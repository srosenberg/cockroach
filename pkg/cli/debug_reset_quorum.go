package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/spf13/cobra"
)

var debugResetQuorumCmd = &cobra.Command{
	Use: "reset-quorum [range ID]",
	Short: "Reset quorum on the given range" +
		" by designating the target node as the sole voter.",
	Long: `
Reset quorum on the given range by designating the current node as 
the sole voter. Any existing data for the range is discarded. 

This command is UNSAFE and should only be used with the supervision 
of Cockroach Labs support. It is a last-resort option to recover a 
specified range after multiple node failures and loss of quorum.

Data on any surviving replicas will not be used to restore quorum. 
Instead, these replicas will be removed irrevocably.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDebugResetQuorum),
}

func runDebugResetQuorum(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31639)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rangeID, err := strconv.ParseInt(args[0], 10, 32)
	if err != nil {
		__antithesis_instrumentation__.Notify(31643)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31644)
	}
	__antithesis_instrumentation__.Notify(31640)

	cc, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(31645)
		log.Errorf(ctx, "connection to server failed: %v", err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31646)
	}
	__antithesis_instrumentation__.Notify(31641)
	defer finish()

	_, err = roachpb.NewInternalClient(cc).ResetQuorum(ctx, &roachpb.ResetQuorumRequest{
		RangeID: int32(rangeID),
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(31647)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31648)
	}
	__antithesis_instrumentation__.Notify(31642)

	fmt.Printf("ok; please verify https://<console>/#/reports/range/%d", rangeID)

	return nil
}
