package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type HealthChecker struct {
	t      test.Test
	c      cluster.Cluster
	nodes  option.NodeListOption
	doneCh chan struct{}
}

func NewHealthChecker(t test.Test, c cluster.Cluster, nodes option.NodeListOption) *HealthChecker {
	__antithesis_instrumentation__.Notify(50273)
	return &HealthChecker{
		t:      t,
		c:      c,
		nodes:  nodes,
		doneCh: make(chan struct{}),
	}
}

func (hc *HealthChecker) Done() {
	__antithesis_instrumentation__.Notify(50274)
	close(hc.doneCh)
}

type gossipAlert struct {
	NodeID, StoreID       int
	Category, Description string
	Value                 float64
}

type gossipAlerts []gossipAlert

func (g gossipAlerts) String() string {
	__antithesis_instrumentation__.Notify(50275)
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)

	for _, a := range g {
		__antithesis_instrumentation__.Notify(50277)
		fmt.Fprintf(tw, "n%d/s%d\t%.2f\t%s\t%s\n", a.NodeID, a.StoreID, a.Value, a.Category, a.Description)
	}
	__antithesis_instrumentation__.Notify(50276)
	_ = tw.Flush()
	return buf.String()
}

func (hc *HealthChecker) Runner(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(50278)
	logger, err := hc.t.L().ChildLogger("health")
	if err != nil {
		__antithesis_instrumentation__.Notify(50281)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50282)
	}
	__antithesis_instrumentation__.Notify(50279)
	defer func() {
		__antithesis_instrumentation__.Notify(50283)
		logger.Printf("health check terminated with %v\n", err)
	}()
	__antithesis_instrumentation__.Notify(50280)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		__antithesis_instrumentation__.Notify(50284)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(50291)
			return ctx.Err()
		case <-hc.doneCh:
			__antithesis_instrumentation__.Notify(50292)
			return nil
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(50293)
		}
		__antithesis_instrumentation__.Notify(50285)

		tBegin := timeutil.Now()

		nodeIdx := 1 + rand.Intn(len(hc.nodes))
		db, err := hc.c.ConnE(ctx, hc.t.L(), nodeIdx)
		if err != nil {
			__antithesis_instrumentation__.Notify(50294)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50295)
		}
		__antithesis_instrumentation__.Notify(50286)

		_, err = db.Exec(`USE system`)
		if err != nil {
			__antithesis_instrumentation__.Notify(50296)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50297)
		}
		__antithesis_instrumentation__.Notify(50287)
		rows, err := db.QueryContext(ctx, `SELECT * FROM crdb_internal.gossip_alerts ORDER BY node_id ASC, store_id ASC `)
		_ = db.Close()
		if err != nil {
			__antithesis_instrumentation__.Notify(50298)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50299)
		}
		__antithesis_instrumentation__.Notify(50288)
		var rr gossipAlerts
		for rows.Next() {
			__antithesis_instrumentation__.Notify(50300)
			a := gossipAlert{StoreID: -1}
			var storeID gosql.NullInt64
			if err := rows.Scan(&a.NodeID, &storeID, &a.Category, &a.Description, &a.Value); err != nil {
				__antithesis_instrumentation__.Notify(50303)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50304)
			}
			__antithesis_instrumentation__.Notify(50301)
			if storeID.Valid {
				__antithesis_instrumentation__.Notify(50305)
				a.StoreID = int(storeID.Int64)
			} else {
				__antithesis_instrumentation__.Notify(50306)
			}
			__antithesis_instrumentation__.Notify(50302)
			rr = append(rr, a)
		}
		__antithesis_instrumentation__.Notify(50289)
		if len(rr) > 0 {
			__antithesis_instrumentation__.Notify(50307)
			logger.Printf(rr.String() + "\n")

		} else {
			__antithesis_instrumentation__.Notify(50308)
		}
		__antithesis_instrumentation__.Notify(50290)

		if elapsed := timeutil.Since(tBegin); elapsed > 10*time.Second {
			__antithesis_instrumentation__.Notify(50309)
			err := errors.Errorf("health check against node %d took %s", nodeIdx, elapsed)
			logger.Printf("%+v", err)

		} else {
			__antithesis_instrumentation__.Notify(50310)
		}
	}
}

type DiskUsageLogger struct {
	t      test.Test
	c      cluster.Cluster
	doneCh chan struct{}
}

func NewDiskUsageLogger(t test.Test, c cluster.Cluster) *DiskUsageLogger {
	__antithesis_instrumentation__.Notify(50311)
	return &DiskUsageLogger{
		t:      t,
		c:      c,
		doneCh: make(chan struct{}),
	}
}

func (dul *DiskUsageLogger) Done() {
	__antithesis_instrumentation__.Notify(50312)
	close(dul.doneCh)
}

func (dul *DiskUsageLogger) Runner(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(50313)
	l, err := dul.t.L().ChildLogger("diskusage")
	if err != nil {
		__antithesis_instrumentation__.Notify(50316)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50317)
	}
	__antithesis_instrumentation__.Notify(50314)
	quietLogger, err := dul.t.L().ChildLogger("diskusage-exec", logger.QuietStdout, logger.QuietStderr)
	if err != nil {
		__antithesis_instrumentation__.Notify(50318)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50319)
	}
	__antithesis_instrumentation__.Notify(50315)

	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		__antithesis_instrumentation__.Notify(50320)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(50325)
			return ctx.Err()
		case <-dul.doneCh:
			__antithesis_instrumentation__.Notify(50326)
			return nil
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(50327)
		}
		__antithesis_instrumentation__.Notify(50321)

		type usage struct {
			nodeNum int
			bytes   int
		}

		var bytesUsed []usage
		for i := 1; i <= dul.c.Spec().NodeCount; i++ {
			__antithesis_instrumentation__.Notify(50328)
			cur, err := getDiskUsageInBytes(ctx, dul.c, quietLogger, i)
			if err != nil {
				__antithesis_instrumentation__.Notify(50330)

				l.Printf("%s", errors.Wrapf(err, "node #%d", i))
				cur = -1
			} else {
				__antithesis_instrumentation__.Notify(50331)
			}
			__antithesis_instrumentation__.Notify(50329)
			bytesUsed = append(bytesUsed, usage{
				nodeNum: i,
				bytes:   cur,
			})
		}
		__antithesis_instrumentation__.Notify(50322)
		sort.Slice(bytesUsed, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(50332)
			return bytesUsed[i].bytes > bytesUsed[j].bytes
		})
		__antithesis_instrumentation__.Notify(50323)

		var s []string
		for _, usage := range bytesUsed {
			__antithesis_instrumentation__.Notify(50333)
			s = append(s, fmt.Sprintf("n#%d: %s", usage.nodeNum, humanizeutil.IBytes(int64(usage.bytes))))
		}
		__antithesis_instrumentation__.Notify(50324)

		l.Printf("%s\n", strings.Join(s, ", "))
	}
}
func registerRestoreNodeShutdown(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50334)
	makeRestoreStarter := func(ctx context.Context, t test.Test, c cluster.Cluster, gatewayNode int) jobStarter {
		__antithesis_instrumentation__.Notify(50337)
		return func(c cluster.Cluster, t test.Test) (string, error) {
			__antithesis_instrumentation__.Notify(50338)
			t.L().Printf("connecting to gateway")
			gatewayDB := c.Conn(ctx, t.L(), gatewayNode)
			defer gatewayDB.Close()

			t.L().Printf("creating bank database")
			if _, err := gatewayDB.Exec("CREATE DATABASE bank"); err != nil {
				__antithesis_instrumentation__.Notify(50343)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(50344)
			}
			__antithesis_instrumentation__.Notify(50339)

			errCh := make(chan error, 1)
			go func() {
				__antithesis_instrumentation__.Notify(50345)
				defer close(errCh)

				restoreQuery := `RESTORE bank.bank FROM
					'gs://cockroach-fixtures/workload/bank/version=1.0.0,payload-bytes=100,ranges=10,rows=10000000,seed=1/bank?AUTH=implicit'`

				t.L().Printf("starting to run the restore job")
				if _, err := gatewayDB.Exec(restoreQuery); err != nil {
					__antithesis_instrumentation__.Notify(50347)
					errCh <- err
				} else {
					__antithesis_instrumentation__.Notify(50348)
				}
				__antithesis_instrumentation__.Notify(50346)
				t.L().Printf("done running restore job")
			}()
			__antithesis_instrumentation__.Notify(50340)

			retryOpts := retry.Options{
				MaxRetries:     50,
				InitialBackoff: 1 * time.Second,
				MaxBackoff:     5 * time.Second,
			}
			for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
				__antithesis_instrumentation__.Notify(50349)
				var jobCount int
				if err := gatewayDB.QueryRowContext(ctx, "SELECT count(*) FROM [SHOW JOBS] WHERE job_type = 'RESTORE'").Scan(&jobCount); err != nil {
					__antithesis_instrumentation__.Notify(50352)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(50353)
				}
				__antithesis_instrumentation__.Notify(50350)

				select {
				case err := <-errCh:
					__antithesis_instrumentation__.Notify(50354)

					return "", err
				default:
					__antithesis_instrumentation__.Notify(50355)
				}
				__antithesis_instrumentation__.Notify(50351)

				if jobCount == 0 {
					__antithesis_instrumentation__.Notify(50356)
					t.L().Printf("waiting for restore job")
				} else {
					__antithesis_instrumentation__.Notify(50357)
					if jobCount == 1 {
						__antithesis_instrumentation__.Notify(50358)
						t.L().Printf("found restore job")
						break
					} else {
						__antithesis_instrumentation__.Notify(50359)
						t.L().Printf("found multiple restore jobs -- erroring")
						return "", errors.New("unexpectedly found multiple restore jobs")
					}
				}
			}
			__antithesis_instrumentation__.Notify(50341)

			var jobID string
			if err := gatewayDB.QueryRowContext(ctx, "SELECT job_id FROM [SHOW JOBS] WHERE job_type = 'RESTORE'").Scan(&jobID); err != nil {
				__antithesis_instrumentation__.Notify(50360)
				return "", errors.Wrap(err, "querying the job ID")
			} else {
				__antithesis_instrumentation__.Notify(50361)
			}
			__antithesis_instrumentation__.Notify(50342)
			return jobID, nil
		}
	}
	__antithesis_instrumentation__.Notify(50335)

	r.Add(registry.TestSpec{
		Name:    "restore/nodeShutdown/worker",
		Owner:   registry.OwnerBulkIO,
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50362)
			gatewayNode := 2
			nodeToShutdown := 3
			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

			jobSurvivesNodeShutdown(ctx, t, c, nodeToShutdown, makeRestoreStarter(ctx, t, c, gatewayNode))
		},
	})
	__antithesis_instrumentation__.Notify(50336)

	r.Add(registry.TestSpec{
		Name:    "restore/nodeShutdown/coordinator",
		Owner:   registry.OwnerBulkIO,
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50363)
			gatewayNode := 2
			nodeToShutdown := 2
			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

			jobSurvivesNodeShutdown(ctx, t, c, nodeToShutdown, makeRestoreStarter(ctx, t, c, gatewayNode))
		},
	})
}

type testDataSet interface {
	name() string

	runRestore(ctx context.Context, c cluster.Cluster)
}

type dataBank2TB struct{}

func (dataBank2TB) name() string {
	__antithesis_instrumentation__.Notify(50364)
	return "2TB"
}

func (dataBank2TB) runRestore(ctx context.Context, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(50365)
	c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "CREATE DATABASE restore2tb"`)
	c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "
				RESTORE csv.bank FROM
				'gs://cockroach-fixtures/workload/bank/version=1.0.0,payload-bytes=10240,ranges=0,rows=65104166,seed=1/bank?AUTH=implicit'
				WITH into_db = 'restore2tb'"`)
}

type tpccIncData struct{}

func (tpccIncData) name() string {
	__antithesis_instrumentation__.Notify(50366)
	return "TPCCInc"
}

func (tpccIncData) runRestore(ctx context.Context, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(50367)

	c.Run(ctx, c.Node(1), `./cockroach sql --insecure -e "
				RESTORE FROM '2021/05/21-020411.00' IN
				'gs://cockroach-fixtures/tpcc-incrementals?AUTH=implicit'
				AS OF SYSTEM TIME '2021-05-21 14:40:22'"`)
}

func registerRestore(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50368)
	largeVolumeSize := 2500

	for _, item := range []struct {
		nodes        int
		cpus         int
		largeVolumes bool
		dataSet      testDataSet

		timeout time.Duration
	}{
		{dataSet: dataBank2TB{}, nodes: 10, timeout: 6 * time.Hour},
		{dataSet: dataBank2TB{}, nodes: 32, timeout: 3 * time.Hour},
		{dataSet: dataBank2TB{}, nodes: 6, timeout: 4 * time.Hour, cpus: 8, largeVolumes: true},
		{dataSet: tpccIncData{}, nodes: 10, timeout: 6 * time.Hour},
	} {
		__antithesis_instrumentation__.Notify(50369)
		item := item
		clusterOpts := make([]spec.Option, 0)
		testName := fmt.Sprintf("restore%s/nodes=%d", item.dataSet.name(), item.nodes)
		if item.cpus != 0 {
			__antithesis_instrumentation__.Notify(50372)
			clusterOpts = append(clusterOpts, spec.CPU(item.cpus))
			testName += fmt.Sprintf("/cpus=%d", item.cpus)
		} else {
			__antithesis_instrumentation__.Notify(50373)
		}
		__antithesis_instrumentation__.Notify(50370)
		if item.largeVolumes {
			__antithesis_instrumentation__.Notify(50374)
			clusterOpts = append(clusterOpts, spec.VolumeSize(largeVolumeSize))
			testName += fmt.Sprintf("/pd-volume=%dGB", largeVolumeSize)
		} else {
			__antithesis_instrumentation__.Notify(50375)
		}
		__antithesis_instrumentation__.Notify(50371)

		r.Add(registry.TestSpec{
			Name:            testName,
			Owner:           registry.OwnerBulkIO,
			Cluster:         r.MakeClusterSpec(item.nodes, clusterOpts...),
			Timeout:         item.timeout,
			EncryptAtRandom: true,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(50376)
				c.Put(ctx, t.Cockroach(), "./cockroach")
				c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
				m := c.NewMonitor(ctx)

				dul := NewDiskUsageLogger(t, c)
				m.Go(dul.Runner)
				hc := NewHealthChecker(t, c, c.All())
				m.Go(hc.Runner)

				tick, perfBuf := initBulkJobPerfArtifacts(testName, item.timeout)
				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(50378)
					defer dul.Done()
					defer hc.Done()
					t.Status(`running restore`)

					if item.cpus >= 8 {
						__antithesis_instrumentation__.Notify(50382)

						c.Run(ctx, c.Node(1),
							`./cockroach sql --insecure -e "SET CLUSTER SETTING kv.bulk_io_write.restore_node_concurrency = 5"`)
						c.Run(ctx, c.Node(1),
							`./cockroach sql --insecure -e "SET CLUSTER SETTING kv.bulk_io_write.concurrent_addsstable_requests = 5"`)
					} else {
						__antithesis_instrumentation__.Notify(50383)
					}
					__antithesis_instrumentation__.Notify(50379)
					tick()
					item.dataSet.runRestore(ctx, c)
					tick()

					dest := filepath.Join(t.PerfArtifactsDir(), "stats.json")
					if err := c.RunE(ctx, c.Node(1), "mkdir -p "+filepath.Dir(dest)); err != nil {
						__antithesis_instrumentation__.Notify(50384)
						log.Errorf(ctx, "failed to create perf dir: %+v", err)
					} else {
						__antithesis_instrumentation__.Notify(50385)
					}
					__antithesis_instrumentation__.Notify(50380)
					if err := c.PutString(ctx, perfBuf.String(), dest, 0755, c.Node(1)); err != nil {
						__antithesis_instrumentation__.Notify(50386)
						log.Errorf(ctx, "failed to upload perf artifacts to node: %s", err.Error())
					} else {
						__antithesis_instrumentation__.Notify(50387)
					}
					__antithesis_instrumentation__.Notify(50381)
					return nil
				})
				__antithesis_instrumentation__.Notify(50377)
				m.Wait()
			},
		})
	}
}

func verifyMetrics(
	ctx context.Context, t test.Test, c cluster.Cluster, m map[string]float64,
) error {
	__antithesis_instrumentation__.Notify(50388)
	const sample = 10 * time.Second

	adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(1))
	if err != nil {
		__antithesis_instrumentation__.Notify(50391)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50392)
	}
	__antithesis_instrumentation__.Notify(50389)
	url := "http://" + adminUIAddrs[0] + "/ts/query"

	request := tspb.TimeSeriesQueryRequest{

		SampleNanos: sample.Nanoseconds(),
	}
	for name := range m {
		__antithesis_instrumentation__.Notify(50393)
		request.Queries = append(request.Queries, tspb.Query{
			Name:             name,
			Downsampler:      tspb.TimeSeriesQueryAggregator_AVG.Enum(),
			SourceAggregator: tspb.TimeSeriesQueryAggregator_SUM.Enum(),
		})
	}
	__antithesis_instrumentation__.Notify(50390)

	ticker := time.NewTicker(sample)
	defer ticker.Stop()
	for {
		__antithesis_instrumentation__.Notify(50394)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(50397)
			return ctx.Err()
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(50398)
		}
		__antithesis_instrumentation__.Notify(50395)

		now := timeutil.Now()
		request.StartNanos = now.Add(-sample * 3).UnixNano()
		request.EndNanos = now.UnixNano()

		var response tspb.TimeSeriesQueryResponse
		if err := httputil.PostJSON(http.Client{}, url, &request, &response); err != nil {
			__antithesis_instrumentation__.Notify(50399)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50400)
		}
		__antithesis_instrumentation__.Notify(50396)

		for i := range request.Queries {
			__antithesis_instrumentation__.Notify(50401)
			name := request.Queries[i].Name
			data := response.Results[i].Datapoints
			n := len(data)
			if n == 0 {
				__antithesis_instrumentation__.Notify(50403)
				continue
			} else {
				__antithesis_instrumentation__.Notify(50404)
			}
			__antithesis_instrumentation__.Notify(50402)
			limit := m[name]
			value := data[n-1].Value
			if value >= limit {
				__antithesis_instrumentation__.Notify(50405)
				return fmt.Errorf("%s: %.1f >= %.1f @ %d", name, value, limit, data[n-1].TimestampNanos)
			} else {
				__antithesis_instrumentation__.Notify(50406)
			}
		}
	}
}

var _ = verifyMetrics
