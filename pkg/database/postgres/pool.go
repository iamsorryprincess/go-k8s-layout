package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolConfig struct {
	URL string `env:"URL"`

	// MaxConns is the maximum size of the pool. The default is the greater of 4 or runtime.NumCPU().
	MaxConns int32 `env:"MAX_CONNS,10"`

	// MinConns is the minimum size of the pool. After connection closes, the pool might dip below MinConns. A low
	// number of MinConns might mean the pool is empty after MaxConnLifetime until the health check has a chance
	// to create new connections.
	MinConns int32 `env:"MIN_CONNS,5"`

	// MinIdleConns is the minimum number of idle connections in the pool. You can increase this to ensure that
	// there are always idle connections available. This can help reduce tail latencies during request processing,
	// as you can avoid the latency of establishing a new connection while handling requests. It is superior
	// to MinConns for this purpose.
	// Similar to MinConns, the pool might temporarily dip below MinIdleConns after connection closes.
	MinIdleConns int32 `env:"MIN_IDLE_CONNS,5"`

	// MaxConnLifetime is the duration since creation after which a connection will be automatically closed.
	MaxConnLifetime time.Duration `env:"MAX_CONN_LIFETIME,1h"`

	// MaxConnLifetimeJitter is the duration after MaxConnLifetime to randomly decide to close a connection.
	// This helps prevent all connections from being closed at the exact same time, starving the pool.
	MaxConnLifetimeJitter time.Duration `env:"MAX_CONN_LIFETIME_JITTER,10m"`

	// MaxConnIdleTime is the duration after which an idle connection will be automatically closed by the health check.
	MaxConnIdleTime time.Duration `env:"MAX_CONN_IDLE_TIME,30m"`

	// PingTimeout is the maximum amount of time to wait for a connection to pong before considering it as unhealthy and
	// destroying it. If zero, the default is no timeout.
	PingTimeout time.Duration `env:"PING_TIMEOUT,0"`
}

type Pool struct {
	config PoolConfig
	logger log.Logger
	*pgxpool.Pool
}

func NewPool(config PoolConfig, logger log.Logger) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.URL)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse config: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MinIdleConns = config.MinIdleConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnLifetimeJitter = config.MaxConnLifetimeJitter
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.PingTimeout = config.PingTimeout

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres: create pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}

	return &Pool{
		config: config,
		logger: logger,
		Pool:   pool,
	}, nil
}

func (p *Pool) Close() {
	p.Pool.Close()
	p.logger.Info().Msg("postgres closed")
}
