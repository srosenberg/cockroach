//go:build acceptance
// +build acceptance

package acceptance

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"os"
	"os/signal"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/acceptance/cluster"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/security/securitytest"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/testcluster"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
)

func MainTest(m *testing.M) {
	__antithesis_instrumentation__.Notify(1152)
	security.SetAssetLoader(securitytest.EmbeddedAssets)
	serverutils.InitTestServerFactory(server.TestServerFactory)
	serverutils.InitTestClusterFactory(testcluster.TestClusterFactory)
	os.Exit(RunTests(m))
}

func RunTests(m *testing.M) int {
	__antithesis_instrumentation__.Notify(1153)
	randutil.SeedForTests()

	ctx := context.Background()
	defer cluster.GenerateCerts(ctx)()

	go func() {
		__antithesis_instrumentation__.Notify(1155)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt)
		<-sig
		stopper.Stop(ctx)
	}()
	__antithesis_instrumentation__.Notify(1154)
	return m.Run()
}
