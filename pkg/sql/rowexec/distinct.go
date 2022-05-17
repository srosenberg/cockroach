package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/stringarena"
	"github.com/cockroachdb/errors"
)

type distinct struct {
	execinfra.ProcessorBase

	input            execinfra.RowSource
	types            []*types.T
	haveLastGroupKey bool
	lastGroupKey     rowenc.EncDatumRow
	arena            stringarena.Arena
	seen             map[string]struct{}
	distinctCols     struct {
		ordered    []uint32
		nonOrdered []uint32
	}
	memAcc           mon.BoundAccount
	datumAlloc       tree.DatumAlloc
	scratch          []byte
	nullsAreDistinct bool
	nullCount        uint32
	errorOnDup       string
}

type sortedDistinct struct {
	*distinct
}

var _ execinfra.Processor = &distinct{}
var _ execinfra.RowSource = &distinct{}
var _ execinfra.OpNode = &distinct{}

const distinctProcName = "distinct"

var _ execinfra.Processor = &sortedDistinct{}
var _ execinfra.RowSource = &sortedDistinct{}
var _ execinfra.OpNode = &sortedDistinct{}

const sortedDistinctProcName = "sorted distinct"

func newDistinct(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.DistinctSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.RowSourcedProcessor, error) {
	__antithesis_instrumentation__.Notify(572114)
	if len(spec.DistinctColumns) == 0 {
		__antithesis_instrumentation__.Notify(572120)
		return nil, errors.AssertionFailedf("0 distinct columns specified for distinct processor")
	} else {
		__antithesis_instrumentation__.Notify(572121)
	}
	__antithesis_instrumentation__.Notify(572115)

	nonOrderedCols := make([]uint32, 0, len(spec.DistinctColumns)-len(spec.OrderedColumns))
	for _, col := range spec.DistinctColumns {
		__antithesis_instrumentation__.Notify(572122)
		ordered := false
		for _, ordCol := range spec.OrderedColumns {
			__antithesis_instrumentation__.Notify(572124)
			if col == ordCol {
				__antithesis_instrumentation__.Notify(572125)
				ordered = true
				break
			} else {
				__antithesis_instrumentation__.Notify(572126)
			}
		}
		__antithesis_instrumentation__.Notify(572123)
		if !ordered {
			__antithesis_instrumentation__.Notify(572127)
			nonOrderedCols = append(nonOrderedCols, col)
		} else {
			__antithesis_instrumentation__.Notify(572128)
		}
	}
	__antithesis_instrumentation__.Notify(572116)

	ctx := flowCtx.EvalCtx.Ctx()
	memMonitor := execinfra.NewMonitor(ctx, flowCtx.EvalCtx.Mon, "distinct-mem")
	d := &distinct{
		input:            input,
		memAcc:           memMonitor.MakeBoundAccount(),
		types:            input.OutputTypes(),
		nullsAreDistinct: spec.NullsAreDistinct,
		errorOnDup:       spec.ErrorOnDup,
	}
	d.distinctCols.ordered = spec.OrderedColumns
	d.distinctCols.nonOrdered = nonOrderedCols

	var returnProcessor execinfra.RowSourcedProcessor = d
	if len(nonOrderedCols) == 0 {
		__antithesis_instrumentation__.Notify(572129)

		sd := &sortedDistinct{distinct: d}
		returnProcessor = sd
	} else {
		__antithesis_instrumentation__.Notify(572130)
	}
	__antithesis_instrumentation__.Notify(572117)

	if err := d.Init(
		d, post, d.types, flowCtx, processorID, output, memMonitor,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{d.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(572131)
				d.close()
				return nil
			},
		}); err != nil {
		__antithesis_instrumentation__.Notify(572132)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572133)
	}
	__antithesis_instrumentation__.Notify(572118)
	d.lastGroupKey = d.OutputHelper.RowAlloc.AllocRow(len(d.types))
	d.haveLastGroupKey = false

	d.arena = stringarena.Make(&d.memAcc)

	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		__antithesis_instrumentation__.Notify(572134)
		d.input = newInputStatCollector(d.input)
		d.ExecStatsForTrace = d.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(572135)
	}
	__antithesis_instrumentation__.Notify(572119)

	return returnProcessor, nil
}

func (d *distinct) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572136)
	ctx = d.StartInternal(ctx, distinctProcName)
	d.input.Start(ctx)
}

func (d *sortedDistinct) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572137)
	ctx = d.StartInternal(ctx, sortedDistinctProcName)
	d.input.Start(ctx)
}

func (d *distinct) matchLastGroupKey(row rowenc.EncDatumRow) (bool, error) {
	__antithesis_instrumentation__.Notify(572138)
	if !d.haveLastGroupKey {
		__antithesis_instrumentation__.Notify(572141)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(572142)
	}
	__antithesis_instrumentation__.Notify(572139)
	for _, colIdx := range d.distinctCols.ordered {
		__antithesis_instrumentation__.Notify(572143)
		res, err := d.lastGroupKey[colIdx].Compare(
			d.types[colIdx], &d.datumAlloc, d.EvalCtx, &row[colIdx],
		)
		if res != 0 || func() bool {
			__antithesis_instrumentation__.Notify(572145)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(572146)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(572147)
		}
		__antithesis_instrumentation__.Notify(572144)

		if d.nullsAreDistinct && func() bool {
			__antithesis_instrumentation__.Notify(572148)
			return d.lastGroupKey[colIdx].IsNull() == true
		}() == true {
			__antithesis_instrumentation__.Notify(572149)
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(572150)
		}
	}
	__antithesis_instrumentation__.Notify(572140)
	return true, nil
}

func (d *distinct) encode(appendTo []byte, row rowenc.EncDatumRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(572151)
	var err error
	foundNull := false
	for _, colIdx := range d.distinctCols.nonOrdered {
		__antithesis_instrumentation__.Notify(572154)
		datum := row[colIdx]

		appendTo, err = datum.Fingerprint(d.Ctx, d.types[colIdx], &d.datumAlloc, appendTo, &d.memAcc)
		if err != nil {
			__antithesis_instrumentation__.Notify(572156)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(572157)
		}
		__antithesis_instrumentation__.Notify(572155)

		if d.nullsAreDistinct && func() bool {
			__antithesis_instrumentation__.Notify(572158)
			return datum.IsNull() == true
		}() == true {
			__antithesis_instrumentation__.Notify(572159)
			foundNull = true
		} else {
			__antithesis_instrumentation__.Notify(572160)
		}
	}
	__antithesis_instrumentation__.Notify(572152)

	if foundNull {
		__antithesis_instrumentation__.Notify(572161)
		appendTo = encoding.EncodeUint32Ascending(appendTo, d.nullCount)
		d.nullCount++
	} else {
		__antithesis_instrumentation__.Notify(572162)
	}
	__antithesis_instrumentation__.Notify(572153)

	return appendTo, nil
}

func (d *distinct) close() {
	__antithesis_instrumentation__.Notify(572163)
	if d.InternalClose() {
		__antithesis_instrumentation__.Notify(572164)
		d.memAcc.Close(d.Ctx)
		d.MemMonitor.Stop(d.Ctx)
	} else {
		__antithesis_instrumentation__.Notify(572165)
	}
}

func (d *distinct) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572166)
	for d.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572168)
		row, meta := d.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(572176)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(572178)
				d.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(572179)
			}
			__antithesis_instrumentation__.Notify(572177)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572180)
		}
		__antithesis_instrumentation__.Notify(572169)
		if row == nil {
			__antithesis_instrumentation__.Notify(572181)
			d.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(572182)
		}
		__antithesis_instrumentation__.Notify(572170)

		encoding, err := d.encode(d.scratch, row)
		if err != nil {
			__antithesis_instrumentation__.Notify(572183)
			d.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(572184)
		}
		__antithesis_instrumentation__.Notify(572171)
		d.scratch = encoding[:0]

		matched, err := d.matchLastGroupKey(row)
		if err != nil {
			__antithesis_instrumentation__.Notify(572185)
			d.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(572186)
		}
		__antithesis_instrumentation__.Notify(572172)

		if !matched {
			__antithesis_instrumentation__.Notify(572187)

			copy(d.lastGroupKey, row)
			d.haveLastGroupKey = true
			if err := d.arena.UnsafeReset(d.Ctx); err != nil {
				__antithesis_instrumentation__.Notify(572189)
				d.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(572190)
			}
			__antithesis_instrumentation__.Notify(572188)
			d.seen = make(map[string]struct{})
		} else {
			__antithesis_instrumentation__.Notify(572191)
		}
		__antithesis_instrumentation__.Notify(572173)

		if _, ok := d.seen[string(encoding)]; ok {
			__antithesis_instrumentation__.Notify(572192)
			if d.errorOnDup != "" {
				__antithesis_instrumentation__.Notify(572194)

				err = pgerror.Newf(pgcode.CardinalityViolation, "%s", d.errorOnDup)
				d.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(572195)
			}
			__antithesis_instrumentation__.Notify(572193)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572196)
		}
		__antithesis_instrumentation__.Notify(572174)
		s, err := d.arena.AllocBytes(d.Ctx, encoding)
		if err != nil {
			__antithesis_instrumentation__.Notify(572197)
			d.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(572198)
		}
		__antithesis_instrumentation__.Notify(572175)
		d.seen[s] = struct{}{}

		if outRow := d.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(572199)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(572200)
		}
	}
	__antithesis_instrumentation__.Notify(572167)
	return nil, d.DrainHelper()
}

func (d *sortedDistinct) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572201)
	for d.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572203)
		row, meta := d.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(572208)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(572210)
				d.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(572211)
			}
			__antithesis_instrumentation__.Notify(572209)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572212)
		}
		__antithesis_instrumentation__.Notify(572204)
		if row == nil {
			__antithesis_instrumentation__.Notify(572213)
			d.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(572214)
		}
		__antithesis_instrumentation__.Notify(572205)
		matched, err := d.matchLastGroupKey(row)
		if err != nil {
			__antithesis_instrumentation__.Notify(572215)
			d.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(572216)
		}
		__antithesis_instrumentation__.Notify(572206)
		if matched {
			__antithesis_instrumentation__.Notify(572217)
			if d.errorOnDup != "" {
				__antithesis_instrumentation__.Notify(572219)

				err = pgerror.Newf(pgcode.CardinalityViolation, "%s", d.errorOnDup)
				d.MoveToDraining(err)
				break
			} else {
				__antithesis_instrumentation__.Notify(572220)
			}
			__antithesis_instrumentation__.Notify(572218)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572221)
		}
		__antithesis_instrumentation__.Notify(572207)

		d.haveLastGroupKey = true
		copy(d.lastGroupKey, row)

		if outRow := d.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(572222)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(572223)
		}
	}
	__antithesis_instrumentation__.Notify(572202)
	return nil, d.DrainHelper()
}

func (d *distinct) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(572224)

	d.close()
}

func (d *distinct) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(572225)
	is, ok := getInputStats(d.input)
	if !ok {
		__antithesis_instrumentation__.Notify(572227)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572228)
	}
	__antithesis_instrumentation__.Notify(572226)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem: optional.MakeUint(uint64(d.MemMonitor.MaximumBytes())),
		},
		Output: d.OutputHelper.Stats(),
	}
}

func (d *distinct) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(572229)
	if _, ok := d.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(572231)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(572232)
	}
	__antithesis_instrumentation__.Notify(572230)
	return 0
}

func (d *distinct) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(572233)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(572235)
		if n, ok := d.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(572237)
			return n
		} else {
			__antithesis_instrumentation__.Notify(572238)
		}
		__antithesis_instrumentation__.Notify(572236)
		panic("input to distinct is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(572239)
	}
	__antithesis_instrumentation__.Notify(572234)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
