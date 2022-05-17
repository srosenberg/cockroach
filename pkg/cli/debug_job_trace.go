package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	tracezipper "github.com/cockroachdb/cockroach/pkg/util/tracing/zipper"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var debugJobTraceFromClusterCmd = &cobra.Command{
	Use:   "job-trace <job_id> --url=<cluster connection string>",
	Short: "get the trace payloads for the executing job",
	Args:  cobra.MinimumNArgs(1),
	RunE:  clierrorplus.MaybeDecorateError(runDebugJobTrace),
}

const jobTraceZipSuffix = "job-trace.zip"

func runDebugJobTrace(_ *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(31094)
	jobID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		__antithesis_instrumentation__.Notify(31098)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31099)
	}
	__antithesis_instrumentation__.Notify(31095)

	sqlConn, err := makeSQLClient("cockroach debug job-trace", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(31100)
		return errors.Wrap(err, "could not establish connection to cluster")
	} else {
		__antithesis_instrumentation__.Notify(31101)
	}
	__antithesis_instrumentation__.Notify(31096)
	defer func() {
		__antithesis_instrumentation__.Notify(31102)
		resErr = errors.CombineErrors(resErr, sqlConn.Close())
	}()
	__antithesis_instrumentation__.Notify(31097)

	return constructJobTraceZipBundle(context.Background(), sqlConn, jobID)
}

func getJobTraceID(sqlConn clisqlclient.Conn, jobID int64) (int64, error) {
	__antithesis_instrumentation__.Notify(31103)
	var traceID int64
	rows, err := sqlConn.Query(context.Background(),
		`SELECT trace_id FROM crdb_internal.jobs WHERE job_id=$1`, jobID)
	if err != nil {
		__antithesis_instrumentation__.Notify(31109)
		return traceID, err
	} else {
		__antithesis_instrumentation__.Notify(31110)
	}
	__antithesis_instrumentation__.Notify(31104)
	vals := make([]driver.Value, 1)
	for {
		__antithesis_instrumentation__.Notify(31111)
		var err error
		if err = rows.Next(vals); err == io.EOF {
			__antithesis_instrumentation__.Notify(31113)
			break
		} else {
			__antithesis_instrumentation__.Notify(31114)
		}
		__antithesis_instrumentation__.Notify(31112)
		if err != nil {
			__antithesis_instrumentation__.Notify(31115)
			return traceID, err
		} else {
			__antithesis_instrumentation__.Notify(31116)
		}
	}
	__antithesis_instrumentation__.Notify(31105)
	if err := rows.Close(); err != nil {
		__antithesis_instrumentation__.Notify(31117)
		return traceID, err
	} else {
		__antithesis_instrumentation__.Notify(31118)
	}
	__antithesis_instrumentation__.Notify(31106)
	if vals[0] == nil {
		__antithesis_instrumentation__.Notify(31119)
		return traceID, errors.Newf("no job entry found for %d", jobID)
	} else {
		__antithesis_instrumentation__.Notify(31120)
	}
	__antithesis_instrumentation__.Notify(31107)
	var ok bool
	traceID, ok = vals[0].(int64)
	if !ok {
		__antithesis_instrumentation__.Notify(31121)
		return traceID, errors.New("failed to parse traceID")
	} else {
		__antithesis_instrumentation__.Notify(31122)
	}
	__antithesis_instrumentation__.Notify(31108)
	return traceID, nil
}

func constructJobTraceZipBundle(ctx context.Context, sqlConn clisqlclient.Conn, jobID int64) error {
	__antithesis_instrumentation__.Notify(31123)

	if cliCtx.cmdTimeout != 0 {
		__antithesis_instrumentation__.Notify(31129)
		if err := sqlConn.Exec(context.Background(),
			`SET statement_timeout = $1`, cliCtx.cmdTimeout.String()); err != nil {
			__antithesis_instrumentation__.Notify(31130)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31131)
		}
	} else {
		__antithesis_instrumentation__.Notify(31132)
	}
	__antithesis_instrumentation__.Notify(31124)

	traceID, err := getJobTraceID(sqlConn, jobID)
	if err != nil {
		__antithesis_instrumentation__.Notify(31133)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31134)
	}
	__antithesis_instrumentation__.Notify(31125)

	zipper := tracezipper.MakeSQLConnInflightTraceZipper(sqlConn.GetDriverConn().(driver.QueryerContext))
	zipBytes, err := zipper.Zip(ctx, traceID)
	if err != nil {
		__antithesis_instrumentation__.Notify(31135)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31136)
	}
	__antithesis_instrumentation__.Notify(31126)

	var f *os.File
	filename := fmt.Sprintf("%d-%s", jobID, jobTraceZipSuffix)
	if f, err = os.Create(filename); err != nil {
		__antithesis_instrumentation__.Notify(31137)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31138)
	}
	__antithesis_instrumentation__.Notify(31127)
	_, err = f.Write(zipBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(31139)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31140)
	}
	__antithesis_instrumentation__.Notify(31128)
	defer f.Close()
	return nil
}
