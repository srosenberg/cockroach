package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type max1RowNode struct {
	plan planNode

	nexted    bool
	values    tree.Datums
	errorText string
}

func (m *max1RowNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(501751)
	return nil
}

func (m *max1RowNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(501752)
	if m.nexted {
		__antithesis_instrumentation__.Notify(501756)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(501757)
	}
	__antithesis_instrumentation__.Notify(501753)
	m.nexted = true

	ok, err := m.plan.Next(params)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(501758)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(501759)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(501760)
	}
	__antithesis_instrumentation__.Notify(501754)
	if ok {
		__antithesis_instrumentation__.Notify(501761)

		m.values = make(tree.Datums, len(m.plan.Values()))
		copy(m.values, m.plan.Values())
		var secondOk bool
		secondOk, err = m.plan.Next(params)
		if secondOk {
			__antithesis_instrumentation__.Notify(501762)

			return false, pgerror.Newf(pgcode.CardinalityViolation, "%s", m.errorText)
		} else {
			__antithesis_instrumentation__.Notify(501763)
		}
	} else {
		__antithesis_instrumentation__.Notify(501764)
	}
	__antithesis_instrumentation__.Notify(501755)
	return ok, err
}

func (m *max1RowNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(501765)
	return m.values
}

func (m *max1RowNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(501766)
	m.plan.Close(ctx)
}
