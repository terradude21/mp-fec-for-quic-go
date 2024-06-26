package utils

import (
	"math"
	"time"

	"github.com/quic-go/quic-go/internal/protocol"
)

// InfDuration is a duration of infinite length
const InfDuration = time.Duration(math.MaxInt64)

// MinNonZeroDuration return the minimum duration that's not zero.
func MinNonZeroDuration(a, b time.Duration) time.Duration {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	return min(a, b)
}

// MinTime returns the earlier time
func MinTime(a, b time.Time) time.Time {
	if a.After(b) {
		return b
	}
	return a
}

// MaxTime returns the later time
func MaxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

// MinByteCount returns the minimum of two ByteCounts
func MinByteCount(a, b protocol.ByteCount) protocol.ByteCount {
	if a < b {
		return a
	}
	return b
}

// MaxByteCount returns the maximum of two ByteCounts
func MaxByteCount(a, b protocol.ByteCount) protocol.ByteCount {
	if a < b {
		return b
	}
	return a
}
