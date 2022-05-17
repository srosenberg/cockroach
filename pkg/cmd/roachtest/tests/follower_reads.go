package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgtype"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func registerFollowerReads(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47727)
	register := func(
		survival survivalGoal, locality localitySetting, rc readConsistency, insufficientQuorum bool,
	) {
		__antithesis_instrumentation__.Notify(47730)
		name := fmt.Sprintf("follower-reads/survival=%s/locality=%s/reads=%s", survival, locality, rc)
		if insufficientQuorum {
			__antithesis_instrumentation__.Notify(47732)
			name = name + "/insufficient-quorum"
		} else {
			__antithesis_instrumentation__.Notify(47733)
		}
		__antithesis_instrumentation__.Notify(47731)
		r.Add(registry.TestSpec{
			Name:  name,
			Owner: registry.OwnerKV,
			Cluster: r.MakeClusterSpec(
				6,
				spec.CPU(4),
				spec.Geo(),
				spec.Zones("us-east1-b,us-east1-b,us-east1-b,us-west1-b,us-west1-b,europe-west2-b"),
			),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(47734)
				c.Put(ctx, t.Cockroach(), "./cockroach")
				c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
				topology := topologySpec{
					multiRegion:       true,
					locality:          locality,
					survival:          survival,
					deadPrimaryRegion: insufficientQuorum,
				}
				data := initFollowerReadsDB(ctx, t, c, topology)
				runFollowerReadsTest(ctx, t, c, topology, rc, data)
			},
		})
	}
	__antithesis_instrumentation__.Notify(47728)
	for _, survival := range []survivalGoal{zone, region} {
		__antithesis_instrumentation__.Notify(47735)
		for _, locality := range []localitySetting{regional, global} {
			__antithesis_instrumentation__.Notify(47737)
			for _, rc := range []readConsistency{strong, exactStaleness, boundedStaleness} {
				__antithesis_instrumentation__.Notify(47738)
				if rc == strong && func() bool {
					__antithesis_instrumentation__.Notify(47740)
					return locality != global == true
				}() == true {
					__antithesis_instrumentation__.Notify(47741)

					continue
				} else {
					__antithesis_instrumentation__.Notify(47742)
				}
				__antithesis_instrumentation__.Notify(47739)
				register(survival, locality, rc, false)
			}
		}
		__antithesis_instrumentation__.Notify(47736)

		register(survival, regional, boundedStaleness, true)
	}
	__antithesis_instrumentation__.Notify(47729)

	r.Add(registry.TestSpec{
		Name:  "follower-reads/mixed-version/single-region",
		Owner: registry.OwnerKV,
		Cluster: r.MakeClusterSpec(
			3,
			spec.CPU(2),
		),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47743)
			runFollowerReadsMixedVersionSingleRegionTest(ctx, t, c, *t.BuildVersion())
		},
	})
}

type survivalGoal string

type localitySetting string

type readConsistency string

const (
	zone   survivalGoal = "zone"
	region survivalGoal = "region"

	regional localitySetting = "regional"
	global   localitySetting = "global"

	strong           readConsistency = "strong"
	exactStaleness   readConsistency = "exact-staleness"
	boundedStaleness readConsistency = "bounded-staleness"
)

type topologySpec struct {
	multiRegion bool

	locality localitySetting
	survival survivalGoal

	deadPrimaryRegion bool
}

func runFollowerReadsTest(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	topology topologySpec,
	rc readConsistency,
	data map[int]int64,
) {
	__antithesis_instrumentation__.Notify(47744)
	var conns []*gosql.DB
	for i := 0; i < c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(47764)
		conns = append(conns, c.Conn(ctx, t.L(), i+1))
		defer conns[i].Close()
	}
	__antithesis_instrumentation__.Notify(47745)
	db := conns[0]

	chooseKV := func() (k int, v int64) {
		__antithesis_instrumentation__.Notify(47765)
		for k, v = range data {
			__antithesis_instrumentation__.Notify(47767)
			return k, v
		}
		__antithesis_instrumentation__.Notify(47766)
		panic("data is empty")
	}
	__antithesis_instrumentation__.Notify(47746)

	var aost string
	switch rc {
	case strong:
		__antithesis_instrumentation__.Notify(47768)
		aost = ""
	case exactStaleness:
		__antithesis_instrumentation__.Notify(47769)
		aost = "AS OF SYSTEM TIME follower_read_timestamp()"
	case boundedStaleness:
		__antithesis_instrumentation__.Notify(47770)
		aost = "AS OF SYSTEM TIME with_max_staleness('10m')"
	default:
		__antithesis_instrumentation__.Notify(47771)
		t.Fatalf("unexpected readConsistency %s", rc)
	}
	__antithesis_instrumentation__.Notify(47747)

	verifySelect := func(ctx context.Context, node, k int, expectedVal int64) func() error {
		__antithesis_instrumentation__.Notify(47772)
		return func() error {
			__antithesis_instrumentation__.Notify(47773)
			nodeDB := conns[node-1]
			q := fmt.Sprintf("SELECT v FROM test.test %s WHERE k = $1", aost)
			r := nodeDB.QueryRowContext(ctx, q, k)
			var got int64
			if err := r.Scan(&got); err != nil {
				__antithesis_instrumentation__.Notify(47776)

				if ctx.Err() != nil {
					__antithesis_instrumentation__.Notify(47778)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(47779)
				}
				__antithesis_instrumentation__.Notify(47777)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47780)
			}
			__antithesis_instrumentation__.Notify(47774)

			if got != expectedVal {
				__antithesis_instrumentation__.Notify(47781)
				return errors.Errorf("Didn't get expected val on node %d: %v != %v",
					node, got, expectedVal)
			} else {
				__antithesis_instrumentation__.Notify(47782)
			}
			__antithesis_instrumentation__.Notify(47775)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(47748)
	doSelects := func(ctx context.Context, node int) func() error {
		__antithesis_instrumentation__.Notify(47783)
		return func() error {
			__antithesis_instrumentation__.Notify(47784)
			for ctx.Err() == nil {
				__antithesis_instrumentation__.Notify(47786)
				k, v := chooseKV()
				err := verifySelect(ctx, node, k, v)()
				if err != nil && func() bool {
					__antithesis_instrumentation__.Notify(47787)
					return ctx.Err() == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(47788)
					return err
				} else {
					__antithesis_instrumentation__.Notify(47789)
				}
			}
			__antithesis_instrumentation__.Notify(47785)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(47749)

	if rc != strong {
		__antithesis_instrumentation__.Notify(47790)

		followerReadDuration, err := computeFollowerReadDuration(ctx, db)
		if err != nil {
			__antithesis_instrumentation__.Notify(47792)
			t.Fatalf("failed to compute follower read duration: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(47793)
		}
		__antithesis_instrumentation__.Notify(47791)
		select {
		case <-time.After(followerReadDuration):
			__antithesis_instrumentation__.Notify(47794)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(47795)
			t.Fatalf("context canceled: %v", ctx.Err())
		}
	} else {
		__antithesis_instrumentation__.Notify(47796)
	}
	__antithesis_instrumentation__.Notify(47750)

	const maxLatencyThreshold = 25 * time.Millisecond
	_, err := db.ExecContext(
		ctx, fmt.Sprintf(
			"SET CLUSTER SETTING sql.trace.stmt.enable_threshold = '%s'",
			maxLatencyThreshold,
		),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(47797)

		if !strings.Contains(err.Error(), "unknown cluster setting") {
			__antithesis_instrumentation__.Notify(47798)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47799)
		}
	} else {
		__antithesis_instrumentation__.Notify(47800)
	}
	__antithesis_instrumentation__.Notify(47751)

	time.Sleep(10 * time.Second)
	followerReadsBefore, err := getFollowerReadCounts(ctx, t, c)
	if err != nil {
		__antithesis_instrumentation__.Notify(47801)
		t.Fatalf("failed to get follower read counts: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(47802)
	}
	__antithesis_instrumentation__.Notify(47752)

	t.L().Printf("warming up reads")
	g, gCtx := errgroup.WithContext(ctx)
	k, v := chooseKV()
	until := timeutil.Now().Add(15 * time.Second)
	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(47803)
		fn := verifySelect(gCtx, i, k, v)
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(47804)
			for {
				__antithesis_instrumentation__.Notify(47805)
				if timeutil.Now().After(until) {
					__antithesis_instrumentation__.Notify(47807)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(47808)
				}
				__antithesis_instrumentation__.Notify(47806)
				if err := fn(); err != nil {
					__antithesis_instrumentation__.Notify(47809)
					return err
				} else {
					__antithesis_instrumentation__.Notify(47810)
				}
			}
		})
	}
	__antithesis_instrumentation__.Notify(47753)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(47811)
		t.Fatalf("error verifying node values: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(47812)
	}
	__antithesis_instrumentation__.Notify(47754)

	expNodesToSeeFollowerReads := 2
	followerReadsAfter, err := getFollowerReadCounts(ctx, t, c)
	if err != nil {
		__antithesis_instrumentation__.Notify(47813)
		t.Fatalf("failed to get follower read counts: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(47814)
	}
	__antithesis_instrumentation__.Notify(47755)
	nodesWhichSawFollowerReads := 0
	for i := 0; i < len(followerReadsAfter); i++ {
		__antithesis_instrumentation__.Notify(47815)
		if followerReadsAfter[i] > followerReadsBefore[i] {
			__antithesis_instrumentation__.Notify(47816)
			nodesWhichSawFollowerReads++
		} else {
			__antithesis_instrumentation__.Notify(47817)
		}
	}
	__antithesis_instrumentation__.Notify(47756)
	if nodesWhichSawFollowerReads < expNodesToSeeFollowerReads {
		__antithesis_instrumentation__.Notify(47818)
		t.Fatalf("fewer than %v follower reads occurred: saw %v before and %v after",
			expNodesToSeeFollowerReads, followerReadsBefore, followerReadsAfter)
	} else {
		__antithesis_instrumentation__.Notify(47819)
	}
	__antithesis_instrumentation__.Notify(47757)

	liveNodes, deadNodes := make(map[int]struct{}), make(map[int]struct{})
	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(47820)
		if topology.deadPrimaryRegion && func() bool {
			__antithesis_instrumentation__.Notify(47821)
			return i <= 3 == true
		}() == true {
			__antithesis_instrumentation__.Notify(47822)
			stopOpts := option.DefaultStopOpts()
			stopOpts.RoachprodOpts.Sig = 9
			c.Stop(ctx, t.L(), stopOpts, c.Node(i))
			deadNodes[i] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(47823)
			liveNodes[i] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(47758)

	t.L().Printf("starting read load")
	const loadDuration = 4 * time.Minute
	timeoutCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	time.AfterFunc(loadDuration, func() {
		__antithesis_instrumentation__.Notify(47824)
		t.L().Printf("stopping load")
		cancel()
	})
	__antithesis_instrumentation__.Notify(47759)
	g, gCtx = errgroup.WithContext(timeoutCtx)
	const concurrency = 32
	var cur int
	for i := 0; cur < concurrency; i++ {
		__antithesis_instrumentation__.Notify(47825)
		node := i%c.Spec().NodeCount + 1
		if _, ok := liveNodes[node]; ok {
			__antithesis_instrumentation__.Notify(47826)
			g.Go(doSelects(gCtx, node))
			cur++
		} else {
			__antithesis_instrumentation__.Notify(47827)
		}
	}
	__antithesis_instrumentation__.Notify(47760)
	start := timeutil.Now()

	if err := g.Wait(); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(47828)
		return timeoutCtx.Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(47829)
		t.Fatalf("error reading data: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(47830)
	}
	__antithesis_instrumentation__.Notify(47761)
	end := timeutil.Now()
	t.L().Printf("load stopped")

	var expectedLowRatioNodes int
	if !topology.multiRegion {
		__antithesis_instrumentation__.Notify(47831)

		expectedLowRatioNodes = 1
	} else {
		__antithesis_instrumentation__.Notify(47832)

		expectedLowRatioNodes = 2

		if topology.deadPrimaryRegion {
			__antithesis_instrumentation__.Notify(47833)
			expectedLowRatioNodes = 1
		} else {
			__antithesis_instrumentation__.Notify(47834)
		}
	}
	__antithesis_instrumentation__.Notify(47762)
	verifyHighFollowerReadRatios(ctx, t, c, liveNodes, start, end, expectedLowRatioNodes)

	if topology.multiRegion {
		__antithesis_instrumentation__.Notify(47835)

		verifySQLLatency(ctx, t, c, liveNodes, start, end, maxLatencyThreshold)
	} else {
		__antithesis_instrumentation__.Notify(47836)
	}
	__antithesis_instrumentation__.Notify(47763)

	for i := range deadNodes {
		__antithesis_instrumentation__.Notify(47837)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(i))
	}
}

func initFollowerReadsDB(
	ctx context.Context, t test.Test, c cluster.Cluster, topology topologySpec,
) (data map[int]int64) {
	__antithesis_instrumentation__.Notify(47838)
	db := c.Conn(ctx, t.L(), 1)

	_, err := db.ExecContext(ctx, "SET CLUSTER SETTING kv.range_split.by_load_enabled = 'false'")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "SET CLUSTER SETTING kv.range_merge.queue_enabled = 'false'")
	require.NoError(t, err)

	if topology.multiRegion {
		__antithesis_instrumentation__.Notify(47847)
		if err := testutils.SucceedsSoonError(func() error {
			__antithesis_instrumentation__.Notify(47848)
			rows, err := db.QueryContext(ctx, "SELECT region, zones[1] FROM [SHOW REGIONS FROM CLUSTER] ORDER BY 1")
			require.NoError(t, err)
			defer rows.Close()

			matrix, err := sqlutils.RowsToStrMatrix(rows)
			require.NoError(t, err)

			expMatrix := [][]string{
				{"europe-west2", "europe-west2-b"},
				{"us-east1", "us-east1-b"},
				{"us-west1", "us-west1-b"},
			}
			if !reflect.DeepEqual(matrix, expMatrix) {
				__antithesis_instrumentation__.Notify(47850)
				return errors.Errorf("unexpected cluster regions: want %+v, got %+v", expMatrix, matrix)
			} else {
				__antithesis_instrumentation__.Notify(47851)
			}
			__antithesis_instrumentation__.Notify(47849)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(47852)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47853)
		}
	} else {
		__antithesis_instrumentation__.Notify(47854)
	}
	__antithesis_instrumentation__.Notify(47839)

	_, err = db.ExecContext(ctx, `CREATE DATABASE test`)
	require.NoError(t, err)
	if topology.multiRegion {
		__antithesis_instrumentation__.Notify(47855)
		_, err = db.ExecContext(ctx, `ALTER DATABASE test SET PRIMARY REGION "us-east1"`)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `ALTER DATABASE test ADD REGION "us-west1"`)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, `ALTER DATABASE test ADD REGION "europe-west2"`)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE test SURVIVE %s FAILURE`, topology.survival))
		require.NoError(t, err)
	} else {
		__antithesis_instrumentation__.Notify(47856)
	}
	__antithesis_instrumentation__.Notify(47840)
	_, err = db.ExecContext(ctx, `CREATE TABLE test.test ( k INT8, v INT8, PRIMARY KEY (k) )`)
	require.NoError(t, err)
	if topology.multiRegion {
		__antithesis_instrumentation__.Notify(47857)
		_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE test.test SET LOCALITY %s`, topology.locality))
		require.NoError(t, err)
	} else {
		__antithesis_instrumentation__.Notify(47858)
	}
	__antithesis_instrumentation__.Notify(47841)

	t.L().Printf("waiting for up-replication...")
	retryOpts := retry.Options{MaxBackoff: 15 * time.Second}
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(47859)

		var votersCol, nonVotersCol string
		if topology.multiRegion {
			__antithesis_instrumentation__.Notify(47864)
			votersCol = "coalesce(array_length(voting_replicas, 1), 0)"
			nonVotersCol = "coalesce(array_length(non_voting_replicas, 1), 0)"
		} else {
			__antithesis_instrumentation__.Notify(47865)

			votersCol = "coalesce(array_length(replicas, 1), 0)"
			nonVotersCol = "0"
		}
		__antithesis_instrumentation__.Notify(47860)

		q1 := fmt.Sprintf(`
			SELECT
				%s, %s
			FROM
				crdb_internal.ranges_no_leases
			WHERE
				table_name = 'test'`, votersCol, nonVotersCol)

		var voters, nonVoters int
		err := db.QueryRowContext(ctx, q1).Scan(&voters, &nonVoters)
		if errors.Is(err, gosql.ErrNoRows) {
			__antithesis_instrumentation__.Notify(47866)
			t.L().Printf("up-replication not complete, missing range")
			continue
		} else {
			__antithesis_instrumentation__.Notify(47867)
		}
		__antithesis_instrumentation__.Notify(47861)
		require.NoError(t, err)

		var ok bool
		if !topology.multiRegion {
			__antithesis_instrumentation__.Notify(47868)
			ok = voters == 3
		} else {
			__antithesis_instrumentation__.Notify(47869)
			if topology.survival == zone {
				__antithesis_instrumentation__.Notify(47870)

				ok = voters == 3 && func() bool {
					__antithesis_instrumentation__.Notify(47871)
					return nonVoters == 2 == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(47872)

				ok = voters == 5 && func() bool {
					__antithesis_instrumentation__.Notify(47873)
					return nonVoters == 0 == true
				}() == true
			}
		}
		__antithesis_instrumentation__.Notify(47862)
		if ok {
			__antithesis_instrumentation__.Notify(47874)
			break
		} else {
			__antithesis_instrumentation__.Notify(47875)
		}
		__antithesis_instrumentation__.Notify(47863)

		t.L().Printf("up-replication not complete, found %d voters and %d non_voters", voters, nonVoters)
	}
	__antithesis_instrumentation__.Notify(47842)

	if topology.multiRegion {
		__antithesis_instrumentation__.Notify(47876)
		for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
			__antithesis_instrumentation__.Notify(47878)

			const q2 = `
			SELECT
				count(distinct substring(unnest(replica_localities), 'region=([^,]*)'))
			FROM
				crdb_internal.ranges_no_leases
			WHERE
				table_name = 'test'`

			var distinctRegions int
			require.NoError(t, db.QueryRowContext(ctx, q2).Scan(&distinctRegions))
			if distinctRegions == 3 {
				__antithesis_instrumentation__.Notify(47880)
				break
			} else {
				__antithesis_instrumentation__.Notify(47881)
			}
			__antithesis_instrumentation__.Notify(47879)

			t.L().Printf("rebalancing not complete, table in %d regions", distinctRegions)
		}
		__antithesis_instrumentation__.Notify(47877)

		if topology.deadPrimaryRegion {
			__antithesis_instrumentation__.Notify(47882)
			for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
				__antithesis_instrumentation__.Notify(47883)

				WaitForUpdatedReplicationReport(ctx, t, db)

				var expAtRisk int
				if topology.survival == zone {
					__antithesis_instrumentation__.Notify(47886)

					expAtRisk = 1
				} else {
					__antithesis_instrumentation__.Notify(47887)

					expAtRisk = 0
				}
				__antithesis_instrumentation__.Notify(47884)

				const q3 = `
				SELECT
					coalesce(sum(at_risk_ranges), 0)
				FROM
					system.replication_critical_localities
				WHERE
					locality LIKE '%region%'
				AND
					locality NOT LIKE '%zone%'`

				var atRisk int
				require.NoError(t, db.QueryRowContext(ctx, q3).Scan(&atRisk))
				if atRisk == expAtRisk {
					__antithesis_instrumentation__.Notify(47888)
					break
				} else {
					__antithesis_instrumentation__.Notify(47889)
				}
				__antithesis_instrumentation__.Notify(47885)

				t.L().Printf("rebalancing not complete, expected %d at risk ranges, "+
					"found %d", expAtRisk, atRisk)
			}
		} else {
			__antithesis_instrumentation__.Notify(47890)
		}
	} else {
		__antithesis_instrumentation__.Notify(47891)
	}
	__antithesis_instrumentation__.Notify(47843)

	const rows = 100
	const concurrency = 32
	sem := make(chan struct{}, concurrency)
	data = make(map[int]int64)
	insert := func(ctx context.Context, k int) func() error {
		__antithesis_instrumentation__.Notify(47892)
		v := rand.Int63()
		data[k] = v
		return func() error {
			__antithesis_instrumentation__.Notify(47893)
			sem <- struct{}{}
			defer func() { __antithesis_instrumentation__.Notify(47895); <-sem }()
			__antithesis_instrumentation__.Notify(47894)
			_, err := db.ExecContext(ctx, "INSERT INTO test.test VALUES ( $1, $2 )", k, v)
			return err
		}
	}
	__antithesis_instrumentation__.Notify(47844)

	g, gCtx := errgroup.WithContext(ctx)
	for i := 0; i < rows; i++ {
		__antithesis_instrumentation__.Notify(47896)
		g.Go(insert(gCtx, i))
	}
	__antithesis_instrumentation__.Notify(47845)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(47897)
		t.Fatalf("failed to insert data: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(47898)
	}
	__antithesis_instrumentation__.Notify(47846)

	return data
}

func computeFollowerReadDuration(ctx context.Context, db *gosql.DB) (time.Duration, error) {
	__antithesis_instrumentation__.Notify(47899)
	var d pgtype.Interval
	err := db.QueryRowContext(ctx, "SELECT now() - follower_read_timestamp()").Scan(&d)
	if err != nil {
		__antithesis_instrumentation__.Notify(47902)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(47903)
	}
	__antithesis_instrumentation__.Notify(47900)
	var lag time.Duration
	err = d.AssignTo(&lag)
	if err != nil {
		__antithesis_instrumentation__.Notify(47904)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(47905)
	}
	__antithesis_instrumentation__.Notify(47901)
	return lag, nil
}

func verifySQLLatency(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	liveNodes map[int]struct{},
	start, end time.Time,
	targetLatency time.Duration,
) {
	__antithesis_instrumentation__.Notify(47906)

	var adminNode int
	for i := range liveNodes {
		__antithesis_instrumentation__.Notify(47913)
		adminNode = i
		break
	}
	__antithesis_instrumentation__.Notify(47907)
	adminURLs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(adminNode))
	if err != nil {
		__antithesis_instrumentation__.Notify(47914)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47915)
	}
	__antithesis_instrumentation__.Notify(47908)
	url := "http://" + adminURLs[0] + "/ts/query"
	var sources []string
	for i := range liveNodes {
		__antithesis_instrumentation__.Notify(47916)
		sources = append(sources, strconv.Itoa(i))
	}
	__antithesis_instrumentation__.Notify(47909)
	request := tspb.TimeSeriesQueryRequest{
		StartNanos: start.UnixNano(),
		EndNanos:   end.UnixNano(),

		SampleNanos: (10 * time.Second).Nanoseconds(),
		Queries: []tspb.Query{{
			Name:    "cr.node.sql.service.latency-p90",
			Sources: sources,
		}},
	}
	var response tspb.TimeSeriesQueryResponse
	if err := httputil.PostJSON(http.Client{}, url, &request, &response); err != nil {
		__antithesis_instrumentation__.Notify(47917)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47918)
	}
	__antithesis_instrumentation__.Notify(47910)
	perTenSeconds := response.Results[0].Datapoints

	if len(perTenSeconds) < 3 {
		__antithesis_instrumentation__.Notify(47919)
		t.Fatalf("not enough ts data to verify latency")
	} else {
		__antithesis_instrumentation__.Notify(47920)
	}
	__antithesis_instrumentation__.Notify(47911)
	perTenSeconds = perTenSeconds[2:]
	var above []time.Duration
	for _, dp := range perTenSeconds {
		__antithesis_instrumentation__.Notify(47921)
		if val := time.Duration(dp.Value); val > targetLatency {
			__antithesis_instrumentation__.Notify(47922)
			above = append(above, val)
		} else {
			__antithesis_instrumentation__.Notify(47923)
		}
	}
	__antithesis_instrumentation__.Notify(47912)
	if permitted := int(.2 * float64(len(perTenSeconds))); len(above) > permitted {
		__antithesis_instrumentation__.Notify(47924)
		t.Fatalf("%d latency values (%v) are above target latency %v, %d permitted",
			len(above), above, targetLatency, permitted)
	} else {
		__antithesis_instrumentation__.Notify(47925)
	}
}

func verifyHighFollowerReadRatios(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	liveNodes map[int]struct{},
	start, end time.Time,
	toleratedNodes int,
) {
	__antithesis_instrumentation__.Notify(47926)

	start = start.Add(10 * time.Second)

	var adminNode int
	for i := range liveNodes {
		__antithesis_instrumentation__.Notify(47934)
		adminNode = i
		break
	}
	__antithesis_instrumentation__.Notify(47927)
	adminURLs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(adminNode))
	require.NoError(t, err)
	url := "http://" + adminURLs[0] + "/ts/query"
	request := tspb.TimeSeriesQueryRequest{
		StartNanos: start.UnixNano(),
		EndNanos:   end.UnixNano(),

		SampleNanos: (10 * time.Second).Nanoseconds(),
	}
	for i := range liveNodes {
		__antithesis_instrumentation__.Notify(47935)
		nodeID := strconv.Itoa(i)
		request.Queries = append(request.Queries, tspb.Query{
			Name:       "cr.store.follower_reads.success_count",
			Sources:    []string{nodeID},
			Derivative: tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE.Enum(),
		})
		request.Queries = append(request.Queries, tspb.Query{
			Name:       "cr.node.sql.select.count",
			Sources:    []string{nodeID},
			Derivative: tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE.Enum(),
		})
	}
	__antithesis_instrumentation__.Notify(47928)

	var response tspb.TimeSeriesQueryResponse
	if err := httputil.PostJSON(http.Client{}, url, &request, &response); err != nil {
		__antithesis_instrumentation__.Notify(47936)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47937)
	}
	__antithesis_instrumentation__.Notify(47929)

	minDataPoints := len(response.Results[0].Datapoints)
	for _, res := range response.Results[1:] {
		__antithesis_instrumentation__.Notify(47938)
		if len(res.Datapoints) < minDataPoints {
			__antithesis_instrumentation__.Notify(47939)
			minDataPoints = len(res.Datapoints)
		} else {
			__antithesis_instrumentation__.Notify(47940)
		}
	}
	__antithesis_instrumentation__.Notify(47930)
	if minDataPoints < 3 {
		__antithesis_instrumentation__.Notify(47941)
		t.Fatalf("not enough ts data to verify follower reads")
	} else {
		__antithesis_instrumentation__.Notify(47942)
	}
	__antithesis_instrumentation__.Notify(47931)

	stats := make([]intervalStats, minDataPoints)
	for i := range stats {
		__antithesis_instrumentation__.Notify(47943)
		var ratios []float64
		for n := 0; n < len(response.Results); n += 2 {
			__antithesis_instrumentation__.Notify(47945)
			followerReadsPerTenSeconds := response.Results[n].Datapoints[i]
			selectsPerTenSeconds := response.Results[n+1].Datapoints[i]
			ratios = append(ratios, followerReadsPerTenSeconds.Value/selectsPerTenSeconds.Value)
		}
		__antithesis_instrumentation__.Notify(47944)
		intervalEnd := timeutil.Unix(0, response.Results[0].Datapoints[i].TimestampNanos)
		stats[i] = intervalStats{
			ratiosPerNode: ratios,
			start:         intervalEnd.Add(-10 * time.Second),
			end:           intervalEnd,
		}
	}
	__antithesis_instrumentation__.Notify(47932)

	t.L().Printf("interval stats: %s", intervalsToString(stats))

	const threshold = 0.9
	var badIntervals []intervalStats
	for _, stat := range stats {
		__antithesis_instrumentation__.Notify(47946)
		var nodesWithLowRatios int
		for _, ratio := range stat.ratiosPerNode {
			__antithesis_instrumentation__.Notify(47948)
			if ratio < threshold {
				__antithesis_instrumentation__.Notify(47949)
				nodesWithLowRatios++
			} else {
				__antithesis_instrumentation__.Notify(47950)
			}
		}
		__antithesis_instrumentation__.Notify(47947)
		if nodesWithLowRatios > toleratedNodes {
			__antithesis_instrumentation__.Notify(47951)
			badIntervals = append(badIntervals, stat)
		} else {
			__antithesis_instrumentation__.Notify(47952)
		}
	}
	__antithesis_instrumentation__.Notify(47933)
	permitted := int(.2 * float64(len(stats)))
	if len(badIntervals) > permitted {
		__antithesis_instrumentation__.Notify(47953)
		t.Fatalf("too many intervals with more than %d nodes with low follower read ratios: "+
			"%d intervals > %d threshold. Bad intervals:\n%s",
			toleratedNodes, len(badIntervals), permitted, intervalsToString(badIntervals))
	} else {
		__antithesis_instrumentation__.Notify(47954)
	}
}

type intervalStats struct {
	ratiosPerNode []float64
	start, end    time.Time
}

func intervalsToString(stats []intervalStats) string {
	__antithesis_instrumentation__.Notify(47955)
	var s strings.Builder
	for _, interval := range stats {
		__antithesis_instrumentation__.Notify(47957)
		s.WriteString(fmt.Sprintf("interval %s-%s: ",
			interval.start.Format("15:04:05"), interval.end.Format("15:04:05")))
		for i, r := range interval.ratiosPerNode {
			__antithesis_instrumentation__.Notify(47959)
			s.WriteString(fmt.Sprintf("n%d ratio: %.3f ", i+1, r))
		}
		__antithesis_instrumentation__.Notify(47958)
		s.WriteRune('\n')
	}
	__antithesis_instrumentation__.Notify(47956)
	return s.String()
}

const followerReadsMetric = "follower_reads_success_count"

func getFollowerReadCounts(ctx context.Context, t test.Test, c cluster.Cluster) ([]int, error) {
	__antithesis_instrumentation__.Notify(47960)
	followerReadCounts := make([]int, c.Spec().NodeCount)
	getFollowerReadCount := func(ctx context.Context, node int) func() error {
		__antithesis_instrumentation__.Notify(47964)
		return func() error {
			__antithesis_instrumentation__.Notify(47965)
			adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(node))
			if err != nil {
				__antithesis_instrumentation__.Notify(47970)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47971)
			}
			__antithesis_instrumentation__.Notify(47966)
			url := "http://" + adminUIAddrs[0] + "/_status/vars"
			resp, err := httputil.Get(ctx, url)
			if err != nil {
				__antithesis_instrumentation__.Notify(47972)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47973)
			}
			__antithesis_instrumentation__.Notify(47967)
			defer resp.Body.Close()
			if resp.StatusCode != 200 {
				__antithesis_instrumentation__.Notify(47974)
				return errors.Errorf("invalid non-200 status code %v from node %d", resp.StatusCode, node)
			} else {
				__antithesis_instrumentation__.Notify(47975)
			}
			__antithesis_instrumentation__.Notify(47968)
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				__antithesis_instrumentation__.Notify(47976)
				m, ok := parsePrometheusMetric(scanner.Text())
				if ok {
					__antithesis_instrumentation__.Notify(47977)
					if m.metric == followerReadsMetric {
						__antithesis_instrumentation__.Notify(47978)
						v, err := strconv.ParseFloat(m.value, 64)
						if err != nil {
							__antithesis_instrumentation__.Notify(47980)
							return err
						} else {
							__antithesis_instrumentation__.Notify(47981)
						}
						__antithesis_instrumentation__.Notify(47979)
						followerReadCounts[node-1] = int(v)
					} else {
						__antithesis_instrumentation__.Notify(47982)
					}
				} else {
					__antithesis_instrumentation__.Notify(47983)
				}
			}
			__antithesis_instrumentation__.Notify(47969)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(47961)
	g, gCtx := errgroup.WithContext(ctx)
	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(47984)
		g.Go(getFollowerReadCount(gCtx, i))
	}
	__antithesis_instrumentation__.Notify(47962)
	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(47985)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(47986)
	}
	__antithesis_instrumentation__.Notify(47963)
	return followerReadCounts, nil
}

var prometheusMetricStringPattern = `^(?P<metric>\w+)(?:\{` +
	`(?P<labelvalues>(\w+=\".*\",)*(\w+=\".*\")?)\})?\s+(?P<value>.*)$`
var promethusMetricStringRE = regexp.MustCompile(prometheusMetricStringPattern)

type prometheusMetric struct {
	metric      string
	labelValues string
	value       string
}

func parsePrometheusMetric(s string) (*prometheusMetric, bool) {
	__antithesis_instrumentation__.Notify(47987)
	matches := promethusMetricStringRE.FindStringSubmatch(s)
	if matches == nil {
		__antithesis_instrumentation__.Notify(47989)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(47990)
	}
	__antithesis_instrumentation__.Notify(47988)
	return &prometheusMetric{
		metric:      matches[1],
		labelValues: matches[2],
		value:       matches[5],
	}, true
}

func runFollowerReadsMixedVersionSingleRegionTest(
	ctx context.Context, t test.Test, c cluster.Cluster, buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(47991)
	predecessorVersion, err := PredecessorVersion(buildVersion)
	require.NoError(t, err)

	const curVersion = ""

	settings := install.MakeClusterSettings()
	settings.Binary = uploadVersion(ctx, t, c, c.All(), predecessorVersion)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.All())
	topology := topologySpec{multiRegion: false}
	data := initFollowerReadsDB(ctx, t, c, topology)

	randNode := 1 + rand.Intn(c.Spec().NodeCount)
	t.L().Printf("upgrading n%d to current version", randNode)
	nodeToUpgrade := c.Node(randNode)
	upgradeNodes(ctx, nodeToUpgrade, curVersion, t, c)
	runFollowerReadsTest(ctx, t, c, topologySpec{multiRegion: false}, exactStaleness, data)

	var remainingNodes option.NodeListOption
	for i := 0; i < c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(47993)
		if i+1 == randNode {
			__antithesis_instrumentation__.Notify(47995)
			continue
		} else {
			__antithesis_instrumentation__.Notify(47996)
		}
		__antithesis_instrumentation__.Notify(47994)
		remainingNodes = remainingNodes.Merge(c.Node(i + 1))
	}
	__antithesis_instrumentation__.Notify(47992)
	t.L().Printf("upgrading nodes %s to current version", remainingNodes)
	upgradeNodes(ctx, remainingNodes, curVersion, t, c)
	runFollowerReadsTest(ctx, t, c, topologySpec{multiRegion: false}, exactStaleness, data)
}
