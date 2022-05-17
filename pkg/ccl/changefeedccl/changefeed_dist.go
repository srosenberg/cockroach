package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeeddist"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func init() {
	rowexec.NewChangeAggregatorProcessor = newChangeAggregatorProcessor
	rowexec.NewChangeFrontierProcessor = newChangeFrontierProcessor
}

const (
	changeAggregatorProcName = `changeagg`
	changeFrontierProcName   = `changefntr`
)

func distChangefeedFlow(
	ctx context.Context,
	execCtx sql.JobExecContext,
	jobID jobspb.JobID,
	details jobspb.ChangefeedDetails,
	progress jobspb.Progress,
	resultsCh chan<- tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(15397)
	var err error
	details, err = validateDetails(details)
	if err != nil {
		__antithesis_instrumentation__.Notify(15402)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15403)
	}

	{
		__antithesis_instrumentation__.Notify(15404)
		h := progress.GetHighWater()
		noHighWater := (h == nil || func() bool {
			__antithesis_instrumentation__.Notify(15406)
			return h.IsEmpty() == true
		}() == true)

		initialScanType, err := initialScanTypeFromOpts(details.Opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(15407)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15408)
		}
		__antithesis_instrumentation__.Notify(15405)
		if noHighWater && func() bool {
			__antithesis_instrumentation__.Notify(15409)
			return initialScanType == changefeedbase.NoInitialScan == true
		}() == true {
			__antithesis_instrumentation__.Notify(15410)

			progress.Progress = &jobspb.Progress_HighWater{HighWater: &details.StatementTime}
		} else {
			__antithesis_instrumentation__.Notify(15411)
		}
	}
	__antithesis_instrumentation__.Notify(15398)

	execCfg := execCtx.ExecCfg()
	var initialHighWater hlc.Timestamp
	var trackedSpans []roachpb.Span
	{
		__antithesis_instrumentation__.Notify(15412)
		spansTS := details.StatementTime
		if h := progress.GetHighWater(); h != nil && func() bool {
			__antithesis_instrumentation__.Notify(15415)
			return !h.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(15416)
			initialHighWater = *h

			spansTS = initialHighWater
		} else {
			__antithesis_instrumentation__.Notify(15417)
		}
		__antithesis_instrumentation__.Notify(15413)

		isRestartAfterCheckpointOrNoInitialScan := progress.GetHighWater() != nil
		if isRestartAfterCheckpointOrNoInitialScan {
			__antithesis_instrumentation__.Notify(15418)
			spansTS = spansTS.Next()
		} else {
			__antithesis_instrumentation__.Notify(15419)
		}
		__antithesis_instrumentation__.Notify(15414)
		var err error
		trackedSpans, err = fetchSpansForTargets(ctx, execCfg, AllTargets(details), spansTS)
		if err != nil {
			__antithesis_instrumentation__.Notify(15420)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15421)
		}
	}
	__antithesis_instrumentation__.Notify(15399)

	var checkpoint jobspb.ChangefeedProgress_Checkpoint
	if cf := progress.GetChangefeed(); cf != nil && func() bool {
		__antithesis_instrumentation__.Notify(15422)
		return cf.Checkpoint != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(15423)
		checkpoint = *cf.Checkpoint
	} else {
		__antithesis_instrumentation__.Notify(15424)
	}
	__antithesis_instrumentation__.Notify(15400)

	var distflowKnobs changefeeddist.TestingKnobs
	if knobs, ok := execCfg.DistSQLSrv.TestingKnobs.Changefeed.(*TestingKnobs); ok && func() bool {
		__antithesis_instrumentation__.Notify(15425)
		return knobs != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(15426)
		distflowKnobs = knobs.DistflowKnobs
	} else {
		__antithesis_instrumentation__.Notify(15427)
	}
	__antithesis_instrumentation__.Notify(15401)

	return changefeeddist.StartDistChangefeed(
		ctx, execCtx, jobID, details, trackedSpans, initialHighWater, checkpoint, resultsCh, distflowKnobs)
}

func fetchSpansForTargets(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	targets []jobspb.ChangefeedTargetSpecification,
	ts hlc.Timestamp,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(15428)
	var spans []roachpb.Span
	fetchSpans := func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(15431)
		spans = nil
		if err := txn.SetFixedTimestamp(ctx, ts); err != nil {
			__antithesis_instrumentation__.Notify(15434)
			return err
		} else {
			__antithesis_instrumentation__.Notify(15435)
		}
		__antithesis_instrumentation__.Notify(15432)
		seen := make(map[descpb.ID]struct{}, len(targets))

		for _, table := range targets {
			__antithesis_instrumentation__.Notify(15436)
			if _, dup := seen[table.TableID]; dup {
				__antithesis_instrumentation__.Notify(15439)
				continue
			} else {
				__antithesis_instrumentation__.Notify(15440)
			}
			__antithesis_instrumentation__.Notify(15437)
			seen[table.TableID] = struct{}{}
			flags := tree.ObjectLookupFlagsWithRequired()
			flags.AvoidLeased = true
			tableDesc, err := descriptors.GetImmutableTableByID(ctx, txn, table.TableID, flags)
			if err != nil {
				__antithesis_instrumentation__.Notify(15441)
				return err
			} else {
				__antithesis_instrumentation__.Notify(15442)
			}
			__antithesis_instrumentation__.Notify(15438)
			spans = append(spans, tableDesc.PrimaryIndexSpan(execCfg.Codec))
		}
		__antithesis_instrumentation__.Notify(15433)
		return nil
	}
	__antithesis_instrumentation__.Notify(15429)
	if err := sql.DescsTxn(ctx, execCfg, fetchSpans); err != nil {
		__antithesis_instrumentation__.Notify(15443)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15444)
	}
	__antithesis_instrumentation__.Notify(15430)
	return spans, nil
}
