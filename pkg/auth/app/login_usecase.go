package app

import (
	"context"
	"strings"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

type LoginUseCase struct {
	refreshSessionRepository domain.RefreshSessionRepository
	refreshTokenHasher       RefreshTokenHasher
	accessTokenIssuer        AccessTokenIssuer
}

func NewLoginUseCase(
	refreshSessionRepository domain.RefreshSessionRepository,
	refreshTokenHasher RefreshTokenHasher,
	accessTokenIssuer AccessTokenIssuer,
) *LoginUseCase {
	return &LoginUseCase{
		refreshSessionRepository: refreshSessionRepository,
		refreshTokenHasher:       refreshTokenHasher,
		accessTokenIssuer:        accessTokenIssuer,
	}
}

// Login 执行登录流程并创建 refresh 会话
func (uc *LoginUseCase) Login(ctx context.Context, command LoginCommand) (*LoginResult, error) {
	if uc == nil {
		return nil, ErrInvalidUseCase
	}
	// 依赖注入完整性校验
	if uc.refreshSessionRepository == nil || uc.refreshTokenHasher == nil || uc.accessTokenIssuer == nil {
		return nil, ErrInvalidUseCaseDependency
	}
	if strings.TrimSpace(command.SessionID) == "" {
		return nil, domain.ErrInvalidRefreshSessionID
	}
	if strings.TrimSpace(command.UserID) == "" {
		return nil, domain.ErrInvalidUserID
	}
	if strings.TrimSpace(command.RefreshToken) == "" {
		return nil, domain.ErrInvalidRefreshToken
	}
	if command.IssuedAt.IsZero() {
		command.IssuedAt = time.Now().UTC()
	} else {
		command.IssuedAt = command.IssuedAt.UTC()
	}
	if !command.ExpiresAt.IsZero() {
		command.ExpiresAt = command.ExpiresAt.UTC()
	}
	if command.ExpiresAt.IsZero() || !command.ExpiresAt.After(command.IssuedAt) {
		return nil, domain.ErrInvalidRefreshTokenTTL
	}

	// 对明文 RefreshToken 进行哈希加密，数据库不存储明文
	refreshTokenHash, err := uc.refreshTokenHasher.HashRefreshToken(command.RefreshToken)
	if err != nil {
		return nil, err
	}

	// 调用领域层创建会话对象（会自动做业务规则校验）
	session, err := domain.NewRefreshSession(
		command.SessionID,
		command.UserID,
		refreshTokenHash,
		command.IssuedAt,
		command.ExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	// 将会话保存到数据库
	if err = uc.refreshSessionRepository.CreateRefreshSession(ctx, session); err != nil {
		return nil, err
	}

	// access token 属于应用流程产物，不进入领域模型
	accessTokenResult, err := uc.accessTokenIssuer.IssueAccessToken(AccessTokenCommand{
		UserID:    session.UserID,
		SessionID: session.SessionID,
		IssuedAt:  command.IssuedAt,
	})
	if err != nil {
		return nil, err
	}
	if accessTokenResult == nil || strings.TrimSpace(accessTokenResult.AccessToken) == "" {
		return nil, ErrInvalidAccessToken
	}
	if accessTokenResult.ExpiresAt.IsZero() || !accessTokenResult.ExpiresAt.After(command.IssuedAt) {
		return nil, ErrInvalidAccessTokenTTL
	}

	// 返回创建成功的会话结果
	return &LoginResult{
		SessionID:             session.SessionID,
		UserID:                session.UserID,
		AccessToken:           accessTokenResult.AccessToken,
		AccessTokenExpiresAt:  accessTokenResult.ExpiresAt,
		RefreshTokenExpiresAt: session.ExpiresAt.UTC(),
	}, nil
}
