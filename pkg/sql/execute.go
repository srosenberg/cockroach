package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func (p *planner) fillInPlaceholders(
	ctx context.Context, ps *PreparedStatement, name string, params tree.Exprs,
) (*tree.PlaceholderInfo, error) {
	__antithesis_instrumentation__.Notify(490950)
	if len(ps.Types) != len(params) {
		__antithesis_instrumentation__.Notify(490953)
		return nil, pgerror.Newf(pgcode.Syntax,
			"wrong number of parameters for prepared statement %q: expected %d, got %d",
			name, len(ps.Types), len(params))
	} else {
		__antithesis_instrumentation__.Notify(490954)
	}
	__antithesis_instrumentation__.Notify(490951)

	qArgs := make(tree.QueryArguments, len(params))
	var semaCtx tree.SemaContext
	for i, e := range params {
		__antithesis_instrumentation__.Notify(490955)
		idx := tree.PlaceholderIdx(i)

		typ, ok := ps.ValueType(idx)
		if !ok {
			__antithesis_instrumentation__.Notify(490959)
			return nil, errors.AssertionFailedf("no type for placeholder %s", idx)
		} else {
			__antithesis_instrumentation__.Notify(490960)
		}
		__antithesis_instrumentation__.Notify(490956)

		if typ.UserDefined() {
			__antithesis_instrumentation__.Notify(490961)
			var err error
			typ, err = p.ResolveTypeByOID(ctx, typ.Oid())
			if err != nil {
				__antithesis_instrumentation__.Notify(490962)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(490963)
			}
		} else {
			__antithesis_instrumentation__.Notify(490964)
		}
		__antithesis_instrumentation__.Notify(490957)
		typedExpr, err := schemaexpr.SanitizeVarFreeExpr(
			ctx, e, typ, "EXECUTE parameter", &semaCtx, tree.VolatilityVolatile,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(490965)
			return nil, pgerror.WithCandidateCode(err, pgcode.WrongObjectType)
		} else {
			__antithesis_instrumentation__.Notify(490966)
		}
		__antithesis_instrumentation__.Notify(490958)

		qArgs[idx] = typedExpr
	}
	__antithesis_instrumentation__.Notify(490952)
	return &tree.PlaceholderInfo{
		Values: qArgs,
		PlaceholderTypesInfo: tree.PlaceholderTypesInfo{
			TypeHints: ps.TypeHints,
			Types:     ps.Types,
		},
	}, nil
}
