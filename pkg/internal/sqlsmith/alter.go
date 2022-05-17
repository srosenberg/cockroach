package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var (
	alters               = append(append(append(altersTableExistence, altersExistingTable...), altersTypeExistence...), altersExistingTypes...)
	altersTableExistence = []statementWeight{
		{10, makeCreateTable},
		{2, makeCreateSchema},
		{1, makeDropTable},
	}
	altersExistingTable = []statementWeight{
		{5, makeRenameTable},

		{10, makeAddColumn},
		{10, makeJSONComputedColumn},
		{10, makeAlterPrimaryKey},
		{1, makeDropColumn},
		{5, makeRenameColumn},
		{5, makeAlterColumnType},

		{10, makeCreateIndex},
		{1, makeDropIndex},
		{5, makeRenameIndex},
	}
	altersTypeExistence = []statementWeight{
		{5, makeCreateType},
	}
	altersExistingTypes = []statementWeight{
		{5, makeAlterTypeDropValue},
		{5, makeAlterTypeAddValue},
		{1, makeAlterTypeRenameValue},
		{1, makeAlterTypeRenameType},
	}
	alterTableMultiregion = []statementWeight{
		{10, makeAlterLocality},
	}
	alterDatabaseMultiregion = []statementWeight{
		{5, makeAlterDatabaseDropRegion},
		{5, makeAlterDatabaseAddRegion},
		{5, makeAlterSurvivalGoal},
		{5, makeAlterDatabasePlacement},
	}
	alterMultiregion = append(alterTableMultiregion, alterDatabaseMultiregion...)
)

func makeAlter(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68743)
	if s.canRecurse() {
		__antithesis_instrumentation__.Notify(68745)

		err := s.ReloadSchemas()
		if err != nil {
			__antithesis_instrumentation__.Notify(68747)

			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(68748)
		}
		__antithesis_instrumentation__.Notify(68746)

		for i := 0; i < retryCount; i++ {
			__antithesis_instrumentation__.Notify(68749)
			stmt, ok := s.alterSampler.Next()(s)
			if ok {
				__antithesis_instrumentation__.Notify(68750)
				return stmt, ok
			} else {
				__antithesis_instrumentation__.Notify(68751)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(68752)
	}
	__antithesis_instrumentation__.Notify(68744)
	return nil, false
}

func makeCreateSchema(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68753)
	return &tree.CreateSchema{
		Schema: tree.ObjectNamePrefix{
			SchemaName:     s.name("schema"),
			ExplicitSchema: true,
		},
	}, true
}

func makeCreateTable(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68754)
	table := randgen.RandCreateTable(s.rnd, "", 0)
	schemaOrd := s.rnd.Intn(len(s.schemas))
	schema := s.schemas[schemaOrd]
	table.Table = tree.MakeTableNameWithSchema(tree.Name(s.dbName), schema.SchemaName, s.name("tab"))
	return table, true
}

func makeDropTable(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68755)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68757)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68758)
	}
	__antithesis_instrumentation__.Notify(68756)

	return &tree.DropTable{
		Names:        tree.TableNames{*tableRef.TableName},
		DropBehavior: s.randDropBehavior(),
	}, true
}

func makeRenameTable(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68759)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68762)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68763)
	}
	__antithesis_instrumentation__.Notify(68760)

	newName, err := tree.NewUnresolvedObjectName(
		1, [3]string{string(s.name("tab"))}, tree.NoAnnotation,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(68764)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68765)
	}
	__antithesis_instrumentation__.Notify(68761)

	return &tree.RenameTable{
		Name:    tableRef.TableName.ToUnresolvedObjectName(),
		NewName: newName,
	}, true
}

func makeRenameColumn(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68766)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68768)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68769)
	}
	__antithesis_instrumentation__.Notify(68767)
	col := tableRef.Columns[s.rnd.Intn(len(tableRef.Columns))]

	return &tree.RenameColumn{
		Table:   *tableRef.TableName,
		Name:    col.Name,
		NewName: s.name("col"),
	}, true
}

func makeAlterColumnType(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68770)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68772)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68773)
	}
	__antithesis_instrumentation__.Notify(68771)
	typ := randgen.RandColumnType(s.rnd)
	col := tableRef.Columns[s.rnd.Intn(len(tableRef.Columns))]

	return &tree.AlterTable{
		Table: tableRef.TableName.ToUnresolvedObjectName(),
		Cmds: tree.AlterTableCmds{
			&tree.AlterTableAlterColumnType{
				Column: col.Name,
				ToType: typ,
			},
		},
	}, true
}

func makeAddColumn(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68774)
	_, _, tableRef, colRefs, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68779)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68780)
	}
	__antithesis_instrumentation__.Notify(68775)
	colRefs.stripTableName()
	t := randgen.RandColumnType(s.rnd)
	col, err := tree.NewColumnTableDef(s.name("col"), t, false, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(68781)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68782)
	}
	__antithesis_instrumentation__.Notify(68776)
	col.Nullable.Nullability = s.randNullability()
	if s.coin() {
		__antithesis_instrumentation__.Notify(68783)
		col.DefaultExpr.Expr = &tree.ParenExpr{Expr: makeScalar(s, t, nil)}
	} else {
		__antithesis_instrumentation__.Notify(68784)
		if s.coin() {
			__antithesis_instrumentation__.Notify(68785)
			col.Computed.Computed = true
			col.Computed.Expr = &tree.ParenExpr{Expr: makeScalar(s, t, colRefs)}
		} else {
			__antithesis_instrumentation__.Notify(68786)
		}
	}
	__antithesis_instrumentation__.Notify(68777)
	for s.coin() {
		__antithesis_instrumentation__.Notify(68787)
		col.CheckExprs = append(col.CheckExprs, tree.ColumnTableDefCheckExpr{
			Expr: makeBoolExpr(s, colRefs),
		})
	}
	__antithesis_instrumentation__.Notify(68778)

	return &tree.AlterTable{
		Table: tableRef.TableName.ToUnresolvedObjectName(),
		Cmds: tree.AlterTableCmds{
			&tree.AlterTableAddColumn{
				ColumnDef: col,
			},
		},
	}, true
}

func makeJSONComputedColumn(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68788)
	_, _, tableRef, colRefs, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68794)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68795)
	}
	__antithesis_instrumentation__.Notify(68789)
	colRefs.stripTableName()

	s.rnd.Shuffle(len(colRefs), func(i, j int) {
		__antithesis_instrumentation__.Notify(68796)
		colRefs[i], colRefs[j] = colRefs[j], colRefs[i]
	})
	__antithesis_instrumentation__.Notify(68790)
	var ref *colRef
	for _, c := range colRefs {
		__antithesis_instrumentation__.Notify(68797)
		if c.typ.Family() == types.JsonFamily {
			__antithesis_instrumentation__.Notify(68798)
			ref = c
			break
		} else {
			__antithesis_instrumentation__.Notify(68799)
		}
	}
	__antithesis_instrumentation__.Notify(68791)

	if ref == nil {
		__antithesis_instrumentation__.Notify(68800)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68801)
	}
	__antithesis_instrumentation__.Notify(68792)
	col, err := tree.NewColumnTableDef(s.name("col"), types.Jsonb, false, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(68802)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68803)
	}
	__antithesis_instrumentation__.Notify(68793)
	col.Computed.Computed = true
	col.Computed.Expr = tree.NewTypedBinaryExpr(
		treebin.MakeBinaryOperator(treebin.JSONFetchText),
		ref.typedExpr(),
		randgen.RandDatumSimple(s.rnd, types.String),
		types.String,
	)

	return &tree.AlterTable{
		Table: tableRef.TableName.ToUnresolvedObjectName(),
		Cmds: tree.AlterTableCmds{
			&tree.AlterTableAddColumn{
				ColumnDef: col,
			},
		},
	}, true
}

func makeDropColumn(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68804)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68806)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68807)
	}
	__antithesis_instrumentation__.Notify(68805)
	col := tableRef.Columns[s.rnd.Intn(len(tableRef.Columns))]

	return &tree.AlterTable{
		Table: tableRef.TableName.ToUnresolvedObjectName(),
		Cmds: tree.AlterTableCmds{
			&tree.AlterTableDropColumn{
				Column:       col.Name,
				DropBehavior: s.randDropBehavior(),
			},
		},
	}, true
}

func makeAlterPrimaryKey(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68808)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68814)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68815)
	}
	__antithesis_instrumentation__.Notify(68809)

	var candidateColumns tree.IndexElemList
	for _, c := range tableRef.Columns {
		__antithesis_instrumentation__.Notify(68816)
		if c.Nullable.Nullability == tree.NotNull {
			__antithesis_instrumentation__.Notify(68817)
			candidateColumns = append(candidateColumns, tree.IndexElem{Column: c.Name})
		} else {
			__antithesis_instrumentation__.Notify(68818)
		}
	}
	__antithesis_instrumentation__.Notify(68810)
	if len(candidateColumns) == 0 {
		__antithesis_instrumentation__.Notify(68819)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68820)
	}
	__antithesis_instrumentation__.Notify(68811)
	s.rnd.Shuffle(len(candidateColumns), func(i, j int) {
		__antithesis_instrumentation__.Notify(68821)
		candidateColumns[i], candidateColumns[j] = candidateColumns[j], candidateColumns[i]
	})
	__antithesis_instrumentation__.Notify(68812)

	i := 1
	for len(candidateColumns) > i && func() bool {
		__antithesis_instrumentation__.Notify(68822)
		return s.rnd.Intn(2) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(68823)
		i++
	}
	__antithesis_instrumentation__.Notify(68813)
	candidateColumns = candidateColumns[:i]
	return &tree.AlterTable{
		Table: tableRef.TableName.ToUnresolvedObjectName(),
		Cmds: tree.AlterTableCmds{
			&tree.AlterTableAlterPrimaryKey{
				Columns: candidateColumns,
			},
		},
	}, true
}

func makeCreateIndex(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68824)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68828)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68829)
	}
	__antithesis_instrumentation__.Notify(68825)
	var cols tree.IndexElemList
	seen := map[tree.Name]bool{}
	inverted := false
	unique := s.coin()
	for len(cols) < 1 || func() bool {
		__antithesis_instrumentation__.Notify(68830)
		return s.coin() == true
	}() == true {
		__antithesis_instrumentation__.Notify(68831)
		col := tableRef.Columns[s.rnd.Intn(len(tableRef.Columns))]
		if seen[col.Name] {
			__antithesis_instrumentation__.Notify(68834)
			continue
		} else {
			__antithesis_instrumentation__.Notify(68835)
		}
		__antithesis_instrumentation__.Notify(68832)
		seen[col.Name] = true

		if len(cols) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(68836)
			return colinfo.ColumnTypeIsInvertedIndexable(tree.MustBeStaticallyKnownType(col.Type)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(68837)
			inverted = true
			unique = false
			cols = append(cols, tree.IndexElem{
				Column: col.Name,
			})
			break
		} else {
			__antithesis_instrumentation__.Notify(68838)
		}
		__antithesis_instrumentation__.Notify(68833)
		if colinfo.ColumnTypeIsIndexable(tree.MustBeStaticallyKnownType(col.Type)) {
			__antithesis_instrumentation__.Notify(68839)
			cols = append(cols, tree.IndexElem{
				Column:    col.Name,
				Direction: s.randDirection(),
			})
		} else {
			__antithesis_instrumentation__.Notify(68840)
		}
	}
	__antithesis_instrumentation__.Notify(68826)
	var storing tree.NameList
	for !inverted && func() bool {
		__antithesis_instrumentation__.Notify(68841)
		return s.coin() == true
	}() == true {
		__antithesis_instrumentation__.Notify(68842)
		col := tableRef.Columns[s.rnd.Intn(len(tableRef.Columns))]
		if seen[col.Name] {
			__antithesis_instrumentation__.Notify(68844)
			continue
		} else {
			__antithesis_instrumentation__.Notify(68845)
		}
		__antithesis_instrumentation__.Notify(68843)
		seen[col.Name] = true
		storing = append(storing, col.Name)
	}
	__antithesis_instrumentation__.Notify(68827)

	return &tree.CreateIndex{
		Name:         s.name("idx"),
		Table:        *tableRef.TableName,
		Unique:       unique,
		Columns:      cols,
		Storing:      storing,
		Inverted:     inverted,
		Concurrently: s.coin(),
	}, true
}

func makeDropIndex(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68846)
	tin, _, _, ok := s.getRandIndex()
	return &tree.DropIndex{
		IndexList:    tree.TableIndexNames{tin},
		DropBehavior: s.randDropBehavior(),
		Concurrently: s.coin(),
	}, ok
}

func makeRenameIndex(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68847)
	tin, _, _, ok := s.getRandIndex()
	return &tree.RenameIndex{
		Index:   tin,
		NewName: tree.UnrestrictedName(s.name("idx")),
	}, ok
}

func makeCreateType(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68848)
	name := s.name("typ")
	return randgen.RandCreateType(s.rnd, string(name), letters), true
}

func rowsToRegionList(rows *gosql.Rows) []string {
	__antithesis_instrumentation__.Notify(68849)

	regionsSet := make(map[string]struct{})
	var region, zone string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(68852)
		if err := rows.Scan(&region, &zone); err != nil {
			__antithesis_instrumentation__.Notify(68854)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(68855)
		}
		__antithesis_instrumentation__.Notify(68853)
		regionsSet[region] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(68850)

	var regions []string
	for region := range regionsSet {
		__antithesis_instrumentation__.Notify(68856)
		regions = append(regions, region)
	}
	__antithesis_instrumentation__.Notify(68851)
	return regions
}

func getClusterRegions(s *Smither) []string {
	__antithesis_instrumentation__.Notify(68857)
	rows, err := s.db.Query("SHOW REGIONS FROM CLUSTER")
	if err != nil {
		__antithesis_instrumentation__.Notify(68859)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(68860)
	}
	__antithesis_instrumentation__.Notify(68858)
	return rowsToRegionList(rows)
}

func getDatabaseRegions(s *Smither) []string {
	__antithesis_instrumentation__.Notify(68861)
	rows, err := s.db.Query("SHOW REGIONS FROM DATABASE defaultdb")
	if err != nil {
		__antithesis_instrumentation__.Notify(68863)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(68864)
	}
	__antithesis_instrumentation__.Notify(68862)
	return rowsToRegionList(rows)
}

func makeAlterLocality(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68865)
	_, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(68868)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68869)
	}
	__antithesis_instrumentation__.Notify(68866)
	regions := getClusterRegions(s)

	localityLevel := tree.LocalityLevel(rand.Intn(3))
	ast := &tree.AlterTableLocality{
		Name: tableRef.TableName.ToUnresolvedObjectName(),
		Locality: &tree.Locality{
			LocalityLevel: localityLevel,
		},
	}
	if localityLevel == tree.LocalityLevelTable {
		__antithesis_instrumentation__.Notify(68870)
		if len(regions) == 0 {
			__antithesis_instrumentation__.Notify(68872)
			return &tree.AlterDatabaseAddRegion{}, false
		} else {
			__antithesis_instrumentation__.Notify(68873)
		}
		__antithesis_instrumentation__.Notify(68871)
		ast.Locality.TableRegion = tree.Name(regions[rand.Intn(len(regions))])
	} else {
		__antithesis_instrumentation__.Notify(68874)
	}
	__antithesis_instrumentation__.Notify(68867)
	return ast, ok
}

func makeAlterDatabaseAddRegion(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68875)
	regions := getClusterRegions(s)

	if len(regions) == 0 {
		__antithesis_instrumentation__.Notify(68877)
		return &tree.AlterDatabaseAddRegion{}, false
	} else {
		__antithesis_instrumentation__.Notify(68878)
	}
	__antithesis_instrumentation__.Notify(68876)

	ast := &tree.AlterDatabaseAddRegion{
		Region: tree.Name(regions[rand.Intn(len(regions))]),
		Name:   tree.Name("defaultdb"),
	}

	return ast, true
}

func makeAlterDatabaseDropRegion(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68879)
	regions := getDatabaseRegions(s)

	if len(regions) == 0 {
		__antithesis_instrumentation__.Notify(68881)
		return &tree.AlterDatabaseDropRegion{}, false
	} else {
		__antithesis_instrumentation__.Notify(68882)
	}
	__antithesis_instrumentation__.Notify(68880)

	ast := &tree.AlterDatabaseDropRegion{
		Region: tree.Name(regions[rand.Intn(len(regions))]),
		Name:   tree.Name("defaultdb"),
	}

	return ast, true
}

func makeAlterSurvivalGoal(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68883)

	survivalGoals := [...]tree.SurvivalGoal{
		tree.SurvivalGoalRegionFailure,
		tree.SurvivalGoalZoneFailure,
	}
	survivalGoal := survivalGoals[rand.Intn(len(survivalGoals))]

	ast := &tree.AlterDatabaseSurvivalGoal{
		Name:         tree.Name("defaultdb"),
		SurvivalGoal: survivalGoal,
	}
	return ast, true
}

func makeAlterDatabasePlacement(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68884)

	dataPlacements := [...]tree.DataPlacement{
		tree.DataPlacementDefault,
		tree.DataPlacementRestricted,
	}
	dataPlacement := dataPlacements[rand.Intn(len(dataPlacements))]

	ast := &tree.AlterDatabasePlacement{
		Name:      tree.Name("defaultdb"),
		Placement: dataPlacement,
	}

	return ast, true
}

func makeAlterTypeDropValue(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68885)
	enumVal, udtName, ok := s.getRandUserDefinedTypeLabel()
	if !ok {
		__antithesis_instrumentation__.Notify(68887)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68888)
	}
	__antithesis_instrumentation__.Notify(68886)
	return &tree.AlterType{
		Type: udtName.ToUnresolvedObjectName(),
		Cmd: &tree.AlterTypeDropValue{
			Val: *enumVal,
		},
	}, ok
}

func makeAlterTypeAddValue(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68889)
	_, udtName, ok := s.getRandUserDefinedTypeLabel()
	if !ok {
		__antithesis_instrumentation__.Notify(68891)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68892)
	}
	__antithesis_instrumentation__.Notify(68890)
	return &tree.AlterType{
		Type: udtName.ToUnresolvedObjectName(),
		Cmd: &tree.AlterTypeAddValue{
			NewVal:      tree.EnumValue(s.name("added_val")),
			IfNotExists: true,
		},
	}, true
}

func makeAlterTypeRenameValue(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68893)
	enumVal, udtName, ok := s.getRandUserDefinedTypeLabel()
	if !ok {
		__antithesis_instrumentation__.Notify(68895)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68896)
	}
	__antithesis_instrumentation__.Notify(68894)
	return &tree.AlterType{
		Type: udtName.ToUnresolvedObjectName(),
		Cmd: &tree.AlterTypeRenameValue{
			OldVal: *enumVal,
			NewVal: tree.EnumValue(s.name("renamed_val")),
		},
	}, true
}

func makeAlterTypeRenameType(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(68897)
	_, udtName, ok := s.getRandUserDefinedTypeLabel()
	if !ok {
		__antithesis_instrumentation__.Notify(68899)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(68900)
	}
	__antithesis_instrumentation__.Notify(68898)
	return &tree.AlterType{
		Type: udtName.ToUnresolvedObjectName(),
		Cmd: &tree.AlterTypeRename{
			NewName: s.name("typ"),
		},
	}, true
}
