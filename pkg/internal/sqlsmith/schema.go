package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"

	_ "github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treebin"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type tableRef struct {
	TableName *tree.TableName
	Columns   []*tree.ColumnTableDef
}

type aliasedTableRef struct {
	*tableRef
	indexFlags *tree.IndexFlags
}

type tableRefs []*tableRef

func (s *Smither) ReloadSchemas() error {
	__antithesis_instrumentation__.Notify(69656)
	if s.db == nil {
		__antithesis_instrumentation__.Notify(69662)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(69663)
	}
	__antithesis_instrumentation__.Notify(69657)
	s.lock.Lock()
	defer s.lock.Unlock()
	var err error
	s.types, err = s.extractTypes()
	if err != nil {
		__antithesis_instrumentation__.Notify(69664)
		return err
	} else {
		__antithesis_instrumentation__.Notify(69665)
	}
	__antithesis_instrumentation__.Notify(69658)
	s.tables, err = s.extractTables()
	if err != nil {
		__antithesis_instrumentation__.Notify(69666)
		return err
	} else {
		__antithesis_instrumentation__.Notify(69667)
	}
	__antithesis_instrumentation__.Notify(69659)
	s.schemas, err = s.extractSchemas()
	if err != nil {
		__antithesis_instrumentation__.Notify(69668)
		return err
	} else {
		__antithesis_instrumentation__.Notify(69669)
	}
	__antithesis_instrumentation__.Notify(69660)
	s.indexes, err = s.extractIndexes(s.tables)
	s.columns = make(map[tree.TableName]map[tree.Name]*tree.ColumnTableDef)
	for _, ref := range s.tables {
		__antithesis_instrumentation__.Notify(69670)
		s.columns[*ref.TableName] = make(map[tree.Name]*tree.ColumnTableDef)
		for _, col := range ref.Columns {
			__antithesis_instrumentation__.Notify(69671)
			s.columns[*ref.TableName][col.Name] = col
		}
	}
	__antithesis_instrumentation__.Notify(69661)
	return err
}

func (s *Smither) getRandTable() (*aliasedTableRef, bool) {
	__antithesis_instrumentation__.Notify(69672)
	s.lock.RLock()
	defer s.lock.RUnlock()
	if len(s.tables) == 0 {
		__antithesis_instrumentation__.Notify(69675)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69676)
	}
	__antithesis_instrumentation__.Notify(69673)
	table := s.tables[s.rnd.Intn(len(s.tables))]
	indexes := s.indexes[*table.TableName]
	var indexFlags tree.IndexFlags
	if s.coin() {
		__antithesis_instrumentation__.Notify(69677)
		indexNames := make([]tree.Name, 0, len(indexes))
		for _, index := range indexes {
			__antithesis_instrumentation__.Notify(69679)
			if !index.Inverted {
				__antithesis_instrumentation__.Notify(69680)
				indexNames = append(indexNames, index.Name)
			} else {
				__antithesis_instrumentation__.Notify(69681)
			}
		}
		__antithesis_instrumentation__.Notify(69678)
		if len(indexNames) > 0 {
			__antithesis_instrumentation__.Notify(69682)
			indexFlags.Index = tree.UnrestrictedName(indexNames[s.rnd.Intn(len(indexNames))])
		} else {
			__antithesis_instrumentation__.Notify(69683)
		}
	} else {
		__antithesis_instrumentation__.Notify(69684)
	}
	__antithesis_instrumentation__.Notify(69674)
	aliased := &aliasedTableRef{
		tableRef:   table,
		indexFlags: &indexFlags,
	}
	return aliased, true
}

func (s *Smither) getRandTableIndex(
	table, alias tree.TableName,
) (*tree.TableIndexName, *tree.CreateIndex, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69685)
	s.lock.RLock()
	indexes := s.indexes[table]
	s.lock.RUnlock()
	if len(indexes) == 0 {
		__antithesis_instrumentation__.Notify(69689)
		return nil, nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69690)
	}
	__antithesis_instrumentation__.Notify(69686)
	names := make([]tree.Name, 0, len(indexes))
	for n := range indexes {
		__antithesis_instrumentation__.Notify(69691)
		names = append(names, n)
	}
	__antithesis_instrumentation__.Notify(69687)
	idx := indexes[names[s.rnd.Intn(len(names))]]
	var refs colRefs
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, col := range idx.Columns {
		__antithesis_instrumentation__.Notify(69692)
		ref := s.columns[table][col.Column]
		if ref == nil {
			__antithesis_instrumentation__.Notify(69694)

			return nil, nil, nil, false
		} else {
			__antithesis_instrumentation__.Notify(69695)
		}
		__antithesis_instrumentation__.Notify(69693)
		refs = append(refs, &colRef{
			typ:  tree.MustBeStaticallyKnownType(ref.Type),
			item: tree.NewColumnItem(&alias, col.Column),
		})
	}
	__antithesis_instrumentation__.Notify(69688)
	return &tree.TableIndexName{
		Table: alias,
		Index: tree.UnrestrictedName(idx.Name),
	}, idx, refs, true
}

func (s *Smither) getRandIndex() (*tree.TableIndexName, *tree.CreateIndex, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69696)
	tableRef, ok := s.getRandTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69698)
		return nil, nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69699)
	}
	__antithesis_instrumentation__.Notify(69697)
	name := *tableRef.TableName
	return s.getRandTableIndex(name, name)
}

func (s *Smither) getRandUserDefinedTypeLabel() (*tree.EnumValue, *tree.TypeName, bool) {
	__antithesis_instrumentation__.Notify(69700)
	typName, ok := s.getRandUserDefinedType()
	if !ok {
		__antithesis_instrumentation__.Notify(69703)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69704)
	}
	__antithesis_instrumentation__.Notify(69701)
	s.lock.RLock()
	defer s.lock.RUnlock()
	udt := s.types.udts[*typName]
	logicalRepresentations := udt.TypeMeta.EnumData.LogicalRepresentations

	if len(logicalRepresentations) == 0 {
		__antithesis_instrumentation__.Notify(69705)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69706)
	}
	__antithesis_instrumentation__.Notify(69702)
	enumVal := tree.EnumValue(logicalRepresentations[s.rnd.Intn(len(logicalRepresentations))])
	return &enumVal, typName, true
}

func (s *Smither) getRandUserDefinedType() (*tree.TypeName, bool) {
	__antithesis_instrumentation__.Notify(69707)
	s.lock.RLock()
	defer s.lock.RUnlock()
	if s.types == nil || func() bool {
		__antithesis_instrumentation__.Notify(69710)
		return len(s.types.udts) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69711)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69712)
	}
	__antithesis_instrumentation__.Notify(69708)
	idx := s.rnd.Intn(len(s.types.udts))
	count := 0
	for typName := range s.types.udts {
		__antithesis_instrumentation__.Notify(69713)
		if count == idx {
			__antithesis_instrumentation__.Notify(69715)
			return &typName, true
		} else {
			__antithesis_instrumentation__.Notify(69716)
		}
		__antithesis_instrumentation__.Notify(69714)
		count++
	}
	__antithesis_instrumentation__.Notify(69709)
	return nil, false
}

func (s *Smither) extractTypes() (*typeInfo, error) {
	__antithesis_instrumentation__.Notify(69717)
	rows, err := s.db.Query(`
SELECT
	schema_name, descriptor_name, descriptor_id, enum_members
FROM
	crdb_internal.create_type_statements
`)
	if err != nil {
		__antithesis_instrumentation__.Notify(69721)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(69722)
	}
	__antithesis_instrumentation__.Notify(69718)
	defer rows.Close()

	evalCtx := tree.EvalContext{}
	udtMapping := make(map[tree.TypeName]*types.T)

	for rows.Next() {
		__antithesis_instrumentation__.Notify(69723)

		var scName, name string
		var id int
		var membersRaw []byte
		if err := rows.Scan(&scName, &name, &id, &membersRaw); err != nil {
			__antithesis_instrumentation__.Notify(69726)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69727)
		}
		__antithesis_instrumentation__.Notify(69724)

		var members []string
		if len(membersRaw) != 0 {
			__antithesis_instrumentation__.Notify(69728)
			arr, _, err := tree.ParseDArrayFromString(&evalCtx, string(membersRaw), types.String)
			if err != nil {
				__antithesis_instrumentation__.Notify(69730)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(69731)
			}
			__antithesis_instrumentation__.Notify(69729)
			for _, d := range arr.Array {
				__antithesis_instrumentation__.Notify(69732)
				members = append(members, string(tree.MustBeDString(d)))
			}
		} else {
			__antithesis_instrumentation__.Notify(69733)
		}
		__antithesis_instrumentation__.Notify(69725)

		switch {
		case len(members) > 0:
			__antithesis_instrumentation__.Notify(69734)
			typ := types.MakeEnum(typedesc.TypeIDToOID(descpb.ID(id)), 0)
			typ.TypeMeta = types.UserDefinedTypeMetadata{
				Name: &types.UserDefinedTypeName{
					Schema: scName,
					Name:   name,
				},
				EnumData: &types.EnumMetadata{
					LogicalRepresentations: members,

					PhysicalRepresentations: make([][]byte, len(members)),
					IsMemberReadOnly:        make([]bool, len(members)),
				},
			}
			key := tree.MakeSchemaQualifiedTypeName(scName, name)
			udtMapping[key] = typ
		default:
			__antithesis_instrumentation__.Notify(69735)
			return nil, errors.New("unsupported SQLSmith type kind")
		}
	}
	__antithesis_instrumentation__.Notify(69719)
	var udts []*types.T
	for _, t := range udtMapping {
		__antithesis_instrumentation__.Notify(69736)
		udts = append(udts, t)
	}
	__antithesis_instrumentation__.Notify(69720)

	udts = udts[:len(udts):len(udts)]

	return &typeInfo{
		udts:        udtMapping,
		scalarTypes: append(udts, types.Scalar...),
		seedTypes:   append(udts, randgen.SeedTypes...),
	}, nil
}

type schemaRef struct {
	SchemaName tree.Name
}

func (s *Smither) extractSchemas() ([]*schemaRef, error) {
	__antithesis_instrumentation__.Notify(69737)
	rows, err := s.db.Query(`
SELECT nspname FROM pg_catalog.pg_namespace
WHERE nspname NOT IN ('crdb_internal', 'pg_catalog', 'pg_extension',
		'information_schema')`)
	if err != nil {
		__antithesis_instrumentation__.Notify(69740)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(69741)
	}
	__antithesis_instrumentation__.Notify(69738)
	defer rows.Close()

	var ret []*schemaRef
	for rows.Next() {
		__antithesis_instrumentation__.Notify(69742)
		var schema tree.Name
		if err := rows.Scan(&schema); err != nil {
			__antithesis_instrumentation__.Notify(69744)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69745)
		}
		__antithesis_instrumentation__.Notify(69743)
		ret = append(ret, &schemaRef{SchemaName: schema})
	}
	__antithesis_instrumentation__.Notify(69739)
	return ret, nil
}

func (s *Smither) extractTables() ([]*tableRef, error) {
	__antithesis_instrumentation__.Notify(69746)
	rows, err := s.db.Query(`
SELECT
	table_catalog,
	table_schema,
	table_name,
	column_name,
	crdb_sql_type,
	generation_expression != '' AS computed,
	is_nullable = 'YES' AS nullable,
	is_hidden = 'YES' AS hidden
FROM
	information_schema.columns
WHERE
	table_schema NOT IN ('crdb_internal', 'pg_catalog', 'pg_extension',
	                     'information_schema')
ORDER BY
	table_catalog, table_schema, table_name
	`)

	if err != nil {
		__antithesis_instrumentation__.Notify(69751)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(69752)
	}
	__antithesis_instrumentation__.Notify(69747)
	defer rows.Close()

	firstTime := true
	var lastCatalog, lastSchema, lastName tree.Name
	var tables []*tableRef
	var currentCols []*tree.ColumnTableDef
	emit := func() error {
		__antithesis_instrumentation__.Notify(69753)
		if len(currentCols) == 0 {
			__antithesis_instrumentation__.Notify(69756)
			return fmt.Errorf("zero columns for %s.%s", lastCatalog, lastName)
		} else {
			__antithesis_instrumentation__.Notify(69757)
		}
		__antithesis_instrumentation__.Notify(69754)

		for i := range colinfo.AllSystemColumnDescs {
			__antithesis_instrumentation__.Notify(69758)
			col := &colinfo.AllSystemColumnDescs[i]
			if s.postgres && func() bool {
				__antithesis_instrumentation__.Notify(69760)
				return col.ID == colinfo.MVCCTimestampColumnID == true
			}() == true {
				__antithesis_instrumentation__.Notify(69761)
				continue
			} else {
				__antithesis_instrumentation__.Notify(69762)
			}
			__antithesis_instrumentation__.Notify(69759)
			currentCols = append(currentCols, &tree.ColumnTableDef{
				Name:   tree.Name(col.Name),
				Type:   col.Type,
				Hidden: true,
			})
		}
		__antithesis_instrumentation__.Notify(69755)
		tableName := tree.MakeTableNameWithSchema(lastCatalog, lastSchema, lastName)
		tables = append(tables, &tableRef{
			TableName: &tableName,
			Columns:   currentCols,
		})
		return nil
	}
	__antithesis_instrumentation__.Notify(69748)
	for rows.Next() {
		__antithesis_instrumentation__.Notify(69763)
		var catalog, schema, name, col tree.Name
		var typ string
		var computed, nullable, hidden bool
		if err := rows.Scan(&catalog, &schema, &name, &col, &typ, &computed, &nullable, &hidden); err != nil {
			__antithesis_instrumentation__.Notify(69771)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69772)
		}
		__antithesis_instrumentation__.Notify(69764)
		if hidden {
			__antithesis_instrumentation__.Notify(69773)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69774)
		}
		__antithesis_instrumentation__.Notify(69765)

		if firstTime {
			__antithesis_instrumentation__.Notify(69775)
			lastCatalog = catalog
			lastSchema = schema
			lastName = name
		} else {
			__antithesis_instrumentation__.Notify(69776)
		}
		__antithesis_instrumentation__.Notify(69766)
		firstTime = false

		if lastCatalog != catalog || func() bool {
			__antithesis_instrumentation__.Notify(69777)
			return lastSchema != schema == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(69778)
			return lastName != name == true
		}() == true {
			__antithesis_instrumentation__.Notify(69779)
			if err := emit(); err != nil {
				__antithesis_instrumentation__.Notify(69781)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(69782)
			}
			__antithesis_instrumentation__.Notify(69780)
			currentCols = nil
		} else {
			__antithesis_instrumentation__.Notify(69783)
		}
		__antithesis_instrumentation__.Notify(69767)

		coltyp, err := s.typeFromSQLTypeSyntax(typ)
		if err != nil {
			__antithesis_instrumentation__.Notify(69784)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69785)
		}
		__antithesis_instrumentation__.Notify(69768)
		column := tree.ColumnTableDef{
			Name: col,
			Type: coltyp,
		}
		if nullable {
			__antithesis_instrumentation__.Notify(69786)
			column.Nullable.Nullability = tree.Null
		} else {
			__antithesis_instrumentation__.Notify(69787)
		}
		__antithesis_instrumentation__.Notify(69769)
		if computed {
			__antithesis_instrumentation__.Notify(69788)
			column.Computed.Computed = true
		} else {
			__antithesis_instrumentation__.Notify(69789)
		}
		__antithesis_instrumentation__.Notify(69770)
		currentCols = append(currentCols, &column)
		lastCatalog = catalog
		lastSchema = schema
		lastName = name
	}
	__antithesis_instrumentation__.Notify(69749)
	if !firstTime {
		__antithesis_instrumentation__.Notify(69790)
		if err := emit(); err != nil {
			__antithesis_instrumentation__.Notify(69791)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69792)
		}
	} else {
		__antithesis_instrumentation__.Notify(69793)
	}
	__antithesis_instrumentation__.Notify(69750)
	return tables, rows.Err()
}

func (s *Smither) extractIndexes(
	tables tableRefs,
) (map[tree.TableName]map[tree.Name]*tree.CreateIndex, error) {
	__antithesis_instrumentation__.Notify(69794)
	ret := map[tree.TableName]map[tree.Name]*tree.CreateIndex{}

	for _, t := range tables {
		__antithesis_instrumentation__.Notify(69796)
		indexes := map[tree.Name]*tree.CreateIndex{}

		rows, err := s.db.Query(fmt.Sprintf(`
			SELECT
			    si.index_name, column_name, storing, direction = 'ASC',
          is_inverted
			FROM
			    [SHOW INDEXES FROM %s] si
      JOIN crdb_internal.table_indexes ti
           ON si.table_name = ti.descriptor_name
           AND si.index_name = ti.index_name
			WHERE
			    column_name != 'rowid'
			`, t.TableName))
		if err != nil {
			__antithesis_instrumentation__.Notify(69802)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69803)
		}
		__antithesis_instrumentation__.Notify(69797)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(69804)
			var idx, col tree.Name
			var storing, ascending, inverted bool
			if err := rows.Scan(&idx, &col, &storing, &ascending, &inverted); err != nil {
				__antithesis_instrumentation__.Notify(69807)
				rows.Close()
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(69808)
			}
			__antithesis_instrumentation__.Notify(69805)
			if _, ok := indexes[idx]; !ok {
				__antithesis_instrumentation__.Notify(69809)
				indexes[idx] = &tree.CreateIndex{
					Name:     idx,
					Table:    *t.TableName,
					Inverted: inverted,
				}
			} else {
				__antithesis_instrumentation__.Notify(69810)
			}
			__antithesis_instrumentation__.Notify(69806)
			create := indexes[idx]
			if storing {
				__antithesis_instrumentation__.Notify(69811)
				create.Storing = append(create.Storing, col)
			} else {
				__antithesis_instrumentation__.Notify(69812)
				dir := tree.Ascending
				if !ascending {
					__antithesis_instrumentation__.Notify(69814)
					dir = tree.Descending
				} else {
					__antithesis_instrumentation__.Notify(69815)
				}
				__antithesis_instrumentation__.Notify(69813)
				create.Columns = append(create.Columns, tree.IndexElem{
					Column:    col,
					Direction: dir,
				})
			}
		}
		__antithesis_instrumentation__.Notify(69798)
		if err := rows.Close(); err != nil {
			__antithesis_instrumentation__.Notify(69816)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69817)
		}
		__antithesis_instrumentation__.Notify(69799)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(69818)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(69819)
		}
		__antithesis_instrumentation__.Notify(69800)

		for name, idx := range indexes {
			__antithesis_instrumentation__.Notify(69820)
			if len(idx.Columns) == 0 {
				__antithesis_instrumentation__.Notify(69821)
				delete(indexes, name)
			} else {
				__antithesis_instrumentation__.Notify(69822)
			}
		}
		__antithesis_instrumentation__.Notify(69801)
		ret[*t.TableName] = indexes
	}
	__antithesis_instrumentation__.Notify(69795)
	return ret, nil
}

type operator struct {
	*tree.BinOp
	Operator treebin.BinaryOperator
}

var operators = func() map[oid.Oid][]operator {
	__antithesis_instrumentation__.Notify(69823)
	m := map[oid.Oid][]operator{}
	for BinaryOperator, overload := range tree.BinOps {
		__antithesis_instrumentation__.Notify(69825)
		for _, ov := range overload {
			__antithesis_instrumentation__.Notify(69826)
			bo := ov.(*tree.BinOp)
			m[bo.ReturnType.Oid()] = append(m[bo.ReturnType.Oid()], operator{
				BinOp:    bo,
				Operator: treebin.MakeBinaryOperator(BinaryOperator),
			})
		}
	}
	__antithesis_instrumentation__.Notify(69824)
	return m
}()

type function struct {
	def      *tree.FunctionDefinition
	overload *tree.Overload
}

var functions = func() map[tree.FunctionClass]map[oid.Oid][]function {
	__antithesis_instrumentation__.Notify(69827)
	m := map[tree.FunctionClass]map[oid.Oid][]function{}
	for _, def := range tree.FunDefs {
		__antithesis_instrumentation__.Notify(69829)
		switch def.Name {
		case "pg_sleep":
			__antithesis_instrumentation__.Notify(69836)
			continue
		case "st_frechetdistance", "st_buffer":
			__antithesis_instrumentation__.Notify(69837)

			continue
		default:
			__antithesis_instrumentation__.Notify(69838)
		}
		__antithesis_instrumentation__.Notify(69830)
		skip := false
		for _, substr := range []string{

			"stream_ingestion",
			"crdb_internal.force_",
			"crdb_internal.unsafe_",
			"crdb_internal.create_join_token",
			"crdb_internal.reset_multi_region_zone_configs_for_database",
			"crdb_internal.reset_index_usage_stats",
			"crdb_internal.start_replication_stream",
			"crdb_internal.replication_stream_progress",
		} {
			__antithesis_instrumentation__.Notify(69839)
			skip = skip || func() bool {
				__antithesis_instrumentation__.Notify(69840)
				return strings.Contains(def.Name, substr) == true
			}() == true
		}
		__antithesis_instrumentation__.Notify(69831)
		if skip {
			__antithesis_instrumentation__.Notify(69841)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69842)
		}
		__antithesis_instrumentation__.Notify(69832)
		if _, ok := m[def.Class]; !ok {
			__antithesis_instrumentation__.Notify(69843)
			m[def.Class] = map[oid.Oid][]function{}
		} else {
			__antithesis_instrumentation__.Notify(69844)
		}
		__antithesis_instrumentation__.Notify(69833)

		if def.Category == "Compatibility" {
			__antithesis_instrumentation__.Notify(69845)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69846)
		}
		__antithesis_instrumentation__.Notify(69834)
		if def.Private {
			__antithesis_instrumentation__.Notify(69847)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69848)
		}
		__antithesis_instrumentation__.Notify(69835)
		for _, ov := range def.Definition {
			__antithesis_instrumentation__.Notify(69849)
			ov := ov.(*tree.Overload)

			if strings.Contains(ov.Info, "Not usable") {
				__antithesis_instrumentation__.Notify(69853)
				continue
			} else {
				__antithesis_instrumentation__.Notify(69854)
			}
			__antithesis_instrumentation__.Notify(69850)
			typ := ov.FixedReturnType()
			found := false
			for _, scalarTyp := range types.Scalar {
				__antithesis_instrumentation__.Notify(69855)
				if typ.Family() == scalarTyp.Family() {
					__antithesis_instrumentation__.Notify(69856)
					found = true
				} else {
					__antithesis_instrumentation__.Notify(69857)
				}
			}
			__antithesis_instrumentation__.Notify(69851)
			if !found {
				__antithesis_instrumentation__.Notify(69858)
				continue
			} else {
				__antithesis_instrumentation__.Notify(69859)
			}
			__antithesis_instrumentation__.Notify(69852)
			m[def.Class][typ.Oid()] = append(m[def.Class][typ.Oid()], function{
				def:      def,
				overload: ov,
			})
		}
	}
	__antithesis_instrumentation__.Notify(69828)
	return m
}()
