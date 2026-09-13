package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"time"

	"github.com/iamsorryprincess/go-k8s-layout/migrations"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/env"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"
)

const (
	commandUp      = "up"
	commandDown    = "down"
	commandStatus  = "status"
	commandVersion = "version"
)

type config struct {
	DBURL string `env:"DB_URL"`

	Timeout time.Duration `env:"MIGRATION_TIMEOUT,5m"`
}

func main() {
	os.Exit(run())
}

func run() int {
	command := flag.String("command", commandUp, "migration command: up, down, status, version")
	flag.Parse()

	logger := log.New(log.Level(os.Getenv("LOG_LEVEL")))

	cfg, err := env.Parse[config]()
	if err != nil {
		logger.Error().Err(err).Msg("failed parse configuration")
		return 1
	}

	if cfg.DBURL == "" {
		logger.Error().Msg("DB_URL is required")
		return 1
	}

	provider, db, err := newProvider(cfg)
	if err != nil {
		logger.Error().Err(err).Msg("failed create migration provider")
		return 1
	}

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err = execute(ctx, provider, logger, *command); err != nil {
		logger.Error().Err(err).Str("command", *command).Msg("migration command failed")
		return 1
	}

	return 0
}

func newProvider(cfg config) (*goose.Provider, *sql.DB, error) {
	fsys, err := fs.Sub(migrations.FS, migrations.PostgresDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read migrations: %w", err)
	}

	db, err := sql.Open("pgx", cfg.DBURL)
	if err != nil {
		return nil, nil, fmt.Errorf("open database: %w", err)
	}

	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		db.Close()
		return nil, nil, fmt.Errorf("create session locker: %w", err)
	}

	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys, goose.WithSessionLocker(locker))
	if err != nil {
		db.Close()
		return nil, nil, err
	}

	return provider, db, nil
}

func execute(ctx context.Context, provider *goose.Provider, logger log.Logger, command string) error {
	switch command {
	case commandUp:
		return up(ctx, provider, logger)
	case commandDown:
		return down(ctx, provider, logger)
	case commandStatus:
		return status(ctx, provider, logger)
	case commandVersion:
		return version(ctx, provider, logger)
	default:
		return fmt.Errorf("unknown command %q", command)
	}
}

func up(ctx context.Context, provider *goose.Provider, logger log.Logger) error {
	results, err := provider.Up(ctx)
	if err != nil {
		return err
	}

	for _, result := range results {
		logger.Info().
			Int64("version", result.Source.Version).
			Str("source", result.Source.Path).
			Str("duration", result.Duration.String()).
			Msg("migration applied")
	}

	if len(results) == 0 {
		logger.Info().Msg("no migrations to apply")
	}

	return version(ctx, provider, logger)
}

func down(ctx context.Context, provider *goose.Provider, logger log.Logger) error {
	result, err := provider.Down(ctx)
	if err != nil {
		if errors.Is(err, goose.ErrNoNextVersion) {
			logger.Info().Msg("no migrations to roll back")
			return nil
		}

		return err
	}

	logger.Info().
		Int64("version", result.Source.Version).
		Str("source", result.Source.Path).
		Str("duration", result.Duration.String()).
		Msg("migration rolled back")

	return version(ctx, provider, logger)
}

func status(ctx context.Context, provider *goose.Provider, logger log.Logger) error {
	statuses, err := provider.Status(ctx)
	if err != nil {
		return err
	}

	for _, migrationStatus := range statuses {
		event := logger.Info().
			Int64("version", migrationStatus.Source.Version).
			Str("source", migrationStatus.Source.Path).
			Str("state", string(migrationStatus.State))

		if !migrationStatus.AppliedAt.IsZero() {
			event = event.Str("applied_at", migrationStatus.AppliedAt.Format(time.RFC3339))
		}

		event.Msg("migration status")
	}

	return nil
}

func version(ctx context.Context, provider *goose.Provider, logger log.Logger) error {
	dbVersion, err := provider.GetDBVersion(ctx)
	if err != nil {
		return err
	}

	logger.Info().Int64("version", dbVersion).Msg("database version")

	return nil
}
