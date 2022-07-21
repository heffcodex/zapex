package zapex

import (
	"net/http"

	"github.com/heffcodex/zapex/console"
	"github.com/heffcodex/zapex/consts"
	"github.com/heffcodex/zapex/sentry"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _logger, _ = zap.NewDevelopment()

func Setup(level string) error {
	hub, err := sentry.NewHub()
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
		console.NewCore(lvlGlobal),
		sentry.NewCore(hub, lvlGlobal),
	)

	_logger = zap.New(core)

	return nil
}

func L() *zap.Logger {
	return _logger
}

func HTTPRequest(r *http.Request) zap.Field {
	if r == nil {
		return zap.Skip()
	}

	return zap.Field{Key: consts.KeyHTTPRequest, Type: zapcore.ReflectType, Interface: r}
}
