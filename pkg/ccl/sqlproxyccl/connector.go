package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/balancer"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenant"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/throttler"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	pgproto3 "github.com/jackc/pgproto3/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const sessionRevivalTokenStartupParam = "crdb:session_revival_token_base64"

const remoteAddrStartupParam = "crdb:remote_addr"

type connector struct {
	ClusterName string
	TenantID    roachpb.TenantID

	DirectoryCache tenant.DirectoryCache

	Balancer *balancer.Balancer

	StartupMsg *pgproto3.StartupMessage

	TLSConfig *tls.Config

	IdleMonitorWrapperFn func(serverConn net.Conn) net.Conn

	testingKnobs struct {
		dialTenantCluster func(ctx context.Context) (net.Conn, error)
		lookupAddr        func(ctx context.Context) (string, error)
		dialSQLServer     func(serverAddr string) (net.Conn, error)
	}
}

func (c *connector) OpenTenantConnWithToken(
	ctx context.Context, token string,
) (retServerConn net.Conn, retErr error) {
	__antithesis_instrumentation__.Notify(21320)
	c.StartupMsg.Parameters[sessionRevivalTokenStartupParam] = token
	defer func() {
		__antithesis_instrumentation__.Notify(21326)

		delete(c.StartupMsg.Parameters, sessionRevivalTokenStartupParam)
	}()
	__antithesis_instrumentation__.Notify(21321)

	serverConn, err := c.dialTenantCluster(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(21327)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21328)
	}
	__antithesis_instrumentation__.Notify(21322)
	defer func() {
		__antithesis_instrumentation__.Notify(21329)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(21330)
			serverConn.Close()
		} else {
			__antithesis_instrumentation__.Notify(21331)
		}
	}()
	__antithesis_instrumentation__.Notify(21323)

	if c.IdleMonitorWrapperFn != nil {
		__antithesis_instrumentation__.Notify(21332)
		serverConn = c.IdleMonitorWrapperFn(serverConn)
	} else {
		__antithesis_instrumentation__.Notify(21333)
	}
	__antithesis_instrumentation__.Notify(21324)

	if err := readTokenAuthResult(serverConn); err != nil {
		__antithesis_instrumentation__.Notify(21334)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21335)
	}
	__antithesis_instrumentation__.Notify(21325)
	log.Infof(ctx, "connected to %s through token-based auth", serverConn.RemoteAddr())
	return serverConn, nil
}

func (c *connector) OpenTenantConnWithAuth(
	ctx context.Context, clientConn net.Conn, throttleHook func(throttler.AttemptStatus) error,
) (retServerConn net.Conn, sentToClient bool, retErr error) {
	__antithesis_instrumentation__.Notify(21336)

	delete(c.StartupMsg.Parameters, sessionRevivalTokenStartupParam)

	serverConn, err := c.dialTenantCluster(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(21341)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(21342)
	}
	__antithesis_instrumentation__.Notify(21337)
	defer func() {
		__antithesis_instrumentation__.Notify(21343)
		if retErr != nil {
			__antithesis_instrumentation__.Notify(21344)
			serverConn.Close()
		} else {
			__antithesis_instrumentation__.Notify(21345)
		}
	}()
	__antithesis_instrumentation__.Notify(21338)

	if c.IdleMonitorWrapperFn != nil {
		__antithesis_instrumentation__.Notify(21346)
		serverConn = c.IdleMonitorWrapperFn(serverConn)
	} else {
		__antithesis_instrumentation__.Notify(21347)
	}
	__antithesis_instrumentation__.Notify(21339)

	if err := authenticate(clientConn, serverConn, throttleHook); err != nil {
		__antithesis_instrumentation__.Notify(21348)
		return nil, true, err
	} else {
		__antithesis_instrumentation__.Notify(21349)
	}
	__antithesis_instrumentation__.Notify(21340)
	log.Infof(ctx, "connected to %s through normal auth", serverConn.RemoteAddr())
	return serverConn, false, nil
}

func (c *connector) dialTenantCluster(ctx context.Context) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(21350)
	if c.testingKnobs.dialTenantCluster != nil {
		__antithesis_instrumentation__.Notify(21354)
		return c.testingKnobs.dialTenantCluster(ctx)
	} else {
		__antithesis_instrumentation__.Notify(21355)
	}
	__antithesis_instrumentation__.Notify(21351)

	retryOpts := retry.Options{
		InitialBackoff: 10 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
	}

	lookupAddrErr := log.Every(time.Minute)
	dialSQLServerErr := log.Every(time.Minute)
	reportFailureErr := log.Every(time.Minute)
	var lookupAddrErrs, dialSQLServerErrs, reportFailureErrs int
	var crdbConn net.Conn
	var serverAddr string
	var err error

	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(21356)

		serverAddr, err = c.lookupAddr(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(21359)
			if isRetriableConnectorError(err) {
				__antithesis_instrumentation__.Notify(21361)
				lookupAddrErrs++
				if lookupAddrErr.ShouldLog() {
					__antithesis_instrumentation__.Notify(21363)
					log.Ops.Errorf(ctx, "lookup address (%d errors skipped): %v",
						lookupAddrErrs, err)
					lookupAddrErrs = 0
				} else {
					__antithesis_instrumentation__.Notify(21364)
				}
				__antithesis_instrumentation__.Notify(21362)
				continue
			} else {
				__antithesis_instrumentation__.Notify(21365)
			}
			__antithesis_instrumentation__.Notify(21360)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21366)
		}
		__antithesis_instrumentation__.Notify(21357)

		crdbConn, err = c.dialSQLServer(serverAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(21367)
			if isRetriableConnectorError(err) {
				__antithesis_instrumentation__.Notify(21369)
				dialSQLServerErrs++
				if dialSQLServerErr.ShouldLog() {
					__antithesis_instrumentation__.Notify(21372)
					log.Ops.Errorf(ctx, "dial SQL server (%d errors skipped): %v",
						dialSQLServerErrs, err)
					dialSQLServerErrs = 0
				} else {
					__antithesis_instrumentation__.Notify(21373)
				}
				__antithesis_instrumentation__.Notify(21370)

				if err = reportFailureToDirectoryCache(
					ctx, c.TenantID, serverAddr, c.DirectoryCache,
				); err != nil {
					__antithesis_instrumentation__.Notify(21374)
					reportFailureErrs++
					if reportFailureErr.ShouldLog() {
						__antithesis_instrumentation__.Notify(21375)
						log.Ops.Errorf(ctx,
							"report failure (%d errors skipped): %v",
							reportFailureErrs,
							err,
						)
						reportFailureErrs = 0
					} else {
						__antithesis_instrumentation__.Notify(21376)
					}
				} else {
					__antithesis_instrumentation__.Notify(21377)
				}
				__antithesis_instrumentation__.Notify(21371)
				continue
			} else {
				__antithesis_instrumentation__.Notify(21378)
			}
			__antithesis_instrumentation__.Notify(21368)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21379)
		}
		__antithesis_instrumentation__.Notify(21358)
		return crdbConn, nil
	}
	__antithesis_instrumentation__.Notify(21352)

	if errors.IsAny(err, context.Canceled, context.DeadlineExceeded) {
		__antithesis_instrumentation__.Notify(21380)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21381)
	}
	__antithesis_instrumentation__.Notify(21353)

	return nil, errors.Mark(err, ctx.Err())
}

func (c *connector) lookupAddr(ctx context.Context) (string, error) {
	__antithesis_instrumentation__.Notify(21382)
	if c.testingKnobs.lookupAddr != nil {
		__antithesis_instrumentation__.Notify(21384)
		return c.testingKnobs.lookupAddr(ctx)
	} else {
		__antithesis_instrumentation__.Notify(21385)
	}
	__antithesis_instrumentation__.Notify(21383)

	pods, err := c.DirectoryCache.LookupTenantPods(ctx, c.TenantID, c.ClusterName)
	switch {
	case err == nil:
		__antithesis_instrumentation__.Notify(21386)
		runningPods := make([]*tenant.Pod, 0, len(pods))
		for _, pod := range pods {
			__antithesis_instrumentation__.Notify(21393)
			if pod.State == tenant.RUNNING {
				__antithesis_instrumentation__.Notify(21394)
				runningPods = append(runningPods, pod)
			} else {
				__antithesis_instrumentation__.Notify(21395)
			}
		}
		__antithesis_instrumentation__.Notify(21387)
		pod, err := c.Balancer.SelectTenantPod(runningPods)
		if err != nil {
			__antithesis_instrumentation__.Notify(21396)

			return "", markAsRetriableConnectorError(err)
		} else {
			__antithesis_instrumentation__.Notify(21397)
		}
		__antithesis_instrumentation__.Notify(21388)
		return pod.Addr, nil

	case status.Code(err) == codes.FailedPrecondition:
		__antithesis_instrumentation__.Notify(21389)
		if st, ok := status.FromError(err); ok {
			__antithesis_instrumentation__.Notify(21398)
			return "", newErrorf(codeUnavailable, "%v", st.Message())
		} else {
			__antithesis_instrumentation__.Notify(21399)
		}
		__antithesis_instrumentation__.Notify(21390)
		return "", newErrorf(codeUnavailable, "unavailable")

	case status.Code(err) == codes.NotFound:
		__antithesis_instrumentation__.Notify(21391)
		return "", newErrorf(codeParamsRoutingFailed,
			"cluster %s-%d not found", c.ClusterName, c.TenantID.ToUint64())

	default:
		__antithesis_instrumentation__.Notify(21392)
		return "", markAsRetriableConnectorError(err)
	}
}

func (c *connector) dialSQLServer(serverAddr string) (net.Conn, error) {
	__antithesis_instrumentation__.Notify(21400)
	if c.testingKnobs.dialSQLServer != nil {
		__antithesis_instrumentation__.Notify(21404)
		return c.testingKnobs.dialSQLServer(serverAddr)
	} else {
		__antithesis_instrumentation__.Notify(21405)
	}
	__antithesis_instrumentation__.Notify(21401)

	tlsConf := c.TLSConfig.Clone()
	if tlsConf != nil {
		__antithesis_instrumentation__.Notify(21406)

		outgoingHost, _, err := addr.SplitHostPort(serverAddr, "")
		if err != nil {
			__antithesis_instrumentation__.Notify(21408)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21409)
		}
		__antithesis_instrumentation__.Notify(21407)

		tlsConf.ServerName = outgoingHost
	} else {
		__antithesis_instrumentation__.Notify(21410)
	}
	__antithesis_instrumentation__.Notify(21402)

	conn, err := BackendDial(c.StartupMsg, serverAddr, tlsConf)
	if err != nil {
		__antithesis_instrumentation__.Notify(21411)
		var codeErr *codeError
		if errors.As(err, &codeErr) && func() bool {
			__antithesis_instrumentation__.Notify(21413)
			return codeErr.code == codeBackendDown == true
		}() == true {
			__antithesis_instrumentation__.Notify(21414)
			return nil, markAsRetriableConnectorError(err)
		} else {
			__antithesis_instrumentation__.Notify(21415)
		}
		__antithesis_instrumentation__.Notify(21412)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21416)
	}
	__antithesis_instrumentation__.Notify(21403)
	return conn, nil
}

var errRetryConnectorSentinel = errors.New("retry connector error")

func markAsRetriableConnectorError(err error) error {
	__antithesis_instrumentation__.Notify(21417)
	return errors.Mark(err, errRetryConnectorSentinel)
}

func isRetriableConnectorError(err error) bool {
	__antithesis_instrumentation__.Notify(21418)
	return errors.Is(err, errRetryConnectorSentinel)
}

var reportFailureToDirectoryCache = func(
	ctx context.Context,
	tenantID roachpb.TenantID,
	addr string,
	directoryCache tenant.DirectoryCache,
) error {
	__antithesis_instrumentation__.Notify(21419)
	return directoryCache.ReportFailure(ctx, tenantID, addr)
}
