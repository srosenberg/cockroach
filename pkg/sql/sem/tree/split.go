package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Split struct {
	TableOrIndex TableIndexName

	Rows *Select

	ExpireExpr Expr
}

func (node *Split) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613664)
	ctx.WriteString("ALTER ")
	if node.TableOrIndex.Index != "" {
		__antithesis_instrumentation__.Notify(613666)
		ctx.WriteString("INDEX ")
	} else {
		__antithesis_instrumentation__.Notify(613667)
		ctx.WriteString("TABLE ")
	}
	__antithesis_instrumentation__.Notify(613665)
	ctx.FormatNode(&node.TableOrIndex)
	ctx.WriteString(" SPLIT AT ")
	ctx.FormatNode(node.Rows)
	if node.ExpireExpr != nil {
		__antithesis_instrumentation__.Notify(613668)
		ctx.WriteString(" WITH EXPIRATION ")
		ctx.FormatNode(node.ExpireExpr)
	} else {
		__antithesis_instrumentation__.Notify(613669)
	}
}

type Unsplit struct {
	TableOrIndex TableIndexName

	Rows *Select
	All  bool
}

func (node *Unsplit) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613670)
	ctx.WriteString("ALTER ")
	if node.TableOrIndex.Index != "" {
		__antithesis_instrumentation__.Notify(613672)
		ctx.WriteString("INDEX ")
	} else {
		__antithesis_instrumentation__.Notify(613673)
		ctx.WriteString("TABLE ")
	}
	__antithesis_instrumentation__.Notify(613671)
	ctx.FormatNode(&node.TableOrIndex)
	if node.All {
		__antithesis_instrumentation__.Notify(613674)
		ctx.WriteString(" UNSPLIT ALL")
	} else {
		__antithesis_instrumentation__.Notify(613675)
		ctx.WriteString(" UNSPLIT AT ")
		ctx.FormatNode(node.Rows)
	}
}

type Relocate struct {
	TableOrIndex TableIndexName

	Rows            *Select
	SubjectReplicas RelocateSubject
}

func (node *Relocate) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613676)
	ctx.WriteString("ALTER ")
	if node.TableOrIndex.Index != "" {
		__antithesis_instrumentation__.Notify(613678)
		ctx.WriteString("INDEX ")
	} else {
		__antithesis_instrumentation__.Notify(613679)
		ctx.WriteString("TABLE ")
	}
	__antithesis_instrumentation__.Notify(613677)
	ctx.FormatNode(&node.TableOrIndex)
	ctx.WriteString(" RELOCATE ")
	ctx.FormatNode(&node.SubjectReplicas)
	ctx.WriteByte(' ')
	ctx.FormatNode(node.Rows)
}

type Scatter struct {
	TableOrIndex TableIndexName

	From, To Exprs
}

func (node *Scatter) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613680)
	ctx.WriteString("ALTER ")
	if node.TableOrIndex.Index != "" {
		__antithesis_instrumentation__.Notify(613682)
		ctx.WriteString("INDEX ")
	} else {
		__antithesis_instrumentation__.Notify(613683)
		ctx.WriteString("TABLE ")
	}
	__antithesis_instrumentation__.Notify(613681)
	ctx.FormatNode(&node.TableOrIndex)
	ctx.WriteString(" SCATTER")
	if node.From != nil {
		__antithesis_instrumentation__.Notify(613684)
		ctx.WriteString(" FROM (")
		ctx.FormatNode(&node.From)
		ctx.WriteString(") TO (")
		ctx.FormatNode(&node.To)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(613685)
	}
}
