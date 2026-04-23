package app

import (
	"context"
	"strings"
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

type LogoutUseCase struct {
	refreshSessionRepository domain.RefreshSessionRepository
}

func NewLogoutUseCase(refreshSessionRepository domain.RefreshSessionRepository) *LogoutUseCase {
	return &LogoutUseCase{refreshSessionRepository: refreshSessionRepository}
}

// Logout 执行登出流程并撤销 refresh 会话
func (uc *LogoutUseCase) Logout(ctx context.Context, command LogoutCommand) error {
	if uc == nil || uc.refreshSessionRepository == nil {
		return ErrInvalidUseCaseDependency
	}
	if strings.TrimSpace(command.SessionID) == "" {
		return domain.ErrInvalidRefreshSessionID
	}
	if command.RevokedAt.IsZero() {
		command.RevokedAt = time.Now().UTC()
	} else {
		command.RevokedAt = command.RevokedAt.UTC()
	}

	// 根据会话ID从数据库查询当前会话信息
	session, err := uc.refreshSessionRepository.GetRefreshSessionBySessionID(ctx, command.SessionID)
	if err != nil {
		return err
	}
	// 已撤销会话直接返回成功，保证 logout 幂等
	if session.RevokedAt != nil {
		return nil
	}
	// 记录更新前的旧版本号，用于乐观锁
	oldVersion := session.Version
	if err = session.Revoke(command.RevokedAt); err != nil {
		return err
	}
	// 调用仓储层执行数据库更新（带乐观锁，确保并发安全）
	return uc.refreshSessionRepository.RevokeRefreshSession(ctx, command.SessionID, command.RevokedAt, oldVersion)
}
