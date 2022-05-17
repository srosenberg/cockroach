package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

func getSchemaIDs(
	ctx context.Context,
	evalPlanner tree.EvalPlanner,
	txn *kv.Txn,
	dbName string,
	acc *mon.BoundAccount,
) ([]int64, error) {
	__antithesis_instrumentation__.Notify(602358)
	query := fmt.Sprintf(`
		SELECT descriptor_id
		FROM %s.crdb_internal.create_schema_statements
		WHERE database_name = $1
		`, dbName)
	it, err := evalPlanner.QueryIteratorEx(
		ctx,
		"crdb_internal.show_create_all_schemas",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		dbName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(602362)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602363)
	}
	__antithesis_instrumentation__.Notify(602359)

	var schemaIDs []int64

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(602364)
		tid := tree.MustBeDInt(it.Cur()[0])

		schemaIDs = append(schemaIDs, int64(tid))
		if err = acc.Grow(ctx, int64(unsafe.Sizeof(tid))); err != nil {
			__antithesis_instrumentation__.Notify(602365)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602366)
		}
	}
	__antithesis_instrumentation__.Notify(602360)
	if err != nil {
		__antithesis_instrumentation__.Notify(602367)
		return schemaIDs, err
	} else {
		__antithesis_instrumentation__.Notify(602368)
	}
	__antithesis_instrumentation__.Notify(602361)

	return schemaIDs, nil
}

func getSchemaCreateStatement(
	ctx context.Context, evalPlanner tree.EvalPlanner, txn *kv.Txn, id int64, dbName string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602369)
	query := fmt.Sprintf(`
		SELECT
			create_statement
		FROM %s.crdb_internal.create_schema_statements
		WHERE descriptor_id = $1
	`, dbName)
	row, err := evalPlanner.QueryRowEx(
		ctx,
		"crdb_internal.show_create_all_schemas",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		id,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(602371)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602372)
	}
	__antithesis_instrumentation__.Notify(602370)
	return row[0], nil
}
