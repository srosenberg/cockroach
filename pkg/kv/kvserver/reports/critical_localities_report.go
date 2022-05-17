package reports

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/errors"
)

const criticalLocalitiesReportID reportID = 2

type localityKey struct {
	ZoneKey
	locality LocalityRepr
}

type LocalityRepr string

type localityStatus struct {
	atRiskRanges int32
}

type LocalityReport map[localityKey]localityStatus

type replicationCriticalLocalitiesReportSaver struct {
	previousVersion     LocalityReport
	lastGenerated       time.Time
	lastUpdatedRowCount int
}

func makeReplicationCriticalLocalitiesReportSaver() replicationCriticalLocalitiesReportSaver {
	__antithesis_instrumentation__.Notify(121663)
	return replicationCriticalLocalitiesReportSaver{}
}

func (r *replicationCriticalLocalitiesReportSaver) LastUpdatedRowCount() int {
	__antithesis_instrumentation__.Notify(121664)
	return r.lastUpdatedRowCount
}

func (r LocalityReport) CountRangeAtRisk(zKey ZoneKey, loc LocalityRepr) {
	__antithesis_instrumentation__.Notify(121665)
	lKey := localityKey{
		ZoneKey:  zKey,
		locality: loc,
	}
	if _, ok := r[lKey]; !ok {
		__antithesis_instrumentation__.Notify(121667)
		r[lKey] = localityStatus{}
	} else {
		__antithesis_instrumentation__.Notify(121668)
	}
	__antithesis_instrumentation__.Notify(121666)
	lStat := r[lKey]
	lStat.atRiskRanges++
	r[lKey] = lStat
}

func (r *replicationCriticalLocalitiesReportSaver) loadPreviousVersion(
	ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(121669)

	if !r.lastGenerated.IsZero() {
		__antithesis_instrumentation__.Notify(121673)
		generated, err := getReportGenerationTime(ctx, criticalLocalitiesReportID, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121675)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121676)
		}
		__antithesis_instrumentation__.Notify(121674)

		if generated == r.lastGenerated {
			__antithesis_instrumentation__.Notify(121677)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(121678)
		}
	} else {
		__antithesis_instrumentation__.Notify(121679)
	}
	__antithesis_instrumentation__.Notify(121670)
	const prevViolations = "select zone_id, subzone_id, locality, at_risk_ranges " +
		"from system.replication_critical_localities"
	it, err := ex.QueryIterator(
		ctx, "get-previous-replication-critical-localities", txn, prevViolations,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(121680)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121681)
	}
	__antithesis_instrumentation__.Notify(121671)

	r.previousVersion = make(LocalityReport)
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(121682)
		row := it.Cur()
		key := localityKey{}
		key.ZoneID = (config.ObjectID)(*row[0].(*tree.DInt))
		key.SubzoneID = base.SubzoneID(*row[1].(*tree.DInt))
		key.locality = (LocalityRepr)(*row[2].(*tree.DString))
		r.previousVersion[key] = localityStatus{(int32)(*row[3].(*tree.DInt))}
	}
	__antithesis_instrumentation__.Notify(121672)
	return err
}

func (r *replicationCriticalLocalitiesReportSaver) updateTimestamp(
	ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn, reportTS time.Time,
) error {
	__antithesis_instrumentation__.Notify(121683)
	if !r.lastGenerated.IsZero() && func() bool {
		__antithesis_instrumentation__.Notify(121685)
		return reportTS == r.lastGenerated == true
	}() == true {
		__antithesis_instrumentation__.Notify(121686)
		return errors.Errorf(
			"The new time %s is the same as the time of the last update %s",
			reportTS.String(),
			r.lastGenerated.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(121687)
	}
	__antithesis_instrumentation__.Notify(121684)

	_, err := ex.Exec(
		ctx,
		"timestamp-upsert-replication-critical-localities",
		txn,
		"upsert into system.reports_meta(id, generated) values($1, $2)",
		criticalLocalitiesReportID,
		reportTS,
	)
	return err
}

func (r *replicationCriticalLocalitiesReportSaver) Save(
	ctx context.Context,
	report LocalityReport,
	reportTS time.Time,
	db *kv.DB,
	ex sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(121688)
	r.lastUpdatedRowCount = 0
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(121690)
		err := r.loadPreviousVersion(ctx, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121695)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121696)
		}
		__antithesis_instrumentation__.Notify(121691)

		err = r.updateTimestamp(ctx, ex, txn, reportTS)
		if err != nil {
			__antithesis_instrumentation__.Notify(121697)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121698)
		}
		__antithesis_instrumentation__.Notify(121692)

		for key, status := range report {
			__antithesis_instrumentation__.Notify(121699)
			if err := r.upsertLocality(
				ctx, reportTS, txn, key, status, db, ex,
			); err != nil {
				__antithesis_instrumentation__.Notify(121700)
				return err
			} else {
				__antithesis_instrumentation__.Notify(121701)
			}
		}
		__antithesis_instrumentation__.Notify(121693)

		for key := range r.previousVersion {
			__antithesis_instrumentation__.Notify(121702)
			if _, ok := report[key]; !ok {
				__antithesis_instrumentation__.Notify(121703)
				_, err := ex.Exec(
					ctx,
					"delete-old-replication-critical-localities",
					txn,
					"delete from system.replication_critical_localities "+
						"where zone_id = $1 and subzone_id = $2 and locality = $3",
					key.ZoneID,
					key.SubzoneID,
					key.locality,
				)

				if err != nil {
					__antithesis_instrumentation__.Notify(121705)
					return err
				} else {
					__antithesis_instrumentation__.Notify(121706)
				}
				__antithesis_instrumentation__.Notify(121704)
				r.lastUpdatedRowCount++
			} else {
				__antithesis_instrumentation__.Notify(121707)
			}
		}
		__antithesis_instrumentation__.Notify(121694)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(121708)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121709)
	}
	__antithesis_instrumentation__.Notify(121689)

	r.lastGenerated = reportTS
	r.previousVersion = report

	return nil
}

func (r *replicationCriticalLocalitiesReportSaver) upsertLocality(
	ctx context.Context,
	reportTS time.Time,
	txn *kv.Txn,
	key localityKey,
	status localityStatus,
	db *kv.DB,
	ex sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(121710)
	var err error
	previousStatus, hasOldVersion := r.previousVersion[key]
	if hasOldVersion && func() bool {
		__antithesis_instrumentation__.Notify(121713)
		return previousStatus.atRiskRanges == status.atRiskRanges == true
	}() == true {
		__antithesis_instrumentation__.Notify(121714)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(121715)
	}
	__antithesis_instrumentation__.Notify(121711)

	_, err = ex.Exec(
		ctx, "upsert-replication-critical-localities", txn,
		"upsert into system.replication_critical_localities(report_id, zone_id, subzone_id, "+
			"locality, at_risk_ranges) values($1, $2, $3, $4, $5)",
		criticalLocalitiesReportID,
		key.ZoneID, key.SubzoneID, key.locality, status.atRiskRanges,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(121716)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121717)
	}
	__antithesis_instrumentation__.Notify(121712)

	r.lastUpdatedRowCount++
	return nil
}

type criticalLocalitiesVisitor struct {
	allLocalities map[roachpb.NodeID]map[string]roachpb.Locality
	cfg           *config.SystemConfig

	storeResolver StoreResolver
	nodeChecker   nodeChecker

	report   LocalityReport
	visitErr bool

	prevZoneKey ZoneKey
}

var _ rangeVisitor = &criticalLocalitiesVisitor{}

func makeCriticalLocalitiesVisitor(
	ctx context.Context,
	nodeLocalities map[roachpb.NodeID]roachpb.Locality,
	cfg *config.SystemConfig,
	storeResolver StoreResolver,
	nodeChecker nodeChecker,
) criticalLocalitiesVisitor {
	__antithesis_instrumentation__.Notify(121718)
	allLocalities := expandLocalities(nodeLocalities)
	v := criticalLocalitiesVisitor{
		allLocalities: allLocalities,
		cfg:           cfg,
		storeResolver: storeResolver,
		nodeChecker:   nodeChecker,
	}
	v.reset(ctx)
	return v
}

func expandLocalities(
	nodeLocalities map[roachpb.NodeID]roachpb.Locality,
) map[roachpb.NodeID]map[string]roachpb.Locality {
	__antithesis_instrumentation__.Notify(121719)
	res := make(map[roachpb.NodeID]map[string]roachpb.Locality)
	for nid, loc := range nodeLocalities {
		__antithesis_instrumentation__.Notify(121721)
		if len(loc.Tiers) == 0 {
			__antithesis_instrumentation__.Notify(121723)
			res[nid] = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(121724)
		}
		__antithesis_instrumentation__.Notify(121722)
		res[nid] = make(map[string]roachpb.Locality, len(loc.Tiers))
		for i := range loc.Tiers {
			__antithesis_instrumentation__.Notify(121725)
			partialLoc := roachpb.Locality{Tiers: make([]roachpb.Tier, i+1)}
			copy(partialLoc.Tiers, loc.Tiers[:i+1])
			res[nid][partialLoc.String()] = partialLoc
		}
	}
	__antithesis_instrumentation__.Notify(121720)
	return res
}

func (v *criticalLocalitiesVisitor) failed() bool {
	__antithesis_instrumentation__.Notify(121726)
	return v.visitErr
}

func (v *criticalLocalitiesVisitor) Report() LocalityReport {
	__antithesis_instrumentation__.Notify(121727)
	return v.report
}

func (v *criticalLocalitiesVisitor) reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(121728)
	*v = criticalLocalitiesVisitor{
		allLocalities: v.allLocalities,
		cfg:           v.cfg,
		storeResolver: v.storeResolver,
		nodeChecker:   v.nodeChecker,
		report:        make(LocalityReport, len(v.report)),
	}
}

func (v *criticalLocalitiesVisitor) visitNewZone(
	ctx context.Context, r *roachpb.RangeDescriptor,
) (retErr error) {
	__antithesis_instrumentation__.Notify(121729)

	defer func() {
		__antithesis_instrumentation__.Notify(121734)
		v.visitErr = retErr != nil
	}()
	__antithesis_instrumentation__.Notify(121730)

	var zKey ZoneKey
	found, err := visitZones(ctx, r, v.cfg, ignoreSubzonePlaceholders,
		func(_ context.Context, zone *zonepb.ZoneConfig, key ZoneKey) bool {
			__antithesis_instrumentation__.Notify(121735)
			if !zoneChangesReplication(zone) {
				__antithesis_instrumentation__.Notify(121737)
				return false
			} else {
				__antithesis_instrumentation__.Notify(121738)
			}
			__antithesis_instrumentation__.Notify(121736)
			zKey = key
			return true
		})
	__antithesis_instrumentation__.Notify(121731)
	if err != nil {
		__antithesis_instrumentation__.Notify(121739)
		return errors.NewAssertionErrorWithWrappedErrf(err, "unexpected error visiting zones")
	} else {
		__antithesis_instrumentation__.Notify(121740)
	}
	__antithesis_instrumentation__.Notify(121732)
	if !found {
		__antithesis_instrumentation__.Notify(121741)
		return errors.AssertionFailedf("no suitable zone config found for range: %s", r)
	} else {
		__antithesis_instrumentation__.Notify(121742)
	}
	__antithesis_instrumentation__.Notify(121733)
	v.prevZoneKey = zKey

	v.countRange(ctx, zKey, r)
	return nil
}

func (v *criticalLocalitiesVisitor) visitSameZone(ctx context.Context, r *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(121743)
	v.countRange(ctx, v.prevZoneKey, r)
}

func (v *criticalLocalitiesVisitor) countRange(
	ctx context.Context, zoneKey ZoneKey, r *roachpb.RangeDescriptor,
) {
	__antithesis_instrumentation__.Notify(121744)
	stores := v.storeResolver(r)

	dedupLocal := make(map[string]roachpb.Locality)
	for _, rep := range r.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(121746)
		for s, loc := range v.allLocalities[rep.NodeID] {
			__antithesis_instrumentation__.Notify(121747)
			if _, ok := dedupLocal[s]; ok {
				__antithesis_instrumentation__.Notify(121749)
				continue
			} else {
				__antithesis_instrumentation__.Notify(121750)
			}
			__antithesis_instrumentation__.Notify(121748)
			dedupLocal[s] = loc
		}
	}
	__antithesis_instrumentation__.Notify(121745)

	for _, loc := range dedupLocal {
		__antithesis_instrumentation__.Notify(121751)
		processLocalityForRange(ctx, r, zoneKey, loc, v.nodeChecker, stores, v.report)
	}
}

func processLocalityForRange(
	ctx context.Context,
	r *roachpb.RangeDescriptor,
	zoneKey ZoneKey,
	loc roachpb.Locality,
	nodeChecker nodeChecker,
	storeDescs []roachpb.StoreDescriptor,
	rep LocalityReport,
) {
	__antithesis_instrumentation__.Notify(121752)

	inLoc := func(other roachpb.Locality) bool {
		__antithesis_instrumentation__.Notify(121755)

		i := 0
		for i < len(loc.Tiers) && func() bool {
			__antithesis_instrumentation__.Notify(121757)
			return i < len(other.Tiers) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(121758)
			return loc.Tiers[i] == other.Tiers[i] == true
		}() == true {
			__antithesis_instrumentation__.Notify(121759)
			i++
		}
		__antithesis_instrumentation__.Notify(121756)

		return i == len(loc.Tiers)
	}
	__antithesis_instrumentation__.Notify(121753)

	unavailableWithoutLoc := !r.Replicas().CanMakeProgress(func(rDesc roachpb.ReplicaDescriptor) bool {
		__antithesis_instrumentation__.Notify(121760)
		alive := nodeChecker(rDesc.NodeID)

		var replicaInLoc bool
		var sDesc roachpb.StoreDescriptor
		for _, sd := range storeDescs {
			__antithesis_instrumentation__.Notify(121763)
			if sd.StoreID == rDesc.StoreID {
				__antithesis_instrumentation__.Notify(121764)
				sDesc = sd
				break
			} else {
				__antithesis_instrumentation__.Notify(121765)
			}
		}
		__antithesis_instrumentation__.Notify(121761)
		if sDesc.StoreID == 0 {
			__antithesis_instrumentation__.Notify(121766)
			replicaInLoc = false
		} else {
			__antithesis_instrumentation__.Notify(121767)
			replicaInLoc = inLoc(sDesc.Node.Locality)
		}
		__antithesis_instrumentation__.Notify(121762)
		return alive && func() bool {
			__antithesis_instrumentation__.Notify(121768)
			return !replicaInLoc == true
		}() == true
	})
	__antithesis_instrumentation__.Notify(121754)

	if unavailableWithoutLoc {
		__antithesis_instrumentation__.Notify(121769)
		rep.CountRangeAtRisk(zoneKey, LocalityRepr(loc.String()))
	} else {
		__antithesis_instrumentation__.Notify(121770)
	}
}
