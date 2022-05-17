package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterTenantSetClusterSetting struct {
	SetClusterSetting
	TenantID  Expr
	TenantAll bool
}

func (n *AlterTenantSetClusterSetting) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614521)
	ctx.WriteString("ALTER TENANT ")
	if n.TenantAll {
		__antithesis_instrumentation__.Notify(614523)
		ctx.WriteString("ALL")
	} else {
		__antithesis_instrumentation__.Notify(614524)
		ctx.FormatNode(n.TenantID)
	}
	__antithesis_instrumentation__.Notify(614522)
	ctx.WriteByte(' ')
	ctx.FormatNode(&n.SetClusterSetting)
}

type ShowTenantClusterSetting struct {
	*ShowClusterSetting
	TenantID Expr
}

func (node *ShowTenantClusterSetting) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614525)
	ctx.FormatNode(node.ShowClusterSetting)
	ctx.WriteString(" FOR TENANT ")
	ctx.FormatNode(node.TenantID)
}

type ShowTenantClusterSettingList struct {
	*ShowClusterSettingList
	TenantID Expr
}

func (node *ShowTenantClusterSettingList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614526)
	ctx.FormatNode(node.ShowClusterSettingList)
	ctx.WriteString(" FOR TENANT ")
	ctx.FormatNode(node.TenantID)
}
