package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/http"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
)

func RunBuildInfo(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(45896)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	var details serverpb.DetailsResponse
	adminUIAddrs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(1))
	if err != nil {
		__antithesis_instrumentation__.Notify(45899)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(45900)
	}
	__antithesis_instrumentation__.Notify(45897)
	url := `http://` + adminUIAddrs[0] + `/_status/details/local`
	err = httputil.GetJSON(http.Client{}, url, &details)
	if err != nil {
		__antithesis_instrumentation__.Notify(45901)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(45902)
	}
	__antithesis_instrumentation__.Notify(45898)

	bi := details.BuildInfo
	testData := map[string]string{
		"go_version": bi.GoVersion,
		"tag":        bi.Tag,
		"time":       bi.Time,
		"revision":   bi.Revision,
	}
	for key, val := range testData {
		__antithesis_instrumentation__.Notify(45903)
		if val == "" {
			__antithesis_instrumentation__.Notify(45904)
			t.Fatalf("build info not set for \"%s\"", key)
		} else {
			__antithesis_instrumentation__.Notify(45905)
		}
	}
}

func RunBuildAnalyze(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(45906)

	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(45909)

		t.Skip("local execution not supported")
	} else {
		__antithesis_instrumentation__.Notify(45910)
	}
	__antithesis_instrumentation__.Notify(45907)

	c.Put(ctx, t.Cockroach(), "./cockroach")

	c.Run(ctx, c.Node(1), "sudo apt-get update")
	c.Run(ctx, c.Node(1), "sudo apt-get -qqy install pax-utils")

	result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(1), "scanelf -qe cockroach")
	if err != nil {
		__antithesis_instrumentation__.Notify(45911)
		t.Fatalf("scanelf failed: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(45912)
	}
	__antithesis_instrumentation__.Notify(45908)
	output := strings.TrimSpace(result.Stdout)
	if len(output) > 0 {
		__antithesis_instrumentation__.Notify(45913)
		t.Fatalf("scanelf returned non-empty output (executable stack): %s", output)
	} else {
		__antithesis_instrumentation__.Notify(45914)
	}
}
