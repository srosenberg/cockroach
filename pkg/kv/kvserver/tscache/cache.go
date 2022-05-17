// Package tscache provides a timestamp cache structure that records the maximum
// timestamp that key ranges were read from and written to.
package tscache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const MinRetentionWindow = 10 * time.Second

type Cache interface {
	Add(start, end roachpb.Key, ts hlc.Timestamp, txnID uuid.UUID)

	GetMax(start, end roachpb.Key) (hlc.Timestamp, uuid.UUID)

	Metrics() Metrics

	clear(lowWater hlc.Timestamp)

	getLowWater() hlc.Timestamp
}

func New(clock *hlc.Clock) Cache {
	__antithesis_instrumentation__.Notify(126796)
	if envutil.EnvOrDefaultBool("COCKROACH_USE_TREE_TSCACHE", false) {
		__antithesis_instrumentation__.Notify(126798)
		return newTreeImpl(clock)
	} else {
		__antithesis_instrumentation__.Notify(126799)
	}
	__antithesis_instrumentation__.Notify(126797)
	return newSklImpl(clock)
}

type cacheValue struct {
	ts    hlc.Timestamp
	txnID uuid.UUID
}

var noTxnID uuid.UUID

func (v cacheValue) String() string {
	__antithesis_instrumentation__.Notify(126800)
	var txnIDStr string
	switch v.txnID {
	case noTxnID:
		__antithesis_instrumentation__.Notify(126802)
		txnIDStr = "none"
	default:
		__antithesis_instrumentation__.Notify(126803)
		txnIDStr = v.txnID.String()
	}
	__antithesis_instrumentation__.Notify(126801)
	return fmt.Sprintf("{ts: %s, txnID: %s}", v.ts, txnIDStr)
}

func ratchetValue(old, new cacheValue) (res cacheValue, updated bool) {
	__antithesis_instrumentation__.Notify(126804)
	if old.ts.Less(new.ts) {
		__antithesis_instrumentation__.Notify(126808)

		return new, true
	} else {
		__antithesis_instrumentation__.Notify(126809)
		if new.ts.Less(old.ts) {
			__antithesis_instrumentation__.Notify(126810)

			return old, false
		} else {
			__antithesis_instrumentation__.Notify(126811)
		}
	}
	__antithesis_instrumentation__.Notify(126805)

	if new.ts.Synthetic != old.ts.Synthetic {
		__antithesis_instrumentation__.Notify(126812)

		new.ts.Synthetic = false
	} else {
		__antithesis_instrumentation__.Notify(126813)
	}
	__antithesis_instrumentation__.Notify(126806)
	if new.txnID != old.txnID {
		__antithesis_instrumentation__.Notify(126814)

		new.txnID = noTxnID
	} else {
		__antithesis_instrumentation__.Notify(126815)
	}
	__antithesis_instrumentation__.Notify(126807)
	return new, new != old
}
