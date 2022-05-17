package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func initWindowBuiltins() {
	__antithesis_instrumentation__.Notify(602456)

	for k, v := range windows {
		__antithesis_instrumentation__.Notify(602457)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(602461)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(602462)
		}
		__antithesis_instrumentation__.Notify(602458)

		if v.props.Class != tree.WindowClass {
			__antithesis_instrumentation__.Notify(602463)
			panic(errors.AssertionFailedf("%s: window functions should be marked with the tree.WindowClass "+
				"function class, found %v", k, v))
		} else {
			__antithesis_instrumentation__.Notify(602464)
		}
		__antithesis_instrumentation__.Notify(602459)
		for _, w := range v.overloads {
			__antithesis_instrumentation__.Notify(602465)
			if w.WindowFunc == nil {
				__antithesis_instrumentation__.Notify(602466)
				panic(errors.AssertionFailedf("%s: window functions should have tree.WindowFunc constructors, "+
					"found %v", k, w))
			} else {
				__antithesis_instrumentation__.Notify(602467)
			}
		}
		__antithesis_instrumentation__.Notify(602460)
		builtins[k] = v
	}
}

func winProps() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(602468)
	return tree.FunctionProperties{
		Class: tree.WindowClass,
	}
}

var windows = map[string]builtinDefinition{
	"row_number": makeBuiltin(winProps(),
		makeWindowOverload(
			tree.ArgTypes{},
			types.Int,
			newRowNumberWindow,
			"Calculates the number of the current row within its partition, counting from 1.",
			tree.VolatilityImmutable,
		),
	),
	"rank": makeBuiltin(winProps(),
		makeWindowOverload(
			tree.ArgTypes{},
			types.Int,
			newRankWindow,
			"Calculates the rank of the current row with gaps; same as row_number of its first peer.",
			tree.VolatilityImmutable,
		),
	),
	"dense_rank": makeBuiltin(winProps(),
		makeWindowOverload(
			tree.ArgTypes{},
			types.Int,
			newDenseRankWindow,
			"Calculates the rank of the current row without gaps; this function counts peer groups.",
			tree.VolatilityImmutable,
		),
	),
	"percent_rank": makeBuiltin(winProps(),
		makeWindowOverload(
			tree.ArgTypes{},
			types.Float,
			newPercentRankWindow,
			"Calculates the relative rank of the current row: (rank - 1) / (total rows - 1).",
			tree.VolatilityImmutable,
		),
	),
	"cume_dist": makeBuiltin(winProps(),
		makeWindowOverload(
			tree.ArgTypes{},
			types.Float,
			newCumulativeDistWindow,
			"Calculates the relative rank of the current row: "+
				"(number of rows preceding or peer with current row) / (total rows).",
			tree.VolatilityImmutable,
		),
	),
	"ntile": makeBuiltin(winProps(),
		makeWindowOverload(
			tree.ArgTypes{{"n", types.Int}},
			types.Int,
			newNtileWindow,
			"Calculates an integer ranging from 1 to `n`, dividing the partition as equally as possible.",
			tree.VolatilityImmutable,
		),
	),
	"lag": collectOverloads(
		winProps(),
		types.Scalar,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602469)
			return makeWindowOverload(
				tree.ArgTypes{{"val", t}},
				t,
				makeLeadLagWindowConstructor(false, false, false),
				"Returns `val` evaluated at the previous row within current row's partition; "+
					"if there is no such row, instead returns null.",
				tree.VolatilityImmutable,
			)
		},
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602470)
			return makeWindowOverload(
				tree.ArgTypes{{"val", t}, {"n", types.Int}},
				t,
				makeLeadLagWindowConstructor(false, true, false),
				"Returns `val` evaluated at the row that is `n` rows before the current row within its partition; "+
					"if there is no such row, instead returns null. `n` is evaluated with respect to the current row.",
				tree.VolatilityImmutable,
			)
		},

		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602471)
			return makeWindowOverload(
				tree.ArgTypes{
					{"val", t}, {"n", types.Int}, {"default", t},
				},
				t,
				makeLeadLagWindowConstructor(false, true, true),
				"Returns `val` evaluated at the row that is `n` rows before the current row within its partition; "+
					"if there is no such, row, instead returns `default` (which must be of the same type as `val`). "+
					"Both `n` and `default` are evaluated with respect to the current row.",
				tree.VolatilityImmutable,
			)
		},
	),
	"lead": collectOverloads(winProps(), types.Scalar,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602472)
			return makeWindowOverload(
				tree.ArgTypes{{"val", t}},
				t,
				makeLeadLagWindowConstructor(true, false, false),
				"Returns `val` evaluated at the following row within current row's partition; "+""+
					"if there is no such row, instead returns null.",
				tree.VolatilityImmutable,
			)
		},
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602473)
			return makeWindowOverload(
				tree.ArgTypes{{"val", t}, {"n", types.Int}},
				t,
				makeLeadLagWindowConstructor(true, true, false),
				"Returns `val` evaluated at the row that is `n` rows after the current row within its partition; "+
					"if there is no such row, instead returns null. `n` is evaluated with respect to the current row.",
				tree.VolatilityImmutable,
			)
		},
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602474)
			return makeWindowOverload(
				tree.ArgTypes{
					{"val", t}, {"n", types.Int}, {"default", t},
				},
				t,
				makeLeadLagWindowConstructor(true, true, true),
				"Returns `val` evaluated at the row that is `n` rows after the current row within its partition; "+
					"if there is no such, row, instead returns `default` (which must be of the same type as `val`). "+
					"Both `n` and `default` are evaluated with respect to the current row.",
				tree.VolatilityImmutable,
			)
		},
	),
	"first_value": collectOverloads(
		winProps(),
		types.Scalar,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602475)
			return makeWindowOverload(
				tree.ArgTypes{{"val", t}},
				t,
				newFirstValueWindow,
				"Returns `val` evaluated at the row that is the first row of the window frame.",
				tree.VolatilityImmutable,
			)
		}),
	"last_value": collectOverloads(
		winProps(),
		types.Scalar,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602476)
			return makeWindowOverload(
				tree.ArgTypes{{"val", t}},
				t,
				newLastValueWindow,
				"Returns `val` evaluated at the row that is the last row of the window frame.",
				tree.VolatilityImmutable,
			)
		}),
	"nth_value": collectOverloads(winProps(), types.Scalar,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(602477)
			return makeWindowOverload(
				tree.ArgTypes{
					{"val", t}, {"n", types.Int},
				},
				t,
				newNthValueWindow,
				"Returns `val` evaluated at the row that is the `n`th row of the window frame (counting from 1); "+
					"null if no such row.",
				tree.VolatilityImmutable,
			)
		}),
}

func makeWindowOverload(
	in tree.ArgTypes,
	ret *types.T,
	f func([]*types.T, *tree.EvalContext) tree.WindowFunc,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(602478)
	return tree.Overload{
		Types:      in,
		ReturnType: tree.FixedReturnType(ret),
		WindowFunc: f,
		Info:       info,
		Volatility: volatility,
	}
}

var _ tree.WindowFunc = &aggregateWindowFunc{}
var _ tree.WindowFunc = &framableAggregateWindowFunc{}
var _ tree.WindowFunc = &rowNumberWindow{}
var _ tree.WindowFunc = &rankWindow{}
var _ tree.WindowFunc = &denseRankWindow{}
var _ tree.WindowFunc = &percentRankWindow{}
var _ tree.WindowFunc = &cumulativeDistWindow{}
var _ tree.WindowFunc = &ntileWindow{}
var _ tree.WindowFunc = &leadLagWindow{}
var _ tree.WindowFunc = &firstValueWindow{}
var _ tree.WindowFunc = &lastValueWindow{}
var _ tree.WindowFunc = &nthValueWindow{}

type aggregateWindowFunc struct {
	agg     tree.AggregateFunc
	peerRes tree.Datum

	peerFrameStartIdx, peerFrameEndIdx int
}

func NewAggregateWindowFunc(
	aggConstructor func(*tree.EvalContext, tree.Datums) tree.AggregateFunc,
) func(*tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602479)
	return func(evalCtx *tree.EvalContext) tree.WindowFunc {
		__antithesis_instrumentation__.Notify(602480)
		return &aggregateWindowFunc{agg: aggConstructor(evalCtx, nil)}
	}
}

func (w *aggregateWindowFunc) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602481)
	if !wfr.FirstInPeerGroup() && func() bool {
		__antithesis_instrumentation__.Notify(602485)
		return wfr.Frame.DefaultFrameExclusion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(602486)
		return w.peerRes, nil
	} else {
		__antithesis_instrumentation__.Notify(602487)
	}
	__antithesis_instrumentation__.Notify(602482)

	peerGroupRowCount := wfr.PeerHelper.GetRowCount(wfr.CurRowPeerGroupNum)
	for i := 0; i < peerGroupRowCount; i++ {
		__antithesis_instrumentation__.Notify(602488)
		if skipped, err := wfr.IsRowSkipped(ctx, wfr.RowIdx+i); err != nil {
			__antithesis_instrumentation__.Notify(602492)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602493)
			if skipped {
				__antithesis_instrumentation__.Notify(602494)
				continue
			} else {
				__antithesis_instrumentation__.Notify(602495)
			}
		}
		__antithesis_instrumentation__.Notify(602489)
		args, err := wfr.ArgsWithRowOffset(ctx, i)
		if err != nil {
			__antithesis_instrumentation__.Notify(602496)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602497)
		}
		__antithesis_instrumentation__.Notify(602490)
		var value tree.Datum
		var others tree.Datums

		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(602498)
			value = args[0]
			others = args[1:]
		} else {
			__antithesis_instrumentation__.Notify(602499)
		}
		__antithesis_instrumentation__.Notify(602491)
		if err := w.agg.Add(ctx, value, others...); err != nil {
			__antithesis_instrumentation__.Notify(602500)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602501)
		}
	}
	__antithesis_instrumentation__.Notify(602483)

	peerRes, err := w.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(602502)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602503)
	}
	__antithesis_instrumentation__.Notify(602484)
	w.peerRes = peerRes
	w.peerFrameStartIdx = wfr.RowIdx
	w.peerFrameEndIdx = wfr.RowIdx + peerGroupRowCount
	return w.peerRes, nil
}

func (w *aggregateWindowFunc) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(602504)
	w.agg.Reset(ctx)
	w.peerRes = nil
	w.peerFrameStartIdx = 0
	w.peerFrameEndIdx = 0
}

func (w *aggregateWindowFunc) Close(ctx context.Context, _ *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602505)
	w.agg.Close(ctx)
}

func ShouldReset(w tree.WindowFunc) {
	__antithesis_instrumentation__.Notify(602506)
	if f, ok := w.(*framableAggregateWindowFunc); ok {
		__antithesis_instrumentation__.Notify(602507)
		f.shouldReset = true
	} else {
		__antithesis_instrumentation__.Notify(602508)
	}
}

type framableAggregateWindowFunc struct {
	agg            *aggregateWindowFunc
	aggConstructor func(*tree.EvalContext, tree.Datums) tree.AggregateFunc
	shouldReset    bool
}

func newFramableAggregateWindow(
	agg tree.AggregateFunc, aggConstructor func(*tree.EvalContext, tree.Datums) tree.AggregateFunc,
) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602509)

	_, shouldReset := agg.(*jsonObjectAggregate)
	return &framableAggregateWindowFunc{
		agg:            &aggregateWindowFunc{agg: agg, peerRes: tree.DNull},
		aggConstructor: aggConstructor,
		shouldReset:    shouldReset,
	}
}

func (w *framableAggregateWindowFunc) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602510)
	if wfr.FullPartitionIsInWindow() {
		__antithesis_instrumentation__.Notify(602518)

		if wfr.RowIdx > 0 {
			__antithesis_instrumentation__.Notify(602519)
			return w.agg.peerRes, nil
		} else {
			__antithesis_instrumentation__.Notify(602520)
		}
	} else {
		__antithesis_instrumentation__.Notify(602521)
	}
	__antithesis_instrumentation__.Notify(602511)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602522)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602523)
	}
	__antithesis_instrumentation__.Notify(602512)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602524)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602525)
	}
	__antithesis_instrumentation__.Notify(602513)
	if !wfr.FirstInPeerGroup() && func() bool {
		__antithesis_instrumentation__.Notify(602526)
		return wfr.Frame.DefaultFrameExclusion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(602527)

		if frameStartIdx == w.agg.peerFrameStartIdx && func() bool {
			__antithesis_instrumentation__.Notify(602528)
			return frameEndIdx == w.agg.peerFrameEndIdx == true
		}() == true {
			__antithesis_instrumentation__.Notify(602529)

			return w.agg.peerRes, nil
		} else {
			__antithesis_instrumentation__.Notify(602530)
		}

	} else {
		__antithesis_instrumentation__.Notify(602531)
	}
	__antithesis_instrumentation__.Notify(602514)
	if !w.shouldReset {
		__antithesis_instrumentation__.Notify(602532)

		return w.agg.Compute(ctx, evalCtx, wfr)
	} else {
		__antithesis_instrumentation__.Notify(602533)
	}
	__antithesis_instrumentation__.Notify(602515)

	w.agg.Close(ctx, evalCtx)

	*w.agg = aggregateWindowFunc{
		agg:     w.aggConstructor(evalCtx, nil),
		peerRes: tree.DNull,
	}

	for i := frameStartIdx; i < frameEndIdx; i++ {
		__antithesis_instrumentation__.Notify(602534)
		if skipped, err := wfr.IsRowSkipped(ctx, i); err != nil {
			__antithesis_instrumentation__.Notify(602538)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602539)
			if skipped {
				__antithesis_instrumentation__.Notify(602540)
				continue
			} else {
				__antithesis_instrumentation__.Notify(602541)
			}
		}
		__antithesis_instrumentation__.Notify(602535)
		args, err := wfr.ArgsByRowIdx(ctx, i)
		if err != nil {
			__antithesis_instrumentation__.Notify(602542)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602543)
		}
		__antithesis_instrumentation__.Notify(602536)
		var value tree.Datum
		var others tree.Datums

		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(602544)
			value = args[0]
			others = args[1:]
		} else {
			__antithesis_instrumentation__.Notify(602545)
		}
		__antithesis_instrumentation__.Notify(602537)
		if err := w.agg.agg.Add(ctx, value, others...); err != nil {
			__antithesis_instrumentation__.Notify(602546)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602547)
		}
	}
	__antithesis_instrumentation__.Notify(602516)

	peerRes, err := w.agg.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(602548)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602549)
	}
	__antithesis_instrumentation__.Notify(602517)
	w.agg.peerRes = peerRes
	w.agg.peerFrameStartIdx = frameStartIdx
	w.agg.peerFrameEndIdx = frameEndIdx
	return w.agg.peerRes, nil
}

func (w *framableAggregateWindowFunc) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(602550)
	w.agg.Reset(ctx)
}

func (w *framableAggregateWindowFunc) Close(ctx context.Context, evalCtx *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602551)
	w.agg.Close(ctx, evalCtx)
}

type rowNumberWindow struct{}

func newRowNumberWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602552)
	return &rowNumberWindow{}
}

func (rowNumberWindow) Compute(
	_ context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602553)
	return tree.NewDInt(tree.DInt(wfr.RowIdx + 1)), nil
}

func (rowNumberWindow) Reset(context.Context) { __antithesis_instrumentation__.Notify(602554) }

func (rowNumberWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602555)
}

type rankWindow struct {
	peerRes *tree.DInt
}

func newRankWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602556)
	return &rankWindow{}
}

func (w *rankWindow) Compute(
	_ context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602557)
	if wfr.FirstInPeerGroup() {
		__antithesis_instrumentation__.Notify(602559)
		w.peerRes = tree.NewDInt(tree.DInt(wfr.Rank()))
	} else {
		__antithesis_instrumentation__.Notify(602560)
	}
	__antithesis_instrumentation__.Notify(602558)
	return w.peerRes, nil
}

func (w *rankWindow) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(602561)
	w.peerRes = nil
}

func (w *rankWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602562)
}

type denseRankWindow struct {
	denseRank int
	peerRes   *tree.DInt
}

func newDenseRankWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602563)
	return &denseRankWindow{}
}

func (w *denseRankWindow) Compute(
	_ context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602564)
	if wfr.FirstInPeerGroup() {
		__antithesis_instrumentation__.Notify(602566)
		w.denseRank++
		w.peerRes = tree.NewDInt(tree.DInt(w.denseRank))
	} else {
		__antithesis_instrumentation__.Notify(602567)
	}
	__antithesis_instrumentation__.Notify(602565)
	return w.peerRes, nil
}

func (w *denseRankWindow) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(602568)
	w.denseRank = 0
	w.peerRes = nil
}

func (w *denseRankWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602569)
}

type percentRankWindow struct {
	peerRes *tree.DFloat
}

func newPercentRankWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602570)
	return &percentRankWindow{}
}

var dfloatZero = tree.NewDFloat(0)

func (w *percentRankWindow) Compute(
	_ context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602571)

	if wfr.PartitionSize() <= 1 {
		__antithesis_instrumentation__.Notify(602574)
		return dfloatZero, nil
	} else {
		__antithesis_instrumentation__.Notify(602575)
	}
	__antithesis_instrumentation__.Notify(602572)

	if wfr.FirstInPeerGroup() {
		__antithesis_instrumentation__.Notify(602576)

		w.peerRes = tree.NewDFloat(tree.DFloat(wfr.Rank()-1) / tree.DFloat(wfr.PartitionSize()-1))
	} else {
		__antithesis_instrumentation__.Notify(602577)
	}
	__antithesis_instrumentation__.Notify(602573)
	return w.peerRes, nil
}

func (w *percentRankWindow) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(602578)
	w.peerRes = nil
}

func (w *percentRankWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602579)
}

type cumulativeDistWindow struct {
	peerRes *tree.DFloat
}

func newCumulativeDistWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602580)
	return &cumulativeDistWindow{}
}

func (w *cumulativeDistWindow) Compute(
	_ context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602581)
	if wfr.FirstInPeerGroup() {
		__antithesis_instrumentation__.Notify(602583)

		w.peerRes = tree.NewDFloat(tree.DFloat(wfr.DefaultFrameSize()) / tree.DFloat(wfr.PartitionSize()))
	} else {
		__antithesis_instrumentation__.Notify(602584)
	}
	__antithesis_instrumentation__.Notify(602582)
	return w.peerRes, nil
}

func (w *cumulativeDistWindow) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(602585)
	w.peerRes = nil
}

func (w *cumulativeDistWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602586)
}

type ntileWindow struct {
	ntile          *tree.DInt
	curBucketCount int
	boundary       int
	remainder      int
}

func newNtileWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602587)
	return &ntileWindow{}
}

var ErrInvalidArgumentForNtile = pgerror.Newf(
	pgcode.InvalidParameterValue, "argument of ntile() must be greater than zero")

func (w *ntileWindow) Compute(
	ctx context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602588)
	if w.ntile == nil {
		__antithesis_instrumentation__.Notify(602591)

		total := wfr.PartitionSize()

		args, err := wfr.Args(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(602595)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602596)
		}
		__antithesis_instrumentation__.Notify(602592)
		arg := args[0]
		if arg == tree.DNull {
			__antithesis_instrumentation__.Notify(602597)

			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(602598)
		}
		__antithesis_instrumentation__.Notify(602593)

		nbuckets := int(tree.MustBeDInt(arg))
		if nbuckets <= 0 {
			__antithesis_instrumentation__.Notify(602599)

			return nil, ErrInvalidArgumentForNtile
		} else {
			__antithesis_instrumentation__.Notify(602600)
		}
		__antithesis_instrumentation__.Notify(602594)

		w.ntile = tree.NewDInt(1)
		w.curBucketCount = 0
		w.boundary = total / nbuckets
		if w.boundary <= 0 {
			__antithesis_instrumentation__.Notify(602601)
			w.boundary = 1
		} else {
			__antithesis_instrumentation__.Notify(602602)

			w.remainder = total % nbuckets
			if w.remainder != 0 {
				__antithesis_instrumentation__.Notify(602603)
				w.boundary++
			} else {
				__antithesis_instrumentation__.Notify(602604)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(602605)
	}
	__antithesis_instrumentation__.Notify(602589)

	w.curBucketCount++
	if w.boundary < w.curBucketCount {
		__antithesis_instrumentation__.Notify(602606)

		if w.remainder != 0 && func() bool {
			__antithesis_instrumentation__.Notify(602608)
			return int(*w.ntile) == w.remainder == true
		}() == true {
			__antithesis_instrumentation__.Notify(602609)
			w.remainder = 0
			w.boundary--
		} else {
			__antithesis_instrumentation__.Notify(602610)
		}
		__antithesis_instrumentation__.Notify(602607)
		w.ntile = tree.NewDInt(*w.ntile + 1)
		w.curBucketCount = 1
	} else {
		__antithesis_instrumentation__.Notify(602611)
	}
	__antithesis_instrumentation__.Notify(602590)
	return w.ntile, nil
}

func (w *ntileWindow) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(602612)
	w.boundary = 0
	w.curBucketCount = 0
	w.ntile = nil
	w.remainder = 0
}

func (w *ntileWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602613)
}

type leadLagWindow struct {
	forward     bool
	withOffset  bool
	withDefault bool
}

func newLeadLagWindow(forward, withOffset, withDefault bool) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602614)
	return &leadLagWindow{
		forward:     forward,
		withOffset:  withOffset,
		withDefault: withDefault,
	}
}

func makeLeadLagWindowConstructor(
	forward, withOffset, withDefault bool,
) func([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602615)
	return func([]*types.T, *tree.EvalContext) tree.WindowFunc {
		__antithesis_instrumentation__.Notify(602616)
		return newLeadLagWindow(forward, withOffset, withDefault)
	}
}

func (w *leadLagWindow) Compute(
	ctx context.Context, _ *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602617)
	offset := 1
	if w.withOffset {
		__antithesis_instrumentation__.Notify(602622)
		args, err := wfr.Args(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(602625)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602626)
		}
		__antithesis_instrumentation__.Notify(602623)
		offsetArg := args[1]
		if offsetArg == tree.DNull {
			__antithesis_instrumentation__.Notify(602627)
			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(602628)
		}
		__antithesis_instrumentation__.Notify(602624)
		offset = int(tree.MustBeDInt(offsetArg))
	} else {
		__antithesis_instrumentation__.Notify(602629)
	}
	__antithesis_instrumentation__.Notify(602618)
	if !w.forward {
		__antithesis_instrumentation__.Notify(602630)
		offset *= -1
	} else {
		__antithesis_instrumentation__.Notify(602631)
	}
	__antithesis_instrumentation__.Notify(602619)

	if targetRow := wfr.RowIdx + offset; targetRow < 0 || func() bool {
		__antithesis_instrumentation__.Notify(602632)
		return targetRow >= wfr.PartitionSize() == true
	}() == true {
		__antithesis_instrumentation__.Notify(602633)

		if w.withDefault {
			__antithesis_instrumentation__.Notify(602635)
			args, err := wfr.Args(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(602637)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602638)
			}
			__antithesis_instrumentation__.Notify(602636)
			return args[2], nil
		} else {
			__antithesis_instrumentation__.Notify(602639)
		}
		__antithesis_instrumentation__.Notify(602634)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602640)
	}
	__antithesis_instrumentation__.Notify(602620)

	args, err := wfr.ArgsWithRowOffset(ctx, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(602641)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602642)
	}
	__antithesis_instrumentation__.Notify(602621)
	return args[0], nil
}

func (w *leadLagWindow) Reset(context.Context) { __antithesis_instrumentation__.Notify(602643) }

func (w *leadLagWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602644)
}

type firstValueWindow struct{}

func newFirstValueWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602645)
	return &firstValueWindow{}
}

func (firstValueWindow) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602646)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602650)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602651)
	}
	__antithesis_instrumentation__.Notify(602647)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602652)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602653)
	}
	__antithesis_instrumentation__.Notify(602648)
	for idx := frameStartIdx; idx < frameEndIdx; idx++ {
		__antithesis_instrumentation__.Notify(602654)
		if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
			__antithesis_instrumentation__.Notify(602655)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602656)
			if !skipped {
				__antithesis_instrumentation__.Notify(602657)
				row, err := wfr.Rows.GetRow(ctx, idx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602659)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602660)
				}
				__antithesis_instrumentation__.Notify(602658)
				return row.GetDatum(int(wfr.ArgsIdxs[0]))
			} else {
				__antithesis_instrumentation__.Notify(602661)
			}
		}
	}
	__antithesis_instrumentation__.Notify(602649)

	return tree.DNull, nil
}

func (firstValueWindow) Reset(context.Context) { __antithesis_instrumentation__.Notify(602662) }

func (firstValueWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602663)
}

type lastValueWindow struct{}

func newLastValueWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602664)
	return &lastValueWindow{}
}

func (lastValueWindow) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602665)
	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602669)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602670)
	}
	__antithesis_instrumentation__.Notify(602666)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602671)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602672)
	}
	__antithesis_instrumentation__.Notify(602667)
	for idx := frameEndIdx - 1; idx >= frameStartIdx; idx-- {
		__antithesis_instrumentation__.Notify(602673)
		if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
			__antithesis_instrumentation__.Notify(602674)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602675)
			if !skipped {
				__antithesis_instrumentation__.Notify(602676)
				row, err := wfr.Rows.GetRow(ctx, idx)
				if err != nil {
					__antithesis_instrumentation__.Notify(602678)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(602679)
				}
				__antithesis_instrumentation__.Notify(602677)
				return row.GetDatum(int(wfr.ArgsIdxs[0]))
			} else {
				__antithesis_instrumentation__.Notify(602680)
			}
		}
	}
	__antithesis_instrumentation__.Notify(602668)

	return tree.DNull, nil
}

func (lastValueWindow) Reset(context.Context) { __antithesis_instrumentation__.Notify(602681) }

func (lastValueWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602682)
}

type nthValueWindow struct{}

func newNthValueWindow([]*types.T, *tree.EvalContext) tree.WindowFunc {
	__antithesis_instrumentation__.Notify(602683)
	return &nthValueWindow{}
}

var ErrInvalidArgumentForNthValue = pgerror.Newf(
	pgcode.InvalidParameterValue, "argument of nth_value() must be greater than zero")

func (nthValueWindow) Compute(
	ctx context.Context, evalCtx *tree.EvalContext, wfr *tree.WindowFrameRun,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602684)
	args, err := wfr.Args(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602693)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602694)
	}
	__antithesis_instrumentation__.Notify(602685)
	arg := args[1]
	if arg == tree.DNull {
		__antithesis_instrumentation__.Notify(602695)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602696)
	}
	__antithesis_instrumentation__.Notify(602686)

	nth := int(tree.MustBeDInt(arg))
	if nth <= 0 {
		__antithesis_instrumentation__.Notify(602697)
		return nil, ErrInvalidArgumentForNthValue
	} else {
		__antithesis_instrumentation__.Notify(602698)
	}
	__antithesis_instrumentation__.Notify(602687)

	frameStartIdx, err := wfr.FrameStartIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602699)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602700)
	}
	__antithesis_instrumentation__.Notify(602688)
	frameEndIdx, err := wfr.FrameEndIdx(ctx, evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602701)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602702)
	}
	__antithesis_instrumentation__.Notify(602689)
	if nth > frameEndIdx-frameStartIdx {
		__antithesis_instrumentation__.Notify(602703)

		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(602704)
	}
	__antithesis_instrumentation__.Notify(602690)
	var idx int

	if wfr.Frame.DefaultFrameExclusion() {
		__antithesis_instrumentation__.Notify(602705)

		idx = frameStartIdx + nth - 1
	} else {
		__antithesis_instrumentation__.Notify(602706)
		ith := 0
		for idx = frameStartIdx; idx < frameEndIdx; idx++ {
			__antithesis_instrumentation__.Notify(602708)
			if skipped, err := wfr.IsRowSkipped(ctx, idx); err != nil {
				__antithesis_instrumentation__.Notify(602709)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(602710)
				if !skipped {
					__antithesis_instrumentation__.Notify(602711)
					ith++
					if ith == nth {
						__antithesis_instrumentation__.Notify(602712)

						break
					} else {
						__antithesis_instrumentation__.Notify(602713)
					}
				} else {
					__antithesis_instrumentation__.Notify(602714)
				}
			}
		}
		__antithesis_instrumentation__.Notify(602707)
		if idx == frameEndIdx {
			__antithesis_instrumentation__.Notify(602715)

			return tree.DNull, nil
		} else {
			__antithesis_instrumentation__.Notify(602716)
		}
	}
	__antithesis_instrumentation__.Notify(602691)
	row, err := wfr.Rows.GetRow(ctx, idx)
	if err != nil {
		__antithesis_instrumentation__.Notify(602717)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602718)
	}
	__antithesis_instrumentation__.Notify(602692)
	return row.GetDatum(int(wfr.ArgsIdxs[0]))
}

func (nthValueWindow) Reset(context.Context) { __antithesis_instrumentation__.Notify(602719) }

func (nthValueWindow) Close(context.Context, *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(602720)
}
