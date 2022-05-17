package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type savepoint struct {
	active bool

	txnID uuid.UUID
	epoch enginepb.TxnEpoch

	seqNum enginepb.TxnSeq

	refreshSpans   []roachpb.Span
	refreshInvalid bool
}

var _ kv.SavepointToken = (*savepoint)(nil)

var initialSavepoint = savepoint{}

func (s *savepoint) Initial() bool {
	__antithesis_instrumentation__.Notify(88410)
	return !s.active
}

func (tc *TxnCoordSender) CreateSavepoint(ctx context.Context) (kv.SavepointToken, error) {
	__antithesis_instrumentation__.Notify(88411)
	if tc.typ != kv.RootTxn {
		__antithesis_instrumentation__.Notify(88417)
		return nil, errors.AssertionFailedf("cannot get savepoint in non-root txn")
	} else {
		__antithesis_instrumentation__.Notify(88418)
	}
	__antithesis_instrumentation__.Notify(88412)

	tc.mu.Lock()
	defer tc.mu.Unlock()

	if err := tc.assertNotFinalized(); err != nil {
		__antithesis_instrumentation__.Notify(88419)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(88420)
	}
	__antithesis_instrumentation__.Notify(88413)

	if tc.mu.txnState != txnPending {
		__antithesis_instrumentation__.Notify(88421)
		return nil, ErrSavepointOperationInErrorTxn
	} else {
		__antithesis_instrumentation__.Notify(88422)
	}
	__antithesis_instrumentation__.Notify(88414)

	if !tc.mu.active {
		__antithesis_instrumentation__.Notify(88423)

		return &initialSavepoint, nil
	} else {
		__antithesis_instrumentation__.Notify(88424)
	}
	__antithesis_instrumentation__.Notify(88415)

	s := &savepoint{
		active: true,
		txnID:  tc.mu.txn.ID,
		epoch:  tc.mu.txn.Epoch,
	}
	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88425)
		reqInt.createSavepointLocked(ctx, s)
	}
	__antithesis_instrumentation__.Notify(88416)

	return s, nil
}

func (tc *TxnCoordSender) RollbackToSavepoint(ctx context.Context, s kv.SavepointToken) error {
	__antithesis_instrumentation__.Notify(88426)
	if tc.typ != kv.RootTxn {
		__antithesis_instrumentation__.Notify(88433)
		return errors.AssertionFailedf("cannot rollback savepoint in non-root txn")
	} else {
		__antithesis_instrumentation__.Notify(88434)
	}
	__antithesis_instrumentation__.Notify(88427)

	tc.mu.Lock()
	defer tc.mu.Unlock()

	if err := tc.assertNotFinalized(); err != nil {
		__antithesis_instrumentation__.Notify(88435)
		return err
	} else {
		__antithesis_instrumentation__.Notify(88436)
	}
	__antithesis_instrumentation__.Notify(88428)

	if tc.mu.txnState == txnError {
		__antithesis_instrumentation__.Notify(88437)
		return unimplemented.New("rollback_error", "cannot rollback to savepoint after error")
	} else {
		__antithesis_instrumentation__.Notify(88438)
	}
	__antithesis_instrumentation__.Notify(88429)

	sp := s.(*savepoint)
	err := tc.checkSavepointLocked(sp)
	if err != nil {
		__antithesis_instrumentation__.Notify(88439)
		if errors.Is(err, errSavepointInvalidAfterTxnRestart) {
			__antithesis_instrumentation__.Notify(88441)
			err = roachpb.NewTransactionRetryWithProtoRefreshError(
				"cannot rollback to savepoint after a transaction restart",
				tc.mu.txn.ID,
				tc.mu.txn,
			)
		} else {
			__antithesis_instrumentation__.Notify(88442)
		}
		__antithesis_instrumentation__.Notify(88440)
		return err
	} else {
		__antithesis_instrumentation__.Notify(88443)
	}
	__antithesis_instrumentation__.Notify(88430)

	tc.mu.active = sp.active

	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88444)
		reqInt.rollbackToSavepointLocked(ctx, *sp)
	}
	__antithesis_instrumentation__.Notify(88431)

	if sp.seqNum < tc.interceptorAlloc.txnSeqNumAllocator.writeSeq {
		__antithesis_instrumentation__.Notify(88445)
		tc.mu.txn.AddIgnoredSeqNumRange(
			enginepb.IgnoredSeqNumRange{
				Start: sp.seqNum + 1, End: tc.interceptorAlloc.txnSeqNumAllocator.writeSeq,
			})
	} else {
		__antithesis_instrumentation__.Notify(88446)
	}
	__antithesis_instrumentation__.Notify(88432)

	return nil
}

func (tc *TxnCoordSender) ReleaseSavepoint(ctx context.Context, s kv.SavepointToken) error {
	__antithesis_instrumentation__.Notify(88447)
	if tc.typ != kv.RootTxn {
		__antithesis_instrumentation__.Notify(88451)
		return errors.AssertionFailedf("cannot release savepoint in non-root txn")
	} else {
		__antithesis_instrumentation__.Notify(88452)
	}
	__antithesis_instrumentation__.Notify(88448)

	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.mu.txnState != txnPending {
		__antithesis_instrumentation__.Notify(88453)
		return ErrSavepointOperationInErrorTxn
	} else {
		__antithesis_instrumentation__.Notify(88454)
	}
	__antithesis_instrumentation__.Notify(88449)

	sp := s.(*savepoint)
	err := tc.checkSavepointLocked(sp)
	if errors.Is(err, errSavepointInvalidAfterTxnRestart) {
		__antithesis_instrumentation__.Notify(88455)
		err = roachpb.NewTransactionRetryWithProtoRefreshError(
			"cannot release savepoint after a transaction restart",
			tc.mu.txn.ID,
			tc.mu.txn,
		)
	} else {
		__antithesis_instrumentation__.Notify(88456)
	}
	__antithesis_instrumentation__.Notify(88450)
	return err
}

type errSavepointOperationInErrorTxn struct{}

var ErrSavepointOperationInErrorTxn error = errSavepointOperationInErrorTxn{}

func (err errSavepointOperationInErrorTxn) Error() string {
	__antithesis_instrumentation__.Notify(88457)
	return "cannot create or release savepoint after an error has occurred"
}

func (tc *TxnCoordSender) assertNotFinalized() error {
	__antithesis_instrumentation__.Notify(88458)
	if tc.mu.txnState == txnFinalized {
		__antithesis_instrumentation__.Notify(88460)
		return errors.AssertionFailedf("operation invalid for finalized txns")
	} else {
		__antithesis_instrumentation__.Notify(88461)
	}
	__antithesis_instrumentation__.Notify(88459)
	return nil
}

var errSavepointInvalidAfterTxnRestart = errors.New("savepoint invalid after transaction restart")

func (tc *TxnCoordSender) checkSavepointLocked(s *savepoint) error {
	__antithesis_instrumentation__.Notify(88462)

	if s.Initial() {
		__antithesis_instrumentation__.Notify(88467)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88468)
	}
	__antithesis_instrumentation__.Notify(88463)
	if s.txnID != tc.mu.txn.ID {
		__antithesis_instrumentation__.Notify(88469)
		return errSavepointInvalidAfterTxnRestart
	} else {
		__antithesis_instrumentation__.Notify(88470)
	}
	__antithesis_instrumentation__.Notify(88464)
	if s.epoch != tc.mu.txn.Epoch {
		__antithesis_instrumentation__.Notify(88471)
		return errSavepointInvalidAfterTxnRestart
	} else {
		__antithesis_instrumentation__.Notify(88472)
	}
	__antithesis_instrumentation__.Notify(88465)

	if s.seqNum < 0 || func() bool {
		__antithesis_instrumentation__.Notify(88473)
		return s.seqNum > tc.interceptorAlloc.txnSeqNumAllocator.writeSeq == true
	}() == true {
		__antithesis_instrumentation__.Notify(88474)
		return errors.AssertionFailedf("invalid savepoint: got %d, expected 0-%d",
			s.seqNum, tc.interceptorAlloc.txnSeqNumAllocator.writeSeq)
	} else {
		__antithesis_instrumentation__.Notify(88475)
	}
	__antithesis_instrumentation__.Notify(88466)

	return nil
}
