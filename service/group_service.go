package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
)

// Group represents a named group chat.
type Group struct {
	Name      string    `json:"name"`
	PhotoURL  string    `json:"photoUrl"`
	CreatedAt time.Time `json:"createdAt"`
	Members   []string  `json:"members"`
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
}

type groupServiceImpl struct {
	db *database.AppDatabase
}

// NewGroupService constructs a GroupService using the given AppDatabase.
func NewGroupService(db *database.AppDatabase) GroupService {
	return &groupServiceImpl{db: db}
}

func (s *groupServiceImpl) CreateGroup(ctx context.Context, name, photoURL string, members []string) (Group, error) {
	// 1) Create a conversation row for this group
	convID := uuid.New().String()
	nowStr := globaltime.Now().Format(time.RFC3339)
	// 1) Create the conversation row for this group
	if err := s.db.CreateConversation(ctx, convID, "group", nowStr); err != nil {
		return Group{}, err
	}

	// 2) Create group row pointing at that conversation
	if err := s.db.CreateGroup(ctx, name, photoURL, nowStr, convID); err != nil {
		return Group{}, err
	}

	// 3) Add participants into both group_members and conversation_participants
	for _, u := range members {
		if err := s.db.AddGroupMember(ctx, name, u); err != nil {
			return Group{}, err
		}
		if err := s.db.AddParticipant(ctx, convID, u); err != nil {
			return Group{}, err
		}
	}

	// 4) Return fully populated group
	return s.GetGroup(ctx, name)
}

func (s *groupServiceImpl) GetGroup(ctx context.Context, name string) (Group, error) {
	// Fetch group metadata
	row, err := s.db.GetGroup(ctx, name)
	if err != nil {
		return Group{}, err
	}

	// Parse timestamp
	createdAt, _ := time.Parse(time.RFC3339, row.CreatedAt)

	// Fetch members
	members, err := s.db.GetGroupMembers(ctx, name)
	if err != nil {
		return Group{}, err
	}

	return Group{
		Name:      row.Name,
		PhotoURL:  row.PhotoURL,
		CreatedAt: createdAt,
		Members:   members,
	}, nil
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
