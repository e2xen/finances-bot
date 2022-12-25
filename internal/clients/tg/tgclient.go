package tg

import (
	"context"
	"time"

	"go.uber.org/zap"
	"max.ks1230/finances-bot/internal/logger"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"max.ks1230/finances-bot/internal/model/messages"
)

const (
	defaultUpdateOffset = 0
	timeoutSeconds      = 5
)

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

func (c *Client) ListenUpdates(ctx context.Context, msgModel *messages.Service) {
	u := tgbotapi.NewUpdate(defaultUpdateOffset)
	u.Timeout = 60

	updates := c.client.GetUpdatesChan(u)

	logger.Info("Start listening for messages")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stop listening for messages")
			return
		case update := <-updates:
			c.listenOnce(ctx, update, msgModel)
		}
	}
}

func (c *Client) listenOnce(ctx context.Context, update tgbotapi.Update, msgModel *messages.Service) {
	if update.Message != nil {
		logger.Info(update.Message.Text, zap.String("user", update.Message.From.UserName))

		ctx, cancel := context.WithTimeout(ctx, time.Second*timeoutSeconds)
		defer cancel()

		err := msgModel.HandleIncomingMessage(ctx, messages.Message{
			Text:   update.Message.Text,
			UserID: update.Message.From.ID,
		})
		if err != nil {
			logger.Error("error processing message:", zap.Error(err))
		}
	}
}
