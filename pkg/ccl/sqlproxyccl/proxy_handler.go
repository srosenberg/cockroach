package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/balancer"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/denylist"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/idle"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenant"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/tenantdirsvr"
	"github.com/cockroachdb/cockroach/pkg/ccl/sqlproxyccl/throttler"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security/certmgr"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/jackc/pgproto3/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	clusterIdentifierLongOptionRE = regexp.MustCompile(`(?:-c\s*|--)cluster=([\S]*)`)

	clusterNameRegex = regexp.MustCompile("^[a-z0-9][a-z0-9-]{4,18}[a-z0-9]$")
)

const (
	clusterTenantSep = "-"
)

type ProxyOptions struct {
	Denylist string

	ListenAddr string

	ListenCert string

	ListenKey string

	MetricsAddress string

	SkipVerify bool

	Insecure bool

	RoutingRule string

	DirectoryAddr string

	RatelimitBaseDelay time.Duration

	ValidateAccessInterval time.Duration

	PollConfigInterval time.Duration

	DrainTimeout time.Duration

	ThrottleBaseDelay time.Duration

	testingKnobs struct {
		dirOpts []tenant.DirOption

		directoryServer tenant.DirectoryServer
	}
}

type proxyHandler struct {
	ProxyOptions

	metrics *metrics

	stopper *stop.Stopper

	incomingCert certmgr.Cert

	denyListWatcher *denylist.Watcher

	throttleService throttler.Service

	idleMonitor *idle.Monitor

	directoryCache tenant.DirectoryCache

	balancer *balancer.Balancer

	connTracker *balancer.ConnTracker

	certManager *certmgr.CertManager
}

const throttledErrorHint string = `Connection throttling is triggered by repeated authentication failure. Make
sure the username and password are correct.
`

var throttledError = errors.WithHint(
	newErrorf(codeProxyRefusedConnection, "connection attempt throttled"),
	throttledErrorHint)

func newProxyHandler(
	ctx context.Context, stopper *stop.Stopper, proxyMetrics *metrics, options ProxyOptions,
) (*proxyHandler, error) {
	__antithesis_instrumentation__.Notify(21806)
	ctx, _ = stopper.WithCancelOnQuiesce(ctx)

	handler := proxyHandler{
		stopper:      stopper,
		metrics:      proxyMetrics,
		ProxyOptions: options,
		certManager:  certmgr.NewCertManager(ctx),
	}

	err := handler.setupIncomingCert(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(21815)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21816)
	}
	__antithesis_instrumentation__.Notify(21807)

	if options.Denylist != "" {
		__antithesis_instrumentation__.Notify(21817)
		handler.denyListWatcher = denylist.WatcherFromFile(ctx, options.Denylist,
			denylist.WithPollingInterval(options.PollConfigInterval))
	} else {
		__antithesis_instrumentation__.Notify(21818)
		handler.denyListWatcher = denylist.NilWatcher()
	}
	__antithesis_instrumentation__.Notify(21808)

	handler.throttleService = throttler.NewLocalService(
		throttler.WithBaseDelay(handler.ThrottleBaseDelay),
	)

	var conn *grpc.ClientConn
	if handler.DirectoryAddr != "" {
		__antithesis_instrumentation__.Notify(21819)
		conn, err = grpc.Dial(
			handler.DirectoryAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(21820)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21821)
		}
	} else {
		__antithesis_instrumentation__.Notify(21822)

		directoryServer, grpcServer := tenantdirsvr.NewTestSimpleDirectoryServer(handler.RoutingRule)
		ln, err := tenantdirsvr.ListenAndServeInMemGRPC(ctx, stopper, grpcServer)
		if err != nil {
			__antithesis_instrumentation__.Notify(21825)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21826)
		}
		__antithesis_instrumentation__.Notify(21823)
		handler.testingKnobs.directoryServer = directoryServer

		dialerFunc := func(ctx context.Context, addr string) (net.Conn, error) {
			__antithesis_instrumentation__.Notify(21827)
			return ln.DialContext(ctx)
		}
		__antithesis_instrumentation__.Notify(21824)
		conn, err = grpc.DialContext(
			ctx,
			"",
			grpc.WithContextDialer(dialerFunc),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(21828)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(21829)
		}
	}
	__antithesis_instrumentation__.Notify(21809)
	stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(21830)
		_ = conn.Close()
	}))
	__antithesis_instrumentation__.Notify(21810)

	var dirOpts []tenant.DirOption
	if options.DrainTimeout != 0 {
		__antithesis_instrumentation__.Notify(21831)
		handler.idleMonitor = idle.NewMonitor(ctx, options.DrainTimeout)

		podWatcher := make(chan *tenant.Pod)
		go handler.startPodWatcher(ctx, podWatcher)
		dirOpts = append(dirOpts, tenant.PodWatcher(podWatcher))
	} else {
		__antithesis_instrumentation__.Notify(21832)
	}
	__antithesis_instrumentation__.Notify(21811)
	if handler.testingKnobs.dirOpts != nil {
		__antithesis_instrumentation__.Notify(21833)
		dirOpts = append(dirOpts, handler.testingKnobs.dirOpts...)
	} else {
		__antithesis_instrumentation__.Notify(21834)
	}
	__antithesis_instrumentation__.Notify(21812)

	client := tenant.NewDirectoryClient(conn)
	handler.directoryCache, err = tenant.NewDirectoryCache(ctx, stopper, client, dirOpts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(21835)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21836)
	}
	__antithesis_instrumentation__.Notify(21813)

	handler.connTracker = balancer.NewConnTracker()
	handler.balancer, err = balancer.NewBalancer(ctx, stopper, handler.directoryCache, handler.connTracker)
	if err != nil {
		__antithesis_instrumentation__.Notify(21837)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21838)
	}
	__antithesis_instrumentation__.Notify(21814)

	return &handler, nil
}

func (handler *proxyHandler) handle(ctx context.Context, incomingConn *proxyConn) error {
	__antithesis_instrumentation__.Notify(21839)
	fe := FrontendAdmit(incomingConn, handler.incomingTLSConfig())
	defer func() { __antithesis_instrumentation__.Notify(21855); _ = fe.conn.Close() }()
	__antithesis_instrumentation__.Notify(21840)
	if fe.err != nil {
		__antithesis_instrumentation__.Notify(21856)
		SendErrToClient(fe.conn, fe.err)
		return fe.err
	} else {
		__antithesis_instrumentation__.Notify(21857)
	}
	__antithesis_instrumentation__.Notify(21841)

	if fe.msg == nil {
		__antithesis_instrumentation__.Notify(21858)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21859)
	}
	__antithesis_instrumentation__.Notify(21842)

	backendStartupMsg, clusterName, tenID, err := clusterNameAndTenantFromParams(ctx, fe)
	if err != nil {
		__antithesis_instrumentation__.Notify(21860)
		clientErr := &codeError{codeParamsRoutingFailed, err}
		log.Errorf(ctx, "unable to extract cluster name and tenant id: %s", err.Error())
		updateMetricsAndSendErrToClient(clientErr, fe.conn, handler.metrics)
		return clientErr
	} else {
		__antithesis_instrumentation__.Notify(21861)
	}
	__antithesis_instrumentation__.Notify(21843)

	ctx = logtags.AddTag(ctx, "cluster", clusterName)
	ctx = logtags.AddTag(ctx, "tenant", tenID)

	ipAddr, _, err := addr.SplitHostPort(fe.conn.RemoteAddr().String(), "")
	if err != nil {
		__antithesis_instrumentation__.Notify(21862)
		clientErr := newErrorf(codeParamsRoutingFailed, "unexpected connection address")
		log.Errorf(ctx, "could not parse address: %v", err.Error())
		updateMetricsAndSendErrToClient(clientErr, fe.conn, handler.metrics)
		return clientErr
	} else {
		__antithesis_instrumentation__.Notify(21863)
	}
	__antithesis_instrumentation__.Notify(21844)

	errConnection := make(chan error, 1)

	removeListener, err := handler.denyListWatcher.ListenForDenied(
		denylist.ConnectionTags{IP: ipAddr, Cluster: tenID.String()},
		func(err error) {
			__antithesis_instrumentation__.Notify(21864)
			err = newErrorf(codeExpiredClientConnection, "connection added to deny list: %v", err)
			select {
			case errConnection <- err:
				__antithesis_instrumentation__.Notify(21865)
			default:
				__antithesis_instrumentation__.Notify(21866)
			}
		},
	)
	__antithesis_instrumentation__.Notify(21845)
	if err != nil {
		__antithesis_instrumentation__.Notify(21867)
		log.Errorf(ctx, "connection matched denylist: %v", err)
		err = newErrorf(codeProxyRefusedConnection, "connection refused")
		updateMetricsAndSendErrToClient(err, fe.conn, handler.metrics)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21868)
	}
	__antithesis_instrumentation__.Notify(21846)
	defer removeListener()

	throttleTags := throttler.ConnectionTags{IP: ipAddr, TenantID: tenID.String()}
	throttleTime, err := handler.throttleService.LoginCheck(throttleTags)
	if err != nil {
		__antithesis_instrumentation__.Notify(21869)
		log.Errorf(ctx, "throttler refused connection: %v", err.Error())
		err = throttledError
		updateMetricsAndSendErrToClient(err, fe.conn, handler.metrics)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21870)
	}
	__antithesis_instrumentation__.Notify(21847)

	connector := &connector{
		ClusterName:    clusterName,
		TenantID:       tenID,
		DirectoryCache: handler.directoryCache,
		Balancer:       handler.balancer,
		StartupMsg:     backendStartupMsg,
	}

	if !handler.Insecure {
		__antithesis_instrumentation__.Notify(21871)
		connector.TLSConfig = &tls.Config{InsecureSkipVerify: handler.SkipVerify}
	} else {
		__antithesis_instrumentation__.Notify(21872)
	}
	__antithesis_instrumentation__.Notify(21848)

	if handler.idleMonitor != nil {
		__antithesis_instrumentation__.Notify(21873)
		connector.IdleMonitorWrapperFn = func(serverConn net.Conn) net.Conn {
			__antithesis_instrumentation__.Notify(21874)
			return handler.idleMonitor.DetectIdle(serverConn, func() {
				__antithesis_instrumentation__.Notify(21875)
				err := newErrorf(codeIdleDisconnect, "idle connection closed")
				select {
				case errConnection <- err:
					__antithesis_instrumentation__.Notify(21876)
				default:
					__antithesis_instrumentation__.Notify(21877)
				}
			})
		}
	} else {
		__antithesis_instrumentation__.Notify(21878)
	}
	__antithesis_instrumentation__.Notify(21849)

	crdbConn, sentToClient, err := connector.OpenTenantConnWithAuth(ctx, fe.conn,
		func(status throttler.AttemptStatus) error {
			__antithesis_instrumentation__.Notify(21879)
			if err := handler.throttleService.ReportAttempt(
				ctx, throttleTags, throttleTime, status,
			); err != nil {
				__antithesis_instrumentation__.Notify(21881)
				log.Errorf(ctx, "throttler refused connection after authentication: %v", err.Error())
				return throttledError
			} else {
				__antithesis_instrumentation__.Notify(21882)
			}
			__antithesis_instrumentation__.Notify(21880)
			return nil
		},
	)
	__antithesis_instrumentation__.Notify(21850)
	if err != nil {
		__antithesis_instrumentation__.Notify(21883)
		log.Errorf(ctx, "could not connect to cluster: %v", err.Error())
		if sentToClient {
			__antithesis_instrumentation__.Notify(21885)
			handler.metrics.updateForError(err)
		} else {
			__antithesis_instrumentation__.Notify(21886)
			updateMetricsAndSendErrToClient(err, fe.conn, handler.metrics)
		}
		__antithesis_instrumentation__.Notify(21884)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21887)
	}
	__antithesis_instrumentation__.Notify(21851)
	defer func() { __antithesis_instrumentation__.Notify(21888); _ = crdbConn.Close() }()
	__antithesis_instrumentation__.Notify(21852)

	handler.metrics.SuccessfulConnCount.Inc(1)

	log.Infof(ctx, "new connection")
	connBegin := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(21889)
		log.Infof(ctx, "closing after %.2fs", timeutil.Since(connBegin).Seconds())
	}()
	__antithesis_instrumentation__.Notify(21853)

	f, err := forward(ctx, connector, handler.metrics, fe.conn, crdbConn)
	if err != nil {
		__antithesis_instrumentation__.Notify(21890)

		handler.metrics.updateForError(err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21891)
	}
	__antithesis_instrumentation__.Notify(21854)
	defer f.Close()

	handler.connTracker.OnConnect(tenID, f)
	defer handler.connTracker.OnDisconnect(tenID, f)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(21892)

		handler.metrics.updateForError(ctx.Err())
		return ctx.Err()
	case err := <-f.errCh:
		__antithesis_instrumentation__.Notify(21893)
		handler.metrics.updateForError(err)
		return err
	case err := <-errConnection:
		__antithesis_instrumentation__.Notify(21894)
		handler.metrics.updateForError(err)
		return err
	case <-handler.stopper.ShouldQuiesce():
		__antithesis_instrumentation__.Notify(21895)
		err := context.Canceled
		handler.metrics.updateForError(err)
		return err
	}
}

func (handler *proxyHandler) startPodWatcher(ctx context.Context, podWatcher chan *tenant.Pod) {
	__antithesis_instrumentation__.Notify(21896)
	for {
		__antithesis_instrumentation__.Notify(21897)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(21898)
			return
		case pod := <-podWatcher:
			__antithesis_instrumentation__.Notify(21899)
			if pod.State == tenant.DRAINING {
				__antithesis_instrumentation__.Notify(21900)
				handler.idleMonitor.SetIdleChecks(pod.Addr)
			} else {
				__antithesis_instrumentation__.Notify(21901)

				handler.idleMonitor.ClearIdleChecks(pod.Addr)
			}
		}
	}
}

func (handler *proxyHandler) incomingTLSConfig() *tls.Config {
	__antithesis_instrumentation__.Notify(21902)
	if handler.incomingCert == nil {
		__antithesis_instrumentation__.Notify(21905)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21906)
	}
	__antithesis_instrumentation__.Notify(21903)

	cert := handler.incomingCert.TLSCert()
	if cert == nil {
		__antithesis_instrumentation__.Notify(21907)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21908)
	}
	__antithesis_instrumentation__.Notify(21904)

	return &tls.Config{Certificates: []tls.Certificate{*cert}}
}

func (handler *proxyHandler) setupIncomingCert(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(21909)
	if (handler.ListenKey == "") != (handler.ListenCert == "") {
		__antithesis_instrumentation__.Notify(21914)
		return errors.New("must specify either both or neither of cert and key")
	} else {
		__antithesis_instrumentation__.Notify(21915)
	}
	__antithesis_instrumentation__.Notify(21910)

	if handler.ListenCert == "" {
		__antithesis_instrumentation__.Notify(21916)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(21917)
	}
	__antithesis_instrumentation__.Notify(21911)

	certMgr := certmgr.NewCertManager(ctx)
	var cert certmgr.Cert
	if handler.ListenCert == "*" {
		__antithesis_instrumentation__.Notify(21918)
		cert = certmgr.NewSelfSignedCert(0, 3, 0, 0)
	} else {
		__antithesis_instrumentation__.Notify(21919)
		if handler.ListenCert != "" {
			__antithesis_instrumentation__.Notify(21920)
			cert = certmgr.NewFileCert(handler.ListenCert, handler.ListenKey)
		} else {
			__antithesis_instrumentation__.Notify(21921)
		}
	}
	__antithesis_instrumentation__.Notify(21912)
	cert.Reload(ctx)
	err := cert.Err()
	if err != nil {
		__antithesis_instrumentation__.Notify(21922)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21923)
	}
	__antithesis_instrumentation__.Notify(21913)
	certMgr.ManageCert("client", cert)
	handler.certManager = certMgr
	handler.incomingCert = cert

	return nil
}

func clusterNameAndTenantFromParams(
	ctx context.Context, fe *FrontendAdmitInfo,
) (*pgproto3.StartupMessage, string, roachpb.TenantID, error) {
	__antithesis_instrumentation__.Notify(21924)
	clusterIdentifierDB, databaseName, err := parseDatabaseParam(fe.msg.Parameters["database"])
	if err != nil {
		__antithesis_instrumentation__.Notify(21936)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21937)
	}
	__antithesis_instrumentation__.Notify(21925)

	clusterIdentifierOpt, newOptionsParam, err := parseOptionsParam(fe.msg.Parameters["options"])
	if err != nil {
		__antithesis_instrumentation__.Notify(21938)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21939)
	}
	__antithesis_instrumentation__.Notify(21926)

	sniTenID, sniPresent := parseSNI(fe.sniServerName)

	if clusterIdentifierDB == "" && func() bool {
		__antithesis_instrumentation__.Notify(21940)
		return clusterIdentifierOpt == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(21941)
		if sniPresent {
			__antithesis_instrumentation__.Notify(21943)
			return fe.msg, "", sniTenID, nil
		} else {
			__antithesis_instrumentation__.Notify(21944)
		}
		__antithesis_instrumentation__.Notify(21942)
		err := errors.New("missing cluster identifier")
		err = errors.WithHint(err, clusterIdentifierHint)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21945)
	}
	__antithesis_instrumentation__.Notify(21927)

	if clusterIdentifierDB != "" && func() bool {
		__antithesis_instrumentation__.Notify(21946)
		return clusterIdentifierOpt != "" == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(21947)
		return clusterIdentifierDB != clusterIdentifierOpt == true
	}() == true {
		__antithesis_instrumentation__.Notify(21948)
		err := errors.New("multiple different cluster identifiers provided")
		err = errors.WithHintf(err,
			"Is '%s' or '%s' the identifier for the cluster that you're connecting to?",
			clusterIdentifierDB, clusterIdentifierOpt)
		err = errors.WithHint(err, clusterIdentifierHint)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21949)
	}
	__antithesis_instrumentation__.Notify(21928)

	if clusterIdentifierDB == "" {
		__antithesis_instrumentation__.Notify(21950)
		clusterIdentifierDB = clusterIdentifierOpt
	} else {
		__antithesis_instrumentation__.Notify(21951)
	}
	__antithesis_instrumentation__.Notify(21929)

	sepIdx := strings.LastIndex(clusterIdentifierDB, clusterTenantSep)

	if sepIdx == -1 || func() bool {
		__antithesis_instrumentation__.Notify(21952)
		return sepIdx == len(clusterIdentifierDB)-1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(21953)
		err := errors.Errorf("invalid cluster identifier '%s'", clusterIdentifierDB)
		err = errors.WithHint(err, missingTenantIDHint)
		err = errors.WithHint(err, clusterNameFormHint)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21954)
	}
	__antithesis_instrumentation__.Notify(21930)

	clusterName, tenantIDStr := clusterIdentifierDB[:sepIdx], clusterIdentifierDB[sepIdx+1:]

	if !clusterNameRegex.MatchString(clusterName) {
		__antithesis_instrumentation__.Notify(21955)
		err := errors.Errorf("invalid cluster identifier '%s'", clusterIdentifierDB)
		err = errors.WithHintf(err, "Is '%s' a valid cluster name?", clusterName)
		err = errors.WithHint(err, clusterNameFormHint)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21956)
	}
	__antithesis_instrumentation__.Notify(21931)

	tenID, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(21957)

		log.Errorf(ctx, "cannot parse tenant ID in %s: %v", clusterIdentifierDB, err)
		err := errors.Errorf("invalid cluster identifier '%s'", clusterIdentifierDB)
		err = errors.WithHintf(err, "Is '%s' a valid tenant ID?", tenantIDStr)
		err = errors.WithHint(err, clusterNameFormHint)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21958)
	}
	__antithesis_instrumentation__.Notify(21932)

	if tenID < roachpb.MinTenantID.ToUint64() {
		__antithesis_instrumentation__.Notify(21959)

		log.Errorf(ctx, "%s contains an invalid tenant ID", clusterIdentifierDB)
		err := errors.Errorf("invalid cluster identifier '%s'", clusterIdentifierDB)
		err = errors.WithHintf(err, "Tenant ID %d is invalid.", tenID)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21960)
	}
	__antithesis_instrumentation__.Notify(21933)

	paramsOut := map[string]string{}
	for key, value := range fe.msg.Parameters {
		__antithesis_instrumentation__.Notify(21961)
		if key == "database" {
			__antithesis_instrumentation__.Notify(21962)
			paramsOut[key] = databaseName
		} else {
			__antithesis_instrumentation__.Notify(21963)
			if key == "options" {
				__antithesis_instrumentation__.Notify(21964)
				if newOptionsParam != "" {
					__antithesis_instrumentation__.Notify(21965)
					paramsOut[key] = newOptionsParam
				} else {
					__antithesis_instrumentation__.Notify(21966)
				}
			} else {
				__antithesis_instrumentation__.Notify(21967)
				paramsOut[key] = value
			}
		}
	}
	__antithesis_instrumentation__.Notify(21934)

	if sniPresent && func() bool {
		__antithesis_instrumentation__.Notify(21968)
		return tenID != sniTenID.InternalValue == true
	}() == true {
		__antithesis_instrumentation__.Notify(21969)
		err := errors.New("multiple different tenant IDs provided")
		err = errors.WithHintf(err,
			"Is '%d' (SNI) or '%d' (database/options) the identifier for the cluster that you're connecting to?",
			sniTenID.InternalValue, tenID)
		err = errors.WithHint(err, clusterIdentifierHint)
		return fe.msg, "", roachpb.MaxTenantID, err
	} else {
		__antithesis_instrumentation__.Notify(21970)
	}
	__antithesis_instrumentation__.Notify(21935)

	outMsg := &pgproto3.StartupMessage{
		ProtocolVersion: fe.msg.ProtocolVersion,
		Parameters:      paramsOut,
	}
	return outMsg, clusterName, roachpb.MakeTenantID(tenID), nil
}

func parseSNI(sniServerName string) (roachpb.TenantID, bool) {
	__antithesis_instrumentation__.Notify(21971)
	if sniServerName == "" {
		__antithesis_instrumentation__.Notify(21976)
		return roachpb.MaxTenantID, false
	} else {
		__antithesis_instrumentation__.Notify(21977)
	}
	__antithesis_instrumentation__.Notify(21972)

	parts := strings.Split(sniServerName, ".")
	if len(parts) == 0 {
		__antithesis_instrumentation__.Notify(21978)
		return roachpb.MaxTenantID, false
	} else {
		__antithesis_instrumentation__.Notify(21979)
	}
	__antithesis_instrumentation__.Notify(21973)

	hostname := parts[0]
	hostnameParts := strings.Split(hostname, "-")

	if len(hostnameParts) != 2 || func() bool {
		__antithesis_instrumentation__.Notify(21980)
		return !strings.EqualFold("serverless", hostnameParts[0]) == true
	}() == true {
		__antithesis_instrumentation__.Notify(21981)
		return roachpb.MaxTenantID, false
	} else {
		__antithesis_instrumentation__.Notify(21982)
	}
	__antithesis_instrumentation__.Notify(21974)

	tenID, err := strconv.ParseUint(hostnameParts[1], 10, 64)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(21983)
		return tenID < roachpb.MinTenantID.ToUint64() == true
	}() == true {
		__antithesis_instrumentation__.Notify(21984)
		return roachpb.MaxTenantID, false
	} else {
		__antithesis_instrumentation__.Notify(21985)
	}
	__antithesis_instrumentation__.Notify(21975)

	return roachpb.MakeTenantID(tenID), true
}

func parseDatabaseParam(databaseParam string) (clusterIdentifier, databaseName string, err error) {
	__antithesis_instrumentation__.Notify(21986)

	if databaseParam == "" {
		__antithesis_instrumentation__.Notify(21990)
		return "", "", nil
	} else {
		__antithesis_instrumentation__.Notify(21991)
	}
	__antithesis_instrumentation__.Notify(21987)

	parts := strings.Split(databaseParam, ".")

	if len(parts) <= 1 {
		__antithesis_instrumentation__.Notify(21992)
		return "", databaseParam, nil
	} else {
		__antithesis_instrumentation__.Notify(21993)
	}
	__antithesis_instrumentation__.Notify(21988)

	clusterIdentifier, databaseName = parts[0], parts[1]

	if len(parts) > 2 || func() bool {
		__antithesis_instrumentation__.Notify(21994)
		return clusterIdentifier == "" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(21995)
		return databaseName == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(21996)
		return "", "", errors.New("invalid database param")
	} else {
		__antithesis_instrumentation__.Notify(21997)
	}
	__antithesis_instrumentation__.Notify(21989)

	return clusterIdentifier, databaseName, nil
}

func parseOptionsParam(optionsParam string) (clusterIdentifier, newOptionsParam string, err error) {
	__antithesis_instrumentation__.Notify(21998)

	matches := clusterIdentifierLongOptionRE.FindAllStringSubmatch(optionsParam, 2)
	if len(matches) == 0 {
		__antithesis_instrumentation__.Notify(22003)
		return "", optionsParam, nil
	} else {
		__antithesis_instrumentation__.Notify(22004)
	}
	__antithesis_instrumentation__.Notify(21999)

	if len(matches) > 1 {
		__antithesis_instrumentation__.Notify(22005)

		return "", "", errors.New("multiple cluster flags provided")
	} else {
		__antithesis_instrumentation__.Notify(22006)
	}
	__antithesis_instrumentation__.Notify(22000)

	if len(matches[0]) != 2 {
		__antithesis_instrumentation__.Notify(22007)

		return "", "", errors.New("internal server error")
	} else {
		__antithesis_instrumentation__.Notify(22008)
	}
	__antithesis_instrumentation__.Notify(22001)

	if len(matches[0][1]) == 0 {
		__antithesis_instrumentation__.Notify(22009)
		return "", "", errors.New("invalid cluster flag")
	} else {
		__antithesis_instrumentation__.Notify(22010)
	}
	__antithesis_instrumentation__.Notify(22002)

	newOptionsParam = strings.ReplaceAll(optionsParam, matches[0][0], "")
	newOptionsParam = strings.TrimSpace(newOptionsParam)
	return matches[0][1], newOptionsParam, nil
}

const clusterIdentifierHint = `Ensure that your cluster identifier is uniquely specified using any of the
following methods:

1) Database parameter:
   Use "<cluster identifier>.<database name>" as the database parameter.
   (e.g. database="active-roach-42.defaultdb")

2) Options parameter:
   Use "--cluster=<cluster identifier>" as the options parameter.
   (e.g. options="--cluster=active-roach-42")

3) Host name:
   Use a driver that supports server name identification (SNI) with TLS 
   connection and the hostname assigned to your cluster 
   (e.g. serverless-101.5xj.gcp-us-central1.cockroachlabs.cloud)

For more details, please visit our docs site at:
	https://www.cockroachlabs.com/docs/cockroachcloud/connect-to-a-serverless-cluster
`

const clusterNameFormHint = "Cluster identifiers come in the form of <name>-<tenant ID> (e.g. lazy-roach-3)."

const missingTenantIDHint = "Did you forget to include your tenant ID in the cluster identifier?"
