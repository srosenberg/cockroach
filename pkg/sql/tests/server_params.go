package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
)

func CreateTestServerParams() (base.TestServerArgs, *CommandFilters) {
	__antithesis_instrumentation__.Notify(628377)
	var cmdFilters CommandFilters
	cmdFilters.AppendFilter(CheckEndTxnTrigger, true)
	params := base.TestServerArgs{}
	params.Knobs = CreateTestingKnobs()
	params.Knobs.Store = &kvserver.StoreTestingKnobs{
		EvalKnobs: kvserverbase.BatchEvalTestingKnobs{
			TestingEvalFilter: cmdFilters.RunFilters,
		},
	}
	return params, &cmdFilters
}

func CreateTestTenantParams(tenantID roachpb.TenantID) base.TestTenantArgs {
	__antithesis_instrumentation__.Notify(628378)
	return base.TestTenantArgs{
		TenantID:     tenantID,
		TestingKnobs: CreateTestingKnobs(),
	}
}

func CreateTestingKnobs() base.TestingKnobs {
	__antithesis_instrumentation__.Notify(628379)
	return base.TestingKnobs{
		SQLStatsKnobs: &sqlstats.TestingKnobs{
			AOSTClause: "AS OF SYSTEM TIME '-1us'",
		},
	}
}
