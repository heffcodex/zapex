package zapex

import (
	"net/http"
	"reflect"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	coreSentryFlushTimeout = 5 * time.Second
	errorKey               = "error"
	httpRequestKey         = "http_request"
	omitHeadStackFrames    = 4
)

var sentryLevelMap = map[zapcore.Level]sentry.Level{
	zapcore.FatalLevel:  sentry.LevelFatal,
	zapcore.PanicLevel:  sentry.LevelFatal,
	zapcore.DPanicLevel: sentry.LevelFatal,
	zapcore.ErrorLevel:  sentry.LevelError,
	zapcore.WarnLevel:   sentry.LevelWarning,
	zapcore.InfoLevel:   sentry.LevelInfo,
	zapcore.DebugLevel:  sentry.LevelDebug,
}

type coreSentry struct {
	zapcore.LevelEnabler
	hub    *sentry.Hub
	fields []zapcore.Field
}

func newCoreSentry(hub *sentry.Hub, enab zapcore.LevelEnabler) zapcore.Core {
	lvlError := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return enab.Enabled(lvl) && lvl >= zapcore.ErrorLevel
	})

	return &coreSentry{
		LevelEnabler: lvlError,
		hub:          hub,
		fields:       make([]zapcore.Field, 0),
	}
}

func (c *coreSentry) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()

	clone.fields = append(clone.fields, fields...)

	return clone
}

func (c *coreSentry) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *coreSentry) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	var httpRequest *sentry.Request

	exceptions := make([]sentry.Exception, 0, 1)
	extra := make(map[string]interface{})

	for _, f := range fields {
		switch f.Key {
		case errorKey:
			e := f.Interface.(error)

			exceptions = append(exceptions, sentry.Exception{
				Type:       reflect.TypeOf(e).String(),
				Value:      e.Error(),
				Stacktrace: c.prepareStacktrace(e),
			})
		case httpRequestKey:
			httpRequest = sentry.NewRequest(f.Interface.(*http.Request))
		default:
			extra[f.Key] = f
		}
	}

	event := &sentry.Event{
		Level:     sentryLevelMap[ent.Level],
		Message:   ent.Message,
		Exception: exceptions,
		Extra:     extra,
		Request:   httpRequest,
	}

	if evtID := c.hub.CaptureEvent(event); evtID == nil {
		return errors.New("there's no `Scope` or `Client` available")
	}

	if ent.Level > zapcore.ErrorLevel {
		_ = c.Sync()
	}

	return nil
}

func (c *coreSentry) Sync() (err error) {
	if !c.hub.Flush(coreSentryFlushTimeout) {
		err = errors.New("cannot flush hub")
	}

	return
}

func (c *coreSentry) clone() *coreSentry {
	fieldsCopy := make([]zapcore.Field, 0, len(c.fields))

	copy(fieldsCopy, c.fields)

	return &coreSentry{
		LevelEnabler: c.LevelEnabler,
		hub:          c.hub.Clone(),
		fields:       fieldsCopy,
	}
}

func (c *coreSentry) prepareStacktrace(err error) *sentry.Stacktrace {
	stack := sentry.ExtractStacktrace(err)
	if stack != nil {
		return stack
	}

	stack = sentry.NewStacktrace()
	stack.Frames = stack.Frames[:len(stack.Frames)-omitHeadStackFrames]

	return stack
}
