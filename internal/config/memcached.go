package config

type MemcachedConfig struct {
	NodeHosts []string `yaml:"hosts"`
}

func (s *MemcachedConfig) Hosts() []string {
	return s.NodeHosts
}
