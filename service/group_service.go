// service/group_service.go
package service

import "context"

// Group represents a chat group.
type Group struct {
	GroupName       string
	MemberUsernames []string
	PhotoURL        string
	CreatedAt       string
}

// GroupService defines create/join/leave/update operations.
type GroupService interface {
	CreateGroup(ctx context.Context, name string, members []string, initialMsg string) (string, error)
	LeaveGroup(ctx context.Context, groupName, username string) error
	AddGroupMember(ctx context.Context, groupName, username string) error
	RemoveGroupMember(ctx context.Context, groupName, username string) error
	UpdateGroupPhoto(ctx context.Context, groupName, photoURL string) error
}

type groupServiceImpl struct {
	// db *database.AppDatabase
}

func NewGroupService( /*db *database.AppDatabase*/ ) GroupService {
	return &groupServiceImpl{ /*db: db*/ }
}

func (s *groupServiceImpl) CreateGroup(ctx context.Context, name string, members []string, initialMsg string) (string, error) {
	// TODO: db.CreateGroup, db.AddGroupMember for each member, db.InsertMessage…
	return "", nil
}

func (s *groupServiceImpl) LeaveGroup(ctx context.Context, groupName, username string) error {
	// TODO: db.RemoveGroupMember…
	return nil
}

func (s *groupServiceImpl) AddGroupMember(ctx context.Context, groupName, username string) error {
	// TODO: db.AddGroupMember…
	return nil
}

func (s *groupServiceImpl) RemoveGroupMember(ctx context.Context, groupName, username string) error {
	// TODO: db.RemoveGroupMember…
	return nil
}

func (s *groupServiceImpl) UpdateGroupPhoto(ctx context.Context, groupName, photoURL string) error {
	// TODO: db.SetGroupPhoto…
	return nil
}
