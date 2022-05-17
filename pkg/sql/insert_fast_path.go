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
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var insertFastPathNodePool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(497466)
		return &insertFastPathNode{}
	},
}

type insertFastPathNode struct {
	input [][]tree.TypedExpr

	columns colinfo.ResultColumns

	run insertFastPathRun
}

var _ mutationPlanNode = &insertFastPathNode{}

type insertFastPathRun struct {
	insertRun

	fkChecks []insertFastPathFKCheck

	numInputCols int

	inputBuf tree.Datums

	fkBatch roachpb.BatchRequest

	fkSpanInfo []insertFastPathFKSpanInfo

	fkSpanMap map[string]struct{}
}

type insertFastPathFKSpanInfo struct {
	check  *insertFastPathFKCheck
	rowIdx int
}

type insertFastPathFKCheck struct {
	exec.InsertFastPathFKCheck

	tabDesc      catalog.TableDescriptor
	idx          catalog.Index
	keyPrefix    []byte
	colMap       catalog.TableColMap
	spanBuilder  span.Builder
	spanSplitter span.Splitter
}

func (c *insertFastPathFKCheck) init(params runParams) error {
	idx := c.ReferencedIndex.(*optIndex)
	c.tabDesc = c.ReferencedTable.(*optTable).desc
	c.idx = idx.idx

	codec := params.ExecCfg().Codec
	c.keyPrefix = rowenc.MakeIndexKeyPrefix(codec, c.tabDesc.GetID(), c.idx.GetID())
	c.spanBuilder.Init(params.EvalContext(), codec, c.tabDesc, c.idx)
	c.spanSplitter = span.MakeSplitter(c.tabDesc, c.idx, util.FastIntSet{})

	if len(c.InsertCols) > idx.numLaxKeyCols {
		return errors.AssertionFailedf(
			"%d FK cols, only %d cols in index", len(c.InsertCols), idx.numLaxKeyCols,
		)
	}
	for i, ord := range c.InsertCols {
		var colID descpb.ColumnID
		if i < c.idx.NumKeyColumns() {
			colID = c.idx.GetKeyColumnID(i)
		} else {
			colID = c.idx.GetKeySuffixColumnID(i - c.idx.NumKeyColumns())
		}

		c.colMap.Set(colID, int(ord))
	}
	return nil
}

func (c *insertFastPathFKCheck) generateSpan(inputRow tree.Datums) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(497467)
	return row.FKCheckSpan(&c.spanBuilder, c.spanSplitter, inputRow, c.colMap, len(c.InsertCols))
}

func (c *insertFastPathFKCheck) errorForRow(inputRow tree.Datums) error {
	__antithesis_instrumentation__.Notify(497468)
	values := make(tree.Datums, len(c.InsertCols))
	for i, ord := range c.InsertCols {
		__antithesis_instrumentation__.Notify(497470)
		values[i] = inputRow[ord]
	}
	__antithesis_instrumentation__.Notify(497469)
	return c.MkErr(values)
}

func (r *insertFastPathRun) inputRow(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(497471)
	start := rowIdx * r.numInputCols
	end := start + r.numInputCols
	return r.inputBuf[start:end:end]
}

func (r *insertFastPathRun) addFKChecks(
	ctx context.Context, rowIdx int, inputRow tree.Datums,
) error {
	__antithesis_instrumentation__.Notify(497472)
	for i := range r.fkChecks {
		__antithesis_instrumentation__.Notify(497474)
		c := &r.fkChecks[i]

		numNulls := 0
		for _, ord := range c.InsertCols {
			__antithesis_instrumentation__.Notify(497480)
			if inputRow[ord] == tree.DNull {
				__antithesis_instrumentation__.Notify(497481)
				numNulls++
			} else {
				__antithesis_instrumentation__.Notify(497482)
			}
		}
		__antithesis_instrumentation__.Notify(497475)
		if numNulls > 0 {
			__antithesis_instrumentation__.Notify(497483)
			if c.MatchMethod == tree.MatchFull && func() bool {
				__antithesis_instrumentation__.Notify(497485)
				return numNulls != len(c.InsertCols) == true
			}() == true {
				__antithesis_instrumentation__.Notify(497486)
				return c.errorForRow(inputRow)
			} else {
				__antithesis_instrumentation__.Notify(497487)
			}
			__antithesis_instrumentation__.Notify(497484)

			continue
		} else {
			__antithesis_instrumentation__.Notify(497488)
		}
		__antithesis_instrumentation__.Notify(497476)

		span, err := c.generateSpan(inputRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(497489)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497490)
		}
		__antithesis_instrumentation__.Notify(497477)
		if r.fkSpanMap != nil {
			__antithesis_instrumentation__.Notify(497491)
			_, exists := r.fkSpanMap[string(span.Key)]
			if exists {
				__antithesis_instrumentation__.Notify(497493)

				continue
			} else {
				__antithesis_instrumentation__.Notify(497494)
			}
			__antithesis_instrumentation__.Notify(497492)
			r.fkSpanMap[string(span.Key)] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(497495)
		}
		__antithesis_instrumentation__.Notify(497478)
		if r.traceKV {
			__antithesis_instrumentation__.Notify(497496)
			log.VEventf(ctx, 2, "FKScan %s", span)
		} else {
			__antithesis_instrumentation__.Notify(497497)
		}
		__antithesis_instrumentation__.Notify(497479)
		reqIdx := len(r.fkBatch.Requests)
		r.fkBatch.Requests = append(r.fkBatch.Requests, roachpb.RequestUnion{})
		r.fkBatch.Requests[reqIdx].MustSetInner(&roachpb.ScanRequest{
			RequestHeader: roachpb.RequestHeaderFromSpan(span),
		})
		r.fkSpanInfo = append(r.fkSpanInfo, insertFastPathFKSpanInfo{
			check:  c,
			rowIdx: rowIdx,
		})
	}
	__antithesis_instrumentation__.Notify(497473)
	return nil
}

func (n *insertFastPathNode) runFKChecks(params runParams) error {
	__antithesis_instrumentation__.Notify(497498)
	if len(n.run.fkBatch.Requests) == 0 {
		__antithesis_instrumentation__.Notify(497502)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(497503)
	}
	__antithesis_instrumentation__.Notify(497499)
	defer n.run.fkBatch.Reset()

	br, err := params.p.txn.Send(params.ctx, n.run.fkBatch)
	if err != nil {
		__antithesis_instrumentation__.Notify(497504)
		return err.GoError()
	} else {
		__antithesis_instrumentation__.Notify(497505)
	}
	__antithesis_instrumentation__.Notify(497500)

	for i := range br.Responses {
		__antithesis_instrumentation__.Notify(497506)
		resp := br.Responses[i].GetInner().(*roachpb.ScanResponse)
		if len(resp.Rows) == 0 {
			__antithesis_instrumentation__.Notify(497507)

			info := n.run.fkSpanInfo[i]
			return info.check.errorForRow(n.run.inputRow(info.rowIdx))
		} else {
			__antithesis_instrumentation__.Notify(497508)
		}
	}
	__antithesis_instrumentation__.Notify(497501)

	return nil
}

func (n *insertFastPathNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(497509)

	n.run.traceKV = params.p.ExtendedEvalContext().Tracing.KVTracingEnabled()

	n.run.initRowContainer(params, n.columns)

	n.run.numInputCols = len(n.input[0])
	n.run.inputBuf = make(tree.Datums, len(n.input)*n.run.numInputCols)

	if len(n.input) > 1 {
		__antithesis_instrumentation__.Notify(497512)
		n.run.fkSpanMap = make(map[string]struct{})
	} else {
		__antithesis_instrumentation__.Notify(497513)
	}
	__antithesis_instrumentation__.Notify(497510)

	if len(n.run.fkChecks) > 0 {
		__antithesis_instrumentation__.Notify(497514)
		for i := range n.run.fkChecks {
			__antithesis_instrumentation__.Notify(497516)
			if err := n.run.fkChecks[i].init(params); err != nil {
				__antithesis_instrumentation__.Notify(497517)
				return err
			} else {
				__antithesis_instrumentation__.Notify(497518)
			}
		}
		__antithesis_instrumentation__.Notify(497515)
		maxSpans := len(n.run.fkChecks) * len(n.input)
		n.run.fkBatch.Requests = make([]roachpb.RequestUnion, 0, maxSpans)
		n.run.fkSpanInfo = make([]insertFastPathFKSpanInfo, 0, maxSpans)
		if len(n.input) > 1 {
			__antithesis_instrumentation__.Notify(497519)
			n.run.fkSpanMap = make(map[string]struct{}, maxSpans)
		} else {
			__antithesis_instrumentation__.Notify(497520)
		}
	} else {
		__antithesis_instrumentation__.Notify(497521)
	}
	__antithesis_instrumentation__.Notify(497511)

	return n.run.ti.init(params.ctx, params.p.txn, params.EvalContext(), &params.EvalContext().Settings.SV)
}

func (n *insertFastPathNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(497522)
	panic("not valid")
}

func (n *insertFastPathNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(497523)
	panic("not valid")
}

func (n *insertFastPathNode) BatchedNext(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(497524)
	if n.run.done {
		__antithesis_instrumentation__.Notify(497529)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(497530)
	}
	__antithesis_instrumentation__.Notify(497525)

	for rowIdx, tupleRow := range n.input {
		__antithesis_instrumentation__.Notify(497531)
		if err := params.p.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(497535)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(497536)
		}
		__antithesis_instrumentation__.Notify(497532)
		inputRow := n.run.inputRow(rowIdx)
		for col, typedExpr := range tupleRow {
			__antithesis_instrumentation__.Notify(497537)
			var err error
			inputRow[col], err = typedExpr.Eval(params.EvalContext())
			if err != nil {
				__antithesis_instrumentation__.Notify(497538)
				err = interceptAlterColumnTypeParseError(n.run.insertCols, col, err)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(497539)
			}
		}
		__antithesis_instrumentation__.Notify(497533)

		if err := n.run.processSourceRow(params, inputRow); err != nil {
			__antithesis_instrumentation__.Notify(497540)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(497541)
		}
		__antithesis_instrumentation__.Notify(497534)

		if len(n.run.fkChecks) > 0 {
			__antithesis_instrumentation__.Notify(497542)
			if err := n.run.addFKChecks(params.ctx, rowIdx, inputRow); err != nil {
				__antithesis_instrumentation__.Notify(497543)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(497544)
			}
		} else {
			__antithesis_instrumentation__.Notify(497545)
		}
	}
	__antithesis_instrumentation__.Notify(497526)

	if err := n.runFKChecks(params); err != nil {
		__antithesis_instrumentation__.Notify(497546)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(497547)
	}
	__antithesis_instrumentation__.Notify(497527)

	n.run.ti.setRowsWrittenLimit(params.extendedEvalCtx.SessionData())
	if err := n.run.ti.finalize(params.ctx); err != nil {
		__antithesis_instrumentation__.Notify(497548)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(497549)
	}
	__antithesis_instrumentation__.Notify(497528)

	n.run.done = true

	params.ExecCfg().StatsRefresher.NotifyMutation(n.run.ti.ri.Helper.TableDesc, len(n.input))

	return true, nil
}

func (n *insertFastPathNode) BatchedCount() int {
	__antithesis_instrumentation__.Notify(497550)
	return len(n.input)
}

func (n *insertFastPathNode) BatchedValues(rowIdx int) tree.Datums {
	__antithesis_instrumentation__.Notify(497551)
	return n.run.ti.rows.At(rowIdx)
}

func (n *insertFastPathNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(497552)
	n.run.ti.close(ctx)
	*n = insertFastPathNode{}
	insertFastPathNodePool.Put(n)
}

func (n *insertFastPathNode) rowsWritten() int64 {
	__antithesis_instrumentation__.Notify(497553)
	return n.run.ti.rowsWritten
}

func (n *insertFastPathNode) enableAutoCommit() {
	__antithesis_instrumentation__.Notify(497554)
	n.run.ti.enableAutoCommit()
}

func interceptAlterColumnTypeParseError(insertCols []catalog.Column, colNum int, err error) error {
	__antithesis_instrumentation__.Notify(497555)

	if colNum >= len(insertCols) {
		__antithesis_instrumentation__.Notify(497560)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497561)
	}
	__antithesis_instrumentation__.Notify(497556)
	var insertCol catalog.Column

	wrapParseError := func(insertCol catalog.Column, colNum int, err error) (bool, error) {
		__antithesis_instrumentation__.Notify(497562)
		if insertCol.ColumnDesc().AlterColumnTypeInProgress {
			__antithesis_instrumentation__.Notify(497564)
			code := pgerror.GetPGCode(err)
			if code == pgcode.InvalidTextRepresentation {
				__antithesis_instrumentation__.Notify(497565)
				if colNum != -1 {
					__antithesis_instrumentation__.Notify(497567)

					return true, errors.Wrapf(err,
						"This table is still undergoing the ALTER COLUMN TYPE schema change, "+
							"this insert is not supported until the schema change is finalized")
				} else {
					__antithesis_instrumentation__.Notify(497568)
				}
				__antithesis_instrumentation__.Notify(497566)

				return true, errors.Wrap(err,
					"This table is still undergoing the ALTER COLUMN TYPE schema change, "+
						"this insert may not be supported until the schema change is finalized")
			} else {
				__antithesis_instrumentation__.Notify(497569)
			}
		} else {
			__antithesis_instrumentation__.Notify(497570)
		}
		__antithesis_instrumentation__.Notify(497563)
		return false, err
	}
	__antithesis_instrumentation__.Notify(497557)

	if colNum != -1 {
		__antithesis_instrumentation__.Notify(497571)
		insertCol = insertCols[colNum]
		_, err = wrapParseError(insertCol, colNum, err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(497572)
	}
	__antithesis_instrumentation__.Notify(497558)

	for _, insertCol = range insertCols {
		__antithesis_instrumentation__.Notify(497573)
		var changed bool
		changed, err = wrapParseError(insertCol, colNum, err)
		if changed {
			__antithesis_instrumentation__.Notify(497574)
			return err
		} else {
			__antithesis_instrumentation__.Notify(497575)
		}
	}
	__antithesis_instrumentation__.Notify(497559)

	return err
}
