package messages

import (
	"fmt"
	"strings"
	"time"

	apiv1 "max.ks1230/finances-bot/api/grpc"

	"max.ks1230/finances-bot/internal/entity/user"
)

const commandParts = 2

func location() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.UTC
	}
	return loc
}

func parseCommand(text string) (cmd, arg string) {
	text = strings.TrimSpace(text)
	split := strings.SplitN(text, " ", commandParts)

	if len(split) == commandParts {
		return split[0], split[1]
	}
	if strings.HasPrefix(text, "/") {
		return text, ""
	}
	return "", text
}

func convertExpenseToBase(exp *user.ExpenseRecord, rate float64) {
	exp.Amount /= rate
}

func convertToBase(amount float64, rate float64) float64 {
	return amount / rate
}

func formatReport(report *apiv1.ReportResult) string {
	res := make([]string, 0)
	for _, rec := range report.GetRecords() {
		res = append(res, fmt.Sprintf("%s: %.2f", rec.GetCategory(), rec.GetAmount()))
	}
	res = append(res, "", fmt.Sprintf("Total: %.2f", report.GetTotalAmount()))
	return strings.Join(res, "\n")
}
