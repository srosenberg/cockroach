package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
)

func MakeSchemaName(ifNotExists bool, schema string, authRole tree.RoleSpec) *tree.CreateSchema {
	__antithesis_instrumentation__.Notify(564502)
	return &tree.CreateSchema{
		IfNotExists: ifNotExists,
		Schema: tree.ObjectNamePrefix{
			SchemaName:     tree.Name(schema),
			ExplicitSchema: true,
		},
		AuthRole: authRole,
	}
}

func RandCreateType(rng *rand.Rand, name, alphabet string) tree.Statement {
	__antithesis_instrumentation__.Notify(564503)
	numLabels := rng.Intn(6) + 1
	labels := make(tree.EnumValueList, numLabels)
	labelsMap := make(map[string]struct{})
	i := 0
	for i < numLabels {
		__antithesis_instrumentation__.Notify(564506)
		s := RandString(rng, rng.Intn(6)+1, alphabet)
		if _, ok := labelsMap[s]; !ok {
			__antithesis_instrumentation__.Notify(564507)
			labels[i] = tree.EnumValue(s)
			labelsMap[s] = struct{}{}
			i++
		} else {
			__antithesis_instrumentation__.Notify(564508)
		}
	}
	__antithesis_instrumentation__.Notify(564504)
	un, err := tree.NewUnresolvedObjectName(1, [3]string{name}, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(564509)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(564510)
	}
	__antithesis_instrumentation__.Notify(564505)
	return &tree.CreateType{
		TypeName:   un,
		Variety:    tree.Enum,
		EnumLabels: labels,
	}
}

func RandCreateTables(
	rng *rand.Rand, prefix string, num int, mutators ...Mutator,
) []tree.Statement {
	__antithesis_instrumentation__.Notify(564511)
	if num < 1 {
		__antithesis_instrumentation__.Notify(564515)
		panic("at least one table required")
	} else {
		__antithesis_instrumentation__.Notify(564516)
	}
	__antithesis_instrumentation__.Notify(564512)

	tables := make([]tree.Statement, num)
	for i := 0; i < num; i++ {
		__antithesis_instrumentation__.Notify(564517)
		t := RandCreateTable(rng, prefix, i+1)
		tables[i] = t
	}
	__antithesis_instrumentation__.Notify(564513)

	for _, m := range mutators {
		__antithesis_instrumentation__.Notify(564518)
		tables, _ = m.Mutate(rng, tables)
	}
	__antithesis_instrumentation__.Notify(564514)

	return tables
}

func RandCreateTable(rng *rand.Rand, prefix string, tableIdx int) *tree.CreateTable {
	__antithesis_instrumentation__.Notify(564519)
	return RandCreateTableWithColumnIndexNumberGenerator(rng, prefix, tableIdx, nil)
}

func RandCreateTableWithColumnIndexNumberGenerator(
	rng *rand.Rand, prefix string, tableIdx int, generateColumnIndexNumber func() int64,
) *tree.CreateTable {
	__antithesis_instrumentation__.Notify(564520)

	nColumns := randutil.RandIntInRange(rng, 1, 20)
	columnDefs := make([]*tree.ColumnTableDef, 0, nColumns)

	defs := make(tree.TableDefs, 0, len(columnDefs))

	colIdx := func(ordinal int) int {
		__antithesis_instrumentation__.Notify(564527)
		if generateColumnIndexNumber != nil {
			__antithesis_instrumentation__.Notify(564529)
			return int(generateColumnIndexNumber())
		} else {
			__antithesis_instrumentation__.Notify(564530)
		}
		__antithesis_instrumentation__.Notify(564528)
		return ordinal
	}
	__antithesis_instrumentation__.Notify(564521)

	nComputedColumns := randutil.RandIntInRange(rng, 0, (nColumns+1)/2)
	nNormalColumns := nColumns - nComputedColumns
	for i := 0; i < nNormalColumns; i++ {
		__antithesis_instrumentation__.Notify(564531)
		columnDef := randColumnTableDef(rng, tableIdx, colIdx(i))
		columnDefs = append(columnDefs, columnDef)
		defs = append(defs, columnDef)
	}
	__antithesis_instrumentation__.Notify(564522)

	normalColDefs := columnDefs
	for i := nNormalColumns; i < nColumns; i++ {
		__antithesis_instrumentation__.Notify(564532)
		columnDef := randComputedColumnTableDef(rng, normalColDefs, tableIdx, colIdx(i))
		columnDefs = append(columnDefs, columnDef)
		defs = append(defs, columnDef)
	}
	__antithesis_instrumentation__.Notify(564523)

	if rng.Intn(8) != 0 {
		__antithesis_instrumentation__.Notify(564533)
		indexDef, ok := randIndexTableDefFromCols(rng, columnDefs, false)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(564535)
			return !indexDef.Inverted == true
		}() == true {
			__antithesis_instrumentation__.Notify(564536)
			defs = append(defs, &tree.UniqueConstraintTableDef{
				PrimaryKey:    true,
				IndexTableDef: indexDef,
			})
		} else {
			__antithesis_instrumentation__.Notify(564537)
		}
		__antithesis_instrumentation__.Notify(564534)

		for _, col := range columnDefs {
			__antithesis_instrumentation__.Notify(564538)
			for _, elem := range indexDef.Columns {
				__antithesis_instrumentation__.Notify(564539)
				if col.Name == elem.Column {
					__antithesis_instrumentation__.Notify(564540)
					col.Nullable.Nullability = tree.NotNull
				} else {
					__antithesis_instrumentation__.Notify(564541)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(564542)
	}
	__antithesis_instrumentation__.Notify(564524)

	nIdxs := rng.Intn(10)
	for i := 0; i < nIdxs; i++ {
		__antithesis_instrumentation__.Notify(564543)
		indexDef, ok := randIndexTableDefFromCols(rng, columnDefs, true)
		if !ok {
			__antithesis_instrumentation__.Notify(564545)
			continue
		} else {
			__antithesis_instrumentation__.Notify(564546)
		}
		__antithesis_instrumentation__.Notify(564544)

		unique := !indexDef.Inverted && func() bool {
			__antithesis_instrumentation__.Notify(564547)
			return rng.Intn(2) == 0 == true
		}() == true
		if unique {
			__antithesis_instrumentation__.Notify(564548)
			defs = append(defs, &tree.UniqueConstraintTableDef{
				IndexTableDef: indexDef,
			})
		} else {
			__antithesis_instrumentation__.Notify(564549)
			defs = append(defs, &indexDef)
		}
	}
	__antithesis_instrumentation__.Notify(564525)

	ret := &tree.CreateTable{
		Table: tree.MakeUnqualifiedTableName(tree.Name(fmt.Sprintf("%s%d", prefix, tableIdx))),
		Defs:  defs,
	}

	if rng.Intn(2) == 0 {
		__antithesis_instrumentation__.Notify(564550)
		ColumnFamilyMutator(rng, ret)
	} else {
		__antithesis_instrumentation__.Notify(564551)
	}
	__antithesis_instrumentation__.Notify(564526)

	res, _ := IndexStoringMutator(rng, []tree.Statement{ret})
	return res[0].(*tree.CreateTable)
}

func parseCreateStatement(createStmtSQL string) (*tree.CreateTable, error) {
	__antithesis_instrumentation__.Notify(564552)
	var p parser.Parser
	stmts, err := p.Parse(createStmtSQL)
	if err != nil {
		__antithesis_instrumentation__.Notify(564556)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564557)
	}
	__antithesis_instrumentation__.Notify(564553)
	if len(stmts) != 1 {
		__antithesis_instrumentation__.Notify(564558)
		return nil, errors.Errorf("parsed CreateStatement string yielded more than one parsed statment")
	} else {
		__antithesis_instrumentation__.Notify(564559)
	}
	__antithesis_instrumentation__.Notify(564554)
	tableStmt, ok := stmts[0].AST.(*tree.CreateTable)
	if !ok {
		__antithesis_instrumentation__.Notify(564560)
		return nil, errors.Errorf("AST could not be cast to *tree.CreateTable")
	} else {
		__antithesis_instrumentation__.Notify(564561)
	}
	__antithesis_instrumentation__.Notify(564555)
	return tableStmt, nil
}

func generateInsertStmtVals(rng *rand.Rand, colTypes []*types.T, nullable []bool) string {
	__antithesis_instrumentation__.Notify(564562)
	var valBuilder strings.Builder
	valBuilder.WriteString("(")
	comma := ""
	for j := 0; j < len(colTypes); j++ {
		__antithesis_instrumentation__.Notify(564564)
		valBuilder.WriteString(comma)
		var d tree.Datum
		if rand.Intn(10) < 4 {
			__antithesis_instrumentation__.Notify(564568)

			d = randInterestingDatum(rng, colTypes[j])
		} else {
			__antithesis_instrumentation__.Notify(564569)
		}
		__antithesis_instrumentation__.Notify(564565)
		if colTypes[j] == types.RegType {
			__antithesis_instrumentation__.Notify(564570)

			d = tree.NewDOid(tree.DInt(rand.Intn(len(types.OidToType))))
		} else {
			__antithesis_instrumentation__.Notify(564571)
		}
		__antithesis_instrumentation__.Notify(564566)
		if d == nil {
			__antithesis_instrumentation__.Notify(564572)
			d = RandDatum(rng, colTypes[j], nullable[j])
		} else {
			__antithesis_instrumentation__.Notify(564573)
		}
		__antithesis_instrumentation__.Notify(564567)
		valBuilder.WriteString(tree.AsStringWithFlags(d, tree.FmtParsable))
		comma = ", "
	}
	__antithesis_instrumentation__.Notify(564563)
	valBuilder.WriteString(")")
	return valBuilder.String()
}

func PopulateTableWithRandData(
	rng *rand.Rand, db *gosql.DB, tableName string, numInserts int,
) (numRowsInserted int, err error) {
	__antithesis_instrumentation__.Notify(564574)
	var createStmtSQL string
	res := db.QueryRow(fmt.Sprintf("SELECT create_statement FROM [SHOW CREATE TABLE %s]", tableName))
	err = res.Scan(&createStmtSQL)
	if err != nil {
		__antithesis_instrumentation__.Notify(564580)
		return 0, errors.Wrapf(err, "table does not exist in db")
	} else {
		__antithesis_instrumentation__.Notify(564581)
	}
	__antithesis_instrumentation__.Notify(564575)
	createStmt, err := parseCreateStatement(createStmtSQL)
	if err != nil {
		__antithesis_instrumentation__.Notify(564582)
		return 0, errors.Wrapf(err, "failed to determine table schema")
	} else {
		__antithesis_instrumentation__.Notify(564583)
	}
	__antithesis_instrumentation__.Notify(564576)

	var hasFK = map[string]bool{}
	for _, def := range createStmt.Defs {
		__antithesis_instrumentation__.Notify(564584)
		if fk, ok := def.(*tree.ForeignKeyConstraintTableDef); ok {
			__antithesis_instrumentation__.Notify(564585)
			for _, col := range fk.FromCols {
				__antithesis_instrumentation__.Notify(564586)
				hasFK[col.String()] = true
			}
		} else {
			__antithesis_instrumentation__.Notify(564587)
		}
	}
	__antithesis_instrumentation__.Notify(564577)

	colTypes := make([]*types.T, 0)
	nullable := make([]bool, 0)
	var colNameBuilder strings.Builder
	comma := ""
	for _, def := range createStmt.Defs {
		__antithesis_instrumentation__.Notify(564588)
		if col, ok := def.(*tree.ColumnTableDef); ok {
			__antithesis_instrumentation__.Notify(564589)
			if _, ok := hasFK[col.Name.String()]; ok {
				__antithesis_instrumentation__.Notify(564592)

				if col.Nullable.Nullability == tree.Null {
					__antithesis_instrumentation__.Notify(564593)
					continue
				} else {
					__antithesis_instrumentation__.Notify(564594)
				}
			} else {
				__antithesis_instrumentation__.Notify(564595)
			}
			__antithesis_instrumentation__.Notify(564590)
			if col.Computed.Computed || func() bool {
				__antithesis_instrumentation__.Notify(564596)
				return col.Hidden == true
			}() == true {
				__antithesis_instrumentation__.Notify(564597)

				continue
			} else {
				__antithesis_instrumentation__.Notify(564598)
			}
			__antithesis_instrumentation__.Notify(564591)
			colTypes = append(colTypes, tree.MustBeStaticallyKnownType(col.Type.(*types.T)))
			nullable = append(nullable, col.Nullable.Nullability == tree.Null)

			colNameBuilder.WriteString(comma)
			colNameBuilder.WriteString(col.Name.String())
			comma = ", "
		} else {
			__antithesis_instrumentation__.Notify(564599)
		}
	}
	__antithesis_instrumentation__.Notify(564578)

	for i := 0; i < numInserts; i++ {
		__antithesis_instrumentation__.Notify(564600)
		insertVals := generateInsertStmtVals(rng, colTypes, nullable)
		insertStmt := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;",
			tableName,
			colNameBuilder.String(),
			insertVals)
		_, err := db.Exec(insertStmt)
		if err == nil {
			__antithesis_instrumentation__.Notify(564601)
			numRowsInserted++
		} else {
			__antithesis_instrumentation__.Notify(564602)
		}
	}
	__antithesis_instrumentation__.Notify(564579)
	return numRowsInserted, nil
}

func GenerateRandInterestingTable(db *gosql.DB, dbName, tableName string) error {
	__antithesis_instrumentation__.Notify(564603)
	var (
		randTypes []*types.T
		colNames  []string
	)
	numRows := 0
	for _, v := range randInterestingDatums {
		__antithesis_instrumentation__.Notify(564608)
		colTyp := v[0].ResolvedType()
		randTypes = append(randTypes, colTyp)
		colNames = append(colNames, colTyp.Name())
		if len(v) > numRows {
			__antithesis_instrumentation__.Notify(564609)
			numRows = len(v)
		} else {
			__antithesis_instrumentation__.Notify(564610)
		}
	}
	__antithesis_instrumentation__.Notify(564604)

	var columns strings.Builder
	comma := ""
	for i, typ := range randTypes {
		__antithesis_instrumentation__.Notify(564611)
		columns.WriteString(comma)
		columns.WriteString(colNames[i])
		columns.WriteString(" ")
		columns.WriteString(typ.SQLString())
		comma = ", "
	}
	__antithesis_instrumentation__.Notify(564605)

	createStatement := fmt.Sprintf("CREATE TABLE %s.%s (%s)", dbName, tableName, columns.String())
	if _, err := db.Exec(createStatement); err != nil {
		__antithesis_instrumentation__.Notify(564612)
		return err
	} else {
		__antithesis_instrumentation__.Notify(564613)
	}
	__antithesis_instrumentation__.Notify(564606)

	row := make([]string, len(randTypes))
	for i := 0; i < numRows; i++ {
		__antithesis_instrumentation__.Notify(564614)
		for j, typ := range randTypes {
			__antithesis_instrumentation__.Notify(564617)
			datums := randInterestingDatums[typ.Family()]
			var d tree.Datum
			if i < len(datums) {
				__antithesis_instrumentation__.Notify(564619)
				d = datums[i]
			} else {
				__antithesis_instrumentation__.Notify(564620)
				d = tree.DNull
			}
			__antithesis_instrumentation__.Notify(564618)
			row[j] = tree.AsStringWithFlags(d, tree.FmtParsable)
		}
		__antithesis_instrumentation__.Notify(564615)
		var builder strings.Builder
		comma := ""
		for _, d := range row {
			__antithesis_instrumentation__.Notify(564621)
			builder.WriteString(comma)
			builder.WriteString(d)
			comma = ", "
		}
		__antithesis_instrumentation__.Notify(564616)
		insertStmt := fmt.Sprintf("INSERT INTO %s.%s VALUES (%s)", dbName, tableName, builder.String())
		if _, err := db.Exec(insertStmt); err != nil {
			__antithesis_instrumentation__.Notify(564622)
			return err
		} else {
			__antithesis_instrumentation__.Notify(564623)
		}
	}
	__antithesis_instrumentation__.Notify(564607)
	return nil
}

func randColumnTableDef(rand *rand.Rand, tableIdx int, colIdx int) *tree.ColumnTableDef {
	__antithesis_instrumentation__.Notify(564624)
	columnDef := &tree.ColumnTableDef{

		Name: tree.Name(fmt.Sprintf("col%d_%d", tableIdx, colIdx)),
		Type: RandColumnType(rand),
	}
	columnDef.Nullable.Nullability = tree.Nullability(rand.Intn(int(tree.SilentNull) + 1))
	return columnDef
}

func randComputedColumnTableDef(
	rng *rand.Rand, normalColDefs []*tree.ColumnTableDef, tableIdx int, colIdx int,
) *tree.ColumnTableDef {
	__antithesis_instrumentation__.Notify(564625)
	newDef := randColumnTableDef(rng, tableIdx, colIdx)
	newDef.Computed.Computed = true
	newDef.Computed.Virtual = (rng.Intn(2) == 0)

	expr, typ, nullability, _ := randExpr(rng, normalColDefs, true)
	newDef.Computed.Expr = expr
	newDef.Type = typ
	newDef.Nullable.Nullability = nullability

	return newDef
}

func randIndexTableDefFromCols(
	rng *rand.Rand, columnTableDefs []*tree.ColumnTableDef, allowExpressions bool,
) (def tree.IndexTableDef, ok bool) {
	__antithesis_instrumentation__.Notify(564626)
	cpy := make([]*tree.ColumnTableDef, len(columnTableDefs))
	copy(cpy, columnTableDefs)
	rng.Shuffle(len(cpy), func(i, j int) { __antithesis_instrumentation__.Notify(564630); cpy[i], cpy[j] = cpy[j], cpy[i] })
	__antithesis_instrumentation__.Notify(564627)
	nCols := rng.Intn(len(cpy)) + 1

	cols := cpy[:nCols]

	eligibleExprIndexRefs := nonComputedColumnTableDefs(columnTableDefs)
	removeColsFromExprIndexRefCols := func(cols map[tree.Name]struct{}) {
		__antithesis_instrumentation__.Notify(564631)
		i := 0
		for j := range eligibleExprIndexRefs {
			__antithesis_instrumentation__.Notify(564633)
			eligibleExprIndexRefs[i] = eligibleExprIndexRefs[j]
			name := eligibleExprIndexRefs[j].Name
			if _, ok := cols[name]; !ok {
				__antithesis_instrumentation__.Notify(564634)
				i++
			} else {
				__antithesis_instrumentation__.Notify(564635)
			}
		}
		__antithesis_instrumentation__.Notify(564632)
		eligibleExprIndexRefs = eligibleExprIndexRefs[:i]
	}
	__antithesis_instrumentation__.Notify(564628)

	def.Columns = make(tree.IndexElemList, 0, len(cols))
	for i := range cols {
		__antithesis_instrumentation__.Notify(564636)
		semType := tree.MustBeStaticallyKnownType(cols[i].Type)
		elem := tree.IndexElem{
			Column:    cols[i].Name,
			Direction: tree.Direction(rng.Intn(int(tree.Descending) + 1)),
		}

		if allowExpressions && func() bool {
			__antithesis_instrumentation__.Notify(564640)
			return len(eligibleExprIndexRefs) > 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(564641)
			return rng.Intn(10) == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(564642)
			var expr tree.Expr

			var referencedCols map[tree.Name]struct{}
			expr, semType, _, referencedCols = randExpr(rng, eligibleExprIndexRefs, false)
			removeColsFromExprIndexRefCols(referencedCols)
			elem.Expr = expr
			elem.Column = ""
		} else {
			__antithesis_instrumentation__.Notify(564643)
		}
		__antithesis_instrumentation__.Notify(564637)

		if isLastCol := i == len(cols)-1; !isLastCol && func() bool {
			__antithesis_instrumentation__.Notify(564644)
			return !colinfo.ColumnTypeIsIndexable(semType) == true
		}() == true {
			__antithesis_instrumentation__.Notify(564645)
			return tree.IndexTableDef{}, false
		} else {
			__antithesis_instrumentation__.Notify(564646)
		}
		__antithesis_instrumentation__.Notify(564638)

		if colinfo.ColumnTypeIsInvertedIndexable(semType) {
			__antithesis_instrumentation__.Notify(564647)
			def.Inverted = true
		} else {
			__antithesis_instrumentation__.Notify(564648)
		}
		__antithesis_instrumentation__.Notify(564639)

		def.Columns = append(def.Columns, elem)
	}
	__antithesis_instrumentation__.Notify(564629)

	return def, true
}

func nonComputedColumnTableDefs(cols []*tree.ColumnTableDef) []*tree.ColumnTableDef {
	__antithesis_instrumentation__.Notify(564649)
	nonComputedCols := make([]*tree.ColumnTableDef, 0, len(cols))
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(564651)
		if !col.Computed.Computed {
			__antithesis_instrumentation__.Notify(564652)
			nonComputedCols = append(nonComputedCols, col)
		} else {
			__antithesis_instrumentation__.Notify(564653)
		}
	}
	__antithesis_instrumentation__.Notify(564650)
	return nonComputedCols
}

func TestingMakePrimaryIndexKey(
	desc catalog.TableDescriptor, vals ...interface{},
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(564654)
	return TestingMakePrimaryIndexKeyForTenant(desc, keys.SystemSQLCodec, vals...)
}

func TestingMakePrimaryIndexKeyForTenant(
	desc catalog.TableDescriptor, codec keys.SQLCodec, vals ...interface{},
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(564655)
	index := desc.GetPrimaryIndex()
	if len(vals) > index.NumKeyColumns() {
		__antithesis_instrumentation__.Notify(564660)
		return nil, errors.Errorf("got %d values, PK has %d columns", len(vals), index.NumKeyColumns())
	} else {
		__antithesis_instrumentation__.Notify(564661)
	}
	__antithesis_instrumentation__.Notify(564656)
	datums := make([]tree.Datum, len(vals))
	for i, v := range vals {
		__antithesis_instrumentation__.Notify(564662)
		switch v := v.(type) {
		case bool:
			__antithesis_instrumentation__.Notify(564664)
			datums[i] = tree.MakeDBool(tree.DBool(v))
		case int:
			__antithesis_instrumentation__.Notify(564665)
			datums[i] = tree.NewDInt(tree.DInt(v))
		case string:
			__antithesis_instrumentation__.Notify(564666)
			datums[i] = tree.NewDString(v)
		case tree.Datum:
			__antithesis_instrumentation__.Notify(564667)
			datums[i] = v
		default:
			__antithesis_instrumentation__.Notify(564668)
			return nil, errors.Errorf("unexpected value type %T", v)
		}
		__antithesis_instrumentation__.Notify(564663)

		colID := index.GetKeyColumnID(i)
		col, _ := desc.FindColumnWithID(colID)
		if col != nil && func() bool {
			__antithesis_instrumentation__.Notify(564669)
			return col.Public() == true
		}() == true {
			__antithesis_instrumentation__.Notify(564670)
			colTyp := datums[i].ResolvedType()
			if t := colTyp.Family(); t != col.GetType().Family() {
				__antithesis_instrumentation__.Notify(564671)
				return nil, errors.Errorf("column %d of type %s, got value of type %s", i, col.GetType().Family(), t)
			} else {
				__antithesis_instrumentation__.Notify(564672)
			}
		} else {
			__antithesis_instrumentation__.Notify(564673)
		}
	}
	__antithesis_instrumentation__.Notify(564657)

	var colIDToRowIndex catalog.TableColMap
	for i := range vals {
		__antithesis_instrumentation__.Notify(564674)
		colIDToRowIndex.Set(index.GetKeyColumnID(i), i)
	}
	__antithesis_instrumentation__.Notify(564658)

	keyPrefix := rowenc.MakeIndexKeyPrefix(codec, desc.GetID(), index.GetID())
	key, _, err := rowenc.EncodeIndexKey(desc, index, colIDToRowIndex, datums, keyPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(564675)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564676)
	}
	__antithesis_instrumentation__.Notify(564659)
	return key, nil
}

func TestingMakeSecondaryIndexKey(
	desc catalog.TableDescriptor, index catalog.Index, codec keys.SQLCodec, vals ...interface{},
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(564677)
	if len(vals) > index.NumKeyColumns() {
		__antithesis_instrumentation__.Notify(564682)
		return nil, errors.Errorf("got %d values, index %s has %d columns", len(vals), index.GetName(), index.NumKeyColumns())
	} else {
		__antithesis_instrumentation__.Notify(564683)
	}
	__antithesis_instrumentation__.Notify(564678)

	datums := make([]tree.Datum, len(vals))
	for i, v := range vals {
		__antithesis_instrumentation__.Notify(564684)
		switch v := v.(type) {
		case bool:
			__antithesis_instrumentation__.Notify(564686)
			datums[i] = tree.MakeDBool(tree.DBool(v))
		case int:
			__antithesis_instrumentation__.Notify(564687)
			datums[i] = tree.NewDInt(tree.DInt(v))
		case string:
			__antithesis_instrumentation__.Notify(564688)
			datums[i] = tree.NewDString(v)
		case tree.Datum:
			__antithesis_instrumentation__.Notify(564689)
			datums[i] = v
		default:
			__antithesis_instrumentation__.Notify(564690)
			return nil, errors.Errorf("unexpected value type %T", v)
		}
		__antithesis_instrumentation__.Notify(564685)

		colID := index.GetKeyColumnID(i)
		col, _ := desc.FindColumnWithID(colID)
		if col != nil && func() bool {
			__antithesis_instrumentation__.Notify(564691)
			return col.Public() == true
		}() == true {
			__antithesis_instrumentation__.Notify(564692)
			colTyp := datums[i].ResolvedType()
			if t := colTyp.Family(); t != col.GetType().Family() {
				__antithesis_instrumentation__.Notify(564693)
				return nil, errors.Errorf("column %d of type %s, got value of type %s", i, col.GetType().Family(), t)
			} else {
				__antithesis_instrumentation__.Notify(564694)
			}
		} else {
			__antithesis_instrumentation__.Notify(564695)
		}
	}
	__antithesis_instrumentation__.Notify(564679)

	var colIDToRowIndex catalog.TableColMap
	for i := range vals {
		__antithesis_instrumentation__.Notify(564696)
		colIDToRowIndex.Set(index.GetKeyColumnID(i), i)
	}
	__antithesis_instrumentation__.Notify(564680)

	keyPrefix := rowenc.MakeIndexKeyPrefix(codec, desc.GetID(), index.GetID())
	key, _, err := rowenc.EncodeIndexKey(desc, index, colIDToRowIndex, datums, keyPrefix)

	if err != nil {
		__antithesis_instrumentation__.Notify(564697)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(564698)
	}
	__antithesis_instrumentation__.Notify(564681)
	return key, nil
}
