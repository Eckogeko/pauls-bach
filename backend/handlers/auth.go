package handlers

import (
	"encoding/json"
	"net/http"
	"pauls-bach/middleware"
	"pauls-bach/models"
	"pauls-bach/store"

	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	Store     *store.Store
	JWTSecret string
}

type loginRequest struct {
	Username string `json:"username"`
	PIN      string `json:"pin"`
}

type authResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}
	if req.Username == "" || req.PIN == "" {
		jsonError(w, "username and pin required", http.StatusBadRequest)
		return
	}
	if req.Username == "admin" {
		jsonError(w, "cannot register as admin", http.StatusBadRequest)
		return
	}

	store.WriteLock()
	defer store.WriteUnlock()

	if existing, _ := h.Store.Users.GetByUsername(req.Username); existing != nil {
		jsonError(w, "username already taken", http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}

	user := &models.User{
		Username: req.Username,
		PinHash:  string(hash),
		Balance:  1000,
		IsAdmin:  false,
	}
	if err := h.Store.Users.Create(user); err != nil {
		jsonError(w, "failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := middleware.GenerateToken(h.JWTSecret, user.ID, user.Username, user.IsAdmin)
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonResp(w, authResponse{Token: token, User: user}, http.StatusCreated)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request", http.StatusBadRequest)
		return
	}

	store.ReadLock()
	defer store.ReadUnlock()

	user, err := h.Store.Users.GetByUsername(req.Username)
	if err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(req.PIN)); err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := middleware.GenerateToken(h.JWTSecret, user.ID, user.Username, user.IsAdmin)
	if err != nil {
		jsonError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	jsonResp(w, authResponse{Token: token, User: user}, http.StatusOK)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	store.ReadLock()
	defer store.ReadUnlock()

	user, err := h.Store.Users.GetByID(userID)
	if err != nil {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}

	jsonResp(w, user, http.StatusOK)
}
