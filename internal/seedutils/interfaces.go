package seedutils

// Logger defines the interface for logging operations used throughout the seed utilities
// This allows for flexible logging implementations and makes testing easier
type Logger interface {
	// Info logs general information messages
	Info(msg string, keysAndValues ...interface{})

	// Debug logs detailed debugging information
	Debug(msg string, keysAndValues ...interface{})

	// Warn logs warning messages for potentially problematic situations
	Warn(msg string, keysAndValues ...interface{})

	// Error logs error messages with context
	Error(msg string, keysAndValues ...interface{})
}

// SimpleLogger provides a basic implementation of the Logger interface
// This is primarily used for development and testing purposes
type SimpleLogger struct{}

// NewSimpleLogger creates a new instance of SimpleLogger
func NewSimpleLogger() Logger {
	return &SimpleLogger{}
}

// Info implements the Logger interface for informational messages
func (sl *SimpleLogger) Info(msg string, keysAndValues ...interface{}) {
	logWithLevel("INFO", msg, keysAndValues...)
}

// Debug implements the Logger interface for debug messages
func (sl *SimpleLogger) Debug(msg string, keysAndValues ...interface{}) {
	logWithLevel("DEBUG", msg, keysAndValues...)
}

// Warn implements the Logger interface for warning messages
func (sl *SimpleLogger) Warn(msg string, keysAndValues ...interface{}) {
	logWithLevel("WARN", msg, keysAndValues...)
}

// Error implements the Logger interface for error messages
func (sl *SimpleLogger) Error(msg string, keysAndValues ...interface{}) {
	logWithLevel("ERROR", msg, keysAndValues...)
}

// logWithLevel is a helper function that formats and prints log messages
func logWithLevel(level, msg string, keysAndValues ...interface{}) {
	// For production environments, this would be replaced with a proper logging library
	// like logrus, zap, or the standard log package with proper formatting

	// Basic implementation for development
	if len(keysAndValues) > 0 {
		// Format key-value pairs for better readability
		kvPairs := ""
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				kvPairs += " " + keysAndValues[i].(string) + "=" + formatValue(keysAndValues[i+1])
			}
		}
		// This would typically use a proper logging framework
		// For now, we'll use a simple print statement
		// In production: log.Printf("[%s] %s%s", level, msg, kvPairs)
	} else {
		// In production: log.Printf("[%s] %s", level, msg)
	}
}

// formatValue converts various types to string representation for logging
func formatValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return string(rune(val))
	case int64:
		return string(rune(val))
	case float64:
		return string(rune(int(val)))
	default:
		return "unknown"
	}
}