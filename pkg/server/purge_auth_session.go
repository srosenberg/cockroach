package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	math_rand "math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

var (
	webSessionPurgeTTL = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.web_session.purge.ttl",
		"if nonzero, entries in system.web_sessions older than this duration are periodically purged",
		time.Hour,
	).WithPublic()

	webSessionAutoLogoutTimeout = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.web_session.auto_logout.timeout",
		"the duration that web sessions will survive before being periodically purged, since they were last used",
		7*24*time.Hour,
		settings.NonNegativeDuration,
	).WithPublic()

	webSessionPurgePeriod = settings.RegisterDurationSetting(
		settings.TenantWritable,
		"server.web_session.purge.period",
		"the time until old sessions are deleted",
		time.Hour,
		settings.NonNegativeDuration,
	).WithPublic()

	webSessionPurgeLimit = settings.RegisterIntSetting(
		settings.TenantWritable,
		"server.web_session.purge.max_deletions_per_cycle",
		"the maximum number of old sessions to delete for each purge",
		10,
	).WithPublic()
)

func startPurgeOldSessions(ctx context.Context, s *authenticationServer) error {
	__antithesis_instrumentation__.Notify(195413)
	return s.sqlServer.stopper.RunAsyncTask(ctx, "purge-old-sessions", func(context.Context) {
		__antithesis_instrumentation__.Notify(195414)
		settingsValues := &s.sqlServer.execCfg.Settings.SV
		period := webSessionPurgePeriod.Get(settingsValues)

		timer := timeutil.NewTimer()
		defer timer.Stop()
		timer.Reset(jitteredInterval(period))

		for ; ; timer.Reset(webSessionPurgePeriod.Get(settingsValues)) {
			__antithesis_instrumentation__.Notify(195415)
			select {
			case <-timer.C:
				__antithesis_instrumentation__.Notify(195416)
				timer.Read = true
				s.purgeOldSessions(ctx)
			case <-s.sqlServer.stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(195417)
				return
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(195418)
				return
			}
		}
	},
	)
}

func (s *authenticationServer) purgeOldSessions(ctx context.Context) {
	__antithesis_instrumentation__.Notify(195419)
	var (
		deleteOldExpiredSessionsStmt = `
DELETE FROM system.web_sessions
WHERE "expiresAt" < $1
ORDER BY random()
LIMIT $2
RETURNING 1
`
		deleteOldRevokedSessionsStmt = `
DELETE FROM system.web_sessions
WHERE "revokedAt" < $1
ORDER BY random()
LIMIT $2
RETURNING 1
`
		deleteSessionsAutoLogoutStmt = `
DELETE FROM system.web_sessions
WHERE "lastUsedAt" < $1
ORDER BY random()
LIMIT $2
RETURNING 1
`
		settingsValues   = &s.sqlServer.execCfg.Settings.SV
		internalExecutor = s.sqlServer.internalExecutor
		currTime         = s.sqlServer.execCfg.Clock.PhysicalTime()

		purgeTTL          = webSessionPurgeTTL.Get(settingsValues)
		autoLogoutTimeout = webSessionAutoLogoutTimeout.Get(settingsValues)
		limit             = webSessionPurgeLimit.Get(settingsValues)

		purgeTime      = currTime.Add(purgeTTL * time.Duration(-1))
		autoLogoutTime = currTime.Add(autoLogoutTimeout * time.Duration(-1))
	)

	if _, err := internalExecutor.ExecEx(
		ctx,
		"delete-old-expired-sessions",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		deleteOldExpiredSessionsStmt,
		purgeTime,
		limit,
	); err != nil {
		__antithesis_instrumentation__.Notify(195422)
		log.Errorf(ctx, "error while deleting old expired web sessions: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(195423)
	}
	__antithesis_instrumentation__.Notify(195420)

	if _, err := internalExecutor.ExecEx(
		ctx,
		"delete-old-revoked-sessions",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		deleteOldRevokedSessionsStmt,
		purgeTime,
		limit,
	); err != nil {
		__antithesis_instrumentation__.Notify(195424)
		log.Errorf(ctx, "error while deleting old revoked web sessions: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(195425)
	}
	__antithesis_instrumentation__.Notify(195421)

	if _, err := internalExecutor.ExecEx(
		ctx,
		"delete-sessions-timeout",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		deleteSessionsAutoLogoutStmt,
		autoLogoutTime,
		limit,
	); err != nil {
		__antithesis_instrumentation__.Notify(195426)
		log.Errorf(ctx, "error while deleting web sessions older than auto-logout timeout: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(195427)
	}
}

func jitteredInterval(interval time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(195428)
	return time.Duration(float64(interval) * (0.75 + 0.5*math_rand.Float64()))
}
