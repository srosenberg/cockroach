package idxusage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
)

type Controller struct {
	statusServer serverpb.SQLStatusServer
}

func NewController(status serverpb.SQLStatusServer) *Controller {
	__antithesis_instrumentation__.Notify(492994)
	return &Controller{
		statusServer: status,
	}
}

func (s *Controller) ResetIndexUsageStats(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(492995)
	req := &serverpb.ResetIndexUsageStatsRequest{}
	_, err := s.statusServer.ResetIndexUsageStats(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(492997)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492998)
	}
	__antithesis_instrumentation__.Notify(492996)
	return nil
}
