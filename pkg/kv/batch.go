package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

const (
	raw    = true
	notRaw = false
)

type Batch struct {
	txn *Txn

	Results []Result

	Header roachpb.Header

	AdmissionHeader roachpb.AdmissionHeader
	reqs            []roachpb.RequestUnion

	approxMutationReqBytes int

	raw bool

	response *roachpb.BatchResponse

	pErr *roachpb.Error

	resultsBuf    [8]Result
	rowsBuf       []KeyValue
	rowsStaticBuf [8]KeyValue
	rowsStaticIdx int
}

func (b *Batch) ApproximateMutationBytes() int {
	__antithesis_instrumentation__.Notify(85927)
	return b.approxMutationReqBytes
}

func (b *Batch) RawResponse() *roachpb.BatchResponse {
	__antithesis_instrumentation__.Notify(85928)
	return b.response
}

func (b *Batch) MustPErr() *roachpb.Error {
	__antithesis_instrumentation__.Notify(85929)
	if b.pErr == nil {
		__antithesis_instrumentation__.Notify(85931)
		panic(errors.Errorf("expected non-nil pErr for batch %+v", b))
	} else {
		__antithesis_instrumentation__.Notify(85932)
	}
	__antithesis_instrumentation__.Notify(85930)
	return b.pErr
}

func (b *Batch) validate() error {
	__antithesis_instrumentation__.Notify(85933)
	err := b.resultErr()
	if err != nil {
		__antithesis_instrumentation__.Notify(85935)

		b.pErr = roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(85936)
	}
	__antithesis_instrumentation__.Notify(85934)
	return err
}

func (b *Batch) initResult(calls, numRows int, raw bool, err error) {
	__antithesis_instrumentation__.Notify(85937)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(85941)
		return b.raw == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(85942)
		return !raw == true
	}() == true {
		__antithesis_instrumentation__.Notify(85943)
		err = errors.Errorf("must not use non-raw operations on a raw batch")
	} else {
		__antithesis_instrumentation__.Notify(85944)
	}
	__antithesis_instrumentation__.Notify(85938)

	r := Result{calls: calls, Err: err}
	if numRows > 0 && func() bool {
		__antithesis_instrumentation__.Notify(85945)
		return !b.raw == true
	}() == true {
		__antithesis_instrumentation__.Notify(85946)
		if b.rowsStaticIdx+numRows <= len(b.rowsStaticBuf) {
			__antithesis_instrumentation__.Notify(85947)
			r.Rows = b.rowsStaticBuf[b.rowsStaticIdx : b.rowsStaticIdx+numRows]
			b.rowsStaticIdx += numRows
		} else {
			__antithesis_instrumentation__.Notify(85948)

			switch numRows {
			case 1:
				__antithesis_instrumentation__.Notify(85949)

				if cap(b.rowsBuf)-len(b.rowsBuf) == 0 {
					__antithesis_instrumentation__.Notify(85952)
					const minSize = 16
					const maxSize = 128
					size := cap(b.rowsBuf) * 2
					if size < minSize {
						__antithesis_instrumentation__.Notify(85954)
						size = minSize
					} else {
						__antithesis_instrumentation__.Notify(85955)
						if size > maxSize {
							__antithesis_instrumentation__.Notify(85956)
							size = maxSize
						} else {
							__antithesis_instrumentation__.Notify(85957)
						}
					}
					__antithesis_instrumentation__.Notify(85953)
					b.rowsBuf = make([]KeyValue, 0, size)
				} else {
					__antithesis_instrumentation__.Notify(85958)
				}
				__antithesis_instrumentation__.Notify(85950)
				pos := len(b.rowsBuf)
				r.Rows = b.rowsBuf[pos : pos+1 : pos+1]
				b.rowsBuf = b.rowsBuf[:pos+1]
			default:
				__antithesis_instrumentation__.Notify(85951)
				r.Rows = make([]KeyValue, numRows)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(85959)
	}
	__antithesis_instrumentation__.Notify(85939)
	if b.Results == nil {
		__antithesis_instrumentation__.Notify(85960)
		b.Results = b.resultsBuf[:0]
	} else {
		__antithesis_instrumentation__.Notify(85961)
	}
	__antithesis_instrumentation__.Notify(85940)
	b.Results = append(b.Results, r)
}

func (b *Batch) fillResults(ctx context.Context) {
	__antithesis_instrumentation__.Notify(85962)

	if b.raw {
		__antithesis_instrumentation__.Notify(85964)
		return
	} else {
		__antithesis_instrumentation__.Notify(85965)
	}
	__antithesis_instrumentation__.Notify(85963)

	offset := 0
	for i := range b.Results {
		__antithesis_instrumentation__.Notify(85966)
		result := &b.Results[i]

		for k := 0; k < result.calls; k++ {
			__antithesis_instrumentation__.Notify(85968)
			args := b.reqs[offset+k].GetInner()

			var reply roachpb.Response

			if result.Err == nil {
				__antithesis_instrumentation__.Notify(85971)

				result.Err = b.pErr.GoError()
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(85972)

					if b.response != nil && func() bool {
						__antithesis_instrumentation__.Notify(85973)
						return offset+k < len(b.response.Responses) == true
					}() == true {
						__antithesis_instrumentation__.Notify(85974)
						reply = b.response.Responses[offset+k].GetInner()
					} else {
						__antithesis_instrumentation__.Notify(85975)
						if args.Method() != roachpb.EndTxn {
							__antithesis_instrumentation__.Notify(85976)

							panic(errors.Errorf("not enough responses for calls: (%T) %+v\nresponses: %+v",
								args, args, b.response))
						} else {
							__antithesis_instrumentation__.Notify(85977)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(85978)
				}
			} else {
				__antithesis_instrumentation__.Notify(85979)
			}
			__antithesis_instrumentation__.Notify(85969)

			switch req := args.(type) {
			case *roachpb.GetRequest:
				__antithesis_instrumentation__.Notify(85980)
				row := &result.Rows[k]
				row.Key = []byte(req.Key)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86015)
					row.Value = reply.(*roachpb.GetResponse).Value
				} else {
					__antithesis_instrumentation__.Notify(86016)
				}
			case *roachpb.PutRequest:
				__antithesis_instrumentation__.Notify(85981)
				row := &result.Rows[k]
				row.Key = []byte(req.Key)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86017)
					row.Value = &req.Value
				} else {
					__antithesis_instrumentation__.Notify(86018)
				}
			case *roachpb.ConditionalPutRequest:
				__antithesis_instrumentation__.Notify(85982)
				row := &result.Rows[k]
				row.Key = []byte(req.Key)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86019)
					row.Value = &req.Value
				} else {
					__antithesis_instrumentation__.Notify(86020)
				}
			case *roachpb.InitPutRequest:
				__antithesis_instrumentation__.Notify(85983)
				row := &result.Rows[k]
				row.Key = []byte(req.Key)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86021)
					row.Value = &req.Value
				} else {
					__antithesis_instrumentation__.Notify(86022)
				}
			case *roachpb.IncrementRequest:
				__antithesis_instrumentation__.Notify(85984)
				row := &result.Rows[k]
				row.Key = []byte(req.Key)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86023)
					t := reply.(*roachpb.IncrementResponse)
					row.Value = &roachpb.Value{}
					row.Value.SetInt(t.NewValue)
				} else {
					__antithesis_instrumentation__.Notify(86024)
				}
			case *roachpb.ScanRequest:
				__antithesis_instrumentation__.Notify(85985)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86025)
					t := reply.(*roachpb.ScanResponse)
					result.Rows = make([]KeyValue, len(t.Rows))
					for j := range t.Rows {
						__antithesis_instrumentation__.Notify(86026)
						src := &t.Rows[j]
						dst := &result.Rows[j]
						dst.Key = src.Key
						dst.Value = &src.Value
					}
				} else {
					__antithesis_instrumentation__.Notify(86027)
				}
			case *roachpb.ReverseScanRequest:
				__antithesis_instrumentation__.Notify(85986)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86028)
					t := reply.(*roachpb.ReverseScanResponse)
					result.Rows = make([]KeyValue, len(t.Rows))
					for j := range t.Rows {
						__antithesis_instrumentation__.Notify(86029)
						src := &t.Rows[j]
						dst := &result.Rows[j]
						dst.Key = src.Key
						dst.Value = &src.Value
					}
				} else {
					__antithesis_instrumentation__.Notify(86030)
				}
			case *roachpb.DeleteRequest:
				__antithesis_instrumentation__.Notify(85987)
				row := &result.Rows[k]
				row.Key = []byte(args.(*roachpb.DeleteRequest).Key)
			case *roachpb.DeleteRangeRequest:
				__antithesis_instrumentation__.Notify(85988)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86031)
					result.Keys = reply.(*roachpb.DeleteRangeResponse).Keys
				} else {
					__antithesis_instrumentation__.Notify(86032)
				}

			case *roachpb.EndTxnRequest:
				__antithesis_instrumentation__.Notify(85989)
			case *roachpb.AdminMergeRequest:
				__antithesis_instrumentation__.Notify(85990)
			case *roachpb.AdminSplitRequest:
				__antithesis_instrumentation__.Notify(85991)
			case *roachpb.AdminUnsplitRequest:
				__antithesis_instrumentation__.Notify(85992)
			case *roachpb.AdminTransferLeaseRequest:
				__antithesis_instrumentation__.Notify(85993)
			case *roachpb.AdminChangeReplicasRequest:
				__antithesis_instrumentation__.Notify(85994)
			case *roachpb.AdminRelocateRangeRequest:
				__antithesis_instrumentation__.Notify(85995)
			case *roachpb.HeartbeatTxnRequest:
				__antithesis_instrumentation__.Notify(85996)
			case *roachpb.GCRequest:
				__antithesis_instrumentation__.Notify(85997)
			case *roachpb.LeaseInfoRequest:
				__antithesis_instrumentation__.Notify(85998)
			case *roachpb.PushTxnRequest:
				__antithesis_instrumentation__.Notify(85999)
			case *roachpb.QueryTxnRequest:
				__antithesis_instrumentation__.Notify(86000)
			case *roachpb.QueryIntentRequest:
				__antithesis_instrumentation__.Notify(86001)
			case *roachpb.ResolveIntentRequest:
				__antithesis_instrumentation__.Notify(86002)
			case *roachpb.ResolveIntentRangeRequest:
				__antithesis_instrumentation__.Notify(86003)
			case *roachpb.MergeRequest:
				__antithesis_instrumentation__.Notify(86004)
			case *roachpb.TruncateLogRequest:
				__antithesis_instrumentation__.Notify(86005)
			case *roachpb.RequestLeaseRequest:
				__antithesis_instrumentation__.Notify(86006)
			case *roachpb.CheckConsistencyRequest:
				__antithesis_instrumentation__.Notify(86007)
			case *roachpb.AdminScatterRequest:
				__antithesis_instrumentation__.Notify(86008)
			case *roachpb.AddSSTableRequest:
				__antithesis_instrumentation__.Notify(86009)
			case *roachpb.MigrateRequest:
				__antithesis_instrumentation__.Notify(86010)
			case *roachpb.QueryResolvedTimestampRequest:
				__antithesis_instrumentation__.Notify(86011)
			case *roachpb.BarrierRequest:
				__antithesis_instrumentation__.Notify(86012)
			case *roachpb.ScanInterleavedIntentsRequest:
				__antithesis_instrumentation__.Notify(86013)
			default:
				__antithesis_instrumentation__.Notify(86014)
				if result.Err == nil {
					__antithesis_instrumentation__.Notify(86033)
					result.Err = errors.Errorf("unsupported reply: %T for %T",
						reply, args)
				} else {
					__antithesis_instrumentation__.Notify(86034)
				}
			}
			__antithesis_instrumentation__.Notify(85970)

			if result.Err == nil && func() bool {
				__antithesis_instrumentation__.Notify(86035)
				return reply != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(86036)
				if h := reply.Header(); h.ResumeSpan != nil {
					__antithesis_instrumentation__.Notify(86037)
					result.ResumeSpan = h.ResumeSpan
					result.ResumeReason = h.ResumeReason
					result.ResumeNextBytes = h.ResumeNextBytes
				} else {
					__antithesis_instrumentation__.Notify(86038)
				}
			} else {
				__antithesis_instrumentation__.Notify(86039)
			}
		}
		__antithesis_instrumentation__.Notify(85967)
		offset += result.calls
	}
}

func (b *Batch) resultErr() error {
	__antithesis_instrumentation__.Notify(86040)
	for i := range b.Results {
		__antithesis_instrumentation__.Notify(86042)
		if err := b.Results[i].Err; err != nil {
			__antithesis_instrumentation__.Notify(86043)
			return err
		} else {
			__antithesis_instrumentation__.Notify(86044)
		}
	}
	__antithesis_instrumentation__.Notify(86041)
	return nil
}

func (b *Batch) growReqs(n int) {
	__antithesis_instrumentation__.Notify(86045)
	if len(b.reqs)+n > cap(b.reqs) {
		__antithesis_instrumentation__.Notify(86047)
		newSize := 2 * cap(b.reqs)
		if newSize == 0 {
			__antithesis_instrumentation__.Notify(86050)
			newSize = 8
		} else {
			__antithesis_instrumentation__.Notify(86051)
		}
		__antithesis_instrumentation__.Notify(86048)
		for newSize < len(b.reqs)+n {
			__antithesis_instrumentation__.Notify(86052)
			newSize *= 2
		}
		__antithesis_instrumentation__.Notify(86049)
		newReqs := make([]roachpb.RequestUnion, len(b.reqs), newSize)
		copy(newReqs, b.reqs)
		b.reqs = newReqs
	} else {
		__antithesis_instrumentation__.Notify(86053)
	}
	__antithesis_instrumentation__.Notify(86046)
	b.reqs = b.reqs[:len(b.reqs)+n]
}

func (b *Batch) appendReqs(args ...roachpb.Request) {
	__antithesis_instrumentation__.Notify(86054)
	n := len(b.reqs)
	b.growReqs(len(args))
	for i := range args {
		__antithesis_instrumentation__.Notify(86055)
		b.reqs[n+i].MustSetInner(args[i])
	}
}

func (b *Batch) AddRawRequest(reqs ...roachpb.Request) {
	__antithesis_instrumentation__.Notify(86056)
	b.raw = true
	for _, args := range reqs {
		__antithesis_instrumentation__.Notify(86057)
		numRows := 0
		switch args.(type) {
		case *roachpb.GetRequest,
			*roachpb.PutRequest,
			*roachpb.ConditionalPutRequest,
			*roachpb.IncrementRequest,
			*roachpb.DeleteRequest:
			__antithesis_instrumentation__.Notify(86059)
			numRows = 1
		}
		__antithesis_instrumentation__.Notify(86058)
		b.appendReqs(args)
		b.initResult(1, numRows, raw, nil)
	}
}

func (b *Batch) get(key interface{}, forUpdate bool) {
	__antithesis_instrumentation__.Notify(86060)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86062)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86063)
	}
	__antithesis_instrumentation__.Notify(86061)
	b.appendReqs(roachpb.NewGet(k, forUpdate))
	b.initResult(1, 1, notRaw, nil)
}

func (b *Batch) Get(key interface{}) {
	__antithesis_instrumentation__.Notify(86064)
	b.get(key, false)
}

func (b *Batch) GetForUpdate(key interface{}) {
	__antithesis_instrumentation__.Notify(86065)
	b.get(key, true)
}

func (b *Batch) put(key, value interface{}, inline bool) {
	__antithesis_instrumentation__.Notify(86066)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86070)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86071)
	}
	__antithesis_instrumentation__.Notify(86067)
	v, err := marshalValue(value)
	if err != nil {
		__antithesis_instrumentation__.Notify(86072)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86073)
	}
	__antithesis_instrumentation__.Notify(86068)
	if inline {
		__antithesis_instrumentation__.Notify(86074)
		b.appendReqs(roachpb.NewPutInline(k, v))
	} else {
		__antithesis_instrumentation__.Notify(86075)
		b.appendReqs(roachpb.NewPut(k, v))
	}
	__antithesis_instrumentation__.Notify(86069)
	b.approxMutationReqBytes += len(k) + len(v.RawBytes)
	b.initResult(1, 1, notRaw, nil)
}

func (b *Batch) Put(key, value interface{}) {
	__antithesis_instrumentation__.Notify(86076)
	if value == nil {
		__antithesis_instrumentation__.Notify(86078)

		panic("can't Put an empty Value; did you mean to Del() instead?")
	} else {
		__antithesis_instrumentation__.Notify(86079)
	}
	__antithesis_instrumentation__.Notify(86077)
	b.put(key, value, false)
}

func (b *Batch) PutInline(key, value interface{}) {
	__antithesis_instrumentation__.Notify(86080)
	b.put(key, value, true)
}

func (b *Batch) CPut(key, value interface{}, expValue []byte) {
	__antithesis_instrumentation__.Notify(86081)
	b.cputInternal(key, value, expValue, false, false)
}

func (b *Batch) CPutAllowingIfNotExists(key, value interface{}, expValue []byte) {
	__antithesis_instrumentation__.Notify(86082)
	b.cputInternal(key, value, expValue, true, false)
}

func (b *Batch) CPutInline(key, value interface{}, expValue []byte) {
	__antithesis_instrumentation__.Notify(86083)
	b.cputInternal(key, value, expValue, false, true)
}

func (b *Batch) cputInternal(
	key, value interface{}, expValue []byte, allowNotExist bool, inline bool,
) {
	__antithesis_instrumentation__.Notify(86084)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86088)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86089)
	}
	__antithesis_instrumentation__.Notify(86085)
	v, err := marshalValue(value)
	if err != nil {
		__antithesis_instrumentation__.Notify(86090)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86091)
	}
	__antithesis_instrumentation__.Notify(86086)
	if inline {
		__antithesis_instrumentation__.Notify(86092)
		b.appendReqs(roachpb.NewConditionalPutInline(k, v, expValue, allowNotExist))
	} else {
		__antithesis_instrumentation__.Notify(86093)
		b.appendReqs(roachpb.NewConditionalPut(k, v, expValue, allowNotExist))
	}
	__antithesis_instrumentation__.Notify(86087)
	b.approxMutationReqBytes += len(k) + len(v.RawBytes)
	b.initResult(1, 1, notRaw, nil)
}

func (b *Batch) InitPut(key, value interface{}, failOnTombstones bool) {
	__antithesis_instrumentation__.Notify(86094)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86097)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86098)
	}
	__antithesis_instrumentation__.Notify(86095)
	v, err := marshalValue(value)
	if err != nil {
		__antithesis_instrumentation__.Notify(86099)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86100)
	}
	__antithesis_instrumentation__.Notify(86096)
	b.appendReqs(roachpb.NewInitPut(k, v, failOnTombstones))
	b.approxMutationReqBytes += len(k) + len(v.RawBytes)
	b.initResult(1, 1, notRaw, nil)
}

func (b *Batch) Inc(key interface{}, value int64) {
	__antithesis_instrumentation__.Notify(86101)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86103)
		b.initResult(0, 1, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86104)
	}
	__antithesis_instrumentation__.Notify(86102)
	b.appendReqs(roachpb.NewIncrement(k, value))
	b.initResult(1, 1, notRaw, nil)
}

func (b *Batch) scan(s, e interface{}, isReverse, forUpdate bool) {
	__antithesis_instrumentation__.Notify(86105)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86109)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86110)
	}
	__antithesis_instrumentation__.Notify(86106)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86111)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86112)
	}
	__antithesis_instrumentation__.Notify(86107)
	if !isReverse {
		__antithesis_instrumentation__.Notify(86113)
		b.appendReqs(roachpb.NewScan(begin, end, forUpdate))
	} else {
		__antithesis_instrumentation__.Notify(86114)
		b.appendReqs(roachpb.NewReverseScan(begin, end, forUpdate))
	}
	__antithesis_instrumentation__.Notify(86108)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) Scan(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86115)
	b.scan(s, e, false, false)
}

func (b *Batch) ScanForUpdate(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86116)
	b.scan(s, e, false, true)
}

func (b *Batch) ReverseScan(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86117)
	b.scan(s, e, true, false)
}

func (b *Batch) ReverseScanForUpdate(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86118)
	b.scan(s, e, true, true)
}

func (b *Batch) Del(keys ...interface{}) {
	__antithesis_instrumentation__.Notify(86119)
	reqs := make([]roachpb.Request, 0, len(keys))
	for _, key := range keys {
		__antithesis_instrumentation__.Notify(86121)
		k, err := marshalKey(key)
		if err != nil {
			__antithesis_instrumentation__.Notify(86123)
			b.initResult(0, len(keys), notRaw, err)
			return
		} else {
			__antithesis_instrumentation__.Notify(86124)
		}
		__antithesis_instrumentation__.Notify(86122)
		reqs = append(reqs, roachpb.NewDelete(k))
		b.approxMutationReqBytes += len(k)
	}
	__antithesis_instrumentation__.Notify(86120)
	b.appendReqs(reqs...)
	b.initResult(len(reqs), len(reqs), notRaw, nil)
}

func (b *Batch) DelRange(s, e interface{}, returnKeys bool) {
	__antithesis_instrumentation__.Notify(86125)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86128)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86129)
	}
	__antithesis_instrumentation__.Notify(86126)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86130)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86131)
	}
	__antithesis_instrumentation__.Notify(86127)
	b.appendReqs(roachpb.NewDeleteRange(begin, end, returnKeys))
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) adminMerge(key interface{}) {
	__antithesis_instrumentation__.Notify(86132)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86134)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86135)
	}
	__antithesis_instrumentation__.Notify(86133)
	req := &roachpb.AdminMergeRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: k,
		},
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) adminSplit(
	splitKeyIn interface{}, expirationTime hlc.Timestamp, predicateKeys []roachpb.Key,
) {
	__antithesis_instrumentation__.Notify(86136)
	splitKey, err := marshalKey(splitKeyIn)
	if err != nil {
		__antithesis_instrumentation__.Notify(86138)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86139)
	}
	__antithesis_instrumentation__.Notify(86137)
	req := &roachpb.AdminSplitRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: splitKey,
		},
		SplitKey:       splitKey,
		ExpirationTime: expirationTime,
		PredicateKeys:  predicateKeys,
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) adminUnsplit(splitKeyIn interface{}) {
	__antithesis_instrumentation__.Notify(86140)
	splitKey, err := marshalKey(splitKeyIn)
	if err != nil {
		__antithesis_instrumentation__.Notify(86142)
		b.initResult(0, 0, notRaw, err)
	} else {
		__antithesis_instrumentation__.Notify(86143)
	}
	__antithesis_instrumentation__.Notify(86141)
	req := &roachpb.AdminUnsplitRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: splitKey,
		},
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) adminTransferLease(key interface{}, target roachpb.StoreID) {
	__antithesis_instrumentation__.Notify(86144)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86146)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86147)
	}
	__antithesis_instrumentation__.Notify(86145)
	req := &roachpb.AdminTransferLeaseRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: k,
		},
		Target: target,
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) adminChangeReplicas(
	key interface{}, expDesc roachpb.RangeDescriptor, chgs []roachpb.ReplicationChange,
) {
	__antithesis_instrumentation__.Notify(86148)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86150)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86151)
	}
	__antithesis_instrumentation__.Notify(86149)
	req := &roachpb.AdminChangeReplicasRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: k,
		},
		ExpDesc: expDesc,
	}
	req.AddChanges(chgs...)

	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) adminRelocateRange(
	key interface{},
	voterTargets, nonVoterTargets []roachpb.ReplicationTarget,
	transferLeaseToFirstVoter bool,
) {
	__antithesis_instrumentation__.Notify(86152)
	k, err := marshalKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(86154)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86155)
	}
	__antithesis_instrumentation__.Notify(86153)
	req := &roachpb.AdminRelocateRangeRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: k,
		},
		VoterTargets:                      voterTargets,
		NonVoterTargets:                   nonVoterTargets,
		TransferLeaseToFirstVoter:         transferLeaseToFirstVoter,
		TransferLeaseToFirstVoterAccurate: true,
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) addSSTable(
	s, e interface{},
	data []byte,
	disallowConflicts bool,
	disallowShadowing bool,
	disallowShadowingBelow hlc.Timestamp,
	stats *enginepb.MVCCStats,
	ingestAsWrites bool,
	sstTimestampToRequestTimestamp hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(86156)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86159)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86160)
	}
	__antithesis_instrumentation__.Notify(86157)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86161)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86162)
	}
	__antithesis_instrumentation__.Notify(86158)
	req := &roachpb.AddSSTableRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    begin,
			EndKey: end,
		},
		Data:                           data,
		DisallowConflicts:              disallowConflicts,
		DisallowShadowing:              disallowShadowing,
		DisallowShadowingBelow:         disallowShadowingBelow,
		MVCCStats:                      stats,
		IngestAsWrites:                 ingestAsWrites,
		SSTTimestampToRequestTimestamp: sstTimestampToRequestTimestamp,
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) migrate(s, e interface{}, version roachpb.Version) {
	__antithesis_instrumentation__.Notify(86163)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86166)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86167)
	}
	__antithesis_instrumentation__.Notify(86164)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86168)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86169)
	}
	__antithesis_instrumentation__.Notify(86165)
	req := &roachpb.MigrateRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    begin,
			EndKey: end,
		},
		Version: version,
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) queryResolvedTimestamp(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86170)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86173)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86174)
	}
	__antithesis_instrumentation__.Notify(86171)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86175)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86176)
	}
	__antithesis_instrumentation__.Notify(86172)
	req := &roachpb.QueryResolvedTimestampRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    begin,
			EndKey: end,
		},
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) scanInterleavedIntents(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86177)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86180)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86181)
	}
	__antithesis_instrumentation__.Notify(86178)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86182)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86183)
	}
	__antithesis_instrumentation__.Notify(86179)
	req := &roachpb.ScanInterleavedIntentsRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    begin,
			EndKey: end,
		},
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}

func (b *Batch) barrier(s, e interface{}) {
	__antithesis_instrumentation__.Notify(86184)
	begin, err := marshalKey(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(86187)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86188)
	}
	__antithesis_instrumentation__.Notify(86185)
	end, err := marshalKey(e)
	if err != nil {
		__antithesis_instrumentation__.Notify(86189)
		b.initResult(0, 0, notRaw, err)
		return
	} else {
		__antithesis_instrumentation__.Notify(86190)
	}
	__antithesis_instrumentation__.Notify(86186)
	req := &roachpb.BarrierRequest{
		RequestHeader: roachpb.RequestHeader{
			Key:    begin,
			EndKey: end,
		},
	}
	b.appendReqs(req)
	b.initResult(1, 0, notRaw, nil)
}
