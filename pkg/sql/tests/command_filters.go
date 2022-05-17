package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/testutils/storageutils"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type CommandFilters struct {
	syncutil.RWMutex
	filters []struct {
		id         int
		idempotent bool
		filter     kvserverbase.ReplicaCommandFilter
	}
	nextID int

	numFiltersTrackingReplays int
	replayProtection          kvserverbase.ReplicaCommandFilter
}

func (c *CommandFilters) RunFilters(args kvserverbase.FilterArgs) *roachpb.Error {
	__antithesis_instrumentation__.Notify(628311)
	c.RLock()
	defer c.RUnlock()

	if c.replayProtection != nil {
		__antithesis_instrumentation__.Notify(628313)
		return c.replayProtection(args)
	} else {
		__antithesis_instrumentation__.Notify(628314)
	}
	__antithesis_instrumentation__.Notify(628312)
	return c.runFiltersInternal(args)
}

func (c *CommandFilters) runFiltersInternal(args kvserverbase.FilterArgs) *roachpb.Error {
	__antithesis_instrumentation__.Notify(628315)
	for _, f := range c.filters {
		__antithesis_instrumentation__.Notify(628317)
		if pErr := f.filter(args); pErr != nil {
			__antithesis_instrumentation__.Notify(628318)
			return pErr
		} else {
			__antithesis_instrumentation__.Notify(628319)
		}
	}
	__antithesis_instrumentation__.Notify(628316)
	return nil
}

func (c *CommandFilters) AppendFilter(
	filter kvserverbase.ReplicaCommandFilter, idempotent bool,
) func() {
	__antithesis_instrumentation__.Notify(628320)

	c.Lock()
	defer c.Unlock()
	id := c.nextID
	c.nextID++
	c.filters = append(c.filters, struct {
		id         int
		idempotent bool
		filter     kvserverbase.ReplicaCommandFilter
	}{id, idempotent, filter})

	if !idempotent {
		__antithesis_instrumentation__.Notify(628322)
		if c.numFiltersTrackingReplays == 0 {
			__antithesis_instrumentation__.Notify(628324)
			c.replayProtection =
				storageutils.WrapFilterForReplayProtection(c.runFiltersInternal)
		} else {
			__antithesis_instrumentation__.Notify(628325)
		}
		__antithesis_instrumentation__.Notify(628323)
		c.numFiltersTrackingReplays++
	} else {
		__antithesis_instrumentation__.Notify(628326)
	}
	__antithesis_instrumentation__.Notify(628321)

	return func() {
		__antithesis_instrumentation__.Notify(628327)
		c.removeFilter(id)
	}
}

func (c *CommandFilters) removeFilter(id int) {
	__antithesis_instrumentation__.Notify(628328)
	c.Lock()
	defer c.Unlock()
	for i, f := range c.filters {
		__antithesis_instrumentation__.Notify(628330)
		if f.id == id {
			__antithesis_instrumentation__.Notify(628331)
			if !f.idempotent {
				__antithesis_instrumentation__.Notify(628333)
				c.numFiltersTrackingReplays--
				if c.numFiltersTrackingReplays == 0 {
					__antithesis_instrumentation__.Notify(628334)
					c.replayProtection = nil
				} else {
					__antithesis_instrumentation__.Notify(628335)
				}
			} else {
				__antithesis_instrumentation__.Notify(628336)
			}
			__antithesis_instrumentation__.Notify(628332)
			c.filters = append(c.filters[:i], c.filters[i+1:]...)
			return
		} else {
			__antithesis_instrumentation__.Notify(628337)
		}
	}
	__antithesis_instrumentation__.Notify(628329)
	panic(errors.AssertionFailedf("failed to find filter with id: %d.", id))
}
