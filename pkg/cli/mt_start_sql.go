package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/sdnotify"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/spf13/cobra"
)

var mtStartSQLCmd = &cobra.Command{
	Use:   "start-sql",
	Short: "start a standalone SQL server",
	Long: `
Start a standalone SQL server.

This functionality is **experimental** and for internal use only.

The following certificates are required:

- ca.crt, node.{crt,key}: CA cert and key pair for serving the SQL endpoint.
  Note that under no circumstances should the node certs be shared with those of
  the same name used at the KV layer, as this would pose a severe security risk.
- ca-client-tenant.crt, client-tenant.X.{crt,key}: CA cert and key pair for
  authentication and authorization with the KV layer (as tenant X).
- ca-server-tenant.crt: to authenticate KV layer.

                 ca.crt        ca-client-tenant.crt        ca-server-tenant.crt
user ---------------> sql server ----------------------------> kv
 client.Y.crt    node.crt      client-tenant.X.crt         server-tenant.crt
 client.Y.key    node.key      client-tenant.X.key         server-tenant.key

Note that CA certificates need to be present on the "other" end of the arrow as
well unless it can be verified using a trusted root certificate store. That is,

- ca.crt needs to be passed in the Postgres connection string (sslrootcert) if
  sslmode=verify-ca.
- ca-server-tenant.crt needs to be present on the SQL server.
- ca-client-tenant.crt needs to be present on the KV server.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runStartSQL),
}

func runStartSQL(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(33465)
	tBegin := timeutil.Now()

	if ok, err := maybeRerunBackground(); ok {
		__antithesis_instrumentation__.Notify(33475)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33476)
	}
	__antithesis_instrumentation__.Notify(33466)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, drainSignals...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ambientCtx := serverCfg.AmbientCtx
	ctx = ambientCtx.AnnotateCtx(ctx)

	const clusterName = ""

	stopper, err := setupAndInitializeLoggingAndProfiling(ctx, cmd, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(33477)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33478)
	}
	__antithesis_instrumentation__.Notify(33467)
	defer stopper.Stop(ctx)
	stopper.SetTracer(serverCfg.BaseConfig.AmbientCtx.Tracer)

	st := serverCfg.BaseConfig.Settings

	if err := clusterversion.Initialize(
		ctx, st.Version.BinaryMinSupportedVersion(), &st.SV,
	); err != nil {
		__antithesis_instrumentation__.Notify(33479)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33480)
	}
	__antithesis_instrumentation__.Notify(33468)

	if serverCfg.SQLConfig.TempStorageConfig, err = initTempStorageConfig(
		ctx, serverCfg.Settings, stopper, serverCfg.Stores,
	); err != nil {
		__antithesis_instrumentation__.Notify(33481)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33482)
	}
	__antithesis_instrumentation__.Notify(33469)

	initGEOS(ctx)

	sqlServer, err := server.StartTenant(
		ctx,
		stopper,
		clusterName,
		serverCfg.BaseConfig,
		serverCfg.SQLConfig,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(33483)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33484)
	}
	__antithesis_instrumentation__.Notify(33470)

	if startCtx.pidFile != "" {
		__antithesis_instrumentation__.Notify(33485)
		log.Ops.Infof(ctx, "PID file: %s", startCtx.pidFile)
		if err := ioutil.WriteFile(startCtx.pidFile, []byte(fmt.Sprintf("%d\n", os.Getpid())), 0644); err != nil {
			__antithesis_instrumentation__.Notify(33486)
			log.Ops.Errorf(ctx, "failed writing the PID: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(33487)
		}
	} else {
		__antithesis_instrumentation__.Notify(33488)
	}
	__antithesis_instrumentation__.Notify(33471)

	log.Flush()

	if err := sdnotify.Ready(); err != nil {
		__antithesis_instrumentation__.Notify(33489)
		log.Ops.Errorf(ctx, "failed to signal readiness using systemd protocol: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(33490)
	}
	__antithesis_instrumentation__.Notify(33472)

	if !cluster.TelemetryOptOut() {
		__antithesis_instrumentation__.Notify(33491)
		sqlServer.StartDiagnostics(ctx)
	} else {
		__antithesis_instrumentation__.Notify(33492)
	}
	__antithesis_instrumentation__.Notify(33473)

	tenantClusterID := sqlServer.LogicalClusterID()

	if err := reportServerInfo(ctx, tBegin, &serverCfg, st, false, false, tenantClusterID); err != nil {
		__antithesis_instrumentation__.Notify(33493)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33494)
	}
	__antithesis_instrumentation__.Notify(33474)

	errChan := make(chan error, 1)
	var serverStatusMu serverStatus
	serverStatusMu.started = true

	return waitForShutdown(
		func() serverShutdownInterface { __antithesis_instrumentation__.Notify(33495); return sqlServer },
		stopper,
		errChan, signalCh,
		&serverStatusMu)
}
