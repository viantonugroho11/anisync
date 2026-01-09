package anisync

import (
	opt "anisync/options"
	"context"
)

// Lock is the minimal interface representing an acquired lock that can be released.
type Lock interface {
	Release(ctx context.Context) error
}

// FencedLock is a Lock that also carries a monotonically increasing fencing token.
// Downstream systems should persist and compare tokens, rejecting operations that
// present a smaller token to prevent stale owners from acting (anti split-brain).
type FencedLock interface {
	Lock
	Token() int64
}

// Backend abstracts lock providers (e.g., Redis, Postgres). Implementations must
// provide blocking Acquire and non-blocking TryAcquire behaviors using the given options.
type Backend interface {
	Acquire(ctx context.Context, key string, opts ...opt.Option) (Lock, error)
	TryAcquire(ctx context.Context, key string, opts ...opt.Option) (Lock, error)
}
