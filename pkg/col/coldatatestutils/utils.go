package coldatatestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func CopyBatch(
	original coldata.Batch, typs []*types.T, factory coldata.ColumnFactory,
) coldata.Batch {
	__antithesis_instrumentation__.Notify(54847)
	b := coldata.NewMemBatchWithCapacity(typs, original.Length(), factory)
	b.SetLength(original.Length())
	for colIdx, col := range original.ColVecs() {
		__antithesis_instrumentation__.Notify(54849)
		b.ColVec(colIdx).Copy(coldata.SliceArgs{
			Src:       col,
			SrcEndIdx: original.Length(),
		})
	}
	__antithesis_instrumentation__.Notify(54848)
	return b
}
