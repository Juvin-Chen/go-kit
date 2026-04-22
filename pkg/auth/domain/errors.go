package domain

import "errors"

var (
	ErrInvalidRefreshToken      = errors.New("invalid refresh token")
	ErrInvalidRefreshSessionID  = errors.New("invalid refresh session id")
	ErrInvalidRefreshTokenID    = ErrInvalidRefreshSessionID
	ErrInvalidUserID            = errors.New("invalid user id")
	ErrRefreshTokenHash         = errors.New("invalid refresh token hash")
	ErrInvalidRefreshTokenTTL   = errors.New("invalid refresh token ttl")
	ErrRefreshSessionRevoked    = errors.New("refresh session revoked")
	ErrRefreshSessionExpired    = errors.New("refresh session expired")
)
