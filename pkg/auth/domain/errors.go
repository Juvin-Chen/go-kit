package domain

import "errors"

var (
	ErrInvalidRefreshSession        = errors.New("invalid refresh session")
	ErrInvalidRefreshSessionID      = errors.New("invalid refresh session id")
	ErrInvalidRefreshSessionVersion = errors.New("invalid refresh session version")
	ErrInvalidUserID                = errors.New("invalid user id")
	ErrInvalidRefreshToken          = errors.New("invalid refresh token")
	ErrInvalidRefreshTokenHash      = errors.New("invalid refresh token hash")
	ErrInvalidRefreshTokenTTL       = errors.New("invalid refresh token ttl")
	ErrRefreshSessionRevoked        = errors.New("refresh session revoked")
	ErrRefreshSessionExpired        = errors.New("refresh session expired")
	ErrRefreshSessionNotFound       = errors.New("refresh session not found")
	ErrRefreshSessionConflict       = errors.New("refresh session version conflict")
	ErrInvalidCleanupLimit          = errors.New("invalid cleanup limit")
)
