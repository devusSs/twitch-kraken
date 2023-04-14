package logging

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func CheckErrorLogs(logsDir string) (string, error) {
	if !logsDirExist(logsDir) {
		return "", errors.New("logs directory does not exist")
	}

	f, err := loadErrorLog(logsDir)
	if err != nil {
		return "", err
	}

	scanner := bufio.NewScanner(f)

	foundErrs := []string{}
	for scanner.Scan() {
		line := scanner.Text()
		// Remove any logging information in front of the error.
		re := regexp.MustCompile(`^\d{4}/\d{2}/\d{2}\s\d{2}:\d{2}:\d{2}\s`)
		output := re.ReplaceAllString(line, "")

		foundErrs = append(foundErrs, output)
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	res := fmt.Sprintf("\nFound %d error(s) in error log file.\nThe following errors have been found:\n\n%s", len(foundErrs), strings.Join(foundErrs, "\n"))

	if len(foundErrs) == 0 {
		res = ""
	}

	return res, f.Close()
}

func logsDirExist(logsDir string) bool {
	if _, err := os.Stat(logsDir); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func loadErrorLog(logsDir string) (*os.File, error) {
	return os.Open(fmt.Sprintf("%s/error_%s.log", logsDir, formatDate))
}
