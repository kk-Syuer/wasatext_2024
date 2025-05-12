package api

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kk-Syuer/wasatext_2024/service"
	"github.com/kk-Syuer/wasatext_2024/service/database"
)

// UserHandler manages all /users endpoints.
type UserHandler struct {
	// UserService provides business logic for user operations.
	UserService service.UserService
}

// NewUserHandler constructs a UserHandler with the given UserService.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{UserService: svc}
}

// ListUsers handles GET /users and returns all users.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

// GetUser handles GET /users/:username and returns a single user.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// Extract username from URL params
	ps := httprouter.ParamsFromContext(r.Context())
	username := ps.ByName("username")

	user, err := h.UserService.GetUser(r.Context(), username)
	if err != nil {
		if err == database.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get user", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

// updateNameReq defines the expected payload for PATCH /users/:username/name
type updateNameReq struct {
	Name string `json:"name"`
}

// UpdateName handles PATCH /users/:username/name to change user's display name.
func (h *UserHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ps := httprouter.ParamsFromContext(ctx)
	username := ps.ByName("username")

	var req updateNameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Name cannot be empty", http.StatusBadRequest)
		return
	}
	// Call business logic
	err := h.UserService.UpdateName(ctx, username, req.Name)
	if err != nil {
		if err == database.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update name", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updatePhotoReq defines the expected payload for PUT /users/:username/photo
type updatePhotoReq struct {
	PhotoURL string `json:"photoUrl"`
}

// UpdatePhoto handles PUT /users/:username/photo to change user's profile picture.
func (h *UserHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	ps := httprouter.ParamsFromContext(ctx)
	username := ps.ByName("username")

	var req updatePhotoReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.PhotoURL == "" {
		http.Error(w, "photoUrl cannot be empty", http.StatusBadRequest)
		return
	}
	// Call business logic
	err := h.UserService.UpdatePhoto(ctx, username, req.PhotoURL)
	if err != nil {
		if err == database.ErrNotFound {
			http.Error(w, "User not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
