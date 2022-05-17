package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/sslocal"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
)

type Controller struct {
	*sslocal.Controller
	db *kv.DB
	ie sqlutil.InternalExecutor
	st *cluster.Settings
}

func NewController(
	sqlStats *PersistedSQLStats,
	status serverpb.SQLStatusServer,
	db *kv.DB,
	ie sqlutil.InternalExecutor,
) *Controller {
	__antithesis_instrumentation__.Notify(624675)
	return &Controller{
		Controller: sslocal.NewController(sqlStats.SQLStats, status),
		db:         db,
		ie:         ie,
		st:         sqlStats.cfg.Settings,
	}
}

func (s *Controller) CreateSQLStatsCompactionSchedule(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(624676)
	return s.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624677)
		_, err := CreateSQLStatsCompactionScheduleIfNotYetExist(ctx, s.ie, txn, s.st)
		return err
	})
}

func (s *Controller) ResetClusterSQLStats(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(624678)
	if err := s.Controller.ResetClusterSQLStats(ctx); err != nil {
		__antithesis_instrumentation__.Notify(624682)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624683)
	}
	__antithesis_instrumentation__.Notify(624679)

	resetSysTableStats := func(tableName string) error {
		__antithesis_instrumentation__.Notify(624684)
		if _, err := s.ie.ExecEx(
			ctx,
			"reset-sql-stats",
			nil,
			sessiondata.InternalExecutorOverride{
				User: security.NodeUserName(),
			},
			"TRUNCATE "+tableName); err != nil {
			__antithesis_instrumentation__.Notify(624686)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624687)
		}
		__antithesis_instrumentation__.Notify(624685)

		return nil
	}
	__antithesis_instrumentation__.Notify(624680)
	if err := resetSysTableStats("system.statement_statistics"); err != nil {
		__antithesis_instrumentation__.Notify(624688)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624689)
	}
	__antithesis_instrumentation__.Notify(624681)

	return resetSysTableStats("system.transaction_statistics")
}
