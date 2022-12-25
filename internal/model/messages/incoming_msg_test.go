package messages

import (
	"context"
	"testing"

	"google.golang.org/protobuf/proto"
	apiv1 "max.ks1230/finances-bot/api/kafka"

	"github.com/bradfitz/gomemcache/memcache"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"max.ks1230/finances-bot/internal/entity/currency"
	"max.ks1230/finances-bot/internal/entity/user"
	"max.ks1230/finances-bot/internal/model/messages/mock"
)

func Test_OnStartCommand_ShouldAnswerWithIntroMessage(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	sender.SendMessageMock.
		Expect("Hello! I am FinancesRoute bot 🤖", int64(123)).
		Return(nil)

	storage.SaveUserByIDMock.
		Inspect(func(_ context.Context, userID int64, rec user.Record) {
			assert.Equal(m, int64(123), userID)
			assert.Equal(m, user.Record{}, rec)
		}).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
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
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	sender.SendMessageMock.
		Expect("I don't understand you :(", int64(123)).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
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
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	u := user.Record{}
	u.SetPreferredCurrency("USD")
	storage.
		GetUserByIDMock.
		Inspect(func(_ context.Context, userID int64) {
			assert.Equal(m, int64(123), userID)
		}).
		Return(user.Record{}, nil).
		SaveUserByIDMock.
		Inspect(func(_ context.Context, userID int64, rec user.Record) {
			assert.Equal(m, int64(123), userID)
			assert.Equal(m, u, rec)
		}).
		Return(nil)

	sender.SendMessageMock.
		Expect("Gotcha!", int64(123)).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
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
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	storage.
		GetUserByIDMock.
		Inspect(func(_ context.Context, userID int64) {
			assert.Equal(m, int64(123), userID)
		}).
		Return(user.Record{}, nil).
		SaveUserByIDMock.
		Inspect(func(_ context.Context, userID int64, rec user.Record) {
			assert.Equal(m, int64(123), userID)
			assert.Equal(m, user.Record{MonthLimit: 1000}, rec)
		}).
		Return(nil).
		GetRateMock.
		Inspect(func(_ context.Context, name string) {
			assert.Equal(m, "RUB", name)
		}).
		Return(currency.Rate{BaseRate: 1}, nil)

	sender.SendMessageMock.
		Expect("Gotcha!", int64(123)).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
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
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
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
		Inspect(func(_ context.Context, userID int64) {
			assert.Equal(m, int64(123), userID)
		}).
		Return(user.Record{}, nil).
		GetRateMock.
		Inspect(func(_ context.Context, name string) {
			assert.Equal(m, "RUB", name)
		}).
		Return(currency.Rate{BaseRate: 1}, nil)

	cache.
		InvalidateCacheMock.
		Inspect(func(id int64, options []string) {
			assert.Equal(m, int64(123), id)
			assert.Contains(m, options, "")
			assert.Contains(m, options, "week")
			assert.Contains(m, options, "month")
			assert.Contains(m, options, "year")
		}).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/expense Internet 500",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnReportCommand_ShouldShowGeneratingMessage(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	producerMessage, _ := proto.Marshal(&apiv1.ReportRequest{
		UserID: 123,
		Period: "",
	})

	producer.
		ProduceMessageMock.
		Expect(producerMessage).
		Return(nil)

	cache.
		GetReportMock.
		Expect(int64(123), "").
		Return("", memcache.ErrCacheMiss)

	sender.SendMessageMock.
		Expect("Generating report...", int64(123)).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/report",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnReportCommand_ShouldShowCachedReport(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	cachedReport := "Shopping: 1600.00\nInternet: 1000.00\n\nTotal: 2600.00"
	cache.
		GetReportMock.
		Expect(int64(123), "").
		Return(cachedReport, nil)

	sender.SendMessageMock.
		Expect(cachedReport, int64(123)).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/report",
		UserID: 123,
	})

	assert.NoError(t, err)
}

func Test_OnReportCommand_ShouldShowReportInPreferredCurrency(t *testing.T) {
	t.Skip()
	ctx := context.Background()

	m := minimock.NewController(t)
	defer m.Finish()
	sender := mock.NewMessageSenderMock(m)
	storage := mock.NewUserStorageMock(m)
	cache := mock.NewReportCacheMock(m)
	producer := mock.NewReportRequestProducerMock(m)
	cfg := mock.NewConfigMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	u := user.Record{}
	u.SetPreferredCurrency("USD")
	storage.
		/*GetUserExpensesMock.
		Inspect(func(_ context.Context, userID int64) {
			assert.Equal(m, int64(123), userID)
		}).
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
		}, nil).*/
		GetUserByIDMock.
		Inspect(func(_ context.Context, userID int64) {
			assert.Equal(m, int64(123), userID)
		}).
		Return(u, nil).
		GetRateMock.
		Inspect(func(_ context.Context, name string) {
			assert.Equal(m, "USD", name)
		}).
		Return(currency.Rate{BaseRate: 0.1}, nil)

	expectedReport := "Shopping: 160.00\nInternet: 100.00\n\nTotal: 260.00"
	cache.
		GetReportMock.
		Expect(int64(123), "").
		Return("", memcache.ErrCacheMiss).
		CacheReportMock.
		Expect(int64(123), "", expectedReport).
		Return(nil)

	sender.SendMessageMock.
		Expect(expectedReport, int64(123)).
		Return(nil)

	model := NewService(cfg, sender, storage, cache, producer)
	err := model.HandleIncomingMessage(ctx, Message{
		Text:   "/report",
		UserID: 123,
	})

	assert.NoError(t, err)
}
