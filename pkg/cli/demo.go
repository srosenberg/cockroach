package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/democluster"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "open a demo sql shell",
	Long: `
Start an in-memory, standalone, single-node CockroachDB instance, and open an
interactive SQL prompt to it. Various datasets are available to be preloaded as
subcommands: e.g. "cockroach demo startrek". See --help for a full list.

By default, the 'movr' dataset is pre-loaded. You can also use --no-example-database
to avoid pre-loading a dataset.

cockroach demo attempts to connect to a Cockroach Labs server to obtain a
temporary enterprise license for demoing enterprise features and enable
telemetry back to Cockroach Labs. In order to disable this behavior, set the
environment variable "COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING" to true.
`,
	Example: `  cockroach demo`,
	Args:    cobra.NoArgs,
}

func init() {
	demoCmd.RunE = clierrorplus.MaybeDecorateError(func(cmd *cobra.Command, _ []string) error {
		return runDemo(cmd, nil)
	})
}

const defaultGeneratorName = "movr"

var defaultGenerator workload.Generator

var demoNodeCacheSizeValue = newBytesOrPercentageValue(
	&demoCtx.CacheSize,
	memoryPercentResolver,
)
var demoNodeSQLMemSizeValue = newBytesOrPercentageValue(
	&demoCtx.SQLPoolMemorySize,
	memoryPercentResolver,
)

func init() {
	for _, meta := range workload.Registered() {
		gen := meta.New()

		if meta.Name == defaultGeneratorName {

			defaultGenerator = gen
		}

		var genFlags *pflag.FlagSet
		if f, ok := gen.(workload.Flagser); ok {
			genFlags = f.Flags().FlagSet
		}

		genDemoCmd := &cobra.Command{
			Use:   meta.Name,
			Short: meta.Description,
			Args:  cobra.ArbitraryArgs,
			RunE: clierrorplus.MaybeDecorateError(func(cmd *cobra.Command, _ []string) error {
				return runDemo(cmd, gen)
			}),
		}
		if !meta.PublicFacing {
			genDemoCmd.Hidden = true
		}
		demoCmd.AddCommand(genDemoCmd)
		genDemoCmd.Flags().AddFlagSet(genFlags)
	}
}

func incrementTelemetryCounters(cmd *cobra.Command) {
	__antithesis_instrumentation__.Notify(31829)
	incrementDemoCounter(demo)
	if flagSetForCmd(cmd).Lookup(cliflags.DemoNodes.Name).Changed {
		__antithesis_instrumentation__.Notify(31833)
		incrementDemoCounter(nodes)
	} else {
		__antithesis_instrumentation__.Notify(31834)
	}
	__antithesis_instrumentation__.Notify(31830)
	if demoCtx.Localities != nil {
		__antithesis_instrumentation__.Notify(31835)
		incrementDemoCounter(demoLocality)
	} else {
		__antithesis_instrumentation__.Notify(31836)
	}
	__antithesis_instrumentation__.Notify(31831)
	if demoCtx.RunWorkload {
		__antithesis_instrumentation__.Notify(31837)
		incrementDemoCounter(withLoad)
	} else {
		__antithesis_instrumentation__.Notify(31838)
	}
	__antithesis_instrumentation__.Notify(31832)
	if demoCtx.GeoPartitionedReplicas {
		__antithesis_instrumentation__.Notify(31839)
		incrementDemoCounter(geoPartitionedReplicas)
	} else {
		__antithesis_instrumentation__.Notify(31840)
	}
}

func checkDemoConfiguration(
	cmd *cobra.Command, gen workload.Generator,
) (workload.Generator, error) {
	__antithesis_instrumentation__.Notify(31841)
	f := flagSetForCmd(cmd)
	if gen == nil && func() bool {
		__antithesis_instrumentation__.Notify(31847)
		return !demoCtx.NoExampleDatabase == true
	}() == true {
		__antithesis_instrumentation__.Notify(31848)

		gen = defaultGenerator
	} else {
		__antithesis_instrumentation__.Notify(31849)
	}
	__antithesis_instrumentation__.Notify(31842)

	if demoCtx.RunWorkload && func() bool {
		__antithesis_instrumentation__.Notify(31850)
		return demoCtx.NoExampleDatabase == true
	}() == true {
		__antithesis_instrumentation__.Notify(31851)
		return nil, errors.New("cannot run a workload when generation of the example database is disabled")
	} else {
		__antithesis_instrumentation__.Notify(31852)
	}
	__antithesis_instrumentation__.Notify(31843)

	if demoCtx.NumNodes <= 0 {
		__antithesis_instrumentation__.Notify(31853)
		return nil, errors.Newf("--%s has invalid value (expected positive, got %d)", cliflags.DemoNodes.Name, demoCtx.NumNodes)
	} else {
		__antithesis_instrumentation__.Notify(31854)
	}
	__antithesis_instrumentation__.Notify(31844)

	if demoCtx.SimulateLatency && func() bool {
		__antithesis_instrumentation__.Notify(31855)
		return demoCtx.Localities != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(31856)
		return nil, errors.Newf("--%s cannot be used with --%s", cliflags.Global.Name, cliflags.DemoNodeLocality.Name)
	} else {
		__antithesis_instrumentation__.Notify(31857)
	}
	__antithesis_instrumentation__.Notify(31845)

	demoCtx.DisableTelemetry = cluster.TelemetryOptOut()

	demoCtx.DisableLicenseAcquisition =
		demoCtx.DisableTelemetry || func() bool {
			__antithesis_instrumentation__.Notify(31858)
			return (democluster.GetAndApplyLicense == nil) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(31859)
			return demoCtx.DisableLicenseAcquisition == true
		}() == true

	if demoCtx.GeoPartitionedReplicas {
		__antithesis_instrumentation__.Notify(31860)
		geoFlag := "--" + cliflags.DemoGeoPartitionedReplicas.Name
		if demoCtx.DisableLicenseAcquisition {
			__antithesis_instrumentation__.Notify(31867)
			return nil, errors.Newf("enterprise features are needed for this demo (%s)", geoFlag)
		} else {
			__antithesis_instrumentation__.Notify(31868)
		}
		__antithesis_instrumentation__.Notify(31861)

		if demoCtx.NoExampleDatabase {
			__antithesis_instrumentation__.Notify(31869)
			return nil, errors.New("cannot setup geo-partitioned replicas topology without generating an example database")
		} else {
			__antithesis_instrumentation__.Notify(31870)
		}
		__antithesis_instrumentation__.Notify(31862)

		if gen == nil || func() bool {
			__antithesis_instrumentation__.Notify(31871)
			return gen.Meta().Name != "movr" == true
		}() == true {
			__antithesis_instrumentation__.Notify(31872)
			return nil, errors.Newf("%s must be used with the Movr dataset", geoFlag)
		} else {
			__antithesis_instrumentation__.Notify(31873)
		}
		__antithesis_instrumentation__.Notify(31863)

		if demoCtx.Localities != nil {
			__antithesis_instrumentation__.Notify(31874)
			return nil, errors.Newf("--demo-locality cannot be used with %s", geoFlag)
		} else {
			__antithesis_instrumentation__.Notify(31875)
		}
		__antithesis_instrumentation__.Notify(31864)

		if f.Lookup(cliflags.DemoNodes.Name).Changed {
			__antithesis_instrumentation__.Notify(31876)
			if demoCtx.NumNodes != 9 {
				__antithesis_instrumentation__.Notify(31877)
				return nil, errors.Newf("--nodes with a value different from 9 cannot be used with %s", geoFlag)
			} else {
				__antithesis_instrumentation__.Notify(31878)
			}
		} else {
			__antithesis_instrumentation__.Notify(31879)
			demoCtx.NumNodes = 9
			cliCtx.PrintlnUnlessEmbedded(

				`#
# --geo-partitioned replicas operates on a 9 node cluster.
# The cluster size has been changed from the default to 9 nodes.`)
		}
		__antithesis_instrumentation__.Notify(31865)

		configErr := errors.Newf(
			"workload %s is not configured to have a partitioning step", gen.Meta().Name)
		hookser, ok := gen.(workload.Hookser)
		if !ok {
			__antithesis_instrumentation__.Notify(31880)
			return nil, configErr
		} else {
			__antithesis_instrumentation__.Notify(31881)
		}
		__antithesis_instrumentation__.Notify(31866)
		if hookser.Hooks().Partition == nil {
			__antithesis_instrumentation__.Notify(31882)
			return nil, configErr
		} else {
			__antithesis_instrumentation__.Notify(31883)
		}
	} else {
		__antithesis_instrumentation__.Notify(31884)
	}
	__antithesis_instrumentation__.Notify(31846)

	return gen, nil
}

func runDemo(cmd *cobra.Command, gen workload.Generator) (resErr error) {
	__antithesis_instrumentation__.Notify(31885)
	closeFn, err := sqlCtx.Open(os.Stdin)
	if err != nil {
		__antithesis_instrumentation__.Notify(31897)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31898)
	}
	__antithesis_instrumentation__.Notify(31886)
	defer closeFn()

	if gen, err = checkDemoConfiguration(cmd, gen); err != nil {
		__antithesis_instrumentation__.Notify(31899)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31900)
	}
	__antithesis_instrumentation__.Notify(31887)

	incrementTelemetryCounters(cmd)

	ctx := context.Background()

	demoCtx.WorkloadGenerator = gen

	c, err := democluster.NewDemoCluster(ctx, &demoCtx,
		log.Infof,
		log.Warningf,
		log.Ops.Shoutf,
		func(ctx context.Context) (*stop.Stopper, error) {
			__antithesis_instrumentation__.Notify(31901)

			serverCfg.Stores.Specs = nil
			return setupAndInitializeLoggingAndProfiling(ctx, cmd, false)
		},
		getAdminClient,
		func(ctx context.Context, ac serverpb.AdminClient) error {
			__antithesis_instrumentation__.Notify(31902)
			return drainAndShutdown(ctx, ac, "local")
		},
	)
	__antithesis_instrumentation__.Notify(31888)
	if err != nil {
		__antithesis_instrumentation__.Notify(31903)
		c.Close(ctx)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31904)
	}
	__antithesis_instrumentation__.Notify(31889)
	defer c.Close(ctx)

	initGEOS(ctx)

	if err := c.Start(ctx, runInitialSQL); err != nil {
		__antithesis_instrumentation__.Notify(31905)
		return clierrorplus.CheckAndMaybeShout(err)
	} else {
		__antithesis_instrumentation__.Notify(31906)
	}
	__antithesis_instrumentation__.Notify(31890)
	sqlCtx.ShellCtx.DemoCluster = c

	if cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(31907)
		cliCtx.PrintfUnlessEmbedded(`#
# Welcome to the CockroachDB demo database!
#
# You are connected to a temporary, in-memory CockroachDB cluster of %d node%s.
`, demoCtx.NumNodes, util.Pluralize(int64(demoCtx.NumNodes)))

		if demoCtx.Multitenant {
			__antithesis_instrumentation__.Notify(31910)
			cliCtx.PrintfUnlessEmbedded(`#
# You are connected to tenant 1, but can connect to the system tenant with
# \connect and the SQL url below.
`)
		} else {
			__antithesis_instrumentation__.Notify(31911)
		}
		__antithesis_instrumentation__.Notify(31908)

		if demoCtx.SimulateLatency {
			__antithesis_instrumentation__.Notify(31912)
			cliCtx.PrintfUnlessEmbedded(
				`# Communication between nodes will simulate real world latencies.
#
# WARNING: the use of --%s is experimental. Some features may not work as expected.
`,
				cliflags.Global.Name,
			)
		} else {
			__antithesis_instrumentation__.Notify(31913)
		}
		__antithesis_instrumentation__.Notify(31909)

		if demoCtx.DisableTelemetry {
			__antithesis_instrumentation__.Notify(31914)
			cliCtx.PrintlnUnlessEmbedded("#\n# Telemetry and automatic license acquisition disabled by configuration.")
		} else {
			__antithesis_instrumentation__.Notify(31915)
			if demoCtx.DisableLicenseAcquisition {
				__antithesis_instrumentation__.Notify(31916)
				cliCtx.PrintlnUnlessEmbedded("#\n# Enterprise features disabled by OSS-only build.")
			} else {
				__antithesis_instrumentation__.Notify(31917)
				cliCtx.PrintlnUnlessEmbedded("#\n# This demo session will attempt to enable enterprise features\n" +
					"# by acquiring a temporary license from Cockroach Labs in the background.\n" +
					"# To disable this behavior, set the environment variable\n" +
					"# COCKROACH_SKIP_ENABLING_DIAGNOSTIC_REPORTING=true.")
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(31918)
	}
	__antithesis_instrumentation__.Notify(31891)

	licenseDone, err := c.AcquireDemoLicense(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(31919)
		return clierrorplus.CheckAndMaybeShout(err)
	} else {
		__antithesis_instrumentation__.Notify(31920)
	}
	__antithesis_instrumentation__.Notify(31892)

	if err := c.SetupWorkload(ctx, licenseDone); err != nil {
		__antithesis_instrumentation__.Notify(31921)
		return clierrorplus.CheckAndMaybeShout(err)
	} else {
		__antithesis_instrumentation__.Notify(31922)
	}
	__antithesis_instrumentation__.Notify(31893)

	if cliCtx.IsInteractive {
		__antithesis_instrumentation__.Notify(31923)
		if gen != nil {
			__antithesis_instrumentation__.Notify(31926)
			fmt.Printf("#\n# The cluster has been preloaded with the %q dataset\n# (%s).\n",
				gen.Meta().Name, gen.Meta().Description)
		} else {
			__antithesis_instrumentation__.Notify(31927)
		}
		__antithesis_instrumentation__.Notify(31924)

		fmt.Println(`#
# Reminder: your changes to data stored in the demo session will not be saved!`)

		var nodeList strings.Builder
		c.ListDemoNodes(&nodeList, stderr, true)
		cliCtx.PrintlnUnlessEmbedded(

			`#
# If you wish to access this demo cluster using another tool, you will need
# the following details:
#
#   - Connection parameters:
#  `,
			strings.ReplaceAll(strings.TrimSuffix(nodeList.String(), "\n"), "\n", "\n#   "))

		if !demoCtx.Insecure {
			__antithesis_instrumentation__.Notify(31928)
			adminUser, adminPassword, certsDir := c.GetSQLCredentials()

			fmt.Printf(`#   - Username: %q, password: %q
#   - Directory with certificate files (for certain SQL drivers/tools): %s
#
`,
				adminUser,
				adminPassword,
				certsDir,
			)
		} else {
			__antithesis_instrumentation__.Notify(31929)
		}
		__antithesis_instrumentation__.Notify(31925)

		go func() {
			__antithesis_instrumentation__.Notify(31930)
			if err := waitForLicense(licenseDone); err != nil {
				__antithesis_instrumentation__.Notify(31931)
				_ = clierrorplus.CheckAndMaybeShout(err)
			} else {
				__antithesis_instrumentation__.Notify(31932)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(31933)

		if err := waitForLicense(licenseDone); err != nil {
			__antithesis_instrumentation__.Notify(31934)
			return clierrorplus.CheckAndMaybeShout(err)
		} else {
			__antithesis_instrumentation__.Notify(31935)
		}
	}
	__antithesis_instrumentation__.Notify(31894)

	conn, err := sqlCtx.MakeConn(c.GetConnURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(31936)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31937)
	}
	__antithesis_instrumentation__.Notify(31895)
	defer func() {
		__antithesis_instrumentation__.Notify(31938)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(31896)

	sqlCtx.ShellCtx.ParseURL = makeURLParser(cmd)
	return sqlCtx.Run(conn)
}

func waitForLicense(licenseDone <-chan error) error {
	__antithesis_instrumentation__.Notify(31939)
	err := <-licenseDone
	return err
}
