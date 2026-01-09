# anisync 🌿

anisync is a lightweight, production-ready distributed synchronization library for Go. It provides safe, portable primitives for running mutually exclusive tasks across processes and hosts, with first-class support for Redis and Postgres backends, auto-renewal, leader election, and Prometheus metrics.

Built for cron jobs, background workers, and highly available microservices.

---

## Highlights

- Distributed locks with safe release (ownership-verified)
- Auto-renew (heartbeats) for long-running tasks
- Blocking and non-blocking acquisition
- Backend-agnostic leader election
- Prometheus metrics + ready-to-import Grafana dashboard
- Postgres V2 backend with monotonic fencing tokens (anti split-brain)

---

## Installation

```bash
go get github.com/yourorg/anisync
```

Requires Go 1.23+.

---

## Repository Structure

```
.
├── backends/
│   ├── redis/
│   │   ├── lock.go
│   │   ├── lua.go
│   │   └── redis.go
│   └── postgres/
│       ├── lock.go
│       ├── schema.sql
│       └── postgres.go
│
├── election/
│   └── leader.go
│
├── fencing/
│   └── token.go
│
├── metrics/
│   └── prometheus.go
│
├── options/
│   └── options.go
│
├── internal/
│   └── retry.go
│
├── lock.go                      # public interfaces (Lock, FencedLock, Backend)
├── dashboards/
│   └── anisync-grafana.json     # Grafana dashboard for anisync metrics
├── examples/
│   └── postgres/
│       └── main.go              # Postgres backend example
├── errors.go
├── test/
│   ├── election_test.go
│   └── lock_test.go
├── go.mod
└── README.md
```

---

## Quick Start

### Redis backend

```go
be := redisbackend.New(rdb) // rdb: github.com/redis/go-redis/v9 UniversalClient
lock, err := be.Acquire(ctx, "job-daily",
    options.WithTTL(15*time.Second),
    options.WithAutoRenew(),
)
if err != nil { /* handle */ }
defer lock.Release(ctx)
```

Leader election (backend-agnostic API):

```go
be := redisbackend.New(rdb)
leader, err := election.ElectLeader(ctx, be, "service-a")
if err != nil {
    // follower path
    return
}
defer leader.Release(ctx)
```

### Postgres backend (with fencing tokens)

The Postgres backend generates a monotonically increasing fencing token on every successful acquire. Downstream consumers must reject operations with a token smaller than the last observed token, preventing split-brain side effects.

```go
pool, _ := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
pg := postgres.New(pool)
if err := pg.EnsureSchema(ctx); err != nil { /* handle */ }

lock, err := pg.Acquire(ctx, "report-job",
    options.WithTTL(15*time.Second),
    options.WithAutoRenew(),
)
if err != nil { /* handle */ }
defer lock.Release(ctx)

// Optional: read fencing token if the backend provides it
if fl, ok := lock.(interface{ Token() int64 }); ok {
    _ = fl.Token() // propagate to downstream
}
```

Schema (created automatically by `EnsureSchema`):
- Table `anisync_locks(key text primary key, value text, expires_at timestamptz, token bigint)`
- Sequence `anisync_fencing_token_seq`

---

## API Overview

Public interfaces (simplified):

```go
type Lock interface {
    Release(ctx context.Context) error
}

type FencedLock interface {
    Lock
    Token() int64
}

type Backend interface {
    Acquire(ctx context.Context, key string, opts ...options.Option) (Lock, error)
    TryAcquire(ctx context.Context, key string, opts ...options.Option) (Lock, error)
}
```

Options:
- `WithTTL(d time.Duration)`: expiration for the lock
- `WithAutoRenew()`: keep-alive using heartbeats (interval ≈ TTL/2)
- `WithRetry(d time.Duration)`: backoff between attempts for blocking Acquire

Errors:
- `anisync.ErrLockAlreadyHeld`
- `anisync.ErrLockNotHeld`
- `anisync.ErrAcquireTimeout`

---

## Semantics & Guarantees

- Mutual exclusion: A lock is acquired if and only if no other valid (non-expired) owner exists.
- Ownership-verified release:
  - Redis: release via Lua script only deletes if key value matches the owner’s token.
  - Postgres: `DELETE ... WHERE key = $1 AND value = $2`.
- TTL and auto-renew:
  - When enabled, a background ticker refreshes expiration roughly every TTL/2.
  - If the process crashes or the ticker stops, the lock naturally expires.
- Fencing tokens (Postgres only):
  - Each successful acquisition receives a strictly increasing token.
  - Downstream services must store and reject any operation with an older token.
- Time sources:
  - Redis: expiration is managed by Redis itself.
  - Postgres: expiration checks use `NOW()` (database server clock).

---

## Observability (Prometheus + Grafana)

Metrics exposed (default Prometheus registry):
- `anisync_lock_acquire_total{success="true|false"}`
- `anisync_locks_held`

Import the Grafana dashboard from `dashboards/anisync-grafana.json`.

Prometheus scrape example:

```yaml
scrape_configs:
  - job_name: anisync
    static_configs:
      - targets: ["localhost:2112"]  # your metrics endpoint
```

Tip: expose metrics via:

```go
http.Handle("/metrics", promhttp.Handler())
```

---

## Testing

```bash
go mod download
go test ./... -v
```

Tests use `miniredis` for an in-memory Redis-compatible server.

---

## Performance Notes

- Redis backend: O(1) per operation; suited for high-throughput distributed locks.
- Postgres backend: single-row upsert with conditional update on expiration; ensure adequate indexing and connection pooling.

---

## Security & Operational Considerations

- Use dedicated Redis/DB namespaces/keys for isolation.
- Ensure `EnsureSchema` is called on Postgres before acquiring locks.
- For fencing tokens, downstream systems must enforce token monotonicity to prevent stale owners from performing actions.

---

## Compatibility

- Go: 1.23+
- Redis client: `github.com/redis/go-redis/v9`
- Postgres client: `github.com/jackc/pgx/v5`

