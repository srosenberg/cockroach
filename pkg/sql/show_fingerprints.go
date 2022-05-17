package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type showFingerprintsNode struct {
	optColumnsSlot

	tableDesc catalog.TableDescriptor
	indexes   []catalog.Index

	run showFingerprintsRun
}

func (p *planner) ShowFingerprints(
	ctx context.Context, n *tree.ShowFingerprints,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(623148)

	tableDesc, err := p.ResolveUncachedTableDescriptorEx(
		ctx, n.Table, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(623151)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623152)
	}
	__antithesis_instrumentation__.Notify(623149)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.SELECT); err != nil {
		__antithesis_instrumentation__.Notify(623153)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623154)
	}
	__antithesis_instrumentation__.Notify(623150)

	return &showFingerprintsNode{
		tableDesc: tableDesc,
		indexes:   tableDesc.NonDropIndexes(),
	}, nil
}

type showFingerprintsRun struct {
	rowIdx int

	values []tree.Datum
}

func (n *showFingerprintsNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623155)
	n.run.values = []tree.Datum{tree.DNull, tree.DNull}
	return nil
}

func (n *showFingerprintsNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623156)
	if n.run.rowIdx >= len(n.indexes) {
		__antithesis_instrumentation__.Notify(623163)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(623164)
	}
	__antithesis_instrumentation__.Notify(623157)
	index := n.indexes[n.run.rowIdx]

	cols := make([]string, 0, len(n.tableDesc.PublicColumns()))
	addColumn := func(col catalog.Column) {
		__antithesis_instrumentation__.Notify(623165)

		name := col.GetName()
		switch col.GetType().Family() {
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(623166)
			cols = append(cols, fmt.Sprintf("%s:::bytes", tree.NameStringP(&name)))
		default:
			__antithesis_instrumentation__.Notify(623167)
			cols = append(cols, fmt.Sprintf("%s::string::bytes", tree.NameStringP(&name)))
		}
	}
	__antithesis_instrumentation__.Notify(623158)

	if index.Primary() {
		__antithesis_instrumentation__.Notify(623168)
		for _, col := range n.tableDesc.PublicColumns() {
			__antithesis_instrumentation__.Notify(623169)
			addColumn(col)
		}
	} else {
		__antithesis_instrumentation__.Notify(623170)
		for i := 0; i < index.NumKeyColumns(); i++ {
			__antithesis_instrumentation__.Notify(623173)
			col, err := n.tableDesc.FindColumnWithID(index.GetKeyColumnID(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(623175)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(623176)
			}
			__antithesis_instrumentation__.Notify(623174)
			addColumn(col)
		}
		__antithesis_instrumentation__.Notify(623171)
		for i := 0; i < index.NumKeySuffixColumns(); i++ {
			__antithesis_instrumentation__.Notify(623177)
			col, err := n.tableDesc.FindColumnWithID(index.GetKeySuffixColumnID(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(623179)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(623180)
			}
			__antithesis_instrumentation__.Notify(623178)
			addColumn(col)
		}
		__antithesis_instrumentation__.Notify(623172)
		for i := 0; i < index.NumSecondaryStoredColumns(); i++ {
			__antithesis_instrumentation__.Notify(623181)
			col, err := n.tableDesc.FindColumnWithID(index.GetStoredColumnID(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(623183)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(623184)
			}
			__antithesis_instrumentation__.Notify(623182)
			addColumn(col)
		}
	}
	__antithesis_instrumentation__.Notify(623159)

	sql := fmt.Sprintf(`SELECT
	  xor_agg(fnv64(%s))::string AS fingerprint
	  FROM [%d AS t]@{FORCE_INDEX=[%d]}
	`, strings.Join(cols, `,`), n.tableDesc.GetID(), index.GetID())

	if params.p.EvalContext().AsOfSystemTime != nil {
		__antithesis_instrumentation__.Notify(623185)
		ts := params.p.txn.ReadTimestamp()
		sql = sql + " AS OF SYSTEM TIME " + ts.AsOfSystemTime()
	} else {
		__antithesis_instrumentation__.Notify(623186)
	}
	__antithesis_instrumentation__.Notify(623160)

	fingerprintCols, err := params.extendedEvalCtx.ExecCfg.InternalExecutor.QueryRowEx(
		params.ctx, "hash-fingerprint",
		params.p.txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		sql,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(623187)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(623188)
	}
	__antithesis_instrumentation__.Notify(623161)

	if len(fingerprintCols) != 1 {
		__antithesis_instrumentation__.Notify(623189)
		return false, errors.AssertionFailedf(
			"unexpected number of columns returned: 1 vs %d",
			len(fingerprintCols))
	} else {
		__antithesis_instrumentation__.Notify(623190)
	}
	__antithesis_instrumentation__.Notify(623162)
	fingerprint := fingerprintCols[0]

	n.run.values[0] = tree.NewDString(index.GetName())
	n.run.values[1] = fingerprint
	n.run.rowIdx++
	return true, nil
}

func (n *showFingerprintsNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623191)
	return n.run.values
}
func (n *showFingerprintsNode) Close(_ context.Context) {
	__antithesis_instrumentation__.Notify(623192)
}
