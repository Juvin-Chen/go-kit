# auth 架构与流程

## 分层边界

- `domain` 只建模 refresh session 及其业务规则
- `app` 负责 UseCase 编排与端口调用
- `interfaces` 负责错误码与外部协议表达
- `infrastructure` 提供仓储 哈希 签发器等具体实现

## 依赖方向

```text
interfaces  -> app -> domain
infrastructure -> app
infrastructure -> domain

domain 不依赖 app interfaces infrastructure
app 不依赖 interfaces infrastructure
```

## 为什么 domain 不放 access token

- access token 是对外访问凭证 不是 refresh session 聚合内状态
- domain 只关注会话事实 是否过期 是否撤销 版本是否冲突
- access token 生成策略可变 JWT PASETO 自定义 opaque token
- 将其放在 app 端口可保持 domain 稳定并便于替换实现

## Login 时序

```text
上层适配层
  -> 生成 session_id refresh_token
  -> 调用 LoginUseCase

LoginUseCase
  -> RefreshTokenHasher.HashRefreshToken
  -> domain.NewRefreshSession
  -> RefreshSessionRepository.CreateRefreshSession
  -> AccessTokenIssuer.IssueAccessToken
  -> 返回 refresh session + access token
```

## Refresh 时序

```text
上层适配层
  -> 传入 session_id current_refresh_token new_refresh_token
  -> 调用 RefreshTokenUseCase

RefreshTokenUseCase
  -> Repository.GetRefreshSessionBySessionID
  -> session.EnsureActive
  -> RefreshTokenHasher.HashRefreshToken(current)
  -> session.VerifyTokenHash
  -> RefreshTokenHasher.HashRefreshToken(new)
  -> session.Rotate
  -> Repository.UpdateRefreshSessionOnRotate
  -> AccessTokenIssuer.IssueAccessToken
  -> 返回 rotated refresh session + new access token
```

## 错误映射入口

- 协议适配层统一调用 `interfaces.ResolveError(err)`
- HTTP 或 gRPC 仅负责把 `ErrorCode` 映射到对应协议状态码
