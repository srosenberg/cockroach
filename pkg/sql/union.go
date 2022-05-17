package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type unionNode struct {
	right, left planNode

	columns colinfo.ResultColumns

	inverted bool

	emitAll bool

	unionType tree.UnionType

	all bool

	streamingOrdering colinfo.ColumnOrdering

	reqOrdering ReqOrdering

	hardLimit uint64
}

func (p *planner) newUnionNode(
	typ tree.UnionType,
	all bool,
	left, right planNode,
	streamingOrdering colinfo.ColumnOrdering,
	reqOrdering ReqOrdering,
	hardLimit uint64,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(631219)
	emitAll := false
	switch typ {
	case tree.UnionOp:
		__antithesis_instrumentation__.Notify(631224)
		if all {
			__antithesis_instrumentation__.Notify(631228)
			emitAll = true
		} else {
			__antithesis_instrumentation__.Notify(631229)
		}
	case tree.IntersectOp:
		__antithesis_instrumentation__.Notify(631225)
	case tree.ExceptOp:
		__antithesis_instrumentation__.Notify(631226)
	default:
		__antithesis_instrumentation__.Notify(631227)
		return nil, errors.Errorf("%v is not supported", typ)
	}
	__antithesis_instrumentation__.Notify(631220)

	leftColumns := planColumns(left)
	rightColumns := planColumns(right)
	if len(leftColumns) != len(rightColumns) {
		__antithesis_instrumentation__.Notify(631230)
		return nil, pgerror.Newf(
			pgcode.Syntax,
			"each %v query must have the same number of columns: %d vs %d",
			typ, len(leftColumns), len(rightColumns),
		)
	} else {
		__antithesis_instrumentation__.Notify(631231)
	}
	__antithesis_instrumentation__.Notify(631221)
	unionColumns := append(colinfo.ResultColumns(nil), leftColumns...)
	for i := 0; i < len(unionColumns); i++ {
		__antithesis_instrumentation__.Notify(631232)
		l := leftColumns[i]
		r := rightColumns[i]

		if !(l.Typ.Equivalent(r.Typ) || func() bool {
			__antithesis_instrumentation__.Notify(631234)
			return l.Typ.Family() == types.UnknownFamily == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(631235)
			return r.Typ.Family() == types.UnknownFamily == true
		}() == true) {
			__antithesis_instrumentation__.Notify(631236)
			return nil, pgerror.Newf(pgcode.DatatypeMismatch,
				"%v types %s and %s cannot be matched", typ, l.Typ, r.Typ)
		} else {
			__antithesis_instrumentation__.Notify(631237)
		}
		__antithesis_instrumentation__.Notify(631233)
		if l.Typ.Family() == types.UnknownFamily {
			__antithesis_instrumentation__.Notify(631238)
			unionColumns[i].Typ = r.Typ
		} else {
			__antithesis_instrumentation__.Notify(631239)
		}
	}
	__antithesis_instrumentation__.Notify(631222)

	inverted := false
	if typ != tree.ExceptOp {
		__antithesis_instrumentation__.Notify(631240)

		left, right = right, left
		inverted = true
	} else {
		__antithesis_instrumentation__.Notify(631241)
	}
	__antithesis_instrumentation__.Notify(631223)

	node := &unionNode{
		right:             right,
		left:              left,
		columns:           unionColumns,
		inverted:          inverted,
		emitAll:           emitAll,
		unionType:         typ,
		all:               all,
		streamingOrdering: streamingOrdering,
		reqOrdering:       reqOrdering,
		hardLimit:         hardLimit,
	}
	return node, nil
}

func (n *unionNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(631242)
	panic("unionNode cannot be run in local mode")
}

func (n *unionNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631243)
	panic("unionNode cannot be run in local mode")
}

func (n *unionNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(631244)
	panic("unionNode cannot be run in local mode")
}

func (n *unionNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(631245)
	n.right.Close(ctx)
	n.left.Close(ctx)
}
