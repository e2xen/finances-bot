package tg

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"max.ks1230/project-base/internal/model/messages"
)

const defaultUpdateOffset = 0

type tokenGetter interface {
	Token() string
}

type Client struct {
	client *tgbotapi.BotAPI
}

func New(tokenGetter tokenGetter) (*Client, error) {
	client, err := tgbotapi.NewBotAPI(tokenGetter.Token())
	if err != nil {
		return nil, errors.Wrap(err, "cannot NewBotApi")
	}
	return &Client{client}, nil
}

func (c *Client) SendMessage(text string, userID int64) error {
	_, err := c.client.Send(tgbotapi.NewMessage(userID, text))
	if err != nil {
		return errors.Wrap(err, "client.Send")
	}
	return nil
}

func (c *Client) ListenUpdates(msgModel *messages.Service, ctx context.Context) {
	u := tgbotapi.NewUpdate(defaultUpdateOffset)
	u.Timeout = 60

	updates := c.client.GetUpdatesChan(u)

	log.Println("Start listening for messages")

	for {
		select {
		case <-ctx.Done():
			log.Println("Stop listening for messages")
			return
		case update := <-updates:
			c.listenOnce(update, msgModel)
		}
	}
}

func (c *Client) listenOnce(update tgbotapi.Update, msgModel *messages.Service) {
	if update.Message != nil {
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		err := msgModel.IncomingMessage(messages.Message{
			Text:   update.Message.Text,
			UserID: update.Message.From.ID,
		})
		if err != nil {
			log.Println("error processing message:", err)
		}
	}
}
