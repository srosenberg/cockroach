package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/heapprofiler"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/marusama/semaphore"
	"github.com/spf13/cobra"
)

type zipRequest struct {
	fn       func(ctx context.Context) (interface{}, error)
	pathName string
}

var customQuery = map[string]string{
	"crdb_internal.node_inflight_trace_spans": "WITH spans AS (" +
		"SELECT * FROM crdb_internal.node_inflight_trace_spans " +
		"WHERE duration > INTERVAL '10' ORDER BY trace_id ASC, duration DESC" +
		") SELECT * FROM spans, LATERAL crdb_internal.payloads_for_span(span_id)",
	"system.jobs":       "SELECT *, to_hex(payload) AS hex_payload, to_hex(progress) AS hex_progress FROM system.jobs",
	"system.descriptor": "SELECT *, to_hex(descriptor) AS hex_descriptor FROM system.descriptor",
	"system.settings":   "SELECT *, to_hex(value) as hex_value FROM system.settings",
}

type debugZipContext struct {
	z              *zipper
	clusterPrinter *zipReporter
	timeout        time.Duration
	admin          serverpb.AdminClient
	status         serverpb.StatusClient

	firstNodeSQLConn clisqlclient.Conn

	sem semaphore.Semaphore
}

func (zc *debugZipContext) runZipFn(
	ctx context.Context, s *zipReporter, fn func(ctx context.Context) error,
) error {
	__antithesis_instrumentation__.Notify(35176)
	return zc.runZipFnWithTimeout(ctx, s, zc.timeout, fn)
}

func (zc *debugZipContext) runZipFnWithTimeout(
	ctx context.Context, s *zipReporter, timeout time.Duration, fn func(ctx context.Context) error,
) error {
	__antithesis_instrumentation__.Notify(35177)
	err := contextutil.RunWithTimeout(ctx, s.prefix, timeout, fn)
	s.progress("received response")
	return err
}

func (zc *debugZipContext) runZipRequest(ctx context.Context, zr *zipReporter, r zipRequest) error {
	__antithesis_instrumentation__.Notify(35178)
	s := zr.start("requesting data for %s", r.pathName)
	var data interface{}
	err := zc.runZipFn(ctx, s, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(35180)
		thisData, err := r.fn(ctx)
		data = thisData
		return err
	})
	__antithesis_instrumentation__.Notify(35179)
	return zc.z.createJSONOrError(s, r.pathName+".json", data, err)
}

func (zc *debugZipContext) forAllNodes(
	ctx context.Context,
	ni nodesInfo,
	fn func(ctx context.Context, nodeDetails serverpb.NodeDetails, nodeStatus *statuspb.NodeStatus) error,
) error {
	__antithesis_instrumentation__.Notify(35181)
	if ni.nodesListResponse == nil {
		__antithesis_instrumentation__.Notify(35187)

		return errors.AssertionFailedf("nodes list is empty")
	} else {
		__antithesis_instrumentation__.Notify(35188)
	}
	__antithesis_instrumentation__.Notify(35182)
	if ni.nodesStatusResponse != nil && func() bool {
		__antithesis_instrumentation__.Notify(35189)
		return len(ni.nodesStatusResponse.Nodes) != len(ni.nodesListResponse.Nodes) == true
	}() == true {
		__antithesis_instrumentation__.Notify(35190)
		return errors.AssertionFailedf("mismatching node status response and node list")
	} else {
		__antithesis_instrumentation__.Notify(35191)
	}
	__antithesis_instrumentation__.Notify(35183)
	if zipCtx.concurrency == 1 {
		__antithesis_instrumentation__.Notify(35192)

		for index, nodeDetails := range ni.nodesListResponse.Nodes {
			__antithesis_instrumentation__.Notify(35194)
			var nodeStatus *statuspb.NodeStatus

			if ni.nodesStatusResponse != nil {
				__antithesis_instrumentation__.Notify(35196)
				nodeStatus = &ni.nodesStatusResponse.Nodes[index]
			} else {
				__antithesis_instrumentation__.Notify(35197)
			}
			__antithesis_instrumentation__.Notify(35195)
			if err := fn(ctx, nodeDetails, nodeStatus); err != nil {
				__antithesis_instrumentation__.Notify(35198)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35199)
			}
		}
		__antithesis_instrumentation__.Notify(35193)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(35200)
	}
	__antithesis_instrumentation__.Notify(35184)

	nodeErrs := make(chan error, len(ni.nodesListResponse.Nodes))

	var wg sync.WaitGroup
	for index, nodeDetails := range ni.nodesListResponse.Nodes {
		__antithesis_instrumentation__.Notify(35201)
		wg.Add(1)
		var nodeStatus *statuspb.NodeStatus
		if ni.nodesStatusResponse != nil {
			__antithesis_instrumentation__.Notify(35203)
			nodeStatus = &ni.nodesStatusResponse.Nodes[index]
		} else {
			__antithesis_instrumentation__.Notify(35204)
		}
		__antithesis_instrumentation__.Notify(35202)
		go func(nodeDetails serverpb.NodeDetails, nodeStatus *statuspb.NodeStatus) {
			__antithesis_instrumentation__.Notify(35205)
			defer wg.Done()
			if err := zc.sem.Acquire(ctx, 1); err != nil {
				__antithesis_instrumentation__.Notify(35207)
				nodeErrs <- err
				return
			} else {
				__antithesis_instrumentation__.Notify(35208)
			}
			__antithesis_instrumentation__.Notify(35206)
			defer zc.sem.Release(1)

			nodeErrs <- fn(ctx, nodeDetails, nodeStatus)
		}(nodeDetails, nodeStatus)
	}
	__antithesis_instrumentation__.Notify(35185)
	wg.Wait()

	var err error
	for range ni.nodesListResponse.Nodes {
		__antithesis_instrumentation__.Notify(35209)
		err = errors.CombineErrors(err, <-nodeErrs)
	}
	__antithesis_instrumentation__.Notify(35186)
	return err
}

type nodeLivenesses = map[roachpb.NodeID]livenesspb.NodeLivenessStatus

func runDebugZip(_ *cobra.Command, args []string) (retErr error) {
	__antithesis_instrumentation__.Notify(35210)
	if err := zipCtx.files.validate(); err != nil {
		__antithesis_instrumentation__.Notify(35223)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35224)
	}
	__antithesis_instrumentation__.Notify(35211)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	zr := zipCtx.newZipReporter("cluster")

	s := zr.start("establishing RPC connection to %s", serverCfg.AdvertiseAddr)
	conn, _, finish, err := getClientGRPCConn(ctx, serverCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(35225)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35226)
	}
	__antithesis_instrumentation__.Notify(35212)
	defer finish()

	status := serverpb.NewStatusClient(conn)
	admin := serverpb.NewAdminClient(conn)
	s.done()

	s = zr.start("retrieving the node status to get the SQL address")
	firstNodeDetails, err := status.Details(ctx, &serverpb.DetailsRequest{NodeId: "local"})
	if err != nil {
		__antithesis_instrumentation__.Notify(35227)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35228)
	}
	__antithesis_instrumentation__.Notify(35213)
	s.done()

	sqlAddr := firstNodeDetails.SQLAddress
	if sqlAddr.IsEmpty() {
		__antithesis_instrumentation__.Notify(35229)

		sqlAddr = firstNodeDetails.Address
	} else {
		__antithesis_instrumentation__.Notify(35230)
	}
	__antithesis_instrumentation__.Notify(35214)
	s = zr.start("using SQL address: %s", sqlAddr.AddressField)

	cliCtx.clientConnHost, cliCtx.clientConnPort, err = net.SplitHostPort(sqlAddr.AddressField)
	if err != nil {
		__antithesis_instrumentation__.Notify(35231)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35232)
	}
	__antithesis_instrumentation__.Notify(35215)

	cliCtx.IsInteractive = false
	sqlExecCtx.TerminalOutput = false
	sqlExecCtx.ShowTimes = false

	sqlExecCtx.TableDisplayFormat = clisqlexec.TableDisplayTSV

	sqlConn, err := makeSQLClient("cockroach zip", useSystemDb)
	if err != nil {
		__antithesis_instrumentation__.Notify(35233)
		_ = s.fail(errors.Wrap(err, "unable to open a SQL session. Debug information will be incomplete"))
	} else {
		__antithesis_instrumentation__.Notify(35234)

		defer func() {
			__antithesis_instrumentation__.Notify(35236)
			retErr = errors.CombineErrors(retErr, sqlConn.Close())
		}()
		__antithesis_instrumentation__.Notify(35235)
		s.progress("using SQL connection URL: %s", sqlConn.GetURL())
		s.done()
	}
	__antithesis_instrumentation__.Notify(35216)

	name := args[0]
	s = zr.start("creating output file %s", name)
	out, err := os.Create(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(35237)
		return s.fail(err)
	} else {
		__antithesis_instrumentation__.Notify(35238)
	}
	__antithesis_instrumentation__.Notify(35217)

	z := newZipper(out)
	defer func() {
		__antithesis_instrumentation__.Notify(35239)
		cErr := z.close()
		retErr = errors.CombineErrors(retErr, cErr)
	}()
	__antithesis_instrumentation__.Notify(35218)
	s.done()

	timeout := 10 * time.Second
	if cliCtx.cmdTimeout != 0 {
		__antithesis_instrumentation__.Notify(35240)
		timeout = cliCtx.cmdTimeout
	} else {
		__antithesis_instrumentation__.Notify(35241)
	}
	__antithesis_instrumentation__.Notify(35219)

	zc := debugZipContext{
		clusterPrinter:   zr,
		z:                z,
		timeout:          timeout,
		admin:            admin,
		status:           status,
		firstNodeSQLConn: sqlConn,
		sem:              semaphore.New(zipCtx.concurrency),
	}

	ni, livenessByNodeID, err := zc.collectClusterData(ctx, firstNodeDetails)
	if err != nil {
		__antithesis_instrumentation__.Notify(35242)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35243)
	}
	__antithesis_instrumentation__.Notify(35220)

	if err := zc.collectCPUProfiles(ctx, ni, livenessByNodeID); err != nil {
		__antithesis_instrumentation__.Notify(35244)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35245)
	}
	__antithesis_instrumentation__.Notify(35221)

	if err := zc.forAllNodes(ctx, ni, func(ctx context.Context, nodeDetails serverpb.NodeDetails, nodesStatus *statuspb.NodeStatus) error {
		__antithesis_instrumentation__.Notify(35246)
		return zc.collectPerNodeData(ctx, nodeDetails, nodesStatus, livenessByNodeID)
	}); err != nil {
		__antithesis_instrumentation__.Notify(35247)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35248)
	}

	{
		__antithesis_instrumentation__.Notify(35249)
		s := zc.clusterPrinter.start("pprof summary script")
		if err := z.createRaw(s, debugBase+"/pprof-summary.sh", []byte(`#!/bin/sh
find . -name cpu.pprof -print0 | xargs -0 go tool pprof -tags
`)); err != nil {
			__antithesis_instrumentation__.Notify(35250)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35251)
		}
	}

	{
		__antithesis_instrumentation__.Notify(35252)
		s := zc.clusterPrinter.start("hot range summary script")
		if err := z.createRaw(s, debugBase+"/hot-ranges.sh", []byte(`#!/bin/sh
find . -path './nodes/*/ranges/*.json' -print0 | xargs -0 grep per_second | sort -rhk3 | head -n 20
`)); err != nil {
			__antithesis_instrumentation__.Notify(35253)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35254)
		}
	}

	{
		__antithesis_instrumentation__.Notify(35255)
		s := zc.clusterPrinter.start("tenant hot range summary script")
		if err := z.createRaw(s, debugBase+"/hot-ranges-tenant.sh", []byte(`#!/bin/sh
find . -path './tenant_ranges/*/*.json' -print0 | xargs -0 grep per_second | sort -rhk3 | head -n 20`)); err != nil {
			__antithesis_instrumentation__.Notify(35256)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35257)
		}
	}
	__antithesis_instrumentation__.Notify(35222)

	return nil
}

func maybeAddProfileSuffix(name string) string {
	__antithesis_instrumentation__.Notify(35258)
	switch {
	case strings.HasPrefix(name, heapprofiler.HeapFileNamePrefix+".") && func() bool {
		__antithesis_instrumentation__.Notify(35264)
		return !strings.HasSuffix(name, heapprofiler.HeapFileNameSuffix) == true
	}() == true:
		__antithesis_instrumentation__.Notify(35260)
		name += heapprofiler.HeapFileNameSuffix
	case strings.HasPrefix(name, heapprofiler.StatsFileNamePrefix+".") && func() bool {
		__antithesis_instrumentation__.Notify(35265)
		return !strings.HasSuffix(name, heapprofiler.StatsFileNameSuffix) == true
	}() == true:
		__antithesis_instrumentation__.Notify(35261)
		name += heapprofiler.StatsFileNameSuffix
	case strings.HasPrefix(name, heapprofiler.JemallocFileNamePrefix+".") && func() bool {
		__antithesis_instrumentation__.Notify(35266)
		return !strings.HasSuffix(name, heapprofiler.JemallocFileNameSuffix) == true
	}() == true:
		__antithesis_instrumentation__.Notify(35262)
		name += heapprofiler.JemallocFileNameSuffix
	default:
		__antithesis_instrumentation__.Notify(35263)
	}
	__antithesis_instrumentation__.Notify(35259)
	return name
}

func (zc *debugZipContext) dumpTableDataForZip(
	zr *zipReporter, conn clisqlclient.Conn, base, table, query string,
) error {
	__antithesis_instrumentation__.Notify(35267)

	fullQuery := fmt.Sprintf(`SET statement_timeout = '%s'; %s`, zc.timeout, query)
	baseName := base + "/" + sanitizeFilename(table)

	s := zr.start("retrieving SQL data for %s", table)
	const maxRetries = 5
	suffix := ""
	for numRetries := 1; numRetries <= maxRetries; numRetries++ {
		__antithesis_instrumentation__.Notify(35269)
		name := baseName + suffix + ".txt"
		s.progress("writing output: %s", name)
		sqlErr := func() error {
			__antithesis_instrumentation__.Notify(35272)
			zc.z.Lock()
			defer zc.z.Unlock()

			w, err := zc.z.createLocked(name, time.Time{})
			if err != nil {
				__antithesis_instrumentation__.Notify(35274)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35275)
			}
			__antithesis_instrumentation__.Notify(35273)

			return sqlExecCtx.RunQueryAndFormatResults(
				context.Background(),
				conn, w, stderr, clisqlclient.MakeQuery(fullQuery))
		}()
		__antithesis_instrumentation__.Notify(35270)
		if sqlErr != nil {
			__antithesis_instrumentation__.Notify(35276)
			if cErr := zc.z.createError(s, name, sqlErr); cErr != nil {
				__antithesis_instrumentation__.Notify(35280)
				return cErr
			} else {
				__antithesis_instrumentation__.Notify(35281)
			}
			__antithesis_instrumentation__.Notify(35277)
			var pqErr *pq.Error
			if !errors.As(sqlErr, &pqErr) {
				__antithesis_instrumentation__.Notify(35282)

				break
			} else {
				__antithesis_instrumentation__.Notify(35283)
			}
			__antithesis_instrumentation__.Notify(35278)
			if pgcode.MakeCode(string(pqErr.Code)) != pgcode.SerializationFailure {
				__antithesis_instrumentation__.Notify(35284)

				break
			} else {
				__antithesis_instrumentation__.Notify(35285)
			}
			__antithesis_instrumentation__.Notify(35279)

			suffix = fmt.Sprintf(".%d", numRetries)
			s = zr.start("retrying %s", table)
			continue
		} else {
			__antithesis_instrumentation__.Notify(35286)
		}
		__antithesis_instrumentation__.Notify(35271)
		s.done()
		break
	}
	__antithesis_instrumentation__.Notify(35268)
	return nil
}

func sanitizeFilename(f string) string {
	__antithesis_instrumentation__.Notify(35287)
	return strings.TrimPrefix(f, `"".`)
}
