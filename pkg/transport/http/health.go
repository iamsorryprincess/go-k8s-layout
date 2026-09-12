package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

const (
	statusOK          = "ok"
	statusUnavailable = "unavailable"
)

type Check struct {
	Name string
	Func func(ctx context.Context) error
}

func NewCheck(name string, checkFunc func(ctx context.Context) error) Check {
	return Check{
		Name: name,
		Func: checkFunc,
	}
}

var ErrAppNotStarted = errors.New("app not started")

type AppRunningChecker interface {
	IsRunning() bool
}

func NewAppRunningCheck(checker AppRunningChecker) Check {
	return NewCheck("app_running", func(_ context.Context) error {
		if checker.IsRunning() {
			return nil
		}
		return ErrAppNotStarted
	})
}

type LivezConfig struct {
	Path string `env:"PATH,/livez"`

	Timeout time.Duration `env:"TIMEOUT,1s"`

	LogInterval time.Duration `env:"LOG_INTERVAL,0s"`
}

type ReadyzConfig struct {
	Path string `env:"PATH,/readyz"`

	Timeout time.Duration `env:"TIMEOUT,1s"`

	LogInterval time.Duration `env:"LOG_INTERVAL,0s"`
}

type StartupzConfig struct {
	Path string `env:"PATH,/startupz"`

	Timeout time.Duration `env:"TIMEOUT,1s"`

	LogInterval time.Duration `env:"LOG_INTERVAL,0s"`
}

type probeParams struct {
	name        string
	path        string
	timeout     time.Duration
	logInterval time.Duration
}

type checkState struct {
	check      Check
	failed     atomic.Bool
	failedAt   atomic.Int64
	lastLogged atomic.Int64
}

type healthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

type HealthHandler struct {
	logger log.Logger
	mux    *http.ServeMux
}

func NewHealthHandler(logger log.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
		mux:    http.NewServeMux(),
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *HealthHandler) UseLivez(config LivezConfig, checks ...Check) *HealthHandler {
	return h.use(probeParams{
		name:        "livez",
		path:        config.Path,
		timeout:     config.Timeout,
		logInterval: config.LogInterval,
	}, checks)
}

func (h *HealthHandler) UseReadyz(config ReadyzConfig, checks ...Check) *HealthHandler {
	return h.use(probeParams{
		name:        "readyz",
		path:        config.Path,
		timeout:     config.Timeout,
		logInterval: config.LogInterval,
	}, checks)
}

func (h *HealthHandler) UseStartupz(config StartupzConfig, checks ...Check) *HealthHandler {
	return h.use(probeParams{
		name:        "startupz",
		path:        config.Path,
		timeout:     config.Timeout,
		logInterval: config.LogInterval,
	}, checks)
}

func (h *HealthHandler) use(params probeParams, checks []Check) *HealthHandler {
	states := make([]*checkState, 0, len(checks))
	for _, check := range checks {
		states = append(states, &checkState{check: check})
	}

	h.mux.HandleFunc("GET "+params.path, func(w http.ResponseWriter, r *http.Request) {
		h.serve(w, r, params, states)
	})

	return h
}

func (h *HealthHandler) serve(w http.ResponseWriter, r *http.Request, params probeParams, states []*checkState) {
	ctx := r.Context()

	if params.timeout > 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, params.timeout)
		defer cancel()
		ctx = timeoutCtx
	}

	response := healthResponse{
		Status: statusOK,
	}

	if len(states) > 0 {
		response.Checks = make(map[string]string, len(states))
	}

	for _, state := range states {
		if err := state.check.Func(ctx); err != nil {
			message := singleLine(err.Error())

			response.Status = statusUnavailable
			response.Checks[state.check.Name] = message

			h.logFailed(params, state, message)

			continue
		}

		response.Checks[state.check.Name] = statusOK

		if state.failed.CompareAndSwap(true, false) {
			h.logger.Info().
				Str("probe", params.name).
				Str("check", state.check.Name).
				Msg("health check recovered")
		}
	}

	statusCode := http.StatusOK
	if response.Status != statusOK {
		statusCode = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error().Err(err).
			Str("probe", params.name).
			Msg("failed to write health response")
	}
}

func (h *HealthHandler) logFailed(params probeParams, state *checkState, message string) {
	now := time.Now().UnixNano()

	if state.failed.CompareAndSwap(false, true) {
		state.failedAt.Store(now)
		state.lastLogged.Store(now)

		h.logger.Warn().
			Str("probe", params.name).
			Str("check", state.check.Name).
			Str("error", message).
			Msg("health check failed")

		return
	}

	if params.logInterval <= 0 {
		return
	}

	lastLogged := state.lastLogged.Load()
	if now-lastLogged < int64(params.logInterval) {
		return
	}

	if !state.lastLogged.CompareAndSwap(lastLogged, now) {
		return
	}

	h.logger.Warn().
		Str("probe", params.name).
		Str("check", state.check.Name).
		Str("error", message).
		Str("failing_for", time.Duration(now-state.failedAt.Load()).Truncate(time.Second).String()).
		Msg("health check still failing")
}

func singleLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
