package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

var scanNodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(576373)
		return &scanNode{}
	},
}

type scanNode struct {
	_ util.NoCopy

	desc  catalog.TableDescriptor
	index catalog.Index

	colCfg scanColumnsConfig

	cols []catalog.Column

	resultColumns colinfo.ResultColumns

	spans   []roachpb.Span
	reverse bool

	reqOrdering ReqOrdering

	hardLimit int64

	softLimit int64

	disableBatchLimits bool

	parallelize bool

	isFull bool

	isCheck bool

	estimatedRowCount uint64

	lockingStrength   descpb.ScanLockingStrength
	lockingWaitPolicy descpb.ScanLockingWaitPolicy

	containsSystemColumns bool

	localityOptimized bool
}

type scanColumnsConfig struct {
	wantedColumns []tree.ColumnID

	invertedColumnID   tree.ColumnID
	invertedColumnType *types.T
}

func (cfg scanColumnsConfig) assertValidReqOrdering(reqOrdering exec.OutputOrdering) error {
	__antithesis_instrumentation__.Notify(576374)
	for i := range reqOrdering {
		__antithesis_instrumentation__.Notify(576376)
		if reqOrdering[i].ColIdx >= len(cfg.wantedColumns) {
			__antithesis_instrumentation__.Notify(576377)
			return errors.Errorf("invalid reqOrdering: %v", reqOrdering)
		} else {
			__antithesis_instrumentation__.Notify(576378)
		}
	}
	__antithesis_instrumentation__.Notify(576375)
	return nil
}

func (p *planner) Scan() *scanNode {
	__antithesis_instrumentation__.Notify(576379)
	n := scanNodePool.Get().(*scanNode)
	return n
}

var _ tree.IndexedVarContainer = &scanNode{}

func (n *scanNode) IndexedVarEval(idx int, ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(576380)
	panic("scanNode can't be run in local mode")
}

func (n *scanNode) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(576381)
	return n.resultColumns[idx].Typ
}

func (n *scanNode) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(576382)
	return (*tree.Name)(&n.resultColumns[idx].Name)
}

func (n *scanNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(576383)
	panic("scanNode can't be run in local mode")
}

func (n *scanNode) Close(context.Context) {
	__antithesis_instrumentation__.Notify(576384)
	*n = scanNode{}
	scanNodePool.Put(n)
}

func (n *scanNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(576385)
	panic("scanNode can't be run in local mode")
}

func (n *scanNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(576386)
	panic("scanNode can't be run in local mode")
}

func (n *scanNode) disableBatchLimit() {
	__antithesis_instrumentation__.Notify(576387)
	n.disableBatchLimits = true
	n.hardLimit = 0
	n.softLimit = 0
}

func (n *scanNode) initTable(
	ctx context.Context, p *planner, desc catalog.TableDescriptor, colCfg scanColumnsConfig,
) error {
	__antithesis_instrumentation__.Notify(576388)
	n.desc = desc

	n.containsSystemColumns = scanContainsSystemColumns(&colCfg)

	return n.initDescDefaults(colCfg)
}

func initColsForScan(
	desc catalog.TableDescriptor, colCfg scanColumnsConfig,
) (cols []catalog.Column, err error) {
	__antithesis_instrumentation__.Notify(576389)
	if colCfg.wantedColumns == nil {
		__antithesis_instrumentation__.Notify(576392)
		return nil, errors.AssertionFailedf("wantedColumns is nil")
	} else {
		__antithesis_instrumentation__.Notify(576393)
	}
	__antithesis_instrumentation__.Notify(576390)

	cols = make([]catalog.Column, len(colCfg.wantedColumns))
	for i, colID := range colCfg.wantedColumns {
		__antithesis_instrumentation__.Notify(576394)
		col, err := desc.FindColumnWithID(colID)
		if err != nil {
			__antithesis_instrumentation__.Notify(576397)
			return cols, err
		} else {
			__antithesis_instrumentation__.Notify(576398)
		}
		__antithesis_instrumentation__.Notify(576395)

		if colCfg.invertedColumnID == colID && func() bool {
			__antithesis_instrumentation__.Notify(576399)
			return !colCfg.invertedColumnType.Identical(col.GetType()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(576400)
			col = col.DeepCopy()
			col.ColumnDesc().Type = colCfg.invertedColumnType
		} else {
			__antithesis_instrumentation__.Notify(576401)
		}
		__antithesis_instrumentation__.Notify(576396)
		cols[i] = col
	}
	__antithesis_instrumentation__.Notify(576391)

	return cols, nil
}

func (n *scanNode) initDescDefaults(colCfg scanColumnsConfig) error {
	__antithesis_instrumentation__.Notify(576402)
	n.colCfg = colCfg
	n.index = n.desc.GetPrimaryIndex()

	var err error
	n.cols, err = initColsForScan(n.desc, n.colCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(576404)
		return err
	} else {
		__antithesis_instrumentation__.Notify(576405)
	}
	__antithesis_instrumentation__.Notify(576403)

	n.resultColumns = colinfo.ResultColumnsFromColumns(n.desc.GetID(), n.cols)
	return nil
}
