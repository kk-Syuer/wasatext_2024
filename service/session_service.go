// service/session_service.go
package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
)

// SessionService handles “login or register” semantics and returns a bearer identifier.
type SessionService interface {
	// Login returns the user’s identifier (creating a new user record if needed).
	Login(ctx context.Context, username string) (string, error)
	// Validate returns the username for a given session token, or error.
	Validate(ctx context.Context, token string) (string, error)
}

// sessionServiceImpl is our default SessionService.
type sessionServiceImpl struct {
	db *database.AppDatabase
}

func NewSessionService(db *database.AppDatabase) SessionService {
	return &sessionServiceImpl{db: db}
}

func (s *sessionServiceImpl) Login(ctx context.Context, username string) (string, error) {
	// 1) Ensure the user row exists (generate an ID for new users):
	if _, err := s.db.GetUser(ctx, username); err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			// user doesn’t exist → create them
			userID := uuid.New().String()
			joined := globaltime.Now().Format(time.RFC3339)
			if err2 := s.db.CreateUser(ctx, userID, username, "", "", joined); err2 != nil {
				return "", err2
			}
		} else {
			// some other DB error
			return "", err
		}
	}

	// 2) new session token
	token := uuid.New().String()
	now := time.Now().Format(time.RFC3339)
	if err := s.db.CreateSession(ctx, token, username, now); err != nil {
		return "", err
	}
	return token, nil
}

func (s *sessionServiceImpl) Validate(ctx context.Context, token string) (string, error) {
	return s.db.GetSessionUsername(ctx, token)
}
