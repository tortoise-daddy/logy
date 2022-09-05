package logy

import (
	"context"
	"io"
	"time"
)

var (
	// std is the name of the standard logger in stdlib `log`
	std = New()
)

func StandardLogger() *Logger {
	return std
}

// SetOutput sets the standard logger output.
func SetOutput(out io.Writer) {
	std.SetOutput(out)
}

// SetFormatter sets the standard logger formatter.
func SetFormatter(formatter Formatter) {
	std.SetFormatter(formatter)
}


func SetReportCaller(include bool) {
	std.SetReportCaller(include)
}


func SetLevel(level Level) {
	std.SetLevel(level)
}


func GetLevel() Level {
	return std.GetLevel()
}


func IsLevelEnabled(level Level) bool {
	return std.IsLevelEnabled(level)
}


func WithError(err error) *Entry {
	return std.WithField(ErrorKey, err)
}

func WithContext(ctx context.Context) *Entry {
	return std.WithContext(ctx)
}

func WithField(key string, value interface{}) *Entry {
	return std.WithField(key, value)
}


func WithFields(fields Fields) *Entry {
	return std.WithFields(fields)
}

func WithTime(t time.Time) *Entry {
	return std.WithTime(t)
}

// Trace logs a message at level Trace on the standard logger.
func Trace(args ...interface{}) {
	std.Trace(args...)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(args ...interface{}) {
	std.Debug(args...)
}
/*
// Print logs a message at level Info on the standard logger.
func Print(args ...interface{}) {
	std.Print(args...)
}
*/
// Info logs a message at level Info on the standard logger.
func Info(args ...interface{}) {
	std.Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(args ...interface{}) {
	std.Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(args ...interface{}) {
	std.Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(args ...interface{}) {
	std.Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	std.Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger then the process will exit with status set to 1.
func Fatal(args ...interface{}) {
	std.Fatal(args...)
}
