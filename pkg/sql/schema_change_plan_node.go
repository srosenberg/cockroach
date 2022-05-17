package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/redact"
)

func (p *planner) FormatAstAsRedactableString(
	statement tree.Statement, annotations *tree.Annotations,
) redact.RedactableString {
	__antithesis_instrumentation__.Notify(577017)
	return formatStmtKeyAsRedactableString(p.getVirtualTabler(),
		statement,
		annotations, tree.FmtSimple)
}

func (p *planner) SchemaChange(ctx context.Context, stmt tree.Statement) (planNode, bool, error) {
	__antithesis_instrumentation__.Notify(577018)

	mode := p.extendedEvalCtx.SchemaChangerState.mode

	if mode == sessiondatapb.UseNewSchemaChangerOff || func() bool {
		__antithesis_instrumentation__.Notify(577022)
		return ((mode == sessiondatapb.UseNewSchemaChangerOn || func() bool {
			__antithesis_instrumentation__.Notify(577023)
			return mode == sessiondatapb.UseNewSchemaChangerUnsafe == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(577024)
			return !p.extendedEvalCtx.TxnIsSingleStmt == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(577025)
		return nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(577026)
	}
	__antithesis_instrumentation__.Notify(577019)
	scs := p.extendedEvalCtx.SchemaChangerState
	scs.stmts = append(scs.stmts, p.stmt.SQL)
	deps := scdeps.NewBuilderDependencies(
		p.ExecCfg().LogicalClusterID(),
		p.ExecCfg().Codec,
		p.Txn(),
		p.Descriptors(),
		p,
		p,
		p,
		p,
		p.SessionData(),
		p.ExecCfg().Settings,
		scs.stmts,
	)
	state, err := scbuild.Build(ctx, deps, scs.state, stmt)
	if scerrors.HasNotImplemented(err) && func() bool {
		__antithesis_instrumentation__.Notify(577027)
		return mode != sessiondatapb.UseNewSchemaChangerUnsafeAlways == true
	}() == true {
		__antithesis_instrumentation__.Notify(577028)
		return nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(577029)
	}
	__antithesis_instrumentation__.Notify(577020)
	if err != nil {
		__antithesis_instrumentation__.Notify(577030)

		if scerrors.ConcurrentSchemaChangeDescID(err) != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(577032)
			p.Descriptors().ReleaseLeases(ctx)
		} else {
			__antithesis_instrumentation__.Notify(577033)
		}
		__antithesis_instrumentation__.Notify(577031)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(577034)
	}
	__antithesis_instrumentation__.Notify(577021)
	return &schemaChangePlanNode{plannedState: state}, true, nil
}

func (p *planner) waitForDescriptorSchemaChanges(
	ctx context.Context, descID descpb.ID, scs SchemaChangerState,
) error {
	__antithesis_instrumentation__.Notify(577035)

	if knobs := p.ExecCfg().DeclarativeSchemaChangerTestingKnobs; knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(577038)
		return knobs.BeforeWaitingForConcurrentSchemaChanges != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(577039)
		knobs.BeforeWaitingForConcurrentSchemaChanges(scs.stmts)
	} else {
		__antithesis_instrumentation__.Notify(577040)
	}
	__antithesis_instrumentation__.Notify(577036)

	retryErr := p.txn.PrepareRetryableError(ctx,
		fmt.Sprintf("schema change waiting for concurrent schema changes on descriptor %d", descID))
	p.txn.CleanupOnError(ctx, retryErr)
	p.Descriptors().ReleaseAll(ctx)

	start := timeutil.Now()
	logEvery := log.Every(30 * time.Second)
	for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
		__antithesis_instrumentation__.Notify(577041)
		now := p.ExecCfg().Clock.Now()
		if logEvery.ShouldLog() {
			__antithesis_instrumentation__.Notify(577044)
			log.Infof(ctx,
				"schema change waiting for concurrent schema changes on descriptor %d,"+
					" waited %v so far", descID, timeutil.Since(start),
			)
		} else {
			__antithesis_instrumentation__.Notify(577045)
		}
		__antithesis_instrumentation__.Notify(577042)
		blocked := false
		if err := p.ExecCfg().CollectionFactory.Txn(
			ctx, p.ExecCfg().InternalExecutor, p.ExecCfg().DB,
			func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
				__antithesis_instrumentation__.Notify(577046)
				if err := txn.SetFixedTimestamp(ctx, now); err != nil {
					__antithesis_instrumentation__.Notify(577049)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577050)
				}
				__antithesis_instrumentation__.Notify(577047)
				desc, err := descriptors.GetImmutableDescriptorByID(ctx, txn, descID,
					tree.CommonLookupFlags{
						Required:    true,
						AvoidLeased: true,
					})
				if err != nil {
					__antithesis_instrumentation__.Notify(577051)
					return err
				} else {
					__antithesis_instrumentation__.Notify(577052)
				}
				__antithesis_instrumentation__.Notify(577048)
				blocked = desc.HasConcurrentSchemaChanges()
				return nil
			}); err != nil {
			__antithesis_instrumentation__.Notify(577053)
			return err
		} else {
			__antithesis_instrumentation__.Notify(577054)
		}
		__antithesis_instrumentation__.Notify(577043)
		if !blocked {
			__antithesis_instrumentation__.Notify(577055)
			break
		} else {
			__antithesis_instrumentation__.Notify(577056)
		}
	}
	__antithesis_instrumentation__.Notify(577037)
	log.Infof(
		ctx,
		"done waiting for concurrent schema changes on descriptor %d after %v",
		descID, timeutil.Since(start),
	)
	return nil
}

type schemaChangePlanNode struct {
	plannedState scpb.CurrentState
}

func (s *schemaChangePlanNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(577057)
	p := params.p
	scs := p.ExtendedEvalContext().SchemaChangerState
	runDeps := newSchemaChangerTxnRunDependencies(
		p.SessionData(),
		p.User(),
		p.ExecCfg(),
		p.Txn(),
		p.Descriptors(),
		p.EvalContext(),
		p.ExtendedEvalContext().Tracing.KVTracingEnabled(),
		scs.jobID,
		scs.stmts,
	)
	after, jobID, err := scrun.RunStatementPhase(
		params.ctx, p.ExecCfg().DeclarativeSchemaChangerTestingKnobs, runDeps, s.plannedState,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(577059)
		return err
	} else {
		__antithesis_instrumentation__.Notify(577060)
	}
	__antithesis_instrumentation__.Notify(577058)
	scs.state = after
	scs.jobID = jobID
	return nil
}

func newSchemaChangerTxnRunDependencies(
	sessionData *sessiondata.SessionData,
	user security.SQLUsername,
	execCfg *ExecutorConfig,
	txn *kv.Txn,
	descriptors *descs.Collection,
	evalContext *tree.EvalContext,
	kvTrace bool,
	schemaChangerJobID jobspb.JobID,
	stmts []string,
) scexec.Dependencies {
	__antithesis_instrumentation__.Notify(577061)
	return scdeps.NewExecutorDependencies(
		execCfg.Codec,
		sessionData,
		txn,
		user,
		descriptors,
		execCfg.JobRegistry,
		execCfg.IndexBackfiller,

		scdeps.NewNoOpBackfillTracker(execCfg.Codec),
		scdeps.NewNoopPeriodicProgressFlusher(),
		execCfg.IndexValidator,
		scdeps.NewConstantClock(evalContext.GetTxnTimestamp(time.Microsecond).Time),
		execCfg.DescMetadaUpdaterFactory,
		NewSchemaChangerEventLogger(txn, execCfg, 1),
		kvTrace,
		schemaChangerJobID,
		stmts,
	)
}

func (s schemaChangePlanNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(577062)
	return false, nil
}
func (s schemaChangePlanNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(577063)
	return tree.Datums{}
}
func (s schemaChangePlanNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(577064)
}

var _ (planNode) = (*schemaChangePlanNode)(nil)
