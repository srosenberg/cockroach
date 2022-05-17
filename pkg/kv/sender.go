package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type TxnType int

const (
	_ TxnType = iota

	RootTxn

	LeafTxn
)

type Sender interface {
	Send(context.Context, roachpb.BatchRequest) (*roachpb.BatchResponse, *roachpb.Error)
}

type TxnSender interface {
	Sender

	AnchorOnSystemConfigRange() error

	GetLeafTxnInputState(context.Context, TxnStatusOpt) (*roachpb.LeafTxnInputState, error)

	GetLeafTxnFinalState(context.Context, TxnStatusOpt) (*roachpb.LeafTxnFinalState, error)

	UpdateRootWithLeafFinalState(context.Context, *roachpb.LeafTxnFinalState)

	SetUserPriority(roachpb.UserPriority) error

	SetDebugName(name string)

	String() string

	TxnStatus() roachpb.TransactionStatus

	CreateSavepoint(context.Context) (SavepointToken, error)

	RollbackToSavepoint(context.Context, SavepointToken) error

	ReleaseSavepoint(context.Context, SavepointToken) error

	SetFixedTimestamp(ctx context.Context, ts hlc.Timestamp) error

	ManualRestart(context.Context, roachpb.UserPriority, hlc.Timestamp)

	UpdateStateOnRemoteRetryableErr(context.Context, *roachpb.Error) *roachpb.Error

	DisablePipelining() error

	ReadTimestamp() hlc.Timestamp

	CommitTimestamp() hlc.Timestamp

	CommitTimestampFixed() bool

	ProvisionalCommitTimestamp() hlc.Timestamp

	RequiredFrontier() hlc.Timestamp

	IsSerializablePushAndRefreshNotPossible() bool

	Active() bool

	Epoch() enginepb.TxnEpoch

	IsLocking() bool

	PrepareRetryableError(ctx context.Context, msg string) error

	TestingCloneTxn() *roachpb.Transaction

	Step(context.Context) error

	SetReadSeqNum(seq enginepb.TxnSeq) error

	ConfigureStepping(ctx context.Context, mode SteppingMode) (prevMode SteppingMode)

	GetSteppingMode(ctx context.Context) (curMode SteppingMode)

	ManualRefresh(ctx context.Context) error

	DeferCommitWait(ctx context.Context) func(context.Context) error

	GetTxnRetryableErr(ctx context.Context) *roachpb.TransactionRetryWithProtoRefreshError

	ClearTxnRetryableErr(ctx context.Context)
}

type SteppingMode bool

const (
	SteppingDisabled SteppingMode = false

	SteppingEnabled SteppingMode = true
)

type SavepointToken interface {
	Initial() bool
}

type TxnStatusOpt int

const (
	AnyTxnStatus TxnStatusOpt = iota

	OnlyPending
)

type TxnSenderFactory interface {
	RootTransactionalSender(
		txn *roachpb.Transaction, pri roachpb.UserPriority,
	) TxnSender

	LeafTransactionalSender(tis *roachpb.LeafTxnInputState) TxnSender

	NonTransactionalSender() Sender
}

type SenderFunc func(context.Context, roachpb.BatchRequest) (*roachpb.BatchResponse, *roachpb.Error)

func (f SenderFunc) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127709)
	return f(ctx, ba)
}

type NonTransactionalFactoryFunc SenderFunc

var _ TxnSenderFactory = NonTransactionalFactoryFunc(nil)

func (f NonTransactionalFactoryFunc) RootTransactionalSender(
	_ *roachpb.Transaction, _ roachpb.UserPriority,
) TxnSender {
	__antithesis_instrumentation__.Notify(127710)
	panic("not supported")
}

func (f NonTransactionalFactoryFunc) LeafTransactionalSender(
	_ *roachpb.LeafTxnInputState,
) TxnSender {
	__antithesis_instrumentation__.Notify(127711)
	panic("not supported")
}

func (f NonTransactionalFactoryFunc) NonTransactionalSender() Sender {
	__antithesis_instrumentation__.Notify(127712)
	return SenderFunc(f)
}

func SendWrappedWith(
	ctx context.Context, sender Sender, h roachpb.Header, args roachpb.Request,
) (roachpb.Response, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127713)
	return SendWrappedWithAdmission(ctx, sender, h, roachpb.AdmissionHeader{}, args)
}

func SendWrappedWithAdmission(
	ctx context.Context,
	sender Sender,
	h roachpb.Header,
	ah roachpb.AdmissionHeader,
	args roachpb.Request,
) (roachpb.Response, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127714)
	ba := roachpb.BatchRequest{}
	ba.Header = h
	ba.AdmissionHeader = ah
	ba.Add(args)

	br, pErr := sender.Send(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(127716)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(127717)
	}
	__antithesis_instrumentation__.Notify(127715)
	unwrappedReply := br.Responses[0].GetInner()
	header := unwrappedReply.Header()
	header.Txn = br.Txn
	unwrappedReply.SetHeader(header)
	return unwrappedReply, nil
}

func SendWrapped(
	ctx context.Context, sender Sender, args roachpb.Request,
) (roachpb.Response, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127718)
	return SendWrappedWith(ctx, sender, roachpb.Header{}, args)
}

func Wrap(sender Sender, f func(roachpb.BatchRequest) roachpb.BatchRequest) Sender {
	__antithesis_instrumentation__.Notify(127719)
	return SenderFunc(func(ctx context.Context, ba roachpb.BatchRequest) (*roachpb.BatchResponse, *roachpb.Error) {
		__antithesis_instrumentation__.Notify(127720)
		return sender.Send(ctx, f(ba))
	})
}
