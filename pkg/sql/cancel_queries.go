package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type cancelQueriesNode struct {
	rows     planNode
	ifExists bool
}

func (n *cancelQueriesNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(247175)
	return nil
}

func (n *cancelQueriesNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(247176)

	if ok, err := n.rows.Next(params); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(247183)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(247184)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(247185)
	}
	__antithesis_instrumentation__.Notify(247177)

	datum := n.rows.Values()[0]
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(247186)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(247187)
	}
	__antithesis_instrumentation__.Notify(247178)

	queryIDString, ok := tree.AsDString(datum)
	if !ok {
		__antithesis_instrumentation__.Notify(247188)
		return false, errors.AssertionFailedf("%q: expected *DString, found %T", datum, datum)
	} else {
		__antithesis_instrumentation__.Notify(247189)
	}
	__antithesis_instrumentation__.Notify(247179)

	queryID, err := StringToClusterWideID(string(queryIDString))
	if err != nil {
		__antithesis_instrumentation__.Notify(247190)
		return false, pgerror.Wrapf(err, pgcode.Syntax, "invalid query ID %s", datum)
	} else {
		__antithesis_instrumentation__.Notify(247191)
	}
	__antithesis_instrumentation__.Notify(247180)

	nodeID := 0xFFFFFFFF & queryID.Lo

	request := &serverpb.CancelQueryRequest{
		NodeId:   fmt.Sprintf("%d", nodeID),
		QueryID:  string(queryIDString),
		Username: params.SessionData().User().Normalized(),
	}

	response, err := params.extendedEvalCtx.SQLStatusServer.CancelQuery(params.ctx, request)
	if err != nil {
		__antithesis_instrumentation__.Notify(247192)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247193)
	}
	__antithesis_instrumentation__.Notify(247181)

	if !response.Canceled && func() bool {
		__antithesis_instrumentation__.Notify(247194)
		return !n.ifExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(247195)
		return false, errors.Newf("could not cancel query %s: %s", queryID, response.Error)
	} else {
		__antithesis_instrumentation__.Notify(247196)
	}
	__antithesis_instrumentation__.Notify(247182)

	return true, nil
}

func (*cancelQueriesNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(247197)
	return nil
}

func (n *cancelQueriesNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(247198)
	n.rows.Close(ctx)
}
