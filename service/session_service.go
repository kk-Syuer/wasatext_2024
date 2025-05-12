// service/session_service.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
)

// SessionService handles “login or register” semantics and returns a bearer identifier.
type SessionService interface {
	// Login returns the user’s identifier (creating a new user record if needed).
	Login(ctx context.Context, username string) (string, error)
}

// sessionServiceImpl is our default SessionService.
type sessionServiceImpl struct {
	db *database.AppDatabase
}

// NewSessionService constructs a SessionService backed by db.
func NewSessionService(db *database.AppDatabase) SessionService {
	return &sessionServiceImpl{db: db}
}

func (s *sessionServiceImpl) Login(ctx context.Context, username string) (string, error) {
	// 1) Try to fetch an existing user ID
	id, err := s.db.GetUserID(ctx, username)
	if err != nil {
		if err == database.ErrNotFound {
			// 2) Not found: create a new user
			id = uuid.New().String()
			now := globaltime.Now().Format(time.RFC3339)
			// Use the username itself as the default display name
			if err := s.db.CreateUser(ctx, id, username, username, "", now); err != nil {
				return "", err
			}
		} else {
			// Some other DB error
			return "", err
		}
	}
	// 3) Return the found or newly created ID
	return id, nil
}
