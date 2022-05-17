package reports

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
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

const replicationConstraintsReportID reportID = 1

type ConstraintReport map[ConstraintStatusKey]ConstraintStatus

type replicationConstraintStatsReportSaver struct {
	previousVersion     ConstraintReport
	lastGenerated       time.Time
	lastUpdatedRowCount int
}

func makeReplicationConstraintStatusReportSaver() replicationConstraintStatsReportSaver {
	__antithesis_instrumentation__.Notify(121534)
	return replicationConstraintStatsReportSaver{}
}

func (r *replicationConstraintStatsReportSaver) LastUpdatedRowCount() int {
	__antithesis_instrumentation__.Notify(121535)
	return r.lastUpdatedRowCount
}

type ConstraintStatus struct {
	FailRangeCount int
}

type ConstraintType string

const (
	Constraint ConstraintType = "constraint"
)

func (t ConstraintType) Less(other ConstraintType) bool {
	__antithesis_instrumentation__.Notify(121536)
	return -1 == strings.Compare(string(t), string(other))
}

type ConstraintRepr string

func (c ConstraintRepr) Less(other ConstraintRepr) bool {
	__antithesis_instrumentation__.Notify(121537)
	return -1 == strings.Compare(string(c), string(other))
}

type ConstraintStatusKey struct {
	ZoneKey
	ViolationType ConstraintType
	Constraint    ConstraintRepr
}

func (k ConstraintStatusKey) String() string {
	__antithesis_instrumentation__.Notify(121538)
	return fmt.Sprintf("zone:%s type:%s constraint:%s", k.ZoneKey, k.ViolationType, k.Constraint)
}

func (k ConstraintStatusKey) Less(other ConstraintStatusKey) bool {
	__antithesis_instrumentation__.Notify(121539)
	if k.ZoneKey.Less(other.ZoneKey) {
		__antithesis_instrumentation__.Notify(121544)
		return true
	} else {
		__antithesis_instrumentation__.Notify(121545)
	}
	__antithesis_instrumentation__.Notify(121540)
	if other.ZoneKey.Less(k.ZoneKey) {
		__antithesis_instrumentation__.Notify(121546)
		return false
	} else {
		__antithesis_instrumentation__.Notify(121547)
	}
	__antithesis_instrumentation__.Notify(121541)
	if k.ViolationType.Less(other.ViolationType) {
		__antithesis_instrumentation__.Notify(121548)
		return true
	} else {
		__antithesis_instrumentation__.Notify(121549)
	}
	__antithesis_instrumentation__.Notify(121542)
	if other.ViolationType.Less(k.ViolationType) {
		__antithesis_instrumentation__.Notify(121550)
		return true
	} else {
		__antithesis_instrumentation__.Notify(121551)
	}
	__antithesis_instrumentation__.Notify(121543)
	return k.Constraint.Less(other.Constraint)
}

func (r ConstraintReport) AddViolation(z ZoneKey, t ConstraintType, c ConstraintRepr) {
	__antithesis_instrumentation__.Notify(121552)
	k := ConstraintStatusKey{
		ZoneKey:       z,
		ViolationType: t,
		Constraint:    c,
	}
	if _, ok := r[k]; !ok {
		__antithesis_instrumentation__.Notify(121554)
		r[k] = ConstraintStatus{}
	} else {
		__antithesis_instrumentation__.Notify(121555)
	}
	__antithesis_instrumentation__.Notify(121553)
	cRep := r[k]
	cRep.FailRangeCount++
	r[k] = cRep
}

func (r ConstraintReport) ensureEntry(z ZoneKey, t ConstraintType, c ConstraintRepr) {
	__antithesis_instrumentation__.Notify(121556)
	k := ConstraintStatusKey{
		ZoneKey:       z,
		ViolationType: t,
		Constraint:    c,
	}
	if _, ok := r[k]; !ok {
		__antithesis_instrumentation__.Notify(121557)
		r[k] = ConstraintStatus{}
	} else {
		__antithesis_instrumentation__.Notify(121558)
	}
}

func (r ConstraintReport) ensureEntries(key ZoneKey, zone *zonepb.ZoneConfig) {
	__antithesis_instrumentation__.Notify(121559)
	for _, conjunction := range zone.Constraints {
		__antithesis_instrumentation__.Notify(121561)
		r.ensureEntry(key, Constraint, ConstraintRepr(conjunction.String()))
	}
	__antithesis_instrumentation__.Notify(121560)
	for i, sz := range zone.Subzones {
		__antithesis_instrumentation__.Notify(121562)
		szKey := ZoneKey{ZoneID: key.ZoneID, SubzoneID: base.SubzoneIDFromIndex(i)}
		r.ensureEntries(szKey, &sz.Config)
	}
}

func (r *replicationConstraintStatsReportSaver) loadPreviousVersion(
	ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(121563)

	if !r.lastGenerated.IsZero() {
		__antithesis_instrumentation__.Notify(121567)
		generated, err := getReportGenerationTime(ctx, replicationConstraintsReportID, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121569)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121570)
		}
		__antithesis_instrumentation__.Notify(121568)

		if generated == r.lastGenerated {
			__antithesis_instrumentation__.Notify(121571)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(121572)
		}
	} else {
		__antithesis_instrumentation__.Notify(121573)
	}
	__antithesis_instrumentation__.Notify(121564)
	const prevViolations = "select zone_id, subzone_id, type, config, " +
		"violating_ranges from system.replication_constraint_stats"
	it, err := ex.QueryIterator(
		ctx, "get-previous-replication-constraint-stats", txn, prevViolations,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(121574)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121575)
	}
	__antithesis_instrumentation__.Notify(121565)

	r.previousVersion = make(ConstraintReport)
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(121576)
		row := it.Cur()
		key := ConstraintStatusKey{}
		key.ZoneID = (config.ObjectID)(*row[0].(*tree.DInt))
		key.SubzoneID = base.SubzoneID((*row[1].(*tree.DInt)))
		key.ViolationType = (ConstraintType)(*row[2].(*tree.DString))
		key.Constraint = (ConstraintRepr)(*row[3].(*tree.DString))
		r.previousVersion[key] = ConstraintStatus{(int)(*row[4].(*tree.DInt))}
	}
	__antithesis_instrumentation__.Notify(121566)
	return err
}

func (r *replicationConstraintStatsReportSaver) updateTimestamp(
	ctx context.Context, ex sqlutil.InternalExecutor, txn *kv.Txn, reportTS time.Time,
) error {
	__antithesis_instrumentation__.Notify(121577)
	if !r.lastGenerated.IsZero() && func() bool {
		__antithesis_instrumentation__.Notify(121579)
		return reportTS == r.lastGenerated == true
	}() == true {
		__antithesis_instrumentation__.Notify(121580)
		return errors.Errorf(
			"The new time %s is the same as the time of the last update %s",
			reportTS.String(),
			r.lastGenerated.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(121581)
	}
	__antithesis_instrumentation__.Notify(121578)

	_, err := ex.Exec(
		ctx,
		"timestamp-upsert-replication-constraint-stats",
		txn,
		"upsert into system.reports_meta(id, generated) values($1, $2)",
		replicationConstraintsReportID,
		reportTS,
	)
	return err
}

func (r *replicationConstraintStatsReportSaver) Save(
	ctx context.Context,
	report ConstraintReport,
	reportTS time.Time,
	db *kv.DB,
	ex sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(121582)
	r.lastUpdatedRowCount = 0
	if err := db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(121584)
		err := r.loadPreviousVersion(ctx, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(121589)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121590)
		}
		__antithesis_instrumentation__.Notify(121585)

		err = r.updateTimestamp(ctx, ex, txn, reportTS)
		if err != nil {
			__antithesis_instrumentation__.Notify(121591)
			return err
		} else {
			__antithesis_instrumentation__.Notify(121592)
		}
		__antithesis_instrumentation__.Notify(121586)

		for k, zoneCons := range report {
			__antithesis_instrumentation__.Notify(121593)
			if err := r.upsertConstraintStatus(
				ctx, reportTS, txn, k, zoneCons.FailRangeCount, db, ex,
			); err != nil {
				__antithesis_instrumentation__.Notify(121594)
				return err
			} else {
				__antithesis_instrumentation__.Notify(121595)
			}
		}
		__antithesis_instrumentation__.Notify(121587)

		for key := range r.previousVersion {
			__antithesis_instrumentation__.Notify(121596)
			if _, ok := report[key]; !ok {
				__antithesis_instrumentation__.Notify(121597)
				_, err := ex.Exec(
					ctx,
					"delete-old-replication-constraint-stats",
					txn,
					"delete from system.replication_constraint_stats "+
						"where zone_id = $1 and subzone_id = $2 and type = $3 and config = $4",
					key.ZoneID,
					key.SubzoneID,
					key.ViolationType,
					key.Constraint,
				)

				if err != nil {
					__antithesis_instrumentation__.Notify(121599)
					return err
				} else {
					__antithesis_instrumentation__.Notify(121600)
				}
				__antithesis_instrumentation__.Notify(121598)
				r.lastUpdatedRowCount++
			} else {
				__antithesis_instrumentation__.Notify(121601)
			}
		}
		__antithesis_instrumentation__.Notify(121588)

		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(121602)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121603)
	}
	__antithesis_instrumentation__.Notify(121583)

	r.lastGenerated = reportTS
	r.previousVersion = report

	return nil
}

func (r *replicationConstraintStatsReportSaver) upsertConstraintStatus(
	ctx context.Context,
	reportTS time.Time,
	txn *kv.Txn,
	key ConstraintStatusKey,
	violationCount int,
	db *kv.DB,
	ex sqlutil.InternalExecutor,
) error {
	__antithesis_instrumentation__.Notify(121604)
	var err error
	previousStatus, hasOldVersion := r.previousVersion[key]
	if hasOldVersion && func() bool {
		__antithesis_instrumentation__.Notify(121607)
		return previousStatus.FailRangeCount == violationCount == true
	}() == true {
		__antithesis_instrumentation__.Notify(121608)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(121609)
		if violationCount != 0 {
			__antithesis_instrumentation__.Notify(121610)
			if previousStatus.FailRangeCount != 0 {
				__antithesis_instrumentation__.Notify(121611)

				_, err = ex.Exec(
					ctx, "upsert-replication-constraint-stat", txn,
					"upsert into system.replication_constraint_stats(report_id, zone_id, subzone_id, type, "+
						"config, violating_ranges) values($1, $2, $3, $4, $5, $6)",
					replicationConstraintsReportID,
					key.ZoneID, key.SubzoneID, key.ViolationType, key.Constraint, violationCount,
				)
			} else {
				__antithesis_instrumentation__.Notify(121612)
				if previousStatus.FailRangeCount == 0 {
					__antithesis_instrumentation__.Notify(121613)

					_, err = ex.Exec(
						ctx, "upsert-replication-constraint-stat", txn,
						"upsert into system.replication_constraint_stats(report_id, zone_id, subzone_id, type, "+
							"config, violating_ranges, violation_start) values($1, $2, $3, $4, $5, $6, $7)",
						replicationConstraintsReportID,
						key.ZoneID, key.SubzoneID, key.ViolationType, key.Constraint, violationCount, reportTS,
					)
				} else {
					__antithesis_instrumentation__.Notify(121614)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(121615)

			_, err = ex.Exec(
				ctx, "upsert-replication-constraint-stat", txn,
				"upsert into system.replication_constraint_stats(report_id, zone_id, subzone_id, type, config, "+
					"violating_ranges, violation_start) values($1, $2, $3, $4, $5, $6, null)",
				replicationConstraintsReportID,
				key.ZoneID, key.SubzoneID, key.ViolationType, key.Constraint, violationCount,
			)
		}
	}
	__antithesis_instrumentation__.Notify(121605)

	if err != nil {
		__antithesis_instrumentation__.Notify(121616)
		return err
	} else {
		__antithesis_instrumentation__.Notify(121617)
	}
	__antithesis_instrumentation__.Notify(121606)

	r.lastUpdatedRowCount++
	return nil
}

type constraintConformanceVisitor struct {
	cfg           *config.SystemConfig
	storeResolver StoreResolver

	report   ConstraintReport
	visitErr bool

	prevZoneKey     ZoneKey
	prevConstraints []zonepb.ConstraintsConjunction
}

var _ rangeVisitor = &constraintConformanceVisitor{}

func makeConstraintConformanceVisitor(
	ctx context.Context, cfg *config.SystemConfig, storeResolver StoreResolver,
) constraintConformanceVisitor {
	__antithesis_instrumentation__.Notify(121618)
	v := constraintConformanceVisitor{
		cfg:           cfg,
		storeResolver: storeResolver,
	}
	v.reset(ctx)
	return v
}

func (v *constraintConformanceVisitor) failed() bool {
	__antithesis_instrumentation__.Notify(121619)
	return v.visitErr
}

func (v *constraintConformanceVisitor) Report() ConstraintReport {
	__antithesis_instrumentation__.Notify(121620)
	return v.report
}

func (v *constraintConformanceVisitor) reset(ctx context.Context) {
	__antithesis_instrumentation__.Notify(121621)
	*v = constraintConformanceVisitor{
		cfg:           v.cfg,
		storeResolver: v.storeResolver,
		report:        make(ConstraintReport, len(v.report)),
	}

	maxObjectID, err := v.cfg.GetLargestObjectID(
		0, keys.PseudoTableIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(121623)
		log.Fatalf(ctx, "unexpected failure to compute max object id: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(121624)
	}
	__antithesis_instrumentation__.Notify(121622)
	for i := config.ObjectID(1); i <= maxObjectID; i++ {
		__antithesis_instrumentation__.Notify(121625)
		zone, err := getZoneByID(i, v.cfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(121628)
			log.Fatalf(ctx, "unexpected failure to compute max object id: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(121629)
		}
		__antithesis_instrumentation__.Notify(121626)
		if zone == nil {
			__antithesis_instrumentation__.Notify(121630)
			continue
		} else {
			__antithesis_instrumentation__.Notify(121631)
		}
		__antithesis_instrumentation__.Notify(121627)
		v.report.ensureEntries(MakeZoneKey(i, NoSubzone), zone)
	}
}

func (v *constraintConformanceVisitor) visitNewZone(
	ctx context.Context, r *roachpb.RangeDescriptor,
) (retErr error) {
	__antithesis_instrumentation__.Notify(121632)

	defer func() {
		__antithesis_instrumentation__.Notify(121636)
		v.visitErr = retErr != nil
	}()
	__antithesis_instrumentation__.Notify(121633)

	var constraints []zonepb.ConstraintsConjunction
	var zKey ZoneKey
	_, err := visitZones(ctx, r, v.cfg, ignoreSubzonePlaceholders,
		func(_ context.Context, zone *zonepb.ZoneConfig, key ZoneKey) bool {
			__antithesis_instrumentation__.Notify(121637)
			if zone.Constraints == nil {
				__antithesis_instrumentation__.Notify(121639)
				return false
			} else {
				__antithesis_instrumentation__.Notify(121640)
			}
			__antithesis_instrumentation__.Notify(121638)
			constraints = zone.Constraints
			zKey = key
			return true
		})
	__antithesis_instrumentation__.Notify(121634)
	if err != nil {
		__antithesis_instrumentation__.Notify(121641)
		return errors.Wrap(err, "unexpected error visiting zones")
	} else {
		__antithesis_instrumentation__.Notify(121642)
	}
	__antithesis_instrumentation__.Notify(121635)
	v.prevZoneKey = zKey
	v.prevConstraints = constraints
	v.countRange(ctx, r, zKey, constraints)
	return nil
}

func (v *constraintConformanceVisitor) visitSameZone(
	ctx context.Context, r *roachpb.RangeDescriptor,
) {
	__antithesis_instrumentation__.Notify(121643)
	v.countRange(ctx, r, v.prevZoneKey, v.prevConstraints)
}

func (v *constraintConformanceVisitor) countRange(
	ctx context.Context,
	r *roachpb.RangeDescriptor,
	key ZoneKey,
	constraints []zonepb.ConstraintsConjunction,
) {
	__antithesis_instrumentation__.Notify(121644)
	storeDescs := v.storeResolver(r)
	violated := getViolations(ctx, storeDescs, constraints)
	for _, c := range violated {
		__antithesis_instrumentation__.Notify(121645)
		v.report.AddViolation(key, Constraint, c)
	}
}

func getViolations(
	ctx context.Context,
	storeDescs []roachpb.StoreDescriptor,
	constraintConjunctions []zonepb.ConstraintsConjunction,
) []ConstraintRepr {
	__antithesis_instrumentation__.Notify(121646)
	var res []ConstraintRepr

	for _, conjunction := range constraintConjunctions {
		__antithesis_instrumentation__.Notify(121648)
		replicasRequiredToMatch := int(conjunction.NumReplicas)
		if replicasRequiredToMatch == 0 {
			__antithesis_instrumentation__.Notify(121650)
			replicasRequiredToMatch = len(storeDescs)
		} else {
			__antithesis_instrumentation__.Notify(121651)
		}
		__antithesis_instrumentation__.Notify(121649)
		for _, c := range conjunction.Constraints {
			__antithesis_instrumentation__.Notify(121652)
			if !constraintSatisfied(c, replicasRequiredToMatch, storeDescs) {
				__antithesis_instrumentation__.Notify(121653)
				res = append(res, ConstraintRepr(conjunction.String()))
				break
			} else {
				__antithesis_instrumentation__.Notify(121654)
			}
		}
	}
	__antithesis_instrumentation__.Notify(121647)
	return res
}

func constraintSatisfied(
	c zonepb.Constraint, replicasRequiredToMatch int, storeDescs []roachpb.StoreDescriptor,
) bool {
	__antithesis_instrumentation__.Notify(121655)
	passCount := 0
	for _, storeDesc := range storeDescs {
		__antithesis_instrumentation__.Notify(121657)

		if storeDesc.StoreID == 0 {
			__antithesis_instrumentation__.Notify(121659)
			passCount++
			continue
		} else {
			__antithesis_instrumentation__.Notify(121660)
		}
		__antithesis_instrumentation__.Notify(121658)
		if zonepb.StoreSatisfiesConstraint(storeDesc, c) {
			__antithesis_instrumentation__.Notify(121661)
			passCount++
		} else {
			__antithesis_instrumentation__.Notify(121662)
		}
	}
	__antithesis_instrumentation__.Notify(121656)
	return replicasRequiredToMatch <= passCount
}
