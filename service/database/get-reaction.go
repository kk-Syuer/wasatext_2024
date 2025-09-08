package database

import (
	"context"
)

// ReactionRow mirrors the reactions table.
type ReactionRow struct {
	ID           string
	MessageID    string
	Emoji        string
	UserUsername string
	CreatedAt    string
}

// GetReactionsForMessage returns all reactions on a given message.
func (adb *AppDatabase) GetReactionsForMessage(ctx context.Context, messageID string) ([]ReactionRow, error) {
	const query = `
    SELECT id, message_id, emoji, user_username, created_at
      FROM reactions
     WHERE message_id = ?
     ORDER BY created_at ASC
    `
	rows, err := adb.db.QueryContext(ctx, query, messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ReactionRow
	for rows.Next() {
		var r ReactionRow
		if err := rows.Scan(&r.ID, &r.MessageID, &r.Emoji, &r.UserUsername, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}
