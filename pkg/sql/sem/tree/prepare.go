package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/lexbase"

type Prepare struct {
	Name      Name
	Types     []ResolvableTypeReference
	Statement Statement
}

func (node *Prepare) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611832)
	ctx.WriteString("PREPARE ")
	ctx.FormatNode(&node.Name)
	if len(node.Types) > 0 {
		__antithesis_instrumentation__.Notify(611834)
		ctx.WriteString(" (")
		for i, t := range node.Types {
			__antithesis_instrumentation__.Notify(611836)
			if i > 0 {
				__antithesis_instrumentation__.Notify(611838)
				ctx.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(611839)
			}
			__antithesis_instrumentation__.Notify(611837)
			ctx.WriteString(t.SQLString())
		}
		__antithesis_instrumentation__.Notify(611835)
		ctx.WriteRune(')')
	} else {
		__antithesis_instrumentation__.Notify(611840)
	}
	__antithesis_instrumentation__.Notify(611833)
	ctx.WriteString(" AS ")
	ctx.FormatNode(node.Statement)
}

type CannedOptPlan struct {
	Plan string
}

func (node *CannedOptPlan) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611841)

	ctx.WriteString("OPT PLAN ")
	ctx.WriteString(lexbase.EscapeSQLString(node.Plan))
}

type Execute struct {
	Name   Name
	Params Exprs

	DiscardRows bool
}

func (node *Execute) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611842)
	ctx.WriteString("EXECUTE ")
	ctx.FormatNode(&node.Name)
	if len(node.Params) > 0 {
		__antithesis_instrumentation__.Notify(611844)
		ctx.WriteString(" (")
		ctx.FormatNode(&node.Params)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(611845)
	}
	__antithesis_instrumentation__.Notify(611843)
	if node.DiscardRows {
		__antithesis_instrumentation__.Notify(611846)
		ctx.WriteString(" DISCARD ROWS")
	} else {
		__antithesis_instrumentation__.Notify(611847)
	}
}

type Deallocate struct {
	Name Name
}

func (node *Deallocate) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(611848)
	ctx.WriteString("DEALLOCATE ")
	if node.Name == "" {
		__antithesis_instrumentation__.Notify(611849)
		ctx.WriteString("ALL")
	} else {
		__antithesis_instrumentation__.Notify(611850)

		if ctx.HasFlags(FmtHideConstants) {
			__antithesis_instrumentation__.Notify(611851)
			ctx.WriteByte('_')
		} else {
			__antithesis_instrumentation__.Notify(611852)
			ctx.FormatNode(&node.Name)
		}
	}
}
