package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"

	"github.com/Juvin-Chen/go-kit/pkg/auth/app"
)

// 哈希器密钥配置错误
var ErrRefreshTokenHasherSecretEmpty = errors.New("refresh token hasher secret is empty")
var ErrRefreshTokenHasherSecretTooShort = errors.New("refresh token hasher secret is too short")

// 哈希输入参数错误
var ErrInvalidPlainRefreshToken = errors.New("invalid refresh token")

// 最小密钥长度为 16 字节
const minRefreshTokenHasherSecretLength = 16

type RefreshTokenHMACHasher struct {
	secretKey []byte
}

var _ app.RefreshTokenHasher = (*RefreshTokenHMACHasher)(nil)

// 创建一个刷新令牌哈希器
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

// 对刷新令牌进行哈希
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
