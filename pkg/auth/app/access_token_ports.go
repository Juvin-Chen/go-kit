package app

import "time"

type AccessTokenCommand struct {
	UserID    string
	SessionID string
	IssuedAt  time.Time
}

type AccessTokenResult struct {
	AccessToken string
	ExpiresAt   time.Time
}

// AccessTokenIssuer 是应用层端口
// 上层项目可注入 JWT PASETO 或自定义签发器实现
type AccessTokenIssuer interface {
	IssueAccessToken(command AccessTokenCommand) (*AccessTokenResult, error)
}
