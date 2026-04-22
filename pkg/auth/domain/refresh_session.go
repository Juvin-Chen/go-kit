package domain

import (
	"crypto/subtle"
	"strings"
	"time"
)

// 刷新令牌的会话档案
type RefreshSession struct {
	SessionID        string
	UserID           string
	RefreshTokenHash string
	IssuedAt         time.Time
	ExpiresAt        time.Time
	RevokedAt        *time.Time
	Version          uint64
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

func (s *RefreshSession) Validate() error {
	if s == nil {
		return ErrInvalidRefreshToken
	}
	if strings.TrimSpace(s.SessionID) == "" {
		return ErrInvalidRefreshSessionID
	}
	if strings.TrimSpace(s.UserID) == "" {
		return ErrInvalidUserID
	}
	if strings.TrimSpace(s.RefreshTokenHash) == "" {
		return ErrRefreshTokenHash
	}
	if s.IssuedAt.IsZero() {
		return ErrInvalidRefreshToken
	}
	if s.ExpiresAt.IsZero() || !s.ExpiresAt.After(s.IssuedAt) {
		return ErrInvalidRefreshTokenTTL
	}
	return nil
}

func (s *RefreshSession) EnsureActive(now time.Time) error {
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

func (s *RefreshSession) VerifyTokenHash(tokenHash string) error {
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

// Rotate 用新令牌哈希替换当前会话令牌，实现 refresh token 轮换。
// 当前版本采用“同一 session 原地轮换”；后续可演进为 token family + replay detection。
func (s *RefreshSession) Rotate(newRefreshTokenHash string, newExpiresAt time.Time, now time.Time) error {
	if err := s.EnsureActive(now); err != nil {
		return err
	}
	newRefreshTokenHash = strings.TrimSpace(newRefreshTokenHash)
	if newRefreshTokenHash == "" {
		return ErrRefreshTokenHash
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

func (s *RefreshSession) Revoke(now time.Time) {
	if now.IsZero() {
		now = time.Now()
	}
	s.RevokedAt = &now
}
