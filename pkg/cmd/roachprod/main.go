package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/roachprod"
	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	rperrors "github.com/cockroachdb/cockroach/pkg/roachprod/errors"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/ui"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "roachprod [command] (flags)",
	Short: "roachprod tool for manipulating test clusters",
	Long: `roachprod is a tool for manipulating ephemeral test clusters, allowing easy
creating, destruction, starting, stopping and wiping of clusters along with
running load generators.

Examples:

  roachprod create local -n 3
  roachprod start local
  roachprod sql local:2 -- -e "select * from crdb_internal.node_runtime_info"
  roachprod stop local
  roachprod wipe local
  roachprod destroy local

The above commands will create a "local" 3 node cluster, start a cockroach
cluster on these nodes, run a sql command on the 2nd node, stop, wipe and
destroy the cluster.
`,
	Version: "details:\n" + build.GetInfo().Long(),
}

func wrap(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	__antithesis_instrumentation__.Notify(42945)
	return func(cmd *cobra.Command, args []string) {
		__antithesis_instrumentation__.Notify(42946)
		err := f(cmd, args)
		if err != nil {
			__antithesis_instrumentation__.Notify(42947)
			roachprodError, ok := rperrors.AsError(err)
			if !ok {
				__antithesis_instrumentation__.Notify(42949)
				roachprodError = rperrors.Unclassified{Err: err}
				err = roachprodError
			} else {
				__antithesis_instrumentation__.Notify(42950)
			}
			__antithesis_instrumentation__.Notify(42948)

			cmd.Printf("Error: %+v\n", err)

			os.Exit(roachprodError.ExitCode())
		} else {
			__antithesis_instrumentation__.Notify(42951)
		}
	}
}

var createCmd = &cobra.Command{
	Use:   "create <cluster>",
	Short: "create a cluster",
	Long: `Create a local or cloud-based cluster.

A cluster is composed of a set of nodes, configured during cluster creation via
the --nodes flag. Creating a cluster does not start any processes on the nodes
other than the base system processes (e.g. sshd). See "roachprod start" for
starting cockroach nodes and "roachprod {run,ssh}" for running arbitrary
commands on the nodes of a cluster.

Cloud Clusters

  Cloud-based clusters are ephemeral and come with a lifetime (specified by the
  --lifetime flag) after which they will be automatically
  destroyed. Cloud-based clusters require the associated command line tool for
  the cloud to be installed and configured (e.g. "gcloud auth login").

  Clusters names are required to be prefixed by the authenticated user of the
  cloud service. The suffix is an arbitrary string used to distinguish
  clusters. For example, "marc-test" is a valid cluster name for the user
  "marc". The authenticated user for the cloud service is automatically
  detected and can be override by the ROACHPROD_USER environment variable or
  the --username flag.

  The machine type and the use of local SSD storage can be specified during
  cluster creation via the --{cloud}-machine-type and --local-ssd flags. The
  machine-type is cloud specified. For example, --gce-machine-type=n1-highcpu-8
  requests the "n1-highcpu-8" machine type for a GCE-based cluster. No attempt
  is made (or desired) to abstract machine types across cloud providers. See
  the cloud provider's documentation for details on the machine types
  available.

  The underlying filesystem can be provided using the --filesystem flag.
  Use --filesystem=zfs, for zfs, and --filesystem=ext4, for ext4. The default
  file system is ext4. The filesystem flag only works on gce currently.

Local Clusters

  A local cluster stores the per-node data in ${HOME}/local on the machine
  roachprod is being run on. Whether a cluster is local is specified on creation
  by using the name 'local' or 'local-<anything>'. Local clusters have no expiration.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) (retErr error) {
		__antithesis_instrumentation__.Notify(42952)
		createVMOpts.ClusterName = args[0]
		return roachprod.Create(context.Background(), roachprodLibraryLogger, username, numNodes, createVMOpts, providerOptsContainer)
	}),
}

var setupSSHCmd = &cobra.Command{
	Use:   "setup-ssh <cluster>",
	Short: "set up ssh for a cluster",
	Long: `Sets up the keys and host keys for the vms in the cluster.

It first resets the machine credentials as though the cluster were newly created
using the cloud provider APIs and then proceeds to ensure that the hosts can
SSH into eachother and lastly adds additional public keys to AWS hosts as read
from the GCP project. This operation is performed as the last step of creating
a new cluster but can be useful to re-run if the operation failed previously or
if the user would like to update the keys on the remote hosts.
`,

	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) (retErr error) {
		__antithesis_instrumentation__.Notify(42953)
		return roachprod.SetupSSH(context.Background(), roachprodLibraryLogger, args[0])
	}),
}

var destroyCmd = &cobra.Command{
	Use:   "destroy [ --all-mine | --all-local | <cluster 1> [<cluster 2> ...] ]",
	Short: "destroy clusters",
	Long: `Destroy one or more local or cloud-based clusters.

The destroy command accepts the names of the clusters to destroy. Alternatively,
the --all-mine flag can be provided to destroy all (non-local) clusters that are
owned by the current user, or the --all-local flag can be provided to destroy
all local clusters.

Destroying a cluster releases the resources for a cluster. For a cloud-based
cluster the machine and associated disk resources are freed. For a local
cluster, any processes started by roachprod are stopped, and the node
directories inside ${HOME}/local directory are removed.
`,
	Args: cobra.ArbitraryArgs,
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42954)
		return roachprod.Destroy(roachprodLibraryLogger, destroyAllMine, destroyAllLocal, args...)
	}),
}

var cachedHostsCmd = &cobra.Command{
	Use:   "cached-hosts",
	Short: "list all clusters (and optionally their host numbers) from local cache",
	Args:  cobra.NoArgs,
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42955)
		roachprod.CachedClusters(roachprodLibraryLogger, func(clusterName string, numVMs int) {
			__antithesis_instrumentation__.Notify(42957)
			if strings.HasPrefix(clusterName, "teamcity") {
				__antithesis_instrumentation__.Notify(42960)
				return
			} else {
				__antithesis_instrumentation__.Notify(42961)
			}
			__antithesis_instrumentation__.Notify(42958)
			fmt.Printf("%s", clusterName)

			if strings.HasPrefix(cachedHostsCluster, clusterName) {
				__antithesis_instrumentation__.Notify(42962)
				for i := 1; i <= numVMs; i++ {
					__antithesis_instrumentation__.Notify(42963)
					fmt.Printf(" %s:%d", clusterName, i)
				}
			} else {
				__antithesis_instrumentation__.Notify(42964)
			}
			__antithesis_instrumentation__.Notify(42959)
			fmt.Printf("\n")
		})
		__antithesis_instrumentation__.Notify(42956)
		return nil
	}),
}

var listCmd = &cobra.Command{
	Use:   "list [--details | --json] [ --mine | --pattern ]",
	Short: "list all clusters",
	Long: `List all clusters.

The list command accepts a flag --pattern which is a regular
expression that will be matched against the cluster name pattern.  Alternatively,
the --mine flag can be provided to list the clusters that are owned by the current
user.

The default output shows one line per cluster, including the local cluster if
it exists:

  ~ roachprod list
  local:     [local]    1  (-)
  marc-test: [aws gce]  4  (5h34m35s)
  Syncing...

The second column lists the cloud providers that host VMs for the cluster.

The third and fourth columns are the number of nodes in the cluster and the
time remaining before the cluster will be automatically destroyed. Note that
local clusters do not have an expiration.

The --details flag adjusts the output format to include per-node details:

  ~ roachprod list --details
  local [local]: (no expiration)
    localhost		127.0.0.1	127.0.0.1
  marc-test: [aws gce] 5h33m57s remaining
    marc-test-0001	marc-test-0001.us-east1-b.cockroach-ephemeral	10.142.0.18	35.229.60.91
    marc-test-0002	marc-test-0002.us-east1-b.cockroach-ephemeral	10.142.0.17	35.231.0.44
    marc-test-0003	marc-test-0003.us-east1-b.cockroach-ephemeral	10.142.0.19	35.229.111.100
    marc-test-0004	marc-test-0004.us-east1-b.cockroach-ephemeral	10.142.0.20	35.231.102.125
  Syncing...

The first and second column are the node hostname and fully qualified name
respectively. The third and fourth column are the private and public IP
addresses.

The --json flag sets the format of the command output to json.

Listing clusters has the side-effect of syncing ssh keys/configs and the local
hosts file.
`,
	Args: cobra.NoArgs,
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42965)
		if listJSON && func() bool {
			__antithesis_instrumentation__.Notify(42970)
			return listDetails == true
		}() == true {
			__antithesis_instrumentation__.Notify(42971)
			return errors.New("'json' option cannot be combined with 'details' option")
		} else {
			__antithesis_instrumentation__.Notify(42972)
		}
		__antithesis_instrumentation__.Notify(42966)
		filteredCloud, err := roachprod.List(roachprodLibraryLogger, listMine, listPattern)
		if err != nil {
			__antithesis_instrumentation__.Notify(42973)
			return err
		} else {
			__antithesis_instrumentation__.Notify(42974)
		}
		__antithesis_instrumentation__.Notify(42967)

		names := make([]string, len(filteredCloud.Clusters))
		i := 0
		for name := range filteredCloud.Clusters {
			__antithesis_instrumentation__.Notify(42975)
			names[i] = name
			i++
		}
		__antithesis_instrumentation__.Notify(42968)
		sort.Strings(names)

		if listJSON {
			__antithesis_instrumentation__.Notify(42976)
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			if err := enc.Encode(filteredCloud); err != nil {
				__antithesis_instrumentation__.Notify(42977)
				return err
			} else {
				__antithesis_instrumentation__.Notify(42978)
			}
		} else {
			__antithesis_instrumentation__.Notify(42979)

			tw := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
			for _, name := range names {
				__antithesis_instrumentation__.Notify(42982)
				c := filteredCloud.Clusters[name]
				if listDetails {
					__antithesis_instrumentation__.Notify(42983)
					c.PrintDetails(roachprodLibraryLogger)
				} else {
					__antithesis_instrumentation__.Notify(42984)
					fmt.Fprintf(tw, "%s\t%s\t%d", c.Name, c.Clouds(), len(c.VMs))
					if !c.IsLocal() {
						__antithesis_instrumentation__.Notify(42986)
						fmt.Fprintf(tw, "\t(%s)", c.LifetimeRemaining().Round(time.Second))
					} else {
						__antithesis_instrumentation__.Notify(42987)
						fmt.Fprintf(tw, "\t(-)")
					}
					__antithesis_instrumentation__.Notify(42985)
					fmt.Fprintf(tw, "\n")
				}
			}
			__antithesis_instrumentation__.Notify(42980)
			if err := tw.Flush(); err != nil {
				__antithesis_instrumentation__.Notify(42988)
				return err
			} else {
				__antithesis_instrumentation__.Notify(42989)
			}
			__antithesis_instrumentation__.Notify(42981)

			if listDetails {
				__antithesis_instrumentation__.Notify(42990)
				collated := filteredCloud.BadInstanceErrors()

				var errors ui.ErrorsByError
				for err := range collated {
					__antithesis_instrumentation__.Notify(42992)
					errors = append(errors, err)
				}
				__antithesis_instrumentation__.Notify(42991)
				sort.Sort(errors)

				for _, e := range errors {
					__antithesis_instrumentation__.Notify(42993)
					fmt.Printf("%s: %s\n", e, collated[e].Names())
				}
			} else {
				__antithesis_instrumentation__.Notify(42994)
			}
		}
		__antithesis_instrumentation__.Notify(42969)
		return nil
	}),
}

var bashCompletion = os.ExpandEnv("$HOME/.roachprod/bash-completion.sh")

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "sync ssh keys/config and hosts files",
	Long:  ``,
	Args:  cobra.NoArgs,
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42995)
		_, err := roachprod.Sync(roachprodLibraryLogger)
		_ = rootCmd.GenBashCompletionFile(bashCompletion)
		return err
	}),
}

var gcCmd = &cobra.Command{
	Use:   "gc",
	Short: "GC expired clusters and unused AWS keypairs\n",
	Long: `Garbage collect expired clusters and unused SSH keypairs in AWS.

Destroys expired clusters, sending email if properly configured. Usually run
hourly by a cronjob so it is not necessary to run manually.
`,
	Args: cobra.NoArgs,
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42996)
		return roachprod.GC(roachprodLibraryLogger, dryrun)
	}),
}

var extendCmd = &cobra.Command{
	Use:   "extend <cluster>",
	Short: "extend the lifetime of a cluster",
	Long: `Extend the lifetime of the specified cluster to prevent it from being
destroyed:

  roachprod extend marc-test --lifetime=6h
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42997)
		return roachprod.Extend(roachprodLibraryLogger, args[0], extendLifetime)
	}),
}

const tagHelp = `
The --tag flag can be used to to associate a tag with the process. This tag can
then be used to restrict the processes which are operated on by the status and
stop commands. Tags can have a hierarchical component by utilizing a slash
separated string similar to a filesystem path. A tag matches if a prefix of the
components match. For example, the tag "a/b" will match both "a/b" and
"a/b/c/d".
`

var startCmd = &cobra.Command{
	Use:   "start <cluster>",
	Short: "start nodes on a cluster",
	Long: `Start nodes on a cluster.

The --secure flag can be used to start nodes in secure mode (i.e. using
certs). When specified, there is a one time initialization for the cluster to
create and distribute the certs. Note that running some modes in secure mode
and others in insecure mode is not a supported Cockroach configuration.

As a debugging aid, the --sequential flag starts the nodes sequentially so node
IDs match hostnames. Otherwise nodes are started in parallel.

The --binary flag specifies the remote binary to run. It is up to the roachprod
user to ensure this binary exists, usually via "roachprod put". Note that no
cockroach software is installed by default on a newly created cluster.

The --args and --env flags can be used to pass arbitrary command line flags and
environment variables to the cockroach process.
` + tagHelp + `
The "start" command takes care of setting up the --join address and specifying
reasonable defaults for other flags. One side-effect of this convenience is
that node 1 is special and if started, is used to auto-initialize the cluster.
The --skip-init flag can be used to avoid auto-initialization (which can then
separately be done using the "init" command).

If the COCKROACH_DEV_LICENSE environment variable is set the enterprise.license
cluster setting will be set to its value.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42998)
		clusterSettingsOpts := []install.ClusterSettingOption{
			install.TagOption(tag),
			install.PGUrlCertsDirOption(pgurlCertsDir),
			install.SecureOption(secure),
			install.UseTreeDistOption(useTreeDist),
			install.EnvOption(nodeEnv),
			install.NumRacksOption(numRacks),
		}
		return roachprod.Start(context.Background(), roachprodLibraryLogger, args[0], startOpts, clusterSettingsOpts...)
	}),
}

var stopCmd = &cobra.Command{
	Use:   "stop <cluster> [--sig] [--wait]",
	Short: "stop nodes on a cluster",
	Long: `Stop nodes on a cluster.

Stop roachprod created processes running on the nodes in a cluster, including
processes started by the "start", "run" and "ssh" commands. Every process
started by roachprod is tagged with a ROACHPROD environment variable which is
used by "stop" to locate the processes and terminate them. By default processes
are killed with signal 9 (SIGKILL) giving them no chance for a graceful exit.

The --sig flag will pass a signal to kill to allow us finer control over how we
shutdown cockroach. The --wait flag causes stop to loop waiting for all
processes with the right ROACHPROD environment variable to exit. Note that stop
will wait forever if you specify --wait with a non-terminating signal (e.g.
SIGHUP). --wait defaults to true for signal 9 (SIGKILL) and false for all other
signals.
` + tagHelp + `
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(42999)
		wait := waitFlag
		if sig == 9 && func() bool {
			__antithesis_instrumentation__.Notify(43001)
			return !cmd.Flags().Changed("wait") == true
		}() == true {
			__antithesis_instrumentation__.Notify(43002)
			wait = true
		} else {
			__antithesis_instrumentation__.Notify(43003)
		}
		__antithesis_instrumentation__.Notify(43000)
		stopOpts := roachprod.StopOpts{Wait: wait, ProcessTag: tag, Sig: sig}
		return roachprod.Stop(context.Background(), roachprodLibraryLogger, args[0], stopOpts)
	}),
}

var startTenantCmd = &cobra.Command{
	Use:   "start-tenant <tenant-cluster> --host-cluster <host-cluster>",
	Short: "start a tenant",
	Long: `Start SQL instances for a non-system tenant.

The --host-cluster flag must be used to specify a host cluster (with optional
node selector) which is already running. The command will create the tenant on
the host cluster if it does not exist already. The host and tenant can use the
same underlying cluster, as long as different subsets of nodes are selected.

The --tenant-id flag can be used to specify the tenant ID; it defaults to 2.

The --secure flag can be used to start nodes in secure mode (i.e. using
certs). When specified, there is a one time initialization for the cluster to
create and distribute the certs. Note that running some modes in secure mode
and others in insecure mode is not a supported Cockroach configuration.

As a debugging aid, the --sequential flag starts the nodes sequentially so node
IDs match hostnames. Otherwise nodes are started in parallel.

The --binary flag specifies the remote binary to run. It is up to the roachprod
user to ensure this binary exists, usually via "roachprod put". Note that no
cockroach software is installed by default on a newly created cluster.

The --args and --env flags can be used to pass arbitrary command line flags and
environment variables to the cockroach process.
` + tagHelp + `
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43004)
		tenantCluster := args[0]
		clusterSettingsOpts := []install.ClusterSettingOption{
			install.TagOption(tag),
			install.PGUrlCertsDirOption(pgurlCertsDir),
			install.SecureOption(secure),
			install.UseTreeDistOption(useTreeDist),
			install.EnvOption(nodeEnv),
			install.NumRacksOption(numRacks),
		}
		return roachprod.StartTenant(context.Background(), roachprodLibraryLogger, tenantCluster, hostCluster, startOpts, clusterSettingsOpts...)
	}),
}

var initCmd = &cobra.Command{
	Use:   "init <cluster>",
	Short: "initialize the cluster",
	Long: `Initialize the cluster.

The "init" command bootstraps the cluster (using "cockroach init"). It also sets
default cluster settings. It's intended to be used in conjunction with
'roachprod start --skip-init'.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43005)
		return roachprod.Init(context.Background(), roachprodLibraryLogger, args[0])
	}),
}

var statusCmd = &cobra.Command{
	Use:   "status <cluster>",
	Short: "retrieve the status of nodes in a cluster",
	Long: `Retrieve the status of nodes in a cluster.

The "status" command outputs the binary and PID for the specified nodes:

  ~ roachprod status local
  local: status 3/3
     1: cockroach 29688
     2: cockroach 29687
     3: cockroach 29689
` + tagHelp + `
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43006)
		return roachprod.Status(context.Background(), roachprodLibraryLogger, args[0], tag)
	}),
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "retrieve and merge logs in a cluster",
	Long: `Retrieve and merge logs in a cluster.

The "logs" command runs until terminated. It works similarly to get but is
specifically focused on retrieving logs periodically and then merging them
into a single stream.
`,
	Args: cobra.RangeArgs(1, 2),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43007)
		logsOpts := roachprod.LogsOpts{
			Dir: logsDir, Filter: logsFilter, ProgramFilter: logsProgramFilter,
			Interval: logsInterval, From: logsFrom, To: logsTo, Out: cmd.OutOrStdout(),
		}
		var dest string
		if len(args) == 2 {
			__antithesis_instrumentation__.Notify(43009)
			dest = args[1]
		} else {
			__antithesis_instrumentation__.Notify(43010)
			dest = args[0] + ".logs"
		}
		__antithesis_instrumentation__.Notify(43008)
		return roachprod.Logs(roachprodLibraryLogger, args[0], dest, username, logsOpts)
	}),
}

var monitorCmd = &cobra.Command{
	Use:   "monitor",
	Short: "monitor the status of nodes in a cluster",
	Long: `Monitor the status of cockroach nodes in a cluster.

The "monitor" command runs until terminated. At startup it outputs a line for
each specified node indicating the status of the node (either the PID of the
node if alive, or "dead" otherwise). It then watches for changes in the status
of nodes, outputting a line whenever a change is detected:

  ~ roachprod monitor local
  1: 29688
  3: 29689
  2: 29687
  3: dead
  3: 30718
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43011)
		messages, err := roachprod.Monitor(context.Background(), roachprodLibraryLogger, args[0], monitorOpts)
		if err != nil {
			__antithesis_instrumentation__.Notify(43014)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43015)
		}
		__antithesis_instrumentation__.Notify(43012)
		for msg := range messages {
			__antithesis_instrumentation__.Notify(43016)
			if msg.Err != nil {
				__antithesis_instrumentation__.Notify(43019)
				msg.Msg += "error: " + msg.Err.Error()
			} else {
				__antithesis_instrumentation__.Notify(43020)
			}
			__antithesis_instrumentation__.Notify(43017)
			thisError := errors.Newf("%d: %s", msg.Node, msg.Msg)
			if msg.Err != nil || func() bool {
				__antithesis_instrumentation__.Notify(43021)
				return strings.Contains(msg.Msg, "dead") == true
			}() == true {
				__antithesis_instrumentation__.Notify(43022)
				err = errors.CombineErrors(err, thisError)
			} else {
				__antithesis_instrumentation__.Notify(43023)
			}
			__antithesis_instrumentation__.Notify(43018)
			fmt.Println(thisError.Error())
		}
		__antithesis_instrumentation__.Notify(43013)
		return err
	}),
}

var wipeCmd = &cobra.Command{
	Use:   "wipe <cluster>",
	Short: "wipe a cluster",
	Long: `Wipe the nodes in a cluster.

The "wipe" command first stops any processes running on the nodes in a cluster
(via the "stop" command) and then deletes the data directories used by the
nodes.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43024)
		return roachprod.Wipe(context.Background(), roachprodLibraryLogger, args[0], wipePreserveCerts)
	}),
}

var reformatCmd = &cobra.Command{
	Use:   "reformat <cluster> <filesystem>",
	Short: "reformat disks in a cluster\n",
	Long: `
Reformat disks in a cluster to use the specified filesystem.

WARNING: Reformatting will delete all existing data in the cluster.

Filesystem options:
  ext4
  zfs

When running with ZFS, you can create a snapshot of the filesystem's current
state using the 'zfs snapshot' command:

  $ roachprod run <cluster> 'sudo zfs snapshot data1@pristine'

You can then nearly instantaneously restore the filesystem to this state with
the 'zfs rollback' command:

  $ roachprod run <cluster> 'sudo zfs rollback data1@pristine'

`,

	Args: cobra.ExactArgs(2),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43025)
		return roachprod.Reformat(context.Background(), roachprodLibraryLogger, args[0], args[1])
	}),
}

var runCmd = &cobra.Command{
	Use:     "run <cluster> <command> [args]",
	Aliases: []string{"ssh"},
	Short:   "run a command on the nodes in a cluster",
	Long: `Run a command on the nodes in a cluster.
`,
	Args: cobra.MinimumNArgs(1),
	Run: wrap(func(_ *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43026)
		return roachprod.Run(context.Background(), roachprodLibraryLogger, args[0], extraSSHOptions, tag, secure, os.Stdout, os.Stderr, args[1:])
	}),
}

var resetCmd = &cobra.Command{
	Use:   "reset <cluster>",
	Short: "reset *all* VMs in a cluster",
	Long: `Reset a cloud VM. This may not be implemented for all
environments and will fall back to a no-op.`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) (retErr error) {
		__antithesis_instrumentation__.Notify(43027)
		return roachprod.Reset(roachprodLibraryLogger, args[0])
	}),
}

var installCmd = &cobra.Command{
	Use:   "install <cluster> <software>",
	Short: "install 3rd party software",
	Long: `Install third party software. Currently available installation options are:

    ` + strings.Join(install.SortedCmds(), "\n    ") + `
`,
	Args: cobra.MinimumNArgs(2),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43028)
		return roachprod.Install(context.Background(), roachprodLibraryLogger, args[0], args[1:])
	}),
}

var downloadCmd = &cobra.Command{
	Use:   "download <cluster> <url> <sha256> [DESTINATION]",
	Short: "download 3rd party tools",
	Long:  "Downloads 3rd party tools, using a GCS cache if possible.",
	Args:  cobra.RangeArgs(3, 4),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43029)
		src, sha := args[1], args[2]
		var dest string
		if len(args) == 4 {
			__antithesis_instrumentation__.Notify(43031)
			dest = args[3]
		} else {
			__antithesis_instrumentation__.Notify(43032)
		}
		__antithesis_instrumentation__.Notify(43030)
		return roachprod.Download(context.Background(), roachprodLibraryLogger, args[0], src, sha, dest)
	}),
}

var stageURLCmd = &cobra.Command{
	Use:   "stageurl <application> [<sha/version>]",
	Short: "print URL to cockroach binaries",
	Long: `Prints URL for release and edge binaries.

Currently available application options are:
  cockroach - Cockroach Unofficial. Can provide an optional SHA, otherwise
              latest build version is used.
  workload  - Cockroach workload application.
  release   - Official CockroachDB Release. Must provide a specific release
              version.
`,
	Args: cobra.RangeArgs(1, 2),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43033)
		versionArg := ""
		if len(args) == 2 {
			__antithesis_instrumentation__.Notify(43037)
			versionArg = args[1]
		} else {
			__antithesis_instrumentation__.Notify(43038)
		}
		__antithesis_instrumentation__.Notify(43034)
		urls, err := roachprod.StageURL(roachprodLibraryLogger, args[0], versionArg, stageOS)
		if err != nil {
			__antithesis_instrumentation__.Notify(43039)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43040)
		}
		__antithesis_instrumentation__.Notify(43035)
		for _, u := range urls {
			__antithesis_instrumentation__.Notify(43041)
			fmt.Println(u)
		}
		__antithesis_instrumentation__.Notify(43036)
		return nil
	}),
}

var stageCmd = &cobra.Command{
	Use:   "stage <cluster> <application> [<sha/version>]",
	Short: "stage cockroach binaries",
	Long: `Stages release and edge binaries to the cluster.

Currently available application options are:
  cockroach - Cockroach Unofficial. Can provide an optional SHA, otherwise
              latest build version is used.
  workload  - Cockroach workload application.
  release   - Official CockroachDB Release. Must provide a specific release
              version.

Some examples of usage:
  -- stage edge build of cockroach build at a specific SHA:
  roachprod stage my-cluster cockroach e90e6903fee7dd0f88e20e345c2ddfe1af1e5a97

  -- Stage the most recent edge build of the workload tool:
  roachprod stage my-cluster workload

  -- Stage the official release binary of CockroachDB at version 2.0.5
  roachprod stage my-cluster release v2.0.5
`,
	Args: cobra.RangeArgs(2, 3),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43042)
		versionArg := ""
		if len(args) == 3 {
			__antithesis_instrumentation__.Notify(43044)
			versionArg = args[2]
		} else {
			__antithesis_instrumentation__.Notify(43045)
		}
		__antithesis_instrumentation__.Notify(43043)
		return roachprod.Stage(context.Background(), roachprodLibraryLogger, args[0], stageOS, stageDir, args[1], versionArg)
	}),
}

var distributeCertsCmd = &cobra.Command{
	Use:   "distribute-certs <cluster>",
	Short: "distribute certificates to the nodes in a cluster",
	Long: `Distribute certificates to the nodes in a cluster.
If the certificates already exist, no action is taken. Note that this command is
invoked automatically when a secure cluster is bootstrapped by "roachprod
start."
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43046)
		return roachprod.DistributeCerts(context.Background(), roachprodLibraryLogger, args[0])
	}),
}

var putCmd = &cobra.Command{
	Use:   "put <cluster> <src> [<dest>]",
	Short: "copy a local file to the nodes in a cluster",
	Long: `Copy a local file to the nodes in a cluster.
`,
	Args: cobra.RangeArgs(2, 3),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43047)
		src := args[1]
		dest := path.Base(src)
		if len(args) == 3 {
			__antithesis_instrumentation__.Notify(43049)
			dest = args[2]
		} else {
			__antithesis_instrumentation__.Notify(43050)
		}
		__antithesis_instrumentation__.Notify(43048)
		return roachprod.Put(context.Background(), roachprodLibraryLogger, args[0], src, dest, useTreeDist)
	}),
}

var getCmd = &cobra.Command{
	Use:   "get <cluster> <src> [<dest>]",
	Short: "copy a remote file from the nodes in a cluster",
	Long: `Copy a remote file from the nodes in a cluster. If the file is retrieved from
multiple nodes the destination file name will be prefixed with the node number.
`,
	Args: cobra.RangeArgs(2, 3),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43051)
		src := args[1]
		dest := path.Base(src)
		if len(args) == 3 {
			__antithesis_instrumentation__.Notify(43053)
			dest = args[2]
		} else {
			__antithesis_instrumentation__.Notify(43054)
		}
		__antithesis_instrumentation__.Notify(43052)
		return roachprod.Get(roachprodLibraryLogger, args[0], src, dest)
	}),
}

var sqlCmd = &cobra.Command{
	Use:   "sql <cluster> -- [args]",
	Short: "run `cockroach sql` on a remote cluster",
	Long:  "Run `cockroach sql` on a remote cluster.\n",
	Args:  cobra.MinimumNArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43055)
		return roachprod.SQL(context.Background(), roachprodLibraryLogger, args[0], secure, args[1:])
	}),
}

var pgurlCmd = &cobra.Command{
	Use:   "pgurl <cluster>",
	Short: "generate pgurls for the nodes in a cluster",
	Long: `Generate pgurls for the nodes in a cluster.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43056)
		urls, err := roachprod.PgURL(context.Background(), roachprodLibraryLogger, args[0], pgurlCertsDir, external, secure)
		if err != nil {
			__antithesis_instrumentation__.Notify(43058)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43059)
		}
		__antithesis_instrumentation__.Notify(43057)
		fmt.Println(strings.Join(urls, " "))
		return nil
	}),
}

var pprofCmd = &cobra.Command{
	Use:     "pprof <cluster>",
	Args:    cobra.ExactArgs(1),
	Aliases: []string{"pprof-heap"},
	Short:   "capture a pprof profile from the specified nodes",
	Long: `Capture a pprof profile from the specified nodes.

Examples:

    # Capture CPU profile for all nodes in the cluster
    roachprod pprof CLUSTERNAME
    # Capture CPU profile for the first node in the cluster for 60 seconds
    roachprod pprof CLUSTERNAME:1 --duration 60s
    # Capture a Heap profile for the first node in the cluster
    roachprod pprof CLUSTERNAME:1 --heap
    # Same as above
    roachprod pprof-heap CLUSTERNAME:1
`,
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43060)
		if cmd.CalledAs() == "pprof-heap" {
			__antithesis_instrumentation__.Notify(43062)
			pprofOpts.Heap = true
		} else {
			__antithesis_instrumentation__.Notify(43063)
		}
		__antithesis_instrumentation__.Notify(43061)
		return roachprod.Pprof(roachprodLibraryLogger, args[0], pprofOpts)
	}),
}

var adminurlCmd = &cobra.Command{
	Use:     "adminurl <cluster>",
	Aliases: []string{"admin", "adminui"},
	Short:   "generate admin UI URLs for the nodes in a cluster\n",
	Long: `Generate admin UI URLs for the nodes in a cluster.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43064)
		urls, err := roachprod.AdminURL(roachprodLibraryLogger, args[0], adminurlPath, adminurlIPs, adminurlOpen, secure)
		if err != nil {
			__antithesis_instrumentation__.Notify(43067)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43068)
		}
		__antithesis_instrumentation__.Notify(43065)
		for _, url := range urls {
			__antithesis_instrumentation__.Notify(43069)
			fmt.Println(url)
		}
		__antithesis_instrumentation__.Notify(43066)
		return nil
	}),
}

var ipCmd = &cobra.Command{
	Use:   "ip <cluster>",
	Short: "get the IP addresses of the nodes in a cluster",
	Long: `Get the IP addresses of the nodes in a cluster.
`,
	Args: cobra.ExactArgs(1),
	Run: wrap(func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43070)
		ips, err := roachprod.IP(context.Background(), roachprodLibraryLogger, args[0], external)
		if err != nil {
			__antithesis_instrumentation__.Notify(43073)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43074)
		}
		__antithesis_instrumentation__.Notify(43071)
		for _, ip := range ips {
			__antithesis_instrumentation__.Notify(43075)
			fmt.Println(ip)
		}
		__antithesis_instrumentation__.Notify(43072)
		return nil
	}),
}

var versionCmd = &cobra.Command{
	Use:   `version`,
	Short: `print version information`,
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43076)
		fmt.Println(roachprod.Version(roachprodLibraryLogger))
		return nil
	},
}

var getProvidersCmd = &cobra.Command{
	Use:   `get-providers`,
	Short: `print providers state (active/inactive)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(43077)
		providers := roachprod.InitProviders()
		for provider, state := range providers {
			__antithesis_instrumentation__.Notify(43079)
			fmt.Printf("%s: %s\n", provider, state)
		}
		__antithesis_instrumentation__.Notify(43078)
		return nil
	},
}

func main() {
	__antithesis_instrumentation__.Notify(43080)
	loggerCfg := logger.Config{Stdout: os.Stdout, Stderr: os.Stderr}
	var loggerError error
	roachprodLibraryLogger, loggerError = loggerCfg.NewLogger("")
	if loggerError != nil {
		__antithesis_instrumentation__.Notify(43086)
		fmt.Fprintf(os.Stderr, "unable to configure logger: %s\n", loggerError)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(43087)
	}
	__antithesis_instrumentation__.Notify(43081)

	_ = roachprod.InitProviders()
	providerOptsContainer = vm.CreateProviderOptionsContainer()

	cobra.EnableCommandSorting = false
	rootCmd.AddCommand(
		createCmd,
		resetCmd,
		destroyCmd,
		extendCmd,
		listCmd,
		syncCmd,
		gcCmd,
		setupSSHCmd,

		statusCmd,
		monitorCmd,
		startCmd,
		stopCmd,
		startTenantCmd,
		initCmd,
		runCmd,
		wipeCmd,
		reformatCmd,
		installCmd,
		distributeCertsCmd,
		putCmd,
		getCmd,
		stageCmd,
		stageURLCmd,
		downloadCmd,
		sqlCmd,
		ipCmd,
		pgurlCmd,
		adminurlCmd,
		logsCmd,
		pprofCmd,
		cachedHostsCmd,
		versionCmd,
		getProvidersCmd,
	)
	setBashCompletionFunction()

	for _, cmd := range []*cobra.Command{
		getCmd, putCmd, runCmd, startCmd, statusCmd, stopCmd,
		wipeCmd, pgurlCmd, adminurlCmd, sqlCmd, installCmd,
	} {
		__antithesis_instrumentation__.Notify(43088)
		if cmd.Long == "" {
			__antithesis_instrumentation__.Notify(43090)
			cmd.Long = cmd.Short
		} else {
			__antithesis_instrumentation__.Notify(43091)
		}
		__antithesis_instrumentation__.Notify(43089)
		cmd.Long += fmt.Sprintf(`
Node specification

  By default the operation is performed on all nodes in <cluster>. A subset of
  nodes can be specified by appending :<nodes> to the cluster name. The syntax
  of <nodes> is a comma separated list of specific node IDs or range of
  IDs. For example:

    roachprod %[1]s marc-test:1-3,8-9

  will perform %[1]s on:

    marc-test-1
    marc-test-2
    marc-test-3
    marc-test-8
    marc-test-9
`, cmd.Name())
	}
	__antithesis_instrumentation__.Notify(43082)

	initFlags()

	var err error
	config.OSUser, err = user.Current()
	if err != nil {
		__antithesis_instrumentation__.Notify(43092)
		fmt.Fprintf(os.Stderr, "unable to lookup current user: %s\n", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(43093)
	}
	__antithesis_instrumentation__.Notify(43083)

	if err := roachprod.InitDirs(); err != nil {
		__antithesis_instrumentation__.Notify(43094)
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(43095)
	}
	__antithesis_instrumentation__.Notify(43084)

	if err := roachprod.LoadClusters(); err != nil {
		__antithesis_instrumentation__.Notify(43096)

		fmt.Printf("problem loading clusters: %s\n", err)
	} else {
		__antithesis_instrumentation__.Notify(43097)
	}
	__antithesis_instrumentation__.Notify(43085)

	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(43098)

		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(43099)
	}
}
