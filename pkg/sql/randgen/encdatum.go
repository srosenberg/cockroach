package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func RandDatumEncoding(rng *rand.Rand) descpb.DatumEncoding {
	__antithesis_instrumentation__.Notify(563948)
	return descpb.DatumEncoding(rng.Intn(len(descpb.DatumEncoding_value)))
}

func RandEncDatum(rng *rand.Rand) (rowenc.EncDatum, *types.T) {
	__antithesis_instrumentation__.Notify(563949)
	typ := RandEncodableType(rng)
	datum := RandDatum(rng, typ, true)
	return rowenc.DatumToEncDatum(typ, datum), typ
}

func RandSortingEncDatumSlice(rng *rand.Rand, numVals int) ([]rowenc.EncDatum, *types.T) {
	__antithesis_instrumentation__.Notify(563950)
	typ := RandSortingType(rng)
	vals := make([]rowenc.EncDatum, numVals)
	for i := range vals {
		__antithesis_instrumentation__.Notify(563952)
		vals[i] = rowenc.DatumToEncDatum(typ, RandDatum(rng, typ, true))
	}
	__antithesis_instrumentation__.Notify(563951)
	return vals, typ
}

func RandSortingEncDatumSlices(
	rng *rand.Rand, numSets, numValsPerSet int,
) ([][]rowenc.EncDatum, []*types.T) {
	__antithesis_instrumentation__.Notify(563953)
	vals := make([][]rowenc.EncDatum, numSets)
	types := make([]*types.T, numSets)
	for i := range vals {
		__antithesis_instrumentation__.Notify(563955)
		val, typ := RandSortingEncDatumSlice(rng, numValsPerSet)
		vals[i], types[i] = val, typ
	}
	__antithesis_instrumentation__.Notify(563954)
	return vals, types
}

func RandEncDatumRowOfTypes(rng *rand.Rand, types []*types.T) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(563956)
	vals := make([]rowenc.EncDatum, len(types))
	for i := range types {
		__antithesis_instrumentation__.Notify(563958)
		vals[i] = rowenc.DatumToEncDatum(types[i], RandDatum(rng, types[i], true))
	}
	__antithesis_instrumentation__.Notify(563957)
	return vals
}

func RandEncDatumRows(rng *rand.Rand, numRows, numCols int) (rowenc.EncDatumRows, []*types.T) {
	__antithesis_instrumentation__.Notify(563959)
	types := RandEncodableColumnTypes(rng, numCols)
	return RandEncDatumRowsOfTypes(rng, numRows, types), types
}

func RandEncDatumRowsOfTypes(rng *rand.Rand, numRows int, types []*types.T) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563960)
	vals := make(rowenc.EncDatumRows, numRows)
	for i := range vals {
		__antithesis_instrumentation__.Notify(563962)
		vals[i] = RandEncDatumRowOfTypes(rng, types)
	}
	__antithesis_instrumentation__.Notify(563961)
	return vals
}

func IntEncDatum(i int) rowenc.EncDatum {
	__antithesis_instrumentation__.Notify(563963)
	return rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(i))}
}

func NullEncDatum() rowenc.EncDatum {
	__antithesis_instrumentation__.Notify(563964)
	return rowenc.EncDatum{Datum: tree.DNull}
}

func GenEncDatumRowsInt(inputRows [][]int) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563965)
	rows := make(rowenc.EncDatumRows, len(inputRows))
	for i, inputRow := range inputRows {
		__antithesis_instrumentation__.Notify(563967)
		for _, x := range inputRow {
			__antithesis_instrumentation__.Notify(563968)
			if x < 0 {
				__antithesis_instrumentation__.Notify(563969)
				rows[i] = append(rows[i], NullEncDatum())
			} else {
				__antithesis_instrumentation__.Notify(563970)
				rows[i] = append(rows[i], IntEncDatum(x))
			}
		}
	}
	__antithesis_instrumentation__.Notify(563966)
	return rows
}

func MakeIntRows(numRows, numCols int) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563971)
	rows := make(rowenc.EncDatumRows, numRows)
	for i := range rows {
		__antithesis_instrumentation__.Notify(563973)
		rows[i] = make(rowenc.EncDatumRow, numCols)
		for j := 0; j < numCols; j++ {
			__antithesis_instrumentation__.Notify(563974)
			rows[i][j] = IntEncDatum(i + j)
		}
	}
	__antithesis_instrumentation__.Notify(563972)
	return rows
}

func MakeRandIntRows(rng *rand.Rand, numRows int, numCols int) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563975)
	rows := make(rowenc.EncDatumRows, numRows)
	for i := range rows {
		__antithesis_instrumentation__.Notify(563977)
		rows[i] = make(rowenc.EncDatumRow, numCols)
		for j := 0; j < numCols; j++ {
			__antithesis_instrumentation__.Notify(563978)
			rows[i][j] = IntEncDatum(rng.Int())
		}
	}
	__antithesis_instrumentation__.Notify(563976)
	return rows
}

func MakeRandIntRowsInRange(
	rng *rand.Rand, numRows int, numCols int, maxNum int, nullProbability float64,
) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563979)
	rows := make(rowenc.EncDatumRows, numRows)
	for i := range rows {
		__antithesis_instrumentation__.Notify(563981)
		rows[i] = make(rowenc.EncDatumRow, numCols)
		for j := 0; j < numCols; j++ {
			__antithesis_instrumentation__.Notify(563982)
			rows[i][j] = IntEncDatum(rng.Intn(maxNum))
			if rng.Float64() < nullProbability {
				__antithesis_instrumentation__.Notify(563983)
				rows[i][j] = NullEncDatum()
			} else {
				__antithesis_instrumentation__.Notify(563984)
			}
		}
	}
	__antithesis_instrumentation__.Notify(563980)
	return rows
}

func MakeRepeatedIntRows(n int, numRows int, numCols int) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563985)
	rows := make(rowenc.EncDatumRows, numRows)
	for i := range rows {
		__antithesis_instrumentation__.Notify(563987)
		rows[i] = make(rowenc.EncDatumRow, numCols)
		for j := 0; j < numCols; j++ {
			__antithesis_instrumentation__.Notify(563988)
			rows[i][j] = IntEncDatum(i/n + j)
		}
	}
	__antithesis_instrumentation__.Notify(563986)
	return rows
}

func GenEncDatumRowsString(inputRows [][]string) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563989)
	rows := make(rowenc.EncDatumRows, len(inputRows))
	for i, inputRow := range inputRows {
		__antithesis_instrumentation__.Notify(563991)
		for _, x := range inputRow {
			__antithesis_instrumentation__.Notify(563992)
			if x == "" {
				__antithesis_instrumentation__.Notify(563993)
				rows[i] = append(rows[i], NullEncDatum())
			} else {
				__antithesis_instrumentation__.Notify(563994)
				rows[i] = append(rows[i], stringEncDatum(x))
			}
		}
	}
	__antithesis_instrumentation__.Notify(563990)
	return rows
}

func stringEncDatum(s string) rowenc.EncDatum {
	__antithesis_instrumentation__.Notify(563995)
	return rowenc.EncDatum{Datum: tree.NewDString(s)}
}

func GenEncDatumRowsBytes(inputRows [][][]byte) rowenc.EncDatumRows {
	__antithesis_instrumentation__.Notify(563996)
	rows := make(rowenc.EncDatumRows, len(inputRows))
	for i, inputRow := range inputRows {
		__antithesis_instrumentation__.Notify(563998)
		for _, x := range inputRow {
			__antithesis_instrumentation__.Notify(563999)
			rows[i] = append(rows[i], rowenc.EncDatumFromEncoded(descpb.DatumEncoding_VALUE, x))
		}
	}
	__antithesis_instrumentation__.Notify(563997)
	return rows
}
