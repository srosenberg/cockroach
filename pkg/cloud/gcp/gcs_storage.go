package gcp

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	gcs "cloud.google.com/go/storage"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const (
	GoogleBillingProjectParam = "GOOGLE_BILLING_PROJECT"

	CredentialsParam = "CREDENTIALS"
)

var gcsChunkingEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"cloudstorage.gs.chunking.enabled",
	"enable chunking of file upload to Google Cloud Storage",
	true,
)

func parseGSURL(_ cloud.ExternalStorageURIContext, uri *url.URL) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36283)
	conf := roachpb.ExternalStorage{}
	conf.Provider = roachpb.ExternalStorageProvider_gs
	conf.GoogleCloudConfig = &roachpb.ExternalStorage_GCS{
		Bucket:         uri.Host,
		Prefix:         uri.Path,
		Auth:           uri.Query().Get(cloud.AuthParam),
		BillingProject: uri.Query().Get(GoogleBillingProjectParam),
		Credentials:    uri.Query().Get(CredentialsParam),
	}
	conf.GoogleCloudConfig.Prefix = strings.TrimLeft(conf.GoogleCloudConfig.Prefix, "/")
	return conf, nil
}

type gcsStorage struct {
	bucket   *gcs.BucketHandle
	client   *gcs.Client
	conf     *roachpb.ExternalStorage_GCS
	ioConf   base.ExternalIODirConfig
	prefix   string
	settings *cluster.Settings
}

var _ cloud.ExternalStorage = &gcsStorage{}

func (g *gcsStorage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(36284)
	return roachpb.ExternalStorage{
		Provider:          roachpb.ExternalStorageProvider_gs,
		GoogleCloudConfig: g.conf,
	}
}

func (g *gcsStorage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(36285)
	return g.ioConf
}

func (g *gcsStorage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(36286)
	return g.settings
}

func makeGCSStorage(
	ctx context.Context, args cloud.ExternalStorageContext, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36287)
	telemetry.Count("external-io.google_cloud")
	conf := dest.GoogleCloudConfig
	if conf == nil {
		__antithesis_instrumentation__.Notify(36293)
		return nil, errors.Errorf("google cloud storage upload requested but info missing")
	} else {
		__antithesis_instrumentation__.Notify(36294)
	}
	__antithesis_instrumentation__.Notify(36288)
	const scope = gcs.ScopeReadWrite
	opts := []option.ClientOption{option.WithScopes(scope)}

	if args.IOConf.DisableImplicitCredentials && func() bool {
		__antithesis_instrumentation__.Notify(36295)
		return conf.Auth == cloud.AuthParamImplicit == true
	}() == true {
		__antithesis_instrumentation__.Notify(36296)
		return nil, errors.New(
			"implicit credentials disallowed for gs due to --external-io-disable-implicit-credentials flag")
	} else {
		__antithesis_instrumentation__.Notify(36297)
	}
	__antithesis_instrumentation__.Notify(36289)

	switch conf.Auth {
	case cloud.AuthParamImplicit:
		__antithesis_instrumentation__.Notify(36298)

	default:
		__antithesis_instrumentation__.Notify(36299)
		if conf.Credentials == "" {
			__antithesis_instrumentation__.Notify(36303)
			return nil, errors.Errorf(
				"%s must be set unless %q is %q",
				CredentialsParam,
				cloud.AuthParam,
				cloud.AuthParamImplicit,
			)
		} else {
			__antithesis_instrumentation__.Notify(36304)
		}
		__antithesis_instrumentation__.Notify(36300)
		decodedKey, err := base64.StdEncoding.DecodeString(conf.Credentials)
		if err != nil {
			__antithesis_instrumentation__.Notify(36305)
			return nil, errors.Wrapf(err, "decoding value of %s", CredentialsParam)
		} else {
			__antithesis_instrumentation__.Notify(36306)
		}
		__antithesis_instrumentation__.Notify(36301)
		source, err := google.JWTConfigFromJSON(decodedKey, scope)
		if err != nil {
			__antithesis_instrumentation__.Notify(36307)
			return nil, errors.Wrap(err, "creating GCS oauth token source from specified credentials")
		} else {
			__antithesis_instrumentation__.Notify(36308)
		}
		__antithesis_instrumentation__.Notify(36302)
		opts = append(opts, option.WithTokenSource(source.TokenSource(ctx)))
	}
	__antithesis_instrumentation__.Notify(36290)
	g, err := gcs.NewClient(ctx, opts...)
	if err != nil {
		__antithesis_instrumentation__.Notify(36309)
		return nil, errors.Wrap(err, "failed to create google cloud client")
	} else {
		__antithesis_instrumentation__.Notify(36310)
	}
	__antithesis_instrumentation__.Notify(36291)
	bucket := g.Bucket(conf.Bucket)
	if conf.BillingProject != `` {
		__antithesis_instrumentation__.Notify(36311)
		bucket = bucket.UserProject(conf.BillingProject)
	} else {
		__antithesis_instrumentation__.Notify(36312)
	}
	__antithesis_instrumentation__.Notify(36292)
	return &gcsStorage{
		bucket:   bucket,
		client:   g,
		conf:     conf,
		ioConf:   args.IOConf,
		prefix:   conf.Prefix,
		settings: args.Settings,
	}, nil
}

func (g *gcsStorage) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(36313)
	_, sp := tracing.ChildSpan(ctx, "gcs.Writer")
	defer sp.Finish()
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("gcs.Writer: %s",
		path.Join(g.prefix, basename))})

	w := g.bucket.Object(path.Join(g.prefix, basename)).NewWriter(ctx)
	if !gcsChunkingEnabled.Get(&g.settings.SV) {
		__antithesis_instrumentation__.Notify(36315)
		w.ChunkSize = 0
	} else {
		__antithesis_instrumentation__.Notify(36316)
	}
	__antithesis_instrumentation__.Notify(36314)
	return w, nil
}

func (g *gcsStorage) ReadFile(ctx context.Context, basename string) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(36317)
	reader, _, err := g.ReadFileAt(ctx, basename, 0)
	return reader, err
}

func (g *gcsStorage) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36318)
	object := path.Join(g.prefix, basename)

	ctx, sp := tracing.ChildSpan(ctx, "gcs.ReadFileAt")
	defer sp.Finish()
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("gcs.ReadFileAt: %s",
		path.Join(g.prefix, basename))})

	r := cloud.NewResumingReader(ctx,
		func(ctx context.Context, pos int64) (io.ReadCloser, error) {
			__antithesis_instrumentation__.Notify(36321)
			return g.bucket.Object(object).NewRangeReader(ctx, pos, -1)
		},
		nil,
		offset,
		cloud.IsResumableHTTPError,
		nil,
	)
	__antithesis_instrumentation__.Notify(36319)

	if err := r.Open(ctx); err != nil {
		__antithesis_instrumentation__.Notify(36322)
		if errors.Is(err, gcs.ErrObjectNotExist) {
			__antithesis_instrumentation__.Notify(36324)

			err = errors.Wrapf(
				errors.Wrap(cloud.ErrFileDoesNotExist, "gcs object does not exist"),
				"%v",
				err.Error(),
			)
		} else {
			__antithesis_instrumentation__.Notify(36325)
		}
		__antithesis_instrumentation__.Notify(36323)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(36326)
	}
	__antithesis_instrumentation__.Notify(36320)
	return r, r.Reader.(*gcs.Reader).Attrs.Size, nil
}

func (g *gcsStorage) List(ctx context.Context, prefix, delim string, fn cloud.ListingFn) error {
	__antithesis_instrumentation__.Notify(36327)
	dest := cloud.JoinPathPreservingTrailingSlash(g.prefix, prefix)
	ctx, sp := tracing.ChildSpan(ctx, "gcs.List")
	defer sp.Finish()
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("gcs.List: %s", dest)})

	it := g.bucket.Objects(ctx, &gcs.Query{Prefix: dest, Delimiter: delim})
	for {
		__antithesis_instrumentation__.Notify(36328)
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			__antithesis_instrumentation__.Notify(36332)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(36333)
		}
		__antithesis_instrumentation__.Notify(36329)
		if err != nil {
			__antithesis_instrumentation__.Notify(36334)
			return errors.Wrap(err, "unable to list files in gcs bucket")
		} else {
			__antithesis_instrumentation__.Notify(36335)
		}
		__antithesis_instrumentation__.Notify(36330)
		name := attrs.Name
		if name == "" {
			__antithesis_instrumentation__.Notify(36336)
			name = attrs.Prefix
		} else {
			__antithesis_instrumentation__.Notify(36337)
		}
		__antithesis_instrumentation__.Notify(36331)
		if err := fn(strings.TrimPrefix(name, dest)); err != nil {
			__antithesis_instrumentation__.Notify(36338)
			return err
		} else {
			__antithesis_instrumentation__.Notify(36339)
		}
	}
}

func (g *gcsStorage) Delete(ctx context.Context, basename string) error {
	__antithesis_instrumentation__.Notify(36340)
	return contextutil.RunWithTimeout(ctx, "delete gcs file",
		cloud.Timeout.Get(&g.settings.SV),
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(36341)
			return g.bucket.Object(path.Join(g.prefix, basename)).Delete(ctx)
		})
}

func (g *gcsStorage) Size(ctx context.Context, basename string) (int64, error) {
	__antithesis_instrumentation__.Notify(36342)
	var r *gcs.Reader
	if err := contextutil.RunWithTimeout(ctx, "size gcs file",
		cloud.Timeout.Get(&g.settings.SV),
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(36344)
			var err error
			r, err = g.bucket.Object(path.Join(g.prefix, basename)).NewReader(ctx)
			return err
		}); err != nil {
		__antithesis_instrumentation__.Notify(36345)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(36346)
	}
	__antithesis_instrumentation__.Notify(36343)
	sz := r.Attrs.Size
	_ = r.Close()
	return sz, nil
}

func (g *gcsStorage) Close() error {
	__antithesis_instrumentation__.Notify(36347)
	return g.client.Close()
}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_gs,
		parseGSURL, makeGCSStorage, cloud.RedactedParams(CredentialsParam), "gs")
}
