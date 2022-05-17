package diagnostics

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics/diagnosticspb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const (
	updateCheckFrequency = time.Hour * 24

	updateCheckPostStartup    = time.Minute * 5
	updateCheckRetryFrequency = time.Hour
	updateMaxVersionsToReport = 3
)

type versionInfo struct {
	Version string `json:"version"`
	Details string `json:"details"`
}

type UpdateChecker struct {
	StartTime  time.Time
	AmbientCtx *log.AmbientContext
	Config     *base.Config
	Settings   *cluster.Settings

	StorageClusterID func() uuid.UUID

	LogicalClusterID func() uuid.UUID

	NodeID func() roachpb.NodeID

	SQLInstanceID func() base.SQLInstanceID

	TestingKnobs *TestingKnobs
}

func (u *UpdateChecker) PeriodicallyCheckForUpdates(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(193180)
	_ = stopper.RunAsyncTaskEx(ctx, stop.TaskOpts{
		TaskName: "update-checker",
		SpanOpt:  stop.SterileRootSpan,
	}, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(193181)
		defer logcrash.RecoverAndReportNonfatalPanic(ctx, &u.Settings.SV)
		nextUpdateCheck := u.StartTime

		var timer timeutil.Timer
		defer timer.Stop()
		for {
			__antithesis_instrumentation__.Notify(193182)
			now := timeutil.Now()
			runningTime := now.Sub(u.StartTime)

			nextUpdateCheck = u.maybeCheckForUpdates(ctx, now, nextUpdateCheck, runningTime)

			timer.Reset(addJitter(nextUpdateCheck.Sub(timeutil.Now())))
			select {
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(193183)
				return
			case <-timer.C:
				__antithesis_instrumentation__.Notify(193184)
				timer.Read = true
			}
		}
	})
}

func (u *UpdateChecker) CheckForUpdates(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(193185)
	ctx, span := u.AmbientCtx.AnnotateCtxWithSpan(ctx, "version update check")
	defer span.Finish()

	url := u.buildUpdatesURL(ctx)
	if url == nil {
		__antithesis_instrumentation__.Notify(193192)
		return true
	} else {
		__antithesis_instrumentation__.Notify(193193)
	}
	__antithesis_instrumentation__.Notify(193186)

	res, err := httputil.Get(ctx, url.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(193194)

		return false
	} else {
		__antithesis_instrumentation__.Notify(193195)
	}
	__antithesis_instrumentation__.Notify(193187)
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		__antithesis_instrumentation__.Notify(193196)
		b, err := ioutil.ReadAll(res.Body)
		log.Infof(ctx, "failed to check for updates: status: %s, body: %s, error: %v",
			res.Status, b, err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193197)
	}
	__antithesis_instrumentation__.Notify(193188)

	decoder := json.NewDecoder(res.Body)
	r := struct {
		Details []versionInfo `json:"details"`
	}{}

	err = decoder.Decode(&r)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(193198)
		return err != io.EOF == true
	}() == true {
		__antithesis_instrumentation__.Notify(193199)
		log.Warningf(ctx, "error decoding updates info: %v", err)
		return false
	} else {
		__antithesis_instrumentation__.Notify(193200)
	}
	__antithesis_instrumentation__.Notify(193189)

	if len(r.Details) > updateMaxVersionsToReport {
		__antithesis_instrumentation__.Notify(193201)
		r.Details = r.Details[len(r.Details)-updateMaxVersionsToReport:]
	} else {
		__antithesis_instrumentation__.Notify(193202)
	}
	__antithesis_instrumentation__.Notify(193190)
	for _, v := range r.Details {
		__antithesis_instrumentation__.Notify(193203)
		log.Infof(ctx, "a new version is available: %s, details: %s", v.Version, v.Details)
	}
	__antithesis_instrumentation__.Notify(193191)
	return true
}

func (u *UpdateChecker) maybeCheckForUpdates(
	ctx context.Context, now, scheduled time.Time, runningTime time.Duration,
) time.Time {
	__antithesis_instrumentation__.Notify(193204)
	if scheduled.After(now) {
		__antithesis_instrumentation__.Notify(193209)
		return scheduled
	} else {
		__antithesis_instrumentation__.Notify(193210)
	}
	__antithesis_instrumentation__.Notify(193205)

	if !logcrash.DiagnosticsReportingEnabled.Get(&u.Settings.SV) {
		__antithesis_instrumentation__.Notify(193211)
		return now.Add(updateCheckFrequency)
	} else {
		__antithesis_instrumentation__.Notify(193212)
	}
	__antithesis_instrumentation__.Notify(193206)

	if succeeded := u.CheckForUpdates(ctx); !succeeded {
		__antithesis_instrumentation__.Notify(193213)
		return now.Add(updateCheckRetryFrequency)
	} else {
		__antithesis_instrumentation__.Notify(193214)
	}
	__antithesis_instrumentation__.Notify(193207)

	if runningTime < updateCheckPostStartup {
		__antithesis_instrumentation__.Notify(193215)
		return now.Add(time.Hour - runningTime)
	} else {
		__antithesis_instrumentation__.Notify(193216)
	}
	__antithesis_instrumentation__.Notify(193208)

	return now.Add(updateCheckFrequency)
}

func (u *UpdateChecker) buildUpdatesURL(ctx context.Context) *url.URL {
	__antithesis_instrumentation__.Notify(193217)
	clusterInfo := ClusterInfo{
		StorageClusterID: u.StorageClusterID(),
		LogicalClusterID: u.LogicalClusterID(),
		TenantID:         roachpb.SystemTenantID,
		IsInsecure:       u.Config.Insecure,
		IsInternal:       sql.ClusterIsInternal(&u.Settings.SV),
	}

	var env diagnosticspb.Environment
	env.Build = build.GetInfo()
	env.LicenseType = getLicenseType(ctx, u.Settings)
	populateHardwareInfo(ctx, &env)

	sqlInfo := diagnosticspb.SQLInstanceInfo{
		SQLInstanceID: u.SQLInstanceID(),
		Uptime:        int64(timeutil.Since(u.StartTime).Seconds()),
	}

	url := updatesURL
	if u.TestingKnobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(193219)
		return u.TestingKnobs.OverrideUpdatesURL != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(193220)
		url = *u.TestingKnobs.OverrideUpdatesURL
	} else {
		__antithesis_instrumentation__.Notify(193221)
	}
	__antithesis_instrumentation__.Notify(193218)
	return addInfoToURL(url, &clusterInfo, &env, u.NodeID(), &sqlInfo)
}
