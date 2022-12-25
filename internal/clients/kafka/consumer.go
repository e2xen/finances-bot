package kafka

import (
	"context"
	"fmt"

	apiv12 "max.ks1230/finances-bot/api/grpc"

	"google.golang.org/protobuf/proto"
	apiv1 "max.ks1230/finances-bot/api/kafka"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/Shopify/sarama"
	"max.ks1230/finances-bot/internal/logger"
)

type consumerConfig interface {
	producerConfig
	ConsumerGroup() string
}

type reportGenerator interface {
	GenerateReport(ctx context.Context, userID int64, period string) (report *apiv12.ReportResult, err error)
}

type reportSender interface {
	SendReport(ctx context.Context, report *apiv12.ReportResult) error
}

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	generator     reportGenerator
	sender        reportSender
}

func NewConsumer(cfg consumerConfig, generator reportGenerator, sender reportSender) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers(), cfg.ConsumerGroup(), config)
	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         cfg.ReportsTopic(),
		generator:     generator,
		sender:        sender,
	}, err
}

func (c *Consumer) StartConsuming(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			err := c.consumerGroup.Consume(ctx, []string{c.topic}, c)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("consume from %s", c.topic))
			}
		}
	}
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	logger.Info("consumer - setup")
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	logger.Info("consumer - cleanup")
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var req apiv1.ReportRequest
		err := proto.Unmarshal(message.Value, &req)
		if err != nil {
			logger.Error("cannot unmarshal kafka message", zap.Error(err))
		} else {
			logger.Info(
				"received report request",
				zap.ByteString("key", message.Key),
				zap.Int64("userID", req.UserID),
				zap.String("period", req.Period),
			)
			c.processRequest(session.Context(), &req)
		}
		session.MarkMessage(message, "")
	}

	return nil
}

func (c *Consumer) processRequest(ctx context.Context, req *apiv1.ReportRequest) {
	report, _ := c.generator.GenerateReport(ctx, req.GetUserID(), req.GetPeriod())
	err := c.sender.SendReport(ctx, report)
	if err != nil {
		logger.Error("failed to send report", zap.Error(err))
	}
}
