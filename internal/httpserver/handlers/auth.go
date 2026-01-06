package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
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

type AuthHandler struct {
	svc user.Service
}

func NewAuthHandler(svc user.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid json")
		return
	}
	u, err := h.svc.Register(req.Email, req.Password)
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

	u, err := h.svc.Authenticate(req.Email, req.Password)
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
	WriteJSON(w, http.StatusOK, u)
}
