package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	encjson "encoding/json"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/errors"
)

var showTableStatsColumns = colinfo.ResultColumns{
	{Name: "statistics_name", Typ: types.String},
	{Name: "column_names", Typ: types.StringArray},
	{Name: "created", Typ: types.Timestamp},
	{Name: "row_count", Typ: types.Int},
	{Name: "distinct_count", Typ: types.Int},
	{Name: "null_count", Typ: types.Int},
	{Name: "histogram_id", Typ: types.Int},
}

var showTableStatsColumnsAvgSizeVer = colinfo.ResultColumns{
	{Name: "statistics_name", Typ: types.String},
	{Name: "column_names", Typ: types.StringArray},
	{Name: "created", Typ: types.Timestamp},
	{Name: "row_count", Typ: types.Int},
	{Name: "distinct_count", Typ: types.Int},
	{Name: "null_count", Typ: types.Int},
	{Name: "avg_size", Typ: types.Int},
	{Name: "histogram_id", Typ: types.Int},
}

var showTableStatsJSONColumns = colinfo.ResultColumns{
	{Name: "statistics", Typ: types.Jsonb},
}

func (p *planner) ShowTableStats(ctx context.Context, n *tree.ShowTableStats) (planNode, error) {
	__antithesis_instrumentation__.Notify(623217)

	desc, err := p.ResolveUncachedTableDescriptorEx(ctx, n.Table, true, tree.ResolveRequireTableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(623222)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623223)
	}
	__antithesis_instrumentation__.Notify(623218)
	if err := p.CheckAnyPrivilege(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(623224)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623225)
	}
	__antithesis_instrumentation__.Notify(623219)
	avgSizeColVerActive := p.ExtendedEvalContext().ExecCfg.Settings.Version.IsActive(ctx, clusterversion.AlterSystemTableStatisticsAddAvgSizeCol)
	columns := showTableStatsColumnsAvgSizeVer
	if !avgSizeColVerActive {
		__antithesis_instrumentation__.Notify(623226)
		columns = showTableStatsColumns
	} else {
		__antithesis_instrumentation__.Notify(623227)
	}
	__antithesis_instrumentation__.Notify(623220)
	if n.UsingJSON {
		__antithesis_instrumentation__.Notify(623228)
		columns = showTableStatsJSONColumns
	} else {
		__antithesis_instrumentation__.Notify(623229)
	}
	__antithesis_instrumentation__.Notify(623221)

	return &delayedNode{
		name:    n.String(),
		columns: columns,
		constructor: func(ctx context.Context, p *planner) (_ planNode, err error) {
			__antithesis_instrumentation__.Notify(623230)

			var avgSize string
			if avgSizeColVerActive {
				__antithesis_instrumentation__.Notify(623237)
				avgSize = `
					"avgSize",`
			} else {
				__antithesis_instrumentation__.Notify(623238)
			}
			__antithesis_instrumentation__.Notify(623231)
			stmt := fmt.Sprintf(`SELECT "statisticID",
																				 name,
																				 "columnIDs",
																				 "createdAt",
																				 "rowCount",
																				 "distinctCount",
																				 "nullCount",
																				 %s
																				 histogram
																	FROM system.table_statistics
																	WHERE "tableID" = $1
																	ORDER BY "createdAt"`, avgSize)
			rows, err := p.ExtendedEvalContext().ExecCfg.InternalExecutor.QueryBuffered(
				ctx,
				"read-table-stats",
				p.txn,
				stmt,
				desc.GetID(),
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(623239)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623240)
			}
			__antithesis_instrumentation__.Notify(623232)

			const (
				statIDIdx = iota
				nameIdx
				columnIDsIdx
				createdAtIdx
				rowCountIdx
				distinctCountIdx
				nullCountIdx
				avgSizeIdx
				histogramIdx
				numCols
			)

			histIdx := histogramIdx
			nCols := numCols
			if !avgSizeColVerActive {
				__antithesis_instrumentation__.Notify(623241)
				histIdx = histogramIdx - 1
				nCols = numCols - 1
			} else {
				__antithesis_instrumentation__.Notify(623242)
			}
			__antithesis_instrumentation__.Notify(623233)

			defer func() {
				__antithesis_instrumentation__.Notify(623243)
				if r := recover(); r != nil {
					__antithesis_instrumentation__.Notify(623244)

					if ok, e := errorutil.ShouldCatch(r); ok {
						__antithesis_instrumentation__.Notify(623245)
						err = e
					} else {
						__antithesis_instrumentation__.Notify(623246)

						panic(r)
					}
				} else {
					__antithesis_instrumentation__.Notify(623247)
				}
			}()
			__antithesis_instrumentation__.Notify(623234)

			v := p.newContainerValuesNode(columns, 0)
			if n.UsingJSON {
				__antithesis_instrumentation__.Notify(623248)
				result := make([]stats.JSONStatistic, len(rows))
				for i, r := range rows {
					__antithesis_instrumentation__.Notify(623253)
					result[i].CreatedAt = tree.AsStringWithFlags(r[createdAtIdx], tree.FmtBareStrings)
					result[i].RowCount = (uint64)(*r[rowCountIdx].(*tree.DInt))
					result[i].DistinctCount = (uint64)(*r[distinctCountIdx].(*tree.DInt))
					result[i].NullCount = (uint64)(*r[nullCountIdx].(*tree.DInt))
					if avgSizeColVerActive {
						__antithesis_instrumentation__.Notify(623257)
						result[i].AvgSize = (uint64)(*r[avgSizeIdx].(*tree.DInt))
					} else {
						__antithesis_instrumentation__.Notify(623258)
					}
					__antithesis_instrumentation__.Notify(623254)
					if r[nameIdx] != tree.DNull {
						__antithesis_instrumentation__.Notify(623259)
						result[i].Name = string(*r[nameIdx].(*tree.DString))
					} else {
						__antithesis_instrumentation__.Notify(623260)
					}
					__antithesis_instrumentation__.Notify(623255)
					colIDs := r[columnIDsIdx].(*tree.DArray).Array
					result[i].Columns = make([]string, len(colIDs))
					for j, d := range colIDs {
						__antithesis_instrumentation__.Notify(623261)
						result[i].Columns[j] = statColumnString(desc, d)
					}
					__antithesis_instrumentation__.Notify(623256)
					if err := result[i].DecodeAndSetHistogram(ctx, &p.semaCtx, r[histIdx]); err != nil {
						__antithesis_instrumentation__.Notify(623262)
						v.Close(ctx)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(623263)
					}
				}
				__antithesis_instrumentation__.Notify(623249)
				encoded, err := encjson.Marshal(result)
				if err != nil {
					__antithesis_instrumentation__.Notify(623264)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623265)
				}
				__antithesis_instrumentation__.Notify(623250)
				jsonResult, err := json.ParseJSON(string(encoded))
				if err != nil {
					__antithesis_instrumentation__.Notify(623266)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623267)
				}
				__antithesis_instrumentation__.Notify(623251)
				if _, err := v.rows.AddRow(ctx, tree.Datums{tree.NewDJSON(jsonResult)}); err != nil {
					__antithesis_instrumentation__.Notify(623268)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623269)
				}
				__antithesis_instrumentation__.Notify(623252)
				return v, nil
			} else {
				__antithesis_instrumentation__.Notify(623270)
			}
			__antithesis_instrumentation__.Notify(623235)

			for _, r := range rows {
				__antithesis_instrumentation__.Notify(623271)
				if len(r) != nCols {
					__antithesis_instrumentation__.Notify(623276)
					v.Close(ctx)
					return nil, errors.Errorf("incorrect columns from internal query")
				} else {
					__antithesis_instrumentation__.Notify(623277)
				}
				__antithesis_instrumentation__.Notify(623272)

				colIDs := r[columnIDsIdx].(*tree.DArray).Array
				colNames := tree.NewDArray(types.String)
				colNames.Array = make(tree.Datums, len(colIDs))
				for i, d := range colIDs {
					__antithesis_instrumentation__.Notify(623278)
					colNames.Array[i] = tree.NewDString(statColumnString(desc, d))
				}
				__antithesis_instrumentation__.Notify(623273)

				histogramID := tree.DNull
				if r[histIdx] != tree.DNull {
					__antithesis_instrumentation__.Notify(623279)
					histogramID = r[statIDIdx]
				} else {
					__antithesis_instrumentation__.Notify(623280)
				}
				__antithesis_instrumentation__.Notify(623274)

				var res tree.Datums
				if avgSizeColVerActive {
					__antithesis_instrumentation__.Notify(623281)
					res = tree.Datums{
						r[nameIdx],
						colNames,
						r[createdAtIdx],
						r[rowCountIdx],
						r[distinctCountIdx],
						r[nullCountIdx],
						r[avgSizeIdx],
						histogramID,
					}
				} else {
					__antithesis_instrumentation__.Notify(623282)
					res = tree.Datums{
						r[nameIdx],
						colNames,
						r[createdAtIdx],
						r[rowCountIdx],
						r[distinctCountIdx],
						r[nullCountIdx],
						histogramID,
					}
				}
				__antithesis_instrumentation__.Notify(623275)

				if _, err := v.rows.AddRow(ctx, res); err != nil {
					__antithesis_instrumentation__.Notify(623283)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623284)
				}
			}
			__antithesis_instrumentation__.Notify(623236)
			return v, nil
		},
	}, nil
}

func statColumnString(desc catalog.TableDescriptor, colID tree.Datum) string {
	__antithesis_instrumentation__.Notify(623285)
	id := descpb.ColumnID(*colID.(*tree.DInt))
	colDesc, err := desc.FindColumnWithID(id)
	if err != nil {
		__antithesis_instrumentation__.Notify(623287)

		return "<unknown>"
	} else {
		__antithesis_instrumentation__.Notify(623288)
	}
	__antithesis_instrumentation__.Notify(623286)
	return colDesc.GetName()
}
