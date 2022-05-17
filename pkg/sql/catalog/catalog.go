package catalog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type MutableDescriptor interface {
	Descriptor

	MaybeIncrementVersion()

	SetDrainingNames([]descpb.NameInfo)

	AddDrainingName(descpb.NameInfo)

	OriginalName() string
	OriginalID() descpb.ID
	OriginalVersion() descpb.DescriptorVersion

	ImmutableCopy() Descriptor

	IsNew() bool

	SetPublic()

	SetDropped()

	SetOffline(reason string)

	SetDeclarativeSchemaChangerState(*scpb.DescriptorState)
}

type VirtualSchemas interface {
	GetVirtualSchema(schemaName string) (VirtualSchema, bool)
	GetVirtualSchemaByID(id descpb.ID) (VirtualSchema, bool)
	GetVirtualObjectByID(id descpb.ID) (VirtualObject, bool)
}

type VirtualSchema interface {
	Desc() SchemaDescriptor
	NumTables() int
	VisitTables(func(object VirtualObject))
	GetObjectByName(name string, flags tree.ObjectLookupFlags) (VirtualObject, error)
}

type VirtualObject interface {
	Desc() Descriptor
}

type ResolvedObjectPrefix struct {
	ExplicitDatabase, ExplicitSchema bool

	Database DatabaseDescriptor

	Schema SchemaDescriptor
}

func (p ResolvedObjectPrefix) NamePrefix() tree.ObjectNamePrefix {
	__antithesis_instrumentation__.Notify(247299)
	var n tree.ObjectNamePrefix
	n.ExplicitCatalog = p.ExplicitDatabase
	n.ExplicitSchema = p.ExplicitSchema
	if p.Database != nil {
		__antithesis_instrumentation__.Notify(247302)
		n.CatalogName = tree.Name(p.Database.GetName())
	} else {
		__antithesis_instrumentation__.Notify(247303)
	}
	__antithesis_instrumentation__.Notify(247300)
	if p.Schema != nil {
		__antithesis_instrumentation__.Notify(247304)
		n.SchemaName = tree.Name(p.Schema.GetName())
	} else {
		__antithesis_instrumentation__.Notify(247305)
	}
	__antithesis_instrumentation__.Notify(247301)
	return n
}

const NumSystemColumns = 2

const SmallestSystemColumnColumnID = math.MaxUint32 - NumSystemColumns + 1
