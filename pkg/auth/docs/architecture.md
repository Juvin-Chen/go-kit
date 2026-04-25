# Auth 架构设计

## 1. 设计目标

- 协议无关：`auth` 核心不依赖 HTTP/gRPC 框架
- 分层稳定：业务规则与技术实现分离，支持替换 infra 实现
- 一致性可控：refresh session 作为状态事实源，支持幂等与乐观锁
- 安全可演进：Access Token 与 Refresh Token 分离治理

## 2. 分层模型

### 2.1 目录结构

```text
pkg/auth
├── domain   // 领域模型与规则
├── app      // 用例编排与端口定义
├── infra    // mysql / jwt / hasher 实现
└── api      // 错误目录与适配层入口
```

### 2.2 依赖方向

```text
api   -> app -> domain
infra -> app
infra -> domain

domain 不依赖 app/api/infra
app 不依赖 api/infra
```

## 3. 关键设计决策

### 3.1 为什么 Access Token 不放入 domain

- Access Token 是外部访问凭证，不是 `RefreshSession` 聚合事实
- 领域层只关心会话生命周期：创建、轮换、撤销、过期、版本冲突
- Token 形态可替换（JWT/PASETO/Opaque），应通过 `app.AccessTokenProvider` 解耦

### 3.2 为什么 Refresh Token 只存哈希

- 数据库不存明文 refresh token，降低泄露风险
- `infra/hasher` 统一使用 HMAC-SHA256 输出固定长度摘要

### 3.3 为什么引入乐观锁

- `refresh_sessions.version` 作为并发控制字段
- 轮换和撤销都要求 `WHERE version = ?`
- 冲突时返回领域错误 `ErrRefreshSessionConflict`，交由上层决定重试策略

## 4. 核心时序

### 4.1 Login

```text
Adapter
  -> 生成 session_id 与明文 refresh_token
  -> 调用 LoginUseCase.Login

LoginUseCase
  -> 参数与依赖校验
  -> 生成 access token
  -> 哈希 refresh token
  -> domain.NewRefreshSession
  -> Repository.CreateRefreshSession
  -> 返回 LoginResult
```

关键语义：
- 先签发 access token，再写 session，避免落库成功但无法返回令牌
- `issued_at` 与 `refresh/access` 过期时间由服务端统一生成，避免客户端篡改

### 4.2 Refresh

```text
Adapter
  -> 传入 session_id/current_token/new_refresh_token/new_refresh_expiry
  -> 调用 RefreshTokenUseCase.Refresh

RefreshTokenUseCase
  -> 查询会话
  -> EnsureActive
  -> 先签发新 access token
  -> 校验 current token hash
  -> 轮换 new refresh token hash
  -> Repository.UpdateRefreshSessionOnRotate(version)
  -> 冲突场景有限重试
  -> 返回 RefreshTokenResult
```

关键语义：
- 先签发 access token，后做持久化，降低会话与令牌不一致窗口
- 乐观锁冲突最多重试固定次数，避免无限重试

### 4.3 Logout

```text
LogoutUseCase
  -> 查询会话
  -> 已撤销直接成功（幂等）
  -> session.Revoke
  -> Repository.RevokeRefreshSession(version + revoked_at is null)
```

## 5. 错误模型

- 领域和应用内部错误在 `api.ResolveError` 统一收敛
- 输出标准结构：`ErrorDescriptor{Code, Message, Retryable}`
- 协议层只做映射，不承载业务判断

建议映射：
- `AUTH_0400_INVALID_ARGUMENT` -> HTTP 400
- `AUTH_0401_INVALID_TOKEN` -> HTTP 401
- `AUTH_0404_SESSION_NOT_FOUND` -> HTTP 404
- `AUTH_0409_SESSION_CONFLICT` -> HTTP 409
- `AUTH_0500_INTERNAL_ERROR` -> HTTP 500

## 6. 生产落地建议

- 所有时间统一 UTC，数据库字段使用 `datetime(3)`
- JWT 密钥与 refresh hasher 密钥分离并轮换
- 接口层对冲突错误实施指数退避重试
- 维护过期会话清理任务，定时调用 `DeleteExpiredRefreshSessions`
- 监控关键指标：登录成功率、刷新冲突率、撤销命中率、清理耗时

## 7. 扩展点

- 多端登录策略：按 `device_id` 维度扩展 session 唯一约束
- Token Family：增强 refresh token 重放检测
- 风控增强：在 `app` 层增加设备指纹与风险评分校验
