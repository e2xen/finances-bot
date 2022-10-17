package messages

import (
	"github.com/gojuno/minimock/v3"
	"testing"

	"github.com/stretchr/testify/assert"
	"max.ks1230/project-base/internal/model/messages/mock"
)

func Test_OnStartCommand_ShouldAnswerWithIntroMessage(t *testing.T) {
	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	sender.SendMessageMock.
		Expect("Hello! I am FinancesRoute bot 🤖", int64(123)).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.IncomingMessage(Message{
		Text:   "/start",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnUnknownCommand_ShouldAnswerWithHelpMessage(t *testing.T) {
	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	sender.SendMessageMock.
		Expect("I don't understand you :(", int64(123)).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.IncomingMessage(Message{
		Text:   "/none",
		UserID: 123,
	})

	assert.NoError(t, err)
}
