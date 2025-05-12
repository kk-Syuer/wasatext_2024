package database

import (
	"context"
	"database/sql"
)

// GetConversationsForUser returns all conversation IDs that a given user participates in.
func (adb *AppDatabase) GetConversationsForUser(ctx context.Context, username string) ([]string, error) {
	rows, err := adb.db.QueryContext(ctx,
		`SELECT conversation_id 
		   FROM conversation_participants 
		  WHERE username = ?`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ids, nil
}

// GetConversationByID returns the type and updated_at timestamp for a conversation.
func (adb *AppDatabase) GetConversationByID(ctx context.Context, id string) (convType string, updatedAt string, err error) {
	err = adb.db.QueryRowContext(ctx,
		`SELECT type, updated_at 
		   FROM conversations 
		  WHERE id = ?`,
		id,
	).Scan(&convType, &updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", ErrNotFound
		}
		return "", "", err
	}
	return convType, updatedAt, nil
}

// GetConversationParticipants returns the usernames of all participants in a conversation.
func (adb *AppDatabase) GetConversationParticipants(ctx context.Context, conversationID string) ([]string, error) {
	rows, err := adb.db.QueryContext(ctx,
		`SELECT username 
		   FROM conversation_participants 
		  WHERE conversation_id = ?`,
		conversationID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return users, nil
}
