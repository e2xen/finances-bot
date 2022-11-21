package messages

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	apiv12 "max.ks1230/project-base/api/grpc"
	"max.ks1230/project-base/internal/model/reports"

	"google.golang.org/protobuf/proto"
	apiv1 "max.ks1230/project-base/api/kafka"

	"github.com/opentracing/opentracing-go"

	"go.uber.org/zap"
	"max.ks1230/project-base/internal/logger"

	"github.com/pkg/errors"
	"max.ks1230/project-base/internal/entity/currency"
	"max.ks1230/project-base/internal/entity/user"
	"max.ks1230/project-base/internal/model/customerr"
	"max.ks1230/project-base/internal/utils"
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
	generatingReport      = "Generating report..."

	incorrectUsageMessage    = "That is an incorrect command usage"
	incorrectExpenseMessage  = "Your expense amount is incorrect"
	incorrectLimitMessage    = "Your limit amount is incorrect"
	incorrectDateMessage     = "The date is incorrect. Should be dd.mm.yyyy"
	cannotGetExpensesMessage = "Can't get your expenses atm. Try later"
	cannotSaveExpenseMessage = "Can't save your expense atm. Try later"
	cannotSetCurrencyMessage = "Can't set your preferred currency atm. Try later"
	cannotSetLimitMessage    = "Can't set your month limit atm. Try later"
	cannotGetRateMessage     = "Can't get currencies rates atm. Try later"
	cannotGenReportMessage   = "Can't generate report atm. Try later"
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

type reportRequestProducer interface {
	ProduceMessage(message []byte) error
}

type userStorage interface {
	GetUserByID(ctx context.Context, userID int64) (user.Record, error)
	SaveUserByID(ctx context.Context, userID int64, rec user.Record) error
	GetRate(ctx context.Context, name string) (currency.Rate, error)
	SaveExpense(ctx context.Context, userID int64, record user.ExpenseRecord) error
}

type reportCache interface {
	CacheReport(userID int64, option string, report string) error
	GetReport(userID int64, option string) (string, error)
	InvalidateCache(userID int64, options []string) error
}

type config interface {
	BaseCurrency() string
}

type handler func(ctx context.Context, arg string, user int64) (string, error)

type handlerMap map[string]handler

type HandlerService struct {
	handlersMap     handlerMap
	storage         userStorage
	cache           reportCache
	producer        reportRequestProducer
	defaultCurrency string
}

func newHandler(config config,
	userStorage userStorage,
	cache reportCache,
	producer reportRequestProducer) *HandlerService {
	res := &HandlerService{
		handlersMap:     nil,
		storage:         userStorage,
		cache:           cache,
		producer:        producer,
		defaultCurrency: config.BaseCurrency(),
	}
	res.handlersMap = newMap(res)
	return res
}

func (s *HandlerService) HandleMessage(ctx context.Context, text string, userID int64) (string, error) {
	cmd, arg := parseCommand(text)

	span, ctx := opentracing.StartSpanFromContext(ctx, "handleCommand")
	defer span.Finish()
	span.SetTag("cmd", cmd)

	handler, ok := s.handlersMap[cmd]
	if ok {
		return handler(ctx, arg, userID)
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

func (s *HandlerService) handleStart(ctx context.Context, _ string, userID int64) (string, error) {
	logger.Info("handleStart - start", zap.Int64("userID", userID))
	defer logger.Info("handleStart - end")

	err := s.storage.SaveUserByID(ctx, userID, user.Record{})
	if err != nil {
		return helloFailedMessage, errors.Wrap(err, "handle start")
	}
	return helloMessage, nil
}

func (s *HandlerService) handleExpense(ctx context.Context, arg string, userID int64) (res string, err error) {
	logger.Info("handleExpense - start", zap.Int64("userID", userID), zap.String("arg", arg))
	defer logger.Info("handleExpense - end")

	defer func() {
		// invalidate cache in case of success
		if err == nil {
			opts := reports.ReportPeriods()
			cacheErr := s.cache.InvalidateCache(userID, opts)
			if cacheErr != nil {
				logger.Error("failed to invalidate cache", zap.Error(cacheErr))
			}
		}
	}()

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

	userRec, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		return cannotGetExpensesMessage, errors.Wrap(err, "handle expense")
	}

	rate, err := s.storage.GetRate(ctx, userRec.PreferredCurrencyOrDefault(s.defaultCurrency))
	if err != nil {
		return cannotGetRateMessage, errors.Wrap(err, "handle expense")
	}

	convertExpenseToBase(&expense, rate.BaseRate)
	err = s.storage.SaveExpense(ctx, userID, expense)
	if err != nil {
		var limErr *customerr.LimitError
		if errors.As(err, &limErr) {
			return limitExceededMessage, err
		}
		return cannotSaveExpenseMessage, errors.Wrap(err, "handle expense")
	}
	return okMessage, nil
}

func (s *HandlerService) handleReport(_ context.Context, arg string, userID int64) (result string, err error) {
	logger.Info("handleReport - start", zap.Int64("userID", userID), zap.String("arg", arg))
	defer logger.Info("handleReport - end")

	report, err := s.cache.GetReport(userID, arg)
	if err == nil {
		return report, nil
	}

	logger.Info(
		"failed to get report from cache, request generation",
		zap.Int64("userID", userID),
		zap.String("arg", arg),
		zap.NamedError("cacheErr", err),
	)

	req, err := proto.Marshal(&apiv1.ReportRequest{
		UserID: userID,
		Period: arg,
	})
	if err != nil {
		return cannotGenReportMessage, errors.Wrap(err, "handle report")
	}

	err = s.producer.ProduceMessage(req)
	if err != nil {
		return cannotGenReportMessage, errors.Wrap(err, "handle report")
	}

	return generatingReport, nil
}

func (s *HandlerService) AcceptReport(_ context.Context, report *apiv12.ReportResult) (result string, err error) {
	logger.Info("acceptReport - start", zap.Int64("userID", report.GetUserID()))
	defer logger.Info("acceptReport - end")

	defer func() {
		// cache report in case of successful generation
		if err == nil {
			cacheErr := s.cache.CacheReport(report.GetUserID(), report.GetPeriod(), result)
			if cacheErr != nil {
				logger.Error("error caching report", zap.Error(cacheErr))
			}
		}
	}()

	if !report.GetStatus().GetSuccess() {
		return cannotGenReportMessage, errors.Wrap(errors.New((*report).GetStatus().GetError()), "accept report")
	}

	if len(report.GetRecords()) == 0 {
		return noExpensesMessage, nil
	}

	return formatReport(report), nil
}

func (s *HandlerService) handleCurrency(ctx context.Context, arg string, userID int64) (string, error) {
	logger.Info("handleCurrency - start", zap.Int64("userID", userID), zap.String("arg", arg))
	defer logger.Info("handleCurrency - end")

	curr := arg
	if !utils.Contains(currency.Currencies, curr) {
		return fmt.Sprintf(invalidCurrencyTemplate, strings.Join(currency.Currencies, ", ")),
			errors.New("handle currency")
	}

	u, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		return cannotSetCurrencyMessage, errors.Wrap(err, "handle currency")
	}
	u.SetPreferredCurrency(curr)
	err = s.storage.SaveUserByID(ctx, userID, u)
	if err != nil {
		return cannotSetCurrencyMessage, errors.Wrap(err, "handle currency")
	}

	return okMessage, nil
}

func (s *HandlerService) handleLimit(ctx context.Context, arg string, userID int64) (string, error) {
	logger.Info("handleLimit - start", zap.Int64("userID", userID), zap.String("arg", arg))
	defer logger.Info("handleLimit - end")

	limit, err := strconv.ParseFloat(arg, floatBitSize)
	if err != nil {
		return incorrectLimitMessage, errors.Wrap(err, "handle limit")
	}

	u, err := s.storage.GetUserByID(ctx, userID)
	if err != nil {
		return cannotSetLimitMessage, errors.Wrap(err, "handle limit")
	}
	rate, err := s.storage.GetRate(ctx, u.PreferredCurrencyOrDefault(s.defaultCurrency))
	if err != nil {
		return cannotGetRateMessage, errors.Wrap(err, "handle limit")
	}
	u.MonthLimit = convertToBase(limit, rate.BaseRate)
	if err = s.storage.SaveUserByID(ctx, userID, u); err != nil {
		return cannotSetLimitMessage, errors.Wrap(err, "handle limit")
	}

	return okMessage, nil
}

func (s *HandlerService) handleNoCommand(_ context.Context, arg string, userID int64) (string, error) {
	logger.Info("handleNoCommand - start", zap.Int64("userID", userID), zap.String("arg", arg))
	defer logger.Info("handleNoCommand - end")

	return loveToTalkMessage, nil
}
