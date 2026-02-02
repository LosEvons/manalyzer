package manalyzer

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime/debug"
)

var (
	appLogger     *log.Logger
	logFile       *os.File
	logFilePath   string
)

// InitLogger initializes the application logger
func InitLogger() error {
	// Get log directory
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %v", err)
		}
		logFilePath = filepath.Join(homeDir, ".manalyzer", "manalyzer.log")
	} else {
		logFilePath = filepath.Join(configDir, "manalyzer", "manalyzer.log")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(logFilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file in append mode
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create logger
	appLogger = log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	
	// Log session start
	appLogger.Println("========================================")
	appLogger.Println("Manalyzer session started")
	appLogger.Printf("Log file: %s", logFilePath)
	appLogger.Println("========================================")

	return nil
}

// CloseLogger closes the log file
func CloseLogger() {
	if logFile != nil {
		appLogger.Println("========================================")
		appLogger.Println("Manalyzer session ended")
		appLogger.Println("========================================")
		logFile.Close()
	}
}

// LogInfo logs an informational message
func LogInfo(format string, v ...interface{}) {
	if appLogger != nil {
		appLogger.Printf("[INFO] "+format, v...)
	}
	// Also log to stdout for development
	log.Printf("[INFO] "+format, v...)
}

// LogError logs an error message
func LogError(format string, v ...interface{}) {
	if appLogger != nil {
		appLogger.Printf("[ERROR] "+format, v...)
	}
	// Also log to stderr for development
	log.Printf("[ERROR] "+format, v...)
}

// LogPanic logs a panic with stack trace
func LogPanic(recovered interface{}) {
	if appLogger != nil {
		appLogger.Printf("[PANIC] Recovered from panic: %v", recovered)
		appLogger.Printf("[PANIC] Stack trace:\n%s", debug.Stack())
	}
	// Also log to stderr
	log.Printf("[PANIC] Recovered from panic: %v", recovered)
	log.Printf("[PANIC] Stack trace:\n%s", debug.Stack())
}

// RecoverFromPanic should be used with defer to recover from panics
func RecoverFromPanic(context string) {
	if r := recover(); r != nil {
		LogPanic(r)
		LogError("Panic in %s: %v", context, r)
	}
}

// GetLogFilePath returns the path to the log file
func GetLogFilePath() string {
	return logFilePath
}
