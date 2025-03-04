package utils

import (
	"os"

	"github.com/charmbracelet/log"
)

// LogLevel represents a log level
type LogLevel string

const (
	// LogLevelDebug represents debug level
	LogLevelDebug LogLevel = "debug"

	// LogLevelInfo represents info level
	LogLevelInfo LogLevel = "info"

	// LogLevelWarn represents warn level
	LogLevelWarn LogLevel = "warn"

	// LogLevelError represents error level
	LogLevelError LogLevel = "error"
)

// InitLogging initializes logging
func InitLogging(level LogLevel) {
	// Set the log level
	switch level {
	case LogLevelDebug:
		log.SetLevel(log.DebugLevel)
	case LogLevelInfo:
		log.SetLevel(log.InfoLevel)
	case LogLevelWarn:
		log.SetLevel(log.WarnLevel)
	case LogLevelError:
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Set the log format
	log.SetReportCaller(true)
	log.SetOutput(os.Stderr)
}

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	log.Debug(msg, args...)
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	log.Info(msg, args...)
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	log.Warn(msg, args...)
}

// Error logs an error message
func Error(msg string, args ...interface{}) {
	log.Error(msg, args...)
}

// Fatal logs a fatal message and exits
func Fatal(msg string, args ...interface{}) {
	log.Fatal(msg, args...)
}

// GetLogLevel returns the log level from a string
func GetLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "warn":
		return LogLevelWarn
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}
