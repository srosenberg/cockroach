package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
)

func (og *operationGenerator) tableExists(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695751)
	return og.scanBool(ctx, tx, `SELECT EXISTS (
	SELECT table_name
    FROM information_schema.tables 
   WHERE table_schema = $1
     AND table_name = $2
   )`, tableName.Schema(), tableName.Object())
}

func (og *operationGenerator) viewExists(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695752)
	return og.scanBool(ctx, tx, `SELECT EXISTS (
	SELECT table_name
    FROM information_schema.views 
   WHERE table_schema = $1
     AND table_name = $2
   )`, tableName.Schema(), tableName.Object())
}

func (og *operationGenerator) sequenceExists(
	ctx context.Context, tx pgx.Tx, seqName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695753)
	return og.scanBool(ctx, tx, `SELECT EXISTS (
	SELECT sequence_name
    FROM information_schema.sequences
   WHERE sequence_schema = $1
     AND sequence_name = $2
   )`, seqName.Schema(), seqName.Object())
}

func (og *operationGenerator) columnExistsOnTable(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695754)
	return og.scanBool(ctx, tx, `SELECT EXISTS (
	SELECT column_name
    FROM information_schema.columns 
   WHERE table_schema = $1
     AND table_name = $2
     AND column_name = $3
   )`, tableName.Schema(), tableName.Object(), columnName)
}

func (og *operationGenerator) tableHasRows(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695755)
	return og.scanBool(ctx, tx, fmt.Sprintf(`SELECT EXISTS (SELECT * FROM %s)`, tableName.String()))
}

func (og *operationGenerator) scanBool(
	ctx context.Context, tx pgx.Tx, query string, args ...interface{},
) (b bool, err error) {
	__antithesis_instrumentation__.Notify(695756)
	err = tx.QueryRow(ctx, query, args...).Scan(&b)
	if err == nil {
		__antithesis_instrumentation__.Notify(695758)
		og.LogQueryResults(
			fmt.Sprintf("%q %q", query, args),
			fmt.Sprintf("%t", b),
		)
	} else {
		__antithesis_instrumentation__.Notify(695759)
	}
	__antithesis_instrumentation__.Notify(695757)
	return b, errors.Wrapf(err, "scanBool: %q %q", query, args)
}

func scanString(
	ctx context.Context, tx pgx.Tx, query string, args ...interface{},
) (s string, err error) {
	__antithesis_instrumentation__.Notify(695760)
	err = tx.QueryRow(ctx, query, args...).Scan(&s)
	return s, errors.Wrapf(err, "scanString: %q %q", query, args)
}

func (og *operationGenerator) schemaExists(
	ctx context.Context, tx pgx.Tx, schemaName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695761)
	return og.scanBool(ctx, tx, `SELECT EXISTS (
	SELECT schema_name
		FROM information_schema.schemata
   WHERE schema_name = $1
	)`, schemaName)
}

func (og *operationGenerator) tableHasDependencies(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695762)
	return og.scanBool(ctx, tx, `
	SELECT EXISTS(
        SELECT fd.descriptor_name
          FROM crdb_internal.forward_dependencies AS fd
         WHERE fd.descriptor_id
               = (
                    SELECT c.oid
                      FROM pg_catalog.pg_class AS c
                      JOIN pg_catalog.pg_namespace AS ns ON
                            ns.oid = c.relnamespace
                     WHERE c.relname = $1 AND ns.nspname = $2
                )
           AND fd.descriptor_id != fd.dependedonby_id
           AND fd.dependedonby_type != 'sequence'
       )
	`, tableName.Object(), tableName.Schema())
}

func (og *operationGenerator) columnIsDependedOn(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695763)

	return og.scanBool(ctx, tx, `SELECT EXISTS(
		SELECT source.column_id
			FROM (
			   SELECT DISTINCT column_id
			     FROM (
			           SELECT unnest(
			                   string_to_array(
			                    rtrim(
			                     ltrim(
			                      fd.dependedonby_details,
			                      'Columns: ['
			                     ),
			                     ']'
			                    ),
			                    ' '
			                   )::INT8[]
			                  ) AS column_id
			             FROM crdb_internal.forward_dependencies
			                   AS fd
			            WHERE fd.descriptor_id
			                  = $1::REGCLASS
                    AND fd.dependedonby_type != 'sequence'
			          )
			   UNION  (
			           SELECT unnest(confkey) AS column_id
			             FROM pg_catalog.pg_constraint
			            WHERE confrelid = $1::REGCLASS
			          )
			 ) AS cons
			 INNER JOIN (
			   SELECT ordinal_position AS column_id
			     FROM information_schema.columns
			    WHERE table_schema = $2
			      AND table_name = $3
			      AND column_name = $4
			  ) AS source ON source.column_id = cons.column_id
)`, tableName.String(), tableName.Schema(), tableName.Object(), columnName)
}

func (og *operationGenerator) colIsPrimaryKey(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695764)
	primaryColumns, err := og.scanStringArray(ctx, tx,
		`
SELECT array_agg(column_name)
  FROM (
        SELECT DISTINCT column_name
          FROM information_schema.statistics
         WHERE index_name
               IN (
                  SELECT index_name
                    FROM crdb_internal.table_indexes
                   WHERE index_type = 'primary' AND descriptor_id = $3::REGCLASS
                )
               AND table_schema = $1
               AND table_name = $2
               AND storing = 'NO'
       );
	`, tableName.Schema(), tableName.Object(), tableName.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(695767)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(695768)
	}
	__antithesis_instrumentation__.Notify(695765)

	for _, primaryColumn := range primaryColumns {
		__antithesis_instrumentation__.Notify(695769)
		if primaryColumn == columnName {
			__antithesis_instrumentation__.Notify(695770)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(695771)
		}
	}
	__antithesis_instrumentation__.Notify(695766)
	return false, nil
}

func (og *operationGenerator) violatesUniqueConstraints(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columns []string, rows [][]string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695772)

	if len(rows) == 0 {
		__antithesis_instrumentation__.Notify(695776)
		return false, fmt.Errorf("violatesUniqueConstraints: no rows provided")
	} else {
		__antithesis_instrumentation__.Notify(695777)
	}
	__antithesis_instrumentation__.Notify(695773)

	constraints, err := scanStringArrayRows(ctx, tx, `
	 SELECT DISTINCT array_agg(cols.column_name ORDER BY cols.column_name)
					    FROM (
					          SELECT d.oid,
					                 d.table_name,
					                 d.schema_name,
					                 conname,
					                 contype,
					                 unnest(conkey) AS position
					            FROM (
					                  SELECT c.oid AS oid,
					                         c.relname AS table_name,
					                         ns.nspname AS schema_name
					                    FROM pg_catalog.pg_class AS c
					                    JOIN pg_catalog.pg_namespace AS ns ON
					                          ns.oid = c.relnamespace
					                   WHERE ns.nspname = $1
					                     AND c.relname = $2
					                 ) AS d
					            JOIN (
					                  SELECT conname, conkey, conrelid, contype
					                    FROM pg_catalog.pg_constraint
					                   WHERE contype = 'p' OR contype = 'u'
					                 ) ON conrelid = d.oid
					         ) AS cons
					    JOIN (
					          SELECT table_name,
					                 table_schema,
					                 column_name,
					                 ordinal_position
					            FROM information_schema.columns
					         ) AS cols ON cons.schema_name = cols.table_schema
					                  AND cols.table_name = cons.table_name
					                  AND cols.ordinal_position = cons.position
					GROUP BY cons.conname;
`, tableName.Schema(), tableName.Object())
	if err != nil {
		__antithesis_instrumentation__.Notify(695778)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(695779)
	}
	__antithesis_instrumentation__.Notify(695774)

	for _, constraint := range constraints {
		__antithesis_instrumentation__.Notify(695780)

		previousRows := map[string]bool{}
		for _, row := range rows {
			__antithesis_instrumentation__.Notify(695781)
			violation, err := og.violatesUniqueConstraintsHelper(
				ctx, tx, tableName, columns, constraint, row, previousRows,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(695783)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(695784)
			}
			__antithesis_instrumentation__.Notify(695782)
			if violation {
				__antithesis_instrumentation__.Notify(695785)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(695786)
			}
		}
	}
	__antithesis_instrumentation__.Notify(695775)

	return false, nil
}

var ErrSchemaChangesDisallowedDueToPkSwap = errors.New("not schema changes allowed on selected table due to PK swap")

func (og *operationGenerator) tableHasPrimaryKeySwapActive(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) error {
	__antithesis_instrumentation__.Notify(695787)

	indexName, err := og.scanStringArray(
		ctx,
		tx,
		`
SELECT array_agg(index_name)
  FROM (
SELECT
	index_name
FROM
	crdb_internal.table_indexes
WHERE
	index_type = 'primary'
	AND descriptor_id = $1::REGCLASS
       );
	`, tableName.String(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(695791)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695792)
	}
	__antithesis_instrumentation__.Notify(695788)

	allowed, err := og.scanBool(
		ctx,
		tx,
		`
SELECT count(*) > 0
  FROM crdb_internal.schema_changes
 WHERE type = 'INDEX'
       AND table_id = $1::REGCLASS
       AND  target_name = $2
       AND direction = 'DROP';
`,
		tableName.String(),
		indexName[0],
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(695793)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695794)
	}
	__antithesis_instrumentation__.Notify(695789)
	if !allowed {
		__antithesis_instrumentation__.Notify(695795)
		return ErrSchemaChangesDisallowedDueToPkSwap
	} else {
		__antithesis_instrumentation__.Notify(695796)
	}
	__antithesis_instrumentation__.Notify(695790)
	return nil
}

func (og *operationGenerator) violatesUniqueConstraintsHelper(
	ctx context.Context,
	tx pgx.Tx,
	tableName *tree.TableName,
	columns []string,
	constraint []string,
	row []string,
	previousRows map[string]bool,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695797)

	columnsToValues := map[string]string{}
	for i := 0; i < len(columns); i++ {
		__antithesis_instrumentation__.Notify(695804)
		columnsToValues[columns[i]] = row[i]
	}
	__antithesis_instrumentation__.Notify(695798)

	query := strings.Builder{}
	query.WriteString(fmt.Sprintf(`SELECT EXISTS (
			SELECT *
				FROM %s
       WHERE
		`, tableName.String()))

	atLeastOneNonNullValue := false
	for _, column := range constraint {
		__antithesis_instrumentation__.Notify(695805)

		if columnsToValues[column] != "NULL" {
			__antithesis_instrumentation__.Notify(695806)
			if atLeastOneNonNullValue {
				__antithesis_instrumentation__.Notify(695808)
				query.WriteString(fmt.Sprintf(` AND %s = %s`, column, columnsToValues[column]))
			} else {
				__antithesis_instrumentation__.Notify(695809)
				query.WriteString(fmt.Sprintf(`%s = %s`, column, columnsToValues[column]))
			}
			__antithesis_instrumentation__.Notify(695807)

			atLeastOneNonNullValue = true
		} else {
			__antithesis_instrumentation__.Notify(695810)
		}
	}
	__antithesis_instrumentation__.Notify(695799)
	query.WriteString(")")

	if !atLeastOneNonNullValue {
		__antithesis_instrumentation__.Notify(695811)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(695812)
	}
	__antithesis_instrumentation__.Notify(695800)

	queryString := query.String()

	if _, duplicateEntry := previousRows[queryString]; duplicateEntry {
		__antithesis_instrumentation__.Notify(695813)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(695814)
	}
	__antithesis_instrumentation__.Notify(695801)
	previousRows[queryString] = true

	exists, err := og.scanBool(ctx, tx, queryString)
	if err != nil {
		__antithesis_instrumentation__.Notify(695815)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(695816)
	}
	__antithesis_instrumentation__.Notify(695802)
	if exists {
		__antithesis_instrumentation__.Notify(695817)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(695818)
	}
	__antithesis_instrumentation__.Notify(695803)

	return false, nil
}

func scanStringArrayRows(
	ctx context.Context, tx pgx.Tx, query string, args ...interface{},
) ([][]string, error) {
	__antithesis_instrumentation__.Notify(695819)
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		__antithesis_instrumentation__.Notify(695822)
		return nil, errors.Wrapf(err, "scanStringArrayRows: %q %q", query, args)
	} else {
		__antithesis_instrumentation__.Notify(695823)
	}
	__antithesis_instrumentation__.Notify(695820)
	defer rows.Close()

	results := [][]string{}
	for rows.Next() {
		__antithesis_instrumentation__.Notify(695824)
		var columnNames []string
		err := rows.Scan(&columnNames)
		if err != nil {
			__antithesis_instrumentation__.Notify(695826)
			return nil, errors.Wrapf(err, "scan: %q, args %q, scanArgs %q", query, columnNames, args)
		} else {
			__antithesis_instrumentation__.Notify(695827)
		}
		__antithesis_instrumentation__.Notify(695825)
		results = append(results, columnNames)
	}
	__antithesis_instrumentation__.Notify(695821)

	return results, nil
}

func (og *operationGenerator) indexExists(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, indexName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695828)
	return og.scanBool(ctx, tx, `SELECT EXISTS(
			SELECT *
			  FROM information_schema.statistics
			 WHERE table_schema = $1
			   AND table_name = $2
			   AND index_name = $3
  )`, tableName.Schema(), tableName.Object(), indexName)
}

func (og *operationGenerator) scanStringArray(
	ctx context.Context, tx pgx.Tx, query string, args ...interface{},
) (b []string, err error) {
	__antithesis_instrumentation__.Notify(695829)
	err = tx.QueryRow(ctx, query, args...).Scan(&b)
	if err == nil {
		__antithesis_instrumentation__.Notify(695831)
		og.LogQueryResultArray(
			fmt.Sprintf("%q %q", query, args),
			b,
		)
	} else {
		__antithesis_instrumentation__.Notify(695832)
	}
	__antithesis_instrumentation__.Notify(695830)
	return b, errors.Wrapf(err, "scanStringArray %q %q", query, args)
}

func (og *operationGenerator) canApplyUniqueConstraint(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columns []string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695833)
	columnNames := strings.Join(columns, ", ")

	whereNotNullClause := strings.Builder{}
	for idx, column := range columns {
		__antithesis_instrumentation__.Notify(695835)
		whereNotNullClause.WriteString(fmt.Sprintf("%s IS NOT NULL ", column))
		if idx != len(columns)-1 {
			__antithesis_instrumentation__.Notify(695836)
			whereNotNullClause.WriteString("OR ")
		} else {
			__antithesis_instrumentation__.Notify(695837)
		}
	}
	__antithesis_instrumentation__.Notify(695834)

	return og.scanBool(ctx, tx,
		fmt.Sprintf(`
		SELECT (
	       SELECT count(*)
	         FROM (
	               SELECT DISTINCT %s
	                 FROM %s
	                WHERE %s
	              )
	      )
	      = (
	        SELECT count(*)
	          FROM %s
	         WHERE %s
	       );
	`, columnNames, tableName.String(), whereNotNullClause.String(), tableName.String(), whereNotNullClause.String()))

}

func (og *operationGenerator) columnContainsNull(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695838)
	return og.scanBool(ctx, tx, fmt.Sprintf(`SELECT EXISTS (
		SELECT %s
		  FROM %s
	   WHERE %s IS NULL
	)`, columnName, tableName.String(), columnName))
}

func (og *operationGenerator) constraintIsPrimary(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, constraintName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695839)
	return og.scanBool(ctx, tx, fmt.Sprintf(`
	SELECT EXISTS(
	        SELECT *
	          FROM pg_catalog.pg_constraint
	         WHERE conrelid = '%s'::REGCLASS::INT
	           AND conname = '%s'
	           AND (contype = 'p')
	       );
	`, tableName.String(), constraintName))
}

func (og *operationGenerator) columnHasSingleUniqueConstraint(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695840)
	return og.scanBool(ctx, tx, `
	SELECT EXISTS(
	        SELECT column_name
	          FROM (
	                SELECT table_schema, table_name, column_name, ordinal_position,
	                       concat(table_schema,'.',table_name)::REGCLASS::INT8 AS tableid
	                  FROM information_schema.columns
	               ) AS cols
	          JOIN (
	                SELECT contype, conkey, conrelid
	                  FROM pg_catalog.pg_constraint
	               ) AS cons ON cons.conrelid = cols.tableid
	         WHERE table_schema = $1
	           AND table_name = $2
	           AND column_name = $3
	           AND (contype = 'u' OR contype = 'p')
	           AND array_length(conkey, 1) = 1
					   AND conkey[1] = ordinal_position
	       )
	`, tableName.Schema(), tableName.Object(), columnName)
}
func (og *operationGenerator) constraintIsUnique(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, constraintName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695841)
	return og.scanBool(ctx, tx, fmt.Sprintf(`
	SELECT EXISTS(
	        SELECT *
	          FROM pg_catalog.pg_constraint
	         WHERE conrelid = '%s'::REGCLASS::INT
	           AND conname = '%s'
	           AND (contype = 'u')
	       );
	`, tableName.String(), constraintName))
}

func (og *operationGenerator) columnIsStoredComputed(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695842)

	return og.scanBool(ctx, tx, `
SELECT COALESCE(
        (
            SELECT attgenerated
              FROM pg_catalog.pg_attribute
             WHERE attrelid = $1:::REGCLASS AND attname = $2
        )
        = 's',
        false
       );
`, tableName.String(), columnName)
}

func (og *operationGenerator) columnIsComputed(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695843)

	return og.scanBool(ctx, tx, `
SELECT COALESCE(
        (
            SELECT attgenerated
              FROM pg_catalog.pg_attribute
             WHERE attrelid = $1:::REGCLASS AND attname = $2
        )
        != '',
        false
       );
`, tableName.String(), columnName)
}

func (og *operationGenerator) constraintExists(
	ctx context.Context, tx pgx.Tx, constraintName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695844)
	return og.scanBool(ctx, tx, fmt.Sprintf(`
	SELECT EXISTS(
	        SELECT *
	          FROM pg_catalog.pg_constraint
	           WHERE conname = '%s'
	       );
	`, constraintName))
}

func (og *operationGenerator) rowsSatisfyFkConstraint(
	ctx context.Context,
	tx pgx.Tx,
	parentTable *tree.TableName,
	parentColumn *column,
	childTable *tree.TableName,
	childColumn *column,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695845)

	if parentTable.Schema() == childTable.Schema() && func() bool {
		__antithesis_instrumentation__.Notify(695847)
		return parentTable.Object() == childTable.Object() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(695848)
		return parentColumn.name == childColumn.name == true
	}() == true {
		__antithesis_instrumentation__.Notify(695849)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(695850)
	}
	__antithesis_instrumentation__.Notify(695846)
	return og.scanBool(ctx, tx, fmt.Sprintf(`
	SELECT NOT EXISTS(
	  SELECT *
	    FROM %s as t1
		  LEFT JOIN %s as t2
				     ON t1.%s = t2.%s
	   WHERE t2.%s IS NULL
  )`, childTable.String(), parentTable.String(), childColumn.name, parentColumn.name, parentColumn.name))
}

func (og *operationGenerator) violatesFkConstraints(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columns []string, rows [][]string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695851)
	fkConstraints, err := scanStringArrayRows(ctx, tx, fmt.Sprintf(`
		SELECT array[parent.table_schema, parent.table_name, parent.column_name, child.column_name]
		  FROM (
		        SELECT conkey, confkey, conrelid, confrelid
		          FROM pg_constraint
		         WHERE contype = 'f'
		           AND conrelid = '%s'::REGCLASS::INT8
		       ) AS con
		  JOIN (
		        SELECT column_name, ordinal_position, column_default
		          FROM information_schema.columns
		         WHERE table_schema = '%s'
		           AND table_name = '%s'
		       ) AS child ON conkey[1] = child.ordinal_position
		  JOIN (
		        SELECT pc.oid,
		               cols.table_schema,
		               cols.table_name,
		               cols.column_name,
		               cols.ordinal_position
		          FROM pg_class AS pc
		          JOIN pg_namespace AS pn ON pc.relnamespace = pn.oid
		          JOIN information_schema.columns AS cols ON (pc.relname = cols.table_name AND pn.nspname = cols.table_schema)
		       ) AS parent ON (
		                       con.confkey[1] = parent.ordinal_position
		                       AND con.confrelid = parent.oid
		                      )
		 WHERE child.column_name != 'rowid';
`, tableName.String(), tableName.Schema(), tableName.Object()))
	if err != nil {
		__antithesis_instrumentation__.Notify(695855)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(695856)
	}
	__antithesis_instrumentation__.Notify(695852)

	columnNameToIndexMap := map[string]int{}
	for i, name := range columns {
		__antithesis_instrumentation__.Notify(695857)
		columnNameToIndexMap[name] = i
	}
	__antithesis_instrumentation__.Notify(695853)
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(695858)
		for _, constraint := range fkConstraints {
			__antithesis_instrumentation__.Notify(695859)
			parentTableSchema := constraint[0]
			parentTableName := constraint[1]
			parentColumnName := constraint[2]
			childColumnName := constraint[3]

			if parentTableSchema == tableName.Schema() && func() bool {
				__antithesis_instrumentation__.Notify(695862)
				return parentTableName == tableName.Object() == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(695863)
				return parentColumnName == childColumnName == true
			}() == true {
				__antithesis_instrumentation__.Notify(695864)
				continue
			} else {
				__antithesis_instrumentation__.Notify(695865)
			}
			__antithesis_instrumentation__.Notify(695860)

			violation, err := og.violatesFkConstraintsHelper(
				ctx, tx, columnNameToIndexMap, parentTableSchema, parentTableName, parentColumnName, childColumnName, row,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(695866)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(695867)
			}
			__antithesis_instrumentation__.Notify(695861)

			if violation {
				__antithesis_instrumentation__.Notify(695868)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(695869)
			}
		}
	}
	__antithesis_instrumentation__.Notify(695854)

	return false, nil
}

func (og *operationGenerator) violatesFkConstraintsHelper(
	ctx context.Context,
	tx pgx.Tx,
	columnNameToIndexMap map[string]int,
	parentTableSchema, parentTableName, parentColumn, childColumn string,
	row []string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695870)

	childValue := row[columnNameToIndexMap[childColumn]]
	if childValue == "NULL" {
		__antithesis_instrumentation__.Notify(695872)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(695873)
	}
	__antithesis_instrumentation__.Notify(695871)

	return og.scanBool(ctx, tx, fmt.Sprintf(`
	SELECT NOT EXISTS (
	    SELECT * from %s.%s
	    WHERE %s = %s
	)
	`, parentTableSchema, parentTableName, parentColumn, childValue))
}

func (og *operationGenerator) columnIsInDroppingIndex(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695874)
	return og.scanBool(ctx, tx, `
SELECT EXISTS(
        SELECT index_id
          FROM (
                SELECT DISTINCT index_id
                  FROM crdb_internal.index_columns
                 WHERE descriptor_id = $1::REGCLASS AND column_name = $2
               ) AS indexes
          JOIN crdb_internal.schema_changes AS sc ON sc.target_id
                                                     = indexes.index_id
                                                 AND table_id = $1::REGCLASS
                                                 AND type = 'INDEX'
                                                 AND direction = 'DROP'
       );
`, tableName.String(), columnName)
}

const descriptorsAndConstraintMutationsCTE = `descriptors AS (
                    SELECT crdb_internal.pb_to_json(
                            'cockroach.sql.sqlbase.Descriptor',
                            descriptor
                           )->'table' AS d
                      FROM system.descriptor
                     WHERE id = $1::REGCLASS
                   ),
       constraint_mutations AS (
                                SELECT mut
                                  FROM (
                                        SELECT json_array_elements(
                                                d->'mutations'
                                               ) AS mut
                                          FROM descriptors
                                       )
                                 WHERE (mut->'constraint') IS NOT NULL
                            )`

func (og *operationGenerator) constraintInDroppingState(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, constraintName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695875)

	return og.scanBool(ctx, tx, `
  WITH `+descriptorsAndConstraintMutationsCTE+`
SELECT true
       IN (
            SELECT (t.f).value @> json_set('{"validity": "Dropping"}', ARRAY['name'], to_json($2:::STRING))
              FROM (
                    SELECT json_each(mut->'constraint') AS f
                      FROM constraint_mutations
                   ) AS t
        );
`, tableName.String(), constraintName)
}

func (og *operationGenerator) columnNotNullConstraintInMutation(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName, columnName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695876)
	return og.scanBool(ctx, tx, `
  WITH `+descriptorsAndConstraintMutationsCTE+`,
       col AS (
            SELECT (c->>'id')::INT8 AS id
              FROM (
                    SELECT json_array_elements(d->'columns') AS c
                      FROM descriptors
                   )
             WHERE c->>'name' = $2
           )
SELECT EXISTS(
        SELECT *
          FROM constraint_mutations
          JOIN col ON mut->'constraint'->>'constraintType' = 'NOT_NULL'
                  AND (mut->'constraint'->>'notNullColumn')::INT8 = id
       );
`, tableName.String(), columnName)
}

func (og *operationGenerator) schemaContainsTypesWithCrossSchemaReferences(
	ctx context.Context, tx pgx.Tx, schemaName string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695877)
	return og.scanBool(ctx, tx, `
  WITH database_id AS (
                    SELECT id
                      FROM system.namespace
                     WHERE "parentID" = 0
                       AND "parentSchemaID" = 0
                       AND name = current_database()
                   ),
       schema_id AS (
                    SELECT nsp.id
                      FROM system.namespace AS nsp
                      JOIN database_id ON "parentID" = database_id.id
                                      AND "parentSchemaID" = 0
                                      AND name = $1
                 ),
       descriptor_ids AS (
                        SELECT nsp.id
                          FROM system.namespace AS nsp,
                               schema_id,
                               database_id
                         WHERE nsp."parentID" = database_id.id
                           AND nsp."parentSchemaID" = schema_id.id
                      ),
       descriptors AS (
                    SELECT crdb_internal.pb_to_json(
                            'cockroach.sql.sqlbase.Descriptor',
                            descriptor
                           ) AS descriptor
                      FROM system.descriptor AS descriptors
                      JOIN descriptor_ids ON descriptors.id
                                             = descriptor_ids.id
                   ),
       types AS (
                SELECT descriptor
                  FROM descriptors
                 WHERE (descriptor->'type') IS NOT NULL
             ),
       table_references AS (
                            SELECT json_array_elements(
                                    descriptor->'table'->'dependedOnBy'
                                   ) AS ref
                              FROM descriptors
                             WHERE (descriptor->'table') IS NOT NULL
                        ),
       dependent AS (
                    SELECT (ref->>'id')::INT8 AS id FROM table_references
                 ),
       referenced_descriptors AS (
                                SELECT json_array_elements_text(
                                        descriptor->'type'->'referencingDescriptorIds'
                                       )::INT8 AS id
                                  FROM types
                              )
SELECT EXISTS(
        SELECT *
          FROM system.namespace
         WHERE id IN (SELECT id FROM referenced_descriptors)
           AND "parentSchemaID" NOT IN (SELECT id FROM schema_id)
           AND id NOT IN (SELECT id FROM dependent)
       );`, schemaName)
}

func (og *operationGenerator) enumMemberPresent(
	ctx context.Context, tx pgx.Tx, enum string, val string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695878)
	return og.scanBool(ctx, tx, `
WITH enum_members AS (
	SELECT
				json_array_elements(
						crdb_internal.pb_to_json(
								'cockroach.sql.sqlbase.Descriptor',
								descriptor
						)->'type'->'enumMembers'
				)->>'logicalRepresentation'
				AS v
		FROM
				system.descriptor
		WHERE
				id = ($1::REGTYPE::INT8 - 100000)
)
SELECT
	CASE WHEN EXISTS (
		SELECT v FROM enum_members WHERE v = $2::string
	) THEN true
	ELSE false
	END AS exists
`,
		enum,
		val,
	)
}

func (og *operationGenerator) tableHasOngoingSchemaChanges(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695879)
	return og.scanBool(
		ctx,
		tx,
		`
SELECT
	json_array_length(
		COALESCE(
			crdb_internal.pb_to_json(
				'cockroach.sql.sqlbase.Descriptor',
				descriptor
			)->'table'->'mutations',
			'[]'
		)
	)
	> 0
FROM
	system.descriptor
WHERE
	id = $1::REGCLASS;
		`,
		tableName.String(),
	)
}

func (og *operationGenerator) tableHasOngoingAlterPKSchemaChanges(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695880)
	return og.scanBool(
		ctx,
		tx,
		`
WITH
	descriptors
		AS (
			SELECT
				crdb_internal.pb_to_json(
					'cockroach.sql.sqlbase.Descriptor',
					descriptor
				)->'table'
					AS d
			FROM
				system.descriptor
			WHERE
				id = $1::REGCLASS
		)
SELECT
	EXISTS(
		SELECT
			mut
		FROM
			(
				SELECT
					json_array_elements(d->'mutations')
						AS mut
				FROM
					descriptors
			)
		WHERE
			(mut->'primaryKeySwap') IS NOT NULL
	);
		`,
		tableName.String(),
	)
}

func (og *operationGenerator) getRegionColumn(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (string, error) {
	__antithesis_instrumentation__.Notify(695881)
	isTableRegionalByRow, err := og.tableIsRegionalByRow(ctx, tx, tableName)
	if err != nil {
		__antithesis_instrumentation__.Notify(695885)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695886)
	}
	__antithesis_instrumentation__.Notify(695882)
	if !isTableRegionalByRow {
		__antithesis_instrumentation__.Notify(695887)
		return "", errors.AssertionFailedf(
			"invalid call to get region column of table %s which is not a REGIONAL BY ROW table",
			tableName.String())
	} else {
		__antithesis_instrumentation__.Notify(695888)
	}
	__antithesis_instrumentation__.Notify(695883)

	regionCol, err := scanString(
		ctx,
		tx,
		`
WITH
	descriptors
		AS (
			SELECT
				crdb_internal.pb_to_json(
					'cockroach.sql.sqlbase.Descriptor',
					descriptor
				)->'table'
					AS d
			FROM
				system.descriptor
			WHERE
				id = $1::REGCLASS
		)
SELECT
	COALESCE (d->'localityConfig'->'regionalByRow'->>'as', $2)
FROM
	descriptors;
`,
		tableName.String(),
		tree.RegionalByRowRegionDefaultCol,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(695889)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(695890)
	}
	__antithesis_instrumentation__.Notify(695884)

	return regionCol, nil
}

func (og *operationGenerator) tableIsRegionalByRow(
	ctx context.Context, tx pgx.Tx, tableName *tree.TableName,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695891)
	return og.scanBool(
		ctx,
		tx,
		`
WITH
	descriptors
		AS (
			SELECT
				crdb_internal.pb_to_json(
					'cockroach.sql.sqlbase.Descriptor',
					descriptor
				)->'table'
					AS d
			FROM
				system.descriptor
			WHERE
				id = $1::REGCLASS
		)
SELECT
	EXISTS(
		SELECT
			1
		FROM
			descriptors
		WHERE
			d->'localityConfig'->'regionalByRow' IS NOT NULL
	);
		`,
		tableName.String(),
	)
}

func (og *operationGenerator) databaseIsMultiRegion(ctx context.Context, tx pgx.Tx) (bool, error) {
	__antithesis_instrumentation__.Notify(695892)
	return og.scanBool(
		ctx,
		tx,
		`SELECT EXISTS (SELECT * FROM [SHOW REGIONS FROM DATABASE])`,
	)
}

func (og *operationGenerator) databaseHasRegionChange(
	ctx context.Context, tx pgx.Tx,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695893)
	isMultiRegion, err := og.scanBool(
		ctx,
		tx,
		`SELECT EXISTS (SELECT * FROM [SHOW REGIONS FROM DATABASE])`,
	)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(695895)
		return !isMultiRegion == true
	}() == true {
		__antithesis_instrumentation__.Notify(695896)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(695897)
	}
	__antithesis_instrumentation__.Notify(695894)
	return og.scanBool(
		ctx,
		tx,
		`
WITH enum_members AS (
	SELECT
				json_array_elements(
						crdb_internal.pb_to_json(
								'cockroach.sql.sqlbase.Descriptor',
								descriptor
						)->'type'->'enumMembers'
				)
				AS v
		FROM
				system.descriptor
		WHERE
				id = ('public.crdb_internal_region'::REGTYPE::INT8 - 100000)
)
SELECT EXISTS (
	SELECT 1 FROM enum_members
	WHERE v->>'direction' <> 'NONE'
)
		`,
	)
}

func (og *operationGenerator) databaseHasRegionalByRowChange(
	ctx context.Context, tx pgx.Tx,
) (bool, error) {
	__antithesis_instrumentation__.Notify(695898)
	return og.scanBool(
		ctx,
		tx,
		`
WITH
	descriptors
		AS (
			SELECT
				crdb_internal.pb_to_json(
					'cockroach.sql.sqlbase.Descriptor',
					descriptor
				)->'table'
					AS d
			FROM
				system.descriptor
			WHERE
				id IN (
					SELECT id FROM system.namespace
					WHERE "parentID" = (
						SELECT id FROM system.namespace
						WHERE name = (SELECT database FROM [SHOW DATABASE])
						AND "parentID" = 0
					) AND "parentSchemaID" <> 0
				)
		)
SELECT (
	EXISTS(
		SELECT
			mut
		FROM
			(
				-- no schema changes on regional by row tables
				SELECT
					json_array_elements(d->'mutations')
						AS mut
				FROM (
					SELECT
						d
					FROM
						descriptors
					WHERE
						d->'localityConfig'->'regionalByRow' IS NOT NULL
				)
			)
	) OR EXISTS (
		-- no primary key swaps in the current database
		SELECT mut FROM (
			SELECT
				json_array_elements(d->'mutations')
					AS mut
			FROM descriptors
		)
		WHERE
			(mut->'primaryKeySwap') IS NOT NULL
	)
);
		`,
	)
}
