package schemadesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catprivilege"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type synthetic struct {
	syntheticBase
}

type syntheticBase interface {
	kindName() string
	kind() catalog.ResolvedSchemaKind
}

func (p synthetic) GetParentSchemaID() descpb.ID {
	__antithesis_instrumentation__.Notify(267737)
	return descpb.InvalidID
}
func (p synthetic) IsUncommittedVersion() bool {
	__antithesis_instrumentation__.Notify(267738)
	return false
}
func (p synthetic) GetVersion() descpb.DescriptorVersion {
	__antithesis_instrumentation__.Notify(267739)
	return 1
}
func (p synthetic) GetModificationTime() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(267740)
	return hlc.Timestamp{}
}

func (p synthetic) GetDrainingNames() []descpb.NameInfo {
	__antithesis_instrumentation__.Notify(267741)
	return nil
}
func (p synthetic) GetPrivileges() *catpb.PrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(267742)
	log.Fatalf(context.TODO(), "cannot access privileges on a %s descriptor", p.kindName())
	return nil
}
func (p synthetic) DescriptorType() catalog.DescriptorType {
	__antithesis_instrumentation__.Notify(267743)
	return catalog.Schema
}
func (p synthetic) GetAuditMode() descpb.TableDescriptor_AuditMode {
	__antithesis_instrumentation__.Notify(267744)
	return descpb.TableDescriptor_DISABLED
}
func (p synthetic) Public() bool {
	__antithesis_instrumentation__.Notify(267745)
	return true
}
func (p synthetic) Adding() bool {
	__antithesis_instrumentation__.Notify(267746)
	return false
}
func (p synthetic) Dropped() bool {
	__antithesis_instrumentation__.Notify(267747)
	return false
}
func (p synthetic) Offline() bool {
	__antithesis_instrumentation__.Notify(267748)
	return false
}
func (p synthetic) GetOfflineReason() string {
	__antithesis_instrumentation__.Notify(267749)
	return ""
}
func (p synthetic) DescriptorProto() *descpb.Descriptor {
	__antithesis_instrumentation__.Notify(267750)
	log.Fatalf(context.TODO(),
		"%s schema cannot be encoded", p.kindName())
	return nil
}
func (p synthetic) ByteSize() int64 {
	__antithesis_instrumentation__.Notify(267751)
	return 0
}
func (p synthetic) NewBuilder() catalog.DescriptorBuilder {
	__antithesis_instrumentation__.Notify(267752)
	log.Fatalf(context.TODO(),
		"%s schema cannot create a builder", p.kindName())
	return nil
}
func (p synthetic) GetReferencedDescIDs() (catalog.DescriptorIDSet, error) {
	__antithesis_instrumentation__.Notify(267753)
	return catalog.DescriptorIDSet{}, nil
}
func (p synthetic) ValidateSelf(_ catalog.ValidationErrorAccumulator) {
	__antithesis_instrumentation__.Notify(267754)
}
func (p synthetic) ValidateCrossReferences(
	_ catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(267755)
}
func (p synthetic) ValidateTxnCommit(
	_ catalog.ValidationErrorAccumulator, _ catalog.ValidationDescGetter,
) {
	__antithesis_instrumentation__.Notify(267756)
}
func (p synthetic) SchemaKind() catalog.ResolvedSchemaKind {
	__antithesis_instrumentation__.Notify(267757)
	return p.kind()
}
func (p synthetic) SchemaDesc() *descpb.SchemaDescriptor {
	__antithesis_instrumentation__.Notify(267758)
	log.Fatalf(context.TODO(),
		"synthetic %s cannot be encoded", p.kindName())
	return nil
}
func (p synthetic) GetDeclarativeSchemaChangerState() *scpb.DescriptorState {
	__antithesis_instrumentation__.Notify(267759)
	return nil
}
func (p synthetic) GetPostDeserializationChanges() catalog.PostDeserializationChanges {
	__antithesis_instrumentation__.Notify(267760)
	return catalog.PostDeserializationChanges{}
}

func (p synthetic) HasConcurrentSchemaChanges() bool {
	__antithesis_instrumentation__.Notify(267761)
	return false
}

func (p synthetic) GetDefaultPrivilegeDescriptor() catalog.DefaultPrivilegeDescriptor {
	__antithesis_instrumentation__.Notify(267762)
	return catprivilege.MakeDefaultPrivileges(catprivilege.MakeDefaultPrivilegeDescriptor(catpb.DefaultPrivilegeDescriptor_SCHEMA))
}
