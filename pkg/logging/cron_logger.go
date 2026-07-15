// Package logging adapts the project's slog-based logging to third-party
// logger interfaces so every log line, including ones emitted from inside
// vendored dependencies, goes through the same structured sink.
package logging

import "log/slog"

// CronLogger adapts *slog.Logger to github.com/robfig/cron/v3's Logger
// interface (Info(msg, kv...) / Error(err, msg, kv...)), so cron's internal
// recover/skip-if-still-running middleware logs through slog as well.
type CronLogger struct {
	logger *slog.Logger
}

func NewCronLogger(logger *slog.Logger) CronLogger {
	return CronLogger{logger: logger}
}

func (l CronLogger) Info(msg string, keysAndValues ...any) {
	l.logger.Info(msg, keysAndValues...)
}

func (l CronLogger) Error(err error, msg string, keysAndValues ...any) {
	l.logger.Error(msg, append(keysAndValues, "error", err)...)
}
