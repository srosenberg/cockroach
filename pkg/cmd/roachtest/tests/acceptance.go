package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

func registerAcceptance(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45217)
	testCases := map[registry.Owner][]struct {
		name            string
		fn              func(ctx context.Context, t test.Test, c cluster.Cluster)
		skip            string
		minVersion      string
		numNodes        int
		timeout         time.Duration
		encryptAtRandom bool
	}{
		registry.OwnerKV: {
			{name: "decommission-self", fn: runDecommissionSelf},
			{name: "event-log", fn: runEventLog},
			{name: "gossip/peerings", fn: runGossipPeerings},
			{name: "gossip/restart", fn: runGossipRestart},
			{
				name: "gossip/restart-node-one",
				fn:   runGossipRestartNodeOne,
			},
			{name: "gossip/locality-address", fn: runCheckLocalityIPAddress},
			{
				name:       "multitenant",
				minVersion: "v20.2.0",
				fn:         runAcceptanceMultitenant,
			},
			{name: "reset-quorum", fn: runResetQuorum, numNodes: 8},
			{
				name: "many-splits", fn: runManySplits,
				minVersion:      "v19.2.0",
				encryptAtRandom: true,
			},
			{
				name: "version-upgrade",
				fn: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(45219)
					runVersionUpgrade(ctx, t, c)
				},

				minVersion: "v19.2.0",
				timeout:    30 * time.Minute,
			},
		},
		registry.OwnerServer: {
			{name: "build-info", fn: RunBuildInfo},
			{name: "build-analyze", fn: RunBuildAnalyze},
			{name: "cli/node-status", fn: runCLINodeStatus},
			{name: "cluster-init", fn: runClusterInit},
			{name: "rapid-restart", fn: runRapidRestart},
			{name: "status-server", fn: runStatusServer},
		},
	}
	__antithesis_instrumentation__.Notify(45218)
	tags := []string{"default", "quick"}
	specTemplate := registry.TestSpec{

		Name:    "acceptance",
		Timeout: 10 * time.Minute,
		Tags:    tags,
	}

	for owner, tests := range testCases {
		__antithesis_instrumentation__.Notify(45220)
		for _, tc := range tests {
			__antithesis_instrumentation__.Notify(45221)
			tc := tc
			numNodes := 4
			if tc.numNodes != 0 {
				__antithesis_instrumentation__.Notify(45225)
				numNodes = tc.numNodes
			} else {
				__antithesis_instrumentation__.Notify(45226)
			}
			__antithesis_instrumentation__.Notify(45222)

			spec := specTemplate
			spec.Owner = owner
			spec.Cluster = r.MakeClusterSpec(numNodes)
			spec.Skip = tc.skip
			spec.Name = specTemplate.Name + "/" + tc.name
			if tc.timeout != 0 {
				__antithesis_instrumentation__.Notify(45227)
				spec.Timeout = tc.timeout
			} else {
				__antithesis_instrumentation__.Notify(45228)
			}
			__antithesis_instrumentation__.Notify(45223)
			spec.EncryptAtRandom = tc.encryptAtRandom
			spec.Run = func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(45229)
				tc.fn(ctx, t, c)
			}
			__antithesis_instrumentation__.Notify(45224)
			r.Add(spec)
		}
	}
}
