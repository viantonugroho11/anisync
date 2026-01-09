package anisync

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Leader struct {
	*Lock
}

func ElectLeader(ctx context.Context, rdb redis.UniversalClient, name string) (*Leader, error) {
	lock, err := Acquire(
		ctx,
		rdb,
		"leader:"+name,
		WithTTL(15*time.Second),
		WithAutoRenew(),
	)
	if err != nil {
		return nil, err
	}
	return &Leader{lock}, nil
}
