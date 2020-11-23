package auth

import "time"

var store Store

type Store interface {
	// 放入令牌，指定到期时间
	Set(token string, expiration time.Duration) error

	Remove(token string) error
	// 检查令牌是否存在
	Check(token string) (bool, error)
	// 关闭存储
	Close() error
}

func InitStore(_store Store) {
	store = _store
}
