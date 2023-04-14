package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
)

// All of these vars will be assigned automatically, do not modify!
var (
	// "✓" in green
	SuccessSign = color.GreenString("✓")
	// "i" in yellow
	InfoSign = color.YellowString("i")
	// "x" in red
	ErrorSign = color.RedString("x")
	// "!" in magenta
	WarnSign = color.MagentaString("!")

	// # NOTE: Do not modify manually!
	logDir      = ""
	currentDate = time.Now().Local()
	// y_m_d - does not display time
	formatDate = fmt.Sprintf("%d_%d_%d", currentDate.Year(), currentDate.Month(), currentDate.Day())

	// # NOTE: Do not modify manually!
	openLogFiles []*os.File

	// Logger used to write to app.log file.
	appLogger *log.Logger
	// Logger used to write to error.log file.
	errorLogger *log.Logger

	// Logger used to write to stdout.
	consoleInfoLogger *log.Logger
	// Logger used to write to stderr.
	consoleErrorLogger *log.Logger
)

// Creates a directory with specified name and stores the name in a variable.
//
// Returns a nil error if the directory already exists.
func CreateLogsDirectory(dir string) error {
	logDir = dir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.Mkdir(dir, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}

// Creates the app.log and error.log file depending on date.
//
// Appends to files if they already exist.
//
// Does not create any custom log files.
//
// # Use the CreateNewLogFile function for that.
func CreateFileLoggers() error {
	appLog := fmt.Sprintf("%s/app_%s.log", logDir, formatDate)
	errorLog := fmt.Sprintf("%s/error_%s.log", logDir, formatDate)

	var err error

	appLogFile, err := os.OpenFile(appLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	errorLogFile, err := os.OpenFile(errorLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	openLogFiles = append(openLogFiles, appLogFile)
	openLogFiles = append(openLogFiles, errorLogFile)

	appLogger = log.New(appLogFile, "", log.LstdFlags)
	errorLogger = log.New(errorLogFile, "", log.LstdFlags)

	return nil
}

// Creates the loggers for console and file logging.
//
// # NOTE: They will not work after closing the log files.
func CreateConsoleLoggers() {
	consoleInfoLogger = log.New(os.Stdout, "", log.LstdFlags)
	consoleErrorLogger = log.New(os.Stderr, "", log.LstdFlags)
}

// Creates a new log file in logs directory with specified name.
//
// Automatically appends the new log file to open log files and closes it on app exit.
func CreateNewLogFile(name string) (*os.File, error) {
	if logDir == "" {
		return nil, errors.New("logDir not set")
	}

	f, err := os.OpenFile(fmt.Sprintf("%s/%s_%s.log", logDir, name, formatDate), os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return nil, err
	}

	openLogFiles = append(openLogFiles, f)

	return f, nil
}

// Writes a message of type interface to os.Stdout and the app.log file.
//
// Prepends time, date and InfoSign to signalise it's an info log process.
func WriteInfo(message interface{}) {
	appLogger.Println(message)
	consoleInfoLogger.Printf("[%s] %v\n", InfoSign, message)
}

// Writes a message of type interface to os.Stdout and the app.log file.
//
// Prepends time, date and SuccessSign to signalise it's a success log process.
func WriteSuccess(message interface{}) {
	appLogger.Println(message)
	consoleInfoLogger.Printf("[%s] %v\n", SuccessSign, message)
}

// Writes a message of type interface to os.Stderr and the error.log file.
//
// Prepends time, date and ErrorSign to signalise it's an error log process.
//
// # NOTE: does not shutdown the app, use default log.Fatal or os.Exit(1) for that
func WriteError(message interface{}) {
	errorLogger.Println(message)
	consoleErrorLogger.Printf("[%s] %v\n", ErrorSign, message)
}

// Writes a message of type interface to os.Stdout and the app.log file.
//
// Prepends time, date and WarnSign to signalise it's a warn log process.
//
// # NOTE: does not shutdown the app, use default log.Fatal or os.Exit(1) for that
func WriteWarn(message interface{}) {
	appLogger.Println(message)
	consoleInfoLogger.Printf("[%s] %v\n", WarnSign, message)
}

// This function closes every open log file and logger.
//
// # NOTE: only use it on app exit, else it might kill the app.
func CloseLogFiles() error {
	for _, f := range openLogFiles {
		err := f.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
