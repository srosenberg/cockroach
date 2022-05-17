package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

func registerSyncTest(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51241)
	const nemesisScript = `#!/usr/bin/env bash

if [[ $1 == "on" ]]; then
  charybdefs-nemesis --probability
else
  charybdefs-nemesis --clear
fi
`

	r.Add(registry.TestSpec{
		Skip:  "#48603: broken on Pebble",
		Name:  "synctest",
		Owner: registry.OwnerStorage,

		Cluster: r.MakeClusterSpec(1, spec.ReuseNone()),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51242)
			n := c.Node(1)
			tmpDir, err := ioutil.TempDir("", "synctest")
			if err != nil {
				__antithesis_instrumentation__.Notify(51247)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51248)
			}
			__antithesis_instrumentation__.Notify(51243)
			defer func() {
				__antithesis_instrumentation__.Notify(51249)
				_ = os.RemoveAll(tmpDir)
			}()
			__antithesis_instrumentation__.Notify(51244)
			nemesis := filepath.Join(tmpDir, "nemesis")

			if err := ioutil.WriteFile(nemesis, []byte(nemesisScript), 0755); err != nil {
				__antithesis_instrumentation__.Notify(51250)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51251)
			}
			__antithesis_instrumentation__.Notify(51245)

			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Put(ctx, nemesis, "./nemesis")
			c.Run(ctx, n, "chmod +x nemesis")
			c.Run(ctx, n, "sudo umount {store-dir}/faulty || true")
			c.Run(ctx, n, "mkdir -p {store-dir}/{real,faulty} || true")
			t.Status("setting up charybdefs")

			if err := c.Install(ctx, t.L(), n, "charybdefs"); err != nil {
				__antithesis_instrumentation__.Notify(51252)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51253)
			}
			__antithesis_instrumentation__.Notify(51246)
			c.Run(ctx, n, "sudo charybdefs {store-dir}/faulty -oallow_other,modules=subdir,subdir={store-dir}/real && chmod 777 {store-dir}/{real,faulty}")

			t.Status("running synctest")
			c.Run(ctx, n, "./cockroach debug synctest {store-dir}/faulty ./nemesis")
		},
	})
}
