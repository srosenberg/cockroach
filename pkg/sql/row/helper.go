package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/rowencpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

const (
	maxRowSizeFloor = 1 << 10

	maxRowSizeCeil = 1 << 30
)

var maxRowSizeLog = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.guardrails.max_row_size_log",
	"maximum size of row (or column family if multiple column families are in use) that SQL can "+
		"write to the database, above which an event is logged to SQL_PERF (or SQL_INTERNAL_PERF "+
		"if the mutating statement was internal); use 0 to disable",
	kvserver.MaxCommandSizeDefault,
	func(size int64) error {
		__antithesis_instrumentation__.Notify(568024)
		if size != 0 && func() bool {
			__antithesis_instrumentation__.Notify(568026)
			return size < maxRowSizeFloor == true
		}() == true {
			__antithesis_instrumentation__.Notify(568027)
			return errors.Newf(
				"cannot set sql.guardrails.max_row_size_log to %v, must be 0 or >= %v",
				size, maxRowSizeFloor,
			)
		} else {
			__antithesis_instrumentation__.Notify(568028)
			if size > maxRowSizeCeil {
				__antithesis_instrumentation__.Notify(568029)
				return errors.Newf(
					"cannot set sql.guardrails.max_row_size_log to %v, must be <= %v",
					size, maxRowSizeCeil,
				)
			} else {
				__antithesis_instrumentation__.Notify(568030)
			}
		}
		__antithesis_instrumentation__.Notify(568025)
		return nil
	},
).WithPublic()

var maxRowSizeErr = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.guardrails.max_row_size_err",
	"maximum size of row (or column family if multiple column families are in use) that SQL can "+
		"write to the database, above which an error is returned; use 0 to disable",
	512<<20,
	func(size int64) error {
		__antithesis_instrumentation__.Notify(568031)
		if size != 0 && func() bool {
			__antithesis_instrumentation__.Notify(568033)
			return size < maxRowSizeFloor == true
		}() == true {
			__antithesis_instrumentation__.Notify(568034)
			return errors.Newf(
				"cannot set sql.guardrails.max_row_size_err to %v, must be 0 or >= %v",
				size, maxRowSizeFloor,
			)
		} else {
			__antithesis_instrumentation__.Notify(568035)
			if size > maxRowSizeCeil {
				__antithesis_instrumentation__.Notify(568036)
				return errors.Newf(
					"cannot set sql.guardrails.max_row_size_err to %v, must be <= %v",
					size, maxRowSizeCeil,
				)
			} else {
				__antithesis_instrumentation__.Notify(568037)
			}
		}
		__antithesis_instrumentation__.Notify(568032)
		return nil
	},
).WithPublic()

type rowHelper struct {
	Codec keys.SQLCodec

	TableDesc catalog.TableDescriptor

	Indexes      []catalog.Index
	indexEntries map[catalog.Index][]rowenc.IndexEntry

	primIndexValDirs []encoding.Direction
	secIndexValDirs  [][]encoding.Direction

	primaryIndexKeyPrefix []byte
	primaryIndexKeyCols   catalog.TableColSet
	primaryIndexValueCols catalog.TableColSet
	sortedColumnFamilies  map[descpb.FamilyID][]descpb.ColumnID

	maxRowSizeLog, maxRowSizeErr uint32
	internal                     bool
	metrics                      *Metrics
}

func newRowHelper(
	codec keys.SQLCodec,
	desc catalog.TableDescriptor,
	indexes []catalog.Index,
	sv *settings.Values,
	internal bool,
	metrics *Metrics,
) rowHelper {
	__antithesis_instrumentation__.Notify(568038)
	rh := rowHelper{
		Codec:     codec,
		TableDesc: desc,
		Indexes:   indexes,
		internal:  internal,
		metrics:   metrics,
	}

	rh.primIndexValDirs = catalogkeys.IndexKeyValDirs(rh.TableDesc.GetPrimaryIndex())

	rh.secIndexValDirs = make([][]encoding.Direction, len(rh.Indexes))
	for i := range rh.Indexes {
		__antithesis_instrumentation__.Notify(568040)
		rh.secIndexValDirs[i] = catalogkeys.IndexKeyValDirs(rh.Indexes[i])
	}
	__antithesis_instrumentation__.Notify(568039)

	rh.maxRowSizeLog = uint32(maxRowSizeLog.Get(sv))
	rh.maxRowSizeErr = uint32(maxRowSizeErr.Get(sv))

	return rh
}

func (rh *rowHelper) encodeIndexes(
	colIDtoRowIndex catalog.TableColMap,
	values []tree.Datum,
	ignoreIndexes util.FastIntSet,
	includeEmpty bool,
) (
	primaryIndexKey []byte,
	secondaryIndexEntries map[catalog.Index][]rowenc.IndexEntry,
	err error,
) {
	__antithesis_instrumentation__.Notify(568041)
	primaryIndexKey, err = rh.encodePrimaryIndex(colIDtoRowIndex, values)
	if err != nil {
		__antithesis_instrumentation__.Notify(568044)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(568045)
	}
	__antithesis_instrumentation__.Notify(568042)
	secondaryIndexEntries, err = rh.encodeSecondaryIndexes(colIDtoRowIndex, values, ignoreIndexes, includeEmpty)
	if err != nil {
		__antithesis_instrumentation__.Notify(568046)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(568047)
	}
	__antithesis_instrumentation__.Notify(568043)
	return primaryIndexKey, secondaryIndexEntries, nil
}

func (rh *rowHelper) encodePrimaryIndex(
	colIDtoRowIndex catalog.TableColMap, values []tree.Datum,
) (primaryIndexKey []byte, err error) {
	__antithesis_instrumentation__.Notify(568048)
	if rh.primaryIndexKeyPrefix == nil {
		__antithesis_instrumentation__.Notify(568051)
		rh.primaryIndexKeyPrefix = rowenc.MakeIndexKeyPrefix(
			rh.Codec, rh.TableDesc.GetID(), rh.TableDesc.GetPrimaryIndexID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(568052)
	}
	__antithesis_instrumentation__.Notify(568049)
	idx := rh.TableDesc.GetPrimaryIndex()
	primaryIndexKey, containsNull, err := rowenc.EncodeIndexKey(
		rh.TableDesc, idx, colIDtoRowIndex, values, rh.primaryIndexKeyPrefix,
	)
	if containsNull {
		__antithesis_instrumentation__.Notify(568053)
		return nil, rowenc.MakeNullPKError(rh.TableDesc, idx, colIDtoRowIndex, values)
	} else {
		__antithesis_instrumentation__.Notify(568054)
	}
	__antithesis_instrumentation__.Notify(568050)
	return primaryIndexKey, err
}

func (rh *rowHelper) encodeSecondaryIndexes(
	colIDtoRowIndex catalog.TableColMap,
	values []tree.Datum,
	ignoreIndexes util.FastIntSet,
	includeEmpty bool,
) (secondaryIndexEntries map[catalog.Index][]rowenc.IndexEntry, err error) {
	__antithesis_instrumentation__.Notify(568055)

	if rh.indexEntries == nil {
		__antithesis_instrumentation__.Notify(568059)
		rh.indexEntries = make(map[catalog.Index][]rowenc.IndexEntry, len(rh.Indexes))
	} else {
		__antithesis_instrumentation__.Notify(568060)
	}
	__antithesis_instrumentation__.Notify(568056)

	for i := range rh.indexEntries {
		__antithesis_instrumentation__.Notify(568061)
		rh.indexEntries[i] = rh.indexEntries[i][:0]
	}
	__antithesis_instrumentation__.Notify(568057)

	for i := range rh.Indexes {
		__antithesis_instrumentation__.Notify(568062)
		index := rh.Indexes[i]
		if !ignoreIndexes.Contains(int(index.GetID())) {
			__antithesis_instrumentation__.Notify(568063)
			entries, err := rowenc.EncodeSecondaryIndex(rh.Codec, rh.TableDesc, index, colIDtoRowIndex, values, includeEmpty)
			if err != nil {
				__antithesis_instrumentation__.Notify(568065)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(568066)
			}
			__antithesis_instrumentation__.Notify(568064)
			rh.indexEntries[index] = append(rh.indexEntries[index], entries...)
		} else {
			__antithesis_instrumentation__.Notify(568067)
		}
	}
	__antithesis_instrumentation__.Notify(568058)

	return rh.indexEntries, nil
}

func (rh *rowHelper) skipColumnNotInPrimaryIndexValue(
	colID descpb.ColumnID, value tree.Datum,
) (bool, error) {
	__antithesis_instrumentation__.Notify(568068)
	if rh.primaryIndexKeyCols.Empty() {
		__antithesis_instrumentation__.Notify(568072)
		rh.primaryIndexKeyCols = rh.TableDesc.GetPrimaryIndex().CollectKeyColumnIDs()
		rh.primaryIndexValueCols = rh.TableDesc.GetPrimaryIndex().CollectPrimaryStoredColumnIDs()
	} else {
		__antithesis_instrumentation__.Notify(568073)
	}
	__antithesis_instrumentation__.Notify(568069)
	if !rh.primaryIndexKeyCols.Contains(colID) {
		__antithesis_instrumentation__.Notify(568074)
		return !rh.primaryIndexValueCols.Contains(colID), nil
	} else {
		__antithesis_instrumentation__.Notify(568075)
	}
	__antithesis_instrumentation__.Notify(568070)
	if cdatum, ok := value.(tree.CompositeDatum); ok {
		__antithesis_instrumentation__.Notify(568076)

		return !cdatum.IsComposite(), nil
	} else {
		__antithesis_instrumentation__.Notify(568077)
	}
	__antithesis_instrumentation__.Notify(568071)

	return true, nil
}

func (rh *rowHelper) sortedColumnFamily(famID descpb.FamilyID) ([]descpb.ColumnID, bool) {
	__antithesis_instrumentation__.Notify(568078)
	if rh.sortedColumnFamilies == nil {
		__antithesis_instrumentation__.Notify(568080)
		rh.sortedColumnFamilies = make(map[descpb.FamilyID][]descpb.ColumnID, rh.TableDesc.NumFamilies())

		_ = rh.TableDesc.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
			__antithesis_instrumentation__.Notify(568081)
			colIDs := append([]descpb.ColumnID{}, family.ColumnIDs...)
			sort.Sort(descpb.ColumnIDs(colIDs))
			rh.sortedColumnFamilies[family.ID] = colIDs
			return nil
		})
	} else {
		__antithesis_instrumentation__.Notify(568082)
	}
	__antithesis_instrumentation__.Notify(568079)
	colIDs, ok := rh.sortedColumnFamilies[famID]
	return colIDs, ok
}

func (rh *rowHelper) checkRowSize(
	ctx context.Context, key *roachpb.Key, value *roachpb.Value, family descpb.FamilyID,
) error {
	__antithesis_instrumentation__.Notify(568083)
	size := uint32(len(*key)) + uint32(len(value.RawBytes))
	shouldLog := rh.maxRowSizeLog != 0 && func() bool {
		__antithesis_instrumentation__.Notify(568088)
		return size > rh.maxRowSizeLog == true
	}() == true
	shouldErr := rh.maxRowSizeErr != 0 && func() bool {
		__antithesis_instrumentation__.Notify(568089)
		return size > rh.maxRowSizeErr == true
	}() == true
	if !shouldLog && func() bool {
		__antithesis_instrumentation__.Notify(568090)
		return !shouldErr == true
	}() == true {
		__antithesis_instrumentation__.Notify(568091)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(568092)
	}
	__antithesis_instrumentation__.Notify(568084)
	details := eventpb.CommonLargeRowDetails{
		RowSize:    size,
		TableID:    uint32(rh.TableDesc.GetID()),
		FamilyID:   uint32(family),
		PrimaryKey: keys.PrettyPrint(rh.primIndexValDirs, *key),
	}
	if rh.internal && func() bool {
		__antithesis_instrumentation__.Notify(568093)
		return shouldErr == true
	}() == true {
		__antithesis_instrumentation__.Notify(568094)

		shouldErr = false
		shouldLog = true
	} else {
		__antithesis_instrumentation__.Notify(568095)
	}
	__antithesis_instrumentation__.Notify(568085)
	if shouldLog {
		__antithesis_instrumentation__.Notify(568096)
		if rh.metrics != nil {
			__antithesis_instrumentation__.Notify(568099)
			rh.metrics.MaxRowSizeLogCount.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(568100)
		}
		__antithesis_instrumentation__.Notify(568097)
		var event eventpb.EventPayload
		if rh.internal {
			__antithesis_instrumentation__.Notify(568101)
			event = &eventpb.LargeRowInternal{CommonLargeRowDetails: details}
		} else {
			__antithesis_instrumentation__.Notify(568102)
			event = &eventpb.LargeRow{CommonLargeRowDetails: details}
		}
		__antithesis_instrumentation__.Notify(568098)
		log.StructuredEvent(ctx, event)
	} else {
		__antithesis_instrumentation__.Notify(568103)
	}
	__antithesis_instrumentation__.Notify(568086)
	if shouldErr {
		__antithesis_instrumentation__.Notify(568104)
		if rh.metrics != nil {
			__antithesis_instrumentation__.Notify(568106)
			rh.metrics.MaxRowSizeErrCount.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(568107)
		}
		__antithesis_instrumentation__.Notify(568105)
		return pgerror.WithCandidateCode(&details, pgcode.ProgramLimitExceeded)
	} else {
		__antithesis_instrumentation__.Notify(568108)
	}
	__antithesis_instrumentation__.Notify(568087)
	return nil
}

var deleteEncoding protoutil.Message = &rowencpb.IndexValueWrapper{
	Value:   nil,
	Deleted: true,
}

func (rh *rowHelper) deleteIndexEntry(
	ctx context.Context,
	batch *kv.Batch,
	index catalog.Index,
	valDirs []encoding.Direction,
	entry *rowenc.IndexEntry,
	traceKV bool,
) error {
	__antithesis_instrumentation__.Notify(568109)
	if index.UseDeletePreservingEncoding() {
		__antithesis_instrumentation__.Notify(568111)
		if traceKV {
			__antithesis_instrumentation__.Notify(568113)
			log.VEventf(ctx, 2, "Put (delete) %s", entry.Key)
		} else {
			__antithesis_instrumentation__.Notify(568114)
		}
		__antithesis_instrumentation__.Notify(568112)

		batch.Put(entry.Key, deleteEncoding)
	} else {
		__antithesis_instrumentation__.Notify(568115)
		if traceKV {
			__antithesis_instrumentation__.Notify(568117)
			if valDirs != nil {
				__antithesis_instrumentation__.Notify(568118)
				log.VEventf(ctx, 2, "Del %s", keys.PrettyPrint(valDirs, entry.Key))
			} else {
				__antithesis_instrumentation__.Notify(568119)
				log.VEventf(ctx, 2, "Del %s", entry.Key)
			}
		} else {
			__antithesis_instrumentation__.Notify(568120)
		}
		__antithesis_instrumentation__.Notify(568116)

		batch.Del(entry.Key)
	}
	__antithesis_instrumentation__.Notify(568110)
	return nil
}
