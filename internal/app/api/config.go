package api

import "github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"

type Config struct {
	Postgres postgres.PoolConfig `env:"DB"`
}
