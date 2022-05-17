package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/spf13/cobra"
)

var debugListFilesCmd = &cobra.Command{
	Use:   "list-files",
	Short: "list files available for retrieval via 'debug zip'",
	RunE:  clierrorplus.MaybeDecorateError(runDebugListFiles),
}

func runDebugListFiles(cmd *cobra.Command, _ []string) error {
	__antithesis_instrumentation__.Notify(31141)
	if err := zipCtx.files.validate(); err != nil {
		__antithesis_instrumentation__.Notify(31152)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31153)
	}
	__antithesis_instrumentation__.Notify(31142)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(31154)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31155)
	}
	__antithesis_instrumentation__.Notify(31143)
	defer finish()

	status := serverpb.NewStatusClient(conn)

	firstNodeDetails, err := status.Details(ctx, &serverpb.DetailsRequest{NodeId: "local"})
	if err != nil {
		__antithesis_instrumentation__.Notify(31156)
		return err
	} else {
		__antithesis_instrumentation__.Notify(31157)
	}
	__antithesis_instrumentation__.Notify(31144)

	nodes, err := status.Nodes(ctx, &serverpb.NodesRequest{})
	if err != nil {
		__antithesis_instrumentation__.Notify(31158)
		log.Warningf(ctx, "cannot retrieve node list: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(31159)
	}
	__antithesis_instrumentation__.Notify(31145)

	inputNodeList := []statuspb.NodeStatus{{Desc: roachpb.NodeDescriptor{
		NodeID:     firstNodeDetails.NodeID,
		Address:    firstNodeDetails.Address,
		SQLAddress: firstNodeDetails.SQLAddress,
	}}}
	if nodes != nil {
		__antithesis_instrumentation__.Notify(31160)

		inputNodeList = nodes.Nodes
	} else {
		__antithesis_instrumentation__.Notify(31161)
	}
	__antithesis_instrumentation__.Notify(31146)

	var nodeList []roachpb.NodeID
	for _, n := range inputNodeList {
		__antithesis_instrumentation__.Notify(31162)
		if zipCtx.nodes.isIncluded(n.Desc.NodeID) {
			__antithesis_instrumentation__.Notify(31163)
			nodeList = append(nodeList, n.Desc.NodeID)
		} else {
			__antithesis_instrumentation__.Notify(31164)
		}
	}
	__antithesis_instrumentation__.Notify(31147)

	fileTypes := make([]int, 0, len(serverpb.FileType_value))
	for _, v := range serverpb.FileType_value {
		__antithesis_instrumentation__.Notify(31165)
		fileTypes = append(fileTypes, int(v))
	}
	__antithesis_instrumentation__.Notify(31148)
	sort.Ints(fileTypes)

	logFiles := make(map[roachpb.NodeID][]logpb.FileInfo)

	otherFiles := make(map[roachpb.NodeID]map[int32][]*serverpb.File)

	for _, nodeID := range nodeList {
		__antithesis_instrumentation__.Notify(31166)
		nodeIDs := fmt.Sprintf("%d", nodeID)
		nodeLogs, err := status.LogFilesList(ctx, &serverpb.LogFilesListRequest{NodeId: nodeIDs})
		if err != nil {
			__antithesis_instrumentation__.Notify(31168)
			log.Warningf(ctx, "cannot retrieve log file list from node %d: %v", nodeID, err)
		} else {
			__antithesis_instrumentation__.Notify(31169)
			logFiles[nodeID] = nodeLogs.Files
		}
		__antithesis_instrumentation__.Notify(31167)

		otherFiles[nodeID] = make(map[int32][]*serverpb.File)
		for _, fileTypeI := range fileTypes {
			__antithesis_instrumentation__.Notify(31170)
			fileType := int32(fileTypeI)
			nodeFiles, err := status.GetFiles(ctx, &serverpb.GetFilesRequest{
				NodeId:   nodeIDs,
				ListOnly: true,
				Type:     serverpb.FileType(fileType),
				Patterns: zipCtx.files.retrievalPatterns(),
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(31171)
				log.Warningf(ctx, "cannot retrieve %s file list from node %d: %v", serverpb.FileType_name[fileType], nodeID, err)
			} else {
				__antithesis_instrumentation__.Notify(31172)
				otherFiles[nodeID][fileType] = nodeFiles.Files
			}
		}
	}
	__antithesis_instrumentation__.Notify(31149)

	fileTypeNames := map[int32]string{}
	for t, n := range serverpb.FileType_name {
		__antithesis_instrumentation__.Notify(31173)
		fileTypeNames[t] = strings.ToLower(n)
	}
	__antithesis_instrumentation__.Notify(31150)
	var totalSize int64
	fileTableHeaders := []string{"node_id", "type", "file_name", "ctime_utc", "mtime_utc", "size"}
	alignment := "lllr"
	var rows [][]string
	for _, nodeID := range nodeList {
		__antithesis_instrumentation__.Notify(31174)
		nodeIDs := fmt.Sprintf("%d", nodeID)
		for _, logFile := range logFiles[nodeID] {
			__antithesis_instrumentation__.Notify(31176)
			ctime := extractTimeFromFileName(logFile.Name)
			mtime := timeutil.Unix(0, logFile.ModTimeNanos)
			if !zipCtx.files.isIncluded(logFile.Name, ctime, mtime) {
				__antithesis_instrumentation__.Notify(31178)
				continue
			} else {
				__antithesis_instrumentation__.Notify(31179)
			}
			__antithesis_instrumentation__.Notify(31177)
			totalSize += logFile.SizeBytes
			ctimes := formatTimeSimple(ctime)
			mtimes := formatTimeSimple(mtime)
			rows = append(rows, []string{nodeIDs, "log", logFile.Name, ctimes, mtimes, fmt.Sprintf("%d", logFile.SizeBytes)})
		}
		__antithesis_instrumentation__.Notify(31175)
		for _, ft := range fileTypes {
			__antithesis_instrumentation__.Notify(31180)
			fileType := int32(ft)
			for _, other := range otherFiles[nodeID][fileType] {
				__antithesis_instrumentation__.Notify(31181)
				ctime := extractTimeFromFileName(other.Name)
				if !zipCtx.files.isIncluded(other.Name, ctime, ctime) {
					__antithesis_instrumentation__.Notify(31183)
					continue
				} else {
					__antithesis_instrumentation__.Notify(31184)
				}
				__antithesis_instrumentation__.Notify(31182)
				totalSize += other.FileSize
				ctimes := formatTimeSimple(ctime)
				rows = append(rows, []string{nodeIDs, fileTypeNames[fileType], other.Name, ctimes, ctimes, fmt.Sprintf("%d", other.FileSize)})
			}
		}
	}
	__antithesis_instrumentation__.Notify(31151)

	rows = append(rows, []string{"", "total", fmt.Sprintf("(%s)", humanizeutil.IBytes(totalSize)), "", "", fmt.Sprintf("%d", totalSize)})

	return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, fileTableHeaders, clisqlexec.NewRowSliceIter(rows, alignment))
}

var tzRe = regexp.MustCompile(`\d\d\d\d-\d\d-\d\dT\d\d_\d\d_\d\d`)

func formatTimeSimple(t time.Time) string {
	__antithesis_instrumentation__.Notify(31185)
	return t.Format("2006-01-02 15:04")
}

func extractTimeFromFileName(f string) time.Time {
	__antithesis_instrumentation__.Notify(31186)
	ts := tzRe.FindString(f)
	if ts == "" {
		__antithesis_instrumentation__.Notify(31189)

		return time.Time{}
	} else {
		__antithesis_instrumentation__.Notify(31190)
	}
	__antithesis_instrumentation__.Notify(31187)
	tm, err := time.ParseInLocation("2006-01-02T15_04_05", ts, time.UTC)
	if err != nil {
		__antithesis_instrumentation__.Notify(31191)

		return time.Time{}
	} else {
		__antithesis_instrumentation__.Notify(31192)
	}
	__antithesis_instrumentation__.Notify(31188)
	return tm
}
