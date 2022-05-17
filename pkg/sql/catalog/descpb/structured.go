package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

func (dir IndexDescriptor_Direction) ToEncodingDirection() (encoding.Direction, error) {
	__antithesis_instrumentation__.Notify(252530)
	switch dir {
	case IndexDescriptor_ASC:
		__antithesis_instrumentation__.Notify(252531)
		return encoding.Ascending, nil
	case IndexDescriptor_DESC:
		__antithesis_instrumentation__.Notify(252532)
		return encoding.Descending, nil
	default:
		__antithesis_instrumentation__.Notify(252533)
		return encoding.Ascending, errors.Errorf("invalid direction: %s", dir)
	}
}

type ID = catid.DescID

const InvalidID = catid.InvalidDescID

type IDs []ID

func (ids IDs) Len() int { __antithesis_instrumentation__.Notify(252534); return len(ids) }
func (ids IDs) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(252535)
	return ids[i] < ids[j]
}
func (ids IDs) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(252536)
	ids[i], ids[j] = ids[j], ids[i]
}

type FormatVersion uint32

const (
	_ FormatVersion = iota

	BaseFormatVersion

	FamilyFormatVersion

	InterleavedFormatVersion
)

type FamilyID = catid.FamilyID

type IndexID = catid.IndexID

type ConstraintID = catid.ConstraintID

type DescriptorVersion uint64

func (DescriptorVersion) SafeValue() { __antithesis_instrumentation__.Notify(252537) }

type IndexDescriptorVersion uint32

func (IndexDescriptorVersion) SafeValue() { __antithesis_instrumentation__.Notify(252538) }

const (
	BaseIndexFormatVersion IndexDescriptorVersion = iota

	SecondaryIndexFamilyFormatVersion

	EmptyArraysInInvertedIndexesVersion

	StrictIndexColumnIDGuaranteesVersion

	PrimaryIndexWithStoredColumnsVersion

	LatestIndexDescriptorVersion = PrimaryIndexWithStoredColumnsVersion
)

type ColumnID = catid.ColumnID

type ColumnIDs []ColumnID

func (c ColumnIDs) Len() int { __antithesis_instrumentation__.Notify(252539); return len(c) }
func (c ColumnIDs) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(252540)
	c[i], c[j] = c[j], c[i]
}
func (c ColumnIDs) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(252541)
	return c[i] < c[j]
}

func (c ColumnIDs) HasPrefix(input ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(252542)
	if len(input) > len(c) {
		__antithesis_instrumentation__.Notify(252545)
		return false
	} else {
		__antithesis_instrumentation__.Notify(252546)
	}
	__antithesis_instrumentation__.Notify(252543)
	for i := range input {
		__antithesis_instrumentation__.Notify(252547)
		if input[i] != c[i] {
			__antithesis_instrumentation__.Notify(252548)
			return false
		} else {
			__antithesis_instrumentation__.Notify(252549)
		}
	}
	__antithesis_instrumentation__.Notify(252544)
	return true
}

func (c ColumnIDs) Equals(input ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(252550)
	if len(input) != len(c) {
		__antithesis_instrumentation__.Notify(252553)
		return false
	} else {
		__antithesis_instrumentation__.Notify(252554)
	}
	__antithesis_instrumentation__.Notify(252551)
	for i := range input {
		__antithesis_instrumentation__.Notify(252555)
		if input[i] != c[i] {
			__antithesis_instrumentation__.Notify(252556)
			return false
		} else {
			__antithesis_instrumentation__.Notify(252557)
		}
	}
	__antithesis_instrumentation__.Notify(252552)
	return true
}

func (c ColumnIDs) PermutationOf(input ColumnIDs) bool {
	__antithesis_instrumentation__.Notify(252558)
	ourColsSet := util.MakeFastIntSet()
	for _, col := range c {
		__antithesis_instrumentation__.Notify(252561)
		ourColsSet.Add(int(col))
	}
	__antithesis_instrumentation__.Notify(252559)

	inputColsSet := util.MakeFastIntSet()
	for _, inputCol := range input {
		__antithesis_instrumentation__.Notify(252562)
		inputColsSet.Add(int(inputCol))
	}
	__antithesis_instrumentation__.Notify(252560)

	return inputColsSet.Equals(ourColsSet)
}

func (c ColumnIDs) Contains(i ColumnID) bool {
	__antithesis_instrumentation__.Notify(252563)
	for _, id := range c {
		__antithesis_instrumentation__.Notify(252565)
		if i == id {
			__antithesis_instrumentation__.Notify(252566)
			return true
		} else {
			__antithesis_instrumentation__.Notify(252567)
		}
	}
	__antithesis_instrumentation__.Notify(252564)
	return false
}

type IndexDescriptorEncodingType uint32

const (
	SecondaryIndexEncoding IndexDescriptorEncodingType = iota

	PrimaryIndexEncoding
)

var _ = SecondaryIndexEncoding

type MutationID uint32

func (MutationID) SafeValue() { __antithesis_instrumentation__.Notify(252568) }

const InvalidMutationID MutationID = 0

func (f ForeignKeyReference) IsSet() bool {
	__antithesis_instrumentation__.Notify(252569)
	return f.Table != 0
}

func (desc *TableDescriptor) Public() bool {
	__antithesis_instrumentation__.Notify(252570)
	return desc.State == DescriptorState_PUBLIC
}

func (desc *TableDescriptor) Offline() bool {
	__antithesis_instrumentation__.Notify(252571)
	return desc.State == DescriptorState_OFFLINE
}

func (desc *TableDescriptor) Dropped() bool {
	__antithesis_instrumentation__.Notify(252572)
	return desc.State == DescriptorState_DROP
}

func (desc *TableDescriptor) Adding() bool {
	__antithesis_instrumentation__.Notify(252573)
	return desc.State == DescriptorState_ADD
}

func (desc *TableDescriptor) IsTable() bool {
	__antithesis_instrumentation__.Notify(252574)
	return !desc.IsView() && func() bool {
		__antithesis_instrumentation__.Notify(252575)
		return !desc.IsSequence() == true
	}() == true
}

func (desc *TableDescriptor) IsView() bool {
	__antithesis_instrumentation__.Notify(252576)
	return desc.ViewQuery != ""
}

func (desc *TableDescriptor) MaterializedView() bool {
	__antithesis_instrumentation__.Notify(252577)
	return desc.IsMaterializedView
}

func (desc *TableDescriptor) IsPhysicalTable() bool {
	__antithesis_instrumentation__.Notify(252578)
	return desc.IsSequence() || func() bool {
		__antithesis_instrumentation__.Notify(252579)
		return (desc.IsTable() && func() bool {
			__antithesis_instrumentation__.Notify(252580)
			return !desc.IsVirtualTable() == true
		}() == true) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(252581)
		return desc.MaterializedView() == true
	}() == true
}

func (desc *TableDescriptor) IsAs() bool {
	__antithesis_instrumentation__.Notify(252582)
	return desc.CreateQuery != ""
}

func (desc *TableDescriptor) IsSequence() bool {
	__antithesis_instrumentation__.Notify(252583)
	return desc.SequenceOpts != nil
}

func (desc *TableDescriptor) IsVirtualTable() bool {
	__antithesis_instrumentation__.Notify(252584)
	return IsVirtualTable(desc.ID)
}

func (desc *TableDescriptor) Persistence() tree.Persistence {
	__antithesis_instrumentation__.Notify(252585)
	if desc.Temporary {
		__antithesis_instrumentation__.Notify(252587)
		return tree.PersistenceTemporary
	} else {
		__antithesis_instrumentation__.Notify(252588)
	}
	__antithesis_instrumentation__.Notify(252586)
	return tree.PersistencePermanent
}

func IsVirtualTable(id ID) bool {
	__antithesis_instrumentation__.Notify(252589)
	return catconstants.MinVirtualID <= id
}

func IsSystemConfigID(id ID) bool {
	__antithesis_instrumentation__.Notify(252590)
	return id > 0 && func() bool {
		__antithesis_instrumentation__.Notify(252591)
		return id <= keys.MaxSystemConfigDescID == true
	}() == true
}

var AnonymousTable = tree.TableName{}

func (opts *TableDescriptor_SequenceOpts) HasOwner() bool {
	__antithesis_instrumentation__.Notify(252592)
	return !opts.SequenceOwner.Equal(TableDescriptor_SequenceOpts_SequenceOwner{})
}

func (opts *TableDescriptor_SequenceOpts) EffectiveCacheSize() int64 {
	__antithesis_instrumentation__.Notify(252593)
	if opts.CacheSize == 0 {
		__antithesis_instrumentation__.Notify(252595)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(252596)
	}
	__antithesis_instrumentation__.Notify(252594)
	return opts.CacheSize
}

func (ConstraintValidity) SafeValue() { __antithesis_instrumentation__.Notify(252597) }

func (DescriptorMutation_Direction) SafeValue() { __antithesis_instrumentation__.Notify(252598) }

func (DescriptorMutation_State) SafeValue() { __antithesis_instrumentation__.Notify(252599) }

func (DescriptorState) SafeValue() { __antithesis_instrumentation__.Notify(252600) }

func (ConstraintType) SafeValue() { __antithesis_instrumentation__.Notify(252601) }

type UniqueConstraint interface {
	IsValidReferencedUniqueConstraint(referencedColIDs ColumnIDs) bool

	GetName() string
}

var _ UniqueConstraint = &UniqueWithoutIndexConstraint{}
var _ UniqueConstraint = &IndexDescriptor{}

func (u *UniqueWithoutIndexConstraint) IsValidReferencedUniqueConstraint(
	referencedColIDs ColumnIDs,
) bool {
	__antithesis_instrumentation__.Notify(252602)
	return ColumnIDs(u.ColumnIDs).PermutationOf(referencedColIDs)
}

func (u *UniqueWithoutIndexConstraint) GetName() string {
	__antithesis_instrumentation__.Notify(252603)
	return u.Name
}

func (u *UniqueWithoutIndexConstraint) IsPartial() bool {
	__antithesis_instrumentation__.Notify(252604)
	return u.Predicate != ""
}

func (ni NameInfo) GetParentID() ID {
	__antithesis_instrumentation__.Notify(252605)
	return ni.ParentID
}

func (ni NameInfo) GetParentSchemaID() ID {
	__antithesis_instrumentation__.Notify(252606)
	return ni.ParentSchemaID
}

func (ni NameInfo) GetName() string {
	__antithesis_instrumentation__.Notify(252607)
	return ni.Name
}

func init() {
	protoreflect.RegisterShorthands((*Descriptor)(nil), "descriptor", "desc")
}
