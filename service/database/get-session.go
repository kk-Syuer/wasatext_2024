package database

import (
	"context"
	"database/sql"
	"errors"
)

var ErrSessionNotFound = errors.New("session not found")

func (adb *AppDatabase) GetSessionUsername(ctx context.Context, id string) (string, error) {
	const q = `SELECT username FROM sessions WHERE id = ?`
	var u string
	err := adb.db.QueryRowContext(ctx, q, id).Scan(&u)
	if err == sql.ErrNoRows {
		return "", ErrSessionNotFound
	}
	return u, err
}
