// service/session_service.go
package service

import (
	"context"
	"errors"
	"github.com/kk-Syuer/wasatext_2024/service/database"
	"github.com/kk-Syuer/wasatext_2024/service/globaltime"
	"time"
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
	if _, err := s.db.GetUser(ctx, username); err != nil {
		if errors.Is(err, database.ErrUserNotFound) {
			joined := globaltime.Now().Format(time.RFC3339)
			if err2 := s.db.CreateUser(ctx, username, "", joined); err2 != nil {
				return "", err2
			}
		} else {
			return "", err
		}
	}
	// identifier == username
	return username, nil
}

func (s *sessionServiceImpl) Validate(ctx context.Context, token string) (string, error) {
	// In this model, "token" == username
	if _, err := s.db.GetUser(ctx, token); err != nil {
		return "", err
	}
	return token, nil
}
