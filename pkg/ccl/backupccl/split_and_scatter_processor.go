package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type splitAndScatterer interface {
	split(ctx context.Context, codec keys.SQLCodec, splitKey roachpb.Key) error

	scatter(ctx context.Context, codec keys.SQLCodec, scatterKey roachpb.Key) (roachpb.NodeID, error)
}

type noopSplitAndScatterer struct{}

var _ splitAndScatterer = noopSplitAndScatterer{}

func (n noopSplitAndScatterer) split(_ context.Context, _ keys.SQLCodec, _ roachpb.Key) error {
	__antithesis_instrumentation__.Notify(13110)
	return nil
}

func (n noopSplitAndScatterer) scatter(
	_ context.Context, _ keys.SQLCodec, _ roachpb.Key,
) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(13111)
	return 0, nil
}

type dbSplitAndScatterer struct {
	db *kv.DB
	kr *KeyRewriter
}

var _ splitAndScatterer = dbSplitAndScatterer{}

func makeSplitAndScatterer(db *kv.DB, kr *KeyRewriter) splitAndScatterer {
	__antithesis_instrumentation__.Notify(13112)
	return dbSplitAndScatterer{db: db, kr: kr}
}

func (s dbSplitAndScatterer) split(
	ctx context.Context, codec keys.SQLCodec, splitKey roachpb.Key,
) error {
	__antithesis_instrumentation__.Notify(13113)
	if s.kr == nil {
		__antithesis_instrumentation__.Notify(13119)
		return errors.AssertionFailedf("KeyRewriter was not set when expected to be")
	} else {
		__antithesis_instrumentation__.Notify(13120)
	}
	__antithesis_instrumentation__.Notify(13114)
	if s.db == nil {
		__antithesis_instrumentation__.Notify(13121)
		return errors.AssertionFailedf("split and scatterer's database was not set when expected")
	} else {
		__antithesis_instrumentation__.Notify(13122)
	}
	__antithesis_instrumentation__.Notify(13115)

	expirationTime := s.db.Clock().Now().Add(time.Hour.Nanoseconds(), 0)
	newSplitKey, err := rewriteBackupSpanKey(codec, s.kr, splitKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(13123)
		return err
	} else {
		__antithesis_instrumentation__.Notify(13124)
	}
	__antithesis_instrumentation__.Notify(13116)
	if splitAt, err := keys.EnsureSafeSplitKey(newSplitKey); err != nil {
		__antithesis_instrumentation__.Notify(13125)

	} else {
		__antithesis_instrumentation__.Notify(13126)
		if len(splitAt) != 0 {
			__antithesis_instrumentation__.Notify(13127)
			newSplitKey = splitAt
		} else {
			__antithesis_instrumentation__.Notify(13128)
		}
	}
	__antithesis_instrumentation__.Notify(13117)
	log.VEventf(ctx, 1, "presplitting new key %+v", newSplitKey)
	if err := s.db.AdminSplit(ctx, newSplitKey, expirationTime); err != nil {
		__antithesis_instrumentation__.Notify(13129)
		return errors.Wrapf(err, "splitting key %s", newSplitKey)
	} else {
		__antithesis_instrumentation__.Notify(13130)
	}
	__antithesis_instrumentation__.Notify(13118)

	return nil
}

func (s dbSplitAndScatterer) scatter(
	ctx context.Context, codec keys.SQLCodec, scatterKey roachpb.Key,
) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(13131)
	if s.kr == nil {
		__antithesis_instrumentation__.Notify(13137)
		return 0, errors.AssertionFailedf("KeyRewriter was not set when expected to be")
	} else {
		__antithesis_instrumentation__.Notify(13138)
	}
	__antithesis_instrumentation__.Notify(13132)
	if s.db == nil {
		__antithesis_instrumentation__.Notify(13139)
		return 0, errors.AssertionFailedf("split and scatterer's database was not set when expected")
	} else {
		__antithesis_instrumentation__.Notify(13140)
	}
	__antithesis_instrumentation__.Notify(13133)

	newScatterKey, err := rewriteBackupSpanKey(codec, s.kr, scatterKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(13141)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(13142)
	}
	__antithesis_instrumentation__.Notify(13134)
	if scatterAt, err := keys.EnsureSafeSplitKey(newScatterKey); err != nil {
		__antithesis_instrumentation__.Notify(13143)

	} else {
		__antithesis_instrumentation__.Notify(13144)
		if len(scatterAt) != 0 {
			__antithesis_instrumentation__.Notify(13145)
			newScatterKey = scatterAt
		} else {
			__antithesis_instrumentation__.Notify(13146)
		}
	}
	__antithesis_instrumentation__.Notify(13135)

	log.VEventf(ctx, 1, "scattering new key %+v", newScatterKey)
	req := &roachpb.AdminScatterRequest{
		RequestHeader: roachpb.RequestHeaderFromSpan(roachpb.Span{
			Key:    newScatterKey,
			EndKey: newScatterKey.Next(),
		}),

		RandomizeLeases: true,
		MaxSize:         1,
	}

	res, pErr := kv.SendWrapped(ctx, s.db.NonTransactionalSender(), req)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(13147)

		if !strings.Contains(pErr.String(), "existing range size") {
			__antithesis_instrumentation__.Notify(13149)

			log.Errorf(ctx, "failed to scatter span [%s,%s): %+v",
				newScatterKey, newScatterKey.Next(), pErr.GoError())
		} else {
			__antithesis_instrumentation__.Notify(13150)
		}
		__antithesis_instrumentation__.Notify(13148)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(13151)
	}
	__antithesis_instrumentation__.Notify(13136)

	return s.findDestination(res.(*roachpb.AdminScatterResponse)), nil
}

func (s dbSplitAndScatterer) findDestination(res *roachpb.AdminScatterResponse) roachpb.NodeID {
	__antithesis_instrumentation__.Notify(13152)

	if len(res.RangeInfos) > 0 {
		__antithesis_instrumentation__.Notify(13154)

		return res.RangeInfos[0].Lease.Replica.NodeID
	} else {
		__antithesis_instrumentation__.Notify(13155)
	}
	__antithesis_instrumentation__.Notify(13153)

	return roachpb.NodeID(0)
}

const splitAndScatterProcessorName = "splitAndScatter"

var splitAndScatterOutputTypes = []*types.T{
	types.Bytes,
	types.Bytes,
}

type splitAndScatterProcessor struct {
	execinfra.ProcessorBase

	flowCtx *execinfra.FlowCtx
	spec    execinfrapb.SplitAndScatterSpec
	output  execinfra.RowReceiver

	scatterer splitAndScatterer

	cancelScatterAndWaitForWorker func()

	doneScatterCh chan entryNode

	routingDatumCache map[roachpb.NodeID]rowenc.EncDatum
	scatterErr        error
}

var _ execinfra.Processor = &splitAndScatterProcessor{}

func newSplitAndScatterProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.SplitAndScatterSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(13156)

	if spec.Validation != jobspb.RestoreValidation_DefaultRestore {
		__antithesis_instrumentation__.Notify(13161)
		return nil, errors.New("Split and Scatter Processor does not support validation yet")
	} else {
		__antithesis_instrumentation__.Notify(13162)
	}
	__antithesis_instrumentation__.Notify(13157)

	numEntries := 0
	for _, chunk := range spec.Chunks {
		__antithesis_instrumentation__.Notify(13163)
		numEntries += len(chunk.Entries)
	}
	__antithesis_instrumentation__.Notify(13158)

	db := flowCtx.Cfg.DB
	kr, err := makeKeyRewriterFromRekeys(flowCtx.Codec(), spec.TableRekeys, spec.TenantRekeys)
	if err != nil {
		__antithesis_instrumentation__.Notify(13164)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(13165)
	}
	__antithesis_instrumentation__.Notify(13159)

	scatterer := makeSplitAndScatterer(db, kr)
	ssp := &splitAndScatterProcessor{
		flowCtx:   flowCtx,
		spec:      spec,
		output:    output,
		scatterer: scatterer,

		doneScatterCh:     make(chan entryNode, numEntries),
		routingDatumCache: make(map[roachpb.NodeID]rowenc.EncDatum),
	}
	if err := ssp.Init(ssp, post, splitAndScatterOutputTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{
			InputsToDrain: nil,
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(13166)
				ssp.close()
				return nil
			},
		}); err != nil {
		__antithesis_instrumentation__.Notify(13167)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(13168)
	}
	__antithesis_instrumentation__.Notify(13160)
	return ssp, nil
}

func (ssp *splitAndScatterProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(13169)
	ctx = logtags.AddTag(ctx, "job", ssp.spec.JobID)
	ctx = ssp.StartInternal(ctx, splitAndScatterProcessorName)

	scatterCtx, cancel := context.WithCancel(ctx)
	workerDone := make(chan struct{})
	ssp.cancelScatterAndWaitForWorker = func() {
		__antithesis_instrumentation__.Notify(13171)
		cancel()
		<-workerDone
	}
	__antithesis_instrumentation__.Notify(13170)
	if err := ssp.flowCtx.Stopper().RunAsyncTaskEx(scatterCtx, stop.TaskOpts{
		TaskName: "splitAndScatter-worker",
		SpanOpt:  stop.ChildSpan,
	}, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(13172)
		ssp.scatterErr = runSplitAndScatter(scatterCtx, ssp.flowCtx, &ssp.spec, ssp.scatterer, ssp.doneScatterCh)
		cancel()
		close(ssp.doneScatterCh)
		close(workerDone)
	}); err != nil {
		__antithesis_instrumentation__.Notify(13173)
		ssp.scatterErr = err
		cancel()
		close(workerDone)
	} else {
		__antithesis_instrumentation__.Notify(13174)
	}
}

type entryNode struct {
	entry execinfrapb.RestoreSpanEntry
	node  roachpb.NodeID
}

func (ssp *splitAndScatterProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(13175)
	if ssp.State != execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(13179)
		return nil, ssp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(13180)
	}
	__antithesis_instrumentation__.Notify(13176)

	scatteredEntry, ok := <-ssp.doneScatterCh
	if ok {
		__antithesis_instrumentation__.Notify(13181)
		entry := scatteredEntry.entry
		entryBytes, err := protoutil.Marshal(&entry)
		if err != nil {
			__antithesis_instrumentation__.Notify(13184)
			ssp.MoveToDraining(err)
			return nil, ssp.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(13185)
		}
		__antithesis_instrumentation__.Notify(13182)

		routingDatum, ok := ssp.routingDatumCache[scatteredEntry.node]
		if !ok {
			__antithesis_instrumentation__.Notify(13186)
			routingDatum, _ = routingDatumsForSQLInstance(base.SQLInstanceID(scatteredEntry.node))
			ssp.routingDatumCache[scatteredEntry.node] = routingDatum
		} else {
			__antithesis_instrumentation__.Notify(13187)
		}
		__antithesis_instrumentation__.Notify(13183)

		row := rowenc.EncDatumRow{
			routingDatum,
			rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(entryBytes))),
		}
		return row, nil
	} else {
		__antithesis_instrumentation__.Notify(13188)
	}
	__antithesis_instrumentation__.Notify(13177)

	if ssp.scatterErr != nil {
		__antithesis_instrumentation__.Notify(13189)
		ssp.MoveToDraining(ssp.scatterErr)
		return nil, ssp.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(13190)
	}
	__antithesis_instrumentation__.Notify(13178)

	ssp.MoveToDraining(nil)
	return nil, ssp.DrainHelper()
}

func (ssp *splitAndScatterProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(13191)

	ssp.close()
}

func (ssp *splitAndScatterProcessor) close() {
	__antithesis_instrumentation__.Notify(13192)
	ssp.cancelScatterAndWaitForWorker()
	ssp.InternalClose()
}

type scatteredChunk struct {
	destination roachpb.NodeID
	entries     []execinfrapb.RestoreSpanEntry
}

func runSplitAndScatter(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.SplitAndScatterSpec,
	scatterer splitAndScatterer,
	doneScatterCh chan<- entryNode,
) error {
	__antithesis_instrumentation__.Notify(13193)
	g := ctxgroup.WithContext(ctx)

	importSpanChunksCh := make(chan scatteredChunk)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(13196)

		defer close(importSpanChunksCh)
		for i, importSpanChunk := range spec.Chunks {
			__antithesis_instrumentation__.Notify(13198)
			scatterKey := importSpanChunk.Entries[0].Span.Key
			if i+1 < len(spec.Chunks) {
				__antithesis_instrumentation__.Notify(13201)

				splitKey := spec.Chunks[i+1].Entries[0].Span.Key
				if err := scatterer.split(ctx, flowCtx.Codec(), splitKey); err != nil {
					__antithesis_instrumentation__.Notify(13202)
					return err
				} else {
					__antithesis_instrumentation__.Notify(13203)
				}
			} else {
				__antithesis_instrumentation__.Notify(13204)
			}
			__antithesis_instrumentation__.Notify(13199)
			chunkDestination, err := scatterer.scatter(ctx, flowCtx.Codec(), scatterKey)
			if err != nil {
				__antithesis_instrumentation__.Notify(13205)
				return err
			} else {
				__antithesis_instrumentation__.Notify(13206)
			}
			__antithesis_instrumentation__.Notify(13200)

			sc := scatteredChunk{
				destination: chunkDestination,
				entries:     importSpanChunk.Entries,
			}

			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(13207)
				return ctx.Err()
			case importSpanChunksCh <- sc:
				__antithesis_instrumentation__.Notify(13208)
			}
		}
		__antithesis_instrumentation__.Notify(13197)
		return nil
	})
	__antithesis_instrumentation__.Notify(13194)

	splitScatterWorkers := 2
	for worker := 0; worker < splitScatterWorkers; worker++ {
		__antithesis_instrumentation__.Notify(13209)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(13210)
			for importSpanChunk := range importSpanChunksCh {
				__antithesis_instrumentation__.Notify(13212)
				chunkDestination := importSpanChunk.destination
				for i, importEntry := range importSpanChunk.entries {
					__antithesis_instrumentation__.Notify(13213)
					nextChunkIdx := i + 1

					log.VInfof(ctx, 2, "processing a span [%s,%s)", importEntry.Span.Key, importEntry.Span.EndKey)
					var splitKey roachpb.Key
					if nextChunkIdx < len(importSpanChunk.entries) {
						__antithesis_instrumentation__.Notify(13215)

						splitKey = importSpanChunk.entries[nextChunkIdx].Span.Key
						if err := scatterer.split(ctx, flowCtx.Codec(), splitKey); err != nil {
							__antithesis_instrumentation__.Notify(13216)
							return err
						} else {
							__antithesis_instrumentation__.Notify(13217)
						}
					} else {
						__antithesis_instrumentation__.Notify(13218)
					}
					__antithesis_instrumentation__.Notify(13214)

					scatteredEntry := entryNode{
						entry: importEntry,
						node:  chunkDestination,
					}

					select {
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(13219)
						return ctx.Err()
					case doneScatterCh <- scatteredEntry:
						__antithesis_instrumentation__.Notify(13220)
					}
				}
			}
			__antithesis_instrumentation__.Notify(13211)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(13195)

	return g.Wait()
}

func routingDatumsForSQLInstance(
	sqlInstanceID base.SQLInstanceID,
) (rowenc.EncDatum, rowenc.EncDatum) {
	__antithesis_instrumentation__.Notify(13221)
	routingBytes := roachpb.Key(fmt.Sprintf("node%d", sqlInstanceID))
	startDatum := rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(routingBytes)))
	endDatum := rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(routingBytes.Next())))
	return startDatum, endDatum
}

func routingSpanForSQLInstance(sqlInstanceID base.SQLInstanceID) ([]byte, []byte, error) {
	__antithesis_instrumentation__.Notify(13222)
	var alloc tree.DatumAlloc
	startDatum, endDatum := routingDatumsForSQLInstance(sqlInstanceID)

	startBytes, endBytes := make([]byte, 0), make([]byte, 0)
	startBytes, err := startDatum.Encode(splitAndScatterOutputTypes[0], &alloc, descpb.DatumEncoding_ASCENDING_KEY, startBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(13225)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(13226)
	}
	__antithesis_instrumentation__.Notify(13223)
	endBytes, err = endDatum.Encode(splitAndScatterOutputTypes[0], &alloc, descpb.DatumEncoding_ASCENDING_KEY, endBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(13227)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(13228)
	}
	__antithesis_instrumentation__.Notify(13224)
	return startBytes, endBytes, nil
}

func init() {
	rowexec.NewSplitAndScatterProcessor = newSplitAndScatterProcessor
}
