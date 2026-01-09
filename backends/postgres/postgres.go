package postgres

import (
	"context"
	"time"

	"anisync"
	"anisync/metrics"
	opt "anisync/options"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const schemaSQL = `
CREATE SEQUENCE IF NOT EXISTS anisync_fencing_token_seq;

CREATE TABLE IF NOT EXISTS anisync_locks (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	expires_at TIMESTAMPTZ NOT NULL,
	token BIGINT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_anisync_locks_expires_at ON anisync_locks (expires_at);
`

type Backend struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Backend {
	return &Backend{pool: pool}
}

func defaultOptions() *opt.Options {
	return &opt.Options{
		TTL:        30 * time.Second,
		RetryDelay: 200 * time.Millisecond,
	}
}

func (b *Backend) EnsureSchema(ctx context.Context) error {
	_, err := b.pool.Exec(ctx, schemaSQL)
	return err
}

func (b *Backend) Acquire(ctx context.Context, key string, opts ...opt.Option) (anisync.Lock, error) {
	return b.acquire(ctx, key, false, opts...)
}

func (b *Backend) TryAcquire(ctx context.Context, key string, opts ...opt.Option) (anisync.Lock, error) {
	return b.acquire(ctx, key, true, opts...)
}

func (b *Backend) acquire(ctx context.Context, key string, once bool, opts ...opt.Option) (*Lock, error) {
	o := defaultOptions()
	for _, fn := range opts {
		fn(o)
	}
	l := &Lock{
		key:   "anisync:lock:" + key,
		value: uuid.NewString(),
		ttl:   o.TTL,
		pg:    b,
		stop:  make(chan struct{}),
	}
	for {
		token, acquired, err := b.tryAcquireOnce(ctx, l.key, l.value, time.Now().Add(l.ttl))
		if err != nil {
			return nil, err
		}
		if acquired {
			l.token = token
			metrics.RecordAcquire(true)
			if o.Renew {
				go l.autoRenew()
			}
			return l, nil
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

func (b *Backend) tryAcquireOnce(ctx context.Context, key, value string, expiresAt time.Time) (int64, bool, error) {
	var token int64
	err := b.pool.QueryRow(ctx, `
WITH new_token AS (
  SELECT nextval('anisync_fencing_token_seq') AS token
)
INSERT INTO anisync_locks(key, value, expires_at, token)
SELECT $1, $2, $3, new_token.token FROM new_token
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    expires_at = EXCLUDED.expires_at,
    token = EXCLUDED.token
WHERE anisync_locks.expires_at <= NOW()
RETURNING token
`, key, value, expiresAt).Scan(&token)
	if err == nil {
		return token, true, nil
	}
	if err == pgx.ErrNoRows {
		return 0, false, nil
	}
	return 0, false, err
}


