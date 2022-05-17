package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sessionphase"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type StatsCollector struct {
	sqlstats.ApplicationStats

	phaseTimes *sessionphase.Times

	previousPhaseTimes *sessionphase.Times

	flushTarget sqlstats.ApplicationStats
	st          *cluster.Settings
	knobs       *sqlstats.TestingKnobs
}

var _ sqlstats.ApplicationStats = &StatsCollector{}

func NewStatsCollector(
	st *cluster.Settings,
	appStats sqlstats.ApplicationStats,
	phaseTime *sessionphase.Times,
	knobs *sqlstats.TestingKnobs,
) *StatsCollector {
	__antithesis_instrumentation__.Notify(625446)
	return &StatsCollector{
		ApplicationStats: appStats,
		phaseTimes:       phaseTime.Clone(),
		st:               st,
		knobs:            knobs,
	}
}

func (s *StatsCollector) PhaseTimes() *sessionphase.Times {
	__antithesis_instrumentation__.Notify(625447)
	return s.phaseTimes
}

func (s *StatsCollector) PreviousPhaseTimes() *sessionphase.Times {
	__antithesis_instrumentation__.Notify(625448)
	return s.previousPhaseTimes
}

func (s *StatsCollector) Reset(appStats sqlstats.ApplicationStats, phaseTime *sessionphase.Times) {
	__antithesis_instrumentation__.Notify(625449)
	previousPhaseTime := s.phaseTimes
	if s.isInExplicitTransaction() {
		__antithesis_instrumentation__.Notify(625451)
		s.flushTarget = appStats
	} else {
		__antithesis_instrumentation__.Notify(625452)
		s.ApplicationStats = appStats
	}
	__antithesis_instrumentation__.Notify(625450)

	s.previousPhaseTimes = previousPhaseTime
	s.phaseTimes = phaseTime.Clone()
}

func (s *StatsCollector) StartExplicitTransaction() {
	__antithesis_instrumentation__.Notify(625453)
	s.flushTarget = s.ApplicationStats
	s.ApplicationStats = s.flushTarget.NewApplicationStatsWithInheritedOptions()
}

func (s *StatsCollector) EndExplicitTransaction(
	ctx context.Context, transactionFingerprintID roachpb.TransactionFingerprintID,
) {
	__antithesis_instrumentation__.Notify(625454)

	if !AssociateStmtWithTxnFingerprint.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(625458)
		transactionFingerprintID = roachpb.InvalidTransactionFingerprintID
	} else {
		__antithesis_instrumentation__.Notify(625459)
	}
	__antithesis_instrumentation__.Notify(625455)

	var discardedStats uint64
	discardedStats += s.flushTarget.MergeApplicationStatementStats(
		ctx,
		s.ApplicationStats,
		func(statistics *roachpb.CollectedStatementStatistics) {
			__antithesis_instrumentation__.Notify(625460)
			statistics.Key.TransactionFingerprintID = transactionFingerprintID
		},
	)
	__antithesis_instrumentation__.Notify(625456)

	discardedStats += s.flushTarget.MergeApplicationTransactionStats(
		ctx,
		s.ApplicationStats,
	)

	if discardedStats > 0 {
		__antithesis_instrumentation__.Notify(625461)
		log.Warningf(ctx, "%d statement statistics discarded due to memory limit", discardedStats)
	} else {
		__antithesis_instrumentation__.Notify(625462)
	}
	__antithesis_instrumentation__.Notify(625457)

	s.ApplicationStats.Free(ctx)
	s.ApplicationStats = s.flushTarget
	s.flushTarget = nil
}

func (s *StatsCollector) ShouldSaveLogicalPlanDesc(
	fingerprint string, implicitTxn bool, database string,
) bool {
	__antithesis_instrumentation__.Notify(625463)
	if s.isInExplicitTransaction() {
		__antithesis_instrumentation__.Notify(625465)
		return s.flushTarget.ShouldSaveLogicalPlanDesc(fingerprint, implicitTxn, database) && func() bool {
			__antithesis_instrumentation__.Notify(625466)
			return s.ApplicationStats.ShouldSaveLogicalPlanDesc(fingerprint, implicitTxn, database) == true
		}() == true
	} else {
		__antithesis_instrumentation__.Notify(625467)
	}
	__antithesis_instrumentation__.Notify(625464)

	return s.ApplicationStats.ShouldSaveLogicalPlanDesc(fingerprint, implicitTxn, database)
}

func (s *StatsCollector) isInExplicitTransaction() bool {
	__antithesis_instrumentation__.Notify(625468)
	return s.flushTarget != nil
}
