package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/streaming"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

func initReplicationBuiltins() {
	__antithesis_instrumentation__.Notify(602309)

	for k, v := range replicationBuiltins {
		__antithesis_instrumentation__.Notify(602310)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(602312)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(602313)
		}
		__antithesis_instrumentation__.Notify(602311)
		builtins[k] = v
	}
}

var replicationBuiltins = map[string]builtinDefinition{
	"crdb_internal.complete_stream_ingestion_job": makeBuiltin(
		tree.FunctionProperties{
			Category:         categoryStreamIngestion,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"job_id", types.Int},
				{"cutover_ts", types.TimestampTZ},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602314)
				mgr, err := streaming.GetReplicationStreamManager(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602317)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602318)
				}
				__antithesis_instrumentation__.Notify(602315)

				streamID := streaming.StreamID(*args[0].(*tree.DInt))
				cutoverTime := args[1].(*tree.DTimestampTZ).Time
				cutoverTimestamp := hlc.Timestamp{WallTime: cutoverTime.UnixNano()}
				err = mgr.CompleteStreamIngestion(evalCtx, evalCtx.Txn, streamID, cutoverTimestamp)
				if err != nil {
					__antithesis_instrumentation__.Notify(602319)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602320)
				}
				__antithesis_instrumentation__.Notify(602316)
				return tree.NewDInt(tree.DInt(streamID)), err
			},
			Info: "This function can be used to signal a running stream ingestion job to complete. " +
				"The job will eventually stop ingesting, revert to the specified timestamp and leave the " +
				"cluster in a consistent state. The specified timestamp can only be specified up to the" +
				" microsecond. " +
				"This function does not wait for the job to reach a terminal state, " +
				"but instead returns the job id as soon as it has signaled the job to complete. " +
				"This builtin can be used in conjunction with SHOW JOBS WHEN COMPLETE to ensure that the" +
				" job has left the cluster in a consistent state.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.start_replication_stream": makeBuiltin(
		tree.FunctionProperties{
			Category:         categoryStreamIngestion,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"tenant_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Int),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602321)
				mgr, err := streaming.GetReplicationStreamManager(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602325)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602326)
				}
				__antithesis_instrumentation__.Notify(602322)
				tenantID, err := mustBeDIntInTenantRange(args[0])
				if err != nil {
					__antithesis_instrumentation__.Notify(602327)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602328)
				}
				__antithesis_instrumentation__.Notify(602323)
				jobID, err := mgr.StartReplicationStream(evalCtx, evalCtx.Txn, uint64(tenantID))
				if err != nil {
					__antithesis_instrumentation__.Notify(602329)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602330)
				}
				__antithesis_instrumentation__.Notify(602324)
				return tree.NewDInt(tree.DInt(jobID)), err
			},
			Info: "This function can be used on the producer side to start a replication stream for " +
				"the specified tenant. The returned stream ID uniquely identifies created stream. " +
				"The caller must periodically invoke crdb_internal.heartbeat_stream() function to " +
				"notify that the replication is still ongoing.",
			Volatility: tree.VolatilityVolatile,
		},
	),

	"crdb_internal.replication_stream_progress": makeBuiltin(
		tree.FunctionProperties{
			Category:         categoryStreamIngestion,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"stream_id", types.Int},
				{"frontier_ts", types.String},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602331)
				mgr, err := streaming.GetReplicationStreamManager(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602336)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602337)
				}
				__antithesis_instrumentation__.Notify(602332)
				frontier, err := hlc.ParseTimestamp(string(tree.MustBeDString(args[1])))
				if err != nil {
					__antithesis_instrumentation__.Notify(602338)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602339)
				}
				__antithesis_instrumentation__.Notify(602333)
				streamID := streaming.StreamID(int(tree.MustBeDInt(args[0])))
				sps, err := mgr.UpdateReplicationStreamProgress(evalCtx, streamID, frontier, evalCtx.Txn)
				if err != nil {
					__antithesis_instrumentation__.Notify(602340)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602341)
				}
				__antithesis_instrumentation__.Notify(602334)
				rawStatus, err := protoutil.Marshal(&sps)
				if err != nil {
					__antithesis_instrumentation__.Notify(602342)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602343)
				}
				__antithesis_instrumentation__.Notify(602335)
				return tree.NewDBytes(tree.DBytes(rawStatus)), nil
			},
			Info: "This function can be used on the consumer side to heartbeat its replication progress to " +
				"a replication stream in the source cluster. The returns a StreamReplicationStatus message " +
				"that indicates stream status (RUNNING, PAUSED, or STOPPED).",
			Volatility: tree.VolatilityVolatile,
		},
	),
	"crdb_internal.stream_partition": makeBuiltin(
		tree.FunctionProperties{
			Category:         categoryStreamIngestion,
			DistsqlBlocklist: false,
			Class:            tree.GeneratorClass,
		},
		makeGeneratorOverload(
			tree.ArgTypes{
				{"stream_id", types.Int},
				{"partition_spec", types.Bytes},
			},
			types.MakeLabeledTuple(
				[]*types.T{types.Bytes},
				[]string{"stream_event"},
			),
			func(evalCtx *tree.EvalContext, args tree.Datums) (tree.ValueGenerator, error) {
				__antithesis_instrumentation__.Notify(602344)
				mgr, err := streaming.GetReplicationStreamManager(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602346)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602347)
				}
				__antithesis_instrumentation__.Notify(602345)
				return mgr.StreamPartition(
					evalCtx,
					streaming.StreamID(tree.MustBeDInt(args[0])),
					[]byte(tree.MustBeDBytes(args[1])),
				)
			},
			"Stream partition data",
			tree.VolatilityVolatile,
		),
	),

	"crdb_internal.replication_stream_spec": makeBuiltin(
		tree.FunctionProperties{
			Category:         categoryStreamIngestion,
			DistsqlBlocklist: true,
		},
		tree.Overload{
			Types: tree.ArgTypes{
				{"stream_id", types.Int},
			},
			ReturnType: tree.FixedReturnType(types.Bytes),
			Fn: func(evalCtx *tree.EvalContext, args tree.Datums) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(602348)
				mgr, err := streaming.GetReplicationStreamManager(evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602352)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602353)
				}
				__antithesis_instrumentation__.Notify(602349)

				streamID := int64(tree.MustBeDInt(args[0]))
				spec, err := mgr.GetReplicationStreamSpec(evalCtx, evalCtx.Txn, streaming.StreamID(streamID))
				if err != nil {
					__antithesis_instrumentation__.Notify(602354)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602355)
				}
				__antithesis_instrumentation__.Notify(602350)
				rawSpec, err := protoutil.Marshal(spec)
				if err != nil {
					__antithesis_instrumentation__.Notify(602356)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602357)
				}
				__antithesis_instrumentation__.Notify(602351)
				return tree.NewDBytes(tree.DBytes(rawSpec)), err
			},
			Info: "This function can be used on the consumer side to get a replication stream specification " +
				"for the specified stream starting from the specified 'start_from' timestamp. The consumer will " +
				"later call 'stream_partition' to a partition with the spec to start streaming.",
			Volatility: tree.VolatilityVolatile,
		},
	),
}
