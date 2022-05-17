package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type valuesProcessor struct {
	execinfra.ProcessorBase

	typs []*types.T
	data [][]byte

	numRows uint64
	rowBuf  rowenc.EncDatumRow
}

var _ execinfra.Processor = &valuesProcessor{}
var _ execinfra.RowSource = &valuesProcessor{}
var _ execinfra.OpNode = &valuesProcessor{}

const valuesProcName = "values"

func newValuesProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.ValuesCoreSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*valuesProcessor, error) {
	__antithesis_instrumentation__.Notify(575213)
	if len(spec.Columns) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(575217)
		return uint64(len(spec.RawBytes)) != spec.NumRows == true
	}() == true {
		__antithesis_instrumentation__.Notify(575218)
		return nil, errors.AssertionFailedf(
			"malformed ValuesCoreSpec: len(RawBytes) = %d does not equal NumRows = %d",
			len(spec.RawBytes), spec.NumRows,
		)
	} else {
		__antithesis_instrumentation__.Notify(575219)
	}
	__antithesis_instrumentation__.Notify(575214)
	v := &valuesProcessor{
		typs:    make([]*types.T, len(spec.Columns)),
		data:    spec.RawBytes,
		numRows: spec.NumRows,
		rowBuf:  make(rowenc.EncDatumRow, len(spec.Columns)),
	}
	for i := range spec.Columns {
		__antithesis_instrumentation__.Notify(575220)
		v.typs[i] = spec.Columns[i].Type
	}
	__antithesis_instrumentation__.Notify(575215)
	if err := v.Init(
		v, post, v.typs, flowCtx, processorID, output, nil, execinfra.ProcStateOpts{},
	); err != nil {
		__antithesis_instrumentation__.Notify(575221)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(575222)
	}
	__antithesis_instrumentation__.Notify(575216)
	return v, nil
}

func (v *valuesProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575223)
	v.StartInternal(ctx, valuesProcName)
}

func (v *valuesProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575224)
	for v.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(575226)
		if v.numRows == 0 {
			__antithesis_instrumentation__.Notify(575229)
			v.MoveToDraining(nil)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575230)
		}
		__antithesis_instrumentation__.Notify(575227)

		if len(v.typs) != 0 {
			__antithesis_instrumentation__.Notify(575231)
			rowData := v.data[0]
			for i, typ := range v.typs {
				__antithesis_instrumentation__.Notify(575234)
				var err error
				v.rowBuf[i], rowData, err = rowenc.EncDatumFromBuffer(
					typ, descpb.DatumEncoding_VALUE, rowData,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(575235)
					v.MoveToDraining(err)
					return nil, v.DrainHelper()
				} else {
					__antithesis_instrumentation__.Notify(575236)
				}
			}
			__antithesis_instrumentation__.Notify(575232)
			if len(rowData) != 0 {
				__antithesis_instrumentation__.Notify(575237)
				panic(errors.AssertionFailedf(
					"malformed ValuesCoreSpec row: %x, numRows %d", rowData, v.numRows,
				))
			} else {
				__antithesis_instrumentation__.Notify(575238)
			}
			__antithesis_instrumentation__.Notify(575233)
			v.data = v.data[1:]
		} else {
			__antithesis_instrumentation__.Notify(575239)
		}
		__antithesis_instrumentation__.Notify(575228)
		v.numRows--

		if outRow := v.ProcessRowHelper(v.rowBuf); outRow != nil {
			__antithesis_instrumentation__.Notify(575240)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(575241)
		}
	}
	__antithesis_instrumentation__.Notify(575225)

	return nil, v.DrainHelper()
}

func (v *valuesProcessor) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(575242)
	return 0
}

func (v *valuesProcessor) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(575243)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
