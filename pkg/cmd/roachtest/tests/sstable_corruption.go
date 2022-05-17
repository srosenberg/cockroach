package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

func runSSTableCorruption(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(51156)
	crdbNodes := c.Range(1, c.Spec().NodeCount)
	workloadNode := c.Node(1)

	corruptNodes := crdbNodes

	t.Status("installing cockroach")
	c.Put(ctx, t.Cockroach(), "./cockroach", crdbNodes)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes)

	{
		__antithesis_instrumentation__.Notify(51160)
		m := c.NewMonitor(ctx, crdbNodes)

		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(51162)

			t.Status("importing tpcc fixture")
			c.Run(ctx, workloadNode,
				"./cockroach workload fixtures import tpcc --warehouses=100 --fks=false --checks=false")
			return nil
		})
		__antithesis_instrumentation__.Notify(51161)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(51157)

	opts := option.DefaultStopOpts()
	opts.RoachprodOpts.Wait = true
	c.Stop(ctx, t.L(), opts, crdbNodes)

	const nTables = 6
	var dumpManifestCmd = "" +

		"ls -tr {store-dir}/MANIFEST-* | tail -n1 | " +

		"xargs ./cockroach debug pebble manifest dump | " +

		"grep -v added | grep -v deleted | grep '/Table/'"
	var findTablesCmd = dumpManifestCmd + "| " +

		"shuf | " +

		fmt.Sprintf("tail -n %d", nTables)

	for _, node := range corruptNodes {
		__antithesis_instrumentation__.Notify(51163)
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(node), findTablesCmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(51166)
			t.Fatalf("could not find tables to corrupt: %s\nstdout: %s\nstderr: %s", err, result.Stdout, result.Stderr)
		} else {
			__antithesis_instrumentation__.Notify(51167)
		}
		__antithesis_instrumentation__.Notify(51164)
		tableSSTs := strings.Split(strings.TrimSpace(result.Stdout), "\n")
		if len(tableSSTs) != nTables {
			__antithesis_instrumentation__.Notify(51168)

			cmd := "ls -l {store-dir}"
			result, err = c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(node), cmd)
			if err == nil {
				__antithesis_instrumentation__.Notify(51173)
				t.Status("store dir contents:\n", result.Stdout)
			} else {
				__antithesis_instrumentation__.Notify(51174)
			}
			__antithesis_instrumentation__.Notify(51169)

			result, err = c.RunWithDetailsSingleNode(
				ctx, t.L(), c.Node(node),
				"tar czf {store-dir}/manifests.tar.gz {store-dir}/MANIFEST-*",
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(51175)
				t.Fatalf("could not create manifest file archive: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(51176)
			}
			__antithesis_instrumentation__.Notify(51170)
			result, err = c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(node), "echo", "-n", "{store-dir}")
			if err != nil {
				__antithesis_instrumentation__.Notify(51177)
				t.Fatalf("could not infer store directory: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(51178)
			}
			__antithesis_instrumentation__.Notify(51171)
			storeDirectory := result.Stdout
			srcPath := filepath.Join(storeDirectory, "manifests.tar.gz")
			dstPath := filepath.Join(t.ArtifactsDir(), fmt.Sprintf("manifests.%d.tar.gz", node))
			err = c.Get(ctx, t.L(), srcPath, dstPath, c.Node(node))
			if err != nil {
				__antithesis_instrumentation__.Notify(51179)
				t.Fatalf("could not fetch manifest archive: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(51180)
			}
			__antithesis_instrumentation__.Notify(51172)
			t.Fatalf(
				"expected %d SSTables containing table keys, got %d: %s",
				nTables, len(tableSSTs), tableSSTs,
			)
		} else {
			__antithesis_instrumentation__.Notify(51181)
		}
		__antithesis_instrumentation__.Notify(51165)

		for _, sstLine := range tableSSTs {
			__antithesis_instrumentation__.Notify(51182)
			sstLine = strings.TrimSpace(sstLine)
			firstFileIdx := strings.Index(sstLine, ":")
			if firstFileIdx < 0 {
				__antithesis_instrumentation__.Notify(51185)
				t.Fatalf("unexpected format for sst line: %q", sstLine)
			} else {
				__antithesis_instrumentation__.Notify(51186)
			}
			__antithesis_instrumentation__.Notify(51183)
			_, err = strconv.Atoi(sstLine[:firstFileIdx])
			if err != nil {
				__antithesis_instrumentation__.Notify(51187)
				t.Fatalf("error when converting %s to int: %s", sstLine[:firstFileIdx], err.Error())
			} else {
				__antithesis_instrumentation__.Notify(51188)
			}
			__antithesis_instrumentation__.Notify(51184)

			t.Status(fmt.Sprintf("corrupting sstable %s on node %d", sstLine[:firstFileIdx], node))
			c.Run(ctx, c.Node(node), fmt.Sprintf("dd if=/dev/urandom of={store-dir}/%s.sst seek=256 count=128 bs=1 conv=notrunc", sstLine[:firstFileIdx]))
		}
	}
	__antithesis_instrumentation__.Notify(51158)

	if err := c.StartE(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), crdbNodes); err != nil {
		__antithesis_instrumentation__.Notify(51189)

		_ = c.WipeE(ctx, t.L(), corruptNodes)
		return
	} else {
		__antithesis_instrumentation__.Notify(51190)
	}

	{
		__antithesis_instrumentation__.Notify(51191)
		m := c.NewMonitor(ctx)

		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(51193)
			const timeout = 10 * time.Minute
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			for {
				__antithesis_instrumentation__.Notify(51194)
				err := errors.CombineErrors(
					c.RunE(
						ctx, workloadNode,
						"./cockroach workload run tpcc --warehouses=100 --tolerate-errors",
					),
					errors.New("workload unexpectedly returned nil"),
				)

				if err != nil {
					__antithesis_instrumentation__.Notify(51196)
					t.L().Printf("workload failed: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(51197)
				}
				__antithesis_instrumentation__.Notify(51195)
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(51198)
					return ctx.Err()
				case <-time.After(time.Second):
					__antithesis_instrumentation__.Notify(51199)

				}
			}
		})
		__antithesis_instrumentation__.Notify(51192)

		t.L().Printf("waiting for monitor to observe error ...")
		err := m.WaitE()
		t.L().Printf("monitor observed error: %s", err)
		if errors.Is(err, context.DeadlineExceeded) || func() bool {
			__antithesis_instrumentation__.Notify(51200)
			return errors.Is(err, context.Canceled) == true
		}() == true {
			__antithesis_instrumentation__.Notify(51201)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51202)
		}
	}
	__antithesis_instrumentation__.Notify(51159)

	_ = c.WipeE(ctx, t.L(), corruptNodes)
}

func registerSSTableCorruption(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51203)
	r.Add(registry.TestSpec{
		Name:    "sstable-corruption/table",
		Owner:   registry.OwnerStorage,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51204)
			runSSTableCorruption(ctx, t, c)
		},
	})
}
