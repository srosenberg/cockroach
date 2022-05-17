package arrowserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import flatbuffers "github.com/google/flatbuffers/go"

type Footer struct {
	_tab flatbuffers.Table
}

func GetRootAsFooter(buf []byte, offset flatbuffers.UOffsetT) *Footer {
	__antithesis_instrumentation__.Notify(54999)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Footer{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Footer) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55000)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Footer) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55001)
	return rcv._tab
}

func (rcv *Footer) Version() int16 {
	__antithesis_instrumentation__.Notify(55002)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55004)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55005)
	}
	__antithesis_instrumentation__.Notify(55003)
	return 0
}

func (rcv *Footer) MutateVersion(n int16) bool {
	__antithesis_instrumentation__.Notify(55006)
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *Footer) Schema(obj *Schema) *Schema {
	__antithesis_instrumentation__.Notify(55007)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55009)
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			__antithesis_instrumentation__.Notify(55011)
			obj = new(Schema)
		} else {
			__antithesis_instrumentation__.Notify(55012)
		}
		__antithesis_instrumentation__.Notify(55010)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55013)
	}
	__antithesis_instrumentation__.Notify(55008)
	return nil
}

func (rcv *Footer) Dictionaries(obj *Block, j int) bool {
	__antithesis_instrumentation__.Notify(55014)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55016)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 24
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55017)
	}
	__antithesis_instrumentation__.Notify(55015)
	return false
}

func (rcv *Footer) DictionariesLength() int {
	__antithesis_instrumentation__.Notify(55018)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55020)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55021)
	}
	__antithesis_instrumentation__.Notify(55019)
	return 0
}

func (rcv *Footer) RecordBatches(obj *Block, j int) bool {
	__antithesis_instrumentation__.Notify(55022)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55024)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 24
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55025)
	}
	__antithesis_instrumentation__.Notify(55023)
	return false
}

func (rcv *Footer) RecordBatchesLength() int {
	__antithesis_instrumentation__.Notify(55026)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55028)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55029)
	}
	__antithesis_instrumentation__.Notify(55027)
	return 0
}

func FooterStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55030)
	builder.StartObject(4)
}
func FooterAddVersion(builder *flatbuffers.Builder, version int16) {
	__antithesis_instrumentation__.Notify(55031)
	builder.PrependInt16Slot(0, version, 0)
}
func FooterAddSchema(builder *flatbuffers.Builder, schema flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55032)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(schema), 0)
}
func FooterAddDictionaries(builder *flatbuffers.Builder, dictionaries flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55033)
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(dictionaries), 0)
}
func FooterStartDictionariesVector(
	builder *flatbuffers.Builder, numElems int,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55034)
	return builder.StartVector(24, numElems, 8)
}
func FooterAddRecordBatches(builder *flatbuffers.Builder, recordBatches flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55035)
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(recordBatches), 0)
}
func FooterStartRecordBatchesVector(
	builder *flatbuffers.Builder, numElems int,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55036)
	return builder.StartVector(24, numElems, 8)
}
func FooterEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55037)
	return builder.EndObject()
}

type Block struct {
	_tab flatbuffers.Struct
}

func (rcv *Block) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55038)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Block) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55039)
	return rcv._tab.Table
}

func (rcv *Block) Offset() int64 {
	__antithesis_instrumentation__.Notify(55040)
	return rcv._tab.GetInt64(rcv._tab.Pos + flatbuffers.UOffsetT(0))
}

func (rcv *Block) MutateOffset(n int64) bool {
	__antithesis_instrumentation__.Notify(55041)
	return rcv._tab.MutateInt64(rcv._tab.Pos+flatbuffers.UOffsetT(0), n)
}

func (rcv *Block) MetaDataLength() int32 {
	__antithesis_instrumentation__.Notify(55042)
	return rcv._tab.GetInt32(rcv._tab.Pos + flatbuffers.UOffsetT(8))
}

func (rcv *Block) MutateMetaDataLength(n int32) bool {
	__antithesis_instrumentation__.Notify(55043)
	return rcv._tab.MutateInt32(rcv._tab.Pos+flatbuffers.UOffsetT(8), n)
}

func (rcv *Block) BodyLength() int64 {
	__antithesis_instrumentation__.Notify(55044)
	return rcv._tab.GetInt64(rcv._tab.Pos + flatbuffers.UOffsetT(16))
}

func (rcv *Block) MutateBodyLength(n int64) bool {
	__antithesis_instrumentation__.Notify(55045)
	return rcv._tab.MutateInt64(rcv._tab.Pos+flatbuffers.UOffsetT(16), n)
}

func CreateBlock(
	builder *flatbuffers.Builder, offset int64, metaDataLength int32, bodyLength int64,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55046)
	builder.Prep(8, 24)
	builder.PrependInt64(bodyLength)
	builder.Pad(4)
	builder.PrependInt32(metaDataLength)
	builder.PrependInt64(offset)
	return builder.Offset()
}
