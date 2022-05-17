package schemadesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
)

type SchemaDescriptorBuilder interface {
	catalog.DescriptorBuilder
	BuildImmutableSchema() catalog.SchemaDescriptor
	BuildExistingMutableSchema() *Mutable
	BuildCreatedMutableSchema() *Mutable
}

type schemaDescriptorBuilder struct {
	original             *descpb.SchemaDescriptor
	maybeModified        *descpb.SchemaDescriptor
	isUncommittedVersion bool
	changes              catalog.PostDeserializationChanges
}

var _ SchemaDescriptorBuilder = &schemaDescriptorBuilder{}

func NewBuilder(desc *descpb.SchemaDescriptor) SchemaDescriptorBuilder {
	__antithesis_instrumentation__.Notify(267716)
	return newBuilder(desc, false,
		catalog.PostDeserializationChanges{})
}

func newBuilder(
	desc *descpb.SchemaDescriptor,
	isUncommittedVersion bool,
	changes catalog.PostDeserializationChanges,
) SchemaDescriptorBuilder {
	__antithesis_instrumentation__.Notify(267717)
	return &schemaDescriptorBuilder{
		original:             protoutil.Clone(desc).(*descpb.SchemaDescriptor),
		isUncommittedVersion: isUncommittedVersion,
		changes:              changes,
	}
}

func (sdb *schemaDescriptorBuilder) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(267718)
	return catalog.Schema
}

func (sdb *schemaDescriptorBuilder) RunPostDeserializationChanges() error {
	__antithesis_instrumentation__.Notify(267719)
	sdb.maybeModified = protoutil.Clone(sdb.original).(*descpb.SchemaDescriptor)
	privsChanged := catprivilege.MaybeFixPrivileges(
		&sdb.maybeModified.Privileges,
		sdb.maybeModified.GetParentID(),
		descpb.InvalidID,
		privilege.Schema,
		sdb.maybeModified.GetName(),
	)
	addedGrantOptions := catprivilege.MaybeUpdateGrantOptions(sdb.maybeModified.Privileges)
	if privsChanged || func() bool {
		__antithesis_instrumentation__.Notify(267721)
		return addedGrantOptions == true
	}() == true {
		__antithesis_instrumentation__.Notify(267722)
		sdb.changes.Add(catalog.UpgradedPrivileges)
	} else {
		__antithesis_instrumentation__.Notify(267723)
	}
	__antithesis_instrumentation__.Notify(267720)
	return nil
}

func (sdb *schemaDescriptorBuilder) RunRestoreChanges(
	_ func(id descpb.ID) catalog.Descriptor,
) error {
	__antithesis_instrumentation__.Notify(267724)
	return nil
}

func (sdb *schemaDescriptorBuilder) BuildImmutable() catalog.Descriptor {
	__antithesis_instrumentation__.Notify(267725)
	return sdb.BuildImmutableSchema()
}

func (sdb *schemaDescriptorBuilder) BuildImmutableSchema() catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(267726)
	desc := sdb.maybeModified
	if desc == nil {
		__antithesis_instrumentation__.Notify(267728)
		desc = sdb.original
	} else {
		__antithesis_instrumentation__.Notify(267729)
	}
	__antithesis_instrumentation__.Notify(267727)
	return &immutable{
		SchemaDescriptor:     *desc,
		changes:              sdb.changes,
		isUncommittedVersion: sdb.isUncommittedVersion,
	}
}

func (sdb *schemaDescriptorBuilder) BuildExistingMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(267730)
	return sdb.BuildExistingMutableSchema()
}

func (sdb *schemaDescriptorBuilder) BuildExistingMutableSchema() *Mutable {
	__antithesis_instrumentation__.Notify(267731)
	if sdb.maybeModified == nil {
		__antithesis_instrumentation__.Notify(267733)
		sdb.maybeModified = protoutil.Clone(sdb.original).(*descpb.SchemaDescriptor)
	} else {
		__antithesis_instrumentation__.Notify(267734)
	}
	__antithesis_instrumentation__.Notify(267732)
	return &Mutable{
		immutable: immutable{
			SchemaDescriptor:     *sdb.maybeModified,
			changes:              sdb.changes,
			isUncommittedVersion: sdb.isUncommittedVersion,
		},
		ClusterVersion: &immutable{SchemaDescriptor: *sdb.original},
	}
}

func (sdb *schemaDescriptorBuilder) BuildCreatedMutable() catalog.MutableDescriptor {
	__antithesis_instrumentation__.Notify(267735)
	return sdb.BuildCreatedMutableSchema()
}

func (sdb *schemaDescriptorBuilder) BuildCreatedMutableSchema() *Mutable {
	__antithesis_instrumentation__.Notify(267736)
	return &Mutable{
		immutable: immutable{
			SchemaDescriptor: *sdb.original,
			changes:          sdb.changes,
		},
	}
}
