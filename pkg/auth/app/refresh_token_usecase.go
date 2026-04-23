package app

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

type RefreshTokenUseCase struct {
	refreshSessionRepository domain.RefreshSessionRepository
	refreshTokenHasher       RefreshTokenHasher
}

func NewRefreshTokenUseCase(
	refreshSessionRepository domain.RefreshSessionRepository,
	refreshTokenHasher RefreshTokenHasher,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		refreshSessionRepository: refreshSessionRepository,
		refreshTokenHasher:       refreshTokenHasher,
	}
}

// Refresh 执行 refresh token 轮换流程
// 流程：校验参数 → 查询会话 → 校验有效性 → 验证旧令牌 → 生成新令牌 → 乐观锁更新
func (uc *RefreshTokenUseCase) Refresh(ctx context.Context, command RefreshTokenCommand) (*RefreshTokenResult, error) {
	if uc == nil {
		return nil, domain.ErrInvalidRefreshSession
	}
	if uc.refreshSessionRepository == nil || uc.refreshTokenHasher == nil {
		return nil, domain.ErrInvalidRefreshSession
	}
	if strings.TrimSpace(command.SessionID) == "" {
		return nil, domain.ErrInvalidRefreshSessionID
	}
	if strings.TrimSpace(command.CurrentToken) == "" || strings.TrimSpace(command.NewRefreshToken) == "" {
		return nil, domain.ErrInvalidRefreshToken
	}
	if command.Now.IsZero() {
		// 应用层统一使用 UTC，避免多时区服务时间语义不一致
		command.Now = time.Now().UTC()
	} else {
		command.Now = command.Now.UTC()
	}
	if command.NewRefreshExpiry.IsZero() {
		return nil, domain.ErrInvalidRefreshTokenTTL
	}
	command.NewRefreshExpiry = command.NewRefreshExpiry.UTC()

	// 乐观锁冲突重试机制，最多重试 2 次，避免并发刷新失败
	const maxRetry = 2
	for retryCount := 0; retryCount < maxRetry; retryCount++ {
		// 1. 根据会话ID查询当前会话信息
		session, err := uc.refreshSessionRepository.GetRefreshSessionBySessionID(ctx, command.SessionID)
		if err != nil {
			return nil, err
		}

		// 2. 校验会话是否有效：未过期、未被撤销
		if err = session.EnsureActive(command.Now); err != nil {
			return nil, err
		}

		// 3. 对当前传入的旧令牌做哈希
		currentTokenHash, err := uc.refreshTokenHasher.HashRefreshToken(command.CurrentToken)
		if err != nil {
			return nil, err
		}

		// 4. 校验旧令牌是否与数据库记录匹配
		if err = session.VerifyTokenHash(currentTokenHash); err != nil {
			return nil, err
		}

		// 5. 新令牌加密
		newRefreshTokenHash, err := uc.refreshTokenHasher.HashRefreshToken(command.NewRefreshToken)
		if err != nil {
			return nil, err
		}

		// 6. 记住旧版本号
		oldVersion := session.Version

		// 7. 执行轮换：
		// 内存里：换令牌、换过期时间、版本号 +1
		if err = session.Rotate(newRefreshTokenHash, command.NewRefreshExpiry, command.Now); err != nil {
			return nil, err
		}

		// 8，执行数据库更新
		err = uc.refreshSessionRepository.UpdateRefreshSessionOnRotate(ctx, session, oldVersion)
		if err == nil {
			return &RefreshTokenResult{RefreshSession: session}, nil
		}
		if !errors.Is(err, domain.ErrRefreshSessionConflict) {
			return nil, err
		}
	}
	return nil, domain.ErrRefreshSessionConflict
}
