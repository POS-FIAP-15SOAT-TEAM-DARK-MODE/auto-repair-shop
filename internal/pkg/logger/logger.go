package logger

import (
	"context"
	"fmt"
	"strings"

	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop/internal/pkg/env"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type customLogger struct {
	*zap.Logger
}

type ctxKey struct{}

var globalLog *customLogger

func init() {
	initGlobalLogger()
}

func initGlobalLogger() {
	globalLog = &customLogger{initLogger()}
}

func initLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "timestamp"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.MessageKey = "message"
	cfg.Level = getLogLevel(env.GetString("LOG_LEVEL", "INFO"))

	l, err := cfg.Build()
	if err != nil {
		panic(err)
	}

	return l
}

func getLogLevel(level string) zap.AtomicLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	case "PANIC":
		return zap.NewAtomicLevelAt(zap.PanicLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func Global() *customLogger {
	return globalLog
}

func Request(ctx context.Context) (context.Context, string) {
	reqId := uuid.New().String()
	l := initLogger().With(
		zap.String("request_id", reqId),
	)
	return withContext(ctx, &customLogger{l}), reqId
}

func Of(ctx context.Context) *customLogger {
	if ctx == nil {
		return Global()
	}

	if l, ok := ctx.Value(ctxKey{}).(*customLogger); ok {
		return l
	}
	return Global()
}

func withContext(ctx context.Context, l *customLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

func (l *customLogger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

func (l *customLogger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

func (l *customLogger) Error(err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	l.Logger.Error(err.Error(), fields...)
}

func (l *customLogger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

func (l *customLogger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

func (l *customLogger) Sync() {
	if l != nil {
		if err := l.Logger.Sync(); err != nil {
			err = fmt.Errorf("failed to sync logger: %w", err)
			l.Error(err)
		}
	}
}
