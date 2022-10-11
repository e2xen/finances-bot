package main

import (
	"log"
	"max.ks1230/project-base/internal/clients/fixer"
	"max.ks1230/project-base/internal/clients/tg"
	"max.ks1230/project-base/internal/config"
	"max.ks1230/project-base/internal/model/messages"
	"max.ks1230/project-base/internal/model/rates"
	"max.ks1230/project-base/internal/model/storage"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal("failed to init config:", err)
	}

	tgClient, err := tg.New(conf.Telegram())
	if err != nil {
		log.Fatal("failed to init client:", err)
	}

	fixerClient := fixer.New(conf.Fixer())

	userStorage := storage.NewInMemStorage()
	msgService := messages.NewService(tgClient, userStorage, conf.App())
	ratesPuller, err := rates.NewPuller(userStorage, fixerClient, conf.App())
	if err != nil {
		log.Fatal("failed to init puller:", err)
	}

	go ratesPuller.Pull()

	tgClient.ListenUpdates(msgService)
}
