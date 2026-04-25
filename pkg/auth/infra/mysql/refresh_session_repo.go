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

// MySQL 仓储初始化错误
var ErrMySQLDBNil = errors.New("mysql repository db is nil")

func NewMySQLRefreshSessionRepository(db *gorm.DB) *MySQLRefreshSessionRepository {
	return &MySQLRefreshSessionRepository{db: db}
}

// UpdateRefreshSessionOnRotate 使用版本号执行乐观锁更新
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
			"updated_at":         time.Now().UTC(),
		})
	if updateResult.Error != nil {
		return updateResult.Error
	}
	if updateResult.RowsAffected == 0 {
		return domain.ErrRefreshSessionConflict
	}
	return nil
}

// CreateRefreshSession 持久化一条新的 refresh 会话记录
func (r *MySQLRefreshSessionRepository) CreateRefreshSession(ctx context.Context, session *domain.RefreshSession) error {
	if r == nil || r.db == nil {
		return ErrMySQLDBNil
	}
	if session == nil {
		return domain.ErrInvalidRefreshSession
	}

	createResult := r.db.WithContext(ctx).Create(ToRefreshSessionPO(session))
	if createResult.Error != nil {
		// session_id作为唯一键，同一个 session_id 被重复插入时，MySQL 会返回 1062 duplicate key
		// 基础设施层在此处把数据库错误转换为领域错误，避免把数据库细节泄漏到 application
		if isDuplicateKeyError(createResult.Error) {
			return domain.ErrRefreshSessionConflict
		}
		return createResult.Error
	}
	return nil
}

// GetRefreshSessionBySessionID 根据会话 ID 查询 refresh 会话记录
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

// RevokeRefreshSession 撤销 refresh 会话
// 功能：把一个 refresh token 标记为“已撤销”，让它立刻失效
func (r *MySQLRefreshSessionRepository) RevokeRefreshSession(
	ctx context.Context,
	sessionID string,
	revokedAt time.Time,
	expectedVersion uint64,
) error {
	if r == nil || r.db == nil {
		return ErrMySQLDBNil
	}
	trimmedSessionID := strings.TrimSpace(sessionID)
	if trimmedSessionID == "" {
		return domain.ErrInvalidRefreshSessionID
	}
	if expectedVersion == 0 {
		return domain.ErrInvalidRefreshSessionVersion
	}
	if revokedAt.IsZero() {
		return domain.ErrInvalidRefreshSession
	}

	// 幂等，这里是第一层幂等，WHERE session_id = ? AND version = ? AND revoked_at IS NULL
	updateResult := r.db.WithContext(ctx).
		Model(&RefreshSessionPO{}).
		Where("session_id = ? AND version = ? AND revoked_at IS NULL", trimmedSessionID, expectedVersion).
		Updates(map[string]any{
			"revoked_at": &revokedAt,
			"version":    expectedVersion + 1,
			"updated_at": time.Now().UTC(),
		})
	if updateResult.Error != nil {
		return updateResult.Error
	}
	// 第二层幂等，兜底
	if updateResult.RowsAffected == 0 {
		var currentState RefreshSessionPO
		stateQueryResult := r.db.WithContext(ctx).
			Select("revoked_at", "version").
			Where("session_id = ?", trimmedSessionID).
			First(&currentState)
		if stateQueryResult.Error != nil {
			if errors.Is(stateQueryResult.Error, gorm.ErrRecordNotFound) {
				return domain.ErrRefreshSessionNotFound
			}
			return stateQueryResult.Error
		}
		if currentState.RevokedAt != nil {
			return nil
		}
		return domain.ErrRefreshSessionConflict
	}
	return nil
}

// DeleteExpiredRefreshSessions 删除过期的 refresh 会话记录，释放空间
// expiredBefore = 在此时间之前过期的 → 都删掉
// limit = 一次最多删多少条（防止一次删太多把数据库卡爆）
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
