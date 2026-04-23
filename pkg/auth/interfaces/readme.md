# 前置说明

auth 模块中遵循 Clean Architecture 原则，将业务逻辑（UseCase）与传输层（HTTP/Gin）完全解耦。


auth 包仅提供纯业务能力，由上层项目决定使用 HTTP、gRPC 或其他协议进行适配，从而保证模块的可复用性和可测试性。

---

# interfaces 层说明

## 定位

`interfaces` 层用于承接 `app` 和 `domain` 向外暴露时的适配需求  
该层不实现业务规则 仅负责对外协议可消费的表达

## 错误目录

`error_catalog.go` 提供统一错误映射函数 `ResolveError(err)`  
输出结构为 `ErrorDescriptor`

字段说明
- `Code` 业务错误码 供前端与调用方稳定识别
- `Message` 错误消息 默认使用模块内部错误消息
- `Retryable` 是否建议重试 供客户端退避策略使用

## 接入方式

HTTP 适配层
- 调用 `ResolveError(err)` 获取 `ErrorDescriptor`
- 将 `Code` 映射为 HTTP Status
- 将 `Code` `Message` 写入标准响应体

gRPC 适配层
- 调用 `ResolveError(err)` 获取 `ErrorDescriptor`
- 将 `Code` 映射为 gRPC `codes.Code`
- 将 `Message` 和 `Code` 附加到 `status` 或 metadata

## 设计约束

- `interfaces` 可依赖 `app` 与 `domain`
- `app` 与 `domain` 不反向依赖 `interfaces`
- 错误码一旦对外发布需保持兼容 避免随意变更

## 关联文档

- 分层与时序说明见 `pkg/auth/docs/architecture.md`
