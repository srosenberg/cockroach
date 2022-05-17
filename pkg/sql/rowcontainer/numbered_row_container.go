package rowcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"
	"math"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type DiskBackedNumberedRowContainer struct {
	deDup bool
	rc    *DiskBackedRowContainer

	deduper DeDupingRowContainer

	storedTypes []*types.T
	idx         int

	rowIter *numberedDiskRowIterator

	cacheMap      map[int]*cacheElement
	rowIterMemAcc mon.BoundAccount
	DisableCache  bool
}

func NewDiskBackedNumberedRowContainer(
	deDup bool,
	types []*types.T,
	evalCtx *tree.EvalContext,
	engine diskmap.Factory,
	memoryMonitor *mon.BytesMonitor,
	diskMonitor *mon.BytesMonitor,
) *DiskBackedNumberedRowContainer {
	__antithesis_instrumentation__.Notify(569431)
	d := &DiskBackedNumberedRowContainer{
		deDup:         deDup,
		storedTypes:   types,
		rowIterMemAcc: memoryMonitor.MakeBoundAccount(),
	}
	d.rc = &DiskBackedRowContainer{}
	d.rc.Init(nil, types, evalCtx, engine, memoryMonitor, diskMonitor)
	if deDup {
		__antithesis_instrumentation__.Notify(569433)
		ordering := make(colinfo.ColumnOrdering, len(types))
		for i := range types {
			__antithesis_instrumentation__.Notify(569435)
			ordering[i].ColIdx = i
			ordering[i].Direction = encoding.Ascending
		}
		__antithesis_instrumentation__.Notify(569434)
		deduper := &DiskBackedRowContainer{}
		deduper.Init(ordering, types, evalCtx, engine, memoryMonitor, diskMonitor)
		deduper.DoDeDuplicate()
		d.deduper = deduper
	} else {
		__antithesis_instrumentation__.Notify(569436)
	}
	__antithesis_instrumentation__.Notify(569432)
	return d
}

func (d *DiskBackedNumberedRowContainer) UsingDisk() bool {
	__antithesis_instrumentation__.Notify(569437)
	return d.rc.UsingDisk()
}

func (d *DiskBackedNumberedRowContainer) Spilled() bool {
	__antithesis_instrumentation__.Notify(569438)
	return d.rc.Spilled()
}

func (d *DiskBackedNumberedRowContainer) SpillToDisk(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(569439)
	if d.rc.UsingDisk() && func() bool {
		__antithesis_instrumentation__.Notify(569443)
		return (!d.deDup || func() bool {
			__antithesis_instrumentation__.Notify(569444)
			return d.deduper.(*DiskBackedRowContainer).UsingDisk() == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(569445)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(569446)
	}
	__antithesis_instrumentation__.Notify(569440)
	if !d.rc.UsingDisk() {
		__antithesis_instrumentation__.Notify(569447)
		if err := d.rc.SpillToDisk(ctx); err != nil {
			__antithesis_instrumentation__.Notify(569448)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(569449)
		}
	} else {
		__antithesis_instrumentation__.Notify(569450)
	}
	__antithesis_instrumentation__.Notify(569441)
	if d.deDup && func() bool {
		__antithesis_instrumentation__.Notify(569451)
		return !d.deduper.(*DiskBackedRowContainer).UsingDisk() == true
	}() == true {
		__antithesis_instrumentation__.Notify(569452)
		if err := d.deduper.(*DiskBackedRowContainer).SpillToDisk(ctx); err != nil {
			__antithesis_instrumentation__.Notify(569453)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(569454)
		}
	} else {
		__antithesis_instrumentation__.Notify(569455)
	}
	__antithesis_instrumentation__.Notify(569442)
	return true, nil
}

func (d *DiskBackedNumberedRowContainer) AddRow(
	ctx context.Context, row rowenc.EncDatumRow,
) (int, error) {
	__antithesis_instrumentation__.Notify(569456)
	if d.deDup {
		__antithesis_instrumentation__.Notify(569458)
		assignedIdx, err := d.deduper.AddRowWithDeDup(ctx, row)
		if err != nil {
			__antithesis_instrumentation__.Notify(569460)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(569461)
		}
		__antithesis_instrumentation__.Notify(569459)
		if assignedIdx < d.idx {
			__antithesis_instrumentation__.Notify(569462)

			return assignedIdx, nil
		} else {
			__antithesis_instrumentation__.Notify(569463)
			if assignedIdx != d.idx {
				__antithesis_instrumentation__.Notify(569464)
				panic(errors.AssertionFailedf("DiskBackedNumberedRowContainer bug: assignedIdx %d != d.idx %d",
					assignedIdx, d.idx))
			} else {
				__antithesis_instrumentation__.Notify(569465)
			}
		}

	} else {
		__antithesis_instrumentation__.Notify(569466)
	}
	__antithesis_instrumentation__.Notify(569457)
	idx := d.idx

	d.idx++
	return idx, d.rc.AddRow(ctx, row)
}

func (d *DiskBackedNumberedRowContainer) SetupForRead(ctx context.Context, accesses [][]int) {
	__antithesis_instrumentation__.Notify(569467)
	if !d.rc.UsingDisk() {
		__antithesis_instrumentation__.Notify(569472)
		return
	} else {
		__antithesis_instrumentation__.Notify(569473)
	}
	__antithesis_instrumentation__.Notify(569468)
	rowIter := d.rc.drc.newNumberedIterator(ctx)
	meanBytesPerRow := d.rc.drc.MeanEncodedRowBytes()
	if meanBytesPerRow == 0 {
		__antithesis_instrumentation__.Notify(569474)
		meanBytesPerRow = 100
	} else {
		__antithesis_instrumentation__.Notify(569475)
	}
	__antithesis_instrumentation__.Notify(569469)

	const bytesPerSSBlock = 32 * 1024
	meanRowsPerSSBlock := bytesPerSSBlock / meanBytesPerRow
	const maxCacheSize = 4096
	cacheSize := maxCacheSize
	if d.DisableCache {
		__antithesis_instrumentation__.Notify(569476)

		cacheSize = 0
	} else {
		__antithesis_instrumentation__.Notify(569477)
	}
	__antithesis_instrumentation__.Notify(569470)
	if d.cacheMap == nil {
		__antithesis_instrumentation__.Notify(569478)
		d.cacheMap = make(map[int]*cacheElement)
	} else {
		__antithesis_instrumentation__.Notify(569479)
	}
	__antithesis_instrumentation__.Notify(569471)
	d.rowIter = newNumberedDiskRowIterator(
		ctx, rowIter, accesses, meanRowsPerSSBlock, cacheSize, d.cacheMap, &d.rowIterMemAcc)
}

func (d *DiskBackedNumberedRowContainer) GetRow(
	ctx context.Context, idx int, skip bool,
) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569480)
	if !d.rc.UsingDisk() {
		__antithesis_instrumentation__.Notify(569482)
		if skip {
			__antithesis_instrumentation__.Notify(569484)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(569485)
		}
		__antithesis_instrumentation__.Notify(569483)
		return d.rc.mrc.EncRow(idx), nil
	} else {
		__antithesis_instrumentation__.Notify(569486)
	}
	__antithesis_instrumentation__.Notify(569481)
	return d.rowIter.getRow(ctx, idx, skip)
}

func (d *DiskBackedNumberedRowContainer) UnsafeReset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569487)
	if d.rowIter != nil {
		__antithesis_instrumentation__.Notify(569491)
		d.rowIter.close()
		d.rowIterMemAcc.Clear(ctx)
		d.rowIter = nil
	} else {
		__antithesis_instrumentation__.Notify(569492)
	}
	__antithesis_instrumentation__.Notify(569488)
	d.idx = 0
	if err := d.rc.UnsafeReset(ctx); err != nil {
		__antithesis_instrumentation__.Notify(569493)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569494)
	}
	__antithesis_instrumentation__.Notify(569489)
	if d.deduper != nil {
		__antithesis_instrumentation__.Notify(569495)
		if err := d.deduper.UnsafeReset(ctx); err != nil {
			__antithesis_instrumentation__.Notify(569496)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569497)
		}
	} else {
		__antithesis_instrumentation__.Notify(569498)
	}
	__antithesis_instrumentation__.Notify(569490)
	return nil
}

func (d *DiskBackedNumberedRowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569499)
	if d.rowIter != nil {
		__antithesis_instrumentation__.Notify(569502)
		d.rowIter.close()
	} else {
		__antithesis_instrumentation__.Notify(569503)
	}
	__antithesis_instrumentation__.Notify(569500)
	d.rowIterMemAcc.Close(ctx)
	d.rc.Close(ctx)
	if d.deduper != nil {
		__antithesis_instrumentation__.Notify(569504)
		d.deduper.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(569505)
	}
	__antithesis_instrumentation__.Notify(569501)
	d.cacheMap = nil
}

type numberedDiskRowIterator struct {
	rowIter *numberedRowIterator

	isPositioned bool

	idxRowIter int

	meanRowsPerSSBlock int

	maxCacheSize int
	memAcc       *mon.BoundAccount

	cache map[int]*cacheElement

	accessIdx int

	cacheHeap  cacheMaxNextAccessHeap
	datumAlloc tree.DatumAlloc
	rowAlloc   rowenc.EncDatumRowAlloc

	hitCount  int
	missCount int
}

type cacheElement struct {
	accesses []int

	row rowenc.EncDatumRow

	heapElement cacheRowHeapElement

	numAccesses int
}

var cacheElementSyncPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(569506)
		return &cacheElement{}
	},
}

func freeCacheElement(elem *cacheElement) {
	__antithesis_instrumentation__.Notify(569507)
	elem.accesses = nil
	elem.row = nil
	elem.numAccesses = 0
	cacheElementSyncPool.Put(elem)
}

func newCacheElement() *cacheElement {
	__antithesis_instrumentation__.Notify(569508)
	return cacheElementSyncPool.Get().(*cacheElement)
}

type cacheRowHeapElement struct {
	rowIdx int

	nextAccess int

	heapIdx int
}

type cacheMaxNextAccessHeap []*cacheRowHeapElement

func (h cacheMaxNextAccessHeap) Len() int {
	__antithesis_instrumentation__.Notify(569509)
	return len(h)
}
func (h cacheMaxNextAccessHeap) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(569510)
	return h[i].nextAccess > h[j].nextAccess
}
func (h cacheMaxNextAccessHeap) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(569511)
	h[i], h[j] = h[j], h[i]
	h[i].heapIdx = i
	h[j].heapIdx = j
}
func (h *cacheMaxNextAccessHeap) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(569512)
	n := len(*h)
	elem := x.(*cacheRowHeapElement)
	elem.heapIdx = n
	*h = append(*h, elem)
}
func (h *cacheMaxNextAccessHeap) Pop() interface{} {
	__antithesis_instrumentation__.Notify(569513)
	old := *h
	n := len(old)
	elem := old[n-1]
	elem.heapIdx = -1
	*h = old[0 : n-1]
	return elem
}

func newNumberedDiskRowIterator(
	_ context.Context,
	rowIter *numberedRowIterator,
	accesses [][]int,
	meanRowsPerSSBlock int,
	maxCacheSize int,
	cache map[int]*cacheElement,
	memAcc *mon.BoundAccount,
) *numberedDiskRowIterator {
	__antithesis_instrumentation__.Notify(569514)
	n := &numberedDiskRowIterator{
		rowIter:            rowIter,
		meanRowsPerSSBlock: meanRowsPerSSBlock,
		maxCacheSize:       maxCacheSize,
		memAcc:             memAcc,
		cache:              cache,
	}
	var numAccesses int
	for _, accSlice := range accesses {
		__antithesis_instrumentation__.Notify(569517)
		for _, rowIdx := range accSlice {
			__antithesis_instrumentation__.Notify(569518)
			elem := n.cache[rowIdx]
			if elem == nil {
				__antithesis_instrumentation__.Notify(569520)
				elem = newCacheElement()
				elem.heapElement.rowIdx = rowIdx
				n.cache[rowIdx] = elem
			} else {
				__antithesis_instrumentation__.Notify(569521)
			}
			__antithesis_instrumentation__.Notify(569519)
			elem.numAccesses++
			numAccesses++
		}
	}
	__antithesis_instrumentation__.Notify(569515)
	allAccesses := make([]int, numAccesses)
	accessIdx := 0
	for _, accSlice := range accesses {
		__antithesis_instrumentation__.Notify(569522)
		for _, rowIdx := range accSlice {
			__antithesis_instrumentation__.Notify(569523)
			elem := n.cache[rowIdx]
			if elem.accesses == nil {
				__antithesis_instrumentation__.Notify(569525)

				elem.accesses = allAccesses[0:0:elem.numAccesses]
				allAccesses = allAccesses[elem.numAccesses:]
			} else {
				__antithesis_instrumentation__.Notify(569526)
			}
			__antithesis_instrumentation__.Notify(569524)
			elem.accesses = append(elem.accesses, accessIdx)
			accessIdx++
		}
	}
	__antithesis_instrumentation__.Notify(569516)
	return n
}

func (n *numberedDiskRowIterator) close() {
	__antithesis_instrumentation__.Notify(569527)
	n.rowIter.Close()
	for k, v := range n.cache {
		__antithesis_instrumentation__.Notify(569528)
		freeCacheElement(v)
		delete(n.cache, k)
	}
}

func (n *numberedDiskRowIterator) getRow(
	ctx context.Context, idx int, skip bool,
) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569529)
	thisAccessIdx := n.accessIdx
	n.accessIdx++
	elem, ok := n.cache[idx]
	if !ok {
		__antithesis_instrumentation__.Notify(569537)
		return nil, errors.Errorf("caller is accessing a row that was not specified up front")
	} else {
		__antithesis_instrumentation__.Notify(569538)
	}
	__antithesis_instrumentation__.Notify(569530)
	if len(elem.accesses) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(569539)
		return elem.accesses[0] != thisAccessIdx == true
	}() == true {
		__antithesis_instrumentation__.Notify(569540)
		return nil, errors.Errorf("caller is no longer synchronized with future accesses")
	} else {
		__antithesis_instrumentation__.Notify(569541)
	}
	__antithesis_instrumentation__.Notify(569531)
	elem.accesses = elem.accesses[1:]
	var nextAccess int
	if len(elem.accesses) > 0 {
		__antithesis_instrumentation__.Notify(569542)
		nextAccess = elem.accesses[0]
	} else {
		__antithesis_instrumentation__.Notify(569543)
		nextAccess = math.MaxInt32
	}
	__antithesis_instrumentation__.Notify(569532)

	if elem.row != nil {
		__antithesis_instrumentation__.Notify(569544)
		n.hitCount++
		elem.heapElement.nextAccess = nextAccess
		heap.Fix(&n.cacheHeap, elem.heapElement.heapIdx)
		if skip {
			__antithesis_instrumentation__.Notify(569546)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(569547)
		}
		__antithesis_instrumentation__.Notify(569545)
		return elem.row, nil
	} else {
		__antithesis_instrumentation__.Notify(569548)
	}
	__antithesis_instrumentation__.Notify(569533)

	n.missCount++

	if skip {
		__antithesis_instrumentation__.Notify(569549)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(569550)
	}
	__antithesis_instrumentation__.Notify(569534)

	if n.isPositioned && func() bool {
		__antithesis_instrumentation__.Notify(569551)
		return idx >= n.idxRowIter == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(569552)
		return (idx-n.idxRowIter <= n.meanRowsPerSSBlock) == true
	}() == true {
		__antithesis_instrumentation__.Notify(569553)

		for i := idx - n.idxRowIter; i > 0; {
			__antithesis_instrumentation__.Notify(569555)
			n.rowIter.Next()
			if valid, err := n.rowIter.Valid(); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(569561)
				return !valid == true
			}() == true {
				__antithesis_instrumentation__.Notify(569562)
				if err != nil {
					__antithesis_instrumentation__.Notify(569564)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(569565)
				}
				__antithesis_instrumentation__.Notify(569563)
				return nil, errors.Errorf("caller is asking for index higher than any added index")
			} else {
				__antithesis_instrumentation__.Notify(569566)
			}
			__antithesis_instrumentation__.Notify(569556)
			n.idxRowIter++
			i--
			if i == 0 {
				__antithesis_instrumentation__.Notify(569567)
				break
			} else {
				__antithesis_instrumentation__.Notify(569568)
			}
			__antithesis_instrumentation__.Notify(569557)

			preElem, ok := n.cache[n.idxRowIter]
			if !ok {
				__antithesis_instrumentation__.Notify(569569)

				continue
			} else {
				__antithesis_instrumentation__.Notify(569570)
			}
			__antithesis_instrumentation__.Notify(569558)
			if preElem.row != nil {
				__antithesis_instrumentation__.Notify(569571)

				continue
			} else {
				__antithesis_instrumentation__.Notify(569572)
			}
			__antithesis_instrumentation__.Notify(569559)
			if len(preElem.accesses) == 0 {
				__antithesis_instrumentation__.Notify(569573)

				continue
			} else {
				__antithesis_instrumentation__.Notify(569574)
			}
			__antithesis_instrumentation__.Notify(569560)
			if err := n.tryAddCache(ctx, preElem); err != nil {
				__antithesis_instrumentation__.Notify(569575)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(569576)
			}
		}
		__antithesis_instrumentation__.Notify(569554)

		return n.tryAddCacheAndReturnRow(ctx, elem)
	} else {
		__antithesis_instrumentation__.Notify(569577)
	}
	__antithesis_instrumentation__.Notify(569535)
	n.rowIter.seekToIndex(idx)
	n.isPositioned = true
	n.idxRowIter = idx
	if valid, err := n.rowIter.Valid(); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(569578)
		return !valid == true
	}() == true {
		__antithesis_instrumentation__.Notify(569579)
		if err != nil {
			__antithesis_instrumentation__.Notify(569581)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(569582)
		}
		__antithesis_instrumentation__.Notify(569580)
		return nil, errors.Errorf("caller is asking for index higher than any added index")
	} else {
		__antithesis_instrumentation__.Notify(569583)
	}
	__antithesis_instrumentation__.Notify(569536)

	return n.tryAddCacheAndReturnRow(ctx, elem)
}

func (n *numberedDiskRowIterator) ensureDecoded(row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569584)
	for i := range row {
		__antithesis_instrumentation__.Notify(569586)
		if err := row[i].EnsureDecoded(n.rowIter.rowContainer.types[i], &n.datumAlloc); err != nil {
			__antithesis_instrumentation__.Notify(569587)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569588)
		}
	}
	__antithesis_instrumentation__.Notify(569585)
	return nil
}

func (n *numberedDiskRowIterator) tryAddCacheAndReturnRow(
	ctx context.Context, elem *cacheElement,
) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569589)
	r, err := n.rowIter.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(569593)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569594)
	}
	__antithesis_instrumentation__.Notify(569590)
	if err = n.ensureDecoded(r); err != nil {
		__antithesis_instrumentation__.Notify(569595)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569596)
	}
	__antithesis_instrumentation__.Notify(569591)
	if len(elem.accesses) == 0 {
		__antithesis_instrumentation__.Notify(569597)
		return r, nil
	} else {
		__antithesis_instrumentation__.Notify(569598)
	}
	__antithesis_instrumentation__.Notify(569592)
	return r, n.tryAddCacheHelper(ctx, elem, r, true)
}

func (n *numberedDiskRowIterator) tryAddCache(ctx context.Context, elem *cacheElement) error {
	__antithesis_instrumentation__.Notify(569599)

	cacheSize := len(n.cacheHeap)
	if cacheSize == n.maxCacheSize && func() bool {
		__antithesis_instrumentation__.Notify(569602)
		return (cacheSize == 0 || func() bool {
			__antithesis_instrumentation__.Notify(569603)
			return n.cacheHeap[0].nextAccess <= elem.accesses[0] == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(569604)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(569605)
	}
	__antithesis_instrumentation__.Notify(569600)
	row, err := n.rowIter.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(569606)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569607)
	}
	__antithesis_instrumentation__.Notify(569601)
	return n.tryAddCacheHelper(ctx, elem, row, false)
}

func (n *numberedDiskRowIterator) tryAddCacheHelper(
	ctx context.Context, elem *cacheElement, row rowenc.EncDatumRow, alreadyDecoded bool,
) error {
	__antithesis_instrumentation__.Notify(569608)
	if elem.row != nil {
		__antithesis_instrumentation__.Notify(569613)
		log.Fatalf(ctx, "adding row to cache when it is already in cache")
	} else {
		__antithesis_instrumentation__.Notify(569614)
	}
	__antithesis_instrumentation__.Notify(569609)
	nextAccess := elem.accesses[0]
	evict := func() (rowenc.EncDatumRow, error) {
		__antithesis_instrumentation__.Notify(569615)
		heapElem := heap.Pop(&n.cacheHeap).(*cacheRowHeapElement)
		evictElem, ok := n.cache[heapElem.rowIdx]
		if !ok {
			__antithesis_instrumentation__.Notify(569617)
			return nil, errors.Errorf("bug: element not in cache map")
		} else {
			__antithesis_instrumentation__.Notify(569618)
		}
		__antithesis_instrumentation__.Notify(569616)
		bytes := evictElem.row.Size()
		n.memAcc.Shrink(ctx, int64(bytes))
		evictedRow := evictElem.row
		evictElem.row = nil
		return evictedRow, nil
	}
	__antithesis_instrumentation__.Notify(569610)
	rowBytesUsage := -1
	var rowToReuse rowenc.EncDatumRow
	for {
		__antithesis_instrumentation__.Notify(569619)
		if n.maxCacheSize == 0 {
			__antithesis_instrumentation__.Notify(569625)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(569626)
		}
		__antithesis_instrumentation__.Notify(569620)
		if len(n.cacheHeap) == n.maxCacheSize && func() bool {
			__antithesis_instrumentation__.Notify(569627)
			return n.cacheHeap[0].nextAccess <= nextAccess == true
		}() == true {
			__antithesis_instrumentation__.Notify(569628)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(569629)
		}
		__antithesis_instrumentation__.Notify(569621)
		var err error
		if len(n.cacheHeap) >= n.maxCacheSize {
			__antithesis_instrumentation__.Notify(569630)
			if rowToReuse, err = evict(); err != nil {
				__antithesis_instrumentation__.Notify(569632)
				return err
			} else {
				__antithesis_instrumentation__.Notify(569633)
			}
			__antithesis_instrumentation__.Notify(569631)
			continue
		} else {
			__antithesis_instrumentation__.Notify(569634)
		}
		__antithesis_instrumentation__.Notify(569622)

		if !alreadyDecoded {
			__antithesis_instrumentation__.Notify(569635)
			err = n.ensureDecoded(row)
			if err != nil {
				__antithesis_instrumentation__.Notify(569637)
				return err
			} else {
				__antithesis_instrumentation__.Notify(569638)
			}
			__antithesis_instrumentation__.Notify(569636)
			alreadyDecoded = true
		} else {
			__antithesis_instrumentation__.Notify(569639)
		}
		__antithesis_instrumentation__.Notify(569623)
		if rowBytesUsage == -1 {
			__antithesis_instrumentation__.Notify(569640)
			rowBytesUsage = int(row.Size())
		} else {
			__antithesis_instrumentation__.Notify(569641)
		}
		__antithesis_instrumentation__.Notify(569624)
		if err := n.memAcc.Grow(ctx, int64(rowBytesUsage)); err != nil {
			__antithesis_instrumentation__.Notify(569642)
			if sqlerrors.IsOutOfMemoryError(err) {
				__antithesis_instrumentation__.Notify(569643)

				n.maxCacheSize = len(n.cacheHeap)
				continue
			} else {
				__antithesis_instrumentation__.Notify(569644)
				return err
			}
		} else {
			__antithesis_instrumentation__.Notify(569645)

			break
		}
	}
	__antithesis_instrumentation__.Notify(569611)

	elem.heapElement.nextAccess = nextAccess

	if rowToReuse == nil {
		__antithesis_instrumentation__.Notify(569646)
		elem.row = n.rowAlloc.CopyRow(row)
	} else {
		__antithesis_instrumentation__.Notify(569647)
		copy(rowToReuse, row)
		elem.row = rowToReuse
	}
	__antithesis_instrumentation__.Notify(569612)
	heap.Push(&n.cacheHeap, &elem.heapElement)
	return nil
}
