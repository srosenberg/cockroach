package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

const AutoStatsClusterSettingName = "sql.stats.automatic_collection.enabled"

var AutomaticStatisticsClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	AutoStatsClusterSettingName,
	"automatic statistics collection mode",
	true,
).WithPublic()

var MultiColumnStatisticsClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.stats.multi_column_collection.enabled",
	"multi-column statistics collection mode",
	true,
).WithPublic()

var AutomaticStatisticsMaxIdleTime = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"sql.stats.automatic_collection.max_fraction_idle",
	"maximum fraction of time that automatic statistics sampler processors are idle",
	0.9,
	func(val float64) error {
		__antithesis_instrumentation__.Notify(626045)
		if val < 0 || func() bool {
			__antithesis_instrumentation__.Notify(626047)
			return val >= 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(626048)
			return pgerror.Newf(pgcode.InvalidParameterValue,
				"sql.stats.automatic_collection.max_fraction_idle must be >= 0 and < 1 but found: %v", val)
		} else {
			__antithesis_instrumentation__.Notify(626049)
		}
		__antithesis_instrumentation__.Notify(626046)
		return nil
	},
)

var AutomaticStatisticsFractionStaleRows = func() *settings.FloatSetting {
	__antithesis_instrumentation__.Notify(626050)
	s := settings.RegisterFloatSetting(
		settings.TenantWritable,
		"sql.stats.automatic_collection.fraction_stale_rows",
		"target fraction of stale rows per table that will trigger a statistics refresh",
		0.2,
		settings.NonNegativeFloat,
	)
	s.SetVisibility(settings.Public)
	return s
}()

var AutomaticStatisticsMinStaleRows = func() *settings.IntSetting {
	__antithesis_instrumentation__.Notify(626051)
	s := settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.stats.automatic_collection.min_stale_rows",
		"target minimum number of stale rows per table that will trigger a statistics refresh",
		500,
		settings.NonNegativeInt,
	)
	s.SetVisibility(settings.Public)
	return s
}()

var DefaultRefreshInterval = time.Minute

var DefaultAsOfTime = 30 * time.Second

var bufferedChanFullLogLimiter = log.Every(time.Second)

const (
	defaultAverageTimeBetweenRefreshes = 12 * time.Hour

	refreshChanBufferLen = 256
)

type Refresher struct {
	log.AmbientContext
	st      *cluster.Settings
	ex      sqlutil.InternalExecutor
	cache   *TableStatisticsCache
	randGen autoStatsRand

	mutations chan mutation

	asOfTime time.Duration

	extraTime time.Duration

	mutationCounts map[descpb.ID]int64
}

type mutation struct {
	tableID      descpb.ID
	rowsAffected int
}

func MakeRefresher(
	ambientCtx log.AmbientContext,
	st *cluster.Settings,
	ex sqlutil.InternalExecutor,
	cache *TableStatisticsCache,
	asOfTime time.Duration,
) *Refresher {
	__antithesis_instrumentation__.Notify(626052)
	randSource := rand.NewSource(rand.Int63())

	return &Refresher{
		AmbientContext: ambientCtx,
		st:             st,
		ex:             ex,
		cache:          cache,
		randGen:        makeAutoStatsRand(randSource),
		mutations:      make(chan mutation, refreshChanBufferLen),
		asOfTime:       asOfTime,
		extraTime:      time.Duration(rand.Int63n(int64(time.Hour))),
		mutationCounts: make(map[descpb.ID]int64, 16),
	}
}

func (r *Refresher) Start(
	ctx context.Context, stopper *stop.Stopper, refreshInterval time.Duration,
) error {
	__antithesis_instrumentation__.Notify(626053)
	bgCtx := r.AnnotateCtx(context.Background())
	_ = stopper.RunAsyncTask(bgCtx, "refresher", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(626055)

		refreshInterval -= r.asOfTime
		if refreshInterval < 0 {
			__antithesis_instrumentation__.Notify(626057)
			refreshInterval = 0
		} else {
			__antithesis_instrumentation__.Notify(626058)
		}
		__antithesis_instrumentation__.Notify(626056)

		timer := time.NewTimer(refreshInterval)
		defer timer.Stop()

		const initialTableCollectionDelay = time.Second
		initialTableCollection := time.After(initialTableCollectionDelay)

		for {
			__antithesis_instrumentation__.Notify(626059)
			select {
			case <-initialTableCollection:
				__antithesis_instrumentation__.Notify(626060)
				r.ensureAllTables(ctx, &r.st.SV, initialTableCollectionDelay)

			case <-timer.C:
				__antithesis_instrumentation__.Notify(626061)
				mutationCounts := r.mutationCounts
				if err := stopper.RunAsyncTask(
					ctx, "stats.Refresher: maybeRefreshStats", func(ctx context.Context) {
						__antithesis_instrumentation__.Notify(626065)

						timerAsOf := time.NewTimer(r.asOfTime)
						defer timerAsOf.Stop()
						select {
						case <-timerAsOf.C:
							__antithesis_instrumentation__.Notify(626068)
							break
						case <-stopper.ShouldQuiesce():
							__antithesis_instrumentation__.Notify(626069)
							return
						}
						__antithesis_instrumentation__.Notify(626066)

						for tableID, rowsAffected := range mutationCounts {
							__antithesis_instrumentation__.Notify(626070)

							if !AutomaticStatisticsClusterMode.Get(&r.st.SV) {
								__antithesis_instrumentation__.Notify(626072)
								break
							} else {
								__antithesis_instrumentation__.Notify(626073)
							}
							__antithesis_instrumentation__.Notify(626071)

							r.maybeRefreshStats(ctx, stopper, tableID, rowsAffected, r.asOfTime)

							select {
							case <-stopper.ShouldQuiesce():
								__antithesis_instrumentation__.Notify(626074)

								return
							default:
								__antithesis_instrumentation__.Notify(626075)
							}
						}
						__antithesis_instrumentation__.Notify(626067)
						timer.Reset(refreshInterval)
					}); err != nil {
					__antithesis_instrumentation__.Notify(626076)
					log.Errorf(ctx, "failed to refresh stats: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(626077)
				}
				__antithesis_instrumentation__.Notify(626062)
				r.mutationCounts = make(map[descpb.ID]int64, len(r.mutationCounts))

			case mut := <-r.mutations:
				__antithesis_instrumentation__.Notify(626063)
				r.mutationCounts[mut.tableID] += int64(mut.rowsAffected)

			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(626064)
				return
			}
		}
	})
	__antithesis_instrumentation__.Notify(626054)
	return nil
}

func (r *Refresher) ensureAllTables(
	ctx context.Context, settings *settings.Values, initialTableCollectionDelay time.Duration,
) {
	__antithesis_instrumentation__.Notify(626078)
	if !AutomaticStatisticsClusterMode.Get(settings) {
		__antithesis_instrumentation__.Notify(626081)

		return
	} else {
		__antithesis_instrumentation__.Notify(626082)
	}
	__antithesis_instrumentation__.Notify(626079)

	getAllTablesQuery := fmt.Sprintf(
		`
SELECT
	tbl.table_id
FROM
	crdb_internal.tables AS tbl
	INNER JOIN system.descriptor AS d ON d.id = tbl.table_id
		AS OF SYSTEM TIME '-%s'
WHERE
	tbl.database_name IS NOT NULL
	AND tbl.database_name <> '%s'
	AND tbl.drop_time IS NULL
	AND (
			crdb_internal.pb_to_json('cockroach.sql.sqlbase.Descriptor', d.descriptor, false)->'table'->>'viewQuery'
		) IS NULL;`,
		initialTableCollectionDelay,
		systemschema.SystemDatabaseName,
	)

	it, err := r.ex.QueryIterator(
		ctx,
		"get-tables",
		nil,
		getAllTablesQuery,
	)
	if err == nil {
		__antithesis_instrumentation__.Notify(626083)
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(626084)
			row := it.Cur()
			tableID := descpb.ID(*row[0].(*tree.DInt))

			if !descpb.IsVirtualTable(tableID) {
				__antithesis_instrumentation__.Notify(626085)
				r.mutationCounts[tableID] += 0
			} else {
				__antithesis_instrumentation__.Notify(626086)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(626087)
	}
	__antithesis_instrumentation__.Notify(626080)
	if err != nil {
		__antithesis_instrumentation__.Notify(626088)

		log.Errorf(ctx, "failed to get tables for automatic stats: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(626089)
	}
}

func (r *Refresher) NotifyMutation(table catalog.TableDescriptor, rowsAffected int) {
	__antithesis_instrumentation__.Notify(626090)
	if !AutomaticStatisticsClusterMode.Get(&r.st.SV) {
		__antithesis_instrumentation__.Notify(626093)

		return
	} else {
		__antithesis_instrumentation__.Notify(626094)
	}
	__antithesis_instrumentation__.Notify(626091)
	if !hasStatistics(table) {
		__antithesis_instrumentation__.Notify(626095)

		return
	} else {
		__antithesis_instrumentation__.Notify(626096)
	}
	__antithesis_instrumentation__.Notify(626092)

	select {
	case r.mutations <- mutation{tableID: table.GetID(), rowsAffected: rowsAffected}:
		__antithesis_instrumentation__.Notify(626097)
	default:
		__antithesis_instrumentation__.Notify(626098)

		if bufferedChanFullLogLimiter.ShouldLog() {
			__antithesis_instrumentation__.Notify(626099)
			log.Warningf(context.TODO(),
				"buffered channel is full. Unable to refresh stats for table %q (%d) with %d rows affected",
				table.GetName(), table.GetID(), rowsAffected)
		} else {
			__antithesis_instrumentation__.Notify(626100)
		}
	}
}

func (r *Refresher) maybeRefreshStats(
	ctx context.Context,
	stopper *stop.Stopper,
	tableID descpb.ID,
	rowsAffected int64,
	asOf time.Duration,
) {
	__antithesis_instrumentation__.Notify(626101)
	tableStats, err := r.cache.getTableStatsFromCache(ctx, tableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(626105)
		log.Errorf(ctx, "failed to get table statistics: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(626106)
	}
	__antithesis_instrumentation__.Notify(626102)

	var rowCount float64
	mustRefresh := false
	if stat := mostRecentAutomaticStat(tableStats); stat != nil {
		__antithesis_instrumentation__.Notify(626107)

		maxTimeBetweenRefreshes := stat.CreatedAt.Add(2*avgRefreshTime(tableStats) + r.extraTime)
		if timeutil.Now().After(maxTimeBetweenRefreshes) {
			__antithesis_instrumentation__.Notify(626109)
			mustRefresh = true
		} else {
			__antithesis_instrumentation__.Notify(626110)
		}
		__antithesis_instrumentation__.Notify(626108)
		rowCount = float64(stat.RowCount)
	} else {
		__antithesis_instrumentation__.Notify(626111)

		mustRefresh = true
	}
	__antithesis_instrumentation__.Notify(626103)

	targetRows := int64(rowCount*AutomaticStatisticsFractionStaleRows.Get(&r.st.SV)) +
		AutomaticStatisticsMinStaleRows.Get(&r.st.SV)
	if !mustRefresh && func() bool {
		__antithesis_instrumentation__.Notify(626112)
		return rowsAffected < math.MaxInt32 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(626113)
		return r.randGen.randInt(targetRows) >= rowsAffected == true
	}() == true {
		__antithesis_instrumentation__.Notify(626114)

		return
	} else {
		__antithesis_instrumentation__.Notify(626115)
	}
	__antithesis_instrumentation__.Notify(626104)

	if err := r.refreshStats(ctx, tableID, asOf); err != nil {
		__antithesis_instrumentation__.Notify(626116)
		if errors.Is(err, ConcurrentCreateStatsError) {
			__antithesis_instrumentation__.Notify(626118)

			if mustRefresh {
				__antithesis_instrumentation__.Notify(626120)

				r.mutations <- mutation{tableID: tableID, rowsAffected: 0}
			} else {
				__antithesis_instrumentation__.Notify(626121)

				r.mutations <- mutation{tableID: tableID, rowsAffected: math.MaxInt32}
			}
			__antithesis_instrumentation__.Notify(626119)
			return
		} else {
			__antithesis_instrumentation__.Notify(626122)
		}
		__antithesis_instrumentation__.Notify(626117)

		log.Warningf(ctx, "failed to create statistics on table %d: %v", tableID, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(626123)
	}
}

func (r *Refresher) refreshStats(ctx context.Context, tableID descpb.ID, asOf time.Duration) error {
	__antithesis_instrumentation__.Notify(626124)

	_, err := r.ex.Exec(
		ctx,
		"create-stats",
		nil,
		fmt.Sprintf(
			"CREATE STATISTICS %s FROM [%d] WITH OPTIONS THROTTLING %g AS OF SYSTEM TIME '-%s'",
			jobspb.AutoStatsName,
			tableID,
			AutomaticStatisticsMaxIdleTime.Get(&r.st.SV),
			asOf.String(),
		),
	)
	return err
}

func mostRecentAutomaticStat(tableStats []*TableStatistic) *TableStatistic {
	__antithesis_instrumentation__.Notify(626125)

	for _, stat := range tableStats {
		__antithesis_instrumentation__.Notify(626127)
		if stat.Name == jobspb.AutoStatsName {
			__antithesis_instrumentation__.Notify(626128)
			return stat
		} else {
			__antithesis_instrumentation__.Notify(626129)
		}
	}
	__antithesis_instrumentation__.Notify(626126)
	return nil
}

func avgRefreshTime(tableStats []*TableStatistic) time.Duration {
	__antithesis_instrumentation__.Notify(626130)
	var reference *TableStatistic
	var sum time.Duration
	var count int
	for _, stat := range tableStats {
		__antithesis_instrumentation__.Notify(626133)
		if stat.Name != jobspb.AutoStatsName {
			__antithesis_instrumentation__.Notify(626137)
			continue
		} else {
			__antithesis_instrumentation__.Notify(626138)
		}
		__antithesis_instrumentation__.Notify(626134)
		if reference == nil {
			__antithesis_instrumentation__.Notify(626139)
			reference = stat
			continue
		} else {
			__antithesis_instrumentation__.Notify(626140)
		}
		__antithesis_instrumentation__.Notify(626135)
		if !areEqual(stat.ColumnIDs, reference.ColumnIDs) {
			__antithesis_instrumentation__.Notify(626141)
			continue
		} else {
			__antithesis_instrumentation__.Notify(626142)
		}
		__antithesis_instrumentation__.Notify(626136)

		sum += reference.CreatedAt.Sub(stat.CreatedAt)
		count++
		reference = stat
	}
	__antithesis_instrumentation__.Notify(626131)
	if count == 0 {
		__antithesis_instrumentation__.Notify(626143)
		return defaultAverageTimeBetweenRefreshes
	} else {
		__antithesis_instrumentation__.Notify(626144)
	}
	__antithesis_instrumentation__.Notify(626132)
	return sum / time.Duration(count)
}

func areEqual(a, b []descpb.ColumnID) bool {
	__antithesis_instrumentation__.Notify(626145)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(626148)
		return false
	} else {
		__antithesis_instrumentation__.Notify(626149)
	}
	__antithesis_instrumentation__.Notify(626146)
	for i := range a {
		__antithesis_instrumentation__.Notify(626150)
		if a[i] != b[i] {
			__antithesis_instrumentation__.Notify(626151)
			return false
		} else {
			__antithesis_instrumentation__.Notify(626152)
		}
	}
	__antithesis_instrumentation__.Notify(626147)
	return true
}

type autoStatsRand struct {
	*syncutil.Mutex
	*rand.Rand
}

func makeAutoStatsRand(source rand.Source) autoStatsRand {
	__antithesis_instrumentation__.Notify(626153)
	return autoStatsRand{
		Mutex: &syncutil.Mutex{},
		Rand:  rand.New(source),
	}
}

func (r autoStatsRand) randInt(n int64) int64 {
	__antithesis_instrumentation__.Notify(626154)
	r.Lock()
	defer r.Unlock()
	return r.Int63n(n)
}

type concurrentCreateStatisticsError struct{}

var _ error = concurrentCreateStatisticsError{}

func (concurrentCreateStatisticsError) Error() string {
	__antithesis_instrumentation__.Notify(626155)
	return "another CREATE STATISTICS job is already running"
}

var ConcurrentCreateStatsError error = concurrentCreateStatisticsError{}
