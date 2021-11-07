package limiter

import (
	"math/bits"
	"time"
)

type LogLimiter struct {
	size       uint32
	tokens     []int64
	head, tail uint32
	msk        uint32
}

func NewLogLimiter(plainSize uint32) *LogLimiter {
	size := plainSize
	if bits.OnesCount32(size) != 1 {
		size = 1 << uint32(32-bits.LeadingZeros32(size))
	}

	return &LogLimiter{
		size:   size,
		tokens: make([]int64, size),
		msk:    size - 1,
	}
}

func (l *LogLimiter) Limit(nowNanos int64) bool {
	if l.acquired() < l.size {
		l.add(nowNanos)
		return true
	}

	if nowNanos-l.tokens[l.head&l.msk] > int64(time.Second) {
		l.add(nowNanos)
		l.head++
		return true
	}
	return false
}

func (l *LogLimiter) add(nowNanos int64) {
	l.tokens[l.tail&l.msk] = nowNanos
	l.tail++
}

func (l *LogLimiter) acquired() uint32 {
	return l.tail - l.head
}
