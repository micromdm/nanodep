package slog

import (
	"log/slog"

	"github.com/micromdm/nanolib/log"
)

// Logger wraps stdlib slog for logging.
type Logger struct {
	lctx    []any
	slogger *slog.Logger
}

func New(slogger *slog.Logger) *Logger {
	if slogger == nil {
		panic("nil slogger")
	}
	return &Logger{slogger: slogger}
}

func extractMsg(logs []any) (string, []any) {
	if len(logs) < 2 {
		return "", logs
	}
	for i := 0; i < len(logs); i += 2 {
		if field, ok := logs[i].(string); ok && field == "msg" {
			if i+1 < len(logs) {
				if field, ok = logs[i+1].(string); ok {
					return field, append(logs[0:i], logs[i+2:]...)
				}
			}
		}
	}
	return "", logs
}

// Info logs to the slog at the info level.
func (logger *Logger) Info(logs ...interface{}) {
	allLogs := append(logger.lctx, logs...)
	msg, msglessLogs := extractMsg(allLogs)
	logger.slogger.Info(msg, msglessLogs...)
}

// Debug logs to the slog with at the debug level.
func (logger *Logger) Debug(logs ...interface{}) {
	allLogs := append(logger.lctx, logs...)
	msg, msglessLogs := extractMsg(allLogs)
	logger.slogger.Debug(msg, msglessLogs...)
}

// With returns a logger with additoinal context.
func (logger *Logger) With(logs ...interface{}) log.Logger {
	logger2 := *logger
	logger2.lctx = append(logger2.lctx, logs...)
	return &logger2
}
