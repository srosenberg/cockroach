package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type batchedPlanNode interface {
	planNode

	BatchedNext(params runParams) (bool, error)

	BatchedCount() int

	BatchedValues(rowIdx int) tree.Datums
}

var _ batchedPlanNode = &deleteNode{}
var _ batchedPlanNode = &updateNode{}

type serializeNode struct {
	source batchedPlanNode

	fastPath bool

	rowCount int

	rowIdx int
}

var _ mutationPlanNode = &serializeNode{}

func (s *serializeNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(562615)
	if f, ok := s.source.(planNodeFastPath); ok {
		__antithesis_instrumentation__.Notify(562617)
		s.rowCount, s.fastPath = f.FastPathResults()
	} else {
		__antithesis_instrumentation__.Notify(562618)
	}
	__antithesis_instrumentation__.Notify(562616)
	return nil
}

func (s *serializeNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(562619)
	if s.fastPath {
		__antithesis_instrumentation__.Notify(562622)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(562623)
	}
	__antithesis_instrumentation__.Notify(562620)
	if s.rowIdx+1 >= s.rowCount {
		__antithesis_instrumentation__.Notify(562624)

		if next, err := s.source.BatchedNext(params); !next {
			__antithesis_instrumentation__.Notify(562626)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(562627)
		}
		__antithesis_instrumentation__.Notify(562625)
		s.rowCount = s.source.BatchedCount()
		s.rowIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(562628)

		s.rowIdx++
	}
	__antithesis_instrumentation__.Notify(562621)
	return s.rowCount > 0, nil
}

func (s *serializeNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(562629)
	return s.source.BatchedValues(s.rowIdx)
}
func (s *serializeNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562630)
	s.source.Close(ctx)
}

func (s *serializeNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(562631)
	return s.rowCount, s.fastPath
}

func (s *serializeNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(562632)
	m, ok := s.source.(mutationPlanNode)
	if !ok {
		__antithesis_instrumentation__.Notify(562634)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(562635)
	}
	__antithesis_instrumentation__.Notify(562633)
	return m.rowsWritten()
}

func (s *serializeNode) requireSpool() { __antithesis_instrumentation__.Notify(562636) }

type rowCountNode struct {
	source   batchedPlanNode
	rowCount int
}

var _ mutationPlanNode = &rowCountNode{}

func (r *rowCountNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(562637)
	done := false
	if f, ok := r.source.(planNodeFastPath); ok {
		__antithesis_instrumentation__.Notify(562640)
		r.rowCount, done = f.FastPathResults()
	} else {
		__antithesis_instrumentation__.Notify(562641)
	}
	__antithesis_instrumentation__.Notify(562638)
	if !done {
		__antithesis_instrumentation__.Notify(562642)
		for {
			__antithesis_instrumentation__.Notify(562643)
			if next, err := r.source.BatchedNext(params); !next {
				__antithesis_instrumentation__.Notify(562645)
				return err
			} else {
				__antithesis_instrumentation__.Notify(562646)
			}
			__antithesis_instrumentation__.Notify(562644)
			r.rowCount += r.source.BatchedCount()
		}
	} else {
		__antithesis_instrumentation__.Notify(562647)
	}
	__antithesis_instrumentation__.Notify(562639)
	return nil
}

func (r *rowCountNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(562648)
	return false, nil
}
func (r *rowCountNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(562649)
	return nil
}
func (r *rowCountNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562650)
	r.source.Close(ctx)
}

func (r *rowCountNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(562651)
	return r.rowCount, true
}

func (r *rowCountNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(562652)
	m, ok := r.source.(mutationPlanNode)
	if !ok {
		__antithesis_instrumentation__.Notify(562654)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(562655)
	}
	__antithesis_instrumentation__.Notify(562653)
	return m.rowsWritten()
}
