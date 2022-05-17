package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/errors"
)

func (dsp *DistSQLPlanner) SetupAllNodesPlanning(
	ctx context.Context, evalCtx *extendedEvalContext, execCfg *ExecutorConfig,
) (*PlanningCtx, []base.SQLInstanceID, error) {
	__antithesis_instrumentation__.Notify(467634)
	if dsp.codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(467636)
		return dsp.setupAllNodesPlanningSystem(ctx, evalCtx, execCfg)
	} else {
		__antithesis_instrumentation__.Notify(467637)
	}
	__antithesis_instrumentation__.Notify(467635)
	return dsp.setupAllNodesPlanningTenant(ctx, evalCtx, execCfg)
}

func (dsp *DistSQLPlanner) setupAllNodesPlanningSystem(
	ctx context.Context, evalCtx *extendedEvalContext, execCfg *ExecutorConfig,
) (*PlanningCtx, []base.SQLInstanceID, error) {
	__antithesis_instrumentation__.Notify(467638)
	planCtx := dsp.NewPlanningCtx(ctx, evalCtx, nil, nil,
		DistributionTypeAlways)

	ss, err := execCfg.NodesStatusServer.OptionalNodesStatusServer(47900)
	if err != nil {
		__antithesis_instrumentation__.Notify(467644)
		return planCtx, []base.SQLInstanceID{dsp.gatewaySQLInstanceID}, nil
	} else {
		__antithesis_instrumentation__.Notify(467645)
	}
	__antithesis_instrumentation__.Notify(467639)
	resp, err := ss.ListNodesInternal(ctx, &serverpb.NodesRequest{})
	if err != nil {
		__antithesis_instrumentation__.Notify(467646)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(467647)
	}
	__antithesis_instrumentation__.Notify(467640)

	for _, node := range resp.Nodes {
		__antithesis_instrumentation__.Notify(467648)
		_ = dsp.CheckInstanceHealthAndVersion(ctx, planCtx, base.SQLInstanceID(node.Desc.NodeID))
	}
	__antithesis_instrumentation__.Notify(467641)
	nodes := make([]base.SQLInstanceID, 0, len(planCtx.NodeStatuses))
	for nodeID, status := range planCtx.NodeStatuses {
		__antithesis_instrumentation__.Notify(467649)
		if status == NodeOK {
			__antithesis_instrumentation__.Notify(467650)
			nodes = append(nodes, nodeID)
		} else {
			__antithesis_instrumentation__.Notify(467651)
		}
	}
	__antithesis_instrumentation__.Notify(467642)

	rand.Shuffle(len(nodes), func(i, j int) {
		__antithesis_instrumentation__.Notify(467652)
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	__antithesis_instrumentation__.Notify(467643)
	return planCtx, nodes, nil
}

func (dsp *DistSQLPlanner) setupAllNodesPlanningTenant(
	ctx context.Context, evalCtx *extendedEvalContext, execCfg *ExecutorConfig,
) (*PlanningCtx, []base.SQLInstanceID, error) {
	__antithesis_instrumentation__.Notify(467653)
	if dsp.sqlInstanceProvider == nil {
		__antithesis_instrumentation__.Notify(467658)
		return nil, nil, errors.New("sql instance provider not available in multi-tenant environment")
	} else {
		__antithesis_instrumentation__.Notify(467659)
	}
	__antithesis_instrumentation__.Notify(467654)
	planCtx := dsp.NewPlanningCtx(ctx, evalCtx, nil, nil,
		DistributionTypeAlways)
	pods, err := dsp.sqlInstanceProvider.GetAllInstances(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(467660)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(467661)
	}
	__antithesis_instrumentation__.Notify(467655)
	sqlInstanceIDs := make([]base.SQLInstanceID, len(pods))
	for i, pod := range pods {
		__antithesis_instrumentation__.Notify(467662)
		sqlInstanceIDs[i] = pod.InstanceID
	}
	__antithesis_instrumentation__.Notify(467656)

	rand.Shuffle(len(sqlInstanceIDs), func(i, j int) {
		__antithesis_instrumentation__.Notify(467663)
		sqlInstanceIDs[i], sqlInstanceIDs[j] = sqlInstanceIDs[j], sqlInstanceIDs[i]
	})
	__antithesis_instrumentation__.Notify(467657)
	return planCtx, sqlInstanceIDs, nil
}
