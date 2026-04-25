package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

type loginTestRepo struct {
	generatedBeforeCreate *bool
	createCalled          bool
}

func (r *loginTestRepo) CreateRefreshSession(_ context.Context, _ *domain.RefreshSession) error {
	r.createCalled = true
	if r.generatedBeforeCreate != nil && !*r.generatedBeforeCreate {
		return errors.New("access token not generated before create")
	}
	return nil
}

func (r *loginTestRepo) GetRefreshSessionBySessionID(_ context.Context, _ string) (*domain.RefreshSession, error) {
	return nil, domain.ErrRefreshSessionNotFound
}

func (r *loginTestRepo) UpdateRefreshSessionOnRotate(_ context.Context, _ *domain.RefreshSession, _ uint64) error {
	return nil
}

func (r *loginTestRepo) RevokeRefreshSession(_ context.Context, _ string, _ time.Time, _ uint64) error {
	return nil
}

func (r *loginTestRepo) DeleteExpiredRefreshSessions(_ context.Context, _ time.Time, _ uint64) (uint64, error) {
	return 0, nil
}

type loginTestHasher struct{}

func (h *loginTestHasher) HashRefreshToken(plainToken string) (string, error) {
	if plainToken == "" {
		return "", domain.ErrInvalidRefreshToken
	}
	return "hmac-sha256-hash-value-hmac-sha256-hash-value-1234567890abcdef123456", nil
}

type loginTestProvider struct {
	generated *bool
}

func (p *loginTestProvider) GenerateAccessToken(_ string, _ time.Duration) (string, error) {
	if p.generated != nil {
		*p.generated = true
	}
	return "signed-access-token", nil
}

func (p *loginTestProvider) ParseAccessToken(_ string) (string, error) {
	return "", nil
}

type loginFailingProvider struct{}

func (p *loginFailingProvider) GenerateAccessToken(_ string, _ time.Duration) (string, error) {
	return "", errors.New("provider failure")
}

func (p *loginFailingProvider) ParseAccessToken(_ string) (string, error) {
	return "", nil
}

func TestLoginUseCaseGeneratesAccessTokenBeforeCreateSession(t *testing.T) {
	generated := false
	repo := &loginTestRepo{generatedBeforeCreate: &generated}
	uc := NewLoginUseCase(
		repo,
		&loginTestHasher{},
		&loginTestProvider{generated: &generated},
		time.Hour,
		15*time.Minute,
	)

	result, err := uc.Login(context.Background(), LoginCommand{
		SessionID:    "session-1001",
		UserID:       "user-1001",
		RefreshToken: "refresh-token-1001",
	})
	if err != nil {
		t.Fatalf("Login returned error: %v", err)
	}
	if !repo.createCalled {
		t.Fatal("expected CreateRefreshSession to be called")
	}
	if result == nil || result.AccessToken == "" {
		t.Fatal("expected access token in login result")
	}
}

func TestLoginUseCaseStopsBeforeCreateWhenTokenGenerationFails(t *testing.T) {
	repo := &loginTestRepo{}
	uc := NewLoginUseCase(
		repo,
		&loginTestHasher{},
		&loginFailingProvider{},
		time.Hour,
		15*time.Minute,
	)

	_, err := uc.Login(context.Background(), LoginCommand{
		SessionID:    "session-1002",
		UserID:       "user-1002",
		RefreshToken: "refresh-token-1002",
	})
	if err == nil {
		t.Fatal("expected error when token generation fails")
	}
	if repo.createCalled {
		t.Fatal("CreateRefreshSession must not be called when token generation fails")
	}
}
