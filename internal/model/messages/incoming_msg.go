package messages

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

type messageSender interface {
	SendMessage(text string, userID int64) error
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, text string, userID int64) (string, error)
}

type Service struct {
	tgClient messageSender
	handler  MessageHandler
}

func NewService(tgClient messageSender, storage userStorage, config config) *Service {
	return &Service{
		tgClient: tgClient,
		handler:  newHandler(storage, config),
	}
}

type Message struct {
	Text   string
	UserID int64
}

func (s *Service) HandleIncomingMessage(ctx context.Context, msg Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handleMessage")
	defer span.Finish()

	start := time.Now()
	err := s.handle(ctx, msg)
	elapsed := time.Since(start)

	observeResponse(elapsed, err != nil)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return err
}

func (s *Service) handle(ctx context.Context, msg Message) error {
	resp, err := s.handler.HandleMessage(ctx, msg.Text, msg.UserID)
	if err != nil {
		_ = s.tgClient.SendMessage("Sorry, something wrong happened...\n"+resp, msg.UserID)
		return err
	}
	return s.tgClient.SendMessage(resp, msg.UserID)
}
