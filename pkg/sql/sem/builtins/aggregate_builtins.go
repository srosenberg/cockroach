package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"unsafe"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geos"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/arith"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/twpayne/go-geom"
)

func initAggregateBuiltins() {
	__antithesis_instrumentation__.Notify(595748)

	for k, v := range aggregates {
		__antithesis_instrumentation__.Notify(595749)
		if _, exists := builtins[k]; exists {
			__antithesis_instrumentation__.Notify(595753)
			panic("duplicate builtin: " + k)
		} else {
			__antithesis_instrumentation__.Notify(595754)
		}
		__antithesis_instrumentation__.Notify(595750)

		if v.props.Class != tree.AggregateClass {
			__antithesis_instrumentation__.Notify(595755)
			panic(errors.AssertionFailedf("%s: aggregate functions should be marked with the tree.AggregateClass "+
				"function class, found %v", k, v))
		} else {
			__antithesis_instrumentation__.Notify(595756)
		}
		__antithesis_instrumentation__.Notify(595751)
		for _, a := range v.overloads {
			__antithesis_instrumentation__.Notify(595757)
			if a.AggregateFunc == nil {
				__antithesis_instrumentation__.Notify(595759)
				panic(errors.AssertionFailedf("%s: aggregate functions should have tree.AggregateFunc constructors, "+
					"found %v", k, a))
			} else {
				__antithesis_instrumentation__.Notify(595760)
			}
			__antithesis_instrumentation__.Notify(595758)
			if a.WindowFunc == nil {
				__antithesis_instrumentation__.Notify(595761)
				panic(errors.AssertionFailedf("%s: aggregate functions should have tree.WindowFunc constructors, "+
					"found %v", k, a))
			} else {
				__antithesis_instrumentation__.Notify(595762)
			}
		}
		__antithesis_instrumentation__.Notify(595752)

		builtins[k] = v
	}
}

func aggProps() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(595763)
	return tree.FunctionProperties{Class: tree.AggregateClass}
}

func aggPropsNullableArgs() tree.FunctionProperties {
	__antithesis_instrumentation__.Notify(595764)
	f := aggProps()
	f.NullableArgs = true
	return f
}

var allMaxMinAggregateTypes = append(
	[]*types.T{types.AnyCollatedString, types.AnyEnum},
	types.Scalar...,
)

var aggregates = map[string]builtinDefinition{
	"array_agg": setProps(aggPropsNullableArgs(),
		arrayBuiltin(func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(595765)
			return makeAggOverloadWithReturnType(
				[]*types.T{t},
				func(args []tree.TypedExpr) *types.T {
					__antithesis_instrumentation__.Notify(595766)
					if len(args) == 0 {
						__antithesis_instrumentation__.Notify(595768)
						return types.MakeArray(t)
					} else {
						__antithesis_instrumentation__.Notify(595769)
					}
					__antithesis_instrumentation__.Notify(595767)

					return types.MakeArray(args[0].ResolvedType())
				},
				newArrayAggregate,
				"Aggregates the selected values into an array.",
				tree.VolatilityImmutable,
			)
		})),

	"avg": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntAvgAggregate,
			"Calculates the average of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatAvgAggregate,
			"Calculates the average of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalAvgAggregate,
			"Calculates the average of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Interval}, types.Interval, newIntervalAvgAggregate,
			"Calculates the average of the selected values.", tree.VolatilityImmutable),
	),

	"bit_and": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Int, newIntBitAndAggregate,
			"Calculates the bitwise AND of all non-null input values, or null if none.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.VarBit}, types.VarBit, newBitBitAndAggregate,
			"Calculates the bitwise AND of all non-null input values, or null if none.", tree.VolatilityImmutable),
	),

	"bit_or": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Int, newIntBitOrAggregate,
			"Calculates the bitwise OR of all non-null input values, or null if none.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.VarBit}, types.VarBit, newBitBitOrAggregate,
			"Calculates the bitwise OR of all non-null input values, or null if none.", tree.VolatilityImmutable),
	),

	"bool_and": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Bool}, types.Bool, newBoolAndAggregate,
			"Calculates the boolean value of `AND`ing all selected values.", tree.VolatilityImmutable),
	),

	"bool_or": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Bool}, types.Bool, newBoolOrAggregate,
			"Calculates the boolean value of `OR`ing all selected values.", tree.VolatilityImmutable),
	),

	"concat_agg": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.String}, types.String, newStringConcatAggregate,
			"Concatenates all selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Bytes}, types.Bytes, newBytesConcatAggregate,
			"Concatenates all selected values.", tree.VolatilityImmutable),
	),

	"corr": makeRegressionAggregateBuiltin(
		newCorrAggregate,
		"Calculates the correlation coefficient of the selected values.",
	),

	"covar_pop": makeRegressionAggregateBuiltin(
		newCovarPopAggregate,
		"Calculates the population covariance of the selected values.",
	),

	"final_covar_pop": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalCovarPopAggregate,
			"Calculates the population covariance of the selected values in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_regr_sxx": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegrSXXAggregate,
			"Calculates sum of squares of the independent variable in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_regr_sxy": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegrSXYAggregate,
			"Calculates sum of products of independent times dependent variable in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_regr_syy": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegrSYYAggregate,
			"Calculates sum of squares of the dependent variable in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_regr_avgx": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegressionAvgXAggregate,
			"Calculates the average of the independent variable (sum(X)/N) in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_regr_avgy": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegressionAvgYAggregate,
			"Calculates the average of the dependent variable (sum(Y)/N) in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_regr_intercept": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegressionInterceptAggregate,
			"Calculates y-intercept of the least-squares-fit linear equation determined by the (X, Y) pairs in final stage.",
			tree.VolatilityImmutable,
		)),
	),

	"final_regr_r2": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegressionR2Aggregate,
			"Calculates square of the correlation coefficient in final stage.",
			tree.VolatilityImmutable,
		)),
	),

	"final_regr_slope": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalRegressionSlopeAggregate,
			"Calculates slope of the least-squares-fit linear equation determined by the (X, Y) pairs in final stage.",
			tree.VolatilityImmutable,
		)),
	),

	"final_corr": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalCorrAggregate,
			"Calculates the correlation coefficient of the selected values in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_covar_samp": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.DecimalArray}, types.Float, newFinalCovarSampAggregate,
			"Calculates the sample covariance of the selected values in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"final_sqrdiff": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Decimal, types.Decimal, types.Int},
			types.Decimal,
			newDecimalFinalSqrdiffAggregate,
			"Calculates the sum of squared differences from the mean of the selected values in final stage.",
			tree.VolatilityImmutable,
		),
		makeAggOverload(
			[]*types.T{types.Float, types.Float, types.Int},
			types.Float,
			newFloatFinalSqrdiffAggregate,
			"Calculates the sum of squared differences from the mean of the selected values in final stage.",
			tree.VolatilityImmutable,
		),
	)),

	"transition_regression_aggregate": makePrivate(makeTransitionRegressionAggregateBuiltin()),

	"covar_samp": makeRegressionAggregateBuiltin(
		newCovarSampAggregate,
		"Calculates the sample covariance of the selected values.",
	),

	"regr_avgx": makeRegressionAggregateBuiltin(
		newRegressionAvgXAggregate, "Calculates the average of the independent variable (sum(X)/N).",
	),

	"regr_avgy": makeRegressionAggregateBuiltin(
		newRegressionAvgYAggregate, "Calculates the average of the dependent variable (sum(Y)/N).",
	),

	"regr_intercept": makeRegressionAggregateBuiltin(
		newRegressionInterceptAggregate, "Calculates y-intercept of the least-squares-fit linear equation determined by the (X, Y) pairs.",
	),

	"regr_r2": makeRegressionAggregateBuiltin(
		newRegressionR2Aggregate, "Calculates square of the correlation coefficient.",
	),

	"regr_slope": makeRegressionAggregateBuiltin(
		newRegressionSlopeAggregate, "Calculates slope of the least-squares-fit linear equation determined by the (X, Y) pairs.",
	),

	"regr_sxx": makeRegressionAggregateBuiltin(
		newRegressionSXXAggregate, "Calculates sum of squares of the independent variable.",
	),

	"regr_sxy": makeRegressionAggregateBuiltin(
		newRegressionSXYAggregate, "Calculates sum of products of independent times dependent variable.",
	),

	"regr_syy": makeRegressionAggregateBuiltin(
		newRegressionSYYAggregate, "Calculates sum of squares of the dependent variable.",
	),

	"regr_count": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Float, types.Float}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int, types.Int}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal, types.Decimal}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float, types.Int}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float, types.Decimal}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int, types.Float}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int, types.Decimal}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal, types.Float}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal, types.Int}, types.Int, newRegressionCountAggregate,
			"Calculates number of input rows in which both expressions are nonnull.", tree.VolatilityImmutable),
	),

	"count": makeBuiltin(aggPropsNullableArgs(),
		makeAggOverload([]*types.T{types.Any}, types.Int, newCountAggregate,
			"Calculates the number of selected elements.", tree.VolatilityImmutable),
	),

	"count_rows": makeBuiltin(aggProps(),
		tree.Overload{
			Types:         tree.ArgTypes{},
			ReturnType:    tree.FixedReturnType(types.Int),
			AggregateFunc: newCountRowsAggregate,
			WindowFunc: func(params []*types.T, evalCtx *tree.EvalContext) tree.WindowFunc {
				__antithesis_instrumentation__.Notify(595770)
				return newFramableAggregateWindow(
					newCountRowsAggregate(params, evalCtx, nil),
					func(evalCtx *tree.EvalContext, arguments tree.Datums) tree.AggregateFunc {
						__antithesis_instrumentation__.Notify(595771)
						return newCountRowsAggregate(params, evalCtx, arguments)
					},
				)
			},
			Info:       "Calculates the number of rows.",
			Volatility: tree.VolatilityImmutable,
		},
	),

	"every": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Bool}, types.Bool, newBoolAndAggregate,
			"Calculates the boolean value of `AND`ing all selected values.", tree.VolatilityImmutable),
	),

	"max": collectOverloads(aggProps(), allMaxMinAggregateTypes,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(595772)
			info := "Identifies the maximum selected value."
			vol := tree.VolatilityImmutable
			return makeAggOverloadWithReturnType(
				[]*types.T{t}, tree.IdentityReturnType(0), newMaxAggregate, info, vol,
			)
		}),

	"min": collectOverloads(aggProps(), allMaxMinAggregateTypes,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(595773)
			info := "Identifies the minimum selected value."
			vol := tree.VolatilityImmutable
			return makeAggOverloadWithReturnType(
				[]*types.T{t}, tree.IdentityReturnType(0), newMinAggregate, info, vol,
			)
		}),

	"string_agg": makeBuiltin(aggPropsNullableArgs(),
		makeAggOverload([]*types.T{types.String, types.String}, types.String, newStringConcatAggregate,
			"Concatenates all selected values using the provided delimiter.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Bytes, types.Bytes}, types.Bytes, newBytesConcatAggregate,
			"Concatenates all selected values using the provided delimiter.", tree.VolatilityImmutable),
	),

	"sum_int": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Int, newSmallIntSumAggregate,
			"Calculates the sum of the selected values.", tree.VolatilityImmutable),
	),

	"sum": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntSumAggregate,
			"Calculates the sum of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatSumAggregate,
			"Calculates the sum of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalSumAggregate,
			"Calculates the sum of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Interval}, types.Interval, newIntervalSumAggregate,
			"Calculates the sum of the selected values.", tree.VolatilityImmutable),
	),

	"sqrdiff": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntSqrDiffAggregate,
			"Calculates the sum of squared differences from the mean of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalSqrDiffAggregate,
			"Calculates the sum of squared differences from the mean of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatSqrDiffAggregate,
			"Calculates the sum of squared differences from the mean of the selected values.", tree.VolatilityImmutable),
	),

	"final_variance": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Decimal, types.Decimal, types.Int},
			types.Decimal,
			newDecimalFinalVarianceAggregate,
			"Calculates the variance from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
		makeAggOverload(
			[]*types.T{types.Float, types.Float, types.Int},
			types.Float,
			newFloatFinalVarianceAggregate,
			"Calculates the variance from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
	)),

	"final_var_pop": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Decimal, types.Decimal, types.Int},
			types.Decimal,
			newDecimalFinalVarPopAggregate,
			"Calculates the population variance from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
		makeAggOverload(
			[]*types.T{types.Float, types.Float, types.Int},
			types.Float,
			newFloatFinalVarPopAggregate,
			"Calculates the population variance from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
	)),

	"final_stddev": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Decimal, types.Decimal, types.Int},
			types.Decimal,
			newDecimalFinalStdDevAggregate,
			"Calculates the standard deviation from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
		makeAggOverload(
			[]*types.T{types.Float, types.Float, types.Int},
			types.Float,
			newFloatFinalStdDevAggregate,
			"Calculates the standard deviation from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
	)),

	"final_stddev_pop": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Decimal, types.Decimal, types.Int},
			types.Decimal,
			newDecimalFinalStdDevPopAggregate,
			"Calculates the population standard deviation from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
		makeAggOverload(
			[]*types.T{types.Float, types.Float, types.Int},
			types.Float,
			newFloatFinalStdDevPopAggregate,
			"Calculates the population standard deviation from the selected locally-computed squared difference values.",
			tree.VolatilityImmutable,
		),
	)),

	"variance": makeVarianceBuiltin(),
	"var_samp": makeVarianceBuiltin(),
	"var_pop": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntVarPopAggregate,
			"Calculates the population variance of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalVarPopAggregate,
			"Calculates the population variance of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatVarPopAggregate,
			"Calculates the population variance of the selected values.", tree.VolatilityImmutable),
	),

	"stddev":      makeStdDevBuiltin(),
	"stddev_samp": makeStdDevBuiltin(),
	"stddev_pop": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntStdDevPopAggregate,
			"Calculates the population standard deviation of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalStdDevPopAggregate,
			"Calculates the population standard deviation of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatStdDevPopAggregate,
			"Calculates the population standard deviation of the selected values.", tree.VolatilityImmutable),
	),

	"xor_agg": makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Bytes}, types.Bytes, newBytesXorAggregate,
			"Calculates the bitwise XOR of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int}, types.Int, newIntXorAggregate,
			"Calculates the bitwise XOR of the selected values.", tree.VolatilityImmutable),
	),

	"json_agg": makeBuiltin(aggPropsNullableArgs(),
		makeAggOverload([]*types.T{types.Any}, types.Jsonb, newJSONAggregate,
			"Aggregates values as a JSON or JSONB array.", tree.VolatilityStable),
	),

	"jsonb_agg": makeBuiltin(aggPropsNullableArgs(),
		makeAggOverload([]*types.T{types.Any}, types.Jsonb, newJSONAggregate,
			"Aggregates values as a JSON or JSONB array.", tree.VolatilityStable),
	),

	"json_object_agg": makeBuiltin(aggPropsNullableArgs(),
		makeAggOverload([]*types.T{types.String, types.Any}, types.Jsonb, newJSONObjectAggregate,
			"Aggregates values as a JSON or JSONB object.", tree.VolatilityStable),
	),
	"jsonb_object_agg": makeBuiltin(aggPropsNullableArgs(),
		makeAggOverload([]*types.T{types.String, types.Any}, types.Jsonb, newJSONObjectAggregate,
			"Aggregates values as a JSON or JSONB object.", tree.VolatilityStable),
	),

	"st_makeline": makeBuiltin(
		tree.FunctionProperties{
			Class:                   tree.AggregateClass,
			NullableArgs:            true,
			AvailableOnPublicSchema: true,
		},
		makeAggOverload(
			[]*types.T{types.Geometry},
			types.Geometry,
			func(
				params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
			) tree.AggregateFunc {
				__antithesis_instrumentation__.Notify(595774)
				return &stMakeLineAgg{
					acc: evalCtx.Mon.MakeBoundAccount(),
				}
			},
			infoBuilder{
				info: "Forms a LineString from Point, MultiPoint or LineStrings. Other shapes will be ignored.",
			}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_extent": makeBuiltin(
		tree.FunctionProperties{
			Class:                   tree.AggregateClass,
			NullableArgs:            true,
			AvailableOnPublicSchema: true,
		},
		makeAggOverload(
			[]*types.T{types.Geometry},
			types.Box2D,
			func(
				params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
			) tree.AggregateFunc {
				__antithesis_instrumentation__.Notify(595775)
				return &stExtentAgg{}
			},
			infoBuilder{
				info: "Forms a Box2D that encapsulates all provided geometries.",
			}.String(),
			tree.VolatilityImmutable,
		),
	),
	"st_union":      makeSTUnionBuiltin(),
	"st_memunion":   makeSTUnionBuiltin(),
	"st_collect":    makeSTCollectBuiltin(),
	"st_memcollect": makeSTCollectBuiltin(),

	AnyNotNull: makePrivate(makeBuiltin(aggProps(),
		makeAggOverloadWithReturnType(
			[]*types.T{types.Any},
			tree.IdentityReturnType(0),
			newAnyNotNullAggregate,
			"Returns an arbitrary not-NULL value, or NULL if none exists.",
			tree.VolatilityImmutable,
		))),

	"percentile_disc": makeBuiltin(aggProps(),
		makeAggOverloadWithReturnType(
			[]*types.T{types.Float},
			func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(595776)
				return tree.UnknownReturnType
			},
			builtinMustNotRun,
			"Discrete percentile: returns the first input value whose position in the ordering equals or "+
				"exceeds the specified fraction.",
			tree.VolatilityImmutable),
		makeAggOverloadWithReturnType(
			[]*types.T{types.FloatArray},
			func(args []tree.TypedExpr) *types.T {
				__antithesis_instrumentation__.Notify(595777)
				return tree.UnknownReturnType
			},
			builtinMustNotRun,
			"Discrete percentile: returns input values whose position in the ordering equals or "+
				"exceeds the specified fractions.",
			tree.VolatilityImmutable),
	),
	"percentile_disc_impl": makePrivate(collectOverloads(aggProps(), types.Scalar,
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(595778)
			return makeAggOverload([]*types.T{types.Float, t}, t, newPercentileDiscAggregate,
				"Implementation of percentile_disc.",
				tree.VolatilityImmutable)
		},
		func(t *types.T) tree.Overload {
			__antithesis_instrumentation__.Notify(595779)
			return makeAggOverload([]*types.T{types.FloatArray, t}, types.MakeArray(t), newPercentileDiscAggregate,
				"Implementation of percentile_disc.",
				tree.VolatilityImmutable)
		},
	)),
	"percentile_cont": makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Float},
			types.Float,
			builtinMustNotRun,
			"Continuous percentile: returns a float corresponding to the specified fraction in the ordering, "+
				"interpolating between adjacent input floats if needed.",
			tree.VolatilityImmutable),
		makeAggOverload(
			[]*types.T{types.Float},
			types.Interval,
			builtinMustNotRun,
			"Continuous percentile: returns an interval corresponding to the specified fraction in the ordering, "+
				"interpolating between adjacent input intervals if needed.",
			tree.VolatilityImmutable),
		makeAggOverload(
			[]*types.T{types.FloatArray},
			types.MakeArray(types.Float),
			builtinMustNotRun,
			"Continuous percentile: returns floats corresponding to the specified fractions in the ordering, "+
				"interpolating between adjacent input floats if needed.",
			tree.VolatilityImmutable),
		makeAggOverload(
			[]*types.T{types.FloatArray},
			types.MakeArray(types.Interval),
			builtinMustNotRun,
			"Continuous percentile: returns intervals corresponding to the specified fractions in the ordering, "+
				"interpolating between adjacent input intervals if needed.",
			tree.VolatilityImmutable),
	),
	"percentile_cont_impl": makePrivate(makeBuiltin(aggProps(),
		makeAggOverload(
			[]*types.T{types.Float, types.Float},
			types.Float,
			newPercentileContAggregate,
			"Implementation of percentile_cont.",
			tree.VolatilityImmutable),
		makeAggOverload(
			[]*types.T{types.Float, types.Interval},
			types.Interval,
			newPercentileContAggregate,
			"Implementation of percentile_cont.",
			tree.VolatilityImmutable),
		makeAggOverload(
			[]*types.T{types.FloatArray, types.Float},
			types.MakeArray(types.Float),
			newPercentileContAggregate,
			"Implementation of percentile_cont.",
			tree.VolatilityImmutable),
		makeAggOverload(
			[]*types.T{types.FloatArray, types.Interval},
			types.MakeArray(types.Interval),
			newPercentileContAggregate,
			"Implementation of percentile_cont.",
			tree.VolatilityImmutable),
	)),
}

const AnyNotNull = "any_not_null"

func makePrivate(b builtinDefinition) builtinDefinition {
	__antithesis_instrumentation__.Notify(595780)
	b.props.Private = true
	return b
}

func makeAggOverload(
	in []*types.T,
	ret *types.T,
	f func([]*types.T, *tree.EvalContext, tree.Datums) tree.AggregateFunc,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(595781)
	return makeAggOverloadWithReturnType(
		in,
		tree.FixedReturnType(ret),
		f,
		info,
		volatility,
	)
}

func makeAggOverloadWithReturnType(
	in []*types.T,
	retType tree.ReturnTyper,
	f func([]*types.T, *tree.EvalContext, tree.Datums) tree.AggregateFunc,
	info string,
	volatility tree.Volatility,
) tree.Overload {
	__antithesis_instrumentation__.Notify(595782)
	argTypes := make(tree.ArgTypes, len(in))
	for i, typ := range in {
		__antithesis_instrumentation__.Notify(595784)
		argTypes[i].Name = fmt.Sprintf("arg%d", i+1)
		argTypes[i].Typ = typ
	}
	__antithesis_instrumentation__.Notify(595783)

	return tree.Overload{

		Types:         argTypes,
		ReturnType:    retType,
		AggregateFunc: f,
		WindowFunc: func(params []*types.T, evalCtx *tree.EvalContext) tree.WindowFunc {
			__antithesis_instrumentation__.Notify(595785)
			aggWindowFunc := f(params, evalCtx, nil)
			switch w := aggWindowFunc.(type) {
			case *minAggregate:
				__antithesis_instrumentation__.Notify(595787)
				min := &slidingWindowFunc{}
				min.sw = makeSlidingWindow(evalCtx, func(evalCtx *tree.EvalContext, a, b tree.Datum) int {
					__antithesis_instrumentation__.Notify(595796)
					return -a.Compare(evalCtx, b)
				})
				__antithesis_instrumentation__.Notify(595788)
				return min
			case *maxAggregate:
				__antithesis_instrumentation__.Notify(595789)
				max := &slidingWindowFunc{}
				max.sw = makeSlidingWindow(evalCtx, func(evalCtx *tree.EvalContext, a, b tree.Datum) int {
					__antithesis_instrumentation__.Notify(595797)
					return a.Compare(evalCtx, b)
				})
				__antithesis_instrumentation__.Notify(595790)
				return max
			case *intSumAggregate:
				__antithesis_instrumentation__.Notify(595791)
				return newSlidingWindowSumFunc(aggWindowFunc)
			case *decimalSumAggregate:
				__antithesis_instrumentation__.Notify(595792)
				return newSlidingWindowSumFunc(aggWindowFunc)
			case *floatSumAggregate:
				__antithesis_instrumentation__.Notify(595793)
				return newSlidingWindowSumFunc(aggWindowFunc)
			case *intervalSumAggregate:
				__antithesis_instrumentation__.Notify(595794)
				return newSlidingWindowSumFunc(aggWindowFunc)
			case *avgAggregate:
				__antithesis_instrumentation__.Notify(595795)

				return &avgWindowFunc{sum: newSlidingWindowSumFunc(w.agg)}
			}
			__antithesis_instrumentation__.Notify(595786)

			return newFramableAggregateWindow(
				aggWindowFunc,
				func(evalCtx *tree.EvalContext, arguments tree.Datums) tree.AggregateFunc {
					__antithesis_instrumentation__.Notify(595798)
					return f(params, evalCtx, arguments)
				},
			)
		},
		Info:       info,
		Volatility: volatility,
	}
}

func makeStdDevBuiltin() builtinDefinition {
	__antithesis_instrumentation__.Notify(595799)
	return makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntStdDevAggregate,
			"Calculates the standard deviation of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalStdDevAggregate,
			"Calculates the standard deviation of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatStdDevAggregate,
			"Calculates the standard deviation of the selected values.", tree.VolatilityImmutable),
	)
}

func makeSTCollectBuiltin() builtinDefinition {
	__antithesis_instrumentation__.Notify(595800)
	return makeBuiltin(
		tree.FunctionProperties{
			Class:                   tree.AggregateClass,
			NullableArgs:            true,
			AvailableOnPublicSchema: true,
		},
		makeAggOverload(
			[]*types.T{types.Geometry},
			types.Geometry,
			newSTCollectAgg,
			infoBuilder{
				info: "Collects geometries into a GeometryCollection or multi-type as appropriate.",
			}.String(),
			tree.VolatilityImmutable,
		),
	)
}

func makeSTUnionBuiltin() builtinDefinition {
	__antithesis_instrumentation__.Notify(595801)
	return makeBuiltin(
		tree.FunctionProperties{
			Class:                   tree.AggregateClass,
			NullableArgs:            true,
			AvailableOnPublicSchema: true,
		},
		makeAggOverload(
			[]*types.T{types.Geometry},
			types.Geometry,
			func(
				params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
			) tree.AggregateFunc {
				__antithesis_instrumentation__.Notify(595802)
				return &stUnionAgg{
					acc: evalCtx.Mon.MakeBoundAccount(),
				}
			},
			infoBuilder{
				info: "Applies a spatial union to the geometries provided.",
			}.String(),
			tree.VolatilityImmutable,
		),
	)
}

func makeRegressionAggregateBuiltin(
	aggregateFunc func([]*types.T, *tree.EvalContext, tree.Datums) tree.AggregateFunc, info string,
) builtinDefinition {
	__antithesis_instrumentation__.Notify(595803)
	return makeRegressionAggregate(aggregateFunc, info, types.Float)
}

func makeTransitionRegressionAggregateBuiltin() builtinDefinition {
	__antithesis_instrumentation__.Notify(595804)
	return makeRegressionAggregate(
		makeTransitionRegressionAccumulatorDecimalBase,
		"Calculates transition values for regression functions in local stage.",
		types.DecimalArray,
	)
}

func makeRegressionAggregate(
	aggregateFunc func([]*types.T, *tree.EvalContext, tree.Datums) tree.AggregateFunc,
	info string,
	ret *types.T,
) builtinDefinition {
	__antithesis_instrumentation__.Notify(595805)
	return makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Float, types.Float}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int, types.Int}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal, types.Decimal}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float, types.Int}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float, types.Decimal}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int, types.Float}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Int, types.Decimal}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal, types.Float}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal, types.Int}, ret, aggregateFunc,
			info, tree.VolatilityImmutable),
	)
}

type stMakeLineAgg struct {
	flatCoords []float64
	layout     geom.Layout
	acc        mon.BoundAccount
}

func (agg *stMakeLineAgg) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(595806)
	if firstArg == tree.DNull {
		__antithesis_instrumentation__.Notify(595811)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(595812)
	}
	__antithesis_instrumentation__.Notify(595807)
	geomArg := tree.MustBeDGeometry(firstArg)

	g, err := geomArg.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(595813)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595814)
	}
	__antithesis_instrumentation__.Notify(595808)

	if len(agg.flatCoords) == 0 {
		__antithesis_instrumentation__.Notify(595815)
		agg.layout = g.Layout()
	} else {
		__antithesis_instrumentation__.Notify(595816)
		if agg.layout != g.Layout() {
			__antithesis_instrumentation__.Notify(595817)
			return errors.Newf(
				"mixed dimensionality not allowed (adding dimension %s to dimension %s)",
				g.Layout(),
				agg.layout,
			)
		} else {
			__antithesis_instrumentation__.Notify(595818)
		}
	}
	__antithesis_instrumentation__.Notify(595809)
	switch g.(type) {
	case *geom.Point, *geom.LineString, *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(595819)
		if err := agg.acc.Grow(ctx, int64(len(g.FlatCoords())*8)); err != nil {
			__antithesis_instrumentation__.Notify(595821)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595822)
		}
		__antithesis_instrumentation__.Notify(595820)
		agg.flatCoords = append(agg.flatCoords, g.FlatCoords()...)
	}
	__antithesis_instrumentation__.Notify(595810)
	return nil
}

func (agg *stMakeLineAgg) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(595823)
	if len(agg.flatCoords) == 0 {
		__antithesis_instrumentation__.Notify(595826)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(595827)
	}
	__antithesis_instrumentation__.Notify(595824)
	g, err := geo.MakeGeometryFromGeomT(geom.NewLineStringFlat(agg.layout, agg.flatCoords))
	if err != nil {
		__antithesis_instrumentation__.Notify(595828)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595829)
	}
	__antithesis_instrumentation__.Notify(595825)
	return tree.NewDGeometry(g), nil
}

func (agg *stMakeLineAgg) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595830)
	agg.flatCoords = agg.flatCoords[:0]
	agg.acc.Empty(ctx)
}

func (agg *stMakeLineAgg) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595831)
	agg.acc.Close(ctx)
}

func (agg *stMakeLineAgg) Size() int64 {
	__antithesis_instrumentation__.Notify(595832)
	return sizeOfSTMakeLineAggregate
}

type stUnionAgg struct {
	srid geopb.SRID

	ewkb geopb.EWKB
	acc  mon.BoundAccount
	set  bool
}

func (agg *stUnionAgg) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(595833)
	if firstArg == tree.DNull {
		__antithesis_instrumentation__.Notify(595839)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(595840)
	}
	__antithesis_instrumentation__.Notify(595834)
	geomArg := tree.MustBeDGeometry(firstArg)
	if !agg.set {
		__antithesis_instrumentation__.Notify(595841)
		agg.ewkb = geomArg.EWKB()
		agg.set = true
		agg.srid = geomArg.SRID()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(595842)
	}
	__antithesis_instrumentation__.Notify(595835)
	if agg.srid != geomArg.SRID() {
		__antithesis_instrumentation__.Notify(595843)
		c, err := geo.ParseGeometryFromEWKB(agg.ewkb)
		if err != nil {
			__antithesis_instrumentation__.Notify(595845)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595846)
		}
		__antithesis_instrumentation__.Notify(595844)
		return geo.NewMismatchingSRIDsError(geomArg.Geometry.SpatialObject(), c.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(595847)
	}
	__antithesis_instrumentation__.Notify(595836)
	if err := agg.acc.Grow(ctx, int64(len(geomArg.EWKB()))); err != nil {
		__antithesis_instrumentation__.Notify(595848)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595849)
	}
	__antithesis_instrumentation__.Notify(595837)
	var err error

	agg.ewkb, err = geos.Union(agg.ewkb, geomArg.EWKB())
	if err != nil {
		__antithesis_instrumentation__.Notify(595850)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595851)
	}
	__antithesis_instrumentation__.Notify(595838)
	return agg.acc.ResizeTo(ctx, int64(len(agg.ewkb)))
}

func (agg *stUnionAgg) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(595852)
	if !agg.set {
		__antithesis_instrumentation__.Notify(595855)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(595856)
	}
	__antithesis_instrumentation__.Notify(595853)
	g, err := geo.ParseGeometryFromEWKB(agg.ewkb)
	if err != nil {
		__antithesis_instrumentation__.Notify(595857)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595858)
	}
	__antithesis_instrumentation__.Notify(595854)
	return tree.NewDGeometry(g), nil
}

func (agg *stUnionAgg) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595859)
	agg.ewkb = nil
	agg.set = false
	agg.acc.Empty(ctx)
}

func (agg *stUnionAgg) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595860)
	agg.acc.Close(ctx)
}

func (agg *stUnionAgg) Size() int64 {
	__antithesis_instrumentation__.Notify(595861)
	return sizeOfSTUnionAggregate
}

type stCollectAgg struct {
	acc  mon.BoundAccount
	coll geom.T
}

func newSTCollectAgg(_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595862)
	return &stCollectAgg{
		acc: evalCtx.Mon.MakeBoundAccount(),
	}
}

func (agg *stCollectAgg) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(595863)
	if firstArg == tree.DNull {
		__antithesis_instrumentation__.Notify(595871)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(595872)
	}
	__antithesis_instrumentation__.Notify(595864)
	if err := agg.acc.Grow(ctx, int64(firstArg.Size())); err != nil {
		__antithesis_instrumentation__.Notify(595873)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595874)
	}
	__antithesis_instrumentation__.Notify(595865)
	geomArg := tree.MustBeDGeometry(firstArg)
	t, err := geomArg.AsGeomT()
	if err != nil {
		__antithesis_instrumentation__.Notify(595875)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595876)
	}
	__antithesis_instrumentation__.Notify(595866)
	if agg.coll != nil && func() bool {
		__antithesis_instrumentation__.Notify(595877)
		return agg.coll.SRID() != t.SRID() == true
	}() == true {
		__antithesis_instrumentation__.Notify(595878)
		c, err := geo.MakeGeometryFromGeomT(agg.coll)
		if err != nil {
			__antithesis_instrumentation__.Notify(595880)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595881)
		}
		__antithesis_instrumentation__.Notify(595879)
		return geo.NewMismatchingSRIDsError(geomArg.Geometry.SpatialObject(), c.SpatialObject())
	} else {
		__antithesis_instrumentation__.Notify(595882)
	}
	__antithesis_instrumentation__.Notify(595867)

	if gc, ok := agg.coll.(*geom.GeometryCollection); ok {
		__antithesis_instrumentation__.Notify(595883)
		return gc.Push(t)
	} else {
		__antithesis_instrumentation__.Notify(595884)
	}
	__antithesis_instrumentation__.Notify(595868)

	switch t := t.(type) {
	case *geom.Point:
		__antithesis_instrumentation__.Notify(595885)
		if agg.coll == nil {
			__antithesis_instrumentation__.Notify(595891)
			agg.coll = geom.NewMultiPoint(t.Layout()).SetSRID(t.SRID())
		} else {
			__antithesis_instrumentation__.Notify(595892)
		}
		__antithesis_instrumentation__.Notify(595886)
		if multi, ok := agg.coll.(*geom.MultiPoint); ok {
			__antithesis_instrumentation__.Notify(595893)
			return multi.Push(t)
		} else {
			__antithesis_instrumentation__.Notify(595894)
		}
	case *geom.LineString:
		__antithesis_instrumentation__.Notify(595887)
		if agg.coll == nil {
			__antithesis_instrumentation__.Notify(595895)
			agg.coll = geom.NewMultiLineString(t.Layout()).SetSRID(t.SRID())
		} else {
			__antithesis_instrumentation__.Notify(595896)
		}
		__antithesis_instrumentation__.Notify(595888)
		if multi, ok := agg.coll.(*geom.MultiLineString); ok {
			__antithesis_instrumentation__.Notify(595897)
			return multi.Push(t)
		} else {
			__antithesis_instrumentation__.Notify(595898)
		}
	case *geom.Polygon:
		__antithesis_instrumentation__.Notify(595889)
		if agg.coll == nil {
			__antithesis_instrumentation__.Notify(595899)
			agg.coll = geom.NewMultiPolygon(t.Layout()).SetSRID(t.SRID())
		} else {
			__antithesis_instrumentation__.Notify(595900)
		}
		__antithesis_instrumentation__.Notify(595890)
		if multi, ok := agg.coll.(*geom.MultiPolygon); ok {
			__antithesis_instrumentation__.Notify(595901)
			return multi.Push(t)
		} else {
			__antithesis_instrumentation__.Notify(595902)
		}
	}
	__antithesis_instrumentation__.Notify(595869)

	var gc *geom.GeometryCollection
	if agg.coll != nil {
		__antithesis_instrumentation__.Notify(595903)

		usedMem := agg.acc.Used()
		if err := agg.acc.Grow(ctx, usedMem); err != nil {
			__antithesis_instrumentation__.Notify(595906)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595907)
		}
		__antithesis_instrumentation__.Notify(595904)
		gc, err = agg.multiToCollection(agg.coll)
		if err != nil {
			__antithesis_instrumentation__.Notify(595908)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595909)
		}
		__antithesis_instrumentation__.Notify(595905)
		agg.coll = nil
		agg.acc.Shrink(ctx, usedMem)
	} else {
		__antithesis_instrumentation__.Notify(595910)
		gc = geom.NewGeometryCollection().SetSRID(t.SRID())
	}
	__antithesis_instrumentation__.Notify(595870)
	agg.coll = gc
	return gc.Push(t)
}

func (agg *stCollectAgg) multiToCollection(multi geom.T) (*geom.GeometryCollection, error) {
	__antithesis_instrumentation__.Notify(595911)
	gc := geom.NewGeometryCollection().SetSRID(multi.SRID())
	switch t := multi.(type) {
	case *geom.MultiPoint:
		__antithesis_instrumentation__.Notify(595913)
		for i := 0; i < t.NumPoints(); i++ {
			__antithesis_instrumentation__.Notify(595917)
			if err := gc.Push(t.Point(i)); err != nil {
				__antithesis_instrumentation__.Notify(595918)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(595919)
			}
		}
	case *geom.MultiLineString:
		__antithesis_instrumentation__.Notify(595914)
		for i := 0; i < t.NumLineStrings(); i++ {
			__antithesis_instrumentation__.Notify(595920)
			if err := gc.Push(t.LineString(i)); err != nil {
				__antithesis_instrumentation__.Notify(595921)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(595922)
			}
		}
	case *geom.MultiPolygon:
		__antithesis_instrumentation__.Notify(595915)
		for i := 0; i < t.NumPolygons(); i++ {
			__antithesis_instrumentation__.Notify(595923)
			if err := gc.Push(t.Polygon(i)); err != nil {
				__antithesis_instrumentation__.Notify(595924)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(595925)
			}
		}
	default:
		__antithesis_instrumentation__.Notify(595916)
		return nil, errors.AssertionFailedf("unexpected geometry type: %T", t)
	}
	__antithesis_instrumentation__.Notify(595912)
	return gc, nil
}

func (agg *stCollectAgg) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(595926)
	if agg.coll == nil {
		__antithesis_instrumentation__.Notify(595929)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(595930)
	}
	__antithesis_instrumentation__.Notify(595927)
	g, err := geo.MakeGeometryFromGeomT(agg.coll)
	if err != nil {
		__antithesis_instrumentation__.Notify(595931)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(595932)
	}
	__antithesis_instrumentation__.Notify(595928)
	return tree.NewDGeometry(g), nil
}

func (agg *stCollectAgg) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595933)
	agg.coll = nil
	agg.acc.Empty(ctx)
}

func (agg *stCollectAgg) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595934)
	agg.acc.Close(ctx)
}

func (agg *stCollectAgg) Size() int64 {
	__antithesis_instrumentation__.Notify(595935)
	return sizeOfSTCollectAggregate
}

type stExtentAgg struct {
	bbox *geo.CartesianBoundingBox
}

func (agg *stExtentAgg) Add(_ context.Context, firstArg tree.Datum, otherArgs ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(595936)
	if firstArg == tree.DNull {
		__antithesis_instrumentation__.Notify(595939)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(595940)
	}
	__antithesis_instrumentation__.Notify(595937)
	geomArg := tree.MustBeDGeometry(firstArg)
	if geomArg.Empty() {
		__antithesis_instrumentation__.Notify(595941)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(595942)
	}
	__antithesis_instrumentation__.Notify(595938)
	b := geomArg.CartesianBoundingBox()
	agg.bbox = agg.bbox.WithPoint(b.LoX, b.LoY).WithPoint(b.HiX, b.HiY)
	return nil
}

func (agg *stExtentAgg) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(595943)
	if agg.bbox == nil {
		__antithesis_instrumentation__.Notify(595945)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(595946)
	}
	__antithesis_instrumentation__.Notify(595944)
	return tree.NewDBox2D(*agg.bbox), nil
}

func (agg *stExtentAgg) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(595947)
	agg.bbox = nil
}

func (agg *stExtentAgg) Close(context.Context) { __antithesis_instrumentation__.Notify(595948) }

func (agg *stExtentAgg) Size() int64 {
	__antithesis_instrumentation__.Notify(595949)
	return sizeOfSTExtentAggregate
}

func makeVarianceBuiltin() builtinDefinition {
	__antithesis_instrumentation__.Notify(595950)
	return makeBuiltin(aggProps(),
		makeAggOverload([]*types.T{types.Int}, types.Decimal, newIntVarianceAggregate,
			"Calculates the variance of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Decimal}, types.Decimal, newDecimalVarianceAggregate,
			"Calculates the variance of the selected values.", tree.VolatilityImmutable),
		makeAggOverload([]*types.T{types.Float}, types.Float, newFloatVarianceAggregate,
			"Calculates the variance of the selected values.", tree.VolatilityImmutable),
	)
}

func builtinMustNotRun(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595951)
	panic("builtin must be overridden and cannot be run directly")
}

var _ tree.AggregateFunc = &arrayAggregate{}
var _ tree.AggregateFunc = &avgAggregate{}
var _ tree.AggregateFunc = &corrAggregate{}
var _ tree.AggregateFunc = &countAggregate{}
var _ tree.AggregateFunc = &countRowsAggregate{}
var _ tree.AggregateFunc = &maxAggregate{}
var _ tree.AggregateFunc = &minAggregate{}
var _ tree.AggregateFunc = &smallIntSumAggregate{}
var _ tree.AggregateFunc = &intSumAggregate{}
var _ tree.AggregateFunc = &decimalSumAggregate{}
var _ tree.AggregateFunc = &floatSumAggregate{}
var _ tree.AggregateFunc = &intervalSumAggregate{}
var _ tree.AggregateFunc = &intSqrDiffAggregate{}
var _ tree.AggregateFunc = &floatSqrDiffAggregate{}
var _ tree.AggregateFunc = &decimalSqrDiffAggregate{}
var _ tree.AggregateFunc = &floatSumSqrDiffsAggregate{}
var _ tree.AggregateFunc = &decimalSumSqrDiffsAggregate{}
var _ tree.AggregateFunc = &floatVarianceAggregate{}
var _ tree.AggregateFunc = &decimalVarianceAggregate{}
var _ tree.AggregateFunc = &floatStdDevAggregate{}
var _ tree.AggregateFunc = &decimalStdDevAggregate{}
var _ tree.AggregateFunc = &anyNotNullAggregate{}
var _ tree.AggregateFunc = &concatAggregate{}
var _ tree.AggregateFunc = &boolAndAggregate{}
var _ tree.AggregateFunc = &boolOrAggregate{}
var _ tree.AggregateFunc = &bytesXorAggregate{}
var _ tree.AggregateFunc = &intXorAggregate{}
var _ tree.AggregateFunc = &jsonAggregate{}
var _ tree.AggregateFunc = &intBitAndAggregate{}
var _ tree.AggregateFunc = &bitBitAndAggregate{}
var _ tree.AggregateFunc = &intBitOrAggregate{}
var _ tree.AggregateFunc = &bitBitOrAggregate{}
var _ tree.AggregateFunc = &percentileDiscAggregate{}
var _ tree.AggregateFunc = &percentileContAggregate{}
var _ tree.AggregateFunc = &stMakeLineAgg{}
var _ tree.AggregateFunc = &stUnionAgg{}
var _ tree.AggregateFunc = &stExtentAgg{}
var _ tree.AggregateFunc = &regressionAccumulatorDecimalBase{}
var _ tree.AggregateFunc = &finalRegressionAccumulatorDecimalBase{}
var _ tree.AggregateFunc = &covarPopAggregate{}
var _ tree.AggregateFunc = &finalCorrAggregate{}
var _ tree.AggregateFunc = &finalCovarSampAggregate{}
var _ tree.AggregateFunc = &finalCovarPopAggregate{}
var _ tree.AggregateFunc = &finalRegrSXXAggregate{}
var _ tree.AggregateFunc = &finalRegrSXYAggregate{}
var _ tree.AggregateFunc = &finalRegrSYYAggregate{}
var _ tree.AggregateFunc = &finalRegressionAvgXAggregate{}
var _ tree.AggregateFunc = &finalRegressionAvgYAggregate{}
var _ tree.AggregateFunc = &finalRegressionInterceptAggregate{}
var _ tree.AggregateFunc = &finalRegressionR2Aggregate{}
var _ tree.AggregateFunc = &finalRegressionSlopeAggregate{}
var _ tree.AggregateFunc = &covarSampAggregate{}
var _ tree.AggregateFunc = &regressionInterceptAggregate{}
var _ tree.AggregateFunc = &regressionR2Aggregate{}
var _ tree.AggregateFunc = &regressionSlopeAggregate{}
var _ tree.AggregateFunc = &regressionSXXAggregate{}
var _ tree.AggregateFunc = &regressionSXYAggregate{}
var _ tree.AggregateFunc = &regressionSYYAggregate{}
var _ tree.AggregateFunc = &regressionCountAggregate{}
var _ tree.AggregateFunc = &regressionAvgXAggregate{}
var _ tree.AggregateFunc = &regressionAvgYAggregate{}

const sizeOfArrayAggregate = int64(unsafe.Sizeof(arrayAggregate{}))
const sizeOfAvgAggregate = int64(unsafe.Sizeof(avgAggregate{}))
const sizeOfRegressionAccumulatorDecimalBase = int64(unsafe.Sizeof(regressionAccumulatorDecimalBase{}))
const sizeOfFinalRegressionAccumulatorDecimalBase = int64(unsafe.Sizeof(finalRegressionAccumulatorDecimalBase{}))
const sizeOfCountAggregate = int64(unsafe.Sizeof(countAggregate{}))
const sizeOfRegressionCountAggregate = int64(unsafe.Sizeof(regressionCountAggregate{}))
const sizeOfCountRowsAggregate = int64(unsafe.Sizeof(countRowsAggregate{}))
const sizeOfMaxAggregate = int64(unsafe.Sizeof(maxAggregate{}))
const sizeOfMinAggregate = int64(unsafe.Sizeof(minAggregate{}))
const sizeOfSmallIntSumAggregate = int64(unsafe.Sizeof(smallIntSumAggregate{}))
const sizeOfIntSumAggregate = int64(unsafe.Sizeof(intSumAggregate{}))
const sizeOfDecimalSumAggregate = int64(unsafe.Sizeof(decimalSumAggregate{}))
const sizeOfFloatSumAggregate = int64(unsafe.Sizeof(floatSumAggregate{}))
const sizeOfIntervalSumAggregate = int64(unsafe.Sizeof(intervalSumAggregate{}))
const sizeOfIntSqrDiffAggregate = int64(unsafe.Sizeof(intSqrDiffAggregate{}))
const sizeOfFloatSqrDiffAggregate = int64(unsafe.Sizeof(floatSqrDiffAggregate{}))
const sizeOfDecimalSqrDiffAggregate = int64(unsafe.Sizeof(decimalSqrDiffAggregate{}))
const sizeOfFloatSumSqrDiffsAggregate = int64(unsafe.Sizeof(floatSumSqrDiffsAggregate{}))
const sizeOfDecimalSumSqrDiffsAggregate = int64(unsafe.Sizeof(decimalSumSqrDiffsAggregate{}))
const sizeOfFloatVarianceAggregate = int64(unsafe.Sizeof(floatVarianceAggregate{}))
const sizeOfDecimalVarianceAggregate = int64(unsafe.Sizeof(decimalVarianceAggregate{}))
const sizeOfFloatVarPopAggregate = int64(unsafe.Sizeof(floatVarPopAggregate{}))
const sizeOfDecimalVarPopAggregate = int64(unsafe.Sizeof(decimalVarPopAggregate{}))
const sizeOfFloatStdDevAggregate = int64(unsafe.Sizeof(floatStdDevAggregate{}))
const sizeOfDecimalStdDevAggregate = int64(unsafe.Sizeof(decimalStdDevAggregate{}))
const sizeOfAnyNotNullAggregate = int64(unsafe.Sizeof(anyNotNullAggregate{}))
const sizeOfConcatAggregate = int64(unsafe.Sizeof(concatAggregate{}))
const sizeOfBoolAndAggregate = int64(unsafe.Sizeof(boolAndAggregate{}))
const sizeOfBoolOrAggregate = int64(unsafe.Sizeof(boolOrAggregate{}))
const sizeOfBytesXorAggregate = int64(unsafe.Sizeof(bytesXorAggregate{}))
const sizeOfIntXorAggregate = int64(unsafe.Sizeof(intXorAggregate{}))
const sizeOfJSONAggregate = int64(unsafe.Sizeof(jsonAggregate{}))
const sizeOfJSONObjectAggregate = int64(unsafe.Sizeof(jsonObjectAggregate{}))
const sizeOfIntBitAndAggregate = int64(unsafe.Sizeof(intBitAndAggregate{}))
const sizeOfBitBitAndAggregate = int64(unsafe.Sizeof(bitBitAndAggregate{}))
const sizeOfIntBitOrAggregate = int64(unsafe.Sizeof(intBitOrAggregate{}))
const sizeOfBitBitOrAggregate = int64(unsafe.Sizeof(bitBitOrAggregate{}))
const sizeOfPercentileDiscAggregate = int64(unsafe.Sizeof(percentileDiscAggregate{}))
const sizeOfPercentileContAggregate = int64(unsafe.Sizeof(percentileContAggregate{}))
const sizeOfSTMakeLineAggregate = int64(unsafe.Sizeof(stMakeLineAgg{}))
const sizeOfSTUnionAggregate = int64(unsafe.Sizeof(stUnionAgg{}))
const sizeOfSTCollectAggregate = int64(unsafe.Sizeof(stCollectAgg{}))
const sizeOfSTExtentAggregate = int64(unsafe.Sizeof(stExtentAgg{}))

type singleDatumAggregateBase struct {
	mode singleDatumAggregateBaseMode
	acc  *mon.BoundAccount

	accountedFor int64
}

type singleDatumAggregateBaseMode int

const (
	sharedSingleDatumAggregateBaseMode singleDatumAggregateBaseMode = iota

	nonSharedSingleDatumAggregateBaseMode
)

func makeSingleDatumAggregateBase(evalCtx *tree.EvalContext) singleDatumAggregateBase {
	__antithesis_instrumentation__.Notify(595952)
	if evalCtx.SingleDatumAggMemAccount == nil {
		__antithesis_instrumentation__.Notify(595954)
		newAcc := evalCtx.Mon.MakeBoundAccount()
		return singleDatumAggregateBase{
			mode: nonSharedSingleDatumAggregateBaseMode,
			acc:  &newAcc,
		}
	} else {
		__antithesis_instrumentation__.Notify(595955)
	}
	__antithesis_instrumentation__.Notify(595953)
	return singleDatumAggregateBase{
		mode: sharedSingleDatumAggregateBaseMode,
		acc:  evalCtx.SingleDatumAggMemAccount,
	}
}

func (b *singleDatumAggregateBase) updateMemoryUsage(ctx context.Context, newUsage int64) error {
	__antithesis_instrumentation__.Notify(595956)
	if err := b.acc.Grow(ctx, newUsage-b.accountedFor); err != nil {
		__antithesis_instrumentation__.Notify(595958)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595959)
	}
	__antithesis_instrumentation__.Notify(595957)
	b.accountedFor = newUsage
	return nil
}

func (b *singleDatumAggregateBase) reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595960)
	switch b.mode {
	case sharedSingleDatumAggregateBaseMode:
		__antithesis_instrumentation__.Notify(595961)
		b.acc.Shrink(ctx, b.accountedFor)
		b.accountedFor = 0
	case nonSharedSingleDatumAggregateBaseMode:
		__antithesis_instrumentation__.Notify(595962)
		b.acc.Clear(ctx)
	default:
		__antithesis_instrumentation__.Notify(595963)
		panic(errors.Errorf("unexpected singleDatumAggregateBaseMode: %d", b.mode))
	}
}

func (b *singleDatumAggregateBase) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595964)
	switch b.mode {
	case sharedSingleDatumAggregateBaseMode:
		__antithesis_instrumentation__.Notify(595965)
		b.acc.Shrink(ctx, b.accountedFor)
		b.accountedFor = 0
	case nonSharedSingleDatumAggregateBaseMode:
		__antithesis_instrumentation__.Notify(595966)
		b.acc.Close(ctx)
	default:
		__antithesis_instrumentation__.Notify(595967)
		panic(errors.Errorf("unexpected singleDatumAggregateBaseMode: %d", b.mode))
	}
}

type anyNotNullAggregate struct {
	singleDatumAggregateBase

	val tree.Datum
}

func NewAnyNotNullAggregate(evalCtx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595968)
	return &anyNotNullAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		val:                      tree.DNull,
	}
}

func newAnyNotNullAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, datums tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595969)
	return NewAnyNotNullAggregate(evalCtx, datums)
}

func (a *anyNotNullAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(595970)
	if a.val == tree.DNull && func() bool {
		__antithesis_instrumentation__.Notify(595972)
		return datum != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(595973)
		a.val = datum
		if err := a.updateMemoryUsage(ctx, int64(datum.Size())); err != nil {
			__antithesis_instrumentation__.Notify(595974)
			return err
		} else {
			__antithesis_instrumentation__.Notify(595975)
		}
	} else {
		__antithesis_instrumentation__.Notify(595976)
	}
	__antithesis_instrumentation__.Notify(595971)
	return nil
}

func (a *anyNotNullAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(595977)
	return a.val, nil
}

func (a *anyNotNullAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595978)
	a.val = tree.DNull
	a.reset(ctx)
}

func (a *anyNotNullAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595979)
	a.close(ctx)
}

func (a *anyNotNullAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(595980)
	return sizeOfAnyNotNullAggregate
}

type arrayAggregate struct {
	arr *tree.DArray

	acc mon.BoundAccount
}

func newArrayAggregate(
	params []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595981)
	return &arrayAggregate{
		arr: tree.NewDArray(params[0]),
		acc: evalCtx.Mon.MakeBoundAccount(),
	}
}

func (a *arrayAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(595982)
	if err := a.acc.Grow(ctx, int64(datum.Size())); err != nil {
		__antithesis_instrumentation__.Notify(595984)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595985)
	}
	__antithesis_instrumentation__.Notify(595983)
	return a.arr.Append(datum)
}

func (a *arrayAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(595986)
	if len(a.arr.Array) > 0 {
		__antithesis_instrumentation__.Notify(595988)
		arrCopy := *a.arr
		return &arrCopy, nil
	} else {
		__antithesis_instrumentation__.Notify(595989)
	}
	__antithesis_instrumentation__.Notify(595987)
	return tree.DNull, nil
}

func (a *arrayAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595990)
	a.arr = tree.NewDArray(a.arr.ParamTyp)
	a.acc.Empty(ctx)
}

func (a *arrayAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(595991)
	a.acc.Close(ctx)
}

func (a *arrayAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(595992)
	return sizeOfArrayAggregate
}

type avgAggregate struct {
	agg   tree.AggregateFunc
	count int
}

func newIntAvgAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595993)
	return &avgAggregate{agg: newIntSumAggregate(params, evalCtx, arguments)}
}
func newFloatAvgAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595994)
	return &avgAggregate{agg: newFloatSumAggregate(params, evalCtx, arguments)}
}
func newDecimalAvgAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595995)
	return &avgAggregate{agg: newDecimalSumAggregate(params, evalCtx, arguments)}
}
func newIntervalAvgAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(595996)
	return &avgAggregate{agg: newIntervalSumAggregate(params, evalCtx, arguments)}
}

func (a *avgAggregate) Add(ctx context.Context, datum tree.Datum, other ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(595997)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596000)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596001)
	}
	__antithesis_instrumentation__.Notify(595998)
	if err := a.agg.Add(ctx, datum); err != nil {
		__antithesis_instrumentation__.Notify(596002)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596003)
	}
	__antithesis_instrumentation__.Notify(595999)
	a.count++
	return nil
}

func (a *avgAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596004)
	sum, err := a.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(596007)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596008)
	}
	__antithesis_instrumentation__.Notify(596005)
	if sum == tree.DNull {
		__antithesis_instrumentation__.Notify(596009)
		return sum, nil
	} else {
		__antithesis_instrumentation__.Notify(596010)
	}
	__antithesis_instrumentation__.Notify(596006)
	switch t := sum.(type) {
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(596011)
		return tree.NewDFloat(*t / tree.DFloat(a.count)), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(596012)
		count := apd.New(int64(a.count), 0)
		_, err := tree.DecimalCtx.Quo(&t.Decimal, &t.Decimal, count)
		return t, err
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(596013)
		return &tree.DInterval{Duration: t.Duration.Div(int64(a.count))}, nil
	default:
		__antithesis_instrumentation__.Notify(596014)
		return nil, errors.AssertionFailedf("unexpected SUM result type: %s", t)
	}
}

func (a *avgAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596015)
	a.agg.Reset(ctx)
	a.count = 0
}

func (a *avgAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596016)
	a.agg.Close(ctx)
}

func (a *avgAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596017)
	return sizeOfAvgAggregate
}

type concatAggregate struct {
	singleDatumAggregateBase

	forBytes   bool
	sawNonNull bool
	delimiter  string
	result     bytes.Buffer
}

func newBytesConcatAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596018)
	concatAgg := &concatAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		forBytes:                 true,
	}
	if len(arguments) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(596020)
		return arguments[0] != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596021)
		concatAgg.delimiter = string(tree.MustBeDBytes(arguments[0]))
	} else {
		__antithesis_instrumentation__.Notify(596022)
		if len(arguments) > 1 {
			__antithesis_instrumentation__.Notify(596023)
			panic(errors.AssertionFailedf("too many arguments passed in, expected < 2, got %d", len(arguments)))
		} else {
			__antithesis_instrumentation__.Notify(596024)
		}
	}
	__antithesis_instrumentation__.Notify(596019)
	return concatAgg
}

func newStringConcatAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596025)
	concatAgg := &concatAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
	}
	if len(arguments) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(596027)
		return arguments[0] != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596028)
		concatAgg.delimiter = string(tree.MustBeDString(arguments[0]))
	} else {
		__antithesis_instrumentation__.Notify(596029)
		if len(arguments) > 1 {
			__antithesis_instrumentation__.Notify(596030)
			panic(errors.AssertionFailedf("too many arguments passed in, expected < 2, got %d", len(arguments)))
		} else {
			__antithesis_instrumentation__.Notify(596031)
		}
	}
	__antithesis_instrumentation__.Notify(596026)
	return concatAgg
}

func (a *concatAggregate) Add(ctx context.Context, datum tree.Datum, others ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596032)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596037)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596038)
	}
	__antithesis_instrumentation__.Notify(596033)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596039)
		a.sawNonNull = true
	} else {
		__antithesis_instrumentation__.Notify(596040)
		delimiter := a.delimiter

		if len(others) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(596042)
			return others[0] != tree.DNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(596043)
			if a.forBytes {
				__antithesis_instrumentation__.Notify(596044)
				delimiter = string(tree.MustBeDBytes(others[0]))
			} else {
				__antithesis_instrumentation__.Notify(596045)
				delimiter = string(tree.MustBeDString(others[0]))
			}
		} else {
			__antithesis_instrumentation__.Notify(596046)
			if len(others) > 1 {
				__antithesis_instrumentation__.Notify(596047)
				panic(errors.AssertionFailedf("too many other datums passed in, expected < 2, got %d", len(others)))
			} else {
				__antithesis_instrumentation__.Notify(596048)
			}
		}
		__antithesis_instrumentation__.Notify(596041)
		if len(delimiter) > 0 {
			__antithesis_instrumentation__.Notify(596049)
			a.result.WriteString(delimiter)
		} else {
			__antithesis_instrumentation__.Notify(596050)
		}
	}
	__antithesis_instrumentation__.Notify(596034)
	var arg string
	if a.forBytes {
		__antithesis_instrumentation__.Notify(596051)
		arg = string(tree.MustBeDBytes(datum))
	} else {
		__antithesis_instrumentation__.Notify(596052)
		arg = string(tree.MustBeDString(datum))
	}
	__antithesis_instrumentation__.Notify(596035)
	a.result.WriteString(arg)
	if err := a.updateMemoryUsage(ctx, int64(a.result.Cap())); err != nil {
		__antithesis_instrumentation__.Notify(596053)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596054)
	}
	__antithesis_instrumentation__.Notify(596036)
	return nil
}

func (a *concatAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596055)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596058)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596059)
	}
	__antithesis_instrumentation__.Notify(596056)
	if a.forBytes {
		__antithesis_instrumentation__.Notify(596060)
		res := tree.DBytes(a.result.String())
		return &res, nil
	} else {
		__antithesis_instrumentation__.Notify(596061)
	}
	__antithesis_instrumentation__.Notify(596057)
	res := tree.DString(a.result.String())
	return &res, nil
}

func (a *concatAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596062)
	a.sawNonNull = false
	a.result.Reset()

}

func (a *concatAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596063)
	a.close(ctx)
}

func (a *concatAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596064)
	return sizeOfConcatAggregate
}

type intBitAndAggregate struct {
	sawNonNull bool
	result     int64
}

func newIntBitAndAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596065)
	return &intBitAndAggregate{}
}

func (a *intBitAndAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596066)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596069)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596070)
	}
	__antithesis_instrumentation__.Notify(596067)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596071)

		a.result = int64(tree.MustBeDInt(datum))
		a.sawNonNull = true
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596072)
	}
	__antithesis_instrumentation__.Notify(596068)

	a.result = a.result & int64(tree.MustBeDInt(datum))
	return nil
}

func (a *intBitAndAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596073)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596075)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596076)
	}
	__antithesis_instrumentation__.Notify(596074)
	return tree.NewDInt(tree.DInt(a.result)), nil
}

func (a *intBitAndAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596077)
	a.sawNonNull = false
	a.result = 0
}

func (a *intBitAndAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596078) }

func (a *intBitAndAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596079)
	return sizeOfIntBitAndAggregate
}

type bitBitAndAggregate struct {
	sawNonNull bool
	result     bitarray.BitArray
}

func newBitBitAndAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596080)
	return &bitBitAndAggregate{}
}

func (a *bitBitAndAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596081)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596085)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596086)
	}
	__antithesis_instrumentation__.Notify(596082)
	bits := &tree.MustBeDBitArray(datum).BitArray
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596087)

		a.result = *bits
		a.sawNonNull = true
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596088)
	}
	__antithesis_instrumentation__.Notify(596083)

	if a.result.BitLen() != bits.BitLen() {
		__antithesis_instrumentation__.Notify(596089)
		return tree.NewCannotMixBitArraySizesError("AND")
	} else {
		__antithesis_instrumentation__.Notify(596090)
	}
	__antithesis_instrumentation__.Notify(596084)

	a.result = bitarray.And(a.result, *bits)
	return nil
}

func (a *bitBitAndAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596091)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596093)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596094)
	}
	__antithesis_instrumentation__.Notify(596092)
	return &tree.DBitArray{BitArray: a.result}, nil
}

func (a *bitBitAndAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596095)
	a.sawNonNull = false
	a.result = bitarray.BitArray{}
}

func (a *bitBitAndAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596096) }

func (a *bitBitAndAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596097)
	return sizeOfBitBitAndAggregate
}

type intBitOrAggregate struct {
	sawNonNull bool
	result     int64
}

func newIntBitOrAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596098)
	return &intBitOrAggregate{}
}

func (a *intBitOrAggregate) Add(
	_ context.Context, datum tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596099)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596102)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596103)
	}
	__antithesis_instrumentation__.Notify(596100)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596104)

		a.result = int64(tree.MustBeDInt(datum))
		a.sawNonNull = true
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596105)
	}
	__antithesis_instrumentation__.Notify(596101)

	a.result = a.result | int64(tree.MustBeDInt(datum))
	return nil
}

func (a *intBitOrAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596106)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596108)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596109)
	}
	__antithesis_instrumentation__.Notify(596107)
	return tree.NewDInt(tree.DInt(a.result)), nil
}

func (a *intBitOrAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596110)
	a.sawNonNull = false
	a.result = 0
}

func (a *intBitOrAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596111) }

func (a *intBitOrAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596112)
	return sizeOfIntBitOrAggregate
}

type bitBitOrAggregate struct {
	sawNonNull bool
	result     bitarray.BitArray
}

func newBitBitOrAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596113)
	return &bitBitOrAggregate{}
}

func (a *bitBitOrAggregate) Add(
	_ context.Context, datum tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596114)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596118)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596119)
	}
	__antithesis_instrumentation__.Notify(596115)
	bits := &tree.MustBeDBitArray(datum).BitArray
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596120)

		a.result = *bits
		a.sawNonNull = true
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596121)
	}
	__antithesis_instrumentation__.Notify(596116)

	if a.result.BitLen() != bits.BitLen() {
		__antithesis_instrumentation__.Notify(596122)
		return tree.NewCannotMixBitArraySizesError("OR")
	} else {
		__antithesis_instrumentation__.Notify(596123)
	}
	__antithesis_instrumentation__.Notify(596117)

	a.result = bitarray.Or(a.result, *bits)
	return nil
}

func (a *bitBitOrAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596124)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596126)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596127)
	}
	__antithesis_instrumentation__.Notify(596125)
	return &tree.DBitArray{BitArray: a.result}, nil
}

func (a *bitBitOrAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596128)
	a.sawNonNull = false
	a.result = bitarray.BitArray{}
}

func (a *bitBitOrAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596129) }

func (a *bitBitOrAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596130)
	return sizeOfBitBitOrAggregate
}

type boolAndAggregate struct {
	sawNonNull bool
	result     bool
}

func newBoolAndAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596131)
	return &boolAndAggregate{}
}

func (a *boolAndAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596132)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596135)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596136)
	}
	__antithesis_instrumentation__.Notify(596133)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596137)
		a.sawNonNull = true
		a.result = true
	} else {
		__antithesis_instrumentation__.Notify(596138)
	}
	__antithesis_instrumentation__.Notify(596134)
	a.result = a.result && func() bool {
		__antithesis_instrumentation__.Notify(596139)
		return bool(*datum.(*tree.DBool)) == true
	}() == true
	return nil
}

func (a *boolAndAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596140)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596142)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596143)
	}
	__antithesis_instrumentation__.Notify(596141)
	return tree.MakeDBool(tree.DBool(a.result)), nil
}

func (a *boolAndAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596144)
	a.sawNonNull = false
	a.result = false
}

func (a *boolAndAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596145) }

func (a *boolAndAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596146)
	return sizeOfBoolAndAggregate
}

type boolOrAggregate struct {
	sawNonNull bool
	result     bool
}

func newBoolOrAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596147)
	return &boolOrAggregate{}
}

func (a *boolOrAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596148)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596150)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596151)
	}
	__antithesis_instrumentation__.Notify(596149)
	a.sawNonNull = true
	a.result = a.result || func() bool {
		__antithesis_instrumentation__.Notify(596152)
		return bool(*datum.(*tree.DBool)) == true
	}() == true
	return nil
}

func (a *boolOrAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596153)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596155)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596156)
	}
	__antithesis_instrumentation__.Notify(596154)
	return tree.MakeDBool(tree.DBool(a.result)), nil
}

func (a *boolOrAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596157)
	a.sawNonNull = false
	a.result = false
}

func (a *boolOrAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596158) }

func (a *boolOrAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596159)
	return sizeOfBoolOrAggregate
}

type regressionAccumulatorDecimalBase struct {
	singleDatumAggregateBase

	ed                       *apd.ErrDecimal
	n, sx, sxx, sy, syy, sxy apd.Decimal

	tmpX, tmpY, tmpSx, tmpSxx, tmpSy, tmpSyy, tmpSxy apd.Decimal
	scale, tmp, tmpN                                 apd.Decimal
}

func makeRegressionAccumulatorDecimalBase(
	evalCtx *tree.EvalContext,
) regressionAccumulatorDecimalBase {
	__antithesis_instrumentation__.Notify(596160)
	ed := apd.MakeErrDecimal(tree.HighPrecisionCtx)
	return regressionAccumulatorDecimalBase{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		ed:                       &ed,
	}
}

const regrFieldsTotal = 6

func makeTransitionRegressionAccumulatorDecimalBase(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596161)
	ed := apd.MakeErrDecimal(tree.HighPrecisionCtx)
	return &regressionAccumulatorDecimalBase{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		ed:                       &ed,
	}
}

func (a *regressionAccumulatorDecimalBase) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596162)
	res := tree.NewDArray(types.Decimal)
	vals := []*apd.Decimal{&a.n, &a.sx, &a.sxx, &a.sy, &a.syy, &a.sxy}
	for _, v := range vals {
		__antithesis_instrumentation__.Notify(596164)
		dd := &tree.DDecimal{}
		dd.Set(v)
		err := res.Append(dd)
		if err != nil {
			__antithesis_instrumentation__.Notify(596165)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(596166)
		}
	}
	__antithesis_instrumentation__.Notify(596163)
	return res, nil
}

func (a *regressionAccumulatorDecimalBase) Add(
	ctx context.Context, datumY tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596167)
	if datumY == tree.DNull {
		__antithesis_instrumentation__.Notify(596172)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596173)
	}
	__antithesis_instrumentation__.Notify(596168)

	datumX := otherArgs[0]
	if datumX == tree.DNull {
		__antithesis_instrumentation__.Notify(596174)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596175)
	}
	__antithesis_instrumentation__.Notify(596169)
	x, err := a.decimalVal(datumX)
	if err != nil {
		__antithesis_instrumentation__.Notify(596176)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596177)
	}
	__antithesis_instrumentation__.Notify(596170)

	y, err := a.decimalVal(datumY)
	if err != nil {
		__antithesis_instrumentation__.Notify(596178)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596179)
	}
	__antithesis_instrumentation__.Notify(596171)
	return a.add(ctx, y, x)
}

func (a *regressionAccumulatorDecimalBase) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596180)
	*a = regressionAccumulatorDecimalBase{
		singleDatumAggregateBase: a.singleDatumAggregateBase,
		ed:                       a.ed,
	}
	a.reset(ctx)
}

func (a *regressionAccumulatorDecimalBase) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596181)
	a.close(ctx)
}

func (a *regressionAccumulatorDecimalBase) Size() int64 {
	__antithesis_instrumentation__.Notify(596182)
	return sizeOfRegressionAccumulatorDecimalBase
}

func (a *regressionAccumulatorDecimalBase) add(
	ctx context.Context, y *apd.Decimal, x *apd.Decimal,
) error {
	__antithesis_instrumentation__.Notify(596183)
	a.tmpN.Set(&a.n)
	a.tmpSx.Set(&a.sx)
	a.tmpSxx.Set(&a.sxx)
	a.tmpSy.Set(&a.sy)
	a.tmpSyy.Set(&a.syy)
	a.tmpSxy.Set(&a.sxy)

	a.ed.Add(&a.tmpN, &a.tmpN, decimalOne)
	a.ed.Add(&a.tmpSx, &a.tmpSx, x)
	a.ed.Add(&a.tmpSy, &a.tmpSy, y)

	if a.n.Cmp(decimalZero) > 0 {
		__antithesis_instrumentation__.Notify(596186)
		a.ed.Sub(&a.tmpX, a.ed.Mul(&a.tmp, x, &a.tmpN), &a.tmpSx)
		a.ed.Sub(&a.tmpY, a.ed.Mul(&a.tmp, y, &a.tmpN), &a.tmpSy)
		a.ed.Quo(&a.scale, decimalOne, a.ed.Mul(&a.tmp, &a.tmpN, &a.n))
		a.ed.Add(&a.tmpSxx, &a.tmpSxx, a.ed.Mul(&a.tmp, &a.tmpX, a.ed.Mul(&a.tmp, &a.tmpX, &a.scale)))
		a.ed.Add(&a.tmpSyy, &a.tmpSyy, a.ed.Mul(&a.tmp, &a.tmpY, a.ed.Mul(&a.tmp, &a.tmpY, &a.scale)))
		a.ed.Add(&a.tmpSxy, &a.tmpSxy, a.ed.Mul(&a.tmp, &a.tmpX, a.ed.Mul(&a.tmp, &a.tmpY, &a.scale)))

		if isInf(&a.tmpSx) || func() bool {
			__antithesis_instrumentation__.Notify(596187)
			return isInf(&a.tmpSxx) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(596188)
			return isInf(&a.tmpSy) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(596189)
			return isInf(&a.tmpSyy) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(596190)
			return isInf(&a.tmpSxy) == true
		}() == true {
			__antithesis_instrumentation__.Notify(596191)
			if ((isInf(&a.tmpSx) || func() bool {
				__antithesis_instrumentation__.Notify(596195)
				return isInf(&a.tmpSxx) == true
			}() == true) && func() bool {
				__antithesis_instrumentation__.Notify(596196)
				return !isInf(&a.sx) == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(596197)
				return !isInf(x) == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(596198)
				return ((isInf(&a.tmpSy) || func() bool {
					__antithesis_instrumentation__.Notify(596199)
					return isInf(&a.tmpSyy) == true
				}() == true) && func() bool {
					__antithesis_instrumentation__.Notify(596200)
					return !isInf(&a.sy) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(596201)
					return !isInf(y) == true
				}() == true) == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(596202)
				return (isInf(&a.tmpSxy) && func() bool {
					__antithesis_instrumentation__.Notify(596203)
					return !isInf(&a.sx) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(596204)
					return !isInf(x) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(596205)
					return !isInf(&a.sy) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(596206)
					return !isInf(y) == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(596207)
				return tree.ErrFloatOutOfRange
			} else {
				__antithesis_instrumentation__.Notify(596208)
			}
			__antithesis_instrumentation__.Notify(596192)

			if isInf(&a.tmpSxx) {
				__antithesis_instrumentation__.Notify(596209)
				a.tmpSxx = *decimalNaN
			} else {
				__antithesis_instrumentation__.Notify(596210)
			}
			__antithesis_instrumentation__.Notify(596193)
			if isInf(&a.tmpSyy) {
				__antithesis_instrumentation__.Notify(596211)
				a.tmpSyy = *decimalNaN
			} else {
				__antithesis_instrumentation__.Notify(596212)
			}
			__antithesis_instrumentation__.Notify(596194)
			if isInf(&a.tmpSxy) {
				__antithesis_instrumentation__.Notify(596213)
				a.tmpSxy = *decimalNaN
			} else {
				__antithesis_instrumentation__.Notify(596214)
			}
		} else {
			__antithesis_instrumentation__.Notify(596215)
		}
	} else {
		__antithesis_instrumentation__.Notify(596216)

		if isNaN(x) || func() bool {
			__antithesis_instrumentation__.Notify(596218)
			return isInf(x) == true
		}() == true {
			__antithesis_instrumentation__.Notify(596219)
			a.tmpSxx = *decimalNaN
			a.tmpSxy = *decimalNaN
		} else {
			__antithesis_instrumentation__.Notify(596220)
		}
		__antithesis_instrumentation__.Notify(596217)
		if isNaN(y) || func() bool {
			__antithesis_instrumentation__.Notify(596221)
			return isInf(y) == true
		}() == true {
			__antithesis_instrumentation__.Notify(596222)
			a.tmpSyy = *decimalNaN
			a.tmpSxy = *decimalNaN
		} else {
			__antithesis_instrumentation__.Notify(596223)
		}
	}
	__antithesis_instrumentation__.Notify(596184)

	a.n.Set(&a.tmpN)
	a.sx.Set(&a.tmpSx)
	a.sy.Set(&a.tmpSy)
	a.sxx.Set(&a.tmpSxx)
	a.syy.Set(&a.tmpSyy)
	a.sxy.Set(&a.tmpSxy)

	if err := a.updateMemoryUsage(ctx, a.memoryUsage()); err != nil {
		__antithesis_instrumentation__.Notify(596224)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596225)
	}
	__antithesis_instrumentation__.Notify(596185)

	return a.ed.Err()
}

func (a *regressionAccumulatorDecimalBase) memoryUsage() int64 {
	__antithesis_instrumentation__.Notify(596226)
	return int64(a.n.Size() +
		a.sx.Size() +
		a.sxx.Size() +
		a.sy.Size() +
		a.syy.Size() +
		a.sxy.Size() +
		a.tmpX.Size() +
		a.tmpY.Size() +
		a.scale.Size() +
		a.tmpN.Size() +
		a.tmpSx.Size() +
		a.tmpSxx.Size() +
		a.tmpSy.Size() +
		a.tmpSyy.Size() +
		a.tmpSxy.Size())
}

func (a *regressionAccumulatorDecimalBase) covarPopLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596227)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596229)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596230)
	}
	__antithesis_instrumentation__.Notify(596228)

	a.ed.Quo(&a.tmp, &a.sxy, &a.n)
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) corrLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596231)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596234)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596235)
	}
	__antithesis_instrumentation__.Notify(596232)

	if a.sxx.Cmp(decimalZero) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(596236)
		return a.syy.Cmp(decimalZero) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(596237)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596238)
	}
	__antithesis_instrumentation__.Notify(596233)

	a.ed.Quo(&a.tmp, &a.sxy, a.ed.Sqrt(&a.tmp, a.ed.Mul(&a.tmp, &a.sxx, &a.syy)))
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) covarSampLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596239)
	if a.n.Cmp(decimalTwo) < 0 {
		__antithesis_instrumentation__.Notify(596241)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596242)
	}
	__antithesis_instrumentation__.Notify(596240)

	a.ed.Quo(&a.tmp, &a.sxy, a.ed.Sub(&a.tmp, &a.n, decimalOne))
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regrSXXLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596243)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596245)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596246)
	}
	__antithesis_instrumentation__.Notify(596244)
	return mapToDFloat(&a.sxx, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regrSXYLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596247)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596249)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596250)
	}
	__antithesis_instrumentation__.Notify(596248)
	return mapToDFloat(&a.sxy, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regrSYYLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596251)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596253)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596254)
	}
	__antithesis_instrumentation__.Notify(596252)
	return mapToDFloat(&a.syy, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regressionAvgXLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596255)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596257)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596258)
	}
	__antithesis_instrumentation__.Notify(596256)

	a.ed.Quo(&a.tmp, &a.sx, &a.n)
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regressionAvgYLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596259)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596261)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596262)
	}
	__antithesis_instrumentation__.Notify(596260)

	a.ed.Quo(&a.tmp, &a.sy, &a.n)
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regressionInterceptLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596263)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596266)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596267)
	}
	__antithesis_instrumentation__.Notify(596264)
	if a.sxx.Cmp(decimalZero) == 0 {
		__antithesis_instrumentation__.Notify(596268)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596269)
	}
	__antithesis_instrumentation__.Notify(596265)

	a.ed.Quo(
		&a.tmp,
		a.ed.Sub(&a.tmp, &a.sy, a.ed.Mul(&a.tmp, &a.sx, a.ed.Quo(&a.tmp, &a.sxy, &a.sxx))),
		&a.n,
	)
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regressionR2LastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596270)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596274)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596275)
	}
	__antithesis_instrumentation__.Notify(596271)
	if a.sxx.Cmp(decimalZero) == 0 {
		__antithesis_instrumentation__.Notify(596276)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596277)
	}
	__antithesis_instrumentation__.Notify(596272)
	if a.syy.Cmp(decimalZero) == 0 {
		__antithesis_instrumentation__.Notify(596278)
		return tree.NewDFloat(tree.DFloat(1.0)), nil
	} else {
		__antithesis_instrumentation__.Notify(596279)
	}
	__antithesis_instrumentation__.Notify(596273)

	a.ed.Quo(
		&a.tmp,
		a.ed.Mul(&a.tmp, &a.sxy, &a.sxy),
		a.ed.Mul(&a.tmpN, &a.sxx, &a.syy),
	)
	return mapToDFloat(&a.tmp, a.ed.Err())
}

func (a *regressionAccumulatorDecimalBase) regressionSlopeLastStage() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596280)
	if a.n.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596283)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596284)
	}
	__antithesis_instrumentation__.Notify(596281)
	if a.sxx.Cmp(decimalZero) == 0 {
		__antithesis_instrumentation__.Notify(596285)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596286)
	}
	__antithesis_instrumentation__.Notify(596282)

	a.ed.Quo(&a.tmp, &a.sxy, &a.sxx)
	return mapToDFloat(&a.tmp, a.ed.Err())
}

type finalRegressionAccumulatorDecimalBase struct {
	regressionAccumulatorDecimalBase
	otherTransitionValues [regrFieldsTotal]apd.Decimal
}

func (a *finalRegressionAccumulatorDecimalBase) Add(
	ctx context.Context, arrayDatum tree.Datum, _ ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596287)
	if arrayDatum == tree.DNull {
		__antithesis_instrumentation__.Notify(596291)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596292)
	}
	__antithesis_instrumentation__.Notify(596288)

	arr := tree.MustBeDArray(arrayDatum)
	if arr.Len() != regrFieldsTotal {
		__antithesis_instrumentation__.Notify(596293)
		return errors.Newf(
			"regression combine should have %d elements, was %d",
			regrFieldsTotal, arr.Len(),
		)
	} else {
		__antithesis_instrumentation__.Notify(596294)
	}
	__antithesis_instrumentation__.Notify(596289)

	for i, d := range arr.Array {
		__antithesis_instrumentation__.Notify(596295)
		if d == tree.DNull {
			__antithesis_instrumentation__.Notify(596297)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(596298)
		}
		__antithesis_instrumentation__.Notify(596296)
		v := tree.MustBeDDecimal(d)
		a.otherTransitionValues[i].Set(&v.Decimal)
	}
	__antithesis_instrumentation__.Notify(596290)

	return a.combine(ctx)
}

func (a *finalRegressionAccumulatorDecimalBase) combine(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(596299)
	if a.n.Cmp(decimalZero) == 0 {
		__antithesis_instrumentation__.Notify(596307)
		a.n.Set(&a.otherTransitionValues[0])
		a.sx.Set(&a.otherTransitionValues[1])
		a.sxx.Set(&a.otherTransitionValues[2])
		a.sy.Set(&a.otherTransitionValues[3])
		a.syy.Set(&a.otherTransitionValues[4])
		a.sxy.Set(&a.otherTransitionValues[5])
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596308)
		if a.otherTransitionValues[0].Cmp(decimalZero) == 0 {
			__antithesis_instrumentation__.Notify(596309)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(596310)
		}
	}
	__antithesis_instrumentation__.Notify(596300)

	n2 := &a.otherTransitionValues[0]
	sx2 := &a.otherTransitionValues[1]
	sxx2 := &a.otherTransitionValues[2]
	sy2 := &a.otherTransitionValues[3]
	syy2 := &a.otherTransitionValues[4]
	sxy2 := &a.otherTransitionValues[5]

	a.ed.Mul(&a.tmpN, &a.n, a.ed.Quo(&a.tmp, n2, a.ed.Add(&a.tmp, &a.n, n2)))

	a.ed.Add(&a.tmpSx, &a.sx, sx2)
	if isInf(&a.tmpSx) && func() bool {
		__antithesis_instrumentation__.Notify(596311)
		return !isInf(&a.sx) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(596312)
		return !isInf(sx2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(596313)
		return tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596314)
	}
	__antithesis_instrumentation__.Notify(596301)

	a.ed.Sub(&a.tmpX, a.ed.Quo(&a.tmpX, &a.sx, &a.n), a.ed.Quo(&a.tmp, sx2, n2))

	a.ed.Add(&a.tmpSxx, &a.sxx, a.ed.Add(&a.tmp, sxx2, a.ed.Mul(
		&a.tmp, &a.tmpX, a.ed.Mul(&a.tmp, &a.tmpX, &a.tmpN)),
	))
	if isInf(&a.tmpSxx) && func() bool {
		__antithesis_instrumentation__.Notify(596315)
		return !isInf(&a.sxx) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(596316)
		return !isInf(sxx2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(596317)
		return tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596318)
	}
	__antithesis_instrumentation__.Notify(596302)

	a.ed.Add(&a.tmpSy, &a.sy, sy2)
	if isInf(&a.tmpSy) && func() bool {
		__antithesis_instrumentation__.Notify(596319)
		return !isInf(&a.sy) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(596320)
		return !isInf(sy2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(596321)
		return tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596322)
	}
	__antithesis_instrumentation__.Notify(596303)

	a.ed.Sub(&a.tmpY, a.ed.Quo(&a.tmpY, &a.sy, &a.n), a.ed.Quo(&a.tmp, sy2, n2))

	a.ed.Add(&a.tmpSyy, &a.syy, a.ed.Add(&a.tmp, syy2, a.ed.Mul(
		&a.tmp, &a.tmpY, a.ed.Mul(&a.tmp, &a.tmpY, &a.tmpN)),
	))
	if isInf(&a.tmpSyy) && func() bool {
		__antithesis_instrumentation__.Notify(596323)
		return !isInf(&a.syy) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(596324)
		return !isInf(syy2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(596325)
		return tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596326)
	}
	__antithesis_instrumentation__.Notify(596304)

	a.ed.Add(&a.tmpSxy, &a.sxy, a.ed.Add(&a.tmp, sxy2, a.ed.Mul(
		&a.tmp, &a.tmpX, a.ed.Mul(&a.tmp, &a.tmpY, &a.tmpN)),
	))
	if isInf(&a.tmpSxy) && func() bool {
		__antithesis_instrumentation__.Notify(596327)
		return !isInf(&a.sxy) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(596328)
		return !isInf(sxy2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(596329)
		return tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596330)
	}
	__antithesis_instrumentation__.Notify(596305)

	a.ed.Add(&a.n, &a.n, n2)
	a.sx.Set(&a.tmpSx)
	a.sy.Set(&a.tmpSy)
	a.sxx.Set(&a.tmpSxx)
	a.syy.Set(&a.tmpSyy)
	a.sxy.Set(&a.tmpSxy)

	size := a.memoryUsage() +
		int64(a.otherTransitionValues[0].Size()+
			a.otherTransitionValues[1].Size()+
			a.otherTransitionValues[2].Size()+
			a.otherTransitionValues[3].Size()+
			a.otherTransitionValues[4].Size()+
			a.otherTransitionValues[5].Size())

	if err := a.updateMemoryUsage(ctx, size); err != nil {
		__antithesis_instrumentation__.Notify(596331)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596332)
	}
	__antithesis_instrumentation__.Notify(596306)

	return a.ed.Err()
}

func (a *regressionAccumulatorDecimalBase) decimalVal(datum tree.Datum) (*apd.Decimal, error) {
	__antithesis_instrumentation__.Notify(596333)
	res := apd.Decimal{}
	switch val := datum.(type) {
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(596334)
		return res.SetFloat64(float64(*val))
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(596335)
		return res.SetInt64(int64(*val)), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(596336)
		return res.Set(&val.Decimal), nil
	default:
		__antithesis_instrumentation__.Notify(596337)
		return decimalNaN, fmt.Errorf("invalid type %T (%v)", val, val)
	}
}

func mapToDFloat(d *apd.Decimal, err error) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596338)
	if err != nil {
		__antithesis_instrumentation__.Notify(596342)
		return tree.DNull, err
	} else {
		__antithesis_instrumentation__.Notify(596343)
	}
	__antithesis_instrumentation__.Notify(596339)

	res, err := d.Float64()

	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(596344)
		return errors.Is(err, strconv.ErrRange) == true
	}() == true {
		__antithesis_instrumentation__.Notify(596345)
		return tree.DNull, tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596346)
	}
	__antithesis_instrumentation__.Notify(596340)

	if math.IsInf(res, 0) {
		__antithesis_instrumentation__.Notify(596347)
		return tree.DNull, tree.ErrFloatOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596348)
	}
	__antithesis_instrumentation__.Notify(596341)

	return tree.NewDFloat(tree.DFloat(res)), err
}

func isInf(d *apd.Decimal) bool {
	__antithesis_instrumentation__.Notify(596349)
	return d.Form == apd.Infinite
}

func isNaN(d *apd.Decimal) bool {
	__antithesis_instrumentation__.Notify(596350)
	return d.Form == apd.NaN
}

type corrAggregate struct {
	regressionAccumulatorDecimalBase
}

func newCorrAggregate(_ []*types.T, ctx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596351)
	return &corrAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *corrAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596352)
	return a.corrLastStage()
}

type finalCorrAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalCorrAggregate(_ []*types.T, ctx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596353)
	return &finalCorrAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalCorrAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596354)
	return a.corrLastStage()
}

type covarPopAggregate struct {
	regressionAccumulatorDecimalBase
}

func newCovarPopAggregate(_ []*types.T, ctx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596355)
	return &covarPopAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *covarPopAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596356)
	return a.covarPopLastStage()
}

type finalCovarPopAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalCovarPopAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596357)
	return &finalCovarPopAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegressionAccumulatorDecimalBase) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596358)
	a.regressionAccumulatorDecimalBase = regressionAccumulatorDecimalBase{
		singleDatumAggregateBase: a.singleDatumAggregateBase,
		ed:                       a.ed,
	}
	a.reset(ctx)
}

func (a *finalRegressionAccumulatorDecimalBase) Size() int64 {
	__antithesis_instrumentation__.Notify(596359)
	return sizeOfFinalRegressionAccumulatorDecimalBase
}

func (a *finalCovarPopAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596360)
	return a.covarPopLastStage()
}

type finalRegrSXXAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegrSXXAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596361)
	return &finalRegrSXXAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegrSXXAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596362)
	return a.regrSXXLastStage()
}

type finalRegrSXYAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegrSXYAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596363)
	return &finalRegrSXYAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegrSXYAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596364)
	return a.regrSXYLastStage()
}

type finalRegrSYYAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegrSYYAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596365)
	return &finalRegrSYYAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegrSYYAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596366)
	return a.regrSYYLastStage()
}

type covarSampAggregate struct {
	regressionAccumulatorDecimalBase
}

func newCovarSampAggregate(_ []*types.T, ctx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596367)
	return &covarSampAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *covarSampAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596368)
	return a.covarSampLastStage()
}

type finalCovarSampAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalCovarSampAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596369)
	return &finalCovarSampAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalCovarSampAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596370)
	return a.covarSampLastStage()
}

type regressionAvgXAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionAvgXAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596371)
	return &regressionAvgXAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionAvgXAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596372)
	return a.regressionAvgXLastStage()
}

type finalRegressionAvgXAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegressionAvgXAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596373)
	return &finalRegressionAvgXAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegressionAvgXAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596374)
	return a.regressionAvgXLastStage()
}

type regressionAvgYAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionAvgYAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596375)
	return &regressionAvgYAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionAvgYAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596376)
	return a.regressionAvgYLastStage()
}

type finalRegressionAvgYAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegressionAvgYAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596377)
	return &finalRegressionAvgYAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegressionAvgYAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596378)
	return a.regressionAvgYLastStage()
}

type regressionInterceptAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionInterceptAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596379)
	return &regressionInterceptAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionInterceptAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596380)
	return a.regressionInterceptLastStage()
}

type finalRegressionInterceptAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegressionInterceptAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596381)
	return &finalRegressionInterceptAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegressionInterceptAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596382)
	return a.regressionInterceptLastStage()
}

type regressionR2Aggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionR2Aggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596383)
	return &regressionR2Aggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionR2Aggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596384)
	return a.regressionR2LastStage()
}

type finalRegressionR2Aggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegressionR2Aggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596385)
	return &finalRegressionR2Aggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegressionR2Aggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596386)
	return a.regressionR2LastStage()
}

type regressionSlopeAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionSlopeAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596387)
	return &regressionSlopeAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionSlopeAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596388)
	return a.regressionSlopeLastStage()
}

type finalRegressionSlopeAggregate struct {
	finalRegressionAccumulatorDecimalBase
}

func newFinalRegressionSlopeAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596389)
	return &finalRegressionSlopeAggregate{
		finalRegressionAccumulatorDecimalBase{
			regressionAccumulatorDecimalBase: makeRegressionAccumulatorDecimalBase(ctx),
		},
	}
}

func (a *finalRegressionSlopeAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596390)
	return a.regressionSlopeLastStage()
}

type regressionSXXAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionSXXAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596391)
	return &regressionSXXAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionSXXAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596392)
	return a.regrSXXLastStage()
}

type regressionSXYAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionSXYAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596393)
	return &regressionSXYAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionSXYAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596394)
	return a.regrSXYLastStage()
}

type regressionSYYAggregate struct {
	regressionAccumulatorDecimalBase
}

func newRegressionSYYAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596395)
	return &regressionSYYAggregate{
		makeRegressionAccumulatorDecimalBase(ctx),
	}
}

func (a *regressionSYYAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596396)
	return a.regrSYYLastStage()
}

type regressionCountAggregate struct {
	count int
}

func newRegressionCountAggregate([]*types.T, *tree.EvalContext, tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596397)
	return &regressionCountAggregate{}
}

func (a *regressionCountAggregate) Add(
	_ context.Context, datumY tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596398)
	if datumY == tree.DNull {
		__antithesis_instrumentation__.Notify(596401)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596402)
	}
	__antithesis_instrumentation__.Notify(596399)

	datumX := otherArgs[0]
	if datumX == tree.DNull {
		__antithesis_instrumentation__.Notify(596403)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596404)
	}
	__antithesis_instrumentation__.Notify(596400)

	a.count++
	return nil
}

func (a *regressionCountAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596405)
	return tree.NewDInt(tree.DInt(a.count)), nil
}

func (a *regressionCountAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596406)
	a.count = 0
}

func (a *regressionCountAggregate) Close(context.Context) {
	__antithesis_instrumentation__.Notify(596407)
}

func (a *regressionCountAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596408)
	return sizeOfRegressionCountAggregate
}

type countAggregate struct {
	count int
}

func newCountAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596409)
	return &countAggregate{}
}

func (a *countAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596410)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596412)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596413)
	}
	__antithesis_instrumentation__.Notify(596411)
	a.count++
	return nil
}

func (a *countAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596414)
	return tree.NewDInt(tree.DInt(a.count)), nil
}

func (a *countAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596415)
	a.count = 0
}

func (a *countAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596416) }

func (a *countAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596417)
	return sizeOfCountAggregate
}

type countRowsAggregate struct {
	count int
}

func newCountRowsAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596418)
	return &countRowsAggregate{}
}

func (a *countRowsAggregate) Add(_ context.Context, _ tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596419)
	a.count++
	return nil
}

func (a *countRowsAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596420)
	return tree.NewDInt(tree.DInt(a.count)), nil
}

func (a *countRowsAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596421)
	a.count = 0
}

func (a *countRowsAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596422) }

func (a *countRowsAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596423)
	return sizeOfCountRowsAggregate
}

type maxAggregate struct {
	singleDatumAggregateBase

	max               tree.Datum
	evalCtx           *tree.EvalContext
	variableDatumSize bool
}

func newMaxAggregate(
	params []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596424)
	_, variable := tree.DatumTypeSize(params[0])

	return &maxAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		evalCtx:                  evalCtx,
		variableDatumSize:        variable,
	}
}

func (a *maxAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596425)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596429)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596430)
	}
	__antithesis_instrumentation__.Notify(596426)
	if a.max == nil {
		__antithesis_instrumentation__.Notify(596431)
		if err := a.updateMemoryUsage(ctx, int64(datum.Size())); err != nil {
			__antithesis_instrumentation__.Notify(596433)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596434)
		}
		__antithesis_instrumentation__.Notify(596432)
		a.max = datum
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596435)
	}
	__antithesis_instrumentation__.Notify(596427)
	c := a.max.Compare(a.evalCtx, datum)
	if c < 0 {
		__antithesis_instrumentation__.Notify(596436)
		a.max = datum
		if a.variableDatumSize {
			__antithesis_instrumentation__.Notify(596437)
			if err := a.updateMemoryUsage(ctx, int64(datum.Size())); err != nil {
				__antithesis_instrumentation__.Notify(596438)
				return err
			} else {
				__antithesis_instrumentation__.Notify(596439)
			}
		} else {
			__antithesis_instrumentation__.Notify(596440)
		}
	} else {
		__antithesis_instrumentation__.Notify(596441)
	}
	__antithesis_instrumentation__.Notify(596428)
	return nil
}

func (a *maxAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596442)
	if a.max == nil {
		__antithesis_instrumentation__.Notify(596444)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596445)
	}
	__antithesis_instrumentation__.Notify(596443)
	return a.max, nil
}

func (a *maxAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596446)
	a.max = nil
	a.reset(ctx)
}

func (a *maxAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596447)
	a.close(ctx)
}

func (a *maxAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596448)
	return sizeOfMaxAggregate
}

type minAggregate struct {
	singleDatumAggregateBase

	min               tree.Datum
	evalCtx           *tree.EvalContext
	variableDatumSize bool
}

func newMinAggregate(
	params []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596449)
	_, variable := tree.DatumTypeSize(params[0])

	return &minAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		evalCtx:                  evalCtx,
		variableDatumSize:        variable,
	}
}

func (a *minAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596450)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596454)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596455)
	}
	__antithesis_instrumentation__.Notify(596451)
	if a.min == nil {
		__antithesis_instrumentation__.Notify(596456)
		if err := a.updateMemoryUsage(ctx, int64(datum.Size())); err != nil {
			__antithesis_instrumentation__.Notify(596458)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596459)
		}
		__antithesis_instrumentation__.Notify(596457)
		a.min = datum
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596460)
	}
	__antithesis_instrumentation__.Notify(596452)
	c := a.min.Compare(a.evalCtx, datum)
	if c > 0 {
		__antithesis_instrumentation__.Notify(596461)
		a.min = datum
		if a.variableDatumSize {
			__antithesis_instrumentation__.Notify(596462)
			if err := a.updateMemoryUsage(ctx, int64(datum.Size())); err != nil {
				__antithesis_instrumentation__.Notify(596463)
				return err
			} else {
				__antithesis_instrumentation__.Notify(596464)
			}
		} else {
			__antithesis_instrumentation__.Notify(596465)
		}
	} else {
		__antithesis_instrumentation__.Notify(596466)
	}
	__antithesis_instrumentation__.Notify(596453)
	return nil
}

func (a *minAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596467)
	if a.min == nil {
		__antithesis_instrumentation__.Notify(596469)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596470)
	}
	__antithesis_instrumentation__.Notify(596468)
	return a.min, nil
}

func (a *minAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596471)
	a.min = nil
	a.reset(ctx)
}

func (a *minAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596472)
	a.close(ctx)
}

func (a *minAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596473)
	return sizeOfMinAggregate
}

type smallIntSumAggregate struct {
	sum         int64
	seenNonNull bool
}

func newSmallIntSumAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596474)
	return &smallIntSumAggregate{}
}

func (a *smallIntSumAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596475)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596478)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596479)
	}
	__antithesis_instrumentation__.Notify(596476)

	var ok bool
	a.sum, ok = arith.AddWithOverflow(a.sum, int64(tree.MustBeDInt(datum)))
	if !ok {
		__antithesis_instrumentation__.Notify(596480)
		return tree.ErrIntOutOfRange
	} else {
		__antithesis_instrumentation__.Notify(596481)
	}
	__antithesis_instrumentation__.Notify(596477)
	a.seenNonNull = true
	return nil
}

func (a *smallIntSumAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596482)
	if !a.seenNonNull {
		__antithesis_instrumentation__.Notify(596484)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596485)
	}
	__antithesis_instrumentation__.Notify(596483)
	return tree.NewDInt(tree.DInt(a.sum)), nil
}

func (a *smallIntSumAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596486)
	a.sum = 0
	a.seenNonNull = false
}

func (a *smallIntSumAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596487) }

func (a *smallIntSumAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596488)
	return sizeOfSmallIntSumAggregate
}

type intSumAggregate struct {
	singleDatumAggregateBase

	intSum      int64
	decSum      apd.Decimal
	tmpDec      apd.Decimal
	large       bool
	seenNonNull bool
}

func newIntSumAggregate(_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596489)
	return &intSumAggregate{singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx)}
}

func (a *intSumAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596490)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596493)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596494)
	}
	__antithesis_instrumentation__.Notify(596491)

	t := int64(tree.MustBeDInt(datum))
	if t != 0 {
		__antithesis_instrumentation__.Notify(596495)

		if !a.large {
			__antithesis_instrumentation__.Notify(596497)
			r, ok := arith.AddWithOverflow(a.intSum, t)
			if ok {
				__antithesis_instrumentation__.Notify(596498)
				a.intSum = r
			} else {
				__antithesis_instrumentation__.Notify(596499)

				a.large = true
				a.decSum.SetInt64(a.intSum)
			}
		} else {
			__antithesis_instrumentation__.Notify(596500)
		}
		__antithesis_instrumentation__.Notify(596496)

		if a.large {
			__antithesis_instrumentation__.Notify(596501)
			a.tmpDec.SetInt64(t)
			_, err := tree.ExactCtx.Add(&a.decSum, &a.decSum, &a.tmpDec)
			if err != nil {
				__antithesis_instrumentation__.Notify(596503)
				return err
			} else {
				__antithesis_instrumentation__.Notify(596504)
			}
			__antithesis_instrumentation__.Notify(596502)
			if err := a.updateMemoryUsage(ctx, int64(a.decSum.Size())); err != nil {
				__antithesis_instrumentation__.Notify(596505)
				return err
			} else {
				__antithesis_instrumentation__.Notify(596506)
			}
		} else {
			__antithesis_instrumentation__.Notify(596507)
		}
	} else {
		__antithesis_instrumentation__.Notify(596508)
	}
	__antithesis_instrumentation__.Notify(596492)
	a.seenNonNull = true
	return nil
}

func (a *intSumAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596509)
	if !a.seenNonNull {
		__antithesis_instrumentation__.Notify(596512)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596513)
	}
	__antithesis_instrumentation__.Notify(596510)
	dd := &tree.DDecimal{}
	if a.large {
		__antithesis_instrumentation__.Notify(596514)
		dd.Set(&a.decSum)
	} else {
		__antithesis_instrumentation__.Notify(596515)
		dd.SetInt64(a.intSum)
	}
	__antithesis_instrumentation__.Notify(596511)
	return dd, nil
}

func (a *intSumAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596516)

	a.seenNonNull = false
	a.intSum = 0
	a.large = false
}

func (a *intSumAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596517)
	a.close(ctx)
}

func (a *intSumAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596518)
	return sizeOfIntSumAggregate
}

type decimalSumAggregate struct {
	singleDatumAggregateBase

	sum        apd.Decimal
	sawNonNull bool
}

func newDecimalSumAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596519)
	return &decimalSumAggregate{singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx)}
}

func (a *decimalSumAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596520)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596524)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596525)
	}
	__antithesis_instrumentation__.Notify(596521)
	t := datum.(*tree.DDecimal)
	_, err := tree.ExactCtx.Add(&a.sum, &a.sum, &t.Decimal)
	if err != nil {
		__antithesis_instrumentation__.Notify(596526)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596527)
	}
	__antithesis_instrumentation__.Notify(596522)

	if err := a.updateMemoryUsage(ctx, int64(a.sum.Size())); err != nil {
		__antithesis_instrumentation__.Notify(596528)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596529)
	}
	__antithesis_instrumentation__.Notify(596523)

	a.sawNonNull = true
	return nil
}

func (a *decimalSumAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596530)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596532)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596533)
	}
	__antithesis_instrumentation__.Notify(596531)
	dd := &tree.DDecimal{}
	dd.Set(&a.sum)
	return dd, nil
}

func (a *decimalSumAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596534)
	a.sum.SetInt64(0)
	a.sawNonNull = false
	a.reset(ctx)
}

func (a *decimalSumAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596535)
	a.close(ctx)
}

func (a *decimalSumAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596536)
	return sizeOfDecimalSumAggregate
}

type floatSumAggregate struct {
	sum        float64
	sawNonNull bool
}

func newFloatSumAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596537)
	return &floatSumAggregate{}
}

func (a *floatSumAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596538)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596540)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596541)
	}
	__antithesis_instrumentation__.Notify(596539)
	t := datum.(*tree.DFloat)
	a.sum += float64(*t)
	a.sawNonNull = true
	return nil
}

func (a *floatSumAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596542)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596544)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596545)
	}
	__antithesis_instrumentation__.Notify(596543)
	return tree.NewDFloat(tree.DFloat(a.sum)), nil
}

func (a *floatSumAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596546)
	a.sawNonNull = false
	a.sum = 0
}

func (a *floatSumAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596547) }

func (a *floatSumAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596548)
	return sizeOfFloatSumAggregate
}

type intervalSumAggregate struct {
	sum        duration.Duration
	sawNonNull bool
}

func newIntervalSumAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596549)
	return &intervalSumAggregate{}
}

func (a *intervalSumAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596550)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596552)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596553)
	}
	__antithesis_instrumentation__.Notify(596551)
	t := datum.(*tree.DInterval).Duration
	a.sum = a.sum.Add(t)
	a.sawNonNull = true
	return nil
}

func (a *intervalSumAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596554)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596556)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596557)
	}
	__antithesis_instrumentation__.Notify(596555)
	return &tree.DInterval{Duration: a.sum}, nil
}

func (a *intervalSumAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596558)
	a.sum = a.sum.Sub(a.sum)
	a.sawNonNull = false
}

func (a *intervalSumAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596559) }

func (a *intervalSumAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596560)
	return sizeOfIntervalSumAggregate
}

var (
	decimalZero = apd.New(0, 0)
	decimalOne  = apd.New(1, 0)
	decimalTwo  = apd.New(2, 0)

	decimalNaN = &apd.Decimal{Form: apd.NaN}
)

type intSqrDiffAggregate struct {
	agg decimalSqrDiff

	tmpDec tree.DDecimal
}

func newIntSqrDiff(evalCtx *tree.EvalContext) decimalSqrDiff {
	__antithesis_instrumentation__.Notify(596561)
	return &intSqrDiffAggregate{agg: newDecimalSqrDiff(evalCtx)}
}

func newIntSqrDiffAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596562)
	return newIntSqrDiff(evalCtx)
}

func (a *intSqrDiffAggregate) Count() *apd.Decimal {
	__antithesis_instrumentation__.Notify(596563)
	return a.agg.Count()
}

func (a *intSqrDiffAggregate) Tmp() *apd.Decimal {
	__antithesis_instrumentation__.Notify(596564)
	return a.agg.Tmp()
}

func (a *intSqrDiffAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596565)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596567)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596568)
	}
	__antithesis_instrumentation__.Notify(596566)

	a.tmpDec.SetInt64(int64(tree.MustBeDInt(datum)))
	return a.agg.Add(ctx, &a.tmpDec)
}

func (a *intSqrDiffAggregate) intermediateResult() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596569)
	return a.agg.intermediateResult()
}

func (a *intSqrDiffAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596570)
	return a.agg.Result()
}

func (a *intSqrDiffAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596571)
	a.agg.Reset(ctx)
}

func (a *intSqrDiffAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596572)
	a.agg.Close(ctx)
}

func (a *intSqrDiffAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596573)
	return sizeOfIntSqrDiffAggregate
}

type floatSqrDiffAggregate struct {
	count   int64
	mean    float64
	sqrDiff float64
}

func newFloatSqrDiff() floatSqrDiff {
	__antithesis_instrumentation__.Notify(596574)
	return &floatSqrDiffAggregate{}
}

func newFloatSqrDiffAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596575)
	return newFloatSqrDiff()
}

func (a *floatSqrDiffAggregate) Count() int64 {
	__antithesis_instrumentation__.Notify(596576)
	return a.count
}

func (a *floatSqrDiffAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596577)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596579)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596580)
	}
	__antithesis_instrumentation__.Notify(596578)
	f := float64(*datum.(*tree.DFloat))

	a.count++
	delta := f - a.mean

	a.mean += delta / float64(a.count)
	a.sqrDiff += delta * (f - a.mean)
	return nil
}

func (a *floatSqrDiffAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596581)
	if a.count < 1 {
		__antithesis_instrumentation__.Notify(596583)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596584)
	}
	__antithesis_instrumentation__.Notify(596582)
	return tree.NewDFloat(tree.DFloat(a.sqrDiff)), nil
}

func (a *floatSqrDiffAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596585)
	a.count = 0
	a.mean = 0
	a.sqrDiff = 0
}

func (a *floatSqrDiffAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596586) }

func (a *floatSqrDiffAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596587)
	return sizeOfFloatSqrDiffAggregate
}

type decimalSqrDiffAggregate struct {
	singleDatumAggregateBase

	ed      *apd.ErrDecimal
	count   apd.Decimal
	mean    apd.Decimal
	sqrDiff apd.Decimal

	delta apd.Decimal
	tmp   apd.Decimal
}

func newDecimalSqrDiff(evalCtx *tree.EvalContext) decimalSqrDiff {
	__antithesis_instrumentation__.Notify(596588)
	ed := apd.MakeErrDecimal(tree.IntermediateCtx)
	return &decimalSqrDiffAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		ed:                       &ed,
	}
}

func newDecimalSqrDiffAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596589)
	return newDecimalSqrDiff(evalCtx)
}

func (a *decimalSqrDiffAggregate) Count() *apd.Decimal {
	__antithesis_instrumentation__.Notify(596590)
	return &a.count
}

func (a *decimalSqrDiffAggregate) Tmp() *apd.Decimal {
	__antithesis_instrumentation__.Notify(596591)
	return &a.tmp
}

func (a *decimalSqrDiffAggregate) Add(
	ctx context.Context, datum tree.Datum, _ ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596592)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596595)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596596)
	}
	__antithesis_instrumentation__.Notify(596593)
	d := &datum.(*tree.DDecimal).Decimal

	a.ed.Add(&a.count, &a.count, decimalOne)
	a.ed.Sub(&a.delta, d, &a.mean)
	a.ed.Quo(&a.tmp, &a.delta, &a.count)
	a.ed.Add(&a.mean, &a.mean, &a.tmp)
	a.ed.Sub(&a.tmp, d, &a.mean)
	a.ed.Add(&a.sqrDiff, &a.sqrDiff, a.ed.Mul(&a.delta, &a.delta, &a.tmp))

	size := int64(a.count.Size() +
		a.mean.Size() +
		a.sqrDiff.Size() +
		a.delta.Size() +
		a.tmp.Size())
	if err := a.updateMemoryUsage(ctx, size); err != nil {
		__antithesis_instrumentation__.Notify(596597)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596598)
	}
	__antithesis_instrumentation__.Notify(596594)

	return a.ed.Err()
}

func (a *decimalSqrDiffAggregate) intermediateResult() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596599)
	if a.count.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596601)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596602)
	}
	__antithesis_instrumentation__.Notify(596600)
	dd := &tree.DDecimal{}
	dd.Set(&a.sqrDiff)

	dd.Decimal.Reduce(&dd.Decimal)
	return dd, nil
}

func (a *decimalSqrDiffAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596603)
	res, err := a.intermediateResult()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(596606)
		return res == tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596607)
		return res, err
	} else {
		__antithesis_instrumentation__.Notify(596608)
	}
	__antithesis_instrumentation__.Notify(596604)

	dd := res.(*tree.DDecimal)

	_, err = tree.DecimalCtx.Round(&dd.Decimal, &a.sqrDiff)
	if err != nil {
		__antithesis_instrumentation__.Notify(596609)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596610)
	}
	__antithesis_instrumentation__.Notify(596605)

	dd.Decimal.Reduce(&dd.Decimal)
	return dd, nil
}

func (a *decimalSqrDiffAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596611)
	a.count.SetInt64(0)
	a.mean.SetInt64(0)
	a.sqrDiff.SetInt64(0)
	a.reset(ctx)
}

func (a *decimalSqrDiffAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596612)
	a.close(ctx)
}

func (a *decimalSqrDiffAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596613)
	return sizeOfDecimalSqrDiffAggregate
}

func newFloatFinalSqrdiffAggregate(
	_ []*types.T, _ *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596614)
	return newFloatSumSqrDiffs()
}

func newDecimalFinalSqrdiffAggregate(
	_ []*types.T, ctx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596615)
	return newDecimalSumSqrDiffs(ctx)
}

type floatSumSqrDiffsAggregate struct {
	count   int64
	mean    float64
	sqrDiff float64
}

func newFloatSumSqrDiffs() floatSqrDiff {
	__antithesis_instrumentation__.Notify(596616)
	return &floatSumSqrDiffsAggregate{}
}

func (a *floatSumSqrDiffsAggregate) Count() int64 {
	__antithesis_instrumentation__.Notify(596617)
	return a.count
}

func (a *floatSumSqrDiffsAggregate) Add(
	_ context.Context, sqrDiffD tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596618)
	sumD := otherArgs[0]
	countD := otherArgs[1]
	if sqrDiffD == tree.DNull || func() bool {
		__antithesis_instrumentation__.Notify(596621)
		return sumD == tree.DNull == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(596622)
		return countD == tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596623)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596624)
	}
	__antithesis_instrumentation__.Notify(596619)

	sqrDiff := float64(*sqrDiffD.(*tree.DFloat))
	sum := float64(*sumD.(*tree.DFloat))
	count := int64(*countD.(*tree.DInt))

	mean := sum / float64(count)
	delta := mean - a.mean

	totalCount, ok := arith.AddWithOverflow(a.count, count)
	if !ok {
		__antithesis_instrumentation__.Notify(596625)
		return pgerror.Newf(pgcode.NumericValueOutOfRange,
			"number of values in aggregate exceed max count of %d", math.MaxInt64,
		)
	} else {
		__antithesis_instrumentation__.Notify(596626)
	}
	__antithesis_instrumentation__.Notify(596620)

	a.sqrDiff += sqrDiff + delta*delta*float64(count)*float64(a.count)/float64(totalCount)
	a.count = totalCount
	a.mean += delta * float64(count) / float64(a.count)
	return nil
}

func (a *floatSumSqrDiffsAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596627)
	if a.count < 1 {
		__antithesis_instrumentation__.Notify(596629)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596630)
	}
	__antithesis_instrumentation__.Notify(596628)
	return tree.NewDFloat(tree.DFloat(a.sqrDiff)), nil
}

func (a *floatSumSqrDiffsAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596631)
	a.count = 0
	a.mean = 0
	a.sqrDiff = 0
}

func (a *floatSumSqrDiffsAggregate) Close(context.Context) {
	__antithesis_instrumentation__.Notify(596632)
}

func (a *floatSumSqrDiffsAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596633)
	return sizeOfFloatSumSqrDiffsAggregate
}

type decimalSumSqrDiffsAggregate struct {
	singleDatumAggregateBase

	ed      *apd.ErrDecimal
	count   apd.Decimal
	mean    apd.Decimal
	sqrDiff apd.Decimal

	tmpCount apd.Decimal
	tmpMean  apd.Decimal
	delta    apd.Decimal
	tmp      apd.Decimal
}

func newDecimalSumSqrDiffs(evalCtx *tree.EvalContext) decimalSqrDiff {
	__antithesis_instrumentation__.Notify(596634)
	ed := apd.MakeErrDecimal(tree.IntermediateCtx)
	return &decimalSumSqrDiffsAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		ed:                       &ed,
	}
}

func (a *decimalSumSqrDiffsAggregate) Count() *apd.Decimal {
	__antithesis_instrumentation__.Notify(596635)
	return &a.count
}

func (a *decimalSumSqrDiffsAggregate) Tmp() *apd.Decimal {
	__antithesis_instrumentation__.Notify(596636)
	return &a.tmp
}

func (a *decimalSumSqrDiffsAggregate) Add(
	ctx context.Context, sqrDiffD tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596637)
	sumD := otherArgs[0]
	countD := otherArgs[1]
	if sqrDiffD == tree.DNull || func() bool {
		__antithesis_instrumentation__.Notify(596640)
		return sumD == tree.DNull == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(596641)
		return countD == tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596642)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596643)
	}
	__antithesis_instrumentation__.Notify(596638)
	sqrDiff := &sqrDiffD.(*tree.DDecimal).Decimal
	sum := &sumD.(*tree.DDecimal).Decimal
	a.tmpCount.SetInt64(int64(*countD.(*tree.DInt)))

	a.ed.Quo(&a.tmpMean, sum, &a.tmpCount)
	a.ed.Sub(&a.delta, &a.tmpMean, &a.mean)

	a.ed.Add(&a.tmp, &a.tmpCount, &a.count)
	a.ed.Quo(&a.tmp, &a.count, &a.tmp)
	a.ed.Mul(&a.tmp, &a.tmpCount, &a.tmp)
	a.ed.Mul(&a.tmp, &a.delta, &a.tmp)
	a.ed.Mul(&a.tmp, &a.delta, &a.tmp)
	a.ed.Add(&a.tmp, sqrDiff, &a.tmp)

	a.ed.Add(&a.sqrDiff, &a.sqrDiff, &a.tmp)

	a.ed.Add(&a.count, &a.count, &a.tmpCount)

	a.ed.Mul(&a.tmp, &a.delta, &a.tmpCount)
	a.ed.Quo(&a.tmp, &a.tmp, &a.count)

	a.ed.Add(&a.mean, &a.mean, &a.tmp)

	size := int64(a.count.Size() +
		a.mean.Size() +
		a.sqrDiff.Size() +
		a.tmpCount.Size() +
		a.tmpMean.Size() +
		a.delta.Size() +
		a.tmp.Size())
	if err := a.updateMemoryUsage(ctx, size); err != nil {
		__antithesis_instrumentation__.Notify(596644)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596645)
	}
	__antithesis_instrumentation__.Notify(596639)

	return a.ed.Err()
}

func (a *decimalSumSqrDiffsAggregate) intermediateResult() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596646)
	if a.count.Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596648)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596649)
	}
	__antithesis_instrumentation__.Notify(596647)
	dd := &tree.DDecimal{Decimal: a.sqrDiff}
	return dd, nil
}

func (a *decimalSumSqrDiffsAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596650)
	res, err := a.intermediateResult()
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(596653)
		return res == tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596654)
		return res, err
	} else {
		__antithesis_instrumentation__.Notify(596655)
	}
	__antithesis_instrumentation__.Notify(596651)

	dd := res.(*tree.DDecimal)
	_, err = tree.DecimalCtx.Round(&dd.Decimal, &dd.Decimal)
	if err != nil {
		__antithesis_instrumentation__.Notify(596656)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596657)
	}
	__antithesis_instrumentation__.Notify(596652)

	dd.Reduce(&dd.Decimal)
	return dd, nil
}

func (a *decimalSumSqrDiffsAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596658)
	a.count.SetInt64(0)
	a.mean.SetInt64(0)
	a.sqrDiff.SetInt64(0)
	a.reset(ctx)
}

func (a *decimalSumSqrDiffsAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596659)
	a.close(ctx)
}

func (a *decimalSumSqrDiffsAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596660)
	return sizeOfDecimalSumSqrDiffsAggregate
}

type floatSqrDiff interface {
	tree.AggregateFunc
	Count() int64
}

type decimalSqrDiff interface {
	tree.AggregateFunc
	Count() *apd.Decimal
	Tmp() *apd.Decimal

	intermediateResult() (tree.Datum, error)
}

type floatVarianceAggregate struct {
	agg floatSqrDiff
}

type decimalVarianceAggregate struct {
	agg decimalSqrDiff
}

func newIntVarianceAggregate(
	params []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596661)
	return &decimalVarianceAggregate{agg: newIntSqrDiff(evalCtx)}
}

func newFloatVarianceAggregate(
	_ []*types.T, _ *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596662)
	return &floatVarianceAggregate{agg: newFloatSqrDiff()}
}

func newDecimalVarianceAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596663)
	return &decimalVarianceAggregate{agg: newDecimalSqrDiff(evalCtx)}
}

func newFloatFinalVarianceAggregate(
	_ []*types.T, _ *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596664)
	return &floatVarianceAggregate{agg: newFloatSumSqrDiffs()}
}

func newDecimalFinalVarianceAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596665)
	return &decimalVarianceAggregate{agg: newDecimalSumSqrDiffs(evalCtx)}
}

func (a *floatVarianceAggregate) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596666)
	return a.agg.Add(ctx, firstArg, otherArgs...)
}

func (a *decimalVarianceAggregate) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596667)
	return a.agg.Add(ctx, firstArg, otherArgs...)
}

func (a *floatVarianceAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596668)
	if a.agg.Count() < 2 {
		__antithesis_instrumentation__.Notify(596671)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596672)
	}
	__antithesis_instrumentation__.Notify(596669)
	sqrDiff, err := a.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(596673)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596674)
	}
	__antithesis_instrumentation__.Notify(596670)
	return tree.NewDFloat(tree.DFloat(float64(*sqrDiff.(*tree.DFloat)) / (float64(a.agg.Count()) - 1))), nil
}

func (a *decimalVarianceAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596675)
	if a.agg.Count().Cmp(decimalTwo) < 0 {
		__antithesis_instrumentation__.Notify(596680)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596681)
	}
	__antithesis_instrumentation__.Notify(596676)
	sqrDiff, err := a.agg.intermediateResult()
	if err != nil {
		__antithesis_instrumentation__.Notify(596682)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596683)
	}
	__antithesis_instrumentation__.Notify(596677)
	if _, err = tree.IntermediateCtx.Sub(a.agg.Tmp(), a.agg.Count(), decimalOne); err != nil {
		__antithesis_instrumentation__.Notify(596684)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596685)
	}
	__antithesis_instrumentation__.Notify(596678)
	dd := &tree.DDecimal{}
	if _, err = tree.DecimalCtx.Quo(&dd.Decimal, &sqrDiff.(*tree.DDecimal).Decimal, a.agg.Tmp()); err != nil {
		__antithesis_instrumentation__.Notify(596686)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596687)
	}
	__antithesis_instrumentation__.Notify(596679)

	dd.Decimal.Reduce(&dd.Decimal)
	return dd, nil
}

func (a *floatVarianceAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596688)
	a.agg.Reset(ctx)
}

func (a *floatVarianceAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596689)
	a.agg.Close(ctx)
}

func (a *floatVarianceAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596690)
	return sizeOfFloatVarianceAggregate
}

func (a *decimalVarianceAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596691)
	a.agg.Reset(ctx)
}

func (a *decimalVarianceAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596692)
	a.agg.Close(ctx)
}

func (a *decimalVarianceAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596693)
	return sizeOfDecimalVarianceAggregate
}

type floatVarPopAggregate struct {
	agg floatSqrDiff
}

type decimalVarPopAggregate struct {
	agg decimalSqrDiff
}

func newIntVarPopAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596694)
	return &decimalVarPopAggregate{agg: newIntSqrDiff(evalCtx)}
}

func newFloatVarPopAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596695)
	return &floatVarPopAggregate{agg: newFloatSqrDiff()}
}

func newDecimalVarPopAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596696)
	return &decimalVarPopAggregate{agg: newDecimalSqrDiff(evalCtx)}
}

func newFloatFinalVarPopAggregate(
	_ []*types.T, _ *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596697)
	return &floatVarPopAggregate{agg: newFloatSumSqrDiffs()}
}

func newDecimalFinalVarPopAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596698)
	return &decimalVarPopAggregate{agg: newDecimalSumSqrDiffs(evalCtx)}
}

func (a *floatVarPopAggregate) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596699)
	return a.agg.Add(ctx, firstArg, otherArgs...)
}

func (a *decimalVarPopAggregate) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596700)
	return a.agg.Add(ctx, firstArg, otherArgs...)
}

func (a *floatVarPopAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596701)
	if a.agg.Count() < 1 {
		__antithesis_instrumentation__.Notify(596704)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596705)
	}
	__antithesis_instrumentation__.Notify(596702)
	sqrDiff, err := a.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(596706)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596707)
	}
	__antithesis_instrumentation__.Notify(596703)
	return tree.NewDFloat(tree.DFloat(float64(*sqrDiff.(*tree.DFloat)) / (float64(a.agg.Count())))), nil
}

func (a *decimalVarPopAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596708)
	if a.agg.Count().Cmp(decimalOne) < 0 {
		__antithesis_instrumentation__.Notify(596712)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596713)
	}
	__antithesis_instrumentation__.Notify(596709)
	sqrDiff, err := a.agg.intermediateResult()
	if err != nil {
		__antithesis_instrumentation__.Notify(596714)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596715)
	}
	__antithesis_instrumentation__.Notify(596710)
	dd := &tree.DDecimal{}
	if _, err = tree.DecimalCtx.Quo(&dd.Decimal, &sqrDiff.(*tree.DDecimal).Decimal, a.agg.Count()); err != nil {
		__antithesis_instrumentation__.Notify(596716)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596717)
	}
	__antithesis_instrumentation__.Notify(596711)

	dd.Decimal.Reduce(&dd.Decimal)
	return dd, nil
}

func (a *floatVarPopAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596718)
	a.agg.Reset(ctx)
}

func (a *floatVarPopAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596719)
	a.agg.Close(ctx)
}

func (a *floatVarPopAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596720)
	return sizeOfFloatVarPopAggregate
}

func (a *decimalVarPopAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596721)
	a.agg.Reset(ctx)
}

func (a *decimalVarPopAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596722)
	a.agg.Close(ctx)
}

func (a *decimalVarPopAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596723)
	return sizeOfDecimalVarPopAggregate
}

type floatStdDevAggregate struct {
	agg tree.AggregateFunc
}

type decimalStdDevAggregate struct {
	agg tree.AggregateFunc
}

func newIntStdDevAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596724)
	return &decimalStdDevAggregate{agg: newIntVarianceAggregate(params, evalCtx, arguments)}
}

func newFloatStdDevAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596725)
	return &floatStdDevAggregate{agg: newFloatVarianceAggregate(params, evalCtx, arguments)}
}

func newDecimalStdDevAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596726)
	return &decimalStdDevAggregate{agg: newDecimalVarianceAggregate(params, evalCtx, arguments)}
}

func newFloatFinalStdDevAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596727)
	return &floatStdDevAggregate{agg: newFloatFinalVarianceAggregate(params, evalCtx, arguments)}
}

func newDecimalFinalStdDevAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596728)
	return &decimalStdDevAggregate{agg: newDecimalFinalVarianceAggregate(params, evalCtx, arguments)}
}

func newIntStdDevPopAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596729)
	return &decimalStdDevAggregate{agg: newIntVarPopAggregate(params, evalCtx, arguments)}
}

func newFloatStdDevPopAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596730)
	return &floatStdDevAggregate{agg: newFloatVarPopAggregate(params, evalCtx, arguments)}
}

func newDecimalStdDevPopAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596731)
	return &decimalStdDevAggregate{agg: newDecimalVarPopAggregate(params, evalCtx, arguments)}
}

func newDecimalFinalStdDevPopAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596732)
	return &decimalStdDevAggregate{agg: newDecimalFinalVarPopAggregate(params, evalCtx, arguments)}
}

func newFloatFinalStdDevPopAggregate(
	params []*types.T, evalCtx *tree.EvalContext, arguments tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596733)
	return &floatStdDevAggregate{agg: newFloatFinalVarPopAggregate(params, evalCtx, arguments)}
}

func (a *floatStdDevAggregate) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596734)
	return a.agg.Add(ctx, firstArg, otherArgs...)
}

func (a *decimalStdDevAggregate) Add(
	ctx context.Context, firstArg tree.Datum, otherArgs ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596735)
	return a.agg.Add(ctx, firstArg, otherArgs...)
}

func (a *floatStdDevAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596736)
	variance, err := a.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(596739)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596740)
	}
	__antithesis_instrumentation__.Notify(596737)
	if variance == tree.DNull {
		__antithesis_instrumentation__.Notify(596741)
		return variance, nil
	} else {
		__antithesis_instrumentation__.Notify(596742)
	}
	__antithesis_instrumentation__.Notify(596738)
	return tree.NewDFloat(tree.DFloat(math.Sqrt(float64(*variance.(*tree.DFloat))))), nil
}

func (a *decimalStdDevAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596743)

	variance, err := a.agg.Result()
	if err != nil {
		__antithesis_instrumentation__.Notify(596746)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(596747)
	}
	__antithesis_instrumentation__.Notify(596744)
	if variance == tree.DNull {
		__antithesis_instrumentation__.Notify(596748)
		return variance, nil
	} else {
		__antithesis_instrumentation__.Notify(596749)
	}
	__antithesis_instrumentation__.Notify(596745)
	varianceDec := variance.(*tree.DDecimal)
	_, err = tree.DecimalCtx.Sqrt(&varianceDec.Decimal, &varianceDec.Decimal)
	return varianceDec, err
}

func (a *floatStdDevAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596750)
	a.agg.Reset(ctx)
}

func (a *floatStdDevAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596751)
	a.agg.Close(ctx)
}

func (a *floatStdDevAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596752)
	return sizeOfFloatStdDevAggregate
}

func (a *decimalStdDevAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596753)
	a.agg.Reset(ctx)
}

func (a *decimalStdDevAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596754)
	a.agg.Close(ctx)
}

func (a *decimalStdDevAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596755)
	return sizeOfDecimalStdDevAggregate
}

type bytesXorAggregate struct {
	singleDatumAggregateBase

	sum        []byte
	sawNonNull bool
}

func newBytesXorAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596756)
	return &bytesXorAggregate{singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx)}
}

func (a *bytesXorAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596757)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596760)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596761)
	}
	__antithesis_instrumentation__.Notify(596758)
	t := []byte(*datum.(*tree.DBytes))
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596762)
		if err := a.updateMemoryUsage(ctx, int64(len(t))); err != nil {
			__antithesis_instrumentation__.Notify(596764)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596765)
		}
		__antithesis_instrumentation__.Notify(596763)
		a.sum = append([]byte(nil), t...)
	} else {
		__antithesis_instrumentation__.Notify(596766)
		if len(a.sum) != len(t) {
			__antithesis_instrumentation__.Notify(596767)
			return pgerror.Newf(pgcode.InvalidParameterValue,
				"arguments to xor must all be the same length %d vs %d", len(a.sum), len(t),
			)
		} else {
			__antithesis_instrumentation__.Notify(596768)
			for i := range t {
				__antithesis_instrumentation__.Notify(596769)
				a.sum[i] = a.sum[i] ^ t[i]
			}
		}
	}
	__antithesis_instrumentation__.Notify(596759)
	a.sawNonNull = true
	return nil
}

func (a *bytesXorAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596770)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596772)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596773)
	}
	__antithesis_instrumentation__.Notify(596771)
	return tree.NewDBytes(tree.DBytes(a.sum)), nil
}

func (a *bytesXorAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596774)
	a.sum = nil
	a.sawNonNull = false
	a.reset(ctx)
}

func (a *bytesXorAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596775)
	a.close(ctx)
}

func (a *bytesXorAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596776)
	return sizeOfBytesXorAggregate
}

type intXorAggregate struct {
	sum        int64
	sawNonNull bool
}

func newIntXorAggregate(_ []*types.T, _ *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596777)
	return &intXorAggregate{}
}

func (a *intXorAggregate) Add(_ context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596778)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596780)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(596781)
	}
	__antithesis_instrumentation__.Notify(596779)
	x := int64(*datum.(*tree.DInt))
	a.sum = a.sum ^ x
	a.sawNonNull = true
	return nil
}

func (a *intXorAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596782)
	if !a.sawNonNull {
		__antithesis_instrumentation__.Notify(596784)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596785)
	}
	__antithesis_instrumentation__.Notify(596783)
	return tree.NewDInt(tree.DInt(a.sum)), nil
}

func (a *intXorAggregate) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(596786)
	a.sum = 0
	a.sawNonNull = false
}

func (a *intXorAggregate) Close(context.Context) { __antithesis_instrumentation__.Notify(596787) }

func (a *intXorAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596788)
	return sizeOfIntXorAggregate
}

type jsonAggregate struct {
	singleDatumAggregateBase

	evalCtx    *tree.EvalContext
	builder    *json.ArrayBuilderWithCounter
	sawNonNull bool
}

func newJSONAggregate(_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596789)
	return &jsonAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		evalCtx:                  evalCtx,
		builder:                  json.NewArrayBuilderWithCounter(),
		sawNonNull:               false,
	}
}

func (a *jsonAggregate) Add(ctx context.Context, datum tree.Datum, _ ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(596790)
	j, err := tree.AsJSON(
		datum,
		a.evalCtx.SessionData().DataConversionConfig,
		a.evalCtx.GetLocation(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(596793)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596794)
	}
	__antithesis_instrumentation__.Notify(596791)
	a.builder.Add(j)
	if err = a.updateMemoryUsage(ctx, int64(a.builder.Size())); err != nil {
		__antithesis_instrumentation__.Notify(596795)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596796)
	}
	__antithesis_instrumentation__.Notify(596792)
	a.sawNonNull = true
	return nil
}

func (a *jsonAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596797)
	if a.sawNonNull {
		__antithesis_instrumentation__.Notify(596799)
		return tree.NewDJSON(a.builder.Build()), nil
	} else {
		__antithesis_instrumentation__.Notify(596800)
	}
	__antithesis_instrumentation__.Notify(596798)
	return tree.DNull, nil
}

func (a *jsonAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596801)
	a.builder = json.NewArrayBuilderWithCounter()
	a.sawNonNull = false
	a.reset(ctx)
}

func (a *jsonAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596802)
	a.close(ctx)
}

func (a *jsonAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596803)
	return sizeOfJSONAggregate
}

func validateInputFractions(datum tree.Datum) ([]float64, bool, error) {
	__antithesis_instrumentation__.Notify(596804)
	fractions := make([]float64, 0)
	singleInput := false

	validate := func(fraction float64) error {
		__antithesis_instrumentation__.Notify(596807)
		if fraction < 0 || func() bool {
			__antithesis_instrumentation__.Notify(596809)
			return fraction > 1.0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(596810)
			return pgerror.Newf(pgcode.NumericValueOutOfRange,
				"percentile value %f is not between 0 and 1", fraction)
		} else {
			__antithesis_instrumentation__.Notify(596811)
		}
		__antithesis_instrumentation__.Notify(596808)
		return nil
	}
	__antithesis_instrumentation__.Notify(596805)

	if datum.ResolvedType().Identical(types.Float) {
		__antithesis_instrumentation__.Notify(596812)
		fraction := float64(tree.MustBeDFloat(datum))
		singleInput = true
		if err := validate(fraction); err != nil {
			__antithesis_instrumentation__.Notify(596814)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(596815)
		}
		__antithesis_instrumentation__.Notify(596813)
		fractions = append(fractions, fraction)
	} else {
		__antithesis_instrumentation__.Notify(596816)
		if datum.ResolvedType().Equivalent(types.FloatArray) {
			__antithesis_instrumentation__.Notify(596817)
			fractionsDatum := tree.MustBeDArray(datum)
			for _, f := range fractionsDatum.Array {
				__antithesis_instrumentation__.Notify(596818)
				fraction := float64(tree.MustBeDFloat(f))
				if err := validate(fraction); err != nil {
					__antithesis_instrumentation__.Notify(596820)
					return nil, false, err
				} else {
					__antithesis_instrumentation__.Notify(596821)
				}
				__antithesis_instrumentation__.Notify(596819)
				fractions = append(fractions, fraction)
			}
		} else {
			__antithesis_instrumentation__.Notify(596822)
			panic(errors.AssertionFailedf("unexpected input type, %s", datum.ResolvedType()))
		}
	}
	__antithesis_instrumentation__.Notify(596806)
	return fractions, singleInput, nil
}

type percentileDiscAggregate struct {
	arr *tree.DArray

	acc mon.BoundAccount

	singleInput bool
	fractions   []float64
}

func newPercentileDiscAggregate(
	params []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596823)
	return &percentileDiscAggregate{
		arr: tree.NewDArray(params[1]),
		acc: evalCtx.Mon.MakeBoundAccount(),
	}
}

func (a *percentileDiscAggregate) Add(
	ctx context.Context, datum tree.Datum, others ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596824)
	if len(a.fractions) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(596827)
		return datum != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596828)
		fractions, singleInput, err := validateInputFractions(datum)
		if err != nil {
			__antithesis_instrumentation__.Notify(596830)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596831)
		}
		__antithesis_instrumentation__.Notify(596829)
		a.fractions = fractions
		a.singleInput = singleInput
	} else {
		__antithesis_instrumentation__.Notify(596832)
	}
	__antithesis_instrumentation__.Notify(596825)

	if len(others) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(596833)
		return others[0] != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596834)
		if err := a.acc.Grow(ctx, int64(others[0].Size())); err != nil {
			__antithesis_instrumentation__.Notify(596836)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596837)
		}
		__antithesis_instrumentation__.Notify(596835)
		return a.arr.Append(others[0])
	} else {
		__antithesis_instrumentation__.Notify(596838)
		if len(others) != 1 {
			__antithesis_instrumentation__.Notify(596839)
			panic(errors.AssertionFailedf("unexpected number of other datums passed in, expected 1, got %d", len(others)))
		} else {
			__antithesis_instrumentation__.Notify(596840)
		}
	}
	__antithesis_instrumentation__.Notify(596826)
	return nil
}

func (a *percentileDiscAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596841)

	if a.arr.Len() == 0 {
		__antithesis_instrumentation__.Notify(596844)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596845)
	}
	__antithesis_instrumentation__.Notify(596842)

	if len(a.fractions) > 0 {
		__antithesis_instrumentation__.Notify(596846)
		res := tree.NewDArray(a.arr.ParamTyp)
		for _, fraction := range a.fractions {
			__antithesis_instrumentation__.Notify(596849)

			if fraction == 0.0 {
				__antithesis_instrumentation__.Notify(596851)
				if err := res.Append(a.arr.Array[0]); err != nil {
					__antithesis_instrumentation__.Notify(596853)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(596854)
				}
				__antithesis_instrumentation__.Notify(596852)
				continue
			} else {
				__antithesis_instrumentation__.Notify(596855)
			}
			__antithesis_instrumentation__.Notify(596850)

			rowIndex := int(math.Ceil(fraction*float64(a.arr.Len()))) - 1
			if err := res.Append(a.arr.Array[rowIndex]); err != nil {
				__antithesis_instrumentation__.Notify(596856)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(596857)
			}
		}
		__antithesis_instrumentation__.Notify(596847)

		if a.singleInput {
			__antithesis_instrumentation__.Notify(596858)
			return res.Array[0], nil
		} else {
			__antithesis_instrumentation__.Notify(596859)
		}
		__antithesis_instrumentation__.Notify(596848)
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(596860)
	}
	__antithesis_instrumentation__.Notify(596843)
	panic("input must either be a single fraction, or an array of fractions")
}

func (a *percentileDiscAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596861)
	a.arr = tree.NewDArray(a.arr.ParamTyp)
	a.acc.Empty(ctx)
	a.singleInput = false
	a.fractions = a.fractions[:0]
}

func (a *percentileDiscAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596862)
	a.acc.Close(ctx)
}

func (a *percentileDiscAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596863)
	return sizeOfPercentileDiscAggregate
}

type percentileContAggregate struct {
	arr *tree.DArray

	acc mon.BoundAccount

	singleInput bool
	fractions   []float64
}

func newPercentileContAggregate(
	params []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596864)
	return &percentileContAggregate{
		arr: tree.NewDArray(params[1]),
		acc: evalCtx.Mon.MakeBoundAccount(),
	}
}

func (a *percentileContAggregate) Add(
	ctx context.Context, datum tree.Datum, others ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596865)
	if len(a.fractions) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(596868)
		return datum != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596869)
		fractions, singleInput, err := validateInputFractions(datum)
		if err != nil {
			__antithesis_instrumentation__.Notify(596871)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596872)
		}
		__antithesis_instrumentation__.Notify(596870)
		a.fractions = fractions
		a.singleInput = singleInput
	} else {
		__antithesis_instrumentation__.Notify(596873)
	}
	__antithesis_instrumentation__.Notify(596866)

	if len(others) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(596874)
		return others[0] != tree.DNull == true
	}() == true {
		__antithesis_instrumentation__.Notify(596875)
		if err := a.acc.Grow(ctx, int64(others[0].Size())); err != nil {
			__antithesis_instrumentation__.Notify(596877)
			return err
		} else {
			__antithesis_instrumentation__.Notify(596878)
		}
		__antithesis_instrumentation__.Notify(596876)
		return a.arr.Append(others[0])
	} else {
		__antithesis_instrumentation__.Notify(596879)
		if len(others) != 1 {
			__antithesis_instrumentation__.Notify(596880)
			panic(errors.AssertionFailedf("unexpected number of other datums passed in, expected 1, got %d", len(others)))
		} else {
			__antithesis_instrumentation__.Notify(596881)
		}
	}
	__antithesis_instrumentation__.Notify(596867)
	return nil
}

func (a *percentileContAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596882)

	if a.arr.Len() == 0 {
		__antithesis_instrumentation__.Notify(596885)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(596886)
	}
	__antithesis_instrumentation__.Notify(596883)

	if len(a.fractions) > 0 {
		__antithesis_instrumentation__.Notify(596887)
		res := tree.NewDArray(a.arr.ParamTyp)
		for _, fraction := range a.fractions {
			__antithesis_instrumentation__.Notify(596890)
			rowNumber := 1.0 + (fraction * (float64(a.arr.Len()) - 1.0))
			ceilRowNumber := int(math.Ceil(rowNumber))
			floorRowNumber := int(math.Floor(rowNumber))

			if a.arr.ParamTyp.Identical(types.Float) {
				__antithesis_instrumentation__.Notify(596891)
				var target float64
				if rowNumber == float64(ceilRowNumber) && func() bool {
					__antithesis_instrumentation__.Notify(596893)
					return rowNumber == float64(floorRowNumber) == true
				}() == true {
					__antithesis_instrumentation__.Notify(596894)
					target = float64(tree.MustBeDFloat(a.arr.Array[int(rowNumber)-1]))
				} else {
					__antithesis_instrumentation__.Notify(596895)
					floorValue := float64(tree.MustBeDFloat(a.arr.Array[floorRowNumber-1]))
					ceilValue := float64(tree.MustBeDFloat(a.arr.Array[ceilRowNumber-1]))
					target = (float64(ceilRowNumber)-rowNumber)*floorValue +
						(rowNumber-float64(floorRowNumber))*ceilValue
				}
				__antithesis_instrumentation__.Notify(596892)
				if err := res.Append(tree.NewDFloat(tree.DFloat(target))); err != nil {
					__antithesis_instrumentation__.Notify(596896)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(596897)
				}
			} else {
				__antithesis_instrumentation__.Notify(596898)
				if a.arr.ParamTyp.Family() == types.IntervalFamily {
					__antithesis_instrumentation__.Notify(596899)
					var target *tree.DInterval
					if rowNumber == float64(ceilRowNumber) && func() bool {
						__antithesis_instrumentation__.Notify(596901)
						return rowNumber == float64(floorRowNumber) == true
					}() == true {
						__antithesis_instrumentation__.Notify(596902)
						target = tree.MustBeDInterval(a.arr.Array[int(rowNumber)-1])
					} else {
						__antithesis_instrumentation__.Notify(596903)
						floorValue := tree.MustBeDInterval(a.arr.Array[floorRowNumber-1]).AsFloat64()
						ceilValue := tree.MustBeDInterval(a.arr.Array[ceilRowNumber-1]).AsFloat64()
						targetDuration := duration.FromFloat64(
							(float64(ceilRowNumber)-rowNumber)*floorValue +
								(rowNumber-float64(floorRowNumber))*ceilValue)
						target = tree.NewDInterval(targetDuration, types.DefaultIntervalTypeMetadata)
					}
					__antithesis_instrumentation__.Notify(596900)
					if err := res.Append(target); err != nil {
						__antithesis_instrumentation__.Notify(596904)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(596905)
					}
				} else {
					__antithesis_instrumentation__.Notify(596906)
					panic(errors.AssertionFailedf("argument type must be float or interval, got %s", a.arr.ParamTyp.String()))
				}
			}
		}
		__antithesis_instrumentation__.Notify(596888)
		if a.singleInput {
			__antithesis_instrumentation__.Notify(596907)
			return res.Array[0], nil
		} else {
			__antithesis_instrumentation__.Notify(596908)
		}
		__antithesis_instrumentation__.Notify(596889)
		return res, nil
	} else {
		__antithesis_instrumentation__.Notify(596909)
	}
	__antithesis_instrumentation__.Notify(596884)
	panic("input must either be a single fraction, or an array of fractions")
}

func (a *percentileContAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596910)
	a.arr = tree.NewDArray(a.arr.ParamTyp)
	a.acc.Empty(ctx)
	a.singleInput = false
	a.fractions = a.fractions[:0]
}

func (a *percentileContAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596911)
	a.acc.Close(ctx)
}

func (a *percentileContAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596912)
	return sizeOfPercentileContAggregate
}

type jsonObjectAggregate struct {
	singleDatumAggregateBase

	evalCtx    *tree.EvalContext
	builder    *json.ObjectBuilderWithCounter
	sawNonNull bool
}

func newJSONObjectAggregate(
	_ []*types.T, evalCtx *tree.EvalContext, _ tree.Datums,
) tree.AggregateFunc {
	__antithesis_instrumentation__.Notify(596913)
	return &jsonObjectAggregate{
		singleDatumAggregateBase: makeSingleDatumAggregateBase(evalCtx),
		evalCtx:                  evalCtx,
		builder:                  json.NewObjectBuilderWithCounter(),
		sawNonNull:               false,
	}
}

func (a *jsonObjectAggregate) Add(
	ctx context.Context, datum tree.Datum, others ...tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(596914)
	if len(others) != 1 {
		__antithesis_instrumentation__.Notify(596920)
		return errors.Errorf("wrong number of arguments, expected key/value pair")
	} else {
		__antithesis_instrumentation__.Notify(596921)
	}
	__antithesis_instrumentation__.Notify(596915)

	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(596922)
		return pgerror.New(pgcode.InvalidParameterValue,
			"field name must not be null")
	} else {
		__antithesis_instrumentation__.Notify(596923)
	}
	__antithesis_instrumentation__.Notify(596916)

	key, err := asJSONBuildObjectKey(
		datum,
		a.evalCtx.SessionData().DataConversionConfig,
		a.evalCtx.GetLocation(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(596924)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596925)
	}
	__antithesis_instrumentation__.Notify(596917)
	val, err := tree.AsJSON(
		others[0],
		a.evalCtx.SessionData().DataConversionConfig,
		a.evalCtx.GetLocation(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(596926)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596927)
	}
	__antithesis_instrumentation__.Notify(596918)
	a.builder.Add(key, val)
	if err = a.updateMemoryUsage(ctx, int64(a.builder.Size())); err != nil {
		__antithesis_instrumentation__.Notify(596928)
		return err
	} else {
		__antithesis_instrumentation__.Notify(596929)
	}
	__antithesis_instrumentation__.Notify(596919)
	a.sawNonNull = true
	return nil
}

func (a *jsonObjectAggregate) Result() (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(596930)
	if a.sawNonNull {
		__antithesis_instrumentation__.Notify(596932)
		return tree.NewDJSON(a.builder.Build()), nil
	} else {
		__antithesis_instrumentation__.Notify(596933)
	}
	__antithesis_instrumentation__.Notify(596931)
	return tree.DNull, nil
}

func (a *jsonObjectAggregate) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596934)
	a.builder = json.NewObjectBuilderWithCounter()
	a.sawNonNull = false
	a.reset(ctx)
}

func (a *jsonObjectAggregate) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(596935)
	a.close(ctx)
}

func (a *jsonObjectAggregate) Size() int64 {
	__antithesis_instrumentation__.Notify(596936)
	return sizeOfJSONObjectAggregate
}
