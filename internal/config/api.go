package config

type HTTPConfig struct {
	Host string `env:"HOST"`
	Port int    `env:"PORT"`
}

type PostgresConfig struct {
	DSN       string `env:"DSN"`
	IdleConns int    `env:"CONN_IDLE"`
}

type ApiConfig struct {
	HTTPConfig     HTTPConfig     `env:"HTTP"`
	PostgresConfig PostgresConfig `env:"POSTGRES"`
}
