package log

import (
	"fmt"

	errs "errors"

	"go.uber.org/zap"
)

type Context map[string]interface{}

type ContextError interface {
	Error() string
	Log(int)
	Add(Context)
}

func (c Context) Log(msg string, lvl Level) {
	logCallerAdjusted(1, msg, lvl, c)
}

func (c Context) NewError(err interface{}, lvl Level) ContextError {
	switch e := err.(type) {
	case error:
		return &ce{e, lvl, c}
	case string:
		return &ce{errs.New(e), lvl, c}
	default:
		return &ce{fmt.Errorf("%+v", e), lvl, c}
	}
}

type ce struct {
	Err     error
	Lvl     Level
	Context map[string]interface{}
}

func (e ce) Log(addToCallLevel int) {
	logCallerAdjusted(1+addToCallLevel, e.Err.Error(), e.Lvl, e.Context)
}

func (e ce) Error() string {
	return e.Err.Error()
}

func (e *ce) Add(c Context) {
	for k, v := range c {
		e.Context[k] = v
	}
}

func logCallerAdjusted(adjust int, msg string, lvl Level, c Context) {
	f := fields(c)
	l := Log.WithOptions(zap.AddCallerSkip(1 + adjust)).Sugar()
	switch lvl {
	case debug:
		l.Debugw(msg, f...)
	case info:
		l.Infow(msg, f...)
	case warn:
		l.Warnw(msg, f...)
	case errors:
		l.Errorw(msg, f...)
	case dpanic:
		l.DPanicw(msg, f...)
	case panics:
		l.Panicw(msg, f...)
	case fatal:
		l.Fatalw(msg, f...)
	}
}

func fields(c Context) []interface{} {
	fields := make([]interface{}, 0, len(c))
	for k, v := range c {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
