package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

type RangeIterator struct {
	ds      *DistSender
	scanDir ScanDirection
	key     roachpb.RKey

	token rangecache.EvictionToken
	init  bool
	err   error
}

func MakeRangeIterator(ds *DistSender) RangeIterator {
	__antithesis_instrumentation__.Notify(87825)
	return RangeIterator{
		ds: ds,
	}
}

type ScanDirection byte

const (
	Ascending ScanDirection = iota

	Descending
)

func (ri *RangeIterator) Key() roachpb.RKey {
	__antithesis_instrumentation__.Notify(87826)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87828)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87829)
	}
	__antithesis_instrumentation__.Notify(87827)
	return ri.key
}

func (ri *RangeIterator) Desc() *roachpb.RangeDescriptor {
	__antithesis_instrumentation__.Notify(87830)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87832)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87833)
	}
	__antithesis_instrumentation__.Notify(87831)
	return ri.token.Desc()
}

func (ri *RangeIterator) Leaseholder() *roachpb.ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(87834)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87836)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87837)
	}
	__antithesis_instrumentation__.Notify(87835)
	return ri.token.Leaseholder()
}

func (ri *RangeIterator) ClosedTimestampPolicy() roachpb.RangeClosedTimestampPolicy {
	__antithesis_instrumentation__.Notify(87838)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87840)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87841)
	}
	__antithesis_instrumentation__.Notify(87839)
	return ri.token.ClosedTimestampPolicy()
}

func (ri *RangeIterator) Token() rangecache.EvictionToken {
	__antithesis_instrumentation__.Notify(87842)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87844)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87845)
	}
	__antithesis_instrumentation__.Notify(87843)
	return ri.token
}

func (ri *RangeIterator) NeedAnother(rs roachpb.RSpan) bool {
	__antithesis_instrumentation__.Notify(87846)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87850)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87851)
	}
	__antithesis_instrumentation__.Notify(87847)
	if rs.EndKey == nil {
		__antithesis_instrumentation__.Notify(87852)
		panic("NeedAnother() undefined for spans representing a single key")
	} else {
		__antithesis_instrumentation__.Notify(87853)
	}
	__antithesis_instrumentation__.Notify(87848)
	if ri.scanDir == Ascending {
		__antithesis_instrumentation__.Notify(87854)
		return ri.Desc().EndKey.Less(rs.EndKey)
	} else {
		__antithesis_instrumentation__.Notify(87855)
	}
	__antithesis_instrumentation__.Notify(87849)
	return rs.Key.Less(ri.Desc().StartKey)
}

func (ri *RangeIterator) Valid() bool {
	__antithesis_instrumentation__.Notify(87856)
	return ri.Error() == nil
}

var errRangeIterNotInitialized = errors.New("range iterator not intialized with Seek()")

func (ri *RangeIterator) Error() error {
	__antithesis_instrumentation__.Notify(87857)
	if !ri.init {
		__antithesis_instrumentation__.Notify(87859)
		return errRangeIterNotInitialized
	} else {
		__antithesis_instrumentation__.Notify(87860)
	}
	__antithesis_instrumentation__.Notify(87858)
	return ri.err
}

func (ri *RangeIterator) Reset() {
	__antithesis_instrumentation__.Notify(87861)
	*ri = RangeIterator{ds: ri.ds}
}

var _ = (*RangeIterator)(nil).Reset

func (ri *RangeIterator) Next(ctx context.Context) {
	__antithesis_instrumentation__.Notify(87862)
	if !ri.Valid() {
		__antithesis_instrumentation__.Notify(87864)
		panic(ri.Error())
	} else {
		__antithesis_instrumentation__.Notify(87865)
	}
	__antithesis_instrumentation__.Notify(87863)

	if ri.scanDir == Ascending {
		__antithesis_instrumentation__.Notify(87866)
		ri.Seek(ctx, ri.Desc().EndKey, ri.scanDir)
	} else {
		__antithesis_instrumentation__.Notify(87867)
		ri.Seek(ctx, ri.Desc().StartKey, ri.scanDir)
	}
}

func (ri *RangeIterator) Seek(ctx context.Context, key roachpb.RKey, scanDir ScanDirection) {
	__antithesis_instrumentation__.Notify(87868)
	if log.HasSpanOrEvent(ctx) {
		__antithesis_instrumentation__.Notify(87872)
		rev := ""
		if scanDir == Descending {
			__antithesis_instrumentation__.Notify(87874)
			rev = " (rev)"
		} else {
			__antithesis_instrumentation__.Notify(87875)
		}
		__antithesis_instrumentation__.Notify(87873)
		log.Eventf(ctx, "querying next range at %s%s", key, rev)
	} else {
		__antithesis_instrumentation__.Notify(87876)
	}
	__antithesis_instrumentation__.Notify(87869)
	ri.scanDir = scanDir
	ri.init = true
	ri.err = nil
	ri.key = key

	if (scanDir == Ascending && func() bool {
		__antithesis_instrumentation__.Notify(87877)
		return key.Equal(roachpb.RKeyMax) == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(87878)
		return (scanDir == Descending && func() bool {
			__antithesis_instrumentation__.Notify(87879)
			return key.Equal(roachpb.RKeyMin) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(87880)
		ri.err = errors.Errorf("RangeIterator seek to invalid key %s", key)
		return
	} else {
		__antithesis_instrumentation__.Notify(87881)
	}
	__antithesis_instrumentation__.Notify(87870)

	var err error
	for r := retry.StartWithCtx(ctx, ri.ds.rpcRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(87882)
		var rngInfo rangecache.EvictionToken
		rngInfo, err = ri.ds.getRoutingInfo(ctx, ri.key, ri.token, ri.scanDir == Descending)

		if err != nil {
			__antithesis_instrumentation__.Notify(87885)
			log.VEventf(ctx, 1, "range descriptor lookup failed: %s", err)
			if !rangecache.IsRangeLookupErrorRetryable(err) {
				__antithesis_instrumentation__.Notify(87887)
				break
			} else {
				__antithesis_instrumentation__.Notify(87888)
			}
			__antithesis_instrumentation__.Notify(87886)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87889)
		}
		__antithesis_instrumentation__.Notify(87883)
		if log.V(2) {
			__antithesis_instrumentation__.Notify(87890)
			log.Infof(ctx, "key: %s, desc: %s", ri.key, rngInfo.Desc())
		} else {
			__antithesis_instrumentation__.Notify(87891)
		}
		__antithesis_instrumentation__.Notify(87884)

		ri.token = rngInfo
		return
	}
	__antithesis_instrumentation__.Notify(87871)

	if deducedErr := ri.ds.deduceRetryEarlyExitError(ctx); deducedErr != nil {
		__antithesis_instrumentation__.Notify(87892)
		ri.err = deducedErr
	} else {
		__antithesis_instrumentation__.Notify(87893)
		ri.err = errors.Wrapf(err, "RangeIterator failed to seek to %s", key)
	}
}
