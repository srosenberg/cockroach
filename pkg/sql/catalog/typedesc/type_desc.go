// Package typedesc contains the concrete implementations of
// catalog.TypeDescriptor.
package typedesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/enum"
	"github.com/cockroachdb/cockroach/pkg/sql/oidext"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

var _ catalog.TypeDescriptor = (*immutable)(nil)
var _ catalog.TypeDescriptor = (*Mutable)(nil)
var _ catalog.MutableDescriptor = (*Mutable)(nil)

func MakeSimpleAlias(typ *types.T, parentSchemaID descpb.ID) catalog.TypeDescriptor {
	__antithesis_instrumentation__.Notify(271930)
	return NewBuilder(&descpb.TypeDescriptor{

		ParentID:       descpb.InvalidID,
		ParentSchemaID: parentSchemaID,
		Name:           typ.Name(),

		ID:    descpb.InvalidID,
		Kind:  descpb.TypeDescriptor_ALIAS,
		Alias: typ,
	}).BuildImmutableType()
}

type Mutable struct {
	immutable

	ClusterVersion *immutable
}

func (desc *Mutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(271931)
	return desc.IsNew() || func() bool {
		__antithesis_instrumentation__.Notify(271932)
		return desc.ClusterVersion.GetVersion() != desc.GetVersion() == true
	}() == true
}

type immutable struct {
	descpb.TypeDescriptor

	logicalReps     []string
	physicalReps    [][]byte
	readOnlyMembers []bool

	isUncommittedVersion bool

	changes catalog.PostDeserializationChanges
}

func UpdateCachedFieldsOnModifiedMutable(desc catalog.TypeDescriptor) (*Mutable, error) {
	__antithesis_instrumentation__.Notify(271933)
	mutable, ok := desc.(*Mutable)
	if !ok {
		__antithesis_instrumentation__.Notify(271935)
		return nil, errors.AssertionFailedf("type descriptor was not mutable")
	} else {
		__antithesis_instrumentation__.Notify(271936)
	}
	__antithesis_instrumentation__.Notify(271934)
	mutable.immutable = *mutable.ImmutableCopy().(*immutable)
	return mutable, nil
}

func TypeIDToOID(id descpb.ID) oid.Oid {
	__antithesis_instrumentation__.Notify(271937)
	return oid.Oid(id) + oidext.CockroachPredefinedOIDMax
}

func UserDefinedTypeOIDToID(oid oid.Oid) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(271938)
	if descpb.ID(oid) <= oidext.CockroachPredefinedOIDMax {
		__antithesis_instrumentation__.Notify(271940)
		return 0, errors.Newf("user-defined OID %d should be greater "+
			"than predefined Max: %d.", oid, oidext.CockroachPredefinedOIDMax)
	} else {
		__antithesis_instrumentation__.Notify(271941)
	}
	__antithesis_instrumentation__.Notify(271939)
	return descpb.ID(oid) - oidext.CockroachPredefinedOIDMax, nil
}

func GetUserDefinedTypeDescID(t *types.T) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(271942)
	return UserDefinedTypeOIDToID(t.Oid())
}

func GetUserDefinedArrayTypeDescID(t *types.T) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(271943)
	return UserDefinedTypeOIDToID(t.UserDefinedArrayOID())
}

func (desc *immutable) TypeDesc() *descpb.TypeDescriptor {
	__antithesis_instrumentation__.Notify(271944)
	return &desc.TypeDescriptor
}

func (desc *immutable) Public() bool {
	__antithesis_instrumentation__.Notify(271945)
	return desc.State == descpb.DescriptorState_PUBLIC
}

func (desc *immutable) Adding() bool {
	__antithesis_instrumentation__.Notify(271946)
	return false
}

func (desc *immutable) Offline() bool {
	__antithesis_instrumentation__.Notify(271947)
	return desc.State == descpb.DescriptorState_OFFLINE
}

func (desc *immutable) Dropped() bool {
	__antithesis_instrumentation__.Notify(271948)
	return desc.State == descpb.DescriptorState_DROP
}

func (desc *immutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(271949)
	return desc.isUncommittedVersion
}

func (desc *immutable) DescriptorProto() *descpb.Descriptor {
	__antithesis_instrumentation__.Notify(271950)
	return &descpb.Descriptor{
		Union: &descpb.Descriptor_Type{
			Type: &desc.TypeDescriptor,
		},
	}
}

func (desc *immutable) ByteSize() int64 {
	__antithesis_instrumentation__.Notify(271951)
	return int64(desc.Size())
}

func (desc *Mutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(271952)
	return newBuilder(desc.TypeDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *immutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(271953)
	return newBuilder(desc.TypeDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *immutable) PrimaryRegionName() (catpb.RegionName, error) {
	__antithesis_instrumentation__.Notify(271954)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271956)
		return "", errors.AssertionFailedf(
			"can not get primary region of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID())
	} else {
		__antithesis_instrumentation__.Notify(271957)
	}
	__antithesis_instrumentation__.Notify(271955)
	return desc.RegionConfig.PrimaryRegion, nil
}

func (desc *immutable) RegionNames() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271958)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271961)
		return nil, errors.AssertionFailedf(
			"can not get regions of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271962)
	}
	__antithesis_instrumentation__.Notify(271959)
	var regions catpb.RegionNames
	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(271963)
		if member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY {
			__antithesis_instrumentation__.Notify(271965)
			continue
		} else {
			__antithesis_instrumentation__.Notify(271966)
		}
		__antithesis_instrumentation__.Notify(271964)
		regions = append(regions, catpb.RegionName(member.LogicalRepresentation))
	}
	__antithesis_instrumentation__.Notify(271960)
	return regions, nil
}

func (desc *immutable) TransitioningRegionNames() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271967)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271970)
		return nil, errors.AssertionFailedf(
			"can not get regions of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271971)
	}
	__antithesis_instrumentation__.Notify(271968)
	var regions catpb.RegionNames
	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(271972)
		if member.Direction != descpb.TypeDescriptor_EnumMember_NONE {
			__antithesis_instrumentation__.Notify(271973)
			regions = append(regions, catpb.RegionName(member.LogicalRepresentation))
		} else {
			__antithesis_instrumentation__.Notify(271974)
		}
	}
	__antithesis_instrumentation__.Notify(271969)
	return regions, nil
}

func (desc *immutable) RegionNamesForValidation() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271975)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271978)
		return nil, errors.AssertionFailedf(
			"can not get regions of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271979)
	}
	__antithesis_instrumentation__.Notify(271976)
	var regions catpb.RegionNames
	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(271980)
		if member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY && func() bool {
			__antithesis_instrumentation__.Notify(271982)
			return member.Direction == descpb.TypeDescriptor_EnumMember_ADD == true
		}() == true {
			__antithesis_instrumentation__.Notify(271983)
			continue
		} else {
			__antithesis_instrumentation__.Notify(271984)
		}
		__antithesis_instrumentation__.Notify(271981)
		regions = append(regions, catpb.RegionName(member.LogicalRepresentation))
	}
	__antithesis_instrumentation__.Notify(271977)
	return regions, nil
}

func (desc *immutable) SuperRegions() ([]descpb.SuperRegion, error) {
	__antithesis_instrumentation__.Notify(271985)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271987)
		return nil, errors.AssertionFailedf(
			"can not get regions of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271988)
	}
	__antithesis_instrumentation__.Notify(271986)

	return desc.RegionConfig.SuperRegions, nil
}

func (desc *immutable) RegionNamesIncludingTransitioning() (catpb.RegionNames, error) {
	__antithesis_instrumentation__.Notify(271989)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271992)
		return nil, errors.AssertionFailedf(
			"can not get regions of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271993)
	}
	__antithesis_instrumentation__.Notify(271990)
	var regions catpb.RegionNames
	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(271994)
		regions = append(regions, catpb.RegionName(member.LogicalRepresentation))
	}
	__antithesis_instrumentation__.Notify(271991)
	return regions, nil
}

func (desc *immutable) ZoneConfigExtensions() (descpb.ZoneConfigExtensions, error) {
	__antithesis_instrumentation__.Notify(271995)
	if desc.Kind != descpb.TypeDescriptor_MULTIREGION_ENUM {
		__antithesis_instrumentation__.Notify(271997)
		return descpb.ZoneConfigExtensions{}, errors.AssertionFailedf(
			"can not get the zone config extensions of a non multi-region enum %q (%d)", desc.GetName(), desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(271998)
	}
	__antithesis_instrumentation__.Notify(271996)
	return desc.RegionConfig.ZoneConfigExtensions, nil
}

func (desc *Mutable) SetDrainingNames(names []descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(271999)
	desc.DrainingNames = names
}

func (desc *immutable) GetAuditMode() descpb.TableDescriptor_AuditMode {
	__antithesis_instrumentation__.Notify(272000)
	return descpb.TableDescriptor_DISABLED
}

func (desc *immutable) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(272001)
	return catalog.Type
}

func (desc *Mutable) MaybeIncrementVersion() {
	__antithesis_instrumentation__.Notify(272002)

	if desc.ClusterVersion == nil || func() bool {
		__antithesis_instrumentation__.Notify(272004)
		return desc.Version == desc.ClusterVersion.Version+1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(272005)
		return
	} else {
		__antithesis_instrumentation__.Notify(272006)
	}
	__antithesis_instrumentation__.Notify(272003)
	desc.Version++
	desc.ModificationTime = hlc.Timestamp{}
}

func (desc *Mutable) OriginalName() string {
	__antithesis_instrumentation__.Notify(272007)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(272009)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(272010)
	}
	__antithesis_instrumentation__.Notify(272008)
	return desc.ClusterVersion.Name
}

func (desc *Mutable) OriginalID() descpb.ID {
	__antithesis_instrumentation__.Notify(272011)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(272013)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(272014)
	}
	__antithesis_instrumentation__.Notify(272012)
	return desc.ClusterVersion.ID
}

func (desc *Mutable) OriginalVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(272015)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(272017)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(272018)
	}
	__antithesis_instrumentation__.Notify(272016)
	return desc.ClusterVersion.Version
}

func (desc *Mutable) ImmutableCopy() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(272019)
	return desc.NewBuilder().(TypeDescriptorBuilder).BuildImmutableType()
}

func (desc *Mutable) IsNew() bool {
	__antithesis_instrumentation__.Notify(272020)
	return desc.ClusterVersion == nil
}

func (desc *Mutable) SetPublic() {
	__antithesis_instrumentation__.Notify(272021)
	desc.State = descpb.DescriptorState_PUBLIC
	desc.OfflineReason = ""
}

func (desc *Mutable) SetDropped() {
	__antithesis_instrumentation__.Notify(272022)
	desc.State = descpb.DescriptorState_DROP
	desc.OfflineReason = ""
}

func (desc *Mutable) SetOffline(reason string) {
	__antithesis_instrumentation__.Notify(272023)
	desc.State = descpb.DescriptorState_OFFLINE
	desc.OfflineReason = reason
}

func (desc *Mutable) DropEnumValue(value tree.EnumValue) {
	__antithesis_instrumentation__.Notify(272024)
	for i := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(272025)
		member := &desc.EnumMembers[i]
		if member.LogicalRepresentation == string(value) {
			__antithesis_instrumentation__.Notify(272026)
			member.Capability = descpb.TypeDescriptor_EnumMember_READ_ONLY
			member.Direction = descpb.TypeDescriptor_EnumMember_REMOVE
			break
		} else {
			__antithesis_instrumentation__.Notify(272027)
		}
	}
}

func (desc *Mutable) AddEnumValue(node *tree.AlterTypeAddValue) error {
	__antithesis_instrumentation__.Notify(272028)
	getPhysicalRep := func(idx int) []byte {
		__antithesis_instrumentation__.Notify(272032)
		if idx < 0 || func() bool {
			__antithesis_instrumentation__.Notify(272034)
			return idx >= len(desc.EnumMembers) == true
		}() == true {
			__antithesis_instrumentation__.Notify(272035)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(272036)
		}
		__antithesis_instrumentation__.Notify(272033)
		return desc.EnumMembers[idx].PhysicalRepresentation
	}
	__antithesis_instrumentation__.Notify(272029)

	pos := len(desc.EnumMembers) - 1
	if node.Placement != nil {
		__antithesis_instrumentation__.Notify(272037)

		foundIndex := -1
		existing := string(node.Placement.ExistingVal)
		for i, member := range desc.EnumMembers {
			__antithesis_instrumentation__.Notify(272040)
			if member.LogicalRepresentation == existing {
				__antithesis_instrumentation__.Notify(272041)
				foundIndex = i
			} else {
				__antithesis_instrumentation__.Notify(272042)
			}
		}
		__antithesis_instrumentation__.Notify(272038)
		if foundIndex == -1 {
			__antithesis_instrumentation__.Notify(272043)
			return pgerror.Newf(pgcode.InvalidParameterValue, "%q is not an existing enum value", existing)
		} else {
			__antithesis_instrumentation__.Notify(272044)
		}
		__antithesis_instrumentation__.Notify(272039)

		pos = foundIndex

		if node.Placement.Before {
			__antithesis_instrumentation__.Notify(272045)
			pos--
		} else {
			__antithesis_instrumentation__.Notify(272046)
		}
	} else {
		__antithesis_instrumentation__.Notify(272047)
	}
	__antithesis_instrumentation__.Notify(272030)

	newPhysicalRep := enum.GenByteStringBetween(getPhysicalRep(pos), getPhysicalRep(pos+1), enum.SpreadSpacing)
	newMember := descpb.TypeDescriptor_EnumMember{
		LogicalRepresentation:  string(node.NewVal),
		PhysicalRepresentation: newPhysicalRep,
		Capability:             descpb.TypeDescriptor_EnumMember_READ_ONLY,
		Direction:              descpb.TypeDescriptor_EnumMember_ADD,
	}

	if len(desc.EnumMembers) == 0 {
		__antithesis_instrumentation__.Notify(272048)
		desc.EnumMembers = []descpb.TypeDescriptor_EnumMember{newMember}
	} else {
		__antithesis_instrumentation__.Notify(272049)
		if pos < 0 {
			__antithesis_instrumentation__.Notify(272050)

			desc.EnumMembers = append([]descpb.TypeDescriptor_EnumMember{newMember}, desc.EnumMembers...)
		} else {
			__antithesis_instrumentation__.Notify(272051)

			desc.EnumMembers = append(desc.EnumMembers, descpb.TypeDescriptor_EnumMember{})
			copy(desc.EnumMembers[pos+2:], desc.EnumMembers[pos+1:])
			desc.EnumMembers[pos+1] = newMember
		}
	}
	__antithesis_instrumentation__.Notify(272031)
	return nil
}

func (desc *Mutable) AddReferencingDescriptorID(new descpb.ID) {
	__antithesis_instrumentation__.Notify(272052)
	for _, id := range desc.ReferencingDescriptorIDs {
		__antithesis_instrumentation__.Notify(272054)
		if new == id {
			__antithesis_instrumentation__.Notify(272055)
			return
		} else {
			__antithesis_instrumentation__.Notify(272056)
		}
	}
	__antithesis_instrumentation__.Notify(272053)
	desc.ReferencingDescriptorIDs = append(desc.ReferencingDescriptorIDs, new)
}

func (desc *Mutable) RemoveReferencingDescriptorID(remove descpb.ID) {
	__antithesis_instrumentation__.Notify(272057)
	for i, id := range desc.ReferencingDescriptorIDs {
		__antithesis_instrumentation__.Notify(272058)
		if id == remove {
			__antithesis_instrumentation__.Notify(272059)
			desc.ReferencingDescriptorIDs = append(desc.ReferencingDescriptorIDs[:i], desc.ReferencingDescriptorIDs[i+1:]...)
			return
		} else {
			__antithesis_instrumentation__.Notify(272060)
		}
	}
}

func (desc *Mutable) SetParentSchemaID(schemaID descpb.ID) {
	__antithesis_instrumentation__.Notify(272061)
	desc.ParentSchemaID = schemaID
}

func (desc *Mutable) AddDrainingName(name descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(272062)
	desc.DrainingNames = append(desc.DrainingNames, name)
}

func (desc *Mutable) SetName(name string) {
	__antithesis_instrumentation__.Notify(272063)
	desc.Name = name
}

type EnumMembers []descpb.TypeDescriptor_EnumMember

func (e EnumMembers) Len() int { __antithesis_instrumentation__.Notify(272064); return len(e) }
func (e EnumMembers) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(272065)
	return bytes.Compare(e[i].PhysicalRepresentation, e[j].PhysicalRepresentation) < 0
}
func (e EnumMembers) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(272066)
	e[i], e[j] = e[j], e[i]
}

func (desc *immutable) ValidateSelf(vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(272067)

	vea.Report(catalog.ValidateName(desc.Name, "type"))
	if desc.GetID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(272072)
		vea.Report(errors.AssertionFailedf("invalid ID %d", desc.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(272073)
	}
	__antithesis_instrumentation__.Notify(272068)
	if desc.GetParentID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(272074)
		vea.Report(errors.AssertionFailedf("invalid parentID %d", desc.GetParentID()))
	} else {
		__antithesis_instrumentation__.Notify(272075)
	}
	__antithesis_instrumentation__.Notify(272069)
	if desc.GetParentSchemaID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(272076)
		vea.Report(errors.AssertionFailedf("invalid parent schema ID %d", desc.GetParentSchemaID()))
	} else {
		__antithesis_instrumentation__.Notify(272077)
	}
	__antithesis_instrumentation__.Notify(272070)

	if desc.Privileges == nil {
		__antithesis_instrumentation__.Notify(272078)
		vea.Report(errors.AssertionFailedf("privileges not set"))
	} else {
		__antithesis_instrumentation__.Notify(272079)
		if desc.Kind != descpb.TypeDescriptor_ALIAS {
			__antithesis_instrumentation__.Notify(272080)
			vea.Report(catprivilege.Validate(*desc.Privileges, desc, privilege.Type))
		} else {
			__antithesis_instrumentation__.Notify(272081)
		}
	}
	__antithesis_instrumentation__.Notify(272071)

	switch desc.Kind {
	case descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272082)

		if desc.RegionConfig == nil {
			__antithesis_instrumentation__.Notify(272091)
			vea.Report(errors.AssertionFailedf("no region config on %s type desc", desc.Kind.String()))
		} else {
			__antithesis_instrumentation__.Notify(272092)
		}
		__antithesis_instrumentation__.Notify(272083)
		if desc.validateEnumMembers(vea) {
			__antithesis_instrumentation__.Notify(272093)

			for i := 0; i < len(desc.EnumMembers)-1; i++ {
				__antithesis_instrumentation__.Notify(272094)
				if desc.EnumMembers[i].LogicalRepresentation > desc.EnumMembers[i+1].LogicalRepresentation {
					__antithesis_instrumentation__.Notify(272095)
					vea.Report(errors.AssertionFailedf(
						"multi-region enum is out of order %q > %q",
						desc.EnumMembers[i].LogicalRepresentation,
						desc.EnumMembers[i+1].LogicalRepresentation,
					))
				} else {
					__antithesis_instrumentation__.Notify(272096)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(272097)
		}
	case descpb.TypeDescriptor_ENUM:
		__antithesis_instrumentation__.Notify(272084)
		if desc.RegionConfig != nil {
			__antithesis_instrumentation__.Notify(272098)
			vea.Report(errors.AssertionFailedf("found region config on %s type desc", desc.Kind.String()))
		} else {
			__antithesis_instrumentation__.Notify(272099)
		}
		__antithesis_instrumentation__.Notify(272085)
		desc.validateEnumMembers(vea)
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(272086)
		if desc.RegionConfig != nil {
			__antithesis_instrumentation__.Notify(272100)
			vea.Report(errors.AssertionFailedf("found region config on %s type desc", desc.Kind.String()))
		} else {
			__antithesis_instrumentation__.Notify(272101)
		}
		__antithesis_instrumentation__.Notify(272087)
		if desc.Alias == nil {
			__antithesis_instrumentation__.Notify(272102)
			vea.Report(errors.AssertionFailedf("ALIAS type desc has nil alias type"))
		} else {
			__antithesis_instrumentation__.Notify(272103)
		}
		__antithesis_instrumentation__.Notify(272088)
		if desc.GetArrayTypeID() != descpb.InvalidID {
			__antithesis_instrumentation__.Notify(272104)
			vea.Report(errors.AssertionFailedf("ALIAS type desc has array type ID %d", desc.GetArrayTypeID()))
		} else {
			__antithesis_instrumentation__.Notify(272105)
		}
	case descpb.TypeDescriptor_TABLE_IMPLICIT_RECORD_TYPE:
		__antithesis_instrumentation__.Notify(272089)
		vea.Report(errors.AssertionFailedf("invalid type descriptor: kind %s should never be serialized or validated", desc.Kind.String()))
	default:
		__antithesis_instrumentation__.Notify(272090)
		vea.Report(errors.AssertionFailedf("invalid type descriptor kind %s", desc.Kind.String()))
	}
}

func (desc *immutable) validateEnumMembers(vea catalog.ValidationErrorAccumulator) (isSorted bool) {
	__antithesis_instrumentation__.Notify(272106)

	isSorted = sort.IsSorted(EnumMembers(desc.EnumMembers))
	if !isSorted {
		__antithesis_instrumentation__.Notify(272109)
		vea.Report(errors.AssertionFailedf("enum members are not sorted %v", desc.EnumMembers))
	} else {
		__antithesis_instrumentation__.Notify(272110)
	}
	__antithesis_instrumentation__.Notify(272107)

	physicalMap := make(map[string]struct{}, len(desc.EnumMembers))
	logicalMap := make(map[string]struct{}, len(desc.EnumMembers))
	for _, member := range desc.EnumMembers {
		__antithesis_instrumentation__.Notify(272111)

		_, duplicatePhysical := physicalMap[string(member.PhysicalRepresentation)]
		if duplicatePhysical {
			__antithesis_instrumentation__.Notify(272114)
			vea.Report(errors.AssertionFailedf("duplicate enum physical rep %v", member.PhysicalRepresentation))
		} else {
			__antithesis_instrumentation__.Notify(272115)
		}
		__antithesis_instrumentation__.Notify(272112)
		physicalMap[string(member.PhysicalRepresentation)] = struct{}{}

		_, duplicateLogical := logicalMap[member.LogicalRepresentation]
		if duplicateLogical {
			__antithesis_instrumentation__.Notify(272116)
			vea.Report(errors.AssertionFailedf("duplicate enum member %q", member.LogicalRepresentation))
		} else {
			__antithesis_instrumentation__.Notify(272117)
		}
		__antithesis_instrumentation__.Notify(272113)
		logicalMap[member.LogicalRepresentation] = struct{}{}

		switch member.Capability {
		case descpb.TypeDescriptor_EnumMember_READ_ONLY:
			__antithesis_instrumentation__.Notify(272118)
			if member.Direction == descpb.TypeDescriptor_EnumMember_NONE {
				__antithesis_instrumentation__.Notify(272121)
				vea.Report(errors.AssertionFailedf(
					"read only capability member must have transition direction set"))
			} else {
				__antithesis_instrumentation__.Notify(272122)
			}
		case descpb.TypeDescriptor_EnumMember_ALL:
			__antithesis_instrumentation__.Notify(272119)
			if member.Direction != descpb.TypeDescriptor_EnumMember_NONE {
				__antithesis_instrumentation__.Notify(272123)
				vea.Report(errors.AssertionFailedf("public enum member can not have transition direction set"))
			} else {
				__antithesis_instrumentation__.Notify(272124)
			}
		default:
			__antithesis_instrumentation__.Notify(272120)
			vea.Report(errors.AssertionFailedf("invalid member capability %s", member.Capability))
		}
	}
	__antithesis_instrumentation__.Notify(272108)
	return isSorted
}

func (desc *immutable) GetReferencedDescIDs() (catalog.DescriptorIDSet, error) {
	__antithesis_instrumentation__.Notify(272125)
	ids := catalog.MakeDescriptorIDSet(desc.GetReferencingDescriptorIDs()...)
	ids.Add(desc.GetParentID())

	if desc.GetParentSchemaID() != keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(272129)
		ids.Add(desc.GetParentSchemaID())
	} else {
		__antithesis_instrumentation__.Notify(272130)
	}
	__antithesis_instrumentation__.Notify(272126)
	children, err := desc.GetIDClosure()
	if err != nil {
		__antithesis_instrumentation__.Notify(272131)
		return catalog.DescriptorIDSet{}, err
	} else {
		__antithesis_instrumentation__.Notify(272132)
	}
	__antithesis_instrumentation__.Notify(272127)
	for id := range children {
		__antithesis_instrumentation__.Notify(272133)
		ids.Add(id)
	}
	__antithesis_instrumentation__.Notify(272128)
	return ids, nil
}

func (desc *immutable) ValidateCrossReferences(
	vea catalog.ValidationErrorAccumulator, vdg catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(272134)

	dbDesc, err := vdg.GetDatabaseDescriptor(desc.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(272139)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(272140)
		if dbDesc.Dropped() {
			__antithesis_instrumentation__.Notify(272141)
			vea.Report(errors.AssertionFailedf("parent database %q (%d) is dropped",
				dbDesc.GetName(), dbDesc.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(272142)
		}
	}
	__antithesis_instrumentation__.Notify(272135)

	if desc.GetParentSchemaID() != keys.PublicSchemaID {
		__antithesis_instrumentation__.Notify(272143)
		schemaDesc, err := vdg.GetSchemaDescriptor(desc.GetParentSchemaID())
		vea.Report(err)
		if schemaDesc != nil && func() bool {
			__antithesis_instrumentation__.Notify(272145)
			return dbDesc != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(272146)
			return schemaDesc.GetParentID() != dbDesc.GetID() == true
		}() == true {
			__antithesis_instrumentation__.Notify(272147)
			vea.Report(errors.AssertionFailedf("parent schema %d is in different database %d",
				desc.GetParentSchemaID(), schemaDesc.GetParentID()))
		} else {
			__antithesis_instrumentation__.Notify(272148)
		}
		__antithesis_instrumentation__.Notify(272144)
		if schemaDesc != nil && func() bool {
			__antithesis_instrumentation__.Notify(272149)
			return schemaDesc.Dropped() == true
		}() == true {
			__antithesis_instrumentation__.Notify(272150)
			vea.Report(errors.AssertionFailedf("parent schema %q (%d) is dropped",
				schemaDesc.GetName(), schemaDesc.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(272151)
		}
	} else {
		__antithesis_instrumentation__.Notify(272152)
	}
	__antithesis_instrumentation__.Notify(272136)

	if desc.GetKind() == descpb.TypeDescriptor_MULTIREGION_ENUM && func() bool {
		__antithesis_instrumentation__.Notify(272153)
		return dbDesc != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(272154)
		desc.validateMultiRegion(dbDesc, vea)
	} else {
		__antithesis_instrumentation__.Notify(272155)
	}
	__antithesis_instrumentation__.Notify(272137)

	switch desc.GetKind() {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272156)

		if typ, err := vdg.GetTypeDescriptor(desc.GetArrayTypeID()); err != nil {
			__antithesis_instrumentation__.Notify(272159)
			vea.Report(errors.Wrapf(err, "arrayTypeID %d does not exist for %q", desc.GetArrayTypeID(), desc.GetKind()))
		} else {
			__antithesis_instrumentation__.Notify(272160)
			if typ.Dropped() {
				__antithesis_instrumentation__.Notify(272161)
				vea.Report(errors.AssertionFailedf("array type %q (%d) is dropped", typ.GetName(), typ.GetID()))
			} else {
				__antithesis_instrumentation__.Notify(272162)
			}
		}
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(272157)
		if desc.GetAlias().UserDefined() {
			__antithesis_instrumentation__.Notify(272163)
			aliasedID, err := UserDefinedTypeOIDToID(desc.GetAlias().Oid())
			if err != nil {
				__antithesis_instrumentation__.Notify(272165)
				vea.Report(err)
			} else {
				__antithesis_instrumentation__.Notify(272166)
			}
			__antithesis_instrumentation__.Notify(272164)
			if typ, err := vdg.GetTypeDescriptor(aliasedID); err != nil {
				__antithesis_instrumentation__.Notify(272167)
				vea.Report(errors.Wrapf(err, "aliased type %d does not exist", aliasedID))
			} else {
				__antithesis_instrumentation__.Notify(272168)
				if typ.Dropped() {
					__antithesis_instrumentation__.Notify(272169)
					vea.Report(errors.AssertionFailedf("aliased type %q (%d) is dropped", typ.GetName(), typ.GetID()))
				} else {
					__antithesis_instrumentation__.Notify(272170)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(272171)
		}
	default:
		__antithesis_instrumentation__.Notify(272158)
	}
	__antithesis_instrumentation__.Notify(272138)

	for _, id := range desc.GetReferencingDescriptorIDs() {
		__antithesis_instrumentation__.Notify(272172)
		tableDesc, err := vdg.GetTableDescriptor(id)
		if err != nil {
			__antithesis_instrumentation__.Notify(272174)
			vea.Report(err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(272175)
		}
		__antithesis_instrumentation__.Notify(272173)
		if tableDesc.Dropped() {
			__antithesis_instrumentation__.Notify(272176)
			vea.Report(errors.AssertionFailedf(
				"referencing table %d was dropped without dependency unlinking", id))
		} else {
			__antithesis_instrumentation__.Notify(272177)
		}
	}
}

func (desc *immutable) validateMultiRegion(
	dbDesc catalog.DatabaseDescriptor, vea catalog.ValidationErrorAccumulator,
) {
	__antithesis_instrumentation__.Notify(272178)

	if !dbDesc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(272188)
		vea.Report(errors.AssertionFailedf("parent database is not a multi-region database"))
		return
	} else {
		__antithesis_instrumentation__.Notify(272189)
	}
	__antithesis_instrumentation__.Notify(272179)

	primaryRegion, err := desc.PrimaryRegionName()
	if err != nil {
		__antithesis_instrumentation__.Notify(272190)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(272191)
	}

	{
		__antithesis_instrumentation__.Notify(272192)
		found := false
		for _, member := range desc.EnumMembers {
			__antithesis_instrumentation__.Notify(272194)
			if catpb.RegionName(member.LogicalRepresentation) == primaryRegion {
				__antithesis_instrumentation__.Notify(272195)
				found = true
			} else {
				__antithesis_instrumentation__.Notify(272196)
			}
		}
		__antithesis_instrumentation__.Notify(272193)
		if !found {
			__antithesis_instrumentation__.Notify(272197)
			vea.Report(
				errors.AssertionFailedf("primary region %q not found in list of enum members",
					primaryRegion,
				))
		} else {
			__antithesis_instrumentation__.Notify(272198)
		}
	}
	__antithesis_instrumentation__.Notify(272180)

	dbPrimaryRegion, err := dbDesc.PrimaryRegionName()
	if err != nil {
		__antithesis_instrumentation__.Notify(272199)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(272200)
	}
	__antithesis_instrumentation__.Notify(272181)
	if dbPrimaryRegion != primaryRegion {
		__antithesis_instrumentation__.Notify(272201)
		vea.Report(errors.AssertionFailedf("unexpected primary region on db desc: %q expected %q",
			dbPrimaryRegion, primaryRegion))
	} else {
		__antithesis_instrumentation__.Notify(272202)
	}
	__antithesis_instrumentation__.Notify(272182)

	regionNames, err := desc.RegionNames()
	if err != nil {
		__antithesis_instrumentation__.Notify(272203)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(272204)
	}
	__antithesis_instrumentation__.Notify(272183)
	if dbDesc.GetRegionConfig().SurvivalGoal == descpb.SurvivalGoal_REGION_FAILURE {
		__antithesis_instrumentation__.Notify(272205)
		if len(regionNames) < 3 {
			__antithesis_instrumentation__.Notify(272206)
			vea.Report(
				errors.AssertionFailedf(
					"expected >= 3 regions, got %d: %s",
					len(regionNames),
					strings.Join(regionNames.ToStrings(), ","),
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(272207)
		}
	} else {
		__antithesis_instrumentation__.Notify(272208)
	}
	__antithesis_instrumentation__.Notify(272184)

	superRegions, err := desc.SuperRegions()
	if err != nil {
		__antithesis_instrumentation__.Notify(272209)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(272210)
	}
	__antithesis_instrumentation__.Notify(272185)
	multiregion.ValidateSuperRegions(superRegions, dbDesc.GetRegionConfig().SurvivalGoal, regionNames, func(err error) {
		__antithesis_instrumentation__.Notify(272211)
		vea.Report(err)
	})
	__antithesis_instrumentation__.Notify(272186)

	zoneCfgExtensions, err := desc.ZoneConfigExtensions()
	if err != nil {
		__antithesis_instrumentation__.Notify(272212)
		vea.Report(err)
	} else {
		__antithesis_instrumentation__.Notify(272213)
	}
	__antithesis_instrumentation__.Notify(272187)
	multiregion.ValidateZoneConfigExtensions(regionNames, zoneCfgExtensions, func(err error) {
		__antithesis_instrumentation__.Notify(272214)
		vea.Report(err)
	})
}

func (desc *immutable) ValidateTxnCommit(
	_ catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(272215)

}

type TypeLookupFunc func(ctx context.Context, id descpb.ID) (tree.TypeName, catalog.TypeDescriptor, error)

func (t TypeLookupFunc) GetTypeDescriptor(
	ctx context.Context, id descpb.ID,
) (tree.TypeName, catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(272216)
	return t(ctx, id)
}

func (desc *immutable) MakeTypesT(
	ctx context.Context, name *tree.TypeName, res catalog.TypeDescriptorResolver,
) (*types.T, error) {
	__antithesis_instrumentation__.Notify(272217)
	switch t := desc.Kind; t {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272218)
		typ := types.MakeEnum(TypeIDToOID(desc.GetID()), TypeIDToOID(desc.ArrayTypeID))
		if err := desc.HydrateTypeInfoWithName(ctx, typ, name, res); err != nil {
			__antithesis_instrumentation__.Notify(272223)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(272224)
		}
		__antithesis_instrumentation__.Notify(272219)
		return typ, nil
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(272220)

		if err := desc.HydrateTypeInfoWithName(ctx, desc.Alias, name, res); err != nil {
			__antithesis_instrumentation__.Notify(272225)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(272226)
		}
		__antithesis_instrumentation__.Notify(272221)
		return desc.Alias, nil
	default:
		__antithesis_instrumentation__.Notify(272222)
		return nil, errors.AssertionFailedf("unknown type kind %s", t.String())
	}
}

func EnsureTypeIsHydrated(
	ctx context.Context, t *types.T, res catalog.TypeDescriptorResolver,
) error {
	__antithesis_instrumentation__.Notify(272227)
	if t.Family() == types.TupleFamily {
		__antithesis_instrumentation__.Notify(272232)
		for _, typ := range t.TupleContents() {
			__antithesis_instrumentation__.Notify(272234)
			if err := EnsureTypeIsHydrated(ctx, typ, res); err != nil {
				__antithesis_instrumentation__.Notify(272235)
				return err
			} else {
				__antithesis_instrumentation__.Notify(272236)
			}
		}
		__antithesis_instrumentation__.Notify(272233)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272237)
	}
	__antithesis_instrumentation__.Notify(272228)
	if !t.UserDefined() || func() bool {
		__antithesis_instrumentation__.Notify(272238)
		return t.IsHydrated() == true
	}() == true {
		__antithesis_instrumentation__.Notify(272239)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272240)
	}
	__antithesis_instrumentation__.Notify(272229)
	id, err := GetUserDefinedTypeDescID(t)
	if err != nil {
		__antithesis_instrumentation__.Notify(272241)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272242)
	}
	__antithesis_instrumentation__.Notify(272230)
	elemTypName, elemTypDesc, err := res.GetTypeDescriptor(ctx, id)
	if err != nil {
		__antithesis_instrumentation__.Notify(272243)
		return err
	} else {
		__antithesis_instrumentation__.Notify(272244)
	}
	__antithesis_instrumentation__.Notify(272231)
	return elemTypDesc.HydrateTypeInfoWithName(ctx, t, &elemTypName, res)
}

func HydrateTypesInTableDescriptor(
	ctx context.Context, desc *descpb.TableDescriptor, res catalog.TypeDescriptorResolver,
) error {
	__antithesis_instrumentation__.Notify(272245)
	for i := range desc.Columns {
		__antithesis_instrumentation__.Notify(272248)
		if err := EnsureTypeIsHydrated(ctx, desc.Columns[i].Type, res); err != nil {
			__antithesis_instrumentation__.Notify(272249)
			return err
		} else {
			__antithesis_instrumentation__.Notify(272250)
		}
	}
	__antithesis_instrumentation__.Notify(272246)
	for i := range desc.Mutations {
		__antithesis_instrumentation__.Notify(272251)
		mut := &desc.Mutations[i]
		if col := mut.GetColumn(); col != nil {
			__antithesis_instrumentation__.Notify(272252)
			if err := EnsureTypeIsHydrated(ctx, col.Type, res); err != nil {
				__antithesis_instrumentation__.Notify(272253)
				return err
			} else {
				__antithesis_instrumentation__.Notify(272254)
			}
		} else {
			__antithesis_instrumentation__.Notify(272255)
		}
	}
	__antithesis_instrumentation__.Notify(272247)
	return nil
}

func (desc *immutable) HydrateTypeInfoWithName(
	ctx context.Context, typ *types.T, name *tree.TypeName, res catalog.TypeDescriptorResolver,
) error {
	__antithesis_instrumentation__.Notify(272256)
	if typ.IsHydrated() {
		__antithesis_instrumentation__.Notify(272258)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(272259)
	}
	__antithesis_instrumentation__.Notify(272257)
	typ.TypeMeta.Name = &types.UserDefinedTypeName{
		Catalog:        name.Catalog(),
		ExplicitSchema: name.ExplicitSchema,
		Schema:         name.Schema(),
		Name:           name.Object(),
	}
	typ.TypeMeta.Version = uint32(desc.Version)
	switch desc.Kind {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272260)
		if typ.Family() != types.EnumFamily {
			__antithesis_instrumentation__.Notify(272265)
			return errors.New("cannot hydrate a non-enum type with an enum type descriptor")
		} else {
			__antithesis_instrumentation__.Notify(272266)
		}
		__antithesis_instrumentation__.Notify(272261)
		typ.TypeMeta.EnumData = &types.EnumMetadata{
			LogicalRepresentations:  desc.logicalReps,
			PhysicalRepresentations: desc.physicalReps,
			IsMemberReadOnly:        desc.readOnlyMembers,
		}
		return nil
	case descpb.TypeDescriptor_ALIAS:
		__antithesis_instrumentation__.Notify(272262)
		if typ.UserDefined() {
			__antithesis_instrumentation__.Notify(272267)
			switch typ.Family() {
			case types.ArrayFamily:
				__antithesis_instrumentation__.Notify(272268)

				elemType := typ.ArrayContents()
				return EnsureTypeIsHydrated(ctx, elemType, res)
			case types.TupleFamily:
				__antithesis_instrumentation__.Notify(272269)
				return EnsureTypeIsHydrated(ctx, typ, res)
			default:
				__antithesis_instrumentation__.Notify(272270)
				return errors.AssertionFailedf("unhandled alias type family %s", typ.Family())
			}
		} else {
			__antithesis_instrumentation__.Notify(272271)
		}
		__antithesis_instrumentation__.Notify(272263)
		return nil
	default:
		__antithesis_instrumentation__.Notify(272264)
		return errors.AssertionFailedf("unknown type descriptor kind %s", desc.Kind)
	}
}

func (desc *immutable) NumEnumMembers() int {
	__antithesis_instrumentation__.Notify(272272)
	return len(desc.EnumMembers)
}

func (desc *immutable) GetMemberPhysicalRepresentation(enumMemberOrdinal int) []byte {
	__antithesis_instrumentation__.Notify(272273)
	return desc.physicalReps[enumMemberOrdinal]
}

func (desc *immutable) GetMemberLogicalRepresentation(enumMemberOrdinal int) string {
	__antithesis_instrumentation__.Notify(272274)
	return desc.logicalReps[enumMemberOrdinal]
}

func (desc *immutable) IsMemberReadOnly(enumMemberOrdinal int) bool {
	__antithesis_instrumentation__.Notify(272275)
	return desc.readOnlyMembers[enumMemberOrdinal]
}

func (desc *immutable) NumReferencingDescriptors() int {
	__antithesis_instrumentation__.Notify(272276)
	return len(desc.ReferencingDescriptorIDs)
}

func (desc *immutable) GetReferencingDescriptorID(refOrdinal int) descpb.ID {
	__antithesis_instrumentation__.Notify(272277)
	return desc.ReferencingDescriptorIDs[refOrdinal]
}

func (desc *immutable) IsCompatibleWith(other catalog.TypeDescriptor) error {
	__antithesis_instrumentation__.Notify(272278)

	switch desc.Kind {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272279)
		if other.GetKind() != desc.Kind {
			__antithesis_instrumentation__.Notify(272283)
			return errors.Newf("%q of type %q is not compatible with type %q",
				other.GetName(), other.GetKind(), desc.Kind)
		} else {
			__antithesis_instrumentation__.Notify(272284)
		}
		__antithesis_instrumentation__.Notify(272280)

		for _, thisMember := range desc.EnumMembers {
			__antithesis_instrumentation__.Notify(272285)
			found := false
			for i := 0; i < other.NumEnumMembers(); i++ {
				__antithesis_instrumentation__.Notify(272287)
				if thisMember.LogicalRepresentation == other.GetMemberLogicalRepresentation(i) {
					__antithesis_instrumentation__.Notify(272288)

					if !bytes.Equal(thisMember.PhysicalRepresentation, other.GetMemberPhysicalRepresentation(i)) {
						__antithesis_instrumentation__.Notify(272290)
						return errors.Newf(
							"%q has differing physical representation for value %q",
							other.GetName(),
							thisMember.LogicalRepresentation,
						)
					} else {
						__antithesis_instrumentation__.Notify(272291)
					}
					__antithesis_instrumentation__.Notify(272289)
					found = true
				} else {
					__antithesis_instrumentation__.Notify(272292)
				}
			}
			__antithesis_instrumentation__.Notify(272286)
			if !found {
				__antithesis_instrumentation__.Notify(272293)
				return errors.Newf(
					"could not find enum value %q in %q", thisMember.LogicalRepresentation, other.GetName())
			} else {
				__antithesis_instrumentation__.Notify(272294)
			}
		}
		__antithesis_instrumentation__.Notify(272281)
		return nil
	default:
		__antithesis_instrumentation__.Notify(272282)
		return errors.Newf("compatibility comparison unsupported for type kind %s", desc.Kind.String())
	}
}

func (desc *immutable) HasPendingSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(272295)
	switch desc.Kind {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272296)

		for i := range desc.EnumMembers {
			__antithesis_instrumentation__.Notify(272299)
			if desc.EnumMembers[i].Capability != descpb.TypeDescriptor_EnumMember_ALL {
				__antithesis_instrumentation__.Notify(272300)
				return true
			} else {
				__antithesis_instrumentation__.Notify(272301)
			}
		}
		__antithesis_instrumentation__.Notify(272297)
		return false
	default:
		__antithesis_instrumentation__.Notify(272298)
		return false
	}
}

func (desc *immutable) GetPostDeserializationChanges() catalog.PostDeserializationChanges {
	__antithesis_instrumentation__.Notify(272302)
	return desc.changes
}

func (desc *immutable) HasConcurrentSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(272303)
	if desc.DeclarativeSchemaChangerState != nil && func() bool {
		__antithesis_instrumentation__.Notify(272305)
		return desc.DeclarativeSchemaChangerState.JobID != catpb.InvalidJobID == true
	}() == true {
		__antithesis_instrumentation__.Notify(272306)
		return true
	} else {
		__antithesis_instrumentation__.Notify(272307)
	}
	__antithesis_instrumentation__.Notify(272304)

	return false
}

func (desc *immutable) GetIDClosure() (map[descpb.ID]struct{}, error) {
	__antithesis_instrumentation__.Notify(272308)
	ret := make(map[descpb.ID]struct{})

	ret[desc.ID] = struct{}{}
	if desc.Kind == descpb.TypeDescriptor_ALIAS {
		__antithesis_instrumentation__.Notify(272310)

		children, err := GetTypeDescriptorClosure(desc.Alias)
		if err != nil {
			__antithesis_instrumentation__.Notify(272312)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(272313)
		}
		__antithesis_instrumentation__.Notify(272311)
		for id := range children {
			__antithesis_instrumentation__.Notify(272314)
			ret[id] = struct{}{}
		}
	} else {
		__antithesis_instrumentation__.Notify(272315)

		ret[desc.ArrayTypeID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(272309)
	return ret, nil
}

func GetTypeDescriptorClosure(typ *types.T) (map[descpb.ID]struct{}, error) {
	__antithesis_instrumentation__.Notify(272316)
	if !typ.UserDefined() {
		__antithesis_instrumentation__.Notify(272320)
		return map[descpb.ID]struct{}{}, nil
	} else {
		__antithesis_instrumentation__.Notify(272321)
	}
	__antithesis_instrumentation__.Notify(272317)
	id, err := GetUserDefinedTypeDescID(typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(272322)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(272323)
	}
	__antithesis_instrumentation__.Notify(272318)

	ret := map[descpb.ID]struct{}{
		id: {},
	}
	switch typ.Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(272324)

		children, err := GetTypeDescriptorClosure(typ.ArrayContents())
		if err != nil {
			__antithesis_instrumentation__.Notify(272329)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(272330)
		}
		__antithesis_instrumentation__.Notify(272325)
		for id := range children {
			__antithesis_instrumentation__.Notify(272331)
			ret[id] = struct{}{}
		}
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(272326)

		for _, elt := range typ.TupleContents() {
			__antithesis_instrumentation__.Notify(272332)
			children, err := GetTypeDescriptorClosure(elt)
			if err != nil {
				__antithesis_instrumentation__.Notify(272334)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(272335)
			}
			__antithesis_instrumentation__.Notify(272333)
			for id := range children {
				__antithesis_instrumentation__.Notify(272336)
				ret[id] = struct{}{}
			}
		}
	default:
		__antithesis_instrumentation__.Notify(272327)

		id, err := GetUserDefinedArrayTypeDescID(typ)
		if err != nil {
			__antithesis_instrumentation__.Notify(272337)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(272338)
		}
		__antithesis_instrumentation__.Notify(272328)
		ret[id] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(272319)
	return ret, nil
}

func (desc *Mutable) SetDeclarativeSchemaChangerState(state *scpb.DescriptorState) {
	__antithesis_instrumentation__.Notify(272339)
	desc.DeclarativeSchemaChangerState = state
}
