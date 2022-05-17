package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type FullBackupClause struct {
	AlwaysFull bool
	Recurrence Expr
}

type ScheduleLabelSpec struct {
	IfNotExists bool
	Label       Expr
}

type ScheduledBackup struct {
	ScheduleLabelSpec ScheduleLabelSpec
	Recurrence        Expr
	FullBackup        *FullBackupClause
	Targets           *TargetList
	To                StringOrPlaceholderOptList
	BackupOptions     BackupOptions
	ScheduleOptions   KVOptions
}

var _ Statement = &ScheduledBackup{}

func (node *ScheduledBackup) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613006)
	ctx.WriteString("CREATE SCHEDULE")

	if node.ScheduleLabelSpec.IfNotExists {
		__antithesis_instrumentation__.Notify(613013)
		ctx.WriteString(" IF NOT EXISTS")
	} else {
		__antithesis_instrumentation__.Notify(613014)
	}
	__antithesis_instrumentation__.Notify(613007)
	if node.ScheduleLabelSpec.Label != nil {
		__antithesis_instrumentation__.Notify(613015)
		ctx.WriteString(" ")
		ctx.FormatNode(node.ScheduleLabelSpec.Label)
	} else {
		__antithesis_instrumentation__.Notify(613016)
	}
	__antithesis_instrumentation__.Notify(613008)

	ctx.WriteString(" FOR BACKUP")
	if node.Targets != nil {
		__antithesis_instrumentation__.Notify(613017)
		ctx.WriteString(" ")
		ctx.FormatNode(node.Targets)
	} else {
		__antithesis_instrumentation__.Notify(613018)
	}
	__antithesis_instrumentation__.Notify(613009)

	ctx.WriteString(" INTO ")
	ctx.FormatNode(&node.To)

	if !node.BackupOptions.IsDefault() {
		__antithesis_instrumentation__.Notify(613019)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.BackupOptions)
	} else {
		__antithesis_instrumentation__.Notify(613020)
	}
	__antithesis_instrumentation__.Notify(613010)

	ctx.WriteString(" RECURRING ")
	if node.Recurrence == nil {
		__antithesis_instrumentation__.Notify(613021)
		ctx.WriteString("NEVER")
	} else {
		__antithesis_instrumentation__.Notify(613022)
		ctx.FormatNode(node.Recurrence)
	}
	__antithesis_instrumentation__.Notify(613011)

	if node.FullBackup != nil {
		__antithesis_instrumentation__.Notify(613023)

		if node.FullBackup.Recurrence != nil {
			__antithesis_instrumentation__.Notify(613024)
			ctx.WriteString(" FULL BACKUP ")
			ctx.FormatNode(node.FullBackup.Recurrence)
		} else {
			__antithesis_instrumentation__.Notify(613025)
			if node.FullBackup.AlwaysFull {
				__antithesis_instrumentation__.Notify(613026)
				ctx.WriteString(" FULL BACKUP ALWAYS")
			} else {
				__antithesis_instrumentation__.Notify(613027)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(613028)
	}
	__antithesis_instrumentation__.Notify(613012)

	if node.ScheduleOptions != nil {
		__antithesis_instrumentation__.Notify(613029)
		ctx.WriteString(" WITH SCHEDULE OPTIONS ")
		ctx.FormatNode(&node.ScheduleOptions)
	} else {
		__antithesis_instrumentation__.Notify(613030)
	}
}

func (node ScheduledBackup) Coverage() DescriptorCoverage {
	__antithesis_instrumentation__.Notify(613031)
	if node.Targets == nil {
		__antithesis_instrumentation__.Notify(613033)
		return AllDescriptors
	} else {
		__antithesis_instrumentation__.Notify(613034)
	}
	__antithesis_instrumentation__.Notify(613032)
	return RequestedDescriptors
}
