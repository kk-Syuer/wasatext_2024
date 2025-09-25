// service/message_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
)

// Message represents a chat message.
type Message struct {
	ID                 string    `json:"id"`
	ConversationID     string    `json:"conversationId"`
	SenderUsername     string    `json:"senderUsername"`
	ContentType        string    `json:"contentType"` // e.g. "text", "image"
	ContentURL         string    `json:"contentUrl,omitempty"`
	Text               string    `json:"text,omitempty"`
	Timestamp          time.Time `json:"timestamp"`
	ReplyTo            string    `json:"replyTo,omitempty"`
	ForwardedFrom      string    `json:"forwardedFrom,omitempty"`
	ForwardedTimestamp time.Time `json:"forwardedTimestamp,omitempty"`
	OriginalContent    string    `json:"originalContent,omitempty"`
	// Reactions can be filled in later once you add DB support
	Reactions []Reaction `json:"reactions,omitempty"`
}

// Reaction represents an emoji reaction to a message.
type Reaction struct {
	ID        string    `json:"id"`
	MessageID string    `json:"messageId"`
	Emoji     string    `json:"emoji"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"createdAt"`
}

// MessageService defines operations on chat messages.
type MessageService interface {
	// SendMessage creates and stores a new message.
	SendMessage(ctx context.Context, msg Message) (Message, error)
	// GetMessage retrieves one message by its ID.
	GetMessage(ctx context.Context, id string) (Message, error)
	// ListMessages returns all messages in a conversation.
	ListMessages(ctx context.Context, conversationID string) ([]Message, error)
	// ForwardMessage forwards an existing message into another conversation.
	ForwardMessage(ctx context.Context, originalMessageID, toConversationID string) (Message, error)
	// ReplyMessage creates a reply to an existing message.
	ReplyMessage(ctx context.Context, originalMessageID, replyText string) (Message, error)
	// React adds or removes a reaction to a message.
	React(ctx context.Context, messageID, emoji, username string) error
	// DeleteMessage 删除指定 ID 的消息
	DeleteMessage(ctx context.Context, messageID string) error
	Unreact(ctx context.Context, messageID, username string) error
	// MarkConversationRead marks everything up to now as read for this user.
	MarkConversationRead(ctx context.Context, conversationID, username string) error
	MarkConversationReadAt(ctx context.Context, conversationID, username string, at time.Time) error
}

type messageServiceImpl struct {
	db *database.AppDatabase
}

func (s *messageServiceImpl) Unreact(ctx context.Context, messageID, username string) error {
	return s.db.RemoveReaction(ctx, messageID, username)
}

// NewMessageService constructs a MessageService backed by the given AppDatabase.
func NewMessageService(db *database.AppDatabase) MessageService {
	return &messageServiceImpl{db: db}
}

func (s *messageServiceImpl) SendMessage(ctx context.Context, msg Message) (Message, error) {
	// Assign ID and timestamp
	msg.ID = uuid.New().String()
	msg.Timestamp = globaltime.Now()

	// Prepare DB row
	row := database.MessageRow{
		ID:                 msg.ID,
		ConversationID:     msg.ConversationID,
		SenderUsername:     msg.SenderUsername,
		ContentType:        msg.ContentType,
		ContentURL:         msg.ContentURL,
		Text:               msg.Text,
		Timestamp:          msg.Timestamp.Format(time.RFC3339),
		ReplyTo:            msg.ReplyTo,
		ForwardedFrom:      msg.ForwardedFrom,
		ForwardedTimestamp: msg.ForwardedTimestamp.Format(time.RFC3339),
		OriginalContent:    msg.OriginalContent,
	}

	if err := s.db.CreateMessage(ctx, row); err != nil {
		return Message{}, err
	}
	return msg, nil
}

func (s *messageServiceImpl) GetMessage(ctx context.Context, id string) (Message, error) {
	row, err := s.db.GetMessageByID(ctx, id)
	if err != nil {
		return Message{}, err
	}

	// Parse timestamps
	ts, _ := time.Parse(time.RFC3339, row.Timestamp)
	fts, _ := time.Parse(time.RFC3339, row.ForwardedTimestamp)

	msg := Message{
		ID:                 row.ID,
		ConversationID:     row.ConversationID,
		SenderUsername:     row.SenderUsername,
		ContentType:        row.ContentType,
		ContentURL:         row.ContentURL,
		Text:               row.Text,
		Timestamp:          ts,
		ReplyTo:            row.ReplyTo,
		ForwardedFrom:      row.ForwardedFrom,
		ForwardedTimestamp: fts,
		OriginalContent:    row.OriginalContent,
	}

	// attach reactions
	if rrows, rerr := s.db.GetReactionsForMessage(ctx, id); rerr == nil {
		msg.Reactions = make([]Reaction, 0, len(rrows))
		for _, r := range rrows {
			cat, _ := time.Parse(time.RFC3339, r.CreatedAt)
			msg.Reactions = append(msg.Reactions, Reaction{
				ID:        r.ID,
				MessageID: r.MessageID,
				Emoji:     r.Emoji,
				Username:  r.UserUsername,
				CreatedAt: cat,
			})
		}
	}

	return msg, nil
}

func (s *messageServiceImpl) ListMessages(ctx context.Context, conversationID string) ([]Message, error) {
	rows, err := s.db.GetMessagesForConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	// batch reactions once for the whole conversation
	rmap, _ := s.db.GetReactionsForConversation(ctx, conversationID)
	msgs := make([]Message, 0, len(rows))
	for _, row := range rows {
		ts, _ := time.Parse(time.RFC3339, row.Timestamp)
		fts, _ := time.Parse(time.RFC3339, row.ForwardedTimestamp)

		m := Message{
			ID:                 row.ID,
			ConversationID:     row.ConversationID,
			SenderUsername:     row.SenderUsername,
			ContentType:        row.ContentType,
			ContentURL:         row.ContentURL,
			Text:               row.Text,
			Timestamp:          ts,
			ReplyTo:            row.ReplyTo,
			ForwardedFrom:      row.ForwardedFrom,
			ForwardedTimestamp: fts,
			OriginalContent:    row.OriginalContent,
		}

		// attach reactions if any
		if rr, ok := rmap[row.ID]; ok && len(rr) > 0 {
			m.Reactions = make([]Reaction, 0, len(rr))
			for _, r := range rr {
				cat, _ := time.Parse(time.RFC3339, r.CreatedAt)
				m.Reactions = append(m.Reactions, Reaction{
					ID:        r.ID,
					MessageID: r.MessageID,
					Emoji:     r.Emoji,
					Username:  r.UserUsername,
					CreatedAt: cat,
				})
			}
		}

		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (s *messageServiceImpl) ForwardMessage(ctx context.Context, originalMessageID, toConversationID string) (Message, error) {
	// Load the original message
	orig, err := s.GetMessage(ctx, originalMessageID)
	if err != nil {
		return Message{}, err
	}

	// Create a new message row that forwards the original
	forward := Message{
		ConversationID:     toConversationID,
		SenderUsername:     orig.SenderUsername,
		ContentType:        orig.ContentType,
		ContentURL:         orig.ContentURL,
		Text:               orig.Text,
		Timestamp:          globaltime.Now(),
		ReplyTo:            "",
		ForwardedFrom:      orig.SenderUsername,
		ForwardedTimestamp: orig.Timestamp,
		OriginalContent:    orig.Text,
	}

	return s.SendMessage(ctx, forward)
}

func (s *messageServiceImpl) ReplyMessage(ctx context.Context, originalMessageID, replyText string) (Message, error) {
	// Load the original message to determine conversation
	orig, err := s.GetMessage(ctx, originalMessageID)
	if err != nil {
		return Message{}, err
	}

	// Create reply message in same conversation
	reply := Message{
		ConversationID: orig.ConversationID,
		SenderUsername: orig.SenderUsername, // or get from session
		ContentType:    "text",
		Text:           replyText,
		Timestamp:      globaltime.Now(),
		ReplyTo:        orig.ID,
	}
	return s.SendMessage(ctx, reply)
}

func (s *messageServiceImpl) React(ctx context.Context, messageID, emoji, username string) error {
	// 1) Generate a new reaction ID and timestamp
	id := uuid.New().String()
	now := globaltime.Now().Format(time.RFC3339)

	// 2) Insert it
	return s.db.AddReaction(ctx, database.ReactionRow{
		ID:           id,
		MessageID:    messageID,
		Emoji:        emoji,
		UserUsername: username,
		CreatedAt:    now,
	})
}

// DeleteMessage 删除一条消息；若 DB 返回 ErrNotFound，则映射为 service.ErrNotFound
func (s *messageServiceImpl) DeleteMessage(ctx context.Context, messageID string) error {
	err := s.db.DeleteMessage(ctx, messageID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *messageServiceImpl) MarkConversationRead(ctx context.Context, conversationID, username string) error {
	return s.db.UpsertConversationRead(ctx, conversationID, username, globaltime.Now())
}

func (s *messageServiceImpl) MarkConversationReadAt(ctx context.Context, conversationID, username string, at time.Time) error {
	return s.db.UpsertConversationRead(ctx, conversationID, username, at)
}
