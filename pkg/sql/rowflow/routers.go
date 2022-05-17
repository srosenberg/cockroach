package rowflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
)

type router interface {
	execinfra.RowReceiver
	flowinfra.Startable
	init(ctx context.Context, flowCtx *execinfra.FlowCtx, types []*types.T)
}

func makeRouter(
	spec *execinfrapb.OutputRouterSpec, streams []execinfra.RowReceiver,
) (router, error) {
	__antithesis_instrumentation__.Notify(575944)
	if len(streams) == 0 {
		__antithesis_instrumentation__.Notify(575946)
		return nil, errors.Errorf("no streams in router")
	} else {
		__antithesis_instrumentation__.Notify(575947)
	}
	__antithesis_instrumentation__.Notify(575945)

	var rb routerBase
	rb.setupStreams(spec, streams)

	switch spec.Type {
	case execinfrapb.OutputRouterSpec_BY_HASH:
		__antithesis_instrumentation__.Notify(575948)
		return makeHashRouter(rb, spec.HashColumns)

	case execinfrapb.OutputRouterSpec_MIRROR:
		__antithesis_instrumentation__.Notify(575949)
		return makeMirrorRouter(rb)

	case execinfrapb.OutputRouterSpec_BY_RANGE:
		__antithesis_instrumentation__.Notify(575950)
		return makeRangeRouter(rb, spec.RangeRouterSpec)

	default:
		__antithesis_instrumentation__.Notify(575951)
		return nil, errors.Errorf("router type %s not supported", spec.Type)
	}
}

const routerRowBufSize = execinfra.RowChannelBufSize

type routerOutput struct {
	stream   execinfra.RowReceiver
	streamID execinfrapb.StreamID
	mu       struct {
		syncutil.Mutex

		cond         *sync.Cond
		streamStatus execinfra.ConsumerStatus

		metadataBuf []*execinfrapb.ProducerMetadata

		rowBuf                [routerRowBufSize]rowenc.EncDatumRow
		rowBufLeft, rowBufLen uint32

		rowContainer rowcontainer.DiskBackedRowContainer
		producerDone bool
	}

	stats execinfrapb.ComponentStats

	memoryMonitor, diskMonitor *mon.BytesMonitor

	rowAlloc         rowenc.EncDatumRowAlloc
	rowBufToPushFrom [routerRowBufSize]rowenc.EncDatumRow

	rowBufToPushFromMon *mon.BytesMonitor
	rowBufToPushFromAcc *mon.BoundAccount

	rowBufToPushFromRowSize [routerRowBufSize]int64
}

func (ro *routerOutput) addMetadataLocked(meta *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575952)

	ro.mu.metadataBuf = append(ro.mu.metadataBuf, meta)
}

func (ro *routerOutput) addRowLocked(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(575953)
	if ro.mu.streamStatus != execinfra.NeedMoreRows {
		__antithesis_instrumentation__.Notify(575956)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(575957)
	}
	__antithesis_instrumentation__.Notify(575954)
	if ro.mu.rowBufLen == routerRowBufSize {
		__antithesis_instrumentation__.Notify(575958)

		evictedRow := ro.mu.rowBuf[ro.mu.rowBufLeft]
		if err := ro.mu.rowContainer.AddRow(ctx, evictedRow); err != nil {
			__antithesis_instrumentation__.Notify(575960)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575961)
		}
		__antithesis_instrumentation__.Notify(575959)

		ro.mu.rowBufLeft = (ro.mu.rowBufLeft + 1) % routerRowBufSize
		ro.mu.rowBufLen--
	} else {
		__antithesis_instrumentation__.Notify(575962)
	}
	__antithesis_instrumentation__.Notify(575955)
	ro.mu.rowBuf[(ro.mu.rowBufLeft+ro.mu.rowBufLen)%routerRowBufSize] = row
	ro.mu.rowBufLen++
	return nil
}

func (ro *routerOutput) popRowsLocked(ctx context.Context) ([]rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(575963)
	n := 0

	addToRowBufToPushFrom := func(row rowenc.EncDatumRow) error {
		__antithesis_instrumentation__.Notify(575967)

		rowSize := int64(row.Size())
		delta := rowSize - ro.rowBufToPushFromRowSize[n]
		ro.rowBufToPushFromRowSize[n] = rowSize
		if err := ro.rowBufToPushFromAcc.Grow(ctx, delta); err != nil {
			__antithesis_instrumentation__.Notify(575969)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575970)
		}
		__antithesis_instrumentation__.Notify(575968)
		ro.rowBufToPushFrom[n] = row
		return nil
	}
	__antithesis_instrumentation__.Notify(575964)

	if ro.mu.rowContainer.Len() > 0 {
		__antithesis_instrumentation__.Notify(575971)
		if err := func() error {
			__antithesis_instrumentation__.Notify(575972)
			i := ro.mu.rowContainer.NewFinalIterator(ctx)
			defer i.Close()
			for i.Rewind(); n < len(ro.rowBufToPushFrom); i.Next() {
				__antithesis_instrumentation__.Notify(575974)
				if ok, err := i.Valid(); err != nil {
					__antithesis_instrumentation__.Notify(575978)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575979)
					if !ok {
						__antithesis_instrumentation__.Notify(575980)
						break
					} else {
						__antithesis_instrumentation__.Notify(575981)
					}
				}
				__antithesis_instrumentation__.Notify(575975)
				row, err := i.Row()
				if err != nil {
					__antithesis_instrumentation__.Notify(575982)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575983)
				}
				__antithesis_instrumentation__.Notify(575976)
				if err = addToRowBufToPushFrom(ro.rowAlloc.CopyRow(row)); err != nil {
					__antithesis_instrumentation__.Notify(575984)
					return err
				} else {
					__antithesis_instrumentation__.Notify(575985)
				}
				__antithesis_instrumentation__.Notify(575977)
				n++
			}
			__antithesis_instrumentation__.Notify(575973)
			return nil
		}(); err != nil {
			__antithesis_instrumentation__.Notify(575986)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575987)
		}
	} else {
		__antithesis_instrumentation__.Notify(575988)
	}
	__antithesis_instrumentation__.Notify(575965)

	for ; n < len(ro.rowBufToPushFrom) && func() bool {
		__antithesis_instrumentation__.Notify(575989)
		return ro.mu.rowBufLen > 0 == true
	}() == true; n++ {
		__antithesis_instrumentation__.Notify(575990)
		if err := addToRowBufToPushFrom(ro.mu.rowBuf[ro.mu.rowBufLeft]); err != nil {
			__antithesis_instrumentation__.Notify(575992)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575993)
		}
		__antithesis_instrumentation__.Notify(575991)
		ro.mu.rowBufLeft = (ro.mu.rowBufLeft + 1) % routerRowBufSize
		ro.mu.rowBufLen--
	}
	__antithesis_instrumentation__.Notify(575966)
	return ro.rowBufToPushFrom[:n], nil
}

const semaphorePeriod = 8

type routerBase struct {
	types []*types.T

	outputs []routerOutput

	numNonDrainingStreams int32

	aggregatedStatus uint32

	semaphore chan struct{}

	semaphoreCount int32

	statsCollectionEnabled bool
}

func (rb *routerBase) aggStatus() execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(575994)
	return execinfra.ConsumerStatus(atomic.LoadUint32(&rb.aggregatedStatus))
}

func (rb *routerBase) setupStreams(
	spec *execinfrapb.OutputRouterSpec, streams []execinfra.RowReceiver,
) {
	__antithesis_instrumentation__.Notify(575995)
	rb.numNonDrainingStreams = int32(len(streams))
	n := len(streams)
	if spec.DisableBuffering {
		__antithesis_instrumentation__.Notify(575997)

		n = 1

	} else {
		__antithesis_instrumentation__.Notify(575998)
	}
	__antithesis_instrumentation__.Notify(575996)
	rb.semaphore = make(chan struct{}, n)
	rb.outputs = make([]routerOutput, len(streams))
	for i := range rb.outputs {
		__antithesis_instrumentation__.Notify(575999)
		ro := &rb.outputs[i]
		ro.stream = streams[i]
		ro.streamID = spec.Streams[i].StreamID
		ro.mu.cond = sync.NewCond(&ro.mu.Mutex)
		ro.mu.streamStatus = execinfra.NeedMoreRows
	}
}

func (rb *routerBase) init(ctx context.Context, flowCtx *execinfra.FlowCtx, types []*types.T) {

	if s := tracing.SpanFromContext(ctx); s != nil && s.IsVerbose() {
		rb.statsCollectionEnabled = true
	}

	rb.types = types
	for i := range rb.outputs {

		evalCtx := flowCtx.NewEvalCtx()
		rb.outputs[i].memoryMonitor = execinfra.NewLimitedMonitor(
			ctx, evalCtx.Mon, flowCtx,
			fmt.Sprintf("router-limited-%d", rb.outputs[i].streamID),
		)
		rb.outputs[i].diskMonitor = execinfra.NewMonitor(
			ctx, flowCtx.DiskMonitor,
			fmt.Sprintf("router-disk-%d", rb.outputs[i].streamID),
		)

		rb.outputs[i].rowBufToPushFromMon = execinfra.NewMonitor(
			ctx, evalCtx.Mon, fmt.Sprintf("router-unlimited-%d", rb.outputs[i].streamID),
		)
		memAcc := rb.outputs[i].rowBufToPushFromMon.MakeBoundAccount()
		rb.outputs[i].rowBufToPushFromAcc = &memAcc

		rb.outputs[i].mu.rowContainer.Init(
			nil,
			types,
			evalCtx,
			flowCtx.Cfg.TempStorage,
			rb.outputs[i].memoryMonitor,
			rb.outputs[i].diskMonitor,
		)

		if o, ok := rb.outputs[i].stream.(*flowinfra.Outbox); ok {
			o.Init(types)
		}
	}
}

func (rb *routerBase) Start(ctx context.Context, wg *sync.WaitGroup, _ context.CancelFunc) {
	__antithesis_instrumentation__.Notify(576000)
	wg.Add(len(rb.outputs))
	for i := range rb.outputs {
		__antithesis_instrumentation__.Notify(576001)
		go func(ctx context.Context, rb *routerBase, ro *routerOutput, wg *sync.WaitGroup) {
			__antithesis_instrumentation__.Notify(576002)
			var span *tracing.Span
			if rb.statsCollectionEnabled {
				__antithesis_instrumentation__.Notify(576005)
				ctx, span = execinfra.ProcessorSpan(ctx, "router output")
				defer span.Finish()
				span.SetTag(execinfrapb.StreamIDTagKey, attribute.IntValue(int(ro.streamID)))
				ro.stats.Inputs = make([]execinfrapb.InputStats, 1)
			} else {
				__antithesis_instrumentation__.Notify(576006)
			}
			__antithesis_instrumentation__.Notify(576003)

			drain := false
			streamStatus := execinfra.NeedMoreRows
			ro.mu.Lock()
			for {
				__antithesis_instrumentation__.Notify(576007)

				if len(ro.mu.metadataBuf) > 0 {
					__antithesis_instrumentation__.Notify(576011)
					m := ro.mu.metadataBuf[0]

					ro.mu.metadataBuf[0] = nil
					ro.mu.metadataBuf = ro.mu.metadataBuf[1:]

					ro.mu.Unlock()

					rb.semaphore <- struct{}{}
					status := ro.stream.Push(nil, m)
					<-rb.semaphore

					rb.updateStreamState(&streamStatus, status)
					ro.mu.Lock()
					ro.mu.streamStatus = streamStatus
					continue
				} else {
					__antithesis_instrumentation__.Notify(576012)
				}
				__antithesis_instrumentation__.Notify(576008)

				if !drain {
					__antithesis_instrumentation__.Notify(576013)

					if rows, err := ro.popRowsLocked(ctx); err != nil {
						__antithesis_instrumentation__.Notify(576014)
						ro.mu.Unlock()
						rb.fwdMetadata(&execinfrapb.ProducerMetadata{Err: err})
						ro.mu.Lock()
						atomic.StoreUint32(&rb.aggregatedStatus, uint32(execinfra.DrainRequested))
						drain = true
						continue
					} else {
						__antithesis_instrumentation__.Notify(576015)
						if len(rows) > 0 {
							__antithesis_instrumentation__.Notify(576016)
							ro.mu.Unlock()
							rb.semaphore <- struct{}{}
							for _, row := range rows {
								__antithesis_instrumentation__.Notify(576019)
								status := ro.stream.Push(row, nil)
								rb.updateStreamState(&streamStatus, status)
							}
							__antithesis_instrumentation__.Notify(576017)
							<-rb.semaphore
							if rb.statsCollectionEnabled {
								__antithesis_instrumentation__.Notify(576020)
								ro.stats.Inputs[0].NumTuples.Add(int64(len(rows)))
							} else {
								__antithesis_instrumentation__.Notify(576021)
							}
							__antithesis_instrumentation__.Notify(576018)
							ro.mu.Lock()
							ro.mu.streamStatus = streamStatus
							continue
						} else {
							__antithesis_instrumentation__.Notify(576022)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(576023)
				}
				__antithesis_instrumentation__.Notify(576009)

				if ro.mu.producerDone {
					__antithesis_instrumentation__.Notify(576024)
					if rb.statsCollectionEnabled {
						__antithesis_instrumentation__.Notify(576026)
						ro.stats.Exec.MaxAllocatedMem.Set(uint64(ro.memoryMonitor.MaximumBytes()))
						ro.stats.Exec.MaxAllocatedDisk.Set(uint64(ro.diskMonitor.MaximumBytes()))
						span.RecordStructured(&ro.stats)
						if meta := execinfra.GetTraceDataAsMetadata(span); meta != nil {
							__antithesis_instrumentation__.Notify(576027)
							ro.mu.Unlock()
							rb.semaphore <- struct{}{}
							status := ro.stream.Push(nil, meta)
							rb.updateStreamState(&streamStatus, status)
							<-rb.semaphore
							ro.mu.Lock()
						} else {
							__antithesis_instrumentation__.Notify(576028)
						}
					} else {
						__antithesis_instrumentation__.Notify(576029)
					}
					__antithesis_instrumentation__.Notify(576025)
					ro.stream.ProducerDone()
					break
				} else {
					__antithesis_instrumentation__.Notify(576030)
				}
				__antithesis_instrumentation__.Notify(576010)

				ro.mu.cond.Wait()
			}
			__antithesis_instrumentation__.Notify(576004)
			ro.mu.rowContainer.Close(ctx)
			ro.mu.Unlock()

			ro.rowBufToPushFromAcc.Close(ctx)
			ro.memoryMonitor.Stop(ctx)
			ro.diskMonitor.Stop(ctx)
			ro.rowBufToPushFromMon.Stop(ctx)

			wg.Done()
		}(ctx, rb, &rb.outputs[i], wg)
	}
}

func (rb *routerBase) ProducerDone() {
	__antithesis_instrumentation__.Notify(576031)
	for i := range rb.outputs {
		__antithesis_instrumentation__.Notify(576032)
		o := &rb.outputs[i]
		o.mu.Lock()
		o.mu.producerDone = true
		o.mu.Unlock()
		o.mu.cond.Signal()
	}
}

func (rb *routerBase) updateStreamState(
	streamStatus *execinfra.ConsumerStatus, newState execinfra.ConsumerStatus,
) {
	__antithesis_instrumentation__.Notify(576033)
	if newState != *streamStatus {
		__antithesis_instrumentation__.Notify(576034)
		if *streamStatus == execinfra.NeedMoreRows {
			__antithesis_instrumentation__.Notify(576036)

			if atomic.AddInt32(&rb.numNonDrainingStreams, -1) == 0 {
				__antithesis_instrumentation__.Notify(576037)

				atomic.CompareAndSwapUint32(
					&rb.aggregatedStatus,
					uint32(execinfra.NeedMoreRows),
					uint32(execinfra.DrainRequested),
				)
			} else {
				__antithesis_instrumentation__.Notify(576038)
			}
		} else {
			__antithesis_instrumentation__.Notify(576039)
		}
		__antithesis_instrumentation__.Notify(576035)
		*streamStatus = newState
	} else {
		__antithesis_instrumentation__.Notify(576040)
	}
}

func (rb *routerBase) fwdMetadata(meta *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(576041)
	if meta == nil {
		__antithesis_instrumentation__.Notify(576045)
		log.Fatalf(context.TODO(), "asked to fwd empty metadata")
		return
	} else {
		__antithesis_instrumentation__.Notify(576046)
	}
	__antithesis_instrumentation__.Notify(576042)

	rb.semaphore <- struct{}{}
	defer func() {
		__antithesis_instrumentation__.Notify(576047)
		<-rb.semaphore
	}()
	__antithesis_instrumentation__.Notify(576043)
	if metaErr := meta.Err; metaErr != nil {
		__antithesis_instrumentation__.Notify(576048)

		if rb.fwdErrMetadata(metaErr) {
			__antithesis_instrumentation__.Notify(576049)
			return
		} else {
			__antithesis_instrumentation__.Notify(576050)
		}
	} else {
		__antithesis_instrumentation__.Notify(576051)

		for i := range rb.outputs {
			__antithesis_instrumentation__.Notify(576052)
			ro := &rb.outputs[i]
			ro.mu.Lock()
			if ro.mu.streamStatus != execinfra.ConsumerClosed {
				__antithesis_instrumentation__.Notify(576054)
				ro.addMetadataLocked(meta)
				ro.mu.Unlock()
				ro.mu.cond.Signal()
				return
			} else {
				__antithesis_instrumentation__.Notify(576055)
			}
			__antithesis_instrumentation__.Notify(576053)
			ro.mu.Unlock()
		}
	}
	__antithesis_instrumentation__.Notify(576044)

	atomic.StoreUint32(&rb.aggregatedStatus, uint32(execinfra.ConsumerClosed))
}

func (rb *routerBase) fwdErrMetadata(err error) bool {
	__antithesis_instrumentation__.Notify(576056)
	forwarded := false
	for i := range rb.outputs {
		__antithesis_instrumentation__.Notify(576058)
		ro := &rb.outputs[i]
		ro.mu.Lock()
		if ro.mu.streamStatus != execinfra.ConsumerClosed {
			__antithesis_instrumentation__.Notify(576059)
			meta := &execinfrapb.ProducerMetadata{Err: err}
			ro.addMetadataLocked(meta)
			ro.mu.Unlock()
			ro.mu.cond.Signal()
			forwarded = true
		} else {
			__antithesis_instrumentation__.Notify(576060)
			ro.mu.Unlock()
		}
	}
	__antithesis_instrumentation__.Notify(576057)
	return forwarded
}

func (rb *routerBase) shouldUseSemaphore() bool {
	__antithesis_instrumentation__.Notify(576061)
	rb.semaphoreCount++
	if rb.semaphoreCount >= semaphorePeriod {
		__antithesis_instrumentation__.Notify(576063)
		rb.semaphoreCount = 0
		return true
	} else {
		__antithesis_instrumentation__.Notify(576064)
	}
	__antithesis_instrumentation__.Notify(576062)
	return false
}

type mirrorRouter struct {
	routerBase
}

type hashRouter struct {
	routerBase

	hashCols []uint32
	buffer   []byte
	alloc    tree.DatumAlloc
}

type rangeRouter struct {
	routerBase

	alloc tree.DatumAlloc

	b         []byte
	encodings []execinfrapb.OutputRouterSpec_RangeRouterSpec_ColumnEncoding
	spans     []execinfrapb.OutputRouterSpec_RangeRouterSpec_Span

	defaultDest *int
}

var _ execinfra.RowReceiver = &mirrorRouter{}
var _ execinfra.RowReceiver = &hashRouter{}
var _ execinfra.RowReceiver = &rangeRouter{}

func makeMirrorRouter(rb routerBase) (router, error) {
	__antithesis_instrumentation__.Notify(576065)
	if len(rb.outputs) < 2 {
		__antithesis_instrumentation__.Notify(576067)
		return nil, errors.Errorf("need at least two streams for mirror router")
	} else {
		__antithesis_instrumentation__.Notify(576068)
	}
	__antithesis_instrumentation__.Notify(576066)
	return &mirrorRouter{routerBase: rb}, nil
}

func (mr *mirrorRouter) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(576069)
	aggStatus := mr.aggStatus()
	if meta != nil {
		__antithesis_instrumentation__.Notify(576075)
		mr.fwdMetadata(meta)

		return mr.aggStatus()
	} else {
		__antithesis_instrumentation__.Notify(576076)
	}
	__antithesis_instrumentation__.Notify(576070)
	if aggStatus != execinfra.NeedMoreRows {
		__antithesis_instrumentation__.Notify(576077)
		return aggStatus
	} else {
		__antithesis_instrumentation__.Notify(576078)
	}
	__antithesis_instrumentation__.Notify(576071)

	useSema := mr.shouldUseSemaphore()
	if useSema {
		__antithesis_instrumentation__.Notify(576079)
		mr.semaphore <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(576080)
	}
	__antithesis_instrumentation__.Notify(576072)

	for i := range mr.outputs {
		__antithesis_instrumentation__.Notify(576081)
		ro := &mr.outputs[i]
		ro.mu.Lock()
		err := ro.addRowLocked(context.TODO(), row)
		ro.mu.Unlock()
		if err != nil {
			__antithesis_instrumentation__.Notify(576083)
			if useSema {
				__antithesis_instrumentation__.Notify(576085)
				<-mr.semaphore
			} else {
				__antithesis_instrumentation__.Notify(576086)
			}
			__antithesis_instrumentation__.Notify(576084)
			mr.fwdMetadata(&execinfrapb.ProducerMetadata{Err: err})
			atomic.StoreUint32(&mr.aggregatedStatus, uint32(execinfra.ConsumerClosed))
			return execinfra.ConsumerClosed
		} else {
			__antithesis_instrumentation__.Notify(576087)
		}
		__antithesis_instrumentation__.Notify(576082)
		ro.mu.cond.Signal()
	}
	__antithesis_instrumentation__.Notify(576073)
	if useSema {
		__antithesis_instrumentation__.Notify(576088)
		<-mr.semaphore
	} else {
		__antithesis_instrumentation__.Notify(576089)
	}
	__antithesis_instrumentation__.Notify(576074)
	return aggStatus
}

var crc32Table = crc32.MakeTable(crc32.Castagnoli)

func makeHashRouter(rb routerBase, hashCols []uint32) (router, error) {
	__antithesis_instrumentation__.Notify(576090)
	if len(rb.outputs) < 2 {
		__antithesis_instrumentation__.Notify(576093)
		return nil, errors.Errorf("need at least two streams for hash router")
	} else {
		__antithesis_instrumentation__.Notify(576094)
	}
	__antithesis_instrumentation__.Notify(576091)
	if len(hashCols) == 0 {
		__antithesis_instrumentation__.Notify(576095)
		return nil, errors.Errorf("no hash columns for BY_HASH router")
	} else {
		__antithesis_instrumentation__.Notify(576096)
	}
	__antithesis_instrumentation__.Notify(576092)
	return &hashRouter{hashCols: hashCols, routerBase: rb}, nil
}

func (hr *hashRouter) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(576097)
	aggStatus := hr.aggStatus()
	if meta != nil {
		__antithesis_instrumentation__.Notify(576104)
		hr.fwdMetadata(meta)

		return hr.aggStatus()
	} else {
		__antithesis_instrumentation__.Notify(576105)
	}
	__antithesis_instrumentation__.Notify(576098)
	if aggStatus != execinfra.NeedMoreRows {
		__antithesis_instrumentation__.Notify(576106)
		return aggStatus
	} else {
		__antithesis_instrumentation__.Notify(576107)
	}
	__antithesis_instrumentation__.Notify(576099)

	useSema := hr.shouldUseSemaphore()
	if useSema {
		__antithesis_instrumentation__.Notify(576108)
		hr.semaphore <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(576109)
	}
	__antithesis_instrumentation__.Notify(576100)

	streamIdx, err := hr.computeDestination(row)
	if err == nil {
		__antithesis_instrumentation__.Notify(576110)
		ro := &hr.outputs[streamIdx]
		ro.mu.Lock()
		err = ro.addRowLocked(context.TODO(), row)
		ro.mu.Unlock()
		ro.mu.cond.Signal()
	} else {
		__antithesis_instrumentation__.Notify(576111)
	}
	__antithesis_instrumentation__.Notify(576101)
	if useSema {
		__antithesis_instrumentation__.Notify(576112)
		<-hr.semaphore
	} else {
		__antithesis_instrumentation__.Notify(576113)
	}
	__antithesis_instrumentation__.Notify(576102)
	if err != nil {
		__antithesis_instrumentation__.Notify(576114)
		hr.fwdMetadata(&execinfrapb.ProducerMetadata{Err: err})
		atomic.StoreUint32(&hr.aggregatedStatus, uint32(execinfra.ConsumerClosed))
		return execinfra.ConsumerClosed
	} else {
		__antithesis_instrumentation__.Notify(576115)
	}
	__antithesis_instrumentation__.Notify(576103)
	return aggStatus
}

func (hr *hashRouter) computeDestination(row rowenc.EncDatumRow) (int, error) {
	__antithesis_instrumentation__.Notify(576116)
	hr.buffer = hr.buffer[:0]
	for _, col := range hr.hashCols {
		__antithesis_instrumentation__.Notify(576118)
		if int(col) >= len(row) {
			__antithesis_instrumentation__.Notify(576120)
			err := errors.Errorf("hash column %d, row with only %d columns", col, len(row))
			return -1, err
		} else {
			__antithesis_instrumentation__.Notify(576121)
		}
		__antithesis_instrumentation__.Notify(576119)
		var err error

		hr.buffer, err = row[col].Fingerprint(context.TODO(), hr.types[col], &hr.alloc, hr.buffer, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(576122)
			return -1, err
		} else {
			__antithesis_instrumentation__.Notify(576123)
		}
	}
	__antithesis_instrumentation__.Notify(576117)

	return int(crc32.Update(0, crc32Table, hr.buffer) % uint32(len(hr.outputs))), nil
}

func makeRangeRouter(
	rb routerBase, spec execinfrapb.OutputRouterSpec_RangeRouterSpec,
) (*rangeRouter, error) {
	__antithesis_instrumentation__.Notify(576124)
	if len(spec.Encodings) == 0 {
		__antithesis_instrumentation__.Notify(576128)
		return nil, errors.New("missing encodings")
	} else {
		__antithesis_instrumentation__.Notify(576129)
	}
	__antithesis_instrumentation__.Notify(576125)
	var defaultDest *int
	if spec.DefaultDest != nil {
		__antithesis_instrumentation__.Notify(576130)
		i := int(*spec.DefaultDest)
		defaultDest = &i
	} else {
		__antithesis_instrumentation__.Notify(576131)
	}
	__antithesis_instrumentation__.Notify(576126)
	var prevKey []byte

	for i, span := range spec.Spans {
		__antithesis_instrumentation__.Notify(576132)
		if bytes.Compare(prevKey, span.Start) > 0 {
			__antithesis_instrumentation__.Notify(576134)
			return nil, errors.Errorf("span %d not after previous span", i)
		} else {
			__antithesis_instrumentation__.Notify(576135)
		}
		__antithesis_instrumentation__.Notify(576133)
		prevKey = span.End
	}
	__antithesis_instrumentation__.Notify(576127)
	return &rangeRouter{
		routerBase:  rb,
		spans:       spec.Spans,
		defaultDest: defaultDest,
		encodings:   spec.Encodings,
	}, nil
}

func (rr *rangeRouter) Push(
	row rowenc.EncDatumRow, meta *execinfrapb.ProducerMetadata,
) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(576136)
	aggStatus := rr.aggStatus()
	if meta != nil {
		__antithesis_instrumentation__.Notify(576142)
		rr.fwdMetadata(meta)

		return rr.aggStatus()
	} else {
		__antithesis_instrumentation__.Notify(576143)
	}
	__antithesis_instrumentation__.Notify(576137)

	useSema := rr.shouldUseSemaphore()
	if useSema {
		__antithesis_instrumentation__.Notify(576144)
		rr.semaphore <- struct{}{}
	} else {
		__antithesis_instrumentation__.Notify(576145)
	}
	__antithesis_instrumentation__.Notify(576138)

	streamIdx, err := rr.computeDestination(row)
	if err == nil {
		__antithesis_instrumentation__.Notify(576146)
		ro := &rr.outputs[streamIdx]
		ro.mu.Lock()
		err = ro.addRowLocked(context.TODO(), row)
		ro.mu.Unlock()
		ro.mu.cond.Signal()
	} else {
		__antithesis_instrumentation__.Notify(576147)
	}
	__antithesis_instrumentation__.Notify(576139)
	if useSema {
		__antithesis_instrumentation__.Notify(576148)
		<-rr.semaphore
	} else {
		__antithesis_instrumentation__.Notify(576149)
	}
	__antithesis_instrumentation__.Notify(576140)
	if err != nil {
		__antithesis_instrumentation__.Notify(576150)
		rr.fwdMetadata(&execinfrapb.ProducerMetadata{Err: err})
		atomic.StoreUint32(&rr.aggregatedStatus, uint32(execinfra.ConsumerClosed))
		return execinfra.ConsumerClosed
	} else {
		__antithesis_instrumentation__.Notify(576151)
	}
	__antithesis_instrumentation__.Notify(576141)
	return aggStatus
}

func (rr *rangeRouter) computeDestination(row rowenc.EncDatumRow) (int, error) {
	__antithesis_instrumentation__.Notify(576152)
	var err error
	rr.b = rr.b[:0]
	for _, enc := range rr.encodings {
		__antithesis_instrumentation__.Notify(576155)
		col := enc.Column
		rr.b, err = row[col].Encode(rr.types[col], &rr.alloc, enc.Encoding, rr.b)
		if err != nil {
			__antithesis_instrumentation__.Notify(576156)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(576157)
		}
	}
	__antithesis_instrumentation__.Notify(576153)
	i := rr.spanForData(rr.b)
	if i == -1 {
		__antithesis_instrumentation__.Notify(576158)
		if rr.defaultDest == nil {
			__antithesis_instrumentation__.Notify(576160)
			return 0, errors.New("no span found for key")
		} else {
			__antithesis_instrumentation__.Notify(576161)
		}
		__antithesis_instrumentation__.Notify(576159)
		return *rr.defaultDest, nil
	} else {
		__antithesis_instrumentation__.Notify(576162)
	}
	__antithesis_instrumentation__.Notify(576154)
	return i, nil
}

func (rr *rangeRouter) spanForData(data []byte) int {
	__antithesis_instrumentation__.Notify(576163)
	i := sort.Search(len(rr.spans), func(i int) bool {
		__antithesis_instrumentation__.Notify(576167)
		return bytes.Compare(rr.spans[i].End, data) > 0
	})
	__antithesis_instrumentation__.Notify(576164)

	if i == len(rr.spans) {
		__antithesis_instrumentation__.Notify(576168)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(576169)
	}
	__antithesis_instrumentation__.Notify(576165)

	if bytes.Compare(rr.spans[i].Start, data) > 0 {
		__antithesis_instrumentation__.Notify(576170)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(576171)
	}
	__antithesis_instrumentation__.Notify(576166)
	return int(rr.spans[i].Stream)
}
