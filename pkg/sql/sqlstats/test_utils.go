package sqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "time"

type TestingKnobs struct {
	OnStmtStatsFlushFinished func()

	OnTxnStatsFlushFinished func()

	OnCleanupStartForShard func(shardIdx int, existingCountInShard, shardLimit int64)

	StubTimeNow func() time.Time

	AOSTClause string
}

func (*TestingKnobs) ModuleTestingKnobs() { __antithesis_instrumentation__.Notify(625773) }

func (knobs *TestingKnobs) GetAOSTClause() string {
	__antithesis_instrumentation__.Notify(625774)
	if knobs != nil {
		__antithesis_instrumentation__.Notify(625776)
		return knobs.AOSTClause
	} else {
		__antithesis_instrumentation__.Notify(625777)
	}
	__antithesis_instrumentation__.Notify(625775)

	return "AS OF SYSTEM TIME follower_read_timestamp()"
}
