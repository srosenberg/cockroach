package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterRole struct {
	Name      RoleSpec
	IfExists  bool
	IsRole    bool
	KVOptions KVOptions
}

func (node *AlterRole) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603035)
	ctx.WriteString("ALTER")
	if node.IsRole {
		__antithesis_instrumentation__.Notify(603038)
		ctx.WriteString(" ROLE ")
	} else {
		__antithesis_instrumentation__.Notify(603039)
		ctx.WriteString(" USER ")
	}
	__antithesis_instrumentation__.Notify(603036)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603040)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603041)
	}
	__antithesis_instrumentation__.Notify(603037)
	ctx.FormatNode(&node.Name)

	if len(node.KVOptions) > 0 {
		__antithesis_instrumentation__.Notify(603042)
		ctx.WriteString(" WITH")
		node.KVOptions.formatAsRoleOptions(ctx)
	} else {
		__antithesis_instrumentation__.Notify(603043)
	}
}

type AlterRoleSet struct {
	RoleName     RoleSpec
	IfExists     bool
	IsRole       bool
	AllRoles     bool
	DatabaseName Name
	SetOrReset   *SetVar
}

func (node *AlterRoleSet) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603044)
	ctx.WriteString("ALTER")
	if node.IsRole {
		__antithesis_instrumentation__.Notify(603049)
		ctx.WriteString(" ROLE ")
	} else {
		__antithesis_instrumentation__.Notify(603050)
		ctx.WriteString(" USER ")
	}
	__antithesis_instrumentation__.Notify(603045)
	if node.IfExists {
		__antithesis_instrumentation__.Notify(603051)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(603052)
	}
	__antithesis_instrumentation__.Notify(603046)
	if node.AllRoles {
		__antithesis_instrumentation__.Notify(603053)
		ctx.WriteString("ALL ")
	} else {
		__antithesis_instrumentation__.Notify(603054)
		ctx.FormatNode(&node.RoleName)
		ctx.WriteString(" ")
	}
	__antithesis_instrumentation__.Notify(603047)
	if node.DatabaseName != "" {
		__antithesis_instrumentation__.Notify(603055)
		ctx.WriteString("IN DATABASE ")
		ctx.FormatNode(&node.DatabaseName)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(603056)
	}
	__antithesis_instrumentation__.Notify(603048)
	ctx.FormatNode(node.SetOrReset)
}
