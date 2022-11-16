package reports

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
	"max.ks1230/project-base/internal/logger"

	apiv1 "max.ks1230/project-base/api/grpc"

	"github.com/jinzhu/now"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/entity/currency"

	"max.ks1230/project-base/internal/entity/user"
)

var reportFilters = map[string]time.Time{
	"":      {},
	"week":  now.BeginningOfWeek(),
	"month": now.BeginningOfMonth(),
	"year":  now.BeginningOfYear(),
}

type expensesStorage interface {
	GetUserExpenses(ctx context.Context, userID int64) ([]user.ExpenseRecord, error)
	GetUserByID(ctx context.Context, userID int64) (user.Record, error)
	GetRate(ctx context.Context, name string) (currency.Rate, error)
}

type Generator struct {
	storage         expensesStorage
	defaultCurrency string
}

type config interface {
	BaseCurrency() string
}

func NewGenerator(config config, storage expensesStorage) *Generator {
	return &Generator{
		storage:         storage,
		defaultCurrency: config.BaseCurrency(),
	}
}

func (g *Generator) GenerateReport(ctx context.Context, userID int64, period string) (report *apiv1.ReportResult, err error) {
	logger.Info("GenerateReport - start", zap.Int64("userID", userID), zap.String("period", period))
	defer logger.Info("GenerateReport - end")

	defer func() {
		if report == nil {
			report = &apiv1.ReportResult{}
		}
		if err == nil {
			report.Status = &apiv1.OperationStatus{Success: true}
		} else {
			errMsg := err.Error()
			report.Status = &apiv1.OperationStatus{Success: false, Error: &errMsg}
		}
		report.UserID = userID
		report.Period = period
	}()

	userRec, err := g.storage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "generate report")
	}

	expenses, err := g.storage.GetUserExpenses(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "generate report")
	}
	if len(expenses) == 0 {
		return nil, nil
	}

	filter, ok := reportFilters[period]
	if !ok {
		return nil, errors.Wrap(
			fmt.Errorf("report period %s is not supported", period),
			"generate report",
		)
	}
	expenses = filterExpensesAfter(expenses, filter)

	rate, err := g.storage.GetRate(ctx, userRec.PreferredCurrencyOrDefault(g.defaultCurrency))
	if err != nil {
		return nil, errors.Wrap(err, "generate report")
	}
	expenses = convertExpensesFromBase(expenses, rate.BaseRate)

	report = groupExpenses(expenses)
	return report, nil
}

func filterExpensesAfter(exps []user.ExpenseRecord, after time.Time) []user.ExpenseRecord {
	res := make([]user.ExpenseRecord, 0)
	for _, exp := range exps {
		if after.Before(exp.Created) {
			res = append(res, exp)
		}
	}
	return res
}

func convertExpensesFromBase(expenses []user.ExpenseRecord, rate float64) (result []user.ExpenseRecord) {
	result = make([]user.ExpenseRecord, 0, len(expenses))
	for _, exp := range expenses {
		exp.Amount *= rate
		result = append(result, exp)
	}
	return
}

func groupExpenses(exps []user.ExpenseRecord) *apiv1.ReportResult {
	m := make(map[string]float64)
	for _, exp := range exps {
		m[exp.Category] += exp.Amount
	}
	records := make([]*apiv1.ReportRecord, 0)
	total := 0.0
	for cat, am := range m {
		records = append(records, &apiv1.ReportRecord{Category: cat, Amount: am})
		total += am
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].Amount > records[j].Amount
	})
	return &apiv1.ReportResult{
		Records:     records,
		TotalAmount: total,
	}
}

func ReportPeriods() []string {
	res := make([]string, 0, len(reportFilters))
	for k := range reportFilters {
		res = append(res, k)
	}
	return res
}
