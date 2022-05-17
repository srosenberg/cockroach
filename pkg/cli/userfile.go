package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/cloud/userfile"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

const (
	defaultUserfileScheme         = "userfile"
	defaultQualifiedNamePrefix    = "defaultdb.public.userfiles_"
	defaultQualifiedHexNamePrefix = "defaultdb.public.userfilesx_"
	tmpSuffix                     = ".tmp"
	fileTableNameSuffix           = "_upload_files"
)

var userFileUploadCmd = &cobra.Command{
	Use:   "upload <source> <destination>",
	Short: "upload file from source to destination",
	Long: `
Uploads a single file, or, with the -r flag, all the files in the subtree rooted
at a directory, to the user-scoped file storage using a SQL connection.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: clierrorplus.MaybeShoutError(runUserFileUpload),
}

var userFileListCmd = &cobra.Command{
	Use:   "list <file|dir glob>",
	Short: "list files matching the provided pattern",
	Long: `
Lists the files stored in the user scoped file storage which match the provided pattern,
using a SQL connection. If no pattern is provided, all files in the specified
(or default, if unspecified) user scoped file storage will be listed.
`,
	Args:    cobra.MinimumNArgs(0),
	RunE:    clierrorplus.MaybeShoutError(runUserFileList),
	Aliases: []string{"ls"},
}

var userFileGetCmd = &cobra.Command{
	Use:   "get <file|dir glob> [destination]",
	Short: "get file(s) matching the provided pattern",
	Long: `
Fetch the files stored in the user scoped file storage which match the provided pattern,
using a SQL connection, to the current directory or 'destination' if provided.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: clierrorplus.MaybeShoutError(runUserFileGet),
}

var userFileDeleteCmd = &cobra.Command{
	Use:   "delete <file|dir glob>",
	Short: "delete files matching the provided pattern",
	Long: `
Deletes the files stored in the user scoped file storage which match the provided pattern,
using a SQL connection. If passed pattern '*', all files in the specified
(or default, if unspecified) user scoped file storage will be deleted. Deletions are not
atomic, and all deletions prior to the first failure will occur.
`,
	Args:    cobra.MinimumNArgs(1),
	RunE:    clierrorplus.MaybeShoutError(runUserFileDelete),
	Aliases: []string{"rm"},
}

func runUserFileDelete(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34894)
	conn, err := makeSQLClient("cockroach userfile", useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34899)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34900)
	}
	__antithesis_instrumentation__.Notify(34895)
	defer func() {
		__antithesis_instrumentation__.Notify(34901)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34896)

	glob := args[0]

	var deletedFiles []string
	if deletedFiles, err = deleteUserFile(context.Background(), conn, glob); err != nil {
		__antithesis_instrumentation__.Notify(34902)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34903)
	}
	__antithesis_instrumentation__.Notify(34897)

	telemetry.Count("userfile.command.delete")
	for _, file := range deletedFiles {
		__antithesis_instrumentation__.Notify(34904)
		fmt.Printf("successfully deleted %s\n", file)
	}
	__antithesis_instrumentation__.Notify(34898)

	return nil
}

func runUserFileList(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34905)
	conn, err := makeSQLClient("cockroach userfile", useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34911)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34912)
	}
	__antithesis_instrumentation__.Notify(34906)
	defer func() {
		__antithesis_instrumentation__.Notify(34913)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34907)

	var glob string
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(34914)
		glob = args[0]
	} else {
		__antithesis_instrumentation__.Notify(34915)
	}
	__antithesis_instrumentation__.Notify(34908)

	var files []string
	if files, err = listUserFile(context.Background(), conn, glob); err != nil {
		__antithesis_instrumentation__.Notify(34916)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34917)
	}
	__antithesis_instrumentation__.Notify(34909)

	telemetry.Count("userfile.command.list")
	for _, file := range files {
		__antithesis_instrumentation__.Notify(34918)
		fmt.Println(file)
	}
	__antithesis_instrumentation__.Notify(34910)

	return nil
}

func uploadUserFileRecursive(conn clisqlclient.Conn, srcDir, dstDir string) error {
	__antithesis_instrumentation__.Notify(34919)
	srcHasTrailingSlash := strings.HasSuffix(srcDir, "/")
	var err error
	srcDir, err = filepath.Abs(srcDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(34924)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34925)
	}
	__antithesis_instrumentation__.Notify(34920)
	dstDir = strings.TrimSuffix(dstDir, "/")

	srcDirBase := filepath.Base(srcDir)
	if dstDir == "" {
		__antithesis_instrumentation__.Notify(34926)
		dstDir = srcDirBase
	} else {
		__antithesis_instrumentation__.Notify(34927)
		if !srcHasTrailingSlash {
			__antithesis_instrumentation__.Notify(34928)
			dstDir = dstDir + "/" + srcDirBase
		} else {
			__antithesis_instrumentation__.Notify(34929)
		}
	}
	__antithesis_instrumentation__.Notify(34921)

	ctx := context.Background()

	err = filepath.WalkDir(srcDir,
		func(path string, info fs.DirEntry, err error) error {
			__antithesis_instrumentation__.Notify(34930)
			if err != nil {
				__antithesis_instrumentation__.Notify(34933)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34934)
			}
			__antithesis_instrumentation__.Notify(34931)
			if !info.IsDir() {
				__antithesis_instrumentation__.Notify(34935)
				relativePath := strings.TrimPrefix(path, srcDir+"/")
				fmt.Printf("uploading: %s\n", relativePath)

				uploadedFile, err := uploadUserFile(ctx, conn, path, dstDir+"/"+relativePath)
				if err != nil {
					__antithesis_instrumentation__.Notify(34937)
					return err
				} else {
					__antithesis_instrumentation__.Notify(34938)
				}
				__antithesis_instrumentation__.Notify(34936)
				fmt.Printf("successfully uploaded to %s\n", uploadedFile)
			} else {
				__antithesis_instrumentation__.Notify(34939)
			}
			__antithesis_instrumentation__.Notify(34932)
			return nil
		})
	__antithesis_instrumentation__.Notify(34922)
	if err != nil {
		__antithesis_instrumentation__.Notify(34940)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34941)
	}
	__antithesis_instrumentation__.Notify(34923)

	fmt.Printf("successfully uploaded all files in the subtree rooted at %s\n", filepath.Base(srcDir))
	return nil
}

func runUserFileUpload(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34942)
	conn, err := makeSQLClient("cockroach userfile", useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34947)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34948)
	}
	__antithesis_instrumentation__.Notify(34943)
	defer func() {
		__antithesis_instrumentation__.Notify(34949)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34944)

	source := args[0]

	var destination string
	if len(args) == 2 {
		__antithesis_instrumentation__.Notify(34950)
		destination = args[1]
	} else {
		__antithesis_instrumentation__.Notify(34951)
	}
	__antithesis_instrumentation__.Notify(34945)

	if userfileCtx.recursive {
		__antithesis_instrumentation__.Notify(34952)
		if err := uploadUserFileRecursive(conn, source, destination); err != nil {
			__antithesis_instrumentation__.Notify(34953)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34954)
		}
	} else {
		__antithesis_instrumentation__.Notify(34955)
		uploadedFile, err := uploadUserFile(context.Background(), conn, source,
			destination)
		if err != nil {
			__antithesis_instrumentation__.Notify(34957)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34958)
		}
		__antithesis_instrumentation__.Notify(34956)
		fmt.Printf("successfully uploaded to %s\n", uploadedFile)
	}
	__antithesis_instrumentation__.Notify(34946)

	telemetry.Count("userfile.command.upload")
	return nil
}

func runUserFileGet(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34959)
	conn, err := makeSQLClient("cockroach userfile", useDefaultDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34968)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34969)
	}
	__antithesis_instrumentation__.Notify(34960)
	defer func() {
		__antithesis_instrumentation__.Notify(34970)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34961)
	ctx := context.Background()

	var dest string
	if len(args) > 1 {
		__antithesis_instrumentation__.Notify(34971)
		dest = args[len(args)-1]
	} else {
		__antithesis_instrumentation__.Notify(34972)
	}
	__antithesis_instrumentation__.Notify(34962)

	conf, err := getUserfileConf(ctx, conn, args[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(34973)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34974)
	}
	__antithesis_instrumentation__.Notify(34963)

	fullPath := conf.Path
	conf.Path = cloud.GetPrefixBeforeWildcard(fullPath)
	pattern := fullPath[len(conf.Path):]
	displayPath := strings.TrimPrefix(conf.Path, "/")

	f, err := userfile.MakeSQLConnFileTableStorage(ctx, conf, conn.GetDriverConn().(cloud.SQLConnI))
	if err != nil {
		__antithesis_instrumentation__.Notify(34975)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34976)
	}
	__antithesis_instrumentation__.Notify(34964)
	defer f.Close()

	var files []string
	if err := f.List(ctx, "", "", func(s string) error {
		__antithesis_instrumentation__.Notify(34977)
		if pattern != "" {
			__antithesis_instrumentation__.Notify(34979)
			if ok, err := path.Match(pattern, s); err != nil || func() bool {
				__antithesis_instrumentation__.Notify(34980)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(34981)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34982)
			}
		} else {
			__antithesis_instrumentation__.Notify(34983)
		}
		__antithesis_instrumentation__.Notify(34978)
		files = append(files, s)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(34984)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34985)
	}
	__antithesis_instrumentation__.Notify(34965)

	if len(files) == 0 {
		__antithesis_instrumentation__.Notify(34986)
		return errors.New("no files matched requested path or path pattern")
	} else {
		__antithesis_instrumentation__.Notify(34987)
	}
	__antithesis_instrumentation__.Notify(34966)

	for _, src := range files {
		__antithesis_instrumentation__.Notify(34988)
		file := displayPath + src
		var fileDest string
		if len(files) > 1 {
			__antithesis_instrumentation__.Notify(34991)

			fileDest = filepath.Join(dest, filepath.FromSlash(file))
		} else {
			__antithesis_instrumentation__.Notify(34992)
			filename := path.Base(file)

			if dest == "" {
				__antithesis_instrumentation__.Notify(34993)
				fileDest = filename
			} else {
				__antithesis_instrumentation__.Notify(34994)

				stat, err := os.Stat(dest)
				if err != nil && func() bool {
					__antithesis_instrumentation__.Notify(34996)
					return !errors.Is(err, os.ErrNotExist) == true
				}() == true {
					__antithesis_instrumentation__.Notify(34997)
					return err
				} else {
					__antithesis_instrumentation__.Notify(34998)
				}
				__antithesis_instrumentation__.Notify(34995)
				if err == nil && func() bool {
					__antithesis_instrumentation__.Notify(34999)
					return stat.IsDir() == true
				}() == true {
					__antithesis_instrumentation__.Notify(35000)
					fileDest = filepath.Join(dest, filename)
				} else {
					__antithesis_instrumentation__.Notify(35001)
					fileDest = dest
				}
			}
		}
		__antithesis_instrumentation__.Notify(34989)
		fmt.Printf("downloading %s... ", file)
		sz, err := downloadUserfile(ctx, f, src, fileDest)
		if err != nil {
			__antithesis_instrumentation__.Notify(35002)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35003)
		}
		__antithesis_instrumentation__.Notify(34990)
		fmt.Printf("\rdownloaded %s to %s (%s)\n", file, fileDest, humanizeutil.IBytes(sz))
	}
	__antithesis_instrumentation__.Notify(34967)

	return nil

}

func openUserFile(source string) (io.ReadCloser, error) {
	__antithesis_instrumentation__.Notify(35004)
	f, err := os.Open(source)
	if err != nil {
		__antithesis_instrumentation__.Notify(35008)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35009)
	}
	__antithesis_instrumentation__.Notify(35005)
	stat, err := f.Stat()
	if err != nil {
		__antithesis_instrumentation__.Notify(35010)
		return nil, errors.Wrapf(err, "unable to get source file stats for %s", source)
	} else {
		__antithesis_instrumentation__.Notify(35011)
	}
	__antithesis_instrumentation__.Notify(35006)
	if stat.IsDir() {
		__antithesis_instrumentation__.Notify(35012)
		return nil, fmt.Errorf("source file %s is a directory, not a file", source)
	} else {
		__antithesis_instrumentation__.Notify(35013)
	}
	__antithesis_instrumentation__.Notify(35007)
	return f, nil
}

func getDefaultQualifiedTableName(user security.SQLUsername) string {
	__antithesis_instrumentation__.Notify(35014)
	normalizedUsername := user.Normalized()
	if lexbase.IsBareIdentifier(normalizedUsername) {
		__antithesis_instrumentation__.Notify(35016)
		return defaultQualifiedNamePrefix + normalizedUsername
	} else {
		__antithesis_instrumentation__.Notify(35017)
	}
	__antithesis_instrumentation__.Notify(35015)
	return defaultQualifiedHexNamePrefix + fmt.Sprintf("%x", normalizedUsername)
}

func constructUserfileDestinationURI(source, destination string, user security.SQLUsername) string {
	__antithesis_instrumentation__.Notify(35018)

	if destination == "" {
		__antithesis_instrumentation__.Notify(35021)
		sourceFilename := path.Base(source)
		userFileURL := url.URL{
			Scheme: defaultUserfileScheme,
			Host:   getDefaultQualifiedTableName(user),
			Path:   sourceFilename,
		}
		return userFileURL.String()
	} else {
		__antithesis_instrumentation__.Notify(35022)
	}
	__antithesis_instrumentation__.Notify(35019)

	var userfileURI *url.URL
	var err error
	if userfileURI, err = url.ParseRequestURI(destination); err == nil {
		__antithesis_instrumentation__.Notify(35023)
		if userfileURI.Scheme == defaultUserfileScheme {
			__antithesis_instrumentation__.Notify(35024)
			if userfileURI.Host == "" {
				__antithesis_instrumentation__.Notify(35026)
				userfileURI.Host = getDefaultQualifiedTableName(user)
			} else {
				__antithesis_instrumentation__.Notify(35027)
			}
			__antithesis_instrumentation__.Notify(35025)
			return userfileURI.String()
		} else {
			__antithesis_instrumentation__.Notify(35028)
		}
	} else {
		__antithesis_instrumentation__.Notify(35029)
	}
	__antithesis_instrumentation__.Notify(35020)

	userFileURL := url.URL{
		Scheme: defaultUserfileScheme,
		Host:   getDefaultQualifiedTableName(user),
		Path:   destination,
	}
	return userFileURL.String()
}

func constructUserfileListURI(glob string, user security.SQLUsername) string {
	__antithesis_instrumentation__.Notify(35030)

	if userfileURL, err := url.ParseRequestURI(glob); err == nil {
		__antithesis_instrumentation__.Notify(35032)
		if userfileURL.Scheme == defaultUserfileScheme {
			__antithesis_instrumentation__.Notify(35033)
			return userfileURL.String()
		} else {
			__antithesis_instrumentation__.Notify(35034)
		}
	} else {
		__antithesis_instrumentation__.Notify(35035)
	}
	__antithesis_instrumentation__.Notify(35031)

	userfileURL := url.URL{
		Scheme: defaultUserfileScheme,
		Host:   getDefaultQualifiedTableName(user),
		Path:   glob,
	}

	return userfileURL.String()
}

func getUserfileConf(
	ctx context.Context, conn clisqlclient.Conn, glob string,
) (roachpb.ExternalStorage_FileTable, error) {
	__antithesis_instrumentation__.Notify(35036)
	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(35041)
		return roachpb.ExternalStorage_FileTable{}, err
	} else {
		__antithesis_instrumentation__.Notify(35042)
	}
	__antithesis_instrumentation__.Notify(35037)

	connURL, err := url.Parse(conn.GetURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(35043)
		return roachpb.ExternalStorage_FileTable{}, err
	} else {
		__antithesis_instrumentation__.Notify(35044)
	}
	__antithesis_instrumentation__.Notify(35038)

	reqUsername, _ := security.MakeSQLUsernameFromUserInput(connURL.User.Username(), security.UsernameValidation)

	userfileListURI := constructUserfileListURI(glob, reqUsername)
	unescapedUserfileListURI, err := url.PathUnescape(userfileListURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(35045)
		return roachpb.ExternalStorage_FileTable{}, err
	} else {
		__antithesis_instrumentation__.Notify(35046)
	}
	__antithesis_instrumentation__.Notify(35039)

	userFileTableConf, err := cloud.ExternalStorageConfFromURI(unescapedUserfileListURI, reqUsername)
	if err != nil {
		__antithesis_instrumentation__.Notify(35047)
		return roachpb.ExternalStorage_FileTable{}, err
	} else {
		__antithesis_instrumentation__.Notify(35048)
	}
	__antithesis_instrumentation__.Notify(35040)
	return userFileTableConf.FileTableConfig, nil

}

func listUserFile(ctx context.Context, conn clisqlclient.Conn, glob string) ([]string, error) {
	__antithesis_instrumentation__.Notify(35049)
	conf, err := getUserfileConf(ctx, conn, glob)
	if err != nil {
		__antithesis_instrumentation__.Notify(35053)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35054)
	}
	__antithesis_instrumentation__.Notify(35050)

	fullPath := conf.Path
	conf.Path = cloud.GetPrefixBeforeWildcard(fullPath)
	pattern := fullPath[len(conf.Path):]

	f, err := userfile.MakeSQLConnFileTableStorage(ctx, conf, conn.GetDriverConn().(cloud.SQLConnI))
	if err != nil {
		__antithesis_instrumentation__.Notify(35055)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35056)
	}
	__antithesis_instrumentation__.Notify(35051)
	defer f.Close()

	displayPrefix := strings.TrimPrefix(conf.Path, "/")
	var res []string
	if err := f.List(ctx, "", "", func(s string) error {
		__antithesis_instrumentation__.Notify(35057)
		if pattern != "" {
			__antithesis_instrumentation__.Notify(35059)
			ok, err := path.Match(pattern, s)
			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(35060)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(35061)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35062)
			}
		} else {
			__antithesis_instrumentation__.Notify(35063)
		}
		__antithesis_instrumentation__.Notify(35058)
		res = append(res, displayPrefix+s)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(35064)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35065)
	}
	__antithesis_instrumentation__.Notify(35052)
	return res, nil
}

func downloadUserfile(
	ctx context.Context, store cloud.ExternalStorage, src, dst string,
) (int64, error) {
	__antithesis_instrumentation__.Notify(35066)
	remoteFile, err := store.ReadFile(ctx, src)
	if err != nil {
		__antithesis_instrumentation__.Notify(35070)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(35071)
	}
	__antithesis_instrumentation__.Notify(35067)
	defer remoteFile.Close(ctx)

	localDir := path.Dir(dst)
	if err := os.MkdirAll(localDir, 0700); err != nil {
		__antithesis_instrumentation__.Notify(35072)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(35073)
	}
	__antithesis_instrumentation__.Notify(35068)

	localFile, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		__antithesis_instrumentation__.Notify(35074)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(35075)
	}
	__antithesis_instrumentation__.Notify(35069)
	defer localFile.Close()

	return io.Copy(localFile, ioctx.ReaderCtxAdapter(ctx, remoteFile))
}

func deleteUserFile(ctx context.Context, conn clisqlclient.Conn, glob string) ([]string, error) {
	__antithesis_instrumentation__.Notify(35076)
	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(35083)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35084)
	}
	__antithesis_instrumentation__.Notify(35077)

	connURL, err := url.Parse(conn.GetURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(35085)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35086)
	}
	__antithesis_instrumentation__.Notify(35078)

	reqUsername, _ := security.MakeSQLUsernameFromUserInput(connURL.User.Username(), security.UsernameValidation)

	userfileListURI := constructUserfileListURI(glob, reqUsername)
	unescapedUserfileListURI, err := url.PathUnescape(userfileListURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(35087)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35088)
	}
	__antithesis_instrumentation__.Notify(35079)

	userFileTableConf, err := cloud.ExternalStorageConfFromURI(unescapedUserfileListURI, reqUsername)
	if err != nil {
		__antithesis_instrumentation__.Notify(35089)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35090)
	}
	__antithesis_instrumentation__.Notify(35080)

	fullPath := userFileTableConf.FileTableConfig.Path
	userFileTableConf.FileTableConfig.Path = cloud.GetPrefixBeforeWildcard(fullPath)
	pattern := fullPath[len(userFileTableConf.FileTableConfig.Path):]

	f, err := userfile.MakeSQLConnFileTableStorage(ctx, userFileTableConf.FileTableConfig,
		conn.GetDriverConn().(cloud.SQLConnI))
	if err != nil {
		__antithesis_instrumentation__.Notify(35091)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35092)
	}
	__antithesis_instrumentation__.Notify(35081)

	displayRoot := strings.TrimPrefix(userFileTableConf.FileTableConfig.Path, "/")
	var deleted []string

	if err := f.List(ctx, "", "", func(s string) error {
		__antithesis_instrumentation__.Notify(35093)
		if pattern != "" {
			__antithesis_instrumentation__.Notify(35096)
			ok, err := path.Match(pattern, s)
			if err != nil || func() bool {
				__antithesis_instrumentation__.Notify(35097)
				return !ok == true
			}() == true {
				__antithesis_instrumentation__.Notify(35098)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35099)
			}
		} else {
			__antithesis_instrumentation__.Notify(35100)
		}
		__antithesis_instrumentation__.Notify(35094)
		if err := errors.WithDetailf(f.Delete(ctx, s), "deleting %s failed", s); err != nil {
			__antithesis_instrumentation__.Notify(35101)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35102)
		}
		__antithesis_instrumentation__.Notify(35095)
		deleted = append(deleted, displayRoot+s)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(35103)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(35104)
	}
	__antithesis_instrumentation__.Notify(35082)

	return deleted, nil
}

func renameUserFile(
	ctx context.Context, conn clisqlclient.Conn, oldFilename,
	newFilename, qualifiedTableName string,
) error {
	__antithesis_instrumentation__.Notify(35105)
	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(35113)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35114)
	}
	__antithesis_instrumentation__.Notify(35106)

	ex := conn.GetDriverConn()
	if _, err := ex.ExecContext(ctx, `BEGIN`, nil); err != nil {
		__antithesis_instrumentation__.Notify(35115)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35116)
	}
	__antithesis_instrumentation__.Notify(35107)

	stmt, err := conn.GetDriverConn().Prepare(fmt.Sprintf(`UPDATE %s SET filename=$1 WHERE filename=$2`,
		qualifiedTableName+fileTableNameSuffix))
	if err != nil {
		__antithesis_instrumentation__.Notify(35117)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35118)
	}
	__antithesis_instrumentation__.Notify(35108)

	defer func() {
		__antithesis_instrumentation__.Notify(35119)
		if stmt != nil {
			__antithesis_instrumentation__.Notify(35120)
			_ = stmt.Close()
			_, _ = ex.ExecContext(ctx, `ROLLBACK`, nil)
		} else {
			__antithesis_instrumentation__.Notify(35121)
		}
	}()
	__antithesis_instrumentation__.Notify(35109)

	_, err = stmt.Exec([]driver.Value{newFilename, oldFilename})
	if err != nil {
		__antithesis_instrumentation__.Notify(35122)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35123)
	}
	__antithesis_instrumentation__.Notify(35110)

	if err := stmt.Close(); err != nil {
		__antithesis_instrumentation__.Notify(35124)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35125)
	}
	__antithesis_instrumentation__.Notify(35111)
	stmt = nil

	if _, err := ex.ExecContext(ctx, `COMMIT`, nil); err != nil {
		__antithesis_instrumentation__.Notify(35126)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35127)
	}
	__antithesis_instrumentation__.Notify(35112)

	return nil
}

func uploadUserFile(
	ctx context.Context, conn clisqlclient.Conn, source, destination string,
) (string, error) {
	__antithesis_instrumentation__.Notify(35128)
	reader, err := openUserFile(source)
	if err != nil {
		__antithesis_instrumentation__.Notify(35142)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35143)
	}
	__antithesis_instrumentation__.Notify(35129)
	defer reader.Close()

	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(35144)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35145)
	}
	__antithesis_instrumentation__.Notify(35130)

	ex := conn.GetDriverConn()
	if _, err := ex.ExecContext(ctx, `BEGIN`, nil); err != nil {
		__antithesis_instrumentation__.Notify(35146)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35147)
	}
	__antithesis_instrumentation__.Notify(35131)

	connURL, err := url.Parse(conn.GetURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(35148)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35149)
	}
	__antithesis_instrumentation__.Notify(35132)

	username, err := security.MakeSQLUsernameFromUserInput(connURL.User.Username(), security.UsernameCreation)
	if err != nil {
		__antithesis_instrumentation__.Notify(35150)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35151)
	}
	__antithesis_instrumentation__.Notify(35133)

	userfileURI := constructUserfileDestinationURI(source, destination, username)

	unescapedUserfileURL, err := url.PathUnescape(userfileURI)

	unescapedUserfileURL = unescapedUserfileURL + tmpSuffix
	if err != nil {
		__antithesis_instrumentation__.Notify(35152)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35153)
	}
	__antithesis_instrumentation__.Notify(35134)
	stmt, err := conn.GetDriverConn().Prepare(sql.CopyInFileStmt(unescapedUserfileURL, sql.CrdbInternalName,
		sql.UserFileUploadTable))
	if err != nil {
		__antithesis_instrumentation__.Notify(35154)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35155)
	}
	__antithesis_instrumentation__.Notify(35135)

	defer func() {
		__antithesis_instrumentation__.Notify(35156)
		if stmt != nil {
			__antithesis_instrumentation__.Notify(35157)
			_ = stmt.Close()
			_, _ = ex.ExecContext(ctx, `ROLLBACK`, nil)
		} else {
			__antithesis_instrumentation__.Notify(35158)
		}
	}()
	__antithesis_instrumentation__.Notify(35136)

	send := make([]byte, chunkSize)
	for {
		__antithesis_instrumentation__.Notify(35159)
		n, err := reader.Read(send)
		if n > 0 {
			__antithesis_instrumentation__.Notify(35160)

			_, err = stmt.Exec([]driver.Value{string(send[:n])})
			if err != nil {
				__antithesis_instrumentation__.Notify(35161)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(35162)
			}
		} else {
			__antithesis_instrumentation__.Notify(35163)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(35164)
				break
			} else {
				__antithesis_instrumentation__.Notify(35165)
				if err != nil {
					__antithesis_instrumentation__.Notify(35166)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(35167)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(35137)
	if err := stmt.Close(); err != nil {
		__antithesis_instrumentation__.Notify(35168)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35169)
	}
	__antithesis_instrumentation__.Notify(35138)
	stmt = nil

	if _, err := ex.ExecContext(ctx, `COMMIT`, nil); err != nil {
		__antithesis_instrumentation__.Notify(35170)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35171)
	}
	__antithesis_instrumentation__.Notify(35139)

	tmpURL, err := url.Parse(unescapedUserfileURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(35172)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35173)
	}
	__antithesis_instrumentation__.Notify(35140)
	err = renameUserFile(ctx, conn, tmpURL.Path, strings.TrimSuffix(tmpURL.Path, tmpSuffix),
		tmpURL.Host)
	if err != nil {
		__antithesis_instrumentation__.Notify(35174)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(35175)
	}
	__antithesis_instrumentation__.Notify(35141)

	return strings.TrimSuffix(unescapedUserfileURL, tmpSuffix), nil
}

var userFileCmds = []*cobra.Command{
	userFileUploadCmd,
	userFileListCmd,
	userFileGetCmd,
	userFileDeleteCmd,
}

var userFileCmd = &cobra.Command{
	Use:   "userfile [command]",
	Short: "upload, list and delete user scoped files",
	Long:  "Upload, list and delete files from the user scoped file storage.",
	RunE:  UsageAndErr,
}

func init() {
	userFileCmd.AddCommand(userFileCmds...)
}
