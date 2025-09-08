// service/conversation_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kk-Syuer/wasatext_2024/service/database"
)

// ConversationType indicates whether this is a one‐on‐one or group chat.
type ConversationType string

const (
	ConversationTypeIndividual ConversationType = "individual"
	ConversationTypeGroup      ConversationType = "group"
)

// Conversation holds the data for one chat.
type Conversation struct {
	ID           string
	Type         ConversationType
	Participants []string
	UpdatedAt    time.Time
}

// DeliveryStatusEntry represents the delivery status of one message to one recipient.
type DeliveryStatusEntry struct {
	MessageID string
	Recipient string
	Status    string    // "sent", "received", or "read"
	UpdatedAt time.Time // timestamp of the last status update
}

// ConversationService defines all conversation‐related business operations.
type ConversationService interface {
	// ListConversations returns every conversation the given user participates in.
	ListConversations(ctx context.Context, username string) ([]Conversation, error)
	// CreateConversation creates a new conversation of the given type with the given participants.
	CreateConversation(ctx context.Context, convType ConversationType, participants []string) (Conversation, error)
	// GetConversation fetches a single conversation by ID, including its participants.
	GetConversation(ctx context.Context, id string) (Conversation, error)
	// GetDeliveryStatus returns the per‐message delivery status for a conversation.
	GetDeliveryStatus(ctx context.Context, conversationID string) ([]DeliveryStatusEntry, error)
	CreateWithMessage(
		ctx context.Context,
		convoType ConversationType,
		userA, userB string,
		initial Message,
	) (Conversation, string, error)

	// GetMessageStatuses 返回会话中所有消息的投递状态
	GetMessageStatuses(ctx context.Context, conversationID string) ([]DeliveryStatusEntry, error)
}

func (s *conversationServiceImpl) CreateWithMessage(
	ctx context.Context,
	convoType ConversationType,
	userA, userB string,
	initial Message,
) (Conversation, string, error) {
	// 1) Upsert or verify both users exist (if you need to)
	//    …(left unchanged)…

	// 2) Find existing or create new conversation:
	conv, err := s.CreateConversation(ctx, convoType, []string{userA, userB})
	if err != nil {
		return Conversation{}, "", err
	}

	// 3) Send the initial message:
	msgSvc := NewMessageService(s.db)
	createdMsg, err := msgSvc.SendMessage(ctx, initial)
	if err != nil {
		return Conversation{}, "", err
	}

	// 4) Update the conversation’s updated_at:
	if err := s.db.UpdateConversationTimestamp(ctx, conv.ID, createdMsg.Timestamp.Format(time.RFC3339)); err != nil {
		return Conversation{}, "", err
	}

	// Return both the conversation and the new message’s ID
	return conv, createdMsg.ID, nil
}

type conversationServiceImpl struct {
	db *database.AppDatabase
}

// NewConversationService constructs a ConversationService backed by your AppDatabase.
func NewConversationService(db *database.AppDatabase) ConversationService {
	return &conversationServiceImpl{db: db}
}

func (s *conversationServiceImpl) ListConversations(ctx context.Context, username string) ([]Conversation, error) {
	ids, err := s.db.GetConversationsForUser(ctx, username)
	if err != nil {
		return nil, err
	}

	convs := make([]Conversation, 0, len(ids))
	for _, id := range ids {
		conv, err := s.GetConversation(ctx, id)
		if err != nil {
			// Skip conversations that have vanished or cannot be loaded
			continue
		}
		convs = append(convs, conv)
	}
	return convs, nil
}

func (s *conversationServiceImpl) CreateConversation(ctx context.Context, convType ConversationType, participants []string) (Conversation, error) {
	// 0) Look for an existing conversation with exactly these participants
	if existingID, err := s.db.FindConversationByParticipants(ctx, participants); err != nil {
		return Conversation{}, err
	} else if existingID != "" {
		// Return it without creating a new one
		return s.GetConversation(ctx, existingID)
	}

	// 1) Otherwise, create a fresh one
	id := uuid.New().String()
	now := time.Now().Format(time.RFC3339)

	if err := s.db.CreateConversation(ctx, id, string(convType), now); err != nil {
		return Conversation{}, err
	}
	for _, u := range participants {
		if err := s.db.AddParticipant(ctx, id, u); err != nil {
			return Conversation{}, err
		}
	}

	// 2) Load and return the newly created conversation
	return s.GetConversation(ctx, id)
}

func (s *conversationServiceImpl) GetConversation(ctx context.Context, id string) (Conversation, error) {
	typ, updatedAtStr, err := s.db.GetConversationByID(ctx, id)
	if err != nil {
		return Conversation{}, err
	}
	parts, err := s.db.GetConversationParticipants(ctx, id)
	if err != nil {
		return Conversation{}, err
	}

	updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
	if err != nil {
		// Fallback to zero‐time on parse error
		updatedAt = time.Time{}
	}

	return Conversation{
		ID:           id,
		Type:         ConversationType(typ),
		Participants: parts,
		UpdatedAt:    updatedAt,
	}, nil
}

func (s *conversationServiceImpl) GetDeliveryStatus(ctx context.Context, conversationID string) ([]DeliveryStatusEntry, error) {
	// Fetch raw rows from the database
	rows, err := s.db.GetDeliveryStatusForConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Map each row into our service‐level type
	var entries []DeliveryStatusEntry
	for _, r := range rows {
		ts, err := time.Parse(time.RFC3339, r.UpdatedAt)
		if err != nil {
			// If the timestamp is malformed, fall back to zero time
			ts = time.Time{}
		}
		entries = append(entries, DeliveryStatusEntry{
			MessageID: r.MessageID,
			Recipient: r.Recipient,
			Status:    r.Status,
			UpdatedAt: ts,
		})
	}

	return entries, nil
}

// GetMessageStatuses 返回会话中所有消息的投递状态
func (s *conversationServiceImpl) GetMessageStatuses(ctx context.Context, conversationID string) ([]DeliveryStatusEntry, error) {
	// 调用已有的 GetDeliveryStatus，处理可能的“未找到”错误
	statuses, err := s.GetDeliveryStatus(ctx, conversationID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return statuses, nil
}
