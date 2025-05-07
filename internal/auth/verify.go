package auth

import (
	"context"

	"github.com/vlourme/go-proxy/internal/config"
)

// VerifyCredentials verifies the credentials of the user
func Verify(username, password string) bool {
	switch config.Get().Auth.Type {
	case config.AuthTypeCredentials:
		return verifyCredentials(username, password)
	case config.AuthTypeRedis:
		return verifyRedisCredentials(username, password)
	case config.AuthTypeNone:
		return true
	default:
		return false
	}
}

// verifyCredentials verifies the credentials of the user
func verifyCredentials(username, password string) bool {
	cfgUsername, cfgPassword := config.Get().Auth.Credentials.Username, config.Get().Auth.Credentials.Password

	return username == cfgUsername && password == cfgPassword
}

// verifyRedisCredentials verifies the credentials of the user using Redis
func verifyRedisCredentials(username, password string) bool {
	client := GetRedisClient()

	return client.Get(context.Background(), username).Val() == password
}
