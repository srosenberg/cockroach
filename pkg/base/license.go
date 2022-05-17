package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var errEnterpriseNotEnabled = errors.New("OSS binaries do not include enterprise features")

var CheckEnterpriseEnabled = func(_ *cluster.Settings, _ uuid.UUID, org, feature string) error {
	__antithesis_instrumentation__.Notify(1467)
	return errEnterpriseNotEnabled
}

var licenseTTLMetadata = metric.Metadata{

	Name:        "seconds_until_enterprise_license_expiry",
	Help:        "Seconds until enterprise license expiry (0 if no license present or running without enterprise features)",
	Measurement: "Seconds",
	Unit:        metric.Unit_SECONDS,
}

var LicenseTTL = metric.NewGauge(licenseTTLMetadata)

var UpdateMetricOnLicenseChange = func(
	ctx context.Context,
	st *cluster.Settings,
	metric *metric.Gauge,
	ts timeutil.TimeSource,
	stopper *stop.Stopper,
) error {
	__antithesis_instrumentation__.Notify(1468)
	return nil
}

var LicenseType = func(st *cluster.Settings) (string, error) {
	__antithesis_instrumentation__.Notify(1469)
	return "OSS", nil
}
