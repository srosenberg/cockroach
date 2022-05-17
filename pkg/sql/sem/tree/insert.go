package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Insert struct {
	With       *With
	Table      TableExpr
	Columns    NameList
	Rows       *Select
	OnConflict *OnConflict
	Returning  ReturningClause
}

func (node *Insert) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(610070)
	ctx.FormatNode(node.With)
	if node.OnConflict.IsUpsertAlias() {
		__antithesis_instrumentation__.Notify(610075)
		ctx.WriteString("UPSERT")
	} else {
		__antithesis_instrumentation__.Notify(610076)
		ctx.WriteString("INSERT")
	}
	__antithesis_instrumentation__.Notify(610071)
	ctx.WriteString(" INTO ")
	ctx.FormatNode(node.Table)
	if node.Columns != nil {
		__antithesis_instrumentation__.Notify(610077)
		ctx.WriteByte('(')
		ctx.FormatNode(&node.Columns)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(610078)
	}
	__antithesis_instrumentation__.Notify(610072)
	if node.DefaultValues() {
		__antithesis_instrumentation__.Notify(610079)
		ctx.WriteString(" DEFAULT VALUES")
	} else {
		__antithesis_instrumentation__.Notify(610080)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Rows)
	}
	__antithesis_instrumentation__.Notify(610073)
	if node.OnConflict != nil && func() bool {
		__antithesis_instrumentation__.Notify(610081)
		return !node.OnConflict.IsUpsertAlias() == true
	}() == true {
		__antithesis_instrumentation__.Notify(610082)
		ctx.WriteString(" ON CONFLICT")
		if node.OnConflict.Constraint != "" {
			__antithesis_instrumentation__.Notify(610086)
			ctx.WriteString(" ON CONSTRAINT ")
			ctx.FormatNode(&node.OnConflict.Constraint)
		} else {
			__antithesis_instrumentation__.Notify(610087)
		}
		__antithesis_instrumentation__.Notify(610083)
		if len(node.OnConflict.Columns) > 0 {
			__antithesis_instrumentation__.Notify(610088)
			ctx.WriteString(" (")
			ctx.FormatNode(&node.OnConflict.Columns)
			ctx.WriteString(")")
		} else {
			__antithesis_instrumentation__.Notify(610089)
		}
		__antithesis_instrumentation__.Notify(610084)
		if node.OnConflict.ArbiterPredicate != nil {
			__antithesis_instrumentation__.Notify(610090)
			ctx.WriteString(" WHERE ")
			ctx.FormatNode(node.OnConflict.ArbiterPredicate)
		} else {
			__antithesis_instrumentation__.Notify(610091)
		}
		__antithesis_instrumentation__.Notify(610085)
		if node.OnConflict.DoNothing {
			__antithesis_instrumentation__.Notify(610092)
			ctx.WriteString(" DO NOTHING")
		} else {
			__antithesis_instrumentation__.Notify(610093)
			ctx.WriteString(" DO UPDATE SET ")
			ctx.FormatNode(&node.OnConflict.Exprs)
			if node.OnConflict.Where != nil {
				__antithesis_instrumentation__.Notify(610094)
				ctx.WriteByte(' ')
				ctx.FormatNode(node.OnConflict.Where)
			} else {
				__antithesis_instrumentation__.Notify(610095)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(610096)
	}
	__antithesis_instrumentation__.Notify(610074)
	if HasReturningClause(node.Returning) {
		__antithesis_instrumentation__.Notify(610097)
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Returning)
	} else {
		__antithesis_instrumentation__.Notify(610098)
	}
}

func (node *Insert) DefaultValues() bool {
	__antithesis_instrumentation__.Notify(610099)
	return node.Rows.Select == nil
}

type OnConflict struct {
	Columns NameList

	Constraint       Name
	ArbiterPredicate Expr
	Exprs            UpdateExprs
	Where            *Where
	DoNothing        bool
}

func (oc *OnConflict) IsUpsertAlias() bool {
	__antithesis_instrumentation__.Notify(610100)
	return oc != nil && func() bool {
		__antithesis_instrumentation__.Notify(610101)
		return oc.Columns == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610102)
		return oc.Constraint == "" == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610103)
		return oc.ArbiterPredicate == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610104)
		return oc.Exprs == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610105)
		return oc.Where == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(610106)
		return !oc.DoNothing == true
	}() == true
}
