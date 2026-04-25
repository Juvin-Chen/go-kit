# Auth 模块说明

`pkg/auth` 是一个协议无关的认证领域模块，提供双 Token 认证的核心能力

- Access Token：短期令牌，用于资源访问鉴权
- Refresh Token：长期令牌，用于换发 Access Token
- Session：以 `refresh_sessions` 为事实源，承载撤销、过期、版本冲突控制

模块遵循 DDD + Clean Architecture，不依赖 HTTP、Gin、gRPC 等传输协议，适合在不同接口层复用

## 当前能力

- 登录：创建 refresh session，并返回 access token
- 刷新：校验并轮换 refresh token，返回新 access token
- 登出：撤销 refresh session，具备幂等语义
- 错误：统一错误收敛为稳定业务码，便于 API 层映射

## 分层职责

- `domain`：聚合与规则，维护 session 状态机
- `app`：UseCase 编排，不感知传输协议和 DB 细节
- `infra`：MySQL/JWT/Hasher 具体实现
- `api`：错误码目录与协议适配入口

详细设计见 [architecture.md](file:///d:/github_projects/go-kit/pkg/auth/docs/architecture.md)

## 核心端口

- `app.RefreshTokenHasher`
- `app.AccessTokenProvider`
- `domain.RefreshSessionRepository`

上层项目通过依赖注入提供具体实现，`app` 只依赖接口

## 快速接入

1. 初始化基础设施实现
2. 组装 UseCase
3. 在 HTTP/gRPC Handler 中调用 UseCase
4. 用 `api.ResolveError(err)` 统一映射错误码

### 示例

```go
package main

import (
	"time"

	"github.com/Juvin-Chen/go-kit/pkg/auth/app"
	"github.com/Juvin-Chen/go-kit/pkg/auth/infra/hasher"
	"github.com/Juvin-Chen/go-kit/pkg/auth/infra/jwt"
	"github.com/Juvin-Chen/go-kit/pkg/auth/infra/mysql"
	"gorm.io/gorm"
)

type AuthUseCases struct {
	LoginUseCase        *app.LoginUseCase
	RefreshTokenUseCase *app.RefreshTokenUseCase
	LogoutUseCase       *app.LogoutUseCase
}

func NewAuthUseCases(db *gorm.DB, refreshHasherSecret string, accessTokenSecret string) (*AuthUseCases, error) {
	refreshSessionRepository := mysql.NewMySQLRefreshSessionRepository(db)

	refreshTokenHasher, err := hasher.NewRefreshTokenHMACHasher(refreshHasherSecret)
	if err != nil {
		return nil, err
	}
	accessTokenProvider, err := jwt.NewJWTProvider(accessTokenSecret)
	if err != nil {
		return nil, err
	}

	const refreshTokenTTL = 7 * 24 * time.Hour
	const accessTokenTTL = 15 * time.Minute

	return &AuthUseCases{
		LoginUseCase: app.NewLoginUseCase(
			refreshSessionRepository,
			refreshTokenHasher,
			accessTokenProvider,
			refreshTokenTTL,
			accessTokenTTL,
		),
		RefreshTokenUseCase: app.NewRefreshTokenUseCase(
			refreshSessionRepository,
			refreshTokenHasher,
			accessTokenProvider,
			accessTokenTTL,
		),
		LogoutUseCase: app.NewLogoutUseCase(refreshSessionRepository),
	}, nil
}
```

## Handler 调用建议

- 登录入参：`session_id`、`user_id`、`refresh_token`
- 刷新入参：`session_id`、`current_token`、`new_refresh_token`、`new_refresh_expiry`
- 登出入参：`session_id`
- 错误处理：统一使用 `api.ResolveError(err)`，不要直接泄露内部错误

## 生产建议

- 所有时间统一使用 UTC
- AccessToken 密钥与 RefreshHasher 密钥分离管理
- 开启密钥轮换策略，并支持灰度窗口
- 对 `AUTH_0409_SESSION_CONFLICT` 做有限重试与退避
- 清理任务定时调用 `DeleteExpiredRefreshSessions`

## 参考文档

- 架构设计：[architecture.md](file:///d:/github_projects/go-kit/pkg/auth/docs/architecture.md)
- 数据库设计：[db.md](file:///d:/github_projects/go-kit/pkg/auth/docs/db.md)
- 错误映射：[api/readme.md](file:///d:/github_projects/go-kit/pkg/auth/api/readme.md)
