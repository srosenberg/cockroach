package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/errorspb"
	"github.com/gogo/protobuf/proto"
)

type lockSpansOverBudgetError struct {
	lockSpansBytes int64
	limitBytes     int64
	baSummary      string
	txnDetails     string
}

func newLockSpansOverBudgetError(
	lockSpansBytes, limitBytes int64, ba roachpb.BatchRequest,
) lockSpansOverBudgetError {
	__antithesis_instrumentation__.Notify(87811)
	return lockSpansOverBudgetError{
		lockSpansBytes: lockSpansBytes,
		limitBytes:     limitBytes,
		baSummary:      ba.Summary(),
		txnDetails:     ba.Txn.String(),
	}
}

func (l lockSpansOverBudgetError) Error() string {
	__antithesis_instrumentation__.Notify(87812)
	return fmt.Sprintf("the transaction is locking too many rows and exceeded its lock-tracking memory budget; "+
		"lock spans: %d bytes > budget: %d bytes. Request pushing transaction over the edge: %s. "+
		"Transaction details: %s.", l.lockSpansBytes, l.limitBytes, l.baSummary, l.txnDetails)
}

func encodeLockSpansOverBudgetError(
	_ context.Context, err error,
) (msgPrefix string, safe []string, details proto.Message) {
	__antithesis_instrumentation__.Notify(87813)
	t := err.(lockSpansOverBudgetError)
	details = &errorspb.StringsPayload{
		Details: []string{
			strconv.FormatInt(t.lockSpansBytes, 10), strconv.FormatInt(t.limitBytes, 10),
			t.baSummary, t.txnDetails,
		},
	}
	msgPrefix = "the transaction is locking too many rows"
	return msgPrefix, nil, details
}

func decodeLockSpansOverBudgetError(
	_ context.Context, msgPrefix string, safeDetails []string, payload proto.Message,
) error {
	__antithesis_instrumentation__.Notify(87814)
	m, ok := payload.(*errorspb.StringsPayload)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(87818)
		return len(m.Details) < 4 == true
	}() == true {
		__antithesis_instrumentation__.Notify(87819)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(87820)
	}
	__antithesis_instrumentation__.Notify(87815)
	lockBytes, decodeErr := strconv.ParseInt(m.Details[0], 10, 64)
	if decodeErr != nil {
		__antithesis_instrumentation__.Notify(87821)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(87822)
	}
	__antithesis_instrumentation__.Notify(87816)
	limitBytes, decodeErr := strconv.ParseInt(m.Details[1], 10, 64)
	if decodeErr != nil {
		__antithesis_instrumentation__.Notify(87823)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(87824)
	}
	__antithesis_instrumentation__.Notify(87817)
	return lockSpansOverBudgetError{
		lockSpansBytes: lockBytes,
		limitBytes:     limitBytes,
		baSummary:      m.Details[2],
		txnDetails:     m.Details[3],
	}
}

func init() {
	pKey := errors.GetTypeKey(lockSpansOverBudgetError{})
	errors.RegisterLeafEncoder(pKey, encodeLockSpansOverBudgetError)
	errors.RegisterLeafDecoder(pKey, decodeLockSpansOverBudgetError)
}
