package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var errEmptyIndexName = pgerror.New(pgcode.Syntax, "empty index name")

type renameIndexNode struct {
	n         *tree.RenameIndex
	tableDesc *tabledesc.Mutable
	idx       catalog.Index
}

func (p *planner) RenameIndex(ctx context.Context, n *tree.RenameIndex) (planNode, error) {
	__antithesis_instrumentation__.Notify(565862)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"RENAME INDEX",
	); err != nil {
		__antithesis_instrumentation__.Notify(565868)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565869)
	}
	__antithesis_instrumentation__.Notify(565863)

	_, tableDesc, err := expandMutableIndexName(ctx, p, n.Index, !n.IfExists)
	if err != nil {
		__antithesis_instrumentation__.Notify(565870)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565871)
	}
	__antithesis_instrumentation__.Notify(565864)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(565872)

		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(565873)
	}
	__antithesis_instrumentation__.Notify(565865)

	idx, err := tableDesc.FindIndexWithName(string(n.Index.Index))
	if err != nil {
		__antithesis_instrumentation__.Notify(565874)
		if n.IfExists {
			__antithesis_instrumentation__.Notify(565876)

			return newZeroNode(nil), nil
		} else {
			__antithesis_instrumentation__.Notify(565877)
		}
		__antithesis_instrumentation__.Notify(565875)

		return nil, pgerror.WithCandidateCode(err, pgcode.UndefinedObject)
	} else {
		__antithesis_instrumentation__.Notify(565878)
	}
	__antithesis_instrumentation__.Notify(565866)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(565879)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565880)
	}
	__antithesis_instrumentation__.Notify(565867)

	return &renameIndexNode{n: n, idx: idx, tableDesc: tableDesc}, nil
}

func (n *renameIndexNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(565881) }

func (n *renameIndexNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(565882)
	p := params.p
	ctx := params.ctx
	tableDesc := n.tableDesc
	idx := n.idx

	for _, tableRef := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(565888)
		if tableRef.IndexID != idx.GetID() {
			__antithesis_instrumentation__.Notify(565890)
			continue
		} else {
			__antithesis_instrumentation__.Notify(565891)
		}
		__antithesis_instrumentation__.Notify(565889)
		return p.dependentViewError(
			ctx, "index", n.n.Index.Index.String(), tableDesc.ParentID, tableRef.ID, "rename",
		)
	}
	__antithesis_instrumentation__.Notify(565883)

	if n.n.NewName == "" {
		__antithesis_instrumentation__.Notify(565892)
		return errEmptyIndexName
	} else {
		__antithesis_instrumentation__.Notify(565893)
	}
	__antithesis_instrumentation__.Notify(565884)

	if n.n.Index.Index == n.n.NewName {
		__antithesis_instrumentation__.Notify(565894)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(565895)
	}
	__antithesis_instrumentation__.Notify(565885)

	if foundIndex, _ := tableDesc.FindIndexWithName(string(n.n.NewName)); foundIndex != nil {
		__antithesis_instrumentation__.Notify(565896)
		return pgerror.Newf(pgcode.DuplicateRelation, "index name %q already exists", string(n.n.NewName))
	} else {
		__antithesis_instrumentation__.Notify(565897)
	}
	__antithesis_instrumentation__.Notify(565886)

	idx.IndexDesc().Name = string(n.n.NewName)

	if err := validateDescriptor(ctx, p, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(565898)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565899)
	}
	__antithesis_instrumentation__.Notify(565887)

	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()))
}

func (n *renameIndexNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565900)
	return false, nil
}
func (n *renameIndexNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565901)
	return tree.Datums{}
}
func (n *renameIndexNode) Close(context.Context) { __antithesis_instrumentation__.Notify(565902) }
