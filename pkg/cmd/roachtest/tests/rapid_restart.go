package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"net/http"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func runRapidRestart(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(50060)

	node := c.Node(1)
	c.Put(ctx, t.Cockroach(), "./cockroach", node)

	deadline := timeutil.Now().Add(time.Minute)
	done := func() bool {
		__antithesis_instrumentation__.Notify(50063)
		return timeutil.Now().After(deadline)
	}
	__antithesis_instrumentation__.Notify(50061)
	for j := 1; !done(); j++ {
		__antithesis_instrumentation__.Notify(50064)
		c.Wipe(ctx, node)

		for i := 0; i < 3; i++ {
			__antithesis_instrumentation__.Notify(50067)
			startOpts := option.DefaultStartOpts()
			startOpts.RoachprodOpts.SkipInit = true
			if err := c.StartE(ctx, t.L(), startOpts, install.MakeClusterSettings(), node); err != nil {
				__antithesis_instrumentation__.Notify(50070)
				t.Fatalf("error during start: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(50071)
			}
			__antithesis_instrumentation__.Notify(50068)

			if i == 2 {
				__antithesis_instrumentation__.Notify(50072)
				break
			} else {
				__antithesis_instrumentation__.Notify(50073)
			}
			__antithesis_instrumentation__.Notify(50069)

			waitTime := time.Duration(rand.Int63n(int64(time.Second)))
			time.Sleep(waitTime)

			sig := [2]int{2, 9}[rand.Intn(2)]
			stopOpts := option.DefaultStopOpts()
			stopOpts.RoachprodOpts.Sig = sig
			if err := c.StopE(ctx, t.L(), stopOpts, node); err != nil {
				__antithesis_instrumentation__.Notify(50074)
				t.Fatalf("error during stop: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(50075)
			}
		}
		__antithesis_instrumentation__.Notify(50065)

		httpClient := httputil.NewClientWithTimeout(15 * time.Second)

		for !done() {
			__antithesis_instrumentation__.Notify(50076)
			adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), node)
			if err != nil {
				__antithesis_instrumentation__.Notify(50078)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50079)
			}
			__antithesis_instrumentation__.Notify(50077)
			base := `http://` + adminUIAddrs[0]

			url := base + `/_status/vars`
			resp, err := httpClient.Get(ctx, url)
			if err == nil {
				__antithesis_instrumentation__.Notify(50080)
				resp.Body.Close()
				if resp.StatusCode != http.StatusNotFound && func() bool {
					__antithesis_instrumentation__.Notify(50082)
					return resp.StatusCode != http.StatusOK == true
				}() == true {
					__antithesis_instrumentation__.Notify(50083)
					t.Fatalf("unexpected status code from %s: %d", url, resp.StatusCode)
				} else {
					__antithesis_instrumentation__.Notify(50084)
				}
				__antithesis_instrumentation__.Notify(50081)
				break
			} else {
				__antithesis_instrumentation__.Notify(50085)
			}
		}
		__antithesis_instrumentation__.Notify(50066)

		t.L().Printf("%d OK\n", j)
	}
	__antithesis_instrumentation__.Notify(50062)

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), node)
	c.Wipe(ctx, node)
}
