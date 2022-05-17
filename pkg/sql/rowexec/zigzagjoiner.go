package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/scrub"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type zigzagJoiner struct {
	joinerBase

	cancelChecker cancelchecker.CancelChecker

	numTables int

	side int

	infos []*zigzagJoinerInfo

	baseRow rowenc.EncDatumRow

	rowAlloc           rowenc.EncDatumRowAlloc
	fetchedInititalRow bool

	scanStats execinfra.ScanStats
}

var zigzagJoinerBatchSize = rowinfra.RowLimit(util.ConstantWithMetamorphicTestValue(
	"zig-zag-joiner-batch-size",
	5,
	1,
))

var _ execinfra.Processor = &zigzagJoiner{}
var _ execinfra.RowSource = &zigzagJoiner{}
var _ execinfra.OpNode = &zigzagJoiner{}

const zigzagJoinerProcName = "zigzagJoiner"

func newZigzagJoiner(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.ZigzagJoinerSpec,
	fixedValues []rowenc.EncDatumRow,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*zigzagJoiner, error) {
	__antithesis_instrumentation__.Notify(575556)
	tables := make([]catalog.TableDescriptor, len(spec.Tables))
	for i := range spec.Tables {
		__antithesis_instrumentation__.Notify(575565)
		tables[i] = flowCtx.TableDescriptor(&spec.Tables[i])
	}
	__antithesis_instrumentation__.Notify(575557)
	if len(tables) != 2 {
		__antithesis_instrumentation__.Notify(575566)
		return nil, errors.AssertionFailedf("zigzag joins only of two tables (or indexes) are supported, %d requested", len(tables))
	} else {
		__antithesis_instrumentation__.Notify(575567)
	}
	__antithesis_instrumentation__.Notify(575558)
	if spec.Type != descpb.InnerJoin {
		__antithesis_instrumentation__.Notify(575568)
		return nil, errors.AssertionFailedf("only inner zigzag joins are supported, %s requested", spec.Type)
	} else {
		__antithesis_instrumentation__.Notify(575569)
	}
	__antithesis_instrumentation__.Notify(575559)
	z := &zigzagJoiner{}

	leftColumnTypes := catalog.ColumnTypes(tables[0].PublicColumns())
	rightColumnTypes := catalog.ColumnTypes(tables[1].PublicColumns())
	err := z.joinerBase.init(
		z,
		flowCtx,
		processorID,
		leftColumnTypes,
		rightColumnTypes,
		spec.Type,
		spec.OnExpr,
		false,
		post,
		output,
		execinfra.ProcStateOpts{
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(575570)

				trailingMeta := z.generateMeta()
				z.close()
				return trailingMeta
			},
		},
	)
	__antithesis_instrumentation__.Notify(575560)
	if err != nil {
		__antithesis_instrumentation__.Notify(575571)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(575572)
	}
	__antithesis_instrumentation__.Notify(575561)

	z.numTables = len(tables)
	z.infos = make([]*zigzagJoinerInfo, z.numTables)
	for i := range z.infos {
		__antithesis_instrumentation__.Notify(575573)
		z.infos[i] = &zigzagJoinerInfo{}
	}
	__antithesis_instrumentation__.Notify(575562)

	collectingStats := false
	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(575574)
		collectingStats = true
		z.ExecStatsForTrace = z.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(575575)
	}
	__antithesis_instrumentation__.Notify(575563)

	colOffset := 0
	for i := 0; i < z.numTables; i++ {
		__antithesis_instrumentation__.Notify(575576)
		if fixedValues != nil && func() bool {
			__antithesis_instrumentation__.Notify(575579)
			return i < len(fixedValues) == true
		}() == true {
			__antithesis_instrumentation__.Notify(575580)

			z.infos[i].fixedValues = fixedValues[i]
		} else {
			__antithesis_instrumentation__.Notify(575581)
			if i < len(spec.FixedValues) {
				__antithesis_instrumentation__.Notify(575582)
				z.infos[i].fixedValues, err = valuesSpecToEncDatum(spec.FixedValues[i])
				if err != nil {
					__antithesis_instrumentation__.Notify(575583)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(575584)
				}
			} else {
				__antithesis_instrumentation__.Notify(575585)
			}
		}
		__antithesis_instrumentation__.Notify(575577)
		if err := z.setupInfo(flowCtx, spec, i, colOffset, tables, collectingStats); err != nil {
			__antithesis_instrumentation__.Notify(575586)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575587)
		}
		__antithesis_instrumentation__.Notify(575578)
		colOffset += len(z.infos[i].table.PublicColumns())
	}
	__antithesis_instrumentation__.Notify(575564)
	z.side = 0
	return z, nil
}

func valuesSpecToEncDatum(
	valuesSpec *execinfrapb.ValuesCoreSpec,
) (res []rowenc.EncDatum, err error) {
	__antithesis_instrumentation__.Notify(575588)
	res = make([]rowenc.EncDatum, len(valuesSpec.Columns))
	rem := valuesSpec.RawBytes[0]
	for i, colInfo := range valuesSpec.Columns {
		__antithesis_instrumentation__.Notify(575590)
		res[i], rem, err = rowenc.EncDatumFromBuffer(colInfo.Type, colInfo.Encoding, rem)
		if err != nil {
			__antithesis_instrumentation__.Notify(575591)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575592)
		}
	}
	__antithesis_instrumentation__.Notify(575589)
	return res, nil
}

func (z *zigzagJoiner) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575593)
	ctx = z.StartInternal(ctx, zigzagJoinerProcName)
	z.cancelChecker.Reset(ctx)
	log.VEventf(ctx, 2, "starting zigzag joiner run")
}

type zigzagJoinerInfo struct {
	fetcher      rowFetcher
	row          rowenc.EncDatumRow
	rowColIdxMap catalog.TableColMap

	rowsRead   int64
	alloc      *tree.DatumAlloc
	table      catalog.TableDescriptor
	index      catalog.Index
	indexTypes []*types.T
	indexDirs  []descpb.IndexDescriptor_Direction

	container rowenc.EncDatumRowContainer

	eqColumns []uint32

	fixedValues rowenc.EncDatumRow

	key roachpb.Key

	prefix []byte

	endKey roachpb.Key

	spanBuilder span.Builder
}

func (z *zigzagJoiner) setupInfo(
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.ZigzagJoinerSpec,
	side int,
	colOffset int,
	tables []catalog.TableDescriptor,
	collectingStats bool,
) error {
	__antithesis_instrumentation__.Notify(575594)
	z.side = side
	info := z.infos[side]

	info.alloc = &tree.DatumAlloc{}
	info.table = tables[side]
	info.eqColumns = spec.EqColumns[side].Columns
	indexOrdinal := spec.IndexOrdinals[side]
	info.index = info.table.ActiveIndexes()[indexOrdinal]

	info.indexDirs = info.table.IndexFullColumnDirections(info.index)
	columns := info.table.IndexFullColumns(info.index)
	info.indexTypes = make([]*types.T, len(columns))
	columnTypes := catalog.ColumnTypes(info.table.PublicColumns())
	colIdxMap := catalog.ColumnIDToOrdinalMap(info.table.PublicColumns())
	for i, col := range columns {
		__antithesis_instrumentation__.Notify(575602)
		if col == nil {
			__antithesis_instrumentation__.Notify(575604)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575605)
		}
		__antithesis_instrumentation__.Notify(575603)
		columnID := col.GetID()
		if info.index.GetType() == descpb.IndexDescriptor_INVERTED && func() bool {
			__antithesis_instrumentation__.Notify(575606)
			return columnID == info.index.InvertedColumnID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(575607)

			info.indexTypes[i] = types.Bytes
		} else {
			__antithesis_instrumentation__.Notify(575608)
			info.indexTypes[i] = columnTypes[colIdxMap.GetDefault(columnID)]
		}
	}
	__antithesis_instrumentation__.Notify(575595)

	neededCols := util.MakeFastIntSet()
	outCols := z.OutputHelper.NeededColumns()
	maxCol := colOffset + len(info.table.PublicColumns())
	for i, ok := outCols.Next(colOffset); ok && func() bool {
		__antithesis_instrumentation__.Notify(575609)
		return i < maxCol == true
	}() == true; i, ok = outCols.Next(i + 1) {
		__antithesis_instrumentation__.Notify(575610)
		neededCols.Add(i - colOffset)
	}
	__antithesis_instrumentation__.Notify(575596)

	for i := 0; i < len(info.fixedValues); i++ {
		__antithesis_instrumentation__.Notify(575611)
		neededCols.Add(colIdxMap.GetDefault(columns[i].GetID()))
	}
	__antithesis_instrumentation__.Notify(575597)

	for _, col := range info.eqColumns {
		__antithesis_instrumentation__.Notify(575612)
		neededCols.Add(int(col))
	}
	__antithesis_instrumentation__.Notify(575598)

	z.addColumnsNeededByOnExpr(&neededCols, colOffset, maxCol)

	info.container.Reset()

	info.spanBuilder.Init(flowCtx.EvalCtx, flowCtx.Codec(), info.table, info.index)

	fetcher, err := makeRowFetcherLegacy(
		flowCtx,
		info.table,
		int(indexOrdinal),
		false,
		neededCols,
		flowCtx.EvalCtx.Mon,
		info.alloc,

		descpb.ScanLockingStrength_FOR_NONE,
		descpb.ScanLockingWaitPolicy_BLOCK,
		false,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(575613)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575614)
	}
	__antithesis_instrumentation__.Notify(575599)
	info.row = make(rowenc.EncDatumRow, len(info.table.PublicColumns()))
	info.rowColIdxMap = catalog.ColumnIDToOrdinalMap(info.table.PublicColumns())

	if collectingStats {
		__antithesis_instrumentation__.Notify(575615)
		info.fetcher = newRowFetcherStatCollector(fetcher)
	} else {
		__antithesis_instrumentation__.Notify(575616)
		info.fetcher = fetcher
	}
	__antithesis_instrumentation__.Notify(575600)

	info.prefix = rowenc.MakeIndexKeyPrefix(flowCtx.Codec(), info.table.GetID(), info.index.GetID())
	span, err := z.produceSpanFromBaseRow()

	if err != nil {
		__antithesis_instrumentation__.Notify(575617)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575618)
	}
	__antithesis_instrumentation__.Notify(575601)
	info.key = span.Key
	info.endKey = span.EndKey
	return nil
}

func (z *zigzagJoiner) close() {
	__antithesis_instrumentation__.Notify(575619)
	if z.InternalClose() {
		__antithesis_instrumentation__.Notify(575620)
		for i := range z.infos {
			__antithesis_instrumentation__.Notify(575622)
			z.infos[i].fetcher.Close(z.Ctx)
		}
		__antithesis_instrumentation__.Notify(575621)
		log.VEventf(z.Ctx, 2, "exiting zigzag joiner run")
	} else {
		__antithesis_instrumentation__.Notify(575623)
	}
}

func findColumnOrdinalInIndex(index catalog.Index, t descpb.ColumnID) int {
	__antithesis_instrumentation__.Notify(575624)
	for i := 0; i < index.NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(575626)
		if index.GetKeyColumnID(i) == t {
			__antithesis_instrumentation__.Notify(575627)
			return i
		} else {
			__antithesis_instrumentation__.Notify(575628)
		}
	}
	__antithesis_instrumentation__.Notify(575625)
	return -1
}

func (z *zigzagJoiner) fetchRow(ctx context.Context) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(575629)
	return z.fetchRowFromSide(ctx, z.side)
}

func (z *zigzagJoiner) fetchRowFromSide(
	ctx context.Context, side int,
) (fetchedRow rowenc.EncDatumRow, err error) {
	__antithesis_instrumentation__.Notify(575630)
	info := z.infos[side]

	hasNull := func(row rowenc.EncDatumRow) bool {
		__antithesis_instrumentation__.Notify(575633)
		for _, c := range info.eqColumns {
			__antithesis_instrumentation__.Notify(575635)
			if row[c].IsNull() {
				__antithesis_instrumentation__.Notify(575636)
				return true
			} else {
				__antithesis_instrumentation__.Notify(575637)
			}
		}
		__antithesis_instrumentation__.Notify(575634)
		return false
	}
	__antithesis_instrumentation__.Notify(575631)
	fetchedRow = info.row
	for {
		__antithesis_instrumentation__.Notify(575638)
		ok, err := info.fetcher.NextRowInto(ctx, fetchedRow, info.rowColIdxMap)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(575640)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(575641)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575642)
		}
		__antithesis_instrumentation__.Notify(575639)
		info.rowsRead++
		if !hasNull(fetchedRow) {
			__antithesis_instrumentation__.Notify(575643)
			break
		} else {
			__antithesis_instrumentation__.Notify(575644)
		}
	}
	__antithesis_instrumentation__.Notify(575632)
	return fetchedRow, nil
}

func (z *zigzagJoiner) extractEqDatums(row rowenc.EncDatumRow, side int) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(575645)
	eqCols := z.infos[side].eqColumns
	eqDatums := make(rowenc.EncDatumRow, len(eqCols))
	for i, col := range eqCols {
		__antithesis_instrumentation__.Notify(575647)
		eqDatums[i] = row[col]
	}
	__antithesis_instrumentation__.Notify(575646)
	return eqDatums
}

func (z *zigzagJoiner) produceInvertedIndexKey(
	info *zigzagJoinerInfo, datums rowenc.EncDatumRow,
) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(575648)

	var colMap catalog.TableColMap
	decodedDatums := make([]tree.Datum, len(datums))

	for i, encDatum := range datums {
		__antithesis_instrumentation__.Notify(575653)
		err := encDatum.EnsureDecoded(info.indexTypes[i], info.alloc)
		if err != nil {
			__antithesis_instrumentation__.Notify(575655)
			return roachpb.Span{}, err
		} else {
			__antithesis_instrumentation__.Notify(575656)
		}
		__antithesis_instrumentation__.Notify(575654)

		decodedDatums[i] = encDatum.Datum
		if i < info.index.NumKeyColumns() {
			__antithesis_instrumentation__.Notify(575657)
			colMap.Set(info.index.GetKeyColumnID(i), i)
		} else {
			__antithesis_instrumentation__.Notify(575658)

			colMap.Set(info.index.GetKeySuffixColumnID(i-info.index.NumKeyColumns()), i)
		}
	}
	__antithesis_instrumentation__.Notify(575649)

	keyPrefix, err := rowenc.EncodeInvertedIndexPrefixKeys(
		info.index,
		colMap,
		decodedDatums,
		info.prefix,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(575659)
		return roachpb.Span{}, err
	} else {
		__antithesis_instrumentation__.Notify(575660)
	}
	__antithesis_instrumentation__.Notify(575650)

	invOrd, ok := colMap.Get(info.index.InvertedColumnID())
	if !ok {
		__antithesis_instrumentation__.Notify(575661)
		return roachpb.Span{}, errors.AssertionFailedf("inverted column not found in colMap")
	} else {
		__antithesis_instrumentation__.Notify(575662)
	}
	__antithesis_instrumentation__.Notify(575651)
	invertedKey, ok := decodedDatums[invOrd].(*tree.DBytes)
	if !ok {
		__antithesis_instrumentation__.Notify(575663)
		return roachpb.Span{}, errors.AssertionFailedf("inverted key must be type DBytes")
	} else {
		__antithesis_instrumentation__.Notify(575664)
	}
	__antithesis_instrumentation__.Notify(575652)
	keyPrefix = append(keyPrefix, []byte(*invertedKey)...)

	keyBytes, _, err := rowenc.EncodeColumns(
		info.index.IndexDesc().KeySuffixColumnIDs[:len(datums)-1],
		info.indexDirs[1:],
		colMap,
		decodedDatums,
		keyPrefix,
	)
	key := roachpb.Key(keyBytes)
	return roachpb.Span{Key: key, EndKey: key.PrefixEnd()}, err
}

func (z *zigzagJoiner) produceSpanFromBaseRow() (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(575665)
	info := z.infos[z.side]
	neededDatums := info.fixedValues
	if z.baseRow != nil {
		__antithesis_instrumentation__.Notify(575668)
		eqDatums := z.extractEqDatums(z.baseRow, z.prevSide())
		neededDatums = append(neededDatums, eqDatums...)
	} else {
		__antithesis_instrumentation__.Notify(575669)
	}
	__antithesis_instrumentation__.Notify(575666)

	if info.index.GetType() == descpb.IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(575670)
		return z.produceInvertedIndexKey(info, neededDatums)
	} else {
		__antithesis_instrumentation__.Notify(575671)
	}
	__antithesis_instrumentation__.Notify(575667)

	s, _, err := info.spanBuilder.SpanFromEncDatums(neededDatums)
	return s, err
}

func (zi *zigzagJoinerInfo) eqColTypes() []*types.T {
	__antithesis_instrumentation__.Notify(575672)
	eqColTypes := make([]*types.T, len(zi.eqColumns))
	colTypes := catalog.ColumnTypes(zi.table.PublicColumns())
	for i := range eqColTypes {
		__antithesis_instrumentation__.Notify(575674)
		eqColTypes[i] = colTypes[zi.eqColumns[i]]
	}
	__antithesis_instrumentation__.Notify(575673)
	return eqColTypes
}

func (zi *zigzagJoinerInfo) eqOrdering() (colinfo.ColumnOrdering, error) {
	__antithesis_instrumentation__.Notify(575675)
	ordering := make(colinfo.ColumnOrdering, len(zi.eqColumns))
	for i := range zi.eqColumns {
		__antithesis_instrumentation__.Notify(575677)
		colID := zi.table.PublicColumns()[zi.eqColumns[i]].GetID()

		var direction encoding.Direction
		var err error
		if idx := findColumnOrdinalInIndex(zi.index, colID); idx != -1 {
			__antithesis_instrumentation__.Notify(575679)
			direction, err = zi.index.GetKeyColumnDirection(idx).ToEncodingDirection()
			if err != nil {
				__antithesis_instrumentation__.Notify(575680)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(575681)
			}
		} else {
			__antithesis_instrumentation__.Notify(575682)
			if idx := findColumnOrdinalInIndex(zi.table.GetPrimaryIndex(), colID); idx != -1 {
				__antithesis_instrumentation__.Notify(575683)
				direction, err = zi.table.GetPrimaryIndex().GetKeyColumnDirection(idx).ToEncodingDirection()
				if err != nil {
					__antithesis_instrumentation__.Notify(575684)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(575685)
				}
			} else {
				__antithesis_instrumentation__.Notify(575686)
				return nil, errors.New("ordering of equality column not found in index or primary key")
			}
		}
		__antithesis_instrumentation__.Notify(575678)
		ordering[i] = colinfo.ColumnOrderInfo{ColIdx: i, Direction: direction}
	}
	__antithesis_instrumentation__.Notify(575676)
	return ordering, nil
}

func (z *zigzagJoiner) matchBase(curRow rowenc.EncDatumRow, side int) (bool, error) {
	__antithesis_instrumentation__.Notify(575687)
	if len(curRow) == 0 {
		__antithesis_instrumentation__.Notify(575691)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(575692)
	}
	__antithesis_instrumentation__.Notify(575688)

	prevEqDatums := z.extractEqDatums(z.baseRow, z.prevSide())
	curEqDatums := z.extractEqDatums(curRow, side)

	eqColTypes := z.infos[side].eqColTypes()
	ordering, err := z.infos[side].eqOrdering()
	if err != nil {
		__antithesis_instrumentation__.Notify(575693)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(575694)
	}
	__antithesis_instrumentation__.Notify(575689)

	da := &tree.DatumAlloc{}
	cmp, err := prevEqDatums.Compare(eqColTypes, da, ordering, z.FlowCtx.EvalCtx, curEqDatums)
	if err != nil {
		__antithesis_instrumentation__.Notify(575695)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(575696)
	}
	__antithesis_instrumentation__.Notify(575690)
	return cmp == 0, nil
}

func (z *zigzagJoiner) emitFromContainers() (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(575697)
	right := z.prevSide()
	left := z.sideBefore(right)
	for !z.infos[right].container.IsEmpty() {
		__antithesis_instrumentation__.Notify(575699)
		leftRow := z.infos[left].container.Pop()
		rightRow := z.infos[right].container.Peek()

		if left == int(rightSide) {
			__antithesis_instrumentation__.Notify(575703)
			leftRow, rightRow = rightRow, leftRow
		} else {
			__antithesis_instrumentation__.Notify(575704)
		}
		__antithesis_instrumentation__.Notify(575700)
		renderedRow, err := z.render(leftRow, rightRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(575705)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575706)
		}
		__antithesis_instrumentation__.Notify(575701)
		if z.infos[left].container.IsEmpty() {
			__antithesis_instrumentation__.Notify(575707)
			z.infos[right].container.Pop()
		} else {
			__antithesis_instrumentation__.Notify(575708)
		}
		__antithesis_instrumentation__.Notify(575702)
		if renderedRow != nil {
			__antithesis_instrumentation__.Notify(575709)

			return renderedRow, nil
		} else {
			__antithesis_instrumentation__.Notify(575710)
		}
	}
	__antithesis_instrumentation__.Notify(575698)

	z.infos[left].container.Reset()
	z.infos[right].container.Reset()

	return nil, nil
}

func (z *zigzagJoiner) nextRow(ctx context.Context, txn *kv.Txn) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(575711)
	for {
		__antithesis_instrumentation__.Notify(575712)
		if err := z.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(575721)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575722)
		}
		__antithesis_instrumentation__.Notify(575713)

		if rowToEmit, err := z.emitFromContainers(); err != nil {
			__antithesis_instrumentation__.Notify(575723)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575724)
			if rowToEmit != nil {
				__antithesis_instrumentation__.Notify(575725)
				return rowToEmit, nil
			} else {
				__antithesis_instrumentation__.Notify(575726)
			}
		}
		__antithesis_instrumentation__.Notify(575714)

		if len(z.baseRow) == 0 {
			__antithesis_instrumentation__.Notify(575727)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(575728)
		}
		__antithesis_instrumentation__.Notify(575715)

		curInfo := z.infos[z.side]

		span, err := z.produceSpanFromBaseRow()
		if err != nil {
			__antithesis_instrumentation__.Notify(575729)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575730)
		}
		__antithesis_instrumentation__.Notify(575716)
		curInfo.key = span.Key

		err = curInfo.fetcher.StartScan(
			ctx,
			txn,
			roachpb.Spans{roachpb.Span{Key: curInfo.key, EndKey: curInfo.endKey}},
			rowinfra.DefaultBatchBytesLimit,
			zigzagJoinerBatchSize,
			z.FlowCtx.TraceKV,
			z.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(575731)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575732)
		}
		__antithesis_instrumentation__.Notify(575717)

		fetchedRow, err := z.fetchRow(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(575733)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575734)
		}
		__antithesis_instrumentation__.Notify(575718)

		if fetchedRow == nil {
			__antithesis_instrumentation__.Notify(575735)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(575736)
		}
		__antithesis_instrumentation__.Notify(575719)

		matched, err := z.matchBase(fetchedRow, z.side)
		if err != nil {
			__antithesis_instrumentation__.Notify(575737)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575738)
		}
		__antithesis_instrumentation__.Notify(575720)
		if matched {
			__antithesis_instrumentation__.Notify(575739)

			prevSide := z.prevSide()

			prevRow := z.rowAlloc.AllocRow(len(z.baseRow))
			copy(prevRow, z.baseRow)
			z.infos[prevSide].container.Push(prevRow)
			curRow := z.rowAlloc.AllocRow(len(fetchedRow))
			copy(curRow, fetchedRow)
			curInfo.container.Push(curRow)

			prevNext, err := z.collectAllMatches(ctx, prevSide)
			if err != nil {
				__antithesis_instrumentation__.Notify(575745)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(575746)
			}
			__antithesis_instrumentation__.Notify(575740)
			curNext, err := z.collectAllMatches(ctx, z.side)
			if err != nil {
				__antithesis_instrumentation__.Notify(575747)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(575748)
			}
			__antithesis_instrumentation__.Notify(575741)

			if len(prevNext) == 0 || func() bool {
				__antithesis_instrumentation__.Notify(575749)
				return len(curNext) == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(575750)
				z.baseRow = nil
				continue
			} else {
				__antithesis_instrumentation__.Notify(575751)
			}
			__antithesis_instrumentation__.Notify(575742)

			prevEqCols := z.extractEqDatums(prevNext, prevSide)
			currentEqCols := z.extractEqDatums(curNext, z.side)
			eqColTypes := curInfo.eqColTypes()
			ordering, err := curInfo.eqOrdering()
			if err != nil {
				__antithesis_instrumentation__.Notify(575752)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(575753)
			}
			__antithesis_instrumentation__.Notify(575743)
			da := &tree.DatumAlloc{}
			cmp, err := prevEqCols.Compare(eqColTypes, da, ordering, z.FlowCtx.EvalCtx, currentEqCols)
			if err != nil {
				__antithesis_instrumentation__.Notify(575754)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(575755)
			}
			__antithesis_instrumentation__.Notify(575744)

			if cmp < 0 {
				__antithesis_instrumentation__.Notify(575756)

				z.side = z.nextSide()
				z.baseRow = curNext
			} else {
				__antithesis_instrumentation__.Notify(575757)

				z.baseRow = prevNext
			}
		} else {
			__antithesis_instrumentation__.Notify(575758)

			z.baseRow = fetchedRow
			z.baseRow = z.rowAlloc.AllocRow(len(fetchedRow))
			copy(z.baseRow, fetchedRow)
			z.side = z.nextSide()
		}
	}
}

func (z *zigzagJoiner) nextSide() int {
	__antithesis_instrumentation__.Notify(575759)
	return (z.side + 1) % z.numTables
}

func (z *zigzagJoiner) prevSide() int {
	__antithesis_instrumentation__.Notify(575760)
	return z.sideBefore(z.side)
}

func (z *zigzagJoiner) sideBefore(side int) int {
	__antithesis_instrumentation__.Notify(575761)
	return (side + z.numTables - 1) % z.numTables
}

func (z *zigzagJoiner) collectAllMatches(
	ctx context.Context, side int,
) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(575762)
	matched := true
	var row rowenc.EncDatumRow
	for matched {
		__antithesis_instrumentation__.Notify(575764)
		var err error
		fetchedRow, err := z.fetchRowFromSide(ctx, side)
		row = z.rowAlloc.AllocRow(len(fetchedRow))
		copy(row, fetchedRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(575767)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575768)
		}
		__antithesis_instrumentation__.Notify(575765)
		matched, err = z.matchBase(row, side)
		if err != nil {
			__antithesis_instrumentation__.Notify(575769)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(575770)
		}
		__antithesis_instrumentation__.Notify(575766)
		if matched {
			__antithesis_instrumentation__.Notify(575771)
			z.infos[side].container.Push(row)
		} else {
			__antithesis_instrumentation__.Notify(575772)
		}
	}
	__antithesis_instrumentation__.Notify(575763)
	return row, nil
}

func (z *zigzagJoiner) maybeFetchInitialRow() error {
	__antithesis_instrumentation__.Notify(575773)
	if !z.fetchedInititalRow {
		__antithesis_instrumentation__.Notify(575775)
		z.fetchedInititalRow = true

		curInfo := z.infos[z.side]
		err := curInfo.fetcher.StartScan(
			z.Ctx,
			z.FlowCtx.Txn,
			roachpb.Spans{roachpb.Span{Key: curInfo.key, EndKey: curInfo.endKey}},
			rowinfra.DefaultBatchBytesLimit,
			zigzagJoinerBatchSize,
			z.FlowCtx.TraceKV,
			z.EvalCtx.TestingKnobs.ForceProductionBatchSizes,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(575778)
			log.Errorf(z.Ctx, "scan error: %s", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575779)
		}
		__antithesis_instrumentation__.Notify(575776)
		fetchedRow, err := z.fetchRow(z.Ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(575780)
			return scrub.UnwrapScrubError(err)
		} else {
			__antithesis_instrumentation__.Notify(575781)
		}
		__antithesis_instrumentation__.Notify(575777)
		z.baseRow = z.rowAlloc.AllocRow(len(fetchedRow))
		copy(z.baseRow, fetchedRow)
		z.side = z.nextSide()
	} else {
		__antithesis_instrumentation__.Notify(575782)
	}
	__antithesis_instrumentation__.Notify(575774)
	return nil
}

func (z *zigzagJoiner) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575783)
	for z.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(575785)
		if err := z.maybeFetchInitialRow(); err != nil {
			__antithesis_instrumentation__.Notify(575789)
			z.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(575790)
		}
		__antithesis_instrumentation__.Notify(575786)
		row, err := z.nextRow(z.Ctx, z.FlowCtx.Txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(575791)
			z.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(575792)
		}
		__antithesis_instrumentation__.Notify(575787)
		if row == nil {
			__antithesis_instrumentation__.Notify(575793)
			z.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(575794)
		}
		__antithesis_instrumentation__.Notify(575788)

		if outRow := z.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(575795)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(575796)
		}
	}
	__antithesis_instrumentation__.Notify(575784)

	return nil, z.DrainHelper()
}

func (z *zigzagJoiner) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(575797)
	z.close()
}

func (z *zigzagJoiner) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(575798)
	z.scanStats = execinfra.GetScanStats(z.Ctx)

	kvStats := execinfrapb.KVStats{
		BytesRead:      optional.MakeUint(uint64(z.getBytesRead())),
		ContentionTime: optional.MakeTimeValue(execinfra.GetCumulativeContentionTime(z.Ctx)),
	}
	execinfra.PopulateKVMVCCStats(&kvStats, &z.scanStats)
	for i := range z.infos {
		__antithesis_instrumentation__.Notify(575800)
		fis, ok := getFetcherInputStats(z.infos[i].fetcher)
		if !ok {
			__antithesis_instrumentation__.Notify(575802)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(575803)
		}
		__antithesis_instrumentation__.Notify(575801)
		kvStats.TuplesRead.MaybeAdd(fis.NumTuples)
		kvStats.KVTime.MaybeAdd(fis.WaitTime)
	}
	__antithesis_instrumentation__.Notify(575799)
	return &execinfrapb.ComponentStats{
		KV:     kvStats,
		Output: z.OutputHelper.Stats(),
	}
}

func (z *zigzagJoiner) getBytesRead() int64 {
	__antithesis_instrumentation__.Notify(575804)
	var bytesRead int64
	for i := range z.infos {
		__antithesis_instrumentation__.Notify(575806)
		bytesRead += z.infos[i].fetcher.GetBytesRead()
	}
	__antithesis_instrumentation__.Notify(575805)
	return bytesRead
}

func (z *zigzagJoiner) getRowsRead() int64 {
	__antithesis_instrumentation__.Notify(575807)
	var rowsRead int64
	for i := range z.infos {
		__antithesis_instrumentation__.Notify(575809)
		rowsRead += z.infos[i].rowsRead
	}
	__antithesis_instrumentation__.Notify(575808)
	return rowsRead
}

func (z *zigzagJoiner) generateMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(575810)
	trailingMeta := make([]execinfrapb.ProducerMetadata, 1, 2)
	meta := &trailingMeta[0]
	meta.Metrics = execinfrapb.GetMetricsMeta()
	meta.Metrics.BytesRead = z.getBytesRead()
	meta.Metrics.RowsRead = z.getRowsRead()
	if tfs := execinfra.GetLeafTxnFinalState(z.Ctx, z.FlowCtx.Txn); tfs != nil {
		__antithesis_instrumentation__.Notify(575812)
		trailingMeta = append(trailingMeta, execinfrapb.ProducerMetadata{LeafTxnFinalState: tfs})
	} else {
		__antithesis_instrumentation__.Notify(575813)
	}
	__antithesis_instrumentation__.Notify(575811)
	return trailingMeta
}

func (z *zigzagJoiner) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(575814)
	return 0
}

func (z *zigzagJoiner) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(575815)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
