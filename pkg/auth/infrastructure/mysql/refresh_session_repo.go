package mysql

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
	"gorm.io/gorm"
)

type MySQLRefreshSessionRepository struct {
	db *gorm.DB
}

var ErrMySQLDBNil = errors.New("mysql repository db is nil")

func NewMySQLRefreshSessionRepository(db *gorm.DB) *MySQLRefreshSessionRepository {
	return &MySQLRefreshSessionRepository{db: db}
}

// UpdateRefreshSessionOnRotate 使用版本号执行乐观锁更新
// 条件不满足（包括并发更新导致版本变化）时返回 ErrRefreshSessionConflict
func (r *MySQLRefreshSessionRepository) UpdateRefreshSessionOnRotate(
	ctx context.Context,
	session *domain.RefreshSession,
	expectedVersion uint64,
) error {
	if r == nil || r.db == nil {
		return ErrMySQLDBNil
	}
	if session == nil {
		return domain.ErrInvalidRefreshSession
	}
	if expectedVersion == 0 {
		return domain.ErrInvalidRefreshSessionVersion
	}
	if strings.TrimSpace(session.SessionID) == "" {
		return domain.ErrInvalidRefreshSessionID
	}

	// session.Version 是上层已经算好的【新版本】
	// expectedVersion 是数据库里还没更新的【旧版本】
	updateResult := r.db.WithContext(ctx).
		Model(&RefreshSessionPO{}).
		Where("session_id = ? AND version = ?", session.SessionID, expectedVersion).
		Updates(map[string]any{
			"refresh_token_hash": session.RefreshTokenHash,
			"issued_at":          session.IssuedAt,
			"expires_at":         session.ExpiresAt,
			"version":            session.Version,
			"updated_at":         time.Now(),
		})
	if updateResult.Error != nil {
		return updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return domain.ErrRefreshSessionConflict
	}
	return nil
}

func (r *MySQLRefreshSessionRepository) CreateRefreshSession(ctx context.Context, session *domain.RefreshSession) error {
	if r == nil || r.db == nil {
		return ErrMySQLDBNil
	}
	if session == nil {
		return domain.ErrInvalidRefreshSession
	}

	createResult := r.db.WithContext(ctx).Create(ToRefreshSessionPO(session))
	return createResult.Error
}

func (r *MySQLRefreshSessionRepository) GetRefreshSessionBySessionID(ctx context.Context, sessionID string) (*domain.RefreshSession, error) {
	if r == nil || r.db == nil {
		return nil, ErrMySQLDBNil
	}
	if strings.TrimSpace(sessionID) == "" {
		return nil, domain.ErrInvalidRefreshSessionID
	}

	var po RefreshSessionPO
	queryResult := r.db.WithContext(ctx).
		Where("session_id = ?", strings.TrimSpace(sessionID)).
		First(&po)
	if queryResult.Error != nil {
		if errors.Is(queryResult.Error, gorm.ErrRecordNotFound) {
			return nil, domain.ErrRefreshSessionNotFound
		}
		return nil, queryResult.Error
	}
	return ToDomainRefreshSession(&po)
}

func (r *MySQLRefreshSessionRepository) RevokeRefreshSession(
	ctx context.Context,
	sessionID string,
	revokedAt time.Time,
	expectedVersion uint64,
) error {
	if r == nil || r.db == nil {
		return ErrMySQLDBNil
	}
	if strings.TrimSpace(sessionID) == "" {
		return domain.ErrInvalidRefreshSessionID
	}
	if expectedVersion == 0 {
		return domain.ErrInvalidRefreshSessionVersion
	}
	if revokedAt.IsZero() {
		return domain.ErrInvalidRefreshSession
	}

	updateResult := r.db.WithContext(ctx).
		Model(&RefreshSessionPO{}).
		Where("session_id = ? AND version = ?", strings.TrimSpace(sessionID), expectedVersion).
		Updates(map[string]any{
			"revoked_at": &revokedAt,
			"version":    expectedVersion + 1,
			"updated_at": time.Now(),
		})
	if updateResult.Error != nil {
		return updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return domain.ErrRefreshSessionConflict
	}
	return nil
}

func (r *MySQLRefreshSessionRepository) DeleteExpiredRefreshSessions(
	ctx context.Context,
	expiredBefore time.Time,
	limit uint64,
) (uint64, error) {
	if r == nil || r.db == nil {
		return 0, ErrMySQLDBNil
	}
	if expiredBefore.IsZero() {
		return 0, domain.ErrInvalidRefreshTokenTTL
	}
	if limit == 0 {
		return 0, domain.ErrInvalidCleanupLimit
	}

	subQuery := r.db.WithContext(ctx).
		Model(&RefreshSessionPO{}).
		Select("id").
		Where("expires_at < ?", expiredBefore).
		Limit(int(limit))
	deleteResult := r.db.WithContext(ctx).
		Where("id IN (?)", subQuery).
		Delete(&RefreshSessionPO{})
	if deleteResult.Error != nil {
		return 0, deleteResult.Error
	}
	return uint64(deleteResult.RowsAffected), nil
}
