package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func makePerNodeZipRequests(prefix, id string, status serverpb.StatusClient) []zipRequest {
	__antithesis_instrumentation__.Notify(35504)
	return []zipRequest{
		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35505)
				return status.Details(ctx, &serverpb.DetailsRequest{NodeId: id})
			},
			pathName: prefix + "/details",
		},
		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35506)
				return status.Gossip(ctx, &serverpb.GossipRequest{NodeId: id})
			},
			pathName: prefix + "/gossip",
		},
		{
			fn: func(ctx context.Context) (interface{}, error) {
				__antithesis_instrumentation__.Notify(35507)
				return status.EngineStats(ctx, &serverpb.EngineStatsRequest{NodeId: id})
			},
			pathName: prefix + "/enginestats",
		},
	}
}

var debugZipTablesPerNode = []string{
	"crdb_internal.feature_usage",

	"crdb_internal.gossip_alerts",
	"crdb_internal.gossip_liveness",
	"crdb_internal.gossip_network",
	"crdb_internal.gossip_nodes",

	"crdb_internal.leases",

	"crdb_internal.node_build_info",
	"crdb_internal.node_contention_events",
	"crdb_internal.node_distsql_flows",
	"crdb_internal.node_inflight_trace_spans",
	"crdb_internal.node_metrics",
	"crdb_internal.node_queries",
	"crdb_internal.node_runtime_info",
	"crdb_internal.node_sessions",
	"crdb_internal.node_statement_statistics",
	"crdb_internal.node_transaction_statistics",
	"crdb_internal.node_transactions",
	"crdb_internal.node_txn_stats",
	"crdb_internal.active_range_feeds",
}

func (zc *debugZipContext) collectCPUProfiles(
	ctx context.Context, ni nodesInfo, livenessByNodeID nodeLivenesses,
) error {
	__antithesis_instrumentation__.Notify(35508)
	if zipCtx.cpuProfDuration <= 0 {
		__antithesis_instrumentation__.Notify(35513)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(35514)
	}
	__antithesis_instrumentation__.Notify(35509)

	var wg sync.WaitGroup
	type profData struct {
		data []byte
		err  error
	}

	zc.clusterPrinter.info("requesting CPU profiles")

	if ni.nodesListResponse == nil {
		__antithesis_instrumentation__.Notify(35515)
		return errors.AssertionFailedf("nodes list is empty; nothing to do")
	} else {
		__antithesis_instrumentation__.Notify(35516)
	}
	__antithesis_instrumentation__.Notify(35510)

	nodeList := ni.nodesListResponse.Nodes

	resps := make([]profData, len(nodeList))
	for i := range nodeList {
		__antithesis_instrumentation__.Notify(35517)
		nodeID := roachpb.NodeID(nodeList[i].NodeID)
		if livenessByNodeID[nodeID] == livenesspb.NodeLivenessStatus_DECOMMISSIONED {
			__antithesis_instrumentation__.Notify(35519)
			continue
		} else {
			__antithesis_instrumentation__.Notify(35520)
		}
		__antithesis_instrumentation__.Notify(35518)
		wg.Add(1)
		go func(ctx context.Context, i int) {
			__antithesis_instrumentation__.Notify(35521)
			defer wg.Done()

			secs := int32(zipCtx.cpuProfDuration / time.Second)
			if secs < 1 {
				__antithesis_instrumentation__.Notify(35524)
				secs = 1
			} else {
				__antithesis_instrumentation__.Notify(35525)
			}
			__antithesis_instrumentation__.Notify(35522)

			var pd profData
			err := contextutil.RunWithTimeout(ctx, "fetch cpu profile", zc.timeout+zipCtx.cpuProfDuration, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(35526)
				resp, err := zc.status.Profile(ctx, &serverpb.ProfileRequest{
					NodeId:  fmt.Sprintf("%d", nodeID),
					Type:    serverpb.ProfileRequest_CPU,
					Seconds: secs,
					Labels:  true,
				})
				if err != nil {
					__antithesis_instrumentation__.Notify(35528)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35529)
				}
				__antithesis_instrumentation__.Notify(35527)
				pd = profData{data: resp.Data}
				return nil
			})
			__antithesis_instrumentation__.Notify(35523)
			if err != nil {
				__antithesis_instrumentation__.Notify(35530)
				resps[i] = profData{err: err}
			} else {
				__antithesis_instrumentation__.Notify(35531)
				resps[i] = pd
			}
		}(ctx, i)
	}
	__antithesis_instrumentation__.Notify(35511)

	wg.Wait()
	zc.clusterPrinter.info("profiles generated")

	for i, pd := range resps {
		__antithesis_instrumentation__.Notify(35532)
		if len(pd.data) == 0 && func() bool {
			__antithesis_instrumentation__.Notify(35534)
			return pd.err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(35535)
			continue
		} else {
			__antithesis_instrumentation__.Notify(35536)
		}
		__antithesis_instrumentation__.Notify(35533)
		nodeID := nodeList[i].NodeID
		prefix := fmt.Sprintf("%s/%s", nodesPrefix, fmt.Sprintf("%d", nodeID))
		s := zc.clusterPrinter.start("profile for node %d", nodeID)
		if err := zc.z.createRawOrError(s, prefix+"/cpu.pprof", pd.data, pd.err); err != nil {
			__antithesis_instrumentation__.Notify(35537)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35538)
		}
	}
	__antithesis_instrumentation__.Notify(35512)
	return nil
}

func (zc *debugZipContext) collectPerNodeData(
	ctx context.Context,
	nodeDetails serverpb.NodeDetails,
	nodeStatus *statuspb.NodeStatus,
	livenessByNodeID nodeLivenesses,
) error {
	__antithesis_instrumentation__.Notify(35539)
	nodeID := roachpb.NodeID(nodeDetails.NodeID)

	if livenessByNodeID != nil {
		__antithesis_instrumentation__.Notify(35556)
		liveness := livenessByNodeID[nodeID]
		if liveness == livenesspb.NodeLivenessStatus_DECOMMISSIONED {
			__antithesis_instrumentation__.Notify(35557)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(35558)
		}
	} else {
		__antithesis_instrumentation__.Notify(35559)
	}
	__antithesis_instrumentation__.Notify(35540)
	nodePrinter := zipCtx.newZipReporter("node %d", nodeID)
	id := fmt.Sprintf("%d", nodeID)
	prefix := fmt.Sprintf("%s/%s", nodesPrefix, id)

	if !zipCtx.nodes.isIncluded(nodeID) {
		__antithesis_instrumentation__.Notify(35560)
		if err := zc.z.createRaw(nodePrinter.start("skipping node"), prefix+".skipped",
			[]byte(fmt.Sprintf("skipping excluded node %d\n", nodeID))); err != nil {
			__antithesis_instrumentation__.Notify(35562)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35563)
		}
		__antithesis_instrumentation__.Notify(35561)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(35564)
	}
	__antithesis_instrumentation__.Notify(35541)
	if nodeStatus != nil {
		__antithesis_instrumentation__.Notify(35565)

		if err := zc.z.createJSON(nodePrinter.start("node status"), prefix+"/status.json", *nodeStatus); err != nil {
			__antithesis_instrumentation__.Notify(35566)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35567)
		}
	} else {
		__antithesis_instrumentation__.Notify(35568)
		if err := zc.z.createJSON(nodePrinter.start("node status"), prefix+"/status.json", nodeDetails); err != nil {
			__antithesis_instrumentation__.Notify(35569)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35570)
		}
	}
	__antithesis_instrumentation__.Notify(35542)

	sqlAddr := nodeDetails.SQLAddress
	if sqlAddr.IsEmpty() {
		__antithesis_instrumentation__.Notify(35571)
		sqlAddr = nodeDetails.Address
	} else {
		__antithesis_instrumentation__.Notify(35572)
	}
	__antithesis_instrumentation__.Notify(35543)
	curSQLConn := guessNodeURL(zc.firstNodeSQLConn.GetURL(), sqlAddr.AddressField)
	nodePrinter.info("using SQL connection URL: %s", curSQLConn.GetURL())

	for _, table := range debugZipTablesPerNode {
		__antithesis_instrumentation__.Notify(35573)
		query := fmt.Sprintf(`SELECT * FROM %s`, table)
		if override, ok := customQuery[table]; ok {
			__antithesis_instrumentation__.Notify(35575)
			query = override
		} else {
			__antithesis_instrumentation__.Notify(35576)
		}
		__antithesis_instrumentation__.Notify(35574)
		if err := zc.dumpTableDataForZip(nodePrinter, curSQLConn, prefix, table, query); err != nil {
			__antithesis_instrumentation__.Notify(35577)
			return errors.Wrapf(err, "fetching %s", table)
		} else {
			__antithesis_instrumentation__.Notify(35578)
		}
	}
	__antithesis_instrumentation__.Notify(35544)

	perNodeZipRequests := makePerNodeZipRequests(prefix, id, zc.status)

	for _, r := range perNodeZipRequests {
		__antithesis_instrumentation__.Notify(35579)
		if err := zc.runZipRequest(ctx, nodePrinter, r); err != nil {
			__antithesis_instrumentation__.Notify(35580)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35581)
		}
	}
	__antithesis_instrumentation__.Notify(35545)

	var stacksData []byte
	s := nodePrinter.start("requesting stacks")
	requestErr := zc.runZipFn(ctx, s,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35582)
			stacks, err := zc.status.Stacks(ctx, &serverpb.StacksRequest{
				NodeId: id,
				Type:   serverpb.StacksType_GOROUTINE_STACKS,
			})
			if err == nil {
				__antithesis_instrumentation__.Notify(35584)
				stacksData = stacks.Data
			} else {
				__antithesis_instrumentation__.Notify(35585)
			}
			__antithesis_instrumentation__.Notify(35583)
			return err
		})
	__antithesis_instrumentation__.Notify(35546)
	if err := zc.z.createRawOrError(s, prefix+"/stacks.txt", stacksData, requestErr); err != nil {
		__antithesis_instrumentation__.Notify(35586)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35587)
	}
	__antithesis_instrumentation__.Notify(35547)

	var stacksDataWithLabels []byte
	s = nodePrinter.start("requesting stacks with labels")

	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(35588)
		stacksDataWithLabels = []byte("disabled in race mode, see 74133")
	} else {
		__antithesis_instrumentation__.Notify(35589)
		requestErr = zc.runZipFn(ctx, s,
			func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(35590)
				stacks, err := zc.status.Stacks(ctx, &serverpb.StacksRequest{
					NodeId: id,
					Type:   serverpb.StacksType_GOROUTINE_STACKS_DEBUG_1,
				})
				if err == nil {
					__antithesis_instrumentation__.Notify(35592)
					stacksDataWithLabels = stacks.Data
				} else {
					__antithesis_instrumentation__.Notify(35593)
				}
				__antithesis_instrumentation__.Notify(35591)
				return err
			})
	}
	__antithesis_instrumentation__.Notify(35548)
	if err := zc.z.createRawOrError(s, prefix+"/stacks_with_labels.txt", stacksDataWithLabels, requestErr); err != nil {
		__antithesis_instrumentation__.Notify(35594)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35595)
	}
	__antithesis_instrumentation__.Notify(35549)

	var heapData []byte
	s = nodePrinter.start("requesting heap profile")
	requestErr = zc.runZipFn(ctx, s,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35596)
			heap, err := zc.status.Profile(ctx, &serverpb.ProfileRequest{
				NodeId: id,
				Type:   serverpb.ProfileRequest_HEAP,
			})
			if err == nil {
				__antithesis_instrumentation__.Notify(35598)
				heapData = heap.Data
			} else {
				__antithesis_instrumentation__.Notify(35599)
			}
			__antithesis_instrumentation__.Notify(35597)
			return err
		})
	__antithesis_instrumentation__.Notify(35550)
	if err := zc.z.createRawOrError(s, prefix+"/heap.pprof", heapData, requestErr); err != nil {
		__antithesis_instrumentation__.Notify(35600)
		return err
	} else {
		__antithesis_instrumentation__.Notify(35601)
	}
	__antithesis_instrumentation__.Notify(35551)

	var profiles *serverpb.GetFilesResponse
	s = nodePrinter.start("requesting heap file list")
	if requestErr := zc.runZipFn(ctx, s,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35602)
			var err error
			profiles, err = zc.status.GetFiles(ctx, &serverpb.GetFilesRequest{
				NodeId:   id,
				Type:     serverpb.FileType_HEAP,
				Patterns: zipCtx.files.retrievalPatterns(),
				ListOnly: true,
			})
			return err
		}); requestErr != nil {
		__antithesis_instrumentation__.Notify(35603)
		if err := zc.z.createError(s, prefix+"/heapprof", requestErr); err != nil {
			__antithesis_instrumentation__.Notify(35604)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35605)
		}
	} else {
		__antithesis_instrumentation__.Notify(35606)
		s.done()

		nodePrinter.info("%d heap profiles found", len(profiles.Files))
		for _, file := range profiles.Files {
			__antithesis_instrumentation__.Notify(35607)
			ctime := extractTimeFromFileName(file.Name)
			if !zipCtx.files.isIncluded(file.Name, ctime, ctime) {
				__antithesis_instrumentation__.Notify(35609)
				nodePrinter.info("skipping excluded heap profile: %s", file.Name)
				continue
			} else {
				__antithesis_instrumentation__.Notify(35610)
			}
			__antithesis_instrumentation__.Notify(35608)

			fName := maybeAddProfileSuffix(file.Name)
			name := prefix + "/heapprof/" + fName
			fs := nodePrinter.start("retrieving %s", file.Name)
			var oneprof *serverpb.GetFilesResponse
			if fileErr := zc.runZipFn(ctx, fs, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(35611)
				var err error
				oneprof, err = zc.status.GetFiles(ctx, &serverpb.GetFilesRequest{
					NodeId:   id,
					Type:     serverpb.FileType_HEAP,
					Patterns: []string{file.Name},
					ListOnly: false,
				})
				return err
			}); fileErr != nil {
				__antithesis_instrumentation__.Notify(35612)
				if err := zc.z.createError(fs, name, fileErr); err != nil {
					__antithesis_instrumentation__.Notify(35613)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35614)
				}
			} else {
				__antithesis_instrumentation__.Notify(35615)
				fs.done()

				if len(oneprof.Files) < 1 {
					__antithesis_instrumentation__.Notify(35617)

					continue
				} else {
					__antithesis_instrumentation__.Notify(35618)
				}
				__antithesis_instrumentation__.Notify(35616)
				file := oneprof.Files[0]
				if err := zc.z.createRaw(nodePrinter.start("writing profile"), name, file.Contents); err != nil {
					__antithesis_instrumentation__.Notify(35619)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35620)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(35552)

	var goroutinesResp *serverpb.GetFilesResponse
	s = nodePrinter.start("requesting goroutine dump list")
	if requestErr := zc.runZipFn(ctx, s,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35621)
			var err error
			goroutinesResp, err = zc.status.GetFiles(ctx, &serverpb.GetFilesRequest{
				NodeId:   id,
				Type:     serverpb.FileType_GOROUTINES,
				Patterns: zipCtx.files.retrievalPatterns(),
				ListOnly: true,
			})
			return err
		}); requestErr != nil {
		__antithesis_instrumentation__.Notify(35622)
		if err := zc.z.createError(s, prefix+"/goroutines", requestErr); err != nil {
			__antithesis_instrumentation__.Notify(35623)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35624)
		}
	} else {
		__antithesis_instrumentation__.Notify(35625)
		s.done()

		nodePrinter.info("%d goroutine dumps found", len(goroutinesResp.Files))
		for _, file := range goroutinesResp.Files {
			__antithesis_instrumentation__.Notify(35626)
			ctime := extractTimeFromFileName(file.Name)
			if !zipCtx.files.isIncluded(file.Name, ctime, ctime) {
				__antithesis_instrumentation__.Notify(35628)
				nodePrinter.info("skipping excluded goroutine dump: %s", file.Name)
				continue
			} else {
				__antithesis_instrumentation__.Notify(35629)
			}
			__antithesis_instrumentation__.Notify(35627)

			name := prefix + "/goroutines/" + file.Name

			fs := nodePrinter.start("retrieving %s", file.Name)
			var onedump *serverpb.GetFilesResponse
			if fileErr := zc.runZipFn(ctx, fs, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(35630)
				var err error
				onedump, err = zc.status.GetFiles(ctx, &serverpb.GetFilesRequest{
					NodeId:   id,
					Type:     serverpb.FileType_GOROUTINES,
					Patterns: []string{file.Name},
					ListOnly: false,
				})
				return err
			}); fileErr != nil {
				__antithesis_instrumentation__.Notify(35631)
				if err := zc.z.createError(fs, name, fileErr); err != nil {
					__antithesis_instrumentation__.Notify(35632)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35633)
				}
			} else {
				__antithesis_instrumentation__.Notify(35634)
				fs.done()

				if len(onedump.Files) < 1 {
					__antithesis_instrumentation__.Notify(35636)

					continue
				} else {
					__antithesis_instrumentation__.Notify(35637)
				}
				__antithesis_instrumentation__.Notify(35635)
				file := onedump.Files[0]
				if err := zc.z.createRaw(nodePrinter.start("writing dump"), name, file.Contents); err != nil {
					__antithesis_instrumentation__.Notify(35638)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35639)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(35553)

	var logs *serverpb.LogFilesListResponse
	s = nodePrinter.start("requesting log files list")
	if requestErr := zc.runZipFn(ctx, s,
		func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(35640)
			var err error
			logs, err = zc.status.LogFilesList(
				ctx, &serverpb.LogFilesListRequest{NodeId: id})
			return err
		}); requestErr != nil {
		__antithesis_instrumentation__.Notify(35641)
		if err := zc.z.createError(s, prefix+"/logs", requestErr); err != nil {
			__antithesis_instrumentation__.Notify(35642)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35643)
		}
	} else {
		__antithesis_instrumentation__.Notify(35644)
		s.done()

		nodePrinter.info("%d log files found", len(logs.Files))
		for _, file := range logs.Files {
			__antithesis_instrumentation__.Notify(35645)
			ctime := extractTimeFromFileName(file.Name)
			mtime := timeutil.Unix(0, file.ModTimeNanos)
			if !zipCtx.files.isIncluded(file.Name, ctime, mtime) {
				__antithesis_instrumentation__.Notify(35649)
				nodePrinter.info("skipping excluded log file: %s", file.Name)
				continue
			} else {
				__antithesis_instrumentation__.Notify(35650)
			}
			__antithesis_instrumentation__.Notify(35646)

			logPrinter := nodePrinter.withPrefix("log file: %s", file.Name)
			name := prefix + "/logs/" + file.Name
			var entries *serverpb.LogEntriesResponse
			sf := logPrinter.start("requesting file")
			if requestErr := zc.runZipFn(ctx, sf,
				func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(35651)
					var err error
					entries, err = zc.status.LogFile(
						ctx, &serverpb.LogFileRequest{
							NodeId: id, File: file.Name, Redact: zipCtx.redactLogs,
						})
					return err
				}); requestErr != nil {
				__antithesis_instrumentation__.Notify(35652)
				if err := zc.z.createError(sf, name, requestErr); err != nil {
					__antithesis_instrumentation__.Notify(35654)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35655)
				}
				__antithesis_instrumentation__.Notify(35653)
				continue
			} else {
				__antithesis_instrumentation__.Notify(35656)
			}
			__antithesis_instrumentation__.Notify(35647)
			sf.progress("writing output: %s", name)
			warnRedactLeak := false
			if err := func() error {
				__antithesis_instrumentation__.Notify(35657)

				zc.z.Lock()
				defer zc.z.Unlock()

				logOut, err := zc.z.createLocked(name, timeutil.Unix(0, file.ModTimeNanos))
				if err != nil {
					__antithesis_instrumentation__.Notify(35660)
					return err
				} else {
					__antithesis_instrumentation__.Notify(35661)
				}
				__antithesis_instrumentation__.Notify(35658)
				for _, e := range entries.Entries {
					__antithesis_instrumentation__.Notify(35662)

					if zipCtx.redactLogs && func() bool {
						__antithesis_instrumentation__.Notify(35664)
						return !e.Redactable == true
					}() == true {
						__antithesis_instrumentation__.Notify(35665)
						e.Message = "REDACTEDBYZIP"

						warnRedactLeak = true
					} else {
						__antithesis_instrumentation__.Notify(35666)
					}
					__antithesis_instrumentation__.Notify(35663)
					if err := log.FormatLegacyEntry(e, logOut); err != nil {
						__antithesis_instrumentation__.Notify(35667)
						return err
					} else {
						__antithesis_instrumentation__.Notify(35668)
					}
				}
				__antithesis_instrumentation__.Notify(35659)
				return nil
			}(); err != nil {
				__antithesis_instrumentation__.Notify(35669)
				return sf.fail(err)
			} else {
				__antithesis_instrumentation__.Notify(35670)
			}
			__antithesis_instrumentation__.Notify(35648)
			sf.done()
			if warnRedactLeak {
				__antithesis_instrumentation__.Notify(35671)

				defer func(fileName string) {
					__antithesis_instrumentation__.Notify(35672)
					fmt.Fprintf(stderr, "WARNING: server-side redaction failed for %s, completed client-side (--redact-logs=true)\n", fileName)
				}(file.Name)
			} else {
				__antithesis_instrumentation__.Notify(35673)
			}
		}
	}
	__antithesis_instrumentation__.Notify(35554)

	var ranges *serverpb.RangesResponse
	s = nodePrinter.start("requesting ranges")
	if requestErr := zc.runZipFn(ctx, s, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(35674)
		var err error
		ranges, err = zc.status.Ranges(ctx, &serverpb.RangesRequest{NodeId: id})
		return err
	}); requestErr != nil {
		__antithesis_instrumentation__.Notify(35675)
		if err := zc.z.createError(s, prefix+"/ranges", requestErr); err != nil {
			__antithesis_instrumentation__.Notify(35676)
			return err
		} else {
			__antithesis_instrumentation__.Notify(35677)
		}
	} else {
		__antithesis_instrumentation__.Notify(35678)
		s.done()
		nodePrinter.info("%d ranges found", len(ranges.Ranges))
		sort.Slice(ranges.Ranges, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(35680)
			return ranges.Ranges[i].State.Desc.RangeID <
				ranges.Ranges[j].State.Desc.RangeID
		})
		__antithesis_instrumentation__.Notify(35679)
		for _, r := range ranges.Ranges {
			__antithesis_instrumentation__.Notify(35681)
			s := nodePrinter.start("writing range %d", r.State.Desc.RangeID)
			name := fmt.Sprintf("%s/ranges/%s", prefix, r.State.Desc.RangeID)
			if err := zc.z.createJSON(s, name+".json", r); err != nil {
				__antithesis_instrumentation__.Notify(35682)
				return err
			} else {
				__antithesis_instrumentation__.Notify(35683)
			}
		}
	}
	__antithesis_instrumentation__.Notify(35555)
	return nil
}

func guessNodeURL(workingURL string, hostport string) clisqlclient.Conn {
	__antithesis_instrumentation__.Notify(35684)
	u, err := url.Parse(workingURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(35686)
		u = &url.URL{Host: "invalid"}
	} else {
		__antithesis_instrumentation__.Notify(35687)
	}
	__antithesis_instrumentation__.Notify(35685)
	u.Host = hostport
	return sqlConnCtx.MakeSQLConn(os.Stdout, stderr, u.String())
}
