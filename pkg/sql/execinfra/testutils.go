package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

const StaticSQLInstanceID = base.SQLInstanceID(3)

type RepeatableRowSource struct {
	nextRowIdx int
	rows       rowenc.EncDatumRows

	types []*types.T
}

var _ RowSource = &RepeatableRowSource{}

func NewRepeatableRowSource(types []*types.T, rows rowenc.EncDatumRows) *RepeatableRowSource {
	__antithesis_instrumentation__.Notify(471488)
	if types == nil {
		__antithesis_instrumentation__.Notify(471490)
		panic("types required")
	} else {
		__antithesis_instrumentation__.Notify(471491)
	}
	__antithesis_instrumentation__.Notify(471489)
	return &RepeatableRowSource{rows: rows, types: types}
}

func (r *RepeatableRowSource) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(471492)
	return r.types
}

func (r *RepeatableRowSource) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(471493)
}

func (r *RepeatableRowSource) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(471494)

	if r.nextRowIdx >= len(r.rows) {
		__antithesis_instrumentation__.Notify(471496)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(471497)
	}
	__antithesis_instrumentation__.Notify(471495)
	nextRow := r.rows[r.nextRowIdx]
	r.nextRowIdx++
	return nextRow, nil
}

func (r *RepeatableRowSource) Reset() {
	__antithesis_instrumentation__.Notify(471498)
	r.nextRowIdx = 0
}

func (r *RepeatableRowSource) ConsumerDone() { __antithesis_instrumentation__.Notify(471499) }

func (r *RepeatableRowSource) ConsumerClosed() { __antithesis_instrumentation__.Notify(471500) }

func NewTestMemMonitor(ctx context.Context, st *cluster.Settings) *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(471501)
	memMonitor := mon.NewMonitor(
		"test-mem",
		mon.MemoryResource,
		nil,
		nil,
		-1,
		math.MaxInt64,
		st,
	)
	memMonitor.Start(ctx, nil, mon.MakeStandaloneBudget(math.MaxInt64))
	return memMonitor
}

func NewTestDiskMonitor(ctx context.Context, st *cluster.Settings) *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(471502)
	diskMonitor := mon.NewMonitor(
		"test-disk",
		mon.DiskResource,
		nil,
		nil,
		-1,
		math.MaxInt64,
		st,
	)
	diskMonitor.Start(ctx, nil, mon.MakeStandaloneBudget(math.MaxInt64))
	return diskMonitor
}

func GenerateValuesSpec(
	colTypes []*types.T, rows rowenc.EncDatumRows,
) (execinfrapb.ValuesCoreSpec, error) {
	__antithesis_instrumentation__.Notify(471503)
	var spec execinfrapb.ValuesCoreSpec
	spec.Columns = make([]execinfrapb.DatumInfo, len(colTypes))
	for i := range spec.Columns {
		__antithesis_instrumentation__.Notify(471506)
		spec.Columns[i].Type = colTypes[i]
		spec.Columns[i].Encoding = descpb.DatumEncoding_VALUE
	}
	__antithesis_instrumentation__.Notify(471504)

	spec.NumRows = uint64(len(rows))
	if len(colTypes) != 0 {
		__antithesis_instrumentation__.Notify(471507)
		var a tree.DatumAlloc
		for i := 0; i < len(rows); i++ {
			__antithesis_instrumentation__.Notify(471508)
			var buf []byte
			for j, info := range spec.Columns {
				__antithesis_instrumentation__.Notify(471510)
				var err error
				buf, err = rows[i][j].Encode(colTypes[j], &a, info.Encoding, buf)
				if err != nil {
					__antithesis_instrumentation__.Notify(471511)
					return execinfrapb.ValuesCoreSpec{}, err
				} else {
					__antithesis_instrumentation__.Notify(471512)
				}
			}
			__antithesis_instrumentation__.Notify(471509)
			spec.RawBytes = append(spec.RawBytes, buf)
		}
	} else {
		__antithesis_instrumentation__.Notify(471513)
	}
	__antithesis_instrumentation__.Notify(471505)
	return spec, nil
}
