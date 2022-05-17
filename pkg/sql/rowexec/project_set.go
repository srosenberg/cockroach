package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/errors"
)

type projectSetProcessor struct {
	execinfra.ProcessorBase

	input execinfra.RowSource
	spec  *execinfrapb.ProjectSetSpec

	exprHelpers []*execinfrapb.ExprHelper

	funcs []*tree.FuncExpr

	inputRowReady bool

	rowBuffer rowenc.EncDatumRow

	gens []tree.ValueGenerator

	done []bool

	cancelChecker cancelchecker.CancelChecker
}

var _ execinfra.Processor = &projectSetProcessor{}
var _ execinfra.RowSource = &projectSetProcessor{}
var _ execinfra.OpNode = &projectSetProcessor{}

const projectSetProcName = "projectSet"

func newProjectSetProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.ProjectSetSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*projectSetProcessor, error) {
	__antithesis_instrumentation__.Notify(574341)
	outputTypes := append(input.OutputTypes(), spec.GeneratedColumns...)
	ps := &projectSetProcessor{
		input:       input,
		spec:        spec,
		exprHelpers: make([]*execinfrapb.ExprHelper, len(spec.Exprs)),
		funcs:       make([]*tree.FuncExpr, len(spec.Exprs)),
		rowBuffer:   make(rowenc.EncDatumRow, len(outputTypes)),
		gens:        make([]tree.ValueGenerator, len(spec.Exprs)),
		done:        make([]bool, len(spec.Exprs)),
	}
	if err := ps.Init(
		ps,
		post,
		outputTypes,
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{ps.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(574344)
				ps.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574345)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574346)
	}
	__antithesis_instrumentation__.Notify(574342)

	semaCtx := ps.FlowCtx.NewSemaContext(ps.EvalCtx.Txn)
	for i, expr := range ps.spec.Exprs {
		__antithesis_instrumentation__.Notify(574347)
		var helper execinfrapb.ExprHelper
		err := helper.Init(expr, ps.input.OutputTypes(), semaCtx, ps.EvalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(574350)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(574351)
		}
		__antithesis_instrumentation__.Notify(574348)
		if tFunc, ok := helper.Expr.(*tree.FuncExpr); ok && func() bool {
			__antithesis_instrumentation__.Notify(574352)
			return tFunc.IsGeneratorApplication() == true
		}() == true {
			__antithesis_instrumentation__.Notify(574353)

			ps.funcs[i] = tFunc
		} else {
			__antithesis_instrumentation__.Notify(574354)
		}
		__antithesis_instrumentation__.Notify(574349)
		ps.exprHelpers[i] = &helper
	}
	__antithesis_instrumentation__.Notify(574343)
	return ps, nil
}

func (ps *projectSetProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574355)
	ctx = ps.StartInternal(ctx, projectSetProcName)
	ps.input.Start(ctx)
	ps.cancelChecker.Reset(ctx)
}

func (ps *projectSetProcessor) nextInputRow() (
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
	error,
) {
	__antithesis_instrumentation__.Notify(574356)
	row, meta := ps.input.Next()
	if row == nil {
		__antithesis_instrumentation__.Notify(574359)
		return nil, meta, nil
	} else {
		__antithesis_instrumentation__.Notify(574360)
	}
	__antithesis_instrumentation__.Notify(574357)

	for i := range ps.exprHelpers {
		__antithesis_instrumentation__.Notify(574361)
		if fn := ps.funcs[i]; fn != nil {
			__antithesis_instrumentation__.Notify(574363)

			ps.exprHelpers[i].Row = row

			ps.EvalCtx.IVarContainer = ps.exprHelpers[i]
			gen, err := fn.EvalArgsAndGetGenerator(ps.EvalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(574367)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(574368)
			}
			__antithesis_instrumentation__.Notify(574364)
			if gen == nil {
				__antithesis_instrumentation__.Notify(574369)
				gen = builtins.EmptyGenerator()
			} else {
				__antithesis_instrumentation__.Notify(574370)
			}
			__antithesis_instrumentation__.Notify(574365)
			if err := gen.Start(ps.Ctx, ps.FlowCtx.Txn); err != nil {
				__antithesis_instrumentation__.Notify(574371)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(574372)
			}
			__antithesis_instrumentation__.Notify(574366)
			ps.gens[i] = gen
		} else {
			__antithesis_instrumentation__.Notify(574373)
		}
		__antithesis_instrumentation__.Notify(574362)
		ps.done[i] = false
	}
	__antithesis_instrumentation__.Notify(574358)

	return row, nil, nil
}

func (ps *projectSetProcessor) nextGeneratorValues() (newValAvail bool, err error) {
	__antithesis_instrumentation__.Notify(574374)
	colIdx := len(ps.input.OutputTypes())
	for i := range ps.exprHelpers {
		__antithesis_instrumentation__.Notify(574376)

		if gen := ps.gens[i]; gen != nil {
			__antithesis_instrumentation__.Notify(574377)

			numCols := int(ps.spec.NumColsPerGen[i])
			if !ps.done[i] {
				__antithesis_instrumentation__.Notify(574378)

				hasVals, err := gen.Next(ps.Ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(574380)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(574381)
				}
				__antithesis_instrumentation__.Notify(574379)
				if hasVals {
					__antithesis_instrumentation__.Notify(574382)

					values, err := gen.Values()
					if err != nil {
						__antithesis_instrumentation__.Notify(574385)
						return false, err
					} else {
						__antithesis_instrumentation__.Notify(574386)
					}
					__antithesis_instrumentation__.Notify(574383)
					for _, value := range values {
						__antithesis_instrumentation__.Notify(574387)
						ps.rowBuffer[colIdx] = ps.toEncDatum(value, colIdx)
						colIdx++
					}
					__antithesis_instrumentation__.Notify(574384)
					newValAvail = true
				} else {
					__antithesis_instrumentation__.Notify(574388)
					ps.done[i] = true

					for j := 0; j < numCols; j++ {
						__antithesis_instrumentation__.Notify(574389)
						ps.rowBuffer[colIdx] = ps.toEncDatum(tree.DNull, colIdx)
						colIdx++
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(574390)

				colIdx += numCols
			}
		} else {
			__antithesis_instrumentation__.Notify(574391)

			if !ps.done[i] {
				__antithesis_instrumentation__.Notify(574392)

				value, err := ps.exprHelpers[i].Eval(ps.rowBuffer)
				if err != nil {
					__antithesis_instrumentation__.Notify(574394)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(574395)
				}
				__antithesis_instrumentation__.Notify(574393)
				ps.rowBuffer[colIdx] = ps.toEncDatum(value, colIdx)
				colIdx++
				newValAvail = true
				ps.done[i] = true
			} else {
				__antithesis_instrumentation__.Notify(574396)

				ps.rowBuffer[colIdx] = ps.toEncDatum(tree.DNull, colIdx)
				colIdx++
			}
		}
	}
	__antithesis_instrumentation__.Notify(574375)
	return newValAvail, nil
}

func (ps *projectSetProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(574397)
	for ps.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(574399)
		if err := ps.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(574403)
			ps.MoveToDraining(err)
			return nil, ps.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(574404)
		}
		__antithesis_instrumentation__.Notify(574400)

		if !ps.inputRowReady {
			__antithesis_instrumentation__.Notify(574405)

			row, meta, err := ps.nextInputRow()
			if meta != nil {
				__antithesis_instrumentation__.Notify(574409)
				if meta.Err != nil {
					__antithesis_instrumentation__.Notify(574411)
					ps.MoveToDraining(nil)
				} else {
					__antithesis_instrumentation__.Notify(574412)
				}
				__antithesis_instrumentation__.Notify(574410)
				return nil, meta
			} else {
				__antithesis_instrumentation__.Notify(574413)
			}
			__antithesis_instrumentation__.Notify(574406)
			if err != nil {
				__antithesis_instrumentation__.Notify(574414)
				ps.MoveToDraining(err)
				return nil, ps.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(574415)
			}
			__antithesis_instrumentation__.Notify(574407)
			if row == nil {
				__antithesis_instrumentation__.Notify(574416)
				ps.MoveToDraining(nil)
				return nil, ps.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(574417)
			}
			__antithesis_instrumentation__.Notify(574408)

			copy(ps.rowBuffer, row)
			ps.inputRowReady = true
		} else {
			__antithesis_instrumentation__.Notify(574418)
		}
		__antithesis_instrumentation__.Notify(574401)

		newValAvail, err := ps.nextGeneratorValues()
		if err != nil {
			__antithesis_instrumentation__.Notify(574419)
			ps.MoveToDraining(err)
			return nil, ps.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(574420)
		}
		__antithesis_instrumentation__.Notify(574402)
		if newValAvail {
			__antithesis_instrumentation__.Notify(574421)
			if outRow := ps.ProcessRowHelper(ps.rowBuffer); outRow != nil {
				__antithesis_instrumentation__.Notify(574422)
				return outRow, nil
			} else {
				__antithesis_instrumentation__.Notify(574423)
			}
		} else {
			__antithesis_instrumentation__.Notify(574424)

			ps.inputRowReady = false
		}
	}
	__antithesis_instrumentation__.Notify(574398)
	return nil, ps.DrainHelper()
}

func (ps *projectSetProcessor) toEncDatum(d tree.Datum, colIdx int) rowenc.EncDatum {
	__antithesis_instrumentation__.Notify(574425)
	generatedColIdx := colIdx - len(ps.input.OutputTypes())
	ctyp := ps.spec.GeneratedColumns[generatedColIdx]
	return rowenc.DatumToEncDatum(ctyp, d)
}

func (ps *projectSetProcessor) close() {
	__antithesis_instrumentation__.Notify(574426)
	ps.InternalCloseEx(func() {
		__antithesis_instrumentation__.Notify(574427)
		for _, gen := range ps.gens {
			__antithesis_instrumentation__.Notify(574428)
			if gen != nil {
				__antithesis_instrumentation__.Notify(574429)
				gen.Close(ps.Ctx)
			} else {
				__antithesis_instrumentation__.Notify(574430)
			}
		}
	})
}

func (ps *projectSetProcessor) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(574431)
	ps.close()
}

func (ps *projectSetProcessor) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(574432)
	if _, ok := ps.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(574434)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(574435)
	}
	__antithesis_instrumentation__.Notify(574433)
	return 0
}

func (ps *projectSetProcessor) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(574436)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(574438)
		if n, ok := ps.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(574440)
			return n
		} else {
			__antithesis_instrumentation__.Notify(574441)
		}
		__antithesis_instrumentation__.Notify(574439)
		panic("input to projectSetProcessor is not an execinfra.OpNode")

	} else {
		__antithesis_instrumentation__.Notify(574442)
	}
	__antithesis_instrumentation__.Notify(574437)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
