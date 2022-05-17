package kvserverbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var MergeQueueEnabled = settings.RegisterBoolSetting(
	settings.SystemOnly,
	"kv.range_merge.queue_enabled",
	"whether the automatic merge queue is enabled",
	true,
)

const TxnCleanupThreshold = time.Hour

type CmdIDKey string

func (s CmdIDKey) SafeFormat(sp redact.SafePrinter, verb rune) {
	__antithesis_instrumentation__.Notify(101790)
	sp.Printf("%q", redact.SafeString(s))
}

func (s CmdIDKey) String() string {
	__antithesis_instrumentation__.Notify(101791)
	return redact.StringWithoutMarkers(s)
}

var _ redact.SafeFormatter = CmdIDKey("")

type FilterArgs struct {
	Ctx     context.Context
	CmdID   CmdIDKey
	Index   int
	Sid     roachpb.StoreID
	Req     roachpb.Request
	Hdr     roachpb.Header
	Version roachpb.Version
	Err     error
}

type ProposalFilterArgs struct {
	Ctx        context.Context
	Cmd        kvserverpb.RaftCommand
	QuotaAlloc *quotapool.IntAlloc
	CmdID      CmdIDKey
	Req        roachpb.BatchRequest
}

type ApplyFilterArgs struct {
	kvserverpb.ReplicatedEvalResult
	CmdID       CmdIDKey
	RangeID     roachpb.RangeID
	StoreID     roachpb.StoreID
	Req         *roachpb.BatchRequest
	ForcedError *roachpb.Error
}

func (f *FilterArgs) InRaftCmd() bool {
	__antithesis_instrumentation__.Notify(101792)
	return f.CmdID != ""
}

type ReplicaRequestFilter func(context.Context, roachpb.BatchRequest) *roachpb.Error

type ReplicaConcurrencyRetryFilter func(context.Context, roachpb.BatchRequest, *roachpb.Error)

type ReplicaCommandFilter func(args FilterArgs) *roachpb.Error

type ReplicaProposalFilter func(args ProposalFilterArgs) *roachpb.Error

type ReplicaApplyFilter func(args ApplyFilterArgs) (int, *roachpb.Error)

type ReplicaResponseFilter func(context.Context, roachpb.BatchRequest, *roachpb.BatchResponse) *roachpb.Error

type ReplicaRangefeedFilter func(
	args *roachpb.RangeFeedRequest, stream roachpb.Internal_RangeFeedServer,
) *roachpb.Error

func ContainsKey(desc *roachpb.RangeDescriptor, key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(101793)
	if bytes.HasPrefix(key, keys.LocalRangeIDPrefix) {
		__antithesis_instrumentation__.Notify(101796)
		return bytes.HasPrefix(key, keys.MakeRangeIDPrefix(desc.RangeID))
	} else {
		__antithesis_instrumentation__.Notify(101797)
	}
	__antithesis_instrumentation__.Notify(101794)
	keyAddr, err := keys.Addr(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(101798)
		return false
	} else {
		__antithesis_instrumentation__.Notify(101799)
	}
	__antithesis_instrumentation__.Notify(101795)
	return desc.ContainsKey(keyAddr)
}

func ContainsKeyRange(desc *roachpb.RangeDescriptor, start, end roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(101800)
	startKeyAddr, err := keys.Addr(start)
	if err != nil {
		__antithesis_instrumentation__.Notify(101803)
		return false
	} else {
		__antithesis_instrumentation__.Notify(101804)
	}
	__antithesis_instrumentation__.Notify(101801)
	endKeyAddr, err := keys.Addr(end)
	if err != nil {
		__antithesis_instrumentation__.Notify(101805)
		return false
	} else {
		__antithesis_instrumentation__.Notify(101806)
	}
	__antithesis_instrumentation__.Notify(101802)
	return desc.ContainsKeyRange(startKeyAddr, endKeyAddr)
}

func IntersectSpan(
	span roachpb.Span, desc *roachpb.RangeDescriptor,
) (middle *roachpb.Span, outside []roachpb.Span) {
	__antithesis_instrumentation__.Notify(101807)
	start, end := desc.StartKey.AsRawKey(), desc.EndKey.AsRawKey()
	if len(span.EndKey) == 0 {
		__antithesis_instrumentation__.Notify(101813)
		outside = append(outside, span)
		return
	} else {
		__antithesis_instrumentation__.Notify(101814)
	}
	__antithesis_instrumentation__.Notify(101808)
	if bytes.Compare(span.Key, keys.LocalRangeMax) < 0 {
		__antithesis_instrumentation__.Notify(101815)
		if bytes.Compare(span.EndKey, keys.LocalRangeMax) >= 0 {
			__antithesis_instrumentation__.Notify(101818)
			panic(fmt.Sprintf("a local intent range may not have a non-local portion: %s", span))
		} else {
			__antithesis_instrumentation__.Notify(101819)
		}
		__antithesis_instrumentation__.Notify(101816)
		if ContainsKeyRange(desc, span.Key, span.EndKey) {
			__antithesis_instrumentation__.Notify(101820)
			return &span, nil
		} else {
			__antithesis_instrumentation__.Notify(101821)
		}
		__antithesis_instrumentation__.Notify(101817)
		return nil, append(outside, span)
	} else {
		__antithesis_instrumentation__.Notify(101822)
	}
	__antithesis_instrumentation__.Notify(101809)

	if bytes.Compare(span.Key, start) < 0 {
		__antithesis_instrumentation__.Notify(101823)

		iCopy := span
		if bytes.Compare(start, span.EndKey) < 0 {
			__antithesis_instrumentation__.Notify(101825)
			iCopy.EndKey = start
		} else {
			__antithesis_instrumentation__.Notify(101826)
		}
		__antithesis_instrumentation__.Notify(101824)
		span.Key = iCopy.EndKey
		outside = append(outside, iCopy)
	} else {
		__antithesis_instrumentation__.Notify(101827)
	}
	__antithesis_instrumentation__.Notify(101810)
	if bytes.Compare(span.Key, span.EndKey) < 0 && func() bool {
		__antithesis_instrumentation__.Notify(101828)
		return bytes.Compare(end, span.EndKey) < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(101829)

		iCopy := span
		if bytes.Compare(iCopy.Key, end) < 0 {
			__antithesis_instrumentation__.Notify(101831)
			iCopy.Key = end
		} else {
			__antithesis_instrumentation__.Notify(101832)
		}
		__antithesis_instrumentation__.Notify(101830)
		span.EndKey = iCopy.Key
		outside = append(outside, iCopy)
	} else {
		__antithesis_instrumentation__.Notify(101833)
	}
	__antithesis_instrumentation__.Notify(101811)
	if bytes.Compare(span.Key, span.EndKey) < 0 && func() bool {
		__antithesis_instrumentation__.Notify(101834)
		return bytes.Compare(span.Key, start) >= 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(101835)
		return bytes.Compare(end, span.EndKey) >= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(101836)
		middle = &span
	} else {
		__antithesis_instrumentation__.Notify(101837)
	}
	__antithesis_instrumentation__.Notify(101812)
	return
}

var SplitByLoadMergeDelay = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.range_split.by_load_merge_delay",
	"the delay that range splits created due to load will wait before considering being merged away",
	5*time.Minute,
	func(v time.Duration) error {
		__antithesis_instrumentation__.Notify(101838)
		const minDelay = 5 * time.Second
		if v < minDelay {
			__antithesis_instrumentation__.Notify(101840)
			return errors.Errorf("cannot be set to a value below %s", minDelay)
		} else {
			__antithesis_instrumentation__.Notify(101841)
		}
		__antithesis_instrumentation__.Notify(101839)
		return nil
	},
)
