package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

var csvOutputTypes = []*types.T{
	types.Bytes,
	types.Bytes,
}

const readImportDataProcessorName = "readImportDataProcessor"

var importPKAdderBufferSize = func() *settings.ByteSizeSetting {
	__antithesis_instrumentation__.Notify(494530)
	s := settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"kv.bulk_ingest.pk_buffer_size",
		"the initial size of the BulkAdder buffer handling primary index imports",
		32<<20,
	)
	return s
}()

var importPKAdderMaxBufferSize = func() *settings.ByteSizeSetting {
	__antithesis_instrumentation__.Notify(494531)
	s := settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"kv.bulk_ingest.max_pk_buffer_size",
		"the maximum size of the BulkAdder buffer handling primary index imports",
		128<<20,
	)
	return s
}()

var importIndexAdderBufferSize = func() *settings.ByteSizeSetting {
	__antithesis_instrumentation__.Notify(494532)
	s := settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"kv.bulk_ingest.index_buffer_size",
		"the initial size of the BulkAdder buffer handling secondary index imports",
		32<<20,
	)
	return s
}()

var importIndexAdderMaxBufferSize = func() *settings.ByteSizeSetting {
	__antithesis_instrumentation__.Notify(494533)
	s := settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		"kv.bulk_ingest.max_index_buffer_size",
		"the maximum size of the BulkAdder buffer handling secondary index imports",
		512<<20,
	)
	return s
}()

var importAtNow = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"bulkio.import_at_current_time.enabled",
	"write imported data at the current timestamp, when each batch is flushed",
	true,
)

var readerParallelismSetting = settings.RegisterIntSetting(
	settings.TenantWritable,
	"bulkio.import.reader_parallelism",
	"number of parallel workers to use to convert read data for formats that support parallel conversion; 0 indicates number of cores",
	0,
	settings.NonNegativeInt,
)

func importBufferConfigSizes(st *cluster.Settings, isPKAdder bool) (int64, func() int64) {
	__antithesis_instrumentation__.Notify(494534)
	if isPKAdder {
		__antithesis_instrumentation__.Notify(494536)
		return importPKAdderBufferSize.Get(&st.SV),
			func() int64 {
				__antithesis_instrumentation__.Notify(494537)
				return importPKAdderMaxBufferSize.Get(&st.SV)
			}
	} else {
		__antithesis_instrumentation__.Notify(494538)
	}
	__antithesis_instrumentation__.Notify(494535)
	return importIndexAdderBufferSize.Get(&st.SV),
		func() int64 {
			__antithesis_instrumentation__.Notify(494539)
			return importIndexAdderMaxBufferSize.Get(&st.SV)
		}
}

type readImportDataProcessor struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.ReadImportDataSpec
	output  execinfra.RowReceiver

	progCh chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress

	seqChunkProvider *row.SeqChunkProvider

	importErr error
	summary   *roachpb.BulkOpSummary
}

var (
	_ execinfra.Processor = &readImportDataProcessor{}
	_ execinfra.RowSource = &readImportDataProcessor{}
)

func newReadImportDataProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.ReadImportDataSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(494540)
	cp := &readImportDataProcessor{
		flowCtx: flowCtx,
		spec:    spec,
		output:  output,
		progCh:  make(chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress),
	}
	if err := cp.Init(cp, post, csvOutputTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{

			InputsToDrain: nil,
		}); err != nil {
		__antithesis_instrumentation__.Notify(494543)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494544)
	}
	__antithesis_instrumentation__.Notify(494541)

	if cp.flowCtx.Cfg.JobRegistry != nil {
		__antithesis_instrumentation__.Notify(494545)
		cp.seqChunkProvider = &row.SeqChunkProvider{
			JobID:    cp.spec.Progress.JobID,
			Registry: cp.flowCtx.Cfg.JobRegistry,
		}
	} else {
		__antithesis_instrumentation__.Notify(494546)
	}
	__antithesis_instrumentation__.Notify(494542)

	return cp, nil
}

func (idp *readImportDataProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(494547)
	ctx = logtags.AddTag(ctx, "job", idp.spec.JobID)
	ctx = idp.StartInternal(ctx, readImportDataProcessorName)

	go func() {
		__antithesis_instrumentation__.Notify(494548)
		defer close(idp.progCh)
		idp.summary, idp.importErr = runImport(ctx, idp.flowCtx, &idp.spec, idp.progCh,
			idp.seqChunkProvider)
	}()
}

func (idp *readImportDataProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(494549)
	if idp.State != execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(494555)
		return nil, idp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(494556)
	}
	__antithesis_instrumentation__.Notify(494550)

	for prog := range idp.progCh {
		__antithesis_instrumentation__.Notify(494557)
		p := prog
		return nil, &execinfrapb.ProducerMetadata{BulkProcessorProgress: &p}
	}
	__antithesis_instrumentation__.Notify(494551)

	if idp.importErr != nil {
		__antithesis_instrumentation__.Notify(494558)
		idp.MoveToDraining(idp.importErr)
		return nil, idp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(494559)
	}
	__antithesis_instrumentation__.Notify(494552)

	if idp.summary == nil {
		__antithesis_instrumentation__.Notify(494560)
		err := errors.Newf("no summary generated by %s", readImportDataProcessorName)
		idp.MoveToDraining(err)
		return nil, idp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(494561)
	}
	__antithesis_instrumentation__.Notify(494553)

	countsBytes, err := protoutil.Marshal(idp.summary)
	idp.MoveToDraining(err)
	if err != nil {
		__antithesis_instrumentation__.Notify(494562)
		return nil, idp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(494563)
	}
	__antithesis_instrumentation__.Notify(494554)

	return rowenc.EncDatumRow{
		rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(countsBytes))),
		rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes([]byte{}))),
	}, nil
}

func injectTimeIntoEvalCtx(ctx *tree.EvalContext, walltime int64) {
	__antithesis_instrumentation__.Notify(494564)
	sec := walltime / int64(time.Second)
	nsec := walltime % int64(time.Second)
	unixtime := timeutil.Unix(sec, nsec)
	ctx.StmtTimestamp = unixtime
	ctx.TxnTimestamp = unixtime
}

func makeInputConverter(
	ctx context.Context,
	semaCtx *tree.SemaContext,
	spec *execinfrapb.ReadImportDataSpec,
	evalCtx *tree.EvalContext,
	kvCh chan row.KVBatch,
	seqChunkProvider *row.SeqChunkProvider,
) (inputConverter, error) {
	__antithesis_instrumentation__.Notify(494565)
	injectTimeIntoEvalCtx(evalCtx, spec.WalltimeNanos)
	var singleTable catalog.TableDescriptor
	var singleTableTargetCols tree.NameList
	if len(spec.Tables) == 1 {
		__antithesis_instrumentation__.Notify(494571)
		for _, table := range spec.Tables {
			__antithesis_instrumentation__.Notify(494572)
			singleTable = tabledesc.NewBuilder(table.Desc).BuildImmutableTable()
			singleTableTargetCols = make(tree.NameList, len(table.TargetCols))
			for i, colName := range table.TargetCols {
				__antithesis_instrumentation__.Notify(494573)
				singleTableTargetCols[i] = tree.Name(colName)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(494574)
	}
	__antithesis_instrumentation__.Notify(494566)

	if format := spec.Format.Format; singleTable == nil && func() bool {
		__antithesis_instrumentation__.Notify(494575)
		return !isMultiTableFormat(format) == true
	}() == true {
		__antithesis_instrumentation__.Notify(494576)
		return nil, errors.Errorf("%s only supports reading a single, pre-specified table", format.String())
	} else {
		__antithesis_instrumentation__.Notify(494577)
	}
	__antithesis_instrumentation__.Notify(494567)

	if singleTable != nil {
		__antithesis_instrumentation__.Notify(494578)
		if idx := catalog.FindDeletableNonPrimaryIndex(singleTable, func(idx catalog.Index) bool {
			__antithesis_instrumentation__.Notify(494580)
			return idx.IsPartial()
		}); idx != nil {
			__antithesis_instrumentation__.Notify(494581)
			return nil, unimplemented.NewWithIssue(50225, "cannot import into table with partial indexes")
		} else {
			__antithesis_instrumentation__.Notify(494582)
		}
		__antithesis_instrumentation__.Notify(494579)

		if len(singleTableTargetCols) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(494583)
			return !formatHasNamedColumns(spec.Format.Format) == true
		}() == true {
			__antithesis_instrumentation__.Notify(494584)
			for _, col := range singleTable.VisibleColumns() {
				__antithesis_instrumentation__.Notify(494585)
				if col.IsComputed() {
					__antithesis_instrumentation__.Notify(494586)
					return nil, unimplemented.NewWithIssueDetail(56002, "import.computed",
						"to use computed columns, use IMPORT INTO")
				} else {
					__antithesis_instrumentation__.Notify(494587)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(494588)
		}
	} else {
		__antithesis_instrumentation__.Notify(494589)
	}
	__antithesis_instrumentation__.Notify(494568)

	readerParallelism := int(spec.ReaderParallelism)
	if readerParallelism <= 0 {
		__antithesis_instrumentation__.Notify(494590)
		readerParallelism = int(readerParallelismSetting.Get(&evalCtx.Settings.SV))
	} else {
		__antithesis_instrumentation__.Notify(494591)
	}
	__antithesis_instrumentation__.Notify(494569)
	if readerParallelism <= 0 {
		__antithesis_instrumentation__.Notify(494592)
		readerParallelism = runtime.GOMAXPROCS(0)
	} else {
		__antithesis_instrumentation__.Notify(494593)
	}
	__antithesis_instrumentation__.Notify(494570)

	switch spec.Format.Format {
	case roachpb.IOFileFormat_CSV:
		__antithesis_instrumentation__.Notify(494594)
		isWorkload := true
		for _, file := range spec.Uri {
			__antithesis_instrumentation__.Notify(494603)
			if _, err := parseWorkloadConfig(file); err != nil {
				__antithesis_instrumentation__.Notify(494604)
				isWorkload = false
				break
			} else {
				__antithesis_instrumentation__.Notify(494605)
			}
		}
		__antithesis_instrumentation__.Notify(494595)
		if isWorkload {
			__antithesis_instrumentation__.Notify(494606)
			return newWorkloadReader(semaCtx, evalCtx, singleTable, kvCh, readerParallelism), nil
		} else {
			__antithesis_instrumentation__.Notify(494607)
		}
		__antithesis_instrumentation__.Notify(494596)
		return newCSVInputReader(
			semaCtx, kvCh, spec.Format.Csv, spec.WalltimeNanos, readerParallelism,
			singleTable, singleTableTargetCols, evalCtx, seqChunkProvider), nil
	case roachpb.IOFileFormat_MysqlOutfile:
		__antithesis_instrumentation__.Notify(494597)
		return newMysqloutfileReader(
			semaCtx, spec.Format.MysqlOut, kvCh, spec.WalltimeNanos,
			readerParallelism, singleTable, singleTableTargetCols, evalCtx)
	case roachpb.IOFileFormat_Mysqldump:
		__antithesis_instrumentation__.Notify(494598)
		return newMysqldumpReader(ctx, semaCtx, kvCh, spec.WalltimeNanos, spec.Tables, evalCtx,
			spec.Format.MysqlDump)
	case roachpb.IOFileFormat_PgCopy:
		__antithesis_instrumentation__.Notify(494599)
		return newPgCopyReader(semaCtx, spec.Format.PgCopy, kvCh, spec.WalltimeNanos,
			readerParallelism, singleTable, singleTableTargetCols, evalCtx)
	case roachpb.IOFileFormat_PgDump:
		__antithesis_instrumentation__.Notify(494600)
		return newPgDumpReader(ctx, semaCtx, int64(spec.Progress.JobID), kvCh, spec.Format.PgDump,
			spec.WalltimeNanos, spec.Tables, evalCtx)
	case roachpb.IOFileFormat_Avro:
		__antithesis_instrumentation__.Notify(494601)
		return newAvroInputReader(
			semaCtx, kvCh, singleTable, spec.Format.Avro, spec.WalltimeNanos,
			readerParallelism, evalCtx)
	default:
		__antithesis_instrumentation__.Notify(494602)
		return nil, errors.Errorf(
			"Requested IMPORT format (%d) not supported by this node", spec.Format.Format)
	}
}

type tableAndIndex struct {
	tableID catid.DescID
	indexID catid.IndexID
}

func ingestKvs(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.ReadImportDataSpec,
	progCh chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
	kvCh <-chan row.KVBatch,
) (*roachpb.BulkOpSummary, error) {
	__antithesis_instrumentation__.Notify(494608)
	ctx, span := tracing.ChildSpan(ctx, "import-ingest-kvs")
	defer span.Finish()

	writeTS := hlc.Timestamp{WallTime: spec.WalltimeNanos}
	writeAtBatchTimestamp := true
	if !importAtNow.Get(&flowCtx.Cfg.Settings.SV) {
		__antithesis_instrumentation__.Notify(494623)
		log.Warningf(ctx, "ingesting import data with raw timestamps due to cluster setting")
		writeAtBatchTimestamp = false
	} else {
		__antithesis_instrumentation__.Notify(494624)
		if !flowCtx.Cfg.Settings.Version.IsActive(ctx, clusterversion.MVCCAddSSTable) {
			__antithesis_instrumentation__.Notify(494625)
			log.Warningf(ctx, "ingesting import data with raw timestamps due to cluster version")
			writeAtBatchTimestamp = false
		} else {
			__antithesis_instrumentation__.Notify(494626)
		}
	}
	__antithesis_instrumentation__.Notify(494609)

	var pkAdderName, indexAdderName = "rows", "indexes"
	if len(spec.Tables) == 1 {
		__antithesis_instrumentation__.Notify(494627)
		for k := range spec.Tables {
			__antithesis_instrumentation__.Notify(494628)
			pkAdderName = fmt.Sprintf("%s rows", k)
			indexAdderName = fmt.Sprintf("%s indexes", k)
		}
	} else {
		__antithesis_instrumentation__.Notify(494629)
	}
	__antithesis_instrumentation__.Notify(494610)

	isPK := make(map[tableAndIndex]bool, len(spec.Tables))
	for _, t := range spec.Tables {
		__antithesis_instrumentation__.Notify(494630)
		isPK[tableAndIndex{tableID: t.Desc.ID, indexID: t.Desc.PrimaryIndex.ID}] = true
	}
	__antithesis_instrumentation__.Notify(494611)

	minBufferSize, maxBufferSize := importBufferConfigSizes(flowCtx.Cfg.Settings,
		true)
	pkIndexAdder, err := flowCtx.Cfg.BulkAdder(ctx, flowCtx.Cfg.DB, writeTS, kvserverbase.BulkAdderOptions{
		Name:                     pkAdderName,
		DisallowShadowingBelow:   writeTS,
		SkipDuplicates:           true,
		MinBufferSize:            minBufferSize,
		MaxBufferSize:            maxBufferSize,
		InitialSplitsIfUnordered: int(spec.InitialSplits),
		WriteAtBatchTimestamp:    writeAtBatchTimestamp,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(494631)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494632)
	}
	__antithesis_instrumentation__.Notify(494612)
	defer pkIndexAdder.Close(ctx)

	minBufferSize, maxBufferSize = importBufferConfigSizes(flowCtx.Cfg.Settings,
		false)
	indexAdder, err := flowCtx.Cfg.BulkAdder(ctx, flowCtx.Cfg.DB, writeTS, kvserverbase.BulkAdderOptions{
		Name:                     indexAdderName,
		DisallowShadowingBelow:   writeTS,
		SkipDuplicates:           true,
		MinBufferSize:            minBufferSize,
		MaxBufferSize:            maxBufferSize,
		InitialSplitsIfUnordered: int(spec.InitialSplits),
		WriteAtBatchTimestamp:    writeAtBatchTimestamp,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(494633)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494634)
	}
	__antithesis_instrumentation__.Notify(494613)
	defer indexAdder.Close(ctx)

	writtenRow := make([]int64, len(spec.Uri))
	writtenFraction := make([]uint32, len(spec.Uri))

	pkFlushedRow := make([]int64, len(spec.Uri))
	idxFlushedRow := make([]int64, len(spec.Uri))

	bulkSummaryMu := &struct {
		syncutil.Mutex
		summary roachpb.BulkOpSummary
	}{}

	pkIndexAdder.SetOnFlush(func(summary roachpb.BulkOpSummary) {
		__antithesis_instrumentation__.Notify(494635)
		for i, emitted := range writtenRow {
			__antithesis_instrumentation__.Notify(494637)
			atomic.StoreInt64(&pkFlushedRow[i], emitted)
			bulkSummaryMu.Lock()
			bulkSummaryMu.summary.Add(summary)
			bulkSummaryMu.Unlock()
		}
		__antithesis_instrumentation__.Notify(494636)
		if indexAdder.IsEmpty() {
			__antithesis_instrumentation__.Notify(494638)
			for i, emitted := range writtenRow {
				__antithesis_instrumentation__.Notify(494639)
				atomic.StoreInt64(&idxFlushedRow[i], emitted)
			}
		} else {
			__antithesis_instrumentation__.Notify(494640)
		}
	})
	__antithesis_instrumentation__.Notify(494614)
	indexAdder.SetOnFlush(func(summary roachpb.BulkOpSummary) {
		__antithesis_instrumentation__.Notify(494641)
		for i, emitted := range writtenRow {
			__antithesis_instrumentation__.Notify(494642)
			atomic.StoreInt64(&idxFlushedRow[i], emitted)
			bulkSummaryMu.Lock()
			bulkSummaryMu.summary.Add(summary)
			bulkSummaryMu.Unlock()
		}
	})
	__antithesis_instrumentation__.Notify(494615)

	offsets := make(map[int32]int, len(spec.Uri))
	var offset int
	for i := range spec.Uri {
		__antithesis_instrumentation__.Notify(494643)
		offsets[i] = offset
		offset++
	}
	__antithesis_instrumentation__.Notify(494616)

	pushProgress := func() {
		__antithesis_instrumentation__.Notify(494644)
		var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
		prog.ResumePos = make(map[int32]int64)
		prog.CompletedFraction = make(map[int32]float32)
		for file, offset := range offsets {
			__antithesis_instrumentation__.Notify(494646)
			pk := atomic.LoadInt64(&pkFlushedRow[offset])
			idx := atomic.LoadInt64(&idxFlushedRow[offset])

			if idx > pk {
				__antithesis_instrumentation__.Notify(494648)
				prog.ResumePos[file] = pk
			} else {
				__antithesis_instrumentation__.Notify(494649)
				prog.ResumePos[file] = idx
			}
			__antithesis_instrumentation__.Notify(494647)
			prog.CompletedFraction[file] = math.Float32frombits(atomic.LoadUint32(&writtenFraction[offset]))

			bulkSummaryMu.Lock()
			prog.BulkSummary = bulkSummaryMu.summary
			bulkSummaryMu.summary.Reset()
			bulkSummaryMu.Unlock()
		}
		__antithesis_instrumentation__.Notify(494645)
		progCh <- prog
	}
	__antithesis_instrumentation__.Notify(494617)

	stopProgress := make(chan struct{})
	g := ctxgroup.WithContext(ctx)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(494650)
		tick := time.NewTicker(time.Second * 10)
		defer tick.Stop()
		done := ctx.Done()
		for {
			__antithesis_instrumentation__.Notify(494651)
			select {
			case <-done:
				__antithesis_instrumentation__.Notify(494652)
				return ctx.Err()
			case <-stopProgress:
				__antithesis_instrumentation__.Notify(494653)
				return nil
			case <-tick.C:
				__antithesis_instrumentation__.Notify(494654)
				pushProgress()
			}
		}
	})
	__antithesis_instrumentation__.Notify(494618)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(494655)
		defer close(stopProgress)

		for kvBatch := range kvCh {
			__antithesis_instrumentation__.Notify(494657)
			for _, kv := range kvBatch.KVs {
				__antithesis_instrumentation__.Notify(494659)
				_, tableID, indexID, indexErr := flowCtx.Codec().DecodeIndexPrefix(kv.Key)
				if indexErr != nil {
					__antithesis_instrumentation__.Notify(494661)
					return indexErr
				} else {
					__antithesis_instrumentation__.Notify(494662)
				}
				__antithesis_instrumentation__.Notify(494660)

				if isPK[tableAndIndex{tableID: catid.DescID(tableID), indexID: catid.IndexID(indexID)}] {
					__antithesis_instrumentation__.Notify(494663)
					if err := pkIndexAdder.Add(ctx, kv.Key, kv.Value.RawBytes); err != nil {
						__antithesis_instrumentation__.Notify(494664)
						if errors.HasType(err, (*kvserverbase.DuplicateKeyError)(nil)) {
							__antithesis_instrumentation__.Notify(494666)
							return errors.Wrap(err, "duplicate key in primary index")
						} else {
							__antithesis_instrumentation__.Notify(494667)
						}
						__antithesis_instrumentation__.Notify(494665)
						return err
					} else {
						__antithesis_instrumentation__.Notify(494668)
					}
				} else {
					__antithesis_instrumentation__.Notify(494669)
					if err := indexAdder.Add(ctx, kv.Key, kv.Value.RawBytes); err != nil {
						__antithesis_instrumentation__.Notify(494670)
						if errors.HasType(err, (*kvserverbase.DuplicateKeyError)(nil)) {
							__antithesis_instrumentation__.Notify(494672)
							return errors.Wrap(err, "duplicate key in index")
						} else {
							__antithesis_instrumentation__.Notify(494673)
						}
						__antithesis_instrumentation__.Notify(494671)
						return err
					} else {
						__antithesis_instrumentation__.Notify(494674)
					}
				}
			}
			__antithesis_instrumentation__.Notify(494658)
			offset := offsets[kvBatch.Source]
			writtenRow[offset] = kvBatch.LastRow
			atomic.StoreUint32(&writtenFraction[offset], math.Float32bits(kvBatch.Progress))
			if flowCtx.Cfg.TestingKnobs.BulkAdderFlushesEveryBatch {
				__antithesis_instrumentation__.Notify(494675)
				_ = pkIndexAdder.Flush(ctx)
				_ = indexAdder.Flush(ctx)
				pushProgress()
			} else {
				__antithesis_instrumentation__.Notify(494676)
			}
		}
		__antithesis_instrumentation__.Notify(494656)
		return nil
	})
	__antithesis_instrumentation__.Notify(494619)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(494677)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494678)
	}
	__antithesis_instrumentation__.Notify(494620)

	if err := pkIndexAdder.Flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(494679)
		if errors.HasType(err, (*kvserverbase.DuplicateKeyError)(nil)) {
			__antithesis_instrumentation__.Notify(494681)
			return nil, errors.Wrap(err, "duplicate key in primary index")
		} else {
			__antithesis_instrumentation__.Notify(494682)
		}
		__antithesis_instrumentation__.Notify(494680)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494683)
	}
	__antithesis_instrumentation__.Notify(494621)

	if err := indexAdder.Flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(494684)
		if errors.HasType(err, (*kvserverbase.DuplicateKeyError)(nil)) {
			__antithesis_instrumentation__.Notify(494686)
			return nil, errors.Wrap(err, "duplicate key in index")
		} else {
			__antithesis_instrumentation__.Notify(494687)
		}
		__antithesis_instrumentation__.Notify(494685)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(494688)
	}
	__antithesis_instrumentation__.Notify(494622)

	addedSummary := pkIndexAdder.GetSummary()
	addedSummary.Add(indexAdder.GetSummary())
	return &addedSummary, nil
}

func init() {
	rowexec.NewReadImportDataProcessor = newReadImportDataProcessor
}
