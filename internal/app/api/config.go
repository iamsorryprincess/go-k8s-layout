package api

import (
	"github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/transport/http"
)

type Config struct {
	Postgres postgres.PoolConfig `env:"DB"`
	HTTP     http.ServerConfig   `env:"HTTP"`
}
