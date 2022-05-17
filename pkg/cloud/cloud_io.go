package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

var Timeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"cloudstorage.timeout",
	"the timeout for import/export storage operations",
	10*time.Minute,
).WithPublic()

var httpCustomCA = settings.RegisterStringSetting(
	settings.TenantWritable,
	"cloudstorage.http.custom_ca",
	"custom root CA (appended to system's default CAs) for verifying certificates when interacting with HTTPS storage",
	"",
).WithPublic()

var HTTPRetryOptions = retry.Options{
	InitialBackoff: 100 * time.Millisecond,
	MaxBackoff:     2 * time.Second,
	MaxRetries:     32,
	Multiplier:     4,
}

func MakeHTTPClient(settings *cluster.Settings) (*http.Client, error) {
	__antithesis_instrumentation__.Notify(35981)
	var tlsConf *tls.Config
	if pem := httpCustomCA.Get(&settings.SV); pem != "" {
		__antithesis_instrumentation__.Notify(35983)
		roots, err := x509.SystemCertPool()
		if err != nil {
			__antithesis_instrumentation__.Notify(35986)
			return nil, errors.Wrap(err, "could not load system root CA pool")
		} else {
			__antithesis_instrumentation__.Notify(35987)
		}
		__antithesis_instrumentation__.Notify(35984)
		if !roots.AppendCertsFromPEM([]byte(pem)) {
			__antithesis_instrumentation__.Notify(35988)
			return nil, errors.Errorf("failed to parse root CA certificate from %q", pem)
		} else {
			__antithesis_instrumentation__.Notify(35989)
		}
		__antithesis_instrumentation__.Notify(35985)
		tlsConf = &tls.Config{RootCAs: roots}
	} else {
		__antithesis_instrumentation__.Notify(35990)
	}
	__antithesis_instrumentation__.Notify(35982)
	t := http.DefaultTransport.(*http.Transport).Clone()

	t.TLSClientConfig = tlsConf
	return &http.Client{Transport: t}, nil
}

const MaxDelayedRetryAttempts = 3

func DelayedRetry(
	ctx context.Context, opName string, customDelay func(error) time.Duration, fn func() error,
) error {
	__antithesis_instrumentation__.Notify(35991)
	span := tracing.SpanFromContext(ctx)
	attemptNumber := int32(1)
	return retry.WithMaxAttempts(ctx, base.DefaultRetryOptions(), MaxDelayedRetryAttempts, func() error {
		__antithesis_instrumentation__.Notify(35992)
		err := fn()
		if err == nil {
			__antithesis_instrumentation__.Notify(35996)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(35997)
		}
		__antithesis_instrumentation__.Notify(35993)
		retryEvent := &roachpb.RetryTracingEvent{
			Operation:     opName,
			AttemptNumber: attemptNumber,
			RetryError:    tracing.RedactAndTruncateError(err),
		}
		span.RecordStructured(retryEvent)
		if customDelay != nil {
			__antithesis_instrumentation__.Notify(35998)
			if d := customDelay(err); d > 0 {
				__antithesis_instrumentation__.Notify(35999)
				select {
				case <-time.After(d):
					__antithesis_instrumentation__.Notify(36000)
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(36001)
				}
			} else {
				__antithesis_instrumentation__.Notify(36002)
			}
		} else {
			__antithesis_instrumentation__.Notify(36003)
		}
		__antithesis_instrumentation__.Notify(35994)

		if strings.Contains(err.Error(), "net/http: timeout awaiting response headers") {
			__antithesis_instrumentation__.Notify(36004)
			select {
			case <-time.After(time.Second * 5):
				__antithesis_instrumentation__.Notify(36005)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(36006)
			}
		} else {
			__antithesis_instrumentation__.Notify(36007)
		}
		__antithesis_instrumentation__.Notify(35995)
		attemptNumber++
		return err
	})
}

func IsResumableHTTPError(err error) bool {
	__antithesis_instrumentation__.Notify(36008)
	return errors.Is(err, io.ErrUnexpectedEOF) || func() bool {
		__antithesis_instrumentation__.Notify(36009)
		return sysutil.IsErrConnectionReset(err) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(36010)
		return sysutil.IsErrConnectionRefused(err) == true
	}() == true
}

const maxNoProgressReads = 3

type ReaderOpenerAt func(ctx context.Context, pos int64) (io.ReadCloser, error)

type ResumingReader struct {
	Opener       ReaderOpenerAt
	Reader       io.ReadCloser
	Pos          int64
	RetryOnErrFn func(error) bool

	ErrFn func(error) time.Duration
}

var _ ioctx.ReadCloserCtx = &ResumingReader{}

func NewResumingReader(
	ctx context.Context,
	opener ReaderOpenerAt,
	reader io.ReadCloser,
	pos int64,
	retryOnErrFn func(error) bool,
	errFn func(error) time.Duration,
) *ResumingReader {
	__antithesis_instrumentation__.Notify(36011)
	r := &ResumingReader{
		Opener:       opener,
		Reader:       reader,
		Pos:          pos,
		RetryOnErrFn: retryOnErrFn,
		ErrFn:        errFn,
	}
	if r.RetryOnErrFn == nil {
		__antithesis_instrumentation__.Notify(36013)
		log.Warning(ctx, "no RetryOnErrFn specified when configuring ResumingReader, setting to default value")
		r.RetryOnErrFn = sysutil.IsErrConnectionReset
	} else {
		__antithesis_instrumentation__.Notify(36014)
	}
	__antithesis_instrumentation__.Notify(36012)
	return r
}

func (r *ResumingReader) Open(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(36015)
	return DelayedRetry(ctx, "ResumingReader.Opener", r.ErrFn, func() error {
		__antithesis_instrumentation__.Notify(36016)
		var readErr error
		r.Reader, readErr = r.Opener(ctx, r.Pos)
		return readErr
	})
}

func (r *ResumingReader) Read(ctx context.Context, p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36017)
	var lastErr error
	for retries := 0; lastErr == nil; retries++ {
		__antithesis_instrumentation__.Notify(36019)
		if r.Reader == nil {
			__antithesis_instrumentation__.Notify(36023)
			lastErr = r.Open(ctx)
		} else {
			__antithesis_instrumentation__.Notify(36024)
		}
		__antithesis_instrumentation__.Notify(36020)

		if lastErr == nil {
			__antithesis_instrumentation__.Notify(36025)
			n, readErr := r.Reader.Read(p)
			if readErr == nil || func() bool {
				__antithesis_instrumentation__.Notify(36027)
				return readErr == io.EOF == true
			}() == true {
				__antithesis_instrumentation__.Notify(36028)
				r.Pos += int64(n)
				return n, readErr
			} else {
				__antithesis_instrumentation__.Notify(36029)
			}
			__antithesis_instrumentation__.Notify(36026)
			lastErr = readErr
		} else {
			__antithesis_instrumentation__.Notify(36030)
		}
		__antithesis_instrumentation__.Notify(36021)

		if !errors.IsAny(lastErr, io.EOF, io.ErrUnexpectedEOF) {
			__antithesis_instrumentation__.Notify(36031)
			log.Errorf(ctx, "Read err: %s", lastErr)
		} else {
			__antithesis_instrumentation__.Notify(36032)
		}
		__antithesis_instrumentation__.Notify(36022)

		if r.RetryOnErrFn(lastErr) {
			__antithesis_instrumentation__.Notify(36033)
			span := tracing.SpanFromContext(ctx)
			retryEvent := &roachpb.RetryTracingEvent{
				Operation:     "ResumingReader.Reader.Read",
				AttemptNumber: int32(retries + 1),
				RetryError:    tracing.RedactAndTruncateError(lastErr),
			}
			span.RecordStructured(retryEvent)
			if retries >= maxNoProgressReads {
				__antithesis_instrumentation__.Notify(36036)
				return 0, errors.Wrap(lastErr, "multiple Read calls return no data")
			} else {
				__antithesis_instrumentation__.Notify(36037)
			}
			__antithesis_instrumentation__.Notify(36034)
			log.Errorf(ctx, "Retry IO: error %s", lastErr)
			lastErr = nil
			if r.Reader != nil {
				__antithesis_instrumentation__.Notify(36038)
				r.Reader.Close()
			} else {
				__antithesis_instrumentation__.Notify(36039)
			}
			__antithesis_instrumentation__.Notify(36035)
			r.Reader = nil
		} else {
			__antithesis_instrumentation__.Notify(36040)
		}
	}
	__antithesis_instrumentation__.Notify(36018)

	return 0, lastErr
}

func (r *ResumingReader) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(36041)
	if r.Reader != nil {
		__antithesis_instrumentation__.Notify(36043)
		return r.Reader.Close()
	} else {
		__antithesis_instrumentation__.Notify(36044)
	}
	__antithesis_instrumentation__.Notify(36042)
	return nil
}

func CheckHTTPContentRangeHeader(h string, pos int64) (int64, error) {
	__antithesis_instrumentation__.Notify(36045)
	if len(h) == 0 {
		__antithesis_instrumentation__.Notify(36052)
		return 0, errors.New("http server does not honor download resume")
	} else {
		__antithesis_instrumentation__.Notify(36053)
	}
	__antithesis_instrumentation__.Notify(36046)

	h = strings.TrimPrefix(h, "bytes ")
	dash := strings.IndexByte(h, '-')
	if dash <= 0 {
		__antithesis_instrumentation__.Notify(36054)
		return 0, errors.Errorf("malformed Content-Range header: %s", h)
	} else {
		__antithesis_instrumentation__.Notify(36055)
	}
	__antithesis_instrumentation__.Notify(36047)

	resume, err := strconv.ParseInt(h[:dash], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(36056)
		return 0, errors.Errorf("malformed start offset in Content-Range header: %s", h)
	} else {
		__antithesis_instrumentation__.Notify(36057)
	}
	__antithesis_instrumentation__.Notify(36048)

	if resume != pos {
		__antithesis_instrumentation__.Notify(36058)
		return 0, errors.Errorf(
			"expected resume position %d, found %d instead in Content-Range header: %s",
			pos, resume, h)
	} else {
		__antithesis_instrumentation__.Notify(36059)
	}
	__antithesis_instrumentation__.Notify(36049)

	slash := strings.IndexByte(h, '/')
	if slash <= 0 {
		__antithesis_instrumentation__.Notify(36060)
		return 0, errors.Errorf("malformed Content-Range header: %s", h)
	} else {
		__antithesis_instrumentation__.Notify(36061)
	}
	__antithesis_instrumentation__.Notify(36050)
	size, err := strconv.ParseInt(h[slash+1:], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(36062)
		return 0, errors.Errorf("malformed slash offset in Content-Range header: %s", h)
	} else {
		__antithesis_instrumentation__.Notify(36063)
	}
	__antithesis_instrumentation__.Notify(36051)

	return size, nil
}

func BackgroundPipe(
	ctx context.Context, fn func(ctx context.Context, pr io.Reader) error,
) io.WriteCloser {
	__antithesis_instrumentation__.Notify(36064)
	pr, pw := io.Pipe()
	w := &backgroundPipe{w: pw, grp: ctxgroup.WithContext(ctx), ctx: ctx}
	w.grp.GoCtx(func(ctc context.Context) error {
		__antithesis_instrumentation__.Notify(36066)
		err := fn(ctx, pr)
		if err != nil {
			__antithesis_instrumentation__.Notify(36068)
			closeErr := pr.CloseWithError(err)
			err = errors.CombineErrors(err, closeErr)
		} else {
			__antithesis_instrumentation__.Notify(36069)
			err = pr.Close()
		}
		__antithesis_instrumentation__.Notify(36067)
		return err
	})
	__antithesis_instrumentation__.Notify(36065)
	return w
}

type backgroundPipe struct {
	w   *io.PipeWriter
	grp ctxgroup.Group
	ctx context.Context
}

func (s *backgroundPipe) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36070)
	return s.w.Write(p)
}

func (s *backgroundPipe) Close() error {
	__antithesis_instrumentation__.Notify(36071)
	err := s.w.CloseWithError(s.ctx.Err())
	return errors.CombineErrors(err, s.grp.Wait())
}

func WriteFile(ctx context.Context, dest ExternalStorage, basename string, src io.Reader) error {
	__antithesis_instrumentation__.Notify(36072)
	var span *tracing.Span
	ctx, span = tracing.ChildSpan(ctx, fmt.Sprintf("%s.WriteFile", dest.Conf().Provider.String()))
	defer span.Finish()

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	w, err := dest.Writer(ctx, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36075)
		return errors.Wrap(err, "opening object for writing")
	} else {
		__antithesis_instrumentation__.Notify(36076)
	}
	__antithesis_instrumentation__.Notify(36073)
	if _, err := io.Copy(w, src); err != nil {
		__antithesis_instrumentation__.Notify(36077)
		cancel()
		return errors.CombineErrors(w.Close(), err)
	} else {
		__antithesis_instrumentation__.Notify(36078)
	}
	__antithesis_instrumentation__.Notify(36074)
	return errors.Wrap(w.Close(), "closing object")
}
