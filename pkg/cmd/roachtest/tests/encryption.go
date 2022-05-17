package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

func registerEncryption(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47558)

	runEncryption := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(47560)
		nodes := c.Spec().NodeCount
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.EncryptedStores = true
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, nodes))

		adminAddrs, err := c.InternalAdminUIAddr(ctx, t.L(), c.Range(1, nodes))
		if err != nil {
			__antithesis_instrumentation__.Notify(47566)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47567)
		}
		__antithesis_instrumentation__.Notify(47561)
		for _, addr := range adminAddrs {
			__antithesis_instrumentation__.Notify(47568)
			if err := c.RunE(ctx, c.Node(nodes), fmt.Sprintf(`curl http://%s/_status/stores/local | (! grep '"encryptionStatus": null')`, addr)); err != nil {
				__antithesis_instrumentation__.Notify(47569)
				t.Fatalf("encryption status from /_status/stores/local endpoint is null")
			} else {
				__antithesis_instrumentation__.Notify(47570)
			}
		}
		__antithesis_instrumentation__.Notify(47562)

		for i := 1; i <= nodes; i++ {
			__antithesis_instrumentation__.Notify(47571)
			if err := c.StopCockroachGracefullyOnNode(ctx, t.L(), i); err != nil {
				__antithesis_instrumentation__.Notify(47572)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47573)
			}
		}
		__antithesis_instrumentation__.Notify(47563)

		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, nodes))

		testCLIGenKey := func(size int) error {
			__antithesis_instrumentation__.Notify(47574)

			if err := c.RunE(ctx, c.Node(nodes), fmt.Sprintf("./cockroach gen encryption-key -s=%[1]d aes-%[1]d.key", size)); err != nil {
				__antithesis_instrumentation__.Notify(47577)
				return errors.Wrapf(err, "failed to generate AES key with size %d through CLI", size)
			} else {
				__antithesis_instrumentation__.Notify(47578)
			}
			__antithesis_instrumentation__.Notify(47575)

			if err := c.RunE(ctx, c.Node(nodes), fmt.Sprintf(`size=$(wc -c <"aes-%d.key"); test $size -eq %d && exit 0 || exit 1`, size, 32+size/8)); err != nil {
				__antithesis_instrumentation__.Notify(47579)
				return errors.Errorf("expected aes-%d.key has size %d bytes, but got different size", size, 32+size/8)
			} else {
				__antithesis_instrumentation__.Notify(47580)
			}
			__antithesis_instrumentation__.Notify(47576)

			return nil
		}
		__antithesis_instrumentation__.Notify(47564)

		for _, size := range []int{128, 192, 256} {
			__antithesis_instrumentation__.Notify(47581)
			if err := testCLIGenKey(size); err != nil {
				__antithesis_instrumentation__.Notify(47582)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47583)
			}
		}
		__antithesis_instrumentation__.Notify(47565)

		for _, size := range []int{20, 88, 90} {
			__antithesis_instrumentation__.Notify(47584)

			if err := testCLIGenKey(size); err == nil {
				__antithesis_instrumentation__.Notify(47585)
				t.Fatalf("expected error from CLI gen encryption-key, but got nil")
			} else {
				__antithesis_instrumentation__.Notify(47586)
			}
		}
	}
	__antithesis_instrumentation__.Notify(47559)

	for _, n := range []int{1} {
		__antithesis_instrumentation__.Notify(47587)
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("encryption/nodes=%d", n),
			Owner:   registry.OwnerStorage,
			Cluster: r.MakeClusterSpec(n),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(47588)
				runEncryption(ctx, t, c)
			},
		})
	}
}
