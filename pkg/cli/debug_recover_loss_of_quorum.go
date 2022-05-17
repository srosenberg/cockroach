package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery/loqrecoverypb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/hintdetail"
	"github.com/spf13/cobra"
)

type confirmActionFlag int

const (
	prompt confirmActionFlag = iota
	allNo
	allYes
)

func (l *confirmActionFlag) Type() string {
	__antithesis_instrumentation__.Notify(31477)
	return "confirmAction"
}

func (l *confirmActionFlag) String() string {
	__antithesis_instrumentation__.Notify(31478)
	switch *l {
	case allYes:
		__antithesis_instrumentation__.Notify(31480)
		return "y"
	case allNo:
		__antithesis_instrumentation__.Notify(31481)
		return "n"
	case prompt:
		__antithesis_instrumentation__.Notify(31482)
		return "p"
	default:
		__antithesis_instrumentation__.Notify(31483)
	}
	__antithesis_instrumentation__.Notify(31479)
	log.Fatalf(context.Background(), "unknown confirm action flag value %d", *l)
	return ""
}

func (l *confirmActionFlag) Set(value string) error {
	__antithesis_instrumentation__.Notify(31484)
	switch strings.ToLower(value) {
	case "y", "yes":
		__antithesis_instrumentation__.Notify(31486)
		*l = allYes
	case "n", "no":
		__antithesis_instrumentation__.Notify(31487)
		*l = allNo
	case "p", "ask":
		__antithesis_instrumentation__.Notify(31488)
		*l = prompt
	default:
		__antithesis_instrumentation__.Notify(31489)
		return errors.Errorf("unrecognized value for confirmation flag: %s", value)
	}
	__antithesis_instrumentation__.Notify(31485)
	return nil
}

var debugRecoverCmd = &cobra.Command{
	Use:   "recover [command]",
	Short: "commands to recover unavailable ranges in case of quorum loss",
	Long: `Set of commands to recover unavailable ranges.

If cluster permanently loses several nodes containing multiple replicas of
the range it is possible for that range to lose quorum. Those ranges can
not converge on their final state and can't proceed further. This is a 
potential data and consistency loss situation. In most cases cluster should 
be restored from the latest backup.

As a workaround, when consistency could be sacrificed temporarily to restore
availability quickly it is possible to instruct remaining replicas to act as
if they have the most up to date view of data and continue operation. There's
no way to guarantee which replica contains most up-to-date information if at
all so heuristics based on raft state are used to pick one.

As a result of such recovery, some of the data could be lost, indexes could
become corrupted and database constraints could be violated. So manual
recovery steps to verify the data and ensure database consistency should be
taken ASAP. Those actions should be done at application level.

'debug recover' set of commands is used as a last resort to perform range
recovery operation. To perform recovery one should perform this sequence
of actions:

0. Decommission failed nodes preemptively to eliminate the possibility of
them coming back online and conflicting with the recovered state. Note that
if system ranges suffer loss of quorum, it may be impossible to decommission
nodes. In that case, recovery can proceed, but those nodes must be prevented
from communicating with the cluster and must be decommissioned once the cluster
is back online after recovery.

1. Stop the cluster

2. Run 'cockroach debug recover collect-info' on every node to collect
replication state from all surviving nodes. Outputs of these invocations
should be collected and made locally available for the next step.

3. Run 'cockroach debug recover make-plan' providing all files generated
on step 1. Planner will decide which replicas should survive and
up-replicate.

4. Run 'cockroach debug recover execute-plan' on every node using plan
generated on the previous step. Each node will pick relevant portion of
the plan and update local replicas accordingly to restore quorum.

5. Start the cluster.

If it was possible to produce and apply the plan, then cluster should
become operational again. It is not guaranteed that there's no data loss
and that all database consistency was not compromised.

Example run:

If we have a cluster of 5 nodes 1-5 where we lost nodes 3 and 4. Each node
has two stores and they are numbered as 1,2 on node 1; 3,4 on node 2 etc.
Recovery commands to recover unavailable ranges would be:

Decommission dead nodes and stop the cluster.

[cockroach@node1 ~]$ cockroach debug recover collect-info --store=/mnt/cockroach-data-1 --store=/mnt/cockroach-data-2 >info-node1.json
[cockroach@node2 ~]$ cockroach debug recover collect-info --store=/mnt/cockroach-data-1 --store=/mnt/cockroach-data-2 >info-node2.json
[cockroach@node5 ~]$ cockroach debug recover collect-info --store=/mnt/cockroach-data-1 --store=/mnt/cockroach-data-2 >info-node5.json

[cockroach@base ~]$ scp cockroach@node1:info-node1.json .
[cockroach@base ~]$ scp cockroach@node2:info-node1.json .
[cockroach@base ~]$ scp cockroach@node5:info-node1.json .

[cockroach@base ~]$ cockroach debug recover make-plan --dead-store-ids=5,6,7,8 info-node1.json info-node2.json info-node5.json >recover-plan.json

[cockroach@base ~]$ scp recover-plan.json cockroach@node1:
[cockroach@base ~]$ scp recover-plan.json cockroach@node2:
[cockroach@base ~]$ scp recover-plan.json cockroach@node5:

[cockroach@node1 ~]$ cockroach debug recover apply-plan --store=/mnt/cockroach-data-1 --store=/mnt/cockroach-data-2 recover-plan.json
Info:
  updating replica for r1
Proceed with above changes [y/N]y
Info:
  updated store s2
[cockroach@node2 ~]$ cockroach debug recover apply-plan --store=/mnt/cockroach-data-1 --store=/mnt/cockroach-data-2 recover-plan.json
[cockroach@node5 ~]$ cockroach debug recover apply-plan --store=/mnt/cockroach-data-1 --store=/mnt/cockroach-data-2 recover-plan.json

Now the cluster could be started again.
`,
	RunE: UsageAndErr,
}

func init() {
	debugRecoverCmd.AddCommand(
		debugRecoverCollectInfoCmd,
		debugRecoverPlanCmd,
		debugRecoverExecuteCmd)
}

var debugRecoverCollectInfoCmd = &cobra.Command{
	Use:   "collect-info [destination-file]",
	Short: "collect replica information from the given stores",
	Long: `
Collect information about replicas by reading data from underlying stores. Store
locations must be provided using --store flags.

Collected information is written to a destination file if file name is provided,
or to stdout.

Multiple store locations could be provided to the command to collect all info from
node at once. It is also possible to call it per store, in that case all resulting
files should be fed to plan subcommand.

See debug recover command help for more details on how to use this command.
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDebugDeadReplicaCollect,
}

var debugRecoverCollectInfoOpts struct {
	Stores base.StoreSpecList
}

func runDebugDeadReplicaCollect(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31490)
	stopper := stop.NewStopper()
	defer stopper.Stop(cmd.Context())

	var stores []storage.Engine
	for _, storeSpec := range debugRecoverCollectInfoOpts.Stores.Specs {
		__antithesis_instrumentation__.Notify(31496)
		db, err := OpenExistingStore(storeSpec.Path, stopper, true, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(31498)
			return errors.Wrapf(err, "failed to open store at path %q, ensure that store path is "+
				"correct and that it is not used by another process", storeSpec.Path)
		} else {
			__antithesis_instrumentation__.Notify(31499)
		}
		__antithesis_instrumentation__.Notify(31497)
		stores = append(stores, db)
	}
	__antithesis_instrumentation__.Notify(31491)

	replicaInfo, err := loqrecovery.CollectReplicaInfo(cmd.Context(), stores)
	if err != nil {
		__antithesis_instrumentation__.Notify(31500)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31501)
	}
	__antithesis_instrumentation__.Notify(31492)

	var writer io.Writer = os.Stdout
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(31502)
		filename := args[0]
		if _, err = os.Stat(filename); err == nil {
			__antithesis_instrumentation__.Notify(31505)
			return errors.Newf("file %q already exists", filename)
		} else {
			__antithesis_instrumentation__.Notify(31506)
		}
		__antithesis_instrumentation__.Notify(31503)

		outFile, err := os.Create(filename)
		if err != nil {
			__antithesis_instrumentation__.Notify(31507)
			return errors.Wrapf(err, "failed to create file %q", filename)
		} else {
			__antithesis_instrumentation__.Notify(31508)
		}
		__antithesis_instrumentation__.Notify(31504)
		defer outFile.Close()
		writer = outFile
	} else {
		__antithesis_instrumentation__.Notify(31509)
	}
	__antithesis_instrumentation__.Notify(31493)
	jsonpb := protoutil.JSONPb{Indent: "  "}
	var out []byte
	if out, err = jsonpb.Marshal(replicaInfo); err != nil {
		__antithesis_instrumentation__.Notify(31510)
		return errors.Wrap(err, "failed to marshal collected replica info")
	} else {
		__antithesis_instrumentation__.Notify(31511)
	}
	__antithesis_instrumentation__.Notify(31494)
	if _, err = writer.Write(out); err != nil {
		__antithesis_instrumentation__.Notify(31512)
		return errors.Wrap(err, "failed to write collected replica info")
	} else {
		__antithesis_instrumentation__.Notify(31513)
	}
	__antithesis_instrumentation__.Notify(31495)
	_, _ = fmt.Fprintf(stderr, "Collected info about %d replicas.\n", len(replicaInfo.Replicas))
	return nil
}

var debugRecoverPlanCmd = &cobra.Command{
	Use:   "make-plan [replica-files]",
	Short: "generate a plan to recover ranges that lost quorum",
	Long: `
Devise a plan to restore ranges that lost a quorum.

This command will read files with information about replicas collected from all
surviving nodes of a cluster and make a decision which replicas should be survivors
for the ranges where quorum was lost.
Decision is then written into a file or stdout.

This command only creates a plan and doesn't change any data.'

See debug recover command help for more details on how to use this command.
`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDebugPlanReplicaRemoval,
}

var debugRecoverPlanOpts struct {
	outputFileName string
	deadStoreIDs   []int
	confirmAction  confirmActionFlag
	force          bool
}

func runDebugPlanReplicaRemoval(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31514)
	replicas, err := readReplicaInfoData(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(31528)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31529)
	}
	__antithesis_instrumentation__.Notify(31515)

	var deadStoreIDs []roachpb.StoreID
	for _, id := range debugRecoverPlanOpts.deadStoreIDs {
		__antithesis_instrumentation__.Notify(31530)
		deadStoreIDs = append(deadStoreIDs, roachpb.StoreID(id))
	}
	__antithesis_instrumentation__.Notify(31516)

	plan, report, err := loqrecovery.PlanReplicas(cmd.Context(), replicas, deadStoreIDs)
	if err != nil {
		__antithesis_instrumentation__.Notify(31531)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31532)
	}
	__antithesis_instrumentation__.Notify(31517)

	_, _ = fmt.Fprintf(stderr, `Total replicas analyzed: %d
Ranges without quorum:   %d
Discarded live replicas: %d

`, report.TotalReplicas, len(report.PlannedUpdates), report.DiscardedNonSurvivors)
	for _, r := range report.PlannedUpdates {
		__antithesis_instrumentation__.Notify(31533)
		_, _ = fmt.Fprintf(stderr, "Recovering range r%d:%s updating replica %s to %s. "+
			"Discarding available replicas: [%s], discarding dead replicas: [%s].\n",
			r.RangeID, r.StartKey, r.OldReplica, r.Replica,
			r.DiscardedAvailableReplicas, r.DiscardedDeadReplicas)
	}
	__antithesis_instrumentation__.Notify(31518)

	deadStoreMsg := fmt.Sprintf("\nDiscovered dead stores from provided files: %s",
		joinStoreIDs(report.MissingStores))
	if len(deadStoreIDs) > 0 {
		__antithesis_instrumentation__.Notify(31534)
		_, _ = fmt.Fprintf(stderr, "%s, (matches --dead-store-ids)\n\n", deadStoreMsg)
	} else {
		__antithesis_instrumentation__.Notify(31535)
		_, _ = fmt.Fprintf(stderr, "%s\n\n", deadStoreMsg)
	}
	__antithesis_instrumentation__.Notify(31519)

	planningErr := report.Error()
	if planningErr != nil {
		__antithesis_instrumentation__.Notify(31536)

		_, _ = fmt.Fprintf(stderr,
			"Found replica inconsistencies:\n\n%s\n\nOnly proceed as a last resort!\n",
			hintdetail.FlattenDetails(planningErr))
	} else {
		__antithesis_instrumentation__.Notify(31537)
	}
	__antithesis_instrumentation__.Notify(31520)

	if debugRecoverPlanOpts.confirmAction == allNo {
		__antithesis_instrumentation__.Notify(31538)
		return errors.New("abort")
	} else {
		__antithesis_instrumentation__.Notify(31539)
	}
	__antithesis_instrumentation__.Notify(31521)

	switch debugRecoverPlanOpts.confirmAction {
	case prompt:
		__antithesis_instrumentation__.Notify(31540)
		opts := "y/N"
		if planningErr != nil {
			__antithesis_instrumentation__.Notify(31544)
			opts = "f/N"
		} else {
			__antithesis_instrumentation__.Notify(31545)
		}
		__antithesis_instrumentation__.Notify(31541)
		done := false
		for !done {
			__antithesis_instrumentation__.Notify(31546)
			_, _ = fmt.Fprintf(stderr, "Proceed with plan creation [%s] ", opts)
			reader := bufio.NewReader(os.Stdin)
			line, err := reader.ReadString('\n')
			if err != nil {
				__antithesis_instrumentation__.Notify(31549)
				return errors.Wrap(err, "failed to read user input")
			} else {
				__antithesis_instrumentation__.Notify(31550)
			}
			__antithesis_instrumentation__.Notify(31547)
			line = strings.ToLower(strings.TrimSpace(line))
			if len(line) == 0 {
				__antithesis_instrumentation__.Notify(31551)
				line = "n"
			} else {
				__antithesis_instrumentation__.Notify(31552)
			}
			__antithesis_instrumentation__.Notify(31548)
			switch line {
			case "y":
				__antithesis_instrumentation__.Notify(31553)

				if planningErr != nil {
					__antithesis_instrumentation__.Notify(31558)
					continue
				} else {
					__antithesis_instrumentation__.Notify(31559)
				}
				__antithesis_instrumentation__.Notify(31554)
				done = true
			case "f":
				__antithesis_instrumentation__.Notify(31555)
				done = true
			case "n":
				__antithesis_instrumentation__.Notify(31556)
				return errors.New("abort")
			default:
				__antithesis_instrumentation__.Notify(31557)
			}
		}
	case allYes:
		__antithesis_instrumentation__.Notify(31542)
		if planningErr != nil && func() bool {
			__antithesis_instrumentation__.Notify(31560)
			return !debugRecoverPlanOpts.force == true
		}() == true {
			__antithesis_instrumentation__.Notify(31561)
			return errors.Errorf(
				"can not create plan because of errors and no --force flag is given")
		} else {
			__antithesis_instrumentation__.Notify(31562)
		}
	default:
		__antithesis_instrumentation__.Notify(31543)
		return errors.New("unexpected CLI error, try using different --confirm option value")
	}
	__antithesis_instrumentation__.Notify(31522)

	if len(plan.Updates) == 0 {
		__antithesis_instrumentation__.Notify(31563)
		_, _ = fmt.Fprintln(stderr, "Found no ranges in need of recovery, nothing to do.")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(31564)
	}
	__antithesis_instrumentation__.Notify(31523)

	var writer io.Writer = os.Stdout
	if len(debugRecoverPlanOpts.outputFileName) > 0 {
		__antithesis_instrumentation__.Notify(31565)
		if _, err = os.Stat(debugRecoverPlanOpts.outputFileName); err == nil {
			__antithesis_instrumentation__.Notify(31568)
			return errors.Newf("file %q already exists", debugRecoverPlanOpts.outputFileName)
		} else {
			__antithesis_instrumentation__.Notify(31569)
		}
		__antithesis_instrumentation__.Notify(31566)
		outFile, err := os.Create(debugRecoverPlanOpts.outputFileName)
		if err != nil {
			__antithesis_instrumentation__.Notify(31570)
			return errors.Wrapf(err, "failed to create file %q", debugRecoverPlanOpts.outputFileName)
		} else {
			__antithesis_instrumentation__.Notify(31571)
		}
		__antithesis_instrumentation__.Notify(31567)
		defer outFile.Close()
		writer = outFile
	} else {
		__antithesis_instrumentation__.Notify(31572)
	}
	__antithesis_instrumentation__.Notify(31524)

	jsonpb := protoutil.JSONPb{Indent: "  "}
	var out []byte
	if out, err = jsonpb.Marshal(plan); err != nil {
		__antithesis_instrumentation__.Notify(31573)
		return errors.Wrap(err, "failed to marshal recovery plan")
	} else {
		__antithesis_instrumentation__.Notify(31574)
	}
	__antithesis_instrumentation__.Notify(31525)
	if _, err = writer.Write(out); err != nil {
		__antithesis_instrumentation__.Notify(31575)
		return errors.Wrap(err, "failed to write recovery plan")
	} else {
		__antithesis_instrumentation__.Notify(31576)
	}
	__antithesis_instrumentation__.Notify(31526)

	_, _ = fmt.Fprint(stderr, "Plan created\nTo complete recovery, distribute the plan to the"+
		" below nodes and invoke `debug recover apply-plan` on:\n")
	for node, stores := range report.UpdatedNodes {
		__antithesis_instrumentation__.Notify(31577)
		_, _ = fmt.Fprintf(stderr, "- node n%d, store(s) %s\n", node, joinStoreIDs(stores))
	}
	__antithesis_instrumentation__.Notify(31527)

	return nil
}

func readReplicaInfoData(fileNames []string) ([]loqrecoverypb.NodeReplicaInfo, error) {
	__antithesis_instrumentation__.Notify(31578)
	var replicas []loqrecoverypb.NodeReplicaInfo
	for _, filename := range fileNames {
		__antithesis_instrumentation__.Notify(31580)
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			__antithesis_instrumentation__.Notify(31583)
			return nil, errors.Wrapf(err, "failed to read replica info file %q", filename)
		} else {
			__antithesis_instrumentation__.Notify(31584)
		}
		__antithesis_instrumentation__.Notify(31581)

		var nodeReplicas loqrecoverypb.NodeReplicaInfo
		jsonpb := protoutil.JSONPb{}
		if err = jsonpb.Unmarshal(data, &nodeReplicas); err != nil {
			__antithesis_instrumentation__.Notify(31585)
			return nil, errors.Wrapf(err, "failed to unmarshal replica info from file %q", filename)
		} else {
			__antithesis_instrumentation__.Notify(31586)
		}
		__antithesis_instrumentation__.Notify(31582)
		replicas = append(replicas, nodeReplicas)
	}
	__antithesis_instrumentation__.Notify(31579)
	return replicas, nil
}

var debugRecoverExecuteCmd = &cobra.Command{
	Use:   "apply-plan plan-file",
	Short: "update replicas in given stores according to recovery plan",
	Long: `
Apply changes to replicas in the provided stores using a plan.

This command will read a plan and update replicas that belong to the
given stores. Stores must be provided using --store flags. 

See debug recover command help for more details on how to use this command.
`,
	Args: cobra.ExactArgs(1),
	RunE: runDebugExecuteRecoverPlan,
}

var debugRecoverExecuteOpts struct {
	Stores        base.StoreSpecList
	confirmAction confirmActionFlag
}

func runDebugExecuteRecoverPlan(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31587)
	stopper := stop.NewStopper()
	defer stopper.Stop(cmd.Context())

	planFile := args[0]
	data, err := ioutil.ReadFile(planFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(31596)
		return errors.Wrapf(err, "failed to read plan file %q", planFile)
	} else {
		__antithesis_instrumentation__.Notify(31597)
	}
	__antithesis_instrumentation__.Notify(31588)

	var nodeUpdates loqrecoverypb.ReplicaUpdatePlan
	jsonpb := protoutil.JSONPb{Indent: "  "}
	if err = jsonpb.Unmarshal(data, &nodeUpdates); err != nil {
		__antithesis_instrumentation__.Notify(31598)
		return errors.Wrapf(err, "failed to unmarshal plan from file %q", planFile)
	} else {
		__antithesis_instrumentation__.Notify(31599)
	}
	__antithesis_instrumentation__.Notify(31589)

	var localNodeID roachpb.NodeID
	batches := make(map[roachpb.StoreID]storage.Batch)
	for _, storeSpec := range debugRecoverExecuteOpts.Stores.Specs {
		__antithesis_instrumentation__.Notify(31600)
		store, err := OpenExistingStore(storeSpec.Path, stopper, false, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(31604)
			return errors.Wrapf(err, "failed to open store at path %q. ensure that store path is "+
				"correct and that it is not used by another process", storeSpec.Path)
		} else {
			__antithesis_instrumentation__.Notify(31605)
		}
		__antithesis_instrumentation__.Notify(31601)
		batch := store.NewBatch()
		defer store.Close()
		defer batch.Close()

		storeIdent, err := kvserver.ReadStoreIdent(cmd.Context(), store)
		if err != nil {
			__antithesis_instrumentation__.Notify(31606)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31607)
		}
		__antithesis_instrumentation__.Notify(31602)
		if localNodeID != storeIdent.NodeID {
			__antithesis_instrumentation__.Notify(31608)
			if localNodeID != roachpb.NodeID(0) {
				__antithesis_instrumentation__.Notify(31610)
				return errors.Errorf("found stores from multiple node IDs n%d, n%d. "+
					"can only run in context of single node.", localNodeID, storeIdent.NodeID)
			} else {
				__antithesis_instrumentation__.Notify(31611)
			}
			__antithesis_instrumentation__.Notify(31609)
			localNodeID = storeIdent.NodeID
		} else {
			__antithesis_instrumentation__.Notify(31612)
		}
		__antithesis_instrumentation__.Notify(31603)
		batches[storeIdent.StoreID] = batch
	}
	__antithesis_instrumentation__.Notify(31590)

	updateTime := timeutil.Now()
	prepReport, err := loqrecovery.PrepareUpdateReplicas(
		cmd.Context(), nodeUpdates, uuid.DefaultGenerator, updateTime, localNodeID, batches)
	if err != nil {
		__antithesis_instrumentation__.Notify(31613)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31614)
	}
	__antithesis_instrumentation__.Notify(31591)

	for _, r := range prepReport.SkippedReplicas {
		__antithesis_instrumentation__.Notify(31615)
		_, _ = fmt.Fprintf(stderr, "Replica %s for range r%d is already updated.\n",
			r.Replica, r.RangeID())
	}
	__antithesis_instrumentation__.Notify(31592)

	if len(prepReport.UpdatedReplicas) == 0 {
		__antithesis_instrumentation__.Notify(31616)
		if len(prepReport.MissingStores) > 0 {
			__antithesis_instrumentation__.Notify(31618)
			return errors.Newf("stores %s expected on the node but no paths were provided",
				joinStoreIDs(prepReport.MissingStores))
		} else {
			__antithesis_instrumentation__.Notify(31619)
		}
		__antithesis_instrumentation__.Notify(31617)
		_, _ = fmt.Fprintf(stderr, "No updates planned on this node.\n")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(31620)
	}
	__antithesis_instrumentation__.Notify(31593)

	for _, r := range prepReport.UpdatedReplicas {
		__antithesis_instrumentation__.Notify(31621)
		message := fmt.Sprintf(
			"Replica %s for range %d:%s will be updated to %s with peer replica(s) removed: %s",
			r.OldReplica, r.RangeID(), r.StartKey(), r.Replica, r.RemovedReplicas)
		if r.AbortedTransaction {
			__antithesis_instrumentation__.Notify(31623)
			message += fmt.Sprintf(", and range update transaction %s aborted.",
				r.AbortedTransactionID.Short())
		} else {
			__antithesis_instrumentation__.Notify(31624)
		}
		__antithesis_instrumentation__.Notify(31622)
		_, _ = fmt.Fprintf(stderr, "%s\n", message)
	}
	__antithesis_instrumentation__.Notify(31594)

	switch debugRecoverExecuteOpts.confirmAction {
	case prompt:
		__antithesis_instrumentation__.Notify(31625)
		_, _ = fmt.Fprintf(stderr, "\nProceed with above changes [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			__antithesis_instrumentation__.Notify(31629)
			return errors.Wrap(err, "failed to read user input")
		} else {
			__antithesis_instrumentation__.Notify(31630)
		}
		__antithesis_instrumentation__.Notify(31626)
		_, _ = fmt.Fprintf(stderr, "\n")
		if len(line) < 1 || func() bool {
			__antithesis_instrumentation__.Notify(31631)
			return (line[0] != 'y' && func() bool {
				__antithesis_instrumentation__.Notify(31632)
				return line[0] != 'Y' == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(31633)
			_, _ = fmt.Fprint(stderr, "Aborted at user request\n")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(31634)
		}
	case allYes:
		__antithesis_instrumentation__.Notify(31627)

	default:
		__antithesis_instrumentation__.Notify(31628)
		return errors.New("Aborted by --confirm option")
	}
	__antithesis_instrumentation__.Notify(31595)

	applyReport, err := loqrecovery.CommitReplicaChanges(batches)
	_, _ = fmt.Fprintf(stderr, "Updated store(s): %s\n", joinStoreIDs(applyReport.UpdatedStores))
	return err
}

func joinStoreIDs(storeIDs []roachpb.StoreID) string {
	__antithesis_instrumentation__.Notify(31635)
	storeNames := make([]string, 0, len(storeIDs))
	for _, id := range storeIDs {
		__antithesis_instrumentation__.Notify(31637)
		storeNames = append(storeNames, fmt.Sprintf("s%d", id))
	}
	__antithesis_instrumentation__.Notify(31636)
	return strings.Join(storeNames, ", ")
}

func setDebugRecoverContextDefaults() {
	__antithesis_instrumentation__.Notify(31638)
	debugRecoverCollectInfoOpts.Stores.Specs = nil
	debugRecoverPlanOpts.outputFileName = ""
	debugRecoverPlanOpts.confirmAction = prompt
	debugRecoverPlanOpts.deadStoreIDs = nil
	debugRecoverExecuteOpts.Stores.Specs = nil
	debugRecoverExecuteOpts.confirmAction = prompt
}
