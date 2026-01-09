package internal

import (
	"context"
	"time"
)

// SleepWithContext pauses for the specified duration or returns early if the
// context is canceled. Prefer this over time.Sleep inside retry loops to make
// the wait interruptible and responsive to shutdown signals.
func SleepWithContext(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
