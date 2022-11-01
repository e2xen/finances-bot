package messages

import (
	"context"
	"strconv"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type messageSender interface {
	SendMessage(text string, userID int64) error
}

type MessageHandler interface {
	HandleMessage(ctx context.Context, text string, userID int64) (string, error)
}

type Service struct {
	tgClient messageSender
	handler  MessageHandler
}

func NewService(tgClient messageSender, storage userStorage, config config) *Service {
	return &Service{
		tgClient: tgClient,
		handler:  newHandler(storage, config),
	}
}

type Message struct {
	Text   string
	UserID int64
}

var HistogramResponseTime = promauto.NewHistogramVec(
	prometheus.HistogramOpts{
		Namespace: "route256",
		Subsystem: "telegram",
		Name:      "histogram_response_time_seconds",
		Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2},
	},
	[]string{"status"},
)

func (s *Service) HandleIncomingMessage(ctx context.Context, msg Message) error {
	span, ctx := opentracing.StartSpanFromContext(ctx, "handleMessage")
	defer span.Finish()

	start := time.Now()
	err := s.handle(ctx, msg)
	elapsed := time.Since(start)

	HistogramResponseTime.
		WithLabelValues(strconv.FormatBool(err != nil)).
		Observe(elapsed.Seconds())

	if err != nil {
		ext.Error.Set(span, true)
	}
	return err
}

func (s *Service) handle(ctx context.Context, msg Message) error {
	resp, err := s.handler.HandleMessage(ctx, msg.Text, msg.UserID)
	if err != nil {
		_ = s.tgClient.SendMessage("Sorry, something wrong happened...\n"+resp, msg.UserID)
		return err
	}
	return s.tgClient.SendMessage(resp, msg.UserID)
}
