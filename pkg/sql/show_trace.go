package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
)

type showTraceNode struct {
	columns colinfo.ResultColumns
	compact bool

	kvTracingEnabled bool

	run traceRun
}

func (p *planner) ShowTrace(ctx context.Context, n *tree.ShowTraceForSession) (planNode, error) {
	__antithesis_instrumentation__.Notify(623289)
	var node planNode = p.makeShowTraceNode(n.Compact, n.TraceType == tree.ShowTraceKV)

	ageColIdx := colinfo.GetTraceAgeColumnIdx(n.Compact)
	node = &sortNode{
		plan: node,
		ordering: colinfo.ColumnOrdering{
			colinfo.ColumnOrderInfo{ColIdx: ageColIdx, Direction: encoding.Ascending},
		},
	}

	if n.TraceType == tree.ShowTraceReplica {
		__antithesis_instrumentation__.Notify(623291)
		node = &showTraceReplicaNode{plan: node}
	} else {
		__antithesis_instrumentation__.Notify(623292)
	}
	__antithesis_instrumentation__.Notify(623290)
	return node, nil
}

func (p *planner) makeShowTraceNode(compact bool, kvTracingEnabled bool) *showTraceNode {
	__antithesis_instrumentation__.Notify(623293)
	n := &showTraceNode{
		kvTracingEnabled: kvTracingEnabled,
		compact:          compact,
	}
	if compact {
		__antithesis_instrumentation__.Notify(623295)

		n.columns = append(n.columns, colinfo.ShowCompactTraceColumns...)
	} else {
		__antithesis_instrumentation__.Notify(623296)
		n.columns = append(n.columns, colinfo.ShowTraceColumns...)
	}
	__antithesis_instrumentation__.Notify(623294)
	return n
}

type traceRun struct {
	resultRows []tree.Datums
	curRow     int
}

func (n *showTraceNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(623297)

	traceRows, err := params.extendedEvalCtx.Tracing.getSessionTrace()
	if err != nil {
		__antithesis_instrumentation__.Notify(623299)
		return err
	} else {
		__antithesis_instrumentation__.Notify(623300)
	}
	__antithesis_instrumentation__.Notify(623298)
	n.processTraceRows(params.EvalContext(), traceRows)
	return nil
}

func (n *showTraceNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623301)
	if n.run.curRow >= len(n.run.resultRows) {
		__antithesis_instrumentation__.Notify(623303)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(623304)
	}
	__antithesis_instrumentation__.Notify(623302)
	n.run.curRow++
	return true, nil
}

func (n *showTraceNode) processTraceRows(evalCtx *tree.EvalContext, traceRows []traceRow) {
	__antithesis_instrumentation__.Notify(623305)

	if n.kvTracingEnabled {
		__antithesis_instrumentation__.Notify(623308)
		res := make([]traceRow, 0, len(traceRows))
		for _, r := range traceRows {
			__antithesis_instrumentation__.Notify(623310)
			msg := r[traceMsgCol].(*tree.DString)
			if kvMsgRegexp.MatchString(string(*msg)) {
				__antithesis_instrumentation__.Notify(623311)
				res = append(res, r)
			} else {
				__antithesis_instrumentation__.Notify(623312)
			}
		}
		__antithesis_instrumentation__.Notify(623309)
		traceRows = res
	} else {
		__antithesis_instrumentation__.Notify(623313)
	}
	__antithesis_instrumentation__.Notify(623306)
	if len(traceRows) == 0 {
		__antithesis_instrumentation__.Notify(623314)
		return
	} else {
		__antithesis_instrumentation__.Notify(623315)
	}
	__antithesis_instrumentation__.Notify(623307)

	n.run.resultRows = make([]tree.Datums, len(traceRows))
	for i, r := range traceRows {
		__antithesis_instrumentation__.Notify(623316)
		ts := r[traceTimestampCol].(*tree.DTimestampTZ)
		loc := r[traceLocCol]
		tag := r[traceTagCol]
		msg := r[traceMsgCol]
		spanIdx := r[traceSpanIdxCol]
		op := r[traceOpCol]
		age := r[traceAgeCol]

		if !n.compact {
			__antithesis_instrumentation__.Notify(623317)
			n.run.resultRows[i] = tree.Datums{ts, age, msg, tag, loc, op, spanIdx}
		} else {
			__antithesis_instrumentation__.Notify(623318)
			msgStr := msg.(*tree.DString)
			if locStr := string(*loc.(*tree.DString)); locStr != "" {
				__antithesis_instrumentation__.Notify(623320)
				msgStr = tree.NewDString(fmt.Sprintf("%s %s", locStr, string(*msgStr)))
			} else {
				__antithesis_instrumentation__.Notify(623321)
			}
			__antithesis_instrumentation__.Notify(623319)
			n.run.resultRows[i] = tree.Datums{age, msgStr, tag, op}
		}
	}
}

func (n *showTraceNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623322)
	return n.run.resultRows[n.run.curRow-1]
}

func (n *showTraceNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623323)
	n.run.resultRows = nil
}

var kvMsgRegexp = regexp.MustCompile(
	strings.Join([]string{
		"^fetched: ",
		"^CPut ",
		"^Put ",
		"^InitPut ",
		"^DelRange ",
		"^ClearRange ",
		"^Del ",
		"^Get ",
		"^Scan ",
		"^querying next range at ",
		"^output row: ",
		"^rows affected: ",
		"^execution failed after ",
		"^r.*: sending batch ",
		"^cascading ",
		"^fast path completed",
	}, "|"),
)
