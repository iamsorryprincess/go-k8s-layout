package background

import (
	"context"
	"os"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/env"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type App interface {
	Run(ctx context.Context, fatal chan<- error) error
	Close()
}

type CreateAppFunc[TConfig any, TApp App] func(config TConfig, logger log.Logger) TApp

func Run[TConfig any, TApp App](createAppFunc CreateAppFunc[TConfig, TApp]) int {
	logger := log.New(log.Level(os.Getenv("LOG_LEVEL")))

	cfg, err := env.Parse[TConfig]()
	if err != nil {
		logger.Error().Err(err).Msg("failed parse configuration")
		return 1
	}

	isStopping := false

	defer func() {
		if isStopping {
			logger.Info().Msg("app stopped")
		}
	}()

	app := createAppFunc(cfg, logger)
	defer app.Close()

	fatal := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info().Msg("app starting")

	if err = app.Run(ctx, fatal); err != nil {
		logger.Error().Err(err).Msg("failed start app")
		return 1
	}

	logger.Info().Msg("app started")

	s, err := Wait(fatal)
	if err != nil {
		logger.Error().Err(err).Msg("app fatal exit")
		return 1
	}

	isStopping = true
	logger.Info().Str("signal", s.String()).Msg("app stopping")
	return 0
}
