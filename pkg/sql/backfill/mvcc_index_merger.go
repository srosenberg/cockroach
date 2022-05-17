package backfill

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

var indexBackfillMergeBatchSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.index_backfill.merge_batch_size",
	"the number of rows we merge between temporary and adding indexes in a single batch",
	1000,
	settings.NonNegativeInt,
)

var indexBackfillMergeBatchBytes = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.index_backfill.merge_batch_bytes",
	"the max number of bytes we merge between temporary and adding indexes in a single batch",
	16<<20,
	settings.NonNegativeInt,
)

var indexBackfillMergeNumWorkers = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.index_backfill.merge_num_workers",
	"the number of parallel merges per node in the cluster",
	4,
	settings.PositiveInt,
)

type IndexBackfillMerger struct {
	spec execinfrapb.IndexBackfillMergerSpec

	desc catalog.TableDescriptor

	out execinfra.ProcOutputHelper

	flowCtx *execinfra.FlowCtx

	evalCtx *tree.EvalContext

	output execinfra.RowReceiver

	mon            *mon.BytesMonitor
	muBoundAccount muBoundAccount
}

func (ibm *IndexBackfillMerger) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(245884)
	return nil
}

func (ibm *IndexBackfillMerger) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(245885)
	return false
}

const indexBackfillMergeProgressReportInterval = 10 * time.Second

func (ibm *IndexBackfillMerger) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(245886)
	opName := "IndexBackfillMerger"
	ctx = logtags.AddTag(ctx, opName, int(ibm.spec.Table.ID))
	ctx, span := execinfra.ProcessorSpan(ctx, opName)
	defer span.Finish()
	defer ibm.output.ProducerDone()
	defer execinfra.SendTraceData(ctx, ibm.output)

	mu := struct {
		syncutil.Mutex
		completedSpans   []roachpb.Span
		completedSpanIdx []int32
	}{}

	storeChunkProgress := func(chunk mergeChunk) {
		__antithesis_instrumentation__.Notify(245895)
		mu.Lock()
		defer mu.Unlock()
		mu.completedSpans = append(mu.completedSpans, chunk.completedSpan)
		mu.completedSpanIdx = append(mu.completedSpanIdx, chunk.spanIdx)
	}
	__antithesis_instrumentation__.Notify(245887)

	getStoredProgressForPush := func() execinfrapb.RemoteProducerMetadata_BulkProcessorProgress {
		__antithesis_instrumentation__.Notify(245896)
		var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
		mu.Lock()
		defer mu.Unlock()

		prog.CompletedSpans = append(prog.CompletedSpans, mu.completedSpans...)
		prog.CompletedSpanIdx = append(prog.CompletedSpanIdx, mu.completedSpanIdx...)
		mu.completedSpans, mu.completedSpanIdx = nil, nil
		return prog
	}
	__antithesis_instrumentation__.Notify(245888)

	pushProgress := func() {
		__antithesis_instrumentation__.Notify(245897)
		p := getStoredProgressForPush()
		if p.CompletedSpans != nil {
			__antithesis_instrumentation__.Notify(245899)
			log.VEventf(ctx, 2, "sending coordinator completed spans: %+v", p.CompletedSpans)
		} else {
			__antithesis_instrumentation__.Notify(245900)
		}
		__antithesis_instrumentation__.Notify(245898)
		ibm.output.Push(nil, &execinfrapb.ProducerMetadata{BulkProcessorProgress: &p})
	}
	__antithesis_instrumentation__.Notify(245889)

	semaCtx := tree.MakeSemaContext()
	if err := ibm.out.Init(&execinfrapb.PostProcessSpec{}, nil, &semaCtx, ibm.flowCtx.NewEvalCtx()); err != nil {
		__antithesis_instrumentation__.Notify(245901)
		ibm.output.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
		return
	} else {
		__antithesis_instrumentation__.Notify(245902)
	}
	__antithesis_instrumentation__.Notify(245890)

	numWorkers := int(indexBackfillMergeNumWorkers.Get(&ibm.evalCtx.Settings.SV))
	mergeCh := make(chan mergeChunk)
	mergeTimestamp := ibm.spec.MergeTimestamp

	g := ctxgroup.WithContext(ctx)
	runWorker := func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(245903)
		for mergeChunk := range mergeCh {
			__antithesis_instrumentation__.Notify(245905)
			err := ibm.merge(ctx, ibm.evalCtx.Codec, ibm.desc, ibm.spec.TemporaryIndexes[mergeChunk.spanIdx],
				ibm.spec.AddedIndexes[mergeChunk.spanIdx], mergeChunk.keys, mergeChunk.completedSpan)
			if err != nil {
				__antithesis_instrumentation__.Notify(245907)
				return err
			} else {
				__antithesis_instrumentation__.Notify(245908)
			}
			__antithesis_instrumentation__.Notify(245906)

			storeChunkProgress(mergeChunk)

			mergeChunk.keys = nil
			ibm.shrinkBoundAccount(ctx, mergeChunk.memUsed)

			if knobs, ok := ibm.flowCtx.Cfg.TestingKnobs.IndexBackfillMergerTestingKnobs.(*IndexBackfillMergerTestingKnobs); ok {
				__antithesis_instrumentation__.Notify(245909)
				if knobs != nil {
					__antithesis_instrumentation__.Notify(245910)
					if knobs.PushesProgressEveryChunk {
						__antithesis_instrumentation__.Notify(245912)
						pushProgress()
					} else {
						__antithesis_instrumentation__.Notify(245913)
					}
					__antithesis_instrumentation__.Notify(245911)

					if knobs.RunAfterMergeChunk != nil {
						__antithesis_instrumentation__.Notify(245914)
						knobs.RunAfterMergeChunk()
					} else {
						__antithesis_instrumentation__.Notify(245915)
					}
				} else {
					__antithesis_instrumentation__.Notify(245916)
				}
			} else {
				__antithesis_instrumentation__.Notify(245917)
			}
		}
		__antithesis_instrumentation__.Notify(245904)
		return nil
	}
	__antithesis_instrumentation__.Notify(245891)

	for worker := 0; worker < numWorkers; worker++ {
		__antithesis_instrumentation__.Notify(245918)
		g.GoCtx(runWorker)
	}
	__antithesis_instrumentation__.Notify(245892)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(245919)
		defer close(mergeCh)
		for i := range ibm.spec.Spans {
			__antithesis_instrumentation__.Notify(245921)
			sp := ibm.spec.Spans[i]
			idx := ibm.spec.SpanIdx[i]

			key := sp.Key
			for key != nil {
				__antithesis_instrumentation__.Notify(245922)
				chunk, nextKey, err := ibm.scan(ctx, idx, key, sp.EndKey, mergeTimestamp)
				if err != nil {
					__antithesis_instrumentation__.Notify(245925)
					return err
				} else {
					__antithesis_instrumentation__.Notify(245926)
				}
				__antithesis_instrumentation__.Notify(245923)
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(245927)
					return ctx.Err()
				case mergeCh <- chunk:
					__antithesis_instrumentation__.Notify(245928)
				}
				__antithesis_instrumentation__.Notify(245924)
				key = nextKey

				if knobs, ok := ibm.flowCtx.Cfg.TestingKnobs.IndexBackfillMergerTestingKnobs.(*IndexBackfillMergerTestingKnobs); ok {
					__antithesis_instrumentation__.Notify(245929)
					if knobs != nil && func() bool {
						__antithesis_instrumentation__.Notify(245930)
						return knobs.RunAfterScanChunk != nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(245931)
						knobs.RunAfterScanChunk()
					} else {
						__antithesis_instrumentation__.Notify(245932)
					}
				} else {
					__antithesis_instrumentation__.Notify(245933)
				}
			}
		}
		__antithesis_instrumentation__.Notify(245920)
		return nil
	})
	__antithesis_instrumentation__.Notify(245893)

	workersDoneCh := make(chan error)
	go func() { __antithesis_instrumentation__.Notify(245934); workersDoneCh <- g.Wait() }()
	__antithesis_instrumentation__.Notify(245894)

	tick := time.NewTicker(indexBackfillMergeProgressReportInterval)
	defer tick.Stop()
	var err error
	for {
		__antithesis_instrumentation__.Notify(245935)
		select {
		case <-tick.C:
			__antithesis_instrumentation__.Notify(245936)
			pushProgress()
		case err = <-workersDoneCh:
			__antithesis_instrumentation__.Notify(245937)
			if err != nil {
				__antithesis_instrumentation__.Notify(245939)
				ibm.output.Push(nil, &execinfrapb.ProducerMetadata{Err: err})
			} else {
				__antithesis_instrumentation__.Notify(245940)
			}
			__antithesis_instrumentation__.Notify(245938)
			return
		}
	}
}

var _ execinfra.Processor = &IndexBackfillMerger{}

type mergeChunk struct {
	completedSpan roachpb.Span
	keys          []roachpb.Key
	spanIdx       int32
	memUsed       int64
}

func (ibm *IndexBackfillMerger) scan(
	ctx context.Context,
	spanIdx int32,
	startKey roachpb.Key,
	endKey roachpb.Key,
	readAsOf hlc.Timestamp,
) (mergeChunk, roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(245941)
	if knobs, ok := ibm.flowCtx.Cfg.TestingKnobs.IndexBackfillMergerTestingKnobs.(*IndexBackfillMergerTestingKnobs); ok {
		__antithesis_instrumentation__.Notify(245945)
		if knobs != nil && func() bool {
			__antithesis_instrumentation__.Notify(245946)
			return knobs.RunBeforeScanChunk != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(245947)
			if err := knobs.RunBeforeScanChunk(startKey); err != nil {
				__antithesis_instrumentation__.Notify(245948)
				return mergeChunk{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(245949)
			}
		} else {
			__antithesis_instrumentation__.Notify(245950)
		}
	} else {
		__antithesis_instrumentation__.Notify(245951)
	}
	__antithesis_instrumentation__.Notify(245942)
	chunkSize := indexBackfillMergeBatchSize.Get(&ibm.evalCtx.Settings.SV)
	chunkBytes := indexBackfillMergeBatchBytes.Get(&ibm.evalCtx.Settings.SV)

	var nextStart roachpb.Key
	var br *roachpb.BatchResponse
	if err := ibm.flowCtx.Cfg.DB.TxnWithAdmissionControl(ctx, roachpb.AdmissionHeader_FROM_SQL, admission.BulkNormalPri,
		func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(245952)
			if err := txn.SetFixedTimestamp(ctx, readAsOf); err != nil {
				__antithesis_instrumentation__.Notify(245956)
				return err
			} else {
				__antithesis_instrumentation__.Notify(245957)
			}
			__antithesis_instrumentation__.Notify(245953)

			log.VInfof(ctx, 2, "scanning batch [%s, %s) at %v to merge", startKey, endKey, readAsOf)
			var ba roachpb.BatchRequest
			ba.TargetBytes = chunkBytes
			if err := ibm.growBoundAccount(ctx, chunkBytes); err != nil {
				__antithesis_instrumentation__.Notify(245958)
				return errors.Wrap(err, "failed to fetch keys to merge from temp index")
			} else {
				__antithesis_instrumentation__.Notify(245959)
			}
			__antithesis_instrumentation__.Notify(245954)
			defer ibm.shrinkBoundAccount(ctx, chunkBytes)

			ba.MaxSpanRequestKeys = chunkSize
			ba.Add(&roachpb.ScanRequest{
				RequestHeader: roachpb.RequestHeader{
					Key:    startKey,
					EndKey: endKey,
				},
				ScanFormat: roachpb.KEY_VALUES,
			})
			var pErr *roachpb.Error
			br, pErr = txn.Send(ctx, ba)
			if pErr != nil {
				__antithesis_instrumentation__.Notify(245960)
				return pErr.GoError()
			} else {
				__antithesis_instrumentation__.Notify(245961)
			}
			__antithesis_instrumentation__.Notify(245955)
			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(245962)
		return mergeChunk{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(245963)
	}
	__antithesis_instrumentation__.Notify(245943)

	resp := br.Responses[0].GetScan()
	chunk := mergeChunk{
		spanIdx: spanIdx,
	}
	var chunkMem int64
	if len(resp.Rows) == 0 {
		__antithesis_instrumentation__.Notify(245964)
		chunk.completedSpan = roachpb.Span{Key: startKey, EndKey: endKey}
	} else {
		__antithesis_instrumentation__.Notify(245965)
		nextStart = resp.Rows[len(resp.Rows)-1].Key.Next()
		chunk.completedSpan = roachpb.Span{Key: startKey, EndKey: nextStart}

		ibm.muBoundAccount.Lock()
		for i := range resp.Rows {
			__antithesis_instrumentation__.Notify(245967)
			chunk.keys = append(chunk.keys, resp.Rows[i].Key)
			if err := ibm.muBoundAccount.boundAccount.Grow(ctx, int64(len(resp.Rows[i].Key))); err != nil {
				__antithesis_instrumentation__.Notify(245969)
				ibm.muBoundAccount.Unlock()
				return mergeChunk{}, nil, errors.Wrap(err, "failed to allocate space for merge keys")
			} else {
				__antithesis_instrumentation__.Notify(245970)
			}
			__antithesis_instrumentation__.Notify(245968)
			chunkMem += int64(len(resp.Rows[i].Key))
		}
		__antithesis_instrumentation__.Notify(245966)
		ibm.muBoundAccount.Unlock()
	}
	__antithesis_instrumentation__.Notify(245944)
	chunk.memUsed = chunkMem
	return chunk, nextStart, nil
}

func (ibm *IndexBackfillMerger) merge(
	ctx context.Context,
	codec keys.SQLCodec,
	table catalog.TableDescriptor,
	sourceID descpb.IndexID,
	destinationID descpb.IndexID,
	sourceKeys []roachpb.Key,
	sourceSpan roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(245971)
	sourcePrefix := rowenc.MakeIndexKeyPrefix(codec, table.GetID(), sourceID)
	destPrefix := rowenc.MakeIndexKeyPrefix(codec, table.GetID(), destinationID)

	err := ibm.flowCtx.Cfg.DB.TxnWithAdmissionControl(ctx, roachpb.AdmissionHeader_FROM_SQL, admission.BulkNormalPri,
		func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(245973)
			var deletedCount int
			txn.AddCommitTrigger(func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(245979)
				log.VInfof(ctx, 2, "merged batch of %d keys (%d deletes) (span: %s) (commit timestamp: %s)",
					len(sourceKeys),
					deletedCount,
					sourceSpan,
					txn.CommitTimestamp(),
				)
			})
			__antithesis_instrumentation__.Notify(245974)
			if len(sourceKeys) == 0 {
				__antithesis_instrumentation__.Notify(245980)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(245981)
			}
			__antithesis_instrumentation__.Notify(245975)

			wb, memUsedInMerge, deletedKeys, err := ibm.constructMergeBatch(ctx, txn, sourceKeys, sourcePrefix, destPrefix)
			if err != nil {
				__antithesis_instrumentation__.Notify(245982)
				return err
			} else {
				__antithesis_instrumentation__.Notify(245983)
			}
			__antithesis_instrumentation__.Notify(245976)

			defer ibm.shrinkBoundAccount(ctx, memUsedInMerge)
			deletedCount = deletedKeys
			if err := txn.Run(ctx, wb); err != nil {
				__antithesis_instrumentation__.Notify(245984)
				return err
			} else {
				__antithesis_instrumentation__.Notify(245985)
			}
			__antithesis_instrumentation__.Notify(245977)

			if knobs, ok := ibm.flowCtx.Cfg.TestingKnobs.IndexBackfillMergerTestingKnobs.(*IndexBackfillMergerTestingKnobs); ok {
				__antithesis_instrumentation__.Notify(245986)
				if knobs != nil && func() bool {
					__antithesis_instrumentation__.Notify(245987)
					return knobs.RunDuringMergeTxn != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(245988)
					if err := knobs.RunDuringMergeTxn(ctx, txn, sourceSpan.Key, sourceSpan.EndKey); err != nil {
						__antithesis_instrumentation__.Notify(245989)
						return err
					} else {
						__antithesis_instrumentation__.Notify(245990)
					}
				} else {
					__antithesis_instrumentation__.Notify(245991)
				}
			} else {
				__antithesis_instrumentation__.Notify(245992)
			}
			__antithesis_instrumentation__.Notify(245978)
			return nil
		})
	__antithesis_instrumentation__.Notify(245972)

	return err
}

func (ibm *IndexBackfillMerger) constructMergeBatch(
	ctx context.Context,
	txn *kv.Txn,
	sourceKeys []roachpb.Key,
	sourcePrefix []byte,
	destPrefix []byte,
) (*kv.Batch, int64, int, error) {
	__antithesis_instrumentation__.Notify(245993)
	rb := txn.NewBatch()
	for i := range sourceKeys {
		__antithesis_instrumentation__.Notify(245998)
		rb.Get(sourceKeys[i])
	}
	__antithesis_instrumentation__.Notify(245994)
	if err := txn.Run(ctx, rb); err != nil {
		__antithesis_instrumentation__.Notify(245999)
		return nil, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(246000)
	}
	__antithesis_instrumentation__.Notify(245995)

	var memUsedInMerge int64
	ibm.muBoundAccount.Lock()
	defer ibm.muBoundAccount.Unlock()
	for i := range rb.Results {
		__antithesis_instrumentation__.Notify(246001)

		if rb.Results[i].Rows[0].Value == nil {
			__antithesis_instrumentation__.Notify(246004)
			return nil, 0, 0, errors.AssertionFailedf("expected value to be present in temp index for key=%s", rb.Results[i].Rows[0].Key)
		} else {
			__antithesis_instrumentation__.Notify(246005)
		}
		__antithesis_instrumentation__.Notify(246002)
		rowMem := int64(len(rb.Results[i].Rows[0].Key)) + int64(len(rb.Results[i].Rows[0].Value.RawBytes))
		if err := ibm.muBoundAccount.boundAccount.Grow(ctx, rowMem); err != nil {
			__antithesis_instrumentation__.Notify(246006)
			return nil, 0, 0, errors.Wrap(err, "failed to allocate space to read latest keys from temp index")
		} else {
			__antithesis_instrumentation__.Notify(246007)
		}
		__antithesis_instrumentation__.Notify(246003)
		memUsedInMerge += rowMem
	}
	__antithesis_instrumentation__.Notify(245996)

	prefixLen := len(sourcePrefix)
	destKey := make([]byte, len(destPrefix))
	var deletedCount int
	wb := txn.NewBatch()
	for i := range rb.Results {
		__antithesis_instrumentation__.Notify(246008)
		sourceKV := &rb.Results[i].Rows[0]
		if len(sourceKV.Key) < prefixLen {
			__antithesis_instrumentation__.Notify(246012)
			return nil, 0, 0, errors.Errorf("key for index entry %v does not start with prefix %v", sourceKV, sourcePrefix)
		} else {
			__antithesis_instrumentation__.Notify(246013)
		}
		__antithesis_instrumentation__.Notify(246009)

		destKey = destKey[:0]
		destKey = append(destKey, destPrefix...)
		destKey = append(destKey, sourceKV.Key[prefixLen:]...)

		mergedEntry, deleted, err := mergeEntry(sourceKV, destKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(246014)
			return nil, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(246015)
		}
		__antithesis_instrumentation__.Notify(246010)

		entryBytes := mergedEntryBytes(mergedEntry, deleted)
		if err := ibm.muBoundAccount.boundAccount.Grow(ctx, entryBytes); err != nil {
			__antithesis_instrumentation__.Notify(246016)
			return nil, 0, 0, errors.Wrap(err, "failed to allocate space to merge entry from temp index")
		} else {
			__antithesis_instrumentation__.Notify(246017)
		}
		__antithesis_instrumentation__.Notify(246011)
		memUsedInMerge += entryBytes
		if deleted {
			__antithesis_instrumentation__.Notify(246018)
			deletedCount++
			wb.Del(mergedEntry.Key)
		} else {
			__antithesis_instrumentation__.Notify(246019)
			wb.Put(mergedEntry.Key, mergedEntry.Value)
		}
	}
	__antithesis_instrumentation__.Notify(245997)

	return wb, memUsedInMerge, deletedCount, nil
}

func mergedEntryBytes(entry *kv.KeyValue, deleted bool) int64 {
	__antithesis_instrumentation__.Notify(246020)
	if deleted {
		__antithesis_instrumentation__.Notify(246022)
		return int64(len(entry.Key))
	} else {
		__antithesis_instrumentation__.Notify(246023)
	}
	__antithesis_instrumentation__.Notify(246021)

	return int64(len(entry.Key) + len(entry.Value.RawBytes))
}

func mergeEntry(sourceKV *kv.KeyValue, destKey roachpb.Key) (*kv.KeyValue, bool, error) {
	__antithesis_instrumentation__.Notify(246024)
	var destTagAndData []byte
	var deleted bool

	tempWrapper, err := rowenc.DecodeWrapper(sourceKV.Value)
	if err != nil {
		__antithesis_instrumentation__.Notify(246027)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(246028)
	}
	__antithesis_instrumentation__.Notify(246025)

	if tempWrapper.Deleted {
		__antithesis_instrumentation__.Notify(246029)
		deleted = true
	} else {
		__antithesis_instrumentation__.Notify(246030)
		destTagAndData = tempWrapper.Value
	}
	__antithesis_instrumentation__.Notify(246026)

	value := &roachpb.Value{}
	value.SetTagAndData(destTagAndData)

	return &kv.KeyValue{
		Key:   destKey.Clone(),
		Value: value,
	}, deleted, nil
}

func (ibm *IndexBackfillMerger) growBoundAccount(ctx context.Context, growBy int64) error {
	__antithesis_instrumentation__.Notify(246031)
	defer ibm.muBoundAccount.Unlock()
	ibm.muBoundAccount.Lock()
	return ibm.muBoundAccount.boundAccount.Grow(ctx, growBy)
}

func (ibm *IndexBackfillMerger) shrinkBoundAccount(ctx context.Context, shrinkBy int64) {
	__antithesis_instrumentation__.Notify(246032)
	defer ibm.muBoundAccount.Unlock()
	ibm.muBoundAccount.Lock()
	ibm.muBoundAccount.boundAccount.Shrink(ctx, shrinkBy)
}

func NewIndexBackfillMerger(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	spec execinfrapb.IndexBackfillMergerSpec,
	output execinfra.RowReceiver,
) (*IndexBackfillMerger, error) {
	__antithesis_instrumentation__.Notify(246033)
	mergerMon := execinfra.NewMonitor(ctx, flowCtx.Cfg.BackfillerMonitor,
		"index-backfiller-merger-mon")

	ibm := &IndexBackfillMerger{
		spec:    spec,
		desc:    tabledesc.NewUnsafeImmutable(&spec.Table),
		flowCtx: flowCtx,
		evalCtx: flowCtx.NewEvalCtx(),
		output:  output,
		mon:     mergerMon,
	}

	ibm.muBoundAccount.boundAccount = mergerMon.MakeBoundAccount()
	return ibm, nil
}

type IndexBackfillMergerTestingKnobs struct {
	RunBeforeScanChunk func(startKey roachpb.Key) error

	RunAfterScanChunk func()

	RunDuringMergeTxn func(ctx context.Context, txn *kv.Txn, startKey roachpb.Key, endKey roachpb.Key) error

	PushesProgressEveryChunk bool

	RunAfterMergeChunk func()
}

var _ base.ModuleTestingKnobs = &IndexBackfillMergerTestingKnobs{}

func (*IndexBackfillMergerTestingKnobs) ModuleTestingKnobs() {
	__antithesis_instrumentation__.Notify(246034)
}
