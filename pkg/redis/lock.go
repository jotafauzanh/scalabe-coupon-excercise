package redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	ErrLockNotAcquired = errors.New("lock not acquired")
	ErrLockNotHeld     = errors.New("lock not held by this instance")
)

type RedisLock struct {
	client *redis.Client
	key    string
	value  string
	ttl    time.Duration
}

func NewRedisLock(client *redis.Client, key string, ttl time.Duration) *RedisLock {
	return &RedisLock{
		client: client,
		key:    "lock:" + key,
		ttl:    ttl,
	}
}

func (l *RedisLock) Lock(ctx context.Context) error {
	// Generate random value for this lock instance
	valueBytes := make([]byte, 16)
	if _, err := rand.Read(valueBytes); err != nil {
		return err
	}
	l.value = hex.EncodeToString(valueBytes)

	// Try to acquire lock with SET NX EX
	result, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
	if err != nil {
		return err
	}
	if !result {
		return ErrLockNotAcquired
	}

	return nil
}

func (l *RedisLock) Unlock(ctx context.Context) error {
	if l.value == "" {
		return ErrLockNotHeld
	}

	// Use Lua script to safely unlock only if we hold the lock
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("DEL", KEYS[1])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
	if err != nil {
		return err
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}

func (l *RedisLock) Extend(ctx context.Context) error {
	if l.value == "" {
		return ErrLockNotHeld
	}

	// Use Lua script to safely extend only if we hold the lock
	script := `
		if redis.call("GET", KEYS[1]) == ARGV[1] then
			return redis.call("EXPIRE", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := l.client.Eval(ctx, script, []string{l.key}, l.value, int(l.ttl.Seconds())).Result()
	if err != nil {
		return err
	}

	if result.(int64) == 0 {
		return ErrLockNotHeld
	}

	return nil
}
