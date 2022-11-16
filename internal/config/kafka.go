package config

type KafkaConfig struct {
	BrokerList []string `yaml:"brokers"`
	Consumer   string   `yaml:"consumer-group"`
	RepTopic   string   `yaml:"reports-topic"`
}

func (s *KafkaConfig) Brokers() []string {
	return s.BrokerList
}

func (s *KafkaConfig) ConsumerGroup() string {
	return s.Consumer
}

func (s *KafkaConfig) ReportsTopic() string {
	return s.RepTopic
}
