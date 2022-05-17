package rowenc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/rowencpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/unique"
	"github.com/cockroachdb/errors"
)

func MakeIndexKeyPrefix(codec keys.SQLCodec, tableID descpb.ID, indexID descpb.IndexID) []byte {
	__antithesis_instrumentation__.Notify(570156)
	return codec.IndexPrefix(uint32(tableID), uint32(indexID))
}

func EncodeIndexKey(
	tableDesc catalog.TableDescriptor,
	index catalog.Index,
	colMap catalog.TableColMap,
	values []tree.Datum,
	keyPrefix []byte,
) (key []byte, containsNull bool, err error) {
	__antithesis_instrumentation__.Notify(570157)
	keyAndSuffixCols := tableDesc.IndexFetchSpecKeyAndSuffixColumns(index)
	keyCols := keyAndSuffixCols[:index.NumKeyColumns()]
	key, containsNull, err = EncodePartialIndexKey(
		keyCols,
		colMap,
		values,
		keyPrefix,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(570159)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(570160)
	}
	__antithesis_instrumentation__.Notify(570158)
	return key, containsNull, err
}

func EncodePartialIndexSpan(
	keyCols []descpb.IndexFetchSpec_KeyColumn,
	colMap catalog.TableColMap,
	values []tree.Datum,
	keyPrefix []byte,
) (span roachpb.Span, containsNull bool, err error) {
	__antithesis_instrumentation__.Notify(570161)
	var key roachpb.Key
	key, containsNull, err = EncodePartialIndexKey(keyCols, colMap, values, keyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(570163)
		return span, false, err
	} else {
		__antithesis_instrumentation__.Notify(570164)
	}
	__antithesis_instrumentation__.Notify(570162)
	return roachpb.Span{Key: key, EndKey: key.PrefixEnd()}, containsNull, nil
}

func EncodePartialIndexKey(
	keyCols []descpb.IndexFetchSpec_KeyColumn,
	colMap catalog.TableColMap,
	values []tree.Datum,
	keyPrefix []byte,
) (key []byte, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(570165)

	key = growKey(keyPrefix, len(keyPrefix)+2*len(values))

	for i := range keyCols {
		__antithesis_instrumentation__.Notify(570167)
		keyCol := &keyCols[i]
		val := findColumnValue(keyCol.ColumnID, colMap, values)
		if val == tree.DNull {
			__antithesis_instrumentation__.Notify(570170)
			containsNull = true
		} else {
			__antithesis_instrumentation__.Notify(570171)
		}
		__antithesis_instrumentation__.Notify(570168)

		dir, err := keyCol.Direction.ToEncodingDirection()
		if err != nil {
			__antithesis_instrumentation__.Notify(570172)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570173)
		}
		__antithesis_instrumentation__.Notify(570169)

		if key, err = keyside.Encode(key, val, dir); err != nil {
			__antithesis_instrumentation__.Notify(570174)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570175)
		}
	}
	__antithesis_instrumentation__.Notify(570166)
	return key, containsNull, nil
}

type directions []descpb.IndexDescriptor_Direction

func (d directions) get(i int) (encoding.Direction, error) {
	__antithesis_instrumentation__.Notify(570176)
	if i < len(d) {
		__antithesis_instrumentation__.Notify(570178)
		return d[i].ToEncodingDirection()
	} else {
		__antithesis_instrumentation__.Notify(570179)
	}
	__antithesis_instrumentation__.Notify(570177)
	return encoding.Ascending, nil
}

func MakeSpanFromEncDatums(
	values EncDatumRow,
	keyCols []descpb.IndexFetchSpec_KeyColumn,
	alloc *tree.DatumAlloc,
	keyPrefix []byte,
) (_ roachpb.Span, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(570180)
	startKey, containsNull, err := MakeKeyFromEncDatums(values, keyCols, alloc, keyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(570182)
		return roachpb.Span{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(570183)
	}
	__antithesis_instrumentation__.Notify(570181)
	return roachpb.Span{Key: startKey, EndKey: startKey.PrefixEnd()}, containsNull, nil
}

func NeededColumnFamilyIDs(
	neededColOrdinals util.FastIntSet, table catalog.TableDescriptor, index catalog.Index,
) []descpb.FamilyID {
	__antithesis_instrumentation__.Notify(570184)
	if table.NumFamilies() == 1 {
		__antithesis_instrumentation__.Notify(570194)
		return []descpb.FamilyID{table.GetFamilies()[0].ID}
	} else {
		__antithesis_instrumentation__.Notify(570195)
	}
	__antithesis_instrumentation__.Notify(570185)

	columns := table.DeletableColumns()
	colIdxMap := catalog.ColumnIDToOrdinalMap(columns)
	var indexedCols util.FastIntSet
	var compositeCols util.FastIntSet
	var extraCols util.FastIntSet
	for i := 0; i < index.NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(570196)
		columnID := index.GetKeyColumnID(i)
		columnOrdinal := colIdxMap.GetDefault(columnID)
		indexedCols.Add(columnOrdinal)
	}
	__antithesis_instrumentation__.Notify(570186)
	for i := 0; i < index.NumCompositeColumns(); i++ {
		__antithesis_instrumentation__.Notify(570197)
		columnID := index.GetCompositeColumnID(i)
		columnOrdinal := colIdxMap.GetDefault(columnID)
		compositeCols.Add(columnOrdinal)
	}
	__antithesis_instrumentation__.Notify(570187)
	for i := 0; i < index.NumKeySuffixColumns(); i++ {
		__antithesis_instrumentation__.Notify(570198)
		columnID := index.GetKeySuffixColumnID(i)
		columnOrdinal := colIdxMap.GetDefault(columnID)
		extraCols.Add(columnOrdinal)
	}
	__antithesis_instrumentation__.Notify(570188)

	var family0 *descpb.ColumnFamilyDescriptor
	hasSecondaryEncoding := index.GetEncodingType() == descpb.SecondaryIndexEncoding

	family0Needed := false
	mvccColumnRequested := false
	nc := neededColOrdinals.Copy()
	neededColOrdinals.ForEach(func(columnOrdinal int) {
		__antithesis_instrumentation__.Notify(570199)
		if indexedCols.Contains(columnOrdinal) && func() bool {
			__antithesis_instrumentation__.Notify(570202)
			return !compositeCols.Contains(columnOrdinal) == true
		}() == true {
			__antithesis_instrumentation__.Notify(570203)

			nc.Remove(columnOrdinal)
		} else {
			__antithesis_instrumentation__.Notify(570204)
		}
		__antithesis_instrumentation__.Notify(570200)
		if hasSecondaryEncoding && func() bool {
			__antithesis_instrumentation__.Notify(570205)
			return (compositeCols.Contains(columnOrdinal) || func() bool {
				__antithesis_instrumentation__.Notify(570206)
				return extraCols.Contains(columnOrdinal) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(570207)

			family0Needed = true
			nc.Remove(columnOrdinal)
		} else {
			__antithesis_instrumentation__.Notify(570208)
		}
		__antithesis_instrumentation__.Notify(570201)

		if columnOrdinal >= len(columns) {
			__antithesis_instrumentation__.Notify(570209)
			mvccColumnRequested = true
		} else {
			__antithesis_instrumentation__.Notify(570210)
		}
	})
	__antithesis_instrumentation__.Notify(570189)

	if mvccColumnRequested {
		__antithesis_instrumentation__.Notify(570211)
		families := make([]descpb.FamilyID, 0, table.NumFamilies())
		_ = table.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
			__antithesis_instrumentation__.Notify(570213)
			families = append(families, family.ID)
			return nil
		})
		__antithesis_instrumentation__.Notify(570212)
		return families
	} else {
		__antithesis_instrumentation__.Notify(570214)
	}
	__antithesis_instrumentation__.Notify(570190)

	var neededFamilyIDs []descpb.FamilyID
	allFamiliesNullable := true
	_ = table.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
		__antithesis_instrumentation__.Notify(570215)
		needed := false
		nullable := true
		if family.ID == 0 {
			__antithesis_instrumentation__.Notify(570219)

			family0 = family
			if family0Needed {
				__antithesis_instrumentation__.Notify(570221)
				needed = true
			} else {
				__antithesis_instrumentation__.Notify(570222)
			}
			__antithesis_instrumentation__.Notify(570220)
			nullable = false
		} else {
			__antithesis_instrumentation__.Notify(570223)
		}
		__antithesis_instrumentation__.Notify(570216)
		for _, columnID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(570224)
			if needed && func() bool {
				__antithesis_instrumentation__.Notify(570227)
				return !nullable == true
			}() == true {
				__antithesis_instrumentation__.Notify(570228)

				break
			} else {
				__antithesis_instrumentation__.Notify(570229)
			}
			__antithesis_instrumentation__.Notify(570225)
			columnOrdinal := colIdxMap.GetDefault(columnID)
			if nc.Contains(columnOrdinal) {
				__antithesis_instrumentation__.Notify(570230)
				needed = true
			} else {
				__antithesis_instrumentation__.Notify(570231)
			}
			__antithesis_instrumentation__.Notify(570226)
			if !columns[columnOrdinal].IsNullable() && func() bool {
				__antithesis_instrumentation__.Notify(570232)
				return !indexedCols.Contains(columnOrdinal) == true
			}() == true {
				__antithesis_instrumentation__.Notify(570233)

				nullable = false
			} else {
				__antithesis_instrumentation__.Notify(570234)
			}
		}
		__antithesis_instrumentation__.Notify(570217)
		if needed {
			__antithesis_instrumentation__.Notify(570235)
			neededFamilyIDs = append(neededFamilyIDs, family.ID)
			allFamiliesNullable = allFamiliesNullable && func() bool {
				__antithesis_instrumentation__.Notify(570236)
				return nullable == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(570237)
		}
		__antithesis_instrumentation__.Notify(570218)
		return nil
	})
	__antithesis_instrumentation__.Notify(570191)
	if family0 == nil {
		__antithesis_instrumentation__.Notify(570238)
		panic(errors.AssertionFailedf("column family 0 not found"))
	} else {
		__antithesis_instrumentation__.Notify(570239)
	}
	__antithesis_instrumentation__.Notify(570192)

	if allFamiliesNullable {
		__antithesis_instrumentation__.Notify(570240)

		neededFamilyIDs = append(neededFamilyIDs, 0)
		copy(neededFamilyIDs[1:], neededFamilyIDs)
		neededFamilyIDs[0] = family0.ID
	} else {
		__antithesis_instrumentation__.Notify(570241)
	}
	__antithesis_instrumentation__.Notify(570193)

	return neededFamilyIDs
}

func SplitRowKeyIntoFamilySpans(
	appendTo roachpb.Spans, key roachpb.Key, neededFamilies []descpb.FamilyID,
) roachpb.Spans {
	__antithesis_instrumentation__.Notify(570242)
	key = key[:len(key):len(key)]
	for i, familyID := range neededFamilies {
		__antithesis_instrumentation__.Notify(570244)
		var famSpan roachpb.Span
		famSpan.Key = keys.MakeFamilyKey(key, uint32(familyID))

		if i > 0 && func() bool {
			__antithesis_instrumentation__.Notify(570245)
			return familyID == neededFamilies[i-1]+1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(570246)

			appendTo[len(appendTo)-1].EndKey = famSpan.Key.PrefixEnd()
		} else {
			__antithesis_instrumentation__.Notify(570247)
			appendTo = append(appendTo, famSpan)
		}
	}
	__antithesis_instrumentation__.Notify(570243)
	return appendTo
}

func MakeKeyFromEncDatums(
	values EncDatumRow,
	keyCols []descpb.IndexFetchSpec_KeyColumn,
	alloc *tree.DatumAlloc,
	keyPrefix []byte,
) (_ roachpb.Key, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(570248)

	if len(values) > len(keyCols) {
		__antithesis_instrumentation__.Notify(570251)
		return nil, false, errors.Errorf("%d values, %d key cols", len(values), len(keyCols))
	} else {
		__antithesis_instrumentation__.Notify(570252)
	}
	__antithesis_instrumentation__.Notify(570249)

	key := make(roachpb.Key, len(keyPrefix), len(keyPrefix)*2)
	copy(key, keyPrefix)

	for i, val := range values {
		__antithesis_instrumentation__.Notify(570253)
		encoding := descpb.DatumEncoding_ASCENDING_KEY
		if keyCols[i].Direction == descpb.IndexDescriptor_DESC {
			__antithesis_instrumentation__.Notify(570256)
			encoding = descpb.DatumEncoding_DESCENDING_KEY
		} else {
			__antithesis_instrumentation__.Notify(570257)
		}
		__antithesis_instrumentation__.Notify(570254)
		if val.IsNull() {
			__antithesis_instrumentation__.Notify(570258)
			containsNull = true
		} else {
			__antithesis_instrumentation__.Notify(570259)
		}
		__antithesis_instrumentation__.Notify(570255)
		var err error
		key, err = val.Encode(keyCols[i].Type, alloc, encoding, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(570260)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570261)
		}
	}
	__antithesis_instrumentation__.Notify(570250)
	return key, containsNull, nil
}

func findColumnValue(
	column descpb.ColumnID, colMap catalog.TableColMap, values []tree.Datum,
) tree.Datum {
	__antithesis_instrumentation__.Notify(570262)
	if i, ok := colMap.Get(column); ok {
		__antithesis_instrumentation__.Notify(570264)

		return values[i]
	} else {
		__antithesis_instrumentation__.Notify(570265)
	}
	__antithesis_instrumentation__.Notify(570263)
	return tree.DNull
}

func DecodePartialTableIDIndexID(key []byte) ([]byte, descpb.ID, descpb.IndexID, error) {
	__antithesis_instrumentation__.Notify(570266)
	key, tableID, indexID, err := keys.DecodeTableIDIndexID(key)
	return key, descpb.ID(tableID), descpb.IndexID(indexID), err
}

func DecodeIndexKeyPrefix(
	codec keys.SQLCodec, expectedTableID descpb.ID, key []byte,
) (indexID descpb.IndexID, remaining []byte, err error) {
	__antithesis_instrumentation__.Notify(570267)
	key, err = codec.StripTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(570271)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(570272)
	}
	__antithesis_instrumentation__.Notify(570268)
	var tableID descpb.ID
	key, tableID, indexID, err = DecodePartialTableIDIndexID(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(570273)
		return 0, nil, err
	} else {
		__antithesis_instrumentation__.Notify(570274)
	}
	__antithesis_instrumentation__.Notify(570269)
	if tableID != expectedTableID {
		__antithesis_instrumentation__.Notify(570275)
		return 0, nil, errors.Errorf(
			"unexpected table ID %d, expected %d instead", tableID, expectedTableID)
	} else {
		__antithesis_instrumentation__.Notify(570276)
	}
	__antithesis_instrumentation__.Notify(570270)
	return indexID, key, err
}

func DecodeIndexKey(
	codec keys.SQLCodec,
	types []*types.T,
	vals []EncDatum,
	colDirs []descpb.IndexDescriptor_Direction,
	key []byte,
) (remainingKey []byte, foundNull bool, _ error) {
	__antithesis_instrumentation__.Notify(570277)
	key, err := codec.StripTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(570281)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(570282)
	}
	__antithesis_instrumentation__.Notify(570278)
	key, _, _, err = DecodePartialTableIDIndexID(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(570283)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(570284)
	}
	__antithesis_instrumentation__.Notify(570279)
	remainingKey, foundNull, err = DecodeKeyVals(types, vals, colDirs, key)
	if err != nil {
		__antithesis_instrumentation__.Notify(570285)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(570286)
	}
	__antithesis_instrumentation__.Notify(570280)
	return remainingKey, foundNull, nil
}

func DecodeKeyVals(
	types []*types.T, vals []EncDatum, directions []descpb.IndexDescriptor_Direction, key []byte,
) (remainingKey []byte, foundNull bool, _ error) {
	__antithesis_instrumentation__.Notify(570287)
	if directions != nil && func() bool {
		__antithesis_instrumentation__.Notify(570290)
		return len(directions) != len(vals) == true
	}() == true {
		__antithesis_instrumentation__.Notify(570291)
		return nil, false, errors.Errorf("encoding directions doesn't parallel vals: %d vs %d.",
			len(directions), len(vals))
	} else {
		__antithesis_instrumentation__.Notify(570292)
	}
	__antithesis_instrumentation__.Notify(570288)
	for j := range vals {
		__antithesis_instrumentation__.Notify(570293)
		enc := descpb.DatumEncoding_ASCENDING_KEY
		if directions != nil && func() bool {
			__antithesis_instrumentation__.Notify(570296)
			return (directions[j] == descpb.IndexDescriptor_DESC) == true
		}() == true {
			__antithesis_instrumentation__.Notify(570297)
			enc = descpb.DatumEncoding_DESCENDING_KEY
		} else {
			__antithesis_instrumentation__.Notify(570298)
		}
		__antithesis_instrumentation__.Notify(570294)
		var err error
		vals[j], key, err = EncDatumFromBuffer(types[j], enc, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(570299)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570300)
		}
		__antithesis_instrumentation__.Notify(570295)
		foundNull = foundNull || func() bool {
			__antithesis_instrumentation__.Notify(570301)
			return vals[j].IsNull() == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(570289)
	return key, foundNull, nil
}

func DecodeKeyValsUsingSpec(
	keyCols []descpb.IndexFetchSpec_KeyColumn, key []byte, vals []EncDatum,
) (remainingKey []byte, foundNull bool, _ error) {
	__antithesis_instrumentation__.Notify(570302)
	for j := range vals {
		__antithesis_instrumentation__.Notify(570304)
		c := keyCols[j]
		enc := descpb.DatumEncoding_ASCENDING_KEY
		if c.Direction == descpb.IndexDescriptor_DESC {
			__antithesis_instrumentation__.Notify(570307)
			enc = descpb.DatumEncoding_DESCENDING_KEY
		} else {
			__antithesis_instrumentation__.Notify(570308)
		}
		__antithesis_instrumentation__.Notify(570305)
		var err error
		vals[j], key, err = EncDatumFromBuffer(c.Type, enc, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(570309)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570310)
		}
		__antithesis_instrumentation__.Notify(570306)
		foundNull = foundNull || func() bool {
			__antithesis_instrumentation__.Notify(570311)
			return vals[j].IsNull() == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(570303)
	return key, foundNull, nil
}

type IndexEntry struct {
	Key   roachpb.Key
	Value roachpb.Value

	Family descpb.FamilyID
}

type valueEncodedColumn struct {
	id          descpb.ColumnID
	isComposite bool
}

type byID []valueEncodedColumn

func (a byID) Len() int      { __antithesis_instrumentation__.Notify(570312); return len(a) }
func (a byID) Swap(i, j int) { __antithesis_instrumentation__.Notify(570313); a[i], a[j] = a[j], a[i] }
func (a byID) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(570314)
	return a[i].id < a[j].id
}

func EncodeInvertedIndexKeys(
	index catalog.Index, colMap catalog.TableColMap, values []tree.Datum, keyPrefix []byte,
) (key [][]byte, err error) {
	__antithesis_instrumentation__.Notify(570315)
	keyPrefix, err = EncodeInvertedIndexPrefixKeys(index, colMap, values, keyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(570319)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570320)
	}
	__antithesis_instrumentation__.Notify(570316)

	var val tree.Datum
	if i, ok := colMap.Get(index.InvertedColumnID()); ok {
		__antithesis_instrumentation__.Notify(570321)
		val = values[i]
	} else {
		__antithesis_instrumentation__.Notify(570322)
		val = tree.DNull
	}
	__antithesis_instrumentation__.Notify(570317)
	indexGeoConfig := index.GetGeoConfig()
	if !geoindex.IsEmptyConfig(&indexGeoConfig) {
		__antithesis_instrumentation__.Notify(570323)
		return EncodeGeoInvertedIndexTableKeys(val, keyPrefix, indexGeoConfig)
	} else {
		__antithesis_instrumentation__.Notify(570324)
	}
	__antithesis_instrumentation__.Notify(570318)
	return EncodeInvertedIndexTableKeys(val, keyPrefix, index.GetVersion())
}

func EncodeInvertedIndexPrefixKeys(
	index catalog.Index, colMap catalog.TableColMap, values []tree.Datum, keyPrefix []byte,
) (_ []byte, err error) {
	__antithesis_instrumentation__.Notify(570325)
	numColumns := index.NumKeyColumns()

	if numColumns > 1 {
		__antithesis_instrumentation__.Notify(570327)

		colIDs := index.IndexDesc().KeyColumnIDs[:numColumns-1]
		dirs := directions(index.IndexDesc().KeyColumnDirections)

		keyPrefix = growKey(keyPrefix, len(keyPrefix))

		keyPrefix, _, err = EncodeColumns(colIDs, dirs, colMap, values, keyPrefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(570328)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570329)
		}
	} else {
		__antithesis_instrumentation__.Notify(570330)
	}
	__antithesis_instrumentation__.Notify(570326)
	return keyPrefix, nil
}

func EncodeInvertedIndexTableKeys(
	val tree.Datum, inKey []byte, version descpb.IndexDescriptorVersion,
) (key [][]byte, err error) {
	__antithesis_instrumentation__.Notify(570331)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(570334)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(570335)
	}
	__antithesis_instrumentation__.Notify(570332)
	datum := tree.UnwrapDatum(nil, val)
	switch val.ResolvedType().Family() {
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(570336)

		return json.EncodeInvertedIndexKeys(inKey, val.(*tree.DJSON).JSON)
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(570337)
		return encodeArrayInvertedIndexTableKeys(val.(*tree.DArray), inKey, version, false)
	default:
		__antithesis_instrumentation__.Notify(570338)
	}
	__antithesis_instrumentation__.Notify(570333)
	return nil, errors.AssertionFailedf("trying to apply inverted index to unsupported type %s", datum.ResolvedType())
}

func EncodeContainingInvertedIndexSpans(
	evalCtx *tree.EvalContext, val tree.Datum,
) (invertedExpr inverted.Expression, err error) {
	__antithesis_instrumentation__.Notify(570339)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(570341)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(570342)
	}
	__antithesis_instrumentation__.Notify(570340)
	datum := tree.UnwrapDatum(evalCtx, val)
	switch val.ResolvedType().Family() {
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(570343)
		return json.EncodeContainingInvertedIndexSpans(nil, val.(*tree.DJSON).JSON)
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(570344)
		return encodeContainingArrayInvertedIndexSpans(val.(*tree.DArray), nil)
	default:
		__antithesis_instrumentation__.Notify(570345)
		return nil, errors.AssertionFailedf(
			"trying to apply inverted index to unsupported type %s", datum.ResolvedType(),
		)
	}
}

func EncodeContainedInvertedIndexSpans(
	evalCtx *tree.EvalContext, val tree.Datum,
) (invertedExpr inverted.Expression, err error) {
	__antithesis_instrumentation__.Notify(570346)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(570348)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(570349)
	}
	__antithesis_instrumentation__.Notify(570347)
	datum := tree.UnwrapDatum(evalCtx, val)
	switch val.ResolvedType().Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(570350)
		return encodeContainedArrayInvertedIndexSpans(val.(*tree.DArray), nil)
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(570351)
		return json.EncodeContainedInvertedIndexSpans(nil, val.(*tree.DJSON).JSON)
	default:
		__antithesis_instrumentation__.Notify(570352)
		return nil, errors.AssertionFailedf(
			"trying to apply inverted index to unsupported type %s", datum.ResolvedType(),
		)
	}
}

func encodeArrayInvertedIndexTableKeys(
	val *tree.DArray, inKey []byte, version descpb.IndexDescriptorVersion, excludeNulls bool,
) (key [][]byte, err error) {
	__antithesis_instrumentation__.Notify(570353)
	if val.Array.Len() == 0 {
		__antithesis_instrumentation__.Notify(570356)
		if version >= descpb.EmptyArraysInInvertedIndexesVersion {
			__antithesis_instrumentation__.Notify(570357)
			return [][]byte{encoding.EncodeEmptyArray(inKey)}, nil
		} else {
			__antithesis_instrumentation__.Notify(570358)
		}
	} else {
		__antithesis_instrumentation__.Notify(570359)
	}
	__antithesis_instrumentation__.Notify(570354)

	outKeys := make([][]byte, 0, len(val.Array))
	for i := range val.Array {
		__antithesis_instrumentation__.Notify(570360)
		d := val.Array[i]
		if d == tree.DNull && func() bool {
			__antithesis_instrumentation__.Notify(570363)
			return (version < descpb.EmptyArraysInInvertedIndexesVersion || func() bool {
				__antithesis_instrumentation__.Notify(570364)
				return excludeNulls == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(570365)

			continue
		} else {
			__antithesis_instrumentation__.Notify(570366)
		}
		__antithesis_instrumentation__.Notify(570361)
		outKey := make([]byte, len(inKey))
		copy(outKey, inKey)
		newKey, err := keyside.Encode(outKey, d, encoding.Ascending)
		if err != nil {
			__antithesis_instrumentation__.Notify(570367)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570368)
		}
		__antithesis_instrumentation__.Notify(570362)
		outKeys = append(outKeys, newKey)
	}
	__antithesis_instrumentation__.Notify(570355)
	outKeys = unique.UniquifyByteSlices(outKeys)
	return outKeys, nil
}

func encodeContainingArrayInvertedIndexSpans(
	val *tree.DArray, inKey []byte,
) (invertedExpr inverted.Expression, err error) {
	__antithesis_instrumentation__.Notify(570369)
	if val.Array.Len() == 0 {
		__antithesis_instrumentation__.Notify(570374)

		invertedExpr = inverted.ExprForSpan(
			inverted.MakeSingleValSpan(inKey), true,
		)
		return invertedExpr, nil
	} else {
		__antithesis_instrumentation__.Notify(570375)
	}
	__antithesis_instrumentation__.Notify(570370)

	if val.HasNulls {
		__antithesis_instrumentation__.Notify(570376)

		return &inverted.SpanExpression{Tight: true, Unique: true}, nil
	} else {
		__antithesis_instrumentation__.Notify(570377)
	}
	__antithesis_instrumentation__.Notify(570371)

	keys, err := encodeArrayInvertedIndexTableKeys(val, inKey, descpb.LatestIndexDescriptorVersion, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(570378)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570379)
	}
	__antithesis_instrumentation__.Notify(570372)
	for _, key := range keys {
		__antithesis_instrumentation__.Notify(570380)
		spanExpr := inverted.ExprForSpan(
			inverted.MakeSingleValSpan(key), true,
		)
		spanExpr.Unique = true
		if invertedExpr == nil {
			__antithesis_instrumentation__.Notify(570381)
			invertedExpr = spanExpr
		} else {
			__antithesis_instrumentation__.Notify(570382)
			invertedExpr = inverted.And(invertedExpr, spanExpr)
		}
	}
	__antithesis_instrumentation__.Notify(570373)
	return invertedExpr, nil
}

func encodeContainedArrayInvertedIndexSpans(
	val *tree.DArray, inKey []byte,
) (invertedExpr inverted.Expression, err error) {
	__antithesis_instrumentation__.Notify(570383)

	emptyArrSpanExpr := inverted.ExprForSpan(
		inverted.MakeSingleValSpan(encoding.EncodeEmptyArray(inKey)), false,
	)
	emptyArrSpanExpr.Unique = true

	if val.Array.Len() == 0 {
		__antithesis_instrumentation__.Notify(570387)
		return emptyArrSpanExpr, nil
	} else {
		__antithesis_instrumentation__.Notify(570388)
	}
	__antithesis_instrumentation__.Notify(570384)

	keys, err := encodeArrayInvertedIndexTableKeys(val, inKey, descpb.LatestIndexDescriptorVersion, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(570389)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570390)
	}
	__antithesis_instrumentation__.Notify(570385)
	invertedExpr = emptyArrSpanExpr
	for _, key := range keys {
		__antithesis_instrumentation__.Notify(570391)
		spanExpr := inverted.ExprForSpan(
			inverted.MakeSingleValSpan(key), false,
		)
		invertedExpr = inverted.Or(invertedExpr, spanExpr)
	}
	__antithesis_instrumentation__.Notify(570386)

	invertedExpr.SetNotTight()
	return invertedExpr, nil
}

func EncodeGeoInvertedIndexTableKeys(
	val tree.Datum, inKey []byte, indexGeoConfig geoindex.Config,
) (key [][]byte, err error) {
	__antithesis_instrumentation__.Notify(570392)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(570394)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(570395)
	}
	__antithesis_instrumentation__.Notify(570393)
	switch val.ResolvedType().Family() {
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(570396)
		index := geoindex.NewS2GeographyIndex(*indexGeoConfig.S2Geography)
		intKeys, bbox, err := index.InvertedIndexKeys(context.TODO(), val.(*tree.DGeography).Geography)
		if err != nil {
			__antithesis_instrumentation__.Notify(570401)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570402)
		}
		__antithesis_instrumentation__.Notify(570397)
		return encodeGeoKeys(encoding.EncodeGeoInvertedAscending(inKey), intKeys, bbox)
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(570398)
		index := geoindex.NewS2GeometryIndex(*indexGeoConfig.S2Geometry)
		intKeys, bbox, err := index.InvertedIndexKeys(context.TODO(), val.(*tree.DGeometry).Geometry)
		if err != nil {
			__antithesis_instrumentation__.Notify(570403)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570404)
		}
		__antithesis_instrumentation__.Notify(570399)
		return encodeGeoKeys(encoding.EncodeGeoInvertedAscending(inKey), intKeys, bbox)
	default:
		__antithesis_instrumentation__.Notify(570400)
		return nil, errors.Errorf("internal error: unexpected type: %s", val.ResolvedType().Family())
	}
}

func encodeGeoKeys(
	inKey []byte, geoKeys []geoindex.Key, bbox geopb.BoundingBox,
) (keys [][]byte, err error) {
	__antithesis_instrumentation__.Notify(570405)
	encodedBBox := make([]byte, 0, encoding.MaxGeoInvertedBBoxLen)
	encodedBBox = encoding.EncodeGeoInvertedBBox(encodedBBox, bbox.LoX, bbox.LoY, bbox.HiX, bbox.HiY)

	b := make([]byte, 0, len(geoKeys)*(len(inKey)+encoding.MaxVarintLen+len(encodedBBox)))
	keys = make([][]byte, len(geoKeys))
	for i, k := range geoKeys {
		__antithesis_instrumentation__.Notify(570407)
		prev := len(b)
		b = append(b, inKey...)
		b = encoding.EncodeUvarintAscending(b, uint64(k))
		b = append(b, encodedBBox...)

		newKey := b[prev:len(b):len(b)]
		keys[i] = newKey
	}
	__antithesis_instrumentation__.Notify(570406)
	return keys, nil
}

func EncodePrimaryIndex(
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	index catalog.Index,
	colMap catalog.TableColMap,
	values []tree.Datum,
	includeEmpty bool,
) ([]IndexEntry, error) {
	__antithesis_instrumentation__.Notify(570408)
	keyPrefix := MakeIndexKeyPrefix(codec, tableDesc.GetID(), index.GetID())
	indexKey, containsNull, err := EncodeIndexKey(tableDesc, index, colMap, values, keyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(570413)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570414)
	}
	__antithesis_instrumentation__.Notify(570409)
	if containsNull {
		__antithesis_instrumentation__.Notify(570415)
		return nil, MakeNullPKError(tableDesc, index, colMap, values)
	} else {
		__antithesis_instrumentation__.Notify(570416)
	}
	__antithesis_instrumentation__.Notify(570410)
	indexedColumns := index.CollectKeyColumnIDs()
	var entryValue []byte
	indexEntries := make([]IndexEntry, 0, tableDesc.NumFamilies())
	var columnsToEncode []valueEncodedColumn
	var called bool
	if err := tableDesc.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
		__antithesis_instrumentation__.Notify(570417)
		if !called {
			__antithesis_instrumentation__.Notify(570423)
			called = true
		} else {
			__antithesis_instrumentation__.Notify(570424)
			indexKey = indexKey[:len(indexKey):len(indexKey)]
			entryValue = entryValue[:0]
			columnsToEncode = columnsToEncode[:0]
		}
		__antithesis_instrumentation__.Notify(570418)
		familyKey := keys.MakeFamilyKey(indexKey, uint32(family.ID))

		if len(family.ColumnIDs) == 1 && func() bool {
			__antithesis_instrumentation__.Notify(570425)
			return family.ColumnIDs[0] == family.DefaultColumnID == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(570426)
			return family.ID != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(570427)
			datum := findColumnValue(family.DefaultColumnID, colMap, values)

			if datum != tree.DNull || func() bool {
				__antithesis_instrumentation__.Notify(570429)
				return includeEmpty == true
			}() == true {
				__antithesis_instrumentation__.Notify(570430)
				col, err := tableDesc.FindColumnWithID(family.DefaultColumnID)
				if err != nil {
					__antithesis_instrumentation__.Notify(570433)
					return err
				} else {
					__antithesis_instrumentation__.Notify(570434)
				}
				__antithesis_instrumentation__.Notify(570431)
				value, err := valueside.MarshalLegacy(col.GetType(), datum)
				if err != nil {
					__antithesis_instrumentation__.Notify(570435)
					return err
				} else {
					__antithesis_instrumentation__.Notify(570436)
				}
				__antithesis_instrumentation__.Notify(570432)
				indexEntries = append(indexEntries, IndexEntry{Key: familyKey, Value: value, Family: family.ID})
			} else {
				__antithesis_instrumentation__.Notify(570437)
			}
			__antithesis_instrumentation__.Notify(570428)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(570438)
		}
		__antithesis_instrumentation__.Notify(570419)

		for _, colID := range family.ColumnIDs {
			__antithesis_instrumentation__.Notify(570439)
			if !indexedColumns.Contains(colID) {
				__antithesis_instrumentation__.Notify(570441)
				columnsToEncode = append(columnsToEncode, valueEncodedColumn{id: colID})
				continue
			} else {
				__antithesis_instrumentation__.Notify(570442)
			}
			__antithesis_instrumentation__.Notify(570440)
			if cdatum, ok := values[colMap.GetDefault(colID)].(tree.CompositeDatum); ok {
				__antithesis_instrumentation__.Notify(570443)
				if cdatum.IsComposite() {
					__antithesis_instrumentation__.Notify(570444)
					columnsToEncode = append(columnsToEncode, valueEncodedColumn{id: colID, isComposite: true})
					continue
				} else {
					__antithesis_instrumentation__.Notify(570445)
				}
			} else {
				__antithesis_instrumentation__.Notify(570446)
			}
		}
		__antithesis_instrumentation__.Notify(570420)
		sort.Sort(byID(columnsToEncode))
		entryValue, err = writeColumnValues(entryValue, colMap, values, columnsToEncode)
		if err != nil {
			__antithesis_instrumentation__.Notify(570447)
			return err
		} else {
			__antithesis_instrumentation__.Notify(570448)
		}
		__antithesis_instrumentation__.Notify(570421)
		if family.ID != 0 && func() bool {
			__antithesis_instrumentation__.Notify(570449)
			return len(entryValue) == 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(570450)
			return !includeEmpty == true
		}() == true {
			__antithesis_instrumentation__.Notify(570451)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(570452)
		}
		__antithesis_instrumentation__.Notify(570422)
		entry := IndexEntry{Key: familyKey, Family: family.ID}
		entry.Value.SetTuple(entryValue)
		indexEntries = append(indexEntries, entry)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(570453)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570454)
	}
	__antithesis_instrumentation__.Notify(570411)

	if index.UseDeletePreservingEncoding() {
		__antithesis_instrumentation__.Notify(570455)
		if err := wrapIndexEntries(indexEntries); err != nil {
			__antithesis_instrumentation__.Notify(570456)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570457)
		}
	} else {
		__antithesis_instrumentation__.Notify(570458)
	}
	__antithesis_instrumentation__.Notify(570412)

	return indexEntries, nil
}

func MakeNullPKError(
	table catalog.TableDescriptor,
	index catalog.Index,
	colMap catalog.TableColMap,
	values []tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(570459)
	for _, col := range table.IndexKeyColumns(index) {
		__antithesis_instrumentation__.Notify(570461)
		ord, ok := colMap.Get(col.GetID())
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(570462)
			return values[ord] == tree.DNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(570463)
			return sqlerrors.NewNonNullViolationError(col.GetName())
		} else {
			__antithesis_instrumentation__.Notify(570464)
		}
	}
	__antithesis_instrumentation__.Notify(570460)
	return errors.AssertionFailedf("NULL value in unknown key column")
}

func EncodeSecondaryIndex(
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	secondaryIndex catalog.Index,
	colMap catalog.TableColMap,
	values []tree.Datum,
	includeEmpty bool,
) ([]IndexEntry, error) {
	__antithesis_instrumentation__.Notify(570465)
	secondaryIndexKeyPrefix := MakeIndexKeyPrefix(codec, tableDesc.GetID(), secondaryIndex.GetID())

	if secondaryIndex.GetEncodingType() == descpb.PrimaryIndexEncoding {
		__antithesis_instrumentation__.Notify(570472)
		return EncodePrimaryIndex(codec, tableDesc, secondaryIndex, colMap, values, includeEmpty)
	} else {
		__antithesis_instrumentation__.Notify(570473)
	}
	__antithesis_instrumentation__.Notify(570466)

	var containsNull = false
	var secondaryKeys [][]byte
	var err error
	if secondaryIndex.GetType() == descpb.IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(570474)
		secondaryKeys, err = EncodeInvertedIndexKeys(secondaryIndex, colMap, values, secondaryIndexKeyPrefix)
	} else {
		__antithesis_instrumentation__.Notify(570475)
		var secondaryIndexKey []byte
		secondaryIndexKey, containsNull, err = EncodeIndexKey(
			tableDesc, secondaryIndex, colMap, values, secondaryIndexKeyPrefix)

		secondaryKeys = [][]byte{secondaryIndexKey}
	}
	__antithesis_instrumentation__.Notify(570467)
	if err != nil {
		__antithesis_instrumentation__.Notify(570476)
		return []IndexEntry{}, err
	} else {
		__antithesis_instrumentation__.Notify(570477)
	}
	__antithesis_instrumentation__.Notify(570468)

	extraKey, _, err := EncodeColumns(
		secondaryIndex.IndexDesc().KeySuffixColumnIDs,
		nil,
		colMap,
		values,
		nil,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(570478)
		return []IndexEntry{}, err
	} else {
		__antithesis_instrumentation__.Notify(570479)
	}
	__antithesis_instrumentation__.Notify(570469)

	entries := make([]IndexEntry, 0, len(secondaryKeys))
	for _, key := range secondaryKeys {
		__antithesis_instrumentation__.Notify(570480)
		if !secondaryIndex.IsUnique() || func() bool {
			__antithesis_instrumentation__.Notify(570482)
			return containsNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(570483)

			key = append(key, extraKey...)
		} else {
			__antithesis_instrumentation__.Notify(570484)
		}
		__antithesis_instrumentation__.Notify(570481)

		if tableDesc.NumFamilies() == 1 || func() bool {
			__antithesis_instrumentation__.Notify(570485)
			return secondaryIndex.GetType() == descpb.IndexDescriptor_INVERTED == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(570486)
			return secondaryIndex.GetVersion() == descpb.BaseIndexFormatVersion == true
		}() == true {
			__antithesis_instrumentation__.Notify(570487)

			entry, err := encodeSecondaryIndexNoFamilies(secondaryIndex, colMap, key, values, extraKey)
			if err != nil {
				__antithesis_instrumentation__.Notify(570489)
				return []IndexEntry{}, err
			} else {
				__antithesis_instrumentation__.Notify(570490)
			}
			__antithesis_instrumentation__.Notify(570488)
			entries = append(entries, entry)
		} else {
			__antithesis_instrumentation__.Notify(570491)

			familyToColumns := make(map[descpb.FamilyID][]valueEncodedColumn)
			addToFamilyColMap := func(id descpb.FamilyID, column valueEncodedColumn) {
				__antithesis_instrumentation__.Notify(570495)
				if _, ok := familyToColumns[id]; !ok {
					__antithesis_instrumentation__.Notify(570497)
					familyToColumns[id] = []valueEncodedColumn{}
				} else {
					__antithesis_instrumentation__.Notify(570498)
				}
				__antithesis_instrumentation__.Notify(570496)
				familyToColumns[id] = append(familyToColumns[id], column)
			}
			__antithesis_instrumentation__.Notify(570492)

			familyToColumns[0] = []valueEncodedColumn{}

			for i := 0; i < secondaryIndex.NumCompositeColumns(); i++ {
				__antithesis_instrumentation__.Notify(570499)
				id := secondaryIndex.GetCompositeColumnID(i)
				addToFamilyColMap(0, valueEncodedColumn{id: id, isComposite: true})
			}
			__antithesis_instrumentation__.Notify(570493)
			_ = tableDesc.ForeachFamily(func(family *descpb.ColumnFamilyDescriptor) error {
				__antithesis_instrumentation__.Notify(570500)
				for i := 0; i < secondaryIndex.NumSecondaryStoredColumns(); i++ {
					__antithesis_instrumentation__.Notify(570502)
					id := secondaryIndex.GetStoredColumnID(i)
					for _, col := range family.ColumnIDs {
						__antithesis_instrumentation__.Notify(570503)
						if id == col {
							__antithesis_instrumentation__.Notify(570504)
							addToFamilyColMap(family.ID, valueEncodedColumn{id: id, isComposite: false})
						} else {
							__antithesis_instrumentation__.Notify(570505)
						}
					}
				}
				__antithesis_instrumentation__.Notify(570501)
				return nil
			})
			__antithesis_instrumentation__.Notify(570494)
			entries, err = encodeSecondaryIndexWithFamilies(
				familyToColumns, secondaryIndex, colMap, key, values, extraKey, entries, includeEmpty)
			if err != nil {
				__antithesis_instrumentation__.Notify(570506)
				return []IndexEntry{}, err
			} else {
				__antithesis_instrumentation__.Notify(570507)
			}
		}
	}
	__antithesis_instrumentation__.Notify(570470)

	if secondaryIndex.UseDeletePreservingEncoding() {
		__antithesis_instrumentation__.Notify(570508)
		if err := wrapIndexEntries(entries); err != nil {
			__antithesis_instrumentation__.Notify(570509)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570510)
		}
	} else {
		__antithesis_instrumentation__.Notify(570511)
	}
	__antithesis_instrumentation__.Notify(570471)

	return entries, nil
}

func encodeSecondaryIndexWithFamilies(
	familyMap map[descpb.FamilyID][]valueEncodedColumn,
	index catalog.Index,
	colMap catalog.TableColMap,
	key []byte,
	row []tree.Datum,
	extraKeyCols []byte,
	results []IndexEntry,
	includeEmpty bool,
) ([]IndexEntry, error) {
	__antithesis_instrumentation__.Notify(570512)
	var (
		value []byte
		err   error
	)
	origKeyLen := len(key)

	familyIDs := make([]int, 0, len(familyMap))
	for familyID := range familyMap {
		__antithesis_instrumentation__.Notify(570515)
		familyIDs = append(familyIDs, int(familyID))
	}
	__antithesis_instrumentation__.Notify(570513)
	sort.Ints(familyIDs)
	for _, familyID := range familyIDs {
		__antithesis_instrumentation__.Notify(570516)
		storedColsInFam := familyMap[descpb.FamilyID(familyID)]

		key = key[:origKeyLen:origKeyLen]

		if len(storedColsInFam) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(570522)
			return familyID != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(570523)
			continue
		} else {
			__antithesis_instrumentation__.Notify(570524)
		}
		__antithesis_instrumentation__.Notify(570517)

		sort.Sort(byID(storedColsInFam))

		key = keys.MakeFamilyKey(key, uint32(familyID))
		if index.IsUnique() && func() bool {
			__antithesis_instrumentation__.Notify(570525)
			return familyID == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(570526)

			value = extraKeyCols
		} else {
			__antithesis_instrumentation__.Notify(570527)

			value = []byte{}
		}
		__antithesis_instrumentation__.Notify(570518)

		value, err = writeColumnValues(value, colMap, row, storedColsInFam)
		if err != nil {
			__antithesis_instrumentation__.Notify(570528)
			return []IndexEntry{}, err
		} else {
			__antithesis_instrumentation__.Notify(570529)
		}
		__antithesis_instrumentation__.Notify(570519)
		entry := IndexEntry{Key: key, Family: descpb.FamilyID(familyID)}

		if familyID != 0 && func() bool {
			__antithesis_instrumentation__.Notify(570530)
			return len(value) == 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(570531)
			return !includeEmpty == true
		}() == true {
			__antithesis_instrumentation__.Notify(570532)
			continue
		} else {
			__antithesis_instrumentation__.Notify(570533)
		}
		__antithesis_instrumentation__.Notify(570520)

		if familyID == 0 {
			__antithesis_instrumentation__.Notify(570534)
			entry.Value.SetBytes(value)
		} else {
			__antithesis_instrumentation__.Notify(570535)
			entry.Value.SetTuple(value)
		}
		__antithesis_instrumentation__.Notify(570521)
		results = append(results, entry)
	}
	__antithesis_instrumentation__.Notify(570514)
	return results, nil
}

func encodeSecondaryIndexNoFamilies(
	index catalog.Index,
	colMap catalog.TableColMap,
	key []byte,
	row []tree.Datum,
	extraKeyCols []byte,
) (IndexEntry, error) {
	__antithesis_instrumentation__.Notify(570536)
	var (
		value []byte
		err   error
	)

	key = keys.MakeFamilyKey(key, 0)
	if index.IsUnique() {
		__antithesis_instrumentation__.Notify(570541)

		value = append(value, extraKeyCols...)
	} else {
		__antithesis_instrumentation__.Notify(570542)

		value = []byte{}
	}
	__antithesis_instrumentation__.Notify(570537)
	var cols []valueEncodedColumn

	for i := 0; i < index.NumSecondaryStoredColumns(); i++ {
		__antithesis_instrumentation__.Notify(570543)
		id := index.GetStoredColumnID(i)
		cols = append(cols, valueEncodedColumn{id: id, isComposite: false})
	}
	__antithesis_instrumentation__.Notify(570538)
	for i := 0; i < index.NumCompositeColumns(); i++ {
		__antithesis_instrumentation__.Notify(570544)
		id := index.GetCompositeColumnID(i)

		if index.GetType() == descpb.IndexDescriptor_INVERTED && func() bool {
			__antithesis_instrumentation__.Notify(570546)
			return id == index.GetKeyColumnID(0) == true
		}() == true {
			__antithesis_instrumentation__.Notify(570547)
			continue
		} else {
			__antithesis_instrumentation__.Notify(570548)
		}
		__antithesis_instrumentation__.Notify(570545)
		cols = append(cols, valueEncodedColumn{id: id, isComposite: true})
	}
	__antithesis_instrumentation__.Notify(570539)
	sort.Sort(byID(cols))
	value, err = writeColumnValues(value, colMap, row, cols)
	if err != nil {
		__antithesis_instrumentation__.Notify(570549)
		return IndexEntry{}, err
	} else {
		__antithesis_instrumentation__.Notify(570550)
	}
	__antithesis_instrumentation__.Notify(570540)
	entry := IndexEntry{Key: key, Family: 0}
	entry.Value.SetBytes(value)
	return entry, nil
}

func writeColumnValues(
	value []byte, colMap catalog.TableColMap, row []tree.Datum, columns []valueEncodedColumn,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(570551)
	var lastColID descpb.ColumnID
	for _, col := range columns {
		__antithesis_instrumentation__.Notify(570553)
		val := findColumnValue(col.id, colMap, row)
		if val == tree.DNull || func() bool {
			__antithesis_instrumentation__.Notify(570555)
			return (col.isComposite && func() bool {
				__antithesis_instrumentation__.Notify(570556)
				return !val.(tree.CompositeDatum).IsComposite() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(570557)
			continue
		} else {
			__antithesis_instrumentation__.Notify(570558)
		}
		__antithesis_instrumentation__.Notify(570554)
		colIDDelta := valueside.MakeColumnIDDelta(lastColID, col.id)
		lastColID = col.id
		var err error
		value, err = valueside.Encode(value, colIDDelta, val, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(570559)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570560)
		}
	}
	__antithesis_instrumentation__.Notify(570552)
	return value, nil
}

func EncodeSecondaryIndexes(
	ctx context.Context,
	codec keys.SQLCodec,
	tableDesc catalog.TableDescriptor,
	indexes []catalog.Index,
	colMap catalog.TableColMap,
	values []tree.Datum,
	secondaryIndexEntries []IndexEntry,
	includeEmpty bool,
	indexBoundAccount *mon.BoundAccount,
) ([]IndexEntry, int64, error) {
	__antithesis_instrumentation__.Notify(570561)
	var memUsedEncodingSecondaryIdxs int64
	if len(secondaryIndexEntries) > 0 {
		__antithesis_instrumentation__.Notify(570565)
		panic(errors.AssertionFailedf("length of secondaryIndexEntries was non-zero"))
	} else {
		__antithesis_instrumentation__.Notify(570566)
	}
	__antithesis_instrumentation__.Notify(570562)

	if indexBoundAccount == nil || func() bool {
		__antithesis_instrumentation__.Notify(570567)
		return indexBoundAccount.Monitor() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(570568)
		panic(errors.AssertionFailedf("memory monitor passed to EncodeSecondaryIndexes was nil"))
	} else {
		__antithesis_instrumentation__.Notify(570569)
	}
	__antithesis_instrumentation__.Notify(570563)
	const sizeOfIndexEntry = int64(unsafe.Sizeof(IndexEntry{}))

	for i := range indexes {
		__antithesis_instrumentation__.Notify(570570)
		entries, err := EncodeSecondaryIndex(codec, tableDesc, indexes[i], colMap, values, includeEmpty)
		if err != nil {
			__antithesis_instrumentation__.Notify(570574)
			return secondaryIndexEntries, 0, err
		} else {
			__antithesis_instrumentation__.Notify(570575)
		}
		__antithesis_instrumentation__.Notify(570571)

		if cap(secondaryIndexEntries)-len(secondaryIndexEntries) < len(entries) {
			__antithesis_instrumentation__.Notify(570576)
			resliceSize := sizeOfIndexEntry * int64(cap(secondaryIndexEntries))
			if err := indexBoundAccount.Grow(ctx, resliceSize); err != nil {
				__antithesis_instrumentation__.Notify(570578)
				return nil, 0, errors.Wrap(err,
					"failed to re-slice index entries buffer")
			} else {
				__antithesis_instrumentation__.Notify(570579)
			}
			__antithesis_instrumentation__.Notify(570577)
			memUsedEncodingSecondaryIdxs += resliceSize
		} else {
			__antithesis_instrumentation__.Notify(570580)
		}
		__antithesis_instrumentation__.Notify(570572)

		for _, index := range entries {
			__antithesis_instrumentation__.Notify(570581)
			if err := indexBoundAccount.Grow(ctx, int64(len(index.Key))); err != nil {
				__antithesis_instrumentation__.Notify(570584)
				return nil, 0, errors.Wrap(err, "failed to allocate space for index keys")
			} else {
				__antithesis_instrumentation__.Notify(570585)
			}
			__antithesis_instrumentation__.Notify(570582)
			memUsedEncodingSecondaryIdxs += int64(len(index.Key))
			if err := indexBoundAccount.Grow(ctx, int64(len(index.Value.RawBytes))); err != nil {
				__antithesis_instrumentation__.Notify(570586)
				return nil, 0, errors.Wrap(err, "failed to allocate space for index values")
			} else {
				__antithesis_instrumentation__.Notify(570587)
			}
			__antithesis_instrumentation__.Notify(570583)
			memUsedEncodingSecondaryIdxs += int64(len(index.Value.RawBytes))
		}
		__antithesis_instrumentation__.Notify(570573)

		secondaryIndexEntries = append(secondaryIndexEntries, entries...)
	}
	__antithesis_instrumentation__.Notify(570564)
	return secondaryIndexEntries, memUsedEncodingSecondaryIdxs, nil
}

func EncodeColumns(
	columnIDs []descpb.ColumnID,
	directions directions,
	colMap catalog.TableColMap,
	values []tree.Datum,
	keyPrefix []byte,
) (key []byte, containsNull bool, err error) {
	__antithesis_instrumentation__.Notify(570588)
	key = keyPrefix
	for colIdx, id := range columnIDs {
		__antithesis_instrumentation__.Notify(570590)
		val := findColumnValue(id, colMap, values)
		if val == tree.DNull {
			__antithesis_instrumentation__.Notify(570593)
			containsNull = true
		} else {
			__antithesis_instrumentation__.Notify(570594)
		}
		__antithesis_instrumentation__.Notify(570591)

		dir, err := directions.get(colIdx)
		if err != nil {
			__antithesis_instrumentation__.Notify(570595)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570596)
		}
		__antithesis_instrumentation__.Notify(570592)

		if key, err = keyside.Encode(key, val, dir); err != nil {
			__antithesis_instrumentation__.Notify(570597)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(570598)
		}
	}
	__antithesis_instrumentation__.Notify(570589)
	return key, containsNull, nil
}

func growKey(key []byte, additionalCapacity int) []byte {
	__antithesis_instrumentation__.Notify(570599)
	newKey := make([]byte, len(key), len(key)+additionalCapacity)
	copy(newKey, key)
	return newKey
}

func getIndexValueWrapperBytes(entry *IndexEntry) ([]byte, error) {
	__antithesis_instrumentation__.Notify(570600)
	var value []byte
	if entry.Value.IsPresent() {
		__antithesis_instrumentation__.Notify(570602)
		value = entry.Value.TagAndDataBytes()
	} else {
		__antithesis_instrumentation__.Notify(570603)
	}
	__antithesis_instrumentation__.Notify(570601)
	tempKV := rowencpb.IndexValueWrapper{
		Value:   value,
		Deleted: false,
	}

	return protoutil.Marshal(&tempKV)
}

func wrapIndexEntries(indexEntries []IndexEntry) error {
	__antithesis_instrumentation__.Notify(570604)
	for i := range indexEntries {
		__antithesis_instrumentation__.Notify(570606)
		encodedEntry, err := getIndexValueWrapperBytes(&indexEntries[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(570608)
			return err
		} else {
			__antithesis_instrumentation__.Notify(570609)
		}
		__antithesis_instrumentation__.Notify(570607)

		indexEntries[i].Value.SetBytes(encodedEntry)
	}
	__antithesis_instrumentation__.Notify(570605)

	return nil
}

func DecodeWrapper(value *roachpb.Value) (*rowencpb.IndexValueWrapper, error) {
	__antithesis_instrumentation__.Notify(570610)
	var wrapper rowencpb.IndexValueWrapper

	valueBytes, err := value.GetBytes()
	if err != nil {
		__antithesis_instrumentation__.Notify(570613)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570614)
	}
	__antithesis_instrumentation__.Notify(570611)

	if err := protoutil.Unmarshal(valueBytes, &wrapper); err != nil {
		__antithesis_instrumentation__.Notify(570615)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570616)
	}
	__antithesis_instrumentation__.Notify(570612)

	return &wrapper, nil
}
