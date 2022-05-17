package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"google.golang.org/grpc"
)

type ContextTestingKnobs struct {
	UnaryClientInterceptor func(target string, class ConnectionClass) grpc.UnaryClientInterceptor

	StreamClientInterceptor func(target string, class ConnectionClass) grpc.StreamClientInterceptor

	ArtificialLatencyMap map[string]int

	StorageClusterID *uuid.UUID
}

func NewInsecureTestingContext(
	ctx context.Context, clock *hlc.Clock, stopper *stop.Stopper,
) *Context {
	__antithesis_instrumentation__.Notify(184717)
	clusterID := uuid.MakeV4()
	return NewInsecureTestingContextWithClusterID(ctx, clock, stopper, clusterID)
}

func NewInsecureTestingContextWithClusterID(
	ctx context.Context, clock *hlc.Clock, stopper *stop.Stopper, storageClusterID uuid.UUID,
) *Context {
	__antithesis_instrumentation__.Notify(184718)
	return NewInsecureTestingContextWithKnobs(ctx,
		clock, stopper, ContextTestingKnobs{
			StorageClusterID: &storageClusterID,
		})
}

func NewInsecureTestingContextWithKnobs(
	ctx context.Context, clock *hlc.Clock, stopper *stop.Stopper, knobs ContextTestingKnobs,
) *Context {
	__antithesis_instrumentation__.Notify(184719)
	return NewContext(ctx,
		ContextOptions{
			TenantID: roachpb.SystemTenantID,
			Config:   &base.Config{Insecure: true},
			Clock:    clock,
			Stopper:  stopper,
			Settings: cluster.MakeTestingClusterSettings(),
			Knobs:    knobs,
		})
}
