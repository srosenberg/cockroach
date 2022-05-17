package catformat

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

type IndexDisplayMode int

const (
	IndexDisplayShowCreate IndexDisplayMode = iota

	IndexDisplayDefOnly
)

func IndexForDisplay(
	ctx context.Context,
	table catalog.TableDescriptor,
	tableName *tree.TableName,
	index catalog.Index,
	partition string,
	formatFlags tree.FmtFlags,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	displayMode IndexDisplayMode,
) (string, error) {
	__antithesis_instrumentation__.Notify(247365)
	return indexForDisplay(
		ctx,
		table,
		tableName,
		index.IndexDesc(),
		index.Primary(),
		partition,
		formatFlags,
		semaCtx,
		sessionData,
		displayMode,
	)
}

func indexForDisplay(
	ctx context.Context,
	table catalog.TableDescriptor,
	tableName *tree.TableName,
	index *descpb.IndexDescriptor,
	isPrimary bool,
	partition string,
	formatFlags tree.FmtFlags,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
	displayMode IndexDisplayMode,
) (string, error) {
	__antithesis_instrumentation__.Notify(247366)

	if displayMode == IndexDisplayShowCreate && func() bool {
		__antithesis_instrumentation__.Notify(247378)
		return *tableName == descpb.AnonymousTable == true
	}() == true {
		__antithesis_instrumentation__.Notify(247379)
		return "", errors.New("tableName must be set for IndexDisplayShowCreate mode")
	} else {
		__antithesis_instrumentation__.Notify(247380)
	}
	__antithesis_instrumentation__.Notify(247367)

	f := tree.NewFmtCtx(formatFlags)
	if displayMode == IndexDisplayShowCreate {
		__antithesis_instrumentation__.Notify(247381)
		f.WriteString("CREATE ")
	} else {
		__antithesis_instrumentation__.Notify(247382)
	}
	__antithesis_instrumentation__.Notify(247368)
	if index.Unique {
		__antithesis_instrumentation__.Notify(247383)
		f.WriteString("UNIQUE ")
	} else {
		__antithesis_instrumentation__.Notify(247384)
	}
	__antithesis_instrumentation__.Notify(247369)
	if !f.HasFlags(tree.FmtPGCatalog) && func() bool {
		__antithesis_instrumentation__.Notify(247385)
		return index.Type == descpb.IndexDescriptor_INVERTED == true
	}() == true {
		__antithesis_instrumentation__.Notify(247386)
		f.WriteString("INVERTED ")
	} else {
		__antithesis_instrumentation__.Notify(247387)
	}
	__antithesis_instrumentation__.Notify(247370)
	f.WriteString("INDEX ")
	f.FormatNameP(&index.Name)
	if *tableName != descpb.AnonymousTable {
		__antithesis_instrumentation__.Notify(247388)
		f.WriteString(" ON ")
		f.FormatNode(tableName)
	} else {
		__antithesis_instrumentation__.Notify(247389)
	}
	__antithesis_instrumentation__.Notify(247371)

	if f.HasFlags(tree.FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(247390)
		f.WriteString(" USING")
		if index.Type == descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(247391)
			f.WriteString(" gin")
		} else {
			__antithesis_instrumentation__.Notify(247392)
			f.WriteString(" btree")
		}
	} else {
		__antithesis_instrumentation__.Notify(247393)
	}
	__antithesis_instrumentation__.Notify(247372)

	f.WriteString(" (")
	if err := FormatIndexElements(ctx, table, index, f, semaCtx, sessionData); err != nil {
		__antithesis_instrumentation__.Notify(247394)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(247395)
	}
	__antithesis_instrumentation__.Notify(247373)
	f.WriteByte(')')

	if index.IsSharded() {
		__antithesis_instrumentation__.Notify(247396)
		if f.HasFlags(tree.FmtPGCatalog) {
			__antithesis_instrumentation__.Notify(247397)
			fmt.Fprintf(f, " USING HASH WITH (bucket_count=%v)",
				index.Sharded.ShardBuckets)
		} else {
			__antithesis_instrumentation__.Notify(247398)
			f.WriteString(" USING HASH")
		}
	} else {
		__antithesis_instrumentation__.Notify(247399)
	}
	__antithesis_instrumentation__.Notify(247374)

	if !isPrimary && func() bool {
		__antithesis_instrumentation__.Notify(247400)
		return len(index.StoreColumnNames) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(247401)
		f.WriteString(" STORING (")
		for i := range index.StoreColumnNames {
			__antithesis_instrumentation__.Notify(247403)
			if i > 0 {
				__antithesis_instrumentation__.Notify(247405)
				f.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(247406)
			}
			__antithesis_instrumentation__.Notify(247404)
			f.FormatNameP(&index.StoreColumnNames[i])
		}
		__antithesis_instrumentation__.Notify(247402)
		f.WriteByte(')')
	} else {
		__antithesis_instrumentation__.Notify(247407)
	}
	__antithesis_instrumentation__.Notify(247375)

	f.WriteString(partition)

	if !f.HasFlags(tree.FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(247408)
		if err := formatStorageConfigs(table, index, f); err != nil {
			__antithesis_instrumentation__.Notify(247409)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(247410)
		}
	} else {
		__antithesis_instrumentation__.Notify(247411)
	}
	__antithesis_instrumentation__.Notify(247376)

	if index.IsPartial() {
		__antithesis_instrumentation__.Notify(247412)
		predFmtFlag := tree.FmtParsable
		if f.HasFlags(tree.FmtPGCatalog) {
			__antithesis_instrumentation__.Notify(247415)
			predFmtFlag = tree.FmtPGCatalog
		} else {
			__antithesis_instrumentation__.Notify(247416)
		}
		__antithesis_instrumentation__.Notify(247413)
		pred, err := schemaexpr.FormatExprForDisplay(ctx, table, index.Predicate, semaCtx, sessionData, predFmtFlag)
		if err != nil {
			__antithesis_instrumentation__.Notify(247417)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(247418)
		}
		__antithesis_instrumentation__.Notify(247414)

		f.WriteString(" WHERE ")
		if f.HasFlags(tree.FmtPGCatalog) {
			__antithesis_instrumentation__.Notify(247419)
			f.WriteString("(")
			f.WriteString(pred)
			f.WriteString(")")
		} else {
			__antithesis_instrumentation__.Notify(247420)
			f.WriteString(pred)
		}
	} else {
		__antithesis_instrumentation__.Notify(247421)
	}
	__antithesis_instrumentation__.Notify(247377)

	return f.CloseAndGetString(), nil
}

func FormatIndexElements(
	ctx context.Context,
	table catalog.TableDescriptor,
	index *descpb.IndexDescriptor,
	f *tree.FmtCtx,
	semaCtx *tree.SemaContext,
	sessionData *sessiondata.SessionData,
) error {
	__antithesis_instrumentation__.Notify(247422)
	elemFmtFlag := tree.FmtParsable
	if f.HasFlags(tree.FmtPGCatalog) {
		__antithesis_instrumentation__.Notify(247425)
		elemFmtFlag = tree.FmtPGCatalog
	} else {
		__antithesis_instrumentation__.Notify(247426)
	}
	__antithesis_instrumentation__.Notify(247423)

	startIdx := index.ExplicitColumnStartIdx()
	for i, n := startIdx, len(index.KeyColumnIDs); i < n; i++ {
		__antithesis_instrumentation__.Notify(247427)
		col, err := table.FindColumnWithID(index.KeyColumnIDs[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(247431)
			return err
		} else {
			__antithesis_instrumentation__.Notify(247432)
		}
		__antithesis_instrumentation__.Notify(247428)
		if i > startIdx {
			__antithesis_instrumentation__.Notify(247433)
			f.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(247434)
		}
		__antithesis_instrumentation__.Notify(247429)
		if col.IsExpressionIndexColumn() {
			__antithesis_instrumentation__.Notify(247435)
			expr, err := schemaexpr.FormatExprForExpressionIndexDisplay(
				ctx, table, col.GetComputeExpr(), semaCtx, sessionData, elemFmtFlag,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(247437)
				return err
			} else {
				__antithesis_instrumentation__.Notify(247438)
			}
			__antithesis_instrumentation__.Notify(247436)
			f.WriteString(expr)
		} else {
			__antithesis_instrumentation__.Notify(247439)
			f.FormatNameP(&index.KeyColumnNames[i])
		}
		__antithesis_instrumentation__.Notify(247430)
		if index.Type != descpb.IndexDescriptor_INVERTED {
			__antithesis_instrumentation__.Notify(247440)
			f.WriteByte(' ')
			f.WriteString(index.KeyColumnDirections[i].String())
		} else {
			__antithesis_instrumentation__.Notify(247441)
		}
	}
	__antithesis_instrumentation__.Notify(247424)
	return nil
}

func formatStorageConfigs(
	table catalog.TableDescriptor, index *descpb.IndexDescriptor, f *tree.FmtCtx,
) error {
	__antithesis_instrumentation__.Notify(247442)
	numCustomSettings := 0
	if index.GeoConfig.S2Geometry != nil || func() bool {
		__antithesis_instrumentation__.Notify(247446)
		return index.GeoConfig.S2Geography != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(247447)
		var s2Config *geoindex.S2Config

		if index.GeoConfig.S2Geometry != nil {
			__antithesis_instrumentation__.Notify(247451)
			s2Config = index.GeoConfig.S2Geometry.S2Config
		} else {
			__antithesis_instrumentation__.Notify(247452)
		}
		__antithesis_instrumentation__.Notify(247448)
		if index.GeoConfig.S2Geography != nil {
			__antithesis_instrumentation__.Notify(247453)
			s2Config = index.GeoConfig.S2Geography.S2Config
		} else {
			__antithesis_instrumentation__.Notify(247454)
		}
		__antithesis_instrumentation__.Notify(247449)

		defaultS2Config := geoindex.DefaultS2Config()
		if *s2Config != *defaultS2Config {
			__antithesis_instrumentation__.Notify(247455)
			for _, check := range []struct {
				key        string
				val        int32
				defaultVal int32
			}{
				{`s2_max_level`, s2Config.MaxLevel, defaultS2Config.MaxLevel},
				{`s2_level_mod`, s2Config.LevelMod, defaultS2Config.LevelMod},
				{`s2_max_cells`, s2Config.MaxCells, defaultS2Config.MaxCells},
			} {
				__antithesis_instrumentation__.Notify(247456)
				if check.val != check.defaultVal {
					__antithesis_instrumentation__.Notify(247457)
					if numCustomSettings > 0 {
						__antithesis_instrumentation__.Notify(247459)
						f.WriteString(", ")
					} else {
						__antithesis_instrumentation__.Notify(247460)
						f.WriteString(" WITH (")
					}
					__antithesis_instrumentation__.Notify(247458)
					numCustomSettings++
					f.WriteString(check.key)
					f.WriteString("=")
					f.WriteString(strconv.Itoa(int(check.val)))
				} else {
					__antithesis_instrumentation__.Notify(247461)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(247462)
		}
		__antithesis_instrumentation__.Notify(247450)

		if index.GeoConfig.S2Geometry != nil {
			__antithesis_instrumentation__.Notify(247463)
			col, err := table.FindColumnWithID(index.InvertedColumnID())
			if err != nil {
				__antithesis_instrumentation__.Notify(247466)
				return errors.Wrapf(err, "expected column %q to exist in table", index.InvertedColumnName())
			} else {
				__antithesis_instrumentation__.Notify(247467)
			}
			__antithesis_instrumentation__.Notify(247464)
			defaultConfig, err := geoindex.GeometryIndexConfigForSRID(col.GetType().GeoSRIDOrZero())
			if err != nil {
				__antithesis_instrumentation__.Notify(247468)
				return errors.Wrapf(err, "expected SRID definition for %d", col.GetType().GeoSRIDOrZero())
			} else {
				__antithesis_instrumentation__.Notify(247469)
			}
			__antithesis_instrumentation__.Notify(247465)
			cfg := index.GeoConfig.S2Geometry

			for _, check := range []struct {
				key        string
				val        float64
				defaultVal float64
			}{
				{`geometry_min_x`, cfg.MinX, defaultConfig.S2Geometry.MinX},
				{`geometry_max_x`, cfg.MaxX, defaultConfig.S2Geometry.MaxX},
				{`geometry_min_y`, cfg.MinY, defaultConfig.S2Geometry.MinY},
				{`geometry_max_y`, cfg.MaxY, defaultConfig.S2Geometry.MaxY},
			} {
				__antithesis_instrumentation__.Notify(247470)
				if check.val != check.defaultVal {
					__antithesis_instrumentation__.Notify(247471)
					if numCustomSettings > 0 {
						__antithesis_instrumentation__.Notify(247473)
						f.WriteString(", ")
					} else {
						__antithesis_instrumentation__.Notify(247474)
						f.WriteString(" WITH (")
					}
					__antithesis_instrumentation__.Notify(247472)
					numCustomSettings++
					f.WriteString(check.key)
					f.WriteString("=")
					f.WriteString(strconv.FormatFloat(check.val, 'f', -1, 64))
				} else {
					__antithesis_instrumentation__.Notify(247475)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(247476)
		}
	} else {
		__antithesis_instrumentation__.Notify(247477)
	}
	__antithesis_instrumentation__.Notify(247443)

	if index.IsSharded() {
		__antithesis_instrumentation__.Notify(247478)
		if numCustomSettings > 0 {
			__antithesis_instrumentation__.Notify(247480)
			f.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(247481)
			f.WriteString(" WITH (")
		}
		__antithesis_instrumentation__.Notify(247479)
		f.WriteString(`bucket_count=`)
		f.WriteString(strconv.FormatInt(int64(index.Sharded.ShardBuckets), 10))
		numCustomSettings++
	} else {
		__antithesis_instrumentation__.Notify(247482)
	}
	__antithesis_instrumentation__.Notify(247444)

	if numCustomSettings > 0 {
		__antithesis_instrumentation__.Notify(247483)
		f.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(247484)
	}
	__antithesis_instrumentation__.Notify(247445)

	return nil
}
