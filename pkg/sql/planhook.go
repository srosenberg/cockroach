package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type planHookFn func(
	context.Context, tree.Statement, PlanHookState,
) (fn PlanHookRowFn, header colinfo.ResultColumns, subplans []planNode, avoidBuffering bool, err error)

type PlanHookRowFn func(context.Context, []planNode, chan<- tree.Datums) error

type planHook struct {
	name string
	fn   planHookFn
}

var planHooks []planHook

func (p *planner) RunParams(ctx context.Context) runParams {
	__antithesis_instrumentation__.Notify(563045)
	return runParams{ctx, p.ExtendedEvalContext(), p}
}

type PlanHookState interface {
	resolver.SchemaResolver
	RunParams(ctx context.Context) runParams
	SemaCtx() *tree.SemaContext
	ExtendedEvalContext() *extendedEvalContext
	SessionData() *sessiondata.SessionData
	SessionDataMutatorIterator() *sessionDataMutatorIterator
	ExecCfg() *ExecutorConfig
	DistSQLPlanner() *DistSQLPlanner
	LeaseMgr() *lease.Manager
	TypeAsString(ctx context.Context, e tree.Expr, op string) (func() (string, error), error)
	TypeAsStringArray(ctx context.Context, e tree.Exprs, op string) (func() ([]string, error), error)
	TypeAsStringOpts(
		ctx context.Context, opts tree.KVOptions, optsValidate map[string]KVStringOptValidate,
	) (func() (map[string]string, error), error)
	User() security.SQLUsername
	AuthorizationAccessor

	GetAllRoles(ctx context.Context) (map[security.SQLUsername]bool, error)
	BumpRoleMembershipTableVersion(ctx context.Context) error
	EvalAsOfTimestamp(
		ctx context.Context,
		asOf tree.AsOfClause,
		opts ...tree.EvalAsOfTimestampOption,
	) (tree.AsOfSystemTime, error)
	ResolveMutableTableDescriptor(ctx context.Context, tn *tree.TableName, required bool, requiredType tree.RequiredTableKind) (prefix catalog.ResolvedObjectPrefix, table *tabledesc.Mutable, err error)
	ShowCreate(
		ctx context.Context, dbPrefix string, allDescs []descpb.Descriptor, desc catalog.TableDescriptor, displayOptions ShowCreateDisplayOptions,
	) (string, error)
	CreateSchemaNamespaceEntry(ctx context.Context, schemaNameKey roachpb.Key,
		schemaID descpb.ID) error
	MigrationJobDeps() migration.JobDeps
	SpanConfigReconciler() spanconfig.Reconciler
	BufferClientNotice(ctx context.Context, notice pgnotice.Notice)
}

func AddPlanHook(name string, fn planHookFn) {
	__antithesis_instrumentation__.Notify(563046)
	planHooks = append(planHooks, planHook{name: name, fn: fn})
}

func ClearPlanHooks() {
	__antithesis_instrumentation__.Notify(563047)
	planHooks = nil
}

type hookFnNode struct {
	optColumnsSlot

	name     string
	f        PlanHookRowFn
	header   colinfo.ResultColumns
	subplans []planNode

	run hookFnRun
}

var _ planNode = &hookFnNode{}

type hookFnRun struct {
	resultsCh chan tree.Datums
	errCh     chan error

	row tree.Datums
}

func newHookFnNode(
	name string, fn PlanHookRowFn, header colinfo.ResultColumns, subplans []planNode,
) *hookFnNode {
	__antithesis_instrumentation__.Notify(563048)
	return &hookFnNode{name: name, f: fn, header: header, subplans: subplans}
}

func (f *hookFnNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(563049)

	f.run.resultsCh = make(chan tree.Datums)
	f.run.errCh = make(chan error)

	subplanCtx, sp := tracing.ChildSpan(params.ctx, f.name)
	go func() {
		__antithesis_instrumentation__.Notify(563051)
		defer sp.Finish()
		err := f.f(subplanCtx, f.subplans, f.run.resultsCh)
		select {
		case <-params.ctx.Done():
			__antithesis_instrumentation__.Notify(563053)
		case f.run.errCh <- err:
			__antithesis_instrumentation__.Notify(563054)
		}
		__antithesis_instrumentation__.Notify(563052)
		close(f.run.errCh)
		close(f.run.resultsCh)
	}()
	__antithesis_instrumentation__.Notify(563050)
	return nil
}

func (f *hookFnNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(563055)
	select {
	case <-params.ctx.Done():
		__antithesis_instrumentation__.Notify(563056)
		return false, params.ctx.Err()
	case err := <-f.run.errCh:
		__antithesis_instrumentation__.Notify(563057)
		return false, err
	case f.run.row = <-f.run.resultsCh:
		__antithesis_instrumentation__.Notify(563058)
		return true, nil
	}
}

func (f *hookFnNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(563059)
	return f.run.row
}

func (f *hookFnNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(563060)
	for _, sub := range f.subplans {
		__antithesis_instrumentation__.Notify(563061)
		sub.Close(ctx)
	}
}
