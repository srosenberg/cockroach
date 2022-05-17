package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type txnMetricRecorder struct {
	wrapped lockedSender
	metrics *TxnMetrics
	clock   *hlc.Clock

	txn            *roachpb.Transaction
	txnStartNanos  int64
	onePCCommit    bool
	parallelCommit bool
}

func (m *txnMetricRecorder) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88727)
	if m.txnStartNanos == 0 {
		__antithesis_instrumentation__.Notify(88731)
		m.txnStartNanos = timeutil.Now().UnixNano()
	} else {
		__antithesis_instrumentation__.Notify(88732)
	}
	__antithesis_instrumentation__.Notify(88728)

	br, pErr := m.wrapped.SendLocked(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88733)
		return br, pErr
	} else {
		__antithesis_instrumentation__.Notify(88734)
	}
	__antithesis_instrumentation__.Notify(88729)

	if length := len(br.Responses); length > 0 {
		__antithesis_instrumentation__.Notify(88735)
		if et := br.Responses[length-1].GetEndTxn(); et != nil {
			__antithesis_instrumentation__.Notify(88736)

			m.onePCCommit = et.OnePhaseCommit

			m.parallelCommit = !et.StagingTimestamp.IsEmpty()
		} else {
			__antithesis_instrumentation__.Notify(88737)
		}
	} else {
		__antithesis_instrumentation__.Notify(88738)
	}
	__antithesis_instrumentation__.Notify(88730)
	return br, nil
}

func (m *txnMetricRecorder) setWrapped(wrapped lockedSender) {
	__antithesis_instrumentation__.Notify(88739)
	m.wrapped = wrapped
}

func (*txnMetricRecorder) populateLeafInputState(*roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(88740)
}

func (*txnMetricRecorder) populateLeafFinalState(*roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88741)
}

func (*txnMetricRecorder) importLeafFinalState(context.Context, *roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88742)
}

func (*txnMetricRecorder) epochBumpedLocked() { __antithesis_instrumentation__.Notify(88743) }

func (*txnMetricRecorder) createSavepointLocked(context.Context, *savepoint) {
	__antithesis_instrumentation__.Notify(88744)
}

func (*txnMetricRecorder) rollbackToSavepointLocked(context.Context, savepoint) {
	__antithesis_instrumentation__.Notify(88745)
}

func (m *txnMetricRecorder) closeLocked() {
	__antithesis_instrumentation__.Notify(88746)
	if m.onePCCommit {
		__antithesis_instrumentation__.Notify(88751)
		m.metrics.Commits1PC.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(88752)
	}
	__antithesis_instrumentation__.Notify(88747)
	if m.parallelCommit {
		__antithesis_instrumentation__.Notify(88753)
		m.metrics.ParallelCommits.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(88754)
	}
	__antithesis_instrumentation__.Notify(88748)

	if m.txnStartNanos != 0 {
		__antithesis_instrumentation__.Notify(88755)
		duration := timeutil.Now().UnixNano() - m.txnStartNanos
		if duration >= 0 {
			__antithesis_instrumentation__.Notify(88756)
			m.metrics.Durations.RecordValue(duration)
		} else {
			__antithesis_instrumentation__.Notify(88757)
		}
	} else {
		__antithesis_instrumentation__.Notify(88758)
	}
	__antithesis_instrumentation__.Notify(88749)
	restarts := int64(m.txn.Epoch)
	status := m.txn.Status

	if restarts > 0 {
		__antithesis_instrumentation__.Notify(88759)
		m.metrics.Restarts.RecordValue(restarts)
	} else {
		__antithesis_instrumentation__.Notify(88760)
	}
	__antithesis_instrumentation__.Notify(88750)
	switch status {
	case roachpb.ABORTED:
		__antithesis_instrumentation__.Notify(88761)
		m.metrics.Aborts.Inc(1)
	case roachpb.PENDING:
		__antithesis_instrumentation__.Notify(88762)

		m.metrics.Aborts.Inc(1)

		m.metrics.RollbacksFailed.Inc(1)
	case roachpb.COMMITTED:
		__antithesis_instrumentation__.Notify(88763)

		m.metrics.Commits.Inc(1)
	default:
		__antithesis_instrumentation__.Notify(88764)
	}
}
