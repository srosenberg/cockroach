package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/geo/geoprojbase"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var pgExtension = virtualSchema{
	name: catconstants.PgExtensionSchemaName,
	tableDefs: map[descpb.ID]virtualSchemaDef{
		catconstants.PgExtensionGeographyColumnsTableID: pgExtensionGeographyColumnsTable,
		catconstants.PgExtensionGeometryColumnsTableID:  pgExtensionGeometryColumnsTable,
		catconstants.PgExtensionSpatialRefSysTableID:    pgExtensionSpatialRefSysTable,
	},
	validWithNoDatabaseContext: false,
}

func postgisColumnsTablePopulator(
	matchingFamily types.Family,
) func(context.Context, *planner, catalog.DatabaseDescriptor, func(...tree.Datum) error) error {
	__antithesis_instrumentation__.Notify(558742)
	return func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558743)
		return forEachTableDesc(
			ctx,
			p,
			dbContext,
			hideVirtual,
			func(db catalog.DatabaseDescriptor, scName string, table catalog.TableDescriptor) error {
				__antithesis_instrumentation__.Notify(558744)
				if !table.IsPhysicalTable() {
					__antithesis_instrumentation__.Notify(558748)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558749)
				}
				__antithesis_instrumentation__.Notify(558745)
				if p.CheckAnyPrivilege(ctx, table) != nil {
					__antithesis_instrumentation__.Notify(558750)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(558751)
				}
				__antithesis_instrumentation__.Notify(558746)
				for _, col := range table.PublicColumns() {
					__antithesis_instrumentation__.Notify(558752)
					if col.GetType().Family() != matchingFamily {
						__antithesis_instrumentation__.Notify(558757)
						continue
					} else {
						__antithesis_instrumentation__.Notify(558758)
					}
					__antithesis_instrumentation__.Notify(558753)
					m, err := col.GetType().GeoMetadata()
					if err != nil {
						__antithesis_instrumentation__.Notify(558759)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558760)
					}
					__antithesis_instrumentation__.Notify(558754)

					var datumNDims tree.Datum
					switch m.ShapeType {
					case geopb.ShapeType_Geometry, geopb.ShapeType_Unset:
						__antithesis_instrumentation__.Notify(558761)

						if matchingFamily == types.GeometryFamily {
							__antithesis_instrumentation__.Notify(558763)
							datumNDims = tree.NewDInt(2)
						} else {
							__antithesis_instrumentation__.Notify(558764)
							datumNDims = tree.DNull
						}
					default:
						__antithesis_instrumentation__.Notify(558762)
						zm := m.ShapeType & (geopb.ZShapeTypeFlag | geopb.MShapeTypeFlag)
						switch zm {
						case geopb.ZShapeTypeFlag | geopb.MShapeTypeFlag:
							__antithesis_instrumentation__.Notify(558765)
							datumNDims = tree.NewDInt(4)
						case geopb.ZShapeTypeFlag, geopb.MShapeTypeFlag:
							__antithesis_instrumentation__.Notify(558766)
							datumNDims = tree.NewDInt(3)
						default:
							__antithesis_instrumentation__.Notify(558767)
							datumNDims = tree.NewDInt(2)
						}
					}
					__antithesis_instrumentation__.Notify(558755)

					shapeName := geopb.ShapeType_Geometry.String()
					if matchingFamily == types.GeometryFamily {
						__antithesis_instrumentation__.Notify(558768)
						if m.ShapeType == geopb.ShapeType_Unset {
							__antithesis_instrumentation__.Notify(558769)
							shapeName = strings.ToUpper(shapeName)
						} else {
							__antithesis_instrumentation__.Notify(558770)
							shapeName = strings.ToUpper(m.ShapeType.To2D().String())
						}
					} else {
						__antithesis_instrumentation__.Notify(558771)
						if m.ShapeType != geopb.ShapeType_Unset {
							__antithesis_instrumentation__.Notify(558772)
							shapeName = m.ShapeType.String()
						} else {
							__antithesis_instrumentation__.Notify(558773)
						}
					}
					__antithesis_instrumentation__.Notify(558756)

					if err := addRow(
						tree.NewDString(db.GetName()),
						tree.NewDString(scName),
						tree.NewDString(table.GetName()),
						tree.NewDString(col.GetName()),
						datumNDims,
						tree.NewDInt(tree.DInt(m.SRID)),
						tree.NewDString(shapeName),
					); err != nil {
						__antithesis_instrumentation__.Notify(558774)
						return err
					} else {
						__antithesis_instrumentation__.Notify(558775)
					}
				}
				__antithesis_instrumentation__.Notify(558747)
				return nil
			},
		)
	}
}

var pgExtensionGeographyColumnsTable = virtualSchemaTable{
	comment: `Shows all defined geography columns. Matches PostGIS' geography_columns functionality.`,
	schema: `
CREATE TABLE pg_extension.geography_columns (
	f_table_catalog name,
	f_table_schema name,
	f_table_name name,
	f_geography_column name,
	coord_dimension integer,
	srid integer,
	type text
)`,
	populate: postgisColumnsTablePopulator(types.GeographyFamily),
}

var pgExtensionGeometryColumnsTable = virtualSchemaTable{
	comment: `Shows all defined geometry columns. Matches PostGIS' geometry_columns functionality.`,
	schema: `
CREATE TABLE pg_extension.geometry_columns (
	f_table_catalog name,
	f_table_schema name,
	f_table_name name,
	f_geometry_column name,
	coord_dimension integer,
	srid integer,
	type text
)`,
	populate: postgisColumnsTablePopulator(types.GeometryFamily),
}

var pgExtensionSpatialRefSysTable = virtualSchemaTable{
	comment: `Shows all defined Spatial Reference Identifiers (SRIDs). Matches PostGIS' spatial_ref_sys table.`,
	schema: `
CREATE TABLE pg_extension.spatial_ref_sys (
	srid integer,
	auth_name varchar(256),
	auth_srid integer,
	srtext varchar(2048),
	proj4text varchar(2048)
)`,
	populate: func(ctx context.Context, p *planner, dbContext catalog.DatabaseDescriptor, addRow func(...tree.Datum) error) error {
		__antithesis_instrumentation__.Notify(558776)
		for _, projection := range geoprojbase.AllProjections() {
			__antithesis_instrumentation__.Notify(558778)
			if err := addRow(
				tree.NewDInt(tree.DInt(projection.SRID)),
				tree.NewDString(projection.AuthName),
				tree.NewDInt(tree.DInt(projection.AuthSRID)),
				tree.NewDString(projection.SRText),
				tree.NewDString(projection.Proj4Text.String()),
			); err != nil {
				__antithesis_instrumentation__.Notify(558779)
				return err
			} else {
				__antithesis_instrumentation__.Notify(558780)
			}
		}
		__antithesis_instrumentation__.Notify(558777)
		return nil
	},
}
