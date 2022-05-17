package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type rowSourceToPlanNode struct {
	source    execinfra.RowSource
	forwarder metadataForwarder

	originalPlanNode planNode

	planCols colinfo.ResultColumns

	row      rowenc.EncDatumRow
	da       tree.DatumAlloc
	datumRow tree.Datums
}

var _ planNode = &rowSourceToPlanNode{}

func makeRowSourceToPlanNode(
	s execinfra.RowSource,
	forwarder metadataForwarder,
	planCols colinfo.ResultColumns,
	originalPlanNode planNode,
) *rowSourceToPlanNode {
	__antithesis_instrumentation__.Notify(568906)
	row := make(tree.Datums, len(planCols))

	return &rowSourceToPlanNode{
		source:           s,
		datumRow:         row,
		forwarder:        forwarder,
		planCols:         planCols,
		originalPlanNode: originalPlanNode,
	}
}

func (r *rowSourceToPlanNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(568907)
	r.source.Start(params.ctx)
	return nil
}

func (r *rowSourceToPlanNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(568908)
	for {
		__antithesis_instrumentation__.Notify(568909)
		var p *execinfrapb.ProducerMetadata
		r.row, p = r.source.Next()

		if p != nil {
			__antithesis_instrumentation__.Notify(568913)
			if p.Err != nil {
				__antithesis_instrumentation__.Notify(568917)
				return false, p.Err
			} else {
				__antithesis_instrumentation__.Notify(568918)
			}
			__antithesis_instrumentation__.Notify(568914)
			if r.forwarder != nil {
				__antithesis_instrumentation__.Notify(568919)
				r.forwarder.forwardMetadata(p)
				continue
			} else {
				__antithesis_instrumentation__.Notify(568920)
			}
			__antithesis_instrumentation__.Notify(568915)
			if p.TraceData != nil {
				__antithesis_instrumentation__.Notify(568921)

				continue
			} else {
				__antithesis_instrumentation__.Notify(568922)
			}
			__antithesis_instrumentation__.Notify(568916)
			return false, fmt.Errorf("unexpected producer metadata: %+v", p)
		} else {
			__antithesis_instrumentation__.Notify(568923)
		}
		__antithesis_instrumentation__.Notify(568910)

		if r.row == nil {
			__antithesis_instrumentation__.Notify(568924)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(568925)
		}
		__antithesis_instrumentation__.Notify(568911)

		types := r.source.OutputTypes()
		for i := range r.planCols {
			__antithesis_instrumentation__.Notify(568926)
			encDatum := r.row[i]
			err := encDatum.EnsureDecoded(types[i], &r.da)
			if err != nil {
				__antithesis_instrumentation__.Notify(568928)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(568929)
			}
			__antithesis_instrumentation__.Notify(568927)
			r.datumRow[i] = encDatum.Datum
		}
		__antithesis_instrumentation__.Notify(568912)

		return true, nil
	}
}

func (r *rowSourceToPlanNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(568930)
	return r.datumRow
}

func (r *rowSourceToPlanNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(568931)
	if r.source != nil {
		__antithesis_instrumentation__.Notify(568933)
		r.source.ConsumerClosed()
		r.source = nil
	} else {
		__antithesis_instrumentation__.Notify(568934)
	}
	__antithesis_instrumentation__.Notify(568932)
	if r.originalPlanNode != nil {
		__antithesis_instrumentation__.Notify(568935)
		r.originalPlanNode.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(568936)
	}
}
