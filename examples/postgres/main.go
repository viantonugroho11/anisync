package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"anisync/backends/postgres"
	"anisync/options"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("set DATABASE_URL first")
		return
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	pg := postgres.New(pool)
	if err := pg.EnsureSchema(ctx); err != nil {
		panic(err)
	}

	lock, err := pg.Acquire(ctx, "daily-job",
		options.WithTTL(15*time.Second),
		options.WithAutoRenew(),
	)
	if err != nil {
		fmt.Println("failed to acquire:", err)
		return
	}
	defer lock.Release(ctx)

	// Read fencing token if the backend exposes it (used for rejecting stale operations downstream).
	if fl, ok := lock.(interface{ Token() int64 }); ok {
		fmt.Println("acquired with fencing token:", fl.Token())
	}
	// do protected work here
	time.Sleep(2 * time.Second)
}
