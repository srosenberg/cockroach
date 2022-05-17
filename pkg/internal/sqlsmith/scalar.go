package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treewindow"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var (
	scalars = []scalarExprWeight{
		{10, scalarNoContext(makeAnd)},
		{1, scalarNoContext(makeCaseExpr)},
		{1, scalarNoContext(makeCoalesceExpr)},
		{50, scalarNoContext(makeColRef)},
		{10, scalarNoContext(makeBinOp)},
		{2, scalarNoContext(makeScalarSubquery)},
		{2, scalarNoContext(makeExists)},
		{2, scalarNoContext(makeIn)},
		{2, scalarNoContext(makeStringComparison)},
		{5, scalarNoContext(makeAnd)},
		{5, scalarNoContext(makeOr)},
		{5, scalarNoContext(makeNot)},
		{10, makeFunc},
		{10, func(s *Smither, ctx Context, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
			__antithesis_instrumentation__.Notify(69424)
			return makeConstExpr(s, typ, refs), true
		}},
	}

	bools = []scalarExprWeight{
		{1, scalarNoContext(makeColRef)},
		{1, scalarNoContext(makeAnd)},
		{1, scalarNoContext(makeOr)},
		{1, scalarNoContext(makeNot)},
		{1, scalarNoContext(makeCompareOp)},
		{1, scalarNoContext(makeIn)},
		{1, scalarNoContext(makeStringComparison)},
		{1, func(s *Smither, ctx Context, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
			__antithesis_instrumentation__.Notify(69425)
			return makeScalar(s, typ, refs), true
		}},
		{1, scalarNoContext(makeExists)},
		{1, makeFunc},
	}
)

func scalarNoContext(fn func(*Smither, *types.T, colRefs) (tree.TypedExpr, bool)) scalarExpr {
	__antithesis_instrumentation__.Notify(69426)
	return func(s *Smither, ctx Context, t *types.T, refs colRefs) (tree.TypedExpr, bool) {
		__antithesis_instrumentation__.Notify(69427)
		return fn(s, t, refs)
	}
}

func makeScalar(s *Smither, typ *types.T, refs colRefs) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69428)
	return makeScalarContext(s, emptyCtx, typ, refs)
}

func makeScalarContext(s *Smither, ctx Context, typ *types.T, refs colRefs) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69429)
	return makeScalarSample(s.scalarExprSampler, s, ctx, typ, refs)
}

func makeBoolExpr(s *Smither, refs colRefs) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69430)
	return makeBoolExprContext(s, emptyCtx, refs)
}

func makeBoolExprWithPlaceholders(s *Smither, refs colRefs) (tree.Expr, []interface{}) {
	__antithesis_instrumentation__.Notify(69431)
	expr := makeBoolExprContext(s, emptyCtx, refs)

	visitor := replaceDatumPlaceholderVisitor{}
	exprFmt := expr.Walk(&visitor)
	return exprFmt, visitor.Args
}

func makeBoolExprContext(s *Smither, ctx Context, refs colRefs) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69432)
	return makeScalarSample(s.boolExprSampler, s, ctx, types.Bool, refs)
}

func makeScalarSample(
	sampler *scalarExprSampler, s *Smither, ctx Context, typ *types.T, refs colRefs,
) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69433)

	if ctx.fnClass == tree.AggregateClass {
		__antithesis_instrumentation__.Notify(69437)
		if expr, ok := makeFunc(s, ctx, typ, refs); ok {
			__antithesis_instrumentation__.Notify(69438)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(69439)
		}
	} else {
		__antithesis_instrumentation__.Notify(69440)
	}
	__antithesis_instrumentation__.Notify(69434)
	if s.canRecurse() {
		__antithesis_instrumentation__.Notify(69441)
		for {
			__antithesis_instrumentation__.Notify(69442)

			result, ok := sampler.Next()(s, ctx, typ, refs)
			if ok {
				__antithesis_instrumentation__.Notify(69443)
				return result
			} else {
				__antithesis_instrumentation__.Notify(69444)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(69445)
	}
	__antithesis_instrumentation__.Notify(69435)

	if s.coin() {
		__antithesis_instrumentation__.Notify(69446)
		if expr, ok := makeColRef(s, typ, refs); ok {
			__antithesis_instrumentation__.Notify(69447)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(69448)
		}
	} else {
		__antithesis_instrumentation__.Notify(69449)
	}
	__antithesis_instrumentation__.Notify(69436)
	return makeConstExpr(s, typ, refs)
}

func makeCaseExpr(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69450)
	typ = s.pickAnyType(typ)
	condition := makeScalar(s, types.Bool, refs)
	trueExpr := makeScalar(s, typ, refs)
	falseExpr := makeScalar(s, typ, refs)
	expr, err := tree.NewTypedCaseExpr(
		nil,
		[]*tree.When{{
			Cond: condition,
			Val:  trueExpr,
		}},
		falseExpr,
		typ,
	)
	return expr, err == nil
}

func makeCoalesceExpr(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69451)
	typ = s.pickAnyType(typ)
	firstExpr := makeScalar(s, typ, refs)
	secondExpr := makeScalar(s, typ, refs)
	return tree.NewTypedCoalesceExpr(
		tree.TypedExprs{
			firstExpr,
			secondExpr,
		},
		typ,
	), true
}

func makeConstExpr(s *Smither, typ *types.T, refs colRefs) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69452)
	typ = s.pickAnyType(typ)

	if s.avoidConsts {
		__antithesis_instrumentation__.Notify(69455)
		if expr, ok := makeColRef(s, typ, refs); ok {
			__antithesis_instrumentation__.Notify(69456)
			return expr
		} else {
			__antithesis_instrumentation__.Notify(69457)
		}
	} else {
		__antithesis_instrumentation__.Notify(69458)
	}
	__antithesis_instrumentation__.Notify(69453)

	expr := tree.TypedExpr(makeConstDatum(s, typ))

	if s.postgres {
		__antithesis_instrumentation__.Notify(69459)

		if typ.Family() == types.OidFamily {
			__antithesis_instrumentation__.Notify(69461)
			typ = types.Oid
		} else {
			__antithesis_instrumentation__.Notify(69462)
		}
		__antithesis_instrumentation__.Notify(69460)
		expr = tree.NewTypedCastExpr(expr, typ)
	} else {
		__antithesis_instrumentation__.Notify(69463)
	}
	__antithesis_instrumentation__.Notify(69454)
	return expr
}

func makeConstDatum(s *Smither, typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(69464)
	var datum tree.Datum
	s.lock.Lock()
	datum = randgen.RandDatumWithNullChance(s.rnd, typ, 6)
	if f := datum.ResolvedType().Family(); f != types.UnknownFamily && func() bool {
		__antithesis_instrumentation__.Notify(69466)
		return s.simpleDatums == true
	}() == true {
		__antithesis_instrumentation__.Notify(69467)
		datum = randgen.RandDatumSimple(s.rnd, typ)
	} else {
		__antithesis_instrumentation__.Notify(69468)
	}
	__antithesis_instrumentation__.Notify(69465)
	s.lock.Unlock()

	return datum
}

func makeColRef(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69469)
	expr, _, ok := getColRef(s, typ, refs)
	return expr, ok
}

func getColRef(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, *colRef, bool) {
	__antithesis_instrumentation__.Notify(69470)

	cols := make(colRefs, 0, len(refs))
	for _, c := range refs {
		__antithesis_instrumentation__.Notify(69473)
		if typ.Family() == types.AnyFamily || func() bool {
			__antithesis_instrumentation__.Notify(69474)
			return c.typ.Equivalent(typ) == true
		}() == true {
			__antithesis_instrumentation__.Notify(69475)
			cols = append(cols, c)
		} else {
			__antithesis_instrumentation__.Notify(69476)
		}
	}
	__antithesis_instrumentation__.Notify(69471)
	if len(cols) == 0 {
		__antithesis_instrumentation__.Notify(69477)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69478)
	}
	__antithesis_instrumentation__.Notify(69472)
	col := cols[s.rnd.Intn(len(cols))]
	return col.typedExpr(), col, true
}

func castType(expr tree.TypedExpr, typ *types.T) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69479)

	if typ.IsAmbiguous() {
		__antithesis_instrumentation__.Notify(69481)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(69482)
	}
	__antithesis_instrumentation__.Notify(69480)
	return makeTypedExpr(&tree.CastExpr{
		Expr:       expr,
		Type:       typ,
		SyntaxMode: tree.CastShort,
	}, typ)
}

func typedParen(expr tree.TypedExpr, typ *types.T) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69483)
	return makeTypedExpr(&tree.ParenExpr{Expr: expr}, typ)
}

func makeOr(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69484)
	switch typ.Family() {
	case types.BoolFamily, types.AnyFamily:
		__antithesis_instrumentation__.Notify(69486)
	default:
		__antithesis_instrumentation__.Notify(69487)
		return nil, false
	}
	__antithesis_instrumentation__.Notify(69485)
	left := makeBoolExpr(s, refs)
	right := makeBoolExpr(s, refs)
	return typedParen(tree.NewTypedOrExpr(left, right), types.Bool), true
}

func makeAnd(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69488)
	switch typ.Family() {
	case types.BoolFamily, types.AnyFamily:
		__antithesis_instrumentation__.Notify(69490)
	default:
		__antithesis_instrumentation__.Notify(69491)
		return nil, false
	}
	__antithesis_instrumentation__.Notify(69489)
	left := makeBoolExpr(s, refs)
	right := makeBoolExpr(s, refs)
	return typedParen(tree.NewTypedAndExpr(left, right), types.Bool), true
}

func makeNot(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69492)
	switch typ.Family() {
	case types.BoolFamily, types.AnyFamily:
		__antithesis_instrumentation__.Notify(69494)
	default:
		__antithesis_instrumentation__.Notify(69495)
		return nil, false
	}
	__antithesis_instrumentation__.Notify(69493)
	expr := makeBoolExpr(s, refs)
	return typedParen(tree.NewTypedNotExpr(expr), types.Bool), true
}

var compareOps = [...]treecmp.ComparisonOperatorSymbol{
	treecmp.EQ,
	treecmp.LT,
	treecmp.GT,
	treecmp.LE,
	treecmp.GE,
	treecmp.NE,
	treecmp.IsDistinctFrom,
	treecmp.IsNotDistinctFrom,
}

func makeCompareOp(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69496)
	if f := typ.Family(); f != types.BoolFamily && func() bool {
		__antithesis_instrumentation__.Notify(69499)
		return f != types.AnyFamily == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(69500)
		return f != types.VoidFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(69501)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69502)
	}
	__antithesis_instrumentation__.Notify(69497)
	typ = s.randScalarType()
	op := compareOps[s.rnd.Intn(len(compareOps))]
	if _, ok := tree.CmpOps[op].LookupImpl(typ, typ); !ok {
		__antithesis_instrumentation__.Notify(69503)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69504)
	}
	__antithesis_instrumentation__.Notify(69498)
	left := makeScalar(s, typ, refs)
	right := makeScalar(s, typ, refs)
	return typedParen(tree.NewTypedComparisonExpr(treecmp.MakeComparisonOperator(op), left, right), typ), true
}

func makeBinOp(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69505)
	typ = s.pickAnyType(typ)
	ops := operators[typ.Oid()]
	if len(ops) == 0 {
		__antithesis_instrumentation__.Notify(69509)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69510)
	}
	__antithesis_instrumentation__.Notify(69506)
	n := s.rnd.Intn(len(ops))
	op := ops[n]
	if s.postgres {
		__antithesis_instrumentation__.Notify(69511)
		if ignorePostgresBinOps[binOpTriple{
			op.LeftType.Family(),
			op.Operator.Symbol,
			op.RightType.Family(),
		}] {
			__antithesis_instrumentation__.Notify(69512)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(69513)
		}
	} else {
		__antithesis_instrumentation__.Notify(69514)
	}
	__antithesis_instrumentation__.Notify(69507)
	if s.postgres {
		__antithesis_instrumentation__.Notify(69515)
		if transform, needTransform := postgresBinOpTransformations[binOpTriple{
			op.LeftType.Family(),
			op.Operator.Symbol,
			op.RightType.Family(),
		}]; needTransform {
			__antithesis_instrumentation__.Notify(69516)
			op.LeftType = transform.leftType
			op.RightType = transform.rightType
		} else {
			__antithesis_instrumentation__.Notify(69517)
		}
	} else {
		__antithesis_instrumentation__.Notify(69518)
	}
	__antithesis_instrumentation__.Notify(69508)
	left := makeScalar(s, op.LeftType, refs)
	right := makeScalar(s, op.RightType, refs)
	return castType(
		typedParen(
			tree.NewTypedBinaryExpr(op.Operator, castType(left, op.LeftType), castType(right, op.RightType), typ),
			typ,
		),
		typ,
	), true
}

type binOpTriple struct {
	left  types.Family
	op    treebin.BinaryOperatorSymbol
	right types.Family
}

type binOpOperands struct {
	leftType  *types.T
	rightType *types.T
}

var ignorePostgresBinOps = map[binOpTriple]bool{

	{types.IntFamily, treebin.Div, types.IntFamily}: true,

	{types.FloatFamily, treebin.Mult, types.DateFamily}: true,
	{types.DateFamily, treebin.Mult, types.FloatFamily}: true,
	{types.DateFamily, treebin.Div, types.FloatFamily}:  true,

	{types.IntFamily, treebin.FloorDiv, types.IntFamily}:         true,
	{types.FloatFamily, treebin.FloorDiv, types.FloatFamily}:     true,
	{types.DecimalFamily, treebin.FloorDiv, types.DecimalFamily}: true,
	{types.DecimalFamily, treebin.FloorDiv, types.IntFamily}:     true,
	{types.IntFamily, treebin.FloorDiv, types.DecimalFamily}:     true,

	{types.FloatFamily, treebin.Mod, types.FloatFamily}: true,
}

var postgresBinOpTransformations = map[binOpTriple]binOpOperands{
	{types.IntFamily, treebin.Plus, types.DateFamily}:          {types.Int4, types.Date},
	{types.DateFamily, treebin.Plus, types.IntFamily}:          {types.Date, types.Int4},
	{types.IntFamily, treebin.Minus, types.DateFamily}:         {types.Int4, types.Date},
	{types.DateFamily, treebin.Minus, types.IntFamily}:         {types.Date, types.Int4},
	{types.JsonFamily, treebin.JSONFetchVal, types.IntFamily}:  {types.Jsonb, types.Int4},
	{types.JsonFamily, treebin.JSONFetchText, types.IntFamily}: {types.Jsonb, types.Int4},
	{types.JsonFamily, treebin.Minus, types.IntFamily}:         {types.Jsonb, types.Int4},
}

func makeFunc(s *Smither, ctx Context, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69519)
	typ = s.pickAnyType(typ)

	class := ctx.fnClass

	if class == tree.WindowClass && func() bool {
		__antithesis_instrumentation__.Notify(69527)
		return s.d6() != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69528)
		class = tree.NormalClass
	} else {
		__antithesis_instrumentation__.Notify(69529)
	}
	__antithesis_instrumentation__.Notify(69520)
	fns := functions[class][typ.Oid()]
	if len(fns) == 0 {
		__antithesis_instrumentation__.Notify(69530)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69531)
	}
	__antithesis_instrumentation__.Notify(69521)
	fn := fns[s.rnd.Intn(len(fns))]
	if s.disableImpureFns && func() bool {
		__antithesis_instrumentation__.Notify(69532)
		return fn.overload.Volatility > tree.VolatilityImmutable == true
	}() == true {
		__antithesis_instrumentation__.Notify(69533)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69534)
	}
	__antithesis_instrumentation__.Notify(69522)
	for _, ignore := range s.ignoreFNs {
		__antithesis_instrumentation__.Notify(69535)
		if ignore.MatchString(fn.def.Name) {
			__antithesis_instrumentation__.Notify(69536)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(69537)
		}
	}
	__antithesis_instrumentation__.Notify(69523)

	args := make(tree.TypedExprs, 0)
	for _, argTyp := range fn.overload.Types.Types() {
		__antithesis_instrumentation__.Notify(69538)

		if s.postgres && func() bool {
			__antithesis_instrumentation__.Notify(69542)
			return argTyp.Family() == types.IntFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(69543)
			argTyp = types.Int4
		} else {
			__antithesis_instrumentation__.Notify(69544)
		}
		__antithesis_instrumentation__.Notify(69539)
		var arg tree.TypedExpr

		if class == tree.AggregateClass || func() bool {
			__antithesis_instrumentation__.Notify(69545)
			return class == tree.WindowClass == true
		}() == true {
			__antithesis_instrumentation__.Notify(69546)
			var ok bool
			arg, ok = makeColRef(s, argTyp, refs)
			if !ok {
				__antithesis_instrumentation__.Notify(69547)

				arg = makeConstExpr(s, typ, refs)
			} else {
				__antithesis_instrumentation__.Notify(69548)
			}
		} else {
			__antithesis_instrumentation__.Notify(69549)
		}
		__antithesis_instrumentation__.Notify(69540)
		if arg == nil {
			__antithesis_instrumentation__.Notify(69550)
			arg = makeScalar(s, argTyp, refs)
		} else {
			__antithesis_instrumentation__.Notify(69551)
		}
		__antithesis_instrumentation__.Notify(69541)
		args = append(args, castType(arg, argTyp))
	}
	__antithesis_instrumentation__.Notify(69524)

	if fn.def.Class == tree.WindowClass && func() bool {
		__antithesis_instrumentation__.Notify(69552)
		return s.disableWindowFuncs == true
	}() == true {
		__antithesis_instrumentation__.Notify(69553)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69554)
	}
	__antithesis_instrumentation__.Notify(69525)

	var window *tree.WindowDef

	if fn.def.Class == tree.WindowClass || func() bool {
		__antithesis_instrumentation__.Notify(69555)
		return (!s.disableWindowFuncs && func() bool {
			__antithesis_instrumentation__.Notify(69556)
			return !ctx.noWindow == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(69557)
			return s.d6() == 1 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(69558)
			return fn.def.Class == tree.AggregateClass == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(69559)
		var parts tree.Exprs
		s.sample(len(refs), 2, func(i int) {
			__antithesis_instrumentation__.Notify(69563)
			parts = append(parts, refs[i].item)
		})
		__antithesis_instrumentation__.Notify(69560)
		var (
			order      tree.OrderBy
			orderTypes []*types.T
		)
		s.sample(len(refs)-len(parts), 2, func(i int) {
			__antithesis_instrumentation__.Notify(69564)
			ref := refs[i+len(parts)]
			order = append(order, &tree.Order{
				Expr:      ref.item,
				Direction: s.randDirection(),
			})
			orderTypes = append(orderTypes, ref.typ)
		})
		__antithesis_instrumentation__.Notify(69561)
		var frame *tree.WindowFrame
		if s.coin() {
			__antithesis_instrumentation__.Notify(69565)
			frame = makeWindowFrame(s, refs, orderTypes)
		} else {
			__antithesis_instrumentation__.Notify(69566)
		}
		__antithesis_instrumentation__.Notify(69562)
		window = &tree.WindowDef{
			Partitions: parts,
			OrderBy:    order,
			Frame:      frame,
		}
	} else {
		__antithesis_instrumentation__.Notify(69567)
	}
	__antithesis_instrumentation__.Notify(69526)

	return castType(tree.NewTypedFuncExpr(
		tree.ResolvableFunctionReference{FunctionReference: fn.def},
		0,
		args,
		nil,
		window,
		typ,
		&fn.def.FunctionProperties,
		fn.overload,
	), typ), true
}

var windowFrameModes = []treewindow.WindowFrameMode{
	treewindow.RANGE,
	treewindow.ROWS,
	treewindow.GROUPS,
}

func randWindowFrameMode(s *Smither) treewindow.WindowFrameMode {
	__antithesis_instrumentation__.Notify(69568)
	return windowFrameModes[s.rnd.Intn(len(windowFrameModes))]
}

func makeWindowFrame(s *Smither, refs colRefs, orderTypes []*types.T) *tree.WindowFrame {
	__antithesis_instrumentation__.Notify(69569)
	var frameMode treewindow.WindowFrameMode
	for {
		__antithesis_instrumentation__.Notify(69573)
		frameMode = randWindowFrameMode(s)
		if len(orderTypes) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(69574)
			return frameMode != treewindow.GROUPS == true
		}() == true {
			__antithesis_instrumentation__.Notify(69575)

			break
		} else {
			__antithesis_instrumentation__.Notify(69576)
		}
	}
	__antithesis_instrumentation__.Notify(69570)

	var startBound tree.WindowFrameBound
	var endBound *tree.WindowFrameBound

	allowRangeWithOffsets := false
	if len(orderTypes) == 1 {
		__antithesis_instrumentation__.Notify(69577)
		switch orderTypes[0].Family() {
		case types.IntFamily, types.FloatFamily,
			types.DecimalFamily, types.IntervalFamily:
			__antithesis_instrumentation__.Notify(69578)
			allowRangeWithOffsets = true
		default:
			__antithesis_instrumentation__.Notify(69579)
		}
	} else {
		__antithesis_instrumentation__.Notify(69580)
	}
	__antithesis_instrumentation__.Notify(69571)
	if frameMode == treewindow.RANGE && func() bool {
		__antithesis_instrumentation__.Notify(69581)
		return !allowRangeWithOffsets == true
	}() == true {
		__antithesis_instrumentation__.Notify(69582)
		if s.coin() {
			__antithesis_instrumentation__.Notify(69584)
			startBound.BoundType = treewindow.UnboundedPreceding
		} else {
			__antithesis_instrumentation__.Notify(69585)
			startBound.BoundType = treewindow.CurrentRow
		}
		__antithesis_instrumentation__.Notify(69583)
		if s.coin() {
			__antithesis_instrumentation__.Notify(69586)
			endBound = new(tree.WindowFrameBound)
			if s.coin() {
				__antithesis_instrumentation__.Notify(69587)
				endBound.BoundType = treewindow.CurrentRow
			} else {
				__antithesis_instrumentation__.Notify(69588)
				endBound.BoundType = treewindow.UnboundedFollowing
			}
		} else {
			__antithesis_instrumentation__.Notify(69589)
		}
	} else {
		__antithesis_instrumentation__.Notify(69590)

		startBound.BoundType = treewindow.WindowFrameBoundType(s.rnd.Intn(4))
		if startBound.BoundType == treewindow.OffsetFollowing {
			__antithesis_instrumentation__.Notify(69594)

			endBound = new(tree.WindowFrameBound)
			if s.coin() {
				__antithesis_instrumentation__.Notify(69595)
				endBound.BoundType = treewindow.OffsetFollowing
			} else {
				__antithesis_instrumentation__.Notify(69596)
				endBound.BoundType = treewindow.UnboundedFollowing
			}
		} else {
			__antithesis_instrumentation__.Notify(69597)
		}
		__antithesis_instrumentation__.Notify(69591)
		if endBound == nil && func() bool {
			__antithesis_instrumentation__.Notify(69598)
			return s.coin() == true
		}() == true {
			__antithesis_instrumentation__.Notify(69599)
			endBound = new(tree.WindowFrameBound)

			endBoundProhibitedChoices := int(startBound.BoundType)
			if startBound.BoundType == treewindow.UnboundedPreceding {
				__antithesis_instrumentation__.Notify(69601)

				endBoundProhibitedChoices = 1
			} else {
				__antithesis_instrumentation__.Notify(69602)
			}
			__antithesis_instrumentation__.Notify(69600)
			endBound.BoundType = treewindow.WindowFrameBoundType(endBoundProhibitedChoices + s.rnd.Intn(5-endBoundProhibitedChoices))
		} else {
			__antithesis_instrumentation__.Notify(69603)
		}
		__antithesis_instrumentation__.Notify(69592)

		typ := types.Int
		if frameMode == treewindow.RANGE {
			__antithesis_instrumentation__.Notify(69604)
			typ = orderTypes[0]
		} else {
			__antithesis_instrumentation__.Notify(69605)
		}
		__antithesis_instrumentation__.Notify(69593)
		startBound.OffsetExpr = makeScalar(s, typ, refs)
		if endBound != nil {
			__antithesis_instrumentation__.Notify(69606)
			endBound.OffsetExpr = makeScalar(s, typ, refs)
		} else {
			__antithesis_instrumentation__.Notify(69607)
		}
	}
	__antithesis_instrumentation__.Notify(69572)
	return &tree.WindowFrame{
		Mode: frameMode,
		Bounds: tree.WindowFrameBounds{
			StartBound: &startBound,
			EndBound:   endBound,
		},
	}
}

func makeExists(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69608)
	switch typ.Family() {
	case types.BoolFamily, types.AnyFamily:
		__antithesis_instrumentation__.Notify(69611)
	default:
		__antithesis_instrumentation__.Notify(69612)
		return nil, false
	}
	__antithesis_instrumentation__.Notify(69609)

	selectStmt, _, ok := s.makeSelect(s.makeDesiredTypes(), refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69613)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69614)
	}
	__antithesis_instrumentation__.Notify(69610)

	subq := &tree.Subquery{
		Select: &tree.ParenSelect{Select: selectStmt},
		Exists: true,
	}
	subq.SetType(types.Bool)
	return subq, true
}

func makeIn(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69615)
	switch typ.Family() {
	case types.BoolFamily, types.AnyFamily:
		__antithesis_instrumentation__.Notify(69619)
	default:
		__antithesis_instrumentation__.Notify(69620)
		return nil, false
	}
	__antithesis_instrumentation__.Notify(69616)

	t := s.randScalarType()
	var rhs tree.TypedExpr
	if s.coin() {
		__antithesis_instrumentation__.Notify(69621)
		rhs = makeTuple(s, t, refs)
	} else {
		__antithesis_instrumentation__.Notify(69622)
		selectStmt, _, ok := s.makeSelect([]*types.T{t}, refs)
		if !ok {
			__antithesis_instrumentation__.Notify(69624)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(69625)
		}
		__antithesis_instrumentation__.Notify(69623)

		clause := selectStmt.Select.(*tree.SelectClause)
		clause.Exprs[0].Expr = &tree.CastExpr{
			Expr:       clause.Exprs[0].Expr,
			Type:       t,
			SyntaxMode: tree.CastShort,
		}
		subq := &tree.Subquery{
			Select: &tree.ParenSelect{Select: selectStmt},
		}
		subq.SetType(types.MakeTuple([]*types.T{t}))
		rhs = subq
	}
	__antithesis_instrumentation__.Notify(69617)
	op := treecmp.In
	if s.coin() {
		__antithesis_instrumentation__.Notify(69626)
		op = treecmp.NotIn
	} else {
		__antithesis_instrumentation__.Notify(69627)
	}
	__antithesis_instrumentation__.Notify(69618)
	return tree.NewTypedComparisonExpr(
		treecmp.MakeComparisonOperator(op),

		castType(makeScalar(s, t, refs), t),
		rhs,
	), true
}

func makeStringComparison(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69628)
	stringComparison := s.randStringComparison()
	switch typ.Family() {
	case types.BoolFamily, types.AnyFamily:
		__antithesis_instrumentation__.Notify(69630)
	default:
		__antithesis_instrumentation__.Notify(69631)
		return nil, false
	}
	__antithesis_instrumentation__.Notify(69629)
	return tree.NewTypedComparisonExpr(
		stringComparison,
		makeScalar(s, types.String, refs),
		makeScalar(s, types.String, refs),
	), true
}

func makeTuple(s *Smither, typ *types.T, refs colRefs) *tree.Tuple {
	__antithesis_instrumentation__.Notify(69632)
	n := s.rnd.Intn(5)

	if n == 0 && func() bool {
		__antithesis_instrumentation__.Notify(69635)
		return s.simpleDatums == true
	}() == true {
		__antithesis_instrumentation__.Notify(69636)
		n++
	} else {
		__antithesis_instrumentation__.Notify(69637)
	}
	__antithesis_instrumentation__.Notify(69633)
	exprs := make(tree.Exprs, n)
	for i := range exprs {
		__antithesis_instrumentation__.Notify(69638)
		if s.d9() == 1 {
			__antithesis_instrumentation__.Notify(69639)
			exprs[i] = makeConstDatum(s, typ)
		} else {
			__antithesis_instrumentation__.Notify(69640)
			exprs[i] = makeScalar(s, typ, refs)
		}
	}
	__antithesis_instrumentation__.Notify(69634)
	return tree.NewTypedTuple(types.MakeTuple([]*types.T{typ}), exprs)
}

func makeScalarSubquery(s *Smither, typ *types.T, refs colRefs) (tree.TypedExpr, bool) {
	__antithesis_instrumentation__.Notify(69641)
	if s.disableLimits {
		__antithesis_instrumentation__.Notify(69644)

		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69645)
	}
	__antithesis_instrumentation__.Notify(69642)
	selectStmt, _, ok := s.makeSelect([]*types.T{typ}, refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69646)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69647)
	}
	__antithesis_instrumentation__.Notify(69643)
	selectStmt.Limit = &tree.Limit{Count: tree.NewDInt(1)}

	subq := &tree.Subquery{
		Select: &tree.ParenSelect{Select: selectStmt},
	}
	subq.SetType(typ)

	return subq, true
}

type replaceDatumPlaceholderVisitor struct {
	Args []interface{}
}

var _ tree.Visitor = &replaceDatumPlaceholderVisitor{}

func (v *replaceDatumPlaceholderVisitor) VisitPre(
	expr tree.Expr,
) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(69648)
	switch t := expr.(type) {
	case tree.Datum:
		__antithesis_instrumentation__.Notify(69650)
		if t.ResolvedType().IsNumeric() || func() bool {
			__antithesis_instrumentation__.Notify(69652)
			return t.ResolvedType() == types.Bool == true
		}() == true {
			__antithesis_instrumentation__.Notify(69653)
			v.Args = append(v.Args, expr)
			placeholder, _ := tree.NewPlaceholder(strconv.Itoa(len(v.Args)))
			return false, placeholder
		} else {
			__antithesis_instrumentation__.Notify(69654)
		}
		__antithesis_instrumentation__.Notify(69651)
		return false, expr
	}
	__antithesis_instrumentation__.Notify(69649)
	return true, expr
}

func (*replaceDatumPlaceholderVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(69655)
	return expr
}
