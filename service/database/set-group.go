package database

import "context"

func (adb *AppDatabase) CreateGroup(
	ctx context.Context,
	groupName, photoURL, createdAt, conversationID string,
) error {
	const stmt = `
    INSERT INTO groups (group_name, photo_url, created_at, conversation_id)
         VALUES (?, ?, ?, ?)
    `
	_, err := adb.db.ExecContext(ctx, stmt, groupName, photoURL, createdAt, conversationID)
	return err
}

// AddGroupMember adds a user to a group.
func (adb *AppDatabase) AddGroupMember(
	ctx context.Context,
	groupName, username string,
) error {
	const stmt = `
	INSERT INTO group_members (group_name, username)
	     VALUES (?, ?)
	`
	_, err := adb.db.ExecContext(ctx, stmt, groupName, username)
	return err
}

// RemoveGroupMember removes a user from a group.
func (adb *AppDatabase) RemoveGroupMember(
	ctx context.Context,
	groupName, username string,
) error {
	const stmt = `
	DELETE FROM group_members
	WHERE group_name = ? AND username = ?
	`
	_, err := adb.db.ExecContext(ctx, stmt, groupName, username)
	return err
}

// UpdateGroupPhoto changes the photo_url for a group.
func (adb *AppDatabase) UpdateGroupPhoto(
	ctx context.Context,
	groupName, photoURL string,
) error {
	const stmt = `
	UPDATE groups
	   SET photo_url = ?
	 WHERE group_name = ?
	`
	res, err := adb.db.ExecContext(ctx, stmt, photoURL, groupName)
	if err != nil {
		return err
	}
	// Optionally check that a row was updated
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrNotFound
	}
	return nil
}
