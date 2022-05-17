package grpcutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/gogo/protobuf/types"
)

type TestServerImpl struct {
	UU func(context.Context, *types.Any) (*types.Any, error)
	US func(*types.Any, GRPCTest_UnaryStreamServer) error
	SU func(server GRPCTest_StreamUnaryServer) error
	SS func(server GRPCTest_StreamStreamServer) error
}

var _ GRPCTestServer = (*TestServerImpl)(nil)

func (s *TestServerImpl) UnaryUnary(ctx context.Context, any *types.Any) (*types.Any, error) {
	__antithesis_instrumentation__.Notify(644248)
	return s.UU(ctx, any)
}

func (s *TestServerImpl) UnaryStream(any *types.Any, server GRPCTest_UnaryStreamServer) error {
	__antithesis_instrumentation__.Notify(644249)
	return s.US(any, server)
}

func (s *TestServerImpl) StreamUnary(server GRPCTest_StreamUnaryServer) error {
	__antithesis_instrumentation__.Notify(644250)
	return s.SU(server)
}

func (s *TestServerImpl) StreamStream(server GRPCTest_StreamStreamServer) error {
	__antithesis_instrumentation__.Notify(644251)
	return s.SS(server)
}
