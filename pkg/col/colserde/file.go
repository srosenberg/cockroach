package colserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/colserde/arrowserde"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	mmap "github.com/edsrzf/mmap-go"
	flatbuffers "github.com/google/flatbuffers/go"
)

const fileMagic = `ARROW1`

var fileMagicPadding [8 - len(fileMagic)]byte

type fileBlock struct {
	offset      int64
	metadataLen int32
	bodyLen     int64
}

type FileSerializer struct {
	scratch [4]byte

	w    *countingWriter
	typs []*types.T
	fb   *flatbuffers.Builder
	a    *ArrowBatchConverter
	rb   *RecordBatchSerializer

	recordBatches []fileBlock
}

func NewFileSerializer(w io.Writer, typs []*types.T) (*FileSerializer, error) {
	__antithesis_instrumentation__.Notify(55600)
	a, err := NewArrowBatchConverter(typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(55603)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(55604)
	}
	__antithesis_instrumentation__.Notify(55601)
	rb, err := NewRecordBatchSerializer(typs)
	if err != nil {
		__antithesis_instrumentation__.Notify(55605)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(55606)
	}
	__antithesis_instrumentation__.Notify(55602)
	s := &FileSerializer{
		typs: typs,
		fb:   flatbuffers.NewBuilder(flatbufferBuilderInitialCapacity),
		a:    a,
		rb:   rb,
	}
	return s, s.Reset(w)
}

func (s *FileSerializer) Reset(w io.Writer) error {
	__antithesis_instrumentation__.Notify(55607)
	if s.w != nil {
		__antithesis_instrumentation__.Notify(55612)
		return errors.New(`Finish must be called before Reset`)
	} else {
		__antithesis_instrumentation__.Notify(55613)
	}
	__antithesis_instrumentation__.Notify(55608)
	s.w = &countingWriter{wrapped: w}
	s.recordBatches = s.recordBatches[:0]
	if _, err := io.WriteString(s.w, fileMagic); err != nil {
		__antithesis_instrumentation__.Notify(55614)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55615)
	}
	__antithesis_instrumentation__.Notify(55609)

	if _, err := s.w.Write(fileMagicPadding[:]); err != nil {
		__antithesis_instrumentation__.Notify(55616)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55617)
	}
	__antithesis_instrumentation__.Notify(55610)

	s.fb.Reset()
	messageOffset := schemaMessage(s.fb, s.typs)
	s.fb.Finish(messageOffset)
	schemaBytes := s.fb.FinishedBytes()
	if _, err := s.w.Write(schemaBytes); err != nil {
		__antithesis_instrumentation__.Notify(55618)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55619)
	}
	__antithesis_instrumentation__.Notify(55611)
	_, err := s.w.Write(make([]byte, calculatePadding(len(schemaBytes))))
	return err
}

func (s *FileSerializer) AppendBatch(batch coldata.Batch) error {
	__antithesis_instrumentation__.Notify(55620)
	offset := int64(s.w.written)

	arrow, err := s.a.BatchToArrow(batch)
	if err != nil {
		__antithesis_instrumentation__.Notify(55623)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55624)
	}
	__antithesis_instrumentation__.Notify(55621)
	metadataLen, bodyLen, err := s.rb.Serialize(s.w, arrow, batch.Length())
	if err != nil {
		__antithesis_instrumentation__.Notify(55625)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55626)
	}
	__antithesis_instrumentation__.Notify(55622)

	s.recordBatches = append(s.recordBatches, fileBlock{
		offset:      offset,
		metadataLen: int32(metadataLen),
		bodyLen:     int64(bodyLen),
	})
	return nil
}

func (s *FileSerializer) Finish() error {
	__antithesis_instrumentation__.Notify(55627)
	defer func() {
		__antithesis_instrumentation__.Notify(55631)
		s.w = nil
	}()
	__antithesis_instrumentation__.Notify(55628)

	s.fb.Reset()
	footerOffset := fileFooter(s.fb, s.typs, s.recordBatches)
	s.fb.Finish(footerOffset)
	footerBytes := s.fb.FinishedBytes()
	if _, err := s.w.Write(footerBytes); err != nil {
		__antithesis_instrumentation__.Notify(55632)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55633)
	}
	__antithesis_instrumentation__.Notify(55629)

	binary.LittleEndian.PutUint32(s.scratch[:], uint32(len(footerBytes)))
	if _, err := s.w.Write(s.scratch[:]); err != nil {
		__antithesis_instrumentation__.Notify(55634)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55635)
	}
	__antithesis_instrumentation__.Notify(55630)

	_, err := io.WriteString(s.w, fileMagic)
	return err
}

type FileDeserializer struct {
	buf        []byte
	bufCloseFn func() error

	recordBatches []fileBlock

	idx  int
	end  int
	typs []*types.T
	a    *ArrowBatchConverter
	rb   *RecordBatchSerializer

	arrowScratch []*array.Data
}

func NewFileDeserializerFromBytes(typs []*types.T, buf []byte) (*FileDeserializer, error) {
	__antithesis_instrumentation__.Notify(55636)
	return newFileDeserializer(typs, buf, func() error { __antithesis_instrumentation__.Notify(55637); return nil })
}

func NewFileDeserializerFromPath(typs []*types.T, path string) (*FileDeserializer, error) {
	__antithesis_instrumentation__.Notify(55638)
	f, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(55641)
		return nil, pgerror.Wrapf(err, pgcode.Io, `opening %s`, path)
	} else {
		__antithesis_instrumentation__.Notify(55642)
	}
	__antithesis_instrumentation__.Notify(55639)

	buf, err := mmap.Map(f, mmap.COPY, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(55643)
		return nil, pgerror.Wrapf(err, pgcode.Io, `mmaping %s`, path)
	} else {
		__antithesis_instrumentation__.Notify(55644)
	}
	__antithesis_instrumentation__.Notify(55640)
	return newFileDeserializer(typs, buf, buf.Unmap)
}

func newFileDeserializer(
	typs []*types.T, buf []byte, bufCloseFn func() error,
) (*FileDeserializer, error) {
	__antithesis_instrumentation__.Notify(55645)
	d := &FileDeserializer{
		buf:        buf,
		bufCloseFn: bufCloseFn,
		end:        len(buf),
	}
	var err error
	if err = d.init(); err != nil {
		__antithesis_instrumentation__.Notify(55649)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(55650)
	}
	__antithesis_instrumentation__.Notify(55646)
	d.typs = typs

	if d.a, err = NewArrowBatchConverter(typs); err != nil {
		__antithesis_instrumentation__.Notify(55651)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(55652)
	}
	__antithesis_instrumentation__.Notify(55647)
	if d.rb, err = NewRecordBatchSerializer(typs); err != nil {
		__antithesis_instrumentation__.Notify(55653)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(55654)
	}
	__antithesis_instrumentation__.Notify(55648)
	d.arrowScratch = make([]*array.Data, 0, len(typs))

	return d, nil
}

func (d *FileDeserializer) Close() error {
	__antithesis_instrumentation__.Notify(55655)
	return d.bufCloseFn()
}

func (d *FileDeserializer) Typs() []*types.T {
	__antithesis_instrumentation__.Notify(55656)
	return d.typs
}

func (d *FileDeserializer) NumBatches() int {
	__antithesis_instrumentation__.Notify(55657)
	return len(d.recordBatches)
}

func (d *FileDeserializer) GetBatch(batchIdx int, b coldata.Batch) error {
	__antithesis_instrumentation__.Notify(55658)
	rb := d.recordBatches[batchIdx]
	d.idx = int(rb.offset)
	buf, err := d.read(metadataLengthNumBytes + int(rb.metadataLen) + int(rb.bodyLen))
	if err != nil {
		__antithesis_instrumentation__.Notify(55661)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55662)
	}
	__antithesis_instrumentation__.Notify(55659)
	d.arrowScratch = d.arrowScratch[:0]
	batchLength, err := d.rb.Deserialize(&d.arrowScratch, buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(55663)
		return err
	} else {
		__antithesis_instrumentation__.Notify(55664)
	}
	__antithesis_instrumentation__.Notify(55660)
	return d.a.ArrowToBatch(d.arrowScratch, batchLength, b)
}

func (d *FileDeserializer) read(n int) ([]byte, error) {
	__antithesis_instrumentation__.Notify(55665)
	if d.idx+n > d.end {
		__antithesis_instrumentation__.Notify(55667)
		return nil, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(55668)
	}
	__antithesis_instrumentation__.Notify(55666)
	start := d.idx
	d.idx += n
	return d.buf[start:d.idx], nil
}

func (d *FileDeserializer) readBackward(n int) ([]byte, error) {
	__antithesis_instrumentation__.Notify(55669)
	if d.idx+n > d.end {
		__antithesis_instrumentation__.Notify(55671)
		return nil, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(55672)
	}
	__antithesis_instrumentation__.Notify(55670)
	end := d.end
	d.end -= n
	return d.buf[d.end:end], nil
}

func (d *FileDeserializer) init() error {

	if magic, err := d.read(8); err != nil {
		return pgerror.Wrap(err, pgcode.DataException, `verifying arrow file header magic`)
	} else if !bytes.Equal([]byte(fileMagic), magic[:len(fileMagic)]) {
		return errors.New(`arrow file header magic mismatch`)
	}
	if magic, err := d.readBackward(len(fileMagic)); err != nil {
		return pgerror.Wrap(err, pgcode.DataException, `verifying arrow file footer magic`)
	} else if !bytes.Equal([]byte(fileMagic), magic) {
		return errors.New(`arrow file magic footer mismatch`)
	}

	footerSize, err := d.readBackward(4)
	if err != nil {
		return pgerror.Wrap(err, pgcode.DataException, `reading arrow file footer`)
	}
	footerBytes, err := d.readBackward(int(binary.LittleEndian.Uint32(footerSize)))
	if err != nil {
		return pgerror.Wrap(err, pgcode.DataException, `reading arrow file footer`)
	}
	footer := arrowserde.GetRootAsFooter(footerBytes, 0)
	if footer.Version() != arrowserde.MetadataVersionV1 {
		return errors.Errorf(`only arrow V1 is supported got %d`, footer.Version())
	}

	var block arrowserde.Block
	d.recordBatches = d.recordBatches[:0]
	for blockIdx := 0; blockIdx < footer.RecordBatchesLength(); blockIdx++ {
		footer.RecordBatches(&block, blockIdx)
		d.recordBatches = append(d.recordBatches, fileBlock{
			offset:      block.Offset(),
			metadataLen: block.MetaDataLength(),
			bodyLen:     block.BodyLength(),
		})
	}

	return nil
}

type countingWriter struct {
	wrapped io.Writer
	written int
}

func (w *countingWriter) Write(buf []byte) (int, error) {
	__antithesis_instrumentation__.Notify(55673)
	n, err := w.wrapped.Write(buf)
	w.written += n
	return n, err
}

func schema(fb *flatbuffers.Builder, typs []*types.T) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55674)
	fieldOffsets := make([]flatbuffers.UOffsetT, len(typs))
	for idx, typ := range typs {
		__antithesis_instrumentation__.Notify(55677)
		var fbTyp byte
		var fbTypOffset flatbuffers.UOffsetT
		switch typeconv.TypeFamilyToCanonicalTypeFamily(typ.Family()) {
		case types.BoolFamily:
			__antithesis_instrumentation__.Notify(55679)
			arrowserde.BoolStart(fb)
			fbTypOffset = arrowserde.BoolEnd(fb)
			fbTyp = arrowserde.TypeBool
		case types.BytesFamily, types.JsonFamily:
			__antithesis_instrumentation__.Notify(55680)
			arrowserde.BinaryStart(fb)
			fbTypOffset = arrowserde.BinaryEnd(fb)
			fbTyp = arrowserde.TypeBinary
		case types.IntFamily:
			__antithesis_instrumentation__.Notify(55681)
			switch typ.Width() {
			case 16:
				__antithesis_instrumentation__.Notify(55688)
				arrowserde.IntStart(fb)
				arrowserde.IntAddBitWidth(fb, 16)
				arrowserde.IntAddIsSigned(fb, 1)
				fbTypOffset = arrowserde.IntEnd(fb)
				fbTyp = arrowserde.TypeInt
			case 32:
				__antithesis_instrumentation__.Notify(55689)
				arrowserde.IntStart(fb)
				arrowserde.IntAddBitWidth(fb, 32)
				arrowserde.IntAddIsSigned(fb, 1)
				fbTypOffset = arrowserde.IntEnd(fb)
				fbTyp = arrowserde.TypeInt
			case 0, 64:
				__antithesis_instrumentation__.Notify(55690)
				arrowserde.IntStart(fb)
				arrowserde.IntAddBitWidth(fb, 64)
				arrowserde.IntAddIsSigned(fb, 1)
				fbTypOffset = arrowserde.IntEnd(fb)
				fbTyp = arrowserde.TypeInt
			default:
				__antithesis_instrumentation__.Notify(55691)
				panic(errors.Errorf(`unexpected int width %d`, typ.Width()))
			}
		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(55682)
			arrowserde.FloatingPointStart(fb)
			arrowserde.FloatingPointAddPrecision(fb, arrowserde.PrecisionDOUBLE)
			fbTypOffset = arrowserde.FloatingPointEnd(fb)
			fbTyp = arrowserde.TypeFloatingPoint
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(55683)

			arrowserde.BinaryStart(fb)
			fbTypOffset = arrowserde.BinaryEnd(fb)
			fbTyp = arrowserde.TypeDecimal
		case types.TimestampTZFamily:
			__antithesis_instrumentation__.Notify(55684)

			arrowserde.BinaryStart(fb)
			fbTypOffset = arrowserde.BinaryEnd(fb)
			fbTyp = arrowserde.TypeTimestamp
		case types.IntervalFamily:
			__antithesis_instrumentation__.Notify(55685)

			arrowserde.BinaryStart(fb)
			fbTypOffset = arrowserde.BinaryEnd(fb)
			fbTyp = arrowserde.TypeInterval
		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(55686)

			arrowserde.BinaryStart(fb)
			fbTypOffset = arrowserde.BinaryEnd(fb)
			fbTyp = arrowserde.TypeUtf8
		default:
			__antithesis_instrumentation__.Notify(55687)
			panic(errors.Errorf(`don't know how to map %s`, typ))
		}
		__antithesis_instrumentation__.Notify(55678)
		arrowserde.FieldStart(fb)
		arrowserde.FieldAddTypeType(fb, fbTyp)
		arrowserde.FieldAddType(fb, fbTypOffset)
		fieldOffsets[idx] = arrowserde.FieldEnd(fb)
	}
	__antithesis_instrumentation__.Notify(55675)

	arrowserde.SchemaStartFieldsVector(fb, len(typs))

	for i := len(fieldOffsets) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(55692)
		fb.PrependUOffsetT(fieldOffsets[i])
	}
	__antithesis_instrumentation__.Notify(55676)
	fields := fb.EndVector(len(typs))

	arrowserde.SchemaStart(fb)
	arrowserde.SchemaAddFields(fb, fields)
	return arrowserde.SchemaEnd(fb)
}

func schemaMessage(fb *flatbuffers.Builder, typs []*types.T) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55693)
	schemaOffset := schema(fb, typs)
	arrowserde.MessageStart(fb)
	arrowserde.MessageAddVersion(fb, arrowserde.MetadataVersionV1)
	arrowserde.MessageAddHeaderType(fb, arrowserde.MessageHeaderSchema)
	arrowserde.MessageAddHeader(fb, schemaOffset)
	return arrowserde.MessageEnd(fb)
}

func fileFooter(
	fb *flatbuffers.Builder, typs []*types.T, recordBatches []fileBlock,
) flatbuffers.UOffsetT {
	__antithesis_instrumentation__.Notify(55694)
	schemaOffset := schema(fb, typs)
	arrowserde.FooterStartRecordBatchesVector(fb, len(recordBatches))

	for i := len(recordBatches) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(55696)
		rb := recordBatches[i]
		arrowserde.CreateBlock(fb, rb.offset, rb.metadataLen, rb.bodyLen)
	}
	__antithesis_instrumentation__.Notify(55695)
	recordBatchesOffset := fb.EndVector(len(recordBatches))
	arrowserde.FooterStart(fb)
	arrowserde.FooterAddVersion(fb, arrowserde.MetadataVersionV1)
	arrowserde.FooterAddSchema(fb, schemaOffset)
	arrowserde.FooterAddRecordBatches(fb, recordBatchesOffset)
	return arrowserde.FooterEnd(fb)
}
