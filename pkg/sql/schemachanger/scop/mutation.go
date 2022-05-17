package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
)

type mutationOp struct{ baseOp }

var _ = mutationOp{baseOp: baseOp{}}

func (mutationOp) Type() Type { __antithesis_instrumentation__.Notify(582414); return MutationType }

type NotImplemented struct {
	mutationOp
	ElementType string
}

type MakeAddedIndexDeleteOnly struct {
	mutationOp
	Index              scpb.Index
	IsSecondaryIndex   bool
	IsDeletePreserving bool
}

type SetAddedIndexPartialPredicate struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
	Expr    catpb.Expression
}

type MakeAddedIndexDeleteAndWriteOnly struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type MakeAddedSecondaryIndexPublic struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type MakeAddedPrimaryIndexPublic struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type MakeDroppedPrimaryIndexDeleteAndWriteOnly struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type CreateGcJobForTable struct {
	mutationOp
	TableID descpb.ID
	StatementForDropJob
}

type CreateGcJobForDatabase struct {
	mutationOp
	DatabaseID descpb.ID
	StatementForDropJob
}

type CreateGcJobForIndex struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
	StatementForDropJob
}

type MarkDescriptorAsDroppedSynthetically struct {
	mutationOp
	DescID descpb.ID
}

type MarkDescriptorAsDropped struct {
	mutationOp
	DescID descpb.ID
}

type DrainDescriptorName struct {
	mutationOp
	Namespace scpb.Namespace
}

type MakeAddedColumnDeleteAndWriteOnly struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type MakeDroppedNonPrimaryIndexDeleteAndWriteOnly struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type MakeDroppedIndexDeleteOnly struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type RemoveDroppedIndexPartialPredicate struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type MakeIndexAbsent struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type MakeAddedColumnDeleteOnly struct {
	mutationOp
	Column scpb.Column
}

type SetAddedColumnType struct {
	mutationOp
	ColumnType scpb.ColumnType
}

type MakeColumnPublic struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type MakeDroppedColumnDeleteAndWriteOnly struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type MakeDroppedColumnDeleteOnly struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type RemoveDroppedColumnType struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type MakeColumnAbsent struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type RemoveOwnerBackReferenceInSequence struct {
	mutationOp
	SequenceID descpb.ID
}

type RemoveSequenceOwner struct {
	mutationOp
	TableID         descpb.ID
	ColumnID        descpb.ColumnID
	OwnedSequenceID descpb.ID
}

type RemoveCheckConstraint struct {
	mutationOp
	TableID      descpb.ID
	ConstraintID descpb.ConstraintID
}

type RemoveForeignKeyConstraint struct {
	mutationOp
	TableID      descpb.ID
	ConstraintID descpb.ConstraintID
}

type RemoveForeignKeyBackReference struct {
	mutationOp
	ReferencedTableID  descpb.ID
	OriginTableID      descpb.ID
	OriginConstraintID descpb.ConstraintID
}

type RemoveSchemaParent struct {
	mutationOp
	Parent scpb.SchemaParent
}

type AddIndexPartitionInfo struct {
	mutationOp
	Partitioning scpb.IndexPartitioning
}

type LogEvent struct {
	mutationOp
	TargetMetadata scpb.TargetMetadata
	Authorization  scpb.Authorization
	Statement      string
	StatementTag   string
	Element        scpb.ElementProto
	TargetStatus   scpb.Status
}

type AddColumnFamily struct {
	mutationOp
	TableID  descpb.ID
	FamilyID descpb.FamilyID
	Name     string
}

type AddColumnDefaultExpression struct {
	mutationOp
	Default scpb.ColumnDefaultExpression
}

type RemoveColumnDefaultExpression struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type AddColumnOnUpdateExpression struct {
	mutationOp
	OnUpdate scpb.ColumnOnUpdateExpression
}

type RemoveColumnOnUpdateExpression struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type UpdateTableBackReferencesInTypes struct {
	mutationOp
	TypeIDs               []descpb.ID
	BackReferencedTableID descpb.ID
}

type RemoveBackReferenceInTypes struct {
	mutationOp
	BackReferencedDescID descpb.ID
	TypeIDs              []descpb.ID
}

type UpdateBackReferencesInSequences struct {
	mutationOp
	BackReferencedTableID  descpb.ID
	BackReferencedColumnID descpb.ColumnID
	SequenceIDs            []descpb.ID
}

type RemoveViewBackReferencesInRelations struct {
	mutationOp
	BackReferencedViewID descpb.ID
	RelationIDs          []descpb.ID
}

type SetColumnName struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
	Name     string
}

type SetIndexName struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
	Name    string
}

type DeleteDescriptor struct {
	mutationOp
	DescriptorID descpb.ID
}

type RemoveJobStateFromDescriptor struct {
	mutationOp
	DescriptorID descpb.ID
	JobID        jobspb.JobID
}

type SetJobStateOnDescriptor struct {
	mutationOp
	DescriptorID descpb.ID

	Initialize bool
	State      scpb.DescriptorState
}

type UpdateSchemaChangerJob struct {
	mutationOp
	IsNonCancelable bool
	JobID           jobspb.JobID
	RunningStatus   string
}

type CreateSchemaChangerJob struct {
	mutationOp
	JobID         jobspb.JobID
	Authorization scpb.Authorization
	Statements    []scpb.Statement
	DescriptorIDs []descpb.ID

	NonCancelable bool
	RunningStatus string
}

type RemoveAllTableComments struct {
	mutationOp
	TableID descpb.ID
}

type RemoveTableComment struct {
	mutationOp
	TableID descpb.ID
}

type RemoveDatabaseComment struct {
	mutationOp
	DatabaseID descpb.ID
}

type RemoveSchemaComment struct {
	mutationOp
	SchemaID descpb.ID
}

type RemoveIndexComment struct {
	mutationOp
	TableID descpb.ID
	IndexID descpb.IndexID
}

type RemoveColumnComment struct {
	mutationOp
	TableID  descpb.ID
	ColumnID descpb.ColumnID
}

type RemoveConstraintComment struct {
	mutationOp
	TableID      descpb.ID
	ConstraintID descpb.ConstraintID
}

type RemoveDatabaseRoleSettings struct {
	mutationOp
	DatabaseID descpb.ID
}

type DeleteSchedule struct {
	mutationOp
	ScheduleID int64
}
