// Package lock provides type definitions for locking-related concepts used by
// concurrency control in the key-value layer.
package lock

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/redact"

func (lw Waiter) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(99165)
	expand := w.Flag('+')

	txnIDRedactableString := redact.Sprint(nil)
	if lw.WaitingTxn != nil {
		__antithesis_instrumentation__.Notify(99167)
		if expand {
			__antithesis_instrumentation__.Notify(99168)
			txnIDRedactableString = redact.Sprint(lw.WaitingTxn.ID)
		} else {
			__antithesis_instrumentation__.Notify(99169)
			txnIDRedactableString = redact.Sprint(lw.WaitingTxn.Short())
		}
	} else {
		__antithesis_instrumentation__.Notify(99170)
	}
	__antithesis_instrumentation__.Notify(99166)
	w.Printf("waiting_txn:%s active_waiter:%t strength:%s wait_duration:%s", txnIDRedactableString, lw.ActiveWaiter, lw.Strength, lw.WaitDuration)
}
