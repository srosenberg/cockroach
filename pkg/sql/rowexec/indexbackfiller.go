package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type indexBackfiller struct {
	backfill.IndexBackfiller

	adder kvserverbase.BulkAdder

	desc catalog.TableDescriptor

	spec execinfrapb.BackfillerSpec

	out execinfra.ProcOutputHelper

	flowCtx *execinfra.FlowCtx

	output execinfra.RowReceiver

	filter backfill.MutationFilter
}

var _ execinfra.Processor = &indexBackfiller{}

var backfillerBufferSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"schemachanger.backfiller.buffer_size", "the initial size of the BulkAdder buffer handling index backfills", 32<<20,
)

var backfillerMaxBufferSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"schemachanger.backfiller.max_buffer_size", "the maximum size of the BulkAdder buffer handling index backfills", 512<<20,
)

func newIndexBackfiller(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.BackfillerSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*indexBackfiller, error) {
	__antithesis_instrumentation__.Notify(572488)
	indexBackfillerMon := execinfra.NewMonitor(ctx, flowCtx.Cfg.BackfillerMonitor,
		"index-backfill-mon")
	ib := &indexBackfiller{
		desc:    flowCtx.TableDescriptor(&spec.Table),
		spec:    spec,
		flowCtx: flowCtx,
		output:  output,
		filter:  backfill.IndexMutationFilter,
	}

	if err := ib.IndexBackfiller.InitForDistributedUse(ctx, flowCtx, ib.desc,
		indexBackfillerMon); err != nil {
		__antithesis_instrumentation__.Notify(572490)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572491)
	}
	__antithesis_instrumentation__.Notify(572489)

	return ib, nil
}

func (ib *indexBackfiller) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(572492)

	return nil
}

func (ib *indexBackfiller) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(572493)
	return false
}

type indexEntryBatch struct {
	indexEntries         []rowenc.IndexEntry
	completedSpan        roachpb.Span
	memUsedBuildingBatch int64
}

func (ib *indexBackfiller) constructIndexEntries(
	ctx context.Context, indexEntriesCh chan indexEntryBatch,
) error {
	__antithesis_instrumentation__.Notify(572494)
	var memUsedBuildingBatch int64
	var err error
	var entries []rowenc.IndexEntry
	for i := range ib.spec.Spans {
		__antithesis_instrumentation__.Notify(572496)
		log.VEventf(ctx, 2, "index backfiller starting span %d of %d: %s",
			i+1, len(ib.spec.Spans), ib.spec.Spans[i])
		todo := ib.spec.Spans[i]
		for todo.Key != nil {
			__antithesis_instrumentation__.Notify(572497)
			startKey := todo.Key
			readAsOf := ib.spec.ReadAsOf
			if readAsOf.IsEmpty() {
				__antithesis_instrumentation__.Notify(572502)
				readAsOf = ib.spec.WriteAsOf
			} else {
				__antithesis_instrumentation__.Notify(572503)
			}
			__antithesis_instrumentation__.Notify(572498)
			todo.Key, entries, memUsedBuildingBatch, err = ib.buildIndexEntryBatch(ctx, todo,
				readAsOf)
			if err != nil {
				__antithesis_instrumentation__.Notify(572504)
				return err
			} else {
				__antithesis_instrumentation__.Notify(572505)
			}
			__antithesis_instrumentation__.Notify(572499)

			completedSpan := ib.spec.Spans[i]
			if todo.Key != nil {
				__antithesis_instrumentation__.Notify(572506)
				completedSpan.Key = startKey
				completedSpan.EndKey = todo.Key
			} else {
				__antithesis_instrumentation__.Notify(572507)
			}
			__antithesis_instrumentation__.Notify(572500)

			log.VEventf(ctx, 2, "index entries built for span %s", completedSpan)
			indexBatch := indexEntryBatch{completedSpan: completedSpan, indexEntries: entries,
				memUsedBuildingBatch: memUsedBuildingBatch}

			select {
			case indexEntriesCh <- indexBatch:
				__antithesis_instrumentation__.Notify(572508)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(572509)
				return ctx.Err()
			}
			__antithesis_instrumentation__.Notify(572501)

			knobs := ib.flowCtx.Cfg.TestingKnobs

			if knobs.SerializeIndexBackfillCreationAndIngestion != nil {
				__antithesis_instrumentation__.Notify(572510)
				<-knobs.SerializeIndexBackfillCreationAndIngestion
			} else {
				__antithesis_instrumentation__.Notify(572511)
			}
		}
	}
	__antithesis_instrumentation__.Notify(572495)

	return nil
}

func (ib *indexBackfiller) ingestIndexEntries(
	ctx context.Context,
	indexEntryCh <-chan indexEntryBatch,
	progCh chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
) error {
	__antithesis_instrumentation__.Notify(572512)
	ctx, span := tracing.ChildSpan(ctx, "ingestIndexEntries")
	defer span.Finish()

	minBufferSize := backfillerBufferSize.Get(&ib.flowCtx.Cfg.Settings.SV)
	maxBufferSize := func() int64 {
		__antithesis_instrumentation__.Notify(572521)
		return backfillerMaxBufferSize.Get(&ib.flowCtx.Cfg.Settings.SV)
	}
	__antithesis_instrumentation__.Notify(572513)
	opts := kvserverbase.BulkAdderOptions{
		Name:                     ib.desc.GetName() + " backfill",
		MinBufferSize:            minBufferSize,
		MaxBufferSize:            maxBufferSize,
		SkipDuplicates:           ib.ContainsInvertedIndex(),
		BatchTimestamp:           ib.spec.ReadAsOf,
		InitialSplitsIfUnordered: int(ib.spec.InitialSplits),
		WriteAtBatchTimestamp:    ib.spec.WriteAtBatchTimestamp,
	}
	adder, err := ib.flowCtx.Cfg.BulkAdder(ctx, ib.flowCtx.Cfg.DB, ib.spec.WriteAsOf, opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(572522)
		return err
	} else {
		__antithesis_instrumentation__.Notify(572523)
	}
	__antithesis_instrumentation__.Notify(572514)
	ib.adder = adder
	defer ib.adder.Close(ctx)

	mu := struct {
		syncutil.Mutex
		completedSpans []roachpb.Span
		addedSpans     []roachpb.Span
	}{}

	adder.SetOnFlush(func(_ roachpb.BulkOpSummary) {
		__antithesis_instrumentation__.Notify(572524)
		mu.Lock()
		defer mu.Unlock()
		mu.completedSpans = append(mu.completedSpans, mu.addedSpans...)
		mu.addedSpans = nil
	})
	__antithesis_instrumentation__.Notify(572515)

	pushProgress := func() {
		__antithesis_instrumentation__.Notify(572525)
		mu.Lock()
		var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
		prog.CompletedSpans = append(prog.CompletedSpans, mu.completedSpans...)
		mu.completedSpans = nil
		mu.Unlock()

		progCh <- prog
	}
	__antithesis_instrumentation__.Notify(572516)

	stopProgress := make(chan struct{})
	g := ctxgroup.WithContext(ctx)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(572526)
		tick := time.NewTicker(ib.getProgressReportInterval())
		defer tick.Stop()
		done := ctx.Done()
		for {
			__antithesis_instrumentation__.Notify(572527)
			select {
			case <-done:
				__antithesis_instrumentation__.Notify(572528)
				return ctx.Err()
			case <-stopProgress:
				__antithesis_instrumentation__.Notify(572529)
				return nil
			case <-tick.C:
				__antithesis_instrumentation__.Notify(572530)
				pushProgress()
			}
		}
	})
	__antithesis_instrumentation__.Notify(572517)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(572531)
		defer close(stopProgress)

		for indexBatch := range indexEntryCh {
			__antithesis_instrumentation__.Notify(572533)
			for _, indexEntry := range indexBatch.indexEntries {
				__antithesis_instrumentation__.Notify(572537)
				if err := ib.adder.Add(ctx, indexEntry.Key, indexEntry.Value.RawBytes); err != nil {
					__antithesis_instrumentation__.Notify(572538)
					return ib.wrapDupError(ctx, err)
				} else {
					__antithesis_instrumentation__.Notify(572539)
				}
			}
			__antithesis_instrumentation__.Notify(572534)

			mu.Lock()
			mu.addedSpans = append(mu.addedSpans, indexBatch.completedSpan)
			mu.Unlock()

			indexBatch.indexEntries = nil
			ib.ShrinkBoundAccount(ctx, indexBatch.memUsedBuildingBatch)

			knobs := &ib.flowCtx.Cfg.TestingKnobs
			if knobs.BulkAdderFlushesEveryBatch {
				__antithesis_instrumentation__.Notify(572540)
				if err := ib.adder.Flush(ctx); err != nil {
					__antithesis_instrumentation__.Notify(572542)
					return ib.wrapDupError(ctx, err)
				} else {
					__antithesis_instrumentation__.Notify(572543)
				}
				__antithesis_instrumentation__.Notify(572541)
				pushProgress()
			} else {
				__antithesis_instrumentation__.Notify(572544)
			}
			__antithesis_instrumentation__.Notify(572535)

			if knobs.RunAfterBackfillChunk != nil {
				__antithesis_instrumentation__.Notify(572545)
				knobs.RunAfterBackfillChunk()
			} else {
				__antithesis_instrumentation__.Notify(572546)
			}
			__antithesis_instrumentation__.Notify(572536)

			if knobs.SerializeIndexBackfillCreationAndIngestion != nil {
				__antithesis_instrumentation__.Notify(572547)
				knobs.SerializeIndexBackfillCreationAndIngestion <- struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(572548)
			}
		}
		__antithesis_instrumentation__.Notify(572532)
		return nil
	})
	__antithesis_instrumentation__.Notify(572518)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(572549)
		return err
	} else {
		__antithesis_instrumentation__.Notify(572550)
	}
	__antithesis_instrumentation__.Notify(572519)

	if err := ib.adder.Flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(572551)
		return ib.wrapDupError(ctx, err)
	} else {
		__antithesis_instrumentation__.Notify(572552)
	}
	__antithesis_instrumentation__.Notify(572520)

	pushProgress()

	return nil
}

func (ib *indexBackfiller) runBackfill(
	ctx context.Context, progCh chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
) error {
	__antithesis_instrumentation__.Notify(572553)

	indexEntriesCh := make(chan indexEntryBatch, 10)

	group := ctxgroup.WithContext(ctx)

	group.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(572557)
		defer close(indexEntriesCh)
		ctx, span := tracing.ChildSpan(ctx, "buildIndexEntries")
		defer span.Finish()
		err := ib.constructIndexEntries(ctx, indexEntriesCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(572559)
			return errors.Wrap(err, "failed to construct index entries during backfill")
		} else {
			__antithesis_instrumentation__.Notify(572560)
		}
		__antithesis_instrumentation__.Notify(572558)
		return nil
	})
	__antithesis_instrumentation__.Notify(572554)

	group.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(572561)
		err := ib.ingestIndexEntries(ctx, indexEntriesCh, progCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(572563)
			return errors.Wrap(err, "failed to ingest index entries during backfill")
		} else {
			__antithesis_instrumentation__.Notify(572564)
		}
		__antithesis_instrumentation__.Notify(572562)
		return nil
	})
	__antithesis_instrumentation__.Notify(572555)

	if err := group.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(572565)
		return err
	} else {
		__antithesis_instrumentation__.Notify(572566)
	}
	__antithesis_instrumentation__.Notify(572556)

	return nil
}

func (ib *indexBackfiller) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572567)
	opName := "indexBackfillerProcessor"
	ctx = logtags.AddTag(ctx, "job", ib.spec.JobID)
	ctx = logtags.AddTag(ctx, opName, int(ib.spec.Table.ID))
	ctx, span := execinfra.ProcessorSpan(ctx, opName)
	defer span.Finish()
	defer ib.output.ProducerDone()
	defer execinfra.SendTraceData(ctx, ib.output)
	defer ib.Close(ctx)

	progCh := make(chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress)

	semaCtx := tree.MakeSemaContext()
	if err := ib.out.Init(&execinfrapb.PostProcessSpec{}, nil, &semaCtx, ib.flowCtx.NewEvalCtx()); err != nil {
		__antithesis_instrumentation__.Notify(572571)
		ib.output.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
		return
	} else {
		__antithesis_instrumentation__.Notify(572572)
	}
	__antithesis_instrumentation__.Notify(572568)

	var err error

	go func() {
		__antithesis_instrumentation__.Notify(572573)
		defer close(progCh)
		err = ib.runBackfill(ctx, progCh)
	}()
	__antithesis_instrumentation__.Notify(572569)

	for prog := range progCh {
		__antithesis_instrumentation__.Notify(572574)

		p := prog
		if p.CompletedSpans != nil {
			__antithesis_instrumentation__.Notify(572576)
			log.VEventf(ctx, 2, "sending coordinator completed spans: %+v", p.CompletedSpans)
		} else {
			__antithesis_instrumentation__.Notify(572577)
		}
		__antithesis_instrumentation__.Notify(572575)
		ib.output.Push(nil, &execinfrapb.ProducerMetadata{BulkProcessorProgress: &p})
	}
	__antithesis_instrumentation__.Notify(572570)

	if err != nil {
		__antithesis_instrumentation__.Notify(572578)
		ib.output.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
		return
	} else {
		__antithesis_instrumentation__.Notify(572579)
	}
}

func (ib *indexBackfiller) wrapDupError(ctx context.Context, orig error) error {
	__antithesis_instrumentation__.Notify(572580)
	if orig == nil {
		__antithesis_instrumentation__.Notify(572584)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572585)
	}
	__antithesis_instrumentation__.Notify(572581)
	var typed *kvserverbase.DuplicateKeyError
	if !errors.As(orig, &typed) {
		__antithesis_instrumentation__.Notify(572586)
		return orig
	} else {
		__antithesis_instrumentation__.Notify(572587)
	}
	__antithesis_instrumentation__.Notify(572582)

	desc, err := ib.desc.MakeFirstMutationPublic(catalog.IncludeConstraints)
	if err != nil {
		__antithesis_instrumentation__.Notify(572588)
		return err
	} else {
		__antithesis_instrumentation__.Notify(572589)
	}
	__antithesis_instrumentation__.Notify(572583)
	v := &roachpb.Value{RawBytes: typed.Value}
	return row.NewUniquenessConstraintViolationError(ctx, desc, typed.Key, v)
}

const indexBackfillProgressReportInterval = 10 * time.Second

func (ib *indexBackfiller) getProgressReportInterval() time.Duration {
	__antithesis_instrumentation__.Notify(572590)
	knobs := &ib.flowCtx.Cfg.TestingKnobs
	if knobs.IndexBackfillProgressReportInterval > 0 {
		__antithesis_instrumentation__.Notify(572592)
		return knobs.IndexBackfillProgressReportInterval
	} else {
		__antithesis_instrumentation__.Notify(572593)
	}
	__antithesis_instrumentation__.Notify(572591)

	return indexBackfillProgressReportInterval
}

func (ib *indexBackfiller) buildIndexEntryBatch(
	tctx context.Context, sp roachpb.Span, readAsOf hlc.Timestamp,
) (roachpb.Key, []rowenc.IndexEntry, int64, error) {
	__antithesis_instrumentation__.Notify(572594)
	knobs := &ib.flowCtx.Cfg.TestingKnobs
	var memUsedBuildingBatch int64
	if knobs.RunBeforeBackfillChunk != nil {
		__antithesis_instrumentation__.Notify(572597)
		if err := knobs.RunBeforeBackfillChunk(sp); err != nil {
			__antithesis_instrumentation__.Notify(572598)
			return nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(572599)
		}
	} else {
		__antithesis_instrumentation__.Notify(572600)
	}
	__antithesis_instrumentation__.Notify(572595)
	var key roachpb.Key

	ctx, traceSpan := tracing.ChildSpan(tctx, "indexBatch")
	defer traceSpan.Finish()
	start := timeutil.Now()
	var entries []rowenc.IndexEntry
	if err := ib.flowCtx.Cfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(572601)
		if err := txn.SetFixedTimestamp(ctx, readAsOf); err != nil {
			__antithesis_instrumentation__.Notify(572603)
			return err
		} else {
			__antithesis_instrumentation__.Notify(572604)
		}
		__antithesis_instrumentation__.Notify(572602)

		var err error
		entries, key, memUsedBuildingBatch, err = ib.BuildIndexEntriesChunk(ctx, txn, ib.desc, sp,
			ib.spec.ChunkSize, false)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(572605)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(572606)
	}
	__antithesis_instrumentation__.Notify(572596)
	prepTime := timeutil.Since(start)
	log.VEventf(ctx, 3, "index backfill stats: entries %d, prepare %+v",
		len(entries), prepTime)

	return key, entries, memUsedBuildingBatch, nil
}
