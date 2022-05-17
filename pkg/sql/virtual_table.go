package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/constraint"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/exec"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

type virtualTableGenerator func() (tree.Datums, error)

type cleanupFunc func(ctx context.Context)

type rowPusher interface {
	pushRow(...tree.Datum) error
}

type funcRowPusher func(...tree.Datum) error

func (f funcRowPusher) pushRow(datums ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(632638)
	return f(datums...)
}

type virtualTableGeneratorResponse struct {
	datums tree.Datums
	err    error
}

func setupGenerator(
	ctx context.Context,
	worker func(ctx context.Context, pusher rowPusher) error,
	stopper *stop.Stopper,
) (next virtualTableGenerator, cleanup cleanupFunc, setupError error) {
	__antithesis_instrumentation__.Notify(632639)
	var cancel func()
	ctx, cancel = context.WithCancel(ctx)
	var wg sync.WaitGroup
	cleanup = func(context.Context) {
		__antithesis_instrumentation__.Notify(632644)
		cancel()
		wg.Wait()
	}
	__antithesis_instrumentation__.Notify(632640)

	comm := make(chan virtualTableGeneratorResponse)
	addRow := func(datums ...tree.Datum) error {
		__antithesis_instrumentation__.Notify(632645)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(632648)
			return cancelchecker.QueryCanceledError
		case comm <- virtualTableGeneratorResponse{datums: datums}:
			__antithesis_instrumentation__.Notify(632649)
		}
		__antithesis_instrumentation__.Notify(632646)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(632650)
			return cancelchecker.QueryCanceledError
		case <-comm:
			__antithesis_instrumentation__.Notify(632651)
		}
		__antithesis_instrumentation__.Notify(632647)
		return nil
	}
	__antithesis_instrumentation__.Notify(632641)

	wg.Add(1)
	if setupError = stopper.RunAsyncTaskEx(ctx,
		stop.TaskOpts{
			TaskName: "sql.rowPusher: send rows",
			SpanOpt:  stop.ChildSpan,
		},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(632652)
			defer wg.Done()

			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(632655)
				return
			case <-comm:
				__antithesis_instrumentation__.Notify(632656)
			}
			__antithesis_instrumentation__.Notify(632653)
			err := worker(ctx, funcRowPusher(addRow))

			if errors.Is(err, cancelchecker.QueryCanceledError) {
				__antithesis_instrumentation__.Notify(632657)
				return
			} else {
				__antithesis_instrumentation__.Notify(632658)
			}
			__antithesis_instrumentation__.Notify(632654)

			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(632659)
				return
			case comm <- virtualTableGeneratorResponse{err: err}:
				__antithesis_instrumentation__.Notify(632660)
			}
		}); setupError != nil {
		__antithesis_instrumentation__.Notify(632661)

		wg.Done()
	} else {
		__antithesis_instrumentation__.Notify(632662)
	}
	__antithesis_instrumentation__.Notify(632642)

	next = func() (tree.Datums, error) {
		__antithesis_instrumentation__.Notify(632663)

		select {
		case comm <- virtualTableGeneratorResponse{}:
			__antithesis_instrumentation__.Notify(632665)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(632666)
			return nil, cancelchecker.QueryCanceledError
		}
		__antithesis_instrumentation__.Notify(632664)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(632667)
			return nil, cancelchecker.QueryCanceledError
		case resp := <-comm:
			__antithesis_instrumentation__.Notify(632668)
			return resp.datums, resp.err
		}
	}
	__antithesis_instrumentation__.Notify(632643)
	return next, cleanup, setupError
}

type virtualTableNode struct {
	columns    colinfo.ResultColumns
	next       virtualTableGenerator
	cleanup    func(ctx context.Context)
	currentRow tree.Datums
}

func (p *planner) newVirtualTableNode(
	columns colinfo.ResultColumns, next virtualTableGenerator, cleanup func(ctx context.Context),
) *virtualTableNode {
	__antithesis_instrumentation__.Notify(632669)
	return &virtualTableNode{
		columns: columns,
		next:    next,
		cleanup: cleanup,
	}
}

func (n *virtualTableNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(632670)
	return nil
}

func (n *virtualTableNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(632671)
	row, err := n.next()
	if err != nil {
		__antithesis_instrumentation__.Notify(632673)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(632674)
	}
	__antithesis_instrumentation__.Notify(632672)
	n.currentRow = row
	return row != nil, nil
}

func (n *virtualTableNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(632675)
	return n.currentRow
}

func (n *virtualTableNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(632676)
	if n.cleanup != nil {
		__antithesis_instrumentation__.Notify(632677)
		n.cleanup(ctx)
	} else {
		__antithesis_instrumentation__.Notify(632678)
	}
}

type vTableLookupJoinNode struct {
	input planNode

	dbName string
	db     catalog.DatabaseDescriptor
	table  catalog.TableDescriptor
	index  catalog.Index

	eqCol             int
	virtualTableEntry *virtualDefEntry

	joinType descpb.JoinType

	columns colinfo.ResultColumns

	pred *joinPredicate

	inputCols colinfo.ResultColumns

	vtableCols colinfo.ResultColumns

	lookupCols exec.TableColumnOrdinalSet

	run struct {
		row tree.Datums

		rows   *rowcontainer.RowContainer
		keyCtx constraint.KeyContext

		indexKeyDatums []tree.Datum

		params *runParams
	}
}

var _ planNode = &vTableLookupJoinNode{}
var _ rowPusher = &vTableLookupJoinNode{}

func (v *vTableLookupJoinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(632679)
	v.run.keyCtx = constraint.KeyContext{EvalCtx: params.EvalContext()}
	v.run.rows = rowcontainer.NewRowContainer(
		params.EvalContext().Mon.MakeBoundAccount(),
		colinfo.ColTypeInfoFromResCols(v.columns),
	)
	v.run.indexKeyDatums = make(tree.Datums, len(v.columns))
	var err error
	db, err := params.p.Descriptors().GetImmutableDatabaseByName(
		params.ctx,
		params.p.txn,
		v.dbName,
		tree.DatabaseLookupFlags{
			Required: true, AvoidLeased: params.p.avoidLeasedDescriptors,
		},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(632681)
		return err
	} else {
		__antithesis_instrumentation__.Notify(632682)
	}
	__antithesis_instrumentation__.Notify(632680)
	v.db = db
	return err
}

func (v *vTableLookupJoinNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(632683)

	v.run.params = &params
	for {
		__antithesis_instrumentation__.Notify(632684)

		if v.run.rows.Len() > 0 {
			__antithesis_instrumentation__.Notify(632688)
			copy(v.run.row, v.run.rows.At(0))
			v.run.rows.PopFirst(params.ctx)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(632689)
		}
		__antithesis_instrumentation__.Notify(632685)

		ok, err := v.input.Next(params)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(632690)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(632691)
			return ok, err
		} else {
			__antithesis_instrumentation__.Notify(632692)
		}
		__antithesis_instrumentation__.Notify(632686)
		inputRow := v.input.Values()
		var span constraint.Span
		datum := inputRow[v.eqCol]

		key := constraint.MakeKey(datum)
		span.Init(key, constraint.IncludeBoundary, key, constraint.IncludeBoundary)
		var idxConstraint constraint.Constraint
		idxConstraint.InitSingleSpan(&v.run.keyCtx, &span)

		genFunc := v.virtualTableEntry.makeConstrainedRowsGenerator(
			params.p, v.db, v.index,
			v.run.indexKeyDatums,
			catalog.ColumnIDToOrdinalMap(v.table.PublicColumns()),
			&idxConstraint,
			v.vtableCols,
		)

		v.run.row = append(v.run.row[:0], inputRow...)

		if err := genFunc(params.ctx, v); err != nil {
			__antithesis_instrumentation__.Notify(632693)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(632694)
		}
		__antithesis_instrumentation__.Notify(632687)
		if v.run.rows.Len() == 0 && func() bool {
			__antithesis_instrumentation__.Notify(632695)
			return v.joinType == descpb.LeftOuterJoin == true
		}() == true {
			__antithesis_instrumentation__.Notify(632696)

			v.run.row = v.run.row[:len(v.inputCols)]
			for i := len(inputRow); i < len(v.columns); i++ {
				__antithesis_instrumentation__.Notify(632698)
				v.run.row = append(v.run.row, tree.DNull)
			}
			__antithesis_instrumentation__.Notify(632697)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(632699)
		}
	}
}

func (v *vTableLookupJoinNode) pushRow(lookedUpRow ...tree.Datum) error {
	__antithesis_instrumentation__.Notify(632700)

	v.run.row = v.run.row[:len(v.inputCols)]

	for i, ok := v.lookupCols.Next(0); ok; i, ok = v.lookupCols.Next(i + 1) {
		__antithesis_instrumentation__.Notify(632703)

		v.run.row = append(v.run.row, lookedUpRow[i-1])
	}
	__antithesis_instrumentation__.Notify(632701)

	if ok, err := v.pred.eval(v.run.params.EvalContext(),
		v.run.row[:len(v.inputCols)],
		v.run.row[len(v.inputCols):]); !ok || func() bool {
		__antithesis_instrumentation__.Notify(632704)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(632705)
		return err
	} else {
		__antithesis_instrumentation__.Notify(632706)
	}
	__antithesis_instrumentation__.Notify(632702)
	_, err := v.run.rows.AddRow(v.run.params.ctx, v.run.row)
	return err
}

func (v *vTableLookupJoinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(632707)
	return v.run.row
}

func (v *vTableLookupJoinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(632708)
	v.input.Close(ctx)
	v.run.rows.Close(ctx)
}
