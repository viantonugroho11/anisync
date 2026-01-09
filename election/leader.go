package election

import (
	"context"
	"time"

	"anisync"
	opt "anisync/options"
)

// ElectLeader attempts to acquire a leader lock using the provided Backend.
// The lock is namespaced under "leader:<name>" and uses a default TTL with
// auto-renew enabled so leadership is maintained until Release is called or
// the process stops renewing. This function is backend-agnostic (Redis, Postgres, etc.).
func ElectLeader(ctx context.Context, backend anisync.Backend, name string) (anisync.Lock, error) {
	return backend.Acquire(
		ctx,
		"leader:"+name,
		opt.WithTTL(15*time.Second),
		opt.WithAutoRenew(),
	)
}
