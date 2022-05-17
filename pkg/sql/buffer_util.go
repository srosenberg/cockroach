package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
)

type rowContainerHelper struct {
	memMonitor  *mon.BytesMonitor
	diskMonitor *mon.BytesMonitor
	rows        *rowcontainer.DiskBackedRowContainer
	scratch     rowenc.EncDatumRow
}

func (c *rowContainerHelper) Init(
	typs []*types.T, evalContext *extendedEvalContext, opName string,
) {
	__antithesis_instrumentation__.Notify(247142)
	c.initMonitors(evalContext, opName)
	distSQLCfg := &evalContext.DistSQLPlanner.distSQLSrv.ServerConfig
	c.rows = &rowcontainer.DiskBackedRowContainer{}
	c.rows.Init(
		colinfo.NoOrdering, typs, &evalContext.EvalContext,
		distSQLCfg.TempStorage, c.memMonitor, c.diskMonitor,
	)
	c.scratch = make(rowenc.EncDatumRow, len(typs))
}

func (c *rowContainerHelper) InitWithDedup(
	typs []*types.T, evalContext *extendedEvalContext, opName string,
) {
	__antithesis_instrumentation__.Notify(247143)
	c.initMonitors(evalContext, opName)
	distSQLCfg := &evalContext.DistSQLPlanner.distSQLSrv.ServerConfig
	c.rows = &rowcontainer.DiskBackedRowContainer{}

	ordering := make(colinfo.ColumnOrdering, len(typs))
	for i := range ordering {
		__antithesis_instrumentation__.Notify(247145)
		ordering[i].ColIdx = i
		ordering[i].Direction = encoding.Ascending
	}
	__antithesis_instrumentation__.Notify(247144)
	c.rows.Init(
		ordering, typs, &evalContext.EvalContext,
		distSQLCfg.TempStorage, c.memMonitor, c.diskMonitor,
	)
	c.rows.DoDeDuplicate()
	c.scratch = make(rowenc.EncDatumRow, len(typs))
}

func (c *rowContainerHelper) initMonitors(evalContext *extendedEvalContext, opName string) {
	__antithesis_instrumentation__.Notify(247146)
	distSQLCfg := &evalContext.DistSQLPlanner.distSQLSrv.ServerConfig
	c.memMonitor = execinfra.NewLimitedMonitorNoFlowCtx(
		evalContext.Context, evalContext.Mon, distSQLCfg, evalContext.SessionData(),
		fmt.Sprintf("%s-limited", opName),
	)
	c.diskMonitor = execinfra.NewMonitor(
		evalContext.Context, distSQLCfg.ParentDiskMonitor, fmt.Sprintf("%s-disk", opName),
	)
}

func (c *rowContainerHelper) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(247147)
	for i := range row {
		__antithesis_instrumentation__.Notify(247149)
		c.scratch[i].Datum = row[i]
	}
	__antithesis_instrumentation__.Notify(247148)
	return c.rows.AddRow(ctx, c.scratch)
}

func (c *rowContainerHelper) AddRowWithDedup(
	ctx context.Context, row tree.Datums,
) (added bool, _ error) {
	__antithesis_instrumentation__.Notify(247150)
	for i := range row {
		__antithesis_instrumentation__.Notify(247153)
		c.scratch[i].Datum = row[i]
	}
	__antithesis_instrumentation__.Notify(247151)
	lenBefore := c.rows.Len()
	if _, err := c.rows.AddRowWithDeDup(ctx, c.scratch); err != nil {
		__antithesis_instrumentation__.Notify(247154)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247155)
	}
	__antithesis_instrumentation__.Notify(247152)
	return c.rows.Len() > lenBefore, nil
}

func (c *rowContainerHelper) Len() int {
	__antithesis_instrumentation__.Notify(247156)
	return c.rows.Len()
}

func (c *rowContainerHelper) Clear(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(247157)
	return c.rows.UnsafeReset(ctx)
}

func (c *rowContainerHelper) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(247158)
	if c.rows != nil {
		__antithesis_instrumentation__.Notify(247159)
		c.rows.Close(ctx)
		c.memMonitor.Stop(ctx)
		c.diskMonitor.Stop(ctx)
		c.rows = nil
	} else {
		__antithesis_instrumentation__.Notify(247160)
	}
}

type rowContainerIterator struct {
	iter rowcontainer.RowIterator

	typs   []*types.T
	datums tree.Datums
	da     tree.DatumAlloc
}

func newRowContainerIterator(
	ctx context.Context, c rowContainerHelper, typs []*types.T,
) *rowContainerIterator {
	__antithesis_instrumentation__.Notify(247161)
	i := &rowContainerIterator{
		iter:   c.rows.NewIterator(ctx),
		typs:   typs,
		datums: make(tree.Datums, len(typs)),
	}
	i.iter.Rewind()
	return i
}

func (i *rowContainerIterator) Next() (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(247162)
	defer i.iter.Next()
	if valid, err := i.iter.Valid(); err != nil {
		__antithesis_instrumentation__.Notify(247166)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(247167)
		if !valid {
			__antithesis_instrumentation__.Notify(247168)

			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(247169)
		}
	}
	__antithesis_instrumentation__.Notify(247163)
	row, err := i.iter.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(247170)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(247171)
	}
	__antithesis_instrumentation__.Notify(247164)
	if err = rowenc.EncDatumRowToDatums(i.typs, i.datums, row, &i.da); err != nil {
		__antithesis_instrumentation__.Notify(247172)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(247173)
	}
	__antithesis_instrumentation__.Notify(247165)
	return i.datums, nil
}

func (i *rowContainerIterator) Close() {
	__antithesis_instrumentation__.Notify(247174)
	i.iter.Close()
}
