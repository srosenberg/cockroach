package result

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func FromAcquiredLocks(txn *roachpb.Transaction, keys ...roachpb.Key) Result {
	__antithesis_instrumentation__.Notify(97660)
	var pd Result
	if txn == nil {
		__antithesis_instrumentation__.Notify(97663)
		return pd
	} else {
		__antithesis_instrumentation__.Notify(97664)
	}
	__antithesis_instrumentation__.Notify(97661)
	pd.Local.AcquiredLocks = make([]roachpb.LockAcquisition, len(keys))
	for i := range pd.Local.AcquiredLocks {
		__antithesis_instrumentation__.Notify(97665)
		pd.Local.AcquiredLocks[i] = roachpb.MakeLockAcquisition(txn, keys[i], lock.Replicated)
	}
	__antithesis_instrumentation__.Notify(97662)
	return pd
}

type EndTxnIntents struct {
	Txn    *roachpb.Transaction
	Always bool
	Poison bool
}

func FromEndTxn(txn *roachpb.Transaction, alwaysReturn, poison bool) Result {
	__antithesis_instrumentation__.Notify(97666)
	var pd Result
	if len(txn.LockSpans) == 0 {
		__antithesis_instrumentation__.Notify(97668)
		return pd
	} else {
		__antithesis_instrumentation__.Notify(97669)
	}
	__antithesis_instrumentation__.Notify(97667)
	pd.Local.EndTxns = []EndTxnIntents{{Txn: txn, Always: alwaysReturn, Poison: poison}}
	return pd
}
