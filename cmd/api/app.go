package main

import (
	"context"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/background"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type App struct {
	background.BaseApp

	config Config
	logger log.Logger
}

func NewApp(config Config, logger log.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (a *App) Run(ctx context.Context, fatal chan<- error) error {
	return nil
}
