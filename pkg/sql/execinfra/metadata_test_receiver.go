package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type MetadataTestReceiver struct {
	ProcessorBase
	input RowSource

	trailingErrMeta []execinfrapb.ProducerMetadata

	senders   []string
	rowCounts map[string]rowNumCounter
}

type rowNumCounter struct {
	expected, actual int32
	seen             util.FastIntSet
	err              error
}

var _ Processor = &MetadataTestReceiver{}
var _ RowSource = &MetadataTestReceiver{}

const metadataTestReceiverProcName = "meta receiver"

func NewMetadataTestReceiver(
	flowCtx *FlowCtx,
	processorID int32,
	input RowSource,
	post *execinfrapb.PostProcessSpec,
	output RowReceiver,
	senders []string,
) (*MetadataTestReceiver, error) {
	__antithesis_instrumentation__.Notify(471020)
	mtr := &MetadataTestReceiver{
		input:     input,
		senders:   senders,
		rowCounts: make(map[string]rowNumCounter),
	}
	if err := mtr.Init(
		mtr,
		post,
		input.OutputTypes(),
		flowCtx,
		processorID,
		output,
		nil,
		ProcStateOpts{
			InputsToDrain: []RowSource{input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(471022)
				var trailingMeta []execinfrapb.ProducerMetadata
				if mtr.rowCounts != nil {
					__antithesis_instrumentation__.Notify(471024)
					if meta := mtr.checkRowNumMetadata(); meta != nil {
						__antithesis_instrumentation__.Notify(471025)
						trailingMeta = append(trailingMeta, *meta)
					} else {
						__antithesis_instrumentation__.Notify(471026)
					}
				} else {
					__antithesis_instrumentation__.Notify(471027)
				}
				__antithesis_instrumentation__.Notify(471023)
				mtr.InternalClose()
				return trailingMeta
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(471028)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(471029)
	}
	__antithesis_instrumentation__.Notify(471021)
	return mtr, nil
}

func (mtr *MetadataTestReceiver) checkRowNumMetadata() *execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(471030)
	defer func() { __antithesis_instrumentation__.Notify(471034); mtr.rowCounts = nil }()
	__antithesis_instrumentation__.Notify(471031)

	if len(mtr.rowCounts) != len(mtr.senders) {
		__antithesis_instrumentation__.Notify(471035)
		var missingSenders string
		for _, sender := range mtr.senders {
			__antithesis_instrumentation__.Notify(471037)
			if _, exists := mtr.rowCounts[sender]; !exists {
				__antithesis_instrumentation__.Notify(471038)
				if missingSenders == "" {
					__antithesis_instrumentation__.Notify(471039)
					missingSenders = sender
				} else {
					__antithesis_instrumentation__.Notify(471040)
					missingSenders += fmt.Sprintf(", %s", sender)
				}
			} else {
				__antithesis_instrumentation__.Notify(471041)
			}
		}
		__antithesis_instrumentation__.Notify(471036)
		return &execinfrapb.ProducerMetadata{
			Err: fmt.Errorf(
				"expected %d metadata senders but found %d; missing %s",
				len(mtr.senders), len(mtr.rowCounts), missingSenders,
			),
		}
	} else {
		__antithesis_instrumentation__.Notify(471042)
	}
	__antithesis_instrumentation__.Notify(471032)
	for id, cnt := range mtr.rowCounts {
		__antithesis_instrumentation__.Notify(471043)
		if cnt.err != nil {
			__antithesis_instrumentation__.Notify(471046)
			return &execinfrapb.ProducerMetadata{Err: cnt.err}
		} else {
			__antithesis_instrumentation__.Notify(471047)
		}
		__antithesis_instrumentation__.Notify(471044)
		if cnt.expected != cnt.actual {
			__antithesis_instrumentation__.Notify(471048)
			return &execinfrapb.ProducerMetadata{
				Err: fmt.Errorf(
					"dropped metadata from sender %s: expected %d RowNum messages but got %d",
					id, cnt.expected, cnt.actual),
			}
		} else {
			__antithesis_instrumentation__.Notify(471049)
		}
		__antithesis_instrumentation__.Notify(471045)
		for i := 0; i < int(cnt.expected); i++ {
			__antithesis_instrumentation__.Notify(471050)
			if !cnt.seen.Contains(i) {
				__antithesis_instrumentation__.Notify(471051)
				return &execinfrapb.ProducerMetadata{
					Err: fmt.Errorf(
						"dropped and repeated metadata from sender %s: have %d messages but missing RowNum #%d",
						id, cnt.expected, i+1),
				}
			} else {
				__antithesis_instrumentation__.Notify(471052)
			}
		}
	}
	__antithesis_instrumentation__.Notify(471033)

	return nil
}

func (mtr *MetadataTestReceiver) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(471053)
	ctx = mtr.StartInternal(ctx, metadataTestReceiverProcName)
	mtr.input.Start(ctx)
}

func (mtr *MetadataTestReceiver) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(471054)
	for {
		__antithesis_instrumentation__.Notify(471055)
		if mtr.State == StateTrailingMeta {
			__antithesis_instrumentation__.Notify(471064)
			if meta := mtr.popTrailingMeta(); meta != nil {
				__antithesis_instrumentation__.Notify(471065)
				return nil, meta
			} else {
				__antithesis_instrumentation__.Notify(471066)
			}

		} else {
			__antithesis_instrumentation__.Notify(471067)
		}
		__antithesis_instrumentation__.Notify(471056)
		if mtr.State == StateExhausted {
			__antithesis_instrumentation__.Notify(471068)
			if len(mtr.trailingErrMeta) > 0 {
				__antithesis_instrumentation__.Notify(471070)
				meta := mtr.trailingErrMeta[0]
				mtr.trailingErrMeta = mtr.trailingErrMeta[1:]
				return nil, &meta
			} else {
				__antithesis_instrumentation__.Notify(471071)
			}
			__antithesis_instrumentation__.Notify(471069)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(471072)
		}
		__antithesis_instrumentation__.Notify(471057)

		row, meta := mtr.input.Next()

		if meta != nil {
			__antithesis_instrumentation__.Notify(471073)
			if meta.RowNum != nil {
				__antithesis_instrumentation__.Notify(471076)
				rowNum := meta.RowNum
				rcnt, exists := mtr.rowCounts[rowNum.SenderID]
				if !exists {
					__antithesis_instrumentation__.Notify(471080)
					rcnt.expected = -1
				} else {
					__antithesis_instrumentation__.Notify(471081)
				}
				__antithesis_instrumentation__.Notify(471077)
				if rcnt.err != nil {
					__antithesis_instrumentation__.Notify(471082)
					return nil, meta
				} else {
					__antithesis_instrumentation__.Notify(471083)
				}
				__antithesis_instrumentation__.Notify(471078)
				if rowNum.LastMsg {
					__antithesis_instrumentation__.Notify(471084)
					if rcnt.expected != -1 {
						__antithesis_instrumentation__.Notify(471086)
						rcnt.err = fmt.Errorf(
							"repeated metadata from reader %s: received more than one RowNum with LastMsg set",
							rowNum.SenderID)
						mtr.rowCounts[rowNum.SenderID] = rcnt
						return nil, meta
					} else {
						__antithesis_instrumentation__.Notify(471087)
					}
					__antithesis_instrumentation__.Notify(471085)
					rcnt.expected = rowNum.RowNum
				} else {
					__antithesis_instrumentation__.Notify(471088)
					rcnt.actual++
					rcnt.seen.Add(int(rowNum.RowNum - 1))
				}
				__antithesis_instrumentation__.Notify(471079)
				mtr.rowCounts[rowNum.SenderID] = rcnt
			} else {
				__antithesis_instrumentation__.Notify(471089)
			}
			__antithesis_instrumentation__.Notify(471074)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(471090)

				mtr.trailingErrMeta = append(mtr.trailingErrMeta, *meta)
				continue
			} else {
				__antithesis_instrumentation__.Notify(471091)
			}
			__antithesis_instrumentation__.Notify(471075)

			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(471092)
		}
		__antithesis_instrumentation__.Notify(471058)

		if row == nil {
			__antithesis_instrumentation__.Notify(471093)
			mtr.moveToTrailingMeta()
			continue
		} else {
			__antithesis_instrumentation__.Notify(471094)
		}
		__antithesis_instrumentation__.Notify(471059)

		outRow, ok, err := mtr.OutputHelper.ProcessRow(mtr.Ctx, row)
		if err != nil {
			__antithesis_instrumentation__.Notify(471095)
			mtr.trailingMeta = append(mtr.trailingMeta, execinfrapb.ProducerMetadata{Err: err})
			continue
		} else {
			__antithesis_instrumentation__.Notify(471096)
		}
		__antithesis_instrumentation__.Notify(471060)
		if outRow == nil {
			__antithesis_instrumentation__.Notify(471097)
			if !ok {
				__antithesis_instrumentation__.Notify(471099)
				mtr.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(471100)
			}
			__antithesis_instrumentation__.Notify(471098)
			continue
		} else {
			__antithesis_instrumentation__.Notify(471101)
		}
		__antithesis_instrumentation__.Notify(471061)

		if mtr.State == StateDraining {
			__antithesis_instrumentation__.Notify(471102)
			continue
		} else {
			__antithesis_instrumentation__.Notify(471103)
		}
		__antithesis_instrumentation__.Notify(471062)

		if !ok {
			__antithesis_instrumentation__.Notify(471104)
			mtr.MoveToDraining(nil)
		} else {
			__antithesis_instrumentation__.Notify(471105)
		}
		__antithesis_instrumentation__.Notify(471063)

		return outRow, nil
	}
}
