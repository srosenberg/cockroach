package cdctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/linkedin/goavro/v2"
)

type SchemaRegistry struct {
	server *httptest.Server
	mu     struct {
		syncutil.Mutex
		idAlloc  int32
		schemas  map[int32]string
		subjects map[string]int32
	}
}

func StartTestSchemaRegistry() *SchemaRegistry {
	__antithesis_instrumentation__.Notify(14954)
	r := makeTestSchemaRegistry()
	r.server.Start()
	return r
}

func StartTestSchemaRegistryWithTLS(certificate *tls.Certificate) (*SchemaRegistry, error) {
	__antithesis_instrumentation__.Notify(14955)
	r := makeTestSchemaRegistry()
	if certificate != nil {
		__antithesis_instrumentation__.Notify(14957)
		r.server.TLS = &tls.Config{
			Certificates: []tls.Certificate{*certificate},
		}
	} else {
		__antithesis_instrumentation__.Notify(14958)
	}
	__antithesis_instrumentation__.Notify(14956)
	r.server.StartTLS()
	return r, nil
}

func makeTestSchemaRegistry() *SchemaRegistry {
	__antithesis_instrumentation__.Notify(14959)
	r := &SchemaRegistry{}
	r.mu.schemas = make(map[int32]string)
	r.mu.subjects = make(map[string]int32)
	r.server = httptest.NewUnstartedServer(http.HandlerFunc(r.requestHandler))
	return r
}

func (r *SchemaRegistry) Close() {
	__antithesis_instrumentation__.Notify(14960)
	r.server.Close()
}

func (r *SchemaRegistry) URL() string {
	__antithesis_instrumentation__.Notify(14961)
	return r.server.URL
}

func (r *SchemaRegistry) Subjects() (subjects []string) {
	__antithesis_instrumentation__.Notify(14962)
	r.mu.Lock()
	defer r.mu.Unlock()
	for subject := range r.mu.subjects {
		__antithesis_instrumentation__.Notify(14964)
		subjects = append(subjects, subject)
	}
	__antithesis_instrumentation__.Notify(14963)
	return
}

func (r *SchemaRegistry) SchemaForSubject(subject string) string {
	__antithesis_instrumentation__.Notify(14965)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.mu.schemas[r.mu.subjects[subject]]
}

func (r *SchemaRegistry) registerSchema(subject string, schema string) int32 {
	__antithesis_instrumentation__.Notify(14966)
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.mu.idAlloc
	r.mu.idAlloc++
	r.mu.schemas[id] = schema
	r.mu.subjects[subject] = id
	return id
}

var (
	subjectVersionsRegexp = regexp.MustCompile("^/subjects/[^/]+/versions$")
)

func (r *SchemaRegistry) requestHandler(hw http.ResponseWriter, hr *http.Request) {
	__antithesis_instrumentation__.Notify(14967)
	path := hr.URL.Path
	method := hr.Method

	var err error
	switch {
	case method == http.MethodPost && func() bool {
		__antithesis_instrumentation__.Notify(14972)
		return subjectVersionsRegexp.MatchString(path) == true
	}() == true:
		__antithesis_instrumentation__.Notify(14969)
		err = r.register(hw, hr)
	case method == http.MethodGet && func() bool {
		__antithesis_instrumentation__.Notify(14973)
		return path == "/mode" == true
	}() == true:
		__antithesis_instrumentation__.Notify(14970)
		err = r.mode(hw, hr)
	default:
		__antithesis_instrumentation__.Notify(14971)
		hw.WriteHeader(http.StatusNotFound)
		return
	}
	__antithesis_instrumentation__.Notify(14968)
	if err != nil {
		__antithesis_instrumentation__.Notify(14974)
		http.Error(hw, err.Error(), http.StatusInternalServerError)
	} else {
		__antithesis_instrumentation__.Notify(14975)
	}
}

func (r *SchemaRegistry) register(hw http.ResponseWriter, hr *http.Request) (err error) {
	__antithesis_instrumentation__.Notify(14976)
	type confluentSchemaVersionRequest struct {
		Schema string `json:"schema"`
	}
	type confluentSchemaVersionResponse struct {
		ID int32 `json:"id"`
	}

	defer func() {
		__antithesis_instrumentation__.Notify(14980)
		err = hr.Body.Close()
	}()
	__antithesis_instrumentation__.Notify(14977)

	var req confluentSchemaVersionRequest
	if err := json.NewDecoder(hr.Body).Decode(&req); err != nil {
		__antithesis_instrumentation__.Notify(14981)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14982)
	}
	__antithesis_instrumentation__.Notify(14978)

	subject := strings.Split(hr.URL.Path, "/")[2]
	id := r.registerSchema(subject, req.Schema)
	res, err := json.Marshal(confluentSchemaVersionResponse{ID: id})
	if err != nil {
		__antithesis_instrumentation__.Notify(14983)
		return err
	} else {
		__antithesis_instrumentation__.Notify(14984)
	}
	__antithesis_instrumentation__.Notify(14979)

	hw.Header().Set(`Content-type`, `application/json`)
	_, err = hw.Write(res)
	return err
}

func (r *SchemaRegistry) mode(hw http.ResponseWriter, _ *http.Request) error {
	__antithesis_instrumentation__.Notify(14985)
	_, err := hw.Write([]byte("{}"))
	return err
}

func (r *SchemaRegistry) EncodedAvroToNative(b []byte) (interface{}, error) {
	__antithesis_instrumentation__.Notify(14986)
	if len(b) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(14990)
		return b[0] != changefeedbase.ConfluentAvroWireFormatMagic == true
	}() == true {
		__antithesis_instrumentation__.Notify(14991)
		return ``, errors.Errorf(`bad magic byte`)
	} else {
		__antithesis_instrumentation__.Notify(14992)
	}
	__antithesis_instrumentation__.Notify(14987)
	b = b[1:]
	if len(b) < 4 {
		__antithesis_instrumentation__.Notify(14993)
		return ``, errors.Errorf(`missing registry id`)
	} else {
		__antithesis_instrumentation__.Notify(14994)
	}
	__antithesis_instrumentation__.Notify(14988)
	id := int32(binary.BigEndian.Uint32(b[:4]))
	b = b[4:]

	r.mu.Lock()
	jsonSchema := r.mu.schemas[id]
	r.mu.Unlock()
	codec, err := goavro.NewCodec(jsonSchema)
	if err != nil {
		__antithesis_instrumentation__.Notify(14995)
		return ``, err
	} else {
		__antithesis_instrumentation__.Notify(14996)
	}
	__antithesis_instrumentation__.Notify(14989)
	native, _, err := codec.NativeFromBinary(b)
	return native, err
}

func (r *SchemaRegistry) AvroToJSON(avroBytes []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(14997)
	if len(avroBytes) == 0 {
		__antithesis_instrumentation__.Notify(15000)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(15001)
	}
	__antithesis_instrumentation__.Notify(14998)
	native, err := r.EncodedAvroToNative(avroBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(15002)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(15003)
	}
	__antithesis_instrumentation__.Notify(14999)

	return json.Marshal(native)
}
