package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/lib/pq/oid"
)

var (
	SeedTypes []*types.T

	arrayContentsTypes []*types.T
	collationLocales   = [...]string{"da", "de", "en"}
)

func init() {
	for _, typ := range types.OidToType {
		switch typ.Oid() {
		case oid.T_unknown, oid.T_anyelement:

		case oid.T_anyarray, oid.T_oidvector, oid.T_int2vector:

			SeedTypes = append(SeedTypes, typ)
		default:

			if typ.Family() != types.ArrayFamily {
				SeedTypes = append(SeedTypes, typ)
			}
		}
	}

	for _, typ := range types.OidToType {
		if IsAllowedForArray(typ) {
			arrayContentsTypes = append(arrayContentsTypes, typ)
		}
	}

	sort.Slice(SeedTypes, func(i, j int) bool {
		return SeedTypes[i].String() < SeedTypes[j].String()
	})
	sort.Slice(arrayContentsTypes, func(i, j int) bool {
		return arrayContentsTypes[i].String() < arrayContentsTypes[j].String()
	})
}

func IsAllowedForArray(typ *types.T) bool {
	__antithesis_instrumentation__.Notify(564702)

	encTyp, err := valueside.DatumTypeToArrayElementEncodingType(typ)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(564705)
		return encTyp == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(564706)
		return false
	} else {
		__antithesis_instrumentation__.Notify(564707)
	}
	__antithesis_instrumentation__.Notify(564703)

	if typ.Family() == types.OidFamily && func() bool {
		__antithesis_instrumentation__.Notify(564708)
		return typ.Oid() != oid.T_oid == true
	}() == true {
		__antithesis_instrumentation__.Notify(564709)
		return false
	} else {
		__antithesis_instrumentation__.Notify(564710)
	}
	__antithesis_instrumentation__.Notify(564704)

	return true
}

func RandType(rng *rand.Rand) *types.T {
	__antithesis_instrumentation__.Notify(564711)
	return RandTypeFromSlice(rng, SeedTypes)
}

func RandArrayContentsType(rng *rand.Rand) *types.T {
	__antithesis_instrumentation__.Notify(564712)
	return RandTypeFromSlice(rng, arrayContentsTypes)
}

func RandTypeFromSlice(rng *rand.Rand, typs []*types.T) *types.T {
	__antithesis_instrumentation__.Notify(564713)
	typ := typs[rng.Intn(len(typs))]
	switch typ.Family() {
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(564715)
		return types.MakeBit(int32(rng.Intn(50)))
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(564716)
		return types.MakeCollatedString(types.String, *RandCollationLocale(rng))
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(564717)
		if typ.ArrayContents().Family() == types.AnyFamily {
			__antithesis_instrumentation__.Notify(564722)
			inner := RandArrayContentsType(rng)
			if inner.Family() == types.CollatedStringFamily {
				__antithesis_instrumentation__.Notify(564724)

				inner = types.String
			} else {
				__antithesis_instrumentation__.Notify(564725)
			}
			__antithesis_instrumentation__.Notify(564723)
			return types.MakeArray(inner)
		} else {
			__antithesis_instrumentation__.Notify(564726)
		}
		__antithesis_instrumentation__.Notify(564718)
		if typ.ArrayContents().Family() == types.TupleFamily {
			__antithesis_instrumentation__.Notify(564727)

			len := rng.Intn(5)
			contents := make([]*types.T, len)
			for i := range contents {
				__antithesis_instrumentation__.Notify(564729)
				contents[i] = RandTypeFromSlice(rng, typs)
			}
			__antithesis_instrumentation__.Notify(564728)
			return types.MakeArray(types.MakeTuple(contents))
		} else {
			__antithesis_instrumentation__.Notify(564730)
		}
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(564719)

		len := rng.Intn(5)
		contents := make([]*types.T, len)
		for i := range contents {
			__antithesis_instrumentation__.Notify(564731)
			contents[i] = RandTypeFromSlice(rng, typs)
		}
		__antithesis_instrumentation__.Notify(564720)
		return types.MakeTuple(contents)
	default:
		__antithesis_instrumentation__.Notify(564721)
	}
	__antithesis_instrumentation__.Notify(564714)
	return typ
}

func RandColumnType(rng *rand.Rand) *types.T {
	__antithesis_instrumentation__.Notify(564732)
	for {
		__antithesis_instrumentation__.Notify(564733)
		typ := RandType(rng)
		if IsLegalColumnType(typ) {
			__antithesis_instrumentation__.Notify(564734)
			return typ
		} else {
			__antithesis_instrumentation__.Notify(564735)
		}
	}
}

func IsLegalColumnType(typ *types.T) bool {
	__antithesis_instrumentation__.Notify(564736)
	switch typ.Oid() {
	case oid.T_int2vector, oid.T_oidvector:
		__antithesis_instrumentation__.Notify(564738)

		return false
	default:
		__antithesis_instrumentation__.Notify(564739)
	}
	__antithesis_instrumentation__.Notify(564737)
	return colinfo.ValidateColumnDefType(typ) == nil
}

func RandArrayType(rng *rand.Rand) *types.T {
	__antithesis_instrumentation__.Notify(564740)
	for {
		__antithesis_instrumentation__.Notify(564741)
		typ := RandColumnType(rng)
		resTyp := types.MakeArray(typ)
		if err := colinfo.ValidateColumnDefType(resTyp); err == nil {
			__antithesis_instrumentation__.Notify(564742)
			return resTyp
		} else {
			__antithesis_instrumentation__.Notify(564743)
		}
	}
}

func RandColumnTypes(rng *rand.Rand, numCols int) []*types.T {
	__antithesis_instrumentation__.Notify(564744)
	types := make([]*types.T, numCols)
	for i := range types {
		__antithesis_instrumentation__.Notify(564746)
		types[i] = RandColumnType(rng)
	}
	__antithesis_instrumentation__.Notify(564745)
	return types
}

func RandSortingType(rng *rand.Rand) *types.T {
	__antithesis_instrumentation__.Notify(564747)
	typ := RandType(rng)
	for colinfo.MustBeValueEncoded(typ) || func() bool {
		__antithesis_instrumentation__.Notify(564749)
		return typ == types.Void == true
	}() == true {
		__antithesis_instrumentation__.Notify(564750)
		typ = RandType(rng)
	}
	__antithesis_instrumentation__.Notify(564748)
	return typ
}

func RandSortingTypes(rng *rand.Rand, numCols int) []*types.T {
	__antithesis_instrumentation__.Notify(564751)
	types := make([]*types.T, numCols)
	for i := range types {
		__antithesis_instrumentation__.Notify(564753)
		types[i] = RandSortingType(rng)
	}
	__antithesis_instrumentation__.Notify(564752)
	return types
}

func RandCollationLocale(rng *rand.Rand) *string {
	__antithesis_instrumentation__.Notify(564754)
	return &collationLocales[rng.Intn(len(collationLocales))]
}

func RandEncodableType(rng *rand.Rand) *types.T {
	__antithesis_instrumentation__.Notify(564755)
	var isEncodableType func(t *types.T) bool
	isEncodableType = func(t *types.T) bool {
		__antithesis_instrumentation__.Notify(564757)
		switch t.Family() {
		case types.ArrayFamily:
			__antithesis_instrumentation__.Notify(564759)

			if t.ArrayContents().Oid() == oid.T_name {
				__antithesis_instrumentation__.Notify(564764)
				return false
			} else {
				__antithesis_instrumentation__.Notify(564765)
			}
			__antithesis_instrumentation__.Notify(564760)
			return isEncodableType(t.ArrayContents())

		case types.TupleFamily:
			__antithesis_instrumentation__.Notify(564761)
			for i := range t.TupleContents() {
				__antithesis_instrumentation__.Notify(564766)
				if !isEncodableType(t.TupleContents()[i]) {
					__antithesis_instrumentation__.Notify(564767)
					return false
				} else {
					__antithesis_instrumentation__.Notify(564768)
				}
			}

		case types.VoidFamily:
			__antithesis_instrumentation__.Notify(564762)
			return false
		default:
			__antithesis_instrumentation__.Notify(564763)

		}
		__antithesis_instrumentation__.Notify(564758)
		return true
	}
	__antithesis_instrumentation__.Notify(564756)

	for {
		__antithesis_instrumentation__.Notify(564769)
		typ := RandType(rng)
		if isEncodableType(typ) {
			__antithesis_instrumentation__.Notify(564770)
			return typ
		} else {
			__antithesis_instrumentation__.Notify(564771)
		}
	}
}

func RandEncodableColumnTypes(rng *rand.Rand, numCols int) []*types.T {
	__antithesis_instrumentation__.Notify(564772)
	types := make([]*types.T, numCols)
	for i := range types {
		__antithesis_instrumentation__.Notify(564774)
		for {
			__antithesis_instrumentation__.Notify(564775)
			types[i] = RandEncodableType(rng)
			if err := colinfo.ValidateColumnDefType(types[i]); err == nil {
				__antithesis_instrumentation__.Notify(564776)
				break
			} else {
				__antithesis_instrumentation__.Notify(564777)
			}
		}
	}
	__antithesis_instrumentation__.Notify(564773)
	return types
}
