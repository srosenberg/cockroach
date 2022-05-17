package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type mergeJoiner struct {
	joinerBase

	cancelChecker cancelchecker.CancelChecker

	leftSource, rightSource execinfra.RowSource
	leftRows, rightRows     []rowenc.EncDatumRow
	leftIdx, rightIdx       int
	trackMatchedRight       bool
	emitUnmatchedRight      bool
	matchedRight            util.FastIntSet
	matchedRightCount       int

	streamMerger streamMerger
}

var _ execinfra.Processor = &mergeJoiner{}
var _ execinfra.RowSource = &mergeJoiner{}
var _ execinfra.OpNode = &mergeJoiner{}

const mergeJoinerProcName = "merge joiner"

func newMergeJoiner(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.MergeJoinerSpec,
	leftSource execinfra.RowSource,
	rightSource execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*mergeJoiner, error) {
	__antithesis_instrumentation__.Notify(573901)
	m := &mergeJoiner{
		leftSource:  leftSource,
		rightSource: rightSource,
		trackMatchedRight: shouldEmitUnmatchedRow(rightSide, spec.Type) || func() bool {
			__antithesis_instrumentation__.Notify(573905)
			return spec.Type == descpb.RightSemiJoin == true
		}() == true,
	}

	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(573906)
		m.leftSource = newInputStatCollector(m.leftSource)
		m.rightSource = newInputStatCollector(m.rightSource)
		m.ExecStatsForTrace = m.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(573907)
	}
	__antithesis_instrumentation__.Notify(573902)

	if err := m.joinerBase.init(
		m, flowCtx, processorID, leftSource.OutputTypes(), rightSource.OutputTypes(),
		spec.Type, spec.OnExpr, false,
		post, output,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{leftSource, rightSource},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(573908)
				m.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(573909)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(573910)
	}
	__antithesis_instrumentation__.Notify(573903)

	m.MemMonitor = execinfra.NewMonitor(flowCtx.EvalCtx.Ctx(), flowCtx.EvalCtx.Mon, "mergejoiner-mem")

	var err error
	m.streamMerger, err = makeStreamMerger(
		m.leftSource,
		execinfrapb.ConvertToColumnOrdering(spec.LeftOrdering),
		m.rightSource,
		execinfrapb.ConvertToColumnOrdering(spec.RightOrdering),
		spec.NullEquality,
		m.MemMonitor,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(573911)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(573912)
	}
	__antithesis_instrumentation__.Notify(573904)

	return m, nil
}

func (m *mergeJoiner) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573913)
	ctx = m.StartInternal(ctx, mergeJoinerProcName)
	m.streamMerger.start(ctx)
	m.cancelChecker.Reset(ctx)
}

func (m *mergeJoiner) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(573914)
	for m.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(573916)
		row, meta := m.nextRow()
		if meta != nil {
			__antithesis_instrumentation__.Notify(573919)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(573921)
				m.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(573922)
			}
			__antithesis_instrumentation__.Notify(573920)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(573923)
		}
		__antithesis_instrumentation__.Notify(573917)
		if row == nil {
			__antithesis_instrumentation__.Notify(573924)
			m.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(573925)
		}
		__antithesis_instrumentation__.Notify(573918)

		if outRow := m.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(573926)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(573927)
		}
	}
	__antithesis_instrumentation__.Notify(573915)
	return nil, m.DrainHelper()
}

func (m *mergeJoiner) nextRow() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(573928)

	for {
		__antithesis_instrumentation__.Notify(573929)
		for m.leftIdx < len(m.leftRows) {
			__antithesis_instrumentation__.Notify(573934)

			lrow := m.leftRows[m.leftIdx]
			for m.rightIdx < len(m.rightRows) {
				__antithesis_instrumentation__.Notify(573939)

				ridx := m.rightIdx
				m.rightIdx++
				if (m.joinType == descpb.RightSemiJoin || func() bool {
					__antithesis_instrumentation__.Notify(573942)
					return m.joinType == descpb.RightAntiJoin == true
				}() == true) && func() bool {
					__antithesis_instrumentation__.Notify(573943)
					return m.matchedRight.Contains(ridx) == true
				}() == true {
					__antithesis_instrumentation__.Notify(573944)

					continue
				} else {
					__antithesis_instrumentation__.Notify(573945)
				}
				__antithesis_instrumentation__.Notify(573940)
				renderedRow, err := m.render(lrow, m.rightRows[ridx])
				if err != nil {
					__antithesis_instrumentation__.Notify(573946)
					return nil, &execinfrapb.ProducerMetadata{Err: err}
				} else {
					__antithesis_instrumentation__.Notify(573947)
				}
				__antithesis_instrumentation__.Notify(573941)
				if renderedRow != nil {
					__antithesis_instrumentation__.Notify(573948)
					m.matchedRightCount++
					if m.trackMatchedRight {
						__antithesis_instrumentation__.Notify(573953)
						m.matchedRight.Add(ridx)
					} else {
						__antithesis_instrumentation__.Notify(573954)
					}
					__antithesis_instrumentation__.Notify(573949)
					if m.joinType == descpb.LeftAntiJoin || func() bool {
						__antithesis_instrumentation__.Notify(573955)
						return m.joinType == descpb.ExceptAllJoin == true
					}() == true {
						__antithesis_instrumentation__.Notify(573956)

						break
					} else {
						__antithesis_instrumentation__.Notify(573957)
					}
					__antithesis_instrumentation__.Notify(573950)
					if m.joinType == descpb.RightAntiJoin {
						__antithesis_instrumentation__.Notify(573958)

						continue
					} else {
						__antithesis_instrumentation__.Notify(573959)
					}
					__antithesis_instrumentation__.Notify(573951)
					if m.joinType == descpb.LeftSemiJoin || func() bool {
						__antithesis_instrumentation__.Notify(573960)
						return m.joinType == descpb.IntersectAllJoin == true
					}() == true {
						__antithesis_instrumentation__.Notify(573961)

						m.rightIdx = len(m.rightRows)
					} else {
						__antithesis_instrumentation__.Notify(573962)
					}
					__antithesis_instrumentation__.Notify(573952)
					return renderedRow, nil
				} else {
					__antithesis_instrumentation__.Notify(573963)
				}
			}
			__antithesis_instrumentation__.Notify(573935)

			if err := m.cancelChecker.Check(); err != nil {
				__antithesis_instrumentation__.Notify(573964)
				return nil, &execinfrapb.ProducerMetadata{Err: err}
			} else {
				__antithesis_instrumentation__.Notify(573965)
			}
			__antithesis_instrumentation__.Notify(573936)

			m.leftIdx++
			m.rightIdx = 0

			if m.joinType.IsSetOpJoin() {
				__antithesis_instrumentation__.Notify(573966)
				m.rightIdx = m.leftIdx
			} else {
				__antithesis_instrumentation__.Notify(573967)
			}
			__antithesis_instrumentation__.Notify(573937)

			if m.matchedRightCount == 0 && func() bool {
				__antithesis_instrumentation__.Notify(573968)
				return shouldEmitUnmatchedRow(leftSide, m.joinType) == true
			}() == true {
				__antithesis_instrumentation__.Notify(573969)
				return m.renderUnmatchedRow(lrow, leftSide), nil
			} else {
				__antithesis_instrumentation__.Notify(573970)
			}
			__antithesis_instrumentation__.Notify(573938)

			m.matchedRightCount = 0
		}
		__antithesis_instrumentation__.Notify(573930)

		if m.emitUnmatchedRight {
			__antithesis_instrumentation__.Notify(573971)
			for m.rightIdx < len(m.rightRows) {
				__antithesis_instrumentation__.Notify(573973)
				ridx := m.rightIdx
				m.rightIdx++
				if m.matchedRight.Contains(ridx) {
					__antithesis_instrumentation__.Notify(573975)
					continue
				} else {
					__antithesis_instrumentation__.Notify(573976)
				}
				__antithesis_instrumentation__.Notify(573974)
				return m.renderUnmatchedRow(m.rightRows[ridx], rightSide), nil
			}
			__antithesis_instrumentation__.Notify(573972)
			m.emitUnmatchedRight = false
		} else {
			__antithesis_instrumentation__.Notify(573977)
		}
		__antithesis_instrumentation__.Notify(573931)

		var meta *execinfrapb.ProducerMetadata

		m.leftRows, m.rightRows, meta = m.streamMerger.NextBatch(m.Ctx, m.EvalCtx)
		if meta != nil {
			__antithesis_instrumentation__.Notify(573978)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(573979)
		}
		__antithesis_instrumentation__.Notify(573932)
		if m.leftRows == nil && func() bool {
			__antithesis_instrumentation__.Notify(573980)
			return m.rightRows == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(573981)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(573982)
		}
		__antithesis_instrumentation__.Notify(573933)

		m.emitUnmatchedRight = shouldEmitUnmatchedRow(rightSide, m.joinType)
		m.leftIdx, m.rightIdx = 0, 0
		if m.trackMatchedRight {
			__antithesis_instrumentation__.Notify(573983)
			m.matchedRight = util.FastIntSet{}
		} else {
			__antithesis_instrumentation__.Notify(573984)
		}
	}
}

func (m *mergeJoiner) close() {
	__antithesis_instrumentation__.Notify(573985)
	if m.InternalClose() {
		__antithesis_instrumentation__.Notify(573986)
		m.streamMerger.close(m.Ctx)
		m.MemMonitor.Stop(m.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(573987)
	}
}

func (m *mergeJoiner) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(573988)

	m.close()
}

func (m *mergeJoiner) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(573989)
	lis, ok := getInputStats(m.leftSource)
	if !ok {
		__antithesis_instrumentation__.Notify(573992)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573993)
	}
	__antithesis_instrumentation__.Notify(573990)
	ris, ok := getInputStats(m.rightSource)
	if !ok {
		__antithesis_instrumentation__.Notify(573994)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(573995)
	}
	__antithesis_instrumentation__.Notify(573991)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{lis, ris},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem: optional.MakeUint(uint64(m.MemMonitor.MaximumBytes())),
		},
		Output: m.OutputHelper.Stats(),
	}
}

func (m *mergeJoiner) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(573996)
	if _, ok := m.leftSource.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(573998)
		if _, ok := m.rightSource.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(573999)
			return 2
		} else {
			__antithesis_instrumentation__.Notify(574000)
		}
	} else {
		__antithesis_instrumentation__.Notify(574001)
	}
	__antithesis_instrumentation__.Notify(573997)
	return 0
}

func (m *mergeJoiner) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(574002)
	switch nth {
	case 0:
		__antithesis_instrumentation__.Notify(574003)
		if n, ok := m.leftSource.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(574008)
			return n
		} else {
			__antithesis_instrumentation__.Notify(574009)
		}
		__antithesis_instrumentation__.Notify(574004)
		panic("left input to mergeJoiner is not an execinfra.OpNode")
	case 1:
		__antithesis_instrumentation__.Notify(574005)
		if n, ok := m.rightSource.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(574010)
			return n
		} else {
			__antithesis_instrumentation__.Notify(574011)
		}
		__antithesis_instrumentation__.Notify(574006)
		panic("right input to mergeJoiner is not an execinfra.OpNode")
	default:
		__antithesis_instrumentation__.Notify(574007)
		panic(errors.AssertionFailedf("invalid index %d", nth))
	}
}
