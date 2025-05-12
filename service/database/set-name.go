package database

import (
	"context"
)

// SetName updates the name field for a user.
// Returns ErrNotFound if no rows were affected.
func (adb *AppDatabase) SetName(ctx context.Context, username, newName string) error {
	res, err := adb.db.ExecContext(ctx,
		"UPDATE users SET name = ? WHERE username = ?",
		newName, username,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// SetPhoto updates the photo_url field for a user.
// Returns ErrNotFound if no rows were affected.
func (adb *AppDatabase) SetPhoto(ctx context.Context, username, photoURL string) error {
	res, err := adb.db.ExecContext(ctx,
		"UPDATE users SET photo_url = ? WHERE username = ?",
		photoURL, username,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
