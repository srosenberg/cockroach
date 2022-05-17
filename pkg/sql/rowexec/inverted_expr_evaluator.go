package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/errors"
)

type KeyIndex = int

type setContainer []KeyIndex

func (s setContainer) Len() int {
	__antithesis_instrumentation__.Notify(572607)
	return len(s)
}

func (s setContainer) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(572608)
	return s[i] < s[j]
}

func (s setContainer) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(572609)
	s[i], s[j] = s[j], s[i]
}

func unionSetContainers(a, b setContainer) setContainer {
	__antithesis_instrumentation__.Notify(572610)
	if len(a) == 0 {
		__antithesis_instrumentation__.Notify(572616)
		return b
	} else {
		__antithesis_instrumentation__.Notify(572617)
	}
	__antithesis_instrumentation__.Notify(572611)
	if len(b) == 0 {
		__antithesis_instrumentation__.Notify(572618)
		return a
	} else {
		__antithesis_instrumentation__.Notify(572619)
	}
	__antithesis_instrumentation__.Notify(572612)
	var out setContainer
	var i, j int
	for i < len(a) && func() bool {
		__antithesis_instrumentation__.Notify(572620)
		return j < len(b) == true
	}() == true {
		__antithesis_instrumentation__.Notify(572621)
		if a[i] < b[j] {
			__antithesis_instrumentation__.Notify(572622)
			out = append(out, a[i])
			i++
		} else {
			__antithesis_instrumentation__.Notify(572623)
			if a[i] > b[j] {
				__antithesis_instrumentation__.Notify(572624)
				out = append(out, b[j])
				j++
			} else {
				__antithesis_instrumentation__.Notify(572625)
				out = append(out, a[i])
				i++
				j++
			}
		}
	}
	__antithesis_instrumentation__.Notify(572613)
	for ; i < len(a); i++ {
		__antithesis_instrumentation__.Notify(572626)
		out = append(out, a[i])
	}
	__antithesis_instrumentation__.Notify(572614)
	for ; j < len(b); j++ {
		__antithesis_instrumentation__.Notify(572627)
		out = append(out, b[j])
	}
	__antithesis_instrumentation__.Notify(572615)
	return out
}

func intersectSetContainers(a, b setContainer) setContainer {
	__antithesis_instrumentation__.Notify(572628)
	var out setContainer
	var i, j int

	for i < len(a) && func() bool {
		__antithesis_instrumentation__.Notify(572630)
		return j < len(b) == true
	}() == true {
		__antithesis_instrumentation__.Notify(572631)
		if a[i] < b[j] {
			__antithesis_instrumentation__.Notify(572632)
			i++
		} else {
			__antithesis_instrumentation__.Notify(572633)
			if a[i] > b[j] {
				__antithesis_instrumentation__.Notify(572634)
				j++
			} else {
				__antithesis_instrumentation__.Notify(572635)
				out = append(out, a[i])
				i++
				j++
			}
		}
	}
	__antithesis_instrumentation__.Notify(572629)
	return out
}

type setExpression struct {
	op inverted.SetOperator

	unionSetIndex int
	left          *setExpression
	right         *setExpression
}

type invertedSpan = inverted.SpanExpressionProto_Span
type invertedSpans = inverted.SpanExpressionProtoSpans
type spanExpression = inverted.SpanExpressionProto_Node

type spansAndSetIndex struct {
	spans    []invertedSpan
	setIndex int
}

type invertedExprEvaluator struct {
	setExpr *setExpression

	sets []setContainer

	spansIndex []spansAndSetIndex
}

func newInvertedExprEvaluator(expr *spanExpression) *invertedExprEvaluator {
	__antithesis_instrumentation__.Notify(572636)
	eval := &invertedExprEvaluator{}
	eval.setExpr = eval.initSetExpr(expr)
	return eval
}

func (ev *invertedExprEvaluator) initSetExpr(expr *spanExpression) *setExpression {
	__antithesis_instrumentation__.Notify(572637)

	i := len(ev.sets)
	ev.sets = append(ev.sets, nil)
	sx := &setExpression{op: expr.Operator, unionSetIndex: i}
	if len(expr.FactoredUnionSpans) > 0 {
		__antithesis_instrumentation__.Notify(572641)
		ev.spansIndex = append(ev.spansIndex,
			spansAndSetIndex{spans: expr.FactoredUnionSpans, setIndex: i})
	} else {
		__antithesis_instrumentation__.Notify(572642)
	}
	__antithesis_instrumentation__.Notify(572638)
	if expr.Left != nil {
		__antithesis_instrumentation__.Notify(572643)
		sx.left = ev.initSetExpr(expr.Left)
	} else {
		__antithesis_instrumentation__.Notify(572644)
	}
	__antithesis_instrumentation__.Notify(572639)
	if expr.Right != nil {
		__antithesis_instrumentation__.Notify(572645)
		sx.right = ev.initSetExpr(expr.Right)
	} else {
		__antithesis_instrumentation__.Notify(572646)
	}
	__antithesis_instrumentation__.Notify(572640)
	return sx
}

func (ev *invertedExprEvaluator) getSpansAndSetIndex() []spansAndSetIndex {
	__antithesis_instrumentation__.Notify(572647)
	return ev.spansIndex
}

func (ev *invertedExprEvaluator) addIndexRow(setIndex int, keyIndex KeyIndex) {
	__antithesis_instrumentation__.Notify(572648)

	ev.sets[setIndex] = append(ev.sets[setIndex], keyIndex)
}

func (ev *invertedExprEvaluator) evaluate() []KeyIndex {
	__antithesis_instrumentation__.Notify(572649)

	for i, c := range ev.sets {
		__antithesis_instrumentation__.Notify(572651)
		if len(c) == 0 {
			__antithesis_instrumentation__.Notify(572654)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572655)
		}
		__antithesis_instrumentation__.Notify(572652)
		sort.Sort(c)

		set := c[:0]
		for j := range c {
			__antithesis_instrumentation__.Notify(572656)
			if len(set) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(572658)
				return c[j] == set[len(set)-1] == true
			}() == true {
				__antithesis_instrumentation__.Notify(572659)
				continue
			} else {
				__antithesis_instrumentation__.Notify(572660)
			}
			__antithesis_instrumentation__.Notify(572657)
			set = append(set, c[j])
		}
		__antithesis_instrumentation__.Notify(572653)
		ev.sets[i] = set
	}
	__antithesis_instrumentation__.Notify(572650)
	return ev.evaluateSetExpr(ev.setExpr)
}

func (ev *invertedExprEvaluator) evaluateSetExpr(sx *setExpression) setContainer {
	__antithesis_instrumentation__.Notify(572661)
	var left, right setContainer
	if sx.left != nil {
		__antithesis_instrumentation__.Notify(572665)
		left = ev.evaluateSetExpr(sx.left)
	} else {
		__antithesis_instrumentation__.Notify(572666)
	}
	__antithesis_instrumentation__.Notify(572662)
	if sx.right != nil {
		__antithesis_instrumentation__.Notify(572667)
		right = ev.evaluateSetExpr(sx.right)
	} else {
		__antithesis_instrumentation__.Notify(572668)
	}
	__antithesis_instrumentation__.Notify(572663)
	var childrenSet setContainer
	switch sx.op {
	case inverted.SetUnion:
		__antithesis_instrumentation__.Notify(572669)
		childrenSet = unionSetContainers(left, right)
	case inverted.SetIntersection:
		__antithesis_instrumentation__.Notify(572670)
		childrenSet = intersectSetContainers(left, right)
	default:
		__antithesis_instrumentation__.Notify(572671)
	}
	__antithesis_instrumentation__.Notify(572664)
	return unionSetContainers(ev.sets[sx.unionSetIndex], childrenSet)
}

type exprAndSetIndex struct {
	exprIndex int

	setIndex int
}

type exprAndSetIndexSorter []exprAndSetIndex

func (esis exprAndSetIndexSorter) Len() int {
	__antithesis_instrumentation__.Notify(572672)
	return len(esis)
}
func (esis exprAndSetIndexSorter) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(572673)
	esis[i], esis[j] = esis[j], esis[i]
}
func (esis exprAndSetIndexSorter) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(572674)
	return esis[i].exprIndex < esis[j].exprIndex
}

type invertedSpanRoutingInfo struct {
	span invertedSpan

	exprAndSetIndexList []exprAndSetIndex

	exprIndexList []int
}

type invertedSpanRoutingInfosByEndKey []invertedSpanRoutingInfo

func (s invertedSpanRoutingInfosByEndKey) Len() int {
	__antithesis_instrumentation__.Notify(572675)
	return len(s)
}
func (s invertedSpanRoutingInfosByEndKey) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(572676)
	s[i], s[j] = s[j], s[i]
}
func (s invertedSpanRoutingInfosByEndKey) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(572677)
	return bytes.Compare(s[i].span.End, s[j].span.End) < 0
}

type preFilterer interface {
	PreFilter(enc inverted.EncVal, preFilters []interface{}, result []bool) (bool, error)
}

type batchedInvertedExprEvaluator struct {
	filterer preFilterer
	exprs    []*inverted.SpanExpressionProto

	preFilterState []interface{}

	tempPreFilters      []interface{}
	tempPreFilterResult []bool

	exprEvals []*invertedExprEvaluator

	nonInvertedPrefixes []roachpb.Key

	fragmentedSpans []invertedSpanRoutingInfo

	routingIndex int

	routingSpans       []invertedSpanRoutingInfo
	coveringSpans      []invertedSpan
	pendingSpansToSort invertedSpanRoutingInfosByEndKey
}

func (b *batchedInvertedExprEvaluator) fragmentPendingSpans(
	pendingSpans []invertedSpanRoutingInfo, fragmentUntil inverted.EncVal,
) []invertedSpanRoutingInfo {
	__antithesis_instrumentation__.Notify(572678)

	b.pendingSpansToSort = invertedSpanRoutingInfosByEndKey(pendingSpans)
	sort.Sort(&b.pendingSpansToSort)
	for len(pendingSpans) > 0 {
		__antithesis_instrumentation__.Notify(572680)
		if fragmentUntil != nil && func() bool {
			__antithesis_instrumentation__.Notify(572685)
			return bytes.Compare(fragmentUntil, pendingSpans[0].span.Start) <= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(572686)
			break
		} else {
			__antithesis_instrumentation__.Notify(572687)
		}
		__antithesis_instrumentation__.Notify(572681)

		var removeSize int

		var end inverted.EncVal

		var nextStart inverted.EncVal
		if fragmentUntil != nil && func() bool {
			__antithesis_instrumentation__.Notify(572688)
			return bytes.Compare(fragmentUntil, pendingSpans[0].span.End) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(572689)

			removeSize = 0
			end = fragmentUntil
			nextStart = end
		} else {
			__antithesis_instrumentation__.Notify(572690)

			removeSize = b.pendingLenWithSameEnd(pendingSpans)
			end = pendingSpans[0].span.End
			nextStart = end
		}
		__antithesis_instrumentation__.Notify(572682)

		nextSpan := invertedSpanRoutingInfo{
			span: invertedSpan{
				Start: pendingSpans[0].span.Start,
				End:   end,
			},
		}
		for i := 0; i < len(pendingSpans); i++ {
			__antithesis_instrumentation__.Notify(572691)
			if i >= removeSize {
				__antithesis_instrumentation__.Notify(572693)

				pendingSpans[i].span.Start = nextStart
			} else {
				__antithesis_instrumentation__.Notify(572694)
			}
			__antithesis_instrumentation__.Notify(572692)

			nextSpan.exprAndSetIndexList =
				append(nextSpan.exprAndSetIndexList, pendingSpans[i].exprAndSetIndexList...)
		}
		__antithesis_instrumentation__.Notify(572683)

		sort.Sort(exprAndSetIndexSorter(nextSpan.exprAndSetIndexList))
		nextSpan.exprIndexList = make([]int, 0, len(nextSpan.exprAndSetIndexList))
		for i := range nextSpan.exprAndSetIndexList {
			__antithesis_instrumentation__.Notify(572695)
			length := len(nextSpan.exprIndexList)
			exprIndex := nextSpan.exprAndSetIndexList[i].exprIndex
			if length == 0 || func() bool {
				__antithesis_instrumentation__.Notify(572696)
				return nextSpan.exprIndexList[length-1] != exprIndex == true
			}() == true {
				__antithesis_instrumentation__.Notify(572697)
				nextSpan.exprIndexList = append(nextSpan.exprIndexList, exprIndex)
			} else {
				__antithesis_instrumentation__.Notify(572698)
			}
		}
		__antithesis_instrumentation__.Notify(572684)
		b.fragmentedSpans = append(b.fragmentedSpans, nextSpan)
		pendingSpans = pendingSpans[removeSize:]
		if removeSize == 0 {
			__antithesis_instrumentation__.Notify(572699)

			break
		} else {
			__antithesis_instrumentation__.Notify(572700)
		}
	}
	__antithesis_instrumentation__.Notify(572679)
	return pendingSpans
}

func (b *batchedInvertedExprEvaluator) pendingLenWithSameEnd(
	pendingSpans []invertedSpanRoutingInfo,
) int {
	__antithesis_instrumentation__.Notify(572701)
	length := 1
	for i := 1; i < len(pendingSpans); i++ {
		__antithesis_instrumentation__.Notify(572703)
		if !bytes.Equal(pendingSpans[0].span.End, pendingSpans[i].span.End) {
			__antithesis_instrumentation__.Notify(572705)
			break
		} else {
			__antithesis_instrumentation__.Notify(572706)
		}
		__antithesis_instrumentation__.Notify(572704)
		length++
	}
	__antithesis_instrumentation__.Notify(572702)
	return length
}

func (b *batchedInvertedExprEvaluator) init() (invertedSpans, error) {
	if len(b.nonInvertedPrefixes) > 0 && len(b.nonInvertedPrefixes) != len(b.exprs) {
		return nil, errors.AssertionFailedf("length of non-empty nonInvertedPrefixes must equal length of exprs")
	}
	if cap(b.exprEvals) < len(b.exprs) {
		b.exprEvals = make([]*invertedExprEvaluator, len(b.exprs))
	} else {
		b.exprEvals = b.exprEvals[:len(b.exprs)]
	}

	for i, expr := range b.exprs {
		if expr == nil {
			b.exprEvals[i] = nil
			continue
		}
		var prefixKey roachpb.Key
		if len(b.nonInvertedPrefixes) > 0 {
			prefixKey = b.nonInvertedPrefixes[i]
		}
		b.exprEvals[i] = newInvertedExprEvaluator(&expr.Node)
		exprSpans := b.exprEvals[i].getSpansAndSetIndex()
		for _, spans := range exprSpans {
			for _, span := range spans.spans {
				if len(prefixKey) > 0 {

					span = prefixInvertedSpan(prefixKey, span)
				}
				b.routingSpans = append(b.routingSpans,
					invertedSpanRoutingInfo{
						span:                span,
						exprAndSetIndexList: []exprAndSetIndex{{exprIndex: i, setIndex: spans.setIndex}},
					},
				)
			}
		}
	}
	if len(b.routingSpans) == 0 {
		return nil, nil
	}

	sort.Slice(b.routingSpans, func(i, j int) bool {
		cmp := bytes.Compare(b.routingSpans[i].span.Start, b.routingSpans[j].span.Start)
		if cmp == 0 {
			cmp = bytes.Compare(b.routingSpans[i].span.End, b.routingSpans[j].span.End)
		}
		return cmp < 0
	})

	currentCoveringSpan := b.routingSpans[0].span

	pendingSpans := b.routingSpans[:1]

	for i := 1; i < len(b.routingSpans); i++ {
		span := b.routingSpans[i]
		if bytes.Compare(pendingSpans[0].span.Start, span.span.Start) < 0 {
			pendingSpans = b.fragmentPendingSpans(pendingSpans, span.span.Start)
			if bytes.Compare(currentCoveringSpan.End, span.span.Start) < 0 {
				b.coveringSpans = append(b.coveringSpans, currentCoveringSpan)
				currentCoveringSpan = span.span
			} else if bytes.Compare(currentCoveringSpan.End, span.span.End) < 0 {
				currentCoveringSpan.End = span.span.End
			}
		} else if bytes.Compare(currentCoveringSpan.End, span.span.End) < 0 {
			currentCoveringSpan.End = span.span.End
		}

		pendingSpans = pendingSpans[:len(pendingSpans)+1]
	}
	b.fragmentPendingSpans(pendingSpans, nil)
	b.coveringSpans = append(b.coveringSpans, currentCoveringSpan)
	return b.coveringSpans, nil
}

func (b *batchedInvertedExprEvaluator) prepareAddIndexRow(
	enc inverted.EncVal, encFull inverted.EncVal,
) (bool, error) {
	__antithesis_instrumentation__.Notify(572707)
	routingEnc := enc
	if encFull != nil {
		__antithesis_instrumentation__.Notify(572712)
		routingEnc = encFull
	} else {
		__antithesis_instrumentation__.Notify(572713)
	}
	__antithesis_instrumentation__.Notify(572708)

	i := sort.Search(len(b.fragmentedSpans), func(i int) bool {
		__antithesis_instrumentation__.Notify(572714)
		return bytes.Compare(b.fragmentedSpans[i].span.Start, routingEnc) > 0
	})
	__antithesis_instrumentation__.Notify(572709)

	i--
	if i < 0 {
		__antithesis_instrumentation__.Notify(572715)

		return false, errors.AssertionFailedf("unexpectedly negative routing index %d", i)
	} else {
		__antithesis_instrumentation__.Notify(572716)
	}
	__antithesis_instrumentation__.Notify(572710)
	if bytes.Compare(b.fragmentedSpans[i].span.End, routingEnc) <= 0 {
		__antithesis_instrumentation__.Notify(572717)
		return false, errors.AssertionFailedf(
			"unexpectedly the end of the routing span %d is not greater "+
				"than encoded routing value", i,
		)
	} else {
		__antithesis_instrumentation__.Notify(572718)
	}
	__antithesis_instrumentation__.Notify(572711)
	b.routingIndex = i
	return b.prefilter(enc)
}

func (b *batchedInvertedExprEvaluator) prefilter(enc inverted.EncVal) (bool, error) {
	__antithesis_instrumentation__.Notify(572719)
	if b.filterer != nil {
		__antithesis_instrumentation__.Notify(572721)
		exprIndexList := b.fragmentedSpans[b.routingIndex].exprIndexList
		if len(exprIndexList) > cap(b.tempPreFilters) {
			__antithesis_instrumentation__.Notify(572724)
			b.tempPreFilters = make([]interface{}, len(exprIndexList))
			b.tempPreFilterResult = make([]bool, len(exprIndexList))
		} else {
			__antithesis_instrumentation__.Notify(572725)
			b.tempPreFilters = b.tempPreFilters[:len(exprIndexList)]
			b.tempPreFilterResult = b.tempPreFilterResult[:len(exprIndexList)]
		}
		__antithesis_instrumentation__.Notify(572722)
		for j := range exprIndexList {
			__antithesis_instrumentation__.Notify(572726)
			b.tempPreFilters[j] = b.preFilterState[exprIndexList[j]]
		}
		__antithesis_instrumentation__.Notify(572723)
		return b.filterer.PreFilter(enc, b.tempPreFilters, b.tempPreFilterResult)
	} else {
		__antithesis_instrumentation__.Notify(572727)
	}
	__antithesis_instrumentation__.Notify(572720)
	return true, nil
}

func (b *batchedInvertedExprEvaluator) addIndexRow(keyIndex KeyIndex) error {
	__antithesis_instrumentation__.Notify(572728)
	i := b.routingIndex
	if b.filterer != nil {
		__antithesis_instrumentation__.Notify(572730)
		exprIndexes := b.fragmentedSpans[i].exprIndexList
		exprSetIndexes := b.fragmentedSpans[i].exprAndSetIndexList
		if len(exprIndexes) != len(b.tempPreFilterResult) {
			__antithesis_instrumentation__.Notify(572732)
			return errors.Errorf("non-matching lengths of tempPreFilterResult and exprIndexes")
		} else {
			__antithesis_instrumentation__.Notify(572733)
		}
		__antithesis_instrumentation__.Notify(572731)

		j := 0
		for k := range exprSetIndexes {
			__antithesis_instrumentation__.Notify(572734)
			elem := exprSetIndexes[k]
			if elem.exprIndex > exprIndexes[j] {
				__antithesis_instrumentation__.Notify(572736)
				j++
				if exprIndexes[j] != elem.exprIndex {
					__antithesis_instrumentation__.Notify(572737)
					return errors.Errorf("non-matching expr indexes")
				} else {
					__antithesis_instrumentation__.Notify(572738)
				}
			} else {
				__antithesis_instrumentation__.Notify(572739)
			}
			__antithesis_instrumentation__.Notify(572735)
			if b.tempPreFilterResult[j] {
				__antithesis_instrumentation__.Notify(572740)
				b.exprEvals[elem.exprIndex].addIndexRow(elem.setIndex, keyIndex)
			} else {
				__antithesis_instrumentation__.Notify(572741)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(572742)
		for _, elem := range b.fragmentedSpans[i].exprAndSetIndexList {
			__antithesis_instrumentation__.Notify(572743)
			b.exprEvals[elem.exprIndex].addIndexRow(elem.setIndex, keyIndex)
		}
	}
	__antithesis_instrumentation__.Notify(572729)
	return nil
}

func (b *batchedInvertedExprEvaluator) evaluate() [][]KeyIndex {
	__antithesis_instrumentation__.Notify(572744)
	result := make([][]KeyIndex, len(b.exprs))
	for i := range b.exprEvals {
		__antithesis_instrumentation__.Notify(572746)
		if b.exprEvals[i] == nil {
			__antithesis_instrumentation__.Notify(572748)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572749)
		}
		__antithesis_instrumentation__.Notify(572747)
		result[i] = b.exprEvals[i].evaluate()
	}
	__antithesis_instrumentation__.Notify(572745)
	return result
}

func (b *batchedInvertedExprEvaluator) reset() {
	__antithesis_instrumentation__.Notify(572750)
	b.exprs = b.exprs[:0]
	b.preFilterState = b.preFilterState[:0]
	b.exprEvals = b.exprEvals[:0]
	b.fragmentedSpans = b.fragmentedSpans[:0]
	b.routingSpans = b.routingSpans[:0]
	b.coveringSpans = b.coveringSpans[:0]
	b.nonInvertedPrefixes = b.nonInvertedPrefixes[:0]
}

func prefixInvertedSpan(prefix roachpb.Key, span invertedSpan) invertedSpan {
	__antithesis_instrumentation__.Notify(572751)
	newSpan := invertedSpan{
		Start: make(roachpb.Key, 0, len(prefix)+len(span.Start)),
		End:   make(roachpb.Key, 0, len(prefix)+len(span.End)),
	}
	newSpan.Start = append(newSpan.Start, prefix...)
	newSpan.Start = append(newSpan.Start, span.Start...)
	newSpan.End = append(newSpan.End, prefix...)
	newSpan.End = append(newSpan.End, span.End...)
	return newSpan
}
