package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/Juvin-Chen/go-kit/pkg/auth/app"
)

// errors
var ErrRefreshTokenHasherSecretEmpty = errors.New("refresh token hasher secret is empty")
var ErrRefreshTokenHasherSecretTooShort = errors.New("refresh token hasher secret is too short")
var ErrInvalidPlainRefreshToken = errors.New("invalid refresh token")

const minRefreshTokenHasherSecretLength = 16

type RefreshTokenHMACHasher struct {
	secretKey []byte
}

var _ app.RefreshTokenHasher = (*RefreshTokenHMACHasher)(nil)

func NewRefreshTokenHMACHasher(secretKey string) (*RefreshTokenHMACHasher, error) {
	trimmedSecretKey := strings.TrimSpace(secretKey)
	if trimmedSecretKey == "" {
		return nil, ErrRefreshTokenHasherSecretEmpty
	}
	if len(trimmedSecretKey) < minRefreshTokenHasherSecretLength {
		return nil, ErrRefreshTokenHasherSecretTooShort
	}
	return &RefreshTokenHMACHasher{
		secretKey: []byte(trimmedSecretKey),
	}, nil
}

func (h *RefreshTokenHMACHasher) HashRefreshToken(plainToken string) (string, error) {
	if h == nil || len(h.secretKey) == 0 {
		return "", ErrRefreshTokenHasherSecretEmpty
	}
	trimmedToken := strings.TrimSpace(plainToken)
	if trimmedToken == "" {
		return "", ErrInvalidPlainRefreshToken
	}

	mac := hmac.New(sha256.New, h.secretKey)
	_, _ = mac.Write([]byte(trimmedToken))
	return hex.EncodeToString(mac.Sum(nil)), nil
}
