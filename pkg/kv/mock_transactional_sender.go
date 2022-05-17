package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type MockTransactionalSender struct {
	senderFunc func(
		context.Context, *roachpb.Transaction, roachpb.BatchRequest,
	) (*roachpb.BatchResponse, *roachpb.Error)
	txn roachpb.Transaction
}

func NewMockTransactionalSender(
	f func(
		context.Context, *roachpb.Transaction, roachpb.BatchRequest,
	) (*roachpb.BatchResponse, *roachpb.Error),
	txn *roachpb.Transaction,
) *MockTransactionalSender {
	__antithesis_instrumentation__.Notify(127558)
	return &MockTransactionalSender{senderFunc: f, txn: *txn}
}

func (m *MockTransactionalSender) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127559)
	return m.senderFunc(ctx, &m.txn, ba)
}

func (m *MockTransactionalSender) GetLeafTxnInputState(
	context.Context, TxnStatusOpt,
) (*roachpb.LeafTxnInputState, error) {
	__antithesis_instrumentation__.Notify(127560)
	panic("unimplemented")
}

func (m *MockTransactionalSender) GetLeafTxnFinalState(
	context.Context, TxnStatusOpt,
) (*roachpb.LeafTxnFinalState, error) {
	__antithesis_instrumentation__.Notify(127561)
	panic("unimplemented")
}

func (m *MockTransactionalSender) UpdateRootWithLeafFinalState(
	context.Context, *roachpb.LeafTxnFinalState,
) {
	__antithesis_instrumentation__.Notify(127562)
	panic("unimplemented")
}

func (m *MockTransactionalSender) AnchorOnSystemConfigRange() error {
	__antithesis_instrumentation__.Notify(127563)
	return errors.New("unimplemented")
}

func (m *MockTransactionalSender) TxnStatus() roachpb.TransactionStatus {
	__antithesis_instrumentation__.Notify(127564)
	return m.txn.Status
}

func (m *MockTransactionalSender) SetUserPriority(pri roachpb.UserPriority) error {
	__antithesis_instrumentation__.Notify(127565)
	m.txn.Priority = roachpb.MakePriority(pri)
	return nil
}

func (m *MockTransactionalSender) SetDebugName(name string) {
	__antithesis_instrumentation__.Notify(127566)
	m.txn.Name = name
}

func (m *MockTransactionalSender) String() string {
	__antithesis_instrumentation__.Notify(127567)
	return m.txn.String()
}

func (m *MockTransactionalSender) ReadTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127568)
	return m.txn.ReadTimestamp
}

func (m *MockTransactionalSender) ProvisionalCommitTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127569)
	return m.txn.WriteTimestamp
}

func (m *MockTransactionalSender) CommitTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127570)
	return m.txn.ReadTimestamp
}

func (m *MockTransactionalSender) CommitTimestampFixed() bool {
	__antithesis_instrumentation__.Notify(127571)
	return m.txn.CommitTimestampFixed
}

func (m *MockTransactionalSender) SetFixedTimestamp(_ context.Context, ts hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(127572)
	m.txn.WriteTimestamp = ts
	m.txn.ReadTimestamp = ts
	m.txn.GlobalUncertaintyLimit = ts
	m.txn.CommitTimestampFixed = true

	m.txn.MinTimestamp.Backward(ts)
	return nil
}

func (m *MockTransactionalSender) RequiredFrontier() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127573)
	return m.txn.RequiredFrontier()
}

func (m *MockTransactionalSender) ManualRestart(
	ctx context.Context, pri roachpb.UserPriority, ts hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(127574)
	m.txn.Restart(pri, 0, ts)
}

func (m *MockTransactionalSender) IsSerializablePushAndRefreshNotPossible() bool {
	__antithesis_instrumentation__.Notify(127575)
	return false
}

func (m *MockTransactionalSender) CreateSavepoint(context.Context) (SavepointToken, error) {
	__antithesis_instrumentation__.Notify(127576)
	panic("unimplemented")
}

func (m *MockTransactionalSender) RollbackToSavepoint(context.Context, SavepointToken) error {
	__antithesis_instrumentation__.Notify(127577)
	panic("unimplemented")
}

func (m *MockTransactionalSender) ReleaseSavepoint(context.Context, SavepointToken) error {
	__antithesis_instrumentation__.Notify(127578)
	panic("unimplemented")
}

func (m *MockTransactionalSender) Epoch() enginepb.TxnEpoch {
	__antithesis_instrumentation__.Notify(127579)
	panic("unimplemented")
}

func (m *MockTransactionalSender) IsLocking() bool {
	__antithesis_instrumentation__.Notify(127580)
	return false
}

func (m *MockTransactionalSender) TestingCloneTxn() *roachpb.Transaction {
	__antithesis_instrumentation__.Notify(127581)
	return m.txn.Clone()
}

func (m *MockTransactionalSender) Active() bool {
	__antithesis_instrumentation__.Notify(127582)
	panic("unimplemented")
}

func (m *MockTransactionalSender) UpdateStateOnRemoteRetryableErr(
	ctx context.Context, pErr *roachpb.Error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(127583)
	panic("unimplemented")
}

func (m *MockTransactionalSender) DisablePipelining() error {
	__antithesis_instrumentation__.Notify(127584)
	return nil
}

func (m *MockTransactionalSender) PrepareRetryableError(ctx context.Context, msg string) error {
	__antithesis_instrumentation__.Notify(127585)
	return roachpb.NewTransactionRetryWithProtoRefreshError(msg, m.txn.ID, *m.txn.Clone())
}

func (m *MockTransactionalSender) Step(_ context.Context) error {
	__antithesis_instrumentation__.Notify(127586)

	return nil
}

func (m *MockTransactionalSender) SetReadSeqNum(_ enginepb.TxnSeq) error {
	__antithesis_instrumentation__.Notify(127587)
	return nil
}

func (m *MockTransactionalSender) ConfigureStepping(context.Context, SteppingMode) SteppingMode {
	__antithesis_instrumentation__.Notify(127588)

	return SteppingDisabled
}

func (m *MockTransactionalSender) GetSteppingMode(context.Context) SteppingMode {
	__antithesis_instrumentation__.Notify(127589)
	return SteppingDisabled
}

func (m *MockTransactionalSender) ManualRefresh(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(127590)
	panic("unimplemented")
}

func (m *MockTransactionalSender) DeferCommitWait(ctx context.Context) func(context.Context) error {
	__antithesis_instrumentation__.Notify(127591)
	panic("unimplemented")
}

func (m *MockTransactionalSender) GetTxnRetryableErr(
	ctx context.Context,
) *roachpb.TransactionRetryWithProtoRefreshError {
	__antithesis_instrumentation__.Notify(127592)
	return nil
}

func (m *MockTransactionalSender) ClearTxnRetryableErr(ctx context.Context) {
	__antithesis_instrumentation__.Notify(127593)
}

type MockTxnSenderFactory struct {
	senderFunc func(context.Context, *roachpb.Transaction, roachpb.BatchRequest) (
		*roachpb.BatchResponse, *roachpb.Error)
	nonTxnSenderFunc Sender
}

var _ TxnSenderFactory = MockTxnSenderFactory{}

func MakeMockTxnSenderFactory(
	senderFunc func(
		context.Context, *roachpb.Transaction, roachpb.BatchRequest,
	) (*roachpb.BatchResponse, *roachpb.Error),
) MockTxnSenderFactory {
	__antithesis_instrumentation__.Notify(127594)
	return MockTxnSenderFactory{
		senderFunc: senderFunc,
	}
}

func MakeMockTxnSenderFactoryWithNonTxnSender(
	senderFunc func(
		context.Context, *roachpb.Transaction, roachpb.BatchRequest,
	) (*roachpb.BatchResponse, *roachpb.Error),
	nonTxnSenderFunc SenderFunc,
) MockTxnSenderFactory {
	__antithesis_instrumentation__.Notify(127595)
	return MockTxnSenderFactory{
		senderFunc:       senderFunc,
		nonTxnSenderFunc: nonTxnSenderFunc,
	}
}

func (f MockTxnSenderFactory) RootTransactionalSender(
	txn *roachpb.Transaction, _ roachpb.UserPriority,
) TxnSender {
	__antithesis_instrumentation__.Notify(127596)
	return NewMockTransactionalSender(f.senderFunc, txn)
}

func (f MockTxnSenderFactory) LeafTransactionalSender(tis *roachpb.LeafTxnInputState) TxnSender {
	__antithesis_instrumentation__.Notify(127597)
	return NewMockTransactionalSender(f.senderFunc, &tis.Txn)
}

func (f MockTxnSenderFactory) NonTransactionalSender() Sender {
	__antithesis_instrumentation__.Notify(127598)
	return f.nonTxnSenderFunc
}
