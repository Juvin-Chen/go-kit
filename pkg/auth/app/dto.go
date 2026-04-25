package app

import (
	"time"
)

type LoginCommand struct {
	SessionID    string
	UserID       string
	RefreshToken string
}

type LoginResult struct {
	SessionID             string
	UserID                string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type RefreshTokenCommand struct {
	SessionID        string
	CurrentToken     string
	NewRefreshToken  string
	NewRefreshExpiry time.Time
	Now              time.Time
}

type RefreshTokenResult struct {
	SessionID             string
	UserID                string
	AccessToken           string
	AccessTokenExpiresAt  time.Time
	RefreshTokenExpiresAt time.Time
}

type LogoutCommand struct {
	SessionID string
	RevokedAt time.Time
}
