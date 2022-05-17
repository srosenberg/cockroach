package scexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec/scmutationexec"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func executeBackfillOps(ctx context.Context, deps Dependencies, execute []scop.Op) (err error) {
	__antithesis_instrumentation__.Notify(581317)
	backfillsToExecute := extractBackfillsFromOps(execute)
	tables, err := getTableDescriptorsForBackfills(ctx, deps.Catalog(), backfillsToExecute)
	if err != nil {
		__antithesis_instrumentation__.Notify(581320)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581321)
	}
	__antithesis_instrumentation__.Notify(581318)
	tracker := deps.BackfillProgressTracker()
	progresses, err := loadProgressesAndMaybePerformInitialScan(
		ctx, deps, backfillsToExecute, tracker, tables,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(581322)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581323)
	}
	__antithesis_instrumentation__.Notify(581319)
	return runBackfills(ctx, deps, tracker, progresses, tables)
}

func getTableDescriptorsForBackfills(
	ctx context.Context, cat Catalog, backfills []Backfill,
) (_ map[descpb.ID]catalog.TableDescriptor, err error) {
	__antithesis_instrumentation__.Notify(581324)
	var descIDs catalog.DescriptorIDSet
	for _, bf := range backfills {
		__antithesis_instrumentation__.Notify(581327)
		descIDs.Add(bf.TableID)
	}
	__antithesis_instrumentation__.Notify(581325)
	tables := make(map[descpb.ID]catalog.TableDescriptor, descIDs.Len())
	for _, id := range descIDs.Ordered() {
		__antithesis_instrumentation__.Notify(581328)
		desc, err := cat.MustReadImmutableDescriptors(ctx, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(581331)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581332)
		}
		__antithesis_instrumentation__.Notify(581329)
		tbl, ok := desc[0].(catalog.TableDescriptor)
		if !ok {
			__antithesis_instrumentation__.Notify(581333)
			return nil, errors.AssertionFailedf("descriptor %d is not a table", id)
		} else {
			__antithesis_instrumentation__.Notify(581334)
		}
		__antithesis_instrumentation__.Notify(581330)
		tables[id] = tbl
	}
	__antithesis_instrumentation__.Notify(581326)
	return tables, nil
}

func extractBackfillsFromOps(execute []scop.Op) []Backfill {
	__antithesis_instrumentation__.Notify(581335)
	var backfillsToExecute []Backfill
	for _, op := range execute {
		__antithesis_instrumentation__.Notify(581337)
		switch op := op.(type) {
		case *scop.BackfillIndex:
			__antithesis_instrumentation__.Notify(581338)
			backfillsToExecute = append(backfillsToExecute, Backfill{
				TableID:       op.TableID,
				SourceIndexID: op.SourceIndexID,
				DestIndexIDs:  []descpb.IndexID{op.IndexID},
			})
		default:
			__antithesis_instrumentation__.Notify(581339)
			panic("unimplemented")
		}
	}
	__antithesis_instrumentation__.Notify(581336)
	return mergeBackfillsFromSameSource(backfillsToExecute)
}

func mergeBackfillsFromSameSource(toExecute []Backfill) []Backfill {
	__antithesis_instrumentation__.Notify(581340)
	sort.Slice(toExecute, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(581345)
		if toExecute[i].TableID == toExecute[j].TableID {
			__antithesis_instrumentation__.Notify(581347)
			return toExecute[i].SourceIndexID < toExecute[j].SourceIndexID
		} else {
			__antithesis_instrumentation__.Notify(581348)
		}
		__antithesis_instrumentation__.Notify(581346)
		return toExecute[i].TableID < toExecute[j].TableID
	})
	__antithesis_instrumentation__.Notify(581341)
	truncated := toExecute[:0]
	sameSource := func(a, b Backfill) bool {
		__antithesis_instrumentation__.Notify(581349)
		return a.TableID == b.TableID && func() bool {
			__antithesis_instrumentation__.Notify(581350)
			return a.SourceIndexID == b.SourceIndexID == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(581342)
	for _, bf := range toExecute {
		__antithesis_instrumentation__.Notify(581351)
		if len(truncated) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(581352)
			return !sameSource(truncated[len(truncated)-1], bf) == true
		}() == true {
			__antithesis_instrumentation__.Notify(581353)
			truncated = append(truncated, bf)
		} else {
			__antithesis_instrumentation__.Notify(581354)
			ord := len(truncated) - 1
			curIDs := truncated[ord].DestIndexIDs
			curIDs = curIDs[:len(curIDs):len(curIDs)]
			truncated[ord].DestIndexIDs = append(curIDs, bf.DestIndexIDs...)
		}
	}
	__antithesis_instrumentation__.Notify(581343)
	toExecute = truncated

	for i := range toExecute {
		__antithesis_instrumentation__.Notify(581355)
		dstIDs := toExecute[i].DestIndexIDs
		sort.Slice(dstIDs, func(i, j int) bool { __antithesis_instrumentation__.Notify(581356); return dstIDs[i] < dstIDs[j] })
	}
	__antithesis_instrumentation__.Notify(581344)
	return toExecute
}

func loadProgressesAndMaybePerformInitialScan(
	ctx context.Context,
	deps Dependencies,
	backfillsToExecute []Backfill,
	tracker BackfillTracker,
	tables map[descpb.ID]catalog.TableDescriptor,
) ([]BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(581357)
	progresses, err := loadProgresses(ctx, backfillsToExecute, tracker)
	if err != nil {
		__antithesis_instrumentation__.Notify(581359)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(581360)
	}
	{
		__antithesis_instrumentation__.Notify(581361)
		didScan, err := maybeScanDestinationIndexes(ctx, deps, progresses, tables, tracker)
		if err != nil {
			__antithesis_instrumentation__.Notify(581363)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581364)
		}
		__antithesis_instrumentation__.Notify(581362)
		if didScan {
			__antithesis_instrumentation__.Notify(581365)
			if err := tracker.FlushCheckpoint(ctx); err != nil {
				__antithesis_instrumentation__.Notify(581366)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(581367)
			}
		} else {
			__antithesis_instrumentation__.Notify(581368)
		}
	}
	__antithesis_instrumentation__.Notify(581358)
	return progresses, nil
}

func maybeScanDestinationIndexes(
	ctx context.Context,
	deps Dependencies,
	progresses []BackfillProgress,
	tables map[descpb.ID]catalog.TableDescriptor,
	tracker BackfillProgressWriter,
) (didScan bool, _ error) {
	__antithesis_instrumentation__.Notify(581369)
	for i := range progresses {
		__antithesis_instrumentation__.Notify(581373)
		if didScan = didScan || func() bool {
			__antithesis_instrumentation__.Notify(581374)
			return progresses[i].MinimumWriteTimestamp.IsEmpty() == true
		}() == true; didScan {
			__antithesis_instrumentation__.Notify(581375)
			break
		} else {
			__antithesis_instrumentation__.Notify(581376)
		}
	}
	__antithesis_instrumentation__.Notify(581370)
	if !didScan {
		__antithesis_instrumentation__.Notify(581377)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(581378)
	}
	__antithesis_instrumentation__.Notify(581371)
	const op = "scan destination indexes"
	if err := forEachProgressConcurrently(ctx, op, progresses, func(
		ctx context.Context, p *BackfillProgress,
	) error {
		__antithesis_instrumentation__.Notify(581379)
		if !p.MinimumWriteTimestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(581382)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(581383)
		}
		__antithesis_instrumentation__.Notify(581380)
		updated, err := deps.IndexBackfiller().MaybePrepareDestIndexesForBackfill(
			ctx, *p, tables[p.TableID],
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(581384)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581385)
		}
		__antithesis_instrumentation__.Notify(581381)
		*p = updated
		return tracker.SetBackfillProgress(ctx, updated)
	}); err != nil {
		__antithesis_instrumentation__.Notify(581386)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(581387)
	}
	__antithesis_instrumentation__.Notify(581372)
	return true, nil
}

func forEachProgressConcurrently(
	ctx context.Context,
	op redact.SafeString,
	progresses []BackfillProgress,
	f func(context.Context, *BackfillProgress) error,
) error {
	__antithesis_instrumentation__.Notify(581388)
	g := ctxgroup.WithContext(ctx)
	run := func(i int) {
		__antithesis_instrumentation__.Notify(581391)
		g.GoCtx(func(ctx context.Context) (err error) {
			__antithesis_instrumentation__.Notify(581392)
			defer func() {
				__antithesis_instrumentation__.Notify(581394)
				switch r := recover().(type) {
				case nil:
					__antithesis_instrumentation__.Notify(581395)
					return
				case error:
					__antithesis_instrumentation__.Notify(581396)
					err = errors.Wrapf(r, "failed to %s", op)
				default:
					__antithesis_instrumentation__.Notify(581397)
					err = errors.AssertionFailedf("failed to %s: %v", op, r)
				}
			}()
			__antithesis_instrumentation__.Notify(581393)
			return f(ctx, &progresses[i])
		})
	}
	__antithesis_instrumentation__.Notify(581389)
	for i := range progresses {
		__antithesis_instrumentation__.Notify(581398)
		run(i)
	}
	__antithesis_instrumentation__.Notify(581390)
	return g.Wait()
}

func loadProgresses(
	ctx context.Context, backfillsToExecute []Backfill, tracker BackfillProgressReader,
) ([]BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(581399)
	var progresses []BackfillProgress
	for _, bf := range backfillsToExecute {
		__antithesis_instrumentation__.Notify(581401)
		progress, err := tracker.GetBackfillProgress(ctx, bf)
		if err != nil {
			__antithesis_instrumentation__.Notify(581403)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581404)
		}
		__antithesis_instrumentation__.Notify(581402)
		progresses = append(progresses, progress)
	}
	__antithesis_instrumentation__.Notify(581400)
	return progresses, nil
}

func runBackfills(
	ctx context.Context,
	deps Dependencies,
	tracker BackfillTracker,
	progresses []BackfillProgress,
	tables map[descpb.ID]catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(581405)

	stop := deps.PeriodicProgressFlusher().StartPeriodicUpdates(ctx, tracker)
	defer func() { __antithesis_instrumentation__.Notify(581410); _ = stop() }()
	__antithesis_instrumentation__.Notify(581406)
	bf := deps.IndexBackfiller()
	const op = "run backfills"
	if err := forEachProgressConcurrently(ctx, op, progresses, func(
		ctx context.Context, p *BackfillProgress,
	) error {
		__antithesis_instrumentation__.Notify(581411)
		return runBackfill(
			ctx, deps.IndexSpanSplitter(), bf, *p, tracker, tables[p.TableID],
		)
	}); err != nil {
		__antithesis_instrumentation__.Notify(581412)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581413)
	}
	__antithesis_instrumentation__.Notify(581407)
	if err := stop(); err != nil {
		__antithesis_instrumentation__.Notify(581414)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581415)
	}
	__antithesis_instrumentation__.Notify(581408)
	if err := tracker.FlushFractionCompleted(ctx); err != nil {
		__antithesis_instrumentation__.Notify(581416)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581417)
	}
	__antithesis_instrumentation__.Notify(581409)
	return tracker.FlushCheckpoint(ctx)
}

func runBackfill(
	ctx context.Context,
	splitter IndexSpanSplitter,
	backfiller Backfiller,
	progress BackfillProgress,
	tracker BackfillProgressWriter,
	table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(581418)

	for _, destIndexID := range progress.DestIndexIDs {
		__antithesis_instrumentation__.Notify(581420)
		mut, err := scmutationexec.FindMutation(table,
			scmutationexec.MakeIndexIDMutationSelector(destIndexID))
		if err != nil {
			__antithesis_instrumentation__.Notify(581422)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581423)
		}
		__antithesis_instrumentation__.Notify(581421)

		idxToBackfill := mut.AsIndex()
		if err := splitter.MaybeSplitIndexSpans(ctx, table, idxToBackfill); err != nil {
			__antithesis_instrumentation__.Notify(581424)
			return err
		} else {
			__antithesis_instrumentation__.Notify(581425)
		}
	}
	__antithesis_instrumentation__.Notify(581419)

	return backfiller.BackfillIndex(ctx, progress, tracker, table)
}
