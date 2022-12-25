package reports

import (
	"context"
	"testing"
	"time"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/assert"
	"max.ks1230/finances-bot/internal/entity/currency"
	"max.ks1230/finances-bot/internal/entity/user"
	"max.ks1230/finances-bot/internal/model/reports/mock"
)

func Test_OnGenerateReport_ShouldReturnReportInUSD(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	cfg := mock.NewConfigMock(m)
	storage := mock.NewExpensesStorageMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	u := user.Record{}
	u.SetPreferredCurrency("USD")
	storage.
		GetUserExpensesMock.
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
		}, nil).
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

	generator := NewGenerator(cfg, storage)
	report, err := generator.GenerateReport(ctx, 123, "")
	assert.NoError(m, err)
	assert.Equal(m, true, report.GetStatus().GetSuccess())
	assert.Equal(m, 260.0, report.GetTotalAmount())
	assert.Equal(m, "Shopping", report.GetRecords()[0].GetCategory())
	assert.Equal(m, 160.0, report.GetRecords()[0].GetAmount())
	assert.Equal(m, "Internet", report.GetRecords()[1].GetCategory())
	assert.Equal(m, 100.0, report.GetRecords()[1].GetAmount())
}

func Test_OnGenerateReport_ShouldReturnReportInRUB(t *testing.T) {
	ctx := context.Background()

	m := minimock.NewController(t)
	cfg := mock.NewConfigMock(m)
	storage := mock.NewExpensesStorageMock(m)

	cfg.BaseCurrencyMock.Return("RUB")

	u := user.Record{}
	storage.
		GetUserExpensesMock.
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
		}, nil).
		GetUserByIDMock.
		Inspect(func(_ context.Context, userID int64) {
			assert.Equal(m, int64(123), userID)
		}).
		Return(u, nil).
		GetRateMock.
		Inspect(func(_ context.Context, name string) {
			assert.Equal(m, "RUB", name)
		}).
		Return(currency.Rate{BaseRate: 1}, nil)

	generator := NewGenerator(cfg, storage)
	report, err := generator.GenerateReport(ctx, 123, "")
	assert.NoError(m, err)
	assert.Equal(m, true, report.GetStatus().GetSuccess())
	assert.Equal(m, 2600.0, report.GetTotalAmount())
	assert.Equal(m, "Shopping", report.GetRecords()[0].GetCategory())
	assert.Equal(m, 1600.0, report.GetRecords()[0].GetAmount())
	assert.Equal(m, "Internet", report.GetRecords()[1].GetCategory())
	assert.Equal(m, 1000.0, report.GetRecords()[1].GetAmount())
}
