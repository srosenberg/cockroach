package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/featureflag"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var createStatsPostEvents = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.stats.post_events.enabled",
	"if set, an event is logged for every CREATE STATISTICS job",
	false,
).WithPublic()

var featureStatsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"feature.stats.enabled",
	"set to true to enable CREATE STATISTICS/ANALYZE, false to disable; default is true",
	featureflag.FeatureFlagEnabledDefault,
).WithPublic()

const nonIndexColHistogramBuckets = 2

func StubTableStats(
	desc catalog.TableDescriptor, name string, multiColEnabled bool,
) ([]*stats.TableStatisticProto, error) {
	__antithesis_instrumentation__.Notify(463537)
	colStats, err := createStatsDefaultColumns(desc, multiColEnabled)
	if err != nil {
		__antithesis_instrumentation__.Notify(463540)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463541)
	}
	__antithesis_instrumentation__.Notify(463538)
	statistics := make([]*stats.TableStatisticProto, len(colStats))
	for i, colStat := range colStats {
		__antithesis_instrumentation__.Notify(463542)
		statistics[i] = &stats.TableStatisticProto{
			TableID:   desc.GetID(),
			Name:      name,
			ColumnIDs: colStat.ColumnIDs,
		}
	}
	__antithesis_instrumentation__.Notify(463539)
	return statistics, nil
}

type createStatsNode struct {
	tree.CreateStats
	p *planner

	runAsJob bool

	run createStatsRun
}

type createStatsRun struct {
	resultsCh chan tree.Datums
	errCh     chan error
}

func (n *createStatsNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(463543)
	telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter("stats"))
	n.run.resultsCh = make(chan tree.Datums)
	n.run.errCh = make(chan error)
	go func() {
		__antithesis_instrumentation__.Notify(463545)
		err := n.startJob(params.ctx, n.run.resultsCh)
		select {
		case <-params.ctx.Done():
			__antithesis_instrumentation__.Notify(463547)
		case n.run.errCh <- err:
			__antithesis_instrumentation__.Notify(463548)
		}
		__antithesis_instrumentation__.Notify(463546)
		close(n.run.errCh)
		close(n.run.resultsCh)
	}()
	__antithesis_instrumentation__.Notify(463544)
	return nil
}

func (n *createStatsNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(463549)
	select {
	case <-params.ctx.Done():
		__antithesis_instrumentation__.Notify(463550)
		return false, params.ctx.Err()
	case err := <-n.run.errCh:
		__antithesis_instrumentation__.Notify(463551)
		return false, err
	case <-n.run.resultsCh:
		__antithesis_instrumentation__.Notify(463552)
		return true, nil
	}
}

func (*createStatsNode) Close(context.Context) { __antithesis_instrumentation__.Notify(463553) }
func (*createStatsNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(463554)
	return nil
}

func (n *createStatsNode) startJob(ctx context.Context, resultsCh chan<- tree.Datums) error {
	__antithesis_instrumentation__.Notify(463555)
	record, err := n.makeJobRecord(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(463561)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463562)
	}
	__antithesis_instrumentation__.Notify(463556)

	if n.Name == jobspb.AutoStatsName {
		__antithesis_instrumentation__.Notify(463563)

		if err := checkRunningJobs(ctx, nil, n.p); err != nil {
			__antithesis_instrumentation__.Notify(463564)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463565)
		}
	} else {
		__antithesis_instrumentation__.Notify(463566)
		telemetry.Inc(sqltelemetry.CreateStatisticsUseCounter)
	}
	__antithesis_instrumentation__.Notify(463557)

	var job *jobs.StartableJob
	jobID := n.p.ExecCfg().JobRegistry.MakeJobID()
	if err := n.p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(463567)
		return n.p.ExecCfg().JobRegistry.CreateStartableJobWithTxn(ctx, &job, jobID, txn, *record)
	}); err != nil {
		__antithesis_instrumentation__.Notify(463568)
		if job != nil {
			__antithesis_instrumentation__.Notify(463570)
			if cleanupErr := job.CleanupOnRollback(ctx); cleanupErr != nil {
				__antithesis_instrumentation__.Notify(463571)
				log.Warningf(ctx, "failed to cleanup StartableJob: %v", cleanupErr)
			} else {
				__antithesis_instrumentation__.Notify(463572)
			}
		} else {
			__antithesis_instrumentation__.Notify(463573)
		}
		__antithesis_instrumentation__.Notify(463569)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463574)
	}
	__antithesis_instrumentation__.Notify(463558)
	if err := job.Start(ctx); err != nil {
		__antithesis_instrumentation__.Notify(463575)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463576)
	}
	__antithesis_instrumentation__.Notify(463559)

	if err := job.AwaitCompletion(ctx); err != nil {
		__antithesis_instrumentation__.Notify(463577)
		if errors.Is(err, stats.ConcurrentCreateStatsError) {
			__antithesis_instrumentation__.Notify(463579)

			const stmt = `DELETE FROM system.jobs WHERE id = $1`
			if _, delErr := n.p.ExecCfg().InternalExecutor.Exec(
				ctx, "delete-job", nil, stmt, jobID,
			); delErr != nil {
				__antithesis_instrumentation__.Notify(463580)
				log.Warningf(ctx, "failed to delete job: %v", delErr)
			} else {
				__antithesis_instrumentation__.Notify(463581)
			}
		} else {
			__antithesis_instrumentation__.Notify(463582)
		}
		__antithesis_instrumentation__.Notify(463578)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463583)
	}
	__antithesis_instrumentation__.Notify(463560)
	return nil
}

func (n *createStatsNode) makeJobRecord(ctx context.Context) (*jobs.Record, error) {
	__antithesis_instrumentation__.Notify(463584)
	var tableDesc catalog.TableDescriptor
	var fqTableName string
	var err error
	switch t := n.Table.(type) {
	case *tree.UnresolvedObjectName:
		__antithesis_instrumentation__.Notify(463592)
		tableDesc, err = n.p.ResolveExistingObjectEx(ctx, t, true, tree.ResolveRequireTableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(463597)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463598)
		}
		__antithesis_instrumentation__.Notify(463593)
		fqTableName = n.p.ResolvedName(t).FQString()

	case *tree.TableRef:
		__antithesis_instrumentation__.Notify(463594)
		flags := tree.ObjectLookupFlags{CommonLookupFlags: tree.CommonLookupFlags{
			AvoidLeased: n.p.avoidLeasedDescriptors,
		}}
		tableDesc, err = n.p.Descriptors().GetImmutableTableByID(ctx, n.p.txn, descpb.ID(t.TableID), flags)
		if err != nil {
			__antithesis_instrumentation__.Notify(463599)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463600)
		}
		__antithesis_instrumentation__.Notify(463595)
		fqName, err := n.p.getQualifiedTableName(ctx, tableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(463601)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463602)
		}
		__antithesis_instrumentation__.Notify(463596)
		fqTableName = fqName.FQString()
	}
	__antithesis_instrumentation__.Notify(463585)

	if tableDesc.IsVirtualTable() {
		__antithesis_instrumentation__.Notify(463603)
		return nil, pgerror.New(
			pgcode.WrongObjectType, "cannot create statistics on virtual tables",
		)
	} else {
		__antithesis_instrumentation__.Notify(463604)
	}
	__antithesis_instrumentation__.Notify(463586)

	if tableDesc.IsView() {
		__antithesis_instrumentation__.Notify(463605)
		return nil, pgerror.New(
			pgcode.WrongObjectType, "cannot create statistics on views",
		)
	} else {
		__antithesis_instrumentation__.Notify(463606)
	}
	__antithesis_instrumentation__.Notify(463587)

	if err := n.p.CheckPrivilege(ctx, tableDesc, privilege.SELECT); err != nil {
		__antithesis_instrumentation__.Notify(463607)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463608)
	}
	__antithesis_instrumentation__.Notify(463588)

	var colStats []jobspb.CreateStatsDetails_ColStat
	if len(n.ColumnNames) == 0 {
		__antithesis_instrumentation__.Notify(463609)
		multiColEnabled := stats.MultiColumnStatisticsClusterMode.Get(&n.p.ExecCfg().Settings.SV)
		if colStats, err = createStatsDefaultColumns(tableDesc, multiColEnabled); err != nil {
			__antithesis_instrumentation__.Notify(463610)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463611)
		}
	} else {
		__antithesis_instrumentation__.Notify(463612)
		columns, err := tabledesc.FindPublicColumnsWithNames(tableDesc, n.ColumnNames)
		if err != nil {
			__antithesis_instrumentation__.Notify(463616)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463617)
		}
		__antithesis_instrumentation__.Notify(463613)

		columnIDs := make([]descpb.ColumnID, len(columns))
		for i := range columns {
			__antithesis_instrumentation__.Notify(463618)
			if columns[i].IsVirtual() {
				__antithesis_instrumentation__.Notify(463620)
				return nil, pgerror.Newf(
					pgcode.InvalidColumnReference,
					"cannot create statistics on virtual column %q",
					columns[i].ColName(),
				)
			} else {
				__antithesis_instrumentation__.Notify(463621)
			}
			__antithesis_instrumentation__.Notify(463619)
			columnIDs[i] = columns[i].GetID()
		}
		__antithesis_instrumentation__.Notify(463614)
		col, err := tableDesc.FindColumnWithID(columnIDs[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(463622)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463623)
		}
		__antithesis_instrumentation__.Notify(463615)
		isInvIndex := colinfo.ColumnTypeIsInvertedIndexable(col.GetType())
		colStats = []jobspb.CreateStatsDetails_ColStat{{
			ColumnIDs: columnIDs,

			HasHistogram: len(columnIDs) == 1 && func() bool {
				__antithesis_instrumentation__.Notify(463624)
				return !isInvIndex == true
			}() == true,
			HistogramMaxBuckets: stats.DefaultHistogramBuckets,
		}}

		if len(columnIDs) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(463625)
			return isInvIndex == true
		}() == true {
			__antithesis_instrumentation__.Notify(463626)
			colStats = append(colStats, jobspb.CreateStatsDetails_ColStat{
				ColumnIDs:           columnIDs,
				HasHistogram:        true,
				Inverted:            true,
				HistogramMaxBuckets: stats.DefaultHistogramBuckets,
			})
		} else {
			__antithesis_instrumentation__.Notify(463627)
		}
	}
	__antithesis_instrumentation__.Notify(463589)

	var asOfTimestamp *hlc.Timestamp
	if n.Options.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(463628)
		asOf, err := n.p.EvalAsOfTimestamp(ctx, n.Options.AsOf)
		if err != nil {
			__antithesis_instrumentation__.Notify(463630)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463631)
		}
		__antithesis_instrumentation__.Notify(463629)
		asOfTimestamp = &asOf.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(463632)
	}
	__antithesis_instrumentation__.Notify(463590)

	statement := tree.AsStringWithFQNames(n, n.p.EvalContext().Annotations)
	eventLogStatement := statement
	var description string
	if n.Name == jobspb.AutoStatsName {
		__antithesis_instrumentation__.Notify(463633)

		description = fmt.Sprintf("Table statistics refresh for %s", fqTableName)
	} else {
		__antithesis_instrumentation__.Notify(463634)

		description = statement
		statement = ""
	}
	__antithesis_instrumentation__.Notify(463591)
	return &jobs.Record{
		Description: description,
		Statements:  []string{statement},
		Username:    n.p.User(),
		Details: jobspb.CreateStatsDetails{
			Name:            string(n.Name),
			FQTableName:     fqTableName,
			Table:           *tableDesc.TableDesc(),
			ColumnStats:     colStats,
			Statement:       eventLogStatement,
			AsOf:            asOfTimestamp,
			MaxFractionIdle: n.Options.Throttling,
		},
		Progress: jobspb.CreateStatsProgress{},
	}, nil
}

const maxNonIndexCols = 100

func createStatsDefaultColumns(
	desc catalog.TableDescriptor, multiColEnabled bool,
) ([]jobspb.CreateStatsDetails_ColStat, error) {
	__antithesis_instrumentation__.Notify(463635)
	colStats := make([]jobspb.CreateStatsDetails_ColStat, 0, len(desc.ActiveIndexes()))

	requestedStats := make(map[string]struct{})

	trackStatsIfNotExists := func(colIDs []descpb.ColumnID) bool {
		__antithesis_instrumentation__.Notify(463641)
		key := makeColStatKey(colIDs)
		if _, ok := requestedStats[key]; ok {
			__antithesis_instrumentation__.Notify(463643)
			return false
		} else {
			__antithesis_instrumentation__.Notify(463644)
		}
		__antithesis_instrumentation__.Notify(463642)
		requestedStats[key] = struct{}{}
		return true
	}
	__antithesis_instrumentation__.Notify(463636)

	addIndexColumnStatsIfNotExists := func(colID descpb.ColumnID, isInverted bool) error {
		__antithesis_instrumentation__.Notify(463645)
		col, err := desc.FindColumnWithID(colID)
		if err != nil {
			__antithesis_instrumentation__.Notify(463650)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463651)
		}
		__antithesis_instrumentation__.Notify(463646)

		if col.IsVirtual() {
			__antithesis_instrumentation__.Notify(463652)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(463653)
		}
		__antithesis_instrumentation__.Notify(463647)

		colList := []descpb.ColumnID{colID}

		if !trackStatsIfNotExists(colList) {
			__antithesis_instrumentation__.Notify(463654)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(463655)
		}
		__antithesis_instrumentation__.Notify(463648)

		colStat := jobspb.CreateStatsDetails_ColStat{
			ColumnIDs:           colList,
			HasHistogram:        !isInverted,
			HistogramMaxBuckets: stats.DefaultHistogramBuckets,
		}
		colStats = append(colStats, colStat)

		if isInverted {
			__antithesis_instrumentation__.Notify(463656)
			colStat.Inverted = true
			colStat.HasHistogram = true
			colStats = append(colStats, colStat)
		} else {
			__antithesis_instrumentation__.Notify(463657)
		}
		__antithesis_instrumentation__.Notify(463649)

		return nil
	}
	__antithesis_instrumentation__.Notify(463637)

	primaryIdx := desc.GetPrimaryIndex()
	for i := 0; i < primaryIdx.NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(463658)

		err := addIndexColumnStatsIfNotExists(primaryIdx.GetKeyColumnID(i), false)
		if err != nil {
			__antithesis_instrumentation__.Notify(463662)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463663)
		}
		__antithesis_instrumentation__.Notify(463659)

		if i == 0 || func() bool {
			__antithesis_instrumentation__.Notify(463664)
			return !multiColEnabled == true
		}() == true {
			__antithesis_instrumentation__.Notify(463665)
			continue
		} else {
			__antithesis_instrumentation__.Notify(463666)
		}
		__antithesis_instrumentation__.Notify(463660)

		colIDs := make([]descpb.ColumnID, i+1)
		for j := 0; j <= i; j++ {
			__antithesis_instrumentation__.Notify(463667)
			colIDs[j] = desc.GetPrimaryIndex().GetKeyColumnID(j)
		}
		__antithesis_instrumentation__.Notify(463661)

		trackStatsIfNotExists(colIDs)

		colStats = append(colStats, jobspb.CreateStatsDetails_ColStat{
			ColumnIDs:    colIDs,
			HasHistogram: false,
		})
	}
	__antithesis_instrumentation__.Notify(463638)

	for _, idx := range desc.PublicNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(463668)
		for j, n := 0, idx.NumKeyColumns(); j < n; j++ {
			__antithesis_instrumentation__.Notify(463670)
			colID := idx.GetKeyColumnID(j)
			isInverted := idx.GetType() == descpb.IndexDescriptor_INVERTED && func() bool {
				__antithesis_instrumentation__.Notify(463676)
				return colID == idx.InvertedColumnID() == true
			}() == true

			if err := addIndexColumnStatsIfNotExists(colID, isInverted); err != nil {
				__antithesis_instrumentation__.Notify(463677)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(463678)
			}
			__antithesis_instrumentation__.Notify(463671)

			if j == 0 || func() bool {
				__antithesis_instrumentation__.Notify(463679)
				return !multiColEnabled == true
			}() == true {
				__antithesis_instrumentation__.Notify(463680)
				continue
			} else {
				__antithesis_instrumentation__.Notify(463681)
			}
			__antithesis_instrumentation__.Notify(463672)

			colIDs := make([]descpb.ColumnID, 0, j+1)
			for k := 0; k <= j; k++ {
				__antithesis_instrumentation__.Notify(463682)
				col, err := desc.FindColumnWithID(idx.GetKeyColumnID(k))
				if err != nil {
					__antithesis_instrumentation__.Notify(463685)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(463686)
				}
				__antithesis_instrumentation__.Notify(463683)
				if col.IsVirtual() {
					__antithesis_instrumentation__.Notify(463687)
					continue
				} else {
					__antithesis_instrumentation__.Notify(463688)
				}
				__antithesis_instrumentation__.Notify(463684)
				colIDs = append(colIDs, col.GetID())
			}
			__antithesis_instrumentation__.Notify(463673)

			if len(colIDs) == 0 {
				__antithesis_instrumentation__.Notify(463689)
				continue
			} else {
				__antithesis_instrumentation__.Notify(463690)
			}
			__antithesis_instrumentation__.Notify(463674)

			if !trackStatsIfNotExists(colIDs) {
				__antithesis_instrumentation__.Notify(463691)
				continue
			} else {
				__antithesis_instrumentation__.Notify(463692)
			}
			__antithesis_instrumentation__.Notify(463675)

			colStats = append(colStats, jobspb.CreateStatsDetails_ColStat{
				ColumnIDs:    colIDs,
				HasHistogram: false,
			})
		}
		__antithesis_instrumentation__.Notify(463669)

		if idx.IsPartial() {
			__antithesis_instrumentation__.Notify(463693)
			expr, err := parser.ParseExpr(idx.GetPredicate())
			if err != nil {
				__antithesis_instrumentation__.Notify(463696)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(463697)
			}
			__antithesis_instrumentation__.Notify(463694)

			colIDs, err := schemaexpr.ExtractColumnIDs(desc, expr)
			if err != nil {
				__antithesis_instrumentation__.Notify(463698)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(463699)
			}
			__antithesis_instrumentation__.Notify(463695)

			for _, colID := range colIDs.Ordered() {
				__antithesis_instrumentation__.Notify(463700)
				col, err := desc.FindColumnWithID(colID)
				if err != nil {
					__antithesis_instrumentation__.Notify(463702)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(463703)
				}
				__antithesis_instrumentation__.Notify(463701)
				isInverted := colinfo.ColumnTypeIsInvertedIndexable(col.GetType())
				if err := addIndexColumnStatsIfNotExists(colID, isInverted); err != nil {
					__antithesis_instrumentation__.Notify(463704)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(463705)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(463706)
		}
	}
	__antithesis_instrumentation__.Notify(463639)

	nonIdxCols := 0
	for i := 0; i < len(desc.PublicColumns()) && func() bool {
		__antithesis_instrumentation__.Notify(463707)
		return nonIdxCols < maxNonIndexCols == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(463708)
		col := desc.PublicColumns()[i]

		if col.IsVirtual() {
			__antithesis_instrumentation__.Notify(463712)
			continue
		} else {
			__antithesis_instrumentation__.Notify(463713)
		}
		__antithesis_instrumentation__.Notify(463709)

		colList := []descpb.ColumnID{col.GetID()}

		if !trackStatsIfNotExists(colList) {
			__antithesis_instrumentation__.Notify(463714)
			continue
		} else {
			__antithesis_instrumentation__.Notify(463715)
		}
		__antithesis_instrumentation__.Notify(463710)

		maxHistBuckets := uint32(nonIndexColHistogramBuckets)
		if col.GetType().Family() == types.BoolFamily || func() bool {
			__antithesis_instrumentation__.Notify(463716)
			return col.GetType().Family() == types.EnumFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(463717)
			maxHistBuckets = stats.DefaultHistogramBuckets
		} else {
			__antithesis_instrumentation__.Notify(463718)
		}
		__antithesis_instrumentation__.Notify(463711)
		colStats = append(colStats, jobspb.CreateStatsDetails_ColStat{
			ColumnIDs:           colList,
			HasHistogram:        !colinfo.ColumnTypeIsInvertedIndexable(col.GetType()),
			HistogramMaxBuckets: maxHistBuckets,
		})
		nonIdxCols++
	}
	__antithesis_instrumentation__.Notify(463640)

	return colStats, nil
}

func makeColStatKey(cols []descpb.ColumnID) string {
	__antithesis_instrumentation__.Notify(463719)
	var colSet util.FastIntSet
	for _, c := range cols {
		__antithesis_instrumentation__.Notify(463721)
		colSet.Add(int(c))
	}
	__antithesis_instrumentation__.Notify(463720)
	return colSet.String()
}

type createStatsResumer struct {
	job     *jobs.Job
	tableID descpb.ID
}

var _ jobs.Resumer = &createStatsResumer{}

func (r *createStatsResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(463722)
	p := execCtx.(JobExecContext)
	details := r.job.Details().(jobspb.CreateStatsDetails)
	if details.Name == jobspb.AutoStatsName {
		__antithesis_instrumentation__.Notify(463726)

		if err := checkRunningJobs(ctx, r.job, p); err != nil {
			__antithesis_instrumentation__.Notify(463727)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463728)
		}
	} else {
		__antithesis_instrumentation__.Notify(463729)
	}
	__antithesis_instrumentation__.Notify(463723)

	r.tableID = details.Table.ID
	evalCtx := p.ExtendedEvalContext()

	dsp := p.DistSQLPlanner()
	if err := p.ExecCfg().DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(463730)

		evalCtx.Txn = txn

		if details.AsOf != nil {
			__antithesis_instrumentation__.Notify(463733)
			p.ExtendedEvalContext().AsOfSystemTime = &tree.AsOfSystemTime{Timestamp: *details.AsOf}
			p.ExtendedEvalContext().SetTxnTimestamp(details.AsOf.GoTime())
			if err := txn.SetFixedTimestamp(ctx, *details.AsOf); err != nil {
				__antithesis_instrumentation__.Notify(463734)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463735)
			}
		} else {
			__antithesis_instrumentation__.Notify(463736)
		}
		__antithesis_instrumentation__.Notify(463731)

		planCtx := dsp.NewPlanningCtx(ctx, evalCtx, nil, txn,
			DistributionTypeSystemTenantOnly)

		resultWriter := NewRowResultWriter(nil)
		if err := dsp.planAndRunCreateStats(
			ctx, evalCtx, planCtx, txn, r.job, resultWriter,
		); err != nil {
			__antithesis_instrumentation__.Notify(463737)

			if grpcutil.IsContextCanceled(err) {
				__antithesis_instrumentation__.Notify(463741)
				return jobs.MarkAsRetryJobError(err)
			} else {
				__antithesis_instrumentation__.Notify(463742)
			}
			__antithesis_instrumentation__.Notify(463738)

			txnForJobProgress := txn
			if details.AsOf != nil {
				__antithesis_instrumentation__.Notify(463743)
				txnForJobProgress = nil
			} else {
				__antithesis_instrumentation__.Notify(463744)
			}
			__antithesis_instrumentation__.Notify(463739)

			if jobErr := r.job.FractionProgressed(
				ctx, txnForJobProgress,
				func(ctx context.Context, _ jobspb.ProgressDetails) float32 {
					__antithesis_instrumentation__.Notify(463745)

					return 0
				},
			); jobErr != nil {
				__antithesis_instrumentation__.Notify(463746)
				return jobErr
			} else {
				__antithesis_instrumentation__.Notify(463747)
			}
			__antithesis_instrumentation__.Notify(463740)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463748)
		}
		__antithesis_instrumentation__.Notify(463732)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(463749)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463750)
	}
	__antithesis_instrumentation__.Notify(463724)

	if !createStatsPostEvents.Get(&evalCtx.Settings.SV) {
		__antithesis_instrumentation__.Notify(463751)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(463752)
	}
	__antithesis_instrumentation__.Notify(463725)

	return evalCtx.ExecCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(463753)
		return logEventInternalForSQLStatements(ctx,
			evalCtx.ExecCfg, txn,
			0,
			eventLogOptions{dst: LogEverywhere},
			eventpb.CommonSQLEventDetails{
				Statement:         redact.Sprint(details.Statement),
				Tag:               "CREATE STATISTICS",
				User:              evalCtx.SessionData().User().Normalized(),
				ApplicationName:   evalCtx.SessionData().ApplicationName,
				PlaceholderValues: []string{},
			},
			eventLogEntry{
				targetID: int32(details.Table.ID),
				event: &eventpb.CreateStatistics{
					TableName: details.FQTableName,
				},
			},
		)
	})
}

func checkRunningJobs(ctx context.Context, job *jobs.Job, p JobExecContext) error {
	__antithesis_instrumentation__.Notify(463754)
	jobID := jobspb.InvalidJobID
	if job != nil {
		__antithesis_instrumentation__.Notify(463759)
		jobID = job.ID()
	} else {
		__antithesis_instrumentation__.Notify(463760)
	}
	__antithesis_instrumentation__.Notify(463755)
	exists, err := jobs.RunningJobExists(ctx, jobID, p.ExecCfg().InternalExecutor, nil, func(payload *jobspb.Payload) bool {
		__antithesis_instrumentation__.Notify(463761)
		return payload.Type() == jobspb.TypeCreateStats || func() bool {
			__antithesis_instrumentation__.Notify(463762)
			return payload.Type() == jobspb.TypeAutoCreateStats == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(463756)

	if err != nil {
		__antithesis_instrumentation__.Notify(463763)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463764)
	}
	__antithesis_instrumentation__.Notify(463757)

	if exists {
		__antithesis_instrumentation__.Notify(463765)
		return stats.ConcurrentCreateStatsError
	} else {
		__antithesis_instrumentation__.Notify(463766)
	}
	__antithesis_instrumentation__.Notify(463758)

	return nil
}

func (r *createStatsResumer) OnFailOrCancel(context.Context, interface{}) error {
	__antithesis_instrumentation__.Notify(463767)
	return nil
}

func init() {
	createResumerFn := func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &createStatsResumer{job: job}
	}
	jobs.RegisterConstructor(jobspb.TypeCreateStats, createResumerFn)
	jobs.RegisterConstructor(jobspb.TypeAutoCreateStats, createResumerFn)
}
