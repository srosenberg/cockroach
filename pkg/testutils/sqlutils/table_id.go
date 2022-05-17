package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"testing"
)

func QueryDatabaseID(t testing.TB, sqlDB DBHandle, dbName string) uint32 {
	__antithesis_instrumentation__.Notify(646363)
	dbIDQuery := `
		SELECT id FROM system.namespace
		WHERE name = $1 AND "parentSchemaID" = 0 AND "parentID" = 0
	`
	var dbID uint32
	result := sqlDB.QueryRowContext(context.Background(), dbIDQuery, dbName)
	if err := result.Scan(&dbID); err != nil {
		__antithesis_instrumentation__.Notify(646365)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646366)
	}
	__antithesis_instrumentation__.Notify(646364)
	return dbID
}

func QuerySchemaID(t testing.TB, sqlDB DBHandle, dbName, schemaName string) uint32 {
	__antithesis_instrumentation__.Notify(646367)
	tableIDQuery := `
 SELECT schemas.id FROM system.namespace schemas
   JOIN system.namespace dbs ON dbs.id = schemas."parentID"
   WHERE dbs.name = $1 AND schemas.name = $2
 `
	var schemaID uint32
	result := sqlDB.QueryRowContext(
		context.Background(),
		tableIDQuery, dbName,
		schemaName,
	)
	if err := result.Scan(&schemaID); err != nil {
		__antithesis_instrumentation__.Notify(646369)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646370)
	}
	__antithesis_instrumentation__.Notify(646368)
	return schemaID
}

func QueryTableID(
	t testing.TB, sqlDB DBHandle, dbName, schemaName string, tableName string,
) uint32 {
	__antithesis_instrumentation__.Notify(646371)
	tableIDQuery := `
 SELECT tables.id FROM system.namespace tables
   JOIN system.namespace dbs ON dbs.id = tables."parentID"
	 JOIN system.namespace schemas ON schemas.id = tables."parentSchemaID"
   WHERE dbs.name = $1 AND schemas.name = $2 AND tables.name = $3
 `
	var tableID uint32
	result := sqlDB.QueryRowContext(
		context.Background(),
		tableIDQuery, dbName,
		schemaName,
		tableName,
	)
	if err := result.Scan(&tableID); err != nil {
		__antithesis_instrumentation__.Notify(646373)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(646374)
	}
	__antithesis_instrumentation__.Notify(646372)
	return tableID
}
