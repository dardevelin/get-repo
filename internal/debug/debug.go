//go:build debug

package debug

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

var debugLogger *log.Logger

func init() {
	// Create debug.log file
	file, err := os.OpenFile("debug.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("Failed to open debug.log: %v\n", err)
		return
	}
	
	debugLogger = log.New(file, "", log.LstdFlags|log.Lmicroseconds)
	debugLogger.Println("=== DEBUG SESSION STARTED ===")
}

// Log writes a debug message with caller information
func Log(format string, args ...interface{}) {
	if debugLogger == nil {
		return
	}
	
	// Get caller information
	_, file, line, ok := runtime.Caller(1)
	if ok {
		file = filepath.Base(file)
		prefix := fmt.Sprintf("[%s:%d] ", file, line)
		message := fmt.Sprintf(format, args...)
		debugLogger.Printf("%s%s", prefix, message)
	} else {
		debugLogger.Printf(format, args...)
	}
}

// LogFunction logs function entry and exit
func LogFunction(name string) func() {
	Log("ENTER: %s", name)
	start := time.Now()
	return func() {
		Log("EXIT:  %s (took %v)", name, time.Since(start))
	}
}

// LogError logs an error with context
func LogError(err error, context string) {
	if err != nil {
		Log("ERROR: %s - %v", context, err)
	}
}

// LogState logs application state changes
func LogState(component, oldState, newState string) {
	Log("STATE: %s changed from %s to %s", component, oldState, newState)
}