package rowenc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func EncodingDirToDatumEncoding(dir encoding.Direction) descpb.DatumEncoding {
	__antithesis_instrumentation__.Notify(569908)
	switch dir {
	case encoding.Ascending:
		__antithesis_instrumentation__.Notify(569909)
		return descpb.DatumEncoding_ASCENDING_KEY
	case encoding.Descending:
		__antithesis_instrumentation__.Notify(569910)
		return descpb.DatumEncoding_DESCENDING_KEY
	default:
		__antithesis_instrumentation__.Notify(569911)
		panic(errors.AssertionFailedf("invalid encoding direction: %d", dir))
	}
}

type EncDatum struct {
	encoding descpb.DatumEncoding

	encoded []byte

	Datum tree.Datum
}

func (ed *EncDatum) stringWithAlloc(typ *types.T, a *tree.DatumAlloc) string {
	__antithesis_instrumentation__.Notify(569912)
	if ed.Datum == nil {
		__antithesis_instrumentation__.Notify(569914)
		if ed.encoded == nil {
			__antithesis_instrumentation__.Notify(569917)
			return "<unset>"
		} else {
			__antithesis_instrumentation__.Notify(569918)
		}
		__antithesis_instrumentation__.Notify(569915)
		if a == nil {
			__antithesis_instrumentation__.Notify(569919)
			a = &tree.DatumAlloc{}
		} else {
			__antithesis_instrumentation__.Notify(569920)
		}
		__antithesis_instrumentation__.Notify(569916)
		err := ed.EnsureDecoded(typ, a)
		if err != nil {
			__antithesis_instrumentation__.Notify(569921)
			return fmt.Sprintf("<error: %v>", err)
		} else {
			__antithesis_instrumentation__.Notify(569922)
		}
	} else {
		__antithesis_instrumentation__.Notify(569923)
	}
	__antithesis_instrumentation__.Notify(569913)
	return ed.Datum.String()
}

func (ed *EncDatum) String(typ *types.T) string {
	__antithesis_instrumentation__.Notify(569924)
	return ed.stringWithAlloc(typ, nil)
}

func (ed *EncDatum) BytesEqual(b []byte) bool {
	__antithesis_instrumentation__.Notify(569925)
	return bytes.Equal(ed.encoded, b)
}

func (ed *EncDatum) EncodedString() string {
	__antithesis_instrumentation__.Notify(569926)
	return string(ed.encoded)
}

func (ed *EncDatum) EncodedBytes() []byte {
	__antithesis_instrumentation__.Notify(569927)
	return ed.encoded
}

const EncDatumOverhead = unsafe.Sizeof(EncDatum{})

func (ed EncDatum) Size() uintptr {
	__antithesis_instrumentation__.Notify(569928)
	size := EncDatumOverhead
	if ed.encoded != nil {
		__antithesis_instrumentation__.Notify(569931)
		size += uintptr(len(ed.encoded))
	} else {
		__antithesis_instrumentation__.Notify(569932)
	}
	__antithesis_instrumentation__.Notify(569929)
	if ed.Datum != nil {
		__antithesis_instrumentation__.Notify(569933)
		size += ed.Datum.Size()
	} else {
		__antithesis_instrumentation__.Notify(569934)
	}
	__antithesis_instrumentation__.Notify(569930)
	return size
}

func (ed EncDatum) DiskSize() uintptr {
	__antithesis_instrumentation__.Notify(569935)
	if ed.encoded != nil {
		__antithesis_instrumentation__.Notify(569938)
		return uintptr(len(ed.encoded))
	} else {
		__antithesis_instrumentation__.Notify(569939)
	}
	__antithesis_instrumentation__.Notify(569936)
	if ed.Datum != nil {
		__antithesis_instrumentation__.Notify(569940)
		return ed.Datum.Size()
	} else {
		__antithesis_instrumentation__.Notify(569941)
	}
	__antithesis_instrumentation__.Notify(569937)
	return uintptr(0)
}

func EncDatumFromEncoded(enc descpb.DatumEncoding, encoded []byte) EncDatum {
	__antithesis_instrumentation__.Notify(569942)
	if len(encoded) == 0 {
		__antithesis_instrumentation__.Notify(569944)
		panic(errors.AssertionFailedf("empty encoded value"))
	} else {
		__antithesis_instrumentation__.Notify(569945)
	}
	__antithesis_instrumentation__.Notify(569943)
	return EncDatum{
		encoding: enc,
		encoded:  encoded,
		Datum:    nil,
	}
}

func EncDatumFromBuffer(
	typ *types.T, enc descpb.DatumEncoding, buf []byte,
) (EncDatum, []byte, error) {
	__antithesis_instrumentation__.Notify(569946)
	if len(buf) == 0 {
		__antithesis_instrumentation__.Notify(569948)
		return EncDatum{}, nil, errors.New("empty encoded value")
	} else {
		__antithesis_instrumentation__.Notify(569949)
	}
	__antithesis_instrumentation__.Notify(569947)
	switch enc {
	case descpb.DatumEncoding_ASCENDING_KEY, descpb.DatumEncoding_DESCENDING_KEY:
		__antithesis_instrumentation__.Notify(569950)
		var encLen int
		var err error
		encLen, err = encoding.PeekLength(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(569955)
			return EncDatum{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(569956)
		}
		__antithesis_instrumentation__.Notify(569951)
		ed := EncDatumFromEncoded(enc, buf[:encLen])
		return ed, buf[encLen:], nil
	case descpb.DatumEncoding_VALUE:
		__antithesis_instrumentation__.Notify(569952)
		typeOffset, encLen, err := encoding.PeekValueLength(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(569957)
			return EncDatum{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(569958)
		}
		__antithesis_instrumentation__.Notify(569953)
		ed := EncDatumFromEncoded(enc, buf[typeOffset:encLen])
		return ed, buf[encLen:], nil
	default:
		__antithesis_instrumentation__.Notify(569954)
		panic(errors.AssertionFailedf("unknown encoding %s", enc))
	}
}

func EncDatumValueFromBufferWithOffsetsAndType(
	buf []byte, typeOffset int, dataOffset int, typ encoding.Type,
) (EncDatum, []byte, error) {
	__antithesis_instrumentation__.Notify(569959)
	encLen, err := encoding.PeekValueLengthWithOffsetsAndType(buf, dataOffset, typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(569961)
		return EncDatum{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(569962)
	}
	__antithesis_instrumentation__.Notify(569960)
	ed := EncDatumFromEncoded(descpb.DatumEncoding_VALUE, buf[typeOffset:encLen])
	return ed, buf[encLen:], nil
}

func DatumToEncDatum(ctyp *types.T, d tree.Datum) EncDatum {
	__antithesis_instrumentation__.Notify(569963)
	if d == nil {
		__antithesis_instrumentation__.Notify(569966)
		panic(errors.AssertionFailedf("cannot convert nil datum to EncDatum"))
	} else {
		__antithesis_instrumentation__.Notify(569967)
	}
	__antithesis_instrumentation__.Notify(569964)

	dTyp := d.ResolvedType()
	if d != tree.DNull && func() bool {
		__antithesis_instrumentation__.Notify(569968)
		return !ctyp.Equivalent(dTyp) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(569969)
		return !dTyp.IsAmbiguous() == true
	}() == true {
		__antithesis_instrumentation__.Notify(569970)
		panic(errors.AssertionFailedf("invalid datum type given: %s, expected %s", dTyp, ctyp))
	} else {
		__antithesis_instrumentation__.Notify(569971)
	}
	__antithesis_instrumentation__.Notify(569965)
	return EncDatum{Datum: d}
}

func (ed *EncDatum) UnsetDatum() {
	__antithesis_instrumentation__.Notify(569972)
	ed.encoded = nil
	ed.Datum = nil
	ed.encoding = 0
}

func (ed *EncDatum) IsUnset() bool {
	__antithesis_instrumentation__.Notify(569973)
	return ed.encoded == nil && func() bool {
		__antithesis_instrumentation__.Notify(569974)
		return ed.Datum == nil == true
	}() == true
}

func (ed *EncDatum) IsNull() bool {
	__antithesis_instrumentation__.Notify(569975)
	if ed.Datum != nil {
		__antithesis_instrumentation__.Notify(569978)
		return ed.Datum == tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(569979)
	}
	__antithesis_instrumentation__.Notify(569976)
	if ed.encoded == nil {
		__antithesis_instrumentation__.Notify(569980)
		panic(errors.AssertionFailedf("IsNull on unset EncDatum"))
	} else {
		__antithesis_instrumentation__.Notify(569981)
	}
	__antithesis_instrumentation__.Notify(569977)
	switch ed.encoding {
	case descpb.DatumEncoding_ASCENDING_KEY, descpb.DatumEncoding_DESCENDING_KEY:
		__antithesis_instrumentation__.Notify(569982)
		_, isNull := encoding.DecodeIfNull(ed.encoded)
		return isNull

	case descpb.DatumEncoding_VALUE:
		__antithesis_instrumentation__.Notify(569983)
		_, _, _, typ, err := encoding.DecodeValueTag(ed.encoded)
		if err != nil {
			__antithesis_instrumentation__.Notify(569986)
			panic(errors.WithAssertionFailure(err))
		} else {
			__antithesis_instrumentation__.Notify(569987)
		}
		__antithesis_instrumentation__.Notify(569984)
		return typ == encoding.Null

	default:
		__antithesis_instrumentation__.Notify(569985)
		panic(errors.AssertionFailedf("unknown encoding %s", ed.encoding))
	}
}

func (ed *EncDatum) EnsureDecoded(typ *types.T, a *tree.DatumAlloc) error {
	__antithesis_instrumentation__.Notify(569988)
	if ed.Datum != nil {
		__antithesis_instrumentation__.Notify(569994)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(569995)
	}
	__antithesis_instrumentation__.Notify(569989)
	if ed.encoded == nil {
		__antithesis_instrumentation__.Notify(569996)
		return errors.AssertionFailedf("decoding unset EncDatum")
	} else {
		__antithesis_instrumentation__.Notify(569997)
	}
	__antithesis_instrumentation__.Notify(569990)
	var err error
	var rem []byte
	switch ed.encoding {
	case descpb.DatumEncoding_ASCENDING_KEY:
		__antithesis_instrumentation__.Notify(569998)
		ed.Datum, rem, err = keyside.Decode(a, typ, ed.encoded, encoding.Ascending)
	case descpb.DatumEncoding_DESCENDING_KEY:
		__antithesis_instrumentation__.Notify(569999)
		ed.Datum, rem, err = keyside.Decode(a, typ, ed.encoded, encoding.Descending)
	case descpb.DatumEncoding_VALUE:
		__antithesis_instrumentation__.Notify(570000)
		ed.Datum, rem, err = valueside.Decode(a, typ, ed.encoded)
	default:
		__antithesis_instrumentation__.Notify(570001)
		return errors.AssertionFailedf("unknown encoding %d", redact.Safe(ed.encoding))
	}
	__antithesis_instrumentation__.Notify(569991)
	if err != nil {
		__antithesis_instrumentation__.Notify(570002)
		return errors.Wrapf(err, "error decoding %d bytes", redact.Safe(len(ed.encoded)))
	} else {
		__antithesis_instrumentation__.Notify(570003)
	}
	__antithesis_instrumentation__.Notify(569992)
	if len(rem) != 0 {
		__antithesis_instrumentation__.Notify(570004)
		ed.Datum = nil
		return errors.AssertionFailedf(
			"%d trailing bytes in encoded value: %+v", redact.Safe(len(rem)), rem)
	} else {
		__antithesis_instrumentation__.Notify(570005)
	}
	__antithesis_instrumentation__.Notify(569993)
	return nil
}

func (ed *EncDatum) Encoding() (descpb.DatumEncoding, bool) {
	__antithesis_instrumentation__.Notify(570006)
	if ed.encoded == nil {
		__antithesis_instrumentation__.Notify(570008)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(570009)
	}
	__antithesis_instrumentation__.Notify(570007)
	return ed.encoding, true
}

func (ed *EncDatum) Encode(
	typ *types.T, a *tree.DatumAlloc, enc descpb.DatumEncoding, appendTo []byte,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(570010)
	if ed.encoded != nil && func() bool {
		__antithesis_instrumentation__.Notify(570013)
		return enc == ed.encoding == true
	}() == true {
		__antithesis_instrumentation__.Notify(570014)

		return append(appendTo, ed.encoded...), nil
	} else {
		__antithesis_instrumentation__.Notify(570015)
	}
	__antithesis_instrumentation__.Notify(570011)
	if err := ed.EnsureDecoded(typ, a); err != nil {
		__antithesis_instrumentation__.Notify(570016)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570017)
	}
	__antithesis_instrumentation__.Notify(570012)
	switch enc {
	case descpb.DatumEncoding_ASCENDING_KEY:
		__antithesis_instrumentation__.Notify(570018)
		return keyside.Encode(appendTo, ed.Datum, encoding.Ascending)
	case descpb.DatumEncoding_DESCENDING_KEY:
		__antithesis_instrumentation__.Notify(570019)
		return keyside.Encode(appendTo, ed.Datum, encoding.Descending)
	case descpb.DatumEncoding_VALUE:
		__antithesis_instrumentation__.Notify(570020)
		return valueside.Encode(appendTo, valueside.NoColumnID, ed.Datum, nil)
	default:
		__antithesis_instrumentation__.Notify(570021)
		panic(errors.AssertionFailedf("unknown encoding requested %s", enc))
	}
}

func (ed *EncDatum) Fingerprint(
	ctx context.Context, typ *types.T, a *tree.DatumAlloc, appendTo []byte, acc *mon.BoundAccount,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(570022)

	var fingerprint []byte
	var err error
	memUsageBefore := ed.Size()
	switch typ.Family() {
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(570025)
		if err = ed.EnsureDecoded(typ, a); err != nil {
			__antithesis_instrumentation__.Notify(570028)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(570029)
		}
		__antithesis_instrumentation__.Notify(570026)

		fingerprint, err = valueside.Encode(appendTo, valueside.NoColumnID, ed.Datum, nil)
	default:
		__antithesis_instrumentation__.Notify(570027)

		fingerprint, err = ed.Encode(typ, a, descpb.DatumEncoding_ASCENDING_KEY, appendTo)
	}
	__antithesis_instrumentation__.Notify(570023)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(570030)
		return acc != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(570031)
		return fingerprint, acc.Grow(ctx, int64(ed.Size()-memUsageBefore))
	} else {
		__antithesis_instrumentation__.Notify(570032)
	}
	__antithesis_instrumentation__.Notify(570024)
	return fingerprint, err
}

func (ed *EncDatum) Compare(
	typ *types.T, a *tree.DatumAlloc, evalCtx *tree.EvalContext, rhs *EncDatum,
) (int, error) {
	__antithesis_instrumentation__.Notify(570033)

	if ed.encoding == rhs.encoding && func() bool {
		__antithesis_instrumentation__.Notify(570037)
		return ed.encoded != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(570038)
		return rhs.encoded != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(570039)
		switch ed.encoding {
		case descpb.DatumEncoding_ASCENDING_KEY:
			__antithesis_instrumentation__.Notify(570040)
			return bytes.Compare(ed.encoded, rhs.encoded), nil
		case descpb.DatumEncoding_DESCENDING_KEY:
			__antithesis_instrumentation__.Notify(570041)
			return bytes.Compare(rhs.encoded, ed.encoded), nil
		default:
			__antithesis_instrumentation__.Notify(570042)
		}
	} else {
		__antithesis_instrumentation__.Notify(570043)
	}
	__antithesis_instrumentation__.Notify(570034)
	if err := ed.EnsureDecoded(typ, a); err != nil {
		__antithesis_instrumentation__.Notify(570044)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(570045)
	}
	__antithesis_instrumentation__.Notify(570035)
	if err := rhs.EnsureDecoded(typ, a); err != nil {
		__antithesis_instrumentation__.Notify(570046)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(570047)
	}
	__antithesis_instrumentation__.Notify(570036)
	return ed.Datum.Compare(evalCtx, rhs.Datum), nil
}

func (ed *EncDatum) GetInt() (int64, error) {
	__antithesis_instrumentation__.Notify(570048)
	if ed.Datum != nil {
		__antithesis_instrumentation__.Notify(570050)
		if ed.Datum == tree.DNull {
			__antithesis_instrumentation__.Notify(570052)
			return 0, errors.Errorf("NULL INT value")
		} else {
			__antithesis_instrumentation__.Notify(570053)
		}
		__antithesis_instrumentation__.Notify(570051)
		return int64(*ed.Datum.(*tree.DInt)), nil
	} else {
		__antithesis_instrumentation__.Notify(570054)
	}
	__antithesis_instrumentation__.Notify(570049)

	switch ed.encoding {
	case descpb.DatumEncoding_ASCENDING_KEY:
		__antithesis_instrumentation__.Notify(570055)
		if _, isNull := encoding.DecodeIfNull(ed.encoded); isNull {
			__antithesis_instrumentation__.Notify(570063)
			return 0, errors.Errorf("NULL INT value")
		} else {
			__antithesis_instrumentation__.Notify(570064)
		}
		__antithesis_instrumentation__.Notify(570056)
		_, val, err := encoding.DecodeVarintAscending(ed.encoded)
		return val, err

	case descpb.DatumEncoding_DESCENDING_KEY:
		__antithesis_instrumentation__.Notify(570057)
		if _, isNull := encoding.DecodeIfNull(ed.encoded); isNull {
			__antithesis_instrumentation__.Notify(570065)
			return 0, errors.Errorf("NULL INT value")
		} else {
			__antithesis_instrumentation__.Notify(570066)
		}
		__antithesis_instrumentation__.Notify(570058)
		_, val, err := encoding.DecodeVarintDescending(ed.encoded)
		return val, err

	case descpb.DatumEncoding_VALUE:
		__antithesis_instrumentation__.Notify(570059)
		_, dataOffset, _, typ, err := encoding.DecodeValueTag(ed.encoded)
		if err != nil {
			__antithesis_instrumentation__.Notify(570067)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(570068)
		}
		__antithesis_instrumentation__.Notify(570060)

		if typ == encoding.Null {
			__antithesis_instrumentation__.Notify(570069)
			return 0, errors.Errorf("NULL INT value")
		} else {
			__antithesis_instrumentation__.Notify(570070)
		}
		__antithesis_instrumentation__.Notify(570061)

		_, val, err := encoding.DecodeUntaggedIntValue(ed.encoded[dataOffset:])
		return val, err

	default:
		__antithesis_instrumentation__.Notify(570062)
		return 0, errors.Errorf("unknown encoding %s", ed.encoding)
	}
}

type EncDatumRow []EncDatum

func (r EncDatumRow) stringToBuf(types []*types.T, a *tree.DatumAlloc, b *bytes.Buffer) {
	__antithesis_instrumentation__.Notify(570071)
	if len(types) != len(r) {
		__antithesis_instrumentation__.Notify(570074)
		panic(errors.AssertionFailedf("mismatched types (%v) and row (%v)", types, r))
	} else {
		__antithesis_instrumentation__.Notify(570075)
	}
	__antithesis_instrumentation__.Notify(570072)
	b.WriteString("[")
	for i := range r {
		__antithesis_instrumentation__.Notify(570076)
		if i > 0 {
			__antithesis_instrumentation__.Notify(570078)
			b.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(570079)
		}
		__antithesis_instrumentation__.Notify(570077)
		b.WriteString(r[i].stringWithAlloc(types[i], a))
	}
	__antithesis_instrumentation__.Notify(570073)
	b.WriteString("]")
}

func (r EncDatumRow) Copy() EncDatumRow {
	__antithesis_instrumentation__.Notify(570080)
	if r == nil {
		__antithesis_instrumentation__.Notify(570082)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(570083)
	}
	__antithesis_instrumentation__.Notify(570081)
	rCopy := make(EncDatumRow, len(r))
	copy(rCopy, r)
	return rCopy
}

func (r EncDatumRow) String(types []*types.T) string {
	__antithesis_instrumentation__.Notify(570084)
	var b bytes.Buffer
	r.stringToBuf(types, &tree.DatumAlloc{}, &b)
	return b.String()
}

const EncDatumRowOverhead = unsafe.Sizeof(EncDatumRow{})

func (r EncDatumRow) Size() uintptr {
	__antithesis_instrumentation__.Notify(570085)
	size := EncDatumRowOverhead
	for _, ed := range r {
		__antithesis_instrumentation__.Notify(570087)
		size += ed.Size()
	}
	__antithesis_instrumentation__.Notify(570086)
	return size
}

func EncDatumRowToDatums(
	types []*types.T, datums tree.Datums, row EncDatumRow, da *tree.DatumAlloc,
) error {
	__antithesis_instrumentation__.Notify(570088)
	if len(types) != len(row) {
		__antithesis_instrumentation__.Notify(570092)
		return errors.AssertionFailedf(
			"mismatched types (%v) and row (%v)", types, row)
	} else {
		__antithesis_instrumentation__.Notify(570093)
	}
	__antithesis_instrumentation__.Notify(570089)
	if len(row) != len(datums) {
		__antithesis_instrumentation__.Notify(570094)
		return errors.AssertionFailedf(
			"Length mismatch (%d and %d) between datums and row", len(datums), len(row))
	} else {
		__antithesis_instrumentation__.Notify(570095)
	}
	__antithesis_instrumentation__.Notify(570090)
	for i, encDatum := range row {
		__antithesis_instrumentation__.Notify(570096)
		if encDatum.IsUnset() {
			__antithesis_instrumentation__.Notify(570099)
			datums[i] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(570100)
		}
		__antithesis_instrumentation__.Notify(570097)
		err := encDatum.EnsureDecoded(types[i], da)
		if err != nil {
			__antithesis_instrumentation__.Notify(570101)
			return err
		} else {
			__antithesis_instrumentation__.Notify(570102)
		}
		__antithesis_instrumentation__.Notify(570098)
		datums[i] = encDatum.Datum
	}
	__antithesis_instrumentation__.Notify(570091)
	return nil
}

func (r EncDatumRow) Compare(
	types []*types.T,
	a *tree.DatumAlloc,
	ordering colinfo.ColumnOrdering,
	evalCtx *tree.EvalContext,
	rhs EncDatumRow,
) (int, error) {
	__antithesis_instrumentation__.Notify(570103)
	if len(r) != len(types) || func() bool {
		__antithesis_instrumentation__.Notify(570106)
		return len(rhs) != len(types) == true
	}() == true {
		__antithesis_instrumentation__.Notify(570107)
		panic(errors.AssertionFailedf("length mismatch: %d types, %d lhs, %d rhs\n%+v\n%+v\n%+v", len(types), len(r), len(rhs), types, r, rhs))
	} else {
		__antithesis_instrumentation__.Notify(570108)
	}
	__antithesis_instrumentation__.Notify(570104)
	for _, c := range ordering {
		__antithesis_instrumentation__.Notify(570109)
		cmp, err := r[c.ColIdx].Compare(types[c.ColIdx], a, evalCtx, &rhs[c.ColIdx])
		if err != nil {
			__antithesis_instrumentation__.Notify(570111)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(570112)
		}
		__antithesis_instrumentation__.Notify(570110)
		if cmp != 0 {
			__antithesis_instrumentation__.Notify(570113)
			if c.Direction == encoding.Descending {
				__antithesis_instrumentation__.Notify(570115)
				cmp = -cmp
			} else {
				__antithesis_instrumentation__.Notify(570116)
			}
			__antithesis_instrumentation__.Notify(570114)
			return cmp, nil
		} else {
			__antithesis_instrumentation__.Notify(570117)
		}
	}
	__antithesis_instrumentation__.Notify(570105)
	return 0, nil
}

func (r EncDatumRow) CompareToDatums(
	types []*types.T,
	a *tree.DatumAlloc,
	ordering colinfo.ColumnOrdering,
	evalCtx *tree.EvalContext,
	rhs tree.Datums,
) (int, error) {
	__antithesis_instrumentation__.Notify(570118)
	for _, c := range ordering {
		__antithesis_instrumentation__.Notify(570120)
		if err := r[c.ColIdx].EnsureDecoded(types[c.ColIdx], a); err != nil {
			__antithesis_instrumentation__.Notify(570122)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(570123)
		}
		__antithesis_instrumentation__.Notify(570121)
		cmp := r[c.ColIdx].Datum.Compare(evalCtx, rhs[c.ColIdx])
		if cmp != 0 {
			__antithesis_instrumentation__.Notify(570124)
			if c.Direction == encoding.Descending {
				__antithesis_instrumentation__.Notify(570126)
				cmp = -cmp
			} else {
				__antithesis_instrumentation__.Notify(570127)
			}
			__antithesis_instrumentation__.Notify(570125)
			return cmp, nil
		} else {
			__antithesis_instrumentation__.Notify(570128)
		}
	}
	__antithesis_instrumentation__.Notify(570119)
	return 0, nil
}

type EncDatumRows []EncDatumRow

func (r EncDatumRows) String(types []*types.T) string {
	__antithesis_instrumentation__.Notify(570129)
	var a tree.DatumAlloc
	var b bytes.Buffer
	b.WriteString("[")
	for i, r := range r {
		__antithesis_instrumentation__.Notify(570131)
		if i > 0 {
			__antithesis_instrumentation__.Notify(570133)
			b.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(570134)
		}
		__antithesis_instrumentation__.Notify(570132)
		r.stringToBuf(types, &a, &b)
	}
	__antithesis_instrumentation__.Notify(570130)
	b.WriteString("]")
	return b.String()
}

type EncDatumRowContainer struct {
	rows  EncDatumRows
	index int
}

func (c *EncDatumRowContainer) Peek() EncDatumRow {
	__antithesis_instrumentation__.Notify(570135)
	return c.rows[c.index]
}

func (c *EncDatumRowContainer) Pop() EncDatumRow {
	__antithesis_instrumentation__.Notify(570136)
	if c.index < 0 {
		__antithesis_instrumentation__.Notify(570138)
		c.index = len(c.rows) - 1
	} else {
		__antithesis_instrumentation__.Notify(570139)
	}
	__antithesis_instrumentation__.Notify(570137)
	row := c.rows[c.index]
	c.index--
	return row
}

func (c *EncDatumRowContainer) Push(row EncDatumRow) {
	__antithesis_instrumentation__.Notify(570140)
	c.rows = append(c.rows, row)
	c.index = len(c.rows) - 1
}

func (c *EncDatumRowContainer) Reset() {
	__antithesis_instrumentation__.Notify(570141)
	c.rows = c.rows[:0]
	c.index = -1
}

func (c *EncDatumRowContainer) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(570142)
	return c.index == -1
}

type EncDatumRowAlloc struct {
	buf []EncDatum

	prealloc [16]EncDatum
}

func (a *EncDatumRowAlloc) AllocRow(cols int) EncDatumRow {
	__antithesis_instrumentation__.Notify(570143)
	if a.buf == nil {
		__antithesis_instrumentation__.Notify(570146)

		a.buf = a.prealloc[:]
	} else {
		__antithesis_instrumentation__.Notify(570147)
	}
	__antithesis_instrumentation__.Notify(570144)
	if len(a.buf) < cols {
		__antithesis_instrumentation__.Notify(570148)

		bufLen := cols
		if cols <= 16 {
			__antithesis_instrumentation__.Notify(570150)
			bufLen *= 16
		} else {
			__antithesis_instrumentation__.Notify(570151)
			if cols <= 64 {
				__antithesis_instrumentation__.Notify(570152)
				bufLen *= 4
			} else {
				__antithesis_instrumentation__.Notify(570153)
			}
		}
		__antithesis_instrumentation__.Notify(570149)
		a.buf = make([]EncDatum, bufLen)
	} else {
		__antithesis_instrumentation__.Notify(570154)
	}
	__antithesis_instrumentation__.Notify(570145)

	result := EncDatumRow(a.buf[:cols:cols])
	a.buf = a.buf[cols:]
	return result
}

func (a *EncDatumRowAlloc) CopyRow(row EncDatumRow) EncDatumRow {
	__antithesis_instrumentation__.Notify(570155)
	rowCopy := a.AllocRow(len(row))
	copy(rowCopy, row)
	return rowCopy
}
