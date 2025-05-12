package api

import (
	"encoding/json"
	"net/http"

	// import the business-logic layer
	"github.com/kk-Syuer/wasatext_2024/service"
)

// SessionHandler handles user login sessions.
type SessionHandler struct {
	// SessionService provides the login/create-user logic.
	SessionService service.SessionService
}

// NewSessionHandler constructs a SessionHandler.
func NewSessionHandler(svc service.SessionService) *SessionHandler {
	return &SessionHandler{SessionService: svc}
}

// LoginRequest is the expected payload for POST /session.
type LoginRequest struct {
	Username string `json:"username"` // 3–16 characters
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	Identifier string `json:"identifier"` // auth token
	Username   string `json:"username"`   // echo
}

// DoLogin decodes the request, calls the business layer, and returns JSON.
// Registered via router.POST("/session", ...).
func (h *SessionHandler) DoLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	// 1) Decode JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// 2) Basic validation
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Username must be 3–16 characters", http.StatusBadRequest)
		return
	}

	// 3) Business logic: login or create user, return token
	token, err := h.SessionService.Login(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	// 4) Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(LoginResponse{
		Identifier: token,
		Username:   req.Username,
	})
}
