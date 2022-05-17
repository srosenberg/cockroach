package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type Deleter struct {
	Helper    rowHelper
	FetchCols []catalog.Column

	FetchColIDtoRowIndex catalog.TableColMap

	key roachpb.Key
}

func MakeDeleter(
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	requestedCols []catalog.Column,
	sv *settings.Values,
	internal bool,
	metrics *Metrics,
) Deleter {
	__antithesis_instrumentation__.Notify(567323)
	indexes := tableDesc.DeletableNonPrimaryIndexes()

	var fetchCols []catalog.Column
	var fetchColIDtoRowIndex catalog.TableColMap
	if requestedCols != nil {
		__antithesis_instrumentation__.Notify(567325)
		fetchCols = requestedCols[:len(requestedCols):len(requestedCols)]
		fetchColIDtoRowIndex = ColIDtoRowIndexFromCols(fetchCols)
	} else {
		__antithesis_instrumentation__.Notify(567326)
		maybeAddCol := func(colID descpb.ColumnID) error {
			__antithesis_instrumentation__.Notify(567329)
			if _, ok := fetchColIDtoRowIndex.Get(colID); !ok {
				__antithesis_instrumentation__.Notify(567331)
				col, err := tableDesc.FindColumnWithID(colID)
				if err != nil {
					__antithesis_instrumentation__.Notify(567333)
					return err
				} else {
					__antithesis_instrumentation__.Notify(567334)
				}
				__antithesis_instrumentation__.Notify(567332)
				fetchColIDtoRowIndex.Set(col.GetID(), len(fetchCols))
				fetchCols = append(fetchCols, col)
			} else {
				__antithesis_instrumentation__.Notify(567335)
			}
			__antithesis_instrumentation__.Notify(567330)
			return nil
		}
		__antithesis_instrumentation__.Notify(567327)
		for j := 0; j < tableDesc.GetPrimaryIndex().NumKeyColumns(); j++ {
			__antithesis_instrumentation__.Notify(567336)
			colID := tableDesc.GetPrimaryIndex().GetKeyColumnID(j)
			if err := maybeAddCol(colID); err != nil {
				__antithesis_instrumentation__.Notify(567337)
				return Deleter{}
			} else {
				__antithesis_instrumentation__.Notify(567338)
			}
		}
		__antithesis_instrumentation__.Notify(567328)
		for _, index := range indexes {
			__antithesis_instrumentation__.Notify(567339)
			for j := 0; j < index.NumKeyColumns(); j++ {
				__antithesis_instrumentation__.Notify(567341)
				colID := index.GetKeyColumnID(j)
				if err := maybeAddCol(colID); err != nil {
					__antithesis_instrumentation__.Notify(567342)
					return Deleter{}
				} else {
					__antithesis_instrumentation__.Notify(567343)
				}
			}
			__antithesis_instrumentation__.Notify(567340)

			for j := 0; j < index.NumKeySuffixColumns(); j++ {
				__antithesis_instrumentation__.Notify(567344)
				colID := index.GetKeySuffixColumnID(j)
				if err := maybeAddCol(colID); err != nil {
					__antithesis_instrumentation__.Notify(567345)
					return Deleter{}
				} else {
					__antithesis_instrumentation__.Notify(567346)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(567324)

	rd := Deleter{
		Helper:               newRowHelper(codec, tableDesc, indexes, sv, internal, metrics),
		FetchCols:            fetchCols,
		FetchColIDtoRowIndex: fetchColIDtoRowIndex,
	}

	return rd
}

func (rd *Deleter) DeleteRow(
	ctx context.Context, b *kv.Batch, values []tree.Datum, pm PartialIndexUpdateHelper, traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(567347)

	for i := range rd.Helper.Indexes {
		__antithesis_instrumentation__.Notify(567350)

		if pm.IgnoreForDel.Contains(int(rd.Helper.Indexes[i].GetID())) {
			__antithesis_instrumentation__.Notify(567353)
			continue
		} else {
			__antithesis_instrumentation__.Notify(567354)
		}
		__antithesis_instrumentation__.Notify(567351)

		entries, err := rowenc.EncodeSecondaryIndex(
			rd.Helper.Codec,
			rd.Helper.TableDesc,
			rd.Helper.Indexes[i],
			rd.FetchColIDtoRowIndex,
			values,
			true,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(567355)
			return err
		} else {
			__antithesis_instrumentation__.Notify(567356)
		}
		__antithesis_instrumentation__.Notify(567352)
		for _, e := range entries {
			__antithesis_instrumentation__.Notify(567357)
			if err := rd.Helper.deleteIndexEntry(ctx, b, rd.Helper.Indexes[i], rd.Helper.secIndexValDirs[i], &e, traceKV); err != nil {
				__antithesis_instrumentation__.Notify(567358)
				return err
			} else {
				__antithesis_instrumentation__.Notify(567359)
			}
		}
	}
	__antithesis_instrumentation__.Notify(567348)

	primaryIndexKey, err := rd.Helper.encodePrimaryIndex(rd.FetchColIDtoRowIndex, values)
	if err != nil {
		__antithesis_instrumentation__.Notify(567360)
		return err
	} else {
		__antithesis_instrumentation__.Notify(567361)
	}
	__antithesis_instrumentation__.Notify(567349)

	var called bool
	return rd.Helper.TableDesc.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
		__antithesis_instrumentation__.Notify(567362)
		if called {
			__antithesis_instrumentation__.Notify(567365)

			primaryIndexKey = primaryIndexKey[:len(primaryIndexKey):len(primaryIndexKey)]
		} else {
			__antithesis_instrumentation__.Notify(567366)
			called = true
		}
		__antithesis_instrumentation__.Notify(567363)
		familyID := family.ID
		rd.key = keys.MakeFamilyKey(primaryIndexKey, uint32(familyID))
		if traceKV {
			__antithesis_instrumentation__.Notify(567367)
			log.VEventf(ctx, 2, "Del %s", keys.PrettyPrint(rd.Helper.primIndexValDirs, rd.key))
		} else {
			__antithesis_instrumentation__.Notify(567368)
		}
		__antithesis_instrumentation__.Notify(567364)
		b.Del(&rd.key)
		rd.key = nil
		return nil
	})
}
