package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type saveTableNode struct {
	source planNode

	target tree.TableName

	colNames []string

	run struct {
		vals tree.ValuesClause
	}
}

const saveTableInsertBatch = 100

func (p *planner) makeSaveTable(
	source planNode, target *tree.TableName, colNames []string,
) planNode {
	__antithesis_instrumentation__.Notify(576343)
	return &saveTableNode{source: source, target: *target, colNames: colNames}
}

func (n *saveTableNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(576344)
	create := &tree.CreateTable{
		Table: n.target,
	}

	cols := planColumns(n.source)
	if len(n.colNames) != len(cols) {
		__antithesis_instrumentation__.Notify(576347)
		return errors.AssertionFailedf(
			"number of column names (%d) does not match number of columns (%d)",
			len(n.colNames), len(cols),
		)
	} else {
		__antithesis_instrumentation__.Notify(576348)
	}
	__antithesis_instrumentation__.Notify(576345)
	for i := 0; i < len(cols); i++ {
		__antithesis_instrumentation__.Notify(576349)
		def := &tree.ColumnTableDef{
			Name: tree.Name(n.colNames[i]),
			Type: cols[i].Typ,
		}
		def.Nullable.Nullability = tree.SilentNull
		create.Defs = append(create.Defs, def)
	}
	__antithesis_instrumentation__.Notify(576346)

	_, err := params.p.ExtendedEvalContext().ExecCfg.InternalExecutor.Exec(
		params.ctx,
		"create save table",
		nil,
		create.String(),
	)
	return err
}

func (n *saveTableNode) issue(params runParams) error {
	__antithesis_instrumentation__.Notify(576350)
	if v := &n.run.vals; len(v.Rows) > 0 {
		__antithesis_instrumentation__.Notify(576352)
		stmt := fmt.Sprintf("INSERT INTO %s %s", n.target.String(), v.String())
		if _, err := params.p.ExtendedEvalContext().ExecCfg.InternalExecutor.Exec(
			params.ctx,
			"insert into save table",
			nil,
			stmt,
		); err != nil {
			__antithesis_instrumentation__.Notify(576354)
			return errors.Wrapf(err, "while running %s", stmt)
		} else {
			__antithesis_instrumentation__.Notify(576355)
		}
		__antithesis_instrumentation__.Notify(576353)
		v.Rows = nil
	} else {
		__antithesis_instrumentation__.Notify(576356)
	}
	__antithesis_instrumentation__.Notify(576351)
	return nil
}

func (n *saveTableNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(576357)
	res, err := n.source.Next(params)
	if err != nil {
		__antithesis_instrumentation__.Notify(576362)
		return res, err
	} else {
		__antithesis_instrumentation__.Notify(576363)
	}
	__antithesis_instrumentation__.Notify(576358)
	if !res {
		__antithesis_instrumentation__.Notify(576364)

		err := n.issue(params)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(576365)
	}
	__antithesis_instrumentation__.Notify(576359)
	row := n.source.Values()
	exprs := make(tree.Exprs, len(row))
	for i := range row {
		__antithesis_instrumentation__.Notify(576366)
		exprs[i] = row[i]
	}
	__antithesis_instrumentation__.Notify(576360)
	n.run.vals.Rows = append(n.run.vals.Rows, exprs)
	if len(n.run.vals.Rows) >= saveTableInsertBatch {
		__antithesis_instrumentation__.Notify(576367)
		if err := n.issue(params); err != nil {
			__antithesis_instrumentation__.Notify(576368)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(576369)
		}
	} else {
		__antithesis_instrumentation__.Notify(576370)
	}
	__antithesis_instrumentation__.Notify(576361)
	return true, nil
}

func (n *saveTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(576371)
	return n.source.Values()
}

func (n *saveTableNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(576372)
	n.source.Close(ctx)
}
