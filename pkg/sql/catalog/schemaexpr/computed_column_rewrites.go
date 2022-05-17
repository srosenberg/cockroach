package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

type ComputedColumnRewritesMap map[string]tree.Expr

func ParseComputedColumnRewrites(val string) (ComputedColumnRewritesMap, error) {
	__antithesis_instrumentation__.Notify(268052)
	if val == "" {
		__antithesis_instrumentation__.Notify(268057)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(268058)
	}
	__antithesis_instrumentation__.Notify(268053)
	stmt, err := parser.ParseOne(fmt.Sprintf("SET ROW (%s)", val))
	if err != nil {
		__antithesis_instrumentation__.Notify(268059)
		return nil, errors.Wrapf(err, "failed to parse column rewrites")
	} else {
		__antithesis_instrumentation__.Notify(268060)
	}
	__antithesis_instrumentation__.Notify(268054)
	set, ok := stmt.AST.(*tree.SetVar)
	if !ok {
		__antithesis_instrumentation__.Notify(268061)
		return nil, errors.AssertionFailedf("expected a SET statement, but found %T", stmt)
	} else {
		__antithesis_instrumentation__.Notify(268062)
	}
	__antithesis_instrumentation__.Notify(268055)
	result := make(ComputedColumnRewritesMap, len(set.Values))
	for _, v := range set.Values {
		__antithesis_instrumentation__.Notify(268063)
		binExpr, ok := v.(*tree.BinaryExpr)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(268067)
			return binExpr.Operator.Symbol != treebin.JSONFetchVal == true
		}() == true {
			__antithesis_instrumentation__.Notify(268068)
			return nil, errors.Newf("invalid column rewrites expression (expected -> operator)")
		} else {
			__antithesis_instrumentation__.Notify(268069)
		}
		__antithesis_instrumentation__.Notify(268064)
		left, ok := binExpr.Left.(*tree.ParenExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(268070)
			return nil, errors.Newf("missing parens around \"before\" expression")
		} else {
			__antithesis_instrumentation__.Notify(268071)
		}
		__antithesis_instrumentation__.Notify(268065)
		right, ok := binExpr.Right.(*tree.ParenExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(268072)
			return nil, errors.Newf("missing parens around \"after\" expression")
		} else {
			__antithesis_instrumentation__.Notify(268073)
		}
		__antithesis_instrumentation__.Notify(268066)
		result[tree.Serialize(left.Expr)] = right.Expr
	}
	__antithesis_instrumentation__.Notify(268056)
	return result, nil
}

func MaybeRewriteComputedColumn(expr tree.Expr, sessionData *sessiondata.SessionData) tree.Expr {
	__antithesis_instrumentation__.Notify(268074)
	rewritesStr := sessionData.ExperimentalComputedColumnRewrites
	if rewritesStr == "" {
		__antithesis_instrumentation__.Notify(268078)
		return expr
	} else {
		__antithesis_instrumentation__.Notify(268079)
	}
	__antithesis_instrumentation__.Notify(268075)
	rewrites, err := ParseComputedColumnRewrites(rewritesStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(268080)

		return expr
	} else {
		__antithesis_instrumentation__.Notify(268081)
	}
	__antithesis_instrumentation__.Notify(268076)
	if newExpr, ok := rewrites[tree.Serialize(expr)]; ok {
		__antithesis_instrumentation__.Notify(268082)
		return newExpr
	} else {
		__antithesis_instrumentation__.Notify(268083)
	}
	__antithesis_instrumentation__.Notify(268077)
	return expr
}
