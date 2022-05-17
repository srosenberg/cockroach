package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

var errTLSInfoMissing = authError("TLSInfo is not available in request context")

func authError(msg string) error {
	__antithesis_instrumentation__.Notify(183984)
	return status.Error(codes.Unauthenticated, msg)
}

func authErrorf(format string, a ...interface{}) error {
	__antithesis_instrumentation__.Notify(183985)
	return status.Errorf(codes.Unauthenticated, format, a...)
}

type kvAuth struct {
	tenant tenantAuthorizer
}

func (a kvAuth) AuthUnary() grpc.UnaryServerInterceptor {
	__antithesis_instrumentation__.Notify(183986)
	return a.unaryInterceptor
}
func (a kvAuth) AuthStream() grpc.StreamServerInterceptor {
	__antithesis_instrumentation__.Notify(183987)
	return a.streamInterceptor
}

func (a kvAuth) unaryInterceptor(
	ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler,
) (interface{}, error) {
	__antithesis_instrumentation__.Notify(183988)

	if info.FullMethod == "/cockroach.server.serverpb.Admin/RequestCA" {
		__antithesis_instrumentation__.Notify(183993)
		return handler(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(183994)
	}
	__antithesis_instrumentation__.Notify(183989)

	if info.FullMethod == "/cockroach.server.serverpb.Admin/RequestCertBundle" {
		__antithesis_instrumentation__.Notify(183995)
		return handler(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(183996)
	}
	__antithesis_instrumentation__.Notify(183990)

	tenID, err := a.authenticate(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(183997)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(183998)
	}
	__antithesis_instrumentation__.Notify(183991)
	if tenID != (roachpb.TenantID{}) {
		__antithesis_instrumentation__.Notify(183999)
		ctx = contextWithTenant(ctx, tenID)
		if err := a.tenant.authorize(tenID, info.FullMethod, req); err != nil {
			__antithesis_instrumentation__.Notify(184000)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(184001)
		}
	} else {
		__antithesis_instrumentation__.Notify(184002)
	}
	__antithesis_instrumentation__.Notify(183992)
	return handler(ctx, req)
}

func (a kvAuth) streamInterceptor(
	srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler,
) error {
	__antithesis_instrumentation__.Notify(184003)
	ctx := ss.Context()
	tenID, err := a.authenticate(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(184006)
		return err
	} else {
		__antithesis_instrumentation__.Notify(184007)
	}
	__antithesis_instrumentation__.Notify(184004)
	if tenID != (roachpb.TenantID{}) {
		__antithesis_instrumentation__.Notify(184008)
		ctx = contextWithTenant(ctx, tenID)
		origSS := ss
		ss = &wrappedServerStream{
			ServerStream: origSS,
			ctx:          ctx,
			recv: func(m interface{}) error {
				__antithesis_instrumentation__.Notify(184009)
				if err := origSS.RecvMsg(m); err != nil {
					__antithesis_instrumentation__.Notify(184011)
					return err
				} else {
					__antithesis_instrumentation__.Notify(184012)
				}
				__antithesis_instrumentation__.Notify(184010)

				return a.tenant.authorize(tenID, info.FullMethod, m)
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(184013)
	}
	__antithesis_instrumentation__.Notify(184005)
	return handler(srv, ss)
}

func (a kvAuth) authenticate(ctx context.Context) (roachpb.TenantID, error) {
	__antithesis_instrumentation__.Notify(184014)
	if grpcutil.IsLocalRequestContext(ctx) {
		__antithesis_instrumentation__.Notify(184021)

		return roachpb.TenantID{}, nil
	} else {
		__antithesis_instrumentation__.Notify(184022)
	}
	__antithesis_instrumentation__.Notify(184015)

	p, ok := peer.FromContext(ctx)
	if !ok {
		__antithesis_instrumentation__.Notify(184023)
		return roachpb.TenantID{}, errTLSInfoMissing
	} else {
		__antithesis_instrumentation__.Notify(184024)
	}
	__antithesis_instrumentation__.Notify(184016)

	tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(184025)
		return len(tlsInfo.State.PeerCertificates) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(184026)
		return roachpb.TenantID{}, errTLSInfoMissing
	} else {
		__antithesis_instrumentation__.Notify(184027)
	}
	__antithesis_instrumentation__.Notify(184017)

	certUsers, err := security.GetCertificateUsers(&tlsInfo.State)
	if err != nil {
		__antithesis_instrumentation__.Notify(184028)
		return roachpb.TenantID{}, err
	} else {
		__antithesis_instrumentation__.Notify(184029)
	}
	__antithesis_instrumentation__.Notify(184018)

	clientCert := tlsInfo.State.PeerCertificates[0]
	if a.tenant.tenantID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(184030)

		if security.IsTenantCertificate(clientCert) {
			__antithesis_instrumentation__.Notify(184031)

			return tenantFromCommonName(clientCert.Subject.CommonName)
		} else {
			__antithesis_instrumentation__.Notify(184032)
		}
	} else {
		__antithesis_instrumentation__.Notify(184033)

		if security.IsTenantCertificate(clientCert) {
			__antithesis_instrumentation__.Notify(184034)

			return roachpb.TenantID{}, nil
		} else {
			__antithesis_instrumentation__.Notify(184035)
		}
	}
	__antithesis_instrumentation__.Notify(184019)

	if !security.Contains(certUsers, security.NodeUser) && func() bool {
		__antithesis_instrumentation__.Notify(184036)
		return !security.Contains(certUsers, security.RootUser) == true
	}() == true {
		__antithesis_instrumentation__.Notify(184037)
		return roachpb.TenantID{}, authErrorf("user %s is not allowed to perform this RPC", certUsers)
	} else {
		__antithesis_instrumentation__.Notify(184038)
	}
	__antithesis_instrumentation__.Notify(184020)

	return roachpb.TenantID{}, nil
}
