package messages

import (
	"fmt"
	"max.ks1230/project-base/internal/entity/user"
	"sort"
	"strings"
	"time"
)

const commandParts = 2

func location() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.UTC
	}
	return loc
}

func parseCommand(text string) (cmd, arg string, err error) {
	text = strings.TrimSpace(text)
	split := strings.SplitN(text, " ", commandParts)

	if len(split) == commandParts {
		return split[0], split[1], nil
	}
	if strings.HasPrefix(text, "/") {
		return text, "", nil
	}
	return "", text, nil
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

func groupByCategory(exps []user.ExpenseRecord) []string {
	m := make(map[string]float64)
	for _, exp := range exps {
		m[exp.Category] += exp.Amount
	}
	records := make([]struct {
		string
		float64
	}, 0, len(m))
	total := 0.0
	for cat, am := range m {
		records = append(records, struct {
			string
			float64
		}{cat, am})
		total += am
	}
	sort.Slice(records, func(i, j int) bool {
		return records[i].float64 > records[j].float64
	})
	res := make([]string, 0)
	for _, rec := range records {
		res = append(res, fmt.Sprintf("%s: %.2f", rec.string, rec.float64))
	}
	res = append(res, "", fmt.Sprintf("Total: %.2f", total))
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

func convertExpenseToBase(exp *user.ExpenseRecord, rate float64) {
	exp.Amount /= rate
}

func convertToBase(amount float64, rate float64) float64 {
	return amount / rate
}
