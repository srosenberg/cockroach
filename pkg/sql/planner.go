package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/lease"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/idxusage"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/querycache"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/lib/pq/oid"
)

type extendedEvalContext struct {
	tree.EvalContext

	SessionID ClusterWideID

	VirtualSchemas VirtualTabler

	Tracing *SessionTracing

	NodesStatusServer serverpb.OptionalNodesStatusServer

	RegionsServer serverpb.RegionsServer

	SQLStatusServer serverpb.SQLStatusServer

	MemMetrics *MemoryMetrics

	Descs *descs.Collection

	ExecCfg *ExecutorConfig

	DistSQLPlanner *DistSQLPlanner

	TxnModesSetter txnModesSetter

	Jobs *jobsCollection

	SchemaChangeJobRecords map[descpb.ID]*jobs.Record

	statsProvider *persistedsqlstats.PersistedSQLStats

	indexUsageStats *idxusage.LocalIndexUsageStats

	SchemaChangerState *SchemaChangerState

	statementPreparer statementPreparer
}

func (evalCtx *extendedEvalContext) copyFromExecCfg(execCfg *ExecutorConfig) {
	__antithesis_instrumentation__.Notify(563062)
	evalCtx.ExecCfg = execCfg
	evalCtx.Settings = execCfg.Settings
	evalCtx.Codec = execCfg.Codec
	evalCtx.Tracer = execCfg.AmbientCtx.Tracer
	evalCtx.DB = execCfg.DB
	evalCtx.SQLLivenessReader = execCfg.SQLLiveness
	evalCtx.CompactEngineSpan = execCfg.CompactEngineSpanFunc
	evalCtx.TestingKnobs = execCfg.EvalContextTestingKnobs
	evalCtx.ClusterID = execCfg.LogicalClusterID()
	evalCtx.ClusterName = execCfg.RPCContext.ClusterName()
	evalCtx.NodeID = execCfg.NodeID
	evalCtx.Locality = execCfg.Locality
	evalCtx.NodesStatusServer = execCfg.NodesStatusServer
	evalCtx.RegionsServer = execCfg.RegionsServer
	evalCtx.SQLStatusServer = execCfg.SQLStatusServer
	evalCtx.DistSQLPlanner = execCfg.DistSQLPlanner
	evalCtx.VirtualSchemas = execCfg.VirtualSchemas
	evalCtx.KVStoresIterator = execCfg.KVStoresIterator
}

func (evalCtx *extendedEvalContext) copy() *extendedEvalContext {
	__antithesis_instrumentation__.Notify(563063)
	cpy := *evalCtx
	cpy.EvalContext = *evalCtx.EvalContext.Copy()
	return &cpy
}

func (evalCtx *extendedEvalContext) QueueJob(
	ctx context.Context, record jobs.Record,
) (*jobs.Job, error) {
	__antithesis_instrumentation__.Notify(563064)
	jobID := evalCtx.ExecCfg.JobRegistry.MakeJobID()
	job, err := evalCtx.ExecCfg.JobRegistry.CreateJobWithTxn(
		ctx,
		record,
		jobID,
		evalCtx.Txn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(563066)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563067)
	}
	__antithesis_instrumentation__.Notify(563065)
	evalCtx.Jobs.add(jobID)
	return job, nil
}

type planner struct {
	txn *kv.Txn

	isInternalPlanner bool

	stmt Statement

	instrumentation instrumentationHelper

	semaCtx         tree.SemaContext
	extendedEvalCtx extendedEvalContext

	sessionDataMutatorIterator *sessionDataMutatorIterator

	execCfg *ExecutorConfig

	preparedStatements preparedStatementsAccessor

	sqlCursors sqlCursors

	createdSequences createdSequences

	avoidLeasedDescriptors bool

	autoCommit bool

	cancelChecker cancelchecker.CancelChecker

	isPreparing bool

	curPlan planTop

	txCtx                 transform.ExprTransformContext
	nameResolutionVisitor schemaexpr.NameResolutionVisitor
	tableName             tree.TableName

	alloc *tree.DatumAlloc

	optPlanningCtx optPlanningCtx

	noticeSender noticeSender

	queryCacheSession querycache.Session

	contextDatabaseID descpb.ID
}

func (evalCtx *extendedEvalContext) setSessionID(sessionID ClusterWideID) {
	__antithesis_instrumentation__.Notify(563068)
	evalCtx.SessionID = sessionID
}

var noteworthyInternalMemoryUsageBytes = envutil.EnvOrDefaultInt64("COCKROACH_NOTEWORTHY_INTERNAL_MEMORY_USAGE", 1<<20)

type internalPlannerParams struct {
	collection *descs.Collection
}

type InternalPlannerParamsOption func(*internalPlannerParams)

func WithDescCollection(collection *descs.Collection) InternalPlannerParamsOption {
	__antithesis_instrumentation__.Notify(563069)
	return func(params *internalPlannerParams) {
		__antithesis_instrumentation__.Notify(563070)
		params.collection = collection
	}
}

func NewInternalPlanner(
	opName string,
	txn *kv.Txn,
	user security.SQLUsername,
	memMetrics *MemoryMetrics,
	execCfg *ExecutorConfig,
	sessionData sessiondatapb.SessionData,
	opts ...InternalPlannerParamsOption,
) (interface{}, func()) {
	__antithesis_instrumentation__.Notify(563071)
	return newInternalPlanner(opName, txn, user, memMetrics, execCfg, sessionData, opts...)
}

func newInternalPlanner(
	opName string,
	txn *kv.Txn,
	user security.SQLUsername,
	memMetrics *MemoryMetrics,
	execCfg *ExecutorConfig,
	sessionData sessiondatapb.SessionData,
	opts ...InternalPlannerParamsOption,
) (*planner, func()) {
	__antithesis_instrumentation__.Notify(563072)

	params := &internalPlannerParams{}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(563077)
		opt(params)
	}
	__antithesis_instrumentation__.Notify(563073)
	callerSuppliedDescsCollection := params.collection != nil

	ctx := logtags.AddTag(context.Background(), opName, "")

	sd := &sessiondata.SessionData{
		SessionData:   sessionData,
		SearchPath:    sessiondata.DefaultSearchPathForUser(user),
		SequenceState: sessiondata.NewSequenceState(),
		Location:      time.UTC,
	}
	sd.SessionData.Database = "system"
	sd.SessionData.UserProto = user.EncodeProto()
	sd.SessionData.Internal = true
	sds := sessiondata.NewStack(sd)

	if params.collection == nil {
		__antithesis_instrumentation__.Notify(563078)
		params.collection = execCfg.CollectionFactory.NewCollection(ctx, descs.NewTemporarySchemaProvider(sds))
	} else {
		__antithesis_instrumentation__.Notify(563079)
	}
	__antithesis_instrumentation__.Notify(563074)

	var ts time.Time
	if txn != nil {
		__antithesis_instrumentation__.Notify(563080)
		readTimestamp := txn.ReadTimestamp()
		if readTimestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(563082)
			panic("makeInternalPlanner called with a transaction without timestamps")
		} else {
			__antithesis_instrumentation__.Notify(563083)
		}
		__antithesis_instrumentation__.Notify(563081)
		ts = readTimestamp.GoTime()
	} else {
		__antithesis_instrumentation__.Notify(563084)
	}
	__antithesis_instrumentation__.Notify(563075)

	p := &planner{execCfg: execCfg, alloc: &tree.DatumAlloc{}}

	p.txn = txn
	p.stmt = Statement{}
	p.cancelChecker.Reset(ctx)
	p.isInternalPlanner = true

	p.semaCtx = tree.MakeSemaContext()
	if p.execCfg.Settings.Version.IsActive(ctx, clusterversion.DateStyleIntervalStyleCastRewrite) {
		__antithesis_instrumentation__.Notify(563085)
		p.semaCtx.IntervalStyleEnabled = true
		p.semaCtx.DateStyleEnabled = true
	} else {
		__antithesis_instrumentation__.Notify(563086)
		p.semaCtx.IntervalStyleEnabled = sd.IntervalStyleEnabled
		p.semaCtx.DateStyleEnabled = sd.DateStyleEnabled
	}
	__antithesis_instrumentation__.Notify(563076)
	p.semaCtx.SearchPath = sd.SearchPath
	p.semaCtx.TypeResolver = p
	p.semaCtx.DateStyle = sd.GetDateStyle()
	p.semaCtx.IntervalStyle = sd.GetIntervalStyle()

	plannerMon := mon.NewMonitor(fmt.Sprintf("internal-planner.%s.%s", user, opName),
		mon.MemoryResource,
		memMetrics.CurBytesCount, memMetrics.MaxBytesHist,
		-1,
		noteworthyInternalMemoryUsageBytes, execCfg.Settings)
	plannerMon.Start(ctx, execCfg.RootMemoryMonitor, mon.BoundAccount{})

	smi := &sessionDataMutatorIterator{
		sds: sds,
		sessionDataMutatorBase: sessionDataMutatorBase{
			defaults: SessionDefaults(map[string]string{
				"application_name": "crdb-internal",
				"database":         "system",
			}),
			settings: execCfg.Settings,
		},
		sessionDataMutatorCallbacks: sessionDataMutatorCallbacks{},
	}

	p.extendedEvalCtx = internalExtendedEvalCtx(ctx, sds, params.collection, txn, ts, ts, execCfg, plannerMon)
	p.extendedEvalCtx.Planner = p
	p.extendedEvalCtx.PrivilegedAccessor = p
	p.extendedEvalCtx.SessionAccessor = p
	p.extendedEvalCtx.ClientNoticeSender = p
	p.extendedEvalCtx.Sequence = p
	p.extendedEvalCtx.Tenant = p
	p.extendedEvalCtx.Regions = p
	p.extendedEvalCtx.JoinTokenCreator = p
	p.extendedEvalCtx.ClusterID = execCfg.LogicalClusterID()
	p.extendedEvalCtx.ClusterName = execCfg.RPCContext.ClusterName()
	p.extendedEvalCtx.NodeID = execCfg.NodeID
	p.extendedEvalCtx.Locality = execCfg.Locality

	p.sessionDataMutatorIterator = smi
	p.autoCommit = false

	p.extendedEvalCtx.MemMetrics = memMetrics
	p.extendedEvalCtx.ExecCfg = execCfg
	p.extendedEvalCtx.Placeholders = &p.semaCtx.Placeholders
	p.extendedEvalCtx.Annotations = &p.semaCtx.Annotations
	p.extendedEvalCtx.Descs = params.collection

	p.queryCacheSession.Init()
	p.optPlanningCtx.init(p)
	p.createdSequences = emptyCreatedSequences{}

	return p, func() {
		__antithesis_instrumentation__.Notify(563087)

		if !callerSuppliedDescsCollection {
			__antithesis_instrumentation__.Notify(563089)

			p.Descriptors().ReleaseAll(ctx)
		} else {
			__antithesis_instrumentation__.Notify(563090)
		}
		__antithesis_instrumentation__.Notify(563088)

		plannerMon.Stop(ctx)
	}
}

func internalExtendedEvalCtx(
	ctx context.Context,
	sds *sessiondata.Stack,
	tables *descs.Collection,
	txn *kv.Txn,
	txnTimestamp time.Time,
	stmtTimestamp time.Time,
	execCfg *ExecutorConfig,
	plannerMon *mon.BytesMonitor,
) extendedEvalContext {
	__antithesis_instrumentation__.Notify(563091)
	evalContextTestingKnobs := execCfg.EvalContextTestingKnobs

	var indexUsageStats *idxusage.LocalIndexUsageStats
	var sqlStatsController tree.SQLStatsController
	var indexUsageStatsController tree.IndexUsageStatsController
	if execCfg.InternalExecutor != nil {
		__antithesis_instrumentation__.Notify(563093)
		if execCfg.InternalExecutor.s != nil {
			__antithesis_instrumentation__.Notify(563094)
			indexUsageStats = execCfg.InternalExecutor.s.indexUsageStats
			sqlStatsController = execCfg.InternalExecutor.s.sqlStatsController
			indexUsageStatsController = execCfg.InternalExecutor.s.indexUsageStatsController
		} else {
			__antithesis_instrumentation__.Notify(563095)

			indexUsageStats = idxusage.NewLocalIndexUsageStats(&idxusage.Config{
				Setting: execCfg.Settings,
			})
			sqlStatsController = &persistedsqlstats.Controller{}
			indexUsageStatsController = &idxusage.Controller{}
		}
	} else {
		__antithesis_instrumentation__.Notify(563096)
	}
	__antithesis_instrumentation__.Notify(563092)
	ret := extendedEvalContext{
		EvalContext: tree.EvalContext{
			Txn:                       txn,
			SessionDataStack:          sds,
			TxnReadOnly:               false,
			TxnImplicit:               true,
			TxnIsSingleStmt:           true,
			Context:                   ctx,
			Mon:                       plannerMon,
			TestingKnobs:              evalContextTestingKnobs,
			StmtTimestamp:             stmtTimestamp,
			TxnTimestamp:              txnTimestamp,
			SQLStatsController:        sqlStatsController,
			IndexUsageStatsController: indexUsageStatsController,
		},
		Tracing:         &SessionTracing{},
		Descs:           tables,
		indexUsageStats: indexUsageStats,
	}
	ret.copyFromExecCfg(execCfg)
	return ret
}

func (p *planner) Accessor() catalog.Accessor {
	__antithesis_instrumentation__.Notify(563097)
	return p.Descriptors()
}

func (p *planner) SemaCtx() *tree.SemaContext {
	__antithesis_instrumentation__.Notify(563098)
	return &p.semaCtx
}

func (p *planner) ExtendedEvalContext() *extendedEvalContext {
	__antithesis_instrumentation__.Notify(563099)
	return &p.extendedEvalCtx
}

func (p *planner) ExtendedEvalContextCopy() *extendedEvalContext {
	__antithesis_instrumentation__.Notify(563100)
	return p.extendedEvalCtx.copy()
}

func (p *planner) CurrentDatabase() string {
	__antithesis_instrumentation__.Notify(563101)
	return p.SessionData().Database
}

func (p *planner) CurrentSearchPath() sessiondata.SearchPath {
	__antithesis_instrumentation__.Notify(563102)
	return p.SessionData().SearchPath
}

func (p *planner) EvalContext() *tree.EvalContext {
	__antithesis_instrumentation__.Notify(563103)
	return &p.extendedEvalCtx.EvalContext
}

func (p *planner) Descriptors() *descs.Collection {
	__antithesis_instrumentation__.Notify(563104)
	return p.extendedEvalCtx.Descs
}

func (p *planner) ExecCfg() *ExecutorConfig {
	__antithesis_instrumentation__.Notify(563105)
	return p.extendedEvalCtx.ExecCfg
}

func (p *planner) GetOrInitSequenceCache() sessiondatapb.SequenceCache {
	__antithesis_instrumentation__.Notify(563106)
	if p.SessionData().SequenceCache == nil {
		__antithesis_instrumentation__.Notify(563108)
		p.sessionDataMutatorIterator.applyOnEachMutator(
			func(m sessionDataMutator) {
				__antithesis_instrumentation__.Notify(563109)
				m.initSequenceCache()
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(563110)
	}
	__antithesis_instrumentation__.Notify(563107)
	return p.SessionData().SequenceCache
}

func (p *planner) LeaseMgr() *lease.Manager {
	__antithesis_instrumentation__.Notify(563111)
	return p.execCfg.LeaseManager
}

func (p *planner) Txn() *kv.Txn {
	__antithesis_instrumentation__.Notify(563112)
	return p.txn
}

func (p *planner) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(563113)
	return p.SessionData().User()
}

func (p *planner) TemporarySchemaName() string {
	__antithesis_instrumentation__.Notify(563114)
	return temporarySchemaName(p.ExtendedEvalContext().SessionID)
}

func (p *planner) DistSQLPlanner() *DistSQLPlanner {
	__antithesis_instrumentation__.Notify(563115)
	return p.extendedEvalCtx.DistSQLPlanner
}

func (p *planner) MigrationJobDeps() migration.JobDeps {
	__antithesis_instrumentation__.Notify(563116)
	return p.execCfg.MigrationJobDeps
}

func (p *planner) SpanConfigReconciler() spanconfig.Reconciler {
	__antithesis_instrumentation__.Notify(563117)
	return p.execCfg.SpanConfigReconciler
}

func (p *planner) GetTypeFromValidSQLSyntax(sql string) (*types.T, error) {
	__antithesis_instrumentation__.Notify(563118)
	ref, err := parser.GetTypeFromValidSQLSyntax(sql)
	if err != nil {
		__antithesis_instrumentation__.Notify(563120)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563121)
	}
	__antithesis_instrumentation__.Notify(563119)
	return tree.ResolveType(context.TODO(), ref, p.semaCtx.GetTypeResolver())
}

func (p *planner) ParseQualifiedTableName(sql string) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(563122)
	return parser.ParseQualifiedTableName(sql)
}

func (p *planner) ResolveTableName(ctx context.Context, tn *tree.TableName) (tree.ID, error) {
	__antithesis_instrumentation__.Notify(563123)
	flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveAnyTableKind)
	_, desc, err := resolver.ResolveExistingTableObject(ctx, p, tn, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(563125)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(563126)
	}
	__antithesis_instrumentation__.Notify(563124)
	return tree.ID(desc.GetID()), nil
}

func (p *planner) LookupTableByID(
	ctx context.Context, tableID descpb.ID,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(563127)
	const required = true
	table, err := p.Descriptors().GetImmutableTableByID(ctx, p.txn, tableID, p.ObjectLookupFlags(
		required, false))
	if err != nil {
		__antithesis_instrumentation__.Notify(563129)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563130)
	}
	__antithesis_instrumentation__.Notify(563128)
	return table, nil
}

func (p *planner) TypeAsString(
	ctx context.Context, e tree.Expr, op string,
) (func() (string, error), error) {
	__antithesis_instrumentation__.Notify(563131)
	typedE, err := tree.TypeCheckAndRequire(ctx, e, &p.semaCtx, types.String, op)
	if err != nil {
		__antithesis_instrumentation__.Notify(563133)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563134)
	}
	__antithesis_instrumentation__.Notify(563132)
	evalFn := p.makeStringEvalFn(typedE)
	return func() (string, error) {
		__antithesis_instrumentation__.Notify(563135)
		isNull, str, err := evalFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(563138)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(563139)
		}
		__antithesis_instrumentation__.Notify(563136)
		if isNull {
			__antithesis_instrumentation__.Notify(563140)
			return "", errors.Errorf("expected string, got NULL")
		} else {
			__antithesis_instrumentation__.Notify(563141)
		}
		__antithesis_instrumentation__.Notify(563137)
		return str, nil
	}, nil
}

func (p *planner) TypeAsStringOrNull(
	ctx context.Context, e tree.Expr, op string,
) (func() (bool, string, error), error) {
	__antithesis_instrumentation__.Notify(563142)
	typedE, err := tree.TypeCheckAndRequire(ctx, e, &p.semaCtx, types.String, op)
	if err != nil {
		__antithesis_instrumentation__.Notify(563144)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(563145)
	}
	__antithesis_instrumentation__.Notify(563143)
	return p.makeStringEvalFn(typedE), nil
}

func (p *planner) makeStringEvalFn(typedE tree.TypedExpr) func() (bool, string, error) {
	__antithesis_instrumentation__.Notify(563146)
	return func() (bool, string, error) {
		__antithesis_instrumentation__.Notify(563147)
		d, err := typedE.Eval(p.EvalContext())
		if err != nil {
			__antithesis_instrumentation__.Notify(563151)
			return false, "", err
		} else {
			__antithesis_instrumentation__.Notify(563152)
		}
		__antithesis_instrumentation__.Notify(563148)
		if d == tree.DNull {
			__antithesis_instrumentation__.Notify(563153)
			return true, "", nil
		} else {
			__antithesis_instrumentation__.Notify(563154)
		}
		__antithesis_instrumentation__.Notify(563149)
		str, ok := d.(*tree.DString)
		if !ok {
			__antithesis_instrumentation__.Notify(563155)
			return false, "", errors.Errorf("failed to cast %T to string", d)
		} else {
			__antithesis_instrumentation__.Notify(563156)
		}
		__antithesis_instrumentation__.Notify(563150)
		return false, string(*str), nil
	}
}

type KVStringOptValidate string

const (
	KVStringOptAny            KVStringOptValidate = `any`
	KVStringOptRequireNoValue KVStringOptValidate = `no-value`
	KVStringOptRequireValue   KVStringOptValidate = `value`
)

func evalStringOptions(
	evalCtx *tree.EvalContext, opts []exec.KVOption, optValidate map[string]KVStringOptValidate,
) (map[string]string, error) {
	__antithesis_instrumentation__.Notify(563157)
	res := make(map[string]string, len(opts))
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(563159)
		k := opt.Key
		validate, ok := optValidate[k]
		if !ok {
			__antithesis_instrumentation__.Notify(563162)
			return nil, errors.Errorf("invalid option %q", k)
		} else {
			__antithesis_instrumentation__.Notify(563163)
		}
		__antithesis_instrumentation__.Notify(563160)
		val, err := opt.Value.Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(563164)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(563165)
		}
		__antithesis_instrumentation__.Notify(563161)
		if val == tree.DNull {
			__antithesis_instrumentation__.Notify(563166)
			if validate == KVStringOptRequireValue {
				__antithesis_instrumentation__.Notify(563168)
				return nil, errors.Errorf("option %q requires a value", k)
			} else {
				__antithesis_instrumentation__.Notify(563169)
			}
			__antithesis_instrumentation__.Notify(563167)
			res[k] = ""
		} else {
			__antithesis_instrumentation__.Notify(563170)
			if validate == KVStringOptRequireNoValue {
				__antithesis_instrumentation__.Notify(563173)
				return nil, errors.Errorf("option %q does not take a value", k)
			} else {
				__antithesis_instrumentation__.Notify(563174)
			}
			__antithesis_instrumentation__.Notify(563171)
			str, ok := val.(*tree.DString)
			if !ok {
				__antithesis_instrumentation__.Notify(563175)
				return nil, errors.Errorf("expected string value, got %T", val)
			} else {
				__antithesis_instrumentation__.Notify(563176)
			}
			__antithesis_instrumentation__.Notify(563172)
			res[k] = string(*str)
		}
	}
	__antithesis_instrumentation__.Notify(563158)
	return res, nil
}

func (p *planner) TypeAsStringOpts(
	ctx context.Context, opts tree.KVOptions, optValidate map[string]KVStringOptValidate,
) (func() (map[string]string, error), error) {
	__antithesis_instrumentation__.Notify(563177)
	typed := make(map[string]tree.TypedExpr, len(opts))
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(563180)
		k := string(opt.Key)
		validate, ok := optValidate[k]
		if !ok {
			__antithesis_instrumentation__.Notify(563185)
			return nil, errors.Errorf("invalid option %q", k)
		} else {
			__antithesis_instrumentation__.Notify(563186)
		}
		__antithesis_instrumentation__.Notify(563181)

		if opt.Value == nil {
			__antithesis_instrumentation__.Notify(563187)
			if validate == KVStringOptRequireValue {
				__antithesis_instrumentation__.Notify(563189)
				return nil, errors.Errorf("option %q requires a value", k)
			} else {
				__antithesis_instrumentation__.Notify(563190)
			}
			__antithesis_instrumentation__.Notify(563188)
			typed[k] = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(563191)
		}
		__antithesis_instrumentation__.Notify(563182)
		if validate == KVStringOptRequireNoValue {
			__antithesis_instrumentation__.Notify(563192)
			return nil, errors.Errorf("option %q does not take a value", k)
		} else {
			__antithesis_instrumentation__.Notify(563193)
		}
		__antithesis_instrumentation__.Notify(563183)
		r, err := tree.TypeCheckAndRequire(ctx, opt.Value, &p.semaCtx, types.String, k)
		if err != nil {
			__antithesis_instrumentation__.Notify(563194)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(563195)
		}
		__antithesis_instrumentation__.Notify(563184)
		typed[k] = r
	}
	__antithesis_instrumentation__.Notify(563178)
	fn := func() (map[string]string, error) {
		__antithesis_instrumentation__.Notify(563196)
		res := make(map[string]string, len(typed))
		for name, e := range typed {
			__antithesis_instrumentation__.Notify(563198)
			if e == nil {
				__antithesis_instrumentation__.Notify(563202)
				res[name] = ""
				continue
			} else {
				__antithesis_instrumentation__.Notify(563203)
			}
			__antithesis_instrumentation__.Notify(563199)
			d, err := e.Eval(p.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(563204)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(563205)
			}
			__antithesis_instrumentation__.Notify(563200)
			str, ok := d.(*tree.DString)
			if !ok {
				__antithesis_instrumentation__.Notify(563206)
				return res, errors.Errorf("failed to cast %T to string", d)
			} else {
				__antithesis_instrumentation__.Notify(563207)
			}
			__antithesis_instrumentation__.Notify(563201)
			res[name] = string(*str)
		}
		__antithesis_instrumentation__.Notify(563197)
		return res, nil
	}
	__antithesis_instrumentation__.Notify(563179)
	return fn, nil
}

func (p *planner) TypeAsStringArray(
	ctx context.Context, exprs tree.Exprs, op string,
) (func() ([]string, error), error) {
	__antithesis_instrumentation__.Notify(563208)
	typedExprs := make([]tree.TypedExpr, len(exprs))
	for i := range exprs {
		__antithesis_instrumentation__.Notify(563211)
		typedE, err := tree.TypeCheckAndRequire(ctx, exprs[i], &p.semaCtx, types.String, op)
		if err != nil {
			__antithesis_instrumentation__.Notify(563213)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(563214)
		}
		__antithesis_instrumentation__.Notify(563212)
		typedExprs[i] = typedE
	}
	__antithesis_instrumentation__.Notify(563209)
	fn := func() ([]string, error) {
		__antithesis_instrumentation__.Notify(563215)
		strs := make([]string, len(exprs))
		for i := range exprs {
			__antithesis_instrumentation__.Notify(563217)
			d, err := typedExprs[i].Eval(p.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(563220)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(563221)
			}
			__antithesis_instrumentation__.Notify(563218)
			str, ok := d.(*tree.DString)
			if !ok {
				__antithesis_instrumentation__.Notify(563222)
				return strs, errors.Errorf("failed to cast %T to string", d)
			} else {
				__antithesis_instrumentation__.Notify(563223)
			}
			__antithesis_instrumentation__.Notify(563219)
			strs[i] = string(*str)
		}
		__antithesis_instrumentation__.Notify(563216)
		return strs, nil
	}
	__antithesis_instrumentation__.Notify(563210)
	return fn, nil
}

func (p *planner) SessionData() *sessiondata.SessionData {
	__antithesis_instrumentation__.Notify(563224)
	return p.EvalContext().SessionData()
}

func (p *planner) SessionDataMutatorIterator() *sessionDataMutatorIterator {
	__antithesis_instrumentation__.Notify(563225)
	return p.sessionDataMutatorIterator
}

func (p *planner) Ann() *tree.Annotations {
	__antithesis_instrumentation__.Notify(563226)
	return p.ExtendedEvalContext().EvalContext.Annotations
}

func (p *planner) ExecutorConfig() interface{} {
	__antithesis_instrumentation__.Notify(563227)
	return p.execCfg
}

type statementPreparer interface {
	addPreparedStmt(
		ctx context.Context,
		name string,
		stmt Statement,
		placeholderHints tree.PlaceholderTypes,
		rawTypeHints []oid.Oid,
		origin PreparedStatementOrigin,
	) (*PreparedStatement, error)
}

var _ statementPreparer = &connExecutor{}

type txnModesSetter interface {
	setTransactionModes(ctx context.Context, modes tree.TransactionModes, asOfTs hlc.Timestamp) error
}

func validateDescriptor(ctx context.Context, p *planner, descriptor catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(563228)
	return p.Descriptors().Validate(
		ctx,
		p.Txn(),
		catalog.NoValidationTelemetry,
		catalog.ValidationLevelCrossReferences,
		descriptor,
	)
}

func (p *planner) QueryRowEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	override sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(563229)
	ie := p.ExecCfg().InternalExecutorFactory(ctx, p.SessionData())
	return ie.QueryRowEx(ctx, opName, txn, override, stmt, qargs...)
}

func (p *planner) QueryIteratorEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	override sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.InternalRows, error) {
	__antithesis_instrumentation__.Notify(563230)
	ie := p.ExecCfg().InternalExecutorFactory(ctx, p.SessionData())
	rows, err := ie.QueryIteratorEx(ctx, opName, txn, override, stmt, qargs...)
	return rows.(tree.InternalRows), err
}
