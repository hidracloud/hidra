package utils

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	strip "github.com/grokify/html-strip-tags-go"
	log "github.com/sirupsen/logrus"
)

const (
	errInvalidDuration = "time: invalid duration"
)

// SetLogLevelFromStr sets the log level from a string
func SetLogLevelFromStr(level string) {
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	case "panic":
		log.SetLevel(log.PanicLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	log.Debug("Log level set to: ", level)
}

// AutoDiscoverYML find yaml in given path
func AutoDiscoverYML(path string) ([]string, error) {
	filesPath := []string{}
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		if f.IsDir() {
			dirFiles, err := AutoDiscoverYML(path + "/" + f.Name())
			if err != nil {
				return nil, err
			}

			filesPath = append(filesPath, dirFiles...)
		}
		if strings.HasSuffix(f.Name(), ".yml") || strings.HasSuffix(f.Name(), ".yaml") {
			filesPath = append(filesPath, path+"/"+f.Name())
		}
	}

	return filesPath, nil
}

// ExtractFileNameWithoutExtension extracts the file name without extension
func ExtractFileNameWithoutExtension(path string) string {
	fileName := path[strings.LastIndex(path, "/")+1:]
	fileName = fileName[:strings.LastIndex(fileName, ".")]

	return fileName
}

// EqualSlices checks if two slices are equal
func EqualSlices[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Map2Hash converts a map to a hash
func Map2Hash(m map[string]string) string {
	var hash string
	for k, v := range m {
		hash += k + v
	}

	return hash
}

// PrintTable prints a table
func PrintTable(table [][]string) {
	// get number of columns from the first table row
	columnLengths := make([]int, len(table[0]))
	for _, line := range table {
		for i, val := range line {
			if len(val) > columnLengths[i] {
				columnLengths[i] = len(val)
			}
		}
	}

	var lineLength int
	for _, c := range columnLengths {
		lineLength += c + 3 // +3 for 3 additional characters before and after each field: "| %s "
	}
	lineLength += 1 // +1 for the last "|" in the line

	for i, line := range table {
		if i == 0 { // table header
			fmt.Printf("+%s+\n", strings.Repeat("-", lineLength-2)) // lineLength-2 because of "+" as first and last character
		}
		for j, val := range line {
			fmt.Printf("| %-*s ", columnLengths[j], val)
			if j == len(line)-1 {
				fmt.Printf("|\n")
			}
		}
		if i == 0 || i == len(table)-1 { // table header or last line
			fmt.Printf("+%s+\n", strings.Repeat("-", lineLength-2)) // lineLength-2 because of "+" as first and last character
		}
	}
}

// Include checks if a string is included in a slice
func Include(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
}

var unitMap = map[string]int64{
	"ns": int64(time.Nanosecond),
	"us": int64(time.Microsecond),
	"µs": int64(time.Microsecond), // U+00B5 = micro symbol
	"μs": int64(time.Microsecond), // U+03BC = Greek letter mu
	"ms": int64(time.Millisecond),
	"s":  int64(time.Second),
	"m":  int64(time.Minute),
	"h":  int64(time.Hour),
	"d":  int64(time.Hour) * 24,
	"w":  int64(time.Hour) * 168,
}

// ParseDuration parses a duration string.
// A duration string is a possibly signed sequence of
// decimal numbers, each with optional fraction and a unit suffix,
// such as "300ms", "-1.5h" or "2h45m".
// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h", "d", "w".
func ParseDuration(s string) (time.Duration, error) {
	// [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+
	orig := s
	var d int64
	neg := false

	// Consume [-+]?
	if s != "" {
		c := s[0]
		if c == '-' || c == '+' {
			neg = c == '-'
			s = s[1:]
		}
	}
	// Special case: if all that is left is "0", this is zero.
	if s == "0" {
		return 0, nil
	}
	if s == "" {
		return 0, errors.New(errInvalidDuration + quote(orig))
	}
	for s != "" {
		var (
			v, f  int64       // integers before, after decimal point
			scale float64 = 1 // value = v + f/scale
		)

		var err error

		// The next character must be [0-9.]
		if !(s[0] == '.' || '0' <= s[0] && s[0] <= '9') {
			return 0, errors.New(errInvalidDuration + quote(orig))
		}
		// Consume [0-9]*
		pl := len(s)
		v, s, err = leadingInt(s)
		if err != nil {
			return 0, errors.New(errInvalidDuration + quote(orig))
		}
		pre := pl != len(s) // whether we consumed anything before a period

		// Consume (\.[0-9]*)?
		post := false
		if s != "" && s[0] == '.' {
			s = s[1:]
			pl := len(s)
			f, scale, s = leadingFraction(s)
			post = pl != len(s)
		}
		if !pre && !post {
			// no digits (e.g. ".s" or "-.s")
			return 0, errors.New(errInvalidDuration + quote(orig))
		}

		// Consume unit.
		i := 0
		for ; i < len(s); i++ {
			c := s[i]
			if c == '.' || '0' <= c && c <= '9' {
				break
			}
		}
		if i == 0 {
			return 0, errors.New("time: missing unit in duration " + quote(orig))
		}
		u := s[:i]
		s = s[i:]
		unit, ok := unitMap[u]
		if !ok {
			return 0, errors.New("time: unknown unit " + quote(u) + " in duration " + quote(orig))
		}
		if v > (1<<63-1)/unit {
			// overflow
			return 0, errors.New(errInvalidDuration + quote(orig))
		}
		v *= unit
		if f > 0 {
			// float64 is needed to be nanosecond accurate for fractions of hours.
			// v >= 0 && (f*unit/scale) <= 3.6e+12 (ns/h, h is the largest unit)
			v += int64(float64(f) * (float64(unit) / scale))
			if v < 0 {
				// overflow
				return 0, errors.New(errInvalidDuration + quote(orig))
			}
		}
		d += v
		if d < 0 {
			// overflow
			return 0, errors.New(errInvalidDuration + quote(orig))
		}
	}

	if neg {
		d = -d
	}
	return time.Duration(d), nil
}

func quote(s string) string {
	return "\"" + s + "\""
}

var errLeadingInt = errors.New("time: bad [0-9]*") // never printed

// leadingInt consumes the leading [0-9]* from s.
func leadingInt(s string) (x int64, rem string, err error) {
	i := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if x > (1<<63-1)/10 {
			// overflow
			return 0, "", errLeadingInt
		}
		x = x*10 + int64(c) - '0'
		if x < 0 {
			// overflow
			return 0, "", errLeadingInt
		}
	}
	return x, s[i:], nil
}

// leadingFraction consumes the leading [0-9]* from s.
// It is used only for fractions, so does not return an error on overflow,
// it just stops accumulating precision.
func leadingFraction(s string) (x int64, scale float64, rem string) {
	i := 0
	scale = 1
	overflow := false
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			break
		}
		if overflow {
			continue
		}
		if x > (1<<63-1)/10 {
			// It's possible for overflow to give a positive number, so take care.
			overflow = true
			continue
		}
		y := x*10 + int64(c) - '0'
		if y < 0 {
			overflow = true
			continue
		}
		x = y
		scale *= 10
	}
	return x, scale, s[i:]
}

var envToMapCache map[string]string

// EnvToMap returns a map of environment variables from the given environment
func EnvToMap() map[string]string {
	if envToMapCache != nil {
		return envToMapCache
	}

	envMap := make(map[string]string)
	for _, v := range os.Environ() {
		splitV := strings.SplitN(v, "=", 2)
		envMap[splitV[0]] = strings.Join(splitV[1:], "=")
	}

	envToMapCache = envMap

	return envMap
}

// HTMLStripTags removes all HTML tags from a string
func HTMLStripTags(s string) string {
	return strip.StripTags(s)
}

// Base64Encode encodes a string to base64
func Base64Encode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// isHeadless returns true if user doesn't have a graphical environment
func IsHeadless() bool {
	if runtime.GOOS == "windows" {
		return false
	} else if runtime.GOOS == "darwin" {
		return false
	}
	return os.Getenv("DISPLAY") == ""
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.Sleep(time.Second),
		chromedp.FullScreenshot(res, quality),
	}
}

// BytesToLowerCase converts all bytes in a byte slice to lowercase
func BytesToLowerCase(b []byte) []byte {
	for i := 0; i < len(b); i++ {
		b[i] = bytes.ToLower([]byte{b[i]})[0]
	}
	return b
}

// StringToInt converts a string to an int
func StringToInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return i
}

// BytesContainsString checks if a byte slice contains a string
func BytesContainsString(b []byte, s string) bool {
	return bytes.Contains(b, []byte(s))
}

// BytesContainsStringXTimes checks if a byte slice contains a string x times
func BytesContainsStringTimes(b []byte, s string) int {
	return bytes.Count(b, []byte(s))
}

// TakeScreenshotWithChromedp takes a screenshot of the current browser window.
func TakeScreenshotWithChromedp(url, file string) error {
	// create context
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)

	// add a timeout
	ctx, cancelTimeout := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	defer cancelTimeout()

	var buf []byte

	// capture entire browser viewport, returning png with quality=90
	if err := chromedp.Run(ctx, fullScreenshot(url, 90, &buf)); err != nil {
		return err
	}
	if err := os.WriteFile(file, buf, 0o644); err != nil {
		return err
	}

	return nil
}

// CamelCaseToSnakeCase converts a string from camel case to snake case
func CamelCaseToSnakeCase(s string) string {
	// detect upper case letters
	upper := regexp.MustCompile(`[A-Z]`)

	// replace upper case letters with lower case letters and a preceding underscore
	newString := strings.ToLower(upper.ReplaceAllStringFunc(s, func(s string) string {
		return "_" + s
	}))

	// remove leading underscore
	return strings.TrimPrefix(newString, "_")
}
