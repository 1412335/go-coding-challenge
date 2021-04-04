package log

import (
	"context"
	"os"

	"go.uber.org/zap/zapcore"
)

var (
	DefaultLogger Factory = NewFactory(WithLevel(os.Getenv("LOG_LEVEL")))
)

// Logger factory is a generic logging interface
type Factory interface {
	// Init initializes options
	Init(options ...Option) error
	// Get logger
	Bg() Logger
	// Get logger w context
	For(ctx context.Context) Logger
	// Fields set fields to always be logged
	With(fields ...zapcore.Field) Factory
	// Info logs an info msg with fields
	Info(msg string, fields ...zapcore.Field)
	// Error logs an error msg with fields
	Error(msg string, fields ...zapcore.Field)
	// Fatal logs a fatal error msg with fields
	Fatal(msg string, fields ...zapcore.Field)
}

// Init initializes options
func Init(opts ...Option) error {
	return DefaultLogger.Init(opts...)
}

// Get logger
func Bg() Logger {
	return DefaultLogger.Bg()
}

// Get logger w context
func For(ctx context.Context) Logger {
	return DefaultLogger.For(ctx)
}

// Fields set fields to always be logged
func With(fields ...zapcore.Field) Factory {
	return DefaultLogger.With(fields...)
}

// Info logs an info msg with fields
func Info(msg string, fields ...zapcore.Field) {
	DefaultLogger.Info(msg, fields...)
}

// Error logs an error msg with fields
func Error(msg string, fields ...zapcore.Field) {
	DefaultLogger.Error(msg, fields...)
}

// Fatal logs a fatal error msg with fields
func Fatal(msg string, fields ...zapcore.Field) {
	DefaultLogger.Fatal(msg, fields...)
}
