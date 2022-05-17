package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type ClusterIDContainer struct {
	clusterID atomic.Value

	OnSet func(uuid.UUID)
}

func (c *ClusterIDContainer) String() string {
	__antithesis_instrumentation__.Notify(1381)
	val := c.Get()
	if val == uuid.Nil {
		__antithesis_instrumentation__.Notify(1383)
		return "?"
	} else {
		__antithesis_instrumentation__.Notify(1384)
	}
	__antithesis_instrumentation__.Notify(1382)
	return val.String()
}

func (c *ClusterIDContainer) Get() uuid.UUID {
	__antithesis_instrumentation__.Notify(1385)
	v := c.clusterID.Load()
	if v == nil {
		__antithesis_instrumentation__.Notify(1387)
		return uuid.Nil
	} else {
		__antithesis_instrumentation__.Notify(1388)
	}
	__antithesis_instrumentation__.Notify(1386)
	return v.(uuid.UUID)
}

func (c *ClusterIDContainer) Set(ctx context.Context, val uuid.UUID) {
	__antithesis_instrumentation__.Notify(1389)

	cur := c.Get()
	if cur == uuid.Nil {
		__antithesis_instrumentation__.Notify(1391)
		c.clusterID.Store(val)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(1392)
			log.Infof(ctx, "ClusterID set to %s", val)
		} else {
			__antithesis_instrumentation__.Notify(1393)
		}
	} else {
		__antithesis_instrumentation__.Notify(1394)
		if cur != val {
			__antithesis_instrumentation__.Notify(1395)
			log.Fatalf(ctx, "different ClusterIDs set: %s, then %s", cur, val)
		} else {
			__antithesis_instrumentation__.Notify(1396)
		}
	}
	__antithesis_instrumentation__.Notify(1390)
	if c.OnSet != nil {
		__antithesis_instrumentation__.Notify(1397)
		c.OnSet(val)
	} else {
		__antithesis_instrumentation__.Notify(1398)
	}
}

func (c *ClusterIDContainer) Reset(val uuid.UUID) {
	__antithesis_instrumentation__.Notify(1399)
	c.clusterID.Store(val)
}
