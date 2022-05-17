package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type commentOnSchemaNode struct {
	n               *tree.CommentOnSchema
	schemaDesc      catalog.SchemaDescriptor
	metadataUpdater scexec.DescriptorMetadataUpdater
}

func (p *planner) CommentOnSchema(ctx context.Context, n *tree.CommentOnSchema) (planNode, error) {
	__antithesis_instrumentation__.Notify(457033)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"COMMENT ON SCHEMA",
	); err != nil {
		__antithesis_instrumentation__.Notify(457039)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457040)
	}
	__antithesis_instrumentation__.Notify(457034)

	dbName := p.CurrentDatabase()
	if dbName == "" {
		__antithesis_instrumentation__.Notify(457041)
		return nil, pgerror.New(pgcode.UndefinedDatabase,
			"cannot comment schema without being connected to a database")
	} else {
		__antithesis_instrumentation__.Notify(457042)
	}
	__antithesis_instrumentation__.Notify(457035)

	db, err := p.Descriptors().GetImmutableDatabaseByName(ctx, p.txn,
		dbName, tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(457043)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457044)
	}
	__antithesis_instrumentation__.Notify(457036)

	schemaDesc, err := p.Descriptors().GetImmutableSchemaByID(ctx, p.txn,
		db.GetSchemaID(string(n.Name)), tree.SchemaLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(457045)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457046)
	}
	__antithesis_instrumentation__.Notify(457037)

	if err := p.CheckPrivilege(ctx, db, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(457047)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(457048)
	}
	__antithesis_instrumentation__.Notify(457038)

	return &commentOnSchemaNode{
		n:          n,
		schemaDesc: schemaDesc,
		metadataUpdater: p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
			ctx,
			p.txn,
			p.SessionData(),
		),
	}, nil
}

func (n *commentOnSchemaNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(457049)
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(457051)
		err := n.metadataUpdater.UpsertDescriptorComment(
			int64(n.schemaDesc.GetID()), 0, keys.SchemaCommentType, *n.n.Comment)
		if err != nil {
			__antithesis_instrumentation__.Notify(457052)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457053)
		}
	} else {
		__antithesis_instrumentation__.Notify(457054)
		err := n.metadataUpdater.DeleteDescriptorComment(
			int64(n.schemaDesc.GetID()), 0, keys.SchemaCommentType)
		if err != nil {
			__antithesis_instrumentation__.Notify(457055)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457056)
		}
	}
	__antithesis_instrumentation__.Notify(457050)

	return nil
}

func (n *commentOnSchemaNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(457057)
	return false, nil
}
func (n *commentOnSchemaNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(457058)
	return tree.Datums{}
}
func (n *commentOnSchemaNode) Close(context.Context) { __antithesis_instrumentation__.Notify(457059) }
