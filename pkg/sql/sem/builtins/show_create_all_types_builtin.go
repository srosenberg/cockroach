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

func getTypeIDs(
	ctx context.Context,
	evalPlanner tree.EvalPlanner,
	txn *kv.Txn,
	dbName string,
	acc *mon.BoundAccount,
) ([]int64, error) {
	__antithesis_instrumentation__.Notify(602441)
	query := fmt.Sprintf(`
		SELECT descriptor_id
		FROM %s.crdb_internal.create_type_statements
		WHERE database_name = $1
		`, dbName)
	it, err := evalPlanner.QueryIteratorEx(
		ctx,
		"crdb_internal.show_create_all_types",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		dbName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(602445)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602446)
	}
	__antithesis_instrumentation__.Notify(602442)

	var typeIDs []int64

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(602447)
		tid := tree.MustBeDInt(it.Cur()[0])

		typeIDs = append(typeIDs, int64(tid))
		if err = acc.Grow(ctx, int64(unsafe.Sizeof(tid))); err != nil {
			__antithesis_instrumentation__.Notify(602448)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602449)
		}
	}
	__antithesis_instrumentation__.Notify(602443)
	if err != nil {
		__antithesis_instrumentation__.Notify(602450)
		return typeIDs, err
	} else {
		__antithesis_instrumentation__.Notify(602451)
	}
	__antithesis_instrumentation__.Notify(602444)

	return typeIDs, nil
}

func getTypeCreateStatement(
	ctx context.Context, evalPlanner tree.EvalPlanner, txn *kv.Txn, id int64, dbName string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602452)
	query := fmt.Sprintf(`
		SELECT
			create_statement
		FROM %s.crdb_internal.create_type_statements
		WHERE descriptor_id = $1
	`, dbName)
	row, err := evalPlanner.QueryRowEx(
		ctx,
		"crdb_internal.show_create_all_types",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		id,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(602454)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602455)
	}
	__antithesis_instrumentation__.Notify(602453)
	return row[0], nil
}
