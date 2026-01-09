package anisync

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Lock struct {
	key   string
	value string
	ttl   time.Duration
	rdb   redis.UniversalClient
	stop  chan struct{}
}

func Acquire(ctx context.Context, rdb redis.UniversalClient, key string, opts ...Option) (*Lock, error) {
	return acquire(ctx, rdb, key, false, opts...)
}

func TryAcquire(ctx context.Context, rdb redis.UniversalClient, key string, opts ...Option) (*Lock, error) {
	return acquire(ctx, rdb, key, true, opts...)
}

func acquire(ctx context.Context, rdb redis.UniversalClient, key string, once bool, opts ...Option) (*Lock, error) {
	o := defaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	lock := &Lock{
		key:   "anisync:lock:" + key,
		value: uuid.NewString(),
		ttl:   o.TTL,
		rdb:   rdb,
		stop:  make(chan struct{}),
	}

	for {
		ok, err := rdb.SetNX(ctx, lock.key, lock.value, lock.ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			observeAcquire(true)
			if o.Renew {
				go lock.autoRenew()
			}
			return lock, nil
		}

		observeAcquire(false)

		if once {
			return nil, ErrLockAlreadyHeld
		}

		select {
		case <-ctx.Done():
			return nil, ErrAcquireTimeout
		case <-time.After(o.RetryDelay):
		}
	}
}


func (l *Lock) autoRenew() {
	ticker := time.NewTicker(l.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.rdb.Expire(context.Background(), l.key, l.ttl)
		case <-l.stop:
			return
		}
	}
}
