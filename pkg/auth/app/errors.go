package app

import "errors"

var (
	// UseCase 结构与依赖注入相关错误
	ErrInvalidUseCase           = errors.New("invalid use case")
	ErrInvalidUseCaseDependency = errors.New("invalid use case dependency")

	// 应用层令牌签发结果相关错误
	ErrInvalidAccessToken    = errors.New("invalid access token")
	ErrInvalidAccessTokenTTL = errors.New("invalid access token ttl")

	// 应用层兜底错误
	ErrAuthInternal = errors.New("auth internal error")
)
