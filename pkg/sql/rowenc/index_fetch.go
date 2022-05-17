package rowenc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

func InitIndexFetchSpec(
	s *descpb.IndexFetchSpec,
	codec keys.SQLCodec,
	table catalog.TableDescriptor,
	index catalog.Index,
	fetchColumnIDs []descpb.ColumnID,
) error {
	__antithesis_instrumentation__.Notify(570617)
	oldFetchedCols := s.FetchedColumns
	*s = descpb.IndexFetchSpec{
		Version:             descpb.IndexFetchSpecVersionInitial,
		TableID:             table.GetID(),
		TableName:           table.GetName(),
		IndexID:             index.GetID(),
		IndexName:           index.GetName(),
		IsSecondaryIndex:    !index.Primary(),
		IsUniqueIndex:       index.IsUnique(),
		EncodingType:        index.GetEncodingType(),
		NumKeySuffixColumns: uint32(index.NumKeySuffixColumns()),
	}

	maxKeysPerRow := table.IndexKeysPerRow(index)
	s.MaxKeysPerRow = uint32(maxKeysPerRow)
	s.KeyPrefixLength = uint32(len(codec.TenantPrefix()) +
		encoding.EncodedLengthUvarintAscending(uint64(s.TableID)) +
		encoding.EncodedLengthUvarintAscending(uint64(index.GetID())))

	s.FamilyDefaultColumns = table.FamilyDefaultColumns()

	families := table.GetFamilies()
	for i := range families {
		__antithesis_instrumentation__.Notify(570623)
		if id := families[i].ID; id > s.MaxFamilyID {
			__antithesis_instrumentation__.Notify(570624)
			s.MaxFamilyID = id
		} else {
			__antithesis_instrumentation__.Notify(570625)
		}
	}
	__antithesis_instrumentation__.Notify(570618)

	s.KeyAndSuffixColumns = table.IndexFetchSpecKeyAndSuffixColumns(index)

	var invertedColumnID descpb.ColumnID
	if index.GetType() == descpb.IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(570626)
		invertedColumnID = index.InvertedColumnID()
	} else {
		__antithesis_instrumentation__.Notify(570627)
	}
	__antithesis_instrumentation__.Notify(570619)

	if cap(oldFetchedCols) >= len(fetchColumnIDs) {
		__antithesis_instrumentation__.Notify(570628)
		s.FetchedColumns = oldFetchedCols[:len(fetchColumnIDs)]
	} else {
		__antithesis_instrumentation__.Notify(570629)
		s.FetchedColumns = make([]descpb.IndexFetchSpec_Column, len(fetchColumnIDs))
	}
	__antithesis_instrumentation__.Notify(570620)
	for i, colID := range fetchColumnIDs {
		__antithesis_instrumentation__.Notify(570630)
		col, err := table.FindColumnWithID(colID)
		if err != nil {
			__antithesis_instrumentation__.Notify(570633)
			return err
		} else {
			__antithesis_instrumentation__.Notify(570634)
		}
		__antithesis_instrumentation__.Notify(570631)
		typ := col.GetType()
		if colID == invertedColumnID {
			__antithesis_instrumentation__.Notify(570635)
			typ = index.InvertedColumnKeyType()
		} else {
			__antithesis_instrumentation__.Notify(570636)
		}
		__antithesis_instrumentation__.Notify(570632)
		s.FetchedColumns[i] = descpb.IndexFetchSpec_Column{
			Name:     col.GetName(),
			ColumnID: colID,
			Type:     typ,
			IsNonNullable: !col.IsNullable() && func() bool {
				__antithesis_instrumentation__.Notify(570637)
				return col.Public() == true
			}() == true,
		}
	}
	__antithesis_instrumentation__.Notify(570621)

	if buildutil.CrdbTestBuild && func() bool {
		__antithesis_instrumentation__.Notify(570638)
		return s.IsSecondaryIndex == true
	}() == true {
		__antithesis_instrumentation__.Notify(570639)
		colIDs := index.CollectKeyColumnIDs()
		colIDs.UnionWith(index.CollectSecondaryStoredColumnIDs())
		colIDs.UnionWith(index.CollectKeySuffixColumnIDs())
		for i := range s.FetchedColumns {
			__antithesis_instrumentation__.Notify(570640)
			if !colIDs.Contains(s.FetchedColumns[i].ColumnID) {
				__antithesis_instrumentation__.Notify(570641)
				return errors.AssertionFailedf("requested column %s not in index", s.FetchedColumns[i].Name)
			} else {
				__antithesis_instrumentation__.Notify(570642)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(570643)
	}
	__antithesis_instrumentation__.Notify(570622)

	return nil
}
