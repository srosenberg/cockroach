package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

func getKVBatchSize(forceProductionKVBatchSize bool) rowinfra.KeyLimit {
	__antithesis_instrumentation__.Notify(568166)
	if forceProductionKVBatchSize {
		__antithesis_instrumentation__.Notify(568168)
		return rowinfra.ProductionKVBatchSize
	} else {
		__antithesis_instrumentation__.Notify(568169)
	}
	__antithesis_instrumentation__.Notify(568167)
	return defaultKVBatchSize
}

var defaultKVBatchSize = rowinfra.KeyLimit(util.ConstantWithMetamorphicTestValue(
	"kv-batch-size",
	int(rowinfra.ProductionKVBatchSize),
	1,
))

type sendFunc func(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, error)

type txnKVFetcher struct {
	sendFn sendFunc

	spans roachpb.Spans

	spansScratch roachpb.Spans

	newFetchSpansIdx int

	firstBatchKeyLimit rowinfra.KeyLimit

	batchBytesLimit rowinfra.BytesLimit

	reverse bool

	lockStrength lock.Strength

	lockWaitPolicy lock.WaitPolicy

	lockTimeout time.Duration

	alreadyFetched bool
	batchIdx       int

	responses        []roachpb.ResponseUnion
	remainingBatches [][]byte

	acc *mon.BoundAccount

	spansAccountedFor         int64
	batchResponseAccountedFor int64

	forceProductionKVBatchSize bool

	requestAdmissionHeader roachpb.AdmissionHeader
	responseAdmissionQ     *admission.WorkQueue
}

var _ KVBatchFetcher = &txnKVFetcher{}

func (f *txnKVFetcher) getBatchKeyLimit() rowinfra.KeyLimit {
	__antithesis_instrumentation__.Notify(568170)
	return f.getBatchKeyLimitForIdx(f.batchIdx)
}

func (f *txnKVFetcher) getBatchKeyLimitForIdx(batchIdx int) rowinfra.KeyLimit {
	__antithesis_instrumentation__.Notify(568171)
	if f.firstBatchKeyLimit == 0 {
		__antithesis_instrumentation__.Notify(568174)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(568175)
	}
	__antithesis_instrumentation__.Notify(568172)

	kvBatchSize := getKVBatchSize(f.forceProductionKVBatchSize)
	if f.firstBatchKeyLimit >= kvBatchSize {
		__antithesis_instrumentation__.Notify(568176)
		return kvBatchSize
	} else {
		__antithesis_instrumentation__.Notify(568177)
	}
	__antithesis_instrumentation__.Notify(568173)

	switch batchIdx {
	case 0:
		__antithesis_instrumentation__.Notify(568178)
		return f.firstBatchKeyLimit

	case 1:
		__antithesis_instrumentation__.Notify(568179)

		secondBatch := f.firstBatchKeyLimit * 10
		switch {
		case secondBatch < kvBatchSize/10:
			__antithesis_instrumentation__.Notify(568181)
			return kvBatchSize / 10
		case secondBatch > kvBatchSize:
			__antithesis_instrumentation__.Notify(568182)
			return kvBatchSize
		default:
			__antithesis_instrumentation__.Notify(568183)
			return secondBatch
		}

	default:
		__antithesis_instrumentation__.Notify(568180)
		return kvBatchSize
	}
}

func makeKVBatchFetcherDefaultSendFunc(txn *kv.Txn) sendFunc {
	__antithesis_instrumentation__.Notify(568184)
	return func(
		ctx context.Context,
		ba roachpb.BatchRequest,
	) (*roachpb.BatchResponse, error) {
		__antithesis_instrumentation__.Notify(568185)
		res, err := txn.Send(ctx, ba)
		if err != nil {
			__antithesis_instrumentation__.Notify(568187)
			return nil, err.GoError()
		} else {
			__antithesis_instrumentation__.Notify(568188)
		}
		__antithesis_instrumentation__.Notify(568186)
		return res, nil
	}
}

func makeKVBatchFetcher(
	ctx context.Context,
	sendFn sendFunc,
	spans roachpb.Spans,
	reverse bool,
	batchBytesLimit rowinfra.BytesLimit,
	firstBatchKeyLimit rowinfra.KeyLimit,
	lockStrength descpb.ScanLockingStrength,
	lockWaitPolicy descpb.ScanLockingWaitPolicy,
	lockTimeout time.Duration,
	acc *mon.BoundAccount,
	forceProductionKVBatchSize bool,
	requestAdmissionHeader roachpb.AdmissionHeader,
	responseAdmissionQ *admission.WorkQueue,
) (txnKVFetcher, error) {
	__antithesis_instrumentation__.Notify(568189)
	if firstBatchKeyLimit < 0 || func() bool {
		__antithesis_instrumentation__.Notify(568194)
		return (batchBytesLimit == 0 && func() bool {
			__antithesis_instrumentation__.Notify(568195)
			return firstBatchKeyLimit != 0 == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(568196)

		return txnKVFetcher{}, errors.Errorf("invalid batch limit %d (batchBytesLimit: %d)",
			firstBatchKeyLimit, batchBytesLimit)
	} else {
		__antithesis_instrumentation__.Notify(568197)
	}
	__antithesis_instrumentation__.Notify(568190)

	if batchBytesLimit != 0 {
		__antithesis_instrumentation__.Notify(568198)

		for i := 1; i < len(spans); i++ {
			__antithesis_instrumentation__.Notify(568199)
			prevKey := spans[i-1].EndKey
			if prevKey == nil {
				__antithesis_instrumentation__.Notify(568201)

				prevKey = spans[i-1].Key
			} else {
				__antithesis_instrumentation__.Notify(568202)
			}
			__antithesis_instrumentation__.Notify(568200)
			if spans[i].Key.Compare(prevKey) < 0 {
				__antithesis_instrumentation__.Notify(568203)
				return txnKVFetcher{}, errors.Errorf("unordered spans (%s %s)", spans[i-1], spans[i])
			} else {
				__antithesis_instrumentation__.Notify(568204)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(568205)
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(568206)

			for i := 1; i < len(spans); i++ {
				__antithesis_instrumentation__.Notify(568207)
				prevEndKey := spans[i-1].EndKey
				if prevEndKey == nil {
					__antithesis_instrumentation__.Notify(568211)
					prevEndKey = spans[i-1].Key
				} else {
					__antithesis_instrumentation__.Notify(568212)
				}
				__antithesis_instrumentation__.Notify(568208)
				curEndKey := spans[i].EndKey
				if curEndKey == nil {
					__antithesis_instrumentation__.Notify(568213)
					curEndKey = spans[i].Key
				} else {
					__antithesis_instrumentation__.Notify(568214)
				}
				__antithesis_instrumentation__.Notify(568209)
				if spans[i].Key.Compare(prevEndKey) >= 0 {
					__antithesis_instrumentation__.Notify(568215)

					continue
				} else {
					__antithesis_instrumentation__.Notify(568216)
					if curEndKey.Compare(spans[i-1].Key) <= 0 {
						__antithesis_instrumentation__.Notify(568217)

						continue
					} else {
						__antithesis_instrumentation__.Notify(568218)
					}
				}
				__antithesis_instrumentation__.Notify(568210)

				return txnKVFetcher{}, errors.Errorf("overlapping neighbor spans (%s %s)", spans[i-1], spans[i])
			}
		} else {
			__antithesis_instrumentation__.Notify(568219)
		}
	}
	__antithesis_instrumentation__.Notify(568191)

	f := txnKVFetcher{
		sendFn:                     sendFn,
		reverse:                    reverse,
		batchBytesLimit:            batchBytesLimit,
		firstBatchKeyLimit:         firstBatchKeyLimit,
		lockStrength:               getKeyLockingStrength(lockStrength),
		lockWaitPolicy:             GetWaitPolicy(lockWaitPolicy),
		lockTimeout:                lockTimeout,
		acc:                        acc,
		forceProductionKVBatchSize: forceProductionKVBatchSize,
		requestAdmissionHeader:     requestAdmissionHeader,
		responseAdmissionQ:         responseAdmissionQ,
	}

	if f.acc != nil {
		__antithesis_instrumentation__.Notify(568220)
		f.spansAccountedFor = spans.MemUsage()
		if err := f.acc.Grow(ctx, f.spansAccountedFor); err != nil {
			__antithesis_instrumentation__.Notify(568221)
			return txnKVFetcher{}, err
		} else {
			__antithesis_instrumentation__.Notify(568222)
		}
	} else {
		__antithesis_instrumentation__.Notify(568223)
	}
	__antithesis_instrumentation__.Notify(568192)

	f.spans = spans
	if reverse {
		__antithesis_instrumentation__.Notify(568224)

		i, j := 0, len(spans)-1
		for i < j {
			__antithesis_instrumentation__.Notify(568225)
			f.spans[i], f.spans[j] = f.spans[j], f.spans[i]
			i++
			j--
		}
	} else {
		__antithesis_instrumentation__.Notify(568226)
	}
	__antithesis_instrumentation__.Notify(568193)

	f.spansScratch = f.spans

	return f, nil
}

func (f *txnKVFetcher) fetch(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(568227)
	var ba roachpb.BatchRequest
	ba.Header.WaitPolicy = f.lockWaitPolicy
	ba.Header.LockTimeout = f.lockTimeout
	ba.Header.TargetBytes = int64(f.batchBytesLimit)
	ba.Header.MaxSpanRequestKeys = int64(f.getBatchKeyLimit())
	ba.AdmissionHeader = f.requestAdmissionHeader
	ba.Requests = spansToRequests(f.spans, f.reverse, f.lockStrength)

	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(568235)
		log.VEventf(ctx, 2, "Scan %s", f.spans)
	} else {
		__antithesis_instrumentation__.Notify(568236)
	}
	__antithesis_instrumentation__.Notify(568228)

	monitoring := f.acc != nil

	const tokenFetchAllocation = 1 << 10
	if !monitoring || func() bool {
		__antithesis_instrumentation__.Notify(568237)
		return f.batchResponseAccountedFor < tokenFetchAllocation == true
	}() == true {
		__antithesis_instrumentation__.Notify(568238)

		ba.AdmissionHeader.NoMemoryReservedAtSource = true
	} else {
		__antithesis_instrumentation__.Notify(568239)
	}
	__antithesis_instrumentation__.Notify(568229)
	if monitoring && func() bool {
		__antithesis_instrumentation__.Notify(568240)
		return f.batchResponseAccountedFor < tokenFetchAllocation == true
	}() == true {
		__antithesis_instrumentation__.Notify(568241)

		f.batchResponseAccountedFor = tokenFetchAllocation
		if err := f.acc.Grow(ctx, f.batchResponseAccountedFor); err != nil {
			__antithesis_instrumentation__.Notify(568242)
			return err
		} else {
			__antithesis_instrumentation__.Notify(568243)
		}
	} else {
		__antithesis_instrumentation__.Notify(568244)
	}
	__antithesis_instrumentation__.Notify(568230)

	br, err := f.sendFn(ctx, ba)
	if err != nil {
		__antithesis_instrumentation__.Notify(568245)
		return err
	} else {
		__antithesis_instrumentation__.Notify(568246)
	}
	__antithesis_instrumentation__.Notify(568231)
	if br != nil {
		__antithesis_instrumentation__.Notify(568247)
		f.responses = br.Responses
	} else {
		__antithesis_instrumentation__.Notify(568248)
		f.responses = nil
	}
	__antithesis_instrumentation__.Notify(568232)
	returnedBytes := int64(br.Size())
	if monitoring && func() bool {
		__antithesis_instrumentation__.Notify(568249)
		return (returnedBytes > int64(f.batchBytesLimit) || func() bool {
			__antithesis_instrumentation__.Notify(568250)
			return returnedBytes > f.batchResponseAccountedFor == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(568251)

		used := f.acc.Used()
		delta := returnedBytes - f.batchResponseAccountedFor
		if err := f.acc.Resize(ctx, used, used+delta); err != nil {
			__antithesis_instrumentation__.Notify(568253)
			return err
		} else {
			__antithesis_instrumentation__.Notify(568254)
		}
		__antithesis_instrumentation__.Notify(568252)
		f.batchResponseAccountedFor = returnedBytes
	} else {
		__antithesis_instrumentation__.Notify(568255)
	}
	__antithesis_instrumentation__.Notify(568233)

	if br != nil && func() bool {
		__antithesis_instrumentation__.Notify(568256)
		return f.responseAdmissionQ != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(568257)
		responseAdmission := admission.WorkInfo{
			TenantID:   roachpb.SystemTenantID,
			Priority:   admission.WorkPriority(f.requestAdmissionHeader.Priority),
			CreateTime: f.requestAdmissionHeader.CreateTime,
		}
		if _, err := f.responseAdmissionQ.Admit(ctx, responseAdmission); err != nil {
			__antithesis_instrumentation__.Notify(568258)
			return err
		} else {
			__antithesis_instrumentation__.Notify(568259)
		}
	} else {
		__antithesis_instrumentation__.Notify(568260)
	}
	__antithesis_instrumentation__.Notify(568234)

	f.batchIdx++
	f.newFetchSpansIdx = 0
	f.alreadyFetched = true

	return nil
}

func popBatch(batches [][]byte) (batch []byte, remainingBatches [][]byte) {
	__antithesis_instrumentation__.Notify(568261)
	batch, remainingBatches = batches[0], batches[1:]
	batches[0] = nil
	return batch, remainingBatches
}

func (f *txnKVFetcher) nextBatch(
	ctx context.Context,
) (ok bool, kvs []roachpb.KeyValue, batchResp []byte, err error) {
	__antithesis_instrumentation__.Notify(568262)

	if len(f.remainingBatches) > 0 {
		__antithesis_instrumentation__.Notify(568267)

		batchResp, f.remainingBatches = popBatch(f.remainingBatches)
		return true, nil, batchResp, nil
	} else {
		__antithesis_instrumentation__.Notify(568268)
	}
	__antithesis_instrumentation__.Notify(568263)

	for len(f.responses) > 0 {
		__antithesis_instrumentation__.Notify(568269)
		reply := f.responses[0].GetInner()
		f.responses[0] = roachpb.ResponseUnion{}
		f.responses = f.responses[1:]

		origSpan := f.spans[0]
		f.spans[0] = roachpb.Span{}
		f.spans = f.spans[1:]

		header := reply.Header()
		if header.NumKeys > 0 && func() bool {
			__antithesis_instrumentation__.Notify(568272)
			return f.newFetchSpansIdx > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(568273)
			return false, nil, nil, errors.Errorf(
				"span with results after resume span; it shouldn't happen given that "+
					"we're only scanning non-overlapping spans. New spans: %s",
				catalogkeys.PrettySpans(nil, f.spans, 0))
		} else {
			__antithesis_instrumentation__.Notify(568274)
		}
		__antithesis_instrumentation__.Notify(568270)

		if resumeSpan := header.ResumeSpan; resumeSpan != nil {
			__antithesis_instrumentation__.Notify(568275)
			f.spansScratch[f.newFetchSpansIdx] = *resumeSpan
			f.newFetchSpansIdx++
		} else {
			__antithesis_instrumentation__.Notify(568276)
		}
		__antithesis_instrumentation__.Notify(568271)

		switch t := reply.(type) {
		case *roachpb.ScanResponse:
			__antithesis_instrumentation__.Notify(568277)
			if len(t.BatchResponses) > 0 {
				__antithesis_instrumentation__.Notify(568288)
				batchResp, f.remainingBatches = popBatch(t.BatchResponses)
			} else {
				__antithesis_instrumentation__.Notify(568289)
			}
			__antithesis_instrumentation__.Notify(568278)
			if len(t.Rows) > 0 {
				__antithesis_instrumentation__.Notify(568290)
				return false, nil, nil, errors.AssertionFailedf(
					"unexpectedly got a ScanResponse using KEY_VALUES response format",
				)
			} else {
				__antithesis_instrumentation__.Notify(568291)
			}
			__antithesis_instrumentation__.Notify(568279)
			if len(t.IntentRows) > 0 {
				__antithesis_instrumentation__.Notify(568292)
				return false, nil, nil, errors.AssertionFailedf(
					"unexpectedly got a ScanResponse with non-nil IntentRows",
				)
			} else {
				__antithesis_instrumentation__.Notify(568293)
			}
			__antithesis_instrumentation__.Notify(568280)

			return true, nil, batchResp, nil
		case *roachpb.ReverseScanResponse:
			__antithesis_instrumentation__.Notify(568281)
			if len(t.BatchResponses) > 0 {
				__antithesis_instrumentation__.Notify(568294)
				batchResp, f.remainingBatches = popBatch(t.BatchResponses)
			} else {
				__antithesis_instrumentation__.Notify(568295)
			}
			__antithesis_instrumentation__.Notify(568282)
			if len(t.Rows) > 0 {
				__antithesis_instrumentation__.Notify(568296)
				return false, nil, nil, errors.AssertionFailedf(
					"unexpectedly got a ReverseScanResponse using KEY_VALUES response format",
				)
			} else {
				__antithesis_instrumentation__.Notify(568297)
			}
			__antithesis_instrumentation__.Notify(568283)
			if len(t.IntentRows) > 0 {
				__antithesis_instrumentation__.Notify(568298)
				return false, nil, nil, errors.AssertionFailedf(
					"unexpectedly got a ReverseScanResponse with non-nil IntentRows",
				)
			} else {
				__antithesis_instrumentation__.Notify(568299)
			}
			__antithesis_instrumentation__.Notify(568284)

			return true, nil, batchResp, nil
		case *roachpb.GetResponse:
			__antithesis_instrumentation__.Notify(568285)
			if t.IntentValue != nil {
				__antithesis_instrumentation__.Notify(568300)
				return false, nil, nil, errors.AssertionFailedf("unexpectedly got an IntentValue back from a SQL GetRequest %v", *t.IntentValue)
			} else {
				__antithesis_instrumentation__.Notify(568301)
			}
			__antithesis_instrumentation__.Notify(568286)
			if t.Value == nil {
				__antithesis_instrumentation__.Notify(568302)

				continue
			} else {
				__antithesis_instrumentation__.Notify(568303)
			}
			__antithesis_instrumentation__.Notify(568287)
			return true, []roachpb.KeyValue{{Key: origSpan.Key, Value: *t.Value}}, nil, nil
		}
	}
	__antithesis_instrumentation__.Notify(568264)

	if f.alreadyFetched {
		__antithesis_instrumentation__.Notify(568304)
		if f.newFetchSpansIdx == 0 {
			__antithesis_instrumentation__.Notify(568306)

			return false, nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(568307)
		}
		__antithesis_instrumentation__.Notify(568305)

		f.spans = f.spansScratch[:f.newFetchSpansIdx]
		if f.acc != nil {
			__antithesis_instrumentation__.Notify(568308)
			used := f.acc.Used()
			delta := f.spans.MemUsage() - f.spansAccountedFor
			if err := f.acc.Resize(ctx, used, used+delta); err != nil {
				__antithesis_instrumentation__.Notify(568310)
				return false, nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(568311)
			}
			__antithesis_instrumentation__.Notify(568309)
			f.spansAccountedFor += delta
		} else {
			__antithesis_instrumentation__.Notify(568312)
		}
	} else {
		__antithesis_instrumentation__.Notify(568313)
	}
	__antithesis_instrumentation__.Notify(568265)

	if err := f.fetch(ctx); err != nil {
		__antithesis_instrumentation__.Notify(568314)
		return false, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(568315)
	}
	__antithesis_instrumentation__.Notify(568266)

	return f.nextBatch(ctx)
}

func (f *txnKVFetcher) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568316)
	f.responses = nil
	f.remainingBatches = nil
	f.spans = nil
	f.spansScratch = nil
	f.acc.Clear(ctx)
}

func spansToRequests(
	spans roachpb.Spans, reverse bool, keyLocking lock.Strength,
) []roachpb.RequestUnion {
	__antithesis_instrumentation__.Notify(568317)
	reqs := make([]roachpb.RequestUnion, len(spans))

	nGets := 0
	for i := range spans {
		__antithesis_instrumentation__.Notify(568320)
		if spans[i].EndKey == nil {
			__antithesis_instrumentation__.Notify(568321)
			nGets++
		} else {
			__antithesis_instrumentation__.Notify(568322)
		}
	}
	__antithesis_instrumentation__.Notify(568318)
	gets := make([]struct {
		req   roachpb.GetRequest
		union roachpb.RequestUnion_Get
	}, nGets)

	curGet := 0
	if reverse {
		__antithesis_instrumentation__.Notify(568323)
		scans := make([]struct {
			req   roachpb.ReverseScanRequest
			union roachpb.RequestUnion_ReverseScan
		}, len(spans)-nGets)
		for i := range spans {
			__antithesis_instrumentation__.Notify(568324)
			if spans[i].EndKey == nil {
				__antithesis_instrumentation__.Notify(568326)

				gets[curGet].req.Key = spans[i].Key
				gets[curGet].req.KeyLocking = keyLocking
				gets[curGet].union.Get = &gets[curGet].req
				reqs[i].Value = &gets[curGet].union
				curGet++
				continue
			} else {
				__antithesis_instrumentation__.Notify(568327)
			}
			__antithesis_instrumentation__.Notify(568325)
			curScan := i - curGet
			scans[curScan].req.SetSpan(spans[i])
			scans[curScan].req.ScanFormat = roachpb.BATCH_RESPONSE
			scans[curScan].req.KeyLocking = keyLocking
			scans[curScan].union.ReverseScan = &scans[curScan].req
			reqs[i].Value = &scans[curScan].union
		}
	} else {
		__antithesis_instrumentation__.Notify(568328)
		scans := make([]struct {
			req   roachpb.ScanRequest
			union roachpb.RequestUnion_Scan
		}, len(spans)-nGets)
		for i := range spans {
			__antithesis_instrumentation__.Notify(568329)
			if spans[i].EndKey == nil {
				__antithesis_instrumentation__.Notify(568331)

				gets[curGet].req.Key = spans[i].Key
				gets[curGet].req.KeyLocking = keyLocking
				gets[curGet].union.Get = &gets[curGet].req
				reqs[i].Value = &gets[curGet].union
				curGet++
				continue
			} else {
				__antithesis_instrumentation__.Notify(568332)
			}
			__antithesis_instrumentation__.Notify(568330)
			curScan := i - curGet
			scans[curScan].req.SetSpan(spans[i])
			scans[curScan].req.ScanFormat = roachpb.BATCH_RESPONSE
			scans[curScan].req.KeyLocking = keyLocking
			scans[curScan].union.Scan = &scans[curScan].req
			reqs[i].Value = &scans[curScan].union
		}
	}
	__antithesis_instrumentation__.Notify(568319)
	return reqs
}
