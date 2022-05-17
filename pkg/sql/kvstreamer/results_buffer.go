package kvstreamer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type resultsBuffer interface {
	init(_ context.Context, numExpectedResponses int) error

	get(context.Context) (_ []Result, allComplete bool, _ error)

	wait()

	releaseOne()

	close(context.Context)

	add([]Result)

	spill(_ context.Context, atLeastBytes int64, spillingPriority int) (bool, error)

	error() error

	numUnreleased() int

	setError(error)
}

type resultsBufferBase struct {
	budget *budget
	syncutil.Mutex

	numExpectedResponses int

	numCompleteResponses int

	numUnreleasedResults int

	hasResults chan struct{}
	err        error
}

func newResultsBufferBase(budget *budget) *resultsBufferBase {
	__antithesis_instrumentation__.Notify(498928)
	return &resultsBufferBase{
		budget:     budget,
		hasResults: make(chan struct{}, 1),
	}
}

func (b *resultsBufferBase) initLocked(isEmpty bool, numExpectedResponses int) error {
	__antithesis_instrumentation__.Notify(498929)
	b.Mutex.AssertHeld()
	if b.numExpectedResponses != b.numCompleteResponses {
		__antithesis_instrumentation__.Notify(498933)
		b.setErrorLocked(errors.AssertionFailedf("Enqueue is called before the previous requests have been completed"))
		return b.err
	} else {
		__antithesis_instrumentation__.Notify(498934)
	}
	__antithesis_instrumentation__.Notify(498930)
	if !isEmpty {
		__antithesis_instrumentation__.Notify(498935)
		b.setErrorLocked(errors.AssertionFailedf("Enqueue is called before the results of the previous requests have been retrieved"))
		return b.err
	} else {
		__antithesis_instrumentation__.Notify(498936)
	}
	__antithesis_instrumentation__.Notify(498931)
	if b.numUnreleasedResults > 0 {
		__antithesis_instrumentation__.Notify(498937)
		b.setErrorLocked(errors.AssertionFailedf("unexpectedly there are some unreleased Results"))
		return b.err
	} else {
		__antithesis_instrumentation__.Notify(498938)
	}
	__antithesis_instrumentation__.Notify(498932)
	b.numExpectedResponses = numExpectedResponses
	b.numCompleteResponses = 0
	return nil
}

func (b *resultsBufferBase) findCompleteResponses(results []Result) {
	__antithesis_instrumentation__.Notify(498939)
	for i := range results {
		__antithesis_instrumentation__.Notify(498940)
		if results[i].GetResp != nil || func() bool {
			__antithesis_instrumentation__.Notify(498941)
			return results[i].ScanResp.Complete == true
		}() == true {
			__antithesis_instrumentation__.Notify(498942)
			b.numCompleteResponses++
		} else {
			__antithesis_instrumentation__.Notify(498943)
		}
	}
}

func (b *resultsBufferBase) signal() {
	__antithesis_instrumentation__.Notify(498944)
	select {
	case b.hasResults <- struct{}{}:
		__antithesis_instrumentation__.Notify(498945)
	default:
		__antithesis_instrumentation__.Notify(498946)
	}
}

func (b *resultsBufferBase) wait() {
	__antithesis_instrumentation__.Notify(498947)
	<-b.hasResults
}

func (b *resultsBufferBase) numUnreleased() int {
	__antithesis_instrumentation__.Notify(498948)
	b.Lock()
	defer b.Unlock()
	return b.numUnreleasedResults
}

func (b *resultsBufferBase) releaseOne() {
	__antithesis_instrumentation__.Notify(498949)
	b.Lock()
	defer b.Unlock()
	b.numUnreleasedResults--
}

func (b *resultsBufferBase) setError(err error) {
	__antithesis_instrumentation__.Notify(498950)
	b.Lock()
	defer b.Unlock()
	b.setErrorLocked(err)
}

func (b *resultsBufferBase) setErrorLocked(err error) {
	__antithesis_instrumentation__.Notify(498951)
	b.Mutex.AssertHeld()
	if b.err == nil {
		__antithesis_instrumentation__.Notify(498953)
		b.err = err
	} else {
		__antithesis_instrumentation__.Notify(498954)
	}
	__antithesis_instrumentation__.Notify(498952)
	b.signal()
}

func (b *resultsBufferBase) error() error {
	__antithesis_instrumentation__.Notify(498955)
	b.Lock()
	defer b.Unlock()
	return b.err
}

func resultsToString(results []Result) string {
	__antithesis_instrumentation__.Notify(498956)
	result := "results for positions "
	for i, r := range results {
		__antithesis_instrumentation__.Notify(498958)
		if i > 0 {
			__antithesis_instrumentation__.Notify(498960)
			result += ", "
		} else {
			__antithesis_instrumentation__.Notify(498961)
		}
		__antithesis_instrumentation__.Notify(498959)
		result += fmt.Sprintf("%d", r.position)
	}
	__antithesis_instrumentation__.Notify(498957)
	return result
}

type outOfOrderResultsBuffer struct {
	*resultsBufferBase
	results []Result
}

var _ resultsBuffer = &outOfOrderResultsBuffer{}

func newOutOfOrderResultsBuffer(budget *budget) resultsBuffer {
	__antithesis_instrumentation__.Notify(498962)
	return &outOfOrderResultsBuffer{resultsBufferBase: newResultsBufferBase(budget)}
}

func (b *outOfOrderResultsBuffer) init(_ context.Context, numExpectedResponses int) error {
	b.Lock()
	defer b.Unlock()
	if err := b.initLocked(len(b.results) == 0, numExpectedResponses); err != nil {
		b.setErrorLocked(err)
		return err
	}
	return nil
}

func (b *outOfOrderResultsBuffer) add(results []Result) {
	__antithesis_instrumentation__.Notify(498963)
	b.Lock()
	defer b.Unlock()
	b.results = append(b.results, results...)
	b.findCompleteResponses(results)
	b.numUnreleasedResults += len(results)
	b.signal()
}

func (b *outOfOrderResultsBuffer) get(context.Context) ([]Result, bool, error) {
	__antithesis_instrumentation__.Notify(498964)
	b.Lock()
	defer b.Unlock()
	results := b.results
	b.results = nil
	allComplete := b.numCompleteResponses == b.numExpectedResponses
	return results, allComplete, b.err
}

func (b *outOfOrderResultsBuffer) spill(context.Context, int64, int) (bool, error) {
	__antithesis_instrumentation__.Notify(498965)

	b.budget.mu.AssertHeld()
	return false, nil
}

func (b *outOfOrderResultsBuffer) close(context.Context) {
	__antithesis_instrumentation__.Notify(498966)

	b.signal()
}

type inOrderResultsBuffer struct {
	*resultsBufferBase

	headOfLinePosition int

	buffered []inOrderBufferedResult
	disk     struct {
		initialized bool

		container rowcontainer.DiskRowContainer

		iter      rowcontainer.RowIterator
		iterRowID int

		engine     diskmap.Factory
		monitor    *mon.BytesMonitor
		rowScratch rowenc.EncDatumRow
		alloc      tree.DatumAlloc
	}
}

var _ resultsBuffer = &inOrderResultsBuffer{}
var _ heap.Interface = &inOrderResultsBuffer{}

func newInOrderResultsBuffer(
	budget *budget, engine diskmap.Factory, diskMonitor *mon.BytesMonitor,
) resultsBuffer {
	__antithesis_instrumentation__.Notify(498967)
	if engine == nil || func() bool {
		__antithesis_instrumentation__.Notify(498969)
		return diskMonitor == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(498970)
		panic(errors.AssertionFailedf("either engine or diskMonitor is nil"))
	} else {
		__antithesis_instrumentation__.Notify(498971)
	}
	__antithesis_instrumentation__.Notify(498968)
	b := &inOrderResultsBuffer{resultsBufferBase: newResultsBufferBase(budget)}
	b.disk.engine = engine
	b.disk.monitor = diskMonitor
	return b
}

func (b *inOrderResultsBuffer) Len() int {
	__antithesis_instrumentation__.Notify(498972)
	return len(b.buffered)
}

func (b *inOrderResultsBuffer) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(498973)
	return b.buffered[i].position < b.buffered[j].position
}

func (b *inOrderResultsBuffer) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(498974)
	b.buffered[i], b.buffered[j] = b.buffered[j], b.buffered[i]
}

func (b *inOrderResultsBuffer) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(498975)
	b.buffered = append(b.buffered, x.(inOrderBufferedResult))
}

func (b *inOrderResultsBuffer) Pop() interface{} {
	__antithesis_instrumentation__.Notify(498976)
	x := b.buffered[len(b.buffered)-1]
	b.buffered = b.buffered[:len(b.buffered)-1]
	return x
}

func (b *inOrderResultsBuffer) init(ctx context.Context, numExpectedResponses int) error {
	b.Lock()
	defer b.Unlock()
	if err := b.initLocked(len(b.buffered) == 0, numExpectedResponses); err != nil {
		b.setErrorLocked(err)
		return err
	}
	b.headOfLinePosition = 0
	if b.disk.initialized {

		if b.disk.iter != nil {
			b.disk.iter.Close()
			b.disk.iter = nil
			b.disk.iterRowID = 0
		}
		if err := b.disk.container.UnsafeReset(ctx); err != nil {
			b.setErrorLocked(err)
			return err
		}
	}
	return nil
}

func (b *inOrderResultsBuffer) add(results []Result) {
	__antithesis_instrumentation__.Notify(498977)
	b.Lock()
	defer b.Unlock()

	b.findCompleteResponses(results)
	foundHeadOfLine := false
	for _, r := range results {
		__antithesis_instrumentation__.Notify(498979)
		if debug {
			__antithesis_instrumentation__.Notify(498981)
			fmt.Printf("adding a result for position %d of size %d\n", r.position, r.memoryTok.toRelease)
		} else {
			__antithesis_instrumentation__.Notify(498982)
		}
		__antithesis_instrumentation__.Notify(498980)

		heap.Push(b, inOrderBufferedResult{Result: r, onDisk: false})
		if r.position == b.headOfLinePosition {
			__antithesis_instrumentation__.Notify(498983)
			foundHeadOfLine = true
		} else {
			__antithesis_instrumentation__.Notify(498984)
		}
	}
	__antithesis_instrumentation__.Notify(498978)
	if foundHeadOfLine {
		__antithesis_instrumentation__.Notify(498985)
		if debug {
			__antithesis_instrumentation__.Notify(498987)
			fmt.Println("found head-of-the-line")
		} else {
			__antithesis_instrumentation__.Notify(498988)
		}
		__antithesis_instrumentation__.Notify(498986)
		b.signal()
	} else {
		__antithesis_instrumentation__.Notify(498989)
	}
}

func (b *inOrderResultsBuffer) get(ctx context.Context) ([]Result, bool, error) {
	__antithesis_instrumentation__.Notify(498990)

	b.budget.mu.Lock()
	defer b.budget.mu.Unlock()
	b.Lock()
	defer b.Unlock()
	var res []Result
	if debug {
		__antithesis_instrumentation__.Notify(498994)
		fmt.Printf("attempting to get results, current headOfLinePosition = %d\n", b.headOfLinePosition)
	} else {
		__antithesis_instrumentation__.Notify(498995)
	}
	__antithesis_instrumentation__.Notify(498991)
	for len(b.buffered) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(498996)
		return b.buffered[0].position == b.headOfLinePosition == true
	}() == true {
		__antithesis_instrumentation__.Notify(498997)
		result, toConsume, err := b.buffered[0].get(ctx, b)
		if err != nil {
			__antithesis_instrumentation__.Notify(499000)
			b.setErrorLocked(err)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(499001)
		}
		__antithesis_instrumentation__.Notify(498998)
		if toConsume > 0 {
			__antithesis_instrumentation__.Notify(499002)
			if err = b.budget.consumeLocked(ctx, toConsume, len(res) == 0); err != nil {
				__antithesis_instrumentation__.Notify(499003)
				if len(res) > 0 {
					__antithesis_instrumentation__.Notify(499005)

					b.buffered[0].spill(b.buffered[0].diskRowID)
					break
				} else {
					__antithesis_instrumentation__.Notify(499006)
				}
				__antithesis_instrumentation__.Notify(499004)
				b.setErrorLocked(err)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(499007)
			}
		} else {
			__antithesis_instrumentation__.Notify(499008)
		}
		__antithesis_instrumentation__.Notify(498999)
		res = append(res, result)
		heap.Remove(b, 0)
		if result.GetResp != nil || func() bool {
			__antithesis_instrumentation__.Notify(499009)
			return result.ScanResp.Complete == true
		}() == true {
			__antithesis_instrumentation__.Notify(499010)

			b.headOfLinePosition++
		} else {
			__antithesis_instrumentation__.Notify(499011)
		}
	}
	__antithesis_instrumentation__.Notify(498992)

	b.numUnreleasedResults += len(res)
	if debug {
		__antithesis_instrumentation__.Notify(499012)
		if len(res) > 0 {
			__antithesis_instrumentation__.Notify(499013)
			fmt.Printf("returning %s to the client, headOfLinePosition is now %d\n", resultsToString(res), b.headOfLinePosition)
		} else {
			__antithesis_instrumentation__.Notify(499014)
		}
	} else {
		__antithesis_instrumentation__.Notify(499015)
	}
	__antithesis_instrumentation__.Notify(498993)

	allComplete := b.numCompleteResponses == b.numExpectedResponses && func() bool {
		__antithesis_instrumentation__.Notify(499016)
		return len(b.buffered) == 0 == true
	}() == true
	return res, allComplete, b.err
}

func (b *inOrderResultsBuffer) stringLocked() string {
	__antithesis_instrumentation__.Notify(499017)
	b.Mutex.AssertHeld()
	result := "buffered for "
	for i := range b.buffered {
		__antithesis_instrumentation__.Notify(499019)
		if i > 0 {
			__antithesis_instrumentation__.Notify(499022)
			result += ", "
		} else {
			__antithesis_instrumentation__.Notify(499023)
		}
		__antithesis_instrumentation__.Notify(499020)
		var onDiskInfo string
		if b.buffered[i].onDisk {
			__antithesis_instrumentation__.Notify(499024)
			onDiskInfo = " (on disk)"
		} else {
			__antithesis_instrumentation__.Notify(499025)
		}
		__antithesis_instrumentation__.Notify(499021)
		result += fmt.Sprintf("[%d]%s: size %d", b.buffered[i].position, onDiskInfo, b.buffered[i].memoryTok.toRelease)
	}
	__antithesis_instrumentation__.Notify(499018)
	return result
}

func (b *inOrderResultsBuffer) spill(
	ctx context.Context, atLeastBytes int64, spillingPriority int,
) (spilled bool, _ error) {
	__antithesis_instrumentation__.Notify(499026)
	b.budget.mu.AssertHeld()
	b.Lock()
	defer b.Unlock()
	if buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(499032)

		defer func() {
			__antithesis_instrumentation__.Notify(499033)
			if !spilled {
				__antithesis_instrumentation__.Notify(499034)
				for i := range b.buffered {
					__antithesis_instrumentation__.Notify(499035)
					if b.buffered[i].position > spillingPriority && func() bool {
						__antithesis_instrumentation__.Notify(499036)
						return !b.buffered[i].onDisk == true
					}() == true {
						__antithesis_instrumentation__.Notify(499037)
						panic(errors.AssertionFailedf(
							"unexpectedly result for position %d wasn't spilled, spilling priority %d\n%s\n",
							b.buffered[i].position, spillingPriority, b.stringLocked()),
						)
					} else {
						__antithesis_instrumentation__.Notify(499038)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(499039)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(499040)
	}
	__antithesis_instrumentation__.Notify(499027)
	if len(b.buffered) == 0 {
		__antithesis_instrumentation__.Notify(499041)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(499042)
	}
	__antithesis_instrumentation__.Notify(499028)

	if debug {
		__antithesis_instrumentation__.Notify(499043)
		fmt.Printf(
			"want to spill at least %d bytes with priority %d\t%s\n",
			atLeastBytes, spillingPriority, b.stringLocked(),
		)
	} else {
		__antithesis_instrumentation__.Notify(499044)
	}
	__antithesis_instrumentation__.Notify(499029)
	if !b.disk.initialized {
		__antithesis_instrumentation__.Notify(499045)
		b.disk.container = rowcontainer.MakeDiskRowContainer(
			b.disk.monitor,
			inOrderResultsBufferSpillTypeSchema,
			colinfo.ColumnOrdering{},
			b.disk.engine,
		)
		b.disk.initialized = true
		b.disk.rowScratch = make(rowenc.EncDatumRow, len(inOrderResultsBufferSpillTypeSchema))
	} else {
		__antithesis_instrumentation__.Notify(499046)
	}
	__antithesis_instrumentation__.Notify(499030)

	for idx := len(b.buffered) - 1; idx >= 0; idx-- {
		__antithesis_instrumentation__.Notify(499047)
		if r := &b.buffered[idx]; !r.onDisk && func() bool {
			__antithesis_instrumentation__.Notify(499048)
			return r.position > spillingPriority == true
		}() == true {
			__antithesis_instrumentation__.Notify(499049)
			if debug {
				__antithesis_instrumentation__.Notify(499054)
				fmt.Printf(
					"spilling a result for position %d which will free up %d bytes\n",
					r.position, r.memoryTok.toRelease,
				)
			} else {
				__antithesis_instrumentation__.Notify(499055)
			}
			__antithesis_instrumentation__.Notify(499050)
			if err := b.buffered[idx].serialize(b.disk.rowScratch, &b.disk.alloc); err != nil {
				__antithesis_instrumentation__.Notify(499056)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(499057)
			}
			__antithesis_instrumentation__.Notify(499051)
			if err := b.disk.container.AddRow(ctx, b.disk.rowScratch); err != nil {
				__antithesis_instrumentation__.Notify(499058)
				b.setErrorLocked(err)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(499059)
			}
			__antithesis_instrumentation__.Notify(499052)

			if b.disk.iter != nil {
				__antithesis_instrumentation__.Notify(499060)
				b.disk.iter.Close()
				b.disk.iter = nil
				b.disk.iterRowID = 0
			} else {
				__antithesis_instrumentation__.Notify(499061)
			}
			__antithesis_instrumentation__.Notify(499053)

			r.spill(b.disk.container.Len() - 1)
			b.budget.releaseLocked(ctx, r.memoryTok.toRelease)
			atLeastBytes -= r.memoryTok.toRelease
			if atLeastBytes <= 0 {
				__antithesis_instrumentation__.Notify(499062)
				if debug {
					__antithesis_instrumentation__.Notify(499064)
					fmt.Println("the spill was successful")
				} else {
					__antithesis_instrumentation__.Notify(499065)
				}
				__antithesis_instrumentation__.Notify(499063)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(499066)
			}
		} else {
			__antithesis_instrumentation__.Notify(499067)
		}
	}
	__antithesis_instrumentation__.Notify(499031)
	return false, nil
}

func (b *inOrderResultsBuffer) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(499068)
	b.Lock()
	defer b.Unlock()
	if b.disk.initialized {
		__antithesis_instrumentation__.Notify(499070)
		if b.disk.iter != nil {
			__antithesis_instrumentation__.Notify(499072)
			b.disk.iter.Close()
			b.disk.iter = nil
		} else {
			__antithesis_instrumentation__.Notify(499073)
		}
		__antithesis_instrumentation__.Notify(499071)
		b.disk.container.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(499074)
	}
	__antithesis_instrumentation__.Notify(499069)

	b.signal()
}

type inOrderBufferedResult struct {
	Result

	onDisk    bool
	diskRowID int
}

func (r *inOrderBufferedResult) spill(diskRowID int) {
	__antithesis_instrumentation__.Notify(499075)
	isScanComplete := r.ScanResp.Complete
	*r = inOrderBufferedResult{
		Result:    Result{memoryTok: r.memoryTok, position: r.position},
		onDisk:    true,
		diskRowID: diskRowID,
	}
	r.ScanResp.Complete = isScanComplete
}

func (r *inOrderBufferedResult) get(
	ctx context.Context, b *inOrderResultsBuffer,
) (_ Result, toConsume int64, _ error) {
	__antithesis_instrumentation__.Notify(499076)
	if !r.onDisk {
		__antithesis_instrumentation__.Notify(499083)
		return r.Result, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(499084)
	}
	__antithesis_instrumentation__.Notify(499077)

	if b.disk.iter == nil {
		__antithesis_instrumentation__.Notify(499085)
		b.disk.iter = b.disk.container.NewIterator(ctx)
		b.disk.iter.Rewind()
	} else {
		__antithesis_instrumentation__.Notify(499086)
	}
	__antithesis_instrumentation__.Notify(499078)
	if r.diskRowID < b.disk.iterRowID {
		__antithesis_instrumentation__.Notify(499087)
		b.disk.iter.Rewind()
		b.disk.iterRowID = 0
	} else {
		__antithesis_instrumentation__.Notify(499088)
	}
	__antithesis_instrumentation__.Notify(499079)
	for b.disk.iterRowID < r.diskRowID {
		__antithesis_instrumentation__.Notify(499089)
		b.disk.iter.Next()
		b.disk.iterRowID++
	}
	__antithesis_instrumentation__.Notify(499080)

	serialized, err := b.disk.iter.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(499090)
		return Result{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(499091)
	}
	__antithesis_instrumentation__.Notify(499081)
	if err = deserialize(&r.Result, serialized, &b.disk.alloc); err != nil {
		__antithesis_instrumentation__.Notify(499092)
		return Result{}, 0, err
	} else {
		__antithesis_instrumentation__.Notify(499093)
	}
	__antithesis_instrumentation__.Notify(499082)
	return r.Result, r.memoryTok.toRelease, err
}

var inOrderResultsBufferSpillTypeSchema = []*types.T{
	types.Bool,

	types.Bytes, types.Int, types.Int, types.Bool,

	types.BytesArray,
	types.IntArray,
}

type resultSerializationIndex int

const (
	isGetIdx resultSerializationIndex = iota
	getRawBytesIdx
	getTSWallTimeIdx
	getTSLogicalIdx
	getTSSyntheticIdx
	scanBatchResponsesIdx
	enqueueKeysSatisfiedIdx
)

func (r *Result) serialize(row rowenc.EncDatumRow, alloc *tree.DatumAlloc) error {
	__antithesis_instrumentation__.Notify(499094)
	row[isGetIdx] = rowenc.EncDatum{Datum: tree.MakeDBool(r.GetResp != nil)}
	if r.GetResp != nil && func() bool {
		__antithesis_instrumentation__.Notify(499097)
		return r.GetResp.Value != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(499098)

		v := r.GetResp.Value
		row[getRawBytesIdx] = rowenc.EncDatum{Datum: alloc.NewDBytes(tree.DBytes(v.RawBytes))}
		row[getTSWallTimeIdx] = rowenc.EncDatum{Datum: alloc.NewDInt(tree.DInt(v.Timestamp.WallTime))}
		row[getTSLogicalIdx] = rowenc.EncDatum{Datum: alloc.NewDInt(tree.DInt(v.Timestamp.Logical))}
		row[getTSSyntheticIdx] = rowenc.EncDatum{Datum: tree.MakeDBool(tree.DBool(v.Timestamp.Synthetic))}
		row[scanBatchResponsesIdx] = rowenc.EncDatum{Datum: tree.DNull}
	} else {
		__antithesis_instrumentation__.Notify(499099)
		row[getRawBytesIdx] = rowenc.EncDatum{Datum: tree.DNull}
		row[getTSWallTimeIdx] = rowenc.EncDatum{Datum: tree.DNull}
		row[getTSLogicalIdx] = rowenc.EncDatum{Datum: tree.DNull}
		row[getTSSyntheticIdx] = rowenc.EncDatum{Datum: tree.DNull}
		if r.GetResp != nil {
			__antithesis_instrumentation__.Notify(499100)

			row[scanBatchResponsesIdx] = rowenc.EncDatum{Datum: tree.DNull}
		} else {
			__antithesis_instrumentation__.Notify(499101)

			batchResponses := tree.NewDArray(types.Bytes)
			batchResponses.Array = make(tree.Datums, 0, len(r.ScanResp.BatchResponses))
			for _, b := range r.ScanResp.BatchResponses {
				__antithesis_instrumentation__.Notify(499103)
				if err := batchResponses.Append(alloc.NewDBytes(tree.DBytes(b))); err != nil {
					__antithesis_instrumentation__.Notify(499104)
					return err
				} else {
					__antithesis_instrumentation__.Notify(499105)
				}
			}
			__antithesis_instrumentation__.Notify(499102)
			row[scanBatchResponsesIdx] = rowenc.EncDatum{Datum: batchResponses}
		}
	}
	__antithesis_instrumentation__.Notify(499095)
	enqueueKeysSatisfied := tree.NewDArray(types.Int)
	enqueueKeysSatisfied.Array = make(tree.Datums, 0, len(r.EnqueueKeysSatisfied))
	for _, k := range r.EnqueueKeysSatisfied {
		__antithesis_instrumentation__.Notify(499106)
		if err := enqueueKeysSatisfied.Append(alloc.NewDInt(tree.DInt(k))); err != nil {
			__antithesis_instrumentation__.Notify(499107)
			return err
		} else {
			__antithesis_instrumentation__.Notify(499108)
		}
	}
	__antithesis_instrumentation__.Notify(499096)
	row[enqueueKeysSatisfiedIdx] = rowenc.EncDatum{Datum: enqueueKeysSatisfied}
	return nil
}

func deserialize(r *Result, row rowenc.EncDatumRow, alloc *tree.DatumAlloc) error {
	__antithesis_instrumentation__.Notify(499109)
	for i := range row {
		__antithesis_instrumentation__.Notify(499113)
		if err := row[i].EnsureDecoded(inOrderResultsBufferSpillTypeSchema[i], alloc); err != nil {
			__antithesis_instrumentation__.Notify(499114)
			return err
		} else {
			__antithesis_instrumentation__.Notify(499115)
		}
	}
	__antithesis_instrumentation__.Notify(499110)
	if isGet := tree.MustBeDBool(row[isGetIdx].Datum); isGet {
		__antithesis_instrumentation__.Notify(499116)
		r.GetResp = &roachpb.GetResponse{}
		if row[getRawBytesIdx].Datum != tree.DNull {
			__antithesis_instrumentation__.Notify(499117)
			r.GetResp.Value = &roachpb.Value{
				RawBytes: []byte(tree.MustBeDBytes(row[getRawBytesIdx].Datum)),
				Timestamp: hlc.Timestamp{
					WallTime:  int64(tree.MustBeDInt(row[getTSWallTimeIdx].Datum)),
					Logical:   int32(tree.MustBeDInt(row[getTSLogicalIdx].Datum)),
					Synthetic: bool(tree.MustBeDBool(row[getTSSyntheticIdx].Datum)),
				},
			}
		} else {
			__antithesis_instrumentation__.Notify(499118)
		}
	} else {
		__antithesis_instrumentation__.Notify(499119)
		r.ScanResp.ScanResponse = &roachpb.ScanResponse{}
		batchResponses := tree.MustBeDArray(row[scanBatchResponsesIdx].Datum)
		r.ScanResp.ScanResponse.BatchResponses = make([][]byte, batchResponses.Len())
		for i := range batchResponses.Array {
			__antithesis_instrumentation__.Notify(499120)
			r.ScanResp.ScanResponse.BatchResponses[i] = []byte(tree.MustBeDBytes(batchResponses.Array[i]))
		}
	}
	__antithesis_instrumentation__.Notify(499111)
	enqueueKeysSatisfied := tree.MustBeDArray(row[enqueueKeysSatisfiedIdx].Datum)
	r.EnqueueKeysSatisfied = make([]int, enqueueKeysSatisfied.Len())
	for i := range enqueueKeysSatisfied.Array {
		__antithesis_instrumentation__.Notify(499121)
		r.EnqueueKeysSatisfied[i] = int(tree.MustBeDInt(enqueueKeysSatisfied.Array[i]))
	}
	__antithesis_instrumentation__.Notify(499112)
	return nil
}
