package redisbackend

import (
	"context"
	"time"

	"anisync"
	opt "anisync/options"

	"github.com/redis/go-redis/v9"
)

// Backend wraps a redis.UniversalClient and implements anisync.Backend.
// It provides blocking and non-blocking lock acquisition using Redis SET NX + TTL,
// applies backend-specific defaults (TTL, retry delay), and delegates auto-renewal
// to the per-lock ticker. Use Acquire for blocking attempts with retry/backoff and
// TryAcquire for a single non-blocking attempt.
type Backend struct {
	rdb redis.UniversalClient
}

func New(rdb redis.UniversalClient) *Backend {
	return &Backend{rdb: rdb}
}

func defaultOptions() *opt.Options {
	return &opt.Options{
		TTL:        30 * time.Second,
		RetryDelay: 200 * time.Millisecond,
	}
}

func (b *Backend) Acquire(ctx context.Context, key string, opts ...opt.Option) (anisync.Lock, error) {
	return b.acquire(ctx, key, false, opts...)
}

func (b *Backend) TryAcquire(ctx context.Context, key string, opts ...opt.Option) (anisync.Lock, error) {
	return b.acquire(ctx, key, true, opts...)
}
