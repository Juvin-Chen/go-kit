package mysql

import "time"

type RefreshSessionPO struct {
	ID               uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	SessionID        string     `gorm:"column:session_id;type:varchar(64);not null;uniqueIndex:uk_refresh_session_id"`
	UserID           string     `gorm:"column:user_id;type:varchar(64);not null;index:idx_refresh_user_id"`
	RefreshTokenHash string     `gorm:"column:refresh_token_hash;type:varchar(255);not null"`
	IssuedAt         time.Time  `gorm:"column:issued_at;not null"`
	ExpiresAt        time.Time  `gorm:"column:expires_at;not null;index:idx_refresh_expires_at"`
	RevokedAt        *time.Time `gorm:"column:revoked_at"`
	Version          uint64     `gorm:"column:version;not null"`
	CreatedAt        time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (RefreshSessionPO) TableName() string {
	return "refresh_sessions"
}
