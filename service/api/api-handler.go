package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
	"github.com/kk-Syuer/wasatext_2024/service"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

/* ------------------------- SESSION ------------------------- */

type SessionHandler struct {
	SessionService service.SessionService
}

func NewSessionHandler(svc service.SessionService) *SessionHandler {
	return &SessionHandler{SessionService: svc}
}

type LoginRequest struct {
	Username string `json:"username"`
}

type LoginResponse struct {
	Identifier string `json:"identifier"`
	Username   string `json:"username"`
}

func (h *SessionHandler) DoLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	token, err := h.SessionService.Login(r.Context(), req.Username)
	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}
	resp := LoginResponse{Identifier: token, Username: req.Username}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

/* --------------------------- USERS ------------------------- */

type UserHandler struct {
	UserService service.UserService
}

func NewUserHandler(svc service.UserService) *UserHandler { return &UserHandler{UserService: svc} }

// GET /users  -> { "usernames": [...] }
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserService.ListUsers(r.Context())
	if err != nil {
		http.Error(w, "Failed to list users", http.StatusInternalServerError)
		return
	}
	out := struct {
		Usernames []string `json:"usernames"`
	}{Usernames: make([]string, 0, len(users))}
	for _, u := range users {
		out.Usernames = append(out.Usernames, u.Username)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(out)
}

// PATCH /user/name  body: { "username": "<new>" }  -> { "username": "<new>" }
func (h *UserHandler) UpdateMyName(w http.ResponseWriter, r *http.Request) {
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Username) == "" {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	newName := strings.TrimSpace(body.Username)

	// basic constraints (align with your frontend)
	if len(newName) < 3 || len(newName) > 16 {
		http.Error(w, "Username must be 3-16 characters", http.StatusBadRequest)
		return
	}
	for _, r := range newName {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-') {
			http.Error(w, "Username may contain letters, numbers, and hyphen only", http.StatusBadRequest)
			return
		}
	}

	if err := h.UserService.UpdateName(r.Context(), me, newName); err != nil {
		// Log precise cause for troubleshooting
		log.Printf("rename %q -> %q failed: %v", me, newName, err)

		low := strings.ToLower(err.Error())
		switch {
		case strings.Contains(low, "already taken"), strings.Contains(low, "unique"):
			http.Error(w, "Username already taken", http.StatusConflict)
		case strings.Contains(low, "not found"):
			http.Error(w, "User not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to update name", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Username string `json:"username"`
	}{newName})
}

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

// PATCH /user/photo  multipart field: photo  -> { "photoUrl": "<url>" }
func (h *UserHandler) UpdateMyPhoto(w http.ResponseWriter, r *http.Request) {
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid multipart payload", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	name := uuid.New().String() + ext
	if err := os.MkdirAll("./uploads", 0o755); err != nil {
		http.Error(w, "failed to create upload dir", http.StatusInternalServerError)
		return
	}
	dst, err := os.Create(filepath.Join("./uploads", name))
	if err != nil {
		http.Error(w, "failed to store file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "failed to save file", http.StatusInternalServerError)
		return
	}

	// Save a relative URL; the frontend will absolutize with fullUrl()
	photoURL := "/uploads/" + name

	if err := h.UserService.UpdatePhoto(r.Context(), me, photoURL); err != nil {
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		PhotoURL string `json:"photoUrl"`
	}{photoURL})
}

/* ---------------------- CONVERSATIONS ----------------------- */

type ConversationHandler struct {
	ConvSvc service.ConversationService
}

func NewConversationHandler(svc service.ConversationService) *ConversationHandler {
	return &ConversationHandler{ConvSvc: svc}
}

// POST /conversations  body: { "type":"individual", "recipient":"...", "initialMessage":"..." }
func (h *ConversationHandler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Type           string `json:"type"`
		Recipient      string `json:"recipient"`
		InitialMessage string `json:"initialMessage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil ||
		body.Type != "individual" || body.Recipient == "" || body.InitialMessage == "" {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}

	conv, _, err := h.ConvSvc.CreateWithMessage(
		r.Context(),
		service.ConversationTypeIndividual,
		me, body.Recipient,
		service.Message{SenderUsername: me, ContentType: "text", Text: body.InitialMessage},
	)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.Error(w, "Recipient not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(conv)
}

// GET /conversations
func (h *ConversationHandler) ListConversations(w http.ResponseWriter, r *http.Request) {
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	convs, err := h.ConvSvc.ListConversations(r.Context(), me)
	if err != nil {
		http.Error(w, "Failed to list conversations", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Conversations []service.Conversation `json:"conversations"`
	}{Conversations: convs})
}

// GET /conversations/:id
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

// GET /conversations/:id/messages/status
func (h *ConversationHandler) GetMessageStatuses(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	id := ps.ByName("id")
	statuses, err := h.ConvSvc.GetMessageStatuses(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get statuses", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Statuses []service.DeliveryStatusEntry `json:"statuses"`
	}{Statuses: statuses})
}

/* -------------------------- MESSAGES ------------------------ */

type MessageHandler struct {
	MsgSvc service.MessageService
}

func NewMessageHandler(svc service.MessageService) *MessageHandler {
	return &MessageHandler{MsgSvc: svc}
}

// POST /messages  (multipart: conversationId, contentType, text?, file?)
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Invalid multipart", http.StatusBadRequest)
		return
	}
	convID := r.FormValue("conversationId")
	ctype := r.FormValue("contentType")
	if convID == "" || (ctype != "text" && ctype != "image" && ctype != "gif") {
		http.Error(w, "Invalid fields", http.StatusBadRequest)
		return
	}

	msg := service.Message{
		ConversationID: convID,
		SenderUsername: me,
		ContentType:    ctype,
	}

	if ctype == "text" {
		txt := r.FormValue("text")
		if txt == "" {
			http.Error(w, "text required", http.StatusBadRequest)
			return
		}
		msg.Text = txt
	} else {
		f, hdr, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "file required", http.StatusBadRequest)
			return
		}
		defer f.Close()
		ext := filepath.Ext(hdr.Filename)
		if ext == "" {
			ext = ".bin"
		}
		fname := uuid.New().String() + ext
		if err := os.MkdirAll("./uploads", 0o755); err != nil {
			http.Error(w, "failed to create upload dir", http.StatusInternalServerError)
			return
		}
		dst, err := os.Create(filepath.Join("./uploads", fname))
		if err != nil {
			http.Error(w, "failed to store file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()
		if _, err := io.Copy(dst, f); err != nil {
			http.Error(w, "failed to save file", http.StatusInternalServerError)
			return
		}
		msg.ContentURL = "/uploads/" + fname

		// Optional caption (allowed for image/gif)
		if cap := r.FormValue("text"); strings.TrimSpace(cap) != "" {
			msg.Text = cap
		}
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

// GET /messages/:id
func (h *MessageHandler) GetMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	id := ps.ByName("id")
	msg, err := h.MsgSvc.GetMessage(r.Context(), id)
	if err != nil {
		http.Error(w, "Message not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(msg)
}

// GET /conversations/:id/messages  -> { "messages": [...] }
func (h *MessageHandler) ListMessages(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	convID := ps.ByName("id")
	user := UsernameFromContext(r.Context())

	// 1) Load messages
	msgs, err := h.MsgSvc.ListMessages(r.Context(), convID)
	if err != nil {
		http.Error(w, "Failed to list messages", http.StatusInternalServerError)
		return
	}

	// 2) Mark "read" up to the newest message we actually returned
	if user != "" {
		var at time.Time
		if len(msgs) > 0 {
			// DB returns DESC by timestamp; the first is newest
			at = msgs[0].Timestamp
		} else {
			at = time.Now().UTC()
		}
		_ = h.MsgSvc.MarkConversationReadAt(r.Context(), convID, user, at)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Messages []service.Message `json:"messages"`
	}{Messages: msgs})
}

// DELETE /messages/:id
func (h *MessageHandler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	id := ps.ByName("id")
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	if err := h.MsgSvc.DeleteMessage(r.Context(), id, me); err != nil {
		if errors.Is(err, service.ErrForbidden) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /messages/:id/forward  body: { "targetConversationId": "<id>" }
func (h *MessageHandler) ForwardMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	orig := ps.ByName("id")
	var body struct {
		TargetConversationID string `json:"targetConversationId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TargetConversationID == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	msg, err := h.MsgSvc.ForwardMessage(r.Context(), orig, body.TargetConversationID, me)
	if err != nil {
		http.Error(w, "Forward failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// POST /messages/:id/reply  body: { "text": "..." }
func (h *MessageHandler) ReplyMessage(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	parentID := ps.ByName("id")

	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// service expects (ctx, parentID, text)
	msg, err := h.MsgSvc.ReplyMessage(r.Context(), parentID, me, body.Text)
	if err != nil {
		http.Error(w, "Reply failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(msg)
}

// POST /messages/:id/reaction  body: { "emoji": "😀" } -> 201 Reaction
func (h *MessageHandler) React(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	msgID := ps.ByName("id")
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var body struct {
		Emoji string `json:"emoji"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Emoji == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.MsgSvc.React(r.Context(), msgID, body.Emoji, me); err != nil {
		http.Error(w, "React failed", http.StatusInternalServerError)
		return
	}
	out := struct {
		ID        string `json:"id"`
		MessageID string `json:"messageId"`
		Emoji     string `json:"emoji"`
		User      string `json:"user"`
		CreatedAt string `json:"createdAt"`
	}{
		ID:        uuid.New().String(),
		MessageID: msgID,
		Emoji:     body.Emoji,
		User:      me,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *MessageHandler) Unreact(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	msgID := ps.ByName("id")
	_ = ps.ByName("reactionId") // optional, not needed with current DB model

	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := h.MsgSvc.Unreact(r.Context(), msgID, me); err != nil {
		// database.RemoveReaction returns database.ErrNotFound when nothing to remove
		if errors.Is(err, database.ErrNotFound) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Unreact failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

/* ---------------------------- GROUPS ------------------------ */

type GroupHandler struct {
	Gsvc service.GroupService
	Msg  service.MessageService
}

func NewGroupHandler(gsvc service.GroupService, msg service.MessageService) *GroupHandler {
	return &GroupHandler{Gsvc: gsvc, Msg: msg}
}

// POST /groups
// body: { "groupName": "...", "members": ["a","b",...], "initialMessage": "..." }
// inside api-handler.go
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	var body struct {
		GroupName      string   `json:"groupName"`
		Members        []string `json:"members"`
		InitialMessage string   `json:"initialMessage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// groupName required; initialMessage is mandatory per your spec
	name := strings.TrimSpace(body.GroupName)
	initial := strings.TrimSpace(body.InitialMessage)
	if name == "" || initial == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Build members = (unique of body.Members) ∪ {creator}
	creator := UsernameFromContext(r.Context())
	seen := make(map[string]struct{}, len(body.Members)+1)
	all := make([]string, 0, len(body.Members)+1)

	add := func(u string) {
		u = strings.TrimSpace(u)
		if u == "" {
			return
		}
		if _, ok := seen[u]; ok {
			return
		}
		seen[u] = struct{}{}
		all = append(all, u)
	}

	for _, m := range body.Members {
		add(m)
	}
	add(creator) // ensure creator is a member

	if len(all) < 2 {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Create the group (service returns ConversationID)
	grp, err := h.Gsvc.CreateGroup(r.Context(), name, "", all)
	if err != nil {
		low := strings.ToLower(err.Error())
		switch {
		case strings.Contains(low, "unique") || strings.Contains(low, "constraint failed") && strings.Contains(low, "groups.name"):
			http.Error(w, "Group name already exists", http.StatusConflict) // 409
			return
		case strings.Contains(low, "foreign key"), strings.Contains(low, "no such user"):
			http.Error(w, "One or more members not found", http.StatusNotFound) // 404
			return
		default:
			http.Error(w, "Failed to create group", http.StatusInternalServerError) // 500
			return
		}
	}

	// Post the mandatory initial message into the new conversation
	if h.Msg != nil && grp.ConversationID != "" {
		if _, err := h.Msg.SendMessage(r.Context(), service.Message{
			ConversationID: grp.ConversationID,
			SenderUsername: creator,
			ContentType:    "text",
			Text:           initial,
		}); err != nil {
			http.Error(w, "Failed to send initial message", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(grp)
}

func (h *GroupHandler) ListGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.Gsvc.ListGroupsDetailed(r.Context())
	if err != nil {
		http.Error(w, "Failed to list groups", http.StatusInternalServerError)
		return
	}
	// Response shape expected by your frontend helper: either an array or { groups: [...] }
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"groups": groups,
	})
}

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

// POST /groups/:name/members  body: { "username": "<user>" } -> { "username": "<user>" }
func (h *GroupHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")
	var body struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Username == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := h.Gsvc.AddMember(r.Context(), name, body.Username); err != nil {
		http.Error(w, "Failed to add member", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Username string `json:"username"`
	}{body.Username})
}

func (h *GroupHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")
	username := ps.ByName("username")
	if err := h.Gsvc.RemoveMember(r.Context(), name, username); err != nil {
		http.Error(w, "Failed to remove member", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PATCH /groups/:name/photo  multipart field: photo  -> { "photoUrl": "<url>" }
func (h *GroupHandler) UpdatePhoto(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Invalid multipart payload", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		http.Error(w, "photo file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	filename := uuid.New().String() + ext
	if err := os.MkdirAll("./uploads", 0o755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}
	dst, err := os.Create(filepath.Join("./uploads", filename))
	if err != nil {
		http.Error(w, "Failed to store photo", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save photo", http.StatusInternalServerError)
		return
	}

	photoURL := "/uploads/" + filename

	if err := h.Gsvc.UpdatePhoto(r.Context(), name, photoURL); err != nil {
		http.Error(w, "Failed to update photo", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		PhotoURL string `json:"photoUrl"`
	}{photoURL})
}

func (h *GroupHandler) LeaveGroup(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	name := ps.ByName("name")
	me := UsernameFromContext(r.Context())
	if me == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Implementation detail: remove member, delete their reactions, keep messages = handled in service/db if needed
	if err := h.Gsvc.RemoveMember(r.Context(), name, me); err != nil {
		http.Error(w, "Failed to leave group", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// api-handler.go
func (h *GroupHandler) UpdateName(w http.ResponseWriter, r *http.Request) {
	ps := httprouter.ParamsFromContext(r.Context())
	old := ps.ByName("name")

	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Name) == "" {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	newName := strings.TrimSpace(body.Name)

	if err := h.Gsvc.UpdateName(r.Context(), old, newName); err != nil {
		low := strings.ToLower(err.Error())
		switch {
		case strings.Contains(low, "unique"), strings.Contains(low, "constraint failed"):
			http.Error(w, "Group name already exists", http.StatusConflict)
		case strings.Contains(low, "not found"):
			http.Error(w, "Group not found", http.StatusNotFound)
		default:
			http.Error(w, "Failed to rename group", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Name string `json:"name"`
	}{newName})
}
