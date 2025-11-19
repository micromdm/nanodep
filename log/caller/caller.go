package caller

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/micromdm/nanolib/log"
)

// Logger wraps another logger and adds callsite information to the logs.
// Be careful about placement and usage of this logger as correct
// reporting of the callsite depends on call depth/stack frame location.
type Logger struct {
	logger log.Logger
	lctx   []interface{}
	skip   int
}

type Option func(*Logger)

// WithSkip skips over any additional stack frames.
// The argument skip is the number of additional stack frames to ascend.
// This option is unnecessary if this logger is the "top" of the logging stack.
func WithSkip(skip int) Option {
	return func(l *Logger) {
		l.skip = skip
	}
}

func New(logger log.Logger, opts ...Option) *Logger {
	if logger == nil {
		panic("nil logger")
	}
	l := &Logger{logger: logger}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// caller returns filename:lineno
func caller(skip int) string {
	_, filename, line, _ := runtime.Caller(skip + 1) // skip past *this* func
	return fmt.Sprintf("%s:%d", filepath.Base(filename), line)
}

// Info logs at the info level with caller info.
func (logger *Logger) Info(logs ...interface{}) {
	logs = append(logs, "caller", caller(logger.skip+1))
	logger.logger.Info(append(logger.lctx, logs...)...)
}

// Debug logs at the debug level with caller info.
func (logger *Logger) Debug(logs ...interface{}) {
	logs = append(logs, "caller", caller(logger.skip+1))
	logger.logger.Debug(append(logger.lctx, logs...)...)
}

// With returns a logger with additoinal context.
func (logger *Logger) With(logs ...interface{}) log.Logger {
	logger2 := *logger
	logger2.lctx = append(logger2.lctx, logs...)
	return &logger2
}
