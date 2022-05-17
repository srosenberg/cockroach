package amazon

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
)

const (
	AWSAccessKeyParam = "AWS_ACCESS_KEY_ID"

	AWSSecretParam = "AWS_SECRET_ACCESS_KEY"

	AWSTempTokenParam = "AWS_SESSION_TOKEN"

	AWSEndpointParam = "AWS_ENDPOINT"

	AWSServerSideEncryptionMode = "AWS_SERVER_ENC_MODE"

	AWSServerSideEncryptionKMSID = "AWS_SERVER_KMS_ID"

	S3StorageClassParam = "S3_STORAGE_CLASS"

	S3RegionParam = "AWS_REGION"

	KMSRegionParam = "REGION"
)

type s3Storage struct {
	bucket   *string
	conf     *roachpb.ExternalStorage_S3
	ioConf   base.ExternalIODirConfig
	settings *cluster.Settings
	prefix   string

	opts   s3ClientConfig
	cached *s3Client
}

type s3Client struct {
	client   *s3.S3
	uploader *s3manager.Uploader
}

var reuseSession = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"cloudstorage.s3.session_reuse.enabled",
	"persist the last opened s3 session and re-use it when opening a new session with the same arguments",
	true,
)

type s3ClientConfig struct {
	endpoint, region, bucket, accessKey, secret, tempToken, auth string

	verbose bool
}

func clientConfig(conf *roachpb.ExternalStorage_S3) s3ClientConfig {
	__antithesis_instrumentation__.Notify(35737)
	return s3ClientConfig{
		endpoint:  conf.Endpoint,
		region:    conf.Region,
		bucket:    conf.Bucket,
		accessKey: conf.AccessKey,
		secret:    conf.Secret,
		tempToken: conf.TempToken,
		auth:      conf.Auth,
		verbose:   log.V(2),
	}
}

var s3ClientCache struct {
	syncutil.Mutex

	key    s3ClientConfig
	client *s3Client
}

var _ cloud.ExternalStorage = &s3Storage{}

type serverSideEncMode string

const (
	kmsEnc    serverSideEncMode = "aws:kms"
	aes256Enc serverSideEncMode = "AES256"
)

func S3URI(bucket, path string, conf *roachpb.ExternalStorage_S3) string {
	__antithesis_instrumentation__.Notify(35738)
	q := make(url.Values)
	setIf := func(key, value string) {
		__antithesis_instrumentation__.Notify(35740)
		if value != "" {
			__antithesis_instrumentation__.Notify(35741)
			q.Set(key, value)
		} else {
			__antithesis_instrumentation__.Notify(35742)
		}
	}
	__antithesis_instrumentation__.Notify(35739)
	setIf(AWSAccessKeyParam, conf.AccessKey)
	setIf(AWSSecretParam, conf.Secret)
	setIf(AWSTempTokenParam, conf.TempToken)
	setIf(AWSEndpointParam, conf.Endpoint)
	setIf(S3RegionParam, conf.Region)
	setIf(cloud.AuthParam, conf.Auth)
	setIf(AWSServerSideEncryptionMode, conf.ServerEncMode)
	setIf(AWSServerSideEncryptionKMSID, conf.ServerKMSID)
	setIf(S3StorageClassParam, conf.StorageClass)

	s3URL := url.URL{
		Scheme:   "s3",
		Host:     bucket,
		Path:     path,
		RawQuery: q.Encode(),
	}

	return s3URL.String()
}

func parseS3URL(_ cloud.ExternalStorageURIContext, uri *url.URL) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(35743)
	conf := roachpb.ExternalStorage{}
	conf.Provider = roachpb.ExternalStorageProvider_s3
	conf.S3Config = &roachpb.ExternalStorage_S3{
		Bucket:        uri.Host,
		Prefix:        uri.Path,
		AccessKey:     uri.Query().Get(AWSAccessKeyParam),
		Secret:        uri.Query().Get(AWSSecretParam),
		TempToken:     uri.Query().Get(AWSTempTokenParam),
		Endpoint:      uri.Query().Get(AWSEndpointParam),
		Region:        uri.Query().Get(S3RegionParam),
		Auth:          uri.Query().Get(cloud.AuthParam),
		ServerEncMode: uri.Query().Get(AWSServerSideEncryptionMode),
		ServerKMSID:   uri.Query().Get(AWSServerSideEncryptionKMSID),
		StorageClass:  uri.Query().Get(S3StorageClassParam),
	}
	conf.S3Config.Prefix = strings.TrimLeft(conf.S3Config.Prefix, "/")

	conf.S3Config.Secret = strings.Replace(conf.S3Config.Secret, " ", "+", -1)
	return conf, nil
}

func MakeS3Storage(
	ctx context.Context, args cloud.ExternalStorageContext, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(35744)
	telemetry.Count("external-io.s3")
	conf := dest.S3Config
	if conf == nil {
		__antithesis_instrumentation__.Notify(35752)
		return nil, errors.Errorf("s3 upload requested but info missing")
	} else {
		__antithesis_instrumentation__.Notify(35753)
	}
	__antithesis_instrumentation__.Notify(35745)

	if conf.Endpoint != "" {
		__antithesis_instrumentation__.Notify(35754)
		if args.IOConf.DisableHTTP {
			__antithesis_instrumentation__.Notify(35755)
			return nil, errors.New(
				"custom endpoints disallowed for s3 due to --external-io-disable-http flag")
		} else {
			__antithesis_instrumentation__.Notify(35756)
		}
	} else {
		__antithesis_instrumentation__.Notify(35757)
	}
	__antithesis_instrumentation__.Notify(35746)

	switch conf.Auth {
	case "", cloud.AuthParamSpecified:
		__antithesis_instrumentation__.Notify(35758)
		if conf.AccessKey == "" {
			__antithesis_instrumentation__.Notify(35762)
			return nil, errors.Errorf(
				"%s is set to '%s', but %s is not set",
				cloud.AuthParam,
				cloud.AuthParamSpecified,
				AWSAccessKeyParam,
			)
		} else {
			__antithesis_instrumentation__.Notify(35763)
		}
		__antithesis_instrumentation__.Notify(35759)
		if conf.Secret == "" {
			__antithesis_instrumentation__.Notify(35764)
			return nil, errors.Errorf(
				"%s is set to '%s', but %s is not set",
				cloud.AuthParam,
				cloud.AuthParamSpecified,
				AWSSecretParam,
			)
		} else {
			__antithesis_instrumentation__.Notify(35765)
		}
	case cloud.AuthParamImplicit:
		__antithesis_instrumentation__.Notify(35760)
		if args.IOConf.DisableImplicitCredentials {
			__antithesis_instrumentation__.Notify(35766)
			return nil, errors.New(
				"implicit credentials disallowed for s3 due to --external-io-implicit-credentials flag")
		} else {
			__antithesis_instrumentation__.Notify(35767)
		}
	default:
		__antithesis_instrumentation__.Notify(35761)
		return nil, errors.Errorf("unsupported value %s for %s", conf.Auth, cloud.AuthParam)
	}
	__antithesis_instrumentation__.Notify(35747)

	if conf.ServerEncMode != "" {
		__antithesis_instrumentation__.Notify(35768)
		switch conf.ServerEncMode {
		case string(aes256Enc):
			__antithesis_instrumentation__.Notify(35769)
		case string(kmsEnc):
			__antithesis_instrumentation__.Notify(35770)
			if conf.ServerKMSID == "" {
				__antithesis_instrumentation__.Notify(35772)
				return nil, errors.New("AWS_SERVER_KMS_ID param must be set" +
					" when using aws:kms server side encryption mode.")
			} else {
				__antithesis_instrumentation__.Notify(35773)
			}
		default:
			__antithesis_instrumentation__.Notify(35771)
			return nil, errors.Newf("unsupported server encryption mode %s. "+
				"Supported values are `aws:kms` and `AES256`.", conf.ServerEncMode)
		}
	} else {
		__antithesis_instrumentation__.Notify(35774)
	}
	__antithesis_instrumentation__.Notify(35748)

	s := &s3Storage{
		bucket:   aws.String(conf.Bucket),
		conf:     conf,
		ioConf:   args.IOConf,
		prefix:   conf.Prefix,
		settings: args.Settings,
		opts:     clientConfig(conf),
	}

	reuse := reuseSession.Get(&args.Settings.SV)
	if !reuse {
		__antithesis_instrumentation__.Notify(35775)
		return s, nil
	} else {
		__antithesis_instrumentation__.Notify(35776)
	}
	__antithesis_instrumentation__.Notify(35749)

	s3ClientCache.Lock()
	defer s3ClientCache.Unlock()

	if s3ClientCache.key == s.opts {
		__antithesis_instrumentation__.Notify(35777)
		s.cached = s3ClientCache.client
		return s, nil
	} else {
		__antithesis_instrumentation__.Notify(35778)
	}
	__antithesis_instrumentation__.Notify(35750)

	client, _, err := newClient(ctx, s.opts, s.settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(35779)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35780)
	}
	__antithesis_instrumentation__.Notify(35751)
	s.cached = &client
	s3ClientCache.key = s.opts
	s3ClientCache.client = &client
	return s, nil
}

func newClient(
	ctx context.Context, conf s3ClientConfig, settings *cluster.Settings,
) (s3Client, string, error) {
	__antithesis_instrumentation__.Notify(35781)

	if conf.region == "" || func() bool {
		__antithesis_instrumentation__.Notify(35788)
		return conf.auth == cloud.AuthParamImplicit == true
	}() == true {
		__antithesis_instrumentation__.Notify(35789)
		var sp *tracing.Span
		ctx, sp = tracing.ChildSpan(ctx, "open s3 client")
		defer sp.Finish()
	} else {
		__antithesis_instrumentation__.Notify(35790)
	}
	__antithesis_instrumentation__.Notify(35782)

	opts := session.Options{}

	if conf.endpoint != "" {
		__antithesis_instrumentation__.Notify(35791)
		opts.Config.Endpoint = aws.String(conf.endpoint)
		opts.Config.S3ForcePathStyle = aws.Bool(true)

		if conf.region == "" {
			__antithesis_instrumentation__.Notify(35794)
			conf.region = "default-region"
		} else {
			__antithesis_instrumentation__.Notify(35795)
		}
		__antithesis_instrumentation__.Notify(35792)

		client, err := cloud.MakeHTTPClient(settings)
		if err != nil {
			__antithesis_instrumentation__.Notify(35796)
			return s3Client{}, "", err
		} else {
			__antithesis_instrumentation__.Notify(35797)
		}
		__antithesis_instrumentation__.Notify(35793)
		opts.Config.HTTPClient = client
	} else {
		__antithesis_instrumentation__.Notify(35798)
	}
	__antithesis_instrumentation__.Notify(35783)

	switch conf.auth {
	case "", cloud.AuthParamSpecified:
		__antithesis_instrumentation__.Notify(35799)
		opts.Config.WithCredentials(credentials.NewStaticCredentials(conf.accessKey, conf.secret, conf.tempToken))
	case cloud.AuthParamImplicit:
		__antithesis_instrumentation__.Notify(35800)
		opts.SharedConfigState = session.SharedConfigEnable
	default:
		__antithesis_instrumentation__.Notify(35801)
	}
	__antithesis_instrumentation__.Notify(35784)

	opts.Config.MaxRetries = aws.Int(10)

	opts.Config.CredentialsChainVerboseErrors = aws.Bool(true)

	if conf.verbose {
		__antithesis_instrumentation__.Notify(35802)
		opts.Config.LogLevel = aws.LogLevel(aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors)
	} else {
		__antithesis_instrumentation__.Notify(35803)
	}
	__antithesis_instrumentation__.Notify(35785)

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		__antithesis_instrumentation__.Notify(35804)
		return s3Client{}, "", errors.Wrap(err, "new aws session")
	} else {
		__antithesis_instrumentation__.Notify(35805)
	}
	__antithesis_instrumentation__.Notify(35786)

	region := conf.region
	if region == "" {
		__antithesis_instrumentation__.Notify(35806)
		if err := cloud.DelayedRetry(ctx, "s3manager.GetBucketRegion", s3ErrDelay, func() error {
			__antithesis_instrumentation__.Notify(35807)
			region, err = s3manager.GetBucketRegion(ctx, sess, conf.bucket, "us-east-1")
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(35808)
			return s3Client{}, "", errors.Wrap(err, "could not find s3 bucket's region")
		} else {
			__antithesis_instrumentation__.Notify(35809)
		}
	} else {
		__antithesis_instrumentation__.Notify(35810)
	}
	__antithesis_instrumentation__.Notify(35787)
	sess.Config.Region = aws.String(region)

	c := s3.New(sess)
	u := s3manager.NewUploader(sess)
	return s3Client{client: c, uploader: u}, region, nil
}

func (s *s3Storage) getClient(ctx context.Context) (*s3.S3, error) {
	__antithesis_instrumentation__.Notify(35811)
	if s.cached != nil {
		__antithesis_instrumentation__.Notify(35815)
		return s.cached.client, nil
	} else {
		__antithesis_instrumentation__.Notify(35816)
	}
	__antithesis_instrumentation__.Notify(35812)
	client, region, err := newClient(ctx, s.opts, s.settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(35817)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35818)
	}
	__antithesis_instrumentation__.Notify(35813)
	if s.opts.region == "" {
		__antithesis_instrumentation__.Notify(35819)
		s.opts.region = region
	} else {
		__antithesis_instrumentation__.Notify(35820)
	}
	__antithesis_instrumentation__.Notify(35814)
	return client.client, nil
}

func (s *s3Storage) getUploader(ctx context.Context) (*s3manager.Uploader, error) {
	__antithesis_instrumentation__.Notify(35821)
	if s.cached != nil {
		__antithesis_instrumentation__.Notify(35825)
		return s.cached.uploader, nil
	} else {
		__antithesis_instrumentation__.Notify(35826)
	}
	__antithesis_instrumentation__.Notify(35822)
	client, region, err := newClient(ctx, s.opts, s.settings)
	if err != nil {
		__antithesis_instrumentation__.Notify(35827)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35828)
	}
	__antithesis_instrumentation__.Notify(35823)
	if s.opts.region == "" {
		__antithesis_instrumentation__.Notify(35829)
		s.opts.region = region
	} else {
		__antithesis_instrumentation__.Notify(35830)
	}
	__antithesis_instrumentation__.Notify(35824)
	return client.uploader, nil
}

func (s *s3Storage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(35831)
	return roachpb.ExternalStorage{
		Provider: roachpb.ExternalStorageProvider_s3,
		S3Config: s.conf,
	}
}

func (s *s3Storage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(35832)
	return s.ioConf
}

func (s *s3Storage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(35833)
	return s.settings
}

func (s *s3Storage) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(35834)
	uploader, err := s.getUploader(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(35836)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35837)
	}
	__antithesis_instrumentation__.Notify(35835)

	ctx, sp := tracing.ChildSpan(ctx, "s3.Writer")
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("s3.Writer: %s", path.Join(s.prefix, basename))})
	return cloud.BackgroundPipe(ctx, func(ctx context.Context, r io.Reader) error {
		__antithesis_instrumentation__.Notify(35838)
		defer sp.Finish()

		_, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket:               s.bucket,
			Key:                  aws.String(path.Join(s.prefix, basename)),
			Body:                 r,
			ServerSideEncryption: nilIfEmpty(s.conf.ServerEncMode),
			SSEKMSKeyId:          nilIfEmpty(s.conf.ServerKMSID),
			StorageClass:         nilIfEmpty(s.conf.StorageClass),
		})
		return errors.Wrap(err, "upload failed")
	}), nil
}

func (s *s3Storage) openStreamAt(
	ctx context.Context, basename string, pos int64,
) (*s3.GetObjectOutput, error) {
	__antithesis_instrumentation__.Notify(35839)
	client, err := s.getClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(35843)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35844)
	}
	__antithesis_instrumentation__.Notify(35840)
	req := &s3.GetObjectInput{Bucket: s.bucket, Key: aws.String(path.Join(s.prefix, basename))}
	if pos != 0 {
		__antithesis_instrumentation__.Notify(35845)
		req.Range = aws.String(fmt.Sprintf("bytes=%d-", pos))
	} else {
		__antithesis_instrumentation__.Notify(35846)
	}
	__antithesis_instrumentation__.Notify(35841)

	out, err := client.GetObjectWithContext(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(35847)
		if aerr := (awserr.Error)(nil); errors.As(err, &aerr) {
			__antithesis_instrumentation__.Notify(35849)
			switch aerr.Code() {

			case s3.ErrCodeNoSuchBucket, s3.ErrCodeNoSuchKey:
				__antithesis_instrumentation__.Notify(35850)

				return nil, errors.Wrapf(
					errors.Wrap(cloud.ErrFileDoesNotExist, "s3 object does not exist"),
					"%v",
					err.Error(),
				)
			default:
				__antithesis_instrumentation__.Notify(35851)
			}
		} else {
			__antithesis_instrumentation__.Notify(35852)
		}
		__antithesis_instrumentation__.Notify(35848)
		return nil, errors.Wrap(err, "failed to get s3 object")
	} else {
		__antithesis_instrumentation__.Notify(35853)
	}
	__antithesis_instrumentation__.Notify(35842)
	return out, nil
}

func (s *s3Storage) ReadFile(ctx context.Context, basename string) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(35854)
	reader, _, err := s.ReadFileAt(ctx, basename, 0)
	return reader, err
}

func (s *s3Storage) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(35855)
	ctx, sp := tracing.ChildSpan(ctx, "s3.ReadFileAt")
	defer sp.Finish()
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("s3.ReadFileAt: %s", path.Join(s.prefix, basename))})

	stream, err := s.openStreamAt(ctx, basename, offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(35859)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(35860)
	}
	__antithesis_instrumentation__.Notify(35856)
	var size int64
	if offset != 0 {
		__antithesis_instrumentation__.Notify(35861)
		if stream.ContentRange == nil {
			__antithesis_instrumentation__.Notify(35863)
			return nil, 0, errors.New("expected content range for read at offset")
		} else {
			__antithesis_instrumentation__.Notify(35864)
		}
		__antithesis_instrumentation__.Notify(35862)
		size, err = cloud.CheckHTTPContentRangeHeader(*stream.ContentRange, offset)
		if err != nil {
			__antithesis_instrumentation__.Notify(35865)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(35866)
		}
	} else {
		__antithesis_instrumentation__.Notify(35867)
		if stream.ContentLength == nil {
			__antithesis_instrumentation__.Notify(35868)
			log.Warningf(ctx, "Content length missing from S3 GetObject (is this actually s3?); attempting to lookup size with separate call...")

			x, err := s.Size(ctx, basename)
			if err != nil {
				__antithesis_instrumentation__.Notify(35870)
				return nil, 0, errors.Wrap(err, "content-length missing from GetObject and Size() failed")
			} else {
				__antithesis_instrumentation__.Notify(35871)
			}
			__antithesis_instrumentation__.Notify(35869)
			size = x
		} else {
			__antithesis_instrumentation__.Notify(35872)
			size = *stream.ContentLength
		}
	}
	__antithesis_instrumentation__.Notify(35857)
	opener := func(ctx context.Context, pos int64) (io.ReadCloser, error) {
		__antithesis_instrumentation__.Notify(35873)
		s, err := s.openStreamAt(ctx, basename, pos)
		if err != nil {
			__antithesis_instrumentation__.Notify(35875)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(35876)
		}
		__antithesis_instrumentation__.Notify(35874)
		return s.Body, nil
	}
	__antithesis_instrumentation__.Notify(35858)
	return cloud.NewResumingReader(ctx, opener, stream.Body, offset,
		cloud.IsResumableHTTPError, s3ErrDelay), size, nil
}

func (s *s3Storage) List(ctx context.Context, prefix, delim string, fn cloud.ListingFn) error {
	__antithesis_instrumentation__.Notify(35877)
	ctx, sp := tracing.ChildSpan(ctx, "s3.List")
	defer sp.Finish()

	dest := cloud.JoinPathPreservingTrailingSlash(s.prefix, prefix)
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("s3.List: %s", dest)})

	client, err := s.getClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(35881)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35882)
	}
	__antithesis_instrumentation__.Notify(35878)

	var fnErr error
	pageFn := func(page *s3.ListObjectsOutput, lastPage bool) bool {
		__antithesis_instrumentation__.Notify(35883)
		for _, x := range page.CommonPrefixes {
			__antithesis_instrumentation__.Notify(35886)
			if fnErr = fn(strings.TrimPrefix(*x.Prefix, dest)); fnErr != nil {
				__antithesis_instrumentation__.Notify(35887)
				return false
			} else {
				__antithesis_instrumentation__.Notify(35888)
			}
		}
		__antithesis_instrumentation__.Notify(35884)
		for _, fileObject := range page.Contents {
			__antithesis_instrumentation__.Notify(35889)
			if fnErr = fn(strings.TrimPrefix(*fileObject.Key, dest)); fnErr != nil {
				__antithesis_instrumentation__.Notify(35890)
				return false
			} else {
				__antithesis_instrumentation__.Notify(35891)
			}
		}
		__antithesis_instrumentation__.Notify(35885)

		return true
	}
	__antithesis_instrumentation__.Notify(35879)

	if err := client.ListObjectsPagesWithContext(
		ctx, &s3.ListObjectsInput{Bucket: s.bucket, Prefix: aws.String(dest), Delimiter: nilIfEmpty(delim)}, pageFn,
	); err != nil {
		__antithesis_instrumentation__.Notify(35892)
		return errors.Wrap(err, `failed to list s3 bucket`)
	} else {
		__antithesis_instrumentation__.Notify(35893)
	}
	__antithesis_instrumentation__.Notify(35880)

	return fnErr
}

func (s *s3Storage) Delete(ctx context.Context, basename string) error {
	__antithesis_instrumentation__.Notify(35894)
	client, err := s.getClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(35896)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35897)
	}
	__antithesis_instrumentation__.Notify(35895)
	return contextutil.RunWithTimeout(ctx, "delete s3 object",
		cloud.Timeout.Get(&s.settings.SV),
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35898)
			_, err := client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
				Bucket: s.bucket,
				Key:    aws.String(path.Join(s.prefix, basename)),
			})
			return err
		})
}

func (s *s3Storage) Size(ctx context.Context, basename string) (int64, error) {
	__antithesis_instrumentation__.Notify(35899)
	client, err := s.getClient(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(35903)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(35904)
	}
	__antithesis_instrumentation__.Notify(35900)
	var out *s3.HeadObjectOutput
	err = contextutil.RunWithTimeout(ctx, "get s3 object header",
		cloud.Timeout.Get(&s.settings.SV),
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35905)
			var err error
			out, err = client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
				Bucket: s.bucket,
				Key:    aws.String(path.Join(s.prefix, basename)),
			})
			return err
		})
	__antithesis_instrumentation__.Notify(35901)
	if err != nil {
		__antithesis_instrumentation__.Notify(35906)
		return 0, errors.Wrap(err, "failed to get s3 object headers")
	} else {
		__antithesis_instrumentation__.Notify(35907)
	}
	__antithesis_instrumentation__.Notify(35902)
	return *out.ContentLength, nil
}

func (s *s3Storage) Close() error {
	__antithesis_instrumentation__.Notify(35908)
	return nil
}

func nilIfEmpty(s string) *string {
	__antithesis_instrumentation__.Notify(35909)
	if s == "" {
		__antithesis_instrumentation__.Notify(35911)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(35912)
	}
	__antithesis_instrumentation__.Notify(35910)
	return aws.String(s)
}

func s3ErrDelay(err error) time.Duration {
	__antithesis_instrumentation__.Notify(35913)
	var s3err s3.RequestFailure
	if errors.As(err, &s3err) {
		__antithesis_instrumentation__.Notify(35915)

		if s3err.StatusCode() == 503 {
			__antithesis_instrumentation__.Notify(35916)
			return time.Second * 5
		} else {
			__antithesis_instrumentation__.Notify(35917)
		}
	} else {
		__antithesis_instrumentation__.Notify(35918)
	}
	__antithesis_instrumentation__.Notify(35914)
	return 0
}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_s3,
		parseS3URL, MakeS3Storage, cloud.RedactedParams(AWSSecretParam, AWSTempTokenParam), "s3")
}
