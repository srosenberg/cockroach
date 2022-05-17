package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/docs"
	"github.com/cockroachdb/cockroach/pkg/geo/geoindex"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/paramparse"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgnotice"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type createIndexNode struct {
	n         *tree.CreateIndex
	tableDesc *tabledesc.Mutable
}

func (p *planner) CreateIndex(ctx context.Context, n *tree.CreateIndex) (planNode, error) {
	__antithesis_instrumentation__.Notify(462931)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"CREATE INDEX",
	); err != nil {
		__antithesis_instrumentation__.Notify(462938)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462939)
	}
	__antithesis_instrumentation__.Notify(462932)
	_, tableDesc, err := p.ResolveMutableTableDescriptor(
		ctx, &n.Table, true, tree.ResolveRequireTableOrViewDesc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(462940)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462941)
	}
	__antithesis_instrumentation__.Notify(462933)

	if tableDesc.IsView() && func() bool {
		__antithesis_instrumentation__.Notify(462942)
		return !tableDesc.MaterializedView() == true
	}() == true {
		__antithesis_instrumentation__.Notify(462943)
		return nil, pgerror.Newf(pgcode.WrongObjectType, "%q is not a table or materialized view", tableDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(462944)
	}
	__antithesis_instrumentation__.Notify(462934)

	if tableDesc.MaterializedView() {
		__antithesis_instrumentation__.Notify(462945)
		if n.Sharded != nil {
			__antithesis_instrumentation__.Notify(462946)
			return nil, pgerror.New(pgcode.InvalidObjectDefinition,
				"cannot create hash sharded index on materialized view")
		} else {
			__antithesis_instrumentation__.Notify(462947)
		}
	} else {
		__antithesis_instrumentation__.Notify(462948)
	}
	__antithesis_instrumentation__.Notify(462935)

	if err := p.CheckPrivilege(ctx, tableDesc, privilege.CREATE); err != nil {
		__antithesis_instrumentation__.Notify(462949)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462950)
	}
	__antithesis_instrumentation__.Notify(462936)

	if tableDesc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(462951)
		if err := p.checkNoRegionChangeUnderway(
			ctx,
			tableDesc.GetParentID(),
			"CREATE INDEX on a REGIONAL BY ROW table",
		); err != nil {
			__antithesis_instrumentation__.Notify(462952)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(462953)
		}
	} else {
		__antithesis_instrumentation__.Notify(462954)
	}
	__antithesis_instrumentation__.Notify(462937)

	return &createIndexNode{tableDesc: tableDesc, n: n}, nil
}

func (p *planner) maybeSetupConstraintForShard(
	ctx context.Context, tableDesc *tabledesc.Mutable, shardCol catalog.Column, buckets int32,
) error {
	__antithesis_instrumentation__.Notify(462955)

	version := p.ExecCfg().Settings.Version.ActiveVersion(ctx)
	if err := tableDesc.AllocateIDs(ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(462961)
		return err
	} else {
		__antithesis_instrumentation__.Notify(462962)
	}
	__antithesis_instrumentation__.Notify(462956)

	ckDef, err := makeShardCheckConstraintDef(int(buckets), shardCol)
	if err != nil {
		__antithesis_instrumentation__.Notify(462963)
		return err
	} else {
		__antithesis_instrumentation__.Notify(462964)
	}
	__antithesis_instrumentation__.Notify(462957)
	ckBuilder := schemaexpr.MakeCheckConstraintBuilder(ctx, p.tableName, tableDesc, &p.semaCtx)
	ckDesc, err := ckBuilder.Build(ckDef)
	if err != nil {
		__antithesis_instrumentation__.Notify(462965)
		return err
	} else {
		__antithesis_instrumentation__.Notify(462966)
	}
	__antithesis_instrumentation__.Notify(462958)

	curConstraintInfos, err := tableDesc.GetConstraintInfo()
	if err != nil {
		__antithesis_instrumentation__.Notify(462967)
		return err
	} else {
		__antithesis_instrumentation__.Notify(462968)
	}
	__antithesis_instrumentation__.Notify(462959)

	for _, info := range curConstraintInfos {
		__antithesis_instrumentation__.Notify(462969)
		if info.CheckConstraint != nil && func() bool {
			__antithesis_instrumentation__.Notify(462970)
			return info.CheckConstraint.Expr == ckDesc.Expr == true
		}() == true {
			__antithesis_instrumentation__.Notify(462971)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(462972)
		}
	}
	__antithesis_instrumentation__.Notify(462960)

	ckDesc.Validity = descpb.ConstraintValidity_Validating
	tableDesc.AddCheckMutation(ckDesc, descpb.DescriptorMutation_ADD)
	return nil
}

func makeIndexDescriptor(
	params runParams, n tree.CreateIndex, tableDesc *tabledesc.Mutable,
) (*descpb.IndexDescriptor, error) {
	__antithesis_instrumentation__.Notify(462973)
	if n.Sharded == nil && func() bool {
		__antithesis_instrumentation__.Notify(462988)
		return n.StorageParams.GetVal(`bucket_count`) != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(462989)
		return nil, pgerror.New(
			pgcode.InvalidParameterValue,
			`"bucket_count" storage param should only be set with "USING HASH" for hash sharded index`,
		)
	} else {
		__antithesis_instrumentation__.Notify(462990)
	}
	__antithesis_instrumentation__.Notify(462974)

	columns := make(tree.IndexElemList, len(n.Columns))
	copy(columns, n.Columns)

	if err := validateColumnsAreAccessible(tableDesc, columns); err != nil {
		__antithesis_instrumentation__.Notify(462991)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462992)
	}
	__antithesis_instrumentation__.Notify(462975)

	tn, err := params.p.getQualifiedTableName(params.ctx, tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(462993)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462994)
	}
	__antithesis_instrumentation__.Notify(462976)

	if err := replaceExpressionElemsWithVirtualCols(
		params.ctx,
		tableDesc,
		tn,
		columns,
		n.Inverted,
		false,
		params.p.SemaCtx(),
	); err != nil {
		__antithesis_instrumentation__.Notify(462995)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462996)
	}
	__antithesis_instrumentation__.Notify(462977)

	if err := validateIndexColumnsExist(tableDesc, columns); err != nil {
		__antithesis_instrumentation__.Notify(462997)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(462998)
	}
	__antithesis_instrumentation__.Notify(462978)

	if idx, _ := tableDesc.FindIndexWithName(string(n.Name)); idx != nil {
		__antithesis_instrumentation__.Notify(462999)
		if idx.Dropped() {
			__antithesis_instrumentation__.Notify(463001)
			return nil, pgerror.Newf(pgcode.DuplicateRelation, "index with name %q already exists and is being dropped, try again later", n.Name)
		} else {
			__antithesis_instrumentation__.Notify(463002)
		}
		__antithesis_instrumentation__.Notify(463000)
		return nil, pgerror.Newf(pgcode.DuplicateRelation, "index with name %q already exists", n.Name)
	} else {
		__antithesis_instrumentation__.Notify(463003)
	}
	__antithesis_instrumentation__.Notify(462979)

	indexDesc := descpb.IndexDescriptor{
		Name:              string(n.Name),
		Unique:            n.Unique,
		StoreColumnNames:  n.Storing.ToStrings(),
		CreatedExplicitly: true,
		CreatedAtNanos:    params.EvalContext().GetTxnTimestamp(time.Microsecond).UnixNano(),
	}

	if n.Inverted {
		__antithesis_instrumentation__.Notify(463004)
		if n.Sharded != nil {
			__antithesis_instrumentation__.Notify(463009)
			return nil, pgerror.New(pgcode.InvalidSQLStatementName, "inverted indexes don't support hash sharding")
		} else {
			__antithesis_instrumentation__.Notify(463010)
		}
		__antithesis_instrumentation__.Notify(463005)

		if len(indexDesc.StoreColumnNames) > 0 {
			__antithesis_instrumentation__.Notify(463011)
			return nil, pgerror.New(pgcode.InvalidSQLStatementName, "inverted indexes don't support stored columns")
		} else {
			__antithesis_instrumentation__.Notify(463012)
		}
		__antithesis_instrumentation__.Notify(463006)

		if n.Unique {
			__antithesis_instrumentation__.Notify(463013)
			return nil, pgerror.New(pgcode.InvalidSQLStatementName, "inverted indexes can't be unique")
		} else {
			__antithesis_instrumentation__.Notify(463014)
		}
		__antithesis_instrumentation__.Notify(463007)

		indexDesc.Type = descpb.IndexDescriptor_INVERTED
		column, err := tableDesc.FindColumnWithName(columns[len(columns)-1].Column)
		if err != nil {
			__antithesis_instrumentation__.Notify(463015)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463016)
		}
		__antithesis_instrumentation__.Notify(463008)
		switch column.GetType().Family() {
		case types.GeometryFamily:
			__antithesis_instrumentation__.Notify(463017)
			config, err := geoindex.GeometryIndexConfigForSRID(column.GetType().GeoSRIDOrZero())
			if err != nil {
				__antithesis_instrumentation__.Notify(463021)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(463022)
			}
			__antithesis_instrumentation__.Notify(463018)
			indexDesc.GeoConfig = *config
		case types.GeographyFamily:
			__antithesis_instrumentation__.Notify(463019)
			indexDesc.GeoConfig = *geoindex.DefaultGeographyIndexConfig()
		default:
			__antithesis_instrumentation__.Notify(463020)
		}
	} else {
		__antithesis_instrumentation__.Notify(463023)
	}
	__antithesis_instrumentation__.Notify(462980)

	if n.Sharded != nil {
		__antithesis_instrumentation__.Notify(463024)
		if n.PartitionByIndex.ContainsPartitions() {
			__antithesis_instrumentation__.Notify(463027)
			return nil, pgerror.New(pgcode.FeatureNotSupported, "sharded indexes don't support explicit partitioning")
		} else {
			__antithesis_instrumentation__.Notify(463028)
		}
		__antithesis_instrumentation__.Notify(463025)

		shardCol, newColumns, err := setupShardedIndex(
			params.ctx,
			params.EvalContext(),
			&params.p.semaCtx,
			columns,
			n.Sharded.ShardBuckets,
			tableDesc,
			&indexDesc,
			n.StorageParams,
			false)
		if err != nil {
			__antithesis_instrumentation__.Notify(463029)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463030)
		}
		__antithesis_instrumentation__.Notify(463026)
		columns = newColumns
		if err := params.p.maybeSetupConstraintForShard(
			params.ctx, tableDesc, shardCol, indexDesc.Sharded.ShardBuckets,
		); err != nil {
			__antithesis_instrumentation__.Notify(463031)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463032)
		}
	} else {
		__antithesis_instrumentation__.Notify(463033)
	}
	__antithesis_instrumentation__.Notify(462981)

	if n.Predicate != nil {
		__antithesis_instrumentation__.Notify(463034)
		expr, err := schemaexpr.ValidatePartialIndexPredicate(
			params.ctx, tableDesc, n.Predicate, &n.Table, params.p.SemaCtx(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(463036)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(463037)
		}
		__antithesis_instrumentation__.Notify(463035)
		indexDesc.Predicate = expr
	} else {
		__antithesis_instrumentation__.Notify(463038)
	}
	__antithesis_instrumentation__.Notify(462982)

	if err := indexDesc.FillColumns(columns); err != nil {
		__antithesis_instrumentation__.Notify(463039)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463040)
	}
	__antithesis_instrumentation__.Notify(462983)

	if err := paramparse.SetStorageParameters(
		params.ctx,
		params.p.SemaCtx(),
		params.EvalContext(),
		n.StorageParams,
		&paramparse.IndexStorageParamObserver{IndexDesc: &indexDesc},
	); err != nil {
		__antithesis_instrumentation__.Notify(463041)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463042)
	}
	__antithesis_instrumentation__.Notify(462984)

	if indexDesc.Type == descpb.IndexDescriptor_INVERTED {
		__antithesis_instrumentation__.Notify(463043)
		telemetry.Inc(sqltelemetry.InvertedIndexCounter)
		if geoindex.IsGeometryConfig(&indexDesc.GeoConfig) {
			__antithesis_instrumentation__.Notify(463047)
			telemetry.Inc(sqltelemetry.GeometryInvertedIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(463048)
		}
		__antithesis_instrumentation__.Notify(463044)
		if geoindex.IsGeographyConfig(&indexDesc.GeoConfig) {
			__antithesis_instrumentation__.Notify(463049)
			telemetry.Inc(sqltelemetry.GeographyInvertedIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(463050)
		}
		__antithesis_instrumentation__.Notify(463045)
		if indexDesc.IsPartial() {
			__antithesis_instrumentation__.Notify(463051)
			telemetry.Inc(sqltelemetry.PartialInvertedIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(463052)
		}
		__antithesis_instrumentation__.Notify(463046)
		if len(indexDesc.KeyColumnNames) > 1 {
			__antithesis_instrumentation__.Notify(463053)
			telemetry.Inc(sqltelemetry.MultiColumnInvertedIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(463054)
		}
	} else {
		__antithesis_instrumentation__.Notify(463055)
	}
	__antithesis_instrumentation__.Notify(462985)
	if indexDesc.IsSharded() {
		__antithesis_instrumentation__.Notify(463056)
		telemetry.Inc(sqltelemetry.HashShardedIndexCounter)
	} else {
		__antithesis_instrumentation__.Notify(463057)
	}
	__antithesis_instrumentation__.Notify(462986)
	if indexDesc.IsPartial() {
		__antithesis_instrumentation__.Notify(463058)
		telemetry.Inc(sqltelemetry.PartialIndexCounter)
	} else {
		__antithesis_instrumentation__.Notify(463059)
	}
	__antithesis_instrumentation__.Notify(462987)

	return &indexDesc, nil
}

func validateColumnsAreAccessible(desc *tabledesc.Mutable, columns tree.IndexElemList) error {
	__antithesis_instrumentation__.Notify(463060)
	for _, column := range columns {
		__antithesis_instrumentation__.Notify(463062)

		if column.Expr != nil {
			__antithesis_instrumentation__.Notify(463065)
			continue
		} else {
			__antithesis_instrumentation__.Notify(463066)
		}
		__antithesis_instrumentation__.Notify(463063)
		foundColumn, err := desc.FindColumnWithName(column.Column)
		if err != nil {
			__antithesis_instrumentation__.Notify(463067)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463068)
		}
		__antithesis_instrumentation__.Notify(463064)
		if foundColumn.IsInaccessible() {
			__antithesis_instrumentation__.Notify(463069)
			return pgerror.Newf(
				pgcode.UndefinedColumn,
				"column %q is inaccessible and cannot be referenced",
				foundColumn.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(463070)
		}
	}
	__antithesis_instrumentation__.Notify(463061)
	return nil
}

func validateIndexColumnsExist(desc *tabledesc.Mutable, columns tree.IndexElemList) error {
	__antithesis_instrumentation__.Notify(463071)
	for _, column := range columns {
		__antithesis_instrumentation__.Notify(463073)
		if column.Expr != nil {
			__antithesis_instrumentation__.Notify(463076)
			return errors.AssertionFailedf("index elem expression should have been replaced with a column")
		} else {
			__antithesis_instrumentation__.Notify(463077)
		}
		__antithesis_instrumentation__.Notify(463074)
		foundColumn, err := desc.FindColumnWithName(column.Column)
		if err != nil {
			__antithesis_instrumentation__.Notify(463078)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463079)
		}
		__antithesis_instrumentation__.Notify(463075)
		if foundColumn.Dropped() {
			__antithesis_instrumentation__.Notify(463080)
			return colinfo.NewUndefinedColumnError(string(column.Column))
		} else {
			__antithesis_instrumentation__.Notify(463081)
		}
	}
	__antithesis_instrumentation__.Notify(463072)
	return nil
}

func replaceExpressionElemsWithVirtualCols(
	ctx context.Context,
	desc *tabledesc.Mutable,
	tn *tree.TableName,
	elems tree.IndexElemList,
	isInverted bool,
	isNewTable bool,
	semaCtx *tree.SemaContext,
) error {
	__antithesis_instrumentation__.Notify(463082)
	findExistingExprIndexCol := func(expr string) (colName string, ok bool) {
		__antithesis_instrumentation__.Notify(463085)
		for _, col := range desc.AllColumns() {
			__antithesis_instrumentation__.Notify(463087)
			if col.IsExpressionIndexColumn() && func() bool {
				__antithesis_instrumentation__.Notify(463088)
				return col.GetComputeExpr() == expr == true
			}() == true {
				__antithesis_instrumentation__.Notify(463089)
				return col.GetName(), true
			} else {
				__antithesis_instrumentation__.Notify(463090)
			}
		}
		__antithesis_instrumentation__.Notify(463086)
		return "", false
	}
	__antithesis_instrumentation__.Notify(463083)

	lastColumnIdx := len(elems) - 1
	for i := range elems {
		__antithesis_instrumentation__.Notify(463091)
		elem := &elems[i]
		if elem.Expr != nil {
			__antithesis_instrumentation__.Notify(463092)

			colDef := &tree.ColumnTableDef{
				Type: types.Any,
			}
			colDef.Computed.Computed = true
			colDef.Computed.Expr = elem.Expr
			colDef.Computed.Virtual = true

			expr, typ, err := schemaexpr.ValidateComputedColumnExpression(
				ctx,
				desc,
				colDef,
				tn,
				"index element",
				semaCtx,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(463100)
				return err
			} else {
				__antithesis_instrumentation__.Notify(463101)
			}
			__antithesis_instrumentation__.Notify(463093)

			if existingColName, ok := findExistingExprIndexCol(expr); ok {
				__antithesis_instrumentation__.Notify(463102)

				elem.Column = tree.Name(existingColName)
				elem.Expr = nil

				telemetry.Inc(sqltelemetry.ExpressionIndexCounter)
				continue
			} else {
				__antithesis_instrumentation__.Notify(463103)
			}
			__antithesis_instrumentation__.Notify(463094)

			if typ.IsAmbiguous() {
				__antithesis_instrumentation__.Notify(463104)
				return errors.WithHint(
					pgerror.Newf(
						pgcode.InvalidTableDefinition,
						"type of index element %s is ambiguous",
						elem.Expr.String(),
					),
					"consider adding a type cast to the expression",
				)
			} else {
				__antithesis_instrumentation__.Notify(463105)
			}
			__antithesis_instrumentation__.Notify(463095)

			if !isInverted && func() bool {
				__antithesis_instrumentation__.Notify(463106)
				return !colinfo.ColumnTypeIsIndexable(typ) == true
			}() == true {
				__antithesis_instrumentation__.Notify(463107)
				if colinfo.ColumnTypeIsInvertedIndexable(typ) {
					__antithesis_instrumentation__.Notify(463109)
					return errors.WithHint(
						pgerror.Newf(
							pgcode.InvalidTableDefinition,
							"index element %s of type %s is not indexable in a non-inverted index",
							elem.Expr.String(),
							typ.Name(),
						),
						"you may want to create an inverted index instead. See the documentation for inverted indexes: "+docs.URL("inverted-indexes.html"),
					)
				} else {
					__antithesis_instrumentation__.Notify(463110)
				}
				__antithesis_instrumentation__.Notify(463108)
				return pgerror.Newf(
					pgcode.InvalidTableDefinition,
					"index element %s of type %s is not indexable",
					elem.Expr.String(),
					typ.Name(),
				)
			} else {
				__antithesis_instrumentation__.Notify(463111)
			}
			__antithesis_instrumentation__.Notify(463096)

			if isInverted {
				__antithesis_instrumentation__.Notify(463112)
				if i < lastColumnIdx && func() bool {
					__antithesis_instrumentation__.Notify(463114)
					return !colinfo.ColumnTypeIsIndexable(typ) == true
				}() == true {
					__antithesis_instrumentation__.Notify(463115)
					return errors.WithHint(
						pgerror.Newf(
							pgcode.InvalidTableDefinition,
							"index element %s of type %s is not allowed as a prefix column in an inverted index",
							elem.Expr.String(),
							typ.Name(),
						),
						"see the documentation for more information about inverted indexes: "+docs.URL("inverted-indexes.html"),
					)
				} else {
					__antithesis_instrumentation__.Notify(463116)
				}
				__antithesis_instrumentation__.Notify(463113)
				if i == lastColumnIdx && func() bool {
					__antithesis_instrumentation__.Notify(463117)
					return !colinfo.ColumnTypeIsInvertedIndexable(typ) == true
				}() == true {
					__antithesis_instrumentation__.Notify(463118)
					return errors.WithHint(
						pgerror.Newf(
							pgcode.InvalidTableDefinition,
							"index element %s of type %s is not allowed as the last column in an inverted index",
							elem.Expr.String(),
							typ.Name(),
						),
						"see the documentation for more information about inverted indexes: "+docs.URL("inverted-indexes.html"),
					)
				} else {
					__antithesis_instrumentation__.Notify(463119)
				}
			} else {
				__antithesis_instrumentation__.Notify(463120)
			}
			__antithesis_instrumentation__.Notify(463097)

			colName := tabledesc.GenerateUniqueName("crdb_internal_idx_expr", func(name string) bool {
				__antithesis_instrumentation__.Notify(463121)
				_, err := desc.FindColumnWithName(tree.Name(name))
				return err == nil
			})
			__antithesis_instrumentation__.Notify(463098)
			col := &descpb.ColumnDescriptor{
				Name:         colName,
				Inaccessible: true,
				Type:         typ,
				ComputeExpr:  &expr,
				Virtual:      true,
				Nullable:     true,
			}

			if isNewTable {
				__antithesis_instrumentation__.Notify(463122)
				desc.AddColumn(col)
			} else {
				__antithesis_instrumentation__.Notify(463123)
				desc.AddColumnMutation(col, descpb.DescriptorMutation_ADD)
			}
			__antithesis_instrumentation__.Notify(463099)

			elem.Column = tree.Name(colName)
			elem.Expr = nil

			telemetry.Inc(sqltelemetry.ExpressionIndexCounter)
		} else {
			__antithesis_instrumentation__.Notify(463124)
		}
	}
	__antithesis_instrumentation__.Notify(463084)

	return nil
}

func (n *createIndexNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(463125) }

func setupShardedIndex(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	columns tree.IndexElemList,
	bucketsExpr tree.Expr,
	tableDesc *tabledesc.Mutable,
	indexDesc *descpb.IndexDescriptor,
	storageParams tree.StorageParams,
	isNewTable bool,
) (shard catalog.Column, newColumns tree.IndexElemList, err error) {
	__antithesis_instrumentation__.Notify(463126)
	if !isNewTable && func() bool {
		__antithesis_instrumentation__.Notify(463131)
		return tableDesc.IsPartitionAllBy() == true
	}() == true {
		__antithesis_instrumentation__.Notify(463132)
		partitionAllBy, err := partitionByFromTableDesc(evalCtx.Codec, tableDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(463134)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(463135)
		}
		__antithesis_instrumentation__.Notify(463133)
		if anyColumnIsPartitioningField(columns, partitionAllBy) {
			__antithesis_instrumentation__.Notify(463136)
			return nil, nil, pgerror.New(
				pgcode.FeatureNotSupported,
				`hash sharded indexes cannot include implicit partitioning columns from "PARTITION ALL BY" or "LOCALITY REGIONAL BY ROW"`,
			)
		} else {
			__antithesis_instrumentation__.Notify(463137)
		}
	} else {
		__antithesis_instrumentation__.Notify(463138)
	}
	__antithesis_instrumentation__.Notify(463127)

	colNames := make([]string, 0, len(columns))
	for _, c := range columns {
		__antithesis_instrumentation__.Notify(463139)
		colNames = append(colNames, string(c.Column))
	}
	__antithesis_instrumentation__.Notify(463128)
	buckets, err := tabledesc.EvalShardBucketCount(ctx, semaCtx, evalCtx, bucketsExpr, storageParams)
	if err != nil {
		__antithesis_instrumentation__.Notify(463140)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463141)
	}
	__antithesis_instrumentation__.Notify(463129)
	shardCol, err := maybeCreateAndAddShardCol(int(buckets), tableDesc,
		colNames, isNewTable)

	if err != nil {
		__antithesis_instrumentation__.Notify(463142)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(463143)
	}
	__antithesis_instrumentation__.Notify(463130)
	shardIdxElem := tree.IndexElem{
		Column:    tree.Name(shardCol.GetName()),
		Direction: tree.Ascending,
	}
	newColumns = append(tree.IndexElemList{shardIdxElem}, columns...)
	indexDesc.Sharded = catpb.ShardedDescriptor{
		IsSharded:    true,
		Name:         shardCol.GetName(),
		ShardBuckets: buckets,
		ColumnNames:  colNames,
	}
	return shardCol, newColumns, nil
}

func maybeCreateAndAddShardCol(
	shardBuckets int, desc *tabledesc.Mutable, colNames []string, isNewTable bool,
) (col catalog.Column, err error) {
	__antithesis_instrumentation__.Notify(463144)
	shardColDesc, err := makeShardColumnDesc(colNames, shardBuckets)
	if err != nil {
		__antithesis_instrumentation__.Notify(463149)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463150)
	}
	__antithesis_instrumentation__.Notify(463145)
	existingShardCol, err := desc.FindColumnWithName(tree.Name(shardColDesc.Name))
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(463151)
		return !existingShardCol.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(463152)

		if !existingShardCol.IsHidden() {
			__antithesis_instrumentation__.Notify(463154)

			return nil, pgerror.Newf(pgcode.DuplicateColumn,
				"column %s already specified; can't be used for sharding", shardColDesc.Name)
		} else {
			__antithesis_instrumentation__.Notify(463155)
		}
		__antithesis_instrumentation__.Notify(463153)
		return existingShardCol, nil
	} else {
		__antithesis_instrumentation__.Notify(463156)
	}
	__antithesis_instrumentation__.Notify(463146)
	columnIsUndefined := sqlerrors.IsUndefinedColumnError(err)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(463157)
		return !columnIsUndefined == true
	}() == true {
		__antithesis_instrumentation__.Notify(463158)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(463159)
	}
	__antithesis_instrumentation__.Notify(463147)
	if columnIsUndefined || func() bool {
		__antithesis_instrumentation__.Notify(463160)
		return existingShardCol.Dropped() == true
	}() == true {
		__antithesis_instrumentation__.Notify(463161)
		if isNewTable {
			__antithesis_instrumentation__.Notify(463162)
			desc.AddColumn(shardColDesc)
		} else {
			__antithesis_instrumentation__.Notify(463163)
			desc.AddColumnMutation(shardColDesc, descpb.DescriptorMutation_ADD)
		}
	} else {
		__antithesis_instrumentation__.Notify(463164)
	}
	__antithesis_instrumentation__.Notify(463148)
	shardCol, err := desc.FindColumnWithName(tree.Name(shardColDesc.Name))
	return shardCol, err
}

func (n *createIndexNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(463165)
	telemetry.Inc(sqltelemetry.SchemaChangeCreateCounter("index"))
	foundIndex, err := n.tableDesc.FindIndexWithName(string(n.n.Name))
	if err == nil {
		__antithesis_instrumentation__.Notify(463179)
		if foundIndex.Dropped() {
			__antithesis_instrumentation__.Notify(463181)
			return pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
				"index %q being dropped, try again later", string(n.n.Name))
		} else {
			__antithesis_instrumentation__.Notify(463182)
		}
		__antithesis_instrumentation__.Notify(463180)
		if n.n.IfNotExists {
			__antithesis_instrumentation__.Notify(463183)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(463184)
		}
	} else {
		__antithesis_instrumentation__.Notify(463185)
	}
	__antithesis_instrumentation__.Notify(463166)

	if n.n.Concurrently {
		__antithesis_instrumentation__.Notify(463186)
		params.p.BufferClientNotice(
			params.ctx,
			pgnotice.Newf("CONCURRENTLY is not required as all indexes are created concurrently"),
		)
	} else {
		__antithesis_instrumentation__.Notify(463187)
	}
	__antithesis_instrumentation__.Notify(463167)

	if n.n.PartitionByIndex == nil && func() bool {
		__antithesis_instrumentation__.Notify(463188)
		return n.tableDesc.GetPrimaryIndex().GetPartitioning().NumColumns() > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(463189)
		return !n.tableDesc.IsPartitionAllBy() == true
	}() == true {
		__antithesis_instrumentation__.Notify(463190)
		params.p.BufferClientNotice(
			params.ctx,
			errors.WithHint(
				pgnotice.Newf("creating non-partitioned index on partitioned table may not be performant"),
				"Consider modifying the index such that it is also partitioned.",
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(463191)
	}
	__antithesis_instrumentation__.Notify(463168)

	indexDesc, err := makeIndexDescriptor(params, *n.n, n.tableDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(463192)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463193)
	}
	__antithesis_instrumentation__.Notify(463169)

	if len(indexDesc.StoreColumnNames) > 1 && func() bool {
		__antithesis_instrumentation__.Notify(463194)
		return len(n.tableDesc.Families) > 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(463195)
		telemetry.Inc(sqltelemetry.SecondaryIndexColumnFamiliesCounter)
	} else {
		__antithesis_instrumentation__.Notify(463196)
	}
	__antithesis_instrumentation__.Notify(463170)

	indexDesc.Version = descpb.StrictIndexColumnIDGuaranteesVersion

	if n.n.PartitionByIndex != nil && func() bool {
		__antithesis_instrumentation__.Notify(463197)
		return n.tableDesc.GetLocalityConfig() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(463198)
		return pgerror.New(
			pgcode.FeatureNotSupported,
			"cannot define PARTITION BY on a new INDEX in a multi-region database",
		)
	} else {
		__antithesis_instrumentation__.Notify(463199)
	}
	__antithesis_instrumentation__.Notify(463171)

	*indexDesc, err = params.p.configureIndexDescForNewIndexPartitioning(
		params.ctx,
		n.tableDesc,
		*indexDesc,
		n.n.PartitionByIndex,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(463200)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463201)
	}
	__antithesis_instrumentation__.Notify(463172)

	if indexDesc.Type == descpb.IndexDescriptor_INVERTED && func() bool {
		__antithesis_instrumentation__.Notify(463202)
		return indexDesc.Partitioning.NumColumns != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(463203)
		telemetry.Inc(sqltelemetry.PartitionedInvertedIndexCounter)
	} else {
		__antithesis_instrumentation__.Notify(463204)
	}
	__antithesis_instrumentation__.Notify(463173)

	mutationIdx := len(n.tableDesc.Mutations)
	if err := n.tableDesc.AddIndexMutation(params.ctx, indexDesc, descpb.DescriptorMutation_ADD, params.p.ExecCfg().Settings); err != nil {
		__antithesis_instrumentation__.Notify(463205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463206)
	}
	__antithesis_instrumentation__.Notify(463174)
	version := params.ExecCfg().Settings.Version.ActiveVersion(params.ctx)
	if err := n.tableDesc.AllocateIDs(params.ctx, version); err != nil {
		__antithesis_instrumentation__.Notify(463207)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463208)
	}
	__antithesis_instrumentation__.Notify(463175)

	if err := params.p.configureZoneConfigForNewIndexPartitioning(
		params.ctx,
		n.tableDesc,
		*indexDesc,
	); err != nil {
		__antithesis_instrumentation__.Notify(463209)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463210)
	}
	__antithesis_instrumentation__.Notify(463176)

	index := n.tableDesc.Mutations[mutationIdx].GetIndex()
	indexName := index.Name

	mutationID := n.tableDesc.ClusterVersion().NextMutationID
	if err := params.p.writeSchemaChange(
		params.ctx, n.tableDesc, mutationID, tree.AsStringWithFQNames(n.n, params.Ann()),
	); err != nil {
		__antithesis_instrumentation__.Notify(463211)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463212)
	}
	__antithesis_instrumentation__.Notify(463177)

	if err := params.p.addBackRefsFromAllTypesInTable(params.ctx, n.tableDesc); err != nil {
		__antithesis_instrumentation__.Notify(463213)
		return err
	} else {
		__antithesis_instrumentation__.Notify(463214)
	}
	__antithesis_instrumentation__.Notify(463178)

	return params.p.logEvent(params.ctx,
		n.tableDesc.ID,
		&eventpb.CreateIndex{
			TableName:  n.n.Table.FQString(),
			IndexName:  indexName,
			MutationID: uint32(mutationID),
		})
}

func (*createIndexNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(463215)
	return false, nil
}
func (*createIndexNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(463216)
	return tree.Datums{}
}
func (*createIndexNode) Close(context.Context) { __antithesis_instrumentation__.Notify(463217) }

func (p *planner) configureIndexDescForNewIndexPartitioning(
	ctx context.Context,
	tableDesc *tabledesc.Mutable,
	indexDesc descpb.IndexDescriptor,
	partitionByIndex *tree.PartitionByIndex,
) (descpb.IndexDescriptor, error) {
	__antithesis_instrumentation__.Notify(463218)
	var err error
	if partitionByIndex.ContainsPartitioningClause() || func() bool {
		__antithesis_instrumentation__.Notify(463220)
		return tableDesc.IsPartitionAllBy() == true
	}() == true {
		__antithesis_instrumentation__.Notify(463221)
		var partitionBy *tree.PartitionBy
		if !tableDesc.IsPartitionAllBy() {
			__antithesis_instrumentation__.Notify(463223)
			if partitionByIndex.ContainsPartitions() {
				__antithesis_instrumentation__.Notify(463224)
				partitionBy = partitionByIndex.PartitionBy
			} else {
				__antithesis_instrumentation__.Notify(463225)
			}
		} else {
			__antithesis_instrumentation__.Notify(463226)
			if partitionByIndex.ContainsPartitioningClause() {
				__antithesis_instrumentation__.Notify(463227)
				return indexDesc, pgerror.New(
					pgcode.FeatureNotSupported,
					"cannot define PARTITION BY on an index if the table has a PARTITION ALL BY definition",
				)
			} else {
				__antithesis_instrumentation__.Notify(463228)
				partitionBy, err = partitionByFromTableDesc(p.ExecCfg().Codec, tableDesc)
				if err != nil {
					__antithesis_instrumentation__.Notify(463229)
					return indexDesc, err
				} else {
					__antithesis_instrumentation__.Notify(463230)
				}
			}
		}
		__antithesis_instrumentation__.Notify(463222)
		allowImplicitPartitioning := p.EvalContext().SessionData().ImplicitColumnPartitioningEnabled || func() bool {
			__antithesis_instrumentation__.Notify(463231)
			return tableDesc.IsLocalityRegionalByRow() == true
		}() == true
		if partitionBy != nil {
			__antithesis_instrumentation__.Notify(463232)
			newImplicitCols, newPartitioning, err := CreatePartitioning(
				ctx,
				p.ExecCfg().Settings,
				p.EvalContext(),
				tableDesc,
				indexDesc,
				partitionBy,
				nil,
				allowImplicitPartitioning,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(463234)
				return indexDesc, err
			} else {
				__antithesis_instrumentation__.Notify(463235)
			}
			__antithesis_instrumentation__.Notify(463233)
			tabledesc.UpdateIndexPartitioning(&indexDesc, false, newImplicitCols, newPartitioning)
		} else {
			__antithesis_instrumentation__.Notify(463236)
		}
	} else {
		__antithesis_instrumentation__.Notify(463237)
	}
	__antithesis_instrumentation__.Notify(463219)
	return indexDesc, nil
}

func (p *planner) configureZoneConfigForNewIndexPartitioning(
	ctx context.Context, tableDesc *tabledesc.Mutable, indexDesc descpb.IndexDescriptor,
) error {
	__antithesis_instrumentation__.Notify(463238)
	if indexDesc.ID == 0 {
		__antithesis_instrumentation__.Notify(463241)
		return errors.AssertionFailedf("index %s does not have id", indexDesc.Name)
	} else {
		__antithesis_instrumentation__.Notify(463242)
	}
	__antithesis_instrumentation__.Notify(463239)

	if tableDesc.IsLocalityRegionalByRow() {
		__antithesis_instrumentation__.Notify(463243)
		regionConfig, err := SynthesizeRegionConfig(ctx, p.txn, tableDesc.GetParentID(), p.Descriptors())
		if err != nil {
			__antithesis_instrumentation__.Notify(463246)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463247)
		}
		__antithesis_instrumentation__.Notify(463244)

		indexIDs := []descpb.IndexID{indexDesc.ID}
		if idx := catalog.FindCorrespondingTemporaryIndexByID(tableDesc, indexDesc.ID); idx != nil {
			__antithesis_instrumentation__.Notify(463248)
			indexIDs = append(indexIDs, idx.GetID())
		} else {
			__antithesis_instrumentation__.Notify(463249)
		}
		__antithesis_instrumentation__.Notify(463245)

		if err := ApplyZoneConfigForMultiRegionTable(
			ctx,
			p.txn,
			p.ExecCfg(),
			regionConfig,
			tableDesc,
			applyZoneConfigForMultiRegionTableOptionNewIndexes(indexIDs...),
		); err != nil {
			__antithesis_instrumentation__.Notify(463250)
			return err
		} else {
			__antithesis_instrumentation__.Notify(463251)
		}
	} else {
		__antithesis_instrumentation__.Notify(463252)
	}
	__antithesis_instrumentation__.Notify(463240)
	return nil
}

func anyColumnIsPartitioningField(columns tree.IndexElemList, partitionBy *tree.PartitionBy) bool {
	__antithesis_instrumentation__.Notify(463253)
	for _, field := range partitionBy.Fields {
		__antithesis_instrumentation__.Notify(463255)
		for _, column := range columns {
			__antithesis_instrumentation__.Notify(463256)
			if field == column.Column {
				__antithesis_instrumentation__.Notify(463257)
				return true
			} else {
				__antithesis_instrumentation__.Notify(463258)
			}
		}
	}
	__antithesis_instrumentation__.Notify(463254)
	return false
}
