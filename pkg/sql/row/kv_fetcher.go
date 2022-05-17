package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type KVFetcher struct {
	KVBatchFetcher

	kvs []roachpb.KeyValue

	batchResponse []byte
	newSpan       bool

	atomics struct {
		bytesRead int64
	}
}

func NewKVFetcher(
	ctx context.Context,
	txn *kv.Txn,
	spans roachpb.Spans,
	bsHeader *roachpb.BoundedStalenessHeader,
	reverse bool,
	batchBytesLimit rowinfra.BytesLimit,
	firstBatchLimit rowinfra.KeyLimit,
	lockStrength descpb.ScanLockingStrength,
	lockWaitPolicy descpb.ScanLockingWaitPolicy,
	lockTimeout time.Duration,
	acc *mon.BoundAccount,
	forceProductionKVBatchSize bool,
) (*KVFetcher, error) {
	__antithesis_instrumentation__.Notify(568385)
	var sendFn sendFunc

	if bsHeader == nil {
		__antithesis_instrumentation__.Notify(568387)
		sendFn = makeKVBatchFetcherDefaultSendFunc(txn)
	} else {
		__antithesis_instrumentation__.Notify(568388)
		negotiated := false
		sendFn = func(ctx context.Context, ba roachpb.BatchRequest) (br *roachpb.BatchResponse, _ error) {
			__antithesis_instrumentation__.Notify(568389)
			ba.RoutingPolicy = roachpb.RoutingPolicy_NEAREST
			var pErr *roachpb.Error

			if !negotiated {
				__antithesis_instrumentation__.Notify(568392)
				ba.BoundedStaleness = bsHeader
				br, pErr = txn.NegotiateAndSend(ctx, ba)
				negotiated = true
			} else {
				__antithesis_instrumentation__.Notify(568393)
				br, pErr = txn.Send(ctx, ba)
			}
			__antithesis_instrumentation__.Notify(568390)
			if pErr != nil {
				__antithesis_instrumentation__.Notify(568394)
				return nil, pErr.GoError()
			} else {
				__antithesis_instrumentation__.Notify(568395)
			}
			__antithesis_instrumentation__.Notify(568391)
			return br, nil
		}
	}
	__antithesis_instrumentation__.Notify(568386)

	kvBatchFetcher, err := makeKVBatchFetcher(
		ctx,
		sendFn,
		spans,
		reverse,
		batchBytesLimit,
		firstBatchLimit,
		lockStrength,
		lockWaitPolicy,
		lockTimeout,
		acc,
		forceProductionKVBatchSize,
		txn.AdmissionHeader(),
		txn.DB().SQLKVResponseAdmissionQ,
	)
	return newKVFetcher(&kvBatchFetcher), err
}

func NewKVStreamingFetcher(streamer *TxnKVStreamer) *KVFetcher {
	__antithesis_instrumentation__.Notify(568396)
	return &KVFetcher{
		KVBatchFetcher: streamer,
	}
}

func newKVFetcher(batchFetcher KVBatchFetcher) *KVFetcher {
	__antithesis_instrumentation__.Notify(568397)
	return &KVFetcher{
		KVBatchFetcher: batchFetcher,
	}
}

func (f *KVFetcher) GetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(568398)
	if f == nil {
		__antithesis_instrumentation__.Notify(568400)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(568401)
	}
	__antithesis_instrumentation__.Notify(568399)
	return atomic.LoadInt64(&f.atomics.bytesRead)
}

func (f *KVFetcher) ResetBytesRead() int64 {
	__antithesis_instrumentation__.Notify(568402)
	if f == nil {
		__antithesis_instrumentation__.Notify(568404)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(568405)
	}
	__antithesis_instrumentation__.Notify(568403)
	return atomic.SwapInt64(&f.atomics.bytesRead, 0)
}

type MVCCDecodingStrategy int

const (
	MVCCDecodingNotRequired MVCCDecodingStrategy = iota

	MVCCDecodingRequired
)

func (f *KVFetcher) NextKV(
	ctx context.Context, mvccDecodeStrategy MVCCDecodingStrategy,
) (ok bool, kv roachpb.KeyValue, finalReferenceToBatch bool, err error) {
	__antithesis_instrumentation__.Notify(568406)
	for {
		__antithesis_instrumentation__.Notify(568407)

		nKvs := len(f.kvs)
		if nKvs != 0 {
			__antithesis_instrumentation__.Notify(568412)
			kv = f.kvs[0]
			f.kvs = f.kvs[1:]

			return true, kv, false, nil
		} else {
			__antithesis_instrumentation__.Notify(568413)
		}
		__antithesis_instrumentation__.Notify(568408)
		if len(f.batchResponse) > 0 {
			__antithesis_instrumentation__.Notify(568414)
			var key []byte
			var rawBytes []byte
			var err error
			var ts hlc.Timestamp
			switch mvccDecodeStrategy {
			case MVCCDecodingRequired:
				__antithesis_instrumentation__.Notify(568418)
				key, ts, rawBytes, f.batchResponse, err = enginepb.ScanDecodeKeyValue(f.batchResponse)
			case MVCCDecodingNotRequired:
				__antithesis_instrumentation__.Notify(568419)
				key, rawBytes, f.batchResponse, err = enginepb.ScanDecodeKeyValueNoTS(f.batchResponse)
			default:
				__antithesis_instrumentation__.Notify(568420)
			}
			__antithesis_instrumentation__.Notify(568415)
			if err != nil {
				__antithesis_instrumentation__.Notify(568421)
				return false, kv, false, err
			} else {
				__antithesis_instrumentation__.Notify(568422)
			}
			__antithesis_instrumentation__.Notify(568416)

			lastKey := len(f.batchResponse) == 0
			if lastKey {
				__antithesis_instrumentation__.Notify(568423)
				f.batchResponse = nil
			} else {
				__antithesis_instrumentation__.Notify(568424)
			}
			__antithesis_instrumentation__.Notify(568417)
			return true, roachpb.KeyValue{
				Key: key[:len(key):len(key)],
				Value: roachpb.Value{
					RawBytes:  rawBytes[:len(rawBytes):len(rawBytes)],
					Timestamp: ts,
				},
			}, lastKey, nil
		} else {
			__antithesis_instrumentation__.Notify(568425)
		}
		__antithesis_instrumentation__.Notify(568409)

		ok, f.kvs, f.batchResponse, err = f.nextBatch(ctx)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(568426)
			return !ok == true
		}() == true {
			__antithesis_instrumentation__.Notify(568427)
			return ok, kv, false, err
		} else {
			__antithesis_instrumentation__.Notify(568428)
		}
		__antithesis_instrumentation__.Notify(568410)
		f.newSpan = true
		nBytes := len(f.batchResponse)
		for i := range f.kvs {
			__antithesis_instrumentation__.Notify(568429)
			nBytes += len(f.kvs[i].Key)
			nBytes += len(f.kvs[i].Value.RawBytes)
		}
		__antithesis_instrumentation__.Notify(568411)
		atomic.AddInt64(&f.atomics.bytesRead, int64(nBytes))
	}
}

func (f *KVFetcher) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568430)
	f.KVBatchFetcher.close(ctx)
}

type SpanKVFetcher struct {
	KVs []roachpb.KeyValue
}

func (f *SpanKVFetcher) nextBatch(
	ctx context.Context,
) (ok bool, kvs []roachpb.KeyValue, batchResponse []byte, err error) {
	__antithesis_instrumentation__.Notify(568431)
	if len(f.KVs) == 0 {
		__antithesis_instrumentation__.Notify(568433)
		return false, nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(568434)
	}
	__antithesis_instrumentation__.Notify(568432)
	res := f.KVs
	f.KVs = nil
	return true, res, nil, nil
}

func (f *SpanKVFetcher) close(context.Context) { __antithesis_instrumentation__.Notify(568435) }

type BackupSSTKVFetcher struct {
	iter          storage.SimpleMVCCIterator
	endKeyMVCC    storage.MVCCKey
	startTime     hlc.Timestamp
	endTime       hlc.Timestamp
	withRevisions bool
}

func MakeBackupSSTKVFetcher(
	startKeyMVCC, endKeyMVCC storage.MVCCKey,
	iter storage.SimpleMVCCIterator,
	startTime hlc.Timestamp,
	endTime hlc.Timestamp,
	withRev bool,
) BackupSSTKVFetcher {
	__antithesis_instrumentation__.Notify(568436)
	res := BackupSSTKVFetcher{
		iter,
		endKeyMVCC,
		startTime,
		endTime,
		withRev,
	}
	res.iter.SeekGE(startKeyMVCC)
	return res
}

func (f *BackupSSTKVFetcher) nextBatch(
	ctx context.Context,
) (ok bool, kvs []roachpb.KeyValue, batchResponse []byte, err error) {
	__antithesis_instrumentation__.Notify(568437)
	res := make([]roachpb.KeyValue, 0)

	copyKV := func(mvccKey storage.MVCCKey, value []byte) roachpb.KeyValue {
		__antithesis_instrumentation__.Notify(568441)
		keyCopy := make([]byte, len(mvccKey.Key))
		copy(keyCopy, mvccKey.Key)
		valueCopy := make([]byte, len(value))
		copy(valueCopy, value)
		return roachpb.KeyValue{
			Key:   keyCopy,
			Value: roachpb.Value{RawBytes: valueCopy, Timestamp: mvccKey.Timestamp},
		}
	}
	__antithesis_instrumentation__.Notify(568438)

	for {
		__antithesis_instrumentation__.Notify(568442)
		valid, err := f.iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(568447)
			err = errors.Wrapf(err, "iter key value of table data")
			return false, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(568448)
		}
		__antithesis_instrumentation__.Notify(568443)

		if !valid || func() bool {
			__antithesis_instrumentation__.Notify(568449)
			return !f.iter.UnsafeKey().Less(f.endKeyMVCC) == true
		}() == true {
			__antithesis_instrumentation__.Notify(568450)
			break
		} else {
			__antithesis_instrumentation__.Notify(568451)
		}
		__antithesis_instrumentation__.Notify(568444)

		if !f.endTime.IsEmpty() {
			__antithesis_instrumentation__.Notify(568452)
			if f.endTime.Less(f.iter.UnsafeKey().Timestamp) {
				__antithesis_instrumentation__.Notify(568453)
				f.iter.Next()
				continue
			} else {
				__antithesis_instrumentation__.Notify(568454)
			}
		} else {
			__antithesis_instrumentation__.Notify(568455)
		}
		__antithesis_instrumentation__.Notify(568445)

		if f.withRevisions {
			__antithesis_instrumentation__.Notify(568456)
			if f.iter.UnsafeKey().Timestamp.Less(f.startTime) {
				__antithesis_instrumentation__.Notify(568457)
				f.iter.NextKey()
				continue
			} else {
				__antithesis_instrumentation__.Notify(568458)
			}
		} else {
			__antithesis_instrumentation__.Notify(568459)
			if len(f.iter.UnsafeValue()) == 0 {
				__antithesis_instrumentation__.Notify(568460)
				if f.endTime.IsEmpty() || func() bool {
					__antithesis_instrumentation__.Notify(568461)
					return f.iter.UnsafeKey().Timestamp.Less(f.endTime) == true
				}() == true {
					__antithesis_instrumentation__.Notify(568462)

					f.iter.NextKey()
					continue
				} else {
					__antithesis_instrumentation__.Notify(568463)

					f.iter.Next()
					continue
				}
			} else {
				__antithesis_instrumentation__.Notify(568464)
			}
		}
		__antithesis_instrumentation__.Notify(568446)

		res = append(res, copyKV(f.iter.UnsafeKey(), f.iter.UnsafeValue()))

		if f.withRevisions {
			__antithesis_instrumentation__.Notify(568465)
			f.iter.Next()
		} else {
			__antithesis_instrumentation__.Notify(568466)
			f.iter.NextKey()
		}

	}
	__antithesis_instrumentation__.Notify(568439)
	if len(res) == 0 {
		__antithesis_instrumentation__.Notify(568467)
		return false, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(568468)
	}
	__antithesis_instrumentation__.Notify(568440)
	return true, res, nil, nil
}

func (f *BackupSSTKVFetcher) close(context.Context) {
	__antithesis_instrumentation__.Notify(568469)
	f.iter.Close()
}
