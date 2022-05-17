package schemadesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
)

func NewTemporarySchema(name string, id descpb.ID, parentDB descpb.ID) catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(267763)
	return &temporary{
		synthetic: synthetic{temporaryBase{}},
		id:        id,
		name:      name,
		parentID:  parentDB,
	}
}

type temporary struct {
	synthetic
	id       descpb.ID
	name     string
	parentID descpb.ID
}

var _ catalog.SchemaDescriptor = temporary{}

func (p temporary) GetID() descpb.ID { __antithesis_instrumentation__.Notify(267764); return p.id }
func (p temporary) GetName() string  { __antithesis_instrumentation__.Notify(267765); return p.name }
func (p temporary) GetParentID() descpb.ID {
	__antithesis_instrumentation__.Notify(267766)
	return p.parentID
}
func (p temporary) GetPrivileges() *catpb.PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(267767)
	return catpb.NewTemporarySchemaPrivilegeDescriptor()
}

type temporaryBase struct{}

func (t temporaryBase) kindName() string {
	__antithesis_instrumentation__.Notify(267768)
	return "temporary"
}
func (t temporaryBase) kind() catalog.ResolvedSchemaKind {
	__antithesis_instrumentation__.Notify(267769)
	return catalog.SchemaTemporary
}

var _ syntheticBase = temporaryBase{}
