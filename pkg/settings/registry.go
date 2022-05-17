package settings

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
	"unicode"
	"unicode/utf8"
)

var registry = make(map[string]internalSetting)

var slotTable [MaxSettings]internalSetting

func TestingSaveRegistry() func() {
	__antithesis_instrumentation__.Notify(240045)
	var origRegistry = make(map[string]internalSetting)
	for k, v := range registry {
		__antithesis_instrumentation__.Notify(240047)
		origRegistry[k] = v
	}
	__antithesis_instrumentation__.Notify(240046)
	return func() {
		__antithesis_instrumentation__.Notify(240048)
		registry = origRegistry
	}
}

var retiredSettings = map[string]struct{}{

	"kv.gc.batch_size":                     {},
	"kv.transaction.max_intents":           {},
	"diagnostics.reporting.report_metrics": {},

	"kv.allocator.stat_based_rebalancing.enabled": {},
	"kv.allocator.stat_rebalance_threshold":       {},

	"kv.raft_log.synchronize": {},

	"schemachanger.bulk_index_backfill.enabled":            {},
	"rocksdb.ingest_backpressure.delay_l0_file":            {},
	"server.heap_profile.system_memory_threshold_fraction": {},
	"timeseries.storage.10s_resolution_ttl":                {},
	"changefeed.push.enabled":                              {},
	"sql.defaults.optimizer":                               {},
	"kv.bulk_io_write.addsstable_max_rate":                 {},

	"schemachanger.lease.duration":           {},
	"schemachanger.lease.renew_fraction":     {},
	"diagnostics.forced_stat_reset.interval": {},

	"rocksdb.ingest_backpressure.pending_compaction_threshold":         {},
	"sql.distsql.temp_storage.joins":                                   {},
	"sql.distsql.temp_storage.sorts":                                   {},
	"sql.distsql.distribute_index_joins":                               {},
	"sql.distsql.merge_joins.enabled":                                  {},
	"sql.defaults.optimizer_foreign_keys.enabled":                      {},
	"sql.defaults.experimental_optimizer_foreign_key_cascades.enabled": {},
	"sql.parallel_scans.enabled":                                       {},
	"backup.table_statistics.enabled":                                  {},

	"sql.distsql.interleaved_joins.enabled": {},
	"sql.testing.vectorize.batch_size":      {},
	"sql.testing.mutations.max_batch_size":  {},
	"sql.testing.mock_contention.enabled":   {},
	"kv.atomic_replication_changes.enabled": {},

	"kv.tenant_rate_limiter.read_requests.rate_limit":   {},
	"kv.tenant_rate_limiter.read_requests.burst_limit":  {},
	"kv.tenant_rate_limiter.write_requests.rate_limit":  {},
	"kv.tenant_rate_limiter.write_requests.burst_limit": {},
	"kv.tenant_rate_limiter.read_bytes.rate_limit":      {},
	"kv.tenant_rate_limiter.read_bytes.burst_limit":     {},
	"kv.tenant_rate_limiter.write_bytes.rate_limit":     {},
	"kv.tenant_rate_limiter.write_bytes.burst_limit":    {},

	"sql.defaults.vectorize_row_count_threshold":                     {},
	"cloudstorage.gs.default.key":                                    {},
	"storage.sst_export.max_intents_per_error":                       {},
	"jobs.registry.leniency":                                         {},
	"sql.defaults.experimental_expression_based_indexes.enabled":     {},
	"kv.transaction.write_pipelining_max_outstanding_size":           {},
	"sql.defaults.optimizer_improve_disjunction_selectivity.enabled": {},
	"bulkio.backup.proxy_file_writes.enabled":                        {},
	"sql.distsql.prefer_local_execution.enabled":                     {},
	"kv.follower_read.target_multiple":                               {},
	"kv.closed_timestamp.close_fraction":                             {},
	"sql.telemetry.query_sampling.qps_threshold":                     {},
	"sql.telemetry.query_sampling.sample_rate":                       {},
	"diagnostics.sql_stat_reset.interval":                            {},
	"changefeed.mem.pushback_enabled":                                {},
	"sql.distsql.index_join_limit_hint.enabled":                      {},

	"sql.defaults.drop_enum_value.enabled":                             {},
	"trace.lightstep.token":                                            {},
	"trace.datadog.agent":                                              {},
	"trace.datadog.project":                                            {},
	"sql.defaults.interleaved_tables.enabled":                          {},
	"sql.defaults.copy_partitioning_when_deinterleaving_table.enabled": {},
	"server.declined_reservation_timeout":                              {},
	"bulkio.backup.resolve_destination_in_job.enabled":                 {},
	"sql.defaults.experimental_hash_sharded_indexes.enabled":           {},
	"schemachanger.backfiller.max_sst_size":                            {},
	"kv.bulk_ingest.buffer_increment":                                  {},
	"schemachanger.backfiller.buffer_increment":                        {},
}

func register(class Class, key, desc string, s internalSetting) {
	__antithesis_instrumentation__.Notify(240049)
	if _, ok := retiredSettings[key]; ok {
		__antithesis_instrumentation__.Notify(240055)
		panic(fmt.Sprintf("cannot reuse previously defined setting name: %s", key))
	} else {
		__antithesis_instrumentation__.Notify(240056)
	}
	__antithesis_instrumentation__.Notify(240050)
	if _, ok := registry[key]; ok {
		__antithesis_instrumentation__.Notify(240057)
		panic(fmt.Sprintf("setting already defined: %s", key))
	} else {
		__antithesis_instrumentation__.Notify(240058)
	}
	__antithesis_instrumentation__.Notify(240051)
	if len(desc) == 0 {
		__antithesis_instrumentation__.Notify(240059)
		panic(fmt.Sprintf("setting missing description: %s", key))
	} else {
		__antithesis_instrumentation__.Notify(240060)
	}
	__antithesis_instrumentation__.Notify(240052)
	if r, _ := utf8.DecodeRuneInString(desc); unicode.IsUpper(r) {
		__antithesis_instrumentation__.Notify(240061)
		panic(fmt.Sprintf(
			"setting descriptions should start with a lowercase letter: %q, %q", key, desc,
		))
	} else {
		__antithesis_instrumentation__.Notify(240062)
	}
	__antithesis_instrumentation__.Notify(240053)
	for _, c := range desc {
		__antithesis_instrumentation__.Notify(240063)
		if c == unicode.ReplacementChar {
			__antithesis_instrumentation__.Notify(240065)
			panic(fmt.Sprintf("setting descriptions must be valid UTF-8: %q, %q", key, desc))
		} else {
			__antithesis_instrumentation__.Notify(240066)
		}
		__antithesis_instrumentation__.Notify(240064)
		if unicode.IsControl(c) {
			__antithesis_instrumentation__.Notify(240067)
			panic(fmt.Sprintf(
				"setting descriptions cannot contain control character %q: %q, %q", c, key, desc,
			))
		} else {
			__antithesis_instrumentation__.Notify(240068)
		}
	}
	__antithesis_instrumentation__.Notify(240054)
	slot := slotIdx(len(registry))
	s.init(class, key, desc, slot)
	registry[key] = s
	slotTable[slot] = s
}

func NumRegisteredSettings() int { __antithesis_instrumentation__.Notify(240069); return len(registry) }

func Keys(forSystemTenant bool) (res []string) {
	__antithesis_instrumentation__.Notify(240070)
	res = make([]string, 0, len(registry))
	for k, v := range registry {
		__antithesis_instrumentation__.Notify(240072)
		if v.isRetired() {
			__antithesis_instrumentation__.Notify(240075)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240076)
		}
		__antithesis_instrumentation__.Notify(240073)
		if !forSystemTenant && func() bool {
			__antithesis_instrumentation__.Notify(240077)
			return v.Class() == SystemOnly == true
		}() == true {
			__antithesis_instrumentation__.Notify(240078)
			continue
		} else {
			__antithesis_instrumentation__.Notify(240079)
		}
		__antithesis_instrumentation__.Notify(240074)
		res = append(res, k)
	}
	__antithesis_instrumentation__.Notify(240071)
	sort.Strings(res)
	return res
}

func Lookup(name string, purpose LookupPurpose, forSystemTenant bool) (Setting, bool) {
	__antithesis_instrumentation__.Notify(240080)
	s, ok := registry[name]
	if !ok {
		__antithesis_instrumentation__.Notify(240084)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(240085)
	}
	__antithesis_instrumentation__.Notify(240081)
	if !forSystemTenant && func() bool {
		__antithesis_instrumentation__.Notify(240086)
		return s.Class() == SystemOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(240087)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(240088)
	}
	__antithesis_instrumentation__.Notify(240082)
	if purpose == LookupForReporting && func() bool {
		__antithesis_instrumentation__.Notify(240089)
		return !s.isReportable() == true
	}() == true {
		__antithesis_instrumentation__.Notify(240090)
		return &MaskedSetting{setting: s}, true
	} else {
		__antithesis_instrumentation__.Notify(240091)
	}
	__antithesis_instrumentation__.Notify(240083)
	return s, true
}

type LookupPurpose int

const (
	LookupForReporting LookupPurpose = iota

	LookupForLocalAccess
)

const ForSystemTenant = true

var ReadableTypes = map[string]string{
	"s": "string",
	"i": "integer",
	"f": "float",
	"b": "boolean",
	"z": "byte size",
	"d": "duration",
	"e": "enumeration",

	"m": "version",
}

func RedactedValue(name string, values *Values, forSystemTenant bool) string {
	__antithesis_instrumentation__.Notify(240092)
	if setting, ok := Lookup(name, LookupForReporting, forSystemTenant); ok {
		__antithesis_instrumentation__.Notify(240094)
		return setting.String(values)
	} else {
		__antithesis_instrumentation__.Notify(240095)
	}
	__antithesis_instrumentation__.Notify(240093)
	return "<unknown>"
}
