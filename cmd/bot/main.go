package main

import (
	"context"
	"log"
	"max.ks1230/project-base/internal/clients/fixer"
	"max.ks1230/project-base/internal/clients/tg"
	"max.ks1230/project-base/internal/config"
	"max.ks1230/project-base/internal/model/messages"
	"max.ks1230/project-base/internal/model/rates"
	"max.ks1230/project-base/internal/model/storage"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cancelOnSignals(cancel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	conf, err := config.New()
	if err != nil {
		log.Fatal("failed to init config:", err)
	}

	tgClient, err := tg.New(conf.Telegram())
	if err != nil {
		log.Fatal("failed to init client:", err)
	}

	fixerClient := fixer.New(conf.Fixer())

	userStorage, err := storage.NewPostgresStorage(ctx, conf.Postgres())
	if err != nil {
		log.Fatal("failed to init postgres:", err)
	}
	msgService := messages.NewService(tgClient, userStorage, conf.App())
	ratesPuller, err := rates.NewPuller(ctx, userStorage, fixerClient, conf.App())
	if err != nil {
		log.Fatal("failed to init puller:", err)
	}

	go ratesPuller.Pull()

	tgClient.ListenUpdates(ctx, msgService)
}

func cancelOnSignals(cancel context.CancelFunc, signals ...os.Signal) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, signals...)
	go func() {
		<-sigChan
		cancel()
	}()
}
