package redisbackend

import (
	"context"

	"anisync"
	"anisync/metrics"

	"github.com/redis/go-redis/v9"
)

var releaseScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

func (l *Lock) Release(ctx context.Context) error {
	close(l.stop)
	res, err := releaseScript.Run(ctx, l.rdb, []string{l.key}, l.value).Int()
	if err != nil {
		return err
	}
	if res == 0 {
		metrics.RecordRelease(false)
		return anisync.ErrLockNotHeld
	}
	metrics.RecordRelease(true)
	return nil
}
