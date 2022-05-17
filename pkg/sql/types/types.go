package types

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/oidext"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/proto"
	"github.com/lib/pq/oid"
)

type T struct {
	InternalType InternalType

	TypeMeta UserDefinedTypeMetadata
}

type UserDefinedTypeMetadata struct {
	Name *UserDefinedTypeName

	Version uint32

	EnumData *EnumMetadata
}

type EnumMetadata struct {
	PhysicalRepresentations [][]byte

	LogicalRepresentations []string

	IsMemberReadOnly []bool
}

func (e *EnumMetadata) debugString() string {
	__antithesis_instrumentation__.Notify(629572)
	return fmt.Sprintf(
		"PhysicalReps: %v; LogicalReps: %s",
		e.PhysicalRepresentations,
		e.LogicalRepresentations,
	)
}

type UserDefinedTypeName struct {
	Catalog        string
	ExplicitSchema bool
	Schema         string
	Name           string
}

func (u UserDefinedTypeName) Basename() string {
	__antithesis_instrumentation__.Notify(629573)
	return u.Name
}

func (u UserDefinedTypeName) FQName() string {
	__antithesis_instrumentation__.Notify(629574)
	var sb strings.Builder

	if u.ExplicitSchema {
		__antithesis_instrumentation__.Notify(629576)
		sb.WriteString(u.Schema)
		sb.WriteString(".")
	} else {
		__antithesis_instrumentation__.Notify(629577)
	}
	__antithesis_instrumentation__.Notify(629575)
	sb.WriteString(u.Name)
	return sb.String()
}

var (
	Unknown = &T{InternalType: InternalType{
		Family: UnknownFamily, Oid: oid.T_unknown, Locale: &emptyLocale}}

	Bool = &T{InternalType: InternalType{
		Family: BoolFamily, Oid: oid.T_bool, Locale: &emptyLocale}}

	VarBit = &T{InternalType: InternalType{
		Family: BitFamily, Oid: oid.T_varbit, Locale: &emptyLocale}}

	Int = &T{InternalType: InternalType{
		Family: IntFamily, Width: 64, Oid: oid.T_int8, Locale: &emptyLocale}}

	Int4 = &T{InternalType: InternalType{
		Family: IntFamily, Width: 32, Oid: oid.T_int4, Locale: &emptyLocale}}

	Int2 = &T{InternalType: InternalType{
		Family: IntFamily, Width: 16, Oid: oid.T_int2, Locale: &emptyLocale}}

	Float = &T{InternalType: InternalType{
		Family: FloatFamily, Width: 64, Oid: oid.T_float8, Locale: &emptyLocale}}

	Float4 = &T{InternalType: InternalType{
		Family: FloatFamily, Width: 32, Oid: oid.T_float4, Locale: &emptyLocale}}

	Decimal = &T{InternalType: InternalType{
		Family: DecimalFamily, Oid: oid.T_numeric, Locale: &emptyLocale}}

	String = &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_text, Locale: &emptyLocale}}

	VarChar = &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_varchar, Locale: &emptyLocale}}

	QChar = &T{InternalType: InternalType{
		Family: StringFamily, Width: 1, Oid: oid.T_char, Locale: &emptyLocale}}

	Name = &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_name, Locale: &emptyLocale}}

	Bytes = &T{InternalType: InternalType{
		Family: BytesFamily, Oid: oid.T_bytea, Locale: &emptyLocale}}

	Date = &T{InternalType: InternalType{
		Family: DateFamily, Oid: oid.T_date, Locale: &emptyLocale}}

	Time = &T{InternalType: InternalType{
		Family:             TimeFamily,
		Precision:          0,
		TimePrecisionIsSet: false,
		Oid:                oid.T_time,
		Locale:             &emptyLocale,
	}}

	TimeTZ = &T{InternalType: InternalType{
		Family:             TimeTZFamily,
		Precision:          0,
		TimePrecisionIsSet: false,
		Oid:                oid.T_timetz,
		Locale:             &emptyLocale,
	}}

	Timestamp = &T{InternalType: InternalType{
		Family:             TimestampFamily,
		Precision:          0,
		TimePrecisionIsSet: false,
		Oid:                oid.T_timestamp,
		Locale:             &emptyLocale,
	}}

	TimestampTZ = &T{InternalType: InternalType{
		Family:             TimestampTZFamily,
		Precision:          0,
		TimePrecisionIsSet: false,
		Oid:                oid.T_timestamptz,
		Locale:             &emptyLocale,
	}}

	Interval = &T{InternalType: InternalType{
		Family:                IntervalFamily,
		Precision:             0,
		TimePrecisionIsSet:    false,
		Oid:                   oid.T_interval,
		Locale:                &emptyLocale,
		IntervalDurationField: &IntervalDurationField{},
	}}

	Jsonb = &T{InternalType: InternalType{
		Family: JsonFamily, Oid: oid.T_jsonb, Locale: &emptyLocale}}

	Uuid = &T{InternalType: InternalType{
		Family: UuidFamily, Oid: oid.T_uuid, Locale: &emptyLocale}}

	INet = &T{InternalType: InternalType{
		Family: INetFamily, Oid: oid.T_inet, Locale: &emptyLocale}}

	Geometry = &T{
		InternalType: InternalType{
			Family:      GeometryFamily,
			Oid:         oidext.T_geometry,
			Locale:      &emptyLocale,
			GeoMetadata: &GeoMetadata{},
		},
	}

	Geography = &T{
		InternalType: InternalType{
			Family:      GeographyFamily,
			Oid:         oidext.T_geography,
			Locale:      &emptyLocale,
			GeoMetadata: &GeoMetadata{},
		},
	}

	Box2D = &T{
		InternalType: InternalType{
			Family: Box2DFamily,
			Oid:    oidext.T_box2d,
			Locale: &emptyLocale,
		},
	}

	Void = &T{
		InternalType: InternalType{
			Family: VoidFamily,
			Oid:    oid.T_void,
			Locale: &emptyLocale,
		},
	}

	Scalar = []*T{
		Bool,
		Box2D,
		Int,
		Float,
		Decimal,
		Date,
		Timestamp,
		Interval,
		Geography,
		Geometry,
		String,
		Bytes,
		TimestampTZ,
		Oid,
		Uuid,
		INet,
		Time,
		TimeTZ,
		Jsonb,
		VarBit,
	}

	Any = &T{InternalType: InternalType{
		Family: AnyFamily, Oid: oid.T_anyelement, Locale: &emptyLocale}}

	AnyArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Any, Oid: oid.T_anyarray, Locale: &emptyLocale}}

	AnyEnum = &T{InternalType: InternalType{
		Family: EnumFamily, Locale: &emptyLocale, Oid: oid.T_anyenum}}

	AnyTuple = &T{InternalType: InternalType{
		Family: TupleFamily, TupleContents: []*T{Any}, Oid: oid.T_record, Locale: &emptyLocale}}

	AnyTupleArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: AnyTuple, Oid: oid.T__record, Locale: &emptyLocale}}

	AnyCollatedString = &T{InternalType: InternalType{
		Family: CollatedStringFamily, Oid: oid.T_text, Locale: &emptyLocale}}

	EmptyTuple = &T{InternalType: InternalType{
		Family: TupleFamily, Oid: oid.T_record, Locale: &emptyLocale}}

	StringArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: String, Oid: oid.T__text, Locale: &emptyLocale}}

	BytesArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Bytes, Oid: oid.T__bytea, Locale: &emptyLocale}}

	IntArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Int, Oid: oid.T__int8, Locale: &emptyLocale}}

	FloatArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Float, Oid: oid.T__float8, Locale: &emptyLocale}}

	DecimalArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Decimal, Oid: oid.T__numeric, Locale: &emptyLocale}}

	BoolArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Bool, Oid: oid.T__bool, Locale: &emptyLocale}}

	UUIDArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Uuid, Oid: oid.T__uuid, Locale: &emptyLocale}}

	DateArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Date, Oid: oid.T__date, Locale: &emptyLocale}}

	TimeArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Time, Oid: oid.T__time, Locale: &emptyLocale}}

	TimeTZArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: TimeTZ, Oid: oid.T__timetz, Locale: &emptyLocale}}

	TimestampArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Timestamp, Oid: oid.T__timestamp, Locale: &emptyLocale}}

	TimestampTZArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: TimestampTZ, Oid: oid.T__timestamptz, Locale: &emptyLocale}}

	IntervalArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Interval, Oid: oid.T__interval, Locale: &emptyLocale}}

	INetArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: INet, Oid: oid.T__inet, Locale: &emptyLocale}}

	VarBitArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: VarBit, Oid: oid.T__varbit, Locale: &emptyLocale}}

	AnyEnumArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: AnyEnum, Oid: oid.T_anyarray, Locale: &emptyLocale}}

	JSONArray = &T{InternalType: InternalType{
		Family: ArrayFamily, ArrayContents: Jsonb, Oid: oid.T__jsonb, Locale: &emptyLocale}}

	Int2Vector = &T{InternalType: InternalType{
		Family: ArrayFamily, Oid: oid.T_int2vector, ArrayContents: Int2, Locale: &emptyLocale}}
)

var (
	typeBit = &T{InternalType: InternalType{
		Family: BitFamily, Oid: oid.T_bit, Locale: &emptyLocale}}

	typeBpChar = &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_bpchar, Locale: &emptyLocale}}

	typeQChar = &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_char, Locale: &emptyLocale}}
)

const (
	name Family = 11

	int2vector Family = 200

	oidvector Family = 201

	visibleNONE = 0

	visibleINTEGER = 1

	visibleSMALLINT = 2

	visibleBIGINT = 3

	visibleBIT = 4

	visibleREAL = 5

	visibleDOUBLE = 6

	visibleVARCHAR = 7

	visibleCHAR = 8

	visibleQCHAR = 9

	visibleVARBIT = 10

	unknownArrayOid = 0
)

const (
	defaultTimePrecision = 6
)

var (
	emptyLocale = ""
)

func MakeScalar(family Family, o oid.Oid, precision, width int32, locale string) *T {
	__antithesis_instrumentation__.Notify(629578)
	t := OidToType[o]
	if family != t.Family() {
		__antithesis_instrumentation__.Notify(629585)
		if family != CollatedStringFamily || func() bool {
			__antithesis_instrumentation__.Notify(629586)
			return StringFamily != t.Family() == true
		}() == true {
			__antithesis_instrumentation__.Notify(629587)
			panic(errors.AssertionFailedf("oid %d does not match %s", o, family))
		} else {
			__antithesis_instrumentation__.Notify(629588)
		}
	} else {
		__antithesis_instrumentation__.Notify(629589)
	}
	__antithesis_instrumentation__.Notify(629579)
	if family == ArrayFamily || func() bool {
		__antithesis_instrumentation__.Notify(629590)
		return family == TupleFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(629591)
		panic(errors.AssertionFailedf("cannot make non-scalar type %s", family))
	} else {
		__antithesis_instrumentation__.Notify(629592)
	}
	__antithesis_instrumentation__.Notify(629580)
	if family != CollatedStringFamily && func() bool {
		__antithesis_instrumentation__.Notify(629593)
		return locale != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(629594)
		panic(errors.AssertionFailedf("non-collation type cannot have locale %s", locale))
	} else {
		__antithesis_instrumentation__.Notify(629595)
	}
	__antithesis_instrumentation__.Notify(629581)

	timePrecisionIsSet := false
	var intervalDurationField *IntervalDurationField
	var geoMetadata *GeoMetadata
	switch family {
	case IntervalFamily:
		__antithesis_instrumentation__.Notify(629596)
		intervalDurationField = &IntervalDurationField{}
		if precision < 0 || func() bool {
			__antithesis_instrumentation__.Notify(629602)
			return precision > 6 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629603)
			panic(errors.AssertionFailedf("precision must be between 0 and 6 inclusive"))
		} else {
			__antithesis_instrumentation__.Notify(629604)
		}
		__antithesis_instrumentation__.Notify(629597)
		timePrecisionIsSet = true
	case TimestampFamily, TimestampTZFamily, TimeFamily, TimeTZFamily:
		__antithesis_instrumentation__.Notify(629598)
		if precision < 0 || func() bool {
			__antithesis_instrumentation__.Notify(629605)
			return precision > 6 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629606)
			panic(errors.AssertionFailedf("precision must be between 0 and 6 inclusive"))
		} else {
			__antithesis_instrumentation__.Notify(629607)
		}
		__antithesis_instrumentation__.Notify(629599)
		timePrecisionIsSet = true
	case DecimalFamily:
		__antithesis_instrumentation__.Notify(629600)
		if precision < 0 {
			__antithesis_instrumentation__.Notify(629608)
			panic(errors.AssertionFailedf("negative precision is not allowed"))
		} else {
			__antithesis_instrumentation__.Notify(629609)
		}
	default:
		__antithesis_instrumentation__.Notify(629601)
		if precision != 0 {
			__antithesis_instrumentation__.Notify(629610)
			panic(errors.AssertionFailedf("type %s cannot have precision", family))
		} else {
			__antithesis_instrumentation__.Notify(629611)
		}
	}
	__antithesis_instrumentation__.Notify(629582)

	if width < 0 {
		__antithesis_instrumentation__.Notify(629612)
		panic(errors.AssertionFailedf("negative width is not allowed"))
	} else {
		__antithesis_instrumentation__.Notify(629613)
	}
	__antithesis_instrumentation__.Notify(629583)
	switch family {
	case IntFamily:
		__antithesis_instrumentation__.Notify(629614)
		switch width {
		case 16, 32, 64:
			__antithesis_instrumentation__.Notify(629621)
		default:
			__antithesis_instrumentation__.Notify(629622)
			panic(errors.AssertionFailedf("invalid width %d for IntFamily type", width))
		}
	case FloatFamily:
		__antithesis_instrumentation__.Notify(629615)
		switch width {
		case 32, 64:
			__antithesis_instrumentation__.Notify(629623)
		default:
			__antithesis_instrumentation__.Notify(629624)
			panic(errors.AssertionFailedf("invalid width %d for FloatFamily type", width))
		}
	case DecimalFamily:
		__antithesis_instrumentation__.Notify(629616)
		if width > precision {
			__antithesis_instrumentation__.Notify(629625)
			panic(errors.AssertionFailedf(
				"decimal scale %d cannot be larger than precision %d", width, precision))
		} else {
			__antithesis_instrumentation__.Notify(629626)
		}
	case StringFamily, BytesFamily, CollatedStringFamily, BitFamily:
		__antithesis_instrumentation__.Notify(629617)

	case GeometryFamily:
		__antithesis_instrumentation__.Notify(629618)
		geoMetadata = &GeoMetadata{}
	case GeographyFamily:
		__antithesis_instrumentation__.Notify(629619)
		geoMetadata = &GeoMetadata{}
	default:
		__antithesis_instrumentation__.Notify(629620)
		if width != 0 {
			__antithesis_instrumentation__.Notify(629627)
			panic(errors.AssertionFailedf("type %s cannot have width", family))
		} else {
			__antithesis_instrumentation__.Notify(629628)
		}
	}
	__antithesis_instrumentation__.Notify(629584)

	return &T{InternalType: InternalType{
		Family:                family,
		Oid:                   o,
		Precision:             precision,
		TimePrecisionIsSet:    timePrecisionIsSet,
		Width:                 width,
		Locale:                &locale,
		IntervalDurationField: intervalDurationField,
		GeoMetadata:           geoMetadata,
	}}
}

func MakeBit(width int32) *T {
	__antithesis_instrumentation__.Notify(629629)
	if width == 0 {
		__antithesis_instrumentation__.Notify(629632)
		return typeBit
	} else {
		__antithesis_instrumentation__.Notify(629633)
	}
	__antithesis_instrumentation__.Notify(629630)
	if width < 0 {
		__antithesis_instrumentation__.Notify(629634)
		panic(errors.AssertionFailedf("width %d cannot be negative", width))
	} else {
		__antithesis_instrumentation__.Notify(629635)
	}
	__antithesis_instrumentation__.Notify(629631)
	return &T{InternalType: InternalType{
		Family: BitFamily, Oid: oid.T_bit, Width: width, Locale: &emptyLocale}}
}

func MakeVarBit(width int32) *T {
	__antithesis_instrumentation__.Notify(629636)
	if width == 0 {
		__antithesis_instrumentation__.Notify(629639)
		return VarBit
	} else {
		__antithesis_instrumentation__.Notify(629640)
	}
	__antithesis_instrumentation__.Notify(629637)
	if width < 0 {
		__antithesis_instrumentation__.Notify(629641)
		panic(errors.AssertionFailedf("width %d cannot be negative", width))
	} else {
		__antithesis_instrumentation__.Notify(629642)
	}
	__antithesis_instrumentation__.Notify(629638)
	return &T{InternalType: InternalType{
		Family: BitFamily, Width: width, Oid: oid.T_varbit, Locale: &emptyLocale}}
}

func MakeString(width int32) *T {
	__antithesis_instrumentation__.Notify(629643)
	if width == 0 {
		__antithesis_instrumentation__.Notify(629646)
		return String
	} else {
		__antithesis_instrumentation__.Notify(629647)
	}
	__antithesis_instrumentation__.Notify(629644)
	if width < 0 {
		__antithesis_instrumentation__.Notify(629648)
		panic(errors.AssertionFailedf("width %d cannot be negative", width))
	} else {
		__antithesis_instrumentation__.Notify(629649)
	}
	__antithesis_instrumentation__.Notify(629645)
	return &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_text, Width: width, Locale: &emptyLocale}}
}

func MakeVarChar(width int32) *T {
	__antithesis_instrumentation__.Notify(629650)
	if width == 0 {
		__antithesis_instrumentation__.Notify(629653)
		return VarChar
	} else {
		__antithesis_instrumentation__.Notify(629654)
	}
	__antithesis_instrumentation__.Notify(629651)
	if width < 0 {
		__antithesis_instrumentation__.Notify(629655)
		panic(errors.AssertionFailedf("width %d cannot be negative", width))
	} else {
		__antithesis_instrumentation__.Notify(629656)
	}
	__antithesis_instrumentation__.Notify(629652)
	return &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_varchar, Width: width, Locale: &emptyLocale}}
}

func MakeChar(width int32) *T {
	__antithesis_instrumentation__.Notify(629657)
	if width == 0 {
		__antithesis_instrumentation__.Notify(629660)
		return typeBpChar
	} else {
		__antithesis_instrumentation__.Notify(629661)
	}
	__antithesis_instrumentation__.Notify(629658)
	if width < 0 {
		__antithesis_instrumentation__.Notify(629662)
		panic(errors.AssertionFailedf("width %d cannot be negative", width))
	} else {
		__antithesis_instrumentation__.Notify(629663)
	}
	__antithesis_instrumentation__.Notify(629659)
	return &T{InternalType: InternalType{
		Family: StringFamily, Oid: oid.T_bpchar, Width: width, Locale: &emptyLocale}}
}

func oidCanBeCollatedString(o oid.Oid) bool {
	__antithesis_instrumentation__.Notify(629664)
	switch o {
	case oid.T_text, oid.T_varchar, oid.T_bpchar, oid.T_char, oid.T_name:
		__antithesis_instrumentation__.Notify(629666)
		return true
	default:
		__antithesis_instrumentation__.Notify(629667)
	}
	__antithesis_instrumentation__.Notify(629665)
	return false
}

func MakeCollatedString(strType *T, locale string) *T {
	__antithesis_instrumentation__.Notify(629668)
	if oidCanBeCollatedString(strType.Oid()) {
		__antithesis_instrumentation__.Notify(629670)
		return &T{InternalType: InternalType{
			Family: CollatedStringFamily, Oid: strType.Oid(), Width: strType.Width(), Locale: &locale}}
	} else {
		__antithesis_instrumentation__.Notify(629671)
	}
	__antithesis_instrumentation__.Notify(629669)
	panic(errors.AssertionFailedf("cannot apply collation to non-string type: %s", strType))
}

func MakeDecimal(precision, scale int32) *T {
	__antithesis_instrumentation__.Notify(629672)
	if precision == 0 && func() bool {
		__antithesis_instrumentation__.Notify(629677)
		return scale == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(629678)
		return Decimal
	} else {
		__antithesis_instrumentation__.Notify(629679)
	}
	__antithesis_instrumentation__.Notify(629673)
	if precision < 0 {
		__antithesis_instrumentation__.Notify(629680)
		panic(errors.AssertionFailedf("precision %d cannot be negative", precision))
	} else {
		__antithesis_instrumentation__.Notify(629681)
	}
	__antithesis_instrumentation__.Notify(629674)
	if scale < 0 {
		__antithesis_instrumentation__.Notify(629682)
		panic(errors.AssertionFailedf("scale %d cannot be negative", scale))
	} else {
		__antithesis_instrumentation__.Notify(629683)
	}
	__antithesis_instrumentation__.Notify(629675)
	if scale > precision {
		__antithesis_instrumentation__.Notify(629684)
		panic(errors.AssertionFailedf(
			"scale %d cannot be larger than precision %d", scale, precision))
	} else {
		__antithesis_instrumentation__.Notify(629685)
	}
	__antithesis_instrumentation__.Notify(629676)
	return &T{InternalType: InternalType{
		Family:    DecimalFamily,
		Oid:       oid.T_numeric,
		Precision: precision,
		Width:     scale,
		Locale:    &emptyLocale,
	}}
}

func MakeTime(precision int32) *T {
	__antithesis_instrumentation__.Notify(629686)
	return &T{InternalType: InternalType{
		Family:             TimeFamily,
		Oid:                oid.T_time,
		Precision:          precision,
		TimePrecisionIsSet: true,
		Locale:             &emptyLocale,
	}}
}

func MakeTimeTZ(precision int32) *T {
	__antithesis_instrumentation__.Notify(629687)
	return &T{InternalType: InternalType{
		Family:             TimeTZFamily,
		Oid:                oid.T_timetz,
		Precision:          precision,
		TimePrecisionIsSet: true,
		Locale:             &emptyLocale,
	}}
}

func MakeGeometry(shape geopb.ShapeType, srid geopb.SRID) *T {
	__antithesis_instrumentation__.Notify(629688)

	if srid < 0 {
		__antithesis_instrumentation__.Notify(629690)
		srid = 0
	} else {
		__antithesis_instrumentation__.Notify(629691)
	}
	__antithesis_instrumentation__.Notify(629689)
	return &T{InternalType: InternalType{
		Family: GeometryFamily,
		Oid:    oidext.T_geometry,
		Locale: &emptyLocale,
		GeoMetadata: &GeoMetadata{
			ShapeType: shape,
			SRID:      srid,
		},
	}}
}

func MakeGeography(shape geopb.ShapeType, srid geopb.SRID) *T {
	__antithesis_instrumentation__.Notify(629692)

	if srid < 0 {
		__antithesis_instrumentation__.Notify(629694)
		srid = 0
	} else {
		__antithesis_instrumentation__.Notify(629695)
	}
	__antithesis_instrumentation__.Notify(629693)
	return &T{InternalType: InternalType{
		Family: GeographyFamily,
		Oid:    oidext.T_geography,
		Locale: &emptyLocale,
		GeoMetadata: &GeoMetadata{
			ShapeType: shape,
			SRID:      srid,
		},
	}}
}

func (t *T) GeoMetadata() (*GeoMetadata, error) {
	__antithesis_instrumentation__.Notify(629696)
	if t.InternalType.GeoMetadata == nil {
		__antithesis_instrumentation__.Notify(629698)
		return nil, errors.Newf("GeoMetadata does not exist on type")
	} else {
		__antithesis_instrumentation__.Notify(629699)
	}
	__antithesis_instrumentation__.Notify(629697)
	return t.InternalType.GeoMetadata, nil
}

func (t *T) GeoSRIDOrZero() geopb.SRID {
	__antithesis_instrumentation__.Notify(629700)
	if t.InternalType.GeoMetadata != nil {
		__antithesis_instrumentation__.Notify(629702)
		return t.InternalType.GeoMetadata.SRID
	} else {
		__antithesis_instrumentation__.Notify(629703)
	}
	__antithesis_instrumentation__.Notify(629701)
	return 0
}

var (
	DefaultIntervalTypeMetadata = IntervalTypeMetadata{}
)

type IntervalTypeMetadata struct {
	DurationField IntervalDurationField

	Precision int32

	PrecisionIsSet bool
}

func (m *IntervalDurationField) IsMinuteToSecond() bool {
	__antithesis_instrumentation__.Notify(629704)
	return m.FromDurationType == IntervalDurationType_MINUTE && func() bool {
		__antithesis_instrumentation__.Notify(629705)
		return m.DurationType == IntervalDurationType_SECOND == true
	}() == true
}

func (m *IntervalDurationField) IsDayToHour() bool {
	__antithesis_instrumentation__.Notify(629706)
	return m.FromDurationType == IntervalDurationType_DAY && func() bool {
		__antithesis_instrumentation__.Notify(629707)
		return m.DurationType == IntervalDurationType_HOUR == true
	}() == true
}

func (t *T) IntervalTypeMetadata() (IntervalTypeMetadata, error) {
	__antithesis_instrumentation__.Notify(629708)
	if t.Family() != IntervalFamily {
		__antithesis_instrumentation__.Notify(629710)
		return IntervalTypeMetadata{}, errors.Newf("cannot call IntervalTypeMetadata on non-intervals")
	} else {
		__antithesis_instrumentation__.Notify(629711)
	}
	__antithesis_instrumentation__.Notify(629709)
	return IntervalTypeMetadata{
		DurationField:  *t.InternalType.IntervalDurationField,
		Precision:      t.InternalType.Precision,
		PrecisionIsSet: t.InternalType.TimePrecisionIsSet,
	}, nil
}

func MakeInterval(itm IntervalTypeMetadata) *T {
	__antithesis_instrumentation__.Notify(629712)
	switch itm.DurationField.DurationType {
	case IntervalDurationType_SECOND, IntervalDurationType_UNSET:
		__antithesis_instrumentation__.Notify(629715)
	default:
		__antithesis_instrumentation__.Notify(629716)
		if itm.PrecisionIsSet {
			__antithesis_instrumentation__.Notify(629717)
			panic(errors.Errorf("cannot set precision for duration type %s", itm.DurationField.DurationType))
		} else {
			__antithesis_instrumentation__.Notify(629718)
		}
	}
	__antithesis_instrumentation__.Notify(629713)
	if itm.Precision > 0 && func() bool {
		__antithesis_instrumentation__.Notify(629719)
		return !itm.PrecisionIsSet == true
	}() == true {
		__antithesis_instrumentation__.Notify(629720)
		panic(errors.Errorf("precision must be set if Precision > 0"))
	} else {
		__antithesis_instrumentation__.Notify(629721)
	}
	__antithesis_instrumentation__.Notify(629714)

	return &T{InternalType: InternalType{
		Family:                IntervalFamily,
		Oid:                   oid.T_interval,
		Locale:                &emptyLocale,
		Precision:             itm.Precision,
		TimePrecisionIsSet:    itm.PrecisionIsSet,
		IntervalDurationField: &itm.DurationField,
	}}
}

func MakeTimestamp(precision int32) *T {
	__antithesis_instrumentation__.Notify(629722)
	return &T{InternalType: InternalType{
		Family:             TimestampFamily,
		Oid:                oid.T_timestamp,
		Precision:          precision,
		TimePrecisionIsSet: true,
		Locale:             &emptyLocale,
	}}
}

func MakeTimestampTZ(precision int32) *T {
	__antithesis_instrumentation__.Notify(629723)
	return &T{InternalType: InternalType{
		Family:             TimestampTZFamily,
		Oid:                oid.T_timestamptz,
		Precision:          precision,
		TimePrecisionIsSet: true,
		Locale:             &emptyLocale,
	}}
}

func MakeEnum(typeOID, arrayTypeOID oid.Oid) *T {
	__antithesis_instrumentation__.Notify(629724)
	return &T{InternalType: InternalType{
		Family: EnumFamily,
		Oid:    typeOID,
		Locale: &emptyLocale,
		UDTMetadata: &PersistentUserDefinedTypeMetadata{
			ArrayTypeOID: arrayTypeOID,
		},
	}}
}

func MakeArray(typ *T) *T {
	__antithesis_instrumentation__.Notify(629725)

	if typ.Family() == UnknownFamily {
		__antithesis_instrumentation__.Notify(629727)
		typ = String
	} else {
		__antithesis_instrumentation__.Notify(629728)
	}
	__antithesis_instrumentation__.Notify(629726)
	arr := &T{InternalType: InternalType{
		Family:        ArrayFamily,
		Oid:           CalcArrayOid(typ),
		ArrayContents: typ,
		Locale:        &emptyLocale,
	}}
	return arr
}

func MakeTuple(contents []*T) *T {
	__antithesis_instrumentation__.Notify(629729)
	return &T{InternalType: InternalType{
		Family: TupleFamily, Oid: oid.T_record, TupleContents: contents, Locale: &emptyLocale,
	}}
}

func MakeLabeledTuple(contents []*T, labels []string) *T {
	__antithesis_instrumentation__.Notify(629730)
	if len(contents) != len(labels) && func() bool {
		__antithesis_instrumentation__.Notify(629732)
		return labels != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(629733)
		panic(errors.AssertionFailedf(
			"tuple contents and labels must be of same length: %v, %v", contents, labels))
	} else {
		__antithesis_instrumentation__.Notify(629734)
	}
	__antithesis_instrumentation__.Notify(629731)
	return &T{InternalType: InternalType{
		Family:        TupleFamily,
		Oid:           oid.T_record,
		TupleContents: contents,
		TupleLabels:   labels,
		Locale:        &emptyLocale,
	}}
}

func (t *T) Family() Family {
	__antithesis_instrumentation__.Notify(629735)
	return t.InternalType.Family
}

func (t *T) Oid() oid.Oid {
	__antithesis_instrumentation__.Notify(629736)
	return t.InternalType.Oid
}

func (t *T) Locale() string {
	__antithesis_instrumentation__.Notify(629737)
	return *t.InternalType.Locale
}

func (t *T) Width() int32 {
	__antithesis_instrumentation__.Notify(629738)
	return t.InternalType.Width
}

func (t *T) Precision() int32 {
	__antithesis_instrumentation__.Notify(629739)
	switch t.InternalType.Family {
	case IntervalFamily, TimestampFamily, TimestampTZFamily, TimeFamily, TimeTZFamily:
		__antithesis_instrumentation__.Notify(629741)
		if t.InternalType.Precision == 0 && func() bool {
			__antithesis_instrumentation__.Notify(629743)
			return !t.InternalType.TimePrecisionIsSet == true
		}() == true {
			__antithesis_instrumentation__.Notify(629744)
			return defaultTimePrecision
		} else {
			__antithesis_instrumentation__.Notify(629745)
		}
	default:
		__antithesis_instrumentation__.Notify(629742)
	}
	__antithesis_instrumentation__.Notify(629740)
	return t.InternalType.Precision
}

func (t *T) TypeModifier() int32 {
	__antithesis_instrumentation__.Notify(629746)
	if t.Family() == ArrayFamily {
		__antithesis_instrumentation__.Notify(629750)
		return t.ArrayContents().TypeModifier()
	} else {
		__antithesis_instrumentation__.Notify(629751)
	}
	__antithesis_instrumentation__.Notify(629747)

	if t.Oid() == oid.T_char {
		__antithesis_instrumentation__.Notify(629752)
		return int32(-1)
	} else {
		__antithesis_instrumentation__.Notify(629753)
	}
	__antithesis_instrumentation__.Notify(629748)

	switch t.Family() {
	case StringFamily, CollatedStringFamily:
		__antithesis_instrumentation__.Notify(629754)
		if width := t.Width(); width != 0 {
			__antithesis_instrumentation__.Notify(629758)

			return width + 4
		} else {
			__antithesis_instrumentation__.Notify(629759)
		}
	case BitFamily:
		__antithesis_instrumentation__.Notify(629755)
		if width := t.Width(); width != 0 {
			__antithesis_instrumentation__.Notify(629760)
			return width
		} else {
			__antithesis_instrumentation__.Notify(629761)
		}
	case DecimalFamily:
		__antithesis_instrumentation__.Notify(629756)

		if width, precision := t.Width(), t.Precision(); precision != 0 || func() bool {
			__antithesis_instrumentation__.Notify(629762)
			return width != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629763)
			return ((precision << 16) | width) + 4
		} else {
			__antithesis_instrumentation__.Notify(629764)
		}
	default:
		__antithesis_instrumentation__.Notify(629757)
	}
	__antithesis_instrumentation__.Notify(629749)
	return int32(-1)
}

func (t *T) WithoutTypeModifiers() *T {
	__antithesis_instrumentation__.Notify(629765)
	switch t.Family() {
	case ArrayFamily:
		__antithesis_instrumentation__.Notify(629769)

		newContents := t.ArrayContents().WithoutTypeModifiers()
		if newContents == t.ArrayContents() {
			__antithesis_instrumentation__.Notify(629776)
			return t
		} else {
			__antithesis_instrumentation__.Notify(629777)
		}
		__antithesis_instrumentation__.Notify(629770)
		return MakeArray(newContents)
	case TupleFamily:
		__antithesis_instrumentation__.Notify(629771)

		oldContents := t.TupleContents()
		newContents := make([]*T, len(oldContents))
		changed := false
		for i := range newContents {
			__antithesis_instrumentation__.Notify(629778)
			newContents[i] = oldContents[i].WithoutTypeModifiers()
			if newContents[i] != oldContents[i] {
				__antithesis_instrumentation__.Notify(629779)
				changed = true
			} else {
				__antithesis_instrumentation__.Notify(629780)
			}
		}
		__antithesis_instrumentation__.Notify(629772)
		if !changed {
			__antithesis_instrumentation__.Notify(629781)
			return t
		} else {
			__antithesis_instrumentation__.Notify(629782)
		}
		__antithesis_instrumentation__.Notify(629773)
		return MakeTuple(newContents)
	case EnumFamily:
		__antithesis_instrumentation__.Notify(629774)

		return t
	default:
		__antithesis_instrumentation__.Notify(629775)
	}
	__antithesis_instrumentation__.Notify(629766)

	if oidCanBeCollatedString(t.Oid()) {
		__antithesis_instrumentation__.Notify(629783)
		newT := *t
		newT.InternalType.Width = 0
		return &newT
	} else {
		__antithesis_instrumentation__.Notify(629784)
	}
	__antithesis_instrumentation__.Notify(629767)

	t, ok := OidToType[t.Oid()]
	if !ok {
		__antithesis_instrumentation__.Notify(629785)
		panic(errors.AssertionFailedf("unexpected OID: %d", t.Oid()))
	} else {
		__antithesis_instrumentation__.Notify(629786)
	}
	__antithesis_instrumentation__.Notify(629768)
	return t
}

func (t *T) Scale() int32 {
	__antithesis_instrumentation__.Notify(629787)
	return t.InternalType.Width
}

func (t *T) ArrayContents() *T {
	__antithesis_instrumentation__.Notify(629788)
	return t.InternalType.ArrayContents
}

func (t *T) TupleContents() []*T {
	__antithesis_instrumentation__.Notify(629789)
	return t.InternalType.TupleContents
}

func (t *T) TupleLabels() []string {
	__antithesis_instrumentation__.Notify(629790)
	return t.InternalType.TupleLabels
}

func (t *T) UserDefinedArrayOID() oid.Oid {
	__antithesis_instrumentation__.Notify(629791)
	if t.InternalType.UDTMetadata == nil {
		__antithesis_instrumentation__.Notify(629793)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(629794)
	}
	__antithesis_instrumentation__.Notify(629792)
	return t.InternalType.UDTMetadata.ArrayTypeOID
}

func RemapUserDefinedTypeOIDs(t *T, newOID, newArrayOID oid.Oid) {
	__antithesis_instrumentation__.Notify(629795)
	if newOID != 0 {
		__antithesis_instrumentation__.Notify(629797)
		t.InternalType.Oid = newOID
	} else {
		__antithesis_instrumentation__.Notify(629798)
	}
	__antithesis_instrumentation__.Notify(629796)
	if t.Family() != ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(629799)
		return newArrayOID != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(629800)
		t.InternalType.UDTMetadata.ArrayTypeOID = newArrayOID
	} else {
		__antithesis_instrumentation__.Notify(629801)
	}
}

func (t *T) UserDefined() bool {
	__antithesis_instrumentation__.Notify(629802)
	return IsOIDUserDefinedType(t.Oid())
}

func IsOIDUserDefinedType(o oid.Oid) bool {
	__antithesis_instrumentation__.Notify(629803)

	return o > oidext.CockroachPredefinedOIDMax
}

var familyNames = map[Family]string{
	AnyFamily:            "any",
	ArrayFamily:          "array",
	BitFamily:            "bit",
	BoolFamily:           "bool",
	Box2DFamily:          "box2d",
	BytesFamily:          "bytes",
	CollatedStringFamily: "collatedstring",
	DateFamily:           "date",
	DecimalFamily:        "decimal",
	EnumFamily:           "enum",
	FloatFamily:          "float",
	GeographyFamily:      "geography",
	GeometryFamily:       "geometry",
	INetFamily:           "inet",
	IntFamily:            "int",
	IntervalFamily:       "interval",
	JsonFamily:           "jsonb",
	OidFamily:            "oid",
	StringFamily:         "string",
	TimeFamily:           "time",
	TimestampFamily:      "timestamp",
	TimestampTZFamily:    "timestamptz",
	TimeTZFamily:         "timetz",
	TupleFamily:          "tuple",
	UnknownFamily:        "unknown",
	UuidFamily:           "uuid",
	VoidFamily:           "void",
}

func (f Family) Name() string {
	__antithesis_instrumentation__.Notify(629804)
	ret, ok := familyNames[f]
	if !ok {
		__antithesis_instrumentation__.Notify(629806)
		panic(errors.AssertionFailedf("unexpected Family: %d", f))
	} else {
		__antithesis_instrumentation__.Notify(629807)
	}
	__antithesis_instrumentation__.Notify(629805)
	return ret
}

func (t *T) Name() string {
	__antithesis_instrumentation__.Notify(629808)
	switch fam := t.Family(); fam {
	case AnyFamily:
		__antithesis_instrumentation__.Notify(629809)
		return "anyelement"

	case ArrayFamily:
		__antithesis_instrumentation__.Notify(629810)
		switch t.Oid() {
		case oid.T_oidvector:
			__antithesis_instrumentation__.Notify(629824)
			return "oidvector"
		case oid.T_int2vector:
			__antithesis_instrumentation__.Notify(629825)
			return "int2vector"
		default:
			__antithesis_instrumentation__.Notify(629826)
		}
		__antithesis_instrumentation__.Notify(629811)
		return t.ArrayContents().Name() + "[]"

	case BitFamily:
		__antithesis_instrumentation__.Notify(629812)
		if t.Oid() == oid.T_varbit {
			__antithesis_instrumentation__.Notify(629827)
			return "varbit"
		} else {
			__antithesis_instrumentation__.Notify(629828)
		}
		__antithesis_instrumentation__.Notify(629813)
		return "bit"

	case FloatFamily:
		__antithesis_instrumentation__.Notify(629814)
		switch t.Width() {
		case 64:
			__antithesis_instrumentation__.Notify(629829)
			return "float"
		case 32:
			__antithesis_instrumentation__.Notify(629830)
			return "float4"
		default:
			__antithesis_instrumentation__.Notify(629831)
			panic(errors.AssertionFailedf("programming error: unknown float width: %d", t.Width()))
		}

	case IntFamily:
		__antithesis_instrumentation__.Notify(629815)
		switch t.Width() {
		case 64:
			__antithesis_instrumentation__.Notify(629832)
			return "int"
		case 32:
			__antithesis_instrumentation__.Notify(629833)
			return "int4"
		case 16:
			__antithesis_instrumentation__.Notify(629834)
			return "int2"
		default:
			__antithesis_instrumentation__.Notify(629835)
			panic(errors.AssertionFailedf("programming error: unknown int width: %d", t.Width()))
		}

	case OidFamily:
		__antithesis_instrumentation__.Notify(629816)
		return t.SQLStandardName()

	case StringFamily, CollatedStringFamily:
		__antithesis_instrumentation__.Notify(629817)
		switch t.Oid() {
		case oid.T_text:
			__antithesis_instrumentation__.Notify(629836)
			return "string"
		case oid.T_bpchar:
			__antithesis_instrumentation__.Notify(629837)
			return "char"
		case oid.T_char:
			__antithesis_instrumentation__.Notify(629838)

			return `"char"`
		case oid.T_varchar:
			__antithesis_instrumentation__.Notify(629839)
			return "varchar"
		case oid.T_name:
			__antithesis_instrumentation__.Notify(629840)
			return "name"
		default:
			__antithesis_instrumentation__.Notify(629841)
		}
		__antithesis_instrumentation__.Notify(629818)
		panic(errors.AssertionFailedf("unexpected OID: %d", t.Oid()))

	case TupleFamily:
		__antithesis_instrumentation__.Notify(629819)
		return t.SQLStandardName()

	case EnumFamily:
		__antithesis_instrumentation__.Notify(629820)
		if t.Oid() == oid.T_anyenum {
			__antithesis_instrumentation__.Notify(629842)
			return "anyenum"
		} else {
			__antithesis_instrumentation__.Notify(629843)
		}
		__antithesis_instrumentation__.Notify(629821)

		if t.TypeMeta.Name == nil {
			__antithesis_instrumentation__.Notify(629844)
			return "unknown_enum"
		} else {
			__antithesis_instrumentation__.Notify(629845)
		}
		__antithesis_instrumentation__.Notify(629822)
		return t.TypeMeta.Name.Basename()

	default:
		__antithesis_instrumentation__.Notify(629823)
		return fam.Name()
	}
}

func (t *T) PGName() string {
	__antithesis_instrumentation__.Notify(629846)
	name, ok := oidext.TypeName(t.Oid())
	if ok {
		__antithesis_instrumentation__.Notify(629850)
		return strings.ToLower(name)
	} else {
		__antithesis_instrumentation__.Notify(629851)
	}
	__antithesis_instrumentation__.Notify(629847)

	if t.UserDefined() {
		__antithesis_instrumentation__.Notify(629852)
		return t.TypeMeta.Name.Basename()
	} else {
		__antithesis_instrumentation__.Notify(629853)
	}
	__antithesis_instrumentation__.Notify(629848)

	if t.Family() != ArrayFamily || func() bool {
		__antithesis_instrumentation__.Notify(629854)
		return t.ArrayContents().Family() != UnknownFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(629855)
		panic(errors.AssertionFailedf("unknown PG name for oid %d", t.Oid()))
	} else {
		__antithesis_instrumentation__.Notify(629856)
	}
	__antithesis_instrumentation__.Notify(629849)
	return "_unknown"
}

func (t *T) SQLStandardName() string {
	__antithesis_instrumentation__.Notify(629857)
	return t.SQLStandardNameWithTypmod(false, 0)
}

var telemetryNameReplaceRegex = regexp.MustCompile("[^a-zA-Z0-9]")

func (t *T) TelemetryName() string {
	__antithesis_instrumentation__.Notify(629858)
	return strings.ToLower(telemetryNameReplaceRegex.ReplaceAllString(t.SQLString(), "_"))
}

func (t *T) SQLStandardNameWithTypmod(haveTypmod bool, typmod int) string {
	__antithesis_instrumentation__.Notify(629859)
	var buf strings.Builder
	switch t.Family() {
	case AnyFamily:
		__antithesis_instrumentation__.Notify(629860)
		return "anyelement"
	case ArrayFamily:
		__antithesis_instrumentation__.Notify(629861)
		switch t.Oid() {
		case oid.T_oidvector:
			__antithesis_instrumentation__.Notify(629898)
			return "oidvector"
		case oid.T_int2vector:
			__antithesis_instrumentation__.Notify(629899)
			return "int2vector"
		default:
			__antithesis_instrumentation__.Notify(629900)
		}
		__antithesis_instrumentation__.Notify(629862)
		return t.ArrayContents().SQLStandardName() + "[]"
	case BitFamily:
		__antithesis_instrumentation__.Notify(629863)
		if t.Oid() == oid.T_varbit {
			__antithesis_instrumentation__.Notify(629901)
			buf.WriteString("bit varying")
		} else {
			__antithesis_instrumentation__.Notify(629902)
			buf.WriteString("bit")
		}
		__antithesis_instrumentation__.Notify(629864)
		if !haveTypmod || func() bool {
			__antithesis_instrumentation__.Notify(629903)
			return typmod <= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629904)
			return buf.String()
		} else {
			__antithesis_instrumentation__.Notify(629905)
		}
		__antithesis_instrumentation__.Notify(629865)
		buf.WriteString(fmt.Sprintf("(%d)", typmod))
		return buf.String()
	case BoolFamily:
		__antithesis_instrumentation__.Notify(629866)
		return "boolean"
	case Box2DFamily:
		__antithesis_instrumentation__.Notify(629867)
		return "box2d"
	case BytesFamily:
		__antithesis_instrumentation__.Notify(629868)
		return "bytea"
	case DateFamily:
		__antithesis_instrumentation__.Notify(629869)
		return "date"
	case DecimalFamily:
		__antithesis_instrumentation__.Notify(629870)
		if !haveTypmod || func() bool {
			__antithesis_instrumentation__.Notify(629906)
			return typmod <= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629907)
			return "numeric"
		} else {
			__antithesis_instrumentation__.Notify(629908)
		}
		__antithesis_instrumentation__.Notify(629871)

		typmod -= 4
		return fmt.Sprintf(
			"numeric(%d,%d)",
			(typmod>>16)&0xffff,
			typmod&0xffff,
		)

	case FloatFamily:
		__antithesis_instrumentation__.Notify(629872)
		switch t.Width() {
		case 32:
			__antithesis_instrumentation__.Notify(629909)
			return "real"
		case 64:
			__antithesis_instrumentation__.Notify(629910)
			return "double precision"
		default:
			__antithesis_instrumentation__.Notify(629911)
			panic(errors.AssertionFailedf("programming error: unknown float width: %d", t.Width()))
		}
	case GeometryFamily, GeographyFamily:
		__antithesis_instrumentation__.Notify(629873)
		return t.Name() + t.InternalType.GeoMetadata.SQLString()
	case INetFamily:
		__antithesis_instrumentation__.Notify(629874)
		return "inet"
	case IntFamily:
		__antithesis_instrumentation__.Notify(629875)
		switch t.Width() {
		case 16:
			__antithesis_instrumentation__.Notify(629912)
			return "smallint"
		case 32:
			__antithesis_instrumentation__.Notify(629913)

			return "integer"
		case 64:
			__antithesis_instrumentation__.Notify(629914)
			return "bigint"
		default:
			__antithesis_instrumentation__.Notify(629915)
			panic(errors.AssertionFailedf("programming error: unknown int width: %d", t.Width()))
		}
	case IntervalFamily:
		__antithesis_instrumentation__.Notify(629876)

		return "interval"
	case JsonFamily:
		__antithesis_instrumentation__.Notify(629877)

		return "jsonb"
	case OidFamily:
		__antithesis_instrumentation__.Notify(629878)
		switch t.Oid() {
		case oid.T_oid:
			__antithesis_instrumentation__.Notify(629916)
			return "oid"
		case oid.T_regclass:
			__antithesis_instrumentation__.Notify(629917)
			return "regclass"
		case oid.T_regnamespace:
			__antithesis_instrumentation__.Notify(629918)
			return "regnamespace"
		case oid.T_regproc:
			__antithesis_instrumentation__.Notify(629919)
			return "regproc"
		case oid.T_regprocedure:
			__antithesis_instrumentation__.Notify(629920)
			return "regprocedure"
		case oid.T_regrole:
			__antithesis_instrumentation__.Notify(629921)
			return "regrole"
		case oid.T_regtype:
			__antithesis_instrumentation__.Notify(629922)
			return "regtype"
		default:
			__antithesis_instrumentation__.Notify(629923)
			panic(errors.AssertionFailedf("unexpected Oid: %v", errors.Safe(t.Oid())))
		}
	case StringFamily, CollatedStringFamily:
		__antithesis_instrumentation__.Notify(629879)
		switch t.Oid() {
		case oid.T_text:
			__antithesis_instrumentation__.Notify(629924)
			buf.WriteString("text")
		case oid.T_varchar:
			__antithesis_instrumentation__.Notify(629925)
			buf.WriteString("character varying")
		case oid.T_bpchar:
			__antithesis_instrumentation__.Notify(629926)
			if haveTypmod && func() bool {
				__antithesis_instrumentation__.Notify(629931)
				return typmod < 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(629932)

				return "bpchar"
			} else {
				__antithesis_instrumentation__.Notify(629933)
			}
			__antithesis_instrumentation__.Notify(629927)
			buf.WriteString("character")
		case oid.T_char:
			__antithesis_instrumentation__.Notify(629928)

			return `"char"`
		case oid.T_name:
			__antithesis_instrumentation__.Notify(629929)

			return "name"
		default:
			__antithesis_instrumentation__.Notify(629930)
			panic(errors.AssertionFailedf("unexpected OID: %d", t.Oid()))
		}
		__antithesis_instrumentation__.Notify(629880)
		if !haveTypmod {
			__antithesis_instrumentation__.Notify(629934)
			return buf.String()
		} else {
			__antithesis_instrumentation__.Notify(629935)
		}
		__antithesis_instrumentation__.Notify(629881)

		if t.Oid() != oid.T_text {
			__antithesis_instrumentation__.Notify(629936)
			typmod -= 4
		} else {
			__antithesis_instrumentation__.Notify(629937)
		}
		__antithesis_instrumentation__.Notify(629882)
		if typmod <= 0 {
			__antithesis_instrumentation__.Notify(629938)

			return buf.String()
		} else {
			__antithesis_instrumentation__.Notify(629939)
		}
		__antithesis_instrumentation__.Notify(629883)
		buf.WriteString(fmt.Sprintf("(%d)", typmod))
		return buf.String()

	case TimeFamily:
		__antithesis_instrumentation__.Notify(629884)
		if !haveTypmod || func() bool {
			__antithesis_instrumentation__.Notify(629940)
			return typmod < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629941)
			return "time without time zone"
		} else {
			__antithesis_instrumentation__.Notify(629942)
		}
		__antithesis_instrumentation__.Notify(629885)
		return fmt.Sprintf("time(%d) without time zone", typmod)
	case TimeTZFamily:
		__antithesis_instrumentation__.Notify(629886)
		if !haveTypmod || func() bool {
			__antithesis_instrumentation__.Notify(629943)
			return typmod < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629944)
			return "time with time zone"
		} else {
			__antithesis_instrumentation__.Notify(629945)
		}
		__antithesis_instrumentation__.Notify(629887)
		return fmt.Sprintf("time(%d) with time zone", typmod)
	case TimestampFamily:
		__antithesis_instrumentation__.Notify(629888)
		if !haveTypmod || func() bool {
			__antithesis_instrumentation__.Notify(629946)
			return typmod < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629947)
			return "timestamp without time zone"
		} else {
			__antithesis_instrumentation__.Notify(629948)
		}
		__antithesis_instrumentation__.Notify(629889)
		return fmt.Sprintf("timestamp(%d) without time zone", typmod)
	case TimestampTZFamily:
		__antithesis_instrumentation__.Notify(629890)
		if !haveTypmod || func() bool {
			__antithesis_instrumentation__.Notify(629949)
			return typmod < 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(629950)
			return "timestamp with time zone"
		} else {
			__antithesis_instrumentation__.Notify(629951)
		}
		__antithesis_instrumentation__.Notify(629891)
		return fmt.Sprintf("timestamp(%d) with time zone", typmod)
	case TupleFamily:
		__antithesis_instrumentation__.Notify(629892)
		if t.UserDefined() {
			__antithesis_instrumentation__.Notify(629952)

			return t.TypeMeta.Name.Basename()
		} else {
			__antithesis_instrumentation__.Notify(629953)
		}
		__antithesis_instrumentation__.Notify(629893)
		return "record"
	case UnknownFamily:
		__antithesis_instrumentation__.Notify(629894)
		return "unknown"
	case UuidFamily:
		__antithesis_instrumentation__.Notify(629895)
		return "uuid"
	case EnumFamily:
		__antithesis_instrumentation__.Notify(629896)
		return t.TypeMeta.Name.Basename()
	default:
		__antithesis_instrumentation__.Notify(629897)
		panic(errors.AssertionFailedf("unexpected Family: %v", errors.Safe(t.Family())))
	}
}

func (t *T) InformationSchemaName() string {
	__antithesis_instrumentation__.Notify(629954)

	if t.Family() == ArrayFamily {
		__antithesis_instrumentation__.Notify(629957)
		return "ARRAY"
	} else {
		__antithesis_instrumentation__.Notify(629958)
	}
	__antithesis_instrumentation__.Notify(629955)

	if t.TypeMeta.Name != nil {
		__antithesis_instrumentation__.Notify(629959)
		return "USER-DEFINED"
	} else {
		__antithesis_instrumentation__.Notify(629960)
	}
	__antithesis_instrumentation__.Notify(629956)
	return t.SQLStandardName()
}

func (t *T) SQLString() string {
	__antithesis_instrumentation__.Notify(629961)
	switch t.Family() {
	case BitFamily:
		__antithesis_instrumentation__.Notify(629963)
		o := t.Oid()
		typName := "BIT"
		if o == oid.T_varbit {
			__antithesis_instrumentation__.Notify(629983)
			typName = "VARBIT"
		} else {
			__antithesis_instrumentation__.Notify(629984)
		}
		__antithesis_instrumentation__.Notify(629964)

		if (o != oid.T_varbit && func() bool {
			__antithesis_instrumentation__.Notify(629985)
			return t.Width() > 1 == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(629986)
			return (o == oid.T_varbit && func() bool {
				__antithesis_instrumentation__.Notify(629987)
				return t.Width() > 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(629988)
			typName = fmt.Sprintf("%s(%d)", typName, t.Width())
		} else {
			__antithesis_instrumentation__.Notify(629989)
		}
		__antithesis_instrumentation__.Notify(629965)
		return typName
	case IntFamily:
		__antithesis_instrumentation__.Notify(629966)
		switch t.Width() {
		case 16:
			__antithesis_instrumentation__.Notify(629990)
			return "INT2"
		case 32:
			__antithesis_instrumentation__.Notify(629991)
			return "INT4"
		case 64:
			__antithesis_instrumentation__.Notify(629992)
			return "INT8"
		default:
			__antithesis_instrumentation__.Notify(629993)
			panic(errors.AssertionFailedf("programming error: unknown int width: %d", t.Width()))
		}
	case StringFamily:
		__antithesis_instrumentation__.Notify(629967)
		return t.stringTypeSQL()
	case CollatedStringFamily:
		__antithesis_instrumentation__.Notify(629968)
		return t.collatedStringTypeSQL(false)
	case FloatFamily:
		__antithesis_instrumentation__.Notify(629969)
		const realName = "FLOAT4"
		const doubleName = "FLOAT8"
		if t.Width() == 32 {
			__antithesis_instrumentation__.Notify(629994)
			return realName
		} else {
			__antithesis_instrumentation__.Notify(629995)
		}
		__antithesis_instrumentation__.Notify(629970)
		return doubleName
	case DecimalFamily:
		__antithesis_instrumentation__.Notify(629971)
		if t.Precision() > 0 {
			__antithesis_instrumentation__.Notify(629996)
			if t.Width() > 0 {
				__antithesis_instrumentation__.Notify(629998)
				return fmt.Sprintf("DECIMAL(%d,%d)", t.Precision(), t.Scale())
			} else {
				__antithesis_instrumentation__.Notify(629999)
			}
			__antithesis_instrumentation__.Notify(629997)
			return fmt.Sprintf("DECIMAL(%d)", t.Precision())
		} else {
			__antithesis_instrumentation__.Notify(630000)
		}
	case JsonFamily:
		__antithesis_instrumentation__.Notify(629972)

		return "JSONB"
	case TimestampFamily, TimestampTZFamily, TimeFamily, TimeTZFamily:
		__antithesis_instrumentation__.Notify(629973)
		if t.InternalType.Precision > 0 || func() bool {
			__antithesis_instrumentation__.Notify(630001)
			return t.InternalType.TimePrecisionIsSet == true
		}() == true {
			__antithesis_instrumentation__.Notify(630002)
			return fmt.Sprintf("%s(%d)", strings.ToUpper(t.Name()), t.Precision())
		} else {
			__antithesis_instrumentation__.Notify(630003)
		}
	case GeometryFamily, GeographyFamily:
		__antithesis_instrumentation__.Notify(629974)
		return strings.ToUpper(t.Name() + t.InternalType.GeoMetadata.SQLString())
	case IntervalFamily:
		__antithesis_instrumentation__.Notify(629975)
		switch t.InternalType.IntervalDurationField.DurationType {
		case IntervalDurationType_UNSET:
			__antithesis_instrumentation__.Notify(630004)
			if t.InternalType.Precision > 0 || func() bool {
				__antithesis_instrumentation__.Notify(630008)
				return t.InternalType.TimePrecisionIsSet == true
			}() == true {
				__antithesis_instrumentation__.Notify(630009)
				return fmt.Sprintf("%s(%d)", strings.ToUpper(t.Name()), t.Precision())
			} else {
				__antithesis_instrumentation__.Notify(630010)
			}
		default:
			__antithesis_instrumentation__.Notify(630005)
			fromStr := ""
			if t.InternalType.IntervalDurationField.FromDurationType != IntervalDurationType_UNSET {
				__antithesis_instrumentation__.Notify(630011)
				fromStr = fmt.Sprintf("%s TO ", t.InternalType.IntervalDurationField.FromDurationType.String())
			} else {
				__antithesis_instrumentation__.Notify(630012)
			}
			__antithesis_instrumentation__.Notify(630006)
			precisionStr := ""
			if t.InternalType.Precision > 0 || func() bool {
				__antithesis_instrumentation__.Notify(630013)
				return t.InternalType.TimePrecisionIsSet == true
			}() == true {
				__antithesis_instrumentation__.Notify(630014)
				precisionStr = fmt.Sprintf("(%d)", t.Precision())
			} else {
				__antithesis_instrumentation__.Notify(630015)
			}
			__antithesis_instrumentation__.Notify(630007)
			return fmt.Sprintf(
				"%s %s%s%s",
				strings.ToUpper(t.Name()),
				fromStr,
				t.InternalType.IntervalDurationField.DurationType.String(),
				precisionStr,
			)
		}
	case OidFamily:
		__antithesis_instrumentation__.Notify(629976)
		if name, ok := oidext.TypeName(t.Oid()); ok {
			__antithesis_instrumentation__.Notify(630016)
			return name
		} else {
			__antithesis_instrumentation__.Notify(630017)
		}
	case ArrayFamily:
		__antithesis_instrumentation__.Notify(629977)
		switch t.Oid() {
		case oid.T_oidvector:
			__antithesis_instrumentation__.Notify(630018)
			return "OIDVECTOR"
		case oid.T_int2vector:
			__antithesis_instrumentation__.Notify(630019)
			return "INT2VECTOR"
		default:
			__antithesis_instrumentation__.Notify(630020)
		}
		__antithesis_instrumentation__.Notify(629978)
		if t.ArrayContents().Family() == CollatedStringFamily {
			__antithesis_instrumentation__.Notify(630021)
			return t.ArrayContents().collatedStringTypeSQL(true)
		} else {
			__antithesis_instrumentation__.Notify(630022)
		}
		__antithesis_instrumentation__.Notify(629979)
		return t.ArrayContents().SQLString() + "[]"
	case EnumFamily:
		__antithesis_instrumentation__.Notify(629980)
		if t.Oid() == oid.T_anyenum {
			__antithesis_instrumentation__.Notify(630023)
			return "anyenum"
		} else {
			__antithesis_instrumentation__.Notify(630024)
		}
		__antithesis_instrumentation__.Notify(629981)
		return t.TypeMeta.Name.FQName()
	default:
		__antithesis_instrumentation__.Notify(629982)
	}
	__antithesis_instrumentation__.Notify(629962)
	return strings.ToUpper(t.Name())
}

func (t *T) Equivalent(other *T) bool {
	__antithesis_instrumentation__.Notify(630025)
	if t.Family() == AnyFamily || func() bool {
		__antithesis_instrumentation__.Notify(630029)
		return other.Family() == AnyFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(630030)
		return true
	} else {
		__antithesis_instrumentation__.Notify(630031)
	}
	__antithesis_instrumentation__.Notify(630026)
	if t.Family() != other.Family() {
		__antithesis_instrumentation__.Notify(630032)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630033)
	}
	__antithesis_instrumentation__.Notify(630027)

	switch t.Family() {
	case CollatedStringFamily:
		__antithesis_instrumentation__.Notify(630034)

		if t.Locale() != "" && func() bool {
			__antithesis_instrumentation__.Notify(630042)
			return other.Locale() != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(630043)
			if !lex.LocaleNamesAreEqual(t.Locale(), other.Locale()) {
				__antithesis_instrumentation__.Notify(630044)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630045)
			}
		} else {
			__antithesis_instrumentation__.Notify(630046)
		}

	case TupleFamily:
		__antithesis_instrumentation__.Notify(630035)

		if IsWildcardTupleType(t) || func() bool {
			__antithesis_instrumentation__.Notify(630047)
			return IsWildcardTupleType(other) == true
		}() == true {
			__antithesis_instrumentation__.Notify(630048)
			return true
		} else {
			__antithesis_instrumentation__.Notify(630049)
		}
		__antithesis_instrumentation__.Notify(630036)
		if len(t.TupleContents()) != len(other.TupleContents()) {
			__antithesis_instrumentation__.Notify(630050)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630051)
		}
		__antithesis_instrumentation__.Notify(630037)
		for i := range t.TupleContents() {
			__antithesis_instrumentation__.Notify(630052)
			if !t.TupleContents()[i].Equivalent(other.TupleContents()[i]) {
				__antithesis_instrumentation__.Notify(630053)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630054)
			}
		}

	case ArrayFamily:
		__antithesis_instrumentation__.Notify(630038)
		if !t.ArrayContents().Equivalent(other.ArrayContents()) {
			__antithesis_instrumentation__.Notify(630055)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630056)
		}

	case EnumFamily:
		__antithesis_instrumentation__.Notify(630039)

		if t.Oid() == oid.T_anyenum || func() bool {
			__antithesis_instrumentation__.Notify(630057)
			return other.Oid() == oid.T_anyenum == true
		}() == true {
			__antithesis_instrumentation__.Notify(630058)
			return true
		} else {
			__antithesis_instrumentation__.Notify(630059)
		}
		__antithesis_instrumentation__.Notify(630040)
		if t.Oid() != other.Oid() {
			__antithesis_instrumentation__.Notify(630060)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630061)
		}
	default:
		__antithesis_instrumentation__.Notify(630041)
	}
	__antithesis_instrumentation__.Notify(630028)

	return true
}

func (t *T) EquivalentOrNull(other *T, allowNullTupleEquivalence bool) bool {
	__antithesis_instrumentation__.Notify(630062)

	normalEquivalency := t.Equivalent(other)
	if normalEquivalency {
		__antithesis_instrumentation__.Notify(630064)
		return true
	} else {
		__antithesis_instrumentation__.Notify(630065)
	}
	__antithesis_instrumentation__.Notify(630063)

	switch t.Family() {
	case UnknownFamily:
		__antithesis_instrumentation__.Notify(630066)
		return allowNullTupleEquivalence || func() bool {
			__antithesis_instrumentation__.Notify(630073)
			return other.Family() != TupleFamily == true
		}() == true

	case TupleFamily:
		__antithesis_instrumentation__.Notify(630067)
		if other.Family() != TupleFamily {
			__antithesis_instrumentation__.Notify(630074)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630075)
		}
		__antithesis_instrumentation__.Notify(630068)

		if IsWildcardTupleType(t) || func() bool {
			__antithesis_instrumentation__.Notify(630076)
			return IsWildcardTupleType(other) == true
		}() == true {
			__antithesis_instrumentation__.Notify(630077)
			return true
		} else {
			__antithesis_instrumentation__.Notify(630078)
		}
		__antithesis_instrumentation__.Notify(630069)
		if len(t.TupleContents()) != len(other.TupleContents()) {
			__antithesis_instrumentation__.Notify(630079)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630080)
		}
		__antithesis_instrumentation__.Notify(630070)
		for i := range t.TupleContents() {
			__antithesis_instrumentation__.Notify(630081)
			if !t.TupleContents()[i].EquivalentOrNull(other.TupleContents()[i], allowNullTupleEquivalence) {
				__antithesis_instrumentation__.Notify(630082)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630083)
			}
		}
		__antithesis_instrumentation__.Notify(630071)
		return true

	default:
		__antithesis_instrumentation__.Notify(630072)
		return normalEquivalency
	}
}

func (t *T) Identical(other *T) bool {
	__antithesis_instrumentation__.Notify(630084)
	return t.InternalType.Identical(&other.InternalType)
}

func (t *T) Equal(other *T) bool {
	__antithesis_instrumentation__.Notify(630085)
	return t.Identical(other)
}

func (t *T) Size() (n int) {
	__antithesis_instrumentation__.Notify(630086)

	temp := *t
	err := temp.downgradeType()
	if err != nil {
		__antithesis_instrumentation__.Notify(630088)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "error during Size call"))
	} else {
		__antithesis_instrumentation__.Notify(630089)
	}
	__antithesis_instrumentation__.Notify(630087)
	return temp.InternalType.Size()
}

func (t *InternalType) Identical(other *InternalType) bool {
	__antithesis_instrumentation__.Notify(630090)
	if t.Family != other.Family {
		__antithesis_instrumentation__.Notify(630104)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630105)
	}
	__antithesis_instrumentation__.Notify(630091)
	if t.Width != other.Width {
		__antithesis_instrumentation__.Notify(630106)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630107)
	}
	__antithesis_instrumentation__.Notify(630092)
	if t.Precision != other.Precision {
		__antithesis_instrumentation__.Notify(630108)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630109)
	}
	__antithesis_instrumentation__.Notify(630093)
	if t.TimePrecisionIsSet != other.TimePrecisionIsSet {
		__antithesis_instrumentation__.Notify(630110)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630111)
	}
	__antithesis_instrumentation__.Notify(630094)
	if t.IntervalDurationField != nil && func() bool {
		__antithesis_instrumentation__.Notify(630112)
		return other.IntervalDurationField != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(630113)
		if *t.IntervalDurationField != *other.IntervalDurationField {
			__antithesis_instrumentation__.Notify(630114)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630115)
		}
	} else {
		__antithesis_instrumentation__.Notify(630116)
		if t.IntervalDurationField != nil {
			__antithesis_instrumentation__.Notify(630117)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630118)
			if other.IntervalDurationField != nil {
				__antithesis_instrumentation__.Notify(630119)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630120)
			}
		}
	}
	__antithesis_instrumentation__.Notify(630095)
	if t.GeoMetadata != nil && func() bool {
		__antithesis_instrumentation__.Notify(630121)
		return other.GeoMetadata != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(630122)
		if t.GeoMetadata.ShapeType != other.GeoMetadata.ShapeType {
			__antithesis_instrumentation__.Notify(630124)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630125)
		}
		__antithesis_instrumentation__.Notify(630123)
		if t.GeoMetadata.SRID != other.GeoMetadata.SRID {
			__antithesis_instrumentation__.Notify(630126)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630127)
		}
	} else {
		__antithesis_instrumentation__.Notify(630128)
		if t.GeoMetadata != nil {
			__antithesis_instrumentation__.Notify(630129)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630130)
			if other.GeoMetadata != nil {
				__antithesis_instrumentation__.Notify(630131)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630132)
			}
		}
	}
	__antithesis_instrumentation__.Notify(630096)
	if t.Locale != nil && func() bool {
		__antithesis_instrumentation__.Notify(630133)
		return other.Locale != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(630134)
		if *t.Locale != *other.Locale {
			__antithesis_instrumentation__.Notify(630135)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630136)
		}
	} else {
		__antithesis_instrumentation__.Notify(630137)
		if t.Locale != nil {
			__antithesis_instrumentation__.Notify(630138)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630139)
			if other.Locale != nil {
				__antithesis_instrumentation__.Notify(630140)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630141)
			}
		}
	}
	__antithesis_instrumentation__.Notify(630097)
	if t.ArrayContents != nil && func() bool {
		__antithesis_instrumentation__.Notify(630142)
		return other.ArrayContents != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(630143)
		if !t.ArrayContents.Identical(other.ArrayContents) {
			__antithesis_instrumentation__.Notify(630144)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630145)
		}
	} else {
		__antithesis_instrumentation__.Notify(630146)
		if t.ArrayContents != nil {
			__antithesis_instrumentation__.Notify(630147)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630148)
			if other.ArrayContents != nil {
				__antithesis_instrumentation__.Notify(630149)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630150)
			}
		}
	}
	__antithesis_instrumentation__.Notify(630098)
	if len(t.TupleContents) != len(other.TupleContents) {
		__antithesis_instrumentation__.Notify(630151)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630152)
	}
	__antithesis_instrumentation__.Notify(630099)
	for i := range t.TupleContents {
		__antithesis_instrumentation__.Notify(630153)
		if !t.TupleContents[i].Identical(other.TupleContents[i]) {
			__antithesis_instrumentation__.Notify(630154)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630155)
		}
	}
	__antithesis_instrumentation__.Notify(630100)
	if len(t.TupleLabels) != len(other.TupleLabels) {
		__antithesis_instrumentation__.Notify(630156)
		return false
	} else {
		__antithesis_instrumentation__.Notify(630157)
	}
	__antithesis_instrumentation__.Notify(630101)
	for i := range t.TupleLabels {
		__antithesis_instrumentation__.Notify(630158)
		if t.TupleLabels[i] != other.TupleLabels[i] {
			__antithesis_instrumentation__.Notify(630159)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630160)
		}
	}
	__antithesis_instrumentation__.Notify(630102)
	if t.UDTMetadata != nil && func() bool {
		__antithesis_instrumentation__.Notify(630161)
		return other.UDTMetadata != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(630162)
		if t.UDTMetadata.ArrayTypeOID != other.UDTMetadata.ArrayTypeOID {
			__antithesis_instrumentation__.Notify(630163)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630164)
		}
	} else {
		__antithesis_instrumentation__.Notify(630165)
		if t.UDTMetadata != nil {
			__antithesis_instrumentation__.Notify(630166)
			return false
		} else {
			__antithesis_instrumentation__.Notify(630167)
			if other.UDTMetadata != nil {
				__antithesis_instrumentation__.Notify(630168)
				return false
			} else {
				__antithesis_instrumentation__.Notify(630169)
			}
		}
	}
	__antithesis_instrumentation__.Notify(630103)
	return t.Oid == other.Oid
}

func (t *T) Unmarshal(data []byte) error {
	__antithesis_instrumentation__.Notify(630170)

	err := protoutil.Unmarshal(data, &t.InternalType)
	if err != nil {
		__antithesis_instrumentation__.Notify(630172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(630173)
	}
	__antithesis_instrumentation__.Notify(630171)
	return t.upgradeType()
}

func (t *T) upgradeType() error {
	__antithesis_instrumentation__.Notify(630174)
	switch t.Family() {
	case IntFamily:
		__antithesis_instrumentation__.Notify(630178)

		switch t.InternalType.VisibleType {
		case visibleSMALLINT:
			__antithesis_instrumentation__.Notify(630196)
			t.InternalType.Width = 16
			t.InternalType.Oid = oid.T_int2
		case visibleINTEGER:
			__antithesis_instrumentation__.Notify(630197)
			t.InternalType.Width = 32
			t.InternalType.Oid = oid.T_int4
		case visibleBIGINT:
			__antithesis_instrumentation__.Notify(630198)
			t.InternalType.Width = 64
			t.InternalType.Oid = oid.T_int8
		case visibleBIT, visibleNONE:
			__antithesis_instrumentation__.Notify(630199)

			switch t.Width() {
			case 16:
				__antithesis_instrumentation__.Notify(630201)
				t.InternalType.Oid = oid.T_int2
			case 32:
				__antithesis_instrumentation__.Notify(630202)
				t.InternalType.Oid = oid.T_int4
			default:
				__antithesis_instrumentation__.Notify(630203)

				t.InternalType.Oid = oid.T_int8
				t.InternalType.Width = 64
			}
		default:
			__antithesis_instrumentation__.Notify(630200)
			return errors.AssertionFailedf("unexpected visible type: %d", t.InternalType.VisibleType)
		}

	case FloatFamily:
		__antithesis_instrumentation__.Notify(630179)

		switch t.InternalType.VisibleType {
		case visibleREAL:
			__antithesis_instrumentation__.Notify(630204)
			t.InternalType.Oid = oid.T_float4
			t.InternalType.Width = 32
		case visibleDOUBLE:
			__antithesis_instrumentation__.Notify(630205)
			t.InternalType.Oid = oid.T_float8
			t.InternalType.Width = 64
		case visibleNONE:
			__antithesis_instrumentation__.Notify(630206)
			switch t.Width() {
			case 32:
				__antithesis_instrumentation__.Notify(630208)
				t.InternalType.Oid = oid.T_float4
			case 64:
				__antithesis_instrumentation__.Notify(630209)
				t.InternalType.Oid = oid.T_float8
			default:
				__antithesis_instrumentation__.Notify(630210)

				if t.Precision() >= 1 && func() bool {
					__antithesis_instrumentation__.Notify(630211)
					return t.Precision() <= 24 == true
				}() == true {
					__antithesis_instrumentation__.Notify(630212)
					t.InternalType.Oid = oid.T_float4
					t.InternalType.Width = 32
				} else {
					__antithesis_instrumentation__.Notify(630213)
					t.InternalType.Oid = oid.T_float8
					t.InternalType.Width = 64
				}
			}
		default:
			__antithesis_instrumentation__.Notify(630207)
			return errors.AssertionFailedf("unexpected visible type: %d", t.InternalType.VisibleType)
		}
		__antithesis_instrumentation__.Notify(630180)

		t.InternalType.Precision = 0

	case TimestampFamily, TimestampTZFamily, TimeFamily, TimeTZFamily:
		__antithesis_instrumentation__.Notify(630181)

		if t.InternalType.Precision == -1 {
			__antithesis_instrumentation__.Notify(630214)
			t.InternalType.Precision = 0
			t.InternalType.TimePrecisionIsSet = false
		} else {
			__antithesis_instrumentation__.Notify(630215)
		}
		__antithesis_instrumentation__.Notify(630182)

		if t.InternalType.Precision > 0 {
			__antithesis_instrumentation__.Notify(630216)
			t.InternalType.TimePrecisionIsSet = true
		} else {
			__antithesis_instrumentation__.Notify(630217)
		}
	case IntervalFamily:
		__antithesis_instrumentation__.Notify(630183)

		if t.InternalType.IntervalDurationField == nil {
			__antithesis_instrumentation__.Notify(630218)
			t.InternalType.IntervalDurationField = &IntervalDurationField{}
		} else {
			__antithesis_instrumentation__.Notify(630219)
		}
		__antithesis_instrumentation__.Notify(630184)

		if t.InternalType.Precision > 0 {
			__antithesis_instrumentation__.Notify(630220)
			t.InternalType.TimePrecisionIsSet = true
		} else {
			__antithesis_instrumentation__.Notify(630221)
		}
	case StringFamily, CollatedStringFamily:
		__antithesis_instrumentation__.Notify(630185)

		switch t.InternalType.VisibleType {
		case visibleVARCHAR:
			__antithesis_instrumentation__.Notify(630222)
			t.InternalType.Oid = oid.T_varchar
		case visibleCHAR:
			__antithesis_instrumentation__.Notify(630223)
			t.InternalType.Oid = oid.T_bpchar
		case visibleQCHAR:
			__antithesis_instrumentation__.Notify(630224)
			t.InternalType.Oid = oid.T_char
		case visibleNONE:
			__antithesis_instrumentation__.Notify(630225)
			t.InternalType.Oid = oid.T_text
		default:
			__antithesis_instrumentation__.Notify(630226)
			return errors.AssertionFailedf("unexpected visible type: %d", t.InternalType.VisibleType)
		}
		__antithesis_instrumentation__.Notify(630186)
		if t.InternalType.Family == StringFamily {
			__antithesis_instrumentation__.Notify(630227)
			if t.InternalType.Locale != nil && func() bool {
				__antithesis_instrumentation__.Notify(630228)
				return len(*t.InternalType.Locale) != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(630229)
				return errors.AssertionFailedf(
					"STRING type should not have locale: %s", *t.InternalType.Locale)
			} else {
				__antithesis_instrumentation__.Notify(630230)
			}
		} else {
			__antithesis_instrumentation__.Notify(630231)
		}

	case BitFamily:
		__antithesis_instrumentation__.Notify(630187)

		switch t.InternalType.VisibleType {
		case visibleVARBIT:
			__antithesis_instrumentation__.Notify(630232)
			t.InternalType.Oid = oid.T_varbit
		case visibleNONE:
			__antithesis_instrumentation__.Notify(630233)
			t.InternalType.Oid = oid.T_bit
		default:
			__antithesis_instrumentation__.Notify(630234)
			return errors.AssertionFailedf("unexpected visible type: %d", t.InternalType.VisibleType)
		}

	case ArrayFamily:
		__antithesis_instrumentation__.Notify(630188)
		if t.ArrayContents() == nil {
			__antithesis_instrumentation__.Notify(630235)

			arrayContents := *t
			arrayContents.InternalType.Family = *t.InternalType.ArrayElemType
			arrayContents.InternalType.ArrayDimensions = nil
			arrayContents.InternalType.ArrayElemType = nil
			if err := arrayContents.upgradeType(); err != nil {
				__antithesis_instrumentation__.Notify(630237)
				return err
			} else {
				__antithesis_instrumentation__.Notify(630238)
			}
			__antithesis_instrumentation__.Notify(630236)
			t.InternalType.ArrayContents = &arrayContents
			t.InternalType.Oid = CalcArrayOid(t.ArrayContents())
		} else {
			__antithesis_instrumentation__.Notify(630239)
		}
		__antithesis_instrumentation__.Notify(630189)

		if t.ArrayContents().Family() == ArrayFamily {
			__antithesis_instrumentation__.Notify(630240)
			return errors.AssertionFailedf("nested array should never be unmarshaled")
		} else {
			__antithesis_instrumentation__.Notify(630241)
		}
		__antithesis_instrumentation__.Notify(630190)

		t.InternalType.Width = 0
		t.InternalType.Precision = 0
		t.InternalType.Locale = nil
		t.InternalType.VisibleType = 0
		t.InternalType.ArrayElemType = nil
		t.InternalType.ArrayDimensions = nil

	case int2vector:
		__antithesis_instrumentation__.Notify(630191)
		t.InternalType.Family = ArrayFamily
		t.InternalType.Width = 0
		t.InternalType.Oid = oid.T_int2vector
		t.InternalType.ArrayContents = Int2

	case oidvector:
		__antithesis_instrumentation__.Notify(630192)
		t.InternalType.Family = ArrayFamily
		t.InternalType.Oid = oid.T_oidvector
		t.InternalType.ArrayContents = Oid

	case name:
		__antithesis_instrumentation__.Notify(630193)
		if t.InternalType.Locale != nil {
			__antithesis_instrumentation__.Notify(630242)
			t.InternalType.Family = CollatedStringFamily
		} else {
			__antithesis_instrumentation__.Notify(630243)
			t.InternalType.Family = StringFamily
		}
		__antithesis_instrumentation__.Notify(630194)
		t.InternalType.Oid = oid.T_name
		if t.Width() != 0 {
			__antithesis_instrumentation__.Notify(630244)
			return errors.AssertionFailedf("name type cannot have non-zero width: %d", t.Width())
		} else {
			__antithesis_instrumentation__.Notify(630245)
		}
	default:
		__antithesis_instrumentation__.Notify(630195)
	}
	__antithesis_instrumentation__.Notify(630175)

	if t.InternalType.Oid == 0 {
		__antithesis_instrumentation__.Notify(630246)
		t.InternalType.Oid = familyToOid[t.Family()]
	} else {
		__antithesis_instrumentation__.Notify(630247)
	}
	__antithesis_instrumentation__.Notify(630176)

	t.InternalType.VisibleType = 0

	if t.InternalType.Locale == nil {
		__antithesis_instrumentation__.Notify(630248)
		t.InternalType.Locale = &emptyLocale
	} else {
		__antithesis_instrumentation__.Notify(630249)
	}
	__antithesis_instrumentation__.Notify(630177)

	return nil
}

func (t *T) Marshal() (data []byte, err error) {
	__antithesis_instrumentation__.Notify(630250)

	temp := *t
	if err := temp.downgradeType(); err != nil {
		__antithesis_instrumentation__.Notify(630252)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(630253)
	}
	__antithesis_instrumentation__.Notify(630251)
	return protoutil.Marshal(&temp.InternalType)
}

func (t *T) MarshalToSizedBuffer(data []byte) (int, error) {
	__antithesis_instrumentation__.Notify(630254)
	temp := *t
	if err := temp.downgradeType(); err != nil {
		__antithesis_instrumentation__.Notify(630256)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(630257)
	}
	__antithesis_instrumentation__.Notify(630255)
	return temp.InternalType.MarshalToSizedBuffer(data)
}

func (t *T) MarshalTo(data []byte) (int, error) {
	__antithesis_instrumentation__.Notify(630258)
	temp := *t
	if err := temp.downgradeType(); err != nil {
		__antithesis_instrumentation__.Notify(630260)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(630261)
	}
	__antithesis_instrumentation__.Notify(630259)
	return temp.InternalType.MarshalTo(data)
}

func (t *T) downgradeType() error {
	__antithesis_instrumentation__.Notify(630262)

	switch t.Family() {
	case BitFamily:
		__antithesis_instrumentation__.Notify(630265)
		if t.Oid() == oid.T_varbit {
			__antithesis_instrumentation__.Notify(630272)
			t.InternalType.VisibleType = visibleVARBIT
		} else {
			__antithesis_instrumentation__.Notify(630273)
		}

	case FloatFamily:
		__antithesis_instrumentation__.Notify(630266)
		switch t.Width() {
		case 32:
			__antithesis_instrumentation__.Notify(630274)
			t.InternalType.VisibleType = visibleREAL
		default:
			__antithesis_instrumentation__.Notify(630275)
		}

	case StringFamily, CollatedStringFamily:
		__antithesis_instrumentation__.Notify(630267)
		switch t.Oid() {
		case oid.T_text:
			__antithesis_instrumentation__.Notify(630276)

		case oid.T_varchar:
			__antithesis_instrumentation__.Notify(630277)
			t.InternalType.VisibleType = visibleVARCHAR
		case oid.T_bpchar:
			__antithesis_instrumentation__.Notify(630278)
			t.InternalType.VisibleType = visibleCHAR
		case oid.T_char:
			__antithesis_instrumentation__.Notify(630279)
			t.InternalType.VisibleType = visibleQCHAR
		case oid.T_name:
			__antithesis_instrumentation__.Notify(630280)
			t.InternalType.Family = name
		default:
			__antithesis_instrumentation__.Notify(630281)
			return errors.AssertionFailedf("unexpected Oid: %d", t.Oid())
		}

	case ArrayFamily:
		__antithesis_instrumentation__.Notify(630268)

		if t.ArrayContents().Family() == ArrayFamily {
			__antithesis_instrumentation__.Notify(630282)
			return errors.AssertionFailedf("nested array should never be marshaled")
		} else {
			__antithesis_instrumentation__.Notify(630283)
		}
		__antithesis_instrumentation__.Notify(630269)

		temp := *t.InternalType.ArrayContents
		if err := temp.downgradeType(); err != nil {
			__antithesis_instrumentation__.Notify(630284)
			return err
		} else {
			__antithesis_instrumentation__.Notify(630285)
		}
		__antithesis_instrumentation__.Notify(630270)
		t.InternalType.Width = temp.InternalType.Width
		t.InternalType.Precision = temp.InternalType.Precision
		t.InternalType.Locale = temp.InternalType.Locale
		t.InternalType.VisibleType = temp.InternalType.VisibleType
		t.InternalType.ArrayElemType = &t.InternalType.ArrayContents.InternalType.Family

		switch t.Oid() {
		case oid.T_int2vector:
			__antithesis_instrumentation__.Notify(630286)
			t.InternalType.Family = int2vector
		case oid.T_oidvector:
			__antithesis_instrumentation__.Notify(630287)
			t.InternalType.Family = oidvector
		default:
			__antithesis_instrumentation__.Notify(630288)
		}
	default:
		__antithesis_instrumentation__.Notify(630271)
	}
	__antithesis_instrumentation__.Notify(630263)

	if t.InternalType.Locale != nil && func() bool {
		__antithesis_instrumentation__.Notify(630289)
		return len(*t.InternalType.Locale) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(630290)
		t.InternalType.Locale = nil
	} else {
		__antithesis_instrumentation__.Notify(630291)
	}
	__antithesis_instrumentation__.Notify(630264)

	return nil
}

func (t *T) String() string {
	__antithesis_instrumentation__.Notify(630292)
	switch t.Family() {
	case CollatedStringFamily:
		__antithesis_instrumentation__.Notify(630294)
		if t.Locale() == "" {
			__antithesis_instrumentation__.Notify(630302)

			return fmt.Sprintf("collated%s{*}", t.Name())
		} else {
			__antithesis_instrumentation__.Notify(630303)
		}
		__antithesis_instrumentation__.Notify(630295)
		return fmt.Sprintf("collated%s{%s}", t.Name(), t.Locale())

	case ArrayFamily:
		__antithesis_instrumentation__.Notify(630296)
		switch t.Oid() {
		case oid.T_oidvector, oid.T_int2vector:
			__antithesis_instrumentation__.Notify(630304)
			return t.Name()
		default:
			__antithesis_instrumentation__.Notify(630305)
		}
		__antithesis_instrumentation__.Notify(630297)
		return t.ArrayContents().String() + "[]"

	case TupleFamily:
		__antithesis_instrumentation__.Notify(630298)
		var buf bytes.Buffer
		buf.WriteString("tuple")
		if len(t.TupleContents()) != 0 && func() bool {
			__antithesis_instrumentation__.Notify(630306)
			return !IsWildcardTupleType(t) == true
		}() == true {
			__antithesis_instrumentation__.Notify(630307)
			buf.WriteByte('{')
			for i, typ := range t.TupleContents() {
				__antithesis_instrumentation__.Notify(630309)
				if i != 0 {
					__antithesis_instrumentation__.Notify(630311)
					buf.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(630312)
				}
				__antithesis_instrumentation__.Notify(630310)
				buf.WriteString(typ.String())
				if t.TupleLabels() != nil {
					__antithesis_instrumentation__.Notify(630313)
					buf.WriteString(" AS ")
					buf.WriteString(t.InternalType.TupleLabels[i])
				} else {
					__antithesis_instrumentation__.Notify(630314)
				}
			}
			__antithesis_instrumentation__.Notify(630308)
			buf.WriteByte('}')
		} else {
			__antithesis_instrumentation__.Notify(630315)
		}
		__antithesis_instrumentation__.Notify(630299)
		return buf.String()
	case IntervalFamily, TimestampFamily, TimestampTZFamily, TimeFamily, TimeTZFamily:
		__antithesis_instrumentation__.Notify(630300)
		if t.InternalType.Precision > 0 || func() bool {
			__antithesis_instrumentation__.Notify(630316)
			return t.InternalType.TimePrecisionIsSet == true
		}() == true {
			__antithesis_instrumentation__.Notify(630317)
			return fmt.Sprintf("%s(%d)", t.Name(), t.Precision())
		} else {
			__antithesis_instrumentation__.Notify(630318)
		}
	default:
		__antithesis_instrumentation__.Notify(630301)
	}
	__antithesis_instrumentation__.Notify(630293)
	return t.Name()
}

func (t *T) MarshalText() (text []byte, err error) {
	__antithesis_instrumentation__.Notify(630319)
	var buf bytes.Buffer
	if err := proto.MarshalText(&buf, &t.InternalType); err != nil {
		__antithesis_instrumentation__.Notify(630321)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(630322)
	}
	__antithesis_instrumentation__.Notify(630320)
	return buf.Bytes(), nil
}

func (t *T) DebugString() string {
	__antithesis_instrumentation__.Notify(630323)
	if t.Family() == ArrayFamily && func() bool {
		__antithesis_instrumentation__.Notify(630325)
		return t.ArrayContents().UserDefined() == true
	}() == true {
		__antithesis_instrumentation__.Notify(630326)

		internalTypeCopy := protoutil.Clone(&t.InternalType).(*InternalType)
		internalTypeCopy.ArrayContents.TypeMeta = UserDefinedTypeMetadata{}
		return internalTypeCopy.String()
	} else {
		__antithesis_instrumentation__.Notify(630327)
	}
	__antithesis_instrumentation__.Notify(630324)
	return t.InternalType.String()
}

func (t *T) IsAmbiguous() bool {
	__antithesis_instrumentation__.Notify(630328)
	switch t.Family() {
	case UnknownFamily, AnyFamily:
		__antithesis_instrumentation__.Notify(630330)
		return true
	case CollatedStringFamily:
		__antithesis_instrumentation__.Notify(630331)
		return t.Locale() == ""
	case TupleFamily:
		__antithesis_instrumentation__.Notify(630332)
		if len(t.TupleContents()) == 0 {
			__antithesis_instrumentation__.Notify(630338)
			return true
		} else {
			__antithesis_instrumentation__.Notify(630339)
		}
		__antithesis_instrumentation__.Notify(630333)
		for i := range t.TupleContents() {
			__antithesis_instrumentation__.Notify(630340)
			if t.TupleContents()[i].IsAmbiguous() {
				__antithesis_instrumentation__.Notify(630341)
				return true
			} else {
				__antithesis_instrumentation__.Notify(630342)
			}
		}
		__antithesis_instrumentation__.Notify(630334)
		return false
	case ArrayFamily:
		__antithesis_instrumentation__.Notify(630335)
		return t.ArrayContents().IsAmbiguous()
	case EnumFamily:
		__antithesis_instrumentation__.Notify(630336)
		return t.Oid() == oid.T_anyenum
	default:
		__antithesis_instrumentation__.Notify(630337)
	}
	__antithesis_instrumentation__.Notify(630329)
	return false
}

func (t *T) IsNumeric() bool {
	__antithesis_instrumentation__.Notify(630343)
	switch t.Family() {
	case IntFamily, FloatFamily, DecimalFamily:
		__antithesis_instrumentation__.Notify(630344)
		return true
	default:
		__antithesis_instrumentation__.Notify(630345)
		return false
	}
}

func (t *T) EnumGetIdxOfPhysical(phys []byte) (int, error) {
	__antithesis_instrumentation__.Notify(630346)
	t.ensureHydratedEnum()

	reps := t.TypeMeta.EnumData.PhysicalRepresentations
	for i := range reps {
		__antithesis_instrumentation__.Notify(630348)
		if bytes.Equal(phys, reps[i]) {
			__antithesis_instrumentation__.Notify(630349)
			return i, nil
		} else {
			__antithesis_instrumentation__.Notify(630350)
		}
	}
	__antithesis_instrumentation__.Notify(630347)
	err := errors.Newf(
		"could not find %v in enum %q representation %s %s",
		phys,
		t.TypeMeta.Name.FQName(),
		t.TypeMeta.EnumData.debugString(),
		debug.Stack(),
	)
	return 0, err
}

func (t *T) EnumGetIdxOfLogical(logical string) (int, error) {
	__antithesis_instrumentation__.Notify(630351)
	t.ensureHydratedEnum()
	reps := t.TypeMeta.EnumData.LogicalRepresentations
	for i := range reps {
		__antithesis_instrumentation__.Notify(630353)
		if reps[i] == logical {
			__antithesis_instrumentation__.Notify(630354)
			return i, nil
		} else {
			__antithesis_instrumentation__.Notify(630355)
		}
	}
	__antithesis_instrumentation__.Notify(630352)
	return 0, pgerror.Newf(
		pgcode.InvalidTextRepresentation, "invalid input value for enum %s: %q", t, logical)
}

func (t *T) ensureHydratedEnum() {
	__antithesis_instrumentation__.Notify(630356)
	if t.TypeMeta.EnumData == nil {
		__antithesis_instrumentation__.Notify(630357)
		panic(errors.AssertionFailedf("use of enum metadata before hydration as an enum: %v %p", t, t))
	} else {
		__antithesis_instrumentation__.Notify(630358)
	}
}

func IsStringType(t *T) bool {
	__antithesis_instrumentation__.Notify(630359)
	switch t.Family() {
	case StringFamily, CollatedStringFamily:
		__antithesis_instrumentation__.Notify(630360)
		return true
	default:
		__antithesis_instrumentation__.Notify(630361)
		return false
	}
}

func IsValidArrayElementType(t *T) (valid bool, issueNum int) {
	__antithesis_instrumentation__.Notify(630362)
	switch t.Family() {
	default:
		__antithesis_instrumentation__.Notify(630363)
		return true, 0
	}
}

func CheckArrayElementType(t *T) error {
	__antithesis_instrumentation__.Notify(630364)
	if ok, issueNum := IsValidArrayElementType(t); !ok {
		__antithesis_instrumentation__.Notify(630366)
		return unimplemented.NewWithIssueDetailf(issueNum, t.String(),
			"arrays of %s not allowed", t)
	} else {
		__antithesis_instrumentation__.Notify(630367)
	}
	__antithesis_instrumentation__.Notify(630365)
	return nil
}

func IsDateTimeType(t *T) bool {
	__antithesis_instrumentation__.Notify(630368)
	switch t.Family() {
	case DateFamily:
		__antithesis_instrumentation__.Notify(630369)
		return true
	case TimeFamily:
		__antithesis_instrumentation__.Notify(630370)
		return true
	case TimeTZFamily:
		__antithesis_instrumentation__.Notify(630371)
		return true
	case TimestampFamily:
		__antithesis_instrumentation__.Notify(630372)
		return true
	case TimestampTZFamily:
		__antithesis_instrumentation__.Notify(630373)
		return true
	case IntervalFamily:
		__antithesis_instrumentation__.Notify(630374)
		return true
	default:
		__antithesis_instrumentation__.Notify(630375)
		return false
	}
}

func IsAdditiveType(t *T) bool {
	__antithesis_instrumentation__.Notify(630376)
	switch t.Family() {
	case IntFamily:
		__antithesis_instrumentation__.Notify(630377)
		return true
	case FloatFamily:
		__antithesis_instrumentation__.Notify(630378)
		return true
	case DecimalFamily:
		__antithesis_instrumentation__.Notify(630379)
		return true
	default:
		__antithesis_instrumentation__.Notify(630380)
		return IsDateTimeType(t)
	}
}

func IsWildcardTupleType(t *T) bool {
	__antithesis_instrumentation__.Notify(630381)
	return len(t.TupleContents()) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(630382)
		return t.TupleContents()[0].Family() == AnyFamily == true
	}() == true
}

func (t *T) collatedStringTypeSQL(isArray bool) string {
	__antithesis_instrumentation__.Notify(630383)
	var buf bytes.Buffer
	buf.WriteString(t.stringTypeSQL())
	if isArray {
		__antithesis_instrumentation__.Notify(630385)
		buf.WriteString("[] COLLATE ")
	} else {
		__antithesis_instrumentation__.Notify(630386)
		buf.WriteString(" COLLATE ")
	}
	__antithesis_instrumentation__.Notify(630384)
	lex.EncodeLocaleName(&buf, t.Locale())
	return buf.String()
}

func (t *T) stringTypeSQL() string {
	__antithesis_instrumentation__.Notify(630387)
	typName := "STRING"
	switch t.Oid() {
	case oid.T_varchar:
		__antithesis_instrumentation__.Notify(630390)
		typName = "VARCHAR"
	case oid.T_bpchar:
		__antithesis_instrumentation__.Notify(630391)
		typName = "CHAR"
	case oid.T_char:
		__antithesis_instrumentation__.Notify(630392)

		typName = `"char"`
	case oid.T_name:
		__antithesis_instrumentation__.Notify(630393)
		typName = "NAME"
	default:
		__antithesis_instrumentation__.Notify(630394)
	}
	__antithesis_instrumentation__.Notify(630388)

	if t.Width() > 0 {
		__antithesis_instrumentation__.Notify(630395)
		o := t.Oid()
		if t.Width() != 1 || func() bool {
			__antithesis_instrumentation__.Notify(630396)
			return (o != oid.T_bpchar && func() bool {
				__antithesis_instrumentation__.Notify(630397)
				return o != oid.T_char == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(630398)
			typName = fmt.Sprintf("%s(%d)", typName, t.Width())
		} else {
			__antithesis_instrumentation__.Notify(630399)
		}
	} else {
		__antithesis_instrumentation__.Notify(630400)
	}
	__antithesis_instrumentation__.Notify(630389)

	return typName
}

func (t *T) IsHydrated() bool {
	__antithesis_instrumentation__.Notify(630401)
	return t.UserDefined() && func() bool {
		__antithesis_instrumentation__.Notify(630402)
		return t.TypeMeta != (UserDefinedTypeMetadata{}) == true
	}() == true
}

var typNameLiterals map[string]*T

func init() {
	typNameLiterals = make(map[string]*T)
	for o, t := range OidToType {
		name, ok := oidext.TypeName(o)
		if !ok {
			panic(errors.AssertionFailedf("oid %d has no type name", o))
		}
		name = strings.ToLower(name)
		if _, ok := typNameLiterals[name]; !ok {
			typNameLiterals[name] = t
		}
	}
	for name, t := range unreservedTypeTokens {
		if _, ok := typNameLiterals[name]; !ok {
			typNameLiterals[name] = t
		}
	}
}

func TypeForNonKeywordTypeName(name string) (*T, bool, int) {
	__antithesis_instrumentation__.Notify(630403)
	t, ok := typNameLiterals[name]
	if ok {
		__antithesis_instrumentation__.Notify(630405)
		return t, ok, 0
	} else {
		__antithesis_instrumentation__.Notify(630406)
	}
	__antithesis_instrumentation__.Notify(630404)
	return nil, false, postgresPredefinedTypeIssues[name]
}

var (
	Serial2Type = *Int2
	Serial4Type = *Int4
	Serial8Type = *Int
)

func IsSerialType(typ *T) bool {
	__antithesis_instrumentation__.Notify(630407)

	return typ == &Serial2Type || func() bool {
		__antithesis_instrumentation__.Notify(630408)
		return typ == &Serial4Type == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(630409)
		return typ == &Serial8Type == true
	}() == true
}

var unreservedTypeTokens = map[string]*T{
	"blob":       Bytes,
	"bool":       Bool,
	"bytea":      Bytes,
	"bytes":      Bytes,
	"date":       Date,
	"float4":     Float,
	"float8":     Float,
	"inet":       INet,
	"int2":       Int2,
	"int4":       Int4,
	"int8":       Int,
	"int64":      Int,
	"int2vector": Int2Vector,
	"json":       Jsonb,
	"jsonb":      Jsonb,
	"name":       Name,
	"oid":        Oid,
	"oidvector":  OidVector,

	"regclass":     RegClass,
	"regnamespace": RegNamespace,
	"regproc":      RegProc,
	"regprocedure": RegProcedure,
	"regrole":      RegRole,
	"regtype":      RegType,

	"serial2":     &Serial2Type,
	"serial4":     &Serial4Type,
	"serial8":     &Serial8Type,
	"smallserial": &Serial2Type,
	"bigserial":   &Serial8Type,

	"string": String,
	"uuid":   Uuid,
}

var postgresPredefinedTypeIssues = map[string]int{
	"box":           21286,
	"cidr":          18846,
	"circle":        21286,
	"jsonpath":      22513,
	"line":          21286,
	"lseg":          21286,
	"macaddr":       45813,
	"macaddr8":      45813,
	"money":         41578,
	"path":          21286,
	"pg_lsn":        -1,
	"tsquery":       7821,
	"tsvector":      7821,
	"txid_snapshot": -1,
	"xml":           43355,
}

func (m *GeoMetadata) SQLString() string {
	__antithesis_instrumentation__.Notify(630410)

	if m.SRID != 0 {
		__antithesis_instrumentation__.Notify(630412)
		shapeName := strings.ToLower(m.ShapeType.String())
		if m.ShapeType == geopb.ShapeType_Unset {
			__antithesis_instrumentation__.Notify(630414)
			shapeName = "geometry"
		} else {
			__antithesis_instrumentation__.Notify(630415)
		}
		__antithesis_instrumentation__.Notify(630413)
		return fmt.Sprintf("(%s,%d)", shapeName, m.SRID)
	} else {
		__antithesis_instrumentation__.Notify(630416)
		if m.ShapeType != geopb.ShapeType_Unset {
			__antithesis_instrumentation__.Notify(630417)
			return fmt.Sprintf("(%s)", m.ShapeType)
		} else {
			__antithesis_instrumentation__.Notify(630418)
		}
	}
	__antithesis_instrumentation__.Notify(630411)
	return ""
}
