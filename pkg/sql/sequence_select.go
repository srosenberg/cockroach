package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type sequenceSelectNode struct {
	optColumnsSlot

	desc catalog.TableDescriptor

	val  int64
	done bool
}

var _ planNode = &sequenceSelectNode{}

func (p *planner) SequenceSelectNode(desc catalog.TableDescriptor) (planNode, error) {
	__antithesis_instrumentation__.Notify(617631)
	if desc.GetSequenceOpts() == nil {
		__antithesis_instrumentation__.Notify(617633)
		return nil, errors.New("descriptor is not a sequence")
	} else {
		__antithesis_instrumentation__.Notify(617634)
	}
	__antithesis_instrumentation__.Notify(617632)
	return &sequenceSelectNode{
		desc: desc,
	}, nil
}

func (ss *sequenceSelectNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(617635)
	return nil
}

func (ss *sequenceSelectNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(617636)
	if ss.done {
		__antithesis_instrumentation__.Notify(617639)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(617640)
	}
	__antithesis_instrumentation__.Notify(617637)
	val, err := params.p.GetSequenceValue(params.ctx, params.ExecCfg().Codec, ss.desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(617641)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(617642)
	}
	__antithesis_instrumentation__.Notify(617638)
	ss.val = val
	ss.done = true
	return true, nil
}

func (ss *sequenceSelectNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(617643)
	valDatum := tree.DInt(ss.val)
	cntDatum := tree.DInt(0)
	calledDatum := tree.DBoolTrue
	return []tree.Datum{
		&valDatum,
		&cntDatum,
		calledDatum,
	}
}

func (ss *sequenceSelectNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(617644)
}
