package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterBackup struct {
	Backup Expr
	Subdir Expr
	Cmds   AlterBackupCmds
}

var _ Statement = &AlterBackup{}

func (node *AlterBackup) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602921)
	ctx.WriteString("ALTER BACKUP ")

	if node.Subdir != nil {
		__antithesis_instrumentation__.Notify(602923)
		ctx.FormatNode(node.Subdir)
		ctx.WriteString(" IN ")
	} else {
		__antithesis_instrumentation__.Notify(602924)
	}
	__antithesis_instrumentation__.Notify(602922)

	ctx.FormatNode(node.Backup)
	ctx.FormatNode(&node.Cmds)
}

type AlterBackupCmds []AlterBackupCmd

func (node *AlterBackupCmds) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602925)
	for i, n := range *node {
		__antithesis_instrumentation__.Notify(602926)
		if i > 0 {
			__antithesis_instrumentation__.Notify(602928)
			ctx.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(602929)
		}
		__antithesis_instrumentation__.Notify(602927)
		ctx.FormatNode(n)
	}
}

type AlterBackupCmd interface {
	NodeFormatter
	alterBackupCmd()
}

func (node *AlterBackupKMS) alterBackupCmd() { __antithesis_instrumentation__.Notify(602930) }

var _ AlterBackupCmd = &AlterBackupKMS{}

type AlterBackupKMS struct {
	KMSInfo BackupKMS
}

func (node *AlterBackupKMS) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602931)
	ctx.WriteString(" ADD NEW_KMS=")
	ctx.FormatNode(&node.KMSInfo.NewKMSURI)

	ctx.WriteString(" WITH OLD_KMS=")
	ctx.FormatNode(&node.KMSInfo.OldKMSURI)
}

type BackupKMS struct {
	NewKMSURI StringOrPlaceholderOptList
	OldKMSURI StringOrPlaceholderOptList
}
