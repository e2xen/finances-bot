package main

import (
	"context"
	"os"
	"os/signal"

	"max.ks1230/finances-bot/internal/model/reports"

	"go.uber.org/zap"
	"max.ks1230/finances-bot/internal/clients/kafka"
	"max.ks1230/finances-bot/internal/config"
	"max.ks1230/finances-bot/internal/logger"
	"max.ks1230/finances-bot/internal/model/storage"
)

const acceptorAddr = "127.0.0.1:8080"

func main() {
	logger.Info("Reporter init - start")

	conf, err := config.New()
	if err != nil {
		logger.Fatal("failed to init config:", zap.Error(err))
	}

	db, err := storage.NewPostgresStorage(conf.Postgres())
	if err != nil {
		logger.Fatal("failed to init postgres:", zap.Error(err))
	}

	reportGenerator := reports.NewGenerator(conf.App(), db)

	reportSender, err := reports.NewSender(acceptorAddr)
	if err != nil {
		logger.Fatal("failed to init grpc client", zap.Error(err))
	}
	defer reportSender.Close()

	consumer, err := kafka.NewConsumer(conf.Kafka(), reportGenerator, reportSender)
	if err != nil {
		logger.Fatal("failed to init kafka consumer", zap.Error(err))
	}

	logger.Info("Reporter init - end")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err = consumer.StartConsuming(ctx); err != nil {
		logger.Fatal("failed to start consuming")
	}
}
