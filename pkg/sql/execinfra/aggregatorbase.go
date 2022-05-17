package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type AggregateConstructor func(*tree.EvalContext, tree.Datums) tree.AggregateFunc

func GetAggregateInfo(
	fn execinfrapb.AggregatorSpec_Func, inputTypes ...*types.T,
) (aggregateConstructor AggregateConstructor, returnType *types.T, err error) {
	__antithesis_instrumentation__.Notify(470828)
	if fn == execinfrapb.AnyNotNull {
		__antithesis_instrumentation__.Notify(470831)

		if len(inputTypes) != 1 {
			__antithesis_instrumentation__.Notify(470833)
			return nil, nil, errors.Errorf("any_not_null aggregate needs 1 input")
		} else {
			__antithesis_instrumentation__.Notify(470834)
		}
		__antithesis_instrumentation__.Notify(470832)
		return builtins.NewAnyNotNullAggregate, inputTypes[0], nil
	} else {
		__antithesis_instrumentation__.Notify(470835)
	}
	__antithesis_instrumentation__.Notify(470829)

	props, builtins := builtins.GetBuiltinProperties(strings.ToLower(fn.String()))
	for _, b := range builtins {
		__antithesis_instrumentation__.Notify(470836)
		typs := b.Types.Types()
		if len(typs) != len(inputTypes) {
			__antithesis_instrumentation__.Notify(470839)
			continue
		} else {
			__antithesis_instrumentation__.Notify(470840)
		}
		__antithesis_instrumentation__.Notify(470837)
		match := true
		for i, t := range typs {
			__antithesis_instrumentation__.Notify(470841)
			if !inputTypes[i].Equivalent(t) {
				__antithesis_instrumentation__.Notify(470842)
				if props.NullableArgs && func() bool {
					__antithesis_instrumentation__.Notify(470844)
					return inputTypes[i].IsAmbiguous() == true
				}() == true {
					__antithesis_instrumentation__.Notify(470845)
					continue
				} else {
					__antithesis_instrumentation__.Notify(470846)
				}
				__antithesis_instrumentation__.Notify(470843)
				match = false
				break
			} else {
				__antithesis_instrumentation__.Notify(470847)
			}
		}
		__antithesis_instrumentation__.Notify(470838)
		if match {
			__antithesis_instrumentation__.Notify(470848)

			constructAgg := func(evalCtx *tree.EvalContext, arguments tree.Datums) tree.AggregateFunc {
				__antithesis_instrumentation__.Notify(470850)
				return b.AggregateFunc(inputTypes, evalCtx, arguments)
			}
			__antithesis_instrumentation__.Notify(470849)
			colTyp := b.InferReturnTypeFromInputArgTypes(inputTypes)
			return constructAgg, colTyp, nil
		} else {
			__antithesis_instrumentation__.Notify(470851)
		}
	}
	__antithesis_instrumentation__.Notify(470830)
	return nil, nil, errors.Errorf(
		"no builtin aggregate for %s on %+v", fn, inputTypes,
	)
}

func GetAggregateConstructor(
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	aggInfo *execinfrapb.AggregatorSpec_Aggregation,
	inputTypes []*types.T,
) (constructor AggregateConstructor, arguments tree.Datums, outputType *types.T, err error) {
	__antithesis_instrumentation__.Notify(470852)
	argTypes := make([]*types.T, len(aggInfo.ColIdx)+len(aggInfo.Arguments))
	for j, c := range aggInfo.ColIdx {
		__antithesis_instrumentation__.Notify(470855)
		if c >= uint32(len(inputTypes)) {
			__antithesis_instrumentation__.Notify(470857)
			err = errors.Errorf("ColIdx out of range (%d)", aggInfo.ColIdx)
			return
		} else {
			__antithesis_instrumentation__.Notify(470858)
		}
		__antithesis_instrumentation__.Notify(470856)
		argTypes[j] = inputTypes[c]
	}
	__antithesis_instrumentation__.Notify(470853)
	arguments = make(tree.Datums, len(aggInfo.Arguments))
	var d tree.Datum
	for j, argument := range aggInfo.Arguments {
		__antithesis_instrumentation__.Notify(470859)
		h := execinfrapb.ExprHelper{}

		if err = h.Init(argument, nil, semaCtx, evalCtx); err != nil {
			__antithesis_instrumentation__.Notify(470862)
			err = errors.Wrapf(err, "%s", argument)
			return
		} else {
			__antithesis_instrumentation__.Notify(470863)
		}
		__antithesis_instrumentation__.Notify(470860)
		d, err = h.Eval(nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(470864)
			err = errors.Wrapf(err, "%s", argument)
			return
		} else {
			__antithesis_instrumentation__.Notify(470865)
		}
		__antithesis_instrumentation__.Notify(470861)
		argTypes[len(aggInfo.ColIdx)+j] = d.ResolvedType()
		arguments[j] = d
	}
	__antithesis_instrumentation__.Notify(470854)
	constructor, outputType, err = GetAggregateInfo(aggInfo.Func, argTypes...)
	return
}

func GetWindowFunctionInfo(
	fn execinfrapb.WindowerSpec_Func, inputTypes ...*types.T,
) (windowConstructor func(*tree.EvalContext) tree.WindowFunc, returnType *types.T, err error) {
	__antithesis_instrumentation__.Notify(470866)
	if fn.AggregateFunc != nil && func() bool {
		__antithesis_instrumentation__.Notify(470870)
		return *fn.AggregateFunc == execinfrapb.AnyNotNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(470871)

		if len(inputTypes) != 1 {
			__antithesis_instrumentation__.Notify(470873)
			return nil, nil, errors.Errorf("any_not_null aggregate needs 1 input")
		} else {
			__antithesis_instrumentation__.Notify(470874)
		}
		__antithesis_instrumentation__.Notify(470872)
		return builtins.NewAggregateWindowFunc(builtins.NewAnyNotNullAggregate), inputTypes[0], nil
	} else {
		__antithesis_instrumentation__.Notify(470875)
	}
	__antithesis_instrumentation__.Notify(470867)

	var funcStr string
	if fn.AggregateFunc != nil {
		__antithesis_instrumentation__.Notify(470876)
		funcStr = fn.AggregateFunc.String()
	} else {
		__antithesis_instrumentation__.Notify(470877)
		if fn.WindowFunc != nil {
			__antithesis_instrumentation__.Notify(470878)
			funcStr = fn.WindowFunc.String()
		} else {
			__antithesis_instrumentation__.Notify(470879)
			return nil, nil, errors.Errorf(
				"function is neither an aggregate nor a window function",
			)
		}
	}
	__antithesis_instrumentation__.Notify(470868)
	props, builtins := builtins.GetBuiltinProperties(strings.ToLower(funcStr))
	for _, b := range builtins {
		__antithesis_instrumentation__.Notify(470880)
		typs := b.Types.Types()
		if len(typs) != len(inputTypes) {
			__antithesis_instrumentation__.Notify(470883)
			continue
		} else {
			__antithesis_instrumentation__.Notify(470884)
		}
		__antithesis_instrumentation__.Notify(470881)
		match := true
		for i, t := range typs {
			__antithesis_instrumentation__.Notify(470885)
			if !inputTypes[i].Equivalent(t) {
				__antithesis_instrumentation__.Notify(470886)
				if props.NullableArgs && func() bool {
					__antithesis_instrumentation__.Notify(470888)
					return inputTypes[i].IsAmbiguous() == true
				}() == true {
					__antithesis_instrumentation__.Notify(470889)
					continue
				} else {
					__antithesis_instrumentation__.Notify(470890)
				}
				__antithesis_instrumentation__.Notify(470887)
				match = false
				break
			} else {
				__antithesis_instrumentation__.Notify(470891)
			}
		}
		__antithesis_instrumentation__.Notify(470882)
		if match {
			__antithesis_instrumentation__.Notify(470892)

			constructAgg := func(evalCtx *tree.EvalContext) tree.WindowFunc {
				__antithesis_instrumentation__.Notify(470894)
				return b.WindowFunc(inputTypes, evalCtx)
			}
			__antithesis_instrumentation__.Notify(470893)
			colTyp := b.InferReturnTypeFromInputArgTypes(inputTypes)
			return constructAgg, colTyp, nil
		} else {
			__antithesis_instrumentation__.Notify(470895)
		}
	}
	__antithesis_instrumentation__.Notify(470869)
	return nil, nil, errors.Errorf(
		"no builtin aggregate/window function for %s on %v", funcStr, inputTypes,
	)
}
