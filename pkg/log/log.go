package log

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"
)

// osExit is a variable for os.Exit to make it mockable in tests
var osExit = os.Exit

// LogLevel define log level
type LogLevel int

const (
	// DEBUG debug level, always show
	DEBUG LogLevel = iota
	// INFO info level, always show
	INFO
	// WARNING warning level, always show
	WARNING
	// ERROR error level, always show
	ERROR
	// FATAL fatal level, always show and exit program
	FATAL
)

var (
	// verbose mode, always show
	verbose bool
	// current log level
	level LogLevel = INFO
	// enable color output
	colorEnabled = true
	// enable stack trace on fatal errors
	stackTraceEnabled bool
)

// Environment variable for controlling stack trace
const (
	EnvStackTrace = "GOAT_STACK_TRACE"
)

func init() {
	// Check environment variable for stack trace setting
	stackTraceEnv := os.Getenv(EnvStackTrace)
	stackTraceEnabled = stackTraceEnv == "1" || stackTraceEnv == "true" || stackTraceEnv == "yes"
}

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorPurple = "\033[35m"
)

// SetVerbose set verbose mode, always show
func SetVerbose(v bool) {
	verbose = v
	if v {
		level = DEBUG
	}
}

// IsVerbose return if verbose mode is enabled
func IsVerbose() bool {
	return verbose
}

// SetLevel set log level
func SetLevel(l LogLevel) {
	level = l
}

// GetLevel get current log level
func GetLevel() LogLevel {
	return level
}

// EnableColor enables color output
func EnableColor(enabled bool) {
	colorEnabled = enabled
}

// IsColorEnabled returns if color output is enabled
func IsColorEnabled() bool {
	return colorEnabled
}

// EnableStackTrace enables or disables stack trace on fatal errors
func EnableStackTrace(enabled bool) {
	stackTraceEnabled = enabled
}

// IsStackTraceEnabled returns if stack trace is enabled
func IsStackTraceEnabled() bool {
	return stackTraceEnabled
}

// getLevelColor returns the color for the given log level
func getLevelColor(l LogLevel) string {
	if !colorEnabled {
		return ""
	}

	switch l {
	case DEBUG:
		return ColorCyan
	case INFO:
		return ColorGreen
	case WARNING:
		return ColorYellow
	case ERROR:
		return ColorRed
	case FATAL:
		return ColorPurple
	default:
		return ""
	}
}

// getLevelPrefix returns the colored prefix for the given log level
func getLevelPrefix(prefix string, l LogLevel) string {
	if !colorEnabled {
		return prefix
	}

	color := getLevelColor(l)
	return color + prefix + ColorReset
}

// time format
const timeFormat = "2006/01/02 15:04:05"

// formatted log output
func logf(prefix string, l LogLevel, format string, args ...any) {
	if l < level {
		return
	}
	timeStr := time.Now().Format(timeFormat)
	coloredPrefix := getLevelPrefix(prefix, l)
	fmt.Fprintf(os.Stdout, "[%s] %s: %s\n", timeStr, coloredPrefix, fmt.Sprintf(format, args...))
}

// non-formatted log output
func log(prefix string, l LogLevel, args ...any) {
	if l < level {
		return
	}

	timeStr := time.Now().Format(timeFormat)
	coloredPrefix := getLevelPrefix(prefix, l)
	fmt.Fprintf(os.Stdout, "[%s] %s: %s\n", timeStr, coloredPrefix, fmt.Sprint(args...))
}

// Info output normal info log
func Info(args ...any) {
	log("INFO", INFO, args...)
}

// Infof output formatted normal info log
func Infof(format string, args ...any) {
	logf("INFO", INFO, format, args...)
}

// Error output error log
func Error(args ...any) {
	log("ERROR", ERROR, args...)
}

// Errorf output formatted error log
func Errorf(format string, args ...any) {
	logf("ERROR", ERROR, format, args...)
}

// Warning output warning log
func Warning(args ...any) {
	log("WARNING", WARNING, args...)
}

// Warningf output formatted warning log
func Warningf(format string, args ...any) {
	logf("WARNING", WARNING, format, args...)
}

// Fatal output fatal log and exit program
func Fatal(args ...any) {
	log("FATAL", FATAL, args...)
	if stackTraceEnabled {
		fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
	} else {
		fmt.Fprintln(os.Stderr, "For detailed stack trace, set GOAT_STACK_TRACE=1")
	}
	osExit(1)
}

// Fatalf output formatted fatal log and exit program
func Fatalf(format string, args ...any) {
	logf("FATAL", FATAL, format, args...)
	if stackTraceEnabled {
		fmt.Fprintf(os.Stderr, "Stack trace:\n%s\n", debug.Stack())
	} else {
		fmt.Fprintln(os.Stderr, "For detailed stack trace, set GOAT_STACK_TRACE=1")
	}
	osExit(1)
}

// Debug output debug log (only effective in verbose mode)
func Debug(args ...any) {
	log("DEBUG", DEBUG, args...)
}

// Debugf output formatted debug log (only effective in verbose mode)
func Debugf(format string, args ...any) {
	logf("DEBUG", DEBUG, format, args...)
}
