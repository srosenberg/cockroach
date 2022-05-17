package scbuild

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild/internal/scbuildstmt"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
)

var _ scbuildstmt.EventLogState = (*eventLogState)(nil)

func (e *eventLogState) TargetMetadata() scpb.TargetMetadata {
	__antithesis_instrumentation__.Notify(579720)
	return e.statementMetaData
}

func (e *eventLogState) IncrementSubWorkID() {
	__antithesis_instrumentation__.Notify(579721)
	e.statementMetaData.SubWorkID++
}

func (e *eventLogState) EventLogStateWithNewSourceElementID() scbuildstmt.EventLogState {
	__antithesis_instrumentation__.Notify(579722)
	*e.sourceElementID++
	return &eventLogState{
		statements:      e.statements,
		authorization:   e.authorization,
		sourceElementID: e.sourceElementID,
		statementMetaData: scpb.TargetMetadata{
			StatementID:     e.statementMetaData.StatementID,
			SubWorkID:       e.statementMetaData.SubWorkID,
			SourceElementID: *e.sourceElementID,
		},
	}
}
