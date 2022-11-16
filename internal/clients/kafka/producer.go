package kafka

import (
	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"max.ks1230/project-base/internal/logger"
)

type producerConfig interface {
	Brokers() []string
	ReportsTopic() string
}

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(cfg producerConfig) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_5_0_0
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(cfg.Brokers(), config)
	return &Producer{
		producer: producer,
		topic:    cfg.ReportsTopic(),
	}, err
}

func (p *Producer) ProduceMessage(message []byte) error {
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(message),
	})
	return err
}

func (p *Producer) Close() {
	err := p.producer.Close()
	if err != nil {
		logger.Error("failed to close producer", zap.Error(err))
	}
}
