package tenantcostserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

func (s *instance) ReconfigureTokenBucket(
	ctx context.Context,
	txn *kv.Txn,
	tenantID roachpb.TenantID,
	availableRU float64,
	refillRate float64,
	maxBurstRU float64,
	asOf time.Time,
	asOfConsumedRequestUnits float64,
) error {
	__antithesis_instrumentation__.Notify(20199)
	if err := s.checkTenantID(ctx, txn, tenantID); err != nil {
		__antithesis_instrumentation__.Notify(20203)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20204)
	}
	__antithesis_instrumentation__.Notify(20200)
	h := makeSysTableHelper(ctx, s.executor, txn, tenantID)
	state, err := h.readTenantState()
	if err != nil {
		__antithesis_instrumentation__.Notify(20205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20206)
	}
	__antithesis_instrumentation__.Notify(20201)
	now := s.timeSource.Now()
	state.update(now)
	state.Bucket.Reconfigure(
		ctx, tenantID, availableRU, refillRate, maxBurstRU, asOf, asOfConsumedRequestUnits,
		now, state.Consumption.RU,
	)
	if err := h.updateTenantState(state); err != nil {
		__antithesis_instrumentation__.Notify(20207)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20208)
	}
	__antithesis_instrumentation__.Notify(20202)
	return nil
}

func (s *instance) checkTenantID(
	ctx context.Context, txn *kv.Txn, tenantID roachpb.TenantID,
) error {
	__antithesis_instrumentation__.Notify(20209)
	row, err := s.executor.QueryRowEx(
		ctx, "check-tenant", txn, sessiondata.NodeUserSessionDataOverride,
		`SELECT active FROM system.tenants WHERE id = $1`, tenantID.ToUint64(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20213)
		return err
	} else {
		__antithesis_instrumentation__.Notify(20214)
	}
	__antithesis_instrumentation__.Notify(20210)
	if row == nil {
		__antithesis_instrumentation__.Notify(20215)
		return pgerror.Newf(pgcode.UndefinedObject, "tenant %q does not exist", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(20216)
	}
	__antithesis_instrumentation__.Notify(20211)
	if active := *row[0].(*tree.DBool); !active {
		__antithesis_instrumentation__.Notify(20217)
		return errors.Errorf("tenant %q is not active", tenantID)
	} else {
		__antithesis_instrumentation__.Notify(20218)
	}
	__antithesis_instrumentation__.Notify(20212)
	return nil
}
