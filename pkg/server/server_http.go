package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"net/http"
	"strings"

	"github.com/cockroachdb/cmux"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/debug"
	"github.com/cockroachdb/cockroach/pkg/server/status"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/ui"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/netutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"google.golang.org/grpc/metadata"
)

type httpServer struct {
	cfg   BaseConfig
	mux   http.ServeMux
	proxy *nodeProxy
}

func newHTTPServer(
	cfg BaseConfig,
	rpcContext *rpc.Context,
	parseNodeID ParseNodeIDFn,
	getNodeIDHTTPAddress GetNodeIDHTTPAddressFn,
) *httpServer {
	__antithesis_instrumentation__.Notify(195802)
	return &httpServer{
		cfg: cfg,
		proxy: &nodeProxy{
			scheme:               cfg.HTTPRequestScheme(),
			parseNodeID:          parseNodeID,
			getNodeIDHTTPAddress: getNodeIDHTTPAddress,
			rpcContext:           rpcContext,
		},
	}
}

var HSTSEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"server.hsts.enabled",
	"if true, HSTS headers will be sent along with all HTTP "+
		"requests. The headers will contain a max-age setting of one "+
		"year. Browsers honoring the header will always use HTTPS to "+
		"access the DB Console. Ensure that TLS is correctly configured "+
		"prior to enabling.",
	false,
).WithPublic()

const hstsHeaderKey = "Strict-Transport-Security"

const hstsHeaderValue = "max-age=31536000"

const healthPath = "/health"

func (s *httpServer) handleHealth(healthHandler http.Handler) {
	__antithesis_instrumentation__.Notify(195803)
	s.mux.Handle(healthPath, healthHandler)
}

func (s *httpServer) setupRoutes(
	ctx context.Context,
	authnServer *authenticationServer,
	adminAuthzCheck *adminPrivilegeChecker,
	metricSource metricMarshaler,
	runtimeStatSampler *status.RuntimeStatSampler,
	handleRequestsUnauthenticated http.Handler,
	handleDebugUnauthenticated http.Handler,
	apiServer *apiV2Server,
) error {
	__antithesis_instrumentation__.Notify(195804)

	oidc, err := ConfigureOIDC(
		ctx, s.cfg.Settings, s.cfg.Locality,
		s.mux.Handle, authnServer.UserLoginFromSSO, s.cfg.AmbientCtx, s.cfg.ClusterIDContainer.Get(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(195811)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195812)
	}
	__antithesis_instrumentation__.Notify(195805)

	assetHandler := ui.Handler(ui.Config{
		ExperimentalUseLogin: s.cfg.EnableWebSessionAuthentication,
		LoginEnabled:         s.cfg.RequireWebSession(),
		NodeID:               s.cfg.IDContainer,
		OIDC:                 oidc,
		GetUser: func(ctx context.Context) *string {
			__antithesis_instrumentation__.Notify(195813)
			if u, ok := ctx.Value(webSessionUserKey{}).(string); ok {
				__antithesis_instrumentation__.Notify(195815)
				return &u
			} else {
				__antithesis_instrumentation__.Notify(195816)
			}
			__antithesis_instrumentation__.Notify(195814)
			return nil
		},
	})
	__antithesis_instrumentation__.Notify(195806)

	authenticatedUIHandler := newAuthenticationMuxAllowAnonymous(
		authnServer, assetHandler)
	s.mux.Handle("/", authenticatedUIHandler)

	var authenticatedHandler = handleRequestsUnauthenticated
	if s.cfg.RequireWebSession() {
		__antithesis_instrumentation__.Notify(195817)
		authenticatedHandler = newAuthenticationMux(authnServer, authenticatedHandler)
	} else {
		__antithesis_instrumentation__.Notify(195818)
	}
	__antithesis_instrumentation__.Notify(195807)

	s.mux.Handle(loginPath, handleRequestsUnauthenticated)
	s.mux.Handle(logoutPath, authenticatedHandler)

	if s.cfg.EnableDemoLoginEndpoint {
		__antithesis_instrumentation__.Notify(195819)
		s.mux.Handle(DemoLoginPath, http.HandlerFunc(authnServer.demoLogin))
	} else {
		__antithesis_instrumentation__.Notify(195820)
	}
	__antithesis_instrumentation__.Notify(195808)

	s.mux.Handle(statusPrefix, authenticatedHandler)
	s.mux.Handle(adminPrefix, authenticatedHandler)

	s.mux.Handle(ts.URLPrefix, authenticatedHandler)

	s.mux.Handle(adminHealth, handleRequestsUnauthenticated)

	s.mux.Handle(statusVars, http.HandlerFunc(varsHandler{metricSource, s.cfg.Settings}.handleVars))

	s.mux.Handle(loadStatusVars, http.HandlerFunc(makeStatusLoadHandler(ctx, runtimeStatSampler, metricSource)))

	if apiServer != nil {
		__antithesis_instrumentation__.Notify(195821)

		s.mux.Handle(apiV2Path, apiServer)
	} else {
		__antithesis_instrumentation__.Notify(195822)
	}
	__antithesis_instrumentation__.Notify(195809)

	handleDebugAuthenticated := handleDebugUnauthenticated
	if s.cfg.RequireWebSession() {
		__antithesis_instrumentation__.Notify(195823)

		handleDebugAuthenticated = makeAdminAuthzCheckHandler(adminAuthzCheck, handleDebugAuthenticated)
		handleDebugAuthenticated = newAuthenticationMux(authnServer, handleDebugAuthenticated)
	} else {
		__antithesis_instrumentation__.Notify(195824)
	}
	__antithesis_instrumentation__.Notify(195810)
	s.mux.Handle(debug.Endpoint, handleDebugAuthenticated)

	log.Event(ctx, "added http endpoints")

	return nil
}

func makeAdminAuthzCheckHandler(
	adminAuthzCheck *adminPrivilegeChecker, handler http.Handler,
) http.Handler {
	__antithesis_instrumentation__.Notify(195825)
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		__antithesis_instrumentation__.Notify(195826)

		md := forwardAuthenticationMetadata(req.Context(), req)
		authCtx := metadata.NewIncomingContext(req.Context(), md)

		_, err := adminAuthzCheck.requireAdminUser(authCtx)
		if errors.Is(err, errRequiresAdmin) {
			__antithesis_instrumentation__.Notify(195828)
			http.Error(w, "admin privilege required", http.StatusUnauthorized)
			return
		} else {
			__antithesis_instrumentation__.Notify(195829)
			if err != nil {
				__antithesis_instrumentation__.Notify(195830)
				log.Ops.Infof(authCtx, "web session error: %s", err)
				http.Error(w, "error checking authentication", http.StatusInternalServerError)
				return
			} else {
				__antithesis_instrumentation__.Notify(195831)
			}
		}
		__antithesis_instrumentation__.Notify(195827)

		handler.ServeHTTP(w, req)
	})
}

func (s *httpServer) start(
	ctx, workersCtx context.Context,
	connManager netutil.Server,
	uiTLSConfig *tls.Config,
	stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(195832)
	httpLn, err := ListenAndUpdateAddrs(ctx, &s.cfg.HTTPAddr, &s.cfg.HTTPAdvertiseAddr, "http")
	if err != nil {
		__antithesis_instrumentation__.Notify(195837)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195838)
	}
	__antithesis_instrumentation__.Notify(195833)
	log.Eventf(ctx, "listening on http port %s", s.cfg.HTTPAddr)

	waitQuiesce := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(195839)

		<-stopper.ShouldQuiesce()
		if err := httpLn.Close(); err != nil {
			__antithesis_instrumentation__.Notify(195840)
			log.Ops.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(195841)
		}
	}
	__antithesis_instrumentation__.Notify(195834)
	if err := stopper.RunAsyncTask(workersCtx, "wait-quiesce", waitQuiesce); err != nil {
		__antithesis_instrumentation__.Notify(195842)
		waitQuiesce(workersCtx)
		return err
	} else {
		__antithesis_instrumentation__.Notify(195843)
	}
	__antithesis_instrumentation__.Notify(195835)

	if uiTLSConfig != nil {
		__antithesis_instrumentation__.Notify(195844)
		httpMux := cmux.New(httpLn)
		clearL := httpMux.Match(cmux.HTTP1())
		tlsL := httpMux.Match(cmux.Any())

		if err := stopper.RunAsyncTask(workersCtx, "serve-ui", func(context.Context) {
			__antithesis_instrumentation__.Notify(195847)
			netutil.FatalIfUnexpected(httpMux.Serve())
		}); err != nil {
			__antithesis_instrumentation__.Notify(195848)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195849)
		}
		__antithesis_instrumentation__.Notify(195845)

		if err := stopper.RunAsyncTask(workersCtx, "serve-health", func(context.Context) {
			__antithesis_instrumentation__.Notify(195850)
			mux := http.NewServeMux()
			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				__antithesis_instrumentation__.Notify(195852)
				if HSTSEnabled.Get(&s.cfg.Settings.SV) {
					__antithesis_instrumentation__.Notify(195854)
					w.Header().Set(hstsHeaderKey, hstsHeaderValue)
				} else {
					__antithesis_instrumentation__.Notify(195855)
				}
				__antithesis_instrumentation__.Notify(195853)
				http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusTemporaryRedirect)
			})
			__antithesis_instrumentation__.Notify(195851)
			mux.Handle(healthPath, http.HandlerFunc(s.baseHandler))

			plainRedirectServer := netutil.MakeServer(stopper, uiTLSConfig, mux)

			netutil.FatalIfUnexpected(plainRedirectServer.Serve(clearL))
		}); err != nil {
			__antithesis_instrumentation__.Notify(195856)
			return err
		} else {
			__antithesis_instrumentation__.Notify(195857)
		}
		__antithesis_instrumentation__.Notify(195846)

		httpLn = tls.NewListener(tlsL, uiTLSConfig)
	} else {
		__antithesis_instrumentation__.Notify(195858)
	}
	__antithesis_instrumentation__.Notify(195836)

	return stopper.RunAsyncTask(workersCtx, "server-http", func(context.Context) {
		__antithesis_instrumentation__.Notify(195859)
		netutil.FatalIfUnexpected(connManager.Serve(httpLn))
	})
}

func (s *httpServer) gzipHandler(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(195860)
	ae := r.Header.Get(httputil.AcceptEncodingHeader)
	switch {
	case strings.Contains(ae, httputil.GzipEncoding):
		__antithesis_instrumentation__.Notify(195862)
		w.Header().Set(httputil.ContentEncodingHeader, httputil.GzipEncoding)
		gzw := newGzipResponseWriter(w)
		defer func() {
			__antithesis_instrumentation__.Notify(195865)

			if err := gzw.Close(); err != nil && func() bool {
				__antithesis_instrumentation__.Notify(195866)
				return !errors.Is(err, http.ErrBodyNotAllowed) == true
			}() == true {
				__antithesis_instrumentation__.Notify(195867)
				ctx := s.cfg.AmbientCtx.AnnotateCtx(r.Context())
				log.Ops.Warningf(ctx, "error closing gzip response writer: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(195868)
			}
		}()
		__antithesis_instrumentation__.Notify(195863)
		w = gzw
	default:
		__antithesis_instrumentation__.Notify(195864)
	}
	__antithesis_instrumentation__.Notify(195861)
	s.mux.ServeHTTP(w, r)
}

func (s *httpServer) baseHandler(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(195869)

	w.Header().Set("Cache-control", "no-cache")

	if HSTSEnabled.Get(&s.cfg.Settings.SV) {
		__antithesis_instrumentation__.Notify(195872)
		w.Header().Set("Strict-Transport-Security", hstsHeaderValue)
	} else {
		__antithesis_instrumentation__.Notify(195873)
	}
	__antithesis_instrumentation__.Notify(195870)

	defer func() {
		__antithesis_instrumentation__.Notify(195874)
		if p := recover(); p != nil && func() bool {
			__antithesis_instrumentation__.Notify(195875)
			return p != http.ErrAbortHandler == true
		}() == true {
			__antithesis_instrumentation__.Notify(195876)

			logcrash.ReportPanic(context.Background(), &s.cfg.Settings.SV, p, 1)
			http.Error(w, errAPIInternalErrorString, http.StatusInternalServerError)
		} else {
			__antithesis_instrumentation__.Notify(195877)
		}
	}()
	__antithesis_instrumentation__.Notify(195871)

	s.proxy.nodeProxyHandler(w, r, s.gzipHandler)
}
