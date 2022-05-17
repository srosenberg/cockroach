package rowcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/ring"
	"github.com/cockroachdb/cockroach/pkg/util/sort"
	"github.com/cockroachdb/errors"
)

type SortableRowContainer interface {
	Len() int

	AddRow(context.Context, rowenc.EncDatumRow) error

	Sort(context.Context)

	NewIterator(context.Context) RowIterator

	NewFinalIterator(context.Context) RowIterator

	UnsafeReset(context.Context) error

	InitTopK()

	MaybeReplaceMax(context.Context, rowenc.EncDatumRow) error

	Close(context.Context)
}

type ReorderableRowContainer interface {
	SortableRowContainer

	Reorder(context.Context, colinfo.ColumnOrdering) error
}

type IndexedRowContainer interface {
	ReorderableRowContainer

	GetRow(ctx context.Context, idx int) (tree.IndexedRow, error)
}

type DeDupingRowContainer interface {
	AddRowWithDeDup(context.Context, rowenc.EncDatumRow) (int, error)

	UnsafeReset(context.Context) error

	Close(context.Context)
}

type RowIterator interface {
	Rewind()

	Valid() (bool, error)

	Next()

	Row() (rowenc.EncDatumRow, error)

	Close()
}

type MemRowContainer struct {
	RowContainer
	types         []*types.T
	invertSorting bool
	ordering      colinfo.ColumnOrdering
	scratchRow    tree.Datums
	scratchEncRow rowenc.EncDatumRow

	evalCtx *tree.EvalContext

	datumAlloc tree.DatumAlloc
}

var _ heap.Interface = &MemRowContainer{}
var _ IndexedRowContainer = &MemRowContainer{}

func (mc *MemRowContainer) Init(
	ordering colinfo.ColumnOrdering, types []*types.T, evalCtx *tree.EvalContext,
) {
	__antithesis_instrumentation__.Notify(569648)
	mc.InitWithMon(ordering, types, evalCtx, evalCtx.Mon)
}

func (mc *MemRowContainer) InitWithMon(
	ordering colinfo.ColumnOrdering,
	types []*types.T,
	evalCtx *tree.EvalContext,
	mon *mon.BytesMonitor,
) {
	__antithesis_instrumentation__.Notify(569649)
	acc := mon.MakeBoundAccount()
	mc.RowContainer.Init(acc, colinfo.ColTypeInfoFromColTypes(types), 0)
	mc.types = types
	mc.ordering = ordering
	mc.scratchRow = make(tree.Datums, len(types))
	mc.scratchEncRow = make(rowenc.EncDatumRow, len(types))
	mc.evalCtx = evalCtx
}

func (mc *MemRowContainer) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(569650)
	cmp := colinfo.CompareDatums(mc.ordering, mc.evalCtx, mc.At(i), mc.At(j))
	if mc.invertSorting {
		__antithesis_instrumentation__.Notify(569652)
		cmp = -cmp
	} else {
		__antithesis_instrumentation__.Notify(569653)
	}
	__antithesis_instrumentation__.Notify(569651)
	return cmp < 0
}

func (mc *MemRowContainer) EncRow(idx int) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(569654)
	datums := mc.At(idx)
	for i, d := range datums {
		__antithesis_instrumentation__.Notify(569656)
		mc.scratchEncRow[i] = rowenc.DatumToEncDatum(mc.types[i], d)
	}
	__antithesis_instrumentation__.Notify(569655)
	return mc.scratchEncRow
}

func (mc *MemRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569657)
	if len(row) != len(mc.types) {
		__antithesis_instrumentation__.Notify(569660)
		log.Fatalf(ctx, "invalid row length %d, expected %d", len(row), len(mc.types))
	} else {
		__antithesis_instrumentation__.Notify(569661)
	}
	__antithesis_instrumentation__.Notify(569658)
	for i := range row {
		__antithesis_instrumentation__.Notify(569662)
		err := row[i].EnsureDecoded(mc.types[i], &mc.datumAlloc)
		if err != nil {
			__antithesis_instrumentation__.Notify(569664)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569665)
		}
		__antithesis_instrumentation__.Notify(569663)
		mc.scratchRow[i] = row[i].Datum
	}
	__antithesis_instrumentation__.Notify(569659)
	_, err := mc.RowContainer.AddRow(ctx, mc.scratchRow)
	return err
}

func (mc *MemRowContainer) Sort(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569666)
	mc.invertSorting = false
	var cancelChecker cancelchecker.CancelChecker
	cancelChecker.Reset(ctx)
	sort.Sort(mc, &cancelChecker)
}

func (mc *MemRowContainer) Reorder(_ context.Context, ordering colinfo.ColumnOrdering) error {
	__antithesis_instrumentation__.Notify(569667)
	mc.ordering = ordering
	return nil
}

func (mc *MemRowContainer) Push(_ interface{}) {
	__antithesis_instrumentation__.Notify(569668)
	panic("unimplemented")
}

func (mc *MemRowContainer) Pop() interface{} {
	__antithesis_instrumentation__.Notify(569669)
	panic("unimplemented")
}

func (mc *MemRowContainer) MaybeReplaceMax(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569670)
	max := mc.At(0)
	cmp, err := row.CompareToDatums(mc.types, &mc.datumAlloc, mc.ordering, mc.evalCtx, max)
	if err != nil {
		__antithesis_instrumentation__.Notify(569673)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569674)
	}
	__antithesis_instrumentation__.Notify(569671)
	if cmp < 0 {
		__antithesis_instrumentation__.Notify(569675)

		for i := range row {
			__antithesis_instrumentation__.Notify(569678)
			if err := row[i].EnsureDecoded(mc.types[i], &mc.datumAlloc); err != nil {
				__antithesis_instrumentation__.Notify(569680)
				return err
			} else {
				__antithesis_instrumentation__.Notify(569681)
			}
			__antithesis_instrumentation__.Notify(569679)
			mc.scratchRow[i] = row[i].Datum
		}
		__antithesis_instrumentation__.Notify(569676)
		if err := mc.Replace(ctx, 0, mc.scratchRow); err != nil {
			__antithesis_instrumentation__.Notify(569682)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569683)
		}
		__antithesis_instrumentation__.Notify(569677)
		heap.Fix(mc, 0)
	} else {
		__antithesis_instrumentation__.Notify(569684)
	}
	__antithesis_instrumentation__.Notify(569672)
	return nil
}

func (mc *MemRowContainer) InitTopK() {
	__antithesis_instrumentation__.Notify(569685)
	mc.invertSorting = true
	heap.Init(mc)
}

type memRowIterator struct {
	*MemRowContainer
	curIdx int
}

var _ RowIterator = &memRowIterator{}

func (mc *MemRowContainer) NewIterator(_ context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569686)
	return &memRowIterator{MemRowContainer: mc}
}

func (i *memRowIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569687)
	i.curIdx = 0
}

func (i *memRowIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(569688)
	return i.curIdx < i.Len(), nil
}

func (i *memRowIterator) Next() {
	__antithesis_instrumentation__.Notify(569689)
	i.curIdx++
}

func (i *memRowIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569690)
	return i.EncRow(i.curIdx), nil
}

func (i *memRowIterator) Close() { __antithesis_instrumentation__.Notify(569691) }

type memRowFinalIterator struct {
	*MemRowContainer

	ctx context.Context
}

func (mc *MemRowContainer) NewFinalIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569692)
	return memRowFinalIterator{MemRowContainer: mc, ctx: ctx}
}

func (mc *MemRowContainer) GetRow(ctx context.Context, pos int) (tree.IndexedRow, error) {
	__antithesis_instrumentation__.Notify(569693)
	return IndexedRow{Idx: pos, Row: mc.EncRow(pos)}, nil
}

var _ RowIterator = memRowFinalIterator{}

func (i memRowFinalIterator) Rewind() { __antithesis_instrumentation__.Notify(569694) }

func (i memRowFinalIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(569695)
	return i.Len() > 0, nil
}

func (i memRowFinalIterator) Next() {
	__antithesis_instrumentation__.Notify(569696)
	i.PopFirst(i.ctx)
}

func (i memRowFinalIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569697)
	return i.EncRow(0), nil
}

func (i memRowFinalIterator) Close() { __antithesis_instrumentation__.Notify(569698) }

type DiskBackedRowContainer struct {
	src ReorderableRowContainer

	mrc *MemRowContainer
	drc *DiskRowContainer

	deDuplicate bool
	keyToIndex  map[string]int

	encodings  []descpb.DatumEncoding
	datumAlloc tree.DatumAlloc
	scratchKey []byte

	spilled bool

	engine      diskmap.Factory
	diskMonitor *mon.BytesMonitor
}

var _ ReorderableRowContainer = &DiskBackedRowContainer{}
var _ DeDupingRowContainer = &DiskBackedRowContainer{}

func (f *DiskBackedRowContainer) Init(
	ordering colinfo.ColumnOrdering,
	types []*types.T,
	evalCtx *tree.EvalContext,
	engine diskmap.Factory,
	memoryMonitor *mon.BytesMonitor,
	diskMonitor *mon.BytesMonitor,
) {
	__antithesis_instrumentation__.Notify(569699)
	mrc := MemRowContainer{}
	mrc.InitWithMon(ordering, types, evalCtx, memoryMonitor)
	f.mrc = &mrc
	f.src = &mrc
	f.engine = engine
	f.diskMonitor = diskMonitor
	f.encodings = make([]descpb.DatumEncoding, len(ordering))
	for i, orderInfo := range ordering {
		__antithesis_instrumentation__.Notify(569700)
		f.encodings[i] = rowenc.EncodingDirToDatumEncoding(orderInfo.Direction)
	}
}

func (f *DiskBackedRowContainer) DoDeDuplicate() {
	__antithesis_instrumentation__.Notify(569701)
	f.deDuplicate = true
	f.keyToIndex = make(map[string]int)
}

func (f *DiskBackedRowContainer) Len() int {
	__antithesis_instrumentation__.Notify(569702)
	return f.src.Len()
}

func (f *DiskBackedRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569703)
	if err := f.src.AddRow(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(569705)
		if spilled, spillErr := f.spillIfMemErr(ctx, err); !spilled && func() bool {
			__antithesis_instrumentation__.Notify(569707)
			return spillErr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(569708)

			return err
		} else {
			__antithesis_instrumentation__.Notify(569709)
			if spillErr != nil {
				__antithesis_instrumentation__.Notify(569710)

				return spillErr
			} else {
				__antithesis_instrumentation__.Notify(569711)
			}
		}
		__antithesis_instrumentation__.Notify(569706)

		return f.src.AddRow(ctx, row)
	} else {
		__antithesis_instrumentation__.Notify(569712)
	}
	__antithesis_instrumentation__.Notify(569704)
	return nil
}

func (f *DiskBackedRowContainer) AddRowWithDeDup(
	ctx context.Context, row rowenc.EncDatumRow,
) (int, error) {
	__antithesis_instrumentation__.Notify(569713)
	if !f.UsingDisk() {
		__antithesis_instrumentation__.Notify(569715)
		if err := f.encodeKey(ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(569720)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(569721)
		}
		__antithesis_instrumentation__.Notify(569716)
		encodedStr := string(f.scratchKey)
		idx, ok := f.keyToIndex[encodedStr]
		if ok {
			__antithesis_instrumentation__.Notify(569722)
			return idx, nil
		} else {
			__antithesis_instrumentation__.Notify(569723)
		}
		__antithesis_instrumentation__.Notify(569717)
		idx = f.Len()
		if err := f.AddRow(ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(569724)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(569725)
		}
		__antithesis_instrumentation__.Notify(569718)

		if !f.UsingDisk() {
			__antithesis_instrumentation__.Notify(569726)
			f.keyToIndex[encodedStr] = idx
		} else {
			__antithesis_instrumentation__.Notify(569727)
		}
		__antithesis_instrumentation__.Notify(569719)
		return idx, nil
	} else {
		__antithesis_instrumentation__.Notify(569728)
	}
	__antithesis_instrumentation__.Notify(569714)

	return f.drc.AddRowWithDeDup(ctx, row)
}

func (f *DiskBackedRowContainer) encodeKey(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569729)
	if len(row) != len(f.mrc.types) {
		__antithesis_instrumentation__.Notify(569732)
		log.Fatalf(ctx, "invalid row length %d, expected %d", len(row), len(f.mrc.types))
	} else {
		__antithesis_instrumentation__.Notify(569733)
	}
	__antithesis_instrumentation__.Notify(569730)
	f.scratchKey = f.scratchKey[:0]
	for i, orderInfo := range f.mrc.ordering {
		__antithesis_instrumentation__.Notify(569734)
		col := orderInfo.ColIdx
		var err error
		f.scratchKey, err = row[col].Encode(f.mrc.types[col], &f.datumAlloc, f.encodings[i], f.scratchKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(569735)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569736)
		}
	}
	__antithesis_instrumentation__.Notify(569731)
	return nil
}

func (f *DiskBackedRowContainer) Sort(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569737)
	f.src.Sort(ctx)
}

func (f *DiskBackedRowContainer) Reorder(
	ctx context.Context, ordering colinfo.ColumnOrdering,
) error {
	__antithesis_instrumentation__.Notify(569738)
	return f.src.Reorder(ctx, ordering)
}

func (f *DiskBackedRowContainer) InitTopK() {
	__antithesis_instrumentation__.Notify(569739)
	f.src.InitTopK()
}

func (f *DiskBackedRowContainer) MaybeReplaceMax(
	ctx context.Context, row rowenc.EncDatumRow,
) error {
	__antithesis_instrumentation__.Notify(569740)
	return f.src.MaybeReplaceMax(ctx, row)
}

func (f *DiskBackedRowContainer) NewIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569741)
	return f.src.NewIterator(ctx)
}

func (f *DiskBackedRowContainer) NewFinalIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569742)
	return f.src.NewFinalIterator(ctx)
}

func (f *DiskBackedRowContainer) UnsafeReset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569743)
	if f.deDuplicate {
		__antithesis_instrumentation__.Notify(569746)
		f.keyToIndex = make(map[string]int)
	} else {
		__antithesis_instrumentation__.Notify(569747)
	}
	__antithesis_instrumentation__.Notify(569744)
	if f.drc != nil {
		__antithesis_instrumentation__.Notify(569748)
		f.drc.Close(ctx)
		f.src = f.mrc
		f.drc = nil
		return nil
	} else {
		__antithesis_instrumentation__.Notify(569749)
	}
	__antithesis_instrumentation__.Notify(569745)
	return f.mrc.UnsafeReset(ctx)
}

func (f *DiskBackedRowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569750)
	if f.drc != nil {
		__antithesis_instrumentation__.Notify(569752)
		f.drc.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(569753)
	}
	__antithesis_instrumentation__.Notify(569751)
	f.mrc.Close(ctx)
	if f.deDuplicate {
		__antithesis_instrumentation__.Notify(569754)
		f.keyToIndex = nil
	} else {
		__antithesis_instrumentation__.Notify(569755)
	}
}

func (f *DiskBackedRowContainer) Spilled() bool {
	__antithesis_instrumentation__.Notify(569756)
	return f.spilled
}

func (f *DiskBackedRowContainer) UsingDisk() bool {
	__antithesis_instrumentation__.Notify(569757)
	return f.drc != nil
}

func (f *DiskBackedRowContainer) spillIfMemErr(ctx context.Context, err error) (bool, error) {
	__antithesis_instrumentation__.Notify(569758)
	if !sqlerrors.IsOutOfMemoryError(err) {
		__antithesis_instrumentation__.Notify(569761)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(569762)
	}
	__antithesis_instrumentation__.Notify(569759)
	if spillErr := f.SpillToDisk(ctx); spillErr != nil {
		__antithesis_instrumentation__.Notify(569763)
		return false, spillErr
	} else {
		__antithesis_instrumentation__.Notify(569764)
	}
	__antithesis_instrumentation__.Notify(569760)
	log.VEventf(ctx, 2, "spilled to disk: %v", err)
	return true, nil
}

func (f *DiskBackedRowContainer) SpillToDisk(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569765)
	if f.UsingDisk() {
		__antithesis_instrumentation__.Notify(569769)
		return errors.New("already using disk")
	} else {
		__antithesis_instrumentation__.Notify(569770)
	}
	__antithesis_instrumentation__.Notify(569766)
	drc := MakeDiskRowContainer(f.diskMonitor, f.mrc.types, f.mrc.ordering, f.engine)
	if f.deDuplicate {
		__antithesis_instrumentation__.Notify(569771)
		drc.DoDeDuplicate()

		f.keyToIndex = nil
	} else {
		__antithesis_instrumentation__.Notify(569772)
	}
	__antithesis_instrumentation__.Notify(569767)
	i := f.mrc.NewFinalIterator(ctx)
	defer i.Close()
	for i.Rewind(); ; i.Next() {
		__antithesis_instrumentation__.Notify(569773)
		if ok, err := i.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(569776)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569777)
			if !ok {
				__antithesis_instrumentation__.Notify(569778)
				break
			} else {
				__antithesis_instrumentation__.Notify(569779)
			}
		}
		__antithesis_instrumentation__.Notify(569774)
		memRow, err := i.Row()
		if err != nil {
			__antithesis_instrumentation__.Notify(569780)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569781)
		}
		__antithesis_instrumentation__.Notify(569775)
		if err := drc.AddRow(ctx, memRow); err != nil {
			__antithesis_instrumentation__.Notify(569782)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569783)
		}
	}
	__antithesis_instrumentation__.Notify(569768)
	f.mrc.Clear(ctx)

	f.src = &drc
	f.drc = &drc
	f.spilled = true
	return nil
}

type DiskBackedIndexedRowContainer struct {
	*DiskBackedRowContainer

	scratchEncRow rowenc.EncDatumRow
	storedTypes   []*types.T
	datumAlloc    tree.DatumAlloc
	rowAlloc      rowenc.EncDatumRowAlloc
	idx           uint64

	diskRowIter RowIterator
	idxRowIter  int

	firstCachedRowPos int
	nextPosToCache    int

	indexedRowsCache ring.Buffer

	maxCacheSize int
	cacheMemAcc  mon.BoundAccount
	hitCount     int
	missCount    int

	DisableCache bool
}

var _ IndexedRowContainer = &DiskBackedIndexedRowContainer{}

func NewDiskBackedIndexedRowContainer(
	ordering colinfo.ColumnOrdering,
	typs []*types.T,
	evalCtx *tree.EvalContext,
	engine diskmap.Factory,
	memoryMonitor *mon.BytesMonitor,
	diskMonitor *mon.BytesMonitor,
) *DiskBackedIndexedRowContainer {
	__antithesis_instrumentation__.Notify(569784)
	d := DiskBackedIndexedRowContainer{}

	d.storedTypes = make([]*types.T, len(typs)+1)
	copy(d.storedTypes, typs)
	d.storedTypes[len(d.storedTypes)-1] = types.Int
	d.scratchEncRow = make(rowenc.EncDatumRow, len(d.storedTypes))
	d.DiskBackedRowContainer = &DiskBackedRowContainer{}
	d.DiskBackedRowContainer.Init(ordering, d.storedTypes, evalCtx, engine, memoryMonitor, diskMonitor)
	d.maxCacheSize = maxIndexedRowsCacheSize
	d.cacheMemAcc = memoryMonitor.MakeBoundAccount()
	return &d
}

func (f *DiskBackedIndexedRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569785)
	copy(f.scratchEncRow, row)
	f.scratchEncRow[len(f.scratchEncRow)-1] = rowenc.DatumToEncDatum(
		types.Int,
		tree.NewDInt(tree.DInt(f.idx)),
	)
	f.idx++
	return f.DiskBackedRowContainer.AddRow(ctx, f.scratchEncRow)
}

func (f *DiskBackedIndexedRowContainer) Reorder(
	ctx context.Context, ordering colinfo.ColumnOrdering,
) error {
	__antithesis_instrumentation__.Notify(569786)
	if err := f.DiskBackedRowContainer.Reorder(ctx, ordering); err != nil {
		__antithesis_instrumentation__.Notify(569788)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569789)
	}
	__antithesis_instrumentation__.Notify(569787)
	f.resetCache(ctx)
	f.resetIterator()
	return nil
}

func (f *DiskBackedIndexedRowContainer) resetCache(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569790)
	f.firstCachedRowPos = 0
	f.nextPosToCache = 0
	f.indexedRowsCache.Reset()
	f.cacheMemAcc.Clear(ctx)
}

func (f *DiskBackedIndexedRowContainer) resetIterator() {
	__antithesis_instrumentation__.Notify(569791)
	if f.diskRowIter != nil {
		__antithesis_instrumentation__.Notify(569792)
		f.diskRowIter.Close()
		f.diskRowIter = nil
		f.idxRowIter = 0
	} else {
		__antithesis_instrumentation__.Notify(569793)
	}
}

func (f *DiskBackedIndexedRowContainer) UnsafeReset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569794)
	f.resetCache(ctx)
	f.resetIterator()
	f.idx = 0
	return f.DiskBackedRowContainer.UnsafeReset(ctx)
}

func (f *DiskBackedIndexedRowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569795)
	if f.diskRowIter != nil {
		__antithesis_instrumentation__.Notify(569797)
		f.diskRowIter.Close()
	} else {
		__antithesis_instrumentation__.Notify(569798)
	}
	__antithesis_instrumentation__.Notify(569796)
	f.cacheMemAcc.Close(ctx)
	f.DiskBackedRowContainer.Close(ctx)
}

const maxIndexedRowsCacheSize = 4096

func (f *DiskBackedIndexedRowContainer) GetRow(
	ctx context.Context, pos int,
) (tree.IndexedRow, error) {
	__antithesis_instrumentation__.Notify(569799)
	var rowWithIdx rowenc.EncDatumRow
	var err error
	if f.UsingDisk() {
		__antithesis_instrumentation__.Notify(569802)
		if f.DisableCache {
			__antithesis_instrumentation__.Notify(569807)
			return f.getRowWithoutCache(ctx, pos), nil
		} else {
			__antithesis_instrumentation__.Notify(569808)
		}
		__antithesis_instrumentation__.Notify(569803)

		if pos >= f.firstCachedRowPos && func() bool {
			__antithesis_instrumentation__.Notify(569809)
			return pos < f.nextPosToCache == true
		}() == true {
			__antithesis_instrumentation__.Notify(569810)
			requestedRowCachePos := pos - f.firstCachedRowPos
			f.hitCount++
			return f.indexedRowsCache.Get(requestedRowCachePos).(tree.IndexedRow), nil
		} else {
			__antithesis_instrumentation__.Notify(569811)
		}
		__antithesis_instrumentation__.Notify(569804)
		f.missCount++
		if f.diskRowIter == nil {
			__antithesis_instrumentation__.Notify(569812)
			f.diskRowIter = f.DiskBackedRowContainer.drc.NewIterator(ctx)
			f.diskRowIter.Rewind()
		} else {
			__antithesis_instrumentation__.Notify(569813)
		}
		__antithesis_instrumentation__.Notify(569805)
		if f.idxRowIter > pos {
			__antithesis_instrumentation__.Notify(569814)

			log.VEventf(ctx, 1, "rewinding: cache contains indices [%d, %d) but index %d requested", f.firstCachedRowPos, f.nextPosToCache, pos)
			f.idxRowIter = 0
			f.diskRowIter.Rewind()
			f.resetCache(ctx)
			if pos-maxIndexedRowsCacheSize > f.nextPosToCache {
				__antithesis_instrumentation__.Notify(569815)

				f.nextPosToCache = pos - maxIndexedRowsCacheSize
				f.firstCachedRowPos = f.nextPosToCache
			} else {
				__antithesis_instrumentation__.Notify(569816)
			}
		} else {
			__antithesis_instrumentation__.Notify(569817)
		}
		__antithesis_instrumentation__.Notify(569806)
		for ; ; f.diskRowIter.Next() {
			__antithesis_instrumentation__.Notify(569818)
			if ok, err := f.diskRowIter.Valid(); err != nil {
				__antithesis_instrumentation__.Notify(569821)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(569822)
				if !ok {
					__antithesis_instrumentation__.Notify(569823)
					return nil, errors.Errorf("row at pos %d not found", pos)
				} else {
					__antithesis_instrumentation__.Notify(569824)
				}
			}
			__antithesis_instrumentation__.Notify(569819)
			if f.idxRowIter == f.nextPosToCache {
				__antithesis_instrumentation__.Notify(569825)
				rowWithIdx, err = f.diskRowIter.Row()
				if err != nil {
					__antithesis_instrumentation__.Notify(569829)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(569830)
				}
				__antithesis_instrumentation__.Notify(569826)
				for i := range rowWithIdx {
					__antithesis_instrumentation__.Notify(569831)
					if err := rowWithIdx[i].EnsureDecoded(f.storedTypes[i], &f.datumAlloc); err != nil {
						__antithesis_instrumentation__.Notify(569832)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(569833)
					}
				}
				__antithesis_instrumentation__.Notify(569827)
				row, rowIdx := rowWithIdx[:len(rowWithIdx)-1], rowWithIdx[len(rowWithIdx)-1].Datum
				if idx, ok := rowIdx.(*tree.DInt); ok {
					__antithesis_instrumentation__.Notify(569834)
					if f.indexedRowsCache.Len() == f.maxCacheSize {
						__antithesis_instrumentation__.Notify(569836)

						if err := f.reuseFirstRowInCache(ctx, int(*idx), row); err != nil {
							__antithesis_instrumentation__.Notify(569837)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(569838)
						}
					} else {
						__antithesis_instrumentation__.Notify(569839)

						usage := memsize.Int + int64(row.Size())
						if err := f.cacheMemAcc.Grow(ctx, usage); err != nil {
							__antithesis_instrumentation__.Notify(569840)
							if sqlerrors.IsOutOfMemoryError(err) {
								__antithesis_instrumentation__.Notify(569841)

								if f.indexedRowsCache.Len() == 0 {
									__antithesis_instrumentation__.Notify(569843)

									return nil, err
								} else {
									__antithesis_instrumentation__.Notify(569844)
								}
								__antithesis_instrumentation__.Notify(569842)
								f.maxCacheSize = f.indexedRowsCache.Len()
								if err := f.reuseFirstRowInCache(ctx, int(*idx), row); err != nil {
									__antithesis_instrumentation__.Notify(569845)
									return nil, err
								} else {
									__antithesis_instrumentation__.Notify(569846)
								}
							} else {
								__antithesis_instrumentation__.Notify(569847)
								return nil, err
							}
						} else {
							__antithesis_instrumentation__.Notify(569848)

							ir := IndexedRow{int(*idx), f.rowAlloc.CopyRow(row)}
							f.indexedRowsCache.AddLast(ir)
						}
					}
					__antithesis_instrumentation__.Notify(569835)
					f.nextPosToCache++
				} else {
					__antithesis_instrumentation__.Notify(569849)
					return nil, errors.Errorf("unexpected last column type: should be DInt but found %T", idx)
				}
				__antithesis_instrumentation__.Notify(569828)
				if f.idxRowIter == pos {
					__antithesis_instrumentation__.Notify(569850)
					return f.indexedRowsCache.GetLast().(tree.IndexedRow), nil
				} else {
					__antithesis_instrumentation__.Notify(569851)
				}
			} else {
				__antithesis_instrumentation__.Notify(569852)
			}
			__antithesis_instrumentation__.Notify(569820)
			f.idxRowIter++
		}
	} else {
		__antithesis_instrumentation__.Notify(569853)
	}
	__antithesis_instrumentation__.Notify(569800)
	rowWithIdx = f.DiskBackedRowContainer.mrc.EncRow(pos)
	row, rowIdx := rowWithIdx[:len(rowWithIdx)-1], rowWithIdx[len(rowWithIdx)-1].Datum
	if idx, ok := rowIdx.(*tree.DInt); ok {
		__antithesis_instrumentation__.Notify(569854)
		return IndexedRow{int(*idx), row}, nil
	} else {
		__antithesis_instrumentation__.Notify(569855)
	}
	__antithesis_instrumentation__.Notify(569801)
	return nil, errors.Errorf("unexpected last column type: should be DInt but found %T", rowIdx)
}

func (f *DiskBackedIndexedRowContainer) reuseFirstRowInCache(
	ctx context.Context, idx int, row rowenc.EncDatumRow,
) error {
	__antithesis_instrumentation__.Notify(569856)
	newRowSize := row.Size()
	for {
		__antithesis_instrumentation__.Notify(569857)
		if f.indexedRowsCache.Len() == 0 {
			__antithesis_instrumentation__.Notify(569860)
			return errors.Errorf("unexpectedly the cache of DiskBackedIndexedRowContainer contains zero rows")
		} else {
			__antithesis_instrumentation__.Notify(569861)
		}
		__antithesis_instrumentation__.Notify(569858)
		indexedRowToReuse := f.indexedRowsCache.GetFirst().(IndexedRow)
		oldRowSize := indexedRowToReuse.Row.Size()
		delta := int64(newRowSize - oldRowSize)
		if delta > 0 {
			__antithesis_instrumentation__.Notify(569862)

			if err := f.cacheMemAcc.Grow(ctx, delta); err != nil {
				__antithesis_instrumentation__.Notify(569863)
				if sqlerrors.IsOutOfMemoryError(err) {
					__antithesis_instrumentation__.Notify(569865)

					f.indexedRowsCache.RemoveFirst()
					f.cacheMemAcc.Shrink(ctx, int64(oldRowSize))
					f.maxCacheSize--
					f.firstCachedRowPos++
					if f.indexedRowsCache.Len() == 0 {
						__antithesis_instrumentation__.Notify(569867)
						return err
					} else {
						__antithesis_instrumentation__.Notify(569868)
					}
					__antithesis_instrumentation__.Notify(569866)
					continue
				} else {
					__antithesis_instrumentation__.Notify(569869)
				}
				__antithesis_instrumentation__.Notify(569864)
				return err
			} else {
				__antithesis_instrumentation__.Notify(569870)
			}
		} else {
			__antithesis_instrumentation__.Notify(569871)
			if delta < 0 {
				__antithesis_instrumentation__.Notify(569872)
				f.cacheMemAcc.Shrink(ctx, -delta)
			} else {
				__antithesis_instrumentation__.Notify(569873)
			}
		}
		__antithesis_instrumentation__.Notify(569859)
		indexedRowToReuse.Idx = idx
		copy(indexedRowToReuse.Row, row)
		f.indexedRowsCache.RemoveFirst()
		f.indexedRowsCache.AddLast(indexedRowToReuse)
		f.firstCachedRowPos++
		return nil
	}
}

func (f *DiskBackedIndexedRowContainer) getRowWithoutCache(
	ctx context.Context, pos int,
) tree.IndexedRow {
	__antithesis_instrumentation__.Notify(569874)
	if !f.UsingDisk() {
		__antithesis_instrumentation__.Notify(569878)
		panic(errors.Errorf("getRowWithoutCache is called when the container is using memory"))
	} else {
		__antithesis_instrumentation__.Notify(569879)
	}
	__antithesis_instrumentation__.Notify(569875)
	if f.diskRowIter == nil {
		__antithesis_instrumentation__.Notify(569880)
		f.diskRowIter = f.DiskBackedRowContainer.drc.NewIterator(ctx)
		f.diskRowIter.Rewind()
	} else {
		__antithesis_instrumentation__.Notify(569881)
	}
	__antithesis_instrumentation__.Notify(569876)
	if f.idxRowIter > pos {
		__antithesis_instrumentation__.Notify(569882)

		f.idxRowIter = 0
		f.diskRowIter.Rewind()
	} else {
		__antithesis_instrumentation__.Notify(569883)
	}
	__antithesis_instrumentation__.Notify(569877)
	for ; ; f.diskRowIter.Next() {
		__antithesis_instrumentation__.Notify(569884)
		if ok, err := f.diskRowIter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(569887)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(569888)
			if !ok {
				__antithesis_instrumentation__.Notify(569889)
				panic(errors.AssertionFailedf("row at pos %d not found", pos))
			} else {
				__antithesis_instrumentation__.Notify(569890)
			}
		}
		__antithesis_instrumentation__.Notify(569885)
		if f.idxRowIter == pos {
			__antithesis_instrumentation__.Notify(569891)
			rowWithIdx, err := f.diskRowIter.Row()
			if err != nil {
				__antithesis_instrumentation__.Notify(569895)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(569896)
			}
			__antithesis_instrumentation__.Notify(569892)
			for i := range rowWithIdx {
				__antithesis_instrumentation__.Notify(569897)
				if err := rowWithIdx[i].EnsureDecoded(f.storedTypes[i], &f.datumAlloc); err != nil {
					__antithesis_instrumentation__.Notify(569898)
					panic(err)
				} else {
					__antithesis_instrumentation__.Notify(569899)
				}
			}
			__antithesis_instrumentation__.Notify(569893)
			row, rowIdx := rowWithIdx[:len(rowWithIdx)-1], rowWithIdx[len(rowWithIdx)-1].Datum
			if idx, ok := rowIdx.(*tree.DInt); ok {
				__antithesis_instrumentation__.Notify(569900)
				return IndexedRow{int(*idx), f.rowAlloc.CopyRow(row)}
			} else {
				__antithesis_instrumentation__.Notify(569901)
			}
			__antithesis_instrumentation__.Notify(569894)
			panic(errors.Errorf("unexpected last column type: should be DInt but found %T", rowIdx))
		} else {
			__antithesis_instrumentation__.Notify(569902)
		}
		__antithesis_instrumentation__.Notify(569886)
		f.idxRowIter++
	}
}

type IndexedRow struct {
	Idx int
	Row rowenc.EncDatumRow
}

func (ir IndexedRow) GetIdx() int {
	__antithesis_instrumentation__.Notify(569903)
	return ir.Idx
}

func (ir IndexedRow) GetDatum(colIdx int) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(569904)
	return ir.Row[colIdx].Datum, nil
}

func (ir IndexedRow) GetDatums(startColIdx, endColIdx int) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(569905)
	datums := make(tree.Datums, 0, endColIdx-startColIdx)
	for idx := startColIdx; idx < endColIdx; idx++ {
		__antithesis_instrumentation__.Notify(569907)
		datums = append(datums, ir.Row[idx].Datum)
	}
	__antithesis_instrumentation__.Notify(569906)
	return datums, nil
}
