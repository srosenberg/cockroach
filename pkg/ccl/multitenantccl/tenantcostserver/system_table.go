package tenantcostserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/multitenantccl/tenantcostserver/tenanttokenbucket"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type tenantState struct {
	Present bool

	LastUpdate tree.DTimestamp

	FirstInstance base.SQLInstanceID

	Bucket tenanttokenbucket.State

	Consumption roachpb.TenantConsumption
}

const defaultRefillRate = 100

const defaultInitialRUs = 10 * 1000 * 1000

const maxInstancesCleanup = 10

func (ts *tenantState) update(now time.Time) {
	__antithesis_instrumentation__.Notify(20226)
	if !ts.Present {
		__antithesis_instrumentation__.Notify(20228)
		*ts = tenantState{
			Present:       true,
			LastUpdate:    tree.DTimestamp{Time: now},
			FirstInstance: 0,
			Bucket: tenanttokenbucket.State{
				RURefillRate: defaultRefillRate,
				RUCurrent:    defaultInitialRUs,
			},
		}
		return
	} else {
		__antithesis_instrumentation__.Notify(20229)
	}
	__antithesis_instrumentation__.Notify(20227)
	delta := now.Sub(ts.LastUpdate.Time)
	if delta > 0 {
		__antithesis_instrumentation__.Notify(20230)

		ts.Bucket.Update(delta)
		ts.LastUpdate.Time = now
	} else {
		__antithesis_instrumentation__.Notify(20231)
	}
}

type instanceState struct {
	ID base.SQLInstanceID

	Present bool

	LastUpdate tree.DTimestamp

	NextInstance base.SQLInstanceID

	Lease tree.DBytes

	Seq int64

	Shares float64
}

type sysTableHelper struct {
	ctx      context.Context
	ex       *sql.InternalExecutor
	txn      *kv.Txn
	tenantID roachpb.TenantID
}

func makeSysTableHelper(
	ctx context.Context, ex *sql.InternalExecutor, txn *kv.Txn, tenantID roachpb.TenantID,
) sysTableHelper {
	__antithesis_instrumentation__.Notify(20232)
	return sysTableHelper{
		ctx:      ctx,
		ex:       ex,
		txn:      txn,
		tenantID: tenantID,
	}
}

func (h *sysTableHelper) readTenantState() (tenant tenantState, _ error) {
	__antithesis_instrumentation__.Notify(20233)

	tenant, _, err := h.readTenantAndInstanceState(0)
	return tenant, err
}

func (h *sysTableHelper) readTenantAndInstanceState(
	instanceID base.SQLInstanceID,
) (tenant tenantState, instance instanceState, _ error) {
	__antithesis_instrumentation__.Notify(20234)
	instance.ID = instanceID

	rows, err := h.ex.QueryBufferedEx(
		h.ctx, "tenant-usage-select", h.txn,
		sessiondata.NodeUserSessionDataOverride,
		`SELECT
		  instance_id,               /* 0 */
			next_instance_id,          /* 1 */
			last_update,               /* 2 */
			ru_burst_limit,            /* 3 */
			ru_refill_rate,            /* 4 */
			ru_current,                /* 5 */
			current_share_sum,         /* 6 */
			total_consumption,         /* 7 */
			instance_lease,            /* 8 */
			instance_seq,              /* 9 */
			instance_shares            /* 10 */
		 FROM system.tenant_usage
		 WHERE tenant_id = $1 AND instance_id IN (0, $2)`,
		h.tenantID.ToUint64(),
		int64(instanceID),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20237)
		return tenantState{}, instanceState{}, err
	} else {
		__antithesis_instrumentation__.Notify(20238)
	}
	__antithesis_instrumentation__.Notify(20235)
	for _, r := range rows {
		__antithesis_instrumentation__.Notify(20239)
		instanceID := base.SQLInstanceID(tree.MustBeDInt(r[0]))
		if instanceID == 0 {
			__antithesis_instrumentation__.Notify(20240)

			tenant.Present = true
			tenant.LastUpdate = tree.MustBeDTimestamp(r[2])
			tenant.FirstInstance = base.SQLInstanceID(tree.MustBeDInt(r[1]))
			tenant.Bucket = tenanttokenbucket.State{
				RUBurstLimit:    float64(tree.MustBeDFloat(r[3])),
				RURefillRate:    float64(tree.MustBeDFloat(r[4])),
				RUCurrent:       float64(tree.MustBeDFloat(r[5])),
				CurrentShareSum: float64(tree.MustBeDFloat(r[6])),
			}
			if consumption := r[7]; consumption != tree.DNull {
				__antithesis_instrumentation__.Notify(20241)

				if err := protoutil.Unmarshal(
					[]byte(tree.MustBeDBytes(consumption)), &tenant.Consumption,
				); err != nil {
					__antithesis_instrumentation__.Notify(20242)
					return tenantState{}, instanceState{}, err
				} else {
					__antithesis_instrumentation__.Notify(20243)
				}
			} else {
				__antithesis_instrumentation__.Notify(20244)
			}
		} else {
			__antithesis_instrumentation__.Notify(20245)

			instance.Present = true
			instance.LastUpdate = tree.MustBeDTimestamp(r[2])
			instance.NextInstance = base.SQLInstanceID(tree.MustBeDInt(r[1]))
			instance.Lease = tree.MustBeDBytes(r[8])
			instance.Seq = int64(tree.MustBeDInt(r[9]))
			instance.Shares = float64(tree.MustBeDFloat(r[10]))
		}
	}
	__antithesis_instrumentation__.Notify(20236)

	return tenant, instance, nil
}

func (h *sysTableHelper) updateTenantState(tenant tenantState) error {
	__antithesis_instrumentation__.Notify(20246)
	consumption, err := protoutil.Marshal(&tenant.Consumption)
	if err != nil {
		__antithesis_instrumentation__.Notify(20248)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20249)
	}
	__antithesis_instrumentation__.Notify(20247)

	_, err = h.ex.ExecEx(
		h.ctx, "tenant-usage-upsert", h.txn,
		sessiondata.NodeUserSessionDataOverride,
		`UPSERT INTO system.tenant_usage(
		  tenant_id,
		  instance_id,
			next_instance_id,
			last_update,
			ru_burst_limit,
			ru_refill_rate,
			ru_current,
			current_share_sum,
			total_consumption,
			instance_lease,
			instance_seq,
			instance_shares
		 ) VALUES ($1, 0, $2, $3, $4, $5, $6, $7, $8, NULL, NULL, NULL)
		 `,
		h.tenantID.ToUint64(),
		int64(tenant.FirstInstance),
		&tenant.LastUpdate,
		tenant.Bucket.RUBurstLimit,
		tenant.Bucket.RURefillRate,
		tenant.Bucket.RUCurrent,
		tenant.Bucket.CurrentShareSum,
		tree.NewDBytes(tree.DBytes(consumption)),
	)
	return err
}

func (h *sysTableHelper) updateTenantAndInstanceState(
	tenant tenantState, instance instanceState,
) error {
	__antithesis_instrumentation__.Notify(20250)
	consumption, err := protoutil.Marshal(&tenant.Consumption)
	if err != nil {
		__antithesis_instrumentation__.Notify(20252)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20253)
	}
	__antithesis_instrumentation__.Notify(20251)

	_, err = h.ex.ExecEx(
		h.ctx, "tenant-usage-insert", h.txn,
		sessiondata.NodeUserSessionDataOverride,
		`UPSERT INTO system.tenant_usage(
		  tenant_id,
		  instance_id,
			next_instance_id,
			last_update,
			ru_burst_limit,
			ru_refill_rate,
			ru_current,
			current_share_sum,
			total_consumption,
			instance_lease,
			instance_seq,
			instance_shares
		 ) VALUES
		   ($1, 0,  $2,  $3,  $4,   $5,   $6,   $7,   $8,   NULL, NULL, NULL),
			 ($1, $9, $10, $11, NULL, NULL, NULL, NULL, NULL, $12,  $13,  $14)
		 `,
		h.tenantID.ToUint64(),
		int64(tenant.FirstInstance),
		&tenant.LastUpdate,
		tenant.Bucket.RUBurstLimit,
		tenant.Bucket.RURefillRate,
		tenant.Bucket.RUCurrent,
		tenant.Bucket.CurrentShareSum,
		tree.NewDBytes(tree.DBytes(consumption)),
		int64(instance.ID),
		int64(instance.NextInstance),
		&instance.LastUpdate,
		&instance.Lease,
		instance.Seq,
		instance.Shares,
	)
	return err
}

func (h *sysTableHelper) accomodateNewInstance(tenant *tenantState, instance *instanceState) error {
	__antithesis_instrumentation__.Notify(20254)
	if tenant.FirstInstance == 0 || func() bool {
		__antithesis_instrumentation__.Notify(20258)
		return tenant.FirstInstance > instance.ID == true
	}() == true {
		__antithesis_instrumentation__.Notify(20259)

		instance.NextInstance = tenant.FirstInstance
		tenant.FirstInstance = instance.ID
		return nil
	} else {
		__antithesis_instrumentation__.Notify(20260)
	}
	__antithesis_instrumentation__.Notify(20255)

	row, err := h.ex.QueryRowEx(
		h.ctx, "find-prev-id", h.txn,
		sessiondata.NodeUserSessionDataOverride,
		`SELECT
		  instance_id,               /* 0 */
			next_instance_id,          /* 1 */
			last_update,               /* 2 */
			instance_lease,            /* 3 */
			instance_seq,              /* 4 */
			instance_shares            /* 5 */
		 FROM system.tenant_usage
		 WHERE tenant_id = $1 AND instance_id > 0 AND instance_id < $2
		 ORDER BY instance_id DESC
		 LIMIT 1`,
		h.tenantID.ToUint64(),
		int64(instance.ID),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20261)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20262)
	}
	__antithesis_instrumentation__.Notify(20256)
	if row == nil {
		__antithesis_instrumentation__.Notify(20263)
		return errors.Errorf("could not find row for previous instance")
	} else {
		__antithesis_instrumentation__.Notify(20264)
	}
	__antithesis_instrumentation__.Notify(20257)
	prevInstanceID := base.SQLInstanceID(tree.MustBeDInt(row[0]))
	instance.NextInstance = base.SQLInstanceID(tree.MustBeDInt(row[1]))
	prevInstanceLastUpdate := row[2]
	prevInstanceLease := row[3]
	prevInstanceSeq := row[4]
	prevInstanceShares := row[5]

	_, err = h.ex.ExecEx(
		h.ctx, "update-next-id", h.txn,
		sessiondata.NodeUserSessionDataOverride,

		`UPSERT INTO system.tenant_usage(
		  tenant_id,
		  instance_id,
			next_instance_id,
			last_update,
			ru_burst_limit,
			ru_refill_rate,
			ru_current,
			current_share_sum,
			total_consumption,
			instance_lease,
			instance_seq,
			instance_shares
		 ) VALUES ($1, $2, $3, $4, NULL, NULL, NULL, NULL, NULL, $5, $6, $7)
		`,
		h.tenantID.ToUint64(),
		int64(prevInstanceID),
		int64(instance.ID),
		prevInstanceLastUpdate,
		prevInstanceLease,
		prevInstanceSeq,
		prevInstanceShares,
	)
	return err
}

func (h *sysTableHelper) maybeCleanupStaleInstance(
	cutoff time.Time, instanceID base.SQLInstanceID,
) (deleted bool, nextInstance base.SQLInstanceID, _ error) {
	__antithesis_instrumentation__.Notify(20265)
	ts := tree.MustMakeDTimestamp(cutoff, time.Microsecond)
	row, err := h.ex.QueryRowEx(
		h.ctx, "tenant-usage-delete", h.txn,
		sessiondata.NodeUserSessionDataOverride,
		`DELETE FROM system.tenant_usage
		 WHERE tenant_id = $1 AND instance_id = $2 AND last_update < $3
		 RETURNING next_instance_id`,
		h.tenantID.ToUint64(),
		int32(instanceID),
		ts,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20268)
		return false, -1, err
	} else {
		__antithesis_instrumentation__.Notify(20269)
	}
	__antithesis_instrumentation__.Notify(20266)
	if row == nil {
		__antithesis_instrumentation__.Notify(20270)
		log.VEventf(h.ctx, 1, "tenant %s instance %d not stale", h.tenantID, instanceID)
		return false, -1, nil
	} else {
		__antithesis_instrumentation__.Notify(20271)
	}
	__antithesis_instrumentation__.Notify(20267)
	nextInstance = base.SQLInstanceID(tree.MustBeDInt(row[0]))
	log.VEventf(h.ctx, 1, "cleaned up tenant %s instance %d", h.tenantID, instanceID)
	return true, nextInstance, nil
}

func (h *sysTableHelper) maybeCleanupStaleInstances(
	cutoff time.Time, startID, endID base.SQLInstanceID,
) (nextInstance base.SQLInstanceID, _ error) {
	__antithesis_instrumentation__.Notify(20272)
	log.VEventf(
		h.ctx, 1, "checking stale instances (tenant=%s startID=%d endID=%d)",
		h.tenantID, startID, endID,
	)
	id := startID
	for n := 0; n < maxInstancesCleanup; n++ {
		__antithesis_instrumentation__.Notify(20274)
		deleted, nextInstance, err := h.maybeCleanupStaleInstance(cutoff, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(20277)
			return -1, err
		} else {
			__antithesis_instrumentation__.Notify(20278)
		}
		__antithesis_instrumentation__.Notify(20275)
		if !deleted {
			__antithesis_instrumentation__.Notify(20279)
			break
		} else {
			__antithesis_instrumentation__.Notify(20280)
		}
		__antithesis_instrumentation__.Notify(20276)
		id = nextInstance
		if id == 0 || func() bool {
			__antithesis_instrumentation__.Notify(20281)
			return (endID != -1 && func() bool {
				__antithesis_instrumentation__.Notify(20282)
				return id >= endID == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(20283)
			break
		} else {
			__antithesis_instrumentation__.Notify(20284)
		}
	}
	__antithesis_instrumentation__.Notify(20273)
	return id, nil
}

func (h *sysTableHelper) maybeCheckInvariants() error {
	__antithesis_instrumentation__.Notify(20285)
	if buildutil.CrdbTestBuild && func() bool {
		__antithesis_instrumentation__.Notify(20287)
		return rand.Intn(10) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(20288)
		return h.checkInvariants()
	} else {
		__antithesis_instrumentation__.Notify(20289)
	}
	__antithesis_instrumentation__.Notify(20286)
	return nil
}

func (h *sysTableHelper) checkInvariants() error {
	__antithesis_instrumentation__.Notify(20290)

	rows, err := h.ex.QueryBufferedEx(
		h.ctx, "tenant-usage-select", h.txn,
		sessiondata.NodeUserSessionDataOverride,
		`SELECT
		  instance_id,               /* 0 */
			next_instance_id,          /* 1 */
			last_update,               /* 2 */
			ru_burst_limit,            /* 3 */
			ru_refill_rate,            /* 4 */
			ru_current,                /* 5 */
			current_share_sum,         /* 6 */
			total_consumption,         /* 7 */
			instance_lease,            /* 8 */
			instance_seq,              /* 9 */
			instance_shares            /* 10 */
		 FROM system.tenant_usage
		 WHERE tenant_id = $1
		 ORDER BY instance_id`,
		h.tenantID.ToUint64(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20300)
		if h.ctx.Err() == nil {
			__antithesis_instrumentation__.Notify(20302)
			log.Warningf(h.ctx, "checkInvariants query failed: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(20303)
		}
		__antithesis_instrumentation__.Notify(20301)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(20304)
	}
	__antithesis_instrumentation__.Notify(20291)
	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(20305)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(20306)
	}
	__antithesis_instrumentation__.Notify(20292)

	instanceIDs := make([]base.SQLInstanceID, len(rows))
	for i := range rows {
		__antithesis_instrumentation__.Notify(20307)
		instanceIDs[i] = base.SQLInstanceID(tree.MustBeDInt(rows[i][0]))
		if i > 0 && func() bool {
			__antithesis_instrumentation__.Notify(20308)
			return instanceIDs[i-1] >= instanceIDs[i] == true
		}() == true {
			__antithesis_instrumentation__.Notify(20309)
			return errors.New("instances out of order")
		} else {
			__antithesis_instrumentation__.Notify(20310)
		}
	}
	__antithesis_instrumentation__.Notify(20293)
	if instanceIDs[0] != 0 {
		__antithesis_instrumentation__.Notify(20311)
		return errors.New("instance 0 row missing")
	} else {
		__antithesis_instrumentation__.Notify(20312)
	}
	__antithesis_instrumentation__.Notify(20294)

	for i := range rows {
		__antithesis_instrumentation__.Notify(20313)
		var nullFirst, nullLast int
		if i == 0 {
			__antithesis_instrumentation__.Notify(20315)

			nullFirst, nullLast = 8, 10
		} else {
			__antithesis_instrumentation__.Notify(20316)

			nullFirst, nullLast = 3, 7
		}
		__antithesis_instrumentation__.Notify(20314)
		for j := range rows[i] {
			__antithesis_instrumentation__.Notify(20317)
			isNull := (rows[i][j] == tree.DNull)
			expNull := (j >= nullFirst && func() bool {
				__antithesis_instrumentation__.Notify(20318)
				return j <= nullLast == true
			}() == true)
			if expNull != isNull {
				__antithesis_instrumentation__.Notify(20319)
				if !expNull {
					__antithesis_instrumentation__.Notify(20321)
					return errors.Errorf("expected NULL column %d", j)
				} else {
					__antithesis_instrumentation__.Notify(20322)
				}
				__antithesis_instrumentation__.Notify(20320)

				if i != 7 {
					__antithesis_instrumentation__.Notify(20323)
					return errors.Errorf("expected non-NULL column %d", j)
				} else {
					__antithesis_instrumentation__.Notify(20324)
				}
			} else {
				__antithesis_instrumentation__.Notify(20325)
			}
		}
	}
	__antithesis_instrumentation__.Notify(20295)

	for i := range rows {
		__antithesis_instrumentation__.Notify(20326)
		expNextInstanceID := base.SQLInstanceID(0)
		if i+1 < len(rows) {
			__antithesis_instrumentation__.Notify(20328)
			expNextInstanceID = instanceIDs[i+1]
		} else {
			__antithesis_instrumentation__.Notify(20329)
		}
		__antithesis_instrumentation__.Notify(20327)
		nextInstanceID := base.SQLInstanceID(tree.MustBeDInt(rows[i][1]))
		if expNextInstanceID != nextInstanceID {
			__antithesis_instrumentation__.Notify(20330)
			return errors.Errorf("expected next instance %d, have %d", expNextInstanceID, nextInstanceID)
		} else {
			__antithesis_instrumentation__.Notify(20331)
		}
	}
	__antithesis_instrumentation__.Notify(20296)

	sharesSum := float64(tree.MustBeDFloat(rows[0][6]))
	var expSharesSum float64
	for _, r := range rows[1:] {
		__antithesis_instrumentation__.Notify(20332)
		expSharesSum += float64(tree.MustBeDFloat(r[10]))
	}
	__antithesis_instrumentation__.Notify(20297)

	a, b := sharesSum, expSharesSum
	if a > b {
		__antithesis_instrumentation__.Notify(20333)
		a, b = b, a
	} else {
		__antithesis_instrumentation__.Notify(20334)
	}
	__antithesis_instrumentation__.Notify(20298)

	const ulpTolerance = 1000
	if math.Float64bits(a)+ulpTolerance <= math.Float64bits(b) {
		__antithesis_instrumentation__.Notify(20335)
		return errors.Errorf("expected shares sum %g, have %g", expSharesSum, sharesSum)
	} else {
		__antithesis_instrumentation__.Notify(20336)
	}
	__antithesis_instrumentation__.Notify(20299)
	return nil
}

func InspectTenantMetadata(
	ctx context.Context,
	ex *sql.InternalExecutor,
	txn *kv.Txn,
	tenantID roachpb.TenantID,
	timeFormat string,
) (string, error) {
	__antithesis_instrumentation__.Notify(20337)
	h := makeSysTableHelper(ctx, ex, txn, tenantID)
	tenant, err := h.readTenantState()
	if err != nil {
		__antithesis_instrumentation__.Notify(20342)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(20343)
	}
	__antithesis_instrumentation__.Notify(20338)
	if !tenant.Present {
		__antithesis_instrumentation__.Notify(20344)
		return "empty state", nil
	} else {
		__antithesis_instrumentation__.Notify(20345)
	}
	__antithesis_instrumentation__.Notify(20339)

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "Bucket state: ru-burst-limit=%g  ru-refill-rate=%g  ru-current=%.12g  current-share-sum=%.12g\n",
		tenant.Bucket.RUBurstLimit,
		tenant.Bucket.RURefillRate,
		tenant.Bucket.RUCurrent,
		tenant.Bucket.CurrentShareSum,
	)
	fmt.Fprintf(&buf, "Consumption: ru=%.12g  reads=%d req/%d bytes  writes=%d req/%d bytes  pod-cpu-usage: %g secs  pgwire-egress=%d bytes\n",
		tenant.Consumption.RU,
		tenant.Consumption.ReadRequests,
		tenant.Consumption.ReadBytes,
		tenant.Consumption.WriteRequests,
		tenant.Consumption.WriteBytes,
		tenant.Consumption.SQLPodsCPUSeconds,
		tenant.Consumption.PGWireEgressBytes,
	)
	fmt.Fprintf(&buf, "Last update: %s\n", tenant.LastUpdate.Time.Format(timeFormat))
	fmt.Fprintf(&buf, "First active instance: %d\n", tenant.FirstInstance)

	rows, err := ex.QueryBufferedEx(
		ctx, "inspect-tenant-state", txn,
		sessiondata.NodeUserSessionDataOverride,
		`SELECT
		  instance_id,               /* 0 */
			next_instance_id,          /* 1 */
			last_update,               /* 2 */
			instance_lease,            /* 3 */
			instance_seq,              /* 4 */
			instance_shares            /* 5 */
		 FROM system.tenant_usage
		 WHERE tenant_id = $1 AND instance_id > 0
		 ORDER BY instance_id`,
		tenantID.ToUint64(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20346)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(20347)
	}
	__antithesis_instrumentation__.Notify(20340)
	for _, r := range rows {
		__antithesis_instrumentation__.Notify(20348)
		fmt.Fprintf(
			&buf, "  Instance %s:  lease=%q  seq=%s  shares=%s  next-instance=%s  last-update=%s\n",
			r[0], tree.MustBeDBytes(r[3]), r[4], r[5], r[1], tree.MustBeDTimestamp(r[2]).Time.Format(timeFormat),
		)
	}
	__antithesis_instrumentation__.Notify(20341)
	return buf.String(), nil
}
