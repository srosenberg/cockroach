package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type splitNode struct {
	optColumnsSlot

	tableDesc      catalog.TableDescriptor
	index          catalog.Index
	rows           planNode
	run            splitRun
	expirationTime hlc.Timestamp
}

type splitRun struct {
	lastSplitKey       []byte
	lastExpirationTime hlc.Timestamp
}

func (n *splitNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623672)
	return nil
}

func (n *splitNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623673)

	if ok, err := n.rows.Next(params); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(623677)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(623678)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(623679)
	}
	__antithesis_instrumentation__.Notify(623674)

	rowKey, err := getRowKey(params.ExecCfg().Codec, n.tableDesc, n.index, n.rows.Values())
	if err != nil {
		__antithesis_instrumentation__.Notify(623680)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623681)
	}
	__antithesis_instrumentation__.Notify(623675)

	if err := params.ExecCfg().DB.AdminSplit(params.ctx, rowKey, n.expirationTime); err != nil {
		__antithesis_instrumentation__.Notify(623682)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623683)
	}
	__antithesis_instrumentation__.Notify(623676)

	n.run.lastSplitKey = rowKey
	n.run.lastExpirationTime = n.expirationTime

	return true, nil
}

func (n *splitNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623684)
	splitEnforcedUntil := tree.DNull
	if !n.run.lastExpirationTime.IsEmpty() {
		__antithesis_instrumentation__.Notify(623686)
		splitEnforcedUntil = tree.TimestampToInexactDTimestamp(n.run.lastExpirationTime)
	} else {
		__antithesis_instrumentation__.Notify(623687)
	}
	__antithesis_instrumentation__.Notify(623685)
	return tree.Datums{
		tree.NewDBytes(tree.DBytes(n.run.lastSplitKey)),
		tree.NewDString(catalogkeys.PrettyKey(nil, n.run.lastSplitKey, 2)),
		splitEnforcedUntil,
	}
}

func (n *splitNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623688)
	n.rows.Close(ctx)
}

func getRowKey(
	codec keys.SQLCodec, tableDesc catalog.TableDescriptor, index catalog.Index, values []tree.Datum,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(623689)
	if index.NumKeyColumns() < len(values) {
		__antithesis_instrumentation__.Notify(623693)
		return nil, pgerror.Newf(pgcode.Syntax, "excessive number of values provided: expected %d, got %d", index.NumKeyColumns(), len(values))
	} else {
		__antithesis_instrumentation__.Notify(623694)
	}
	__antithesis_instrumentation__.Notify(623690)
	var colMap catalog.TableColMap
	for i := range values {
		__antithesis_instrumentation__.Notify(623695)
		colMap.Set(index.GetKeyColumnID(i), i)
	}
	__antithesis_instrumentation__.Notify(623691)
	prefix := rowenc.MakeIndexKeyPrefix(codec, tableDesc.GetID(), index.GetID())
	keyCols := tableDesc.IndexFetchSpecKeyAndSuffixColumns(index)
	key, _, err := rowenc.EncodePartialIndexKey(keyCols[:len(values)], colMap, values, prefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(623696)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623697)
	}
	__antithesis_instrumentation__.Notify(623692)
	return key, nil
}

func parseExpirationTime(
	evalCtx *tree.EvalContext, expireExpr tree.TypedExpr,
) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(623698)
	if !tree.IsConst(evalCtx, expireExpr) {
		__antithesis_instrumentation__.Notify(623704)
		return hlc.Timestamp{}, errors.Errorf("SPLIT AT: only constant expressions are allowed for expiration")
	} else {
		__antithesis_instrumentation__.Notify(623705)
	}
	__antithesis_instrumentation__.Notify(623699)
	d, err := expireExpr.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(623706)
		return hlc.Timestamp{}, err
	} else {
		__antithesis_instrumentation__.Notify(623707)
	}
	__antithesis_instrumentation__.Notify(623700)
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(623708)
		return hlc.MaxTimestamp, nil
	} else {
		__antithesis_instrumentation__.Notify(623709)
	}
	__antithesis_instrumentation__.Notify(623701)
	stmtTimestamp := evalCtx.GetStmtTimestamp()
	ts, err := tree.DatumToHLC(evalCtx, stmtTimestamp, d)
	if err != nil {
		__antithesis_instrumentation__.Notify(623710)
		return ts, errors.Wrap(err, "SPLIT AT")
	} else {
		__antithesis_instrumentation__.Notify(623711)
	}
	__antithesis_instrumentation__.Notify(623702)
	if ts.GoTime().Before(stmtTimestamp) {
		__antithesis_instrumentation__.Notify(623712)
		return ts, errors.Errorf("SPLIT AT: expiration time should be greater than or equal to current time")
	} else {
		__antithesis_instrumentation__.Notify(623713)
	}
	__antithesis_instrumentation__.Notify(623703)
	return ts, nil
}
