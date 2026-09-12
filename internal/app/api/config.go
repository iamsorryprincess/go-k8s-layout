package api

import (
	"github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/transport/http"
)

type HealthHTTPConfig struct {
	Server   http.ServerConfig   `env:"SERVER"`
	Livez    http.LivezConfig    `env:"LIVEZ"`
	Readyz   http.ReadyzConfig   `env:"READYZ"`
	Startupz http.StartupzConfig `env:"STARTUPZ"`
}

type KubeHealthConfig struct {
	HTTP HealthHTTPConfig `env:"HTTP"`
}

type Config struct {
	Postgres   postgres.PoolConfig `env:"DB"`
	HTTP       http.ServerConfig   `env:"HTTP"`
	KubeHealth KubeHealthConfig    `env:"K8S"`
}
