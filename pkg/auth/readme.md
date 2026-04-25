# auth 模块定位

auth 模块遵循 Clean Architecture
模块只提供领域模型与 UseCase
不内置 HTTP Gin gRPC 等传输协议
上层业务项目可按自身技术栈完成适配

## 双 token 流程

登录阶段
- 上层生成 `session_id` 与明文 `refresh_token`
- `LoginUseCase` 存储 `refresh_token` 哈希并创建会话
- `LoginUseCase` 通过 `AccessTokenProvider` 签发 `access_token`

鉴权阶段
- 资源接口仅校验 `access_token`
- 不依赖数据库查询会话状态

刷新阶段
- 上层携带 `session_id` 与旧 `refresh_token`
- `RefreshTokenUseCase` 校验并轮换 refresh 会话
- `RefreshTokenUseCase` 再次通过 `AccessTokenProvider` 签发新 `access_token`

登出阶段
- `LogoutUseCase` 撤销 refresh 会话
- 同一会话重复登出保持幂等

## 端口说明

- `RefreshTokenHasher` 负责 refresh token 哈希
- `AccessTokenProvider` 负责 access token 签发与解析
- 以上端口由上层项目注入具体实现

## 错误映射说明

- `api.ResolveError(err)` 提供协议无关的统一错误映射
- `ErrorDescriptor` 包含 `Code` `Message` `Retryable`
- HTTP 适配层可将 `Code` 映射为状态码
- gRPC 适配层可将 `Code` 映射为 `codes.Code`
- 详细说明见 `pkg/auth/api/readme.md`

## 架构文档

- 分层边界 依赖方向 与时序图见 `pkg/auth/docs/architecture.md`
