package config

type TelegramConfig struct {
	APIToken string `yaml:"token"`
}

func (t *TelegramConfig) Token() string {
	return t.APIToken
}
