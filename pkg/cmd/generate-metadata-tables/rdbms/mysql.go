package rdbms

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

const mysqlDescribeSchema = `
	SELECT 
		table_name, 
		column_name, 
		data_type 
	FROM information_schema.columns
	WHERE table_schema = ?
	ORDER BY table_name
`

var mysqlExclusions = []*excludePattern{
	{
		pattern: regexp.MustCompile(`innodb_.+`),
		except:  make(map[string]struct{}),
	},
}

type mysqlMetadataConnection struct {
	*gosql.DB
	catalog string
}

func mysqlConnect(address, user, catalog string) (DBMetadataConnection, error) {
	__antithesis_instrumentation__.Notify(40369)
	db, err := gosql.Open("mysql", fmt.Sprintf("%s@tcp(%s)/%s", user, address, catalog))
	if err != nil {
		__antithesis_instrumentation__.Notify(40371)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40372)
	}
	__antithesis_instrumentation__.Notify(40370)
	return mysqlMetadataConnection{db, catalog}, nil
}

func (conn mysqlMetadataConnection) DatabaseVersion(
	ctx context.Context,
) (version string, err error) {
	__antithesis_instrumentation__.Notify(40373)
	row := conn.QueryRowContext(ctx, "SELECT version()")
	err = row.Scan(&version)
	return version, err
}

func (conn mysqlMetadataConnection) DescribeSchema(
	ctx context.Context,
) (*ColumnMetadataList, error) {
	__antithesis_instrumentation__.Notify(40374)
	metadata := &ColumnMetadataList{exclusions: mysqlExclusions}
	rows, err := conn.QueryContext(ctx, mysqlDescribeSchema, conn.catalog)
	if err != nil {
		__antithesis_instrumentation__.Notify(40377)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40378)
	}
	__antithesis_instrumentation__.Notify(40375)
	defer rows.Close()

	for rows.Next() {
		__antithesis_instrumentation__.Notify(40379)
		row := new(columnMetadata)
		if err := rows.Scan(&row.tableName, &row.columnName, &row.dataTypeName); err != nil {
			__antithesis_instrumentation__.Notify(40381)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(40382)
		}
		__antithesis_instrumentation__.Notify(40380)
		row.tableName = strings.ToLower(row.tableName)
		row.columnName = strings.ToLower(row.columnName)
		metadata.data = append(metadata.data, row)
	}
	__antithesis_instrumentation__.Notify(40376)

	return metadata, nil
}

func (conn mysqlMetadataConnection) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(40383)
	return conn.DB.Close()
}
