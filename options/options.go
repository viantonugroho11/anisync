package options

import "time"

type Options struct {
	TTL        time.Duration
	RetryDelay time.Duration
	Renew      bool
}

type Option func(*Options)

// WithTTL sets the TTL for the lock.
func WithTTL(ttl time.Duration) Option {
	return func(o *Options) { o.TTL = ttl }
}

// WithRetry sets the retry delay for the lock.
func WithRetry(delay time.Duration) Option {
	return func(o *Options) { o.RetryDelay = delay }
}

// WithAutoRenew sets the auto-renew option for the lock.
func WithAutoRenew() Option {
	return func(o *Options) { o.Renew = true }
}
