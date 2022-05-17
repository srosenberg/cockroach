package featureflag

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
)

const FeatureFlagEnabledDefault = true

func CheckEnabled(
	ctx context.Context, config Config, s *settings.BoolSetting, featureName string,
) error {
	__antithesis_instrumentation__.Notify(58473)
	sv := config.SV()

	if enabled := s.Get(sv); !enabled {
		__antithesis_instrumentation__.Notify(58475)

		telemetry.Inc(sqltelemetry.FeatureDeniedByFeatureFlagCounter)

		if config.GetFeatureFlagMetrics() == nil {
			__antithesis_instrumentation__.Notify(58478)
			log.Warningf(
				ctx,
				"executorConfig.FeatureFlagMetrics is uninitiated; feature %s was attempted but disabled via cluster settings",
				featureName,
			)
			return pgerror.Newf(
				pgcode.OperatorIntervention,
				"feature %s was disabled by the database administrator",
				featureName,
			)
		} else {
			__antithesis_instrumentation__.Notify(58479)
		}
		__antithesis_instrumentation__.Notify(58476)
		config.GetFeatureFlagMetrics().FeatureDenialMetric.Inc(1)

		if log.V(2) {
			__antithesis_instrumentation__.Notify(58480)
			log.Warningf(
				ctx,
				"feature %s was attempted but disabled via cluster settings",
				featureName,
			)
		} else {
			__antithesis_instrumentation__.Notify(58481)
		}
		__antithesis_instrumentation__.Notify(58477)

		return pgerror.Newf(
			pgcode.OperatorIntervention,
			"feature %s was disabled by the database administrator",
			featureName,
		)
	} else {
		__antithesis_instrumentation__.Notify(58482)
	}
	__antithesis_instrumentation__.Notify(58474)
	return nil
}

var metaFeatureDenialMetric = metric.Metadata{
	Name:        "sql.feature_flag_denial",
	Help:        "Counter of the number of statements denied by a feature flag",
	Measurement: "Statements",
	Unit:        metric.Unit_COUNT,
}

type DenialMetrics struct {
	FeatureDenialMetric *metric.Counter
}

func (s *DenialMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(58483) }

var _ metric.Struct = (*DenialMetrics)(nil)

func NewFeatureFlagMetrics() *DenialMetrics {
	__antithesis_instrumentation__.Notify(58484)
	return &DenialMetrics{
		FeatureDenialMetric: metric.NewCounter(metaFeatureDenialMetric),
	}
}

type Config interface {
	SV() *settings.Values
	GetFeatureFlagMetrics() *DenialMetrics
}
