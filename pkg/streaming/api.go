package streaming

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/ccl/streamingccl/streampb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type StreamID int64

func (j StreamID) SafeValue() { __antithesis_instrumentation__.Notify(643970) }

const InvalidStreamID StreamID = 0

var GetReplicationStreamManagerHook func(evalCtx *tree.EvalContext) (ReplicationStreamManager, error)

type ReplicationStreamManager interface {
	CompleteStreamIngestion(
		evalCtx *tree.EvalContext,
		txn *kv.Txn,
		streamID StreamID,
		cutoverTimestamp hlc.Timestamp,
	) error

	StartReplicationStream(
		evalCtx *tree.EvalContext,
		txn *kv.Txn,
		tenantID uint64,
	) (StreamID, error)

	UpdateReplicationStreamProgress(
		evalCtx *tree.EvalContext,
		streamID StreamID,
		frontier hlc.Timestamp,
		txn *kv.Txn) (streampb.StreamReplicationStatus, error)

	StreamPartition(
		evalCtx *tree.EvalContext,
		streamID StreamID,
		opaqueSpec []byte,
	) (tree.ValueGenerator, error)

	GetReplicationStreamSpec(
		evalCtx *tree.EvalContext,
		txn *kv.Txn,
		streamID StreamID,
	) (*streampb.ReplicationStreamSpec, error)
}

func GetReplicationStreamManager(evalCtx *tree.EvalContext) (ReplicationStreamManager, error) {
	__antithesis_instrumentation__.Notify(643971)
	if GetReplicationStreamManagerHook == nil {
		__antithesis_instrumentation__.Notify(643973)
		return nil, errors.New("replication streaming requires a CCL binary")
	} else {
		__antithesis_instrumentation__.Notify(643974)
	}
	__antithesis_instrumentation__.Notify(643972)
	return GetReplicationStreamManagerHook(evalCtx)
}
