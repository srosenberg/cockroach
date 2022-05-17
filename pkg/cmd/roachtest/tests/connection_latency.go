package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

const (
	regionUsEast    = "us-east1-b"
	regionUsCentral = "us-central1-b"
	regionUsWest    = "us-west1-b"
	regionEuWest    = "europe-west2-b"
)

func runConnectionLatencyTest(
	ctx context.Context, t test.Test, c cluster.Cluster, numNodes int, numZones int, password bool,
) {
	__antithesis_instrumentation__.Notify(46809)
	err := c.PutE(ctx, t.L(), t.Cockroach(), "./cockroach")
	require.NoError(t, err)

	err = c.PutE(ctx, t.L(), t.DeprecatedWorkload(), "./workload")
	require.NoError(t, err)

	settings := install.MakeClusterSettings(install.SecureOption(true))
	err = c.StartE(ctx, t.L(), option.DefaultStartOpts(), settings)
	require.NoError(t, err)

	var passwordFlag string

	t.L().Printf("creating testuser")
	if password {
		__antithesis_instrumentation__.Notify(46812)
		err = c.RunE(ctx, c.Node(1), `./cockroach sql --certs-dir certs -e "CREATE USER testuser WITH PASSWORD '123' CREATEDB"`)
		require.NoError(t, err)
		err = c.RunE(ctx, c.Node(1), "./workload init connectionlatency --user testuser --password '123' --secure")
		require.NoError(t, err)
		passwordFlag = "--password 123 "
	} else {
		__antithesis_instrumentation__.Notify(46813)

		err = c.RunE(ctx, c.Node(1), `./cockroach sql --certs-dir certs -e "CREATE USER testuser CREATEDB"`)
		require.NoError(t, err)
		require.NoError(t, err)
		err = c.RunE(ctx, c.Node(1), "./workload init connectionlatency --user testuser --secure")
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(46810)

	runWorkload := func(roachNodes, loadNode option.NodeListOption, locality string) {
		__antithesis_instrumentation__.Notify(46814)
		var urlString string
		var urls []string
		externalIps, err := c.ExternalIP(ctx, t.L(), roachNodes)
		require.NoError(t, err)

		if password {
			__antithesis_instrumentation__.Notify(46816)
			urlTemplate := "postgres://testuser:123@%s:26257?sslmode=require&sslrootcert=certs/ca.crt"
			for _, u := range externalIps {
				__antithesis_instrumentation__.Notify(46818)
				url := fmt.Sprintf(urlTemplate, u)
				urls = append(urls, fmt.Sprintf("'%s'", url))
			}
			__antithesis_instrumentation__.Notify(46817)
			urlString = strings.Join(urls, " ")
		} else {
			__antithesis_instrumentation__.Notify(46819)
			urlTemplate := "postgres://testuser@%s:26257?sslcert=certs/client.testuser.crt&sslkey=certs/client.testuser.key&sslrootcert=certs/ca.crt&sslmode=require"
			for _, u := range externalIps {
				__antithesis_instrumentation__.Notify(46821)
				url := fmt.Sprintf(urlTemplate, u)
				urls = append(urls, fmt.Sprintf("'%s'", url))
			}
			__antithesis_instrumentation__.Notify(46820)
			urlString = strings.Join(urls, " ")
		}
		__antithesis_instrumentation__.Notify(46815)

		t.L().Printf("running workload in %q against urls:\n%s", locality, strings.Join(urls, "\n"))

		workloadCmd := fmt.Sprintf(
			`./workload run connectionlatency %s --user testuser --secure %s --duration 30s --histograms=%s/stats.json --locality %s`,
			urlString,
			passwordFlag,
			t.PerfArtifactsDir(),
			locality,
		)
		err = c.RunE(ctx, loadNode, workloadCmd)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(46811)

	if numZones > 1 {
		__antithesis_instrumentation__.Notify(46822)
		numLoadNodes := numZones
		loadGroups := makeLoadGroups(c, numZones, numNodes, numLoadNodes)
		cockroachUsEast := loadGroups[0].loadNodes
		cockroachUsWest := loadGroups[1].loadNodes
		cockroachEuWest := loadGroups[2].loadNodes

		runWorkload(loadGroups[0].roachNodes, cockroachUsEast, regionUsEast)
		runWorkload(loadGroups[1].roachNodes, cockroachUsWest, regionUsWest)
		runWorkload(loadGroups[2].roachNodes, cockroachEuWest, regionEuWest)
	} else {
		__antithesis_instrumentation__.Notify(46823)

		runWorkload(c.Range(1, numNodes), c.Node(numNodes+1), regionUsCentral)
	}
}

func registerConnectionLatencyTest(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46824)

	numNodes := 3
	r.Add(registry.TestSpec{
		Name:  fmt.Sprintf("connection_latency/nodes=%d/certs", numNodes),
		Owner: registry.OwnerSQLExperience,

		Cluster: r.MakeClusterSpec(numNodes+1, spec.Zones(regionUsCentral)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46827)
			runConnectionLatencyTest(ctx, t, c, numNodes, 1, false)
		},
	})
	__antithesis_instrumentation__.Notify(46825)

	geoZones := []string{regionUsEast, regionUsWest, regionEuWest}
	geoZonesStr := strings.Join(geoZones, ",")
	numMultiRegionNodes := 9
	numZones := len(geoZones)
	loadNodes := numZones

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("connection_latency/nodes=%d/multiregion/certs", numMultiRegionNodes),
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(numMultiRegionNodes+loadNodes, spec.Geo(), spec.Zones(geoZonesStr)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46828)
			runConnectionLatencyTest(ctx, t, c, numMultiRegionNodes, numZones, false)
		},
	})
	__antithesis_instrumentation__.Notify(46826)

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("connection_latency/nodes=%d/multiregion/password", numMultiRegionNodes),
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(numMultiRegionNodes+loadNodes, spec.Geo(), spec.Zones(geoZonesStr)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(46829)
			runConnectionLatencyTest(ctx, t, c, numMultiRegionNodes, numZones, true)
		},
	})
}
