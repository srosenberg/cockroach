package schemadesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var _ catalog.SchemaDescriptor = (*immutable)(nil)
var _ catalog.SchemaDescriptor = (*Mutable)(nil)
var _ catalog.MutableDescriptor = (*Mutable)(nil)

type immutable struct {
	descpb.SchemaDescriptor

	isUncommittedVersion bool

	changes catalog.PostDeserializationChanges
}

func (desc *immutable) SchemaKind() catalog.ResolvedSchemaKind {
	__antithesis_instrumentation__.Notify(267616)
	return catalog.SchemaUserDefined
}

func (desc *immutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(267617)
	return formatSafeMessage("schemadesc.immutable", desc)
}

func (desc *Mutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(267618)
	return formatSafeMessage("schemadesc.Mutable", desc)
}

func formatSafeMessage(typeName string, desc catalog.SchemaDescriptor) string {
	__antithesis_instrumentation__.Notify(267619)
	var buf redact.StringBuilder
	buf.Printf(typeName + ": {")
	catalog.FormatSafeDescriptorProperties(&buf, desc)
	buf.Printf("}")
	return buf.String()
}

type Mutable struct {
	immutable

	ClusterVersion *immutable
}

var _ redact.SafeMessager = (*immutable)(nil)

func (desc *Mutable) SetDrainingNames(names []descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(267620)
	desc.DrainingNames = names
}

func (desc *Mutable) AddDrainingName(name descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(267621)
	desc.DrainingNames = append(desc.DrainingNames, name)
}

func (desc *immutable) GetParentSchemaID() descpb.ID {
	__antithesis_instrumentation__.Notify(267622)
	return keys.RootNamespaceID
}

func (desc *immutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(267623)
	return desc.isUncommittedVersion
}

func (desc *immutable) GetAuditMode() descpb.TableDescriptor_AuditMode {
	__antithesis_instrumentation__.Notify(267624)
	return descpb.TableDescriptor_DISABLED
}

func (desc *immutable) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(267625)
	return catalog.Schema
}

func (desc *immutable) SchemaDesc() *descpb.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(267626)
	return &desc.SchemaDescriptor
}

func (desc *immutable) Public() bool {
	__antithesis_instrumentation__.Notify(267627)
	return desc.State == descpb.DescriptorState_PUBLIC
}

func (desc *immutable) Adding() bool {
	__antithesis_instrumentation__.Notify(267628)
	return false
}

func (desc *immutable) Offline() bool {
	__antithesis_instrumentation__.Notify(267629)
	return desc.State == descpb.DescriptorState_OFFLINE
}

func (desc *immutable) Dropped() bool {
	__antithesis_instrumentation__.Notify(267630)
	return desc.State == descpb.DescriptorState_DROP
}

func (desc *immutable) DescriptorProto() *descpb.Descriptor {
	__antithesis_instrumentation__.Notify(267631)
	return &descpb.Descriptor{
		Union: &descpb.Descriptor_Schema{
			Schema: &desc.SchemaDescriptor,
		},
	}
}

func (desc *immutable) ByteSize() int64 {
	__antithesis_instrumentation__.Notify(267632)
	return int64(desc.Size())
}

func (desc *Mutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(267633)
	return newBuilder(desc.SchemaDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *immutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(267634)
	return newBuilder(desc.SchemaDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *immutable) ValidateSelf(vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(267635)

	vea.Report(catalog.ValidateName(desc.GetName(), "descriptor"))
	if desc.GetID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(267638)
		vea.Report(fmt.Errorf("invalid schema ID %d", desc.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(267639)
	}
	__antithesis_instrumentation__.Notify(267636)

	if desc.Privileges == nil {
		__antithesis_instrumentation__.Notify(267640)
		vea.Report(errors.AssertionFailedf("privileges not set"))
	} else {
		__antithesis_instrumentation__.Notify(267641)
		vea.Report(catprivilege.Validate(*desc.Privileges, desc, privilege.Schema))
	}
	__antithesis_instrumentation__.Notify(267637)

	if desc.GetDefaultPrivileges() != nil {
		__antithesis_instrumentation__.Notify(267642)

		vea.Report(catprivilege.ValidateDefaultPrivileges(*desc.GetDefaultPrivileges()))
	} else {
		__antithesis_instrumentation__.Notify(267643)
	}
}

func (desc *immutable) GetReferencedDescIDs() (catalog.DescriptorIDSet, error) {
	__antithesis_instrumentation__.Notify(267644)
	return catalog.MakeDescriptorIDSet(desc.GetID(), desc.GetParentID()), nil
}

func (desc *immutable) ValidateCrossReferences(
	vea catalog.ValidationErrorAccumulator, vdg catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(267645)

	db, err := vdg.GetDatabaseDescriptor(desc.GetParentID())
	if err != nil {
		__antithesis_instrumentation__.Notify(267649)
		vea.Report(err)
		return
	} else {
		__antithesis_instrumentation__.Notify(267650)
	}
	__antithesis_instrumentation__.Notify(267646)
	if db.Dropped() {
		__antithesis_instrumentation__.Notify(267651)
		vea.Report(errors.AssertionFailedf("parent database %q (%d) is dropped",
			db.GetName(), db.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(267652)
	}
	__antithesis_instrumentation__.Notify(267647)

	isInDBSchemas := false
	_ = db.ForEachSchemaInfo(func(id descpb.ID, name string, isDropped bool) error {
		__antithesis_instrumentation__.Notify(267653)
		if id == desc.GetID() {
			__antithesis_instrumentation__.Notify(267656)
			if isDropped {
				__antithesis_instrumentation__.Notify(267659)
				if name == desc.GetName() {
					__antithesis_instrumentation__.Notify(267661)
					vea.Report(errors.AssertionFailedf("present in parent database [%d] schemas mapping but marked as dropped",
						desc.GetParentID()))
				} else {
					__antithesis_instrumentation__.Notify(267662)
				}
				__antithesis_instrumentation__.Notify(267660)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(267663)
			}
			__antithesis_instrumentation__.Notify(267657)
			if name != desc.GetName() {
				__antithesis_instrumentation__.Notify(267664)
				vea.Report(errors.AssertionFailedf("present in parent database [%d] schemas mapping but under name %q",
					desc.GetParentID(), errors.Safe(name)))
				return nil
			} else {
				__antithesis_instrumentation__.Notify(267665)
			}
			__antithesis_instrumentation__.Notify(267658)
			isInDBSchemas = true
			return nil
		} else {
			__antithesis_instrumentation__.Notify(267666)
		}
		__antithesis_instrumentation__.Notify(267654)
		if name == desc.GetName() && func() bool {
			__antithesis_instrumentation__.Notify(267667)
			return !isDropped == true
		}() == true {
			__antithesis_instrumentation__.Notify(267668)
			vea.Report(errors.AssertionFailedf("present in parent database [%d] schemas mapping but name maps to other schema [%d]",
				desc.GetParentID(), id))
		} else {
			__antithesis_instrumentation__.Notify(267669)
		}
		__antithesis_instrumentation__.Notify(267655)
		return nil
	})
	__antithesis_instrumentation__.Notify(267648)
	if !isInDBSchemas {
		__antithesis_instrumentation__.Notify(267670)
		vea.Report(errors.AssertionFailedf("not present in parent database [%d] schemas mapping",
			desc.GetParentID()))
	} else {
		__antithesis_instrumentation__.Notify(267671)
	}
}

func (desc *immutable) ValidateTxnCommit(
	_ catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(267672)

}

func (desc *immutable) GetDefaultPrivilegeDescriptor() catalog.DefaultPrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(267673)
	defaultPrivilegeDescriptor := desc.GetDefaultPrivileges()
	if defaultPrivilegeDescriptor == nil {
		__antithesis_instrumentation__.Notify(267675)
		defaultPrivilegeDescriptor = catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_SCHEMA)
	} else {
		__antithesis_instrumentation__.Notify(267676)
	}
	__antithesis_instrumentation__.Notify(267674)
	return catprivilege.MakeDefaultPrivileges(defaultPrivilegeDescriptor)
}

func (desc *immutable) GetPostDeserializationChanges() catalog.PostDeserializationChanges {
	__antithesis_instrumentation__.Notify(267677)
	return desc.changes
}

func (desc *immutable) HasConcurrentSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(267678)
	return desc.DeclarativeSchemaChangerState != nil && func() bool {
		__antithesis_instrumentation__.Notify(267679)
		return desc.DeclarativeSchemaChangerState.JobID != catpb.InvalidJobID == true
	}() == true
}

func (desc *Mutable) MaybeIncrementVersion() {
	__antithesis_instrumentation__.Notify(267680)

	if desc.ClusterVersion == nil || func() bool {
		__antithesis_instrumentation__.Notify(267682)
		return desc.Version == desc.ClusterVersion.Version+1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(267683)
		return
	} else {
		__antithesis_instrumentation__.Notify(267684)
	}
	__antithesis_instrumentation__.Notify(267681)
	desc.Version++
	desc.ModificationTime = hlc.Timestamp{}
}

func (desc *Mutable) OriginalName() string {
	__antithesis_instrumentation__.Notify(267685)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(267687)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(267688)
	}
	__antithesis_instrumentation__.Notify(267686)
	return desc.ClusterVersion.Name
}

func (desc *Mutable) OriginalID() descpb.ID {
	__antithesis_instrumentation__.Notify(267689)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(267691)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(267692)
	}
	__antithesis_instrumentation__.Notify(267690)
	return desc.ClusterVersion.ID
}

func (desc *Mutable) OriginalVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(267693)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(267695)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(267696)
	}
	__antithesis_instrumentation__.Notify(267694)
	return desc.ClusterVersion.Version
}

func (desc *Mutable) ImmutableCopy() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(267697)
	return desc.NewBuilder().BuildImmutable()
}

func (desc *Mutable) IsNew() bool {
	__antithesis_instrumentation__.Notify(267698)
	return desc.ClusterVersion == nil
}

func (desc *Mutable) SetPublic() {
	__antithesis_instrumentation__.Notify(267699)
	desc.State = descpb.DescriptorState_PUBLIC
	desc.OfflineReason = ""
}

func (desc *Mutable) SetDropped() {
	__antithesis_instrumentation__.Notify(267700)
	desc.State = descpb.DescriptorState_DROP
	desc.OfflineReason = ""
}

func (desc *Mutable) SetOffline(reason string) {
	__antithesis_instrumentation__.Notify(267701)
	desc.State = descpb.DescriptorState_OFFLINE
	desc.OfflineReason = reason
}

func (desc *Mutable) SetName(name string) {
	__antithesis_instrumentation__.Notify(267702)
	desc.Name = name
}

func (desc *Mutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(267703)
	return desc.IsNew() || func() bool {
		__antithesis_instrumentation__.Notify(267704)
		return desc.GetVersion() != desc.ClusterVersion.GetVersion() == true
	}() == true
}

func (desc *Mutable) GetMutableDefaultPrivilegeDescriptor() *catprivilege.Mutable {
	__antithesis_instrumentation__.Notify(267705)
	defaultPrivilegeDescriptor := desc.GetDefaultPrivileges()
	if defaultPrivilegeDescriptor == nil {
		__antithesis_instrumentation__.Notify(267707)
		defaultPrivilegeDescriptor = catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_SCHEMA)
	} else {
		__antithesis_instrumentation__.Notify(267708)
	}
	__antithesis_instrumentation__.Notify(267706)
	return catprivilege.NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
}

func (desc *Mutable) SetDefaultPrivilegeDescriptor(
	defaultPrivilegeDescriptor *catpb.DefaultPrivilegeDescriptor,
) {
	__antithesis_instrumentation__.Notify(267709)
	desc.DefaultPrivileges = defaultPrivilegeDescriptor
}

func (desc *immutable) GetDeclarativeSchemaChangeState() *scpb.DescriptorState {
	__antithesis_instrumentation__.Notify(267710)
	return desc.DeclarativeSchemaChangerState.Clone()
}

func (desc *Mutable) SetDeclarativeSchemaChangerState(state *scpb.DescriptorState) {
	__antithesis_instrumentation__.Notify(267711)
	desc.DeclarativeSchemaChangerState = state
}

func IsSchemaNameValid(name string) error {
	__antithesis_instrumentation__.Notify(267712)

	if strings.HasPrefix(name, catconstants.PgSchemaPrefix) {
		__antithesis_instrumentation__.Notify(267714)
		err := pgerror.Newf(pgcode.ReservedName, "unacceptable schema name %q", name)
		err = errors.WithDetail(err, `The prefix "pg_" is reserved for system schemas.`)
		return err
	} else {
		__antithesis_instrumentation__.Notify(267715)
	}
	__antithesis_instrumentation__.Notify(267713)
	return nil
}
