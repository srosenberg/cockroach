package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"path"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

const confluentSchemaContentType = `application/vnd.schemaregistry.v1+json`

type schemaRegistry interface {
	Ping(ctx context.Context) error

	RegisterSchemaForSubject(ctx context.Context, subject string, schema string) (int32, error)
}

type confluentSchemaVersionRequest struct {
	Schema string `json:"schema"`
}

type confluentSchemaVersionResponse struct {
	ID int32 `json:"id"`
}

type confluentSchemaRegistry struct {
	baseURL *url.URL

	client    *httputil.Client
	retryOpts retry.Options
}

var _ schemaRegistry = (*confluentSchemaRegistry)(nil)

func newConfluentSchemaRegistry(baseURL string) (*confluentSchemaRegistry, error) {
	__antithesis_instrumentation__.Notify(17644)
	u, err := url.Parse(baseURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(17649)
		return nil, errors.Wrap(err, "malformed schema registry url")
	} else {
		__antithesis_instrumentation__.Notify(17650)
	}
	__antithesis_instrumentation__.Notify(17645)

	if u.Scheme != "http" && func() bool {
		__antithesis_instrumentation__.Notify(17651)
		return u.Scheme != "https" == true
	}() == true {
		__antithesis_instrumentation__.Notify(17652)
		return nil, errors.Errorf("unsupported scheme: %q", u.Scheme)
	} else {
		__antithesis_instrumentation__.Notify(17653)
	}
	__antithesis_instrumentation__.Notify(17646)

	query := u.Query()
	var caCert []byte
	if caCertString := query.Get(changefeedbase.RegistryParamCACert); caCertString != "" {
		__antithesis_instrumentation__.Notify(17654)
		err := decodeBase64FromString(caCertString, &caCert)
		if err != nil {
			__antithesis_instrumentation__.Notify(17655)
			return nil, errors.Wrapf(err, "param %s must be base 64 encoded", changefeedbase.RegistryParamCACert)
		} else {
			__antithesis_instrumentation__.Notify(17656)
		}
	} else {
		__antithesis_instrumentation__.Notify(17657)
	}
	__antithesis_instrumentation__.Notify(17647)

	query.Del(changefeedbase.RegistryParamCACert)
	u.RawQuery = query.Encode()

	httpClient, err := setupHTTPClient(u, caCert)
	if err != nil {
		__antithesis_instrumentation__.Notify(17658)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(17659)
	}
	__antithesis_instrumentation__.Notify(17648)

	retryOpts := base.DefaultRetryOptions()
	retryOpts.MaxRetries = 2
	return &confluentSchemaRegistry{
		baseURL:   u,
		client:    httpClient,
		retryOpts: retryOpts,
	}, nil
}

func setupHTTPClient(baseURL *url.URL, caCert []byte) (*httputil.Client, error) {
	__antithesis_instrumentation__.Notify(17660)
	if caCert != nil {
		__antithesis_instrumentation__.Notify(17662)
		httpClient, err := newClientFromTLSKeyPair(caCert)
		if err != nil {
			__antithesis_instrumentation__.Notify(17665)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(17666)
		}
		__antithesis_instrumentation__.Notify(17663)
		if baseURL.Scheme == "http" {
			__antithesis_instrumentation__.Notify(17667)
			log.Warningf(context.Background(), "CA certificate provided but schema registry %s uses HTTP", baseURL)
		} else {
			__antithesis_instrumentation__.Notify(17668)
		}
		__antithesis_instrumentation__.Notify(17664)
		return httpClient, nil
	} else {
		__antithesis_instrumentation__.Notify(17669)
	}
	__antithesis_instrumentation__.Notify(17661)
	return httputil.DefaultClient, nil
}

func (r *confluentSchemaRegistry) Ping(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(17670)
	u := r.urlForPath("mode")
	return r.doWithRetry(ctx, func() error {
		__antithesis_instrumentation__.Notify(17671)
		resp, err := r.client.Get(ctx, u)
		if err != nil {
			__antithesis_instrumentation__.Notify(17674)
			return err
		} else {
			__antithesis_instrumentation__.Notify(17675)
		}
		__antithesis_instrumentation__.Notify(17672)
		defer gracefulClose(ctx, resp.Body)

		if resp.StatusCode >= 500 {
			__antithesis_instrumentation__.Notify(17676)
			return errors.Errorf("unexpected schema registry response: %s", resp.Status)
		} else {
			__antithesis_instrumentation__.Notify(17677)
		}
		__antithesis_instrumentation__.Notify(17673)
		return nil
	})
}

func (r *confluentSchemaRegistry) RegisterSchemaForSubject(
	ctx context.Context, subject string, schema string,
) (int32, error) {
	__antithesis_instrumentation__.Notify(17678)
	u := r.urlForPath(fmt.Sprintf("subjects/%s/versions", subject))
	if log.V(1) {
		__antithesis_instrumentation__.Notify(17683)
		log.Infof(ctx, "registering avro schema %s %s", u, schema)
	} else {
		__antithesis_instrumentation__.Notify(17684)
	}
	__antithesis_instrumentation__.Notify(17679)

	req := confluentSchemaVersionRequest{Schema: schema}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(req); err != nil {
		__antithesis_instrumentation__.Notify(17685)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(17686)
	}
	__antithesis_instrumentation__.Notify(17680)

	var id int32
	err := r.doWithRetry(ctx, func() error {
		__antithesis_instrumentation__.Notify(17687)
		resp, err := r.client.Post(ctx, u, confluentSchemaContentType, &buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(17691)
			return errors.Wrap(err, "contacting confluent schema registry")
		} else {
			__antithesis_instrumentation__.Notify(17692)
		}
		__antithesis_instrumentation__.Notify(17688)
		defer gracefulClose(ctx, resp.Body)
		if resp.StatusCode < 200 || func() bool {
			__antithesis_instrumentation__.Notify(17693)
			return resp.StatusCode >= 300 == true
		}() == true {
			__antithesis_instrumentation__.Notify(17694)
			body, _ := ioutil.ReadAll(resp.Body)
			return errors.Errorf("registering schema to %s %s: %s", u, resp.Status, body)
		} else {
			__antithesis_instrumentation__.Notify(17695)
		}
		__antithesis_instrumentation__.Notify(17689)
		var res confluentSchemaVersionResponse
		if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
			__antithesis_instrumentation__.Notify(17696)
			return errors.Wrap(err, "decoding confluent schema registry reply")
		} else {
			__antithesis_instrumentation__.Notify(17697)
		}
		__antithesis_instrumentation__.Notify(17690)
		id = res.ID
		return nil
	})
	__antithesis_instrumentation__.Notify(17681)
	if err != nil {
		__antithesis_instrumentation__.Notify(17698)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(17699)
	}
	__antithesis_instrumentation__.Notify(17682)
	return id, nil
}

func (r *confluentSchemaRegistry) doWithRetry(ctx context.Context, fn func() error) error {
	__antithesis_instrumentation__.Notify(17700)

	var err error
	for retrier := retry.StartWithCtx(ctx, r.retryOpts); retrier.Next(); {
		__antithesis_instrumentation__.Notify(17702)
		err = fn()
		if err == nil {
			__antithesis_instrumentation__.Notify(17704)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(17705)
		}
		__antithesis_instrumentation__.Notify(17703)
		log.VInfof(ctx, 2, "retrying schema registry operation: %s", err.Error())
	}
	__antithesis_instrumentation__.Notify(17701)
	return changefeedbase.MarkRetryableError(err)
}

func gracefulClose(ctx context.Context, toClose io.ReadCloser) {
	__antithesis_instrumentation__.Notify(17706)

	const respExtraReadLimit = 4096
	_, _ = io.CopyN(ioutil.Discard, toClose, respExtraReadLimit)
	if err := toClose.Close(); err != nil {
		__antithesis_instrumentation__.Notify(17707)
		log.VInfof(ctx, 2, "failure to close schema registry connection", err)
	} else {
		__antithesis_instrumentation__.Notify(17708)
	}
}

func (r *confluentSchemaRegistry) urlForPath(relPath string) string {
	__antithesis_instrumentation__.Notify(17709)
	u := *r.baseURL
	u.Path = path.Join(u.EscapedPath(), relPath)
	return u.String()
}
