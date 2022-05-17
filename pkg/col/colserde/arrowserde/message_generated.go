package arrowserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import flatbuffers "github.com/google/flatbuffers/go"

type MessageHeader = byte

const (
	MessageHeaderNONE            MessageHeader = 0
	MessageHeaderSchema          MessageHeader = 1
	MessageHeaderDictionaryBatch MessageHeader = 2
	MessageHeaderRecordBatch     MessageHeader = 3
	MessageHeaderTensor          MessageHeader = 4
	MessageHeaderSparseTensor    MessageHeader = 5
)

var EnumNamesMessageHeader = map[MessageHeader]string{
	MessageHeaderNONE:            "NONE",
	MessageHeaderSchema:          "Schema",
	MessageHeaderDictionaryBatch: "DictionaryBatch",
	MessageHeaderRecordBatch:     "RecordBatch",
	MessageHeaderTensor:          "Tensor",
	MessageHeaderSparseTensor:    "SparseTensor",
}

type FieldNode struct {
	_tab flatbuffers.Struct
}

func (rcv *FieldNode) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55047)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *FieldNode) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55048)
	return rcv._tab.Table
}

func (rcv *FieldNode) Length() int64 {
	__antithesis_instrumentation__.Notify(55049)
	return rcv._tab.GetInt64(rcv._tab.Pos + flatbuffers.UOffsetT(0))
}

func (rcv *FieldNode) MutateLength(n int64) bool {
	__antithesis_instrumentation__.Notify(55050)
	return rcv._tab.MutateInt64(rcv._tab.Pos+flatbuffers.UOffsetT(0), n)
}

func (rcv *FieldNode) NullCount() int64 {
	__antithesis_instrumentation__.Notify(55051)
	return rcv._tab.GetInt64(rcv._tab.Pos + flatbuffers.UOffsetT(8))
}

func (rcv *FieldNode) MutateNullCount(n int64) bool {
	__antithesis_instrumentation__.Notify(55052)
	return rcv._tab.MutateInt64(rcv._tab.Pos+flatbuffers.UOffsetT(8), n)
}

func CreateFieldNode(
	builder *flatbuffers.Builder, length int64, nullCount int64,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55053)
	builder.Prep(8, 16)
	builder.PrependInt64(nullCount)
	builder.PrependInt64(length)
	return builder.Offset()
}

type RecordBatch struct {
	_tab flatbuffers.Table
}

func GetRootAsRecordBatch(buf []byte, offset flatbuffers.UOffsetT) *RecordBatch {
	__antithesis_instrumentation__.Notify(55054)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &RecordBatch{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *RecordBatch) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55055)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *RecordBatch) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55056)
	return rcv._tab
}

func (rcv *RecordBatch) Length() int64 {
	__antithesis_instrumentation__.Notify(55057)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55059)
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55060)
	}
	__antithesis_instrumentation__.Notify(55058)
	return 0
}

func (rcv *RecordBatch) MutateLength(n int64) bool {
	__antithesis_instrumentation__.Notify(55061)
	return rcv._tab.MutateInt64Slot(4, n)
}

func (rcv *RecordBatch) Nodes(obj *FieldNode, j int) bool {
	__antithesis_instrumentation__.Notify(55062)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55064)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 16
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55065)
	}
	__antithesis_instrumentation__.Notify(55063)
	return false
}

func (rcv *RecordBatch) NodesLength() int {
	__antithesis_instrumentation__.Notify(55066)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55068)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55069)
	}
	__antithesis_instrumentation__.Notify(55067)
	return 0
}

func (rcv *RecordBatch) Buffers(obj *Buffer, j int) bool {
	__antithesis_instrumentation__.Notify(55070)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55072)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 16
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55073)
	}
	__antithesis_instrumentation__.Notify(55071)
	return false
}

func (rcv *RecordBatch) BuffersLength() int {
	__antithesis_instrumentation__.Notify(55074)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55076)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55077)
	}
	__antithesis_instrumentation__.Notify(55075)
	return 0
}

func RecordBatchStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55078)
	builder.StartObject(3)
}
func RecordBatchAddLength(builder *flatbuffers.Builder, length int64) {
	__antithesis_instrumentation__.Notify(55079)
	builder.PrependInt64Slot(0, length, 0)
}
func RecordBatchAddNodes(builder *flatbuffers.Builder, nodes flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55080)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(nodes), 0)
}
func RecordBatchStartNodesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55081)
	return builder.StartVector(16, numElems, 8)
}
func RecordBatchAddBuffers(builder *flatbuffers.Builder, buffers flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55082)
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(buffers), 0)
}
func RecordBatchStartBuffersVector(
	builder *flatbuffers.Builder, numElems int,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55083)
	return builder.StartVector(16, numElems, 8)
}
func RecordBatchEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55084)
	return builder.EndObject()
}

type DictionaryBatch struct {
	_tab flatbuffers.Table
}

func GetRootAsDictionaryBatch(buf []byte, offset flatbuffers.UOffsetT) *DictionaryBatch {
	__antithesis_instrumentation__.Notify(55085)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &DictionaryBatch{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *DictionaryBatch) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55086)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *DictionaryBatch) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55087)
	return rcv._tab
}

func (rcv *DictionaryBatch) Id() int64 {
	__antithesis_instrumentation__.Notify(55088)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55090)
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55091)
	}
	__antithesis_instrumentation__.Notify(55089)
	return 0
}

func (rcv *DictionaryBatch) MutateId(n int64) bool {
	__antithesis_instrumentation__.Notify(55092)
	return rcv._tab.MutateInt64Slot(4, n)
}

func (rcv *DictionaryBatch) Data(obj *RecordBatch) *RecordBatch {
	__antithesis_instrumentation__.Notify(55093)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55095)
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			__antithesis_instrumentation__.Notify(55097)
			obj = new(RecordBatch)
		} else {
			__antithesis_instrumentation__.Notify(55098)
		}
		__antithesis_instrumentation__.Notify(55096)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55099)
	}
	__antithesis_instrumentation__.Notify(55094)
	return nil
}

func (rcv *DictionaryBatch) IsDelta() byte {
	__antithesis_instrumentation__.Notify(55100)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55102)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55103)
	}
	__antithesis_instrumentation__.Notify(55101)
	return 0
}

func (rcv *DictionaryBatch) MutateIsDelta(n byte) bool {
	__antithesis_instrumentation__.Notify(55104)
	return rcv._tab.MutateByteSlot(8, n)
}

func DictionaryBatchStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55105)
	builder.StartObject(3)
}
func DictionaryBatchAddId(builder *flatbuffers.Builder, id int64) {
	__antithesis_instrumentation__.Notify(55106)
	builder.PrependInt64Slot(0, id, 0)
}
func DictionaryBatchAddData(builder *flatbuffers.Builder, data flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55107)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(data), 0)
}
func DictionaryBatchAddIsDelta(builder *flatbuffers.Builder, isDelta byte) {
	__antithesis_instrumentation__.Notify(55108)
	builder.PrependByteSlot(2, isDelta, 0)
}
func DictionaryBatchEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55109)
	return builder.EndObject()
}

type Message struct {
	_tab flatbuffers.Table
}

func GetRootAsMessage(buf []byte, offset flatbuffers.UOffsetT) *Message {
	__antithesis_instrumentation__.Notify(55110)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Message{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Message) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55111)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Message) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55112)
	return rcv._tab
}

func (rcv *Message) Version() int16 {
	__antithesis_instrumentation__.Notify(55113)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55115)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55116)
	}
	__antithesis_instrumentation__.Notify(55114)
	return 0
}

func (rcv *Message) MutateVersion(n int16) bool {
	__antithesis_instrumentation__.Notify(55117)
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *Message) HeaderType() byte {
	__antithesis_instrumentation__.Notify(55118)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55120)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55121)
	}
	__antithesis_instrumentation__.Notify(55119)
	return 0
}

func (rcv *Message) MutateHeaderType(n byte) bool {
	__antithesis_instrumentation__.Notify(55122)
	return rcv._tab.MutateByteSlot(6, n)
}

func (rcv *Message) Header(obj *flatbuffers.Table) bool {
	__antithesis_instrumentation__.Notify(55123)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55125)
		rcv._tab.Union(obj, o)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55126)
	}
	__antithesis_instrumentation__.Notify(55124)
	return false
}

func (rcv *Message) BodyLength() int64 {
	__antithesis_instrumentation__.Notify(55127)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55129)
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55130)
	}
	__antithesis_instrumentation__.Notify(55128)
	return 0
}

func (rcv *Message) MutateBodyLength(n int64) bool {
	__antithesis_instrumentation__.Notify(55131)
	return rcv._tab.MutateInt64Slot(10, n)
}

func MessageStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55132)
	builder.StartObject(4)
}
func MessageAddVersion(builder *flatbuffers.Builder, version int16) {
	__antithesis_instrumentation__.Notify(55133)
	builder.PrependInt16Slot(0, version, 0)
}
func MessageAddHeaderType(builder *flatbuffers.Builder, headerType byte) {
	__antithesis_instrumentation__.Notify(55134)
	builder.PrependByteSlot(1, headerType, 0)
}
func MessageAddHeader(builder *flatbuffers.Builder, header flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55135)
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(header), 0)
}
func MessageAddBodyLength(builder *flatbuffers.Builder, bodyLength int64) {
	__antithesis_instrumentation__.Notify(55136)
	builder.PrependInt64Slot(3, bodyLength, 0)
}
func MessageEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55137)
	return builder.EndObject()
}
