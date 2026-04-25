package app

import "time"

type AccessTokenProvider interface {
	GenerateAccessToken(userID string, expiry time.Duration) (string, error)
	ParseAccessToken(token string) (string, error)
}
