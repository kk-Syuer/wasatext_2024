package database

import (
	"context"
	"database/sql"
	"errors"
)

// UserRow mirrors the users table.
type UserRow struct {
	ID       string
	Username string
	Name     string
	PhotoURL string
	JoinedAt string
}

// ErrUserNotFound is returned when a username lookup fails.
var ErrUserNotFound = errors.New("user not found")

// GetUser looks up a user by username.
func (adb *AppDatabase) GetUser(ctx context.Context, username string) (UserRow, error) {
	const query = `
    SELECT id, username, name, photo_url, joined_at
      FROM users
     WHERE username = ?
    `
	var u UserRow
	err := adb.db.QueryRowContext(ctx, query, username).
		Scan(&u.ID, &u.Username, &u.Name, &u.PhotoURL, &u.JoinedAt)
	if err == sql.ErrNoRows {
		return UserRow{}, ErrUserNotFound
	}
	return u, err
}
