package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type commentOnTableNode struct {
	n               *tree.CommentOnTable
	tableDesc       catalog.TableDescriptor
	metadataUpdater scexec.DescriptorMetadataUpdater
}

func (p *planner) CommentOnTable(ctx context.Context, n *tree.CommentOnTable) (planNode, error) {
	__antithesis_instrumentation__.Notify(457060)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"COMMENT ON TABLE",
	); err != nil {
		__antithesis_instrumentation__.Notify(457064)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457065)
	}
	__antithesis_instrumentation__.Notify(457061)

	tableDesc, err := p.ResolveUncachedTableDescriptorEx(ctx, n.Table, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(457066)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457067)
	}
	__antithesis_instrumentation__.Notify(457062)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(457068)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457069)
	}
	__antithesis_instrumentation__.Notify(457063)

	return &commentOnTableNode{
		n:         n,
		tableDesc: tableDesc,
		metadataUpdater: p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
			ctx,
			p.txn,
			p.SessionData(),
		),
	}, nil
}

func (n *commentOnTableNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(457070)
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(457073)
		err := n.metadataUpdater.UpsertDescriptorComment(
			int64(n.tableDesc.GetID()), 0, keys.TableCommentType, *n.n.Comment)
		if err != nil {
			__antithesis_instrumentation__.Notify(457074)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457075)
		}
	} else {
		__antithesis_instrumentation__.Notify(457076)
		err := n.metadataUpdater.DeleteDescriptorComment(
			int64(n.tableDesc.GetID()), 0, keys.TableCommentType)
		if err != nil {
			__antithesis_instrumentation__.Notify(457077)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457078)
		}
	}
	__antithesis_instrumentation__.Notify(457071)

	comment := ""
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(457079)
		comment = *n.n.Comment
	} else {
		__antithesis_instrumentation__.Notify(457080)
	}
	__antithesis_instrumentation__.Notify(457072)
	return params.p.logEvent(params.ctx,
		n.tableDesc.GetID(),
		&eventpb.CommentOnTable{
			TableName:   params.p.ResolvedName(n.n.Table).FQString(),
			Comment:     comment,
			NullComment: n.n.Comment == nil,
		})
}

func (n *commentOnTableNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(457081)
	return false, nil
}
func (n *commentOnTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(457082)
	return tree.Datums{}
}
func (n *commentOnTableNode) Close(context.Context) { __antithesis_instrumentation__.Notify(457083) }
