package anisync

import "errors"

var (
	ErrLockAlreadyHeld = errors.New("anisync: lock already held")
	ErrLockNotHeld     = errors.New("anisync: lock not held")
	ErrAcquireTimeout  = errors.New("anisync: acquire timeout")
)
