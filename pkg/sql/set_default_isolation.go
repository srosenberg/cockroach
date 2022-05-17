package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
)

func (p *planner) SetSessionCharacteristics(n *tree.SetSessionCharacteristics) (planNode, error) {
	__antithesis_instrumentation__.Notify(621997)

	switch n.Modes.Isolation {
	case tree.SerializableIsolation, tree.UnspecifiedIsolation:
		__antithesis_instrumentation__.Notify(622001)

	default:
		__antithesis_instrumentation__.Notify(622002)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue,
			"unsupported default isolation level: %s", n.Modes.Isolation)
	}
	__antithesis_instrumentation__.Notify(621998)

	if err := p.sessionDataMutatorIterator.applyOnEachMutatorError(func(m sessionDataMutator) error {
		__antithesis_instrumentation__.Notify(622003)

		switch n.Modes.UserPriority {
		case tree.UnspecifiedUserPriority:
			__antithesis_instrumentation__.Notify(622007)
		default:
			__antithesis_instrumentation__.Notify(622008)
			m.SetDefaultTransactionPriority(n.Modes.UserPriority)
		}
		__antithesis_instrumentation__.Notify(622004)

		switch n.Modes.ReadWriteMode {
		case tree.ReadOnly:
			__antithesis_instrumentation__.Notify(622009)
			m.SetDefaultTransactionReadOnly(true)
		case tree.ReadWrite:
			__antithesis_instrumentation__.Notify(622010)
			m.SetDefaultTransactionReadOnly(false)
		case tree.UnspecifiedReadWriteMode:
			__antithesis_instrumentation__.Notify(622011)
		default:
			__antithesis_instrumentation__.Notify(622012)
			return pgerror.Newf(pgcode.InvalidParameterValue,
				"unsupported default read write mode: %s", n.Modes.ReadWriteMode)
		}
		__antithesis_instrumentation__.Notify(622005)

		if n.Modes.AsOf.Expr != nil {
			__antithesis_instrumentation__.Notify(622013)
			if tree.IsFollowerReadTimestampFunction(n.Modes.AsOf, p.semaCtx.SearchPath) {
				__antithesis_instrumentation__.Notify(622014)
				m.SetDefaultTransactionUseFollowerReads(true)
			} else {
				__antithesis_instrumentation__.Notify(622015)
				return pgerror.Newf(pgcode.InvalidParameterValue,
					"unsupported default as of system time expression, only %s() allowed",
					tree.FollowerReadTimestampFunctionName)
			}
		} else {
			__antithesis_instrumentation__.Notify(622016)
		}
		__antithesis_instrumentation__.Notify(622006)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(622017)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(622018)
	}
	__antithesis_instrumentation__.Notify(621999)

	switch n.Modes.Deferrable {
	case tree.NotDeferrable, tree.UnspecifiedDeferrableMode:
		__antithesis_instrumentation__.Notify(622019)

	case tree.Deferrable:
		__antithesis_instrumentation__.Notify(622020)
		return nil, unimplemented.NewWithIssue(53432, "DEFERRABLE transactions")
	default:
		__antithesis_instrumentation__.Notify(622021)
		return nil, pgerror.Newf(pgcode.InvalidParameterValue,
			"unsupported default deferrable mode: %s", n.Modes.Deferrable)
	}
	__antithesis_instrumentation__.Notify(622000)

	return newZeroNode(nil), nil
}
