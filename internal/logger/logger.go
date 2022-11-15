package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
)

const (
	logEnvKey     = "LOG_ENV"
	defaultLogEnv = "dev"
)

var logger *zap.Logger

func init() {
	env := os.Getenv(logEnvKey)
	if env == "" {
		env = defaultLogEnv
	}

	var err error
	if env == "dev" {
		logger, err = zap.NewDevelopment()
	} else if env == "prod" {
		logger, err = zap.NewProduction()
	}

	if err != nil || logger == nil {
		log.Fatal("logger init", err)
	}
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
