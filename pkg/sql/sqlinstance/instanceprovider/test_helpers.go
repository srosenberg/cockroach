package instanceprovider

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance/instancestorage"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

type TestInstanceProvider interface {
	sqlinstance.Provider
	ShutdownSQLInstanceForTest(context.Context)
}

func NewTestInstanceProvider(
	stopper *stop.Stopper, session sqlliveness.Instance, addr string,
) TestInstanceProvider {
	__antithesis_instrumentation__.Notify(623967)
	storage := instancestorage.NewFakeStorage()
	p := &provider{
		storage:      storage,
		stopper:      stopper,
		session:      session,
		instanceAddr: addr,
		initialized:  make(chan struct{}),
	}
	p.mu.started = true
	return p
}

func (p *provider) ShutdownSQLInstanceForTest(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623968)
	p.shutdownSQLInstance(ctx)
}
