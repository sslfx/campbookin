package pkg

import (
	"context"
	"fmt"

	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	defaultMaxConns        = 10
	defaultMaxConnLifeTime = time.Hour
	defaultMaxConnIdleTime = 30 * time.Minute
	defaultConnectTimeout  = 5 * time.Second
)

type Options struct {
	MaxConns        int32
	MaxConnLifeTime time.Duration
	MaxConnIdleTime time.Duration
	ConnectTimeout  time.Duration
}

// postgres://user:password@localhost:5432/dbName?sslMode=disable
func NewPool(ctx context.Context, dsn string, opts ...Options) (*pgxpool.Pool, error) {
	MAX_RETRIES := 3
	DELAY := 500 * time.Millisecond
	var newpool *pgxpool.Pool
	err := Retry(func() error {
		cfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			return fmt.Errorf("postgres.NewPool: %w", err)
		}

		opt := defaultOptions()
		if len(opts) > 0 {
			applyOptions(cfg, opts[0])
		} else {
			applyOptions(cfg, opt)
		}

		connectCtx, cancel := context.WithTimeout(ctx, defaultConnectTimeout)
		defer cancel()

		pool, err := pgxpool.NewWithConfig(connectCtx, cfg)
		if err != nil {
			fmt.Errorf("postgres.NewPool: %w", err)
		}

		if err := pool.Ping(connectCtx); err != nil {
			pool.Close()

			return fmt.Errorf("postgres.NewPool: %w", err)
		}
		newpool = pool
		return nil
	}, MAX_RETRIES, DELAY)
	if err != nil {
		return nil, fmt.Errorf("postgres.NewPool: RetryError: %w", err)
	}
	return newpool, nil
}

func MustNewPool(ctx context.Context, dsn string, opts ...Options) *pgxpool.Pool {
	pool, err := NewPool(ctx, dsn, opts...)
	if err != nil {
		panic(err)
	}

	return pool
}

func defaultOptions() Options {
	return Options{
		MaxConns:        defaultMaxConns,
		MaxConnLifeTime: defaultMaxConnLifeTime,
		MaxConnIdleTime: defaultMaxConnIdleTime,
		ConnectTimeout:  defaultConnectTimeout,
	}
}

func applyOptions(cfg *pgxpool.Config, opts Options) {
	if opts.MaxConns > 0 {
		cfg.MaxConns = opts.MaxConns
	}
	if opts.MaxConnLifeTime > 0 {
		cfg.MaxConnLifetime = opts.MaxConnLifeTime
	}
	if opts.MaxConnIdleTime > 0 {
		cfg.MaxConnIdleTime = opts.MaxConnIdleTime
	}
	if opts.ConnectTimeout > 0 {
		cfg.PingTimeout = opts.ConnectTimeout
	}
}
