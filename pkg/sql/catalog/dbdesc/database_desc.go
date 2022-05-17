// Package dbdesc contains the concrete implementations of
// catalog.DatabaseDescriptor.
package dbdesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/multiregion"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var _ catalog.DatabaseDescriptor = (*immutable)(nil)
var _ catalog.DatabaseDescriptor = (*Mutable)(nil)
var _ catalog.MutableDescriptor = (*Mutable)(nil)

type immutable struct {
	descpb.DatabaseDescriptor

	isUncommittedVersion bool

	changes catalog.PostDeserializationChanges
}

type Mutable struct {
	immutable
	ClusterVersion *immutable
}

func (desc *immutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(251231)
	return formatSafeMessage("dbdesc.immutable", desc)
}

func (desc *Mutable) SafeMessage() string {
	__antithesis_instrumentation__.Notify(251232)
	return formatSafeMessage("dbdesc.Mutable", desc)
}

func formatSafeMessage(typeName string, desc catalog.DatabaseDescriptor) string {
	__antithesis_instrumentation__.Notify(251233)
	var buf redact.StringBuilder
	buf.Print(typeName + ": {")
	catalog.FormatSafeDescriptorProperties(&buf, desc)
	buf.Print("}")
	return buf.String()
}

func (desc *immutable) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(251234)
	return catalog.Database
}

func (desc *immutable) DatabaseDesc() *descpb.DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(251235)
	return &desc.DatabaseDescriptor
}

func (desc *Mutable) SetDrainingNames(names []descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(251236)
	desc.DrainingNames = names
}

func (desc *immutable) GetParentID() descpb.ID {
	__antithesis_instrumentation__.Notify(251237)
	return keys.RootNamespaceID
}

func (desc *immutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(251238)
	return desc.isUncommittedVersion
}

func (desc *immutable) GetParentSchemaID() descpb.ID {
	__antithesis_instrumentation__.Notify(251239)
	return keys.RootNamespaceID
}

func (desc *immutable) GetAuditMode() descpb.TableDescriptor_AuditMode {
	__antithesis_instrumentation__.Notify(251240)
	return descpb.TableDescriptor_DISABLED
}

func (desc *immutable) Public() bool {
	__antithesis_instrumentation__.Notify(251241)
	return desc.State == descpb.DescriptorState_PUBLIC
}

func (desc *immutable) Adding() bool {
	__antithesis_instrumentation__.Notify(251242)
	return false
}

func (desc *immutable) Offline() bool {
	__antithesis_instrumentation__.Notify(251243)
	return desc.State == descpb.DescriptorState_OFFLINE
}

func (desc *immutable) Dropped() bool {
	__antithesis_instrumentation__.Notify(251244)
	return desc.State == descpb.DescriptorState_DROP
}

func (desc *immutable) DescriptorProto() *descpb.Descriptor {
	__antithesis_instrumentation__.Notify(251245)
	return &descpb.Descriptor{
		Union: &descpb.Descriptor_Database{
			Database: &desc.DatabaseDescriptor,
		},
	}
}

func (desc *immutable) ByteSize() int64 {
	__antithesis_instrumentation__.Notify(251246)
	return int64(desc.Size())
}

func (desc *immutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(251247)
	return newBuilder(desc.DatabaseDesc(), desc.isUncommittedVersion, desc.changes)
}

func (desc *Mutable) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(251248)
	return newBuilder(desc.DatabaseDesc(), desc.IsUncommittedVersion(), desc.changes)
}

func (desc *immutable) IsMultiRegion() bool {
	__antithesis_instrumentation__.Notify(251249)
	return desc.RegionConfig != nil
}

func (desc *immutable) PrimaryRegionName() (catpb.RegionName, error) {
	__antithesis_instrumentation__.Notify(251250)
	if !desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(251252)
		return "", errors.AssertionFailedf(
			"can not get the primary region of a non multi-region database")
	} else {
		__antithesis_instrumentation__.Notify(251253)
	}
	__antithesis_instrumentation__.Notify(251251)
	return desc.RegionConfig.PrimaryRegion, nil
}

func (desc *immutable) MultiRegionEnumID() (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(251254)
	if !desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(251256)
		return descpb.InvalidID, errors.AssertionFailedf(
			"can not get multi-region enum ID of a non multi-region database")
	} else {
		__antithesis_instrumentation__.Notify(251257)
	}
	__antithesis_instrumentation__.Notify(251255)
	return desc.RegionConfig.RegionEnumID, nil
}

func (desc *Mutable) SetName(name string) {
	__antithesis_instrumentation__.Notify(251258)
	desc.Name = name
}

func (desc *immutable) ForEachSchemaInfo(
	f func(id descpb.ID, name string, isDropped bool) error,
) error {
	__antithesis_instrumentation__.Notify(251259)
	for name, info := range desc.Schemas {
		__antithesis_instrumentation__.Notify(251261)
		if err := f(info.ID, name, info.Dropped); err != nil {
			__antithesis_instrumentation__.Notify(251262)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(251264)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(251265)
			}
			__antithesis_instrumentation__.Notify(251263)
			return err
		} else {
			__antithesis_instrumentation__.Notify(251266)
		}
	}
	__antithesis_instrumentation__.Notify(251260)
	return nil
}

func (desc *immutable) ForEachNonDroppedSchema(f func(id descpb.ID, name string) error) error {
	__antithesis_instrumentation__.Notify(251267)
	for name, info := range desc.Schemas {
		__antithesis_instrumentation__.Notify(251269)
		if info.Dropped {
			__antithesis_instrumentation__.Notify(251271)
			continue
		} else {
			__antithesis_instrumentation__.Notify(251272)
		}
		__antithesis_instrumentation__.Notify(251270)
		if err := f(info.ID, name); err != nil {
			__antithesis_instrumentation__.Notify(251273)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(251275)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(251276)
			}
			__antithesis_instrumentation__.Notify(251274)
			return err
		} else {
			__antithesis_instrumentation__.Notify(251277)
		}
	}
	__antithesis_instrumentation__.Notify(251268)
	return nil
}

func (desc *immutable) GetSchemaID(name string) descpb.ID {
	__antithesis_instrumentation__.Notify(251278)
	info := desc.Schemas[name]
	if info.Dropped {
		__antithesis_instrumentation__.Notify(251280)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(251281)
	}
	__antithesis_instrumentation__.Notify(251279)
	return info.ID
}

func (desc *immutable) HasPublicSchemaWithDescriptor() bool {
	__antithesis_instrumentation__.Notify(251282)

	if desc.ID == keys.SystemDatabaseID {
		__antithesis_instrumentation__.Notify(251284)
		return false
	} else {
		__antithesis_instrumentation__.Notify(251285)
	}
	__antithesis_instrumentation__.Notify(251283)
	_, found := desc.Schemas[tree.PublicSchema]
	return found
}

func (desc *immutable) GetNonDroppedSchemaName(schemaID descpb.ID) string {
	__antithesis_instrumentation__.Notify(251286)
	for name, info := range desc.Schemas {
		__antithesis_instrumentation__.Notify(251288)
		if !info.Dropped && func() bool {
			__antithesis_instrumentation__.Notify(251289)
			return info.ID == schemaID == true
		}() == true {
			__antithesis_instrumentation__.Notify(251290)
			return name
		} else {
			__antithesis_instrumentation__.Notify(251291)
		}
	}
	__antithesis_instrumentation__.Notify(251287)
	return ""
}

func (desc *immutable) ValidateSelf(vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(251292)

	vea.Report(catalog.ValidateName(desc.GetName(), "descriptor"))
	if desc.GetID() == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(251296)
		vea.Report(fmt.Errorf("invalid database ID %d", desc.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(251297)
	}
	__antithesis_instrumentation__.Notify(251293)

	if desc.Privileges == nil {
		__antithesis_instrumentation__.Notify(251298)
		vea.Report(errors.AssertionFailedf("privileges not set"))
	} else {
		__antithesis_instrumentation__.Notify(251299)
		vea.Report(catprivilege.Validate(*desc.Privileges, desc, privilege.Database))
	}
	__antithesis_instrumentation__.Notify(251294)

	if desc.GetDefaultPrivileges() != nil {
		__antithesis_instrumentation__.Notify(251300)

		vea.Report(catprivilege.ValidateDefaultPrivileges(*desc.GetDefaultPrivileges()))
	} else {
		__antithesis_instrumentation__.Notify(251301)
	}
	__antithesis_instrumentation__.Notify(251295)

	if desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(251302)
		desc.validateMultiRegion(vea)
	} else {
		__antithesis_instrumentation__.Notify(251303)
	}
}

func (desc *immutable) validateMultiRegion(vea catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(251304)
	if desc.RegionConfig.PrimaryRegion == "" {
		__antithesis_instrumentation__.Notify(251305)
		vea.Report(errors.AssertionFailedf(
			"primary region unset on a multi-region db %d", desc.GetID()))
	} else {
		__antithesis_instrumentation__.Notify(251306)
	}
}

func (desc *immutable) GetReferencedDescIDs() (catalog.DescriptorIDSet, error) {
	__antithesis_instrumentation__.Notify(251307)
	ids := catalog.MakeDescriptorIDSet(desc.GetID())
	if desc.IsMultiRegion() {
		__antithesis_instrumentation__.Notify(251310)
		id, err := desc.MultiRegionEnumID()
		if err != nil {
			__antithesis_instrumentation__.Notify(251312)
			return catalog.DescriptorIDSet{}, err
		} else {
			__antithesis_instrumentation__.Notify(251313)
		}
		__antithesis_instrumentation__.Notify(251311)
		ids.Add(id)
	} else {
		__antithesis_instrumentation__.Notify(251314)
	}
	__antithesis_instrumentation__.Notify(251308)
	for _, schema := range desc.Schemas {
		__antithesis_instrumentation__.Notify(251315)
		ids.Add(schema.ID)
	}
	__antithesis_instrumentation__.Notify(251309)
	return ids, nil
}

func (desc *immutable) ValidateCrossReferences(
	vea catalog.ValidationErrorAccumulator, vdg catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(251316)

	if enumID, err := desc.MultiRegionEnumID(); err == nil {
		__antithesis_instrumentation__.Notify(251317)
		report := func(err error) {
			__antithesis_instrumentation__.Notify(251321)
			vea.Report(errors.Wrap(err, "multi-region enum"))
		}
		__antithesis_instrumentation__.Notify(251318)
		typ, err := vdg.GetTypeDescriptor(enumID)
		if err != nil {
			__antithesis_instrumentation__.Notify(251322)
			report(err)
			return
		} else {
			__antithesis_instrumentation__.Notify(251323)
		}
		__antithesis_instrumentation__.Notify(251319)
		if typ.Dropped() {
			__antithesis_instrumentation__.Notify(251324)
			report(errors.Errorf("type descriptor is dropped"))
		} else {
			__antithesis_instrumentation__.Notify(251325)
		}
		__antithesis_instrumentation__.Notify(251320)
		if typ.GetParentID() != desc.GetID() {
			__antithesis_instrumentation__.Notify(251326)
			report(errors.Errorf("parentID is actually %d", typ.GetParentID()))
		} else {
			__antithesis_instrumentation__.Notify(251327)
		}

	} else {
		__antithesis_instrumentation__.Notify(251328)
	}
}

func (desc *immutable) ValidateTxnCommit(
	vea catalog.ValidationErrorAccumulator, vdg catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(251329)

	for schemaName, schemaInfo := range desc.Schemas {
		__antithesis_instrumentation__.Notify(251330)
		if schemaInfo.Dropped {
			__antithesis_instrumentation__.Notify(251336)
			continue
		} else {
			__antithesis_instrumentation__.Notify(251337)
		}
		__antithesis_instrumentation__.Notify(251331)
		report := func(err error) {
			__antithesis_instrumentation__.Notify(251338)
			vea.Report(errors.Wrapf(err, "schema mapping entry %q (%d)",
				errors.Safe(schemaName), schemaInfo.ID))
		}
		__antithesis_instrumentation__.Notify(251332)
		schemaDesc, err := vdg.GetSchemaDescriptor(schemaInfo.ID)
		if err != nil {
			__antithesis_instrumentation__.Notify(251339)
			report(err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(251340)
		}
		__antithesis_instrumentation__.Notify(251333)
		if schemaDesc.GetName() != schemaName {
			__antithesis_instrumentation__.Notify(251341)
			report(errors.Errorf("schema name is actually %q", errors.Safe(schemaDesc.GetName())))
		} else {
			__antithesis_instrumentation__.Notify(251342)
		}
		__antithesis_instrumentation__.Notify(251334)
		if schemaDesc.GetParentID() != desc.GetID() {
			__antithesis_instrumentation__.Notify(251343)
			report(errors.Errorf("schema parentID is actually %d", schemaDesc.GetParentID()))
		} else {
			__antithesis_instrumentation__.Notify(251344)
		}
		__antithesis_instrumentation__.Notify(251335)
		if schemaDesc.Dropped() {
			__antithesis_instrumentation__.Notify(251345)
			report(errors.Errorf("back-referenced schema %q (%d) is dropped",
				schemaDesc.GetName(), schemaDesc.GetID()))
		} else {
			__antithesis_instrumentation__.Notify(251346)
		}
	}
}

func (desc *Mutable) MaybeIncrementVersion() {
	__antithesis_instrumentation__.Notify(251347)

	if desc.ClusterVersion == nil || func() bool {
		__antithesis_instrumentation__.Notify(251349)
		return desc.Version == desc.ClusterVersion.Version+1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(251350)
		return
	} else {
		__antithesis_instrumentation__.Notify(251351)
	}
	__antithesis_instrumentation__.Notify(251348)
	desc.Version++
	desc.ModificationTime = hlc.Timestamp{}
}

func (desc *Mutable) OriginalName() string {
	__antithesis_instrumentation__.Notify(251352)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(251354)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(251355)
	}
	__antithesis_instrumentation__.Notify(251353)
	return desc.ClusterVersion.Name
}

func (desc *Mutable) OriginalID() descpb.ID {
	__antithesis_instrumentation__.Notify(251356)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(251358)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(251359)
	}
	__antithesis_instrumentation__.Notify(251357)
	return desc.ClusterVersion.ID
}

func (desc *Mutable) OriginalVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(251360)
	if desc.ClusterVersion == nil {
		__antithesis_instrumentation__.Notify(251362)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(251363)
	}
	__antithesis_instrumentation__.Notify(251361)
	return desc.ClusterVersion.Version
}

func (desc *Mutable) ImmutableCopy() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(251364)
	return desc.NewBuilder().BuildImmutable()
}

func (desc *Mutable) IsNew() bool {
	__antithesis_instrumentation__.Notify(251365)
	return desc.ClusterVersion == nil
}

func (desc *Mutable) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(251366)
	return desc.IsNew() || func() bool {
		__antithesis_instrumentation__.Notify(251367)
		return desc.GetVersion() != desc.ClusterVersion.GetVersion() == true
	}() == true
}

func (desc *Mutable) SetPublic() {
	__antithesis_instrumentation__.Notify(251368)
	desc.State = descpb.DescriptorState_PUBLIC
	desc.OfflineReason = ""
}

func (desc *Mutable) SetDropped() {
	__antithesis_instrumentation__.Notify(251369)
	desc.State = descpb.DescriptorState_DROP
	desc.OfflineReason = ""
}

func (desc *Mutable) SetOffline(reason string) {
	__antithesis_instrumentation__.Notify(251370)
	desc.State = descpb.DescriptorState_OFFLINE
	desc.OfflineReason = reason
}

func (desc *Mutable) AddDrainingName(name descpb.NameInfo) {
	__antithesis_instrumentation__.Notify(251371)
	desc.DrainingNames = append(desc.DrainingNames, name)
}

func (desc *Mutable) UnsetMultiRegionConfig() {
	__antithesis_instrumentation__.Notify(251372)
	desc.RegionConfig = nil
}

func (desc *Mutable) SetInitialMultiRegionConfig(config *multiregion.RegionConfig) error {
	__antithesis_instrumentation__.Notify(251373)

	if desc.RegionConfig != nil {
		__antithesis_instrumentation__.Notify(251375)
		return errors.AssertionFailedf(
			"expected no region config on database %q with ID %d",
			desc.GetName(),
			desc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(251376)
	}
	__antithesis_instrumentation__.Notify(251374)
	desc.RegionConfig = &descpb.DatabaseDescriptor_RegionConfig{
		SurvivalGoal:  config.SurvivalGoal(),
		PrimaryRegion: config.PrimaryRegion(),
		RegionEnumID:  config.RegionEnumID(),
	}
	return nil
}

func (desc *Mutable) SetRegionConfig(cfg *descpb.DatabaseDescriptor_RegionConfig) {
	__antithesis_instrumentation__.Notify(251377)
	desc.RegionConfig = cfg
}

func (desc *Mutable) SetPlacement(placement descpb.DataPlacement) {
	__antithesis_instrumentation__.Notify(251378)
	desc.RegionConfig.Placement = placement
}

func (desc *immutable) GetPostDeserializationChanges() catalog.PostDeserializationChanges {
	__antithesis_instrumentation__.Notify(251379)
	return desc.changes
}

func (desc *immutable) HasConcurrentSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(251380)
	return desc.DeclarativeSchemaChangerState != nil && func() bool {
		__antithesis_instrumentation__.Notify(251381)
		return desc.DeclarativeSchemaChangerState.JobID != catpb.InvalidJobID == true
	}() == true
}

func (desc *immutable) GetDefaultPrivilegeDescriptor() catalog.DefaultPrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(251382)
	defaultPrivilegeDescriptor := desc.GetDefaultPrivileges()
	if defaultPrivilegeDescriptor == nil {
		__antithesis_instrumentation__.Notify(251384)
		defaultPrivilegeDescriptor = catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_DATABASE)
	} else {
		__antithesis_instrumentation__.Notify(251385)
	}
	__antithesis_instrumentation__.Notify(251383)
	return catprivilege.MakeDefaultPrivileges(defaultPrivilegeDescriptor)
}

func (desc *Mutable) GetMutableDefaultPrivilegeDescriptor() *catprivilege.Mutable {
	__antithesis_instrumentation__.Notify(251386)
	defaultPrivilegeDescriptor := desc.GetDefaultPrivileges()
	if defaultPrivilegeDescriptor == nil {
		__antithesis_instrumentation__.Notify(251388)
		defaultPrivilegeDescriptor = catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_DATABASE)
	} else {
		__antithesis_instrumentation__.Notify(251389)
	}
	__antithesis_instrumentation__.Notify(251387)
	return catprivilege.NewMutableDefaultPrivileges(defaultPrivilegeDescriptor)
}

func (desc *Mutable) SetDefaultPrivilegeDescriptor(
	defaultPrivilegeDescriptor *catpb.DefaultPrivilegeDescriptor,
) {
	__antithesis_instrumentation__.Notify(251390)
	desc.DefaultPrivileges = defaultPrivilegeDescriptor
}

func (desc *Mutable) AddSchemaToDatabase(
	schemaName string, schemaInfo descpb.DatabaseDescriptor_SchemaInfo,
) {
	__antithesis_instrumentation__.Notify(251391)
	if desc.Schemas == nil {
		__antithesis_instrumentation__.Notify(251393)
		desc.Schemas = make(map[string]descpb.DatabaseDescriptor_SchemaInfo)
	} else {
		__antithesis_instrumentation__.Notify(251394)
	}
	__antithesis_instrumentation__.Notify(251392)
	desc.Schemas[schemaName] = schemaInfo
}

func (desc *immutable) GetDeclarativeSchemaChangerState() *scpb.DescriptorState {
	__antithesis_instrumentation__.Notify(251395)
	return desc.DeclarativeSchemaChangerState.Clone()
}

func (desc *Mutable) SetDeclarativeSchemaChangerState(state *scpb.DescriptorState) {
	__antithesis_instrumentation__.Notify(251396)
	desc.DeclarativeSchemaChangerState = state
}

func maybeRemoveDroppedSelfEntryFromSchemas(dbDesc *descpb.DatabaseDescriptor) bool {
	__antithesis_instrumentation__.Notify(251397)
	if dbDesc == nil {
		__antithesis_instrumentation__.Notify(251400)
		return false
	} else {
		__antithesis_instrumentation__.Notify(251401)
	}
	__antithesis_instrumentation__.Notify(251398)
	if sc, ok := dbDesc.Schemas[dbDesc.Name]; ok && func() bool {
		__antithesis_instrumentation__.Notify(251402)
		return sc.ID == dbDesc.ID == true
	}() == true {
		__antithesis_instrumentation__.Notify(251403)
		delete(dbDesc.Schemas, dbDesc.Name)
		return true
	} else {
		__antithesis_instrumentation__.Notify(251404)
	}
	__antithesis_instrumentation__.Notify(251399)
	return false
}
