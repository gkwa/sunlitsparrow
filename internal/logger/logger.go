package logger

import (
	"io"
	"log"
	"os"
)

var (
	InfoLogger  *log.Logger
	DebugLogger *log.Logger
	TraceLogger *log.Logger
	nullWriter  = io.Discard
)

func init() {
	// Set up loggers with prefixes
	InfoLogger = log.New(nullWriter, "INFO: ", log.Ltime)
	DebugLogger = log.New(nullWriter, "DEBUG: ", log.Ltime)
	TraceLogger = log.New(nullWriter, "TRACE: ", log.Ltime)

	// Default to silent
	SetLogLevel(0)
}

// SetLogLevel sets the logging level based on verbosity
func SetLogLevel(verbosity int) {
	// Disable all loggers by default
	InfoLogger.SetOutput(nullWriter)
	DebugLogger.SetOutput(nullWriter)
	TraceLogger.SetOutput(nullWriter)

	// Enable loggers based on verbosity
	if verbosity >= 1 {
		InfoLogger.SetOutput(os.Stderr)
	}
	if verbosity >= 2 {
		DebugLogger.SetOutput(os.Stderr)
	}
	if verbosity >= 3 {
		TraceLogger.SetOutput(os.Stderr)
	}
}

// Info logs an info message if the log level is appropriate
func Info(format string, v ...interface{}) {
	InfoLogger.Printf(format, v...)
}

// Debug logs a debug message if the log level is appropriate
func Debug(format string, v ...interface{}) {
	DebugLogger.Printf(format, v...)
}

// Trace logs a trace message if the log level is appropriate
func Trace(format string, v ...interface{}) {
	TraceLogger.Printf(format, v...)
}
