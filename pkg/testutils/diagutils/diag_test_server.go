package diagutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/server/diagnostics/diagnosticspb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type Server struct {
	httpSrv *httptest.Server
	url     *url.URL

	mu struct {
		syncutil.Mutex

		numRequests int
		last        *RequestData
	}
}

type RequestData struct {
	UUID          string
	TenantID      string
	NodeID        string
	SQLInstanceID string
	Version       string
	LicenseType   string
	Internal      string
	RawReportBody string

	diagnosticspb.DiagnosticReport
}

func NewServer() *Server {
	__antithesis_instrumentation__.Notify(644053)
	srv := &Server{}

	srv.httpSrv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		__antithesis_instrumentation__.Notify(644056)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			__antithesis_instrumentation__.Notify(644059)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(644060)
		}
		__antithesis_instrumentation__.Notify(644057)

		srv.mu.Lock()
		defer srv.mu.Unlock()

		srv.mu.numRequests++

		data := &RequestData{
			UUID:          r.URL.Query().Get("uuid"),
			TenantID:      r.URL.Query().Get("tenantid"),
			NodeID:        r.URL.Query().Get("nodeid"),
			SQLInstanceID: r.URL.Query().Get("sqlid"),
			Version:       r.URL.Query().Get("version"),
			LicenseType:   r.URL.Query().Get("licensetype"),
			Internal:      r.URL.Query().Get("internal"),
			RawReportBody: string(body),
		}

		if err := protoutil.Unmarshal(body, &data.DiagnosticReport); err != nil {
			__antithesis_instrumentation__.Notify(644061)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(644062)
		}
		__antithesis_instrumentation__.Notify(644058)
		srv.mu.last = data
	}))
	__antithesis_instrumentation__.Notify(644054)

	var err error
	srv.url, err = url.Parse(srv.httpSrv.URL)
	if err != nil {
		__antithesis_instrumentation__.Notify(644063)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(644064)
	}
	__antithesis_instrumentation__.Notify(644055)

	return srv
}

func (s *Server) URL() *url.URL {
	__antithesis_instrumentation__.Notify(644065)
	return s.url
}

func (s *Server) Close() {
	__antithesis_instrumentation__.Notify(644066)
	s.httpSrv.Close()
}

func (s *Server) NumRequests() int {
	__antithesis_instrumentation__.Notify(644067)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.numRequests
}

func (s *Server) LastRequestData() *RequestData {
	__antithesis_instrumentation__.Notify(644068)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.last
}
