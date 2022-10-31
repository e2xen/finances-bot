package config

import (
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const configFile = "data/config.yaml"

type config struct {
	Telegram TelegramConfig `yaml:"telegram"`
	Fixer    FixerConfig    `yaml:"fixer"`
	App      AppConfig      `yaml:"app"`
	Postgres PostgresConfig `yaml:"postgres"`
}

type Service struct {
	config config
}

func New() (*Service, error) {
	s := &Service{}

	rawYAML, err := os.ReadFile(configFile)
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}

	err = yaml.Unmarshal(rawYAML, &s.config)
	if err != nil {
		return nil, errors.Wrap(err, "parsing yaml")
	}

	return s, nil
}

func (s *Service) Telegram() *TelegramConfig {
	return &s.config.Telegram
}

func (s *Service) Fixer() *FixerConfig {
	return &s.config.Fixer
}

func (s *Service) App() *AppConfig {
	return &s.config.App
}

func (s *Service) Postgres() *PostgresConfig {
	return &s.config.Postgres
}
