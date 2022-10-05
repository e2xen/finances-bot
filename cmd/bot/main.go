package main

import (
	"log"
	"max.ks1230/project-base/internal/clients/tg"
	"max.ks1230/project-base/internal/config"
	"max.ks1230/project-base/internal/model/messages"
	"max.ks1230/project-base/internal/model/storage"
)

func main() {
	conf, err := config.New()
	if err != nil {
		log.Fatal("failed to init config:", err)
	}

	client, err := tg.New(conf)
	if err != nil {
		log.Fatal("failed to init client:", err)
	}

	userStorage := storage.NewInMemStorage()
	msgService := messages.New(client, userStorage)

	client.ListenUpdates(msgService)
}
