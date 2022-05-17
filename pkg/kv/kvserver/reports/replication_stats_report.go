package reports

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const replicationStatsReportID reportID = 3

type RangeReport map[ZoneKey]zoneRangeStatus

type zoneRangeStatus struct {
	numRanges       int32
	unavailable     int32
	underReplicated int32
	overReplicated  int32
}

type replicationStatsReportSaver struct {
	previousVersion     RangeReport
	lastGenerated       time.Time
	lastUpdatedRowCount int
}

func makeReplicationStatsReportSaver() replicationStatsReportSaver {
	__antithesis_instrumentation__.Notify(121771)
	return replicationStatsReportSaver{}
}

func (r *replicationStatsReportSaver) LastUpdatedRowCount() int {
	__antithesis_instrumentation__.Notify(121772)
	return r.lastUpdatedRowCount
}

func (r RangeReport) EnsureEntry(zKey ZoneKey) {
	__antithesis_instrumentation__.Notify(121773)
	if _, ok := r[zKey]; !ok {
		__antithesis_instrumentation__.Notify(121774)
		r[zKey] = zoneRangeStatus{}
	} else {
		__antithesis_instrumentation__.Notify(121775)
	}
}

func (r RangeReport) CountRange(zKey ZoneKey, status roachpb.RangeStatusReport) {
	__antithesis_instrumentation__.Notify(121776)
	r.EnsureEntry(zKey)
	rStat := r[zKey]
	rStat.numRanges++
	if !status.Available {
		__antithesis_instrumentation__.Notify(121780)
		rStat.unavailable++
	} else {
		__antithesis_instrumentation__.Notify(121781)
	}
	__antithesis_instrumentation__.Notify(121777)
	if status.UnderReplicated {
		__antithesis_instrumentation__.Notify(121782)
		rStat.underReplicated++
	} else {
		__antithesis_instrumentation__.Notify(121783)
	}
	__antithesis_instrumentation__.Notify(121778)
	if status.OverReplicated {
		__antithesis_instrumentation__.Notify(121784)
		rStat.overReplicated++
	} else {
		__antithesis_instrumentation__.Notify(121785)
	}
	__antithesis_instrumentation__.Notify(121779)
	r[zKey] = rStat
}

func (r *replicationStatsReportSaver) loadPreviousVersion(
	ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(121786)

	if !r.lastGenerated.IsZero() {
		__antithesis_instrumentation__.Notify(121790)
		generated, err := getReportGenerationTime(ctx, replicationStatsReportID, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121792)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121793)
		}
		__antithesis_instrumentation__.Notify(121791)

		if generated == r.lastGenerated {
			__antithesis_instrumentation__.Notify(121794)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(121795)
		}
	} else {
		__antithesis_instrumentation__.Notify(121796)
	}
	__antithesis_instrumentation__.Notify(121787)
	const prevViolations = "select zone_id, subzone_id, total_ranges, " +
		"unavailable_ranges, under_replicated_ranges, over_replicated_ranges " +
		"from system.replication_stats"
	it, err := ex.QueryIterator(
		ctx, "get-previous-replication-stats", txn, prevViolations,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(121797)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121798)
	}
	__antithesis_instrumentation__.Notify(121788)

	r.previousVersion = make(RangeReport)
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(121799)
		row := it.Cur()
		key := ZoneKey{}
		key.ZoneID = (config.ObjectID)(*row[0].(*tree.DInt))
		key.SubzoneID = base.SubzoneID(*row[1].(*tree.DInt))
		r.previousVersion[key] = zoneRangeStatus{
			(int32)(*row[2].(*tree.DInt)),
			(int32)(*row[3].(*tree.DInt)),
			(int32)(*row[4].(*tree.DInt)),
			(int32)(*row[5].(*tree.DInt)),
		}
	}
	__antithesis_instrumentation__.Notify(121789)
	return err
}

func (r *replicationStatsReportSaver) updateTimestamp(
	ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn, reportTS time.Time,
) error {
	__antithesis_instrumentation__.Notify(121800)
	if !r.lastGenerated.IsZero() && func() bool {
		__antithesis_instrumentation__.Notify(121802)
		return reportTS == r.lastGenerated == true
	}() == true {
		__antithesis_instrumentation__.Notify(121803)
		return errors.Errorf(
			"The new time %s is the same as the time of the last update %s",
			reportTS.String(),
			r.lastGenerated.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(121804)
	}
	__antithesis_instrumentation__.Notify(121801)

	_, err := ex.Exec(
		ctx,
		"timestamp-upsert-replication-stats",
		txn,
		"upsert into system.reports_meta(id, generated) values($1, $2)",
		replicationStatsReportID,
		reportTS,
	)
	return err
}

func (r *replicationStatsReportSaver) Save(
	ctx context.Context,
	report RangeReport,
	reportTS time.Time,
	db *kv.DB,
	ex sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(121805)
	r.lastUpdatedRowCount = 0
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(121807)
		err := r.loadPreviousVersion(ctx, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121812)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121813)
		}
		__antithesis_instrumentation__.Notify(121808)

		err = r.updateTimestamp(ctx, ex, txn, reportTS)
		if err != nil {
			__antithesis_instrumentation__.Notify(121814)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121815)
		}
		__antithesis_instrumentation__.Notify(121809)

		for key, status := range report {
			__antithesis_instrumentation__.Notify(121816)
			if err := r.upsertStats(ctx, txn, key, status, ex); err != nil {
				__antithesis_instrumentation__.Notify(121817)
				return err
			} else {
				__antithesis_instrumentation__.Notify(121818)
			}
		}
		__antithesis_instrumentation__.Notify(121810)

		for key := range r.previousVersion {
			__antithesis_instrumentation__.Notify(121819)
			if _, ok := report[key]; !ok {
				__antithesis_instrumentation__.Notify(121820)
				_, err := ex.Exec(
					ctx,
					"delete-old-replication-stats",
					txn,
					"delete from system.replication_stats "+
						"where zone_id = $1 and subzone_id = $2",
					key.ZoneID,
					key.SubzoneID,
				)

				if err != nil {
					__antithesis_instrumentation__.Notify(121822)
					return err
				} else {
					__antithesis_instrumentation__.Notify(121823)
				}
				__antithesis_instrumentation__.Notify(121821)
				r.lastUpdatedRowCount++
			} else {
				__antithesis_instrumentation__.Notify(121824)
			}
		}
		__antithesis_instrumentation__.Notify(121811)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(121825)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121826)
	}
	__antithesis_instrumentation__.Notify(121806)

	r.lastGenerated = reportTS
	r.previousVersion = report

	return nil
}

func (r *replicationStatsReportSaver) upsertStats(
	ctx context.Context, txn *kv.Txn, key ZoneKey, stats zoneRangeStatus, ex sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(121827)
	var err error
	previousStats, hasOldVersion := r.previousVersion[key]
	if hasOldVersion && func() bool {
		__antithesis_instrumentation__.Notify(121830)
		return previousStats == stats == true
	}() == true {
		__antithesis_instrumentation__.Notify(121831)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(121832)
	}
	__antithesis_instrumentation__.Notify(121828)

	_, err = ex.Exec(
		ctx, "upsert-replication-stats", txn,
		"upsert into system.replication_stats(report_id, zone_id, subzone_id, "+
			"total_ranges, unavailable_ranges, under_replicated_ranges, "+
			"over_replicated_ranges) values($1, $2, $3, $4, $5, $6, $7)",
		replicationStatsReportID,
		key.ZoneID, key.SubzoneID, stats.numRanges, stats.unavailable,
		stats.underReplicated, stats.overReplicated,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(121833)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121834)
	}
	__antithesis_instrumentation__.Notify(121829)

	r.lastUpdatedRowCount++
	return nil
}

type replicationStatsVisitor struct {
	cfg         *config.SystemConfig
	nodeChecker nodeChecker

	report   RangeReport
	visitErr bool

	prevZoneKey   ZoneKey
	prevNumVoters int
}

var _ rangeVisitor = &replicationStatsVisitor{}

func makeReplicationStatsVisitor(
	ctx context.Context, cfg *config.SystemConfig, nodeChecker nodeChecker,
) replicationStatsVisitor {
	__antithesis_instrumentation__.Notify(121835)
	v := replicationStatsVisitor{
		cfg:         cfg,
		nodeChecker: nodeChecker,
		report:      make(RangeReport),
	}
	v.reset(ctx)
	return v
}

func (v *replicationStatsVisitor) failed() bool {
	__antithesis_instrumentation__.Notify(121836)
	return v.visitErr
}

func (v *replicationStatsVisitor) Report() RangeReport {
	__antithesis_instrumentation__.Notify(121837)
	return v.report
}

func (v *replicationStatsVisitor) reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(121838)
	*v = replicationStatsVisitor{
		cfg:           v.cfg,
		nodeChecker:   v.nodeChecker,
		prevNumVoters: -1,
		report:        make(RangeReport, len(v.report)),
	}

	maxObjectID, err := v.cfg.GetLargestObjectID(
		0, keys.PseudoTableIDs,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(121840)
		log.Fatalf(ctx, "unexpected failure to compute max object id: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(121841)
	}
	__antithesis_instrumentation__.Notify(121839)
	for i := config.ObjectID(1); i <= maxObjectID; i++ {
		__antithesis_instrumentation__.Notify(121842)
		zone, err := getZoneByID(i, v.cfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(121845)
			log.Fatalf(ctx, "unexpected failure to compute max object id: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(121846)
		}
		__antithesis_instrumentation__.Notify(121843)
		if zone == nil {
			__antithesis_instrumentation__.Notify(121847)
			continue
		} else {
			__antithesis_instrumentation__.Notify(121848)
		}
		__antithesis_instrumentation__.Notify(121844)
		v.ensureEntries(MakeZoneKey(i, NoSubzone), zone)
	}
}

func (v *replicationStatsVisitor) ensureEntries(key ZoneKey, zone *zonepb.ZoneConfig) {
	__antithesis_instrumentation__.Notify(121849)
	if zoneChangesReplication(zone) {
		__antithesis_instrumentation__.Notify(121851)
		v.report.EnsureEntry(key)
	} else {
		__antithesis_instrumentation__.Notify(121852)
	}
	__antithesis_instrumentation__.Notify(121850)
	for i, sz := range zone.Subzones {
		__antithesis_instrumentation__.Notify(121853)
		v.ensureEntries(MakeZoneKey(key.ZoneID, base.SubzoneIDFromIndex(i)), &sz.Config)
	}
}

func (v *replicationStatsVisitor) visitNewZone(
	ctx context.Context, r *roachpb.RangeDescriptor,
) (retErr error) {
	__antithesis_instrumentation__.Notify(121854)

	defer func() {
		__antithesis_instrumentation__.Notify(121859)
		v.visitErr = retErr != nil
	}()
	__antithesis_instrumentation__.Notify(121855)
	var zKey ZoneKey
	var zConfig *zonepb.ZoneConfig

	var desiredNumVoters int

	_, err := visitZones(ctx, r, v.cfg, ignoreSubzonePlaceholders,
		func(_ context.Context, zone *zonepb.ZoneConfig, key ZoneKey) bool {
			__antithesis_instrumentation__.Notify(121860)
			if zConfig == nil {
				__antithesis_instrumentation__.Notify(121864)
				if !zoneChangesReplication(zone) {
					__antithesis_instrumentation__.Notify(121868)
					return false
				} else {
					__antithesis_instrumentation__.Notify(121869)
				}
				__antithesis_instrumentation__.Notify(121865)
				zKey = key
				zConfig = zone
				if zone.NumVoters != nil {
					__antithesis_instrumentation__.Notify(121870)
					desiredNumVoters = int(*zone.NumVoters)
					return true
				} else {
					__antithesis_instrumentation__.Notify(121871)
				}
				__antithesis_instrumentation__.Notify(121866)
				if zone.NumReplicas != nil && func() bool {
					__antithesis_instrumentation__.Notify(121872)
					return desiredNumVoters == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(121873)
					desiredNumVoters = int(*zone.NumReplicas)
				} else {
					__antithesis_instrumentation__.Notify(121874)
				}
				__antithesis_instrumentation__.Notify(121867)

				return false
			} else {
				__antithesis_instrumentation__.Notify(121875)
			}
			__antithesis_instrumentation__.Notify(121861)
			if zone.NumVoters != nil {
				__antithesis_instrumentation__.Notify(121876)
				desiredNumVoters = int(*zone.NumVoters)
				return true
			} else {
				__antithesis_instrumentation__.Notify(121877)
			}
			__antithesis_instrumentation__.Notify(121862)
			if zone.NumReplicas != nil && func() bool {
				__antithesis_instrumentation__.Notify(121878)
				return desiredNumVoters == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(121879)
				desiredNumVoters = int(*zone.NumReplicas)
			} else {
				__antithesis_instrumentation__.Notify(121880)
			}
			__antithesis_instrumentation__.Notify(121863)

			return false
		})
	__antithesis_instrumentation__.Notify(121856)
	if err != nil {
		__antithesis_instrumentation__.Notify(121881)
		return errors.NewAssertionErrorWithWrappedErrf(err, "unexpected error visiting zones for range %s", r)
	} else {
		__antithesis_instrumentation__.Notify(121882)
	}
	__antithesis_instrumentation__.Notify(121857)
	if desiredNumVoters == 0 {
		__antithesis_instrumentation__.Notify(121883)
		return errors.AssertionFailedf(
			"no zone config with replication attributes found for range: %s", r)
	} else {
		__antithesis_instrumentation__.Notify(121884)
	}
	__antithesis_instrumentation__.Notify(121858)
	v.prevZoneKey = zKey
	v.prevNumVoters = desiredNumVoters

	v.countRange(ctx, zKey, desiredNumVoters, r)
	return nil
}

func (v *replicationStatsVisitor) visitSameZone(ctx context.Context, r *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(121885)
	v.countRange(ctx, v.prevZoneKey, v.prevNumVoters, r)
}

func (v *replicationStatsVisitor) countRange(
	ctx context.Context, key ZoneKey, replicationFactor int, r *roachpb.RangeDescriptor,
) {
	__antithesis_instrumentation__.Notify(121886)
	status := r.Replicas().ReplicationStatus(func(rDesc roachpb.ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(121888)
		return v.nodeChecker(rDesc.NodeID)
	}, replicationFactor)
	__antithesis_instrumentation__.Notify(121887)

	v.report.CountRange(key, status)
}

func zoneChangesReplication(zone *zonepb.ZoneConfig) bool {
	__antithesis_instrumentation__.Notify(121889)
	return (zone.NumReplicas != nil && func() bool {
		__antithesis_instrumentation__.Notify(121890)
		return *zone.NumReplicas != 0 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(121891)
		return (zone.NumVoters != nil && func() bool {
			__antithesis_instrumentation__.Notify(121892)
			return *zone.NumVoters != 0 == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(121893)
		return zone.Constraints != nil == true
	}() == true
}
