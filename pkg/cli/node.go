package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	localTimeFormat = "2006-01-02 15:04:05.999999-07:00"
)

var lsNodesColumnHeaders = []string{
	"id",
}

var lsNodesCmd = &cobra.Command{
	Use:   "ls",
	Short: "lists the IDs of all nodes in the cluster",
	Long: `
Display the node IDs for all active (that is, running and not decommissioned) members of the cluster.
To retrieve the IDs for inactive members, see 'node status --decommission'.
	`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runLsNodes),
}

func runLsNodes(cmd *cobra.Command, args []string) (resErr error) {
	__antithesis_instrumentation__.Notify(33508)
	conn, err := makeSQLClient("cockroach node ls", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(33513)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33514)
	}
	__antithesis_instrumentation__.Notify(33509)
	defer func() {
		__antithesis_instrumentation__.Notify(33515)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(33510)

	ctx := context.Background()

	if cliCtx.cmdTimeout != 0 {
		__antithesis_instrumentation__.Notify(33516)
		if err := conn.Exec(ctx,
			"SET statement_timeout = $1", cliCtx.cmdTimeout.String()); err != nil {
			__antithesis_instrumentation__.Notify(33517)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33518)
		}
	} else {
		__antithesis_instrumentation__.Notify(33519)
	}
	__antithesis_instrumentation__.Notify(33511)

	_, rows, err := sqlExecCtx.RunQuery(ctx,
		conn,
		clisqlclient.MakeQuery(`SELECT node_id FROM crdb_internal.gossip_liveness
               WHERE membership = 'active' OR split_part(expiration,',',1)::decimal > now()::decimal`),
		false,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(33520)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33521)
	}
	__antithesis_instrumentation__.Notify(33512)

	return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, lsNodesColumnHeaders,
		clisqlexec.NewRowSliceIter(rows, "r"))
}

var baseNodeColumnHeaders = []string{
	"id",
	"address",
	"sql_address",
	"build",
	"started_at",
	"updated_at",
	"locality",
	"is_available",
	"is_live",
}

var statusNodesColumnHeadersForRanges = []string{
	"replicas_leaders",
	"replicas_leaseholders",
	"ranges",
	"ranges_unavailable",
	"ranges_underreplicated",
}

var statusNodesColumnHeadersForStats = []string{
	"live_bytes",
	"key_bytes",
	"value_bytes",
	"intent_bytes",
	"system_bytes",
}

var statusNodesColumnHeadersForDecommission = []string{
	"gossiped_replicas",
	"is_decommissioning",
	"membership",
	"is_draining",
}

var statusNodeCmd = &cobra.Command{
	Use:   "status [<node id>]",
	Short: "shows the status of a node or all nodes",
	Long: `
If a node ID is specified, this will show the status for the corresponding node. If no node ID
is specified, this will display the status for all nodes in the cluster.
	`,
	Args: cobra.MaximumNArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runStatusNode),
}

func runStatusNode(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33522)
	_, rows, err := runStatusNodeInner(nodeCtx.statusShowDecommission || func() bool {
		__antithesis_instrumentation__.Notify(33524)
		return nodeCtx.statusShowAll == true
	}() == true, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(33525)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33526)
	}
	__antithesis_instrumentation__.Notify(33523)

	sliceIter := clisqlexec.NewRowSliceIter(rows, getStatusNodeAlignment())
	return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, getStatusNodeHeaders(), sliceIter)
}

func runStatusNodeInner(
	showDecommissioned bool, args []string,
) (colNames []string, rowVals [][]string, resErr error) {
	__antithesis_instrumentation__.Notify(33527)
	joinUsingID := func(queries []string) (query string) {
		__antithesis_instrumentation__.Notify(33536)
		for i, q := range queries {
			__antithesis_instrumentation__.Notify(33538)
			if i == 0 {
				__antithesis_instrumentation__.Notify(33540)
				query = q
				continue
			} else {
				__antithesis_instrumentation__.Notify(33541)
			}
			__antithesis_instrumentation__.Notify(33539)
			query = "(" + query + ") LEFT JOIN (" + q + ") USING (id)"
		}
		__antithesis_instrumentation__.Notify(33537)
		return
	}
	__antithesis_instrumentation__.Notify(33528)

	maybeAddActiveNodesFilter := func(query string) string {
		__antithesis_instrumentation__.Notify(33542)
		if !showDecommissioned {
			__antithesis_instrumentation__.Notify(33544)
			query += " WHERE membership = 'active' OR split_part(expiration,',',1)::decimal > now()::decimal"
		} else {
			__antithesis_instrumentation__.Notify(33545)
		}
		__antithesis_instrumentation__.Notify(33543)
		return query
	}
	__antithesis_instrumentation__.Notify(33529)

	baseQuery := maybeAddActiveNodesFilter(
		`SELECT node_id AS id,
            address,
            sql_address,
            build_tag AS build,
            started_at,
			updated_at,
			locality,
            CASE WHEN split_part(expiration,',',1)::decimal > now()::decimal
                 THEN true
                 ELSE false
                 END AS is_available,
			ifnull(is_live, false)
     FROM crdb_internal.gossip_liveness LEFT JOIN crdb_internal.gossip_nodes USING (node_id)`,
	)

	const rangesQuery = `
SELECT node_id AS id,
       sum((metrics->>'replicas.leaders')::DECIMAL)::INT AS replicas_leaders,
       sum((metrics->>'replicas.leaseholders')::DECIMAL)::INT AS replicas_leaseholders,
       sum((metrics->>'replicas')::DECIMAL)::INT AS ranges,
       sum((metrics->>'ranges.unavailable')::DECIMAL)::INT AS ranges_unavailable,
       sum((metrics->>'ranges.underreplicated')::DECIMAL)::INT AS ranges_underreplicated
FROM crdb_internal.kv_store_status
GROUP BY node_id`

	const statsQuery = `
SELECT node_id AS id,
       sum((metrics->>'livebytes')::DECIMAL)::INT AS live_bytes,
       sum((metrics->>'keybytes')::DECIMAL)::INT AS key_bytes,
       sum((metrics->>'valbytes')::DECIMAL)::INT AS value_bytes,
       sum((metrics->>'intentbytes')::DECIMAL)::INT AS intent_bytes,
       sum((metrics->>'sysbytes')::DECIMAL)::INT AS system_bytes
FROM crdb_internal.kv_store_status
GROUP BY node_id`

	const decommissionQuery = `
SELECT node_id AS id,
       ranges AS gossiped_replicas,
       membership != 'active' as is_decommissioning,
       membership AS membership,
       draining AS is_draining
FROM crdb_internal.gossip_liveness LEFT JOIN crdb_internal.gossip_nodes USING (node_id)`

	conn, err := makeSQLClient("cockroach node status", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(33546)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(33547)
	}
	__antithesis_instrumentation__.Notify(33530)
	defer func() {
		__antithesis_instrumentation__.Notify(33548)
		resErr = errors.CombineErrors(resErr, conn.Close())
	}()
	__antithesis_instrumentation__.Notify(33531)

	queriesToJoin := []string{baseQuery}

	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33549)
		return nodeCtx.statusShowRanges == true
	}() == true {
		__antithesis_instrumentation__.Notify(33550)
		queriesToJoin = append(queriesToJoin, rangesQuery)
	} else {
		__antithesis_instrumentation__.Notify(33551)
	}
	__antithesis_instrumentation__.Notify(33532)
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33552)
		return nodeCtx.statusShowStats == true
	}() == true {
		__antithesis_instrumentation__.Notify(33553)
		queriesToJoin = append(queriesToJoin, statsQuery)
	} else {
		__antithesis_instrumentation__.Notify(33554)
	}
	__antithesis_instrumentation__.Notify(33533)
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33555)
		return nodeCtx.statusShowDecommission == true
	}() == true {
		__antithesis_instrumentation__.Notify(33556)
		queriesToJoin = append(queriesToJoin, decommissionQuery)
	} else {
		__antithesis_instrumentation__.Notify(33557)
	}
	__antithesis_instrumentation__.Notify(33534)

	ctx := context.Background()

	if cliCtx.cmdTimeout != 0 {
		__antithesis_instrumentation__.Notify(33558)
		if err := conn.Exec(ctx,
			"SET statement_timeout = $1", cliCtx.cmdTimeout.String()); err != nil {
			__antithesis_instrumentation__.Notify(33559)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(33560)
		}
	} else {
		__antithesis_instrumentation__.Notify(33561)
	}
	__antithesis_instrumentation__.Notify(33535)

	queryString := "SELECT * FROM (" + joinUsingID(queriesToJoin) + ")"

	switch len(args) {
	case 0:
		__antithesis_instrumentation__.Notify(33562)
		query := clisqlclient.MakeQuery(queryString + " ORDER BY id")
		return sqlExecCtx.RunQuery(ctx, conn, query, false)
	case 1:
		__antithesis_instrumentation__.Notify(33563)
		nodeID, err := strconv.Atoi(args[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(33568)
			return nil, nil, errors.Errorf("could not parse node_id %s", args[0])
		} else {
			__antithesis_instrumentation__.Notify(33569)
		}
		__antithesis_instrumentation__.Notify(33564)
		query := clisqlclient.MakeQuery(queryString+" WHERE id = $1", nodeID)
		headers, rows, err := sqlExecCtx.RunQuery(ctx, conn, query, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(33570)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(33571)
		}
		__antithesis_instrumentation__.Notify(33565)
		if len(rows) == 0 {
			__antithesis_instrumentation__.Notify(33572)
			return nil, nil, errors.Errorf("node %d doesn't exist", nodeID)
		} else {
			__antithesis_instrumentation__.Notify(33573)
		}
		__antithesis_instrumentation__.Notify(33566)
		return headers, rows, nil
	default:
		__antithesis_instrumentation__.Notify(33567)
		return nil, nil, errors.Errorf("expected no arguments or a single node ID")
	}
}

func getStatusNodeHeaders() []string {
	__antithesis_instrumentation__.Notify(33574)
	headers := baseNodeColumnHeaders

	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33578)
		return nodeCtx.statusShowRanges == true
	}() == true {
		__antithesis_instrumentation__.Notify(33579)
		headers = append(headers, statusNodesColumnHeadersForRanges...)
	} else {
		__antithesis_instrumentation__.Notify(33580)
	}
	__antithesis_instrumentation__.Notify(33575)
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33581)
		return nodeCtx.statusShowStats == true
	}() == true {
		__antithesis_instrumentation__.Notify(33582)
		headers = append(headers, statusNodesColumnHeadersForStats...)
	} else {
		__antithesis_instrumentation__.Notify(33583)
	}
	__antithesis_instrumentation__.Notify(33576)
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33584)
		return nodeCtx.statusShowDecommission == true
	}() == true {
		__antithesis_instrumentation__.Notify(33585)
		headers = append(headers, statusNodesColumnHeadersForDecommission...)
	} else {
		__antithesis_instrumentation__.Notify(33586)
	}
	__antithesis_instrumentation__.Notify(33577)
	return headers
}

func getStatusNodeAlignment() string {
	__antithesis_instrumentation__.Notify(33587)
	align := "rlllll"
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33591)
		return nodeCtx.statusShowRanges == true
	}() == true {
		__antithesis_instrumentation__.Notify(33592)
		align += "rrrrrr"
	} else {
		__antithesis_instrumentation__.Notify(33593)
	}
	__antithesis_instrumentation__.Notify(33588)
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33594)
		return nodeCtx.statusShowStats == true
	}() == true {
		__antithesis_instrumentation__.Notify(33595)
		align += "rrrrrr"
	} else {
		__antithesis_instrumentation__.Notify(33596)
	}
	__antithesis_instrumentation__.Notify(33589)
	if nodeCtx.statusShowAll || func() bool {
		__antithesis_instrumentation__.Notify(33597)
		return nodeCtx.statusShowDecommission == true
	}() == true {
		__antithesis_instrumentation__.Notify(33598)
		align += decommissionResponseAlignment()
	} else {
		__antithesis_instrumentation__.Notify(33599)
	}
	__antithesis_instrumentation__.Notify(33590)
	return align
}

var decommissionNodesColumnHeaders = []string{
	"id",
	"is_live",
	"replicas",
	"is_decommissioning",
	"membership",
	"is_draining",
}

var decommissionNodeCmd = &cobra.Command{
	Use:   "decommission { --self | <node id 1> [<node id 2> ...] }",
	Short: "decommissions the node(s)",
	Long: `
Marks the nodes with the supplied IDs as decommissioning.
This will cause leases and replicas to be removed from these nodes.`,
	Args: cobra.MinimumNArgs(0),
	RunE: clierrorplus.MaybeDecorateError(runDecommissionNode),
}

func parseNodeIDs(strNodeIDs []string) ([]roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(33600)
	nodeIDs := make([]roachpb.NodeID, 0, len(strNodeIDs))
	for _, str := range strNodeIDs {
		__antithesis_instrumentation__.Notify(33602)
		i, err := strconv.ParseInt(str, 10, 32)
		if err != nil {
			__antithesis_instrumentation__.Notify(33604)
			return nil, errors.Wrapf(err, "unable to parse %s", str)
		} else {
			__antithesis_instrumentation__.Notify(33605)
		}
		__antithesis_instrumentation__.Notify(33603)
		nodeIDs = append(nodeIDs, roachpb.NodeID(i))
	}
	__antithesis_instrumentation__.Notify(33601)
	return nodeIDs, nil
}

func runDecommissionNode(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33606)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if nodeCtx.nodeDecommissionSelf {
		__antithesis_instrumentation__.Notify(33615)
		log.Warningf(ctx, "--%s for decommission is deprecated.", cliflags.NodeDecommissionSelf.Name)
	} else {
		__antithesis_instrumentation__.Notify(33616)
	}
	__antithesis_instrumentation__.Notify(33607)

	if !nodeCtx.nodeDecommissionSelf && func() bool {
		__antithesis_instrumentation__.Notify(33617)
		return len(args) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33618)
		return errors.New("no node ID specified")
	} else {
		__antithesis_instrumentation__.Notify(33619)
	}
	__antithesis_instrumentation__.Notify(33608)

	nodeIDs, err := parseNodeIDs(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(33620)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33621)
	}
	__antithesis_instrumentation__.Notify(33609)

	conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(33622)
		return errors.Wrap(err, "failed to connect to the node")
	} else {
		__antithesis_instrumentation__.Notify(33623)
	}
	__antithesis_instrumentation__.Notify(33610)
	defer finish()

	s := serverpb.NewStatusClient(conn)

	localNodeID, err := getLocalNodeID(ctx, s)
	if err != nil {
		__antithesis_instrumentation__.Notify(33624)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33625)
	}
	__antithesis_instrumentation__.Notify(33611)

	nodeIDs, err = handleNodeDecommissionSelf(ctx, nodeIDs, localNodeID, "decommissioning")
	if err != nil {
		__antithesis_instrumentation__.Notify(33626)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33627)
	}
	__antithesis_instrumentation__.Notify(33612)

	if err := expectNodesDecommissioned(ctx, s, nodeIDs, false); err != nil {
		__antithesis_instrumentation__.Notify(33628)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33629)
	}
	__antithesis_instrumentation__.Notify(33613)

	c := serverpb.NewAdminClient(conn)
	if err := runDecommissionNodeImpl(ctx, c, nodeCtx.nodeDecommissionWait, nodeIDs, localNodeID); err != nil {
		__antithesis_instrumentation__.Notify(33630)
		cause := errors.UnwrapAll(err)
		if s, ok := status.FromError(cause); ok && func() bool {
			__antithesis_instrumentation__.Notify(33632)
			return s.Code() == codes.NotFound == true
		}() == true {
			__antithesis_instrumentation__.Notify(33633)

			return errors.New("node does not exist")
		} else {
			__antithesis_instrumentation__.Notify(33634)
		}
		__antithesis_instrumentation__.Notify(33631)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33635)
	}
	__antithesis_instrumentation__.Notify(33614)
	return nil
}

func getLocalNodeID(ctx context.Context, s serverpb.StatusClient) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(33636)
	var nodeID roachpb.NodeID
	resp, err := s.Node(ctx, &serverpb.NodeRequest{NodeId: "local"})
	if err != nil {
		__antithesis_instrumentation__.Notify(33638)
		return nodeID, err
	} else {
		__antithesis_instrumentation__.Notify(33639)
	}
	__antithesis_instrumentation__.Notify(33637)
	nodeID = resp.Desc.NodeID
	return nodeID, nil
}

func handleNodeDecommissionSelf(
	ctx context.Context, nodeIDs []roachpb.NodeID, localNodeID roachpb.NodeID, command string,
) ([]roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(33640)
	if !nodeCtx.nodeDecommissionSelf {
		__antithesis_instrumentation__.Notify(33643)

		return nodeIDs, nil
	} else {
		__antithesis_instrumentation__.Notify(33644)
	}
	__antithesis_instrumentation__.Notify(33641)

	if len(nodeIDs) > 0 {
		__antithesis_instrumentation__.Notify(33645)
		return nil, errors.Newf("cannot use --%s with an explicit list of node IDs",
			cliflags.NodeDecommissionSelf.Name)
	} else {
		__antithesis_instrumentation__.Notify(33646)
	}
	__antithesis_instrumentation__.Notify(33642)

	log.Infof(ctx, "%s node %d", redact.Safe(command), localNodeID)
	return []roachpb.NodeID{localNodeID}, nil
}

func expectNodesDecommissioned(
	ctx context.Context, s serverpb.StatusClient, nodeIDs []roachpb.NodeID, expDecommissioned bool,
) error {
	__antithesis_instrumentation__.Notify(33647)
	resp, err := s.Nodes(ctx, &serverpb.NodesRequest{})
	if err != nil {
		__antithesis_instrumentation__.Notify(33650)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33651)
	}
	__antithesis_instrumentation__.Notify(33648)
	for _, nodeID := range nodeIDs {
		__antithesis_instrumentation__.Notify(33652)
		liveness, ok := resp.LivenessByNodeID[nodeID]
		if !ok {
			__antithesis_instrumentation__.Notify(33654)
			fmt.Fprintln(stderr, "warning: cannot find status of node", nodeID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(33655)
		}
		__antithesis_instrumentation__.Notify(33653)
		if !expDecommissioned {
			__antithesis_instrumentation__.Notify(33656)

			switch liveness {
			case livenesspb.NodeLivenessStatus_DECOMMISSIONING,
				livenesspb.NodeLivenessStatus_DECOMMISSIONED:
				__antithesis_instrumentation__.Notify(33657)
				fmt.Fprintln(stderr, "warning: node", nodeID, "is already decommissioning or decommissioned")
			default:
				__antithesis_instrumentation__.Notify(33658)

			}
		} else {
			__antithesis_instrumentation__.Notify(33659)

			switch liveness {
			case livenesspb.NodeLivenessStatus_DECOMMISSIONING,
				livenesspb.NodeLivenessStatus_DECOMMISSIONED:
				__antithesis_instrumentation__.Notify(33660)

			case livenesspb.NodeLivenessStatus_LIVE:
				__antithesis_instrumentation__.Notify(33661)
				fmt.Fprintln(stderr, "warning: node", nodeID, "is not decommissioned")
			default:
				__antithesis_instrumentation__.Notify(33662)
				fmt.Fprintln(stderr, "warning: node", nodeID, "is in unexpected state", liveness)
			}
		}
	}
	__antithesis_instrumentation__.Notify(33649)
	return nil
}

func runDecommissionNodeImpl(
	ctx context.Context,
	c serverpb.AdminClient,
	wait nodeDecommissionWaitType,
	nodeIDs []roachpb.NodeID,
	localNodeID roachpb.NodeID,
) error {
	__antithesis_instrumentation__.Notify(33663)
	minReplicaCount := int64(math.MaxInt64)
	opts := retry.Options{
		InitialBackoff: 5 * time.Millisecond,
		Multiplier:     2,
		MaxBackoff:     20 * time.Second,
	}

	prevResponse := serverpb.DecommissionStatusResponse{}
	for r := retry.StartWithCtx(ctx, opts); r.Next(); {
		__antithesis_instrumentation__.Notify(33665)
		req := &serverpb.DecommissionRequest{
			NodeIDs:          nodeIDs,
			TargetMembership: livenesspb.MembershipStatus_DECOMMISSIONING,
		}
		resp, err := c.Decommission(ctx, req)
		if err != nil {
			__antithesis_instrumentation__.Notify(33671)
			fmt.Fprintln(stderr)
			return errors.Wrap(err, "while trying to mark as decommissioning")
		} else {
			__antithesis_instrumentation__.Notify(33672)
		}
		__antithesis_instrumentation__.Notify(33666)

		if !reflect.DeepEqual(&prevResponse, resp) {
			__antithesis_instrumentation__.Notify(33673)
			fmt.Fprintln(stderr)
			if err := printDecommissionStatus(*resp); err != nil {
				__antithesis_instrumentation__.Notify(33675)
				return err
			} else {
				__antithesis_instrumentation__.Notify(33676)
			}
			__antithesis_instrumentation__.Notify(33674)
			prevResponse = *resp
		} else {
			__antithesis_instrumentation__.Notify(33677)
			fmt.Fprintf(stderr, ".")
		}
		__antithesis_instrumentation__.Notify(33667)

		anyActive := false
		var replicaCount int64
		for _, status := range resp.Status {
			__antithesis_instrumentation__.Notify(33678)
			anyActive = anyActive || func() bool {
				__antithesis_instrumentation__.Notify(33679)
				return status.Membership.Active() == true
			}() == true
			replicaCount += status.ReplicaCount
		}
		__antithesis_instrumentation__.Notify(33668)

		if !anyActive && func() bool {
			__antithesis_instrumentation__.Notify(33680)
			return replicaCount == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(33681)

			for _, targetNode := range nodeIDs {
				__antithesis_instrumentation__.Notify(33684)
				if targetNode == localNodeID {
					__antithesis_instrumentation__.Notify(33686)

					log.Warningf(ctx,
						"skipping drain step for node n%d; it is decommissioning and serving the request",
						localNodeID,
					)
					continue
				} else {
					__antithesis_instrumentation__.Notify(33687)
				}
				__antithesis_instrumentation__.Notify(33685)
				drainReq := &serverpb.DrainRequest{
					Shutdown: false,
					DoDrain:  true,
					NodeId:   targetNode.String(),
				}
				if _, err = c.Drain(ctx, drainReq); err != nil {
					__antithesis_instrumentation__.Notify(33688)
					fmt.Fprintln(stderr)
					return errors.Wrapf(err, "while trying to drain n%d", targetNode)
				} else {
					__antithesis_instrumentation__.Notify(33689)
				}
			}
			__antithesis_instrumentation__.Notify(33682)

			decommissionReq := &serverpb.DecommissionRequest{
				NodeIDs:          nodeIDs,
				TargetMembership: livenesspb.MembershipStatus_DECOMMISSIONED,
			}
			_, err = c.Decommission(ctx, decommissionReq)
			if err != nil {
				__antithesis_instrumentation__.Notify(33690)
				fmt.Fprintln(stderr)
				return errors.Wrap(err, "while trying to mark as decommissioned")
			} else {
				__antithesis_instrumentation__.Notify(33691)
			}
			__antithesis_instrumentation__.Notify(33683)

			fmt.Fprintln(os.Stdout, "\nNo more data reported on target nodes. "+
				"Please verify cluster health before removing the nodes.")
			return nil
		} else {
			__antithesis_instrumentation__.Notify(33692)
		}
		__antithesis_instrumentation__.Notify(33669)

		if wait == nodeDecommissionWaitNone {
			__antithesis_instrumentation__.Notify(33693)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(33694)
		}
		__antithesis_instrumentation__.Notify(33670)
		if replicaCount < minReplicaCount {
			__antithesis_instrumentation__.Notify(33695)
			minReplicaCount = replicaCount
			r.Reset()
		} else {
			__antithesis_instrumentation__.Notify(33696)
		}
	}
	__antithesis_instrumentation__.Notify(33664)
	return errors.New("maximum number of retries exceeded")
}

func decommissionResponseAlignment() string {
	__antithesis_instrumentation__.Notify(33697)
	return "rcrccc"
}

func decommissionResponseValueToRows(
	statuses []serverpb.DecommissionStatusResponse_Status,
) [][]string {
	__antithesis_instrumentation__.Notify(33698)

	var rows [][]string
	for _, node := range statuses {
		__antithesis_instrumentation__.Notify(33700)
		rows = append(rows, []string{
			strconv.FormatInt(int64(node.NodeID), 10),
			strconv.FormatBool(node.IsLive),
			strconv.FormatInt(node.ReplicaCount, 10),
			strconv.FormatBool(!node.Membership.Active()),
			node.Membership.String(),
			strconv.FormatBool(node.Draining),
		})
	}
	__antithesis_instrumentation__.Notify(33699)
	return rows
}

var recommissionNodeCmd = &cobra.Command{
	Use:   "recommission { --self | <node id 1> [<node id 2> ...] }",
	Short: "recommissions the node(s)",
	Long: `
For the nodes with the supplied IDs, resets the decommissioning states,
signaling the affected nodes to participate in the cluster again.
	`,
	Args: cobra.MinimumNArgs(0),
	RunE: clierrorplus.MaybeDecorateError(runRecommissionNode),
}

func printDecommissionStatus(resp serverpb.DecommissionStatusResponse) error {
	__antithesis_instrumentation__.Notify(33701)
	return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, decommissionNodesColumnHeaders,
		clisqlexec.NewRowSliceIter(decommissionResponseValueToRows(resp.Status), decommissionResponseAlignment()))
}

func runRecommissionNode(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33702)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !nodeCtx.nodeDecommissionSelf && func() bool {
		__antithesis_instrumentation__.Notify(33710)
		return len(args) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33711)
		return errors.New("no node ID specified; use --self to target the node specified with --host")
	} else {
		__antithesis_instrumentation__.Notify(33712)
	}
	__antithesis_instrumentation__.Notify(33703)

	nodeIDs, err := parseNodeIDs(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(33713)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33714)
	}
	__antithesis_instrumentation__.Notify(33704)

	conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(33715)
		return errors.Wrap(err, "failed to connect to the node")
	} else {
		__antithesis_instrumentation__.Notify(33716)
	}
	__antithesis_instrumentation__.Notify(33705)
	defer finish()

	s := serverpb.NewStatusClient(conn)

	localNodeID, err := getLocalNodeID(ctx, s)
	if err != nil {
		__antithesis_instrumentation__.Notify(33717)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33718)
	}
	__antithesis_instrumentation__.Notify(33706)

	nodeIDs, err = handleNodeDecommissionSelf(ctx, nodeIDs, localNodeID, "recommissioning")
	if err != nil {
		__antithesis_instrumentation__.Notify(33719)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33720)
	}
	__antithesis_instrumentation__.Notify(33707)

	if err := expectNodesDecommissioned(ctx, s, nodeIDs, true); err != nil {
		__antithesis_instrumentation__.Notify(33721)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33722)
	}
	__antithesis_instrumentation__.Notify(33708)

	c := serverpb.NewAdminClient(conn)
	req := &serverpb.DecommissionRequest{
		NodeIDs:          nodeIDs,
		TargetMembership: livenesspb.MembershipStatus_ACTIVE,
	}
	resp, err := c.Decommission(ctx, req)
	if err != nil {
		__antithesis_instrumentation__.Notify(33723)
		cause := errors.UnwrapAll(err)

		if s, ok := status.FromError(cause); ok && func() bool {
			__antithesis_instrumentation__.Notify(33726)
			return s.Code() == codes.FailedPrecondition == true
		}() == true {
			__antithesis_instrumentation__.Notify(33727)
			return errors.Newf("%s", s.Message())
		} else {
			__antithesis_instrumentation__.Notify(33728)
		}
		__antithesis_instrumentation__.Notify(33724)
		if s, ok := status.FromError(cause); ok && func() bool {
			__antithesis_instrumentation__.Notify(33729)
			return s.Code() == codes.NotFound == true
		}() == true {
			__antithesis_instrumentation__.Notify(33730)

			fmt.Fprintln(stderr)
			return errors.New("node does not exist")
		} else {
			__antithesis_instrumentation__.Notify(33731)
		}
		__antithesis_instrumentation__.Notify(33725)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33732)
	}
	__antithesis_instrumentation__.Notify(33709)
	return printDecommissionStatus(*resp)
}

var drainNodeCmd = &cobra.Command{
	Use:   "drain { --self | <node id> }",
	Short: "drain a node without shutting it down",
	Long: `
Prepare a server so it becomes ready to be shut down safely.
This causes the server to stop accepting client connections, stop
extant connections, and finally push range leases onto other
nodes, subject to various timeout parameters configurable via
cluster settings.

After a successful drain, the server process is still running;
use a service manager or orchestrator to terminate the process
gracefully using e.g. a unix signal.

If an argument is specified, the command affects the node
whose ID is given. If --self is specified, the command
affects the node that the command is connected to (via --host).
`,
	Args: cobra.MaximumNArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runDrain),
}

func runDrain(cmd *cobra.Command, args []string) (err error) {
	__antithesis_instrumentation__.Notify(33733)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if !quitCtx.nodeDrainSelf && func() bool {
		__antithesis_instrumentation__.Notify(33739)
		return len(args) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33740)
		fmt.Fprintf(stderr, "warning: draining a node without node ID or passing --self explicitly is deprecated.\n")
		quitCtx.nodeDrainSelf = true
	} else {
		__antithesis_instrumentation__.Notify(33741)
	}
	__antithesis_instrumentation__.Notify(33734)
	if quitCtx.nodeDrainSelf && func() bool {
		__antithesis_instrumentation__.Notify(33742)
		return len(args) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33743)
		return errors.Newf("cannot use --%s with an explicit node ID", cliflags.NodeDrainSelf.Name)
	} else {
		__antithesis_instrumentation__.Notify(33744)
	}
	__antithesis_instrumentation__.Notify(33735)

	targetNode := "local"
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(33745)
		targetNode = args[0]
	} else {
		__antithesis_instrumentation__.Notify(33746)
	}
	__antithesis_instrumentation__.Notify(33736)

	defer func() {
		__antithesis_instrumentation__.Notify(33747)
		if err == nil {
			__antithesis_instrumentation__.Notify(33748)
			fmt.Println("ok")
		} else {
			__antithesis_instrumentation__.Notify(33749)
		}
	}()
	__antithesis_instrumentation__.Notify(33737)

	c, finish, err := getAdminClient(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(33750)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33751)
	}
	__antithesis_instrumentation__.Notify(33738)
	defer finish()

	_, _, err = doDrain(ctx, c, targetNode)
	return err
}

var nodeCmds = []*cobra.Command{
	lsNodesCmd,
	statusNodeCmd,
	decommissionNodeCmd,
	recommissionNodeCmd,
	drainNodeCmd,
}

var nodeCmd = &cobra.Command{
	Use:   "node [command]",
	Short: "list, inspect, drain or remove nodes\n",
	Long:  "List, inspect, drain or remove nodes.",
	RunE:  UsageAndErr,
}

func init() {
	nodeCmd.AddCommand(nodeCmds...)
}
