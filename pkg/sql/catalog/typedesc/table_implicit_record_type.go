package typedesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type TableImplicitRecordType struct {
	desc catalog.TableDescriptor

	typ *types.T

	privs *catpb.PrivilegeDescriptor
}

var _ catalog.TypeDescriptor = (*TableImplicitRecordType)(nil)

func CreateImplicitRecordTypeFromTableDesc(
	descriptor catalog.TableDescriptor,
) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(271856)

	cols := descriptor.VisibleColumns()
	typs := make([]*types.T, len(cols))
	names := make([]string, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(271859)
		if col.GetType().UserDefined() && func() bool {
			__antithesis_instrumentation__.Notify(271861)
			return !col.GetType().IsHydrated() == true
		}() == true {
			__antithesis_instrumentation__.Notify(271862)
			return nil, errors.AssertionFailedf("encountered unhydrated col %s while creating implicit record type from"+
				" table %s", col.ColName(), descriptor.GetName())
		} else {
			__antithesis_instrumentation__.Notify(271863)
		}
		__antithesis_instrumentation__.Notify(271860)
		typs[i] = col.GetType()
		names[i] = col.GetName()
	}
	__antithesis_instrumentation__.Notify(271857)

	typ := types.MakeLabeledTuple(typs, names)
	tableID := descriptor.GetID()
	typeOID := TypeIDToOID(tableID)

	typ.InternalType.Oid = typeOID
	typ.TypeMeta = types.UserDefinedTypeMetadata{
		Name: &types.UserDefinedTypeName{
			Name: descriptor.GetName(),
		},
		Version: uint32(descriptor.GetVersion()),
	}

	tablePrivs := descriptor.GetPrivileges()
	newPrivs := make([]catpb.UserPrivileges, len(tablePrivs.Users))
	for i := range tablePrivs.Users {
		__antithesis_instrumentation__.Notify(271864)
		newPrivs[i].UserProto = tablePrivs.Users[i].UserProto

		if privilege.SELECT.IsSetIn(tablePrivs.Users[i].Privileges) {
			__antithesis_instrumentation__.Notify(271865)
			newPrivs[i].Privileges = privilege.USAGE.Mask()
		} else {
			__antithesis_instrumentation__.Notify(271866)
		}
	}
	__antithesis_instrumentation__.Notify(271858)

	return &TableImplicitRecordType{
		desc: descriptor,
		typ:  typ,
		privs: &catpb.PrivilegeDescriptor{
			Users:      newPrivs,
			OwnerProto: tablePrivs.OwnerProto,
			Version:    tablePrivs.Version,
		},
	}, nil
}

func (v TableImplicitRecordType) GetName() string {
	__antithesis_instrumentation__.Notify(271867)
	return v.desc.GetName()
}

func (v TableImplicitRecordType) GetParentID() descpb.ID {
	__antithesis_instrumentation__.Notify(271868)
	return v.desc.GetParentID()
}

func (v TableImplicitRecordType) GetParentSchemaID() descpb.ID {
	__antithesis_instrumentation__.Notify(271869)
	return v.desc.GetParentSchemaID()
}

func (v TableImplicitRecordType) GetID() descpb.ID {
	__antithesis_instrumentation__.Notify(271870)
	return v.desc.GetID()
}

func (v TableImplicitRecordType) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(271871)
	return v.desc.IsUncommittedVersion()
}

func (v TableImplicitRecordType) GetVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(271872)
	return v.desc.GetVersion()
}

func (v TableImplicitRecordType) GetModificationTime() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(271873)
	return v.desc.GetModificationTime()
}

func (v TableImplicitRecordType) GetDrainingNames() []descpb.NameInfo {
	__antithesis_instrumentation__.Notify(271874)

	return nil
}

func (v TableImplicitRecordType) GetPrivileges() *catpb.PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(271875)
	return v.privs
}

func (v TableImplicitRecordType) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(271876)
	return catalog.Type
}

func (v TableImplicitRecordType) GetAuditMode() descpb.TableDescriptor_AuditMode {
	__antithesis_instrumentation__.Notify(271877)
	return descpb.TableDescriptor_DISABLED
}

func (v TableImplicitRecordType) Public() bool {
	__antithesis_instrumentation__.Notify(271878)
	return v.desc.Public()
}

func (v TableImplicitRecordType) Adding() bool {
	__antithesis_instrumentation__.Notify(271879)
	v.panicNotSupported("Adding")
	return false
}

func (v TableImplicitRecordType) Dropped() bool {
	__antithesis_instrumentation__.Notify(271880)
	v.panicNotSupported("Dropped")
	return false
}

func (v TableImplicitRecordType) Offline() bool {
	__antithesis_instrumentation__.Notify(271881)
	v.panicNotSupported("Offline")
	return false
}

func (v TableImplicitRecordType) GetOfflineReason() string {
	__antithesis_instrumentation__.Notify(271882)
	v.panicNotSupported("GetOfflineReason")
	return ""
}

func (v TableImplicitRecordType) DescriptorProto() *descpb.Descriptor {
	__antithesis_instrumentation__.Notify(271883)
	v.panicNotSupported("DescriptorProto")
	return nil
}

func (v TableImplicitRecordType) ByteSize() int64 {
	__antithesis_instrumentation__.Notify(271884)
	mem := v.desc.ByteSize()
	if v.typ != nil {
		__antithesis_instrumentation__.Notify(271887)
		mem += int64(v.typ.Size())
	} else {
		__antithesis_instrumentation__.Notify(271888)
	}
	__antithesis_instrumentation__.Notify(271885)
	if v.privs != nil {
		__antithesis_instrumentation__.Notify(271889)
		mem += int64(v.privs.Size())
	} else {
		__antithesis_instrumentation__.Notify(271890)
	}
	__antithesis_instrumentation__.Notify(271886)
	return mem
}

func (v TableImplicitRecordType) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(271891)
	v.panicNotSupported("NewBuilder")
	return nil
}

func (v TableImplicitRecordType) GetReferencedDescIDs() (catalog.DescriptorIDSet, error) {
	__antithesis_instrumentation__.Notify(271892)
	return catalog.DescriptorIDSet{}, errors.AssertionFailedf(
		"GetReferencedDescIDs are unsupported for implicit table record types")
}

func (v TableImplicitRecordType) ValidateSelf(_ catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(271893)
}

func (v TableImplicitRecordType) ValidateCrossReferences(
	_ catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(271894)
}

func (v TableImplicitRecordType) ValidateTxnCommit(
	_ catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(271895)
}

func (v TableImplicitRecordType) TypeDesc() *descpb.TypeDescriptor {
	__antithesis_instrumentation__.Notify(271896)
	v.panicNotSupported("TypeDesc")
	return nil
}

func (v TableImplicitRecordType) HydrateTypeInfoWithName(
	ctx context.Context, typ *types.T, name *tree.TypeName, res catalog.TypeDescriptorResolver,
) error {
	__antithesis_instrumentation__.Notify(271897)
	if typ.IsHydrated() {
		__antithesis_instrumentation__.Notify(271901)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(271902)
	}
	__antithesis_instrumentation__.Notify(271898)
	if typ.Family() != types.TupleFamily {
		__antithesis_instrumentation__.Notify(271903)
		return errors.AssertionFailedf("unexpected hydration of non-tuple type %s with table implicit record type %d",
			typ, v.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271904)
	}
	__antithesis_instrumentation__.Notify(271899)
	if typ.Oid() != TypeIDToOID(v.GetID()) {
		__antithesis_instrumentation__.Notify(271905)
		return errors.AssertionFailedf("unexpected mismatch during table implicit record type hydration: "+
			"type %s has id %d, descriptor has id %d", typ, typ.Oid(), v.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271906)
	}
	__antithesis_instrumentation__.Notify(271900)
	typ.TypeMeta.Name = &types.UserDefinedTypeName{
		Catalog:        name.Catalog(),
		ExplicitSchema: name.ExplicitSchema,
		Schema:         name.Schema(),
		Name:           name.Object(),
	}
	typ.TypeMeta.Version = uint32(v.desc.GetVersion())
	return EnsureTypeIsHydrated(ctx, typ, res)
}

func (v TableImplicitRecordType) MakeTypesT(
	_ context.Context, _ *tree.TypeName, _ catalog.TypeDescriptorResolver,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(271907)
	return v.typ, nil
}

func (v TableImplicitRecordType) HasPendingSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(271908)
	return false
}

func (v TableImplicitRecordType) GetIDClosure() (map[descpb.ID]struct{}, error) {
	__antithesis_instrumentation__.Notify(271909)
	return nil, errors.AssertionFailedf("IDClosure unsupported for implicit table record types")
}

func (v TableImplicitRecordType) IsCompatibleWith(_ catalog.TypeDescriptor) error {
	__antithesis_instrumentation__.Notify(271910)
	return errors.AssertionFailedf("compatibility comparison unsupported for implicit table record types")
}

func (v TableImplicitRecordType) PrimaryRegionName() (catpb.RegionName, error) {
	__antithesis_instrumentation__.Notify(271911)
	return "", errors.AssertionFailedf(
		"can not get primary region of a implicit table record type")
}

func (v TableImplicitRecordType) RegionNames() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271912)
	return nil, errors.AssertionFailedf(
		"can not get region names of a implicit table record type")
}

func (v TableImplicitRecordType) RegionNamesIncludingTransitioning() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271913)
	return nil, errors.AssertionFailedf(
		"can not get region names of a implicit table record type")
}

func (v TableImplicitRecordType) RegionNamesForValidation() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271914)
	return nil, errors.AssertionFailedf(
		"can not get region names of a implicit table record type")
}

func (v TableImplicitRecordType) TransitioningRegionNames() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271915)
	return nil, errors.AssertionFailedf(
		"can not get region names of a implicit table record type")
}

func (v TableImplicitRecordType) SuperRegions() ([]descpb.SuperRegion, error) {
	__antithesis_instrumentation__.Notify(271916)
	return nil, errors.AssertionFailedf(
		"can not get super regions of a implicit table record type",
	)
}

func (v TableImplicitRecordType) ZoneConfigExtensions() (descpb.ZoneConfigExtensions, error) {
	__antithesis_instrumentation__.Notify(271917)
	return descpb.ZoneConfigExtensions{}, errors.AssertionFailedf(
		"can not get the zone config extensions of a implicit table record type")
}

func (v TableImplicitRecordType) GetArrayTypeID() descpb.ID {
	__antithesis_instrumentation__.Notify(271918)
	return 0
}

func (v TableImplicitRecordType) GetKind() descpb.TypeDescriptor_Kind {
	__antithesis_instrumentation__.Notify(271919)
	return descpb.TypeDescriptor_TABLE_IMPLICIT_RECORD_TYPE
}

func (v TableImplicitRecordType) NumEnumMembers() int {
	__antithesis_instrumentation__.Notify(271920)
	return 0
}

func (v TableImplicitRecordType) GetMemberPhysicalRepresentation(_ int) []byte {
	__antithesis_instrumentation__.Notify(271921)
	return nil
}

func (v TableImplicitRecordType) GetMemberLogicalRepresentation(_ int) string {
	__antithesis_instrumentation__.Notify(271922)
	return ""
}

func (v TableImplicitRecordType) IsMemberReadOnly(_ int) bool {
	__antithesis_instrumentation__.Notify(271923)
	return false
}

func (v TableImplicitRecordType) NumReferencingDescriptors() int {
	__antithesis_instrumentation__.Notify(271924)
	return 0
}

func (v TableImplicitRecordType) GetReferencingDescriptorID(_ int) descpb.ID {
	__antithesis_instrumentation__.Notify(271925)
	return 0
}

func (v TableImplicitRecordType) GetPostDeserializationChanges() catalog.PostDeserializationChanges {
	__antithesis_instrumentation__.Notify(271926)
	return catalog.PostDeserializationChanges{}
}

func (v TableImplicitRecordType) HasConcurrentSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(271927)
	return false
}

func (v TableImplicitRecordType) panicNotSupported(message string) {
	__antithesis_instrumentation__.Notify(271928)
	panic(errors.AssertionFailedf("implicit table record type for table %q: not supported: %s", v.GetName(), message))
}

func (v TableImplicitRecordType) GetDeclarativeSchemaChangerState() *scpb.DescriptorState {
	__antithesis_instrumentation__.Notify(271929)
	v.panicNotSupported("GetDeclarativeSchemaChangeState")
	return nil
}
