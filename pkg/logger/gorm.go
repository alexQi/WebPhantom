package logger

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type GormLogger struct {
	logger *logrus.Logger
}

func NewGormLogger(logger *logrus.Logger) *GormLogger {
	return &GormLogger{logger: logger}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	switch level {
	case logger.Silent:
		newLogger.logger.SetLevel(logrus.PanicLevel)
	case logger.Error:
		newLogger.logger.SetLevel(logrus.ErrorLevel)
	case logger.Warn:
		newLogger.logger.SetLevel(logrus.WarnLevel)
	case logger.Info:
		newLogger.logger.SetLevel(logrus.InfoLevel)
	default:
		newLogger.logger.SetLevel(logrus.InfoLevel)
	}
	return &newLogger
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WithContext(ctx).Infof(msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WithContext(ctx).Warnf(msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.logger.WithContext(ctx).Errorf(msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := logrus.Fields{
		"duration": elapsed,
		"rows":     rows,
	}
	if err != nil {
		fields["error"] = err
		l.logger.WithFields(fields).Error(sql)
	} else {
		l.logger.WithFields(fields).Info(sql)
	}
}
