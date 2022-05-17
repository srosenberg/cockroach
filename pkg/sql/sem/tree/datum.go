package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/stringencoding"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timetz"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/lib/pq/oid"
	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

var (
	constDBoolTrue  DBool = true
	constDBoolFalse DBool = false

	DBoolTrue = &constDBoolTrue

	DBoolFalse = &constDBoolFalse

	DNull Datum = dNull{}

	DZero = NewDInt(0)

	DTimeMaxTimeRegex = regexp.MustCompile(`^([0-9-]*(\s|T))?\s*24:00(:00(.0+)?)?\s*$`)

	MaxSupportedTime = timeutil.Unix(9224318016000-1, 999999000)

	MinSupportedTime = timeutil.Unix(-210866803200, 0)
)

type Datum interface {
	TypedExpr

	AmbiguousFormat() bool

	Compare(ctx *EvalContext, other Datum) int

	CompareError(ctx *EvalContext, other Datum) (int, error)

	Prev(ctx *EvalContext) (Datum, bool)

	IsMin(ctx *EvalContext) bool

	Next(ctx *EvalContext) (Datum, bool)

	IsMax(ctx *EvalContext) bool

	Max(ctx *EvalContext) (Datum, bool)

	Min(ctx *EvalContext) (Datum, bool)

	Size() uintptr
}

type Datums []Datum

func (d Datums) Len() int { __antithesis_instrumentation__.Notify(605441); return len(d) }

func (d *Datums) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605442)
	ctx.WriteByte('(')
	for i, v := range *d {
		__antithesis_instrumentation__.Notify(605444)
		if i > 0 {
			__antithesis_instrumentation__.Notify(605446)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(605447)
		}
		__antithesis_instrumentation__.Notify(605445)
		ctx.FormatNode(v)
	}
	__antithesis_instrumentation__.Notify(605443)
	ctx.WriteByte(')')
}

func (d Datums) Compare(evalCtx *EvalContext, other Datums) int {
	__antithesis_instrumentation__.Notify(605448)
	if len(d) == 0 {
		__antithesis_instrumentation__.Notify(605452)
		panic(errors.AssertionFailedf("empty Datums being compared to other"))
	} else {
		__antithesis_instrumentation__.Notify(605453)
	}
	__antithesis_instrumentation__.Notify(605449)

	for i := range d {
		__antithesis_instrumentation__.Notify(605454)
		if i >= len(other) {
			__antithesis_instrumentation__.Notify(605456)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(605457)
		}
		__antithesis_instrumentation__.Notify(605455)

		compareDatum := d[i].Compare(evalCtx, other[i])
		if compareDatum != 0 {
			__antithesis_instrumentation__.Notify(605458)
			return compareDatum
		} else {
			__antithesis_instrumentation__.Notify(605459)
		}
	}
	__antithesis_instrumentation__.Notify(605450)

	if len(d) < len(other) {
		__antithesis_instrumentation__.Notify(605460)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(605461)
	}
	__antithesis_instrumentation__.Notify(605451)
	return 0
}

func (d Datums) IsDistinctFrom(evalCtx *EvalContext, other Datums) bool {
	__antithesis_instrumentation__.Notify(605462)
	if len(d) != len(other) {
		__antithesis_instrumentation__.Notify(605465)
		return true
	} else {
		__antithesis_instrumentation__.Notify(605466)
	}
	__antithesis_instrumentation__.Notify(605463)
	for i, val := range d {
		__antithesis_instrumentation__.Notify(605467)
		if val == DNull {
			__antithesis_instrumentation__.Notify(605468)
			if other[i] != DNull {
				__antithesis_instrumentation__.Notify(605469)
				return true
			} else {
				__antithesis_instrumentation__.Notify(605470)
			}
		} else {
			__antithesis_instrumentation__.Notify(605471)
			if val.Compare(evalCtx, other[i]) != 0 {
				__antithesis_instrumentation__.Notify(605472)
				return true
			} else {
				__antithesis_instrumentation__.Notify(605473)
			}
		}
	}
	__antithesis_instrumentation__.Notify(605464)
	return false
}

type CompositeDatum interface {
	Datum

	IsComposite() bool
}

type DBool bool

func MakeDBool(d DBool) *DBool {
	__antithesis_instrumentation__.Notify(605474)
	if d {
		__antithesis_instrumentation__.Notify(605476)
		return DBoolTrue
	} else {
		__antithesis_instrumentation__.Notify(605477)
	}
	__antithesis_instrumentation__.Notify(605475)
	return DBoolFalse
}

func MustBeDBool(e Expr) DBool {
	__antithesis_instrumentation__.Notify(605478)
	b, ok := AsDBool(e)
	if !ok {
		__antithesis_instrumentation__.Notify(605480)
		panic(errors.AssertionFailedf("expected *DBool, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(605481)
	}
	__antithesis_instrumentation__.Notify(605479)
	return b
}

func AsDBool(e Expr) (DBool, bool) {
	__antithesis_instrumentation__.Notify(605482)
	switch t := e.(type) {
	case *DBool:
		__antithesis_instrumentation__.Notify(605484)
		return *t, true
	}
	__antithesis_instrumentation__.Notify(605483)
	return false, false
}

func MakeParseError(s string, typ *types.T, err error) error {
	__antithesis_instrumentation__.Notify(605485)
	if err != nil {
		__antithesis_instrumentation__.Notify(605487)
		return pgerror.Wrapf(err, pgcode.InvalidTextRepresentation,
			"could not parse %q as type %s", s, typ)
	} else {
		__antithesis_instrumentation__.Notify(605488)
	}
	__antithesis_instrumentation__.Notify(605486)
	return pgerror.Newf(pgcode.InvalidTextRepresentation,
		"could not parse %q as type %s", s, typ)
}

func makeUnsupportedComparisonMessage(d1, d2 Datum) error {
	__antithesis_instrumentation__.Notify(605489)
	return pgerror.Newf(pgcode.DatatypeMismatch,
		"unsupported comparison: %s to %s",
		errors.Safe(d1.ResolvedType()),
		errors.Safe(d2.ResolvedType()),
	)
}

func isCaseInsensitivePrefix(prefix, s string) bool {
	__antithesis_instrumentation__.Notify(605490)
	if len(prefix) > len(s) {
		__antithesis_instrumentation__.Notify(605492)
		return false
	} else {
		__antithesis_instrumentation__.Notify(605493)
	}
	__antithesis_instrumentation__.Notify(605491)
	return strings.EqualFold(prefix, s[:len(prefix)])
}

func ParseBool(s string) (bool, error) {
	__antithesis_instrumentation__.Notify(605494)
	s = strings.TrimSpace(s)
	if len(s) >= 1 {
		__antithesis_instrumentation__.Notify(605496)
		switch s[0] {
		case 't', 'T':
			__antithesis_instrumentation__.Notify(605497)
			if isCaseInsensitivePrefix(s, "true") {
				__antithesis_instrumentation__.Notify(605505)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(605506)
			}
		case 'f', 'F':
			__antithesis_instrumentation__.Notify(605498)
			if isCaseInsensitivePrefix(s, "false") {
				__antithesis_instrumentation__.Notify(605507)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(605508)
			}
		case 'y', 'Y':
			__antithesis_instrumentation__.Notify(605499)
			if isCaseInsensitivePrefix(s, "yes") {
				__antithesis_instrumentation__.Notify(605509)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(605510)
			}
		case 'n', 'N':
			__antithesis_instrumentation__.Notify(605500)
			if isCaseInsensitivePrefix(s, "no") {
				__antithesis_instrumentation__.Notify(605511)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(605512)
			}
		case '1':
			__antithesis_instrumentation__.Notify(605501)
			if s == "1" {
				__antithesis_instrumentation__.Notify(605513)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(605514)
			}
		case '0':
			__antithesis_instrumentation__.Notify(605502)
			if s == "0" {
				__antithesis_instrumentation__.Notify(605515)
				return false, nil
			} else {
				__antithesis_instrumentation__.Notify(605516)
			}
		case 'o', 'O':
			__antithesis_instrumentation__.Notify(605503)

			if len(s) > 1 {
				__antithesis_instrumentation__.Notify(605517)
				if isCaseInsensitivePrefix(s, "on") {
					__antithesis_instrumentation__.Notify(605519)
					return true, nil
				} else {
					__antithesis_instrumentation__.Notify(605520)
				}
				__antithesis_instrumentation__.Notify(605518)
				if isCaseInsensitivePrefix(s, "off") {
					__antithesis_instrumentation__.Notify(605521)
					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(605522)
				}
			} else {
				__antithesis_instrumentation__.Notify(605523)
			}
		default:
			__antithesis_instrumentation__.Notify(605504)
		}
	} else {
		__antithesis_instrumentation__.Notify(605524)
	}
	__antithesis_instrumentation__.Notify(605495)
	return false, MakeParseError(s, types.Bool, pgerror.New(pgcode.InvalidTextRepresentation, "invalid bool value"))
}

func ParseDBool(s string) (*DBool, error) {
	__antithesis_instrumentation__.Notify(605525)
	v, err := ParseBool(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(605528)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(605529)
	}
	__antithesis_instrumentation__.Notify(605526)
	if v {
		__antithesis_instrumentation__.Notify(605530)
		return DBoolTrue, nil
	} else {
		__antithesis_instrumentation__.Notify(605531)
	}
	__antithesis_instrumentation__.Notify(605527)
	return DBoolFalse, nil
}

func ParseDByte(s string) (*DBytes, error) {
	__antithesis_instrumentation__.Notify(605532)
	res, err := lex.DecodeRawBytesToByteArrayAuto([]byte(s))
	if err != nil {
		__antithesis_instrumentation__.Notify(605534)
		return nil, MakeParseError(s, types.Bytes, err)
	} else {
		__antithesis_instrumentation__.Notify(605535)
	}
	__antithesis_instrumentation__.Notify(605533)
	return NewDBytes(DBytes(res)), nil
}

func ParseDUuidFromString(s string) (*DUuid, error) {
	__antithesis_instrumentation__.Notify(605536)
	uv, err := uuid.FromString(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(605538)
		return nil, MakeParseError(s, types.Uuid, err)
	} else {
		__antithesis_instrumentation__.Notify(605539)
	}
	__antithesis_instrumentation__.Notify(605537)
	return NewDUuid(DUuid{uv}), nil
}

func ParseDUuidFromBytes(b []byte) (*DUuid, error) {
	__antithesis_instrumentation__.Notify(605540)
	uv, err := uuid.FromBytes(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(605542)
		return nil, MakeParseError(string(b), types.Uuid, err)
	} else {
		__antithesis_instrumentation__.Notify(605543)
	}
	__antithesis_instrumentation__.Notify(605541)
	return NewDUuid(DUuid{uv}), nil
}

func ParseDIPAddrFromINetString(s string) (*DIPAddr, error) {
	__antithesis_instrumentation__.Notify(605544)
	var d DIPAddr
	err := ipaddr.ParseINet(s, &d.IPAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(605546)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(605547)
	}
	__antithesis_instrumentation__.Notify(605545)
	return &d, nil
}

func GetBool(d Datum) (DBool, error) {
	__antithesis_instrumentation__.Notify(605548)
	if v, ok := d.(*DBool); ok {
		__antithesis_instrumentation__.Notify(605551)
		return *v, nil
	} else {
		__antithesis_instrumentation__.Notify(605552)
	}
	__antithesis_instrumentation__.Notify(605549)
	if d == DNull {
		__antithesis_instrumentation__.Notify(605553)
		return DBool(false), nil
	} else {
		__antithesis_instrumentation__.Notify(605554)
	}
	__antithesis_instrumentation__.Notify(605550)
	return false, errors.AssertionFailedf("cannot convert %s to type %s", d.ResolvedType(), types.Bool)
}

func (*DBool) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605555)
	return types.Bool
}

func (d *DBool) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605556)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605558)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605559)
	}
	__antithesis_instrumentation__.Notify(605557)
	return res
}

func (d *DBool) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605560)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605563)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605564)
	}
	__antithesis_instrumentation__.Notify(605561)
	v, ok := UnwrapDatum(ctx, other).(*DBool)
	if !ok {
		__antithesis_instrumentation__.Notify(605565)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(605566)
	}
	__antithesis_instrumentation__.Notify(605562)
	res := CompareBools(bool(*d), bool(*v))
	return res, nil
}

func CompareBools(d, v bool) int {
	__antithesis_instrumentation__.Notify(605567)
	if !d && func() bool {
		__antithesis_instrumentation__.Notify(605570)
		return v == true
	}() == true {
		__antithesis_instrumentation__.Notify(605571)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(605572)
	}
	__antithesis_instrumentation__.Notify(605568)
	if d && func() bool {
		__antithesis_instrumentation__.Notify(605573)
		return !v == true
	}() == true {
		__antithesis_instrumentation__.Notify(605574)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(605575)
	}
	__antithesis_instrumentation__.Notify(605569)
	return 0
}

func (*DBool) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605576)
	return DBoolFalse, true
}

func (*DBool) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605577)
	return DBoolTrue, true
}

func (d *DBool) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605578)
	return bool(*d)
}

func (d *DBool) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605579)
	return !bool(*d)
}

func (d *DBool) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605580)
	return DBoolFalse, true
}

func (d *DBool) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605581)
	return DBoolTrue, true
}

func (*DBool) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605582); return false }

func PgwireFormatBool(d bool) byte {
	__antithesis_instrumentation__.Notify(605583)
	if d {
		__antithesis_instrumentation__.Notify(605585)
		return 't'
	} else {
		__antithesis_instrumentation__.Notify(605586)
	}
	__antithesis_instrumentation__.Notify(605584)
	return 'f'
}

func (d *DBool) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605587)
	if ctx.HasFlags(fmtPgwireFormat) {
		__antithesis_instrumentation__.Notify(605589)
		ctx.WriteByte(PgwireFormatBool(bool(*d)))
		return
	} else {
		__antithesis_instrumentation__.Notify(605590)
	}
	__antithesis_instrumentation__.Notify(605588)
	ctx.WriteString(strconv.FormatBool(bool(*d)))
}

func (d *DBool) Size() uintptr {
	__antithesis_instrumentation__.Notify(605591)
	return unsafe.Sizeof(*d)
}

type DBitArray struct {
	bitarray.BitArray
}

func ParseDBitArray(s string) (*DBitArray, error) {
	__antithesis_instrumentation__.Notify(605592)
	var a DBitArray
	var err error
	a.BitArray, err = bitarray.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(605594)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(605595)
	}
	__antithesis_instrumentation__.Notify(605593)
	return &a, nil
}

func NewDBitArray(bitLen uint) *DBitArray {
	__antithesis_instrumentation__.Notify(605596)
	a := MakeDBitArray(bitLen)
	return &a
}

func MakeDBitArray(bitLen uint) DBitArray {
	__antithesis_instrumentation__.Notify(605597)
	return DBitArray{BitArray: bitarray.MakeZeroBitArray(bitLen)}
}

func MustBeDBitArray(e Expr) *DBitArray {
	__antithesis_instrumentation__.Notify(605598)
	b, ok := AsDBitArray(e)
	if !ok {
		__antithesis_instrumentation__.Notify(605600)
		panic(errors.AssertionFailedf("expected *DBitArray, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(605601)
	}
	__antithesis_instrumentation__.Notify(605599)
	return b
}

func AsDBitArray(e Expr) (*DBitArray, bool) {
	__antithesis_instrumentation__.Notify(605602)
	switch t := e.(type) {
	case *DBitArray:
		__antithesis_instrumentation__.Notify(605604)
		return t, true
	}
	__antithesis_instrumentation__.Notify(605603)
	return nil, false
}

var errCannotCastNegativeIntToBitArray = pgerror.Newf(pgcode.CannotCoerce,
	"cannot cast negative integer to bit varying with unbounded width")

func NewDBitArrayFromInt(i int64, width uint) (*DBitArray, error) {
	__antithesis_instrumentation__.Notify(605605)
	if width == 0 && func() bool {
		__antithesis_instrumentation__.Notify(605607)
		return i < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(605608)
		return nil, errCannotCastNegativeIntToBitArray
	} else {
		__antithesis_instrumentation__.Notify(605609)
	}
	__antithesis_instrumentation__.Notify(605606)
	return &DBitArray{
		BitArray: bitarray.MakeBitArrayFromInt64(width, i, 64),
	}, nil
}

func (d *DBitArray) AsDInt(n uint) *DInt {
	__antithesis_instrumentation__.Notify(605610)
	if n == 0 {
		__antithesis_instrumentation__.Notify(605612)
		n = 64
	} else {
		__antithesis_instrumentation__.Notify(605613)
	}
	__antithesis_instrumentation__.Notify(605611)
	return NewDInt(DInt(d.BitArray.AsInt64(n)))
}

func (*DBitArray) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605614)
	return types.VarBit
}

func (d *DBitArray) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605615)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605617)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605618)
	}
	__antithesis_instrumentation__.Notify(605616)
	return res
}

func (d *DBitArray) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605619)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605622)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605623)
	}
	__antithesis_instrumentation__.Notify(605620)
	v, ok := UnwrapDatum(ctx, other).(*DBitArray)
	if !ok {
		__antithesis_instrumentation__.Notify(605624)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(605625)
	}
	__antithesis_instrumentation__.Notify(605621)
	res := bitarray.Compare(d.BitArray, v.BitArray)
	return res, nil
}

func (d *DBitArray) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605626)
	return nil, false
}

func (d *DBitArray) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605627)
	a := bitarray.Next(d.BitArray)
	return &DBitArray{BitArray: a}, true
}

func (d *DBitArray) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605628)
	return false
}

func (d *DBitArray) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605629)
	return d.BitArray.IsEmpty()
}

var bitArrayZero = NewDBitArray(0)

func (d *DBitArray) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605630)
	return bitArrayZero, true
}

func (d *DBitArray) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605631)
	return nil, false
}

func (*DBitArray) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605632); return false }

func (d *DBitArray) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605633)
	f := ctx.flags
	if f.HasFlags(fmtPgwireFormat) {
		__antithesis_instrumentation__.Notify(605634)
		d.BitArray.Format(&ctx.Buffer)
	} else {
		__antithesis_instrumentation__.Notify(605635)
		withQuotes := !f.HasFlags(FmtFlags(lexbase.EncBareStrings))
		if withQuotes {
			__antithesis_instrumentation__.Notify(605637)
			ctx.WriteString("B'")
		} else {
			__antithesis_instrumentation__.Notify(605638)
		}
		__antithesis_instrumentation__.Notify(605636)
		d.BitArray.Format(&ctx.Buffer)
		if withQuotes {
			__antithesis_instrumentation__.Notify(605639)
			ctx.WriteByte('\'')
		} else {
			__antithesis_instrumentation__.Notify(605640)
		}
	}
}

func (d *DBitArray) Size() uintptr {
	__antithesis_instrumentation__.Notify(605641)
	return d.BitArray.Sizeof()
}

type DInt int64

func NewDInt(d DInt) *DInt {
	__antithesis_instrumentation__.Notify(605642)
	return &d
}

func ParseDInt(s string) (*DInt, error) {
	__antithesis_instrumentation__.Notify(605643)
	i, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(605645)
		return nil, MakeParseError(s, types.Int, err)
	} else {
		__antithesis_instrumentation__.Notify(605646)
	}
	__antithesis_instrumentation__.Notify(605644)
	return NewDInt(DInt(i)), nil
}

func AsDInt(e Expr) (DInt, bool) {
	__antithesis_instrumentation__.Notify(605647)
	switch t := e.(type) {
	case *DInt:
		__antithesis_instrumentation__.Notify(605649)
		return *t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(605650)
		return AsDInt(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(605648)
	return 0, false
}

func MustBeDInt(e Expr) DInt {
	__antithesis_instrumentation__.Notify(605651)
	i, ok := AsDInt(e)
	if !ok {
		__antithesis_instrumentation__.Notify(605653)
		panic(errors.AssertionFailedf("expected *DInt, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(605654)
	}
	__antithesis_instrumentation__.Notify(605652)
	return i
}

func (*DInt) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605655)
	return types.Int
}

func (d *DInt) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605656)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605658)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605659)
	}
	__antithesis_instrumentation__.Notify(605657)
	return res
}

func (d *DInt) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605660)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605665)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605666)
	}
	__antithesis_instrumentation__.Notify(605661)
	thisInt := *d
	var v DInt
	switch t := UnwrapDatum(ctx, other).(type) {
	case *DInt:
		__antithesis_instrumentation__.Notify(605667)
		v = *t
	case *DFloat, *DDecimal:
		__antithesis_instrumentation__.Notify(605668)
		res, err := t.CompareError(ctx, d)
		if err != nil {
			__antithesis_instrumentation__.Notify(605672)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(605673)
		}
		__antithesis_instrumentation__.Notify(605669)
		return -res, nil
	case *DOid:
		__antithesis_instrumentation__.Notify(605670)

		thisInt = DInt(uint32(thisInt))
		v = t.DInt
	default:
		__antithesis_instrumentation__.Notify(605671)
		return 0, makeUnsupportedComparisonMessage(d, other)
	}
	__antithesis_instrumentation__.Notify(605662)
	if thisInt < v {
		__antithesis_instrumentation__.Notify(605674)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(605675)
	}
	__antithesis_instrumentation__.Notify(605663)
	if thisInt > v {
		__antithesis_instrumentation__.Notify(605676)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605677)
	}
	__antithesis_instrumentation__.Notify(605664)
	return 0, nil
}

func (d *DInt) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605678)
	return NewDInt(*d - 1), true
}

func (d *DInt) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605679)
	return NewDInt(*d + 1), true
}

func (d *DInt) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605680)
	return *d == math.MaxInt64
}

func (d *DInt) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605681)
	return *d == math.MinInt64
}

var dMaxInt = NewDInt(math.MaxInt64)
var dMinInt = NewDInt(math.MinInt64)

func (d *DInt) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605682)
	return dMaxInt, true
}

func (d *DInt) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605683)
	return dMinInt, true
}

func (*DInt) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605684); return true }

func (d *DInt) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605685)

	disambiguate := ctx.flags.HasFlags(fmtDisambiguateDatumTypes)
	parsable := ctx.flags.HasFlags(FmtParsableNumerics)
	needParens := (disambiguate || func() bool {
		__antithesis_instrumentation__.Notify(605687)
		return parsable == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(605688)
		return *d < 0 == true
	}() == true
	if needParens {
		__antithesis_instrumentation__.Notify(605689)
		ctx.WriteByte('(')
	} else {
		__antithesis_instrumentation__.Notify(605690)
	}
	__antithesis_instrumentation__.Notify(605686)
	ctx.WriteString(strconv.FormatInt(int64(*d), 10))
	if needParens {
		__antithesis_instrumentation__.Notify(605691)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(605692)
	}
}

func (d *DInt) Size() uintptr {
	__antithesis_instrumentation__.Notify(605693)
	return unsafe.Sizeof(*d)
}

type DFloat float64

func MustBeDFloat(e Expr) DFloat {
	__antithesis_instrumentation__.Notify(605694)
	switch t := e.(type) {
	case *DFloat:
		__antithesis_instrumentation__.Notify(605696)
		return *t
	}
	__antithesis_instrumentation__.Notify(605695)
	panic(errors.AssertionFailedf("expected *DFloat, found %T", e))
}

func NewDFloat(d DFloat) *DFloat {
	__antithesis_instrumentation__.Notify(605697)
	return &d
}

func ParseDFloat(s string) (*DFloat, error) {
	__antithesis_instrumentation__.Notify(605698)
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(605700)
		return nil, MakeParseError(s, types.Float, err)
	} else {
		__antithesis_instrumentation__.Notify(605701)
	}
	__antithesis_instrumentation__.Notify(605699)
	return NewDFloat(DFloat(f)), nil
}

func (*DFloat) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605702)
	return types.Float
}

func (d *DFloat) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605703)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605705)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605706)
	}
	__antithesis_instrumentation__.Notify(605704)
	return res
}

func (d *DFloat) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605707)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605714)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605715)
	}
	__antithesis_instrumentation__.Notify(605708)
	var v DFloat
	switch t := UnwrapDatum(ctx, other).(type) {
	case *DFloat:
		__antithesis_instrumentation__.Notify(605716)
		v = *t
	case *DInt:
		__antithesis_instrumentation__.Notify(605717)
		v = DFloat(MustBeDInt(t))
	case *DDecimal:
		__antithesis_instrumentation__.Notify(605718)
		res, err := t.CompareError(ctx, d)
		if err != nil {
			__antithesis_instrumentation__.Notify(605721)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(605722)
		}
		__antithesis_instrumentation__.Notify(605719)
		return -res, nil
	default:
		__antithesis_instrumentation__.Notify(605720)
		return 0, makeUnsupportedComparisonMessage(d, other)
	}
	__antithesis_instrumentation__.Notify(605709)
	if *d < v {
		__antithesis_instrumentation__.Notify(605723)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(605724)
	}
	__antithesis_instrumentation__.Notify(605710)
	if *d > v {
		__antithesis_instrumentation__.Notify(605725)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605726)
	}
	__antithesis_instrumentation__.Notify(605711)

	if *d == v {
		__antithesis_instrumentation__.Notify(605727)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(605728)
	}
	__antithesis_instrumentation__.Notify(605712)
	if math.IsNaN(float64(*d)) {
		__antithesis_instrumentation__.Notify(605729)
		if math.IsNaN(float64(v)) {
			__antithesis_instrumentation__.Notify(605731)
			return 0, nil
		} else {
			__antithesis_instrumentation__.Notify(605732)
		}
		__antithesis_instrumentation__.Notify(605730)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(605733)
	}
	__antithesis_instrumentation__.Notify(605713)
	return 1, nil
}

func (d *DFloat) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605734)
	f := float64(*d)
	if math.IsNaN(f) {
		__antithesis_instrumentation__.Notify(605737)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(605738)
	}
	__antithesis_instrumentation__.Notify(605735)
	if f == math.Inf(-1) {
		__antithesis_instrumentation__.Notify(605739)
		return dNaNFloat, true
	} else {
		__antithesis_instrumentation__.Notify(605740)
	}
	__antithesis_instrumentation__.Notify(605736)
	return NewDFloat(DFloat(math.Nextafter(f, math.Inf(-1)))), true
}

func (d *DFloat) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605741)
	f := float64(*d)
	if math.IsNaN(f) {
		__antithesis_instrumentation__.Notify(605744)
		return dNegInfFloat, true
	} else {
		__antithesis_instrumentation__.Notify(605745)
	}
	__antithesis_instrumentation__.Notify(605742)
	if f == math.Inf(+1) {
		__antithesis_instrumentation__.Notify(605746)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(605747)
	}
	__antithesis_instrumentation__.Notify(605743)
	return NewDFloat(DFloat(math.Nextafter(f, math.Inf(+1)))), true
}

var dZeroFloat = NewDFloat(0.0)
var dPosInfFloat = NewDFloat(DFloat(math.Inf(+1)))
var dNegInfFloat = NewDFloat(DFloat(math.Inf(-1)))
var dNaNFloat = NewDFloat(DFloat(math.NaN()))

func (d *DFloat) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605748)
	return *d == *dPosInfFloat
}

func (d *DFloat) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605749)
	return math.IsNaN(float64(*d))
}

func (d *DFloat) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605750)
	return dPosInfFloat, true
}

func (d *DFloat) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605751)
	return dNaNFloat, true
}

func (*DFloat) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605752); return true }

func (d *DFloat) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605753)
	fl := float64(*d)

	disambiguate := ctx.flags.HasFlags(fmtDisambiguateDatumTypes)
	parsable := ctx.flags.HasFlags(FmtParsableNumerics)
	quote := parsable && func() bool {
		__antithesis_instrumentation__.Notify(605756)
		return (math.IsNaN(fl) || func() bool {
			__antithesis_instrumentation__.Notify(605757)
			return math.IsInf(fl, 0) == true
		}() == true) == true
	}() == true

	needParens := !quote && func() bool {
		__antithesis_instrumentation__.Notify(605758)
		return (disambiguate || func() bool {
			__antithesis_instrumentation__.Notify(605759)
			return parsable == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(605760)
		return math.Signbit(fl) == true
	}() == true

	if quote {
		__antithesis_instrumentation__.Notify(605761)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(605762)
		if needParens {
			__antithesis_instrumentation__.Notify(605763)
			ctx.WriteByte('(')
		} else {
			__antithesis_instrumentation__.Notify(605764)
		}
	}
	__antithesis_instrumentation__.Notify(605754)
	if _, frac := math.Modf(fl); frac == 0 && func() bool {
		__antithesis_instrumentation__.Notify(605765)
		return -1000000 < *d == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(605766)
		return *d < 1000000 == true
	}() == true {
		__antithesis_instrumentation__.Notify(605767)

		ctx.Printf("%.1f", fl)
	} else {
		__antithesis_instrumentation__.Notify(605768)
		ctx.Printf("%g", fl)
	}
	__antithesis_instrumentation__.Notify(605755)
	if quote {
		__antithesis_instrumentation__.Notify(605769)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(605770)
		if needParens {
			__antithesis_instrumentation__.Notify(605771)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(605772)
		}
	}
}

func (d *DFloat) Size() uintptr {
	__antithesis_instrumentation__.Notify(605773)
	return unsafe.Sizeof(*d)
}

func (d *DFloat) IsComposite() bool {
	__antithesis_instrumentation__.Notify(605774)

	return math.Float64bits(float64(*d)) == 1<<63
}

type DDecimal struct {
	apd.Decimal
}

func MustBeDDecimal(e Expr) DDecimal {
	__antithesis_instrumentation__.Notify(605775)
	switch t := e.(type) {
	case *DDecimal:
		__antithesis_instrumentation__.Notify(605777)
		return *t
	}
	__antithesis_instrumentation__.Notify(605776)
	panic(errors.AssertionFailedf("expected *DDecimal, found %T", e))
}

func ParseDDecimal(s string) (*DDecimal, error) {
	__antithesis_instrumentation__.Notify(605778)
	dd := &DDecimal{}
	err := dd.SetString(s)
	return dd, err
}

func (d *DDecimal) SetString(s string) error {
	__antithesis_instrumentation__.Notify(605779)

	_, res, err := ExactCtx.SetString(&d.Decimal, s)
	if res != 0 || func() bool {
		__antithesis_instrumentation__.Notify(605782)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(605783)
		return MakeParseError(s, types.Decimal, nil)
	} else {
		__antithesis_instrumentation__.Notify(605784)
	}
	__antithesis_instrumentation__.Notify(605780)
	switch d.Form {
	case apd.NaNSignaling:
		__antithesis_instrumentation__.Notify(605785)
		d.Form = apd.NaN
		d.Negative = false
	case apd.NaN:
		__antithesis_instrumentation__.Notify(605786)
		d.Negative = false
	case apd.Finite:
		__antithesis_instrumentation__.Notify(605787)
		if d.IsZero() && func() bool {
			__antithesis_instrumentation__.Notify(605789)
			return d.Negative == true
		}() == true {
			__antithesis_instrumentation__.Notify(605790)
			d.Negative = false
		} else {
			__antithesis_instrumentation__.Notify(605791)
		}
	default:
		__antithesis_instrumentation__.Notify(605788)
	}
	__antithesis_instrumentation__.Notify(605781)
	return nil
}

func (*DDecimal) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605792)
	return types.Decimal
}

func (d *DDecimal) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605793)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605795)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605796)
	}
	__antithesis_instrumentation__.Notify(605794)
	return res
}

func (d *DDecimal) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605797)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605800)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605801)
	}
	__antithesis_instrumentation__.Notify(605798)
	var v apd.Decimal
	switch t := UnwrapDatum(ctx, other).(type) {
	case *DDecimal:
		__antithesis_instrumentation__.Notify(605802)
		v.Set(&t.Decimal)
	case *DInt:
		__antithesis_instrumentation__.Notify(605803)
		v.SetInt64(int64(*t))
	case *DFloat:
		__antithesis_instrumentation__.Notify(605804)
		if _, err := v.SetFloat64(float64(*t)); err != nil {
			__antithesis_instrumentation__.Notify(605806)
			panic(errors.NewAssertionErrorWithWrappedErrf(err, "decimal compare, unexpected error"))
		} else {
			__antithesis_instrumentation__.Notify(605807)
		}
	default:
		__antithesis_instrumentation__.Notify(605805)
		return 0, makeUnsupportedComparisonMessage(d, other)
	}
	__antithesis_instrumentation__.Notify(605799)
	res := CompareDecimals(&d.Decimal, &v)
	return res, nil
}

func CompareDecimals(d *apd.Decimal, v *apd.Decimal) int {
	__antithesis_instrumentation__.Notify(605808)

	if dn, vn := d.Form == apd.NaN, v.Form == apd.NaN; dn && func() bool {
		__antithesis_instrumentation__.Notify(605810)
		return !vn == true
	}() == true {
		__antithesis_instrumentation__.Notify(605811)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(605812)
		if !dn && func() bool {
			__antithesis_instrumentation__.Notify(605813)
			return vn == true
		}() == true {
			__antithesis_instrumentation__.Notify(605814)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(605815)
			if dn && func() bool {
				__antithesis_instrumentation__.Notify(605816)
				return vn == true
			}() == true {
				__antithesis_instrumentation__.Notify(605817)
				return 0
			} else {
				__antithesis_instrumentation__.Notify(605818)
			}
		}
	}
	__antithesis_instrumentation__.Notify(605809)
	return d.Cmp(v)
}

func (d *DDecimal) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605819)
	return nil, false
}

func (d *DDecimal) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605820)
	return nil, false
}

var dZeroDecimal = &DDecimal{Decimal: apd.Decimal{}}
var dPosInfDecimal = &DDecimal{Decimal: apd.Decimal{Form: apd.Infinite, Negative: false}}
var dNaNDecimal = &DDecimal{Decimal: apd.Decimal{Form: apd.NaN}}

func (d *DDecimal) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605821)
	return d.Form == apd.Infinite && func() bool {
		__antithesis_instrumentation__.Notify(605822)
		return !d.Negative == true
	}() == true
}

func (d *DDecimal) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605823)
	return d.Form == apd.NaN
}

func (d *DDecimal) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605824)
	return dPosInfDecimal, true
}

func (d *DDecimal) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605825)
	return dNaNDecimal, true
}

func (*DDecimal) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605826); return true }

func (d *DDecimal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605827)

	disambiguate := ctx.flags.HasFlags(fmtDisambiguateDatumTypes)
	parsable := ctx.flags.HasFlags(FmtParsableNumerics)
	quote := parsable && func() bool {
		__antithesis_instrumentation__.Notify(605831)
		return d.Decimal.Form != apd.Finite == true
	}() == true
	needParens := !quote && func() bool {
		__antithesis_instrumentation__.Notify(605832)
		return (disambiguate || func() bool {
			__antithesis_instrumentation__.Notify(605833)
			return parsable == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(605834)
		return d.Negative == true
	}() == true
	if needParens {
		__antithesis_instrumentation__.Notify(605835)
		ctx.WriteByte('(')
	} else {
		__antithesis_instrumentation__.Notify(605836)
	}
	__antithesis_instrumentation__.Notify(605828)
	if quote {
		__antithesis_instrumentation__.Notify(605837)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(605838)
	}
	__antithesis_instrumentation__.Notify(605829)
	ctx.WriteString(d.Decimal.String())
	if quote {
		__antithesis_instrumentation__.Notify(605839)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(605840)
	}
	__antithesis_instrumentation__.Notify(605830)
	if needParens {
		__antithesis_instrumentation__.Notify(605841)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(605842)
	}
}

func (d *DDecimal) Size() uintptr {
	__antithesis_instrumentation__.Notify(605843)
	return d.Decimal.Size()
}

var (
	decimalNegativeZero = &apd.Decimal{Negative: true}
	bigTen              = apd.NewBigInt(10)
)

func (d *DDecimal) IsComposite() bool {
	__antithesis_instrumentation__.Notify(605844)

	if d.Decimal.CmpTotal(decimalNegativeZero) == 0 {
		__antithesis_instrumentation__.Notify(605846)
		return true
	} else {
		__antithesis_instrumentation__.Notify(605847)
	}
	__antithesis_instrumentation__.Notify(605845)

	var r apd.BigInt
	r.Rem(&d.Decimal.Coeff, bigTen)
	return r.Sign() == 0
}

type DString string

func NewDString(d string) *DString {
	__antithesis_instrumentation__.Notify(605848)
	r := DString(d)
	return &r
}

func AsDString(e Expr) (DString, bool) {
	__antithesis_instrumentation__.Notify(605849)
	switch t := e.(type) {
	case *DString:
		__antithesis_instrumentation__.Notify(605851)
		return *t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(605852)
		return AsDString(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(605850)
	return "", false
}

func MustBeDString(e Expr) DString {
	__antithesis_instrumentation__.Notify(605853)
	i, ok := AsDString(e)
	if !ok {
		__antithesis_instrumentation__.Notify(605855)
		panic(errors.AssertionFailedf("expected *DString, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(605856)
	}
	__antithesis_instrumentation__.Notify(605854)
	return i
}

func (*DString) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605857)
	return types.String
}

func (d *DString) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605858)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605860)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605861)
	}
	__antithesis_instrumentation__.Notify(605859)
	return res
}

func (d *DString) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605862)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605867)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605868)
	}
	__antithesis_instrumentation__.Notify(605863)
	v, ok := UnwrapDatum(ctx, other).(*DString)
	if !ok {
		__antithesis_instrumentation__.Notify(605869)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(605870)
	}
	__antithesis_instrumentation__.Notify(605864)
	if *d < *v {
		__antithesis_instrumentation__.Notify(605871)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(605872)
	}
	__antithesis_instrumentation__.Notify(605865)
	if *d > *v {
		__antithesis_instrumentation__.Notify(605873)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605874)
	}
	__antithesis_instrumentation__.Notify(605866)
	return 0, nil
}

func (d *DString) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605875)
	return nil, false
}

func (d *DString) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605876)
	return NewDString(string(roachpb.Key(*d).Next())), true
}

func (*DString) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605877)
	return false
}

func (d *DString) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605878)
	return len(*d) == 0
}

var dEmptyString = NewDString("")

func (d *DString) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605879)
	return dEmptyString, true
}

func (d *DString) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605880)
	return nil, false
}

func (*DString) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605881); return true }

func (d *DString) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605882)
	buf, f := &ctx.Buffer, ctx.flags
	if f.HasFlags(fmtRawStrings) {
		__antithesis_instrumentation__.Notify(605883)
		buf.WriteString(string(*d))
	} else {
		__antithesis_instrumentation__.Notify(605884)
		lexbase.EncodeSQLStringWithFlags(buf, string(*d), f.EncodeFlags())
	}
}

func (d *DString) Size() uintptr {
	__antithesis_instrumentation__.Notify(605885)
	return unsafe.Sizeof(*d) + uintptr(len(*d))
}

type DCollatedString struct {
	Contents string
	Locale   string

	Key []byte
}

type CollationEnvironment struct {
	cache  map[string]collationEnvironmentCacheEntry
	buffer *collate.Buffer
}

type collationEnvironmentCacheEntry struct {
	locale string

	collator *collate.Collator
}

func (env *CollationEnvironment) getCacheEntry(
	locale string,
) (collationEnvironmentCacheEntry, error) {
	__antithesis_instrumentation__.Notify(605886)
	entry, ok := env.cache[locale]
	if !ok {
		__antithesis_instrumentation__.Notify(605888)
		if env.cache == nil {
			__antithesis_instrumentation__.Notify(605891)
			env.cache = make(map[string]collationEnvironmentCacheEntry)
		} else {
			__antithesis_instrumentation__.Notify(605892)
		}
		__antithesis_instrumentation__.Notify(605889)
		tag, err := language.Parse(locale)
		if err != nil {
			__antithesis_instrumentation__.Notify(605893)
			err = errors.NewAssertionErrorWithWrappedErrf(err, "failed to parse locale %q", locale)
			return collationEnvironmentCacheEntry{}, err
		} else {
			__antithesis_instrumentation__.Notify(605894)
		}
		__antithesis_instrumentation__.Notify(605890)

		entry = collationEnvironmentCacheEntry{locale, collate.New(tag)}
		env.cache[locale] = entry
	} else {
		__antithesis_instrumentation__.Notify(605895)
	}
	__antithesis_instrumentation__.Notify(605887)
	return entry, nil
}

func NewDCollatedString(
	contents string, locale string, env *CollationEnvironment,
) (*DCollatedString, error) {
	__antithesis_instrumentation__.Notify(605896)
	entry, err := env.getCacheEntry(locale)
	if err != nil {
		__antithesis_instrumentation__.Notify(605899)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(605900)
	}
	__antithesis_instrumentation__.Notify(605897)
	if env.buffer == nil {
		__antithesis_instrumentation__.Notify(605901)
		env.buffer = &collate.Buffer{}
	} else {
		__antithesis_instrumentation__.Notify(605902)
	}
	__antithesis_instrumentation__.Notify(605898)
	key := entry.collator.KeyFromString(env.buffer, contents)
	d := DCollatedString{contents, entry.locale, make([]byte, len(key))}
	copy(d.Key, key)
	env.buffer.Reset()
	return &d, nil
}

func (*DCollatedString) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(605903)
	return false
}

func (d *DCollatedString) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605904)
	lexbase.EncodeSQLString(&ctx.Buffer, d.Contents)
	ctx.WriteString(" COLLATE ")
	lex.EncodeLocaleName(&ctx.Buffer, d.Locale)
}

func (d *DCollatedString) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605905)
	return types.MakeCollatedString(types.String, d.Locale)
}

func (d *DCollatedString) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605906)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605908)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605909)
	}
	__antithesis_instrumentation__.Notify(605907)
	return res
}

func (d *DCollatedString) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605910)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605913)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605914)
	}
	__antithesis_instrumentation__.Notify(605911)
	v, ok := UnwrapDatum(ctx, other).(*DCollatedString)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(605915)
		return !d.ResolvedType().Equivalent(other.ResolvedType()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(605916)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(605917)
	}
	__antithesis_instrumentation__.Notify(605912)
	res := bytes.Compare(d.Key, v.Key)
	return res, nil
}

func (d *DCollatedString) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605918)
	return nil, false
}

func (d *DCollatedString) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605919)
	return nil, false
}

func (*DCollatedString) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605920)
	return false
}

func (d *DCollatedString) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605921)
	return d.Contents == ""
}

func (d *DCollatedString) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605922)
	return &DCollatedString{"", d.Locale, nil}, true
}

func (d *DCollatedString) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605923)
	return nil, false
}

func (d *DCollatedString) Size() uintptr {
	__antithesis_instrumentation__.Notify(605924)
	return unsafe.Sizeof(*d) + uintptr(len(d.Contents)) + uintptr(len(d.Locale)) + uintptr(len(d.Key))
}

func (d *DCollatedString) IsComposite() bool {
	__antithesis_instrumentation__.Notify(605925)
	return true
}

type DBytes string

func NewDBytes(d DBytes) *DBytes {
	__antithesis_instrumentation__.Notify(605926)
	return &d
}

func MustBeDBytes(e Expr) DBytes {
	__antithesis_instrumentation__.Notify(605927)
	i, ok := AsDBytes(e)
	if !ok {
		__antithesis_instrumentation__.Notify(605929)
		panic(errors.AssertionFailedf("expected *DBytes, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(605930)
	}
	__antithesis_instrumentation__.Notify(605928)
	return i
}

func AsDBytes(e Expr) (DBytes, bool) {
	__antithesis_instrumentation__.Notify(605931)
	switch t := e.(type) {
	case *DBytes:
		__antithesis_instrumentation__.Notify(605933)
		return *t, true
	}
	__antithesis_instrumentation__.Notify(605932)
	return "", false
}

func (*DBytes) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605934)
	return types.Bytes
}

func (d *DBytes) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605935)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605937)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605938)
	}
	__antithesis_instrumentation__.Notify(605936)
	return res
}

func (d *DBytes) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605939)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605944)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605945)
	}
	__antithesis_instrumentation__.Notify(605940)
	v, ok := UnwrapDatum(ctx, other).(*DBytes)
	if !ok {
		__antithesis_instrumentation__.Notify(605946)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(605947)
	}
	__antithesis_instrumentation__.Notify(605941)
	if *d < *v {
		__antithesis_instrumentation__.Notify(605948)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(605949)
	}
	__antithesis_instrumentation__.Notify(605942)
	if *d > *v {
		__antithesis_instrumentation__.Notify(605950)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605951)
	}
	__antithesis_instrumentation__.Notify(605943)
	return 0, nil
}

func (d *DBytes) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605952)
	return nil, false
}

func (d *DBytes) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605953)
	return NewDBytes(DBytes(roachpb.Key(*d).Next())), true
}

func (*DBytes) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605954)
	return false
}

func (d *DBytes) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605955)
	return len(*d) == 0
}

var dEmptyBytes = NewDBytes(DBytes(""))

func (d *DBytes) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605956)
	return dEmptyBytes, true
}

func (d *DBytes) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605957)
	return nil, false
}

func (*DBytes) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605958); return true }

func writeAsHexString(ctx *FmtCtx, d *DBytes) {
	__antithesis_instrumentation__.Notify(605959)
	b := string(*d)
	for i := 0; i < len(b); i++ {
		__antithesis_instrumentation__.Notify(605960)
		ctx.Write(stringencoding.RawHexMap[b[i]])
	}
}

func (d *DBytes) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605961)
	f := ctx.flags
	if f.HasFlags(fmtPgwireFormat) {
		__antithesis_instrumentation__.Notify(605962)
		ctx.WriteString(`"\\x`)
		writeAsHexString(ctx, d)
		ctx.WriteString(`"`)
	} else {
		__antithesis_instrumentation__.Notify(605963)
		if f.HasFlags(fmtFormatByteLiterals) {
			__antithesis_instrumentation__.Notify(605964)
			ctx.WriteByte('x')
			ctx.WriteByte('\'')
			_, _ = hex.NewEncoder(ctx).Write([]byte(*d))
			ctx.WriteByte('\'')
		} else {
			__antithesis_instrumentation__.Notify(605965)
			withQuotes := !f.HasFlags(FmtFlags(lexbase.EncBareStrings))
			if withQuotes {
				__antithesis_instrumentation__.Notify(605967)
				ctx.WriteByte('\'')
			} else {
				__antithesis_instrumentation__.Notify(605968)
			}
			__antithesis_instrumentation__.Notify(605966)
			ctx.WriteString("\\x")
			writeAsHexString(ctx, d)
			if withQuotes {
				__antithesis_instrumentation__.Notify(605969)
				ctx.WriteByte('\'')
			} else {
				__antithesis_instrumentation__.Notify(605970)
			}
		}
	}
}

func (d *DBytes) Size() uintptr {
	__antithesis_instrumentation__.Notify(605971)
	return unsafe.Sizeof(*d) + uintptr(len(*d))
}

type DUuid struct {
	uuid.UUID
}

func NewDUuid(d DUuid) *DUuid {
	__antithesis_instrumentation__.Notify(605972)
	return &d
}

func (*DUuid) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(605973)
	return types.Uuid
}

func (d *DUuid) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(605974)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(605976)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(605977)
	}
	__antithesis_instrumentation__.Notify(605975)
	return res
}

func (d *DUuid) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(605978)
	if other == DNull {
		__antithesis_instrumentation__.Notify(605981)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(605982)
	}
	__antithesis_instrumentation__.Notify(605979)
	v, ok := UnwrapDatum(ctx, other).(*DUuid)
	if !ok {
		__antithesis_instrumentation__.Notify(605983)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(605984)
	}
	__antithesis_instrumentation__.Notify(605980)
	res := bytes.Compare(d.GetBytes(), v.GetBytes())
	return res, nil
}

func (d *DUuid) equal(other *DUuid) bool {
	__antithesis_instrumentation__.Notify(605985)
	return bytes.Equal(d.GetBytes(), other.GetBytes())
}

func (d *DUuid) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605986)
	i := d.ToUint128()
	u := uuid.FromUint128(i.Sub(1))
	return NewDUuid(DUuid{u}), true
}

func (d *DUuid) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605987)
	i := d.ToUint128()
	u := uuid.FromUint128(i.Add(1))
	return NewDUuid(DUuid{u}), true
}

func (d *DUuid) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605988)
	return d.equal(DMaxUUID)
}

func (d *DUuid) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(605989)
	return d.equal(DMinUUID)
}

var DMinUUID = NewDUuid(DUuid{uuid.UUID{}})

var DMaxUUID = NewDUuid(DUuid{uuid.UUID{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}})

func (*DUuid) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605990)
	return DMinUUID, true
}

func (*DUuid) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(605991)
	return DMaxUUID, true
}

func (*DUuid) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(605992); return true }

func (d *DUuid) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(605993)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(605995)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(605996)
	}
	__antithesis_instrumentation__.Notify(605994)
	ctx.WriteString(d.UUID.String())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(605997)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(605998)
	}
}

func (d *DUuid) Size() uintptr {
	__antithesis_instrumentation__.Notify(605999)
	return unsafe.Sizeof(*d)
}

type DIPAddr struct {
	ipaddr.IPAddr
}

func NewDIPAddr(d DIPAddr) *DIPAddr {
	__antithesis_instrumentation__.Notify(606000)
	return &d
}

func AsDIPAddr(e Expr) (DIPAddr, bool) {
	__antithesis_instrumentation__.Notify(606001)
	switch t := e.(type) {
	case *DIPAddr:
		__antithesis_instrumentation__.Notify(606003)
		return *t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606004)
		return AsDIPAddr(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606002)
	return DIPAddr{}, false
}

func MustBeDIPAddr(e Expr) DIPAddr {
	__antithesis_instrumentation__.Notify(606005)
	i, ok := AsDIPAddr(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606007)
		panic(errors.AssertionFailedf("expected *DIPAddr, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606008)
	}
	__antithesis_instrumentation__.Notify(606006)
	return i
}

func (*DIPAddr) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606009)
	return types.INet
}

func (d *DIPAddr) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606010)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606012)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606013)
	}
	__antithesis_instrumentation__.Notify(606011)
	return res
}

func (d *DIPAddr) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606014)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606017)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606018)
	}
	__antithesis_instrumentation__.Notify(606015)
	v, ok := UnwrapDatum(ctx, other).(*DIPAddr)
	if !ok {
		__antithesis_instrumentation__.Notify(606019)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606020)
	}
	__antithesis_instrumentation__.Notify(606016)

	res := d.IPAddr.Compare(&v.IPAddr)
	return res, nil
}

func (d DIPAddr) equal(other *DIPAddr) bool {
	__antithesis_instrumentation__.Notify(606021)
	return d.IPAddr.Equal(&other.IPAddr)
}

func (d *DIPAddr) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606022)

	if d.Family == ipaddr.IPv6family && func() bool {
		__antithesis_instrumentation__.Notify(606024)
		return d.Addr.Equal(dIPv6min) == true
	}() == true {
		__antithesis_instrumentation__.Notify(606025)
		if d.Mask == 0 {
			__antithesis_instrumentation__.Notify(606027)

			return dMaxIPv4Addr, true
		} else {
			__antithesis_instrumentation__.Notify(606028)
		}
		__antithesis_instrumentation__.Notify(606026)

		return NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv6family, Addr: dIPv6max, Mask: d.Mask - 1}}), true
	} else {
		__antithesis_instrumentation__.Notify(606029)
		if d.Family == ipaddr.IPv4family && func() bool {
			__antithesis_instrumentation__.Notify(606030)
			return d.Addr.Equal(dIPv4min) == true
		}() == true {
			__antithesis_instrumentation__.Notify(606031)

			return NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv4family, Addr: dIPv4max, Mask: d.Mask - 1}}), true
		} else {
			__antithesis_instrumentation__.Notify(606032)
		}
	}
	__antithesis_instrumentation__.Notify(606023)

	return NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: d.Family, Addr: d.Addr.Sub(1), Mask: d.Mask}}), true
}

func (d *DIPAddr) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606033)

	if d.Family == ipaddr.IPv4family && func() bool {
		__antithesis_instrumentation__.Notify(606035)
		return d.Addr.Equal(dIPv4max) == true
	}() == true {
		__antithesis_instrumentation__.Notify(606036)
		if d.Mask == 32 {
			__antithesis_instrumentation__.Notify(606038)

			return dMinIPv6Addr, true
		} else {
			__antithesis_instrumentation__.Notify(606039)
		}
		__antithesis_instrumentation__.Notify(606037)

		return NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv4family, Addr: dIPv4min, Mask: d.Mask + 1}}), true
	} else {
		__antithesis_instrumentation__.Notify(606040)
		if d.Family == ipaddr.IPv6family && func() bool {
			__antithesis_instrumentation__.Notify(606041)
			return d.Addr.Equal(dIPv6max) == true
		}() == true {
			__antithesis_instrumentation__.Notify(606042)

			return NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv6family, Addr: dIPv6min, Mask: d.Mask + 1}}), true
		} else {
			__antithesis_instrumentation__.Notify(606043)
		}
	}
	__antithesis_instrumentation__.Notify(606034)

	return NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: d.Family, Addr: d.Addr.Add(1), Mask: d.Mask}}), true
}

func (d *DIPAddr) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606044)
	return d.equal(DMaxIPAddr)
}

func (d *DIPAddr) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606045)
	return d.equal(DMinIPAddr)
}

var dIPv4min = ipaddr.Addr(uint128.FromBytes([]byte(ipaddr.ParseIP("0.0.0.0"))))
var dIPv4max = ipaddr.Addr(uint128.FromBytes([]byte(ipaddr.ParseIP("255.255.255.255"))))
var dIPv6min = ipaddr.Addr(uint128.FromBytes([]byte(ipaddr.ParseIP("::"))))
var dIPv6max = ipaddr.Addr(uint128.FromBytes([]byte(ipaddr.ParseIP("ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff"))))

var dMaxIPv4Addr = NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv4family, Addr: dIPv4max, Mask: 32}})
var dMinIPv6Addr = NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv6family, Addr: dIPv6min, Mask: 0}})

var DMinIPAddr = NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv4family, Addr: dIPv4min, Mask: 0}})

var DMaxIPAddr = NewDIPAddr(DIPAddr{ipaddr.IPAddr{Family: ipaddr.IPv6family, Addr: dIPv6max, Mask: 128}})

func (*DIPAddr) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606046)
	return DMinIPAddr, true
}

func (*DIPAddr) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606047)
	return DMaxIPAddr, true
}

func (*DIPAddr) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(606048)
	return true
}

func (d *DIPAddr) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606049)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606051)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606052)
	}
	__antithesis_instrumentation__.Notify(606050)
	ctx.WriteString(d.IPAddr.String())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606053)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606054)
	}
}

func (d *DIPAddr) Size() uintptr {
	__antithesis_instrumentation__.Notify(606055)
	return unsafe.Sizeof(*d)
}

type DDate struct {
	pgdate.Date
}

func NewDDate(d pgdate.Date) *DDate {
	__antithesis_instrumentation__.Notify(606056)
	return &DDate{Date: d}
}

func MakeDDate(d pgdate.Date) DDate {
	__antithesis_instrumentation__.Notify(606057)
	return DDate{Date: d}
}

func NewDDateFromTime(t time.Time) (*DDate, error) {
	__antithesis_instrumentation__.Notify(606058)
	d, err := pgdate.MakeDateFromTime(t)
	return NewDDate(d), err
}

type ParseTimeContext interface {
	GetRelativeParseTime() time.Time

	GetIntervalStyle() duration.IntervalStyle

	GetDateStyle() pgdate.DateStyle
}

var _ ParseTimeContext = &EvalContext{}
var _ ParseTimeContext = &simpleParseTimeContext{}

type NewParseTimeContextOption func(ret *simpleParseTimeContext)

func NewParseTimeContextOptionDateStyle(dateStyle pgdate.DateStyle) NewParseTimeContextOption {
	__antithesis_instrumentation__.Notify(606059)
	return func(ret *simpleParseTimeContext) {
		__antithesis_instrumentation__.Notify(606060)
		ret.DateStyle = dateStyle
	}
}

func NewParseTimeContext(
	relativeParseTime time.Time, opts ...NewParseTimeContextOption,
) ParseTimeContext {
	__antithesis_instrumentation__.Notify(606061)
	ret := &simpleParseTimeContext{
		RelativeParseTime: relativeParseTime,
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(606063)
		opt(ret)
	}
	__antithesis_instrumentation__.Notify(606062)
	return ret
}

type simpleParseTimeContext struct {
	RelativeParseTime time.Time
	DateStyle         pgdate.DateStyle
	IntervalStyle     duration.IntervalStyle
}

func (ctx simpleParseTimeContext) GetRelativeParseTime() time.Time {
	__antithesis_instrumentation__.Notify(606064)
	return ctx.RelativeParseTime
}

func (ctx simpleParseTimeContext) GetIntervalStyle() duration.IntervalStyle {
	__antithesis_instrumentation__.Notify(606065)
	return ctx.IntervalStyle
}

func (ctx simpleParseTimeContext) GetDateStyle() pgdate.DateStyle {
	__antithesis_instrumentation__.Notify(606066)
	return ctx.DateStyle
}

func relativeParseTime(ctx ParseTimeContext) time.Time {
	__antithesis_instrumentation__.Notify(606067)
	if ctx == nil {
		__antithesis_instrumentation__.Notify(606069)
		return timeutil.Now()
	} else {
		__antithesis_instrumentation__.Notify(606070)
	}
	__antithesis_instrumentation__.Notify(606068)
	return ctx.GetRelativeParseTime()
}

func dateStyle(ctx ParseTimeContext) pgdate.DateStyle {
	__antithesis_instrumentation__.Notify(606071)
	if ctx == nil {
		__antithesis_instrumentation__.Notify(606073)
		return pgdate.DefaultDateStyle()
	} else {
		__antithesis_instrumentation__.Notify(606074)
	}
	__antithesis_instrumentation__.Notify(606072)
	return ctx.GetDateStyle()
}

func intervalStyle(ctx ParseTimeContext) duration.IntervalStyle {
	__antithesis_instrumentation__.Notify(606075)
	if ctx == nil {
		__antithesis_instrumentation__.Notify(606077)
		return duration.IntervalStyle_POSTGRES
	} else {
		__antithesis_instrumentation__.Notify(606078)
	}
	__antithesis_instrumentation__.Notify(606076)
	return ctx.GetIntervalStyle()
}

func ParseDDate(ctx ParseTimeContext, s string) (_ *DDate, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(606079)
	now := relativeParseTime(ctx)
	t, dependsOnContext, err := pgdate.ParseDate(now, dateStyle(ctx), s)
	return NewDDate(t), dependsOnContext, err
}

func AsDDate(e Expr) (DDate, bool) {
	__antithesis_instrumentation__.Notify(606080)
	switch t := e.(type) {
	case *DDate:
		__antithesis_instrumentation__.Notify(606082)
		return *t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606083)
		return AsDDate(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606081)
	return DDate{}, false
}

func MustBeDDate(e Expr) DDate {
	__antithesis_instrumentation__.Notify(606084)
	t, ok := AsDDate(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606086)
		panic(errors.AssertionFailedf("expected *DDate, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606087)
	}
	__antithesis_instrumentation__.Notify(606085)
	return t
}

func (*DDate) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606088)
	return types.Date
}

func (d *DDate) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606089)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606091)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606092)
	}
	__antithesis_instrumentation__.Notify(606090)
	return res
}

func (d *DDate) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606093)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606096)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606097)
	}
	__antithesis_instrumentation__.Notify(606094)
	var v DDate
	switch t := UnwrapDatum(ctx, other).(type) {
	case *DDate:
		__antithesis_instrumentation__.Notify(606098)
		v = *t
	case *DTimestamp, *DTimestampTZ:
		__antithesis_instrumentation__.Notify(606099)
		return compareTimestamps(ctx, d, other)
	default:
		__antithesis_instrumentation__.Notify(606100)
		return 0, makeUnsupportedComparisonMessage(d, other)
	}
	__antithesis_instrumentation__.Notify(606095)
	res := d.Date.Compare(v.Date)
	return res, nil
}

var (
	epochDate, _ = pgdate.MakeDateFromPGEpoch(0)
	dEpochDate   = NewDDate(epochDate)
	dMaxDate     = NewDDate(pgdate.PosInfDate)
	dMinDate     = NewDDate(pgdate.NegInfDate)
	dLowDate     = NewDDate(pgdate.LowDate)
	dHighDate    = NewDDate(pgdate.HighDate)
)

func (d *DDate) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606101)
	switch d.Date {
	case pgdate.PosInfDate:
		__antithesis_instrumentation__.Notify(606104)
		return dHighDate, true
	case pgdate.LowDate:
		__antithesis_instrumentation__.Notify(606105)
		return dMinDate, true
	case pgdate.NegInfDate:
		__antithesis_instrumentation__.Notify(606106)
		return nil, false
	default:
		__antithesis_instrumentation__.Notify(606107)
	}
	__antithesis_instrumentation__.Notify(606102)
	n, err := d.AddDays(-1)
	if err != nil {
		__antithesis_instrumentation__.Notify(606108)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606109)
	}
	__antithesis_instrumentation__.Notify(606103)
	return NewDDate(n), true
}

func (d *DDate) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606110)
	switch d.Date {
	case pgdate.NegInfDate:
		__antithesis_instrumentation__.Notify(606113)
		return dLowDate, true
	case pgdate.HighDate:
		__antithesis_instrumentation__.Notify(606114)
		return dMaxDate, true
	case pgdate.PosInfDate:
		__antithesis_instrumentation__.Notify(606115)
		return nil, false
	default:
		__antithesis_instrumentation__.Notify(606116)
	}
	__antithesis_instrumentation__.Notify(606111)
	n, err := d.AddDays(1)
	if err != nil {
		__antithesis_instrumentation__.Notify(606117)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606118)
	}
	__antithesis_instrumentation__.Notify(606112)
	return NewDDate(n), true
}

func (d *DDate) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606119)
	return d.PGEpochDays() == pgdate.PosInfDate.PGEpochDays()
}

func (d *DDate) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606120)
	return d.PGEpochDays() == pgdate.NegInfDate.PGEpochDays()
}

func (d *DDate) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606121)
	return dMaxDate, true
}

func (d *DDate) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606122)
	return dMinDate, true
}

func (*DDate) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606123); return true }

func FormatDate(d pgdate.Date, ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606124)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606126)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606127)
	}
	__antithesis_instrumentation__.Notify(606125)
	d.Format(&ctx.Buffer)
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606128)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606129)
	}
}

func (d *DDate) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606130)
	FormatDate(d.Date, ctx)
}

func (d *DDate) Size() uintptr {
	__antithesis_instrumentation__.Notify(606131)
	return unsafe.Sizeof(*d)
}

type DTime timeofday.TimeOfDay

func MakeDTime(t timeofday.TimeOfDay) *DTime {
	__antithesis_instrumentation__.Notify(606132)
	d := DTime(t)
	return &d
}

func ParseDTime(
	ctx ParseTimeContext, s string, precision time.Duration,
) (_ *DTime, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(606133)
	now := relativeParseTime(ctx)

	if DTimeMaxTimeRegex.MatchString(s) {
		__antithesis_instrumentation__.Notify(606136)
		return MakeDTime(timeofday.Time2400), false, nil
	} else {
		__antithesis_instrumentation__.Notify(606137)
	}
	__antithesis_instrumentation__.Notify(606134)

	s = timeutil.ReplaceLibPQTimePrefix(s)

	t, dependsOnContext, err := pgdate.ParseTimeWithoutTimezone(now, dateStyle(ctx), s)
	if err != nil {
		__antithesis_instrumentation__.Notify(606138)

		return nil, false, MakeParseError(s, types.Time, nil)
	} else {
		__antithesis_instrumentation__.Notify(606139)
	}
	__antithesis_instrumentation__.Notify(606135)
	return MakeDTime(timeofday.FromTime(t).Round(precision)), dependsOnContext, nil
}

func (*DTime) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606140)
	return types.Time
}

func (d *DTime) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606141)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606143)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606144)
	}
	__antithesis_instrumentation__.Notify(606142)
	return res
}

func (d *DTime) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606145)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606147)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606148)
	}
	__antithesis_instrumentation__.Notify(606146)
	return compareTimestamps(ctx, d, other)
}

func (d *DTime) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606149)
	if d.IsMin(ctx) {
		__antithesis_instrumentation__.Notify(606151)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606152)
	}
	__antithesis_instrumentation__.Notify(606150)
	prev := *d - 1
	return &prev, true
}

func (d *DTime) Round(precision time.Duration) *DTime {
	__antithesis_instrumentation__.Notify(606153)
	return MakeDTime(timeofday.TimeOfDay(*d).Round(precision))
}

func (d *DTime) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606154)
	if d.IsMax(ctx) {
		__antithesis_instrumentation__.Notify(606156)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606157)
	}
	__antithesis_instrumentation__.Notify(606155)
	next := *d + 1
	return &next, true
}

var dTimeMin = MakeDTime(timeofday.Min)
var dTimeMax = MakeDTime(timeofday.Max)

func (d *DTime) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606158)
	return *d == *dTimeMax
}

func (d *DTime) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606159)
	return *d == *dTimeMin
}

func (d *DTime) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606160)
	return dTimeMax, true
}

func (d *DTime) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606161)
	return dTimeMin, true
}

func (*DTime) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606162); return true }

func (d *DTime) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606163)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606165)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606166)
	}
	__antithesis_instrumentation__.Notify(606164)
	ctx.WriteString(timeofday.TimeOfDay(*d).String())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606167)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606168)
	}
}

func (d *DTime) Size() uintptr {
	__antithesis_instrumentation__.Notify(606169)
	return unsafe.Sizeof(*d)
}

type DTimeTZ struct {
	timetz.TimeTZ
}

var (
	dZeroTimeTZ = NewDTimeTZFromOffset(timeofday.Min, 0)

	DMinTimeTZ = NewDTimeTZFromOffset(timeofday.Min, timetz.MinTimeTZOffsetSecs)

	DMaxTimeTZ = NewDTimeTZFromOffset(timeofday.Max, timetz.MaxTimeTZOffsetSecs)
)

func NewDTimeTZ(t timetz.TimeTZ) *DTimeTZ {
	__antithesis_instrumentation__.Notify(606170)
	return &DTimeTZ{t}
}

func NewDTimeTZFromTime(t time.Time) *DTimeTZ {
	__antithesis_instrumentation__.Notify(606171)
	return &DTimeTZ{timetz.MakeTimeTZFromTime(t)}
}

func NewDTimeTZFromOffset(t timeofday.TimeOfDay, offsetSecs int32) *DTimeTZ {
	__antithesis_instrumentation__.Notify(606172)
	return &DTimeTZ{timetz.MakeTimeTZ(t, offsetSecs)}
}

func NewDTimeTZFromLocation(t timeofday.TimeOfDay, loc *time.Location) *DTimeTZ {
	__antithesis_instrumentation__.Notify(606173)
	return &DTimeTZ{timetz.MakeTimeTZFromLocation(t, loc)}
}

func ParseDTimeTZ(
	ctx ParseTimeContext, s string, precision time.Duration,
) (_ *DTimeTZ, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(606174)
	now := relativeParseTime(ctx)
	d, dependsOnContext, err := timetz.ParseTimeTZ(now, dateStyle(ctx), s, precision)
	if err != nil {
		__antithesis_instrumentation__.Notify(606176)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(606177)
	}
	__antithesis_instrumentation__.Notify(606175)
	return NewDTimeTZ(d), dependsOnContext, nil
}

func (*DTimeTZ) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606178)
	return types.TimeTZ
}

func (d *DTimeTZ) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606179)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606181)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606182)
	}
	__antithesis_instrumentation__.Notify(606180)
	return res
}

func (d *DTimeTZ) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606183)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606185)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606186)
	}
	__antithesis_instrumentation__.Notify(606184)
	return compareTimestamps(ctx, d, other)
}

func (d *DTimeTZ) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606187)
	if d.IsMin(ctx) {
		__antithesis_instrumentation__.Notify(606190)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606191)
	}
	__antithesis_instrumentation__.Notify(606188)

	var newTimeOfDay timeofday.TimeOfDay
	var newOffsetSecs int32
	if d.OffsetSecs == timetz.MinTimeTZOffsetSecs || func() bool {
		__antithesis_instrumentation__.Notify(606192)
		return d.TimeOfDay+duration.MicrosPerSec > timeofday.Max == true
	}() == true {
		__antithesis_instrumentation__.Notify(606193)
		newTimeOfDay = d.TimeOfDay - 1
		shiftSeconds := int32((newTimeOfDay - timeofday.Min) / duration.MicrosPerSec)
		if d.OffsetSecs+shiftSeconds > timetz.MaxTimeTZOffsetSecs {
			__antithesis_instrumentation__.Notify(606195)
			shiftSeconds = timetz.MaxTimeTZOffsetSecs - d.OffsetSecs
		} else {
			__antithesis_instrumentation__.Notify(606196)
		}
		__antithesis_instrumentation__.Notify(606194)
		newOffsetSecs = d.OffsetSecs + shiftSeconds
		newTimeOfDay -= timeofday.TimeOfDay(shiftSeconds) * duration.MicrosPerSec
	} else {
		__antithesis_instrumentation__.Notify(606197)
		newTimeOfDay = d.TimeOfDay + duration.MicrosPerSec
		newOffsetSecs = d.OffsetSecs - 1
	}
	__antithesis_instrumentation__.Notify(606189)
	return NewDTimeTZFromOffset(newTimeOfDay, newOffsetSecs), true
}

func (d *DTimeTZ) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606198)
	if d.IsMax(ctx) {
		__antithesis_instrumentation__.Notify(606201)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606202)
	}
	__antithesis_instrumentation__.Notify(606199)

	var newTimeOfDay timeofday.TimeOfDay
	var newOffsetSecs int32
	if d.OffsetSecs == timetz.MaxTimeTZOffsetSecs || func() bool {
		__antithesis_instrumentation__.Notify(606203)
		return d.TimeOfDay-duration.MicrosPerSec < timeofday.Min == true
	}() == true {
		__antithesis_instrumentation__.Notify(606204)
		newTimeOfDay = d.TimeOfDay + 1
		shiftSeconds := int32((timeofday.Max - newTimeOfDay) / duration.MicrosPerSec)
		if d.OffsetSecs-shiftSeconds < timetz.MinTimeTZOffsetSecs {
			__antithesis_instrumentation__.Notify(606206)
			shiftSeconds = d.OffsetSecs - timetz.MinTimeTZOffsetSecs
		} else {
			__antithesis_instrumentation__.Notify(606207)
		}
		__antithesis_instrumentation__.Notify(606205)
		newOffsetSecs = d.OffsetSecs - shiftSeconds
		newTimeOfDay += timeofday.TimeOfDay(shiftSeconds) * duration.MicrosPerSec
	} else {
		__antithesis_instrumentation__.Notify(606208)
		newTimeOfDay = d.TimeOfDay - duration.MicrosPerSec
		newOffsetSecs = d.OffsetSecs + 1
	}
	__antithesis_instrumentation__.Notify(606200)
	return NewDTimeTZFromOffset(newTimeOfDay, newOffsetSecs), true
}

func (d *DTimeTZ) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606209)
	return d.TimeOfDay == DMaxTimeTZ.TimeOfDay && func() bool {
		__antithesis_instrumentation__.Notify(606210)
		return d.OffsetSecs == timetz.MaxTimeTZOffsetSecs == true
	}() == true
}

func (d *DTimeTZ) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606211)
	return d.TimeOfDay == DMinTimeTZ.TimeOfDay && func() bool {
		__antithesis_instrumentation__.Notify(606212)
		return d.OffsetSecs == timetz.MinTimeTZOffsetSecs == true
	}() == true
}

func (d *DTimeTZ) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606213)
	return DMaxTimeTZ, true
}

func (d *DTimeTZ) Round(precision time.Duration) *DTimeTZ {
	__antithesis_instrumentation__.Notify(606214)
	return NewDTimeTZ(d.TimeTZ.Round(precision))
}

func (d *DTimeTZ) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606215)
	return DMinTimeTZ, true
}

func (*DTimeTZ) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606216); return true }

func (d *DTimeTZ) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606217)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606219)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606220)
	}
	__antithesis_instrumentation__.Notify(606218)
	ctx.WriteString(d.TimeTZ.String())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606221)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606222)
	}
}

func (d *DTimeTZ) Size() uintptr {
	__antithesis_instrumentation__.Notify(606223)
	return unsafe.Sizeof(*d)
}

type DTimestamp struct {
	time.Time
}

func MakeDTimestamp(t time.Time, precision time.Duration) (*DTimestamp, error) {
	__antithesis_instrumentation__.Notify(606224)
	ret := t.Round(precision)
	if ret.After(MaxSupportedTime) || func() bool {
		__antithesis_instrumentation__.Notify(606226)
		return ret.Before(MinSupportedTime) == true
	}() == true {
		__antithesis_instrumentation__.Notify(606227)
		return nil, pgerror.Newf(
			pgcode.InvalidTimeZoneDisplacementValue,
			"timestamp %q exceeds supported timestamp bounds", ret.Format(time.RFC3339))
	} else {
		__antithesis_instrumentation__.Notify(606228)
	}
	__antithesis_instrumentation__.Notify(606225)
	return &DTimestamp{Time: ret}, nil
}

func MustMakeDTimestamp(t time.Time, precision time.Duration) *DTimestamp {
	__antithesis_instrumentation__.Notify(606229)
	ret, err := MakeDTimestamp(t, precision)
	if err != nil {
		__antithesis_instrumentation__.Notify(606231)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606232)
	}
	__antithesis_instrumentation__.Notify(606230)
	return ret
}

var dZeroTimestamp = &DTimestamp{}

const (
	timestampTZOutputFormat = "2006-01-02 15:04:05.999999-07:00"

	timestampOutputFormat = "2006-01-02 15:04:05.999999"
)

func ParseDTimestamp(
	ctx ParseTimeContext, s string, precision time.Duration,
) (_ *DTimestamp, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(606233)
	now := relativeParseTime(ctx)
	t, dependsOnContext, err := pgdate.ParseTimestampWithoutTimezone(now, dateStyle(ctx), s)
	if err != nil {
		__antithesis_instrumentation__.Notify(606235)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(606236)
	}
	__antithesis_instrumentation__.Notify(606234)
	d, err := MakeDTimestamp(t, precision)
	return d, dependsOnContext, err
}

func AsDTimestamp(e Expr) (DTimestamp, bool) {
	__antithesis_instrumentation__.Notify(606237)
	switch t := e.(type) {
	case *DTimestamp:
		__antithesis_instrumentation__.Notify(606239)
		return *t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606240)
		return AsDTimestamp(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606238)
	return DTimestamp{}, false
}

func MustBeDTimestamp(e Expr) DTimestamp {
	__antithesis_instrumentation__.Notify(606241)
	t, ok := AsDTimestamp(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606243)
		panic(errors.AssertionFailedf("expected *DTimestamp, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606244)
	}
	__antithesis_instrumentation__.Notify(606242)
	return t
}

func (d *DTimestamp) Round(precision time.Duration) (*DTimestamp, error) {
	__antithesis_instrumentation__.Notify(606245)
	return MakeDTimestamp(d.Time, precision)
}

func (*DTimestamp) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606246)
	return types.Timestamp
}

func timeFromDatumForComparison(ctx *EvalContext, d Datum) (time.Time, error) {
	__antithesis_instrumentation__.Notify(606247)
	d = UnwrapDatum(ctx, d)
	switch t := d.(type) {
	case *DDate:
		__antithesis_instrumentation__.Notify(606248)
		ts, err := MakeDTimestampTZFromDate(ctx.GetLocation(), t)
		if err != nil {
			__antithesis_instrumentation__.Notify(606255)
			return time.Time{}, err
		} else {
			__antithesis_instrumentation__.Notify(606256)
		}
		__antithesis_instrumentation__.Notify(606249)
		return ts.Time, nil
	case *DTimestampTZ:
		__antithesis_instrumentation__.Notify(606250)
		return t.Time, nil
	case *DTimestamp:
		__antithesis_instrumentation__.Notify(606251)

		_, zoneOffset := t.Time.In(ctx.GetLocation()).Zone()
		ts := t.Time.In(ctx.GetLocation()).Add(-time.Duration(zoneOffset) * time.Second)
		return ts, nil
	case *DTime:
		__antithesis_instrumentation__.Notify(606252)

		toTime := timeofday.TimeOfDay(*t).ToTime()
		_, zoneOffsetSecs := toTime.In(ctx.GetLocation()).Zone()
		return toTime.In(ctx.GetLocation()).Add(-time.Duration(zoneOffsetSecs) * time.Second), nil
	case *DTimeTZ:
		__antithesis_instrumentation__.Notify(606253)
		return t.ToTime(), nil
	default:
		__antithesis_instrumentation__.Notify(606254)
		return time.Time{}, errors.AssertionFailedf("unexpected type: %v", t.ResolvedType())
	}
}

type infiniteDateComparison int

const (
	negativeInfinity infiniteDateComparison = iota
	finite
	positiveInfinity
)

func checkInfiniteDate(ctx *EvalContext, d Datum) infiniteDateComparison {
	__antithesis_instrumentation__.Notify(606257)
	if _, isDate := d.(*DDate); isDate {
		__antithesis_instrumentation__.Notify(606259)
		if d.IsMax(ctx) {
			__antithesis_instrumentation__.Notify(606261)
			return positiveInfinity
		} else {
			__antithesis_instrumentation__.Notify(606262)
		}
		__antithesis_instrumentation__.Notify(606260)
		if d.IsMin(ctx) {
			__antithesis_instrumentation__.Notify(606263)
			return negativeInfinity
		} else {
			__antithesis_instrumentation__.Notify(606264)
		}
	} else {
		__antithesis_instrumentation__.Notify(606265)
	}
	__antithesis_instrumentation__.Notify(606258)
	return finite
}

func compareTimestamps(ctx *EvalContext, l Datum, r Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606266)
	leftInf := checkInfiniteDate(ctx, l)
	rightInf := checkInfiniteDate(ctx, r)
	if leftInf != finite || func() bool {
		__antithesis_instrumentation__.Notify(606276)
		return rightInf != finite == true
	}() == true {
		__antithesis_instrumentation__.Notify(606277)

		if leftInf != finite && func() bool {
			__antithesis_instrumentation__.Notify(606279)
			return rightInf != finite == true
		}() == true {
			__antithesis_instrumentation__.Notify(606280)

			return 0, errors.AssertionFailedf("unexpectedly two infinite dates in compareTimestamps")
		} else {
			__antithesis_instrumentation__.Notify(606281)
		}
		__antithesis_instrumentation__.Notify(606278)

		return int(leftInf - rightInf), nil
	} else {
		__antithesis_instrumentation__.Notify(606282)
	}
	__antithesis_instrumentation__.Notify(606267)
	lTime, lErr := timeFromDatumForComparison(ctx, l)
	rTime, rErr := timeFromDatumForComparison(ctx, r)
	if lErr != nil || func() bool {
		__antithesis_instrumentation__.Notify(606283)
		return rErr != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(606284)
		return 0, makeUnsupportedComparisonMessage(l, r)
	} else {
		__antithesis_instrumentation__.Notify(606285)
	}
	__antithesis_instrumentation__.Notify(606268)
	if lTime.Before(rTime) {
		__antithesis_instrumentation__.Notify(606286)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(606287)
	}
	__antithesis_instrumentation__.Notify(606269)
	if rTime.Before(lTime) {
		__antithesis_instrumentation__.Notify(606288)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606289)
	}
	__antithesis_instrumentation__.Notify(606270)

	_, leftIsTimeTZ := l.(*DTimeTZ)
	_, rightIsTimeTZ := r.(*DTimeTZ)

	if !leftIsTimeTZ && func() bool {
		__antithesis_instrumentation__.Notify(606290)
		return !rightIsTimeTZ == true
	}() == true {
		__antithesis_instrumentation__.Notify(606291)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(606292)
	}
	__antithesis_instrumentation__.Notify(606271)

	_, zoneOffset := ctx.GetRelativeParseTime().Zone()
	lOffset := int32(-zoneOffset)
	rOffset := int32(-zoneOffset)

	if leftIsTimeTZ {
		__antithesis_instrumentation__.Notify(606293)
		lOffset = l.(*DTimeTZ).OffsetSecs
	} else {
		__antithesis_instrumentation__.Notify(606294)
	}
	__antithesis_instrumentation__.Notify(606272)
	if rightIsTimeTZ {
		__antithesis_instrumentation__.Notify(606295)
		rOffset = r.(*DTimeTZ).OffsetSecs
	} else {
		__antithesis_instrumentation__.Notify(606296)
	}
	__antithesis_instrumentation__.Notify(606273)

	if lOffset > rOffset {
		__antithesis_instrumentation__.Notify(606297)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606298)
	}
	__antithesis_instrumentation__.Notify(606274)
	if lOffset < rOffset {
		__antithesis_instrumentation__.Notify(606299)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(606300)
	}
	__antithesis_instrumentation__.Notify(606275)
	return 0, nil
}

func (d *DTimestamp) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606301)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606303)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606304)
	}
	__antithesis_instrumentation__.Notify(606302)
	return res
}

func (d *DTimestamp) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606305)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606307)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606308)
	}
	__antithesis_instrumentation__.Notify(606306)
	return compareTimestamps(ctx, d, other)
}

func (d *DTimestamp) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606309)
	if d.IsMin(ctx) {
		__antithesis_instrumentation__.Notify(606311)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606312)
	}
	__antithesis_instrumentation__.Notify(606310)
	return &DTimestamp{Time: d.Add(-time.Microsecond)}, true
}

func (d *DTimestamp) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606313)
	if d.IsMax(ctx) {
		__antithesis_instrumentation__.Notify(606315)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606316)
	}
	__antithesis_instrumentation__.Notify(606314)
	return &DTimestamp{Time: d.Add(time.Microsecond)}, true
}

func (d *DTimestamp) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606317)
	return d.Equal(MaxSupportedTime)
}

func (d *DTimestamp) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606318)
	return d.Equal(MinSupportedTime)
}

func (d *DTimestamp) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606319)
	return &DTimestamp{Time: MinSupportedTime}, true
}

func (d *DTimestamp) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606320)
	return &DTimestamp{Time: MaxSupportedTime}, true
}

func (*DTimestamp) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606321); return true }

func (d *DTimestamp) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606322)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606324)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606325)
	}
	__antithesis_instrumentation__.Notify(606323)
	ctx.WriteString(d.UTC().Format(timestampOutputFormat))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606326)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606327)
	}
}

func (d *DTimestamp) Size() uintptr {
	__antithesis_instrumentation__.Notify(606328)
	return unsafe.Sizeof(*d)
}

type DTimestampTZ struct {
	time.Time
}

func MakeDTimestampTZ(t time.Time, precision time.Duration) (*DTimestampTZ, error) {
	__antithesis_instrumentation__.Notify(606329)
	ret := t.Round(precision)
	if ret.After(MaxSupportedTime) || func() bool {
		__antithesis_instrumentation__.Notify(606331)
		return ret.Before(MinSupportedTime) == true
	}() == true {
		__antithesis_instrumentation__.Notify(606332)
		return nil, pgerror.Newf(
			pgcode.InvalidTimeZoneDisplacementValue,
			"timestamp %q exceeds supported timestamp bounds", ret.Format(time.RFC3339))
	} else {
		__antithesis_instrumentation__.Notify(606333)
	}
	__antithesis_instrumentation__.Notify(606330)
	return &DTimestampTZ{Time: ret}, nil
}

func MustMakeDTimestampTZ(t time.Time, precision time.Duration) *DTimestampTZ {
	__antithesis_instrumentation__.Notify(606334)
	ret, err := MakeDTimestampTZ(t, precision)
	if err != nil {
		__antithesis_instrumentation__.Notify(606336)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606337)
	}
	__antithesis_instrumentation__.Notify(606335)
	return ret
}

func MakeDTimestampTZFromDate(loc *time.Location, d *DDate) (*DTimestampTZ, error) {
	__antithesis_instrumentation__.Notify(606338)
	t, err := d.ToTime()
	if err != nil {
		__antithesis_instrumentation__.Notify(606340)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(606341)
	}
	__antithesis_instrumentation__.Notify(606339)

	t = t.In(loc)
	_, offset := t.Zone()
	return MakeDTimestampTZ(t.Add(time.Duration(-offset)*time.Second), time.Microsecond)
}

func ParseDTimestampTZ(
	ctx ParseTimeContext, s string, precision time.Duration,
) (_ *DTimestampTZ, dependsOnContext bool, _ error) {
	__antithesis_instrumentation__.Notify(606342)
	now := relativeParseTime(ctx)
	t, dependsOnContext, err := pgdate.ParseTimestamp(now, dateStyle(ctx), s)
	if err != nil {
		__antithesis_instrumentation__.Notify(606344)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(606345)
	}
	__antithesis_instrumentation__.Notify(606343)

	d, err := MakeDTimestampTZ(t, precision)
	return d, dependsOnContext, err
}

var dZeroTimestampTZ = &DTimestampTZ{}

func AsDTimestampTZ(e Expr) (DTimestampTZ, bool) {
	__antithesis_instrumentation__.Notify(606346)
	switch t := e.(type) {
	case *DTimestampTZ:
		__antithesis_instrumentation__.Notify(606348)
		return *t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606349)
		return AsDTimestampTZ(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606347)
	return DTimestampTZ{}, false
}

func MustBeDTimestampTZ(e Expr) DTimestampTZ {
	__antithesis_instrumentation__.Notify(606350)
	t, ok := AsDTimestampTZ(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606352)
		panic(errors.AssertionFailedf("expected *DTimestampTZ, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606353)
	}
	__antithesis_instrumentation__.Notify(606351)
	return t
}

func (d *DTimestampTZ) Round(precision time.Duration) (*DTimestampTZ, error) {
	__antithesis_instrumentation__.Notify(606354)
	return MakeDTimestampTZ(d.Time, precision)
}

func (*DTimestampTZ) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606355)
	return types.TimestampTZ
}

func (d *DTimestampTZ) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606356)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606358)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606359)
	}
	__antithesis_instrumentation__.Notify(606357)
	return res
}

func (d *DTimestampTZ) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606360)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606362)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606363)
	}
	__antithesis_instrumentation__.Notify(606361)
	return compareTimestamps(ctx, d, other)
}

func (d *DTimestampTZ) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606364)
	if d.IsMin(ctx) {
		__antithesis_instrumentation__.Notify(606366)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606367)
	}
	__antithesis_instrumentation__.Notify(606365)
	return &DTimestampTZ{Time: d.Add(-time.Microsecond)}, true
}

func (d *DTimestampTZ) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606368)
	if d.IsMax(ctx) {
		__antithesis_instrumentation__.Notify(606370)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(606371)
	}
	__antithesis_instrumentation__.Notify(606369)
	return &DTimestampTZ{Time: d.Add(time.Microsecond)}, true
}

func (d *DTimestampTZ) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606372)
	return d.Equal(MaxSupportedTime)
}

func (d *DTimestampTZ) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606373)
	return d.Equal(MinSupportedTime)
}

func (d *DTimestampTZ) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606374)
	return &DTimestampTZ{Time: MinSupportedTime}, true
}

func (d *DTimestampTZ) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606375)
	return &DTimestampTZ{Time: MaxSupportedTime}, true
}

func (*DTimestampTZ) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(606376)
	return true
}

func (d *DTimestampTZ) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606377)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606380)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606381)
	}
	__antithesis_instrumentation__.Notify(606378)
	ctx.WriteString(d.Time.Format(timestampTZOutputFormat))
	_, offsetSecs := d.Time.Zone()

	if secondOffset := offsetSecs % 60; secondOffset != 0 {
		__antithesis_instrumentation__.Notify(606382)
		if secondOffset < 0 {
			__antithesis_instrumentation__.Notify(606384)
			secondOffset = 60 + secondOffset
		} else {
			__antithesis_instrumentation__.Notify(606385)
		}
		__antithesis_instrumentation__.Notify(606383)
		ctx.WriteByte(':')
		ctx.WriteString(fmt.Sprintf("%02d", secondOffset))
	} else {
		__antithesis_instrumentation__.Notify(606386)
	}
	__antithesis_instrumentation__.Notify(606379)
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606387)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606388)
	}
}

func (d *DTimestampTZ) Size() uintptr {
	__antithesis_instrumentation__.Notify(606389)
	return unsafe.Sizeof(*d)
}

func (d *DTimestampTZ) stripTimeZone(ctx *EvalContext) (*DTimestamp, error) {
	__antithesis_instrumentation__.Notify(606390)
	return d.EvalAtTimeZone(ctx, ctx.GetLocation())
}

func (d *DTimestampTZ) EvalAtTimeZone(ctx *EvalContext, loc *time.Location) (*DTimestamp, error) {
	__antithesis_instrumentation__.Notify(606391)
	_, locOffset := d.Time.In(loc).Zone()
	t := d.Time.UTC().Add(time.Duration(locOffset) * time.Second).UTC()
	return MakeDTimestamp(t, time.Microsecond)
}

type DInterval struct {
	duration.Duration
}

func MustBeDInterval(e Expr) *DInterval {
	__antithesis_instrumentation__.Notify(606392)
	switch t := e.(type) {
	case *DInterval:
		__antithesis_instrumentation__.Notify(606394)
		return t
	}
	__antithesis_instrumentation__.Notify(606393)
	panic(errors.AssertionFailedf("expected *DInterval, found %T", e))
}

func NewDInterval(d duration.Duration, itm types.IntervalTypeMetadata) *DInterval {
	__antithesis_instrumentation__.Notify(606395)
	ret := &DInterval{Duration: d}
	truncateDInterval(ret, itm)
	return ret
}

func ParseDInterval(style duration.IntervalStyle, s string) (*DInterval, error) {
	__antithesis_instrumentation__.Notify(606396)
	return ParseDIntervalWithTypeMetadata(style, s, types.DefaultIntervalTypeMetadata)
}

func truncateDInterval(d *DInterval, itm types.IntervalTypeMetadata) {
	__antithesis_instrumentation__.Notify(606397)
	switch itm.DurationField.DurationType {
	case types.IntervalDurationType_YEAR:
		__antithesis_instrumentation__.Notify(606398)
		d.Duration.Months = d.Duration.Months - d.Duration.Months%12
		d.Duration.Days = 0
		d.Duration.SetNanos(0)
	case types.IntervalDurationType_MONTH:
		__antithesis_instrumentation__.Notify(606399)
		d.Duration.Days = 0
		d.Duration.SetNanos(0)
	case types.IntervalDurationType_DAY:
		__antithesis_instrumentation__.Notify(606400)
		d.Duration.SetNanos(0)
	case types.IntervalDurationType_HOUR:
		__antithesis_instrumentation__.Notify(606401)
		d.Duration.SetNanos(d.Duration.Nanos() - d.Duration.Nanos()%time.Hour.Nanoseconds())
	case types.IntervalDurationType_MINUTE:
		__antithesis_instrumentation__.Notify(606402)
		d.Duration.SetNanos(d.Duration.Nanos() - d.Duration.Nanos()%time.Minute.Nanoseconds())
	case types.IntervalDurationType_SECOND, types.IntervalDurationType_UNSET:
		__antithesis_instrumentation__.Notify(606403)
		if itm.PrecisionIsSet || func() bool {
			__antithesis_instrumentation__.Notify(606405)
			return itm.Precision > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(606406)
			prec := TimeFamilyPrecisionToRoundDuration(itm.Precision)
			d.Duration.SetNanos(time.Duration(d.Duration.Nanos()).Round(prec).Nanoseconds())
		} else {
			__antithesis_instrumentation__.Notify(606407)
		}
	default:
		__antithesis_instrumentation__.Notify(606404)
	}
}

func ParseDIntervalWithTypeMetadata(
	style duration.IntervalStyle, s string, itm types.IntervalTypeMetadata,
) (*DInterval, error) {
	__antithesis_instrumentation__.Notify(606408)
	d, err := parseDInterval(style, s, itm)
	if err != nil {
		__antithesis_instrumentation__.Notify(606410)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(606411)
	}
	__antithesis_instrumentation__.Notify(606409)
	truncateDInterval(d, itm)
	return d, nil
}

func parseDInterval(
	style duration.IntervalStyle, s string, itm types.IntervalTypeMetadata,
) (*DInterval, error) {
	__antithesis_instrumentation__.Notify(606412)

	if len(s) == 0 {
		__antithesis_instrumentation__.Notify(606417)
		return nil, MakeParseError(s, types.Interval, nil)
	} else {
		__antithesis_instrumentation__.Notify(606418)
	}
	__antithesis_instrumentation__.Notify(606413)
	if s[0] == 'P' {
		__antithesis_instrumentation__.Notify(606419)

		dur, err := iso8601ToDuration(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(606421)
			return nil, MakeParseError(s, types.Interval, err)
		} else {
			__antithesis_instrumentation__.Notify(606422)
		}
		__antithesis_instrumentation__.Notify(606420)
		return &DInterval{Duration: dur}, nil
	} else {
		__antithesis_instrumentation__.Notify(606423)
	}
	__antithesis_instrumentation__.Notify(606414)
	if strings.IndexFunc(s, unicode.IsLetter) == -1 {
		__antithesis_instrumentation__.Notify(606424)

		dur, err := sqlStdToDuration(s, itm)
		if err != nil {
			__antithesis_instrumentation__.Notify(606426)
			return nil, MakeParseError(s, types.Interval, err)
		} else {
			__antithesis_instrumentation__.Notify(606427)
		}
		__antithesis_instrumentation__.Notify(606425)
		return &DInterval{Duration: dur}, nil
	} else {
		__antithesis_instrumentation__.Notify(606428)
	}
	__antithesis_instrumentation__.Notify(606415)

	dur, err := parseDuration(style, s, itm)
	if err != nil {
		__antithesis_instrumentation__.Notify(606429)
		return nil, MakeParseError(s, types.Interval, err)
	} else {
		__antithesis_instrumentation__.Notify(606430)
	}
	__antithesis_instrumentation__.Notify(606416)
	return &DInterval{Duration: dur}, nil
}

func (*DInterval) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606431)
	return types.Interval
}

func (d *DInterval) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606432)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606434)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606435)
	}
	__antithesis_instrumentation__.Notify(606433)
	return res
}

func (d *DInterval) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606436)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606439)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606440)
	}
	__antithesis_instrumentation__.Notify(606437)
	v, ok := UnwrapDatum(ctx, other).(*DInterval)
	if !ok {
		__antithesis_instrumentation__.Notify(606441)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606442)
	}
	__antithesis_instrumentation__.Notify(606438)
	res := d.Duration.Compare(v.Duration)
	return res, nil
}

func (d *DInterval) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606443)
	return nil, false
}

func (d *DInterval) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606444)
	return nil, false
}

func (d *DInterval) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606445)
	return d.Duration == dMaxInterval.Duration
}

func (d *DInterval) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606446)
	return d.Duration == dMinInterval.Duration
}

var (
	dZeroInterval = &DInterval{}
	dMaxInterval  = &DInterval{duration.MakeDuration(math.MaxInt64, math.MaxInt64, math.MaxInt64)}
	dMinInterval  = &DInterval{duration.MakeDuration(math.MinInt64, math.MinInt64, math.MinInt64)}
)

func (d *DInterval) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606447)
	return dMaxInterval, true
}

func (d *DInterval) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606448)
	return dMinInterval, true
}

func (d *DInterval) ValueAsISO8601String() string {
	__antithesis_instrumentation__.Notify(606449)
	return d.Duration.ISO8601String()
}

func (*DInterval) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606450); return true }

func FormatDuration(d duration.Duration, ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606451)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606453)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606454)
	}
	__antithesis_instrumentation__.Notify(606452)
	d.FormatWithStyle(&ctx.Buffer, ctx.dataConversionConfig.IntervalStyle)
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606455)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606456)
	}
}

func (d *DInterval) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606457)
	FormatDuration(d.Duration, ctx)
}

func (d *DInterval) Size() uintptr {
	__antithesis_instrumentation__.Notify(606458)
	return unsafe.Sizeof(*d)
}

type DGeography struct {
	geo.Geography
}

func NewDGeography(g geo.Geography) *DGeography {
	__antithesis_instrumentation__.Notify(606459)
	return &DGeography{Geography: g}
}

func AsDGeography(e Expr) (*DGeography, bool) {
	__antithesis_instrumentation__.Notify(606460)
	switch t := e.(type) {
	case *DGeography:
		__antithesis_instrumentation__.Notify(606462)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606463)
		return AsDGeography(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606461)
	return nil, false
}

func MustBeDGeography(e Expr) *DGeography {
	__antithesis_instrumentation__.Notify(606464)
	i, ok := AsDGeography(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606466)
		panic(errors.AssertionFailedf("expected *DGeography, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606467)
	}
	__antithesis_instrumentation__.Notify(606465)
	return i
}

func ParseDGeography(str string) (*DGeography, error) {
	__antithesis_instrumentation__.Notify(606468)
	g, err := geo.ParseGeography(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(606470)
		return nil, errors.Wrapf(err, "could not parse geography")
	} else {
		__antithesis_instrumentation__.Notify(606471)
	}
	__antithesis_instrumentation__.Notify(606469)
	return &DGeography{Geography: g}, nil
}

func (*DGeography) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606472)
	return types.Geography
}

func (d *DGeography) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606473)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606475)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606476)
	}
	__antithesis_instrumentation__.Notify(606474)
	return res
}

func (d *DGeography) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606477)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606480)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606481)
	}
	__antithesis_instrumentation__.Notify(606478)
	v, ok := UnwrapDatum(ctx, other).(*DGeography)
	if !ok {
		__antithesis_instrumentation__.Notify(606482)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606483)
	}
	__antithesis_instrumentation__.Notify(606479)
	res := d.Geography.Compare(v.Geography)
	return res, nil
}

func (d *DGeography) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606484)
	return nil, false
}

func (d *DGeography) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606485)
	return nil, false
}

func (d *DGeography) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606486)
	return false
}

func (d *DGeography) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606487)
	return false
}

func (d *DGeography) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606488)
	return nil, false
}

func (d *DGeography) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606489)
	return nil, false
}

func (*DGeography) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606490); return true }

func (d *DGeography) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606491)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606493)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606494)
	}
	__antithesis_instrumentation__.Notify(606492)
	ctx.WriteString(d.Geography.EWKBHex())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606495)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606496)
	}
}

func (d *DGeography) Size() uintptr {
	__antithesis_instrumentation__.Notify(606497)
	return d.Geography.SpatialObjectRef().MemSize()
}

type DGeometry struct {
	geo.Geometry
}

func NewDGeometry(g geo.Geometry) *DGeometry {
	__antithesis_instrumentation__.Notify(606498)
	return &DGeometry{Geometry: g}
}

func AsDGeometry(e Expr) (*DGeometry, bool) {
	__antithesis_instrumentation__.Notify(606499)
	switch t := e.(type) {
	case *DGeometry:
		__antithesis_instrumentation__.Notify(606501)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606502)
		return AsDGeometry(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606500)
	return nil, false
}

func MustBeDGeometry(e Expr) *DGeometry {
	__antithesis_instrumentation__.Notify(606503)
	i, ok := AsDGeometry(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606505)
		panic(errors.AssertionFailedf("expected *DGeometry, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606506)
	}
	__antithesis_instrumentation__.Notify(606504)
	return i
}

func ParseDGeometry(str string) (*DGeometry, error) {
	__antithesis_instrumentation__.Notify(606507)
	g, err := geo.ParseGeometry(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(606509)
		return nil, errors.Wrapf(err, "could not parse geometry")
	} else {
		__antithesis_instrumentation__.Notify(606510)
	}
	__antithesis_instrumentation__.Notify(606508)
	return &DGeometry{Geometry: g}, nil
}

func (*DGeometry) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606511)
	return types.Geometry
}

func (d *DGeometry) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606512)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606514)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606515)
	}
	__antithesis_instrumentation__.Notify(606513)
	return res
}

func (d *DGeometry) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606516)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606519)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606520)
	}
	__antithesis_instrumentation__.Notify(606517)
	v, ok := UnwrapDatum(ctx, other).(*DGeometry)
	if !ok {
		__antithesis_instrumentation__.Notify(606521)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606522)
	}
	__antithesis_instrumentation__.Notify(606518)
	res := d.Geometry.Compare(v.Geometry)
	return res, nil
}

func (d *DGeometry) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606523)
	return nil, false
}

func (d *DGeometry) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606524)
	return nil, false
}

func (d *DGeometry) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606525)
	return false
}

func (d *DGeometry) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606526)
	return false
}

func (d *DGeometry) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606527)
	return nil, false
}

func (d *DGeometry) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606528)
	return nil, false
}

func (*DGeometry) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606529); return true }

func (d *DGeometry) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606530)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606532)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606533)
	}
	__antithesis_instrumentation__.Notify(606531)
	ctx.WriteString(d.Geometry.EWKBHex())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606534)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606535)
	}
}

func (d *DGeometry) Size() uintptr {
	__antithesis_instrumentation__.Notify(606536)
	return d.Geometry.SpatialObjectRef().MemSize()
}

type DBox2D struct {
	geo.CartesianBoundingBox
}

func NewDBox2D(b geo.CartesianBoundingBox) *DBox2D {
	__antithesis_instrumentation__.Notify(606537)
	return &DBox2D{CartesianBoundingBox: b}
}

func ParseDBox2D(str string) (*DBox2D, error) {
	__antithesis_instrumentation__.Notify(606538)
	b, err := geo.ParseCartesianBoundingBox(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(606540)
		return nil, errors.Wrapf(err, "could not parse geometry")
	} else {
		__antithesis_instrumentation__.Notify(606541)
	}
	__antithesis_instrumentation__.Notify(606539)
	return &DBox2D{CartesianBoundingBox: b}, nil
}

func AsDBox2D(e Expr) (*DBox2D, bool) {
	__antithesis_instrumentation__.Notify(606542)
	switch t := e.(type) {
	case *DBox2D:
		__antithesis_instrumentation__.Notify(606544)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606545)
		return AsDBox2D(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606543)
	return nil, false
}

func MustBeDBox2D(e Expr) *DBox2D {
	__antithesis_instrumentation__.Notify(606546)
	i, ok := AsDBox2D(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606548)
		panic(errors.AssertionFailedf("expected *DBox2D, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606549)
	}
	__antithesis_instrumentation__.Notify(606547)
	return i
}

func (*DBox2D) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606550)
	return types.Box2D
}

func (d *DBox2D) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606551)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606553)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606554)
	}
	__antithesis_instrumentation__.Notify(606552)
	return res
}

func (d *DBox2D) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606555)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606558)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606559)
	}
	__antithesis_instrumentation__.Notify(606556)
	v, ok := UnwrapDatum(ctx, other).(*DBox2D)
	if !ok {
		__antithesis_instrumentation__.Notify(606560)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606561)
	}
	__antithesis_instrumentation__.Notify(606557)
	res := d.CartesianBoundingBox.Compare(&v.CartesianBoundingBox)
	return res, nil
}

func (d *DBox2D) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606562)
	return nil, false
}

func (d *DBox2D) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606563)
	return nil, false
}

func (d *DBox2D) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606564)
	return false
}

func (d *DBox2D) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606565)
	return false
}

func (d *DBox2D) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606566)
	return nil, false
}

func (d *DBox2D) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606567)
	return nil, false
}

func (*DBox2D) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606568); return true }

func (d *DBox2D) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606569)
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lexbase.EncBareStrings))
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606571)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606572)
	}
	__antithesis_instrumentation__.Notify(606570)
	ctx.WriteString(d.CartesianBoundingBox.Repr())
	if !bareStrings {
		__antithesis_instrumentation__.Notify(606573)
		ctx.WriteByte('\'')
	} else {
		__antithesis_instrumentation__.Notify(606574)
	}
}

func (d *DBox2D) Size() uintptr {
	__antithesis_instrumentation__.Notify(606575)
	return unsafe.Sizeof(*d) + unsafe.Sizeof(d.CartesianBoundingBox)
}

type DJSON struct{ json.JSON }

func NewDJSON(j json.JSON) *DJSON {
	__antithesis_instrumentation__.Notify(606576)
	return &DJSON{j}
}

func ParseDJSON(s string) (Datum, error) {
	__antithesis_instrumentation__.Notify(606577)
	j, err := json.ParseJSON(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(606579)
		return nil, pgerror.Wrapf(err, pgcode.Syntax, "could not parse JSON")
	} else {
		__antithesis_instrumentation__.Notify(606580)
	}
	__antithesis_instrumentation__.Notify(606578)
	return NewDJSON(j), nil
}

func MakeDJSON(d interface{}) (Datum, error) {
	__antithesis_instrumentation__.Notify(606581)
	j, err := json.MakeJSON(d)
	if err != nil {
		__antithesis_instrumentation__.Notify(606583)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(606584)
	}
	__antithesis_instrumentation__.Notify(606582)
	return &DJSON{j}, nil
}

var dNullJSON = NewDJSON(json.NullJSONValue)

func AsDJSON(e Expr) (*DJSON, bool) {
	__antithesis_instrumentation__.Notify(606585)
	switch t := e.(type) {
	case *DJSON:
		__antithesis_instrumentation__.Notify(606587)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606588)
		return AsDJSON(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606586)
	return nil, false
}

func MustBeDJSON(e Expr) DJSON {
	__antithesis_instrumentation__.Notify(606589)
	i, ok := AsDJSON(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606591)
		panic(errors.AssertionFailedf("expected *DJSON, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606592)
	}
	__antithesis_instrumentation__.Notify(606590)
	return *i
}

func AsJSON(
	d Datum, dcc sessiondatapb.DataConversionConfig, loc *time.Location,
) (json.JSON, error) {
	__antithesis_instrumentation__.Notify(606593)
	d = UnwrapDatum(nil, d)
	switch t := d.(type) {
	case *DBool:
		__antithesis_instrumentation__.Notify(606594)
		return json.FromBool(bool(*t)), nil
	case *DInt:
		__antithesis_instrumentation__.Notify(606595)
		return json.FromInt(int(*t)), nil
	case *DFloat:
		__antithesis_instrumentation__.Notify(606596)
		return json.FromFloat64(float64(*t))
	case *DDecimal:
		__antithesis_instrumentation__.Notify(606597)
		return json.FromDecimal(t.Decimal), nil
	case *DString:
		__antithesis_instrumentation__.Notify(606598)
		return json.FromString(string(*t)), nil
	case *DCollatedString:
		__antithesis_instrumentation__.Notify(606599)
		return json.FromString(t.Contents), nil
	case *DEnum:
		__antithesis_instrumentation__.Notify(606600)
		return json.FromString(t.LogicalRep), nil
	case *DJSON:
		__antithesis_instrumentation__.Notify(606601)
		return t.JSON, nil
	case *DArray:
		__antithesis_instrumentation__.Notify(606602)
		builder := json.NewArrayBuilder(t.Len())
		for _, e := range t.Array {
			__antithesis_instrumentation__.Notify(606613)
			j, err := AsJSON(e, dcc, loc)
			if err != nil {
				__antithesis_instrumentation__.Notify(606615)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(606616)
			}
			__antithesis_instrumentation__.Notify(606614)
			builder.Add(j)
		}
		__antithesis_instrumentation__.Notify(606603)
		return builder.Build(), nil
	case *DTuple:
		__antithesis_instrumentation__.Notify(606604)
		builder := json.NewObjectBuilder(len(t.D))

		t.maybePopulateType()
		labels := t.typ.TupleLabels()
		for i, e := range t.D {
			__antithesis_instrumentation__.Notify(606617)
			j, err := AsJSON(e, dcc, loc)
			if err != nil {
				__antithesis_instrumentation__.Notify(606620)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(606621)
			}
			__antithesis_instrumentation__.Notify(606618)
			var key string
			if i >= len(labels) {
				__antithesis_instrumentation__.Notify(606622)
				key = fmt.Sprintf("f%d", i+1)
			} else {
				__antithesis_instrumentation__.Notify(606623)
				key = labels[i]
			}
			__antithesis_instrumentation__.Notify(606619)
			builder.Add(key, j)
		}
		__antithesis_instrumentation__.Notify(606605)
		return builder.Build(), nil
	case *DTimestampTZ:
		__antithesis_instrumentation__.Notify(606606)

		return json.FromString(t.Time.In(loc).Format(time.RFC3339Nano)), nil
	case *DTimestamp:
		__antithesis_instrumentation__.Notify(606607)

		return json.FromString(t.UTC().Format("2006-01-02T15:04:05.999999999")), nil
	case *DDate, *DUuid, *DOid, *DInterval, *DBytes, *DIPAddr, *DTime, *DTimeTZ, *DBitArray, *DBox2D:
		__antithesis_instrumentation__.Notify(606608)
		return json.FromString(AsStringWithFlags(t, FmtBareStrings, FmtDataConversionConfig(dcc))), nil
	case *DGeometry:
		__antithesis_instrumentation__.Notify(606609)
		return json.FromSpatialObject(t.Geometry.SpatialObject(), geo.DefaultGeoJSONDecimalDigits)
	case *DGeography:
		__antithesis_instrumentation__.Notify(606610)
		return json.FromSpatialObject(t.Geography.SpatialObject(), geo.DefaultGeoJSONDecimalDigits)
	default:
		__antithesis_instrumentation__.Notify(606611)
		if d == DNull {
			__antithesis_instrumentation__.Notify(606624)
			return json.NullJSONValue, nil
		} else {
			__antithesis_instrumentation__.Notify(606625)
		}
		__antithesis_instrumentation__.Notify(606612)

		return nil, errors.AssertionFailedf("unexpected type %T for AsJSON", d)
	}
}

func (*DJSON) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606626)
	return types.Jsonb
}

func (d *DJSON) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606627)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606629)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606630)
	}
	__antithesis_instrumentation__.Notify(606628)
	return res
}

func (d *DJSON) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606631)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606635)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606636)
	}
	__antithesis_instrumentation__.Notify(606632)
	v, ok := UnwrapDatum(ctx, other).(*DJSON)
	if !ok {
		__antithesis_instrumentation__.Notify(606637)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606638)
	}
	__antithesis_instrumentation__.Notify(606633)
	c, err := d.JSON.Compare(v.JSON)
	if err != nil {
		__antithesis_instrumentation__.Notify(606639)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(606640)
	}
	__antithesis_instrumentation__.Notify(606634)
	return c, nil
}

func (d *DJSON) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606641)
	return nil, false
}

func (d *DJSON) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606642)
	return nil, false
}

func (d *DJSON) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606643)
	return false
}

func (d *DJSON) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606644)
	return d.JSON == json.NullJSONValue
}

func (d *DJSON) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606645)
	return nil, false
}

func (d *DJSON) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606646)
	return &DJSON{json.NullJSONValue}, true
}

func (*DJSON) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606647); return true }

func (d *DJSON) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606648)

	s := d.JSON.String()
	if ctx.flags.HasFlags(fmtRawStrings) {
		__antithesis_instrumentation__.Notify(606649)
		ctx.WriteString(s)
	} else {
		__antithesis_instrumentation__.Notify(606650)

		lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, s, ctx.flags.EncodeFlags())
	}
}

func (d *DJSON) Size() uintptr {
	__antithesis_instrumentation__.Notify(606651)
	return unsafe.Sizeof(*d) + d.JSON.Size()
}

type DTuple struct {
	D Datums

	sorted bool

	typ *types.T
}

func NewDTuple(typ *types.T, d ...Datum) *DTuple {
	__antithesis_instrumentation__.Notify(606652)
	return &DTuple{D: d, typ: typ}
}

func NewDTupleWithLen(typ *types.T, l int) *DTuple {
	__antithesis_instrumentation__.Notify(606653)
	return &DTuple{D: make(Datums, l), typ: typ}
}

func MakeDTuple(typ *types.T, d ...Datum) DTuple {
	__antithesis_instrumentation__.Notify(606654)
	return DTuple{D: d, typ: typ}
}

func AsDTuple(e Expr) (*DTuple, bool) {
	__antithesis_instrumentation__.Notify(606655)
	switch t := e.(type) {
	case *DTuple:
		__antithesis_instrumentation__.Notify(606657)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606658)
		return AsDTuple(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606656)
	return nil, false
}

func MustBeDTuple(e Expr) *DTuple {
	__antithesis_instrumentation__.Notify(606659)
	i, ok := AsDTuple(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606661)
		panic(errors.AssertionFailedf("expected *DTuple, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606662)
	}
	__antithesis_instrumentation__.Notify(606660)
	return i
}

func (d *DTuple) maybePopulateType() {
	__antithesis_instrumentation__.Notify(606663)
	if d.typ == nil {
		__antithesis_instrumentation__.Notify(606664)
		contents := make([]*types.T, len(d.D))
		for i, v := range d.D {
			__antithesis_instrumentation__.Notify(606666)
			contents[i] = v.ResolvedType()
		}
		__antithesis_instrumentation__.Notify(606665)
		d.typ = types.MakeTuple(contents)
	} else {
		__antithesis_instrumentation__.Notify(606667)
	}
}

func (d *DTuple) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606668)
	d.maybePopulateType()
	return d.typ
}

func (d *DTuple) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606669)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606671)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606672)
	}
	__antithesis_instrumentation__.Notify(606670)
	return res
}

func (d *DTuple) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606673)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606680)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606681)
	}
	__antithesis_instrumentation__.Notify(606674)
	v, ok := UnwrapDatum(ctx, other).(*DTuple)
	if !ok {
		__antithesis_instrumentation__.Notify(606682)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606683)
	}
	__antithesis_instrumentation__.Notify(606675)
	n := len(d.D)
	if n > len(v.D) {
		__antithesis_instrumentation__.Notify(606684)
		n = len(v.D)
	} else {
		__antithesis_instrumentation__.Notify(606685)
	}
	__antithesis_instrumentation__.Notify(606676)
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(606686)
		c, err := d.D[i].CompareError(ctx, v.D[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(606688)
			return 0, errors.WithDetailf(err, "type mismatch at record column %d", redact.SafeInt(i+1))
		} else {
			__antithesis_instrumentation__.Notify(606689)
		}
		__antithesis_instrumentation__.Notify(606687)
		if c != 0 {
			__antithesis_instrumentation__.Notify(606690)
			return c, nil
		} else {
			__antithesis_instrumentation__.Notify(606691)
		}
	}
	__antithesis_instrumentation__.Notify(606677)
	if len(d.D) < len(v.D) {
		__antithesis_instrumentation__.Notify(606692)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(606693)
	}
	__antithesis_instrumentation__.Notify(606678)
	if len(d.D) > len(v.D) {
		__antithesis_instrumentation__.Notify(606694)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606695)
	}
	__antithesis_instrumentation__.Notify(606679)
	return 0, nil
}

func (d *DTuple) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606696)

	res := NewDTupleWithLen(d.typ, len(d.D))
	copy(res.D, d.D)
	for i := len(res.D) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(606698)
		if !res.D[i].IsMin(ctx) {
			__antithesis_instrumentation__.Notify(606701)
			prevVal, ok := res.D[i].Prev(ctx)
			if !ok {
				__antithesis_instrumentation__.Notify(606703)
				return nil, false
			} else {
				__antithesis_instrumentation__.Notify(606704)
			}
			__antithesis_instrumentation__.Notify(606702)
			res.D[i] = prevVal
			break
		} else {
			__antithesis_instrumentation__.Notify(606705)
		}
		__antithesis_instrumentation__.Notify(606699)
		maxVal, ok := res.D[i].Max(ctx)
		if !ok {
			__antithesis_instrumentation__.Notify(606706)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(606707)
		}
		__antithesis_instrumentation__.Notify(606700)
		res.D[i] = maxVal
	}
	__antithesis_instrumentation__.Notify(606697)
	return res, true
}

func (d *DTuple) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606708)

	res := NewDTupleWithLen(d.typ, len(d.D))
	copy(res.D, d.D)
	for i := len(res.D) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(606710)
		if !res.D[i].IsMax(ctx) {
			__antithesis_instrumentation__.Notify(606712)
			nextVal, ok := res.D[i].Next(ctx)
			if !ok {
				__antithesis_instrumentation__.Notify(606714)
				return nil, false
			} else {
				__antithesis_instrumentation__.Notify(606715)
			}
			__antithesis_instrumentation__.Notify(606713)
			res.D[i] = nextVal
			break
		} else {
			__antithesis_instrumentation__.Notify(606716)
		}
		__antithesis_instrumentation__.Notify(606711)

		res.D[i] = DNull
	}
	__antithesis_instrumentation__.Notify(606709)
	return res, true
}

func (d *DTuple) Max(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606717)
	res := NewDTupleWithLen(d.typ, len(d.D))
	for i, v := range d.D {
		__antithesis_instrumentation__.Notify(606719)
		m, ok := v.Max(ctx)
		if !ok {
			__antithesis_instrumentation__.Notify(606721)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(606722)
		}
		__antithesis_instrumentation__.Notify(606720)
		res.D[i] = m
	}
	__antithesis_instrumentation__.Notify(606718)
	return res, true
}

func (d *DTuple) Min(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606723)
	res := NewDTupleWithLen(d.typ, len(d.D))
	for i, v := range d.D {
		__antithesis_instrumentation__.Notify(606725)
		m, ok := v.Min(ctx)
		if !ok {
			__antithesis_instrumentation__.Notify(606727)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(606728)
		}
		__antithesis_instrumentation__.Notify(606726)
		res.D[i] = m
	}
	__antithesis_instrumentation__.Notify(606724)
	return res, true
}

func (d *DTuple) IsMax(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606729)
	for _, v := range d.D {
		__antithesis_instrumentation__.Notify(606731)
		if !v.IsMax(ctx) {
			__antithesis_instrumentation__.Notify(606732)
			return false
		} else {
			__antithesis_instrumentation__.Notify(606733)
		}
	}
	__antithesis_instrumentation__.Notify(606730)
	return true
}

func (d *DTuple) IsMin(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606734)
	for _, v := range d.D {
		__antithesis_instrumentation__.Notify(606736)
		if !v.IsMin(ctx) {
			__antithesis_instrumentation__.Notify(606737)
			return false
		} else {
			__antithesis_instrumentation__.Notify(606738)
		}
	}
	__antithesis_instrumentation__.Notify(606735)
	return true
}

func (*DTuple) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606739); return false }

func (d *DTuple) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606740)
	if ctx.HasFlags(fmtPgwireFormat) {
		__antithesis_instrumentation__.Notify(606745)
		d.pgwireFormat(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(606746)
	}
	__antithesis_instrumentation__.Notify(606741)

	typ := d.ResolvedType()
	tupleContents := typ.TupleContents()
	showLabels := len(typ.TupleLabels()) > 0
	if showLabels {
		__antithesis_instrumentation__.Notify(606747)
		ctx.WriteByte('(')
	} else {
		__antithesis_instrumentation__.Notify(606748)
	}
	__antithesis_instrumentation__.Notify(606742)
	ctx.WriteByte('(')
	comma := ""
	parsable := ctx.HasFlags(FmtParsable)
	for i, v := range d.D {
		__antithesis_instrumentation__.Notify(606749)
		ctx.WriteString(comma)
		ctx.FormatNode(v)
		if parsable && func() bool {
			__antithesis_instrumentation__.Notify(606751)
			return (v == DNull) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(606752)
			return len(tupleContents) > i == true
		}() == true {
			__antithesis_instrumentation__.Notify(606753)

			if tupleContents[i].Family() != types.UnknownFamily {
				__antithesis_instrumentation__.Notify(606754)
				nullType := tupleContents[i]
				if ctx.HasFlags(fmtDisambiguateDatumTypes) {
					__antithesis_instrumentation__.Notify(606755)
					ctx.WriteString(":::")
					ctx.FormatTypeReference(nullType)
				} else {
					__antithesis_instrumentation__.Notify(606756)
					ctx.WriteString("::")
					ctx.WriteString(nullType.SQLString())
				}
			} else {
				__antithesis_instrumentation__.Notify(606757)
			}
		} else {
			__antithesis_instrumentation__.Notify(606758)
		}
		__antithesis_instrumentation__.Notify(606750)
		comma = ", "
	}
	__antithesis_instrumentation__.Notify(606743)
	if len(d.D) == 1 {
		__antithesis_instrumentation__.Notify(606759)

		ctx.WriteByte(',')
	} else {
		__antithesis_instrumentation__.Notify(606760)
	}
	__antithesis_instrumentation__.Notify(606744)
	ctx.WriteByte(')')
	if showLabels {
		__antithesis_instrumentation__.Notify(606761)
		ctx.WriteString(" AS ")
		comma := ""
		for i := range typ.TupleLabels() {
			__antithesis_instrumentation__.Notify(606763)
			ctx.WriteString(comma)
			ctx.FormatNode((*Name)(&typ.TupleLabels()[i]))
			comma = ", "
		}
		__antithesis_instrumentation__.Notify(606762)
		ctx.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(606764)
	}
}

func (d *DTuple) Sorted() bool {
	__antithesis_instrumentation__.Notify(606765)
	return d.sorted
}

func (d *DTuple) SetSorted() *DTuple {
	__antithesis_instrumentation__.Notify(606766)
	if d.ContainsNull() {
		__antithesis_instrumentation__.Notify(606768)

		return d
	} else {
		__antithesis_instrumentation__.Notify(606769)
	}
	__antithesis_instrumentation__.Notify(606767)
	d.sorted = true
	return d
}

func (d *DTuple) AssertSorted() {
	__antithesis_instrumentation__.Notify(606770)
	if !d.sorted {
		__antithesis_instrumentation__.Notify(606771)
		panic(errors.AssertionFailedf("expected sorted tuple, found %#v", d))
	} else {
		__antithesis_instrumentation__.Notify(606772)
	}
}

func (d *DTuple) SearchSorted(ctx *EvalContext, target Datum) (int, bool) {
	__antithesis_instrumentation__.Notify(606773)
	d.AssertSorted()
	if target == DNull {
		__antithesis_instrumentation__.Notify(606777)
		panic(errors.AssertionFailedf("NULL target (d: %s)", d))
	} else {
		__antithesis_instrumentation__.Notify(606778)
	}
	__antithesis_instrumentation__.Notify(606774)
	if t, ok := target.(*DTuple); ok && func() bool {
		__antithesis_instrumentation__.Notify(606779)
		return t.ContainsNull() == true
	}() == true {
		__antithesis_instrumentation__.Notify(606780)
		panic(errors.AssertionFailedf("target containing NULLs: %#v (d: %s)", target, d))
	} else {
		__antithesis_instrumentation__.Notify(606781)
	}
	__antithesis_instrumentation__.Notify(606775)
	i := sort.Search(len(d.D), func(i int) bool {
		__antithesis_instrumentation__.Notify(606782)
		return d.D[i].Compare(ctx, target) >= 0
	})
	__antithesis_instrumentation__.Notify(606776)
	found := i < len(d.D) && func() bool {
		__antithesis_instrumentation__.Notify(606783)
		return d.D[i].Compare(ctx, target) == 0 == true
	}() == true
	return i, found
}

func (d *DTuple) Normalize(ctx *EvalContext) {
	__antithesis_instrumentation__.Notify(606784)
	d.sort(ctx)
	d.makeUnique(ctx)
}

func (d *DTuple) sort(ctx *EvalContext) {
	__antithesis_instrumentation__.Notify(606785)
	if !d.sorted {
		__antithesis_instrumentation__.Notify(606786)
		lessFn := func(i, j int) bool {
			__antithesis_instrumentation__.Notify(606789)
			return d.D[i].Compare(ctx, d.D[j]) < 0
		}
		__antithesis_instrumentation__.Notify(606787)

		if !sort.SliceIsSorted(d.D, lessFn) {
			__antithesis_instrumentation__.Notify(606790)
			sort.Slice(d.D, lessFn)
		} else {
			__antithesis_instrumentation__.Notify(606791)
		}
		__antithesis_instrumentation__.Notify(606788)
		d.SetSorted()
	} else {
		__antithesis_instrumentation__.Notify(606792)
	}
}

func (d *DTuple) makeUnique(ctx *EvalContext) {
	__antithesis_instrumentation__.Notify(606793)
	n := 0
	for i := 0; i < len(d.D); i++ {
		__antithesis_instrumentation__.Notify(606795)
		if n == 0 || func() bool {
			__antithesis_instrumentation__.Notify(606796)
			return d.D[n-1].Compare(ctx, d.D[i]) < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(606797)
			d.D[n] = d.D[i]
			n++
		} else {
			__antithesis_instrumentation__.Notify(606798)
		}
	}
	__antithesis_instrumentation__.Notify(606794)
	d.D = d.D[:n]
}

func (d *DTuple) Size() uintptr {
	__antithesis_instrumentation__.Notify(606799)
	sz := unsafe.Sizeof(*d)
	for _, e := range d.D {
		__antithesis_instrumentation__.Notify(606801)
		dsz := e.Size()
		sz += dsz
	}
	__antithesis_instrumentation__.Notify(606800)
	return sz
}

func (d *DTuple) ContainsNull() bool {
	__antithesis_instrumentation__.Notify(606802)
	for _, r := range d.D {
		__antithesis_instrumentation__.Notify(606804)
		if r == DNull {
			__antithesis_instrumentation__.Notify(606806)
			return true
		} else {
			__antithesis_instrumentation__.Notify(606807)
		}
		__antithesis_instrumentation__.Notify(606805)
		if t, ok := r.(*DTuple); ok {
			__antithesis_instrumentation__.Notify(606808)
			if t.ContainsNull() {
				__antithesis_instrumentation__.Notify(606809)
				return true
			} else {
				__antithesis_instrumentation__.Notify(606810)
			}
		} else {
			__antithesis_instrumentation__.Notify(606811)
		}
	}
	__antithesis_instrumentation__.Notify(606803)
	return false
}

type dNull struct{}

func (dNull) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606812)
	return types.Unknown
}

func (d dNull) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606813)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606815)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606816)
	}
	__antithesis_instrumentation__.Notify(606814)
	return res
}

func (d dNull) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606817)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606819)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(606820)
	}
	__antithesis_instrumentation__.Notify(606818)
	return -1, nil
}

func (d dNull) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606821)
	return nil, false
}

func (d dNull) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606822)
	return nil, false
}

func (dNull) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606823)
	return true
}

func (dNull) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606824)
	return true
}

func (dNull) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606825)
	return DNull, true
}

func (dNull) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606826)
	return DNull, true
}

func (dNull) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606827); return false }

func (dNull) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606828)
	if ctx.HasFlags(fmtPgwireFormat) {
		__antithesis_instrumentation__.Notify(606830)

		return
	} else {
		__antithesis_instrumentation__.Notify(606831)
	}
	__antithesis_instrumentation__.Notify(606829)
	ctx.WriteString("NULL")
}

func (d dNull) Size() uintptr {
	__antithesis_instrumentation__.Notify(606832)
	return unsafe.Sizeof(d)
}

type DArray struct {
	ParamTyp *types.T
	Array    Datums

	HasNulls bool

	HasNonNulls bool

	customOid oid.Oid
}

func NewDArray(paramTyp *types.T) *DArray {
	__antithesis_instrumentation__.Notify(606833)
	return &DArray{ParamTyp: paramTyp}
}

func AsDArray(e Expr) (*DArray, bool) {
	__antithesis_instrumentation__.Notify(606834)
	switch t := e.(type) {
	case *DArray:
		__antithesis_instrumentation__.Notify(606836)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(606837)
		return AsDArray(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(606835)
	return nil, false
}

func MustBeDArray(e Expr) *DArray {
	__antithesis_instrumentation__.Notify(606838)
	i, ok := AsDArray(e)
	if !ok {
		__antithesis_instrumentation__.Notify(606840)
		panic(errors.AssertionFailedf("expected *DArray, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(606841)
	}
	__antithesis_instrumentation__.Notify(606839)
	return i
}

func (d *DArray) MaybeSetCustomOid(t *types.T) error {
	__antithesis_instrumentation__.Notify(606842)
	if t.Family() != types.ArrayFamily {
		__antithesis_instrumentation__.Notify(606845)
		return errors.AssertionFailedf("expected array type, got %s", t.SQLString())
	} else {
		__antithesis_instrumentation__.Notify(606846)
	}
	__antithesis_instrumentation__.Notify(606843)
	switch t.Oid() {
	case oid.T_int2vector:
		__antithesis_instrumentation__.Notify(606847)
		d.customOid = oid.T_int2vector
	case oid.T_oidvector:
		__antithesis_instrumentation__.Notify(606848)
		d.customOid = oid.T_oidvector
	default:
		__antithesis_instrumentation__.Notify(606849)
	}
	__antithesis_instrumentation__.Notify(606844)
	return nil
}

func (d *DArray) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606850)
	switch d.customOid {
	case oid.T_int2vector:
		__antithesis_instrumentation__.Notify(606852)
		return types.Int2Vector
	case oid.T_oidvector:
		__antithesis_instrumentation__.Notify(606853)
		return types.OidVector
	default:
		__antithesis_instrumentation__.Notify(606854)
	}
	__antithesis_instrumentation__.Notify(606851)
	return types.MakeArray(d.ParamTyp)
}

func (d *DArray) IsComposite() bool {
	__antithesis_instrumentation__.Notify(606855)
	for _, elem := range d.Array {
		__antithesis_instrumentation__.Notify(606857)
		if cdatum, ok := elem.(CompositeDatum); ok && func() bool {
			__antithesis_instrumentation__.Notify(606858)
			return cdatum.IsComposite() == true
		}() == true {
			__antithesis_instrumentation__.Notify(606859)
			return true
		} else {
			__antithesis_instrumentation__.Notify(606860)
		}
	}
	__antithesis_instrumentation__.Notify(606856)
	return false
}

func (d *DArray) FirstIndex() int {
	__antithesis_instrumentation__.Notify(606861)
	switch d.customOid {
	case oid.T_int2vector, oid.T_oidvector:
		__antithesis_instrumentation__.Notify(606863)
		return 0
	default:
		__antithesis_instrumentation__.Notify(606864)
	}
	__antithesis_instrumentation__.Notify(606862)
	return 1
}

func (d *DArray) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606865)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606867)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606868)
	}
	__antithesis_instrumentation__.Notify(606866)
	return res
}

func (d *DArray) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606869)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606876)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606877)
	}
	__antithesis_instrumentation__.Notify(606870)
	v, ok := UnwrapDatum(ctx, other).(*DArray)
	if !ok {
		__antithesis_instrumentation__.Notify(606878)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606879)
	}
	__antithesis_instrumentation__.Notify(606871)
	n := d.Len()
	if n > v.Len() {
		__antithesis_instrumentation__.Notify(606880)
		n = v.Len()
	} else {
		__antithesis_instrumentation__.Notify(606881)
	}
	__antithesis_instrumentation__.Notify(606872)
	for i := 0; i < n; i++ {
		__antithesis_instrumentation__.Notify(606882)
		c, err := d.Array[i].CompareError(ctx, v.Array[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(606884)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(606885)
		}
		__antithesis_instrumentation__.Notify(606883)
		if c != 0 {
			__antithesis_instrumentation__.Notify(606886)
			return c, nil
		} else {
			__antithesis_instrumentation__.Notify(606887)
		}
	}
	__antithesis_instrumentation__.Notify(606873)
	if d.Len() < v.Len() {
		__antithesis_instrumentation__.Notify(606888)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(606889)
	}
	__antithesis_instrumentation__.Notify(606874)
	if d.Len() > v.Len() {
		__antithesis_instrumentation__.Notify(606890)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606891)
	}
	__antithesis_instrumentation__.Notify(606875)
	return 0, nil
}

func (d *DArray) Prev(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606892)
	return nil, false
}

func (d *DArray) Next(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606893)
	a := DArray{ParamTyp: d.ParamTyp, Array: make(Datums, d.Len()+1)}
	copy(a.Array, d.Array)
	a.Array[len(a.Array)-1] = DNull
	return &a, true
}

func (d *DArray) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606894)
	return nil, false
}

func (d *DArray) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606895)
	return &DArray{ParamTyp: d.ParamTyp}, true
}

func (d *DArray) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606896)
	return false
}

func (d *DArray) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606897)
	return d.Len() == 0
}

func (d *DArray) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(606898)

	if d.ParamTyp.Family() == types.UnknownFamily {
		__antithesis_instrumentation__.Notify(606900)

		return false
	} else {
		__antithesis_instrumentation__.Notify(606901)
	}
	__antithesis_instrumentation__.Notify(606899)
	return !d.HasNonNulls
}

func (d *DArray) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606902)
	if ctx.flags.HasAnyFlags(fmtPgwireFormat | FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(606906)
		d.pgwireFormat(ctx)
		return
	} else {
		__antithesis_instrumentation__.Notify(606907)
	}
	__antithesis_instrumentation__.Notify(606903)

	if ctx.HasFlags(FmtExport) {
		__antithesis_instrumentation__.Notify(606908)
		oldFlags := ctx.flags
		ctx.flags = oldFlags & ^FmtExport | FmtParsable
		defer func() { __antithesis_instrumentation__.Notify(606909); ctx.flags = oldFlags }()
	} else {
		__antithesis_instrumentation__.Notify(606910)
	}
	__antithesis_instrumentation__.Notify(606904)

	ctx.WriteString("ARRAY[")
	comma := ""
	for _, v := range d.Array {
		__antithesis_instrumentation__.Notify(606911)
		ctx.WriteString(comma)
		ctx.FormatNode(v)
		comma = ","
	}
	__antithesis_instrumentation__.Notify(606905)
	ctx.WriteByte(']')
}

const maxArrayLength = math.MaxInt32

var errArrayTooLongError = errors.New("ARRAYs can be at most 2^31-1 elements long")

func (d *DArray) Validate() error {
	__antithesis_instrumentation__.Notify(606912)
	if d.Len() > maxArrayLength {
		__antithesis_instrumentation__.Notify(606914)
		return errors.WithStack(errArrayTooLongError)
	} else {
		__antithesis_instrumentation__.Notify(606915)
	}
	__antithesis_instrumentation__.Notify(606913)
	return nil
}

func (d *DArray) Len() int {
	__antithesis_instrumentation__.Notify(606916)
	return len(d.Array)
}

func (d *DArray) Size() uintptr {
	__antithesis_instrumentation__.Notify(606917)
	sz := unsafe.Sizeof(*d)
	for _, e := range d.Array {
		__antithesis_instrumentation__.Notify(606919)
		dsz := e.Size()
		sz += dsz
	}
	__antithesis_instrumentation__.Notify(606918)
	return sz
}

var errNonHomogeneousArray = pgerror.New(pgcode.ArraySubscript, "multidimensional arrays must have array expressions with matching dimensions")

func (d *DArray) Append(v Datum) error {
	__antithesis_instrumentation__.Notify(606920)

	if !v.ResolvedType().EquivalentOrNull(d.ParamTyp, true) {
		__antithesis_instrumentation__.Notify(606925)
		return errors.AssertionFailedf("cannot append %s to array containing %s", v.ResolvedType(), d.ParamTyp)
	} else {
		__antithesis_instrumentation__.Notify(606926)
	}
	__antithesis_instrumentation__.Notify(606921)
	if d.Len() >= maxArrayLength {
		__antithesis_instrumentation__.Notify(606927)
		return errors.WithStack(errArrayTooLongError)
	} else {
		__antithesis_instrumentation__.Notify(606928)
	}
	__antithesis_instrumentation__.Notify(606922)
	if d.ParamTyp.Family() == types.ArrayFamily {
		__antithesis_instrumentation__.Notify(606929)
		if v == DNull {
			__antithesis_instrumentation__.Notify(606931)
			return errNonHomogeneousArray
		} else {
			__antithesis_instrumentation__.Notify(606932)
		}
		__antithesis_instrumentation__.Notify(606930)
		if d.Len() > 0 {
			__antithesis_instrumentation__.Notify(606933)
			prevItem := d.Array[d.Len()-1]
			if prevItem == DNull {
				__antithesis_instrumentation__.Notify(606935)
				return errNonHomogeneousArray
			} else {
				__antithesis_instrumentation__.Notify(606936)
			}
			__antithesis_instrumentation__.Notify(606934)
			expectedLen := MustBeDArray(prevItem).Len()
			if MustBeDArray(v).Len() != expectedLen {
				__antithesis_instrumentation__.Notify(606937)
				return errNonHomogeneousArray
			} else {
				__antithesis_instrumentation__.Notify(606938)
			}
		} else {
			__antithesis_instrumentation__.Notify(606939)
		}
	} else {
		__antithesis_instrumentation__.Notify(606940)
	}
	__antithesis_instrumentation__.Notify(606923)
	if v == DNull {
		__antithesis_instrumentation__.Notify(606941)
		d.HasNulls = true
	} else {
		__antithesis_instrumentation__.Notify(606942)
		d.HasNonNulls = true
	}
	__antithesis_instrumentation__.Notify(606924)
	d.Array = append(d.Array, v)
	return d.Validate()
}

type DVoid struct{}

var DVoidDatum = &DVoid{}

func (*DVoid) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606943)
	return types.Void
}

func (d *DVoid) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606944)
	ret, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(606946)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(606947)
	}
	__antithesis_instrumentation__.Notify(606945)
	return ret
}

func (d *DVoid) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(606948)
	if other == DNull {
		__antithesis_instrumentation__.Notify(606951)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(606952)
	}
	__antithesis_instrumentation__.Notify(606949)

	_, ok := UnwrapDatum(ctx, other).(*DVoid)
	if !ok {
		__antithesis_instrumentation__.Notify(606953)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(606954)
	}
	__antithesis_instrumentation__.Notify(606950)
	return 0, nil
}

func (d *DVoid) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606955)
	return nil, false
}

func (d *DVoid) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606956)
	return nil, false
}

func (d *DVoid) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606957)
	return false
}

func (d *DVoid) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(606958)
	return false
}

func (d *DVoid) Max(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606959)
	return nil, false
}

func (d *DVoid) Min(_ *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(606960)
	return nil, false
}

func (*DVoid) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(606961); return true }

func (d *DVoid) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606962)
	buf, f := &ctx.Buffer, ctx.flags
	if !f.HasFlags(fmtRawStrings) {
		__antithesis_instrumentation__.Notify(606963)

		lexbase.EncodeSQLStringWithFlags(buf, "", f.EncodeFlags())
	} else {
		__antithesis_instrumentation__.Notify(606964)
	}
}

func (d *DVoid) Size() uintptr {
	__antithesis_instrumentation__.Notify(606965)
	return unsafe.Sizeof(*d)
}

type DEnum struct {
	EnumTyp *types.T

	PhysicalRep []byte

	LogicalRep string
}

func (d *DEnum) Size() uintptr {
	__antithesis_instrumentation__.Notify(606966)

	return unsafe.Sizeof(d.EnumTyp) +
		unsafe.Sizeof(d.PhysicalRep) +
		unsafe.Sizeof(d.LogicalRep)
}

func GetEnumComponentsFromPhysicalRep(typ *types.T, rep []byte) ([]byte, string, error) {
	__antithesis_instrumentation__.Notify(606967)
	idx, err := typ.EnumGetIdxOfPhysical(rep)
	if err != nil {
		__antithesis_instrumentation__.Notify(606969)
		return nil, "", err
	} else {
		__antithesis_instrumentation__.Notify(606970)
	}
	__antithesis_instrumentation__.Notify(606968)
	meta := typ.TypeMeta.EnumData

	return meta.PhysicalRepresentations[idx], meta.LogicalRepresentations[idx], nil
}

func MakeDEnumFromPhysicalRepresentation(typ *types.T, rep []byte) (*DEnum, error) {
	__antithesis_instrumentation__.Notify(606971)

	if typ.Oid() == oid.T_anyenum {
		__antithesis_instrumentation__.Notify(606974)
		return nil, errors.New("cannot create enum of unspecified type")
	} else {
		__antithesis_instrumentation__.Notify(606975)
	}
	__antithesis_instrumentation__.Notify(606972)
	phys, log, err := GetEnumComponentsFromPhysicalRep(typ, rep)
	if err != nil {
		__antithesis_instrumentation__.Notify(606976)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(606977)
	}
	__antithesis_instrumentation__.Notify(606973)
	return &DEnum{
		EnumTyp:     typ,
		PhysicalRep: phys,
		LogicalRep:  log,
	}, nil
}

func MakeDEnumFromLogicalRepresentation(typ *types.T, rep string) (*DEnum, error) {
	__antithesis_instrumentation__.Notify(606978)

	if typ.Oid() == oid.T_anyenum {
		__antithesis_instrumentation__.Notify(606982)
		return nil, errors.New("cannot create enum of unspecified type")
	} else {
		__antithesis_instrumentation__.Notify(606983)
	}
	__antithesis_instrumentation__.Notify(606979)

	idx, err := typ.EnumGetIdxOfLogical(rep)
	if err != nil {
		__antithesis_instrumentation__.Notify(606984)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(606985)
	}
	__antithesis_instrumentation__.Notify(606980)

	if typ.TypeMeta.EnumData.IsMemberReadOnly[idx] {
		__antithesis_instrumentation__.Notify(606986)
		return nil, errors.Newf("enum value %q is not yet public", rep)
	} else {
		__antithesis_instrumentation__.Notify(606987)
	}
	__antithesis_instrumentation__.Notify(606981)
	return &DEnum{
		EnumTyp:     typ,
		PhysicalRep: typ.TypeMeta.EnumData.PhysicalRepresentations[idx],
		LogicalRep:  typ.TypeMeta.EnumData.LogicalRepresentations[idx],
	}, nil
}

func MakeAllDEnumsInType(typ *types.T) []Datum {
	__antithesis_instrumentation__.Notify(606988)
	result := make([]Datum, len(typ.TypeMeta.EnumData.LogicalRepresentations))
	for i := 0; i < len(result); i++ {
		__antithesis_instrumentation__.Notify(606990)
		result[i] = &DEnum{
			EnumTyp:     typ,
			PhysicalRep: typ.TypeMeta.EnumData.PhysicalRepresentations[i],
			LogicalRep:  typ.TypeMeta.EnumData.LogicalRepresentations[i],
		}
	}
	__antithesis_instrumentation__.Notify(606989)
	return result
}

func (d *DEnum) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(606991)
	if ctx.HasFlags(fmtStaticallyFormatUserDefinedTypes) {
		__antithesis_instrumentation__.Notify(606992)
		s := DBytes(d.PhysicalRep)

		ctx.WithFlags(ctx.flags|fmtFormatByteLiterals, func() {
			__antithesis_instrumentation__.Notify(606993)
			s.Format(ctx)
		})
	} else {
		__antithesis_instrumentation__.Notify(606994)
		if ctx.HasFlags(FmtPgwireText) {
			__antithesis_instrumentation__.Notify(606995)
			ctx.WriteString(d.LogicalRep)
		} else {
			__antithesis_instrumentation__.Notify(606996)
			s := DString(d.LogicalRep)
			s.Format(ctx)
		}
	}
}

func (d *DEnum) String() string {
	__antithesis_instrumentation__.Notify(606997)
	return AsString(d)
}

func (d *DEnum) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(606998)
	return d.EnumTyp
}

func (d *DEnum) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(606999)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(607001)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607002)
	}
	__antithesis_instrumentation__.Notify(607000)
	return res
}

func (d *DEnum) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(607003)
	if other == DNull {
		__antithesis_instrumentation__.Notify(607006)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(607007)
	}
	__antithesis_instrumentation__.Notify(607004)
	v, ok := UnwrapDatum(ctx, other).(*DEnum)
	if !ok {
		__antithesis_instrumentation__.Notify(607008)
		return 0, makeUnsupportedComparisonMessage(d, other)
	} else {
		__antithesis_instrumentation__.Notify(607009)
	}
	__antithesis_instrumentation__.Notify(607005)
	res := bytes.Compare(d.PhysicalRep, v.PhysicalRep)
	return res, nil
}

func (d *DEnum) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607010)
	idx, err := d.EnumTyp.EnumGetIdxOfPhysical(d.PhysicalRep)
	if err != nil {
		__antithesis_instrumentation__.Notify(607013)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607014)
	}
	__antithesis_instrumentation__.Notify(607011)
	if idx == 0 {
		__antithesis_instrumentation__.Notify(607015)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(607016)
	}
	__antithesis_instrumentation__.Notify(607012)
	enumData := d.EnumTyp.TypeMeta.EnumData
	return &DEnum{
		EnumTyp:     d.EnumTyp,
		PhysicalRep: enumData.PhysicalRepresentations[idx-1],
		LogicalRep:  enumData.LogicalRepresentations[idx-1],
	}, true
}

func (d *DEnum) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607017)
	idx, err := d.EnumTyp.EnumGetIdxOfPhysical(d.PhysicalRep)
	if err != nil {
		__antithesis_instrumentation__.Notify(607020)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607021)
	}
	__antithesis_instrumentation__.Notify(607018)
	enumData := d.EnumTyp.TypeMeta.EnumData
	if idx == len(enumData.PhysicalRepresentations)-1 {
		__antithesis_instrumentation__.Notify(607022)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(607023)
	}
	__antithesis_instrumentation__.Notify(607019)
	return &DEnum{
		EnumTyp:     d.EnumTyp,
		PhysicalRep: enumData.PhysicalRepresentations[idx+1],
		LogicalRep:  enumData.LogicalRepresentations[idx+1],
	}, true
}

func (d *DEnum) Max(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607024)
	enumData := d.EnumTyp.TypeMeta.EnumData
	if len(enumData.PhysicalRepresentations) == 0 {
		__antithesis_instrumentation__.Notify(607026)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(607027)
	}
	__antithesis_instrumentation__.Notify(607025)
	idx := len(enumData.PhysicalRepresentations) - 1
	return &DEnum{
		EnumTyp:     d.EnumTyp,
		PhysicalRep: enumData.PhysicalRepresentations[idx],
		LogicalRep:  enumData.LogicalRepresentations[idx],
	}, true
}

func (d *DEnum) Min(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607028)
	enumData := d.EnumTyp.TypeMeta.EnumData
	if len(enumData.PhysicalRepresentations) == 0 {
		__antithesis_instrumentation__.Notify(607030)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(607031)
	}
	__antithesis_instrumentation__.Notify(607029)
	return &DEnum{
		EnumTyp:     d.EnumTyp,
		PhysicalRep: enumData.PhysicalRepresentations[0],
		LogicalRep:  enumData.LogicalRepresentations[0],
	}, true
}

func (d *DEnum) IsMax(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607032)
	physReps := d.EnumTyp.TypeMeta.EnumData.PhysicalRepresentations
	idx, err := d.EnumTyp.EnumGetIdxOfPhysical(d.PhysicalRep)
	if err != nil {
		__antithesis_instrumentation__.Notify(607034)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607035)
	}
	__antithesis_instrumentation__.Notify(607033)
	return idx == len(physReps)-1
}

func (d *DEnum) IsMin(_ *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607036)
	idx, err := d.EnumTyp.EnumGetIdxOfPhysical(d.PhysicalRep)
	if err != nil {
		__antithesis_instrumentation__.Notify(607038)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607039)
	}
	__antithesis_instrumentation__.Notify(607037)
	return idx == 0
}

func (d *DEnum) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(607040)
	return true
}

func (d *DEnum) MaxWriteable() (Datum, bool) {
	__antithesis_instrumentation__.Notify(607041)
	enumData := d.EnumTyp.TypeMeta.EnumData
	if len(enumData.PhysicalRepresentations) == 0 {
		__antithesis_instrumentation__.Notify(607044)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(607045)
	}
	__antithesis_instrumentation__.Notify(607042)
	for i := len(enumData.PhysicalRepresentations) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(607046)
		if !enumData.IsMemberReadOnly[i] {
			__antithesis_instrumentation__.Notify(607047)
			return &DEnum{
				EnumTyp:     d.EnumTyp,
				PhysicalRep: enumData.PhysicalRepresentations[i],
				LogicalRep:  enumData.LogicalRepresentations[i],
			}, true
		} else {
			__antithesis_instrumentation__.Notify(607048)
		}
	}
	__antithesis_instrumentation__.Notify(607043)
	return nil, false
}

func (d *DEnum) MinWriteable() (Datum, bool) {
	__antithesis_instrumentation__.Notify(607049)
	enumData := d.EnumTyp.TypeMeta.EnumData
	if len(enumData.PhysicalRepresentations) == 0 {
		__antithesis_instrumentation__.Notify(607052)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(607053)
	}
	__antithesis_instrumentation__.Notify(607050)
	for i := 0; i < len(enumData.PhysicalRepresentations); i++ {
		__antithesis_instrumentation__.Notify(607054)
		if !enumData.IsMemberReadOnly[i] {
			__antithesis_instrumentation__.Notify(607055)
			return &DEnum{
				EnumTyp:     d.EnumTyp,
				PhysicalRep: enumData.PhysicalRepresentations[i],
				LogicalRep:  enumData.LogicalRepresentations[i],
			}, true
		} else {
			__antithesis_instrumentation__.Notify(607056)
		}
	}
	__antithesis_instrumentation__.Notify(607051)
	return nil, false
}

type DOid struct {
	DInt

	semanticType *types.T

	name string
}

func MakeDOid(d DInt) DOid {
	__antithesis_instrumentation__.Notify(607057)
	return DOid{DInt: d, semanticType: types.Oid, name: ""}
}

func NewDOid(d DInt) *DOid {
	__antithesis_instrumentation__.Notify(607058)
	oid := MakeDOid(d)
	return &oid
}

func ParseDOid(ctx *EvalContext, s string, t *types.T) (*DOid, error) {
	__antithesis_instrumentation__.Notify(607059)

	if val, err := ParseDInt(strings.TrimSpace(s)); err == nil {
		__antithesis_instrumentation__.Notify(607061)
		tmpOid := NewDOid(*val)
		oid, err := ctx.Planner.ResolveOIDFromOID(ctx.Ctx(), t, tmpOid)
		if err != nil {
			__antithesis_instrumentation__.Notify(607063)
			oid = tmpOid
			oid.semanticType = t
		} else {
			__antithesis_instrumentation__.Notify(607064)
		}
		__antithesis_instrumentation__.Notify(607062)
		return oid, nil
	} else {
		__antithesis_instrumentation__.Notify(607065)
	}
	__antithesis_instrumentation__.Notify(607060)

	switch t.Oid() {
	case oid.T_regproc, oid.T_regprocedure:
		__antithesis_instrumentation__.Notify(607066)

		s = pgSignatureRegexp.ReplaceAllString(s, "$1")

		substrs, err := splitIdentifierList(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(607082)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(607083)
		}
		__antithesis_instrumentation__.Notify(607067)
		if len(substrs) > 3 {
			__antithesis_instrumentation__.Notify(607084)

			return nil, pgerror.Newf(pgcode.Syntax,
				"invalid function name: %s", s)
		} else {
			__antithesis_instrumentation__.Notify(607085)
		}
		__antithesis_instrumentation__.Notify(607068)
		name := UnresolvedName{NumParts: len(substrs)}
		for i := 0; i < len(substrs); i++ {
			__antithesis_instrumentation__.Notify(607086)
			name.Parts[i] = substrs[len(substrs)-1-i]
		}
		__antithesis_instrumentation__.Notify(607069)
		funcDef, err := name.ResolveFunction(ctx.SessionData().SearchPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(607087)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(607088)
		}
		__antithesis_instrumentation__.Notify(607070)
		if len(funcDef.Definition) > 1 {
			__antithesis_instrumentation__.Notify(607089)
			return nil, pgerror.Newf(pgcode.AmbiguousAlias,
				"more than one function named '%s'", funcDef.Name)
		} else {
			__antithesis_instrumentation__.Notify(607090)
		}
		__antithesis_instrumentation__.Notify(607071)
		def := funcDef.Definition[0]
		overload, ok := def.(*Overload)
		if !ok {
			__antithesis_instrumentation__.Notify(607091)
			return nil, errors.AssertionFailedf("invalid non-overload regproc %s", funcDef.Name)
		} else {
			__antithesis_instrumentation__.Notify(607092)
		}
		__antithesis_instrumentation__.Notify(607072)
		return &DOid{semanticType: t, DInt: DInt(overload.Oid), name: funcDef.Name}, nil
	case oid.T_regtype:
		__antithesis_instrumentation__.Notify(607073)
		parsedTyp, err := ctx.Planner.GetTypeFromValidSQLSyntax(s)
		if err == nil {
			__antithesis_instrumentation__.Notify(607093)
			return &DOid{
				semanticType: t,
				DInt:         DInt(parsedTyp.Oid()),
				name:         parsedTyp.SQLStandardName(),
			}, nil
		} else {
			__antithesis_instrumentation__.Notify(607094)
		}
		__antithesis_instrumentation__.Notify(607074)

		s = strings.TrimSpace(s)
		if len(s) > 1 && func() bool {
			__antithesis_instrumentation__.Notify(607095)
			return s[0] == '"' == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(607096)
			return s[len(s)-1] == '"' == true
		}() == true {
			__antithesis_instrumentation__.Notify(607097)
			s = s[1 : len(s)-1]
		} else {
			__antithesis_instrumentation__.Notify(607098)
		}
		__antithesis_instrumentation__.Notify(607075)

		s = pgSignatureRegexp.ReplaceAllString(s, "$1")

		dOid, missingTypeErr := ctx.Planner.ResolveOIDFromString(ctx.Ctx(), t, NewDString(Name(s).Normalize()))
		if missingTypeErr == nil {
			__antithesis_instrumentation__.Notify(607099)
			return dOid, missingTypeErr
		} else {
			__antithesis_instrumentation__.Notify(607100)
		}
		__antithesis_instrumentation__.Notify(607076)

		switch s {

		case "trigger":
			__antithesis_instrumentation__.Notify(607101)
		default:
			__antithesis_instrumentation__.Notify(607102)
			return nil, missingTypeErr
		}
		__antithesis_instrumentation__.Notify(607077)
		return &DOid{
			semanticType: t,

			DInt: -1,
			name: s,
		}, nil

	case oid.T_regclass:
		__antithesis_instrumentation__.Notify(607078)
		tn, err := castStringToRegClassTableName(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(607103)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(607104)
		}
		__antithesis_instrumentation__.Notify(607079)
		id, err := ctx.Planner.ResolveTableName(ctx.Ctx(), &tn)
		if err != nil {
			__antithesis_instrumentation__.Notify(607105)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(607106)
		}
		__antithesis_instrumentation__.Notify(607080)
		return &DOid{
			semanticType: t,
			DInt:         DInt(id),
			name:         tn.ObjectName.String(),
		}, nil
	default:
		__antithesis_instrumentation__.Notify(607081)
		return ctx.Planner.ResolveOIDFromString(ctx.Ctx(), t, NewDString(s))
	}
}

func castStringToRegClassTableName(s string) (TableName, error) {
	__antithesis_instrumentation__.Notify(607107)
	components, err := splitIdentifierList(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(607112)
		return TableName{}, err
	} else {
		__antithesis_instrumentation__.Notify(607113)
	}
	__antithesis_instrumentation__.Notify(607108)

	if len(components) > 3 {
		__antithesis_instrumentation__.Notify(607114)
		return TableName{}, pgerror.Newf(
			pgcode.InvalidName,
			"too many components: %s",
			s,
		)
	} else {
		__antithesis_instrumentation__.Notify(607115)
	}
	__antithesis_instrumentation__.Notify(607109)
	var retComponents [3]string
	for i := 0; i < len(components); i++ {
		__antithesis_instrumentation__.Notify(607116)
		retComponents[len(components)-1-i] = components[i]
	}
	__antithesis_instrumentation__.Notify(607110)
	u, err := NewUnresolvedObjectName(
		len(components),
		retComponents,
		0,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(607117)
		return TableName{}, err
	} else {
		__antithesis_instrumentation__.Notify(607118)
	}
	__antithesis_instrumentation__.Notify(607111)
	return u.ToTableName(), nil
}

func splitIdentifierList(in string) ([]string, error) {
	__antithesis_instrumentation__.Notify(607119)
	var pos int
	var ret []string
	const separator = '.'

	for pos < len(in) {
		__antithesis_instrumentation__.Notify(607121)
		if isWhitespace(in[pos]) {
			__antithesis_instrumentation__.Notify(607127)
			pos++
			continue
		} else {
			__antithesis_instrumentation__.Notify(607128)
		}
		__antithesis_instrumentation__.Notify(607122)
		if in[pos] == '"' {
			__antithesis_instrumentation__.Notify(607129)
			var b strings.Builder

			for {
				__antithesis_instrumentation__.Notify(607131)
				pos++
				endIdx := strings.IndexByte(in[pos:], '"')
				if endIdx == -1 {
					__antithesis_instrumentation__.Notify(607134)
					return nil, pgerror.Newf(
						pgcode.InvalidName,
						`invalid name: unclosed ": %s`,
						in,
					)
				} else {
					__antithesis_instrumentation__.Notify(607135)
				}
				__antithesis_instrumentation__.Notify(607132)
				b.WriteString(in[pos : pos+endIdx])
				pos += endIdx + 1

				if pos == len(in) || func() bool {
					__antithesis_instrumentation__.Notify(607136)
					return in[pos] != '"' == true
				}() == true {
					__antithesis_instrumentation__.Notify(607137)
					break
				} else {
					__antithesis_instrumentation__.Notify(607138)
				}
				__antithesis_instrumentation__.Notify(607133)
				b.WriteByte('"')
			}
			__antithesis_instrumentation__.Notify(607130)
			ret = append(ret, b.String())
		} else {
			__antithesis_instrumentation__.Notify(607139)
			var b strings.Builder
			for pos < len(in) && func() bool {
				__antithesis_instrumentation__.Notify(607141)
				return in[pos] != separator == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(607142)
				return !isWhitespace(in[pos]) == true
			}() == true {
				__antithesis_instrumentation__.Notify(607143)
				b.WriteByte(in[pos])
				pos++
			}
			__antithesis_instrumentation__.Notify(607140)

			ret = append(ret, strings.ToLower(b.String()))
		}
		__antithesis_instrumentation__.Notify(607123)

		for pos < len(in) && func() bool {
			__antithesis_instrumentation__.Notify(607144)
			return isWhitespace(in[pos]) == true
		}() == true {
			__antithesis_instrumentation__.Notify(607145)
			pos++
		}
		__antithesis_instrumentation__.Notify(607124)

		if pos == len(in) {
			__antithesis_instrumentation__.Notify(607146)
			break
		} else {
			__antithesis_instrumentation__.Notify(607147)
		}
		__antithesis_instrumentation__.Notify(607125)

		if in[pos] != separator {
			__antithesis_instrumentation__.Notify(607148)
			return nil, pgerror.Newf(
				pgcode.InvalidName,
				"invalid name: expected separator %c: %s",
				separator,
				in,
			)
		} else {
			__antithesis_instrumentation__.Notify(607149)
		}
		__antithesis_instrumentation__.Notify(607126)

		pos++
	}
	__antithesis_instrumentation__.Notify(607120)

	return ret, nil
}

func isWhitespace(ch byte) bool {
	__antithesis_instrumentation__.Notify(607150)
	switch ch {
	case ' ', '\t', '\r', '\f', '\n':
		__antithesis_instrumentation__.Notify(607152)
		return true
	default:
		__antithesis_instrumentation__.Notify(607153)
	}
	__antithesis_instrumentation__.Notify(607151)
	return false
}

func AsDOid(e Expr) (*DOid, bool) {
	__antithesis_instrumentation__.Notify(607154)
	switch t := e.(type) {
	case *DOid:
		__antithesis_instrumentation__.Notify(607156)
		return t, true
	case *DOidWrapper:
		__antithesis_instrumentation__.Notify(607157)
		return AsDOid(t.Wrapped)
	}
	__antithesis_instrumentation__.Notify(607155)
	return NewDOid(0), false
}

func MustBeDOid(e Expr) *DOid {
	__antithesis_instrumentation__.Notify(607158)
	i, ok := AsDOid(e)
	if !ok {
		__antithesis_instrumentation__.Notify(607160)
		panic(errors.AssertionFailedf("expected *DOid, found %T", e))
	} else {
		__antithesis_instrumentation__.Notify(607161)
	}
	__antithesis_instrumentation__.Notify(607159)
	return i
}

func NewDOidWithName(d DInt, typ *types.T, name string) *DOid {
	__antithesis_instrumentation__.Notify(607162)
	return &DOid{
		DInt:         d,
		semanticType: typ,
		name:         name,
	}
}

func (d *DOid) AsRegProc(name string) *DOid {
	__antithesis_instrumentation__.Notify(607163)
	d.name = name
	d.semanticType = types.RegProc
	return d
}

func (*DOid) AmbiguousFormat() bool { __antithesis_instrumentation__.Notify(607164); return true }

func (d *DOid) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(607165)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(607167)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607168)
	}
	__antithesis_instrumentation__.Notify(607166)
	return res
}

func (d *DOid) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(607169)
	if other == DNull {
		__antithesis_instrumentation__.Notify(607174)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(607175)
	}
	__antithesis_instrumentation__.Notify(607170)
	var v DInt
	switch t := UnwrapDatum(ctx, other).(type) {
	case *DOid:
		__antithesis_instrumentation__.Notify(607176)
		v = t.DInt
	case *DInt:
		__antithesis_instrumentation__.Notify(607177)

		v = DInt(uint32(*t))
	default:
		__antithesis_instrumentation__.Notify(607178)
		return 0, makeUnsupportedComparisonMessage(d, other)
	}
	__antithesis_instrumentation__.Notify(607171)

	if d.DInt < v {
		__antithesis_instrumentation__.Notify(607179)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(607180)
	}
	__antithesis_instrumentation__.Notify(607172)
	if d.DInt > v {
		__antithesis_instrumentation__.Notify(607181)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(607182)
	}
	__antithesis_instrumentation__.Notify(607173)
	return 0, nil
}

func (d *DOid) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607183)
	if d.semanticType.Oid() == oid.T_oid || func() bool {
		__antithesis_instrumentation__.Notify(607184)
		return d.name == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(607185)

		d.DInt.Format(ctx)
	} else {
		__antithesis_instrumentation__.Notify(607186)
		if ctx.HasFlags(fmtDisambiguateDatumTypes) {
			__antithesis_instrumentation__.Notify(607187)
			ctx.WriteString("crdb_internal.create_")
			ctx.WriteString(d.semanticType.SQLStandardName())
			ctx.WriteByte('(')
			d.DInt.Format(ctx)
			ctx.WriteByte(',')
			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, d.name, lexbase.EncNoFlags)
			ctx.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(607188)

			lexbase.EncodeSQLStringWithFlags(&ctx.Buffer, d.name, lexbase.EncBareStrings)
		}
	}
}

func (d *DOid) IsMax(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607189)
	return d.DInt.IsMax(ctx)
}

func (d *DOid) IsMin(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607190)
	return d.DInt.IsMin(ctx)
}

func (d *DOid) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607191)
	next, ok := d.DInt.Next(ctx)
	return &DOid{*next.(*DInt), d.semanticType, ""}, ok
}

func (d *DOid) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607192)
	prev, ok := d.DInt.Prev(ctx)
	return &DOid{*prev.(*DInt), d.semanticType, ""}, ok
}

func (d *DOid) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(607193)
	return d.semanticType
}

func (d *DOid) Size() uintptr {
	__antithesis_instrumentation__.Notify(607194)
	return unsafe.Sizeof(*d)
}

func (d *DOid) Max(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607195)
	max, ok := d.DInt.Max(ctx)
	return &DOid{*max.(*DInt), d.semanticType, ""}, ok
}

func (d *DOid) Min(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607196)
	min, ok := d.DInt.Min(ctx)
	return &DOid{*min.(*DInt), d.semanticType, ""}, ok
}

type DOidWrapper struct {
	Wrapped Datum
	Oid     oid.Oid
}

const ZeroOidValue = "-"

func wrapWithOid(d Datum, oid oid.Oid) Datum {
	__antithesis_instrumentation__.Notify(607197)
	switch v := d.(type) {
	case nil:
		__antithesis_instrumentation__.Notify(607199)
		return nil
	case *DInt:
		__antithesis_instrumentation__.Notify(607200)
	case *DString:
		__antithesis_instrumentation__.Notify(607201)
	case *DArray:
		__antithesis_instrumentation__.Notify(607202)
	case dNull, *DOidWrapper:
		__antithesis_instrumentation__.Notify(607203)
		panic(errors.AssertionFailedf("cannot wrap %T with an Oid", v))
	default:
		__antithesis_instrumentation__.Notify(607204)

		panic(errors.AssertionFailedf("unsupported Datum type passed to wrapWithOid: %T", d))
	}
	__antithesis_instrumentation__.Notify(607198)
	return &DOidWrapper{
		Wrapped: d,
		Oid:     oid,
	}
}

func wrapAsZeroOid(t *types.T) Datum {
	__antithesis_instrumentation__.Notify(607205)
	tmpOid := NewDOid(0)
	tmpOid.semanticType = t
	if t.Oid() != oid.T_oid {
		__antithesis_instrumentation__.Notify(607207)
		tmpOid.name = ZeroOidValue
	} else {
		__antithesis_instrumentation__.Notify(607208)
	}
	__antithesis_instrumentation__.Notify(607206)
	return tmpOid
}

func UnwrapDatum(evalCtx *EvalContext, d Datum) Datum {
	__antithesis_instrumentation__.Notify(607209)
	if w, ok := d.(*DOidWrapper); ok {
		__antithesis_instrumentation__.Notify(607212)
		return w.Wrapped
	} else {
		__antithesis_instrumentation__.Notify(607213)
	}
	__antithesis_instrumentation__.Notify(607210)
	if p, ok := d.(*Placeholder); ok && func() bool {
		__antithesis_instrumentation__.Notify(607214)
		return evalCtx != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(607215)
		return evalCtx.HasPlaceholders() == true
	}() == true {
		__antithesis_instrumentation__.Notify(607216)
		ret, err := p.Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(607218)

			return d
		} else {
			__antithesis_instrumentation__.Notify(607219)
		}
		__antithesis_instrumentation__.Notify(607217)
		return ret
	} else {
		__antithesis_instrumentation__.Notify(607220)
	}
	__antithesis_instrumentation__.Notify(607211)
	return d
}

func (d *DOidWrapper) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(607221)
	return types.OidToType[d.Oid]
}

func (d *DOidWrapper) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(607222)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(607224)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607225)
	}
	__antithesis_instrumentation__.Notify(607223)
	return res
}

func (d *DOidWrapper) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(607226)
	if other == DNull {
		__antithesis_instrumentation__.Notify(607229)

		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(607230)
	}
	__antithesis_instrumentation__.Notify(607227)
	if v, ok := other.(*DOidWrapper); ok {
		__antithesis_instrumentation__.Notify(607231)
		return d.Wrapped.CompareError(ctx, v.Wrapped)
	} else {
		__antithesis_instrumentation__.Notify(607232)
	}
	__antithesis_instrumentation__.Notify(607228)
	return d.Wrapped.CompareError(ctx, other)
}

func (d *DOidWrapper) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607233)
	prev, ok := d.Wrapped.Prev(ctx)
	return wrapWithOid(prev, d.Oid), ok
}

func (d *DOidWrapper) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607234)
	next, ok := d.Wrapped.Next(ctx)
	return wrapWithOid(next, d.Oid), ok
}

func (d *DOidWrapper) IsMax(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607235)
	return d.Wrapped.IsMax(ctx)
}

func (d *DOidWrapper) IsMin(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607236)
	return d.Wrapped.IsMin(ctx)
}

func (d *DOidWrapper) Max(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607237)
	max, ok := d.Wrapped.Max(ctx)
	return wrapWithOid(max, d.Oid), ok
}

func (d *DOidWrapper) Min(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607238)
	min, ok := d.Wrapped.Min(ctx)
	return wrapWithOid(min, d.Oid), ok
}

func (d *DOidWrapper) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(607239)
	return d.Wrapped.AmbiguousFormat()
}

func (d *DOidWrapper) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(607240)

	ctx.FormatNode(d.Wrapped)
}

func (d *DOidWrapper) Size() uintptr {
	__antithesis_instrumentation__.Notify(607241)
	return unsafe.Sizeof(*d) + d.Wrapped.Size()
}

func (d *Placeholder) AmbiguousFormat() bool {
	__antithesis_instrumentation__.Notify(607242)
	return true
}

func (d *Placeholder) mustGetValue(ctx *EvalContext) Datum {
	__antithesis_instrumentation__.Notify(607243)
	e, ok := ctx.Placeholders.Value(d.Idx)
	if !ok {
		__antithesis_instrumentation__.Notify(607246)
		panic(errors.AssertionFailedf("fail"))
	} else {
		__antithesis_instrumentation__.Notify(607247)
	}
	__antithesis_instrumentation__.Notify(607244)
	out, err := e.Eval(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(607248)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "fail"))
	} else {
		__antithesis_instrumentation__.Notify(607249)
	}
	__antithesis_instrumentation__.Notify(607245)
	return out
}

func (d *Placeholder) Compare(ctx *EvalContext, other Datum) int {
	__antithesis_instrumentation__.Notify(607250)
	res, err := d.CompareError(ctx, other)
	if err != nil {
		__antithesis_instrumentation__.Notify(607252)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(607253)
	}
	__antithesis_instrumentation__.Notify(607251)
	return res
}

func (d *Placeholder) CompareError(ctx *EvalContext, other Datum) (int, error) {
	__antithesis_instrumentation__.Notify(607254)
	return d.mustGetValue(ctx).CompareError(ctx, other)
}

func (d *Placeholder) Prev(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607255)
	return d.mustGetValue(ctx).Prev(ctx)
}

func (d *Placeholder) IsMin(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607256)
	return d.mustGetValue(ctx).IsMin(ctx)
}

func (d *Placeholder) Next(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607257)
	return d.mustGetValue(ctx).Next(ctx)
}

func (d *Placeholder) IsMax(ctx *EvalContext) bool {
	__antithesis_instrumentation__.Notify(607258)
	return d.mustGetValue(ctx).IsMax(ctx)
}

func (d *Placeholder) Max(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607259)
	return d.mustGetValue(ctx).Max(ctx)
}

func (d *Placeholder) Min(ctx *EvalContext) (Datum, bool) {
	__antithesis_instrumentation__.Notify(607260)
	return d.mustGetValue(ctx).Min(ctx)
}

func (d *Placeholder) Size() uintptr {
	__antithesis_instrumentation__.Notify(607261)
	panic(errors.AssertionFailedf("shouldn't get called"))
}

func NewDNameFromDString(d *DString) Datum {
	__antithesis_instrumentation__.Notify(607262)
	return wrapWithOid(d, oid.T_name)
}

func NewDName(d string) Datum {
	__antithesis_instrumentation__.Notify(607263)
	return NewDNameFromDString(NewDString(d))
}

func NewDIntVectorFromDArray(d *DArray) Datum {
	__antithesis_instrumentation__.Notify(607264)
	ret := new(DArray)
	*ret = *d
	ret.customOid = oid.T_int2vector
	return ret
}

func NewDOidVectorFromDArray(d *DArray) Datum {
	__antithesis_instrumentation__.Notify(607265)
	ret := new(DArray)
	*ret = *d
	ret.customOid = oid.T_oidvector
	return ret
}

func NewDefaultDatum(evalCtx *EvalContext, t *types.T) (d Datum, err error) {
	__antithesis_instrumentation__.Notify(607266)
	switch t.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(607267)
		return DBoolFalse, nil
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(607268)
		return DZero, nil
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(607269)
		return dZeroFloat, nil
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(607270)
		return dZeroDecimal, nil
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(607271)
		return dEpochDate, nil
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(607272)
		return dZeroTimestamp, nil
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(607273)
		return dZeroInterval, nil
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(607274)
		return dEmptyString, nil
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(607275)
		return dEmptyBytes, nil
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(607276)
		return dZeroTimestampTZ, nil
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(607277)
		return NewDCollatedString("", t.Locale(), &evalCtx.CollationEnv)
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(607278)
		return NewDOidWithName(DInt(t.Oid()), t, t.SQLStandardName()), nil
	case types.UnknownFamily:
		__antithesis_instrumentation__.Notify(607279)
		return DNull, nil
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(607280)
		return DMinUUID, nil
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(607281)
		return NewDArray(t.ArrayContents()), nil
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(607282)
		return DMinIPAddr, nil
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(607283)
		return dTimeMin, nil
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(607284)
		return dNullJSON, nil
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(607285)
		return dZeroTimeTZ, nil
	case types.GeometryFamily, types.GeographyFamily, types.Box2DFamily:
		__antithesis_instrumentation__.Notify(607286)

		return nil, pgerror.Newf(
			pgcode.FeatureNotSupported,
			"%s must be set or be NULL",
			t.Name(),
		)
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(607287)
		contents := t.TupleContents()
		datums := make([]Datum, len(contents))
		for i, subT := range contents {
			__antithesis_instrumentation__.Notify(607293)
			datums[i], err = NewDefaultDatum(evalCtx, subT)
			if err != nil {
				__antithesis_instrumentation__.Notify(607294)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(607295)
			}
		}
		__antithesis_instrumentation__.Notify(607288)
		return NewDTuple(t, datums...), nil
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(607289)
		return bitArrayZero, nil
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(607290)

		if len(t.TypeMeta.EnumData.PhysicalRepresentations) == 0 {
			__antithesis_instrumentation__.Notify(607296)
			return nil, pgerror.Newf(
				pgcode.NotNullViolation,
				"%s has no values which can be used to satisfy the NOT NULL "+
					"constraint while adding or dropping",
				t.Name(),
			)
		} else {
			__antithesis_instrumentation__.Notify(607297)
		}
		__antithesis_instrumentation__.Notify(607291)

		return MakeDEnumFromPhysicalRepresentation(t,
			t.TypeMeta.EnumData.PhysicalRepresentations[0])
	default:
		__antithesis_instrumentation__.Notify(607292)
		return nil, errors.AssertionFailedf("unhandled type %v", t.SQLString())
	}
}

func DatumTypeSize(t *types.T) (size uintptr, isVarlen bool) {
	__antithesis_instrumentation__.Notify(607298)

	switch t.Family() {
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(607301)
		if types.IsWildcardTupleType(t) {
			__antithesis_instrumentation__.Notify(607307)
			return uintptr(0), false
		} else {
			__antithesis_instrumentation__.Notify(607308)
		}
		__antithesis_instrumentation__.Notify(607302)
		sz := uintptr(0)
		variable := false
		for i := range t.TupleContents() {
			__antithesis_instrumentation__.Notify(607309)
			typsz, typvariable := DatumTypeSize(t.TupleContents()[i])
			sz += typsz
			variable = variable || func() bool {
				__antithesis_instrumentation__.Notify(607310)
				return typvariable == true
			}() == true
		}
		__antithesis_instrumentation__.Notify(607303)
		return sz, variable
	case types.IntFamily, types.FloatFamily:
		__antithesis_instrumentation__.Notify(607304)
		return uintptr(t.Width() / 8), false

	case types.StringFamily:
		__antithesis_instrumentation__.Notify(607305)

		if t.Oid() == oid.T_char {
			__antithesis_instrumentation__.Notify(607311)
			return 1, false
		} else {
			__antithesis_instrumentation__.Notify(607312)
		}
	default:
		__antithesis_instrumentation__.Notify(607306)
	}
	__antithesis_instrumentation__.Notify(607299)

	if bSzInfo, ok := baseDatumTypeSizes[t.Family()]; ok {
		__antithesis_instrumentation__.Notify(607313)
		return bSzInfo.sz, bSzInfo.variable
	} else {
		__antithesis_instrumentation__.Notify(607314)
	}
	__antithesis_instrumentation__.Notify(607300)

	panic(errors.AssertionFailedf("unknown type: %T", t))
}

const (
	fixedSize    = false
	variableSize = true
)

var baseDatumTypeSizes = map[types.Family]struct {
	sz       uintptr
	variable bool
}{
	types.UnknownFamily:        {unsafe.Sizeof(dNull{}), fixedSize},
	types.BoolFamily:           {unsafe.Sizeof(DBool(false)), fixedSize},
	types.Box2DFamily:          {unsafe.Sizeof(DBox2D{CartesianBoundingBox: geo.CartesianBoundingBox{}}), fixedSize},
	types.BitFamily:            {unsafe.Sizeof(DBitArray{}), variableSize},
	types.IntFamily:            {unsafe.Sizeof(DInt(0)), fixedSize},
	types.FloatFamily:          {unsafe.Sizeof(DFloat(0.0)), fixedSize},
	types.DecimalFamily:        {unsafe.Sizeof(DDecimal{}), variableSize},
	types.StringFamily:         {unsafe.Sizeof(DString("")), variableSize},
	types.CollatedStringFamily: {unsafe.Sizeof(DCollatedString{"", "", nil}), variableSize},
	types.BytesFamily:          {unsafe.Sizeof(DBytes("")), variableSize},
	types.DateFamily:           {unsafe.Sizeof(DDate{}), fixedSize},
	types.GeographyFamily:      {unsafe.Sizeof(DGeography{}), variableSize},
	types.GeometryFamily:       {unsafe.Sizeof(DGeometry{}), variableSize},
	types.TimeFamily:           {unsafe.Sizeof(DTime(0)), fixedSize},
	types.TimeTZFamily:         {unsafe.Sizeof(DTimeTZ{}), fixedSize},
	types.TimestampFamily:      {unsafe.Sizeof(DTimestamp{}), fixedSize},
	types.TimestampTZFamily:    {unsafe.Sizeof(DTimestampTZ{}), fixedSize},
	types.IntervalFamily:       {unsafe.Sizeof(DInterval{}), fixedSize},
	types.JsonFamily:           {unsafe.Sizeof(DJSON{}), variableSize},
	types.UuidFamily:           {unsafe.Sizeof(DUuid{}), fixedSize},
	types.INetFamily:           {unsafe.Sizeof(DIPAddr{}), fixedSize},
	types.OidFamily:            {unsafe.Sizeof(DInt(0)), fixedSize},
	types.EnumFamily:           {unsafe.Sizeof(DEnum{}), variableSize},

	types.VoidFamily: {sz: unsafe.Sizeof(DVoid{}), variable: fixedSize},

	types.ArrayFamily: {unsafe.Sizeof(DString("")), variableSize},

	types.AnyFamily: {unsafe.Sizeof(DString("")), variableSize},
}

func MaxDistinctCount(evalCtx *EvalContext, first, last Datum) (_ int64, ok bool) {
	__antithesis_instrumentation__.Notify(607315)
	if !first.ResolvedType().Equivalent(last.ResolvedType()) {
		__antithesis_instrumentation__.Notify(607321)

		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(607322)
	}
	__antithesis_instrumentation__.Notify(607316)
	if first.Compare(evalCtx, last) == 0 {
		__antithesis_instrumentation__.Notify(607323)

		return 1, true
	} else {
		__antithesis_instrumentation__.Notify(607324)
	}
	__antithesis_instrumentation__.Notify(607317)

	var start, end int64

	switch t := first.(type) {
	case *DInt:
		__antithesis_instrumentation__.Notify(607325)
		otherDInt, otherOk := AsDInt(last)
		if otherOk {
			__antithesis_instrumentation__.Notify(607331)
			start = int64(*t)
			end = int64(otherDInt)
		} else {
			__antithesis_instrumentation__.Notify(607332)
		}

	case *DOid:
		__antithesis_instrumentation__.Notify(607326)
		otherDOid, otherOk := AsDOid(last)
		if otherOk {
			__antithesis_instrumentation__.Notify(607333)
			start = int64((*t).DInt)
			end = int64(otherDOid.DInt)
		} else {
			__antithesis_instrumentation__.Notify(607334)
		}

	case *DDate:
		__antithesis_instrumentation__.Notify(607327)
		otherDDate, otherOk := last.(*DDate)
		if otherOk {
			__antithesis_instrumentation__.Notify(607335)
			if !t.IsFinite() || func() bool {
				__antithesis_instrumentation__.Notify(607337)
				return !otherDDate.IsFinite() == true
			}() == true {
				__antithesis_instrumentation__.Notify(607338)

				return 0, false
			} else {
				__antithesis_instrumentation__.Notify(607339)
			}
			__antithesis_instrumentation__.Notify(607336)
			start = int64((*t).PGEpochDays())
			end = int64(otherDDate.PGEpochDays())
		} else {
			__antithesis_instrumentation__.Notify(607340)
		}

	case *DEnum:
		__antithesis_instrumentation__.Notify(607328)
		otherDEnum, otherOk := last.(*DEnum)
		if otherOk {
			__antithesis_instrumentation__.Notify(607341)
			startIdx, err := t.EnumTyp.EnumGetIdxOfPhysical(t.PhysicalRep)
			if err != nil {
				__antithesis_instrumentation__.Notify(607344)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(607345)
			}
			__antithesis_instrumentation__.Notify(607342)
			endIdx, err := t.EnumTyp.EnumGetIdxOfPhysical(otherDEnum.PhysicalRep)
			if err != nil {
				__antithesis_instrumentation__.Notify(607346)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(607347)
			}
			__antithesis_instrumentation__.Notify(607343)
			start, end = int64(startIdx), int64(endIdx)
		} else {
			__antithesis_instrumentation__.Notify(607348)
		}

	case *DBool:
		__antithesis_instrumentation__.Notify(607329)
		otherDBool, otherOk := last.(*DBool)
		if otherOk {
			__antithesis_instrumentation__.Notify(607349)
			if *t {
				__antithesis_instrumentation__.Notify(607351)
				start = 1
			} else {
				__antithesis_instrumentation__.Notify(607352)
			}
			__antithesis_instrumentation__.Notify(607350)
			if *otherDBool {
				__antithesis_instrumentation__.Notify(607353)
				end = 1
			} else {
				__antithesis_instrumentation__.Notify(607354)
			}
		} else {
			__antithesis_instrumentation__.Notify(607355)
		}

	default:
		__antithesis_instrumentation__.Notify(607330)

		return 0, false
	}
	__antithesis_instrumentation__.Notify(607318)

	if start > end {
		__antithesis_instrumentation__.Notify(607356)

		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(607357)
	}
	__antithesis_instrumentation__.Notify(607319)

	delta := end - start
	if delta < 0 {
		__antithesis_instrumentation__.Notify(607358)

		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(607359)
	}
	__antithesis_instrumentation__.Notify(607320)
	return delta + 1, true
}

func ParseDatumPath(evalCtx *EvalContext, str string, typs []types.Family) []Datum {
	__antithesis_instrumentation__.Notify(607360)
	var res []Datum
	for i, valStr := range ParsePath(str) {
		__antithesis_instrumentation__.Notify(607362)
		if i >= len(typs) {
			__antithesis_instrumentation__.Notify(607367)
			panic(errors.AssertionFailedf("invalid types"))
		} else {
			__antithesis_instrumentation__.Notify(607368)
		}
		__antithesis_instrumentation__.Notify(607363)

		if valStr == "NULL" {
			__antithesis_instrumentation__.Notify(607369)
			res = append(res, DNull)
			continue
		} else {
			__antithesis_instrumentation__.Notify(607370)
		}
		__antithesis_instrumentation__.Notify(607364)
		var val Datum
		var err error
		switch typs[i] {
		case types.BoolFamily:
			__antithesis_instrumentation__.Notify(607371)
			val, err = ParseDBool(valStr)
		case types.IntFamily:
			__antithesis_instrumentation__.Notify(607372)
			val, err = ParseDInt(valStr)
		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(607373)
			val, err = ParseDFloat(valStr)
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(607374)
			val, err = ParseDDecimal(valStr)
		case types.DateFamily:
			__antithesis_instrumentation__.Notify(607375)
			val, _, err = ParseDDate(evalCtx, valStr)
		case types.TimestampFamily:
			__antithesis_instrumentation__.Notify(607376)
			val, _, err = ParseDTimestamp(evalCtx, valStr, time.Microsecond)
		case types.TimestampTZFamily:
			__antithesis_instrumentation__.Notify(607377)
			val, _, err = ParseDTimestampTZ(evalCtx, valStr, time.Microsecond)
		case types.StringFamily:
			__antithesis_instrumentation__.Notify(607378)
			val = NewDString(valStr)
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(607379)
			val = NewDBytes(DBytes(valStr))
		case types.OidFamily:
			__antithesis_instrumentation__.Notify(607380)
			dInt, err := ParseDInt(valStr)
			if err == nil {
				__antithesis_instrumentation__.Notify(607386)
				val = NewDOid(*dInt)
			} else {
				__antithesis_instrumentation__.Notify(607387)
			}
		case types.UuidFamily:
			__antithesis_instrumentation__.Notify(607381)
			val, err = ParseDUuidFromString(valStr)
		case types.INetFamily:
			__antithesis_instrumentation__.Notify(607382)
			val, err = ParseDIPAddrFromINetString(valStr)
		case types.TimeFamily:
			__antithesis_instrumentation__.Notify(607383)
			val, _, err = ParseDTime(evalCtx, valStr, time.Microsecond)
		case types.TimeTZFamily:
			__antithesis_instrumentation__.Notify(607384)
			val, _, err = ParseDTimeTZ(evalCtx, valStr, time.Microsecond)
		default:
			__antithesis_instrumentation__.Notify(607385)
			panic(errors.AssertionFailedf("type %s not supported", typs[i].String()))
		}
		__antithesis_instrumentation__.Notify(607365)
		if err != nil {
			__antithesis_instrumentation__.Notify(607388)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(607389)
		}
		__antithesis_instrumentation__.Notify(607366)
		res = append(res, val)
	}
	__antithesis_instrumentation__.Notify(607361)
	return res
}

func ParsePath(str string) []string {
	__antithesis_instrumentation__.Notify(607390)
	if str == "" {
		__antithesis_instrumentation__.Notify(607393)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(607394)
	}
	__antithesis_instrumentation__.Notify(607391)
	if str[0] != '/' {
		__antithesis_instrumentation__.Notify(607395)
		panic(str)
	} else {
		__antithesis_instrumentation__.Notify(607396)
	}
	__antithesis_instrumentation__.Notify(607392)
	return strings.Split(str, "/")[1:]
}

func InferTypes(vals []string) []types.Family {
	__antithesis_instrumentation__.Notify(607397)

	typs := make([]types.Family, len(vals))
	for i := 0; i < len(vals); i++ {
		__antithesis_instrumentation__.Notify(607399)
		typ := types.IntFamily
		_, err := ParseDInt(vals[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(607401)
			typ = types.StringFamily
		} else {
			__antithesis_instrumentation__.Notify(607402)
		}
		__antithesis_instrumentation__.Notify(607400)
		typs[i] = typ
	}
	__antithesis_instrumentation__.Notify(607398)
	return typs
}
