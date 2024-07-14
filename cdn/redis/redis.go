package redis

import (
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Redis server address
		Password: "",               // No password set
		DB:       0,                // Use default DB
	})
)

func Client() *redis.Client {
	if redisClient == nil {
		log.Fatal("Failed to initialize redis")
	}
	return redisClient
}
