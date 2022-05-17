package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/colflow"
	"github.com/cockroachdb/cockroach/pkg/sql/distsql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan/replicaoracle"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlinstance"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type DistSQLPlanner struct {
	planVersion execinfrapb.DistSQLVersion

	st *cluster.Settings

	gatewaySQLInstanceID base.SQLInstanceID
	stopper              *stop.Stopper
	distSQLSrv           *distsql.ServerImpl
	spanResolver         physicalplan.SpanResolver

	metadataTestTolerance execinfra.MetadataTestLevel

	runnerChan chan runnerRequest

	cancelFlowsCoordinator cancelFlowsCoordinator

	gossip gossip.OptionalGossip

	nodeDialer *nodedialer.Dialer

	podNodeDialer *nodedialer.Dialer

	nodeHealth distSQLNodeHealth

	parallelLocalScansSem *quotapool.IntPool

	distSender *kvcoord.DistSender

	nodeDescs kvcoord.NodeDescStore

	rpcCtx *rpc.Context

	sqlInstanceProvider sqlinstance.Provider

	codec keys.SQLCodec
}

type DistributionType int

const (
	DistributionTypeNone = iota

	DistributionTypeAlways

	DistributionTypeSystemTenantOnly
)

var ReplicaOraclePolicy = replicaoracle.BinPackingChoice

var logPlanDiagram = envutil.EnvOrDefaultBool("COCKROACH_DISTSQL_LOG_PLAN", false)

func NewDistSQLPlanner(
	ctx context.Context,
	planVersion execinfrapb.DistSQLVersion,
	st *cluster.Settings,
	sqlInstanceID base.SQLInstanceID,
	rpcCtx *rpc.Context,
	distSQLSrv *distsql.ServerImpl,
	distSender *kvcoord.DistSender,
	nodeDescs kvcoord.NodeDescStore,
	gw gossip.OptionalGossip,
	stopper *stop.Stopper,
	isAvailable func(base.SQLInstanceID) bool,
	nodeDialer *nodedialer.Dialer,
	podNodeDialer *nodedialer.Dialer,
	codec keys.SQLCodec,
	sqlInstanceProvider sqlinstance.Provider,
) *DistSQLPlanner {
	__antithesis_instrumentation__.Notify(466385)
	dsp := &DistSQLPlanner{
		planVersion:          planVersion,
		st:                   st,
		gatewaySQLInstanceID: sqlInstanceID,
		stopper:              stopper,
		distSQLSrv:           distSQLSrv,
		gossip:               gw,
		nodeDialer:           nodeDialer,
		podNodeDialer:        podNodeDialer,
		nodeHealth: distSQLNodeHealth{
			gossip:      gw,
			connHealth:  nodeDialer.ConnHealthTryDial,
			isAvailable: isAvailable,
		},
		distSender:            distSender,
		nodeDescs:             nodeDescs,
		rpcCtx:                rpcCtx,
		metadataTestTolerance: execinfra.NoExplain,
		sqlInstanceProvider:   sqlInstanceProvider,
		codec:                 codec,
	}

	dsp.parallelLocalScansSem = quotapool.NewIntPool("parallel local scans concurrency",
		uint64(localScansConcurrencyLimit.Get(&st.SV)))
	localScansConcurrencyLimit.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(466388)
		dsp.parallelLocalScansSem.UpdateCapacity(uint64(localScansConcurrencyLimit.Get(&st.SV)))
	})
	__antithesis_instrumentation__.Notify(466386)
	if rpcCtx != nil {
		__antithesis_instrumentation__.Notify(466389)

		rpcCtx.Stopper.AddCloser(dsp.parallelLocalScansSem.Closer("stopper"))
	} else {
		__antithesis_instrumentation__.Notify(466390)
	}
	__antithesis_instrumentation__.Notify(466387)

	dsp.initRunners(ctx)
	dsp.initCancelingWorkers(ctx)
	return dsp
}

func (dsp *DistSQLPlanner) shouldPlanTestMetadata() bool {
	__antithesis_instrumentation__.Notify(466391)
	return dsp.distSQLSrv.TestingKnobs.MetadataTestLevel >= dsp.metadataTestTolerance
}

func (dsp *DistSQLPlanner) GetSQLInstanceInfo(
	sqlInstanceID base.SQLInstanceID,
) (*roachpb.NodeDescriptor, error) {
	__antithesis_instrumentation__.Notify(466392)
	return dsp.nodeDescs.GetNodeDescriptor(roachpb.NodeID(sqlInstanceID))
}

func (dsp *DistSQLPlanner) SetSQLInstanceInfo(desc roachpb.NodeDescriptor) {
	__antithesis_instrumentation__.Notify(466393)
	dsp.gatewaySQLInstanceID = base.SQLInstanceID(desc.NodeID)
	if dsp.spanResolver == nil {
		__antithesis_instrumentation__.Notify(466394)
		sr := physicalplan.NewSpanResolver(dsp.st, dsp.distSender, dsp.nodeDescs, desc,
			dsp.rpcCtx, ReplicaOraclePolicy)
		dsp.SetSpanResolver(sr)
	} else {
		__antithesis_instrumentation__.Notify(466395)
	}
}

func (dsp *DistSQLPlanner) GatewayID() base.SQLInstanceID {
	__antithesis_instrumentation__.Notify(466396)
	return dsp.gatewaySQLInstanceID
}

func (dsp *DistSQLPlanner) SetSpanResolver(spanResolver physicalplan.SpanResolver) {
	__antithesis_instrumentation__.Notify(466397)
	dsp.spanResolver = spanResolver
}

type distSQLExprCheckVisitor struct {
	err error
}

var _ tree.Visitor = &distSQLExprCheckVisitor{}

func (v *distSQLExprCheckVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(466398)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(466401)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(466402)
	}
	__antithesis_instrumentation__.Notify(466399)
	switch t := expr.(type) {
	case *tree.FuncExpr:
		__antithesis_instrumentation__.Notify(466403)
		if t.IsDistSQLBlocklist() {
			__antithesis_instrumentation__.Notify(466408)
			v.err = newQueryNotSupportedErrorf("function %s cannot be executed with distsql", t)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(466409)
		}
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(466404)
		v.err = newQueryNotSupportedError("OID expressions are not supported by distsql")
		return false, expr
	case *tree.CastExpr:
		__antithesis_instrumentation__.Notify(466405)

		if typ, ok := tree.GetStaticallyKnownType(t.Type); ok {
			__antithesis_instrumentation__.Notify(466410)
			switch typ.Family() {
			case types.OidFamily:
				__antithesis_instrumentation__.Notify(466411)
				v.err = newQueryNotSupportedErrorf("cast to %s is not supported by distsql", t.Type)
				return false, expr
			default:
				__antithesis_instrumentation__.Notify(466412)
			}
		} else {
			__antithesis_instrumentation__.Notify(466413)
		}
	case *tree.DArray:
		__antithesis_instrumentation__.Notify(466406)

		if t.ResolvedType().ArrayContents() == types.AnyTuple {
			__antithesis_instrumentation__.Notify(466414)
			v.err = newQueryNotSupportedErrorf("array %s cannot be executed with distsql", t)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(466415)
		}
	case *tree.DTuple:
		__antithesis_instrumentation__.Notify(466407)
		if t.ResolvedType() == types.AnyTuple {
			__antithesis_instrumentation__.Notify(466416)
			v.err = newQueryNotSupportedErrorf("tuple %s cannot be executed with distsql", t)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(466417)
		}
	}
	__antithesis_instrumentation__.Notify(466400)
	return true, expr
}

func (v *distSQLExprCheckVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(466418)
	return expr
}

func checkExpr(expr tree.Expr) error {
	__antithesis_instrumentation__.Notify(466419)
	if expr == nil {
		__antithesis_instrumentation__.Notify(466421)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(466422)
	}
	__antithesis_instrumentation__.Notify(466420)
	v := distSQLExprCheckVisitor{}
	tree.WalkExprConst(&v, expr)
	return v.err
}

type distRecommendation int

const (
	cannotDistribute distRecommendation = iota

	shouldNotDistribute

	canDistribute

	shouldDistribute
)

func (a distRecommendation) compose(b distRecommendation) distRecommendation {
	__antithesis_instrumentation__.Notify(466423)
	if a == cannotDistribute || func() bool {
		__antithesis_instrumentation__.Notify(466427)
		return b == cannotDistribute == true
	}() == true {
		__antithesis_instrumentation__.Notify(466428)
		return cannotDistribute
	} else {
		__antithesis_instrumentation__.Notify(466429)
	}
	__antithesis_instrumentation__.Notify(466424)
	if a == shouldNotDistribute || func() bool {
		__antithesis_instrumentation__.Notify(466430)
		return b == shouldNotDistribute == true
	}() == true {
		__antithesis_instrumentation__.Notify(466431)
		return shouldNotDistribute
	} else {
		__antithesis_instrumentation__.Notify(466432)
	}
	__antithesis_instrumentation__.Notify(466425)
	if a == shouldDistribute || func() bool {
		__antithesis_instrumentation__.Notify(466433)
		return b == shouldDistribute == true
	}() == true {
		__antithesis_instrumentation__.Notify(466434)
		return shouldDistribute
	} else {
		__antithesis_instrumentation__.Notify(466435)
	}
	__antithesis_instrumentation__.Notify(466426)
	return canDistribute
}

type queryNotSupportedError struct {
	msg string
}

func (e *queryNotSupportedError) Error() string {
	__antithesis_instrumentation__.Notify(466436)
	return e.msg
}

func newQueryNotSupportedError(msg string) error {
	__antithesis_instrumentation__.Notify(466437)
	return &queryNotSupportedError{msg: msg}
}

func newQueryNotSupportedErrorf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(466438)
	return &queryNotSupportedError{msg: fmt.Sprintf(format, args...)}
}

var planNodeNotSupportedErr = newQueryNotSupportedError("unsupported node")

var cannotDistributeRowLevelLockingErr = newQueryNotSupportedError(
	"scans with row-level locking are not supported by distsql",
)

func (dsp *DistSQLPlanner) mustWrapNode(planCtx *PlanningCtx, node planNode) bool {
	__antithesis_instrumentation__.Notify(466439)
	switch n := node.(type) {

	case *distinctNode:
		__antithesis_instrumentation__.Notify(466441)
	case *exportNode:
		__antithesis_instrumentation__.Notify(466442)
	case *filterNode:
		__antithesis_instrumentation__.Notify(466443)
	case *groupNode:
		__antithesis_instrumentation__.Notify(466444)
	case *indexJoinNode:
		__antithesis_instrumentation__.Notify(466445)
	case *invertedFilterNode:
		__antithesis_instrumentation__.Notify(466446)
	case *invertedJoinNode:
		__antithesis_instrumentation__.Notify(466447)
	case *joinNode:
		__antithesis_instrumentation__.Notify(466448)
	case *limitNode:
		__antithesis_instrumentation__.Notify(466449)
	case *lookupJoinNode:
		__antithesis_instrumentation__.Notify(466450)
	case *ordinalityNode:
		__antithesis_instrumentation__.Notify(466451)
	case *projectSetNode:
		__antithesis_instrumentation__.Notify(466452)
	case *renderNode:
		__antithesis_instrumentation__.Notify(466453)
	case *scanNode:
		__antithesis_instrumentation__.Notify(466454)
	case *sortNode:
		__antithesis_instrumentation__.Notify(466455)
	case *topKNode:
		__antithesis_instrumentation__.Notify(466456)
	case *unaryNode:
		__antithesis_instrumentation__.Notify(466457)
	case *unionNode:
		__antithesis_instrumentation__.Notify(466458)
	case *valuesNode:
		__antithesis_instrumentation__.Notify(466459)
		return mustWrapValuesNode(planCtx, n.specifiedInQuery)
	case *windowNode:
		__antithesis_instrumentation__.Notify(466460)
	case *zeroNode:
		__antithesis_instrumentation__.Notify(466461)
	case *zigzagJoinNode:
		__antithesis_instrumentation__.Notify(466462)
	default:
		__antithesis_instrumentation__.Notify(466463)
		return true
	}
	__antithesis_instrumentation__.Notify(466440)
	return false
}

func mustWrapValuesNode(planCtx *PlanningCtx, specifiedInQuery bool) bool {
	__antithesis_instrumentation__.Notify(466464)

	if !specifiedInQuery || func() bool {
		__antithesis_instrumentation__.Notify(466466)
		return planCtx.isLocal == true
	}() == true {
		__antithesis_instrumentation__.Notify(466467)
		return true
	} else {
		__antithesis_instrumentation__.Notify(466468)
	}
	__antithesis_instrumentation__.Notify(466465)
	return false
}

func checkSupportForPlanNode(node planNode) (distRecommendation, error) {
	__antithesis_instrumentation__.Notify(466469)
	switch n := node.(type) {

	case *distinctNode:
		__antithesis_instrumentation__.Notify(466470)
		return checkSupportForPlanNode(n.plan)

	case *exportNode:
		__antithesis_instrumentation__.Notify(466471)
		return checkSupportForPlanNode(n.source)

	case *filterNode:
		__antithesis_instrumentation__.Notify(466472)
		if err := checkExpr(n.filter); err != nil {
			__antithesis_instrumentation__.Notify(466518)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466519)
		}
		__antithesis_instrumentation__.Notify(466473)
		return checkSupportForPlanNode(n.source.plan)

	case *groupNode:
		__antithesis_instrumentation__.Notify(466474)
		rec, err := checkSupportForPlanNode(n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(466520)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466521)
		}
		__antithesis_instrumentation__.Notify(466475)

		return rec.compose(shouldDistribute), nil

	case *indexJoinNode:
		__antithesis_instrumentation__.Notify(466476)

		if _, err := checkSupportForPlanNode(n.table); err != nil {
			__antithesis_instrumentation__.Notify(466522)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466523)
		}
		__antithesis_instrumentation__.Notify(466477)
		return checkSupportForPlanNode(n.input)

	case *invertedFilterNode:
		__antithesis_instrumentation__.Notify(466478)
		return checkSupportForInvertedFilterNode(n)

	case *invertedJoinNode:
		__antithesis_instrumentation__.Notify(466479)
		if err := checkExpr(n.onExpr); err != nil {
			__antithesis_instrumentation__.Notify(466524)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466525)
		}
		__antithesis_instrumentation__.Notify(466480)
		rec, err := checkSupportForPlanNode(n.input)
		if err != nil {
			__antithesis_instrumentation__.Notify(466526)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466527)
		}
		__antithesis_instrumentation__.Notify(466481)
		return rec.compose(shouldDistribute), nil

	case *joinNode:
		__antithesis_instrumentation__.Notify(466482)
		if err := checkExpr(n.pred.onCond); err != nil {
			__antithesis_instrumentation__.Notify(466528)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466529)
		}
		__antithesis_instrumentation__.Notify(466483)
		recLeft, err := checkSupportForPlanNode(n.left.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(466530)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466531)
		}
		__antithesis_instrumentation__.Notify(466484)
		recRight, err := checkSupportForPlanNode(n.right.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(466532)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466533)
		}
		__antithesis_instrumentation__.Notify(466485)

		rec := recLeft.compose(recRight)

		if len(n.pred.leftEqualityIndices) > 0 {
			__antithesis_instrumentation__.Notify(466534)
			rec = rec.compose(shouldDistribute)
		} else {
			__antithesis_instrumentation__.Notify(466535)
		}
		__antithesis_instrumentation__.Notify(466486)
		return rec, nil

	case *limitNode:
		__antithesis_instrumentation__.Notify(466487)

		return checkSupportForPlanNode(n.plan)

	case *lookupJoinNode:
		__antithesis_instrumentation__.Notify(466488)
		if n.table.lockingStrength != descpb.ScanLockingStrength_FOR_NONE {
			__antithesis_instrumentation__.Notify(466536)

			return cannotDistribute, cannotDistributeRowLevelLockingErr
		} else {
			__antithesis_instrumentation__.Notify(466537)
		}
		__antithesis_instrumentation__.Notify(466489)

		if err := checkExpr(n.lookupExpr); err != nil {
			__antithesis_instrumentation__.Notify(466538)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466539)
		}
		__antithesis_instrumentation__.Notify(466490)
		if err := checkExpr(n.remoteLookupExpr); err != nil {
			__antithesis_instrumentation__.Notify(466540)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466541)
		}
		__antithesis_instrumentation__.Notify(466491)
		if err := checkExpr(n.onCond); err != nil {
			__antithesis_instrumentation__.Notify(466542)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466543)
		}
		__antithesis_instrumentation__.Notify(466492)
		rec, err := checkSupportForPlanNode(n.input)
		if err != nil {
			__antithesis_instrumentation__.Notify(466544)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466545)
		}
		__antithesis_instrumentation__.Notify(466493)
		return rec.compose(canDistribute), nil

	case *ordinalityNode:
		__antithesis_instrumentation__.Notify(466494)

		return cannotDistribute, nil

	case *projectSetNode:
		__antithesis_instrumentation__.Notify(466495)
		return checkSupportForPlanNode(n.source)

	case *renderNode:
		__antithesis_instrumentation__.Notify(466496)
		for _, e := range n.render {
			__antithesis_instrumentation__.Notify(466546)
			if err := checkExpr(e); err != nil {
				__antithesis_instrumentation__.Notify(466547)
				return cannotDistribute, err
			} else {
				__antithesis_instrumentation__.Notify(466548)
			}
		}
		__antithesis_instrumentation__.Notify(466497)
		return checkSupportForPlanNode(n.source.plan)

	case *scanNode:
		__antithesis_instrumentation__.Notify(466498)
		if n.lockingStrength != descpb.ScanLockingStrength_FOR_NONE {
			__antithesis_instrumentation__.Notify(466549)

			return cannotDistribute, cannotDistributeRowLevelLockingErr
		} else {
			__antithesis_instrumentation__.Notify(466550)
		}
		__antithesis_instrumentation__.Notify(466499)

		switch {
		case n.localityOptimized:
			__antithesis_instrumentation__.Notify(466551)

			return cannotDistribute, nil
		case n.isFull:
			__antithesis_instrumentation__.Notify(466552)

			return shouldDistribute, nil
		default:
			__antithesis_instrumentation__.Notify(466553)

			return canDistribute, nil
		}

	case *sortNode:
		__antithesis_instrumentation__.Notify(466500)
		rec, err := checkSupportForPlanNode(n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(466554)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466555)
		}
		__antithesis_instrumentation__.Notify(466501)
		return rec.compose(shouldDistribute), nil

	case *topKNode:
		__antithesis_instrumentation__.Notify(466502)
		rec, err := checkSupportForPlanNode(n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(466556)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466557)
		}
		__antithesis_instrumentation__.Notify(466503)

		return rec.compose(canDistribute), nil

	case *unaryNode:
		__antithesis_instrumentation__.Notify(466504)
		return canDistribute, nil

	case *unionNode:
		__antithesis_instrumentation__.Notify(466505)
		recLeft, err := checkSupportForPlanNode(n.left)
		if err != nil {
			__antithesis_instrumentation__.Notify(466558)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466559)
		}
		__antithesis_instrumentation__.Notify(466506)
		recRight, err := checkSupportForPlanNode(n.right)
		if err != nil {
			__antithesis_instrumentation__.Notify(466560)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466561)
		}
		__antithesis_instrumentation__.Notify(466507)
		return recLeft.compose(recRight), nil

	case *valuesNode:
		__antithesis_instrumentation__.Notify(466508)
		if !n.specifiedInQuery {
			__antithesis_instrumentation__.Notify(466562)

			return cannotDistribute, newQueryNotSupportedErrorf("unsupported valuesNode, not specified in query")
		} else {
			__antithesis_instrumentation__.Notify(466563)
		}
		__antithesis_instrumentation__.Notify(466509)

		for _, tuple := range n.tuples {
			__antithesis_instrumentation__.Notify(466564)
			for _, expr := range tuple {
				__antithesis_instrumentation__.Notify(466565)
				if err := checkExpr(expr); err != nil {
					__antithesis_instrumentation__.Notify(466566)
					return cannotDistribute, err
				} else {
					__antithesis_instrumentation__.Notify(466567)
				}
			}
		}
		__antithesis_instrumentation__.Notify(466510)
		return canDistribute, nil

	case *windowNode:
		__antithesis_instrumentation__.Notify(466511)
		return checkSupportForPlanNode(n.plan)

	case *zeroNode:
		__antithesis_instrumentation__.Notify(466512)
		return canDistribute, nil

	case *zigzagJoinNode:
		__antithesis_instrumentation__.Notify(466513)
		if err := checkExpr(n.onCond); err != nil {
			__antithesis_instrumentation__.Notify(466568)
			return cannotDistribute, err
		} else {
			__antithesis_instrumentation__.Notify(466569)
		}
		__antithesis_instrumentation__.Notify(466514)
		return shouldDistribute, nil

	case *createStatsNode:
		__antithesis_instrumentation__.Notify(466515)
		if n.runAsJob {
			__antithesis_instrumentation__.Notify(466570)
			return cannotDistribute, planNodeNotSupportedErr
		} else {
			__antithesis_instrumentation__.Notify(466571)
		}
		__antithesis_instrumentation__.Notify(466516)
		return shouldDistribute, nil

	default:
		__antithesis_instrumentation__.Notify(466517)
		return cannotDistribute, planNodeNotSupportedErr
	}
}

func checkSupportForInvertedFilterNode(n *invertedFilterNode) (distRecommendation, error) {
	__antithesis_instrumentation__.Notify(466572)
	rec, err := checkSupportForPlanNode(n.input)
	if err != nil {
		__antithesis_instrumentation__.Notify(466575)
		return cannotDistribute, err
	} else {
		__antithesis_instrumentation__.Notify(466576)
	}
	__antithesis_instrumentation__.Notify(466573)

	filterRec := cannotDistribute
	if n.expression.Left == nil && func() bool {
		__antithesis_instrumentation__.Notify(466577)
		return n.expression.Right == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(466578)
		filterRec = shouldDistribute
	} else {
		__antithesis_instrumentation__.Notify(466579)
	}
	__antithesis_instrumentation__.Notify(466574)
	return rec.compose(filterRec), nil
}

type NodeStatus int

const (
	NodeOK NodeStatus = iota

	NodeUnhealthy

	NodeDistSQLVersionIncompatible
)

type PlanningCtx struct {
	ExtendedEvalCtx *extendedEvalContext
	spanIter        physicalplan.SpanResolverIterator

	NodeStatuses map[base.SQLInstanceID]NodeStatus

	infra physicalplan.PhysicalInfrastructure

	isLocal bool
	planner *planner

	ignoreClose bool
	stmtType    tree.StatementReturnType

	planDepth int

	saveFlows func(map[base.SQLInstanceID]*execinfrapb.FlowSpec, execinfra.OpChains) error

	traceMetadata execNodeTraceMetadata

	collectExecStats bool

	parallelizeScansIfLocal bool

	onFlowCleanup []func()
}

var _ physicalplan.ExprContext = &PlanningCtx{}

func (p *PlanningCtx) NewPhysicalPlan() *PhysicalPlan {
	__antithesis_instrumentation__.Notify(466580)
	return &PhysicalPlan{
		PhysicalPlan: physicalplan.MakePhysicalPlan(&p.infra),
	}
}

func (p *PlanningCtx) EvalContext() *tree.EvalContext {
	__antithesis_instrumentation__.Notify(466581)
	if p.ExtendedEvalCtx == nil {
		__antithesis_instrumentation__.Notify(466583)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(466584)
	}
	__antithesis_instrumentation__.Notify(466582)
	return &p.ExtendedEvalCtx.EvalContext
}

func (p *PlanningCtx) IsLocal() bool {
	__antithesis_instrumentation__.Notify(466585)
	return p.isLocal
}

func (p *PlanningCtx) getDefaultSaveFlowsFunc(
	ctx context.Context, planner *planner, typ planComponentType,
) func(map[base.SQLInstanceID]*execinfrapb.FlowSpec, execinfra.OpChains) error {
	__antithesis_instrumentation__.Notify(466586)
	return func(flows map[base.SQLInstanceID]*execinfrapb.FlowSpec, opChains execinfra.OpChains) error {
		__antithesis_instrumentation__.Notify(466587)
		var diagram execinfrapb.FlowDiagram
		if planner.instrumentation.shouldSaveDiagrams() {
			__antithesis_instrumentation__.Notify(466590)
			diagramFlags := execinfrapb.DiagramFlags{
				MakeDeterministic: planner.execCfg.TestingKnobs.DeterministicExplain,
			}
			var err error
			diagram, err = p.flowSpecsToDiagram(ctx, flows, diagramFlags)
			if err != nil {
				__antithesis_instrumentation__.Notify(466591)
				return err
			} else {
				__antithesis_instrumentation__.Notify(466592)
			}
		} else {
			__antithesis_instrumentation__.Notify(466593)
		}
		__antithesis_instrumentation__.Notify(466588)
		var explainVec []string
		var explainVecVerbose []string
		if planner.instrumentation.collectBundle && func() bool {
			__antithesis_instrumentation__.Notify(466594)
			return planner.curPlan.flags.IsSet(planFlagVectorized) == true
		}() == true {
			__antithesis_instrumentation__.Notify(466595)
			flowCtx := newFlowCtxForExplainPurposes(p, planner)
			getExplain := func(verbose bool) []string {
				__antithesis_instrumentation__.Notify(466597)
				explain, cleanup, err := colflow.ExplainVec(
					ctx, flowCtx, flows, p.infra.LocalProcessors, opChains,
					planner.extendedEvalCtx.DistSQLPlanner.gatewaySQLInstanceID,
					verbose, planner.curPlan.flags.IsDistributed(),
				)
				cleanup()
				if err != nil {
					__antithesis_instrumentation__.Notify(466599)

					explain = nil
				} else {
					__antithesis_instrumentation__.Notify(466600)
				}
				__antithesis_instrumentation__.Notify(466598)
				return explain
			}
			__antithesis_instrumentation__.Notify(466596)
			explainVec = getExplain(false)
			explainVecVerbose = getExplain(true)
		} else {
			__antithesis_instrumentation__.Notify(466601)
		}
		__antithesis_instrumentation__.Notify(466589)
		planner.curPlan.distSQLFlowInfos = append(
			planner.curPlan.distSQLFlowInfos,
			flowInfo{
				typ:               typ,
				diagram:           diagram,
				explainVec:        explainVec,
				explainVecVerbose: explainVecVerbose,
				flowsMetadata:     execstats.NewFlowsMetadata(flows),
			},
		)
		return nil
	}
}

func (p *PlanningCtx) flowSpecsToDiagram(
	ctx context.Context,
	flows map[base.SQLInstanceID]*execinfrapb.FlowSpec,
	diagramFlags execinfrapb.DiagramFlags,
) (execinfrapb.FlowDiagram, error) {
	__antithesis_instrumentation__.Notify(466602)
	log.VEvent(ctx, 1, "creating plan diagram")
	var stmtStr string
	if p.planner != nil && func() bool {
		__antithesis_instrumentation__.Notify(466605)
		return p.planner.stmt.AST != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(466606)
		stmtStr = p.planner.stmt.String()
	} else {
		__antithesis_instrumentation__.Notify(466607)
	}
	__antithesis_instrumentation__.Notify(466603)
	diagram, err := execinfrapb.GeneratePlanDiagram(
		stmtStr, flows, diagramFlags,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466608)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(466609)
	}
	__antithesis_instrumentation__.Notify(466604)
	return diagram, nil
}

func (p *PlanningCtx) getCleanupFunc() func() {
	__antithesis_instrumentation__.Notify(466610)
	return func() {
		__antithesis_instrumentation__.Notify(466611)
		for _, r := range p.onFlowCleanup {
			__antithesis_instrumentation__.Notify(466612)
			r()
		}
	}
}

type PhysicalPlan struct {
	physicalplan.PhysicalPlan

	PlanToStreamColMap []int
}

func makePlanToStreamColMap(numCols int) []int {
	__antithesis_instrumentation__.Notify(466613)
	m := make([]int, numCols)
	for i := 0; i < numCols; i++ {
		__antithesis_instrumentation__.Notify(466615)
		m[i] = -1
	}
	__antithesis_instrumentation__.Notify(466614)
	return m
}

func identityMap(buf []int, numCols int) []int {
	__antithesis_instrumentation__.Notify(466616)
	buf = buf[:0]
	for i := 0; i < numCols; i++ {
		__antithesis_instrumentation__.Notify(466618)
		buf = append(buf, i)
	}
	__antithesis_instrumentation__.Notify(466617)
	return buf
}

func identityMapInPlace(slice []int) []int {
	__antithesis_instrumentation__.Notify(466619)
	for i := range slice {
		__antithesis_instrumentation__.Notify(466621)
		slice[i] = i
	}
	__antithesis_instrumentation__.Notify(466620)
	return slice
}

type SpanPartition struct {
	SQLInstanceID base.SQLInstanceID
	Spans         roachpb.Spans
}

type distSQLNodeHealth struct {
	gossip      gossip.OptionalGossip
	isAvailable func(base.SQLInstanceID) bool
	connHealth  func(roachpb.NodeID, rpc.ConnectionClass) error
}

func (h *distSQLNodeHealth) check(ctx context.Context, sqlInstanceID base.SQLInstanceID) error {
	__antithesis_instrumentation__.Notify(466622)
	{
		__antithesis_instrumentation__.Notify(466626)

		err := h.connHealth(roachpb.NodeID(sqlInstanceID), rpc.DefaultClass)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(466627)
			return !errors.Is(err, rpc.ErrNotHeartbeated) == true
		}() == true {
			__antithesis_instrumentation__.Notify(466628)

			log.VEventf(ctx, 1, "marking n%d as unhealthy for this plan: %v", sqlInstanceID, err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(466629)
		}
	}
	__antithesis_instrumentation__.Notify(466623)
	if !h.isAvailable(sqlInstanceID) {
		__antithesis_instrumentation__.Notify(466630)
		return pgerror.Newf(pgcode.CannotConnectNow, "not using n%d since it is not available", sqlInstanceID)
	} else {
		__antithesis_instrumentation__.Notify(466631)
	}
	__antithesis_instrumentation__.Notify(466624)

	if g, ok := h.gossip.Optional(distsql.MultiTenancyIssueNo); ok {
		__antithesis_instrumentation__.Notify(466632)
		drainingInfo := &execinfrapb.DistSQLDrainingInfo{}
		if err := g.GetInfoProto(gossip.MakeDistSQLDrainingKey(sqlInstanceID), drainingInfo); err != nil {
			__antithesis_instrumentation__.Notify(466634)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(466635)
		}
		__antithesis_instrumentation__.Notify(466633)

		if drainingInfo.Draining {
			__antithesis_instrumentation__.Notify(466636)
			err := errors.Newf("not using n%d because it is draining", sqlInstanceID)
			log.VEventf(ctx, 1, "%v", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(466637)
		}
	} else {
		__antithesis_instrumentation__.Notify(466638)
	}
	__antithesis_instrumentation__.Notify(466625)

	return nil
}

func (dsp *DistSQLPlanner) PartitionSpans(
	ctx context.Context, planCtx *PlanningCtx, spans roachpb.Spans,
) ([]SpanPartition, error) {
	__antithesis_instrumentation__.Notify(466639)
	if dsp.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(466641)
		return dsp.partitionSpansSystem(ctx, planCtx, spans)
	} else {
		__antithesis_instrumentation__.Notify(466642)
	}
	__antithesis_instrumentation__.Notify(466640)
	return dsp.partitionSpansTenant(ctx, planCtx, spans)
}

func (dsp *DistSQLPlanner) partitionSpansSystem(
	ctx context.Context, planCtx *PlanningCtx, spans roachpb.Spans,
) ([]SpanPartition, error) {
	__antithesis_instrumentation__.Notify(466643)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(466647)
		panic("no spans")
	} else {
		__antithesis_instrumentation__.Notify(466648)
	}
	__antithesis_instrumentation__.Notify(466644)
	partitions := make([]SpanPartition, 0, 1)
	if planCtx.isLocal {
		__antithesis_instrumentation__.Notify(466649)

		partitions = append(partitions,
			SpanPartition{dsp.gatewaySQLInstanceID, spans})
		return partitions, nil
	} else {
		__antithesis_instrumentation__.Notify(466650)
	}
	__antithesis_instrumentation__.Notify(466645)

	nodeMap := make(map[base.SQLInstanceID]int)
	it := planCtx.spanIter
	for i := range spans {
		__antithesis_instrumentation__.Notify(466651)

		span := spans[i]
		noEndKey := false
		if len(span.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(466655)

			span = roachpb.Span{
				Key:    span.Key,
				EndKey: span.Key.Next(),
			}
			noEndKey = true
		} else {
			__antithesis_instrumentation__.Notify(466656)
		}
		__antithesis_instrumentation__.Notify(466652)

		rSpan, err := keys.SpanAddr(span)
		if err != nil {
			__antithesis_instrumentation__.Notify(466657)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(466658)
		}
		__antithesis_instrumentation__.Notify(466653)

		var lastSQLInstanceID base.SQLInstanceID

		lastKey := rSpan.Key
		if log.V(1) {
			__antithesis_instrumentation__.Notify(466659)
			log.Infof(ctx, "partitioning span %s", span)
		} else {
			__antithesis_instrumentation__.Notify(466660)
		}
		__antithesis_instrumentation__.Notify(466654)

		for it.Seek(ctx, span, kvcoord.Ascending); ; it.Next(ctx) {
			__antithesis_instrumentation__.Notify(466661)
			if !it.Valid() {
				__antithesis_instrumentation__.Notify(466671)
				return nil, it.Error()
			} else {
				__antithesis_instrumentation__.Notify(466672)
			}
			__antithesis_instrumentation__.Notify(466662)
			replDesc, err := it.ReplicaInfo(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(466673)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(466674)
			}
			__antithesis_instrumentation__.Notify(466663)
			desc := it.Desc()
			if log.V(1) {
				__antithesis_instrumentation__.Notify(466675)
				descCpy := desc
				log.Infof(ctx, "lastKey: %s desc: %s", lastKey, &descCpy)
			} else {
				__antithesis_instrumentation__.Notify(466676)
			}
			__antithesis_instrumentation__.Notify(466664)

			if !desc.ContainsKey(lastKey) {
				__antithesis_instrumentation__.Notify(466677)

				log.Fatalf(
					ctx, "next range %v doesn't cover last end key %v. Partitions: %#v",
					desc.RSpan(), lastKey, partitions,
				)
			} else {
				__antithesis_instrumentation__.Notify(466678)
			}
			__antithesis_instrumentation__.Notify(466665)

			endKey := desc.EndKey
			if rSpan.EndKey.Less(endKey) {
				__antithesis_instrumentation__.Notify(466679)
				endKey = rSpan.EndKey
			} else {
				__antithesis_instrumentation__.Notify(466680)
			}
			__antithesis_instrumentation__.Notify(466666)

			sqlInstanceID := base.SQLInstanceID(replDesc.NodeID)
			partitionIdx, inNodeMap := nodeMap[sqlInstanceID]
			if !inNodeMap {
				__antithesis_instrumentation__.Notify(466681)

				status := dsp.CheckInstanceHealthAndVersion(ctx, planCtx, sqlInstanceID)

				if status != NodeOK {
					__antithesis_instrumentation__.Notify(466683)
					log.Eventf(ctx, "not planning on node %d: %s", sqlInstanceID, status)
					sqlInstanceID = dsp.gatewaySQLInstanceID
					partitionIdx, inNodeMap = nodeMap[sqlInstanceID]
				} else {
					__antithesis_instrumentation__.Notify(466684)
				}
				__antithesis_instrumentation__.Notify(466682)

				if !inNodeMap {
					__antithesis_instrumentation__.Notify(466685)
					partitionIdx = len(partitions)
					partitions = append(partitions, SpanPartition{SQLInstanceID: sqlInstanceID})
					nodeMap[sqlInstanceID] = partitionIdx
				} else {
					__antithesis_instrumentation__.Notify(466686)
				}
			} else {
				__antithesis_instrumentation__.Notify(466687)
			}
			__antithesis_instrumentation__.Notify(466667)
			partition := &partitions[partitionIdx]

			if noEndKey {
				__antithesis_instrumentation__.Notify(466688)

				partition.Spans = append(partition.Spans, roachpb.Span{
					Key: lastKey.AsRawKey(),
				})
				break
			} else {
				__antithesis_instrumentation__.Notify(466689)
			}
			__antithesis_instrumentation__.Notify(466668)

			if lastSQLInstanceID == sqlInstanceID {
				__antithesis_instrumentation__.Notify(466690)

				partition.Spans[len(partition.Spans)-1].EndKey = endKey.AsRawKey()
			} else {
				__antithesis_instrumentation__.Notify(466691)
				partition.Spans = append(partition.Spans, roachpb.Span{
					Key:    lastKey.AsRawKey(),
					EndKey: endKey.AsRawKey(),
				})
			}
			__antithesis_instrumentation__.Notify(466669)

			if !endKey.Less(rSpan.EndKey) {
				__antithesis_instrumentation__.Notify(466692)

				break
			} else {
				__antithesis_instrumentation__.Notify(466693)
			}
			__antithesis_instrumentation__.Notify(466670)

			lastKey = endKey
			lastSQLInstanceID = sqlInstanceID
		}
	}
	__antithesis_instrumentation__.Notify(466646)
	return partitions, nil
}

func (dsp *DistSQLPlanner) partitionSpansTenant(
	ctx context.Context, planCtx *PlanningCtx, spans roachpb.Spans,
) ([]SpanPartition, error) {
	__antithesis_instrumentation__.Notify(466694)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(466702)
		panic("no spans")
	} else {
		__antithesis_instrumentation__.Notify(466703)
	}
	__antithesis_instrumentation__.Notify(466695)
	partitions := make([]SpanPartition, 0, 1)
	if planCtx.isLocal {
		__antithesis_instrumentation__.Notify(466704)

		partitions = append(partitions,
			SpanPartition{dsp.gatewaySQLInstanceID, spans})
		return partitions, nil
	} else {
		__antithesis_instrumentation__.Notify(466705)
	}
	__antithesis_instrumentation__.Notify(466696)
	if dsp.sqlInstanceProvider == nil {
		__antithesis_instrumentation__.Notify(466706)
		return nil, errors.New("sql instance provider not available in multi-tenant environment")
	} else {
		__antithesis_instrumentation__.Notify(466707)
	}
	__antithesis_instrumentation__.Notify(466697)

	instances, err := dsp.sqlInstanceProvider.GetAllInstances(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(466708)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(466709)
	}
	__antithesis_instrumentation__.Notify(466698)
	if len(instances) == 0 {
		__antithesis_instrumentation__.Notify(466710)
		return nil, errors.New("no healthy sql instances available for planning")
	} else {
		__antithesis_instrumentation__.Notify(466711)
	}
	__antithesis_instrumentation__.Notify(466699)

	rand.Shuffle(len(instances), func(i, j int) {
		__antithesis_instrumentation__.Notify(466712)
		instances[i], instances[j] = instances[j], instances[i]
	})
	__antithesis_instrumentation__.Notify(466700)

	nodeMap := make(map[base.SQLInstanceID]int)
	for i := range spans {
		__antithesis_instrumentation__.Notify(466713)
		span := spans[i]
		if log.V(1) {
			__antithesis_instrumentation__.Notify(466716)
			log.Infof(ctx, "partitioning span %s", span)
		} else {
			__antithesis_instrumentation__.Notify(466717)
		}
		__antithesis_instrumentation__.Notify(466714)
		sqlInstanceID := instances[i%len(instances)].InstanceID
		partitionIdx, inNodeMap := nodeMap[sqlInstanceID]
		if !inNodeMap {
			__antithesis_instrumentation__.Notify(466718)
			partitionIdx = len(partitions)
			partitions = append(partitions, SpanPartition{SQLInstanceID: sqlInstanceID})
			nodeMap[sqlInstanceID] = partitionIdx
		} else {
			__antithesis_instrumentation__.Notify(466719)
		}
		__antithesis_instrumentation__.Notify(466715)
		partition := &partitions[partitionIdx]
		partition.Spans = append(partition.Spans, span)
	}
	__antithesis_instrumentation__.Notify(466701)
	return partitions, nil
}

func (dsp *DistSQLPlanner) nodeVersionIsCompatible(sqlInstanceID base.SQLInstanceID) bool {
	__antithesis_instrumentation__.Notify(466720)
	g, ok := dsp.gossip.Optional(distsql.MultiTenancyIssueNo)
	if !ok {
		__antithesis_instrumentation__.Notify(466723)
		return true
	} else {
		__antithesis_instrumentation__.Notify(466724)
	}
	__antithesis_instrumentation__.Notify(466721)
	var v execinfrapb.DistSQLVersionGossipInfo
	if err := g.GetInfoProto(gossip.MakeDistSQLNodeVersionKey(sqlInstanceID), &v); err != nil {
		__antithesis_instrumentation__.Notify(466725)
		return false
	} else {
		__antithesis_instrumentation__.Notify(466726)
	}
	__antithesis_instrumentation__.Notify(466722)
	return distsql.FlowVerIsCompatible(dsp.planVersion, v.MinAcceptedVersion, v.Version)
}

func getIndexIdx(index catalog.Index, desc catalog.TableDescriptor) (uint32, error) {
	__antithesis_instrumentation__.Notify(466727)
	if index.Public() {
		__antithesis_instrumentation__.Notify(466729)
		return uint32(index.Ordinal()), nil
	} else {
		__antithesis_instrumentation__.Notify(466730)
	}
	__antithesis_instrumentation__.Notify(466728)
	return 0, errors.Errorf("invalid index %v (table %s)", index, desc.GetName())
}

func initTableReaderSpecTemplate(
	n *scanNode, codec keys.SQLCodec,
) (*execinfrapb.TableReaderSpec, execinfrapb.PostProcessSpec, error) {
	__antithesis_instrumentation__.Notify(466731)
	if n.isCheck {
		__antithesis_instrumentation__.Notify(466736)
		return nil, execinfrapb.PostProcessSpec{}, errors.AssertionFailedf("isCheck no longer supported")
	} else {
		__antithesis_instrumentation__.Notify(466737)
	}
	__antithesis_instrumentation__.Notify(466732)
	colIDs := make([]descpb.ColumnID, len(n.cols))
	for i := range n.cols {
		__antithesis_instrumentation__.Notify(466738)
		colIDs[i] = n.cols[i].GetID()
	}
	__antithesis_instrumentation__.Notify(466733)
	s := physicalplan.NewTableReaderSpec()
	*s = execinfrapb.TableReaderSpec{
		Reverse:                         n.reverse,
		TableDescriptorModificationTime: n.desc.GetModificationTime(),
		LockingStrength:                 n.lockingStrength,
		LockingWaitPolicy:               n.lockingWaitPolicy,
	}
	if err := rowenc.InitIndexFetchSpec(&s.FetchSpec, codec, n.desc, n.index, colIDs); err != nil {
		__antithesis_instrumentation__.Notify(466739)
		return nil, execinfrapb.PostProcessSpec{}, err
	} else {
		__antithesis_instrumentation__.Notify(466740)
	}
	__antithesis_instrumentation__.Notify(466734)

	var post execinfrapb.PostProcessSpec
	if n.hardLimit != 0 {
		__antithesis_instrumentation__.Notify(466741)
		post.Limit = uint64(n.hardLimit)
	} else {
		__antithesis_instrumentation__.Notify(466742)
		if n.softLimit != 0 {
			__antithesis_instrumentation__.Notify(466743)
			s.LimitHint = n.softLimit
		} else {
			__antithesis_instrumentation__.Notify(466744)
		}
	}
	__antithesis_instrumentation__.Notify(466735)
	return s, post, nil
}

func tableOrdinal(desc catalog.TableDescriptor, colID descpb.ColumnID) int {
	__antithesis_instrumentation__.Notify(466745)
	col, _ := desc.FindColumnWithID(colID)
	if col == nil {
		__antithesis_instrumentation__.Notify(466747)
		panic(errors.AssertionFailedf("column %d not in desc.Columns", colID))
	} else {
		__antithesis_instrumentation__.Notify(466748)
	}
	__antithesis_instrumentation__.Notify(466746)
	return col.Ordinal()
}

func (dsp *DistSQLPlanner) convertOrdering(
	reqOrdering ReqOrdering, planToStreamColMap []int,
) execinfrapb.Ordering {
	__antithesis_instrumentation__.Notify(466749)
	if len(reqOrdering) == 0 {
		__antithesis_instrumentation__.Notify(466752)
		return execinfrapb.Ordering{}
	} else {
		__antithesis_instrumentation__.Notify(466753)
	}
	__antithesis_instrumentation__.Notify(466750)
	result := execinfrapb.Ordering{
		Columns: make([]execinfrapb.Ordering_Column, len(reqOrdering)),
	}
	for i, o := range reqOrdering {
		__antithesis_instrumentation__.Notify(466754)
		streamColIdx := o.ColIdx
		if planToStreamColMap != nil {
			__antithesis_instrumentation__.Notify(466758)
			streamColIdx = planToStreamColMap[o.ColIdx]
		} else {
			__antithesis_instrumentation__.Notify(466759)
		}
		__antithesis_instrumentation__.Notify(466755)
		if streamColIdx == -1 {
			__antithesis_instrumentation__.Notify(466760)
			panic("column in ordering not part of processor output")
		} else {
			__antithesis_instrumentation__.Notify(466761)
		}
		__antithesis_instrumentation__.Notify(466756)
		result.Columns[i].ColIdx = uint32(streamColIdx)
		dir := execinfrapb.Ordering_Column_ASC
		if o.Direction == encoding.Descending {
			__antithesis_instrumentation__.Notify(466762)
			dir = execinfrapb.Ordering_Column_DESC
		} else {
			__antithesis_instrumentation__.Notify(466763)
		}
		__antithesis_instrumentation__.Notify(466757)
		result.Columns[i].Direction = dir
	}
	__antithesis_instrumentation__.Notify(466751)
	return result
}

func (dsp *DistSQLPlanner) getInstanceIDForScan(
	ctx context.Context, planCtx *PlanningCtx, spans []roachpb.Span, reverse bool,
) (base.SQLInstanceID, error) {
	__antithesis_instrumentation__.Notify(466764)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(466770)
		panic("no spans")
	} else {
		__antithesis_instrumentation__.Notify(466771)
	}
	__antithesis_instrumentation__.Notify(466765)

	it := planCtx.spanIter
	if reverse {
		__antithesis_instrumentation__.Notify(466772)
		it.Seek(ctx, spans[len(spans)-1], kvcoord.Descending)
	} else {
		__antithesis_instrumentation__.Notify(466773)
		it.Seek(ctx, spans[0], kvcoord.Ascending)
	}
	__antithesis_instrumentation__.Notify(466766)
	if !it.Valid() {
		__antithesis_instrumentation__.Notify(466774)
		return 0, it.Error()
	} else {
		__antithesis_instrumentation__.Notify(466775)
	}
	__antithesis_instrumentation__.Notify(466767)
	replDesc, err := it.ReplicaInfo(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(466776)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(466777)
	}
	__antithesis_instrumentation__.Notify(466768)

	sqlInstanceID := base.SQLInstanceID(replDesc.NodeID)
	status := dsp.CheckInstanceHealthAndVersion(ctx, planCtx, sqlInstanceID)
	if status != NodeOK {
		__antithesis_instrumentation__.Notify(466778)
		log.Eventf(ctx, "not planning on node %d: %s", sqlInstanceID, status)
		return dsp.gatewaySQLInstanceID, nil
	} else {
		__antithesis_instrumentation__.Notify(466779)
	}
	__antithesis_instrumentation__.Notify(466769)
	return sqlInstanceID, nil
}

func (dsp *DistSQLPlanner) CheckInstanceHealthAndVersion(
	ctx context.Context, planCtx *PlanningCtx, sqlInstanceID base.SQLInstanceID,
) NodeStatus {
	__antithesis_instrumentation__.Notify(466780)
	if status, ok := planCtx.NodeStatuses[sqlInstanceID]; ok {
		__antithesis_instrumentation__.Notify(466783)
		return status
	} else {
		__antithesis_instrumentation__.Notify(466784)
	}
	__antithesis_instrumentation__.Notify(466781)

	var status NodeStatus
	if err := dsp.nodeHealth.check(ctx, sqlInstanceID); err != nil {
		__antithesis_instrumentation__.Notify(466785)
		status = NodeUnhealthy
	} else {
		__antithesis_instrumentation__.Notify(466786)
		if !dsp.nodeVersionIsCompatible(sqlInstanceID) {
			__antithesis_instrumentation__.Notify(466787)
			status = NodeDistSQLVersionIncompatible
		} else {
			__antithesis_instrumentation__.Notify(466788)
			status = NodeOK
		}
	}
	__antithesis_instrumentation__.Notify(466782)
	planCtx.NodeStatuses[sqlInstanceID] = status
	return status
}

func (dsp *DistSQLPlanner) createTableReaders(
	ctx context.Context, planCtx *PlanningCtx, n *scanNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(466789)
	spec, post, err := initTableReaderSpecTemplate(n, planCtx.ExtendedEvalCtx.Codec)
	if err != nil {
		__antithesis_instrumentation__.Notify(466791)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(466792)
	}
	__antithesis_instrumentation__.Notify(466790)

	p := planCtx.NewPhysicalPlan()
	err = dsp.planTableReaders(
		ctx,
		planCtx,
		p,
		&tableReaderPlanningInfo{
			spec:              spec,
			post:              post,
			desc:              n.desc,
			spans:             n.spans,
			reverse:           n.reverse,
			parallelize:       n.parallelize,
			estimatedRowCount: n.estimatedRowCount,
			reqOrdering:       n.reqOrdering,
		},
	)
	return p, err
}

type tableReaderPlanningInfo struct {
	spec              *execinfrapb.TableReaderSpec
	post              execinfrapb.PostProcessSpec
	desc              catalog.TableDescriptor
	spans             []roachpb.Span
	reverse           bool
	parallelize       bool
	estimatedRowCount uint64
	reqOrdering       ReqOrdering
}

const defaultLocalScansConcurrencyLimit = 1024

var localScansConcurrencyLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.local_scans.concurrency_limit",
	"maximum number of additional goroutines for performing scans in local plans",
	defaultLocalScansConcurrencyLimit,
	settings.NonNegativeInt,
)

func (dsp *DistSQLPlanner) maybeParallelizeLocalScans(
	ctx context.Context, planCtx *PlanningCtx, info *tableReaderPlanningInfo,
) (spanPartitions []SpanPartition, parallelizeLocal bool) {
	__antithesis_instrumentation__.Notify(466793)

	sd := planCtx.ExtendedEvalCtx.EvalContext.SessionData()

	prohibitParallelScans := sd.LocalityOptimizedSearch && func() bool {
		__antithesis_instrumentation__.Notify(466795)
		return sd.VectorizeMode == sessiondatapb.VectorizeOff == true
	}() == true
	if len(info.reqOrdering) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(466796)
		return info.parallelize == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(466797)
		return planCtx.parallelizeScansIfLocal == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(466798)
		return !prohibitParallelScans == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(466799)
		return dsp.parallelLocalScansSem.ApproximateQuota() > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(466800)
		return planCtx.spanIter != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(466801)
		parallelizeLocal = true

		planCtx.isLocal = false
		var err error
		spanPartitions, err = dsp.PartitionSpans(ctx, planCtx, info.spans)
		planCtx.isLocal = true
		if err != nil {
			__antithesis_instrumentation__.Notify(466805)

			spanPartitions = []SpanPartition{{dsp.gatewaySQLInstanceID, info.spans}}
			parallelizeLocal = false
			return spanPartitions, parallelizeLocal
		} else {
			__antithesis_instrumentation__.Notify(466806)
		}
		__antithesis_instrumentation__.Notify(466802)
		for i := range spanPartitions {
			__antithesis_instrumentation__.Notify(466807)
			spanPartitions[i].SQLInstanceID = dsp.gatewaySQLInstanceID
		}
		__antithesis_instrumentation__.Notify(466803)
		if len(spanPartitions) > 1 {
			__antithesis_instrumentation__.Notify(466808)

			const maxConcurrency = 64
			actualConcurrency := len(spanPartitions)
			if actualConcurrency > maxConcurrency {
				__antithesis_instrumentation__.Notify(466812)
				actualConcurrency = maxConcurrency
			} else {
				__antithesis_instrumentation__.Notify(466813)
			}
			__antithesis_instrumentation__.Notify(466809)
			if quota := int(dsp.parallelLocalScansSem.ApproximateQuota()); actualConcurrency > quota {
				__antithesis_instrumentation__.Notify(466814)
				actualConcurrency = quota
			} else {
				__antithesis_instrumentation__.Notify(466815)
			}
			__antithesis_instrumentation__.Notify(466810)

			var alloc *quotapool.IntAlloc
			for actualConcurrency > 1 {
				__antithesis_instrumentation__.Notify(466816)
				alloc, err = dsp.parallelLocalScansSem.TryAcquire(ctx, uint64(actualConcurrency-1))
				if err == nil {
					__antithesis_instrumentation__.Notify(466818)
					break
				} else {
					__antithesis_instrumentation__.Notify(466819)
				}
				__antithesis_instrumentation__.Notify(466817)
				actualConcurrency--
			}
			__antithesis_instrumentation__.Notify(466811)

			if actualConcurrency > 1 {
				__antithesis_instrumentation__.Notify(466820)

				for extraPartitionIdx := actualConcurrency; extraPartitionIdx < len(spanPartitions); extraPartitionIdx++ {
					__antithesis_instrumentation__.Notify(466822)
					mergeIntoIdx := extraPartitionIdx % actualConcurrency
					spanPartitions[mergeIntoIdx].Spans = append(spanPartitions[mergeIntoIdx].Spans, spanPartitions[extraPartitionIdx].Spans...)
				}
				__antithesis_instrumentation__.Notify(466821)
				spanPartitions = spanPartitions[:actualConcurrency]
				planCtx.onFlowCleanup = append(planCtx.onFlowCleanup, alloc.Release)
			} else {
				__antithesis_instrumentation__.Notify(466823)

				spanPartitions = []SpanPartition{{dsp.gatewaySQLInstanceID, info.spans}}
			}
		} else {
			__antithesis_instrumentation__.Notify(466824)
		}
		__antithesis_instrumentation__.Notify(466804)
		if len(spanPartitions) == 1 {
			__antithesis_instrumentation__.Notify(466825)

			parallelizeLocal = false
		} else {
			__antithesis_instrumentation__.Notify(466826)
		}
	} else {
		__antithesis_instrumentation__.Notify(466827)
		spanPartitions = []SpanPartition{{dsp.gatewaySQLInstanceID, info.spans}}
	}
	__antithesis_instrumentation__.Notify(466794)
	return spanPartitions, parallelizeLocal
}

func (dsp *DistSQLPlanner) planTableReaders(
	ctx context.Context, planCtx *PlanningCtx, p *PhysicalPlan, info *tableReaderPlanningInfo,
) error {
	__antithesis_instrumentation__.Notify(466828)
	var (
		spanPartitions   []SpanPartition
		parallelizeLocal bool
		err              error
	)
	if planCtx.isLocal {
		__antithesis_instrumentation__.Notify(466833)
		spanPartitions, parallelizeLocal = dsp.maybeParallelizeLocalScans(ctx, planCtx, info)
	} else {
		__antithesis_instrumentation__.Notify(466834)
		if info.post.Limit == 0 {
			__antithesis_instrumentation__.Notify(466835)

			spanPartitions, err = dsp.PartitionSpans(ctx, planCtx, info.spans)
			if err != nil {
				__antithesis_instrumentation__.Notify(466836)
				return err
			} else {
				__antithesis_instrumentation__.Notify(466837)
			}
		} else {
			__antithesis_instrumentation__.Notify(466838)

			sqlInstanceID, err := dsp.getInstanceIDForScan(ctx, planCtx, info.spans, info.reverse)
			if err != nil {
				__antithesis_instrumentation__.Notify(466840)
				return err
			} else {
				__antithesis_instrumentation__.Notify(466841)
			}
			__antithesis_instrumentation__.Notify(466839)
			spanPartitions = []SpanPartition{{sqlInstanceID, info.spans}}
		}
	}
	__antithesis_instrumentation__.Notify(466829)

	corePlacement := make([]physicalplan.ProcessorCorePlacement, len(spanPartitions))
	for i, sp := range spanPartitions {
		__antithesis_instrumentation__.Notify(466842)
		var tr *execinfrapb.TableReaderSpec
		if i == 0 {
			__antithesis_instrumentation__.Notify(466845)

			tr = info.spec
		} else {
			__antithesis_instrumentation__.Notify(466846)

			tr = physicalplan.NewTableReaderSpec()
			*tr = *info.spec
		}
		__antithesis_instrumentation__.Notify(466843)

		tr.Spans = sp.Spans

		tr.Parallelize = info.parallelize
		if !tr.Parallelize {
			__antithesis_instrumentation__.Notify(466847)
			tr.BatchBytesLimit = dsp.distSQLSrv.TestingKnobs.TableReaderBatchBytesLimit
		} else {
			__antithesis_instrumentation__.Notify(466848)
		}
		__antithesis_instrumentation__.Notify(466844)
		p.TotalEstimatedScannedRows += info.estimatedRowCount

		corePlacement[i].SQLInstanceID = sp.SQLInstanceID
		corePlacement[i].EstimatedRowCount = info.estimatedRowCount
		corePlacement[i].Core.TableReader = tr
	}
	__antithesis_instrumentation__.Notify(466830)

	typs := make([]*types.T, len(info.spec.FetchSpec.FetchedColumns))
	for i := range typs {
		__antithesis_instrumentation__.Notify(466849)
		typs[i] = info.spec.FetchSpec.FetchedColumns[i].Type
	}
	__antithesis_instrumentation__.Notify(466831)

	p.AddNoInputStage(corePlacement, info.post, typs, execinfrapb.Ordering{})

	p.PlanToStreamColMap = identityMap(make([]int, len(typs)), len(typs))
	p.SetMergeOrdering(dsp.convertOrdering(info.reqOrdering, p.PlanToStreamColMap))

	if parallelizeLocal {
		__antithesis_instrumentation__.Notify(466850)

		p.AddSingleGroupStage(dsp.gatewaySQLInstanceID, execinfrapb.ProcessorCoreUnion{Noop: &execinfrapb.NoopCoreSpec{}}, execinfrapb.PostProcessSpec{}, p.GetResultTypes())
	} else {
		__antithesis_instrumentation__.Notify(466851)
	}
	__antithesis_instrumentation__.Notify(466832)

	return nil
}

func (dsp *DistSQLPlanner) createPlanForRender(
	p *PhysicalPlan, n *renderNode, planCtx *PlanningCtx,
) error {
	__antithesis_instrumentation__.Notify(466852)
	typs, err := getTypesForPlanResult(n, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(466856)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466857)
	}
	__antithesis_instrumentation__.Notify(466853)
	if n.serialize {
		__antithesis_instrumentation__.Notify(466858)

		deferSerialization := len(p.MergeOrdering.Columns) == 0
		if deferSerialization {
			__antithesis_instrumentation__.Notify(466859)
			defer p.EnsureSingleStreamOnGateway()
		} else {
			__antithesis_instrumentation__.Notify(466860)
			p.EnsureSingleStreamOnGateway()
		}
	} else {
		__antithesis_instrumentation__.Notify(466861)
	}
	__antithesis_instrumentation__.Notify(466854)
	newColMap := identityMap(p.PlanToStreamColMap, len(n.render))
	newMergeOrdering := dsp.convertOrdering(n.reqOrdering, newColMap)
	err = p.AddRendering(
		n.render, planCtx, p.PlanToStreamColMap, typs, newMergeOrdering,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(466862)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466863)
	}
	__antithesis_instrumentation__.Notify(466855)
	p.PlanToStreamColMap = newColMap
	return nil
}

func (dsp *DistSQLPlanner) addSorters(
	p *PhysicalPlan, ordering colinfo.ColumnOrdering, alreadyOrderedPrefix int, limit int64,
) {
	__antithesis_instrumentation__.Notify(466864)

	outputOrdering := execinfrapb.ConvertToMappedSpecOrdering(ordering, p.PlanToStreamColMap)

	p.AddNoGroupingStage(
		execinfrapb.ProcessorCoreUnion{
			Sorter: &execinfrapb.SorterSpec{
				OutputOrdering:   outputOrdering,
				OrderingMatchLen: uint32(alreadyOrderedPrefix),
				Limit:            limit,
			},
		},
		execinfrapb.PostProcessSpec{},
		p.GetResultTypes(),
		outputOrdering,
	)

	if limit > 0 && func() bool {
		__antithesis_instrumentation__.Notify(466865)
		return len(p.ResultRouters) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(466866)
		post := execinfrapb.PostProcessSpec{
			Limit: uint64(limit),
		}
		p.AddSingleGroupStage(
			p.GatewaySQLInstanceID,
			execinfrapb.ProcessorCoreUnion{Noop: &execinfrapb.NoopCoreSpec{}},
			post,
			p.GetResultTypes(),
		)
	} else {
		__antithesis_instrumentation__.Notify(466867)
	}
}

type aggregatorPlanningInfo struct {
	aggregations             []execinfrapb.AggregatorSpec_Aggregation
	argumentsColumnTypes     [][]*types.T
	isScalar                 bool
	groupCols                []int
	groupColOrdering         colinfo.ColumnOrdering
	inputMergeOrdering       execinfrapb.Ordering
	reqOrdering              ReqOrdering
	allowPartialDistribution bool
}

func (dsp *DistSQLPlanner) addAggregators(
	planCtx *PlanningCtx, p *PhysicalPlan, n *groupNode,
) error {
	__antithesis_instrumentation__.Notify(466868)
	aggregations := make([]execinfrapb.AggregatorSpec_Aggregation, len(n.funcs))
	argumentsColumnTypes := make([][]*types.T, len(n.funcs))
	for i, fholder := range n.funcs {
		__antithesis_instrumentation__.Notify(466870)
		funcIdx, err := execinfrapb.GetAggregateFuncIdx(fholder.funcName)
		if err != nil {
			__antithesis_instrumentation__.Notify(466874)
			return err
		} else {
			__antithesis_instrumentation__.Notify(466875)
		}
		__antithesis_instrumentation__.Notify(466871)
		aggregations[i].Func = execinfrapb.AggregatorSpec_Func(funcIdx)
		aggregations[i].Distinct = fholder.isDistinct
		for _, renderIdx := range fholder.argRenderIdxs {
			__antithesis_instrumentation__.Notify(466876)
			aggregations[i].ColIdx = append(aggregations[i].ColIdx, uint32(p.PlanToStreamColMap[renderIdx]))
		}
		__antithesis_instrumentation__.Notify(466872)
		if fholder.hasFilter() {
			__antithesis_instrumentation__.Notify(466877)
			col := uint32(p.PlanToStreamColMap[fholder.filterRenderIdx])
			aggregations[i].FilterColIdx = &col
		} else {
			__antithesis_instrumentation__.Notify(466878)
		}
		__antithesis_instrumentation__.Notify(466873)
		aggregations[i].Arguments = make([]execinfrapb.Expression, len(fholder.arguments))
		argumentsColumnTypes[i] = make([]*types.T, len(fholder.arguments))
		for j, argument := range fholder.arguments {
			__antithesis_instrumentation__.Notify(466879)
			var err error
			aggregations[i].Arguments[j], err = physicalplan.MakeExpression(argument, planCtx, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(466881)
				return err
			} else {
				__antithesis_instrumentation__.Notify(466882)
			}
			__antithesis_instrumentation__.Notify(466880)
			argumentsColumnTypes[i][j] = argument.ResolvedType()
		}
	}
	__antithesis_instrumentation__.Notify(466869)

	return dsp.planAggregators(planCtx, p, &aggregatorPlanningInfo{
		aggregations:         aggregations,
		argumentsColumnTypes: argumentsColumnTypes,
		isScalar:             n.isScalar,
		groupCols:            n.groupCols,
		groupColOrdering:     n.groupColOrdering,
		inputMergeOrdering:   dsp.convertOrdering(planReqOrdering(n.plan), p.PlanToStreamColMap),
		reqOrdering:          n.reqOrdering,
	})
}

func (dsp *DistSQLPlanner) planAggregators(
	planCtx *PlanningCtx, p *PhysicalPlan, info *aggregatorPlanningInfo,
) error {
	__antithesis_instrumentation__.Notify(466883)
	aggType := execinfrapb.AggregatorSpec_NON_SCALAR
	if info.isScalar {
		__antithesis_instrumentation__.Notify(466894)
		aggType = execinfrapb.AggregatorSpec_SCALAR
	} else {
		__antithesis_instrumentation__.Notify(466895)
	}
	__antithesis_instrumentation__.Notify(466884)

	inputTypes := p.GetResultTypes()

	groupCols := make([]uint32, len(info.groupCols))
	for i, idx := range info.groupCols {
		__antithesis_instrumentation__.Notify(466896)
		groupCols[i] = uint32(p.PlanToStreamColMap[idx])
	}
	__antithesis_instrumentation__.Notify(466885)
	orderedGroupCols := make([]uint32, len(info.groupColOrdering))
	var orderedGroupColSet util.FastIntSet
	for i, c := range info.groupColOrdering {
		__antithesis_instrumentation__.Notify(466897)
		orderedGroupCols[i] = uint32(p.PlanToStreamColMap[c.ColIdx])
		orderedGroupColSet.Add(c.ColIdx)
	}
	__antithesis_instrumentation__.Notify(466886)

	allDistinct := true
	for _, e := range info.aggregations {
		__antithesis_instrumentation__.Notify(466898)
		if !e.Distinct {
			__antithesis_instrumentation__.Notify(466899)
			allDistinct = false
			break
		} else {
			__antithesis_instrumentation__.Notify(466900)
		}
	}
	__antithesis_instrumentation__.Notify(466887)
	if allDistinct {
		__antithesis_instrumentation__.Notify(466901)
		var distinctColumnsSet util.FastIntSet
		for _, e := range info.aggregations {
			__antithesis_instrumentation__.Notify(466903)
			for _, colIdx := range e.ColIdx {
				__antithesis_instrumentation__.Notify(466904)
				distinctColumnsSet.Add(int(colIdx))
			}
		}
		__antithesis_instrumentation__.Notify(466902)
		if distinctColumnsSet.Len() > 0 {
			__antithesis_instrumentation__.Notify(466905)

			distinctColumns := make([]uint32, 0, distinctColumnsSet.Len())
			distinctColumnsSet.ForEach(func(i int) {
				__antithesis_instrumentation__.Notify(466910)
				distinctColumns = append(distinctColumns, uint32(i))
			})
			__antithesis_instrumentation__.Notify(466906)
			ordering := info.inputMergeOrdering.Columns
			orderedColumns := make([]uint32, 0, len(ordering))
			for _, ord := range ordering {
				__antithesis_instrumentation__.Notify(466911)
				if distinctColumnsSet.Contains(int(ord.ColIdx)) {
					__antithesis_instrumentation__.Notify(466912)

					orderedColumns = append(orderedColumns, ord.ColIdx)
				} else {
					__antithesis_instrumentation__.Notify(466913)
				}
			}
			__antithesis_instrumentation__.Notify(466907)
			sort.Slice(orderedColumns, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(466914)
				return orderedColumns[i] < orderedColumns[j]
			})
			__antithesis_instrumentation__.Notify(466908)
			sort.Slice(distinctColumns, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(466915)
				return distinctColumns[i] < distinctColumns[j]
			})
			__antithesis_instrumentation__.Notify(466909)
			distinctSpec := execinfrapb.ProcessorCoreUnion{
				Distinct: dsp.createDistinctSpec(
					distinctColumns,
					orderedColumns,
					false,
					"",
					p.MergeOrdering,
				),
			}

			p.AddNoGroupingStage(distinctSpec, execinfrapb.PostProcessSpec{}, inputTypes, p.MergeOrdering)
		} else {
			__antithesis_instrumentation__.Notify(466916)
		}
	} else {
		__antithesis_instrumentation__.Notify(466917)
	}
	__antithesis_instrumentation__.Notify(466888)

	prevStageNode := p.Processors[p.ResultRouters[0]].SQLInstanceID
	for i := 1; i < len(p.ResultRouters); i++ {
		__antithesis_instrumentation__.Notify(466918)
		if n := p.Processors[p.ResultRouters[i]].SQLInstanceID; n != prevStageNode {
			__antithesis_instrumentation__.Notify(466919)
			prevStageNode = 0
			break
		} else {
			__antithesis_instrumentation__.Notify(466920)
		}
	}
	__antithesis_instrumentation__.Notify(466889)

	multiStage := prevStageNode == 0
	if multiStage {
		__antithesis_instrumentation__.Notify(466921)
		for _, e := range info.aggregations {
			__antithesis_instrumentation__.Notify(466922)
			if e.Distinct {
				__antithesis_instrumentation__.Notify(466924)
				multiStage = false
				break
			} else {
				__antithesis_instrumentation__.Notify(466925)
			}
			__antithesis_instrumentation__.Notify(466923)

			if _, ok := physicalplan.DistAggregationTable[e.Func]; !ok {
				__antithesis_instrumentation__.Notify(466926)
				multiStage = false
				break
			} else {
				__antithesis_instrumentation__.Notify(466927)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(466928)
	}
	__antithesis_instrumentation__.Notify(466890)

	var finalAggsSpec execinfrapb.AggregatorSpec
	var finalAggsPost execinfrapb.PostProcessSpec

	finalOutputOrdering := dsp.convertOrdering(info.reqOrdering, nil)

	if !multiStage {
		__antithesis_instrumentation__.Notify(466929)
		finalAggsSpec = execinfrapb.AggregatorSpec{
			Type:             aggType,
			Aggregations:     info.aggregations,
			GroupCols:        groupCols,
			OrderedGroupCols: orderedGroupCols,
			OutputOrdering:   finalOutputOrdering,
		}
	} else {
		__antithesis_instrumentation__.Notify(466930)

		nLocalAgg := 0
		nFinalAgg := 0
		needRender := false
		for _, e := range info.aggregations {
			__antithesis_instrumentation__.Notify(466936)
			info := physicalplan.DistAggregationTable[e.Func]
			nLocalAgg += len(info.LocalStage)
			nFinalAgg += len(info.FinalStage)
			if info.FinalRendering != nil {
				__antithesis_instrumentation__.Notify(466937)
				needRender = true
			} else {
				__antithesis_instrumentation__.Notify(466938)
			}
		}
		__antithesis_instrumentation__.Notify(466931)

		localAggs := make([]execinfrapb.AggregatorSpec_Aggregation, 0, nLocalAgg+len(groupCols))
		intermediateTypes := make([]*types.T, 0, nLocalAgg+len(groupCols))
		finalAggs := make([]execinfrapb.AggregatorSpec_Aggregation, 0, nFinalAgg)

		finalIdxMap := make([]uint32, nFinalAgg)

		var finalPreRenderTypes []*types.T
		if needRender {
			__antithesis_instrumentation__.Notify(466939)
			finalPreRenderTypes = make([]*types.T, 0, nFinalAgg)
		} else {
			__antithesis_instrumentation__.Notify(466940)
		}
		__antithesis_instrumentation__.Notify(466932)

		finalIdx := 0
		for _, e := range info.aggregations {
			__antithesis_instrumentation__.Notify(466941)
			info := physicalplan.DistAggregationTable[e.Func]

			relToAbsLocalIdx := make([]uint32, len(info.LocalStage))

			for i, localFunc := range info.LocalStage {
				__antithesis_instrumentation__.Notify(466943)
				localAgg := execinfrapb.AggregatorSpec_Aggregation{
					Func:         localFunc,
					ColIdx:       e.ColIdx,
					FilterColIdx: e.FilterColIdx,
				}

				isNewAgg := true
				for j, prevLocalAgg := range localAggs {
					__antithesis_instrumentation__.Notify(466945)
					if localAgg.Equals(prevLocalAgg) {
						__antithesis_instrumentation__.Notify(466946)

						relToAbsLocalIdx[i] = uint32(j)
						isNewAgg = false
						break
					} else {
						__antithesis_instrumentation__.Notify(466947)
					}
				}
				__antithesis_instrumentation__.Notify(466944)

				if isNewAgg {
					__antithesis_instrumentation__.Notify(466948)

					relToAbsLocalIdx[i] = uint32(len(localAggs))
					localAggs = append(localAggs, localAgg)

					argTypes := make([]*types.T, len(e.ColIdx))
					for j, c := range e.ColIdx {
						__antithesis_instrumentation__.Notify(466951)
						argTypes[j] = inputTypes[c]
					}
					__antithesis_instrumentation__.Notify(466949)
					_, outputType, err := execinfra.GetAggregateInfo(localFunc, argTypes...)
					if err != nil {
						__antithesis_instrumentation__.Notify(466952)
						return err
					} else {
						__antithesis_instrumentation__.Notify(466953)
					}
					__antithesis_instrumentation__.Notify(466950)
					intermediateTypes = append(intermediateTypes, outputType)
				} else {
					__antithesis_instrumentation__.Notify(466954)
				}
			}
			__antithesis_instrumentation__.Notify(466942)

			for _, finalInfo := range info.FinalStage {
				__antithesis_instrumentation__.Notify(466955)

				argIdxs := make([]uint32, len(finalInfo.LocalIdxs))
				for i, relIdx := range finalInfo.LocalIdxs {
					__antithesis_instrumentation__.Notify(466959)
					argIdxs[i] = relToAbsLocalIdx[relIdx]
				}
				__antithesis_instrumentation__.Notify(466956)
				finalAgg := execinfrapb.AggregatorSpec_Aggregation{
					Func:   finalInfo.Fn,
					ColIdx: argIdxs,
				}

				isNewAgg := true
				for i, prevFinalAgg := range finalAggs {
					__antithesis_instrumentation__.Notify(466960)
					if finalAgg.Equals(prevFinalAgg) {
						__antithesis_instrumentation__.Notify(466961)

						finalIdxMap[finalIdx] = uint32(i)
						isNewAgg = false
						break
					} else {
						__antithesis_instrumentation__.Notify(466962)
					}
				}
				__antithesis_instrumentation__.Notify(466957)

				if isNewAgg {
					__antithesis_instrumentation__.Notify(466963)
					finalIdxMap[finalIdx] = uint32(len(finalAggs))
					finalAggs = append(finalAggs, finalAgg)

					if needRender {
						__antithesis_instrumentation__.Notify(466964)
						argTypes := make([]*types.T, len(finalInfo.LocalIdxs))
						for i := range finalInfo.LocalIdxs {
							__antithesis_instrumentation__.Notify(466967)

							argTypes[i] = intermediateTypes[argIdxs[i]]
						}
						__antithesis_instrumentation__.Notify(466965)
						_, outputType, err := execinfra.GetAggregateInfo(finalInfo.Fn, argTypes...)
						if err != nil {
							__antithesis_instrumentation__.Notify(466968)
							return err
						} else {
							__antithesis_instrumentation__.Notify(466969)
						}
						__antithesis_instrumentation__.Notify(466966)
						finalPreRenderTypes = append(finalPreRenderTypes, outputType)
					} else {
						__antithesis_instrumentation__.Notify(466970)
					}
				} else {
					__antithesis_instrumentation__.Notify(466971)
				}
				__antithesis_instrumentation__.Notify(466958)
				finalIdx++
			}
		}
		__antithesis_instrumentation__.Notify(466933)

		finalGroupCols := make([]uint32, len(groupCols))
		finalOrderedGroupCols := make([]uint32, 0, len(orderedGroupCols))
		for i, groupColIdx := range groupCols {
			__antithesis_instrumentation__.Notify(466972)
			agg := execinfrapb.AggregatorSpec_Aggregation{
				Func:   execinfrapb.AnyNotNull,
				ColIdx: []uint32{groupColIdx},
			}

			idx := -1
			for j := range localAggs {
				__antithesis_instrumentation__.Notify(466975)
				if localAggs[j].Equals(agg) {
					__antithesis_instrumentation__.Notify(466976)
					idx = j
					break
				} else {
					__antithesis_instrumentation__.Notify(466977)
				}
			}
			__antithesis_instrumentation__.Notify(466973)
			if idx == -1 {
				__antithesis_instrumentation__.Notify(466978)

				idx = len(localAggs)
				localAggs = append(localAggs, agg)
				intermediateTypes = append(intermediateTypes, inputTypes[groupColIdx])
			} else {
				__antithesis_instrumentation__.Notify(466979)
			}
			__antithesis_instrumentation__.Notify(466974)
			finalGroupCols[i] = uint32(idx)
			if orderedGroupColSet.Contains(info.groupCols[i]) {
				__antithesis_instrumentation__.Notify(466980)
				finalOrderedGroupCols = append(finalOrderedGroupCols, uint32(idx))
			} else {
				__antithesis_instrumentation__.Notify(466981)
			}
		}
		__antithesis_instrumentation__.Notify(466934)

		ordCols := make([]execinfrapb.Ordering_Column, len(info.groupColOrdering))
		for i, o := range info.groupColOrdering {
			__antithesis_instrumentation__.Notify(466982)

			found := false
			for j, col := range info.groupCols {
				__antithesis_instrumentation__.Notify(466985)
				if col == o.ColIdx {
					__antithesis_instrumentation__.Notify(466986)
					ordCols[i].ColIdx = finalGroupCols[j]
					found = true
					break
				} else {
					__antithesis_instrumentation__.Notify(466987)
				}
			}
			__antithesis_instrumentation__.Notify(466983)
			if !found {
				__antithesis_instrumentation__.Notify(466988)
				return errors.AssertionFailedf("group column ordering contains non-grouping column %d", o.ColIdx)
			} else {
				__antithesis_instrumentation__.Notify(466989)
			}
			__antithesis_instrumentation__.Notify(466984)
			if o.Direction == encoding.Descending {
				__antithesis_instrumentation__.Notify(466990)
				ordCols[i].Direction = execinfrapb.Ordering_Column_DESC
			} else {
				__antithesis_instrumentation__.Notify(466991)
				ordCols[i].Direction = execinfrapb.Ordering_Column_ASC
			}
		}
		__antithesis_instrumentation__.Notify(466935)

		localAggsSpec := execinfrapb.AggregatorSpec{
			Type:             aggType,
			Aggregations:     localAggs,
			GroupCols:        groupCols,
			OrderedGroupCols: orderedGroupCols,
			OutputOrdering:   execinfrapb.Ordering{Columns: ordCols},
		}

		p.AddNoGroupingStage(
			execinfrapb.ProcessorCoreUnion{Aggregator: &localAggsSpec},
			execinfrapb.PostProcessSpec{},
			intermediateTypes,
			execinfrapb.Ordering{Columns: ordCols},
		)

		finalAggsSpec = execinfrapb.AggregatorSpec{
			Type:             aggType,
			Aggregations:     finalAggs,
			GroupCols:        finalGroupCols,
			OrderedGroupCols: finalOrderedGroupCols,
			OutputOrdering:   finalOutputOrdering,
		}

		if needRender {
			__antithesis_instrumentation__.Notify(466992)

			renderExprs := make([]execinfrapb.Expression, len(info.aggregations))
			h := tree.MakeTypesOnlyIndexedVarHelper(finalPreRenderTypes)

			finalIdx := 0
			for i, e := range info.aggregations {
				__antithesis_instrumentation__.Notify(466994)
				info := physicalplan.DistAggregationTable[e.Func]
				if info.FinalRendering == nil {
					__antithesis_instrumentation__.Notify(466996)

					mappedIdx := int(finalIdxMap[finalIdx])
					var err error
					renderExprs[i], err = physicalplan.MakeExpression(
						h.IndexedVar(mappedIdx), planCtx, nil)
					if err != nil {
						__antithesis_instrumentation__.Notify(466997)
						return err
					} else {
						__antithesis_instrumentation__.Notify(466998)
					}
				} else {
					__antithesis_instrumentation__.Notify(466999)

					mappedIdxs := make([]int, len(info.FinalStage))
					for j := range info.FinalStage {
						__antithesis_instrumentation__.Notify(467002)
						mappedIdxs[j] = int(finalIdxMap[finalIdx+j])
					}
					__antithesis_instrumentation__.Notify(467000)

					expr, err := info.FinalRendering(&h, mappedIdxs)
					if err != nil {
						__antithesis_instrumentation__.Notify(467003)
						return err
					} else {
						__antithesis_instrumentation__.Notify(467004)
					}
					__antithesis_instrumentation__.Notify(467001)
					renderExprs[i], err = physicalplan.MakeExpression(
						expr, planCtx,
						nil)
					if err != nil {
						__antithesis_instrumentation__.Notify(467005)
						return err
					} else {
						__antithesis_instrumentation__.Notify(467006)
					}
				}
				__antithesis_instrumentation__.Notify(466995)
				finalIdx += len(info.FinalStage)
			}
			__antithesis_instrumentation__.Notify(466993)
			finalAggsPost.RenderExprs = renderExprs
		} else {
			__antithesis_instrumentation__.Notify(467007)
			if len(finalAggs) < len(info.aggregations) {
				__antithesis_instrumentation__.Notify(467008)

				finalAggsPost.Projection = true
				finalAggsPost.OutputColumns = finalIdxMap
			} else {
				__antithesis_instrumentation__.Notify(467009)
			}
		}
	}
	__antithesis_instrumentation__.Notify(466891)

	finalOutTypes := make([]*types.T, len(info.aggregations))
	for i, agg := range info.aggregations {
		__antithesis_instrumentation__.Notify(467010)
		argTypes := make([]*types.T, len(agg.ColIdx)+len(agg.Arguments))
		for j, c := range agg.ColIdx {
			__antithesis_instrumentation__.Notify(467013)
			argTypes[j] = inputTypes[c]
		}
		__antithesis_instrumentation__.Notify(467011)
		copy(argTypes[len(agg.ColIdx):], info.argumentsColumnTypes[i])
		var err error
		_, returnTyp, err := execinfra.GetAggregateInfo(agg.Func, argTypes...)
		if err != nil {
			__antithesis_instrumentation__.Notify(467014)
			return err
		} else {
			__antithesis_instrumentation__.Notify(467015)
		}
		__antithesis_instrumentation__.Notify(467012)
		finalOutTypes[i] = returnTyp
	}
	__antithesis_instrumentation__.Notify(466892)

	p.PlanToStreamColMap = identityMap(p.PlanToStreamColMap, len(info.aggregations))

	if len(finalAggsSpec.GroupCols) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(467016)
		return len(p.ResultRouters) == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(467017)

		node := dsp.gatewaySQLInstanceID
		if prevStageNode != 0 {
			__antithesis_instrumentation__.Notify(467019)
			node = prevStageNode
		} else {
			__antithesis_instrumentation__.Notify(467020)
		}
		__antithesis_instrumentation__.Notify(467018)
		p.AddSingleGroupStage(
			node,
			execinfrapb.ProcessorCoreUnion{Aggregator: &finalAggsSpec},
			finalAggsPost,
			finalOutTypes,
		)
	} else {
		__antithesis_instrumentation__.Notify(467021)

		for _, resultProc := range p.ResultRouters {
			__antithesis_instrumentation__.Notify(467026)
			p.Processors[resultProc].Spec.Output[0] = execinfrapb.OutputRouterSpec{
				Type:        execinfrapb.OutputRouterSpec_BY_HASH,
				HashColumns: finalAggsSpec.GroupCols,
			}
		}
		__antithesis_instrumentation__.Notify(467022)

		stageID := p.NewStage(true, info.allowPartialDistribution)

		pIdxStart := physicalplan.ProcessorIdx(len(p.Processors))
		prevStageResultTypes := p.GetResultTypes()
		for _, resultProc := range p.ResultRouters {
			__antithesis_instrumentation__.Notify(467027)
			proc := physicalplan.Processor{
				SQLInstanceID: p.Processors[resultProc].SQLInstanceID,
				Spec: execinfrapb.ProcessorSpec{
					Input: []execinfrapb.InputSyncSpec{{

						ColumnTypes: prevStageResultTypes,
					}},
					Core: execinfrapb.ProcessorCoreUnion{Aggregator: &finalAggsSpec},
					Post: finalAggsPost,
					Output: []execinfrapb.OutputRouterSpec{{
						Type: execinfrapb.OutputRouterSpec_PASS_THROUGH,
					}},
					StageID:     stageID,
					ResultTypes: finalOutTypes,
				},
			}
			p.AddProcessor(proc)
		}
		__antithesis_instrumentation__.Notify(467023)

		for bucket := 0; bucket < len(p.ResultRouters); bucket++ {
			__antithesis_instrumentation__.Notify(467028)
			pIdx := pIdxStart + physicalplan.ProcessorIdx(bucket)
			p.MergeResultStreams(p.ResultRouters, bucket, p.MergeOrdering, pIdx, 0, false)
		}
		__antithesis_instrumentation__.Notify(467024)

		for i := 0; i < len(p.ResultRouters); i++ {
			__antithesis_instrumentation__.Notify(467029)
			p.ResultRouters[i] = pIdxStart + physicalplan.ProcessorIdx(i)
		}
		__antithesis_instrumentation__.Notify(467025)

		p.SetMergeOrdering(dsp.convertOrdering(info.reqOrdering, p.PlanToStreamColMap))
	}
	__antithesis_instrumentation__.Notify(466893)

	return nil
}

func (dsp *DistSQLPlanner) createPlanForIndexJoin(
	ctx context.Context, planCtx *PlanningCtx, n *indexJoinNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467030)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.input)
	if err != nil {
		__antithesis_instrumentation__.Notify(467037)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467038)
	}
	__antithesis_instrumentation__.Notify(467031)

	pkCols := make([]uint32, len(n.keyCols))
	for i := range n.keyCols {
		__antithesis_instrumentation__.Notify(467039)
		streamColOrd := plan.PlanToStreamColMap[n.keyCols[i]]
		if streamColOrd == -1 {
			__antithesis_instrumentation__.Notify(467041)
			panic("key column not in planToStreamColMap")
		} else {
			__antithesis_instrumentation__.Notify(467042)
		}
		__antithesis_instrumentation__.Notify(467040)
		pkCols[i] = uint32(streamColOrd)
	}
	__antithesis_instrumentation__.Notify(467032)

	plan.AddProjection(pkCols, execinfrapb.Ordering{})

	joinReaderSpec := execinfrapb.JoinReaderSpec{
		Type:              descpb.InnerJoin,
		LockingStrength:   n.table.lockingStrength,
		LockingWaitPolicy: n.table.lockingWaitPolicy,
		MaintainOrdering:  len(n.reqOrdering) > 0,
		LimitHint:         n.limitHint,
	}

	fetchColIDs := make([]descpb.ColumnID, len(n.cols))
	var fetchOrdinals util.FastIntSet
	for i := range n.cols {
		__antithesis_instrumentation__.Notify(467043)
		fetchColIDs[i] = n.cols[i].GetID()
		fetchOrdinals.Add(n.cols[i].Ordinal())
	}
	__antithesis_instrumentation__.Notify(467033)
	index := n.table.desc.GetPrimaryIndex()
	if err := rowenc.InitIndexFetchSpec(
		&joinReaderSpec.FetchSpec,
		planCtx.ExtendedEvalCtx.Codec,
		n.table.desc,
		index,
		fetchColIDs,
	); err != nil {
		__antithesis_instrumentation__.Notify(467044)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467045)
	}
	__antithesis_instrumentation__.Notify(467034)

	splitter := span.MakeSplitter(n.table.desc, index, fetchOrdinals)
	joinReaderSpec.SplitFamilyIDs = splitter.FamilyIDs()

	plan.PlanToStreamColMap = identityMap(plan.PlanToStreamColMap, len(fetchColIDs))

	types, err := getTypesForPlanResult(n, plan.PlanToStreamColMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(467046)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467047)
	}
	__antithesis_instrumentation__.Notify(467035)
	if len(plan.ResultRouters) > 1 {
		__antithesis_instrumentation__.Notify(467048)

		plan.AddNoGroupingStage(
			execinfrapb.ProcessorCoreUnion{JoinReader: &joinReaderSpec},
			execinfrapb.PostProcessSpec{},
			types,
			dsp.convertOrdering(n.reqOrdering, plan.PlanToStreamColMap),
		)
	} else {
		__antithesis_instrumentation__.Notify(467049)

		plan.AddSingleGroupStage(
			plan.Processors[plan.ResultRouters[0]].SQLInstanceID,
			execinfrapb.ProcessorCoreUnion{JoinReader: &joinReaderSpec},
			execinfrapb.PostProcessSpec{},
			types,
		)
	}
	__antithesis_instrumentation__.Notify(467036)
	return plan, nil
}

func (dsp *DistSQLPlanner) createPlanForLookupJoin(
	ctx context.Context, planCtx *PlanningCtx, n *lookupJoinNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467050)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.input)
	if err != nil {
		__antithesis_instrumentation__.Notify(467060)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467061)
	}
	__antithesis_instrumentation__.Notify(467051)

	joinReaderSpec := execinfrapb.JoinReaderSpec{
		Type:              n.joinType,
		LockingStrength:   n.table.lockingStrength,
		LockingWaitPolicy: n.table.lockingWaitPolicy,

		MaintainOrdering: len(n.reqOrdering) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(467062)
			return n.isFirstJoinInPairedJoiner == true
		}() == true,
		LeftJoinWithPairedJoiner:          n.isSecondJoinInPairedJoiner,
		OutputGroupContinuationForLeftRow: n.isFirstJoinInPairedJoiner,
		LookupBatchBytesLimit:             dsp.distSQLSrv.TestingKnobs.JoinReaderBatchBytesLimit,
		LimitHint:                         n.limitHint,
	}

	fetchColIDs := make([]descpb.ColumnID, len(n.table.cols))
	var fetchOrdinals util.FastIntSet
	for i := range n.table.cols {
		__antithesis_instrumentation__.Notify(467063)
		fetchColIDs[i] = n.table.cols[i].GetID()
		fetchOrdinals.Add(n.table.cols[i].Ordinal())
	}
	__antithesis_instrumentation__.Notify(467052)
	if err := rowenc.InitIndexFetchSpec(
		&joinReaderSpec.FetchSpec,
		planCtx.ExtendedEvalCtx.Codec,
		n.table.desc,
		n.table.index,
		fetchColIDs,
	); err != nil {
		__antithesis_instrumentation__.Notify(467064)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467065)
	}
	__antithesis_instrumentation__.Notify(467053)

	splitter := span.MakeSplitter(n.table.desc, n.table.index, fetchOrdinals)
	joinReaderSpec.SplitFamilyIDs = splitter.FamilyIDs()

	joinReaderSpec.LookupColumns = make([]uint32, len(n.eqCols))
	for i, col := range n.eqCols {
		__antithesis_instrumentation__.Notify(467066)
		if plan.PlanToStreamColMap[col] == -1 {
			__antithesis_instrumentation__.Notify(467068)
			panic("lookup column not in planToStreamColMap")
		} else {
			__antithesis_instrumentation__.Notify(467069)
		}
		__antithesis_instrumentation__.Notify(467067)
		joinReaderSpec.LookupColumns[i] = uint32(plan.PlanToStreamColMap[col])
	}
	__antithesis_instrumentation__.Notify(467054)
	joinReaderSpec.LookupColumnsAreKey = n.eqColsAreKey

	inputTypes := plan.GetResultTypes()
	fetchedColumns := joinReaderSpec.FetchSpec.FetchedColumns
	numOutCols := len(inputTypes) + len(fetchedColumns)
	if n.isFirstJoinInPairedJoiner {
		__antithesis_instrumentation__.Notify(467070)

		numOutCols++
	} else {
		__antithesis_instrumentation__.Notify(467071)
	}
	__antithesis_instrumentation__.Notify(467055)

	var outTypes []*types.T
	var planToStreamColMap []int
	if !n.joinType.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(467072)
		if n.isFirstJoinInPairedJoiner {
			__antithesis_instrumentation__.Notify(467074)
			return nil, errors.AssertionFailedf("continuation column without right columns")
		} else {
			__antithesis_instrumentation__.Notify(467075)
		}
		__antithesis_instrumentation__.Notify(467073)
		outTypes = inputTypes
		planToStreamColMap = plan.PlanToStreamColMap
	} else {
		__antithesis_instrumentation__.Notify(467076)
		outTypes = make([]*types.T, numOutCols)
		copy(outTypes, inputTypes)
		planToStreamColMap = plan.PlanToStreamColMap
		for i := range fetchedColumns {
			__antithesis_instrumentation__.Notify(467078)
			outTypes[len(inputTypes)+i] = fetchedColumns[i].Type
			planToStreamColMap = append(planToStreamColMap, len(inputTypes)+i)
		}
		__antithesis_instrumentation__.Notify(467077)
		if n.isFirstJoinInPairedJoiner {
			__antithesis_instrumentation__.Notify(467079)
			outTypes[numOutCols-1] = types.Bool
			planToStreamColMap = append(planToStreamColMap, numOutCols-1)
		} else {
			__antithesis_instrumentation__.Notify(467080)
		}
	}
	__antithesis_instrumentation__.Notify(467056)

	if n.lookupExpr != nil {
		__antithesis_instrumentation__.Notify(467081)
		var err error
		joinReaderSpec.LookupExpr, err = physicalplan.MakeExpression(
			n.lookupExpr, planCtx, nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(467082)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467083)
		}
	} else {
		__antithesis_instrumentation__.Notify(467084)
	}
	__antithesis_instrumentation__.Notify(467057)
	if n.remoteLookupExpr != nil {
		__antithesis_instrumentation__.Notify(467085)
		if n.lookupExpr == nil {
			__antithesis_instrumentation__.Notify(467087)
			return nil, errors.AssertionFailedf("remoteLookupExpr is set but lookupExpr is not")
		} else {
			__antithesis_instrumentation__.Notify(467088)
		}
		__antithesis_instrumentation__.Notify(467086)
		var err error
		joinReaderSpec.RemoteLookupExpr, err = physicalplan.MakeExpression(
			n.remoteLookupExpr, planCtx, nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(467089)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467090)
		}
	} else {
		__antithesis_instrumentation__.Notify(467091)
	}
	__antithesis_instrumentation__.Notify(467058)

	if n.onCond != nil {
		__antithesis_instrumentation__.Notify(467092)
		var err error
		joinReaderSpec.OnExpr, err = physicalplan.MakeExpression(
			n.onCond, planCtx, nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(467093)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467094)
		}
	} else {
		__antithesis_instrumentation__.Notify(467095)
	}
	__antithesis_instrumentation__.Notify(467059)

	plan.AddNoGroupingStage(
		execinfrapb.ProcessorCoreUnion{JoinReader: &joinReaderSpec},
		execinfrapb.PostProcessSpec{},
		outTypes,
		dsp.convertOrdering(planReqOrdering(n), planToStreamColMap),
	)
	plan.PlanToStreamColMap = planToStreamColMap
	return plan, nil
}

func mappingHelperForLookupJoins(
	plan *PhysicalPlan, input planNode, table *scanNode, addContinuationCol bool,
) (
	numInputNodeCols int,
	planToStreamColMap []int,
	post execinfrapb.PostProcessSpec,
	outTypes []*types.T,
) {
	__antithesis_instrumentation__.Notify(467096)

	inputTypes := plan.GetResultTypes()
	numLeftCols := len(inputTypes)
	numOutCols := numLeftCols + len(table.cols)
	if addContinuationCol {
		__antithesis_instrumentation__.Notify(467104)
		numOutCols++
	} else {
		__antithesis_instrumentation__.Notify(467105)
	}
	__antithesis_instrumentation__.Notify(467097)
	post = execinfrapb.PostProcessSpec{Projection: true}

	post.OutputColumns = make([]uint32, numOutCols)
	outTypes = make([]*types.T, numOutCols)

	for i := 0; i < numLeftCols; i++ {
		__antithesis_instrumentation__.Notify(467106)
		outTypes[i] = inputTypes[i]
		post.OutputColumns[i] = uint32(i)
	}
	__antithesis_instrumentation__.Notify(467098)
	for i := range table.cols {
		__antithesis_instrumentation__.Notify(467107)
		outTypes[numLeftCols+i] = table.cols[i].GetType()
		ord := tableOrdinal(table.desc, table.cols[i].GetID())
		post.OutputColumns[numLeftCols+i] = uint32(numLeftCols + ord)
	}
	__antithesis_instrumentation__.Notify(467099)
	if addContinuationCol {
		__antithesis_instrumentation__.Notify(467108)
		outTypes[numOutCols-1] = types.Bool
		post.OutputColumns[numOutCols-1] = uint32(numLeftCols + len(table.desc.DeletableColumns()))
	} else {
		__antithesis_instrumentation__.Notify(467109)
	}
	__antithesis_instrumentation__.Notify(467100)

	numInputNodeCols = len(planColumns(input))
	lenPlanToStreamColMap := numInputNodeCols + len(table.cols)
	if addContinuationCol {
		__antithesis_instrumentation__.Notify(467110)
		lenPlanToStreamColMap++
	} else {
		__antithesis_instrumentation__.Notify(467111)
	}
	__antithesis_instrumentation__.Notify(467101)
	planToStreamColMap = makePlanToStreamColMap(lenPlanToStreamColMap)
	copy(planToStreamColMap, plan.PlanToStreamColMap)
	for i := range table.cols {
		__antithesis_instrumentation__.Notify(467112)
		planToStreamColMap[numInputNodeCols+i] = numLeftCols + i
	}
	__antithesis_instrumentation__.Notify(467102)
	if addContinuationCol {
		__antithesis_instrumentation__.Notify(467113)
		planToStreamColMap[lenPlanToStreamColMap-1] = numLeftCols + len(table.cols)
	} else {
		__antithesis_instrumentation__.Notify(467114)
	}
	__antithesis_instrumentation__.Notify(467103)
	return numInputNodeCols, planToStreamColMap, post, outTypes
}

func makeIndexVarMapForLookupJoins(
	numInputNodeCols int, table *scanNode, plan *PhysicalPlan, post *execinfrapb.PostProcessSpec,
) (indexVarMap []int) {
	__antithesis_instrumentation__.Notify(467115)

	indexVarMap = makePlanToStreamColMap(numInputNodeCols + len(table.cols))
	copy(indexVarMap, plan.PlanToStreamColMap)
	numLeftCols := len(plan.GetResultTypes())
	for i := range table.cols {
		__antithesis_instrumentation__.Notify(467117)
		indexVarMap[numInputNodeCols+i] = int(post.OutputColumns[numLeftCols+i])
	}
	__antithesis_instrumentation__.Notify(467116)
	return indexVarMap
}

func truncateToInputForLookupJoins(
	numInputNodeCols int, planToStreamColMap []int, outputColumns []uint32, outTypes []*types.T,
) ([]int, []uint32, []*types.T) {
	__antithesis_instrumentation__.Notify(467118)
	planToStreamColMap = planToStreamColMap[:numInputNodeCols]
	outputColumns = outputColumns[:numInputNodeCols]
	outTypes = outTypes[:numInputNodeCols]
	return planToStreamColMap, outputColumns, outTypes
}

func (dsp *DistSQLPlanner) createPlanForInvertedJoin(
	ctx context.Context, planCtx *PlanningCtx, n *invertedJoinNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467119)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.input)
	if err != nil {
		__antithesis_instrumentation__.Notify(467126)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467127)
	}
	__antithesis_instrumentation__.Notify(467120)

	invertedJoinerSpec := execinfrapb.InvertedJoinerSpec{
		Table:                             *n.table.desc.TableDesc(),
		Type:                              n.joinType,
		MaintainOrdering:                  len(n.reqOrdering) > 0,
		OutputGroupContinuationForLeftRow: n.isFirstJoinInPairedJoiner,
	}
	invertedJoinerSpec.IndexIdx, err = getIndexIdx(n.table.index, n.table.desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(467128)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467129)
	}
	__antithesis_instrumentation__.Notify(467121)

	numInputNodeCols, planToStreamColMap, post, types :=
		mappingHelperForLookupJoins(plan, n.input, n.table, n.isFirstJoinInPairedJoiner)

	invertedJoinerSpec.PrefixEqualityColumns = make([]uint32, len(n.prefixEqCols))
	for i, col := range n.prefixEqCols {
		__antithesis_instrumentation__.Notify(467130)
		if plan.PlanToStreamColMap[col] == -1 {
			__antithesis_instrumentation__.Notify(467132)
			panic("lookup column not in planToStreamColMap")
		} else {
			__antithesis_instrumentation__.Notify(467133)
		}
		__antithesis_instrumentation__.Notify(467131)
		invertedJoinerSpec.PrefixEqualityColumns[i] = uint32(plan.PlanToStreamColMap[col])
	}
	__antithesis_instrumentation__.Notify(467122)

	indexVarMap := makeIndexVarMapForLookupJoins(numInputNodeCols, n.table, plan, &post)
	if invertedJoinerSpec.InvertedExpr, err = physicalplan.MakeExpression(
		n.invertedExpr, planCtx, indexVarMap,
	); err != nil {
		__antithesis_instrumentation__.Notify(467134)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467135)
	}
	__antithesis_instrumentation__.Notify(467123)

	if n.onExpr != nil {
		__antithesis_instrumentation__.Notify(467136)
		if invertedJoinerSpec.OnExpr, err = physicalplan.MakeExpression(
			n.onExpr, planCtx, indexVarMap,
		); err != nil {
			__antithesis_instrumentation__.Notify(467137)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467138)
		}
	} else {
		__antithesis_instrumentation__.Notify(467139)
	}
	__antithesis_instrumentation__.Notify(467124)

	if !n.joinType.ShouldIncludeRightColsInOutput() {
		__antithesis_instrumentation__.Notify(467140)
		planToStreamColMap, post.OutputColumns, types = truncateToInputForLookupJoins(
			numInputNodeCols, planToStreamColMap, post.OutputColumns, types)
	} else {
		__antithesis_instrumentation__.Notify(467141)
	}
	__antithesis_instrumentation__.Notify(467125)

	plan.AddNoGroupingStage(
		execinfrapb.ProcessorCoreUnion{InvertedJoiner: &invertedJoinerSpec},
		post,
		types,
		dsp.convertOrdering(planReqOrdering(n), planToStreamColMap),
	)
	plan.PlanToStreamColMap = planToStreamColMap
	return plan, nil
}

func (dsp *DistSQLPlanner) createPlanForZigzagJoin(
	planCtx *PlanningCtx, n *zigzagJoinNode,
) (plan *PhysicalPlan, err error) {
	__antithesis_instrumentation__.Notify(467142)

	sides := make([]zigzagPlanningSide, len(n.sides))

	for i, side := range n.sides {
		__antithesis_instrumentation__.Notify(467144)

		typs := getTypesFromResultColumns(side.fixedVals.columns)
		valuesSpec, err := dsp.createValuesSpecFromTuples(planCtx, side.fixedVals.tuples, typs)
		if err != nil {
			__antithesis_instrumentation__.Notify(467146)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467147)
		}
		__antithesis_instrumentation__.Notify(467145)

		sides[i] = zigzagPlanningSide{
			desc:        side.scan.desc,
			index:       side.scan.index,
			cols:        side.scan.cols,
			eqCols:      side.eqCols,
			fixedValues: valuesSpec,
		}
	}
	__antithesis_instrumentation__.Notify(467143)

	return dsp.planZigzagJoin(planCtx, zigzagPlanningInfo{
		sides:       sides,
		columns:     n.columns,
		onCond:      n.onCond,
		reqOrdering: n.reqOrdering,
	})
}

type zigzagPlanningSide struct {
	desc        catalog.TableDescriptor
	index       catalog.Index
	cols        []catalog.Column
	eqCols      []int
	fixedValues *execinfrapb.ValuesCoreSpec
}

type zigzagPlanningInfo struct {
	sides       []zigzagPlanningSide
	columns     colinfo.ResultColumns
	onCond      tree.TypedExpr
	reqOrdering ReqOrdering
}

func (dsp *DistSQLPlanner) planZigzagJoin(
	planCtx *PlanningCtx, pi zigzagPlanningInfo,
) (plan *PhysicalPlan, err error) {
	__antithesis_instrumentation__.Notify(467148)

	plan = planCtx.NewPhysicalPlan()
	tables := make([]descpb.TableDescriptor, len(pi.sides))
	indexOrdinals := make([]uint32, len(pi.sides))
	cols := make([]execinfrapb.Columns, len(pi.sides))
	fixedValues := make([]*execinfrapb.ValuesCoreSpec, len(pi.sides))

	for i, side := range pi.sides {
		__antithesis_instrumentation__.Notify(467152)
		tables[i] = *side.desc.TableDesc()
		indexOrdinals[i], err = getIndexIdx(side.index, side.desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(467155)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467156)
		}
		__antithesis_instrumentation__.Notify(467153)

		cols[i].Columns = make([]uint32, len(side.eqCols))
		for j, col := range side.eqCols {
			__antithesis_instrumentation__.Notify(467157)
			cols[i].Columns[j] = uint32(col)
		}
		__antithesis_instrumentation__.Notify(467154)
		fixedValues[i] = side.fixedValues
	}
	__antithesis_instrumentation__.Notify(467149)

	zigzagJoinerSpec := execinfrapb.ZigzagJoinerSpec{
		Tables:        tables,
		EqColumns:     cols,
		IndexOrdinals: indexOrdinals,
		FixedValues:   fixedValues,
		Type:          descpb.InnerJoin,
	}

	post := execinfrapb.PostProcessSpec{Projection: true}
	numOutCols := len(pi.columns)
	post.OutputColumns = make([]uint32, numOutCols)
	types := make([]*types.T, numOutCols)
	planToStreamColMap := makePlanToStreamColMap(numOutCols)
	colOffset := 0
	i := 0

	for _, side := range pi.sides {
		__antithesis_instrumentation__.Notify(467158)

		for _, col := range side.cols {
			__antithesis_instrumentation__.Notify(467160)
			ord := tableOrdinal(side.desc, col.GetID())
			post.OutputColumns[i] = uint32(colOffset + ord)
			types[i] = col.GetType()
			planToStreamColMap[i] = i
			i++
		}
		__antithesis_instrumentation__.Notify(467159)
		colOffset += len(side.desc.PublicColumns())
	}
	__antithesis_instrumentation__.Notify(467150)

	if pi.onCond != nil {
		__antithesis_instrumentation__.Notify(467161)

		indexVarMap := makePlanToStreamColMap(len(pi.columns))
		for i := 0; i < len(pi.columns); i++ {
			__antithesis_instrumentation__.Notify(467163)
			indexVarMap[i] = int(post.OutputColumns[i])
		}
		__antithesis_instrumentation__.Notify(467162)
		zigzagJoinerSpec.OnExpr, err = physicalplan.MakeExpression(
			pi.onCond, planCtx, indexVarMap,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(467164)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467165)
		}
	} else {
		__antithesis_instrumentation__.Notify(467166)
	}
	__antithesis_instrumentation__.Notify(467151)

	corePlacement := []physicalplan.ProcessorCorePlacement{{
		SQLInstanceID: dsp.gatewaySQLInstanceID,
		Core:          execinfrapb.ProcessorCoreUnion{ZigzagJoiner: &zigzagJoinerSpec},
	}}

	plan.AddNoInputStage(corePlacement, post, types, execinfrapb.Ordering{})
	plan.PlanToStreamColMap = planToStreamColMap

	return plan, nil
}

func (dsp *DistSQLPlanner) createPlanForInvertedFilter(
	ctx context.Context, planCtx *PlanningCtx, n *invertedFilterNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467167)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.input)
	if err != nil {
		__antithesis_instrumentation__.Notify(467173)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467174)
	}
	__antithesis_instrumentation__.Notify(467168)
	invertedFiltererSpec := &execinfrapb.InvertedFiltererSpec{
		InvertedColIdx: uint32(n.invColumn),
		InvertedExpr:   *n.expression.ToProto(),
	}
	if n.preFiltererExpr != nil {
		__antithesis_instrumentation__.Notify(467175)
		invertedFiltererSpec.PreFiltererSpec = &execinfrapb.InvertedFiltererSpec_PreFiltererSpec{
			Type: n.preFiltererType,
		}
		if invertedFiltererSpec.PreFiltererSpec.Expression, err = physicalplan.MakeExpression(
			n.preFiltererExpr, planCtx, nil); err != nil {
			__antithesis_instrumentation__.Notify(467176)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467177)
		}
	} else {
		__antithesis_instrumentation__.Notify(467178)
	}
	__antithesis_instrumentation__.Notify(467169)

	if len(plan.ResultRouters) == 1 {
		__antithesis_instrumentation__.Notify(467179)

		lastSQLInstanceID := plan.Processors[plan.ResultRouters[0]].SQLInstanceID
		plan.AddSingleGroupStage(lastSQLInstanceID,
			execinfrapb.ProcessorCoreUnion{
				InvertedFilterer: invertedFiltererSpec,
			},
			execinfrapb.PostProcessSpec{}, plan.GetResultTypes())
		return plan, nil
	} else {
		__antithesis_instrumentation__.Notify(467180)
	}
	__antithesis_instrumentation__.Notify(467170)

	distributable := n.expression.Left == nil && func() bool {
		__antithesis_instrumentation__.Notify(467181)
		return n.expression.Right == nil == true
	}() == true
	if !distributable {
		__antithesis_instrumentation__.Notify(467182)
		return nil, errors.Errorf("expected distributable inverted filterer")
	} else {
		__antithesis_instrumentation__.Notify(467183)
	}
	__antithesis_instrumentation__.Notify(467171)
	reqOrdering := execinfrapb.Ordering{}

	plan.AddNoGroupingStage(
		execinfrapb.ProcessorCoreUnion{InvertedFilterer: invertedFiltererSpec},
		execinfrapb.PostProcessSpec{}, plan.GetResultTypes(), reqOrdering,
	)

	distinctColumns := make([]uint32, 0, len(n.resultColumns)-1)
	for i := 0; i < len(n.resultColumns); i++ {
		__antithesis_instrumentation__.Notify(467184)
		if i == n.invColumn {
			__antithesis_instrumentation__.Notify(467186)
			continue
		} else {
			__antithesis_instrumentation__.Notify(467187)
		}
		__antithesis_instrumentation__.Notify(467185)
		distinctColumns = append(distinctColumns, uint32(i))
	}
	__antithesis_instrumentation__.Notify(467172)
	plan.AddSingleGroupStage(
		dsp.gatewaySQLInstanceID,
		execinfrapb.ProcessorCoreUnion{
			Distinct: dsp.createDistinctSpec(
				distinctColumns,
				[]uint32{},
				false,
				"",
				reqOrdering,
			),
		},
		execinfrapb.PostProcessSpec{},
		plan.GetResultTypes(),
	)
	return plan, nil
}

func getTypesFromResultColumns(cols colinfo.ResultColumns) []*types.T {
	__antithesis_instrumentation__.Notify(467188)
	typs := make([]*types.T, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(467190)
		typs[i] = col.Typ
	}
	__antithesis_instrumentation__.Notify(467189)
	return typs
}

func getTypesForPlanResult(node planNode, planToStreamColMap []int) ([]*types.T, error) {
	__antithesis_instrumentation__.Notify(467191)
	nodeColumns := planColumns(node)
	if planToStreamColMap == nil {
		__antithesis_instrumentation__.Notify(467195)

		return getTypesFromResultColumns(nodeColumns), nil
	} else {
		__antithesis_instrumentation__.Notify(467196)
	}
	__antithesis_instrumentation__.Notify(467192)
	numCols := 0
	for _, streamCol := range planToStreamColMap {
		__antithesis_instrumentation__.Notify(467197)
		if numCols <= streamCol {
			__antithesis_instrumentation__.Notify(467198)
			numCols = streamCol + 1
		} else {
			__antithesis_instrumentation__.Notify(467199)
		}
	}
	__antithesis_instrumentation__.Notify(467193)
	types := make([]*types.T, numCols)
	for nodeCol, streamCol := range planToStreamColMap {
		__antithesis_instrumentation__.Notify(467200)
		if streamCol != -1 {
			__antithesis_instrumentation__.Notify(467201)
			types[streamCol] = nodeColumns[nodeCol].Typ
		} else {
			__antithesis_instrumentation__.Notify(467202)
		}
	}
	__antithesis_instrumentation__.Notify(467194)
	return types, nil
}

func (dsp *DistSQLPlanner) createPlanForJoin(
	ctx context.Context, planCtx *PlanningCtx, n *joinNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467203)
	leftPlan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.left.plan)
	if err != nil {
		__antithesis_instrumentation__.Notify(467208)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467209)
	}
	__antithesis_instrumentation__.Notify(467204)
	rightPlan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.right.plan)
	if err != nil {
		__antithesis_instrumentation__.Notify(467210)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467211)
	}
	__antithesis_instrumentation__.Notify(467205)

	leftMap, rightMap := leftPlan.PlanToStreamColMap, rightPlan.PlanToStreamColMap
	helper := &joinPlanningHelper{
		numLeftOutCols:          n.pred.numLeftCols,
		numRightOutCols:         n.pred.numRightCols,
		numAllLeftCols:          len(leftPlan.GetResultTypes()),
		leftPlanToStreamColMap:  leftMap,
		rightPlanToStreamColMap: rightMap,
	}
	post, joinToStreamColMap := helper.joinOutColumns(n.pred.joinType, n.columns)
	onExpr, err := helper.remapOnExpr(planCtx, n.pred.onCond)
	if err != nil {
		__antithesis_instrumentation__.Notify(467212)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467213)
	}
	__antithesis_instrumentation__.Notify(467206)

	leftEqCols := eqCols(n.pred.leftEqualityIndices, leftMap)
	rightEqCols := eqCols(n.pred.rightEqualityIndices, rightMap)
	leftMergeOrd := distsqlOrdering(n.mergeJoinOrdering, leftEqCols)
	rightMergeOrd := distsqlOrdering(n.mergeJoinOrdering, rightEqCols)

	joinResultTypes, err := getTypesForPlanResult(n, joinToStreamColMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(467214)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467215)
	}
	__antithesis_instrumentation__.Notify(467207)

	info := joinPlanningInfo{
		leftPlan:           leftPlan,
		rightPlan:          rightPlan,
		joinType:           n.pred.joinType,
		joinResultTypes:    joinResultTypes,
		onExpr:             onExpr,
		post:               post,
		joinToStreamColMap: joinToStreamColMap,
		leftEqCols:         leftEqCols,
		rightEqCols:        rightEqCols,
		leftEqColsAreKey:   n.pred.leftEqKey,
		rightEqColsAreKey:  n.pred.rightEqKey,
		leftMergeOrd:       leftMergeOrd,
		rightMergeOrd:      rightMergeOrd,

		leftPlanDistribution:  leftPlan.GetLastStageDistribution(),
		rightPlanDistribution: rightPlan.GetLastStageDistribution(),
	}
	return dsp.planJoiners(planCtx, &info, n.reqOrdering), nil
}

func (dsp *DistSQLPlanner) planJoiners(
	planCtx *PlanningCtx, info *joinPlanningInfo, reqOrdering ReqOrdering,
) *PhysicalPlan {
	__antithesis_instrumentation__.Notify(467216)

	p := planCtx.NewPhysicalPlan()
	physicalplan.MergePlans(
		&p.PhysicalPlan, &info.leftPlan.PhysicalPlan, &info.rightPlan.PhysicalPlan,
		info.leftPlanDistribution, info.rightPlanDistribution, info.allowPartialDistribution,
	)
	leftRouters := info.leftPlan.ResultRouters
	rightRouters := info.rightPlan.ResultRouters

	var sqlInstances []base.SQLInstanceID
	if numEq := len(info.leftEqCols); numEq != 0 {
		__antithesis_instrumentation__.Notify(467218)
		sqlInstances = findJoinProcessorNodes(leftRouters, rightRouters, p.Processors)
	} else {
		__antithesis_instrumentation__.Notify(467219)

		sqlInstances = []base.SQLInstanceID{dsp.gatewaySQLInstanceID}

		if len(leftRouters) == 1 {
			__antithesis_instrumentation__.Notify(467220)
			sqlInstances[0] = p.Processors[leftRouters[0]].SQLInstanceID
		} else {
			__antithesis_instrumentation__.Notify(467221)
			if len(rightRouters) == 1 {
				__antithesis_instrumentation__.Notify(467222)
				sqlInstances[0] = p.Processors[rightRouters[0]].SQLInstanceID
			} else {
				__antithesis_instrumentation__.Notify(467223)
			}
		}
	}
	__antithesis_instrumentation__.Notify(467217)

	p.AddJoinStage(
		sqlInstances, info.makeCoreSpec(), info.post,
		info.leftEqCols, info.rightEqCols,
		info.leftPlan.GetResultTypes(), info.rightPlan.GetResultTypes(),
		info.leftMergeOrd, info.rightMergeOrd,
		leftRouters, rightRouters, info.joinResultTypes,
	)

	p.PlanToStreamColMap = info.joinToStreamColMap

	p.SetMergeOrdering(dsp.convertOrdering(reqOrdering, p.PlanToStreamColMap))
	return p
}

func (dsp *DistSQLPlanner) createPhysPlan(
	ctx context.Context, planCtx *PlanningCtx, plan planMaybePhysical,
) (physPlan *PhysicalPlan, cleanup func(), err error) {
	__antithesis_instrumentation__.Notify(467224)
	if plan.isPhysicalPlan() {
		__antithesis_instrumentation__.Notify(467226)

		return plan.physPlan.PhysicalPlan, func() { __antithesis_instrumentation__.Notify(467227) }, nil
	} else {
		__antithesis_instrumentation__.Notify(467228)
	}
	__antithesis_instrumentation__.Notify(467225)
	physPlan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, plan.planNode)
	return physPlan, planCtx.getCleanupFunc(), err
}

func (dsp *DistSQLPlanner) createPhysPlanForPlanNode(
	ctx context.Context, planCtx *PlanningCtx, node planNode,
) (plan *PhysicalPlan, err error) {
	__antithesis_instrumentation__.Notify(467229)
	planCtx.planDepth++

	switch n := node.(type) {

	case *distinctNode:
		__antithesis_instrumentation__.Notify(467234)
		plan, err = dsp.createPlanForDistinct(ctx, planCtx, n)

	case *exportNode:
		__antithesis_instrumentation__.Notify(467235)
		plan, err = dsp.createPlanForExport(ctx, planCtx, n)

	case *filterNode:
		__antithesis_instrumentation__.Notify(467236)
		plan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, n.source.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(467266)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467267)
		}
		__antithesis_instrumentation__.Notify(467237)

		if err := plan.AddFilter(n.filter, planCtx, plan.PlanToStreamColMap); err != nil {
			__antithesis_instrumentation__.Notify(467268)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467269)
		}

	case *groupNode:
		__antithesis_instrumentation__.Notify(467238)
		plan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(467270)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467271)
		}
		__antithesis_instrumentation__.Notify(467239)

		if err := dsp.addAggregators(planCtx, plan, n); err != nil {
			__antithesis_instrumentation__.Notify(467272)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467273)
		}

	case *indexJoinNode:
		__antithesis_instrumentation__.Notify(467240)
		plan, err = dsp.createPlanForIndexJoin(ctx, planCtx, n)

	case *invertedFilterNode:
		__antithesis_instrumentation__.Notify(467241)
		plan, err = dsp.createPlanForInvertedFilter(ctx, planCtx, n)

	case *invertedJoinNode:
		__antithesis_instrumentation__.Notify(467242)
		plan, err = dsp.createPlanForInvertedJoin(ctx, planCtx, n)

	case *joinNode:
		__antithesis_instrumentation__.Notify(467243)
		plan, err = dsp.createPlanForJoin(ctx, planCtx, n)

	case *limitNode:
		__antithesis_instrumentation__.Notify(467244)
		plan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(467274)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467275)
		}
		__antithesis_instrumentation__.Notify(467245)
		var count, offset int64
		if count, offset, err = evalLimit(planCtx.EvalContext(), n.countExpr, n.offsetExpr); err != nil {
			__antithesis_instrumentation__.Notify(467276)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467277)
		}
		__antithesis_instrumentation__.Notify(467246)
		if err := plan.AddLimit(count, offset, planCtx); err != nil {
			__antithesis_instrumentation__.Notify(467278)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467279)
		}

	case *lookupJoinNode:
		__antithesis_instrumentation__.Notify(467247)
		plan, err = dsp.createPlanForLookupJoin(ctx, planCtx, n)

	case *ordinalityNode:
		__antithesis_instrumentation__.Notify(467248)
		plan, err = dsp.createPlanForOrdinality(ctx, planCtx, n)

	case *projectSetNode:
		__antithesis_instrumentation__.Notify(467249)
		plan, err = dsp.createPlanForProjectSet(ctx, planCtx, n)

	case *renderNode:
		__antithesis_instrumentation__.Notify(467250)
		plan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, n.source.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(467280)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467281)
		}
		__antithesis_instrumentation__.Notify(467251)
		err = dsp.createPlanForRender(plan, n, planCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(467282)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467283)
		}

	case *scanNode:
		__antithesis_instrumentation__.Notify(467252)
		plan, err = dsp.createTableReaders(ctx, planCtx, n)

	case *sortNode:
		__antithesis_instrumentation__.Notify(467253)
		plan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(467284)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467285)
		}
		__antithesis_instrumentation__.Notify(467254)

		dsp.addSorters(plan, n.ordering, n.alreadyOrderedPrefix, 0)

	case *topKNode:
		__antithesis_instrumentation__.Notify(467255)
		plan, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, n.plan)
		if err != nil {
			__antithesis_instrumentation__.Notify(467286)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467287)
		}
		__antithesis_instrumentation__.Notify(467256)

		if n.k <= 0 {
			__antithesis_instrumentation__.Notify(467288)
			return nil, errors.New("negative or zero value for LIMIT")
		} else {
			__antithesis_instrumentation__.Notify(467289)
		}
		__antithesis_instrumentation__.Notify(467257)
		dsp.addSorters(plan, n.ordering, n.alreadyOrderedPrefix, n.k)

	case *unaryNode:
		__antithesis_instrumentation__.Notify(467258)
		plan, err = dsp.createPlanForUnary(planCtx, n)

	case *unionNode:
		__antithesis_instrumentation__.Notify(467259)
		plan, err = dsp.createPlanForSetOp(ctx, planCtx, n)

	case *valuesNode:
		__antithesis_instrumentation__.Notify(467260)
		if mustWrapValuesNode(planCtx, n.specifiedInQuery) {
			__antithesis_instrumentation__.Notify(467290)
			plan, err = dsp.wrapPlan(ctx, planCtx, n, false)
		} else {
			__antithesis_instrumentation__.Notify(467291)
			colTypes := getTypesFromResultColumns(n.columns)
			var spec *execinfrapb.ValuesCoreSpec
			spec, err = dsp.createValuesSpecFromTuples(planCtx, n.tuples, colTypes)
			if err != nil {
				__antithesis_instrumentation__.Notify(467293)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(467294)
			}
			__antithesis_instrumentation__.Notify(467292)
			plan, err = dsp.createValuesPlan(planCtx, spec, colTypes)
		}

	case *windowNode:
		__antithesis_instrumentation__.Notify(467261)
		plan, err = dsp.createPlanForWindow(ctx, planCtx, n)

	case *zeroNode:
		__antithesis_instrumentation__.Notify(467262)
		plan, err = dsp.createPlanForZero(planCtx, n)

	case *zigzagJoinNode:
		__antithesis_instrumentation__.Notify(467263)
		plan, err = dsp.createPlanForZigzagJoin(planCtx, n)

	case *createStatsNode:
		__antithesis_instrumentation__.Notify(467264)
		if n.runAsJob {
			__antithesis_instrumentation__.Notify(467295)
			plan, err = dsp.wrapPlan(ctx, planCtx, n, false)
		} else {
			__antithesis_instrumentation__.Notify(467296)

			var record *jobs.Record
			record, err = n.makeJobRecord(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(467298)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(467299)
			}
			__antithesis_instrumentation__.Notify(467297)
			plan, err = dsp.createPlanForCreateStats(ctx, planCtx, 0,
				record.Details.(jobspb.CreateStatsDetails))
		}

	default:
		__antithesis_instrumentation__.Notify(467265)

		plan, err = dsp.wrapPlan(ctx, planCtx, n, false)
	}
	__antithesis_instrumentation__.Notify(467230)

	if err != nil {
		__antithesis_instrumentation__.Notify(467300)
		return plan, err
	} else {
		__antithesis_instrumentation__.Notify(467301)
	}
	__antithesis_instrumentation__.Notify(467231)

	if planCtx.traceMetadata != nil {
		__antithesis_instrumentation__.Notify(467302)
		processors := make(execComponents, len(plan.ResultRouters))
		for i, resultProcIdx := range plan.ResultRouters {
			__antithesis_instrumentation__.Notify(467304)
			processors[i] = execinfrapb.ProcessorComponentID(
				plan.Processors[resultProcIdx].SQLInstanceID,
				execinfrapb.FlowID{UUID: planCtx.infra.FlowID},
				int32(resultProcIdx),
			)
		}
		__antithesis_instrumentation__.Notify(467303)
		planCtx.traceMetadata.associateNodeWithComponents(node, processors)
	} else {
		__antithesis_instrumentation__.Notify(467305)
	}
	__antithesis_instrumentation__.Notify(467232)

	if dsp.shouldPlanTestMetadata() {
		__antithesis_instrumentation__.Notify(467306)
		if err := plan.CheckLastStagePost(); err != nil {
			__antithesis_instrumentation__.Notify(467308)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(467309)
		}
		__antithesis_instrumentation__.Notify(467307)
		plan.AddNoGroupingStageWithCoreFunc(
			func(_ int, _ *physicalplan.Processor) execinfrapb.ProcessorCoreUnion {
				__antithesis_instrumentation__.Notify(467310)
				return execinfrapb.ProcessorCoreUnion{
					MetadataTestSender: &execinfrapb.MetadataTestSenderSpec{
						ID: uuid.MakeV4().String(),
					},
				}
			},
			execinfrapb.PostProcessSpec{},
			plan.GetResultTypes(),
			plan.MergeOrdering,
		)
	} else {
		__antithesis_instrumentation__.Notify(467311)
	}
	__antithesis_instrumentation__.Notify(467233)

	return plan, err
}

func (dsp *DistSQLPlanner) wrapPlan(
	ctx context.Context, planCtx *PlanningCtx, n planNode, allowPartialDistribution bool,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467312)
	useFastPath := planCtx.planDepth == 1 && func() bool {
		__antithesis_instrumentation__.Notify(467319)
		return planCtx.stmtType == tree.RowsAffected == true
	}() == true

	seenTop := false
	nParents := uint32(0)
	p := planCtx.NewPhysicalPlan()

	var firstNotWrapped planNode
	if err := walkPlan(ctx, n, planObserver{
		enterNode: func(ctx context.Context, nodeName string, plan planNode) (bool, error) {
			__antithesis_instrumentation__.Notify(467320)
			switch plan.(type) {
			case *explainVecNode, *explainPlanNode, *explainDDLNode:
				__antithesis_instrumentation__.Notify(467324)

				return false, nil
			}
			__antithesis_instrumentation__.Notify(467321)
			if !seenTop {
				__antithesis_instrumentation__.Notify(467325)

				seenTop = true
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(467326)
			}
			__antithesis_instrumentation__.Notify(467322)
			var err error

			if !dsp.mustWrapNode(planCtx, plan) {
				__antithesis_instrumentation__.Notify(467327)
				firstNotWrapped = plan
				p, err = dsp.createPhysPlanForPlanNode(ctx, planCtx, plan)
				if err != nil {
					__antithesis_instrumentation__.Notify(467329)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(467330)
				}
				__antithesis_instrumentation__.Notify(467328)
				nParents++
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(467331)
			}
			__antithesis_instrumentation__.Notify(467323)
			return true, nil
		},
	}); err != nil {
		__antithesis_instrumentation__.Notify(467332)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467333)
	}
	__antithesis_instrumentation__.Notify(467313)
	if nParents > 1 {
		__antithesis_instrumentation__.Notify(467334)
		return nil, errors.Errorf("can't wrap plan %v %T with more than one input", n, n)
	} else {
		__antithesis_instrumentation__.Notify(467335)
	}
	__antithesis_instrumentation__.Notify(467314)

	evalCtx := *planCtx.ExtendedEvalCtx

	wrapper, err := makePlanNodeToRowSource(n,
		runParams{
			extendedEvalCtx: &evalCtx,
			p:               planCtx.planner,
		},
		useFastPath,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(467336)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467337)
	}
	__antithesis_instrumentation__.Notify(467315)
	wrapper.firstNotWrapped = firstNotWrapped

	localProcIdx := p.AddLocalProcessor(wrapper)
	var input []execinfrapb.InputSyncSpec
	if firstNotWrapped != nil {
		__antithesis_instrumentation__.Notify(467338)

		input = []execinfrapb.InputSyncSpec{{
			Type:        execinfrapb.InputSyncSpec_PARALLEL_UNORDERED,
			ColumnTypes: p.GetResultTypes(),
		}}
	} else {
		__antithesis_instrumentation__.Notify(467339)
	}
	__antithesis_instrumentation__.Notify(467316)
	name := nodeName(n)
	proc := physicalplan.Processor{
		SQLInstanceID: dsp.gatewaySQLInstanceID,
		Spec: execinfrapb.ProcessorSpec{
			Input: input,
			Core: execinfrapb.ProcessorCoreUnion{LocalPlanNode: &execinfrapb.LocalPlanNodeSpec{
				RowSourceIdx: uint32(localProcIdx),
				NumInputs:    nParents,
				Name:         name,
			}},
			Post: execinfrapb.PostProcessSpec{},
			Output: []execinfrapb.OutputRouterSpec{{
				Type: execinfrapb.OutputRouterSpec_PASS_THROUGH,
			}},

			StageID:     p.NewStage(false, allowPartialDistribution),
			ResultTypes: wrapper.outputTypes,
		},
	}
	pIdx := p.AddProcessor(proc)
	if firstNotWrapped != nil {
		__antithesis_instrumentation__.Notify(467340)

		p.MergeResultStreams(p.ResultRouters, 0, p.MergeOrdering, pIdx, 0, false)
	} else {
		__antithesis_instrumentation__.Notify(467341)
	}
	__antithesis_instrumentation__.Notify(467317)

	if cap(p.ResultRouters) < 1 {
		__antithesis_instrumentation__.Notify(467342)
		p.ResultRouters = make([]physicalplan.ProcessorIdx, 1)
	} else {
		__antithesis_instrumentation__.Notify(467343)
		p.ResultRouters = p.ResultRouters[:1]
	}
	__antithesis_instrumentation__.Notify(467318)
	p.ResultRouters[0] = pIdx
	p.PlanToStreamColMap = identityMapInPlace(make([]int, len(p.GetResultTypes())))
	return p, nil
}

func (dsp *DistSQLPlanner) createValuesSpec(
	planCtx *PlanningCtx, resultTypes []*types.T, numRows int, rawBytes [][]byte,
) *execinfrapb.ValuesCoreSpec {
	__antithesis_instrumentation__.Notify(467344)
	numColumns := len(resultTypes)
	s := &execinfrapb.ValuesCoreSpec{
		Columns: make([]execinfrapb.DatumInfo, numColumns),
	}

	for i, t := range resultTypes {
		__antithesis_instrumentation__.Notify(467346)
		s.Columns[i].Encoding = descpb.DatumEncoding_VALUE
		s.Columns[i].Type = t
	}
	__antithesis_instrumentation__.Notify(467345)

	s.NumRows = uint64(numRows)
	s.RawBytes = rawBytes

	return s
}

func (dsp *DistSQLPlanner) createValuesPlan(
	planCtx *PlanningCtx, spec *execinfrapb.ValuesCoreSpec, resultTypes []*types.T,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467347)
	p := planCtx.NewPhysicalPlan()

	pIdx := p.AddProcessor(physicalplan.Processor{

		SQLInstanceID: dsp.gatewaySQLInstanceID,
		Spec: execinfrapb.ProcessorSpec{
			Core:        execinfrapb.ProcessorCoreUnion{Values: spec},
			Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
			ResultTypes: resultTypes,
		},
	})
	p.ResultRouters = []physicalplan.ProcessorIdx{pIdx}
	p.Distribution = physicalplan.LocalPlan
	p.PlanToStreamColMap = identityMapInPlace(make([]int, len(resultTypes)))

	return p, nil
}

func (dsp *DistSQLPlanner) createValuesSpecFromTuples(
	planCtx *PlanningCtx, tuples [][]tree.TypedExpr, resultTypes []*types.T,
) (*execinfrapb.ValuesCoreSpec, error) {
	__antithesis_instrumentation__.Notify(467348)
	var a tree.DatumAlloc
	evalCtx := &planCtx.ExtendedEvalCtx.EvalContext
	numRows := len(tuples)
	if len(resultTypes) == 0 {
		__antithesis_instrumentation__.Notify(467351)

		spec := dsp.createValuesSpec(planCtx, resultTypes, numRows, nil)
		return spec, nil
	} else {
		__antithesis_instrumentation__.Notify(467352)
	}
	__antithesis_instrumentation__.Notify(467349)
	rawBytes := make([][]byte, numRows)
	for rowIdx, tuple := range tuples {
		__antithesis_instrumentation__.Notify(467353)
		var buf []byte
		for colIdx, typedExpr := range tuple {
			__antithesis_instrumentation__.Notify(467355)
			datum, err := typedExpr.Eval(evalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(467357)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(467358)
			}
			__antithesis_instrumentation__.Notify(467356)
			encDatum := rowenc.DatumToEncDatum(resultTypes[colIdx], datum)
			buf, err = encDatum.Encode(resultTypes[colIdx], &a, descpb.DatumEncoding_VALUE, buf)
			if err != nil {
				__antithesis_instrumentation__.Notify(467359)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(467360)
			}
		}
		__antithesis_instrumentation__.Notify(467354)
		rawBytes[rowIdx] = buf
	}
	__antithesis_instrumentation__.Notify(467350)
	spec := dsp.createValuesSpec(planCtx, resultTypes, numRows, rawBytes)
	return spec, nil
}

func (dsp *DistSQLPlanner) createPlanForUnary(
	planCtx *PlanningCtx, n *unaryNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467361)
	types, err := getTypesForPlanResult(n, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(467363)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467364)
	}
	__antithesis_instrumentation__.Notify(467362)

	spec := dsp.createValuesSpec(planCtx, types, 1, nil)
	return dsp.createValuesPlan(planCtx, spec, types)
}

func (dsp *DistSQLPlanner) createPlanForZero(
	planCtx *PlanningCtx, n *zeroNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467365)
	types, err := getTypesForPlanResult(n, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(467367)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467368)
	}
	__antithesis_instrumentation__.Notify(467366)

	spec := dsp.createValuesSpec(planCtx, types, 0, nil)
	return dsp.createValuesPlan(planCtx, spec, types)
}

func (dsp *DistSQLPlanner) createDistinctSpec(
	distinctColumns []uint32,
	orderedColumns []uint32,
	nullsAreDistinct bool,
	errorOnDup string,
	outputOrdering execinfrapb.Ordering,
) *execinfrapb.DistinctSpec {
	__antithesis_instrumentation__.Notify(467369)
	return &execinfrapb.DistinctSpec{
		OrderedColumns:   orderedColumns,
		DistinctColumns:  distinctColumns,
		NullsAreDistinct: nullsAreDistinct,
		ErrorOnDup:       errorOnDup,
		OutputOrdering:   outputOrdering,
	}
}

func (dsp *DistSQLPlanner) createPlanForDistinct(
	ctx context.Context, planCtx *PlanningCtx, n *distinctNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467370)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.plan)
	if err != nil {
		__antithesis_instrumentation__.Notify(467372)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467373)
	}
	__antithesis_instrumentation__.Notify(467371)
	spec := dsp.createDistinctSpec(
		convertFastIntSetToUint32Slice(n.distinctOnColIdxs),
		convertFastIntSetToUint32Slice(n.columnsInOrder),
		n.nullsAreDistinct,
		n.errorOnDup,
		dsp.convertOrdering(n.reqOrdering, plan.PlanToStreamColMap),
	)
	dsp.addDistinctProcessors(plan, spec)
	return plan, nil
}

func (dsp *DistSQLPlanner) addDistinctProcessors(
	plan *PhysicalPlan, spec *execinfrapb.DistinctSpec,
) {
	__antithesis_instrumentation__.Notify(467374)
	distinctSpec := execinfrapb.ProcessorCoreUnion{
		Distinct: spec,
	}

	plan.AddNoGroupingStage(distinctSpec, execinfrapb.PostProcessSpec{}, plan.GetResultTypes(), plan.MergeOrdering)
	if !plan.IsLastStageDistributed() {
		__antithesis_instrumentation__.Notify(467376)
		return
	} else {
		__antithesis_instrumentation__.Notify(467377)
	}
	__antithesis_instrumentation__.Notify(467375)

	sqlInstanceIDs := getSQLInstanceIDsOfRouters(plan.ResultRouters, plan.Processors)
	plan.AddStageOnNodes(
		sqlInstanceIDs, distinctSpec, execinfrapb.PostProcessSpec{},
		distinctSpec.Distinct.DistinctColumns, plan.GetResultTypes(),
		plan.GetResultTypes(), plan.MergeOrdering, plan.ResultRouters,
	)
	plan.SetMergeOrdering(spec.OutputOrdering)
}

func (dsp *DistSQLPlanner) createPlanForOrdinality(
	ctx context.Context, planCtx *PlanningCtx, n *ordinalityNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467378)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.source)
	if err != nil {
		__antithesis_instrumentation__.Notify(467380)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467381)
	}
	__antithesis_instrumentation__.Notify(467379)

	ordinalitySpec := execinfrapb.ProcessorCoreUnion{
		Ordinality: &execinfrapb.OrdinalitySpec{},
	}

	plan.PlanToStreamColMap = append(plan.PlanToStreamColMap, len(plan.GetResultTypes()))
	outputTypes := append(plan.GetResultTypes(), types.Int)

	plan.AddSingleGroupStage(dsp.gatewaySQLInstanceID, ordinalitySpec, execinfrapb.PostProcessSpec{}, outputTypes)

	return plan, nil
}

func createProjectSetSpec(
	planCtx *PlanningCtx, n *projectSetPlanningInfo, indexVarMap []int,
) (*execinfrapb.ProjectSetSpec, error) {
	__antithesis_instrumentation__.Notify(467382)
	spec := execinfrapb.ProjectSetSpec{
		Exprs:            make([]execinfrapb.Expression, len(n.exprs)),
		GeneratedColumns: make([]*types.T, len(n.columns)-n.numColsInSource),
		NumColsPerGen:    make([]uint32, len(n.exprs)),
	}
	for i, expr := range n.exprs {
		__antithesis_instrumentation__.Notify(467386)
		var err error
		spec.Exprs[i], err = physicalplan.MakeExpression(expr, planCtx, indexVarMap)
		if err != nil {
			__antithesis_instrumentation__.Notify(467387)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(467388)
		}
	}
	__antithesis_instrumentation__.Notify(467383)
	for i, col := range n.columns[n.numColsInSource:] {
		__antithesis_instrumentation__.Notify(467389)
		spec.GeneratedColumns[i] = col.Typ
	}
	__antithesis_instrumentation__.Notify(467384)
	for i, n := range n.numColsPerGen {
		__antithesis_instrumentation__.Notify(467390)
		spec.NumColsPerGen[i] = uint32(n)
	}
	__antithesis_instrumentation__.Notify(467385)
	return &spec, nil
}

func (dsp *DistSQLPlanner) createPlanForProjectSet(
	ctx context.Context, planCtx *PlanningCtx, n *projectSetNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467391)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.source)
	if err != nil {
		__antithesis_instrumentation__.Notify(467393)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467394)
	}
	__antithesis_instrumentation__.Notify(467392)
	err = dsp.addProjectSet(plan, planCtx, &n.projectSetPlanningInfo)
	return plan, err
}

func (dsp *DistSQLPlanner) addProjectSet(
	plan *PhysicalPlan, planCtx *PlanningCtx, info *projectSetPlanningInfo,
) error {
	__antithesis_instrumentation__.Notify(467395)
	numResults := len(plan.GetResultTypes())

	projectSetSpec, err := createProjectSetSpec(planCtx, info, plan.PlanToStreamColMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(467398)
		return err
	} else {
		__antithesis_instrumentation__.Notify(467399)
	}
	__antithesis_instrumentation__.Notify(467396)
	spec := execinfrapb.ProcessorCoreUnion{
		ProjectSet: projectSetSpec,
	}

	outputTypes := append(plan.GetResultTypes(), projectSetSpec.GeneratedColumns...)
	plan.AddSingleGroupStage(dsp.gatewaySQLInstanceID, spec, execinfrapb.PostProcessSpec{}, outputTypes)

	for i := range projectSetSpec.GeneratedColumns {
		__antithesis_instrumentation__.Notify(467400)
		plan.PlanToStreamColMap = append(plan.PlanToStreamColMap, numResults+i)
	}
	__antithesis_instrumentation__.Notify(467397)
	return nil
}

func (dsp *DistSQLPlanner) isOnlyOnGateway(plan *PhysicalPlan) bool {
	__antithesis_instrumentation__.Notify(467401)
	if len(plan.ResultRouters) == 1 {
		__antithesis_instrumentation__.Notify(467403)
		processorIdx := plan.ResultRouters[0]
		if plan.Processors[processorIdx].SQLInstanceID == dsp.gatewaySQLInstanceID {
			__antithesis_instrumentation__.Notify(467404)
			return true
		} else {
			__antithesis_instrumentation__.Notify(467405)
		}
	} else {
		__antithesis_instrumentation__.Notify(467406)
	}
	__antithesis_instrumentation__.Notify(467402)
	return false
}

func (dsp *DistSQLPlanner) createPlanForSetOp(
	ctx context.Context, planCtx *PlanningCtx, n *unionNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467407)
	leftLogicalPlan := n.left
	leftPlan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.left)
	if err != nil {
		__antithesis_instrumentation__.Notify(467416)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467417)
	}
	__antithesis_instrumentation__.Notify(467408)
	rightLogicalPlan := n.right
	rightPlan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.right)
	if err != nil {
		__antithesis_instrumentation__.Notify(467418)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467419)
	}
	__antithesis_instrumentation__.Notify(467409)
	if n.inverted {
		__antithesis_instrumentation__.Notify(467420)
		leftPlan, rightPlan = rightPlan, leftPlan
		leftLogicalPlan, rightLogicalPlan = rightLogicalPlan, leftLogicalPlan
	} else {
		__antithesis_instrumentation__.Notify(467421)
	}
	__antithesis_instrumentation__.Notify(467410)
	childPhysicalPlans := []*PhysicalPlan{leftPlan, rightPlan}

	if !reflect.DeepEqual(leftPlan.PlanToStreamColMap, rightPlan.PlanToStreamColMap) {
		__antithesis_instrumentation__.Notify(467422)
		return nil, errors.Errorf(
			"planToStreamColMap mismatch: %v, %v", leftPlan.PlanToStreamColMap,
			rightPlan.PlanToStreamColMap)
	} else {
		__antithesis_instrumentation__.Notify(467423)
	}
	__antithesis_instrumentation__.Notify(467411)
	planToStreamColMap := leftPlan.PlanToStreamColMap
	streamCols := make([]uint32, 0, len(planToStreamColMap))
	for _, streamCol := range planToStreamColMap {
		__antithesis_instrumentation__.Notify(467424)
		if streamCol < 0 {
			__antithesis_instrumentation__.Notify(467426)
			continue
		} else {
			__antithesis_instrumentation__.Notify(467427)
		}
		__antithesis_instrumentation__.Notify(467425)
		streamCols = append(streamCols, uint32(streamCol))
	}
	__antithesis_instrumentation__.Notify(467412)

	var distinctSpecs [2]execinfrapb.ProcessorCoreUnion

	if !n.all {
		__antithesis_instrumentation__.Notify(467428)
		var distinctOrds [2]execinfrapb.Ordering
		distinctOrds[0] = execinfrapb.ConvertToMappedSpecOrdering(
			planReqOrdering(leftLogicalPlan), leftPlan.PlanToStreamColMap,
		)
		distinctOrds[1] = execinfrapb.ConvertToMappedSpecOrdering(
			planReqOrdering(rightLogicalPlan), rightPlan.PlanToStreamColMap,
		)

		for side, plan := range childPhysicalPlans {
			__antithesis_instrumentation__.Notify(467429)
			sortCols := make([]uint32, len(distinctOrds[side].Columns))
			for i, ord := range distinctOrds[side].Columns {
				__antithesis_instrumentation__.Notify(467431)
				sortCols[i] = ord.ColIdx
			}
			__antithesis_instrumentation__.Notify(467430)
			distinctSpecs[side].Distinct = dsp.createDistinctSpec(
				streamCols,
				sortCols,
				false,
				"",
				distinctOrds[side],
			)
			if !dsp.isOnlyOnGateway(plan) {
				__antithesis_instrumentation__.Notify(467432)

				plan.AddNoGroupingStage(distinctSpecs[side], execinfrapb.PostProcessSpec{}, plan.GetResultTypes(), distinctOrds[side])
			} else {
				__antithesis_instrumentation__.Notify(467433)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(467434)
	}
	__antithesis_instrumentation__.Notify(467413)

	p := planCtx.NewPhysicalPlan()
	p.SetRowEstimates(&leftPlan.PhysicalPlan, &rightPlan.PhysicalPlan)

	p.PlanToStreamColMap = planToStreamColMap

	resultTypes, err := mergeResultTypesForSetOp(leftPlan, rightPlan)
	if err != nil {
		__antithesis_instrumentation__.Notify(467435)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467436)
	}
	__antithesis_instrumentation__.Notify(467414)

	mergeOrdering := dsp.convertOrdering(n.streamingOrdering, p.PlanToStreamColMap)

	leftRouters := leftPlan.ResultRouters
	rightRouters := rightPlan.ResultRouters
	physicalplan.MergePlans(
		&p.PhysicalPlan, &leftPlan.PhysicalPlan, &rightPlan.PhysicalPlan,

		leftPlan.GetLastStageDistribution(),
		rightPlan.GetLastStageDistribution(),
		false,
	)

	if n.unionType == tree.UnionOp {
		__antithesis_instrumentation__.Notify(467437)

		p.ResultRouters = append(leftRouters, rightRouters...)

		p.SetMergeOrdering(mergeOrdering)

		if !n.all {
			__antithesis_instrumentation__.Notify(467438)
			if n.hardLimit != 0 {
				__antithesis_instrumentation__.Notify(467440)
				return nil, errors.AssertionFailedf("a hard limit is not supported for UNION (only for UNION ALL)")
			} else {
				__antithesis_instrumentation__.Notify(467441)
			}
			__antithesis_instrumentation__.Notify(467439)

			distinctSpec := execinfrapb.ProcessorCoreUnion{
				Distinct: dsp.createDistinctSpec(
					streamCols,
					[]uint32{},
					false,
					"",
					mergeOrdering,
				),
			}
			p.AddSingleGroupStage(dsp.gatewaySQLInstanceID, distinctSpec, execinfrapb.PostProcessSpec{}, resultTypes)
		} else {
			__antithesis_instrumentation__.Notify(467442)

			if n.hardLimit == 0 {
				__antithesis_instrumentation__.Notify(467444)

				p.EnsureSingleStreamPerNode(true, execinfrapb.PostProcessSpec{})
			} else {
				__antithesis_instrumentation__.Notify(467445)
				if p.GetLastStageDistribution() != physicalplan.LocalPlan {
					__antithesis_instrumentation__.Notify(467448)
					return nil, errors.AssertionFailedf("we expect that limited UNION ALL queries are only planned locally")
				} else {
					__antithesis_instrumentation__.Notify(467449)
				}
				__antithesis_instrumentation__.Notify(467446)
				if len(p.MergeOrdering.Columns) != 0 {
					__antithesis_instrumentation__.Notify(467450)
					return nil, errors.AssertionFailedf(
						"we expect that limited UNION ALL queries do not require a specific ordering",
					)
				} else {
					__antithesis_instrumentation__.Notify(467451)
				}
				__antithesis_instrumentation__.Notify(467447)

				p.EnsureSingleStreamPerNode(
					true,
					execinfrapb.PostProcessSpec{Limit: n.hardLimit},
				)
			}
			__antithesis_instrumentation__.Notify(467443)

			if err := p.CheckLastStagePost(); err != nil {
				__antithesis_instrumentation__.Notify(467452)
				p.EnsureSingleStreamOnGateway()
			} else {
				__antithesis_instrumentation__.Notify(467453)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(467454)
		if n.hardLimit != 0 {
			__antithesis_instrumentation__.Notify(467458)
			return nil, errors.AssertionFailedf("a hard limit is not supported for INTERSECT or EXCEPT")
		} else {
			__antithesis_instrumentation__.Notify(467459)
		}
		__antithesis_instrumentation__.Notify(467455)

		joinType := distsqlSetOpJoinType(n.unionType)

		nodes := findJoinProcessorNodes(leftRouters, rightRouters, p.Processors)

		eqCols := streamCols

		post := execinfrapb.PostProcessSpec{Projection: true}
		post.OutputColumns = make([]uint32, len(streamCols))
		copy(post.OutputColumns, streamCols)

		var core execinfrapb.ProcessorCoreUnion
		if len(mergeOrdering.Columns) == 0 {
			__antithesis_instrumentation__.Notify(467460)
			core.HashJoiner = &execinfrapb.HashJoinerSpec{
				LeftEqColumns:  eqCols,
				RightEqColumns: eqCols,
				Type:           joinType,
			}
		} else {
			__antithesis_instrumentation__.Notify(467461)
			if len(mergeOrdering.Columns) < len(streamCols) {
				__antithesis_instrumentation__.Notify(467463)
				return nil, errors.AssertionFailedf("the merge ordering must include all stream columns")
			} else {
				__antithesis_instrumentation__.Notify(467464)
			}
			__antithesis_instrumentation__.Notify(467462)
			core.MergeJoiner = &execinfrapb.MergeJoinerSpec{
				LeftOrdering:  mergeOrdering,
				RightOrdering: mergeOrdering,
				Type:          joinType,
				NullEquality:  true,
			}
		}
		__antithesis_instrumentation__.Notify(467456)

		if n.all {
			__antithesis_instrumentation__.Notify(467465)
			p.AddJoinStage(
				nodes, core, post, eqCols, eqCols,
				leftPlan.GetResultTypes(), rightPlan.GetResultTypes(),
				leftPlan.MergeOrdering, rightPlan.MergeOrdering,
				leftRouters, rightRouters, resultTypes,
			)
		} else {
			__antithesis_instrumentation__.Notify(467466)
			p.AddDistinctSetOpStage(
				nodes, core, distinctSpecs[:], post, eqCols,
				leftPlan.GetResultTypes(), rightPlan.GetResultTypes(),
				leftPlan.MergeOrdering, rightPlan.MergeOrdering,
				leftRouters, rightRouters, resultTypes,
			)
		}
		__antithesis_instrumentation__.Notify(467457)

		p.SetMergeOrdering(mergeOrdering)
	}
	__antithesis_instrumentation__.Notify(467415)

	return p, nil
}

func (dsp *DistSQLPlanner) createPlanForWindow(
	ctx context.Context, planCtx *PlanningCtx, n *windowNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467467)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.plan)
	if err != nil {
		__antithesis_instrumentation__.Notify(467472)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467473)
	}
	__antithesis_instrumentation__.Notify(467468)

	numWindowFuncProcessed := 0
	windowPlanState := createWindowPlanState(n, planCtx, plan)

	for numWindowFuncProcessed < len(n.funcs) {
		__antithesis_instrumentation__.Notify(467474)
		samePartitionFuncs, partitionIdxs := windowPlanState.findUnprocessedWindowFnsWithSamePartition()
		numWindowFuncProcessed += len(samePartitionFuncs)
		windowerSpec := execinfrapb.WindowerSpec{
			PartitionBy: partitionIdxs,
			WindowFns:   make([]execinfrapb.WindowerSpec_WindowFn, len(samePartitionFuncs)),
		}

		newResultTypes := make([]*types.T, len(plan.GetResultTypes())+len(samePartitionFuncs))
		copy(newResultTypes, plan.GetResultTypes())
		for windowFnSpecIdx, windowFn := range samePartitionFuncs {
			__antithesis_instrumentation__.Notify(467477)
			windowFnSpec, outputType, err := windowPlanState.createWindowFnSpec(windowFn)
			if err != nil {
				__antithesis_instrumentation__.Notify(467479)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(467480)
			}
			__antithesis_instrumentation__.Notify(467478)
			newResultTypes[windowFn.outputColIdx] = outputType
			windowerSpec.WindowFns[windowFnSpecIdx] = windowFnSpec
		}
		__antithesis_instrumentation__.Notify(467475)

		prevStageNode := plan.Processors[plan.ResultRouters[0]].SQLInstanceID
		for i := 1; i < len(plan.ResultRouters); i++ {
			__antithesis_instrumentation__.Notify(467481)
			if n := plan.Processors[plan.ResultRouters[i]].SQLInstanceID; n != prevStageNode {
				__antithesis_instrumentation__.Notify(467482)
				prevStageNode = 0
				break
			} else {
				__antithesis_instrumentation__.Notify(467483)
			}
		}
		__antithesis_instrumentation__.Notify(467476)

		sqlInstanceIDs := getSQLInstanceIDsOfRouters(plan.ResultRouters, plan.Processors)
		if len(partitionIdxs) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(467484)
			return len(sqlInstanceIDs) == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(467485)

			sqlInstanceID := dsp.gatewaySQLInstanceID
			if len(sqlInstanceIDs) == 1 {
				__antithesis_instrumentation__.Notify(467487)
				sqlInstanceID = sqlInstanceIDs[0]
			} else {
				__antithesis_instrumentation__.Notify(467488)
			}
			__antithesis_instrumentation__.Notify(467486)
			plan.AddSingleGroupStage(
				sqlInstanceID,
				execinfrapb.ProcessorCoreUnion{Windower: &windowerSpec},
				execinfrapb.PostProcessSpec{},
				newResultTypes,
			)
		} else {
			__antithesis_instrumentation__.Notify(467489)

			for _, resultProc := range plan.ResultRouters {
				__antithesis_instrumentation__.Notify(467491)
				plan.Processors[resultProc].Spec.Output[0] = execinfrapb.OutputRouterSpec{
					Type:        execinfrapb.OutputRouterSpec_BY_HASH,
					HashColumns: partitionIdxs,
				}
			}
			__antithesis_instrumentation__.Notify(467490)

			stageID := plan.NewStage(true, false)

			prevStageRouters := plan.ResultRouters
			prevStageResultTypes := plan.GetResultTypes()
			plan.ResultRouters = make([]physicalplan.ProcessorIdx, 0, len(sqlInstanceIDs))
			for bucket, sqlInstanceID := range sqlInstanceIDs {
				__antithesis_instrumentation__.Notify(467492)
				proc := physicalplan.Processor{
					SQLInstanceID: sqlInstanceID,
					Spec: execinfrapb.ProcessorSpec{
						Input: []execinfrapb.InputSyncSpec{{
							Type:        execinfrapb.InputSyncSpec_PARALLEL_UNORDERED,
							ColumnTypes: prevStageResultTypes,
						}},
						Core: execinfrapb.ProcessorCoreUnion{Windower: &windowerSpec},
						Post: execinfrapb.PostProcessSpec{},
						Output: []execinfrapb.OutputRouterSpec{{
							Type: execinfrapb.OutputRouterSpec_PASS_THROUGH,
						}},
						StageID:     stageID,
						ResultTypes: newResultTypes,
					},
				}
				pIdx := plan.AddProcessor(proc)

				for _, router := range prevStageRouters {
					__antithesis_instrumentation__.Notify(467494)
					plan.Streams = append(plan.Streams, physicalplan.Stream{
						SourceProcessor:  router,
						SourceRouterSlot: bucket,
						DestProcessor:    pIdx,
						DestInput:        0,
					})
				}
				__antithesis_instrumentation__.Notify(467493)
				plan.ResultRouters = append(plan.ResultRouters, pIdx)
			}
		}
	}
	__antithesis_instrumentation__.Notify(467469)

	plan.PlanToStreamColMap = identityMap(plan.PlanToStreamColMap, len(plan.GetResultTypes()))

	plan.SetMergeOrdering(execinfrapb.Ordering{})

	if err := windowPlanState.addRenderingOrProjection(); err != nil {
		__antithesis_instrumentation__.Notify(467495)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467496)
	}
	__antithesis_instrumentation__.Notify(467470)

	if len(plan.GetResultTypes()) != len(plan.PlanToStreamColMap) {
		__antithesis_instrumentation__.Notify(467497)

		plan.PlanToStreamColMap = identityMap(plan.PlanToStreamColMap, len(plan.GetResultTypes()))
	} else {
		__antithesis_instrumentation__.Notify(467498)
	}
	__antithesis_instrumentation__.Notify(467471)

	return plan, nil
}

func (dsp *DistSQLPlanner) createPlanForExport(
	ctx context.Context, planCtx *PlanningCtx, n *exportNode,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467499)
	plan, err := dsp.createPhysPlanForPlanNode(ctx, planCtx, n.source)
	if err != nil {
		__antithesis_instrumentation__.Notify(467502)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467503)
	}
	__antithesis_instrumentation__.Notify(467500)

	var core execinfrapb.ProcessorCoreUnion
	core.Exporter = &execinfrapb.ExportSpec{
		Destination: n.destination,
		NamePattern: n.fileNamePattern,
		Format:      n.format,
		ChunkRows:   int64(n.chunkRows),
		ChunkSize:   n.chunkSize,
		ColNames:    n.colNames,
		UserProto:   planCtx.planner.User().EncodeProto(),
	}

	resTypes := make([]*types.T, len(colinfo.ExportColumns))
	for i := range colinfo.ExportColumns {
		__antithesis_instrumentation__.Notify(467504)
		resTypes[i] = colinfo.ExportColumns[i].Typ
	}
	__antithesis_instrumentation__.Notify(467501)
	plan.AddNoGroupingStage(
		core, execinfrapb.PostProcessSpec{}, resTypes, execinfrapb.Ordering{},
	)

	plan.PlanToStreamColMap = identityMap(plan.PlanToStreamColMap, len(colinfo.ExportColumns))
	return plan, nil
}

func checkScanParallelizationIfLocal(
	ctx context.Context, plan *planComponents,
) (prohibitParallelization, hasScanNodeToParallelize bool) {
	__antithesis_instrumentation__.Notify(467505)
	if plan.main.planNode == nil || func() bool {
		__antithesis_instrumentation__.Notify(467509)
		return len(plan.cascades) != 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(467510)
		return len(plan.checkPlans) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(467511)

		return true, false
	} else {
		__antithesis_instrumentation__.Notify(467512)
	}
	__antithesis_instrumentation__.Notify(467506)
	o := planObserver{
		enterNode: func(ctx context.Context, nodeName string, plan planNode) (bool, error) {
			__antithesis_instrumentation__.Notify(467513)
			if prohibitParallelization {
				__antithesis_instrumentation__.Notify(467515)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(467516)
			}
			__antithesis_instrumentation__.Notify(467514)
			switch n := plan.(type) {
			case *distinctNode:
				__antithesis_instrumentation__.Notify(467517)
				return true, nil
			case *explainPlanNode:
				__antithesis_instrumentation__.Notify(467518)

				plan := n.plan.WrappedPlan.(*planComponents)
				prohibit, has := checkScanParallelizationIfLocal(ctx, plan)
				prohibitParallelization = prohibitParallelization || func() bool {
					__antithesis_instrumentation__.Notify(467535)
					return prohibit == true
				}() == true
				hasScanNodeToParallelize = hasScanNodeToParallelize || func() bool {
					__antithesis_instrumentation__.Notify(467536)
					return has == true
				}() == true
				return false, nil
			case *explainVecNode:
				__antithesis_instrumentation__.Notify(467519)
				return true, nil
			case *filterNode:
				__antithesis_instrumentation__.Notify(467520)

				prohibitParallelization = true
				return false, nil
			case *groupNode:
				__antithesis_instrumentation__.Notify(467521)
				for _, f := range n.funcs {
					__antithesis_instrumentation__.Notify(467537)
					prohibitParallelization = f.hasFilter()
				}
				__antithesis_instrumentation__.Notify(467522)
				return true, nil
			case *indexJoinNode:
				__antithesis_instrumentation__.Notify(467523)
				return true, nil
			case *joinNode:
				__antithesis_instrumentation__.Notify(467524)
				prohibitParallelization = n.pred.onCond != nil
				return true, nil
			case *limitNode:
				__antithesis_instrumentation__.Notify(467525)
				return true, nil
			case *ordinalityNode:
				__antithesis_instrumentation__.Notify(467526)
				return true, nil
			case *renderNode:
				__antithesis_instrumentation__.Notify(467527)

				for _, e := range n.render {
					__antithesis_instrumentation__.Notify(467538)
					if _, isIVar := e.(*tree.IndexedVar); !isIVar {
						__antithesis_instrumentation__.Notify(467539)
						prohibitParallelization = true
					} else {
						__antithesis_instrumentation__.Notify(467540)
					}
				}
				__antithesis_instrumentation__.Notify(467528)
				return true, nil
			case *scanNode:
				__antithesis_instrumentation__.Notify(467529)
				if len(n.reqOrdering) == 0 && func() bool {
					__antithesis_instrumentation__.Notify(467541)
					return n.parallelize == true
				}() == true {
					__antithesis_instrumentation__.Notify(467542)
					hasScanNodeToParallelize = true
				} else {
					__antithesis_instrumentation__.Notify(467543)
				}
				__antithesis_instrumentation__.Notify(467530)
				return true, nil
			case *sortNode:
				__antithesis_instrumentation__.Notify(467531)
				return true, nil
			case *unionNode:
				__antithesis_instrumentation__.Notify(467532)
				return true, nil
			case *valuesNode:
				__antithesis_instrumentation__.Notify(467533)
				return true, nil
			default:
				__antithesis_instrumentation__.Notify(467534)
				prohibitParallelization = true
				return false, nil
			}
		},
	}
	__antithesis_instrumentation__.Notify(467507)
	_ = walkPlan(ctx, plan.main.planNode, o)
	for _, s := range plan.subqueryPlans {
		__antithesis_instrumentation__.Notify(467544)
		_ = walkPlan(ctx, s.plan.planNode, o)
	}
	__antithesis_instrumentation__.Notify(467508)
	return prohibitParallelization, hasScanNodeToParallelize
}

func (dsp *DistSQLPlanner) NewPlanningCtx(
	ctx context.Context,
	evalCtx *extendedEvalContext,
	planner *planner,
	txn *kv.Txn,
	distributionType DistributionType,
) *PlanningCtx {
	__antithesis_instrumentation__.Notify(467545)
	distribute := distributionType == DistributionTypeAlways || func() bool {
		__antithesis_instrumentation__.Notify(467547)
		return (distributionType == DistributionTypeSystemTenantOnly && func() bool {
			__antithesis_instrumentation__.Notify(467548)
			return evalCtx.Codec.ForSystemTenant() == true
		}() == true) == true
	}() == true
	planCtx := &PlanningCtx{
		ExtendedEvalCtx: evalCtx,
		infra:           physicalplan.MakePhysicalInfrastructure(uuid.FastMakeV4(), dsp.gatewaySQLInstanceID),
		isLocal:         !distribute,
		planner:         planner,
	}
	if !distribute {
		__antithesis_instrumentation__.Notify(467549)
		if planner == nil || func() bool {
			__antithesis_instrumentation__.Notify(467552)
			return dsp.spanResolver == nil == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(467553)
			return planner.curPlan.flags.IsSet(planFlagContainsMutation) == true
		}() == true {
			__antithesis_instrumentation__.Notify(467554)

			return planCtx
		} else {
			__antithesis_instrumentation__.Notify(467555)
		}
		__antithesis_instrumentation__.Notify(467550)
		prohibitParallelization, hasScanNodeToParallelize := checkScanParallelizationIfLocal(ctx, &planner.curPlan.planComponents)
		if prohibitParallelization || func() bool {
			__antithesis_instrumentation__.Notify(467556)
			return !hasScanNodeToParallelize == true
		}() == true {
			__antithesis_instrumentation__.Notify(467557)
			return planCtx
		} else {
			__antithesis_instrumentation__.Notify(467558)
		}
		__antithesis_instrumentation__.Notify(467551)

		planCtx.parallelizeScansIfLocal = true
	} else {
		__antithesis_instrumentation__.Notify(467559)
	}
	__antithesis_instrumentation__.Notify(467546)
	planCtx.spanIter = dsp.spanResolver.NewSpanResolverIterator(txn)
	planCtx.NodeStatuses = make(map[base.SQLInstanceID]NodeStatus)
	planCtx.NodeStatuses[dsp.gatewaySQLInstanceID] = NodeOK
	return planCtx
}

func maybeMoveSingleFlowToGateway(planCtx *PlanningCtx, plan *PhysicalPlan, rowCount int64) {
	__antithesis_instrumentation__.Notify(467560)
	if !planCtx.isLocal && func() bool {
		__antithesis_instrumentation__.Notify(467561)
		return planCtx.ExtendedEvalCtx.SessionData().DistSQLMode != sessiondatapb.DistSQLAlways == true
	}() == true {
		__antithesis_instrumentation__.Notify(467562)

		const rowReductionRatio = 10
		keepPlan := rowCount <= 0 || func() bool {
			__antithesis_instrumentation__.Notify(467565)
			return float64(plan.TotalEstimatedScannedRows)/float64(rowCount) >= rowReductionRatio == true
		}() == true
		if keepPlan {
			__antithesis_instrumentation__.Notify(467566)
			return
		} else {
			__antithesis_instrumentation__.Notify(467567)
		}
		__antithesis_instrumentation__.Notify(467563)
		singleFlow := true
		moveFlowToGateway := false
		sqlInstanceID := plan.Processors[0].SQLInstanceID
		for _, p := range plan.Processors[1:] {
			__antithesis_instrumentation__.Notify(467568)
			if p.SQLInstanceID != sqlInstanceID {
				__antithesis_instrumentation__.Notify(467570)
				if p.SQLInstanceID != plan.GatewaySQLInstanceID || func() bool {
					__antithesis_instrumentation__.Notify(467571)
					return p.Spec.Core.Noop == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(467572)

					singleFlow = false
					break
				} else {
					__antithesis_instrumentation__.Notify(467573)
				}
			} else {
				__antithesis_instrumentation__.Notify(467574)
			}
			__antithesis_instrumentation__.Notify(467569)
			core := p.Spec.Core
			if core.JoinReader != nil || func() bool {
				__antithesis_instrumentation__.Notify(467575)
				return core.MergeJoiner != nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(467576)
				return core.HashJoiner != nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(467577)
				return core.ZigzagJoiner != nil == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(467578)
				return core.InvertedJoiner != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(467579)

				moveFlowToGateway = true
			} else {
				__antithesis_instrumentation__.Notify(467580)
			}
		}
		__antithesis_instrumentation__.Notify(467564)
		if singleFlow && func() bool {
			__antithesis_instrumentation__.Notify(467581)
			return moveFlowToGateway == true
		}() == true {
			__antithesis_instrumentation__.Notify(467582)
			for i := range plan.Processors {
				__antithesis_instrumentation__.Notify(467584)
				plan.Processors[i].SQLInstanceID = plan.GatewaySQLInstanceID
			}
			__antithesis_instrumentation__.Notify(467583)
			planCtx.isLocal = true
			planCtx.planner.curPlan.flags.Unset(planFlagFullyDistributed)
			planCtx.planner.curPlan.flags.Unset(planFlagPartiallyDistributed)
			plan.Distribution = physicalplan.LocalPlan
		} else {
			__antithesis_instrumentation__.Notify(467585)
		}
	} else {
		__antithesis_instrumentation__.Notify(467586)
	}
}

func (dsp *DistSQLPlanner) FinalizePlan(planCtx *PlanningCtx, plan *PhysicalPlan) {
	__antithesis_instrumentation__.Notify(467587)
	dsp.finalizePlanWithRowCount(planCtx, plan, -1)
}

func (dsp *DistSQLPlanner) finalizePlanWithRowCount(
	planCtx *PlanningCtx, plan *PhysicalPlan, rowCount int64,
) {
	__antithesis_instrumentation__.Notify(467588)

	var metadataSenders []string
	for _, proc := range plan.Processors {
		__antithesis_instrumentation__.Notify(467592)
		if proc.Spec.Core.MetadataTestSender != nil {
			__antithesis_instrumentation__.Notify(467593)
			metadataSenders = append(metadataSenders, proc.Spec.Core.MetadataTestSender.ID)
		} else {
			__antithesis_instrumentation__.Notify(467594)
		}
	}
	__antithesis_instrumentation__.Notify(467589)

	maybeMoveSingleFlowToGateway(planCtx, plan, rowCount)

	plan.EnsureSingleStreamOnGateway()

	projection := make([]uint32, 0, len(plan.GetResultTypes()))
	for _, outputCol := range plan.PlanToStreamColMap {
		__antithesis_instrumentation__.Notify(467595)
		if outputCol >= 0 {
			__antithesis_instrumentation__.Notify(467596)
			projection = append(projection, uint32(outputCol))
		} else {
			__antithesis_instrumentation__.Notify(467597)
		}
	}
	__antithesis_instrumentation__.Notify(467590)
	plan.AddProjection(projection, execinfrapb.Ordering{})

	plan.PlanToStreamColMap = nil

	if len(metadataSenders) > 0 {
		__antithesis_instrumentation__.Notify(467598)
		plan.AddSingleGroupStage(
			dsp.gatewaySQLInstanceID,
			execinfrapb.ProcessorCoreUnion{
				MetadataTestReceiver: &execinfrapb.MetadataTestReceiverSpec{
					SenderIDs: metadataSenders,
				},
			},
			execinfrapb.PostProcessSpec{},
			plan.GetResultTypes(),
		)
	} else {
		__antithesis_instrumentation__.Notify(467599)
	}
	__antithesis_instrumentation__.Notify(467591)

	plan.PopulateEndpoints()

	finalOut := &plan.Processors[plan.ResultRouters[0]].Spec.Output[0]
	finalOut.Streams = append(finalOut.Streams, execinfrapb.StreamEndpointSpec{
		Type: execinfrapb.StreamEndpointSpec_SYNC_RESPONSE,
	})

	for i := range plan.Processors {
		__antithesis_instrumentation__.Notify(467600)
		plan.Processors[i].Spec.ProcessorID = int32(i)
	}
}
