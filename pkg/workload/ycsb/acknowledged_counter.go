package ycsb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

const (
	windowSize = 1 << 20

	windowMask = windowSize - 1
)

type AcknowledgedCounter struct {
	mu struct {
		syncutil.Mutex
		count  uint64
		window []bool
	}
}

func NewAcknowledgedCounter(initialCount uint64) *AcknowledgedCounter {
	__antithesis_instrumentation__.Notify(699290)
	c := &AcknowledgedCounter{}
	c.mu.count = initialCount
	c.mu.window = make([]bool, windowSize)
	return c
}

func (c *AcknowledgedCounter) Last() uint64 {
	__antithesis_instrumentation__.Notify(699291)
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.mu.count
}

func (c *AcknowledgedCounter) Acknowledge(v uint64) (uint64, error) {
	__antithesis_instrumentation__.Notify(699292)
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.mu.window[v&windowMask] {
		__antithesis_instrumentation__.Notify(699295)
		return 0, errors.Errorf("Number of pending acknowledgements exceeded window size: %d has been acknowledged, but %d is not acknowledged", v, c.mu.count)
	} else {
		__antithesis_instrumentation__.Notify(699296)
	}
	__antithesis_instrumentation__.Notify(699293)

	c.mu.window[v&windowMask] = true
	count := uint64(0)
	for c.mu.window[c.mu.count&windowMask] {
		__antithesis_instrumentation__.Notify(699297)
		c.mu.window[c.mu.count&windowMask] = false
		c.mu.count++
		count++
	}
	__antithesis_instrumentation__.Notify(699294)
	return count, nil
}
