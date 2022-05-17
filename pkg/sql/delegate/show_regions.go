package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

func (d *delegator) delegateShowRegions(n *tree.ShowRegions) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465729)
	zonesClause := `
		SELECT
			region, zones
		FROM crdb_internal.regions
		ORDER BY region
`
	switch n.ShowRegionsFrom {
	case tree.ShowRegionsFromAllDatabases:
		__antithesis_instrumentation__.Notify(465731)
		sqltelemetry.IncrementShowCounter(sqltelemetry.RegionsFromAllDatabases)
		return parse(`
SELECT
	name as database_name,
	regions,
	primary_region
FROM crdb_internal.databases
ORDER BY database_name
			`,
		)

	case tree.ShowRegionsFromDatabase:
		__antithesis_instrumentation__.Notify(465732)
		sqltelemetry.IncrementShowCounter(sqltelemetry.RegionsFromDatabase)
		dbName := string(n.DatabaseName)
		if dbName == "" {
			__antithesis_instrumentation__.Notify(465738)
			dbName = d.evalCtx.SessionData().Database
		} else {
			__antithesis_instrumentation__.Notify(465739)
		}
		__antithesis_instrumentation__.Notify(465733)

		query := fmt.Sprintf(
			`
WITH zones_table(region, zones) AS (%s)
SELECT
	r.name AS "database",
	r.region as "region",
	r.region = r.primary_region AS "primary",
	COALESCE(zones_table.zones, '{}'::string[])
AS
	zones
FROM [
	SELECT
		name,
		unnest(dbs.regions) AS region,
		dbs.primary_region AS primary_region
	FROM crdb_internal.databases dbs
	WHERE dbs.name = %s
] r
LEFT JOIN zones_table ON (r.region = zones_table.region)
ORDER BY "primary" DESC, "region"`,
			zonesClause,
			lexbase.EscapeSQLString(dbName),
		)

		return parse(query)

	case tree.ShowRegionsFromCluster:
		__antithesis_instrumentation__.Notify(465734)
		sqltelemetry.IncrementShowCounter(sqltelemetry.RegionsFromCluster)

		query := fmt.Sprintf(
			`
SELECT
	region, zones
FROM
	(%s)
WHERE
	region IS NOT NULL
ORDER BY
	region`,
			zonesClause,
		)

		return parse(query)
	case tree.ShowRegionsFromDefault:
		__antithesis_instrumentation__.Notify(465735)
		sqltelemetry.IncrementShowCounter(sqltelemetry.Regions)

		query := fmt.Sprintf(
			`
WITH databases_by_region(region, database_names) AS (
	SELECT
		region,
		array_agg(name) as database_names
	FROM [
		SELECT
			name,
			unnest(regions) AS region
		FROM crdb_internal.databases
	] GROUP BY region
),
databases_by_primary_region(region, database_names) AS (
	SELECT
		primary_region,
		array_agg(name)
	FROM crdb_internal.databases
	GROUP BY primary_region
),
zones_table(region, zones) AS (%s)
SELECT
	zones_table.region,
	zones_table.zones,
	COALESCE(databases_by_region.database_names, '{}'::string[]) AS database_names,
	COALESCE(databases_by_primary_region.database_names, '{}'::string[]) AS primary_region_of
FROM zones_table
LEFT JOIN databases_by_region ON (zones_table.region = databases_by_region.region)
LEFT JOIN databases_by_primary_region ON (zones_table.region = databases_by_primary_region.region)
ORDER BY zones_table.region
`,
			zonesClause,
		)
		return parse(query)
	case tree.ShowSuperRegionsFromDatabase:
		__antithesis_instrumentation__.Notify(465736)
		sqltelemetry.IncrementShowCounter(sqltelemetry.SuperRegions)

		query := fmt.Sprintf(
			`
SELECT database_name, super_region_name, regions
  FROM crdb_internal.super_regions
 WHERE database_name = '%s'`, n.DatabaseName)

		return parse(query)
	default:
		__antithesis_instrumentation__.Notify(465737)
	}
	__antithesis_instrumentation__.Notify(465730)
	return nil, errors.Newf("unhandled ShowRegionsFrom: %v", n.ShowRegionsFrom)
}
