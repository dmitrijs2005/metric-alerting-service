package logger

import "go.uber.org/zap"

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
