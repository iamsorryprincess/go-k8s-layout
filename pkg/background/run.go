package background

import (
	"context"
	"os"
	"sync/atomic"
	"time"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/env"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type AppCtx interface {
	Context() context.Context
	IsRunning() bool
	Fatal() chan<- error
}

type appCtx struct {
	ctx     context.Context
	running *atomic.Bool
	fatal   chan error
}

func (c appCtx) Context() context.Context {
	return c.ctx
}

func (c appCtx) IsRunning() bool {
	return c.running.Load()
}

func (c appCtx) Fatal() chan<- error {
	return c.fatal
}

type App interface {
	Run(ctx AppCtx) error
	Close()
}

type CreateAppFunc[TConfig any, TApp App] func(config TConfig, logger log.Logger) TApp

type runConfig struct {
	PreShutdownDelay time.Duration `env:"PRE_SHUTDOWN_DELAY,0s"`
}

func Run[TConfig any, TApp App](createAppFunc CreateAppFunc[TConfig, TApp]) int {
	logger := log.New(log.Level(os.Getenv("LOG_LEVEL")))

	runCfg, err := env.Parse[runConfig]()
	if err != nil {
		logger.Error().Err(err).Msg("failed parse run configuration")
		return 1
	}

	cfg, err := env.Parse[TConfig]()
	if err != nil {
		logger.Error().Err(err).Msg("failed parse configuration")
		return 1
	}

	var running atomic.Bool
	isStopping := false

	defer func() {
		running.Store(false)
		if isStopping {
			logger.Info().Msg("app stopped")
		}
	}()

	app := createAppFunc(cfg, logger)
	defer app.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger.Info().Msg("app starting")

	appContext := appCtx{
		ctx:     ctx,
		running: &running,
		fatal:   make(chan error, 1),
	}

	if err = app.Run(appContext); err != nil {
		logger.Error().Err(err).Msg("failed start app")
		return 1
	}

	running.Store(true)
	logger.Info().Msg("app started")

	s, err := Wait(appContext.fatal)
	running.Store(false)

	if err != nil {
		logger.Error().Err(err).Msg("app fatal exit")
		return 1
	}

	isStopping = true
	logger.Info().Str("signal", s.String()).Msg("app stopping")

	if runCfg.PreShutdownDelay > 0 {
		logger.Info().
			Str("delay", runCfg.PreShutdownDelay.String()).
			Msg("waiting before shutdown")

		time.Sleep(runCfg.PreShutdownDelay)
	}

	return 0
}
