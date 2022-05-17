package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type With struct {
	Recursive bool
	CTEList   []*CTE
}

type CTE struct {
	Name AliasClause
	Mtr  MaterializeClause
	Stmt Statement
}

type MaterializeClause struct {
	Set bool

	Materialize bool
}

func (node *With) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(617141)
	if node == nil {
		__antithesis_instrumentation__.Notify(617145)
		return
	} else {
		__antithesis_instrumentation__.Notify(617146)
	}
	__antithesis_instrumentation__.Notify(617142)
	ctx.WriteString("WITH ")
	if node.Recursive {
		__antithesis_instrumentation__.Notify(617147)
		ctx.WriteString("RECURSIVE ")
	} else {
		__antithesis_instrumentation__.Notify(617148)
	}
	__antithesis_instrumentation__.Notify(617143)
	for i, cte := range node.CTEList {
		__antithesis_instrumentation__.Notify(617149)
		if i != 0 {
			__antithesis_instrumentation__.Notify(617152)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(617153)
		}
		__antithesis_instrumentation__.Notify(617150)
		ctx.FormatNode(&cte.Name)
		ctx.WriteString(" AS ")
		if cte.Mtr.Set {
			__antithesis_instrumentation__.Notify(617154)
			if !cte.Mtr.Materialize {
				__antithesis_instrumentation__.Notify(617156)
				ctx.WriteString("NOT ")
			} else {
				__antithesis_instrumentation__.Notify(617157)
			}
			__antithesis_instrumentation__.Notify(617155)
			ctx.WriteString("MATERIALIZED ")
		} else {
			__antithesis_instrumentation__.Notify(617158)
		}
		__antithesis_instrumentation__.Notify(617151)
		ctx.WriteString("(")
		ctx.FormatNode(cte.Stmt)
		ctx.WriteString(")")
	}
	__antithesis_instrumentation__.Notify(617144)
	ctx.WriteByte(' ')
}
