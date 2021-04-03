package log

import (
	"os"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"go.uber.org/zap/zapcore"
)

var (
	defaultMode       = os.Getenv("GOENV")
	defaultLevel      = "DEBUG"
	defaultTraceLevel = "FATAL"
)

type Option func(*configs.Log) error

func WithMode(mode string) Option {
	return func(o *configs.Log) error {
		o.Mode = mode
		return nil
	}
}

func WithLevel(level string) Option {
	return func(o *configs.Log) error {
		if level != "" {
			o.Level = level
		}
		return nil
	}
}

func WithTraceLevel(level string) Option {
	return func(o *configs.Log) error {
		if level != "" {
			o.TraceLevel = level
		}
		return nil
	}
}

func GetLevel(level string) zapcore.Level {
	switch level {
	case "PANNIC":
		return zapcore.PanicLevel
	case "FATAL":
		return zapcore.FatalLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "WARN":
		return zapcore.WarnLevel
	case "INFO":
		return zapcore.InfoLevel
	case "DEBUG":
		return zapcore.DebugLevel
	default:
		return zapcore.InfoLevel
	}
}
