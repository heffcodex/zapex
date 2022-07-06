package zapex

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var consoleEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "ts",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeTime:     zapcore.RFC3339TimeEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

func newCoreConsole(enab zapcore.LevelEnabler) zapcore.Core {
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

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
