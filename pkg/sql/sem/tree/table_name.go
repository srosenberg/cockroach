package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type TableName struct {
	objName
}

func (t *TableName) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614451)
	if ctx.tableNameFormatter != nil {
		__antithesis_instrumentation__.Notify(614454)
		ctx.tableNameFormatter(ctx, t)
		return
	} else {
		__antithesis_instrumentation__.Notify(614455)
	}
	__antithesis_instrumentation__.Notify(614452)
	t.ObjectNamePrefix.Format(ctx)
	if t.ExplicitSchema || func() bool {
		__antithesis_instrumentation__.Notify(614456)
		return ctx.alwaysFormatTablePrefix() == true
	}() == true {
		__antithesis_instrumentation__.Notify(614457)
		ctx.WriteByte('.')
	} else {
		__antithesis_instrumentation__.Notify(614458)
	}
	__antithesis_instrumentation__.Notify(614453)
	ctx.FormatNode(&t.ObjectName)
}
func (t *TableName) String() string {
	__antithesis_instrumentation__.Notify(614459)
	return AsString(t)
}

func (t *TableName) objectName() { __antithesis_instrumentation__.Notify(614460) }

func (t *TableName) FQString() string {
	__antithesis_instrumentation__.Notify(614461)
	ctx := NewFmtCtx(FmtSimple)
	ctx.FormatNode(&t.CatalogName)
	ctx.WriteByte('.')
	ctx.FormatNode(&t.SchemaName)
	ctx.WriteByte('.')
	ctx.FormatNode(&t.ObjectName)
	return ctx.CloseAndGetString()
}

func (t *TableName) Table() string {
	__antithesis_instrumentation__.Notify(614462)
	return string(t.ObjectName)
}

func (t *TableName) Equals(other *TableName) bool {
	__antithesis_instrumentation__.Notify(614463)
	return *t == *other
}

func (*TableName) tableExpr() { __antithesis_instrumentation__.Notify(614464) }

func NewTableNameWithSchema(db, sc, tbl Name) *TableName {
	__antithesis_instrumentation__.Notify(614465)
	tn := MakeTableNameWithSchema(db, sc, tbl)
	return &tn
}

func MakeTableNameWithSchema(db, schema, tbl Name) TableName {
	__antithesis_instrumentation__.Notify(614466)
	return TableName{objName{
		ObjectName: tbl,
		ObjectNamePrefix: ObjectNamePrefix{
			CatalogName:     db,
			SchemaName:      schema,
			ExplicitSchema:  true,
			ExplicitCatalog: true,
		},
	}}
}

func MakeTableNameFromPrefix(prefix ObjectNamePrefix, object Name) TableName {
	__antithesis_instrumentation__.Notify(614467)
	return TableName{objName{
		ObjectName:       object,
		ObjectNamePrefix: prefix,
	}}
}

func MakeUnqualifiedTableName(tbl Name) TableName {
	__antithesis_instrumentation__.Notify(614468)
	return TableName{objName{
		ObjectName: tbl,
	}}
}

func NewUnqualifiedTableName(tbl Name) *TableName {
	__antithesis_instrumentation__.Notify(614469)
	tn := MakeUnqualifiedTableName(tbl)
	return &tn
}

func makeTableNameFromUnresolvedName(n *UnresolvedName) TableName {
	__antithesis_instrumentation__.Notify(614470)
	return TableName{objName{
		ObjectName:       Name(n.Parts[0]),
		ObjectNamePrefix: makeObjectNamePrefixFromUnresolvedName(n),
	}}
}

func makeObjectNamePrefixFromUnresolvedName(n *UnresolvedName) ObjectNamePrefix {
	__antithesis_instrumentation__.Notify(614471)
	return ObjectNamePrefix{
		SchemaName:      Name(n.Parts[1]),
		CatalogName:     Name(n.Parts[2]),
		ExplicitSchema:  n.NumParts >= 2,
		ExplicitCatalog: n.NumParts >= 3,
	}
}

type TableNames []TableName

func (ts *TableNames) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614472)
	sep := ""
	for i := range *ts {
		__antithesis_instrumentation__.Notify(614473)
		ctx.WriteString(sep)
		ctx.FormatNode(&(*ts)[i])
		sep = ", "
	}
}
func (ts *TableNames) String() string {
	__antithesis_instrumentation__.Notify(614474)
	return AsString(ts)
}

type TableIndexName struct {
	Table TableName
	Index UnrestrictedName
}

func (n *TableIndexName) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614475)
	if n.Index == "" {
		__antithesis_instrumentation__.Notify(614479)
		ctx.FormatNode(&n.Table)
		return
	} else {
		__antithesis_instrumentation__.Notify(614480)
	}
	__antithesis_instrumentation__.Notify(614476)

	if n.Table.ObjectName != "" {
		__antithesis_instrumentation__.Notify(614481)

		ctx.FormatNode(&n.Table)
		ctx.WriteByte('@')
		ctx.FormatNode(&n.Index)
		return
	} else {
		__antithesis_instrumentation__.Notify(614482)
	}
	__antithesis_instrumentation__.Notify(614477)

	if n.Table.ExplicitSchema || func() bool {
		__antithesis_instrumentation__.Notify(614483)
		return ctx.alwaysFormatTablePrefix() == true
	}() == true {
		__antithesis_instrumentation__.Notify(614484)
		ctx.FormatNode(&n.Table.ObjectNamePrefix)
		ctx.WriteByte('.')
	} else {
		__antithesis_instrumentation__.Notify(614485)
	}
	__antithesis_instrumentation__.Notify(614478)

	ctx.FormatNode((*Name)(&n.Index))
}

func (n *TableIndexName) String() string {
	__antithesis_instrumentation__.Notify(614486)
	return AsString(n)
}

type TableIndexNames []*TableIndexName

func (n *TableIndexNames) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614487)
	sep := ""
	for _, tni := range *n {
		__antithesis_instrumentation__.Notify(614488)
		ctx.WriteString(sep)
		ctx.FormatNode(tni)
		sep = ", "
	}
}
