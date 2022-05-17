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

type commentOnDatabaseNode struct {
	n               *tree.CommentOnDatabase
	dbDesc          catalog.DatabaseDescriptor
	metadataUpdater scexec.DescriptorMetadataUpdater
}

func (p *planner) CommentOnDatabase(
	ctx context.Context, n *tree.CommentOnDatabase,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(456985)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"COMMENT ON DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(456989)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456990)
	}
	__antithesis_instrumentation__.Notify(456986)

	dbDesc, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
		string(n.Name), tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(456991)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456992)
	}
	__antithesis_instrumentation__.Notify(456987)
	if err := p.CheckPrivilege(ctx, dbDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(456993)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456994)
	}
	__antithesis_instrumentation__.Notify(456988)

	return &commentOnDatabaseNode{n: n,
		dbDesc: dbDesc,
		metadataUpdater: p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
			ctx,
			p.txn,
			p.SessionData(),
		),
	}, nil
}

func (n *commentOnDatabaseNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(456995)
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(456998)
		err := n.metadataUpdater.UpsertDescriptorComment(
			int64(n.dbDesc.GetID()), 0, keys.DatabaseCommentType, *n.n.Comment)
		if err != nil {
			__antithesis_instrumentation__.Notify(456999)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457000)
		}
	} else {
		__antithesis_instrumentation__.Notify(457001)
		err := n.metadataUpdater.DeleteDescriptorComment(
			int64(n.dbDesc.GetID()), 0, keys.DatabaseCommentType)
		if err != nil {
			__antithesis_instrumentation__.Notify(457002)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457003)
		}
	}
	__antithesis_instrumentation__.Notify(456996)

	dbComment := ""
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(457004)
		dbComment = *n.n.Comment
	} else {
		__antithesis_instrumentation__.Notify(457005)
	}
	__antithesis_instrumentation__.Notify(456997)
	return params.p.logEvent(params.ctx,
		n.dbDesc.GetID(),
		&eventpb.CommentOnDatabase{
			DatabaseName: n.n.Name.String(),
			Comment:      dbComment,
			NullComment:  n.n.Comment == nil,
		})
}

func (n *commentOnDatabaseNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(457006)
	return false, nil
}
func (n *commentOnDatabaseNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(457007)
	return tree.Datums{}
}
func (n *commentOnDatabaseNode) Close(context.Context) { __antithesis_instrumentation__.Notify(457008) }
