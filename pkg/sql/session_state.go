package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/lib/pq/oid"
)

var maxSerializedSessionSize = settings.RegisterByteSizeSetting(
	settings.TenantReadOnly,
	"sql.session_transfer.max_session_size",
	"if set to non-zero, then serializing a session will fail if it requires more"+
		"than the specified size",
	0,
	settings.NonNegativeInt,
)

func (p *planner) SerializeSessionState() (*tree.DBytes, error) {
	__antithesis_instrumentation__.Notify(617750)
	evalCtx := p.EvalContext()
	return serializeSessionState(
		!evalCtx.TxnIsSingleStmt,
		evalCtx.PreparedStatementState,
		p.SessionData(),
		p.ExecCfg(),
	)
}

func serializeSessionState(
	inExplicitTxn bool,
	prepStmtsState tree.PreparedStatementState,
	sd *sessiondata.SessionData,
	execCfg *ExecutorConfig,
) (*tree.DBytes, error) {
	__antithesis_instrumentation__.Notify(617751)
	if inExplicitTxn {
		__antithesis_instrumentation__.Notify(617758)
		return nil, pgerror.Newf(
			pgcode.InvalidTransactionState,
			"cannot serialize a session which is inside a transaction",
		)
	} else {
		__antithesis_instrumentation__.Notify(617759)
	}
	__antithesis_instrumentation__.Notify(617752)

	if prepStmtsState.HasActivePortals() {
		__antithesis_instrumentation__.Notify(617760)
		return nil, pgerror.Newf(
			pgcode.InvalidTransactionState,
			"cannot serialize a session which has active portals",
		)
	} else {
		__antithesis_instrumentation__.Notify(617761)
	}
	__antithesis_instrumentation__.Notify(617753)

	if sd == nil {
		__antithesis_instrumentation__.Notify(617762)
		return nil, pgerror.Newf(
			pgcode.InvalidTransactionState,
			"no session is active",
		)
	} else {
		__antithesis_instrumentation__.Notify(617763)
	}
	__antithesis_instrumentation__.Notify(617754)

	if len(sd.DatabaseIDToTempSchemaID) > 0 {
		__antithesis_instrumentation__.Notify(617764)
		return nil, pgerror.Newf(
			pgcode.InvalidTransactionState,
			"cannot serialize session with temporary schemas",
		)
	} else {
		__antithesis_instrumentation__.Notify(617765)
	}
	__antithesis_instrumentation__.Notify(617755)

	var m sessiondatapb.MigratableSession
	m.SessionData = sd.SessionData
	sessiondata.MarshalNonLocal(sd, &m.SessionData)
	m.LocalOnlySessionData = sd.LocalOnlySessionData
	m.PreparedStatements = prepStmtsState.MigratablePreparedStatements()

	b, err := protoutil.Marshal(&m)
	if err != nil {
		__antithesis_instrumentation__.Notify(617766)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617767)
	}
	__antithesis_instrumentation__.Notify(617756)

	maxSize := maxSerializedSessionSize.Get(&execCfg.Settings.SV)
	if maxSize > 0 && func() bool {
		__antithesis_instrumentation__.Notify(617768)
		return int64(len(b)) > maxSize == true
	}() == true {
		__antithesis_instrumentation__.Notify(617769)
		return nil, pgerror.Newf(
			pgcode.ProgramLimitExceeded,
			"serialized session size %s exceeds max allowed size %s",
			humanizeutil.IBytes(int64(len(b))),
			humanizeutil.IBytes(maxSize),
		)
	} else {
		__antithesis_instrumentation__.Notify(617770)
	}
	__antithesis_instrumentation__.Notify(617757)

	return tree.NewDBytes(tree.DBytes(b)), nil
}

func (p *planner) DeserializeSessionState(state *tree.DBytes) (*tree.DBool, error) {
	__antithesis_instrumentation__.Notify(617771)
	evalCtx := p.ExtendedEvalContext()
	if !evalCtx.TxnIsSingleStmt {
		__antithesis_instrumentation__.Notify(617778)
		return nil, pgerror.Newf(
			pgcode.InvalidTransactionState,
			"cannot deserialize a session whilst inside a multi-statement transaction",
		)
	} else {
		__antithesis_instrumentation__.Notify(617779)
	}
	__antithesis_instrumentation__.Notify(617772)

	var m sessiondatapb.MigratableSession
	if err := protoutil.Unmarshal([]byte(*state), &m); err != nil {
		__antithesis_instrumentation__.Notify(617780)
		return nil, pgerror.Wrapf(err, pgcode.InvalidParameterValue, "error deserializing session")
	} else {
		__antithesis_instrumentation__.Notify(617781)
	}
	__antithesis_instrumentation__.Notify(617773)
	sd, err := sessiondata.UnmarshalNonLocal(m.SessionData)
	if err != nil {
		__antithesis_instrumentation__.Notify(617782)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617783)
	}
	__antithesis_instrumentation__.Notify(617774)
	sd.SessionData = m.SessionData
	sd.LocalUnmigratableSessionData = evalCtx.SessionData().LocalUnmigratableSessionData
	sd.LocalOnlySessionData = m.LocalOnlySessionData
	if sd.SessionUser().Normalized() != evalCtx.SessionData().SessionUser().Normalized() {
		__antithesis_instrumentation__.Notify(617784)
		return nil, pgerror.Newf(
			pgcode.InsufficientPrivilege,
			"can only deserialize matching session users",
		)
	} else {
		__antithesis_instrumentation__.Notify(617785)
	}
	__antithesis_instrumentation__.Notify(617775)
	if err := p.checkCanBecomeUser(evalCtx.Context, sd.User()); err != nil {
		__antithesis_instrumentation__.Notify(617786)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(617787)
	}
	__antithesis_instrumentation__.Notify(617776)

	for _, prepStmt := range m.PreparedStatements {
		__antithesis_instrumentation__.Notify(617788)
		parserStmt, err := parser.ParseOneWithInt(
			prepStmt.SQL,
			parser.NakedIntTypeFromDefaultIntSize(sd.DefaultIntSize),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(617791)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617792)
		}
		__antithesis_instrumentation__.Notify(617789)
		id := GenerateClusterWideID(evalCtx.ExecCfg.Clock.Now(), evalCtx.ExecCfg.NodeID.SQLInstanceID())
		stmt := makeStatement(parserStmt, id)

		var placeholderTypes tree.PlaceholderTypes
		if len(prepStmt.PlaceholderTypeHints) > 0 {
			__antithesis_instrumentation__.Notify(617793)

			placeholderTypes = make(tree.PlaceholderTypes, stmt.NumPlaceholders)
			for i, t := range prepStmt.PlaceholderTypeHints {
				__antithesis_instrumentation__.Notify(617794)

				if t == 0 || func() bool {
					__antithesis_instrumentation__.Notify(617797)
					return t == oid.T_unknown == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(617798)
					return types.IsOIDUserDefinedType(t) == true
				}() == true {
					__antithesis_instrumentation__.Notify(617799)
					placeholderTypes[i] = nil
					continue
				} else {
					__antithesis_instrumentation__.Notify(617800)
				}
				__antithesis_instrumentation__.Notify(617795)
				v, ok := types.OidToType[t]
				if !ok {
					__antithesis_instrumentation__.Notify(617801)
					err := pgwirebase.NewProtocolViolationErrorf("unknown oid type: %v", t)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(617802)
				}
				__antithesis_instrumentation__.Notify(617796)
				placeholderTypes[i] = v
			}
		} else {
			__antithesis_instrumentation__.Notify(617803)
		}
		__antithesis_instrumentation__.Notify(617790)

		_, err = evalCtx.statementPreparer.addPreparedStmt(
			evalCtx.Context,
			prepStmt.Name, stmt, placeholderTypes, prepStmt.PlaceholderTypeHints,
			PreparedStatementOriginSessionMigration,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(617804)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(617805)
		}
	}
	__antithesis_instrumentation__.Notify(617777)

	*p.SessionData() = *sd

	return tree.MakeDBool(true), nil
}
