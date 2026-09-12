package http

import (
	"context"
	"net/http"

	"github.com/iamsorryprincess/go-k8s-layout/internal/domain"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type UserProvider interface {
	GetUser(ctx context.Context, id uint64) (domain.User, error)
	GetUsers(ctx context.Context) ([]domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (uint64, error)
}

type UserHandler struct {
	logger   log.Logger
	provider UserProvider
}

func NewUserHandler(logger log.Logger, provider UserProvider) *UserHandler {
	return &UserHandler{
		logger:   logger,
		provider: provider,
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {}
