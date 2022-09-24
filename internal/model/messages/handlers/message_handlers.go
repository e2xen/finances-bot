package handlers

import (
	"fmt"
	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/model/storage"
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
		if err != nil {
			return "", err
		}
		return resp, nil
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

func (s *Service) handleExpense(arg string, user int64) (string, error) {
	args := strings.Fields(arg)
	if len(args) < 2 {
		return "That is an incorrect command usage", nil
	}
	amount, err := strconv.ParseFloat(args[1], 32)
	if err != nil {
		return "Your expense amount is incorrect", errors.Wrap(err, "handle expense")
	}
	category, date := args[0], time.Now()
	if len(args) > 2 {
		date, _ = time.ParseInLocation("", args[2], nil)
	}

	expense := storage.ExpenseRecord{
		Amount:   amount,
		Category: category,
		Created:  date,
	}

	userRec, _ := s.storage.GetById(user)
	userRec.Expenses = append(userRec.Expenses, expense)
	_ = s.storage.SaveById(user, userRec)
	return "Gotcha!", nil
}

func (s *Service) handleReport(arg string, user int64) (string, error) {
	userRec, _ := s.storage.GetById(user)
	expenses := userRec.Expenses
	if len(expenses) == 0 {
		return "You have no expenses yet", nil
	}
	expensesSlice := make([]string, 0, len(expenses))
	for _, exp := range expenses {
		expensesSlice = append(expensesSlice, fmt.Sprint(exp))
	}
	return strings.Join(expensesSlice, "\n"), nil
}

func (s *Service) handleNoCommand(arg string, user int64) (string, error) {
	return "I would love to talk about it more!", nil
}
