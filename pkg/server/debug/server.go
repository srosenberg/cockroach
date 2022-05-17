package debug

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/pprof"
	"path"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/sidetransport"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/debug/goroutineui"
	"github.com/cockroachdb/cockroach/pkg/server/debug/pprofui"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	pebbletool "github.com/cockroachdb/pebble/tool"
	metrics "github.com/rcrowley/go-metrics"
	"github.com/rcrowley/go-metrics/exp"
	"github.com/spf13/cobra"
	"golang.org/x/net/trace"
)

func init() {

	trace.AuthRequest = func(r *http.Request) (allowed, sensitive bool) {
		return true, true
	}
}

const Endpoint = "/debug/"

var _ = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(190457)

	v := settings.RegisterStringSetting(
		settings.TenantWritable, "server.remote_debugging.mode", "unused", "local")
	v.SetRetired()
	return v
}()

type Server struct {
	ambientCtx log.AmbientContext
	st         *cluster.Settings
	mux        *http.ServeMux
	spy        logSpy
}

func NewServer(
	ambientContext log.AmbientContext,
	st *cluster.Settings,
	hbaConfDebugFn http.HandlerFunc,
	profiler pprofui.Profiler,
) *Server {
	__antithesis_instrumentation__.Notify(190458)
	mux := http.NewServeMux()

	mux.HandleFunc(Endpoint, handleLanding)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(190462)
		CPUProfileHandler(st, w, r)
	})
	__antithesis_instrumentation__.Notify(190459)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.HandleFunc("/debug/requests", trace.Traces)
	mux.HandleFunc("/debug/events", trace.Events)

	mux.Handle("/debug/metrics", exp.ExpHandler(metrics.DefaultRegistry))

	mux.Handle("/debug/vars", expvar.Handler())

	if hbaConfDebugFn != nil {
		__antithesis_instrumentation__.Notify(190463)

		mux.HandleFunc("/debug/hba_conf", hbaConfDebugFn)
	} else {
		__antithesis_instrumentation__.Notify(190464)
	}
	__antithesis_instrumentation__.Notify(190460)

	mux.HandleFunc("/debug/stopper", stop.HandleDebug)

	vsrv := &vmoduleServer{}
	mux.HandleFunc("/debug/vmodule", vsrv.vmoduleHandleDebug)

	spy := logSpy{
		vsrv:         vsrv,
		setIntercept: log.InterceptWith,
	}
	mux.HandleFunc("/debug/logspy", spy.handleDebugLogSpy)

	ps := pprofui.NewServer(pprofui.NewMemStorage(pprofui.ProfileConcurrency, pprofui.ProfileExpiry), profiler)
	mux.Handle("/debug/pprof/ui/", http.StripPrefix("/debug/pprof/ui", ps))

	mux.HandleFunc("/debug/pprof/goroutineui/", func(w http.ResponseWriter, req *http.Request) {
		__antithesis_instrumentation__.Notify(190465)
		dump := goroutineui.NewDump()

		_ = req.ParseForm()
		switch req.Form.Get("sort") {
		case "count":
			__antithesis_instrumentation__.Notify(190467)
			dump.SortCountDesc()
		case "wait":
			__antithesis_instrumentation__.Notify(190468)
			dump.SortWaitDesc()
		default:
			__antithesis_instrumentation__.Notify(190469)
		}
		__antithesis_instrumentation__.Notify(190466)
		_ = dump.HTML(w)
	})
	__antithesis_instrumentation__.Notify(190461)

	return &Server{
		ambientCtx: ambientContext,
		st:         st,
		mux:        mux,
		spy:        spy,
	}
}

func analyzeLSM(dir string, writer io.Writer) error {
	__antithesis_instrumentation__.Notify(190470)
	manifestName, err := ioutil.ReadFile(path.Join(dir, "CURRENT"))
	if err != nil {
		__antithesis_instrumentation__.Notify(190474)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190475)
	}
	__antithesis_instrumentation__.Notify(190471)

	manifestPath := path.Join(dir, string(bytes.TrimSpace(manifestName)))

	t := pebbletool.New(pebbletool.Comparers(storage.EngineComparer))

	var lsm *cobra.Command
	for _, c := range t.Commands {
		__antithesis_instrumentation__.Notify(190476)
		if c.Name() == "lsm" {
			__antithesis_instrumentation__.Notify(190477)
			lsm = c
		} else {
			__antithesis_instrumentation__.Notify(190478)
		}
	}
	__antithesis_instrumentation__.Notify(190472)
	if lsm == nil {
		__antithesis_instrumentation__.Notify(190479)
		return errors.New("no such command")
	} else {
		__antithesis_instrumentation__.Notify(190480)
	}
	__antithesis_instrumentation__.Notify(190473)

	lsm.SetOutput(writer)
	lsm.Run(lsm, []string{manifestPath})
	return nil
}

func (ds *Server) RegisterEngines(specs []base.StoreSpec, engines []storage.Engine) error {
	__antithesis_instrumentation__.Notify(190481)
	if len(specs) != len(engines) {
		__antithesis_instrumentation__.Notify(190486)

		return errors.New("number of store specs must match number of engines")
	} else {
		__antithesis_instrumentation__.Notify(190487)
	}
	__antithesis_instrumentation__.Notify(190482)

	storeIDs := make([]roachpb.StoreIdent, len(engines))
	for i := range engines {
		__antithesis_instrumentation__.Notify(190488)
		id, err := kvserver.ReadStoreIdent(context.Background(), engines[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(190490)
			return err
		} else {
			__antithesis_instrumentation__.Notify(190491)
		}
		__antithesis_instrumentation__.Notify(190489)
		storeIDs[i] = id
	}
	__antithesis_instrumentation__.Notify(190483)

	ds.mux.HandleFunc("/debug/lsm", func(w http.ResponseWriter, req *http.Request) {
		__antithesis_instrumentation__.Notify(190492)
		for i := range engines {
			__antithesis_instrumentation__.Notify(190493)
			fmt.Fprintf(w, "Store %d:\n", storeIDs[i].StoreID)
			_, _ = io.WriteString(w, engines[i].GetMetrics().String())
			fmt.Fprintln(w)
		}
	})
	__antithesis_instrumentation__.Notify(190484)

	for i := 0; i < len(specs); i++ {
		__antithesis_instrumentation__.Notify(190494)
		if specs[i].InMemory {
			__antithesis_instrumentation__.Notify(190496)

			continue
		} else {
			__antithesis_instrumentation__.Notify(190497)
		}
		__antithesis_instrumentation__.Notify(190495)

		dir := specs[i].Path
		ds.mux.HandleFunc(fmt.Sprintf("/debug/lsm-viz/%d", storeIDs[i].StoreID),
			func(w http.ResponseWriter, req *http.Request) {
				__antithesis_instrumentation__.Notify(190498)
				if err := analyzeLSM(dir, w); err != nil {
					__antithesis_instrumentation__.Notify(190499)
					fmt.Fprintf(w, "error analyzing LSM at %s: %v", dir, err)
				} else {
					__antithesis_instrumentation__.Notify(190500)
				}
			})
	}
	__antithesis_instrumentation__.Notify(190485)
	return nil
}

type sidetransportReceiver interface {
	HTML() string
}

func (ds *Server) RegisterClosedTimestampSideTransport(
	sender *sidetransport.Sender, receiver sidetransportReceiver,
) {
	__antithesis_instrumentation__.Notify(190501)
	ds.mux.HandleFunc("/debug/closedts-receiver",
		func(w http.ResponseWriter, req *http.Request) {
			__antithesis_instrumentation__.Notify(190503)
			w.Header().Add("Content-type", "text/html")
			fmt.Fprint(w, receiver.HTML())
		})
	__antithesis_instrumentation__.Notify(190502)
	ds.mux.HandleFunc("/debug/closedts-sender",
		func(w http.ResponseWriter, req *http.Request) {
			__antithesis_instrumentation__.Notify(190504)
			w.Header().Add("Content-type", "text/html")
			fmt.Fprint(w, sender.HTML())
		})
}

func (ds *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(190505)
	handler, _ := ds.mux.Handler(r)
	handler.ServeHTTP(w, r)
}

func handleLanding(w http.ResponseWriter, r *http.Request) {
	__antithesis_instrumentation__.Notify(190506)
	if r.URL.Path != Endpoint {
		__antithesis_instrumentation__.Notify(190508)
		http.Redirect(w, r, Endpoint, http.StatusMovedPermanently)
		return
	} else {
		__antithesis_instrumentation__.Notify(190509)
	}
	__antithesis_instrumentation__.Notify(190507)

	w.Header().Add("Content-type", "text/html")

	fmt.Fprint(w, `
<html>
<head>
<meta charset="UTF-8">
<meta http-equiv="refresh" content="1; url=/#/debug">
<script type="text/javascript">
	window.location.href = "/#/debug"
</script>
<title>Page Redirection</title>
</head>
<body>
This page has moved.
If you are not redirected automatically, follow this <a href='/#/debug'>link</a>.
</body>
</html>
`)
}
