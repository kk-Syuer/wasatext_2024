package database

import (
	"context"
	"database/sql"
)

// GetName retrieves the display name for the given username.
// Returns ErrNotFound if no such user exists.
func (adb *AppDatabase) GetName(ctx context.Context, username string) (string, error) {
	var name string
	err := adb.db.QueryRowContext(ctx,
		"SELECT name FROM users WHERE username = ?",
		username,
	).Scan(&name)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return name, nil
}

// GetPhoto retrieves the photo_url for the given username.
// Returns ErrNotFound if no such user exists.
func (adb *AppDatabase) GetPhoto(ctx context.Context, username string) (string, error) {
	var photo string
	err := adb.db.QueryRowContext(ctx,
		"SELECT photo_url FROM users WHERE username = ?",
		username,
	).Scan(&photo)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return photo, nil
}

// GetAllUsernames returns a slice of all usernames in the users table.
func (adb *AppDatabase) GetAllUsernames(ctx context.Context) ([]string, error) {
	rows, err := adb.db.QueryContext(ctx,
		"SELECT username FROM users",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usernames []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		usernames = append(usernames, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return usernames, nil
}
