package messages

import (
	"fmt"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/model/user"
	"sort"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "02.01.2006"

const (
	dontUnderstandMessage = "I don't understand you :("
	helloMessage          = "Hello! I am FinancesRoute bot 🤖"
	loveToTalkMessage     = "I would love to talk about it more!"
	okMessage             = "Gotcha!"
	noExpensesMessage     = "You have no expenses yet"

	incorrectUsageMessage    = "That is an incorrect command usage"
	incorrectExpenseMessage  = "Your expense amount is incorrect"
	incorrectDateMessage     = "The date is incorrect. Should be dd.mm.yyyy"
	cannotGetExpensesMessage = "Can't get your expenses atm. Try later"
	cannotSaveExpenseMessage = "Can't save your expense atm. Try later"
)

const (
	startCommand   = "/start"
	expenseCommand = "/expense"
	reportCommand  = "/report"
)

type userStorage interface {
	GetByID(userID int64) (user.Record, error)
	SaveByID(userID int64, rec user.Record) error
}

type handler func(arg string, user int64) (string, error)

type handlerMap map[string]handler

type HandlerService struct {
	handlersMap handlerMap
	storage     userStorage
}

func newHandler(userStorage userStorage) *HandlerService {
	res := &HandlerService{
		handlersMap: nil,
		storage:     userStorage,
	}
	res.handlersMap = newMap(res)
	return res
}

func (s *HandlerService) HandleMessage(text string, userID int64) (string, error) {
	cmd, arg, err := parseCommand(text)
	if err != nil {
		return "", errors.Wrap(err, "handle message")
	}

	handler, ok := s.handlersMap[cmd]
	if ok {
		return handler(arg, userID)
	}
	return dontUnderstandMessage, nil
}

func parseCommand(text string) (cmd, arg string, err error) {
	text = strings.TrimSpace(text)
	split := strings.SplitN(text, " ", 2)

	if len(split) == 2 {
		return split[0], split[1], nil
	}
	if strings.HasPrefix(text, "/") {
		return text, "", nil
	}
	return "", text, nil
}

func newMap(s *HandlerService) handlerMap {
	m := make(handlerMap)
	m[startCommand] = s.handleStart
	m[expenseCommand] = s.handleExpense
	m[reportCommand] = s.handleReport

	m[""] = s.handleNoCommand

	return m
}

func (s *HandlerService) handleStart(_ string, _ int64) (string, error) {
	return helloMessage, nil
}

func location() *time.Location {
	loc, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		return time.UTC
	}
	return loc
}

func (s *HandlerService) handleExpense(arg string, userID int64) (string, error) {
	args := strings.Fields(arg)
	if len(args) < 2 {
		return incorrectUsageMessage, nil
	}
	amount, err := strconv.ParseFloat(args[1], 32)
	if err != nil || amount <= 0 {
		return incorrectExpenseMessage, errors.Wrap(err, "handle expense")
	}
	category, date := args[0], time.Now()
	if len(args) > 2 {
		date, err = time.ParseInLocation(dateLayout, args[2], location())
		if err != nil {
			return incorrectDateMessage, errors.Wrap(err, "handle expense")
		}
	}

	expense := user.ExpenseRecord{
		Amount:   amount,
		Category: category,
		Created:  date,
	}

	userRec, err := s.storage.GetByID(userID)
	if err != nil {
		return cannotGetExpensesMessage, errors.Wrap(err, "handle expense")
	}
	userRec.Expenses = append(userRec.Expenses, expense)
	err = s.storage.SaveByID(userID, userRec)
	if err != nil {
		return cannotSaveExpenseMessage, errors.Wrap(err, "handle expense")
	}
	return okMessage, nil
}

func (s *HandlerService) handleReport(arg string, user int64) (string, error) {
	userRec, err := s.storage.GetByID(user)
	if err != nil {
		return cannotGetExpensesMessage, errors.Wrap(err, "handle report")
	}

	expenses := userRec.Expenses
	if len(expenses) == 0 {
		return noExpensesMessage, nil
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
	res := make([]string, 0, len(records)+2)
	for _, rec := range records {
		res = append(res, fmt.Sprintf("%s: %.2f", rec.string, rec.float64))
	}
	res = append(res, "", fmt.Sprintf("Total: %.2f", total))
	return res
}

func (s *HandlerService) handleNoCommand(_ string, _ int64) (string, error) {
	return loveToTalkMessage, nil
}
