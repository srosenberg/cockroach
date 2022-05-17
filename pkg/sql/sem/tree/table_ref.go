package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/sem/catid"

type ID = catid.ColumnID

type ColumnID = catid.ColumnID

type TableRef struct {
	TableID int64

	Columns []ColumnID

	As AliasClause
}

func (n *TableRef) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614503)
	ctx.Printf("[%d", n.TableID)
	if n.Columns != nil {
		__antithesis_instrumentation__.Notify(614506)
		ctx.WriteByte('(')
		for i, c := range n.Columns {
			__antithesis_instrumentation__.Notify(614508)
			if i > 0 {
				__antithesis_instrumentation__.Notify(614510)
				ctx.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(614511)
			}
			__antithesis_instrumentation__.Notify(614509)
			ctx.Printf("%d", c)
		}
		__antithesis_instrumentation__.Notify(614507)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(614512)
	}
	__antithesis_instrumentation__.Notify(614504)
	if n.As.Alias != "" {
		__antithesis_instrumentation__.Notify(614513)
		ctx.WriteString(" AS ")
		ctx.FormatNode(&n.As)
	} else {
		__antithesis_instrumentation__.Notify(614514)
	}
	__antithesis_instrumentation__.Notify(614505)
	ctx.WriteByte(']')
}
func (n *TableRef) String() string { __antithesis_instrumentation__.Notify(614515); return AsString(n) }

func (n *TableRef) tableExpr() { __antithesis_instrumentation__.Notify(614516) }
