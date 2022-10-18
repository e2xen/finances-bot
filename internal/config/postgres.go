package config

type PostgresConfig struct {
	Hostname string `yaml:"host"`
	Db       string `yaml:"db"`
	User     string `yaml:"username"`
	Pswd     string `yaml:"password"`
}

func (s *PostgresConfig) Host() string {
	return s.Hostname
}

func (s *PostgresConfig) Database() string {
	return s.Db
}

func (s *PostgresConfig) Username() string {
	return s.User
}

func (s *PostgresConfig) Password() string {
	return s.Pswd
}
