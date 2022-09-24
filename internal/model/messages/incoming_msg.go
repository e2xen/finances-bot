package messages

import (
	"max.ks1230/project-base/internal/model/messages/handlers"
	"max.ks1230/project-base/internal/model/storage"
)

type MessageSender interface {
	SendMessage(text string, userID int64) error
}

type MessageHandler interface {
	HandleMessage(text string, userID int64) (string, error)
}

type Service struct {
	tgClient MessageSender
	handler  MessageHandler
}

func New(tgClient MessageSender, storage *storage.UserStorage) *Service {
	return &Service{
		tgClient: tgClient,
		handler:  handlers.New(storage),
	}
}

type Message struct {
	Text   string
	UserID int64
}

func (s *Service) IncomingMessage(msg Message) error {
	resp, err := s.handler.HandleMessage(msg.Text, msg.UserID)
	if err != nil {
		_ = s.tgClient.SendMessage("Sorry, something weird happened...\n"+resp, msg.UserID)
		return err
	}
	return s.tgClient.SendMessage(resp, msg.UserID)
}
