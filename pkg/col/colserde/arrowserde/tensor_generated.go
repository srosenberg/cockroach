package arrowserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import flatbuffers "github.com/google/flatbuffers/go"

type SparseTensorIndex = byte

const (
	SparseTensorIndexNONE                 SparseTensorIndex = 0
	SparseTensorIndexSparseTensorIndexCOO SparseTensorIndex = 1
	SparseTensorIndexSparseMatrixIndexCSR SparseTensorIndex = 2
)

var EnumNamesSparseTensorIndex = map[SparseTensorIndex]string{
	SparseTensorIndexNONE:                 "NONE",
	SparseTensorIndexSparseTensorIndexCOO: "SparseTensorIndexCOO",
	SparseTensorIndexSparseMatrixIndexCSR: "SparseMatrixIndexCSR",
}

type TensorDim struct {
	_tab flatbuffers.Table
}

func GetRootAsTensorDim(buf []byte, offset flatbuffers.UOffsetT) *TensorDim {
	__antithesis_instrumentation__.Notify(55455)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &TensorDim{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *TensorDim) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55456)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *TensorDim) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55457)
	return rcv._tab
}

func (rcv *TensorDim) Size() int64 {
	__antithesis_instrumentation__.Notify(55458)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55460)
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55461)
	}
	__antithesis_instrumentation__.Notify(55459)
	return 0
}

func (rcv *TensorDim) MutateSize(n int64) bool {
	__antithesis_instrumentation__.Notify(55462)
	return rcv._tab.MutateInt64Slot(4, n)
}

func (rcv *TensorDim) Name() []byte {
	__antithesis_instrumentation__.Notify(55463)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55465)
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55466)
	}
	__antithesis_instrumentation__.Notify(55464)
	return nil
}

func TensorDimStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55467)
	builder.StartObject(2)
}
func TensorDimAddSize(builder *flatbuffers.Builder, size int64) {
	__antithesis_instrumentation__.Notify(55468)
	builder.PrependInt64Slot(0, size, 0)
}
func TensorDimAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55469)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(name), 0)
}
func TensorDimEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55470)
	return builder.EndObject()
}

type Tensor struct {
	_tab flatbuffers.Table
}

func GetRootAsTensor(buf []byte, offset flatbuffers.UOffsetT) *Tensor {
	__antithesis_instrumentation__.Notify(55471)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Tensor{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Tensor) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55472)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Tensor) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55473)
	return rcv._tab
}

func (rcv *Tensor) TypeType() byte {
	__antithesis_instrumentation__.Notify(55474)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55476)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55477)
	}
	__antithesis_instrumentation__.Notify(55475)
	return 0
}

func (rcv *Tensor) MutateTypeType(n byte) bool {
	__antithesis_instrumentation__.Notify(55478)
	return rcv._tab.MutateByteSlot(4, n)
}

func (rcv *Tensor) Type(obj *flatbuffers.Table) bool {
	__antithesis_instrumentation__.Notify(55479)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55481)
		rcv._tab.Union(obj, o)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55482)
	}
	__antithesis_instrumentation__.Notify(55480)
	return false
}

func (rcv *Tensor) Shape(obj *TensorDim, j int) bool {
	__antithesis_instrumentation__.Notify(55483)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55485)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55486)
	}
	__antithesis_instrumentation__.Notify(55484)
	return false
}

func (rcv *Tensor) ShapeLength() int {
	__antithesis_instrumentation__.Notify(55487)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55489)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55490)
	}
	__antithesis_instrumentation__.Notify(55488)
	return 0
}

func (rcv *Tensor) Strides(j int) int64 {
	__antithesis_instrumentation__.Notify(55491)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55493)
		a := rcv._tab.Vector(o)
		return rcv._tab.GetInt64(a + flatbuffers.UOffsetT(j*8))
	} else {
		__antithesis_instrumentation__.Notify(55494)
	}
	__antithesis_instrumentation__.Notify(55492)
	return 0
}

func (rcv *Tensor) StridesLength() int {
	__antithesis_instrumentation__.Notify(55495)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55497)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55498)
	}
	__antithesis_instrumentation__.Notify(55496)
	return 0
}

func (rcv *Tensor) Data(obj *Buffer) *Buffer {
	__antithesis_instrumentation__.Notify(55499)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55501)
		x := o + rcv._tab.Pos
		if obj == nil {
			__antithesis_instrumentation__.Notify(55503)
			obj = new(Buffer)
		} else {
			__antithesis_instrumentation__.Notify(55504)
		}
		__antithesis_instrumentation__.Notify(55502)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55505)
	}
	__antithesis_instrumentation__.Notify(55500)
	return nil
}

func TensorStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55506)
	builder.StartObject(5)
}
func TensorAddTypeType(builder *flatbuffers.Builder, typeType byte) {
	__antithesis_instrumentation__.Notify(55507)
	builder.PrependByteSlot(0, typeType, 0)
}
func TensorAddType(builder *flatbuffers.Builder, type_ flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55508)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(type_), 0)
}
func TensorAddShape(builder *flatbuffers.Builder, shape flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55509)
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(shape), 0)
}
func TensorStartShapeVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55510)
	return builder.StartVector(4, numElems, 4)
}
func TensorAddStrides(builder *flatbuffers.Builder, strides flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55511)
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(strides), 0)
}
func TensorStartStridesVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55512)
	return builder.StartVector(8, numElems, 8)
}
func TensorAddData(builder *flatbuffers.Builder, data flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55513)
	builder.PrependStructSlot(4, flatbuffers.UOffsetT(data), 0)
}
func TensorEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55514)
	return builder.EndObject()
}

type SparseTensorIndexCOO struct {
	_tab flatbuffers.Table
}

func GetRootAsSparseTensorIndexCOO(buf []byte, offset flatbuffers.UOffsetT) *SparseTensorIndexCOO {
	__antithesis_instrumentation__.Notify(55515)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &SparseTensorIndexCOO{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *SparseTensorIndexCOO) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55516)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *SparseTensorIndexCOO) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55517)
	return rcv._tab
}

func (rcv *SparseTensorIndexCOO) IndicesBuffer(obj *Buffer) *Buffer {
	__antithesis_instrumentation__.Notify(55518)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55520)
		x := o + rcv._tab.Pos
		if obj == nil {
			__antithesis_instrumentation__.Notify(55522)
			obj = new(Buffer)
		} else {
			__antithesis_instrumentation__.Notify(55523)
		}
		__antithesis_instrumentation__.Notify(55521)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55524)
	}
	__antithesis_instrumentation__.Notify(55519)
	return nil
}

func SparseTensorIndexCOOStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55525)
	builder.StartObject(1)
}
func SparseTensorIndexCOOAddIndicesBuffer(
	builder *flatbuffers.Builder, indicesBuffer flatbuffers.UOffsetT,
) {
	__antithesis_instrumentation__.Notify(55526)
	builder.PrependStructSlot(0, flatbuffers.UOffsetT(indicesBuffer), 0)
}
func SparseTensorIndexCOOEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55527)
	return builder.EndObject()
}

type SparseMatrixIndexCSR struct {
	_tab flatbuffers.Table
}

func GetRootAsSparseMatrixIndexCSR(buf []byte, offset flatbuffers.UOffsetT) *SparseMatrixIndexCSR {
	__antithesis_instrumentation__.Notify(55528)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &SparseMatrixIndexCSR{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *SparseMatrixIndexCSR) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55529)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *SparseMatrixIndexCSR) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55530)
	return rcv._tab
}

func (rcv *SparseMatrixIndexCSR) IndptrBuffer(obj *Buffer) *Buffer {
	__antithesis_instrumentation__.Notify(55531)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55533)
		x := o + rcv._tab.Pos
		if obj == nil {
			__antithesis_instrumentation__.Notify(55535)
			obj = new(Buffer)
		} else {
			__antithesis_instrumentation__.Notify(55536)
		}
		__antithesis_instrumentation__.Notify(55534)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55537)
	}
	__antithesis_instrumentation__.Notify(55532)
	return nil
}

func (rcv *SparseMatrixIndexCSR) IndicesBuffer(obj *Buffer) *Buffer {
	__antithesis_instrumentation__.Notify(55538)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55540)
		x := o + rcv._tab.Pos
		if obj == nil {
			__antithesis_instrumentation__.Notify(55542)
			obj = new(Buffer)
		} else {
			__antithesis_instrumentation__.Notify(55543)
		}
		__antithesis_instrumentation__.Notify(55541)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55544)
	}
	__antithesis_instrumentation__.Notify(55539)
	return nil
}

func SparseMatrixIndexCSRStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55545)
	builder.StartObject(2)
}
func SparseMatrixIndexCSRAddIndptrBuffer(
	builder *flatbuffers.Builder, indptrBuffer flatbuffers.UOffsetT,
) {
	__antithesis_instrumentation__.Notify(55546)
	builder.PrependStructSlot(0, flatbuffers.UOffsetT(indptrBuffer), 0)
}
func SparseMatrixIndexCSRAddIndicesBuffer(
	builder *flatbuffers.Builder, indicesBuffer flatbuffers.UOffsetT,
) {
	__antithesis_instrumentation__.Notify(55547)
	builder.PrependStructSlot(1, flatbuffers.UOffsetT(indicesBuffer), 0)
}
func SparseMatrixIndexCSREnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55548)
	return builder.EndObject()
}

type SparseTensor struct {
	_tab flatbuffers.Table
}

func GetRootAsSparseTensor(buf []byte, offset flatbuffers.UOffsetT) *SparseTensor {
	__antithesis_instrumentation__.Notify(55549)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &SparseTensor{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *SparseTensor) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55550)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *SparseTensor) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55551)
	return rcv._tab
}

func (rcv *SparseTensor) TypeType() byte {
	__antithesis_instrumentation__.Notify(55552)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55554)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55555)
	}
	__antithesis_instrumentation__.Notify(55553)
	return 0
}

func (rcv *SparseTensor) MutateTypeType(n byte) bool {
	__antithesis_instrumentation__.Notify(55556)
	return rcv._tab.MutateByteSlot(4, n)
}

func (rcv *SparseTensor) Type(obj *flatbuffers.Table) bool {
	__antithesis_instrumentation__.Notify(55557)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55559)
		rcv._tab.Union(obj, o)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55560)
	}
	__antithesis_instrumentation__.Notify(55558)
	return false
}

func (rcv *SparseTensor) Shape(obj *TensorDim, j int) bool {
	__antithesis_instrumentation__.Notify(55561)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55563)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55564)
	}
	__antithesis_instrumentation__.Notify(55562)
	return false
}

func (rcv *SparseTensor) ShapeLength() int {
	__antithesis_instrumentation__.Notify(55565)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55567)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55568)
	}
	__antithesis_instrumentation__.Notify(55566)
	return 0
}

func (rcv *SparseTensor) NonZeroLength() int64 {
	__antithesis_instrumentation__.Notify(55569)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55571)
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55572)
	}
	__antithesis_instrumentation__.Notify(55570)
	return 0
}

func (rcv *SparseTensor) MutateNonZeroLength(n int64) bool {
	__antithesis_instrumentation__.Notify(55573)
	return rcv._tab.MutateInt64Slot(10, n)
}

func (rcv *SparseTensor) SparseIndexType() byte {
	__antithesis_instrumentation__.Notify(55574)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55576)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55577)
	}
	__antithesis_instrumentation__.Notify(55575)
	return 0
}

func (rcv *SparseTensor) MutateSparseIndexType(n byte) bool {
	__antithesis_instrumentation__.Notify(55578)
	return rcv._tab.MutateByteSlot(12, n)
}

func (rcv *SparseTensor) SparseIndex(obj *flatbuffers.Table) bool {
	__antithesis_instrumentation__.Notify(55579)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55581)
		rcv._tab.Union(obj, o)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55582)
	}
	__antithesis_instrumentation__.Notify(55580)
	return false
}

func (rcv *SparseTensor) Data(obj *Buffer) *Buffer {
	__antithesis_instrumentation__.Notify(55583)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55585)
		x := o + rcv._tab.Pos
		if obj == nil {
			__antithesis_instrumentation__.Notify(55587)
			obj = new(Buffer)
		} else {
			__antithesis_instrumentation__.Notify(55588)
		}
		__antithesis_instrumentation__.Notify(55586)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55589)
	}
	__antithesis_instrumentation__.Notify(55584)
	return nil
}

func SparseTensorStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55590)
	builder.StartObject(7)
}
func SparseTensorAddTypeType(builder *flatbuffers.Builder, typeType byte) {
	__antithesis_instrumentation__.Notify(55591)
	builder.PrependByteSlot(0, typeType, 0)
}
func SparseTensorAddType(builder *flatbuffers.Builder, type_ flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55592)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(type_), 0)
}
func SparseTensorAddShape(builder *flatbuffers.Builder, shape flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55593)
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(shape), 0)
}
func SparseTensorStartShapeVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55594)
	return builder.StartVector(4, numElems, 4)
}
func SparseTensorAddNonZeroLength(builder *flatbuffers.Builder, nonZeroLength int64) {
	__antithesis_instrumentation__.Notify(55595)
	builder.PrependInt64Slot(3, nonZeroLength, 0)
}
func SparseTensorAddSparseIndexType(builder *flatbuffers.Builder, sparseIndexType byte) {
	__antithesis_instrumentation__.Notify(55596)
	builder.PrependByteSlot(4, sparseIndexType, 0)
}
func SparseTensorAddSparseIndex(builder *flatbuffers.Builder, sparseIndex flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55597)
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(sparseIndex), 0)
}
func SparseTensorAddData(builder *flatbuffers.Builder, data flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55598)
	builder.PrependStructSlot(6, flatbuffers.UOffsetT(data), 0)
}
func SparseTensorEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55599)
	return builder.EndObject()
}
