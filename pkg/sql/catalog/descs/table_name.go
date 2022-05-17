package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func GetTableNameByID(
	ctx context.Context, txn *kv.Txn, tc *Collection, tableID descpb.ID,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(264968)
	tbl, err := tc.GetImmutableTableByID(ctx, txn, tableID, tree.ObjectLookupFlagsWithRequired())
	if err != nil {
		__antithesis_instrumentation__.Notify(264970)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264971)
	}
	__antithesis_instrumentation__.Notify(264969)
	return GetTableNameByDesc(ctx, txn, tc, tbl)
}

func GetTableNameByDesc(
	ctx context.Context, txn *kv.Txn, tc *Collection, tbl catalog.TableDescriptor,
) (*tree.TableName, error) {
	__antithesis_instrumentation__.Notify(264972)
	sc, err := tc.GetImmutableSchemaByID(ctx, txn, tbl.GetParentSchemaID(), tree.SchemaLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(264976)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264977)
	}
	__antithesis_instrumentation__.Notify(264973)
	found, db, err := tc.GetImmutableDatabaseByID(ctx, txn, tbl.GetParentID(), tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(264978)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(264979)
	}
	__antithesis_instrumentation__.Notify(264974)
	if !found {
		__antithesis_instrumentation__.Notify(264980)
		return nil, errors.AssertionFailedf("expected database %d to exist", tbl.GetParentID())
	} else {
		__antithesis_instrumentation__.Notify(264981)
	}
	__antithesis_instrumentation__.Notify(264975)
	return tree.NewTableNameWithSchema(tree.Name(db.GetName()), tree.Name(sc.GetName()), tree.Name(tbl.GetName())), nil
}
