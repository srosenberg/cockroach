package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

func SerialColumnNormalizationCounter(inputType, normType string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625826)
	return telemetry.GetCounter(fmt.Sprintf("sql.schema.serial.%s.%s", normType, inputType))
}

func SchemaNewTypeCounter(t string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625827)
	return telemetry.GetCounter("sql.schema.new_column_type." + t)
}

var (
	CreateTempTableCounter = telemetry.GetCounterOnce("sql.schema.create_temp_table")

	CreateTempSequenceCounter = telemetry.GetCounterOnce("sql.schema.create_temp_sequence")

	CreateTempViewCounter = telemetry.GetCounterOnce("sql.schema.create_temp_view")
)

var (
	HashShardedIndexCounter = telemetry.GetCounterOnce("sql.schema.hash_sharded_index")

	InvertedIndexCounter = telemetry.GetCounterOnce("sql.schema.inverted_index")

	MultiColumnInvertedIndexCounter = telemetry.GetCounterOnce("sql.schema.multi_column_inverted_index")

	GeographyInvertedIndexCounter = telemetry.GetCounterOnce("sql.schema.geography_inverted_index")

	GeometryInvertedIndexCounter = telemetry.GetCounterOnce("sql.schema.geometry_inverted_index")

	PartialIndexCounter = telemetry.GetCounterOnce("sql.schema.partial_index")

	PartialInvertedIndexCounter = telemetry.GetCounterOnce("sql.schema.partial_inverted_index")

	PartitionedInvertedIndexCounter = telemetry.GetCounterOnce("sql.schema.partitioned_inverted_index")

	ExpressionIndexCounter = telemetry.GetCounterOnce("sql.schema.expression_index")
)

var (
	TempObjectCleanerDeletionCounter = telemetry.GetCounterOnce("sql.schema.temp_object_cleaner.num_cleaned")
)

func SchemaNewColumnTypeQualificationCounter(qual string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625828)
	return telemetry.GetCounter("sql.schema.new_column.qualification." + qual)
}

func SchemaChangeCreateCounter(typ string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625829)
	return telemetry.GetCounter("sql.schema.create_" + typ)
}

func SchemaChangeDropCounter(typ string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625830)
	return telemetry.GetCounter("sql.schema.drop_" + typ)
}

func SchemaSetZoneConfigCounter(configName, keyChange string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625831)
	return telemetry.GetCounter(
		fmt.Sprintf("sql.schema.zone_config.%s.%s", configName, keyChange),
	)
}

func SchemaChangeAlterCounter(typ string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625832)
	return SchemaChangeAlterCounterWithExtra(typ, "")
}

func SchemaChangeAlterCounterWithExtra(typ string, extra string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625833)
	if extra != "" {
		__antithesis_instrumentation__.Notify(625835)
		extra = "." + extra
	} else {
		__antithesis_instrumentation__.Notify(625836)
	}
	__antithesis_instrumentation__.Notify(625834)
	return telemetry.GetCounter(fmt.Sprintf("sql.schema.alter_%s%s", typ, extra))
}

func SchemaSetAuditModeCounter(mode string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625837)
	return telemetry.GetCounter("sql.schema.set_audit_mode." + mode)
}

func SchemaJobControlCounter(desiredStatus string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625838)
	return telemetry.GetCounter("sql.schema.job.control." + desiredStatus)
}

var SchemaChangeInExplicitTxnCounter = telemetry.GetCounterOnce("sql.schema.change_in_explicit_txn")

var SecondaryIndexColumnFamiliesCounter = telemetry.GetCounterOnce("sql.schema.secondary_index_column_families")

var CreateUnloggedTableCounter = telemetry.GetCounterOnce("sql.schema.create_unlogged_table")

var SchemaRefreshMaterializedView = telemetry.GetCounterOnce("sql.schema.refresh_materialized_view")

func SchemaChangeErrorCounter(typ string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625839)
	return telemetry.GetCounter(fmt.Sprintf("sql.schema_changer.errors.%s", typ))
}

func SetTableStorageParameter(param string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625840)
	return telemetry.GetCounter("sql.schema.table_storage_parameter." + param + ".set")
}

func ResetTableStorageParameter(param string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625841)
	return telemetry.GetCounter("sql.schema.table_storage_parameter." + param + ".reset")
}
