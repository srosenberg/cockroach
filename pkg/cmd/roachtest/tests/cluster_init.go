package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/stretchr/testify/require"
)

func runClusterInit(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(46745)
	c.Put(ctx, t.Cockroach(), "./cockroach")

	t.L().Printf("retrieving VM addresses")
	addrs, err := c.InternalAddr(ctx, t.L(), c.All())
	if err != nil {
		__antithesis_instrumentation__.Notify(46748)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46749)
	}
	__antithesis_instrumentation__.Notify(46746)

	if addrs[0] == "" {
		__antithesis_instrumentation__.Notify(46750)
		t.Fatal("no address for first node")
	} else {
		__antithesis_instrumentation__.Notify(46751)
	}
	__antithesis_instrumentation__.Notify(46747)

	for _, initNode := range []int{2, 1} {
		__antithesis_instrumentation__.Notify(46752)
		c.Wipe(ctx)
		t.L().Printf("starting test with init node %d", initNode)
		startOpts := option.DefaultStartOpts()

		startOpts.RoachprodOpts.SkipInit = true

		startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--join="+strings.Join(addrs, ","))
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings())

		urlMap := make(map[int]string)
		adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.All())
		if err != nil {
			__antithesis_instrumentation__.Notify(46763)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46764)
		}
		__antithesis_instrumentation__.Notify(46753)
		for i, addr := range adminUIAddrs {
			__antithesis_instrumentation__.Notify(46765)
			urlMap[i+1] = `http://` + addr
		}
		__antithesis_instrumentation__.Notify(46754)

		t.L().Printf("waiting for the servers to bind their ports")
		if err := retry.ForDuration(10*time.Second, func() error {
			__antithesis_instrumentation__.Notify(46766)
			for i := 1; i <= c.Spec().NodeCount; i++ {
				__antithesis_instrumentation__.Notify(46768)
				resp, err := httputil.Get(ctx, urlMap[i]+"/health")
				if err != nil {
					__antithesis_instrumentation__.Notify(46770)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46771)
				}
				__antithesis_instrumentation__.Notify(46769)
				resp.Body.Close()
			}
			__antithesis_instrumentation__.Notify(46767)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(46772)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46773)
		}
		__antithesis_instrumentation__.Notify(46755)
		t.L().Printf("all nodes started, establishing SQL connections")

		var dbs []*gosql.DB
		for i := 1; i <= c.Spec().NodeCount; i++ {
			__antithesis_instrumentation__.Notify(46774)
			db := c.Conn(ctx, t.L(), i)
			defer db.Close()
			dbs = append(dbs, db)
		}
		__antithesis_instrumentation__.Notify(46756)

		t.L().Printf("checking that the SQL conns are not failing immediately")
		errCh := make(chan error, len(dbs))
		for _, db := range dbs {
			__antithesis_instrumentation__.Notify(46775)
			db := db
			go func() {
				__antithesis_instrumentation__.Notify(46776)
				var val int
				errCh <- db.QueryRow("SELECT 1").Scan(&val)
			}()
		}
		__antithesis_instrumentation__.Notify(46757)

		time.Sleep(time.Second)
		select {
		case err := <-errCh:
			__antithesis_instrumentation__.Notify(46777)
			t.Fatalf("query finished prematurely with err %v", err)
		default:
			__antithesis_instrumentation__.Notify(46778)
		}
		__antithesis_instrumentation__.Notify(46758)

		httpTests := []struct {
			endpoint       string
			expectedStatus int
		}{
			{"/health", http.StatusOK},
			{"/health?ready=1", http.StatusServiceUnavailable},
			{"/_status/nodes", http.StatusNotFound},
		}
		for _, tc := range httpTests {
			__antithesis_instrumentation__.Notify(46779)
			for _, withCookie := range []bool{false, true} {
				__antithesis_instrumentation__.Notify(46780)
				t.L().Printf("checking for HTTP endpoint %q, using authentication = %v", tc.endpoint, withCookie)
				req, err := http.NewRequest("GET", urlMap[1]+tc.endpoint, nil)
				if err != nil {
					__antithesis_instrumentation__.Notify(46784)
					t.Fatalf("unexpected error while constructing request for %s: %s", tc.endpoint, err)
				} else {
					__antithesis_instrumentation__.Notify(46785)
				}
				__antithesis_instrumentation__.Notify(46781)
				if withCookie {
					__antithesis_instrumentation__.Notify(46786)

					cookie, err := server.EncodeSessionCookie(&serverpb.SessionCookie{}, false)
					if err != nil {
						__antithesis_instrumentation__.Notify(46788)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(46789)
					}
					__antithesis_instrumentation__.Notify(46787)
					req.AddCookie(cookie)
				} else {
					__antithesis_instrumentation__.Notify(46790)
				}
				__antithesis_instrumentation__.Notify(46782)
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					__antithesis_instrumentation__.Notify(46791)
					t.Fatalf("unexpected error hitting %s endpoint: %v", tc.endpoint, err)
				} else {
					__antithesis_instrumentation__.Notify(46792)
				}
				__antithesis_instrumentation__.Notify(46783)
				defer resp.Body.Close()
				if resp.StatusCode != tc.expectedStatus {
					__antithesis_instrumentation__.Notify(46793)
					bodyBytes, _ := ioutil.ReadAll(resp.Body)
					t.Fatalf("unexpected response code %d (expected %d) hitting %s endpoint: %v",
						resp.StatusCode, tc.expectedStatus, tc.endpoint, string(bodyBytes))
				} else {
					__antithesis_instrumentation__.Notify(46794)
				}
			}

		}
		__antithesis_instrumentation__.Notify(46759)

		t.L().Printf("sending init command to node %d", initNode)
		c.Run(ctx, c.Node(initNode),
			fmt.Sprintf(`./cockroach init --insecure --port={pgport:%d}`, initNode))

		err = WaitFor3XReplication(ctx, t, dbs[0])
		require.NoError(t, err)

		execCLI := func(runNode int, extraArgs ...string) (string, error) {
			__antithesis_instrumentation__.Notify(46795)
			args := []string{"./cockroach"}
			args = append(args, extraArgs...)
			args = append(args, "--insecure")
			args = append(args, fmt.Sprintf("--port={pgport:%d}", runNode))
			result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(runNode), args...)
			combinedOutput := result.Stdout + result.Stderr
			t.L().Printf("%s\n", combinedOutput)
			return combinedOutput, err
		}

		{
			__antithesis_instrumentation__.Notify(46796)
			t.L().Printf("checking that double init fails")

			if output, err := execCLI(initNode, "init"); err == nil {
				__antithesis_instrumentation__.Notify(46797)
				t.Fatalf("expected error running init command on initialized cluster\n%s", output)
			} else {
				__antithesis_instrumentation__.Notify(46798)
				if !strings.Contains(output, "cluster has already been initialized") {
					__antithesis_instrumentation__.Notify(46799)
					t.Fatalf("unexpected output when running init command on initialized cluster: %v\n%s",
						err, output)
				} else {
					__antithesis_instrumentation__.Notify(46800)
				}
			}
		}
		__antithesis_instrumentation__.Notify(46760)

		t.L().Printf("waiting for original SQL queries to complete now cluster is initialized")
		deadline := time.After(10 * time.Second)
		for i := 0; i < len(dbs); i++ {
			__antithesis_instrumentation__.Notify(46801)
			select {
			case err := <-errCh:
				__antithesis_instrumentation__.Notify(46802)
				if err != nil {
					__antithesis_instrumentation__.Notify(46804)
					t.Fatalf("querying node %d: %s", i, err)
				} else {
					__antithesis_instrumentation__.Notify(46805)
				}
			case <-deadline:
				__antithesis_instrumentation__.Notify(46803)
				t.Fatalf("timed out waiting for query %d", i)
			}
		}
		__antithesis_instrumentation__.Notify(46761)

		t.L().Printf("testing new SQL queries")
		for i, db := range dbs {
			__antithesis_instrumentation__.Notify(46806)
			var val int
			if err := db.QueryRow("SELECT 1").Scan(&val); err != nil {
				__antithesis_instrumentation__.Notify(46807)
				t.Fatalf("querying node %d: %s", i, err)
			} else {
				__antithesis_instrumentation__.Notify(46808)
			}
		}
		__antithesis_instrumentation__.Notify(46762)

		t.L().Printf("test complete")
	}
}
