package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

type TablePattern interface {
	fmt.Stringer
	NodeFormatter

	NormalizeTablePattern() (TablePattern, error)
}

var _ TablePattern = &UnresolvedName{}
var _ TablePattern = &TableName{}
var _ TablePattern = &AllTablesSelector{}

func (n *UnresolvedName) NormalizeTablePattern() (TablePattern, error) {
	__antithesis_instrumentation__.Notify(614489)
	return classifyTablePattern(n)
}

func (t *TableName) NormalizeTablePattern() (TablePattern, error) {
	__antithesis_instrumentation__.Notify(614490)
	return t, nil
}

type AllTablesSelector struct {
	ObjectNamePrefix
}

func (at *AllTablesSelector) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614491)
	at.ObjectNamePrefix.Format(ctx)
	if at.ExplicitSchema || func() bool {
		__antithesis_instrumentation__.Notify(614493)
		return ctx.alwaysFormatTablePrefix() == true
	}() == true {
		__antithesis_instrumentation__.Notify(614494)
		ctx.WriteByte('.')
	} else {
		__antithesis_instrumentation__.Notify(614495)
	}
	__antithesis_instrumentation__.Notify(614492)
	ctx.WriteByte('*')
}
func (at *AllTablesSelector) String() string {
	__antithesis_instrumentation__.Notify(614496)
	return AsString(at)
}

func (at *AllTablesSelector) NormalizeTablePattern() (TablePattern, error) {
	__antithesis_instrumentation__.Notify(614497)
	return at, nil
}

type TablePatterns []TablePattern

func (tt *TablePatterns) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614498)
	for i, t := range *tt {
		__antithesis_instrumentation__.Notify(614499)
		if i > 0 {
			__antithesis_instrumentation__.Notify(614501)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(614502)
		}
		__antithesis_instrumentation__.Notify(614500)
		ctx.FormatNode(t)
	}
}
