package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

type refreshTestRepo struct {
	session               *domain.RefreshSession
	generatedBeforeUpdate *bool
	updateCalled          bool
}

func (r *refreshTestRepo) CreateRefreshSession(_ context.Context, _ *domain.RefreshSession) error {
	return nil
}

func (r *refreshTestRepo) GetRefreshSessionBySessionID(_ context.Context, _ string) (*domain.RefreshSession, error) {
	return r.session, nil
}

func (r *refreshTestRepo) UpdateRefreshSessionOnRotate(_ context.Context, _ *domain.RefreshSession, _ uint64) error {
	r.updateCalled = true
	if r.generatedBeforeUpdate != nil && !*r.generatedBeforeUpdate {
		return errors.New("access token not generated before update")
	}
	return nil
}

func (r *refreshTestRepo) RevokeRefreshSession(_ context.Context, _ string, _ time.Time, _ uint64) error {
	return nil
}

func (r *refreshTestRepo) DeleteExpiredRefreshSessions(_ context.Context, _ time.Time, _ uint64) (uint64, error) {
	return 0, nil
}

type refreshTestHasher struct{}

func (h *refreshTestHasher) HashRefreshToken(plainToken string) (string, error) {
	if plainToken == "" {
		return "", domain.ErrInvalidRefreshToken
	}
	return "hash-" + plainToken, nil
}

type refreshTestProvider struct {
	generated *bool
}

func (p *refreshTestProvider) GenerateAccessToken(_ string, _ time.Duration) (string, error) {
	if p.generated != nil {
		*p.generated = true
	}
	return "signed-access-token", nil
}

func (p *refreshTestProvider) ParseAccessToken(_ string) (string, error) {
	return "", nil
}

type refreshFailingProvider struct{}

func (p *refreshFailingProvider) GenerateAccessToken(_ string, _ time.Duration) (string, error) {
	return "", errors.New("provider failure")
}

func (p *refreshFailingProvider) ParseAccessToken(_ string) (string, error) {
	return "", nil
}

func TestRefreshUseCaseGeneratesAccessTokenBeforeRotateUpdate(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	session, err := domain.NewRefreshSession(
		"session-2001",
		"user-2001",
		"hash-current-token",
		now.Add(-10*time.Minute),
		now.Add(50*time.Minute),
	)
	if err != nil {
		t.Fatalf("NewRefreshSession returned error: %v", err)
	}

	generated := false
	repo := &refreshTestRepo{
		session:               session,
		generatedBeforeUpdate: &generated,
	}
	uc := NewRefreshTokenUseCase(
		repo,
		&refreshTestHasher{},
		&refreshTestProvider{generated: &generated},
		15*time.Minute,
	)

	result, err := uc.Refresh(context.Background(), RefreshTokenCommand{
		SessionID:        "session-2001",
		CurrentToken:     "current-token",
		NewRefreshToken:  "new-token",
		NewRefreshExpiry: now.Add(2 * time.Hour),
		Now:              now,
	})
	if err != nil {
		t.Fatalf("Refresh returned error: %v", err)
	}
	if !repo.updateCalled {
		t.Fatal("expected UpdateRefreshSessionOnRotate to be called")
	}
	if result == nil || result.AccessToken == "" {
		t.Fatal("expected access token in refresh result")
	}
}

func TestRefreshUseCaseStopsBeforeRotateWhenTokenGenerationFails(t *testing.T) {
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, time.UTC)
	session, err := domain.NewRefreshSession(
		"session-2002",
		"user-2002",
		"hash-current-token",
		now.Add(-10*time.Minute),
		now.Add(50*time.Minute),
	)
	if err != nil {
		t.Fatalf("NewRefreshSession returned error: %v", err)
	}

	repo := &refreshTestRepo{session: session}
	uc := NewRefreshTokenUseCase(
		repo,
		&refreshTestHasher{},
		&refreshFailingProvider{},
		15*time.Minute,
	)

	_, err = uc.Refresh(context.Background(), RefreshTokenCommand{
		SessionID:        "session-2002",
		CurrentToken:     "current-token",
		NewRefreshToken:  "new-token",
		NewRefreshExpiry: now.Add(2 * time.Hour),
		Now:              now,
	})
	if err == nil {
		t.Fatal("expected error when token generation fails")
	}
	if repo.updateCalled {
		t.Fatal("UpdateRefreshSessionOnRotate must not be called when token generation fails")
	}
}
