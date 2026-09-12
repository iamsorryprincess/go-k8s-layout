package api

import (
	"context"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/background"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
	httptransport "github.com/iamsorryprincess/go-k8s-layout/pkg/transport/http"
)

type App struct {
	background.BaseApp

	config Config
	logger log.Logger

	postgresPool *postgres.Pool

	httpServer *httptransport.Server
}

func New(config Config, logger log.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (a *App) Run(_ context.Context, fatal chan<- error) error {
	var err error

	if a.postgresPool, err = postgres.NewPool(a.config.Postgres, a.logger); err != nil {
		return err
	}

	a.DeferCloser(a.postgresPool)
	a.logger.Info().Msg("postgres connected")

	if err = a.initHTTP(fatal); err != nil {
		return err
	}

	return nil
}

func (a *App) initHTTP(fatal chan<- error) error {
	a.httpServer = httptransport.New(a.config.HTTP, a.logger, nil)
	if err := a.httpServer.Start(fatal); err != nil {
		return err
	}

	a.DeferCloser(a.httpServer)
	a.logger.Info().Msg("http server started")

	return nil
}
