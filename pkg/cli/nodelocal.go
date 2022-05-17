package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

const (
	chunkSize = 4 * 1024
)

var nodeLocalUploadCmd = &cobra.Command{
	Use:   "upload <source> <destination>",
	Short: "Upload file from source to destination",
	Long: `
Uploads a file to a gateway node's local file system using a SQL connection.
`,
	Args: cobra.MinimumNArgs(2),
	RunE: clierrorplus.MaybeShoutError(runUpload),
}

func runUpload(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(33752)
	conn, err := makeSQLClient("cockroach nodelocal", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(33756)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33757)
	}
	__antithesis_instrumentation__.Notify(33753)
	defer func() {
		__antithesis_instrumentation__.Notify(33758)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(33754)

	source := args[0]
	destination := args[1]
	reader, err := openSourceFile(source)
	if err != nil {
		__antithesis_instrumentation__.Notify(33759)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33760)
	}
	__antithesis_instrumentation__.Notify(33755)
	defer reader.Close()

	return uploadFile(context.Background(), conn, reader, destination)
}

func openSourceFile(source string) (io.ReadCloser, error) {
	__antithesis_instrumentation__.Notify(33761)
	f, err := os.Open(source)
	if err != nil {
		__antithesis_instrumentation__.Notify(33765)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(33766)
	}
	__antithesis_instrumentation__.Notify(33762)
	stat, err := f.Stat()
	if err != nil {
		__antithesis_instrumentation__.Notify(33767)
		return nil, errors.Wrapf(err, "unable to get source file stats for %s", source)
	} else {
		__antithesis_instrumentation__.Notify(33768)
	}
	__antithesis_instrumentation__.Notify(33763)
	if stat.IsDir() {
		__antithesis_instrumentation__.Notify(33769)
		return nil, fmt.Errorf("source file %s is a directory, not a file", source)
	} else {
		__antithesis_instrumentation__.Notify(33770)
	}
	__antithesis_instrumentation__.Notify(33764)
	return f, nil
}

func uploadFile(
	ctx context.Context, conn clisqlclient.Conn, reader io.Reader, destination string,
) error {
	__antithesis_instrumentation__.Notify(33771)
	if err := conn.EnsureConn(); err != nil {
		__antithesis_instrumentation__.Notify(33780)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33781)
	}
	__antithesis_instrumentation__.Notify(33772)

	ex := conn.GetDriverConn()
	if _, err := ex.ExecContext(ctx, `BEGIN`, nil); err != nil {
		__antithesis_instrumentation__.Notify(33782)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33783)
	}
	__antithesis_instrumentation__.Notify(33773)

	nodelocalURL := url.URL{
		Scheme: "nodelocal",
		Host:   "self",
		Path:   destination,
	}
	stmt, err := conn.GetDriverConn().Prepare(sql.CopyInFileStmt(nodelocalURL.String(), sql.CrdbInternalName,
		sql.NodelocalFileUploadTable))
	if err != nil {
		__antithesis_instrumentation__.Notify(33784)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33785)
	}
	__antithesis_instrumentation__.Notify(33774)

	defer func() {
		__antithesis_instrumentation__.Notify(33786)
		if stmt != nil {
			__antithesis_instrumentation__.Notify(33787)
			_ = stmt.Close()
			_, _ = ex.ExecContext(ctx, `ROLLBACK`, nil)
		} else {
			__antithesis_instrumentation__.Notify(33788)
		}
	}()
	__antithesis_instrumentation__.Notify(33775)

	send := make([]byte, chunkSize)
	for {
		__antithesis_instrumentation__.Notify(33789)
		n, err := reader.Read(send)
		if n > 0 {
			__antithesis_instrumentation__.Notify(33790)

			_, err = stmt.Exec([]driver.Value{string(send[:n])})
			if err != nil {
				__antithesis_instrumentation__.Notify(33791)
				return err
			} else {
				__antithesis_instrumentation__.Notify(33792)
			}
		} else {
			__antithesis_instrumentation__.Notify(33793)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(33794)
				break
			} else {
				__antithesis_instrumentation__.Notify(33795)
				if err != nil {
					__antithesis_instrumentation__.Notify(33796)
					return err
				} else {
					__antithesis_instrumentation__.Notify(33797)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(33776)
	if err := stmt.Close(); err != nil {
		__antithesis_instrumentation__.Notify(33798)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33799)
	}
	__antithesis_instrumentation__.Notify(33777)
	stmt = nil

	if _, err := ex.ExecContext(ctx, `COMMIT`, nil); err != nil {
		__antithesis_instrumentation__.Notify(33800)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33801)
	}
	__antithesis_instrumentation__.Notify(33778)

	nodeID, _, _, err := conn.GetServerMetadata(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(33802)
		return errors.Wrap(err, "unable to get node id")
	} else {
		__antithesis_instrumentation__.Notify(33803)
	}
	__antithesis_instrumentation__.Notify(33779)
	fmt.Printf("successfully uploaded to nodelocal://%s\n", filepath.Join(roachpb.NodeID(nodeID).String(), destination))
	return nil
}

var nodeLocalCmds = []*cobra.Command{
	nodeLocalUploadCmd,
}

var nodeLocalCmd = &cobra.Command{
	Use:   "nodelocal [command]",
	Short: "upload and delete nodelocal files",
	Long:  "Upload and delete files on the gateway node's local file system.",
	RunE:  UsageAndErr,
}

func init() {
	nodeLocalCmd.AddCommand(nodeLocalCmds...)
}
