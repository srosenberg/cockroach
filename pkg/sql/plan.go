package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type runParams struct {
	ctx context.Context

	extendedEvalCtx *extendedEvalContext

	p *planner
}

func (r *runParams) EvalContext() *tree.EvalContext {
	__antithesis_instrumentation__.Notify(562544)
	return &r.extendedEvalCtx.EvalContext
}

func (r *runParams) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(562545)
	return r.extendedEvalCtx.SessionData()
}

func (r *runParams) ExecCfg() *ExecutorConfig {
	__antithesis_instrumentation__.Notify(562546)
	return r.extendedEvalCtx.ExecCfg
}

func (r *runParams) Ann() *tree.Annotations {
	__antithesis_instrumentation__.Notify(562547)
	return r.extendedEvalCtx.EvalContext.Annotations
}

type planNode interface {
	startExec(params runParams) error

	Next(params runParams) (bool, error)

	Values() tree.Datums

	Close(ctx context.Context)
}

type mutationPlanNode interface {
	planNode

	rowsWritten() int64
}

type PlanNode = planNode

type planNodeFastPath interface {
	FastPathResults() (int, bool)
}

type planNodeReadingOwnWrites interface {
	ReadingOwnWrites()
}

var _ planNode = &alterIndexNode{}
var _ planNode = &alterSchemaNode{}
var _ planNode = &alterSequenceNode{}
var _ planNode = &alterTableNode{}
var _ planNode = &alterTableOwnerNode{}
var _ planNode = &alterTableSetSchemaNode{}
var _ planNode = &alterTypeNode{}
var _ planNode = &bufferNode{}
var _ planNode = &cancelQueriesNode{}
var _ planNode = &cancelSessionsNode{}
var _ planNode = &changePrivilegesNode{}
var _ planNode = &createDatabaseNode{}
var _ planNode = &createIndexNode{}
var _ planNode = &createSequenceNode{}
var _ planNode = &createStatsNode{}
var _ planNode = &createTableNode{}
var _ planNode = &createTypeNode{}
var _ planNode = &CreateRoleNode{}
var _ planNode = &createViewNode{}
var _ planNode = &delayedNode{}
var _ planNode = &deleteNode{}
var _ planNode = &deleteRangeNode{}
var _ planNode = &distinctNode{}
var _ planNode = &dropDatabaseNode{}
var _ planNode = &dropIndexNode{}
var _ planNode = &dropSchemaNode{}
var _ planNode = &dropSequenceNode{}
var _ planNode = &dropTableNode{}
var _ planNode = &dropTypeNode{}
var _ planNode = &DropRoleNode{}
var _ planNode = &dropViewNode{}
var _ planNode = &errorIfRowsNode{}
var _ planNode = &explainVecNode{}
var _ planNode = &filterNode{}
var _ planNode = &GrantRoleNode{}
var _ planNode = &groupNode{}
var _ planNode = &hookFnNode{}
var _ planNode = &indexJoinNode{}
var _ planNode = &insertNode{}
var _ planNode = &insertFastPathNode{}
var _ planNode = &joinNode{}
var _ planNode = &limitNode{}
var _ planNode = &max1RowNode{}
var _ planNode = &ordinalityNode{}
var _ planNode = &projectSetNode{}
var _ planNode = &reassignOwnedByNode{}
var _ planNode = &refreshMaterializedViewNode{}
var _ planNode = &recursiveCTENode{}
var _ planNode = &relocateNode{}
var _ planNode = &relocateRange{}
var _ planNode = &renameColumnNode{}
var _ planNode = &renameDatabaseNode{}
var _ planNode = &renameIndexNode{}
var _ planNode = &renameTableNode{}
var _ planNode = &reparentDatabaseNode{}
var _ planNode = &renderNode{}
var _ planNode = &RevokeRoleNode{}
var _ planNode = &rowCountNode{}
var _ planNode = &scanBufferNode{}
var _ planNode = &scanNode{}
var _ planNode = &scatterNode{}
var _ planNode = &serializeNode{}
var _ planNode = &sequenceSelectNode{}
var _ planNode = &showFingerprintsNode{}
var _ planNode = &showTraceNode{}
var _ planNode = &sortNode{}
var _ planNode = &splitNode{}
var _ planNode = &topKNode{}
var _ planNode = &unsplitNode{}
var _ planNode = &unsplitAllNode{}
var _ planNode = &truncateNode{}
var _ planNode = &unaryNode{}
var _ planNode = &unionNode{}
var _ planNode = &updateNode{}
var _ planNode = &upsertNode{}
var _ planNode = &valuesNode{}
var _ planNode = &virtualTableNode{}
var _ planNode = &windowNode{}
var _ planNode = &zeroNode{}

var _ planNodeFastPath = &deleteRangeNode{}
var _ planNodeFastPath = &rowCountNode{}
var _ planNodeFastPath = &serializeNode{}
var _ planNodeFastPath = &setZoneConfigNode{}
var _ planNodeFastPath = &controlJobsNode{}
var _ planNodeFastPath = &controlSchedulesNode{}

var _ planNodeReadingOwnWrites = &alterIndexNode{}
var _ planNodeReadingOwnWrites = &alterSchemaNode{}
var _ planNodeReadingOwnWrites = &alterSequenceNode{}
var _ planNodeReadingOwnWrites = &alterTableNode{}
var _ planNodeReadingOwnWrites = &alterTypeNode{}
var _ planNodeReadingOwnWrites = &createIndexNode{}
var _ planNodeReadingOwnWrites = &createSequenceNode{}
var _ planNodeReadingOwnWrites = &createDatabaseNode{}
var _ planNodeReadingOwnWrites = &createTableNode{}
var _ planNodeReadingOwnWrites = &createTypeNode{}
var _ planNodeReadingOwnWrites = &createViewNode{}
var _ planNodeReadingOwnWrites = &changePrivilegesNode{}
var _ planNodeReadingOwnWrites = &dropSchemaNode{}
var _ planNodeReadingOwnWrites = &dropTypeNode{}
var _ planNodeReadingOwnWrites = &refreshMaterializedViewNode{}
var _ planNodeReadingOwnWrites = &reparentDatabaseNode{}
var _ planNodeReadingOwnWrites = &setZoneConfigNode{}

type planNodeRequireSpool interface {
	requireSpool()
}

var _ planNodeRequireSpool = &serializeNode{}

type planNodeSpooled interface {
	spooled()
}

var _ planNodeSpooled = &spoolNode{}

type flowInfo struct {
	typ     planComponentType
	diagram execinfrapb.FlowDiagram

	explainVec        []string
	explainVecVerbose []string

	flowsMetadata *execstats.FlowsMetadata
}

type planTop struct {
	stmt *Statement

	planComponents

	mem     *memo.Memo
	catalog *optCatalog

	auditEvents []auditEvent

	flags planFlags

	avoidBuffering bool

	distSQLFlowInfos []flowInfo

	instrumentation *instrumentationHelper
}

type physicalPlanTop struct {
	*PhysicalPlan

	planNodesToClose []planNode
}

func (p *physicalPlanTop) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562548)
	for _, plan := range p.planNodesToClose {
		__antithesis_instrumentation__.Notify(562550)
		plan.Close(ctx)
	}
	__antithesis_instrumentation__.Notify(562549)
	p.planNodesToClose = nil
}

type planMaybePhysical struct {
	planNode planNode

	physPlan *physicalPlanTop
}

func makePlanMaybePhysical(physPlan *PhysicalPlan, planNodesToClose []planNode) planMaybePhysical {
	__antithesis_instrumentation__.Notify(562551)
	return planMaybePhysical{
		physPlan: &physicalPlanTop{
			PhysicalPlan:     physPlan,
			planNodesToClose: planNodesToClose,
		},
	}
}

func (p *planMaybePhysical) isPhysicalPlan() bool {
	__antithesis_instrumentation__.Notify(562552)
	return p.physPlan != nil
}

func (p *planMaybePhysical) planColumns() colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(562553)
	if p.isPhysicalPlan() {
		__antithesis_instrumentation__.Notify(562555)
		return p.physPlan.ResultColumns
	} else {
		__antithesis_instrumentation__.Notify(562556)
	}
	__antithesis_instrumentation__.Notify(562554)
	return planColumns(p.planNode)
}

func (p *planMaybePhysical) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562557)
	if p.planNode != nil {
		__antithesis_instrumentation__.Notify(562559)
		p.planNode.Close(ctx)
		p.planNode = nil
	} else {
		__antithesis_instrumentation__.Notify(562560)
	}
	__antithesis_instrumentation__.Notify(562558)
	if p.physPlan != nil {
		__antithesis_instrumentation__.Notify(562561)
		p.physPlan.Close(ctx)
		p.physPlan = nil
	} else {
		__antithesis_instrumentation__.Notify(562562)
	}
}

type planComponentType int

const (
	planComponentTypeUnknown = iota
	planComponentTypeMainQuery
	planComponentTypeSubquery
	planComponentTypePostquery
)

func (t planComponentType) String() string {
	__antithesis_instrumentation__.Notify(562563)
	switch t {
	case planComponentTypeMainQuery:
		__antithesis_instrumentation__.Notify(562564)
		return "main-query"
	case planComponentTypeSubquery:
		__antithesis_instrumentation__.Notify(562565)
		return "subquery"
	case planComponentTypePostquery:
		__antithesis_instrumentation__.Notify(562566)
		return "postquery"
	default:
		__antithesis_instrumentation__.Notify(562567)
		return "unknownquerytype"
	}
}

type planComponents struct {
	subqueryPlans []subquery

	main planMaybePhysical

	mainRowCount int64

	cascades []cascadeMetadata

	checkPlans []checkPlan
}

type cascadeMetadata struct {
	exec.Cascade

	plan planMaybePhysical
}

type checkPlan struct {
	plan planMaybePhysical
}

func (p *planComponents) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562568)
	p.main.Close(ctx)
	for i := range p.subqueryPlans {
		__antithesis_instrumentation__.Notify(562571)
		p.subqueryPlans[i].plan.Close(ctx)
	}
	__antithesis_instrumentation__.Notify(562569)
	for i := range p.cascades {
		__antithesis_instrumentation__.Notify(562572)
		p.cascades[i].plan.Close(ctx)
	}
	__antithesis_instrumentation__.Notify(562570)
	for i := range p.checkPlans {
		__antithesis_instrumentation__.Notify(562573)
		p.checkPlans[i].plan.Close(ctx)
	}
}

func (p *planTop) init(stmt *Statement, instrumentation *instrumentationHelper) {
	*p = planTop{
		stmt:            stmt,
		instrumentation: instrumentation,
	}
}

func (p *planTop) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562574)
	if p.flags.IsSet(planFlagExecDone) {
		__antithesis_instrumentation__.Notify(562576)
		p.savePlanInfo(ctx)
	} else {
		__antithesis_instrumentation__.Notify(562577)
	}
	__antithesis_instrumentation__.Notify(562575)
	p.planComponents.close(ctx)
}

func (p *planTop) savePlanInfo(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562578)
	vectorized := p.flags.IsSet(planFlagVectorized)
	distribution := physicalplan.LocalPlan
	if p.flags.IsSet(planFlagFullyDistributed) {
		__antithesis_instrumentation__.Notify(562580)
		distribution = physicalplan.FullyDistributedPlan
	} else {
		__antithesis_instrumentation__.Notify(562581)
		if p.flags.IsSet(planFlagPartiallyDistributed) {
			__antithesis_instrumentation__.Notify(562582)
			distribution = physicalplan.PartiallyDistributedPlan
		} else {
			__antithesis_instrumentation__.Notify(562583)
		}
	}
	__antithesis_instrumentation__.Notify(562579)
	p.instrumentation.RecordPlanInfo(distribution, vectorized)
}

func startExec(params runParams, plan planNode) error {
	__antithesis_instrumentation__.Notify(562584)
	o := planObserver{
		enterNode: func(ctx context.Context, _ string, p planNode) (bool, error) {
			__antithesis_instrumentation__.Notify(562586)
			switch p.(type) {
			case *explainVecNode, *explainDDLNode:
				__antithesis_instrumentation__.Notify(562588)

				return false, nil
			case *showTraceNode:
				__antithesis_instrumentation__.Notify(562589)

				return false, nil
			}
			__antithesis_instrumentation__.Notify(562587)
			return true, nil
		},
		leaveNode: func(_ string, n planNode) (err error) {
			__antithesis_instrumentation__.Notify(562590)
			if _, ok := n.(planNodeReadingOwnWrites); ok {
				__antithesis_instrumentation__.Notify(562592)
				prevMode := params.p.Txn().ConfigureStepping(params.ctx, kv.SteppingDisabled)
				defer func() {
					__antithesis_instrumentation__.Notify(562593)
					_ = params.p.Txn().ConfigureStepping(params.ctx, prevMode)
				}()
			} else {
				__antithesis_instrumentation__.Notify(562594)
			}
			__antithesis_instrumentation__.Notify(562591)
			return n.startExec(params)
		},
	}
	__antithesis_instrumentation__.Notify(562585)
	return walkPlan(params.ctx, plan, o)
}

func (p *planner) maybePlanHook(ctx context.Context, stmt tree.Statement) (planNode, error) {
	__antithesis_instrumentation__.Notify(562595)

	for _, planHook := range planHooks {
		__antithesis_instrumentation__.Notify(562597)
		if fn, header, subplans, avoidBuffering, err := planHook.fn(ctx, stmt, p); err != nil {
			__antithesis_instrumentation__.Notify(562598)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(562599)
			if fn != nil {
				__antithesis_instrumentation__.Notify(562600)
				if avoidBuffering {
					__antithesis_instrumentation__.Notify(562602)
					p.curPlan.avoidBuffering = true
				} else {
					__antithesis_instrumentation__.Notify(562603)
				}
				__antithesis_instrumentation__.Notify(562601)
				return newHookFnNode(planHook.name, fn, header, subplans), nil
			} else {
				__antithesis_instrumentation__.Notify(562604)
			}
		}
	}
	__antithesis_instrumentation__.Notify(562596)
	return nil, nil
}

func (p *planner) maybeSetSystemConfig(id descpb.ID) error {
	__antithesis_instrumentation__.Notify(562605)
	if !descpb.IsSystemConfigID(id) || func() bool {
		__antithesis_instrumentation__.Notify(562607)
		return p.execCfg.Settings.Version.IsActive(
			p.EvalContext().Ctx(), clusterversion.DisableSystemConfigGossipTrigger,
		) == true
	}() == true {
		__antithesis_instrumentation__.Notify(562608)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(562609)
	}
	__antithesis_instrumentation__.Notify(562606)

	return p.txn.DeprecatedSetSystemConfigTrigger(p.execCfg.Codec.ForSystemTenant())
}

type planFlags uint32

const (
	planFlagOptCacheHit = (1 << iota)

	planFlagOptCacheMiss

	planFlagFullyDistributed

	planFlagPartiallyDistributed

	planFlagNotDistributed

	planFlagExecDone

	planFlagImplicitTxn

	planFlagIsDDL

	planFlagVectorized

	planFlagTenant

	planFlagContainsFullTableScan

	planFlagContainsFullIndexScan

	planFlagContainsLargeFullTableScan

	planFlagContainsLargeFullIndexScan

	planFlagContainsMutation
)

func (pf planFlags) IsSet(flag planFlags) bool {
	__antithesis_instrumentation__.Notify(562610)
	return (pf & flag) != 0
}

func (pf *planFlags) Set(flag planFlags) {
	__antithesis_instrumentation__.Notify(562611)
	*pf |= flag
}

func (pf *planFlags) Unset(flag planFlags) {
	__antithesis_instrumentation__.Notify(562612)
	*pf &= ^flag
}

func (pf planFlags) IsDistributed() bool {
	__antithesis_instrumentation__.Notify(562613)
	return pf.IsSet(planFlagFullyDistributed) || func() bool {
		__antithesis_instrumentation__.Notify(562614)
		return pf.IsSet(planFlagPartiallyDistributed) == true
	}() == true
}
