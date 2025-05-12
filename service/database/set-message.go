package database

import (
	"context"
)

// CreateMessage inserts a new message row into the messages table.
func (adb *AppDatabase) CreateMessage(ctx context.Context, m MessageRow) error {
	const stmt = `
	INSERT INTO messages (
	  id,
	  conversation_id,
	  sender_username,
	  content_type,
	  content_url,
	  text,
	  timestamp,
	  reply_to,
	  forwarded_from,
	  forwarded_timestamp,
	  original_content
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := adb.db.ExecContext(ctx, stmt,
		m.ID,
		m.ConversationID,
		m.SenderUsername,
		m.ContentType,
		m.ContentURL,
		m.Text,
		m.Timestamp,
		m.ReplyTo,
		m.ForwardedFrom,
		m.ForwardedTimestamp,
		m.OriginalContent,
	)
	return err
}
