package app

// 其实这个接口放 app / domain 都可以
type RefreshTokenHasher interface {
	HashRefreshToken(plainToken string) (string, error)
}
