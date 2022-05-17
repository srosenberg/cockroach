package nodelocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/blobs"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func parseNodelocalURL(
	_ cloud.ExternalStorageURIContext, uri *url.URL,
) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36556)
	conf := roachpb.ExternalStorage{}
	if uri.Host == "" {
		__antithesis_instrumentation__.Notify(36559)
		return conf, errors.Errorf(
			"host component of nodelocal URI must be a node ID ("+
				"use 'self' to specify each node should access its own local filesystem): %s",
			uri.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(36560)
		if uri.Host == "self" {
			__antithesis_instrumentation__.Notify(36561)
			uri.Host = "0"
		} else {
			__antithesis_instrumentation__.Notify(36562)
		}
	}
	__antithesis_instrumentation__.Notify(36557)

	nodeID, err := strconv.Atoi(uri.Host)
	if err != nil {
		__antithesis_instrumentation__.Notify(36563)
		return conf, errors.Errorf("host component of nodelocal URI must be a node ID: %s", uri.String())
	} else {
		__antithesis_instrumentation__.Notify(36564)
	}
	__antithesis_instrumentation__.Notify(36558)
	conf.Provider = roachpb.ExternalStorageProvider_nodelocal
	conf.LocalFile.Path = uri.Path
	conf.LocalFile.NodeID = roachpb.NodeID(nodeID)
	return conf, nil
}

type localFileStorage struct {
	cfg        roachpb.ExternalStorage_LocalFilePath
	ioConf     base.ExternalIODirConfig
	base       string
	blobClient blobs.BlobClient
	settings   *cluster.Settings
}

var _ cloud.ExternalStorage = &localFileStorage{}

func MakeLocalStorageURI(path string) string {
	__antithesis_instrumentation__.Notify(36565)
	return fmt.Sprintf("nodelocal://0/%s", path)
}

func TestingMakeLocalStorage(
	ctx context.Context,
	cfg roachpb.ExternalStorage_LocalFilePath,
	settings *cluster.Settings,
	blobClientFactory blobs.BlobClientFactory,
	ioConf base.ExternalIODirConfig,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36566)
	args := cloud.ExternalStorageContext{IOConf: ioConf, BlobClientFactory: blobClientFactory, Settings: settings}
	return makeLocalStorage(ctx, args, roachpb.ExternalStorage{LocalFile: cfg})
}

func makeLocalStorage(
	ctx context.Context, args cloud.ExternalStorageContext, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36567)
	telemetry.Count("external-io.nodelocal")
	if args.BlobClientFactory == nil {
		__antithesis_instrumentation__.Notify(36571)
		return nil, errors.New("nodelocal storage is not available")
	} else {
		__antithesis_instrumentation__.Notify(36572)
	}
	__antithesis_instrumentation__.Notify(36568)
	cfg := dest.LocalFile
	if cfg.Path == "" {
		__antithesis_instrumentation__.Notify(36573)
		return nil, errors.Errorf("local storage requested but path not provided")
	} else {
		__antithesis_instrumentation__.Notify(36574)
	}
	__antithesis_instrumentation__.Notify(36569)
	client, err := args.BlobClientFactory(ctx, cfg.NodeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(36575)
		return nil, errors.Wrap(err, "failed to create blob client")
	} else {
		__antithesis_instrumentation__.Notify(36576)
	}
	__antithesis_instrumentation__.Notify(36570)
	return &localFileStorage{base: cfg.Path, cfg: cfg, ioConf: args.IOConf, blobClient: client,
		settings: args.Settings}, nil
}

func (l *localFileStorage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(36577)
	return roachpb.ExternalStorage{
		Provider:  roachpb.ExternalStorageProvider_nodelocal,
		LocalFile: l.cfg,
	}
}

func (l *localFileStorage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(36578)
	return l.ioConf
}

func (l *localFileStorage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(36579)
	return l.settings
}

func joinRelativePath(filePath string, file string) string {
	__antithesis_instrumentation__.Notify(36580)

	return path.Join(".", filePath, file)
}

func (l *localFileStorage) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(36581)
	return l.blobClient.Writer(ctx, joinRelativePath(l.base, basename))
}

func (l *localFileStorage) ReadFile(
	ctx context.Context, basename string,
) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(36582)
	body, _, err := l.ReadFileAt(ctx, basename, 0)
	return body, err
}

func (l *localFileStorage) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36583)
	reader, size, err := l.blobClient.ReadFile(ctx, joinRelativePath(l.base, basename), offset)
	if err != nil {
		__antithesis_instrumentation__.Notify(36585)

		if oserror.IsNotExist(err) || func() bool {
			__antithesis_instrumentation__.Notify(36587)
			return status.Code(err) == codes.NotFound == true
		}() == true {
			__antithesis_instrumentation__.Notify(36588)

			return nil, 0, errors.WithMessagef(
				errors.Wrap(cloud.ErrFileDoesNotExist, "nodelocal storage file does not exist"),
				"%s",
				err.Error(),
			)
		} else {
			__antithesis_instrumentation__.Notify(36589)
		}
		__antithesis_instrumentation__.Notify(36586)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(36590)
	}
	__antithesis_instrumentation__.Notify(36584)
	return reader, size, nil
}

func (l *localFileStorage) List(
	ctx context.Context, prefix, delim string, fn cloud.ListingFn,
) error {
	__antithesis_instrumentation__.Notify(36591)
	dest := cloud.JoinPathPreservingTrailingSlash(l.base, prefix)

	res, err := l.blobClient.List(ctx, dest)
	if err != nil {
		__antithesis_instrumentation__.Notify(36594)
		return errors.Wrap(err, "unable to match pattern provided")
	} else {
		__antithesis_instrumentation__.Notify(36595)
	}
	__antithesis_instrumentation__.Notify(36592)

	sort.Strings(res)
	var prevPrefix string
	for _, f := range res {
		__antithesis_instrumentation__.Notify(36596)
		f = strings.TrimPrefix(f, dest)
		if delim != "" {
			__antithesis_instrumentation__.Notify(36598)
			if i := strings.Index(f, delim); i >= 0 {
				__antithesis_instrumentation__.Notify(36601)
				f = f[:i+len(delim)]
			} else {
				__antithesis_instrumentation__.Notify(36602)
			}
			__antithesis_instrumentation__.Notify(36599)
			if f == prevPrefix {
				__antithesis_instrumentation__.Notify(36603)
				continue
			} else {
				__antithesis_instrumentation__.Notify(36604)
			}
			__antithesis_instrumentation__.Notify(36600)
			prevPrefix = f
		} else {
			__antithesis_instrumentation__.Notify(36605)
		}
		__antithesis_instrumentation__.Notify(36597)
		if err := fn(f); err != nil {
			__antithesis_instrumentation__.Notify(36606)
			return err
		} else {
			__antithesis_instrumentation__.Notify(36607)
		}
	}
	__antithesis_instrumentation__.Notify(36593)
	return nil
}

func (l *localFileStorage) Delete(ctx context.Context, basename string) error {
	__antithesis_instrumentation__.Notify(36608)
	return l.blobClient.Delete(ctx, joinRelativePath(l.base, basename))
}

func (l *localFileStorage) Size(ctx context.Context, basename string) (int64, error) {
	__antithesis_instrumentation__.Notify(36609)
	stat, err := l.blobClient.Stat(ctx, joinRelativePath(l.base, basename))
	if err != nil {
		__antithesis_instrumentation__.Notify(36611)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(36612)
	}
	__antithesis_instrumentation__.Notify(36610)
	return stat.Filesize, nil
}

func (*localFileStorage) Close() error {
	__antithesis_instrumentation__.Notify(36613)
	return nil
}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_nodelocal,
		parseNodelocalURL, makeLocalStorage, cloud.RedactedParams(), "nodelocal")
}
