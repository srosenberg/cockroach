package typedesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

type TypeDescriptorBuilder interface {
	catalog.DescriptorBuilder
	BuildImmutableType() catalog.TypeDescriptor
	BuildExistingMutableType() *Mutable
	BuildCreatedMutableType() *Mutable
}

type typeDescriptorBuilder struct {
	original      *descpb.TypeDescriptor
	maybeModified *descpb.TypeDescriptor

	isUncommittedVersion bool
	changes              catalog.PostDeserializationChanges
}

var _ TypeDescriptorBuilder = &typeDescriptorBuilder{}

func NewBuilder(desc *descpb.TypeDescriptor) TypeDescriptorBuilder {
	__antithesis_instrumentation__.Notify(272340)
	return newBuilder(desc, false,
		catalog.PostDeserializationChanges{})
}

func newBuilder(
	desc *descpb.TypeDescriptor,
	isUncommittedVersion bool,
	changes catalog.PostDeserializationChanges,
) TypeDescriptorBuilder {
	__antithesis_instrumentation__.Notify(272341)
	b := &typeDescriptorBuilder{
		original:             protoutil.Clone(desc).(*descpb.TypeDescriptor),
		isUncommittedVersion: isUncommittedVersion,
		changes:              changes,
	}
	return b
}

func (tdb *typeDescriptorBuilder) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(272342)
	return catalog.Type
}

func (tdb *typeDescriptorBuilder) RunPostDeserializationChanges() error {
	__antithesis_instrumentation__.Notify(272343)
	tdb.maybeModified = protoutil.Clone(tdb.original).(*descpb.TypeDescriptor)
	fixedPrivileges := catprivilege.MaybeFixPrivileges(
		&tdb.maybeModified.Privileges,
		tdb.maybeModified.GetParentID(),
		tdb.maybeModified.GetParentSchemaID(),
		privilege.Type,
		tdb.maybeModified.GetName(),
	)
	addedGrantOptions := catprivilege.MaybeUpdateGrantOptions(tdb.maybeModified.Privileges)
	if fixedPrivileges || func() bool {
		__antithesis_instrumentation__.Notify(272345)
		return addedGrantOptions == true
	}() == true {
		__antithesis_instrumentation__.Notify(272346)
		tdb.changes.Add(catalog.UpgradedPrivileges)
	} else {
		__antithesis_instrumentation__.Notify(272347)
	}
	__antithesis_instrumentation__.Notify(272344)
	return nil
}

func (tdb *typeDescriptorBuilder) RunRestoreChanges(_ func(id descpb.ID) catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(272348)
	return nil
}

func (tdb *typeDescriptorBuilder) BuildImmutable() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(272349)
	return tdb.BuildImmutableType()
}

func (tdb *typeDescriptorBuilder) BuildImmutableType() catalog.TypeDescriptor {
	__antithesis_instrumentation__.Notify(272350)
	desc := tdb.maybeModified
	if desc == nil {
		__antithesis_instrumentation__.Notify(272352)
		desc = tdb.original
	} else {
		__antithesis_instrumentation__.Notify(272353)
	}
	__antithesis_instrumentation__.Notify(272351)
	imm := makeImmutable(desc, tdb.isUncommittedVersion, tdb.changes)
	return &imm
}

func (tdb *typeDescriptorBuilder) BuildExistingMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(272354)
	return tdb.BuildExistingMutableType()
}

func (tdb *typeDescriptorBuilder) BuildExistingMutableType() *Mutable {
	__antithesis_instrumentation__.Notify(272355)
	if tdb.maybeModified == nil {
		__antithesis_instrumentation__.Notify(272357)
		tdb.maybeModified = protoutil.Clone(tdb.original).(*descpb.TypeDescriptor)
	} else {
		__antithesis_instrumentation__.Notify(272358)
	}
	__antithesis_instrumentation__.Notify(272356)
	clusterVersion := makeImmutable(tdb.original, false,
		catalog.PostDeserializationChanges{})
	return &Mutable{
		immutable:      makeImmutable(tdb.maybeModified, false, tdb.changes),
		ClusterVersion: &clusterVersion,
	}
}

func (tdb *typeDescriptorBuilder) BuildCreatedMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(272359)
	return tdb.BuildCreatedMutableType()
}

func (tdb *typeDescriptorBuilder) BuildCreatedMutableType() *Mutable {
	__antithesis_instrumentation__.Notify(272360)
	return &Mutable{
		immutable: makeImmutable(tdb.original, tdb.isUncommittedVersion, tdb.changes),
	}
}

func makeImmutable(
	desc *descpb.TypeDescriptor,
	isUncommittedVersion bool,
	changes catalog.PostDeserializationChanges,
) immutable {
	__antithesis_instrumentation__.Notify(272361)
	immutDesc := immutable{
		TypeDescriptor:       *desc,
		isUncommittedVersion: isUncommittedVersion,
		changes:              changes,
	}

	switch immutDesc.Kind {
	case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
		__antithesis_instrumentation__.Notify(272363)
		immutDesc.logicalReps = make([]string, len(desc.EnumMembers))
		immutDesc.physicalReps = make([][]byte, len(desc.EnumMembers))
		immutDesc.readOnlyMembers = make([]bool, len(desc.EnumMembers))
		for i := range desc.EnumMembers {
			__antithesis_instrumentation__.Notify(272365)
			member := &desc.EnumMembers[i]
			immutDesc.logicalReps[i] = member.LogicalRepresentation
			immutDesc.physicalReps[i] = member.PhysicalRepresentation
			immutDesc.readOnlyMembers[i] =
				member.Capability == descpb.TypeDescriptor_EnumMember_READ_ONLY
		}
	default:
		__antithesis_instrumentation__.Notify(272364)
	}
	__antithesis_instrumentation__.Notify(272362)

	return immutDesc
}
