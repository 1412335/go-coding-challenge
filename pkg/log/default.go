package log

import (
	"context"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type factory struct {
	opts   *configs.Log
	logger *zap.Logger
}

var _ Factory = (*factory)(nil)

func NewFactory(opts ...Option) Factory {
	// default options
	options := &configs.Log{
		Mode:       defaultMode,
		Level:      defaultLevel,
		TraceLevel: defaultTraceLevel,
	}

	f := &factory{
		opts: options,
	}
	// init logger w custom options
	if err := f.Init(opts...); err != nil {
		f.Bg().Error("config log options failed", zap.Error(err))
	}

	// set log level
	level := GetLevel(f.opts.Level)
	traceLevel := GetLevel(f.opts.TraceLevel)

	// new zap.Logger w mode
	var logger *zap.Logger
	if f.opts.Mode == "pro" || f.opts.Mode == "production" {
		logger, _ = zap.NewProduction(
			zap.IncreaseLevel(level),
			zap.AddStacktrace(traceLevel),
		)
	} else {
		logger, _ = zap.NewDevelopment(
			// zap.IncreaseLevel(level),
			zap.AddStacktrace(traceLevel),
			zap.AddCallerSkip(1),
		)
	}
	f.logger = logger
	return f
}

// override options
func (f *factory) Init(opts ...Option) error {
	for _, opt := range opts {
		if err := opt(f.opts); err != nil {
			return err
		}
	}
	return nil
}

// Load new logger
func (f *factory) LoadLogger(l Logger) Factory {
	f.logger = l.(*logger).logger
	return f
}

// Get logger
func (f *factory) Bg() Logger {
	return logger{logger: f.logger}
}

// Get logger w context
func (f *factory) For(ctx context.Context) Logger {
	if span := opentracing.SpanFromContext(ctx); span != nil {
		logger := spanLogger{span: span, logger: f.logger}
		if jaegerCtx, ok := span.Context().(jaeger.SpanContext); ok {
			logger.spanFields = []zapcore.Field{
				zap.String("trace_id", jaegerCtx.TraceID().String()),
				zap.String("span_id", jaegerCtx.SpanID().String()),
			}
		}
		return logger
	}
	return f.Bg()
}

// Fields set fields to always be logged
func (f *factory) With(fields ...zapcore.Field) Factory {
	return &factory{logger: f.logger.With(fields...)}
}

// Info logs an info msg with fields
func (f *factory) Info(msg string, fields ...zapcore.Field) {
	f.logger.Info(msg, fields...)
}

// Error logs an error msg with fields
func (f *factory) Error(msg string, fields ...zapcore.Field) {
	f.logger.Error(msg, fields...)
}

// Fatal logs a fatal error msg with fields
func (f *factory) Fatal(msg string, fields ...zapcore.Field) {
	f.logger.Fatal(msg, fields...)
}
