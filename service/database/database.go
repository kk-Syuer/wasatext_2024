package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

/*
This file implements the concrete AppDatabase used by your service layer.

Constructor styles supported:
  - New(db *sql.DB)      // template-compatible; you already use this in main.go
  - NewFromDSN(dsn)      // optional convenience (not required by your main.go)
*/

var (
	ErrNotFound     = errors.New("not found")
	ErrUserNotFound = errors.New("user not found")
)

type AppDatabase struct {
	DB *sql.DB
}

/* ------------------------- Constructors ------------------------- */

func New(db *sql.DB) (*AppDatabase, error) {
	if db == nil {
		return nil, errors.New("database is required when building an AppDatabase")
	}
	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		return nil, fmt.Errorf("enable foreign_keys: %w", err)
	}
	ad := &AppDatabase{DB: db}
	if err := ad.initSchema(); err != nil {
		return nil, err
	}
	return ad, nil
}

func NewFromDSN(dsn string) (*AppDatabase, error) {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	ad, err := New(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return ad, nil
}

func (a *AppDatabase) Ping() error { return a.DB.Ping() }

/* --------------------------- Schema ---------------------------- */

func (a *AppDatabase) initSchema() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
  username     TEXT PRIMARY KEY,
  photo_url    TEXT DEFAULT '',
  joined_at    TEXT NOT NULL
);


CREATE TABLE IF NOT EXISTS conversations (
  id           TEXT PRIMARY KEY,
  type         TEXT NOT NULL,  -- "individual" | "group"
  updated_at   TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS conversation_participants (
  conversation_id TEXT NOT NULL,
  username        TEXT NOT NULL,
  PRIMARY KEY (conversation_id, username),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  FOREIGN KEY (username)        REFERENCES users(username)        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS groups (
  name            TEXT PRIMARY KEY,
  photo_url       TEXT DEFAULT '',
  created_at      TEXT NOT NULL,
  conversation_id TEXT NOT NULL UNIQUE,
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS group_members (
  group_name  TEXT NOT NULL,
  username    TEXT NOT NULL,
  PRIMARY KEY (group_name, username),
  FOREIGN KEY (group_name) REFERENCES groups(name) ON DELETE CASCADE,
  FOREIGN KEY (username)   REFERENCES users(username) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS messages (
  id                  TEXT PRIMARY KEY,
  conversation_id     TEXT NOT NULL,
  sender_username     TEXT NOT NULL,
  content_type        TEXT NOT NULL,  -- "text" | "image" | "gif"
  content_url         TEXT DEFAULT '',
  text                TEXT DEFAULT '',
  timestamp           TEXT NOT NULL,
  reply_to            TEXT DEFAULT '',
  forwarded_from      TEXT DEFAULT '',
  forwarded_timestamp TEXT DEFAULT '',
  original_content    TEXT DEFAULT '',
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  FOREIGN KEY (sender_username) REFERENCES users(username)   ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reactions (
  id             TEXT PRIMARY KEY,
  message_id     TEXT NOT NULL,
  emoji          TEXT NOT NULL,
  user_username  TEXT NOT NULL,
  created_at     TEXT NOT NULL,
  UNIQUE (message_id, user_username),
  FOREIGN KEY (message_id)    REFERENCES messages(id) ON DELETE CASCADE,
  FOREIGN KEY (user_username) REFERENCES users(username) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS conversation_reads (
  conversation_id TEXT NOT NULL,
  username        TEXT NOT NULL,
  last_read_at    TEXT NOT NULL,
  PRIMARY KEY (conversation_id, username),
  FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE,
  FOREIGN KEY (username)        REFERENCES users(username)    ON DELETE CASCADE
);
`
	_, err := a.DB.Exec(schema)
	return err
}

/* --------------------------- Types ----------------------------- */

type UserRow struct {
	Username string
	PhotoURL string
	JoinedAt string
}

type GroupRow struct {
	Name           string
	PhotoURL       string
	CreatedAt      string
	ConversationID string
}

type MessageRow struct {
	ID                 string
	ConversationID     string
	SenderUsername     string
	ContentType        string
	ContentURL         string
	Text               string
	Timestamp          string
	ReplyTo            string
	ForwardedFrom      string
	ForwardedTimestamp string
	OriginalContent    string
}

type ReactionRow struct {
	ID           string
	MessageID    string
	Emoji        string
	UserUsername string
	CreatedAt    string
}

type DeliveryStatusRow struct {
	MessageID string
	Recipient string
	Status    string
	UpdatedAt string
}

// BeginTx starts a SQL transaction; the service can call this directly.
func (a *AppDatabase) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return a.DB.BeginTx(ctx, opts)
}


/* ---------------------------- Users ---------------------------- */

func (a *AppDatabase) GetUser(ctx context.Context, username string) (UserRow, error) {
	var u UserRow
	err := a.DB.QueryRowContext(ctx, `
        SELECT username, photo_url, joined_at
          FROM users WHERE username=?`, username).
		Scan(&u.Username, &u.PhotoURL, &u.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	return u, err
}

func (a *AppDatabase) CreateUser(ctx context.Context, username, photoURL, joinedAt string) error {
	_, err := a.DB.ExecContext(ctx, `
        INSERT INTO users (username, photo_url, joined_at)
        VALUES (?, ?, ?)`, username, photoURL, joinedAt)
	return err
}

func (a *AppDatabase) GetAllUsernames(ctx context.Context) ([]string, error) {
	rows, err := a.DB.QueryContext(ctx, `SELECT username FROM users ORDER BY username`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (a *AppDatabase) GetName(ctx context.Context, username string) (string, error) {
	// return the canonical identifier (username) after existence check
	if _, err := a.GetUser(ctx, username); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return "", ErrNotFound
		}
		return "", err
	}
	return username, nil
}

func (a *AppDatabase) SetName(ctx context.Context, oldUsername, newUsername string) error {
	if oldUsername == "" || newUsername == "" {
		return errors.New("invalid username")
	}
	if oldUsername == newUsername {
		return nil // nothing to do
	}

	tx, err := a.DB.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Defer FK checks until COMMIT so we can change parent/children in one txn.
	if _, err = tx.ExecContext(ctx, `PRAGMA defer_foreign_keys = ON;`); err != nil {
		return err
	}

	// 1) Ensure old exists
	var exists int
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM users WHERE username=?`, oldUsername).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		err = ErrNotFound
		return err
	}
	if err != nil {
		return err
	}

	// 2) Ensure new does NOT exist
	err = tx.QueryRowContext(ctx, `SELECT 1 FROM users WHERE username=?`, newUsername).Scan(&exists)
	if err == nil {
		err = errors.New("username already taken")
		return err
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// 3) Update parent first (FK checks are deferred)
	if _, err = tx.ExecContext(ctx, `
        UPDATE users SET username=? WHERE username=?`,
		newUsername, oldUsername); err != nil {
		return err
	}

	// 4) Update children to point to the new username
	if _, err = tx.ExecContext(ctx, `
        UPDATE conversation_participants SET username=? WHERE username=?`,
		newUsername, oldUsername); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
        UPDATE group_members SET username=? WHERE username=?`,
		newUsername, oldUsername); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
        UPDATE messages SET sender_username=? WHERE sender_username=?`,
		newUsername, oldUsername); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
        UPDATE reactions SET user_username=? WHERE user_username=?`,
		newUsername, oldUsername); err != nil {
		return err
	}

	err = tx.Commit()
	return err
}

func (a *AppDatabase) GetPhoto(ctx context.Context, username string) (string, error) {
	var s string
	err := a.DB.QueryRowContext(ctx, `SELECT photo_url FROM users WHERE username=?`, username).Scan(&s)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return s, err
}

func (a *AppDatabase) SetPhoto(ctx context.Context, username, photoURL string) error {
	res, err := a.DB.ExecContext(ctx, `UPDATE users SET photo_url=? WHERE username=?`, photoURL, username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

/* ------------------------ Conversations ----------------------- */

func (a *AppDatabase) CreateConversation(ctx context.Context, id, typ, updatedAt string) error {
	_, err := a.DB.ExecContext(ctx, `
    INSERT INTO conversations (id, type, updated_at) VALUES (?, ?, ?)`, id, typ, updatedAt)
	return err
}

func (a *AppDatabase) AddParticipant(ctx context.Context, conversationID, username string) error {
	_, err := a.DB.ExecContext(ctx, `
        INSERT INTO conversation_participants (conversation_id, username)
        VALUES (?, ?)`, conversationID, username)
	return err
}

func (a *AppDatabase) RemoveParticipant(ctx context.Context, conversationID, username string) error {
	res, err := a.DB.ExecContext(ctx, `
    DELETE FROM conversation_participants WHERE conversation_id=? AND username=?`, conversationID, username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *AppDatabase) GetConversationByID(ctx context.Context, id string) (typ, updatedAt string, err error) {
	err = a.DB.QueryRowContext(ctx, `
    SELECT type, updated_at FROM conversations WHERE id=?`, id).
		Scan(&typ, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", "", ErrNotFound
	}
	return
}

func (a *AppDatabase) GetConversationParticipants(ctx context.Context, id string) ([]string, error) {
	rows, err := a.DB.QueryContext(ctx, `
    SELECT username FROM conversation_participants WHERE conversation_id=? ORDER BY username`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out = append(out, u)
	}
	return out, rows.Err()
}

func (a *AppDatabase) UpdateConversationTimestamp(ctx context.Context, id, updatedAt string) error {
	res, err := a.DB.ExecContext(ctx, `UPDATE conversations SET updated_at=? WHERE id=?`, updatedAt, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *AppDatabase) GetConversationsForUser(ctx context.Context, username string) ([]string, error) {
	rows, err := a.DB.QueryContext(ctx, `
    SELECT c.id
      FROM conversations c
      JOIN conversation_participants p ON p.conversation_id = c.id
     WHERE p.username = ?
     ORDER BY c.updated_at DESC`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (a *AppDatabase) FindConversationByParticipants(ctx context.Context, participants []string) (string, error) {
	if len(participants) == 0 {
		return "", nil
	}
	rows, err := a.DB.QueryContext(ctx, `
    SELECT c.id
      FROM conversations c
      JOIN conversation_participants p ON p.conversation_id = c.id
     WHERE p.username = ?
	 	AND c.type = 'individual'`, participants[0])
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return "", err
		}
		parts, err := a.GetConversationParticipants(ctx, id)
		if err != nil {
			continue
		}
		if equalSet(parts, participants) {
			return id, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return "", nil
}

// --- Tx variants (same SQL, but executed on the provided *sql.Tx) ---

func (a *AppDatabase) TxCreateConversation(ctx context.Context, tx *sql.Tx, id, typ, updatedAt string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO conversations (id, type, updated_at) VALUES (?, ?, ?)`,
		id, typ, updatedAt,
	)
	return err
}

func (a *AppDatabase) TxAddParticipant(ctx context.Context, tx *sql.Tx, conversationID, username string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO conversation_participants (conversation_id, username) VALUES (?, ?)`,
		conversationID, username,
	)
	return err
}


/* ---------------------------- Groups --------------------------- */

func (a *AppDatabase) CreateGroup(ctx context.Context, name, photoURL, createdAt, conversationID string) error {
	_, err := a.DB.ExecContext(ctx, `
    INSERT INTO groups (name, photo_url, created_at, conversation_id)
    VALUES (?, ?, ?, ?)`, name, photoURL, createdAt, conversationID)
	return err
}

func (a *AppDatabase) GetGroup(ctx context.Context, name string) (GroupRow, error) {
	var g GroupRow
	err := a.DB.QueryRowContext(ctx, `
    SELECT name, photo_url, created_at, conversation_id FROM groups WHERE name=?`, name).
		Scan(&g.Name, &g.PhotoURL, &g.CreatedAt, &g.ConversationID)
	if errors.Is(err, sql.ErrNoRows) {
		return GroupRow{}, ErrNotFound
	}
	return g, err
}

func (a *AppDatabase) ListGroups(ctx context.Context) ([]string, error) {
	rows, err := a.DB.QueryContext(ctx, `SELECT name FROM groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (a *AppDatabase) AddGroupMember(ctx context.Context, groupName, username string) error {
	_, err := a.DB.ExecContext(ctx, `
        INSERT INTO group_members (group_name, username)
        VALUES (?, ?)`, groupName, username)
	return err
}

func (a *AppDatabase) RemoveGroupMember(ctx context.Context, groupName, username string) error {
	res, err := a.DB.ExecContext(ctx, `
    DELETE FROM group_members WHERE group_name=? AND username=?`, groupName, username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *AppDatabase) GetGroupMembers(ctx context.Context, groupName string) ([]string, error) {
	rows, err := a.DB.QueryContext(ctx, `
    SELECT username FROM group_members WHERE group_name=? ORDER BY username`, groupName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (a *AppDatabase) UpdateGroupPhoto(ctx context.Context, groupName, photoURL string) error {
	res, err := a.DB.ExecContext(ctx, `
    UPDATE groups SET photo_url=? WHERE name=?`, photoURL, groupName)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (a *AppDatabase) TxCreateGroup(ctx context.Context, tx *sql.Tx, name, photoURL, createdAt, conversationID string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO groups (name, photo_url, created_at, conversation_id) VALUES (?, ?, ?, ?)`,
		name, photoURL, createdAt, conversationID,
	)
	return err
}

func (a *AppDatabase) TxAddGroupMember(ctx context.Context, tx *sql.Tx, groupName, username string) error {
	_, err := tx.ExecContext(ctx,
		`INSERT INTO group_members (group_name, username) VALUES (?, ?)`,
		groupName, username,
	)
	return err
}

// database.go
func (a *AppDatabase) RenameGroup(ctx context.Context, oldName, newName string) error {
    tx, err := a.DB.BeginTx(ctx, nil)
    if err != nil { return err }
    defer tx.Rollback()

    if _, err := tx.ExecContext(ctx, "PRAGMA defer_foreign_keys = ON"); err != nil { return err }

    var cnt int
    if err := tx.QueryRowContext(ctx, "SELECT COUNT(1) FROM groups WHERE name = ?", oldName).Scan(&cnt); err != nil {
        return err
    }
    if cnt == 0 {
        return ErrNotFound
    }

    if _, err := tx.ExecContext(ctx, "UPDATE groups SET name = ? WHERE name = ?", newName, oldName); err != nil {
        return err // hits UNIQUE constraint if taken
    }
    if _, err := tx.ExecContext(ctx, "UPDATE group_members SET group_name = ? WHERE group_name = ?", newName, oldName); err != nil {
        return err
    }
    return tx.Commit()
}


/* --------------------------- Messages -------------------------- */

func (a *AppDatabase) CreateMessage(ctx context.Context, m MessageRow) error {
	_, err := a.DB.ExecContext(ctx, `
    INSERT INTO messages
    (id, conversation_id, sender_username, content_type, content_url, text, timestamp,
     reply_to, forwarded_from, forwarded_timestamp, original_content)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.ID, m.ConversationID, m.SenderUsername, m.ContentType, m.ContentURL, m.Text, m.Timestamp,
		m.ReplyTo, m.ForwardedFrom, m.ForwardedTimestamp, m.OriginalContent)
	return err
}

func (a *AppDatabase) GetMessageByID(ctx context.Context, id string) (MessageRow, error) {
	var m MessageRow
	err := a.DB.QueryRowContext(ctx, `
    SELECT id, conversation_id, sender_username, content_type, content_url, text, timestamp,
           reply_to, forwarded_from, forwarded_timestamp, original_content
      FROM messages WHERE id=?`, id).
		Scan(&m.ID, &m.ConversationID, &m.SenderUsername, &m.ContentType, &m.ContentURL, &m.Text,
			&m.Timestamp, &m.ReplyTo, &m.ForwardedFrom, &m.ForwardedTimestamp, &m.OriginalContent)
	if errors.Is(err, sql.ErrNoRows) {
		return MessageRow{}, ErrNotFound
	}
	return m, err
}

func (a *AppDatabase) GetMessagesForConversation(ctx context.Context, conversationID string) ([]MessageRow, error) {
	rows, err := a.DB.QueryContext(ctx, `
    SELECT id, conversation_id, sender_username, content_type, content_url, text, timestamp,
           reply_to, forwarded_from, forwarded_timestamp, original_content
      FROM messages WHERE conversation_id=?
     ORDER BY timestamp DESC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MessageRow
	for rows.Next() {
		var m MessageRow
		if err := rows.Scan(&m.ID, &m.ConversationID, &m.SenderUsername, &m.ContentType, &m.ContentURL, &m.Text,
			&m.Timestamp, &m.ReplyTo, &m.ForwardedFrom, &m.ForwardedTimestamp, &m.OriginalContent); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (a *AppDatabase) DeleteMessage(ctx context.Context, messageID string) error {
	res, err := a.DB.ExecContext(ctx, `DELETE FROM messages WHERE id=?`, messageID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

/* --------------------------- Reactions ------------------------- */

func (a *AppDatabase) AddReaction(ctx context.Context, r ReactionRow) error {
	// one reaction per (message,user); conflict updates the emoji and timestamp
	_, err := a.DB.ExecContext(ctx, `
    INSERT INTO reactions (id, message_id, emoji, user_username, created_at)
    VALUES (?, ?, ?, ?, ?)
    ON CONFLICT(message_id, user_username) DO UPDATE SET
      emoji=excluded.emoji, created_at=excluded.created_at`, r.ID, r.MessageID, r.Emoji, r.UserUsername, r.CreatedAt)
	return err
}

func (a *AppDatabase) RemoveReaction(ctx context.Context, messageID, username string) error {
	res, err := a.DB.ExecContext(ctx, `
        DELETE FROM reactions WHERE message_id=? AND user_username=?`, messageID, username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}
// Fetch all reactions for a single message.
func (a *AppDatabase) GetReactionsForMessage(ctx context.Context, messageID string) ([]ReactionRow, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT id, message_id, emoji, user_username, created_at
		  FROM reactions
		 WHERE message_id = ?
		 ORDER BY created_at ASC`, messageID)
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
	return out, rows.Err()
}

// Batch: reactions for all messages in a conversation (1 query).
func (a *AppDatabase) GetReactionsForConversation(ctx context.Context, conversationID string) (map[string][]ReactionRow, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT r.id, r.message_id, r.emoji, r.user_username, r.created_at
		  FROM reactions r
		  JOIN messages m ON m.id = r.message_id
		 WHERE m.conversation_id = ?
		 ORDER BY r.created_at ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string][]ReactionRow)
	for rows.Next() {
		var r ReactionRow
		if err := rows.Scan(&r.ID, &r.MessageID, &r.Emoji, &r.UserUsername, &r.CreatedAt); err != nil {
			return nil, err
		}
		out[r.MessageID] = append(out[r.MessageID], r)
	}
	return out, rows.Err()
}
/* ---------------------- Delivery statuses ---------------------- */

func (a *AppDatabase) GetDeliveryStatusForConversation(ctx context.Context, conversationID string) ([]DeliveryStatusRow, error) {
	// Participants in the conversation
	parts, err := a.GetConversationParticipants(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, ErrNotFound
	}

	// All messages for this conversation (DB already returns DESC; order not important here)
	msgs, err := a.GetMessagesForConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return []DeliveryStatusRow{}, nil
	}

	// Load "last read" watermark per participant (if the table exists).
	// If the table hasn't been created yet, we ignore the error and just return "sent".
	readAt := map[string]time.Time{}
	if rows, qerr := a.DB.QueryContext(ctx, `
		SELECT username, last_read_at
		FROM conversation_reads
		WHERE conversation_id = ?
	`, conversationID); qerr == nil {
		defer rows.Close()
		for rows.Next() {
			var u, t string
			if err := rows.Scan(&u, &t); err == nil {
				if ts, perr := time.Parse(time.RFC3339, t); perr == nil {
					readAt[u] = ts
				}
			}
		}
		_ = rows.Err()
		// If query failed with "no such table", we just fall through with an empty map.
	}

	out := make([]DeliveryStatusRow, 0, len(msgs)*len(parts))

	for _, m := range msgs {
		// Parse message timestamp once
		var msgTS time.Time
		if ts, perr := time.Parse(time.RFC3339, m.Timestamp); perr == nil {
			msgTS = ts
		}

		for _, p := range parts {
			// Skip sender; we only compute status for recipients
			if p == m.SenderUsername {
				continue
			}

			status := "sent" // baseline -> single ✓ on the UI

			// If we have a read watermark for this recipient and the message
			// timestamp is <= last_read_at, consider it "read" -> double ✓.
			if !msgTS.IsZero() {
				if rt, ok := readAt[p]; ok && (msgTS.Before(rt) || msgTS.Equal(rt)) {
					status = "read"
				}
			}

			out = append(out, DeliveryStatusRow{
				MessageID: m.ID,
				Recipient: p,
				Status:    status,
				UpdatedAt: m.Timestamp, // keep message ts; or use rt.Format(time.RFC3339) if you prefer
			})
		}
	}

	return out, nil
}

// Upsert a user's last-read watermark for a conversation.
func (a *AppDatabase) UpsertConversationRead(ctx context.Context, conversationID, username string, at time.Time) error {
	_, err := a.DB.ExecContext(ctx, `
		INSERT INTO conversation_reads (conversation_id, username, last_read_at)
		VALUES (?, ?, ?)
		ON CONFLICT(conversation_id, username)
		DO UPDATE SET last_read_at = excluded.last_read_at
	`, conversationID, username, at.UTC().Format(time.RFC3339))
	return err
}

// Load last-read watermark per recipient for a conversation.
func (a *AppDatabase) GetConversationReadMap(ctx context.Context, conversationID string) (map[string]time.Time, error) {
	rows, err := a.DB.QueryContext(ctx, `
		SELECT username, last_read_at
		FROM conversation_reads
		WHERE conversation_id = ?
	`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]time.Time)
	for rows.Next() {
		var u, t string
		if err := rows.Scan(&u, &t); err != nil {
			return nil, err
		}
		if ts, err := time.Parse(time.RFC3339, t); err == nil {
			out[u] = ts
		}
	}
	return out, nil
}

/* ----------------------------- Utils --------------------------- */

func equalSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]int, len(a))
	for _, x := range a {
		m[x]++
	}
	for _, y := range b {
		if m[y] == 0 {
			return false
		}
		m[y]--
	}
	for _, v := range m {
		if v != 0 {
			return false
		}
	}
	return true
}
