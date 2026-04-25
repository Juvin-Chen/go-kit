package jwt

import (
	"errors"
	"strings"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/app"
	golangjwt "github.com/golang-jwt/jwt/v5"
)

var ErrAccessTokenSecretEmpty = errors.New("access token secret is empty")
var ErrAccessTokenSecretTooShort = errors.New("access token secret is too short")
var ErrAccessTokenTTLInvalid = errors.New("access token ttl is invalid")
var ErrAccessTokenUserIDEmpty = errors.New("access token user id is empty")
var ErrAccessTokenParseFailed = errors.New("access token parse failed")
var ErrAccessTokenInvalidSigningMethod = errors.New("access token signing method is invalid")
var ErrAccessTokenInvalidClaims = errors.New("access token claims are invalid")

const minAccessTokenSecretLength = 16

type JWTProvider struct {
	secretKey []byte
	nowFn     func() time.Time
}

var _ app.AccessTokenProvider = (*JWTProvider)(nil)

// NewJWTProvider 创建 JWT Provider, 校验密钥合法性
func NewJWTProvider(secretKey string) (*JWTProvider, error) {
	trimmedSecretKey := strings.TrimSpace(secretKey)
	if trimmedSecretKey == "" {
		return nil, ErrAccessTokenSecretEmpty
	}
	if len(trimmedSecretKey) < minAccessTokenSecretLength {
		return nil, ErrAccessTokenSecretTooShort
	}
	return &JWTProvider{
		secretKey: []byte(trimmedSecretKey),
		nowFn: func() time.Time {
			return time.Now().UTC()
		},
	}, nil
}

// GenerateAccessToken 生成 AccessToken, claims 包含 user_id 与 exp
func (provider *JWTProvider) GenerateAccessToken(userID string, expiry time.Duration) (string, error) {
	if provider == nil || len(provider.secretKey) == 0 {
		return "", ErrAccessTokenSecretEmpty
	}
	trimmedUserID := strings.TrimSpace(userID)
	if trimmedUserID == "" {
		return "", ErrAccessTokenUserIDEmpty
	}
	if expiry <= 0 {
		return "", ErrAccessTokenTTLInvalid
	}
	claims := golangjwt.MapClaims{
		"user_id": trimmedUserID,
		"exp":     provider.nowFn().UTC().Add(expiry).Unix(),
	}
	token := golangjwt.NewWithClaims(golangjwt.SigningMethodHS256, claims)
	return token.SignedString(provider.secretKey)
}

// ParseAccessToken 解析并校验 AccessToken, 返回 user_id
func (provider *JWTProvider) ParseAccessToken(token string) (string, error) {
	if provider == nil || len(provider.secretKey) == 0 {
		return "", ErrAccessTokenSecretEmpty
	}
	trimmedToken := strings.TrimSpace(token)
	if trimmedToken == "" {
		return "", ErrAccessTokenParseFailed
	}
	parsedToken, err := golangjwt.Parse(trimmedToken, func(parsedToken *golangjwt.Token) (interface{}, error) {
		if parsedToken.Method == nil || parsedToken.Method.Alg() != golangjwt.SigningMethodHS256.Alg() {
			return nil, ErrAccessTokenInvalidSigningMethod
		}
		return provider.secretKey, nil
	})
	if err != nil {
		return "", ErrAccessTokenParseFailed
	}
	if parsedToken == nil || !parsedToken.Valid {
		return "", ErrAccessTokenInvalidClaims
	}
	mapClaims, typeAssertionOK := parsedToken.Claims.(golangjwt.MapClaims)
	if !typeAssertionOK {
		return "", ErrAccessTokenInvalidClaims
	}
	userIDValue, userIDExists := mapClaims["user_id"]
	if !userIDExists {
		return "", ErrAccessTokenInvalidClaims
	}
	userID, typeAssertionOK := userIDValue.(string)
	if !typeAssertionOK || strings.TrimSpace(userID) == "" {
		return "", ErrAccessTokenInvalidClaims
	}
	return strings.TrimSpace(userID), nil
}
