package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type Server struct {
	Stopper         *stop.Stopper
	handler         *proxyHandler
	mux             *http.ServeMux
	metrics         *metrics
	metricsRegistry *metric.Registry

	prometheusExporter metric.PrometheusExporter
}

func NewServer(ctx context.Context, stopper *stop.Stopper, options ProxyOptions) (*Server, error) {
	__antithesis_instrumentation__.Notify(22011)
	proxyMetrics := makeProxyMetrics()
	handler, err := newProxyHandler(ctx, stopper, &proxyMetrics, options)
	if err != nil {
		__antithesis_instrumentation__.Notify(22013)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(22014)
	}
	__antithesis_instrumentation__.Notify(22012)

	mux := http.NewServeMux()

	registry := metric.NewRegistry()

	registry.AddMetricStruct(&proxyMetrics)

	s := &Server{
		Stopper:            stopper,
		handler:            handler,
		mux:                mux,
		metrics:            &proxyMetrics,
		metricsRegistry:    registry,
		prometheusExporter: metric.MakePrometheusExporter(),
	}

	mux.HandleFunc("/_status/vars/", s.handleVars)
	mux.HandleFunc("/_status/healthz/", s.handleHealth)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return s, nil
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(22015)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte("OK"))
}

func (s *Server) handleVars(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(22016)
	w.Header().Set(httputil.ContentTypeHeader, httputil.PlaintextContentType)
	scrape := func(pm *metric.PrometheusExporter) {
		__antithesis_instrumentation__.Notify(22018)
		pm.ScrapeRegistry(s.metricsRegistry, true)
	}
	__antithesis_instrumentation__.Notify(22017)
	if err := s.prometheusExporter.ScrapeAndPrintAsText(w, scrape); err != nil {
		__antithesis_instrumentation__.Notify(22019)
		log.Errorf(r.Context(), "%v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		__antithesis_instrumentation__.Notify(22020)
	}
}

func (s *Server) ServeHTTP(ctx context.Context, ln net.Listener) error {
	__antithesis_instrumentation__.Notify(22021)
	srv := http.Server{
		Handler: s.mux,
	}

	go func() {
		__antithesis_instrumentation__.Notify(22024)
		<-ctx.Done()

		_ = contextutil.RunWithTimeout(
			context.Background(),
			"http server shutdown",
			15*time.Second,
			func(shutdownCtx context.Context) error {
				__antithesis_instrumentation__.Notify(22025)

				_ = srv.Shutdown(shutdownCtx)

				return nil
			},
		)
	}()
	__antithesis_instrumentation__.Notify(22022)

	if err := srv.Serve(ln); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(22026)
		return !errors.Is(err, http.ErrServerClosed) == true
	}() == true {
		__antithesis_instrumentation__.Notify(22027)
		return err
	} else {
		__antithesis_instrumentation__.Notify(22028)
	}
	__antithesis_instrumentation__.Notify(22023)

	return nil
}

func (s *Server) Serve(ctx context.Context, ln net.Listener) error {
	__antithesis_instrumentation__.Notify(22029)
	err := s.Stopper.RunAsyncTask(ctx, "listen-quiesce", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(22032)
		<-s.Stopper.ShouldQuiesce()
		if err := ln.Close(); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(22033)
			return !grpcutil.IsClosedConnection(err) == true
		}() == true {
			__antithesis_instrumentation__.Notify(22034)
			log.Fatalf(ctx, "closing proxy listener: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(22035)
		}
	})
	__antithesis_instrumentation__.Notify(22030)
	if err != nil {
		__antithesis_instrumentation__.Notify(22036)
		return err
	} else {
		__antithesis_instrumentation__.Notify(22037)
	}
	__antithesis_instrumentation__.Notify(22031)

	for {
		__antithesis_instrumentation__.Notify(22038)
		origConn, err := ln.Accept()
		if err != nil {
			__antithesis_instrumentation__.Notify(22041)
			return err
		} else {
			__antithesis_instrumentation__.Notify(22042)
		}
		__antithesis_instrumentation__.Notify(22039)
		conn := &proxyConn{
			Conn: origConn,
		}

		err = s.Stopper.RunAsyncTask(ctx, "proxy-con-serve", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(22043)
			defer func() { __antithesis_instrumentation__.Notify(22045); _ = conn.Close() }()
			__antithesis_instrumentation__.Notify(22044)
			s.metrics.CurConnCount.Inc(1)
			defer s.metrics.CurConnCount.Dec(1)
			remoteAddr := conn.RemoteAddr()
			ctxWithTag := logtags.AddTag(ctx, "client", log.SafeOperational(remoteAddr))
			if err := s.handler.handle(ctxWithTag, conn); err != nil {
				__antithesis_instrumentation__.Notify(22046)
				log.Infof(ctxWithTag, "connection error: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(22047)
			}
		})
		__antithesis_instrumentation__.Notify(22040)
		if err != nil {
			__antithesis_instrumentation__.Notify(22048)
			return err
		} else {
			__antithesis_instrumentation__.Notify(22049)
		}
	}
}

type proxyConn struct {
	net.Conn

	mu struct {
		syncutil.Mutex
		closed   bool
		closedCh chan struct{}
	}
}

func (c *proxyConn) done() <-chan struct{} {
	__antithesis_instrumentation__.Notify(22050)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.closedCh == nil {
		__antithesis_instrumentation__.Notify(22052)
		c.mu.closedCh = make(chan struct{})
		if c.mu.closed {
			__antithesis_instrumentation__.Notify(22053)
			close(c.mu.closedCh)
		} else {
			__antithesis_instrumentation__.Notify(22054)
		}
	} else {
		__antithesis_instrumentation__.Notify(22055)
	}
	__antithesis_instrumentation__.Notify(22051)
	return c.mu.closedCh
}

func (c *proxyConn) Close() error {
	__antithesis_instrumentation__.Notify(22056)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.mu.closed {
		__antithesis_instrumentation__.Notify(22059)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(22060)
	}
	__antithesis_instrumentation__.Notify(22057)
	if c.mu.closedCh != nil {
		__antithesis_instrumentation__.Notify(22061)
		close(c.mu.closedCh)
	} else {
		__antithesis_instrumentation__.Notify(22062)
	}
	__antithesis_instrumentation__.Notify(22058)
	c.mu.closed = true
	return c.Conn.Close()
}
