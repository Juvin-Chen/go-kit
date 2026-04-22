package domain

import (
	"context"
	"time"
)

type RefreshSessionRepository interface {
	// CreateRefreshSession 持久化一条新的 refresh 会话记录
	CreateRefreshSession(ctx context.Context, session *RefreshSession) error

	// GetRefreshSessionBySessionID 按会话 ID 查询会话；找不到时返回 (nil, ErrRefreshSessionNotFound)
	GetRefreshSessionBySessionID(ctx context.Context, sessionID string) (*RefreshSession, error)

	// UpdateRefreshSessionOnRotate 在令牌轮换后更新会话状态；版本不匹配时返回 ErrRefreshSessionConflict
	UpdateRefreshSessionOnRotate(ctx context.Context, session *RefreshSession, expectedVersion uint64) error

	// RevokeRefreshSession 撤销指定会话；版本不匹配时返回 ErrRefreshSessionConflict
	RevokeRefreshSession(ctx context.Context, sessionID string, revokedAt time.Time, expectedVersion uint64) error

	// DeleteExpiredRefreshSessions 批量清理已过期会话，返回实际清理数量
	DeleteExpiredRefreshSessions(ctx context.Context, expiredBefore time.Time, limit uint64) (uint64, error)
}
