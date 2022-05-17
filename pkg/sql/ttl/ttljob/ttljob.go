package ttljob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/metric/aggmetric"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

var (
	defaultSelectBatchSize = settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.ttl.default_select_batch_size",
		"default amount of rows to select in a single query during a TTL job",
		500,
		settings.PositiveInt,
	).WithPublic()
	defaultDeleteBatchSize = settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.ttl.default_delete_batch_size",
		"default amount of rows to delete in a single query during a TTL job",
		100,
		settings.PositiveInt,
	).WithPublic()
	defaultRangeConcurrency = settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.ttl.default_range_concurrency",
		"default amount of ranges to process at once during a TTL delete",
		1,
		settings.PositiveInt,
	).WithPublic()
	defaultDeleteRateLimit = settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.ttl.default_delete_rate_limit",
		"default delete rate limit for all TTL jobs. Use 0 to signify no rate limit.",
		0,
		settings.NonNegativeInt,
	).WithPublic()

	jobEnabled = settings.RegisterBoolSetting(
		settings.TenantWritable,
		"sql.ttl.job.enabled",
		"whether the TTL job is enabled",
		true,
	).WithPublic()
	rangeBatchSize = settings.RegisterIntSetting(
		settings.TenantWritable,
		"sql.ttl.range_batch_size",
		"amount of ranges to fetch at a time for a table during the TTL job",
		100,
		settings.PositiveInt,
	).WithPublic()
)

type rowLevelTTLResumer struct {
	job *jobs.Job
	st  *cluster.Settings
}

type RowLevelTTLAggMetrics struct {
	RangeTotalDuration *aggmetric.AggHistogram
	SelectDuration     *aggmetric.AggHistogram
	DeleteDuration     *aggmetric.AggHistogram
	RowSelections      *aggmetric.AggCounter
	RowDeletions       *aggmetric.AggCounter
	NumActiveRanges    *aggmetric.AggGauge
	TotalRows          *aggmetric.AggGauge
	TotalExpiredRows   *aggmetric.AggGauge

	defaultRowLevelMetrics rowLevelTTLMetrics
	mu                     struct {
		syncutil.Mutex
		m map[string]rowLevelTTLMetrics
	}
}

type rowLevelTTLMetrics struct {
	RangeTotalDuration *aggmetric.Histogram
	SelectDuration     *aggmetric.Histogram
	DeleteDuration     *aggmetric.Histogram
	RowSelections      *aggmetric.Counter
	RowDeletions       *aggmetric.Counter
	NumActiveRanges    *aggmetric.Gauge
	TotalRows          *aggmetric.Gauge
	TotalExpiredRows   *aggmetric.Gauge
}

func (m *RowLevelTTLAggMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(628589) }

func (m *RowLevelTTLAggMetrics) metricsWithChildren(children ...string) rowLevelTTLMetrics {
	__antithesis_instrumentation__.Notify(628590)
	return rowLevelTTLMetrics{
		RangeTotalDuration: m.RangeTotalDuration.AddChild(children...),
		SelectDuration:     m.SelectDuration.AddChild(children...),
		DeleteDuration:     m.DeleteDuration.AddChild(children...),
		RowSelections:      m.RowSelections.AddChild(children...),
		RowDeletions:       m.RowDeletions.AddChild(children...),
		NumActiveRanges:    m.NumActiveRanges.AddChild(children...),
		TotalRows:          m.TotalRows.AddChild(children...),
		TotalExpiredRows:   m.TotalExpiredRows.AddChild(children...),
	}
}

var invalidPrometheusRe = regexp.MustCompile(`[^a-zA-Z0-9_]`)

func (m *RowLevelTTLAggMetrics) loadMetrics(labelMetrics bool, relation string) rowLevelTTLMetrics {
	__antithesis_instrumentation__.Notify(628591)
	if !labelMetrics {
		__antithesis_instrumentation__.Notify(628594)
		return m.defaultRowLevelMetrics
	} else {
		__antithesis_instrumentation__.Notify(628595)
	}
	__antithesis_instrumentation__.Notify(628592)
	m.mu.Lock()
	defer m.mu.Unlock()

	relation = invalidPrometheusRe.ReplaceAllString(relation, "_")
	if ret, ok := m.mu.m[relation]; ok {
		__antithesis_instrumentation__.Notify(628596)
		return ret
	} else {
		__antithesis_instrumentation__.Notify(628597)
	}
	__antithesis_instrumentation__.Notify(628593)
	ret := m.metricsWithChildren(relation)
	m.mu.m[relation] = ret
	return ret
}

func makeRowLevelTTLAggMetrics(histogramWindowInterval time.Duration) metric.Struct {
	__antithesis_instrumentation__.Notify(628598)
	sigFigs := 2
	b := aggmetric.MakeBuilder("relation")
	ret := &RowLevelTTLAggMetrics{
		RangeTotalDuration: b.Histogram(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.range_total_duration",
				Help:        "Duration for processing a range during row level TTL.",
				Measurement: "nanoseconds",
				Unit:        metric.Unit_NANOSECONDS,
				MetricType:  io_prometheus_client.MetricType_HISTOGRAM,
			},
			histogramWindowInterval,
			time.Hour.Nanoseconds(),
			sigFigs,
		),
		SelectDuration: b.Histogram(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.select_duration",
				Help:        "Duration for select requests during row level TTL.",
				Measurement: "nanoseconds",
				Unit:        metric.Unit_NANOSECONDS,
				MetricType:  io_prometheus_client.MetricType_HISTOGRAM,
			},
			histogramWindowInterval,
			time.Minute.Nanoseconds(),
			sigFigs,
		),
		DeleteDuration: b.Histogram(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.delete_duration",
				Help:        "Duration for delete requests during row level TTL.",
				Measurement: "nanoseconds",
				Unit:        metric.Unit_NANOSECONDS,
				MetricType:  io_prometheus_client.MetricType_HISTOGRAM,
			},
			histogramWindowInterval,
			time.Minute.Nanoseconds(),
			sigFigs,
		),
		RowSelections: b.Counter(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.rows_selected",
				Help:        "Number of rows selected for deletion by the row level TTL job.",
				Measurement: "num_rows",
				Unit:        metric.Unit_COUNT,
				MetricType:  io_prometheus_client.MetricType_COUNTER,
			},
		),
		RowDeletions: b.Counter(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.rows_deleted",
				Help:        "Number of rows deleted by the row level TTL job.",
				Measurement: "num_rows",
				Unit:        metric.Unit_COUNT,
				MetricType:  io_prometheus_client.MetricType_COUNTER,
			},
		),
		NumActiveRanges: b.Gauge(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.num_active_ranges",
				Help:        "Number of active workers attempting to delete for row level TTL.",
				Measurement: "num_active_workers",
				Unit:        metric.Unit_COUNT,
			},
		),
		TotalRows: b.Gauge(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.total_rows",
				Help:        "Approximate number of rows on the TTL table.",
				Measurement: "total_rows",
				Unit:        metric.Unit_COUNT,
			},
		),
		TotalExpiredRows: b.Gauge(
			metric.Metadata{
				Name:        "jobs.row_level_ttl.total_expired_rows",
				Help:        "Approximate number of rows that have expired the TTL on the TTL table.",
				Measurement: "total_expired_rows",
				Unit:        metric.Unit_COUNT,
			},
		),
	}
	ret.defaultRowLevelMetrics = ret.metricsWithChildren("default")
	ret.mu.m = make(map[string]rowLevelTTLMetrics)
	return ret
}

var _ jobs.Resumer = (*rowLevelTTLResumer)(nil)

func (t rowLevelTTLResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(628599)
	p := execCtx.(sql.JobExecContext)
	db := p.ExecCfg().DB
	descsCol := p.ExtendedEvalContext().Descs

	if enabled := jobEnabled.Get(p.ExecCfg().SV()); !enabled {
		__antithesis_instrumentation__.Notify(628608)
		return errors.Newf(
			"ttl jobs are currently disabled by CLUSTER SETTING %s",
			jobEnabled.Key(),
		)
	} else {
		__antithesis_instrumentation__.Notify(628609)
	}
	__antithesis_instrumentation__.Notify(628600)

	telemetry.Inc(sqltelemetry.RowLevelTTLExecuted)

	var knobs sql.TTLTestingKnobs
	if ttlKnobs := p.ExecCfg().TTLTestingKnobs; ttlKnobs != nil {
		__antithesis_instrumentation__.Notify(628610)
		knobs = *ttlKnobs
	} else {
		__antithesis_instrumentation__.Notify(628611)
	}
	__antithesis_instrumentation__.Notify(628601)

	details := t.job.Details().(jobspb.RowLevelTTLDetails)

	aostDuration := -time.Second * 30
	if knobs.AOSTDuration != nil {
		__antithesis_instrumentation__.Notify(628612)
		aostDuration = *knobs.AOSTDuration
	} else {
		__antithesis_instrumentation__.Notify(628613)
	}
	__antithesis_instrumentation__.Notify(628602)
	aost, err := tree.MakeDTimestampTZ(timeutil.Now().Add(aostDuration), time.Microsecond)
	if err != nil {
		__antithesis_instrumentation__.Notify(628614)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628615)
	}
	__antithesis_instrumentation__.Notify(628603)

	var initialVersion descpb.DescriptorVersion

	var ttlSettings catpb.RowLevelTTL
	var pkColumns []string
	var pkTypes []*types.T
	var relationName string
	var rangeSpan roachpb.Span
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(628616)
		desc, err := descsCol.GetImmutableTableByID(
			ctx,
			txn,
			details.TableID,
			tree.ObjectLookupFlagsWithRequired(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(628623)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628624)
		}
		__antithesis_instrumentation__.Notify(628617)
		initialVersion = desc.GetVersion()

		if desc.GetModificationTime().GoTime().After(aost.Time) {
			__antithesis_instrumentation__.Notify(628625)
			return errors.Newf(
				"found a recent schema change on the table at %s, aborting",
				desc.GetModificationTime().GoTime().Format(time.RFC3339),
			)
		} else {
			__antithesis_instrumentation__.Notify(628626)
		}
		__antithesis_instrumentation__.Notify(628618)
		pkColumns = desc.GetPrimaryIndex().IndexDesc().KeyColumnNames
		for _, id := range desc.GetPrimaryIndex().IndexDesc().KeyColumnIDs {
			__antithesis_instrumentation__.Notify(628627)
			col, err := desc.FindColumnWithID(id)
			if err != nil {
				__antithesis_instrumentation__.Notify(628629)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628630)
			}
			__antithesis_instrumentation__.Notify(628628)
			pkTypes = append(pkTypes, col.GetType())
		}
		__antithesis_instrumentation__.Notify(628619)

		ttl := desc.GetRowLevelTTL()
		if ttl == nil {
			__antithesis_instrumentation__.Notify(628631)
			return errors.Newf("unable to find TTL on table %s", desc.GetName())
		} else {
			__antithesis_instrumentation__.Notify(628632)
		}
		__antithesis_instrumentation__.Notify(628620)

		if ttl.Pause {
			__antithesis_instrumentation__.Notify(628633)
			return errors.Newf("ttl jobs on table %s are currently paused", tree.Name(desc.GetName()))
		} else {
			__antithesis_instrumentation__.Notify(628634)
		}
		__antithesis_instrumentation__.Notify(628621)

		tn, err := descs.GetTableNameByDesc(ctx, txn, descsCol, desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(628635)
			return errors.Wrapf(err, "error fetching table relation name for TTL")
		} else {
			__antithesis_instrumentation__.Notify(628636)
		}
		__antithesis_instrumentation__.Notify(628622)

		relationName = tn.FQString()
		rangeSpan = desc.TableSpan(p.ExecCfg().Codec)
		ttlSettings = *ttl
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(628637)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628638)
	}
	__antithesis_instrumentation__.Notify(628604)

	var metrics = p.ExecCfg().JobRegistry.MetricsStruct().RowLevelTTL.(*RowLevelTTLAggMetrics).loadMetrics(
		ttlSettings.LabelMetrics,
		relationName,
	)
	var rangeDesc roachpb.RangeDescriptor
	var alloc tree.DatumAlloc
	type rangeToProcess struct {
		startPK, endPK tree.Datums
	}

	g := ctxgroup.WithContext(ctx)

	rangeConcurrency := getRangeConcurrency(p.ExecCfg().SV(), ttlSettings)
	selectBatchSize := getSelectBatchSize(p.ExecCfg().SV(), ttlSettings)
	deleteBatchSize := getDeleteBatchSize(p.ExecCfg().SV(), ttlSettings)
	deleteRateLimit := getDeleteRateLimit(p.ExecCfg().SV(), ttlSettings)
	deleteRateLimiter := quotapool.NewRateLimiter(
		"ttl-delete",
		quotapool.Limit(deleteRateLimit),
		deleteRateLimit,
	)

	statsCloseCh := make(chan struct{})
	ch := make(chan rangeToProcess, rangeConcurrency)
	for i := 0; i < rangeConcurrency; i++ {
		__antithesis_instrumentation__.Notify(628639)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(628640)
			for r := range ch {
				__antithesis_instrumentation__.Notify(628642)
				start := timeutil.Now()
				err := runTTLOnRange(
					ctx,
					p.ExecCfg(),
					details,
					p.ExtendedEvalContext().Descs,
					knobs,
					metrics,
					initialVersion,
					r.startPK,
					r.endPK,
					pkColumns,
					relationName,
					selectBatchSize,
					deleteBatchSize,
					deleteRateLimiter,
					*aost,
				)
				metrics.RangeTotalDuration.RecordValue(int64(timeutil.Since(start)))
				if err != nil {
					__antithesis_instrumentation__.Notify(628643)

					for r = range ch {
						__antithesis_instrumentation__.Notify(628645)
					}
					__antithesis_instrumentation__.Notify(628644)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628646)
				}
			}
			__antithesis_instrumentation__.Notify(628641)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(628605)

	if ttlSettings.RowStatsPollInterval != 0 {
		__antithesis_instrumentation__.Notify(628647)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(628648)

			fetchStatistics(ctx, p.ExecCfg(), knobs, relationName, details, metrics, aostDuration)

			for {
				__antithesis_instrumentation__.Notify(628649)
				select {
				case <-statsCloseCh:
					__antithesis_instrumentation__.Notify(628650)
					return nil
				case <-time.After(ttlSettings.RowStatsPollInterval):
					__antithesis_instrumentation__.Notify(628651)
					fetchStatistics(ctx, p.ExecCfg(), knobs, relationName, details, metrics, aostDuration)
				}
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(628652)
	}
	__antithesis_instrumentation__.Notify(628606)

	if err := func() (retErr error) {
		__antithesis_instrumentation__.Notify(628653)
		defer func() {
			__antithesis_instrumentation__.Notify(628656)
			close(ch)
			close(statsCloseCh)
			retErr = errors.CombineErrors(retErr, g.Wait())
		}()
		__antithesis_instrumentation__.Notify(628654)
		done := false

		batchSize := rangeBatchSize.Get(p.ExecCfg().SV())
		for !done {
			__antithesis_instrumentation__.Notify(628657)
			var ranges []kv.KeyValue

			if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(628659)
				metaStart := keys.RangeMetaKey(keys.MustAddr(rangeSpan.Key).Next())
				metaEnd := keys.RangeMetaKey(keys.MustAddr(rangeSpan.EndKey))

				kvs, err := txn.Scan(ctx, metaStart, metaEnd, batchSize)
				if err != nil {
					__antithesis_instrumentation__.Notify(628662)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628663)
				}
				__antithesis_instrumentation__.Notify(628660)
				if len(kvs) < int(batchSize) {
					__antithesis_instrumentation__.Notify(628664)
					done = true
					if len(kvs) == 0 || func() bool {
						__antithesis_instrumentation__.Notify(628665)
						return !kvs[len(kvs)-1].Key.Equal(metaEnd.AsRawKey()) == true
					}() == true {
						__antithesis_instrumentation__.Notify(628666)

						extraKV, err := txn.Scan(ctx, metaEnd, keys.Meta2Prefix.PrefixEnd(), 1)
						if err != nil {
							__antithesis_instrumentation__.Notify(628668)
							return err
						} else {
							__antithesis_instrumentation__.Notify(628669)
						}
						__antithesis_instrumentation__.Notify(628667)
						kvs = append(kvs, extraKV[0])
					} else {
						__antithesis_instrumentation__.Notify(628670)
					}
				} else {
					__antithesis_instrumentation__.Notify(628671)
				}
				__antithesis_instrumentation__.Notify(628661)
				ranges = kvs
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(628672)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628673)
			}
			__antithesis_instrumentation__.Notify(628658)

			for _, r := range ranges {
				__antithesis_instrumentation__.Notify(628674)
				if err := r.ValueProto(&rangeDesc); err != nil {
					__antithesis_instrumentation__.Notify(628678)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628679)
				}
				__antithesis_instrumentation__.Notify(628675)
				rangeSpan.Key = rangeDesc.EndKey.AsRawKey()
				var nextRange rangeToProcess
				nextRange.startPK, err = keyToDatums(rangeDesc.StartKey, p.ExecCfg().Codec, pkTypes, &alloc)
				if err != nil {
					__antithesis_instrumentation__.Notify(628680)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628681)
				}
				__antithesis_instrumentation__.Notify(628676)
				nextRange.endPK, err = keyToDatums(rangeDesc.EndKey, p.ExecCfg().Codec, pkTypes, &alloc)
				if err != nil {
					__antithesis_instrumentation__.Notify(628682)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628683)
				}
				__antithesis_instrumentation__.Notify(628677)
				ch <- nextRange
			}
		}
		__antithesis_instrumentation__.Notify(628655)
		return nil
	}(); err != nil {
		__antithesis_instrumentation__.Notify(628684)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628685)
	}
	__antithesis_instrumentation__.Notify(628607)
	return nil
}

func getSelectBatchSize(sv *settings.Values, ttl catpb.RowLevelTTL) int {
	__antithesis_instrumentation__.Notify(628686)
	if bs := ttl.SelectBatchSize; bs != 0 {
		__antithesis_instrumentation__.Notify(628688)
		return int(bs)
	} else {
		__antithesis_instrumentation__.Notify(628689)
	}
	__antithesis_instrumentation__.Notify(628687)
	return int(defaultSelectBatchSize.Get(sv))
}

func getDeleteBatchSize(sv *settings.Values, ttl catpb.RowLevelTTL) int {
	__antithesis_instrumentation__.Notify(628690)
	if bs := ttl.DeleteBatchSize; bs != 0 {
		__antithesis_instrumentation__.Notify(628692)
		return int(bs)
	} else {
		__antithesis_instrumentation__.Notify(628693)
	}
	__antithesis_instrumentation__.Notify(628691)
	return int(defaultDeleteBatchSize.Get(sv))
}

func getRangeConcurrency(sv *settings.Values, ttl catpb.RowLevelTTL) int {
	__antithesis_instrumentation__.Notify(628694)
	if rc := ttl.RangeConcurrency; rc != 0 {
		__antithesis_instrumentation__.Notify(628696)
		return int(rc)
	} else {
		__antithesis_instrumentation__.Notify(628697)
	}
	__antithesis_instrumentation__.Notify(628695)
	return int(defaultRangeConcurrency.Get(sv))
}

func getDeleteRateLimit(sv *settings.Values, ttl catpb.RowLevelTTL) int64 {
	__antithesis_instrumentation__.Notify(628698)
	val := func() int64 {
		__antithesis_instrumentation__.Notify(628701)
		if bs := ttl.DeleteRateLimit; bs != 0 {
			__antithesis_instrumentation__.Notify(628703)
			return bs
		} else {
			__antithesis_instrumentation__.Notify(628704)
		}
		__antithesis_instrumentation__.Notify(628702)
		return defaultDeleteRateLimit.Get(sv)
	}()
	__antithesis_instrumentation__.Notify(628699)

	if val == 0 {
		__antithesis_instrumentation__.Notify(628705)
		return math.MaxInt64
	} else {
		__antithesis_instrumentation__.Notify(628706)
	}
	__antithesis_instrumentation__.Notify(628700)
	return val
}

func fetchStatistics(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	knobs sql.TTLTestingKnobs,
	relationName string,
	details jobspb.RowLevelTTLDetails,
	metrics rowLevelTTLMetrics,
	aostDuration time.Duration,
) {
	__antithesis_instrumentation__.Notify(628707)
	if err := func() error {
		__antithesis_instrumentation__.Notify(628708)
		aost, err := tree.MakeDTimestampTZ(timeutil.Now().Add(aostDuration), time.Microsecond)
		if err != nil {
			__antithesis_instrumentation__.Notify(628711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628712)
		}
		__antithesis_instrumentation__.Notify(628709)
		for _, c := range []struct {
			opName string
			query  string
			args   []interface{}
			gauge  *aggmetric.Gauge
		}{
			{
				opName: fmt.Sprintf("ttl num rows stats %s", relationName),
				query:  `SELECT count(1) FROM [%d AS t] AS OF SYSTEM TIME %s`,
				gauge:  metrics.TotalRows,
			},
			{
				opName: fmt.Sprintf("ttl num expired rows stats %s", relationName),
				query:  `SELECT count(1) FROM [%d AS t] AS OF SYSTEM TIME %s WHERE ` + colinfo.TTLDefaultExpirationColumnName + ` < $1`,
				args:   []interface{}{details.Cutoff},
				gauge:  metrics.TotalExpiredRows,
			},
		} {
			__antithesis_instrumentation__.Notify(628713)

			qosLevel := sessiondatapb.SystemLow
			datums, err := execCfg.InternalExecutor.QueryRowEx(
				ctx,
				c.opName,
				nil,
				sessiondata.InternalExecutorOverride{
					User:             security.RootUserName(),
					QualityOfService: &qosLevel,
				},
				fmt.Sprintf(c.query, details.TableID, aost.String()),
				c.args...,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(628715)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628716)
			}
			__antithesis_instrumentation__.Notify(628714)
			c.gauge.Update(int64(tree.MustBeDInt(datums[0])))
		}
		__antithesis_instrumentation__.Notify(628710)
		return nil
	}(); err != nil {
		__antithesis_instrumentation__.Notify(628717)
		if onStatisticsError := knobs.OnStatisticsError; onStatisticsError != nil {
			__antithesis_instrumentation__.Notify(628719)
			onStatisticsError(err)
		} else {
			__antithesis_instrumentation__.Notify(628720)
		}
		__antithesis_instrumentation__.Notify(628718)
		log.Warningf(ctx, "failed to get statistics for table id %d: %s", details.TableID, err)
	} else {
		__antithesis_instrumentation__.Notify(628721)
	}
}

func runTTLOnRange(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	details jobspb.RowLevelTTLDetails,
	descriptors *descs.Collection,
	knobs sql.TTLTestingKnobs,
	metrics rowLevelTTLMetrics,
	tableVersion descpb.DescriptorVersion,
	startPK tree.Datums,
	endPK tree.Datums,
	pkColumns []string,
	relationName string,
	selectBatchSize, deleteBatchSize int,
	deleteRateLimiter *quotapool.RateLimiter,
	aost tree.DTimestampTZ,
) error {
	__antithesis_instrumentation__.Notify(628722)
	metrics.NumActiveRanges.Inc(1)
	defer metrics.NumActiveRanges.Dec(1)

	ie := execCfg.InternalExecutor
	db := execCfg.DB

	selectBuilder := makeSelectQueryBuilder(
		details.TableID,
		details.Cutoff,
		pkColumns,
		relationName,
		startPK,
		endPK,
		aost,
		selectBatchSize,
	)
	deleteBuilder := makeDeleteQueryBuilder(
		details.TableID,
		details.Cutoff,
		pkColumns,
		relationName,
		deleteBatchSize,
	)

	for {
		__antithesis_instrumentation__.Notify(628724)
		if f := knobs.OnDeleteLoopStart; f != nil {
			__antithesis_instrumentation__.Notify(628729)
			if err := f(); err != nil {
				__antithesis_instrumentation__.Notify(628730)
				return err
			} else {
				__antithesis_instrumentation__.Notify(628731)
			}
		} else {
			__antithesis_instrumentation__.Notify(628732)
		}
		__antithesis_instrumentation__.Notify(628725)

		if enabled := jobEnabled.Get(execCfg.SV()); !enabled {
			__antithesis_instrumentation__.Notify(628733)
			return errors.Newf(
				"ttl jobs are currently disabled by CLUSTER SETTING %s",
				jobEnabled.Key(),
			)
		} else {
			__antithesis_instrumentation__.Notify(628734)
		}
		__antithesis_instrumentation__.Notify(628726)

		start := timeutil.Now()
		expiredRowsPKs, err := selectBuilder.run(ctx, ie)
		metrics.DeleteDuration.RecordValue(int64(timeutil.Since(start)))
		if err != nil {
			__antithesis_instrumentation__.Notify(628735)
			return errors.Wrapf(err, "error selecting rows to delete")
		} else {
			__antithesis_instrumentation__.Notify(628736)
		}
		__antithesis_instrumentation__.Notify(628727)
		metrics.RowSelections.Inc(int64(len(expiredRowsPKs)))

		for startRowIdx := 0; startRowIdx < len(expiredRowsPKs); startRowIdx += deleteBatchSize {
			__antithesis_instrumentation__.Notify(628737)
			until := startRowIdx + deleteBatchSize
			if until > len(expiredRowsPKs) {
				__antithesis_instrumentation__.Notify(628740)
				until = len(expiredRowsPKs)
			} else {
				__antithesis_instrumentation__.Notify(628741)
			}
			__antithesis_instrumentation__.Notify(628738)
			deleteBatch := expiredRowsPKs[startRowIdx:until]
			if err := db.TxnWithSteppingEnabled(ctx, sessiondatapb.TTLLow, func(ctx context.Context, txn *kv.Txn) error {
				__antithesis_instrumentation__.Notify(628742)

				desc, err := descriptors.GetImmutableTableByID(
					ctx,
					txn,
					details.TableID,
					tree.ObjectLookupFlagsWithRequired(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(628747)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628748)
				}
				__antithesis_instrumentation__.Notify(628743)
				version := desc.GetVersion()
				if mockVersion := knobs.MockDescriptorVersionDuringDelete; mockVersion != nil {
					__antithesis_instrumentation__.Notify(628749)
					version = *mockVersion
				} else {
					__antithesis_instrumentation__.Notify(628750)
				}
				__antithesis_instrumentation__.Notify(628744)
				if version != tableVersion {
					__antithesis_instrumentation__.Notify(628751)
					return errors.Newf(
						"table has had a schema change since the job has started at %s, aborting",
						desc.GetModificationTime().GoTime().Format(time.RFC3339),
					)
				} else {
					__antithesis_instrumentation__.Notify(628752)
				}
				__antithesis_instrumentation__.Notify(628745)

				tokens, err := deleteRateLimiter.Acquire(ctx, int64(len(deleteBatch)))
				if err != nil {
					__antithesis_instrumentation__.Notify(628753)
					return err
				} else {
					__antithesis_instrumentation__.Notify(628754)
				}
				__antithesis_instrumentation__.Notify(628746)
				defer tokens.Consume()

				start := timeutil.Now()
				err = deleteBuilder.run(ctx, ie, txn, deleteBatch)
				metrics.DeleteDuration.RecordValue(int64(timeutil.Since(start)))
				return err
			}); err != nil {
				__antithesis_instrumentation__.Notify(628755)
				return errors.Wrapf(err, "error during row deletion")
			} else {
				__antithesis_instrumentation__.Notify(628756)
			}
			__antithesis_instrumentation__.Notify(628739)
			metrics.RowDeletions.Inc(int64(len(deleteBatch)))
		}
		__antithesis_instrumentation__.Notify(628728)

		if len(expiredRowsPKs) < selectBatchSize {
			__antithesis_instrumentation__.Notify(628757)
			break
		} else {
			__antithesis_instrumentation__.Notify(628758)
		}
	}
	__antithesis_instrumentation__.Notify(628723)
	return nil
}

func keyToDatums(
	key roachpb.RKey, codec keys.SQLCodec, pkTypes []*types.T, alloc *tree.DatumAlloc,
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(628759)
	rKey := key.AsRawKey()

	if _, _, err := codec.DecodeTablePrefix(rKey); err != nil {
		__antithesis_instrumentation__.Notify(628766)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(628767)
	}
	__antithesis_instrumentation__.Notify(628760)
	if _, _, _, err := codec.DecodeIndexPrefix(rKey); err != nil {
		__antithesis_instrumentation__.Notify(628768)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(628769)
	}
	__antithesis_instrumentation__.Notify(628761)

	rKey, err := codec.StripTenantPrefix(key.AsRawKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(628770)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628771)
	}
	__antithesis_instrumentation__.Notify(628762)
	rKey, _, _, err = rowenc.DecodePartialTableIDIndexID(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(628772)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(628773)
	}
	__antithesis_instrumentation__.Notify(628763)
	encDatums := make([]rowenc.EncDatum, 0, len(pkTypes))
	for len(rKey) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(628774)
		return len(encDatums) < len(pkTypes) == true
	}() == true {
		__antithesis_instrumentation__.Notify(628775)
		i := len(encDatums)

		enc := descpb.DatumEncoding_ASCENDING_KEY
		var val rowenc.EncDatum
		val, rKey, err = rowenc.EncDatumFromBuffer(pkTypes[i], enc, rKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(628777)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(628778)
		}
		__antithesis_instrumentation__.Notify(628776)
		encDatums = append(encDatums, val)
	}
	__antithesis_instrumentation__.Notify(628764)

	datums := make(tree.Datums, len(encDatums))
	for i, encDatum := range encDatums {
		__antithesis_instrumentation__.Notify(628779)
		if err := encDatum.EnsureDecoded(pkTypes[i], alloc); err != nil {
			__antithesis_instrumentation__.Notify(628781)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(628782)
		}
		__antithesis_instrumentation__.Notify(628780)
		datums[i] = encDatum.Datum
	}
	__antithesis_instrumentation__.Notify(628765)
	return datums, nil
}

func (t rowLevelTTLResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(628783)
	return nil
}

func init() {
	jobs.RegisterConstructor(jobspb.TypeRowLevelTTL, func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &rowLevelTTLResumer{
			job: job,
			st:  settings,
		}
	})
	jobs.MakeRowLevelTTLMetricsHook = makeRowLevelTTLAggMetrics
}
