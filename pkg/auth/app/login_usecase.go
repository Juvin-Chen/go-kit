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
	accessTokenProvider      AccessTokenProvider
	refreshTokenTTL          time.Duration
	accessTokenTTL           time.Duration
}

func NewLoginUseCase(
	refreshSessionRepository domain.RefreshSessionRepository,
	refreshTokenHasher RefreshTokenHasher,
	accessTokenProvider AccessTokenProvider,
	refreshTokenTTL time.Duration,
	accessTokenTTL time.Duration,
) *LoginUseCase {
	return &LoginUseCase{
		refreshSessionRepository: refreshSessionRepository,
		refreshTokenHasher:       refreshTokenHasher,
		accessTokenProvider:      accessTokenProvider,
		refreshTokenTTL:          refreshTokenTTL,
		accessTokenTTL:           accessTokenTTL,
	}
}

// Login 执行登录流程并创建 refresh 会话
func (uc *LoginUseCase) Login(ctx context.Context, command LoginCommand) (*LoginResult, error) {
	if uc == nil {
		return nil, ErrInvalidUseCase
	}
	// 依赖注入完整性校验
	if uc.refreshSessionRepository == nil || uc.refreshTokenHasher == nil || uc.accessTokenProvider == nil || uc.refreshTokenTTL <= 0 || uc.accessTokenTTL <= 0 {
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
	issuedAt := time.Now().UTC()
	refreshTokenExpiresAt := issuedAt.Add(uc.refreshTokenTTL).UTC()
	if refreshTokenExpiresAt.IsZero() || !refreshTokenExpiresAt.After(issuedAt) {
		return nil, domain.ErrInvalidRefreshTokenTTL
	}

	// 先签发 access token，确保后续会话写库成功时，令牌已可用
	signedAccessToken, err := uc.accessTokenProvider.GenerateAccessToken(command.UserID, uc.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(signedAccessToken) == "" {
		return nil, ErrInvalidAccessToken
	}
	accessTokenExpiresAt := issuedAt.Add(uc.accessTokenTTL).UTC()
	if accessTokenExpiresAt.IsZero() || !accessTokenExpiresAt.After(issuedAt) {
		return nil, ErrInvalidAccessTokenTTL
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
		issuedAt,
		refreshTokenExpiresAt,
	)
	if err != nil {
		return nil, err
	}

	// 将会话保存到数据库
	if err = uc.refreshSessionRepository.CreateRefreshSession(ctx, session); err != nil {
		return nil, err
	}

	// 返回创建成功的会话结果
	return &LoginResult{
		SessionID:             session.SessionID,
		UserID:                session.UserID,
		AccessToken:           signedAccessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshTokenExpiresAt: session.ExpiresAt.UTC(),
	}, nil
}
