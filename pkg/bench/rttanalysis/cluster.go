package rttanalysis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"sync"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type ClusterConstructor func(testing.TB) *Cluster

func MakeClusterConstructor(
	f func(testing.TB, base.TestingKnobs) (_ *gosql.DB, cleanup func()),
) ClusterConstructor {
	__antithesis_instrumentation__.Notify(1824)
	return func(t testing.TB) *Cluster {
		__antithesis_instrumentation__.Notify(1825)
		c := &Cluster{}
		beforePlan := func(trace tracing.Recording, stmt string) {
			__antithesis_instrumentation__.Notify(1827)
			if _, ok := c.stmtToKVBatchRequests.Load(stmt); ok {
				__antithesis_instrumentation__.Notify(1828)
				c.stmtToKVBatchRequests.Store(stmt, trace)
			} else {
				__antithesis_instrumentation__.Notify(1829)
			}
		}
		__antithesis_instrumentation__.Notify(1826)
		c.sql, c.cleanup = f(t, base.TestingKnobs{
			SQLExecutor: &sql.ExecutorTestingKnobs{
				WithStatementTrace: beforePlan,
			},
		})
		return c
	}
}

type Cluster struct {
	stmtToKVBatchRequests sync.Map
	cleanup               func()
	sql                   *gosql.DB
}

func (c *Cluster) conn() *gosql.DB {
	__antithesis_instrumentation__.Notify(1830)
	return c.sql
}

func (c *Cluster) clearStatementTrace(stmt string) {
	__antithesis_instrumentation__.Notify(1831)
	c.stmtToKVBatchRequests.Store(stmt, nil)
}

func (c *Cluster) getStatementTrace(stmt string) (tracing.Recording, bool) {
	__antithesis_instrumentation__.Notify(1832)
	out, _ := c.stmtToKVBatchRequests.Load(stmt)
	r, ok := out.(tracing.Recording)
	return r, ok
}

func (c *Cluster) close() {
	__antithesis_instrumentation__.Notify(1833)
	c.cleanup()
}
