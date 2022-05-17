package rdbms

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq/oid"
)

const getServerVersion = `SELECT current_setting('server_version');`

var unimplementedEquivalencies = map[oid.Oid]oid.Oid{

	oid.Oid(12653): oid.T_int8,
	oid.Oid(12665): oid.T_text,
	oid.Oid(12656): oid.T_text,
	oid.Oid(12658): oid.T_text,
	oid.Oid(12663): oid.T_timestamptz,

	oid.Oid(2277): oid.T__text,
	oid.Oid(3361): oid.T_bytea,
	oid.Oid(3402): oid.T_bytea,
	oid.Oid(5017): oid.T_bytea,

	oid.T__aclitem:     oid.T__text,
	oid.T_pg_node_tree: oid.T_text,
	oid.T_xid:          oid.T_int8,
	oid.T_pg_lsn:       oid.T_text,
}

var postgresExclusions = []*excludePattern{
	{
		pattern: regexp.MustCompile(`^_pg_.+$`),
		except:  make(map[string]struct{}),
	},
}

type pgMetadataConnection struct {
	*pgx.Conn
	catalog string
}

func postgresConnect(address, user, catalog string) (DBMetadataConnection, error) {
	__antithesis_instrumentation__.Notify(40384)
	conf, err := pgx.ParseConfig(fmt.Sprintf("postgresql://%s@%s?sslmode=disable", user, address))
	if err != nil {
		__antithesis_instrumentation__.Notify(40387)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40388)
	}
	__antithesis_instrumentation__.Notify(40385)
	conn, err := pgx.ConnectConfig(context.Background(), conf)
	if err != nil {
		__antithesis_instrumentation__.Notify(40389)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40390)
	}
	__antithesis_instrumentation__.Notify(40386)

	return pgMetadataConnection{conn, catalog}, nil
}

func (conn pgMetadataConnection) DescribeSchema(ctx context.Context) (*ColumnMetadataList, error) {
	__antithesis_instrumentation__.Notify(40391)
	var metadata []*columnMetadata
	rows, err := conn.Query(ctx, sql.GetPGMetadataSQL, conn.catalog)
	if err != nil {
		__antithesis_instrumentation__.Notify(40394)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40395)
	}
	__antithesis_instrumentation__.Notify(40392)
	for rows.Next() {
		__antithesis_instrumentation__.Notify(40396)
		var table, column, dataTypeName string
		var dataTypeOid uint32
		if err := rows.Scan(&table, &column, &dataTypeName, &dataTypeOid); err != nil {
			__antithesis_instrumentation__.Notify(40399)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(40400)
		}
		__antithesis_instrumentation__.Notify(40397)
		mappedTypeName, mappedTypeOid, err := getMappedType(dataTypeName, dataTypeOid)
		if err != nil {
			__antithesis_instrumentation__.Notify(40401)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(40402)
		}
		__antithesis_instrumentation__.Notify(40398)
		row := new(columnMetadata)
		row.tableName = table
		row.columnName = column
		row.dataTypeName = mappedTypeName
		row.dataTypeOid = mappedTypeOid
		metadata = append(metadata, row)
	}
	__antithesis_instrumentation__.Notify(40393)
	return &ColumnMetadataList{data: metadata, exclusions: postgresExclusions}, nil
}

func (conn pgMetadataConnection) DatabaseVersion(
	ctx context.Context,
) (pgVersion string, err error) {
	__antithesis_instrumentation__.Notify(40403)
	row := conn.QueryRow(ctx, getServerVersion)
	err = row.Scan(&pgVersion)
	return pgVersion, err
}

func getMappedType(dataTypeName string, dataTypeOid uint32) (string, uint32, error) {
	__antithesis_instrumentation__.Notify(40404)
	actualOid := oid.Oid(dataTypeOid)
	mappedOid, ok := unimplementedEquivalencies[actualOid]
	if !ok {
		__antithesis_instrumentation__.Notify(40408)

		return dataTypeName, dataTypeOid, nil
	} else {
		__antithesis_instrumentation__.Notify(40409)
	}
	__antithesis_instrumentation__.Notify(40405)

	_, ok = types.OidToType[mappedOid]
	if !ok {
		__antithesis_instrumentation__.Notify(40410)

		return "", 0, fmt.Errorf("type with oid %d is unimplemented", mappedOid)
	} else {
		__antithesis_instrumentation__.Notify(40411)
	}
	__antithesis_instrumentation__.Notify(40406)

	typeName, ok := oid.TypeName[mappedOid]
	if !ok {
		__antithesis_instrumentation__.Notify(40412)

		return "", 0, fmt.Errorf("type name for oid %d does not exist in oid.TypeName map", mappedOid)
	} else {
		__antithesis_instrumentation__.Notify(40413)
	}
	__antithesis_instrumentation__.Notify(40407)

	return typeName, uint32(mappedOid), nil
}
