package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type joinReaderSpanGenerator interface {
	generateSpans(ctx context.Context, rows []rowenc.EncDatumRow) (roachpb.Spans, error)

	getMatchingRowIndices(key roachpb.Key) []int

	maxLookupCols() int

	close(context.Context)
}

var _ joinReaderSpanGenerator = &defaultSpanGenerator{}
var _ joinReaderSpanGenerator = &multiSpanGenerator{}
var _ joinReaderSpanGenerator = &localityOptimizedSpanGenerator{}

type defaultSpanGenerator struct {
	spanBuilder  span.Builder
	spanSplitter span.Splitter
	lookupCols   []uint32

	indexKeyRow rowenc.EncDatumRow

	keyToInputRowIndices map[string][]int

	scratchSpans roachpb.Spans

	memAcc *mon.BoundAccount
}

func (g *defaultSpanGenerator) init(
	evalCtx *tree.EvalContext,
	codec keys.SQLCodec,
	fetchSpec *descpb.IndexFetchSpec,
	splitFamilyIDs []descpb.FamilyID,
	keyToInputRowIndices map[string][]int,
	lookupCols []uint32,
	memAcc *mon.BoundAccount,
) error {
	g.spanBuilder.InitWithFetchSpec(evalCtx, codec, fetchSpec)
	g.spanSplitter = span.MakeSplitterWithFamilyIDs(len(fetchSpec.KeyFullColumns()), splitFamilyIDs)
	g.lookupCols = lookupCols
	if len(lookupCols) > len(fetchSpec.KeyAndSuffixColumns) {
		return errors.AssertionFailedf(
			"%d lookup columns specified, expecting at most %d", len(lookupCols), len(fetchSpec.KeyAndSuffixColumns),
		)
	}
	g.indexKeyRow = nil
	g.keyToInputRowIndices = keyToInputRowIndices
	g.scratchSpans = nil
	g.memAcc = memAcc
	return nil
}

func (g *defaultSpanGenerator) generateSpan(
	row rowenc.EncDatumRow,
) (_ roachpb.Span, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(573548)
	g.indexKeyRow = g.indexKeyRow[:0]
	for _, id := range g.lookupCols {
		__antithesis_instrumentation__.Notify(573550)
		g.indexKeyRow = append(g.indexKeyRow, row[id])
	}
	__antithesis_instrumentation__.Notify(573549)
	return g.spanBuilder.SpanFromEncDatums(g.indexKeyRow)
}

func (g *defaultSpanGenerator) hasNullLookupColumn(row rowenc.EncDatumRow) bool {
	__antithesis_instrumentation__.Notify(573551)
	for _, colIdx := range g.lookupCols {
		__antithesis_instrumentation__.Notify(573553)
		if row[colIdx].IsNull() {
			__antithesis_instrumentation__.Notify(573554)
			return true
		} else {
			__antithesis_instrumentation__.Notify(573555)
		}
	}
	__antithesis_instrumentation__.Notify(573552)
	return false
}

func (g *defaultSpanGenerator) generateSpans(
	ctx context.Context, rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573556)

	for k := range g.keyToInputRowIndices {
		__antithesis_instrumentation__.Notify(573560)
		delete(g.keyToInputRowIndices, k)
	}
	__antithesis_instrumentation__.Notify(573557)

	g.scratchSpans = g.scratchSpans[:0]
	for i, inputRow := range rows {
		__antithesis_instrumentation__.Notify(573561)
		if g.hasNullLookupColumn(inputRow) {
			__antithesis_instrumentation__.Notify(573564)
			continue
		} else {
			__antithesis_instrumentation__.Notify(573565)
		}
		__antithesis_instrumentation__.Notify(573562)
		generatedSpan, containsNull, err := g.generateSpan(inputRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(573566)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(573567)
		}
		__antithesis_instrumentation__.Notify(573563)
		if g.keyToInputRowIndices == nil {
			__antithesis_instrumentation__.Notify(573568)

			g.scratchSpans = g.spanSplitter.MaybeSplitSpanIntoSeparateFamilies(
				g.scratchSpans, generatedSpan, len(g.lookupCols), containsNull)
		} else {
			__antithesis_instrumentation__.Notify(573569)
			inputRowIndices := g.keyToInputRowIndices[string(generatedSpan.Key)]
			if inputRowIndices == nil {
				__antithesis_instrumentation__.Notify(573571)
				g.scratchSpans = g.spanSplitter.MaybeSplitSpanIntoSeparateFamilies(
					g.scratchSpans, generatedSpan, len(g.lookupCols), containsNull)
			} else {
				__antithesis_instrumentation__.Notify(573572)
			}
			__antithesis_instrumentation__.Notify(573570)
			g.keyToInputRowIndices[string(generatedSpan.Key)] = append(inputRowIndices, i)
		}
	}
	__antithesis_instrumentation__.Notify(573558)

	if err := g.memAcc.ResizeTo(ctx, g.memUsage()); err != nil {
		__antithesis_instrumentation__.Notify(573573)
		return nil, addWorkmemHint(err)
	} else {
		__antithesis_instrumentation__.Notify(573574)
	}
	__antithesis_instrumentation__.Notify(573559)

	return g.scratchSpans, nil
}

func (g *defaultSpanGenerator) getMatchingRowIndices(key roachpb.Key) []int {
	__antithesis_instrumentation__.Notify(573575)
	return g.keyToInputRowIndices[string(key)]
}

func (g *defaultSpanGenerator) maxLookupCols() int {
	__antithesis_instrumentation__.Notify(573576)
	return len(g.lookupCols)
}

func (g *defaultSpanGenerator) memUsage() int64 {
	__antithesis_instrumentation__.Notify(573577)

	var size int64
	for k, v := range g.keyToInputRowIndices {
		__antithesis_instrumentation__.Notify(573579)
		size += memsize.MapEntryOverhead
		size += memsize.String + int64(len(k))
		size += memsize.IntSliceOverhead + memsize.Int*int64(cap(v))
	}
	__antithesis_instrumentation__.Notify(573578)
	return size
}

func (g *defaultSpanGenerator) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573580)
	g.memAcc.Close(ctx)
	*g = defaultSpanGenerator{}
}

type spanRowIndex struct {
	span       roachpb.Span
	rowIndices []int
}

type spanRowIndices []spanRowIndex

func (s spanRowIndices) Len() int { __antithesis_instrumentation__.Notify(573581); return len(s) }
func (s spanRowIndices) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(573582)
	s[i], s[j] = s[j], s[i]
}
func (s spanRowIndices) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(573583)
	return s[i].span.Key.Compare(s[j].span.Key) < 0
}

var _ sort.Interface = &spanRowIndices{}

func (s spanRowIndices) memUsage() int64 {
	__antithesis_instrumentation__.Notify(573584)

	sCap := s[:cap(s)]
	size := int64(unsafe.Sizeof(spanRowIndices{}))
	for i := range sCap {
		__antithesis_instrumentation__.Notify(573586)
		size += sCap[i].span.MemUsage()
		size += memsize.IntSliceOverhead + memsize.Int*int64(cap(sCap[i].rowIndices))
	}
	__antithesis_instrumentation__.Notify(573585)
	return size
}

type multiSpanGenerator struct {
	spanBuilder  span.Builder
	spanSplitter span.Splitter

	indexColInfos []multiSpanGeneratorColInfo

	indexKeyRows  []rowenc.EncDatumRow
	indexKeySpans roachpb.Spans

	keyToInputRowIndices map[string][]int

	spanToInputRowIndices spanRowIndices

	spansCount int

	indexOrds util.FastIntSet

	fetchedOrdToIndexKeyOrd util.FastIntMap

	numInputCols int

	inequalityColIdx int

	scratchSpans roachpb.Spans

	memAcc *mon.BoundAccount
}

type multiSpanGeneratorColInfo interface {
	String() string
}

type multiSpanGeneratorValuesColInfo struct {
	constVals tree.Datums
}

func (i multiSpanGeneratorValuesColInfo) String() string {
	__antithesis_instrumentation__.Notify(573587)
	return fmt.Sprintf("[constVals: %s]", i.constVals.String())
}

type multiSpanGeneratorIndexVarColInfo struct {
	inputRowIdx int
}

func (i multiSpanGeneratorIndexVarColInfo) String() string {
	__antithesis_instrumentation__.Notify(573588)
	return fmt.Sprintf("[inputRowIdx: %d]", i.inputRowIdx)
}

type multiSpanGeneratorInequalityColInfo struct {
	start          tree.Datum
	startInclusive bool
	end            tree.Datum
	endInclusive   bool
}

func (i multiSpanGeneratorInequalityColInfo) String() string {
	__antithesis_instrumentation__.Notify(573589)
	var startBoundary byte
	if i.startInclusive {
		__antithesis_instrumentation__.Notify(573592)
		startBoundary = '['
	} else {
		__antithesis_instrumentation__.Notify(573593)
		startBoundary = '('
	}
	__antithesis_instrumentation__.Notify(573590)
	var endBoundary rune
	if i.endInclusive {
		__antithesis_instrumentation__.Notify(573594)
		endBoundary = ']'
	} else {
		__antithesis_instrumentation__.Notify(573595)
		endBoundary = ')'
	}
	__antithesis_instrumentation__.Notify(573591)
	return fmt.Sprintf("%c%v - %v%c", startBoundary, i.start, i.end, endBoundary)
}

var _ multiSpanGeneratorColInfo = &multiSpanGeneratorValuesColInfo{}
var _ multiSpanGeneratorColInfo = &multiSpanGeneratorIndexVarColInfo{}
var _ multiSpanGeneratorColInfo = &multiSpanGeneratorInequalityColInfo{}

func (g *multiSpanGenerator) maxLookupCols() int {
	__antithesis_instrumentation__.Notify(573596)
	return len(g.indexColInfos)
}

func (g *multiSpanGenerator) init(
	evalCtx *tree.EvalContext,
	codec keys.SQLCodec,
	fetchSpec *descpb.IndexFetchSpec,
	splitFamilyIDs []descpb.FamilyID,
	numInputCols int,
	exprHelper *execinfrapb.ExprHelper,
	fetchedOrdToIndexKeyOrd util.FastIntMap,
	memAcc *mon.BoundAccount,
) error {
	g.spanBuilder.InitWithFetchSpec(evalCtx, codec, fetchSpec)
	g.spanSplitter = span.MakeSplitterWithFamilyIDs(len(fetchSpec.KeyFullColumns()), splitFamilyIDs)
	g.numInputCols = numInputCols
	g.keyToInputRowIndices = make(map[string][]int)
	g.fetchedOrdToIndexKeyOrd = fetchedOrdToIndexKeyOrd
	g.inequalityColIdx = -1
	g.memAcc = memAcc

	g.spansCount = 1

	g.indexColInfos = make([]multiSpanGeneratorColInfo, 0, len(fetchSpec.KeyAndSuffixColumns))
	if err := g.fillInIndexColInfos(exprHelper.Expr); err != nil {
		return err
	}

	lookupColsCount := len(g.indexColInfos)
	if lookupColsCount != g.indexOrds.Len() {
		return errors.AssertionFailedf(
			"columns in the join condition do not form a prefix on the index",
		)
	}
	if lookupColsCount > len(fetchSpec.KeyAndSuffixColumns) {
		return errors.AssertionFailedf(
			"%d lookup columns specified, expecting at most %d", lookupColsCount, len(fetchSpec.KeyAndSuffixColumns),
		)
	}

	g.indexKeyRows = make([]rowenc.EncDatumRow, 1, g.spansCount)
	g.indexKeyRows[0] = make(rowenc.EncDatumRow, 0, lookupColsCount)
	for _, info := range g.indexColInfos {
		if valuesInfo, ok := info.(multiSpanGeneratorValuesColInfo); ok {
			for i, n := 0, len(g.indexKeyRows); i < n; i++ {
				indexKeyRow := g.indexKeyRows[i]
				for j := 1; j < len(valuesInfo.constVals); j++ {
					newIndexKeyRow := make(rowenc.EncDatumRow, len(indexKeyRow), lookupColsCount)
					copy(newIndexKeyRow, indexKeyRow)
					newIndexKeyRow = append(newIndexKeyRow, rowenc.EncDatum{Datum: valuesInfo.constVals[j]})
					g.indexKeyRows = append(g.indexKeyRows, newIndexKeyRow)
				}
				g.indexKeyRows[i] = append(indexKeyRow, rowenc.EncDatum{Datum: valuesInfo.constVals[0]})
			}
		} else {
			for i := 0; i < len(g.indexKeyRows); i++ {

				g.indexKeyRows[i] = append(g.indexKeyRows[i], rowenc.EncDatum{})
			}
		}
	}

	g.indexKeySpans = make(roachpb.Spans, 0, g.spansCount)
	return nil
}

func (g *multiSpanGenerator) fillInIndexColInfos(expr tree.TypedExpr) error {
	__antithesis_instrumentation__.Notify(573597)
	switch t := expr.(type) {
	case *tree.AndExpr:
		__antithesis_instrumentation__.Notify(573599)
		if err := g.fillInIndexColInfos(t.Left.(tree.TypedExpr)); err != nil {
			__antithesis_instrumentation__.Notify(573610)
			return err
		} else {
			__antithesis_instrumentation__.Notify(573611)
		}
		__antithesis_instrumentation__.Notify(573600)
		return g.fillInIndexColInfos(t.Right.(tree.TypedExpr))

	case *tree.ComparisonExpr:
		__antithesis_instrumentation__.Notify(573601)
		setOfVals := false
		inequality := false
		switch t.Operator.Symbol {
		case treecmp.EQ, treecmp.In:
			__antithesis_instrumentation__.Notify(573612)
			setOfVals = true
		case treecmp.GE, treecmp.LE, treecmp.GT, treecmp.LT:
			__antithesis_instrumentation__.Notify(573613)
			inequality = true
		default:
			__antithesis_instrumentation__.Notify(573614)

			return errors.AssertionFailedf("comparison operator not supported. Found %s", t.Operator)
		}
		__antithesis_instrumentation__.Notify(573602)

		tabOrd := -1

		var info multiSpanGeneratorColInfo

		getInfo := func(typedExpr tree.TypedExpr) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(573615)
			switch t := typedExpr.(type) {
			case *tree.IndexedVar:
				__antithesis_instrumentation__.Notify(573617)

				if t.Idx >= g.numInputCols {
					__antithesis_instrumentation__.Notify(573620)
					tabOrd = t.Idx - g.numInputCols
				} else {
					__antithesis_instrumentation__.Notify(573621)
					info = multiSpanGeneratorIndexVarColInfo{inputRowIdx: t.Idx}
				}

			case tree.Datum:
				__antithesis_instrumentation__.Notify(573618)
				if setOfVals {
					__antithesis_instrumentation__.Notify(573622)
					var values tree.Datums
					switch t.ResolvedType().Family() {
					case types.TupleFamily:
						__antithesis_instrumentation__.Notify(573624)
						values = t.(*tree.DTuple).D
					default:
						__antithesis_instrumentation__.Notify(573625)
						values = tree.Datums{t}
					}
					__antithesis_instrumentation__.Notify(573623)

					info = multiSpanGeneratorValuesColInfo{constVals: values}
					g.spansCount *= len(values)
				} else {
					__antithesis_instrumentation__.Notify(573626)
					return t, nil
				}

			default:
				__antithesis_instrumentation__.Notify(573619)
				return nil, errors.AssertionFailedf("unhandled comparison argument type %T", t)
			}
			__antithesis_instrumentation__.Notify(573616)
			return nil, nil
		}
		__antithesis_instrumentation__.Notify(573603)

		lval, err := getInfo(t.Left.(tree.TypedExpr))
		if err != nil {
			__antithesis_instrumentation__.Notify(573627)
			return err
		} else {
			__antithesis_instrumentation__.Notify(573628)
		}
		__antithesis_instrumentation__.Notify(573604)

		rval, err := getInfo(t.Right.(tree.TypedExpr))
		if err != nil {
			__antithesis_instrumentation__.Notify(573629)
			return err
		} else {
			__antithesis_instrumentation__.Notify(573630)
		}
		__antithesis_instrumentation__.Notify(573605)

		idxOrd, ok := g.fetchedOrdToIndexKeyOrd.Get(tabOrd)
		if !ok {
			__antithesis_instrumentation__.Notify(573631)
			return errors.AssertionFailedf("table column %d not found in index", tabOrd)
		} else {
			__antithesis_instrumentation__.Notify(573632)
		}
		__antithesis_instrumentation__.Notify(573606)

		if len(g.indexColInfos) <= idxOrd {
			__antithesis_instrumentation__.Notify(573633)
			g.indexColInfos = g.indexColInfos[:idxOrd+1]
		} else {
			__antithesis_instrumentation__.Notify(573634)
		}
		__antithesis_instrumentation__.Notify(573607)

		if inequality {
			__antithesis_instrumentation__.Notify(573635)

			colInfo := g.indexColInfos[idxOrd]
			var inequalityInfo multiSpanGeneratorInequalityColInfo
			if colInfo != nil {
				__antithesis_instrumentation__.Notify(573639)
				inequalityInfo, ok = colInfo.(multiSpanGeneratorInequalityColInfo)
				if !ok {
					__antithesis_instrumentation__.Notify(573640)
					return errors.AssertionFailedf("unexpected colinfo type (%d): %T", idxOrd, colInfo)
				} else {
					__antithesis_instrumentation__.Notify(573641)
				}
			} else {
				__antithesis_instrumentation__.Notify(573642)
			}
			__antithesis_instrumentation__.Notify(573636)

			if lval != nil {
				__antithesis_instrumentation__.Notify(573643)
				if t.Operator.Symbol == treecmp.LT || func() bool {
					__antithesis_instrumentation__.Notify(573644)
					return t.Operator.Symbol == treecmp.LE == true
				}() == true {
					__antithesis_instrumentation__.Notify(573645)
					inequalityInfo.start = lval
					inequalityInfo.startInclusive = t.Operator.Symbol == treecmp.LE
				} else {
					__antithesis_instrumentation__.Notify(573646)
					inequalityInfo.end = lval
					inequalityInfo.endInclusive = t.Operator.Symbol == treecmp.GE
				}
			} else {
				__antithesis_instrumentation__.Notify(573647)
			}
			__antithesis_instrumentation__.Notify(573637)

			if rval != nil {
				__antithesis_instrumentation__.Notify(573648)
				if t.Operator.Symbol == treecmp.LT || func() bool {
					__antithesis_instrumentation__.Notify(573649)
					return t.Operator.Symbol == treecmp.LE == true
				}() == true {
					__antithesis_instrumentation__.Notify(573650)
					inequalityInfo.end = rval
					inequalityInfo.endInclusive = t.Operator.Symbol == treecmp.LE
				} else {
					__antithesis_instrumentation__.Notify(573651)
					inequalityInfo.start = rval
					inequalityInfo.startInclusive = t.Operator.Symbol == treecmp.GE
				}
			} else {
				__antithesis_instrumentation__.Notify(573652)
			}
			__antithesis_instrumentation__.Notify(573638)
			info = inequalityInfo
			g.inequalityColIdx = idxOrd
		} else {
			__antithesis_instrumentation__.Notify(573653)
		}
		__antithesis_instrumentation__.Notify(573608)

		g.indexColInfos[idxOrd] = info
		g.indexOrds.Add(idxOrd)

	default:
		__antithesis_instrumentation__.Notify(573609)
		return errors.AssertionFailedf("unhandled expression type %T", t)
	}
	__antithesis_instrumentation__.Notify(573598)
	return nil
}

func (g *multiSpanGenerator) generateNonNullSpans(row rowenc.EncDatumRow) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573654)

	for i := 0; i < len(g.indexKeyRows); i++ {
		__antithesis_instrumentation__.Notify(573658)
		for j, info := range g.indexColInfos {
			__antithesis_instrumentation__.Notify(573659)
			if inf, ok := info.(multiSpanGeneratorIndexVarColInfo); ok {
				__antithesis_instrumentation__.Notify(573660)
				g.indexKeyRows[i][j] = row[inf.inputRowIdx]
			} else {
				__antithesis_instrumentation__.Notify(573661)
			}
		}
	}
	__antithesis_instrumentation__.Notify(573655)

	g.indexKeySpans = g.indexKeySpans[:0]

	var inequalityInfo multiSpanGeneratorInequalityColInfo
	if g.inequalityColIdx != -1 {
		__antithesis_instrumentation__.Notify(573662)
		inequalityInfo = g.indexColInfos[g.inequalityColIdx].(multiSpanGeneratorInequalityColInfo)
	} else {
		__antithesis_instrumentation__.Notify(573663)
	}
	__antithesis_instrumentation__.Notify(573656)

	for _, indexKeyRow := range g.indexKeyRows {
		__antithesis_instrumentation__.Notify(573664)
		var s roachpb.Span
		var err error
		var containsNull bool
		if g.inequalityColIdx == -1 {
			__antithesis_instrumentation__.Notify(573667)
			s, containsNull, err = g.spanBuilder.SpanFromEncDatums(indexKeyRow[:len(g.indexColInfos)])
		} else {
			__antithesis_instrumentation__.Notify(573668)
			s, containsNull, err = g.spanBuilder.SpanFromEncDatumsWithRange(indexKeyRow, len(g.indexColInfos),
				inequalityInfo.start, inequalityInfo.startInclusive, inequalityInfo.end, inequalityInfo.endInclusive)
		}
		__antithesis_instrumentation__.Notify(573665)

		if err != nil {
			__antithesis_instrumentation__.Notify(573669)
			return roachpb.Spans{}, err
		} else {
			__antithesis_instrumentation__.Notify(573670)
		}
		__antithesis_instrumentation__.Notify(573666)

		if !containsNull {
			__antithesis_instrumentation__.Notify(573671)
			g.indexKeySpans = append(g.indexKeySpans, s)
		} else {
			__antithesis_instrumentation__.Notify(573672)
		}
	}
	__antithesis_instrumentation__.Notify(573657)

	return g.indexKeySpans, nil
}

func (s *spanRowIndices) findInputRowIndicesByKey(key roachpb.Key) []int {
	__antithesis_instrumentation__.Notify(573673)
	i, j := 0, s.Len()
	for i < j {
		__antithesis_instrumentation__.Notify(573675)
		h := (i + j) >> 1
		sp := (*s)[h]
		switch sp.span.CompareKey(key) {
		case 0:
			__antithesis_instrumentation__.Notify(573676)
			return sp.rowIndices
		case -1:
			__antithesis_instrumentation__.Notify(573677)
			j = h
		case 1:
			__antithesis_instrumentation__.Notify(573678)
			i = h + 1
		default:
			__antithesis_instrumentation__.Notify(573679)
		}
	}
	__antithesis_instrumentation__.Notify(573674)

	return nil
}

func (g *multiSpanGenerator) generateSpans(
	ctx context.Context, rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573680)

	for k := range g.keyToInputRowIndices {
		__antithesis_instrumentation__.Notify(573685)
		delete(g.keyToInputRowIndices, k)
	}
	__antithesis_instrumentation__.Notify(573681)
	g.spanToInputRowIndices = g.spanToInputRowIndices[:0]

	g.scratchSpans = g.scratchSpans[:0]
	for i, inputRow := range rows {
		__antithesis_instrumentation__.Notify(573686)
		generatedSpans, err := g.generateNonNullSpans(inputRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(573688)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(573689)
		}
		__antithesis_instrumentation__.Notify(573687)
		for j := range generatedSpans {
			__antithesis_instrumentation__.Notify(573690)
			generatedSpan := &generatedSpans[j]
			inputRowIndices := g.keyToInputRowIndices[string(generatedSpan.Key)]
			if inputRowIndices == nil {
				__antithesis_instrumentation__.Notify(573692)

				if g.inequalityColIdx != -1 {
					__antithesis_instrumentation__.Notify(573693)
					g.scratchSpans = append(g.scratchSpans, *generatedSpan)
				} else {
					__antithesis_instrumentation__.Notify(573694)
					g.scratchSpans = g.spanSplitter.MaybeSplitSpanIntoSeparateFamilies(
						g.scratchSpans, *generatedSpan, len(g.indexColInfos), false)
				}
			} else {
				__antithesis_instrumentation__.Notify(573695)
			}
			__antithesis_instrumentation__.Notify(573691)

			g.keyToInputRowIndices[string(generatedSpan.Key)] = append(inputRowIndices, i)
		}
	}
	__antithesis_instrumentation__.Notify(573682)

	if g.inequalityColIdx != -1 {
		__antithesis_instrumentation__.Notify(573696)
		for _, s := range g.scratchSpans {
			__antithesis_instrumentation__.Notify(573698)
			g.spanToInputRowIndices = append(g.spanToInputRowIndices, spanRowIndex{span: s, rowIndices: g.keyToInputRowIndices[string(s.Key)]})
		}
		__antithesis_instrumentation__.Notify(573697)
		sort.Sort(g.spanToInputRowIndices)

		for k := range g.keyToInputRowIndices {
			__antithesis_instrumentation__.Notify(573699)
			delete(g.keyToInputRowIndices, k)
		}
	} else {
		__antithesis_instrumentation__.Notify(573700)
	}
	__antithesis_instrumentation__.Notify(573683)

	if err := g.memAcc.ResizeTo(ctx, g.memUsage()); err != nil {
		__antithesis_instrumentation__.Notify(573701)
		return nil, addWorkmemHint(err)
	} else {
		__antithesis_instrumentation__.Notify(573702)
	}
	__antithesis_instrumentation__.Notify(573684)

	return g.scratchSpans, nil
}

func (g *multiSpanGenerator) getMatchingRowIndices(key roachpb.Key) []int {
	__antithesis_instrumentation__.Notify(573703)
	if g.inequalityColIdx != -1 {
		__antithesis_instrumentation__.Notify(573705)
		return g.spanToInputRowIndices.findInputRowIndicesByKey(key)
	} else {
		__antithesis_instrumentation__.Notify(573706)
	}
	__antithesis_instrumentation__.Notify(573704)
	return g.keyToInputRowIndices[string(key)]
}

func (g *multiSpanGenerator) memUsage() int64 {
	__antithesis_instrumentation__.Notify(573707)

	var size int64
	for k, v := range g.keyToInputRowIndices {
		__antithesis_instrumentation__.Notify(573709)
		size += memsize.MapEntryOverhead
		size += memsize.String + int64(len(k))
		size += memsize.IntSliceOverhead + memsize.Int*int64(cap(v))
	}
	__antithesis_instrumentation__.Notify(573708)

	size += g.spanToInputRowIndices.memUsage()
	return size
}

func (g *multiSpanGenerator) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573710)
	g.memAcc.Close(ctx)
	*g = multiSpanGenerator{}
}

type localityOptimizedSpanGenerator struct {
	localSpanGen  multiSpanGenerator
	remoteSpanGen multiSpanGenerator
}

func (g *localityOptimizedSpanGenerator) init(
	evalCtx *tree.EvalContext,
	codec keys.SQLCodec,
	fetchSpec *descpb.IndexFetchSpec,
	splitFamilyIDs []descpb.FamilyID,
	numInputCols int,
	localExprHelper *execinfrapb.ExprHelper,
	remoteExprHelper *execinfrapb.ExprHelper,
	fetchedOrdToIndexKeyOrd util.FastIntMap,
	localSpanGenMemAcc *mon.BoundAccount,
	remoteSpanGenMemAcc *mon.BoundAccount,
) error {
	if err := g.localSpanGen.init(
		evalCtx, codec, fetchSpec, splitFamilyIDs,
		numInputCols, localExprHelper, fetchedOrdToIndexKeyOrd, localSpanGenMemAcc,
	); err != nil {
		return err
	}
	if err := g.remoteSpanGen.init(
		evalCtx, codec, fetchSpec, splitFamilyIDs,
		numInputCols, remoteExprHelper, fetchedOrdToIndexKeyOrd, remoteSpanGenMemAcc,
	); err != nil {
		return err
	}

	localLookupCols := g.localSpanGen.maxLookupCols()
	remoteLookupCols := g.remoteSpanGen.maxLookupCols()
	if localLookupCols != remoteLookupCols {
		return errors.AssertionFailedf(
			"local lookup cols (%d) != remote lookup cols (%d)", localLookupCols, remoteLookupCols,
		)
	}
	return nil
}

func (g *localityOptimizedSpanGenerator) maxLookupCols() int {
	__antithesis_instrumentation__.Notify(573711)

	return g.localSpanGen.maxLookupCols()
}

func (g *localityOptimizedSpanGenerator) generateSpans(
	ctx context.Context, rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573712)
	return g.localSpanGen.generateSpans(ctx, rows)
}

func (g *localityOptimizedSpanGenerator) generateRemoteSpans(
	ctx context.Context, rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573713)
	return g.remoteSpanGen.generateSpans(ctx, rows)
}

func (g *localityOptimizedSpanGenerator) getMatchingRowIndices(key roachpb.Key) []int {
	__antithesis_instrumentation__.Notify(573714)
	if res := g.localSpanGen.getMatchingRowIndices(key); len(res) > 0 {
		__antithesis_instrumentation__.Notify(573716)
		return res
	} else {
		__antithesis_instrumentation__.Notify(573717)
	}
	__antithesis_instrumentation__.Notify(573715)
	return g.remoteSpanGen.getMatchingRowIndices(key)
}

func (g *localityOptimizedSpanGenerator) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573718)
	g.localSpanGen.close(ctx)
	g.remoteSpanGen.close(ctx)
	*g = localityOptimizedSpanGenerator{}
}
