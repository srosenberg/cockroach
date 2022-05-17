package changefeedbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/errors"
)

var TableDescriptorPollInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"changefeed.experimental_poll_interval",
	"polling interval for the table descriptors",
	1*time.Second,
	settings.NonNegativeDuration,
)

var DefaultMinCheckpointFrequency = 30 * time.Second

func TestingSetDefaultMinCheckpointFrequency(f time.Duration) func() {
	__antithesis_instrumentation__.Notify(16623)
	old := DefaultMinCheckpointFrequency
	DefaultMinCheckpointFrequency = f
	return func() { __antithesis_instrumentation__.Notify(16624); DefaultMinCheckpointFrequency = old }
}

var PerChangefeedMemLimit = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"changefeed.memory.per_changefeed_limit",
	"controls amount of data that can be buffered per changefeed",
	1<<30,
)

var SlowSpanLogThreshold = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"changefeed.slow_span_log_threshold",
	"a changefeed will log spans with resolved timestamps this far behind the current wall-clock time; if 0, a default value is calculated based on other cluster settings",
	0,
	settings.NonNegativeDuration,
)

var IdleTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"changefeed.idle_timeout",
	"a changefeed will mark itself idle if no changes have been emitted for greater than this duration; if 0, the changefeed will never be marked idle",
	10*time.Minute,
	settings.NonNegativeDuration,
)

var FrontierCheckpointFrequency = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"changefeed.frontier_checkpoint_frequency",
	"controls the frequency with which span level checkpoints will be written; if 0, disabled.",
	10*time.Minute,
	settings.NonNegativeDuration,
)

var FrontierCheckpointMaxBytes = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"changefeed.frontier_checkpoint_max_bytes",
	"controls the maximum size of the checkpoint as a total size of key bytes",
	1<<20,
)

var ScanRequestLimit = settings.RegisterIntSetting(
	settings.TenantWritable,
	"changefeed.backfill.concurrent_scan_requests",
	"number of concurrent scan requests per node issued during a backfill",
	0,
)

var ScanRequestSize = settings.RegisterIntSetting(
	settings.TenantWritable,
	"changefeed.backfill.scan_request_size",
	"the maximum number of bytes returned by each scan request",
	16<<20,
)

type SinkThrottleConfig struct {
	MessageRate float64 `json:",omitempty"`

	MessageBurst float64 `json:",omitempty"`

	ByteRate float64 `json:",omitempty"`

	ByteBurst float64 `json:",omitempty"`

	FlushRate float64 `json:",omitempty"`

	FlushBurst float64 `json:",omitempty"`
}

var NodeSinkThrottleConfig = func() *settings.StringSetting {
	__antithesis_instrumentation__.Notify(16625)
	s := settings.RegisterValidatedStringSetting(
		settings.TenantWritable,
		"changefeed.node_throttle_config",
		"specifies node level throttling configuration for all changefeeeds",
		"",
		validateSinkThrottleConfig,
	)
	s.SetVisibility(settings.Public)
	s.SetReportable(true)
	return s
}()

func validateSinkThrottleConfig(values *settings.Values, configStr string) error {
	__antithesis_instrumentation__.Notify(16626)
	if configStr == "" {
		__antithesis_instrumentation__.Notify(16628)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(16629)
	}
	__antithesis_instrumentation__.Notify(16627)
	var config = &SinkThrottleConfig{}
	return json.Unmarshal([]byte(configStr), config)
}

var MinHighWaterMarkCheckpointAdvance = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"changefeed.min_highwater_advance",
	"minimum amount of time the changefeed high water mark must advance "+
		"for it to be eligible for checkpointing; Default of 0 will checkpoint every time frontier "+
		"advances, as long as the rate of checkpointing keeps up with the rate of frontier changes",
	0,
	settings.NonNegativeDuration,
)

var EventMemoryMultiplier = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"changefeed.event_memory_multiplier",
	"the amount of memory required to process an event is multiplied by this factor",
	3,
	func(v float64) error {
		__antithesis_instrumentation__.Notify(16630)
		if v < 1 {
			__antithesis_instrumentation__.Notify(16632)
			return errors.New("changefeed.event_memory_multiplier must be at least 1")
		} else {
			__antithesis_instrumentation__.Notify(16633)
		}
		__antithesis_instrumentation__.Notify(16631)
		return nil
	},
)

var ProtectTimestampInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"changefeed.protect_timestamp_interval",
	"controls how often the changefeed forwards its protected timestamp to the resolved timestamp",
	10*time.Minute,
	settings.PositiveDuration,
)

var ActiveProtectedTimestampsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"changefeed.active_protected_timestamps.enabled",
	"if set, rather than only protecting changefeed targets from garbage collection during backfills, data will always be protected up to the changefeed's frontier",
	true,
)
