package database

import (
	"context"
	"database/sql"
	"errors"
)

// ErrNotFound is returned when a lookup yields no rows.
var ErrNotFound = errors.New("record not found")

// AppDatabase wraps the SQL connection and provides app-specific data methods.
type AppDatabase struct {
	db *sql.DB
}

// New initializes the schema (creates tables if not exist) and returns *AppDatabase.
func New(db *sql.DB) (*AppDatabase, error) {
	adb := &AppDatabase{db: db}

	schemas := []string{
		`CREATE TABLE IF NOT EXISTS users (
		    id TEXT PRIMARY KEY,
		    username TEXT UNIQUE NOT NULL,
		    name TEXT NOT NULL,
		    photo_url TEXT NOT NULL DEFAULT '',
		    joined_at TEXT NOT NULL
		);`,
		// Conversations for groups are managed in the same table as 1:1 convos:
		`CREATE TABLE IF NOT EXISTS conversations (
			id          TEXT PRIMARY KEY,
			type        TEXT NOT NULL,
			updated_at  TEXT NOT NULL
		  );`,

		`CREATE TABLE IF NOT EXISTS conversation_participants (
			conversation_id TEXT NOT NULL,
			username        TEXT NOT NULL,
			PRIMARY KEY(conversation_id, username),
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
			FOREIGN KEY(username)       REFERENCES users(username)       ON DELETE CASCADE
		  );`,
		`CREATE TABLE IF NOT EXISTS delivery_status (
			message_id TEXT NOT NULL,
			recipient  TEXT NOT NULL,
			status     TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			PRIMARY KEY(message_id, recipient),
			FOREIGN KEY(message_id) REFERENCES messages(id) ON DELETE CASCADE,
			FOREIGN KEY(recipient)   REFERENCES users(username) ON DELETE CASCADE
		);`,
		// Messages table
		`CREATE TABLE IF NOT EXISTS messages (
			id                  TEXT PRIMARY KEY,
			conversation_id     TEXT NOT NULL,
			sender_username     TEXT NOT NULL,
			content_type        TEXT NOT NULL,
			content_url         TEXT,
			text                TEXT,
			timestamp           TEXT NOT NULL,
			reply_to            TEXT,
			forwarded_from      TEXT,
			forwarded_timestamp TEXT,
			original_content    TEXT,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		);`,

		/// Now groups point to those conversation rows:
		`CREATE TABLE IF NOT EXISTS groups (
			group_name     TEXT PRIMARY KEY,
			conversation_id TEXT NOT NULL UNIQUE,
			photo_url      TEXT NOT NULL DEFAULT '',
			created_at     TEXT NOT NULL,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
		  );`,

		`CREATE TABLE IF NOT EXISTS group_members (
			group_name TEXT NOT NULL,
			username   TEXT NOT NULL,
			PRIMARY KEY(group_name, username),
			FOREIGN KEY(group_name)      REFERENCES groups(group_name)      ON DELETE CASCADE,
			FOREIGN KEY(username)        REFERENCES users(username)        ON DELETE CASCADE,
			FOREIGN KEY(conversation_id) REFERENCES conversations(id)       ON DELETE CASCADE
			  USING (conversation_id)  -- SQLite doesn’t support this, see note below
		  );`,

		// Reactions
		`CREATE TABLE IF NOT EXISTS reactions (
            id TEXT PRIMARY KEY,
            message_id TEXT NOT NULL,
            emoji TEXT NOT NULL,
            user_username TEXT NOT NULL,
            created_at TEXT NOT NULL,
            FOREIGN KEY(message_id)    REFERENCES messages(id) ON DELETE CASCADE,
            FOREIGN KEY(user_username) REFERENCES users(username) ON DELETE CASCADE
        );`,
	}

	for _, ddl := range schemas {
		if _, err := adb.db.ExecContext(context.Background(), ddl); err != nil {
			return nil, err
		}
	}

	return adb, nil
}

// GetUserID returns the ID for the given username, or ErrNotFound.
func (adb *AppDatabase) GetUserID(ctx context.Context, username string) (string, error) {
	var id string
	err := adb.db.QueryRowContext(ctx,
		"SELECT id FROM users WHERE username = ?",
		username,
	).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", err
	}
	return id, nil
}

// CreateUser inserts a new user record.
func (adb *AppDatabase) CreateUser(
	ctx context.Context,
	id, username, name, photoURL, joinedAt string,
) error {
	_, err := adb.db.ExecContext(ctx,
		`INSERT INTO users (id, username, name, photo_url, joined_at)
         VALUES (?, ?, ?, ?, ?)`,
		id, username, name, photoURL, joinedAt,
	)
	return err
}
