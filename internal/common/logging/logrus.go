package logging

import (
	"context"
	"github.com/liuzhaoze/MyGo-project/common/tracing"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

type traceHook struct {
}

func (t traceHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (t traceHook) Fire(entry *logrus.Entry) error {
	if entry.Context != nil {
		entry.Data["trace"] = tracing.TraceID(entry.Context)
		entry = entry.WithTime(time.Now())
	}
	return nil
}

func Init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(&traceHook{})
}

func SetFormatter(logger *logrus.Logger) {
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
			logrus.FieldKeyTime:  "time",
			logrus.FieldKeyMsg:   "message",
		},
	})
	if isLocal, _ := strconv.ParseBool(os.Getenv("LOCAL_MODE")); isLocal {
		logger.SetFormatter(&prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			TimestampFormat: time.RFC3339,
		})
	}
}

// Optional: 使用logging.Infof等；或者使用logrus提供的hook

func logf(ctx context.Context, level logrus.Level, fields logrus.Fields, format string, args ...interface{}) {
	logrus.WithContext(ctx).WithFields(fields).Logf(level, format, args...)
}

func Infof(ctx context.Context, fields logrus.Fields, format string, args ...interface{}) {
	logrus.WithContext(ctx).WithFields(fields).Infof(format, args...)
}

func InfofWithCost(ctx context.Context, fields logrus.Fields, start time.Time, format string, args ...interface{}) {
	fields[Cost] = time.Since(start).Milliseconds()
	Infof(ctx, fields, format, args...)
}

func Warnf(ctx context.Context, fields logrus.Fields, format string, args ...interface{}) {
	logrus.WithContext(ctx).WithFields(fields).Warnf(format, args...)
}

func Errorf(ctx context.Context, fields logrus.Fields, format string, args ...interface{}) {
	logrus.WithContext(ctx).WithFields(fields).Errorf(format, args...)
}

func Panicf(ctx context.Context, fields logrus.Fields, format string, args ...interface{}) {
	logrus.WithContext(ctx).WithFields(fields).Panicf(format, args...)
}
