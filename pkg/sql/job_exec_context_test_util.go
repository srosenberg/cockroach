package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type FakeJobExecContext struct {
	JobExecContext
	ExecutorConfig *ExecutorConfig
}

func (p *FakeJobExecContext) ExecCfg() *ExecutorConfig {
	__antithesis_instrumentation__.Notify(498798)
	return p.ExecutorConfig
}

func (p *FakeJobExecContext) SemaCtx() *tree.SemaContext {
	__antithesis_instrumentation__.Notify(498799)
	return nil
}

func (p *FakeJobExecContext) ExtendedEvalContext() *extendedEvalContext {
	__antithesis_instrumentation__.Notify(498800)
	panic("unimplemented")
}

func (p *FakeJobExecContext) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(498801)
	return nil
}

func (p *FakeJobExecContext) SessionDataMutatorIterator() *sessionDataMutatorIterator {
	__antithesis_instrumentation__.Notify(498802)
	panic("unimplemented")
}

func (p *FakeJobExecContext) DistSQLPlanner() *DistSQLPlanner {
	__antithesis_instrumentation__.Notify(498803)
	panic("unimplemented")
}

func (p *FakeJobExecContext) LeaseMgr() *lease.Manager {
	__antithesis_instrumentation__.Notify(498804)
	panic("unimplemented")
}

func (p *FakeJobExecContext) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(498805)
	panic("unimplemented")
}

func (p *FakeJobExecContext) MigrationJobDeps() migration.JobDeps {
	__antithesis_instrumentation__.Notify(498806)
	panic("unimplemented")
}

func (p *FakeJobExecContext) SpanConfigReconciler() spanconfig.Reconciler {
	__antithesis_instrumentation__.Notify(498807)
	panic("unimplemented")
}
