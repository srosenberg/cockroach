package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const (
	systemLogGCPeriod = 10 * time.Minute
)

var (
	rangeLogTTL = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.rangelog.ttl",
		fmt.Sprintf(
			"if nonzero, range log entries older than this duration are deleted every %s. "+
				"Should not be lowered below 24 hours.",
			systemLogGCPeriod,
		),
		30*24*time.Hour,
	).WithPublic()

	eventLogTTL = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.eventlog.ttl",
		fmt.Sprintf(
			"if nonzero, entries in system.eventlog older than this duration are deleted every %s. "+
				"Should not be lowered below 24 hours.",
			systemLogGCPeriod,
		),
		90*24*time.Hour,
	).WithPublic()
)

func (s *Server) gcSystemLog(
	ctx context.Context, table string, timestampLowerBound, timestampUpperBound time.Time,
) (time.Time, int64, error) {
	__antithesis_instrumentation__.Notify(196118)
	var totalRowsAffected int64
	repl, _, err := s.node.stores.GetReplicaForRangeID(ctx, roachpb.RangeID(1))
	if roachpb.IsRangeNotFoundError(err) {
		__antithesis_instrumentation__.Notify(196122)
		return timestampLowerBound, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(196123)
	}
	__antithesis_instrumentation__.Notify(196119)
	if err != nil {
		__antithesis_instrumentation__.Notify(196124)
		return timestampLowerBound, 0, err
	} else {
		__antithesis_instrumentation__.Notify(196125)
	}
	__antithesis_instrumentation__.Notify(196120)

	if !repl.IsFirstRange() || func() bool {
		__antithesis_instrumentation__.Notify(196126)
		return !repl.OwnsValidLease(ctx, s.clock.NowAsClockTimestamp()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(196127)
		return timestampLowerBound, 0, nil
	} else {
		__antithesis_instrumentation__.Notify(196128)
	}
	__antithesis_instrumentation__.Notify(196121)

	deleteStmt := fmt.Sprintf(
		`SELECT count(1), max(timestamp) FROM
[DELETE FROM system.%s WHERE timestamp >= $1 AND timestamp <= $2 LIMIT 1000 RETURNING timestamp]`,
		table,
	)

	for {
		__antithesis_instrumentation__.Notify(196129)
		var rowsAffected int64
		err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(196132)
			var err error
			row, err := s.sqlServer.internalExecutor.QueryRowEx(
				ctx,
				table+"-gc",
				txn,
				sessiondata.InternalExecutorOverride{User: security.RootUserName()},
				deleteStmt,
				timestampLowerBound,
				timestampUpperBound,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(196135)
				return err
			} else {
				__antithesis_instrumentation__.Notify(196136)
			}
			__antithesis_instrumentation__.Notify(196133)

			if row != nil {
				__antithesis_instrumentation__.Notify(196137)
				rowCount, ok := row[0].(*tree.DInt)
				if !ok {
					__antithesis_instrumentation__.Notify(196140)
					return errors.Errorf("row count is of unknown type %T", row[0])
				} else {
					__antithesis_instrumentation__.Notify(196141)
				}
				__antithesis_instrumentation__.Notify(196138)
				if rowCount == nil {
					__antithesis_instrumentation__.Notify(196142)
					return errors.New("error parsing row count")
				} else {
					__antithesis_instrumentation__.Notify(196143)
				}
				__antithesis_instrumentation__.Notify(196139)
				rowsAffected = int64(*rowCount)

				if rowsAffected > 0 {
					__antithesis_instrumentation__.Notify(196144)
					maxTimestamp, ok := row[1].(*tree.DTimestamp)
					if !ok {
						__antithesis_instrumentation__.Notify(196147)
						return errors.Errorf("timestamp is of unknown type %T", row[1])
					} else {
						__antithesis_instrumentation__.Notify(196148)
					}
					__antithesis_instrumentation__.Notify(196145)
					if maxTimestamp == nil {
						__antithesis_instrumentation__.Notify(196149)
						return errors.New("error parsing timestamp")
					} else {
						__antithesis_instrumentation__.Notify(196150)
					}
					__antithesis_instrumentation__.Notify(196146)
					timestampLowerBound = maxTimestamp.Time
				} else {
					__antithesis_instrumentation__.Notify(196151)
				}
			} else {
				__antithesis_instrumentation__.Notify(196152)
			}
			__antithesis_instrumentation__.Notify(196134)
			return nil
		})
		__antithesis_instrumentation__.Notify(196130)
		totalRowsAffected += rowsAffected
		if err != nil {
			__antithesis_instrumentation__.Notify(196153)
			return timestampLowerBound, totalRowsAffected, err
		} else {
			__antithesis_instrumentation__.Notify(196154)
		}
		__antithesis_instrumentation__.Notify(196131)

		if rowsAffected == 0 {
			__antithesis_instrumentation__.Notify(196155)
			return timestampUpperBound, totalRowsAffected, nil
		} else {
			__antithesis_instrumentation__.Notify(196156)
		}
	}
}

type systemLogGCConfig struct {
	ttl *settings.DurationSetting

	timestampLowerBound time.Time
}

func (s *Server) startSystemLogsGC(ctx context.Context) {
	__antithesis_instrumentation__.Notify(196157)
	systemLogsToGC := map[string]*systemLogGCConfig{
		"rangelog": {
			ttl:                 rangeLogTTL,
			timestampLowerBound: timeutil.Unix(0, 0),
		},
		"eventlog": {
			ttl:                 eventLogTTL,
			timestampLowerBound: timeutil.Unix(0, 0),
		},
	}

	_ = s.stopper.RunAsyncTask(ctx, "system-log-gc", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(196158)
		period := systemLogGCPeriod
		if storeKnobs, ok := s.cfg.TestingKnobs.Store.(*kvserver.StoreTestingKnobs); ok && func() bool {
			__antithesis_instrumentation__.Notify(196160)
			return storeKnobs.SystemLogsGCPeriod != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(196161)
			period = storeKnobs.SystemLogsGCPeriod
		} else {
			__antithesis_instrumentation__.Notify(196162)
		}
		__antithesis_instrumentation__.Notify(196159)

		t := time.NewTicker(period)
		defer t.Stop()

		for {
			__antithesis_instrumentation__.Notify(196163)
			select {
			case <-t.C:
				__antithesis_instrumentation__.Notify(196164)
				for table, gcConfig := range systemLogsToGC {
					__antithesis_instrumentation__.Notify(196167)
					ttl := gcConfig.ttl.Get(&s.cfg.Settings.SV)
					if ttl > 0 {
						__antithesis_instrumentation__.Notify(196168)
						timestampUpperBound := timeutil.Unix(0, s.clock.PhysicalNow()-int64(ttl))
						newTimestampLowerBound, rowsAffected, err := s.gcSystemLog(
							ctx,
							table,
							gcConfig.timestampLowerBound,
							timestampUpperBound,
						)
						if err != nil {
							__antithesis_instrumentation__.Notify(196169)
							log.Warningf(
								ctx,
								"error garbage collecting %s: %v",
								table,
								err,
							)
						} else {
							__antithesis_instrumentation__.Notify(196170)
							gcConfig.timestampLowerBound = newTimestampLowerBound
							if log.V(1) {
								__antithesis_instrumentation__.Notify(196171)
								log.Infof(ctx, "garbage collected %d rows from %s", rowsAffected, table)
							} else {
								__antithesis_instrumentation__.Notify(196172)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(196173)
					}
				}
				__antithesis_instrumentation__.Notify(196165)

				if storeKnobs, ok := s.cfg.TestingKnobs.Store.(*kvserver.StoreTestingKnobs); ok && func() bool {
					__antithesis_instrumentation__.Notify(196174)
					return storeKnobs.SystemLogsGCGCDone != nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(196175)
					select {
					case storeKnobs.SystemLogsGCGCDone <- struct{}{}:
						__antithesis_instrumentation__.Notify(196176)
					case <-s.stopper.ShouldQuiesce():
						__antithesis_instrumentation__.Notify(196177)

						return
					}
				} else {
					__antithesis_instrumentation__.Notify(196178)
				}
			case <-s.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(196166)
				return
			}
		}
	})
}
