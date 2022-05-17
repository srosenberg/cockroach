package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"

type ReqOrdering = colinfo.ColumnOrdering

func planReqOrdering(plan planNode) ReqOrdering {
	__antithesis_instrumentation__.Notify(563018)
	switch n := plan.(type) {
	case *limitNode:
		__antithesis_instrumentation__.Notify(563020)
		return planReqOrdering(n.plan)
	case *max1RowNode:
		__antithesis_instrumentation__.Notify(563021)
		return planReqOrdering(n.plan)
	case *spoolNode:
		__antithesis_instrumentation__.Notify(563022)
		return planReqOrdering(n.source)
	case *saveTableNode:
		__antithesis_instrumentation__.Notify(563023)
		return planReqOrdering(n.source)
	case *serializeNode:
		__antithesis_instrumentation__.Notify(563024)
		return planReqOrdering(n.source)
	case *deleteNode:
		__antithesis_instrumentation__.Notify(563025)
		if n.run.rowsNeeded {
			__antithesis_instrumentation__.Notify(563043)
			return planReqOrdering(n.source)
		} else {
			__antithesis_instrumentation__.Notify(563044)
		}

	case *filterNode:
		__antithesis_instrumentation__.Notify(563026)
		return n.reqOrdering

	case *groupNode:
		__antithesis_instrumentation__.Notify(563027)
		return n.reqOrdering

	case *distinctNode:
		__antithesis_instrumentation__.Notify(563028)
		return n.reqOrdering

	case *indexJoinNode:
		__antithesis_instrumentation__.Notify(563029)
		return n.reqOrdering

	case *windowNode:
		__antithesis_instrumentation__.Notify(563030)

	case *joinNode:
		__antithesis_instrumentation__.Notify(563031)
		return n.reqOrdering
	case *unionNode:
		__antithesis_instrumentation__.Notify(563032)
		return n.reqOrdering
	case *insertNode, *insertFastPathNode:
		__antithesis_instrumentation__.Notify(563033)

	case *updateNode, *upsertNode:
		__antithesis_instrumentation__.Notify(563034)

	case *scanNode:
		__antithesis_instrumentation__.Notify(563035)
		return n.reqOrdering
	case *ordinalityNode:
		__antithesis_instrumentation__.Notify(563036)
		return n.reqOrdering
	case *renderNode:
		__antithesis_instrumentation__.Notify(563037)
		return n.reqOrdering
	case *sortNode:
		__antithesis_instrumentation__.Notify(563038)
		return n.ordering
	case *topKNode:
		__antithesis_instrumentation__.Notify(563039)
		return n.ordering
	case *lookupJoinNode:
		__antithesis_instrumentation__.Notify(563040)
		return n.reqOrdering
	case *invertedJoinNode:
		__antithesis_instrumentation__.Notify(563041)
		return n.reqOrdering
	case *zigzagJoinNode:
		__antithesis_instrumentation__.Notify(563042)
		return n.reqOrdering
	}
	__antithesis_instrumentation__.Notify(563019)

	return nil
}
