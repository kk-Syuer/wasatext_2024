package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kk-Syuer/wasatext_2024/service"
)

// SessionHandler handles login/create‐session requests.
type SessionHandler struct {
	SessionService service.SessionService
}

// NewSessionHandler constructs a SessionHandler.
func NewSessionHandler(svc service.SessionService) *SessionHandler {
	return &SessionHandler{SessionService: svc}
}

// LoginRequest is the payload for POST /session.
type LoginRequest struct {
	Username string `json:"username"` // 3–16 characters
}

// LoginResponse is returned on successful login.
type LoginResponse struct {
	Identifier string `json:"identifier"` // auth token
	Username   string `json:"username"`
}

// DoLogin handles POST /session: decodes JSON, calls SessionService.Login, and writes JSON.
func (h *SessionHandler) DoLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Username must be 3–16 characters", http.StatusBadRequest)
		return
	}

	token, err := h.SessionService.Login(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	resp := LoginResponse{
		Identifier: token,
		Username:   req.Username,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------------------------------------

// UserHandler handles /users endpoints.
type UserHandler struct {
	UserService service.UserService
}

// NewUserHandler constructs a UserHandler.
func NewUserHandler(svc service.UserService) *UserHandler {
	return &UserHandler{UserService: svc}
}

// ListUsers handles GET /users: returns all users.
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(users)
}

// GetUser handles GET /users/:username: returns a single user.
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	username := ps.ByName("username")

	user, err := h.UserService.GetUser(r.Context(), username)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(user)
}

// updateNameReq is the payload for PATCH /users/:username/name.
type updateNameReq struct {
	Name string `json:"name"`
}

// UpdateName handles PATCH /users/:username/name: updates the user's display name.
func (h *UserHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
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

	if err := h.UserService.UpdateName(r.Context(), username, req.Name); err != nil {
		http.Error(w, "Failed to update name", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// updatePhotoReq is the payload for PUT /users/:username/photo.
type updatePhotoReq struct {
	PhotoURL string `json:"photoUrl"`
}

// UpdatePhoto handles PUT /users/:username/photo: updates the user's avatar URL.
func (h *UserHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
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

	if err := h.UserService.UpdatePhoto(r.Context(), username, req.PhotoURL); err != nil {
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// -------------------------------------------------------------------------------------------------
// ConversationHandler handles /conversations endpoints.
type ConversationHandler struct {
	ConvSvc service.ConversationService
}

// NewConversationHandler constructs a ConversationHandler.
func NewConversationHandler(svc service.ConversationService) *ConversationHandler {
	return &ConversationHandler{ConvSvc: svc}
}

// ListConversations handles GET /conversations?user={username}
func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get("user")
	if username == "" {
		http.Error(w, "`user` query param is required", http.StatusBadRequest)
		return
	}

	convs, err := h.ConvSvc.ListConversations(r.Context(), username)
	if err != nil {
		http.Error(w, "Failed to list conversations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(convs)
}

// CreateConversationRequest is the payload for POST /conversations
type CreateConversationRequest struct {
	Type         string   `json:"type"`         // "individual" or "group"
	Participants []string `json:"participants"` // list of usernames
}

func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if len(req.Participants) < 2 {
		http.Error(w, "At least two participants required", http.StatusBadRequest)
		return
	}

	convType := service.ConversationType(req.Type)
	conv, err := h.ConvSvc.CreateConversation(r.Context(), convType, req.Participants)
	if err != nil {
		// 1) Log it server-side
		log.Printf("CreateConversation error: %v", err)
		// 2) Return the real error in the response (for debugging)
		http.Error(w, fmt.Sprintf("Failed to create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(conv)
}

// GetConversation handles GET /conversations/:id
func (h *ConversationHandler) GetConversation(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	id := ps.ByName("id")

	conv, err := h.ConvSvc.GetConversation(r.Context(), id)
	if err != nil {
		http.Error(w, "Conversation not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(conv)
}

// GetDeliveryStatus handles GET /conversations/:id/delivery
func (h *ConversationHandler) GetDeliveryStatus(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	id := ps.ByName("id")

	status, err := h.ConvSvc.GetDeliveryStatus(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to fetch delivery status", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(status)
}
