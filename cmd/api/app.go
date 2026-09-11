package main

import (
	"context"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/background"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type App struct {
	background.BaseApp

	config Config
	logger log.Logger

	postgresPool *postgres.Pool
}

func NewApp(config Config, logger log.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (a *App) Run(_ context.Context, _ chan<- error) error {
	var err error

	if a.postgresPool, err = postgres.NewPool(a.config.Postgres, a.logger); err != nil {
		return err
	}

	a.DeferCloser(a.postgresPool)
	a.logger.Info().Msg("postgres connected")

	return nil
}
