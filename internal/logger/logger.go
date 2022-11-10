package logger

import (
	"log"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	// localLogger, err := zap.NewProduction()
	localLogger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatal("logger init", err)
	}

	logger = localLogger
}

func Info(msg string, fields ...zap.Field) {
	logger.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	logger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	logger.Fatal(msg, fields...)
}
