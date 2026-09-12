package api

import (
	postgresrepo "github.com/iamsorryprincess/go-k8s-layout/internal/repository/postgres"
	httptransport "github.com/iamsorryprincess/go-k8s-layout/internal/transport/http"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/background"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/database/postgres"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/transport/http"
)

type App struct {
	background.BaseApp

	config Config
	logger log.Logger

	postgresPool *postgres.Pool

	userRepo *postgresrepo.UserRepository

	httpServer       *http.Server
	httpHealthServer *http.Server
}

func New(config Config, logger log.Logger) *App {
	return &App{
		config: config,
		logger: logger,
	}
}

func (a *App) Run(ctx background.AppCtx) error {
	var err error

	if a.postgresPool, err = postgres.NewPool(a.config.Postgres, a.logger); err != nil {
		return err
	}

	a.DeferCloser(a.postgresPool)
	a.logger.Info().Msg("postgres connected")

	a.userRepo = postgresrepo.NewUserRepository(a.postgresPool)

	if err = a.initHTTP(ctx); err != nil {
		return err
	}

	return nil
}

func (a *App) initHTTP(ctx background.AppCtx) error {
	health := http.NewHealthHandler(a.logger).
		UseLivez(a.config.KubeHealth.HTTP.Livez).
		UseReadyz(a.config.KubeHealth.HTTP.Readyz,
			http.NewAppRunningCheck(ctx),
			http.NewCheck("db", a.postgresPool.Ping),
		).
		UseStartupz(a.config.KubeHealth.HTTP.Startupz)

	a.httpHealthServer = http.NewServer(a.config.KubeHealth.HTTP.Server, a.logger, health)
	if err := a.httpHealthServer.Start(ctx.Fatal()); err != nil {
		return err
	}

	a.DeferCloser(a.httpHealthServer)

	router := http.NewRouter().
		Use(http.Recovery(a.logger)).
		Use(http.CORS)

	users := httptransport.NewUserHandler(a.logger, a.userRepo)
	router.HandleFunc("GET /users/{id}", users.GetUser).
		HandleFunc("GET /users", users.GetUsers).
		HandleFunc("POST /users", users.CreateUser)

	a.httpServer = http.NewServer(a.config.HTTP, a.logger, router)
	if err := a.httpServer.Start(ctx.Fatal()); err != nil {
		return err
	}

	a.DeferCloser(a.httpServer)

	return nil
}
