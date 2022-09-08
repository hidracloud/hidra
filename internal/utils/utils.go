package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/hidracloud/hidra/v3/internal/metrics"
	log "github.com/sirupsen/logrus"
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

// PrintMetrics prints the metrics
func PrintMetrics(metric []metrics.Metric) {

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
