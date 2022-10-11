package config

type FixerConfig struct {
	FixerApiKey string `yaml:"api-key"`
}

func (f *FixerConfig) ApiKey() string {
	return f.FixerApiKey
}
