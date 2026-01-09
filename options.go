package anisync

import "time"

type Options struct {
	TTL        time.Duration
	RetryDelay time.Duration
	Renew      bool
}

type Option func(*Options)

func WithTTL(ttl time.Duration) Option {
	return func(o *Options) { o.TTL = ttl }
}

func WithRetry(delay time.Duration) Option {
	return func(o *Options) { o.RetryDelay = delay }
}

func WithAutoRenew() Option {
	return func(o *Options) { o.Renew = true }
}

func defaultOptions() *Options {
	return &Options{
		TTL:        30 * time.Second,
		RetryDelay: 200 * time.Millisecond,
	}
}
