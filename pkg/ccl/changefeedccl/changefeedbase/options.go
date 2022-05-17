package changefeedbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql"
)

type EnvelopeType string

type FormatType string

type OnErrorType string

type SchemaChangeEventClass string

type SchemaChangePolicy string

type VirtualColumnVisibility string

type InitialScanType int

const (
	InitialScan InitialScanType = iota
	NoInitialScan
	OnlyInitialScan
)

const (
	OptAvroSchemaPrefix         = `avro_schema_prefix`
	OptConfluentSchemaRegistry  = `confluent_schema_registry`
	OptCursor                   = `cursor`
	OptEndTime                  = `end_time`
	OptEnvelope                 = `envelope`
	OptFormat                   = `format`
	OptFullTableName            = `full_table_name`
	OptKeyInValue               = `key_in_value`
	OptTopicInValue             = `topic_in_value`
	OptResolvedTimestamps       = `resolved`
	OptMinCheckpointFrequency   = `min_checkpoint_frequency`
	OptUpdatedTimestamps        = `updated`
	OptMVCCTimestamps           = `mvcc_timestamp`
	OptDiff                     = `diff`
	OptCompression              = `compression`
	OptSchemaChangeEvents       = `schema_change_events`
	OptSchemaChangePolicy       = `schema_change_policy`
	OptSplitColumnFamilies      = `split_column_families`
	OptProtectDataFromGCOnPause = `protect_data_from_gc_on_pause`
	OptWebhookAuthHeader        = `webhook_auth_header`
	OptWebhookClientTimeout     = `webhook_client_timeout`
	OptOnError                  = `on_error`
	OptMetricsScope             = `metrics_label`
	OptVirtualColumns           = `virtual_columns`

	OptVirtualColumnsOmitted VirtualColumnVisibility = `omitted`
	OptVirtualColumnsNull    VirtualColumnVisibility = `null`

	OptSchemaChangeEventClassColumnChange SchemaChangeEventClass = `column_changes`

	OptSchemaChangeEventClassDefault SchemaChangeEventClass = `default`

	OptSchemaChangePolicyBackfill SchemaChangePolicy = `backfill`

	OptSchemaChangePolicyNoBackfill SchemaChangePolicy = `nobackfill`

	OptSchemaChangePolicyStop SchemaChangePolicy = `stop`

	OptSchemaChangePolicyIgnore SchemaChangePolicy = `ignore`

	OptInitialScan = `initial_scan`

	OptNoInitialScan = `no_initial_scan`

	OptEmitAllResolvedTimestamps = ``

	OptInitialScanOnly = `initial_scan_only`

	OptEnvelopeKeyOnly       EnvelopeType = `key_only`
	OptEnvelopeRow           EnvelopeType = `row`
	OptEnvelopeDeprecatedRow EnvelopeType = `deprecated_row`
	OptEnvelopeWrapped       EnvelopeType = `wrapped`

	OptFormatJSON FormatType = `json`
	OptFormatAvro FormatType = `avro`

	OptFormatNative FormatType = `native`

	OptOnErrorFail  OnErrorType = `fail`
	OptOnErrorPause OnErrorType = `pause`

	DeprecatedOptFormatAvro                   = `experimental_avro`
	DeprecatedSinkSchemeCloudStorageAzure     = `experimental-azure`
	DeprecatedSinkSchemeCloudStorageGCS       = `experimental-gs`
	DeprecatedSinkSchemeCloudStorageHTTP      = `experimental-http`
	DeprecatedSinkSchemeCloudStorageHTTPS     = `experimental-https`
	DeprecatedSinkSchemeCloudStorageNodelocal = `experimental-nodelocal`
	DeprecatedSinkSchemeCloudStorageS3        = `experimental-s3`

	OptKafkaSinkConfig   = `kafka_sink_config`
	OptWebhookSinkConfig = `webhook_sink_config`

	OptSink = `sink`

	SinkParamCACert                 = `ca_cert`
	SinkParamClientCert             = `client_cert`
	SinkParamClientKey              = `client_key`
	SinkParamFileSize               = `file_size`
	SinkParamPartitionFormat        = `partition_format`
	SinkParamSchemaTopic            = `schema_topic`
	SinkParamTLSEnabled             = `tls_enabled`
	SinkParamSkipTLSVerify          = `insecure_tls_skip_verify`
	SinkParamTopicPrefix            = `topic_prefix`
	SinkParamTopicName              = `topic_name`
	SinkSchemeCloudStorageAzure     = `azure`
	SinkSchemeCloudStorageGCS       = `gs`
	SinkSchemeCloudStorageHTTP      = `http`
	SinkSchemeCloudStorageHTTPS     = `https`
	SinkSchemeCloudStorageNodelocal = `nodelocal`
	SinkSchemeCloudStorageS3        = `s3`
	SinkSchemeExperimentalSQL       = `experimental-sql`
	SinkSchemeHTTP                  = `http`
	SinkSchemeHTTPS                 = `https`
	SinkSchemeKafka                 = `kafka`
	SinkSchemeNull                  = `null`
	SinkSchemeWebhookHTTP           = `webhook-http`
	SinkSchemeWebhookHTTPS          = `webhook-https`
	SinkParamSASLEnabled            = `sasl_enabled`
	SinkParamSASLHandshake          = `sasl_handshake`
	SinkParamSASLUser               = `sasl_user`
	SinkParamSASLPassword           = `sasl_password`
	SinkParamSASLMechanism          = `sasl_mechanism`

	RegistryParamCACert = `ca_cert`

	Topics = `topics`
)

var ChangefeedOptionExpectValues = map[string]sql.KVStringOptValidate{
	OptAvroSchemaPrefix:         sql.KVStringOptRequireValue,
	OptConfluentSchemaRegistry:  sql.KVStringOptRequireValue,
	OptCursor:                   sql.KVStringOptRequireValue,
	OptEndTime:                  sql.KVStringOptRequireValue,
	OptEnvelope:                 sql.KVStringOptRequireValue,
	OptFormat:                   sql.KVStringOptRequireValue,
	OptFullTableName:            sql.KVStringOptRequireNoValue,
	OptKeyInValue:               sql.KVStringOptRequireNoValue,
	OptTopicInValue:             sql.KVStringOptRequireNoValue,
	OptResolvedTimestamps:       sql.KVStringOptAny,
	OptMinCheckpointFrequency:   sql.KVStringOptRequireValue,
	OptUpdatedTimestamps:        sql.KVStringOptRequireNoValue,
	OptMVCCTimestamps:           sql.KVStringOptRequireNoValue,
	OptDiff:                     sql.KVStringOptRequireNoValue,
	OptCompression:              sql.KVStringOptRequireValue,
	OptSchemaChangeEvents:       sql.KVStringOptRequireValue,
	OptSchemaChangePolicy:       sql.KVStringOptRequireValue,
	OptSplitColumnFamilies:      sql.KVStringOptRequireNoValue,
	OptInitialScan:              sql.KVStringOptAny,
	OptNoInitialScan:            sql.KVStringOptRequireNoValue,
	OptInitialScanOnly:          sql.KVStringOptRequireNoValue,
	OptProtectDataFromGCOnPause: sql.KVStringOptRequireNoValue,
	OptKafkaSinkConfig:          sql.KVStringOptRequireValue,
	OptWebhookSinkConfig:        sql.KVStringOptRequireValue,
	OptWebhookAuthHeader:        sql.KVStringOptRequireValue,
	OptWebhookClientTimeout:     sql.KVStringOptRequireValue,
	OptOnError:                  sql.KVStringOptRequireValue,
	OptMetricsScope:             sql.KVStringOptRequireValue,
	OptVirtualColumns:           sql.KVStringOptRequireValue,
}

func makeStringSet(opts ...string) map[string]struct{} {
	__antithesis_instrumentation__.Notify(16617)
	res := make(map[string]struct{}, len(opts))
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(16619)
		res[opt] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(16618)
	return res
}

var CommonOptions = makeStringSet(OptCursor, OptEndTime, OptEnvelope,
	OptFormat, OptFullTableName,
	OptKeyInValue, OptTopicInValue,
	OptResolvedTimestamps, OptUpdatedTimestamps,
	OptMVCCTimestamps, OptDiff, OptSplitColumnFamilies,
	OptSchemaChangeEvents, OptSchemaChangePolicy,
	OptProtectDataFromGCOnPause, OptOnError,
	OptInitialScan, OptNoInitialScan, OptInitialScanOnly,
	OptMinCheckpointFrequency, OptMetricsScope, OptVirtualColumns, Topics)

var SQLValidOptions map[string]struct{} = nil

var KafkaValidOptions = makeStringSet(OptAvroSchemaPrefix, OptConfluentSchemaRegistry, OptKafkaSinkConfig)

var CloudStorageValidOptions = makeStringSet(OptCompression)

var WebhookValidOptions = makeStringSet(OptWebhookAuthHeader, OptWebhookClientTimeout, OptWebhookSinkConfig)

var PubsubValidOptions = makeStringSet()

var CaseInsensitiveOpts = makeStringSet(OptFormat, OptEnvelope, OptCompression, OptSchemaChangeEvents, OptSchemaChangePolicy, OptOnError)

var NoLongerExperimental = map[string]string{
	DeprecatedOptFormatAvro:                   string(OptFormatAvro),
	DeprecatedSinkSchemeCloudStorageAzure:     SinkSchemeCloudStorageAzure,
	DeprecatedSinkSchemeCloudStorageGCS:       SinkSchemeCloudStorageGCS,
	DeprecatedSinkSchemeCloudStorageHTTP:      SinkSchemeCloudStorageHTTP,
	DeprecatedSinkSchemeCloudStorageHTTPS:     SinkSchemeCloudStorageHTTPS,
	DeprecatedSinkSchemeCloudStorageNodelocal: SinkSchemeCloudStorageNodelocal,
	DeprecatedSinkSchemeCloudStorageS3:        SinkSchemeCloudStorageS3,
}

var AlterChangefeedUnsupportedOptions = makeStringSet(OptCursor, OptInitialScan,
	OptNoInitialScan, OptInitialScanOnly, OptEndTime)

var AlterChangefeedOptionExpectValues = func() map[string]sql.KVStringOptValidate {
	__antithesis_instrumentation__.Notify(16620)
	alterChangefeedOptions := make(map[string]sql.KVStringOptValidate, len(ChangefeedOptionExpectValues)+1)
	for key, value := range ChangefeedOptionExpectValues {
		__antithesis_instrumentation__.Notify(16622)
		alterChangefeedOptions[key] = value
	}
	__antithesis_instrumentation__.Notify(16621)
	alterChangefeedOptions[OptSink] = sql.KVStringOptRequireValue
	return alterChangefeedOptions
}()

var AlterChangefeedTargetOptions = map[string]sql.KVStringOptValidate{
	OptInitialScan:   sql.KVStringOptRequireNoValue,
	OptNoInitialScan: sql.KVStringOptRequireNoValue,
}

var VersionGateOptions = map[string]clusterversion.Key{
	OptEndTime:         clusterversion.EnableNewChangefeedOptions,
	OptInitialScanOnly: clusterversion.EnableNewChangefeedOptions,
	OptInitialScan:     clusterversion.EnableNewChangefeedOptions,
}
