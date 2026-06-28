package cache

import (
	"context"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

func Connect() error {
	client = redis.NewClient(&redis.Options{
		Addr: os.Getenv("REDIS_URL"),
	})

	_, err := client.Ping(context.Background()).Result()
	return err
}

func Set(key string, value []byte, ttl time.Duration) error {
	return client.Set(context.Background(), key, value, ttl).Err()
}

func Get(key string) ([]byte, error) {
	return client.Get(context.Background(), key).Bytes()
}

func Del(keys ...string) error {
	return client.Del(context.Background(), keys...).Err()
}
