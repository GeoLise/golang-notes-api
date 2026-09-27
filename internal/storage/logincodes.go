package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type LoginCodes struct {
	cache *redis.Client
}

const (
	loginCodeTTL     = 5 * time.Minute
	loginCodePending = "pending"
)

func NewLoginCodes(cache *redis.Client) *LoginCodes {
	return &LoginCodes{cache: cache}
}

func loginCodeKey(code string) string {
	return "logincodes:" + code
}

func (c *LoginCodes) Generate(ctx context.Context) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	code := hex.EncodeToString(b)

	err := c.cache.Set(ctx, loginCodeKey(code), loginCodePending, loginCodeTTL).Err()
	if err != nil {
		return "", err
	}

	return code, nil
}

var ErrLoginCodeNotFound = errors.New("login code not found or expired")
var ErrLoginCodeNotAcivated = errors.New("login code is not activated")

func (c *LoginCodes) IsPending(ctx context.Context, code string) (bool, error) {
	val, err := c.cache.Get(ctx, loginCodeKey(code)).Result()
	if errors.Is(err, redis.Nil) {
		return false, ErrLoginCodeNotFound
	}
	if err != nil {
		return false, err
	}
	return val == loginCodePending, nil
}

func (c *LoginCodes) Confirm(ctx context.Context, code string, userID int) error {
	return c.cache.Set(ctx, loginCodeKey(code), strconv.Itoa(userID), loginCodeTTL).Err()
}

func (c *LoginCodes) GetUserIdByCode(ctx context.Context, code string) (int, error) {
	val, err := c.cache.Get(ctx, loginCodeKey(code)).Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrLoginCodeNotFound
	}

	if val == loginCodePending {
		return 0, ErrLoginCodeNotAcivated
	}

	c.cache.Del(ctx, loginCodeKey(code))

	return strconv.Atoi(val)
}
