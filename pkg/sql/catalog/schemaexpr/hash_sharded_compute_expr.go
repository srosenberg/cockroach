package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/sem/tree"

func MakeHashShardComputeExpr(colNames []string, buckets int) *string {
	__antithesis_instrumentation__.Notify(268240)
	unresolvedFunc := func(funcName string) tree.ResolvableFunctionReference {
		__antithesis_instrumentation__.Notify(268245)
		return tree.ResolvableFunctionReference{
			FunctionReference: &tree.UnresolvedName{
				NumParts: 1,
				Parts:    tree.NameParts{funcName},
			},
		}
	}
	__antithesis_instrumentation__.Notify(268241)
	columnItems := func() tree.Exprs {
		__antithesis_instrumentation__.Notify(268246)
		exprs := make(tree.Exprs, len(colNames))
		for i := range exprs {
			__antithesis_instrumentation__.Notify(268248)
			exprs[i] = &tree.ColumnItem{ColumnName: tree.Name(colNames[i])}
		}
		__antithesis_instrumentation__.Notify(268247)
		return exprs
	}
	__antithesis_instrumentation__.Notify(268242)
	hashedColumnsExpr := func() tree.Expr {
		__antithesis_instrumentation__.Notify(268249)
		return &tree.FuncExpr{
			Func: unresolvedFunc("fnv32"),
			Exprs: tree.Exprs{
				&tree.FuncExpr{
					Func:  unresolvedFunc("crdb_internal.datums_to_bytes"),
					Exprs: columnItems(),
				},
			},
		}
	}
	__antithesis_instrumentation__.Notify(268243)
	modBuckets := func(expr tree.Expr) tree.Expr {
		__antithesis_instrumentation__.Notify(268250)
		return &tree.FuncExpr{
			Func: unresolvedFunc("mod"),
			Exprs: tree.Exprs{
				expr,
				tree.NewDInt(tree.DInt(buckets)),
			},
		}
	}
	__antithesis_instrumentation__.Notify(268244)
	res := tree.Serialize(modBuckets(hashedColumnsExpr()))
	return &res
}
