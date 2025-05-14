package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"

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

// -------------------------------------------------------
// UpdateMyName handles PATCH /user/name
func (h *UserHandler) UpdateMyName(w http.ResponseWriter, r *http.Request) {
	// 1) Verify session auth
	username := UsernameFromContext(r.Context())
	if username == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 2) Decode new name from JSON
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// 3) Update via service
	if err := h.UserService.UpdateName(r.Context(), username, req.Name); err != nil {
		http.Error(w, "Failed to update name", http.StatusInternalServerError)
		return
	}

	// 4) Echo back NameResponse { username, name }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"username": username,
		"name":     req.Name,
	})
}

// UpdateMyPhoto handles PATCH /user/photo (multipart/form-data)
func (h *UserHandler) UpdateMyPhoto(w http.ResponseWriter, r *http.Request) {
	// 1) Parse the multipart form, allow up to 10 MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid multipart payload", http.StatusBadRequest)
		return
	}

	// 2) Grab the uploaded file under field “photo”
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3) Determine a safe filename and target directory
	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := uuid.New().String() + ext
	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}
	dstPath := filepath.Join(uploadDir, filename)

	// 4) Write file to disk
	out, err := os.Create(dstPath)
	if err != nil {
		http.Error(w, "Failed to store photo", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, file); err != nil {
		http.Error(w, "Failed to save photo", http.StatusInternalServerError)
		return
	}

	// 5) Build the resulting public URL (adjust to your CDN/host)
	photoURL := fmt.Sprintf("https://%s/uploads/%s", r.Host, filename)

	username := UsernameFromContext(r.Context())
	if username == "" {
		log.Printf("no username in context; headers = %+v", r.Header)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 7) Update in your service layer
	if err := h.UserService.UpdatePhoto(r.Context(), username, photoURL); err != nil {
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}

	// 8) Respond with the spec’s PhotoResponse schema
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"photoUrl": photoURL,
	})
}

// ListConversationsForUser handles GET /users/:username/conversations
// func (h *ConversationHandler) ListConversationsForUser(w http.ResponseWriter, r *http.Request) {
// 	ps := httprouter.ParamsFromContext(r.Context())
// 	username := ps.ByName("username")

// 	convs, err := h.ConvSvc.ListConversations(r.Context(), username)
// 	if err != nil {
// 		http.Error(w, fmt.Sprintf("Failed to list conversations for %q: %v", username, err), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	_ = json.NewEncoder(w).Encode(convs)
// }

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

// CreateConversation handles POST /conversations per OpenAPI spec
func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// New payload shape
	var req struct {
		Recipient      string `json:"recipient"`
		InitialMessage struct {
			ContentType string `json:"contentType"`
			Text        string `json:"text,omitempty"`
			ContentURL  string `json:"contentUrl,omitempty"`
		} `json:"initialMessage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Call your new service method
	conv, msgID, err := h.ConvSvc.CreateWithMessage(
		r.Context(),
		service.ConversationTypeIndividual, // or parse a "type" field if you like
		me,
		req.Recipient,
		service.Message{
			ConversationID: "", // filled inside CreateWithMessage
			SenderUsername: me,
			ContentType:    req.InitialMessage.ContentType,
			Text:           req.InitialMessage.Text,
			ContentURL:     req.InitialMessage.ContentURL,
		},
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create conversation: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"conversation": conv,
		"messageId":    msgID,
	})
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

// MessageHandler handles /messages and related endpoints.
type MessageHandler struct {
	MsgSvc service.MessageService
}

// NewMessageHandler constructs a new MessageHandler.
func NewMessageHandler(svc service.MessageService) *MessageHandler {
	return &MessageHandler{MsgSvc: svc}
}

// SendMessageRequest is the payload for POST /messages
type SendMessageRequest struct {
	ConversationID string `json:"conversationId"`
	SenderUsername string `json:"senderUsername"`
	ContentType    string `json:"contentType"`
	Text           string `json:"text,omitempty"`
	ContentURL     string `json:"contentUrl,omitempty"`
}

// ----------------------------------------------------------------------------------------------------------
// SendMessage handles POST /messages.
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	msg := service.Message{
		ConversationID: req.ConversationID,
		SenderUsername: req.SenderUsername,
		ContentType:    req.ContentType,
		Text:           req.Text,
		ContentURL:     req.ContentURL,
	}
	created, err := h.MsgSvc.SendMessage(r.Context(), msg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send message: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(created)
}

// ListMessages handles GET /conversations/:id/messages.
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	convID := ps.ByName("id")

	msgs, err := h.MsgSvc.ListMessages(r.Context(), convID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list messages: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(msgs)
}

// GetMessage handles GET /messages/:id.
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	msgID := ps.ByName("id")

	msg, err := h.MsgSvc.GetMessage(r.Context(), msgID)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(msg)
}

// ForwardRequest is the payload for POST /messages/:id/forward
type ForwardRequest struct {
	ConversationID string `json:"conversationId"`
}

// ForwardMessage handles POST /messages/:id/forward.
func (h *MessageHandler) ForwardMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	origID := ps.ByName("id")

	var req ForwardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	msg, err := h.MsgSvc.ForwardMessage(r.Context(), origID, req.ConversationID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to forward message: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// ReplyRequest is the payload for POST /messages/:id/reply
type ReplyRequest struct {
	Text string `json:"text"`
}

// ReplyMessage handles POST /messages/:id/reply.
func (h *MessageHandler) ReplyMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	origID := ps.ByName("id")

	var req ReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	msg, err := h.MsgSvc.ReplyMessage(r.Context(), origID, req.Text)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to reply: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// ReactRequest is the payload for POST /messages/:id/reaction
type ReactRequest struct {
	Emoji    string `json:"emoji"`
	Username string `json:"username"`
}

// React handles POST /messages/:id/reaction.
func (h *MessageHandler) React(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	msgID := ps.ByName("id")

	var req ReactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.MsgSvc.React(r.Context(), msgID, req.Emoji, req.Username); err != nil {
		http.Error(w, fmt.Sprintf("Failed to react: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// -----------------------------------------------------------------
// GroupHandler handles /groups endpoints.
type GroupHandler struct {
	Gsvc service.GroupService
}

// NewGroupHandler constructs a GroupHandler.
func NewGroupHandler(svc service.GroupService) *GroupHandler {
	return &GroupHandler{Gsvc: svc}
}

// CreateGroupRequest is the payload for POST /groups.
type CreateGroupRequest struct {
	Name     string   `json:"name"`
	PhotoURL string   `json:"photoUrl,omitempty"`
	Members  []string `json:"members"`
}

// CreateGroup handles POST /groups.
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}
	grp, err := h.Gsvc.CreateGroup(r.Context(), req.Name, req.PhotoURL, req.Members)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create group: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(grp)
}

// ListGroups handles GET /groups.
func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	names, err := h.Gsvc.ListGroups(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list groups: %v", err), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(names)
}

// GetGroup handles GET /groups/:name.
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")

	grp, err := h.Gsvc.GetGroup(r.Context(), name)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(grp)
}

// AddMemberRequest is the payload for POST /groups/:name/members.
type AddMemberRequest struct {
	Username string `json:"username"`
}

// AddMember handles POST /groups/:name/members.
func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if err := h.Gsvc.AddMember(r.Context(), name, req.Username); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add member: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// RemoveMember handles DELETE /groups/:name/members/:username.
func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")
	username := ps.ByName("username")

	if err := h.Gsvc.RemoveMember(r.Context(), name, username); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove member: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// UpdatePhotoRequest is the payload for PATCH /groups/:name/photo.
type UpdatePhotoRequest struct {
	PhotoURL string `json:"photoUrl"`
}

// UpdatePhoto handles PATCH /groups/:name/photo.
func (h *GroupHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")

	var req UpdatePhotoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if req.PhotoURL == "" {
		http.Error(w, "photoUrl is required", http.StatusBadRequest)
		return
	}
	if err := h.Gsvc.UpdatePhoto(r.Context(), name, req.PhotoURL); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update photo: %v", err), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
