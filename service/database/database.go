package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
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
  user_id      TEXT UNIQUE NOT NULL,
  name         TEXT DEFAULT '',
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
`
	_, err := a.DB.Exec(schema)
	return err
}

/* --------------------------- Types ----------------------------- */

type UserRow struct {
	Username string
	UserID   string
	Name     string
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

/* ---------------------------- Users ---------------------------- */

func (a *AppDatabase) GetUser(ctx context.Context, username string) (UserRow, error) {
	var u UserRow
	err := a.DB.QueryRowContext(ctx, `
    SELECT username, user_id, name, photo_url, joined_at
      FROM users WHERE username=?`, username).
		Scan(&u.Username, &u.UserID, &u.Name, &u.PhotoURL, &u.JoinedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return UserRow{}, ErrUserNotFound
	}
	return u, err
}

func (a *AppDatabase) CreateUser(ctx context.Context, userID, username, name, photoURL, joinedAt string) error {
	_, err := a.DB.ExecContext(ctx, `
    INSERT INTO users (username, user_id, name, photo_url, joined_at)
    VALUES (?, ?, ?, ?, ?)`, username, userID, name, photoURL, joinedAt)
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
	var s string
	err := a.DB.QueryRowContext(ctx, `SELECT name FROM users WHERE username=?`, username).Scan(&s)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	return s, err
}

func (a *AppDatabase) SetName(ctx context.Context, username, newName string) error {
	res, err := a.DB.ExecContext(ctx, `UPDATE users SET name=? WHERE username=?`, newName, username)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
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
    INSERT OR IGNORE INTO conversation_participants (conversation_id, username)
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
     WHERE p.username = ?`, participants[0])
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
    INSERT OR IGNORE INTO group_members (group_name, username) VALUES (?, ?)`, groupName, username)
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

/* ---------------------- Delivery statuses ---------------------- */

func (a *AppDatabase) GetDeliveryStatusForConversation(ctx context.Context, conversationID string) ([]DeliveryStatusRow, error) {
	parts, err := a.GetConversationParticipants(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if len(parts) == 0 {
		return nil, ErrNotFound
	}
	msgs, err := a.GetMessagesForConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}
	if len(msgs) == 0 {
		return []DeliveryStatusRow{}, nil
	}
	out := make([]DeliveryStatusRow, 0, len(msgs)*len(parts))
	for _, m := range msgs {
		for _, p := range parts {
			if p == m.SenderUsername {
				continue
			}
			out = append(out, DeliveryStatusRow{
				MessageID: m.ID,
				Recipient: p,
				Status:    "sent",      // upgrade to "received"/"read" when you track reads
				UpdatedAt: m.Timestamp, // baseline
			})
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
