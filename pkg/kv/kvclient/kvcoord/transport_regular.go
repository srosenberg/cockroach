//go:build !race
// +build !race

package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/rpc/nodedialer"

func GRPCTransportFactory(
	opts SendOptions, nodeDialer *nodedialer.Dialer, replicas ReplicaSlice,
) (Transport, error) {
	__antithesis_instrumentation__.Notify(88105)
	return grpcTransportFactoryImpl(opts, nodeDialer, replicas)
}
