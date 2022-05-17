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

var errEmptyColumnName = pgerror.New(pgcode.Syntax, "empty column name")

type renameColumnNode struct {
	n         *tree.RenameColumn
	tableDesc *tabledesc.Mutable
}

func (p *planner) RenameColumn(ctx context.Context, n *tree.RenameColumn) (planNode, error) {
	__antithesis_instrumentation__.Notify(565673)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"RENAME COLUMN",
	); err != nil {
		__antithesis_instrumentation__.Notify(565678)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565679)
	}
	__antithesis_instrumentation__.Notify(565674)

	_, tableDesc, err := p.ResolveMutableTableDescriptor(ctx, &n.Table, !n.IfExists, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(565680)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565681)
	}
	__antithesis_instrumentation__.Notify(565675)
	if tableDesc == nil {
		__antithesis_instrumentation__.Notify(565682)
		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(565683)
	}
	__antithesis_instrumentation__.Notify(565676)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(565684)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565685)
	}
	__antithesis_instrumentation__.Notify(565677)

	return &renameColumnNode{n: n, tableDesc: tableDesc}, nil
}

func (n *renameColumnNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(565686) }

func (n *renameColumnNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(565687)
	p := params.p
	ctx := params.ctx
	tableDesc := n.tableDesc

	descChanged, err := params.p.renameColumn(params.ctx, tableDesc, n.n.Name, n.n.NewName)
	if err != nil {
		__antithesis_instrumentation__.Notify(565691)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565692)
	}
	__antithesis_instrumentation__.Notify(565688)

	if !descChanged {
		__antithesis_instrumentation__.Notify(565693)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(565694)
	}
	__antithesis_instrumentation__.Notify(565689)

	if err := validateDescriptor(ctx, p, tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(565695)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565696)
	}
	__antithesis_instrumentation__.Notify(565690)

	return p.writeSchemaChange(
		ctx, tableDesc, descpb.InvalidMutationID, tree.AsStringWithFQNames(n.n, params.Ann()))
}

func (p *planner) findColumnToRename(
	ctx context.Context, tableDesc *tabledesc.Mutable, oldName, newName tree.Name,
) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(565697)
	if newName == "" {
		__antithesis_instrumentation__.Notify(565704)
		return nil, errEmptyColumnName
	} else {
		__antithesis_instrumentation__.Notify(565705)
	}
	__antithesis_instrumentation__.Notify(565698)

	col, err := tableDesc.FindColumnWithName(oldName)
	if err != nil {
		__antithesis_instrumentation__.Notify(565706)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565707)
	}
	__antithesis_instrumentation__.Notify(565699)

	for _, tableRef := range tableDesc.DependedOnBy {
		__antithesis_instrumentation__.Notify(565708)
		found := false
		for _, colID := range tableRef.ColumnIDs {
			__antithesis_instrumentation__.Notify(565710)
			if colID == col.GetID() {
				__antithesis_instrumentation__.Notify(565711)
				found = true
			} else {
				__antithesis_instrumentation__.Notify(565712)
			}
		}
		__antithesis_instrumentation__.Notify(565709)
		if found {
			__antithesis_instrumentation__.Notify(565713)
			return nil, p.dependentViewError(
				ctx, "column", oldName.String(), tableDesc.ParentID, tableRef.ID, "rename",
			)
		} else {
			__antithesis_instrumentation__.Notify(565714)
		}
	}
	__antithesis_instrumentation__.Notify(565700)
	if oldName == newName {
		__antithesis_instrumentation__.Notify(565715)

		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(565716)
	}
	__antithesis_instrumentation__.Notify(565701)

	if col.IsInaccessible() {
		__antithesis_instrumentation__.Notify(565717)
		return nil, pgerror.Newf(
			pgcode.UndefinedColumn,
			"column %q is inaccessible and cannot be renamed",
			col.GetName(),
		)
	} else {
		__antithesis_instrumentation__.Notify(565718)
	}
	__antithesis_instrumentation__.Notify(565702)

	_, err = checkColumnDoesNotExist(tableDesc, newName)
	if err != nil {
		__antithesis_instrumentation__.Notify(565719)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565720)
	}
	__antithesis_instrumentation__.Notify(565703)
	return col, nil
}

func (p *planner) renameColumn(
	ctx context.Context, tableDesc *tabledesc.Mutable, oldName, newName tree.Name,
) (changed bool, err error) {
	__antithesis_instrumentation__.Notify(565721)
	col, err := p.findColumnToRename(ctx, tableDesc, oldName, newName)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(565725)
		return col == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(565726)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(565727)
	}
	__antithesis_instrumentation__.Notify(565722)
	if tableDesc.IsShardColumn(col) {
		__antithesis_instrumentation__.Notify(565728)
		return false, pgerror.Newf(pgcode.ReservedName, "cannot rename shard column")
	} else {
		__antithesis_instrumentation__.Notify(565729)
	}
	__antithesis_instrumentation__.Notify(565723)
	if err := tabledesc.RenameColumnInTable(tableDesc, col, newName, func(shardCol catalog.Column, newShardColName tree.Name) (bool, error) {
		__antithesis_instrumentation__.Notify(565730)
		if c, err := p.findColumnToRename(ctx, tableDesc, shardCol.ColName(), newShardColName); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(565732)
			return c == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(565733)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(565734)
		}
		__antithesis_instrumentation__.Notify(565731)
		return true, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(565735)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(565736)
	}
	__antithesis_instrumentation__.Notify(565724)
	return true, nil
}

func (n *renameColumnNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565737)
	return false, nil
}
func (n *renameColumnNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565738)
	return tree.Datums{}
}
func (n *renameColumnNode) Close(context.Context) { __antithesis_instrumentation__.Notify(565739) }
