package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
)

type MetadataTestSender struct {
	ProcessorBase
	input RowSource
	id    string

	sendRowNumMeta bool
	rowNumCnt      int32
}

var _ Processor = &MetadataTestSender{}
var _ RowSource = &MetadataTestSender{}

const metadataTestSenderProcName = "meta sender"

func NewMetadataTestSender(
	flowCtx *FlowCtx,
	processorID int32,
	input RowSource,
	post *execinfrapb.PostProcessSpec,
	output RowReceiver,
	id string,
) (*MetadataTestSender, error) {
	__antithesis_instrumentation__.Notify(471106)
	mts := &MetadataTestSender{input: input, id: id}
	if err := mts.Init(
		mts,
		post,
		input.OutputTypes(),
		flowCtx,
		processorID,
		output,
		nil,
		ProcStateOpts{
			InputsToDrain: []RowSource{mts.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(471108)
				mts.InternalClose()

				meta := execinfrapb.ProducerMetadata{
					RowNum: &execinfrapb.RemoteProducerMetadata_RowNum{
						RowNum:   mts.rowNumCnt,
						SenderID: mts.id,
						LastMsg:  true,
					},
				}
				return []execinfrapb.ProducerMetadata{meta}
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(471109)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(471110)
	}
	__antithesis_instrumentation__.Notify(471107)
	return mts, nil
}

func (mts *MetadataTestSender) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(471111)
	ctx = mts.StartInternal(ctx, metadataTestSenderProcName)
	mts.input.Start(ctx)
}

func (mts *MetadataTestSender) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(471112)

	if mts.sendRowNumMeta {
		__antithesis_instrumentation__.Notify(471115)
		mts.sendRowNumMeta = false
		mts.rowNumCnt++
		return nil, &execinfrapb.ProducerMetadata{
			RowNum: &execinfrapb.RemoteProducerMetadata_RowNum{
				RowNum:   mts.rowNumCnt,
				SenderID: mts.id,
				LastMsg:  false,
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(471116)
	}
	__antithesis_instrumentation__.Notify(471113)

	for mts.State == StateRunning {
		__antithesis_instrumentation__.Notify(471117)
		row, meta := mts.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(471120)

			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(471121)
		}
		__antithesis_instrumentation__.Notify(471118)
		if row == nil {
			__antithesis_instrumentation__.Notify(471122)
			mts.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(471123)
		}
		__antithesis_instrumentation__.Notify(471119)

		if outRow := mts.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(471124)
			mts.sendRowNumMeta = true
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(471125)
		}
	}
	__antithesis_instrumentation__.Notify(471114)
	return nil, mts.DrainHelper()
}
