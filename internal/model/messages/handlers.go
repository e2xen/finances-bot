package messages

import (
	"fmt"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/entity/user"
	"max.ks1230/project-base/internal/utils"
	"strconv"
	"strings"
	"time"
)

const dateLayout = "02.01.2006"
const floatBitSize = 32

const (
	expenseCmdParts = 2
)

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
	cannotSetCurrencyMessage = "Can't set your preferred currency atm. Try later"
	cannotGetRateMessage     = "Can't get currencies rates atm. Try later"
	invalidCurrencyTemplate  = "I don't know that currency. Try one of: %s"
)

const (
	startCmd    = "/start"
	expenseCmd  = "/expense"
	reportCmd   = "/report"
	currencyCmd = "/currency"
)

type userStorage interface {
	GetUserByID(userID int64) (user.Record, error)
	SaveUserByID(userID int64, rec user.Record) error
	SetCurrencyForUser(userID int64, curr string) error
	GetRate(name string) (currency.Rate, error)
}

type config interface {
	BaseCurrency() string
}

type handler func(arg string, user int64) (string, error)

type handlerMap map[string]handler

type HandlerService struct {
	handlersMap     handlerMap
	storage         userStorage
	defaultCurrency string
}

func newHandler(userStorage userStorage, config config) *HandlerService {
	res := &HandlerService{
		handlersMap:     nil,
		storage:         userStorage,
		defaultCurrency: config.BaseCurrency(),
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

func newMap(s *HandlerService) handlerMap {
	m := make(handlerMap)
	m[startCmd] = s.handleStart
	m[expenseCmd] = s.handleExpense
	m[reportCmd] = s.handleReport
	m[currencyCmd] = s.handleCurrency

	m[""] = s.handleNoCommand

	return m
}

func (s *HandlerService) handleStart(_ string, _ int64) (string, error) {
	return helloMessage, nil
}

func (s *HandlerService) handleExpense(arg string, userID int64) (string, error) {
	args := strings.Fields(arg)
	if len(args) < expenseCmdParts {
		return incorrectUsageMessage, nil
	}
	amount, err := strconv.ParseFloat(args[1], floatBitSize)
	if err != nil || amount <= 0 {
		return incorrectExpenseMessage, errors.Wrap(err, "handle expense")
	}
	category, date := args[0], time.Now()
	if len(args) > expenseCmdParts {
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

	userRec, err := s.storage.GetUserByID(userID)
	if err != nil {
		return cannotGetExpensesMessage, errors.Wrap(err, "handle expense")
	}

	rate, err := s.storage.GetRate(userRec.PreferredCurrency(s.defaultCurrency))
	if err != nil {
		return cannotGetRateMessage, errors.Wrap(err, "handle report")
	}

	convertExpenseToBase(&expense, rate.BaseRate)
	userRec.Expenses = append(userRec.Expenses, expense)
	err = s.storage.SaveUserByID(userID, userRec)
	if err != nil {
		return cannotSaveExpenseMessage, errors.Wrap(err, "handle expense")
	}
	return okMessage, nil
}

func (s *HandlerService) handleReport(arg string, user int64) (string, error) {
	userRec, err := s.storage.GetUserByID(user)
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

	rate, err := s.storage.GetRate(userRec.PreferredCurrency(s.defaultCurrency))
	if err != nil {
		return cannotGetRateMessage, errors.Wrap(err, "handle report")
	}
	expenses = convertExpensesFromBase(expenses, rate.BaseRate)

	reportExpenses := groupByCategory(expenses)
	return strings.Join(reportExpenses, "\n"), nil
}

func (s *HandlerService) handleCurrency(curr string, userID int64) (string, error) {
	if !utils.Contains(currency.Currencies, curr) {
		return fmt.Sprintf(invalidCurrencyTemplate, strings.Join(currency.Currencies, ", ")),
			errors.New("handle currency")
	}

	err := s.storage.SetCurrencyForUser(userID, curr)
	if err != nil {
		return cannotSetCurrencyMessage, errors.Wrap(err, "handle currency")
	}

	return okMessage, nil
}

func (s *HandlerService) handleNoCommand(_ string, _ int64) (string, error) {
	return loveToTalkMessage, nil
}
