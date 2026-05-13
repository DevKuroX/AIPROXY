package admin

import (
	"encoding/json"
	"net/http"

	"github.com/DevKuroX/AIPROXY/internal/api/middleware"
	"github.com/DevKuroX/AIPROXY/internal/auth"
	"github.com/DevKuroX/AIPROXY/internal/errs"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type meResponse struct {
	ID      string `json:"id"`
	Username string `json:"username"`
	IsAdmin bool   `json:"is_admin"`
}

type Handler struct {
	jwtSecret string
	users     UserStore
}

type UserStore interface {
	GetUserByUsername(username string) (*User, error)
}

type User struct {
	ID       string
	Username string
	Password string
	IsAdmin  bool
}

func NewHandler(jwtSecret string, users UserStore) *Handler {
	return &Handler{
		jwtSecret: jwtSecret,
		users:     users,
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errs.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		errs.WriteJSONError(w, "username and password required", http.StatusBadRequest)
		return
	}

	user, err := h.users.GetUserByUsername(req.Username)
	if err != nil {
		errs.WriteJSONError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		errs.WriteJSONError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID, user.Username, user.IsAdmin, h.jwtSecret)
	if err != nil {
		errs.WriteJSONError(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(loginResponse{Token: token})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out successfully"})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		errs.WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.users.GetUserByUsername(claims.Username)
	if err != nil {
		errs.WriteJSONError(w, "user not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(meResponse{
		ID:       user.ID,
		Username: user.Username,
		IsAdmin:  user.IsAdmin,
	})
}
