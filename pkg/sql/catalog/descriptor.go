package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type DescriptorType string

const (
	Any DescriptorType = "any"

	Database = "database"

	Table = "relation"

	Type = "type"

	Schema = "schema"
)

type MutationPublicationFilter int

const (
	IgnoreConstraints MutationPublicationFilter = 1

	IgnoreConstraintsAndPKSwaps = 2

	IncludeConstraints = 0
)

type DescriptorBuilder interface {
	DescriptorType() DescriptorType

	RunPostDeserializationChanges() error

	RunRestoreChanges(descLookupFn func(id descpb.ID) Descriptor) error

	BuildImmutable() Descriptor

	BuildExistingMutable() MutableDescriptor

	BuildCreatedMutable() MutableDescriptor
}

type IndexOpts struct {
	NonPhysicalPrimaryIndex bool

	DropMutations bool

	AddMutations bool
}

type NameKey interface {
	GetName() string
	GetParentID() descpb.ID
	GetParentSchemaID() descpb.ID
}

type NameEntry interface {
	NameKey
	GetID() descpb.ID
}

var _ NameKey = descpb.NameInfo{}

type LeasableDescriptor interface {
	IsUncommittedVersion() bool

	GetVersion() descpb.DescriptorVersion

	GetModificationTime() hlc.Timestamp

	GetDrainingNames() []descpb.NameInfo
}

type Descriptor interface {
	NameEntry
	LeasableDescriptor

	GetPrivileges() *catpb.PrivilegeDescriptor

	DescriptorType() DescriptorType

	GetAuditMode() descpb.TableDescriptor_AuditMode

	Public() bool

	Adding() bool

	Dropped() bool

	Offline() bool

	GetOfflineReason() string

	DescriptorProto() *descpb.Descriptor

	ByteSize() int64

	NewBuilder() DescriptorBuilder

	GetReferencedDescIDs() (DescriptorIDSet, error)

	ValidateSelf(vea ValidationErrorAccumulator)

	ValidateCrossReferences(vea ValidationErrorAccumulator, vdg ValidationDescGetter)

	ValidateTxnCommit(vea ValidationErrorAccumulator, vdg ValidationDescGetter)

	GetDeclarativeSchemaChangerState() *scpb.DescriptorState

	GetPostDeserializationChanges() PostDeserializationChanges

	HasConcurrentSchemaChanges() bool
}

type DatabaseDescriptor interface {
	Descriptor

	DatabaseDesc() *descpb.DatabaseDescriptor

	GetRegionConfig() *descpb.DatabaseDescriptor_RegionConfig

	IsMultiRegion() bool

	PrimaryRegionName() (catpb.RegionName, error)

	MultiRegionEnumID() (descpb.ID, error)

	ForEachSchemaInfo(func(id descpb.ID, name string, isDropped bool) error) error

	ForEachNonDroppedSchema(func(id descpb.ID, name string) error) error

	GetSchemaID(name string) descpb.ID

	GetNonDroppedSchemaName(schemaID descpb.ID) string

	GetDefaultPrivilegeDescriptor() DefaultPrivilegeDescriptor

	HasPublicSchemaWithDescriptor() bool
}

type TableDescriptor interface {
	Descriptor

	TableDesc() *descpb.TableDescriptor

	GetState() descpb.DescriptorState

	IsTable() bool

	IsView() bool

	IsSequence() bool

	IsTemporary() bool

	IsVirtualTable() bool

	IsPhysicalTable() bool

	MaterializedView() bool

	IsAs() bool

	GetSequenceOpts() *descpb.TableDescriptor_SequenceOpts

	GetCreateQuery() string

	GetCreateAsOfTime() hlc.Timestamp

	GetViewQuery() string

	GetLease() *descpb.TableDescriptor_SchemaChangeLease

	GetDropTime() int64

	GetFormatVersion() descpb.FormatVersion

	GetPrimaryIndexID() descpb.IndexID

	GetPrimaryIndex() Index

	IsPartitionAllBy() bool

	PrimaryIndexSpan(codec keys.SQLCodec) roachpb.Span

	IndexSpan(codec keys.SQLCodec, id descpb.IndexID) roachpb.Span

	AllIndexSpans(codec keys.SQLCodec) roachpb.Spans

	TableSpan(codec keys.SQLCodec) roachpb.Span

	GetIndexMutationCapabilities(id descpb.IndexID) (isMutation, isWriteOnly bool)

	AllIndexes() []Index

	ActiveIndexes() []Index

	NonDropIndexes() []Index

	PartialIndexes() []Index

	PublicNonPrimaryIndexes() []Index

	WritableNonPrimaryIndexes() []Index

	DeletableNonPrimaryIndexes() []Index

	NonPrimaryIndexes() []Index

	DeleteOnlyNonPrimaryIndexes() []Index

	FindIndexWithID(id descpb.IndexID) (Index, error)

	FindIndexWithName(name string) (Index, error)

	GetNextIndexID() descpb.IndexID

	HasPrimaryKey() bool

	AllColumns() []Column

	PublicColumns() []Column

	WritableColumns() []Column

	DeletableColumns() []Column

	NonDropColumns() []Column

	VisibleColumns() []Column

	AccessibleColumns() []Column

	ReadableColumns() []Column

	UserDefinedTypeColumns() []Column

	SystemColumns() []Column

	PublicColumnIDs() []descpb.ColumnID

	IndexColumns(idx Index) []Column

	IndexKeyColumns(idx Index) []Column

	IndexKeyColumnDirections(idx Index) []descpb.IndexDescriptor_Direction

	IndexKeySuffixColumns(idx Index) []Column

	IndexFullColumns(idx Index) []Column

	IndexFullColumnDirections(idx Index) []descpb.IndexDescriptor_Direction

	IndexStoredColumns(idx Index) []Column

	IndexKeysPerRow(idx Index) int

	IndexFetchSpecKeyAndSuffixColumns(idx Index) []descpb.IndexFetchSpec_KeyColumn

	FindColumnWithID(id descpb.ColumnID) (Column, error)

	FindColumnWithName(name tree.Name) (Column, error)

	NamesForColumnIDs(ids descpb.ColumnIDs) ([]string, error)

	ContainsUserDefinedTypes() bool

	GetNextColumnID() descpb.ColumnID

	GetNextConstraintID() descpb.ConstraintID

	CheckConstraintUsesColumn(cc *descpb.TableDescriptor_CheckConstraint, colID descpb.ColumnID) (bool, error)

	IsShardColumn(col Column) bool

	GetFamilies() []descpb.ColumnFamilyDescriptor

	NumFamilies() int

	FindFamilyByID(id descpb.FamilyID) (*descpb.ColumnFamilyDescriptor, error)

	ForeachFamily(f func(family *descpb.ColumnFamilyDescriptor) error) error

	GetNextFamilyID() descpb.FamilyID

	FamilyDefaultColumns() []descpb.IndexFetchSpec_FamilyDefaultColumn

	HasColumnBackfillMutation() bool

	MakeFirstMutationPublic(includeConstraints MutationPublicationFilter) (TableDescriptor, error)

	MakePublic() TableDescriptor

	AllMutations() []Mutation

	GetMutationJobs() []descpb.TableDescriptor_MutationJob

	GetReplacementOf() descpb.TableDescriptor_Replacement

	GetAllReferencedTypeIDs(
		databaseDesc DatabaseDescriptor, getType func(descpb.ID) (TypeDescriptor, error),
	) (referencedAnywhere, referencedInColumns descpb.IDs, _ error)

	ForeachDependedOnBy(f func(dep *descpb.TableDescriptor_Reference) error) error

	GetDependedOnBy() []descpb.TableDescriptor_Reference

	GetDependsOn() []descpb.ID

	GetDependsOnTypes() []descpb.ID

	GetConstraintInfoWithLookup(fn TableLookupFn) (map[string]descpb.ConstraintDetail, error)

	GetConstraintInfo() (map[string]descpb.ConstraintDetail, error)

	GetUniqueWithoutIndexConstraints() []descpb.UniqueWithoutIndexConstraint

	AllActiveAndInactiveUniqueWithoutIndexConstraints() []*descpb.UniqueWithoutIndexConstraint

	ForeachOutboundFK(f func(fk *descpb.ForeignKeyConstraint) error) error

	ForeachInboundFK(f func(fk *descpb.ForeignKeyConstraint) error) error

	AllActiveAndInactiveForeignKeys() []*descpb.ForeignKeyConstraint

	GetChecks() []*descpb.TableDescriptor_CheckConstraint

	AllActiveAndInactiveChecks() []*descpb.TableDescriptor_CheckConstraint

	ActiveChecks() []descpb.TableDescriptor_CheckConstraint

	GetLocalityConfig() *catpb.LocalityConfig

	IsLocalityRegionalByRow() bool

	IsLocalityRegionalByTable() bool

	IsLocalityGlobal() bool

	GetRegionalByTableRegion() (catpb.RegionName, error)

	GetRegionalByRowTableRegionColumnName() (tree.Name, error)

	GetRowLevelTTL() *catpb.RowLevelTTL

	HasRowLevelTTL() bool

	GetExcludeDataFromBackup() bool

	GetStorageParams(spaceBetweenEqual bool) []string
}

type TypeDescriptor interface {
	Descriptor

	TypeDesc() *descpb.TypeDescriptor

	HydrateTypeInfoWithName(ctx context.Context, typ *types.T, name *tree.TypeName, res TypeDescriptorResolver) error

	MakeTypesT(ctx context.Context, name *tree.TypeName, res TypeDescriptorResolver) (*types.T, error)

	HasPendingSchemaChanges() bool

	GetIDClosure() (map[descpb.ID]struct{}, error)

	IsCompatibleWith(other TypeDescriptor) error

	GetArrayTypeID() descpb.ID

	GetKind() descpb.TypeDescriptor_Kind

	PrimaryRegionName() (catpb.RegionName, error)

	RegionNames() (catpb.RegionNames, error)

	RegionNamesIncludingTransitioning() (catpb.RegionNames, error)

	RegionNamesForValidation() (catpb.RegionNames, error)

	TransitioningRegionNames() (catpb.RegionNames, error)

	SuperRegions() ([]descpb.SuperRegion, error)

	ZoneConfigExtensions() (descpb.ZoneConfigExtensions, error)

	NumEnumMembers() int

	GetMemberPhysicalRepresentation(enumMemberOrdinal int) []byte

	GetMemberLogicalRepresentation(enumMemberOrdinal int) string

	IsMemberReadOnly(enumMemberOrdinal int) bool

	NumReferencingDescriptors() int

	GetReferencingDescriptorID(refOrdinal int) descpb.ID
}

type TypeDescriptorResolver interface {
	GetTypeDescriptor(ctx context.Context, id descpb.ID) (tree.TypeName, TypeDescriptor, error)
}

type DefaultPrivilegeDescriptor interface {
	GetDefaultPrivilegesForRole(catpb.DefaultPrivilegesRole) (*catpb.DefaultPrivilegesForRole, bool)
	ForEachDefaultPrivilegeForRole(func(catpb.DefaultPrivilegesForRole) error) error
	GetDefaultPrivilegeDescriptorType() catpb.DefaultPrivilegeDescriptor_DefaultPrivilegeDescriptorType
}

func FilterDescriptorState(desc Descriptor, flags tree.CommonLookupFlags) error {
	__antithesis_instrumentation__.Notify(264020)
	switch {
	case desc.Dropped() && func() bool {
		__antithesis_instrumentation__.Notify(264026)
		return !flags.IncludeDropped == true
	}() == true:
		__antithesis_instrumentation__.Notify(264021)
		return NewInactiveDescriptorError(ErrDescriptorDropped)
	case desc.Offline() && func() bool {
		__antithesis_instrumentation__.Notify(264027)
		return !flags.IncludeOffline == true
	}() == true:
		__antithesis_instrumentation__.Notify(264022)
		err := errors.Errorf("%s %q is offline", desc.DescriptorType(), desc.GetName())
		if desc.GetOfflineReason() != "" {
			__antithesis_instrumentation__.Notify(264028)
			err = errors.Errorf("%s %q is offline: %s", desc.DescriptorType(), desc.GetName(), desc.GetOfflineReason())
		} else {
			__antithesis_instrumentation__.Notify(264029)
		}
		__antithesis_instrumentation__.Notify(264023)
		return NewInactiveDescriptorError(err)
	case desc.Adding():
		__antithesis_instrumentation__.Notify(264024)

		return pgerror.WithCandidateCode(newAddingTableError(desc.(TableDescriptor)),
			pgcode.ObjectNotInPrerequisiteState)
	default:
		__antithesis_instrumentation__.Notify(264025)
		return nil
	}
}

type TableLookupFn func(descpb.ID) (TableDescriptor, error)

type Descriptors []Descriptor

func (d Descriptors) Len() int { __antithesis_instrumentation__.Notify(264030); return len(d) }
func (d Descriptors) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(264031)
	return d[i].GetID() < d[j].GetID()
}
func (d Descriptors) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(264032)
	d[i], d[j] = d[j], d[i]
}

func FormatSafeDescriptorProperties(w *redact.StringBuilder, desc Descriptor) {
	__antithesis_instrumentation__.Notify(264033)
	w.Printf("ID: %d, Version: %d", desc.GetID(), desc.GetVersion())
	if desc.IsUncommittedVersion() {
		__antithesis_instrumentation__.Notify(264036)
		w.Printf(", IsUncommitted: true")
	} else {
		__antithesis_instrumentation__.Notify(264037)
	}
	__antithesis_instrumentation__.Notify(264034)
	w.Printf(", ModificationTime: %q", desc.GetModificationTime())
	if parentID := desc.GetParentID(); parentID != 0 {
		__antithesis_instrumentation__.Notify(264038)
		w.Printf(", ParentID: %d", parentID)
	} else {
		__antithesis_instrumentation__.Notify(264039)
	}
	__antithesis_instrumentation__.Notify(264035)
	if parentSchemaID := desc.GetParentSchemaID(); parentSchemaID != 0 {
		__antithesis_instrumentation__.Notify(264040)
		w.Printf(", ParentSchemaID: %d", parentSchemaID)
	} else {
		__antithesis_instrumentation__.Notify(264041)
	}
	{
		__antithesis_instrumentation__.Notify(264042)
		var state descpb.DescriptorState
		switch {
		case desc.Public():
			__antithesis_instrumentation__.Notify(264044)
			state = descpb.DescriptorState_PUBLIC
		case desc.Dropped():
			__antithesis_instrumentation__.Notify(264045)
			state = descpb.DescriptorState_DROP
		case desc.Adding():
			__antithesis_instrumentation__.Notify(264046)
			state = descpb.DescriptorState_ADD
		case desc.Offline():
			__antithesis_instrumentation__.Notify(264047)
			state = descpb.DescriptorState_OFFLINE
		default:
			__antithesis_instrumentation__.Notify(264048)
		}
		__antithesis_instrumentation__.Notify(264043)
		w.Printf(", State: %v", state)
		if offlineReason := desc.GetOfflineReason(); state == descpb.DescriptorState_OFFLINE && func() bool {
			__antithesis_instrumentation__.Notify(264049)
			return offlineReason != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(264050)
			w.Printf(", OfflineReason: %q", redact.Safe(offlineReason))
		} else {
			__antithesis_instrumentation__.Notify(264051)
		}
	}
}

func IsSystemDescriptor(desc Descriptor) bool {
	__antithesis_instrumentation__.Notify(264052)
	if desc.GetParentID() == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(264055)
		return true
	} else {
		__antithesis_instrumentation__.Notify(264056)
	}
	__antithesis_instrumentation__.Notify(264053)
	if desc.GetID() == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(264057)
		return true
	} else {
		__antithesis_instrumentation__.Notify(264058)
	}
	__antithesis_instrumentation__.Notify(264054)
	return false
}
