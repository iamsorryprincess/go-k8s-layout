package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/iamsorryprincess/go-k8s-layout/internal/domain"
	"github.com/iamsorryprincess/go-k8s-layout/internal/repository"
	"github.com/iamsorryprincess/go-k8s-layout/pkg/log"
)

type UserProvider interface {
	GetUser(ctx context.Context, id uint64) (domain.User, error)
	GetUsers(ctx context.Context) ([]domain.User, error)
	CreateUser(ctx context.Context, user domain.User) (uint64, error)
}

type userResponse struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type createUserRequest struct {
	Name string `json:"name"`
}

type createUserResponse struct {
	ID uint64 `json:"id"`
}

type errorResponse struct {
	Error string `json:"error"`
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

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(r.PathValue("id"), 10, 64)
	if err != nil {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{Error: "invalid user id"})
		return
	}

	user, err := h.provider.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.writeJSON(w, r, http.StatusNotFound, errorResponse{Error: "user not found"})
			return
		}

		h.logger.Error().Err(err).Uint64("user_id", id).Msg("failed to get user")
		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	h.writeJSON(w, r, http.StatusOK, userResponse{
		ID:   user.ID,
		Name: user.Name,
	})
}

func (h *UserHandler) GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.provider.GetUsers(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to get users")
		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	response := make([]userResponse, 0, len(users))
	for _, user := range users {
		response = append(response, userResponse{
			ID:   user.ID,
			Name: user.Name,
		})
	}

	h.writeJSON(w, r, http.StatusOK, response)
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeJSON(w, r, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	id, err := h.provider.CreateUser(r.Context(), domain.User{Name: request.Name})
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to create user")
		h.writeJSON(w, r, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	h.writeJSON(w, r, http.StatusCreated, createUserResponse{ID: id})
}

func (h *UserHandler) writeJSON(w http.ResponseWriter, r *http.Request, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error().Err(err).
			Str("method", r.Method).
			Str("url", r.RequestURI).
			Msg("failed to write response")
	}
}
