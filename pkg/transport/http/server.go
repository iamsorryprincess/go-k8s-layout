package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type ServerConfig struct {
	Addr string `env:"ADDR,:8080"`

	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT,10s"`
}

type Server struct {
	config ServerConfig
	logger log.Logger

	server *http.Server

	wg sync.WaitGroup
}

func NewServer(config ServerConfig, logger log.Logger, handler http.Handler) *Server {
	return &Server{
		config: config,
		logger: logger,
		server: &http.Server{
			Handler: handler,
		},
	}
}

func (s *Server) Start(fatal chan<- error) error {
	ln, err := net.Listen("tcp", s.config.Addr)
	if err != nil {
		return fmt.Errorf("http server listen: %w", err)
	}

	s.wg.Go(func() {
		if err := s.server.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case fatal <- fmt.Errorf("http server serve: %w", err):
			default:
			}
		}
	})

	return nil
}

func (s *Server) Close() {
	defer s.wg.Wait()
	s.logger.Info().Msg("http server stopping")

	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("http server shutdown failed")

		if err = s.server.Close(); err != nil {
			s.logger.Error().Err(err).Msg("http server force close failed")
			return
		}

		s.logger.Info().Msg("http server force stopped")
		return
	}

	s.logger.Info().Msg("http server stopped")
}
