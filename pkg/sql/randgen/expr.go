package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
)

func randPartialIndexPredicateFromCols(
	rng *rand.Rand, columnTableDefs []*tree.ColumnTableDef, tableName *tree.TableName,
) tree.Expr {
	__antithesis_instrumentation__.Notify(564000)

	cpy := make([]*tree.ColumnTableDef, len(columnTableDefs))
	copy(cpy, columnTableDefs)
	rng.Shuffle(len(cpy), func(i, j int) { __antithesis_instrumentation__.Notify(564004); cpy[i], cpy[j] = cpy[j], cpy[i] })
	__antithesis_instrumentation__.Notify(564001)

	nCols := rng.Intn(len(cpy)) + 1
	cols := make([]*tree.ColumnTableDef, 0, nCols)
	for _, col := range cpy {
		__antithesis_instrumentation__.Notify(564005)
		if isAllowedPartialIndexColType(col) {
			__antithesis_instrumentation__.Notify(564007)
			cols = append(cols, col)
		} else {
			__antithesis_instrumentation__.Notify(564008)
		}
		__antithesis_instrumentation__.Notify(564006)
		if len(cols) == nCols {
			__antithesis_instrumentation__.Notify(564009)
			break
		} else {
			__antithesis_instrumentation__.Notify(564010)
		}
	}
	__antithesis_instrumentation__.Notify(564002)

	var e tree.Expr
	for _, columnTableDef := range cols {
		__antithesis_instrumentation__.Notify(564011)
		expr := randBoolColumnExpr(rng, columnTableDef, tableName)

		if e != nil {
			__antithesis_instrumentation__.Notify(564013)
			expr = randAndOrExpr(rng, e, expr)
		} else {
			__antithesis_instrumentation__.Notify(564014)
		}
		__antithesis_instrumentation__.Notify(564012)
		e = expr
	}
	__antithesis_instrumentation__.Notify(564003)
	return e
}

func isAllowedPartialIndexColType(columnTableDef *tree.ColumnTableDef) bool {
	__antithesis_instrumentation__.Notify(564015)
	switch fam := columnTableDef.Type.(*types.T).Family(); fam {
	case types.BoolFamily, types.IntFamily, types.FloatFamily, types.DecimalFamily,
		types.StringFamily, types.DateFamily, types.TimeFamily, types.TimeTZFamily,
		types.TimestampFamily, types.TimestampTZFamily, types.BytesFamily:
		__antithesis_instrumentation__.Notify(564016)
		return true
	default:
		__antithesis_instrumentation__.Notify(564017)
		return false
	}
}

var cmpOps = []treecmp.ComparisonOperatorSymbol{treecmp.EQ, treecmp.NE, treecmp.LT, treecmp.LE, treecmp.GE, treecmp.GT}

func randBoolColumnExpr(
	rng *rand.Rand, columnTableDef *tree.ColumnTableDef, tableName *tree.TableName,
) tree.Expr {
	__antithesis_instrumentation__.Notify(564018)
	varExpr := tree.NewColumnItem(tableName, columnTableDef.Name)
	t := columnTableDef.Type.(*types.T)

	if t.Family() == types.BoolFamily {
		__antithesis_instrumentation__.Notify(564020)
		if rng.Intn(2) == 0 {
			__antithesis_instrumentation__.Notify(564022)
			return &tree.NotExpr{Expr: varExpr}
		} else {
			__antithesis_instrumentation__.Notify(564023)
		}
		__antithesis_instrumentation__.Notify(564021)
		return varExpr
	} else {
		__antithesis_instrumentation__.Notify(564024)
	}
	__antithesis_instrumentation__.Notify(564019)

	op := treecmp.MakeComparisonOperator(cmpOps[rng.Intn(len(cmpOps))])
	datum := randInterestingDatum(rng, t)
	return &tree.ComparisonExpr{Operator: op, Left: varExpr, Right: datum}
}

func randAndOrExpr(rng *rand.Rand, left, right tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(564025)
	if rng.Intn(2) == 0 {
		__antithesis_instrumentation__.Notify(564027)
		return &tree.OrExpr{
			Left:  left,
			Right: right,
		}
	} else {
		__antithesis_instrumentation__.Notify(564028)
	}
	__antithesis_instrumentation__.Notify(564026)

	return &tree.AndExpr{
		Left:  left,
		Right: right,
	}
}

func randExpr(
	rng *rand.Rand, normalColDefs []*tree.ColumnTableDef, nullOk bool,
) (_ tree.Expr, _ *types.T, _ tree.Nullability, referencedCols map[tree.Name]struct{}) {
	__antithesis_instrumentation__.Notify(564029)
	nullability := tree.NotNull
	referencedCols = make(map[tree.Name]struct{})

	if rng.Intn(2) == 0 {
		__antithesis_instrumentation__.Notify(564032)

		var cols []*tree.ColumnTableDef
		var fam types.Family
		for _, idx := range rng.Perm(len(normalColDefs)) {
			__antithesis_instrumentation__.Notify(564034)
			x := normalColDefs[idx]
			xFam := x.Type.(*types.T).Family()

			if len(cols) == 0 {
				__antithesis_instrumentation__.Notify(564035)
				switch xFam {
				case types.IntFamily, types.FloatFamily, types.DecimalFamily:
					__antithesis_instrumentation__.Notify(564036)
					fam = xFam
					cols = append(cols, x)
				default:
					__antithesis_instrumentation__.Notify(564037)
				}
			} else {
				__antithesis_instrumentation__.Notify(564038)
				if fam == xFam {
					__antithesis_instrumentation__.Notify(564039)
					cols = append(cols, x)
					if len(cols) > 1 && func() bool {
						__antithesis_instrumentation__.Notify(564040)
						return rng.Intn(2) == 0 == true
					}() == true {
						__antithesis_instrumentation__.Notify(564041)
						break
					} else {
						__antithesis_instrumentation__.Notify(564042)
					}
				} else {
					__antithesis_instrumentation__.Notify(564043)
				}
			}
		}
		__antithesis_instrumentation__.Notify(564033)
		if len(cols) > 1 {
			__antithesis_instrumentation__.Notify(564044)

			for _, x := range cols {
				__antithesis_instrumentation__.Notify(564047)
				if x.Nullable.Nullability != tree.NotNull {
					__antithesis_instrumentation__.Notify(564048)
					nullability = x.Nullable.Nullability
					break
				} else {
					__antithesis_instrumentation__.Notify(564049)
				}
			}
			__antithesis_instrumentation__.Notify(564045)

			var expr tree.Expr
			expr = tree.NewUnresolvedName(string(cols[0].Name))
			referencedCols[cols[0].Name] = struct{}{}
			for _, x := range cols[1:] {
				__antithesis_instrumentation__.Notify(564050)
				expr = &tree.BinaryExpr{
					Operator: treebin.MakeBinaryOperator(treebin.Plus),
					Left:     expr,
					Right:    tree.NewUnresolvedName(string(x.Name)),
				}
				referencedCols[x.Name] = struct{}{}
			}
			__antithesis_instrumentation__.Notify(564046)
			return expr, cols[0].Type.(*types.T), nullability, referencedCols
		} else {
			__antithesis_instrumentation__.Notify(564051)
		}
	} else {
		__antithesis_instrumentation__.Notify(564052)
	}
	__antithesis_instrumentation__.Notify(564030)

	x := normalColDefs[randutil.RandIntInRange(rng, 0, len(normalColDefs))]
	xTyp := x.Type.(*types.T)
	referencedCols[x.Name] = struct{}{}

	nullability = x.Nullable.Nullability
	nullOk = nullOk && func() bool {
		__antithesis_instrumentation__.Notify(564053)
		return nullability != tree.NotNull == true
	}() == true

	var expr tree.Expr
	var typ *types.T
	switch xTyp.Family() {
	case types.IntFamily, types.FloatFamily, types.DecimalFamily:
		__antithesis_instrumentation__.Notify(564054)
		typ = xTyp
		expr = &tree.BinaryExpr{
			Operator: treebin.MakeBinaryOperator(treebin.Plus),
			Left:     tree.NewUnresolvedName(string(x.Name)),
			Right:    RandDatum(rng, xTyp, nullOk),
		}

	case types.StringFamily:
		__antithesis_instrumentation__.Notify(564055)
		typ = types.String
		expr = &tree.FuncExpr{
			Func:  tree.WrapFunction("lower"),
			Exprs: tree.Exprs{tree.NewUnresolvedName(string(x.Name))},
		}

	default:
		__antithesis_instrumentation__.Notify(564056)
		volatility, ok := tree.LookupCastVolatility(xTyp, types.String, nil)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(564057)
			return volatility <= tree.VolatilityImmutable == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(564058)
			return !typeToStringCastHasIncorrectVolatility(xTyp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(564059)

			typ = types.String
			expr = &tree.FuncExpr{
				Func: tree.WrapFunction("lower"),
				Exprs: tree.Exprs{
					&tree.CastExpr{
						Expr: tree.NewUnresolvedName(string(x.Name)),
						Type: types.String,
					},
				},
			}
		} else {
			__antithesis_instrumentation__.Notify(564060)

			typ = types.String
			expr = &tree.CaseExpr{
				Whens: []*tree.When{
					{
						Cond: &tree.IsNullExpr{
							Expr: tree.NewUnresolvedName(string(x.Name)),
						},
						Val: RandDatum(rng, types.String, nullOk),
					},
				},
				Else: RandDatum(rng, types.String, nullOk),
			}
		}
	}
	__antithesis_instrumentation__.Notify(564031)

	return expr, typ, nullability, referencedCols
}

func typeToStringCastHasIncorrectVolatility(t *types.T) bool {
	__antithesis_instrumentation__.Notify(564061)
	switch t.Family() {
	case types.DateFamily, types.EnumFamily, types.TimestampFamily,
		types.IntervalFamily, types.TupleFamily:
		__antithesis_instrumentation__.Notify(564062)
		return true
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(564063)
		return t == types.RegClass || func() bool {
			__antithesis_instrumentation__.Notify(564065)
			return t == types.RegNamespace == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(564066)
			return t == types.RegProc == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(564067)
			return t == types.RegProcedure == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(564068)
			return t == types.RegRole == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(564069)
			return t == types.RegType == true
		}() == true
	default:
		__antithesis_instrumentation__.Notify(564064)
		return false
	}
}
