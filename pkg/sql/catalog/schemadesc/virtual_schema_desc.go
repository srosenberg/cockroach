package schemadesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
)

func GetVirtualSchemaByID(id descpb.ID) (catalog.SchemaDescriptor, bool) {
	__antithesis_instrumentation__.Notify(267770)
	sc, ok := virtualSchemasByID[id]
	return sc, ok
}

var virtualSchemasByID = func() map[descpb.ID]catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(267771)
	m := make(map[descpb.ID]catalog.SchemaDescriptor, len(catconstants.StaticSchemaIDMap))
	for id, name := range catconstants.StaticSchemaIDMap {
		__antithesis_instrumentation__.Notify(267773)
		id := descpb.ID(id)
		sc := virtual{
			synthetic: synthetic{virtualBase{}},
			id:        id,
			name:      name,
		}
		m[id] = sc
	}
	__antithesis_instrumentation__.Notify(267772)
	return m
}()

type virtual struct {
	synthetic
	id   descpb.ID
	name string
}

var _ catalog.SchemaDescriptor = virtual{}

func (p virtual) GetID() descpb.ID { __antithesis_instrumentation__.Notify(267774); return p.id }
func (p virtual) GetName() string  { __antithesis_instrumentation__.Notify(267775); return p.name }
func (p virtual) GetParentID() descpb.ID {
	__antithesis_instrumentation__.Notify(267776)
	return descpb.InvalidID
}
func (p virtual) GetPrivileges() *catpb.PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(267777)
	return catpb.NewVirtualSchemaPrivilegeDescriptor()
}

type virtualBase struct{}

var _ syntheticBase = virtualBase{}

func (v virtualBase) kindName() string {
	__antithesis_instrumentation__.Notify(267778)
	return "virtual"
}
func (v virtualBase) kind() catalog.ResolvedSchemaKind {
	__antithesis_instrumentation__.Notify(267779)
	return catalog.SchemaVirtual
}
