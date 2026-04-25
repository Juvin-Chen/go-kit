package jwt

import (
	"errors"
	"testing"
	"time"

	golangjwt "github.com/golang-jwt/jwt/v5"
)

func TestJWTProviderGenerateAndParseAccessToken(t *testing.T) {
	provider, err := NewJWTProvider("1234567890abcdef")
	if err != nil {
		t.Fatalf("NewJWTProvider returned error: %v", err)
	}

	accessToken, err := provider.GenerateAccessToken("user-1001", 30*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}
	if accessToken == "" {
		t.Fatal("GenerateAccessToken returned empty token")
	}

	parsedUserID, err := provider.ParseAccessToken(accessToken)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}
	if parsedUserID != "user-1001" {
		t.Fatalf("parsed user id mismatch, got: %s", parsedUserID)
	}
}

func TestJWTProviderParseExpiredToken(t *testing.T) {
	provider, err := NewJWTProvider("1234567890abcdef")
	if err != nil {
		t.Fatalf("NewJWTProvider returned error: %v", err)
	}
	provider.nowFn = func() time.Time {
		return time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)
	}

	accessToken, err := provider.GenerateAccessToken("user-1002", time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	provider.nowFn = func() time.Time {
		return time.Date(2026, 1, 1, 10, 2, 0, 0, time.UTC)
	}
	_, err = provider.ParseAccessToken(accessToken)
	if !errors.Is(err, ErrAccessTokenParseFailed) {
		t.Fatalf("expected ErrAccessTokenParseFailed, got: %v", err)
	}
}

func TestJWTProviderParseRejectsInvalidSigningMethod(t *testing.T) {
	provider, err := NewJWTProvider("1234567890abcdef")
	if err != nil {
		t.Fatalf("NewJWTProvider returned error: %v", err)
	}

	unsignedToken := golangjwt.NewWithClaims(golangjwt.SigningMethodNone, golangjwt.MapClaims{
		"user_id": "user-1003",
		"exp":     time.Now().UTC().Add(10 * time.Minute).Unix(),
	})
	accessToken, err := unsignedToken.SignedString(golangjwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("SignedString returned error: %v", err)
	}

	_, err = provider.ParseAccessToken(accessToken)
	if !errors.Is(err, ErrAccessTokenParseFailed) {
		t.Fatalf("expected ErrAccessTokenParseFailed, got: %v", err)
	}
}

func TestJWTProviderGenerateAccessTokenInputValidation(t *testing.T) {
	provider, err := NewJWTProvider("1234567890abcdef")
	if err != nil {
		t.Fatalf("NewJWTProvider returned error: %v", err)
	}

	_, err = provider.GenerateAccessToken("", time.Minute)
	if !errors.Is(err, ErrAccessTokenUserIDEmpty) {
		t.Fatalf("expected ErrAccessTokenUserIDEmpty, got: %v", err)
	}

	_, err = provider.GenerateAccessToken("user-1004", 0)
	if !errors.Is(err, ErrAccessTokenTTLInvalid) {
		t.Fatalf("expected ErrAccessTokenTTLInvalid, got: %v", err)
	}
}
