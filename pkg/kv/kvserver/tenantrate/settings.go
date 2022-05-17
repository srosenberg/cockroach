package tenantrate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"runtime"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/errors"
)

type Config struct {
	Rate float64

	Burst float64

	ReadRequestUnits float64

	ReadUnitsPerByte float64

	WriteRequestUnits float64

	WriteUnitsPerByte float64
}

var (
	kvcuRateLimit = settings.RegisterFloatSetting(
		settings.TenantWritable,
		"kv.tenant_rate_limiter.rate_limit",
		"per-tenant rate limit in KV Compute Units per second if positive, "+
			"or KV Compute Units per second per CPU if negative",
		-200,
		func(v float64) error {
			__antithesis_instrumentation__.Notify(126703)
			if v == 0 {
				__antithesis_instrumentation__.Notify(126705)
				return errors.New("cannot set to zero value")
			} else {
				__antithesis_instrumentation__.Notify(126706)
			}
			__antithesis_instrumentation__.Notify(126704)
			return nil
		},
	)

	kvcuBurstLimitSeconds = settings.RegisterFloatSetting(
		settings.TenantWritable,
		"kv.tenant_rate_limiter.burst_limit_seconds",
		"per-tenant burst limit as a multiplier of the rate",
		10,
		settings.PositiveFloat,
	)

	readRequestCost = settings.RegisterFloatSetting(
		settings.TenantWritable,
		"kv.tenant_rate_limiter.read_request_cost",
		"base cost of a read request in KV Compute Units",
		0.7,
		settings.PositiveFloat,
	)

	readCostPerMB = settings.RegisterFloatSetting(
		settings.TenantWritable,
		"kv.tenant_rate_limiter.read_cost_per_megabyte",
		"cost of a read in KV Compute Units per MB",
		10.0,
		settings.PositiveFloat,
	)

	writeRequestCost = settings.RegisterFloatSetting(
		settings.TenantWritable,
		"kv.tenant_rate_limiter.write_request_cost",
		"base cost of a write request in KV Compute Units",
		1.0,
		settings.PositiveFloat,
	)

	writeCostPerMB = settings.RegisterFloatSetting(
		settings.TenantWritable,
		"kv.tenant_rate_limiter.write_cost_per_megabyte",
		"cost of a write in KV Compute Units per MB",
		400.0,
		settings.PositiveFloat,
	)

	configSettings = [...]settings.NonMaskedSetting{
		kvcuRateLimit,
		kvcuBurstLimitSeconds,
		readRequestCost,
		readCostPerMB,
		writeRequestCost,
		writeCostPerMB,
	}
)

func absoluteRateFromConfigValue(value float64) float64 {
	__antithesis_instrumentation__.Notify(126707)
	if value < 0 {
		__antithesis_instrumentation__.Notify(126709)

		return -value * float64(runtime.GOMAXPROCS(0))
	} else {
		__antithesis_instrumentation__.Notify(126710)
	}
	__antithesis_instrumentation__.Notify(126708)
	return value
}

func ConfigFromSettings(sv *settings.Values) Config {
	__antithesis_instrumentation__.Notify(126711)
	rate := absoluteRateFromConfigValue(kvcuRateLimit.Get(sv))
	return Config{
		Rate:              rate,
		Burst:             rate * kvcuBurstLimitSeconds.Get(sv),
		ReadRequestUnits:  readRequestCost.Get(sv),
		ReadUnitsPerByte:  readCostPerMB.Get(sv) / (1024 * 1024),
		WriteRequestUnits: writeRequestCost.Get(sv),
		WriteUnitsPerByte: writeCostPerMB.Get(sv) / (1024 * 1024),
	}
}

func DefaultConfig() Config {
	__antithesis_instrumentation__.Notify(126712)
	rate := absoluteRateFromConfigValue(kvcuRateLimit.Default())
	return Config{
		Rate:              rate,
		Burst:             rate * kvcuBurstLimitSeconds.Default(),
		ReadRequestUnits:  readRequestCost.Default(),
		ReadUnitsPerByte:  readCostPerMB.Default() / (1024 * 1024),
		WriteRequestUnits: writeRequestCost.Default(),
		WriteUnitsPerByte: writeCostPerMB.Default() / (1024 * 1024),
	}
}
