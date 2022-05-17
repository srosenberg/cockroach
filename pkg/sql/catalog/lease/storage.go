package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/catkv"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/errors"
)

type storage struct {
	nodeIDContainer  *base.SQLIDContainer
	db               *kv.DB
	clock            *hlc.Clock
	internalExecutor sqlutil.InternalExecutor
	settings         *cluster.Settings
	codec            keys.SQLCodec

	group *singleflight.Group

	outstandingLeases *metric.Gauge
	testingKnobs      StorageTestingKnobs
}

var LeaseRenewalDuration = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.catalog.descriptor_lease_renewal_fraction",
	"controls the default time before a lease expires when acquisition to renew the lease begins",
	base.DefaultDescriptorLeaseRenewalTimeout)

func (s storage) leaseRenewalTimeout() time.Duration {
	__antithesis_instrumentation__.Notify(266607)
	return LeaseRenewalDuration.Get(&s.settings.SV)
}

func (s storage) jitteredLeaseDuration() time.Duration {
	__antithesis_instrumentation__.Notify(266608)
	leaseDuration := LeaseDuration.Get(&s.settings.SV)
	jitterFraction := LeaseJitterFraction.Get(&s.settings.SV)
	return time.Duration(float64(leaseDuration) * (1 - jitterFraction +
		2*jitterFraction*rand.Float64()))
}

func (s storage) acquire(
	ctx context.Context, minExpiration hlc.Timestamp, id descpb.ID,
) (desc catalog.Descriptor, expiration hlc.Timestamp, _ error) {
	__antithesis_instrumentation__.Notify(266609)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	acquireInTxn := func(ctx context.Context, txn *kv.Txn) (err error) {
		__antithesis_instrumentation__.Notify(266612)

		if err := txn.SetUserPriority(roachpb.MaxUserPriority); err != nil {
			__antithesis_instrumentation__.Notify(266621)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266622)
		}
		__antithesis_instrumentation__.Notify(266613)

		nodeID := s.nodeIDContainer.SQLInstanceID()
		if nodeID == 0 {
			__antithesis_instrumentation__.Notify(266623)
			panic("zero nodeID")
		} else {
			__antithesis_instrumentation__.Notify(266624)
		}
		__antithesis_instrumentation__.Notify(266614)

		if !expiration.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(266625)
			return desc != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(266626)
			prevExpirationTS := storedLeaseExpiration(expiration)
			deleteLease := fmt.Sprintf(
				`DELETE FROM system.public.lease WHERE "descID" = %d AND version = %d AND "nodeID" = %d AND expiration = %s`,
				desc.GetID(), desc.GetVersion(), nodeID, &prevExpirationTS,
			)
			if _, err := s.internalExecutor.Exec(
				ctx, "lease-delete-after-ambiguous", txn, deleteLease,
			); err != nil {
				__antithesis_instrumentation__.Notify(266627)
				return errors.Wrap(err, "deleting ambiguously created lease")
			} else {
				__antithesis_instrumentation__.Notify(266628)
			}
		} else {
			__antithesis_instrumentation__.Notify(266629)
		}
		__antithesis_instrumentation__.Notify(266615)

		expiration = txn.ReadTimestamp().Add(int64(s.jitteredLeaseDuration()), 0)
		if expiration.LessEq(minExpiration) {
			__antithesis_instrumentation__.Notify(266630)

			expiration = minExpiration.Add(int64(time.Millisecond), 0)
		} else {
			__antithesis_instrumentation__.Notify(266631)
		}
		__antithesis_instrumentation__.Notify(266616)

		version := s.settings.Version.ActiveVersion(ctx)
		desc, err = catkv.MustGetDescriptorByID(ctx, version, s.codec, txn, nil, id, catalog.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(266632)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266633)
		}
		__antithesis_instrumentation__.Notify(266617)
		if err := catalog.FilterDescriptorState(
			desc, tree.CommonLookupFlags{IncludeOffline: true},
		); err != nil {
			__antithesis_instrumentation__.Notify(266634)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266635)
		}
		__antithesis_instrumentation__.Notify(266618)
		log.VEventf(ctx, 2, "storage attempting to acquire lease %v@%v", desc, expiration)

		ts := storedLeaseExpiration(expiration)
		insertLease := fmt.Sprintf(
			`INSERT INTO system.public.lease ("descID", version, "nodeID", expiration) VALUES (%d, %d, %d, %s)`,
			desc.GetID(), desc.GetVersion(), nodeID, &ts,
		)
		count, err := s.internalExecutor.Exec(ctx, "lease-insert", txn, insertLease)
		if err != nil {
			__antithesis_instrumentation__.Notify(266636)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266637)
		}
		__antithesis_instrumentation__.Notify(266619)
		if count != 1 {
			__antithesis_instrumentation__.Notify(266638)
			return errors.Errorf("%s: expected 1 result, found %d", insertLease, count)
		} else {
			__antithesis_instrumentation__.Notify(266639)
		}
		__antithesis_instrumentation__.Notify(266620)
		return nil
	}
	__antithesis_instrumentation__.Notify(266610)

	for r := retry.StartWithCtx(ctx, retry.Options{}); r.Next(); {
		__antithesis_instrumentation__.Notify(266640)
		err := s.db.Txn(ctx, acquireInTxn)
		var pErr *roachpb.AmbiguousResultError
		if errors.As(err, &pErr) {
			__antithesis_instrumentation__.Notify(266644)
			log.Infof(ctx, "ambiguous error occurred during lease acquisition for %v, retrying: %v", id, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(266645)
		}
		__antithesis_instrumentation__.Notify(266641)
		if err != nil {
			__antithesis_instrumentation__.Notify(266646)
			return nil, hlc.Timestamp{}, err
		} else {
			__antithesis_instrumentation__.Notify(266647)
		}
		__antithesis_instrumentation__.Notify(266642)
		log.VEventf(ctx, 2, "storage acquired lease %v@%v", desc, expiration)
		if s.testingKnobs.LeaseAcquiredEvent != nil {
			__antithesis_instrumentation__.Notify(266648)
			s.testingKnobs.LeaseAcquiredEvent(desc, err)
		} else {
			__antithesis_instrumentation__.Notify(266649)
		}
		__antithesis_instrumentation__.Notify(266643)
		s.outstandingLeases.Inc(1)
		return desc, expiration, nil
	}
	__antithesis_instrumentation__.Notify(266611)
	return nil, hlc.Timestamp{}, ctx.Err()
}

func (s storage) release(ctx context.Context, stopper *stop.Stopper, lease *storedLease) {
	__antithesis_instrumentation__.Notify(266650)
	ctx = multitenant.WithTenantCostControlExemption(ctx)
	retryOptions := base.DefaultRetryOptions()
	retryOptions.Closer = stopper.ShouldQuiesce()
	firstAttempt := true

	for r := retry.Start(retryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(266651)
		log.VEventf(ctx, 2, "storage releasing lease %+v", lease)
		nodeID := s.nodeIDContainer.SQLInstanceID()
		if nodeID == 0 {
			__antithesis_instrumentation__.Notify(266656)
			panic("zero nodeID")
		} else {
			__antithesis_instrumentation__.Notify(266657)
		}
		__antithesis_instrumentation__.Notify(266652)
		const deleteLease = `DELETE FROM system.public.lease ` +
			`WHERE ("descID", version, "nodeID", expiration) = ($1, $2, $3, $4)`
		count, err := s.internalExecutor.Exec(
			ctx,
			"lease-release",
			nil,
			deleteLease,
			lease.id, lease.version, nodeID, &lease.expiration,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(266658)
			log.Warningf(ctx, "error releasing lease %q: %s", lease, err)
			if grpcutil.IsConnectionRejected(err) {
				__antithesis_instrumentation__.Notify(266660)
				return
			} else {
				__antithesis_instrumentation__.Notify(266661)
			}
			__antithesis_instrumentation__.Notify(266659)
			firstAttempt = false
			continue
		} else {
			__antithesis_instrumentation__.Notify(266662)
		}
		__antithesis_instrumentation__.Notify(266653)

		if count > 1 || func() bool {
			__antithesis_instrumentation__.Notify(266663)
			return (count == 0 && func() bool {
				__antithesis_instrumentation__.Notify(266664)
				return firstAttempt == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(266665)
			log.Warningf(ctx, "unexpected results while deleting lease %+v: "+
				"expected 1 result, found %d", lease, count)
		} else {
			__antithesis_instrumentation__.Notify(266666)
		}
		__antithesis_instrumentation__.Notify(266654)

		s.outstandingLeases.Dec(1)
		if s.testingKnobs.LeaseReleasedEvent != nil {
			__antithesis_instrumentation__.Notify(266667)
			s.testingKnobs.LeaseReleasedEvent(
				lease.id, descpb.DescriptorVersion(lease.version), err)
		} else {
			__antithesis_instrumentation__.Notify(266668)
		}
		__antithesis_instrumentation__.Notify(266655)
		break
	}
}

func (s storage) getForExpiration(
	ctx context.Context, expiration hlc.Timestamp, id descpb.ID,
) (catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(266669)
	var desc catalog.Descriptor
	err := s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(266671)
		prevTimestamp := expiration.Prev()
		err := txn.SetFixedTimestamp(ctx, prevTimestamp)
		if err != nil {
			__antithesis_instrumentation__.Notify(266675)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266676)
		}
		__antithesis_instrumentation__.Notify(266672)
		version := s.settings.Version.ActiveVersion(ctx)
		desc, err = catkv.MustGetDescriptorByID(ctx, version, s.codec, txn, nil, id, catalog.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(266677)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266678)
		}
		__antithesis_instrumentation__.Notify(266673)
		if prevTimestamp.LessEq(desc.GetModificationTime()) {
			__antithesis_instrumentation__.Notify(266679)
			return errors.AssertionFailedf("unable to read descriptor"+
				" (%d, %s) found descriptor with modificationTime %s",
				id, expiration, desc.GetModificationTime())
		} else {
			__antithesis_instrumentation__.Notify(266680)
		}
		__antithesis_instrumentation__.Notify(266674)

		return nil
	})
	__antithesis_instrumentation__.Notify(266670)
	return desc, err
}
