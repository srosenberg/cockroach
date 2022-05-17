package schemadesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func GetPublicSchema() catalog.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(267609)
	return publicDesc
}

type public struct {
	synthetic
}

var _ catalog.SchemaDescriptor = public{}

func (p public) GetParentID() descpb.ID {
	__antithesis_instrumentation__.Notify(267610)
	return descpb.InvalidID
}
func (p public) GetID() descpb.ID {
	__antithesis_instrumentation__.Notify(267611)
	return keys.PublicSchemaID
}
func (p public) GetName() string {
	__antithesis_instrumentation__.Notify(267612)
	return tree.PublicSchema
}
func (p public) GetPrivileges() *catpb.PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(267613)
	return catpb.NewPublicSchemaPrivilegeDescriptor()
}

type publicBase struct{}

func (p publicBase) kindName() string { __antithesis_instrumentation__.Notify(267614); return "public" }
func (p publicBase) kind() catalog.ResolvedSchemaKind {
	__antithesis_instrumentation__.Notify(267615)
	return catalog.SchemaPublic
}

var publicDesc catalog.SchemaDescriptor = public{
	synthetic{publicBase{}},
}
