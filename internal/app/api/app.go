package api

import (
	"context"

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

	httpServer *http.Server
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

	a.userRepo = postgresrepo.NewUserRepository(a.postgresPool)

	if err = a.initHTTP(fatal); err != nil {
		return err
	}

	return nil
}

func (a *App) initHTTP(fatal chan<- error) error {
	router := http.NewRouter().
		Use(http.Recovery(a.logger)).
		Use(http.CORS)

	users := httptransport.NewUserHandler(a.logger, a.userRepo)
	router.HandleFunc("GET /users/{id}", users.GetUser).
		HandleFunc("GET /users", users.GetUsers).
		HandleFunc("POST /users", users.CreateUser)

	a.httpServer = http.NewServer(a.config.HTTP, a.logger, router)
	if err := a.httpServer.Start(fatal); err != nil {
		return err
	}

	a.DeferCloser(a.httpServer)
	a.logger.Info().Msg("http server started")

	return nil
}
