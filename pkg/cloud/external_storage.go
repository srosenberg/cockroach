package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"io"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
)

type ExternalStorage interface {
	io.Closer

	Conf() roachpb.ExternalStorage

	ExternalIOConf() base.ExternalIODirConfig

	Settings() *cluster.Settings

	ReadFile(ctx context.Context, basename string) (ioctx.ReadCloserCtx, error)

	ReadFileAt(ctx context.Context, basename string, offset int64) (ioctx.ReadCloserCtx, int64, error)

	Writer(ctx context.Context, basename string) (io.WriteCloser, error)

	List(ctx context.Context, prefix, delimiter string, fn ListingFn) error

	Delete(ctx context.Context, basename string) error

	Size(ctx context.Context, basename string) (int64, error)
}

type ListingFn func(string) error

type ExternalStorageFactory func(ctx context.Context, dest roachpb.ExternalStorage) (ExternalStorage, error)

type ExternalStorageFromURIFactory func(ctx context.Context, uri string,
	user security.SQLUsername) (ExternalStorage, error)

type SQLConnI interface {
	driver.QueryerContext
	driver.ExecerContext
}

var ErrFileDoesNotExist = errors.New("external_storage: file doesn't exist")

var ErrListingUnsupported = errors.New("listing is not supported")

var ErrListingDone = errors.New("listing is done")

func RedactedParams(strs ...string) map[string]struct{} {
	__antithesis_instrumentation__.Notify(36231)
	if len(strs) == 0 {
		__antithesis_instrumentation__.Notify(36234)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(36235)
	}
	__antithesis_instrumentation__.Notify(36232)
	m := make(map[string]struct{}, len(strs))
	for i := range strs {
		__antithesis_instrumentation__.Notify(36236)
		m[strs[i]] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(36233)
	return m
}

type ExternalStorageURIContext struct {
	CurrentUser security.SQLUsername
}

type ExternalStorageURIParser func(ExternalStorageURIContext, *url.URL) (roachpb.ExternalStorage, error)

type ExternalStorageContext struct {
	IOConf            base.ExternalIODirConfig
	Settings          *cluster.Settings
	BlobClientFactory blobs.BlobClientFactory
	InternalExecutor  sqlutil.InternalExecutor
	DB                *kv.DB
}

type ExternalStorageConstructor func(
	context.Context, ExternalStorageContext, roachpb.ExternalStorage,
) (ExternalStorage, error)
