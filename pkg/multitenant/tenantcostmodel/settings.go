package tenantcostmodel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/settings"
)

var (
	readRequestCost = settings.RegisterFloatSetting(
		settings.TenantReadOnly,
		"tenant_cost_model.kv_read_request_cost",
		"base cost of a read request in Request Units",
		0.6993,
		settings.PositiveFloat,
	)

	readCostPerMB = settings.RegisterFloatSetting(
		settings.TenantReadOnly,
		"tenant_cost_model.kv_read_cost_per_megabyte",
		"cost of a read in Request Units per MB",
		107.6563,
		settings.PositiveFloat,
	)

	writeRequestCost = settings.RegisterFloatSetting(
		settings.TenantReadOnly,
		"tenant_cost_model.kv_write_request_cost",
		"base cost of a write request in Request Units",
		5.7733,
		settings.PositiveFloat,
	)

	writeCostPerMB = settings.RegisterFloatSetting(
		settings.TenantReadOnly,
		"tenant_cost_model.kv_write_cost_per_megabyte",
		"cost of a write in Request Units per MB",
		2026.3021,
		settings.PositiveFloat,
	)

	podCPUSecondCost = settings.RegisterFloatSetting(
		settings.TenantReadOnly,
		"tenant_cost_model.pod_cpu_second_cost",
		"cost of a CPU-second on the tenant POD in Request Units",
		1000.0,
		settings.PositiveFloat,
	)

	pgwireEgressCostPerMB = settings.RegisterFloatSetting(
		settings.TenantReadOnly,
		"tenant_cost_model.pgwire_egress_cost_per_megabyte",
		"cost of client <-> SQL ingress/egress per MB",
		878.9063,
		settings.PositiveFloat,
	)

	configSettings = [...]settings.NonMaskedSetting{
		readRequestCost,
		readCostPerMB,
		writeRequestCost,
		writeCostPerMB,
		podCPUSecondCost,
		pgwireEgressCostPerMB,
	}
)

const perMBToPerByte = float64(1) / (1024 * 1024)

func ConfigFromSettings(sv *settings.Values) Config {
	__antithesis_instrumentation__.Notify(128787)
	return Config{
		KVReadRequest:    RU(readRequestCost.Get(sv)),
		KVReadByte:       RU(readCostPerMB.Get(sv) * perMBToPerByte),
		KVWriteRequest:   RU(writeRequestCost.Get(sv)),
		KVWriteByte:      RU(writeCostPerMB.Get(sv) * perMBToPerByte),
		PodCPUSecond:     RU(podCPUSecondCost.Get(sv)),
		PGWireEgressByte: RU(pgwireEgressCostPerMB.Get(sv) * perMBToPerByte),
	}
}

func DefaultConfig() Config {
	__antithesis_instrumentation__.Notify(128788)
	return Config{
		KVReadRequest:    RU(readRequestCost.Default()),
		KVReadByte:       RU(readCostPerMB.Default() * perMBToPerByte),
		KVWriteRequest:   RU(writeRequestCost.Default()),
		KVWriteByte:      RU(writeCostPerMB.Default() * perMBToPerByte),
		PodCPUSecond:     RU(podCPUSecondCost.Default()),
		PGWireEgressByte: RU(pgwireEgressCostPerMB.Default() * perMBToPerByte),
	}
}

func SetOnChange(sv *settings.Values, fn func(context.Context)) {
	__antithesis_instrumentation__.Notify(128789)
	for _, s := range configSettings {
		__antithesis_instrumentation__.Notify(128790)
		s.SetOnChange(sv, fn)
	}
}

var _ = SetOnChange
var _ = ConfigFromSettings
