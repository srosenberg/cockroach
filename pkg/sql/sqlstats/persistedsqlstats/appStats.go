package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/ssmemstorage"
	"github.com/cockroachdb/errors"
)

type ApplicationStats struct {
	sqlstats.ApplicationStats

	memoryPressureSignal chan struct{}
}

var _ sqlstats.ApplicationStats = &ApplicationStats{}

func (s *ApplicationStats) RecordStatement(
	ctx context.Context, key roachpb.StatementStatisticsKey, value sqlstats.RecordedStmtStats,
) (roachpb.StmtFingerprintID, error) {
	__antithesis_instrumentation__.Notify(624418)
	var fingerprintID roachpb.StmtFingerprintID
	err := s.recordStatsOrSendMemoryPressureSignal(func() (err error) {
		__antithesis_instrumentation__.Notify(624420)
		fingerprintID, err = s.ApplicationStats.RecordStatement(ctx, key, value)
		return err
	})
	__antithesis_instrumentation__.Notify(624419)
	return fingerprintID, err
}

func (s *ApplicationStats) ShouldSaveLogicalPlanDesc(
	fingerprint string, implicitTxn bool, database string,
) bool {
	__antithesis_instrumentation__.Notify(624421)
	return s.ApplicationStats.ShouldSaveLogicalPlanDesc(fingerprint, implicitTxn, database)
}

func (s *ApplicationStats) RecordTransaction(
	ctx context.Context, key roachpb.TransactionFingerprintID, value sqlstats.RecordedTxnStats,
) error {
	__antithesis_instrumentation__.Notify(624422)
	return s.recordStatsOrSendMemoryPressureSignal(func() error {
		__antithesis_instrumentation__.Notify(624423)
		return s.ApplicationStats.RecordTransaction(ctx, key, value)
	})
}

func (s *ApplicationStats) recordStatsOrSendMemoryPressureSignal(fn func() error) error {
	__antithesis_instrumentation__.Notify(624424)
	err := fn()
	if errors.Is(err, ssmemstorage.ErrFingerprintLimitReached) || func() bool {
		__antithesis_instrumentation__.Notify(624426)
		return errors.Is(err, ssmemstorage.ErrMemoryPressure) == true
	}() == true {
		__antithesis_instrumentation__.Notify(624427)
		select {
		case s.memoryPressureSignal <- struct{}{}:
			__antithesis_instrumentation__.Notify(624429)

		default:
			__antithesis_instrumentation__.Notify(624430)
		}
		__antithesis_instrumentation__.Notify(624428)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(624431)
	}
	__antithesis_instrumentation__.Notify(624425)
	return err
}
