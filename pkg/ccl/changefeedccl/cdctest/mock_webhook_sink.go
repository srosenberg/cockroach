package cdctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type MockWebhookSink struct {
	basicAuth          bool
	username, password string
	server             *httptest.Server
	mu                 struct {
		syncutil.Mutex
		numCalls         int
		statusCodes      []int
		statusCodesIndex int
		rows             []string
	}
}

func StartMockWebhookSinkInsecure() (*MockWebhookSink, error) {
	__antithesis_instrumentation__.Notify(14681)
	s := makeMockWebhookSink()
	s.server.Start()
	return s, nil
}

func StartMockWebhookSink(certificate *tls.Certificate) (*MockWebhookSink, error) {
	__antithesis_instrumentation__.Notify(14682)
	s := makeMockWebhookSink()
	if certificate == nil {
		__antithesis_instrumentation__.Notify(14684)
		return nil, errors.Errorf("Must pass a CA cert when creating a mock webhook sink.")
	} else {
		__antithesis_instrumentation__.Notify(14685)
	}
	__antithesis_instrumentation__.Notify(14683)
	s.server.TLS = &tls.Config{
		Certificates: []tls.Certificate{*certificate},
	}
	s.server.StartTLS()
	return s, nil
}

func StartMockWebhookSinkSecure(certificate *tls.Certificate) (*MockWebhookSink, error) {
	__antithesis_instrumentation__.Notify(14686)
	s := makeMockWebhookSink()
	if certificate == nil {
		__antithesis_instrumentation__.Notify(14688)
		return nil, errors.Errorf("Must pass a CA cert when creating a mock webhook sink.")
	} else {
		__antithesis_instrumentation__.Notify(14689)
	}
	__antithesis_instrumentation__.Notify(14687)

	s.server.TLS = &tls.Config{
		Certificates: []tls.Certificate{*certificate},
		ClientAuth:   tls.RequireAnyClientCert,
	}

	s.server.StartTLS()
	return s, nil
}

func StartMockWebhookSinkWithBasicAuth(
	certificate *tls.Certificate, username, password string,
) (*MockWebhookSink, error) {
	__antithesis_instrumentation__.Notify(14690)
	s := makeMockWebhookSink()
	s.basicAuth = true
	s.username = username
	s.password = password
	if certificate != nil {
		__antithesis_instrumentation__.Notify(14692)
		s.server.TLS = &tls.Config{
			Certificates: []tls.Certificate{*certificate},
		}
	} else {
		__antithesis_instrumentation__.Notify(14693)
	}
	__antithesis_instrumentation__.Notify(14691)
	s.server.StartTLS()
	return s, nil
}

func makeMockWebhookSink() *MockWebhookSink {
	__antithesis_instrumentation__.Notify(14694)
	s := &MockWebhookSink{}
	s.mu.statusCodes = []int{http.StatusOK}
	s.server = httptest.NewUnstartedServer(http.HandlerFunc(s.requestHandler))
	return s
}

func (s *MockWebhookSink) URL() string {
	__antithesis_instrumentation__.Notify(14695)
	return s.server.URL
}

func (s *MockWebhookSink) GetNumCalls() int {
	__antithesis_instrumentation__.Notify(14696)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.numCalls
}

func (s *MockWebhookSink) SetStatusCodes(statusCodes []int) {
	__antithesis_instrumentation__.Notify(14697)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.statusCodes = statusCodes
}

func (s *MockWebhookSink) Close() {
	__antithesis_instrumentation__.Notify(14698)
	s.server.Close()
	s.server.CloseClientConnections()
}

func (s *MockWebhookSink) Latest() string {
	__antithesis_instrumentation__.Notify(14699)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.mu.rows) == 0 {
		__antithesis_instrumentation__.Notify(14701)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(14702)
	}
	__antithesis_instrumentation__.Notify(14700)
	latest := s.mu.rows[len(s.mu.rows)-1]
	return latest
}

func (s *MockWebhookSink) Pop() string {
	__antithesis_instrumentation__.Notify(14703)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.mu.rows) > 0 {
		__antithesis_instrumentation__.Notify(14705)
		oldest := s.mu.rows[0]
		s.mu.rows = s.mu.rows[1:]
		return oldest
	} else {
		__antithesis_instrumentation__.Notify(14706)
	}
	__antithesis_instrumentation__.Notify(14704)
	return ""
}

func (s *MockWebhookSink) requestHandler(hw http.ResponseWriter, hr *http.Request) {
	__antithesis_instrumentation__.Notify(14707)
	method := hr.Method

	var err error
	switch {
	case method == http.MethodPost:
		__antithesis_instrumentation__.Notify(14709)
		if s.basicAuth {
			__antithesis_instrumentation__.Notify(14712)
			username, password, ok := hr.BasicAuth()
			if !ok || func() bool {
				__antithesis_instrumentation__.Notify(14713)
				return s.username != username == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(14714)
				return s.password != password == true
			}() == true {
				__antithesis_instrumentation__.Notify(14715)
				hw.WriteHeader(http.StatusUnauthorized)
				return
			} else {
				__antithesis_instrumentation__.Notify(14716)
			}
		} else {
			__antithesis_instrumentation__.Notify(14717)
		}
		__antithesis_instrumentation__.Notify(14710)
		err = s.publish(hw, hr)
	default:
		__antithesis_instrumentation__.Notify(14711)
		hw.WriteHeader(http.StatusNotFound)
		return
	}
	__antithesis_instrumentation__.Notify(14708)
	if err != nil {
		__antithesis_instrumentation__.Notify(14718)
		http.Error(hw, err.Error(), http.StatusInternalServerError)
	} else {
		__antithesis_instrumentation__.Notify(14719)
	}
}

func (s *MockWebhookSink) publish(hw http.ResponseWriter, hr *http.Request) error {
	__antithesis_instrumentation__.Notify(14720)
	defer hr.Body.Close()
	row, err := ioutil.ReadAll(hr.Body)
	if err != nil {
		__antithesis_instrumentation__.Notify(14723)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14724)
	}
	__antithesis_instrumentation__.Notify(14721)
	s.mu.Lock()
	s.mu.numCalls++
	if s.mu.statusCodes[s.mu.statusCodesIndex] >= http.StatusOK && func() bool {
		__antithesis_instrumentation__.Notify(14725)
		return s.mu.statusCodes[s.mu.statusCodesIndex] < http.StatusMultipleChoices == true
	}() == true {
		__antithesis_instrumentation__.Notify(14726)
		s.mu.rows = append(s.mu.rows, string(row))
	} else {
		__antithesis_instrumentation__.Notify(14727)
	}
	__antithesis_instrumentation__.Notify(14722)
	hw.WriteHeader(s.mu.statusCodes[s.mu.statusCodesIndex])
	s.mu.statusCodesIndex = (s.mu.statusCodesIndex + 1) % len(s.mu.statusCodes)
	s.mu.Unlock()
	return nil
}
