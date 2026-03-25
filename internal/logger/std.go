package logger

import "log"

type StdLogger struct {
	level Level
}

func New(level Level) *StdLogger {
	return &StdLogger{level}
}

func (l *StdLogger) logf(level Level, prefix, msg string, args ...any) {
	if level < l.level {
		return
	}

	log.Printf("[%s] "+msg, append([]any{prefix}, args...)...)
}

func (l *StdLogger) Debug(msg string, args ...any) {
	l.logf(DEBUG, "DEBUG", msg, args...)
}

func (l *StdLogger) Info(msg string, args ...any) {
	l.logf(INFO, "INFO", msg, args...)
}

func (l *StdLogger) Error(msg string, args ...any) {
	l.logf(ERROR, "ERROR", msg, args...)
}

func (l *StdLogger) Warn(msg string, args ...any) {
	l.logf(WARNING, "WARN", msg, args...)
}
