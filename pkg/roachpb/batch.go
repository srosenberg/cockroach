package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func (h Header) WriteTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(157926)
	ts := h.Timestamp
	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(157928)
		ts.Forward(h.Txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(157929)
	}
	__antithesis_instrumentation__.Notify(157927)
	return ts
}

func (h Header) RequiredFrontier() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(157930)
	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(157932)
		return h.Txn.RequiredFrontier()
	} else {
		__antithesis_instrumentation__.Notify(157933)
	}
	__antithesis_instrumentation__.Notify(157931)
	return h.Timestamp
}

func (ba *BatchRequest) SetActiveTimestamp(clock *hlc.Clock) error {
	__antithesis_instrumentation__.Notify(157934)
	if txn := ba.Txn; txn != nil {
		__antithesis_instrumentation__.Notify(157936)
		if !ba.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(157938)
			return errors.New("transactional request must not set batch timestamp")
		} else {
			__antithesis_instrumentation__.Notify(157939)
		}
		__antithesis_instrumentation__.Notify(157937)

		ba.Timestamp = txn.ReadTimestamp
	} else {
		__antithesis_instrumentation__.Notify(157940)

		if ba.Timestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(157941)
			now := clock.NowAsClockTimestamp()
			ba.Timestamp = now.ToTimestamp()
			ba.TimestampFromServerClock = &now
		} else {
			__antithesis_instrumentation__.Notify(157942)
		}
	}
	__antithesis_instrumentation__.Notify(157935)
	return nil
}

func (ba BatchRequest) EarliestActiveTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(157943)
	ts := ba.Timestamp
	for _, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(157945)
		switch t := ru.GetInner().(type) {
		case *ExportRequest:
			__antithesis_instrumentation__.Notify(157946)
			if !t.StartTime.IsEmpty() {
				__antithesis_instrumentation__.Notify(157950)

				ts.Backward(t.StartTime.Next())
			} else {
				__antithesis_instrumentation__.Notify(157951)
			}
		case *RevertRangeRequest:
			__antithesis_instrumentation__.Notify(157947)

			if !t.IgnoreGcThreshold {
				__antithesis_instrumentation__.Notify(157952)
				ts.Backward(t.TargetTime)
			} else {
				__antithesis_instrumentation__.Notify(157953)
			}
		case *RefreshRequest:
			__antithesis_instrumentation__.Notify(157948)

			ts.Backward(t.RefreshFrom.Next())
		case *RefreshRangeRequest:
			__antithesis_instrumentation__.Notify(157949)

			ts.Backward(t.RefreshFrom.Next())
		}
	}
	__antithesis_instrumentation__.Notify(157944)
	return ts
}

func (ba *BatchRequest) UpdateTxn(o *Transaction) {
	__antithesis_instrumentation__.Notify(157954)
	if o == nil {
		__antithesis_instrumentation__.Notify(157957)
		return
	} else {
		__antithesis_instrumentation__.Notify(157958)
	}
	__antithesis_instrumentation__.Notify(157955)
	o.AssertInitialized(context.TODO())
	if ba.Txn == nil {
		__antithesis_instrumentation__.Notify(157959)
		ba.Txn = o
		return
	} else {
		__antithesis_instrumentation__.Notify(157960)
	}
	__antithesis_instrumentation__.Notify(157956)
	clonedTxn := ba.Txn.Clone()
	clonedTxn.Update(o)
	ba.Txn = clonedTxn
}

func (ba *BatchRequest) IsLeaseRequest() bool {
	__antithesis_instrumentation__.Notify(157961)
	if !ba.IsSingleRequest() {
		__antithesis_instrumentation__.Notify(157963)
		return false
	} else {
		__antithesis_instrumentation__.Notify(157964)
	}
	__antithesis_instrumentation__.Notify(157962)
	_, ok := ba.GetArg(RequestLease)
	return ok
}

func (ba *BatchRequest) AppliesTimestampCache() bool {
	__antithesis_instrumentation__.Notify(157965)
	return ba.hasFlag(appliesTSCache)
}

func (ba *BatchRequest) IsAdmin() bool {
	__antithesis_instrumentation__.Notify(157966)
	return ba.hasFlag(isAdmin)
}

func (ba *BatchRequest) IsWrite() bool {
	__antithesis_instrumentation__.Notify(157967)
	return ba.hasFlag(isWrite)
}

func (ba *BatchRequest) IsReadOnly() bool {
	__antithesis_instrumentation__.Notify(157968)
	return len(ba.Requests) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(157969)
		return !ba.hasFlag(isWrite|isAdmin) == true
	}() == true
}

func (ba *BatchRequest) IsReverse() bool {
	__antithesis_instrumentation__.Notify(157970)
	return ba.hasFlag(isReverse)
}

func (ba *BatchRequest) IsTransactional() bool {
	__antithesis_instrumentation__.Notify(157971)
	return ba.hasFlag(isTxn)
}

func (ba *BatchRequest) IsAllTransactional() bool {
	__antithesis_instrumentation__.Notify(157972)
	return ba.hasFlagForAll(isTxn)
}

func (ba *BatchRequest) IsLocking() bool {
	__antithesis_instrumentation__.Notify(157973)
	return ba.hasFlag(isLocking)
}

func (ba *BatchRequest) IsIntentWrite() bool {
	__antithesis_instrumentation__.Notify(157974)
	return ba.hasFlag(isIntentWrite)
}

func (ba *BatchRequest) IsUnsplittable() bool {
	__antithesis_instrumentation__.Notify(157975)
	return ba.hasFlag(isUnsplittable)
}

func (ba *BatchRequest) IsSingleRequest() bool {
	__antithesis_instrumentation__.Notify(157976)
	return len(ba.Requests) == 1
}

func (ba *BatchRequest) IsSingleSkipsLeaseCheckRequest() bool {
	__antithesis_instrumentation__.Notify(157977)
	return ba.IsSingleRequest() && func() bool {
		__antithesis_instrumentation__.Notify(157978)
		return ba.hasFlag(skipsLeaseCheck) == true
	}() == true
}

func (ba *BatchRequest) isSingleRequestWithMethod(m Method) bool {
	__antithesis_instrumentation__.Notify(157979)
	return ba.IsSingleRequest() && func() bool {
		__antithesis_instrumentation__.Notify(157980)
		return ba.Requests[0].GetInner().Method() == m == true
	}() == true
}

func (ba *BatchRequest) IsSingleTransferLeaseRequest() bool {
	__antithesis_instrumentation__.Notify(157981)
	return ba.isSingleRequestWithMethod(TransferLease)
}

func (ba *BatchRequest) IsSingleLeaseInfoRequest() bool {
	__antithesis_instrumentation__.Notify(157982)
	return ba.isSingleRequestWithMethod(LeaseInfo)
}

func (ba *BatchRequest) IsSingleProbeRequest() bool {
	__antithesis_instrumentation__.Notify(157983)
	return ba.isSingleRequestWithMethod(Probe)
}

func (ba *BatchRequest) IsSinglePushTxnRequest() bool {
	__antithesis_instrumentation__.Notify(157984)
	return ba.isSingleRequestWithMethod(PushTxn)
}

func (ba *BatchRequest) IsSingleHeartbeatTxnRequest() bool {
	__antithesis_instrumentation__.Notify(157985)
	return ba.isSingleRequestWithMethod(HeartbeatTxn)
}

func (ba *BatchRequest) IsSingleEndTxnRequest() bool {
	__antithesis_instrumentation__.Notify(157986)
	return ba.isSingleRequestWithMethod(EndTxn)
}

func (ba *BatchRequest) Require1PC() bool {
	__antithesis_instrumentation__.Notify(157987)
	arg, ok := ba.GetArg(EndTxn)
	if !ok {
		__antithesis_instrumentation__.Notify(157989)
		return false
	} else {
		__antithesis_instrumentation__.Notify(157990)
	}
	__antithesis_instrumentation__.Notify(157988)
	etArg := arg.(*EndTxnRequest)
	return etArg.Require1PC
}

func (ba *BatchRequest) IsSingleAbortTxnRequest() bool {
	__antithesis_instrumentation__.Notify(157991)
	if ba.isSingleRequestWithMethod(EndTxn) {
		__antithesis_instrumentation__.Notify(157993)
		return !ba.Requests[0].GetInner().(*EndTxnRequest).Commit
	} else {
		__antithesis_instrumentation__.Notify(157994)
	}
	__antithesis_instrumentation__.Notify(157992)
	return false
}

func (ba *BatchRequest) IsSingleCommitRequest() bool {
	__antithesis_instrumentation__.Notify(157995)
	if ba.isSingleRequestWithMethod(EndTxn) {
		__antithesis_instrumentation__.Notify(157997)
		return ba.Requests[0].GetInner().(*EndTxnRequest).Commit
	} else {
		__antithesis_instrumentation__.Notify(157998)
	}
	__antithesis_instrumentation__.Notify(157996)
	return false
}

func (ba *BatchRequest) IsSingleRefreshRequest() bool {
	__antithesis_instrumentation__.Notify(157999)
	return ba.isSingleRequestWithMethod(Refresh)
}

func (ba *BatchRequest) IsSingleSubsumeRequest() bool {
	__antithesis_instrumentation__.Notify(158000)
	return ba.isSingleRequestWithMethod(Subsume)
}

func (ba *BatchRequest) IsSingleComputeChecksumRequest() bool {
	__antithesis_instrumentation__.Notify(158001)
	return ba.isSingleRequestWithMethod(ComputeChecksum)
}

func (ba *BatchRequest) IsSingleCheckConsistencyRequest() bool {
	__antithesis_instrumentation__.Notify(158002)
	return ba.isSingleRequestWithMethod(CheckConsistency)
}

func (ba *BatchRequest) RequiresConsensus() bool {
	__antithesis_instrumentation__.Notify(158003)
	return ba.isSingleRequestWithMethod(Barrier) || func() bool {
		__antithesis_instrumentation__.Notify(158004)
		return ba.isSingleRequestWithMethod(Probe) == true
	}() == true
}

func (ba *BatchRequest) IsCompleteTransaction() bool {
	__antithesis_instrumentation__.Notify(158005)
	et, hasET := ba.GetArg(EndTxn)
	if !hasET || func() bool {
		__antithesis_instrumentation__.Notify(158010)
		return !et.(*EndTxnRequest).Commit == true
	}() == true {
		__antithesis_instrumentation__.Notify(158011)
		return false
	} else {
		__antithesis_instrumentation__.Notify(158012)
	}
	__antithesis_instrumentation__.Notify(158006)
	maxSeq := et.Header().Sequence
	switch maxSeq {
	case 0:
		__antithesis_instrumentation__.Notify(158013)

		return false
	case 1:
		__antithesis_instrumentation__.Notify(158014)

		return true
	default:
		__antithesis_instrumentation__.Notify(158015)
	}
	__antithesis_instrumentation__.Notify(158007)
	if int(maxSeq) > len(ba.Requests) {
		__antithesis_instrumentation__.Notify(158016)

		return false
	} else {
		__antithesis_instrumentation__.Notify(158017)
	}
	__antithesis_instrumentation__.Notify(158008)

	nextSeq := enginepb.TxnSeq(1)
	for _, args := range ba.Requests {
		__antithesis_instrumentation__.Notify(158018)
		req := args.GetInner()
		seq := req.Header().Sequence
		if seq > nextSeq {
			__antithesis_instrumentation__.Notify(158020)
			return false
		} else {
			__antithesis_instrumentation__.Notify(158021)
		}
		__antithesis_instrumentation__.Notify(158019)
		if seq == nextSeq {
			__antithesis_instrumentation__.Notify(158022)
			if !IsIntentWrite(req) {
				__antithesis_instrumentation__.Notify(158024)
				return false
			} else {
				__antithesis_instrumentation__.Notify(158025)
			}
			__antithesis_instrumentation__.Notify(158023)
			nextSeq++
			if nextSeq == maxSeq {
				__antithesis_instrumentation__.Notify(158026)
				return true
			} else {
				__antithesis_instrumentation__.Notify(158027)
			}
		} else {
			__antithesis_instrumentation__.Notify(158028)
		}
	}
	__antithesis_instrumentation__.Notify(158009)
	panic("unreachable")
}

func (ba *BatchRequest) hasFlag(flag flag) bool {
	__antithesis_instrumentation__.Notify(158029)
	for _, union := range ba.Requests {
		__antithesis_instrumentation__.Notify(158031)
		if (union.GetInner().flags() & flag) != 0 {
			__antithesis_instrumentation__.Notify(158032)
			return true
		} else {
			__antithesis_instrumentation__.Notify(158033)
		}
	}
	__antithesis_instrumentation__.Notify(158030)
	return false
}

func (ba *BatchRequest) hasFlagForAll(flag flag) bool {
	__antithesis_instrumentation__.Notify(158034)
	if len(ba.Requests) == 0 {
		__antithesis_instrumentation__.Notify(158037)
		return false
	} else {
		__antithesis_instrumentation__.Notify(158038)
	}
	__antithesis_instrumentation__.Notify(158035)
	for _, union := range ba.Requests {
		__antithesis_instrumentation__.Notify(158039)
		if (union.GetInner().flags() & flag) == 0 {
			__antithesis_instrumentation__.Notify(158040)
			return false
		} else {
			__antithesis_instrumentation__.Notify(158041)
		}
	}
	__antithesis_instrumentation__.Notify(158036)
	return true
}

func (ba *BatchRequest) GetArg(method Method) (Request, bool) {
	__antithesis_instrumentation__.Notify(158042)

	if method == EndTxn {
		__antithesis_instrumentation__.Notify(158045)
		if length := len(ba.Requests); length > 0 {
			__antithesis_instrumentation__.Notify(158047)
			if req := ba.Requests[length-1].GetInner(); req.Method() == EndTxn {
				__antithesis_instrumentation__.Notify(158048)
				return req, true
			} else {
				__antithesis_instrumentation__.Notify(158049)
			}
		} else {
			__antithesis_instrumentation__.Notify(158050)
		}
		__antithesis_instrumentation__.Notify(158046)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(158051)
	}
	__antithesis_instrumentation__.Notify(158043)

	for _, arg := range ba.Requests {
		__antithesis_instrumentation__.Notify(158052)
		if req := arg.GetInner(); req.Method() == method {
			__antithesis_instrumentation__.Notify(158053)
			return req, true
		} else {
			__antithesis_instrumentation__.Notify(158054)
		}
	}
	__antithesis_instrumentation__.Notify(158044)
	return nil, false
}

func (br *BatchResponse) String() string {
	__antithesis_instrumentation__.Notify(158055)
	var str []string
	str = append(str, fmt.Sprintf("(err: %v)", br.Error))
	for count, union := range br.Responses {
		__antithesis_instrumentation__.Notify(158057)

		if count >= 20 && func() bool {
			__antithesis_instrumentation__.Notify(158059)
			return count < len(br.Responses)-5 == true
		}() == true {
			__antithesis_instrumentation__.Notify(158060)
			if count == 20 {
				__antithesis_instrumentation__.Notify(158062)
				str = append(str, fmt.Sprintf("... %d skipped ...", len(br.Responses)-25))
			} else {
				__antithesis_instrumentation__.Notify(158063)
			}
			__antithesis_instrumentation__.Notify(158061)
			continue
		} else {
			__antithesis_instrumentation__.Notify(158064)
		}
		__antithesis_instrumentation__.Notify(158058)
		str = append(str, fmt.Sprintf("%T", union.GetInner()))
	}
	__antithesis_instrumentation__.Notify(158056)
	return strings.Join(str, ", ")
}

func (ba *BatchRequest) LockSpanIterate(br *BatchResponse, fn func(Span, lock.Durability)) {
	__antithesis_instrumentation__.Notify(158065)
	for i, arg := range ba.Requests {
		__antithesis_instrumentation__.Notify(158066)
		req := arg.GetInner()
		if !IsLocking(req) {
			__antithesis_instrumentation__.Notify(158069)
			continue
		} else {
			__antithesis_instrumentation__.Notify(158070)
		}
		__antithesis_instrumentation__.Notify(158067)
		var resp Response
		if br != nil {
			__antithesis_instrumentation__.Notify(158071)
			resp = br.Responses[i].GetInner()
		} else {
			__antithesis_instrumentation__.Notify(158072)
		}
		__antithesis_instrumentation__.Notify(158068)
		if span, ok := ActualSpan(req, resp); ok {
			__antithesis_instrumentation__.Notify(158073)
			fn(span, LockingDurability(req))
		} else {
			__antithesis_instrumentation__.Notify(158074)
		}
	}
}

func (ba *BatchRequest) RefreshSpanIterate(br *BatchResponse, fn func(Span)) {
	__antithesis_instrumentation__.Notify(158075)
	for i, arg := range ba.Requests {
		__antithesis_instrumentation__.Notify(158076)
		req := arg.GetInner()
		if !NeedsRefresh(req) {
			__antithesis_instrumentation__.Notify(158079)
			continue
		} else {
			__antithesis_instrumentation__.Notify(158080)
		}
		__antithesis_instrumentation__.Notify(158077)
		var resp Response
		if br != nil {
			__antithesis_instrumentation__.Notify(158081)
			resp = br.Responses[i].GetInner()
		} else {
			__antithesis_instrumentation__.Notify(158082)
		}
		__antithesis_instrumentation__.Notify(158078)
		if span, ok := ActualSpan(req, resp); ok {
			__antithesis_instrumentation__.Notify(158083)
			fn(span)
		} else {
			__antithesis_instrumentation__.Notify(158084)
		}
	}
}

func ActualSpan(req Request, resp Response) (Span, bool) {
	__antithesis_instrumentation__.Notify(158085)
	h := req.Header()
	if resp != nil {
		__antithesis_instrumentation__.Notify(158087)
		resumeSpan := resp.Header().ResumeSpan

		if resumeSpan != nil {
			__antithesis_instrumentation__.Notify(158088)

			if bytes.Equal(resumeSpan.Key, h.Key) {
				__antithesis_instrumentation__.Notify(158090)
				if bytes.Equal(resumeSpan.EndKey, h.EndKey) {
					__antithesis_instrumentation__.Notify(158092)
					return Span{}, false
				} else {
					__antithesis_instrumentation__.Notify(158093)
				}
				__antithesis_instrumentation__.Notify(158091)
				return Span{Key: resumeSpan.EndKey, EndKey: h.EndKey}, true
			} else {
				__antithesis_instrumentation__.Notify(158094)
			}
			__antithesis_instrumentation__.Notify(158089)

			return Span{Key: h.Key, EndKey: resumeSpan.Key}, true
		} else {
			__antithesis_instrumentation__.Notify(158095)
		}
	} else {
		__antithesis_instrumentation__.Notify(158096)
	}
	__antithesis_instrumentation__.Notify(158086)
	return h.Span(), true
}

func (br *BatchResponse) Combine(otherBatch *BatchResponse, positions []int) error {
	__antithesis_instrumentation__.Notify(158097)
	if err := br.BatchResponse_Header.combine(otherBatch.BatchResponse_Header); err != nil {
		__antithesis_instrumentation__.Notify(158100)
		return err
	} else {
		__antithesis_instrumentation__.Notify(158101)
	}
	__antithesis_instrumentation__.Notify(158098)
	for i := range otherBatch.Responses {
		__antithesis_instrumentation__.Notify(158102)
		pos := positions[i]
		if br.Responses[pos] == (ResponseUnion{}) {
			__antithesis_instrumentation__.Notify(158104)
			br.Responses[pos] = otherBatch.Responses[i]
			continue
		} else {
			__antithesis_instrumentation__.Notify(158105)
		}
		__antithesis_instrumentation__.Notify(158103)
		valLeft := br.Responses[pos].GetInner()
		valRight := otherBatch.Responses[i].GetInner()
		if err := CombineResponses(valLeft, valRight); err != nil {
			__antithesis_instrumentation__.Notify(158106)
			return err
		} else {
			__antithesis_instrumentation__.Notify(158107)
		}
	}
	__antithesis_instrumentation__.Notify(158099)
	return nil
}

func (ba *BatchRequest) Add(requests ...Request) {
	__antithesis_instrumentation__.Notify(158108)
	for _, args := range requests {
		__antithesis_instrumentation__.Notify(158109)
		ba.Requests = append(ba.Requests, RequestUnion{})
		ba.Requests[len(ba.Requests)-1].MustSetInner(args)
	}
}

func (br *BatchResponse) Add(reply Response) {
	__antithesis_instrumentation__.Notify(158110)
	br.Responses = append(br.Responses, ResponseUnion{})
	br.Responses[len(br.Responses)-1].MustSetInner(reply)
}

func (ba *BatchRequest) Methods() []Method {
	__antithesis_instrumentation__.Notify(158111)
	var res []Method
	for _, arg := range ba.Requests {
		__antithesis_instrumentation__.Notify(158113)
		res = append(res, arg.GetInner().Method())
	}
	__antithesis_instrumentation__.Notify(158112)
	return res
}

func (ba BatchRequest) Split(canSplitET bool) [][]RequestUnion {
	__antithesis_instrumentation__.Notify(158114)
	compatible := func(exFlags, newFlags flag) bool {
		__antithesis_instrumentation__.Notify(158117)

		if (exFlags&isAlone) != 0 || func() bool {
			__antithesis_instrumentation__.Notify(158121)
			return (newFlags & isAlone) != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(158122)
			return false
		} else {
			__antithesis_instrumentation__.Notify(158123)
		}
		__antithesis_instrumentation__.Notify(158118)

		if exFlags == 0 || func() bool {
			__antithesis_instrumentation__.Notify(158124)
			return newFlags == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(158125)
			return true
		} else {
			__antithesis_instrumentation__.Notify(158126)
		}
		__antithesis_instrumentation__.Notify(158119)

		mask := isWrite | isAdmin
		if (exFlags&isRange) != 0 && func() bool {
			__antithesis_instrumentation__.Notify(158127)
			return (newFlags & isRange) != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(158128)

			mask |= isReverse
		} else {
			__antithesis_instrumentation__.Notify(158129)
		}
		__antithesis_instrumentation__.Notify(158120)
		return (mask & exFlags) == (mask & newFlags)
	}
	__antithesis_instrumentation__.Notify(158115)
	var parts [][]RequestUnion
	for len(ba.Requests) > 0 {
		__antithesis_instrumentation__.Notify(158130)
		part := ba.Requests
		var gFlags, hFlags flag = -1, -1
		for i, union := range ba.Requests {
			__antithesis_instrumentation__.Notify(158132)
			args := union.GetInner()
			flags := args.flags()
			method := args.Method()
			if (flags & isPrefix) != 0 {
				__antithesis_instrumentation__.Notify(158135)

				if hFlags == -1 {
					__antithesis_instrumentation__.Notify(158137)
					for _, nUnion := range ba.Requests[i+1:] {
						__antithesis_instrumentation__.Notify(158138)
						nArgs := nUnion.GetInner()
						nFlags := nArgs.flags()
						nMethod := nArgs.Method()
						if !canSplitET && func() bool {
							__antithesis_instrumentation__.Notify(158140)
							return nMethod == EndTxn == true
						}() == true {
							__antithesis_instrumentation__.Notify(158141)
							nFlags = 0
						} else {
							__antithesis_instrumentation__.Notify(158142)
						}
						__antithesis_instrumentation__.Notify(158139)
						if (nFlags & isPrefix) == 0 {
							__antithesis_instrumentation__.Notify(158143)
							hFlags = nFlags
							break
						} else {
							__antithesis_instrumentation__.Notify(158144)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(158145)
				}
				__antithesis_instrumentation__.Notify(158136)
				if hFlags != -1 && func() bool {
					__antithesis_instrumentation__.Notify(158146)
					return (hFlags & isAlone) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(158147)
					flags = hFlags
				} else {
					__antithesis_instrumentation__.Notify(158148)
				}
			} else {
				__antithesis_instrumentation__.Notify(158149)
				hFlags = -1
			}
			__antithesis_instrumentation__.Notify(158133)
			cmpFlags := flags
			if !canSplitET && func() bool {
				__antithesis_instrumentation__.Notify(158150)
				return method == EndTxn == true
			}() == true {
				__antithesis_instrumentation__.Notify(158151)
				cmpFlags = 0
			} else {
				__antithesis_instrumentation__.Notify(158152)
			}
			__antithesis_instrumentation__.Notify(158134)
			if gFlags == -1 {
				__antithesis_instrumentation__.Notify(158153)

				gFlags = flags
			} else {
				__antithesis_instrumentation__.Notify(158154)
				if !compatible(gFlags, cmpFlags) {
					__antithesis_instrumentation__.Notify(158156)
					part = ba.Requests[:i]
					break
				} else {
					__antithesis_instrumentation__.Notify(158157)
				}
				__antithesis_instrumentation__.Notify(158155)
				gFlags |= flags
			}
		}
		__antithesis_instrumentation__.Notify(158131)
		parts = append(parts, part)
		ba.Requests = ba.Requests[len(part):]
	}
	__antithesis_instrumentation__.Notify(158116)
	return parts
}

func (ba BatchRequest) SafeFormat(s redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(158158)
	for count, arg := range ba.Requests {
		__antithesis_instrumentation__.Notify(158162)

		if count >= 20 && func() bool {
			__antithesis_instrumentation__.Notify(158165)
			return count < len(ba.Requests)-5 == true
		}() == true {
			__antithesis_instrumentation__.Notify(158166)
			if count == 20 {
				__antithesis_instrumentation__.Notify(158168)
				s.Printf(",... %d skipped ...", len(ba.Requests)-25)
			} else {
				__antithesis_instrumentation__.Notify(158169)
			}
			__antithesis_instrumentation__.Notify(158167)
			continue
		} else {
			__antithesis_instrumentation__.Notify(158170)
		}
		__antithesis_instrumentation__.Notify(158163)
		if count > 0 {
			__antithesis_instrumentation__.Notify(158171)
			s.Print(redact.SafeString(", "))
		} else {
			__antithesis_instrumentation__.Notify(158172)
		}
		__antithesis_instrumentation__.Notify(158164)

		req := arg.GetInner()
		if et, ok := req.(*EndTxnRequest); ok {
			__antithesis_instrumentation__.Notify(158173)
			h := req.Header()
			s.Printf("%s(", req.Method())
			if et.Commit {
				__antithesis_instrumentation__.Notify(158176)
				if et.IsParallelCommit() {
					__antithesis_instrumentation__.Notify(158177)
					s.Printf("parallel commit")
				} else {
					__antithesis_instrumentation__.Notify(158178)
					s.Printf("commit")
				}
			} else {
				__antithesis_instrumentation__.Notify(158179)
				s.Printf("abort")
			}
			__antithesis_instrumentation__.Notify(158174)
			if et.InternalCommitTrigger != nil {
				__antithesis_instrumentation__.Notify(158180)
				s.Printf(" %s", et.InternalCommitTrigger.Kind())
			} else {
				__antithesis_instrumentation__.Notify(158181)
			}
			__antithesis_instrumentation__.Notify(158175)
			s.Printf(") [%s]", h.Key)
		} else {
			__antithesis_instrumentation__.Notify(158182)
			h := req.Header()
			if req.Method() == PushTxn {
				__antithesis_instrumentation__.Notify(158184)
				pushReq := req.(*PushTxnRequest)
				s.Printf("PushTxn(%s->%s)", pushReq.PusherTxn.Short(), pushReq.PusheeTxn.Short())
			} else {
				__antithesis_instrumentation__.Notify(158185)
				s.Print(req.Method())
			}
			__antithesis_instrumentation__.Notify(158183)
			s.Printf(" [%s,%s)", h.Key, h.EndKey)
		}
	}
	{
		__antithesis_instrumentation__.Notify(158186)
		if ba.Txn != nil {
			__antithesis_instrumentation__.Notify(158187)
			s.Printf(", [txn: %s]", ba.Txn.Short())
		} else {
			__antithesis_instrumentation__.Notify(158188)
		}
	}
	__antithesis_instrumentation__.Notify(158159)
	if ba.WaitPolicy != lock.WaitPolicy_Block {
		__antithesis_instrumentation__.Notify(158189)
		s.Printf(", [wait-policy: %s]", ba.WaitPolicy)
	} else {
		__antithesis_instrumentation__.Notify(158190)
	}
	__antithesis_instrumentation__.Notify(158160)
	if ba.CanForwardReadTimestamp {
		__antithesis_instrumentation__.Notify(158191)
		s.Printf(", [can-forward-ts]")
	} else {
		__antithesis_instrumentation__.Notify(158192)
	}
	__antithesis_instrumentation__.Notify(158161)
	if cfg := ba.BoundedStaleness; cfg != nil {
		__antithesis_instrumentation__.Notify(158193)
		s.Printf(", [bounded-staleness, min_ts_bound: %s", cfg.MinTimestampBound)
		if cfg.MinTimestampBoundStrict {
			__antithesis_instrumentation__.Notify(158196)
			s.Printf(", min_ts_bound_strict")
		} else {
			__antithesis_instrumentation__.Notify(158197)
		}
		__antithesis_instrumentation__.Notify(158194)
		if !cfg.MaxTimestampBound.IsEmpty() {
			__antithesis_instrumentation__.Notify(158198)
			s.Printf(", max_ts_bound: %s", cfg.MaxTimestampBound)
		} else {
			__antithesis_instrumentation__.Notify(158199)
		}
		__antithesis_instrumentation__.Notify(158195)
		s.Printf("]")
	} else {
		__antithesis_instrumentation__.Notify(158200)
	}
}

func (ba BatchRequest) String() string {
	__antithesis_instrumentation__.Notify(158201)
	return redact.StringWithoutMarkers(ba)
}

func (ba BatchRequest) ValidateForEvaluation() error {
	__antithesis_instrumentation__.Notify(158202)
	if ba.RangeID == 0 {
		__antithesis_instrumentation__.Notify(158206)
		return errors.AssertionFailedf("batch request missing range ID")
	} else {
		__antithesis_instrumentation__.Notify(158207)
		if ba.Replica.StoreID == 0 {
			__antithesis_instrumentation__.Notify(158208)
			return errors.AssertionFailedf("batch request missing store ID")
		} else {
			__antithesis_instrumentation__.Notify(158209)
		}
	}
	__antithesis_instrumentation__.Notify(158203)
	if _, ok := ba.GetArg(EndTxn); ok && func() bool {
		__antithesis_instrumentation__.Notify(158210)
		return ba.Txn == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(158211)
		return errors.AssertionFailedf("EndTxn request without transaction")
	} else {
		__antithesis_instrumentation__.Notify(158212)
	}
	__antithesis_instrumentation__.Notify(158204)
	if ba.Txn != nil {
		__antithesis_instrumentation__.Notify(158213)
		if ba.Txn.WriteTooOld && func() bool {
			__antithesis_instrumentation__.Notify(158214)
			return ba.Txn.ReadTimestamp == ba.Txn.WriteTimestamp == true
		}() == true {
			__antithesis_instrumentation__.Notify(158215)
			return errors.AssertionFailedf("WriteTooOld set but no offset in timestamps. txn: %s", ba.Txn)
		} else {
			__antithesis_instrumentation__.Notify(158216)
		}
	} else {
		__antithesis_instrumentation__.Notify(158217)
	}
	__antithesis_instrumentation__.Notify(158205)
	return nil
}
