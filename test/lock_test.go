package test

import (
	"anisync"
	redisbackend "anisync/backends/redis"
	"anisync/options"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newRedis(t *testing.T) redis.UniversalClient {
	mr := miniredis.RunT(t)
	return redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
}

func TestAcquireAndRelease(t *testing.T) {
	rdb := newRedis(t)
	ctx := context.Background()

	backend := redisbackend.New(rdb)
	lock, err := backend.Acquire(ctx, "test-lock")
	assert.NoError(t, err)
	assert.NotNil(t, lock)

	err = lock.Release(ctx)
	assert.NoError(t, err)
}

func TestTryAcquireFail(t *testing.T) {
	rdb := newRedis(t)
	ctx := context.Background()

	backend := redisbackend.New(rdb)
	_, err := backend.Acquire(ctx, "lock")
	assert.NoError(t, err)

	_, err = backend.TryAcquire(ctx, "lock")
	assert.ErrorIs(t, err, anisync.ErrLockAlreadyHeld)
}

func TestAutoExpire(t *testing.T) {
	rdb := newRedis(t)
	ctx := context.Background()

	backend := redisbackend.New(rdb)
	_, err := backend.Acquire(ctx, "expire-lock", options.WithTTL(100*time.Millisecond))
	assert.NoError(t, err)

	time.Sleep(200 * time.Millisecond)

	lock, err := backend.TryAcquire(ctx, "expire-lock")
	assert.NoError(t, err)

	_ = lock.Release(ctx)
}
