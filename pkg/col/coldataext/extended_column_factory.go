package coldataext

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type extendedColumnFactory struct {
	evalCtx *tree.EvalContext
}

var _ coldata.ColumnFactory = &extendedColumnFactory{}

func NewExtendedColumnFactory(evalCtx *tree.EvalContext) coldata.ColumnFactory {
	__antithesis_instrumentation__.Notify(54705)
	return &extendedColumnFactory{evalCtx: evalCtx}
}

func (cf *extendedColumnFactory) MakeColumn(t *types.T, n int) coldata.Column {
	__antithesis_instrumentation__.Notify(54706)
	if typeconv.TypeFamilyToCanonicalTypeFamily(t.Family()) == typeconv.DatumVecCanonicalTypeFamily {
		__antithesis_instrumentation__.Notify(54708)
		return newDatumVec(t, n, cf.evalCtx)
	} else {
		__antithesis_instrumentation__.Notify(54709)
	}
	__antithesis_instrumentation__.Notify(54707)
	return coldata.StandardColumnFactory.MakeColumn(t, n)
}
