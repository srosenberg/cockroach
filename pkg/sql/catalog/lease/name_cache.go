package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/nstree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

func makeNameCache() nameCache {
	__antithesis_instrumentation__.Notify(266585)
	return nameCache{}
}

type nameCache struct {
	mu          syncutil.RWMutex
	descriptors nstree.Map
}

func (c *nameCache) get(
	ctx context.Context,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
	timestamp hlc.Timestamp,
) *descriptorVersionState {
	__antithesis_instrumentation__.Notify(266586)
	c.mu.RLock()
	desc, ok := c.descriptors.GetByName(
		parentID, parentSchemaID, name,
	).(*descriptorVersionState)
	c.mu.RUnlock()
	if !ok {
		__antithesis_instrumentation__.Notify(266591)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266592)
	}
	__antithesis_instrumentation__.Notify(266587)
	expensiveLogEnabled := log.ExpensiveLogEnabled(ctx, 2)
	desc.mu.Lock()
	if desc.mu.lease == nil {
		__antithesis_instrumentation__.Notify(266593)
		desc.mu.Unlock()

		c.remove(desc)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266594)
	}
	__antithesis_instrumentation__.Notify(266588)

	defer desc.mu.Unlock()

	if !NameMatchesDescriptor(desc, parentID, parentSchemaID, name) {
		__antithesis_instrumentation__.Notify(266595)
		panic(errors.AssertionFailedf("out of sync entry in the name cache. "+
			"Cache entry: (%d, %d, %q) -> %d. Lease: (%d, %d, %q).",
			parentID, parentSchemaID, name,
			desc.GetID(),
			desc.GetParentID(), desc.GetParentSchemaID(), desc.GetName()),
		)
	} else {
		__antithesis_instrumentation__.Notify(266596)
	}
	__antithesis_instrumentation__.Notify(266589)

	if desc.hasExpiredLocked(timestamp) {
		__antithesis_instrumentation__.Notify(266597)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266598)
	}
	__antithesis_instrumentation__.Notify(266590)

	desc.incRefCountLocked(ctx, expensiveLogEnabled)
	return desc
}

func (c *nameCache) insert(desc *descriptorVersionState) {
	__antithesis_instrumentation__.Notify(266599)
	c.mu.Lock()
	defer c.mu.Unlock()
	got, ok := c.descriptors.GetByName(
		desc.GetParentID(), desc.GetParentSchemaID(), desc.GetName(),
	).(*descriptorVersionState)
	if ok && func() bool {
		__antithesis_instrumentation__.Notify(266601)
		return desc.getExpiration().Less(got.getExpiration()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(266602)
		return
	} else {
		__antithesis_instrumentation__.Notify(266603)
	}
	__antithesis_instrumentation__.Notify(266600)
	c.descriptors.Upsert(desc)
}

func (c *nameCache) remove(desc *descriptorVersionState) {
	__antithesis_instrumentation__.Notify(266604)
	c.mu.Lock()
	defer c.mu.Unlock()

	if got := c.descriptors.GetByID(desc.GetID()); got == desc {
		__antithesis_instrumentation__.Notify(266605)
		c.descriptors.Remove(desc.GetID())
	} else {
		__antithesis_instrumentation__.Notify(266606)
	}
}
