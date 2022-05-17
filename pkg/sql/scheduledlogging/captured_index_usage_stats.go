package scheduledlogging

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var telemetryCaptureIndexUsageStatsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.telemetry.capture_index_usage_stats.enabled",
	"enable/disable capturing index usage statistics to the telemetry logging channel",
	true,
)

var telemetryCaptureIndexUsageStatsInterval = settings.RegisterDurationSetting(
	settings.SystemOnly,
	"sql.telemetry.capture_index_usage_stats.interval",
	"the scheduled interval time between capturing index usage statistics when capturing index usage statistics is enabled",
	8*time.Hour,
	settings.NonNegativeDuration,
)

var telemetryCaptureIndexUsageStatsStatusCheckEnabledInterval = settings.RegisterDurationSetting(
	settings.SystemOnly,
	"sql.telemetry.capture_index_usage_stats.check_enabled_interval",
	"the scheduled interval time between checks to see if index usage statistics has been enabled",
	10*time.Minute,
	settings.NonNegativeDuration,
)

var telemetryCaptureIndexUsageStatsLoggingDelay = settings.RegisterDurationSetting(
	settings.SystemOnly,
	"sql.telemetry.capture_index_usage_stats.logging_delay",
	"the time delay between emitting individual index usage stats logs, this is done to "+
		"mitigate the log-line limit of 10 logs per second on the telemetry pipeline",
	500*time.Millisecond,
	settings.NonNegativeDuration,
)

type CaptureIndexUsageStatsTestingKnobs struct {
	getLoggingDuration func() time.Duration

	getOverlapDuration func() time.Duration
}

func (*CaptureIndexUsageStatsTestingKnobs) ModuleTestingKnobs() {
	__antithesis_instrumentation__.Notify(576914)
}

type CaptureIndexUsageStatsLoggingScheduler struct {
	db                      *kv.DB
	st                      *cluster.Settings
	ie                      sqlutil.InternalExecutor
	knobs                   *CaptureIndexUsageStatsTestingKnobs
	currentCaptureStartTime time.Time
}

func (s *CaptureIndexUsageStatsLoggingScheduler) getLoggingDuration() time.Duration {
	__antithesis_instrumentation__.Notify(576915)
	if s.knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(576917)
		return s.knobs.getLoggingDuration != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(576918)
		return s.knobs.getLoggingDuration()
	} else {
		__antithesis_instrumentation__.Notify(576919)
	}
	__antithesis_instrumentation__.Notify(576916)
	return timeutil.Since(s.currentCaptureStartTime)
}

func (s *CaptureIndexUsageStatsLoggingScheduler) durationOnOverlap() time.Duration {
	__antithesis_instrumentation__.Notify(576920)
	if s.knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(576922)
		return s.knobs.getOverlapDuration != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(576923)
		return s.knobs.getOverlapDuration()
	} else {
		__antithesis_instrumentation__.Notify(576924)
	}
	__antithesis_instrumentation__.Notify(576921)

	return 0 * time.Second
}

func (s *CaptureIndexUsageStatsLoggingScheduler) durationUntilNextInterval() time.Duration {
	__antithesis_instrumentation__.Notify(576925)

	if !telemetryCaptureIndexUsageStatsEnabled.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(576928)
		return telemetryCaptureIndexUsageStatsStatusCheckEnabledInterval.Get(&s.st.SV)
	} else {
		__antithesis_instrumentation__.Notify(576929)
	}
	__antithesis_instrumentation__.Notify(576926)

	if s.getLoggingDuration() >= telemetryCaptureIndexUsageStatsInterval.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(576930)
		return s.durationOnOverlap()
	} else {
		__antithesis_instrumentation__.Notify(576931)
	}
	__antithesis_instrumentation__.Notify(576927)

	return telemetryCaptureIndexUsageStatsInterval.Get(&s.st.SV)
}

func Start(
	ctx context.Context,
	stopper *stop.Stopper,
	db *kv.DB,
	cs *cluster.Settings,
	ie sqlutil.InternalExecutor,
	knobs *CaptureIndexUsageStatsTestingKnobs,
) {
	__antithesis_instrumentation__.Notify(576932)
	scheduler := CaptureIndexUsageStatsLoggingScheduler{
		db:    db,
		st:    cs,
		ie:    ie,
		knobs: knobs,
	}
	scheduler.start(ctx, stopper)
}

func (s *CaptureIndexUsageStatsLoggingScheduler) start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(576933)
	_ = stopper.RunAsyncTask(ctx, "capture-index-usage-stats", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(576934)

		for timer := time.NewTimer(0 * time.Second); ; timer.Reset(s.durationUntilNextInterval()) {
			__antithesis_instrumentation__.Notify(576935)
			select {
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(576936)
				timer.Stop()
				return
			case <-timer.C:
				__antithesis_instrumentation__.Notify(576937)
				s.currentCaptureStartTime = timeutil.Now()
				if !telemetryCaptureIndexUsageStatsEnabled.Get(&s.st.SV) {
					__antithesis_instrumentation__.Notify(576939)
					continue
				} else {
					__antithesis_instrumentation__.Notify(576940)
				}
				__antithesis_instrumentation__.Notify(576938)

				err := captureIndexUsageStats(ctx, s.ie, stopper, telemetryCaptureIndexUsageStatsLoggingDelay.Get(&s.st.SV))
				if err != nil {
					__antithesis_instrumentation__.Notify(576941)
					log.Warningf(ctx, "error capturing index usage stats: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(576942)
				}
			}
		}
	})
}

func captureIndexUsageStats(
	ctx context.Context,
	ie sqlutil.InternalExecutor,
	stopper *stop.Stopper,
	loggingDelay time.Duration,
) error {
	__antithesis_instrumentation__.Notify(576943)
	allDatabaseNames, err := getAllDatabaseNames(ctx, ie)
	if err != nil {
		__antithesis_instrumentation__.Notify(576946)
		return err
	} else {
		__antithesis_instrumentation__.Notify(576947)
	}
	__antithesis_instrumentation__.Notify(576944)

	var ok bool
	expectedNumDatums := 9
	var allCapturedIndexUsageStats []eventpb.EventPayload
	for _, databaseName := range allDatabaseNames {
		__antithesis_instrumentation__.Notify(576948)

		if databaseName == "system" || func() bool {
			__antithesis_instrumentation__.Notify(576952)
			return databaseName == "defaultdb" == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(576953)
			return databaseName == "postgres" == true
		}() == true {
			__antithesis_instrumentation__.Notify(576954)
			continue
		} else {
			__antithesis_instrumentation__.Notify(576955)
		}
		__antithesis_instrumentation__.Notify(576949)
		stmt := fmt.Sprintf(`
		SELECT
		 ti.descriptor_name as table_name,
		 ti.descriptor_id as table_id,
		 ti.index_name,
		 ti.index_id,
		 ti.index_type,
		 ti.is_unique,
		 ti.is_inverted,
		 total_reads,
		 last_read
		FROM %s.crdb_internal.index_usage_statistics AS us
		JOIN %s.crdb_internal.table_indexes ti
		ON us.index_id = ti.index_id
		 AND us.table_id = ti.descriptor_id
		ORDER BY total_reads ASC;
	`, databaseName, databaseName)

		it, err := ie.QueryIteratorEx(
			ctx,
			"capture-index-usage-stats",
			nil,
			sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
			stmt,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(576956)
			return err
		} else {
			__antithesis_instrumentation__.Notify(576957)
		}
		__antithesis_instrumentation__.Notify(576950)

		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(576958)
			var row tree.Datums
			if err != nil {
				__antithesis_instrumentation__.Notify(576963)
				return err
			} else {
				__antithesis_instrumentation__.Notify(576964)
			}
			__antithesis_instrumentation__.Notify(576959)
			if row = it.Cur(); row == nil {
				__antithesis_instrumentation__.Notify(576965)
				return errors.New("unexpected null row while capturing index usage stats")
			} else {
				__antithesis_instrumentation__.Notify(576966)
			}
			__antithesis_instrumentation__.Notify(576960)

			if row.Len() != expectedNumDatums {
				__antithesis_instrumentation__.Notify(576967)
				return errors.Newf("expected %d columns, received %d while capturing index usage stats", expectedNumDatums, row.Len())
			} else {
				__antithesis_instrumentation__.Notify(576968)
			}
			__antithesis_instrumentation__.Notify(576961)

			tableName := tree.MustBeDString(row[0])
			tableID := tree.MustBeDInt(row[1])
			indexName := tree.MustBeDString(row[2])
			indexID := tree.MustBeDInt(row[3])
			indexType := tree.MustBeDString(row[4])
			isUnique := tree.MustBeDBool(row[5])
			isInverted := tree.MustBeDBool(row[6])
			totalReads := tree.MustBeDInt(row[7])
			lastRead := time.Time{}
			if row[8] != tree.DNull {
				__antithesis_instrumentation__.Notify(576969)
				lastRead = tree.MustBeDTimestampTZ(row[8]).Time
			} else {
				__antithesis_instrumentation__.Notify(576970)
			}
			__antithesis_instrumentation__.Notify(576962)

			capturedIndexStats := &eventpb.CapturedIndexUsageStats{
				TableID:        uint32(roachpb.TableID(tableID)),
				IndexID:        uint32(roachpb.IndexID(indexID)),
				TotalReadCount: uint64(totalReads),
				LastRead:       lastRead.String(),
				DatabaseName:   databaseName,
				TableName:      string(tableName),
				IndexName:      string(indexName),
				IndexType:      string(indexType),
				IsUnique:       bool(isUnique),
				IsInverted:     bool(isInverted),
			}

			allCapturedIndexUsageStats = append(allCapturedIndexUsageStats, capturedIndexStats)
		}
		__antithesis_instrumentation__.Notify(576951)
		if err = it.Close(); err != nil {
			__antithesis_instrumentation__.Notify(576971)
			return err
		} else {
			__antithesis_instrumentation__.Notify(576972)
		}
	}
	__antithesis_instrumentation__.Notify(576945)
	logIndexUsageStatsWithDelay(ctx, allCapturedIndexUsageStats, stopper, loggingDelay)
	return nil
}

func logIndexUsageStatsWithDelay(
	ctx context.Context, events []eventpb.EventPayload, stopper *stop.Stopper, delay time.Duration,
) {
	__antithesis_instrumentation__.Notify(576973)

	timer := time.NewTimer(0 * time.Second)
	for len(events) > 0 {
		__antithesis_instrumentation__.Notify(576975)
		select {
		case <-stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(576976)
			timer.Stop()
			return
		case <-timer.C:
			__antithesis_instrumentation__.Notify(576977)
			event := events[0]
			log.StructuredEvent(ctx, event)
			events = events[1:]

			timer.Reset(delay)
		}
	}
	__antithesis_instrumentation__.Notify(576974)
	timer.Stop()
}

func getAllDatabaseNames(ctx context.Context, ie sqlutil.InternalExecutor) ([]string, error) {
	__antithesis_instrumentation__.Notify(576978)
	var allDatabaseNames []string
	var ok bool
	var expectedNumDatums = 1

	it, err := ie.QueryIteratorEx(
		ctx,
		"get-all-db-names",
		nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		`SELECT database_name FROM [SHOW DATABASES]`,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(576982)
		return []string{}, err
	} else {
		__antithesis_instrumentation__.Notify(576983)
	}
	__antithesis_instrumentation__.Notify(576979)

	defer func() { __antithesis_instrumentation__.Notify(576984); err = errors.CombineErrors(err, it.Close()) }()
	__antithesis_instrumentation__.Notify(576980)
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(576985)
		var row tree.Datums
		if row = it.Cur(); row == nil {
			__antithesis_instrumentation__.Notify(576988)
			return []string{}, errors.New("unexpected null row while capturing index usage stats")
		} else {
			__antithesis_instrumentation__.Notify(576989)
		}
		__antithesis_instrumentation__.Notify(576986)
		if row.Len() != expectedNumDatums {
			__antithesis_instrumentation__.Notify(576990)
			return []string{}, errors.Newf("expected %d columns, received %d while capturing index usage stats", expectedNumDatums, row.Len())
		} else {
			__antithesis_instrumentation__.Notify(576991)
		}
		__antithesis_instrumentation__.Notify(576987)

		databaseName := string(tree.MustBeDString(row[0]))
		allDatabaseNames = append(allDatabaseNames, databaseName)
	}
	__antithesis_instrumentation__.Notify(576981)
	return allDatabaseNames, nil
}
