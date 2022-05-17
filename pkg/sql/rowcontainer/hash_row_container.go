package rowcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type columns []uint32

type RowMarkerIterator interface {
	RowIterator

	Reset(ctx context.Context, row rowenc.EncDatumRow) error
	Mark(ctx context.Context) error
	IsMarked(ctx context.Context) bool
}

type HashRowContainer interface {
	Init(
		ctx context.Context, shouldMark bool, types []*types.T, storedEqCols columns, encodeNull bool,
	) error
	AddRow(context.Context, rowenc.EncDatumRow) error

	IsEmpty() bool

	NewBucketIterator(
		ctx context.Context, row rowenc.EncDatumRow, probeEqCols columns,
	) (RowMarkerIterator, error)

	NewUnmarkedIterator(context.Context) RowIterator

	Close(context.Context)
}

type columnEncoder struct {
	scratch []byte

	keyTypes   []*types.T
	datumAlloc tree.DatumAlloc
	encodeNull bool
}

func (e *columnEncoder) init(typs []*types.T, keyCols columns, encodeNull bool) {
	e.keyTypes = make([]*types.T, len(keyCols))
	for i, c := range keyCols {
		e.keyTypes[i] = typs[c]
	}
	e.encodeNull = encodeNull
}

func encodeColumnsOfRow(
	da *tree.DatumAlloc,
	appendTo []byte,
	row rowenc.EncDatumRow,
	cols columns,
	colTypes []*types.T,
	encodeNull bool,
) (encoding []byte, hasNull bool, err error) {
	__antithesis_instrumentation__.Notify(569170)
	for i, colIdx := range cols {
		__antithesis_instrumentation__.Notify(569172)
		if row[colIdx].IsNull() && func() bool {
			__antithesis_instrumentation__.Notify(569174)
			return !encodeNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(569175)
			return nil, true, nil
		} else {
			__antithesis_instrumentation__.Notify(569176)
		}
		__antithesis_instrumentation__.Notify(569173)

		appendTo, err = row[colIdx].Encode(colTypes[i], da, descpb.DatumEncoding_ASCENDING_KEY, appendTo)
		if err != nil {
			__antithesis_instrumentation__.Notify(569177)
			return appendTo, false, err
		} else {
			__antithesis_instrumentation__.Notify(569178)
		}
	}
	__antithesis_instrumentation__.Notify(569171)
	return appendTo, false, nil
}

func (e *columnEncoder) encodeEqualityCols(
	ctx context.Context, row rowenc.EncDatumRow, eqCols columns,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(569179)
	encoded, hasNull, err := encodeColumnsOfRow(
		&e.datumAlloc, e.scratch, row, eqCols, e.keyTypes, e.encodeNull,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(569182)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569183)
	}
	__antithesis_instrumentation__.Notify(569180)
	e.scratch = encoded[:0]
	if hasNull {
		__antithesis_instrumentation__.Notify(569184)
		log.Fatal(ctx, "cannot process rows with NULL in an equality column")
	} else {
		__antithesis_instrumentation__.Notify(569185)
	}
	__antithesis_instrumentation__.Notify(569181)
	return encoded, nil
}

func storedEqColsToOrdering(storedEqCols columns) colinfo.ColumnOrdering {
	__antithesis_instrumentation__.Notify(569186)
	ordering := make(colinfo.ColumnOrdering, len(storedEqCols))
	for i := range ordering {
		__antithesis_instrumentation__.Notify(569188)
		ordering[i] = colinfo.ColumnOrderInfo{
			ColIdx:    int(storedEqCols[i]),
			Direction: encoding.Ascending,
		}
	}
	__antithesis_instrumentation__.Notify(569187)
	return ordering
}

type HashMemRowContainer struct {
	*MemRowContainer
	columnEncoder

	shouldMark bool

	marked []bool

	markMemoryReserved bool

	buckets map[string][]int

	bucketsAcc mon.BoundAccount

	storedEqCols columns
}

var _ HashRowContainer = &HashMemRowContainer{}

func MakeHashMemRowContainer(
	evalCtx *tree.EvalContext, memMonitor *mon.BytesMonitor, typs []*types.T, storedEqCols columns,
) HashMemRowContainer {
	__antithesis_instrumentation__.Notify(569189)
	mrc := &MemRowContainer{}
	mrc.InitWithMon(storedEqColsToOrdering(storedEqCols), typs, evalCtx, memMonitor)
	return HashMemRowContainer{
		MemRowContainer: mrc,
		buckets:         make(map[string][]int),
		bucketsAcc:      memMonitor.MakeBoundAccount(),
	}
}

func (h *HashMemRowContainer) Init(
	_ context.Context, shouldMark bool, typs []*types.T, storedEqCols columns, encodeNull bool,
) error {
	__antithesis_instrumentation__.Notify(569190)
	if h.storedEqCols != nil {
		__antithesis_instrumentation__.Notify(569192)
		return errors.New("HashMemRowContainer has already been initialized")
	} else {
		__antithesis_instrumentation__.Notify(569193)
	}
	__antithesis_instrumentation__.Notify(569191)
	h.columnEncoder.init(typs, storedEqCols, encodeNull)
	h.shouldMark = shouldMark
	h.storedEqCols = storedEqCols
	return nil
}

func (h *HashMemRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569194)
	rowIdx := h.Len()

	if err := h.addRowToBucket(ctx, row, rowIdx); err != nil {
		__antithesis_instrumentation__.Notify(569196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569197)
	}
	__antithesis_instrumentation__.Notify(569195)
	return h.MemRowContainer.AddRow(ctx, row)
}

func (h *HashMemRowContainer) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(569198)
	return h.Len() == 0
}

func (h *HashMemRowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569199)
	h.MemRowContainer.Close(ctx)
	h.bucketsAcc.Close(ctx)
}

func (h *HashMemRowContainer) addRowToBucket(
	ctx context.Context, row rowenc.EncDatumRow, rowIdx int,
) error {
	__antithesis_instrumentation__.Notify(569200)
	encoded, err := h.encodeEqualityCols(ctx, row, h.storedEqCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(569204)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569205)
	}
	__antithesis_instrumentation__.Notify(569201)

	bucket, ok := h.buckets[string(encoded)]

	usage := memsize.Int
	if !ok {
		__antithesis_instrumentation__.Notify(569206)
		usage += int64(len(encoded))
		usage += memsize.IntSliceOverhead
	} else {
		__antithesis_instrumentation__.Notify(569207)
	}
	__antithesis_instrumentation__.Notify(569202)

	if err := h.bucketsAcc.Grow(ctx, usage); err != nil {
		__antithesis_instrumentation__.Notify(569208)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569209)
	}
	__antithesis_instrumentation__.Notify(569203)

	h.buckets[string(encoded)] = append(bucket, rowIdx)
	return nil
}

func (h *HashMemRowContainer) ReserveMarkMemoryMaybe(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569210)
	if h.markMemoryReserved {
		__antithesis_instrumentation__.Notify(569213)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(569214)
	}
	__antithesis_instrumentation__.Notify(569211)
	if err := h.bucketsAcc.Grow(ctx, memsize.BoolSliceOverhead+(memsize.Bool*int64(h.Len()))); err != nil {
		__antithesis_instrumentation__.Notify(569215)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569216)
	}
	__antithesis_instrumentation__.Notify(569212)
	h.markMemoryReserved = true
	return nil
}

type hashMemRowBucketIterator struct {
	*HashMemRowContainer
	probeEqCols columns

	rowIdxs []int
	curIdx  int
}

var _ RowMarkerIterator = &hashMemRowBucketIterator{}

func (h *HashMemRowContainer) NewBucketIterator(
	ctx context.Context, row rowenc.EncDatumRow, probeEqCols columns,
) (RowMarkerIterator, error) {
	__antithesis_instrumentation__.Notify(569217)
	ret := &hashMemRowBucketIterator{
		HashMemRowContainer: h,
		probeEqCols:         probeEqCols,
	}

	if err := ret.Reset(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(569219)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569220)
	}
	__antithesis_instrumentation__.Notify(569218)
	return ret, nil
}

func (i *hashMemRowBucketIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569221)
	i.curIdx = 0
}

func (i *hashMemRowBucketIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(569222)
	return i.curIdx < len(i.rowIdxs), nil
}

func (i *hashMemRowBucketIterator) Next() {
	__antithesis_instrumentation__.Notify(569223)
	i.curIdx++
}

func (i *hashMemRowBucketIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569224)
	return i.EncRow(i.rowIdxs[i.curIdx]), nil
}

func (i *hashMemRowBucketIterator) IsMarked(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(569225)
	if !i.shouldMark {
		__antithesis_instrumentation__.Notify(569228)
		log.Fatal(ctx, "hash mem row container not set up for marking")
	} else {
		__antithesis_instrumentation__.Notify(569229)
	}
	__antithesis_instrumentation__.Notify(569226)
	if i.marked == nil {
		__antithesis_instrumentation__.Notify(569230)
		return false
	} else {
		__antithesis_instrumentation__.Notify(569231)
	}
	__antithesis_instrumentation__.Notify(569227)

	return i.marked[i.rowIdxs[i.curIdx]]
}

func (i *hashMemRowBucketIterator) Mark(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569232)
	if !i.shouldMark {
		__antithesis_instrumentation__.Notify(569235)
		log.Fatal(ctx, "hash mem row container not set up for marking")
	} else {
		__antithesis_instrumentation__.Notify(569236)
	}
	__antithesis_instrumentation__.Notify(569233)
	if i.marked == nil {
		__antithesis_instrumentation__.Notify(569237)
		if !i.markMemoryReserved {
			__antithesis_instrumentation__.Notify(569239)
			panic("mark memory should have been reserved already")
		} else {
			__antithesis_instrumentation__.Notify(569240)
		}
		__antithesis_instrumentation__.Notify(569238)
		i.marked = make([]bool, i.Len())
	} else {
		__antithesis_instrumentation__.Notify(569241)
	}
	__antithesis_instrumentation__.Notify(569234)

	i.marked[i.rowIdxs[i.curIdx]] = true
	return nil
}

func (i *hashMemRowBucketIterator) Reset(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569242)
	encoded, err := i.encodeEqualityCols(ctx, row, i.probeEqCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(569244)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569245)
	}
	__antithesis_instrumentation__.Notify(569243)
	i.rowIdxs = i.buckets[string(encoded)]
	return nil
}

func (i *hashMemRowBucketIterator) Close() { __antithesis_instrumentation__.Notify(569246) }

type hashMemRowIterator struct {
	*HashMemRowContainer
	curIdx int

	curKey []byte
}

var _ RowIterator = &hashMemRowIterator{}

func (h *HashMemRowContainer) NewUnmarkedIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569247)
	return &hashMemRowIterator{HashMemRowContainer: h}
}

func (i *hashMemRowIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569248)
	i.curIdx = -1

	i.Next()
}

func (i *hashMemRowIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(569249)
	return i.curIdx < i.Len(), nil
}

func (i *hashMemRowIterator) computeKey() error {
	__antithesis_instrumentation__.Notify(569250)
	valid, err := i.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(569254)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569255)
	}
	__antithesis_instrumentation__.Notify(569251)

	var row rowenc.EncDatumRow
	if valid {
		__antithesis_instrumentation__.Notify(569256)
		row = i.EncRow(i.curIdx)
	} else {
		__antithesis_instrumentation__.Notify(569257)
		if i.curIdx == 0 {
			__antithesis_instrumentation__.Notify(569259)

			i.curKey = nil
			return nil
		} else {
			__antithesis_instrumentation__.Notify(569260)
		}
		__antithesis_instrumentation__.Notify(569258)

		row = i.EncRow(i.curIdx - 1)
	}
	__antithesis_instrumentation__.Notify(569252)

	i.curKey = i.curKey[:0]
	for _, col := range i.storedEqCols {
		__antithesis_instrumentation__.Notify(569261)
		var err error
		i.curKey, err = row[col].Encode(i.types[col], &i.columnEncoder.datumAlloc, descpb.DatumEncoding_ASCENDING_KEY, i.curKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(569262)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569263)
		}
	}
	__antithesis_instrumentation__.Notify(569253)
	i.curKey = encoding.EncodeUvarintAscending(i.curKey, uint64(i.curIdx))
	return nil
}

func (i *hashMemRowIterator) Next() {
	__antithesis_instrumentation__.Notify(569264)

	i.curIdx++
	if i.marked != nil {
		__antithesis_instrumentation__.Notify(569265)
		for ; i.curIdx < len(i.marked) && func() bool {
			__antithesis_instrumentation__.Notify(569266)
			return i.marked[i.curIdx] == true
		}() == true; i.curIdx++ {
			__antithesis_instrumentation__.Notify(569267)
		}
	} else {
		__antithesis_instrumentation__.Notify(569268)
	}
}

func (i *hashMemRowIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569269)
	return i.EncRow(i.curIdx), nil
}

func (i *hashMemRowIterator) Close() { __antithesis_instrumentation__.Notify(569270) }

type HashDiskRowContainer struct {
	DiskRowContainer
	columnEncoder

	diskMonitor *mon.BytesMonitor

	shouldMark    bool
	engine        diskmap.Factory
	scratchEncRow rowenc.EncDatumRow
}

var _ HashRowContainer = &HashDiskRowContainer{}

var encodedTrue = encoding.EncodeBoolValue(nil, encoding.NoColumnID, true)

func MakeHashDiskRowContainer(
	diskMonitor *mon.BytesMonitor, e diskmap.Factory,
) HashDiskRowContainer {
	__antithesis_instrumentation__.Notify(569271)
	return HashDiskRowContainer{
		diskMonitor: diskMonitor,
		engine:      e,
	}
}

func (h *HashDiskRowContainer) Init(
	_ context.Context, shouldMark bool, typs []*types.T, storedEqCols columns, encodeNull bool,
) error {
	__antithesis_instrumentation__.Notify(569272)
	h.columnEncoder.init(typs, storedEqCols, encodeNull)
	h.shouldMark = shouldMark
	storedTypes := typs
	if h.shouldMark {
		__antithesis_instrumentation__.Notify(569274)

		storedTypes = make([]*types.T, len(typs)+1)
		copy(storedTypes, typs)
		storedTypes[len(storedTypes)-1] = types.Bool

		h.scratchEncRow = make(rowenc.EncDatumRow, len(storedTypes))

		h.scratchEncRow[len(h.scratchEncRow)-1] = rowenc.DatumToEncDatum(
			types.Bool,
			tree.MakeDBool(false),
		)
	} else {
		__antithesis_instrumentation__.Notify(569275)
	}
	__antithesis_instrumentation__.Notify(569273)

	h.DiskRowContainer = MakeDiskRowContainer(h.diskMonitor, storedTypes, storedEqColsToOrdering(storedEqCols), h.engine)
	return nil
}

func (h *HashDiskRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569276)
	var err error
	if h.shouldMark {
		__antithesis_instrumentation__.Notify(569278)

		copy(h.scratchEncRow, row)
		err = h.DiskRowContainer.AddRow(ctx, h.scratchEncRow)
	} else {
		__antithesis_instrumentation__.Notify(569279)
		err = h.DiskRowContainer.AddRow(ctx, row)
	}
	__antithesis_instrumentation__.Notify(569277)
	return err
}

func (h *HashDiskRowContainer) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(569280)
	return h.DiskRowContainer.Len() == 0
}

type hashDiskRowBucketIterator struct {
	*diskRowIterator
	*HashDiskRowContainer
	probeEqCols columns

	haveMarkedRows bool

	encodedEqCols []byte

	tmpBuf []byte
}

var _ RowMarkerIterator = &hashDiskRowBucketIterator{}

func (h *HashDiskRowContainer) NewBucketIterator(
	ctx context.Context, row rowenc.EncDatumRow, probeEqCols columns,
) (RowMarkerIterator, error) {
	__antithesis_instrumentation__.Notify(569281)
	ret := &hashDiskRowBucketIterator{
		HashDiskRowContainer: h,
		probeEqCols:          probeEqCols,
		diskRowIterator:      h.NewIterator(ctx).(*diskRowIterator),
	}
	if err := ret.Reset(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(569283)
		ret.Close()
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569284)
	}
	__antithesis_instrumentation__.Notify(569282)
	return ret, nil
}

func (i *hashDiskRowBucketIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569285)
	i.SeekGE(i.encodedEqCols)
}

func (i *hashDiskRowBucketIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(569286)
	ok, err := i.diskRowIterator.Valid()
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(569288)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(569289)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(569290)
	}
	__antithesis_instrumentation__.Notify(569287)

	return bytes.HasPrefix(i.UnsafeKey(), i.encodedEqCols), nil
}

func (i *hashDiskRowBucketIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569291)
	row, err := i.diskRowIterator.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(569294)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569295)
	}
	__antithesis_instrumentation__.Notify(569292)

	if i.HashDiskRowContainer.shouldMark {
		__antithesis_instrumentation__.Notify(569296)
		row = row[:len(row)-1]
	} else {
		__antithesis_instrumentation__.Notify(569297)
	}
	__antithesis_instrumentation__.Notify(569293)
	return row, nil
}

func (i *hashDiskRowBucketIterator) Reset(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569298)
	encoded, err := i.HashDiskRowContainer.encodeEqualityCols(ctx, row, i.probeEqCols)
	if err != nil {
		__antithesis_instrumentation__.Notify(569301)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569302)
	}
	__antithesis_instrumentation__.Notify(569299)
	i.encodedEqCols = append(i.encodedEqCols[:0], encoded...)
	if i.haveMarkedRows {
		__antithesis_instrumentation__.Notify(569303)

		i.haveMarkedRows = false
		i.diskRowIterator.Close()
		i.diskRowIterator = i.HashDiskRowContainer.NewIterator(ctx).(*diskRowIterator)
	} else {
		__antithesis_instrumentation__.Notify(569304)
	}
	__antithesis_instrumentation__.Notify(569300)
	return nil
}

func (i *hashDiskRowBucketIterator) IsMarked(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(569305)
	if !i.HashDiskRowContainer.shouldMark {
		__antithesis_instrumentation__.Notify(569308)
		log.Fatal(ctx, "hash disk row container not set up for marking")
	} else {
		__antithesis_instrumentation__.Notify(569309)
	}
	__antithesis_instrumentation__.Notify(569306)
	ok, err := i.diskRowIterator.Valid()
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(569310)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(569311)
		return false
	} else {
		__antithesis_instrumentation__.Notify(569312)
	}
	__antithesis_instrumentation__.Notify(569307)

	rowVal := i.UnsafeValue()
	return bytes.Equal(rowVal[len(rowVal)-len(encodedTrue):], encodedTrue)
}

func (i *hashDiskRowBucketIterator) Mark(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569313)
	if !i.HashDiskRowContainer.shouldMark {
		__antithesis_instrumentation__.Notify(569315)
		log.Fatal(ctx, "hash disk row container not set up for marking")
	} else {
		__antithesis_instrumentation__.Notify(569316)
	}
	__antithesis_instrumentation__.Notify(569314)
	i.haveMarkedRows = true
	markBytes := encodedTrue

	rowVal := append(i.tmpBuf[:0], i.UnsafeValue()...)
	originalLen := len(rowVal)
	rowVal = append(rowVal, markBytes...)

	copy(rowVal[originalLen-len(markBytes):], rowVal[originalLen:])
	rowVal = rowVal[:originalLen]
	i.tmpBuf = rowVal

	return i.HashDiskRowContainer.bufferedRows.Put(i.UnsafeKey(), rowVal)
}

type hashDiskRowIterator struct {
	*diskRowIterator
}

var _ RowIterator = &hashDiskRowIterator{}

func (h *HashDiskRowContainer) NewUnmarkedIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569317)
	if h.shouldMark {
		__antithesis_instrumentation__.Notify(569319)
		return &hashDiskRowIterator{
			diskRowIterator: h.NewIterator(ctx).(*diskRowIterator),
		}
	} else {
		__antithesis_instrumentation__.Notify(569320)
	}
	__antithesis_instrumentation__.Notify(569318)
	return h.NewIterator(ctx)
}

func (i *hashDiskRowIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569321)
	i.diskRowIterator.Rewind()

	if i.isRowMarked() {
		__antithesis_instrumentation__.Notify(569322)
		i.Next()
	} else {
		__antithesis_instrumentation__.Notify(569323)
	}
}

func (i *hashDiskRowIterator) Next() {
	__antithesis_instrumentation__.Notify(569324)
	i.diskRowIterator.Next()
	for i.isRowMarked() {
		__antithesis_instrumentation__.Notify(569325)
		i.diskRowIterator.Next()
	}
}

func (i *hashDiskRowIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569326)
	row, err := i.diskRowIterator.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(569328)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569329)
	}
	__antithesis_instrumentation__.Notify(569327)

	row = row[:len(row)-1]
	return row, nil
}

func (i *hashDiskRowIterator) isRowMarked() bool {
	__antithesis_instrumentation__.Notify(569330)

	ok, err := i.diskRowIterator.Valid()
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(569332)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(569333)
		return false
	} else {
		__antithesis_instrumentation__.Notify(569334)
	}
	__antithesis_instrumentation__.Notify(569331)

	rowVal := i.UnsafeValue()
	return bytes.Equal(rowVal[len(rowVal)-len(encodedTrue):], encodedTrue)
}

type HashDiskBackedRowContainer struct {
	src HashRowContainer

	hmrc *HashMemRowContainer
	hdrc *HashDiskRowContainer

	shouldMark   bool
	types        []*types.T
	storedEqCols columns
	encodeNull   bool

	evalCtx       *tree.EvalContext
	memoryMonitor *mon.BytesMonitor
	diskMonitor   *mon.BytesMonitor
	engine        diskmap.Factory
	scratchEncRow rowenc.EncDatumRow

	allRowsIterators []*AllRowsIterator
}

var _ HashRowContainer = &HashDiskBackedRowContainer{}

func NewHashDiskBackedRowContainer(
	evalCtx *tree.EvalContext,
	memoryMonitor *mon.BytesMonitor,
	diskMonitor *mon.BytesMonitor,
	engine diskmap.Factory,
) *HashDiskBackedRowContainer {
	__antithesis_instrumentation__.Notify(569335)
	return &HashDiskBackedRowContainer{
		evalCtx:          evalCtx,
		memoryMonitor:    memoryMonitor,
		diskMonitor:      diskMonitor,
		engine:           engine,
		allRowsIterators: make([]*AllRowsIterator, 0, 1),
	}
}

func (h *HashDiskBackedRowContainer) Init(
	ctx context.Context, shouldMark bool, types []*types.T, storedEqCols columns, encodeNull bool,
) error {
	__antithesis_instrumentation__.Notify(569336)
	h.shouldMark = shouldMark
	h.types = types
	h.storedEqCols = storedEqCols
	h.encodeNull = encodeNull
	if shouldMark {
		__antithesis_instrumentation__.Notify(569339)

		h.scratchEncRow = make(rowenc.EncDatumRow, len(types)+1)
	} else {
		__antithesis_instrumentation__.Notify(569340)
	}
	__antithesis_instrumentation__.Notify(569337)

	hmrc := MakeHashMemRowContainer(h.evalCtx, h.memoryMonitor, types, storedEqCols)
	h.hmrc = &hmrc
	h.src = h.hmrc
	if err := h.hmrc.Init(ctx, shouldMark, types, storedEqCols, encodeNull); err != nil {
		__antithesis_instrumentation__.Notify(569341)
		if spilled, spillErr := h.spillIfMemErr(ctx, err); !spilled && func() bool {
			__antithesis_instrumentation__.Notify(569342)
			return spillErr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(569343)

			return err
		} else {
			__antithesis_instrumentation__.Notify(569344)
			if spillErr != nil {
				__antithesis_instrumentation__.Notify(569345)

				return spillErr
			} else {
				__antithesis_instrumentation__.Notify(569346)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(569347)
	}
	__antithesis_instrumentation__.Notify(569338)

	return nil
}

func (h *HashDiskBackedRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569348)
	if err := h.src.AddRow(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(569350)
		if spilled, spillErr := h.spillIfMemErr(ctx, err); !spilled && func() bool {
			__antithesis_instrumentation__.Notify(569352)
			return spillErr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(569353)

			return err
		} else {
			__antithesis_instrumentation__.Notify(569354)
			if spillErr != nil {
				__antithesis_instrumentation__.Notify(569355)

				return spillErr
			} else {
				__antithesis_instrumentation__.Notify(569356)
			}
		}
		__antithesis_instrumentation__.Notify(569351)

		return h.src.AddRow(ctx, row)
	} else {
		__antithesis_instrumentation__.Notify(569357)
	}
	__antithesis_instrumentation__.Notify(569349)
	return nil
}

func (h *HashDiskBackedRowContainer) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(569358)
	return h.src.IsEmpty()
}

func (h *HashDiskBackedRowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569359)
	if h.hdrc != nil {
		__antithesis_instrumentation__.Notify(569361)
		h.hdrc.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(569362)
	}
	__antithesis_instrumentation__.Notify(569360)
	h.hmrc.Close(ctx)
}

func (h *HashDiskBackedRowContainer) UsingDisk() bool {
	__antithesis_instrumentation__.Notify(569363)
	return h.hdrc != nil
}

func (h *HashDiskBackedRowContainer) ReserveMarkMemoryMaybe(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569364)
	if !h.UsingDisk() {
		__antithesis_instrumentation__.Notify(569366)

		if err := h.hmrc.ReserveMarkMemoryMaybe(ctx); err != nil {
			__antithesis_instrumentation__.Notify(569367)
			return h.SpillToDisk(ctx)
		} else {
			__antithesis_instrumentation__.Notify(569368)
		}
	} else {
		__antithesis_instrumentation__.Notify(569369)
	}
	__antithesis_instrumentation__.Notify(569365)
	return nil
}

func (h *HashDiskBackedRowContainer) spillIfMemErr(ctx context.Context, err error) (bool, error) {
	__antithesis_instrumentation__.Notify(569370)
	if !sqlerrors.IsOutOfMemoryError(err) {
		__antithesis_instrumentation__.Notify(569373)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(569374)
	}
	__antithesis_instrumentation__.Notify(569371)
	if spillErr := h.SpillToDisk(ctx); spillErr != nil {
		__antithesis_instrumentation__.Notify(569375)
		return false, spillErr
	} else {
		__antithesis_instrumentation__.Notify(569376)
	}
	__antithesis_instrumentation__.Notify(569372)
	log.VEventf(ctx, 2, "spilled to disk: %v", err)
	return true, nil
}

func (h *HashDiskBackedRowContainer) SpillToDisk(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569377)
	if h.UsingDisk() {
		__antithesis_instrumentation__.Notify(569382)
		return errors.New("already using disk")
	} else {
		__antithesis_instrumentation__.Notify(569383)
	}
	__antithesis_instrumentation__.Notify(569378)
	hdrc := MakeHashDiskRowContainer(h.diskMonitor, h.engine)
	if err := hdrc.Init(ctx, h.shouldMark, h.types, h.storedEqCols, h.encodeNull); err != nil {
		__antithesis_instrumentation__.Notify(569384)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569385)
	}
	__antithesis_instrumentation__.Notify(569379)

	if err := h.computeKeysForAllRowsIterators(); err != nil {
		__antithesis_instrumentation__.Notify(569386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569387)
	}
	__antithesis_instrumentation__.Notify(569380)

	rowIdx := 0
	i := h.hmrc.NewFinalIterator(ctx)
	defer i.Close()
	for i.Rewind(); ; i.Next() {
		__antithesis_instrumentation__.Notify(569388)
		if ok, err := i.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(569392)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569393)
			if !ok {
				__antithesis_instrumentation__.Notify(569394)
				break
			} else {
				__antithesis_instrumentation__.Notify(569395)
			}
		}
		__antithesis_instrumentation__.Notify(569389)
		row, err := i.Row()
		if err != nil {
			__antithesis_instrumentation__.Notify(569396)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569397)
		}
		__antithesis_instrumentation__.Notify(569390)
		if h.shouldMark && func() bool {
			__antithesis_instrumentation__.Notify(569398)
			return h.hmrc.marked != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(569399)

			copy(h.scratchEncRow, row)
			h.scratchEncRow[len(h.types)] = rowenc.EncDatum{Datum: tree.MakeDBool(tree.DBool(h.hmrc.marked[rowIdx]))}
			row = h.scratchEncRow
			rowIdx++
		} else {
			__antithesis_instrumentation__.Notify(569400)
		}
		__antithesis_instrumentation__.Notify(569391)
		if err := hdrc.AddRow(ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(569401)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569402)
		}
	}
	__antithesis_instrumentation__.Notify(569381)
	h.hmrc.Clear(ctx)

	h.src = &hdrc
	h.hdrc = &hdrc

	return h.recreateAllRowsIterators(ctx)
}

func (h *HashDiskBackedRowContainer) NewBucketIterator(
	ctx context.Context, row rowenc.EncDatumRow, probeEqCols columns,
) (RowMarkerIterator, error) {
	__antithesis_instrumentation__.Notify(569403)
	return h.src.NewBucketIterator(ctx, row, probeEqCols)
}

func (h *HashDiskBackedRowContainer) NewUnmarkedIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569404)
	return h.src.NewUnmarkedIterator(ctx)
}

func (h *HashDiskBackedRowContainer) Sort(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569405)
	if !h.UsingDisk() && func() bool {
		__antithesis_instrumentation__.Notify(569406)
		return len(h.storedEqCols) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(569407)

		h.hmrc.Sort(ctx)
	} else {
		__antithesis_instrumentation__.Notify(569408)
	}
}

type AllRowsIterator struct {
	RowIterator

	container *HashDiskBackedRowContainer
}

func (i *AllRowsIterator) Close() {
	__antithesis_instrumentation__.Notify(569409)
	i.RowIterator.Close()
	for j, iterator := range i.container.allRowsIterators {
		__antithesis_instrumentation__.Notify(569410)
		if i == iterator {
			__antithesis_instrumentation__.Notify(569411)
			i.container.allRowsIterators = append(i.container.allRowsIterators[:j], i.container.allRowsIterators[j+1:]...)
			return
		} else {
			__antithesis_instrumentation__.Notify(569412)
		}
	}
}

func (h *HashDiskBackedRowContainer) NewAllRowsIterator(
	ctx context.Context,
) (*AllRowsIterator, error) {
	__antithesis_instrumentation__.Notify(569413)
	if h.shouldMark {
		__antithesis_instrumentation__.Notify(569415)
		return nil, errors.Errorf("AllRowsIterator can only be created when the container doesn't do marking")
	} else {
		__antithesis_instrumentation__.Notify(569416)
	}
	__antithesis_instrumentation__.Notify(569414)
	i := AllRowsIterator{h.src.NewUnmarkedIterator(ctx), h}
	h.allRowsIterators = append(h.allRowsIterators, &i)
	return &i, nil
}

func (h *HashDiskBackedRowContainer) computeKeysForAllRowsIterators() error {
	__antithesis_instrumentation__.Notify(569417)
	var oldIterator *hashMemRowIterator
	var ok bool
	for _, iterator := range h.allRowsIterators {
		__antithesis_instrumentation__.Notify(569419)
		if oldIterator, ok = (*iterator).RowIterator.(*hashMemRowIterator); !ok {
			__antithesis_instrumentation__.Notify(569421)
			return errors.Errorf("the iterator is unexpectedly not hashMemRowIterator")
		} else {
			__antithesis_instrumentation__.Notify(569422)
		}
		__antithesis_instrumentation__.Notify(569420)
		if err := oldIterator.computeKey(); err != nil {
			__antithesis_instrumentation__.Notify(569423)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569424)
		}
	}
	__antithesis_instrumentation__.Notify(569418)
	return nil
}

func (h *HashDiskBackedRowContainer) recreateAllRowsIterators(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569425)
	var oldIterator *hashMemRowIterator
	var ok bool
	for _, iterator := range h.allRowsIterators {
		__antithesis_instrumentation__.Notify(569427)
		if oldIterator, ok = (*iterator).RowIterator.(*hashMemRowIterator); !ok {
			__antithesis_instrumentation__.Notify(569429)
			return errors.Errorf("the iterator is unexpectedly not hashMemRowIterator")
		} else {
			__antithesis_instrumentation__.Notify(569430)
		}
		__antithesis_instrumentation__.Notify(569428)
		newIterator := h.NewUnmarkedIterator(ctx)
		newIterator.(*diskRowIterator).SeekGE(oldIterator.curKey)
		(*iterator).RowIterator.Close()
		iterator.RowIterator = newIterator
	}
	__antithesis_instrumentation__.Notify(569426)
	return nil
}
