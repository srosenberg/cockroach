package multiregionccltestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/testutils/testcluster"
	"github.com/cockroachdb/errors"
)

type multiRegionTestClusterParams struct {
	baseDir         string
	replicationMode base.TestClusterReplicationMode
	useDatabase     string
}

type MultiRegionTestClusterParamsOption func(params *multiRegionTestClusterParams)

func WithBaseDirectory(baseDir string) MultiRegionTestClusterParamsOption {
	__antithesis_instrumentation__.Notify(19927)
	return func(params *multiRegionTestClusterParams) {
		__antithesis_instrumentation__.Notify(19928)
		params.baseDir = baseDir
	}
}

func WithReplicationMode(
	replicationMode base.TestClusterReplicationMode,
) MultiRegionTestClusterParamsOption {
	__antithesis_instrumentation__.Notify(19929)
	return func(params *multiRegionTestClusterParams) {
		__antithesis_instrumentation__.Notify(19930)
		params.replicationMode = replicationMode
	}
}

func WithUseDatabase(db string) MultiRegionTestClusterParamsOption {
	__antithesis_instrumentation__.Notify(19931)
	return func(params *multiRegionTestClusterParams) {
		__antithesis_instrumentation__.Notify(19932)
		params.useDatabase = db
	}
}

func TestingCreateMultiRegionCluster(
	t testing.TB, numServers int, knobs base.TestingKnobs, opts ...MultiRegionTestClusterParamsOption,
) (*testcluster.TestCluster, *gosql.DB, func()) {
	__antithesis_instrumentation__.Notify(19933)
	serverArgs := make(map[int]base.TestServerArgs)
	regionNames := make([]string, numServers)
	for i := 0; i < numServers; i++ {
		__antithesis_instrumentation__.Notify(19938)

		regionNames[i] = fmt.Sprintf("us-east%d", i+1)
	}
	__antithesis_instrumentation__.Notify(19934)

	params := &multiRegionTestClusterParams{}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(19939)
		opt(params)
	}
	__antithesis_instrumentation__.Notify(19935)

	for i := 0; i < numServers; i++ {
		__antithesis_instrumentation__.Notify(19940)
		serverArgs[i] = base.TestServerArgs{
			Knobs:         knobs,
			ExternalIODir: params.baseDir,
			UseDatabase:   params.useDatabase,
			Locality: roachpb.Locality{
				Tiers: []roachpb.Tier{{Key: "region", Value: regionNames[i]}},
			},
		}
	}
	__antithesis_instrumentation__.Notify(19936)

	tc := testcluster.StartTestCluster(t, numServers, base.TestClusterArgs{
		ReplicationMode:   params.replicationMode,
		ServerArgsPerNode: serverArgs,
	})

	ctx := context.Background()
	cleanup := func() {
		__antithesis_instrumentation__.Notify(19941)
		tc.Stopper().Stop(ctx)
	}
	__antithesis_instrumentation__.Notify(19937)

	sqlDB := tc.ServerConn(0)

	return tc, sqlDB, cleanup
}

func TestingEnsureCorrectPartitioning(
	sqlDB *gosql.DB, dbName string, tableName string, expectedIndexes []string,
) error {
	__antithesis_instrumentation__.Notify(19942)
	rows, err := sqlDB.Query("SELECT region FROM [SHOW REGIONS FROM DATABASE db] ORDER BY region")
	if err != nil {
		__antithesis_instrumentation__.Notify(19948)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19949)
	}
	__antithesis_instrumentation__.Notify(19943)
	defer rows.Close()
	var expectedPartitions []string
	for rows.Next() {
		__antithesis_instrumentation__.Notify(19950)
		var regionName string
		if err := rows.Scan(&regionName); err != nil {
			__antithesis_instrumentation__.Notify(19952)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19953)
		}
		__antithesis_instrumentation__.Notify(19951)
		expectedPartitions = append(expectedPartitions, regionName)
	}
	__antithesis_instrumentation__.Notify(19944)

	rows, err = sqlDB.Query(
		fmt.Sprintf(
			"SELECT index_name, partition_name FROM [SHOW PARTITIONS FROM TABLE %s.%s] ORDER BY partition_name",
			dbName,
			tableName,
		),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(19954)
		return err
	} else {
		__antithesis_instrumentation__.Notify(19955)
	}
	__antithesis_instrumentation__.Notify(19945)
	defer rows.Close()

	indexPartitions := make(map[string][]string)
	for rows.Next() {
		__antithesis_instrumentation__.Notify(19956)
		var indexName string
		var partitionName string
		if err := rows.Scan(&indexName, &partitionName); err != nil {
			__antithesis_instrumentation__.Notify(19958)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19959)
		}
		__antithesis_instrumentation__.Notify(19957)

		indexPartitions[indexName] = append(indexPartitions[indexName], partitionName)
	}
	__antithesis_instrumentation__.Notify(19946)

	for _, expectedIndex := range expectedIndexes {
		__antithesis_instrumentation__.Notify(19960)
		partitions, found := indexPartitions[expectedIndex]
		if !found {
			__antithesis_instrumentation__.Notify(19963)
			return errors.AssertionFailedf("did not find index %s", expectedIndex)
		} else {
			__antithesis_instrumentation__.Notify(19964)
		}
		__antithesis_instrumentation__.Notify(19961)

		if len(partitions) != len(expectedPartitions) {
			__antithesis_instrumentation__.Notify(19965)
			return errors.AssertionFailedf(
				"unexpected number of partitions; expected %d, found %d",
				len(partitions),
				len(expectedPartitions),
			)
		} else {
			__antithesis_instrumentation__.Notify(19966)
		}
		__antithesis_instrumentation__.Notify(19962)
		for i, expectedPartition := range expectedPartitions {
			__antithesis_instrumentation__.Notify(19967)
			if expectedPartition != partitions[i] {
				__antithesis_instrumentation__.Notify(19968)
				return errors.AssertionFailedf(
					"unexpected partitions; expected %v, found %v",
					expectedPartitions,
					partitions,
				)
			} else {
				__antithesis_instrumentation__.Notify(19969)
			}
		}
	}
	__antithesis_instrumentation__.Notify(19947)
	return nil
}
