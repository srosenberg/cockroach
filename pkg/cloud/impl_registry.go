package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"math"
	"net/url"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/errors"
)

var redactedQueryParams = map[string]struct{}{}

var confParsers = map[string]ExternalStorageURIParser{}

var implementations = map[roachpb.ExternalStorageProvider]ExternalStorageConstructor{}

type rateAndBurstSettings struct {
	rate  *settings.ByteSizeSetting
	burst *settings.ByteSizeSetting
}

type readAndWriteSettings struct {
	read, write rateAndBurstSettings
}

var limiterSettings = map[roachpb.ExternalStorageProvider]readAndWriteSettings{}

func RegisterExternalStorageProvider(
	providerType roachpb.ExternalStorageProvider,
	parseFn ExternalStorageURIParser,
	constructFn ExternalStorageConstructor,
	redactedParams map[string]struct{},
	schemes ...string,
) {
	__antithesis_instrumentation__.Notify(36441)
	for _, scheme := range schemes {
		__antithesis_instrumentation__.Notify(36445)
		if _, ok := confParsers[scheme]; ok {
			__antithesis_instrumentation__.Notify(36447)
			panic(fmt.Sprintf("external storage provider already registered for %s", scheme))
		} else {
			__antithesis_instrumentation__.Notify(36448)
		}
		__antithesis_instrumentation__.Notify(36446)
		confParsers[scheme] = parseFn
		for param := range redactedParams {
			__antithesis_instrumentation__.Notify(36449)
			redactedQueryParams[param] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(36442)
	if _, ok := implementations[providerType]; ok {
		__antithesis_instrumentation__.Notify(36450)
		panic(fmt.Sprintf("external storage provider already registered for %s", providerType.String()))
	} else {
		__antithesis_instrumentation__.Notify(36451)
	}
	__antithesis_instrumentation__.Notify(36443)
	implementations[providerType] = constructFn

	sinkName := strings.ToLower(providerType.String())
	if sinkName == "null" {
		__antithesis_instrumentation__.Notify(36452)
		sinkName = "nullsink"
	} else {
		__antithesis_instrumentation__.Notify(36453)
	}
	__antithesis_instrumentation__.Notify(36444)

	readRateName := fmt.Sprintf("cloudstorage.%s.read.node_rate_limit", sinkName)
	readBurstName := fmt.Sprintf("cloudstorage.%s.read.node_burst_limit", sinkName)
	writeRateName := fmt.Sprintf("cloudstorage.%s.write.node_rate_limit", sinkName)
	writeBurstName := fmt.Sprintf("cloudstorage.%s.write.node_burst_limit", sinkName)

	limiterSettings[providerType] = readAndWriteSettings{
		read: rateAndBurstSettings{
			rate: settings.RegisterByteSizeSetting(settings.TenantWritable, readRateName,
				"limit on number of bytes per second per node across operations writing to the designated cloud storage provider if non-zero",
				0, settings.NonNegativeInt,
			),
			burst: settings.RegisterByteSizeSetting(settings.TenantWritable, readBurstName,
				"burst limit on number of bytes per second per node across operations writing to the designated cloud storage provider if non-zero",
				0, settings.NonNegativeInt,
			),
		},
		write: rateAndBurstSettings{
			rate: settings.RegisterByteSizeSetting(settings.TenantWritable, writeRateName,
				"limit on number of bytes per second per node across operations writing to the designated cloud storage provider if non-zero",
				0, settings.NonNegativeInt,
			),
			burst: settings.RegisterByteSizeSetting(settings.TenantWritable, writeBurstName,
				"burst limit on number of bytes per second per node across operations writing to the designated cloud storage provider if non-zero",
				0, settings.NonNegativeInt,
			),
		},
	}
}

func ExternalStorageConfFromURI(
	path string, user security.SQLUsername,
) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36454)
	uri, err := url.Parse(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(36457)
		return roachpb.ExternalStorage{}, err
	} else {
		__antithesis_instrumentation__.Notify(36458)
	}
	__antithesis_instrumentation__.Notify(36455)
	if fn, ok := confParsers[uri.Scheme]; ok {
		__antithesis_instrumentation__.Notify(36459)
		return fn(ExternalStorageURIContext{CurrentUser: user}, uri)
	} else {
		__antithesis_instrumentation__.Notify(36460)
	}
	__antithesis_instrumentation__.Notify(36456)

	return roachpb.ExternalStorage{}, errors.Errorf("unsupported storage scheme: %q - refer to docs to find supported"+
		" storage schemes", uri.Scheme)
}

func ExternalStorageFromURI(
	ctx context.Context,
	uri string,
	externalConfig base.ExternalIODirConfig,
	settings *cluster.Settings,
	blobClientFactory blobs.BlobClientFactory,
	user security.SQLUsername,
	ie sqlutil.InternalExecutor,
	kvDB *kv.DB,
	limiters Limiters,
) (ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36461)
	conf, err := ExternalStorageConfFromURI(uri, user)
	if err != nil {
		__antithesis_instrumentation__.Notify(36463)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36464)
	}
	__antithesis_instrumentation__.Notify(36462)
	return MakeExternalStorage(ctx, conf, externalConfig, settings, blobClientFactory, ie, kvDB, limiters)
}

func SanitizeExternalStorageURI(path string, extraParams []string) (string, error) {
	__antithesis_instrumentation__.Notify(36465)
	uri, err := url.Parse(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(36469)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(36470)
	}
	__antithesis_instrumentation__.Notify(36466)
	if uri.Scheme == "experimental-workload" || func() bool {
		__antithesis_instrumentation__.Notify(36471)
		return uri.Scheme == "workload" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(36472)
		return uri.Scheme == "null" == true
	}() == true {
		__antithesis_instrumentation__.Notify(36473)
		return path, nil
	} else {
		__antithesis_instrumentation__.Notify(36474)
	}
	__antithesis_instrumentation__.Notify(36467)

	params := uri.Query()
	for param := range params {
		__antithesis_instrumentation__.Notify(36475)
		if _, ok := redactedQueryParams[param]; ok {
			__antithesis_instrumentation__.Notify(36476)
			params.Set(param, "redacted")
		} else {
			__antithesis_instrumentation__.Notify(36477)
			for _, p := range extraParams {
				__antithesis_instrumentation__.Notify(36478)
				if param == p {
					__antithesis_instrumentation__.Notify(36479)
					params.Set(param, "redacted")
				} else {
					__antithesis_instrumentation__.Notify(36480)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(36468)

	uri.RawQuery = params.Encode()
	return uri.String(), nil
}

func MakeExternalStorage(
	ctx context.Context,
	dest roachpb.ExternalStorage,
	conf base.ExternalIODirConfig,
	settings *cluster.Settings,
	blobClientFactory blobs.BlobClientFactory,
	ie sqlutil.InternalExecutor,
	kvDB *kv.DB,
	limiters Limiters,
) (ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36481)
	args := ExternalStorageContext{
		IOConf:            conf,
		Settings:          settings,
		BlobClientFactory: blobClientFactory,
		InternalExecutor:  ie,
		DB:                kvDB,
	}
	if conf.DisableOutbound && func() bool {
		__antithesis_instrumentation__.Notify(36484)
		return dest.Provider != roachpb.ExternalStorageProvider_userfile == true
	}() == true {
		__antithesis_instrumentation__.Notify(36485)
		return nil, errors.New("external network access is disabled")
	} else {
		__antithesis_instrumentation__.Notify(36486)
	}
	__antithesis_instrumentation__.Notify(36482)
	if fn, ok := implementations[dest.Provider]; ok {
		__antithesis_instrumentation__.Notify(36487)
		e, err := fn(ctx, args, dest)
		if err != nil {
			__antithesis_instrumentation__.Notify(36490)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(36491)
		}
		__antithesis_instrumentation__.Notify(36488)
		if l, ok := limiters[dest.Provider]; ok {
			__antithesis_instrumentation__.Notify(36492)
			return &limitWrapper{ExternalStorage: e, lim: l}, nil
		} else {
			__antithesis_instrumentation__.Notify(36493)
		}
		__antithesis_instrumentation__.Notify(36489)
		return e, nil
	} else {
		__antithesis_instrumentation__.Notify(36494)
	}
	__antithesis_instrumentation__.Notify(36483)
	return nil, errors.Errorf("unsupported external destination type: %s", dest.Provider.String())
}

type rwLimiter struct {
	read, write *quotapool.RateLimiter
}

type Limiters map[roachpb.ExternalStorageProvider]rwLimiter

func makeLimiter(
	ctx context.Context, sv *settings.Values, s rateAndBurstSettings,
) *quotapool.RateLimiter {
	__antithesis_instrumentation__.Notify(36495)
	lim := quotapool.NewRateLimiter(s.rate.Key(), quotapool.Limit(0), 0)
	fn := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(36497)
		rate := quotapool.Limit(s.rate.Get(sv))
		if rate == 0 {
			__antithesis_instrumentation__.Notify(36500)
			rate = quotapool.Limit(math.Inf(1))
		} else {
			__antithesis_instrumentation__.Notify(36501)
		}
		__antithesis_instrumentation__.Notify(36498)
		burst := s.burst.Get(sv)
		if burst == 0 {
			__antithesis_instrumentation__.Notify(36502)
			burst = math.MaxInt64
		} else {
			__antithesis_instrumentation__.Notify(36503)
		}
		__antithesis_instrumentation__.Notify(36499)
		lim.UpdateLimit(rate, burst)
	}
	__antithesis_instrumentation__.Notify(36496)
	s.rate.SetOnChange(sv, fn)
	s.burst.SetOnChange(sv, fn)
	fn(ctx)
	return lim
}

func MakeLimiters(ctx context.Context, sv *settings.Values) Limiters {
	__antithesis_instrumentation__.Notify(36504)
	m := make(Limiters, len(limiterSettings))
	for k := range limiterSettings {
		__antithesis_instrumentation__.Notify(36506)
		l := limiterSettings[k]
		m[k] = rwLimiter{read: makeLimiter(ctx, sv, l.read), write: makeLimiter(ctx, sv, l.write)}
	}
	__antithesis_instrumentation__.Notify(36505)
	return m
}

type limitWrapper struct {
	ExternalStorage
	lim rwLimiter
}

func (l *limitWrapper) ReadFile(ctx context.Context, basename string) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(36507)
	r, err := l.ExternalStorage.ReadFile(ctx, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36509)
		return r, err
	} else {
		__antithesis_instrumentation__.Notify(36510)
	}
	__antithesis_instrumentation__.Notify(36508)

	return &limitedReader{r: r, lim: l.lim.read}, nil
}

func (l *limitWrapper) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36511)
	r, s, err := l.ExternalStorage.ReadFileAt(ctx, basename, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(36513)
		return r, s, err
	} else {
		__antithesis_instrumentation__.Notify(36514)
	}
	__antithesis_instrumentation__.Notify(36512)

	return &limitedReader{r: r, lim: l.lim.read}, s, nil
}

func (l *limitWrapper) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(36515)
	w, err := l.ExternalStorage.Writer(ctx, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36517)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36518)
	}
	__antithesis_instrumentation__.Notify(36516)

	return &limitedWriter{w: w, ctx: ctx, lim: l.lim.write}, nil
}

type limitedReader struct {
	r    ioctx.ReadCloserCtx
	lim  *quotapool.RateLimiter
	pool int64
}

func (l *limitedReader) Read(ctx context.Context, p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36519)
	n, err := l.r.Read(ctx, p)

	l.pool += int64(n)
	const batchedWriteLimit = 128 << 10
	if l.pool > batchedWriteLimit {
		__antithesis_instrumentation__.Notify(36521)
		if err := l.lim.WaitN(ctx, l.pool); err != nil {
			__antithesis_instrumentation__.Notify(36523)
			log.Warningf(ctx, "failed to throttle write: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(36524)
		}
		__antithesis_instrumentation__.Notify(36522)
		l.pool = 0
	} else {
		__antithesis_instrumentation__.Notify(36525)
	}
	__antithesis_instrumentation__.Notify(36520)
	return n, err
}

func (l *limitedReader) Close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(36526)
	if err := l.lim.WaitN(ctx, l.pool); err != nil {
		__antithesis_instrumentation__.Notify(36528)
		log.Warningf(ctx, "failed to throttle closing write: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(36529)
	}
	__antithesis_instrumentation__.Notify(36527)
	return l.r.Close(ctx)
}

type limitedWriter struct {
	w    io.WriteCloser
	ctx  context.Context
	lim  *quotapool.RateLimiter
	pool int64
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(36530)

	l.pool += int64(len(p))
	const batchedWriteLimit = 128 << 10
	if l.pool > batchedWriteLimit {
		__antithesis_instrumentation__.Notify(36532)
		if err := l.lim.WaitN(l.ctx, l.pool); err != nil {
			__antithesis_instrumentation__.Notify(36534)
			log.Warningf(l.ctx, "failed to throttle write: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(36535)
		}
		__antithesis_instrumentation__.Notify(36533)
		l.pool = 0
	} else {
		__antithesis_instrumentation__.Notify(36536)
	}
	__antithesis_instrumentation__.Notify(36531)
	n, err := l.w.Write(p)
	return n, err
}

func (l *limitedWriter) Close() error {
	__antithesis_instrumentation__.Notify(36537)
	if err := l.lim.WaitN(l.ctx, l.pool); err != nil {
		__antithesis_instrumentation__.Notify(36539)
		log.Warningf(l.ctx, "failed to throttle closing write: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(36540)
	}
	__antithesis_instrumentation__.Notify(36538)
	return l.w.Close()
}
