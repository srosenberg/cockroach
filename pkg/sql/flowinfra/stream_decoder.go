package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type StreamDecoder struct {
	typing       []execinfrapb.DatumInfo
	data         []byte
	numEmptyRows int
	metadata     []execinfrapb.ProducerMetadata
	rowAlloc     rowenc.EncDatumRowAlloc

	headerReceived bool
	typingReceived bool
}

func (sd *StreamDecoder) AddMessage(ctx context.Context, msg *execinfrapb.ProducerMessage) error {
	__antithesis_instrumentation__.Notify(492076)
	if msg.Header != nil {
		__antithesis_instrumentation__.Notify(492082)
		if sd.headerReceived {
			__antithesis_instrumentation__.Notify(492084)
			return errors.Errorf("received multiple headers")
		} else {
			__antithesis_instrumentation__.Notify(492085)
		}
		__antithesis_instrumentation__.Notify(492083)
		sd.headerReceived = true
	} else {
		__antithesis_instrumentation__.Notify(492086)
	}
	__antithesis_instrumentation__.Notify(492077)
	if msg.Typing != nil {
		__antithesis_instrumentation__.Notify(492087)
		if sd.typingReceived {
			__antithesis_instrumentation__.Notify(492089)
			return errors.Errorf("typing information received multiple times")
		} else {
			__antithesis_instrumentation__.Notify(492090)
		}
		__antithesis_instrumentation__.Notify(492088)
		sd.typingReceived = true
		sd.typing = msg.Typing
	} else {
		__antithesis_instrumentation__.Notify(492091)
	}
	__antithesis_instrumentation__.Notify(492078)

	if len(msg.Data.RawBytes) > 0 {
		__antithesis_instrumentation__.Notify(492092)
		if !sd.headerReceived || func() bool {
			__antithesis_instrumentation__.Notify(492094)
			return !sd.typingReceived == true
		}() == true {
			__antithesis_instrumentation__.Notify(492095)
			return errors.Errorf("received data before header and/or typing info")
		} else {
			__antithesis_instrumentation__.Notify(492096)
		}
		__antithesis_instrumentation__.Notify(492093)

		if len(sd.data) == 0 {
			__antithesis_instrumentation__.Notify(492097)

			sd.data = msg.Data.RawBytes[:len(msg.Data.RawBytes):len(msg.Data.RawBytes)]
		} else {
			__antithesis_instrumentation__.Notify(492098)

			sd.data = append(sd.data, msg.Data.RawBytes...)
		}
	} else {
		__antithesis_instrumentation__.Notify(492099)
	}
	__antithesis_instrumentation__.Notify(492079)
	if msg.Data.NumEmptyRows > 0 {
		__antithesis_instrumentation__.Notify(492100)
		if len(msg.Data.RawBytes) > 0 {
			__antithesis_instrumentation__.Notify(492102)
			return errors.Errorf("received both data and empty rows")
		} else {
			__antithesis_instrumentation__.Notify(492103)
		}
		__antithesis_instrumentation__.Notify(492101)
		sd.numEmptyRows += int(msg.Data.NumEmptyRows)
	} else {
		__antithesis_instrumentation__.Notify(492104)
	}
	__antithesis_instrumentation__.Notify(492080)
	if len(msg.Data.Metadata) > 0 {
		__antithesis_instrumentation__.Notify(492105)
		for _, md := range msg.Data.Metadata {
			__antithesis_instrumentation__.Notify(492106)
			meta, ok := execinfrapb.RemoteProducerMetaToLocalMeta(ctx, md)
			if !ok {
				__antithesis_instrumentation__.Notify(492108)

				continue
			} else {
				__antithesis_instrumentation__.Notify(492109)
			}
			__antithesis_instrumentation__.Notify(492107)
			sd.metadata = append(sd.metadata, meta)
		}
	} else {
		__antithesis_instrumentation__.Notify(492110)
	}
	__antithesis_instrumentation__.Notify(492081)
	return nil
}

func (sd *StreamDecoder) GetRow(
	rowBuf rowenc.EncDatumRow,
) (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata, error) {
	__antithesis_instrumentation__.Notify(492111)
	if len(sd.metadata) != 0 {
		__antithesis_instrumentation__.Notify(492117)
		r := &sd.metadata[0]
		sd.metadata = sd.metadata[1:]
		return nil, r, nil
	} else {
		__antithesis_instrumentation__.Notify(492118)
	}
	__antithesis_instrumentation__.Notify(492112)

	if sd.numEmptyRows > 0 {
		__antithesis_instrumentation__.Notify(492119)
		sd.numEmptyRows--
		row := make(rowenc.EncDatumRow, 0)
		return row, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(492120)
	}
	__antithesis_instrumentation__.Notify(492113)

	if len(sd.data) == 0 {
		__antithesis_instrumentation__.Notify(492121)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(492122)
	}
	__antithesis_instrumentation__.Notify(492114)
	rowLen := len(sd.typing)
	if cap(rowBuf) >= rowLen {
		__antithesis_instrumentation__.Notify(492123)
		rowBuf = rowBuf[:rowLen]
	} else {
		__antithesis_instrumentation__.Notify(492124)
		rowBuf = sd.rowAlloc.AllocRow(rowLen)
	}
	__antithesis_instrumentation__.Notify(492115)
	for i := range rowBuf {
		__antithesis_instrumentation__.Notify(492125)
		var err error
		rowBuf[i], sd.data, err = rowenc.EncDatumFromBuffer(
			sd.typing[i].Type, sd.typing[i].Encoding, sd.data,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(492126)

			*sd = StreamDecoder{}
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(492127)
		}
	}
	__antithesis_instrumentation__.Notify(492116)
	return rowBuf, nil, nil
}

func (sd *StreamDecoder) Types() []*types.T {
	__antithesis_instrumentation__.Notify(492128)
	types := make([]*types.T, len(sd.typing))
	for i := range types {
		__antithesis_instrumentation__.Notify(492130)
		types[i] = sd.typing[i].Type
	}
	__antithesis_instrumentation__.Notify(492129)
	return types
}
