package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

func ColIDtoRowIndexFromCols(cols []catalog.Column) catalog.TableColMap {
	__antithesis_instrumentation__.Notify(568836)
	var colIDtoRowIndex catalog.TableColMap
	for i := range cols {
		__antithesis_instrumentation__.Notify(568838)
		colIDtoRowIndex.Set(cols[i].GetID(), i)
	}
	__antithesis_instrumentation__.Notify(568837)
	return colIDtoRowIndex
}

func ColMapping(fromCols, toCols []catalog.Column) []int {
	__antithesis_instrumentation__.Notify(568839)

	var colMap util.FastIntMap
	for i := range fromCols {
		__antithesis_instrumentation__.Notify(568843)
		colMap.Set(int(fromCols[i].GetID()), i)
	}
	__antithesis_instrumentation__.Notify(568840)

	result := make([]int, len(fromCols))
	for i := range result {
		__antithesis_instrumentation__.Notify(568844)

		result[i] = -1
	}
	__antithesis_instrumentation__.Notify(568841)

	for toOrd := range toCols {
		__antithesis_instrumentation__.Notify(568845)
		if fromOrd, ok := colMap.Get(int(toCols[toOrd].GetID())); ok {
			__antithesis_instrumentation__.Notify(568846)
			result[fromOrd] = toOrd
		} else {
			__antithesis_instrumentation__.Notify(568847)
		}
	}
	__antithesis_instrumentation__.Notify(568842)

	return result
}

func prepareInsertOrUpdateBatch(
	ctx context.Context,
	batch putter,
	helper *rowHelper,
	primaryIndexKey []byte,
	fetchedCols []catalog.Column,
	values []tree.Datum,
	valColIDMapping catalog.TableColMap,
	marshaledValues []roachpb.Value,
	marshaledColIDMapping catalog.TableColMap,
	kvKey *roachpb.Key,
	kvValue *roachpb.Value,
	rawValueBuf []byte,
	putFn func(ctx context.Context, b putter, key *roachpb.Key, value *roachpb.Value, traceKV bool),
	overwrite, traceKV bool,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(568848)
	families := helper.TableDesc.GetFamilies()
	for i := range families {
		__antithesis_instrumentation__.Notify(568850)
		family := &families[i]
		update := false
		for _, colID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(568858)
			if _, ok := marshaledColIDMapping.Get(colID); ok {
				__antithesis_instrumentation__.Notify(568859)
				update = true
				break
			} else {
				__antithesis_instrumentation__.Notify(568860)
			}
		}
		__antithesis_instrumentation__.Notify(568851)

		if !update && func() bool {
			__antithesis_instrumentation__.Notify(568861)
			return len(family.ColumnIDs) != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(568862)
			continue
		} else {
			__antithesis_instrumentation__.Notify(568863)
		}
		__antithesis_instrumentation__.Notify(568852)

		if i > 0 {
			__antithesis_instrumentation__.Notify(568864)

			primaryIndexKey = primaryIndexKey[:len(primaryIndexKey):len(primaryIndexKey)]
		} else {
			__antithesis_instrumentation__.Notify(568865)
		}
		__antithesis_instrumentation__.Notify(568853)

		*kvKey = keys.MakeFamilyKey(primaryIndexKey, uint32(family.ID))

		if len(family.ColumnIDs) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(568866)
			return family.ColumnIDs[0] == family.DefaultColumnID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(568867)
			return family.ID != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(568868)

			idx, ok := marshaledColIDMapping.Get(family.DefaultColumnID)
			if !ok {
				__antithesis_instrumentation__.Notify(568871)
				continue
			} else {
				__antithesis_instrumentation__.Notify(568872)
			}
			__antithesis_instrumentation__.Notify(568869)

			if marshaledValues[idx].RawBytes == nil {
				__antithesis_instrumentation__.Notify(568873)
				if overwrite {
					__antithesis_instrumentation__.Notify(568874)

					insertDelFn(ctx, batch, kvKey, traceKV)
				} else {
					__antithesis_instrumentation__.Notify(568875)
				}
			} else {
				__antithesis_instrumentation__.Notify(568876)

				if err := helper.checkRowSize(ctx, kvKey, &marshaledValues[idx], family.ID); err != nil {
					__antithesis_instrumentation__.Notify(568878)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(568879)
				}
				__antithesis_instrumentation__.Notify(568877)
				putFn(ctx, batch, kvKey, &marshaledValues[idx], traceKV)
			}
			__antithesis_instrumentation__.Notify(568870)

			continue
		} else {
			__antithesis_instrumentation__.Notify(568880)
		}
		__antithesis_instrumentation__.Notify(568854)

		rawValueBuf = rawValueBuf[:0]

		var lastColID descpb.ColumnID
		familySortedColumnIDs, ok := helper.sortedColumnFamily(family.ID)
		if !ok {
			__antithesis_instrumentation__.Notify(568881)
			return nil, errors.AssertionFailedf("invalid family sorted column id map")
		} else {
			__antithesis_instrumentation__.Notify(568882)
		}
		__antithesis_instrumentation__.Notify(568855)
		for _, colID := range familySortedColumnIDs {
			__antithesis_instrumentation__.Notify(568883)
			idx, ok := valColIDMapping.Get(colID)
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(568887)
				return values[idx] == tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(568888)

				continue
			} else {
				__antithesis_instrumentation__.Notify(568889)
			}
			__antithesis_instrumentation__.Notify(568884)

			if skip, err := helper.skipColumnNotInPrimaryIndexValue(colID, values[idx]); err != nil {
				__antithesis_instrumentation__.Notify(568890)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568891)
				if skip {
					__antithesis_instrumentation__.Notify(568892)
					continue
				} else {
					__antithesis_instrumentation__.Notify(568893)
				}
			}
			__antithesis_instrumentation__.Notify(568885)

			col := fetchedCols[idx]
			if lastColID > col.GetID() {
				__antithesis_instrumentation__.Notify(568894)
				return nil, errors.AssertionFailedf("cannot write column id %d after %d", col.GetID(), lastColID)
			} else {
				__antithesis_instrumentation__.Notify(568895)
			}
			__antithesis_instrumentation__.Notify(568886)
			colIDDelta := valueside.MakeColumnIDDelta(lastColID, col.GetID())
			lastColID = col.GetID()
			var err error
			rawValueBuf, err = valueside.Encode(rawValueBuf, colIDDelta, values[idx], nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(568896)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568897)
			}
		}
		__antithesis_instrumentation__.Notify(568856)

		if family.ID != 0 && func() bool {
			__antithesis_instrumentation__.Notify(568898)
			return len(rawValueBuf) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(568899)
			if overwrite {
				__antithesis_instrumentation__.Notify(568900)

				insertDelFn(ctx, batch, kvKey, traceKV)
			} else {
				__antithesis_instrumentation__.Notify(568901)
			}
		} else {
			__antithesis_instrumentation__.Notify(568902)

			kvValue.SetTuple(rawValueBuf)
			if err := helper.checkRowSize(ctx, kvKey, kvValue, family.ID); err != nil {
				__antithesis_instrumentation__.Notify(568904)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568905)
			}
			__antithesis_instrumentation__.Notify(568903)
			putFn(ctx, batch, kvKey, kvValue, traceKV)
		}
		__antithesis_instrumentation__.Notify(568857)

		*kvKey = nil

		*kvValue = roachpb.Value{}
	}
	__antithesis_instrumentation__.Notify(568849)

	return rawValueBuf, nil
}
