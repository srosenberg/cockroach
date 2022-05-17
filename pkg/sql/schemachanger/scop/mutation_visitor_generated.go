package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

type MutationOp interface {
	Op
	Visit(context.Context, MutationVisitor) error
}

type MutationVisitor interface {
	NotImplemented(context.Context, NotImplemented) error
	MakeAddedIndexDeleteOnly(context.Context, MakeAddedIndexDeleteOnly) error
	SetAddedIndexPartialPredicate(context.Context, SetAddedIndexPartialPredicate) error
	MakeAddedIndexDeleteAndWriteOnly(context.Context, MakeAddedIndexDeleteAndWriteOnly) error
	MakeAddedSecondaryIndexPublic(context.Context, MakeAddedSecondaryIndexPublic) error
	MakeAddedPrimaryIndexPublic(context.Context, MakeAddedPrimaryIndexPublic) error
	MakeDroppedPrimaryIndexDeleteAndWriteOnly(context.Context, MakeDroppedPrimaryIndexDeleteAndWriteOnly) error
	CreateGcJobForTable(context.Context, CreateGcJobForTable) error
	CreateGcJobForDatabase(context.Context, CreateGcJobForDatabase) error
	CreateGcJobForIndex(context.Context, CreateGcJobForIndex) error
	MarkDescriptorAsDroppedSynthetically(context.Context, MarkDescriptorAsDroppedSynthetically) error
	MarkDescriptorAsDropped(context.Context, MarkDescriptorAsDropped) error
	DrainDescriptorName(context.Context, DrainDescriptorName) error
	MakeAddedColumnDeleteAndWriteOnly(context.Context, MakeAddedColumnDeleteAndWriteOnly) error
	MakeDroppedNonPrimaryIndexDeleteAndWriteOnly(context.Context, MakeDroppedNonPrimaryIndexDeleteAndWriteOnly) error
	MakeDroppedIndexDeleteOnly(context.Context, MakeDroppedIndexDeleteOnly) error
	RemoveDroppedIndexPartialPredicate(context.Context, RemoveDroppedIndexPartialPredicate) error
	MakeIndexAbsent(context.Context, MakeIndexAbsent) error
	MakeAddedColumnDeleteOnly(context.Context, MakeAddedColumnDeleteOnly) error
	SetAddedColumnType(context.Context, SetAddedColumnType) error
	MakeColumnPublic(context.Context, MakeColumnPublic) error
	MakeDroppedColumnDeleteAndWriteOnly(context.Context, MakeDroppedColumnDeleteAndWriteOnly) error
	MakeDroppedColumnDeleteOnly(context.Context, MakeDroppedColumnDeleteOnly) error
	RemoveDroppedColumnType(context.Context, RemoveDroppedColumnType) error
	MakeColumnAbsent(context.Context, MakeColumnAbsent) error
	RemoveOwnerBackReferenceInSequence(context.Context, RemoveOwnerBackReferenceInSequence) error
	RemoveSequenceOwner(context.Context, RemoveSequenceOwner) error
	RemoveCheckConstraint(context.Context, RemoveCheckConstraint) error
	RemoveForeignKeyConstraint(context.Context, RemoveForeignKeyConstraint) error
	RemoveForeignKeyBackReference(context.Context, RemoveForeignKeyBackReference) error
	RemoveSchemaParent(context.Context, RemoveSchemaParent) error
	AddIndexPartitionInfo(context.Context, AddIndexPartitionInfo) error
	LogEvent(context.Context, LogEvent) error
	AddColumnFamily(context.Context, AddColumnFamily) error
	AddColumnDefaultExpression(context.Context, AddColumnDefaultExpression) error
	RemoveColumnDefaultExpression(context.Context, RemoveColumnDefaultExpression) error
	AddColumnOnUpdateExpression(context.Context, AddColumnOnUpdateExpression) error
	RemoveColumnOnUpdateExpression(context.Context, RemoveColumnOnUpdateExpression) error
	UpdateTableBackReferencesInTypes(context.Context, UpdateTableBackReferencesInTypes) error
	RemoveBackReferenceInTypes(context.Context, RemoveBackReferenceInTypes) error
	UpdateBackReferencesInSequences(context.Context, UpdateBackReferencesInSequences) error
	RemoveViewBackReferencesInRelations(context.Context, RemoveViewBackReferencesInRelations) error
	SetColumnName(context.Context, SetColumnName) error
	SetIndexName(context.Context, SetIndexName) error
	DeleteDescriptor(context.Context, DeleteDescriptor) error
	RemoveJobStateFromDescriptor(context.Context, RemoveJobStateFromDescriptor) error
	SetJobStateOnDescriptor(context.Context, SetJobStateOnDescriptor) error
	UpdateSchemaChangerJob(context.Context, UpdateSchemaChangerJob) error
	CreateSchemaChangerJob(context.Context, CreateSchemaChangerJob) error
	RemoveAllTableComments(context.Context, RemoveAllTableComments) error
	RemoveTableComment(context.Context, RemoveTableComment) error
	RemoveDatabaseComment(context.Context, RemoveDatabaseComment) error
	RemoveSchemaComment(context.Context, RemoveSchemaComment) error
	RemoveIndexComment(context.Context, RemoveIndexComment) error
	RemoveColumnComment(context.Context, RemoveColumnComment) error
	RemoveConstraintComment(context.Context, RemoveConstraintComment) error
	RemoveDatabaseRoleSettings(context.Context, RemoveDatabaseRoleSettings) error
	DeleteSchedule(context.Context, DeleteSchedule) error
}

func (op NotImplemented) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582415)
	return v.NotImplemented(ctx, op)
}

func (op MakeAddedIndexDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582416)
	return v.MakeAddedIndexDeleteOnly(ctx, op)
}

func (op SetAddedIndexPartialPredicate) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582417)
	return v.SetAddedIndexPartialPredicate(ctx, op)
}

func (op MakeAddedIndexDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582418)
	return v.MakeAddedIndexDeleteAndWriteOnly(ctx, op)
}

func (op MakeAddedSecondaryIndexPublic) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582419)
	return v.MakeAddedSecondaryIndexPublic(ctx, op)
}

func (op MakeAddedPrimaryIndexPublic) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582420)
	return v.MakeAddedPrimaryIndexPublic(ctx, op)
}

func (op MakeDroppedPrimaryIndexDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582421)
	return v.MakeDroppedPrimaryIndexDeleteAndWriteOnly(ctx, op)
}

func (op CreateGcJobForTable) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582422)
	return v.CreateGcJobForTable(ctx, op)
}

func (op CreateGcJobForDatabase) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582423)
	return v.CreateGcJobForDatabase(ctx, op)
}

func (op CreateGcJobForIndex) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582424)
	return v.CreateGcJobForIndex(ctx, op)
}

func (op MarkDescriptorAsDroppedSynthetically) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582425)
	return v.MarkDescriptorAsDroppedSynthetically(ctx, op)
}

func (op MarkDescriptorAsDropped) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582426)
	return v.MarkDescriptorAsDropped(ctx, op)
}

func (op DrainDescriptorName) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582427)
	return v.DrainDescriptorName(ctx, op)
}

func (op MakeAddedColumnDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582428)
	return v.MakeAddedColumnDeleteAndWriteOnly(ctx, op)
}

func (op MakeDroppedNonPrimaryIndexDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582429)
	return v.MakeDroppedNonPrimaryIndexDeleteAndWriteOnly(ctx, op)
}

func (op MakeDroppedIndexDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582430)
	return v.MakeDroppedIndexDeleteOnly(ctx, op)
}

func (op RemoveDroppedIndexPartialPredicate) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582431)
	return v.RemoveDroppedIndexPartialPredicate(ctx, op)
}

func (op MakeIndexAbsent) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582432)
	return v.MakeIndexAbsent(ctx, op)
}

func (op MakeAddedColumnDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582433)
	return v.MakeAddedColumnDeleteOnly(ctx, op)
}

func (op SetAddedColumnType) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582434)
	return v.SetAddedColumnType(ctx, op)
}

func (op MakeColumnPublic) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582435)
	return v.MakeColumnPublic(ctx, op)
}

func (op MakeDroppedColumnDeleteAndWriteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582436)
	return v.MakeDroppedColumnDeleteAndWriteOnly(ctx, op)
}

func (op MakeDroppedColumnDeleteOnly) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582437)
	return v.MakeDroppedColumnDeleteOnly(ctx, op)
}

func (op RemoveDroppedColumnType) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582438)
	return v.RemoveDroppedColumnType(ctx, op)
}

func (op MakeColumnAbsent) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582439)
	return v.MakeColumnAbsent(ctx, op)
}

func (op RemoveOwnerBackReferenceInSequence) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582440)
	return v.RemoveOwnerBackReferenceInSequence(ctx, op)
}

func (op RemoveSequenceOwner) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582441)
	return v.RemoveSequenceOwner(ctx, op)
}

func (op RemoveCheckConstraint) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582442)
	return v.RemoveCheckConstraint(ctx, op)
}

func (op RemoveForeignKeyConstraint) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582443)
	return v.RemoveForeignKeyConstraint(ctx, op)
}

func (op RemoveForeignKeyBackReference) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582444)
	return v.RemoveForeignKeyBackReference(ctx, op)
}

func (op RemoveSchemaParent) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582445)
	return v.RemoveSchemaParent(ctx, op)
}

func (op AddIndexPartitionInfo) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582446)
	return v.AddIndexPartitionInfo(ctx, op)
}

func (op LogEvent) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582447)
	return v.LogEvent(ctx, op)
}

func (op AddColumnFamily) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582448)
	return v.AddColumnFamily(ctx, op)
}

func (op AddColumnDefaultExpression) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582449)
	return v.AddColumnDefaultExpression(ctx, op)
}

func (op RemoveColumnDefaultExpression) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582450)
	return v.RemoveColumnDefaultExpression(ctx, op)
}

func (op AddColumnOnUpdateExpression) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582451)
	return v.AddColumnOnUpdateExpression(ctx, op)
}

func (op RemoveColumnOnUpdateExpression) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582452)
	return v.RemoveColumnOnUpdateExpression(ctx, op)
}

func (op UpdateTableBackReferencesInTypes) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582453)
	return v.UpdateTableBackReferencesInTypes(ctx, op)
}

func (op RemoveBackReferenceInTypes) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582454)
	return v.RemoveBackReferenceInTypes(ctx, op)
}

func (op UpdateBackReferencesInSequences) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582455)
	return v.UpdateBackReferencesInSequences(ctx, op)
}

func (op RemoveViewBackReferencesInRelations) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582456)
	return v.RemoveViewBackReferencesInRelations(ctx, op)
}

func (op SetColumnName) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582457)
	return v.SetColumnName(ctx, op)
}

func (op SetIndexName) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582458)
	return v.SetIndexName(ctx, op)
}

func (op DeleteDescriptor) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582459)
	return v.DeleteDescriptor(ctx, op)
}

func (op RemoveJobStateFromDescriptor) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582460)
	return v.RemoveJobStateFromDescriptor(ctx, op)
}

func (op SetJobStateOnDescriptor) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582461)
	return v.SetJobStateOnDescriptor(ctx, op)
}

func (op UpdateSchemaChangerJob) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582462)
	return v.UpdateSchemaChangerJob(ctx, op)
}

func (op CreateSchemaChangerJob) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582463)
	return v.CreateSchemaChangerJob(ctx, op)
}

func (op RemoveAllTableComments) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582464)
	return v.RemoveAllTableComments(ctx, op)
}

func (op RemoveTableComment) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582465)
	return v.RemoveTableComment(ctx, op)
}

func (op RemoveDatabaseComment) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582466)
	return v.RemoveDatabaseComment(ctx, op)
}

func (op RemoveSchemaComment) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582467)
	return v.RemoveSchemaComment(ctx, op)
}

func (op RemoveIndexComment) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582468)
	return v.RemoveIndexComment(ctx, op)
}

func (op RemoveColumnComment) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582469)
	return v.RemoveColumnComment(ctx, op)
}

func (op RemoveConstraintComment) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582470)
	return v.RemoveConstraintComment(ctx, op)
}

func (op RemoveDatabaseRoleSettings) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582471)
	return v.RemoveDatabaseRoleSettings(ctx, op)
}

func (op DeleteSchedule) Visit(ctx context.Context, v MutationVisitor) error {
	__antithesis_instrumentation__.Notify(582472)
	return v.DeleteSchedule(ctx, op)
}
