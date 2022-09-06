package utils

import (
	"os"
	"strings"

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
