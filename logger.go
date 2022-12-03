package zapex

import (
	"github.com/heffcodex/zapex/console"
	"github.com/heffcodex/zapex/sentry"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var defaultLogger, _ = zap.NewDevelopment()

func New(level string) (*zap.Logger, error) {
	hub, err := sentry.NewHub()
	if err != nil {
		return nil, errors.Wrap(err, "cannot make sentry.Hub")
	}

	var zapLevel zapcore.Level

	if err := zapLevel.Set(level); err != nil {
		return nil, errors.Wrap(err, "cannot set global level")
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
