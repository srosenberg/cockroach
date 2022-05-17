package arrowserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import flatbuffers "github.com/google/flatbuffers/go"

type MetadataVersion = int16

const (
	MetadataVersionV1 MetadataVersion = 0

	MetadataVersionV2 MetadataVersion = 1

	MetadataVersionV3 MetadataVersion = 2

	MetadataVersionV4 MetadataVersion = 3
)

var EnumNamesMetadataVersion = map[MetadataVersion]string{
	MetadataVersionV1: "V1",
	MetadataVersionV2: "V2",
	MetadataVersionV3: "V3",
	MetadataVersionV4: "V4",
}

type UnionMode = int16

const (
	UnionModeSparse UnionMode = 0
	UnionModeDense  UnionMode = 1
)

var EnumNamesUnionMode = map[UnionMode]string{
	UnionModeSparse: "Sparse",
	UnionModeDense:  "Dense",
}

type Precision = int16

const (
	PrecisionHALF   Precision = 0
	PrecisionSINGLE Precision = 1
	PrecisionDOUBLE Precision = 2
)

var EnumNamesPrecision = map[Precision]string{
	PrecisionHALF:   "HALF",
	PrecisionSINGLE: "SINGLE",
	PrecisionDOUBLE: "DOUBLE",
}

type DateUnit = int16

const (
	DateUnitDAY         DateUnit = 0
	DateUnitMILLISECOND DateUnit = 1
)

var EnumNamesDateUnit = map[DateUnit]string{
	DateUnitDAY:         "DAY",
	DateUnitMILLISECOND: "MILLISECOND",
}

type TimeUnit = int16

const (
	TimeUnitSECOND      TimeUnit = 0
	TimeUnitMILLISECOND TimeUnit = 1
	TimeUnitMICROSECOND TimeUnit = 2
	TimeUnitNANOSECOND  TimeUnit = 3
)

var EnumNamesTimeUnit = map[TimeUnit]string{
	TimeUnitSECOND:      "SECOND",
	TimeUnitMILLISECOND: "MILLISECOND",
	TimeUnitMICROSECOND: "MICROSECOND",
	TimeUnitNANOSECOND:  "NANOSECOND",
}

type IntervalUnit = int16

const (
	IntervalUnitYEAR_MONTH IntervalUnit = 0
	IntervalUnitDAY_TIME   IntervalUnit = 1
)

var EnumNamesIntervalUnit = map[IntervalUnit]string{
	IntervalUnitYEAR_MONTH: "YEAR_MONTH",
	IntervalUnitDAY_TIME:   "DAY_TIME",
}

type Type = byte

const (
	TypeNONE            Type = 0
	TypeNull            Type = 1
	TypeInt             Type = 2
	TypeFloatingPoint   Type = 3
	TypeBinary          Type = 4
	TypeUtf8            Type = 5
	TypeBool            Type = 6
	TypeDecimal         Type = 7
	TypeDate            Type = 8
	TypeTime            Type = 9
	TypeTimestamp       Type = 10
	TypeInterval        Type = 11
	TypeList            Type = 12
	TypeStruct_         Type = 13
	TypeUnion           Type = 14
	TypeFixedSizeBinary Type = 15
	TypeFixedSizeList   Type = 16
	TypeMap             Type = 17
)

var EnumNamesType = map[Type]string{
	TypeNONE:            "NONE",
	TypeNull:            "Null",
	TypeInt:             "Int",
	TypeFloatingPoint:   "FloatingPoint",
	TypeBinary:          "Binary",
	TypeUtf8:            "Utf8",
	TypeBool:            "Bool",
	TypeDecimal:         "Decimal",
	TypeDate:            "Date",
	TypeTime:            "Time",
	TypeTimestamp:       "Timestamp",
	TypeInterval:        "Interval",
	TypeList:            "List",
	TypeStruct_:         "Struct_",
	TypeUnion:           "Union",
	TypeFixedSizeBinary: "FixedSizeBinary",
	TypeFixedSizeList:   "FixedSizeList",
	TypeMap:             "Map",
}

type Endianness = int16

const (
	EndiannessLittle Endianness = 0
	EndiannessBig    Endianness = 1
)

var EnumNamesEndianness = map[Endianness]string{
	EndiannessLittle: "Little",
	EndiannessBig:    "Big",
}

type Null struct {
	_tab flatbuffers.Table
}

func GetRootAsNull(buf []byte, offset flatbuffers.UOffsetT) *Null {
	__antithesis_instrumentation__.Notify(55138)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Null{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Null) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55139)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Null) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55140)
	return rcv._tab
}

func NullStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55141)
	builder.StartObject(0)
}
func NullEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55142)
	return builder.EndObject()
}

type Struct_ struct {
	_tab flatbuffers.Table
}

func GetRootAsStruct_(buf []byte, offset flatbuffers.UOffsetT) *Struct_ {
	__antithesis_instrumentation__.Notify(55143)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Struct_{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Struct_) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55144)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Struct_) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55145)
	return rcv._tab
}

func Struct_Start(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55146)
	builder.StartObject(0)
}
func Struct_End(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55147)
	return builder.EndObject()
}

type List struct {
	_tab flatbuffers.Table
}

func GetRootAsList(buf []byte, offset flatbuffers.UOffsetT) *List {
	__antithesis_instrumentation__.Notify(55148)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &List{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *List) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55149)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *List) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55150)
	return rcv._tab
}

func ListStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55151)
	builder.StartObject(0)
}
func ListEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55152)
	return builder.EndObject()
}

type FixedSizeList struct {
	_tab flatbuffers.Table
}

func GetRootAsFixedSizeList(buf []byte, offset flatbuffers.UOffsetT) *FixedSizeList {
	__antithesis_instrumentation__.Notify(55153)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &FixedSizeList{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *FixedSizeList) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55154)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *FixedSizeList) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55155)
	return rcv._tab
}

func (rcv *FixedSizeList) ListSize() int32 {
	__antithesis_instrumentation__.Notify(55156)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55158)
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55159)
	}
	__antithesis_instrumentation__.Notify(55157)
	return 0
}

func (rcv *FixedSizeList) MutateListSize(n int32) bool {
	__antithesis_instrumentation__.Notify(55160)
	return rcv._tab.MutateInt32Slot(4, n)
}

func FixedSizeListStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55161)
	builder.StartObject(1)
}
func FixedSizeListAddListSize(builder *flatbuffers.Builder, listSize int32) {
	__antithesis_instrumentation__.Notify(55162)
	builder.PrependInt32Slot(0, listSize, 0)
}
func FixedSizeListEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55163)
	return builder.EndObject()
}

type Map struct {
	_tab flatbuffers.Table
}

func GetRootAsMap(buf []byte, offset flatbuffers.UOffsetT) *Map {
	__antithesis_instrumentation__.Notify(55164)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Map{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Map) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55165)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Map) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55166)
	return rcv._tab
}

func (rcv *Map) KeysSorted() byte {
	__antithesis_instrumentation__.Notify(55167)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55169)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55170)
	}
	__antithesis_instrumentation__.Notify(55168)
	return 0
}

func (rcv *Map) MutateKeysSorted(n byte) bool {
	__antithesis_instrumentation__.Notify(55171)
	return rcv._tab.MutateByteSlot(4, n)
}

func MapStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55172)
	builder.StartObject(1)
}
func MapAddKeysSorted(builder *flatbuffers.Builder, keysSorted byte) {
	__antithesis_instrumentation__.Notify(55173)
	builder.PrependByteSlot(0, keysSorted, 0)
}
func MapEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55174)
	return builder.EndObject()
}

type Union struct {
	_tab flatbuffers.Table
}

func GetRootAsUnion(buf []byte, offset flatbuffers.UOffsetT) *Union {
	__antithesis_instrumentation__.Notify(55175)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Union{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Union) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55176)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Union) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55177)
	return rcv._tab
}

func (rcv *Union) Mode() int16 {
	__antithesis_instrumentation__.Notify(55178)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55180)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55181)
	}
	__antithesis_instrumentation__.Notify(55179)
	return 0
}

func (rcv *Union) MutateMode(n int16) bool {
	__antithesis_instrumentation__.Notify(55182)
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *Union) TypeIds(j int) int32 {
	__antithesis_instrumentation__.Notify(55183)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55185)
		a := rcv._tab.Vector(o)
		return rcv._tab.GetInt32(a + flatbuffers.UOffsetT(j*4))
	} else {
		__antithesis_instrumentation__.Notify(55186)
	}
	__antithesis_instrumentation__.Notify(55184)
	return 0
}

func (rcv *Union) TypeIdsLength() int {
	__antithesis_instrumentation__.Notify(55187)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55189)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55190)
	}
	__antithesis_instrumentation__.Notify(55188)
	return 0
}

func UnionStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55191)
	builder.StartObject(2)
}
func UnionAddMode(builder *flatbuffers.Builder, mode int16) {
	__antithesis_instrumentation__.Notify(55192)
	builder.PrependInt16Slot(0, mode, 0)
}
func UnionAddTypeIds(builder *flatbuffers.Builder, typeIds flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55193)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(typeIds), 0)
}
func UnionStartTypeIdsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55194)
	return builder.StartVector(4, numElems, 4)
}
func UnionEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55195)
	return builder.EndObject()
}

type Int struct {
	_tab flatbuffers.Table
}

func GetRootAsInt(buf []byte, offset flatbuffers.UOffsetT) *Int {
	__antithesis_instrumentation__.Notify(55196)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Int{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Int) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55197)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Int) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55198)
	return rcv._tab
}

func (rcv *Int) BitWidth() int32 {
	__antithesis_instrumentation__.Notify(55199)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55201)
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55202)
	}
	__antithesis_instrumentation__.Notify(55200)
	return 0
}

func (rcv *Int) MutateBitWidth(n int32) bool {
	__antithesis_instrumentation__.Notify(55203)
	return rcv._tab.MutateInt32Slot(4, n)
}

func (rcv *Int) IsSigned() byte {
	__antithesis_instrumentation__.Notify(55204)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55206)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55207)
	}
	__antithesis_instrumentation__.Notify(55205)
	return 0
}

func (rcv *Int) MutateIsSigned(n byte) bool {
	__antithesis_instrumentation__.Notify(55208)
	return rcv._tab.MutateByteSlot(6, n)
}

func IntStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55209)
	builder.StartObject(2)
}
func IntAddBitWidth(builder *flatbuffers.Builder, bitWidth int32) {
	__antithesis_instrumentation__.Notify(55210)
	builder.PrependInt32Slot(0, bitWidth, 0)
}
func IntAddIsSigned(builder *flatbuffers.Builder, isSigned byte) {
	__antithesis_instrumentation__.Notify(55211)
	builder.PrependByteSlot(1, isSigned, 0)
}
func IntEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55212)
	return builder.EndObject()
}

type FloatingPoint struct {
	_tab flatbuffers.Table
}

func GetRootAsFloatingPoint(buf []byte, offset flatbuffers.UOffsetT) *FloatingPoint {
	__antithesis_instrumentation__.Notify(55213)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &FloatingPoint{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *FloatingPoint) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55214)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *FloatingPoint) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55215)
	return rcv._tab
}

func (rcv *FloatingPoint) Precision() int16 {
	__antithesis_instrumentation__.Notify(55216)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55218)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55219)
	}
	__antithesis_instrumentation__.Notify(55217)
	return 0
}

func (rcv *FloatingPoint) MutatePrecision(n int16) bool {
	__antithesis_instrumentation__.Notify(55220)
	return rcv._tab.MutateInt16Slot(4, n)
}

func FloatingPointStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55221)
	builder.StartObject(1)
}
func FloatingPointAddPrecision(builder *flatbuffers.Builder, precision int16) {
	__antithesis_instrumentation__.Notify(55222)
	builder.PrependInt16Slot(0, precision, 0)
}
func FloatingPointEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55223)
	return builder.EndObject()
}

type Utf8 struct {
	_tab flatbuffers.Table
}

func GetRootAsUtf8(buf []byte, offset flatbuffers.UOffsetT) *Utf8 {
	__antithesis_instrumentation__.Notify(55224)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Utf8{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Utf8) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55225)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Utf8) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55226)
	return rcv._tab
}

func Utf8Start(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55227)
	builder.StartObject(0)
}
func Utf8End(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55228)
	return builder.EndObject()
}

type Binary struct {
	_tab flatbuffers.Table
}

func GetRootAsBinary(buf []byte, offset flatbuffers.UOffsetT) *Binary {
	__antithesis_instrumentation__.Notify(55229)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Binary{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Binary) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55230)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Binary) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55231)
	return rcv._tab
}

func BinaryStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55232)
	builder.StartObject(0)
}
func BinaryEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55233)
	return builder.EndObject()
}

type FixedSizeBinary struct {
	_tab flatbuffers.Table
}

func GetRootAsFixedSizeBinary(buf []byte, offset flatbuffers.UOffsetT) *FixedSizeBinary {
	__antithesis_instrumentation__.Notify(55234)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &FixedSizeBinary{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *FixedSizeBinary) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55235)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *FixedSizeBinary) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55236)
	return rcv._tab
}

func (rcv *FixedSizeBinary) ByteWidth() int32 {
	__antithesis_instrumentation__.Notify(55237)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55239)
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55240)
	}
	__antithesis_instrumentation__.Notify(55238)
	return 0
}

func (rcv *FixedSizeBinary) MutateByteWidth(n int32) bool {
	__antithesis_instrumentation__.Notify(55241)
	return rcv._tab.MutateInt32Slot(4, n)
}

func FixedSizeBinaryStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55242)
	builder.StartObject(1)
}
func FixedSizeBinaryAddByteWidth(builder *flatbuffers.Builder, byteWidth int32) {
	__antithesis_instrumentation__.Notify(55243)
	builder.PrependInt32Slot(0, byteWidth, 0)
}
func FixedSizeBinaryEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55244)
	return builder.EndObject()
}

type Bool struct {
	_tab flatbuffers.Table
}

func GetRootAsBool(buf []byte, offset flatbuffers.UOffsetT) *Bool {
	__antithesis_instrumentation__.Notify(55245)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Bool{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Bool) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55246)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Bool) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55247)
	return rcv._tab
}

func BoolStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55248)
	builder.StartObject(0)
}
func BoolEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55249)
	return builder.EndObject()
}

type Decimal struct {
	_tab flatbuffers.Table
}

func GetRootAsDecimal(buf []byte, offset flatbuffers.UOffsetT) *Decimal {
	__antithesis_instrumentation__.Notify(55250)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Decimal{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Decimal) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55251)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Decimal) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55252)
	return rcv._tab
}

func (rcv *Decimal) Precision() int32 {
	__antithesis_instrumentation__.Notify(55253)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55255)
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55256)
	}
	__antithesis_instrumentation__.Notify(55254)
	return 0
}

func (rcv *Decimal) MutatePrecision(n int32) bool {
	__antithesis_instrumentation__.Notify(55257)
	return rcv._tab.MutateInt32Slot(4, n)
}

func (rcv *Decimal) Scale() int32 {
	__antithesis_instrumentation__.Notify(55258)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55260)
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55261)
	}
	__antithesis_instrumentation__.Notify(55259)
	return 0
}

func (rcv *Decimal) MutateScale(n int32) bool {
	__antithesis_instrumentation__.Notify(55262)
	return rcv._tab.MutateInt32Slot(6, n)
}

func DecimalStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55263)
	builder.StartObject(2)
}
func DecimalAddPrecision(builder *flatbuffers.Builder, precision int32) {
	__antithesis_instrumentation__.Notify(55264)
	builder.PrependInt32Slot(0, precision, 0)
}
func DecimalAddScale(builder *flatbuffers.Builder, scale int32) {
	__antithesis_instrumentation__.Notify(55265)
	builder.PrependInt32Slot(1, scale, 0)
}
func DecimalEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55266)
	return builder.EndObject()
}

type Date struct {
	_tab flatbuffers.Table
}

func GetRootAsDate(buf []byte, offset flatbuffers.UOffsetT) *Date {
	__antithesis_instrumentation__.Notify(55267)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Date{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Date) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55268)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Date) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55269)
	return rcv._tab
}

func (rcv *Date) Unit() int16 {
	__antithesis_instrumentation__.Notify(55270)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55272)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55273)
	}
	__antithesis_instrumentation__.Notify(55271)
	return 1
}

func (rcv *Date) MutateUnit(n int16) bool {
	__antithesis_instrumentation__.Notify(55274)
	return rcv._tab.MutateInt16Slot(4, n)
}

func DateStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55275)
	builder.StartObject(1)
}
func DateAddUnit(builder *flatbuffers.Builder, unit int16) {
	__antithesis_instrumentation__.Notify(55276)
	builder.PrependInt16Slot(0, unit, 1)
}
func DateEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55277)
	return builder.EndObject()
}

type Time struct {
	_tab flatbuffers.Table
}

func GetRootAsTime(buf []byte, offset flatbuffers.UOffsetT) *Time {
	__antithesis_instrumentation__.Notify(55278)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Time{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Time) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55279)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Time) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55280)
	return rcv._tab
}

func (rcv *Time) Unit() int16 {
	__antithesis_instrumentation__.Notify(55281)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55283)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55284)
	}
	__antithesis_instrumentation__.Notify(55282)
	return 1
}

func (rcv *Time) MutateUnit(n int16) bool {
	__antithesis_instrumentation__.Notify(55285)
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *Time) BitWidth() int32 {
	__antithesis_instrumentation__.Notify(55286)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55288)
		return rcv._tab.GetInt32(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55289)
	}
	__antithesis_instrumentation__.Notify(55287)
	return 32
}

func (rcv *Time) MutateBitWidth(n int32) bool {
	__antithesis_instrumentation__.Notify(55290)
	return rcv._tab.MutateInt32Slot(6, n)
}

func TimeStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55291)
	builder.StartObject(2)
}
func TimeAddUnit(builder *flatbuffers.Builder, unit int16) {
	__antithesis_instrumentation__.Notify(55292)
	builder.PrependInt16Slot(0, unit, 1)
}
func TimeAddBitWidth(builder *flatbuffers.Builder, bitWidth int32) {
	__antithesis_instrumentation__.Notify(55293)
	builder.PrependInt32Slot(1, bitWidth, 32)
}
func TimeEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55294)
	return builder.EndObject()
}

type Timestamp struct {
	_tab flatbuffers.Table
}

func GetRootAsTimestamp(buf []byte, offset flatbuffers.UOffsetT) *Timestamp {
	__antithesis_instrumentation__.Notify(55295)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Timestamp{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Timestamp) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55296)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Timestamp) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55297)
	return rcv._tab
}

func (rcv *Timestamp) Unit() int16 {
	__antithesis_instrumentation__.Notify(55298)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55300)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55301)
	}
	__antithesis_instrumentation__.Notify(55299)
	return 0
}

func (rcv *Timestamp) MutateUnit(n int16) bool {
	__antithesis_instrumentation__.Notify(55302)
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *Timestamp) Timezone() []byte {
	__antithesis_instrumentation__.Notify(55303)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55305)
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55306)
	}
	__antithesis_instrumentation__.Notify(55304)
	return nil
}

func TimestampStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55307)
	builder.StartObject(2)
}
func TimestampAddUnit(builder *flatbuffers.Builder, unit int16) {
	__antithesis_instrumentation__.Notify(55308)
	builder.PrependInt16Slot(0, unit, 0)
}
func TimestampAddTimezone(builder *flatbuffers.Builder, timezone flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55309)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(timezone), 0)
}
func TimestampEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55310)
	return builder.EndObject()
}

type Interval struct {
	_tab flatbuffers.Table
}

func GetRootAsInterval(buf []byte, offset flatbuffers.UOffsetT) *Interval {
	__antithesis_instrumentation__.Notify(55311)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Interval{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Interval) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55312)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Interval) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55313)
	return rcv._tab
}

func (rcv *Interval) Unit() int16 {
	__antithesis_instrumentation__.Notify(55314)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55316)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55317)
	}
	__antithesis_instrumentation__.Notify(55315)
	return 0
}

func (rcv *Interval) MutateUnit(n int16) bool {
	__antithesis_instrumentation__.Notify(55318)
	return rcv._tab.MutateInt16Slot(4, n)
}

func IntervalStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55319)
	builder.StartObject(1)
}
func IntervalAddUnit(builder *flatbuffers.Builder, unit int16) {
	__antithesis_instrumentation__.Notify(55320)
	builder.PrependInt16Slot(0, unit, 0)
}
func IntervalEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55321)
	return builder.EndObject()
}

type KeyValue struct {
	_tab flatbuffers.Table
}

func GetRootAsKeyValue(buf []byte, offset flatbuffers.UOffsetT) *KeyValue {
	__antithesis_instrumentation__.Notify(55322)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &KeyValue{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *KeyValue) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55323)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *KeyValue) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55324)
	return rcv._tab
}

func (rcv *KeyValue) Key() []byte {
	__antithesis_instrumentation__.Notify(55325)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55327)
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55328)
	}
	__antithesis_instrumentation__.Notify(55326)
	return nil
}

func (rcv *KeyValue) Value() []byte {
	__antithesis_instrumentation__.Notify(55329)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55331)
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55332)
	}
	__antithesis_instrumentation__.Notify(55330)
	return nil
}

func KeyValueStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55333)
	builder.StartObject(2)
}
func KeyValueAddKey(builder *flatbuffers.Builder, key flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55334)
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(key), 0)
}
func KeyValueAddValue(builder *flatbuffers.Builder, value flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55335)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(value), 0)
}
func KeyValueEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55336)
	return builder.EndObject()
}

type DictionaryEncoding struct {
	_tab flatbuffers.Table
}

func GetRootAsDictionaryEncoding(buf []byte, offset flatbuffers.UOffsetT) *DictionaryEncoding {
	__antithesis_instrumentation__.Notify(55337)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &DictionaryEncoding{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *DictionaryEncoding) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55338)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *DictionaryEncoding) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55339)
	return rcv._tab
}

func (rcv *DictionaryEncoding) Id() int64 {
	__antithesis_instrumentation__.Notify(55340)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55342)
		return rcv._tab.GetInt64(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55343)
	}
	__antithesis_instrumentation__.Notify(55341)
	return 0
}

func (rcv *DictionaryEncoding) MutateId(n int64) bool {
	__antithesis_instrumentation__.Notify(55344)
	return rcv._tab.MutateInt64Slot(4, n)
}

func (rcv *DictionaryEncoding) IndexType(obj *Int) *Int {
	__antithesis_instrumentation__.Notify(55345)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55347)
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			__antithesis_instrumentation__.Notify(55349)
			obj = new(Int)
		} else {
			__antithesis_instrumentation__.Notify(55350)
		}
		__antithesis_instrumentation__.Notify(55348)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55351)
	}
	__antithesis_instrumentation__.Notify(55346)
	return nil
}

func (rcv *DictionaryEncoding) IsOrdered() byte {
	__antithesis_instrumentation__.Notify(55352)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55354)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55355)
	}
	__antithesis_instrumentation__.Notify(55353)
	return 0
}

func (rcv *DictionaryEncoding) MutateIsOrdered(n byte) bool {
	__antithesis_instrumentation__.Notify(55356)
	return rcv._tab.MutateByteSlot(8, n)
}

func DictionaryEncodingStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55357)
	builder.StartObject(3)
}
func DictionaryEncodingAddId(builder *flatbuffers.Builder, id int64) {
	__antithesis_instrumentation__.Notify(55358)
	builder.PrependInt64Slot(0, id, 0)
}
func DictionaryEncodingAddIndexType(builder *flatbuffers.Builder, indexType flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55359)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(indexType), 0)
}
func DictionaryEncodingAddIsOrdered(builder *flatbuffers.Builder, isOrdered byte) {
	__antithesis_instrumentation__.Notify(55360)
	builder.PrependByteSlot(2, isOrdered, 0)
}
func DictionaryEncodingEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55361)
	return builder.EndObject()
}

type Field struct {
	_tab flatbuffers.Table
}

func GetRootAsField(buf []byte, offset flatbuffers.UOffsetT) *Field {
	__antithesis_instrumentation__.Notify(55362)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Field{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Field) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55363)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Field) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55364)
	return rcv._tab
}

func (rcv *Field) Name() []byte {
	__antithesis_instrumentation__.Notify(55365)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55367)
		return rcv._tab.ByteVector(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55368)
	}
	__antithesis_instrumentation__.Notify(55366)
	return nil
}

func (rcv *Field) Nullable() byte {
	__antithesis_instrumentation__.Notify(55369)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55371)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55372)
	}
	__antithesis_instrumentation__.Notify(55370)
	return 0
}

func (rcv *Field) MutateNullable(n byte) bool {
	__antithesis_instrumentation__.Notify(55373)
	return rcv._tab.MutateByteSlot(6, n)
}

func (rcv *Field) TypeType() byte {
	__antithesis_instrumentation__.Notify(55374)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55376)
		return rcv._tab.GetByte(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55377)
	}
	__antithesis_instrumentation__.Notify(55375)
	return 0
}

func (rcv *Field) MutateTypeType(n byte) bool {
	__antithesis_instrumentation__.Notify(55378)
	return rcv._tab.MutateByteSlot(8, n)
}

func (rcv *Field) Type(obj *flatbuffers.Table) bool {
	__antithesis_instrumentation__.Notify(55379)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(10))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55381)
		rcv._tab.Union(obj, o)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55382)
	}
	__antithesis_instrumentation__.Notify(55380)
	return false
}

func (rcv *Field) Dictionary(obj *DictionaryEncoding) *DictionaryEncoding {
	__antithesis_instrumentation__.Notify(55383)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(12))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55385)
		x := rcv._tab.Indirect(o + rcv._tab.Pos)
		if obj == nil {
			__antithesis_instrumentation__.Notify(55387)
			obj = new(DictionaryEncoding)
		} else {
			__antithesis_instrumentation__.Notify(55388)
		}
		__antithesis_instrumentation__.Notify(55386)
		obj.Init(rcv._tab.Bytes, x)
		return obj
	} else {
		__antithesis_instrumentation__.Notify(55389)
	}
	__antithesis_instrumentation__.Notify(55384)
	return nil
}

func (rcv *Field) Children(obj *Field, j int) bool {
	__antithesis_instrumentation__.Notify(55390)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55392)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55393)
	}
	__antithesis_instrumentation__.Notify(55391)
	return false
}

func (rcv *Field) ChildrenLength() int {
	__antithesis_instrumentation__.Notify(55394)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(14))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55396)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55397)
	}
	__antithesis_instrumentation__.Notify(55395)
	return 0
}

func (rcv *Field) CustomMetadata(obj *KeyValue, j int) bool {
	__antithesis_instrumentation__.Notify(55398)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55400)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55401)
	}
	__antithesis_instrumentation__.Notify(55399)
	return false
}

func (rcv *Field) CustomMetadataLength() int {
	__antithesis_instrumentation__.Notify(55402)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(16))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55404)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55405)
	}
	__antithesis_instrumentation__.Notify(55403)
	return 0
}

func FieldStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55406)
	builder.StartObject(7)
}
func FieldAddName(builder *flatbuffers.Builder, name flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55407)
	builder.PrependUOffsetTSlot(0, flatbuffers.UOffsetT(name), 0)
}
func FieldAddNullable(builder *flatbuffers.Builder, nullable byte) {
	__antithesis_instrumentation__.Notify(55408)
	builder.PrependByteSlot(1, nullable, 0)
}
func FieldAddTypeType(builder *flatbuffers.Builder, typeType byte) {
	__antithesis_instrumentation__.Notify(55409)
	builder.PrependByteSlot(2, typeType, 0)
}
func FieldAddType(builder *flatbuffers.Builder, type_ flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55410)
	builder.PrependUOffsetTSlot(3, flatbuffers.UOffsetT(type_), 0)
}
func FieldAddDictionary(builder *flatbuffers.Builder, dictionary flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55411)
	builder.PrependUOffsetTSlot(4, flatbuffers.UOffsetT(dictionary), 0)
}
func FieldAddChildren(builder *flatbuffers.Builder, children flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55412)
	builder.PrependUOffsetTSlot(5, flatbuffers.UOffsetT(children), 0)
}
func FieldStartChildrenVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55413)
	return builder.StartVector(4, numElems, 4)
}
func FieldAddCustomMetadata(builder *flatbuffers.Builder, customMetadata flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55414)
	builder.PrependUOffsetTSlot(6, flatbuffers.UOffsetT(customMetadata), 0)
}
func FieldStartCustomMetadataVector(
	builder *flatbuffers.Builder, numElems int,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55415)
	return builder.StartVector(4, numElems, 4)
}
func FieldEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55416)
	return builder.EndObject()
}

type Buffer struct {
	_tab flatbuffers.Struct
}

func (rcv *Buffer) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55417)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Buffer) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55418)
	return rcv._tab.Table
}

func (rcv *Buffer) Offset() int64 {
	__antithesis_instrumentation__.Notify(55419)
	return rcv._tab.GetInt64(rcv._tab.Pos + flatbuffers.UOffsetT(0))
}

func (rcv *Buffer) MutateOffset(n int64) bool {
	__antithesis_instrumentation__.Notify(55420)
	return rcv._tab.MutateInt64(rcv._tab.Pos+flatbuffers.UOffsetT(0), n)
}

func (rcv *Buffer) Length() int64 {
	__antithesis_instrumentation__.Notify(55421)
	return rcv._tab.GetInt64(rcv._tab.Pos + flatbuffers.UOffsetT(8))
}

func (rcv *Buffer) MutateLength(n int64) bool {
	__antithesis_instrumentation__.Notify(55422)
	return rcv._tab.MutateInt64(rcv._tab.Pos+flatbuffers.UOffsetT(8), n)
}

func CreateBuffer(builder *flatbuffers.Builder, offset int64, length int64) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55423)
	builder.Prep(8, 16)
	builder.PrependInt64(length)
	builder.PrependInt64(offset)
	return builder.Offset()
}

type Schema struct {
	_tab flatbuffers.Table
}

func GetRootAsSchema(buf []byte, offset flatbuffers.UOffsetT) *Schema {
	__antithesis_instrumentation__.Notify(55424)
	n := flatbuffers.GetUOffsetT(buf[offset:])
	x := &Schema{}
	x.Init(buf, n+offset)
	return x
}

func (rcv *Schema) Init(buf []byte, i flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55425)
	rcv._tab.Bytes = buf
	rcv._tab.Pos = i
}

func (rcv *Schema) Table() flatbuffers.Table {
	__antithesis_instrumentation__.Notify(55426)
	return rcv._tab
}

func (rcv *Schema) Endianness() int16 {
	__antithesis_instrumentation__.Notify(55427)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(4))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55429)
		return rcv._tab.GetInt16(o + rcv._tab.Pos)
	} else {
		__antithesis_instrumentation__.Notify(55430)
	}
	__antithesis_instrumentation__.Notify(55428)
	return 0
}

func (rcv *Schema) MutateEndianness(n int16) bool {
	__antithesis_instrumentation__.Notify(55431)
	return rcv._tab.MutateInt16Slot(4, n)
}

func (rcv *Schema) Fields(obj *Field, j int) bool {
	__antithesis_instrumentation__.Notify(55432)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55434)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55435)
	}
	__antithesis_instrumentation__.Notify(55433)
	return false
}

func (rcv *Schema) FieldsLength() int {
	__antithesis_instrumentation__.Notify(55436)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(6))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55438)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55439)
	}
	__antithesis_instrumentation__.Notify(55437)
	return 0
}

func (rcv *Schema) CustomMetadata(obj *KeyValue, j int) bool {
	__antithesis_instrumentation__.Notify(55440)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55442)
		x := rcv._tab.Vector(o)
		x += flatbuffers.UOffsetT(j) * 4
		x = rcv._tab.Indirect(x)
		obj.Init(rcv._tab.Bytes, x)
		return true
	} else {
		__antithesis_instrumentation__.Notify(55443)
	}
	__antithesis_instrumentation__.Notify(55441)
	return false
}

func (rcv *Schema) CustomMetadataLength() int {
	__antithesis_instrumentation__.Notify(55444)
	o := flatbuffers.UOffsetT(rcv._tab.Offset(8))
	if o != 0 {
		__antithesis_instrumentation__.Notify(55446)
		return rcv._tab.VectorLen(o)
	} else {
		__antithesis_instrumentation__.Notify(55447)
	}
	__antithesis_instrumentation__.Notify(55445)
	return 0
}

func SchemaStart(builder *flatbuffers.Builder) {
	__antithesis_instrumentation__.Notify(55448)
	builder.StartObject(3)
}
func SchemaAddEndianness(builder *flatbuffers.Builder, endianness int16) {
	__antithesis_instrumentation__.Notify(55449)
	builder.PrependInt16Slot(0, endianness, 0)
}
func SchemaAddFields(builder *flatbuffers.Builder, fields flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55450)
	builder.PrependUOffsetTSlot(1, flatbuffers.UOffsetT(fields), 0)
}
func SchemaStartFieldsVector(builder *flatbuffers.Builder, numElems int) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55451)
	return builder.StartVector(4, numElems, 4)
}
func SchemaAddCustomMetadata(builder *flatbuffers.Builder, customMetadata flatbuffers.UOffsetT) {
	__antithesis_instrumentation__.Notify(55452)
	builder.PrependUOffsetTSlot(2, flatbuffers.UOffsetT(customMetadata), 0)
}
func SchemaStartCustomMetadataVector(
	builder *flatbuffers.Builder, numElems int,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55453)
	return builder.StartVector(4, numElems, 4)
}
func SchemaEnd(builder *flatbuffers.Builder) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55454)
	return builder.EndObject()
}
