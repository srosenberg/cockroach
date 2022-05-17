// Package schematestutils is a utility package for constructing schema objects
// in the context of cdc.
package schematestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/gogo/protobuf/proto"
)

func MakeTableDesc(
	tableID descpb.ID,
	version descpb.DescriptorVersion,
	modTime hlc.Timestamp,
	cols int,
	primaryKeyIndex int,
) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17956)
	td := descpb.TableDescriptor{
		Name:             "foo",
		ID:               tableID,
		Version:          version,
		ModificationTime: modTime,
		NextColumnID:     1,
		PrimaryIndex: descpb.IndexDescriptor{
			ID: descpb.IndexID(primaryKeyIndex),
		},
	}
	for i := 0; i < cols; i++ {
		__antithesis_instrumentation__.Notify(17958)
		td.Columns = append(td.Columns, *MakeColumnDesc(td.NextColumnID))
		td.NextColumnID++
	}
	__antithesis_instrumentation__.Notify(17957)
	return tabledesc.NewBuilder(&td).BuildImmutableTable()
}

func MakeColumnDesc(id descpb.ColumnID) *descpb.ColumnDescriptor {
	__antithesis_instrumentation__.Notify(17959)
	return &descpb.ColumnDescriptor{
		Name:        "c" + strconv.Itoa(int(id)),
		ID:          id,
		Type:        types.Bool,
		DefaultExpr: proto.String("true"),
	}
}

func SetLocalityRegionalByRow(desc catalog.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17960)
	desc.TableDesc().LocalityConfig = &catpb.LocalityConfig{
		Locality: &catpb.LocalityConfig_RegionalByRow_{
			RegionalByRow: &catpb.LocalityConfig_RegionalByRow{},
		},
	}
	return tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable()
}

func AddColumnDropBackfillMutation(desc catalog.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17961)
	desc.TableDesc().Mutations = append(desc.TableDesc().Mutations, descpb.DescriptorMutation{
		State:       descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		Direction:   descpb.DescriptorMutation_DROP,
		Descriptor_: &descpb.DescriptorMutation_Column{Column: MakeColumnDesc(desc.GetNextColumnID() - 1)},
	})
	return tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable()
}

func AddNewColumnBackfillMutation(desc catalog.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17962)
	desc.TableDesc().Mutations = append(desc.TableDesc().Mutations, descpb.DescriptorMutation{
		Descriptor_: &descpb.DescriptorMutation_Column{Column: MakeColumnDesc(desc.GetNextColumnID())},
		State:       descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		Direction:   descpb.DescriptorMutation_ADD,
		MutationID:  0,
		Rollback:    false,
	})
	return tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable()
}

func AddPrimaryKeySwapMutation(desc catalog.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17963)
	desc.TableDesc().Mutations = append(desc.TableDesc().Mutations, descpb.DescriptorMutation{
		State:       descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		Direction:   descpb.DescriptorMutation_ADD,
		Descriptor_: &descpb.DescriptorMutation_PrimaryKeySwap{PrimaryKeySwap: &descpb.PrimaryKeySwap{}},
	})
	return tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable()
}

func AddNewIndexMutation(desc catalog.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17964)
	desc.TableDesc().Mutations = append(desc.TableDesc().Mutations, descpb.DescriptorMutation{
		State:       descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		Direction:   descpb.DescriptorMutation_ADD,
		Descriptor_: &descpb.DescriptorMutation_Index{Index: &descpb.IndexDescriptor{}},
	})
	return tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable()
}

func AddDropIndexMutation(desc catalog.TableDescriptor) catalog.TableDescriptor {
	__antithesis_instrumentation__.Notify(17965)
	desc.TableDesc().Mutations = append(desc.TableDesc().Mutations, descpb.DescriptorMutation{
		State:       descpb.DescriptorMutation_DELETE_AND_WRITE_ONLY,
		Direction:   descpb.DescriptorMutation_DROP,
		Descriptor_: &descpb.DescriptorMutation_Index{Index: &descpb.IndexDescriptor{}},
	})
	return tabledesc.NewBuilder(desc.TableDesc()).BuildImmutableTable()
}
