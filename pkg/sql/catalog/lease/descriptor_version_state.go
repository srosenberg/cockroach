package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type storedLease struct {
	id         descpb.ID
	version    int
	expiration tree.DTimestamp
}

func (s *storedLease) String() string {
	__antithesis_instrumentation__.Notify(266177)
	return fmt.Sprintf("ID = %d ver=%d expiration=%s", s.id, s.version, s.expiration)
}

type descriptorVersionState struct {
	t *descriptorState

	catalog.Descriptor

	mu struct {
		syncutil.Mutex

		expiration hlc.Timestamp

		refcount int

		lease *storedLease
	}
}

func (s *descriptorVersionState) Release(ctx context.Context) {
	__antithesis_instrumentation__.Notify(266178)
	s.t.release(ctx, s)
}

func (s *descriptorVersionState) Underlying() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(266179)
	return s.Descriptor
}

func (s *descriptorVersionState) Expiration() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(266180)
	return s.getExpiration()
}

func (s *descriptorVersionState) SafeMessage() string {
	__antithesis_instrumentation__.Notify(266181)
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("%d ver=%d:%s, refcount=%d", s.GetID(), s.GetVersion(), s.mu.expiration, s.mu.refcount)
}

func (s *descriptorVersionState) String() string {
	__antithesis_instrumentation__.Notify(266182)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stringLocked()
}

func (s *descriptorVersionState) stringLocked() string {
	__antithesis_instrumentation__.Notify(266183)
	return fmt.Sprintf("%d(%q) ver=%d:%s, refcount=%d", s.GetID(), s.GetName(), s.GetVersion(), s.mu.expiration, s.mu.refcount)
}

func (s *descriptorVersionState) hasExpired(timestamp hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(266184)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hasExpiredLocked(timestamp)
}

func (s *descriptorVersionState) hasExpiredLocked(timestamp hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(266185)
	return s.mu.expiration.LessEq(timestamp)
}

func (s *descriptorVersionState) incRefCount(ctx context.Context, expensiveLogEnabled bool) {
	__antithesis_instrumentation__.Notify(266186)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.incRefCountLocked(ctx, expensiveLogEnabled)
}

func (s *descriptorVersionState) incRefCountLocked(ctx context.Context, expensiveLogEnabled bool) {
	__antithesis_instrumentation__.Notify(266187)
	s.mu.refcount++
	if expensiveLogEnabled {
		__antithesis_instrumentation__.Notify(266188)
		log.VEventf(ctx, 2, "descriptorVersionState.incRefCount: %s", s.stringLocked())
	} else {
		__antithesis_instrumentation__.Notify(266189)
	}
}

func (s *descriptorVersionState) getExpiration() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(266190)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.expiration
}

func storedLeaseExpiration(expiration hlc.Timestamp) tree.DTimestamp {
	__antithesis_instrumentation__.Notify(266191)
	return tree.DTimestamp{Time: timeutil.Unix(0, expiration.WallTime).Round(time.Microsecond)}
}
