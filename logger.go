package zapex

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/heffcodex/zapex/console"
	"github.com/heffcodex/zapex/sentry"
)

var defaultLogger, _ = zap.NewDevelopment()

func New(level string) (*zap.Logger, error) {
	hub, err := sentry.NewHub()
	if err != nil {
		return nil, fmt.Errorf("make sentry.Hub: %w", err)
	}

	var zapLevel zapcore.Level

	if err := zapLevel.Set(level); err != nil {
		return nil, fmt.Errorf("set level: %w", err)
	}

	lvlGlobal := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapLevel
	})

	core := zapcore.NewTee(
		console.NewCore(lvlGlobal),
		sentry.NewCore(hub, lvlGlobal),
	)

	logger := zap.New(core)

	return logger, nil
}

func Default() *zap.Logger {
	return defaultLogger
}

func SetDefault(logger *zap.Logger) {
	defaultLogger = logger
}
