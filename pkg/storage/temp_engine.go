package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/diskmap"
	"github.com/cockroachdb/cockroach/pkg/storage/fs"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/pebble"
)

func NewTempEngine(
	ctx context.Context, tempStorage base.TempStorageConfig, storeSpec base.StoreSpec,
) (diskmap.Factory, fs.FS, error) {
	__antithesis_instrumentation__.Notify(643954)
	return NewPebbleTempEngine(ctx, tempStorage, storeSpec)
}

type pebbleTempEngine struct {
	db *pebble.DB
}

func (r *pebbleTempEngine) Close() {
	__antithesis_instrumentation__.Notify(643955)
	err := r.db.Close()
	if err != nil {
		__antithesis_instrumentation__.Notify(643956)
		log.Fatalf(context.TODO(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(643957)
	}
}

func (r *pebbleTempEngine) NewSortedDiskMap() diskmap.SortedDiskMap {
	__antithesis_instrumentation__.Notify(643958)
	return newPebbleMap(r.db, false)
}

func (r *pebbleTempEngine) NewSortedDiskMultiMap() diskmap.SortedDiskMap {
	__antithesis_instrumentation__.Notify(643959)
	return newPebbleMap(r.db, true)
}

func NewPebbleTempEngine(
	ctx context.Context, tempStorage base.TempStorageConfig, storeSpec base.StoreSpec,
) (diskmap.Factory, fs.FS, error) {
	__antithesis_instrumentation__.Notify(643960)
	return newPebbleTempEngine(ctx, tempStorage, storeSpec)
}

func newPebbleTempEngine(
	ctx context.Context, tempStorage base.TempStorageConfig, storeSpec base.StoreSpec,
) (*pebbleTempEngine, fs.FS, error) {
	__antithesis_instrumentation__.Notify(643961)
	var loc Location
	var cacheSize int64 = 128 << 20
	if tempStorage.InMemory {
		__antithesis_instrumentation__.Notify(643965)
		cacheSize = 8 << 20
		loc = InMemory()
	} else {
		__antithesis_instrumentation__.Notify(643966)
		loc = Filesystem(tempStorage.Path)
	}
	__antithesis_instrumentation__.Notify(643962)

	p, err := Open(ctx, loc,
		CacheSize(cacheSize),
		func(cfg *engineConfig) error {
			__antithesis_instrumentation__.Notify(643967)
			cfg.UseFileRegistry = storeSpec.UseFileRegistry
			cfg.EncryptionOptions = storeSpec.EncryptionOptions

			cfg.Opts.Comparer = pebble.DefaultComparer
			cfg.Opts.DisableWAL = true
			cfg.Opts.TablePropertyCollectors = nil
			cfg.Opts.Experimental.KeyValidationFunc = nil
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(643963)
	if err != nil {
		__antithesis_instrumentation__.Notify(643968)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(643969)
	}
	__antithesis_instrumentation__.Notify(643964)

	p.SetStoreID(ctx, base.TempStoreID)

	return &pebbleTempEngine{db: p.db}, p, nil
}
