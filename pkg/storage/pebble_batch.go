package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

type pebbleBatch struct {
	db    *pebble.DB
	batch *pebble.Batch
	buf   []byte

	prefixIter       pebbleIterator
	normalIter       pebbleIterator
	prefixEngineIter pebbleIterator
	normalEngineIter pebbleIterator
	iter             cloneableIter
	writeOnly        bool
	closed           bool

	wrappedIntentWriter intentDemuxWriter

	scratch []byte
}

var _ Batch = &pebbleBatch{}

var pebbleBatchPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(642455)
		return &pebbleBatch{}
	},
}

func newPebbleBatch(db *pebble.DB, batch *pebble.Batch, writeOnly bool) *pebbleBatch {
	__antithesis_instrumentation__.Notify(642456)
	pb := pebbleBatchPool.Get().(*pebbleBatch)
	*pb = pebbleBatch{
		db:    db,
		batch: batch,
		buf:   pb.buf,
		prefixIter: pebbleIterator{
			lowerBoundBuf: pb.prefixIter.lowerBoundBuf,
			upperBoundBuf: pb.prefixIter.upperBoundBuf,
			reusable:      true,
		},
		normalIter: pebbleIterator{
			lowerBoundBuf: pb.normalIter.lowerBoundBuf,
			upperBoundBuf: pb.normalIter.upperBoundBuf,
			reusable:      true,
		},
		prefixEngineIter: pebbleIterator{
			lowerBoundBuf: pb.prefixEngineIter.lowerBoundBuf,
			upperBoundBuf: pb.prefixEngineIter.upperBoundBuf,
			reusable:      true,
		},
		normalEngineIter: pebbleIterator{
			lowerBoundBuf: pb.normalEngineIter.lowerBoundBuf,
			upperBoundBuf: pb.normalEngineIter.upperBoundBuf,
			reusable:      true,
		},
		writeOnly: writeOnly,
	}
	pb.wrappedIntentWriter = wrapIntentWriter(context.Background(), pb)
	return pb
}

func (p *pebbleBatch) Close() {
	__antithesis_instrumentation__.Notify(642457)
	if p.closed {
		__antithesis_instrumentation__.Notify(642459)
		panic("closing an already-closed pebbleBatch")
	} else {
		__antithesis_instrumentation__.Notify(642460)
	}
	__antithesis_instrumentation__.Notify(642458)
	p.closed = true

	p.iter = nil

	p.prefixIter.destroy()
	p.normalIter.destroy()
	p.prefixEngineIter.destroy()
	p.normalEngineIter.destroy()

	_ = p.batch.Close()
	p.batch = nil

	pebbleBatchPool.Put(p)
}

func (p *pebbleBatch) Closed() bool {
	__antithesis_instrumentation__.Notify(642461)
	return p.closed
}

func (p *pebbleBatch) ExportMVCCToSst(
	ctx context.Context, exportOptions ExportOptions, dest io.Writer,
) (roachpb.BulkOpSummary, roachpb.Key, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(642462)
	panic("unimplemented")
}

func (p *pebbleBatch) MVCCGet(key MVCCKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642463)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642465)
		return nil, emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642466)
	}
	__antithesis_instrumentation__.Notify(642464)
	r := wrapReader(p)

	v, err := r.MVCCGet(key)
	r.Free()
	return v, err
}

func (p *pebbleBatch) rawMVCCGet(key []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(642467)
	r := pebble.Reader(p.batch)
	if p.writeOnly {
		__antithesis_instrumentation__.Notify(642472)
		panic("write-only batch")
	} else {
		__antithesis_instrumentation__.Notify(642473)
	}
	__antithesis_instrumentation__.Notify(642468)
	if !p.batch.Indexed() {
		__antithesis_instrumentation__.Notify(642474)
		r = p.db
	} else {
		__antithesis_instrumentation__.Notify(642475)
	}
	__antithesis_instrumentation__.Notify(642469)

	ret, closer, err := r.Get(key)
	if closer != nil {
		__antithesis_instrumentation__.Notify(642476)
		retCopy := make([]byte, len(ret))
		copy(retCopy, ret)
		ret = retCopy
		closer.Close()
	} else {
		__antithesis_instrumentation__.Notify(642477)
	}
	__antithesis_instrumentation__.Notify(642470)
	if errors.Is(err, pebble.ErrNotFound) || func() bool {
		__antithesis_instrumentation__.Notify(642478)
		return len(ret) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(642479)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(642480)
	}
	__antithesis_instrumentation__.Notify(642471)
	return ret, err
}

func (p *pebbleBatch) MVCCGetProto(
	key MVCCKey, msg protoutil.Message,
) (ok bool, keyBytes, valBytes int64, err error) {
	__antithesis_instrumentation__.Notify(642481)
	return pebbleGetProto(p, key, msg)
}

func (p *pebbleBatch) MVCCIterate(
	start, end roachpb.Key, iterKind MVCCIterKind, f func(MVCCKeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(642482)
	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(642484)
		r := wrapReader(p)

		err := iterateOnReader(r, start, end, iterKind, f)
		r.Free()
		return err
	} else {
		__antithesis_instrumentation__.Notify(642485)
	}
	__antithesis_instrumentation__.Notify(642483)
	return iterateOnReader(p, start, end, iterKind, f)
}

func (p *pebbleBatch) NewMVCCIterator(iterKind MVCCIterKind, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(642486)
	if !opts.Prefix && func() bool {
		__antithesis_instrumentation__.Notify(642495)
		return len(opts.UpperBound) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(642496)
		return len(opts.LowerBound) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(642497)
		panic("iterator must set prefix or upper bound or lower bound")
	} else {
		__antithesis_instrumentation__.Notify(642498)
	}
	__antithesis_instrumentation__.Notify(642487)

	if p.writeOnly {
		__antithesis_instrumentation__.Notify(642499)
		panic("write-only batch")
	} else {
		__antithesis_instrumentation__.Notify(642500)
	}
	__antithesis_instrumentation__.Notify(642488)

	if iterKind == MVCCKeyAndIntentsIterKind {
		__antithesis_instrumentation__.Notify(642501)
		r := wrapReader(p)

		iter := r.NewMVCCIterator(iterKind, opts)
		r.Free()
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(642503)
			iter = wrapInUnsafeIter(iter)
		} else {
			__antithesis_instrumentation__.Notify(642504)
		}
		__antithesis_instrumentation__.Notify(642502)
		return iter
	} else {
		__antithesis_instrumentation__.Notify(642505)
	}
	__antithesis_instrumentation__.Notify(642489)

	if !opts.MinTimestampHint.IsEmpty() {
		__antithesis_instrumentation__.Notify(642506)

		iter := MVCCIterator(newPebbleIterator(p.batch, nil, opts, StandardDurability))
		if util.RaceEnabled {
			__antithesis_instrumentation__.Notify(642508)
			iter = wrapInUnsafeIter(iter)
		} else {
			__antithesis_instrumentation__.Notify(642509)
		}
		__antithesis_instrumentation__.Notify(642507)
		return iter
	} else {
		__antithesis_instrumentation__.Notify(642510)
	}
	__antithesis_instrumentation__.Notify(642490)

	iter := &p.normalIter
	if opts.Prefix {
		__antithesis_instrumentation__.Notify(642511)
		iter = &p.prefixIter
	} else {
		__antithesis_instrumentation__.Notify(642512)
	}
	__antithesis_instrumentation__.Notify(642491)
	if iter.inuse {
		__antithesis_instrumentation__.Notify(642513)
		panic("iterator already in use")
	} else {
		__antithesis_instrumentation__.Notify(642514)
	}
	__antithesis_instrumentation__.Notify(642492)

	checkOptionsForIterReuse(opts)

	if iter.iter != nil {
		__antithesis_instrumentation__.Notify(642515)
		iter.setBounds(opts.LowerBound, opts.UpperBound)
	} else {
		__antithesis_instrumentation__.Notify(642516)
		if p.batch.Indexed() {
			__antithesis_instrumentation__.Notify(642518)
			iter.init(p.batch, p.iter, opts, StandardDurability)
		} else {
			__antithesis_instrumentation__.Notify(642519)
			iter.init(p.db, p.iter, opts, StandardDurability)
		}
		__antithesis_instrumentation__.Notify(642517)
		if p.iter == nil {
			__antithesis_instrumentation__.Notify(642520)

			p.iter = iter.iter
		} else {
			__antithesis_instrumentation__.Notify(642521)
		}
	}
	__antithesis_instrumentation__.Notify(642493)

	iter.inuse = true
	var rv MVCCIterator = iter
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(642522)
		rv = wrapInUnsafeIter(rv)
	} else {
		__antithesis_instrumentation__.Notify(642523)
	}
	__antithesis_instrumentation__.Notify(642494)
	return rv
}

func (p *pebbleBatch) NewEngineIterator(opts IterOptions) EngineIterator {
	__antithesis_instrumentation__.Notify(642524)
	if !opts.Prefix && func() bool {
		__antithesis_instrumentation__.Notify(642530)
		return len(opts.UpperBound) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(642531)
		return len(opts.LowerBound) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(642532)
		panic("iterator must set prefix or upper bound or lower bound")
	} else {
		__antithesis_instrumentation__.Notify(642533)
	}
	__antithesis_instrumentation__.Notify(642525)

	if p.writeOnly {
		__antithesis_instrumentation__.Notify(642534)
		panic("write-only batch")
	} else {
		__antithesis_instrumentation__.Notify(642535)
	}
	__antithesis_instrumentation__.Notify(642526)

	iter := &p.normalEngineIter
	if opts.Prefix {
		__antithesis_instrumentation__.Notify(642536)
		iter = &p.prefixEngineIter
	} else {
		__antithesis_instrumentation__.Notify(642537)
	}
	__antithesis_instrumentation__.Notify(642527)
	if iter.inuse {
		__antithesis_instrumentation__.Notify(642538)
		panic("iterator already in use")
	} else {
		__antithesis_instrumentation__.Notify(642539)
	}
	__antithesis_instrumentation__.Notify(642528)

	checkOptionsForIterReuse(opts)

	if iter.iter != nil {
		__antithesis_instrumentation__.Notify(642540)
		iter.setBounds(opts.LowerBound, opts.UpperBound)
	} else {
		__antithesis_instrumentation__.Notify(642541)
		if p.batch.Indexed() {
			__antithesis_instrumentation__.Notify(642543)
			iter.init(p.batch, p.iter, opts, StandardDurability)
		} else {
			__antithesis_instrumentation__.Notify(642544)
			iter.init(p.db, p.iter, opts, StandardDurability)
		}
		__antithesis_instrumentation__.Notify(642542)
		if p.iter == nil {
			__antithesis_instrumentation__.Notify(642545)

			p.iter = iter.iter
		} else {
			__antithesis_instrumentation__.Notify(642546)
		}
	}
	__antithesis_instrumentation__.Notify(642529)

	iter.inuse = true
	return iter
}

func (p *pebbleBatch) ConsistentIterators() bool {
	__antithesis_instrumentation__.Notify(642547)
	return true
}

func (p *pebbleBatch) PinEngineStateForIterators() error {
	__antithesis_instrumentation__.Notify(642548)
	if p.iter == nil {
		__antithesis_instrumentation__.Notify(642550)
		if p.batch.Indexed() {
			__antithesis_instrumentation__.Notify(642551)
			p.iter = p.batch.NewIter(nil)
		} else {
			__antithesis_instrumentation__.Notify(642552)
			p.iter = p.db.NewIter(nil)
		}
	} else {
		__antithesis_instrumentation__.Notify(642553)
	}
	__antithesis_instrumentation__.Notify(642549)
	return nil
}

func (p *pebbleBatch) ApplyBatchRepr(repr []byte, sync bool) error {
	__antithesis_instrumentation__.Notify(642554)
	var batch pebble.Batch
	if err := batch.SetRepr(repr); err != nil {
		__antithesis_instrumentation__.Notify(642556)
		return err
	} else {
		__antithesis_instrumentation__.Notify(642557)
	}
	__antithesis_instrumentation__.Notify(642555)

	return p.batch.Apply(&batch, nil)
}

func (p *pebbleBatch) ClearMVCC(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(642558)
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(642560)
		panic("ClearMVCC timestamp is empty")
	} else {
		__antithesis_instrumentation__.Notify(642561)
	}
	__antithesis_instrumentation__.Notify(642559)
	return p.clear(key)
}

func (p *pebbleBatch) ClearUnversioned(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642562)
	return p.clear(MVCCKey{Key: key})
}

func (p *pebbleBatch) ClearIntent(
	key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(642563)
	var err error
	p.scratch, err = p.wrappedIntentWriter.ClearIntent(key, txnDidNotUpdateMeta, txnUUID, p.scratch)
	return err
}

func (p *pebbleBatch) ClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(642564)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642566)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642567)
	}
	__antithesis_instrumentation__.Notify(642565)
	p.buf = key.EncodeToBuf(p.buf[:0])
	return p.batch.Delete(p.buf, nil)
}

func (p *pebbleBatch) clear(key MVCCKey) error {
	__antithesis_instrumentation__.Notify(642568)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642570)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642571)
	}
	__antithesis_instrumentation__.Notify(642569)

	p.buf = EncodeMVCCKeyToBuf(p.buf[:0], key)
	return p.batch.Delete(p.buf, nil)
}

func (p *pebbleBatch) SingleClearEngineKey(key EngineKey) error {
	__antithesis_instrumentation__.Notify(642572)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642574)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642575)
	}
	__antithesis_instrumentation__.Notify(642573)

	p.buf = key.EncodeToBuf(p.buf[:0])
	return p.batch.SingleDelete(p.buf, nil)
}

func (p *pebbleBatch) ClearRawRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642576)
	return p.clearRange(MVCCKey{Key: start}, MVCCKey{Key: end})
}

func (p *pebbleBatch) ClearMVCCRangeAndIntents(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642577)
	var err error
	p.scratch, err = p.wrappedIntentWriter.ClearMVCCRangeAndIntents(start, end, p.scratch)
	return err
}

func (p *pebbleBatch) ClearMVCCRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(642578)
	return p.clearRange(start, end)
}

func (p *pebbleBatch) clearRange(start, end MVCCKey) error {
	__antithesis_instrumentation__.Notify(642579)
	p.buf = EncodeMVCCKeyToBuf(p.buf[:0], start)
	buf2 := EncodeMVCCKey(end)
	return p.batch.DeleteRange(p.buf, buf2, nil)
}

func (p *pebbleBatch) ClearIterRange(iter MVCCIterator, start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(642580)

	iter.SetUpperBound(end)
	iter.SeekGE(MakeMVCCMetadataKey(start))

	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(642582)
		valid, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(642584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(642585)
			if !valid {
				__antithesis_instrumentation__.Notify(642586)
				break
			} else {
				__antithesis_instrumentation__.Notify(642587)
			}
		}
		__antithesis_instrumentation__.Notify(642583)

		err = p.batch.Delete(iter.UnsafeRawKey(), nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(642588)
			return err
		} else {
			__antithesis_instrumentation__.Notify(642589)
		}
	}
	__antithesis_instrumentation__.Notify(642581)
	return nil
}

func (p *pebbleBatch) Merge(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642590)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642592)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642593)
	}
	__antithesis_instrumentation__.Notify(642591)

	p.buf = EncodeMVCCKeyToBuf(p.buf[:0], key)
	return p.batch.Merge(p.buf, value, nil)
}

func (p *pebbleBatch) PutMVCC(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642594)
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(642596)
		panic("PutMVCC timestamp is empty")
	} else {
		__antithesis_instrumentation__.Notify(642597)
	}
	__antithesis_instrumentation__.Notify(642595)
	return p.put(key, value)
}

func (p *pebbleBatch) PutUnversioned(key roachpb.Key, value []byte) error {
	__antithesis_instrumentation__.Notify(642598)
	return p.put(MVCCKey{Key: key}, value)
}

func (p *pebbleBatch) PutIntent(
	ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(642599)
	var err error
	p.scratch, err = p.wrappedIntentWriter.PutIntent(ctx, key, value, txnUUID, p.scratch)
	return err
}

func (p *pebbleBatch) PutEngineKey(key EngineKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642600)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642602)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642603)
	}
	__antithesis_instrumentation__.Notify(642601)

	p.buf = key.EncodeToBuf(p.buf[:0])
	return p.batch.Set(p.buf, value, nil)
}

func (p *pebbleBatch) put(key MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(642604)
	if len(key.Key) == 0 {
		__antithesis_instrumentation__.Notify(642606)
		return emptyKeyError()
	} else {
		__antithesis_instrumentation__.Notify(642607)
	}
	__antithesis_instrumentation__.Notify(642605)

	p.buf = EncodeMVCCKeyToBuf(p.buf[:0], key)
	return p.batch.Set(p.buf, value, nil)
}

func (p *pebbleBatch) LogData(data []byte) error {
	__antithesis_instrumentation__.Notify(642608)
	return p.batch.LogData(data, nil)
}

func (p *pebbleBatch) LogLogicalOp(op MVCCLogicalOpType, details MVCCLogicalOpDetails) {
	__antithesis_instrumentation__.Notify(642609)

}

func (p *pebbleBatch) Commit(sync bool) error {
	__antithesis_instrumentation__.Notify(642610)
	opts := pebble.NoSync
	if sync {
		__antithesis_instrumentation__.Notify(642614)
		opts = pebble.Sync
	} else {
		__antithesis_instrumentation__.Notify(642615)
	}
	__antithesis_instrumentation__.Notify(642611)
	if p.batch == nil {
		__antithesis_instrumentation__.Notify(642616)
		panic("called with nil batch")
	} else {
		__antithesis_instrumentation__.Notify(642617)
	}
	__antithesis_instrumentation__.Notify(642612)
	err := p.batch.Commit(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(642618)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(642619)
	}
	__antithesis_instrumentation__.Notify(642613)
	return err
}

func (p *pebbleBatch) Empty() bool {
	__antithesis_instrumentation__.Notify(642620)
	return p.batch.Count() == 0
}

func (p *pebbleBatch) Count() uint32 {
	__antithesis_instrumentation__.Notify(642621)
	return p.batch.Count()
}

func (p *pebbleBatch) Len() int {
	__antithesis_instrumentation__.Notify(642622)
	return len(p.batch.Repr())
}

func (p *pebbleBatch) Repr() []byte {
	__antithesis_instrumentation__.Notify(642623)

	repr := p.batch.Repr()
	reprCopy := make([]byte, len(repr))
	copy(reprCopy, repr)
	return reprCopy
}
