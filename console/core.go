package console

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/heffcodex/zapex/consts"
)

var consoleEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        consts.KeyTime,
	LevelKey:       consts.KeyLevel,
	NameKey:        consts.KeyName,
	CallerKey:      consts.KeyCaller,
	FunctionKey:    zapcore.OmitKey,
	MessageKey:     consts.KeyMessage,
	StacktraceKey:  consts.KeyStacktrace,
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalLevelEncoder,
	EncodeTime:     zapcore.RFC3339TimeEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   zapcore.ShortCallerEncoder,
}

func NewCore(enab zapcore.LevelEnabler) zapcore.Core {
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
