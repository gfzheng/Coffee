package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

const (
	expirationInMs = 604800000
)

var (
	signingKey string = "signingKey"

	ErrUnauthorized error = errors.New("Unauthorized")
)

// JWTAuth 用户认证
type JWTAuth struct {
	SigningMethod string `toml:"signing_method"`
	SigningKey    string `toml:"signing_key"`
	Expired       int    `toml:"expired"`
	Store         string `toml:"store"`
	FilePath      string `toml:"file_path"`
	RedisDB       int    `toml:"redis_db"`
	RedisPrefix   string `toml:"redis_prefix"`
}

// Init 初始化鉴权模块配置
func Init(auth *JWTAuth, store Store) {
	signingKey = auth.SigningKey
	InitStore(store)
}

// GenerateToken 根据用户 ID 创建一个新的 JWT 令牌
func GenerateToken(userId string) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.StandardClaims{
		ExpiresAt: now.Add(time.Millisecond * expirationInMs).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Subject:   userId,
	})

	tokenString, err := token.SignedString([]byte(signingKey))

	if err == nil {
		if err := store.Set("Bearer "+tokenString, time.Hour); err != nil {
			return "", err
		}
	}

	return tokenString, err
}

// DestroyToken 销毁传入的令牌，令该令牌无效。用于实现登出
func DestroyToken(token string) error {
	return store.Remove(token)
}

// Authenticate 根据 HTTP Header 中的 Authorization 解析用户信息的中间件
func Authenticate(bearerToken string) (email string, err error) {
	email = ""
	err = ErrUnauthorized
	if strings.HasPrefix(bearerToken, "Bearer ") {
		token, _ := jwt.ParseWithClaims(bearerToken[len("Bearer "):], &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrUnauthorized
			}
			return []byte(signingKey), nil
		})
		if token != nil && token.Valid {
			var exists bool
			exists, err = store.Check(bearerToken)
			if err != nil || !exists {
				return
			}

			email = token.Claims.(*jwt.StandardClaims).Subject
		}
	}
	return
}
