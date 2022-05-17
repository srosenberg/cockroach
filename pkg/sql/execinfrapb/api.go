package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type ProcessorID int

type StreamID int

func (sid StreamID) String() string {
	__antithesis_instrumentation__.Notify(471521)
	return strconv.Itoa(int(sid))
}

type FlowID struct {
	uuid.UUID
}

func (f FlowID) Equal(other FlowID) bool {
	__antithesis_instrumentation__.Notify(471522)
	return f.UUID.Equal(other.UUID)
}

func (f FlowID) IsUnset() bool {
	__antithesis_instrumentation__.Notify(471523)
	return f.UUID.Equal(uuid.Nil)
}

type DistSQLVersion uint32

func MakeEvalContext(evalCtx *tree.EvalContext) EvalContext {
	__antithesis_instrumentation__.Notify(471524)
	sessionDataProto := evalCtx.SessionData().SessionData
	sessiondata.MarshalNonLocal(evalCtx.SessionData(), &sessionDataProto)
	return EvalContext{
		SessionData:        sessionDataProto,
		StmtTimestampNanos: evalCtx.StmtTimestamp.UnixNano(),
		TxnTimestampNanos:  evalCtx.TxnTimestamp.UnixNano(),
	}
}

func (m *BackupDataSpec) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(471525)
	return m.UserProto.Decode()
}

func (m *ExportSpec) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(471526)
	return m.UserProto.Decode()
}

func (m *ReadImportDataSpec) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(471527)
	return m.UserProto.Decode()
}

func (m *ChangeAggregatorSpec) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(471528)
	return m.UserProto.Decode()
}

func (m *ChangeFrontierSpec) User() security.SQLUsername {
	__antithesis_instrumentation__.Notify(471529)
	return m.UserProto.Decode()
}
