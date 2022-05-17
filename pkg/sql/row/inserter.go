package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type Inserter struct {
	Helper                rowHelper
	InsertCols            []catalog.Column
	InsertColIDtoRowIndex catalog.TableColMap

	marshaled []roachpb.Value
	key       roachpb.Key
	valueBuf  []byte
	value     roachpb.Value
}

func MakeInserter(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	insertCols []catalog.Column,
	alloc *tree.DatumAlloc,
	sv *settings.Values,
	internal bool,
	metrics *Metrics,
) (Inserter, error) {
	__antithesis_instrumentation__.Notify(568121)
	ri := Inserter{
		Helper: newRowHelper(
			codec, tableDesc, tableDesc.WritableNonPrimaryIndexes(), sv, internal, metrics,
		),

		InsertCols:            insertCols,
		InsertColIDtoRowIndex: ColIDtoRowIndexFromCols(insertCols),
		marshaled:             make([]roachpb.Value, len(insertCols)),
	}

	for i := 0; i < tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(568123)
		colID := tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		if _, ok := ri.InsertColIDtoRowIndex.Get(colID); !ok {
			__antithesis_instrumentation__.Notify(568124)
			return Inserter{}, fmt.Errorf("missing %q primary key column", tableDesc.GetPrimaryIndex().GetKeyColumnName(i))
		} else {
			__antithesis_instrumentation__.Notify(568125)
		}
	}
	__antithesis_instrumentation__.Notify(568122)

	return ri, nil
}

func insertCPutFn(
	ctx context.Context, b putter, key *roachpb.Key, value *roachpb.Value, traceKV bool,
) {
	__antithesis_instrumentation__.Notify(568126)

	if traceKV {
		__antithesis_instrumentation__.Notify(568128)
		log.VEventfDepth(ctx, 1, 2, "CPut %s -> %s", *key, value.PrettyPrint())
	} else {
		__antithesis_instrumentation__.Notify(568129)
	}
	__antithesis_instrumentation__.Notify(568127)
	b.CPut(key, value, nil)
}

func insertPutFn(
	ctx context.Context, b putter, key *roachpb.Key, value *roachpb.Value, traceKV bool,
) {
	__antithesis_instrumentation__.Notify(568130)
	if traceKV {
		__antithesis_instrumentation__.Notify(568132)
		log.VEventfDepth(ctx, 1, 2, "Put %s -> %s", *key, value.PrettyPrint())
	} else {
		__antithesis_instrumentation__.Notify(568133)
	}
	__antithesis_instrumentation__.Notify(568131)
	b.Put(key, value)
}

func insertDelFn(ctx context.Context, b putter, key *roachpb.Key, traceKV bool) {
	__antithesis_instrumentation__.Notify(568134)
	if traceKV {
		__antithesis_instrumentation__.Notify(568136)
		log.VEventfDepth(ctx, 1, 2, "Del %s", *key)
	} else {
		__antithesis_instrumentation__.Notify(568137)
	}
	__antithesis_instrumentation__.Notify(568135)
	b.Del(key)
}

func insertInvertedPutFn(
	ctx context.Context, b putter, key *roachpb.Key, value *roachpb.Value, traceKV bool,
) {
	__antithesis_instrumentation__.Notify(568138)
	if traceKV {
		__antithesis_instrumentation__.Notify(568140)
		log.VEventfDepth(ctx, 1, 2, "InitPut %s -> %s", *key, value.PrettyPrint())
	} else {
		__antithesis_instrumentation__.Notify(568141)
	}
	__antithesis_instrumentation__.Notify(568139)
	b.InitPut(key, value, false)
}

type putter interface {
	CPut(key, value interface{}, expValue []byte)
	Put(key, value interface{})
	InitPut(key, value interface{}, failOnTombstones bool)
	Del(key ...interface{})
}

func (ri *Inserter) InsertRow(
	ctx context.Context,
	b putter,
	values []tree.Datum,
	pm PartialIndexUpdateHelper,
	overwrite bool,
	traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(568142)
	if len(values) != len(ri.InsertCols) {
		__antithesis_instrumentation__.Notify(568149)
		return errors.Errorf("got %d values but expected %d", len(values), len(ri.InsertCols))
	} else {
		__antithesis_instrumentation__.Notify(568150)
	}
	__antithesis_instrumentation__.Notify(568143)

	putFn := insertCPutFn
	if overwrite {
		__antithesis_instrumentation__.Notify(568151)
		putFn = insertPutFn
	} else {
		__antithesis_instrumentation__.Notify(568152)
	}
	__antithesis_instrumentation__.Notify(568144)

	for i, val := range values {
		__antithesis_instrumentation__.Notify(568153)

		var err error
		if ri.marshaled[i], err = valueside.MarshalLegacy(ri.InsertCols[i].GetType(), val); err != nil {
			__antithesis_instrumentation__.Notify(568154)
			return err
		} else {
			__antithesis_instrumentation__.Notify(568155)
		}
	}
	__antithesis_instrumentation__.Notify(568145)

	primaryIndexKey, secondaryIndexEntries, err := ri.Helper.encodeIndexes(
		ri.InsertColIDtoRowIndex, values, pm.IgnoreForPut, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(568156)
		return err
	} else {
		__antithesis_instrumentation__.Notify(568157)
	}
	__antithesis_instrumentation__.Notify(568146)

	ri.valueBuf, err = prepareInsertOrUpdateBatch(ctx, b,
		&ri.Helper, primaryIndexKey, ri.InsertCols,
		values, ri.InsertColIDtoRowIndex,
		ri.marshaled, ri.InsertColIDtoRowIndex,
		&ri.key, &ri.value, ri.valueBuf, putFn, overwrite, traceKV)
	if err != nil {
		__antithesis_instrumentation__.Notify(568158)
		return err
	} else {
		__antithesis_instrumentation__.Notify(568159)
	}
	__antithesis_instrumentation__.Notify(568147)

	putFn = insertInvertedPutFn

	for idx := range ri.Helper.Indexes {
		__antithesis_instrumentation__.Notify(568160)
		entries, ok := secondaryIndexEntries[ri.Helper.Indexes[idx]]
		if ok {
			__antithesis_instrumentation__.Notify(568161)
			for i := range entries {
				__antithesis_instrumentation__.Notify(568162)
				e := &entries[i]

				if ri.Helper.Indexes[idx].ForcePut() {
					__antithesis_instrumentation__.Notify(568163)

					insertPutFn(ctx, b, &e.Key, &e.Value, traceKV)
				} else {
					__antithesis_instrumentation__.Notify(568164)
					putFn(ctx, b, &e.Key, &e.Value, traceKV)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(568165)
		}
	}
	__antithesis_instrumentation__.Notify(568148)

	return nil
}
