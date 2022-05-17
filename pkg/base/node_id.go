package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strconv"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type NodeIDContainer struct {
	_ util.NoCopy

	nodeID int32

	sqlInstance bool

	str atomic.Value
}

func (n *NodeIDContainer) String() string {
	__antithesis_instrumentation__.Notify(1470)
	s := n.str.Load()
	if s == nil {
		__antithesis_instrumentation__.Notify(1472)
		if n.sqlInstance {
			__antithesis_instrumentation__.Notify(1474)
			return "sql?"
		} else {
			__antithesis_instrumentation__.Notify(1475)
		}
		__antithesis_instrumentation__.Notify(1473)
		return "?"
	} else {
		__antithesis_instrumentation__.Notify(1476)
	}
	__antithesis_instrumentation__.Notify(1471)
	return s.(string)
}

var _ redact.SafeValue = &NodeIDContainer{}

func (n *NodeIDContainer) SafeValue() { __antithesis_instrumentation__.Notify(1477) }

func (n *NodeIDContainer) Get() roachpb.NodeID {
	__antithesis_instrumentation__.Notify(1478)
	return roachpb.NodeID(atomic.LoadInt32(&n.nodeID))
}

func (n *NodeIDContainer) Set(ctx context.Context, val roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(1479)
	n.setInternal(ctx, int32(val), false)
}

func (n *NodeIDContainer) setInternal(ctx context.Context, val int32, sqlInstance bool) {
	__antithesis_instrumentation__.Notify(1480)
	if val <= 0 {
		__antithesis_instrumentation__.Notify(1484)
		log.Fatalf(ctx, "trying to set invalid NodeID: %d", val)
	} else {
		__antithesis_instrumentation__.Notify(1485)
	}
	__antithesis_instrumentation__.Notify(1481)
	oldVal := atomic.SwapInt32(&n.nodeID, val)
	if oldVal == 0 {
		__antithesis_instrumentation__.Notify(1486)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(1487)
			log.Infof(ctx, "ID set to %d", val)
		} else {
			__antithesis_instrumentation__.Notify(1488)
		}
	} else {
		__antithesis_instrumentation__.Notify(1489)
		if n.sqlInstance != sqlInstance {
			__antithesis_instrumentation__.Notify(1490)
			serverIs := map[bool]redact.SafeString{false: "SQL instance", true: "node"}
			log.Fatalf(ctx, "server is a %v, cannot set %v ID", serverIs[!n.sqlInstance], serverIs[sqlInstance])
		} else {
			__antithesis_instrumentation__.Notify(1491)
			if oldVal != val {
				__antithesis_instrumentation__.Notify(1492)
				log.Fatalf(ctx, "different IDs set: %d, then %d", oldVal, val)
			} else {
				__antithesis_instrumentation__.Notify(1493)
			}
		}
	}
	__antithesis_instrumentation__.Notify(1482)
	prefix := ""
	if sqlInstance {
		__antithesis_instrumentation__.Notify(1494)
		prefix = "sql"
	} else {
		__antithesis_instrumentation__.Notify(1495)
	}
	__antithesis_instrumentation__.Notify(1483)
	n.str.Store(prefix + strconv.Itoa(int(val)))
}

func (n *NodeIDContainer) Reset(val roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(1496)
	atomic.StoreInt32(&n.nodeID, int32(val))
	n.str.Store(strconv.Itoa(int(val)))
}

type StoreIDContainer struct {
	_ util.NoCopy

	storeID int32

	str atomic.Value
}

const TempStoreID = -1

func (s *StoreIDContainer) String() string {
	__antithesis_instrumentation__.Notify(1497)
	str := s.str.Load()
	if str == nil {
		__antithesis_instrumentation__.Notify(1499)
		return "?"
	} else {
		__antithesis_instrumentation__.Notify(1500)
	}
	__antithesis_instrumentation__.Notify(1498)
	return str.(string)
}

var _ redact.SafeValue = &StoreIDContainer{}

func (s *StoreIDContainer) SafeValue() { __antithesis_instrumentation__.Notify(1501) }

func (s *StoreIDContainer) Get() int32 {
	__antithesis_instrumentation__.Notify(1502)
	return atomic.LoadInt32(&s.storeID)
}

func (s *StoreIDContainer) Set(ctx context.Context, val int32) {
	__antithesis_instrumentation__.Notify(1503)
	if val != TempStoreID && func() bool {
		__antithesis_instrumentation__.Notify(1506)
		return val <= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(1507)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(1509)
			log.Infof(
				ctx, "trying to set invalid storeID for the store in the Pebble log: %d",
				val)
		} else {
			__antithesis_instrumentation__.Notify(1510)
		}
		__antithesis_instrumentation__.Notify(1508)
		return
	} else {
		__antithesis_instrumentation__.Notify(1511)
	}
	__antithesis_instrumentation__.Notify(1504)
	oldVal := atomic.SwapInt32(&s.storeID, val)
	if oldVal != 0 && func() bool {
		__antithesis_instrumentation__.Notify(1512)
		return oldVal != val == true
	}() == true {
		__antithesis_instrumentation__.Notify(1513)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(1514)
			log.Infof(
				ctx, "different storeIDs set for the store in the Pebble log: %d, then %d",
				oldVal, val)
		} else {
			__antithesis_instrumentation__.Notify(1515)
		}
	} else {
		__antithesis_instrumentation__.Notify(1516)
	}
	__antithesis_instrumentation__.Notify(1505)
	if val == TempStoreID {
		__antithesis_instrumentation__.Notify(1517)
		s.str.Store("temp")
	} else {
		__antithesis_instrumentation__.Notify(1518)
		s.str.Store(strconv.Itoa(int(val)))
	}
}

type SQLInstanceID int32

func (s SQLInstanceID) String() string {
	__antithesis_instrumentation__.Notify(1519)
	if s == 0 {
		__antithesis_instrumentation__.Notify(1521)
		return "?"
	} else {
		__antithesis_instrumentation__.Notify(1522)
	}
	__antithesis_instrumentation__.Notify(1520)
	return strconv.Itoa(int(s))
}

type SQLIDContainer NodeIDContainer

func NewSQLIDContainerForNode(nodeID *NodeIDContainer) *SQLIDContainer {
	__antithesis_instrumentation__.Notify(1523)
	if nodeID.sqlInstance {
		__antithesis_instrumentation__.Notify(1525)

		log.Fatalf(context.Background(), "programming error: container is already for a SQL instance")
	} else {
		__antithesis_instrumentation__.Notify(1526)
	}
	__antithesis_instrumentation__.Notify(1524)
	return (*SQLIDContainer)(nodeID)
}

func (n *NodeIDContainer) SwitchToSQLIDContainer() *SQLIDContainer {
	__antithesis_instrumentation__.Notify(1527)
	sc := NewSQLIDContainerForNode(n)
	sc.sqlInstance = true
	return sc
}

func (c *SQLIDContainer) SetSQLInstanceID(ctx context.Context, sqlInstanceID SQLInstanceID) error {
	__antithesis_instrumentation__.Notify(1528)
	if !c.sqlInstance {
		__antithesis_instrumentation__.Notify(1530)
		return errors.New("attempting to initialize instance ID when node ID is set")
	} else {
		__antithesis_instrumentation__.Notify(1531)
	}
	__antithesis_instrumentation__.Notify(1529)

	(*NodeIDContainer)(c).setInternal(ctx, int32(sqlInstanceID), true)
	return nil
}

func (c *SQLIDContainer) OptionalNodeID() (roachpb.NodeID, bool) {
	__antithesis_instrumentation__.Notify(1532)
	if (*NodeIDContainer)(c).sqlInstance {
		__antithesis_instrumentation__.Notify(1534)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(1535)
	}
	__antithesis_instrumentation__.Notify(1533)
	return (*NodeIDContainer)(c).Get(), true
}

func (c *SQLIDContainer) OptionalNodeIDErr(issue int) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(1536)
	if (*NodeIDContainer)(c).sqlInstance {
		__antithesis_instrumentation__.Notify(1538)
		return 0, errorutil.UnsupportedWithMultiTenancy(issue)
	} else {
		__antithesis_instrumentation__.Notify(1539)
	}
	__antithesis_instrumentation__.Notify(1537)
	return (*NodeIDContainer)(c).Get(), nil
}

func (c *SQLIDContainer) SQLInstanceID() SQLInstanceID {
	__antithesis_instrumentation__.Notify(1540)
	return SQLInstanceID((*NodeIDContainer)(c).Get())
}

func (c *SQLIDContainer) SafeValue() { __antithesis_instrumentation__.Notify(1541) }

func (c *SQLIDContainer) String() string {
	__antithesis_instrumentation__.Notify(1542)
	return (*NodeIDContainer)(c).String()
}

var TestingIDContainer = func() *SQLIDContainer {
	__antithesis_instrumentation__.Notify(1543)
	var c NodeIDContainer
	sc := c.SwitchToSQLIDContainer()
	if err := sc.SetSQLInstanceID(context.Background(), 10); err != nil {
		__antithesis_instrumentation__.Notify(1545)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(1546)
	}
	__antithesis_instrumentation__.Notify(1544)
	return sc
}()
