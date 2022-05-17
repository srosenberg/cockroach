package serverutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

type TestClusterInterface interface {
	Start(t testing.TB)

	NumServers() int

	Server(idx int) TestServerInterface

	ServerConn(idx int) *gosql.DB

	StopServer(idx int)

	Stopper() *stop.Stopper

	AddVoters(
		startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) (roachpb.RangeDescriptor, error)

	AddVotersMulti(
		kts ...KeyAndTargets,
	) ([]roachpb.RangeDescriptor, []error)

	AddVotersOrFatal(
		t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) roachpb.RangeDescriptor

	RemoveVoters(
		startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) (roachpb.RangeDescriptor, error)

	RemoveVotersOrFatal(
		t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) roachpb.RangeDescriptor

	AddNonVoters(
		startKey roachpb.Key,
		targets ...roachpb.ReplicationTarget,
	) (roachpb.RangeDescriptor, error)

	AddNonVotersOrFatal(
		t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) roachpb.RangeDescriptor

	RemoveNonVoters(
		startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) (roachpb.RangeDescriptor, error)

	RemoveNonVotersOrFatal(
		t testing.TB, startKey roachpb.Key, targets ...roachpb.ReplicationTarget,
	) roachpb.RangeDescriptor

	SwapVoterWithNonVoter(
		startKey roachpb.Key, voterTarget, nonVoterTarget roachpb.ReplicationTarget,
	) (*roachpb.RangeDescriptor, error)

	SwapVoterWithNonVoterOrFatal(
		t *testing.T, startKey roachpb.Key, voterTarget, nonVoterTarget roachpb.ReplicationTarget,
	) *roachpb.RangeDescriptor

	FindRangeLeaseHolder(
		rangeDesc roachpb.RangeDescriptor,
		hint *roachpb.ReplicationTarget,
	) (roachpb.ReplicationTarget, error)

	TransferRangeLease(
		rangeDesc roachpb.RangeDescriptor, dest roachpb.ReplicationTarget,
	) error

	MoveRangeLeaseNonCooperatively(
		ctx context.Context,
		rangeDesc roachpb.RangeDescriptor,
		dest roachpb.ReplicationTarget,
		manual *hlc.HybridManualClock,
	) (*roachpb.Lease, error)

	LookupRange(key roachpb.Key) (roachpb.RangeDescriptor, error)

	LookupRangeOrFatal(t testing.TB, key roachpb.Key) roachpb.RangeDescriptor

	Target(serverIdx int) roachpb.ReplicationTarget

	ReplicationMode() base.TestClusterReplicationMode

	ScratchRange(t testing.TB) roachpb.Key

	WaitForFullReplication() error
}

type TestClusterFactory interface {
	NewTestCluster(t testing.TB, numNodes int, args base.TestClusterArgs) TestClusterInterface
}

var clusterFactoryImpl TestClusterFactory

func InitTestClusterFactory(impl TestClusterFactory) {
	__antithesis_instrumentation__.Notify(645910)
	clusterFactoryImpl = impl
}

func StartNewTestCluster(
	t testing.TB, numNodes int, args base.TestClusterArgs,
) TestClusterInterface {
	__antithesis_instrumentation__.Notify(645911)
	cluster := NewTestCluster(t, numNodes, args)
	cluster.Start(t)
	return cluster
}

func NewTestCluster(t testing.TB, numNodes int, args base.TestClusterArgs) TestClusterInterface {
	__antithesis_instrumentation__.Notify(645912)
	if clusterFactoryImpl == nil {
		__antithesis_instrumentation__.Notify(645914)
		panic("TestClusterFactory not initialized. One needs to be injected " +
			"from the package's TestMain()")
	} else {
		__antithesis_instrumentation__.Notify(645915)
	}
	__antithesis_instrumentation__.Notify(645913)
	return clusterFactoryImpl.NewTestCluster(t, numNodes, args)
}

type KeyAndTargets struct {
	StartKey roachpb.Key
	Targets  []roachpb.ReplicationTarget
}
