package database

import (
	"context"
	"database/sql"
)

// GroupRow mirrors the basic group data.
type GroupRow struct {
	Name           string
	PhotoURL       string
	CreatedAt      string
	ConversationID string
}

// GetGroup retrieves the photo URL and creation timestamp for a group.
func (adb *AppDatabase) GetGroup(ctx context.Context, groupName string) (GroupRow, error) {
	const query = `
	SELECT group_name, photo_url, created_at, conversation_id
	  FROM groups
	 WHERE group_name = ?
	`
	var g GroupRow
	err := adb.db.QueryRowContext(ctx, query, groupName).
		Scan(&g.Name, &g.PhotoURL, &g.CreatedAt, &g.ConversationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return GroupRow{}, ErrNotFound
		}
		return GroupRow{}, err
	}
	return g, nil
}

// GetGroupMembers returns all usernames belonging to the given group.
func (adb *AppDatabase) GetGroupMembers(ctx context.Context, groupName string) ([]string, error) {
	const query = `
	SELECT username
	  FROM group_members
	 WHERE group_name = ?
	`
	rows, err := adb.db.QueryContext(ctx, query, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

// ListGroups returns the names of all groups.
func (adb *AppDatabase) ListGroups(ctx context.Context) ([]string, error) {
	const query = `
	SELECT group_name
	  FROM groups
	 ORDER BY created_at ASC
	`
	rows, err := adb.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
