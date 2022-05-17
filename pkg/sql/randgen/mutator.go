package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
)

var (
	StatisticsMutator MultiStatementMutation = statisticsMutator

	ForeignKeyMutator MultiStatementMutation = foreignKeyMutator

	ColumnFamilyMutator StatementMutator = columnFamilyMutator

	IndexStoringMutator MultiStatementMutation = indexStoringMutator

	PartialIndexMutator MultiStatementMutation = partialIndexMutator

	PostgresMutator StatementStringMutator = postgresMutator

	PostgresCreateTableMutator MultiStatementMutation = postgresCreateTableMutator
)

var (
	_ = IndexStoringMutator
	_ = PostgresCreateTableMutator
)

type StatementMutator func(rng *rand.Rand, stmt tree.Statement) (changed bool)

type MultiStatementMutation func(rng *rand.Rand, stmts []tree.Statement) (mutated []tree.Statement, changed bool)

type Mutator interface {
	Mutate(rng *rand.Rand, stmts []tree.Statement) (mutated []tree.Statement, changed bool)
}

func (sm StatementMutator) Mutate(
	rng *rand.Rand, stmts []tree.Statement,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564070)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564072)
		sc := sm(rng, stmt)
		changed = changed || func() bool {
			__antithesis_instrumentation__.Notify(564073)
			return sc == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(564071)
	return stmts, changed
}

func (msm MultiStatementMutation) Mutate(
	rng *rand.Rand, stmts []tree.Statement,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564074)
	return msm(rng, stmts)
}

func Apply(
	rng *rand.Rand, stmts []tree.Statement, mutators ...Mutator,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564075)
	var mc bool
	for _, m := range mutators {
		__antithesis_instrumentation__.Notify(564077)
		stmts, mc = m.Mutate(rng, stmts)
		changed = changed || func() bool {
			__antithesis_instrumentation__.Notify(564078)
			return mc == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(564076)
	return stmts, changed
}

type StringMutator interface {
	MutateString(*rand.Rand, string) (mutated string, changed bool)
}

type StatementStringMutator func(*rand.Rand, string) string

func (sm StatementStringMutator) Mutate(
	rng *rand.Rand, stmts []tree.Statement,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564079)
	panic("can only be used with MutateString")
}

func (sm StatementStringMutator) MutateString(
	rng *rand.Rand, q string,
) (mutated string, changed bool) {
	__antithesis_instrumentation__.Notify(564080)
	newq := sm(rng, q)
	return newq, newq != q
}

func ApplyString(rng *rand.Rand, input string, mutators ...Mutator) (output string, changed bool) {
	__antithesis_instrumentation__.Notify(564081)
	parsed, err := parser.Parse(input)
	if err != nil {
		__antithesis_instrumentation__.Notify(564087)
		return input, false
	} else {
		__antithesis_instrumentation__.Notify(564088)
	}
	__antithesis_instrumentation__.Notify(564082)

	stmts := make([]tree.Statement, len(parsed))
	for i, p := range parsed {
		__antithesis_instrumentation__.Notify(564089)
		stmts[i] = p.AST
	}
	__antithesis_instrumentation__.Notify(564083)

	var normalMutators []Mutator
	var stringMutators []StringMutator
	for _, m := range mutators {
		__antithesis_instrumentation__.Notify(564090)
		if sm, ok := m.(StringMutator); ok {
			__antithesis_instrumentation__.Notify(564091)
			stringMutators = append(stringMutators, sm)
		} else {
			__antithesis_instrumentation__.Notify(564092)
			normalMutators = append(normalMutators, m)
		}
	}
	__antithesis_instrumentation__.Notify(564084)
	stmts, changed = Apply(rng, stmts, normalMutators...)
	if changed {
		__antithesis_instrumentation__.Notify(564093)
		var sb strings.Builder
		for _, s := range stmts {
			__antithesis_instrumentation__.Notify(564095)
			sb.WriteString(tree.Serialize(s))
			sb.WriteString(";\n")
		}
		__antithesis_instrumentation__.Notify(564094)
		input = sb.String()
	} else {
		__antithesis_instrumentation__.Notify(564096)
	}
	__antithesis_instrumentation__.Notify(564085)
	for _, m := range stringMutators {
		__antithesis_instrumentation__.Notify(564097)
		s, ch := m.MutateString(rng, input)
		if ch {
			__antithesis_instrumentation__.Notify(564098)
			input = s
			changed = true
		} else {
			__antithesis_instrumentation__.Notify(564099)
		}
	}
	__antithesis_instrumentation__.Notify(564086)
	return input, changed
}

func randNonNegInt(rng *rand.Rand) int64 {
	__antithesis_instrumentation__.Notify(564100)
	var v int64
	if n := rng.Intn(20); n == 0 {
		__antithesis_instrumentation__.Notify(564102)

	} else {
		__antithesis_instrumentation__.Notify(564103)
		if n <= 10 {
			__antithesis_instrumentation__.Notify(564104)
			v = rng.Int63n(10) + 1
			for i := 0; i < n; i++ {
				__antithesis_instrumentation__.Notify(564105)
				v *= 10
			}
		} else {
			__antithesis_instrumentation__.Notify(564106)
			v = rng.Int63()
		}
	}
	__antithesis_instrumentation__.Notify(564101)
	return v
}

func statisticsMutator(
	rng *rand.Rand, stmts []tree.Statement,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564107)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564109)
		create, ok := stmt.(*tree.CreateTable)
		if !ok {
			__antithesis_instrumentation__.Notify(564113)
			continue
		} else {
			__antithesis_instrumentation__.Notify(564114)
		}
		__antithesis_instrumentation__.Notify(564110)
		alter := &tree.AlterTable{
			Table: create.Table.ToUnresolvedObjectName(),
		}
		rowCount := randNonNegInt(rng)
		cols := map[tree.Name]*tree.ColumnTableDef{}
		colStats := map[tree.Name]*stats.JSONStatistic{}
		makeHistogram := func(col *tree.ColumnTableDef) {
			__antithesis_instrumentation__.Notify(564115)

			if col == nil {
				__antithesis_instrumentation__.Notify(564118)
				return
			} else {
				__antithesis_instrumentation__.Notify(564119)
			}
			__antithesis_instrumentation__.Notify(564116)

			if rng.Intn(5) == 0 {
				__antithesis_instrumentation__.Notify(564120)
				return
			} else {
				__antithesis_instrumentation__.Notify(564121)
			}
			__antithesis_instrumentation__.Notify(564117)
			colType := tree.MustBeStaticallyKnownType(col.Type)
			h := randHistogram(rng, colType)
			stat := colStats[col.Name]
			if err := stat.SetHistogram(&h); err != nil {
				__antithesis_instrumentation__.Notify(564122)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(564123)
			}
		}
		__antithesis_instrumentation__.Notify(564111)
		for _, def := range create.Defs {
			__antithesis_instrumentation__.Notify(564124)
			switch def := def.(type) {
			case *tree.ColumnTableDef:
				__antithesis_instrumentation__.Notify(564125)
				var nullCount, distinctCount uint64
				if rowCount > 0 {
					__antithesis_instrumentation__.Notify(564129)
					if def.Nullable.Nullability != tree.NotNull {
						__antithesis_instrumentation__.Notify(564131)
						nullCount = uint64(rng.Int63n(rowCount))
					} else {
						__antithesis_instrumentation__.Notify(564132)
					}
					__antithesis_instrumentation__.Notify(564130)
					distinctCount = uint64(rng.Int63n(rowCount))
				} else {
					__antithesis_instrumentation__.Notify(564133)
				}
				__antithesis_instrumentation__.Notify(564126)
				cols[def.Name] = def
				colStats[def.Name] = &stats.JSONStatistic{
					Name:          "__auto__",
					CreatedAt:     "2000-01-01 00:00:00+00:00",
					RowCount:      uint64(rowCount),
					Columns:       []string{def.Name.String()},
					DistinctCount: distinctCount,
					NullCount:     nullCount,
				}
				if (def.Unique.IsUnique && func() bool {
					__antithesis_instrumentation__.Notify(564134)
					return !def.Unique.WithoutIndex == true
				}() == true) || func() bool {
					__antithesis_instrumentation__.Notify(564135)
					return def.PrimaryKey.IsPrimaryKey == true
				}() == true {
					__antithesis_instrumentation__.Notify(564136)
					makeHistogram(def)
				} else {
					__antithesis_instrumentation__.Notify(564137)
				}
			case *tree.IndexTableDef:
				__antithesis_instrumentation__.Notify(564127)

				makeHistogram(cols[def.Columns[0].Column])
			case *tree.UniqueConstraintTableDef:
				__antithesis_instrumentation__.Notify(564128)
				if !def.WithoutIndex {
					__antithesis_instrumentation__.Notify(564138)

					makeHistogram(cols[def.Columns[0].Column])
				} else {
					__antithesis_instrumentation__.Notify(564139)
				}
			}
		}
		__antithesis_instrumentation__.Notify(564112)
		if len(colStats) > 0 {
			__antithesis_instrumentation__.Notify(564140)
			var allStats []*stats.JSONStatistic
			for _, cs := range colStats {
				__antithesis_instrumentation__.Notify(564144)
				allStats = append(allStats, cs)
			}
			__antithesis_instrumentation__.Notify(564141)
			b, err := json.Marshal(allStats)
			if err != nil {
				__antithesis_instrumentation__.Notify(564145)

				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(564146)
			}
			__antithesis_instrumentation__.Notify(564142)
			j, err := tree.ParseDJSON(string(b))
			if err != nil {
				__antithesis_instrumentation__.Notify(564147)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(564148)
			}
			__antithesis_instrumentation__.Notify(564143)
			alter.Cmds = append(alter.Cmds, &tree.AlterTableInjectStats{
				Stats: j,
			})
			stmts = append(stmts, alter)
			changed = true
		} else {
			__antithesis_instrumentation__.Notify(564149)
		}
	}
	__antithesis_instrumentation__.Notify(564108)
	return stmts, changed
}

func randHistogram(rng *rand.Rand, colType *types.T) stats.HistogramData {
	__antithesis_instrumentation__.Notify(564150)
	histogramColType := colType
	if colinfo.ColumnTypeIsInvertedIndexable(colType) {
		__antithesis_instrumentation__.Notify(564157)
		histogramColType = types.Bytes
	} else {
		__antithesis_instrumentation__.Notify(564158)
	}
	__antithesis_instrumentation__.Notify(564151)
	h := stats.HistogramData{
		ColumnType: histogramColType,
	}

	var encodedUpperBounds [][]byte
	for i, numDatums := 0, rng.Intn(10); i < numDatums; i++ {
		__antithesis_instrumentation__.Notify(564159)
		upper := RandDatum(rng, colType, false)
		if colinfo.ColumnTypeIsInvertedIndexable(colType) {
			__antithesis_instrumentation__.Notify(564160)
			encs := encodeInvertedIndexHistogramUpperBounds(colType, upper)
			encodedUpperBounds = append(encodedUpperBounds, encs...)
		} else {
			__antithesis_instrumentation__.Notify(564161)
			enc, err := keyside.Encode(nil, upper, encoding.Ascending)
			if err != nil {
				__antithesis_instrumentation__.Notify(564163)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(564164)
			}
			__antithesis_instrumentation__.Notify(564162)
			encodedUpperBounds = append(encodedUpperBounds, enc)
		}
	}
	__antithesis_instrumentation__.Notify(564152)

	if len(encodedUpperBounds) == 0 {
		__antithesis_instrumentation__.Notify(564165)
		return h
	} else {
		__antithesis_instrumentation__.Notify(564166)
	}
	__antithesis_instrumentation__.Notify(564153)

	sort.Slice(encodedUpperBounds, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(564167)
		return bytes.Compare(encodedUpperBounds[i], encodedUpperBounds[j]) < 0
	})
	__antithesis_instrumentation__.Notify(564154)

	dedupIdx := 1
	for i := 1; i < len(encodedUpperBounds); i++ {
		__antithesis_instrumentation__.Notify(564168)
		if !bytes.Equal(encodedUpperBounds[i], encodedUpperBounds[i-1]) {
			__antithesis_instrumentation__.Notify(564169)
			encodedUpperBounds[dedupIdx] = encodedUpperBounds[i]
			dedupIdx++
		} else {
			__antithesis_instrumentation__.Notify(564170)
		}
	}
	__antithesis_instrumentation__.Notify(564155)
	encodedUpperBounds = encodedUpperBounds[:dedupIdx]

	for i := range encodedUpperBounds {
		__antithesis_instrumentation__.Notify(564171)

		var numRange int64
		var distinctRange float64
		if i > 0 {
			__antithesis_instrumentation__.Notify(564173)
			numRange, distinctRange = randNumRangeAndDistinctRange(rng)
		} else {
			__antithesis_instrumentation__.Notify(564174)
		}
		__antithesis_instrumentation__.Notify(564172)

		h.Buckets = append(h.Buckets, stats.HistogramData_Bucket{
			NumEq:         randNonNegInt(rng),
			NumRange:      numRange,
			DistinctRange: distinctRange,
			UpperBound:    encodedUpperBounds[i],
		})
	}
	__antithesis_instrumentation__.Notify(564156)

	return h
}

func encodeInvertedIndexHistogramUpperBounds(colType *types.T, val tree.Datum) (encs [][]byte) {
	__antithesis_instrumentation__.Notify(564175)
	var keys [][]byte
	var err error
	switch colType.Family() {
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(564179)
		keys, err = rowenc.EncodeGeoInvertedIndexTableKeys(val, nil, *geoindex.DefaultGeometryIndexConfig())
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(564180)
		keys, err = rowenc.EncodeGeoInvertedIndexTableKeys(val, nil, *geoindex.DefaultGeographyIndexConfig())
	default:
		__antithesis_instrumentation__.Notify(564181)
		keys, err = rowenc.EncodeInvertedIndexTableKeys(val, nil, descpb.LatestIndexDescriptorVersion)
	}
	__antithesis_instrumentation__.Notify(564176)

	if err != nil {
		__antithesis_instrumentation__.Notify(564182)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(564183)
	}
	__antithesis_instrumentation__.Notify(564177)

	var da tree.DatumAlloc
	for i := range keys {
		__antithesis_instrumentation__.Notify(564184)

		enc, err := keyside.Encode(nil, da.NewDBytes(tree.DBytes(keys[i])), encoding.Ascending)
		if err != nil {
			__antithesis_instrumentation__.Notify(564186)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(564187)
		}
		__antithesis_instrumentation__.Notify(564185)
		encs = append(encs, enc)
	}
	__antithesis_instrumentation__.Notify(564178)
	return encs
}

func randNumRangeAndDistinctRange(rng *rand.Rand) (numRange int64, distinctRange float64) {
	__antithesis_instrumentation__.Notify(564188)
	numRange = randNonNegInt(rng)

	switch rng.Intn(3) {
	case 0:
		__antithesis_instrumentation__.Notify(564190)
		distinctRange = 0
	case 1:
		__antithesis_instrumentation__.Notify(564191)
		distinctRange = float64(numRange)
	default:
		__antithesis_instrumentation__.Notify(564192)
		distinctRange = rng.Float64() * float64(numRange)
	}
	__antithesis_instrumentation__.Notify(564189)
	return numRange, distinctRange
}

func foreignKeyMutator(
	rng *rand.Rand, stmts []tree.Statement,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564193)

	cols := map[tree.TableName][]*tree.ColumnTableDef{}
	byName := map[tree.TableName]*tree.CreateTable{}

	usedCols := map[tree.TableName]map[tree.Name]bool{}

	dependsOn := map[tree.TableName]map[tree.TableName]bool{}

	var tables []*tree.CreateTable
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564198)
		table, ok := stmt.(*tree.CreateTable)
		if !ok {
			__antithesis_instrumentation__.Notify(564200)
			continue
		} else {
			__antithesis_instrumentation__.Notify(564201)
		}
		__antithesis_instrumentation__.Notify(564199)
		tables = append(tables, table)
		byName[table.Table] = table
		usedCols[table.Table] = map[tree.Name]bool{}
		dependsOn[table.Table] = map[tree.TableName]bool{}
		for _, def := range table.Defs {
			__antithesis_instrumentation__.Notify(564202)
			switch def := def.(type) {
			case *tree.ColumnTableDef:
				__antithesis_instrumentation__.Notify(564203)
				cols[table.Table] = append(cols[table.Table], def)
			}
		}
	}
	__antithesis_instrumentation__.Notify(564194)
	if len(tables) == 0 {
		__antithesis_instrumentation__.Notify(564204)
		return stmts, false
	} else {
		__antithesis_instrumentation__.Notify(564205)
	}
	__antithesis_instrumentation__.Notify(564195)

	toNames := func(cols []*tree.ColumnTableDef) tree.NameList {
		__antithesis_instrumentation__.Notify(564206)
		names := make(tree.NameList, len(cols))
		for i, c := range cols {
			__antithesis_instrumentation__.Notify(564208)
			names[i] = c.Name
		}
		__antithesis_instrumentation__.Notify(564207)
		return names
	}
	__antithesis_instrumentation__.Notify(564196)

	for rng.Intn(2) == 0 {
		__antithesis_instrumentation__.Notify(564209)

		table := tables[rng.Intn(len(tables))]

		var fkCols []*tree.ColumnTableDef
		for _, c := range cols[table.Table] {
			__antithesis_instrumentation__.Notify(564214)
			if c.Computed.Computed {
				__antithesis_instrumentation__.Notify(564217)

				continue
			} else {
				__antithesis_instrumentation__.Notify(564218)
			}
			__antithesis_instrumentation__.Notify(564215)
			if usedCols[table.Table][c.Name] {
				__antithesis_instrumentation__.Notify(564219)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564220)
			}
			__antithesis_instrumentation__.Notify(564216)
			fkCols = append(fkCols, c)
		}
		__antithesis_instrumentation__.Notify(564210)
		if len(fkCols) == 0 {
			__antithesis_instrumentation__.Notify(564221)
			continue
		} else {
			__antithesis_instrumentation__.Notify(564222)
		}
		__antithesis_instrumentation__.Notify(564211)
		rng.Shuffle(len(fkCols), func(i, j int) {
			__antithesis_instrumentation__.Notify(564223)
			fkCols[i], fkCols[j] = fkCols[j], fkCols[i]
		})
		__antithesis_instrumentation__.Notify(564212)

		i := 1
		for len(fkCols) > i && func() bool {
			__antithesis_instrumentation__.Notify(564224)
			return rng.Intn(2) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(564225)
			i++
		}
		__antithesis_instrumentation__.Notify(564213)
		fkCols = fkCols[:i]

	LoopTable:
		for refTable, refCols := range cols {
			__antithesis_instrumentation__.Notify(564226)

			if refTable == table.Table || func() bool {
				__antithesis_instrumentation__.Notify(564234)
				return len(refCols) < len(fkCols) == true
			}() == true {
				__antithesis_instrumentation__.Notify(564235)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564236)
			}

			{
				__antithesis_instrumentation__.Notify(564237)

				stack := []tree.TableName{refTable}
				for i := 0; i < len(stack); i++ {
					__antithesis_instrumentation__.Notify(564238)
					curTable := stack[i]
					if curTable == table.Table {
						__antithesis_instrumentation__.Notify(564240)

						continue LoopTable
					} else {
						__antithesis_instrumentation__.Notify(564241)
					}
					__antithesis_instrumentation__.Notify(564239)
					for t := range dependsOn[curTable] {
						__antithesis_instrumentation__.Notify(564242)
						stack = append(stack, t)
					}
				}
			}
			__antithesis_instrumentation__.Notify(564227)

			availCols := append([]*tree.ColumnTableDef(nil), refCols...)
			var usingCols []*tree.ColumnTableDef
			for len(availCols) > 0 && func() bool {
				__antithesis_instrumentation__.Notify(564243)
				return len(usingCols) < len(fkCols) == true
			}() == true {
				__antithesis_instrumentation__.Notify(564244)
				fkCol := fkCols[len(usingCols)]
				found := false
				for refI, refCol := range availCols {
					__antithesis_instrumentation__.Notify(564246)
					if refCol.Computed.Virtual {
						__antithesis_instrumentation__.Notify(564248)

						continue
					} else {
						__antithesis_instrumentation__.Notify(564249)
					}
					__antithesis_instrumentation__.Notify(564247)
					fkColType := tree.MustBeStaticallyKnownType(fkCol.Type)
					refColType := tree.MustBeStaticallyKnownType(refCol.Type)
					if fkColType.Equivalent(refColType) && func() bool {
						__antithesis_instrumentation__.Notify(564250)
						return colinfo.ColumnTypeIsIndexable(refColType) == true
					}() == true {
						__antithesis_instrumentation__.Notify(564251)
						usingCols = append(usingCols, refCol)
						availCols = append(availCols[:refI], availCols[refI+1:]...)
						found = true
						break
					} else {
						__antithesis_instrumentation__.Notify(564252)
					}
				}
				__antithesis_instrumentation__.Notify(564245)
				if !found {
					__antithesis_instrumentation__.Notify(564253)
					continue LoopTable
				} else {
					__antithesis_instrumentation__.Notify(564254)
				}
			}
			__antithesis_instrumentation__.Notify(564228)

			if len(usingCols) != len(fkCols) {
				__antithesis_instrumentation__.Notify(564255)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564256)
			}
			__antithesis_instrumentation__.Notify(564229)

			ref := byName[refTable]
			refColumns := make(tree.IndexElemList, len(usingCols))
			for i, c := range usingCols {
				__antithesis_instrumentation__.Notify(564257)
				refColumns[i].Column = c.Name
			}
			__antithesis_instrumentation__.Notify(564230)
			for _, c := range fkCols {
				__antithesis_instrumentation__.Notify(564258)
				usedCols[table.Table][c.Name] = true
			}
			__antithesis_instrumentation__.Notify(564231)
			dependsOn[table.Table][ref.Table] = true
			ref.Defs = append(ref.Defs, &tree.UniqueConstraintTableDef{
				IndexTableDef: tree.IndexTableDef{
					Columns: refColumns,
				},
			})

			match := tree.MatchSimple

			var actions tree.ReferenceActions
			if rng.Intn(2) == 0 {
				__antithesis_instrumentation__.Notify(564259)
				actions.Delete = randAction(rng, table)
			} else {
				__antithesis_instrumentation__.Notify(564260)
			}
			__antithesis_instrumentation__.Notify(564232)
			if rng.Intn(2) == 0 {
				__antithesis_instrumentation__.Notify(564261)
				actions.Update = randAction(rng, table)
			} else {
				__antithesis_instrumentation__.Notify(564262)
			}
			__antithesis_instrumentation__.Notify(564233)
			stmts = append(stmts, &tree.AlterTable{
				Table: table.Table.ToUnresolvedObjectName(),
				Cmds: tree.AlterTableCmds{&tree.AlterTableAddConstraint{
					ConstraintDef: &tree.ForeignKeyConstraintTableDef{
						Table:    ref.Table,
						FromCols: toNames(fkCols),
						ToCols:   toNames(usingCols),
						Actions:  actions,
						Match:    match,
					},
				}},
			})
			changed = true
			break
		}
	}
	__antithesis_instrumentation__.Notify(564197)

	return stmts, changed
}

func randAction(rng *rand.Rand, table *tree.CreateTable) tree.ReferenceAction {
	__antithesis_instrumentation__.Notify(564263)
	const highestAction = tree.Cascade

Loop:
	for {
		__antithesis_instrumentation__.Notify(564264)
		action := tree.ReferenceAction(rng.Intn(int(highestAction + 1)))
		for _, def := range table.Defs {
			__antithesis_instrumentation__.Notify(564266)
			col, ok := def.(*tree.ColumnTableDef)
			if !ok {
				__antithesis_instrumentation__.Notify(564268)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564269)
			}
			__antithesis_instrumentation__.Notify(564267)
			switch action {
			case tree.SetNull:
				__antithesis_instrumentation__.Notify(564270)
				if col.Nullable.Nullability == tree.NotNull {
					__antithesis_instrumentation__.Notify(564273)
					continue Loop
				} else {
					__antithesis_instrumentation__.Notify(564274)
				}
			case tree.SetDefault:
				__antithesis_instrumentation__.Notify(564271)
				if col.DefaultExpr.Expr == nil && func() bool {
					__antithesis_instrumentation__.Notify(564275)
					return col.Nullable.Nullability == tree.NotNull == true
				}() == true {
					__antithesis_instrumentation__.Notify(564276)
					continue Loop
				} else {
					__antithesis_instrumentation__.Notify(564277)
				}
			default:
				__antithesis_instrumentation__.Notify(564272)
			}
		}
		__antithesis_instrumentation__.Notify(564265)
		return action
	}
}

var postgresMutatorAtIndex = regexp.MustCompile(`@[\[\]\w]+`)

func postgresMutator(rng *rand.Rand, q string) string {
	__antithesis_instrumentation__.Notify(564278)
	q, _ = ApplyString(rng, q, postgresStatementMutator)

	for from, to := range map[string]string{
		":::":     "::",
		"STRING":  "TEXT",
		"BYTES":   "BYTEA",
		"STORING": "INCLUDE",
		" AS (":   " GENERATED ALWAYS AS (",
		",)":      ")",
	} {
		__antithesis_instrumentation__.Notify(564280)
		q = strings.Replace(q, from, to, -1)
	}
	__antithesis_instrumentation__.Notify(564279)
	q = postgresMutatorAtIndex.ReplaceAllString(q, "")
	return q
}

var postgresStatementMutator MultiStatementMutation = func(rng *rand.Rand, stmts []tree.Statement) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564281)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564283)
		switch stmt := stmt.(type) {
		case *tree.SetClusterSetting, *tree.SetVar, *tree.AlterTenantSetClusterSetting:
			__antithesis_instrumentation__.Notify(564285)
			continue
		case *tree.CreateTable:
			__antithesis_instrumentation__.Notify(564286)
			if stmt.PartitionByTable != nil {
				__antithesis_instrumentation__.Notify(564290)
				stmt.PartitionByTable = nil
				changed = true
			} else {
				__antithesis_instrumentation__.Notify(564291)
			}
			__antithesis_instrumentation__.Notify(564287)
			for i := 0; i < len(stmt.Defs); i++ {
				__antithesis_instrumentation__.Notify(564292)
				switch def := stmt.Defs[i].(type) {
				case *tree.FamilyTableDef:
					__antithesis_instrumentation__.Notify(564293)

					stmt.Defs = append(stmt.Defs[:i], stmt.Defs[i+1:]...)
					i--
					changed = true
				case *tree.ColumnTableDef:
					__antithesis_instrumentation__.Notify(564294)
					if def.HasColumnFamily() {
						__antithesis_instrumentation__.Notify(564299)
						def.Family.Name = ""
						def.Family.Create = false
						changed = true
					} else {
						__antithesis_instrumentation__.Notify(564300)
					}
					__antithesis_instrumentation__.Notify(564295)
					if def.Unique.WithoutIndex {
						__antithesis_instrumentation__.Notify(564301)
						def.Unique.WithoutIndex = false
						changed = true
					} else {
						__antithesis_instrumentation__.Notify(564302)
					}
					__antithesis_instrumentation__.Notify(564296)
					if def.IsVirtual() {
						__antithesis_instrumentation__.Notify(564303)
						def.Computed.Virtual = false
						def.Computed.Computed = true
						changed = true
					} else {
						__antithesis_instrumentation__.Notify(564304)
					}
				case *tree.UniqueConstraintTableDef:
					__antithesis_instrumentation__.Notify(564297)
					if def.PartitionByIndex != nil {
						__antithesis_instrumentation__.Notify(564305)
						def.PartitionByIndex = nil
						changed = true
					} else {
						__antithesis_instrumentation__.Notify(564306)
					}
					__antithesis_instrumentation__.Notify(564298)
					if def.WithoutIndex {
						__antithesis_instrumentation__.Notify(564307)
						def.WithoutIndex = false
						changed = true
					} else {
						__antithesis_instrumentation__.Notify(564308)
					}
				}
			}
		case *tree.AlterTable:
			__antithesis_instrumentation__.Notify(564288)
			for i := 0; i < len(stmt.Cmds); i++ {
				__antithesis_instrumentation__.Notify(564309)

				if _, ok := stmt.Cmds[i].(*tree.AlterTableInjectStats); ok {
					__antithesis_instrumentation__.Notify(564310)
					stmt.Cmds = append(stmt.Cmds[:i], stmt.Cmds[i+1:]...)
					i--
					changed = true
				} else {
					__antithesis_instrumentation__.Notify(564311)
				}
			}
			__antithesis_instrumentation__.Notify(564289)

			if len(stmt.Cmds) == 0 {
				__antithesis_instrumentation__.Notify(564312)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564313)
			}
		}
		__antithesis_instrumentation__.Notify(564284)
		mutated = append(mutated, stmt)
	}
	__antithesis_instrumentation__.Notify(564282)
	return mutated, changed
}

func postgresCreateTableMutator(
	rng *rand.Rand, stmts []tree.Statement,
) (mutated []tree.Statement, changed bool) {
	__antithesis_instrumentation__.Notify(564314)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564316)
		mutated = append(mutated, stmt)
		switch stmt := stmt.(type) {
		case *tree.CreateTable:
			__antithesis_instrumentation__.Notify(564317)

			colTypes := make(map[string]*types.T)
			for _, def := range stmt.Defs {
				__antithesis_instrumentation__.Notify(564320)
				switch def := def.(type) {
				case *tree.ColumnTableDef:
					__antithesis_instrumentation__.Notify(564321)
					colTypes[string(def.Name)] = tree.MustBeStaticallyKnownType(def.Type)
				}
			}
			__antithesis_instrumentation__.Notify(564318)

			var newdefs tree.TableDefs
			for _, def := range stmt.Defs {
				__antithesis_instrumentation__.Notify(564322)
				switch def := def.(type) {
				case *tree.IndexTableDef:
					__antithesis_instrumentation__.Notify(564323)

					var newCols tree.IndexElemList
					for _, col := range def.Columns {
						__antithesis_instrumentation__.Notify(564331)
						isBox2d := false

						if col.Expr == nil {
							__antithesis_instrumentation__.Notify(564333)

							colTypeFamily := colTypes[string(col.Column)].Family()
							if colTypeFamily == types.Box2DFamily {
								__antithesis_instrumentation__.Notify(564334)
								isBox2d = true
							} else {
								__antithesis_instrumentation__.Notify(564335)
							}
						} else {
							__antithesis_instrumentation__.Notify(564336)
						}
						__antithesis_instrumentation__.Notify(564332)
						if isBox2d {
							__antithesis_instrumentation__.Notify(564337)
							changed = true
						} else {
							__antithesis_instrumentation__.Notify(564338)
							newCols = append(newCols, col)
						}
					}
					__antithesis_instrumentation__.Notify(564324)
					if len(newCols) == 0 {
						__antithesis_instrumentation__.Notify(564339)

						break
					} else {
						__antithesis_instrumentation__.Notify(564340)
					}
					__antithesis_instrumentation__.Notify(564325)
					def.Columns = newCols

					if !def.Inverted {
						__antithesis_instrumentation__.Notify(564341)
						mutated = append(mutated, &tree.CreateIndex{
							Name:     def.Name,
							Table:    stmt.Table,
							Inverted: def.Inverted,
							Columns:  newCols,
							Storing:  def.Storing,
						})
						changed = true
					} else {
						__antithesis_instrumentation__.Notify(564342)
					}
				case *tree.UniqueConstraintTableDef:
					__antithesis_instrumentation__.Notify(564326)
					var newCols tree.IndexElemList
					for _, col := range def.Columns {
						__antithesis_instrumentation__.Notify(564343)
						isBox2d := false

						if col.Expr == nil {
							__antithesis_instrumentation__.Notify(564345)

							colTypeFamily := colTypes[string(col.Column)].Family()
							if colTypeFamily == types.Box2DFamily {
								__antithesis_instrumentation__.Notify(564346)
								isBox2d = true
							} else {
								__antithesis_instrumentation__.Notify(564347)
							}
						} else {
							__antithesis_instrumentation__.Notify(564348)
						}
						__antithesis_instrumentation__.Notify(564344)
						if isBox2d {
							__antithesis_instrumentation__.Notify(564349)
							changed = true
						} else {
							__antithesis_instrumentation__.Notify(564350)
							newCols = append(newCols, col)
						}
					}
					__antithesis_instrumentation__.Notify(564327)
					if len(newCols) == 0 {
						__antithesis_instrumentation__.Notify(564351)

						break
					} else {
						__antithesis_instrumentation__.Notify(564352)
					}
					__antithesis_instrumentation__.Notify(564328)
					def.Columns = newCols
					if def.PrimaryKey {
						__antithesis_instrumentation__.Notify(564353)
						for i, col := range def.Columns {
							__antithesis_instrumentation__.Notify(564356)

							if col.Direction != tree.DefaultDirection {
								__antithesis_instrumentation__.Notify(564357)
								def.Columns[i].Direction = tree.DefaultDirection
								changed = true
							} else {
								__antithesis_instrumentation__.Notify(564358)
							}
						}
						__antithesis_instrumentation__.Notify(564354)
						if def.Name != "" {
							__antithesis_instrumentation__.Notify(564359)

							def.Name = ""
							changed = true
						} else {
							__antithesis_instrumentation__.Notify(564360)
						}
						__antithesis_instrumentation__.Notify(564355)
						newdefs = append(newdefs, def)
						break
					} else {
						__antithesis_instrumentation__.Notify(564361)
					}
					__antithesis_instrumentation__.Notify(564329)
					mutated = append(mutated, &tree.CreateIndex{
						Name:     def.Name,
						Table:    stmt.Table,
						Unique:   true,
						Inverted: def.Inverted,
						Columns:  newCols,
						Storing:  def.Storing,
					})
					changed = true
				default:
					__antithesis_instrumentation__.Notify(564330)
					newdefs = append(newdefs, def)
				}
			}
			__antithesis_instrumentation__.Notify(564319)
			stmt.Defs = newdefs
		}
	}
	__antithesis_instrumentation__.Notify(564315)
	return mutated, changed
}

func columnFamilyMutator(rng *rand.Rand, stmt tree.Statement) (changed bool) {
	__antithesis_instrumentation__.Notify(564362)
	ast, ok := stmt.(*tree.CreateTable)
	if !ok {
		__antithesis_instrumentation__.Notify(564368)
		return false
	} else {
		__antithesis_instrumentation__.Notify(564369)
	}
	__antithesis_instrumentation__.Notify(564363)

	var columns []tree.Name
	for _, def := range ast.Defs {
		__antithesis_instrumentation__.Notify(564370)
		switch def := def.(type) {
		case *tree.FamilyTableDef:
			__antithesis_instrumentation__.Notify(564371)
			return false
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(564372)
			if def.HasColumnFamily() {
				__antithesis_instrumentation__.Notify(564374)
				return false
			} else {
				__antithesis_instrumentation__.Notify(564375)
			}
			__antithesis_instrumentation__.Notify(564373)
			if !def.Computed.Virtual {
				__antithesis_instrumentation__.Notify(564376)
				columns = append(columns, def.Name)
			} else {
				__antithesis_instrumentation__.Notify(564377)
			}
		}
	}
	__antithesis_instrumentation__.Notify(564364)

	if len(columns) <= 1 {
		__antithesis_instrumentation__.Notify(564378)
		return false
	} else {
		__antithesis_instrumentation__.Notify(564379)
	}
	__antithesis_instrumentation__.Notify(564365)

	rng.Shuffle(len(columns), func(i, j int) {
		__antithesis_instrumentation__.Notify(564380)
		columns[i], columns[j] = columns[j], columns[i]
	})
	__antithesis_instrumentation__.Notify(564366)
	fd := &tree.FamilyTableDef{}
	for {
		__antithesis_instrumentation__.Notify(564381)
		if len(columns) == 0 {
			__antithesis_instrumentation__.Notify(564383)
			if len(fd.Columns) > 0 {
				__antithesis_instrumentation__.Notify(564385)
				ast.Defs = append(ast.Defs, fd)
			} else {
				__antithesis_instrumentation__.Notify(564386)
			}
			__antithesis_instrumentation__.Notify(564384)
			break
		} else {
			__antithesis_instrumentation__.Notify(564387)
		}
		__antithesis_instrumentation__.Notify(564382)
		fd.Columns = append(fd.Columns, columns[0])
		columns = columns[1:]

		if rng.Intn(2) != 0 {
			__antithesis_instrumentation__.Notify(564388)
			ast.Defs = append(ast.Defs, fd)
			fd = &tree.FamilyTableDef{}
		} else {
			__antithesis_instrumentation__.Notify(564389)
		}
	}
	__antithesis_instrumentation__.Notify(564367)
	return true
}

type tableInfo struct {
	columnNames      []tree.Name
	columnsTableDefs []*tree.ColumnTableDef
	pkCols           []tree.Name
	refColsLists     [][]tree.Name
}

func getTableInfoFromDDLStatements(stmts []tree.Statement) map[tree.Name]tableInfo {
	__antithesis_instrumentation__.Notify(564390)
	tables := make(map[tree.Name]tableInfo)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564392)
		switch ast := stmt.(type) {
		case *tree.CreateTable:
			__antithesis_instrumentation__.Notify(564393)
			info := tableInfo{}
			for _, def := range ast.Defs {
				__antithesis_instrumentation__.Notify(564396)
				switch ast := def.(type) {
				case *tree.ColumnTableDef:
					__antithesis_instrumentation__.Notify(564397)
					info.columnNames = append(info.columnNames, ast.Name)
					info.columnsTableDefs = append(info.columnsTableDefs, ast)
					if ast.PrimaryKey.IsPrimaryKey {
						__antithesis_instrumentation__.Notify(564400)
						info.pkCols = []tree.Name{ast.Name}
					} else {
						__antithesis_instrumentation__.Notify(564401)
					}
				case *tree.UniqueConstraintTableDef:
					__antithesis_instrumentation__.Notify(564398)
					if ast.PrimaryKey {
						__antithesis_instrumentation__.Notify(564402)
						for _, elem := range ast.Columns {
							__antithesis_instrumentation__.Notify(564403)
							info.pkCols = append(info.pkCols, elem.Column)
						}
					} else {
						__antithesis_instrumentation__.Notify(564404)
					}
				case *tree.ForeignKeyConstraintTableDef:
					__antithesis_instrumentation__.Notify(564399)

					if refTableInfo, ok := tables[ast.Table.ObjectName]; ok {
						__antithesis_instrumentation__.Notify(564405)
						refTableInfo.refColsLists = append(refTableInfo.refColsLists, ast.ToCols)
						tables[ast.Table.ObjectName] = refTableInfo
					} else {
						__antithesis_instrumentation__.Notify(564406)
					}
				}
			}
			__antithesis_instrumentation__.Notify(564394)
			tables[ast.Table.ObjectName] = info
		case *tree.AlterTable:
			__antithesis_instrumentation__.Notify(564395)
			for _, cmd := range ast.Cmds {
				__antithesis_instrumentation__.Notify(564407)
				switch alterCmd := cmd.(type) {
				case *tree.AlterTableAddConstraint:
					__antithesis_instrumentation__.Notify(564408)
					switch constraintDef := alterCmd.ConstraintDef.(type) {
					case *tree.ForeignKeyConstraintTableDef:
						__antithesis_instrumentation__.Notify(564409)

						if info, ok := tables[constraintDef.Table.ObjectName]; ok {
							__antithesis_instrumentation__.Notify(564410)
							info.refColsLists = append(info.refColsLists, constraintDef.ToCols)
							tables[constraintDef.Table.ObjectName] = info
						} else {
							__antithesis_instrumentation__.Notify(564411)
						}
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(564391)
	return tables
}

func indexStoringMutator(rng *rand.Rand, stmts []tree.Statement) ([]tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(564412)
	changed := false
	tables := getTableInfoFromDDLStatements(stmts)
	mapFromIndexCols := func(cols []tree.Name) map[tree.Name]struct{} {
		__antithesis_instrumentation__.Notify(564416)
		colMap := map[tree.Name]struct{}{}
		for _, col := range cols {
			__antithesis_instrumentation__.Notify(564418)
			colMap[col] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(564417)
		return colMap
	}
	__antithesis_instrumentation__.Notify(564413)
	generateStoringCols := func(rng *rand.Rand, tableInfo tableInfo, indexCols map[tree.Name]struct{}) []tree.Name {
		__antithesis_instrumentation__.Notify(564419)
		var storingCols []tree.Name
		for colOrdinal, col := range tableInfo.columnNames {
			__antithesis_instrumentation__.Notify(564421)
			if _, ok := indexCols[col]; ok {
				__antithesis_instrumentation__.Notify(564424)

				continue
			} else {
				__antithesis_instrumentation__.Notify(564425)
			}
			__antithesis_instrumentation__.Notify(564422)
			if tableInfo.columnsTableDefs[colOrdinal].Computed.Virtual {
				__antithesis_instrumentation__.Notify(564426)

				continue
			} else {
				__antithesis_instrumentation__.Notify(564427)
			}
			__antithesis_instrumentation__.Notify(564423)
			if rng.Intn(2) == 0 {
				__antithesis_instrumentation__.Notify(564428)
				storingCols = append(storingCols, col)
			} else {
				__antithesis_instrumentation__.Notify(564429)
			}
		}
		__antithesis_instrumentation__.Notify(564420)
		return storingCols
	}
	__antithesis_instrumentation__.Notify(564414)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564430)
		switch ast := stmt.(type) {
		case *tree.CreateIndex:
			__antithesis_instrumentation__.Notify(564431)
			if ast.Inverted {
				__antithesis_instrumentation__.Notify(564436)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564437)
			}
			__antithesis_instrumentation__.Notify(564432)
			info, ok := tables[ast.Table.ObjectName]
			if !ok {
				__antithesis_instrumentation__.Notify(564438)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564439)
			}
			__antithesis_instrumentation__.Notify(564433)

			if ast.Storing == nil && func() bool {
				__antithesis_instrumentation__.Notify(564440)
				return rng.Intn(2) == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(564441)
				indexCols := mapFromIndexCols(info.pkCols)
				for _, elem := range ast.Columns {
					__antithesis_instrumentation__.Notify(564443)
					indexCols[elem.Column] = struct{}{}
				}
				__antithesis_instrumentation__.Notify(564442)
				ast.Storing = generateStoringCols(rng, info, indexCols)
				changed = true
			} else {
				__antithesis_instrumentation__.Notify(564444)
			}
		case *tree.CreateTable:
			__antithesis_instrumentation__.Notify(564434)
			info, ok := tables[ast.Table.ObjectName]
			if !ok {
				__antithesis_instrumentation__.Notify(564445)
				panic("table info could not be found")
			} else {
				__antithesis_instrumentation__.Notify(564446)
			}
			__antithesis_instrumentation__.Notify(564435)
			for _, def := range ast.Defs {
				__antithesis_instrumentation__.Notify(564447)
				var idx *tree.IndexTableDef
				switch defType := def.(type) {
				case *tree.IndexTableDef:
					__antithesis_instrumentation__.Notify(564450)
					idx = defType
				case *tree.UniqueConstraintTableDef:
					__antithesis_instrumentation__.Notify(564451)
					if !defType.PrimaryKey && func() bool {
						__antithesis_instrumentation__.Notify(564452)
						return !defType.WithoutIndex == true
					}() == true {
						__antithesis_instrumentation__.Notify(564453)
						idx = &defType.IndexTableDef
					} else {
						__antithesis_instrumentation__.Notify(564454)
					}
				}
				__antithesis_instrumentation__.Notify(564448)
				if idx == nil || func() bool {
					__antithesis_instrumentation__.Notify(564455)
					return idx.Inverted == true
				}() == true {
					__antithesis_instrumentation__.Notify(564456)
					continue
				} else {
					__antithesis_instrumentation__.Notify(564457)
				}
				__antithesis_instrumentation__.Notify(564449)

				if idx.Storing == nil && func() bool {
					__antithesis_instrumentation__.Notify(564458)
					return rng.Intn(2) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(564459)
					indexCols := mapFromIndexCols(info.pkCols)
					for _, elem := range idx.Columns {
						__antithesis_instrumentation__.Notify(564461)
						indexCols[elem.Column] = struct{}{}
					}
					__antithesis_instrumentation__.Notify(564460)
					idx.Storing = generateStoringCols(rng, info, indexCols)
					changed = true
				} else {
					__antithesis_instrumentation__.Notify(564462)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(564415)
	return stmts, changed
}

func partialIndexMutator(rng *rand.Rand, stmts []tree.Statement) ([]tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(564463)
	changed := false
	tables := getTableInfoFromDDLStatements(stmts)
	for _, stmt := range stmts {
		__antithesis_instrumentation__.Notify(564465)
		switch ast := stmt.(type) {
		case *tree.CreateIndex:
			__antithesis_instrumentation__.Notify(564466)
			info, ok := tables[ast.Table.ObjectName]
			if !ok {
				__antithesis_instrumentation__.Notify(564470)
				continue
			} else {
				__antithesis_instrumentation__.Notify(564471)
			}
			__antithesis_instrumentation__.Notify(564467)

			if ast.Predicate == nil && func() bool {
				__antithesis_instrumentation__.Notify(564472)
				return !hasReferencingConstraint(info, ast.Columns) == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(564473)
				return rng.Intn(2) == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(564474)
				tn := tree.MakeUnqualifiedTableName(ast.Table.ObjectName)
				ast.Predicate = randPartialIndexPredicateFromCols(rng, info.columnsTableDefs, &tn)
				changed = true
			} else {
				__antithesis_instrumentation__.Notify(564475)
			}
		case *tree.CreateTable:
			__antithesis_instrumentation__.Notify(564468)
			info, ok := tables[ast.Table.ObjectName]
			if !ok {
				__antithesis_instrumentation__.Notify(564476)
				panic("table info could not be found")
			} else {
				__antithesis_instrumentation__.Notify(564477)
			}
			__antithesis_instrumentation__.Notify(564469)
			for _, def := range ast.Defs {
				__antithesis_instrumentation__.Notify(564478)
				var idx *tree.IndexTableDef
				switch defType := def.(type) {
				case *tree.IndexTableDef:
					__antithesis_instrumentation__.Notify(564481)
					idx = defType
				case *tree.UniqueConstraintTableDef:
					__antithesis_instrumentation__.Notify(564482)
					if !defType.PrimaryKey && func() bool {
						__antithesis_instrumentation__.Notify(564483)
						return !defType.WithoutIndex == true
					}() == true {
						__antithesis_instrumentation__.Notify(564484)
						idx = &defType.IndexTableDef
					} else {
						__antithesis_instrumentation__.Notify(564485)
					}
				}
				__antithesis_instrumentation__.Notify(564479)

				if idx == nil {
					__antithesis_instrumentation__.Notify(564486)
					continue
				} else {
					__antithesis_instrumentation__.Notify(564487)
				}
				__antithesis_instrumentation__.Notify(564480)

				if idx.Predicate == nil && func() bool {
					__antithesis_instrumentation__.Notify(564488)
					return !hasReferencingConstraint(info, idx.Columns) == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(564489)
					return rng.Intn(2) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(564490)
					tn := tree.MakeUnqualifiedTableName(ast.Table.ObjectName)
					idx.Predicate = randPartialIndexPredicateFromCols(rng, info.columnsTableDefs, &tn)
					changed = true
				} else {
					__antithesis_instrumentation__.Notify(564491)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(564464)
	return stmts, changed
}

func hasReferencingConstraint(info tableInfo, idxColumns tree.IndexElemList) bool {
	__antithesis_instrumentation__.Notify(564492)
RefColsLoop:
	for _, refCols := range info.refColsLists {
		__antithesis_instrumentation__.Notify(564494)
		if len(refCols) != len(idxColumns) {
			__antithesis_instrumentation__.Notify(564497)
			continue RefColsLoop
		} else {
			__antithesis_instrumentation__.Notify(564498)
		}
		__antithesis_instrumentation__.Notify(564495)
		for i := range refCols {
			__antithesis_instrumentation__.Notify(564499)
			if refCols[i] != idxColumns[i].Column {
				__antithesis_instrumentation__.Notify(564500)
				continue RefColsLoop
			} else {
				__antithesis_instrumentation__.Notify(564501)
			}
		}
		__antithesis_instrumentation__.Notify(564496)
		return true
	}
	__antithesis_instrumentation__.Notify(564493)
	return false
}
