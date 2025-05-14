package logger

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Level represents the logging level
type Level int

// Log levels
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// String returns the string representation of the log level
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Options configures the logger
type Options struct {
	// Minimum level to log
	Level Level

	// Output writer (defaults to os.Stdout)
	Output io.Writer

	// Whether to show timestamps
	ShowTimestamp bool

	// Whether to show colors (if the output supports it)
	ShowColors bool
}

// Logger provides logging functionality
type Logger struct {
	opts Options
}

// New creates a new logger with the given options
func New(opts Options) *Logger {
	// Set defaults
	if opts.Output == nil {
		opts.Output = os.Stdout
	}

	return &Logger{
		opts: opts,
	}
}

// Default returns a default logger
func Default() *Logger {
	return New(Options{
		Level:         LevelInfo,
		Output:        os.Stdout,
		ShowTimestamp: true,
		ShowColors:    true,
	})
}

// log logs a message at the specified level
func (l *Logger) log(level Level, format string, args ...interface{}) {
	// Skip logging if level is below minimum
	if level < l.opts.Level {
		return
	}

	// Format timestamp
	timestamp := ""
	if l.opts.ShowTimestamp {
		timestamp = time.Now().Format("2006-01-02 15:04:05") + " "
	}

	// Format level with optional colors
	levelStr := level.String()
	if l.opts.ShowColors {
		switch level {
		case LevelDebug:
			levelStr = fmt.Sprintf("\033[36m%s\033[0m", levelStr) // Cyan
		case LevelInfo:
			levelStr = fmt.Sprintf("\033[32m%s\033[0m", levelStr) // Green
		case LevelWarn:
			levelStr = fmt.Sprintf("\033[33m%s\033[0m", levelStr) // Yellow
		case LevelError:
			levelStr = fmt.Sprintf("\033[31m%s\033[0m", levelStr) // Red
		}
	}

	// Format message
	message := fmt.Sprintf(format, args...)

	// Write log entry
	fmt.Fprintf(l.opts.Output, "%s[%s] %s\n", timestamp, levelStr, message)
}

// Debugf logs a debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(LevelDebug, format, args...)
}

// Infof logs an info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(LevelInfo, format, args...)
}

// Warnf logs a warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(LevelWarn, format, args...)
}

// Errorf logs an error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(LevelError, format, args...)
}

// WithLevel creates a new logger with the specified minimum level
func (l *Logger) WithLevel(level Level) *Logger {
	opts := l.opts
	opts.Level = level
	return New(opts)
}

// WithOutput creates a new logger with the specified output
func (l *Logger) WithOutput(output io.Writer) *Logger {
	opts := l.opts
	opts.Output = output
	return New(opts)
}
