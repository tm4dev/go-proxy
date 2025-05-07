package auth

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/vlourme/go-proxy/internal/config"
)

var once sync.Once
var client *redis.Client

func GetRedisClient() *redis.Client {
	once.Do(func() {
		opt, err := redis.ParseURL(config.Get().Auth.Redis.DSN)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to parse Redis URL")
		}
		client = redis.NewClient(opt)
		_, err = client.Ping(context.Background()).Result()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to ping Redis")
		}
	})

	return client
}
