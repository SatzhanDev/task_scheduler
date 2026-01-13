package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"task_scheduler/internal/auth"
	"task_scheduler/internal/user"
)

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access-token"`
}

type AuthHandler struct {
	userSvc    user.Service
	jwtManager *auth.JWTManager
}

func NewAuthHandler(svc user.Service, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{
		userSvc:    svc,
		jwtManager: jwtManager,
	}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}
	u, err := h.userSvc.Register(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidInput):
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, user.ErrEmailExists):
			WriteError(w, http.StatusConflict, "EMAIL_EXISTS", err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
		}
		return
	}
	WriteJSON(w, http.StatusCreated, u)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}

	u, err := h.userSvc.Authenticate(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidInput):
			WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		case errors.Is(err, user.ErrAuthFailed):
			WriteError(w, http.StatusUnauthorized, "UNAUTHORIZED_USER", err.Error())
		default:
			WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			log.Println("[AUTH] login error:", err)
		}
		return
	}
	token, err := h.jwtManager.Generate(u.ID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "TOKEN_ERROR", "failed to generate token")
		return
	}
	WriteJSON(w, http.StatusOK, loginResponse{AccessToken: token})
}
