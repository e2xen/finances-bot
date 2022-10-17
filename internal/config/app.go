package config

type AppConfig struct {
	BaseCurrencyName        string `yaml:"base-currency"`
	RatePullingDelayMinutes int64  `yaml:"rate-pulling-delay-minutes"`
}

func (s *AppConfig) BaseCurrency() string {
	return s.BaseCurrencyName
}

func (s *AppConfig) PullingDelayMinutes() int64 {
	return s.RatePullingDelayMinutes
}
