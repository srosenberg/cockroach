package userfile

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/cloud/userfile/filetable"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
)

const (
	DefaultQualifiedNamespace = "defaultdb.public."

	DefaultQualifiedNamePrefix = "userfiles_"
)

func parseUserfileURL(
	args cloud.ExternalStorageURIContext, uri *url.URL,
) (roachpb.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36644)
	conf := roachpb.ExternalStorage{}
	qualifiedTableName := uri.Host
	if args.CurrentUser.Undefined() {
		__antithesis_instrumentation__.Notify(36647)
		return conf, errors.Errorf("user creating the FileTable ExternalStorage must be specified")
	} else {
		__antithesis_instrumentation__.Notify(36648)
	}
	__antithesis_instrumentation__.Notify(36645)
	normUser := args.CurrentUser.Normalized()

	if qualifiedTableName == "" {
		__antithesis_instrumentation__.Notify(36649)
		composedTableName := security.MakeSQLUsernameFromPreNormalizedString(
			DefaultQualifiedNamePrefix + normUser)
		qualifiedTableName = DefaultQualifiedNamespace +

			composedTableName.SQLIdentifier()
	} else {
		__antithesis_instrumentation__.Notify(36650)
	}
	__antithesis_instrumentation__.Notify(36646)

	conf.Provider = roachpb.ExternalStorageProvider_userfile
	conf.FileTableConfig.User = normUser
	conf.FileTableConfig.QualifiedTableName = qualifiedTableName
	conf.FileTableConfig.Path = uri.Path
	return conf, nil
}

type fileTableStorage struct {
	fs       *filetable.FileToTableSystem
	cfg      roachpb.ExternalStorage_FileTable
	ioConf   base.ExternalIODirConfig
	db       *kv.DB
	ie       sqlutil.InternalExecutor
	prefix   string
	settings *cluster.Settings
}

var _ cloud.ExternalStorage = &fileTableStorage{}

func makeFileTableStorage(
	ctx context.Context, args cloud.ExternalStorageContext, dest roachpb.ExternalStorage,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36651)
	telemetry.Count("external-io.filetable")

	cfg := dest.FileTableConfig
	if cfg.User == "" || func() bool {
		__antithesis_instrumentation__.Notify(36655)
		return cfg.QualifiedTableName == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(36656)
		return nil, errors.Errorf("FileTable storage requested but username or qualified table name" +
			" not provided")
	} else {
		__antithesis_instrumentation__.Notify(36657)
	}
	__antithesis_instrumentation__.Notify(36652)

	if path.Clean(cfg.Path) != cfg.Path {
		__antithesis_instrumentation__.Notify(36658)

		trimmedPath := strings.TrimSuffix(cfg.Path, ".tmp")
		return nil, errors.Newf("path %s changes after normalization to %s. "+
			"userfile upload does not permit such path constructs",
			trimmedPath, path.Clean(trimmedPath))
	} else {
		__antithesis_instrumentation__.Notify(36659)
	}
	__antithesis_instrumentation__.Notify(36653)

	username := security.MakeSQLUsernameFromPreNormalizedString(cfg.User)
	executor := filetable.MakeInternalFileToTableExecutor(args.InternalExecutor, args.DB)

	fileToTableSystem, err := filetable.NewFileToTableSystem(ctx,
		cfg.QualifiedTableName, executor, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(36660)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36661)
	}
	__antithesis_instrumentation__.Notify(36654)
	return &fileTableStorage{
		fs:       fileToTableSystem,
		cfg:      cfg,
		ioConf:   args.IOConf,
		db:       args.DB,
		ie:       args.InternalExecutor,
		prefix:   cfg.Path,
		settings: args.Settings,
	}, nil
}

func MakeSQLConnFileTableStorage(
	ctx context.Context, cfg roachpb.ExternalStorage_FileTable, conn cloud.SQLConnI,
) (cloud.ExternalStorage, error) {
	__antithesis_instrumentation__.Notify(36662)
	executor := filetable.MakeSQLConnFileToTableExecutor(conn)

	username := security.MakeSQLUsernameFromPreNormalizedString(cfg.User)

	fileToTableSystem, err := filetable.NewFileToTableSystem(ctx,
		cfg.QualifiedTableName, executor, username)
	if err != nil {
		__antithesis_instrumentation__.Notify(36665)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36666)
	}
	__antithesis_instrumentation__.Notify(36663)
	prefix := cfg.Path
	if !strings.HasPrefix(prefix, "/") {
		__antithesis_instrumentation__.Notify(36667)
		prefix = "/" + prefix
	} else {
		__antithesis_instrumentation__.Notify(36668)
	}
	__antithesis_instrumentation__.Notify(36664)
	return &fileTableStorage{
		fs:       fileToTableSystem,
		cfg:      cfg,
		ioConf:   base.ExternalIODirConfig{},
		prefix:   prefix,
		settings: nil,
	}, nil
}

func MakeUserFileStorageURI(qualifiedTableName, filename string) string {
	__antithesis_instrumentation__.Notify(36669)
	return fmt.Sprintf("userfile://%s/%s", qualifiedTableName, filename)
}

func (f *fileTableStorage) Close() error {
	__antithesis_instrumentation__.Notify(36670)
	return nil
}

func (f *fileTableStorage) Conf() roachpb.ExternalStorage {
	__antithesis_instrumentation__.Notify(36671)
	return roachpb.ExternalStorage{
		Provider:        roachpb.ExternalStorageProvider_userfile,
		FileTableConfig: f.cfg,
	}
}

func (f *fileTableStorage) ExternalIOConf() base.ExternalIODirConfig {
	__antithesis_instrumentation__.Notify(36672)
	return f.ioConf
}

func (f *fileTableStorage) Settings() *cluster.Settings {
	__antithesis_instrumentation__.Notify(36673)
	return f.settings
}

func checkBaseAndJoinFilePath(prefix, basename string) (string, error) {
	__antithesis_instrumentation__.Notify(36674)
	if basename == "" {
		__antithesis_instrumentation__.Notify(36677)
		return prefix, nil
	} else {
		__antithesis_instrumentation__.Notify(36678)
	}
	__antithesis_instrumentation__.Notify(36675)

	if path.Clean(basename) != basename {
		__antithesis_instrumentation__.Notify(36679)
		return "", errors.Newf("basename %s changes to %s on normalization. "+
			"userfile does not permit such constructs.", basename, path.Clean(basename))
	} else {
		__antithesis_instrumentation__.Notify(36680)
	}
	__antithesis_instrumentation__.Notify(36676)
	return path.Join(prefix, basename), nil
}

func (f *fileTableStorage) ReadFile(
	ctx context.Context, basename string,
) (ioctx.ReadCloserCtx, error) {
	__antithesis_instrumentation__.Notify(36681)
	body, _, err := f.ReadFileAt(ctx, basename, 0)
	return body, err
}

func (f *fileTableStorage) ReadFileAt(
	ctx context.Context, basename string, offset int64,
) (ioctx.ReadCloserCtx, int64, error) {
	__antithesis_instrumentation__.Notify(36682)
	filepath, err := checkBaseAndJoinFilePath(f.prefix, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36685)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(36686)
	}
	__antithesis_instrumentation__.Notify(36683)
	reader, size, err := f.fs.ReadFile(ctx, filepath, offset)
	if oserror.IsNotExist(err) {
		__antithesis_instrumentation__.Notify(36687)
		return nil, 0, errors.Wrapf(cloud.ErrFileDoesNotExist,
			"file %s does not exist in the UserFileTableSystem", filepath)
	} else {
		__antithesis_instrumentation__.Notify(36688)
	}
	__antithesis_instrumentation__.Notify(36684)

	return reader, size, err
}

func (f *fileTableStorage) Writer(ctx context.Context, basename string) (io.WriteCloser, error) {
	__antithesis_instrumentation__.Notify(36689)
	filepath, err := checkBaseAndJoinFilePath(f.prefix, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36692)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(36693)
	}
	__antithesis_instrumentation__.Notify(36690)

	if f.ie == nil {
		__antithesis_instrumentation__.Notify(36694)
		return nil, errors.New("cannot Write without a configured internal executor")
	} else {
		__antithesis_instrumentation__.Notify(36695)
	}
	__antithesis_instrumentation__.Notify(36691)

	return f.fs.NewFileWriter(ctx, filepath, filetable.ChunkDefaultSize)
}

func (f *fileTableStorage) List(
	ctx context.Context, prefix, delim string, fn cloud.ListingFn,
) error {
	__antithesis_instrumentation__.Notify(36696)
	dest := cloud.JoinPathPreservingTrailingSlash(f.prefix, prefix)

	res, err := f.fs.ListFiles(ctx, dest)
	if err != nil {
		__antithesis_instrumentation__.Notify(36699)
		return errors.Wrap(err, "fail to list destination")
	} else {
		__antithesis_instrumentation__.Notify(36700)
	}
	__antithesis_instrumentation__.Notify(36697)

	sort.Strings(res)
	var prevPrefix string
	for _, f := range res {
		__antithesis_instrumentation__.Notify(36701)
		f = strings.TrimPrefix(f, dest)
		if delim != "" {
			__antithesis_instrumentation__.Notify(36703)
			if i := strings.Index(f, delim); i >= 0 {
				__antithesis_instrumentation__.Notify(36706)
				f = f[:i+len(delim)]
			} else {
				__antithesis_instrumentation__.Notify(36707)
			}
			__antithesis_instrumentation__.Notify(36704)
			if f == prevPrefix {
				__antithesis_instrumentation__.Notify(36708)
				continue
			} else {
				__antithesis_instrumentation__.Notify(36709)
			}
			__antithesis_instrumentation__.Notify(36705)
			prevPrefix = f
		} else {
			__antithesis_instrumentation__.Notify(36710)
		}
		__antithesis_instrumentation__.Notify(36702)
		if err := fn(f); err != nil {
			__antithesis_instrumentation__.Notify(36711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(36712)
		}
	}
	__antithesis_instrumentation__.Notify(36698)

	return nil
}

func (f *fileTableStorage) Delete(ctx context.Context, basename string) error {
	__antithesis_instrumentation__.Notify(36713)
	filepath, err := checkBaseAndJoinFilePath(f.prefix, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36715)
		return err
	} else {
		__antithesis_instrumentation__.Notify(36716)
	}
	__antithesis_instrumentation__.Notify(36714)
	return f.fs.DeleteFile(ctx, filepath)
}

func (f *fileTableStorage) Size(ctx context.Context, basename string) (int64, error) {
	__antithesis_instrumentation__.Notify(36717)
	filepath, err := checkBaseAndJoinFilePath(f.prefix, basename)
	if err != nil {
		__antithesis_instrumentation__.Notify(36719)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(36720)
	}
	__antithesis_instrumentation__.Notify(36718)
	return f.fs.FileSize(ctx, filepath)
}

func init() {
	cloud.RegisterExternalStorageProvider(roachpb.ExternalStorageProvider_userfile,
		parseUserfileURL, makeFileTableStorage, cloud.RedactedParams(), "userfile")
}
