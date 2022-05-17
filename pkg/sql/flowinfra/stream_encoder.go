package flowinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

const PreferredEncoding = descpb.DatumEncoding_ASCENDING_KEY

type StreamEncoder struct {
	infos            []execinfrapb.DatumInfo
	infosInitialized bool

	rowBuf       []byte
	numEmptyRows int
	metadata     []execinfrapb.RemoteProducerMetadata

	headerSent bool

	typingSent bool
	alloc      tree.DatumAlloc

	msg    execinfrapb.ProducerMessage
	msgHdr execinfrapb.ProducerHeader
}

func (se *StreamEncoder) HasHeaderBeenSent() bool {
	__antithesis_instrumentation__.Notify(492131)
	return se.headerSent
}

func (se *StreamEncoder) SetHeaderFields(flowID execinfrapb.FlowID, streamID execinfrapb.StreamID) {
	__antithesis_instrumentation__.Notify(492132)
	se.msgHdr.FlowID = flowID
	se.msgHdr.StreamID = streamID
}

func (se *StreamEncoder) Init(types []*types.T) {
	__antithesis_instrumentation__.Notify(492133)
	se.infos = make([]execinfrapb.DatumInfo, len(types))
	for i := range types {
		__antithesis_instrumentation__.Notify(492134)
		se.infos[i].Type = types[i]
	}
}

func (se *StreamEncoder) AddMetadata(ctx context.Context, meta execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(492135)
	se.metadata = append(se.metadata, execinfrapb.LocalMetaToRemoteProducerMeta(ctx, meta))
}

func (se *StreamEncoder) AddRow(row rowenc.EncDatumRow) error {
	__antithesis_instrumentation__.Notify(492136)
	if se.infos == nil {
		__antithesis_instrumentation__.Notify(492142)
		panic("Init not called")
	} else {
		__antithesis_instrumentation__.Notify(492143)
	}
	__antithesis_instrumentation__.Notify(492137)
	if len(se.infos) != len(row) {
		__antithesis_instrumentation__.Notify(492144)
		return errors.Errorf("inconsistent row length: expected %d, got %d", len(se.infos), len(row))
	} else {
		__antithesis_instrumentation__.Notify(492145)
	}
	__antithesis_instrumentation__.Notify(492138)
	if !se.infosInitialized {
		__antithesis_instrumentation__.Notify(492146)

		for i := range row {
			__antithesis_instrumentation__.Notify(492148)
			enc, ok := row[i].Encoding()
			if !ok {
				__antithesis_instrumentation__.Notify(492151)
				enc = PreferredEncoding
			} else {
				__antithesis_instrumentation__.Notify(492152)
			}
			__antithesis_instrumentation__.Notify(492149)
			sType := se.infos[i].Type
			if enc != descpb.DatumEncoding_VALUE && func() bool {
				__antithesis_instrumentation__.Notify(492153)
				return (colinfo.CanHaveCompositeKeyEncoding(sType) || func() bool {
					__antithesis_instrumentation__.Notify(492154)
					return colinfo.MustBeValueEncoded(sType) == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(492155)

				enc = descpb.DatumEncoding_VALUE
			} else {
				__antithesis_instrumentation__.Notify(492156)
			}
			__antithesis_instrumentation__.Notify(492150)
			se.infos[i].Encoding = enc
		}
		__antithesis_instrumentation__.Notify(492147)
		se.infosInitialized = true
	} else {
		__antithesis_instrumentation__.Notify(492157)
	}
	__antithesis_instrumentation__.Notify(492139)
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(492158)
		se.numEmptyRows++
		return nil
	} else {
		__antithesis_instrumentation__.Notify(492159)
	}
	__antithesis_instrumentation__.Notify(492140)
	for i := range row {
		__antithesis_instrumentation__.Notify(492160)
		var err error
		se.rowBuf, err = row[i].Encode(se.infos[i].Type, &se.alloc, se.infos[i].Encoding, se.rowBuf)
		if err != nil {
			__antithesis_instrumentation__.Notify(492161)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492162)
		}
	}
	__antithesis_instrumentation__.Notify(492141)
	return nil
}

func (se *StreamEncoder) FormMessage(ctx context.Context) *execinfrapb.ProducerMessage {
	__antithesis_instrumentation__.Notify(492163)
	msg := &se.msg
	msg.Header = nil
	msg.Data.RawBytes = se.rowBuf
	msg.Data.NumEmptyRows = int32(se.numEmptyRows)
	msg.Data.Metadata = make([]execinfrapb.RemoteProducerMetadata, len(se.metadata))
	copy(msg.Data.Metadata, se.metadata)
	se.metadata = se.metadata[:0]

	if !se.headerSent {
		__antithesis_instrumentation__.Notify(492166)
		msg.Header = &se.msgHdr
		se.headerSent = true
	} else {
		__antithesis_instrumentation__.Notify(492167)
	}
	__antithesis_instrumentation__.Notify(492164)
	if !se.typingSent {
		__antithesis_instrumentation__.Notify(492168)
		if se.infosInitialized {
			__antithesis_instrumentation__.Notify(492169)
			msg.Typing = se.infos
			se.typingSent = true
		} else {
			__antithesis_instrumentation__.Notify(492170)
		}
	} else {
		__antithesis_instrumentation__.Notify(492171)
		msg.Typing = nil
	}
	__antithesis_instrumentation__.Notify(492165)

	se.rowBuf = se.rowBuf[:0]
	se.numEmptyRows = 0
	return msg
}
