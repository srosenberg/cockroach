package scmutationexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func (m *visitor) LogEvent(ctx context.Context, op scop.LogEvent) error {
	__antithesis_instrumentation__.Notify(581876)
	descID := screl.GetDescID(op.Element.Element())
	fullName, err := m.nr.GetFullyQualifiedName(ctx, descID)
	if err != nil {
		__antithesis_instrumentation__.Notify(581879)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581880)
	}
	__antithesis_instrumentation__.Notify(581877)
	event, err := asEventPayload(ctx, fullName, op.Element.Element(), op.TargetStatus, m)
	if err != nil {
		__antithesis_instrumentation__.Notify(581881)
		return err
	} else {
		__antithesis_instrumentation__.Notify(581882)
	}
	__antithesis_instrumentation__.Notify(581878)
	details := eventpb.CommonSQLEventDetails{
		ApplicationName: op.Authorization.AppName,
		User:            op.Authorization.UserName,
		Statement:       redact.RedactableString(op.Statement),
		Tag:             op.StatementTag,
	}
	return m.s.EnqueueEvent(descID, op.TargetMetadata, details, event)
}

func asEventPayload(
	ctx context.Context, fullName string, e scpb.Element, targetStatus scpb.Status, m *visitor,
) (eventpb.EventPayload, error) {
	__antithesis_instrumentation__.Notify(581883)
	if targetStatus == scpb.Status_ABSENT {
		__antithesis_instrumentation__.Notify(581886)
		switch e.(type) {
		case *scpb.Table:
			__antithesis_instrumentation__.Notify(581887)
			return &eventpb.DropTable{TableName: fullName}, nil
		case *scpb.View:
			__antithesis_instrumentation__.Notify(581888)
			return &eventpb.DropView{ViewName: fullName}, nil
		case *scpb.Sequence:
			__antithesis_instrumentation__.Notify(581889)
			return &eventpb.DropSequence{SequenceName: fullName}, nil
		case *scpb.Database:
			__antithesis_instrumentation__.Notify(581890)
			return &eventpb.DropDatabase{DatabaseName: fullName}, nil
		case *scpb.Schema:
			__antithesis_instrumentation__.Notify(581891)
			return &eventpb.DropSchema{SchemaName: fullName}, nil
		case *scpb.AliasType, *scpb.EnumType:
			__antithesis_instrumentation__.Notify(581892)
			return &eventpb.DropType{TypeName: fullName}, nil
		}
	} else {
		__antithesis_instrumentation__.Notify(581893)
	}
	__antithesis_instrumentation__.Notify(581884)
	switch e := e.(type) {
	case *scpb.Column:
		__antithesis_instrumentation__.Notify(581894)
		tbl, err := m.checkOutTable(ctx, e.TableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(581900)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581901)
		}
		__antithesis_instrumentation__.Notify(581895)
		mutation, err := FindMutation(tbl, MakeColumnIDMutationSelector(e.ColumnID))
		if err != nil {
			__antithesis_instrumentation__.Notify(581902)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581903)
		}
		__antithesis_instrumentation__.Notify(581896)
		return &eventpb.AlterTable{
			TableName:  fullName,
			MutationID: uint32(mutation.MutationID()),
		}, nil
	case *scpb.SecondaryIndex:
		__antithesis_instrumentation__.Notify(581897)
		tbl, err := m.checkOutTable(ctx, e.TableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(581904)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581905)
		}
		__antithesis_instrumentation__.Notify(581898)
		mutation, err := FindMutation(tbl, MakeIndexIDMutationSelector(e.IndexID))
		if err != nil {
			__antithesis_instrumentation__.Notify(581906)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(581907)
		}
		__antithesis_instrumentation__.Notify(581899)
		switch targetStatus {
		case scpb.Status_PUBLIC:
			__antithesis_instrumentation__.Notify(581908)
			return &eventpb.AlterTable{
				TableName:  fullName,
				MutationID: uint32(mutation.MutationID()),
			}, nil
		case scpb.Status_ABSENT:
			__antithesis_instrumentation__.Notify(581909)
			return &eventpb.DropIndex{
				TableName:  fullName,
				IndexName:  mutation.AsIndex().GetName(),
				MutationID: uint32(mutation.MutationID()),
			}, nil
		default:
			__antithesis_instrumentation__.Notify(581910)
			return nil, errors.AssertionFailedf("unknown target status %s", targetStatus)
		}
	}
	__antithesis_instrumentation__.Notify(581885)
	return nil, errors.AssertionFailedf("unknown %s element type %T", targetStatus.String(), e)
}
