package app

type RefreshTokenHasher interface {
	HashRefreshToken(plainToken string) (string, error)
}
