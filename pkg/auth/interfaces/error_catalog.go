package interfaces

import (
	"errors"

	"github.com/Juvin-Chen/go-kit/pkg/auth/app"
	"github.com/Juvin-Chen/go-kit/pkg/auth/domain"
)

// ErrorCode 是协议无关的业务错误码
// 上层可将其映射为 HTTP 状态码或 gRPC 状态码
type ErrorCode string

const (
	// ErrorCodeOK 表示成功
	ErrorCodeOK ErrorCode = "AUTH_0000"

	// ErrorCodeInvalidArgument 表示请求参数或上下文不合法
	ErrorCodeInvalidArgument ErrorCode = "AUTH_0400_INVALID_ARGUMENT"

	// ErrorCodeInvalidToken 表示 access token 或 refresh token 无效
	ErrorCodeInvalidToken ErrorCode = "AUTH_0401_INVALID_TOKEN"

	// ErrorCodeSessionNotFound 表示会话不存在
	ErrorCodeSessionNotFound ErrorCode = "AUTH_0404_SESSION_NOT_FOUND"

	// ErrorCodeSessionConflict 表示乐观锁冲突
	ErrorCodeSessionConflict ErrorCode = "AUTH_0409_SESSION_CONFLICT"

	// ErrorCodeSessionRevoked 表示会话已撤销
	ErrorCodeSessionRevoked ErrorCode = "AUTH_0410_SESSION_REVOKED"

	// ErrorCodeSessionExpired 表示会话已过期
	ErrorCodeSessionExpired ErrorCode = "AUTH_0410_SESSION_EXPIRED"

	// ErrorCodeInternal 表示内部错误兜底
	ErrorCodeInternal ErrorCode = "AUTH_0500_INTERNAL_ERROR"
)

// ErrorDescriptor 是统一错误描述
// 接口层可直接使用该结构组装响应体
type ErrorDescriptor struct {
	Code      ErrorCode
	Message   string
	Retryable bool
}

// ResolveError 将模块内部错误收敛为统一错误描述
// 该函数只做语义映射 不做日志与链路处理
func ResolveError(err error) ErrorDescriptor {
	if err == nil {
		return ErrorDescriptor{
			Code:      ErrorCodeOK,
			Message:   "ok",
			Retryable: false,
		}
	}

	// 参数类错误 适配层通常映射为 400
	if errors.Is(err, app.ErrInvalidUseCase) || errors.Is(err, app.ErrInvalidUseCaseDependency) ||
		errors.Is(err, app.ErrInvalidAccessTokenTTL) || errors.Is(err, domain.ErrInvalidRefreshSession) ||
		errors.Is(err, domain.ErrInvalidRefreshSessionID) || errors.Is(err, domain.ErrInvalidRefreshSessionVersion) ||
		errors.Is(err, domain.ErrInvalidUserID) || errors.Is(err, domain.ErrInvalidRefreshTokenHash) ||
		errors.Is(err, domain.ErrInvalidRefreshTokenTTL) || errors.Is(err, domain.ErrInvalidCleanupLimit) {
		return ErrorDescriptor{
			Code:      ErrorCodeInvalidArgument,
			Message:   err.Error(),
			Retryable: false,
		}
	}

	// 令牌错误 适配层通常映射为 401
	if errors.Is(err, app.ErrInvalidAccessToken) || errors.Is(err, domain.ErrInvalidRefreshToken) {
		return ErrorDescriptor{
			Code:      ErrorCodeInvalidToken,
			Message:   err.Error(),
			Retryable: false,
		}
	}

	if errors.Is(err, domain.ErrRefreshSessionNotFound) {
		return ErrorDescriptor{
			Code:      ErrorCodeSessionNotFound,
			Message:   err.Error(),
			Retryable: false,
		}
	}

	if errors.Is(err, domain.ErrRefreshSessionConflict) {
		return ErrorDescriptor{
			Code:      ErrorCodeSessionConflict,
			Message:   err.Error(),
			Retryable: true,
		}
	}

	if errors.Is(err, domain.ErrRefreshSessionRevoked) {
		return ErrorDescriptor{
			Code:      ErrorCodeSessionRevoked,
			Message:   err.Error(),
			Retryable: false,
		}
	}

	if errors.Is(err, domain.ErrRefreshSessionExpired) {
		return ErrorDescriptor{
			Code:      ErrorCodeSessionExpired,
			Message:   err.Error(),
			Retryable: false,
		}
	}

	return ErrorDescriptor{
		Code:      ErrorCodeInternal,
		Message:   app.ErrAuthInternal.Error(),
		Retryable: false,
	}
}
