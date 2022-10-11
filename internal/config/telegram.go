package config

type TelegramConfig struct {
	ApiToken string `yaml:"token"`
}

func (t *TelegramConfig) Token() string {
	return t.ApiToken
}
