package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/privilege"

type Grant struct {
	Privileges      privilege.List
	Targets         TargetList
	Grantees        RoleSpecList
	WithGrantOption bool
}

type TargetList struct {
	Databases NameList
	Schemas   ObjectNamePrefixList
	Tables    TablePatterns
	TenantID  TenantID
	Types     []*UnresolvedObjectName

	AllTablesInSchema bool

	SystemUser bool

	ForRoles bool
	Roles    RoleSpecList
}

func (tl *TargetList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609916)
	if tl.Databases != nil {
		__antithesis_instrumentation__.Notify(609917)
		ctx.WriteString("DATABASE ")
		ctx.FormatNode(&tl.Databases)
	} else {
		__antithesis_instrumentation__.Notify(609918)
		if tl.AllTablesInSchema {
			__antithesis_instrumentation__.Notify(609919)
			ctx.WriteString("ALL TABLES IN SCHEMA ")
			ctx.FormatNode(&tl.Schemas)
		} else {
			__antithesis_instrumentation__.Notify(609920)
			if tl.Schemas != nil {
				__antithesis_instrumentation__.Notify(609921)
				ctx.WriteString("SCHEMA ")
				ctx.FormatNode(&tl.Schemas)
			} else {
				__antithesis_instrumentation__.Notify(609922)
				if tl.TenantID.Specified {
					__antithesis_instrumentation__.Notify(609923)
					ctx.WriteString("TENANT ")
					ctx.FormatNode(&tl.TenantID)
				} else {
					__antithesis_instrumentation__.Notify(609924)
					if tl.Types != nil {
						__antithesis_instrumentation__.Notify(609925)
						ctx.WriteString("TYPE ")
						for i, typ := range tl.Types {
							__antithesis_instrumentation__.Notify(609926)
							if i != 0 {
								__antithesis_instrumentation__.Notify(609928)
								ctx.WriteString(", ")
							} else {
								__antithesis_instrumentation__.Notify(609929)
							}
							__antithesis_instrumentation__.Notify(609927)
							ctx.FormatNode(typ)
						}
					} else {
						__antithesis_instrumentation__.Notify(609930)
						ctx.WriteString("TABLE ")
						ctx.FormatNode(&tl.Tables)
					}
				}
			}
		}
	}
}

func (node *Grant) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609931)
	ctx.WriteString("GRANT ")
	node.Privileges.Format(&ctx.Buffer)
	ctx.WriteString(" ON ")
	ctx.FormatNode(&node.Targets)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.Grantees)
}

type GrantRole struct {
	Roles       NameList
	Members     RoleSpecList
	AdminOption bool
}

func (node *GrantRole) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(609932)
	ctx.WriteString("GRANT ")
	ctx.FormatNode(&node.Roles)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.Members)
	if node.AdminOption {
		__antithesis_instrumentation__.Notify(609933)
		ctx.WriteString(" WITH ADMIN OPTION")
	} else {
		__antithesis_instrumentation__.Notify(609934)
	}
}
