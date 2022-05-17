package colexectestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/colmem"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/pmezard/go-difflib/difflib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var OrderedDistinctColsToOperators func(input colexecop.Operator, distinctCols []uint32, typs []*types.T, nullsAreDistinct bool,
) (colexecop.ResettableOperator, []bool)

type Tuple []interface{}

func (t Tuple) String() string {
	__antithesis_instrumentation__.Notify(431021)
	var sb strings.Builder
	sb.WriteString("[")
	for i := range t {
		__antithesis_instrumentation__.Notify(431023)
		if i != 0 {
			__antithesis_instrumentation__.Notify(431025)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(431026)
		}
		__antithesis_instrumentation__.Notify(431024)
		if d, ok := t[i].(apd.Decimal); ok {
			__antithesis_instrumentation__.Notify(431027)
			sb.WriteString(d.String())
		} else {
			__antithesis_instrumentation__.Notify(431028)
			if d, ok := t[i].(*apd.Decimal); ok {
				__antithesis_instrumentation__.Notify(431029)
				sb.WriteString(d.String())
			} else {
				__antithesis_instrumentation__.Notify(431030)
				if d, ok := t[i].([]byte); ok {
					__antithesis_instrumentation__.Notify(431031)
					sb.WriteString(string(d))
				} else {
					__antithesis_instrumentation__.Notify(431032)
					if d, ok := t[i].(json.JSON); ok {
						__antithesis_instrumentation__.Notify(431033)
						sb.WriteString(d.String())
					} else {
						__antithesis_instrumentation__.Notify(431034)
						sb.WriteString(fmt.Sprintf("%v", t[i]))
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(431022)
	sb.WriteString("]")
	return sb.String()
}

func (t Tuple) less(other Tuple, evalCtx *tree.EvalContext, tupleFromOtherSet Tuple) bool {
	__antithesis_instrumentation__.Notify(431035)
	for i := range t {
		__antithesis_instrumentation__.Notify(431037)

		if t[i] == nil && func() bool {
			__antithesis_instrumentation__.Notify(431046)
			return other[i] == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(431047)
			continue
		} else {
			__antithesis_instrumentation__.Notify(431048)
			if t[i] == nil && func() bool {
				__antithesis_instrumentation__.Notify(431049)
				return other[i] != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(431050)
				return true
			} else {
				__antithesis_instrumentation__.Notify(431051)
				if t[i] != nil && func() bool {
					__antithesis_instrumentation__.Notify(431052)
					return other[i] == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(431053)
					return false
				} else {
					__antithesis_instrumentation__.Notify(431054)
				}
			}
		}
		__antithesis_instrumentation__.Notify(431038)

		if d1, ok := t[i].(tree.Datum); ok {
			__antithesis_instrumentation__.Notify(431055)
			d2 := other[i].(tree.Datum)
			cmp := d1.Compare(evalCtx, d2)
			if cmp == 0 {
				__antithesis_instrumentation__.Notify(431057)
				continue
			} else {
				__antithesis_instrumentation__.Notify(431058)
			}
			__antithesis_instrumentation__.Notify(431056)
			return cmp < 0
		} else {
			__antithesis_instrumentation__.Notify(431059)
		}
		__antithesis_instrumentation__.Notify(431039)

		lhsVal := reflect.ValueOf(t[i])
		rhsVal := reflect.ValueOf(other[i])

		if lhsVal.Type().Name() == "Decimal" && func() bool {
			__antithesis_instrumentation__.Notify(431060)
			return lhsVal.CanInterface() == true
		}() == true {
			__antithesis_instrumentation__.Notify(431061)
			lhsDecimal := lhsVal.Interface().(apd.Decimal)
			rhsDecimal := rhsVal.Interface().(apd.Decimal)
			cmp := (&lhsDecimal).CmpTotal(&rhsDecimal)
			if cmp == 0 {
				__antithesis_instrumentation__.Notify(431062)
				continue
			} else {
				__antithesis_instrumentation__.Notify(431063)
				if cmp == -1 {
					__antithesis_instrumentation__.Notify(431064)
					return true
				} else {
					__antithesis_instrumentation__.Notify(431065)
					return false
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(431066)
		}
		__antithesis_instrumentation__.Notify(431040)

		if lhsVal.Type().Implements(reflect.TypeOf((*json.JSON)(nil)).Elem()) && func() bool {
			__antithesis_instrumentation__.Notify(431067)
			return lhsVal.CanInterface() == true
		}() == true {
			__antithesis_instrumentation__.Notify(431068)
			lhsJSON := lhsVal.Interface().(json.JSON)
			rhsJSON := rhsVal.Interface().(json.JSON)
			cmp, err := lhsJSON.Compare(rhsJSON)
			if err != nil {
				__antithesis_instrumentation__.Notify(431070)
				colexecerror.InternalError(errors.NewAssertionErrorWithWrappedErrf(err, "failed json compare"))
			} else {
				__antithesis_instrumentation__.Notify(431071)
			}
			__antithesis_instrumentation__.Notify(431069)
			if cmp == 0 {
				__antithesis_instrumentation__.Notify(431072)
				continue
			} else {
				__antithesis_instrumentation__.Notify(431073)
				if cmp == -1 {
					__antithesis_instrumentation__.Notify(431074)
					return true
				} else {
					__antithesis_instrumentation__.Notify(431075)
					return false
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(431076)
		}
		__antithesis_instrumentation__.Notify(431041)

		if lhsVal.Type().String() == "[]uint8" {
			__antithesis_instrumentation__.Notify(431077)
			lhsStr := string(lhsVal.Interface().([]uint8))
			rhsStr := string(rhsVal.Interface().([]uint8))
			if lhsStr == rhsStr {
				__antithesis_instrumentation__.Notify(431078)
				continue
			} else {
				__antithesis_instrumentation__.Notify(431079)
				if lhsStr < rhsStr {
					__antithesis_instrumentation__.Notify(431080)
					return true
				} else {
					__antithesis_instrumentation__.Notify(431081)
					return false
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(431082)
		}
		__antithesis_instrumentation__.Notify(431042)

		if lhsVal.Type().Name() == "Duration" {
			__antithesis_instrumentation__.Notify(431083)
			lhsDuration := lhsVal.Interface().(duration.Duration)
			rhsDuration := rhsVal.Interface().(duration.Duration)
			cmp := lhsDuration.Compare(rhsDuration)
			if cmp == 0 {
				__antithesis_instrumentation__.Notify(431084)
				continue
			} else {
				__antithesis_instrumentation__.Notify(431085)
				if cmp == -1 {
					__antithesis_instrumentation__.Notify(431086)
					return true
				} else {
					__antithesis_instrumentation__.Notify(431087)
					return false
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(431088)
		}
		__antithesis_instrumentation__.Notify(431043)

		if lhsVal.Type().Name() == "Time" {
			__antithesis_instrumentation__.Notify(431089)
			lhsTime := lhsVal.Interface().(time.Time)
			rhsTime := rhsVal.Interface().(time.Time)
			if lhsTime.Equal(rhsTime) {
				__antithesis_instrumentation__.Notify(431090)
				continue
			} else {
				__antithesis_instrumentation__.Notify(431091)
				if lhsTime.Before(rhsTime) {
					__antithesis_instrumentation__.Notify(431092)
					return true
				} else {
					__antithesis_instrumentation__.Notify(431093)
					return false
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(431094)
		}
		__antithesis_instrumentation__.Notify(431044)

		if t[i] == other[i] {
			__antithesis_instrumentation__.Notify(431095)
			continue
		} else {
			__antithesis_instrumentation__.Notify(431096)
		}
		__antithesis_instrumentation__.Notify(431045)

		switch typ := lhsVal.Type().Name(); typ {
		case "int", "int16", "int32", "int64":
			__antithesis_instrumentation__.Notify(431097)
			return lhsVal.Int() < rhsVal.Int()
		case "uint", "uint16", "uint32", "uint64":
			__antithesis_instrumentation__.Notify(431098)
			return lhsVal.Uint() < rhsVal.Uint()
		case "float", "float64":
			__antithesis_instrumentation__.Notify(431099)
			return lhsVal.Float() < rhsVal.Float()
		case "bool":
			__antithesis_instrumentation__.Notify(431100)
			return !lhsVal.Bool() && func() bool {
				__antithesis_instrumentation__.Notify(431104)
				return rhsVal.Bool() == true
			}() == true
		case "string":
			__antithesis_instrumentation__.Notify(431101)
			lString, rString := lhsVal.String(), rhsVal.String()
			if tupleFromOtherSet != nil && func() bool {
				__antithesis_instrumentation__.Notify(431105)
				return len(tupleFromOtherSet) > i == true
			}() == true {
				__antithesis_instrumentation__.Notify(431106)
				if d, ok := tupleFromOtherSet[i].(tree.Datum); ok {
					__antithesis_instrumentation__.Notify(431107)

					d1 := stringToDatum(lString, d.ResolvedType(), evalCtx)
					d2 := stringToDatum(rString, d.ResolvedType(), evalCtx)
					cmp := d1.Compare(evalCtx, d2)
					if cmp == 0 {
						__antithesis_instrumentation__.Notify(431109)
						continue
					} else {
						__antithesis_instrumentation__.Notify(431110)
					}
					__antithesis_instrumentation__.Notify(431108)
					return cmp < 0
				} else {
					__antithesis_instrumentation__.Notify(431111)
				}
			} else {
				__antithesis_instrumentation__.Notify(431112)
			}
			__antithesis_instrumentation__.Notify(431102)
			return lString < rString
		default:
			__antithesis_instrumentation__.Notify(431103)
			colexecerror.InternalError(errors.AssertionFailedf("Unhandled comparison type: %s", typ))
		}
	}
	__antithesis_instrumentation__.Notify(431036)
	return false
}

func (t Tuple) clone() Tuple {
	__antithesis_instrumentation__.Notify(431113)
	b := make(Tuple, len(t))
	for i := range b {
		__antithesis_instrumentation__.Notify(431115)
		b[i] = t[i]
	}
	__antithesis_instrumentation__.Notify(431114)

	return b
}

type Tuples []Tuple

func (t Tuples) Clone() Tuples {
	__antithesis_instrumentation__.Notify(431116)
	b := make(Tuples, len(t))
	for i := range b {
		__antithesis_instrumentation__.Notify(431118)
		b[i] = t[i].clone()
	}
	__antithesis_instrumentation__.Notify(431117)
	return b
}

func (t Tuples) String() string {
	__antithesis_instrumentation__.Notify(431119)
	var sb strings.Builder
	sb.WriteString("[")
	for i := range t {
		__antithesis_instrumentation__.Notify(431121)
		if i != 0 {
			__antithesis_instrumentation__.Notify(431123)
			sb.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(431124)
		}
		__antithesis_instrumentation__.Notify(431122)
		sb.WriteString(t[i].String())
	}
	__antithesis_instrumentation__.Notify(431120)
	sb.WriteString("]")
	return sb.String()
}

func (t Tuples) sort(evalCtx *tree.EvalContext, tupleFromOtherSet Tuple) Tuples {
	__antithesis_instrumentation__.Notify(431125)
	b := make(Tuples, len(t))
	for i := range b {
		__antithesis_instrumentation__.Notify(431128)
		b[i] = make(Tuple, len(t[i]))
		copy(b[i], t[i])
	}
	__antithesis_instrumentation__.Notify(431126)
	sort.SliceStable(b, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(431129)
		lhs := b[i]
		rhs := b[j]
		return lhs.less(rhs, evalCtx, tupleFromOtherSet)
	})
	__antithesis_instrumentation__.Notify(431127)
	return b
}

type VerifierType int

const (
	OrderedVerifier VerifierType = iota

	UnorderedVerifier

	PartialOrderedVerifier
)

func maybeHasNulls(b coldata.Batch) bool {
	__antithesis_instrumentation__.Notify(431130)
	if b.Length() == 0 {
		__antithesis_instrumentation__.Notify(431133)
		return false
	} else {
		__antithesis_instrumentation__.Notify(431134)
	}
	__antithesis_instrumentation__.Notify(431131)
	for i := 0; i < b.Width(); i++ {
		__antithesis_instrumentation__.Notify(431135)
		if b.ColVec(i).MaybeHasNulls() {
			__antithesis_instrumentation__.Notify(431136)
			return true
		} else {
			__antithesis_instrumentation__.Notify(431137)
		}
	}
	__antithesis_instrumentation__.Notify(431132)
	return false
}

type TestRunner func(*testing.T, *colmem.Allocator, []Tuples, [][]*types.T, Tuples, VerifierType, func([]colexecop.Operator) (colexecop.Operator, error))

func RunTests(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	expected Tuples,
	verifier VerifierType,
	constructor func(inputs []colexecop.Operator) (colexecop.Operator, error),
) {
	__antithesis_instrumentation__.Notify(431138)
	RunTestsWithTyps(t, allocator, tups, nil, expected, verifier, constructor)
}

func RunTestsWithTyps(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	typs [][]*types.T,
	expected Tuples,
	verifier VerifierType,
	constructor func(inputs []colexecop.Operator) (colexecop.Operator, error),
) {
	__antithesis_instrumentation__.Notify(431139)
	RunTestsWithOrderedCols(t, allocator, tups, typs, expected, verifier, nil, constructor)
}

func RunTestsWithOrderedCols(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	typs [][]*types.T,
	expected Tuples,
	verifier VerifierType,
	orderedCols []uint32,
	constructor func(inputs []colexecop.Operator) (colexecop.Operator, error),
) {
	__antithesis_instrumentation__.Notify(431140)
	RunTestsWithoutAllNullsInjectionWithErrorHandler(t, allocator, tups, typs, expected, verifier, constructor, func(err error) { __antithesis_instrumentation__.Notify(431141); t.Fatal(err) }, orderedCols)

	{
		__antithesis_instrumentation__.Notify(431142)
		ctx := context.Background()
		evalCtx := tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings())
		defer evalCtx.Stop(ctx)
		log.Info(ctx, "allNullsInjection")

		onlyNullsInTheInput := true
	OUTER:
		for _, tup := range tups {
			__antithesis_instrumentation__.Notify(431147)
			for i := 0; i < len(tup); i++ {
				__antithesis_instrumentation__.Notify(431148)
				for j := 0; j < len(tup[i]); j++ {
					__antithesis_instrumentation__.Notify(431149)
					if tup[i][j] != nil {
						__antithesis_instrumentation__.Notify(431150)
						onlyNullsInTheInput = false
						break OUTER
					} else {
						__antithesis_instrumentation__.Notify(431151)
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(431143)
		opConstructor := func(injectAllNulls bool) colexecop.Operator {
			__antithesis_instrumentation__.Notify(431152)
			inputSources := make([]colexecop.Operator, len(tups))
			var inputTypes []*types.T
			for i, tup := range tups {
				__antithesis_instrumentation__.Notify(431155)
				if typs != nil {
					__antithesis_instrumentation__.Notify(431157)
					inputTypes = typs[i]
				} else {
					__antithesis_instrumentation__.Notify(431158)
				}
				__antithesis_instrumentation__.Notify(431156)
				input := NewOpTestInput(allocator, 1, tup, inputTypes).(*opTestInput)
				input.injectAllNulls = injectAllNulls
				inputSources[i] = input
			}
			__antithesis_instrumentation__.Notify(431153)
			op, err := constructor(inputSources)
			if err != nil {
				__antithesis_instrumentation__.Notify(431159)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(431160)
			}
			__antithesis_instrumentation__.Notify(431154)
			op.Init(ctx)
			return op
		}
		__antithesis_instrumentation__.Notify(431144)
		originalOp := opConstructor(false)
		opWithNulls := opConstructor(true)
		foundDifference := false
		for {
			__antithesis_instrumentation__.Notify(431161)
			originalBatch := originalOp.Next()
			batchWithNulls := opWithNulls.Next()
			if originalBatch.Length() != batchWithNulls.Length() {
				__antithesis_instrumentation__.Notify(431165)
				foundDifference = true
				break
			} else {
				__antithesis_instrumentation__.Notify(431166)
			}
			__antithesis_instrumentation__.Notify(431162)
			if originalBatch.Length() == 0 {
				__antithesis_instrumentation__.Notify(431167)
				break
			} else {
				__antithesis_instrumentation__.Notify(431168)
			}
			__antithesis_instrumentation__.Notify(431163)
			var originalTuples, tuplesWithNulls Tuples
			for i := 0; i < originalBatch.Length(); i++ {
				__antithesis_instrumentation__.Notify(431169)

				originalTuples = append(originalTuples, GetTupleFromBatch(originalBatch, i))
				tuplesWithNulls = append(tuplesWithNulls, GetTupleFromBatch(batchWithNulls, i))
			}
			__antithesis_instrumentation__.Notify(431164)
			if err := AssertTuplesSetsEqual(originalTuples, tuplesWithNulls, evalCtx); err != nil {
				__antithesis_instrumentation__.Notify(431170)

				foundDifference = true
				break
			} else {
				__antithesis_instrumentation__.Notify(431171)
			}
		}
		__antithesis_instrumentation__.Notify(431145)
		if onlyNullsInTheInput {
			__antithesis_instrumentation__.Notify(431172)
			require.False(t, foundDifference, "since there were only "+
				"nulls in the input tuples, we expect for all nulls injection to not "+
				"change the output")
		} else {
			__antithesis_instrumentation__.Notify(431173)
			require.True(t, foundDifference, "since there were "+
				"non-nulls in the input tuples, we expect for all nulls injection to "+
				"change the output")
		}
		__antithesis_instrumentation__.Notify(431146)
		closeIfCloser(t, originalOp)
		closeIfCloser(t, opWithNulls)
	}
}

func closeIfCloser(t *testing.T, op colexecop.Operator) {
	__antithesis_instrumentation__.Notify(431174)
	if c, ok := op.(colexecop.Closer); ok {
		__antithesis_instrumentation__.Notify(431175)
		if err := c.Close(context.Background()); err != nil {
			__antithesis_instrumentation__.Notify(431176)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(431177)
		}
	} else {
		__antithesis_instrumentation__.Notify(431178)
	}
}

func isOperatorChainResettable(op execinfra.OpNode) bool {
	__antithesis_instrumentation__.Notify(431179)
	if _, resettable := op.(colexecop.ResettableOperator); !resettable {
		__antithesis_instrumentation__.Notify(431182)
		return false
	} else {
		__antithesis_instrumentation__.Notify(431183)
	}
	__antithesis_instrumentation__.Notify(431180)
	for i := 0; i < op.ChildCount(true); i++ {
		__antithesis_instrumentation__.Notify(431184)
		if !isOperatorChainResettable(op.Child(i, true)) {
			__antithesis_instrumentation__.Notify(431185)
			return false
		} else {
			__antithesis_instrumentation__.Notify(431186)
		}
	}
	__antithesis_instrumentation__.Notify(431181)
	return true
}

func RunTestsWithoutAllNullsInjection(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	typs [][]*types.T,
	expected Tuples,
	verifier VerifierType,
	constructor func(inputs []colexecop.Operator) (colexecop.Operator, error),
) {
	__antithesis_instrumentation__.Notify(431187)
	RunTestsWithoutAllNullsInjectionWithErrorHandler(t, allocator, tups, typs, expected, verifier, constructor, func(err error) { __antithesis_instrumentation__.Notify(431188); t.Fatal(err) }, nil)
}

var SkipRandomNullsInjection = func(error) { __antithesis_instrumentation__.Notify(431189) }

func RunTestsWithoutAllNullsInjectionWithErrorHandler(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	typs [][]*types.T,
	expected Tuples,
	verifier VerifierType,
	constructor func(inputs []colexecop.Operator) (colexecop.Operator, error),
	errorHandler func(error),
	orderedCols []uint32,
) {
	__antithesis_instrumentation__.Notify(431190)
	ctx := context.Background()
	verifyFn := (*OpTestOutput).VerifyAnyOrder
	skipVerifySelAndNullsResets := true
	if verifier == OrderedVerifier {
		__antithesis_instrumentation__.Notify(431195)
		verifyFn = (*OpTestOutput).Verify

		skipVerifySelAndNullsResets = false
	} else {
		__antithesis_instrumentation__.Notify(431196)
		if verifier == PartialOrderedVerifier {
			__antithesis_instrumentation__.Notify(431197)
			verifyFn = (*OpTestOutput).VerifyPartialOrder
			skipVerifySelAndNullsResets = false
		} else {
			__antithesis_instrumentation__.Notify(431198)
		}
	}
	__antithesis_instrumentation__.Notify(431191)
	skipRandomNullsInjection := &errorHandler == &SkipRandomNullsInjection
	if skipRandomNullsInjection {
		__antithesis_instrumentation__.Notify(431199)
		errorHandler = func(err error) {
			__antithesis_instrumentation__.Notify(431200)
			if err != nil {
				__antithesis_instrumentation__.Notify(431201)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(431202)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(431203)
	}
	__antithesis_instrumentation__.Notify(431192)
	RunTestsWithFn(t, allocator, tups, typs, func(t *testing.T, inputs []colexecop.Operator) {
		__antithesis_instrumentation__.Notify(431204)
		op, err := constructor(inputs)
		if err != nil {
			__antithesis_instrumentation__.Notify(431209)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(431210)
		}
		__antithesis_instrumentation__.Notify(431205)
		out := NewOpTestOutput(op, expected)
		if len(typs) > 0 {
			__antithesis_instrumentation__.Notify(431211)
			out.typs = typs[0]
		} else {
			__antithesis_instrumentation__.Notify(431212)
		}
		__antithesis_instrumentation__.Notify(431206)
		out.orderedCols = orderedCols
		if err := verifyFn(out); err != nil {
			__antithesis_instrumentation__.Notify(431213)
			errorHandler(err)
		} else {
			__antithesis_instrumentation__.Notify(431214)
		}
		__antithesis_instrumentation__.Notify(431207)
		if isOperatorChainResettable(op) {
			__antithesis_instrumentation__.Notify(431215)
			log.Info(ctx, "reusing after reset")
			out.Reset(ctx)
			if err := verifyFn(out); err != nil {
				__antithesis_instrumentation__.Notify(431216)
				errorHandler(err)
			} else {
				__antithesis_instrumentation__.Notify(431217)
			}
		} else {
			__antithesis_instrumentation__.Notify(431218)
		}
		__antithesis_instrumentation__.Notify(431208)
		closeIfCloser(t, op)
	})
	__antithesis_instrumentation__.Notify(431193)

	if !skipVerifySelAndNullsResets {
		__antithesis_instrumentation__.Notify(431219)
		log.Info(ctx, "verifySelAndNullResets")

		var (
			secondBatchHasSelection, secondBatchHasNulls bool
			inputTypes                                   []*types.T
		)
		for round := 0; round < 2; round++ {
			__antithesis_instrumentation__.Notify(431220)
			inputSources := make([]colexecop.Operator, len(tups))
			for i, tup := range tups {
				__antithesis_instrumentation__.Notify(431224)
				if typs != nil {
					__antithesis_instrumentation__.Notify(431226)
					inputTypes = typs[i]
				} else {
					__antithesis_instrumentation__.Notify(431227)
				}
				__antithesis_instrumentation__.Notify(431225)
				inputSources[i] = NewOpTestInput(allocator, 1, tup, inputTypes)
			}
			__antithesis_instrumentation__.Notify(431221)
			op, err := constructor(inputSources)
			if err != nil {
				__antithesis_instrumentation__.Notify(431228)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(431229)
			}
			__antithesis_instrumentation__.Notify(431222)

			defer closeIfCloser(t, op)
			op.Init(ctx)

			lessThanTwoBatches := true
			if err = colexecerror.CatchVectorizedRuntimeError(func() {
				__antithesis_instrumentation__.Notify(431230)
				b := op.Next()
				if b.Length() == 0 {
					__antithesis_instrumentation__.Notify(431235)
					return
				} else {
					__antithesis_instrumentation__.Notify(431236)
				}
				__antithesis_instrumentation__.Notify(431231)
				if round == 1 {
					__antithesis_instrumentation__.Notify(431237)
					if secondBatchHasSelection {
						__antithesis_instrumentation__.Notify(431239)
						b.SetSelection(false)
					} else {
						__antithesis_instrumentation__.Notify(431240)
						b.SetSelection(true)
					}
					__antithesis_instrumentation__.Notify(431238)
					if secondBatchHasNulls {
						__antithesis_instrumentation__.Notify(431241)

						b.ResetInternalBatch()
					} else {
						__antithesis_instrumentation__.Notify(431242)
						for i := 0; i < b.Width(); i++ {
							__antithesis_instrumentation__.Notify(431243)
							b.ColVec(i).Nulls().SetNulls()
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(431244)
				}
				__antithesis_instrumentation__.Notify(431232)
				b = op.Next()
				if b.Length() == 0 {
					__antithesis_instrumentation__.Notify(431245)
					return
				} else {
					__antithesis_instrumentation__.Notify(431246)
				}
				__antithesis_instrumentation__.Notify(431233)
				lessThanTwoBatches = false
				if round == 0 {
					__antithesis_instrumentation__.Notify(431247)
					secondBatchHasSelection = b.Selection() != nil
					secondBatchHasNulls = maybeHasNulls(b)
				} else {
					__antithesis_instrumentation__.Notify(431248)
				}
				__antithesis_instrumentation__.Notify(431234)
				if round == 1 {
					__antithesis_instrumentation__.Notify(431249)
					if secondBatchHasSelection {
						__antithesis_instrumentation__.Notify(431251)
						assert.NotNil(t, b.Selection())
					} else {
						__antithesis_instrumentation__.Notify(431252)
						assert.Nil(t, b.Selection())
					}
					__antithesis_instrumentation__.Notify(431250)
					if secondBatchHasNulls {
						__antithesis_instrumentation__.Notify(431253)
						assert.True(t, maybeHasNulls(b))
					} else {
						__antithesis_instrumentation__.Notify(431254)
						assert.False(t, maybeHasNulls(b))
					}
				} else {
					__antithesis_instrumentation__.Notify(431255)
				}
			}); err != nil {
				__antithesis_instrumentation__.Notify(431256)
				errorHandler(err)
			} else {
				__antithesis_instrumentation__.Notify(431257)
			}
			__antithesis_instrumentation__.Notify(431223)
			if lessThanTwoBatches {
				__antithesis_instrumentation__.Notify(431258)
				return
			} else {
				__antithesis_instrumentation__.Notify(431259)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(431260)
	}
	__antithesis_instrumentation__.Notify(431194)

	if !skipRandomNullsInjection {
		__antithesis_instrumentation__.Notify(431261)
		log.Info(ctx, "randomNullsInjection")

		inputSources := make([]colexecop.Operator, len(tups))
		var inputTypes []*types.T
		for i, tup := range tups {
			__antithesis_instrumentation__.Notify(431265)
			if typs != nil {
				__antithesis_instrumentation__.Notify(431267)
				inputTypes = typs[i]
			} else {
				__antithesis_instrumentation__.Notify(431268)
			}
			__antithesis_instrumentation__.Notify(431266)
			input := NewOpTestInput(allocator, 1, tup, inputTypes).(*opTestInput)
			input.injectRandomNulls = true
			inputSources[i] = input
		}
		__antithesis_instrumentation__.Notify(431262)
		op, err := constructor(inputSources)
		if err != nil {
			__antithesis_instrumentation__.Notify(431269)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(431270)
		}
		__antithesis_instrumentation__.Notify(431263)
		if err = colexecerror.CatchVectorizedRuntimeError(func() {
			__antithesis_instrumentation__.Notify(431271)
			op.Init(ctx)
			for b := op.Next(); b.Length() > 0; b = op.Next() {
				__antithesis_instrumentation__.Notify(431272)
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(431273)
			errorHandler(err)
		} else {
			__antithesis_instrumentation__.Notify(431274)
		}
		__antithesis_instrumentation__.Notify(431264)
		closeIfCloser(t, op)
	} else {
		__antithesis_instrumentation__.Notify(431275)
	}
}

func RunTestsWithFn(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	typs [][]*types.T,
	test func(t *testing.T, inputs []colexecop.Operator),
) {
	__antithesis_instrumentation__.Notify(431276)

	batchSizes := make([]int, 0, 3)
	batchSizes = append(batchSizes, 1)
	smallButGreaterThanOne := int(math.Trunc(.002 * float64(coldata.BatchSize())))
	if smallButGreaterThanOne > 1 {
		__antithesis_instrumentation__.Notify(431278)
		batchSizes = append(batchSizes, smallButGreaterThanOne)
	} else {
		__antithesis_instrumentation__.Notify(431279)
	}
	__antithesis_instrumentation__.Notify(431277)
	batchSizes = append(batchSizes, coldata.BatchSize())

	for _, batchSize := range batchSizes {
		__antithesis_instrumentation__.Notify(431280)
		for _, useSel := range []bool{false, true} {
			__antithesis_instrumentation__.Notify(431281)
			log.Infof(context.Background(), "batchSize=%d/sel=%t", batchSize, useSel)
			inputSources := make([]colexecop.Operator, len(tups))
			var inputTypes []*types.T
			if useSel {
				__antithesis_instrumentation__.Notify(431283)
				for i, tup := range tups {
					__antithesis_instrumentation__.Notify(431284)
					if typs != nil {
						__antithesis_instrumentation__.Notify(431286)
						inputTypes = typs[i]
					} else {
						__antithesis_instrumentation__.Notify(431287)
					}
					__antithesis_instrumentation__.Notify(431285)
					inputSources[i] = newOpTestSelInput(allocator, batchSize, tup, inputTypes)
					inputSources[i].(*opTestInput).batchLengthRandomizationEnabled = true
				}
			} else {
				__antithesis_instrumentation__.Notify(431288)
				for i, tup := range tups {
					__antithesis_instrumentation__.Notify(431289)
					if typs != nil {
						__antithesis_instrumentation__.Notify(431291)
						inputTypes = typs[i]
					} else {
						__antithesis_instrumentation__.Notify(431292)
					}
					__antithesis_instrumentation__.Notify(431290)
					inputSources[i] = NewOpTestInput(allocator, batchSize, tup, inputTypes)
					inputSources[i].(*opTestInput).batchLengthRandomizationEnabled = true
				}
			}
			__antithesis_instrumentation__.Notify(431282)
			test(t, inputSources)
		}
	}
}

func RunTestsWithFixedSel(
	t *testing.T,
	allocator *colmem.Allocator,
	tups []Tuples,
	typs []*types.T,
	sel []int,
	test func(t *testing.T, inputs []colexecop.Operator),
) {
	__antithesis_instrumentation__.Notify(431293)
	for _, batchSize := range []int{1, 2, 3, 16, 1024} {
		__antithesis_instrumentation__.Notify(431294)
		log.Infof(context.Background(), "batchSize=%d/fixedSel", batchSize)
		inputSources := make([]colexecop.Operator, len(tups))
		for i, tup := range tups {
			__antithesis_instrumentation__.Notify(431296)
			inputSources[i] = NewOpFixedSelTestInput(allocator, sel, batchSize, tup, typs)
		}
		__antithesis_instrumentation__.Notify(431295)
		test(t, inputSources)
	}
}

func stringToDatum(val string, typ *types.T, evalCtx *tree.EvalContext) tree.Datum {
	__antithesis_instrumentation__.Notify(431297)
	expr, err := parser.ParseExpr(val)
	if err != nil {
		__antithesis_instrumentation__.Notify(431301)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(431302)
	}
	__antithesis_instrumentation__.Notify(431298)
	semaCtx := tree.MakeSemaContext()
	typedExpr, err := tree.TypeCheck(context.Background(), expr, &semaCtx, typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(431303)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(431304)
	}
	__antithesis_instrumentation__.Notify(431299)
	d, err := typedExpr.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(431305)
		colexecerror.InternalError(err)
	} else {
		__antithesis_instrumentation__.Notify(431306)
	}
	__antithesis_instrumentation__.Notify(431300)
	return d
}

func setColVal(vec coldata.Vec, idx int, val interface{}, evalCtx *tree.EvalContext) {
	__antithesis_instrumentation__.Notify(431307)
	switch vec.CanonicalTypeFamily() {
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(431308)
		var (
			bytesVal []byte
			ok       bool
		)
		bytesVal, ok = val.([]byte)
		if !ok {
			__antithesis_instrumentation__.Notify(431315)
			bytesVal = []byte(val.(string))
		} else {
			__antithesis_instrumentation__.Notify(431316)
		}
		__antithesis_instrumentation__.Notify(431309)
		vec.Bytes().Set(idx, bytesVal)
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(431310)

		if decimalVal, ok := val.(apd.Decimal); ok {
			__antithesis_instrumentation__.Notify(431317)
			vec.Decimal()[idx].Set(&decimalVal)
		} else {
			__antithesis_instrumentation__.Notify(431318)
			floatVal := val.(float64)
			decimalVal, _, err := apd.NewFromString(fmt.Sprintf("%f", floatVal))
			if err != nil {
				__antithesis_instrumentation__.Notify(431320)
				colexecerror.InternalError(
					errors.NewAssertionErrorWithWrappedErrf(err, "unable to set decimal %f", floatVal))
			} else {
				__antithesis_instrumentation__.Notify(431321)
			}
			__antithesis_instrumentation__.Notify(431319)

			vec.Decimal()[idx].Set(decimalVal)
		}
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(431311)
		var j json.JSON
		if j2, ok := val.(json.JSON); ok {
			__antithesis_instrumentation__.Notify(431322)
			j = j2
		} else {
			__antithesis_instrumentation__.Notify(431323)
			s := val.(string)
			var err error
			j, err = json.ParseJSON(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(431324)
				colexecerror.InternalError(
					errors.NewAssertionErrorWithWrappedErrf(err, "unable to set json %s", s))
			} else {
				__antithesis_instrumentation__.Notify(431325)
			}
		}
		__antithesis_instrumentation__.Notify(431312)
		vec.JSON().Set(idx, j)
	case typeconv.DatumVecCanonicalTypeFamily:
		__antithesis_instrumentation__.Notify(431313)
		switch v := val.(type) {
		case tree.Datum:
			__antithesis_instrumentation__.Notify(431326)
			vec.Datum().Set(idx, v)
		case string:
			__antithesis_instrumentation__.Notify(431327)
			vec.Datum().Set(idx, stringToDatum(v, vec.Type(), evalCtx))
		default:
			__antithesis_instrumentation__.Notify(431328)
			colexecerror.InternalError(errors.AssertionFailedf("unexpected type %T of datum-backed value: %v", v, v))
		}
	default:
		__antithesis_instrumentation__.Notify(431314)
		reflect.ValueOf(vec.Col()).Index(idx).Set(reflect.ValueOf(val).Convert(reflect.TypeOf(vec.Col()).Elem()))
	}
}

func extrapolateTypesFromTuples(tups Tuples) []*types.T {
	__antithesis_instrumentation__.Notify(431329)
	typs := make([]*types.T, len(tups[0]))
	for i := range typs {
		__antithesis_instrumentation__.Notify(431331)

		typs[i] = types.Int
		for _, tup := range tups {
			__antithesis_instrumentation__.Notify(431332)
			if tup[i] != nil {
				__antithesis_instrumentation__.Notify(431333)
				typs[i] = typeconv.UnsafeFromGoType(tup[i])
				break
			} else {
				__antithesis_instrumentation__.Notify(431334)
			}
		}
	}
	__antithesis_instrumentation__.Notify(431330)
	return typs
}

type opTestInput struct {
	colexecop.ZeroInputNode

	allocator *colmem.Allocator

	typs []*types.T

	batchSize                       int
	batchLengthRandomizationEnabled bool
	tuples                          Tuples

	initialTuples Tuples
	batch         coldata.Batch
	useSel        bool
	rng           *rand.Rand
	selection     []int
	evalCtx       *tree.EvalContext

	injectAllNulls bool

	injectRandomNulls bool
}

var _ colexecop.ResettableOperator = &opTestInput{}

func NewOpTestInput(
	allocator *colmem.Allocator, batchSize int, tuples Tuples, typs []*types.T,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431335)
	ret := &opTestInput{
		allocator:     allocator,
		batchSize:     batchSize,
		tuples:        tuples,
		initialTuples: tuples,
		typs:          typs,
		evalCtx:       tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings()),
	}
	return ret
}

func newOpTestSelInput(
	allocator *colmem.Allocator, batchSize int, tuples Tuples, typs []*types.T,
) *opTestInput {
	__antithesis_instrumentation__.Notify(431336)
	ret := &opTestInput{
		allocator:     allocator,
		useSel:        true,
		batchSize:     batchSize,
		tuples:        tuples,
		initialTuples: tuples,
		typs:          typs,
		evalCtx:       tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings()),
	}
	return ret
}

func (s *opTestInput) Init(context.Context) {
	__antithesis_instrumentation__.Notify(431337)
	if s.typs == nil {
		__antithesis_instrumentation__.Notify(431339)
		if len(s.tuples) == 0 {
			__antithesis_instrumentation__.Notify(431341)
			colexecerror.InternalError(errors.AssertionFailedf("empty tuple source with no specified types"))
		} else {
			__antithesis_instrumentation__.Notify(431342)
		}
		__antithesis_instrumentation__.Notify(431340)
		s.typs = extrapolateTypesFromTuples(s.tuples)
	} else {
		__antithesis_instrumentation__.Notify(431343)
	}
	__antithesis_instrumentation__.Notify(431338)
	s.batch = s.allocator.NewMemBatchWithMaxCapacity(s.typs)
	s.rng, _ = randutil.NewTestRand()
	s.selection = make([]int, coldata.BatchSize())
	for i := range s.selection {
		__antithesis_instrumentation__.Notify(431344)
		s.selection[i] = i
	}
}

func (s *opTestInput) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431345)
	if len(s.tuples) == 0 {
		__antithesis_instrumentation__.Notify(431353)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(431354)
	}
	__antithesis_instrumentation__.Notify(431346)
	s.batch.ResetInternalBatch()
	batchSize := s.batchSize
	if len(s.tuples) < batchSize {
		__antithesis_instrumentation__.Notify(431355)
		batchSize = len(s.tuples)
	} else {
		__antithesis_instrumentation__.Notify(431356)
	}
	__antithesis_instrumentation__.Notify(431347)
	if s.batchLengthRandomizationEnabled {
		__antithesis_instrumentation__.Notify(431357)

		if s.rng.Float64() < 0.5 {
			__antithesis_instrumentation__.Notify(431358)
			batchSize = s.rng.Intn(batchSize) + 1
		} else {
			__antithesis_instrumentation__.Notify(431359)
		}
	} else {
		__antithesis_instrumentation__.Notify(431360)
	}
	__antithesis_instrumentation__.Notify(431348)
	tups := s.tuples[:batchSize]
	s.tuples = s.tuples[batchSize:]

	tupleLen := len(tups[0])
	for i := range tups {
		__antithesis_instrumentation__.Notify(431361)
		if len(tups[i]) != tupleLen {
			__antithesis_instrumentation__.Notify(431362)
			colexecerror.InternalError(errors.AssertionFailedf("mismatched tuple lens: found %+v expected %d vals",
				tups[i], tupleLen))
		} else {
			__antithesis_instrumentation__.Notify(431363)
		}
	}
	__antithesis_instrumentation__.Notify(431349)

	if s.useSel {
		__antithesis_instrumentation__.Notify(431364)
		for i := range s.selection {
			__antithesis_instrumentation__.Notify(431369)
			s.selection[i] = i
		}
		__antithesis_instrumentation__.Notify(431365)

		s.rng.Shuffle(len(s.selection), func(i, j int) {
			__antithesis_instrumentation__.Notify(431370)
			s.selection[i], s.selection[j] = s.selection[j], s.selection[i]
		})
		__antithesis_instrumentation__.Notify(431366)
		sort.Slice(s.selection[:batchSize], func(i, j int) bool {
			__antithesis_instrumentation__.Notify(431371)
			return s.selection[i] < s.selection[j]
		})
		__antithesis_instrumentation__.Notify(431367)

		for i := range s.selection[batchSize:] {
			__antithesis_instrumentation__.Notify(431372)
			s.selection[batchSize+i] = coldata.BatchSize() + 1
		}
		__antithesis_instrumentation__.Notify(431368)

		s.batch.SetSelection(true)
		copy(s.batch.Selection(), s.selection)
	} else {
		__antithesis_instrumentation__.Notify(431373)
	}
	__antithesis_instrumentation__.Notify(431350)

	for _, colVec := range s.batch.ColVecs() {
		__antithesis_instrumentation__.Notify(431374)
		if colVec.CanonicalTypeFamily() != types.UnknownFamily {
			__antithesis_instrumentation__.Notify(431375)
			colVec.Nulls().UnsetNulls()
		} else {
			__antithesis_instrumentation__.Notify(431376)
		}
	}
	__antithesis_instrumentation__.Notify(431351)

	rng, _ := randutil.NewTestRand()

	for i := range s.typs {
		__antithesis_instrumentation__.Notify(431377)
		vec := s.batch.ColVec(i)

		col := reflect.ValueOf(vec.Col())
		for j := 0; j < batchSize; j++ {
			__antithesis_instrumentation__.Notify(431378)

			outputIdx := s.selection[j]
			injectRandomNull := s.injectRandomNulls && func() bool {
				__antithesis_instrumentation__.Notify(431379)
				return rng.Float64() < 0.5 == true
			}() == true
			if tups[j][i] == nil || func() bool {
				__antithesis_instrumentation__.Notify(431380)
				return s.injectAllNulls == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(431381)
				return injectRandomNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(431382)
				vec.Nulls().SetNull(outputIdx)
				if rng.Float64() < 0.5 {
					__antithesis_instrumentation__.Notify(431383)

					canonicalTypeFamily := vec.CanonicalTypeFamily()
					switch canonicalTypeFamily {
					case types.DecimalFamily:
						__antithesis_instrumentation__.Notify(431384)
						d := apd.Decimal{}
						_, err := d.SetFloat64(rng.Float64())
						if err != nil {
							__antithesis_instrumentation__.Notify(431393)
							colexecerror.InternalError(errors.NewAssertionErrorWithWrappedErrf(err, "error setting float"))
						} else {
							__antithesis_instrumentation__.Notify(431394)
						}
						__antithesis_instrumentation__.Notify(431385)
						col.Index(outputIdx).Set(reflect.ValueOf(d))
					case types.BytesFamily:
						__antithesis_instrumentation__.Notify(431386)
						newBytes := make([]byte, rng.Intn(16)+1)
						rng.Read(newBytes)
						setColVal(vec, outputIdx, newBytes, s.evalCtx)
					case types.IntervalFamily:
						__antithesis_instrumentation__.Notify(431387)
						setColVal(vec, outputIdx, duration.MakeDuration(rng.Int63(), rng.Int63(), rng.Int63()), s.evalCtx)
					case types.JsonFamily:
						__antithesis_instrumentation__.Notify(431388)
						j, err := json.Random(20, rng)
						if err != nil {
							__antithesis_instrumentation__.Notify(431395)
							colexecerror.InternalError(errors.NewAssertionErrorWithWrappedErrf(err, "error creating json"))
						} else {
							__antithesis_instrumentation__.Notify(431396)
						}
						__antithesis_instrumentation__.Notify(431389)
						setColVal(vec, outputIdx, j, s.evalCtx)
					case types.TimestampTZFamily:
						__antithesis_instrumentation__.Notify(431390)
						t := timeutil.Unix(rng.Int63n(2000000000), rng.Int63n(1000000))
						t.Round(tree.TimeFamilyPrecisionToRoundDuration(vec.Type().Precision()))
						setColVal(vec, outputIdx, t, s.evalCtx)
					case typeconv.DatumVecCanonicalTypeFamily:
						__antithesis_instrumentation__.Notify(431391)
						switch vec.Type().Family() {
						case types.CollatedStringFamily:
							__antithesis_instrumentation__.Notify(431397)
							collatedStringType := types.MakeCollatedString(types.String, *randgen.RandCollationLocale(rng))
							randomBytes := make([]byte, rng.Intn(16)+1)
							rng.Read(randomBytes)
							d, err := tree.NewDCollatedString(string(randomBytes), collatedStringType.Locale(), &tree.CollationEnvironment{})
							if err != nil {
								__antithesis_instrumentation__.Notify(431402)
								colexecerror.InternalError(err)
							} else {
								__antithesis_instrumentation__.Notify(431403)
							}
							__antithesis_instrumentation__.Notify(431398)
							setColVal(vec, outputIdx, d, s.evalCtx)
						case types.TimeTZFamily:
							__antithesis_instrumentation__.Notify(431399)
							setColVal(vec, outputIdx, tree.NewDTimeTZFromOffset(timeofday.FromInt(rng.Int63()), rng.Int31()), s.evalCtx)
						case types.TupleFamily:
							__antithesis_instrumentation__.Notify(431400)
							setColVal(vec, outputIdx, stringToDatum("(NULL)", vec.Type(), s.evalCtx), s.evalCtx)
						default:
							__antithesis_instrumentation__.Notify(431401)

							continue
						}
					default:
						__antithesis_instrumentation__.Notify(431392)
						if val, ok := quick.Value(reflect.TypeOf(vec.Col()).Elem(), rng); ok {
							__antithesis_instrumentation__.Notify(431404)
							setColVal(vec, outputIdx, val.Interface(), s.evalCtx)
						} else {
							__antithesis_instrumentation__.Notify(431405)
							colexecerror.InternalError(errors.AssertionFailedf("could not generate a random value of type %s", vec.Type()))
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(431406)
				}
			} else {
				__antithesis_instrumentation__.Notify(431407)
				setColVal(vec, outputIdx, tups[j][i], s.evalCtx)
			}
		}
	}
	__antithesis_instrumentation__.Notify(431352)

	s.batch.SetLength(batchSize)
	return s.batch
}

func (s *opTestInput) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(431408)
	s.tuples = s.initialTuples
}

type opFixedSelTestInput struct {
	colexecop.ZeroInputNode

	allocator *colmem.Allocator

	typs []*types.T

	batchSize int
	tuples    Tuples
	batch     coldata.Batch
	sel       []int
	evalCtx   *tree.EvalContext

	idx int
}

var _ colexecop.ResettableOperator = &opFixedSelTestInput{}

func NewOpFixedSelTestInput(
	allocator *colmem.Allocator, sel []int, batchSize int, tuples Tuples, typs []*types.T,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431409)
	ret := &opFixedSelTestInput{
		allocator: allocator,
		batchSize: batchSize,
		sel:       sel,
		tuples:    tuples,
		typs:      typs,
		evalCtx:   tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings()),
	}
	return ret
}

func (s *opFixedSelTestInput) Init(context.Context) {
	__antithesis_instrumentation__.Notify(431410)
	if s.typs == nil {
		__antithesis_instrumentation__.Notify(431414)
		if len(s.tuples) == 0 {
			__antithesis_instrumentation__.Notify(431416)
			colexecerror.InternalError(errors.AssertionFailedf("empty tuple source with no specified types"))
		} else {
			__antithesis_instrumentation__.Notify(431417)
		}
		__antithesis_instrumentation__.Notify(431415)
		s.typs = extrapolateTypesFromTuples(s.tuples)
	} else {
		__antithesis_instrumentation__.Notify(431418)
	}
	__antithesis_instrumentation__.Notify(431411)

	s.batch = s.allocator.NewMemBatchWithMaxCapacity(s.typs)
	tupleLen := len(s.tuples[0])
	for _, i := range s.sel {
		__antithesis_instrumentation__.Notify(431419)
		if len(s.tuples[i]) != tupleLen {
			__antithesis_instrumentation__.Notify(431420)
			colexecerror.InternalError(errors.AssertionFailedf("mismatched tuple lens: found %+v expected %d vals",
				s.tuples[i], tupleLen))
		} else {
			__antithesis_instrumentation__.Notify(431421)
		}
	}
	__antithesis_instrumentation__.Notify(431412)

	for i := 0; i < s.batch.Width(); i++ {
		__antithesis_instrumentation__.Notify(431422)
		s.batch.ColVec(i).Nulls().UnsetNulls()
	}
	__antithesis_instrumentation__.Notify(431413)

	if s.sel != nil {
		__antithesis_instrumentation__.Notify(431423)
		s.batch.SetSelection(true)

		for i := range s.typs {
			__antithesis_instrumentation__.Notify(431424)
			vec := s.batch.ColVec(i)

			for j := 0; j < len(s.tuples); j++ {
				__antithesis_instrumentation__.Notify(431425)
				if s.tuples[j][i] == nil {
					__antithesis_instrumentation__.Notify(431426)
					vec.Nulls().SetNull(j)
				} else {
					__antithesis_instrumentation__.Notify(431427)
					setColVal(vec, j, s.tuples[j][i], s.evalCtx)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(431428)
	}

}

func (s *opFixedSelTestInput) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431429)
	var batchSize int
	if s.sel == nil {
		__antithesis_instrumentation__.Notify(431431)
		batchSize = s.batchSize
		if len(s.tuples)-s.idx < batchSize {
			__antithesis_instrumentation__.Notify(431433)
			batchSize = len(s.tuples) - s.idx
		} else {
			__antithesis_instrumentation__.Notify(431434)
		}
		__antithesis_instrumentation__.Notify(431432)

		for i := range s.typs {
			__antithesis_instrumentation__.Notify(431435)
			vec := s.batch.ColVec(i)
			vec.Nulls().UnsetNulls()
			for j := 0; j < batchSize; j++ {
				__antithesis_instrumentation__.Notify(431436)
				if s.tuples[s.idx+j][i] == nil {
					__antithesis_instrumentation__.Notify(431437)
					vec.Nulls().SetNull(j)
				} else {
					__antithesis_instrumentation__.Notify(431438)

					setColVal(vec, j, s.tuples[s.idx+j][i], s.evalCtx)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(431439)
		if s.idx == len(s.sel) {
			__antithesis_instrumentation__.Notify(431442)
			return coldata.ZeroBatch
		} else {
			__antithesis_instrumentation__.Notify(431443)
		}
		__antithesis_instrumentation__.Notify(431440)
		batchSize = s.batchSize
		if len(s.sel)-s.idx < batchSize {
			__antithesis_instrumentation__.Notify(431444)
			batchSize = len(s.sel) - s.idx
		} else {
			__antithesis_instrumentation__.Notify(431445)
		}
		__antithesis_instrumentation__.Notify(431441)

		copy(s.batch.Selection(), s.sel[s.idx:s.idx+batchSize])
	}
	__antithesis_instrumentation__.Notify(431430)
	s.batch.SetLength(batchSize)
	s.idx += batchSize
	return s.batch
}

func (s *opFixedSelTestInput) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(431446)
	s.idx = 0
}

type OpTestOutput struct {
	colexecop.OneInputNode
	expected    Tuples
	evalCtx     *tree.EvalContext
	typs        []*types.T
	orderedCols []uint32

	curIdx int
	batch  coldata.Batch
}

func NewOpTestOutput(input colexecop.Operator, expected Tuples) *OpTestOutput {
	__antithesis_instrumentation__.Notify(431447)
	input.Init(context.Background())

	return &OpTestOutput{
		OneInputNode: colexecop.NewOneInputNode(input),
		expected:     expected,
		evalCtx:      tree.NewTestingEvalContext(cluster.MakeTestingClusterSettings()),
	}
}

func GetTupleFromBatch(batch coldata.Batch, tupleIdx int) Tuple {
	__antithesis_instrumentation__.Notify(431448)
	ret := make(Tuple, batch.Width())
	out := reflect.ValueOf(ret)
	if sel := batch.Selection(); sel != nil {
		__antithesis_instrumentation__.Notify(431451)
		tupleIdx = sel[tupleIdx]
	} else {
		__antithesis_instrumentation__.Notify(431452)
	}
	__antithesis_instrumentation__.Notify(431449)
	for colIdx := range ret {
		__antithesis_instrumentation__.Notify(431453)
		vec := batch.ColVec(colIdx)
		if vec.Nulls().NullAt(tupleIdx) {
			__antithesis_instrumentation__.Notify(431454)
			ret[colIdx] = nil
		} else {
			__antithesis_instrumentation__.Notify(431455)
			var val reflect.Value
			family := vec.CanonicalTypeFamily()
			if colBytes, ok := vec.Col().(*coldata.Bytes); ok {
				__antithesis_instrumentation__.Notify(431457)
				val = reflect.ValueOf(append([]byte(nil), colBytes.Get(tupleIdx)...))
			} else {
				__antithesis_instrumentation__.Notify(431458)
				if family == types.DecimalFamily {
					__antithesis_instrumentation__.Notify(431459)
					colDec := vec.Decimal()
					var newDec apd.Decimal
					newDec.Set(&colDec[tupleIdx])
					val = reflect.ValueOf(newDec)
				} else {
					__antithesis_instrumentation__.Notify(431460)
					if family == types.JsonFamily {
						__antithesis_instrumentation__.Notify(431461)
						colJSON := vec.JSON()
						newJSON := colJSON.Get(tupleIdx)
						b, err := json.EncodeJSON(nil, newJSON)
						if err != nil {
							__antithesis_instrumentation__.Notify(431464)
							colexecerror.ExpectedError(err)
						} else {
							__antithesis_instrumentation__.Notify(431465)
						}
						__antithesis_instrumentation__.Notify(431462)
						_, j, err := json.DecodeJSON(b)
						if err != nil {
							__antithesis_instrumentation__.Notify(431466)
							colexecerror.ExpectedError(err)
						} else {
							__antithesis_instrumentation__.Notify(431467)
						}
						__antithesis_instrumentation__.Notify(431463)
						val = reflect.ValueOf(j)
					} else {
						__antithesis_instrumentation__.Notify(431468)
						if family == typeconv.DatumVecCanonicalTypeFamily {
							__antithesis_instrumentation__.Notify(431469)
							val = reflect.ValueOf(vec.Datum().Get(tupleIdx).(tree.Datum))
						} else {
							__antithesis_instrumentation__.Notify(431470)
							val = reflect.ValueOf(vec.Col()).Index(tupleIdx)
						}
					}
				}
			}
			__antithesis_instrumentation__.Notify(431456)
			out.Index(colIdx).Set(val)
		}
	}
	__antithesis_instrumentation__.Notify(431450)
	return ret
}

func (r *OpTestOutput) next() Tuple {
	__antithesis_instrumentation__.Notify(431471)
	if r.batch == nil || func() bool {
		__antithesis_instrumentation__.Notify(431473)
		return r.curIdx >= r.batch.Length() == true
	}() == true {
		__antithesis_instrumentation__.Notify(431474)

		r.batch = r.Input.Next()
		if r.batch.Length() == 0 {
			__antithesis_instrumentation__.Notify(431476)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(431477)
		}
		__antithesis_instrumentation__.Notify(431475)
		r.curIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(431478)
	}
	__antithesis_instrumentation__.Notify(431472)
	ret := GetTupleFromBatch(r.batch, r.curIdx)
	r.curIdx++
	return ret
}

func (r *OpTestOutput) Reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431479)
	if r, ok := r.Input.(colexecop.Resetter); ok {
		__antithesis_instrumentation__.Notify(431481)
		r.Reset(ctx)
	} else {
		__antithesis_instrumentation__.Notify(431482)
	}
	__antithesis_instrumentation__.Notify(431480)
	r.curIdx = 0
	r.batch = nil
}

func (r *OpTestOutput) Verify() error {
	__antithesis_instrumentation__.Notify(431483)
	var actual Tuples
	for {
		__antithesis_instrumentation__.Notify(431485)
		var tup Tuple
		if err := colexecerror.CatchVectorizedRuntimeError(func() {
			__antithesis_instrumentation__.Notify(431488)
			tup = r.next()
		}); err != nil {
			__antithesis_instrumentation__.Notify(431489)
			return err
		} else {
			__antithesis_instrumentation__.Notify(431490)
		}
		__antithesis_instrumentation__.Notify(431486)
		if tup == nil {
			__antithesis_instrumentation__.Notify(431491)
			break
		} else {
			__antithesis_instrumentation__.Notify(431492)
		}
		__antithesis_instrumentation__.Notify(431487)
		actual = append(actual, tup)
	}
	__antithesis_instrumentation__.Notify(431484)
	return assertTuplesOrderedEqual(r.expected, actual, r.evalCtx)
}

func (r *OpTestOutput) VerifyAnyOrder() error {
	__antithesis_instrumentation__.Notify(431493)
	var actual Tuples
	for {
		__antithesis_instrumentation__.Notify(431495)
		var tup Tuple
		if err := colexecerror.CatchVectorizedRuntimeError(func() {
			__antithesis_instrumentation__.Notify(431498)
			tup = r.next()
		}); err != nil {
			__antithesis_instrumentation__.Notify(431499)
			return err
		} else {
			__antithesis_instrumentation__.Notify(431500)
		}
		__antithesis_instrumentation__.Notify(431496)
		if tup == nil {
			__antithesis_instrumentation__.Notify(431501)
			break
		} else {
			__antithesis_instrumentation__.Notify(431502)
		}
		__antithesis_instrumentation__.Notify(431497)
		actual = append(actual, tup)
	}
	__antithesis_instrumentation__.Notify(431494)
	return AssertTuplesSetsEqual(r.expected, actual, r.evalCtx)
}

func (r *OpTestOutput) VerifyPartialOrder() error {
	__antithesis_instrumentation__.Notify(431503)
	distincterInput := &colexecop.FeedOperator{}
	distincter, distinctOutput := OrderedDistinctColsToOperators(
		distincterInput, r.orderedCols, r.typs, false,
	)
	var actual, expected Tuples
	start := 0

	for {
		__antithesis_instrumentation__.Notify(431504)
		count := 0
		if err := colexecerror.CatchVectorizedRuntimeError(func() {
			__antithesis_instrumentation__.Notify(431509)
			for {
				__antithesis_instrumentation__.Notify(431510)

				if r.batch == nil || func() bool {
					__antithesis_instrumentation__.Notify(431513)
					return r.curIdx >= r.batch.Length() == true
				}() == true {
					__antithesis_instrumentation__.Notify(431514)

					r.batch = r.Input.Next()
					if r.batch.Length() == 0 {
						__antithesis_instrumentation__.Notify(431516)
						break
					} else {
						__antithesis_instrumentation__.Notify(431517)
					}
					__antithesis_instrumentation__.Notify(431515)
					distincterInput.SetBatch(r.batch)
					distincter.Next()
					r.curIdx = 0
				} else {
					__antithesis_instrumentation__.Notify(431518)
				}
				__antithesis_instrumentation__.Notify(431511)
				if distinctOutput[r.curIdx] && func() bool {
					__antithesis_instrumentation__.Notify(431519)
					return len(actual) > 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(431520)
					break
				} else {
					__antithesis_instrumentation__.Notify(431521)
				}
				__antithesis_instrumentation__.Notify(431512)
				ret := GetTupleFromBatch(r.batch, r.curIdx)
				actual = append(actual, ret)
				r.curIdx++
				count++
			}
		}); err != nil {
			__antithesis_instrumentation__.Notify(431522)
			return err
		} else {
			__antithesis_instrumentation__.Notify(431523)
		}
		__antithesis_instrumentation__.Notify(431505)
		if count == 0 {
			__antithesis_instrumentation__.Notify(431524)
			if len(r.expected) > start {
				__antithesis_instrumentation__.Notify(431526)
				expected = r.expected[start:]
				return makeError(expected, actual)
			} else {
				__antithesis_instrumentation__.Notify(431527)
			}
			__antithesis_instrumentation__.Notify(431525)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(431528)
		}
		__antithesis_instrumentation__.Notify(431506)
		if start+count > len(r.expected) {
			__antithesis_instrumentation__.Notify(431529)
			expected = r.expected[start:]
			return makeError(expected, actual)
		} else {
			__antithesis_instrumentation__.Notify(431530)
		}
		__antithesis_instrumentation__.Notify(431507)
		expected = r.expected[start : start+count]
		start = start + count
		if err := AssertTuplesSetsEqual(expected, actual, r.evalCtx); err != nil {
			__antithesis_instrumentation__.Notify(431531)
			return err
		} else {
			__antithesis_instrumentation__.Notify(431532)
		}
		__antithesis_instrumentation__.Notify(431508)
		actual = nil
		expected = nil
	}
}

func tupleEquals(expected Tuple, actual Tuple, evalCtx *tree.EvalContext) bool {
	__antithesis_instrumentation__.Notify(431533)
	if len(expected) != len(actual) {
		__antithesis_instrumentation__.Notify(431536)
		return false
	} else {
		__antithesis_instrumentation__.Notify(431537)
	}
	__antithesis_instrumentation__.Notify(431534)
	for i := 0; i < len(actual); i++ {
		__antithesis_instrumentation__.Notify(431538)
		expectedIsNull := expected[i] == nil || func() bool {
			__antithesis_instrumentation__.Notify(431539)
			return expected[i] == tree.DNull == true
		}() == true
		actualIsNull := actual[i] == nil || func() bool {
			__antithesis_instrumentation__.Notify(431540)
			return actual[i] == tree.DNull == true
		}() == true
		if expectedIsNull || func() bool {
			__antithesis_instrumentation__.Notify(431541)
			return actualIsNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(431542)
			if !expectedIsNull || func() bool {
				__antithesis_instrumentation__.Notify(431543)
				return !actualIsNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(431544)
				return false
			} else {
				__antithesis_instrumentation__.Notify(431545)
			}
		} else {
			__antithesis_instrumentation__.Notify(431546)

			if f1, ok := expected[i].(float64); ok {
				__antithesis_instrumentation__.Notify(431551)
				if f2, ok := actual[i].(float64); ok {
					__antithesis_instrumentation__.Notify(431552)
					if math.IsNaN(f1) && func() bool {
						__antithesis_instrumentation__.Notify(431553)
						return math.IsNaN(f2) == true
					}() == true {
						__antithesis_instrumentation__.Notify(431554)
						continue
					} else {
						__antithesis_instrumentation__.Notify(431555)
						if !math.IsNaN(f1) && func() bool {
							__antithesis_instrumentation__.Notify(431556)
							return !math.IsNaN(f2) == true
						}() == true && func() bool {
							__antithesis_instrumentation__.Notify(431557)
							return math.Abs(f1-f2) < 1e-6 == true
						}() == true {
							__antithesis_instrumentation__.Notify(431558)
							continue
						} else {
							__antithesis_instrumentation__.Notify(431559)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(431560)
				}
			} else {
				__antithesis_instrumentation__.Notify(431561)
			}
			__antithesis_instrumentation__.Notify(431547)

			if d1, ok := actual[i].(apd.Decimal); ok {
				__antithesis_instrumentation__.Notify(431562)
				if f2, ok := expected[i].(float64); ok {
					__antithesis_instrumentation__.Notify(431563)
					d2, _, err := apd.NewFromString(fmt.Sprintf("%f", f2))
					if err == nil && func() bool {
						__antithesis_instrumentation__.Notify(431564)
						return d1.Cmp(d2) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(431565)
						continue
					} else {
						__antithesis_instrumentation__.Notify(431566)
						return false
					}
				} else {
					__antithesis_instrumentation__.Notify(431567)
				}
			} else {
				__antithesis_instrumentation__.Notify(431568)
			}
			__antithesis_instrumentation__.Notify(431548)

			if j1, ok := actual[i].(json.JSON); ok {
				__antithesis_instrumentation__.Notify(431569)
				var j2 json.JSON
				switch t := expected[i].(type) {
				case string:
					__antithesis_instrumentation__.Notify(431572)
					var err error
					j2, err = json.ParseJSON(t)
					if err != nil {
						__antithesis_instrumentation__.Notify(431574)
						colexecerror.ExpectedError(err)
					} else {
						__antithesis_instrumentation__.Notify(431575)
					}
				case json.JSON:
					__antithesis_instrumentation__.Notify(431573)
					j2 = t
				}
				__antithesis_instrumentation__.Notify(431570)
				cmp, err := j1.Compare(j2)
				if err != nil {
					__antithesis_instrumentation__.Notify(431576)
					colexecerror.ExpectedError(err)
				} else {
					__antithesis_instrumentation__.Notify(431577)
				}
				__antithesis_instrumentation__.Notify(431571)
				if cmp == 0 {
					__antithesis_instrumentation__.Notify(431578)
					continue
				} else {
					__antithesis_instrumentation__.Notify(431579)
					return false
				}
			} else {
				__antithesis_instrumentation__.Notify(431580)
			}
			__antithesis_instrumentation__.Notify(431549)

			if d1, ok := actual[i].(tree.Datum); ok {
				__antithesis_instrumentation__.Notify(431581)
				var d2 tree.Datum
				switch d := expected[i].(type) {
				case tree.Datum:
					__antithesis_instrumentation__.Notify(431584)
					d2 = d
				case string:
					__antithesis_instrumentation__.Notify(431585)
					d2 = stringToDatum(d, d1.ResolvedType(), evalCtx)
				default:
					__antithesis_instrumentation__.Notify(431586)
					return false
				}
				__antithesis_instrumentation__.Notify(431582)
				if d1.Compare(evalCtx, d2) == 0 {
					__antithesis_instrumentation__.Notify(431587)
					continue
				} else {
					__antithesis_instrumentation__.Notify(431588)
				}
				__antithesis_instrumentation__.Notify(431583)
				return false
			} else {
				__antithesis_instrumentation__.Notify(431589)
			}
			__antithesis_instrumentation__.Notify(431550)

			if !reflect.DeepEqual(
				reflect.ValueOf(actual[i]).Convert(reflect.TypeOf(expected[i])).Interface(),
				expected[i],
			) || func() bool {
				__antithesis_instrumentation__.Notify(431590)
				return !reflect.DeepEqual(
					reflect.ValueOf(expected[i]).Convert(reflect.TypeOf(actual[i])).Interface(),
					actual[i],
				) == true
			}() == true {
				__antithesis_instrumentation__.Notify(431591)
				return false
			} else {
				__antithesis_instrumentation__.Notify(431592)
			}
		}
	}
	__antithesis_instrumentation__.Notify(431535)
	return true
}

func makeError(expected Tuples, actual Tuples) error {
	__antithesis_instrumentation__.Notify(431593)
	var expStr, actStr strings.Builder
	for i := range expected {
		__antithesis_instrumentation__.Notify(431597)
		expStr.WriteString(fmt.Sprintf("%d: %s\n", i, expected[i].String()))
	}
	__antithesis_instrumentation__.Notify(431594)
	for i := range actual {
		__antithesis_instrumentation__.Notify(431598)
		actStr.WriteString(fmt.Sprintf("%d: %s\n", i, actual[i].String()))
	}
	__antithesis_instrumentation__.Notify(431595)

	diff := difflib.UnifiedDiff{
		A:       difflib.SplitLines(expStr.String()),
		B:       difflib.SplitLines(actStr.String()),
		Context: 100,
	}
	text, err := difflib.GetUnifiedDiffString(diff)
	if err != nil {
		__antithesis_instrumentation__.Notify(431599)
		return errors.Wrap(err, "expected didn't match actual, failed to make diff")
	} else {
		__antithesis_instrumentation__.Notify(431600)
	}
	__antithesis_instrumentation__.Notify(431596)
	return errors.Errorf("expected didn't match actual. diff:\n%s", text)
}

func AssertTuplesSetsEqual(expected Tuples, actual Tuples, evalCtx *tree.EvalContext) error {
	__antithesis_instrumentation__.Notify(431601)
	if len(expected) != len(actual) {
		__antithesis_instrumentation__.Notify(431605)
		return makeError(expected, actual)
	} else {
		__antithesis_instrumentation__.Notify(431606)
	}
	__antithesis_instrumentation__.Notify(431602)
	var tupleFromOtherSet Tuple
	if len(expected) > 0 {
		__antithesis_instrumentation__.Notify(431607)
		tupleFromOtherSet = expected[0]
	} else {
		__antithesis_instrumentation__.Notify(431608)
	}
	__antithesis_instrumentation__.Notify(431603)
	actual = actual.sort(evalCtx, tupleFromOtherSet)
	if len(actual) > 0 {
		__antithesis_instrumentation__.Notify(431609)
		tupleFromOtherSet = actual[0]
	} else {
		__antithesis_instrumentation__.Notify(431610)
	}
	__antithesis_instrumentation__.Notify(431604)
	expected = expected.sort(evalCtx, tupleFromOtherSet)
	return assertTuplesOrderedEqual(expected, actual, evalCtx)
}

func assertTuplesOrderedEqual(expected Tuples, actual Tuples, evalCtx *tree.EvalContext) error {
	__antithesis_instrumentation__.Notify(431611)
	if len(expected) != len(actual) {
		__antithesis_instrumentation__.Notify(431614)
		return errors.Errorf("expected %+v, actual %+v", expected, actual)
	} else {
		__antithesis_instrumentation__.Notify(431615)
	}
	__antithesis_instrumentation__.Notify(431612)
	for i := range expected {
		__antithesis_instrumentation__.Notify(431616)
		if !tupleEquals(expected[i], actual[i], evalCtx) {
			__antithesis_instrumentation__.Notify(431617)
			return makeError(expected, actual)
		} else {
			__antithesis_instrumentation__.Notify(431618)
		}
	}
	__antithesis_instrumentation__.Notify(431613)
	return nil
}

type FiniteBatchSource struct {
	colexecop.ZeroInputNode

	repeatableBatch *colexecop.RepeatableBatchSource

	usableCount int
}

var _ colexecop.Operator = &FiniteBatchSource{}

func NewFiniteBatchSource(
	allocator *colmem.Allocator, batch coldata.Batch, typs []*types.T, usableCount int,
) *FiniteBatchSource {
	__antithesis_instrumentation__.Notify(431619)
	return &FiniteBatchSource{
		repeatableBatch: colexecop.NewRepeatableBatchSource(allocator, batch, typs),
		usableCount:     usableCount,
	}
}

func (f *FiniteBatchSource) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431620)
	f.repeatableBatch.Init(ctx)
}

func (f *FiniteBatchSource) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431621)
	if f.usableCount > 0 {
		__antithesis_instrumentation__.Notify(431623)
		f.usableCount--
		return f.repeatableBatch.Next()
	} else {
		__antithesis_instrumentation__.Notify(431624)
	}
	__antithesis_instrumentation__.Notify(431622)
	return coldata.ZeroBatch
}

func (f *FiniteBatchSource) Reset(usableCount int) {
	__antithesis_instrumentation__.Notify(431625)
	f.usableCount = usableCount
}

type finiteChunksSource struct {
	colexecop.ZeroInputNode
	repeatableBatch *colexecop.RepeatableBatchSource

	usableCount int
	matchLen    int
	adjustment  []int64
}

var _ colexecop.Operator = &finiteChunksSource{}

func NewFiniteChunksSource(
	allocator *colmem.Allocator, batch coldata.Batch, typs []*types.T, usableCount int, matchLen int,
) colexecop.Operator {
	__antithesis_instrumentation__.Notify(431626)
	return &finiteChunksSource{
		repeatableBatch: colexecop.NewRepeatableBatchSource(allocator, batch, typs),
		usableCount:     usableCount,
		matchLen:        matchLen,
	}
}

func (f *finiteChunksSource) Init(ctx context.Context) {
	__antithesis_instrumentation__.Notify(431627)
	f.repeatableBatch.Init(ctx)
	f.adjustment = make([]int64, f.matchLen)
}

func (f *finiteChunksSource) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431628)
	if f.usableCount > 0 {
		__antithesis_instrumentation__.Notify(431630)
		f.usableCount--
		batch := f.repeatableBatch.Next()
		if f.matchLen > 0 && func() bool {
			__antithesis_instrumentation__.Notify(431632)
			return f.adjustment[0] == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(431633)

			for col := 0; col < f.matchLen; col++ {
				__antithesis_instrumentation__.Notify(431634)
				firstValue := batch.ColVec(col).Int64()[0]
				lastValue := batch.ColVec(col).Int64()[batch.Length()-1]
				f.adjustment[col] = lastValue - firstValue + 1
			}
		} else {
			__antithesis_instrumentation__.Notify(431635)
			for i := 0; i < f.matchLen; i++ {
				__antithesis_instrumentation__.Notify(431636)
				int64Vec := batch.ColVec(i).Int64()
				for j := range int64Vec {
					__antithesis_instrumentation__.Notify(431638)
					int64Vec[j] += f.adjustment[i]
				}
				__antithesis_instrumentation__.Notify(431637)

				firstValue := batch.ColVec(i).Int64()[0]
				lastValue := batch.ColVec(i).Int64()[batch.Length()-1]
				f.adjustment[i] += lastValue - firstValue + 1
			}
		}
		__antithesis_instrumentation__.Notify(431631)
		return batch
	} else {
		__antithesis_instrumentation__.Notify(431639)
	}
	__antithesis_instrumentation__.Notify(431629)
	return coldata.ZeroBatch
}

type chunkingBatchSource struct {
	colexecop.ZeroInputNode
	allocator *colmem.Allocator
	typs      []*types.T
	cols      []coldata.Vec
	len       int

	curIdx int
	batch  coldata.Batch
}

var _ colexecop.ResettableOperator = &chunkingBatchSource{}

func NewChunkingBatchSource(
	allocator *colmem.Allocator, typs []*types.T, cols []coldata.Vec, len int,
) colexecop.ResettableOperator {
	__antithesis_instrumentation__.Notify(431640)
	return &chunkingBatchSource{
		allocator: allocator,
		typs:      typs,
		cols:      cols,
		len:       len,
	}
}

func (c *chunkingBatchSource) Init(context.Context) {
	__antithesis_instrumentation__.Notify(431641)
	c.batch = c.allocator.NewMemBatchWithMaxCapacity(c.typs)
	for i := range c.cols {
		__antithesis_instrumentation__.Notify(431642)
		c.batch.ColVec(i).SetCol(c.cols[i].Col())
		c.batch.ColVec(i).SetNulls(*c.cols[i].Nulls())
	}
}

func (c *chunkingBatchSource) Next() coldata.Batch {
	__antithesis_instrumentation__.Notify(431643)
	if c.curIdx >= c.len {
		__antithesis_instrumentation__.Notify(431647)
		return coldata.ZeroBatch
	} else {
		__antithesis_instrumentation__.Notify(431648)
	}
	__antithesis_instrumentation__.Notify(431644)

	c.batch.SetSelection(false)
	lastIdx := c.curIdx + coldata.BatchSize()
	if lastIdx > c.len {
		__antithesis_instrumentation__.Notify(431649)
		lastIdx = c.len
	} else {
		__antithesis_instrumentation__.Notify(431650)
	}
	__antithesis_instrumentation__.Notify(431645)
	for i := range c.typs {
		__antithesis_instrumentation__.Notify(431651)

		c.batch.ColVec(i).SetCol(c.cols[i].Window(c.curIdx, lastIdx).Col())
		c.batch.ColVec(i).SetNulls(c.cols[i].Nulls().Slice(c.curIdx, lastIdx))
	}
	__antithesis_instrumentation__.Notify(431646)
	c.batch.SetLength(lastIdx - c.curIdx)
	c.curIdx = lastIdx
	return c.batch
}

func (c *chunkingBatchSource) Reset(context.Context) {
	__antithesis_instrumentation__.Notify(431652)
	c.curIdx = 0
}

const MinBatchSize = 3

func GenerateBatchSize() int {
	__antithesis_instrumentation__.Notify(431653)
	randomizeBatchSize := envutil.EnvOrDefaultBool("COCKROACH_RANDOMIZE_BATCH_SIZE", true)
	if randomizeBatchSize {
		__antithesis_instrumentation__.Notify(431655)
		rng, _ := randutil.NewTestRand()

		var sizesToChooseFrom = []int{
			MinBatchSize,
			MinBatchSize + 1,
			MinBatchSize + 2,
			coldata.BatchSize(),
			MinBatchSize + rng.Intn(coldata.MaxBatchSize-MinBatchSize),
		}
		return sizesToChooseFrom[rng.Intn(len(sizesToChooseFrom))]
	} else {
		__antithesis_instrumentation__.Notify(431656)
	}
	__antithesis_instrumentation__.Notify(431654)
	return coldata.BatchSize()
}

type CallbackMetadataSource struct {
	DrainMetaCb func() []execinfrapb.ProducerMetadata
}

func (s CallbackMetadataSource) DrainMeta() []execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(431657)
	return s.DrainMetaCb()
}

func MakeRandWindowFrameRangeOffset(t *testing.T, rng *rand.Rand, typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(431658)
	isNegative := func(val tree.Datum) bool {
		__antithesis_instrumentation__.Notify(431660)
		switch datumTyp := val.(type) {
		case *tree.DInt:
			__antithesis_instrumentation__.Notify(431661)
			return int64(*datumTyp) < 0
		case *tree.DFloat:
			__antithesis_instrumentation__.Notify(431662)
			return float64(*datumTyp) < 0
		case *tree.DDecimal:
			__antithesis_instrumentation__.Notify(431663)
			return datumTyp.Negative
		case *tree.DInterval:
			__antithesis_instrumentation__.Notify(431664)
			return false
		default:
			__antithesis_instrumentation__.Notify(431665)
			t.Errorf("unexpected error: %v", errors.AssertionFailedf("unsupported datum: %v", datumTyp))
			return false
		}
	}
	__antithesis_instrumentation__.Notify(431659)

	for {
		__antithesis_instrumentation__.Notify(431666)
		val := randgen.RandDatumSimple(rng, typ)
		if isNegative(val) {
			__antithesis_instrumentation__.Notify(431668)

			continue
		} else {
			__antithesis_instrumentation__.Notify(431669)
		}
		__antithesis_instrumentation__.Notify(431667)
		return val
	}
}

func EncodeWindowFrameOffset(t *testing.T, offset tree.Datum) []byte {
	__antithesis_instrumentation__.Notify(431670)
	var encoded, scratch []byte
	encoded, err := valueside.Encode(
		encoded, valueside.NoColumnID, offset, scratch)
	require.NoError(t, err)
	return encoded
}
