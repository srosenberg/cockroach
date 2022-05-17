package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type metadataForwarder interface {
	forwardMetadata(metadata *execinfrapb.ProducerMetadata)
}

type planNodeToRowSource struct {
	execinfra.ProcessorBase

	input execinfra.RowSource

	fastPath bool

	node        planNode
	params      runParams
	outputTypes []*types.T

	firstNotWrapped planNode

	row rowenc.EncDatumRow
}

var _ execinfra.OpNode = &planNodeToRowSource{}

func makePlanNodeToRowSource(
	source planNode, params runParams, fastPath bool,
) (*planNodeToRowSource, error) {
	__antithesis_instrumentation__.Notify(562725)
	var typs []*types.T
	if fastPath {
		__antithesis_instrumentation__.Notify(562727)

		typs = []*types.T{types.Int}
	} else {
		__antithesis_instrumentation__.Notify(562728)
		typs = getTypesFromResultColumns(planColumns(source))
	}
	__antithesis_instrumentation__.Notify(562726)
	row := make(rowenc.EncDatumRow, len(typs))

	return &planNodeToRowSource{
		node:        source,
		params:      params,
		outputTypes: typs,
		row:         row,
		fastPath:    fastPath,
	}, nil
}

var _ execinfra.LocalProcessor = &planNodeToRowSource{}

func (p *planNodeToRowSource) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(562729)

	_, isHookFnNode := p.node.(*hookFnNode)
	return isHookFnNode
}

func (p *planNodeToRowSource) InitWithOutput(
	flowCtx *execinfra.FlowCtx, post *execinfrapb.PostProcessSpec, output execinfra.RowReceiver,
) error {
	__antithesis_instrumentation__.Notify(562730)
	return p.InitWithEvalCtx(
		p,
		post,
		p.outputTypes,
		flowCtx,

		p.params.EvalContext(),
		0,
		output,
		nil,
		execinfra.ProcStateOpts{
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(562731)
				var meta []execinfrapb.ProducerMetadata
				if p.InternalClose() {
					__antithesis_instrumentation__.Notify(562733)

					if m, ok := p.node.(mutationPlanNode); ok {
						__antithesis_instrumentation__.Notify(562734)
						metrics := execinfrapb.GetMetricsMeta()
						metrics.RowsWritten = m.rowsWritten()
						meta = []execinfrapb.ProducerMetadata{{Metrics: metrics}}
					} else {
						__antithesis_instrumentation__.Notify(562735)
					}
				} else {
					__antithesis_instrumentation__.Notify(562736)
				}
				__antithesis_instrumentation__.Notify(562732)
				return meta
			},
		},
	)
}

func (p *planNodeToRowSource) SetInput(ctx context.Context, input execinfra.RowSource) error {
	__antithesis_instrumentation__.Notify(562737)
	if p.firstNotWrapped == nil {
		__antithesis_instrumentation__.Notify(562739)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(562740)
	}
	__antithesis_instrumentation__.Notify(562738)
	p.input = input
	p.AddInputToDrain(input)

	return walkPlan(ctx, p.node, planObserver{
		replaceNode: func(ctx context.Context, nodeName string, plan planNode) (planNode, error) {
			__antithesis_instrumentation__.Notify(562741)
			if plan == p.firstNotWrapped {
				__antithesis_instrumentation__.Notify(562743)
				return makeRowSourceToPlanNode(input, p, planColumns(p.firstNotWrapped), p.firstNotWrapped), nil
			} else {
				__antithesis_instrumentation__.Notify(562744)
			}
			__antithesis_instrumentation__.Notify(562742)
			return nil, nil
		},
	})
}

func (p *planNodeToRowSource) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(562745)
	ctx = p.StartInternalNoSpan(ctx)
	p.params.ctx = ctx

	if err := startExec(p.params, p.node); err != nil {
		__antithesis_instrumentation__.Notify(562746)
		p.MoveToDraining(err)
	} else {
		__antithesis_instrumentation__.Notify(562747)
	}
}

func (p *planNodeToRowSource) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(562748)
	if p.State == execinfra.StateRunning && func() bool {
		__antithesis_instrumentation__.Notify(562751)
		return p.fastPath == true
	}() == true {
		__antithesis_instrumentation__.Notify(562752)
		var count int

		fastPath, ok := p.node.(planNodeFastPath)

		if ok {
			__antithesis_instrumentation__.Notify(562755)
			var res bool
			if count, res = fastPath.FastPathResults(); res {
				__antithesis_instrumentation__.Notify(562756)
				if p.params.extendedEvalCtx.Tracing.Enabled() {
					__antithesis_instrumentation__.Notify(562757)
					log.VEvent(p.params.ctx, 2, "fast path completed")
				} else {
					__antithesis_instrumentation__.Notify(562758)
				}
			} else {
				__antithesis_instrumentation__.Notify(562759)

				count = 0
				ok = false
			}
		} else {
			__antithesis_instrumentation__.Notify(562760)
		}
		__antithesis_instrumentation__.Notify(562753)

		if !ok {
			__antithesis_instrumentation__.Notify(562761)

			next, err := p.node.Next(p.params)
			for ; next; next, err = p.node.Next(p.params) {
				__antithesis_instrumentation__.Notify(562763)
				count++
			}
			__antithesis_instrumentation__.Notify(562762)
			if err != nil {
				__antithesis_instrumentation__.Notify(562764)
				p.MoveToDraining(err)
				return nil, p.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(562765)
			}
		} else {
			__antithesis_instrumentation__.Notify(562766)
		}
		__antithesis_instrumentation__.Notify(562754)
		p.MoveToDraining(nil)

		return rowenc.EncDatumRow{rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(count))}}, nil
	} else {
		__antithesis_instrumentation__.Notify(562767)
	}
	__antithesis_instrumentation__.Notify(562749)

	for p.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(562768)
		valid, err := p.node.Next(p.params)
		if err != nil || func() bool {
			__antithesis_instrumentation__.Notify(562771)
			return !valid == true
		}() == true {
			__antithesis_instrumentation__.Notify(562772)
			p.MoveToDraining(err)
			return nil, p.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(562773)
		}
		__antithesis_instrumentation__.Notify(562769)

		for i, datum := range p.node.Values() {
			__antithesis_instrumentation__.Notify(562774)
			if datum != nil {
				__antithesis_instrumentation__.Notify(562775)
				p.row[i] = rowenc.DatumToEncDatum(p.outputTypes[i], datum)
			} else {
				__antithesis_instrumentation__.Notify(562776)
			}
		}
		__antithesis_instrumentation__.Notify(562770)

		if outRow := p.ProcessRowHelper(p.row); outRow != nil {
			__antithesis_instrumentation__.Notify(562777)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(562778)
		}
	}
	__antithesis_instrumentation__.Notify(562750)
	return nil, p.DrainHelper()
}

func (p *planNodeToRowSource) forwardMetadata(metadata *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(562779)
	p.ProcessorBase.AppendTrailingMeta(*metadata)
}

func (p *planNodeToRowSource) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(562780)
	if _, ok := p.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(562782)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(562783)
	}
	__antithesis_instrumentation__.Notify(562781)
	return 0
}

func (p *planNodeToRowSource) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(562784)
	switch nth {
	case 0:
		__antithesis_instrumentation__.Notify(562785)
		if n, ok := p.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(562788)
			return n
		} else {
			__antithesis_instrumentation__.Notify(562789)
		}
		__antithesis_instrumentation__.Notify(562786)
		panic("input to planNodeToRowSource is not an execinfra.OpNode")
	default:
		__antithesis_instrumentation__.Notify(562787)
		panic(errors.AssertionFailedf("invalid index %d", nth))
	}
}
