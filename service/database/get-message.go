package database

import (
	"context"
	"database/sql"
)

// MessageRow mirrors the messages table, with nullable columns coalesced to empty strings.
type MessageRow struct {
	ID                 string
	ConversationID     string
	SenderUsername     string
	ContentType        string
	ContentURL         string
	Text               string
	Timestamp          string
	ReplyTo            string
	ForwardedFrom      string
	ForwardedTimestamp string
	OriginalContent    string
}

// GetMessageByID retrieves a single message by its ID.
func (adb *AppDatabase) GetMessageByID(ctx context.Context, id string) (MessageRow, error) {
	const query = `
	SELECT 
	  id,
	  conversation_id,
	  sender_username,
	  content_type,
	  COALESCE(content_url, ''),
	  COALESCE(text, ''),
	  timestamp,
	  COALESCE(reply_to, ''),
	  COALESCE(forwarded_from, ''),
	  COALESCE(forwarded_timestamp, ''),
	  COALESCE(original_content, '')
	FROM messages
	WHERE id = ?
	`

	var m MessageRow
	err := adb.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID,
		&m.ConversationID,
		&m.SenderUsername,
		&m.ContentType,
		&m.ContentURL,
		&m.Text,
		&m.Timestamp,
		&m.ReplyTo,
		&m.ForwardedFrom,
		&m.ForwardedTimestamp,
		&m.OriginalContent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return MessageRow{}, ErrNotFound
		}
		return MessageRow{}, err
	}
	return m, nil
}

// GetMessagesForConversation returns all messages in a conversation, ordered by timestamp.
func (adb *AppDatabase) GetMessagesForConversation(ctx context.Context, conversationID string) ([]MessageRow, error) {
	const query = `
	SELECT 
	  id,
	  conversation_id,
	  sender_username,
	  content_type,
	  COALESCE(content_url, ''),
	  COALESCE(text, ''),
	  timestamp,
	  COALESCE(reply_to, ''),
	  COALESCE(forwarded_from, ''),
	  COALESCE(forwarded_timestamp, ''),
	  COALESCE(original_content, '')
	FROM messages
	WHERE conversation_id = ?
	ORDER BY timestamp ASC
	`

	rows, err := adb.db.QueryContext(ctx, query, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []MessageRow
	for rows.Next() {
		var m MessageRow
		if err := rows.Scan(
			&m.ID,
			&m.ConversationID,
			&m.SenderUsername,
			&m.ContentType,
			&m.ContentURL,
			&m.Text,
			&m.Timestamp,
			&m.ReplyTo,
			&m.ForwardedFrom,
			&m.ForwardedTimestamp,
			&m.OriginalContent,
		); err != nil {
			return nil, err
		}
		msgs = append(msgs, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgs, nil
}
