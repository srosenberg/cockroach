package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/errors"
)

type ListenError struct {
	cause error
	Addr  string
}

func (l *ListenError) Error() string {
	__antithesis_instrumentation__.Notify(194227)
	return l.cause.Error()
}

func (l *ListenError) Unwrap() error { __antithesis_instrumentation__.Notify(194228); return l.cause }

func ListenAndUpdateAddrs(
	ctx context.Context, addr, advertiseAddr *string, connName string,
) (net.Listener, error) {
	__antithesis_instrumentation__.Notify(194229)
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(194232)
		return nil, &ListenError{
			cause: err,
			Addr:  *addr,
		}
	} else {
		__antithesis_instrumentation__.Notify(194233)
	}
	__antithesis_instrumentation__.Notify(194230)
	if err := base.UpdateAddrs(ctx, addr, advertiseAddr, ln.Addr()); err != nil {
		__antithesis_instrumentation__.Notify(194234)
		return nil, errors.Wrapf(err, "internal error: cannot parse %s listen address", connName)
	} else {
		__antithesis_instrumentation__.Notify(194235)
	}
	__antithesis_instrumentation__.Notify(194231)
	return ln, nil
}
