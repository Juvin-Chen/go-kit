package mysql

import "time"

type RefreshSessionPO struct {
	ID               uint64     `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement"`
	SessionID        string     `gorm:"column:session_id;type:varchar(64);not null;uniqueIndex:uk_refresh_session_id"`
	UserID           string     `gorm:"column:user_id;type:varchar(64);not null;index:idx_refresh_user_id;index:idx_user_active,priority:1"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash;type:varchar(64);not null"`
	IssuedAt         time.Time  `gorm:"column:issued_at;type:datetime(3);not null"`
	ExpiresAt        time.Time  `gorm:"column:expires_at;type:datetime(3);not null;index:idx_refresh_expires_at;index:idx_user_active,priority:3"`
	RevokedAt        *time.Time `gorm:"column:revoked_at;type:datetime(3);index:idx_user_active,priority:2"`
	Version          uint64     `gorm:"column:version;type:bigint unsigned;not null;default:1"`
	CreatedAt        time.Time  `gorm:"column:created_at;type:datetime(3);autoCreateTime:milli"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;type:datetime(3);autoUpdateTime:milli"`
}

func (RefreshSessionPO) TableName() string {
	return "refresh_sessions"
}
