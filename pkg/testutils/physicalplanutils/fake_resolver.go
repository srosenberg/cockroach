package physicalplanutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/util"
)

func FakeResolverForTestCluster(tc serverutils.TestClusterInterface) physicalplan.SpanResolver {
	__antithesis_instrumentation__.Notify(645900)
	nodeDescs := make([]*roachpb.NodeDescriptor, tc.NumServers())
	for i := range nodeDescs {
		__antithesis_instrumentation__.Notify(645902)
		s := tc.Server(i)
		nodeDescs[i] = &roachpb.NodeDescriptor{
			NodeID:  s.NodeID(),
			Address: util.UnresolvedAddr{AddressField: s.ServingRPCAddr()},
		}
	}
	__antithesis_instrumentation__.Notify(645901)

	return physicalplan.NewFakeSpanResolver(nodeDescs)
}
