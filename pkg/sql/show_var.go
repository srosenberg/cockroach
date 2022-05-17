package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type showVarNode struct {
	name  string
	shown bool
	val   string
}

func (s *showVarNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623348)
	return nil
}

func (s *showVarNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623349)
	if s.shown {
		__antithesis_instrumentation__.Notify(623352)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(623353)
	}
	__antithesis_instrumentation__.Notify(623350)
	s.shown = true

	_, v, err := getSessionVar(s.name, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(623354)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623355)
	}
	__antithesis_instrumentation__.Notify(623351)
	s.val, err = v.Get(params.extendedEvalCtx)
	return true, err
}

func (s *showVarNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623356)
	return tree.Datums{tree.NewDString(s.val)}
}

func (s *showVarNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(623357) }

func (p *planner) ShowVar(ctx context.Context, n *tree.ShowVar) (planNode, error) {
	__antithesis_instrumentation__.Notify(623358)
	return &showVarNode{name: strings.ToLower(n.Name)}, nil
}
