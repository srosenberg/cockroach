package azure

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
)

const (
	AzureAccountNameParam = "AZURE_ACCOUNT_NAME"

	AzureAccountKeyParam = "AZURE_ACCOUNT_KEY"
)

func parseAzureURL(
	_ cloud.ExternalStorageURIContext, uri *url.URL,
) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(35919)
	conf := roachpb.ExternalStorage{}
	conf.Provider = roachpb.ExternalStorageProvider_azure
	conf.AzureConfig = &roachpb.ExternalStorage_Azure{
		Container:   uri.Host,
		Prefix:      uri.Path,
		AccountName: uri.Query().Get(AzureAccountNameParam),
		AccountKey:  uri.Query().Get(AzureAccountKeyParam),
	}
	if conf.AzureConfig.AccountName == "" {
		__antithesis_instrumentation__.Notify(35922)
		return conf, errors.Errorf("azure uri missing %q parameter", AzureAccountNameParam)
	} else {
		__antithesis_instrumentation__.Notify(35923)
	}
	__antithesis_instrumentation__.Notify(35920)
	if conf.AzureConfig.AccountKey == "" {
		__antithesis_instrumentation__.Notify(35924)
		return conf, errors.Errorf("azure uri missing %q parameter", AzureAccountKeyParam)
	} else {
		__antithesis_instrumentation__.Notify(35925)
	}
	__antithesis_instrumentation__.Notify(35921)
	conf.AzureConfig.Prefix = strings.TrimLeft(conf.AzureConfig.Prefix, "/")
	return conf, nil
}

type azureStorage struct {
	conf      *roachpb.ExternalStorage_Azure
	ioConf    base.ExternalIODirConfig
	container azblob.ContainerURL
	prefix    string
	settings  *cluster.Settings
}

var _ cloud.ExternalStorage = &azureStorage{}

func makeAzureStorage(
	_ context.Context, args cloud.ExternalStorageContext, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(35926)
	telemetry.Count("external-io.azure")
	conf := dest.AzureConfig
	if conf == nil {
		__antithesis_instrumentation__.Notify(35930)
		return nil, errors.Errorf("azure upload requested but info missing")
	} else {
		__antithesis_instrumentation__.Notify(35931)
	}
	__antithesis_instrumentation__.Notify(35927)
	credential, err := azblob.NewSharedKeyCredential(conf.AccountName, conf.AccountKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(35932)
		return nil, errors.Wrap(err, "azure credential")
	} else {
		__antithesis_instrumentation__.Notify(35933)
	}
	__antithesis_instrumentation__.Notify(35928)
	p := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", conf.AccountName))
	if err != nil {
		__antithesis_instrumentation__.Notify(35934)
		return nil, errors.Wrap(err, "azure: account name is not valid")
	} else {
		__antithesis_instrumentation__.Notify(35935)
	}
	__antithesis_instrumentation__.Notify(35929)
	serviceURL := azblob.NewServiceURL(*u, p)
	return &azureStorage{
		conf:      conf,
		ioConf:    args.IOConf,
		container: serviceURL.NewContainerURL(conf.Container),
		prefix:    conf.Prefix,
		settings:  args.Settings,
	}, nil
}

func (s *azureStorage) getBlob(basename string) azblob.BlockBlobURL {
	__antithesis_instrumentation__.Notify(35936)
	name := path.Join(s.prefix, basename)
	return s.container.NewBlockBlobURL(name)
}

func (s *azureStorage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(35937)
	return roachpb.ExternalStorage{
		Provider:    roachpb.ExternalStorageProvider_azure,
		AzureConfig: s.conf,
	}
}

func (s *azureStorage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(35938)
	return s.ioConf
}

func (s *azureStorage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(35939)
	return s.settings
}

func (s *azureStorage) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(35940)
	ctx, sp := tracing.ChildSpan(ctx, "azure.Writer")
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("azure.Writer: %s",
		path.Join(s.prefix, basename))})
	blob := s.getBlob(basename)
	return cloud.BackgroundPipe(ctx, func(ctx context.Context, r io.Reader) error {
		__antithesis_instrumentation__.Notify(35941)
		defer sp.Finish()
		_, err := azblob.UploadStreamToBlockBlob(
			ctx, r, blob, azblob.UploadStreamToBlockBlobOptions{
				BufferSize: 4 << 20,
			},
		)
		return err
	}), nil
}

func (s *azureStorage) ReadFile(ctx context.Context, basename string) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(35942)
	reader, _, err := s.ReadFileAt(ctx, basename, 0)
	return reader, err
}

func (s *azureStorage) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(35943)
	ctx, sp := tracing.ChildSpan(ctx, "azure.ReadFileAt")
	defer sp.Finish()
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("azure.ReadFileAt: %s",
		path.Join(s.prefix, basename))})

	blob := s.getBlob(basename)
	get, err := blob.Download(ctx, offset, azblob.CountToEnd, azblob.BlobAccessConditions{},
		false, azblob.ClientProvidedKeyOptions{},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(35946)
		if azerr := (azblob.StorageError)(nil); errors.As(err, &azerr) {
			__antithesis_instrumentation__.Notify(35948)
			switch azerr.ServiceCode() {

			case azblob.ServiceCodeBlobNotFound, azblob.ServiceCodeResourceNotFound:
				__antithesis_instrumentation__.Notify(35949)

				return nil, 0, errors.Wrapf(
					errors.Wrap(cloud.ErrFileDoesNotExist, "azure blob does not exist"),
					"%v",
					err.Error(),
				)
			default:
				__antithesis_instrumentation__.Notify(35950)
			}
		} else {
			__antithesis_instrumentation__.Notify(35951)
		}
		__antithesis_instrumentation__.Notify(35947)
		return nil, 0, errors.Wrap(err, "failed to create azure reader")
	} else {
		__antithesis_instrumentation__.Notify(35952)
	}
	__antithesis_instrumentation__.Notify(35944)
	var size int64
	if offset == 0 {
		__antithesis_instrumentation__.Notify(35953)
		size = get.ContentLength()
	} else {
		__antithesis_instrumentation__.Notify(35954)
		size, err = cloud.CheckHTTPContentRangeHeader(get.ContentRange(), offset)
		if err != nil {
			__antithesis_instrumentation__.Notify(35955)
			return nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(35956)
		}
	}
	__antithesis_instrumentation__.Notify(35945)
	reader := get.Body(azblob.RetryReaderOptions{MaxRetryRequests: 3})

	return ioctx.ReadCloserAdapter(reader), size, nil
}

func (s *azureStorage) List(ctx context.Context, prefix, delim string, fn cloud.ListingFn) error {
	__antithesis_instrumentation__.Notify(35957)
	ctx, sp := tracing.ChildSpan(ctx, "azure.List")
	defer sp.Finish()

	dest := cloud.JoinPathPreservingTrailingSlash(s.prefix, prefix)
	sp.RecordStructured(&types.StringValue{Value: fmt.Sprintf("azure.List: %s", dest)})

	var marker azblob.Marker
	for marker.NotDone() {
		__antithesis_instrumentation__.Notify(35959)
		response, err := s.container.ListBlobsHierarchySegment(
			ctx, marker, delim, azblob.ListBlobsSegmentOptions{Prefix: dest},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(35963)
			return errors.Wrap(err, "unable to list files for specified blob")
		} else {
			__antithesis_instrumentation__.Notify(35964)
		}
		__antithesis_instrumentation__.Notify(35960)
		for _, blob := range response.Segment.BlobPrefixes {
			__antithesis_instrumentation__.Notify(35965)
			if err := fn(strings.TrimPrefix(blob.Name, dest)); err != nil {
				__antithesis_instrumentation__.Notify(35966)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35967)
			}
		}
		__antithesis_instrumentation__.Notify(35961)
		for _, blob := range response.Segment.BlobItems {
			__antithesis_instrumentation__.Notify(35968)
			if err := fn(strings.TrimPrefix(blob.Name, dest)); err != nil {
				__antithesis_instrumentation__.Notify(35969)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35970)
			}
		}
		__antithesis_instrumentation__.Notify(35962)
		marker = response.NextMarker
	}
	__antithesis_instrumentation__.Notify(35958)
	return nil
}

func (s *azureStorage) Delete(ctx context.Context, basename string) error {
	__antithesis_instrumentation__.Notify(35971)
	err := contextutil.RunWithTimeout(ctx, "delete azure file", cloud.Timeout.Get(&s.settings.SV),
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35973)
			blob := s.getBlob(basename)
			_, err := blob.Delete(ctx, azblob.DeleteSnapshotsOptionNone, azblob.BlobAccessConditions{})
			return err
		})
	__antithesis_instrumentation__.Notify(35972)
	return errors.Wrap(err, "delete file")
}

func (s *azureStorage) Size(ctx context.Context, basename string) (int64, error) {
	__antithesis_instrumentation__.Notify(35974)
	var props *azblob.BlobGetPropertiesResponse
	err := contextutil.RunWithTimeout(ctx, "size azure file", cloud.Timeout.Get(&s.settings.SV),
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35977)
			blob := s.getBlob(basename)
			var err error
			props, err = blob.GetProperties(ctx, azblob.BlobAccessConditions{}, azblob.ClientProvidedKeyOptions{})
			return err
		})
	__antithesis_instrumentation__.Notify(35975)
	if err != nil {
		__antithesis_instrumentation__.Notify(35978)
		return 0, errors.Wrap(err, "get file properties")
	} else {
		__antithesis_instrumentation__.Notify(35979)
	}
	__antithesis_instrumentation__.Notify(35976)
	return props.ContentLength(), nil
}

func (s *azureStorage) Close() error {
	__antithesis_instrumentation__.Notify(35980)
	return nil
}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_azure,
		parseAzureURL, makeAzureStorage, cloud.RedactedParams(AzureAccountKeyParam), "azure")
}
