package colserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"io"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/cockroachdb/cockroach/pkg/col/colserde/arrowserde"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	flatbuffers "github.com/google/flatbuffers/go"
)

const (
	metadataLengthNumBytes           = 4
	flatbufferBuilderInitialCapacity = 1024
)

func numBuffersForType(t *types.T) int {
	__antithesis_instrumentation__.Notify(55697)

	numBuffers := 3
	switch typeconv.TypeFamilyToCanonicalTypeFamily(t.Family()) {
	case types.BoolFamily, types.FloatFamily, types.IntFamily:
		__antithesis_instrumentation__.Notify(55699)

		numBuffers = 2
	default:
		__antithesis_instrumentation__.Notify(55700)
	}
	__antithesis_instrumentation__.Notify(55698)
	return numBuffers
}

type RecordBatchSerializer struct {
	numBuffers []int

	builder *flatbuffers.Builder
	scratch struct {
		bufferLens     []int
		metadataLength [metadataLengthNumBytes]byte
		padding        []byte
	}
}

func NewRecordBatchSerializer(typs []*types.T) (*RecordBatchSerializer, error) {
	__antithesis_instrumentation__.Notify(55701)
	s := &RecordBatchSerializer{
		numBuffers: make([]int, len(typs)),
		builder:    flatbuffers.NewBuilder(flatbufferBuilderInitialCapacity),
	}
	for i, t := range typs {
		__antithesis_instrumentation__.Notify(55703)
		s.numBuffers[i] = numBuffersForType(t)
	}
	__antithesis_instrumentation__.Notify(55702)

	s.scratch.padding = make([]byte, 7)
	return s, nil
}

func calculatePadding(numBytes int) int {
	__antithesis_instrumentation__.Notify(55704)
	return (8 - (numBytes & 7)) & 7
}

func (s *RecordBatchSerializer) Serialize(
	w io.Writer, data []*array.Data, headerLength int,
) (metadataLen uint32, dataLen uint64, _ error) {
	__antithesis_instrumentation__.Notify(55705)
	if len(data) != len(s.numBuffers) {
		__antithesis_instrumentation__.Notify(55714)
		return 0, 0, errors.Errorf("mismatched schema length and number of columns: %d != %d", len(s.numBuffers), len(data))
	} else {
		__antithesis_instrumentation__.Notify(55715)
	}
	__antithesis_instrumentation__.Notify(55706)
	for i := range data {
		__antithesis_instrumentation__.Notify(55716)
		if data[i].Len() != headerLength {
			__antithesis_instrumentation__.Notify(55718)
			return 0, 0, errors.Errorf("mismatched data lengths at column %d: %d != %d", i, headerLength, data[i].Len())
		} else {
			__antithesis_instrumentation__.Notify(55719)
		}
		__antithesis_instrumentation__.Notify(55717)
		if len(data[i].Buffers()) != s.numBuffers[i] {
			__antithesis_instrumentation__.Notify(55720)
			return 0, 0, errors.Errorf(
				"mismatched number of buffers at column %d: %d != %d", i, len(data[i].Buffers()), s.numBuffers[i],
			)
		} else {
			__antithesis_instrumentation__.Notify(55721)
		}
	}
	__antithesis_instrumentation__.Notify(55707)

	s.builder.Reset()
	s.scratch.bufferLens = s.scratch.bufferLens[:0]
	totalBufferLen := 0

	arrowserde.RecordBatchStartNodesVector(s.builder, len(data))
	for i := len(data) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(55722)
		col := data[i]
		arrowserde.CreateFieldNode(s.builder, int64(col.Len()), int64(col.NullN()))
		buffers := col.Buffers()
		for j := len(buffers) - 1; j >= 0; j-- {
			__antithesis_instrumentation__.Notify(55723)
			bufferLen := 0

			if buffers[j] != nil {
				__antithesis_instrumentation__.Notify(55725)
				bufferLen = buffers[j].Len()
			} else {
				__antithesis_instrumentation__.Notify(55726)
			}
			__antithesis_instrumentation__.Notify(55724)
			s.scratch.bufferLens = append(s.scratch.bufferLens, bufferLen)
			totalBufferLen += bufferLen
		}
	}
	__antithesis_instrumentation__.Notify(55708)
	nodes := s.builder.EndVector(len(data))

	arrowserde.RecordBatchStartBuffersVector(s.builder, len(s.scratch.bufferLens))
	for i, offset := 0, totalBufferLen; i < len(s.scratch.bufferLens); i++ {
		__antithesis_instrumentation__.Notify(55727)
		bufferLen := s.scratch.bufferLens[i]
		offset -= bufferLen
		arrowserde.CreateBuffer(s.builder, int64(offset), int64(bufferLen))
	}
	__antithesis_instrumentation__.Notify(55709)
	buffers := s.builder.EndVector(len(s.scratch.bufferLens))

	arrowserde.RecordBatchStart(s.builder)
	arrowserde.RecordBatchAddLength(s.builder, int64(headerLength))
	arrowserde.RecordBatchAddNodes(s.builder, nodes)
	arrowserde.RecordBatchAddBuffers(s.builder, buffers)
	header := arrowserde.RecordBatchEnd(s.builder)

	arrowserde.MessageStart(s.builder)
	arrowserde.MessageAddVersion(s.builder, arrowserde.MetadataVersionV1)
	arrowserde.MessageAddHeaderType(s.builder, arrowserde.MessageHeaderRecordBatch)
	arrowserde.MessageAddHeader(s.builder, header)
	arrowserde.MessageAddBodyLength(s.builder, int64(totalBufferLen))
	s.builder.Finish(arrowserde.MessageEnd(s.builder))

	metadataBytes := s.builder.FinishedBytes()

	s.scratch.padding = s.scratch.padding[:calculatePadding(metadataLengthNumBytes+len(metadataBytes))]

	metadataLength := uint32(len(metadataBytes) + len(s.scratch.padding))
	binary.LittleEndian.PutUint32(s.scratch.metadataLength[:], metadataLength)
	if _, err := w.Write(s.scratch.metadataLength[:]); err != nil {
		__antithesis_instrumentation__.Notify(55728)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(55729)
	}
	__antithesis_instrumentation__.Notify(55710)

	if _, err := w.Write(metadataBytes); err != nil {
		__antithesis_instrumentation__.Notify(55730)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(55731)
	}
	__antithesis_instrumentation__.Notify(55711)

	if _, err := w.Write(s.scratch.padding); err != nil {
		__antithesis_instrumentation__.Notify(55732)
		return 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(55733)
	}
	__antithesis_instrumentation__.Notify(55712)

	bodyLength := 0
	for i := 0; i < len(data); i++ {
		__antithesis_instrumentation__.Notify(55734)
		buffers := data[i].Buffers()
		for j := 0; j < len(buffers); j++ {
			__antithesis_instrumentation__.Notify(55736)
			var bufferBytes []byte
			if buffers[j] != nil {
				__antithesis_instrumentation__.Notify(55738)

				bufferBytes = buffers[j].Bytes()
			} else {
				__antithesis_instrumentation__.Notify(55739)
			}
			__antithesis_instrumentation__.Notify(55737)
			bodyLength += len(bufferBytes)
			if _, err := w.Write(bufferBytes); err != nil {
				__antithesis_instrumentation__.Notify(55740)
				return 0, 0, err
			} else {
				__antithesis_instrumentation__.Notify(55741)
			}
		}
		__antithesis_instrumentation__.Notify(55735)

		data[i] = nil
	}
	__antithesis_instrumentation__.Notify(55713)

	s.scratch.padding = s.scratch.padding[:calculatePadding(bodyLength)]
	_, err := w.Write(s.scratch.padding)
	bodyLength += len(s.scratch.padding)
	return metadataLength, uint64(bodyLength), err
}

func (s *RecordBatchSerializer) Deserialize(data *[]*array.Data, bytes []byte) (int, error) {
	__antithesis_instrumentation__.Notify(55742)

	metadataLen := int(binary.LittleEndian.Uint32(bytes[:metadataLengthNumBytes]))
	metadata := arrowserde.GetRootAsMessage(
		bytes[metadataLengthNumBytes:metadataLengthNumBytes+metadataLen], 0,
	)

	bodyBytes := bytes[metadataLengthNumBytes+metadataLen : metadataLengthNumBytes+metadataLen+int(metadata.BodyLength())]

	_ = metadata.Version()

	if metadata.HeaderType() != arrowserde.MessageHeaderRecordBatch {
		__antithesis_instrumentation__.Notify(55747)
		return 0, errors.Errorf(
			`cannot decode RecordBatch from %s message`,
			arrowserde.EnumNamesMessageHeader[metadata.HeaderType()],
		)
	} else {
		__antithesis_instrumentation__.Notify(55748)
	}
	__antithesis_instrumentation__.Notify(55743)

	var (
		headerTab flatbuffers.Table
		header    arrowserde.RecordBatch
	)

	if !metadata.Header(&headerTab) {
		__antithesis_instrumentation__.Notify(55749)
		return 0, errors.New(`unable to decode metadata table`)
	} else {
		__antithesis_instrumentation__.Notify(55750)
	}
	__antithesis_instrumentation__.Notify(55744)

	header.Init(headerTab.Bytes, headerTab.Pos)
	if len(s.numBuffers) != header.NodesLength() {
		__antithesis_instrumentation__.Notify(55751)
		return 0, errors.Errorf(
			`mismatched schema and header lengths: %d != %d`, len(s.numBuffers), header.NodesLength(),
		)
	} else {
		__antithesis_instrumentation__.Notify(55752)
	}
	__antithesis_instrumentation__.Notify(55745)

	var (
		node arrowserde.FieldNode
		buf  arrowserde.Buffer
	)
	for fieldIdx, bufferIdx := 0, 0; fieldIdx < len(s.numBuffers); fieldIdx++ {
		__antithesis_instrumentation__.Notify(55753)
		header.Nodes(&node, fieldIdx)

		if node.Length() != header.Length() {
			__antithesis_instrumentation__.Notify(55756)
			return 0, errors.Errorf(
				`mismatched field and header lengths: %d != %d`, node.Length(), header.Length(),
			)
		} else {
			__antithesis_instrumentation__.Notify(55757)
		}
		__antithesis_instrumentation__.Notify(55754)

		buffers := make([]*memory.Buffer, s.numBuffers[fieldIdx])
		for i := 0; i < s.numBuffers[fieldIdx]; i++ {
			__antithesis_instrumentation__.Notify(55758)
			header.Buffers(&buf, bufferIdx)
			bufData := bodyBytes[buf.Offset() : buf.Offset()+buf.Length()]
			if i < len(buffers)-1 {
				__antithesis_instrumentation__.Notify(55760)

				bufData = bufData[:buf.Length():buf.Length()]
			} else {
				__antithesis_instrumentation__.Notify(55761)
			}
			__antithesis_instrumentation__.Notify(55759)
			buffers[i] = memory.NewBufferBytes(bufData)
			bufferIdx++
		}
		__antithesis_instrumentation__.Notify(55755)

		*data = append(
			*data,
			array.NewData(
				nil,
				int(header.Length()),
				buffers,
				nil,
				int(node.NullCount()),
				0,
			),
		)
	}
	__antithesis_instrumentation__.Notify(55746)

	return int(header.Length()), nil
}
