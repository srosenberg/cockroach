package colinfo

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
)

type ColumnOrderInfo struct {
	ColIdx    int
	Direction encoding.Direction
}

type ColumnOrdering []ColumnOrderInfo

func (ordering ColumnOrdering) String(columns ResultColumns) string {
	__antithesis_instrumentation__.Notify(251159)
	var buf bytes.Buffer
	fmtCtx := tree.NewFmtCtx(tree.FmtSimple)
	for i, o := range ordering {
		__antithesis_instrumentation__.Notify(251161)
		if i > 0 {
			__antithesis_instrumentation__.Notify(251164)
			buf.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(251165)
		}
		__antithesis_instrumentation__.Notify(251162)
		prefix := byte('+')
		if o.Direction == encoding.Descending {
			__antithesis_instrumentation__.Notify(251166)
			prefix = byte('-')
		} else {
			__antithesis_instrumentation__.Notify(251167)
		}
		__antithesis_instrumentation__.Notify(251163)
		buf.WriteByte(prefix)

		fmtCtx.FormatNameP(&columns[o.ColIdx].Name)
		_, _ = fmtCtx.WriteTo(&buf)
	}
	__antithesis_instrumentation__.Notify(251160)
	fmtCtx.Close()
	return buf.String()
}

var NoOrdering ColumnOrdering

func CompareDatums(ordering ColumnOrdering, evalCtx *tree.EvalContext, lhs, rhs tree.Datums) int {
	__antithesis_instrumentation__.Notify(251168)
	for _, c := range ordering {
		__antithesis_instrumentation__.Notify(251170)

		if cmp := lhs[c.ColIdx].Compare(evalCtx, rhs[c.ColIdx]); cmp != 0 {
			__antithesis_instrumentation__.Notify(251171)
			if c.Direction == encoding.Descending {
				__antithesis_instrumentation__.Notify(251173)
				cmp = -cmp
			} else {
				__antithesis_instrumentation__.Notify(251174)
			}
			__antithesis_instrumentation__.Notify(251172)
			return cmp
		} else {
			__antithesis_instrumentation__.Notify(251175)
		}
	}
	__antithesis_instrumentation__.Notify(251169)
	return 0
}
