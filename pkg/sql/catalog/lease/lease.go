// Package lease provides functionality to create and manage sql schema leases.
package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descbuilder"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/catkv"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	kvstorage "github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

var errRenewLease = errors.New("renew lease on id")
var errReadOlderVersion = errors.New("read older descriptor version from store")

var LeaseDuration = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.catalog.descriptor_lease_duration",
	"mean duration of sql descriptor leases, this actual duration is jitterred",
	base.DefaultDescriptorLeaseDuration)

func between0and1inclusive(f float64) error {
	__antithesis_instrumentation__.Notify(266192)
	if f < 0 || func() bool {
		__antithesis_instrumentation__.Notify(266194)
		return f > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(266195)
		return errors.Errorf("value %f must be between 0 and 1", f)
	} else {
		__antithesis_instrumentation__.Notify(266196)
	}
	__antithesis_instrumentation__.Notify(266193)
	return nil
}

var LeaseJitterFraction = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"sql.catalog.descriptor_lease_jitter_fraction",
	"mean duration of sql descriptor leases, this actual duration is jitterred",
	base.DefaultDescriptorLeaseJitterFraction,
	between0and1inclusive)

func (m *Manager) WaitForNoVersion(
	ctx context.Context, id descpb.ID, retryOpts retry.Options,
) error {
	__antithesis_instrumentation__.Notify(266197)
	for lastCount, r := 0, retry.Start(retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(266199)

		now := m.storage.clock.Now()
		stmt := fmt.Sprintf(`SELECT count(1) FROM system.public.lease AS OF SYSTEM TIME '%s' WHERE ("descID" = %d AND expiration > $1)`,
			now.AsOfSystemTime(),
			id)
		values, err := m.storage.internalExecutor.QueryRowEx(
			ctx, "count-leases", nil,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			stmt, now.GoTime(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(266203)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266204)
		}
		__antithesis_instrumentation__.Notify(266200)
		if values == nil {
			__antithesis_instrumentation__.Notify(266205)
			return errors.New("failed to count leases")
		} else {
			__antithesis_instrumentation__.Notify(266206)
		}
		__antithesis_instrumentation__.Notify(266201)
		count := int(tree.MustBeDInt(values[0]))
		if count == 0 {
			__antithesis_instrumentation__.Notify(266207)
			break
		} else {
			__antithesis_instrumentation__.Notify(266208)
		}
		__antithesis_instrumentation__.Notify(266202)
		if count != lastCount {
			__antithesis_instrumentation__.Notify(266209)
			lastCount = count
			log.Infof(ctx, "waiting for %d leases to expire: desc=%d", count, id)
		} else {
			__antithesis_instrumentation__.Notify(266210)
		}
	}
	__antithesis_instrumentation__.Notify(266198)
	return nil
}

func (m *Manager) WaitForOneVersion(
	ctx context.Context, id descpb.ID, retryOpts retry.Options,
) (desc catalog.Descriptor, _ error) {
	__antithesis_instrumentation__.Notify(266211)
	for lastCount, r := 0, retry.Start(retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(266213)
		if err := m.DB().Txn(ctx, func(ctx context.Context, txn *kv.Txn) (err error) {
			__antithesis_instrumentation__.Notify(266217)
			version := m.storage.settings.Version.ActiveVersion(ctx)
			desc, err = catkv.MustGetDescriptorByID(ctx, version, m.Codec(), txn, nil, id, catalog.Any)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(266218)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266219)
		}
		__antithesis_instrumentation__.Notify(266214)

		now := m.storage.clock.Now()
		descs := []IDVersion{NewIDVersionPrev(desc.GetName(), desc.GetID(), desc.GetVersion())}
		count, err := CountLeases(ctx, m.storage.internalExecutor, descs, now)
		if err != nil {
			__antithesis_instrumentation__.Notify(266220)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266221)
		}
		__antithesis_instrumentation__.Notify(266215)
		if count == 0 {
			__antithesis_instrumentation__.Notify(266222)
			break
		} else {
			__antithesis_instrumentation__.Notify(266223)
		}
		__antithesis_instrumentation__.Notify(266216)
		if count != lastCount {
			__antithesis_instrumentation__.Notify(266224)
			lastCount = count
			log.Infof(ctx, "waiting for %d leases to expire: desc=%v", count, descs)
		} else {
			__antithesis_instrumentation__.Notify(266225)
		}
	}
	__antithesis_instrumentation__.Notify(266212)
	return desc, nil
}

type IDVersion struct {
	Name    string
	ID      descpb.ID
	Version descpb.DescriptorVersion
}

func NewIDVersionPrev(name string, id descpb.ID, currVersion descpb.DescriptorVersion) IDVersion {
	__antithesis_instrumentation__.Notify(266226)
	return IDVersion{Name: name, ID: id, Version: currVersion - 1}
}

func ensureVersion(
	ctx context.Context, id descpb.ID, minVersion descpb.DescriptorVersion, m *Manager,
) error {
	__antithesis_instrumentation__.Notify(266227)
	if s := m.findNewest(id); s != nil && func() bool {
		__antithesis_instrumentation__.Notify(266231)
		return minVersion <= s.GetVersion() == true
	}() == true {
		__antithesis_instrumentation__.Notify(266232)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266233)
	}
	__antithesis_instrumentation__.Notify(266228)

	if err := m.AcquireFreshestFromStore(ctx, id); err != nil {
		__antithesis_instrumentation__.Notify(266234)
		return err
	} else {
		__antithesis_instrumentation__.Notify(266235)
	}
	__antithesis_instrumentation__.Notify(266229)

	if s := m.findNewest(id); s != nil && func() bool {
		__antithesis_instrumentation__.Notify(266236)
		return s.GetVersion() < minVersion == true
	}() == true {
		__antithesis_instrumentation__.Notify(266237)
		return errors.Errorf("version %d for descriptor %s does not exist yet", minVersion, s.GetName())
	} else {
		__antithesis_instrumentation__.Notify(266238)
	}
	__antithesis_instrumentation__.Notify(266230)
	return nil
}

type historicalDescriptor struct {
	desc       catalog.Descriptor
	expiration hlc.Timestamp
}

func getDescriptorsFromStoreForInterval(
	ctx context.Context,
	db *kv.DB,
	codec keys.SQLCodec,
	id descpb.ID,
	lowerBound, upperBound hlc.Timestamp,
) ([]historicalDescriptor, error) {
	__antithesis_instrumentation__.Notify(266239)

	if lowerBound.IsEmpty() {
		__antithesis_instrumentation__.Notify(266244)
		return nil, errors.AssertionFailedf(
			"getDescriptorsFromStoreForInterval: lower bound cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(266245)
	}
	__antithesis_instrumentation__.Notify(266240)

	if upperBound.IsEmpty() {
		__antithesis_instrumentation__.Notify(266246)
		return nil, errors.AssertionFailedf(
			"getDescriptorsFromStoreForInterval: upper bound cannot be empty")
	} else {
		__antithesis_instrumentation__.Notify(266247)
	}
	__antithesis_instrumentation__.Notify(266241)

	batchRequestHeader := roachpb.Header{
		Timestamp: upperBound.Prev(),
	}
	descriptorKey := catalogkeys.MakeDescMetadataKey(codec, id)
	requestHeader := roachpb.RequestHeader{
		Key:    descriptorKey,
		EndKey: descriptorKey.PrefixEnd(),
	}
	req := &roachpb.ExportRequest{
		RequestHeader: requestHeader,
		StartTime:     lowerBound.Prev(),
		MVCCFilter:    roachpb.MVCCFilter_All,
		ReturnSST:     true,
	}

	res, pErr := kv.SendWrappedWith(ctx, db.NonTransactionalSender(), batchRequestHeader, req)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(266248)
		return nil, errors.Wrapf(pErr.GoError(), "error in retrieving descs between %s, %s",
			lowerBound, upperBound)
	} else {
		__antithesis_instrumentation__.Notify(266249)
	}
	__antithesis_instrumentation__.Notify(266242)

	var descriptorsRead []historicalDescriptor

	subsequentModificationTime := upperBound
	for _, file := range res.(*roachpb.ExportResponse).Files {
		__antithesis_instrumentation__.Notify(266250)
		if err := func() error {
			__antithesis_instrumentation__.Notify(266251)
			it, err := kvstorage.NewMemSSTIterator(file.SST, false)
			if err != nil {
				__antithesis_instrumentation__.Notify(266253)
				return err
			} else {
				__antithesis_instrumentation__.Notify(266254)
			}
			__antithesis_instrumentation__.Notify(266252)
			defer it.Close()

			for it.SeekGE(kvstorage.NilKey); ; it.Next() {
				__antithesis_instrumentation__.Notify(266255)
				if ok, err := it.Valid(); err != nil {
					__antithesis_instrumentation__.Notify(266259)
					return err
				} else {
					__antithesis_instrumentation__.Notify(266260)
					if !ok {
						__antithesis_instrumentation__.Notify(266261)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(266262)
					}
				}
				__antithesis_instrumentation__.Notify(266256)

				k := it.UnsafeKey()
				descContent := it.UnsafeValue()
				if descContent == nil {
					__antithesis_instrumentation__.Notify(266263)
					return errors.Wrapf(errors.New("unsafe value error"), "error "+
						"extracting raw bytes of descriptor with key %s modified between "+
						"%s, %s", k.String(), k.Timestamp, subsequentModificationTime)
				} else {
					__antithesis_instrumentation__.Notify(266264)
				}
				__antithesis_instrumentation__.Notify(266257)

				value := roachpb.Value{RawBytes: descContent}
				var desc descpb.Descriptor
				if err := value.GetProto(&desc); err != nil {
					__antithesis_instrumentation__.Notify(266265)
					return err
				} else {
					__antithesis_instrumentation__.Notify(266266)
				}
				__antithesis_instrumentation__.Notify(266258)
				descBuilder := descbuilder.NewBuilderWithMVCCTimestamp(&desc, k.Timestamp)

				histDesc := historicalDescriptor{
					desc:       descBuilder.BuildImmutable(),
					expiration: subsequentModificationTime,
				}
				descriptorsRead = append(descriptorsRead, histDesc)

				subsequentModificationTime = k.Timestamp
			}
		}(); err != nil {
			__antithesis_instrumentation__.Notify(266267)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266268)
		}
	}
	__antithesis_instrumentation__.Notify(266243)
	return descriptorsRead, nil
}

func (m *Manager) readOlderVersionForTimestamp(
	ctx context.Context, id descpb.ID, timestamp hlc.Timestamp,
) ([]historicalDescriptor, error) {
	__antithesis_instrumentation__.Notify(266269)

	t := m.findDescriptorState(id, false)

	if t == nil {
		__antithesis_instrumentation__.Notify(266276)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(266277)
	}
	__antithesis_instrumentation__.Notify(266270)
	endTimestamp, done := func() (hlc.Timestamp, bool) {
		__antithesis_instrumentation__.Notify(266278)
		t.mu.Lock()
		defer t.mu.Unlock()

		if len(t.mu.active.data) == 0 {
			__antithesis_instrumentation__.Notify(266282)
			return hlc.Timestamp{}, true
		} else {
			__antithesis_instrumentation__.Notify(266283)
		}
		__antithesis_instrumentation__.Notify(266279)

		i := sort.Search(len(t.mu.active.data), func(i int) bool {
			__antithesis_instrumentation__.Notify(266284)
			return timestamp.Less(t.mu.active.data[i].GetModificationTime())
		})
		__antithesis_instrumentation__.Notify(266280)

		if i == len(t.mu.active.data) || func() bool {
			__antithesis_instrumentation__.Notify(266285)
			return (i > 0 && func() bool {
				__antithesis_instrumentation__.Notify(266286)
				return timestamp.Less(t.mu.active.data[i-1].getExpiration()) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(266287)
			return hlc.Timestamp{}, true
		} else {
			__antithesis_instrumentation__.Notify(266288)
		}
		__antithesis_instrumentation__.Notify(266281)
		return t.mu.active.data[i].GetModificationTime(), false
	}()
	__antithesis_instrumentation__.Notify(266271)
	if done {
		__antithesis_instrumentation__.Notify(266289)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(266290)
	}
	__antithesis_instrumentation__.Notify(266272)

	descs, err := getDescriptorsFromStoreForInterval(ctx, m.DB(), m.Codec(), id, timestamp, endTimestamp)
	if err != nil {
		__antithesis_instrumentation__.Notify(266291)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266292)
	}
	__antithesis_instrumentation__.Notify(266273)

	var earliestModificationTime hlc.Timestamp
	if len(descs) == 0 {
		__antithesis_instrumentation__.Notify(266293)
		earliestModificationTime = endTimestamp
	} else {
		__antithesis_instrumentation__.Notify(266294)
		earliestModificationTime = descs[len(descs)-1].desc.GetModificationTime()
	}
	__antithesis_instrumentation__.Notify(266274)

	if timestamp.Less(earliestModificationTime) {
		__antithesis_instrumentation__.Notify(266295)
		desc, err := m.storage.getForExpiration(ctx, earliestModificationTime, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(266297)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266298)
		}
		__antithesis_instrumentation__.Notify(266296)
		descs = append(descs, historicalDescriptor{
			desc:       desc,
			expiration: earliestModificationTime,
		})
	} else {
		__antithesis_instrumentation__.Notify(266299)
	}
	__antithesis_instrumentation__.Notify(266275)

	return descs, nil
}

func (m *Manager) insertDescriptorVersions(id descpb.ID, versions []historicalDescriptor) {
	__antithesis_instrumentation__.Notify(266300)
	t := m.findDescriptorState(id, false)
	t.mu.Lock()
	defer t.mu.Unlock()
	for i := range versions {
		__antithesis_instrumentation__.Notify(266301)

		existingVersion := t.mu.active.findVersion(versions[i].desc.GetVersion())
		if existingVersion == nil {
			__antithesis_instrumentation__.Notify(266302)
			t.mu.active.insert(
				newDescriptorVersionState(t, versions[i].desc, versions[i].expiration, false))
		} else {
			__antithesis_instrumentation__.Notify(266303)
		}
	}
}

func (m *Manager) AcquireFreshestFromStore(ctx context.Context, id descpb.ID) error {
	__antithesis_instrumentation__.Notify(266304)

	_ = m.findDescriptorState(id, true)

	attemptsMade := 0
	for {
		__antithesis_instrumentation__.Notify(266306)

		didAcquire, err := acquireNodeLease(ctx, m, id)
		if m.testingKnobs.LeaseStoreTestingKnobs.LeaseAcquireResultBlockEvent != nil {
			__antithesis_instrumentation__.Notify(266310)
			m.testingKnobs.LeaseStoreTestingKnobs.LeaseAcquireResultBlockEvent(AcquireFreshestBlock, id)
		} else {
			__antithesis_instrumentation__.Notify(266311)
		}
		__antithesis_instrumentation__.Notify(266307)
		if err != nil {
			__antithesis_instrumentation__.Notify(266312)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266313)
		}
		__antithesis_instrumentation__.Notify(266308)

		if didAcquire {
			__antithesis_instrumentation__.Notify(266314)

			break
		} else {
			__antithesis_instrumentation__.Notify(266315)
			if attemptsMade > 1 {
				__antithesis_instrumentation__.Notify(266316)

				break
			} else {
				__antithesis_instrumentation__.Notify(266317)
			}
		}
		__antithesis_instrumentation__.Notify(266309)
		attemptsMade++
	}
	__antithesis_instrumentation__.Notify(266305)
	return nil
}

func acquireNodeLease(ctx context.Context, m *Manager, id descpb.ID) (bool, error) {
	__antithesis_instrumentation__.Notify(266318)
	var toRelease *storedLease
	resultChan, didAcquire := m.storage.group.DoChan(fmt.Sprintf("acquire%d", id), func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(266321)

		lt := logtags.FromContext(ctx)
		ctx, cancel := m.stopper.WithCancelOnQuiesce(logtags.AddTags(m.ambientCtx.AnnotateCtx(context.Background()), lt))
		defer cancel()
		if m.isDraining() {
			__antithesis_instrumentation__.Notify(266328)
			return nil, errors.New("cannot acquire lease when draining")
		} else {
			__antithesis_instrumentation__.Notify(266329)
		}
		__antithesis_instrumentation__.Notify(266322)
		newest := m.findNewest(id)
		var minExpiration hlc.Timestamp
		if newest != nil {
			__antithesis_instrumentation__.Notify(266330)
			minExpiration = newest.getExpiration()
		} else {
			__antithesis_instrumentation__.Notify(266331)
		}
		__antithesis_instrumentation__.Notify(266323)
		desc, expiration, err := m.storage.acquire(ctx, minExpiration, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(266332)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266333)
		}
		__antithesis_instrumentation__.Notify(266324)
		t := m.findDescriptorState(id, false)
		t.mu.Lock()
		t.mu.takenOffline = false
		defer t.mu.Unlock()
		var newDescVersionState *descriptorVersionState
		newDescVersionState, toRelease, err = t.upsertLeaseLocked(ctx, desc, expiration)
		if err != nil {
			__antithesis_instrumentation__.Notify(266334)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266335)
		}
		__antithesis_instrumentation__.Notify(266325)
		if newDescVersionState != nil {
			__antithesis_instrumentation__.Notify(266336)
			m.names.insert(newDescVersionState)
		} else {
			__antithesis_instrumentation__.Notify(266337)
		}
		__antithesis_instrumentation__.Notify(266326)
		if toRelease != nil {
			__antithesis_instrumentation__.Notify(266338)
			releaseLease(ctx, toRelease, m)
		} else {
			__antithesis_instrumentation__.Notify(266339)
		}
		__antithesis_instrumentation__.Notify(266327)
		return true, nil
	})
	__antithesis_instrumentation__.Notify(266319)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(266340)
		return false, ctx.Err()
	case result := <-resultChan:
		__antithesis_instrumentation__.Notify(266341)
		if result.Err != nil {
			__antithesis_instrumentation__.Notify(266342)
			return false, result.Err
		} else {
			__antithesis_instrumentation__.Notify(266343)
		}
	}
	__antithesis_instrumentation__.Notify(266320)
	return didAcquire, nil
}

func releaseLease(ctx context.Context, lease *storedLease, m *Manager) {
	__antithesis_instrumentation__.Notify(266344)
	if m.isDraining() {
		__antithesis_instrumentation__.Notify(266346)

		m.storage.release(ctx, m.stopper, lease)
		return
	} else {
		__antithesis_instrumentation__.Notify(266347)
	}
	__antithesis_instrumentation__.Notify(266345)

	newCtx := m.ambientCtx.AnnotateCtx(context.Background())

	newCtx = logtags.AddTags(newCtx, logtags.FromContext(ctx))
	if err := m.stopper.RunAsyncTask(
		newCtx, "sql.descriptorState: releasing descriptor lease",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(266348)
			m.storage.release(ctx, m.stopper, lease)
		}); err != nil {
		__antithesis_instrumentation__.Notify(266349)
		log.Warningf(ctx, "error: %s, not releasing lease: %q", err, lease)
	} else {
		__antithesis_instrumentation__.Notify(266350)
	}
}

func purgeOldVersions(
	ctx context.Context,
	db *kv.DB,
	id descpb.ID,
	dropped bool,
	minVersion descpb.DescriptorVersion,
	m *Manager,
) error {
	__antithesis_instrumentation__.Notify(266351)
	t := m.findDescriptorState(id, false)
	if t == nil {
		__antithesis_instrumentation__.Notify(266359)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266360)
	}
	__antithesis_instrumentation__.Notify(266352)
	t.mu.Lock()
	if t.mu.maxVersionSeen < minVersion {
		__antithesis_instrumentation__.Notify(266361)
		t.mu.maxVersionSeen = minVersion
	} else {
		__antithesis_instrumentation__.Notify(266362)
	}
	__antithesis_instrumentation__.Notify(266353)
	empty := len(t.mu.active.data) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(266363)
		return t.mu.acquisitionsInProgress == 0 == true
	}() == true
	t.mu.Unlock()
	if empty && func() bool {
		__antithesis_instrumentation__.Notify(266364)
		return !dropped == true
	}() == true {
		__antithesis_instrumentation__.Notify(266365)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(266366)
	}
	__antithesis_instrumentation__.Notify(266354)

	removeInactives := func(dropped bool) {
		__antithesis_instrumentation__.Notify(266367)
		t.mu.Lock()
		t.mu.takenOffline = dropped
		leases := t.removeInactiveVersions()
		t.mu.Unlock()
		for _, l := range leases {
			__antithesis_instrumentation__.Notify(266368)
			releaseLease(ctx, l, m)
		}
	}
	__antithesis_instrumentation__.Notify(266355)

	if dropped {
		__antithesis_instrumentation__.Notify(266369)
		removeInactives(true)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266370)
	}
	__antithesis_instrumentation__.Notify(266356)

	if err := ensureVersion(ctx, id, minVersion, m); err != nil {
		__antithesis_instrumentation__.Notify(266371)
		return err
	} else {
		__antithesis_instrumentation__.Notify(266372)
	}
	__antithesis_instrumentation__.Notify(266357)

	desc, _, err := t.findForTimestamp(ctx, m.storage.clock.Now())
	if isInactive := catalog.HasInactiveDescriptorError(err); err == nil || func() bool {
		__antithesis_instrumentation__.Notify(266373)
		return isInactive == true
	}() == true {
		__antithesis_instrumentation__.Notify(266374)
		removeInactives(isInactive)
		if desc != nil {
			__antithesis_instrumentation__.Notify(266376)
			t.release(ctx, desc)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(266377)
		}
		__antithesis_instrumentation__.Notify(266375)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266378)
	}
	__antithesis_instrumentation__.Notify(266358)
	return err
}

type AcquireBlockType int

const (
	AcquireBlock AcquireBlockType = iota

	AcquireFreshestBlock
)

type Manager struct {
	rangeFeedFactory *rangefeed.Factory
	storage          storage
	mu               struct {
		syncutil.Mutex

		descriptors map[descpb.ID]*descriptorState

		updatesResolvedTimestamp hlc.Timestamp
	}

	draining atomic.Value

	names        nameCache
	testingKnobs ManagerTestingKnobs
	ambientCtx   log.AmbientContext
	stopper      *stop.Stopper
	sem          *quotapool.IntPool
}

const leaseConcurrencyLimit = 5

func NewLeaseManager(
	ambientCtx log.AmbientContext,
	nodeIDContainer *base.SQLIDContainer,
	db *kv.DB,
	clock *hlc.Clock,
	internalExecutor sqlutil.InternalExecutor,
	settings *cluster.Settings,
	codec keys.SQLCodec,
	testingKnobs ManagerTestingKnobs,
	stopper *stop.Stopper,
	rangeFeedFactory *rangefeed.Factory,
) *Manager {
	__antithesis_instrumentation__.Notify(266379)
	lm := &Manager{
		storage: storage{
			nodeIDContainer:  nodeIDContainer,
			db:               db,
			clock:            clock,
			internalExecutor: internalExecutor,
			settings:         settings,
			codec:            codec,
			group:            &singleflight.Group{},
			testingKnobs:     testingKnobs.LeaseStoreTestingKnobs,
			outstandingLeases: metric.NewGauge(metric.Metadata{
				Name:        "sql.leases.active",
				Help:        "The number of outstanding SQL schema leases.",
				Measurement: "Outstanding leases",
				Unit:        metric.Unit_COUNT,
			}),
		},
		rangeFeedFactory: rangeFeedFactory,
		testingKnobs:     testingKnobs,
		names:            makeNameCache(),
		ambientCtx:       ambientCtx,
		stopper:          stopper,
		sem:              quotapool.NewIntPool("lease manager", leaseConcurrencyLimit),
	}
	lm.stopper.AddCloser(lm.sem.Closer("stopper"))
	lm.mu.descriptors = make(map[descpb.ID]*descriptorState)
	lm.mu.updatesResolvedTimestamp = db.Clock().Now()

	lm.draining.Store(false)
	return lm
}

func NameMatchesDescriptor(
	desc catalog.Descriptor, parentID descpb.ID, parentSchemaID descpb.ID, name string,
) bool {
	__antithesis_instrumentation__.Notify(266380)
	return desc.GetName() == name && func() bool {
		__antithesis_instrumentation__.Notify(266381)
		return desc.GetParentID() == parentID == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(266382)
		return desc.GetParentSchemaID() == parentSchemaID == true
	}() == true
}

func (m *Manager) findNewest(id descpb.ID) *descriptorVersionState {
	__antithesis_instrumentation__.Notify(266383)
	t := m.findDescriptorState(id, false)
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.mu.active.findNewest()
}

func (m *Manager) AcquireByName(
	ctx context.Context,
	timestamp hlc.Timestamp,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
) (LeasedDescriptor, error) {
	__antithesis_instrumentation__.Notify(266384)

	validateDescriptorForReturn := func(desc LeasedDescriptor) (LeasedDescriptor, error) {
		__antithesis_instrumentation__.Notify(266390)
		if desc.Underlying().Offline() {
			__antithesis_instrumentation__.Notify(266392)
			if err := catalog.FilterDescriptorState(
				desc.Underlying(), tree.CommonLookupFlags{},
			); err != nil {
				__antithesis_instrumentation__.Notify(266393)
				desc.Release(ctx)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(266394)
			}
		} else {
			__antithesis_instrumentation__.Notify(266395)
		}
		__antithesis_instrumentation__.Notify(266391)
		return desc, nil
	}
	__antithesis_instrumentation__.Notify(266385)

	descVersion := m.names.get(ctx, parentID, parentSchemaID, name, timestamp)
	if descVersion != nil {
		__antithesis_instrumentation__.Notify(266396)
		if descVersion.GetModificationTime().LessEq(timestamp) {
			__antithesis_instrumentation__.Notify(266399)
			expiration := descVersion.getExpiration()

			durationUntilExpiry := time.Duration(expiration.WallTime - timestamp.WallTime)
			if durationUntilExpiry < m.storage.leaseRenewalTimeout() {
				__antithesis_instrumentation__.Notify(266401)
				if t := m.findDescriptorState(descVersion.GetID(), false); t != nil {
					__antithesis_instrumentation__.Notify(266402)
					if err := t.maybeQueueLeaseRenewal(
						ctx, m, descVersion.GetID(), name); err != nil {
						__antithesis_instrumentation__.Notify(266403)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(266404)
					}
				} else {
					__antithesis_instrumentation__.Notify(266405)
				}
			} else {
				__antithesis_instrumentation__.Notify(266406)
			}
			__antithesis_instrumentation__.Notify(266400)
			return validateDescriptorForReturn(descVersion)
		} else {
			__antithesis_instrumentation__.Notify(266407)
		}
		__antithesis_instrumentation__.Notify(266397)

		descVersion.Release(ctx)

		leasedDesc, err := m.Acquire(ctx, timestamp, descVersion.GetID())
		if err != nil {
			__antithesis_instrumentation__.Notify(266408)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266409)
		}
		__antithesis_instrumentation__.Notify(266398)
		return validateDescriptorForReturn(leasedDesc)
	} else {
		__antithesis_instrumentation__.Notify(266410)
	}
	__antithesis_instrumentation__.Notify(266386)

	var err error
	id, err := m.resolveName(ctx, timestamp, parentID, parentSchemaID, name)
	if err != nil {
		__antithesis_instrumentation__.Notify(266411)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266412)
	}
	__antithesis_instrumentation__.Notify(266387)
	desc, err := m.Acquire(ctx, timestamp, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(266413)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266414)
	}
	__antithesis_instrumentation__.Notify(266388)
	if !NameMatchesDescriptor(desc.Underlying(), parentID, parentSchemaID, name) {
		__antithesis_instrumentation__.Notify(266415)

		desc.Release(ctx)
		if err := m.AcquireFreshestFromStore(ctx, id); err != nil {
			__antithesis_instrumentation__.Notify(266418)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266419)
		}
		__antithesis_instrumentation__.Notify(266416)
		desc, err = m.Acquire(ctx, timestamp, id)
		if err != nil {
			__antithesis_instrumentation__.Notify(266420)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266421)
		}
		__antithesis_instrumentation__.Notify(266417)
		if !NameMatchesDescriptor(desc.Underlying(), parentID, parentSchemaID, name) {
			__antithesis_instrumentation__.Notify(266422)

			desc.Release(ctx)
			return nil, catalog.ErrDescriptorNotFound
		} else {
			__antithesis_instrumentation__.Notify(266423)
		}
	} else {
		__antithesis_instrumentation__.Notify(266424)
	}
	__antithesis_instrumentation__.Notify(266389)
	return validateDescriptorForReturn(desc)
}

func (m *Manager) resolveName(
	ctx context.Context,
	timestamp hlc.Timestamp,
	parentID descpb.ID,
	parentSchemaID descpb.ID,
	name string,
) (id descpb.ID, _ error) {
	__antithesis_instrumentation__.Notify(266425)
	if err := m.storage.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(266428)

		if err := txn.SetUserPriority(roachpb.MaxUserPriority); err != nil {
			__antithesis_instrumentation__.Notify(266431)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266432)
		}
		__antithesis_instrumentation__.Notify(266429)
		if err := txn.SetFixedTimestamp(ctx, timestamp); err != nil {
			__antithesis_instrumentation__.Notify(266433)
			return err
		} else {
			__antithesis_instrumentation__.Notify(266434)
		}
		__antithesis_instrumentation__.Notify(266430)
		var err error
		id, err = catkv.LookupID(ctx, txn, m.storage.codec, parentID, parentSchemaID, name)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(266435)
		return id, err
	} else {
		__antithesis_instrumentation__.Notify(266436)
	}
	__antithesis_instrumentation__.Notify(266426)
	if id == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(266437)
		return id, catalog.ErrDescriptorNotFound
	} else {
		__antithesis_instrumentation__.Notify(266438)
	}
	__antithesis_instrumentation__.Notify(266427)
	return id, nil
}

type LeasedDescriptor interface {
	catalog.NameEntry

	Underlying() catalog.Descriptor

	Expiration() hlc.Timestamp

	Release(ctx context.Context)
}

func (m *Manager) Acquire(
	ctx context.Context, timestamp hlc.Timestamp, id descpb.ID,
) (LeasedDescriptor, error) {
	__antithesis_instrumentation__.Notify(266439)
	for {
		__antithesis_instrumentation__.Notify(266440)
		t := m.findDescriptorState(id, true)
		desc, latest, err := t.findForTimestamp(ctx, timestamp)
		if err == nil {
			__antithesis_instrumentation__.Notify(266442)

			if latest {
				__antithesis_instrumentation__.Notify(266444)
				durationUntilExpiry := time.Duration(desc.getExpiration().WallTime - timestamp.WallTime)
				if durationUntilExpiry < m.storage.leaseRenewalTimeout() {
					__antithesis_instrumentation__.Notify(266445)
					if err := t.maybeQueueLeaseRenewal(ctx, m, id, desc.GetName()); err != nil {
						__antithesis_instrumentation__.Notify(266446)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(266447)
					}
				} else {
					__antithesis_instrumentation__.Notify(266448)
				}
			} else {
				__antithesis_instrumentation__.Notify(266449)
			}
			__antithesis_instrumentation__.Notify(266443)
			return desc, nil
		} else {
			__antithesis_instrumentation__.Notify(266450)
		}
		__antithesis_instrumentation__.Notify(266441)
		switch {
		case errors.Is(err, errRenewLease):
			__antithesis_instrumentation__.Notify(266451)
			if err := func() error {
				__antithesis_instrumentation__.Notify(266456)
				t.markAcquisitionStart(ctx)
				defer t.markAcquisitionDone(ctx)

				_, errLease := acquireNodeLease(ctx, m, id)
				return errLease
			}(); err != nil {
				__antithesis_instrumentation__.Notify(266457)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(266458)
			}
			__antithesis_instrumentation__.Notify(266452)

			if m.testingKnobs.LeaseStoreTestingKnobs.LeaseAcquireResultBlockEvent != nil {
				__antithesis_instrumentation__.Notify(266459)
				m.testingKnobs.LeaseStoreTestingKnobs.LeaseAcquireResultBlockEvent(AcquireBlock, id)
			} else {
				__antithesis_instrumentation__.Notify(266460)
			}

		case errors.Is(err, errReadOlderVersion):
			__antithesis_instrumentation__.Notify(266453)

			versions, errRead := m.readOlderVersionForTimestamp(ctx, id, timestamp)
			if errRead != nil {
				__antithesis_instrumentation__.Notify(266461)
				return nil, errRead
			} else {
				__antithesis_instrumentation__.Notify(266462)
			}
			__antithesis_instrumentation__.Notify(266454)
			m.insertDescriptorVersions(id, versions)

		default:
			__antithesis_instrumentation__.Notify(266455)
			return nil, err
		}
	}
}

func (m *Manager) removeOnceDereferenced() bool {
	__antithesis_instrumentation__.Notify(266463)
	return m.storage.testingKnobs.RemoveOnceDereferenced || func() bool {
		__antithesis_instrumentation__.Notify(266464)
		return m.isDraining() == true
	}() == true
}

func (m *Manager) isDraining() bool {
	__antithesis_instrumentation__.Notify(266465)
	return m.draining.Load().(bool)
}

func (m *Manager) SetDraining(
	ctx context.Context, drain bool, reporter func(int, redact.SafeString),
) {
	__antithesis_instrumentation__.Notify(266466)
	m.draining.Store(drain)
	if !drain {
		__antithesis_instrumentation__.Notify(266468)
		return
	} else {
		__antithesis_instrumentation__.Notify(266469)
	}
	__antithesis_instrumentation__.Notify(266467)

	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.mu.descriptors {
		__antithesis_instrumentation__.Notify(266470)
		t.mu.Lock()
		leases := t.removeInactiveVersions()
		t.mu.Unlock()
		for _, l := range leases {
			__antithesis_instrumentation__.Notify(266472)
			releaseLease(ctx, l, m)
		}
		__antithesis_instrumentation__.Notify(266471)
		if reporter != nil {
			__antithesis_instrumentation__.Notify(266473)

			reporter(len(leases), "descriptor leases")
		} else {
			__antithesis_instrumentation__.Notify(266474)
		}
	}
}

func (m *Manager) findDescriptorState(id descpb.ID, create bool) *descriptorState {
	__antithesis_instrumentation__.Notify(266475)
	m.mu.Lock()
	defer m.mu.Unlock()
	t := m.mu.descriptors[id]
	if t == nil && func() bool {
		__antithesis_instrumentation__.Notify(266477)
		return create == true
	}() == true {
		__antithesis_instrumentation__.Notify(266478)
		t = &descriptorState{m: m, id: id, stopper: m.stopper}
		m.mu.descriptors[id] = t
	} else {
		__antithesis_instrumentation__.Notify(266479)
	}
	__antithesis_instrumentation__.Notify(266476)
	return t
}

func (m *Manager) RefreshLeases(ctx context.Context, s *stop.Stopper, db *kv.DB) {
	__antithesis_instrumentation__.Notify(266480)
	descUpdateCh := make(chan *descpb.Descriptor)
	m.watchForUpdates(ctx, descUpdateCh)
	_ = s.RunAsyncTask(ctx, "refresh-leases", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(266481)
		for {
			__antithesis_instrumentation__.Notify(266482)
			select {
			case desc := <-descUpdateCh:
				__antithesis_instrumentation__.Notify(266483)

				if desc == nil {
					__antithesis_instrumentation__.Notify(266489)
					continue
				} else {
					__antithesis_instrumentation__.Notify(266490)
				}
				__antithesis_instrumentation__.Notify(266484)

				if evFunc := m.testingKnobs.TestingDescriptorUpdateEvent; evFunc != nil {
					__antithesis_instrumentation__.Notify(266491)
					if err := evFunc(desc); err != nil {
						__antithesis_instrumentation__.Notify(266492)
						log.Infof(ctx, "skipping update of %v due to knob: %v",
							desc, err)
						continue
					} else {
						__antithesis_instrumentation__.Notify(266493)
					}
				} else {
					__antithesis_instrumentation__.Notify(266494)
				}
				__antithesis_instrumentation__.Notify(266485)

				id, version, name, state, _, err := descpb.GetDescriptorMetadata(desc)
				if err != nil {
					__antithesis_instrumentation__.Notify(266495)
					log.Fatalf(ctx, "invalid descriptor %v: %v", desc, err)
				} else {
					__antithesis_instrumentation__.Notify(266496)
				}
				__antithesis_instrumentation__.Notify(266486)
				dropped := state == descpb.DescriptorState_DROP

				log.VEventf(ctx, 2, "purging old version of descriptor %d@%d (dropped %v)",
					id, version, dropped)
				if err := purgeOldVersions(ctx, db, id, dropped, version, m); err != nil {
					__antithesis_instrumentation__.Notify(266497)
					log.Warningf(ctx, "error purging leases for descriptor %d(%s): %s",
						id, name, err)
				} else {
					__antithesis_instrumentation__.Notify(266498)
				}
				__antithesis_instrumentation__.Notify(266487)

				if evFunc := m.testingKnobs.TestingDescriptorRefreshedEvent; evFunc != nil {
					__antithesis_instrumentation__.Notify(266499)
					evFunc(desc)
				} else {
					__antithesis_instrumentation__.Notify(266500)
				}

			case <-s.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(266488)
				return
			}
		}
	})
}

func (m *Manager) watchForUpdates(ctx context.Context, descUpdateCh chan<- *descpb.Descriptor) {
	__antithesis_instrumentation__.Notify(266501)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(266504)
		log.Infof(ctx, "using rangefeeds for lease manager updates")
	} else {
		__antithesis_instrumentation__.Notify(266505)
	}
	__antithesis_instrumentation__.Notify(266502)
	descriptorTableStart := m.Codec().TablePrefix(keys.DescriptorTableID)
	descriptorTableSpan := roachpb.Span{
		Key:    descriptorTableStart,
		EndKey: descriptorTableStart.PrefixEnd(),
	}
	handleEvent := func(
		ctx context.Context, ev *roachpb.RangeFeedValue,
	) {
		__antithesis_instrumentation__.Notify(266506)
		if len(ev.Value.RawBytes) == 0 {
			__antithesis_instrumentation__.Notify(266512)
			return
		} else {
			__antithesis_instrumentation__.Notify(266513)
		}
		__antithesis_instrumentation__.Notify(266507)
		var descriptor descpb.Descriptor
		if err := ev.Value.GetProto(&descriptor); err != nil {
			__antithesis_instrumentation__.Notify(266514)
			logcrash.ReportOrPanic(ctx, &m.storage.settings.SV,
				"%s: unable to unmarshal descriptor %v", ev.Key, ev.Value)
			return
		} else {
			__antithesis_instrumentation__.Notify(266515)
		}
		__antithesis_instrumentation__.Notify(266508)
		if descriptor.Union == nil {
			__antithesis_instrumentation__.Notify(266516)
			return
		} else {
			__antithesis_instrumentation__.Notify(266517)
		}
		__antithesis_instrumentation__.Notify(266509)
		descpb.MaybeSetDescriptorModificationTimeFromMVCCTimestamp(&descriptor, ev.Value.Timestamp)
		id, version, name, _, _, err := descpb.GetDescriptorMetadata(&descriptor)
		if err != nil {
			__antithesis_instrumentation__.Notify(266518)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(266519)
		}
		__antithesis_instrumentation__.Notify(266510)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(266520)
			log.Infof(ctx, "%s: refreshing lease on descriptor: %d (%s), version: %d",
				ev.Key, id, name, version)
		} else {
			__antithesis_instrumentation__.Notify(266521)
		}
		__antithesis_instrumentation__.Notify(266511)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(266522)
		case descUpdateCh <- &descriptor:
			__antithesis_instrumentation__.Notify(266523)
		}
	}
	__antithesis_instrumentation__.Notify(266503)

	_, _ = m.rangeFeedFactory.RangeFeed(
		ctx, "lease", []roachpb.Span{descriptorTableSpan}, hlc.Timestamp{}, handleEvent,
	)
}

var leaseRefreshLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.tablecache.lease.refresh_limit",
	"maximum number of descriptors to periodically refresh leases for",
	500,
)

func (m *Manager) PeriodicallyRefreshSomeLeases(ctx context.Context) {
	__antithesis_instrumentation__.Notify(266524)
	_ = m.stopper.RunAsyncTask(ctx, "lease-refresher", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(266525)
		leaseDuration := LeaseDuration.Get(&m.storage.settings.SV)
		if leaseDuration <= 0 {
			__antithesis_instrumentation__.Notify(266527)
			return
		} else {
			__antithesis_instrumentation__.Notify(266528)
		}
		__antithesis_instrumentation__.Notify(266526)
		refreshTimer := timeutil.NewTimer()
		defer refreshTimer.Stop()
		refreshTimer.Reset(m.storage.jitteredLeaseDuration() / 2)
		for {
			__antithesis_instrumentation__.Notify(266529)
			select {
			case <-m.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(266530)
				return

			case <-refreshTimer.C:
				__antithesis_instrumentation__.Notify(266531)
				refreshTimer.Read = true
				refreshTimer.Reset(m.storage.jitteredLeaseDuration() / 2)

				m.refreshSomeLeases(ctx)
			}
		}
	})
}

func (m *Manager) refreshSomeLeases(ctx context.Context) {
	__antithesis_instrumentation__.Notify(266532)
	limit := leaseRefreshLimit.Get(&m.storage.settings.SV)
	if limit <= 0 {
		__antithesis_instrumentation__.Notify(266536)
		return
	} else {
		__antithesis_instrumentation__.Notify(266537)
	}
	__antithesis_instrumentation__.Notify(266533)

	m.mu.Lock()
	ids := make([]descpb.ID, 0, len(m.mu.descriptors))
	var i int64
	for k, desc := range m.mu.descriptors {
		__antithesis_instrumentation__.Notify(266538)
		if i++; i > limit {
			__antithesis_instrumentation__.Notify(266540)
			break
		} else {
			__antithesis_instrumentation__.Notify(266541)
		}
		__antithesis_instrumentation__.Notify(266539)
		desc.mu.Lock()
		takenOffline := desc.mu.takenOffline
		desc.mu.Unlock()
		if !takenOffline {
			__antithesis_instrumentation__.Notify(266542)
			ids = append(ids, k)
		} else {
			__antithesis_instrumentation__.Notify(266543)
		}
	}
	__antithesis_instrumentation__.Notify(266534)
	m.mu.Unlock()

	var wg sync.WaitGroup
	for i := range ids {
		__antithesis_instrumentation__.Notify(266544)
		id := ids[i]
		wg.Add(1)
		if err := m.stopper.RunAsyncTaskEx(
			ctx,
			stop.TaskOpts{
				TaskName:   fmt.Sprintf("refresh descriptor: %d lease", id),
				Sem:        m.sem,
				WaitForSem: true,
			},
			func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(266545)
				defer wg.Done()
				if _, err := acquireNodeLease(ctx, m, id); err != nil {
					__antithesis_instrumentation__.Notify(266546)
					log.Infof(ctx, "refreshing descriptor: %d lease failed: %s", id, err)
				} else {
					__antithesis_instrumentation__.Notify(266547)
				}
			}); err != nil {
			__antithesis_instrumentation__.Notify(266548)
			log.Infof(ctx, "didnt refresh descriptor: %d lease: %s", id, err)
			wg.Done()
		} else {
			__antithesis_instrumentation__.Notify(266549)
		}
	}
	__antithesis_instrumentation__.Notify(266535)
	wg.Wait()
}

func (m *Manager) DeleteOrphanedLeases(ctx context.Context, timeThreshold int64) {
	__antithesis_instrumentation__.Notify(266550)
	if m.testingKnobs.DisableDeleteOrphanedLeases {
		__antithesis_instrumentation__.Notify(266553)
		return
	} else {
		__antithesis_instrumentation__.Notify(266554)
	}
	__antithesis_instrumentation__.Notify(266551)

	nodeID := m.storage.nodeIDContainer.SQLInstanceID()
	if nodeID == 0 {
		__antithesis_instrumentation__.Notify(266555)
		panic("zero nodeID")
	} else {
		__antithesis_instrumentation__.Notify(266556)
	}
	__antithesis_instrumentation__.Notify(266552)

	newCtx := m.ambientCtx.AnnotateCtx(context.Background())

	newCtx = logtags.AddTags(newCtx, logtags.FromContext(ctx))
	_ = m.stopper.RunAsyncTask(newCtx, "del-orphaned-leases", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(266557)

		sqlQuery := fmt.Sprintf(`
SELECT "descID", version, expiration FROM system.public.lease AS OF SYSTEM TIME %d WHERE "nodeID" = %d
`, timeThreshold, nodeID)
		var rows []tree.Datums
		retryOptions := base.DefaultRetryOptions()
		retryOptions.Closer = m.stopper.ShouldQuiesce()

		if err := retry.WithMaxAttempts(ctx, retryOptions, 30, func() error {
			__antithesis_instrumentation__.Notify(266559)
			var err error
			rows, err = m.storage.internalExecutor.QueryBuffered(
				ctx, "read orphaned leases", nil, sqlQuery,
			)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(266560)
			log.Warningf(ctx, "unable to read orphaned leases: %+v", err)
			return
		} else {
			__antithesis_instrumentation__.Notify(266561)
		}
		__antithesis_instrumentation__.Notify(266558)

		var wg sync.WaitGroup
		defer wg.Wait()
		for i := range rows {
			__antithesis_instrumentation__.Notify(266562)

			row := rows[i]
			wg.Add(1)
			lease := storedLease{
				id:         descpb.ID(tree.MustBeDInt(row[0])),
				version:    int(tree.MustBeDInt(row[1])),
				expiration: tree.MustBeDTimestamp(row[2]),
			}
			if err := m.stopper.RunAsyncTaskEx(
				ctx,
				stop.TaskOpts{
					TaskName:   fmt.Sprintf("release lease %+v", lease),
					Sem:        m.sem,
					WaitForSem: true,
				},
				func(ctx context.Context) {
					__antithesis_instrumentation__.Notify(266563)
					m.storage.release(ctx, m.stopper, &lease)
					log.Infof(ctx, "released orphaned lease: %+v", lease)
					wg.Done()
				}); err != nil {
				__antithesis_instrumentation__.Notify(266564)
				wg.Done()
			} else {
				__antithesis_instrumentation__.Notify(266565)
			}
		}
	})
}

func (m *Manager) DB() *kv.DB {
	__antithesis_instrumentation__.Notify(266566)
	return m.storage.db
}

func (m *Manager) Codec() keys.SQLCodec {
	__antithesis_instrumentation__.Notify(266567)
	return m.storage.codec
}

type Metrics struct {
	OutstandingLeases *metric.Gauge
}

func (m *Manager) MetricsStruct() Metrics {
	__antithesis_instrumentation__.Notify(266568)
	return Metrics{
		OutstandingLeases: m.storage.outstandingLeases,
	}
}

func (m *Manager) VisitLeases(
	f func(desc catalog.Descriptor, takenOffline bool, refCount int, expiration tree.DTimestamp) (wantMore bool),
) {
	__antithesis_instrumentation__.Notify(266569)
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ts := range m.mu.descriptors {
		__antithesis_instrumentation__.Notify(266570)
		visitor := func() (wantMore bool) {
			__antithesis_instrumentation__.Notify(266572)
			ts.mu.Lock()
			defer ts.mu.Unlock()

			takenOffline := ts.mu.takenOffline

			for _, state := range ts.mu.active.data {
				__antithesis_instrumentation__.Notify(266574)
				state.mu.Lock()
				lease := state.mu.lease
				refCount := state.mu.refcount
				state.mu.Unlock()

				if lease == nil {
					__antithesis_instrumentation__.Notify(266576)
					continue
				} else {
					__antithesis_instrumentation__.Notify(266577)
				}
				__antithesis_instrumentation__.Notify(266575)

				if !f(state.Descriptor, takenOffline, refCount, lease.expiration) {
					__antithesis_instrumentation__.Notify(266578)
					return false
				} else {
					__antithesis_instrumentation__.Notify(266579)
				}
			}
			__antithesis_instrumentation__.Notify(266573)
			return true
		}
		__antithesis_instrumentation__.Notify(266571)
		if !visitor() {
			__antithesis_instrumentation__.Notify(266580)
			return
		} else {
			__antithesis_instrumentation__.Notify(266581)
		}
	}
}
