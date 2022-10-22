package messages

import (
	"context"
	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/entity/user"
	"max.ks1230/project-base/internal/model/messages/mock"
	"testing"
	"time"
)

func Test_OnStartCommand_ShouldAnswerWithIntroMessage(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	sender.SendMessageMock.
		Expect("Hello! I am FinancesRoute bot 🤖", int64(123)).
		Return(nil)

	storage.SaveUserByIDMock.
		Expect(ctx, 123, user.Record{}).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/start",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnUnknownCommand_ShouldAnswerWithHelpMessage(t *testing.T) {
	ctx := context.Background()

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
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/none",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnCurrencyCommand_ShouldAnswerOkMessage(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	u := user.Record{}
	u.SetPreferredCurrency("USD")
	storage.
		GetUserByIDMock.
		Expect(ctx, 123).
		Return(user.Record{}, nil).
		SaveUserByIDMock.
		Expect(ctx, 123, u).
		Return(nil)

	sender.SendMessageMock.
		Expect("Gotcha!", int64(123)).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/currency USD",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnLimitCommand_ShouldAnswerOkMessage(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	storage.
		GetUserByIDMock.
		Expect(ctx, 123).
		Return(user.Record{}, nil).
		SaveUserByIDMock.
		Expect(ctx, 123, user.Record{MonthLimit: 1000}).
		Return(nil).
		GetRateMock.
		Expect(ctx, "RUB").
		Return(currency.Rate{BaseRate: 1}, nil)

	sender.SendMessageMock.
		Expect("Gotcha!", int64(123)).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/limit 1000",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnExpenseCommand_ShouldAnswerWithOkMessage(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	sender.SendMessageMock.
		Expect("Gotcha!", int64(123)).
		Return(nil)

	storage.
		SaveExpenseMock.
		Inspect(func(_ context.Context, id int64, rec user.ExpenseRecord) {
			assert.Equal(m, int64(123), id)
			assert.Equal(m, float64(500), rec.Amount)
			assert.Equal(m, "Internet", rec.Category)
		}).
		Return(nil).
		GetUserByIDMock.
		Expect(ctx, 123).
		Return(user.Record{}, nil).
		GetRateMock.
		Expect(ctx, "RUB").
		Return(currency.Rate{BaseRate: 1}, nil)

	model := NewService(sender, storage, cfg)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/expense Internet 500",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnReportCommand_ShouldShowReport(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	storage.
		GetUserExpensesMock.
		Expect(ctx, 123).
		Return([]user.ExpenseRecord{
			{
				Amount:   1000,
				Category: "Internet",
				Created:  time.Now(),
			},
			{
				Amount:   1500,
				Category: "Shopping",
				Created:  time.Now(),
			},
			{
				Amount:   100,
				Category: "Shopping",
				Created:  time.Now(),
			},
		}, nil).
		GetUserByIDMock.
		Expect(ctx, 123).
		Return(user.Record{}, nil).
		GetRateMock.
		Expect(ctx, "RUB").
		Return(currency.Rate{BaseRate: 1}, nil)

	sender.SendMessageMock.
		Expect("Shopping: 1600.00\nInternet: 1000.00\n\nTotal: 2600.00", int64(123)).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/report",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnReportCommand_ShouldShowReportInPreferredCurrency(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	u := user.Record{}
	u.SetPreferredCurrency("USD")
	storage.
		GetUserExpensesMock.
		Expect(ctx, 123).
		Return([]user.ExpenseRecord{
			{
				Amount:   1000,
				Category: "Internet",
				Created:  time.Now(),
			},
			{
				Amount:   1500,
				Category: "Shopping",
				Created:  time.Now(),
			},
			{
				Amount:   100,
				Category: "Shopping",
				Created:  time.Now(),
			},
		}, nil).
		GetUserByIDMock.
		Expect(ctx, 123).
		Return(u, nil).
		GetRateMock.
		Expect(ctx, "USD").
		Return(currency.Rate{BaseRate: 0.1}, nil)

	sender.SendMessageMock.
		Expect("Shopping: 160.00\nInternet: 100.00\n\nTotal: 260.00", int64(123)).
		Return(nil)

	model := NewService(sender, storage, cfg)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/report",
		UserID: 123,
	})

	assert.NoError(t, err)
}
