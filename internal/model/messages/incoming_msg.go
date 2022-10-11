package messages

type messageSender interface {
	SendMessage(text string, userID int64) error
}

type MessageHandler interface {
	HandleMessage(text string, userID int64) (string, error)
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

func (s *Service) IncomingMessage(msg Message) error {
	resp, err := s.handler.HandleMessage(msg.Text, msg.UserID)
	if err != nil {
		_ = s.tgClient.SendMessage("Sorry, something weird happened...\n"+resp, msg.UserID)
		return err
	}
	return s.tgClient.SendMessage(resp, msg.UserID)
}
