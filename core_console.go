package zapex

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newCoreConsole(enab zapcore.LevelEnabler) zapcore.Core {
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	lvlError := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return enab.Enabled(lvl) && lvl >= zapcore.ErrorLevel
	})
	lvlMessage := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return enab.Enabled(lvl) && lvl < zapcore.ErrorLevel
	})

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), lvlMessage),
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stderr), lvlError),
	)
}
