package httpsink

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

func parseHTTPURL(
	_ cloud.ExternalStorageURIContext, uri *url.URL,
) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36348)
	conf := roachpb.ExternalStorage{}
	conf.Provider = roachpb.ExternalStorageProvider_http
	conf.HttpPath.BaseUri = uri.String()
	return conf, nil
}

type httpStorage struct {
	base     *url.URL
	client   *http.Client
	hosts    []string
	settings *cluster.Settings
	ioConf   base.ExternalIODirConfig
}

var _ cloud.ExternalStorage = &httpStorage{}

type retryableHTTPError struct {
	cause error
}

func (e *retryableHTTPError) Error() string {
	__antithesis_instrumentation__.Notify(36349)
	return fmt.Sprintf("retryable http error: %s", e.cause)
}

func MakeHTTPStorage(
	ctx context.Context, args cloud.ExternalStorageContext, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36350)
	telemetry.Count("external-io.http")
	if args.IOConf.DisableHTTP {
		__antithesis_instrumentation__.Notify(36355)
		return nil, errors.New("external http access disabled")
	} else {
		__antithesis_instrumentation__.Notify(36356)
	}
	__antithesis_instrumentation__.Notify(36351)
	base := dest.HttpPath.BaseUri
	if base == "" {
		__antithesis_instrumentation__.Notify(36357)
		return nil, errors.Errorf("HTTP storage requested but prefix path not provided")
	} else {
		__antithesis_instrumentation__.Notify(36358)
	}
	__antithesis_instrumentation__.Notify(36352)

	client, err := cloud.MakeHTTPClient(args.Settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(36359)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36360)
	}
	__antithesis_instrumentation__.Notify(36353)
	uri, err := url.Parse(base)
	if err != nil {
		__antithesis_instrumentation__.Notify(36361)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36362)
	}
	__antithesis_instrumentation__.Notify(36354)
	return &httpStorage{
		base:     uri,
		client:   client,
		hosts:    strings.Split(uri.Host, ","),
		settings: args.Settings,
		ioConf:   args.IOConf,
	}, nil
}

func (h *httpStorage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(36363)
	return roachpb.ExternalStorage{
		Provider: roachpb.ExternalStorageProvider_http,
		HttpPath: roachpb.ExternalStorage_Http{
			BaseUri: h.base.String(),
		},
	}
}

func (h *httpStorage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(36364)
	return h.ioConf
}

func (h *httpStorage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(36365)
	return h.settings
}

func (h *httpStorage) ReadFile(ctx context.Context, basename string) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(36366)

	stream, _, err := h.ReadFileAt(ctx, basename, 0)
	return stream, err
}

func (h *httpStorage) openStreamAt(
	ctx context.Context, url string, pos int64,
) (*http.Response, error) {
	__antithesis_instrumentation__.Notify(36367)
	var headers map[string]string
	if pos > 0 {
		__antithesis_instrumentation__.Notify(36371)
		headers = map[string]string{"Range": fmt.Sprintf("bytes=%d-", pos)}
	} else {
		__antithesis_instrumentation__.Notify(36372)
	}
	__antithesis_instrumentation__.Notify(36368)

	for attempt, retries := 0, retry.StartWithCtx(ctx, cloud.HTTPRetryOptions); retries.Next(); attempt++ {
		__antithesis_instrumentation__.Notify(36373)
		resp, err := h.req(ctx, "GET", url, nil, headers)
		if err == nil {
			__antithesis_instrumentation__.Notify(36375)
			return resp, err
		} else {
			__antithesis_instrumentation__.Notify(36376)
		}
		__antithesis_instrumentation__.Notify(36374)

		log.Errorf(ctx, "HTTP:Req error: err=%s (attempt %d)", err, attempt)

		if !errors.HasType(err, (*retryableHTTPError)(nil)) {
			__antithesis_instrumentation__.Notify(36377)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(36378)
		}
	}
	__antithesis_instrumentation__.Notify(36369)
	if ctx.Err() == nil {
		__antithesis_instrumentation__.Notify(36379)
		return nil, errors.New("too many retries; giving up")
	} else {
		__antithesis_instrumentation__.Notify(36380)
	}
	__antithesis_instrumentation__.Notify(36370)

	return nil, ctx.Err()
}

func (h *httpStorage) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36381)
	stream, err := h.openStreamAt(ctx, basename, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(36385)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(36386)
	}
	__antithesis_instrumentation__.Notify(36382)

	var size int64
	if offset == 0 {
		__antithesis_instrumentation__.Notify(36387)
		size = stream.ContentLength
	} else {
		__antithesis_instrumentation__.Notify(36388)
		size, err = cloud.CheckHTTPContentRangeHeader(stream.Header.Get("Content-Range"), offset)
		if err != nil {
			__antithesis_instrumentation__.Notify(36389)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(36390)
		}
	}
	__antithesis_instrumentation__.Notify(36383)

	canResume := stream.Header.Get("Accept-Ranges") == "bytes"
	if canResume {
		__antithesis_instrumentation__.Notify(36391)
		opener := func(ctx context.Context, pos int64) (io.ReadCloser, error) {
			__antithesis_instrumentation__.Notify(36393)
			s, err := h.openStreamAt(ctx, basename, pos)
			if err != nil {
				__antithesis_instrumentation__.Notify(36395)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(36396)
			}
			__antithesis_instrumentation__.Notify(36394)
			return s.Body, err
		}
		__antithesis_instrumentation__.Notify(36392)
		return cloud.NewResumingReader(ctx, opener, stream.Body, offset,
			cloud.IsResumableHTTPError, nil), size, nil
	} else {
		__antithesis_instrumentation__.Notify(36397)
	}
	__antithesis_instrumentation__.Notify(36384)
	return ioctx.ReadCloserAdapter(stream.Body), size, nil
}

func (h *httpStorage) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(36398)
	return cloud.BackgroundPipe(ctx, func(ctx context.Context, r io.Reader) error {
		__antithesis_instrumentation__.Notify(36399)
		_, err := h.reqNoBody(ctx, "PUT", basename, r)
		return err
	}), nil
}

func (h *httpStorage) List(_ context.Context, _, _ string, _ cloud.ListingFn) error {
	__antithesis_instrumentation__.Notify(36400)
	return errors.Mark(errors.New("http storage does not support listing"), cloud.ErrListingUnsupported)
}

func (h *httpStorage) Delete(ctx context.Context, basename string) error {
	__antithesis_instrumentation__.Notify(36401)
	return contextutil.RunWithTimeout(ctx, fmt.Sprintf("DELETE %s", basename),
		cloud.Timeout.Get(&h.settings.SV), func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(36402)
			_, err := h.reqNoBody(ctx, "DELETE", basename, nil)
			return err
		})
}

func (h *httpStorage) Size(ctx context.Context, basename string) (int64, error) {
	__antithesis_instrumentation__.Notify(36403)
	var resp *http.Response
	if err := contextutil.RunWithTimeout(ctx, fmt.Sprintf("HEAD %s", basename),
		cloud.Timeout.Get(&h.settings.SV), func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(36406)
			var err error
			resp, err = h.reqNoBody(ctx, "HEAD", basename, nil)
			return err
		}); err != nil {
		__antithesis_instrumentation__.Notify(36407)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(36408)
	}
	__antithesis_instrumentation__.Notify(36404)
	if resp.ContentLength < 0 {
		__antithesis_instrumentation__.Notify(36409)
		return 0, errors.Errorf("bad ContentLength: %d", resp.ContentLength)
	} else {
		__antithesis_instrumentation__.Notify(36410)
	}
	__antithesis_instrumentation__.Notify(36405)
	return resp.ContentLength, nil
}

func (h *httpStorage) Close() error {
	__antithesis_instrumentation__.Notify(36411)
	return nil
}

func (h *httpStorage) reqNoBody(
	ctx context.Context, method, file string, body io.Reader,
) (*http.Response, error) {
	__antithesis_instrumentation__.Notify(36412)
	resp, err := h.req(ctx, method, file, body, nil)
	if resp != nil {
		__antithesis_instrumentation__.Notify(36414)
		resp.Body.Close()
	} else {
		__antithesis_instrumentation__.Notify(36415)
	}
	__antithesis_instrumentation__.Notify(36413)
	return resp, err
}

func (h *httpStorage) req(
	ctx context.Context, method, file string, body io.Reader, headers map[string]string,
) (*http.Response, error) {
	__antithesis_instrumentation__.Notify(36416)
	dest := *h.base
	if hosts := len(h.hosts); hosts > 1 {
		__antithesis_instrumentation__.Notify(36422)
		if file == "" {
			__antithesis_instrumentation__.Notify(36425)
			return nil, errors.New("cannot use a multi-host HTTP basepath for single file")
		} else {
			__antithesis_instrumentation__.Notify(36426)
		}
		__antithesis_instrumentation__.Notify(36423)
		hash := fnv.New32a()
		if _, err := hash.Write([]byte(file)); err != nil {
			__antithesis_instrumentation__.Notify(36427)
			panic(errors.Wrap(err, `"It never returns an error." -- https://golang.org/pkg/hash`))
		} else {
			__antithesis_instrumentation__.Notify(36428)
		}
		__antithesis_instrumentation__.Notify(36424)
		dest.Host = h.hosts[int(hash.Sum32())%hosts]
	} else {
		__antithesis_instrumentation__.Notify(36429)
	}
	__antithesis_instrumentation__.Notify(36417)
	dest.Path = path.Join(dest.Path, file)
	url := dest.String()
	req, err := http.NewRequest(method, url, body)

	if err != nil {
		__antithesis_instrumentation__.Notify(36430)
		return nil, errors.Wrapf(err, "error constructing request %s %q", method, url)
	} else {
		__antithesis_instrumentation__.Notify(36431)
	}
	__antithesis_instrumentation__.Notify(36418)
	req = req.WithContext(ctx)

	for key, val := range headers {
		__antithesis_instrumentation__.Notify(36432)
		req.Header.Add(key, val)
	}
	__antithesis_instrumentation__.Notify(36419)

	resp, err := h.client.Do(req)
	if err != nil {
		__antithesis_instrumentation__.Notify(36433)

		return nil, &retryableHTTPError{err}
	} else {
		__antithesis_instrumentation__.Notify(36434)
	}
	__antithesis_instrumentation__.Notify(36420)

	switch resp.StatusCode {
	case 200, 201, 204, 206:
		__antithesis_instrumentation__.Notify(36435)

	default:
		__antithesis_instrumentation__.Notify(36436)
		body, _ := ioutil.ReadAll(resp.Body)
		_ = resp.Body.Close()
		err := errors.Errorf("error response from server: %s %q", resp.Status, body)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(36438)
			return resp.StatusCode == 404 == true
		}() == true {
			__antithesis_instrumentation__.Notify(36439)

			err = errors.Wrapf(
				errors.Wrap(cloud.ErrFileDoesNotExist, "http storage file does not exist"),
				"%v",
				err.Error(),
			)
		} else {
			__antithesis_instrumentation__.Notify(36440)
		}
		__antithesis_instrumentation__.Notify(36437)
		return nil, err
	}
	__antithesis_instrumentation__.Notify(36421)
	return resp, nil
}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_http,
		parseHTTPURL, MakeHTTPStorage, cloud.RedactedParams(), "http", "https")
}
