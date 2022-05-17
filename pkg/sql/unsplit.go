package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

type unsplitNode struct {
	optColumnsSlot

	tableDesc catalog.TableDescriptor
	index     catalog.Index
	run       unsplitRun
	rows      planNode
}

type unsplitRun struct {
	lastUnsplitKey []byte
}

func (n *unsplitNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(631246)
	return nil
}

func (n *unsplitNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631247)
	if ok, err := n.rows.Next(params); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(631251)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(631252)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(631253)
	}
	__antithesis_instrumentation__.Notify(631248)

	row := n.rows.Values()
	rowKey, err := getRowKey(params.ExecCfg().Codec, n.tableDesc, n.index, row)
	if err != nil {
		__antithesis_instrumentation__.Notify(631254)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(631255)
	}
	__antithesis_instrumentation__.Notify(631249)

	if err := params.extendedEvalCtx.ExecCfg.DB.AdminUnsplit(params.ctx, rowKey); err != nil {
		__antithesis_instrumentation__.Notify(631256)
		ctx := params.p.EvalContext().FmtCtx(tree.FmtSimple)
		row.Format(ctx)
		return false, errors.Wrapf(err, "could not UNSPLIT AT %s", ctx)
	} else {
		__antithesis_instrumentation__.Notify(631257)
	}
	__antithesis_instrumentation__.Notify(631250)

	n.run.lastUnsplitKey = rowKey

	return true, nil
}

func (n *unsplitNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(631258)
	return tree.Datums{
		tree.NewDBytes(tree.DBytes(n.run.lastUnsplitKey)),
		tree.NewDString(keys.PrettyPrint(nil, n.run.lastUnsplitKey)),
	}
}

func (n *unsplitNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(631259)
	n.rows.Close(ctx)
}

type unsplitAllNode struct {
	optColumnsSlot

	tableDesc catalog.TableDescriptor
	index     catalog.Index
	run       unsplitAllRun
}

type unsplitAllRun struct {
	keys           [][]byte
	lastUnsplitKey []byte
}

func (n *unsplitAllNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(631260)

	statement := `
		SELECT
			start_key
		FROM
			crdb_internal.ranges_no_leases
		WHERE
			database_name=$1 AND table_name=$2 AND index_name=$3 AND split_enforced_until IS NOT NULL
	`
	dbDesc, err := params.p.Descriptors().Direct().MustGetDatabaseDescByID(
		params.ctx, params.p.txn, n.tableDesc.GetParentID(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(631265)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631266)
	}
	__antithesis_instrumentation__.Notify(631261)
	indexName := ""
	if n.index.GetID() != n.tableDesc.GetPrimaryIndexID() {
		__antithesis_instrumentation__.Notify(631267)
		indexName = n.index.GetName()
	} else {
		__antithesis_instrumentation__.Notify(631268)
	}
	__antithesis_instrumentation__.Notify(631262)
	ie := params.p.ExecCfg().InternalExecutorFactory(params.ctx, params.SessionData())
	it, err := ie.QueryIteratorEx(
		params.ctx, "split points query", params.p.txn, sessiondata.NoSessionDataOverride,
		statement,
		dbDesc.GetName(),
		n.tableDesc.GetName(),
		indexName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(631269)
		return err
	} else {
		__antithesis_instrumentation__.Notify(631270)
	}
	__antithesis_instrumentation__.Notify(631263)
	var ok bool
	for ok, err = it.Next(params.ctx); ok; ok, err = it.Next(params.ctx) {
		__antithesis_instrumentation__.Notify(631271)
		n.run.keys = append(n.run.keys, []byte(*(it.Cur()[0].(*tree.DBytes))))
	}
	__antithesis_instrumentation__.Notify(631264)
	return err
}

func (n *unsplitAllNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(631272)
	if len(n.run.keys) == 0 {
		__antithesis_instrumentation__.Notify(631275)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(631276)
	}
	__antithesis_instrumentation__.Notify(631273)
	rowKey := n.run.keys[0]
	n.run.keys = n.run.keys[1:]

	if err := params.extendedEvalCtx.ExecCfg.DB.AdminUnsplit(params.ctx, rowKey); err != nil {
		__antithesis_instrumentation__.Notify(631277)
		return false, errors.Wrapf(err, "could not UNSPLIT AT %s", keys.PrettyPrint(nil, rowKey))
	} else {
		__antithesis_instrumentation__.Notify(631278)
	}
	__antithesis_instrumentation__.Notify(631274)

	n.run.lastUnsplitKey = rowKey

	return true, nil
}

func (n *unsplitAllNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(631279)
	return tree.Datums{
		tree.NewDBytes(tree.DBytes(n.run.lastUnsplitKey)),
		tree.NewDString(keys.PrettyPrint(nil, n.run.lastUnsplitKey)),
	}
}

func (n *unsplitAllNode) Close(ctx context.Context) { __antithesis_instrumentation__.Notify(631280) }
