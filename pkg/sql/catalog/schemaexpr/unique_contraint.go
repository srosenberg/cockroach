package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func ValidateUniqueWithoutIndexPredicate(
	ctx context.Context,
	tn tree.TableName,
	desc catalog.TableDescriptor,
	pred tree.Expr,
	semaCtx *tree.SemaContext,
) (string, error) {
	__antithesis_instrumentation__.Notify(268311)
	expr, _, _, err := DequalifyAndValidateExpr(
		ctx,
		desc,
		pred,
		types.Bool,
		"unique without index predicate",
		semaCtx,
		tree.VolatilityImmutable,
		&tn,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(268313)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(268314)
	}
	__antithesis_instrumentation__.Notify(268312)
	return expr, nil
}
