package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type commentOnColumnNode struct {
	n         *tree.CommentOnColumn
	tableDesc catalog.TableDescriptor
}

func (p *planner) CommentOnColumn(ctx context.Context, n *tree.CommentOnColumn) (planNode, error) {
	__antithesis_instrumentation__.Notify(456919)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"COMMENT ON COLUMN",
	); err != nil {
		__antithesis_instrumentation__.Notify(456924)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456925)
	}
	__antithesis_instrumentation__.Notify(456920)

	var tableName tree.TableName
	if n.ColumnItem.TableName != nil {
		__antithesis_instrumentation__.Notify(456926)
		tableName = n.ColumnItem.TableName.ToTableName()
	} else {
		__antithesis_instrumentation__.Notify(456927)
	}
	__antithesis_instrumentation__.Notify(456921)
	tableDesc, err := p.resolveUncachedTableDescriptor(ctx, &tableName, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(456928)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456929)
	}
	__antithesis_instrumentation__.Notify(456922)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(456930)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456931)
	}
	__antithesis_instrumentation__.Notify(456923)

	return &commentOnColumnNode{n: n, tableDesc: tableDesc}, nil
}

func (n *commentOnColumnNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(456932)
	col, err := n.tableDesc.FindColumnWithName(n.n.ColumnItem.ColumnName)
	if err != nil {
		__antithesis_instrumentation__.Notify(456937)
		return err
	} else {
		__antithesis_instrumentation__.Notify(456938)
	}
	__antithesis_instrumentation__.Notify(456933)

	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(456939)
		_, err := params.p.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
			params.ctx,
			"set-column-comment",
			params.p.Txn(),
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			"UPSERT INTO system.comments VALUES ($1, $2, $3, $4)",
			keys.ColumnCommentType,
			n.tableDesc.GetID(),
			col.GetPGAttributeNum(),
			*n.n.Comment)
		if err != nil {
			__antithesis_instrumentation__.Notify(456940)
			return err
		} else {
			__antithesis_instrumentation__.Notify(456941)
		}
	} else {
		__antithesis_instrumentation__.Notify(456942)
		_, err := params.p.extendedEvalCtx.ExecCfg.InternalExecutor.ExecEx(
			params.ctx,
			"delete-column-comment",
			params.p.Txn(),
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			"DELETE FROM system.comments WHERE type=$1 AND object_id=$2 AND sub_id=$3",
			keys.ColumnCommentType,
			n.tableDesc.GetID(),
			col.GetPGAttributeNum())
		if err != nil {
			__antithesis_instrumentation__.Notify(456943)
			return err
		} else {
			__antithesis_instrumentation__.Notify(456944)
		}
	}
	__antithesis_instrumentation__.Notify(456934)

	comment := ""
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(456945)
		comment = *n.n.Comment
	} else {
		__antithesis_instrumentation__.Notify(456946)
	}
	__antithesis_instrumentation__.Notify(456935)

	tn, err := params.p.getQualifiedTableName(params.ctx, n.tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(456947)
		return err
	} else {
		__antithesis_instrumentation__.Notify(456948)
	}
	__antithesis_instrumentation__.Notify(456936)

	return params.p.logEvent(params.ctx,
		n.tableDesc.GetID(),
		&eventpb.CommentOnColumn{
			TableName:   tn.FQString(),
			ColumnName:  string(n.n.ColumnItem.ColumnName),
			Comment:     comment,
			NullComment: n.n.Comment == nil,
		})
}

func (n *commentOnColumnNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(456949)
	return false, nil
}
func (n *commentOnColumnNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(456950)
	return tree.Datums{}
}
func (n *commentOnColumnNode) Close(context.Context) { __antithesis_instrumentation__.Notify(456951) }
