package redisbackend

import (
	"context"
	"time"

	"anisync"
	"anisync/metrics"
	opt "anisync/options"

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

func (b *Backend) acquire(ctx context.Context, key string, once bool, opts ...opt.Option) (*Lock, error) {
	o := defaultOptions()
	for _, fn := range opts {
		fn(o)
	}
	lock := &Lock{
		key:   "anisync:lock:" + key,
		value: uuid.NewString(),
		ttl:   o.TTL,
		rdb:   b.rdb,
		stop:  make(chan struct{}),
	}

	for {
		ok, err := b.rdb.SetNX(ctx, lock.key, lock.value, lock.ttl).Result()
		if err != nil {
			return nil, err
		}
		if ok {
			metrics.RecordAcquire(true)
			if o.Renew {
				go lock.autoRenew()
			}
			return lock, nil
		}

		metrics.RecordAcquire(false)

		if once {
			return nil, anisync.ErrLockAlreadyHeld
		}
		select {
		case <-ctx.Done():
			return nil, anisync.ErrAcquireTimeout
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
