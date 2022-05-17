package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

const goPath = `/mnt/data1/go`

func installGolang(
	ctx context.Context, t test.Test, c cluster.Cluster, node option.NodeListOption,
) {
	__antithesis_instrumentation__.Notify(47997)
	if err := repeatRunE(
		ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
	); err != nil {
		__antithesis_instrumentation__.Notify(48003)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48004)
	}
	__antithesis_instrumentation__.Notify(47998)

	if err := repeatRunE(
		ctx,
		t,
		c,
		node,
		"install dependencies (go uses C bindings)",
		`sudo apt-get -qq install build-essential`,
	); err != nil {
		__antithesis_instrumentation__.Notify(48005)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48006)
	}
	__antithesis_instrumentation__.Notify(47999)

	if err := repeatRunE(
		ctx, t, c, node, "download go", `curl -fsSL https://dl.google.com/go/go1.17.6.linux-amd64.tar.gz > /tmp/go.tgz`,
	); err != nil {
		__antithesis_instrumentation__.Notify(48007)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48008)
	}
	__antithesis_instrumentation__.Notify(48000)
	if err := repeatRunE(
		ctx, t, c, node, "verify tarball", `sha256sum -c - <<EOF
231654bbf2dab3d86c1619ce799e77b03d96f9b50770297c8f4dff8836fc8ca2 /tmp/go.tgz
EOF`,
	); err != nil {
		__antithesis_instrumentation__.Notify(48009)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48010)
	}
	__antithesis_instrumentation__.Notify(48001)
	if err := repeatRunE(
		ctx, t, c, node, "extract go", `sudo tar -C /usr/local -zxf /tmp/go.tgz && rm /tmp/go.tgz`,
	); err != nil {
		__antithesis_instrumentation__.Notify(48011)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48012)
	}
	__antithesis_instrumentation__.Notify(48002)
	if err := repeatRunE(
		ctx, t, c, node, "force symlink go", "sudo ln -sf /usr/local/go/bin/go /usr/bin",
	); err != nil {
		__antithesis_instrumentation__.Notify(48013)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48014)
	}
}
