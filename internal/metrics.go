package internal

import (
	"math"
	"sync/atomic"
)

// Counter is an in-memory implementation of a Counter.
type Counter struct {
	bits uint64
}

// Add implements Counter.
func (c *Counter) Add(delta float64) {
	for {
		var (
			old  = atomic.LoadUint64(&c.bits)
			newf = math.Float64frombits(old) + delta
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&c.bits, old, new) {
			break
		}
	}
}

// Value returns the current value of the counter.
func (c *Counter) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.bits))
}

// ValueReset returns the current value of the counter, and resets it to zero.
// This is useful for metrics backends whose counter aggregations expect deltas,
// like Graphite.
func (c *Counter) ValueReset() float64 {
	for {
		var (
			old  = atomic.LoadUint64(&c.bits)
			newf = 0.0
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&c.bits, old, new) {
			return math.Float64frombits(old)
		}
	}
}
