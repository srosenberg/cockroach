package rowcontainer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type DiskRowContainer struct {
	diskMap diskmap.SortedDiskMap

	diskAcc mon.BoundAccount

	bufferedRows  diskmap.SortedDiskMapBatchWriter
	scratchKey    []byte
	scratchVal    []byte
	scratchEncRow rowenc.EncDatumRow

	totalEncodedRowBytes uint64

	lastReadKey []byte

	topK int

	rowID uint64

	types []*types.T

	ordering colinfo.ColumnOrdering

	encodings []descpb.DatumEncoding

	valueIdxs []int

	deDuplicate bool

	deDupCache map[string]int

	diskMonitor *mon.BytesMonitor
	engine      diskmap.Factory

	datumAlloc *tree.DatumAlloc
}

var _ SortableRowContainer = &DiskRowContainer{}
var _ DeDupingRowContainer = &DiskRowContainer{}

func MakeDiskRowContainer(
	diskMonitor *mon.BytesMonitor,
	types []*types.T,
	ordering colinfo.ColumnOrdering,
	e diskmap.Factory,
) DiskRowContainer {
	__antithesis_instrumentation__.Notify(569014)
	diskMap := e.NewSortedDiskMap()
	d := DiskRowContainer{
		diskMap:       diskMap,
		diskAcc:       diskMonitor.MakeBoundAccount(),
		types:         types,
		ordering:      ordering,
		scratchEncRow: make(rowenc.EncDatumRow, len(types)),
		diskMonitor:   diskMonitor,
		engine:        e,
		datumAlloc:    &tree.DatumAlloc{},
	}
	d.bufferedRows = d.diskMap.NewBatchWriter()

	orderingIdxs := make(map[int]struct{})
	for _, orderInfo := range d.ordering {
		__antithesis_instrumentation__.Notify(569018)
		orderingIdxs[orderInfo.ColIdx] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(569015)
	d.valueIdxs = make([]int, 0, len(d.types))
	for i := range d.types {
		__antithesis_instrumentation__.Notify(569019)

		if _, ok := orderingIdxs[i]; !ok || func() bool {
			__antithesis_instrumentation__.Notify(569020)
			return colinfo.CanHaveCompositeKeyEncoding(d.types[i]) == true
		}() == true {
			__antithesis_instrumentation__.Notify(569021)
			d.valueIdxs = append(d.valueIdxs, i)
		} else {
			__antithesis_instrumentation__.Notify(569022)
		}
	}
	__antithesis_instrumentation__.Notify(569016)

	d.encodings = make([]descpb.DatumEncoding, len(d.ordering))
	for i, orderInfo := range ordering {
		__antithesis_instrumentation__.Notify(569023)
		d.encodings[i] = rowenc.EncodingDirToDatumEncoding(orderInfo.Direction)
	}
	__antithesis_instrumentation__.Notify(569017)

	return d
}

func (d *DiskRowContainer) DoDeDuplicate() {
	__antithesis_instrumentation__.Notify(569024)
	d.deDuplicate = true
	d.deDupCache = make(map[string]int)
}

func (d *DiskRowContainer) Len() int {
	__antithesis_instrumentation__.Notify(569025)
	return int(d.rowID)
}

func (d *DiskRowContainer) AddRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569026)
	if err := d.encodeRow(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(569031)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569032)
	}
	__antithesis_instrumentation__.Notify(569027)
	if err := d.diskAcc.Grow(ctx, int64(len(d.scratchKey)+len(d.scratchVal))); err != nil {
		__antithesis_instrumentation__.Notify(569033)
		return pgerror.Wrapf(err, pgcode.OutOfMemory,
			"this query requires additional disk space")
	} else {
		__antithesis_instrumentation__.Notify(569034)
	}
	__antithesis_instrumentation__.Notify(569028)
	if err := d.bufferedRows.Put(d.scratchKey, d.scratchVal); err != nil {
		__antithesis_instrumentation__.Notify(569035)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569036)
	}
	__antithesis_instrumentation__.Notify(569029)

	if d.deDuplicate {
		__antithesis_instrumentation__.Notify(569037)
		if d.bufferedRows.NumPutsSinceFlush() == 0 {
			__antithesis_instrumentation__.Notify(569038)
			d.clearDeDupCache()
		} else {
			__antithesis_instrumentation__.Notify(569039)
			d.deDupCache[string(d.scratchKey)] = int(d.rowID)
		}
	} else {
		__antithesis_instrumentation__.Notify(569040)
	}
	__antithesis_instrumentation__.Notify(569030)
	d.totalEncodedRowBytes += uint64(len(d.scratchKey) + len(d.scratchVal))
	d.scratchKey = d.scratchKey[:0]
	d.scratchVal = d.scratchVal[:0]
	d.rowID++
	return nil
}

func (d *DiskRowContainer) AddRowWithDeDup(
	ctx context.Context, row rowenc.EncDatumRow,
) (int, error) {
	__antithesis_instrumentation__.Notify(569041)
	if err := d.encodeRow(ctx, row); err != nil {
		__antithesis_instrumentation__.Notify(569050)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(569051)
	}
	__antithesis_instrumentation__.Notify(569042)
	defer func() {
		__antithesis_instrumentation__.Notify(569052)
		d.scratchKey = d.scratchKey[:0]
		d.scratchVal = d.scratchVal[:0]
	}()
	__antithesis_instrumentation__.Notify(569043)

	entry, ok := d.deDupCache[string(d.scratchKey)]
	if ok {
		__antithesis_instrumentation__.Notify(569053)
		return entry, nil
	} else {
		__antithesis_instrumentation__.Notify(569054)
	}
	__antithesis_instrumentation__.Notify(569044)

	iter := d.diskMap.NewIterator()
	defer iter.Close()
	iter.SeekGE(d.scratchKey)
	valid, err := iter.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(569055)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(569056)
	}
	__antithesis_instrumentation__.Notify(569045)
	if valid && func() bool {
		__antithesis_instrumentation__.Notify(569057)
		return bytes.Equal(iter.UnsafeKey(), d.scratchKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(569058)

		_, idx, err := encoding.DecodeUvarintAscending(iter.UnsafeValue())
		if err != nil {
			__antithesis_instrumentation__.Notify(569060)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(569061)
		}
		__antithesis_instrumentation__.Notify(569059)
		return int(idx), nil
	} else {
		__antithesis_instrumentation__.Notify(569062)
	}
	__antithesis_instrumentation__.Notify(569046)
	if err := d.diskAcc.Grow(ctx, int64(len(d.scratchKey)+len(d.scratchVal))); err != nil {
		__antithesis_instrumentation__.Notify(569063)
		return 0, pgerror.Wrapf(err, pgcode.OutOfMemory,
			"this query requires additional disk space")
	} else {
		__antithesis_instrumentation__.Notify(569064)
	}
	__antithesis_instrumentation__.Notify(569047)
	if err := d.bufferedRows.Put(d.scratchKey, d.scratchVal); err != nil {
		__antithesis_instrumentation__.Notify(569065)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(569066)
	}
	__antithesis_instrumentation__.Notify(569048)
	if d.bufferedRows.NumPutsSinceFlush() == 0 {
		__antithesis_instrumentation__.Notify(569067)
		d.clearDeDupCache()
	} else {
		__antithesis_instrumentation__.Notify(569068)
		d.deDupCache[string(d.scratchKey)] = int(d.rowID)
	}
	__antithesis_instrumentation__.Notify(569049)
	d.totalEncodedRowBytes += uint64(len(d.scratchKey) + len(d.scratchVal))
	idx := int(d.rowID)
	d.rowID++
	return idx, nil
}

func (d *DiskRowContainer) clearDeDupCache() {
	__antithesis_instrumentation__.Notify(569069)
	for k := range d.deDupCache {
		__antithesis_instrumentation__.Notify(569070)
		delete(d.deDupCache, k)
	}
}

func (d *DiskRowContainer) testingFlushBuffer(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569071)
	if err := d.bufferedRows.Flush(); err != nil {
		__antithesis_instrumentation__.Notify(569073)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(569074)
	}
	__antithesis_instrumentation__.Notify(569072)
	d.clearDeDupCache()
}

func (d *DiskRowContainer) encodeRow(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569075)
	if len(row) != len(d.types) {
		__antithesis_instrumentation__.Notify(569079)
		log.Fatalf(ctx, "invalid row length %d, expected %d", len(row), len(d.types))
	} else {
		__antithesis_instrumentation__.Notify(569080)
	}
	__antithesis_instrumentation__.Notify(569076)

	for i, orderInfo := range d.ordering {
		__antithesis_instrumentation__.Notify(569081)
		col := orderInfo.ColIdx
		var err error
		d.scratchKey, err = row[col].Encode(d.types[col], d.datumAlloc, d.encodings[i], d.scratchKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(569082)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569083)
		}
	}
	__antithesis_instrumentation__.Notify(569077)
	if !d.deDuplicate {
		__antithesis_instrumentation__.Notify(569084)
		for _, i := range d.valueIdxs {
			__antithesis_instrumentation__.Notify(569086)
			var err error
			d.scratchVal, err = row[i].Encode(d.types[i], d.datumAlloc, descpb.DatumEncoding_VALUE, d.scratchVal)
			if err != nil {
				__antithesis_instrumentation__.Notify(569087)
				return err
			} else {
				__antithesis_instrumentation__.Notify(569088)
			}
		}
		__antithesis_instrumentation__.Notify(569085)

		d.scratchKey = encoding.EncodeUvarintAscending(d.scratchKey, d.rowID)
	} else {
		__antithesis_instrumentation__.Notify(569089)

		d.scratchVal = encoding.EncodeUvarintAscending(d.scratchVal, d.rowID)
	}
	__antithesis_instrumentation__.Notify(569078)
	return nil
}

func (d *DiskRowContainer) Sort(context.Context) { __antithesis_instrumentation__.Notify(569090) }

func (d *DiskRowContainer) Reorder(ctx context.Context, ordering colinfo.ColumnOrdering) error {
	__antithesis_instrumentation__.Notify(569091)

	newContainer := MakeDiskRowContainer(d.diskMonitor, d.types, ordering, d.engine)
	i := d.NewFinalIterator(ctx)
	defer i.Close()
	for i.Rewind(); ; i.Next() {
		__antithesis_instrumentation__.Notify(569093)
		if ok, err := i.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(569096)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569097)
			if !ok {
				__antithesis_instrumentation__.Notify(569098)
				break
			} else {
				__antithesis_instrumentation__.Notify(569099)
			}
		}
		__antithesis_instrumentation__.Notify(569094)
		row, err := i.Row()
		if err != nil {
			__antithesis_instrumentation__.Notify(569100)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569101)
		}
		__antithesis_instrumentation__.Notify(569095)
		if err := newContainer.AddRow(ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(569102)
			return err
		} else {
			__antithesis_instrumentation__.Notify(569103)
		}
	}
	__antithesis_instrumentation__.Notify(569092)
	d.Close(ctx)
	*d = newContainer
	return nil
}

func (d *DiskRowContainer) InitTopK() {
	__antithesis_instrumentation__.Notify(569104)
	d.topK = d.Len()
}

func (d *DiskRowContainer) MaybeReplaceMax(ctx context.Context, row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(569105)
	return d.AddRow(ctx, row)
}

func (d *DiskRowContainer) MeanEncodedRowBytes() int {
	__antithesis_instrumentation__.Notify(569106)
	if d.rowID == 0 {
		__antithesis_instrumentation__.Notify(569108)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(569109)
	}
	__antithesis_instrumentation__.Notify(569107)
	return int(d.totalEncodedRowBytes / d.rowID)
}

func (d *DiskRowContainer) UnsafeReset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(569110)
	_ = d.bufferedRows.Close(ctx)
	if err := d.diskMap.Clear(); err != nil {
		__antithesis_instrumentation__.Notify(569112)
		return err
	} else {
		__antithesis_instrumentation__.Notify(569113)
	}
	__antithesis_instrumentation__.Notify(569111)
	d.diskAcc.Clear(ctx)
	d.bufferedRows = d.diskMap.NewBatchWriter()
	d.clearDeDupCache()
	d.lastReadKey = nil
	d.rowID = 0
	d.totalEncodedRowBytes = 0
	return nil
}

func (d *DiskRowContainer) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(569114)

	_ = d.bufferedRows.Close(ctx)
	d.diskMap.Close(ctx)
	d.diskAcc.Close(ctx)
}

func (d *DiskRowContainer) keyValToRow(k []byte, v []byte) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569115)
	for i, orderInfo := range d.ordering {
		__antithesis_instrumentation__.Notify(569118)

		if colinfo.CanHaveCompositeKeyEncoding(d.types[orderInfo.ColIdx]) {
			__antithesis_instrumentation__.Notify(569120)

			encLen, err := encoding.PeekLength(k)
			if err != nil {
				__antithesis_instrumentation__.Notify(569122)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(569123)
			}
			__antithesis_instrumentation__.Notify(569121)
			k = k[encLen:]
			continue
		} else {
			__antithesis_instrumentation__.Notify(569124)
		}
		__antithesis_instrumentation__.Notify(569119)
		var err error
		col := orderInfo.ColIdx
		d.scratchEncRow[col], k, err = rowenc.EncDatumFromBuffer(d.types[col], d.encodings[i], k)
		if err != nil {
			__antithesis_instrumentation__.Notify(569125)
			return nil, errors.NewAssertionErrorWithWrappedErrf(err,
				"unable to decode row, column idx %d", errors.Safe(col))
		} else {
			__antithesis_instrumentation__.Notify(569126)
		}
	}
	__antithesis_instrumentation__.Notify(569116)
	for _, i := range d.valueIdxs {
		__antithesis_instrumentation__.Notify(569127)
		var err error
		d.scratchEncRow[i], v, err = rowenc.EncDatumFromBuffer(d.types[i], descpb.DatumEncoding_VALUE, v)
		if err != nil {
			__antithesis_instrumentation__.Notify(569128)
			return nil, errors.NewAssertionErrorWithWrappedErrf(err,
				"unable to decode row, value idx %d", errors.Safe(i))
		} else {
			__antithesis_instrumentation__.Notify(569129)
		}
	}
	__antithesis_instrumentation__.Notify(569117)
	return d.scratchEncRow, nil
}

type diskRowIterator struct {
	rowContainer *DiskRowContainer
	rowBuf       []byte
	diskmap.SortedDiskMapIterator
}

var _ RowIterator = &diskRowIterator{}

func (d *DiskRowContainer) newIterator(ctx context.Context) diskRowIterator {
	__antithesis_instrumentation__.Notify(569130)
	if err := d.bufferedRows.Flush(); err != nil {
		__antithesis_instrumentation__.Notify(569132)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(569133)
	}
	__antithesis_instrumentation__.Notify(569131)
	return diskRowIterator{rowContainer: d, SortedDiskMapIterator: d.diskMap.NewIterator()}
}

func (d *DiskRowContainer) NewIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569134)
	i := d.newIterator(ctx)
	if d.topK > 0 {
		__antithesis_instrumentation__.Notify(569136)
		return &diskRowTopKIterator{RowIterator: &i, k: d.topK}
	} else {
		__antithesis_instrumentation__.Notify(569137)
	}
	__antithesis_instrumentation__.Notify(569135)
	return &i
}

func (r *diskRowIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569138)
	if ok, err := r.Valid(); err != nil {
		__antithesis_instrumentation__.Notify(569141)
		return nil, errors.NewAssertionErrorWithWrappedErrf(err, "unable to check row validity")
	} else {
		__antithesis_instrumentation__.Notify(569142)
		if !ok {
			__antithesis_instrumentation__.Notify(569143)
			return nil, errors.AssertionFailedf("invalid row")
		} else {
			__antithesis_instrumentation__.Notify(569144)
		}
	}
	__antithesis_instrumentation__.Notify(569139)

	k := r.UnsafeKey()
	v := r.UnsafeValue()

	if true || func() bool {
		__antithesis_instrumentation__.Notify(569145)
		return cap(r.rowBuf) < len(k)+len(v) == true
	}() == true {
		__antithesis_instrumentation__.Notify(569146)
		r.rowBuf = make([]byte, 0, len(k)+len(v))
	} else {
		__antithesis_instrumentation__.Notify(569147)
	}
	__antithesis_instrumentation__.Notify(569140)
	r.rowBuf = r.rowBuf[:len(k)+len(v)]
	copy(r.rowBuf, k)
	copy(r.rowBuf[len(k):], v)
	k = r.rowBuf[:len(k)]
	v = r.rowBuf[len(k):]

	return r.rowContainer.keyValToRow(k, v)
}

func (r *diskRowIterator) Close() {
	__antithesis_instrumentation__.Notify(569148)
	if r.SortedDiskMapIterator != nil {
		__antithesis_instrumentation__.Notify(569149)
		r.SortedDiskMapIterator.Close()
	} else {
		__antithesis_instrumentation__.Notify(569150)
	}
}

type numberedRowIterator struct {
	*diskRowIterator
	scratchKey []byte
}

func (d *DiskRowContainer) newNumberedIterator(ctx context.Context) *numberedRowIterator {
	__antithesis_instrumentation__.Notify(569151)
	i := d.newIterator(ctx)
	return &numberedRowIterator{diskRowIterator: &i}
}

func (n numberedRowIterator) seekToIndex(idx int) {
	__antithesis_instrumentation__.Notify(569152)
	n.scratchKey = encoding.EncodeUvarintAscending(n.scratchKey, uint64(idx))
	n.SeekGE(n.scratchKey)
}

type diskRowFinalIterator struct {
	diskRowIterator
}

var _ RowIterator = &diskRowFinalIterator{}

func (d *DiskRowContainer) NewFinalIterator(ctx context.Context) RowIterator {
	__antithesis_instrumentation__.Notify(569153)
	i := diskRowFinalIterator{diskRowIterator: d.newIterator(ctx)}
	if d.topK > 0 {
		__antithesis_instrumentation__.Notify(569155)
		return &diskRowTopKIterator{RowIterator: &i, k: d.topK}
	} else {
		__antithesis_instrumentation__.Notify(569156)
	}
	__antithesis_instrumentation__.Notify(569154)
	return &i
}

func (r *diskRowFinalIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569157)
	r.SeekGE(r.diskRowIterator.rowContainer.lastReadKey)
	if r.diskRowIterator.rowContainer.lastReadKey != nil {
		__antithesis_instrumentation__.Notify(569158)
		r.Next()
	} else {
		__antithesis_instrumentation__.Notify(569159)
	}
}

func (r *diskRowFinalIterator) Row() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(569160)
	row, err := r.diskRowIterator.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(569162)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(569163)
	}
	__antithesis_instrumentation__.Notify(569161)
	r.diskRowIterator.rowContainer.lastReadKey =
		append(r.diskRowIterator.rowContainer.lastReadKey[:0], r.UnsafeKey()...)
	return row, nil
}

type diskRowTopKIterator struct {
	RowIterator
	position int

	k int
}

var _ RowIterator = &diskRowTopKIterator{}

func (d *diskRowTopKIterator) Rewind() {
	__antithesis_instrumentation__.Notify(569164)
	d.RowIterator.Rewind()
	d.position = 0
}

func (d *diskRowTopKIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(569165)
	if d.position >= d.k {
		__antithesis_instrumentation__.Notify(569167)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(569168)
	}
	__antithesis_instrumentation__.Notify(569166)
	return d.RowIterator.Valid()
}

func (d *diskRowTopKIterator) Next() {
	__antithesis_instrumentation__.Notify(569169)
	d.position++
	d.RowIterator.Next()
}
