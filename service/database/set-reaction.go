package database

import (
	"context"
)

// AddReaction inserts a new reaction.
func (adb *AppDatabase) AddReaction(ctx context.Context, r ReactionRow) error {
	const stmt = `
    INSERT INTO reactions (id, message_id, emoji, user_username, created_at)
         VALUES (?, ?, ?, ?, ?)
    `
	_, err := adb.db.ExecContext(ctx, stmt,
		r.ID, r.MessageID, r.Emoji, r.UserUsername, r.CreatedAt,
	)
	return err
}

// RemoveReaction deletes a reaction by its ID.
func (adb *AppDatabase) RemoveReaction(ctx context.Context, reactionID string) error {
	const stmt = `
    DELETE FROM reactions
     WHERE id = ?
    `
	res, err := adb.db.ExecContext(ctx, stmt, reactionID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}
