// service/user_service.go
package service

import (
	"context"
	"errors"

	"github.com/kk-Syuer/wasatext_2024/service/database"
)

// User represents a user account.
type User struct {
	Username string `json:"username"` // unique identifier
	Name     string `json:"name"`     // display name
	PhotoURL string `json:"photoUrl"` // avatar URL
}

// UserService defines operations on users.
type UserService interface {
	ListUsers(ctx context.Context) ([]User, error)
	GetUser(ctx context.Context, username string) (User, error)
	UpdateName(ctx context.Context, username, newName string) error
	UpdatePhoto(ctx context.Context, username, photoURL string) error
}

type userServiceImpl struct {
	db *database.AppDatabase
}

// NewUserService creates a UserService that persists to db.
func NewUserService(db *database.AppDatabase) UserService {
	return &userServiceImpl{db: db}
}

func (s *userServiceImpl) ListUsers(ctx context.Context) ([]User, error) {
	// Step 1: fetch all usernames
	usernames, err := s.db.GetAllUsernames(ctx)
	if err != nil {
		return nil, err
	}

	// Step 2: for each username, fetch the full profile
	users := make([]User, 0, len(usernames))
	for _, uname := range usernames {
		u, err := s.GetUser(ctx, uname)
		if err != nil {
			// if user was deleted between calls, skip
			if errors.Is(err, database.ErrNotFound) {
				continue
			}
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *userServiceImpl) GetUser(ctx context.Context, username string) (User, error) {
	name, err := s.db.GetName(ctx, username)
	if err != nil {
		return User{}, err
	}
	photo, err := s.db.GetPhoto(ctx, username)
	if err != nil {
		return User{}, err
	}
	return User{
		Username: username,
		Name:     name,
		PhotoURL: photo,
	}, nil
}

func (s *userServiceImpl) UpdateName(ctx context.Context, username, newName string) error {
	return s.db.SetName(ctx, username, newName)
}

func (s *userServiceImpl) UpdatePhoto(ctx context.Context, username, photoURL string) error {
	return s.db.SetPhoto(ctx, username, photoURL)
}
