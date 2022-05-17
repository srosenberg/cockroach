package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type Controller struct {
	sqlStats     *SQLStats
	statusServer serverpb.SQLStatusServer
}

func NewController(sqlStats *SQLStats, status serverpb.SQLStatusServer) *Controller {
	__antithesis_instrumentation__.Notify(625373)
	return &Controller{
		sqlStats:     sqlStats,
		statusServer: status,
	}
}

func (s *Controller) ResetClusterSQLStats(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(625374)
	req := &serverpb.ResetSQLStatsRequest{

		ResetPersistedStats: false,
	}
	_, err := s.statusServer.ResetSQLStats(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(625376)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625377)
	}
	__antithesis_instrumentation__.Notify(625375)
	return nil
}

func (s *Controller) ResetLocalSQLStats(ctx context.Context) {
	__antithesis_instrumentation__.Notify(625378)
	err := s.sqlStats.Reset(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(625379)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(625380)
			log.Warningf(ctx, "reported SQL stats memory limit has been exceeded, some fingerprints stats are discarded: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(625381)
		}
	} else {
		__antithesis_instrumentation__.Notify(625382)
	}
}
