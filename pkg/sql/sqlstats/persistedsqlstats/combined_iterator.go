package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/errors"
)

type CombinedStmtStatsIterator struct {
	nextToRead     *roachpb.CollectedStatementStatistics
	expectedColCnt int

	mem struct {
		canBeAdvanced bool
		paused        bool
		it            *memStmtStatsIterator
	}

	disk struct {
		canBeAdvanced bool
		paused        bool
		it            sqlutil.InternalRows
	}
}

func NewCombinedStmtStatsIterator(
	memIter *memStmtStatsIterator, diskIter sqlutil.InternalRows, expectedColCnt int,
) *CombinedStmtStatsIterator {
	__antithesis_instrumentation__.Notify(624441)
	c := &CombinedStmtStatsIterator{
		expectedColCnt: expectedColCnt,
	}

	c.mem.it = memIter
	c.mem.canBeAdvanced = true

	c.disk.it = diskIter
	c.disk.canBeAdvanced = true

	return c
}

func (c *CombinedStmtStatsIterator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(624442)
	var err error

	if c.mem.canBeAdvanced && func() bool {
		__antithesis_instrumentation__.Notify(624450)
		return !c.mem.paused == true
	}() == true {
		__antithesis_instrumentation__.Notify(624451)
		c.mem.canBeAdvanced = c.mem.it.Next()
	} else {
		__antithesis_instrumentation__.Notify(624452)
	}
	__antithesis_instrumentation__.Notify(624443)

	if c.disk.canBeAdvanced && func() bool {
		__antithesis_instrumentation__.Notify(624453)
		return !c.disk.paused == true
	}() == true {
		__antithesis_instrumentation__.Notify(624454)
		c.disk.canBeAdvanced, err = c.disk.it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(624455)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(624456)
		}
	} else {
		__antithesis_instrumentation__.Notify(624457)
	}
	__antithesis_instrumentation__.Notify(624444)

	if !c.mem.canBeAdvanced && func() bool {
		__antithesis_instrumentation__.Notify(624458)
		return !c.disk.canBeAdvanced == true
	}() == true {
		__antithesis_instrumentation__.Notify(624459)

		if c.mem.paused || func() bool {
			__antithesis_instrumentation__.Notify(624461)
			return c.disk.paused == true
		}() == true {
			__antithesis_instrumentation__.Notify(624462)
			return false, errors.AssertionFailedf("bug: leaked iterator")
		} else {
			__antithesis_instrumentation__.Notify(624463)
		}
		__antithesis_instrumentation__.Notify(624460)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(624464)
	}
	__antithesis_instrumentation__.Notify(624445)

	if !c.mem.canBeAdvanced {
		__antithesis_instrumentation__.Notify(624465)
		row := c.disk.it.Cur()
		if row == nil {
			__antithesis_instrumentation__.Notify(624470)
			return false, errors.New("unexpected nil row")
		} else {
			__antithesis_instrumentation__.Notify(624471)
		}
		__antithesis_instrumentation__.Notify(624466)

		if len(row) != c.expectedColCnt {
			__antithesis_instrumentation__.Notify(624472)
			return false, errors.AssertionFailedf("unexpectedly received %d columns", len(row))
		} else {
			__antithesis_instrumentation__.Notify(624473)
		}
		__antithesis_instrumentation__.Notify(624467)

		c.nextToRead, err = rowToStmtStats(c.disk.it.Cur())
		if err != nil {
			__antithesis_instrumentation__.Notify(624474)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(624475)
		}
		__antithesis_instrumentation__.Notify(624468)

		if c.disk.canBeAdvanced {
			__antithesis_instrumentation__.Notify(624476)
			c.disk.paused = false
		} else {
			__antithesis_instrumentation__.Notify(624477)
		}
		__antithesis_instrumentation__.Notify(624469)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(624478)
	}
	__antithesis_instrumentation__.Notify(624446)

	if !c.disk.canBeAdvanced {
		__antithesis_instrumentation__.Notify(624479)
		c.nextToRead = c.mem.it.Cur()

		if c.mem.canBeAdvanced {
			__antithesis_instrumentation__.Notify(624481)
			c.mem.paused = false
		} else {
			__antithesis_instrumentation__.Notify(624482)
		}
		__antithesis_instrumentation__.Notify(624480)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(624483)
	}
	__antithesis_instrumentation__.Notify(624447)

	memCurVal := c.mem.it.Cur()
	diskCurVal, err := rowToStmtStats(c.disk.it.Cur())
	if err != nil {
		__antithesis_instrumentation__.Notify(624484)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(624485)
	}
	__antithesis_instrumentation__.Notify(624448)

	switch compareStmtStats(memCurVal, diskCurVal) {
	case -1:
		__antithesis_instrumentation__.Notify(624486)

		c.nextToRead = memCurVal
		c.mem.paused = false
		c.disk.paused = true
	case 0:
		__antithesis_instrumentation__.Notify(624487)

		c.nextToRead = memCurVal
		c.nextToRead.Stats.Add(&diskCurVal.Stats)
		c.mem.paused = false
		c.disk.paused = false
	case 1:
		__antithesis_instrumentation__.Notify(624488)

		c.nextToRead = diskCurVal
		c.mem.paused = true
		c.disk.paused = false
	default:
		__antithesis_instrumentation__.Notify(624489)
		return false, errors.AssertionFailedf("bug: impossible state")
	}
	__antithesis_instrumentation__.Notify(624449)

	return true, nil
}

func (c *CombinedStmtStatsIterator) Cur() *roachpb.CollectedStatementStatistics {
	__antithesis_instrumentation__.Notify(624490)
	return c.nextToRead
}

func compareStmtStats(lhs, rhs *roachpb.CollectedStatementStatistics) int {
	__antithesis_instrumentation__.Notify(624491)

	if lhs.AggregatedTs.Before(rhs.AggregatedTs) {
		__antithesis_instrumentation__.Notify(624501)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(624502)
	}
	__antithesis_instrumentation__.Notify(624492)
	if lhs.AggregatedTs.After(rhs.AggregatedTs) {
		__antithesis_instrumentation__.Notify(624503)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(624504)
	}
	__antithesis_instrumentation__.Notify(624493)

	cmp := strings.Compare(lhs.Key.App, rhs.Key.App)
	if cmp != 0 {
		__antithesis_instrumentation__.Notify(624505)
		return cmp
	} else {
		__antithesis_instrumentation__.Notify(624506)
	}
	__antithesis_instrumentation__.Notify(624494)

	if lhs.ID < rhs.ID {
		__antithesis_instrumentation__.Notify(624507)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(624508)
	}
	__antithesis_instrumentation__.Notify(624495)
	if lhs.ID > rhs.ID {
		__antithesis_instrumentation__.Notify(624509)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(624510)
	}
	__antithesis_instrumentation__.Notify(624496)

	if lhs.Key.TransactionFingerprintID < rhs.Key.TransactionFingerprintID {
		__antithesis_instrumentation__.Notify(624511)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(624512)
	}
	__antithesis_instrumentation__.Notify(624497)
	if lhs.Key.TransactionFingerprintID > rhs.Key.TransactionFingerprintID {
		__antithesis_instrumentation__.Notify(624513)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(624514)
	}
	__antithesis_instrumentation__.Notify(624498)

	if lhs.Key.PlanHash < rhs.Key.PlanHash {
		__antithesis_instrumentation__.Notify(624515)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(624516)
	}
	__antithesis_instrumentation__.Notify(624499)
	if lhs.Key.PlanHash > rhs.Key.PlanHash {
		__antithesis_instrumentation__.Notify(624517)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(624518)
	}
	__antithesis_instrumentation__.Notify(624500)

	return 0
}

type CombinedTxnStatsIterator struct {
	nextToReadVal  *roachpb.CollectedTransactionStatistics
	expectedColCnt int

	mem struct {
		canBeAdvanced bool
		paused        bool
		it            *memTxnStatsIterator
	}

	disk struct {
		canBeAdvanced bool
		paused        bool
		it            sqlutil.InternalRows
	}
}

func NewCombinedTxnStatsIterator(
	memIter *memTxnStatsIterator, diskIter sqlutil.InternalRows, expectedColCnt int,
) *CombinedTxnStatsIterator {
	__antithesis_instrumentation__.Notify(624519)
	c := &CombinedTxnStatsIterator{
		expectedColCnt: expectedColCnt,
	}

	c.mem.it = memIter
	c.mem.canBeAdvanced = true

	c.disk.it = diskIter
	c.disk.canBeAdvanced = true

	return c
}

func (c *CombinedTxnStatsIterator) Next(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(624520)
	var err error

	if c.mem.canBeAdvanced && func() bool {
		__antithesis_instrumentation__.Notify(624528)
		return !c.mem.paused == true
	}() == true {
		__antithesis_instrumentation__.Notify(624529)
		c.mem.canBeAdvanced = c.mem.it.Next()
	} else {
		__antithesis_instrumentation__.Notify(624530)
	}
	__antithesis_instrumentation__.Notify(624521)

	if c.disk.canBeAdvanced && func() bool {
		__antithesis_instrumentation__.Notify(624531)
		return !c.disk.paused == true
	}() == true {
		__antithesis_instrumentation__.Notify(624532)
		c.disk.canBeAdvanced, err = c.disk.it.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(624533)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(624534)
		}
	} else {
		__antithesis_instrumentation__.Notify(624535)
	}
	__antithesis_instrumentation__.Notify(624522)

	if !c.mem.canBeAdvanced && func() bool {
		__antithesis_instrumentation__.Notify(624536)
		return !c.disk.canBeAdvanced == true
	}() == true {
		__antithesis_instrumentation__.Notify(624537)

		if c.mem.paused || func() bool {
			__antithesis_instrumentation__.Notify(624539)
			return c.disk.paused == true
		}() == true {
			__antithesis_instrumentation__.Notify(624540)
			return false, errors.AssertionFailedf("bug: leaked iterator")
		} else {
			__antithesis_instrumentation__.Notify(624541)
		}
		__antithesis_instrumentation__.Notify(624538)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(624542)
	}
	__antithesis_instrumentation__.Notify(624523)

	if !c.mem.canBeAdvanced {
		__antithesis_instrumentation__.Notify(624543)
		row := c.disk.it.Cur()
		if row == nil {
			__antithesis_instrumentation__.Notify(624548)
			return false, errors.New("unexpected nil row")
		} else {
			__antithesis_instrumentation__.Notify(624549)
		}
		__antithesis_instrumentation__.Notify(624544)

		if len(row) != c.expectedColCnt {
			__antithesis_instrumentation__.Notify(624550)
			return false, errors.AssertionFailedf("unexpectedly received %d columns", len(row))
		} else {
			__antithesis_instrumentation__.Notify(624551)
		}
		__antithesis_instrumentation__.Notify(624545)

		c.nextToReadVal, err = rowToTxnStats(c.disk.it.Cur())
		if err != nil {
			__antithesis_instrumentation__.Notify(624552)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(624553)
		}
		__antithesis_instrumentation__.Notify(624546)

		if c.disk.canBeAdvanced {
			__antithesis_instrumentation__.Notify(624554)
			c.disk.paused = false
		} else {
			__antithesis_instrumentation__.Notify(624555)
		}
		__antithesis_instrumentation__.Notify(624547)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(624556)
	}
	__antithesis_instrumentation__.Notify(624524)

	if !c.disk.canBeAdvanced {
		__antithesis_instrumentation__.Notify(624557)
		c.nextToReadVal = c.mem.it.Cur()

		if c.mem.canBeAdvanced {
			__antithesis_instrumentation__.Notify(624559)
			c.mem.paused = false
		} else {
			__antithesis_instrumentation__.Notify(624560)
		}
		__antithesis_instrumentation__.Notify(624558)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(624561)
	}
	__antithesis_instrumentation__.Notify(624525)

	memCurVal := c.mem.it.Cur()
	diskCurVal, err := rowToTxnStats(c.disk.it.Cur())
	if err != nil {
		__antithesis_instrumentation__.Notify(624562)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(624563)
	}
	__antithesis_instrumentation__.Notify(624526)

	switch compareTxnStats(memCurVal, diskCurVal) {
	case -1:
		__antithesis_instrumentation__.Notify(624564)

		c.nextToReadVal = memCurVal
		c.mem.paused = false
		c.disk.paused = true
	case 0:
		__antithesis_instrumentation__.Notify(624565)

		c.nextToReadVal = memCurVal
		c.nextToReadVal.Stats.Add(&diskCurVal.Stats)
		c.mem.paused = false
		c.disk.paused = false
	case 1:
		__antithesis_instrumentation__.Notify(624566)

		c.nextToReadVal = diskCurVal
		c.mem.paused = true
		c.disk.paused = false
	default:
		__antithesis_instrumentation__.Notify(624567)
		return false, errors.AssertionFailedf("bug: impossible state")
	}
	__antithesis_instrumentation__.Notify(624527)

	return true, nil
}

func (c *CombinedTxnStatsIterator) Cur() *roachpb.CollectedTransactionStatistics {
	__antithesis_instrumentation__.Notify(624568)
	return c.nextToReadVal
}

func compareTxnStats(lhs, rhs *roachpb.CollectedTransactionStatistics) int {
	__antithesis_instrumentation__.Notify(624569)

	if lhs.AggregatedTs.Before(rhs.AggregatedTs) {
		__antithesis_instrumentation__.Notify(624575)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(624576)
	}
	__antithesis_instrumentation__.Notify(624570)
	if lhs.AggregatedTs.After(rhs.AggregatedTs) {
		__antithesis_instrumentation__.Notify(624577)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(624578)
	}
	__antithesis_instrumentation__.Notify(624571)

	cmp := strings.Compare(lhs.App, rhs.App)
	if cmp != 0 {
		__antithesis_instrumentation__.Notify(624579)
		return cmp
	} else {
		__antithesis_instrumentation__.Notify(624580)
	}
	__antithesis_instrumentation__.Notify(624572)

	if lhs.TransactionFingerprintID < rhs.TransactionFingerprintID {
		__antithesis_instrumentation__.Notify(624581)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(624582)
	}
	__antithesis_instrumentation__.Notify(624573)
	if lhs.TransactionFingerprintID > rhs.TransactionFingerprintID {
		__antithesis_instrumentation__.Notify(624583)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(624584)
	}
	__antithesis_instrumentation__.Notify(624574)

	return 0
}
