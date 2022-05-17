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

type cancelSessionsNode struct {
	rows     planNode
	ifExists bool
}

func (n *cancelSessionsNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(247199)
	return nil
}

func (n *cancelSessionsNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(247200)

	if ok, err := n.rows.Next(params); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(247207)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(247208)
		return ok, err
	} else {
		__antithesis_instrumentation__.Notify(247209)
	}
	__antithesis_instrumentation__.Notify(247201)

	datum := n.rows.Values()[0]
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(247210)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(247211)
	}
	__antithesis_instrumentation__.Notify(247202)

	sessionIDString, ok := tree.AsDString(datum)
	if !ok {
		__antithesis_instrumentation__.Notify(247212)
		return false, errors.AssertionFailedf("%q: expected *DString, found %T", datum, datum)
	} else {
		__antithesis_instrumentation__.Notify(247213)
	}
	__antithesis_instrumentation__.Notify(247203)

	sessionID, err := StringToClusterWideID(string(sessionIDString))
	if err != nil {
		__antithesis_instrumentation__.Notify(247214)
		return false, pgerror.Wrapf(err, pgcode.Syntax, "invalid session ID %s", datum)
	} else {
		__antithesis_instrumentation__.Notify(247215)
	}
	__antithesis_instrumentation__.Notify(247204)

	nodeID := sessionID.GetNodeID()

	request := &serverpb.CancelSessionRequest{
		NodeId:    fmt.Sprintf("%d", nodeID),
		SessionID: sessionID.GetBytes(),
		Username:  params.SessionData().User().Normalized(),
	}

	response, err := params.extendedEvalCtx.SQLStatusServer.CancelSession(params.ctx, request)
	if err != nil {
		__antithesis_instrumentation__.Notify(247216)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(247217)
	}
	__antithesis_instrumentation__.Notify(247205)

	if !response.Canceled && func() bool {
		__antithesis_instrumentation__.Notify(247218)
		return !n.ifExists == true
	}() == true {
		__antithesis_instrumentation__.Notify(247219)
		return false, errors.Newf("could not cancel session %s: %s", sessionID, response.Error)
	} else {
		__antithesis_instrumentation__.Notify(247220)
	}
	__antithesis_instrumentation__.Notify(247206)

	return true, nil
}

func (*cancelSessionsNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(247221)
	return nil
}

func (n *cancelSessionsNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(247222)
	n.rows.Close(ctx)
}
