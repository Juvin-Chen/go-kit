package domain

import (
	// subtle：固定时间比较：不管对不对都跑完所有字符，速度永远一样，防止黑客可以靠 “比较时间” 猜出密码产生时间攻击
	"crypto/subtle"

	"strings"
	"time"
)

// RefreshSession 刷新令牌会话结构体
type RefreshSession struct {
	SessionID        string     // 会话ID，一次登录对应一个 ，用于唯一标识会话，方便查找
	UserID           string     // 用户ID
	RefreshTokenHash string     // 令牌哈希值
	IssuedAt         time.Time  // 令牌创建时间
	ExpiresAt        time.Time  // 令牌过期时间
	RevokedAt        *time.Time // 令牌被撤销时间
	Version          uint64     // 令牌版本号
}

func NewRefreshSession(
	sessionID string,
	userID string,
	refreshTokenHash string,
	issuedAt time.Time,
	expiresAt time.Time,
) (*RefreshSession, error) {
	session := &RefreshSession{
		SessionID:        strings.TrimSpace(sessionID),
		UserID:           strings.TrimSpace(userID),
		RefreshTokenHash: strings.TrimSpace(refreshTokenHash),
		IssuedAt:         issuedAt,
		ExpiresAt:        expiresAt,
		Version:          1,
	}
	if err := session.Validate(); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *RefreshSession) Validate() error {
	if s == nil {
		return ErrInvalidRefreshSession
	}
	if strings.TrimSpace(s.SessionID) == "" {
		return ErrInvalidRefreshSessionID
	}
	if strings.TrimSpace(s.UserID) == "" {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(s.RefreshTokenHash) == "" {
		return ErrInvalidRefreshTokenHash
	}
	if s.IssuedAt.IsZero() {
		return ErrInvalidRefreshToken
	}
	if s.ExpiresAt.IsZero() || !s.ExpiresAt.After(s.IssuedAt) {
		return ErrInvalidRefreshTokenTTL
	}
	if s.Version == 0 {
		return ErrInvalidRefreshSessionVersion
	}
	return nil
}

func RebuildRefreshSession(
	sessionID string,
	userID string,
	refreshTokenHash string,
	issuedAt time.Time,
	expiresAt time.Time,
	revokedAt *time.Time,
	version uint64,
) (*RefreshSession, error) {
	session := &RefreshSession{
		SessionID:        strings.TrimSpace(sessionID),
		UserID:           strings.TrimSpace(userID),
		RefreshTokenHash: strings.TrimSpace(refreshTokenHash),
		IssuedAt:         issuedAt,
		ExpiresAt:        expiresAt,
		RevokedAt:        revokedAt,
		Version:          version,
	}
	if err := session.Validate(); err != nil {
		return nil, err
	}
	return session, nil
}

// EnsureActive 确保会话令牌是否有效
func (s *RefreshSession) EnsureActive(now time.Time) error {
	if s == nil {
		return ErrInvalidRefreshSession
	}
	if now.IsZero() {
		now = time.Now()
	}
	if s.RevokedAt != nil {
		return ErrRefreshSessionRevoked
	}
	if !now.Before(s.ExpiresAt) {
		return ErrRefreshSessionExpired
	}
	return nil
}

// VerifyTokenHash 验证会话令牌哈希值是否匹配
func (s *RefreshSession) VerifyTokenHash(tokenHash string) error {
	if s == nil {
		return ErrInvalidRefreshSession
	}
	if strings.TrimSpace(tokenHash) == "" {
		return ErrInvalidRefreshToken
	}
	expectedHashBytes := []byte(s.RefreshTokenHash)
	actualHashBytes := []byte(strings.TrimSpace(tokenHash))
	if len(expectedHashBytes) != len(actualHashBytes) {
		return ErrInvalidRefreshToken
	}
	if subtle.ConstantTimeCompare(expectedHashBytes, actualHashBytes) != 1 {
		return ErrInvalidRefreshToken
	}
	return nil
}

// Rotate 用新令牌哈希替换当前会话令牌。
// 当前实现为“单会话原地轮换”，适合基础版本。
// 后续可演进为 token family + replay detection。
func (s *RefreshSession) Rotate(newRefreshTokenHash string, newExpiresAt time.Time, now time.Time) error {
	if s == nil {
		return ErrInvalidRefreshSession
	}
	if err := s.EnsureActive(now); err != nil {
		return err
	}
	newRefreshTokenHash = strings.TrimSpace(newRefreshTokenHash)
	if newRefreshTokenHash == "" {
		return ErrInvalidRefreshTokenHash
	}
	if newExpiresAt.IsZero() {
		return ErrInvalidRefreshTokenTTL
	}
	if now.IsZero() {
		now = time.Now()
	}
	if !newExpiresAt.After(now) {
		return ErrInvalidRefreshTokenTTL
	}
	s.RefreshTokenHash = newRefreshTokenHash
	s.IssuedAt = now
	s.ExpiresAt = newExpiresAt
	s.Version++
	return nil
}

// Revoke 撤销会话令牌
func (s *RefreshSession) Revoke(now time.Time) error {
	if s == nil {
		return ErrInvalidRefreshSession
	}
	if now.IsZero() {
		now = time.Now()
	}
	s.RevokedAt = &now
	return nil
}
