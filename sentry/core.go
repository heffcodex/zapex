package sentry

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
	defaultLevel        = zapcore.WarnLevel
	flushTimeout        = 5 * time.Second
	omitHeadStackFrames = 4
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

var (
	errType      = reflect.TypeOf((*error)(nil)).Elem()
	errArrayType = reflect.TypeOf([]error{})
)

type Core struct {
	zapcore.LevelEnabler
	hub    *sentry.Hub
	fields []zapcore.Field
}

func NewCore(hub *sentry.Hub, enab zapcore.LevelEnabler) zapcore.Core {
	lvlError := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return enab.Enabled(lvl) && lvl >= defaultLevel
	})

	return &Core{
		LevelEnabler: lvlError,
		hub:          hub,
		fields:       make([]zapcore.Field, 0),
	}
}

func (c *Core) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	clone.fields = append(clone.fields, fields...)

	return clone
}

func (c *Core) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}

	return ce
}

func (c *Core) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	evt := sentry.NewEvent()
	evt.Level = sentryLevelMap[ent.Level]
	evt.Message = ent.Message

	ns := ""

	for _, f := range fields {
		switch f.Type {
		case zapcore.SkipType:
			continue
		case zapcore.NamespaceType:
			if ns == "" {
				ns = f.Key
			} else {
				ns += "." + f.Key
			}

			continue
		}

		writeField(ns, f, evt)
	}

	if evtID := c.hub.CaptureEvent(evt); evtID == nil {
		return errors.New("there's no `Scope` or `Client` available")
	}

	if ent.Level > zapcore.ErrorLevel {
		_ = c.Sync()
	}

	return nil
}

func writeField(ns string, field zapcore.Field, evt *sentry.Event) {
	key := field.Key
	if ns != "" {
		key = ns + "." + field.Key
	}

	vof := reflect.ValueOf(field.Interface)
	tof := vof.Type()

	if tof.ConvertibleTo(errType) {
		evt.Exception = append(evt.Exception, makeExceptions(key, vof.Convert(errType).Interface().(error))...)
		return
	} else if tof.ConvertibleTo(errArrayType) {
		evt.Exception = append(evt.Exception, makeExceptions(key, vof.Convert(errArrayType).Interface().([]error)...)...)
		return
	}

	if v, ok := field.Interface.(*http.Request); ok {
		evt.Request = sentry.NewRequest(v)
		return
	}

	evt.Extra[key] = field
}

func (c *Core) Sync() (err error) {
	if !c.hub.Flush(flushTimeout) {
		err = errors.New("cannot flush hub")
	}

	return
}

func (c *Core) clone() *Core {
	fieldsCopy := make([]zapcore.Field, 0, len(c.fields))

	copy(fieldsCopy, c.fields)

	return &Core{
		LevelEnabler: c.LevelEnabler,
		hub:          c.hub.Clone(),
		fields:       fieldsCopy,
	}
}