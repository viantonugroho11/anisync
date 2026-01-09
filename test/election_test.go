package test

import (
	"anisync"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaderElection(t *testing.T) {
	rdb := newRedis(t)
	ctx := context.Background()

	leader1, err := anisync.ElectLeader(ctx, rdb, "service-a")
	assert.NoError(t, err)
	assert.NotNil(t, leader1)

	_, err = anisync.ElectLeader(ctx, rdb, "service-a")
	assert.Error(t, err)

	_ = leader1.Release(ctx)
}
