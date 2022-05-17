package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
)

type descriptorState struct {
	m       *Manager
	id      descpb.ID
	stopper *stop.Stopper

	renewalInProgress int32

	mu struct {
		syncutil.Mutex

		active descriptorSet

		takenOffline bool

		maxVersionSeen descpb.DescriptorVersion

		acquisitionsInProgress int
	}
}

func (t *descriptorState) findForTimestamp(
	ctx context.Context, timestamp hlc.Timestamp,
) (*descriptorVersionState, bool, error) {
	__antithesis_instrumentation__.Notify(266092)
	expensiveLogEnabled := log.ExpensiveLogEnabled(ctx, 2)
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.mu.active.data) == 0 {
		__antithesis_instrumentation__.Notify(266095)
		return nil, false, errRenewLease
	} else {
		__antithesis_instrumentation__.Notify(266096)
	}
	__antithesis_instrumentation__.Notify(266093)

	for i := len(t.mu.active.data) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(266097)

		if desc := t.mu.active.data[i]; desc.GetModificationTime().LessEq(timestamp) {
			__antithesis_instrumentation__.Notify(266098)
			latest := i+1 == len(t.mu.active.data)
			if !desc.hasExpired(timestamp) {
				__antithesis_instrumentation__.Notify(266101)

				desc.incRefCount(ctx, expensiveLogEnabled)
				return desc, latest, nil
			} else {
				__antithesis_instrumentation__.Notify(266102)
			}
			__antithesis_instrumentation__.Notify(266099)

			if latest {
				__antithesis_instrumentation__.Notify(266103)

				return nil, false, errRenewLease
			} else {
				__antithesis_instrumentation__.Notify(266104)
			}
			__antithesis_instrumentation__.Notify(266100)
			break
		} else {
			__antithesis_instrumentation__.Notify(266105)
		}
	}
	__antithesis_instrumentation__.Notify(266094)

	return nil, false, errReadOlderVersion
}

func (t *descriptorState) upsertLeaseLocked(
	ctx context.Context, desc catalog.Descriptor, expiration hlc.Timestamp,
) (createdDescriptorVersionState *descriptorVersionState, toRelease *storedLease, _ error) {
	__antithesis_instrumentation__.Notify(266106)
	if t.mu.maxVersionSeen < desc.GetVersion() {
		__antithesis_instrumentation__.Notify(266111)
		t.mu.maxVersionSeen = desc.GetVersion()
	} else {
		__antithesis_instrumentation__.Notify(266112)
	}
	__antithesis_instrumentation__.Notify(266107)
	s := t.mu.active.find(desc.GetVersion())
	if s == nil {
		__antithesis_instrumentation__.Notify(266113)
		if t.mu.active.findNewest() != nil {
			__antithesis_instrumentation__.Notify(266115)
			log.Infof(ctx, "new lease: %s", desc)
		} else {
			__antithesis_instrumentation__.Notify(266116)
		}
		__antithesis_instrumentation__.Notify(266114)
		descState := newDescriptorVersionState(t, desc, expiration, true)
		t.mu.active.insert(descState)
		return descState, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(266117)
	}
	__antithesis_instrumentation__.Notify(266108)

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.mu.expiration.Less(expiration) {
		__antithesis_instrumentation__.Notify(266118)

		return nil, nil, errors.AssertionFailedf("lease expiration monotonicity violation, (%s) vs (%s)", s, desc)
	} else {
		__antithesis_instrumentation__.Notify(266119)
	}
	__antithesis_instrumentation__.Notify(266109)

	s.mu.expiration = expiration
	toRelease = s.mu.lease
	s.mu.lease = &storedLease{
		id:         desc.GetID(),
		version:    int(desc.GetVersion()),
		expiration: storedLeaseExpiration(expiration),
	}
	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(266120)
		log.VEventf(ctx, 2, "replaced lease: %s with %s", toRelease, s.mu.lease)
	} else {
		__antithesis_instrumentation__.Notify(266121)
	}
	__antithesis_instrumentation__.Notify(266110)
	return nil, toRelease, nil
}

var _ redact.SafeMessager = (*descriptorVersionState)(nil)

func newDescriptorVersionState(
	t *descriptorState, desc catalog.Descriptor, expiration hlc.Timestamp, isLease bool,
) *descriptorVersionState {
	__antithesis_instrumentation__.Notify(266122)
	descState := &descriptorVersionState{
		t:          t,
		Descriptor: desc,
	}
	descState.mu.expiration = expiration
	if isLease {
		__antithesis_instrumentation__.Notify(266124)
		descState.mu.lease = &storedLease{
			id:         desc.GetID(),
			version:    int(desc.GetVersion()),
			expiration: storedLeaseExpiration(expiration),
		}
	} else {
		__antithesis_instrumentation__.Notify(266125)
	}
	__antithesis_instrumentation__.Notify(266123)
	return descState
}

func (t *descriptorState) removeInactiveVersions() []*storedLease {
	__antithesis_instrumentation__.Notify(266126)
	var leases []*storedLease

	for _, desc := range append([]*descriptorVersionState(nil), t.mu.active.data...) {
		__antithesis_instrumentation__.Notify(266128)
		func() {
			__antithesis_instrumentation__.Notify(266129)
			desc.mu.Lock()
			defer desc.mu.Unlock()
			if desc.mu.refcount == 0 {
				__antithesis_instrumentation__.Notify(266130)
				t.mu.active.remove(desc)
				if l := desc.mu.lease; l != nil {
					__antithesis_instrumentation__.Notify(266131)
					desc.mu.lease = nil
					leases = append(leases, l)
				} else {
					__antithesis_instrumentation__.Notify(266132)
				}
			} else {
				__antithesis_instrumentation__.Notify(266133)
			}
		}()
	}
	__antithesis_instrumentation__.Notify(266127)
	return leases
}

func (t *descriptorState) release(ctx context.Context, s *descriptorVersionState) {
	__antithesis_instrumentation__.Notify(266134)

	expensiveLoggingEnabled := log.ExpensiveLogEnabled(ctx, 2)
	decRefCount := func(s *descriptorVersionState) (shouldRemove bool) {
		__antithesis_instrumentation__.Notify(266138)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.mu.refcount--
		if expensiveLoggingEnabled {
			__antithesis_instrumentation__.Notify(266140)
			log.Infof(ctx, "release: %s", s.stringLocked())
		} else {
			__antithesis_instrumentation__.Notify(266141)
		}
		__antithesis_instrumentation__.Notify(266139)
		return s.mu.refcount == 0
	}
	__antithesis_instrumentation__.Notify(266135)
	maybeMarkRemoveStoredLease := func(s *descriptorVersionState) *storedLease {
		__antithesis_instrumentation__.Notify(266142)

		removeOnceDereferenced :=

			t.mu.takenOffline || func() bool {
				__antithesis_instrumentation__.Notify(266146)
				return s != t.mu.active.findNewest() == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(266147)
				return s.GetVersion() < t.mu.maxVersionSeen == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(266148)
				return t.m.removeOnceDereferenced() == true
			}() == true
		if !removeOnceDereferenced {
			__antithesis_instrumentation__.Notify(266149)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(266150)
		}
		__antithesis_instrumentation__.Notify(266143)
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.mu.refcount < 0 {
			__antithesis_instrumentation__.Notify(266151)
			panic(errors.AssertionFailedf("negative ref count: %s", s))
		} else {
			__antithesis_instrumentation__.Notify(266152)
		}
		__antithesis_instrumentation__.Notify(266144)
		if s.mu.refcount == 0 && func() bool {
			__antithesis_instrumentation__.Notify(266153)
			return s.mu.lease != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(266154)
			return removeOnceDereferenced == true
		}() == true {
			__antithesis_instrumentation__.Notify(266155)
			l := s.mu.lease
			s.mu.lease = nil
			return l
		} else {
			__antithesis_instrumentation__.Notify(266156)
		}
		__antithesis_instrumentation__.Notify(266145)
		return nil
	}
	__antithesis_instrumentation__.Notify(266136)
	maybeRemoveLease := func() *storedLease {
		__antithesis_instrumentation__.Notify(266157)
		if shouldRemove := decRefCount(s); !shouldRemove {
			__antithesis_instrumentation__.Notify(266160)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(266161)
		}
		__antithesis_instrumentation__.Notify(266158)
		t.mu.Lock()
		defer t.mu.Unlock()
		if l := maybeMarkRemoveStoredLease(s); l != nil {
			__antithesis_instrumentation__.Notify(266162)
			t.mu.active.remove(s)
			return l
		} else {
			__antithesis_instrumentation__.Notify(266163)
		}
		__antithesis_instrumentation__.Notify(266159)
		return nil
	}
	__antithesis_instrumentation__.Notify(266137)
	if l := maybeRemoveLease(); l != nil {
		__antithesis_instrumentation__.Notify(266164)
		releaseLease(ctx, l, t.m)
	} else {
		__antithesis_instrumentation__.Notify(266165)
	}
}

func (t *descriptorState) maybeQueueLeaseRenewal(
	ctx context.Context, m *Manager, id descpb.ID, name string,
) error {
	__antithesis_instrumentation__.Notify(266166)
	if !atomic.CompareAndSwapInt32(&t.renewalInProgress, 0, 1) {
		__antithesis_instrumentation__.Notify(266168)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266169)
	}
	__antithesis_instrumentation__.Notify(266167)

	newCtx := m.ambientCtx.AnnotateCtx(context.Background())

	newCtx = logtags.AddTags(newCtx, logtags.FromContext(ctx))
	return t.stopper.RunAsyncTask(newCtx,
		"lease renewal", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(266170)
			t.startLeaseRenewal(ctx, m, id, name)
		})
}

func (t *descriptorState) startLeaseRenewal(
	ctx context.Context, m *Manager, id descpb.ID, name string,
) {
	__antithesis_instrumentation__.Notify(266171)
	log.VEventf(ctx, 1,
		"background lease renewal beginning for id=%d name=%q",
		id, name)
	if _, err := acquireNodeLease(ctx, m, id); err != nil {
		__antithesis_instrumentation__.Notify(266173)
		log.Errorf(ctx,
			"background lease renewal for id=%d name=%q failed: %s",
			id, name, err)
	} else {
		__antithesis_instrumentation__.Notify(266174)
		log.VEventf(ctx, 1,
			"background lease renewal finished for id=%d name=%q",
			id, name)
	}
	__antithesis_instrumentation__.Notify(266172)
	atomic.StoreInt32(&t.renewalInProgress, 0)
}

func (t *descriptorState) markAcquisitionStart(ctx context.Context) {
	__antithesis_instrumentation__.Notify(266175)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mu.acquisitionsInProgress++
}

func (t *descriptorState) markAcquisitionDone(ctx context.Context) {
	__antithesis_instrumentation__.Notify(266176)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mu.acquisitionsInProgress--
}
