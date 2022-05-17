package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
)

var (
	CreateMultiRegionDatabaseCounter = telemetry.GetCounterOnce(
		"sql.multiregion.create_database",
	)

	SetInitialPrimaryRegionCounter = telemetry.GetCounterOnce(
		"sql.multiregion.alter_database.set_primary_region.initial_multiregion",
	)

	SwitchPrimaryRegionCounter = telemetry.GetCounterOnce(
		"sql.multiregion.alter_database.set_primary_region.switch_primary_region",
	)

	AlterDatabaseAddRegionCounter = telemetry.GetCounterOnce(
		"sql.multiregion.add_region",
	)

	AlterDatabaseDropRegionCounter = telemetry.GetCounterOnce(
		"sql.multiregion.drop_region",
	)

	AlterDatabaseDropPrimaryRegionCounter = telemetry.GetCounterOnce(
		"sql.multiregion.drop_primary_region",
	)

	ImportIntoMultiRegionDatabaseCounter = telemetry.GetCounterOnce(
		"sql.multiregion.import",
	)

	OverrideMultiRegionZoneConfigurationUser = telemetry.GetCounterOnce(
		"sql.multiregion.zone_configuration.override.user",
	)

	OverrideMultiRegionDatabaseZoneConfigurationSystem = telemetry.GetCounterOnce(
		"sql.multiregion.zone_configuration.override.system.database",
	)

	OverrideMultiRegionTableZoneConfigurationSystem = telemetry.GetCounterOnce(
		"sql.multiregion.zone_configuration.override.system.table",
	)
)

func CreateDatabaseSurvivalGoalCounter(goal string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625798)
	return telemetry.GetCounter(fmt.Sprintf("sql.multiregion.create_database.survival_goal.%s", goal))
}

func AlterDatabaseSurvivalGoalCounter(goal string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625799)
	return telemetry.GetCounter(fmt.Sprintf("sql.multiregion.alter_database.survival_goal.%s", goal))
}

func CreateDatabasePlacementCounter(placement string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625800)
	return telemetry.GetCounter(
		fmt.Sprintf("sql.multiregion.create_database.placement.%s", placement),
	)
}

func AlterDatabasePlacementCounter(placement string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625801)
	return telemetry.GetCounter(
		fmt.Sprintf("sql.multiregion.alter_database.placement.%s", placement),
	)
}

func CreateTableLocalityCounter(locality string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625802)
	return telemetry.GetCounter(
		fmt.Sprintf("sql.multiregion.create_table.locality.%s", locality),
	)
}

func AlterTableLocalityCounter(from, to string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625803)
	return telemetry.GetCounter(
		fmt.Sprintf("sql.multiregion.alter_table.locality.from.%s.to.%s", from, to),
	)
}
