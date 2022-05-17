package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
)

type commentOnConstraintNode struct {
	n               *tree.CommentOnConstraint
	tableDesc       catalog.TableDescriptor
	metadataUpdater scexec.DescriptorMetadataUpdater
}

func (p *planner) CommentOnConstraint(
	ctx context.Context, n *tree.CommentOnConstraint,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(456952)

	if !p.ExecCfg().Settings.Version.IsActive(ctx, tabledesc.ConstraintIDsAddedToTableDescsVersion) {
		__antithesis_instrumentation__.Notify(456957)
		return nil, pgerror.Newf(pgcode.FeatureNotSupported, "cannot comment on constraint")
	} else {
		__antithesis_instrumentation__.Notify(456958)
	}
	__antithesis_instrumentation__.Notify(456953)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"COMMENT ON CONSTRAINT",
	); err != nil {
		__antithesis_instrumentation__.Notify(456959)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456960)
	}
	__antithesis_instrumentation__.Notify(456954)

	tableDesc, err := p.ResolveUncachedTableDescriptorEx(ctx, n.Table, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(456961)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456962)
	}
	__antithesis_instrumentation__.Notify(456955)
	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(456963)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(456964)
	}
	__antithesis_instrumentation__.Notify(456956)

	return &commentOnConstraintNode{
		n:         n,
		tableDesc: tableDesc,
		metadataUpdater: p.execCfg.DescMetadaUpdaterFactory.NewMetadataUpdater(
			ctx,
			p.txn,
			p.SessionData(),
		),
	}, nil

}

func (n *commentOnConstraintNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(456965)
	info, err := n.tableDesc.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(456970)
		return err
	} else {
		__antithesis_instrumentation__.Notify(456971)
	}
	__antithesis_instrumentation__.Notify(456966)

	constraintName := string(n.n.Constraint)
	constraint, ok := info[constraintName]
	if !ok {
		__antithesis_instrumentation__.Notify(456972)
		return pgerror.Newf(pgcode.UndefinedObject,
			"constraint %q of relation %q does not exist", constraintName, n.tableDesc.GetName())
	} else {
		__antithesis_instrumentation__.Notify(456973)
	}
	__antithesis_instrumentation__.Notify(456967)

	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(456974)
		err := n.metadataUpdater.UpsertConstraintComment(
			n.tableDesc.GetID(),
			constraint.ConstraintID,
			*n.n.Comment,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(456975)
			return err
		} else {
			__antithesis_instrumentation__.Notify(456976)
		}
	} else {
		__antithesis_instrumentation__.Notify(456977)
		err := n.metadataUpdater.DeleteConstraintComment(
			n.tableDesc.GetID(),
			constraint.ConstraintID,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(456978)
			return err
		} else {
			__antithesis_instrumentation__.Notify(456979)
		}
	}
	__antithesis_instrumentation__.Notify(456968)

	comment := ""
	if n.n.Comment != nil {
		__antithesis_instrumentation__.Notify(456980)
		comment = *n.n.Comment
	} else {
		__antithesis_instrumentation__.Notify(456981)
	}
	__antithesis_instrumentation__.Notify(456969)

	return params.p.logEvent(params.ctx,
		n.tableDesc.GetID(),
		&eventpb.CommentOnConstraint{
			TableName:      params.p.ResolvedName(n.n.Table).FQString(),
			ConstraintName: n.n.Constraint.String(),
			Comment:        comment,
			NullComment:    n.n.Comment == nil,
		})
}

func (n *commentOnConstraintNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(456982)
	return false, nil
}
func (n *commentOnConstraintNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(456983)
	return tree.Datums{}
}
func (n *commentOnConstraintNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(456984)
}
