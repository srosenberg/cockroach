package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/spf13/cobra"
)

var connectCmds = []*cobra.Command{
	connectInitCmd,
	connectJoinCmd,
}

var connectCmd = &cobra.Command{
	Use:   "connect [command]",
	Short: "Create certificates for securely connecting with clusters\n",
	Long: `
Bootstrap security certificates for connecting to new or existing clusters.`,
	RunE: UsageAndErr,
}

func init() {
	connectCmd.AddCommand(connectCmds...)
}

var connectInitCmd = &cobra.Command{
	Use:   "init --certs-dir=<path to cockroach certs dir> --init-token=<shared secret> --join=<host 1>,<host 2>,...,<host N>",
	Short: "auto-build TLS certificates for use with the start command",
	Long: `
Connects to other nodes and negotiates an initialization bundle for use with
secure inter-node connections.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runConnectInit),
}

func runConnectInit(cmd *cobra.Command, args []string) (retErr error) {
	__antithesis_instrumentation__.Notify(30232)
	if err := validateConnectInitFlags(cmd, true); err != nil {
		__antithesis_instrumentation__.Notify(30242)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30243)
	}
	__antithesis_instrumentation__.Notify(30233)

	cl := security.MakeCertsLocator(baseCfg.SSLCertsDir)
	if exists, err := cl.HasNodeCert(); err != nil {
		__antithesis_instrumentation__.Notify(30244)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30245)
		if exists {
			__antithesis_instrumentation__.Notify(30246)
			return errors.Newf("node certificate already exists in %s", baseCfg.SSLCertsDir)
		} else {
			__antithesis_instrumentation__.Notify(30247)
		}
	}
	__antithesis_instrumentation__.Notify(30234)

	defer log.Flush()

	peers := []string(serverCfg.JoinList)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = logtags.AddTag(ctx, "connect", nil)

	log.Infof(ctx, "validating the command line network arguments")

	if err := baseCfg.ValidateAddrs(ctx); err != nil {
		__antithesis_instrumentation__.Notify(30248)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30249)
	}
	__antithesis_instrumentation__.Notify(30235)

	log.Ops.Infof(ctx, "starting the initial network listeners")

	rpcLn, err := server.ListenAndUpdateAddrs(ctx, &baseCfg.Addr, &baseCfg.AdvertiseAddr, "rpc")
	if err != nil {
		__antithesis_instrumentation__.Notify(30250)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30251)
	}
	__antithesis_instrumentation__.Notify(30236)
	log.Ops.Infof(ctx, "started rpc listener at: %s", rpcLn.Addr())
	defer func() {
		__antithesis_instrumentation__.Notify(30252)
		_ = rpcLn.Close()
	}()
	__antithesis_instrumentation__.Notify(30237)
	httpLn, err := server.ListenAndUpdateAddrs(ctx, &baseCfg.HTTPAddr, &baseCfg.HTTPAdvertiseAddr, "http")
	if err != nil {
		__antithesis_instrumentation__.Notify(30253)
		return err
	} else {
		__antithesis_instrumentation__.Notify(30254)
	}
	__antithesis_instrumentation__.Notify(30238)
	log.Ops.Infof(ctx, "started http listener at: %s", httpLn.Addr())
	if baseCfg.SplitListenSQL {
		__antithesis_instrumentation__.Notify(30255)
		sqlLn, err := server.ListenAndUpdateAddrs(ctx, &baseCfg.SQLAddr, &baseCfg.SQLAdvertiseAddr, "sql")
		if err != nil {
			__antithesis_instrumentation__.Notify(30257)
			return err
		} else {
			__antithesis_instrumentation__.Notify(30258)
		}
		__antithesis_instrumentation__.Notify(30256)
		log.Ops.Infof(ctx, "started sql listener at: %s", sqlLn.Addr())
		_ = sqlLn.Close()
	} else {
		__antithesis_instrumentation__.Notify(30259)
	}
	__antithesis_instrumentation__.Notify(30239)

	_ = httpLn.Close()

	defer func() {
		__antithesis_instrumentation__.Notify(30260)
		if retErr == nil {
			__antithesis_instrumentation__.Notify(30261)
			fmt.Println("server certificate generation complete.\n\n" +
				"cert files generated in: " + os.ExpandEnv(baseCfg.SSLCertsDir) + "\n\n" +
				"Do not forget to generate a client certificate for the 'root' user!\n" +
				"This must be done manually, preferably from a different unix user account\n" +
				"than the one running the server. Example command:\n\n" +
				"   " + os.Args[0] + " cert create-client root --ca-key=<path-to-client-ca-key>\n")
		} else {
			__antithesis_instrumentation__.Notify(30262)
		}
	}()
	__antithesis_instrumentation__.Notify(30240)

	reporter := func(format string, args ...interface{}) {
		__antithesis_instrumentation__.Notify(30263)
		fmt.Printf(format+"\n", args...)
	}
	__antithesis_instrumentation__.Notify(30241)

	return server.InitHandshake(ctx, reporter, baseCfg, startCtx.initToken, startCtx.numExpectedNodes, peers, baseCfg.SSLCertsDir, rpcLn)
}

func validateConnectInitFlags(cmd *cobra.Command, requireExplicitFlags bool) error {
	__antithesis_instrumentation__.Notify(30264)
	if requireExplicitFlags {
		__antithesis_instrumentation__.Notify(30269)
		f := flagSetForCmd(cmd)
		if !(f.Lookup(cliflags.SingleNode.Name).Changed || func() bool {
			__antithesis_instrumentation__.Notify(30271)
			return (f.Lookup(cliflags.NumExpectedInitialNodes.Name).Changed && func() bool {
				__antithesis_instrumentation__.Notify(30272)
				return f.Lookup(cliflags.InitToken.Name).Changed == true
			}() == true) == true
		}() == true) {
			__antithesis_instrumentation__.Notify(30273)
			return errors.Newf("either --%s must be passed, or both --%s and --%s",
				cliflags.SingleNode.Name, cliflags.NumExpectedInitialNodes.Name, cliflags.InitToken.Name)
		} else {
			__antithesis_instrumentation__.Notify(30274)
		}
		__antithesis_instrumentation__.Notify(30270)
		if f.Lookup(cliflags.SingleNode.Name).Changed && func() bool {
			__antithesis_instrumentation__.Notify(30275)
			return (f.Lookup(cliflags.NumExpectedInitialNodes.Name).Changed || func() bool {
				__antithesis_instrumentation__.Notify(30276)
				return f.Lookup(cliflags.InitToken.Name).Changed == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(30277)
			return errors.Newf("--%s cannot be specified together with --%s or --%s",
				cliflags.SingleNode.Name, cliflags.NumExpectedInitialNodes.Name, cliflags.InitToken.Name)
		} else {
			__antithesis_instrumentation__.Notify(30278)
		}
	} else {
		__antithesis_instrumentation__.Notify(30279)
	}
	__antithesis_instrumentation__.Notify(30265)

	if startCtx.genCertsForSingleNode {
		__antithesis_instrumentation__.Notify(30280)
		startCtx.numExpectedNodes = 1
		startCtx.initToken = "start-single-node"
		return nil
	} else {
		__antithesis_instrumentation__.Notify(30281)
	}
	__antithesis_instrumentation__.Notify(30266)

	if startCtx.numExpectedNodes < 1 {
		__antithesis_instrumentation__.Notify(30282)
		return errors.Newf("flag --%s must be set to a value greater than or equal to 1",
			cliflags.NumExpectedInitialNodes.Name)
	} else {
		__antithesis_instrumentation__.Notify(30283)
	}
	__antithesis_instrumentation__.Notify(30267)
	if startCtx.initToken == "" {
		__antithesis_instrumentation__.Notify(30284)
		return errors.Newf("flag --%s must be set to a non-empty string",
			cliflags.InitToken.Name)
	} else {
		__antithesis_instrumentation__.Notify(30285)
	}
	__antithesis_instrumentation__.Notify(30268)
	return nil
}
