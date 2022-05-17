package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/errors"
)

type txnSeqNumAllocator struct {
	wrapped lockedSender

	writeSeq enginepb.TxnSeq

	readSeq enginepb.TxnSeq

	steppingModeEnabled bool
}

func (s *txnSeqNumAllocator) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89031)
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(89033)
		req := ru.GetInner()

		if roachpb.IsIntentWrite(req) || func() bool {
			__antithesis_instrumentation__.Notify(89036)
			return req.Method() == roachpb.EndTxn == true
		}() == true {
			__antithesis_instrumentation__.Notify(89037)
			s.writeSeq++
		} else {
			__antithesis_instrumentation__.Notify(89038)
		}
		__antithesis_instrumentation__.Notify(89034)

		oldHeader := req.Header()
		oldHeader.Sequence = s.writeSeq
		if s.steppingModeEnabled && func() bool {
			__antithesis_instrumentation__.Notify(89039)
			return roachpb.IsReadOnly(req) == true
		}() == true {
			__antithesis_instrumentation__.Notify(89040)
			oldHeader.Sequence = s.readSeq
		} else {
			__antithesis_instrumentation__.Notify(89041)
		}
		__antithesis_instrumentation__.Notify(89035)
		req.SetHeader(oldHeader)
	}
	__antithesis_instrumentation__.Notify(89032)

	return s.wrapped.SendLocked(ctx, ba)
}

func (s *txnSeqNumAllocator) setWrapped(wrapped lockedSender) {
	__antithesis_instrumentation__.Notify(89042)
	s.wrapped = wrapped
}

func (s *txnSeqNumAllocator) populateLeafInputState(tis *roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(89043)
	tis.Txn.Sequence = s.writeSeq
	tis.SteppingModeEnabled = s.steppingModeEnabled
	tis.ReadSeqNum = s.readSeq
}

func (s *txnSeqNumAllocator) initializeLeaf(tis *roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(89044)
	s.steppingModeEnabled = tis.SteppingModeEnabled
	s.readSeq = tis.ReadSeqNum
}

func (s *txnSeqNumAllocator) populateLeafFinalState(tfs *roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(89045)
}

func (s *txnSeqNumAllocator) importLeafFinalState(
	ctx context.Context, tfs *roachpb.LeafTxnFinalState,
) {
	__antithesis_instrumentation__.Notify(89046)
}

func (s *txnSeqNumAllocator) stepLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(89047)
	if !s.steppingModeEnabled {
		__antithesis_instrumentation__.Notify(89050)
		return errors.AssertionFailedf("stepping mode is not enabled")
	} else {
		__antithesis_instrumentation__.Notify(89051)
	}
	__antithesis_instrumentation__.Notify(89048)
	if s.readSeq > s.writeSeq {
		__antithesis_instrumentation__.Notify(89052)
		return errors.AssertionFailedf(
			"cannot step() after mistaken initialization (%d,%d)", s.writeSeq, s.readSeq)
	} else {
		__antithesis_instrumentation__.Notify(89053)
	}
	__antithesis_instrumentation__.Notify(89049)
	s.readSeq = s.writeSeq
	return nil
}

func (s *txnSeqNumAllocator) configureSteppingLocked(
	newMode kv.SteppingMode,
) (prevMode kv.SteppingMode) {
	__antithesis_instrumentation__.Notify(89054)
	prevEnabled := s.steppingModeEnabled
	enabled := newMode == kv.SteppingEnabled
	s.steppingModeEnabled = enabled
	if !prevEnabled && func() bool {
		__antithesis_instrumentation__.Notify(89057)
		return enabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(89058)
		s.readSeq = s.writeSeq
	} else {
		__antithesis_instrumentation__.Notify(89059)
	}
	__antithesis_instrumentation__.Notify(89055)
	prevMode = kv.SteppingDisabled
	if prevEnabled {
		__antithesis_instrumentation__.Notify(89060)
		prevMode = kv.SteppingEnabled
	} else {
		__antithesis_instrumentation__.Notify(89061)
	}
	__antithesis_instrumentation__.Notify(89056)
	return prevMode
}

func (s *txnSeqNumAllocator) epochBumpedLocked() {
	__antithesis_instrumentation__.Notify(89062)

	s.writeSeq = 0
	s.readSeq = 0
}

func (s *txnSeqNumAllocator) createSavepointLocked(ctx context.Context, sp *savepoint) {
	__antithesis_instrumentation__.Notify(89063)
	sp.seqNum = s.writeSeq
}

func (*txnSeqNumAllocator) rollbackToSavepointLocked(context.Context, savepoint) {
	__antithesis_instrumentation__.Notify(89064)

}

func (*txnSeqNumAllocator) closeLocked() { __antithesis_instrumentation__.Notify(89065) }
