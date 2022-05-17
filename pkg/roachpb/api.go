package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/dustin/go-humanize"
)

type UserPriority float64

func (up UserPriority) String() string {
	__antithesis_instrumentation__.Notify(129031)
	switch up {
	case MinUserPriority:
		__antithesis_instrumentation__.Notify(129032)
		return "low"
	case UnspecifiedUserPriority, NormalUserPriority:
		__antithesis_instrumentation__.Notify(129033)
		return "normal"
	case MaxUserPriority:
		__antithesis_instrumentation__.Notify(129034)
		return "high"
	default:
		__antithesis_instrumentation__.Notify(129035)
		return fmt.Sprintf("%g", float64(up))
	}
}

const (
	MinUserPriority UserPriority = 0.001

	UnspecifiedUserPriority UserPriority = 0

	NormalUserPriority UserPriority = 1

	MaxUserPriority UserPriority = 1000
)

func (rc ReadConsistencyType) SupportsBatch(ba BatchRequest) error {
	__antithesis_instrumentation__.Notify(129036)
	switch rc {
	case CONSISTENT:
		__antithesis_instrumentation__.Notify(129038)
		return nil
	case READ_UNCOMMITTED, INCONSISTENT:
		__antithesis_instrumentation__.Notify(129039)
		for _, ru := range ba.Requests {
			__antithesis_instrumentation__.Notify(129042)
			m := ru.GetInner().Method()
			switch m {
			case Get, Scan, ReverseScan, QueryResolvedTimestamp:
				__antithesis_instrumentation__.Notify(129043)
			default:
				__antithesis_instrumentation__.Notify(129044)
				return errors.Errorf("method %s not allowed with %s batch", m, rc)
			}
		}
		__antithesis_instrumentation__.Notify(129040)
		return nil
	default:
		__antithesis_instrumentation__.Notify(129041)
	}
	__antithesis_instrumentation__.Notify(129037)
	panic("unreachable")
}

type flag int

const (
	isAdmin flag = 1 << iota
	isRead
	isWrite
	isTxn
	isLocking
	isIntentWrite
	isRange
	isReverse
	isAlone
	isPrefix
	isUnsplittable
	skipsLeaseCheck
	appliesTSCache
	updatesTSCache
	updatesTSCacheOnErr
	needsRefresh
	canBackpressure
	bypassesReplicaCircuitBreaker
)

var flagDependencies = map[flag][]flag{
	isAdmin:         {isAlone},
	isLocking:       {isTxn},
	isIntentWrite:   {isWrite, isLocking},
	appliesTSCache:  {isWrite},
	skipsLeaseCheck: {isAlone},
}

var flagExclusions = map[flag][]flag{
	skipsLeaseCheck: {isIntentWrite},
}

func IsReadOnly(args Request) bool {
	__antithesis_instrumentation__.Notify(129045)
	flags := args.flags()
	return (flags&isRead) != 0 && func() bool {
		__antithesis_instrumentation__.Notify(129046)
		return (flags & isWrite) == 0 == true
	}() == true
}

func IsBlindWrite(args Request) bool {
	__antithesis_instrumentation__.Notify(129047)
	flags := args.flags()
	return (flags&isRead) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(129048)
		return (flags & isWrite) != 0 == true
	}() == true
}

func IsTransactional(args Request) bool {
	__antithesis_instrumentation__.Notify(129049)
	return (args.flags() & isTxn) != 0
}

func IsLocking(args Request) bool {
	__antithesis_instrumentation__.Notify(129050)
	return (args.flags() & isLocking) != 0
}

func LockingDurability(args Request) lock.Durability {
	__antithesis_instrumentation__.Notify(129051)
	if IsReadOnly(args) {
		__antithesis_instrumentation__.Notify(129053)
		return lock.Unreplicated
	} else {
		__antithesis_instrumentation__.Notify(129054)
	}
	__antithesis_instrumentation__.Notify(129052)
	return lock.Replicated
}

func IsIntentWrite(args Request) bool {
	__antithesis_instrumentation__.Notify(129055)
	return (args.flags() & isIntentWrite) != 0
}

func IsRange(args Request) bool {
	__antithesis_instrumentation__.Notify(129056)
	return (args.flags() & isRange) != 0
}

func AppliesTimestampCache(args Request) bool {
	__antithesis_instrumentation__.Notify(129057)
	return (args.flags() & appliesTSCache) != 0
}

func UpdatesTimestampCache(args Request) bool {
	__antithesis_instrumentation__.Notify(129058)
	return (args.flags() & updatesTSCache) != 0
}

func UpdatesTimestampCacheOnError(args Request) bool {
	__antithesis_instrumentation__.Notify(129059)
	return (args.flags() & updatesTSCacheOnErr) != 0
}

func NeedsRefresh(args Request) bool {
	__antithesis_instrumentation__.Notify(129060)
	return (args.flags() & needsRefresh) != 0
}

func CanBackpressure(args Request) bool {
	__antithesis_instrumentation__.Notify(129061)
	return (args.flags() & canBackpressure) != 0
}

func BypassesReplicaCircuitBreaker(args Request) bool {
	__antithesis_instrumentation__.Notify(129062)
	return (args.flags() & bypassesReplicaCircuitBreaker) != 0
}

type Request interface {
	protoutil.Message

	Header() RequestHeader

	SetHeader(RequestHeader)

	Method() Method

	ShallowCopy() Request
	flags() flag
}

type SizedWriteRequest interface {
	Request
	WriteBytes() int64
}

var _ SizedWriteRequest = (*PutRequest)(nil)

func (pr *PutRequest) WriteBytes() int64 {
	__antithesis_instrumentation__.Notify(129063)
	return int64(len(pr.Key)) + int64(pr.Value.Size())
}

var _ SizedWriteRequest = (*ConditionalPutRequest)(nil)

func (cpr *ConditionalPutRequest) WriteBytes() int64 {
	__antithesis_instrumentation__.Notify(129064)
	return int64(len(cpr.Key)) + int64(cpr.Value.Size())
}

var _ SizedWriteRequest = (*InitPutRequest)(nil)

func (pr *InitPutRequest) WriteBytes() int64 {
	__antithesis_instrumentation__.Notify(129065)
	return int64(len(pr.Key)) + int64(pr.Value.Size())
}

var _ SizedWriteRequest = (*IncrementRequest)(nil)

func (ir *IncrementRequest) WriteBytes() int64 {
	__antithesis_instrumentation__.Notify(129066)
	return int64(len(ir.Key)) + 8
}

var _ SizedWriteRequest = (*DeleteRequest)(nil)

func (dr *DeleteRequest) WriteBytes() int64 {
	__antithesis_instrumentation__.Notify(129067)
	return int64(len(dr.Key))
}

var _ SizedWriteRequest = (*AddSSTableRequest)(nil)

func (r *AddSSTableRequest) WriteBytes() int64 {
	__antithesis_instrumentation__.Notify(129068)
	return int64(len(r.Data))
}

type Response interface {
	protoutil.Message

	Header() ResponseHeader

	SetHeader(ResponseHeader)

	Verify(req Request) error
}

type combinable interface {
	combine(combinable) error
}

func CombineResponses(left, right Response) error {
	__antithesis_instrumentation__.Notify(129069)
	cLeft, lOK := left.(combinable)
	cRight, rOK := right.(combinable)
	if lOK && func() bool {
		__antithesis_instrumentation__.Notify(129071)
		return rOK == true
	}() == true {
		__antithesis_instrumentation__.Notify(129072)
		return cLeft.combine(cRight)
	} else {
		__antithesis_instrumentation__.Notify(129073)
		if lOK != rOK {
			__antithesis_instrumentation__.Notify(129074)
			return errors.Errorf("can not combine %T and %T", left, right)
		} else {
			__antithesis_instrumentation__.Notify(129075)
		}
	}
	__antithesis_instrumentation__.Notify(129070)
	return nil
}

func (rh *ResponseHeader) combine(otherRH ResponseHeader) error {
	__antithesis_instrumentation__.Notify(129076)
	if rh.Txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(129079)
		return otherRH.Txn == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(129080)
		rh.Txn = nil
	} else {
		__antithesis_instrumentation__.Notify(129081)
	}
	__antithesis_instrumentation__.Notify(129077)
	if rh.ResumeSpan != nil {
		__antithesis_instrumentation__.Notify(129082)
		return errors.Errorf("combining %+v with %+v", rh.ResumeSpan, otherRH.ResumeSpan)
	} else {
		__antithesis_instrumentation__.Notify(129083)
	}
	__antithesis_instrumentation__.Notify(129078)
	rh.ResumeSpan = otherRH.ResumeSpan
	rh.ResumeReason = otherRH.ResumeReason
	rh.NumKeys += otherRH.NumKeys
	rh.NumBytes += otherRH.NumBytes
	return nil
}

func (sr *ScanResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129084)
	otherSR := c.(*ScanResponse)
	if sr != nil {
		__antithesis_instrumentation__.Notify(129086)
		sr.Rows = append(sr.Rows, otherSR.Rows...)
		sr.IntentRows = append(sr.IntentRows, otherSR.IntentRows...)
		sr.BatchResponses = append(sr.BatchResponses, otherSR.BatchResponses...)
		if err := sr.ResponseHeader.combine(otherSR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129087)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129088)
		}
	} else {
		__antithesis_instrumentation__.Notify(129089)
	}
	__antithesis_instrumentation__.Notify(129085)
	return nil
}

var _ combinable = &ScanResponse{}

func (sr *ReverseScanResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129090)
	otherSR := c.(*ReverseScanResponse)
	if sr != nil {
		__antithesis_instrumentation__.Notify(129092)
		sr.Rows = append(sr.Rows, otherSR.Rows...)
		sr.IntentRows = append(sr.IntentRows, otherSR.IntentRows...)
		sr.BatchResponses = append(sr.BatchResponses, otherSR.BatchResponses...)
		if err := sr.ResponseHeader.combine(otherSR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129093)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129094)
		}
	} else {
		__antithesis_instrumentation__.Notify(129095)
	}
	__antithesis_instrumentation__.Notify(129091)
	return nil
}

var _ combinable = &ReverseScanResponse{}

func (dr *DeleteRangeResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129096)
	otherDR := c.(*DeleteRangeResponse)
	if dr != nil {
		__antithesis_instrumentation__.Notify(129098)
		dr.Keys = append(dr.Keys, otherDR.Keys...)
		if err := dr.ResponseHeader.combine(otherDR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129099)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129100)
		}
	} else {
		__antithesis_instrumentation__.Notify(129101)
	}
	__antithesis_instrumentation__.Notify(129097)
	return nil
}

var _ combinable = &DeleteRangeResponse{}

func (dr *RevertRangeResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129102)
	otherDR := c.(*RevertRangeResponse)
	if dr != nil {
		__antithesis_instrumentation__.Notify(129104)
		if err := dr.ResponseHeader.combine(otherDR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129105)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129106)
		}
	} else {
		__antithesis_instrumentation__.Notify(129107)
	}
	__antithesis_instrumentation__.Notify(129103)
	return nil
}

var _ combinable = &RevertRangeResponse{}

func (rr *ResolveIntentRangeResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129108)
	otherRR := c.(*ResolveIntentRangeResponse)
	if rr != nil {
		__antithesis_instrumentation__.Notify(129110)
		if err := rr.ResponseHeader.combine(otherRR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129111)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129112)
		}
	} else {
		__antithesis_instrumentation__.Notify(129113)
	}
	__antithesis_instrumentation__.Notify(129109)
	return nil
}

var _ combinable = &ResolveIntentRangeResponse{}

func (cc *CheckConsistencyResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129114)
	if cc != nil {
		__antithesis_instrumentation__.Notify(129116)
		otherCC := c.(*CheckConsistencyResponse)
		cc.Result = append(cc.Result, otherCC.Result...)
		if err := cc.ResponseHeader.combine(otherCC.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129117)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129118)
		}
	} else {
		__antithesis_instrumentation__.Notify(129119)
	}
	__antithesis_instrumentation__.Notify(129115)
	return nil
}

var _ combinable = &CheckConsistencyResponse{}

func (er *ExportResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129120)
	if er != nil {
		__antithesis_instrumentation__.Notify(129122)
		otherER := c.(*ExportResponse)
		if err := er.ResponseHeader.combine(otherER.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129124)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129125)
		}
		__antithesis_instrumentation__.Notify(129123)
		er.Files = append(er.Files, otherER.Files...)
	} else {
		__antithesis_instrumentation__.Notify(129126)
	}
	__antithesis_instrumentation__.Notify(129121)
	return nil
}

var _ combinable = &ExportResponse{}

func (r *AdminScatterResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129127)
	if r != nil {
		__antithesis_instrumentation__.Notify(129129)
		otherR := c.(*AdminScatterResponse)
		if err := r.ResponseHeader.combine(otherR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129131)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129132)
		}
		__antithesis_instrumentation__.Notify(129130)

		r.RangeInfos = append(r.RangeInfos, otherR.RangeInfos...)
	} else {
		__antithesis_instrumentation__.Notify(129133)
	}
	__antithesis_instrumentation__.Notify(129128)
	return nil
}

var _ combinable = &AdminScatterResponse{}

func (avptr *AdminVerifyProtectedTimestampResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129134)
	other := c.(*AdminVerifyProtectedTimestampResponse)
	if avptr != nil {
		__antithesis_instrumentation__.Notify(129136)
		avptr.DeprecatedFailedRanges = append(avptr.DeprecatedFailedRanges,
			other.DeprecatedFailedRanges...)
		avptr.VerificationFailedRanges = append(avptr.VerificationFailedRanges,
			other.VerificationFailedRanges...)
		if err := avptr.ResponseHeader.combine(other.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129137)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129138)
		}
	} else {
		__antithesis_instrumentation__.Notify(129139)
	}
	__antithesis_instrumentation__.Notify(129135)
	return nil
}

var _ combinable = &AdminVerifyProtectedTimestampResponse{}

func (r *QueryResolvedTimestampResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129140)
	if r != nil {
		__antithesis_instrumentation__.Notify(129142)
		otherR := c.(*QueryResolvedTimestampResponse)
		if err := r.ResponseHeader.combine(otherR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129144)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129145)
		}
		__antithesis_instrumentation__.Notify(129143)

		r.ResolvedTS.Backward(otherR.ResolvedTS)
	} else {
		__antithesis_instrumentation__.Notify(129146)
	}
	__antithesis_instrumentation__.Notify(129141)
	return nil
}

var _ combinable = &QueryResolvedTimestampResponse{}

func (r *BarrierResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129147)
	otherR := c.(*BarrierResponse)
	if r != nil {
		__antithesis_instrumentation__.Notify(129149)
		if err := r.ResponseHeader.combine(otherR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129151)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129152)
		}
		__antithesis_instrumentation__.Notify(129150)
		r.Timestamp.Forward(otherR.Timestamp)
	} else {
		__antithesis_instrumentation__.Notify(129153)
	}
	__antithesis_instrumentation__.Notify(129148)
	return nil
}

var _ combinable = &BarrierResponse{}

func (r *ScanInterleavedIntentsResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129154)
	otherR := c.(*ScanInterleavedIntentsResponse)
	if r != nil {
		__antithesis_instrumentation__.Notify(129156)
		if err := r.ResponseHeader.combine(otherR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129158)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129159)
		}
		__antithesis_instrumentation__.Notify(129157)
		r.Intents = append(r.Intents, otherR.Intents...)
	} else {
		__antithesis_instrumentation__.Notify(129160)
	}
	__antithesis_instrumentation__.Notify(129155)
	return nil
}

var _ combinable = &ScanInterleavedIntentsResponse{}

func (r *QueryLocksResponse) combine(c combinable) error {
	__antithesis_instrumentation__.Notify(129161)
	otherR := c.(*QueryLocksResponse)
	if r != nil {
		__antithesis_instrumentation__.Notify(129163)
		if err := r.ResponseHeader.combine(otherR.Header()); err != nil {
			__antithesis_instrumentation__.Notify(129165)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129166)
		}
		__antithesis_instrumentation__.Notify(129164)
		r.Locks = append(r.Locks, otherR.Locks...)
	} else {
		__antithesis_instrumentation__.Notify(129167)
	}
	__antithesis_instrumentation__.Notify(129162)
	return nil
}

var _ combinable = &QueryLocksResponse{}

func (rh RequestHeader) Header() RequestHeader {
	__antithesis_instrumentation__.Notify(129168)
	return rh
}

func (rh *RequestHeader) SetHeader(other RequestHeader) {
	__antithesis_instrumentation__.Notify(129169)
	*rh = other
}

func (rh RequestHeader) Span() Span {
	__antithesis_instrumentation__.Notify(129170)
	return Span{Key: rh.Key, EndKey: rh.EndKey}
}

func (rh *RequestHeader) SetSpan(s Span) {
	__antithesis_instrumentation__.Notify(129171)
	rh.Key = s.Key
	rh.EndKey = s.EndKey
}

func RequestHeaderFromSpan(s Span) RequestHeader {
	__antithesis_instrumentation__.Notify(129172)
	return RequestHeader{Key: s.Key, EndKey: s.EndKey}
}

func (h *BatchResponse_Header) combine(o BatchResponse_Header) error {
	__antithesis_instrumentation__.Notify(129173)
	if h.Error != nil || func() bool {
		__antithesis_instrumentation__.Notify(129176)
		return o.Error != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(129177)
		return errors.Errorf(
			"can't combine batch responses with errors, have errors %q and %q",
			h.Error, o.Error,
		)
	} else {
		__antithesis_instrumentation__.Notify(129178)
	}
	__antithesis_instrumentation__.Notify(129174)
	h.Timestamp.Forward(o.Timestamp)
	if txn := o.Txn; txn != nil {
		__antithesis_instrumentation__.Notify(129179)
		if h.Txn == nil {
			__antithesis_instrumentation__.Notify(129180)
			h.Txn = txn.Clone()
		} else {
			__antithesis_instrumentation__.Notify(129181)
			h.Txn.Update(txn)
		}
	} else {
		__antithesis_instrumentation__.Notify(129182)
	}
	__antithesis_instrumentation__.Notify(129175)
	h.Now.Forward(o.Now)
	h.RangeInfos = append(h.RangeInfos, o.RangeInfos...)
	h.CollectedSpans = append(h.CollectedSpans, o.CollectedSpans...)
	return nil
}

func (rh *ResponseHeader) SetHeader(other ResponseHeader) {
	__antithesis_instrumentation__.Notify(129183)
	*rh = other
}

func (rh ResponseHeader) Header() ResponseHeader {
	__antithesis_instrumentation__.Notify(129184)
	return rh
}

func (rh *ResponseHeader) Verify(req Request) error {
	__antithesis_instrumentation__.Notify(129185)
	return nil
}

func (gr *GetResponse) Verify(req Request) error {
	__antithesis_instrumentation__.Notify(129186)
	if gr.Value != nil {
		__antithesis_instrumentation__.Notify(129188)
		return gr.Value.Verify(req.Header().Key)
	} else {
		__antithesis_instrumentation__.Notify(129189)
	}
	__antithesis_instrumentation__.Notify(129187)
	return nil
}

func (sr *ScanResponse) Verify(req Request) error {
	__antithesis_instrumentation__.Notify(129190)
	for _, kv := range sr.Rows {
		__antithesis_instrumentation__.Notify(129192)
		if err := kv.Value.Verify(kv.Key); err != nil {
			__antithesis_instrumentation__.Notify(129193)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129194)
		}
	}
	__antithesis_instrumentation__.Notify(129191)
	return nil
}

func (sr *ReverseScanResponse) Verify(req Request) error {
	__antithesis_instrumentation__.Notify(129195)
	for _, kv := range sr.Rows {
		__antithesis_instrumentation__.Notify(129197)
		if err := kv.Value.Verify(kv.Key); err != nil {
			__antithesis_instrumentation__.Notify(129198)
			return err
		} else {
			__antithesis_instrumentation__.Notify(129199)
		}
	}
	__antithesis_instrumentation__.Notify(129196)
	return nil
}

func (*GetRequest) Method() Method { __antithesis_instrumentation__.Notify(129200); return Get }

func (*PutRequest) Method() Method { __antithesis_instrumentation__.Notify(129201); return Put }

func (*ConditionalPutRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129202)
	return ConditionalPut
}

func (*InitPutRequest) Method() Method { __antithesis_instrumentation__.Notify(129203); return InitPut }

func (*IncrementRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129204)
	return Increment
}

func (*DeleteRequest) Method() Method { __antithesis_instrumentation__.Notify(129205); return Delete }

func (*DeleteRangeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129206)
	return DeleteRange
}

func (*ClearRangeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129207)
	return ClearRange
}

func (*RevertRangeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129208)
	return RevertRange
}

func (*ScanRequest) Method() Method { __antithesis_instrumentation__.Notify(129209); return Scan }

func (*ReverseScanRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129210)
	return ReverseScan
}

func (*CheckConsistencyRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129211)
	return CheckConsistency
}

func (*EndTxnRequest) Method() Method { __antithesis_instrumentation__.Notify(129212); return EndTxn }

func (*AdminSplitRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129213)
	return AdminSplit
}

func (*AdminUnsplitRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129214)
	return AdminUnsplit
}

func (*AdminMergeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129215)
	return AdminMerge
}

func (*AdminTransferLeaseRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129216)
	return AdminTransferLease
}

func (*AdminChangeReplicasRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129217)
	return AdminChangeReplicas
}

func (*AdminRelocateRangeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129218)
	return AdminRelocateRange
}

func (*HeartbeatTxnRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129219)
	return HeartbeatTxn
}

func (*GCRequest) Method() Method { __antithesis_instrumentation__.Notify(129220); return GC }

func (*PushTxnRequest) Method() Method { __antithesis_instrumentation__.Notify(129221); return PushTxn }

func (*RecoverTxnRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129222)
	return RecoverTxn
}

func (*QueryTxnRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129223)
	return QueryTxn
}

func (*QueryIntentRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129224)
	return QueryIntent
}

func (*QueryLocksRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129225)
	return QueryLocks
}

func (*ResolveIntentRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129226)
	return ResolveIntent
}

func (*ResolveIntentRangeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129227)
	return ResolveIntentRange
}

func (*MergeRequest) Method() Method { __antithesis_instrumentation__.Notify(129228); return Merge }

func (*TruncateLogRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129229)
	return TruncateLog
}

func (*RequestLeaseRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129230)
	return RequestLease
}

func (*TransferLeaseRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129231)
	return TransferLease
}

func (*ProbeRequest) Method() Method { __antithesis_instrumentation__.Notify(129232); return Probe }

func (*LeaseInfoRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129233)
	return LeaseInfo
}

func (*ComputeChecksumRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129234)
	return ComputeChecksum
}

func (*ExportRequest) Method() Method { __antithesis_instrumentation__.Notify(129235); return Export }

func (*AdminScatterRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129236)
	return AdminScatter
}

func (*AddSSTableRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129237)
	return AddSSTable
}

func (*MigrateRequest) Method() Method { __antithesis_instrumentation__.Notify(129238); return Migrate }

func (*RecomputeStatsRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129239)
	return RecomputeStats
}

func (*RefreshRequest) Method() Method { __antithesis_instrumentation__.Notify(129240); return Refresh }

func (*RefreshRangeRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129241)
	return RefreshRange
}

func (*SubsumeRequest) Method() Method { __antithesis_instrumentation__.Notify(129242); return Subsume }

func (*RangeStatsRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129243)
	return RangeStats
}

func (*AdminVerifyProtectedTimestampRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129244)
	return AdminVerifyProtectedTimestamp
}

func (*QueryResolvedTimestampRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129245)
	return QueryResolvedTimestamp
}

func (*ScanInterleavedIntentsRequest) Method() Method {
	__antithesis_instrumentation__.Notify(129246)
	return ScanInterleavedIntents
}

func (*BarrierRequest) Method() Method { __antithesis_instrumentation__.Notify(129247); return Barrier }

func (gr *GetRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129248)
	shallowCopy := *gr
	return &shallowCopy
}

func (pr *PutRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129249)
	shallowCopy := *pr
	return &shallowCopy
}

func (cpr *ConditionalPutRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129250)
	shallowCopy := *cpr
	return &shallowCopy
}

func (pr *InitPutRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129251)
	shallowCopy := *pr
	return &shallowCopy
}

func (ir *IncrementRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129252)
	shallowCopy := *ir
	return &shallowCopy
}

func (dr *DeleteRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129253)
	shallowCopy := *dr
	return &shallowCopy
}

func (drr *DeleteRangeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129254)
	shallowCopy := *drr
	return &shallowCopy
}

func (crr *ClearRangeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129255)
	shallowCopy := *crr
	return &shallowCopy
}

func (crr *RevertRangeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129256)
	shallowCopy := *crr
	return &shallowCopy
}

func (sr *ScanRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129257)
	shallowCopy := *sr
	return &shallowCopy
}

func (rsr *ReverseScanRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129258)
	shallowCopy := *rsr
	return &shallowCopy
}

func (ccr *CheckConsistencyRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129259)
	shallowCopy := *ccr
	return &shallowCopy
}

func (etr *EndTxnRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129260)
	shallowCopy := *etr
	return &shallowCopy
}

func (asr *AdminSplitRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129261)
	shallowCopy := *asr
	return &shallowCopy
}

func (aur *AdminUnsplitRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129262)
	shallowCopy := *aur
	return &shallowCopy
}

func (amr *AdminMergeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129263)
	shallowCopy := *amr
	return &shallowCopy
}

func (atlr *AdminTransferLeaseRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129264)
	shallowCopy := *atlr
	return &shallowCopy
}

func (acrr *AdminChangeReplicasRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129265)
	shallowCopy := *acrr
	return &shallowCopy
}

func (acrr *AdminRelocateRangeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129266)
	shallowCopy := *acrr
	return &shallowCopy
}

func (htr *HeartbeatTxnRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129267)
	shallowCopy := *htr
	return &shallowCopy
}

func (gcr *GCRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129268)
	shallowCopy := *gcr
	return &shallowCopy
}

func (ptr *PushTxnRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129269)
	shallowCopy := *ptr
	return &shallowCopy
}

func (rtr *RecoverTxnRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129270)
	shallowCopy := *rtr
	return &shallowCopy
}

func (qtr *QueryTxnRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129271)
	shallowCopy := *qtr
	return &shallowCopy
}

func (pir *QueryIntentRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129272)
	shallowCopy := *pir
	return &shallowCopy
}

func (pir *QueryLocksRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129273)
	shallowCopy := *pir
	return &shallowCopy
}

func (rir *ResolveIntentRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129274)
	shallowCopy := *rir
	return &shallowCopy
}

func (rirr *ResolveIntentRangeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129275)
	shallowCopy := *rirr
	return &shallowCopy
}

func (mr *MergeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129276)
	shallowCopy := *mr
	return &shallowCopy
}

func (tlr *TruncateLogRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129277)
	shallowCopy := *tlr
	return &shallowCopy
}

func (rlr *RequestLeaseRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129278)
	shallowCopy := *rlr
	return &shallowCopy
}

func (tlr *TransferLeaseRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129279)
	shallowCopy := *tlr
	return &shallowCopy
}

func (r *ProbeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129280)
	shallowCopy := *r
	return &shallowCopy
}

func (lt *LeaseInfoRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129281)
	shallowCopy := *lt
	return &shallowCopy
}

func (ccr *ComputeChecksumRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129282)
	shallowCopy := *ccr
	return &shallowCopy
}

func (ekr *ExportRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129283)
	shallowCopy := *ekr
	return &shallowCopy
}

func (r *AdminScatterRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129284)
	shallowCopy := *r
	return &shallowCopy
}

func (r *AddSSTableRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129285)
	shallowCopy := *r
	return &shallowCopy
}

func (r *MigrateRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129286)
	shallowCopy := *r
	return &shallowCopy
}

func (r *RecomputeStatsRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129287)
	shallowCopy := *r
	return &shallowCopy
}

func (r *RefreshRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129288)
	shallowCopy := *r
	return &shallowCopy
}

func (r *RefreshRangeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129289)
	shallowCopy := *r
	return &shallowCopy
}

func (r *SubsumeRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129290)
	shallowCopy := *r
	return &shallowCopy
}

func (r *RangeStatsRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129291)
	shallowCopy := *r
	return &shallowCopy
}

func (r *AdminVerifyProtectedTimestampRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129292)
	shallowCopy := *r
	return &shallowCopy
}

func (r *QueryResolvedTimestampRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129293)
	shallowCopy := *r
	return &shallowCopy
}

func (r *ScanInterleavedIntentsRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129294)
	shallowCopy := *r
	return &shallowCopy
}

func (r *BarrierRequest) ShallowCopy() Request {
	__antithesis_instrumentation__.Notify(129295)
	shallowCopy := *r
	return &shallowCopy
}

func NewGet(key Key, forUpdate bool) Request {
	__antithesis_instrumentation__.Notify(129296)
	return &GetRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		KeyLocking: scanLockStrength(forUpdate),
	}
}

func NewIncrement(key Key, increment int64) Request {
	__antithesis_instrumentation__.Notify(129297)
	return &IncrementRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		Increment: increment,
	}
}

func NewPut(key Key, value Value) Request {
	__antithesis_instrumentation__.Notify(129298)
	value.InitChecksum(key)
	return &PutRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		Value: value,
	}
}

func NewPutInline(key Key, value Value) Request {
	__antithesis_instrumentation__.Notify(129299)
	value.InitChecksum(key)
	return &PutRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		Value:  value,
		Inline: true,
	}
}

func NewConditionalPut(key Key, value Value, expValue []byte, allowNotExist bool) Request {
	__antithesis_instrumentation__.Notify(129300)
	value.InitChecksum(key)

	var expValueVal *Value
	if expValue != nil {
		__antithesis_instrumentation__.Notify(129302)
		expValueVal = &Value{}
		expValueVal.SetTagAndData(expValue)

	} else {
		__antithesis_instrumentation__.Notify(129303)
	}
	__antithesis_instrumentation__.Notify(129301)

	return &ConditionalPutRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		Value:               value,
		DeprecatedExpValue:  expValueVal,
		ExpBytes:            expValue,
		AllowIfDoesNotExist: allowNotExist,
	}
}

func NewConditionalPutInline(key Key, value Value, expValue []byte, allowNotExist bool) Request {
	__antithesis_instrumentation__.Notify(129304)
	value.InitChecksum(key)

	var expValueVal *Value
	if expValue != nil {
		__antithesis_instrumentation__.Notify(129306)
		expValueVal = &Value{}
		expValueVal.SetTagAndData(expValue)

	} else {
		__antithesis_instrumentation__.Notify(129307)
	}
	__antithesis_instrumentation__.Notify(129305)

	return &ConditionalPutRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		Value:               value,
		DeprecatedExpValue:  expValueVal,
		ExpBytes:            expValue,
		AllowIfDoesNotExist: allowNotExist,
		Inline:              true,
	}
}

func NewInitPut(key Key, value Value, failOnTombstones bool) Request {
	__antithesis_instrumentation__.Notify(129308)
	value.InitChecksum(key)
	return &InitPutRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
		Value:            value,
		FailOnTombstones: failOnTombstones,
	}
}

func NewDelete(key Key) Request {
	__antithesis_instrumentation__.Notify(129309)
	return &DeleteRequest{
		RequestHeader: RequestHeader{
			Key: key,
		},
	}
}

func NewDeleteRange(startKey, endKey Key, returnKeys bool) Request {
	__antithesis_instrumentation__.Notify(129310)
	return &DeleteRangeRequest{
		RequestHeader: RequestHeader{
			Key:    startKey,
			EndKey: endKey,
		},
		ReturnKeys: returnKeys,
	}
}

func NewScan(key, endKey Key, forUpdate bool) Request {
	__antithesis_instrumentation__.Notify(129311)
	return &ScanRequest{
		RequestHeader: RequestHeader{
			Key:    key,
			EndKey: endKey,
		},
		KeyLocking: scanLockStrength(forUpdate),
	}
}

func NewReverseScan(key, endKey Key, forUpdate bool) Request {
	__antithesis_instrumentation__.Notify(129312)
	return &ReverseScanRequest{
		RequestHeader: RequestHeader{
			Key:    key,
			EndKey: endKey,
		},
		KeyLocking: scanLockStrength(forUpdate),
	}
}

func scanLockStrength(forUpdate bool) lock.Strength {
	__antithesis_instrumentation__.Notify(129313)
	if forUpdate {
		__antithesis_instrumentation__.Notify(129315)
		return lock.Exclusive
	} else {
		__antithesis_instrumentation__.Notify(129316)
	}
	__antithesis_instrumentation__.Notify(129314)
	return lock.None
}

func flagForLockStrength(l lock.Strength) flag {
	__antithesis_instrumentation__.Notify(129317)
	if l != lock.None {
		__antithesis_instrumentation__.Notify(129319)
		return isLocking
	} else {
		__antithesis_instrumentation__.Notify(129320)
	}
	__antithesis_instrumentation__.Notify(129318)
	return 0
}

func (gr *GetRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129321)
	maybeLocking := flagForLockStrength(gr.KeyLocking)
	return isRead | isTxn | maybeLocking | updatesTSCache | needsRefresh
}

func (*PutRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129322)
	return isWrite | isTxn | isLocking | isIntentWrite | appliesTSCache | canBackpressure
}

func (*ConditionalPutRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129323)
	return isRead | isWrite | isTxn | isLocking | isIntentWrite |
		appliesTSCache | updatesTSCache | updatesTSCacheOnErr | canBackpressure
}

func (*InitPutRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129324)
	return isRead | isWrite | isTxn | isLocking | isIntentWrite |
		appliesTSCache | updatesTSCache | updatesTSCacheOnErr | canBackpressure
}

func (*IncrementRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129325)
	return isRead | isWrite | isTxn | isLocking | isIntentWrite | appliesTSCache | canBackpressure
}

func (*DeleteRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129326)
	return isWrite | isTxn | isLocking | isIntentWrite | appliesTSCache | canBackpressure
}

func (drr *DeleteRangeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129327)

	if drr.Inline {
		__antithesis_instrumentation__.Notify(129329)
		return isRead | isWrite | isRange | isAlone
	} else {
		__antithesis_instrumentation__.Notify(129330)
	}
	__antithesis_instrumentation__.Notify(129328)

	return isRead | isWrite | isTxn | isLocking | isIntentWrite | isRange |
		appliesTSCache | updatesTSCache | needsRefresh | canBackpressure
}

func (*ClearRangeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129331)
	return isWrite | isRange | isAlone | bypassesReplicaCircuitBreaker
}

func (*RevertRangeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129332)
	return isWrite | isRange | isAlone | bypassesReplicaCircuitBreaker
}

func (sr *ScanRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129333)
	maybeLocking := flagForLockStrength(sr.KeyLocking)
	return isRead | isRange | isTxn | maybeLocking | updatesTSCache | needsRefresh
}

func (rsr *ReverseScanRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129334)
	maybeLocking := flagForLockStrength(rsr.KeyLocking)
	return isRead | isRange | isReverse | isTxn | maybeLocking | updatesTSCache | needsRefresh
}

func (*EndTxnRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129335)
	return isWrite | isTxn | isAlone | updatesTSCache
}
func (*AdminSplitRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129336)
	return isAdmin | isAlone
}
func (*AdminUnsplitRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129337)
	return isAdmin | isAlone
}
func (*AdminMergeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129338)
	return isAdmin | isAlone
}
func (*AdminTransferLeaseRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129339)
	return isAdmin | isAlone
}
func (*AdminChangeReplicasRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129340)
	return isAdmin | isAlone
}
func (*AdminRelocateRangeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129341)
	return isAdmin | isAlone
}

func (*GCRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129342)

	return isWrite | isRange | bypassesReplicaCircuitBreaker
}

func (*HeartbeatTxnRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129343)
	return isWrite | isTxn | updatesTSCache
}

func (*PushTxnRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129344)
	return isWrite | isAlone | updatesTSCache | updatesTSCache
}
func (*RecoverTxnRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129345)
	return isWrite | isAlone | updatesTSCache
}
func (*QueryTxnRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129346)
	return isRead | isAlone
}

func (*QueryIntentRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129347)
	return isRead | isPrefix | updatesTSCache | updatesTSCacheOnErr
}
func (*QueryLocksRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129348)
	return isRead | isRange
}
func (*ResolveIntentRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129349)
	return isWrite
}
func (*ResolveIntentRangeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129350)
	return isWrite | isRange
}
func (*TruncateLogRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129351)
	return isWrite
}
func (*MergeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129352)
	return isWrite | canBackpressure
}
func (*RequestLeaseRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129353)
	return isWrite | isAlone | skipsLeaseCheck | bypassesReplicaCircuitBreaker
}

func (*LeaseInfoRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129354)
	return isRead | isAlone
}
func (*TransferLeaseRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129355)

	return isWrite | isAlone | skipsLeaseCheck
}
func (*ProbeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129356)
	return isWrite | isAlone | skipsLeaseCheck | bypassesReplicaCircuitBreaker
}
func (*RecomputeStatsRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129357)
	return isWrite | isAlone
}
func (*ComputeChecksumRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129358)
	return isWrite
}
func (*CheckConsistencyRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129359)
	return isAdmin | isRange | isAlone
}
func (*ExportRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129360)
	return isRead | isRange | updatesTSCache | bypassesReplicaCircuitBreaker
}
func (*AdminScatterRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129361)
	return isAdmin | isRange | isAlone
}
func (*AdminVerifyProtectedTimestampRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129362)
	return isAdmin | isRange | isAlone
}
func (r *AddSSTableRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129363)
	flags := isWrite | isRange | isAlone | isUnsplittable | canBackpressure | bypassesReplicaCircuitBreaker
	if r.SSTTimestampToRequestTimestamp.IsSet() {
		__antithesis_instrumentation__.Notify(129365)
		flags |= appliesTSCache
	} else {
		__antithesis_instrumentation__.Notify(129366)
	}
	__antithesis_instrumentation__.Notify(129364)
	return flags
}
func (*MigrateRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129367)
	return isWrite | isRange | isAlone
}

func (r *RefreshRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129368)
	return isRead | isTxn | updatesTSCache
}
func (r *RefreshRangeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129369)
	return isRead | isTxn | isRange | updatesTSCache
}

func (*SubsumeRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129370)
	return isRead | isAlone | updatesTSCache
}
func (*RangeStatsRequest) flags() flag { __antithesis_instrumentation__.Notify(129371); return isRead }
func (*QueryResolvedTimestampRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129372)
	return isRead | isRange
}
func (*ScanInterleavedIntentsRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129373)
	return isRead | isRange
}
func (*BarrierRequest) flags() flag {
	__antithesis_instrumentation__.Notify(129374)
	return isWrite | isRange
}

func (etr *EndTxnRequest) IsParallelCommit() bool {
	__antithesis_instrumentation__.Notify(129375)
	return etr.Commit && func() bool {
		__antithesis_instrumentation__.Notify(129376)
		return len(etr.InFlightWrites) > 0 == true
	}() == true
}

const (
	ExternalStorageAuthImplicit = "implicit"

	ExternalStorageAuthSpecified = "specified"
)

func (m *ExternalStorage) AccessIsWithExplicitAuth() bool {
	__antithesis_instrumentation__.Notify(129377)
	switch m.Provider {
	case ExternalStorageProvider_s3:
		__antithesis_instrumentation__.Notify(129378)

		if m.S3Config.Endpoint != "" {
			__antithesis_instrumentation__.Notify(129387)
			return false
		} else {
			__antithesis_instrumentation__.Notify(129388)
		}
		__antithesis_instrumentation__.Notify(129379)
		return m.S3Config.Auth != ExternalStorageAuthImplicit
	case ExternalStorageProvider_gs:
		__antithesis_instrumentation__.Notify(129380)
		return m.GoogleCloudConfig.Auth == ExternalStorageAuthSpecified
	case ExternalStorageProvider_azure:
		__antithesis_instrumentation__.Notify(129381)

		return true
	case ExternalStorageProvider_userfile:
		__antithesis_instrumentation__.Notify(129382)

		return true
	case ExternalStorageProvider_null:
		__antithesis_instrumentation__.Notify(129383)
		return true
	case ExternalStorageProvider_http:
		__antithesis_instrumentation__.Notify(129384)

		return false
	case ExternalStorageProvider_nodelocal:
		__antithesis_instrumentation__.Notify(129385)

		return false
	default:
		__antithesis_instrumentation__.Notify(129386)
		return false
	}
}

func BulkOpSummaryID(tableID, indexID uint64) uint64 {
	__antithesis_instrumentation__.Notify(129389)
	return (tableID << 32) | indexID
}

func (b *BulkOpSummary) Add(other BulkOpSummary) {
	__antithesis_instrumentation__.Notify(129390)
	b.DataSize += other.DataSize
	b.DeprecatedRows += other.DeprecatedRows
	b.DeprecatedIndexEntries += other.DeprecatedIndexEntries

	if other.EntryCounts != nil && func() bool {
		__antithesis_instrumentation__.Notify(129392)
		return b.EntryCounts == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(129393)
		b.EntryCounts = make(map[uint64]int64, len(other.EntryCounts))
	} else {
		__antithesis_instrumentation__.Notify(129394)
	}
	__antithesis_instrumentation__.Notify(129391)
	for i := range other.EntryCounts {
		__antithesis_instrumentation__.Notify(129395)
		b.EntryCounts[i] += other.EntryCounts[i]
	}
}

func (e *RangeFeedEvent) MustSetValue(value interface{}) {
	__antithesis_instrumentation__.Notify(129396)
	e.Reset()
	if !e.SetValue(value) {
		__antithesis_instrumentation__.Notify(129397)
		panic(errors.AssertionFailedf("%T excludes %T", e, value))
	} else {
		__antithesis_instrumentation__.Notify(129398)
	}
}

func (e *RangeFeedEvent) ShallowCopy() *RangeFeedEvent {
	__antithesis_instrumentation__.Notify(129399)
	cpy := *e
	switch t := cpy.GetValue().(type) {
	case *RangeFeedValue:
		__antithesis_instrumentation__.Notify(129401)
		cpyVal := *t
		cpy.MustSetValue(&cpyVal)
	case *RangeFeedCheckpoint:
		__antithesis_instrumentation__.Notify(129402)
		cpyChk := *t
		cpy.MustSetValue(&cpyChk)
	case *RangeFeedSSTable:
		__antithesis_instrumentation__.Notify(129403)
		cpySST := *t
		cpy.MustSetValue(&cpySST)
	case *RangeFeedError:
		__antithesis_instrumentation__.Notify(129404)
		cpyErr := *t
		cpy.MustSetValue(&cpyErr)
	default:
		__antithesis_instrumentation__.Notify(129405)
		panic(fmt.Sprintf("unexpected RangeFeedEvent variant: %v", t))
	}
	__antithesis_instrumentation__.Notify(129400)
	return &cpy
}

func (e *RangeFeedValue) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(129406)
	return e.Value.Timestamp
}

func MakeReplicationChanges(
	changeType ReplicaChangeType, targets ...ReplicationTarget,
) []ReplicationChange {
	__antithesis_instrumentation__.Notify(129407)
	chgs := make([]ReplicationChange, 0, len(targets))
	for _, target := range targets {
		__antithesis_instrumentation__.Notify(129409)
		chgs = append(chgs, ReplicationChange{
			ChangeType: changeType,
			Target:     target,
		})
	}
	__antithesis_instrumentation__.Notify(129408)
	return chgs
}

func ReplicationChangesForPromotion(target ReplicationTarget) []ReplicationChange {
	__antithesis_instrumentation__.Notify(129410)
	return []ReplicationChange{
		{ChangeType: ADD_VOTER, Target: target}, {ChangeType: REMOVE_NON_VOTER, Target: target},
	}
}

func ReplicationChangesForDemotion(target ReplicationTarget) []ReplicationChange {
	__antithesis_instrumentation__.Notify(129411)
	return []ReplicationChange{
		{ChangeType: ADD_NON_VOTER, Target: target}, {ChangeType: REMOVE_VOTER, Target: target},
	}
}

func (acrr *AdminChangeReplicasRequest) AddChanges(chgs ...ReplicationChange) {
	__antithesis_instrumentation__.Notify(129412)
	acrr.InternalChanges = append(acrr.InternalChanges, chgs...)

	acrr.DeprecatedChangeType = chgs[0].ChangeType
	for _, chg := range chgs {
		__antithesis_instrumentation__.Notify(129413)
		acrr.DeprecatedTargets = append(acrr.DeprecatedTargets, chg.Target)
	}
}

type ReplicationChanges []ReplicationChange

func (rc ReplicationChanges) byType(typ ReplicaChangeType) []ReplicationTarget {
	__antithesis_instrumentation__.Notify(129414)
	var sl []ReplicationTarget
	for _, chg := range rc {
		__antithesis_instrumentation__.Notify(129416)
		if chg.ChangeType == typ {
			__antithesis_instrumentation__.Notify(129417)
			sl = append(sl, chg.Target)
		} else {
			__antithesis_instrumentation__.Notify(129418)
		}
	}
	__antithesis_instrumentation__.Notify(129415)
	return sl
}

func (rc ReplicationChanges) VoterAdditions() []ReplicationTarget {
	__antithesis_instrumentation__.Notify(129419)
	return rc.byType(ADD_VOTER)
}

func (rc ReplicationChanges) VoterRemovals() []ReplicationTarget {
	__antithesis_instrumentation__.Notify(129420)
	return rc.byType(REMOVE_VOTER)
}

func (rc ReplicationChanges) NonVoterAdditions() []ReplicationTarget {
	__antithesis_instrumentation__.Notify(129421)
	return rc.byType(ADD_NON_VOTER)
}

func (rc ReplicationChanges) NonVoterRemovals() []ReplicationTarget {
	__antithesis_instrumentation__.Notify(129422)
	return rc.byType(REMOVE_NON_VOTER)
}

func (acrr *AdminChangeReplicasRequest) Changes() []ReplicationChange {
	__antithesis_instrumentation__.Notify(129423)
	if len(acrr.InternalChanges) > 0 {
		__antithesis_instrumentation__.Notify(129426)
		return acrr.InternalChanges
	} else {
		__antithesis_instrumentation__.Notify(129427)
	}
	__antithesis_instrumentation__.Notify(129424)

	sl := make([]ReplicationChange, len(acrr.DeprecatedTargets))
	for _, target := range acrr.DeprecatedTargets {
		__antithesis_instrumentation__.Notify(129428)
		sl = append(sl, ReplicationChange{
			ChangeType: acrr.DeprecatedChangeType,
			Target:     target,
		})
	}
	__antithesis_instrumentation__.Notify(129425)
	return sl
}

func (rir *ResolveIntentRequest) AsLockUpdate() LockUpdate {
	__antithesis_instrumentation__.Notify(129429)
	return LockUpdate{
		Span:           rir.Span(),
		Txn:            rir.IntentTxn,
		Status:         rir.Status,
		IgnoredSeqNums: rir.IgnoredSeqNums,
	}
}

func (rirr *ResolveIntentRangeRequest) AsLockUpdate() LockUpdate {
	__antithesis_instrumentation__.Notify(129430)
	return LockUpdate{
		Span:           rirr.Span(),
		Txn:            rirr.IntentTxn,
		Status:         rirr.Status,
		IgnoredSeqNums: rirr.IgnoredSeqNums,
	}
}

func (r *JoinNodeResponse) CreateStoreIdent() (StoreIdent, error) {
	__antithesis_instrumentation__.Notify(129431)
	nodeID, storeID := NodeID(r.NodeID), StoreID(r.StoreID)
	clusterID, err := uuid.FromBytes(r.ClusterID)
	if err != nil {
		__antithesis_instrumentation__.Notify(129433)
		return StoreIdent{}, err
	} else {
		__antithesis_instrumentation__.Notify(129434)
	}
	__antithesis_instrumentation__.Notify(129432)

	sIdent := StoreIdent{
		ClusterID: clusterID,
		NodeID:    nodeID,
		StoreID:   storeID,
	}
	return sIdent, nil
}

func (c *ContentionEvent) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(129435)
	w.Printf("conflicted with %s on %s for %.3fs", c.TxnMeta.ID, c.Key, c.Duration.Seconds())
}

func (c *ContentionEvent) String() string {
	__antithesis_instrumentation__.Notify(129436)
	return redact.StringWithoutMarkers(c)
}

func (c *TenantConsumption) Equal(other *TenantConsumption) bool {
	__antithesis_instrumentation__.Notify(129437)
	return *c == *other
}

var _ = (*TenantConsumption).Equal

func (c *TenantConsumption) Add(other *TenantConsumption) {
	__antithesis_instrumentation__.Notify(129438)
	c.RU += other.RU
	c.ReadRequests += other.ReadRequests
	c.ReadBytes += other.ReadBytes
	c.WriteRequests += other.WriteRequests
	c.WriteBytes += other.WriteBytes
	c.SQLPodsCPUSeconds += other.SQLPodsCPUSeconds
	c.PGWireEgressBytes += other.PGWireEgressBytes
}

func (c *TenantConsumption) Sub(other *TenantConsumption) {
	__antithesis_instrumentation__.Notify(129439)
	if c.RU < other.RU {
		__antithesis_instrumentation__.Notify(129446)
		c.RU = 0
	} else {
		__antithesis_instrumentation__.Notify(129447)
		c.RU -= other.RU
	}
	__antithesis_instrumentation__.Notify(129440)

	if c.ReadRequests < other.ReadRequests {
		__antithesis_instrumentation__.Notify(129448)
		c.ReadRequests = 0
	} else {
		__antithesis_instrumentation__.Notify(129449)
		c.ReadRequests -= other.ReadRequests
	}
	__antithesis_instrumentation__.Notify(129441)

	if c.ReadBytes < other.ReadBytes {
		__antithesis_instrumentation__.Notify(129450)
		c.ReadBytes = 0
	} else {
		__antithesis_instrumentation__.Notify(129451)
		c.ReadBytes -= other.ReadBytes
	}
	__antithesis_instrumentation__.Notify(129442)

	if c.WriteRequests < other.WriteRequests {
		__antithesis_instrumentation__.Notify(129452)
		c.WriteRequests = 0
	} else {
		__antithesis_instrumentation__.Notify(129453)
		c.WriteRequests -= other.WriteRequests
	}
	__antithesis_instrumentation__.Notify(129443)

	if c.WriteBytes < other.WriteBytes {
		__antithesis_instrumentation__.Notify(129454)
		c.WriteBytes = 0
	} else {
		__antithesis_instrumentation__.Notify(129455)
		c.WriteBytes -= other.WriteBytes
	}
	__antithesis_instrumentation__.Notify(129444)

	if c.SQLPodsCPUSeconds < other.SQLPodsCPUSeconds {
		__antithesis_instrumentation__.Notify(129456)
		c.SQLPodsCPUSeconds = 0
	} else {
		__antithesis_instrumentation__.Notify(129457)
		c.SQLPodsCPUSeconds -= other.SQLPodsCPUSeconds
	}
	__antithesis_instrumentation__.Notify(129445)

	if c.PGWireEgressBytes < other.PGWireEgressBytes {
		__antithesis_instrumentation__.Notify(129458)
		c.PGWireEgressBytes = 0
	} else {
		__antithesis_instrumentation__.Notify(129459)
		c.PGWireEgressBytes -= other.PGWireEgressBytes
	}
}

func humanizePointCount(n uint64) redact.SafeString {
	__antithesis_instrumentation__.Notify(129460)
	return redact.SafeString(humanize.SI(float64(n), ""))
}

func (s *ScanStats) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(129461)
	w.Printf("scan stats: stepped %d times (%d internal); seeked %d times (%d internal); "+
		"block-bytes: (total %s, cached %s); "+
		"points: (count %s, key-bytes %s, value-bytes %s, tombstoned: %s)",
		s.NumInterfaceSteps, s.NumInternalSteps, s.NumInterfaceSeeks, s.NumInternalSeeks,
		humanizeutil.IBytes(int64(s.BlockBytes)),
		humanizeutil.IBytes(int64(s.BlockBytesInCache)),
		humanizePointCount(s.PointCount),
		humanizeutil.IBytes(int64(s.KeyBytes)),
		humanizeutil.IBytes(int64(s.ValueBytes)),
		humanizePointCount(s.PointsCoveredByRangeTombstones))
}

func (s *ScanStats) String() string {
	__antithesis_instrumentation__.Notify(129462)
	return redact.StringWithoutMarkers(s)
}

type TenantSettingsPrecedence uint32

const (
	SpecificTenantOverrides TenantSettingsPrecedence = 1 + iota

	AllTenantsOverrides
)
