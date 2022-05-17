package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/ts"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	yaml "gopkg.in/yaml.v2"
)

const maxBatchSize = 10000

func maybeImportTS(ctx context.Context, s *Server) (returnErr error) {
	__antithesis_instrumentation__.Notify(193653)
	var deferError func(error)
	{
		__antithesis_instrumentation__.Notify(193668)
		var defErr error
		deferError = func(err error) {
			__antithesis_instrumentation__.Notify(193670)
			log.Infof(ctx, "%v", err)
			defErr = errors.CombineErrors(defErr, err)
		}
		__antithesis_instrumentation__.Notify(193669)
		defer func() {
			__antithesis_instrumentation__.Notify(193671)
			if returnErr == nil {
				__antithesis_instrumentation__.Notify(193672)
				returnErr = defErr
			} else {
				__antithesis_instrumentation__.Notify(193673)
			}
		}()
	}
	__antithesis_instrumentation__.Notify(193654)
	knobs, _ := s.cfg.TestingKnobs.Server.(*TestingKnobs)
	if knobs == nil {
		__antithesis_instrumentation__.Notify(193674)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(193675)
	}
	__antithesis_instrumentation__.Notify(193655)
	tsImport := knobs.ImportTimeseriesFile
	if tsImport == "" {
		__antithesis_instrumentation__.Notify(193676)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(193677)
	}
	__antithesis_instrumentation__.Notify(193656)

	s.node.suppressNodeStatus.Set(true)

	for _, stmt := range []string{
		"SET CLUSTER SETTING kv.raft_log.disable_synchronization_unsafe = 'true';",
		"SET CLUSTER SETTING timeseries.storage.enabled = 'false';",
		"SET CLUSTER SETTING timeseries.storage.resolution_10s.ttl = '99999h';",
		"SET CLUSTER SETTING timeseries.storage.resolution_30m.ttl = '99999h';",
	} {
		__antithesis_instrumentation__.Notify(193678)
		if _, err := s.sqlServer.internalExecutor.ExecEx(
			ctx, "tsdump-cfg", nil,
			sessiondata.InternalExecutorOverride{
				User: security.RootUserName(),
			},
			stmt,
		); err != nil {
			__antithesis_instrumentation__.Notify(193679)
			return errors.Wrapf(err, "%s", stmt)
		} else {
			__antithesis_instrumentation__.Notify(193680)
		}
	}
	__antithesis_instrumentation__.Notify(193657)
	if tsImport == "-" {
		__antithesis_instrumentation__.Notify(193681)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(193682)
	}
	__antithesis_instrumentation__.Notify(193658)

	if !s.InitialStart() || func() bool {
		__antithesis_instrumentation__.Notify(193683)
		return len(s.cfg.JoinList) > 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(193684)
		return len(s.cfg.Stores.Specs) != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(193685)
		return errors.New("cannot import timeseries into an existing cluster or a multi-{store,node} cluster")
	} else {
		__antithesis_instrumentation__.Notify(193686)
	}
	__antithesis_instrumentation__.Notify(193659)

	f, err := os.Open(tsImport)
	if err != nil {
		__antithesis_instrumentation__.Notify(193687)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193688)
	}
	__antithesis_instrumentation__.Notify(193660)
	defer f.Close()

	if knobs.ImportTimeseriesMappingFile == "" {
		__antithesis_instrumentation__.Notify(193689)
		return errors.Errorf("need to specify COCKROACH_DEBUG_TS_IMPORT_MAPPING_FILE; it should point at " +
			"a YAML file that maps StoreID to NodeID. To generate from the source cluster, run the following command:\n \n" +
			"cockroach sql --url \"<(unix/sql) url>\" --format tsv -e \\\n  \"select concat(store_id::string, ': ', node_id::string)" +
			"from crdb_internal.kv_store_status\" | \\\n  grep -E '[0-9]+: [0-9]+' | tee tsdump.gob.yaml\n \n" +
			"To create from a debug.zip file, run the following command:\n \n" +
			"tail -n +2 debug/crdb_internal.kv_store_status.txt | awk '{print $2 \": \" $1}' > tsdump.gob.yaml\n \n" +
			"Then export the created file: export COCKROACH_DEBUG_TS_IMPORT_MAPPING_FILE=tsdump.gob.yaml\n \n")
	} else {
		__antithesis_instrumentation__.Notify(193690)
	}
	__antithesis_instrumentation__.Notify(193661)
	mapBytes, err := ioutil.ReadFile(knobs.ImportTimeseriesMappingFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(193691)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193692)
	}
	__antithesis_instrumentation__.Notify(193662)
	storeToNode := map[roachpb.StoreID]roachpb.NodeID{}
	if err := yaml.NewDecoder(bytes.NewReader(mapBytes)).Decode(&storeToNode); err != nil {
		__antithesis_instrumentation__.Notify(193693)
		return err
	} else {
		__antithesis_instrumentation__.Notify(193694)
	}
	__antithesis_instrumentation__.Notify(193663)
	batch := &kv.Batch{}

	var batchSize int
	maybeFlush := func(force bool) error {
		__antithesis_instrumentation__.Notify(193695)
		if batchSize == 0 {
			__antithesis_instrumentation__.Notify(193699)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(193700)
		}
		__antithesis_instrumentation__.Notify(193696)
		if batchSize < maxBatchSize && func() bool {
			__antithesis_instrumentation__.Notify(193701)
			return !force == true
		}() == true {
			__antithesis_instrumentation__.Notify(193702)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(193703)
		}
		__antithesis_instrumentation__.Notify(193697)
		err := s.db.Run(ctx, batch)
		if err != nil {
			__antithesis_instrumentation__.Notify(193704)
			return err
		} else {
			__antithesis_instrumentation__.Notify(193705)
		}
		__antithesis_instrumentation__.Notify(193698)
		log.Infof(ctx, "imported %d ts pairs\n", batchSize)
		*batch, batchSize = kv.Batch{}, 0
		return nil
	}
	__antithesis_instrumentation__.Notify(193664)

	nodeIDs := map[string]struct{}{}
	storeIDs := map[string]struct{}{}
	dec := gob.NewDecoder(f)
	for {
		__antithesis_instrumentation__.Notify(193706)
		var v roachpb.KeyValue
		err := dec.Decode(&v)
		if err != nil {
			__antithesis_instrumentation__.Notify(193710)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(193712)
				if err := maybeFlush(true); err != nil {
					__antithesis_instrumentation__.Notify(193714)
					return err
				} else {
					__antithesis_instrumentation__.Notify(193715)
				}
				__antithesis_instrumentation__.Notify(193713)
				break
			} else {
				__antithesis_instrumentation__.Notify(193716)
			}
			__antithesis_instrumentation__.Notify(193711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(193717)
		}
		__antithesis_instrumentation__.Notify(193707)

		name, source, _, _, err := ts.DecodeDataKey(v.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(193718)
			deferError(err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(193719)
		}
		__antithesis_instrumentation__.Notify(193708)
		if strings.HasPrefix(name, "cr.node.") {
			__antithesis_instrumentation__.Notify(193720)
			nodeIDs[source] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(193721)
			if strings.HasPrefix(name, "cr.store.") {
				__antithesis_instrumentation__.Notify(193722)
				storeIDs[source] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(193723)
				deferError(errors.Errorf("unknown metric %s", name))
				continue
			}
		}
		__antithesis_instrumentation__.Notify(193709)

		p := roachpb.NewPut(v.Key, v.Value)
		p.(*roachpb.PutRequest).Inline = true
		batch.AddRawRequest(p)
		batchSize++
		if err := maybeFlush(false); err != nil {
			__antithesis_instrumentation__.Notify(193724)
			return err
		} else {
			__antithesis_instrumentation__.Notify(193725)
		}
	}
	__antithesis_instrumentation__.Notify(193665)

	fakeStatuses := makeFakeNodeStatuses(storeToNode)
	if err := checkFakeStatuses(fakeStatuses, storeIDs); err != nil {
		__antithesis_instrumentation__.Notify(193726)

		deferError(errors.Wrapf(err, "consider updating the mapping file %s or restarting the server with "+
			"COCKROACH_DEBUG_TS_IMPORT_FILE=- to ignore the error", knobs.ImportTimeseriesMappingFile))
	} else {
		__antithesis_instrumentation__.Notify(193727)
	}
	__antithesis_instrumentation__.Notify(193666)

	for _, status := range fakeStatuses {
		__antithesis_instrumentation__.Notify(193728)
		key := keys.NodeStatusKey(status.Desc.NodeID)
		if err := s.db.PutInline(ctx, key, &status); err != nil {
			__antithesis_instrumentation__.Notify(193729)
			return err
		} else {
			__antithesis_instrumentation__.Notify(193730)
		}
	}
	__antithesis_instrumentation__.Notify(193667)

	return nil
}

func makeFakeNodeStatuses(storeToNode map[roachpb.StoreID]roachpb.NodeID) []statuspb.NodeStatus {
	__antithesis_instrumentation__.Notify(193731)
	var sl []statuspb.NodeStatus
	nodeToStore := map[roachpb.NodeID][]roachpb.StoreID{}
	for sid, nid := range storeToNode {
		__antithesis_instrumentation__.Notify(193735)
		nodeToStore[nid] = append(nodeToStore[nid], sid)
	}
	__antithesis_instrumentation__.Notify(193732)

	for nodeID, storeIDs := range nodeToStore {
		__antithesis_instrumentation__.Notify(193736)
		sort.Slice(storeIDs, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(193739)
			return storeIDs[i] < storeIDs[j]
		})
		__antithesis_instrumentation__.Notify(193737)
		nodeStatus := statuspb.NodeStatus{
			Desc: roachpb.NodeDescriptor{
				NodeID: nodeID,
			},
		}
		for _, storeID := range storeIDs {
			__antithesis_instrumentation__.Notify(193740)
			nodeStatus.StoreStatuses = append(nodeStatus.StoreStatuses, statuspb.StoreStatus{Desc: roachpb.StoreDescriptor{
				Node:    nodeStatus.Desc,
				StoreID: storeID,
			}})
		}
		__antithesis_instrumentation__.Notify(193738)

		sl = append(sl, nodeStatus)
	}
	__antithesis_instrumentation__.Notify(193733)
	sort.Slice(sl, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(193741)
		return sl[i].Desc.NodeID < sl[j].Desc.NodeID
	})
	__antithesis_instrumentation__.Notify(193734)
	return sl
}

func checkFakeStatuses(fakeStatuses []statuspb.NodeStatus, storeIDs map[string]struct{}) error {
	__antithesis_instrumentation__.Notify(193742)
	for _, status := range fakeStatuses {
		__antithesis_instrumentation__.Notify(193745)
		for _, ss := range status.StoreStatuses {
			__antithesis_instrumentation__.Notify(193746)
			storeID := ss.Desc.StoreID
			strID := fmt.Sprint(storeID)
			if _, ok := storeIDs[strID]; !ok {
				__antithesis_instrumentation__.Notify(193748)

				return errors.Errorf(
					"s%d supplied in input mapping, but no timeseries found for it",
					ss.Desc.StoreID,
				)
			} else {
				__antithesis_instrumentation__.Notify(193749)
			}
			__antithesis_instrumentation__.Notify(193747)

			delete(storeIDs, strID)
		}
	}
	__antithesis_instrumentation__.Notify(193743)
	if len(storeIDs) > 0 {
		__antithesis_instrumentation__.Notify(193750)
		return errors.Errorf(
			"need to map the remaining stores %v to nodes", storeIDs)
	} else {
		__antithesis_instrumentation__.Notify(193751)
	}
	__antithesis_instrumentation__.Notify(193744)
	return nil
}
