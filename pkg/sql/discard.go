package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func (p *planner) Discard(ctx context.Context, s *tree.Discard) (planNode, error) {
	__antithesis_instrumentation__.Notify(466177)
	switch s.Mode {
	case tree.DiscardModeAll:
		__antithesis_instrumentation__.Notify(466179)
		if !p.autoCommit {
			__antithesis_instrumentation__.Notify(466183)
			return nil, pgerror.New(pgcode.ActiveSQLTransaction,
				"DISCARD ALL cannot run inside a transaction block")
		} else {
			__antithesis_instrumentation__.Notify(466184)
		}
		__antithesis_instrumentation__.Notify(466180)

		if err := p.sessionDataMutatorIterator.applyOnEachMutatorError(
			func(m sessionDataMutator) error {
				__antithesis_instrumentation__.Notify(466185)
				return resetSessionVars(ctx, m)
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(466186)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(466187)
		}
		__antithesis_instrumentation__.Notify(466181)

		p.preparedStatements.DeleteAll(ctx)
	default:
		__antithesis_instrumentation__.Notify(466182)
		return nil, errors.AssertionFailedf("unknown mode for DISCARD: %d", s.Mode)
	}
	__antithesis_instrumentation__.Notify(466178)
	return newZeroNode(nil), nil
}

func resetSessionVars(ctx context.Context, m sessionDataMutator) error {
	__antithesis_instrumentation__.Notify(466188)

	if err := resetSessionVar(ctx, m, "datestyle_enabled"); err != nil {
		__antithesis_instrumentation__.Notify(466192)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466193)
	}
	__antithesis_instrumentation__.Notify(466189)
	if err := resetSessionVar(ctx, m, "intervalstyle_enabled"); err != nil {
		__antithesis_instrumentation__.Notify(466194)
		return err
	} else {
		__antithesis_instrumentation__.Notify(466195)
	}
	__antithesis_instrumentation__.Notify(466190)
	for _, varName := range varNames {
		__antithesis_instrumentation__.Notify(466196)
		if err := resetSessionVar(ctx, m, varName); err != nil {
			__antithesis_instrumentation__.Notify(466197)
			return err
		} else {
			__antithesis_instrumentation__.Notify(466198)
		}
	}
	__antithesis_instrumentation__.Notify(466191)
	return nil
}

func resetSessionVar(ctx context.Context, m sessionDataMutator, varName string) error {
	__antithesis_instrumentation__.Notify(466199)
	v := varGen[varName]
	if v.Set != nil {
		__antithesis_instrumentation__.Notify(466201)
		hasDefault, defVal := getSessionVarDefaultString(varName, v, m.sessionDataMutatorBase)
		if hasDefault {
			__antithesis_instrumentation__.Notify(466202)
			if err := v.Set(ctx, m, defVal); err != nil {
				__antithesis_instrumentation__.Notify(466203)
				return err
			} else {
				__antithesis_instrumentation__.Notify(466204)
			}
		} else {
			__antithesis_instrumentation__.Notify(466205)
		}
	} else {
		__antithesis_instrumentation__.Notify(466206)
	}
	__antithesis_instrumentation__.Notify(466200)
	return nil
}
