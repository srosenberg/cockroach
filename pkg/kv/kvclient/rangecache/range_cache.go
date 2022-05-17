package rangecache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/biogo/store/llrb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type rangeCacheKey roachpb.RKey

var minCacheKey interface{} = rangeCacheKey(roachpb.RKeyMin)

func (a rangeCacheKey) String() string {
	__antithesis_instrumentation__.Notify(89257)
	return roachpb.Key(a).String()
}

func (a rangeCacheKey) Compare(b llrb.Comparable) int {
	__antithesis_instrumentation__.Notify(89258)
	return bytes.Compare(a, b.(rangeCacheKey))
}

type RangeDescriptorDB interface {
	RangeLookup(
		ctx context.Context, key roachpb.RKey, useReverseScan bool,
	) ([]roachpb.RangeDescriptor, []roachpb.RangeDescriptor, error)

	FirstRange() (*roachpb.RangeDescriptor, error)
}

type RangeCache struct {
	st      *cluster.Settings
	stopper *stop.Stopper
	tracer  *tracing.Tracer

	db RangeDescriptorDB

	rangeCache struct {
		syncutil.RWMutex
		cache *cache.OrderedCache
	}

	lookupRequests singleflight.Group

	coalesced chan struct{}
}

func makeLookupRequestKey(
	key roachpb.RKey, prevDesc *roachpb.RangeDescriptor, useReverseScan bool,
) string {
	__antithesis_instrumentation__.Notify(89259)
	var ret strings.Builder

	if key.AsRawKey().Compare(keys.Meta1KeyMax) < 0 {
		__antithesis_instrumentation__.Notify(89263)
		ret.Write(keys.Meta1Prefix)
	} else {
		__antithesis_instrumentation__.Notify(89264)
		if key.AsRawKey().Compare(keys.Meta2KeyMax) < 0 {
			__antithesis_instrumentation__.Notify(89265)
			ret.Write(keys.Meta2Prefix)
		} else {
			__antithesis_instrumentation__.Notify(89266)
		}
	}
	__antithesis_instrumentation__.Notify(89260)
	if prevDesc != nil {
		__antithesis_instrumentation__.Notify(89267)
		if useReverseScan {
			__antithesis_instrumentation__.Notify(89268)
			key = prevDesc.EndKey
		} else {
			__antithesis_instrumentation__.Notify(89269)
			key = prevDesc.StartKey
		}
	} else {
		__antithesis_instrumentation__.Notify(89270)
	}
	__antithesis_instrumentation__.Notify(89261)
	ret.Write(key)
	ret.WriteString(":")
	ret.WriteString(strconv.FormatBool(useReverseScan))

	if prevDesc != nil {
		__antithesis_instrumentation__.Notify(89271)
		ret.WriteString(":")
		ret.WriteString(prevDesc.Generation.String())
	} else {
		__antithesis_instrumentation__.Notify(89272)
	}
	__antithesis_instrumentation__.Notify(89262)
	return ret.String()
}

func NewRangeCache(
	st *cluster.Settings,
	db RangeDescriptorDB,
	size func() int64,
	stopper *stop.Stopper,
	tracer *tracing.Tracer,
) *RangeCache {
	__antithesis_instrumentation__.Notify(89273)
	rdc := &RangeCache{st: st, db: db, stopper: stopper, tracer: tracer}
	rdc.rangeCache.cache = cache.NewOrderedCache(cache.Config{
		Policy: cache.CacheLRU,
		ShouldEvict: func(n int, _, _ interface{}) bool {
			__antithesis_instrumentation__.Notify(89275)
			return int64(n) > size()
		},
	})
	__antithesis_instrumentation__.Notify(89274)
	return rdc
}

func (rc *RangeCache) String() string {
	__antithesis_instrumentation__.Notify(89276)
	rc.rangeCache.RLock()
	defer rc.rangeCache.RUnlock()
	return rc.stringLocked()
}

func (rc *RangeCache) stringLocked() string {
	__antithesis_instrumentation__.Notify(89277)
	var buf strings.Builder
	rc.rangeCache.cache.Do(func(k, v interface{}) bool {
		__antithesis_instrumentation__.Notify(89279)
		fmt.Fprintf(&buf, "key=%s desc=%+v\n", roachpb.Key(k.(rangeCacheKey)), v)
		return false
	})
	__antithesis_instrumentation__.Notify(89278)
	return buf.String()
}

type EvictionToken struct {
	rdc *RangeCache

	desc     *roachpb.RangeDescriptor
	lease    *roachpb.Lease
	closedts roachpb.RangeClosedTimestampPolicy

	speculativeDesc *roachpb.RangeDescriptor
}

func (rc *RangeCache) makeEvictionToken(
	entry *CacheEntry, speculativeDesc *roachpb.RangeDescriptor,
) EvictionToken {
	__antithesis_instrumentation__.Notify(89280)
	if speculativeDesc != nil {
		__antithesis_instrumentation__.Notify(89282)

		nextCpy := *speculativeDesc
		nextCpy.Generation = 0
		speculativeDesc = &nextCpy
	} else {
		__antithesis_instrumentation__.Notify(89283)
	}
	__antithesis_instrumentation__.Notify(89281)
	return EvictionToken{
		rdc:             rc,
		desc:            entry.Desc(),
		lease:           entry.leaseEvenIfSpeculative(),
		closedts:        entry.closedts,
		speculativeDesc: speculativeDesc,
	}
}

func (rc *RangeCache) MakeEvictionToken(entry *CacheEntry) EvictionToken {
	__antithesis_instrumentation__.Notify(89284)
	return rc.makeEvictionToken(entry, nil)
}

func (et EvictionToken) String() string {
	__antithesis_instrumentation__.Notify(89285)
	if !et.Valid() {
		__antithesis_instrumentation__.Notify(89287)
		return "<empty>"
	} else {
		__antithesis_instrumentation__.Notify(89288)
	}
	__antithesis_instrumentation__.Notify(89286)
	return fmt.Sprintf("desc:%s lease:%s spec desc: %v", et.desc, et.lease, et.speculativeDesc)
}

func (et EvictionToken) Valid() bool {
	__antithesis_instrumentation__.Notify(89289)
	return et.rdc != nil
}

func (et *EvictionToken) clear() {
	__antithesis_instrumentation__.Notify(89290)
	*et = EvictionToken{}
}

func (et EvictionToken) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(89291)
	if !et.Valid() {
		__antithesis_instrumentation__.Notify(89293)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89294)
	}
	__antithesis_instrumentation__.Notify(89292)
	return et.desc
}

func (et EvictionToken) Leaseholder() *roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(89295)
	if !et.Valid() || func() bool {
		__antithesis_instrumentation__.Notify(89297)
		return et.lease == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(89298)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89299)
	}
	__antithesis_instrumentation__.Notify(89296)
	return &et.lease.Replica
}

func (et EvictionToken) LeaseSeq() roachpb.LeaseSequence {
	__antithesis_instrumentation__.Notify(89300)
	if !et.Valid() {
		__antithesis_instrumentation__.Notify(89303)
		panic("invalid LeaseSeq() call on empty EvictionToken")
	} else {
		__antithesis_instrumentation__.Notify(89304)
	}
	__antithesis_instrumentation__.Notify(89301)
	if et.lease == nil {
		__antithesis_instrumentation__.Notify(89305)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(89306)
	}
	__antithesis_instrumentation__.Notify(89302)
	return et.lease.Sequence
}

func (et EvictionToken) ClosedTimestampPolicy() roachpb.RangeClosedTimestampPolicy {
	__antithesis_instrumentation__.Notify(89307)
	if !et.Valid() {
		__antithesis_instrumentation__.Notify(89309)
		panic("invalid ClosedTimestampPolicy() call on empty EvictionToken")
	} else {
		__antithesis_instrumentation__.Notify(89310)
	}
	__antithesis_instrumentation__.Notify(89308)
	return et.closedts
}

func (et *EvictionToken) syncRLocked(
	ctx context.Context,
) (stillValid bool, cachedEntry *CacheEntry, rawEntry *cache.Entry) {
	__antithesis_instrumentation__.Notify(89311)
	cachedEntry, rawEntry = et.rdc.getCachedRLocked(ctx, et.desc.StartKey, false)
	if cachedEntry == nil || func() bool {
		__antithesis_instrumentation__.Notify(89313)
		return !descsCompatible(cachedEntry.Desc(), et.Desc()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(89314)
		et.clear()
		return false, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(89315)
	}
	__antithesis_instrumentation__.Notify(89312)
	et.desc = cachedEntry.Desc()
	et.lease = cachedEntry.leaseEvenIfSpeculative()
	return true, cachedEntry, rawEntry
}

func (et *EvictionToken) UpdateLease(
	ctx context.Context, l *roachpb.Lease, descGeneration roachpb.RangeGeneration,
) bool {
	__antithesis_instrumentation__.Notify(89316)
	rdc := et.rdc
	rdc.rangeCache.Lock()
	defer rdc.rangeCache.Unlock()

	stillValid, cachedEntry, rawEntry := et.syncRLocked(ctx)
	if !stillValid {
		__antithesis_instrumentation__.Notify(89320)
		return false
	} else {
		__antithesis_instrumentation__.Notify(89321)
	}
	__antithesis_instrumentation__.Notify(89317)
	ok, newEntry := cachedEntry.updateLease(l, descGeneration)
	if !ok {
		__antithesis_instrumentation__.Notify(89322)
		return false
	} else {
		__antithesis_instrumentation__.Notify(89323)
	}
	__antithesis_instrumentation__.Notify(89318)
	if newEntry != nil {
		__antithesis_instrumentation__.Notify(89324)
		et.desc = newEntry.Desc()
		et.lease = newEntry.leaseEvenIfSpeculative()
	} else {
		__antithesis_instrumentation__.Notify(89325)

		et.clear()
	}
	__antithesis_instrumentation__.Notify(89319)
	rdc.swapEntryLocked(ctx, rawEntry, newEntry)
	return newEntry != nil
}

func (et *EvictionToken) UpdateLeaseholder(
	ctx context.Context, lh roachpb.ReplicaDescriptor, descGeneration roachpb.RangeGeneration,
) {
	__antithesis_instrumentation__.Notify(89326)

	l := &roachpb.Lease{Replica: lh}
	et.UpdateLease(ctx, l, descGeneration)
}

func (et *EvictionToken) EvictLease(ctx context.Context) {
	__antithesis_instrumentation__.Notify(89327)
	et.rdc.rangeCache.Lock()
	defer et.rdc.rangeCache.Unlock()

	if et.lease == nil {
		__antithesis_instrumentation__.Notify(89331)
		log.Fatalf(ctx, "attempting to clear lease from cache entry without lease")
	} else {
		__antithesis_instrumentation__.Notify(89332)
	}
	__antithesis_instrumentation__.Notify(89328)

	lh := et.lease.Replica
	stillValid, cachedEntry, rawEntry := et.syncRLocked(ctx)
	if !stillValid {
		__antithesis_instrumentation__.Notify(89333)
		return
	} else {
		__antithesis_instrumentation__.Notify(89334)
	}
	__antithesis_instrumentation__.Notify(89329)
	ok, newEntry := cachedEntry.evictLeaseholder(lh)
	if !ok {
		__antithesis_instrumentation__.Notify(89335)
		return
	} else {
		__antithesis_instrumentation__.Notify(89336)
	}
	__antithesis_instrumentation__.Notify(89330)
	et.desc = newEntry.Desc()
	et.lease = newEntry.leaseEvenIfSpeculative()
	et.rdc.swapEntryLocked(ctx, rawEntry, newEntry)
}

func descsCompatible(a, b *roachpb.RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(89337)
	return (a.RangeID == b.RangeID) && func() bool {
		__antithesis_instrumentation__.Notify(89338)
		return (a.RSpan().Equal(b.RSpan())) == true
	}() == true
}

func (et *EvictionToken) Evict(ctx context.Context) {
	__antithesis_instrumentation__.Notify(89339)
	et.EvictAndReplace(ctx)
}

func (et *EvictionToken) EvictAndReplace(ctx context.Context, newDescs ...roachpb.RangeInfo) {
	__antithesis_instrumentation__.Notify(89340)
	if !et.Valid() {
		__antithesis_instrumentation__.Notify(89343)
		panic("trying to evict an invalid token")
	} else {
		__antithesis_instrumentation__.Notify(89344)
	}
	__antithesis_instrumentation__.Notify(89341)

	et.rdc.rangeCache.Lock()
	defer et.rdc.rangeCache.Unlock()

	et.rdc.evictDescLocked(ctx, et.Desc())

	if len(newDescs) > 0 {
		__antithesis_instrumentation__.Notify(89345)
		log.Eventf(ctx, "evicting cached range descriptor with %d replacements", len(newDescs))
		et.rdc.insertLocked(ctx, newDescs...)
	} else {
		__antithesis_instrumentation__.Notify(89346)
		if et.speculativeDesc != nil {
			__antithesis_instrumentation__.Notify(89347)
			log.Eventf(ctx, "evicting cached range descriptor with replacement from token")
			et.rdc.insertLocked(ctx, roachpb.RangeInfo{
				Desc: *et.speculativeDesc,

				Lease: roachpb.Lease{},

				ClosedTimestampPolicy: et.closedts,
			})
		} else {
			__antithesis_instrumentation__.Notify(89348)
			log.Eventf(ctx, "evicting cached range descriptor")
		}
	}
	__antithesis_instrumentation__.Notify(89342)
	et.clear()
}

func (rc *RangeCache) LookupWithEvictionToken(
	ctx context.Context, key roachpb.RKey, evictToken EvictionToken, useReverseScan bool,
) (EvictionToken, error) {
	__antithesis_instrumentation__.Notify(89349)
	tok, err := rc.lookupInternal(ctx, key, evictToken, useReverseScan)
	if err != nil {
		__antithesis_instrumentation__.Notify(89351)
		return EvictionToken{}, err
	} else {
		__antithesis_instrumentation__.Notify(89352)
	}
	__antithesis_instrumentation__.Notify(89350)
	return tok, nil
}

func (rc *RangeCache) Lookup(ctx context.Context, key roachpb.RKey) (CacheEntry, error) {
	__antithesis_instrumentation__.Notify(89353)
	tok, err := rc.lookupInternal(
		ctx, key, EvictionToken{}, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(89357)
		return CacheEntry{}, err
	} else {
		__antithesis_instrumentation__.Notify(89358)
	}
	__antithesis_instrumentation__.Notify(89354)
	var e CacheEntry
	if tok.desc != nil {
		__antithesis_instrumentation__.Notify(89359)
		e.desc = *tok.desc
	} else {
		__antithesis_instrumentation__.Notify(89360)
	}
	__antithesis_instrumentation__.Notify(89355)
	if tok.lease != nil {
		__antithesis_instrumentation__.Notify(89361)
		e.lease = *tok.lease
	} else {
		__antithesis_instrumentation__.Notify(89362)
	}
	__antithesis_instrumentation__.Notify(89356)
	e.closedts = tok.closedts
	return e, nil
}

func (rc *RangeCache) GetCachedOverlapping(ctx context.Context, span roachpb.RSpan) []*CacheEntry {
	__antithesis_instrumentation__.Notify(89363)
	rc.rangeCache.RLock()
	defer rc.rangeCache.RUnlock()
	rawEntries := rc.getCachedOverlappingRLocked(ctx, span)
	entries := make([]*CacheEntry, len(rawEntries))
	for i, e := range rawEntries {
		__antithesis_instrumentation__.Notify(89365)
		entries[i] = rc.getValue(e)
	}
	__antithesis_instrumentation__.Notify(89364)
	return entries
}

func (rc *RangeCache) getCachedOverlappingRLocked(
	ctx context.Context, span roachpb.RSpan,
) []*cache.Entry {
	__antithesis_instrumentation__.Notify(89366)
	var res []*cache.Entry
	rc.rangeCache.cache.DoRangeReverseEntry(func(e *cache.Entry) (exit bool) {
		__antithesis_instrumentation__.Notify(89369)
		desc := rc.getValue(e).Desc()
		if desc.StartKey.Equal(span.EndKey) {
			__antithesis_instrumentation__.Notify(89372)

			return false
		} else {
			__antithesis_instrumentation__.Notify(89373)
		}
		__antithesis_instrumentation__.Notify(89370)

		if desc.EndKey.Compare(span.Key) <= 0 {
			__antithesis_instrumentation__.Notify(89374)
			return true
		} else {
			__antithesis_instrumentation__.Notify(89375)
		}
		__antithesis_instrumentation__.Notify(89371)
		res = append(res, e)
		return false
	}, rangeCacheKey(span.EndKey), minCacheKey)
	__antithesis_instrumentation__.Notify(89367)

	for i, j := 0, len(res)-1; i < j; i, j = i+1, j-1 {
		__antithesis_instrumentation__.Notify(89376)
		res[i], res[j] = res[j], res[i]
	}
	__antithesis_instrumentation__.Notify(89368)
	return res
}

func (rc *RangeCache) lookupInternal(
	ctx context.Context, key roachpb.RKey, evictToken EvictionToken, useReverseScan bool,
) (EvictionToken, error) {
	__antithesis_instrumentation__.Notify(89377)

	for {
		__antithesis_instrumentation__.Notify(89378)
		newToken, err := rc.tryLookup(ctx, key, evictToken, useReverseScan)
		if errors.HasType(err, (lookupCoalescingError{})) {
			__antithesis_instrumentation__.Notify(89381)
			log.VEventf(ctx, 2, "bad lookup coalescing; retrying: %s", err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(89382)
		}
		__antithesis_instrumentation__.Notify(89379)
		if err != nil {
			__antithesis_instrumentation__.Notify(89383)
			return EvictionToken{}, err
		} else {
			__antithesis_instrumentation__.Notify(89384)
		}
		__antithesis_instrumentation__.Notify(89380)
		return newToken, nil
	}
}

type lookupCoalescingError struct {
	key       roachpb.RKey
	wrongDesc *roachpb.RangeDescriptor
}

func (e lookupCoalescingError) Error() string {
	__antithesis_instrumentation__.Notify(89385)
	return fmt.Sprintf("key %q not contained in range lookup's "+
		"resulting descriptor %v", e.key, e.wrongDesc)
}

func newLookupCoalescingError(key roachpb.RKey, wrongDesc *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(89386)
	return lookupCoalescingError{
		key:       key,
		wrongDesc: wrongDesc,
	}
}

func (rc *RangeCache) tryLookup(
	ctx context.Context, key roachpb.RKey, evictToken EvictionToken, useReverseScan bool,
) (EvictionToken, error) {
	__antithesis_instrumentation__.Notify(89387)
	rc.rangeCache.RLock()
	if entry, _ := rc.getCachedRLocked(ctx, key, useReverseScan); entry != nil {
		__antithesis_instrumentation__.Notify(89398)
		rc.rangeCache.RUnlock()
		returnToken := rc.makeEvictionToken(entry, nil)
		return returnToken, nil
	} else {
		__antithesis_instrumentation__.Notify(89399)
	}
	__antithesis_instrumentation__.Notify(89388)

	log.VEventf(ctx, 2, "looking up range descriptor: key=%s", key)

	var prevDesc *roachpb.RangeDescriptor
	if evictToken.Valid() {
		__antithesis_instrumentation__.Notify(89400)
		prevDesc = evictToken.Desc()
	} else {
		__antithesis_instrumentation__.Notify(89401)
	}
	__antithesis_instrumentation__.Notify(89389)
	requestKey := makeLookupRequestKey(key, prevDesc, useReverseScan)

	reqCtx, reqSpan := tracing.EnsureChildSpan(ctx, rc.tracer, "range lookup")
	resC, leader := rc.lookupRequests.DoChan(requestKey, func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(89402)
		defer reqSpan.Finish()
		var lookupRes EvictionToken
		if err := rc.stopper.RunTaskWithErr(reqCtx, "rangecache: range lookup", func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(89404)

			ctx, cancel := rc.stopper.WithCancelOnQuiesce(
				logtags.WithTags(context.Background(), logtags.FromContext(ctx)))
			defer cancel()
			ctx = tracing.ContextWithSpan(ctx, reqSpan)

			var rs, preRs []roachpb.RangeDescriptor
			if err := contextutil.RunWithTimeout(ctx, "range lookup", 10*time.Second,
				func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(89410)
					var err error
					rs, preRs, err = rc.performRangeLookup(ctx, key, useReverseScan)
					return err
				}); err != nil {
				__antithesis_instrumentation__.Notify(89411)
				return err
			} else {
				__antithesis_instrumentation__.Notify(89412)
			}
			__antithesis_instrumentation__.Notify(89405)

			switch {
			case len(rs) == 0:
				__antithesis_instrumentation__.Notify(89413)
				return fmt.Errorf("no range descriptors returned for %s", key)
			case len(rs) > 2:
				__antithesis_instrumentation__.Notify(89414)
				panic(fmt.Sprintf("more than 2 matching range descriptors returned for %s: %v", key, rs))
			default:
				__antithesis_instrumentation__.Notify(89415)
			}
			__antithesis_instrumentation__.Notify(89406)

			rc.rangeCache.Lock()
			defer rc.rangeCache.Unlock()

			newEntries := make([]*CacheEntry, len(preRs)+1)
			newEntries[0] = &CacheEntry{
				desc: rs[0],

				lease: roachpb.Lease{},

				closedts: roachpb.LAG_BY_CLUSTER_SETTING,
			}
			for i, preR := range preRs {
				__antithesis_instrumentation__.Notify(89416)
				newEntries[i+1] = &CacheEntry{desc: preR}
			}
			__antithesis_instrumentation__.Notify(89407)
			insertedEntries := rc.insertLockedInner(ctx, newEntries)

			entry := insertedEntries[0]

			if entry == nil {
				__antithesis_instrumentation__.Notify(89417)
				entry = &CacheEntry{
					desc:     rs[0],
					lease:    roachpb.Lease{},
					closedts: roachpb.LAG_BY_CLUSTER_SETTING,
				}
			} else {
				__antithesis_instrumentation__.Notify(89418)
			}
			__antithesis_instrumentation__.Notify(89408)
			if len(rs) == 1 {
				__antithesis_instrumentation__.Notify(89419)
				lookupRes = rc.makeEvictionToken(entry, nil)
			} else {
				__antithesis_instrumentation__.Notify(89420)
				lookupRes = rc.makeEvictionToken(entry, &rs[1])
			}
			__antithesis_instrumentation__.Notify(89409)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(89421)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(89422)
		}
		__antithesis_instrumentation__.Notify(89403)
		return lookupRes, nil
	})
	__antithesis_instrumentation__.Notify(89390)

	rc.rangeCache.RUnlock()

	if !leader {
		__antithesis_instrumentation__.Notify(89423)
		log.VEvent(ctx, 2, "coalesced range lookup request onto in-flight one")
		if rc.coalesced != nil {
			__antithesis_instrumentation__.Notify(89425)
			rc.coalesced <- struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(89426)
		}
		__antithesis_instrumentation__.Notify(89424)

		reqSpan.Finish()
	} else {
		__antithesis_instrumentation__.Notify(89427)
	}
	__antithesis_instrumentation__.Notify(89391)

	var res singleflight.Result
	select {
	case res = <-resC:
		__antithesis_instrumentation__.Notify(89428)
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(89429)
		return EvictionToken{}, errors.Wrap(ctx.Err(), "aborted during range descriptor lookup")
	}
	__antithesis_instrumentation__.Notify(89392)

	var s string
	if res.Err != nil {
		__antithesis_instrumentation__.Notify(89430)
		s = res.Err.Error()
	} else {
		__antithesis_instrumentation__.Notify(89431)
		s = res.Val.(EvictionToken).String()
	}
	__antithesis_instrumentation__.Notify(89393)
	if res.Shared {
		__antithesis_instrumentation__.Notify(89432)
		log.VEventf(ctx, 2, "looked up range descriptor with shared request: %s", s)
	} else {
		__antithesis_instrumentation__.Notify(89433)
		log.VEventf(ctx, 2, "looked up range descriptor: %s", s)
	}
	__antithesis_instrumentation__.Notify(89394)
	if res.Err != nil {
		__antithesis_instrumentation__.Notify(89434)
		return EvictionToken{}, res.Err
	} else {
		__antithesis_instrumentation__.Notify(89435)
	}
	__antithesis_instrumentation__.Notify(89395)

	lookupRes := res.Val.(EvictionToken)
	desc := lookupRes.Desc()
	containsFn := (*roachpb.RangeDescriptor).ContainsKey
	if useReverseScan {
		__antithesis_instrumentation__.Notify(89436)
		containsFn = (*roachpb.RangeDescriptor).ContainsKeyInverted
	} else {
		__antithesis_instrumentation__.Notify(89437)
	}
	__antithesis_instrumentation__.Notify(89396)
	if !containsFn(desc, key) {
		__antithesis_instrumentation__.Notify(89438)
		return EvictionToken{}, newLookupCoalescingError(key, desc)
	} else {
		__antithesis_instrumentation__.Notify(89439)
	}
	__antithesis_instrumentation__.Notify(89397)
	return lookupRes, nil
}

func (rc *RangeCache) performRangeLookup(
	ctx context.Context, key roachpb.RKey, useReverseScan bool,
) ([]roachpb.RangeDescriptor, []roachpb.RangeDescriptor, error) {
	__antithesis_instrumentation__.Notify(89440)

	ctx = logtags.AddTag(ctx, "range-lookup", key)

	if keys.RangeMetaKey(key).Equal(roachpb.RKeyMin) {
		__antithesis_instrumentation__.Notify(89442)
		desc, err := rc.db.FirstRange()
		if err != nil {
			__antithesis_instrumentation__.Notify(89444)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(89445)
		}
		__antithesis_instrumentation__.Notify(89443)
		return []roachpb.RangeDescriptor{*desc}, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(89446)
	}
	__antithesis_instrumentation__.Notify(89441)

	return rc.db.RangeLookup(ctx, key, useReverseScan)
}

func (rc *RangeCache) Clear() {
	__antithesis_instrumentation__.Notify(89447)
	rc.rangeCache.Lock()
	defer rc.rangeCache.Unlock()
	rc.rangeCache.cache.Clear()
}

func (rc *RangeCache) EvictByKey(ctx context.Context, descKey roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(89448)
	rc.rangeCache.Lock()
	defer rc.rangeCache.Unlock()

	cachedDesc, entry := rc.getCachedRLocked(ctx, descKey, false)
	if cachedDesc == nil {
		__antithesis_instrumentation__.Notify(89450)
		return false
	} else {
		__antithesis_instrumentation__.Notify(89451)
	}
	__antithesis_instrumentation__.Notify(89449)
	log.VEventf(ctx, 2, "evict cached descriptor: %s", cachedDesc)
	rc.rangeCache.cache.DelEntry(entry)
	return true
}

func (rc *RangeCache) evictDescLocked(ctx context.Context, desc *roachpb.RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(89452)
	cachedEntry, rawEntry := rc.getCachedRLocked(ctx, desc.StartKey, false)
	if cachedEntry == nil {
		__antithesis_instrumentation__.Notify(89455)

		return false
	} else {
		__antithesis_instrumentation__.Notify(89456)
	}
	__antithesis_instrumentation__.Notify(89453)
	cachedDesc := cachedEntry.Desc()
	cachedNewer := cachedDesc.Generation > desc.Generation
	if cachedNewer {
		__antithesis_instrumentation__.Notify(89457)
		return false
	} else {
		__antithesis_instrumentation__.Notify(89458)
	}
	__antithesis_instrumentation__.Notify(89454)

	log.VEventf(ctx, 2, "evict cached descriptor: desc=%s", cachedEntry)
	rc.rangeCache.cache.DelEntry(rawEntry)
	return true
}

func (rc *RangeCache) GetCached(ctx context.Context, key roachpb.RKey, inverted bool) *CacheEntry {
	__antithesis_instrumentation__.Notify(89459)
	rc.rangeCache.RLock()
	defer rc.rangeCache.RUnlock()
	entry, _ := rc.getCachedRLocked(ctx, key, inverted)
	return entry
}

func (rc *RangeCache) getCachedRLocked(
	ctx context.Context, key roachpb.RKey, inverted bool,
) (*CacheEntry, *cache.Entry) {
	__antithesis_instrumentation__.Notify(89460)

	var rawEntry *cache.Entry
	if !inverted {
		__antithesis_instrumentation__.Notify(89465)
		var ok bool
		rawEntry, ok = rc.rangeCache.cache.FloorEntry(rangeCacheKey(key))
		if !ok {
			__antithesis_instrumentation__.Notify(89466)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(89467)
		}
	} else {
		__antithesis_instrumentation__.Notify(89468)
		rc.rangeCache.cache.DoRangeReverseEntry(func(e *cache.Entry) bool {
			__antithesis_instrumentation__.Notify(89470)
			startKey := roachpb.RKey(e.Key.(rangeCacheKey))
			if key.Equal(startKey) {
				__antithesis_instrumentation__.Notify(89472)

				return false
			} else {
				__antithesis_instrumentation__.Notify(89473)
			}
			__antithesis_instrumentation__.Notify(89471)
			rawEntry = e
			return true
		}, rangeCacheKey(key), minCacheKey)
		__antithesis_instrumentation__.Notify(89469)

		if rawEntry == nil {
			__antithesis_instrumentation__.Notify(89474)
			rawEntry, _ = rc.rangeCache.cache.FloorEntry(minCacheKey)
		} else {
			__antithesis_instrumentation__.Notify(89475)
		}
	}
	__antithesis_instrumentation__.Notify(89461)

	if rawEntry == nil {
		__antithesis_instrumentation__.Notify(89476)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(89477)
	}
	__antithesis_instrumentation__.Notify(89462)
	entry := rc.getValue(rawEntry)

	containsFn := (*roachpb.RangeDescriptor).ContainsKey
	if inverted {
		__antithesis_instrumentation__.Notify(89478)
		containsFn = (*roachpb.RangeDescriptor).ContainsKeyInverted
	} else {
		__antithesis_instrumentation__.Notify(89479)
	}
	__antithesis_instrumentation__.Notify(89463)

	if !containsFn(entry.Desc(), key) {
		__antithesis_instrumentation__.Notify(89480)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(89481)
	}
	__antithesis_instrumentation__.Notify(89464)
	return entry, rawEntry
}

func (rc *RangeCache) Insert(ctx context.Context, rs ...roachpb.RangeInfo) {
	__antithesis_instrumentation__.Notify(89482)
	rc.rangeCache.Lock()
	defer rc.rangeCache.Unlock()
	rc.insertLocked(ctx, rs...)
}

func (rc *RangeCache) insertLocked(ctx context.Context, rs ...roachpb.RangeInfo) []*CacheEntry {
	__antithesis_instrumentation__.Notify(89483)
	entries := make([]*CacheEntry, len(rs))
	for i, r := range rs {
		__antithesis_instrumentation__.Notify(89485)
		entries[i] = &CacheEntry{
			desc:     r.Desc,
			lease:    r.Lease,
			closedts: r.ClosedTimestampPolicy,
		}
	}
	__antithesis_instrumentation__.Notify(89484)
	return rc.insertLockedInner(ctx, entries)
}

func (rc *RangeCache) insertLockedInner(ctx context.Context, rs []*CacheEntry) []*CacheEntry {
	__antithesis_instrumentation__.Notify(89486)

	entries := make([]*CacheEntry, len(rs))
	for i, ent := range rs {
		__antithesis_instrumentation__.Notify(89488)
		if !ent.desc.IsInitialized() {
			__antithesis_instrumentation__.Notify(89493)
			log.Fatalf(ctx, "inserting uninitialized desc: %s", ent)
		} else {
			__antithesis_instrumentation__.Notify(89494)
		}
		__antithesis_instrumentation__.Notify(89489)
		if !ent.lease.Empty() {
			__antithesis_instrumentation__.Notify(89495)
			replID := ent.lease.Replica.ReplicaID
			_, ok := ent.desc.GetReplicaDescriptorByID(replID)
			if !ok {
				__antithesis_instrumentation__.Notify(89496)
				log.Fatalf(ctx, "leaseholder replicaID: %d not part of descriptor: %s. lease: %s",
					replID, ent.Desc(), ent.Lease())
			} else {
				__antithesis_instrumentation__.Notify(89497)
			}
		} else {
			__antithesis_instrumentation__.Notify(89498)
		}
		__antithesis_instrumentation__.Notify(89490)

		ok, newerEntry := rc.clearOlderOverlappingLocked(ctx, ent)
		if !ok {
			__antithesis_instrumentation__.Notify(89499)

			entries[i] = newerEntry
			continue
		} else {
			__antithesis_instrumentation__.Notify(89500)
		}
		__antithesis_instrumentation__.Notify(89491)
		rangeKey := ent.Desc().StartKey
		if log.V(2) {
			__antithesis_instrumentation__.Notify(89501)
			log.Infof(ctx, "adding cache entry: value=%s", ent)
		} else {
			__antithesis_instrumentation__.Notify(89502)
		}
		__antithesis_instrumentation__.Notify(89492)
		rc.rangeCache.cache.Add(rangeCacheKey(rangeKey), ent)
		entries[i] = ent
	}
	__antithesis_instrumentation__.Notify(89487)
	return entries
}

func (rc *RangeCache) getValue(entry *cache.Entry) *CacheEntry {
	__antithesis_instrumentation__.Notify(89503)
	return entry.Value.(*CacheEntry)
}

func (rc *RangeCache) clearOlderOverlapping(
	ctx context.Context, newEntry *CacheEntry,
) (ok bool, newerEntry *CacheEntry) {
	__antithesis_instrumentation__.Notify(89504)
	rc.rangeCache.Lock()
	defer rc.rangeCache.Unlock()
	return rc.clearOlderOverlappingLocked(ctx, newEntry)
}

func (rc *RangeCache) clearOlderOverlappingLocked(
	ctx context.Context, newEntry *CacheEntry,
) (ok bool, newerEntry *CacheEntry) {
	__antithesis_instrumentation__.Notify(89505)
	log.VEventf(ctx, 2, "clearing entries overlapping %s", newEntry.Desc())
	newest := true
	var newerFound *CacheEntry
	overlapping := rc.getCachedOverlappingRLocked(ctx, newEntry.Desc().RSpan())
	for _, e := range overlapping {
		__antithesis_instrumentation__.Notify(89507)
		entry := rc.getValue(e)
		if newEntry.overrides(entry) {
			__antithesis_instrumentation__.Notify(89508)
			if log.V(2) {
				__antithesis_instrumentation__.Notify(89510)
				log.Infof(ctx, "clearing overlapping descriptor: key=%s entry=%s", e.Key, rc.getValue(e))
			} else {
				__antithesis_instrumentation__.Notify(89511)
			}
			__antithesis_instrumentation__.Notify(89509)
			rc.rangeCache.cache.DelEntry(e)
		} else {
			__antithesis_instrumentation__.Notify(89512)
			newest = false
			if descsCompatible(entry.Desc(), newEntry.Desc()) {
				__antithesis_instrumentation__.Notify(89513)
				newerFound = entry

				if len(overlapping) != 1 {
					__antithesis_instrumentation__.Notify(89514)
					log.Errorf(ctx, "%s", errors.AssertionFailedf(
						"found compatible descriptor but also got multiple overlapping results. newEntry: %s. overlapping: %s",
						newEntry, overlapping).Error())
				} else {
					__antithesis_instrumentation__.Notify(89515)
				}
			} else {
				__antithesis_instrumentation__.Notify(89516)
			}
		}
	}
	__antithesis_instrumentation__.Notify(89506)
	return newest, newerFound
}

func (rc *RangeCache) swapEntryLocked(
	ctx context.Context, oldEntry *cache.Entry, newEntry *CacheEntry,
) {
	__antithesis_instrumentation__.Notify(89517)
	if newEntry != nil {
		__antithesis_instrumentation__.Notify(89519)
		old := rc.getValue(oldEntry)
		if !descsCompatible(old.Desc(), newEntry.Desc()) {
			__antithesis_instrumentation__.Notify(89520)
			log.Fatalf(ctx, "attempting to swap non-compatible descs: %s vs %s",
				old, newEntry)
		} else {
			__antithesis_instrumentation__.Notify(89521)
		}
	} else {
		__antithesis_instrumentation__.Notify(89522)
	}
	__antithesis_instrumentation__.Notify(89518)

	rc.rangeCache.cache.DelEntry(oldEntry)
	if newEntry != nil {
		__antithesis_instrumentation__.Notify(89523)
		log.VEventf(ctx, 2, "caching new entry: %s", newEntry)
		rc.rangeCache.cache.Add(oldEntry.Key, newEntry)
	} else {
		__antithesis_instrumentation__.Notify(89524)
	}
}

func (rc *RangeCache) DB() RangeDescriptorDB {
	__antithesis_instrumentation__.Notify(89525)
	return rc.db
}

func (rc *RangeCache) TestingSetDB(db RangeDescriptorDB) {
	__antithesis_instrumentation__.Notify(89526)
	rc.db = db
}

func (rc *RangeCache) NumInFlight(name string) int {
	__antithesis_instrumentation__.Notify(89527)
	return rc.lookupRequests.NumCalls(name)
}

type CacheEntry struct {
	desc roachpb.RangeDescriptor

	lease roachpb.Lease

	closedts roachpb.RangeClosedTimestampPolicy
}

func (e CacheEntry) String() string {
	__antithesis_instrumentation__.Notify(89528)
	return fmt.Sprintf("desc:%s, lease:%s", e.Desc(), e.lease)
}

func (e *CacheEntry) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(89529)
	return &e.desc
}

func (e *CacheEntry) Leaseholder() *roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(89530)
	if e.lease.Empty() {
		__antithesis_instrumentation__.Notify(89532)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89533)
	}
	__antithesis_instrumentation__.Notify(89531)
	return &e.lease.Replica
}

func (e *CacheEntry) Lease() *roachpb.Lease {
	__antithesis_instrumentation__.Notify(89534)
	if e.lease.Empty() {
		__antithesis_instrumentation__.Notify(89537)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89538)
	}
	__antithesis_instrumentation__.Notify(89535)
	if e.LeaseSpeculative() {
		__antithesis_instrumentation__.Notify(89539)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89540)
	}
	__antithesis_instrumentation__.Notify(89536)
	return &e.lease
}

func (e *CacheEntry) leaseEvenIfSpeculative() *roachpb.Lease {
	__antithesis_instrumentation__.Notify(89541)
	if e.lease.Empty() {
		__antithesis_instrumentation__.Notify(89543)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(89544)
	}
	__antithesis_instrumentation__.Notify(89542)
	return &e.lease
}

func (e *CacheEntry) ClosedTimestampPolicy() roachpb.RangeClosedTimestampPolicy {
	__antithesis_instrumentation__.Notify(89545)
	return e.closedts
}

func (e *CacheEntry) DescSpeculative() bool {
	__antithesis_instrumentation__.Notify(89546)
	return e.desc.Generation == 0
}

func (e *CacheEntry) LeaseSpeculative() bool {
	__antithesis_instrumentation__.Notify(89547)
	if e.lease.Empty() {
		__antithesis_instrumentation__.Notify(89549)
		panic(fmt.Sprintf("LeaseSpeculative called on entry with empty lease: %s", e))
	} else {
		__antithesis_instrumentation__.Notify(89550)
	}
	__antithesis_instrumentation__.Notify(89548)
	return e.lease.Speculative()
}

func (e *CacheEntry) overrides(o *CacheEntry) bool {
	__antithesis_instrumentation__.Notify(89551)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(89556)
		if _, err := e.Desc().RSpan().Intersect(o.Desc()); err != nil {
			__antithesis_instrumentation__.Notify(89557)
			panic(fmt.Sprintf("descriptors don't intersect: %s vs %s", e.Desc(), o.Desc()))
		} else {
			__antithesis_instrumentation__.Notify(89558)
		}
	} else {
		__antithesis_instrumentation__.Notify(89559)
	}
	__antithesis_instrumentation__.Notify(89552)

	if res := compareEntryDescs(o, e); res != 0 {
		__antithesis_instrumentation__.Notify(89560)
		return res < 0
	} else {
		__antithesis_instrumentation__.Notify(89561)
	}
	__antithesis_instrumentation__.Notify(89553)

	if e.Desc().RangeID != o.Desc().RangeID {
		__antithesis_instrumentation__.Notify(89562)
		panic(fmt.Sprintf("overlapping descriptors with same gen but different IDs: %s vs %s",
			e.Desc(), o.Desc()))
	} else {
		__antithesis_instrumentation__.Notify(89563)
	}
	__antithesis_instrumentation__.Notify(89554)

	if res := compareEntryLeases(o, e); res != 0 {
		__antithesis_instrumentation__.Notify(89564)
		return res < 0
	} else {
		__antithesis_instrumentation__.Notify(89565)
	}
	__antithesis_instrumentation__.Notify(89555)

	return o.closedts != e.closedts
}

func compareEntryDescs(a, b *CacheEntry) int {
	__antithesis_instrumentation__.Notify(89566)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(89572)
		if _, err := a.Desc().RSpan().Intersect(b.Desc()); err != nil {
			__antithesis_instrumentation__.Notify(89573)
			panic(fmt.Sprintf("descriptors don't intersect: %s vs %s", a.Desc(), b.Desc()))
		} else {
			__antithesis_instrumentation__.Notify(89574)
		}
	} else {
		__antithesis_instrumentation__.Notify(89575)
	}
	__antithesis_instrumentation__.Notify(89567)

	if a.desc.Equal(&b.desc) {
		__antithesis_instrumentation__.Notify(89576)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(89577)
	}
	__antithesis_instrumentation__.Notify(89568)

	if a.DescSpeculative() || func() bool {
		__antithesis_instrumentation__.Notify(89578)
		return b.DescSpeculative() == true
	}() == true {
		__antithesis_instrumentation__.Notify(89579)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(89580)
	}
	__antithesis_instrumentation__.Notify(89569)

	if a.Desc().Generation < b.Desc().Generation {
		__antithesis_instrumentation__.Notify(89581)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(89582)
	}
	__antithesis_instrumentation__.Notify(89570)
	if a.Desc().Generation > b.Desc().Generation {
		__antithesis_instrumentation__.Notify(89583)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(89584)
	}
	__antithesis_instrumentation__.Notify(89571)
	return 0
}

func compareEntryLeases(a, b *CacheEntry) int {
	__antithesis_instrumentation__.Notify(89585)
	if aEmpty, bEmpty := a.lease.Empty(), b.lease.Empty(); aEmpty || func() bool {
		__antithesis_instrumentation__.Notify(89590)
		return bEmpty == true
	}() == true {
		__antithesis_instrumentation__.Notify(89591)
		if aEmpty && func() bool {
			__antithesis_instrumentation__.Notify(89594)
			return !bEmpty == true
		}() == true {
			__antithesis_instrumentation__.Notify(89595)
			return -1
		} else {
			__antithesis_instrumentation__.Notify(89596)
		}
		__antithesis_instrumentation__.Notify(89592)
		if !aEmpty && func() bool {
			__antithesis_instrumentation__.Notify(89597)
			return bEmpty == true
		}() == true {
			__antithesis_instrumentation__.Notify(89598)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(89599)
		}
		__antithesis_instrumentation__.Notify(89593)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(89600)
	}
	__antithesis_instrumentation__.Notify(89586)

	if a.LeaseSpeculative() || func() bool {
		__antithesis_instrumentation__.Notify(89601)
		return b.LeaseSpeculative() == true
	}() == true {
		__antithesis_instrumentation__.Notify(89602)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(89603)
	}
	__antithesis_instrumentation__.Notify(89587)

	if a.Lease().Sequence < b.Lease().Sequence {
		__antithesis_instrumentation__.Notify(89604)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(89605)
	}
	__antithesis_instrumentation__.Notify(89588)
	if a.Lease().Sequence > b.Lease().Sequence {
		__antithesis_instrumentation__.Notify(89606)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(89607)
	}
	__antithesis_instrumentation__.Notify(89589)
	return 0
}

func (e *CacheEntry) updateLease(
	l *roachpb.Lease, descGeneration roachpb.RangeGeneration,
) (updated bool, newEntry *CacheEntry) {
	__antithesis_instrumentation__.Notify(89608)

	if l.Sequence != 0 && func() bool {
		__antithesis_instrumentation__.Notify(89612)
		return e.lease.Sequence != 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(89613)
		return l.Sequence < e.lease.Sequence == true
	}() == true {
		__antithesis_instrumentation__.Notify(89614)
		return false, e
	} else {
		__antithesis_instrumentation__.Notify(89615)
	}
	__antithesis_instrumentation__.Notify(89609)

	if l.Equal(e.Lease()) {
		__antithesis_instrumentation__.Notify(89616)
		return false, e
	} else {
		__antithesis_instrumentation__.Notify(89617)
	}
	__antithesis_instrumentation__.Notify(89610)

	_, ok := e.desc.GetReplicaDescriptorByID(l.Replica.ReplicaID)
	if !ok {
		__antithesis_instrumentation__.Notify(89618)

		if descGeneration != 0 && func() bool {
			__antithesis_instrumentation__.Notify(89620)
			return descGeneration < e.desc.Generation == true
		}() == true {
			__antithesis_instrumentation__.Notify(89621)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(89622)
		}
		__antithesis_instrumentation__.Notify(89619)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(89623)
	}
	__antithesis_instrumentation__.Notify(89611)

	return true, &CacheEntry{
		desc:     e.desc,
		lease:    *l,
		closedts: e.closedts,
	}
}

func (e *CacheEntry) evictLeaseholder(
	lh roachpb.ReplicaDescriptor,
) (updated bool, newEntry *CacheEntry) {
	__antithesis_instrumentation__.Notify(89624)
	if e.lease.Replica != lh {
		__antithesis_instrumentation__.Notify(89626)
		return false, e
	} else {
		__antithesis_instrumentation__.Notify(89627)
	}
	__antithesis_instrumentation__.Notify(89625)
	return true, &CacheEntry{
		desc:     e.desc,
		closedts: e.closedts,
	}
}

func IsRangeLookupErrorRetryable(err error) bool {
	__antithesis_instrumentation__.Notify(89628)

	return !grpcutil.IsAuthError(err)
}
