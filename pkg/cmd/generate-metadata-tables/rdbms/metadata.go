package rdbms

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/sql"
)

var ConnectFns = map[string]func(address, user, catalog string) (DBMetadataConnection, error){
	sql.Postgres: postgresConnect,
	sql.MySQL:    mysqlConnect,
}

type columnMetadata struct {
	tableName    string
	columnName   string
	dataTypeName string
	dataTypeOid  uint32
}

type excludePattern struct {
	pattern *regexp.Regexp
	except  map[string]struct{}
}

type ColumnMetadataList struct {
	data       []*columnMetadata
	exclusions []*excludePattern
}

type DBMetadataConnection interface {
	Close(ctx context.Context) error
	DescribeSchema(ctx context.Context) (*ColumnMetadataList, error)
	DatabaseVersion(ctx context.Context) (string, error)
}

func (l *ColumnMetadataList) ForEachRow(addRow func(string, string, string, uint32)) {
	__antithesis_instrumentation__.Notify(40360)
	addRowIfAllowed := func(metadata *columnMetadata) {
		__antithesis_instrumentation__.Notify(40362)
		for _, exclusion := range l.exclusions {
			__antithesis_instrumentation__.Notify(40364)
			tableName := metadata.tableName
			if _, ok := exclusion.except[tableName]; exclusion.pattern.MatchString(tableName) && func() bool {
				__antithesis_instrumentation__.Notify(40365)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(40366)
				return
			} else {
				__antithesis_instrumentation__.Notify(40367)
			}
		}
		__antithesis_instrumentation__.Notify(40363)

		addRow(metadata.tableName, metadata.columnName, metadata.dataTypeName, metadata.dataTypeOid)
	}
	__antithesis_instrumentation__.Notify(40361)
	for _, metadata := range l.data {
		__antithesis_instrumentation__.Notify(40368)
		addRowIfAllowed(metadata)
	}
}
