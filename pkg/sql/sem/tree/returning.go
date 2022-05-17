package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ReturningClause interface {
	NodeFormatter

	statementReturnType() StatementReturnType
	returningClause()
}

var _ ReturningClause = &ReturningExprs{}
var _ ReturningClause = &ReturningNothing{}
var _ ReturningClause = &NoReturningClause{}

type ReturningExprs SelectExprs

func (r *ReturningExprs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612932)
	ctx.WriteString("RETURNING ")
	ctx.FormatNode((*SelectExprs)(r))
}

var ReturningNothingClause = &ReturningNothing{}

type ReturningNothing struct{}

func (*ReturningNothing) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612933)
	ctx.WriteString("RETURNING NOTHING")
}

var AbsentReturningClause = &NoReturningClause{}

type NoReturningClause struct{}

func (*NoReturningClause) Format(_ *FmtCtx) { __antithesis_instrumentation__.Notify(612934) }

func (*ReturningExprs) statementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(612935)
	return Rows
}
func (*ReturningNothing) statementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(612936)
	return RowsAffected
}
func (*NoReturningClause) statementReturnType() StatementReturnType {
	__antithesis_instrumentation__.Notify(612937)
	return RowsAffected
}

func (*ReturningExprs) returningClause()    { __antithesis_instrumentation__.Notify(612938) }
func (*ReturningNothing) returningClause()  { __antithesis_instrumentation__.Notify(612939) }
func (*NoReturningClause) returningClause() { __antithesis_instrumentation__.Notify(612940) }

func HasReturningClause(clause ReturningClause) bool {
	__antithesis_instrumentation__.Notify(612941)
	_, ok := clause.(*NoReturningClause)
	return !ok
}
