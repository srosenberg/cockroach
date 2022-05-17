package inverted

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/keysbase"
	"github.com/cockroachdb/cockroach/pkg/util/treeprinter"
	"github.com/cockroachdb/errors"
)

type EncVal []byte

type Span struct {
	Start, End EncVal
}

func MakeSingleValSpan(val EncVal) Span {
	__antithesis_instrumentation__.Notify(498033)
	end := EncVal(keysbase.PrefixEnd(val))
	return Span{Start: val, End: end}
}

func (s Span) IsSingleVal() bool {
	__antithesis_instrumentation__.Notify(498034)
	return bytes.Equal(keysbase.PrefixEnd(s.Start), s.End)
}

func (s Span) Equals(other Span) bool {
	__antithesis_instrumentation__.Notify(498035)
	if !bytes.Equal(s.Start, other.Start) {
		__antithesis_instrumentation__.Notify(498037)
		return false
	} else {
		__antithesis_instrumentation__.Notify(498038)
	}
	__antithesis_instrumentation__.Notify(498036)
	return bytes.Equal(s.End, other.End)
}

type Spans []Span

func (s Span) ContainsKey(key EncVal) bool {
	__antithesis_instrumentation__.Notify(498039)
	return bytes.Compare(key, s.Start) >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(498040)
		return bytes.Compare(key, s.End) < 0 == true
	}() == true
}

func (is Spans) Equals(other Spans) bool {
	__antithesis_instrumentation__.Notify(498041)
	if len(is) != len(other) {
		__antithesis_instrumentation__.Notify(498044)
		return false
	} else {
		__antithesis_instrumentation__.Notify(498045)
	}
	__antithesis_instrumentation__.Notify(498042)
	for i := range is {
		__antithesis_instrumentation__.Notify(498046)
		if !is[i].Equals(other[i]) {
			__antithesis_instrumentation__.Notify(498047)
			return false
		} else {
			__antithesis_instrumentation__.Notify(498048)
		}
	}
	__antithesis_instrumentation__.Notify(498043)
	return true
}

func (is Spans) Format(tp treeprinter.Node, label string) {
	__antithesis_instrumentation__.Notify(498049)
	if len(is) == 0 {
		__antithesis_instrumentation__.Notify(498052)
		tp.Childf("%s: empty", label)
		return
	} else {
		__antithesis_instrumentation__.Notify(498053)
	}
	__antithesis_instrumentation__.Notify(498050)
	if len(is) == 1 {
		__antithesis_instrumentation__.Notify(498054)
		tp.Childf("%s: %s", label, formatSpan(is[0]))
		return
	} else {
		__antithesis_instrumentation__.Notify(498055)
	}
	__antithesis_instrumentation__.Notify(498051)
	n := tp.Child(label)
	for i := 0; i < len(is); i++ {
		__antithesis_instrumentation__.Notify(498056)
		n.Child(formatSpan(is[i]))
	}
}

func formatSpan(span Span) string {
	__antithesis_instrumentation__.Notify(498057)
	end := span.End
	spanEndOpenOrClosed := ')'
	if span.IsSingleVal() {
		__antithesis_instrumentation__.Notify(498059)
		end = span.Start
		spanEndOpenOrClosed = ']'
	} else {
		__antithesis_instrumentation__.Notify(498060)
	}
	__antithesis_instrumentation__.Notify(498058)
	return fmt.Sprintf("[%s, %s%c", strconv.Quote(string(span.Start)),
		strconv.Quote(string(end)), spanEndOpenOrClosed)
}

func (is Spans) Len() int { __antithesis_instrumentation__.Notify(498061); return len(is) }

func (is Spans) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(498062)
	return bytes.Compare(is[i].Start, is[j].Start) < 0
}

func (is Spans) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(498063)
	is[i], is[j] = is[j], is[i]
}

func (is Spans) Start(i int) []byte {
	__antithesis_instrumentation__.Notify(498064)
	return is[i].Start
}

func (is Spans) End(i int) []byte {
	__antithesis_instrumentation__.Notify(498065)
	return is[i].End
}

type Expression interface {
	IsTight() bool

	SetNotTight()

	Copy() Expression
}

type SpanExpression struct {
	Tight bool

	Unique bool

	SpansToRead Spans

	FactoredUnionSpans Spans

	Operator SetOperator
	Left     Expression
	Right    Expression
}

var _ Expression = (*SpanExpression)(nil)

func (s *SpanExpression) IsTight() bool {
	__antithesis_instrumentation__.Notify(498066)
	return s.Tight
}

func (s *SpanExpression) SetNotTight() {
	__antithesis_instrumentation__.Notify(498067)
	s.Tight = false
}

func (s *SpanExpression) Copy() Expression {
	__antithesis_instrumentation__.Notify(498068)
	res := &SpanExpression{
		Tight:              s.Tight,
		Unique:             s.Unique,
		SpansToRead:        s.SpansToRead,
		FactoredUnionSpans: s.FactoredUnionSpans,
		Operator:           s.Operator,
	}
	if s.Left != nil {
		__antithesis_instrumentation__.Notify(498071)
		res.Left = s.Left.Copy()
	} else {
		__antithesis_instrumentation__.Notify(498072)
	}
	__antithesis_instrumentation__.Notify(498069)
	if s.Right != nil {
		__antithesis_instrumentation__.Notify(498073)
		res.Right = s.Right.Copy()
	} else {
		__antithesis_instrumentation__.Notify(498074)
	}
	__antithesis_instrumentation__.Notify(498070)
	return res
}

func (s *SpanExpression) String() string {
	__antithesis_instrumentation__.Notify(498075)
	tp := treeprinter.New()
	n := tp.Child("span expression")
	s.Format(n, true)
	return tp.String()
}

func (s *SpanExpression) Format(tp treeprinter.Node, includeSpansToRead bool) {
	__antithesis_instrumentation__.Notify(498076)
	tp.Childf("tight: %t, unique: %t", s.Tight, s.Unique)
	if includeSpansToRead {
		__antithesis_instrumentation__.Notify(498080)
		s.SpansToRead.Format(tp, "to read")
	} else {
		__antithesis_instrumentation__.Notify(498081)
	}
	__antithesis_instrumentation__.Notify(498077)
	s.FactoredUnionSpans.Format(tp, "union spans")
	if s.Operator == None {
		__antithesis_instrumentation__.Notify(498082)
		return
	} else {
		__antithesis_instrumentation__.Notify(498083)
	}
	__antithesis_instrumentation__.Notify(498078)
	switch s.Operator {
	case SetUnion:
		__antithesis_instrumentation__.Notify(498084)
		tp = tp.Child("UNION")
	case SetIntersection:
		__antithesis_instrumentation__.Notify(498085)
		tp = tp.Child("INTERSECTION")
	default:
		__antithesis_instrumentation__.Notify(498086)
	}
	__antithesis_instrumentation__.Notify(498079)
	formatExpression(tp, s.Left, includeSpansToRead)
	formatExpression(tp, s.Right, includeSpansToRead)
}

func formatExpression(tp treeprinter.Node, expr Expression, includeSpansToRead bool) {
	__antithesis_instrumentation__.Notify(498087)
	switch e := expr.(type) {
	case *SpanExpression:
		__antithesis_instrumentation__.Notify(498088)
		n := tp.Child("span expression")
		e.Format(n, includeSpansToRead)
	default:
		__antithesis_instrumentation__.Notify(498089)
		tp.Child(fmt.Sprintf("%v", e))
	}
}

func (s *SpanExpression) ToProto() *SpanExpressionProto {
	__antithesis_instrumentation__.Notify(498090)
	if s == nil {
		__antithesis_instrumentation__.Notify(498092)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(498093)
	}
	__antithesis_instrumentation__.Notify(498091)
	proto := &SpanExpressionProto{
		SpansToRead: getProtoSpans(s.SpansToRead),
		Node:        *s.getProtoNode(),
	}
	return proto
}

func getProtoSpans(spans []Span) []SpanExpressionProto_Span {
	__antithesis_instrumentation__.Notify(498094)
	out := make([]SpanExpressionProto_Span, len(spans))
	for i := range spans {
		__antithesis_instrumentation__.Notify(498096)
		out[i] = SpanExpressionProto_Span{
			Start: spans[i].Start,
			End:   spans[i].End,
		}
	}
	__antithesis_instrumentation__.Notify(498095)
	return out
}

func (s *SpanExpression) getProtoNode() *SpanExpressionProto_Node {
	__antithesis_instrumentation__.Notify(498097)
	node := &SpanExpressionProto_Node{
		FactoredUnionSpans: getProtoSpans(s.FactoredUnionSpans),
		Operator:           s.Operator,
	}
	if node.Operator != None {
		__antithesis_instrumentation__.Notify(498099)
		node.Left = s.Left.(*SpanExpression).getProtoNode()
		node.Right = s.Right.(*SpanExpression).getProtoNode()
	} else {
		__antithesis_instrumentation__.Notify(498100)
	}
	__antithesis_instrumentation__.Notify(498098)
	return node
}

type NonInvertedColExpression struct{}

var _ Expression = NonInvertedColExpression{}

func (n NonInvertedColExpression) IsTight() bool {
	__antithesis_instrumentation__.Notify(498101)
	return false
}

func (n NonInvertedColExpression) SetNotTight() { __antithesis_instrumentation__.Notify(498102) }

func (n NonInvertedColExpression) Copy() Expression {
	__antithesis_instrumentation__.Notify(498103)
	return NonInvertedColExpression{}
}

type SpanExpressionProtoSpans []SpanExpressionProto_Span

func (s SpanExpressionProtoSpans) Len() int {
	__antithesis_instrumentation__.Notify(498104)
	return len(s)
}

func (s SpanExpressionProtoSpans) Start(i int) []byte {
	__antithesis_instrumentation__.Notify(498105)
	return s[i].Start
}

func (s SpanExpressionProtoSpans) End(i int) []byte {
	__antithesis_instrumentation__.Notify(498106)
	return s[i].End
}

func ExprForSpan(span Span, tight bool) *SpanExpression {
	__antithesis_instrumentation__.Notify(498107)
	return &SpanExpression{
		Tight:              tight,
		SpansToRead:        []Span{span},
		FactoredUnionSpans: []Span{span},
	}
}

func (s *SpanExpression) ContainsKeys(keys [][]byte) (bool, error) {
	__antithesis_instrumentation__.Notify(498108)
	if s.Operator == None && func() bool {
		__antithesis_instrumentation__.Notify(498115)
		return len(s.FactoredUnionSpans) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(498116)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(498117)
	}
	__antithesis_instrumentation__.Notify(498109)

	if len(s.FactoredUnionSpans) > 0 {
		__antithesis_instrumentation__.Notify(498118)
		for _, span := range s.FactoredUnionSpans {
			__antithesis_instrumentation__.Notify(498119)
			for _, key := range keys {
				__antithesis_instrumentation__.Notify(498120)
				if span.ContainsKey(key) {
					__antithesis_instrumentation__.Notify(498121)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(498122)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(498123)
	}
	__antithesis_instrumentation__.Notify(498110)

	if s.Operator == None {
		__antithesis_instrumentation__.Notify(498124)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(498125)
	}
	__antithesis_instrumentation__.Notify(498111)

	leftRes, err := s.Left.(*SpanExpression).ContainsKeys(keys)
	if err != nil {
		__antithesis_instrumentation__.Notify(498126)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(498127)
	}
	__antithesis_instrumentation__.Notify(498112)
	if leftRes && func() bool {
		__antithesis_instrumentation__.Notify(498128)
		return s.Operator == SetUnion == true
	}() == true {
		__antithesis_instrumentation__.Notify(498129)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(498130)
	}
	__antithesis_instrumentation__.Notify(498113)

	rightRes, err := s.Right.(*SpanExpression).ContainsKeys(keys)
	if err != nil {
		__antithesis_instrumentation__.Notify(498131)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(498132)
	}
	__antithesis_instrumentation__.Notify(498114)
	switch s.Operator {
	case SetIntersection:
		__antithesis_instrumentation__.Notify(498133)
		return leftRes && func() bool {
			__antithesis_instrumentation__.Notify(498136)
			return rightRes == true
		}() == true, nil
	case SetUnion:
		__antithesis_instrumentation__.Notify(498134)
		return leftRes || func() bool {
			__antithesis_instrumentation__.Notify(498137)
			return rightRes == true
		}() == true, nil
	default:
		__antithesis_instrumentation__.Notify(498135)
		return false, errors.AssertionFailedf("invalid operator %v", s.Operator)
	}
}

func And(left, right Expression) Expression {
	__antithesis_instrumentation__.Notify(498138)
	switch l := left.(type) {
	case *SpanExpression:
		__antithesis_instrumentation__.Notify(498139)
		switch r := right.(type) {
		case *SpanExpression:
			__antithesis_instrumentation__.Notify(498142)
			return intersectSpanExpressions(l, r)
		case NonInvertedColExpression:
			__antithesis_instrumentation__.Notify(498143)
			left.SetNotTight()
			return left
		default:
			__antithesis_instrumentation__.Notify(498144)
			return opSpanExpressionAndDefault(l, right, SetIntersection)
		}
	case NonInvertedColExpression:
		__antithesis_instrumentation__.Notify(498140)
		right.SetNotTight()
		return right
	default:
		__antithesis_instrumentation__.Notify(498141)
		switch r := right.(type) {
		case *SpanExpression:
			__antithesis_instrumentation__.Notify(498145)
			return opSpanExpressionAndDefault(r, left, SetIntersection)
		case NonInvertedColExpression:
			__antithesis_instrumentation__.Notify(498146)
			left.SetNotTight()
			return left
		default:
			__antithesis_instrumentation__.Notify(498147)
			return &SpanExpression{
				Tight: left.IsTight() && func() bool {
					__antithesis_instrumentation__.Notify(498148)
					return right.IsTight() == true
				}() == true,
				Operator: SetIntersection,
				Left:     left,
				Right:    right,
			}
		}
	}
}

func Or(left, right Expression) Expression {
	__antithesis_instrumentation__.Notify(498149)
	switch l := left.(type) {
	case *SpanExpression:
		__antithesis_instrumentation__.Notify(498150)
		switch r := right.(type) {
		case *SpanExpression:
			__antithesis_instrumentation__.Notify(498153)
			return unionSpanExpressions(l, r)
		case NonInvertedColExpression:
			__antithesis_instrumentation__.Notify(498154)
			return r
		default:
			__antithesis_instrumentation__.Notify(498155)
			return opSpanExpressionAndDefault(l, right, SetUnion)
		}
	case NonInvertedColExpression:
		__antithesis_instrumentation__.Notify(498151)
		return left
	default:
		__antithesis_instrumentation__.Notify(498152)
		switch r := right.(type) {
		case *SpanExpression:
			__antithesis_instrumentation__.Notify(498156)
			return opSpanExpressionAndDefault(r, left, SetUnion)
		case NonInvertedColExpression:
			__antithesis_instrumentation__.Notify(498157)
			return right
		default:
			__antithesis_instrumentation__.Notify(498158)
			return &SpanExpression{
				Tight: left.IsTight() && func() bool {
					__antithesis_instrumentation__.Notify(498159)
					return right.IsTight() == true
				}() == true,
				Operator: SetUnion,
				Left:     left,
				Right:    right,
			}
		}
	}
}

func opSpanExpressionAndDefault(
	left *SpanExpression, right Expression, op SetOperator,
) *SpanExpression {
	__antithesis_instrumentation__.Notify(498160)
	expr := &SpanExpression{
		Tight: left.IsTight() && func() bool {
			__antithesis_instrumentation__.Notify(498162)
			return right.IsTight() == true
		}() == true,

		SpansToRead: left.SpansToRead,
		Operator:    op,
		Left:        left,
		Right:       right,
	}
	if op == SetUnion {
		__antithesis_instrumentation__.Notify(498163)

		expr.FactoredUnionSpans = left.FactoredUnionSpans
		left.FactoredUnionSpans = nil
	} else {
		__antithesis_instrumentation__.Notify(498164)
	}
	__antithesis_instrumentation__.Notify(498161)

	return expr
}

func intersectSpanExpressions(left, right *SpanExpression) *SpanExpression {
	__antithesis_instrumentation__.Notify(498165)
	expr := &SpanExpression{
		Tight: left.Tight && func() bool {
			__antithesis_instrumentation__.Notify(498167)
			return right.Tight == true
		}() == true,
		Unique: left.Unique && func() bool {
			__antithesis_instrumentation__.Notify(498168)
			return right.Unique == true
		}() == true,

		SpansToRead:        unionSpans(left.SpansToRead, right.SpansToRead),
		FactoredUnionSpans: intersectSpans(left.FactoredUnionSpans, right.FactoredUnionSpans),
		Operator:           SetIntersection,
		Left:               left,
		Right:              right,
	}
	if expr.FactoredUnionSpans != nil {
		__antithesis_instrumentation__.Notify(498169)
		left.FactoredUnionSpans = subtractSpans(left.FactoredUnionSpans, expr.FactoredUnionSpans)
		right.FactoredUnionSpans = subtractSpans(right.FactoredUnionSpans, expr.FactoredUnionSpans)
	} else {
		__antithesis_instrumentation__.Notify(498170)
	}
	__antithesis_instrumentation__.Notify(498166)
	tryPruneChildren(expr, SetIntersection)
	return expr
}

func unionSpanExpressions(left, right *SpanExpression) *SpanExpression {
	__antithesis_instrumentation__.Notify(498171)
	expr := &SpanExpression{
		Tight: left.Tight && func() bool {
			__antithesis_instrumentation__.Notify(498172)
			return right.Tight == true
		}() == true,
		SpansToRead:        unionSpans(left.SpansToRead, right.SpansToRead),
		FactoredUnionSpans: unionSpans(left.FactoredUnionSpans, right.FactoredUnionSpans),
		Operator:           SetUnion,
		Left:               left,
		Right:              right,
	}
	left.FactoredUnionSpans = nil
	right.FactoredUnionSpans = nil
	tryPruneChildren(expr, SetUnion)
	return expr
}

func tryPruneChildren(expr *SpanExpression, op SetOperator) {
	__antithesis_instrumentation__.Notify(498173)
	isEmptyExpr := func(e *SpanExpression) bool {
		__antithesis_instrumentation__.Notify(498180)
		return len(e.FactoredUnionSpans) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(498181)
			return e.Left == nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(498182)
			return e.Right == nil == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(498174)
	if isEmptyExpr(expr.Left.(*SpanExpression)) {
		__antithesis_instrumentation__.Notify(498183)
		expr.Left = nil
	} else {
		__antithesis_instrumentation__.Notify(498184)
	}
	__antithesis_instrumentation__.Notify(498175)
	if isEmptyExpr(expr.Right.(*SpanExpression)) {
		__antithesis_instrumentation__.Notify(498185)
		expr.Right = nil
	} else {
		__antithesis_instrumentation__.Notify(498186)
	}
	__antithesis_instrumentation__.Notify(498176)

	promoteChild := func(child *SpanExpression) {
		__antithesis_instrumentation__.Notify(498187)

		expr.Operator = child.Operator
		expr.Left = child.Left
		expr.Right = child.Right

		if child.FactoredUnionSpans != nil {
			__antithesis_instrumentation__.Notify(498188)
			expr.SpansToRead = expr.FactoredUnionSpans
			if expr.Left != nil {
				__antithesis_instrumentation__.Notify(498190)
				expr.SpansToRead = unionSpans(expr.SpansToRead, expr.Left.(*SpanExpression).SpansToRead)
			} else {
				__antithesis_instrumentation__.Notify(498191)
			}
			__antithesis_instrumentation__.Notify(498189)
			if expr.Right != nil {
				__antithesis_instrumentation__.Notify(498192)
				expr.SpansToRead = unionSpans(expr.SpansToRead, expr.Right.(*SpanExpression).SpansToRead)
			} else {
				__antithesis_instrumentation__.Notify(498193)
			}
		} else {
			__antithesis_instrumentation__.Notify(498194)
		}
	}
	__antithesis_instrumentation__.Notify(498177)
	promoteLeft := expr.Left != nil && func() bool {
		__antithesis_instrumentation__.Notify(498195)
		return expr.Right == nil == true
	}() == true
	promoteRight := expr.Left == nil && func() bool {
		__antithesis_instrumentation__.Notify(498196)
		return expr.Right != nil == true
	}() == true
	if promoteLeft {
		__antithesis_instrumentation__.Notify(498197)
		promoteChild(expr.Left.(*SpanExpression))
	} else {
		__antithesis_instrumentation__.Notify(498198)
	}
	__antithesis_instrumentation__.Notify(498178)
	if promoteRight {
		__antithesis_instrumentation__.Notify(498199)
		promoteChild(expr.Right.(*SpanExpression))
	} else {
		__antithesis_instrumentation__.Notify(498200)
	}
	__antithesis_instrumentation__.Notify(498179)
	if expr.Left == nil && func() bool {
		__antithesis_instrumentation__.Notify(498201)
		return expr.Right == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(498202)
		expr.Operator = None
	} else {
		__antithesis_instrumentation__.Notify(498203)
	}
}

func unionSpans(left []Span, right []Span) []Span {
	__antithesis_instrumentation__.Notify(498204)
	if len(left) == 0 {
		__antithesis_instrumentation__.Notify(498210)
		return right
	} else {
		__antithesis_instrumentation__.Notify(498211)
	}
	__antithesis_instrumentation__.Notify(498205)
	if len(right) == 0 {
		__antithesis_instrumentation__.Notify(498212)
		return left
	} else {
		__antithesis_instrumentation__.Notify(498213)
	}
	__antithesis_instrumentation__.Notify(498206)

	var spans []Span

	var mergeSpan Span

	var i, j int

	swapLeftRight := func() {
		__antithesis_instrumentation__.Notify(498214)
		i, j = j, i
		left, right = right, left
	}
	__antithesis_instrumentation__.Notify(498207)

	makeMergeSpan := func() {
		__antithesis_instrumentation__.Notify(498215)
		if i >= len(left) || func() bool {
			__antithesis_instrumentation__.Notify(498217)
			return (j < len(right) && func() bool {
				__antithesis_instrumentation__.Notify(498218)
				return bytes.Compare(left[i].Start, right[j].Start) > 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(498219)
			swapLeftRight()
		} else {
			__antithesis_instrumentation__.Notify(498220)
		}
		__antithesis_instrumentation__.Notify(498216)
		mergeSpan = left[i]
		i++
	}
	__antithesis_instrumentation__.Notify(498208)
	makeMergeSpan()

	for j < len(right) {
		__antithesis_instrumentation__.Notify(498221)
		cmpEndStart := cmpExcEndWithIncStart(mergeSpan, right[j])
		if cmpEndStart >= 0 {
			__antithesis_instrumentation__.Notify(498223)
			if extendSpanEnd(&mergeSpan, right[j], cmpEndStart) {
				__antithesis_instrumentation__.Notify(498225)

				j++
				swapLeftRight()
			} else {
				__antithesis_instrumentation__.Notify(498226)
				j++
			}
			__antithesis_instrumentation__.Notify(498224)
			continue
		} else {
			__antithesis_instrumentation__.Notify(498227)
		}
		__antithesis_instrumentation__.Notify(498222)

		spans = append(spans, mergeSpan)
		makeMergeSpan()
	}
	__antithesis_instrumentation__.Notify(498209)
	spans = append(spans, mergeSpan)
	spans = append(spans, left[i:]...)
	return spans
}

func intersectSpans(left []Span, right []Span) []Span {
	__antithesis_instrumentation__.Notify(498228)
	if len(left) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(498233)
		return len(right) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(498234)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(498235)
	}
	__antithesis_instrumentation__.Notify(498229)

	var spans []Span

	var i, j int

	var mergeSpan Span
	var mergeSpanInitialized bool
	swapLeftRight := func() {
		__antithesis_instrumentation__.Notify(498236)
		i, j = j, i
		left, right = right, left
	}
	__antithesis_instrumentation__.Notify(498230)

	makeMergeSpan := func() {
		__antithesis_instrumentation__.Notify(498237)
		if bytes.Compare(left[i].Start, right[j].Start) > 0 {
			__antithesis_instrumentation__.Notify(498239)
			swapLeftRight()
		} else {
			__antithesis_instrumentation__.Notify(498240)
		}
		__antithesis_instrumentation__.Notify(498238)
		mergeSpan = left[i]
		mergeSpanInitialized = true
	}
	__antithesis_instrumentation__.Notify(498231)

	for i < len(left) && func() bool {
		__antithesis_instrumentation__.Notify(498241)
		return j < len(right) == true
	}() == true {
		__antithesis_instrumentation__.Notify(498242)
		if !mergeSpanInitialized {
			__antithesis_instrumentation__.Notify(498244)
			makeMergeSpan()
		} else {
			__antithesis_instrumentation__.Notify(498245)
		}
		__antithesis_instrumentation__.Notify(498243)
		cmpEndStart := cmpExcEndWithIncStart(mergeSpan, right[j])
		if cmpEndStart > 0 {
			__antithesis_instrumentation__.Notify(498246)

			mergeSpan.Start = right[j].Start
			mergeSpanEnd := mergeSpan.End
			cmpEnds := cmpEnds(mergeSpan, right[j])
			if cmpEnds > 0 {
				__antithesis_instrumentation__.Notify(498248)

				mergeSpan.End = right[j].End
			} else {
				__antithesis_instrumentation__.Notify(498249)
			}
			__antithesis_instrumentation__.Notify(498247)

			spans = append(spans, mergeSpan)

			if cmpEnds < 0 {
				__antithesis_instrumentation__.Notify(498250)

				i++
				mergeSpan.Start = mergeSpan.End
				mergeSpan.End = right[j].End
				swapLeftRight()
			} else {
				__antithesis_instrumentation__.Notify(498251)
				if cmpEnds == 0 {
					__antithesis_instrumentation__.Notify(498252)

					i++
					j++
					mergeSpanInitialized = false
				} else {
					__antithesis_instrumentation__.Notify(498253)

					j++
					mergeSpan.Start = mergeSpan.End
					mergeSpan.End = mergeSpanEnd
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(498254)

			i++
			mergeSpanInitialized = false
		}
	}
	__antithesis_instrumentation__.Notify(498232)
	return spans
}

func subtractSpans(left []Span, right []Span) []Span {
	__antithesis_instrumentation__.Notify(498255)
	if len(right) == 0 {
		__antithesis_instrumentation__.Notify(498259)
		return left
	} else {
		__antithesis_instrumentation__.Notify(498260)
	}
	__antithesis_instrumentation__.Notify(498256)

	var out []Span

	var mergeSpan Span
	var mergeSpanInitialized bool

	var i, j int
	for j < len(right) {
		__antithesis_instrumentation__.Notify(498261)
		if !mergeSpanInitialized {
			__antithesis_instrumentation__.Notify(498263)
			mergeSpan = left[i]
			mergeSpanInitialized = true
		} else {
			__antithesis_instrumentation__.Notify(498264)
		}
		__antithesis_instrumentation__.Notify(498262)
		cmpEndStart := cmpExcEndWithIncStart(mergeSpan, right[j])
		if cmpEndStart > 0 {
			__antithesis_instrumentation__.Notify(498265)

			cmpStart := bytes.Compare(mergeSpan.Start, right[j].Start)
			if cmpStart < 0 {
				__antithesis_instrumentation__.Notify(498268)

				out = append(out, Span{Start: mergeSpan.Start, End: right[j].Start})
				mergeSpan.Start = right[j].Start
			} else {
				__antithesis_instrumentation__.Notify(498269)
			}
			__antithesis_instrumentation__.Notify(498266)

			cmpEnd := cmpEnds(mergeSpan, right[j])
			if cmpEnd == 0 {
				__antithesis_instrumentation__.Notify(498270)

				i++
				j++
				mergeSpanInitialized = false
				continue
			} else {
				__antithesis_instrumentation__.Notify(498271)
			}
			__antithesis_instrumentation__.Notify(498267)

			mergeSpan.Start = right[j].End
			j++
		} else {
			__antithesis_instrumentation__.Notify(498272)

			out = append(out, mergeSpan)
			i++
			mergeSpanInitialized = false
		}
	}
	__antithesis_instrumentation__.Notify(498257)
	if mergeSpanInitialized {
		__antithesis_instrumentation__.Notify(498273)
		out = append(out, mergeSpan)
		i++
	} else {
		__antithesis_instrumentation__.Notify(498274)
	}
	__antithesis_instrumentation__.Notify(498258)
	out = append(out, left[i:]...)
	return out
}

func cmpExcEndWithIncStart(left, right Span) int {
	__antithesis_instrumentation__.Notify(498275)
	return bytes.Compare(left.End, right.Start)
}

func extendSpanEnd(left *Span, right Span, cmpExcEndIncStart int) bool {
	__antithesis_instrumentation__.Notify(498276)
	if cmpExcEndIncStart == 0 {
		__antithesis_instrumentation__.Notify(498279)

		left.End = right.End
		return true
	} else {
		__antithesis_instrumentation__.Notify(498280)
	}
	__antithesis_instrumentation__.Notify(498277)

	if bytes.Compare(left.End, right.End) < 0 {
		__antithesis_instrumentation__.Notify(498281)
		left.End = right.End
		return true
	} else {
		__antithesis_instrumentation__.Notify(498282)
	}
	__antithesis_instrumentation__.Notify(498278)
	return false
}

func cmpEnds(left, right Span) int {
	__antithesis_instrumentation__.Notify(498283)
	return bytes.Compare(left.End, right.End)
}
