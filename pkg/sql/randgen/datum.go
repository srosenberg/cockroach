package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"math"
	"math/bits"
	"math/rand"
	"time"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geogen"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func RandDatum(rng *rand.Rand, typ *types.T, nullOk bool) tree.Datum {
	__antithesis_instrumentation__.Notify(563760)
	nullDenominator := 10
	if !nullOk {
		__antithesis_instrumentation__.Notify(563762)
		nullDenominator = 0
	} else {
		__antithesis_instrumentation__.Notify(563763)
	}
	__antithesis_instrumentation__.Notify(563761)
	return RandDatumWithNullChance(rng, typ, nullDenominator)
}

func RandDatumWithNullChance(rng *rand.Rand, typ *types.T, nullChance int) tree.Datum {
	__antithesis_instrumentation__.Notify(563764)
	if nullChance != 0 && func() bool {
		__antithesis_instrumentation__.Notify(563767)
		return rng.Intn(nullChance) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(563768)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(563769)
	}
	__antithesis_instrumentation__.Notify(563765)

	if rng.Intn(10) == 0 {
		__antithesis_instrumentation__.Notify(563770)
		if special := randInterestingDatum(rng, typ); special != nil {
			__antithesis_instrumentation__.Notify(563771)
			return special
		} else {
			__antithesis_instrumentation__.Notify(563772)
		}
	} else {
		__antithesis_instrumentation__.Notify(563773)
	}
	__antithesis_instrumentation__.Notify(563766)
	switch typ.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(563774)
		return tree.MakeDBool(rng.Intn(2) == 1)
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(563775)
		switch typ.Width() {
		case 64:
			__antithesis_instrumentation__.Notify(563817)

			return tree.NewDInt(tree.DInt(int64(rng.Uint64())))
		case 32:
			__antithesis_instrumentation__.Notify(563818)

			return tree.NewDInt(tree.DInt(int32(rng.Uint64())))
		case 16:
			__antithesis_instrumentation__.Notify(563819)

			return tree.NewDInt(tree.DInt(int16(rng.Uint64())))
		case 8:
			__antithesis_instrumentation__.Notify(563820)

			return tree.NewDInt(tree.DInt(int8(rng.Uint64())))
		default:
			__antithesis_instrumentation__.Notify(563821)
			panic(errors.AssertionFailedf("int with an unexpected width %d", typ.Width()))
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(563776)
		switch typ.Width() {
		case 64:
			__antithesis_instrumentation__.Notify(563822)
			return tree.NewDFloat(tree.DFloat(rng.NormFloat64()))
		case 32:
			__antithesis_instrumentation__.Notify(563823)
			return tree.NewDFloat(tree.DFloat(float32(rng.NormFloat64())))
		default:
			__antithesis_instrumentation__.Notify(563824)
			panic(errors.AssertionFailedf("float with an unexpected width %d", typ.Width()))
		}
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(563777)
		b := geo.NewCartesianBoundingBox().AddPoint(rng.NormFloat64(), rng.NormFloat64()).AddPoint(rng.NormFloat64(), rng.NormFloat64())
		return tree.NewDBox2D(*b)
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(563778)
		gm, err := typ.GeoMetadata()
		if err != nil {
			__antithesis_instrumentation__.Notify(563825)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(563826)
		}
		__antithesis_instrumentation__.Notify(563779)
		srid := gm.SRID
		if srid == 0 {
			__antithesis_instrumentation__.Notify(563827)
			srid = geopb.DefaultGeographySRID
		} else {
			__antithesis_instrumentation__.Notify(563828)
		}
		__antithesis_instrumentation__.Notify(563780)
		return tree.NewDGeography(geogen.RandomGeography(rng, srid))
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(563781)
		gm, err := typ.GeoMetadata()
		if err != nil {
			__antithesis_instrumentation__.Notify(563829)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(563830)
		}
		__antithesis_instrumentation__.Notify(563782)
		return tree.NewDGeometry(geogen.RandomGeometry(rng, gm.SRID))
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(563783)
		d := &tree.DDecimal{}

		d.Decimal.SetFinite(int64(rng.Uint64()), int32(rng.Intn(40)-20))
		return d
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(563784)
		d, err := pgdate.MakeDateFromUnixEpoch(int64(rng.Intn(10000)))
		if err != nil {
			__antithesis_instrumentation__.Notify(563831)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(563832)
		}
		__antithesis_instrumentation__.Notify(563785)
		return tree.NewDDate(d)
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(563786)
		return tree.MakeDTime(timeofday.Random(rng)).Round(tree.TimeFamilyPrecisionToRoundDuration(typ.Precision()))
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(563787)
		return tree.NewDTimeTZFromOffset(
			timeofday.Random(rng),

			(rng.Int31n(28*60+59)-(14*60+59))*60,
		).Round(tree.TimeFamilyPrecisionToRoundDuration(typ.Precision()))
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(563788)
		return tree.MustMakeDTimestamp(
			timeutil.Unix(rng.Int63n(2000000000), rng.Int63n(1000000)),
			tree.TimeFamilyPrecisionToRoundDuration(typ.Precision()),
		)
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(563789)
		sign := 1 - rng.Int63n(2)*2
		return &tree.DInterval{Duration: duration.MakeDuration(
			sign*rng.Int63n(25*3600*int64(1000000000)),
			sign*rng.Int63n(1000),
			sign*rng.Int63n(1000),
		)}
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(563790)
		gen := uuid.NewGenWithReader(rng)
		return tree.NewDUuid(tree.DUuid{UUID: uuid.Must(gen.NewV4())})
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(563791)
		ipAddr := ipaddr.RandIPAddr(rng)
		return tree.NewDIPAddr(tree.DIPAddr{IPAddr: ipAddr})
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(563792)
		j, err := json.Random(20, rng)
		if err != nil {
			__antithesis_instrumentation__.Notify(563833)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(563834)
		}
		__antithesis_instrumentation__.Notify(563793)
		return &tree.DJSON{JSON: j}
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(563794)
		tuple := tree.DTuple{D: make(tree.Datums, len(typ.TupleContents()))}
		for i := range typ.TupleContents() {
			__antithesis_instrumentation__.Notify(563835)
			tuple.D[i] = RandDatum(rng, typ.TupleContents()[i], true)
		}
		__antithesis_instrumentation__.Notify(563795)

		tuple.ResolvedType()
		return &tuple
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(563796)
		width := typ.Width()
		if width == 0 {
			__antithesis_instrumentation__.Notify(563836)
			width = rng.Int31n(100)
		} else {
			__antithesis_instrumentation__.Notify(563837)
		}
		__antithesis_instrumentation__.Notify(563797)
		r := bitarray.Rand(rng, uint(width))
		return &tree.DBitArray{BitArray: r}
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(563798)

		var length int
		if typ.Oid() == oid.T_char || func() bool {
			__antithesis_instrumentation__.Notify(563838)
			return typ.Oid() == oid.T_bpchar == true
		}() == true {
			__antithesis_instrumentation__.Notify(563839)
			length = 1
		} else {
			__antithesis_instrumentation__.Notify(563840)
			length = rng.Intn(10)
		}
		__antithesis_instrumentation__.Notify(563799)
		p := make([]byte, length)
		for i := range p {
			__antithesis_instrumentation__.Notify(563841)
			p[i] = byte(1 + rng.Intn(127))
		}
		__antithesis_instrumentation__.Notify(563800)
		if typ.Oid() == oid.T_name {
			__antithesis_instrumentation__.Notify(563842)
			return tree.NewDName(string(p))
		} else {
			__antithesis_instrumentation__.Notify(563843)
		}
		__antithesis_instrumentation__.Notify(563801)
		return tree.NewDString(string(p))
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(563802)
		p := make([]byte, rng.Intn(10))
		_, _ = rng.Read(p)
		return tree.NewDBytes(tree.DBytes(p))
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(563803)
		return tree.MustMakeDTimestampTZ(
			timeutil.Unix(rng.Int63n(2000000000), rng.Int63n(1000000)),
			tree.TimeFamilyPrecisionToRoundDuration(typ.Precision()),
		)
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(563804)

		var buf bytes.Buffer
		n := rng.Intn(10)
		for i := 0; i < n; i++ {
			__antithesis_instrumentation__.Notify(563844)
			var r rune
			for {
				__antithesis_instrumentation__.Notify(563846)
				r = rune(rng.Intn(unicode.MaxRune + 1))
				if !unicode.Is(unicode.C, r) {
					__antithesis_instrumentation__.Notify(563847)
					break
				} else {
					__antithesis_instrumentation__.Notify(563848)
				}
			}
			__antithesis_instrumentation__.Notify(563845)
			buf.WriteRune(r)
		}
		__antithesis_instrumentation__.Notify(563805)
		d, err := tree.NewDCollatedString(buf.String(), typ.Locale(), &tree.CollationEnvironment{})
		if err != nil {
			__antithesis_instrumentation__.Notify(563849)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(563850)
		}
		__antithesis_instrumentation__.Notify(563806)
		return d
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(563807)
		return tree.NewDOid(tree.DInt(rng.Uint32()))
	case types.UnknownFamily:
		__antithesis_instrumentation__.Notify(563808)
		return tree.DNull
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(563809)
		return RandArray(rng, typ, 0)
	case types.AnyFamily:
		__antithesis_instrumentation__.Notify(563810)
		return RandDatumWithNullChance(rng, RandType(rng), nullChance)
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(563811)

		if typ.TypeMeta.EnumData == nil {
			__antithesis_instrumentation__.Notify(563851)
			return tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(563852)
		}
		__antithesis_instrumentation__.Notify(563812)
		reps := typ.TypeMeta.EnumData.LogicalRepresentations
		if len(reps) == 0 {
			__antithesis_instrumentation__.Notify(563853)
			return tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(563854)
		}
		__antithesis_instrumentation__.Notify(563813)

		d, err := tree.MakeDEnumFromLogicalRepresentation(typ, reps[rng.Intn(len(reps))])
		if err != nil {
			__antithesis_instrumentation__.Notify(563855)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(563856)
		}
		__antithesis_instrumentation__.Notify(563814)
		return d
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(563815)
		return tree.DVoidDatum
	default:
		__antithesis_instrumentation__.Notify(563816)
		panic(errors.AssertionFailedf("invalid type %v", typ.DebugString()))
	}
}

func RandArray(rng *rand.Rand, typ *types.T, nullChance int) tree.Datum {
	__antithesis_instrumentation__.Notify(563857)
	contents := typ.ArrayContents()
	if contents.Family() == types.AnyFamily {
		__antithesis_instrumentation__.Notify(563860)
		contents = RandArrayContentsType(rng)
	} else {
		__antithesis_instrumentation__.Notify(563861)
	}
	__antithesis_instrumentation__.Notify(563858)
	arr := tree.NewDArray(contents)
	for i := 0; i < rng.Intn(10); i++ {
		__antithesis_instrumentation__.Notify(563862)
		if err := arr.Append(RandDatumWithNullChance(rng, contents, nullChance)); err != nil {
			__antithesis_instrumentation__.Notify(563863)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(563864)
		}
	}
	__antithesis_instrumentation__.Notify(563859)
	return arr
}

func randInterestingDatum(rng *rand.Rand, typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(563865)
	specials, ok := randInterestingDatums[typ.Family()]
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(563867)
		return len(specials) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(563868)
		for _, sc := range types.Scalar {
			__antithesis_instrumentation__.Notify(563870)

			if sc == typ {
				__antithesis_instrumentation__.Notify(563871)
				panic(errors.AssertionFailedf("no interesting datum for type %s found", typ.String()))
			} else {
				__antithesis_instrumentation__.Notify(563872)
			}
		}
		__antithesis_instrumentation__.Notify(563869)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(563873)
	}
	__antithesis_instrumentation__.Notify(563866)

	special := specials[rng.Intn(len(specials))]
	switch typ.Family() {
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(563874)
		switch typ.Width() {
		case 64:
			__antithesis_instrumentation__.Notify(563879)
			return special
		case 32:
			__antithesis_instrumentation__.Notify(563880)
			return tree.NewDInt(tree.DInt(int32(tree.MustBeDInt(special))))
		case 16:
			__antithesis_instrumentation__.Notify(563881)
			return tree.NewDInt(tree.DInt(int16(tree.MustBeDInt(special))))
		case 8:
			__antithesis_instrumentation__.Notify(563882)
			return tree.NewDInt(tree.DInt(int8(tree.MustBeDInt(special))))
		default:
			__antithesis_instrumentation__.Notify(563883)
			panic(errors.AssertionFailedf("int with an unexpected width %d", typ.Width()))
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(563875)
		switch typ.Width() {
		case 64:
			__antithesis_instrumentation__.Notify(563884)
			return special
		case 32:
			__antithesis_instrumentation__.Notify(563885)
			return tree.NewDFloat(tree.DFloat(float32(*special.(*tree.DFloat))))
		default:
			__antithesis_instrumentation__.Notify(563886)
			panic(errors.AssertionFailedf("float with an unexpected width %d", typ.Width()))
		}
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(563876)

		if typ.Width() == 0 || func() bool {
			__antithesis_instrumentation__.Notify(563887)
			return typ.Width() == 64 == true
		}() == true {
			__antithesis_instrumentation__.Notify(563888)
			return special
		} else {
			__antithesis_instrumentation__.Notify(563889)
		}
		__antithesis_instrumentation__.Notify(563877)
		return &tree.DBitArray{BitArray: special.(*tree.DBitArray).ToWidth(uint(typ.Width()))}

	default:
		__antithesis_instrumentation__.Notify(563878)
		return special
	}
}

const simpleRange = 10

func RandDatumSimple(rng *rand.Rand, typ *types.T) tree.Datum {
	__antithesis_instrumentation__.Notify(563890)
	datum := tree.DNull
	switch typ.Family() {
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(563892)
		datum, _ = tree.NewDBitArrayFromInt(rng.Int63n(simpleRange), uint(bits.Len(simpleRange)))
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(563893)
		if rng.Intn(2) == 1 {
			__antithesis_instrumentation__.Notify(563909)
			datum = tree.DBoolTrue
		} else {
			__antithesis_instrumentation__.Notify(563910)
			datum = tree.DBoolFalse
		}
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(563894)
		datum = tree.NewDBytes(tree.DBytes(randStringSimple(rng)))
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(563895)
		date, _ := pgdate.MakeDateFromPGEpoch(rng.Int31n(simpleRange))
		datum = tree.NewDDate(date)
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(563896)
		dd := &tree.DDecimal{}
		dd.SetInt64(rng.Int63n(simpleRange))
		datum = dd
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(563897)
		datum = tree.NewDInt(tree.DInt(rng.Intn(simpleRange)))
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(563898)
		datum = &tree.DInterval{Duration: duration.MakeDuration(
			rng.Int63n(simpleRange)*1e9,
			0,
			0,
		)}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(563899)
		datum = tree.NewDFloat(tree.DFloat(rng.Intn(simpleRange)))
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(563900)
		datum = tree.NewDIPAddr(tree.DIPAddr{
			IPAddr: ipaddr.IPAddr{
				Addr: ipaddr.Addr(uint128.FromInts(0, uint64(rng.Intn(simpleRange)))),
			},
		})
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(563901)
		datum = tree.NewDJSON(randJSONSimple(rng))
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(563902)
		datum = tree.NewDOid(tree.DInt(rng.Intn(simpleRange)))
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(563903)
		datum = tree.NewDString(randStringSimple(rng))
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(563904)
		datum = tree.MakeDTime(timeofday.New(0, rng.Intn(simpleRange), 0, 0))
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(563905)
		datum = tree.MustMakeDTimestamp(time.Date(2000, 1, 1, rng.Intn(simpleRange), 0, 0, 0, time.UTC), time.Microsecond)
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(563906)
		datum = tree.MustMakeDTimestampTZ(time.Date(2000, 1, 1, rng.Intn(simpleRange), 0, 0, 0, time.UTC), time.Microsecond)
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(563907)
		datum = tree.NewDUuid(tree.DUuid{
			UUID: uuid.FromUint128(uint128.FromInts(0, uint64(rng.Intn(simpleRange)))),
		})
	default:
		__antithesis_instrumentation__.Notify(563908)
	}
	__antithesis_instrumentation__.Notify(563891)
	return datum
}

func randStringSimple(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(563911)
	return string(rune('A' + rng.Intn(simpleRange)))
}

func randJSONSimple(rng *rand.Rand) json.JSON {
	__antithesis_instrumentation__.Notify(563912)
	switch rng.Intn(10) {
	case 0:
		__antithesis_instrumentation__.Notify(563913)
		return json.NullJSONValue
	case 1:
		__antithesis_instrumentation__.Notify(563914)
		return json.FalseJSONValue
	case 2:
		__antithesis_instrumentation__.Notify(563915)
		return json.TrueJSONValue
	case 3:
		__antithesis_instrumentation__.Notify(563916)
		return json.FromInt(rng.Intn(simpleRange))
	case 4:
		__antithesis_instrumentation__.Notify(563917)
		return json.FromString(randStringSimple(rng))
	case 5:
		__antithesis_instrumentation__.Notify(563918)
		a := json.NewArrayBuilder(0)
		for i := rng.Intn(3); i >= 0; i-- {
			__antithesis_instrumentation__.Notify(563922)
			a.Add(randJSONSimple(rng))
		}
		__antithesis_instrumentation__.Notify(563919)
		return a.Build()
	default:
		__antithesis_instrumentation__.Notify(563920)
		a := json.NewObjectBuilder(0)
		for i := rng.Intn(3); i >= 0; i-- {
			__antithesis_instrumentation__.Notify(563923)
			a.Add(randStringSimple(rng), randJSONSimple(rng))
		}
		__antithesis_instrumentation__.Notify(563921)
		return a.Build()
	}
}

var (
	randInterestingDatums = map[types.Family][]tree.Datum{
		types.BoolFamily: {
			tree.DBoolTrue,
			tree.DBoolFalse,
		},
		types.IntFamily: {
			tree.NewDInt(tree.DInt(0)),
			tree.NewDInt(tree.DInt(-1)),
			tree.NewDInt(tree.DInt(1)),
			tree.NewDInt(tree.DInt(math.MaxInt8)),
			tree.NewDInt(tree.DInt(math.MinInt8)),
			tree.NewDInt(tree.DInt(math.MaxInt16)),
			tree.NewDInt(tree.DInt(math.MinInt16)),
			tree.NewDInt(tree.DInt(math.MaxInt32)),
			tree.NewDInt(tree.DInt(math.MinInt32)),
			tree.NewDInt(tree.DInt(math.MaxInt64)),

			tree.NewDInt(tree.DInt(math.MinInt64 + 1)),
		},
		types.FloatFamily: {
			tree.NewDFloat(tree.DFloat(0)),
			tree.NewDFloat(tree.DFloat(1)),
			tree.NewDFloat(tree.DFloat(-1)),
			tree.NewDFloat(tree.DFloat(math.SmallestNonzeroFloat32)),
			tree.NewDFloat(tree.DFloat(math.MaxFloat32)),
			tree.NewDFloat(tree.DFloat(math.SmallestNonzeroFloat64)),
			tree.NewDFloat(tree.DFloat(math.MaxFloat64)),
			tree.NewDFloat(tree.DFloat(math.Inf(1))),
			tree.NewDFloat(tree.DFloat(math.Inf(-1))),
			tree.NewDFloat(tree.DFloat(math.NaN())),
		},
		types.DecimalFamily: func() []tree.Datum {
			__antithesis_instrumentation__.Notify(563924)
			var res []tree.Datum
			for _, s := range []string{
				"0",
				"1",
				"-1",
				"Inf",
				"-Inf",
				"NaN",
				"-12.34e400",
			} {
				__antithesis_instrumentation__.Notify(563926)
				d, err := tree.ParseDDecimal(s)
				if err != nil {
					__antithesis_instrumentation__.Notify(563928)
					panic(err)
				} else {
					__antithesis_instrumentation__.Notify(563929)
				}
				__antithesis_instrumentation__.Notify(563927)
				res = append(res, d)
			}
			__antithesis_instrumentation__.Notify(563925)
			return res
		}(),
		types.DateFamily: {
			tree.NewDDate(pgdate.MakeCompatibleDateFromDisk(0)),
			tree.NewDDate(pgdate.LowDate),
			tree.NewDDate(pgdate.HighDate),
			tree.NewDDate(pgdate.PosInfDate),
			tree.NewDDate(pgdate.NegInfDate),
		},
		types.TimeFamily: {
			tree.MakeDTime(timeofday.Min),
			tree.MakeDTime(timeofday.Max),
			tree.MakeDTime(timeofday.Time2400),
		},
		types.TimeTZFamily: {
			tree.DMinTimeTZ,
			tree.DMaxTimeTZ,
		},
		types.TimestampFamily: func() []tree.Datum {
			__antithesis_instrumentation__.Notify(563930)
			res := make([]tree.Datum, len(randTimestampSpecials))
			for i, t := range randTimestampSpecials {
				__antithesis_instrumentation__.Notify(563932)
				res[i] = tree.MustMakeDTimestamp(t, time.Microsecond)
			}
			__antithesis_instrumentation__.Notify(563931)
			return res
		}(),
		types.TimestampTZFamily: func() []tree.Datum {
			__antithesis_instrumentation__.Notify(563933)
			res := make([]tree.Datum, len(randTimestampSpecials))
			for i, t := range randTimestampSpecials {
				__antithesis_instrumentation__.Notify(563935)
				res[i] = tree.MustMakeDTimestampTZ(t, time.Microsecond)
			}
			__antithesis_instrumentation__.Notify(563934)
			return res
		}(),
		types.IntervalFamily: {
			&tree.DInterval{Duration: duration.MakeDuration(0, 0, 0)},
			&tree.DInterval{Duration: duration.MakeDuration(0, 1, 0)},
			&tree.DInterval{Duration: duration.MakeDuration(1, 0, 0)},
			&tree.DInterval{Duration: duration.MakeDuration(1, 1, 1)},

			&tree.DInterval{Duration: duration.MakeDuration(0, 0, 290*12)},
		},
		types.Box2DFamily: {
			&tree.DBox2D{CartesianBoundingBox: geo.CartesianBoundingBox{BoundingBox: geopb.BoundingBox{LoX: -10, HiX: 10, LoY: -10, HiY: 10}}},
		},
		types.GeographyFamily: {

			&tree.DGeography{Geography: geo.MustParseGeography("0101000000000000000000F03F000000000000F03F")},

			&tree.DGeography{Geography: geo.MustParseGeography("010200000002000000000000000000F03F000000000000F03F00000000000000400000000000000040")},

			&tree.DGeography{Geography: geo.MustParseGeography("0103000000010000000500000000000000000000000000000000000000000000000000F03F0000000000000000000000000000F03F000000000000F03F0000000000000000000000000000F03F00000000000000000000000000000000")},

			&tree.DGeography{Geography: geo.MustParseGeography("0103000000020000000500000000000000000000000000000000000000000000000000F03F0000000000000000000000000000F03F000000000000F03F0000000000000000000000000000F03F00000000000000000000000000000000050000009A9999999999C93F9A9999999999C93F9A9999999999C93F9A9999999999D93F9A9999999999D93F9A9999999999D93F9A9999999999D93F9A9999999999C93F9A9999999999C93F9A9999999999C93F")},

			&tree.DGeography{Geography: geo.MustParseGeography("010400000004000000010100000000000000000024400000000000004440010100000000000000000044400000000000003E4001010000000000000000003440000000000000344001010000000000000000003E400000000000002440")},

			&tree.DGeography{Geography: geo.MustParseGeography("010500000002000000010200000003000000000000000000244000000000000024400000000000003440000000000000344000000000000024400000000000004440010200000004000000000000000000444000000000000044400000000000003E400000000000003E40000000000000444000000000000034400000000000003E400000000000002440")},

			&tree.DGeography{Geography: geo.MustParseGeography("01060000000200000001030000000100000004000000000000000000444000000000000044400000000000003440000000000080464000000000008046400000000000003E4000000000000044400000000000004440010300000002000000060000000000000000003440000000000080414000000000000024400000000000003E40000000000000244000000000000024400000000000003E4000000000000014400000000000804640000000000000344000000000000034400000000000804140040000000000000000003E40000000000000344000000000000034400000000000002E40000000000000344000000000000039400000000000003E400000000000003440")},

			&tree.DGeography{Geography: geo.MustParseGeography("01070000000300000001010000000000000000004440000000000000244001020000000300000000000000000024400000000000002440000000000000344000000000000034400000000000002440000000000000444001030000000100000004000000000000000000444000000000000044400000000000003440000000000080464000000000008046400000000000003E4000000000000044400000000000004440")},

			&tree.DGeography{Geography: geo.MustParseGeography("0101000000000000000000F87F000000000000F87F")},

			&tree.DGeography{Geography: geo.MustParseGeography("010200000000000000")},

			&tree.DGeography{Geography: geo.MustParseGeography("010300000000000000")},

			&tree.DGeography{Geography: geo.MustParseGeography("010400000000000000")},

			&tree.DGeography{Geography: geo.MustParseGeography("010500000000000000")},

			&tree.DGeography{Geography: geo.MustParseGeography("010600000000000000")},

			&tree.DGeography{Geography: geo.MustParseGeography("010700000000000000")},
		},
		types.GeometryFamily: {

			&tree.DGeometry{Geometry: geo.MustParseGeometry("0101000000000000000000F03F000000000000F03F")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010200000002000000000000000000F03F000000000000F03F00000000000000400000000000000040")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("0103000000010000000500000000000000000000000000000000000000000000000000F03F0000000000000000000000000000F03F000000000000F03F0000000000000000000000000000F03F00000000000000000000000000000000")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("0103000000020000000500000000000000000000000000000000000000000000000000F03F0000000000000000000000000000F03F000000000000F03F0000000000000000000000000000F03F00000000000000000000000000000000050000009A9999999999C93F9A9999999999C93F9A9999999999C93F9A9999999999D93F9A9999999999D93F9A9999999999D93F9A9999999999D93F9A9999999999C93F9A9999999999C93F9A9999999999C93F")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010400000004000000010100000000000000000024400000000000004440010100000000000000000044400000000000003E4001010000000000000000003440000000000000344001010000000000000000003E400000000000002440")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010500000002000000010200000003000000000000000000244000000000000024400000000000003440000000000000344000000000000024400000000000004440010200000004000000000000000000444000000000000044400000000000003E400000000000003E40000000000000444000000000000034400000000000003E400000000000002440")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("01060000000200000001030000000100000004000000000000000000444000000000000044400000000000003440000000000080464000000000008046400000000000003E4000000000000044400000000000004440010300000002000000060000000000000000003440000000000080414000000000000024400000000000003E40000000000000244000000000000024400000000000003E4000000000000014400000000000804640000000000000344000000000000034400000000000804140040000000000000000003E40000000000000344000000000000034400000000000002E40000000000000344000000000000039400000000000003E400000000000003440")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("01070000000300000001010000000000000000004440000000000000244001020000000300000000000000000024400000000000002440000000000000344000000000000034400000000000002440000000000000444001030000000100000004000000000000000000444000000000000044400000000000003440000000000080464000000000008046400000000000003E4000000000000044400000000000004440")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("0101000000000000000000F87F000000000000F87F")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010200000000000000")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010300000000000000")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010400000000000000")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010500000000000000")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010600000000000000")},

			&tree.DGeometry{Geometry: geo.MustParseGeometry("010700000000000000")},
		},
		types.StringFamily: {
			tree.NewDString(""),
			tree.NewDString("X"),
			tree.NewDString(`"`),
			tree.NewDString(`'`),
			tree.NewDString("\x00"),
			tree.NewDString("\u2603"),
		},
		types.BytesFamily: {
			tree.NewDBytes(""),
			tree.NewDBytes("X"),
			tree.NewDBytes(`"`),
			tree.NewDBytes(`'`),
			tree.NewDBytes("\x00"),
			tree.NewDBytes("\u2603"),
			tree.NewDBytes("\xFF"),
		},
		types.OidFamily: {
			tree.NewDOid(0),
		},
		types.UuidFamily: {
			tree.DMinUUID,
			tree.DMaxUUID,
		},
		types.INetFamily: {
			tree.DMinIPAddr,
			tree.DMaxIPAddr,
		},
		types.JsonFamily: func() []tree.Datum {
			__antithesis_instrumentation__.Notify(563936)
			var res []tree.Datum
			for _, s := range []string{
				`{}`,
				`1`,
				`{"test": "json"}`,
			} {
				__antithesis_instrumentation__.Notify(563938)
				d, err := tree.ParseDJSON(s)
				if err != nil {
					__antithesis_instrumentation__.Notify(563940)
					panic(err)
				} else {
					__antithesis_instrumentation__.Notify(563941)
				}
				__antithesis_instrumentation__.Notify(563939)
				res = append(res, d)
			}
			__antithesis_instrumentation__.Notify(563937)
			return res
		}(),
		types.BitFamily: func() []tree.Datum {
			__antithesis_instrumentation__.Notify(563942)
			var res []tree.Datum
			for _, i := range []int64{
				0,
				1<<63 - 1,
			} {
				__antithesis_instrumentation__.Notify(563944)
				d, err := tree.NewDBitArrayFromInt(i, 64)
				if err != nil {
					__antithesis_instrumentation__.Notify(563946)
					panic(err)
				} else {
					__antithesis_instrumentation__.Notify(563947)
				}
				__antithesis_instrumentation__.Notify(563945)
				res = append(res, d)
			}
			__antithesis_instrumentation__.Notify(563943)
			return res
		}(),
	}
	randTimestampSpecials = []time.Time{
		{},
		time.Date(-2000, time.January, 1, 0, 0, 0, 0, time.UTC),
		time.Date(3000, time.January, 1, 0, 0, 0, 0, time.UTC),

		tree.MinSupportedTime,
		tree.MaxSupportedTime,
	}
)
