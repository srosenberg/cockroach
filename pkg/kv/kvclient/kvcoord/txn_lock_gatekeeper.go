package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
)

type lockedSender interface {
	SendLocked(context.Context, roachpb.BatchRequest) (*roachpb.BatchResponse, *roachpb.Error)
}

type txnLockGatekeeper struct {
	wrapped kv.Sender
	mu      sync.Locker

	allowConcurrentRequests bool

	requestInFlight bool
}

func (gs *txnLockGatekeeper) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89235)

	if !gs.allowConcurrentRequests && func() bool {
		__antithesis_instrumentation__.Notify(89237)
		return !ba.IsSingleHeartbeatTxnRequest() == true
	}() == true {
		__antithesis_instrumentation__.Notify(89238)
		if gs.requestInFlight {
			__antithesis_instrumentation__.Notify(89240)
			return nil, roachpb.NewError(
				errors.AssertionFailedf("concurrent txn use detected. ba: %s", ba))
		} else {
			__antithesis_instrumentation__.Notify(89241)
		}
		__antithesis_instrumentation__.Notify(89239)
		gs.requestInFlight = true
		defer func() {
			__antithesis_instrumentation__.Notify(89242)
			gs.requestInFlight = false
		}()
	} else {
		__antithesis_instrumentation__.Notify(89243)
	}
	__antithesis_instrumentation__.Notify(89236)

	gs.mu.Unlock()
	defer gs.mu.Lock()
	return gs.wrapped.Send(ctx, ba)
}
