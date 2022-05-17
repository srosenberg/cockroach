package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

var CTASPlanResultTypes = []*types.T{
	types.Bytes,
}

type bulkRowWriter struct {
	execinfra.ProcessorBase
	flowCtx        *execinfra.FlowCtx
	processorID    int32
	batchIdxAtomic int64
	tableDesc      catalog.TableDescriptor
	spec           execinfrapb.BulkRowWriterSpec
	input          execinfra.RowSource
	output         execinfra.RowReceiver
	summary        roachpb.BulkOpSummary
}

var _ execinfra.Processor = &bulkRowWriter{}
var _ execinfra.RowSource = &bulkRowWriter{}

func newBulkRowWriterProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.BulkRowWriterSpec,
	input execinfra.RowSource,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(571979)
	c := &bulkRowWriter{
		flowCtx:        flowCtx,
		processorID:    processorID,
		batchIdxAtomic: 0,
		tableDesc:      flowCtx.TableDescriptor(&spec.Table),
		spec:           spec,
		input:          input,
		output:         output,
	}
	if err := c.Init(
		c, &execinfrapb.PostProcessSpec{}, CTASPlanResultTypes, flowCtx, processorID, output,
		nil, execinfra.ProcStateOpts{InputsToDrain: []execinfra.RowSource{input}},
	); err != nil {
		__antithesis_instrumentation__.Notify(571981)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571982)
	}
	__antithesis_instrumentation__.Notify(571980)
	return c, nil
}

func (sp *bulkRowWriter) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(571983)
	ctx = sp.StartInternal(ctx, "bulkRowWriter")
	sp.input.Start(ctx)
	err := sp.work(ctx)
	sp.MoveToDraining(err)
}

func (sp *bulkRowWriter) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(571984)

	if sp.ProcessorBase.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(571986)
		countsBytes, marshalErr := protoutil.Marshal(&sp.summary)
		sp.MoveToDraining(marshalErr)
		if marshalErr == nil {
			__antithesis_instrumentation__.Notify(571987)

			return rowenc.EncDatumRow{
				rowenc.DatumToEncDatum(types.Bytes, tree.NewDBytes(tree.DBytes(countsBytes))),
			}, nil
		} else {
			__antithesis_instrumentation__.Notify(571988)
		}
	} else {
		__antithesis_instrumentation__.Notify(571989)
	}
	__antithesis_instrumentation__.Notify(571985)
	return nil, sp.DrainHelper()
}

func (sp *bulkRowWriter) work(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(571990)
	kvCh := make(chan row.KVBatch, 10)
	var g ctxgroup.Group

	semaCtx := tree.MakeSemaContext()
	conv, err := row.NewDatumRowConverter(
		ctx, &semaCtx, sp.tableDesc, nil, sp.EvalCtx, kvCh, nil,
		sp.flowCtx.GetRowMetrics(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(571995)
		return err
	} else {
		__antithesis_instrumentation__.Notify(571996)
	}
	__antithesis_instrumentation__.Notify(571991)
	if conv.EvalCtx.SessionData() == nil {
		__antithesis_instrumentation__.Notify(571997)
		panic("uninitialized session data")
	} else {
		__antithesis_instrumentation__.Notify(571998)
	}
	__antithesis_instrumentation__.Notify(571992)

	g = ctxgroup.WithContext(ctx)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(571999)
		return sp.ingestLoop(ctx, kvCh)
	})
	__antithesis_instrumentation__.Notify(571993)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(572000)
		return sp.convertLoop(ctx, kvCh, conv)
	})
	__antithesis_instrumentation__.Notify(571994)
	return g.Wait()
}

func (sp *bulkRowWriter) wrapDupError(ctx context.Context, orig error) error {
	__antithesis_instrumentation__.Notify(572001)
	var typed *kvserverbase.DuplicateKeyError
	if !errors.As(orig, &typed) {
		__antithesis_instrumentation__.Notify(572003)
		return orig
	} else {
		__antithesis_instrumentation__.Notify(572004)
	}
	__antithesis_instrumentation__.Notify(572002)
	v := &roachpb.Value{RawBytes: typed.Value}
	return row.NewUniquenessConstraintViolationError(ctx, sp.tableDesc, typed.Key, v)
}

func (sp *bulkRowWriter) ingestLoop(ctx context.Context, kvCh chan row.KVBatch) error {
	__antithesis_instrumentation__.Notify(572005)
	writeTS := sp.spec.Table.CreateAsOfTime
	const bufferSize = 64 << 20
	adder, err := sp.flowCtx.Cfg.BulkAdder(
		ctx, sp.flowCtx.Cfg.DB, writeTS, kvserverbase.BulkAdderOptions{
			Name:          sp.tableDesc.GetName(),
			MinBufferSize: bufferSize,

			DisallowShadowingBelow: writeTS,
			WriteAtBatchTimestamp:  true,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(572009)
		return err
	} else {
		__antithesis_instrumentation__.Notify(572010)
	}
	__antithesis_instrumentation__.Notify(572006)
	defer adder.Close(ctx)

	ingestKvs := func() error {
		__antithesis_instrumentation__.Notify(572011)
		for kvBatch := range kvCh {
			__antithesis_instrumentation__.Notify(572014)
			for _, kv := range kvBatch.KVs {
				__antithesis_instrumentation__.Notify(572015)
				if err := adder.Add(ctx, kv.Key, kv.Value.RawBytes); err != nil {
					__antithesis_instrumentation__.Notify(572016)
					return sp.wrapDupError(ctx, err)
				} else {
					__antithesis_instrumentation__.Notify(572017)
				}
			}
		}
		__antithesis_instrumentation__.Notify(572012)

		if err := adder.Flush(ctx); err != nil {
			__antithesis_instrumentation__.Notify(572018)
			return sp.wrapDupError(ctx, err)
		} else {
			__antithesis_instrumentation__.Notify(572019)
		}
		__antithesis_instrumentation__.Notify(572013)
		return nil
	}
	__antithesis_instrumentation__.Notify(572007)

	if err := ingestKvs(); err != nil {
		__antithesis_instrumentation__.Notify(572020)
		return err
	} else {
		__antithesis_instrumentation__.Notify(572021)
	}
	__antithesis_instrumentation__.Notify(572008)

	sp.summary = adder.GetSummary()
	return nil
}

func (sp *bulkRowWriter) convertLoop(
	ctx context.Context, kvCh chan row.KVBatch, conv *row.DatumRowConverter,
) error {
	__antithesis_instrumentation__.Notify(572022)
	defer close(kvCh)

	done := false
	alloc := &tree.DatumAlloc{}
	typs := sp.input.OutputTypes()

	for {
		__antithesis_instrumentation__.Notify(572024)
		var rows int64
		for {
			__antithesis_instrumentation__.Notify(572028)
			row, meta := sp.input.Next()
			if meta != nil {
				__antithesis_instrumentation__.Notify(572033)
				if meta.Err != nil {
					__antithesis_instrumentation__.Notify(572035)
					return meta.Err
				} else {
					__antithesis_instrumentation__.Notify(572036)
				}
				__antithesis_instrumentation__.Notify(572034)
				sp.AppendTrailingMeta(*meta)
				continue
			} else {
				__antithesis_instrumentation__.Notify(572037)
			}
			__antithesis_instrumentation__.Notify(572029)
			if row == nil {
				__antithesis_instrumentation__.Notify(572038)
				done = true
				break
			} else {
				__antithesis_instrumentation__.Notify(572039)
			}
			__antithesis_instrumentation__.Notify(572030)
			rows++

			for i, ed := range row {
				__antithesis_instrumentation__.Notify(572040)
				if ed.IsNull() {
					__antithesis_instrumentation__.Notify(572043)
					conv.Datums[i] = tree.DNull
					continue
				} else {
					__antithesis_instrumentation__.Notify(572044)
				}
				__antithesis_instrumentation__.Notify(572041)
				if err := ed.EnsureDecoded(typs[i], alloc); err != nil {
					__antithesis_instrumentation__.Notify(572045)
					return err
				} else {
					__antithesis_instrumentation__.Notify(572046)
				}
				__antithesis_instrumentation__.Notify(572042)
				conv.Datums[i] = ed.Datum
			}
			__antithesis_instrumentation__.Notify(572031)

			if err := conv.Row(ctx, sp.processorID, sp.batchIdxAtomic); err != nil {
				__antithesis_instrumentation__.Notify(572047)
				return err
			} else {
				__antithesis_instrumentation__.Notify(572048)
			}
			__antithesis_instrumentation__.Notify(572032)
			atomic.AddInt64(&sp.batchIdxAtomic, 1)
		}
		__antithesis_instrumentation__.Notify(572025)
		if rows < 1 {
			__antithesis_instrumentation__.Notify(572049)
			break
		} else {
			__antithesis_instrumentation__.Notify(572050)
		}
		__antithesis_instrumentation__.Notify(572026)

		if err := conv.SendBatch(ctx); err != nil {
			__antithesis_instrumentation__.Notify(572051)
			return err
		} else {
			__antithesis_instrumentation__.Notify(572052)
		}
		__antithesis_instrumentation__.Notify(572027)

		if done {
			__antithesis_instrumentation__.Notify(572053)
			break
		} else {
			__antithesis_instrumentation__.Notify(572054)
		}
	}
	__antithesis_instrumentation__.Notify(572023)

	return nil
}
