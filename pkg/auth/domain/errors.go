package domain

import "errors"

var (
	// 聚合与值对象基础校验错误
	ErrInvalidRefreshSession        = errors.New("invalid refresh session")
	ErrInvalidRefreshSessionID      = errors.New("invalid refresh session id")
	ErrInvalidRefreshSessionVersion = errors.New("invalid refresh session version")
	ErrInvalidUserID                = errors.New("invalid user id")
	ErrInvalidRefreshToken          = errors.New("invalid refresh token")
	ErrInvalidRefreshTokenHash      = errors.New("invalid refresh token hash")
	ErrInvalidRefreshTokenTTL       = errors.New("invalid refresh token ttl")

	// 会话生命周期状态错误
	ErrRefreshSessionRevoked = errors.New("refresh session revoked")
	ErrRefreshSessionExpired = errors.New("refresh session expired")

	// 仓储契约与并发控制错误
	ErrRefreshSessionNotFound = errors.New("refresh session not found")
	ErrRefreshSessionConflict = errors.New("refresh session version conflict")
	ErrInvalidCleanupLimit    = errors.New("invalid cleanup limit")
)
