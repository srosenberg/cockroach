package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type grpcServer struct {
	*grpc.Server
	mode serveMode
}

func newGRPCServer(rpcCtx *rpc.Context) *grpcServer {
	__antithesis_instrumentation__.Notify(193469)
	s := &grpcServer{}
	s.mode.set(modeInitializing)
	s.Server = rpc.NewServer(rpcCtx, rpc.WithInterceptor(func(path string) error {
		__antithesis_instrumentation__.Notify(193471)
		return s.intercept(path)
	}))
	__antithesis_instrumentation__.Notify(193470)
	return s
}

type serveMode int32

const (
	modeInitializing serveMode = iota

	modeOperational

	modeDraining
)

func (s *grpcServer) setMode(mode serveMode) {
	__antithesis_instrumentation__.Notify(193472)
	s.mode.set(mode)
}

func (s *grpcServer) operational() bool {
	__antithesis_instrumentation__.Notify(193473)
	sMode := s.mode.get()
	return sMode == modeOperational || func() bool {
		__antithesis_instrumentation__.Notify(193474)
		return sMode == modeDraining == true
	}() == true
}

var rpcsAllowedWhileBootstrapping = map[string]struct{}{
	"/cockroach.rpc.Heartbeat/Ping":             {},
	"/cockroach.gossip.Gossip/Gossip":           {},
	"/cockroach.server.serverpb.Init/Bootstrap": {},
	"/cockroach.server.serverpb.Admin/Health":   {},
}

func (s *grpcServer) intercept(fullName string) error {
	__antithesis_instrumentation__.Notify(193475)
	if s.operational() {
		__antithesis_instrumentation__.Notify(193478)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(193479)
	}
	__antithesis_instrumentation__.Notify(193476)
	if _, allowed := rpcsAllowedWhileBootstrapping[fullName]; !allowed {
		__antithesis_instrumentation__.Notify(193480)
		return s.waitingForInitError(fullName)
	} else {
		__antithesis_instrumentation__.Notify(193481)
	}
	__antithesis_instrumentation__.Notify(193477)
	return nil
}

func (s *serveMode) set(mode serveMode) {
	__antithesis_instrumentation__.Notify(193482)
	atomic.StoreInt32((*int32)(s), int32(mode))
}

func (s *serveMode) get() serveMode {
	__antithesis_instrumentation__.Notify(193483)
	return serveMode(atomic.LoadInt32((*int32)(s)))
}

func (s *grpcServer) waitingForInitError(methodName string) error {
	__antithesis_instrumentation__.Notify(193484)
	return grpcstatus.Errorf(codes.Unavailable, "node waiting for init; %s not available", methodName)
}

func IsWaitingForInit(err error) bool {
	__antithesis_instrumentation__.Notify(193485)
	s, ok := grpcstatus.FromError(errors.UnwrapAll(err))
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(193486)
		return s.Code() == codes.Unavailable == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(193487)
		return strings.Contains(err.Error(), "node waiting for init") == true
	}() == true
}
