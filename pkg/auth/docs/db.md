# 数据库设计（auth）

## 表总览

| 表名 | 说明 |
| --- | --- |
| `refresh_sessions` | 刷新令牌会话表，存储 refresh token 哈希、有效期、撤销状态与版本号（乐观锁） |

## `refresh_sessions`

### 字段说明

| 字段名 | 类型 | 约束 | 说明 |
| --- | --- | --- | --- |
| `id` | `BIGINT UNSIGNED` | `PRIMARY KEY AUTO_INCREMENT` | 自增主键 |
| `session_id` | `VARCHAR(64)` | `NOT NULL`, `UNIQUE` | 会话唯一标识 |
| `user_id` | `VARCHAR(64)` | `NOT NULL`, `INDEX`, `INDEX(user_id, revoked_at, expires_at)` | 用户标识 |
| `refresh_token_hash` | `VARCHAR(64)` | `NOT NULL` | refresh token 哈希值（HMAC-SHA256 hex） |
| `issued_at` | `DATETIME(3)` | `NOT NULL` | 签发时间 |
| `expires_at` | `DATETIME(3)` | `NOT NULL`, `INDEX`, `INDEX(user_id, revoked_at, expires_at)` | 过期时间 |
| `revoked_at` | `DATETIME(3)` | `NULL`, `INDEX(user_id, revoked_at, expires_at)` | 撤销时间，未撤销为 `NULL` |
| `version` | `BIGINT UNSIGNED` | `NOT NULL`, `DEFAULT 1` | 版本号（乐观锁） |
| `created_at` | `DATETIME(3)` | `NOT NULL` | 创建时间 |
| `updated_at` | `DATETIME(3)` | `NOT NULL` | 更新时间 |

### 索引说明

| 索引名 | 字段 | 用途 |
| --- | --- | --- |
| `uk_refresh_session_id` | `session_id` | 会话唯一定位 |
| `idx_refresh_user_id` | `user_id` | 按用户查询会话 |
| `idx_refresh_expires_at` | `expires_at` | 过期会话清理 |
| `idx_user_active` | `user_id, revoked_at, expires_at` | 用户活跃会话查询 |

### Revoke 幂等更新

为实现“重复撤销请求幂等”，撤销更新可增加 `revoked_at IS NULL` 条件：

```sql
UPDATE `refresh_sessions`
SET
  `revoked_at` = ?,
  `version` = ?,
  `updated_at` = CURRENT_TIMESTAMP(3)
WHERE
  `session_id` = ?
  AND `version` = ?
  AND `revoked_at` IS NULL;
```

当 `RowsAffected = 0` 时：
- 若会话已撤销：视为幂等成功
- 若未撤销但版本不匹配：返回并发冲突
