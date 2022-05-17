package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type intentDemuxWriter struct {
	w Writer
}

func wrapIntentWriter(ctx context.Context, w Writer) intentDemuxWriter {
	__antithesis_instrumentation__.Notify(639595)
	idw := intentDemuxWriter{w: w}
	return idw
}

func (idw intentDemuxWriter) ClearIntent(
	key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID, buf []byte,
) (_ []byte, _ error) {
	__antithesis_instrumentation__.Notify(639596)
	var engineKey EngineKey
	engineKey, buf = LockTableKey{
		Key:      key,
		Strength: lock.Exclusive,
		TxnUUID:  txnUUID[:],
	}.ToEngineKey(buf)
	if txnDidNotUpdateMeta {
		__antithesis_instrumentation__.Notify(639598)
		return buf, idw.w.SingleClearEngineKey(engineKey)
	} else {
		__antithesis_instrumentation__.Notify(639599)
	}
	__antithesis_instrumentation__.Notify(639597)
	return buf, idw.w.ClearEngineKey(engineKey)
}

func (idw intentDemuxWriter) PutIntent(
	ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID, buf []byte,
) (_ []byte, _ error) {
	__antithesis_instrumentation__.Notify(639600)
	var engineKey EngineKey
	engineKey, buf = LockTableKey{
		Key:      key,
		Strength: lock.Exclusive,
		TxnUUID:  txnUUID[:],
	}.ToEngineKey(buf)
	return buf, idw.w.PutEngineKey(engineKey, value)
}

func (idw intentDemuxWriter) ClearMVCCRangeAndIntents(
	start, end roachpb.Key, buf []byte,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(639601)
	err := idw.w.ClearRawRange(start, end)
	if err != nil {
		__antithesis_instrumentation__.Notify(639603)
		return buf, err
	} else {
		__antithesis_instrumentation__.Notify(639604)
	}
	__antithesis_instrumentation__.Notify(639602)
	lstart, buf := keys.LockTableSingleKey(start, buf)
	lend, _ := keys.LockTableSingleKey(end, nil)
	return buf, idw.w.ClearRawRange(lstart, lend)
}

type wrappableReader interface {
	Reader

	rawMVCCGet(key []byte) (value []byte, err error)
}

func wrapReader(r wrappableReader) *intentInterleavingReader {
	__antithesis_instrumentation__.Notify(639605)
	iiReader := intentInterleavingReaderPool.Get().(*intentInterleavingReader)
	*iiReader = intentInterleavingReader{wrappableReader: r}
	return iiReader
}

type intentInterleavingReader struct {
	wrappableReader
}

var _ Reader = &intentInterleavingReader{}

var intentInterleavingReaderPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(639606)
		return &intentInterleavingReader{}
	},
}

func (imr *intentInterleavingReader) MVCCGet(key MVCCKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(639607)
	val, err := imr.wrappableReader.rawMVCCGet(EncodeMVCCKey(key))
	if val != nil || func() bool {
		__antithesis_instrumentation__.Notify(639610)
		return err != nil == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(639611)
		return !key.Timestamp.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(639612)
		return val, err
	} else {
		__antithesis_instrumentation__.Notify(639613)
	}
	__antithesis_instrumentation__.Notify(639608)

	ltKey, _ := keys.LockTableSingleKey(key.Key, nil)
	iter := imr.wrappableReader.NewEngineIterator(IterOptions{Prefix: true, LowerBound: ltKey})
	defer iter.Close()
	valid, err := iter.SeekEngineKeyGE(EngineKey{Key: ltKey})
	if !valid || func() bool {
		__antithesis_instrumentation__.Notify(639614)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(639615)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(639616)
	}
	__antithesis_instrumentation__.Notify(639609)
	val = iter.Value()
	return val, nil
}

func (imr *intentInterleavingReader) MVCCGetProto(
	key MVCCKey, msg protoutil.Message,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(639617)
	return pebbleGetProto(imr, key, msg)
}

func (imr *intentInterleavingReader) NewMVCCIterator(
	iterKind MVCCIterKind, opts IterOptions,
) MVCCIterator {
	__antithesis_instrumentation__.Notify(639618)
	if (!opts.MinTimestampHint.IsEmpty() || func() bool {
		__antithesis_instrumentation__.Notify(639621)
		return !opts.MaxTimestampHint.IsEmpty() == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(639622)
		return iterKind == MVCCKeyAndIntentsIterKind == true
	}() == true {
		__antithesis_instrumentation__.Notify(639623)
		panic("cannot ask for interleaved intents when specifying timestamp hints")
	} else {
		__antithesis_instrumentation__.Notify(639624)
	}
	__antithesis_instrumentation__.Notify(639619)
	if iterKind == MVCCKeyIterKind {
		__antithesis_instrumentation__.Notify(639625)
		return imr.wrappableReader.NewMVCCIterator(MVCCKeyIterKind, opts)
	} else {
		__antithesis_instrumentation__.Notify(639626)
	}
	__antithesis_instrumentation__.Notify(639620)
	return newIntentInterleavingIterator(imr.wrappableReader, opts)
}

func (imr *intentInterleavingReader) Free() {
	__antithesis_instrumentation__.Notify(639627)
	*imr = intentInterleavingReader{}
	intentInterleavingReaderPool.Put(imr)
}
