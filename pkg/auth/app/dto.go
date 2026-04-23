package app

import (
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

type LoginCommand struct {
	SessionID    string
	UserID       string
	RefreshToken string
	IssuedAt     time.Time
	ExpiresAt    time.Time
}

type LoginResult struct {
	RefreshSession *domain.RefreshSession
}

type RefreshTokenCommand struct {
	SessionID        string
	CurrentToken     string
	NewRefreshToken  string
	NewRefreshExpiry time.Time
	Now              time.Time
}

type RefreshTokenResult struct {
	RefreshSession *domain.RefreshSession
}

type LogoutCommand struct {
	SessionID string
	RevokedAt time.Time
}
