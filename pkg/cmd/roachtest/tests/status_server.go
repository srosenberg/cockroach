package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
)

func runStatusServer(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(51205)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	idMap := make(map[int]roachpb.NodeID)
	urlMap := make(map[int]string)
	adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.All())
	if err != nil {
		__antithesis_instrumentation__.Notify(51211)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51212)
	}
	__antithesis_instrumentation__.Notify(51206)
	for i, addr := range adminUIAddrs {
		__antithesis_instrumentation__.Notify(51213)
		var details serverpb.DetailsResponse
		url := `http://` + addr + `/_status/details/local`

		if err := retry.ForDuration(10*time.Second, func() error {
			__antithesis_instrumentation__.Notify(51215)
			return httputil.GetJSON(http.Client{}, url, &details)
		}); err != nil {
			__antithesis_instrumentation__.Notify(51216)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51217)
		}
		__antithesis_instrumentation__.Notify(51214)
		idMap[i+1] = details.NodeID
		urlMap[i+1] = `http://` + addr
	}
	__antithesis_instrumentation__.Notify(51207)

	httpClient := httputil.NewClientWithTimeout(15 * time.Second)

	get := func(base, rel string) []byte {
		__antithesis_instrumentation__.Notify(51218)
		url := base + rel
		resp, err := httpClient.Get(context.TODO(), url)
		if err != nil {
			__antithesis_instrumentation__.Notify(51222)
			t.Fatalf("could not GET %s - %s", url, err)
		} else {
			__antithesis_instrumentation__.Notify(51223)
		}
		__antithesis_instrumentation__.Notify(51219)
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			__antithesis_instrumentation__.Notify(51224)
			t.Fatalf("could not read body for %s - %s", url, err)
		} else {
			__antithesis_instrumentation__.Notify(51225)
		}
		__antithesis_instrumentation__.Notify(51220)
		if resp.StatusCode != http.StatusOK {
			__antithesis_instrumentation__.Notify(51226)
			t.Fatalf("could not GET %s - statuscode: %d - body: %s", url, resp.StatusCode, body)
		} else {
			__antithesis_instrumentation__.Notify(51227)
		}
		__antithesis_instrumentation__.Notify(51221)
		t.L().Printf("OK response from %s\n", url)
		return body
	}
	__antithesis_instrumentation__.Notify(51208)

	checkNode := func(url string, nodeID, otherNodeID, expectedNodeID roachpb.NodeID) {
		__antithesis_instrumentation__.Notify(51228)
		urlIDs := []string{otherNodeID.String()}
		if nodeID == otherNodeID {
			__antithesis_instrumentation__.Notify(51231)
			urlIDs = append(urlIDs, "local")
		} else {
			__antithesis_instrumentation__.Notify(51232)
		}
		__antithesis_instrumentation__.Notify(51229)
		var details serverpb.DetailsResponse
		for _, urlID := range urlIDs {
			__antithesis_instrumentation__.Notify(51233)
			if err := httputil.GetJSON(http.Client{}, url+`/_status/details/`+urlID, &details); err != nil {
				__antithesis_instrumentation__.Notify(51236)
				t.Fatalf("unable to parse details - %s", err)
			} else {
				__antithesis_instrumentation__.Notify(51237)
			}
			__antithesis_instrumentation__.Notify(51234)
			if details.NodeID != expectedNodeID {
				__antithesis_instrumentation__.Notify(51238)
				t.Fatalf("%d calling %s: node ids don't match - expected %d, actual %d",
					nodeID, urlID, expectedNodeID, details.NodeID)
			} else {
				__antithesis_instrumentation__.Notify(51239)
			}
			__antithesis_instrumentation__.Notify(51235)

			get(url, fmt.Sprintf("/_status/gossip/%s", urlID))
			get(url, fmt.Sprintf("/_status/nodes/%s", urlID))
			get(url, fmt.Sprintf("/_status/logfiles/%s", urlID))
			get(url, fmt.Sprintf("/_status/logs/%s", urlID))
			get(url, fmt.Sprintf("/_status/stacks/%s", urlID))
		}
		__antithesis_instrumentation__.Notify(51230)

		get(url, "/_status/vars")
	}
	__antithesis_instrumentation__.Notify(51209)

	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(51240)
		id := idMap[i]
		checkNode(urlMap[i], id, id, id)
		get(urlMap[i], "/_status/nodes")
	}
	__antithesis_instrumentation__.Notify(51210)

	firstNode := 1
	lastNode := c.Spec().NodeCount
	firstID := idMap[firstNode]
	lastID := idMap[lastNode]
	checkNode(urlMap[firstNode], firstID, lastID, lastID)

	checkNode(urlMap[lastNode], lastID, firstID, firstID)

	checkNode(urlMap[lastNode], lastID, lastID, lastID)
}
