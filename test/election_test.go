package test

import (
	redisbackend "anisync/backends/redis"
	"anisync/election"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLeaderElection(t *testing.T) {
	rdb := newRedis(t)
	ctx := context.Background()

	be := redisbackend.New(rdb)
	leader1, err := election.ElectLeader(ctx, be, "service-a")
	assert.NoError(t, err)
	assert.NotNil(t, leader1)

	_, err = election.ElectLeader(ctx, be, "service-a")
	assert.Error(t, err)

	_ = leader1.Release(ctx)
}
