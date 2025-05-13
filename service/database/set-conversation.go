package database

import "context"

// CreateConversation inserts a new conversation record with the given ID, type, and timestamp.
func (adb *AppDatabase) CreateConversation(ctx context.Context, id, convType, updatedAt string) error {
	_, err := adb.db.ExecContext(ctx,
		`INSERT INTO conversations (id, type, updated_at)
		       VALUES (?, ?, ?)`,
		id, convType, updatedAt,
	)
	return err
}

// AddParticipant adds a user into a conversation’s participant list.
func (adb *AppDatabase) AddParticipant(ctx context.Context, conversationID, username string) error {
	_, err := adb.db.ExecContext(ctx,
		`INSERT INTO conversation_participants (conversation_id, username)
		       VALUES (?, ?)`,
		conversationID, username,
	)
	return err
}

func (adb *AppDatabase) RemoveParticipant(ctx context.Context, conversationID, username string) error {
	_, err := adb.db.ExecContext(ctx,
		`DELETE FROM conversation_participants WHERE conversation_id=? AND username=?`,
		conversationID, username,
	)
	return err
}
