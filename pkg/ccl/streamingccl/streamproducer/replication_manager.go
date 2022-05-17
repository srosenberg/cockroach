package streamproducer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type replicationStreamManagerImpl struct{}

func (r *replicationStreamManagerImpl) CompleteStreamIngestion(
	evalCtx *tree.EvalContext,
	txn *kv.Txn,
	streamID streaming.StreamID,
	cutoverTimestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(26974)
	return completeStreamIngestion(evalCtx, txn, streamID, cutoverTimestamp)
}

func (r *replicationStreamManagerImpl) StartReplicationStream(
	evalCtx *tree.EvalContext, txn *kv.Txn, tenantID uint64,
) (streaming.StreamID, error) {
	__antithesis_instrumentation__.Notify(26975)
	return startReplicationStreamJob(evalCtx, txn, tenantID)
}

func (r *replicationStreamManagerImpl) UpdateReplicationStreamProgress(
	evalCtx *tree.EvalContext, streamID streaming.StreamID, frontier hlc.Timestamp, txn *kv.Txn,
) (streampb.StreamReplicationStatus, error) {
	__antithesis_instrumentation__.Notify(26976)
	return heartbeatReplicationStream(evalCtx, streamID, frontier, txn)
}

func (r *replicationStreamManagerImpl) StreamPartition(
	evalCtx *tree.EvalContext, streamID streaming.StreamID, opaqueSpec []byte,
) (tree.ValueGenerator, error) {
	__antithesis_instrumentation__.Notify(26977)
	return streamPartition(evalCtx, streamID, opaqueSpec)
}

func (r *replicationStreamManagerImpl) GetReplicationStreamSpec(
	evalCtx *tree.EvalContext, txn *kv.Txn, streamID streaming.StreamID,
) (*streampb.ReplicationStreamSpec, error) {
	__antithesis_instrumentation__.Notify(26978)
	return getReplicationStreamSpec(evalCtx, txn, streamID)
}

func newReplicationStreamManagerWithPrivilegesCheck(
	evalCtx *tree.EvalContext,
) (streaming.ReplicationStreamManager, error) {
	__antithesis_instrumentation__.Notify(26979)
	isAdmin, err := evalCtx.SessionAccessor.HasAdminRole(evalCtx.Context)
	if err != nil {
		__antithesis_instrumentation__.Notify(26983)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(26984)
	}
	__antithesis_instrumentation__.Notify(26980)

	if !isAdmin {
		__antithesis_instrumentation__.Notify(26985)
		return nil,
			pgerror.New(pgcode.InsufficientPrivilege, "replication restricted to ADMIN role")
	} else {
		__antithesis_instrumentation__.Notify(26986)
	}
	__antithesis_instrumentation__.Notify(26981)

	execCfg := evalCtx.Planner.ExecutorConfig().(*sql.ExecutorConfig)
	enterpriseCheckErr := utilccl.CheckEnterpriseEnabled(
		execCfg.Settings, execCfg.LogicalClusterID(), execCfg.Organization(), "REPLICATION")
	if enterpriseCheckErr != nil {
		__antithesis_instrumentation__.Notify(26987)
		return nil, pgerror.Wrap(enterpriseCheckErr,
			pgcode.InsufficientPrivilege, "replication requires enterprise license")
	} else {
		__antithesis_instrumentation__.Notify(26988)
	}
	__antithesis_instrumentation__.Notify(26982)

	return &replicationStreamManagerImpl{}, nil
}

func init() {
	streaming.GetReplicationStreamManagerHook = newReplicationStreamManagerWithPrivilegesCheck
}
