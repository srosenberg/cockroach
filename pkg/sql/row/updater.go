package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/unique"
	"github.com/cockroachdb/errors"
)

type Updater struct {
	Helper       rowHelper
	DeleteHelper *rowHelper
	FetchCols    []catalog.Column

	FetchColIDtoRowIndex  catalog.TableColMap
	UpdateCols            []catalog.Column
	UpdateColIDtoRowIndex catalog.TableColMap
	primaryKeyColChange   bool

	rd Deleter
	ri Inserter

	marshaled       []roachpb.Value
	newValues       []tree.Datum
	key             roachpb.Key
	valueBuf        []byte
	value           roachpb.Value
	oldIndexEntries [][]rowenc.IndexEntry
	newIndexEntries [][]rowenc.IndexEntry
}

type rowUpdaterType int

const (
	UpdaterDefault rowUpdaterType = 0

	UpdaterOnlyColumns rowUpdaterType = 1
)

func MakeUpdater(
	ctx context.Context,
	txn *kv.Txn,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	updateCols []catalog.Column,
	requestedCols []catalog.Column,
	updateType rowUpdaterType,
	alloc *tree.DatumAlloc,
	sv *settings.Values,
	internal bool,
	metrics *Metrics,
) (Updater, error) {
	__antithesis_instrumentation__.Notify(568651)
	if requestedCols == nil {
		__antithesis_instrumentation__.Notify(568659)
		return Updater{}, errors.AssertionFailedf("requestedCols is nil in MakeUpdater")
	} else {
		__antithesis_instrumentation__.Notify(568660)
	}
	__antithesis_instrumentation__.Notify(568652)

	updateColIDtoRowIndex := ColIDtoRowIndexFromCols(updateCols)

	var primaryIndexCols catalog.TableColSet
	for i := 0; i < tableDesc.GetPrimaryIndex().NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(568661)
		colID := tableDesc.GetPrimaryIndex().GetKeyColumnID(i)
		primaryIndexCols.Add(colID)
	}
	__antithesis_instrumentation__.Notify(568653)

	var primaryKeyColChange bool
	for _, c := range updateCols {
		__antithesis_instrumentation__.Notify(568662)
		if primaryIndexCols.Contains(c.GetID()) {
			__antithesis_instrumentation__.Notify(568663)
			primaryKeyColChange = true
			break
		} else {
			__antithesis_instrumentation__.Notify(568664)
		}
	}
	__antithesis_instrumentation__.Notify(568654)

	needsUpdate := func(index catalog.Index) bool {
		__antithesis_instrumentation__.Notify(568665)

		if updateType == UpdaterOnlyColumns {
			__antithesis_instrumentation__.Notify(568670)
			return false
		} else {
			__antithesis_instrumentation__.Notify(568671)
		}
		__antithesis_instrumentation__.Notify(568666)

		if primaryKeyColChange {
			__antithesis_instrumentation__.Notify(568672)
			return true
		} else {
			__antithesis_instrumentation__.Notify(568673)
		}
		__antithesis_instrumentation__.Notify(568667)

		if index.IsPartial() {
			__antithesis_instrumentation__.Notify(568674)
			return true
		} else {
			__antithesis_instrumentation__.Notify(568675)
		}
		__antithesis_instrumentation__.Notify(568668)
		colIDs := index.CollectKeyColumnIDs()
		colIDs.UnionWith(index.CollectSecondaryStoredColumnIDs())
		colIDs.UnionWith(index.CollectKeySuffixColumnIDs())
		for _, colID := range colIDs.Ordered() {
			__antithesis_instrumentation__.Notify(568676)
			if _, ok := updateColIDtoRowIndex.Get(colID); ok {
				__antithesis_instrumentation__.Notify(568677)
				return true
			} else {
				__antithesis_instrumentation__.Notify(568678)
			}
		}
		__antithesis_instrumentation__.Notify(568669)
		return false
	}
	__antithesis_instrumentation__.Notify(568655)

	includeIndexes := make([]catalog.Index, 0, len(tableDesc.WritableNonPrimaryIndexes()))
	var deleteOnlyIndexes []catalog.Index
	for _, index := range tableDesc.DeletableNonPrimaryIndexes() {
		__antithesis_instrumentation__.Notify(568679)
		if !needsUpdate(index) {
			__antithesis_instrumentation__.Notify(568681)
			continue
		} else {
			__antithesis_instrumentation__.Notify(568682)
		}
		__antithesis_instrumentation__.Notify(568680)
		if !index.DeleteOnly() {
			__antithesis_instrumentation__.Notify(568683)
			includeIndexes = append(includeIndexes, index)
		} else {
			__antithesis_instrumentation__.Notify(568684)
			if deleteOnlyIndexes == nil {
				__antithesis_instrumentation__.Notify(568686)

				deleteOnlyIndexes = make([]catalog.Index, 0, len(tableDesc.DeleteOnlyNonPrimaryIndexes()))
			} else {
				__antithesis_instrumentation__.Notify(568687)
			}
			__antithesis_instrumentation__.Notify(568685)
			deleteOnlyIndexes = append(deleteOnlyIndexes, index)
		}
	}
	__antithesis_instrumentation__.Notify(568656)

	var deleteOnlyHelper *rowHelper
	if len(deleteOnlyIndexes) > 0 {
		__antithesis_instrumentation__.Notify(568688)
		rh := newRowHelper(codec, tableDesc, deleteOnlyIndexes, sv, internal, metrics)
		deleteOnlyHelper = &rh
	} else {
		__antithesis_instrumentation__.Notify(568689)
	}
	__antithesis_instrumentation__.Notify(568657)

	ru := Updater{
		Helper:                newRowHelper(codec, tableDesc, includeIndexes, sv, internal, metrics),
		DeleteHelper:          deleteOnlyHelper,
		FetchCols:             requestedCols,
		FetchColIDtoRowIndex:  ColIDtoRowIndexFromCols(requestedCols),
		UpdateCols:            updateCols,
		UpdateColIDtoRowIndex: updateColIDtoRowIndex,
		primaryKeyColChange:   primaryKeyColChange,
		marshaled:             make([]roachpb.Value, len(updateCols)),
		oldIndexEntries:       make([][]rowenc.IndexEntry, len(includeIndexes)),
		newIndexEntries:       make([][]rowenc.IndexEntry, len(includeIndexes)),
	}

	if primaryKeyColChange {
		__antithesis_instrumentation__.Notify(568690)

		var err error
		ru.rd = MakeDeleter(codec, tableDesc, requestedCols, sv, internal, metrics)
		if ru.ri, err = MakeInserter(
			ctx, txn, codec, tableDesc, requestedCols, alloc, sv, internal, metrics,
		); err != nil {
			__antithesis_instrumentation__.Notify(568691)
			return Updater{}, err
		} else {
			__antithesis_instrumentation__.Notify(568692)
		}
	} else {
		__antithesis_instrumentation__.Notify(568693)
	}
	__antithesis_instrumentation__.Notify(568658)

	ru.newValues = make(tree.Datums, len(ru.FetchCols))

	return ru, nil
}

func (ru *Updater) UpdateRow(
	ctx context.Context,
	batch *kv.Batch,
	oldValues []tree.Datum,
	updateValues []tree.Datum,
	pm PartialIndexUpdateHelper,
	traceKV bool,
) ([]tree.Datum, error) {
	__antithesis_instrumentation__.Notify(568694)
	if len(oldValues) != len(ru.FetchCols) {
		__antithesis_instrumentation__.Notify(568707)
		return nil, errors.Errorf("got %d values but expected %d", len(oldValues), len(ru.FetchCols))
	} else {
		__antithesis_instrumentation__.Notify(568708)
	}
	__antithesis_instrumentation__.Notify(568695)
	if len(updateValues) != len(ru.UpdateCols) {
		__antithesis_instrumentation__.Notify(568709)
		return nil, errors.Errorf("got %d values but expected %d", len(updateValues), len(ru.UpdateCols))
	} else {
		__antithesis_instrumentation__.Notify(568710)
	}
	__antithesis_instrumentation__.Notify(568696)

	primaryIndexKey, err := ru.Helper.encodePrimaryIndex(ru.FetchColIDtoRowIndex, oldValues)
	if err != nil {
		__antithesis_instrumentation__.Notify(568711)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(568712)
	}
	__antithesis_instrumentation__.Notify(568697)
	var deleteOldSecondaryIndexEntries map[catalog.Index][]rowenc.IndexEntry
	if ru.DeleteHelper != nil {
		__antithesis_instrumentation__.Notify(568713)

		_, deleteOldSecondaryIndexEntries, err = ru.DeleteHelper.encodeIndexes(
			ru.FetchColIDtoRowIndex, oldValues, pm.IgnoreForDel, true)
		if err != nil {
			__antithesis_instrumentation__.Notify(568714)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568715)
		}
	} else {
		__antithesis_instrumentation__.Notify(568716)
	}
	__antithesis_instrumentation__.Notify(568698)

	for i, val := range updateValues {
		__antithesis_instrumentation__.Notify(568717)
		if ru.marshaled[i], err = valueside.MarshalLegacy(ru.UpdateCols[i].GetType(), val); err != nil {
			__antithesis_instrumentation__.Notify(568718)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568719)
		}
	}
	__antithesis_instrumentation__.Notify(568699)

	copy(ru.newValues, oldValues)
	for i, updateCol := range ru.UpdateCols {
		__antithesis_instrumentation__.Notify(568720)
		idx, ok := ru.FetchColIDtoRowIndex.Get(updateCol.GetID())
		if !ok {
			__antithesis_instrumentation__.Notify(568722)
			return nil, errors.AssertionFailedf("update column without a corresponding fetch column")
		} else {
			__antithesis_instrumentation__.Notify(568723)
		}
		__antithesis_instrumentation__.Notify(568721)
		ru.newValues[idx] = updateValues[i]
	}
	__antithesis_instrumentation__.Notify(568700)

	rowPrimaryKeyChanged := false
	if ru.primaryKeyColChange {
		__antithesis_instrumentation__.Notify(568724)
		var newPrimaryIndexKey []byte
		newPrimaryIndexKey, err =
			ru.Helper.encodePrimaryIndex(ru.FetchColIDtoRowIndex, ru.newValues)
		if err != nil {
			__antithesis_instrumentation__.Notify(568726)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568727)
		}
		__antithesis_instrumentation__.Notify(568725)
		rowPrimaryKeyChanged = !bytes.Equal(primaryIndexKey, newPrimaryIndexKey)
	} else {
		__antithesis_instrumentation__.Notify(568728)
	}
	__antithesis_instrumentation__.Notify(568701)

	for i, index := range ru.Helper.Indexes {
		__antithesis_instrumentation__.Notify(568729)

		if pm.IgnoreForDel.Contains(int(index.GetID())) {
			__antithesis_instrumentation__.Notify(568732)
			ru.oldIndexEntries[i] = nil
		} else {
			__antithesis_instrumentation__.Notify(568733)
			ru.oldIndexEntries[i], err = rowenc.EncodeSecondaryIndex(
				ru.Helper.Codec,
				ru.Helper.TableDesc,
				index,
				ru.FetchColIDtoRowIndex,
				oldValues,
				false,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(568734)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568735)
			}
		}
		__antithesis_instrumentation__.Notify(568730)
		if pm.IgnoreForPut.Contains(int(index.GetID())) {
			__antithesis_instrumentation__.Notify(568736)
			ru.newIndexEntries[i] = nil
		} else {
			__antithesis_instrumentation__.Notify(568737)
			ru.newIndexEntries[i], err = rowenc.EncodeSecondaryIndex(
				ru.Helper.Codec,
				ru.Helper.TableDesc,
				index,
				ru.FetchColIDtoRowIndex,
				ru.newValues,
				false,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(568738)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568739)
			}
		}
		__antithesis_instrumentation__.Notify(568731)
		if ru.Helper.Indexes[i].GetType() == descpb.IndexDescriptor_INVERTED && func() bool {
			__antithesis_instrumentation__.Notify(568740)
			return !ru.Helper.Indexes[i].IsTemporaryIndexForBackfill() == true
		}() == true {
			__antithesis_instrumentation__.Notify(568741)

			newIndexEntries := ru.newIndexEntries[i]
			oldIndexEntries := ru.oldIndexEntries[i]
			sort.Slice(oldIndexEntries, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(568745)
				return compareIndexEntries(oldIndexEntries[i], oldIndexEntries[j]) < 0
			})
			__antithesis_instrumentation__.Notify(568742)
			sort.Slice(newIndexEntries, func(i, j int) bool {
				__antithesis_instrumentation__.Notify(568746)
				return compareIndexEntries(newIndexEntries[i], newIndexEntries[j]) < 0
			})
			__antithesis_instrumentation__.Notify(568743)
			oldLen, newLen := unique.UniquifyAcrossSlices(
				oldIndexEntries, newIndexEntries,
				func(l, r int) int {
					__antithesis_instrumentation__.Notify(568747)
					return compareIndexEntries(oldIndexEntries[l], newIndexEntries[r])
				},
				func(i, j int) {
					__antithesis_instrumentation__.Notify(568748)
					oldIndexEntries[i] = oldIndexEntries[j]
				},
				func(i, j int) {
					__antithesis_instrumentation__.Notify(568749)
					newIndexEntries[i] = newIndexEntries[j]
				})
			__antithesis_instrumentation__.Notify(568744)
			ru.oldIndexEntries[i] = oldIndexEntries[:oldLen]
			ru.newIndexEntries[i] = newIndexEntries[:newLen]
		} else {
			__antithesis_instrumentation__.Notify(568750)
		}
	}
	__antithesis_instrumentation__.Notify(568702)

	if rowPrimaryKeyChanged {
		__antithesis_instrumentation__.Notify(568751)
		if err := ru.rd.DeleteRow(ctx, batch, oldValues, pm, traceKV); err != nil {
			__antithesis_instrumentation__.Notify(568754)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568755)
		}
		__antithesis_instrumentation__.Notify(568752)
		if err := ru.ri.InsertRow(
			ctx, batch, ru.newValues, pm, false, traceKV,
		); err != nil {
			__antithesis_instrumentation__.Notify(568756)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(568757)
		}
		__antithesis_instrumentation__.Notify(568753)

		return ru.newValues, nil
	} else {
		__antithesis_instrumentation__.Notify(568758)
	}
	__antithesis_instrumentation__.Notify(568703)

	ru.valueBuf, err = prepareInsertOrUpdateBatch(ctx, batch,
		&ru.Helper, primaryIndexKey, ru.FetchCols,
		ru.newValues, ru.FetchColIDtoRowIndex,
		ru.marshaled, ru.UpdateColIDtoRowIndex,
		&ru.key, &ru.value, ru.valueBuf, insertPutFn, true, traceKV)
	if err != nil {
		__antithesis_instrumentation__.Notify(568759)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(568760)
	}
	__antithesis_instrumentation__.Notify(568704)

	for i, index := range ru.Helper.Indexes {
		__antithesis_instrumentation__.Notify(568761)
		if index.GetType() == descpb.IndexDescriptor_FORWARD {
			__antithesis_instrumentation__.Notify(568762)
			oldIdx, newIdx := 0, 0
			oldEntries, newEntries := ru.oldIndexEntries[i], ru.newIndexEntries[i]

			for oldIdx < len(oldEntries) && func() bool {
				__antithesis_instrumentation__.Notify(568765)
				return newIdx < len(newEntries) == true
			}() == true {
				__antithesis_instrumentation__.Notify(568766)
				oldEntry, newEntry := &oldEntries[oldIdx], &newEntries[newIdx]
				if oldEntry.Family == newEntry.Family {
					__antithesis_instrumentation__.Notify(568767)

					oldIdx++
					newIdx++
					var expValue []byte
					if !bytes.Equal(oldEntry.Key, newEntry.Key) {
						__antithesis_instrumentation__.Notify(568769)
						if err := ru.Helper.deleteIndexEntry(ctx, batch, index, ru.Helper.secIndexValDirs[i], oldEntry, traceKV); err != nil {
							__antithesis_instrumentation__.Notify(568770)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(568771)
						}
					} else {
						__antithesis_instrumentation__.Notify(568772)
						if !newEntry.Value.EqualTagAndData(oldEntry.Value) {
							__antithesis_instrumentation__.Notify(568773)
							expValue = oldEntry.Value.TagAndDataBytes()
						} else {
							__antithesis_instrumentation__.Notify(568774)
							if !index.IsTemporaryIndexForBackfill() {
								__antithesis_instrumentation__.Notify(568775)

								continue
							} else {
								__antithesis_instrumentation__.Notify(568776)
							}
						}
					}
					__antithesis_instrumentation__.Notify(568768)

					if index.ForcePut() {
						__antithesis_instrumentation__.Notify(568777)

						insertPutFn(ctx, batch, &newEntry.Key, &newEntry.Value, traceKV)
					} else {
						__antithesis_instrumentation__.Notify(568778)
						if traceKV {
							__antithesis_instrumentation__.Notify(568780)
							k := keys.PrettyPrint(ru.Helper.secIndexValDirs[i], newEntry.Key)
							v := newEntry.Value.PrettyPrint()
							if expValue != nil {
								__antithesis_instrumentation__.Notify(568781)
								log.VEventf(ctx, 2, "CPut %s -> %v (replacing %v, if exists)", k, v, expValue)
							} else {
								__antithesis_instrumentation__.Notify(568782)
								log.VEventf(ctx, 2, "CPut %s -> %v (expecting does not exist)", k, v)
							}
						} else {
							__antithesis_instrumentation__.Notify(568783)
						}
						__antithesis_instrumentation__.Notify(568779)
						batch.CPutAllowingIfNotExists(newEntry.Key, &newEntry.Value, expValue)
					}
				} else {
					__antithesis_instrumentation__.Notify(568784)
					if oldEntry.Family < newEntry.Family {
						__antithesis_instrumentation__.Notify(568785)
						if oldEntry.Family == descpb.FamilyID(0) {
							__antithesis_instrumentation__.Notify(568788)
							return nil, errors.AssertionFailedf(
								"index entry for family 0 for table %s, index %s was not generated",
								ru.Helper.TableDesc.GetName(), index.GetName(),
							)
						} else {
							__antithesis_instrumentation__.Notify(568789)
						}
						__antithesis_instrumentation__.Notify(568786)

						if err := ru.Helper.deleteIndexEntry(ctx, batch, index, ru.Helper.secIndexValDirs[i], oldEntry, traceKV); err != nil {
							__antithesis_instrumentation__.Notify(568790)
							return nil, err
						} else {
							__antithesis_instrumentation__.Notify(568791)
						}
						__antithesis_instrumentation__.Notify(568787)
						oldIdx++
					} else {
						__antithesis_instrumentation__.Notify(568792)
						if newEntry.Family == descpb.FamilyID(0) {
							__antithesis_instrumentation__.Notify(568795)
							return nil, errors.AssertionFailedf(
								"index entry for family 0 for table %s, index %s was not generated",
								ru.Helper.TableDesc.GetName(), index.GetName(),
							)
						} else {
							__antithesis_instrumentation__.Notify(568796)
						}
						__antithesis_instrumentation__.Notify(568793)

						if index.ForcePut() {
							__antithesis_instrumentation__.Notify(568797)

							insertPutFn(ctx, batch, &newEntry.Key, &newEntry.Value, traceKV)
						} else {
							__antithesis_instrumentation__.Notify(568798)

							if traceKV {
								__antithesis_instrumentation__.Notify(568800)
								k := keys.PrettyPrint(ru.Helper.secIndexValDirs[i], newEntry.Key)
								v := newEntry.Value.PrettyPrint()
								log.VEventf(ctx, 2, "CPut %s -> %v (expecting does not exist)", k, v)
							} else {
								__antithesis_instrumentation__.Notify(568801)
							}
							__antithesis_instrumentation__.Notify(568799)
							batch.CPut(newEntry.Key, &newEntry.Value, nil)
						}
						__antithesis_instrumentation__.Notify(568794)
						newIdx++
					}
				}
			}
			__antithesis_instrumentation__.Notify(568763)
			for oldIdx < len(oldEntries) {
				__antithesis_instrumentation__.Notify(568802)

				oldEntry := &oldEntries[oldIdx]
				if err := ru.Helper.deleteIndexEntry(ctx, batch, index, ru.Helper.secIndexValDirs[i], oldEntry, traceKV); err != nil {
					__antithesis_instrumentation__.Notify(568804)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(568805)
				}
				__antithesis_instrumentation__.Notify(568803)
				oldIdx++
			}
			__antithesis_instrumentation__.Notify(568764)
			for newIdx < len(newEntries) {
				__antithesis_instrumentation__.Notify(568806)

				newEntry := &newEntries[newIdx]
				if index.ForcePut() {
					__antithesis_instrumentation__.Notify(568808)

					insertPutFn(ctx, batch, &newEntry.Key, &newEntry.Value, traceKV)
				} else {
					__antithesis_instrumentation__.Notify(568809)
					if traceKV {
						__antithesis_instrumentation__.Notify(568811)
						k := keys.PrettyPrint(ru.Helper.secIndexValDirs[i], newEntry.Key)
						v := newEntry.Value.PrettyPrint()
						log.VEventf(ctx, 2, "CPut %s -> %v (expecting does not exist)", k, v)
					} else {
						__antithesis_instrumentation__.Notify(568812)
					}
					__antithesis_instrumentation__.Notify(568810)
					batch.CPut(newEntry.Key, &newEntry.Value, nil)
				}
				__antithesis_instrumentation__.Notify(568807)
				newIdx++
			}
		} else {
			__antithesis_instrumentation__.Notify(568813)

			for j := range ru.oldIndexEntries[i] {
				__antithesis_instrumentation__.Notify(568815)
				if err := ru.Helper.deleteIndexEntry(ctx, batch, index, nil, &ru.oldIndexEntries[i][j], traceKV); err != nil {
					__antithesis_instrumentation__.Notify(568816)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(568817)
				}
			}
			__antithesis_instrumentation__.Notify(568814)

			for j := range ru.newIndexEntries[i] {
				__antithesis_instrumentation__.Notify(568818)
				if index.ForcePut() {
					__antithesis_instrumentation__.Notify(568819)

					insertPutFn(ctx, batch, &ru.newIndexEntries[i][j].Key, &ru.newIndexEntries[i][j].Value, traceKV)
				} else {
					__antithesis_instrumentation__.Notify(568820)
					insertInvertedPutFn(ctx, batch, &ru.newIndexEntries[i][j].Key, &ru.newIndexEntries[i][j].Value, traceKV)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(568705)

	if ru.DeleteHelper != nil {
		__antithesis_instrumentation__.Notify(568821)

		for idx := range ru.DeleteHelper.Indexes {
			__antithesis_instrumentation__.Notify(568822)
			index := ru.DeleteHelper.Indexes[idx]
			deletedSecondaryIndexEntries, ok := deleteOldSecondaryIndexEntries[index]

			if ok {
				__antithesis_instrumentation__.Notify(568823)
				for _, deletedSecondaryIndexEntry := range deletedSecondaryIndexEntries {
					__antithesis_instrumentation__.Notify(568824)
					if err := ru.DeleteHelper.deleteIndexEntry(ctx, batch, index, nil, &deletedSecondaryIndexEntry, traceKV); err != nil {
						__antithesis_instrumentation__.Notify(568825)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(568826)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(568827)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(568828)
	}
	__antithesis_instrumentation__.Notify(568706)

	return ru.newValues, nil
}

func compareIndexEntries(left, right rowenc.IndexEntry) int {
	__antithesis_instrumentation__.Notify(568829)
	cmp := bytes.Compare(left.Key, right.Key)
	if cmp != 0 {
		__antithesis_instrumentation__.Notify(568831)
		return cmp
	} else {
		__antithesis_instrumentation__.Notify(568832)
	}
	__antithesis_instrumentation__.Notify(568830)

	return bytes.Compare(left.Value.RawBytes, right.Value.RawBytes)
}

func (ru *Updater) IsColumnOnlyUpdate() bool {
	__antithesis_instrumentation__.Notify(568833)

	return !ru.primaryKeyColChange && func() bool {
		__antithesis_instrumentation__.Notify(568834)
		return ru.DeleteHelper == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(568835)
		return len(ru.Helper.Indexes) == 0 == true
	}() == true
}
