package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/privilege"

type Revoke struct {
	Privileges     privilege.List
	Targets        TargetList
	Grantees       RoleSpecList
	GrantOptionFor bool
}

func (node *Revoke) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612942)
	ctx.WriteString("REVOKE ")

	node.Privileges.Format(&ctx.Buffer)
	ctx.WriteString(" ON ")
	ctx.FormatNode(&node.Targets)
	ctx.WriteString(" FROM ")
	ctx.FormatNode(&node.Grantees)
}

type RevokeRole struct {
	Roles       NameList
	Members     RoleSpecList
	AdminOption bool
}

func (node *RevokeRole) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612943)
	ctx.WriteString("REVOKE ")
	if node.AdminOption {
		__antithesis_instrumentation__.Notify(612945)
		ctx.WriteString("ADMIN OPTION FOR ")
	} else {
		__antithesis_instrumentation__.Notify(612946)
	}
	__antithesis_instrumentation__.Notify(612944)
	ctx.FormatNode(&node.Roles)
	ctx.WriteString(" FROM ")
	ctx.FormatNode(&node.Members)
}
