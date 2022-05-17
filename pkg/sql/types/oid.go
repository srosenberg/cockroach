package types

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/oidext"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

var (
	Oid = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_oid, Locale: &emptyLocale}}

	RegClass = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_regclass, Locale: &emptyLocale}}

	RegNamespace = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_regnamespace, Locale: &emptyLocale}}

	RegProc = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_regproc, Locale: &emptyLocale}}

	RegProcedure = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_regprocedure, Locale: &emptyLocale}}

	RegRole = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_regrole, Locale: &emptyLocale}}

	RegType = &T{InternalType: InternalType{
		Family: OidFamily, Oid: oid.T_regtype, Locale: &emptyLocale}}

	OidVector = &T{InternalType: InternalType{
		Family: ArrayFamily, Oid: oid.T_oidvector, ArrayContents: Oid, Locale: &emptyLocale}}
)

var OidToType = map[oid.Oid]*T{
	oid.T_anyelement:   Any,
	oid.T_bit:          typeBit,
	oid.T_bool:         Bool,
	oid.T_bpchar:       typeBpChar,
	oid.T_bytea:        Bytes,
	oid.T_char:         QChar,
	oid.T_date:         Date,
	oid.T_float4:       Float4,
	oid.T_float8:       Float,
	oid.T_int2:         Int2,
	oid.T_int2vector:   Int2Vector,
	oid.T_int4:         Int4,
	oid.T_int8:         Int,
	oid.T_inet:         INet,
	oid.T_interval:     Interval,
	oid.T_jsonb:        Jsonb,
	oid.T_name:         Name,
	oid.T_numeric:      Decimal,
	oid.T_oid:          Oid,
	oid.T_oidvector:    OidVector,
	oid.T_record:       AnyTuple,
	oid.T_regclass:     RegClass,
	oid.T_regnamespace: RegNamespace,
	oid.T_regproc:      RegProc,
	oid.T_regprocedure: RegProcedure,
	oid.T_regrole:      RegRole,
	oid.T_regtype:      RegType,
	oid.T_text:         String,
	oid.T_time:         Time,
	oid.T_timetz:       TimeTZ,
	oid.T_timestamp:    Timestamp,
	oid.T_timestamptz:  TimestampTZ,
	oid.T_unknown:      Unknown,
	oid.T_uuid:         Uuid,
	oid.T_varbit:       VarBit,
	oid.T_varchar:      VarChar,
	oid.T_void:         Void,

	oidext.T_geometry:  Geometry,
	oidext.T_geography: Geography,
	oidext.T_box2d:     Box2D,
}

var oidToArrayOid = map[oid.Oid]oid.Oid{
	oid.T_anyelement:   oid.T_anyarray,
	oid.T_bit:          oid.T__bit,
	oid.T_bool:         oid.T__bool,
	oid.T_bpchar:       oid.T__bpchar,
	oid.T_bytea:        oid.T__bytea,
	oid.T_char:         oid.T__char,
	oid.T_date:         oid.T__date,
	oid.T_float4:       oid.T__float4,
	oid.T_float8:       oid.T__float8,
	oid.T_inet:         oid.T__inet,
	oid.T_int2:         oid.T__int2,
	oid.T_int2vector:   oid.T__int2vector,
	oid.T_int4:         oid.T__int4,
	oid.T_int8:         oid.T__int8,
	oid.T_interval:     oid.T__interval,
	oid.T_jsonb:        oid.T__jsonb,
	oid.T_name:         oid.T__name,
	oid.T_numeric:      oid.T__numeric,
	oid.T_oid:          oid.T__oid,
	oid.T_oidvector:    oid.T__oidvector,
	oid.T_record:       oid.T__record,
	oid.T_regclass:     oid.T__regclass,
	oid.T_regnamespace: oid.T__regnamespace,
	oid.T_regproc:      oid.T__regproc,
	oid.T_regprocedure: oid.T__regprocedure,
	oid.T_regrole:      oid.T__regrole,
	oid.T_regtype:      oid.T__regtype,
	oid.T_text:         oid.T__text,
	oid.T_time:         oid.T__time,
	oid.T_timetz:       oid.T__timetz,
	oid.T_timestamp:    oid.T__timestamp,
	oid.T_timestamptz:  oid.T__timestamptz,
	oid.T_uuid:         oid.T__uuid,
	oid.T_varbit:       oid.T__varbit,
	oid.T_varchar:      oid.T__varchar,

	oidext.T_geometry:  oidext.T__geometry,
	oidext.T_geography: oidext.T__geography,
	oidext.T_box2d:     oidext.T__box2d,
}

var familyToOid = map[Family]oid.Oid{
	BoolFamily:           oid.T_bool,
	IntFamily:            oid.T_int8,
	FloatFamily:          oid.T_float8,
	DecimalFamily:        oid.T_numeric,
	DateFamily:           oid.T_date,
	TimestampFamily:      oid.T_timestamp,
	IntervalFamily:       oid.T_interval,
	StringFamily:         oid.T_text,
	BytesFamily:          oid.T_bytea,
	TimestampTZFamily:    oid.T_timestamptz,
	CollatedStringFamily: oid.T_text,
	OidFamily:            oid.T_oid,
	UnknownFamily:        oid.T_unknown,
	UuidFamily:           oid.T_uuid,
	ArrayFamily:          oid.T_anyarray,
	INetFamily:           oid.T_inet,
	TimeFamily:           oid.T_time,
	TimeTZFamily:         oid.T_timetz,
	JsonFamily:           oid.T_jsonb,
	TupleFamily:          oid.T_record,
	BitFamily:            oid.T_bit,
	AnyFamily:            oid.T_anyelement,

	GeometryFamily:  oidext.T_geometry,
	GeographyFamily: oidext.T_geography,
	Box2DFamily:     oidext.T_box2d,
}

var ArrayOids = map[oid.Oid]struct{}{}

func init() {
	for o, ao := range oidToArrayOid {
		ArrayOids[ao] = struct{}{}
		OidToType[ao] = MakeArray(OidToType[o])
	}
}

func CalcArrayOid(elemTyp *T) oid.Oid {
	__antithesis_instrumentation__.Notify(629555)
	o := elemTyp.Oid()
	switch elemTyp.Family() {
	case ArrayFamily:
		__antithesis_instrumentation__.Notify(629558)

		switch o {
		case oid.T_int2vector, oid.T_oidvector:
			__antithesis_instrumentation__.Notify(629563)

		default:
			__antithesis_instrumentation__.Notify(629564)
			return o
		}

	case UnknownFamily:
		__antithesis_instrumentation__.Notify(629559)

		return unknownArrayOid

	case EnumFamily:
		__antithesis_instrumentation__.Notify(629560)
		return elemTyp.UserDefinedArrayOID()

	case TupleFamily:
		__antithesis_instrumentation__.Notify(629561)
		if elemTyp.UserDefined() {
			__antithesis_instrumentation__.Notify(629565)

			return oid.T__record
		} else {
			__antithesis_instrumentation__.Notify(629566)
		}
	default:
		__antithesis_instrumentation__.Notify(629562)
	}
	__antithesis_instrumentation__.Notify(629556)

	o = oidToArrayOid[o]
	if o == 0 {
		__antithesis_instrumentation__.Notify(629567)
		panic(errors.AssertionFailedf("oid %d couldn't be mapped to array oid", o))
	} else {
		__antithesis_instrumentation__.Notify(629568)
	}
	__antithesis_instrumentation__.Notify(629557)
	return o
}
