package messages

import (
	"fmt"
	"github.com/jinzhu/now"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/entity/user"
	"max.ks1230/project-base/internal/model/customerr"
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
	helloFailedMessage    = "Haven't heard you. Please try /start one more time"
	loveToTalkMessage     = "I would love to talk about it more!"
	okMessage             = "Gotcha!"
	noExpensesMessage     = "You have no expenses yet"

	incorrectUsageMessage    = "That is an incorrect command usage"
	incorrectExpenseMessage  = "Your expense amount is incorrect"
	incorrectLimitMessage    = "Your limit amount is incorrect"
	incorrectDateMessage     = "The date is incorrect. Should be dd.mm.yyyy"
	cannotGetExpensesMessage = "Can't get your expenses atm. Try later"
	cannotSaveExpenseMessage = "Can't save your expense atm. Try later"
	cannotSetCurrencyMessage = "Can't set your preferred currency atm. Try later"
	cannotSetLimitMessage    = "Can't set your month limit atm. Try later"
	cannotGetRateMessage     = "Can't get currencies rates atm. Try later"
	limitExceededMessage     = "You exceeded your limit and I'm not writing that down! Congrats!"
	invalidCurrencyTemplate  = "I don't know that currency. Try one of: %s"
)

const (
	startCmd    = "/start"
	expenseCmd  = "/expense"
	reportCmd   = "/report"
	currencyCmd = "/currency"
	limitCmd    = "/limit"
)

type userStorage interface {
	GetUserByID(userID int64) (user.Record, error)
	SaveUserByID(userID int64, rec user.Record) error
	GetRate(name string) (currency.Rate, error)
	SaveExpense(userID int64, record user.ExpenseRecord) error
	GetUserExpenses(userID int64) ([]user.ExpenseRecord, error)
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
	m[limitCmd] = s.handleLimit

	m[""] = s.handleNoCommand

	return m
}

func (s *HandlerService) handleStart(_ string, userID int64) (string, error) {
	err := s.storage.SaveUserByID(userID, user.Record{})
	if err != nil {
		return helloFailedMessage, errors.Wrap(err, "handle start")
	}
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

	rate, err := s.storage.GetRate(userRec.PreferredCurrencyOrDefault(s.defaultCurrency))
	if err != nil {
		return cannotGetRateMessage, errors.Wrap(err, "handle expense")
	}

	convertExpenseToBase(&expense, rate.BaseRate)
	err = s.storage.SaveExpense(userID, expense)
	if err != nil {
		var limErr *customerr.LimitError
		if errors.As(err, &limErr) {
			return limitExceededMessage, err
		}
		return cannotSaveExpenseMessage, errors.Wrap(err, "handle expense")
	}
	return okMessage, nil
}

func (s *HandlerService) handleReport(arg string, userID int64) (string, error) {
	userRec, err := s.storage.GetUserByID(userID)
	if err != nil {
		return cannotGetExpensesMessage, errors.Wrap(err, "handle report")
	}

	expenses, err := s.storage.GetUserExpenses(userID)
	if err != nil {
		return cannotGetExpensesMessage, errors.Wrap(err, "handle report")
	}
	if len(expenses) == 0 {
		return noExpensesMessage, nil
	}

	switch arg {
	case "week":
		expenses = filterExpensesAfter(expenses, now.BeginningOfWeek())
	case "month":
		expenses = filterExpensesAfter(expenses, now.BeginningOfMonth())
	case "year":
		expenses = filterExpensesAfter(expenses, now.BeginningOfYear())
	}

	rate, err := s.storage.GetRate(userRec.PreferredCurrencyOrDefault(s.defaultCurrency))
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

	u, err := s.storage.GetUserByID(userID)
	if err != nil {
		return cannotSetCurrencyMessage, errors.Wrap(err, "handle currency")
	}
	u.SetPreferredCurrency(curr)
	err = s.storage.SaveUserByID(userID, u)
	if err != nil {
		return cannotSetCurrencyMessage, errors.Wrap(err, "handle currency")
	}

	return okMessage, nil
}

func (s *HandlerService) handleLimit(arg string, userID int64) (string, error) {
	limit, err := strconv.ParseFloat(arg, floatBitSize)
	if err != nil {
		return incorrectLimitMessage, errors.Wrap(err, "handle limit")
	}

	u, err := s.storage.GetUserByID(userID)
	if err != nil {
		return cannotSetLimitMessage, errors.Wrap(err, "handle limit")
	}
	rate, err := s.storage.GetRate(u.PreferredCurrencyOrDefault(s.defaultCurrency))
	if err != nil {
		return cannotGetRateMessage, errors.Wrap(err, "handle limit")
	}
	u.MonthLimit = convertToBase(limit, rate.BaseRate)
	if err = s.storage.SaveUserByID(userID, u); err != nil {
		return cannotSetLimitMessage, errors.Wrap(err, "handle limit")
	}

	return okMessage, nil
}

func (s *HandlerService) handleNoCommand(_ string, _ int64) (string, error) {
	return loveToTalkMessage, nil
}
