package sessionphase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "time"

type SessionPhase int

const (
	SessionInit SessionPhase = iota

	SessionQueryReceived

	SessionStartParse

	SessionEndParse

	PlannerStartLogicalPlan

	PlannerEndLogicalPlan

	PlannerStartExecStmt

	PlannerEndExecStmt

	SessionQueryServiced

	SessionTransactionReceived

	SessionFirstStartExecTransaction

	SessionMostRecentStartExecTransaction

	SessionEndExecTransaction

	SessionStartTransactionCommit

	SessionEndTransactionCommit

	SessionStartPostCommitJob

	SessionEndPostCommitJob

	SessionNumPhases
)

type Times struct {
	times [SessionNumPhases]time.Time
}

func NewTimes() *Times {
	__antithesis_instrumentation__.Notify(621691)
	return &Times{}
}

func (t *Times) SetSessionPhaseTime(sp SessionPhase, time time.Time) {
	__antithesis_instrumentation__.Notify(621692)
	t.times[sp] = time
}

func (t *Times) GetSessionPhaseTime(sp SessionPhase) time.Time {
	__antithesis_instrumentation__.Notify(621693)
	return t.times[sp]
}

func (t *Times) Clone() *Times {
	__antithesis_instrumentation__.Notify(621694)
	tCopy := &Times{}
	*tCopy = *t
	return tCopy
}

func (t *Times) GetServiceLatencyNoOverhead() time.Duration {
	__antithesis_instrumentation__.Notify(621695)

	var queryReceivedTime time.Time
	if t.times[SessionEndParse].IsZero() {
		__antithesis_instrumentation__.Notify(621698)
		queryReceivedTime = t.times[SessionStartParse]
	} else {
		__antithesis_instrumentation__.Notify(621699)
		queryReceivedTime = t.times[SessionQueryReceived]
	}
	__antithesis_instrumentation__.Notify(621696)
	parseLatency := t.times[SessionEndParse].Sub(queryReceivedTime)

	var queryEndExecTime time.Time
	if t.times[PlannerEndExecStmt].IsZero() {
		__antithesis_instrumentation__.Notify(621700)
		queryEndExecTime = t.times[PlannerEndLogicalPlan]
	} else {
		__antithesis_instrumentation__.Notify(621701)
		queryEndExecTime = t.times[PlannerEndExecStmt]
	}
	__antithesis_instrumentation__.Notify(621697)
	planAndExecuteLatency := queryEndExecTime.Sub(t.times[PlannerStartLogicalPlan])
	return parseLatency + planAndExecuteLatency
}

func (t *Times) GetServiceLatencyTotal() time.Duration {
	__antithesis_instrumentation__.Notify(621702)
	return t.times[SessionQueryServiced].Sub(t.times[SessionQueryReceived])
}

func (t *Times) GetRunLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621703)
	return t.times[PlannerEndExecStmt].Sub(t.times[PlannerStartExecStmt])
}

func (t *Times) GetPlanningLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621704)
	return t.times[PlannerEndLogicalPlan].Sub(t.times[PlannerStartLogicalPlan])
}

func (t *Times) GetParsingLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621705)
	return t.times[SessionEndParse].Sub(t.times[SessionStartParse])
}

func (t *Times) GetPostCommitJobsLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621706)
	return t.times[SessionEndPostCommitJob].Sub(t.times[SessionStartPostCommitJob])
}

func (t *Times) GetTransactionRetryLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621707)
	return t.times[SessionMostRecentStartExecTransaction].Sub(t.times[SessionFirstStartExecTransaction])
}

func (t *Times) GetTransactionServiceLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621708)
	return t.times[SessionEndExecTransaction].Sub(t.times[SessionTransactionReceived])
}

func (t *Times) GetCommitLatency() time.Duration {
	__antithesis_instrumentation__.Notify(621709)
	return t.times[SessionEndTransactionCommit].Sub(t.times[SessionStartTransactionCommit])
}

func (t *Times) GetSessionAge() time.Duration {
	__antithesis_instrumentation__.Notify(621710)
	return t.times[PlannerEndExecStmt].Sub(t.times[SessionInit])
}
