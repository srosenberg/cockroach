package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

const defaultBatchCapacityBytes = 4096

var tempStorageID uint64

func generateTempStorageID() uint64 {
	__antithesis_instrumentation__.Notify(633602)
	return atomic.AddUint64(&tempStorageID, 1)
}

type pebbleMapBatchWriter struct {
	capacity int

	makeKey           func(k []byte) []byte
	batch             *pebble.Batch
	numPutsSinceFlush int
	store             *pebble.DB
}

type pebbleMapIterator struct {
	allowDuplicates bool
	iter            *pebble.Iterator

	makeKey func(k []byte) []byte

	prefix []byte
}

type pebbleMap struct {
	prefix          []byte
	store           *pebble.DB
	allowDuplicates bool
	keyID           int64
}

var _ diskmap.SortedDiskMapBatchWriter = &pebbleMapBatchWriter{}
var _ diskmap.SortedDiskMapIterator = &pebbleMapIterator{}
var _ diskmap.SortedDiskMap = &pebbleMap{}

func newPebbleMap(e *pebble.DB, allowDuplicates bool) *pebbleMap {
	__antithesis_instrumentation__.Notify(633603)
	prefix := generateTempStorageID()
	return &pebbleMap{
		prefix:          encoding.EncodeUvarintAscending([]byte(nil), prefix),
		store:           e,
		allowDuplicates: allowDuplicates,
	}
}

func (r *pebbleMap) makeKey(k []byte) []byte {
	__antithesis_instrumentation__.Notify(633604)
	prefixLen := len(r.prefix)
	r.prefix = append(r.prefix, k...)
	key := r.prefix
	r.prefix = r.prefix[:prefixLen]
	return key
}

func (r *pebbleMap) makeKeyWithSequence(k []byte) []byte {
	__antithesis_instrumentation__.Notify(633605)
	byteKey := r.makeKey(k)
	if r.allowDuplicates {
		__antithesis_instrumentation__.Notify(633607)
		r.keyID++
		byteKey = encoding.EncodeUint64Ascending(byteKey, uint64(r.keyID))
	} else {
		__antithesis_instrumentation__.Notify(633608)
	}
	__antithesis_instrumentation__.Notify(633606)
	return byteKey
}

func (r *pebbleMap) NewIterator() diskmap.SortedDiskMapIterator {
	__antithesis_instrumentation__.Notify(633609)
	return &pebbleMapIterator{
		allowDuplicates: r.allowDuplicates,
		iter: r.store.NewIter(&pebble.IterOptions{
			UpperBound: roachpb.Key(r.prefix).PrefixEnd(),
		}),
		makeKey: r.makeKey,
		prefix:  r.prefix,
	}
}

func (r *pebbleMap) NewBatchWriter() diskmap.SortedDiskMapBatchWriter {
	__antithesis_instrumentation__.Notify(633610)
	return r.NewBatchWriterCapacity(defaultBatchCapacityBytes)
}

func (r *pebbleMap) NewBatchWriterCapacity(capacityBytes int) diskmap.SortedDiskMapBatchWriter {
	__antithesis_instrumentation__.Notify(633611)
	makeKey := r.makeKey
	if r.allowDuplicates {
		__antithesis_instrumentation__.Notify(633613)
		makeKey = r.makeKeyWithSequence
	} else {
		__antithesis_instrumentation__.Notify(633614)
	}
	__antithesis_instrumentation__.Notify(633612)
	return &pebbleMapBatchWriter{
		capacity: capacityBytes,
		makeKey:  makeKey,
		batch:    r.store.NewBatch(),
		store:    r.store,
	}
}

func (r *pebbleMap) Clear() error {
	__antithesis_instrumentation__.Notify(633615)
	if err := r.store.DeleteRange(
		r.prefix,
		roachpb.Key(r.prefix).PrefixEnd(),
		pebble.NoSync,
	); err != nil {
		__antithesis_instrumentation__.Notify(633617)
		return errors.Wrapf(err, "unable to clear range with prefix %v", r.prefix)
	} else {
		__antithesis_instrumentation__.Notify(633618)
	}
	__antithesis_instrumentation__.Notify(633616)

	_, err := r.store.AsyncFlush()
	return err
}

func (r *pebbleMap) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(633619)
	if err := r.Clear(); err != nil {
		__antithesis_instrumentation__.Notify(633620)
		log.Errorf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(633621)
	}
}

func (i *pebbleMapIterator) SeekGE(k []byte) {
	__antithesis_instrumentation__.Notify(633622)
	i.iter.SeekGE(i.makeKey(k))
}

func (i *pebbleMapIterator) Rewind() {
	__antithesis_instrumentation__.Notify(633623)
	i.iter.SeekGE(i.makeKey(nil))
}

func (i *pebbleMapIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(633624)
	return i.iter.Valid(), nil
}

func (i *pebbleMapIterator) Next() {
	__antithesis_instrumentation__.Notify(633625)
	i.iter.Next()
}

func (i *pebbleMapIterator) UnsafeKey() []byte {
	__antithesis_instrumentation__.Notify(633626)
	unsafeKey := i.iter.Key()
	end := len(unsafeKey)
	if i.allowDuplicates {
		__antithesis_instrumentation__.Notify(633628)

		end -= 8
	} else {
		__antithesis_instrumentation__.Notify(633629)
	}
	__antithesis_instrumentation__.Notify(633627)
	return unsafeKey[len(i.prefix):end]
}

func (i *pebbleMapIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(633630)
	return i.iter.Value()
}

func (i *pebbleMapIterator) Close() {
	__antithesis_instrumentation__.Notify(633631)
	_ = i.iter.Close()
}

func (b *pebbleMapBatchWriter) Put(k []byte, v []byte) error {
	__antithesis_instrumentation__.Notify(633632)
	key := b.makeKey(k)
	if err := b.batch.Set(key, v, nil); err != nil {
		__antithesis_instrumentation__.Notify(633635)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633636)
	}
	__antithesis_instrumentation__.Notify(633633)
	b.numPutsSinceFlush++
	if len(b.batch.Repr()) >= b.capacity {
		__antithesis_instrumentation__.Notify(633637)
		return b.Flush()
	} else {
		__antithesis_instrumentation__.Notify(633638)
	}
	__antithesis_instrumentation__.Notify(633634)
	return nil
}

func (b *pebbleMapBatchWriter) Flush() error {
	__antithesis_instrumentation__.Notify(633639)
	if err := b.batch.Commit(pebble.NoSync); err != nil {
		__antithesis_instrumentation__.Notify(633641)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633642)
	}
	__antithesis_instrumentation__.Notify(633640)
	b.numPutsSinceFlush = 0
	b.batch = b.store.NewBatch()
	return nil
}

func (b *pebbleMapBatchWriter) NumPutsSinceFlush() int {
	__antithesis_instrumentation__.Notify(633643)
	return b.numPutsSinceFlush
}

func (b *pebbleMapBatchWriter) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(633644)
	err := b.Flush()
	if err != nil {
		__antithesis_instrumentation__.Notify(633646)
		return err
	} else {
		__antithesis_instrumentation__.Notify(633647)
	}
	__antithesis_instrumentation__.Notify(633645)
	return b.batch.Close()
}
