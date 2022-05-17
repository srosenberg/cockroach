package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type SetVar struct {
	Name     string
	Local    bool
	Values   Exprs
	Reset    bool
	ResetAll bool
}

func (node *SetVar) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613440)
	if node.ResetAll {
		__antithesis_instrumentation__.Notify(613444)
		ctx.WriteString("RESET ALL")
		return
	} else {
		__antithesis_instrumentation__.Notify(613445)
	}
	__antithesis_instrumentation__.Notify(613441)
	if node.Reset {
		__antithesis_instrumentation__.Notify(613446)
		ctx.WriteString("RESET ")
		ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
			__antithesis_instrumentation__.Notify(613448)

			ctx.FormatNameP(&node.Name)
		})
		__antithesis_instrumentation__.Notify(613447)
		return
	} else {
		__antithesis_instrumentation__.Notify(613449)
	}
	__antithesis_instrumentation__.Notify(613442)
	ctx.WriteString("SET ")
	if node.Local {
		__antithesis_instrumentation__.Notify(613450)
		ctx.WriteString("LOCAL ")
	} else {
		__antithesis_instrumentation__.Notify(613451)
	}
	__antithesis_instrumentation__.Notify(613443)
	if node.Name == "" {
		__antithesis_instrumentation__.Notify(613452)
		ctx.WriteString("ROW (")
		ctx.FormatNode(&node.Values)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(613453)
		ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
			__antithesis_instrumentation__.Notify(613455)

			ctx.FormatNameP(&node.Name)
		})
		__antithesis_instrumentation__.Notify(613454)

		ctx.WriteString(" = ")
		ctx.FormatNode(&node.Values)
	}
}

type SetClusterSetting struct {
	Name  string
	Value Expr
}

func (node *SetClusterSetting) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613456)
	ctx.WriteString("SET CLUSTER SETTING ")

	ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
		__antithesis_instrumentation__.Notify(613458)
		ctx.FormatNameP(&node.Name)
	})
	__antithesis_instrumentation__.Notify(613457)

	ctx.WriteString(" = ")

	switch v := node.Value.(type) {
	case *DBool, *DInt:
		__antithesis_instrumentation__.Notify(613459)
		ctx.WithFlags(ctx.flags & ^FmtAnonymize & ^FmtMarkRedactionNode, func() {
			__antithesis_instrumentation__.Notify(613461)
			ctx.FormatNode(v)
		})
	default:
		__antithesis_instrumentation__.Notify(613460)
		ctx.FormatNode(v)
	}
}

type SetTransaction struct {
	Modes TransactionModes
}

func (node *SetTransaction) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613462)
	ctx.WriteString("SET TRANSACTION")
	ctx.FormatNode(&node.Modes)
}

type SetSessionAuthorizationDefault struct{}

func (node *SetSessionAuthorizationDefault) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613463)
	ctx.WriteString("SET SESSION AUTHORIZATION DEFAULT")
}

type SetSessionCharacteristics struct {
	Modes TransactionModes
}

func (node *SetSessionCharacteristics) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613464)
	ctx.WriteString("SET SESSION CHARACTERISTICS AS TRANSACTION")
	ctx.FormatNode(&node.Modes)
}

type SetTracing struct {
	Values Exprs
}

func (node *SetTracing) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613465)
	ctx.WriteString("SET TRACING = ")

	ctx.WithFlags(ctx.flags&^FmtMarkRedactionNode, func() {
		__antithesis_instrumentation__.Notify(613466)
		ctx.FormatNode(&node.Values)
	})
}
