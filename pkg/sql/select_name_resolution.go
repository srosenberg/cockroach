package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (p *planner) resolveNames(
	expr tree.Expr, source *colinfo.DataSourceInfo, ivarHelper tree.IndexedVarHelper,
) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(595744)
	if expr == nil {
		__antithesis_instrumentation__.Notify(595746)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(595747)
	}
	__antithesis_instrumentation__.Notify(595745)
	return schemaexpr.ResolveNamesUsingVisitor(&p.nameResolutionVisitor, expr, source, ivarHelper, p.SessionData().SearchPath)
}
