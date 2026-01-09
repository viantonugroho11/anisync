package fencing

// Token represents a monotonically increasing fencing token issued on each
// successful lock acquisition. Downstream systems should persist the last
// observed token and reject operations bearing a smaller token to prevent
// stale owners from acting (anti split-brain).
type Token int64

func (t Token) Int64() int64 { return int64(t) }
