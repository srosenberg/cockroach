package builtins

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

const mapEntryOverhead = 64

const alterAddFKStatements = "alter_statements"

const alterValidateFKStatements = "validate_statements"

const foreignKeyValidationWarning = "-- Validate foreign key constraints. These can fail if there was unvalidated data during the SHOW CREATE ALL TABLES"

func getTopologicallySortedTableIDs(
	ctx context.Context,
	evalPlanner tree.EvalPlanner,
	txn *kv.Txn,
	dbName string,
	acc *mon.BoundAccount,
) ([]int64, error) {
	__antithesis_instrumentation__.Notify(602373)
	ids, err := getTableIDs(ctx, evalPlanner, txn, dbName, acc)
	if err != nil {
		__antithesis_instrumentation__.Notify(602381)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602382)
	}
	__antithesis_instrumentation__.Notify(602374)

	if len(ids) == 0 {
		__antithesis_instrumentation__.Notify(602383)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(602384)
	}
	__antithesis_instrumentation__.Notify(602375)

	sizeOfMap := int64(0)

	dependsOnIDs := make(map[int64][]int64)
	for _, tid := range ids {
		__antithesis_instrumentation__.Notify(602385)
		query := fmt.Sprintf(`
		SELECT dependson_id
		FROM %s.crdb_internal.backward_dependencies
		WHERE descriptor_id = $1
		`, dbName)
		it, err := evalPlanner.QueryIteratorEx(
			ctx,
			"crdb_internal.show_create_all_tables",
			txn,
			sessiondata.NoSessionDataOverride,
			query,
			tid,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(602390)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602391)
		}
		__antithesis_instrumentation__.Notify(602386)

		var refs []int64
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(602392)
			id := tree.MustBeDInt(it.Cur()[0])
			refs = append(refs, int64(id))
		}
		__antithesis_instrumentation__.Notify(602387)
		if err != nil {
			__antithesis_instrumentation__.Notify(602393)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602394)
		}
		__antithesis_instrumentation__.Notify(602388)

		sizeOfKeyValue := int64(unsafe.Sizeof(tid)) + int64(len(refs))*memsize.Int64
		sizeOfMap += sizeOfKeyValue + mapEntryOverhead
		if err = acc.Grow(ctx, sizeOfKeyValue+mapEntryOverhead); err != nil {
			__antithesis_instrumentation__.Notify(602395)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602396)
		}
		__antithesis_instrumentation__.Notify(602389)

		dependsOnIDs[tid] = refs
	}
	__antithesis_instrumentation__.Notify(602376)

	sort.Slice(ids, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(602397)
		return ids[i] < ids[j]
	})
	__antithesis_instrumentation__.Notify(602377)

	var topologicallyOrderedIDs []int64

	if err = acc.Grow(ctx, int64(len(ids))*memsize.Int64); err != nil {
		__antithesis_instrumentation__.Notify(602398)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602399)
	}
	__antithesis_instrumentation__.Notify(602378)
	seen := make(map[int64]struct{})
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(602400)
		if err := topologicalSort(
			ctx, id, dependsOnIDs, seen, &topologicallyOrderedIDs, acc,
		); err != nil {
			__antithesis_instrumentation__.Notify(602401)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602402)
		}
	}
	__antithesis_instrumentation__.Notify(602379)

	sizeOfSeen := len(seen)
	seen = nil
	acc.Shrink(ctx, int64(sizeOfSeen)*(memsize.Int64+mapEntryOverhead))

	if len(ids) != len(topologicallyOrderedIDs) {
		__antithesis_instrumentation__.Notify(602403)
		return nil, errors.AssertionFailedf("show_create_all_tables_builtin failed. "+
			"len(ids):% d not equal to len(topologicallySortedIDs): %d",
			len(ids), len(topologicallyOrderedIDs))
	} else {
		__antithesis_instrumentation__.Notify(602404)
	}
	__antithesis_instrumentation__.Notify(602380)

	acc.Shrink(ctx, int64(len(ids))*memsize.Int64)
	acc.Shrink(ctx, sizeOfMap)
	return topologicallyOrderedIDs, nil
}

func getTableIDs(
	ctx context.Context,
	evalPlanner tree.EvalPlanner,
	txn *kv.Txn,
	dbName string,
	acc *mon.BoundAccount,
) ([]int64, error) {
	__antithesis_instrumentation__.Notify(602405)
	query := fmt.Sprintf(`
		SELECT descriptor_id
		FROM %s.crdb_internal.create_statements
		WHERE database_name = $1 
		AND is_virtual = FALSE
		AND is_temporary = FALSE
		`, dbName)
	it, err := evalPlanner.QueryIteratorEx(
		ctx,
		"crdb_internal.show_create_all_tables",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		dbName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(602409)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602410)
	}
	__antithesis_instrumentation__.Notify(602406)

	var tableIDs []int64

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(602411)
		tid := tree.MustBeDInt(it.Cur()[0])

		tableIDs = append(tableIDs, int64(tid))
		if err = acc.Grow(ctx, int64(unsafe.Sizeof(tid))); err != nil {
			__antithesis_instrumentation__.Notify(602412)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(602413)
		}
	}
	__antithesis_instrumentation__.Notify(602407)
	if err != nil {
		__antithesis_instrumentation__.Notify(602414)
		return tableIDs, err
	} else {
		__antithesis_instrumentation__.Notify(602415)
	}
	__antithesis_instrumentation__.Notify(602408)

	return tableIDs, nil
}

func topologicalSort(
	ctx context.Context,
	tid int64,
	dependsOnIDs map[int64][]int64,
	seen map[int64]struct{},
	collected *[]int64,
	acc *mon.BoundAccount,
) error {
	__antithesis_instrumentation__.Notify(602416)

	if _, isPresent := seen[tid]; isPresent {
		__antithesis_instrumentation__.Notify(602422)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(602423)
	}
	__antithesis_instrumentation__.Notify(602417)

	if _, exists := dependsOnIDs[tid]; !exists {
		__antithesis_instrumentation__.Notify(602424)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(602425)
	}
	__antithesis_instrumentation__.Notify(602418)

	if err := acc.Grow(ctx, memsize.Int64+mapEntryOverhead); err != nil {
		__antithesis_instrumentation__.Notify(602426)
		return err
	} else {
		__antithesis_instrumentation__.Notify(602427)
	}
	__antithesis_instrumentation__.Notify(602419)
	seen[tid] = struct{}{}
	for _, dep := range dependsOnIDs[tid] {
		__antithesis_instrumentation__.Notify(602428)
		if err := topologicalSort(ctx, dep, dependsOnIDs, seen, collected, acc); err != nil {
			__antithesis_instrumentation__.Notify(602429)
			return err
		} else {
			__antithesis_instrumentation__.Notify(602430)
		}
	}
	__antithesis_instrumentation__.Notify(602420)

	if err := acc.Grow(ctx, int64(unsafe.Sizeof(tid))); err != nil {
		__antithesis_instrumentation__.Notify(602431)
		return err
	} else {
		__antithesis_instrumentation__.Notify(602432)
	}
	__antithesis_instrumentation__.Notify(602421)
	*collected = append(*collected, tid)

	return nil
}

func getCreateStatement(
	ctx context.Context, evalPlanner tree.EvalPlanner, txn *kv.Txn, id int64, dbName string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602433)
	query := fmt.Sprintf(`
		SELECT
			create_nofks
		FROM %s.crdb_internal.create_statements
		WHERE descriptor_id = $1
	`, dbName)
	row, err := evalPlanner.QueryRowEx(
		ctx,
		"crdb_internal.show_create_all_tables",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		id,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(602435)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602436)
	}
	__antithesis_instrumentation__.Notify(602434)
	return row[0], nil
}

func getAlterStatements(
	ctx context.Context,
	evalPlanner tree.EvalPlanner,
	txn *kv.Txn,
	id int64,
	dbName string,
	statementType string,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(602437)
	query := fmt.Sprintf(`
		SELECT
			%s
		FROM %s.crdb_internal.create_statements
		WHERE descriptor_id = $1
	`, statementType, dbName)
	row, err := evalPlanner.QueryRowEx(
		ctx,
		"crdb_internal.show_create_all_tables",
		txn,
		sessiondata.NoSessionDataOverride,
		query,
		id,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(602439)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602440)
	}
	__antithesis_instrumentation__.Notify(602438)

	return row[0], nil
}
