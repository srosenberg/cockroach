package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/gc"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type cachedProtectedTimestampState struct {
	readAt                      hlc.Timestamp
	earliestProtectionTimestamp hlc.Timestamp
}

func (ts *cachedProtectedTimestampState) clearIfNotNewer(existing cachedProtectedTimestampState) {
	__antithesis_instrumentation__.Notify(118332)
	if !existing.readAt.Less(ts.readAt) {
		__antithesis_instrumentation__.Notify(118333)
		*ts = cachedProtectedTimestampState{}
	} else {
		__antithesis_instrumentation__.Notify(118334)
	}
}

func (r *Replica) maybeUpdateCachedProtectedTS(ts *cachedProtectedTimestampState) {
	__antithesis_instrumentation__.Notify(118335)
	if *ts == (cachedProtectedTimestampState{}) {
		__antithesis_instrumentation__.Notify(118337)
		return
	} else {
		__antithesis_instrumentation__.Notify(118338)
	}
	__antithesis_instrumentation__.Notify(118336)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mu.cachedProtectedTS.readAt.Less(ts.readAt) {
		__antithesis_instrumentation__.Notify(118339)
		r.mu.cachedProtectedTS = *ts
	} else {
		__antithesis_instrumentation__.Notify(118340)
	}
}

func (r *Replica) readProtectedTimestampsRLocked(
	ctx context.Context,
) (ts cachedProtectedTimestampState, _ error) {
	__antithesis_instrumentation__.Notify(118341)
	desc := r.descRLocked()
	gcThreshold := *r.mu.state.GCThreshold

	sp := roachpb.Span{
		Key:    roachpb.Key(desc.StartKey),
		EndKey: roachpb.Key(desc.EndKey),
	}
	var protectionTimestamps []hlc.Timestamp
	var err error
	protectionTimestamps, ts.readAt, err = r.store.protectedtsReader.GetProtectionTimestamps(ctx, sp)
	if err != nil {
		__antithesis_instrumentation__.Notify(118344)
		return ts, err
	} else {
		__antithesis_instrumentation__.Notify(118345)
	}
	__antithesis_instrumentation__.Notify(118342)
	earliestTS := hlc.Timestamp{}
	for _, protectionTimestamp := range protectionTimestamps {
		__antithesis_instrumentation__.Notify(118346)

		if isValid := gcThreshold.LessEq(protectionTimestamp); !isValid {
			__antithesis_instrumentation__.Notify(118348)
			continue
		} else {
			__antithesis_instrumentation__.Notify(118349)
		}
		__antithesis_instrumentation__.Notify(118347)

		log.VEventf(ctx, 2, "span: %s has a protection policy protecting: %s",
			sp.String(), protectionTimestamp.String())

		if earliestTS.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(118350)
			return protectionTimestamp.Less(earliestTS) == true
		}() == true {
			__antithesis_instrumentation__.Notify(118351)
			earliestTS = protectionTimestamp
		} else {
			__antithesis_instrumentation__.Notify(118352)
		}
	}
	__antithesis_instrumentation__.Notify(118343)
	ts.earliestProtectionTimestamp = earliestTS
	return ts, nil
}

func (r *Replica) checkProtectedTimestampsForGC(
	ctx context.Context, gcTTL time.Duration,
) (canGC bool, cacheTimestamp, gcTimestamp, oldThreshold, newThreshold hlc.Timestamp, _ error) {
	__antithesis_instrumentation__.Notify(118353)

	var read cachedProtectedTimestampState
	defer r.maybeUpdateCachedProtectedTS(&read)
	r.mu.RLock()
	defer r.mu.RUnlock()
	defer read.clearIfNotNewer(r.mu.cachedProtectedTS)

	oldThreshold = *r.mu.state.GCThreshold
	lease := *r.mu.state.Lease

	var err error
	read, err = r.readProtectedTimestampsRLocked(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(118358)
		return false, hlc.Timestamp{}, hlc.Timestamp{}, hlc.Timestamp{}, hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(118359)
	}
	__antithesis_instrumentation__.Notify(118354)

	if read.readAt.IsEmpty() {
		__antithesis_instrumentation__.Notify(118360)

		log.VEventf(ctx, 1,
			"not gc'ing replica %v because protected timestamp information is unavailable", r)
		return false, hlc.Timestamp{}, hlc.Timestamp{}, hlc.Timestamp{}, hlc.Timestamp{}, nil
	} else {
		__antithesis_instrumentation__.Notify(118361)
	}
	__antithesis_instrumentation__.Notify(118355)

	gcTimestamp = read.readAt
	if !read.earliestProtectionTimestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(118362)

		impliedGCTimestamp := gc.TimestampForThreshold(read.earliestProtectionTimestamp.Prev(), gcTTL)
		if impliedGCTimestamp.Less(gcTimestamp) {
			__antithesis_instrumentation__.Notify(118363)
			gcTimestamp = impliedGCTimestamp
		} else {
			__antithesis_instrumentation__.Notify(118364)
		}
	} else {
		__antithesis_instrumentation__.Notify(118365)
	}
	__antithesis_instrumentation__.Notify(118356)

	if gcTimestamp.Less(lease.Start.ToTimestamp()) {
		__antithesis_instrumentation__.Notify(118366)
		log.VEventf(ctx, 1, "not gc'ing replica %v due to new lease %v started after %v",
			r, lease, gcTimestamp)
		return false, hlc.Timestamp{}, hlc.Timestamp{}, hlc.Timestamp{}, hlc.Timestamp{}, nil
	} else {
		__antithesis_instrumentation__.Notify(118367)
	}
	__antithesis_instrumentation__.Notify(118357)

	newThreshold = gc.CalculateThreshold(gcTimestamp, gcTTL)

	return true, read.readAt, gcTimestamp, oldThreshold, newThreshold, nil
}

func (r *Replica) markPendingGC(readAt, newThreshold hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(118368)
	r.protectedTimestampMu.Lock()
	defer r.protectedTimestampMu.Unlock()
	if readAt.Less(r.protectedTimestampMu.minStateReadTimestamp) {
		__antithesis_instrumentation__.Notify(118370)
		return errors.Errorf("cannot set gc threshold to %v because read at %v < min %v",
			newThreshold, readAt, r.protectedTimestampMu.minStateReadTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(118371)
	}
	__antithesis_instrumentation__.Notify(118369)
	r.protectedTimestampMu.pendingGCThreshold = newThreshold
	return nil
}
