package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ControlJobs struct {
	Jobs    *Select
	Command JobCommand
	Reason  Expr
}

type JobCommand int

const (
	PauseJob JobCommand = iota
	CancelJob
	ResumeJob
)

var JobCommandToStatement = map[JobCommand]string{
	PauseJob:  "PAUSE",
	CancelJob: "CANCEL",
	ResumeJob: "RESUME",
}

func (n *ControlJobs) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612987)
	ctx.WriteString(JobCommandToStatement[n.Command])
	ctx.WriteString(" JOBS ")
	ctx.FormatNode(n.Jobs)
	if n.Reason != nil {
		__antithesis_instrumentation__.Notify(612988)
		ctx.WriteString(" WITH REASON = ")
		ctx.FormatNode(n.Reason)
	} else {
		__antithesis_instrumentation__.Notify(612989)
	}
}

type CancelQueries struct {
	Queries  *Select
	IfExists bool
}

func (node *CancelQueries) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612990)
	ctx.WriteString("CANCEL QUERIES ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(612992)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(612993)
	}
	__antithesis_instrumentation__.Notify(612991)
	ctx.FormatNode(node.Queries)
}

type CancelSessions struct {
	Sessions *Select
	IfExists bool
}

func (node *CancelSessions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(612994)
	ctx.WriteString("CANCEL SESSIONS ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(612996)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(612997)
	}
	__antithesis_instrumentation__.Notify(612995)
	ctx.FormatNode(node.Sessions)
}

type ScheduleCommand int

const (
	PauseSchedule ScheduleCommand = iota
	ResumeSchedule
	DropSchedule
)

func (c ScheduleCommand) String() string {
	__antithesis_instrumentation__.Notify(612998)
	switch c {
	case PauseSchedule:
		__antithesis_instrumentation__.Notify(612999)
		return "PAUSE"
	case ResumeSchedule:
		__antithesis_instrumentation__.Notify(613000)
		return "RESUME"
	case DropSchedule:
		__antithesis_instrumentation__.Notify(613001)
		return "DROP"
	default:
		__antithesis_instrumentation__.Notify(613002)
		panic("unhandled schedule command")
	}
}

type ControlSchedules struct {
	Schedules *Select
	Command   ScheduleCommand
}

var _ Statement = &ControlSchedules{}

func (n *ControlSchedules) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613003)
	ctx.WriteString(n.Command.String())
	ctx.WriteString(" SCHEDULES ")
	ctx.FormatNode(n.Schedules)
}

type ControlJobsForSchedules struct {
	Schedules *Select
	Command   JobCommand
}

type ControlJobsOfType struct {
	Type    string
	Command JobCommand
}

func (n *ControlJobsOfType) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613004)
	ctx.WriteString(JobCommandToStatement[n.Command])
	ctx.WriteString(" ALL ")
	ctx.WriteString(n.Type)
	ctx.WriteString(" JOBS")
}

func (n *ControlJobsForSchedules) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(613005)
	ctx.WriteString(JobCommandToStatement[n.Command])
	ctx.WriteString(" JOBS FOR SCHEDULES ")
	ctx.FormatNode(n.Schedules)
}

var _ Statement = &ControlJobsForSchedules{}
var _ Statement = &ControlJobsOfType{}
