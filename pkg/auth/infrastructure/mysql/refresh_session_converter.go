package mysql

import "github.com/Juvin-Chen/go-kit/pkg/auth/domain"

// 创建/查询时会用到这两个函数

func ToRefreshSessionPO(session *domain.RefreshSession) *RefreshSessionPO {
	if session == nil {
		return nil
	}
	return &RefreshSessionPO{
		SessionID:        session.SessionID,
		UserID:           session.UserID,
		RefreshTokenHash: session.RefreshTokenHash,
		IssuedAt:         session.IssuedAt,
		ExpiresAt:        session.ExpiresAt,
		RevokedAt:        session.RevokedAt,
		Version:          session.Version,
	}
}

func ToDomainRefreshSession(po *RefreshSessionPO) (*domain.RefreshSession, error) {
	if po == nil {
		return nil, domain.ErrInvalidRefreshSession
	}
	return domain.RebuildRefreshSession(
		po.SessionID,
		po.UserID,
		po.RefreshTokenHash,
		po.IssuedAt,
		po.ExpiresAt,
		po.RevokedAt,
		po.Version,
	)
}
