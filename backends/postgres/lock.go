package postgres

import (
	"context"
	"time"

	"anisync"
	"anisync/metrics"
)

type Lock struct {
	key   string
	value string
	ttl   time.Duration
	pg    *Backend
	stop  chan struct{}
	token int64
}

func (l *Lock) Token() int64 { return l.token }

func (l *Lock) autoRenew() {
	ticker := time.NewTicker(l.ttl / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			_ = l.renew(context.Background())
		case <-l.stop:
			return
		}
	}
}

func (l *Lock) renew(ctx context.Context) error {
	_, err := l.pg.pool.Exec(ctx, `
UPDATE anisync_locks
SET expires_at = $3
WHERE key = $1 AND value = $2
`, l.key, l.value, time.Now().Add(l.ttl))
	return err
}

func (l *Lock) Release(ctx context.Context) error {
	close(l.stop)
	ct, err := l.pg.pool.Exec(ctx, `
DELETE FROM anisync_locks
WHERE key = $1 AND value = $2
`, l.key, l.value)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		metrics.RecordRelease(false)
		return anisync.ErrLockNotHeld
	}
	metrics.RecordRelease(true)
	return nil
}


