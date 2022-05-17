package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/settings"

var DummyVars = map[string]sessionVar{
	"enable_seqscan": makeDummyBooleanSessionVar(
		"enable_seqscan",
		func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631281)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().EnableSeqScan), nil
		},
		func(m sessionDataMutator, v bool) {
			__antithesis_instrumentation__.Notify(631282)
			m.SetEnableSeqScan(v)
		},
		func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(631283); return "on" },
	),
	"synchronous_commit": makeDummyBooleanSessionVar(
		"synchronous_commit",
		func(evalCtx *extendedEvalContext) (string, error) {
			__antithesis_instrumentation__.Notify(631284)
			return formatBoolAsPostgresSetting(evalCtx.SessionData().SynchronousCommit), nil
		},
		func(m sessionDataMutator, v bool) {
			__antithesis_instrumentation__.Notify(631285)
			m.SetSynchronousCommit(v)
		},
		func(sv *settings.Values) string { __antithesis_instrumentation__.Notify(631286); return "on" },
	),
}

var UnsupportedVars = func(ss ...string) map[string]struct{} {
	__antithesis_instrumentation__.Notify(631287)
	m := map[string]struct{}{}
	for _, s := range ss {
		__antithesis_instrumentation__.Notify(631289)
		m[s] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(631288)
	return m
}(

	"optimize_bounded_sort",

	"array_nulls",
	"backend_flush_after",

	"commit_delay",
	"commit_siblings",
	"constraint_exclusion",
	"cpu_index_tuple_cost",
	"cpu_operator_cost",
	"cpu_tuple_cost",
	"cursor_tuple_fraction",
	"deadlock_timeout",
	"debug_deadlocks",
	"debug_pretty_print",
	"debug_print_parse",
	"debug_print_plan",
	"debug_print_rewritten",
	"default_statistics_target",
	"default_text_search_config",
	"default_transaction_deferrable",

	"dynamic_library_path",
	"effective_cache_size",
	"enable_bitmapscan",
	"enable_gathermerge",
	"enable_hashagg",
	"enable_hashjoin",
	"enable_indexonlyscan",
	"enable_indexscan",
	"enable_material",
	"enable_mergejoin",
	"enable_nestloop",

	"enable_sort",
	"enable_tidscan",

	"exit_on_error",

	"force_parallel_mode",
	"from_collapse_limit",
	"geqo",
	"geqo_effort",
	"geqo_generations",
	"geqo_pool_size",
	"geqo_seed",
	"geqo_selection_bias",
	"geqo_threshold",
	"gin_fuzzy_search_limit",
	"gin_pending_list_limit",

	"ignore_checksum_failure",
	"join_collapse_limit",

	"lo_compat_privileges",
	"local_preload_libraries",

	"log_btree_build_stats",
	"log_duration",
	"log_error_verbosity",
	"log_executor_stats",
	"log_lock_waits",
	"log_min_duration_statement",
	"log_min_error_statement",
	"log_min_messages",
	"log_parser_stats",
	"log_planner_stats",
	"log_replication_commands",
	"log_statement",
	"log_statement_stats",
	"log_temp_files",
	"maintenance_work_mem",
	"max_parallel_workers",
	"max_parallel_workers_per_gather",
	"max_stack_depth",
	"min_parallel_index_scan_size",
	"min_parallel_table_scan_size",
	"operator_precedence_warning",
	"parallel_setup_cost",
	"parallel_tuple_cost",

	"quote_all_identifiers",
	"random_page_cost",
	"replacement_sort_tuples",

	"seed",
	"seq_page_cost",

	"session_preload_libraries",
	"session_replication_role",

	"tcp_keepalives_count",
	"tcp_keepalives_idle",
	"tcp_keepalives_interval",
	"temp_buffers",
	"temp_file_limit",
	"temp_tablespaces",
	"timezone_abbreviations",
	"trace_lock_oidmin",
	"trace_lock_table",
	"trace_locks",
	"trace_lwlocks",
	"trace_notify",
	"trace_sort",
	"trace_syncscan",
	"trace_userlocks",
	"track_activities",
	"track_counts",
	"track_functions",
	"track_io_timing",
	"transaction_deferrable",

	"transform_null_equals",
	"update_process_title",
	"vacuum_cost_delay",
	"vacuum_cost_limit",
	"vacuum_cost_page_dirty",
	"vacuum_cost_page_hit",
	"vacuum_cost_page_miss",
	"vacuum_freeze_min_age",
	"vacuum_freeze_table_age",
	"vacuum_multixact_freeze_min_age",
	"vacuum_multixact_freeze_table_age",
	"wal_compression",
	"wal_consistency_checking",
	"wal_debug",
	"work_mem",
	"xmlbinary",

	"zero_damaged_pages",
)
