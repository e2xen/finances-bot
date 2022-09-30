package handlers

import (
	"fmt"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/model/storage"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Handler func(arg string, user int64) (string, error)

type Map map[string]Handler

type Service struct {
	handlersMap Map
	storage     storage.UserStorage
}

func New(userStorage *storage.UserStorage) *Service {
	res := &Service{
		handlersMap: nil,
		storage:     *userStorage,
	}
	res.handlersMap = newMap(res)
	return res
}

func (s *Service) HandleMessage(text string, userID int64) (string, error) {
	cmd, arg, err := parseCommand(text)
	if err != nil {
		return "", errors.Wrap(err, "handle message")
	}

	handler, ok := s.handlersMap[cmd]
	if ok {
		resp, err := handler(arg, userID)
		return resp, err
	}
	return "I don't understand you :(", nil
}

func parseCommand(text string) (cmd, arg string, err error) {
	text = strings.TrimSpace(text)
	split := strings.SplitN(text, " ", 2)
	if len(split) < 2 {
		if strings.HasPrefix(text, "/") {
			return text, "", nil
		} else {
			return "", text, nil
		}
	}
	return split[0], split[1], nil
}

func newMap(s *Service) Map {
	m := make(Map)
	m["/start"] = s.handleStart
	m["/expense"] = s.handleExpense
	m["/report"] = s.handleReport

	m[""] = s.handleNoCommand

	return m
}

func (s *Service) handleStart(arg string, user int64) (string, error) {
	return "Hello! I am FinancesRoute bot 🤖", nil
}

const dateLayout = "02.01.2006"

var loc, _ = time.LoadLocation("Europe/Moscow")

func (s *Service) handleExpense(arg string, user int64) (string, error) {
	args := strings.Fields(arg)
	if len(args) < 2 {
		return "That is an incorrect command usage", nil
	}
	amount, err := strconv.ParseFloat(args[1], 32)
	if err != nil || amount <= 0 {
		return "Your expense amount is incorrect", errors.Wrap(err, "handle expense")
	}
	category, date := args[0], time.Now()
	if len(args) > 2 {
		date, err = time.ParseInLocation(dateLayout, args[2], loc)
		if err != nil {
			return "The date is incorrect. Should be dd.mm.yyyy", errors.Wrap(err, "handle expense")
		}
	}

	expense := storage.ExpenseRecord{
		Amount:   amount,
		Category: category,
		Created:  date,
	}

	userRec, _ := s.storage.GetById(user)
	userRec.Expenses = append(userRec.Expenses, expense)
	err = s.storage.SaveById(user, userRec)
	if err != nil {
		return "Can't save your expense atm. Try later", errors.Wrap(err, "handle expense")
	}
	return "Gotcha!", nil
}

func (s *Service) handleReport(arg string, user int64) (string, error) {
	userRec, err := s.storage.GetById(user)
	if err != nil {
		return "Can't load your expenses atm. Try later", errors.Wrap(err, "handle report")
	}

	expenses := userRec.Expenses
	if len(expenses) == 0 {
		return "You have no expenses yet", nil
	}

	switch arg {
	case "week":
		expenses = filterExpensesAfter(expenses, time.Now().AddDate(0, 0, -7))
	case "month":
		expenses = filterExpensesAfter(expenses, time.Now().AddDate(0, -1, 0))
	case "year":
		expenses = filterExpensesAfter(expenses, time.Now().AddDate(-1, 0, 0))
	}

	reportExpenses := groupByCategory(expenses)
	return strings.Join(reportExpenses, "\n"), nil
}

func filterExpensesAfter(exps []storage.ExpenseRecord, after time.Time) []storage.ExpenseRecord {
	res := make([]storage.ExpenseRecord, 0)
	for _, exp := range exps {
		if after.Before(exp.Created) {
			res = append(res, exp)
		}
	}
	return res
}

func groupByCategory(exps []storage.ExpenseRecord) []string {
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
	res := make([]string, 0, len(records)+2)
	for _, rec := range records {
		res = append(res, fmt.Sprintf("%s: %.2f", rec.string, rec.float64))
	}
	res = append(res, "")
	res = append(res, fmt.Sprintf("Total: %.2f", total))
	return res
}

func (s *Service) handleNoCommand(arg string, user int64) (string, error) {
	return "I would love to talk about it more!", nil
}
