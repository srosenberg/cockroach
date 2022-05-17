package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/caller"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errorspb"
	_ "github.com/cockroachdb/errors/extgrpc"
	"github.com/cockroachdb/redact"
)

type ClientVisibleRetryError interface {
	ClientVisibleRetryError()
}

type ClientVisibleAmbiguousError interface {
	ClientVisibleAmbiguousError()
}

func (e *UnhandledRetryableError) Error() string {
	__antithesis_instrumentation__.Notify(164277)
	return e.String()
}

var _ error = &UnhandledRetryableError{}

func (e *UnhandledRetryableError) SafeFormat(s redact.SafePrinter, r rune) {
	__antithesis_instrumentation__.Notify(164278)
	e.PErr.SafeFormat(s, r)
}

func (e *UnhandledRetryableError) String() string {
	__antithesis_instrumentation__.Notify(164279)
	return redact.StringWithoutMarkers(e)
}

type transactionRestartError interface {
	canRestartTransaction() TransactionRestart
}

func ErrorUnexpectedlySet(culprit, response interface{}) error {
	__antithesis_instrumentation__.Notify(164280)
	return errors.AssertionFailedf("error is unexpectedly set, culprit is %T:\n%+v", culprit, response)
}

type ErrorPriority int

const (
	_ ErrorPriority = iota

	ErrorScoreTxnRestart

	ErrorScoreUnambiguousError

	ErrorScoreNonRetriable

	ErrorScoreTxnAbort
)

func ErrPriority(err error) ErrorPriority {
	__antithesis_instrumentation__.Notify(164281)

	var detail ErrorDetailInterface
	switch tErr := err.(type) {
	case nil:
		__antithesis_instrumentation__.Notify(164284)
		return 0
	case ErrorDetailInterface:
		__antithesis_instrumentation__.Notify(164285)
		detail = tErr
	case *internalError:
		__antithesis_instrumentation__.Notify(164286)
		detail = (*Error)(tErr).GetDetail()
	case *UnhandledRetryableError:
		__antithesis_instrumentation__.Notify(164287)
		if _, ok := tErr.PErr.GetDetail().(*TransactionAbortedError); ok {
			__antithesis_instrumentation__.Notify(164289)
			return ErrorScoreTxnAbort
		} else {
			__antithesis_instrumentation__.Notify(164290)
		}
		__antithesis_instrumentation__.Notify(164288)
		return ErrorScoreTxnRestart
	}
	__antithesis_instrumentation__.Notify(164282)

	switch v := detail.(type) {
	case *TransactionRetryWithProtoRefreshError:
		__antithesis_instrumentation__.Notify(164291)
		if v.PrevTxnAborted() {
			__antithesis_instrumentation__.Notify(164294)
			return ErrorScoreTxnAbort
		} else {
			__antithesis_instrumentation__.Notify(164295)
		}
		__antithesis_instrumentation__.Notify(164292)
		return ErrorScoreTxnRestart
	case *ConditionFailedError, *WriteIntentError:
		__antithesis_instrumentation__.Notify(164293)

		return ErrorScoreUnambiguousError
	}
	__antithesis_instrumentation__.Notify(164283)
	return ErrorScoreNonRetriable
}

func NewError(err error) *Error {
	__antithesis_instrumentation__.Notify(164296)
	if err == nil {
		__antithesis_instrumentation__.Notify(164299)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(164300)
	}
	__antithesis_instrumentation__.Notify(164297)
	e := &Error{
		EncodedError: errors.EncodeError(context.Background(), err),
	}

	{
		__antithesis_instrumentation__.Notify(164301)
		if intErr, ok := err.(*internalError); ok {
			__antithesis_instrumentation__.Notify(164302)
			*e = *(*Error)(intErr)
		} else {
			__antithesis_instrumentation__.Notify(164303)
			if detail := ErrorDetailInterface(nil); errors.As(err, &detail) {
				__antithesis_instrumentation__.Notify(164304)
				e.deprecatedMessage = detail.message(e)
				if r, ok := detail.(transactionRestartError); ok {
					__antithesis_instrumentation__.Notify(164306)
					e.deprecatedTransactionRestart = r.canRestartTransaction()
				} else {
					__antithesis_instrumentation__.Notify(164307)
					e.deprecatedTransactionRestart = TransactionRestart_NONE
				}
				__antithesis_instrumentation__.Notify(164305)
				e.deprecatedDetail.MustSetInner(detail)
				e.checkTxnStatusValid()
			} else {
				__antithesis_instrumentation__.Notify(164308)
				e.deprecatedMessage = err.Error()
			}
		}
	}
	__antithesis_instrumentation__.Notify(164298)
	return e
}

func NewErrorWithTxn(err error, txn *Transaction) *Error {
	__antithesis_instrumentation__.Notify(164309)
	e := NewError(err)
	e.SetTxn(txn)
	return e
}

func NewErrorf(format string, a ...interface{}) *Error {
	__antithesis_instrumentation__.Notify(164310)
	err := errors.Newf(format, a...)
	file, line, _ := caller.Lookup(1)
	err = errors.Wrapf(err, "%s:%d", file, line)
	return NewError(err)
}

func (e *Error) SafeFormat(s redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(164311)
	if e == nil {
		__antithesis_instrumentation__.Notify(164314)
		s.Print(nil)
		return
	} else {
		__antithesis_instrumentation__.Notify(164315)
	}
	__antithesis_instrumentation__.Notify(164312)

	if e.EncodedError != (errors.EncodedError{}) {
		__antithesis_instrumentation__.Notify(164316)
		err := errors.DecodeError(context.Background(), e.EncodedError)
		var iface ErrorDetailInterface
		if errors.As(err, &iface) {
			__antithesis_instrumentation__.Notify(164318)

			deprecatedMsg := iface.message(e)
			if deprecatedMsg != err.Error() {
				__antithesis_instrumentation__.Notify(164319)
				s.Print(deprecatedMsg)
				return
			} else {
				__antithesis_instrumentation__.Notify(164320)
			}
		} else {
			__antithesis_instrumentation__.Notify(164321)
		}
		__antithesis_instrumentation__.Notify(164317)
		s.Print(err)
	} else {
		__antithesis_instrumentation__.Notify(164322)

		switch t := e.GetDetail().(type) {
		case nil:
			__antithesis_instrumentation__.Notify(164323)
			s.Print(e.deprecatedMessage)
		default:
			__antithesis_instrumentation__.Notify(164324)

			s.Print(t)
		}
	}
	__antithesis_instrumentation__.Notify(164313)

	if txn := e.GetTxn(); txn != nil {
		__antithesis_instrumentation__.Notify(164325)
		s.SafeString(": ")
		s.Print(txn)
	} else {
		__antithesis_instrumentation__.Notify(164326)
	}
}

func (e *Error) String() string {
	__antithesis_instrumentation__.Notify(164327)
	return redact.StringWithoutMarkers(e)
}

func (e *Error) TransactionRestart() TransactionRestart {
	__antithesis_instrumentation__.Notify(164328)
	if e.EncodedError == (errorspb.EncodedError{}) {
		__antithesis_instrumentation__.Notify(164331)

		return e.deprecatedTransactionRestart
	} else {
		__antithesis_instrumentation__.Notify(164332)
	}
	__antithesis_instrumentation__.Notify(164329)
	var iface transactionRestartError
	if errors.As(errors.DecodeError(context.Background(), e.EncodedError), &iface) {
		__antithesis_instrumentation__.Notify(164333)
		return iface.canRestartTransaction()
	} else {
		__antithesis_instrumentation__.Notify(164334)
	}
	__antithesis_instrumentation__.Notify(164330)
	return TransactionRestart_NONE
}

type internalError Error

func (e *internalError) Error() string {
	__antithesis_instrumentation__.Notify(164335)
	return (*Error)(e).String()
}

type ErrorDetailInterface interface {
	error
	protoutil.Message

	message(*Error) string

	Type() ErrorDetailType
}

type ErrorDetailType int

const (
	NotLeaseHolderErrType                   ErrorDetailType = 1
	RangeNotFoundErrType                    ErrorDetailType = 2
	RangeKeyMismatchErrType                 ErrorDetailType = 3
	ReadWithinUncertaintyIntervalErrType    ErrorDetailType = 4
	TransactionAbortedErrType               ErrorDetailType = 5
	TransactionPushErrType                  ErrorDetailType = 6
	TransactionRetryErrType                 ErrorDetailType = 7
	TransactionStatusErrType                ErrorDetailType = 8
	WriteIntentErrType                      ErrorDetailType = 9
	WriteTooOldErrType                      ErrorDetailType = 10
	OpRequiresTxnErrType                    ErrorDetailType = 11
	ConditionFailedErrType                  ErrorDetailType = 12
	LeaseRejectedErrType                    ErrorDetailType = 13
	NodeUnavailableErrType                  ErrorDetailType = 14
	RaftGroupDeletedErrType                 ErrorDetailType = 16
	ReplicaCorruptionErrType                ErrorDetailType = 17
	ReplicaTooOldErrType                    ErrorDetailType = 18
	AmbiguousResultErrType                  ErrorDetailType = 26
	StoreNotFoundErrType                    ErrorDetailType = 27
	TransactionRetryWithProtoRefreshErrType ErrorDetailType = 28
	IntegerOverflowErrType                  ErrorDetailType = 31
	UnsupportedRequestErrType               ErrorDetailType = 32
	BatchTimestampBeforeGCErrType           ErrorDetailType = 34
	TxnAlreadyEncounteredErrType            ErrorDetailType = 35
	IntentMissingErrType                    ErrorDetailType = 36
	MergeInProgressErrType                  ErrorDetailType = 37
	RangeFeedRetryErrType                   ErrorDetailType = 38
	IndeterminateCommitErrType              ErrorDetailType = 39
	InvalidLeaseErrType                     ErrorDetailType = 40
	OptimisticEvalConflictsErrType          ErrorDetailType = 41
	MinTimestampBoundUnsatisfiableErrType   ErrorDetailType = 42
	RefreshFailedErrType                    ErrorDetailType = 43
	MVCCHistoryMutationErrType              ErrorDetailType = 44

	CommunicationErrType ErrorDetailType = 22

	InternalErrType ErrorDetailType = 25

	NumErrors int = 45
)

func (e *Error) GoError() error {
	__antithesis_instrumentation__.Notify(164336)
	if e == nil {
		__antithesis_instrumentation__.Notify(164341)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(164342)
	}
	__antithesis_instrumentation__.Notify(164337)
	if e.EncodedError != (errorspb.EncodedError{}) {
		__antithesis_instrumentation__.Notify(164343)
		err := errors.DecodeError(context.Background(), e.EncodedError)
		var iface transactionRestartError
		if errors.As(err, &iface) {
			__antithesis_instrumentation__.Notify(164345)
			if txnRestart := iface.canRestartTransaction(); txnRestart != TransactionRestart_NONE {
				__antithesis_instrumentation__.Notify(164346)

				return &UnhandledRetryableError{
					PErr: *e,
				}
			} else {
				__antithesis_instrumentation__.Notify(164347)
			}
		} else {
			__antithesis_instrumentation__.Notify(164348)
		}
		__antithesis_instrumentation__.Notify(164344)
		return err
	} else {
		__antithesis_instrumentation__.Notify(164349)
	}
	__antithesis_instrumentation__.Notify(164338)

	if e.TransactionRestart() != TransactionRestart_NONE {
		__antithesis_instrumentation__.Notify(164350)
		return &UnhandledRetryableError{
			PErr: *e,
		}
	} else {
		__antithesis_instrumentation__.Notify(164351)
	}
	__antithesis_instrumentation__.Notify(164339)
	if detail := e.GetDetail(); detail != nil {
		__antithesis_instrumentation__.Notify(164352)
		return detail
	} else {
		__antithesis_instrumentation__.Notify(164353)
	}
	__antithesis_instrumentation__.Notify(164340)
	return (*internalError)(e)
}

func (e *Error) GetDetail() ErrorDetailInterface {
	__antithesis_instrumentation__.Notify(164354)
	if e == nil {
		__antithesis_instrumentation__.Notify(164357)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(164358)
	}
	__antithesis_instrumentation__.Notify(164355)
	var detail ErrorDetailInterface
	if e.EncodedError != (errorspb.EncodedError{}) {
		__antithesis_instrumentation__.Notify(164359)
		errors.As(errors.DecodeError(context.Background(), e.EncodedError), &detail)
	} else {
		__antithesis_instrumentation__.Notify(164360)

		detail, _ = e.deprecatedDetail.GetInner().(ErrorDetailInterface)
	}
	__antithesis_instrumentation__.Notify(164356)
	return detail
}

func (e *Error) SetTxn(txn *Transaction) {
	__antithesis_instrumentation__.Notify(164361)
	e.UnexposedTxn = nil
	e.UpdateTxn(txn)
}

func (e *Error) UpdateTxn(o *Transaction) {
	__antithesis_instrumentation__.Notify(164362)
	if o == nil {
		__antithesis_instrumentation__.Notify(164366)
		return
	} else {
		__antithesis_instrumentation__.Notify(164367)
	}
	__antithesis_instrumentation__.Notify(164363)
	if e.UnexposedTxn == nil {
		__antithesis_instrumentation__.Notify(164368)
		e.UnexposedTxn = o.Clone()
	} else {
		__antithesis_instrumentation__.Notify(164369)
		e.UnexposedTxn.Update(o)
	}
	__antithesis_instrumentation__.Notify(164364)
	if sErr, ok := e.deprecatedDetail.GetInner().(ErrorDetailInterface); ok {
		__antithesis_instrumentation__.Notify(164370)

		e.deprecatedMessage = sErr.message(e)
	} else {
		__antithesis_instrumentation__.Notify(164371)
	}
	__antithesis_instrumentation__.Notify(164365)
	e.checkTxnStatusValid()
}

func (e *Error) checkTxnStatusValid() {
	__antithesis_instrumentation__.Notify(164372)

	txn := e.UnexposedTxn
	err := e.deprecatedDetail.GetInner()
	if txn == nil {
		__antithesis_instrumentation__.Notify(164376)
		return
	} else {
		__antithesis_instrumentation__.Notify(164377)
	}
	__antithesis_instrumentation__.Notify(164373)
	if e.deprecatedTransactionRestart == TransactionRestart_NONE {
		__antithesis_instrumentation__.Notify(164378)
		return
	} else {
		__antithesis_instrumentation__.Notify(164379)
	}
	__antithesis_instrumentation__.Notify(164374)
	if errors.HasType(err, (*TransactionAbortedError)(nil)) {
		__antithesis_instrumentation__.Notify(164380)
		return
	} else {
		__antithesis_instrumentation__.Notify(164381)
	}
	__antithesis_instrumentation__.Notify(164375)
	if txn.Status.IsFinalized() {
		__antithesis_instrumentation__.Notify(164382)
		log.Fatalf(context.TODO(), "transaction unexpectedly finalized in (%T): %v", err, e)
	} else {
		__antithesis_instrumentation__.Notify(164383)
	}
}

func (e *Error) GetTxn() *Transaction {
	__antithesis_instrumentation__.Notify(164384)
	if e == nil {
		__antithesis_instrumentation__.Notify(164386)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(164387)
	}
	__antithesis_instrumentation__.Notify(164385)
	return e.UnexposedTxn
}

func (e *Error) SetErrorIndex(index int32) {
	__antithesis_instrumentation__.Notify(164388)
	e.Index = &ErrPosition{Index: index}
}

func (e *NodeUnavailableError) Error() string {
	__antithesis_instrumentation__.Notify(164389)
	return e.message(nil)
}

func (e *NodeUnavailableError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164390)
	return NodeUnavailableErrType
}

func (*NodeUnavailableError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164391)
	return "node unavailable; try another peer"
}

var _ ErrorDetailInterface = &NodeUnavailableError{}

func (e *NotLeaseHolderError) Error() string {
	__antithesis_instrumentation__.Notify(164392)
	return e.message(nil)
}

func (e *NotLeaseHolderError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164393)
	return NotLeaseHolderErrType
}

func (e *NotLeaseHolderError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164394)
	var buf strings.Builder
	buf.WriteString("[NotLeaseHolderError] ")
	if e.CustomMsg != "" {
		__antithesis_instrumentation__.Notify(164398)
		buf.WriteString(e.CustomMsg)
		buf.WriteString("; ")
	} else {
		__antithesis_instrumentation__.Notify(164399)
	}
	__antithesis_instrumentation__.Notify(164395)
	fmt.Fprintf(&buf, "r%d: ", e.RangeID)
	if e.Replica != (ReplicaDescriptor{}) {
		__antithesis_instrumentation__.Notify(164400)
		fmt.Fprintf(&buf, "replica %s not lease holder; ", e.Replica)
	} else {
		__antithesis_instrumentation__.Notify(164401)
		fmt.Fprint(&buf, "replica not lease holder; ")
	}
	__antithesis_instrumentation__.Notify(164396)
	if e.LeaseHolder == nil {
		__antithesis_instrumentation__.Notify(164402)
		fmt.Fprint(&buf, "lease holder unknown")
	} else {
		__antithesis_instrumentation__.Notify(164403)
		if e.Lease != nil {
			__antithesis_instrumentation__.Notify(164404)
			fmt.Fprintf(&buf, "current lease is %s", e.Lease)
		} else {
			__antithesis_instrumentation__.Notify(164405)
			fmt.Fprintf(&buf, "replica %s is", *e.LeaseHolder)
		}
	}
	__antithesis_instrumentation__.Notify(164397)
	return buf.String()
}

var _ ErrorDetailInterface = &NotLeaseHolderError{}

func (e *LeaseRejectedError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164406)
	return LeaseRejectedErrType
}

func (e *LeaseRejectedError) Error() string {
	__antithesis_instrumentation__.Notify(164407)
	return e.message(nil)
}

func (e *LeaseRejectedError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164408)
	return fmt.Sprintf("cannot replace lease %s with %s: %s", e.Existing, e.Requested.String(), e.Message)
}

var _ ErrorDetailInterface = &LeaseRejectedError{}

func NewRangeNotFoundError(rangeID RangeID, storeID StoreID) *RangeNotFoundError {
	__antithesis_instrumentation__.Notify(164409)
	return &RangeNotFoundError{
		RangeID: rangeID,
		StoreID: storeID,
	}
}

func (e *RangeNotFoundError) Error() string {
	__antithesis_instrumentation__.Notify(164410)
	return e.message(nil)
}

func (e *RangeNotFoundError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164411)
	msg := fmt.Sprintf("r%d was not found", e.RangeID)
	if e.StoreID != 0 {
		__antithesis_instrumentation__.Notify(164413)
		msg += fmt.Sprintf(" on s%d", e.StoreID)
	} else {
		__antithesis_instrumentation__.Notify(164414)
	}
	__antithesis_instrumentation__.Notify(164412)
	return msg
}

func (e *RangeNotFoundError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164415)
	return RangeNotFoundErrType
}

var _ ErrorDetailInterface = &RangeNotFoundError{}

func IsRangeNotFoundError(err error) bool {
	__antithesis_instrumentation__.Notify(164416)
	return errors.HasType(err, (*RangeNotFoundError)(nil))
}

func NewRangeKeyMismatchErrorWithCTPolicy(
	ctx context.Context,
	start, end Key,
	desc *RangeDescriptor,
	lease *Lease,
	ctPolicy RangeClosedTimestampPolicy,
) *RangeKeyMismatchError {
	__antithesis_instrumentation__.Notify(164417)
	if desc == nil {
		__antithesis_instrumentation__.Notify(164421)
		panic("NewRangeKeyMismatchError with nil descriptor")
	} else {
		__antithesis_instrumentation__.Notify(164422)
	}
	__antithesis_instrumentation__.Notify(164418)
	if !desc.IsInitialized() {
		__antithesis_instrumentation__.Notify(164423)

		panic(fmt.Sprintf("descriptor is not initialized: %+v", desc))
	} else {
		__antithesis_instrumentation__.Notify(164424)
	}
	__antithesis_instrumentation__.Notify(164419)
	var l Lease
	if lease != nil {
		__antithesis_instrumentation__.Notify(164425)

		_, ok := desc.GetReplicaDescriptorByID(lease.Replica.ReplicaID)
		if ok {
			__antithesis_instrumentation__.Notify(164426)
			l = *lease
		} else {
			__antithesis_instrumentation__.Notify(164427)
		}
	} else {
		__antithesis_instrumentation__.Notify(164428)
	}
	__antithesis_instrumentation__.Notify(164420)
	e := &RangeKeyMismatchError{
		RequestStartKey: start,
		RequestEndKey:   end,
	}
	ri := RangeInfo{
		Desc:                  *desc,
		Lease:                 l,
		ClosedTimestampPolicy: ctPolicy,
	}

	e.AppendRangeInfo(ctx, ri)
	return e
}

func NewRangeKeyMismatchError(
	ctx context.Context, start, end Key, desc *RangeDescriptor, lease *Lease,
) *RangeKeyMismatchError {
	__antithesis_instrumentation__.Notify(164429)
	return NewRangeKeyMismatchErrorWithCTPolicy(ctx,
		start,
		end,
		desc,
		lease,
		LAG_BY_CLUSTER_SETTING,
	)
}

func (e *RangeKeyMismatchError) Error() string {
	__antithesis_instrumentation__.Notify(164430)
	return e.message(nil)
}

func (e *RangeKeyMismatchError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164431)
	mr, err := e.MismatchedRange()
	if err != nil {
		__antithesis_instrumentation__.Notify(164433)
		return err.Error()
	} else {
		__antithesis_instrumentation__.Notify(164434)
	}
	__antithesis_instrumentation__.Notify(164432)
	return fmt.Sprintf("key range %s-%s outside of bounds of range %s-%s; suggested ranges: %s",
		e.RequestStartKey, e.RequestEndKey, mr.Desc.StartKey, mr.Desc.EndKey, e.Ranges)
}

func (e *RangeKeyMismatchError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164435)
	return RangeKeyMismatchErrType
}

func (e *RangeKeyMismatchError) MismatchedRange() (RangeInfo, error) {
	__antithesis_instrumentation__.Notify(164436)
	if len(e.Ranges) == 0 {
		__antithesis_instrumentation__.Notify(164438)
		return RangeInfo{}, errors.AssertionFailedf(
			"RangeKeyMismatchError (key range %s-%s) with empty RangeInfo slice", e.RequestStartKey, e.RequestEndKey,
		)
	} else {
		__antithesis_instrumentation__.Notify(164439)
	}
	__antithesis_instrumentation__.Notify(164437)
	return e.Ranges[0], nil
}

func (e *RangeKeyMismatchError) AppendRangeInfo(ctx context.Context, ris ...RangeInfo) {
	__antithesis_instrumentation__.Notify(164440)
	for _, ri := range ris {
		__antithesis_instrumentation__.Notify(164441)
		if !ri.Lease.Empty() {
			__antithesis_instrumentation__.Notify(164443)
			if _, ok := ri.Desc.GetReplicaDescriptorByID(ri.Lease.Replica.ReplicaID); !ok {
				__antithesis_instrumentation__.Notify(164444)
				log.Fatalf(ctx, "lease names missing replica; lease: %s, desc: %s", ri.Lease, ri.Desc)
			} else {
				__antithesis_instrumentation__.Notify(164445)
			}
		} else {
			__antithesis_instrumentation__.Notify(164446)
		}
		__antithesis_instrumentation__.Notify(164442)
		e.Ranges = append(e.Ranges, ri)
	}
}

var _ ErrorDetailInterface = &RangeKeyMismatchError{}

func (e *AmbiguousResultError) ClientVisibleAmbiguousError() {
	__antithesis_instrumentation__.Notify(164447)
}

var _ ErrorDetailInterface = &AmbiguousResultError{}
var _ ClientVisibleAmbiguousError = &AmbiguousResultError{}

func (e *TransactionAbortedError) Error() string {
	__antithesis_instrumentation__.Notify(164448)
	return fmt.Sprintf("TransactionAbortedError(%s)", e.Reason)
}

func (e *TransactionAbortedError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164449)
	return fmt.Sprintf("TransactionAbortedError(%s): %s", e.Reason, pErr.GetTxn())
}

func (*TransactionAbortedError) canRestartTransaction() TransactionRestart {
	__antithesis_instrumentation__.Notify(164450)
	return TransactionRestart_IMMEDIATE
}

func (e *TransactionAbortedError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164451)
	return TransactionAbortedErrType
}

var _ ErrorDetailInterface = &TransactionAbortedError{}
var _ transactionRestartError = &TransactionAbortedError{}

func (e *TransactionRetryWithProtoRefreshError) ClientVisibleRetryError() {
	__antithesis_instrumentation__.Notify(164452)
}

func (e *TransactionRetryWithProtoRefreshError) Error() string {
	__antithesis_instrumentation__.Notify(164453)
	return e.message(nil)
}

func (e *TransactionRetryWithProtoRefreshError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164454)
	return fmt.Sprintf("TransactionRetryWithProtoRefreshError: %s", e.Msg)
}

func (e *TransactionRetryWithProtoRefreshError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164455)
	return TransactionRetryWithProtoRefreshErrType
}

var _ ClientVisibleRetryError = &TransactionRetryWithProtoRefreshError{}
var _ ErrorDetailInterface = &TransactionRetryWithProtoRefreshError{}

func NewTransactionAbortedError(reason TransactionAbortedReason) *TransactionAbortedError {
	__antithesis_instrumentation__.Notify(164456)
	return &TransactionAbortedError{
		Reason: reason,
	}
}

func NewTransactionRetryWithProtoRefreshError(
	msg string, txnID uuid.UUID, txn Transaction,
) *TransactionRetryWithProtoRefreshError {
	__antithesis_instrumentation__.Notify(164457)
	return &TransactionRetryWithProtoRefreshError{
		Msg:         msg,
		TxnID:       txnID,
		Transaction: txn,
	}
}

func (e *TransactionRetryWithProtoRefreshError) PrevTxnAborted() bool {
	__antithesis_instrumentation__.Notify(164458)
	return !e.TxnID.Equal(e.Transaction.ID)
}

func NewTransactionPushError(pusheeTxn Transaction) *TransactionPushError {
	__antithesis_instrumentation__.Notify(164459)

	return &TransactionPushError{PusheeTxn: pusheeTxn}
}

func (e *TransactionPushError) Error() string {
	__antithesis_instrumentation__.Notify(164460)
	return e.message(nil)
}

func (e *TransactionPushError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164461)
	s := fmt.Sprintf("failed to push %s", e.PusheeTxn)
	if pErr.GetTxn() == nil {
		__antithesis_instrumentation__.Notify(164463)
		return s
	} else {
		__antithesis_instrumentation__.Notify(164464)
	}
	__antithesis_instrumentation__.Notify(164462)
	return fmt.Sprintf("txn %s %s", pErr.GetTxn(), s)
}

func (*TransactionPushError) canRestartTransaction() TransactionRestart {
	__antithesis_instrumentation__.Notify(164465)
	return TransactionRestart_IMMEDIATE
}

func (e *TransactionPushError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164466)
	return TransactionPushErrType
}

var _ ErrorDetailInterface = &TransactionPushError{}
var _ transactionRestartError = &TransactionPushError{}

func NewTransactionRetryError(
	reason TransactionRetryReason, extraMsg string,
) *TransactionRetryError {
	__antithesis_instrumentation__.Notify(164467)
	return &TransactionRetryError{
		Reason:   reason,
		ExtraMsg: extraMsg,
	}
}

func (e *TransactionRetryError) Error() string {
	__antithesis_instrumentation__.Notify(164468)
	msg := ""
	if e.ExtraMsg != "" {
		__antithesis_instrumentation__.Notify(164470)
		msg = " - " + e.ExtraMsg
	} else {
		__antithesis_instrumentation__.Notify(164471)
	}
	__antithesis_instrumentation__.Notify(164469)
	return fmt.Sprintf("TransactionRetryError: retry txn (%s%s)", e.Reason, msg)
}

func (e *TransactionRetryError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164472)
	return fmt.Sprintf("%s: %s", e.Error(), pErr.GetTxn())
}

func (e *TransactionRetryError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164473)
	return TransactionRetryErrType
}

func (*TransactionRetryError) canRestartTransaction() TransactionRestart {
	__antithesis_instrumentation__.Notify(164474)
	return TransactionRestart_IMMEDIATE
}

var _ ErrorDetailInterface = &TransactionRetryError{}
var _ transactionRestartError = &TransactionRetryError{}

func NewTransactionStatusError(
	reason TransactionStatusError_Reason, msg string,
) *TransactionStatusError {
	__antithesis_instrumentation__.Notify(164475)
	return &TransactionStatusError{
		Msg:    msg,
		Reason: reason,
	}
}

func (e *TransactionStatusError) Error() string {
	__antithesis_instrumentation__.Notify(164476)
	return fmt.Sprintf("TransactionStatusError: %s (%s)", e.Msg, e.Reason)
}

func (e *TransactionStatusError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164477)
	return TransactionStatusErrType
}

func (e *TransactionStatusError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164478)
	return fmt.Sprintf("%s: %s", e.Error(), pErr.GetTxn())
}

var _ ErrorDetailInterface = &TransactionStatusError{}

func (e *WriteIntentError) Error() string {
	__antithesis_instrumentation__.Notify(164479)
	return e.message(nil)
}

func (e *WriteIntentError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164480)
	var buf strings.Builder
	buf.WriteString("conflicting intents on ")

	const maxBegin = 5
	const maxEnd = 5
	var begin, end []Intent
	if len(e.Intents) <= maxBegin+maxEnd {
		__antithesis_instrumentation__.Notify(164485)
		begin = e.Intents
	} else {
		__antithesis_instrumentation__.Notify(164486)
		begin = e.Intents[0:maxBegin]
		end = e.Intents[len(e.Intents)-maxEnd : len(e.Intents)]
	}
	__antithesis_instrumentation__.Notify(164481)

	for i := range begin {
		__antithesis_instrumentation__.Notify(164487)
		if i > 0 {
			__antithesis_instrumentation__.Notify(164489)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(164490)
		}
		__antithesis_instrumentation__.Notify(164488)
		buf.WriteString(begin[i].Key.String())
	}
	__antithesis_instrumentation__.Notify(164482)
	if end != nil {
		__antithesis_instrumentation__.Notify(164491)
		buf.WriteString(" ... ")
		for i := range end {
			__antithesis_instrumentation__.Notify(164492)
			if i > 0 {
				__antithesis_instrumentation__.Notify(164494)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(164495)
			}
			__antithesis_instrumentation__.Notify(164493)
			buf.WriteString(end[i].Key.String())
		}
	} else {
		__antithesis_instrumentation__.Notify(164496)
	}
	__antithesis_instrumentation__.Notify(164483)

	switch e.Reason {
	case WriteIntentError_REASON_UNSPECIFIED:
		__antithesis_instrumentation__.Notify(164497)

	case WriteIntentError_REASON_WAIT_POLICY:
		__antithesis_instrumentation__.Notify(164498)
		buf.WriteString(" [reason=wait_policy]")
	case WriteIntentError_REASON_LOCK_TIMEOUT:
		__antithesis_instrumentation__.Notify(164499)
		buf.WriteString(" [reason=lock_timeout]")
	case WriteIntentError_REASON_LOCK_WAIT_QUEUE_MAX_LENGTH_EXCEEDED:
		__antithesis_instrumentation__.Notify(164500)
		buf.WriteString(" [reason=lock_wait_queue_max_length_exceeded]")
	default:
		__antithesis_instrumentation__.Notify(164501)

	}
	__antithesis_instrumentation__.Notify(164484)
	return buf.String()
}

func (e *WriteIntentError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164502)
	return WriteIntentErrType
}

var _ ErrorDetailInterface = &WriteIntentError{}

func NewWriteTooOldError(operationTS, actualTS hlc.Timestamp, key Key) *WriteTooOldError {
	__antithesis_instrumentation__.Notify(164503)
	if len(key) > 0 {
		__antithesis_instrumentation__.Notify(164505)
		oldKey := key
		key = make([]byte, len(oldKey))
		copy(key, oldKey)
	} else {
		__antithesis_instrumentation__.Notify(164506)
	}
	__antithesis_instrumentation__.Notify(164504)
	return &WriteTooOldError{
		Timestamp:       operationTS,
		ActualTimestamp: actualTS,
		Key:             key,
	}
}

func (e *WriteTooOldError) Error() string {
	__antithesis_instrumentation__.Notify(164507)
	return e.message(nil)
}

func (e *WriteTooOldError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164508)
	if len(e.Key) > 0 {
		__antithesis_instrumentation__.Notify(164510)
		return fmt.Sprintf("WriteTooOldError: write for key %s at timestamp %s too old; wrote at %s",
			e.Key, e.Timestamp, e.ActualTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(164511)
	}
	__antithesis_instrumentation__.Notify(164509)
	return fmt.Sprintf("WriteTooOldError: write at timestamp %s too old; wrote at %s",
		e.Timestamp, e.ActualTimestamp)
}

func (*WriteTooOldError) canRestartTransaction() TransactionRestart {
	__antithesis_instrumentation__.Notify(164512)
	return TransactionRestart_IMMEDIATE
}

func (e *WriteTooOldError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164513)
	return WriteTooOldErrType
}

func (e *WriteTooOldError) RetryTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(164514)
	return e.ActualTimestamp
}

var _ ErrorDetailInterface = &WriteTooOldError{}
var _ transactionRestartError = &WriteTooOldError{}

func NewReadWithinUncertaintyIntervalError(
	readTS, existingTS, localUncertaintyLimit hlc.Timestamp, txn *Transaction,
) *ReadWithinUncertaintyIntervalError {
	__antithesis_instrumentation__.Notify(164515)
	rwue := &ReadWithinUncertaintyIntervalError{
		ReadTimestamp:         readTS,
		ExistingTimestamp:     existingTS,
		LocalUncertaintyLimit: localUncertaintyLimit,
	}
	if txn != nil {
		__antithesis_instrumentation__.Notify(164517)
		rwue.GlobalUncertaintyLimit = txn.GlobalUncertaintyLimit
		rwue.ObservedTimestamps = txn.ObservedTimestamps
	} else {
		__antithesis_instrumentation__.Notify(164518)
	}
	__antithesis_instrumentation__.Notify(164516)
	return rwue
}

func (e *ReadWithinUncertaintyIntervalError) SafeFormat(s redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(164519)
	s.Printf("ReadWithinUncertaintyIntervalError: read at time %s encountered "+
		"previous write with future timestamp %s within uncertainty interval `t <= "+
		"(local=%v, global=%v)`; "+
		"observed timestamps: ",
		e.ReadTimestamp, e.ExistingTimestamp, e.LocalUncertaintyLimit, e.GlobalUncertaintyLimit)

	s.SafeRune('[')
	for i, ot := range observedTimestampSlice(e.ObservedTimestamps) {
		__antithesis_instrumentation__.Notify(164521)
		if i > 0 {
			__antithesis_instrumentation__.Notify(164523)
			s.SafeRune(' ')
		} else {
			__antithesis_instrumentation__.Notify(164524)
		}
		__antithesis_instrumentation__.Notify(164522)
		s.Printf("{%d %v}", ot.NodeID, ot.Timestamp)
	}
	__antithesis_instrumentation__.Notify(164520)
	s.SafeRune(']')
}

func (e *ReadWithinUncertaintyIntervalError) String() string {
	__antithesis_instrumentation__.Notify(164525)
	return redact.StringWithoutMarkers(e)
}

func (e *ReadWithinUncertaintyIntervalError) Error() string {
	__antithesis_instrumentation__.Notify(164526)
	return e.String()
}

func (e *ReadWithinUncertaintyIntervalError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164527)
	return e.String()
}

func (e *ReadWithinUncertaintyIntervalError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164528)
	return ReadWithinUncertaintyIntervalErrType
}

func (*ReadWithinUncertaintyIntervalError) canRestartTransaction() TransactionRestart {
	__antithesis_instrumentation__.Notify(164529)
	return TransactionRestart_IMMEDIATE
}

func (e *ReadWithinUncertaintyIntervalError) RetryTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(164530)

	ts := e.ExistingTimestamp.Next()

	ts.Forward(e.LocalUncertaintyLimit)
	return ts
}

var _ ErrorDetailInterface = &ReadWithinUncertaintyIntervalError{}
var _ transactionRestartError = &ReadWithinUncertaintyIntervalError{}

func (e *OpRequiresTxnError) Error() string {
	__antithesis_instrumentation__.Notify(164531)
	return e.message(nil)
}

func (e *OpRequiresTxnError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164532)
	return "the operation requires transactional context"
}

func (e *OpRequiresTxnError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164533)
	return OpRequiresTxnErrType
}

var _ ErrorDetailInterface = &OpRequiresTxnError{}

func (e *ConditionFailedError) Error() string {
	__antithesis_instrumentation__.Notify(164534)
	return e.message(nil)
}

func (e *ConditionFailedError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164535)
	return fmt.Sprintf("unexpected value: %s", e.ActualValue)
}

func (e *ConditionFailedError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164536)
	return ConditionFailedErrType
}

var _ ErrorDetailInterface = &ConditionFailedError{}

func (e *RaftGroupDeletedError) Error() string {
	__antithesis_instrumentation__.Notify(164537)
	return e.message(nil)
}

func (*RaftGroupDeletedError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164538)
	return "raft group deleted"
}

func (e *RaftGroupDeletedError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164539)
	return RaftGroupDeletedErrType
}

var _ ErrorDetailInterface = &RaftGroupDeletedError{}

func NewReplicaCorruptionError(err error) *ReplicaCorruptionError {
	__antithesis_instrumentation__.Notify(164540)
	return &ReplicaCorruptionError{ErrorMsg: err.Error()}
}

func (e *ReplicaCorruptionError) Error() string {
	__antithesis_instrumentation__.Notify(164541)
	return e.message(nil)
}

func (e *ReplicaCorruptionError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164542)
	msg := fmt.Sprintf("replica corruption (processed=%t)", e.Processed)
	if e.ErrorMsg != "" {
		__antithesis_instrumentation__.Notify(164544)
		msg += ": " + e.ErrorMsg
	} else {
		__antithesis_instrumentation__.Notify(164545)
	}
	__antithesis_instrumentation__.Notify(164543)
	return msg
}

func (e *ReplicaCorruptionError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164546)
	return ReplicaCorruptionErrType
}

var _ ErrorDetailInterface = &ReplicaCorruptionError{}

func NewReplicaTooOldError(replicaID ReplicaID) *ReplicaTooOldError {
	__antithesis_instrumentation__.Notify(164547)
	return &ReplicaTooOldError{
		ReplicaID: replicaID,
	}
}

func (e *ReplicaTooOldError) Error() string {
	__antithesis_instrumentation__.Notify(164548)
	return e.message(nil)
}

func (*ReplicaTooOldError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164549)
	return "sender replica too old, discarding message"
}

func (e *ReplicaTooOldError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164550)
	return ReplicaTooOldErrType
}

var _ ErrorDetailInterface = &ReplicaTooOldError{}

func NewStoreNotFoundError(storeID StoreID) *StoreNotFoundError {
	__antithesis_instrumentation__.Notify(164551)
	return &StoreNotFoundError{
		StoreID: storeID,
	}
}

func (e *StoreNotFoundError) Error() string {
	__antithesis_instrumentation__.Notify(164552)
	return e.message(nil)
}

func (e *StoreNotFoundError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164553)
	return fmt.Sprintf("store %d was not found", e.StoreID)
}

func (e *StoreNotFoundError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164554)
	return StoreNotFoundErrType
}

var _ ErrorDetailInterface = &StoreNotFoundError{}

func (e *TxnAlreadyEncounteredErrorError) Error() string {
	__antithesis_instrumentation__.Notify(164555)
	return e.message(nil)
}

func (e *TxnAlreadyEncounteredErrorError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164556)
	return fmt.Sprintf(
		"txn already encountered an error; cannot be used anymore (previous err: %s)",
		e.PrevError,
	)
}

func (e *TxnAlreadyEncounteredErrorError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164557)
	return TxnAlreadyEncounteredErrType
}

var _ ErrorDetailInterface = &TxnAlreadyEncounteredErrorError{}

func (e *IntegerOverflowError) Error() string {
	__antithesis_instrumentation__.Notify(164558)
	return e.message(nil)
}

func (e *IntegerOverflowError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164559)
	return fmt.Sprintf(
		"key %s with value %d incremented by %d results in overflow",
		e.Key, e.CurrentValue, e.IncrementValue)
}

func (e *IntegerOverflowError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164560)
	return IntegerOverflowErrType
}

var _ ErrorDetailInterface = &IntegerOverflowError{}

func (e *UnsupportedRequestError) Error() string {
	__antithesis_instrumentation__.Notify(164561)
	return e.message(nil)
}

func (e *UnsupportedRequestError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164562)
	return "unsupported request"
}

func (e *UnsupportedRequestError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164563)
	return UnsupportedRequestErrType
}

var _ ErrorDetailInterface = &UnsupportedRequestError{}

func (e *BatchTimestampBeforeGCError) Error() string {
	__antithesis_instrumentation__.Notify(164564)
	return e.message(nil)
}

func (e *BatchTimestampBeforeGCError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164565)
	return fmt.Sprintf("batch timestamp %v must be after replica GC threshold %v", e.Timestamp, e.Threshold)
}

func (e *BatchTimestampBeforeGCError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164566)
	return BatchTimestampBeforeGCErrType
}

var _ ErrorDetailInterface = &BatchTimestampBeforeGCError{}

func (e *MVCCHistoryMutationError) Error() string {
	__antithesis_instrumentation__.Notify(164567)
	return e.message(nil)
}

func (e *MVCCHistoryMutationError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164568)
	return fmt.Sprintf("unexpected MVCC history mutation in span %s", e.Span)
}

func (e *MVCCHistoryMutationError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164569)
	return MVCCHistoryMutationErrType
}

var _ ErrorDetailInterface = &MVCCHistoryMutationError{}

func NewIntentMissingError(key Key, wrongIntent *Intent) *IntentMissingError {
	__antithesis_instrumentation__.Notify(164570)
	return &IntentMissingError{
		Key:         key,
		WrongIntent: wrongIntent,
	}
}

func (e *IntentMissingError) Error() string {
	__antithesis_instrumentation__.Notify(164571)
	return e.message(nil)
}

func (e *IntentMissingError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164572)
	var detail string
	if e.WrongIntent != nil {
		__antithesis_instrumentation__.Notify(164574)
		detail = fmt.Sprintf("; found intent %v at key instead", e.WrongIntent)
	} else {
		__antithesis_instrumentation__.Notify(164575)
	}
	__antithesis_instrumentation__.Notify(164573)
	return fmt.Sprintf("intent missing%s", detail)
}

func (e *IntentMissingError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164576)
	return IntentMissingErrType
}

func (*IntentMissingError) canRestartTransaction() TransactionRestart {
	__antithesis_instrumentation__.Notify(164577)
	return TransactionRestart_IMMEDIATE
}

var _ ErrorDetailInterface = &IntentMissingError{}
var _ transactionRestartError = &IntentMissingError{}

func (e *MergeInProgressError) Error() string {
	__antithesis_instrumentation__.Notify(164578)
	return e.message(nil)
}

func (e *MergeInProgressError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164579)
	return "merge in progress"
}

func (e *MergeInProgressError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164580)
	return MergeInProgressErrType
}

var _ ErrorDetailInterface = &MergeInProgressError{}

func NewRangeFeedRetryError(reason RangeFeedRetryError_Reason) *RangeFeedRetryError {
	__antithesis_instrumentation__.Notify(164581)
	return &RangeFeedRetryError{
		Reason: reason,
	}
}

func (e *RangeFeedRetryError) Error() string {
	__antithesis_instrumentation__.Notify(164582)
	return e.message(nil)
}

func (e *RangeFeedRetryError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164583)
	return fmt.Sprintf("retry rangefeed (%s)", e.Reason)
}

func (e *RangeFeedRetryError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164584)
	return RangeFeedRetryErrType
}

var _ ErrorDetailInterface = &RangeFeedRetryError{}

func NewIndeterminateCommitError(txn Transaction) *IndeterminateCommitError {
	__antithesis_instrumentation__.Notify(164585)
	return &IndeterminateCommitError{StagingTxn: txn}
}

func (e *IndeterminateCommitError) Error() string {
	__antithesis_instrumentation__.Notify(164586)
	return e.message(nil)
}

func (e *IndeterminateCommitError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164587)
	s := fmt.Sprintf("found txn in indeterminate STAGING state %s", e.StagingTxn)
	if pErr.GetTxn() == nil {
		__antithesis_instrumentation__.Notify(164589)
		return s
	} else {
		__antithesis_instrumentation__.Notify(164590)
	}
	__antithesis_instrumentation__.Notify(164588)
	return fmt.Sprintf("txn %s %s", pErr.GetTxn(), s)
}

func (e *IndeterminateCommitError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164591)
	return IndeterminateCommitErrType
}

var _ ErrorDetailInterface = &IndeterminateCommitError{}

func (e *InvalidLeaseError) Error() string {
	__antithesis_instrumentation__.Notify(164592)
	return e.message(nil)
}

func (e *InvalidLeaseError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164593)
	return "invalid lease"
}

func (e *InvalidLeaseError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164594)
	return InvalidLeaseErrType
}

var _ ErrorDetailInterface = &InvalidLeaseError{}

func NewOptimisticEvalConflictsError() *OptimisticEvalConflictsError {
	__antithesis_instrumentation__.Notify(164595)
	return &OptimisticEvalConflictsError{}
}

func (e *OptimisticEvalConflictsError) Error() string {
	__antithesis_instrumentation__.Notify(164596)
	return e.message(nil)
}

func (e *OptimisticEvalConflictsError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164597)
	return "optimistic eval encountered conflict"
}

func (e *OptimisticEvalConflictsError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164598)
	return OptimisticEvalConflictsErrType
}

var _ ErrorDetailInterface = &OptimisticEvalConflictsError{}

func NewMinTimestampBoundUnsatisfiableError(
	minTimestampBound, resolvedTimestamp hlc.Timestamp,
) *MinTimestampBoundUnsatisfiableError {
	__antithesis_instrumentation__.Notify(164599)
	return &MinTimestampBoundUnsatisfiableError{
		MinTimestampBound: minTimestampBound,
		ResolvedTimestamp: resolvedTimestamp,
	}
}

func (e *MinTimestampBoundUnsatisfiableError) Error() string {
	__antithesis_instrumentation__.Notify(164600)
	return e.message(nil)
}

func (e *MinTimestampBoundUnsatisfiableError) message(pErr *Error) string {
	__antithesis_instrumentation__.Notify(164601)
	return fmt.Sprintf("bounded staleness read with minimum timestamp "+
		"bound of %s could not be satisfied by a local resolved timestamp of %s",
		e.MinTimestampBound, e.ResolvedTimestamp)
}

func (e *MinTimestampBoundUnsatisfiableError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164602)
	return MinTimestampBoundUnsatisfiableErrType
}

var _ ErrorDetailInterface = &MinTimestampBoundUnsatisfiableError{}

func NewRefreshFailedError(
	reason RefreshFailedError_Reason, key Key, ts hlc.Timestamp,
) *RefreshFailedError {
	__antithesis_instrumentation__.Notify(164603)
	return &RefreshFailedError{
		Reason:    reason,
		Key:       key,
		Timestamp: ts,
	}
}

func (e *RefreshFailedError) Error() string {
	__antithesis_instrumentation__.Notify(164604)
	return e.message(nil)
}

func (e *RefreshFailedError) FailureReason() string {
	__antithesis_instrumentation__.Notify(164605)
	var r string
	switch e.Reason {
	case RefreshFailedError_REASON_COMMITTED_VALUE:
		__antithesis_instrumentation__.Notify(164607)
		r = "committed value"
	case RefreshFailedError_REASON_INTENT:
		__antithesis_instrumentation__.Notify(164608)
		r = "intent"
	default:
		__antithesis_instrumentation__.Notify(164609)
		r = "UNKNOWN"
	}
	__antithesis_instrumentation__.Notify(164606)
	return r
}

func (e *RefreshFailedError) message(_ *Error) string {
	__antithesis_instrumentation__.Notify(164610)
	return fmt.Sprintf("encountered recently written %s %s @%s", e.FailureReason(), e.Key, e.Timestamp)
}

func (e *RefreshFailedError) Type() ErrorDetailType {
	__antithesis_instrumentation__.Notify(164611)
	return RefreshFailedErrType
}

var _ ErrorDetailInterface = &RefreshFailedError{}

func (e *InsufficientSpaceError) Error() string {
	__antithesis_instrumentation__.Notify(164612)
	return fmt.Sprintf("store %d has insufficient remaining capacity to %s (remaining: %s / %.1f%%, min required: %.1f%%)",
		e.StoreID, e.Op, humanizeutil.IBytes(e.Available), float64(e.Available)/float64(e.Capacity)*100, e.Required*100)
}
