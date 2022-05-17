package utilccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl/licenseccl"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

var enterpriseLicense = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(27151)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		"enterprise.license",
		"the encoded cluster license",
		"",
		func(sv *settings.Values, s string) error {
			__antithesis_instrumentation__.Notify(27153)
			_, err := decode(s)
			return err
		},
	)
	__antithesis_instrumentation__.Notify(27152)

	s.SetReportable(false)
	s.SetVisibility(settings.Public)
	return s
}()

var enterpriseStatus int32 = deferToLicense

const (
	deferToLicense    = 0
	enterpriseEnabled = 1
)

var errEnterpriseRequired = pgerror.New(pgcode.CCLValidLicenseRequired,
	"a valid enterprise license is required")

type licenseCacheKey string

func TestingEnableEnterprise() func() {
	__antithesis_instrumentation__.Notify(27154)
	before := atomic.LoadInt32(&enterpriseStatus)
	atomic.StoreInt32(&enterpriseStatus, enterpriseEnabled)
	return func() {
		__antithesis_instrumentation__.Notify(27155)
		atomic.StoreInt32(&enterpriseStatus, before)
	}
}

func TestingDisableEnterprise() func() {
	__antithesis_instrumentation__.Notify(27156)
	before := atomic.LoadInt32(&enterpriseStatus)
	atomic.StoreInt32(&enterpriseStatus, deferToLicense)
	return func() {
		__antithesis_instrumentation__.Notify(27157)
		atomic.StoreInt32(&enterpriseStatus, before)
	}
}

func ApplyTenantLicense() error {
	__antithesis_instrumentation__.Notify(27158)
	license, ok := envutil.EnvString("COCKROACH_TENANT_LICENSE", 0)
	if !ok {
		__antithesis_instrumentation__.Notify(27161)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(27162)
	}
	__antithesis_instrumentation__.Notify(27159)
	if _, err := decode(license); err != nil {
		__antithesis_instrumentation__.Notify(27163)
		return errors.Wrap(err, "COCKROACH_TENANT_LICENSE encoding is invalid")
	} else {
		__antithesis_instrumentation__.Notify(27164)
	}
	__antithesis_instrumentation__.Notify(27160)
	atomic.StoreInt32(&enterpriseStatus, enterpriseEnabled)
	return nil
}

func CheckEnterpriseEnabled(st *cluster.Settings, cluster uuid.UUID, org, feature string) error {
	__antithesis_instrumentation__.Notify(27165)
	return checkEnterpriseEnabledAt(st, timeutil.Now(), cluster, org, feature, true)
}

func IsEnterpriseEnabled(st *cluster.Settings, cluster uuid.UUID, org, feature string) bool {
	__antithesis_instrumentation__.Notify(27166)
	return checkEnterpriseEnabledAt(
		st, timeutil.Now(), cluster, org, feature, false) == nil
}

func init() {
	base.CheckEnterpriseEnabled = CheckEnterpriseEnabled
	base.LicenseType = getLicenseType
	base.UpdateMetricOnLicenseChange = UpdateMetricOnLicenseChange
	server.ApplyTenantLicense = ApplyTenantLicense
}

var licenseMetricUpdateFrequency = 1 * time.Minute

func UpdateMetricOnLicenseChange(
	ctx context.Context,
	st *cluster.Settings,
	metric *metric.Gauge,
	ts timeutil.TimeSource,
	stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(27167)
	enterpriseLicense.SetOnChange(&st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(27169)
		updateMetricWithLicenseTTL(ctx, st, metric, ts)
	})
	__antithesis_instrumentation__.Notify(27168)
	return stopper.RunAsyncTask(ctx, "write-license-expiry-metric", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(27170)
		ticker := time.NewTicker(licenseMetricUpdateFrequency)
		defer ticker.Stop()
		for {
			__antithesis_instrumentation__.Notify(27171)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(27172)
				updateMetricWithLicenseTTL(ctx, st, metric, ts)
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(27173)
				return
			}
		}
	})
}

func updateMetricWithLicenseTTL(
	ctx context.Context, st *cluster.Settings, metric *metric.Gauge, ts timeutil.TimeSource,
) {
	__antithesis_instrumentation__.Notify(27174)
	license, err := getLicense(st)
	if err != nil {
		__antithesis_instrumentation__.Notify(27177)
		log.Errorf(ctx, "unable to update license expiry metric: %v", err)
		metric.Update(0)
		return
	} else {
		__antithesis_instrumentation__.Notify(27178)
	}
	__antithesis_instrumentation__.Notify(27175)
	if license == nil {
		__antithesis_instrumentation__.Notify(27179)
		metric.Update(0)
		return
	} else {
		__antithesis_instrumentation__.Notify(27180)
	}
	__antithesis_instrumentation__.Notify(27176)
	sec := timeutil.Unix(license.ValidUntilUnixSec, 0).Sub(ts.Now()).Seconds()
	metric.Update(int64(sec))
}

func checkEnterpriseEnabledAt(
	st *cluster.Settings, at time.Time, cluster uuid.UUID, org, feature string, withDetails bool,
) error {
	__antithesis_instrumentation__.Notify(27181)
	if atomic.LoadInt32(&enterpriseStatus) == enterpriseEnabled {
		__antithesis_instrumentation__.Notify(27184)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(27185)
	}
	__antithesis_instrumentation__.Notify(27182)
	license, err := getLicense(st)
	if err != nil {
		__antithesis_instrumentation__.Notify(27186)
		return err
	} else {
		__antithesis_instrumentation__.Notify(27187)
	}
	__antithesis_instrumentation__.Notify(27183)
	return check(license, at, cluster, org, feature, withDetails)
}

func getLicense(st *cluster.Settings) (*licenseccl.License, error) {
	__antithesis_instrumentation__.Notify(27188)
	str := enterpriseLicense.Get(&st.SV)
	if str == "" {
		__antithesis_instrumentation__.Notify(27192)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(27193)
	}
	__antithesis_instrumentation__.Notify(27189)
	cacheKey := licenseCacheKey(str)
	if cachedLicense, ok := st.Cache.Load(cacheKey); ok {
		__antithesis_instrumentation__.Notify(27194)
		return cachedLicense.(*licenseccl.License), nil
	} else {
		__antithesis_instrumentation__.Notify(27195)
	}
	__antithesis_instrumentation__.Notify(27190)
	license, err := decode(str)
	if err != nil {
		__antithesis_instrumentation__.Notify(27196)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(27197)
	}
	__antithesis_instrumentation__.Notify(27191)
	st.Cache.Store(cacheKey, license)
	return license, nil
}

func getLicenseType(st *cluster.Settings) (string, error) {
	__antithesis_instrumentation__.Notify(27198)
	license, err := getLicense(st)
	if err != nil {
		__antithesis_instrumentation__.Notify(27200)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(27201)
		if license == nil {
			__antithesis_instrumentation__.Notify(27202)
			return "None", nil
		} else {
			__antithesis_instrumentation__.Notify(27203)
		}
	}
	__antithesis_instrumentation__.Notify(27199)
	return license.Type.String(), nil
}

func decode(s string) (*licenseccl.License, error) {
	__antithesis_instrumentation__.Notify(27204)
	lic, err := licenseccl.Decode(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(27206)
		return nil, pgerror.WithCandidateCode(err, pgcode.Syntax)
	} else {
		__antithesis_instrumentation__.Notify(27207)
	}
	__antithesis_instrumentation__.Notify(27205)
	return lic, nil
}

func check(
	l *licenseccl.License, at time.Time, cluster uuid.UUID, org, feature string, withDetails bool,
) error {
	__antithesis_instrumentation__.Notify(27208)
	if l == nil {
		__antithesis_instrumentation__.Notify(27215)
		if !withDetails {
			__antithesis_instrumentation__.Notify(27217)
			return errEnterpriseRequired
		} else {
			__antithesis_instrumentation__.Notify(27218)
		}
		__antithesis_instrumentation__.Notify(27216)

		link := "https://cockroachlabs.com/pricing?cluster="
		return pgerror.Newf(pgcode.CCLValidLicenseRequired,
			"use of %s requires an enterprise license. "+
				"see %s%s for details on how to enable enterprise features",
			errors.Safe(feature),
			link,
			cluster.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(27219)
	}
	__antithesis_instrumentation__.Notify(27209)

	if l.ValidUntilUnixSec > 0 && func() bool {
		__antithesis_instrumentation__.Notify(27220)
		return l.Type != licenseccl.License_Enterprise == true
	}() == true {
		__antithesis_instrumentation__.Notify(27221)
		if expiration := timeutil.Unix(l.ValidUntilUnixSec, 0); at.After(expiration) {
			__antithesis_instrumentation__.Notify(27222)
			if !withDetails {
				__antithesis_instrumentation__.Notify(27225)
				return errEnterpriseRequired
			} else {
				__antithesis_instrumentation__.Notify(27226)
			}
			__antithesis_instrumentation__.Notify(27223)
			licensePrefix := redact.SafeString("")
			switch l.Type {
			case licenseccl.License_NonCommercial:
				__antithesis_instrumentation__.Notify(27227)
				licensePrefix = "non-commercial "
			case licenseccl.License_Evaluation:
				__antithesis_instrumentation__.Notify(27228)
				licensePrefix = "evaluation "
			default:
				__antithesis_instrumentation__.Notify(27229)
			}
			__antithesis_instrumentation__.Notify(27224)
			return pgerror.Newf(pgcode.CCLValidLicenseRequired,
				"Use of %s requires an enterprise license. Your %slicense expired on %s. If you're "+
					"interested in getting a new license, please contact subscriptions@cockroachlabs.com "+
					"and we can help you out.",
				errors.Safe(feature),
				licensePrefix,
				expiration.Format("January 2, 2006"),
			)
		} else {
			__antithesis_instrumentation__.Notify(27230)
		}
	} else {
		__antithesis_instrumentation__.Notify(27231)
	}
	__antithesis_instrumentation__.Notify(27210)

	if l.ClusterID == nil {
		__antithesis_instrumentation__.Notify(27232)
		if strings.EqualFold(l.OrganizationName, org) {
			__antithesis_instrumentation__.Notify(27235)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(27236)
		}
		__antithesis_instrumentation__.Notify(27233)
		if !withDetails {
			__antithesis_instrumentation__.Notify(27237)
			return errEnterpriseRequired
		} else {
			__antithesis_instrumentation__.Notify(27238)
		}
		__antithesis_instrumentation__.Notify(27234)
		return pgerror.Newf(pgcode.CCLValidLicenseRequired,
			"license valid only for %q", l.OrganizationName)
	} else {
		__antithesis_instrumentation__.Notify(27239)
	}
	__antithesis_instrumentation__.Notify(27211)

	for _, c := range l.ClusterID {
		__antithesis_instrumentation__.Notify(27240)
		if cluster == c {
			__antithesis_instrumentation__.Notify(27241)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(27242)
		}
	}
	__antithesis_instrumentation__.Notify(27212)

	if !withDetails {
		__antithesis_instrumentation__.Notify(27243)
		return errEnterpriseRequired
	} else {
		__antithesis_instrumentation__.Notify(27244)
	}
	__antithesis_instrumentation__.Notify(27213)
	var matches bytes.Buffer
	for i, c := range l.ClusterID {
		__antithesis_instrumentation__.Notify(27245)
		if i > 0 {
			__antithesis_instrumentation__.Notify(27247)
			matches.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(27248)
		}
		__antithesis_instrumentation__.Notify(27246)
		matches.WriteString(c.String())
	}
	__antithesis_instrumentation__.Notify(27214)
	return pgerror.Newf(pgcode.CCLValidLicenseRequired,
		"license for cluster(s) %s is not valid for cluster %s",
		matches.String(), cluster.String(),
	)
}
