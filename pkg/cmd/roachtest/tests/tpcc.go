package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/prometheus"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/util/search"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/cockroach/pkg/workload/tpcc"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/ttycolor"
	"github.com/lib/pq"
	promapi "github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/require"
)

type tpccSetupType int

const (
	usingImport tpccSetupType = iota
	usingInit
	usingExistingData
)

type tpccOptions struct {
	Warehouses     int
	ExtraRunArgs   string
	ExtraSetupArgs string
	Chaos          func() Chaos
	During         func(context.Context) error
	Duration       time.Duration
	SetupType      tpccSetupType

	PrometheusConfig *prometheus.Config

	DisablePrometheus bool

	WorkloadInstances []workloadInstance

	ChaosEventsProcessor func(
		prometheusNode option.NodeListOption,
		workloadInstances []workloadInstance,
	) (tpccChaosEventProcessor, error)

	Start func(context.Context, test.Test, cluster.Cluster)

	EnableCircuitBreakers bool
}

type workloadInstance struct {
	nodes option.NodeListOption

	prometheusPort int

	extraRunArgs string
}

const workloadPProfStartPort = 33333
const workloadPrometheusPort = 2112

func tpccImportCmd(warehouses int, extraArgs ...string) string {
	__antithesis_instrumentation__.Notify(51490)
	return tpccImportCmdWithCockroachBinary("./cockroach", warehouses, extraArgs...)
}

func tpccImportCmdWithCockroachBinary(
	crdbBinary string, warehouses int, extraArgs ...string,
) string {
	__antithesis_instrumentation__.Notify(51491)
	return fmt.Sprintf("./%s workload fixtures import tpcc --warehouses=%d %s",
		crdbBinary, warehouses, strings.Join(extraArgs, " "))
}

func setupTPCC(
	ctx context.Context, t test.Test, c cluster.Cluster, opts tpccOptions,
) (crdbNodes, workloadNode option.NodeListOption) {
	__antithesis_instrumentation__.Notify(51492)

	crdbNodes = c.Range(1, c.Spec().NodeCount-1)
	workloadNode = c.Node(c.Spec().NodeCount)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(51496)
		opts.Warehouses = 1
	} else {
		__antithesis_instrumentation__.Notify(51497)
	}
	__antithesis_instrumentation__.Notify(51493)

	if opts.Start == nil {
		__antithesis_instrumentation__.Notify(51498)
		opts.Start = func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51499)

			c.Put(ctx, t.Cockroach(), "./cockroach", c.All())

			c.Put(ctx, t.DeprecatedWorkload(), "./workload", workloadNode)
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)
		}
	} else {
		__antithesis_instrumentation__.Notify(51500)
	}
	__antithesis_instrumentation__.Notify(51494)

	func() {
		__antithesis_instrumentation__.Notify(51501)
		opts.Start(ctx, t, c)
		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()
		if opts.EnableCircuitBreakers {
			__antithesis_instrumentation__.Notify(51504)
			_, err := db.Exec(`SET CLUSTER SETTING kv.replica_circuit_breaker.slow_replication_threshold = '15s'`)
			require.NoError(t, err)
		} else {
			__antithesis_instrumentation__.Notify(51505)
		}
		__antithesis_instrumentation__.Notify(51502)
		err := WaitFor3XReplication(ctx, t, c.Conn(ctx, t.L(), crdbNodes[0]))
		require.NoError(t, err)
		switch opts.SetupType {
		case usingExistingData:
			__antithesis_instrumentation__.Notify(51506)

		case usingImport:
			__antithesis_instrumentation__.Notify(51507)
			t.Status("loading fixture")
			c.Run(ctx, crdbNodes[:1], tpccImportCmd(opts.Warehouses, opts.ExtraSetupArgs))
		case usingInit:
			__antithesis_instrumentation__.Notify(51508)
			t.Status("initializing tables")
			extraArgs := opts.ExtraSetupArgs
			if !t.BuildVersion().AtLeast(version.MustParse("v20.2.0")) {
				__antithesis_instrumentation__.Notify(51511)
				extraArgs += " --deprecated-fk-indexes"
			} else {
				__antithesis_instrumentation__.Notify(51512)
			}
			__antithesis_instrumentation__.Notify(51509)
			cmd := fmt.Sprintf(
				"./workload init tpcc --warehouses=%d %s {pgurl:1}",
				opts.Warehouses, extraArgs,
			)
			c.Run(ctx, workloadNode, cmd)
		default:
			__antithesis_instrumentation__.Notify(51510)
			t.Fatal("unknown tpcc setup type")
		}
		__antithesis_instrumentation__.Notify(51503)
		t.Status("")
	}()
	__antithesis_instrumentation__.Notify(51495)
	return crdbNodes, workloadNode
}

func runTPCC(ctx context.Context, t test.Test, c cluster.Cluster, opts tpccOptions) {
	__antithesis_instrumentation__.Notify(51513)
	workloadInstances := opts.WorkloadInstances
	if len(workloadInstances) == 0 {
		__antithesis_instrumentation__.Notify(51521)
		workloadInstances = append(
			workloadInstances,
			workloadInstance{
				nodes:          c.Range(1, c.Spec().NodeCount-1),
				prometheusPort: workloadPrometheusPort,
			},
		)
	} else {
		__antithesis_instrumentation__.Notify(51522)
	}
	__antithesis_instrumentation__.Notify(51514)
	var pgURLs []string
	for _, workloadInstance := range workloadInstances {
		__antithesis_instrumentation__.Notify(51523)
		pgURLs = append(pgURLs, fmt.Sprintf("{pgurl%s}", workloadInstance.nodes.String()))
	}
	__antithesis_instrumentation__.Notify(51515)

	var ep *tpccChaosEventProcessor
	promCfg, cleanupFunc := setupPrometheus(ctx, t, c, opts, workloadInstances)
	defer cleanupFunc()
	if opts.ChaosEventsProcessor != nil {
		__antithesis_instrumentation__.Notify(51524)
		if promCfg == nil {
			__antithesis_instrumentation__.Notify(51527)
			t.Skip("skipping test as prometheus is needed, but prometheus does not yet work locally")
			return
		} else {
			__antithesis_instrumentation__.Notify(51528)
		}
		__antithesis_instrumentation__.Notify(51525)
		cep, err := opts.ChaosEventsProcessor(
			promCfg.PrometheusNode,
			workloadInstances,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(51529)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51530)
		}
		__antithesis_instrumentation__.Notify(51526)
		cep.listen(ctx, t.L())
		ep = &cep
	} else {
		__antithesis_instrumentation__.Notify(51531)
	}
	__antithesis_instrumentation__.Notify(51516)

	rampDuration := 5 * time.Minute
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(51532)
		opts.Warehouses = 1
		if opts.Duration > time.Minute {
			__antithesis_instrumentation__.Notify(51534)
			opts.Duration = time.Minute
		} else {
			__antithesis_instrumentation__.Notify(51535)
		}
		__antithesis_instrumentation__.Notify(51533)
		rampDuration = 30 * time.Second
	} else {
		__antithesis_instrumentation__.Notify(51536)
	}
	__antithesis_instrumentation__.Notify(51517)
	crdbNodes, workloadNode := setupTPCC(ctx, t, c, opts)
	t.Status("waiting")
	m := c.NewMonitor(ctx, crdbNodes)
	for i := range workloadInstances {
		__antithesis_instrumentation__.Notify(51537)

		i := i
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(51538)

			var statsPrefix string
			if len(workloadInstances) > 1 {
				__antithesis_instrumentation__.Notify(51540)
				statsPrefix = fmt.Sprintf("workload_%d.", i)
			} else {
				__antithesis_instrumentation__.Notify(51541)
			}
			__antithesis_instrumentation__.Notify(51539)
			t.WorkerStatus(fmt.Sprintf("running tpcc idx %d on %s", i, pgURLs[i]))
			cmd := fmt.Sprintf(
				"./cockroach workload run tpcc --warehouses=%d --histograms="+t.PerfArtifactsDir()+"/%sstats.json "+
					opts.ExtraRunArgs+" --ramp=%s --duration=%s --prometheus-port=%d --pprofport=%d %s %s",
				opts.Warehouses,
				statsPrefix,
				rampDuration,
				opts.Duration,
				workloadInstances[i].prometheusPort,
				workloadPProfStartPort+i,
				workloadInstances[i].extraRunArgs,
				pgURLs[i],
			)
			return c.RunE(ctx, workloadNode, cmd)
		})
	}
	__antithesis_instrumentation__.Notify(51518)
	if opts.Chaos != nil {
		__antithesis_instrumentation__.Notify(51542)
		chaos := opts.Chaos()
		m.Go(chaos.Runner(c, t, m))
	} else {
		__antithesis_instrumentation__.Notify(51543)
	}
	__antithesis_instrumentation__.Notify(51519)
	if opts.During != nil {
		__antithesis_instrumentation__.Notify(51544)
		m.Go(opts.During)
	} else {
		__antithesis_instrumentation__.Notify(51545)
	}
	__antithesis_instrumentation__.Notify(51520)
	m.Wait()

	c.Run(ctx, workloadNode, fmt.Sprintf(
		"./cockroach workload check tpcc --warehouses=%d {pgurl:1}", opts.Warehouses))

	if ep != nil {
		__antithesis_instrumentation__.Notify(51546)
		if err := ep.err(); err != nil {
			__antithesis_instrumentation__.Notify(51547)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51548)
		}
	} else {
		__antithesis_instrumentation__.Notify(51549)
	}
}

var tpccSupportedWarehouses = []struct {
	hardware   string
	v          *version.Version
	warehouses int
}{

	{hardware: "gce-n4cpu16", v: version.MustParse(`v2.1.0-0`), warehouses: 1300},
	{hardware: "gce-n4cpu16", v: version.MustParse(`v19.1.0-0`), warehouses: 1250},
	{hardware: "aws-n4cpu16", v: version.MustParse(`v19.1.0-0`), warehouses: 2100},

	{hardware: "gce-n5cpu16", v: version.MustParse(`v19.1.0-0`), warehouses: 1300},

	{hardware: "gce-n5cpu16", v: version.MustParse(`v2.1.0-0`), warehouses: 1300},
}

func maxSupportedTPCCWarehouses(
	buildVersion version.Version, cloud string, nodes spec.ClusterSpec,
) int {
	__antithesis_instrumentation__.Notify(51550)
	var v *version.Version
	var warehouses int
	hardware := fmt.Sprintf(`%s-%s`, cloud, &nodes)
	for _, x := range tpccSupportedWarehouses {
		__antithesis_instrumentation__.Notify(51553)
		if x.hardware != hardware {
			__antithesis_instrumentation__.Notify(51555)
			continue
		} else {
			__antithesis_instrumentation__.Notify(51556)
		}
		__antithesis_instrumentation__.Notify(51554)
		if buildVersion.AtLeast(x.v) && func() bool {
			__antithesis_instrumentation__.Notify(51557)
			return (v == nil || func() bool {
				__antithesis_instrumentation__.Notify(51558)
				return buildVersion.AtLeast(v) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(51559)
			v = x.v
			warehouses = x.warehouses
		} else {
			__antithesis_instrumentation__.Notify(51560)
		}
	}
	__antithesis_instrumentation__.Notify(51551)
	if v == nil {
		__antithesis_instrumentation__.Notify(51561)
		panic(fmt.Sprintf(`could not find max tpcc warehouses for %s`, hardware))
	} else {
		__antithesis_instrumentation__.Notify(51562)
	}
	__antithesis_instrumentation__.Notify(51552)
	return warehouses
}

func registerTPCC(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51563)
	cloud := r.MakeClusterSpec(1).Cloud
	headroomSpec := r.MakeClusterSpec(4, spec.CPU(16), spec.RandomlyUseZfs())
	r.Add(registry.TestSpec{

		Name:            "tpcc/headroom/" + headroomSpec.String(),
		Owner:           registry.OwnerKV,
		Tags:            []string{`default`, `release_qualification`},
		Cluster:         headroomSpec,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51570)
			maxWarehouses := maxSupportedTPCCWarehouses(*t.BuildVersion(), cloud, t.Spec().(*registry.TestSpec).Cluster)
			headroomWarehouses := int(float64(maxWarehouses) * 0.7)
			t.L().Printf("computed headroom warehouses of %d\n", headroomWarehouses)
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses:            headroomWarehouses,
				Duration:              120 * time.Minute,
				SetupType:             usingImport,
				EnableCircuitBreakers: true,
			})
		},
	})
	__antithesis_instrumentation__.Notify(51564)
	mixedHeadroomSpec := r.MakeClusterSpec(5, spec.CPU(16), spec.RandomlyUseZfs())

	r.Add(registry.TestSpec{

		Name:  "tpcc/mixed-headroom/" + mixedHeadroomSpec.String(),
		Owner: registry.OwnerKV,

		Tags:            []string{`default`},
		Cluster:         mixedHeadroomSpec,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51571)
			crdbNodes := c.Range(1, 4)
			workloadNode := c.Node(5)

			maxWarehouses := maxSupportedTPCCWarehouses(*t.BuildVersion(), cloud, t.Spec().(*registry.TestSpec).Cluster)
			headroomWarehouses := int(float64(maxWarehouses) * 0.7)
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(51576)
				headroomWarehouses = 10
			} else {
				__antithesis_instrumentation__.Notify(51577)
			}
			__antithesis_instrumentation__.Notify(51572)

			tpccBackgroundStepper := backgroundStepper{
				nodes: crdbNodes,
				run: func(ctx context.Context, u *versionUpgradeTest) error {
					__antithesis_instrumentation__.Notify(51578)
					const duration = 120 * time.Minute
					t.L().Printf("running background TPCC workload")
					runTPCC(ctx, t, c, tpccOptions{
						Warehouses: headroomWarehouses,
						Duration:   duration,
						SetupType:  usingExistingData,
						Start: func(ctx context.Context, t test.Test, c cluster.Cluster) {
							__antithesis_instrumentation__.Notify(51580)

						},
					})
					__antithesis_instrumentation__.Notify(51579)
					return nil
				}}
			__antithesis_instrumentation__.Notify(51573)
			const (
				mainBinary = ""
				n1         = 1
			)

			bankRows := 65104166 / 2
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(51581)
				bankRows = 1000
			} else {
				__antithesis_instrumentation__.Notify(51582)
			}
			__antithesis_instrumentation__.Notify(51574)

			oldV, err := PredecessorVersion(*t.BuildVersion())
			if err != nil {
				__antithesis_instrumentation__.Notify(51583)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51584)
			}
			__antithesis_instrumentation__.Notify(51575)

			newVersionUpgradeTest(c,
				uploadAndStartFromCheckpointFixture(crdbNodes, oldV),
				waitForUpgradeStep(crdbNodes),
				preventAutoUpgradeStep(n1),

				importTPCCStep(oldV, headroomWarehouses, crdbNodes),

				importLargeBankStep(oldV, bankRows, crdbNodes),

				binaryUpgradeStep(crdbNodes, mainBinary),
				uploadVersionStep(workloadNode, mainBinary),

				tpccBackgroundStepper.launch,

				allowAutoUpgradeStep(n1),
				setClusterSettingVersionStep,

				tpccBackgroundStepper.wait,
			).run(ctx, t)
		},
	})
	__antithesis_instrumentation__.Notify(51565)
	r.Add(registry.TestSpec{
		Name:            "tpcc-nowait/nodes=3/w=1",
		Owner:           registry.OwnerKV,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51585)
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses:   1,
				Duration:     10 * time.Minute,
				ExtraRunArgs: "--wait=false",
				SetupType:    usingImport,
			})
		},
	})
	__antithesis_instrumentation__.Notify(51566)
	r.Add(registry.TestSpec{
		Name:    "weekly/tpcc/headroom",
		Owner:   registry.OwnerKV,
		Tags:    []string{`weekly`},
		Cluster: r.MakeClusterSpec(4, spec.CPU(16)),

		Timeout:         4*24*time.Hour + 10*time.Hour,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51586)
			warehouses := 1000
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses: warehouses,
				Duration:   4 * 24 * time.Hour,
				SetupType:  usingImport,
			})
		},
	})

	{
		__antithesis_instrumentation__.Notify(51587)
		mrSetup := []struct {
			region string
			zones  string
		}{
			{region: "us-east1", zones: "us-east1-b"},
			{region: "us-west1", zones: "us-west1-b"},
			{region: "europe-west2", zones: "europe-west2-b"},
		}
		zs := []string{}

		regions := []string{}
		for _, s := range mrSetup {
			__antithesis_instrumentation__.Notify(51590)
			regions = append(regions, s.region)
			zs = append(zs, s.zones)
		}
		__antithesis_instrumentation__.Notify(51588)
		const nodesPerRegion = 3
		const warehousesPerRegion = 20

		multiRegionTests := []struct {
			desc         string
			name         string
			survivalGoal string

			chaosTarget       func(iter int) option.NodeListOption
			workloadInstances []workloadInstance
		}{
			{
				desc:         "test zone survivability works when single nodes are down",
				name:         "tpcc/multiregion/survive=zone/chaos=true",
				survivalGoal: "zone",
				chaosTarget: func(iter int) option.NodeListOption {
					__antithesis_instrumentation__.Notify(51591)
					return option.NodeListOption{(iter % (len(regions) * nodesPerRegion)) + 1}
				},
				workloadInstances: func() []workloadInstance {
					__antithesis_instrumentation__.Notify(51592)
					const prometheusPortStart = 2110
					ret := []workloadInstance{}
					for i := 0; i < len(regions)*nodesPerRegion; i++ {
						__antithesis_instrumentation__.Notify(51594)
						ret = append(
							ret,
							workloadInstance{
								nodes:          option.NodeListOption{i + 1},
								prometheusPort: prometheusPortStart + i,
								extraRunArgs:   fmt.Sprintf("--partition-affinity=%d", i/nodesPerRegion),
							},
						)
					}
					__antithesis_instrumentation__.Notify(51593)
					return ret
				}(),
			},
			{
				desc:         "test region survivability works when regions going down",
				name:         "tpcc/multiregion/survive=region/chaos=true",
				survivalGoal: "region",
				chaosTarget: func(iter int) option.NodeListOption {
					__antithesis_instrumentation__.Notify(51595)
					regionIdx := iter % len(regions)
					return option.NewNodeListOptionRange(
						(nodesPerRegion*regionIdx)+1,
						(nodesPerRegion * (regionIdx + 1)),
					)
				},
				workloadInstances: func() []workloadInstance {
					__antithesis_instrumentation__.Notify(51596)

					const prometheusLocalPortStart = 2110
					const prometheusRemotePortStart = 2120
					ret := []workloadInstance{}
					for i := 0; i < len(regions); i++ {
						__antithesis_instrumentation__.Notify(51598)

						ret = append(
							ret,
							workloadInstance{
								nodes:          option.NewNodeListOptionRange((i*nodesPerRegion)+1, ((i + 1) * nodesPerRegion)),
								prometheusPort: prometheusLocalPortStart + i,
								extraRunArgs:   fmt.Sprintf("--partition-affinity=%d", i),
							},
						)

						ret = append(
							ret,
							workloadInstance{
								nodes:          option.NewNodeListOptionRange((i*nodesPerRegion)+1, ((i + 1) * nodesPerRegion)),
								prometheusPort: prometheusRemotePortStart + i,
								extraRunArgs:   fmt.Sprintf("--partition-affinity=%d", (i+1)%len(regions)),
							},
						)
					}
					__antithesis_instrumentation__.Notify(51597)
					return ret
				}(),
			},
		}
		__antithesis_instrumentation__.Notify(51589)

		for i := range multiRegionTests {
			__antithesis_instrumentation__.Notify(51599)
			tc := multiRegionTests[i]
			r.Add(registry.TestSpec{
				Name:  tc.name,
				Owner: registry.OwnerMultiRegion,

				Cluster:         r.MakeClusterSpec(len(regions)*nodesPerRegion+1, spec.Geo(), spec.Zones(strings.Join(zs, ","))),
				EncryptAtRandom: true,
				Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(51600)
					t.Status(tc.desc)
					duration := 90 * time.Minute
					partitionArgs := fmt.Sprintf(
						`--survival-goal=%s --regions=%s --partitions=%d`,
						tc.survivalGoal,
						strings.Join(regions, ","),
						len(regions),
					)
					iter := 0
					chaosEventCh := make(chan ChaosEvent)
					runTPCC(ctx, t, c, tpccOptions{
						Warehouses:     len(regions) * warehousesPerRegion,
						Duration:       duration,
						ExtraSetupArgs: partitionArgs,
						ExtraRunArgs:   `--method=simple --wait=false --tolerate-errors ` + partitionArgs,
						Chaos: func() Chaos {
							__antithesis_instrumentation__.Notify(51601)
							return Chaos{
								Timer: Periodic{
									Period:   300 * time.Second,
									DownTime: 300 * time.Second,
								},
								Target: func() option.NodeListOption {
									__antithesis_instrumentation__.Notify(51602)
									ret := tc.chaosTarget(iter)
									iter++
									return ret
								},
								Stopper:      time.After(duration),
								DrainAndQuit: false,
								ChaosEventCh: chaosEventCh,
							}
						},
						ChaosEventsProcessor: func(
							prometheusNode option.NodeListOption,
							workloadInstances []workloadInstance,
						) (tpccChaosEventProcessor, error) {
							__antithesis_instrumentation__.Notify(51603)
							prometheusNodeIP, err := c.ExternalIP(ctx, t.L(), prometheusNode)
							if err != nil {
								__antithesis_instrumentation__.Notify(51606)
								return tpccChaosEventProcessor{}, err
							} else {
								__antithesis_instrumentation__.Notify(51607)
							}
							__antithesis_instrumentation__.Notify(51604)
							client, err := promapi.NewClient(promapi.Config{
								Address: fmt.Sprintf("http://%s:9090", prometheusNodeIP[0]),
							})
							if err != nil {
								__antithesis_instrumentation__.Notify(51608)
								return tpccChaosEventProcessor{}, err
							} else {
								__antithesis_instrumentation__.Notify(51609)
							}
							__antithesis_instrumentation__.Notify(51605)
							return tpccChaosEventProcessor{
								workloadInstances: workloadInstances,
								workloadNodeIP:    prometheusNodeIP[0],
								ops: []string{
									"newOrder",
									"delivery",
									"payment",
									"orderStatus",
									"stockLevel",
								},
								ch:         chaosEventCh,
								promClient: promv1.NewAPI(client),

								maxErrorsDuringUptime: warehousesPerRegion * tpcc.NumWorkersPerWarehouse,

								allowZeroSuccessDuringUptime: true,
							}, nil
						},
						SetupType:         usingInit,
						WorkloadInstances: tc.workloadInstances,
					})
				},
			})
		}
	}
	__antithesis_instrumentation__.Notify(51567)

	r.Add(registry.TestSpec{
		Name:            "tpcc/w=100/nodes=3/chaos=true",
		Owner:           registry.OwnerKV,
		Cluster:         r.MakeClusterSpec(4),
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51610)
			duration := 30 * time.Minute
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses: 100,
				Duration:   duration,

				ExtraRunArgs: "--method=simple --wait=false --tolerate-errors",
				Chaos: func() Chaos {
					__antithesis_instrumentation__.Notify(51611)
					return Chaos{
						Timer: Periodic{
							Period:   45 * time.Second,
							DownTime: 10 * time.Second,
						},
						Target: func() option.NodeListOption {
							__antithesis_instrumentation__.Notify(51612)
							return c.Node(1 + rand.Intn(c.Spec().NodeCount-1))
						},
						Stopper:      time.After(duration),
						DrainAndQuit: false,
					}
				},
				SetupType: usingImport,
			})
		},
	})
	__antithesis_instrumentation__.Notify(51568)
	r.Add(registry.TestSpec{
		Name:            "tpcc/interleaved/nodes=3/cpu=16/w=500",
		Owner:           registry.OwnerSQLQueries,
		Cluster:         r.MakeClusterSpec(4, spec.CPU(16)),
		Timeout:         6 * time.Hour,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51613)
			skip.WithIssue(t, 53886)
			runTPCC(ctx, t, c, tpccOptions{

				Warehouses:     500,
				Duration:       time.Minute * 15,
				ExtraSetupArgs: "--interleaved=true",
				SetupType:      usingInit,
			})
		},
	})
	__antithesis_instrumentation__.Notify(51569)

	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes: 3,
		CPUs:  4,

		LoadWarehouses: 1000,
		EstimatedMax:   gceOrAws(cloud, 650, 800),
	})
	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes: 3,
		CPUs:  16,

		LoadWarehouses: gceOrAws(cloud, 3000, 3500),
		EstimatedMax:   gceOrAws(cloud, 2400, 3000),
	})
	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes:                    3,
		CPUs:                     16,
		AdmissionControlDisabled: true,

		LoadWarehouses: gceOrAws(cloud, 3000, 3500),
		EstimatedMax:   gceOrAws(cloud, 2400, 3000),
	})
	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes: 12,
		CPUs:  16,

		LoadWarehouses: gceOrAws(cloud, 8000, 10000),
		EstimatedMax:   gceOrAws(cloud, 7000, 8000),

		Tags: []string{`weekly`},
	})
	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes:        6,
		CPUs:         16,
		Distribution: multiZone,

		LoadWarehouses: 5000,
		EstimatedMax:   2500,
	})
	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes:        9,
		CPUs:         4,
		Distribution: multiRegion,
		LoadConfig:   multiLoadgen,

		LoadWarehouses: 3000,
		EstimatedMax:   2000,
	})
	registerTPCCBenchSpec(r, tpccBenchSpec{
		Nodes:      9,
		CPUs:       4,
		Chaos:      true,
		LoadConfig: singlePartitionedLoadgen,

		LoadWarehouses: 2000,
		EstimatedMax:   900,
	})
}

func gceOrAws(cloud string, gce, aws int) int {
	__antithesis_instrumentation__.Notify(51614)
	if cloud == "aws" {
		__antithesis_instrumentation__.Notify(51616)
		return aws
	} else {
		__antithesis_instrumentation__.Notify(51617)
	}
	__antithesis_instrumentation__.Notify(51615)
	return gce
}

type tpccBenchDistribution int

const (
	singleZone tpccBenchDistribution = iota

	multiZone

	multiRegion
)

func (d tpccBenchDistribution) zones() []string {
	__antithesis_instrumentation__.Notify(51618)
	switch d {
	case singleZone:
		__antithesis_instrumentation__.Notify(51619)
		return []string{"us-central1-b"}
	case multiZone:
		__antithesis_instrumentation__.Notify(51620)

		return []string{"us-central1-f", "us-central1-b", "us-central1-c"}
	case multiRegion:
		__antithesis_instrumentation__.Notify(51621)
		return []string{"us-east1-b", "us-west1-b", "europe-west2-b"}
	default:
		__antithesis_instrumentation__.Notify(51622)
		panic("unexpected")
	}
}

type tpccBenchLoadConfig int

const (
	singleLoadgen tpccBenchLoadConfig = iota

	singlePartitionedLoadgen

	multiLoadgen
)

func (l tpccBenchLoadConfig) numLoadNodes(d tpccBenchDistribution) int {
	__antithesis_instrumentation__.Notify(51623)
	switch l {
	case singleLoadgen:
		__antithesis_instrumentation__.Notify(51624)
		return 1
	case singlePartitionedLoadgen:
		__antithesis_instrumentation__.Notify(51625)
		return 1
	case multiLoadgen:
		__antithesis_instrumentation__.Notify(51626)
		return len(d.zones())
	default:
		__antithesis_instrumentation__.Notify(51627)
		panic("unexpected")
	}
}

type tpccBenchSpec struct {
	Nodes                    int
	CPUs                     int
	Chaos                    bool
	AdmissionControlDisabled bool
	Distribution             tpccBenchDistribution
	LoadConfig               tpccBenchLoadConfig

	LoadWarehouses int

	EstimatedMax int

	MinVersion string

	Tags []string
}

func (s tpccBenchSpec) partitions() int {
	__antithesis_instrumentation__.Notify(51628)
	switch s.LoadConfig {
	case singleLoadgen:
		__antithesis_instrumentation__.Notify(51629)
		return 0
	case singlePartitionedLoadgen:
		__antithesis_instrumentation__.Notify(51630)
		return s.Nodes / 3
	case multiLoadgen:
		__antithesis_instrumentation__.Notify(51631)
		return len(s.Distribution.zones())
	default:
		__antithesis_instrumentation__.Notify(51632)
		panic("unexpected")
	}
}

func (s tpccBenchSpec) startOpts() (option.StartOpts, install.ClusterSettings) {
	__antithesis_instrumentation__.Notify(51633)
	startOpts := option.DefaultStartOpts()
	settings := install.MakeClusterSettings()

	settings.Env = append(settings.Env, "COCKROACH_MEMPROF_INTERVAL=15s")
	if s.LoadConfig == singlePartitionedLoadgen {
		__antithesis_instrumentation__.Notify(51635)
		settings.NumRacks = s.partitions()
	} else {
		__antithesis_instrumentation__.Notify(51636)
	}
	__antithesis_instrumentation__.Notify(51634)
	return startOpts, settings
}

func registerTPCCBenchSpec(r registry.Registry, b tpccBenchSpec) {
	__antithesis_instrumentation__.Notify(51637)
	nameParts := []string{
		"tpccbench",
		fmt.Sprintf("nodes=%d", b.Nodes),
		fmt.Sprintf("cpu=%d", b.CPUs),
	}
	if b.Chaos {
		__antithesis_instrumentation__.Notify(51643)
		nameParts = append(nameParts, "chaos")
	} else {
		__antithesis_instrumentation__.Notify(51644)
	}
	__antithesis_instrumentation__.Notify(51638)
	if b.AdmissionControlDisabled {
		__antithesis_instrumentation__.Notify(51645)
		nameParts = append(nameParts, "no-admission")
	} else {
		__antithesis_instrumentation__.Notify(51646)
	}
	__antithesis_instrumentation__.Notify(51639)

	opts := []spec.Option{spec.CPU(b.CPUs)}
	switch b.Distribution {
	case singleZone:
		__antithesis_instrumentation__.Notify(51647)

	case multiZone:
		__antithesis_instrumentation__.Notify(51648)
		nameParts = append(nameParts, "multi-az")
		opts = append(opts, spec.Geo(), spec.Zones(strings.Join(b.Distribution.zones(), ",")))
	case multiRegion:
		__antithesis_instrumentation__.Notify(51649)
		nameParts = append(nameParts, "multi-region")
		opts = append(opts, spec.Geo(), spec.Zones(strings.Join(b.Distribution.zones(), ",")))
	default:
		__antithesis_instrumentation__.Notify(51650)
		panic("unexpected")
	}
	__antithesis_instrumentation__.Notify(51640)

	switch b.LoadConfig {
	case singleLoadgen:
		__antithesis_instrumentation__.Notify(51651)

	case singlePartitionedLoadgen:
		__antithesis_instrumentation__.Notify(51652)
		nameParts = append(nameParts, "partition")
	case multiLoadgen:
		__antithesis_instrumentation__.Notify(51653)

	default:
		__antithesis_instrumentation__.Notify(51654)
		panic("unexpected")
	}
	__antithesis_instrumentation__.Notify(51641)

	name := strings.Join(nameParts, "/")

	numNodes := b.Nodes + b.LoadConfig.numLoadNodes(b.Distribution)
	nodes := r.MakeClusterSpec(numNodes, opts...)

	minVersion := b.MinVersion
	if minVersion == "" {
		__antithesis_instrumentation__.Notify(51655)
		minVersion = "v19.1.0"
	} else {
		__antithesis_instrumentation__.Notify(51656)
	}
	__antithesis_instrumentation__.Notify(51642)

	r.Add(registry.TestSpec{
		Name:    name,
		Owner:   registry.OwnerKV,
		Cluster: nodes,
		Tags:    b.Tags,

		EncryptAtRandom: false,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51657)
			runTPCCBench(ctx, t, c, b)
		},
	})
}

func loadTPCCBench(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	b tpccBenchSpec,
	roachNodes, loadNode option.NodeListOption,
) error {
	__antithesis_instrumentation__.Notify(51658)
	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	if _, err := db.ExecContext(ctx, `USE tpcc`); err == nil {
		__antithesis_instrumentation__.Notify(51665)
		t.L().Printf("found existing tpcc database\n")

		var curWarehouses int
		if err := db.QueryRowContext(ctx,
			`SELECT count(*) FROM tpcc.warehouse`,
		).Scan(&curWarehouses); err != nil {
			__antithesis_instrumentation__.Notify(51668)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51669)
		}
		__antithesis_instrumentation__.Notify(51666)
		if curWarehouses >= b.LoadWarehouses {
			__antithesis_instrumentation__.Notify(51670)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(51671)
		}
		__antithesis_instrumentation__.Notify(51667)

		c.Wipe(ctx, roachNodes)
		startOpts, settings := b.startOpts()
		c.Start(ctx, t.L(), startOpts, settings, roachNodes)
	} else {
		__antithesis_instrumentation__.Notify(51672)
		if pqErr := (*pq.Error)(nil); !(errors.As(err, &pqErr) && func() bool {
			__antithesis_instrumentation__.Notify(51673)
			return pgcode.MakeCode(string(pqErr.Code)) == pgcode.InvalidCatalogName == true
		}() == true) {
			__antithesis_instrumentation__.Notify(51674)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51675)
		}
	}
	__antithesis_instrumentation__.Notify(51659)

	var loadArgs string
	var rebalanceWait time.Duration
	switch b.LoadConfig {
	case singleLoadgen:
		__antithesis_instrumentation__.Notify(51676)
		loadArgs = `--checks=false`
		rebalanceWait = time.Duration(b.LoadWarehouses/250) * time.Minute
	case singlePartitionedLoadgen:
		__antithesis_instrumentation__.Notify(51677)
		loadArgs = fmt.Sprintf(`--checks=false --partitions=%d`, b.partitions())
		rebalanceWait = time.Duration(b.LoadWarehouses/125) * time.Minute
	case multiLoadgen:
		__antithesis_instrumentation__.Notify(51678)
		loadArgs = fmt.Sprintf(`--checks=false --partitions=%d --zones="%s"`,
			b.partitions(), strings.Join(b.Distribution.zones(), ","))
		rebalanceWait = time.Duration(b.LoadWarehouses/50) * time.Minute
	default:
		__antithesis_instrumentation__.Notify(51679)
		panic("unexpected")
	}
	__antithesis_instrumentation__.Notify(51660)

	t.L().Printf("restoring tpcc fixture\n")
	err := WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)
	cmd := tpccImportCmd(b.LoadWarehouses, loadArgs)
	if err = c.RunE(ctx, roachNodes[:1], cmd); err != nil {
		__antithesis_instrumentation__.Notify(51680)
		return err
	} else {
		__antithesis_instrumentation__.Notify(51681)
	}
	__antithesis_instrumentation__.Notify(51661)
	if rebalanceWait == 0 || func() bool {
		__antithesis_instrumentation__.Notify(51682)
		return len(roachNodes) <= 3 == true
	}() == true {
		__antithesis_instrumentation__.Notify(51683)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(51684)
	}
	__antithesis_instrumentation__.Notify(51662)

	t.L().Printf("waiting %v for rebalancing\n", rebalanceWait)
	_, err = db.ExecContext(ctx, `SET CLUSTER SETTING kv.snapshot_rebalance.max_rate='128MiB'`)
	if err != nil {
		__antithesis_instrumentation__.Notify(51685)
		return err
	} else {
		__antithesis_instrumentation__.Notify(51686)
	}
	__antithesis_instrumentation__.Notify(51663)

	const txnsPerWarehousePerSecond = 12.8 * (23.0 / 10.0) * (1.0 / 60.0)
	rateAtExpected := txnsPerWarehousePerSecond * float64(b.EstimatedMax)
	maxRate := int(rateAtExpected / 2)
	rampTime := (1 * rebalanceWait) / 4
	loadTime := (3 * rebalanceWait) / 4
	cmd = fmt.Sprintf("./cockroach workload run tpcc --warehouses=%d --workers=%d --max-rate=%d "+
		"--wait=false --ramp=%s --duration=%s --scatter --tolerate-errors {pgurl%s}",
		b.LoadWarehouses, b.LoadWarehouses, maxRate, rampTime, loadTime, roachNodes)
	if _, err := c.RunWithDetailsSingleNode(ctx, t.L(), loadNode, cmd); err != nil {
		__antithesis_instrumentation__.Notify(51687)
		return err
	} else {
		__antithesis_instrumentation__.Notify(51688)
	}
	__antithesis_instrumentation__.Notify(51664)

	_, err = db.ExecContext(ctx, `SET CLUSTER SETTING kv.snapshot_rebalance.max_rate='2MiB'`)
	return err
}

func runTPCCBench(ctx context.Context, t test.Test, c cluster.Cluster, b tpccBenchSpec) {
	__antithesis_instrumentation__.Notify(51689)

	numLoadGroups := b.LoadConfig.numLoadNodes(b.Distribution)
	numZones := len(b.Distribution.zones())
	loadGroups := makeLoadGroups(c, numZones, b.Nodes, numLoadGroups)
	roachNodes := loadGroups.roachNodes()
	loadNodes := loadGroups.loadNodes()
	c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)

	c.Put(ctx, t.Cockroach(), "./cockroach", loadNodes)

	startOpts, settings := b.startOpts()
	c.Start(ctx, t.L(), startOpts, settings, roachNodes)
	SetAdmissionControl(ctx, t, c, !b.AdmissionControlDisabled)
	useHAProxy := b.Chaos
	const restartWait = 15 * time.Second
	{
		__antithesis_instrumentation__.Notify(51694)

		time.Sleep(restartWait)
		if useHAProxy {
			__antithesis_instrumentation__.Notify(51697)
			if len(loadNodes) > 1 {
				__antithesis_instrumentation__.Notify(51701)
				t.Fatal("distributed chaos benchmarking not supported")
			} else {
				__antithesis_instrumentation__.Notify(51702)
			}
			__antithesis_instrumentation__.Notify(51698)
			t.Status("installing haproxy")
			if err := c.Install(ctx, t.L(), loadNodes, "haproxy"); err != nil {
				__antithesis_instrumentation__.Notify(51703)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51704)
			}
			__antithesis_instrumentation__.Notify(51699)
			c.Run(ctx, loadNodes, "./cockroach gen haproxy --insecure --url {pgurl:1}")

			c.Run(ctx, loadNodes, "sed -i 's/maxconn [0-9]\\+/maxconn 21000/' haproxy.cfg")
			if b.LoadWarehouses > 1e4 {
				__antithesis_instrumentation__.Notify(51705)
				t.Fatal("HAProxy config supports up to 10k warehouses")
			} else {
				__antithesis_instrumentation__.Notify(51706)
			}
			__antithesis_instrumentation__.Notify(51700)
			c.Run(ctx, loadNodes, "haproxy -f haproxy.cfg -D")
		} else {
			__antithesis_instrumentation__.Notify(51707)
		}
		__antithesis_instrumentation__.Notify(51695)

		m := c.NewMonitor(ctx, roachNodes)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(51708)
			t.Status("setting up dataset")
			return loadTPCCBench(ctx, t, c, b, roachNodes, c.Node(loadNodes[0]))
		})
		__antithesis_instrumentation__.Notify(51696)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(51690)

	precision := int(math.Max(1.0, float64(b.LoadWarehouses/200)))
	initStepSize := precision

	resultsDir, err := ioutil.TempDir("", "roachtest-tpcc")
	if err != nil {
		__antithesis_instrumentation__.Notify(51709)
		t.Fatal(errors.Wrap(err, "failed to create temp dir"))
	} else {
		__antithesis_instrumentation__.Notify(51710)
	}
	__antithesis_instrumentation__.Notify(51691)
	defer func() { __antithesis_instrumentation__.Notify(51711); _ = os.RemoveAll(resultsDir) }()
	__antithesis_instrumentation__.Notify(51692)

	restart := func() {
		__antithesis_instrumentation__.Notify(51712)

		if err := c.Reset(ctx, t.L()); err != nil {
			__antithesis_instrumentation__.Notify(51716)

			t.L().Printf("failed to reset VMs, proceeding anyway: %s", err)
			_ = err
		} else {
			__antithesis_instrumentation__.Notify(51717)
		}
		__antithesis_instrumentation__.Notify(51713)
		var ok bool
		for i := 0; i < 10; i++ {
			__antithesis_instrumentation__.Notify(51718)
			if err := ctx.Err(); err != nil {
				__antithesis_instrumentation__.Notify(51721)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51722)
			}
			__antithesis_instrumentation__.Notify(51719)
			shortCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
			if err := c.StopE(shortCtx, t.L(), option.DefaultStopOpts(), roachNodes); err != nil {
				__antithesis_instrumentation__.Notify(51723)
				cancel()
				t.L().Printf("unable to stop cluster; retrying to allow vm to recover: %s", err)

				select {
				case <-time.After(30 * time.Second):
					__antithesis_instrumentation__.Notify(51725)
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(51726)
				}
				__antithesis_instrumentation__.Notify(51724)
				continue
			} else {
				__antithesis_instrumentation__.Notify(51727)
			}
			__antithesis_instrumentation__.Notify(51720)
			cancel()
			ok = true
			break
		}
		__antithesis_instrumentation__.Notify(51714)
		if !ok {
			__antithesis_instrumentation__.Notify(51728)
			t.Fatalf("VM is hosed; giving up")
		} else {
			__antithesis_instrumentation__.Notify(51729)
		}
		__antithesis_instrumentation__.Notify(51715)

		startOpts, settings := b.startOpts()
		c.Start(ctx, t.L(), startOpts, settings, roachNodes)
		SetAdmissionControl(ctx, t, c, !b.AdmissionControlDisabled)
	}
	__antithesis_instrumentation__.Notify(51693)

	s := search.NewLineSearcher(1, b.LoadWarehouses, b.EstimatedMax, initStepSize, precision)
	iteration := 0
	if res, err := s.Search(func(warehouses int) (bool, error) {
		__antithesis_instrumentation__.Notify(51730)
		iteration++
		t.L().Printf("initializing cluster for %d warehouses (search attempt: %d)", warehouses, iteration)

		restart()

		time.Sleep(restartWait)

		rampDur := 5 * time.Minute
		loadDur := 10 * time.Minute
		loadDone := make(chan time.Time, numLoadGroups)

		m := c.NewMonitor(ctx, roachNodes)

		if b.Chaos {
			__antithesis_instrumentation__.Notify(51736)

			ch := Chaos{
				Timer:   Periodic{Period: 90 * time.Second, DownTime: 5 * time.Second},
				Target:  roachNodes.RandNode,
				Stopper: loadDone,
			}
			m.Go(ch.Runner(c, t, m))
		} else {
			__antithesis_instrumentation__.Notify(51737)
		}
		__antithesis_instrumentation__.Notify(51731)
		if b.Distribution == multiRegion {
			__antithesis_instrumentation__.Notify(51738)
			rampDur = 3 * time.Minute
			loadDur = 15 * time.Minute
		} else {
			__antithesis_instrumentation__.Notify(51739)
		}
		__antithesis_instrumentation__.Notify(51732)

		resultChan := make(chan *tpcc.Result, numLoadGroups)
		for groupIdx, group := range loadGroups {
			__antithesis_instrumentation__.Notify(51740)

			groupIdx := groupIdx
			group := group
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(51741)
				sqlGateways := group.roachNodes
				if useHAProxy {
					__antithesis_instrumentation__.Notify(51748)
					sqlGateways = group.loadNodes
				} else {
					__antithesis_instrumentation__.Notify(51749)
				}
				__antithesis_instrumentation__.Notify(51742)

				extraFlags := ""
				switch b.LoadConfig {
				case singleLoadgen:
					__antithesis_instrumentation__.Notify(51750)

				case singlePartitionedLoadgen:
					__antithesis_instrumentation__.Notify(51751)
					extraFlags = fmt.Sprintf(` --partitions=%d`, b.partitions())
				case multiLoadgen:
					__antithesis_instrumentation__.Notify(51752)
					extraFlags = fmt.Sprintf(` --partitions=%d --partition-affinity=%d`,
						b.partitions(), groupIdx)
				default:
					__antithesis_instrumentation__.Notify(51753)

					t.Fatalf("unimplemented LoadConfig %v", b.LoadConfig)
				}
				__antithesis_instrumentation__.Notify(51743)
				if b.Chaos {
					__antithesis_instrumentation__.Notify(51754)

					extraFlags += " --method=simple"
				} else {
					__antithesis_instrumentation__.Notify(51755)
				}
				__antithesis_instrumentation__.Notify(51744)
				t.Status(fmt.Sprintf("running benchmark, warehouses=%d", warehouses))
				histogramsPath := fmt.Sprintf("%s/warehouses=%d/stats.json", t.PerfArtifactsDir(), warehouses)
				cmd := fmt.Sprintf("./cockroach workload run tpcc --warehouses=%d --active-warehouses=%d "+
					"--tolerate-errors --ramp=%s --duration=%s%s --histograms=%s {pgurl%s}",
					b.LoadWarehouses, warehouses, rampDur,
					loadDur, extraFlags, histogramsPath, sqlGateways)
				err := c.RunE(ctx, group.loadNodes, cmd)
				loadDone <- timeutil.Now()
				if err != nil {
					__antithesis_instrumentation__.Notify(51756)

					return errors.Wrapf(err, "error running tpcc load generator")
				} else {
					__antithesis_instrumentation__.Notify(51757)
				}
				__antithesis_instrumentation__.Notify(51745)
				roachtestHistogramsPath := filepath.Join(resultsDir, fmt.Sprintf("%d.%d-stats.json", warehouses, groupIdx))
				if err := c.Get(
					ctx, t.L(), histogramsPath, roachtestHistogramsPath, group.loadNodes,
				); err != nil {
					__antithesis_instrumentation__.Notify(51758)

					return err
				} else {
					__antithesis_instrumentation__.Notify(51759)
				}
				__antithesis_instrumentation__.Notify(51746)
				snapshots, err := histogram.DecodeSnapshots(roachtestHistogramsPath)
				if err != nil {
					__antithesis_instrumentation__.Notify(51760)

					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(51761)
				}
				__antithesis_instrumentation__.Notify(51747)
				result := tpcc.NewResultWithSnapshots(warehouses, 0, snapshots)
				resultChan <- result
				return nil
			})
		}
		__antithesis_instrumentation__.Notify(51733)
		failErr := m.WaitE()
		close(resultChan)

		var res *tpcc.Result
		if failErr != nil {
			__antithesis_instrumentation__.Notify(51762)
			if t.Failed() {
				__antithesis_instrumentation__.Notify(51764)

				return false, err
			} else {
				__antithesis_instrumentation__.Notify(51765)
			}
			__antithesis_instrumentation__.Notify(51763)

			res = &tpcc.Result{
				ActiveWarehouses: warehouses,
			}
		} else {
			__antithesis_instrumentation__.Notify(51766)

			var results []*tpcc.Result
			for partial := range resultChan {
				__antithesis_instrumentation__.Notify(51768)
				results = append(results, partial)
			}
			__antithesis_instrumentation__.Notify(51767)
			res = tpcc.MergeResults(results...)
			failErr = res.FailureError()
		}
		__antithesis_instrumentation__.Notify(51734)

		if failErr == nil {
			__antithesis_instrumentation__.Notify(51769)
			ttycolor.Stdout(ttycolor.Green)
			t.L().Printf("--- SEARCH ITER PASS: TPCC %d resulted in %.1f tpmC (%.1f%% of max tpmC)\n\n",
				warehouses, res.TpmC(), res.Efficiency())
		} else {
			__antithesis_instrumentation__.Notify(51770)
			ttycolor.Stdout(ttycolor.Red)
			t.L().Printf("--- SEARCH ITER FAIL: TPCC %d resulted in %.1f tpmC and failed due to %v",
				warehouses, res.TpmC(), failErr)
		}
		__antithesis_instrumentation__.Notify(51735)
		ttycolor.Stdout(ttycolor.Reset)
		return failErr == nil, nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(51771)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51772)

		restart()
		ttycolor.Stdout(ttycolor.Green)
		t.L().Printf("------\nMAX WAREHOUSES = %d\n------\n\n", res)
		ttycolor.Stdout(ttycolor.Reset)
	}
}

func registerTPCCBench(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51773)
	specs := []tpccBenchSpec{
		{
			Nodes: 3,
			CPUs:  4,

			LoadWarehouses: 1000,
			EstimatedMax:   325,
		},
		{
			Nodes: 3,
			CPUs:  16,

			LoadWarehouses: 2000,
			EstimatedMax:   1300,
		},

		{
			Nodes: 30,
			CPUs:  16,

			LoadWarehouses: 10000,
			EstimatedMax:   5300,
		},

		{
			Nodes:      18,
			CPUs:       16,
			LoadConfig: singlePartitionedLoadgen,

			LoadWarehouses: 10000,
			EstimatedMax:   8000,
		},

		{
			Nodes: 7,
			CPUs:  16,
			Chaos: true,

			LoadWarehouses: 5000,
			EstimatedMax:   2000,
		},

		{
			Nodes:        3,
			CPUs:         16,
			Distribution: multiZone,

			LoadWarehouses: 2000,
			EstimatedMax:   1000,
		},

		{
			Nodes:        9,
			CPUs:         16,
			Distribution: multiRegion,
			LoadConfig:   multiLoadgen,

			LoadWarehouses: 12000,
			EstimatedMax:   8000,
		},

		{
			Nodes: 64,
			CPUs:  16,

			LoadWarehouses: 50000,
			EstimatedMax:   40000,
		},

		{
			Nodes: 6,
			CPUs:  16,

			LoadWarehouses: 5000,
			EstimatedMax:   3000,
			LoadConfig:     singlePartitionedLoadgen,
		},
		{
			Nodes: 12,
			CPUs:  16,

			LoadWarehouses: 10000,
			EstimatedMax:   6000,
			LoadConfig:     singlePartitionedLoadgen,
		},
		{
			Nodes: 24,
			CPUs:  16,

			LoadWarehouses: 20000,
			EstimatedMax:   12000,
			LoadConfig:     singlePartitionedLoadgen,
		},

		{
			Nodes: 11,
			CPUs:  32,

			LoadWarehouses: 10000,
			EstimatedMax:   8000,
		},
	}

	for _, b := range specs {
		__antithesis_instrumentation__.Notify(51774)
		registerTPCCBenchSpec(r, b)
	}
}

func makeWorkloadScrapeNodes(
	workloadNode option.NodeListOption, workloadInstances []workloadInstance,
) []prometheus.ScrapeNode {
	__antithesis_instrumentation__.Notify(51775)
	workloadScrapeNodes := make([]prometheus.ScrapeNode, len(workloadInstances))
	for i, workloadInstance := range workloadInstances {
		__antithesis_instrumentation__.Notify(51777)
		workloadScrapeNodes[i] = prometheus.ScrapeNode{
			Nodes: workloadNode,
			Port:  workloadInstance.prometheusPort,
		}
	}
	__antithesis_instrumentation__.Notify(51776)
	return workloadScrapeNodes
}

func setupPrometheus(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	opts tpccOptions,
	workloadInstances []workloadInstance,
) (*prometheus.Config, func()) {
	__antithesis_instrumentation__.Notify(51778)
	cfg := opts.PrometheusConfig
	if cfg == nil {
		__antithesis_instrumentation__.Notify(51784)

		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(51787)
			return nil, func() { __antithesis_instrumentation__.Notify(51788) }
		} else {
			__antithesis_instrumentation__.Notify(51789)
		}
		__antithesis_instrumentation__.Notify(51785)
		if opts.DisablePrometheus {
			__antithesis_instrumentation__.Notify(51790)
			return nil, func() { __antithesis_instrumentation__.Notify(51791) }
		} else {
			__antithesis_instrumentation__.Notify(51792)
		}
		__antithesis_instrumentation__.Notify(51786)
		workloadNode := c.Node(c.Spec().NodeCount)
		cfg = &prometheus.Config{
			PrometheusNode: workloadNode,

			ScrapeConfigs: []prometheus.ScrapeConfig{
				prometheus.MakeInsecureCockroachScrapeConfig(
					"cockroach",
					c.Range(1, c.Spec().NodeCount-1),
				),
				prometheus.MakeWorkloadScrapeConfig("workload", makeWorkloadScrapeNodes(workloadNode, workloadInstances)),
			},
		}
	} else {
		__antithesis_instrumentation__.Notify(51793)
	}
	__antithesis_instrumentation__.Notify(51779)
	if opts.DisablePrometheus {
		__antithesis_instrumentation__.Notify(51794)
		t.Fatal("test has PrometheusConfig but DisablePrometheus was on")
	} else {
		__antithesis_instrumentation__.Notify(51795)
	}
	__antithesis_instrumentation__.Notify(51780)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(51796)
		t.Skip("skipping test as prometheus is needed, but prometheus does not yet work locally")
		return nil, func() { __antithesis_instrumentation__.Notify(51797) }
	} else {
		__antithesis_instrumentation__.Notify(51798)
	}
	__antithesis_instrumentation__.Notify(51781)
	p, err := prometheus.Init(
		ctx,
		*cfg,
		c,
		t.L(),
		func(ctx context.Context, nodes option.NodeListOption, operation string, args ...string) error {
			__antithesis_instrumentation__.Notify(51799)
			return repeatRunE(
				ctx,
				t,
				c,
				nodes,
				operation,
				args...,
			)
		},
	)
	__antithesis_instrumentation__.Notify(51782)
	if err != nil {
		__antithesis_instrumentation__.Notify(51800)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51801)
	}
	__antithesis_instrumentation__.Notify(51783)

	return cfg, func() {
		__antithesis_instrumentation__.Notify(51802)

		snapshotCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if err := p.Snapshot(
			snapshotCtx,
			c,
			t.L(),
			t.ArtifactsDir(),
		); err != nil {
			__antithesis_instrumentation__.Notify(51803)
			t.L().Printf("failed to get prometheus snapshot: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(51804)
		}
	}
}
