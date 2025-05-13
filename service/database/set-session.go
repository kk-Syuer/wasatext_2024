package database

import (
	"context"
)

func (adb *AppDatabase) CreateSession(ctx context.Context, id, username, createdAt string) error {
	const stmt = `
      INSERT INTO sessions (id, username, created_at)
           VALUES (?, ?, ?)
    `
	_, err := adb.db.ExecContext(ctx, stmt, id, username, createdAt)
	return err
}
