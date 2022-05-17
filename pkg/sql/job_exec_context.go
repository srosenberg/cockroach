package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
)

type plannerJobExecContext struct {
	p *planner
}

func MakeJobExecContext(
	opName string, user security.SQLUsername, memMetrics *MemoryMetrics, execCfg *ExecutorConfig,
) (JobExecContext, func()) {
	__antithesis_instrumentation__.Notify(498787)
	plannerInterface, close := NewInternalPlanner(
		opName,
		nil,
		user,
		memMetrics,
		execCfg,
		sessiondatapb.SessionData{},
	)
	p := plannerInterface.(*planner)
	return &plannerJobExecContext{p: p}, close
}

func (e *plannerJobExecContext) SemaCtx() *tree.SemaContext {
	__antithesis_instrumentation__.Notify(498788)
	return e.p.SemaCtx()
}
func (e *plannerJobExecContext) ExtendedEvalContext() *extendedEvalContext {
	__antithesis_instrumentation__.Notify(498789)
	return e.p.ExtendedEvalContext()
}
func (e *plannerJobExecContext) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(498790)
	return e.p.SessionData()
}
func (e *plannerJobExecContext) SessionDataMutatorIterator() *sessionDataMutatorIterator {
	__antithesis_instrumentation__.Notify(498791)
	return e.p.SessionDataMutatorIterator()
}
func (e *plannerJobExecContext) ExecCfg() *ExecutorConfig {
	__antithesis_instrumentation__.Notify(498792)
	return e.p.ExecCfg()
}
func (e *plannerJobExecContext) DistSQLPlanner() *DistSQLPlanner {
	__antithesis_instrumentation__.Notify(498793)
	return e.p.DistSQLPlanner()
}
func (e *plannerJobExecContext) LeaseMgr() *lease.Manager {
	__antithesis_instrumentation__.Notify(498794)
	return e.p.LeaseMgr()
}
func (e *plannerJobExecContext) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(498795)
	return e.p.User()
}
func (e *plannerJobExecContext) MigrationJobDeps() migration.JobDeps {
	__antithesis_instrumentation__.Notify(498796)
	return e.p.MigrationJobDeps()
}
func (e *plannerJobExecContext) SpanConfigReconciler() spanconfig.Reconciler {
	__antithesis_instrumentation__.Notify(498797)
	return e.p.SpanConfigReconciler()
}

type JobExecContext interface {
	SemaCtx() *tree.SemaContext
	ExtendedEvalContext() *extendedEvalContext
	SessionData() *sessiondata.SessionData
	SessionDataMutatorIterator() *sessionDataMutatorIterator
	ExecCfg() *ExecutorConfig
	DistSQLPlanner() *DistSQLPlanner
	LeaseMgr() *lease.Manager
	User() security.SQLUsername
	MigrationJobDeps() migration.JobDeps
	SpanConfigReconciler() spanconfig.Reconciler
}
