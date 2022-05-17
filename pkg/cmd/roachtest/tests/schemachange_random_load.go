package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

type randomLoadBenchSpec struct {
	Nodes       int
	Ops         int
	Concurrency int
}

func registerSchemaChangeRandomLoad(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50701)
	geoZones := []string{"us-east1-b", "us-west1-b", "europe-west2-b"}
	if r.MakeClusterSpec(1).Cloud == spec.AWS {
		__antithesis_instrumentation__.Notify(50704)
		geoZones = []string{"us-east-2b", "us-west-1a", "eu-west-1a"}
	} else {
		__antithesis_instrumentation__.Notify(50705)
	}
	__antithesis_instrumentation__.Notify(50702)
	geoZonesStr := strings.Join(geoZones, ",")
	r.Add(registry.TestSpec{
		Name:  "schemachange/random-load",
		Owner: registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(
			3,
			spec.Geo(),
			spec.Zones(geoZonesStr),
		),

		NonReleaseBlocker: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50706)
			maxOps := 5000
			concurrency := 20
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(50708)
				maxOps = 200
				concurrency = 2
			} else {
				__antithesis_instrumentation__.Notify(50709)
			}
			__antithesis_instrumentation__.Notify(50707)
			runSchemaChangeRandomLoad(ctx, t, c, maxOps, concurrency)
		},
	})
	__antithesis_instrumentation__.Notify(50703)

	registerRandomLoadBenchSpec(r, randomLoadBenchSpec{
		Nodes:       3,
		Ops:         2000,
		Concurrency: 1,
	})

	registerRandomLoadBenchSpec(r, randomLoadBenchSpec{
		Nodes:       3,
		Ops:         10000,
		Concurrency: 20,
	})
}

func registerRandomLoadBenchSpec(r registry.Registry, b randomLoadBenchSpec) {
	__antithesis_instrumentation__.Notify(50710)
	nameParts := []string{
		"scbench",
		"randomload",
		fmt.Sprintf("nodes=%d", b.Nodes),
		fmt.Sprintf("ops=%d", b.Ops),
		fmt.Sprintf("conc=%d", b.Concurrency),
	}
	name := strings.Join(nameParts, "/")

	r.Add(registry.TestSpec{
		Name:    name,
		Owner:   registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(b.Nodes),
		Skip:    "https://github.com/cockroachdb/cockroach/issues/56230",

		NonReleaseBlocker: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50711)
			runSchemaChangeRandomLoad(ctx, t, c, b.Ops, b.Concurrency)
		},
	})
}

func runSchemaChangeRandomLoad(
	ctx context.Context, t test.Test, c cluster.Cluster, maxOps, concurrency int,
) {
	__antithesis_instrumentation__.Notify(50712)
	validate := func(db *gosql.DB) {
		__antithesis_instrumentation__.Notify(50717)
		var (
			id           int
			databaseName string
			schemaName   string
			objName      string
			objError     string
		)
		numInvalidObjects := 0
		rows, err := db.QueryContext(ctx, `SELECT id, database_name, schema_name, obj_name, error FROM crdb_internal.invalid_objects`)
		if err != nil {
			__antithesis_instrumentation__.Notify(50721)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50722)
		}
		__antithesis_instrumentation__.Notify(50718)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(50723)
			numInvalidObjects++
			if err := rows.Scan(&id, &databaseName, &schemaName, &objName, &objError); err != nil {
				__antithesis_instrumentation__.Notify(50725)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50726)
			}
			__antithesis_instrumentation__.Notify(50724)
			t.L().Errorf(
				"invalid object found: id: %d, database_name: %s, schema_name: %s, obj_name: %s, error: %s",
				id, databaseName, schemaName, objName, objError,
			)
		}
		__antithesis_instrumentation__.Notify(50719)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(50727)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50728)
		}
		__antithesis_instrumentation__.Notify(50720)
		if numInvalidObjects > 0 {
			__antithesis_instrumentation__.Notify(50729)
			t.Fatalf("found %d invalid objects", numInvalidObjects)
		} else {
			__antithesis_instrumentation__.Notify(50730)
		}
	}
	__antithesis_instrumentation__.Notify(50713)

	loadNode := c.Node(1)
	roachNodes := c.Range(1, c.Spec().NodeCount)
	t.Status("copying binaries")
	c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", loadNode)

	t.Status("starting cockroach nodes")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)
	c.Run(ctx, loadNode, "./workload init schemachange")

	result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(1), "echo", "-n", "{store-dir}")
	if err != nil {
		__antithesis_instrumentation__.Notify(50731)
		t.L().Printf("Failed to retrieve store directory from node 1: %v\n", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(50732)
	}
	__antithesis_instrumentation__.Notify(50714)
	storeDirectory := result.Stdout

	runCmd := []string{
		"./workload run schemachange --verbose=1",
		"--tolerate-errors=false",

		" --histograms=" + t.PerfArtifactsDir() + "/stats.json",
		fmt.Sprintf("--max-ops %d", maxOps),
		fmt.Sprintf("--concurrency %d", concurrency),
		fmt.Sprintf("--txn-log %s", filepath.Join(storeDirectory, "transactions.json")),
	}
	t.Status("running schemachange workload")
	err = c.RunE(ctx, loadNode, runCmd...)
	if err != nil {
		__antithesis_instrumentation__.Notify(50733)
		saveArtifacts(ctx, t, c, storeDirectory)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50734)
	}
	__antithesis_instrumentation__.Notify(50715)

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	t.Status("performing validation after workload")
	validate(db)
	t.Status("dropping database")
	_, err = db.ExecContext(ctx, `USE defaultdb; DROP DATABASE schemachange CASCADE;`)
	if err != nil {
		__antithesis_instrumentation__.Notify(50735)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50736)
	}
	__antithesis_instrumentation__.Notify(50716)
	t.Status("performing validation after dropping database")
	validate(db)
}

func saveArtifacts(ctx context.Context, t test.Test, c cluster.Cluster, storeDirectory string) {
	__antithesis_instrumentation__.Notify(50737)
	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	_, err := db.Exec("BACKUP DATABASE schemachange to 'nodelocal://1/schemachange'")
	if err != nil {
		__antithesis_instrumentation__.Notify(50740)
		t.L().Printf("Failed execute backup command on node 1: %v\n", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(50741)
	}
	__antithesis_instrumentation__.Notify(50738)

	remoteBackupFilePath := filepath.Join(storeDirectory, "extern", "schemachange")
	localBackupFilePath := filepath.Join(t.ArtifactsDir(), "backup")
	remoteTransactionsFilePath := filepath.Join(storeDirectory, "transactions.ndjson")
	localTransactionsFilePath := filepath.Join(t.ArtifactsDir(), "transactions.ndjson")

	err = c.Get(ctx, t.L(), remoteBackupFilePath, localBackupFilePath, c.Node(1))
	if err != nil {
		__antithesis_instrumentation__.Notify(50742)
		t.L().Printf("Failed to copy backup file from node 1 to artifacts directory: %v\n", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(50743)
	}
	__antithesis_instrumentation__.Notify(50739)

	err = c.Get(ctx, t.L(), remoteTransactionsFilePath, localTransactionsFilePath, c.Node(1))
	if err != nil {
		__antithesis_instrumentation__.Notify(50744)
		t.L().Printf("Failed to copy txn log file from node 1 to artifacts directory: %v\n", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(50745)
	}
}
