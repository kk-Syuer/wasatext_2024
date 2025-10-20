package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
	"sort"
	"strings"
	"time"
)

// Group represents a named group chat.
type Group struct {
	Name           string    `json:"name"`
	PhotoURL       string    `json:"photoUrl"`
	CreatedAt      time.Time `json:"createdAt"`
	Members        []string  `json:"members"`
	ConversationID string    `json:"conversationId"`
}

// GroupService defines operations on groups.
type GroupService interface {
	// CreateGroup creates a new group with the given name, photo URL, and members.
	CreateGroup(ctx context.Context, name, photoURL string, members []string) (Group, error)
	// GetGroup retrieves the group metadata and member list.
	GetGroup(ctx context.Context, name string) (Group, error)
	// ListGroups returns the names of all groups.
	ListGroups(ctx context.Context) ([]string, error)
	// AddMember adds a user to the group.
	AddMember(ctx context.Context, groupName, username string) error
	// RemoveMember removes a user from the group.
	RemoveMember(ctx context.Context, groupName, username string) error
	// UpdatePhoto updates the group’s photo URL.
	UpdatePhoto(ctx context.Context, groupName, photoURL string) error
	// LeaveGroup 让指定用户退出群组
	LeaveGroup(ctx context.Context, groupName, username string) error
	ListGroupsDetailed(ctx context.Context) ([]Group, error)
	UpdateName(ctx context.Context, oldName, newName string) error
}

type groupServiceImpl struct {
	db *database.AppDatabase
}

// NewGroupService constructs a GroupService using the given AppDatabase.
func NewGroupService(db *database.AppDatabase) GroupService {
	return &groupServiceImpl{db: db}
}

// parse common SQLite/ISO formats into time.Time, fallback to now
func createdAtFromDB(s string) time.Time {
	ts := strings.TrimSpace(s)
	if ts == "" {
		return globaltime.Now()
	}
	// Try the most precise RFCs first
	if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		return t
	}
	// Common SQLite datetime format without timezone
	if t, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
		return t
	}
	// Last resort
	return globaltime.Now()
}

func (s *groupServiceImpl) CreateGroup(
	ctx context.Context,
	name string,
	photoURL string,
	members []string,
) (Group, error) {
	if strings.TrimSpace(name) == "" {
		return Group{}, ErrBadRequest
	}

	// De-dupe and sanitize member list
	uniq := make(map[string]struct{}, len(members))
	for _, m := range members {
		m = strings.TrimSpace(m)
		if m != "" {
			uniq[m] = struct{}{}
		}
	}
	if len(uniq) < 2 {
		return Group{}, ErrBadRequest // need at least 2 people for a group
	}

	convID := uuid.New().String()
	nowTS := globaltime.Now().UTC()
	nowRFC3339 := nowTS.Format(time.RFC3339)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Group{}, err
	}
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
				err = fmt.Errorf("rollback failed: %w (original: %v)", rbErr, err)
			}
		} else {
			err = tx.Commit()
		}
	}()

	if err := s.db.TxCreateConversation(ctx, tx, convID, "group", nowRFC3339); err != nil {
		return Group{}, err
	}

	for m := range uniq {
		if err := s.db.TxAddParticipant(ctx, tx, convID, m); err != nil {
			return Group{}, err
		}
	}

	if err := s.db.TxCreateGroup(ctx, tx, name, photoURL, nowRFC3339, convID); err != nil {
		return Group{}, err
	}

	for m := range uniq {
		if err := s.db.TxAddGroupMember(ctx, tx, name, m); err != nil {
			return Group{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return Group{}, err
	}

	outMembers := make([]string, 0, len(uniq))
	for m := range uniq {
		outMembers = append(outMembers, m)
	}
	sort.Strings(outMembers)

	return Group{
		Name:           name,
		PhotoURL:       photoURL,
		CreatedAt:      nowTS,
		Members:        outMembers,
		ConversationID: convID,
	}, nil
}

func (s *groupServiceImpl) GetGroup(ctx context.Context, name string) (Group, error) {
	row, err := s.db.GetGroup(ctx, name)
	if err != nil {
		return Group{}, err
	}
	members, _ := s.db.GetGroupMembers(ctx, name)

	return Group{
		Name:           row.Name,
		PhotoURL:       row.PhotoURL,
		CreatedAt:      createdAtFromDB(row.CreatedAt), // <-- convert string -> time.Time
		Members:        members,
		ConversationID: row.ConversationID,
	}, nil
}
func (s *groupServiceImpl) ListGroupsDetailed(ctx context.Context) ([]Group, error) {
	names, err := s.ListGroups(ctx) // this calls s.db.ListGroups(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]Group, 0, len(names))
	for _, n := range names {
		if n == "" {
			continue
		}
		g, err := s.GetGroup(ctx, n)
		if err != nil {
			// skip broken rows instead of failing the whole list
			continue
		}
		out = append(out, g)
	}
	return out, nil
}

func (s *groupServiceImpl) ListGroups(ctx context.Context) ([]string, error) {
	return s.db.ListGroups(ctx)
}

func (s *groupServiceImpl) UpdatePhoto(ctx context.Context, groupName, photoURL string) error {
	return s.db.UpdateGroupPhoto(ctx, groupName, photoURL)
}
func (s *groupServiceImpl) AddMember(ctx context.Context, name, username string) error {
	// 1) Add to group_members
	if err := s.db.AddGroupMember(ctx, name, username); err != nil {
		return err
	}
	// 2) Lookup conversation_id for the group
	row, err := s.db.GetGroup(ctx, name)
	if err != nil {
		return err
	}
	// 3) Add to conversation_participants
	return s.db.AddParticipant(ctx, row.ConversationID, username)
}

func (s *groupServiceImpl) RemoveMember(ctx context.Context, name, username string) error {
	// 1) Remove from group_members
	if err := s.db.RemoveGroupMember(ctx, name, username); err != nil {
		return err
	}
	// 2) Lookup linked conversation
	row, err := s.db.GetGroup(ctx, name)
	if err != nil {
		return err
	}
	// 3) Remove from conversation_participants
	return s.db.RemoveParticipant(ctx, row.ConversationID, username)
}

func (s *groupServiceImpl) LeaveGroup(ctx context.Context, groupName, username string) error {
	grp, err := s.db.GetGroup(ctx, groupName)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	members, err := s.db.GetGroupMembers(ctx, groupName)
	if err != nil {
		return err
	}
	found := false
	for _, m := range members {
		if m == username {
			found = true
			break
		}
	}
	if !found {
		return ErrForbidden
	}
	if err := s.db.RemoveGroupMember(ctx, groupName, username); err != nil {
		return err
	}
	if err := s.db.RemoveParticipant(ctx, grp.ConversationID, username); err != nil {
		return err
	}
	return nil
}
func (s *groupServiceImpl) UpdateName(ctx context.Context, oldName, newName string) error {
	newName = strings.TrimSpace(newName)
	if newName == "" {
		return ErrBadRequest
	}
	return s.db.RenameGroup(ctx, oldName, newName)
}
