package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"google.golang.org/grpc"
)

type Dialer interface {
	DialNoBreaker(context.Context, roachpb.NodeID, rpc.ConnectionClass) (*grpc.ClientConn, error)
}

func GetConnForOutbox(
	ctx context.Context, dialer Dialer, sqlInstanceID base.SQLInstanceID, timeout time.Duration,
) (conn *grpc.ClientConn, err error) {
	__antithesis_instrumentation__.Notify(471132)
	firstConnectionAttempt := timeutil.Now()
	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(471134)
		conn, err = dialer.DialNoBreaker(ctx, roachpb.NodeID(sqlInstanceID), rpc.DefaultClass)
		if err == nil || func() bool {
			__antithesis_instrumentation__.Notify(471135)
			return timeutil.Since(firstConnectionAttempt) > timeout == true
		}() == true {
			__antithesis_instrumentation__.Notify(471136)
			break
		} else {
			__antithesis_instrumentation__.Notify(471137)
		}
	}
	__antithesis_instrumentation__.Notify(471133)
	return
}
