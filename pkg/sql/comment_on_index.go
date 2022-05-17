package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type commentOnIndexNode struct {
	n               *tree.CommentOnIndex
	tableDesc       *tabledesc.Mutable
	index           catalog.Index
	metadataUpdater scexec.DescriptorMetadataUpdater
}

func (p *planner) CommentOnIndex(ctx context.Context, n *tree.CommentOnIndex) (planNode, error) {
	__antithesis_instrumentation__.Notify(457009)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"COMMENT ON INDEX",
	); err != nil {
		__antithesis_instrumentation__.Notify(457012)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457013)
	}
	__antithesis_instrumentation__.Notify(457010)

	tableDesc, index, err := p.getTableAndIndex(ctx, &n.Index, privilege.CREATE)
	if err != nil {
		__antithesis_instrumentation__.Notify(457014)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457015)
	}
	__antithesis_instrumentation__.Notify(457011)

	return &commentOnIndexNode{
		n:         n,
		tableDesc: tableDesc,
		index:     index,
		metadataUpdater: p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
			ctx,
			p.txn,
			p.SessionData(),
		)}, nil
}

func (n *commentOnIndexNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(457016)
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(457020)
		err := n.metadataUpdater.UpsertDescriptorComment(
			int64(n.tableDesc.ID),
			int64(n.index.GetID()),
			keys.IndexCommentType,
			*n.n.Comment,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(457021)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457022)
		}
	} else {
		__antithesis_instrumentation__.Notify(457023)
		err := n.metadataUpdater.DeleteDescriptorComment(
			int64(n.tableDesc.ID), int64(n.index.GetID()), keys.IndexCommentType)
		if err != nil {
			__antithesis_instrumentation__.Notify(457024)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457025)
		}
	}
	__antithesis_instrumentation__.Notify(457017)

	comment := ""
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(457026)
		comment = *n.n.Comment
	} else {
		__antithesis_instrumentation__.Notify(457027)
	}
	__antithesis_instrumentation__.Notify(457018)

	tn, err := params.p.getQualifiedTableName(params.ctx, n.tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(457028)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457029)
	}
	__antithesis_instrumentation__.Notify(457019)

	return params.p.logEvent(params.ctx,
		n.tableDesc.ID,
		&eventpb.CommentOnIndex{
			TableName:   tn.FQString(),
			IndexName:   string(n.n.Index.Index),
			Comment:     comment,
			NullComment: n.n.Comment == nil,
		})
}

func (n *commentOnIndexNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(457030)
	return false, nil
}
func (n *commentOnIndexNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(457031)
	return tree.Datums{}
}
func (n *commentOnIndexNode) Close(context.Context) { __antithesis_instrumentation__.Notify(457032) }
