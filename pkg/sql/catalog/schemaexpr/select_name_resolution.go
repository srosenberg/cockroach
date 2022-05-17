package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type NameResolutionVisitor struct {
	err        error
	iVarHelper tree.IndexedVarHelper
	searchPath sessiondata.SearchPath
	resolver   colinfo.ColumnResolver
}

var _ tree.Visitor = &NameResolutionVisitor{}

func (v *NameResolutionVisitor) VisitPre(expr tree.Expr) (recurse bool, newNode tree.Expr) {
	__antithesis_instrumentation__.Notify(268282)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(268285)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(268286)
	}
	__antithesis_instrumentation__.Notify(268283)

	switch t := expr.(type) {
	case *tree.IndexedVar:
		__antithesis_instrumentation__.Notify(268287)

		t, v.err = v.iVarHelper.BindIfUnbound(t)
		if v.err != nil {
			__antithesis_instrumentation__.Notify(268297)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(268298)
		}
		__antithesis_instrumentation__.Notify(268288)

		return false, t

	case *tree.UnresolvedName:
		__antithesis_instrumentation__.Notify(268289)
		vn, err := t.NormalizeVarName()
		if err != nil {
			__antithesis_instrumentation__.Notify(268299)
			v.err = err
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(268300)
		}
		__antithesis_instrumentation__.Notify(268290)
		return v.VisitPre(vn)

	case *tree.ColumnItem:
		__antithesis_instrumentation__.Notify(268291)
		_, err := colinfo.ResolveColumnItem(context.TODO(), &v.resolver, t)
		if err != nil {
			__antithesis_instrumentation__.Notify(268301)
			v.err = err
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(268302)
		}
		__antithesis_instrumentation__.Notify(268292)

		colIdx := v.resolver.ResolverState.ColIdx
		ivar := v.iVarHelper.IndexedVar(colIdx)
		return true, ivar

	case *tree.FuncExpr:
		__antithesis_instrumentation__.Notify(268293)

		if len(t.Exprs) != 1 {
			__antithesis_instrumentation__.Notify(268303)
			break
		} else {
			__antithesis_instrumentation__.Notify(268304)
		}
		__antithesis_instrumentation__.Notify(268294)
		vn, ok := t.Exprs[0].(tree.VarName)
		if !ok {
			__antithesis_instrumentation__.Notify(268305)
			break
		} else {
			__antithesis_instrumentation__.Notify(268306)
		}
		__antithesis_instrumentation__.Notify(268295)
		vn, v.err = vn.NormalizeVarName()
		if v.err != nil {
			__antithesis_instrumentation__.Notify(268307)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(268308)
		}
		__antithesis_instrumentation__.Notify(268296)

		t.Exprs[0] = vn
		return true, t
	}
	__antithesis_instrumentation__.Notify(268284)

	return true, expr
}

func (*NameResolutionVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(268309)
	return expr
}

func ResolveNamesUsingVisitor(
	v *NameResolutionVisitor,
	expr tree.Expr,
	source *colinfo.DataSourceInfo,
	ivarHelper tree.IndexedVarHelper,
	searchPath sessiondata.SearchPath,
) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(268310)
	*v = NameResolutionVisitor{
		iVarHelper: ivarHelper,
		searchPath: searchPath,
		resolver: colinfo.ColumnResolver{
			Source: source,
		},
	}

	expr, _ = tree.WalkExpr(v, expr)
	return expr, v.err
}
