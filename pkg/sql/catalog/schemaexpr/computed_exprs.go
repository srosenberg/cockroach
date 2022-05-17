package schemaexpr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type RowIndexedVarContainer struct {
	CurSourceRow tree.Datums

	Cols    []catalog.Column
	Mapping catalog.TableColMap
}

var _ tree.IndexedVarContainer = &RowIndexedVarContainer{}

func (r *RowIndexedVarContainer) IndexedVarEval(
	idx int, ctx *tree.EvalContext,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(268084)
	rowIdx, ok := r.Mapping.Get(r.Cols[idx].GetID())
	if !ok {
		__antithesis_instrumentation__.Notify(268086)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(268087)
	}
	__antithesis_instrumentation__.Notify(268085)
	return r.CurSourceRow[rowIdx], nil
}

func (*RowIndexedVarContainer) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(268088)
	panic("unsupported")
}

func (*RowIndexedVarContainer) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(268089)
	return nil
}

func CannotWriteToComputedColError(colName string) error {
	__antithesis_instrumentation__.Notify(268090)
	return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
		"cannot write directly to computed column %q", tree.ErrNameString(colName))
}
