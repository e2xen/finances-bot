package config

type FixerConfig struct {
	FixerAPIKey string `yaml:"api-key"`
}

func (f *FixerConfig) APIKey() string {
	return f.FixerAPIKey
}
