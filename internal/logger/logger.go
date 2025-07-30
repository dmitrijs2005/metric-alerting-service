// Package logger defines a logging interface and wrapper around zap.Logger,
// providing helper methods and a default logger instance.
package logger

import "go.uber.org/zap"

type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infow(msg string, args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Error(args ...interface{})
	Errorw(msg string, args ...interface{})
	Errorf(format string, args ...interface{})
}

var (
	logger *zap.Logger
)

func GetLogger() *zap.SugaredLogger {
	var err error
	logger, err = zap.NewDevelopment()
	if err != nil {
		// вызываем панику, если ошибка
		panic("cannot initialize zap")
	}
	return logger.Sugar()
}

func Sync() {
	if logger != nil {
		logger.Sync()
	}
}
