package messages

import (
	"context"
	"time"

	"go.uber.org/zap"
	apiv1 "max.ks1230/project-base/api/grpc"
	"max.ks1230/project-base/internal/logger"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const (
	errorPrefix = "Sorry, something wrong happened...\n"
)

type messageSender interface {
	SendMessage(text string, userID int64) error
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, text string, userID int64) (string, error)
	AcceptReport(ctx context.Context, report *apiv1.ReportResult) (result string, err error)
}

type Service struct {
	tgClient messageSender
	handler  MessageHandler
}

func NewService(config config,
	tgClient messageSender,
	storage userStorage,
	cache reportCache,
	producer reportRequestProducer) *Service {
	return &Service{
		tgClient: tgClient,
		handler:  newHandler(config, storage, cache, producer),
	}
}

type Message struct {
	Text   string
	UserID int64
}

func (s *Service) HandleIncomingMessage(ctx context.Context, msg Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handleMessage")
	defer span.Finish()

	start := time.Now()
	err := s.handle(ctx, msg)
	elapsed := time.Since(start)

	observeResponse(elapsed, err != nil)
	if err != nil {
		ext.Error.Set(span, true)
	}
	return err
}

func (s *Service) handle(ctx context.Context, msg Message) error {
	resp, err := s.handler.HandleMessage(ctx, msg.Text, msg.UserID)
	return s.sendResponse(resp, err, msg.UserID)
}

func (s *Service) AcceptReport(ctx context.Context, report *apiv1.ReportResult) error {
	resp, err := s.handler.AcceptReport(ctx, report)
	return s.sendResponse(resp, err, report.GetUserID())
}

func (s *Service) sendResponse(response string, err error, userID int64) error {
	if err != nil {
		senderErr := s.tgClient.SendMessage(errorPrefix+response, userID)
		if senderErr != nil {
			logger.Error("failed to send error message", zap.NamedError("senderErr", senderErr))
		}
		return err
	}
	return s.tgClient.SendMessage(response, userID)
}
