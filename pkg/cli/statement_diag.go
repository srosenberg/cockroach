package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"text/tabwriter"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var stmtDiagCmd = &cobra.Command{
	Use:   "statement-diag [command] [options]",
	Short: "commands for managing statement diagnostics bundles",
	Long: `This set of commands can be used to manage and download statement diagnostic
bundles, and to cancel outstanding diagnostics activation requests. Statement
diagnostics can be activated from the UI or using EXPLAIN ANALYZE (DEBUG).`,
	RunE: UsageAndErr,
}

var stmtDiagListCmd = &cobra.Command{
	Use:   "list",
	Short: "list available bundles and outstanding activation requests",
	Long: `List statement diagnostics that are available for download and outstanding
diagnostics activation requests.`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runStmtDiagList),
}

func runStmtDiagList(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34520)
	const timeFmt = "2006-01-02 15:04:05 MST"

	conn, err := makeSQLClient("cockroach statement-diag", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34527)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34528)
	}
	__antithesis_instrumentation__.Notify(34521)
	defer func() {
		__antithesis_instrumentation__.Notify(34529)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34522)

	ctx := context.Background()

	bundles, err := clisqlclient.StmtDiagListBundles(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(34530)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34531)
	}
	__antithesis_instrumentation__.Notify(34523)

	var buf bytes.Buffer
	if len(bundles) == 0 {
		__antithesis_instrumentation__.Notify(34532)
		fmt.Printf("No statement diagnostics bundles available.\n")
	} else {
		__antithesis_instrumentation__.Notify(34533)
		fmt.Printf("Statement diagnostics bundles:\n")
		w := tabwriter.NewWriter(&buf, 4, 0, 2, ' ', 0)
		fmt.Fprint(w, "  ID\tCollection time\tStatement\n")
		for _, b := range bundles {
			__antithesis_instrumentation__.Notify(34535)
			fmt.Fprintf(w, "  %d\t%s\t%s\n", b.ID, b.CollectedAt.UTC().Format(timeFmt), b.Statement)
		}
		__antithesis_instrumentation__.Notify(34534)
		_ = w.Flush()

		fmt.Println(buf.String())
	}
	__antithesis_instrumentation__.Notify(34524)

	reqs, err := clisqlclient.StmtDiagListOutstandingRequests(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(34536)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34537)
	}
	__antithesis_instrumentation__.Notify(34525)

	buf.Reset()
	if len(reqs) == 0 {
		__antithesis_instrumentation__.Notify(34538)
		fmt.Printf("No outstanding activation requests.\n")
	} else {
		__antithesis_instrumentation__.Notify(34539)
		fmt.Printf("Outstanding activation requests:\n")
		w := tabwriter.NewWriter(&buf, 4, 0, 2, ' ', 0)
		fmt.Fprint(w, "  ID\tActivation time\tStatement\tMin execution latency\tExpires at\n")
		for _, r := range reqs {
			__antithesis_instrumentation__.Notify(34541)
			minExecLatency := "N/A"
			if r.MinExecutionLatency != 0 {
				__antithesis_instrumentation__.Notify(34544)
				minExecLatency = r.MinExecutionLatency.String()
			} else {
				__antithesis_instrumentation__.Notify(34545)
			}
			__antithesis_instrumentation__.Notify(34542)
			expiresAt := "never"
			if !r.ExpiresAt.IsZero() {
				__antithesis_instrumentation__.Notify(34546)
				expiresAt = r.ExpiresAt.String()
			} else {
				__antithesis_instrumentation__.Notify(34547)
			}
			__antithesis_instrumentation__.Notify(34543)
			fmt.Fprintf(
				w, "  %d\t%s\t%s\t%s\t%s\n",
				r.ID, r.RequestedAt.UTC().Format(timeFmt), r.Statement, minExecLatency, expiresAt,
			)
		}
		__antithesis_instrumentation__.Notify(34540)
		_ = w.Flush()
		fmt.Print(buf.String())
	}
	__antithesis_instrumentation__.Notify(34526)

	return nil
}

var stmtDiagDownloadCmd = &cobra.Command{
	Use:   "download <bundle id> [<filename>]",
	Short: "download statement diagnostics bundle into a zip file",
	Long: `Download statement diagnostics bundle into a zip file, using an ID returned by
the list command.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: clierrorplus.MaybeDecorateError(runStmtDiagDownload),
}

func runStmtDiagDownload(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34548)
	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(34554)
		return id < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(34555)
		return errors.New("invalid bundle ID")
	} else {
		__antithesis_instrumentation__.Notify(34556)
	}
	__antithesis_instrumentation__.Notify(34549)
	var filename string
	if len(args) > 1 {
		__antithesis_instrumentation__.Notify(34557)
		filename = args[1]
	} else {
		__antithesis_instrumentation__.Notify(34558)
		filename = fmt.Sprintf("stmt-bundle-%d.zip", id)
	}
	__antithesis_instrumentation__.Notify(34550)

	conn, err := makeSQLClient("cockroach statement-diag", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34559)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34560)
	}
	__antithesis_instrumentation__.Notify(34551)
	defer func() {
		__antithesis_instrumentation__.Notify(34561)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34552)

	if err := clisqlclient.StmtDiagDownloadBundle(
		context.Background(), conn, id, filename); err != nil {
		__antithesis_instrumentation__.Notify(34562)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34563)
	}
	__antithesis_instrumentation__.Notify(34553)
	fmt.Printf("Bundle saved to %q\n", filename)
	return nil
}

var stmtDiagDeleteCmd = &cobra.Command{
	Use:   "delete { --all | <bundle id> }",
	Short: "delete statement diagnostics bundles",
	Long: `Delete a statement diagnostics bundle using an ID returned by the list
command, or delete all bundles.`,
	Args: cobra.MaximumNArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runStmtDiagDelete),
}

func runStmtDiagDelete(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34564)
	conn, err := makeSQLClient("cockroach statement-diag", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34570)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34571)
	}
	__antithesis_instrumentation__.Notify(34565)
	defer func() {
		__antithesis_instrumentation__.Notify(34572)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34566)

	ctx := context.Background()

	if stmtDiagCtx.all {
		__antithesis_instrumentation__.Notify(34573)
		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(34575)
			return errors.New("extra arguments with --all")
		} else {
			__antithesis_instrumentation__.Notify(34576)
		}
		__antithesis_instrumentation__.Notify(34574)
		return clisqlclient.StmtDiagDeleteAllBundles(ctx, conn)
	} else {
		__antithesis_instrumentation__.Notify(34577)
	}
	__antithesis_instrumentation__.Notify(34567)
	if len(args) != 1 {
		__antithesis_instrumentation__.Notify(34578)
		return fmt.Errorf("accepts 1 arg, received %d", len(args))
	} else {
		__antithesis_instrumentation__.Notify(34579)
	}
	__antithesis_instrumentation__.Notify(34568)

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(34580)
		return id < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(34581)
		return errors.New("invalid ID")
	} else {
		__antithesis_instrumentation__.Notify(34582)
	}
	__antithesis_instrumentation__.Notify(34569)

	return clisqlclient.StmtDiagDeleteBundle(ctx, conn, id)
}

var stmtDiagCancelCmd = &cobra.Command{
	Use:   "cancel { -all | <request id> }",
	Short: "cancel outstanding activation requests",
	Long: `Cancel an outstanding activation request, using an ID returned by the
list command, or cancel all outstanding requests.`,
	Args: cobra.MaximumNArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runStmtDiagCancel),
}

func runStmtDiagCancel(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(34583)
	conn, err := makeSQLClient("cockroach statement-diag", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(34589)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34590)
	}
	__antithesis_instrumentation__.Notify(34584)
	defer func() {
		__antithesis_instrumentation__.Notify(34591)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(34585)

	ctx := context.Background()

	if stmtDiagCtx.all {
		__antithesis_instrumentation__.Notify(34592)
		if len(args) > 0 {
			__antithesis_instrumentation__.Notify(34594)
			return errors.New("extra arguments with --all")
		} else {
			__antithesis_instrumentation__.Notify(34595)
		}
		__antithesis_instrumentation__.Notify(34593)
		return clisqlclient.StmtDiagCancelAllOutstandingRequests(ctx, conn)
	} else {
		__antithesis_instrumentation__.Notify(34596)
	}
	__antithesis_instrumentation__.Notify(34586)
	if len(args) != 1 {
		__antithesis_instrumentation__.Notify(34597)
		return fmt.Errorf("accepts 1 arg, received %d", len(args))
	} else {
		__antithesis_instrumentation__.Notify(34598)
	}
	__antithesis_instrumentation__.Notify(34587)

	id, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil || func() bool {
		__antithesis_instrumentation__.Notify(34599)
		return id < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(34600)
		return errors.New("invalid ID")
	} else {
		__antithesis_instrumentation__.Notify(34601)
	}
	__antithesis_instrumentation__.Notify(34588)

	return clisqlclient.StmtDiagCancelOutstandingRequest(ctx, conn, id)
}

var stmtDiagCmds = []*cobra.Command{
	stmtDiagListCmd,
	stmtDiagDownloadCmd,
	stmtDiagDeleteCmd,
	stmtDiagCancelCmd,
}

func init() {
	stmtDiagCmd.AddCommand(stmtDiagCmds...)
}
