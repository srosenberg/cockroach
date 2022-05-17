package sqllivenesstestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type alwaysAliveSession string

func NewAlwaysAliveSession(name string) sqlliveness.Session {
	__antithesis_instrumentation__.Notify(624413)
	return alwaysAliveSession(name)
}

func (f alwaysAliveSession) ID() sqlliveness.SessionID {
	__antithesis_instrumentation__.Notify(624414)
	return sqlliveness.SessionID(f)
}

func (f alwaysAliveSession) Expiration() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(624415)
	return hlc.MaxTimestamp
}

func (f alwaysAliveSession) Start() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(624416)
	return hlc.MinTimestamp
}

func (f alwaysAliveSession) RegisterCallbackForSessionExpiry(func(context.Context)) {
	__antithesis_instrumentation__.Notify(624417)
}
