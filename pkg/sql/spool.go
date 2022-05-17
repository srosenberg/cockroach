package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type spoolNode struct {
	source    planNode
	rows      *rowcontainer.RowContainer
	hardLimit int64
	curRowIdx int
}

var _ mutationPlanNode = &spoolNode{}

func (s *spoolNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623714)

	if f, ok := s.source.(planNodeFastPath); ok {
		__antithesis_instrumentation__.Notify(623717)
		_, done := f.FastPathResults()
		if done {
			__antithesis_instrumentation__.Notify(623718)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(623719)
		}
	} else {
		__antithesis_instrumentation__.Notify(623720)
	}
	__antithesis_instrumentation__.Notify(623715)

	s.rows = rowcontainer.NewRowContainer(
		params.EvalContext().Mon.MakeBoundAccount(),
		colinfo.ColTypeInfoFromResCols(planColumns(s.source)),
	)

	for {
		__antithesis_instrumentation__.Notify(623721)
		next, err := s.source.Next(params)
		if err != nil {
			__antithesis_instrumentation__.Notify(623724)
			return err
		} else {
			__antithesis_instrumentation__.Notify(623725)
		}
		__antithesis_instrumentation__.Notify(623722)
		if !next {
			__antithesis_instrumentation__.Notify(623726)
			break
		} else {
			__antithesis_instrumentation__.Notify(623727)
		}
		__antithesis_instrumentation__.Notify(623723)
		if s.hardLimit == 0 || func() bool {
			__antithesis_instrumentation__.Notify(623728)
			return int64(s.rows.Len()) < s.hardLimit == true
		}() == true {
			__antithesis_instrumentation__.Notify(623729)
			if _, err := s.rows.AddRow(params.ctx, s.source.Values()); err != nil {
				__antithesis_instrumentation__.Notify(623730)
				return err
			} else {
				__antithesis_instrumentation__.Notify(623731)
			}
		} else {
			__antithesis_instrumentation__.Notify(623732)
		}
	}
	__antithesis_instrumentation__.Notify(623716)
	s.curRowIdx = -1
	return nil
}

func (s *spoolNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(623733)

	if f, ok := s.source.(planNodeFastPath); ok {
		__antithesis_instrumentation__.Notify(623735)
		return f.FastPathResults()
	} else {
		__antithesis_instrumentation__.Notify(623736)
	}
	__antithesis_instrumentation__.Notify(623734)
	return 0, false
}

func (s *spoolNode) spooled() { __antithesis_instrumentation__.Notify(623737) }

func (s *spoolNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623738)
	s.curRowIdx++
	return s.curRowIdx < s.rows.Len(), nil
}

func (s *spoolNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623739)
	return s.rows.At(s.curRowIdx)
}

func (s *spoolNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623740)
	s.source.Close(ctx)
	if s.rows != nil {
		__antithesis_instrumentation__.Notify(623741)
		s.rows.Close(ctx)
		s.rows = nil
	} else {
		__antithesis_instrumentation__.Notify(623742)
	}
}

func (s *spoolNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(623743)
	m, ok := s.source.(mutationPlanNode)
	if !ok {
		__antithesis_instrumentation__.Notify(623745)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(623746)
	}
	__antithesis_instrumentation__.Notify(623744)
	return m.rowsWritten()
}
