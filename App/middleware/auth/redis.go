package auth

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

// NewRedisStore 创建基于redis存储实例
func NewRedisStore(options *redis.Options, prefix string) *RedisStore {
	cli := redis.NewClient(options)
	return &RedisStore{
		cli:    cli,
		prefix: prefix,
	}
}

// RedisStore redis存储
type RedisStore struct {
	cli    *redis.Client
	prefix string
}

func (a *RedisStore) wrapperKey(key string) string {
	return fmt.Sprintf("%s%s", a.prefix, key)
}

// Set ...
func (a *RedisStore) Set(tokenString string, expiration time.Duration) error {
	cmd := a.cli.Set(a.wrapperKey(tokenString), "1", expiration)
	return cmd.Err()
}

// Check ...
func (a *RedisStore) Check(tokenString string) (bool, error) {
	cmd := a.cli.Exists(a.wrapperKey(tokenString))
	if err := cmd.Err(); err != nil {
		return false, err
	}
	return cmd.Val() > 0, nil
}

// Remove ...
func (a *RedisStore) Remove(tokenString string) error {
	cmd := a.cli.Del(tokenString)
	return cmd.Err()
}

// Close ...
func (a *RedisStore) Close() error {
	return a.cli.Close()
}
