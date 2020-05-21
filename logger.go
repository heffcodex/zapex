package zapex

import (
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _logger *zap.Logger

func Setup(level string) error {
	hub, err := newSentryHub()
	if err != nil {
		return errors.Wrap(err, "cannot make Sentry hub")
	}

	var zapLevel zapcore.Level

	if err := zapLevel.Set(level); err != nil {
		return errors.Wrap(err, "cannot set global level")
	}

	lvlGlobal := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapLevel
	})

	core := zapcore.NewTee(
		newCoreConsole(lvlGlobal),
		newCoreSentry(hub, lvlGlobal),
	)

	_logger = zap.New(core)

	return nil
}

func L() *zap.Logger {
	return _logger
}
