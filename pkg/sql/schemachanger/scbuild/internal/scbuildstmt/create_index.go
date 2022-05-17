package scbuildstmt

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/catid"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func CreateIndex(b BuildCtx, n *tree.CreateIndex) {
	__antithesis_instrumentation__.Notify(579848)

	relationElements := b.ResolveRelation(n.Table.ToUnresolvedObjectName(), ResolveParams{
		IsExistenceOptional: false,
		RequiredPrivilege:   privilege.CREATE,
	})
	index := scpb.Index{
		IsUnique:       n.Unique,
		IsInverted:     n.Inverted,
		IsConcurrently: n.Concurrently,
	}
	var relation scpb.Element
	var source *scpb.PrimaryIndex
	relationElements.ForEachElementStatus(func(_ scpb.Status, target scpb.TargetStatus, e scpb.Element) {
		__antithesis_instrumentation__.Notify(579858)
		switch t := e.(type) {
		case *scpb.Table:
			__antithesis_instrumentation__.Notify(579859)
			n.Table.ObjectNamePrefix = b.NamePrefix(t)
			if descpb.IsVirtualTable(t.TableID) {
				__antithesis_instrumentation__.Notify(579868)
				return
			} else {
				__antithesis_instrumentation__.Notify(579869)
			}
			__antithesis_instrumentation__.Notify(579860)
			index.TableID = t.TableID
			relation = e

		case *scpb.View:
			__antithesis_instrumentation__.Notify(579861)
			n.Table.ObjectNamePrefix = b.NamePrefix(t)
			if !t.IsMaterialized {
				__antithesis_instrumentation__.Notify(579870)
				return
			} else {
				__antithesis_instrumentation__.Notify(579871)
			}
			__antithesis_instrumentation__.Notify(579862)
			if n.Sharded != nil {
				__antithesis_instrumentation__.Notify(579872)
				panic(pgerror.New(pgcode.InvalidObjectDefinition,
					"cannot create hash sharded index on materialized view"))
			} else {
				__antithesis_instrumentation__.Notify(579873)
			}
			__antithesis_instrumentation__.Notify(579863)
			index.TableID = t.ViewID
			relation = e

		case *scpb.TableLocalityGlobal, *scpb.TableLocalityPrimaryRegion, *scpb.TableLocalitySecondaryRegion:
			__antithesis_instrumentation__.Notify(579864)
			if n.PartitionByIndex != nil {
				__antithesis_instrumentation__.Notify(579874)
				panic(pgerror.New(pgcode.FeatureNotSupported,
					"cannot define PARTITION BY on a new INDEX in a multi-region database",
				))
			} else {
				__antithesis_instrumentation__.Notify(579875)
			}

		case *scpb.TableLocalityRegionalByRow:
			__antithesis_instrumentation__.Notify(579865)
			if n.PartitionByIndex != nil {
				__antithesis_instrumentation__.Notify(579876)
				panic(pgerror.New(pgcode.FeatureNotSupported,
					"cannot define PARTITION BY on a new INDEX in a multi-region database",
				))
			} else {
				__antithesis_instrumentation__.Notify(579877)
			}
			__antithesis_instrumentation__.Notify(579866)
			if n.Sharded != nil {
				__antithesis_instrumentation__.Notify(579878)
				panic(pgerror.New(pgcode.FeatureNotSupported, "hash sharded indexes are not compatible with REGIONAL BY ROW tables"))
			} else {
				__antithesis_instrumentation__.Notify(579879)
			}

		case *scpb.PrimaryIndex:
			__antithesis_instrumentation__.Notify(579867)
			if target == scpb.ToPublic {
				__antithesis_instrumentation__.Notify(579880)
				source = t
			} else {
				__antithesis_instrumentation__.Notify(579881)
			}
		}
	})
	__antithesis_instrumentation__.Notify(579849)
	if index.TableID == catid.InvalidDescID || func() bool {
		__antithesis_instrumentation__.Notify(579882)
		return source == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(579883)
		panic(pgerror.Newf(pgcode.WrongObjectType,
			"%q is not an indexable table or a materialized view", n.Table.ObjectName))
	} else {
		__antithesis_instrumentation__.Notify(579884)
	}

	{
		__antithesis_instrumentation__.Notify(579885)
		indexElements := b.ResolveIndex(index.TableID, n.Name, ResolveParams{
			IsExistenceOptional: true,
			RequiredPrivilege:   privilege.CREATE,
		})
		if _, target, sec := scpb.FindSecondaryIndex(indexElements); sec != nil {
			__antithesis_instrumentation__.Notify(579886)
			if n.IfNotExists {
				__antithesis_instrumentation__.Notify(579889)
				return
			} else {
				__antithesis_instrumentation__.Notify(579890)
			}
			__antithesis_instrumentation__.Notify(579887)
			if target == scpb.ToAbsent {
				__antithesis_instrumentation__.Notify(579891)
				panic(pgerror.Newf(pgcode.ObjectNotInPrerequisiteState,
					"index %q being dropped, try again later", n.Name.String()))
			} else {
				__antithesis_instrumentation__.Notify(579892)
			}
			__antithesis_instrumentation__.Notify(579888)
			panic(pgerror.Newf(pgcode.DuplicateRelation, "index with name %q already exists", n.Name))
		} else {
			__antithesis_instrumentation__.Notify(579893)
		}
	}
	__antithesis_instrumentation__.Notify(579850)

	columnRefs := map[string]struct{}{}
	for _, columnNode := range n.Columns {
		__antithesis_instrumentation__.Notify(579894)
		colName := columnNode.Column.String()
		if _, found := columnRefs[colName]; found {
			__antithesis_instrumentation__.Notify(579896)
			panic(pgerror.Newf(pgcode.InvalidObjectDefinition,
				"index %q contains duplicate column %q", n.Name, colName))
		} else {
			__antithesis_instrumentation__.Notify(579897)
		}
		__antithesis_instrumentation__.Notify(579895)
		columnRefs[colName] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(579851)
	for _, storingNode := range n.Storing {
		__antithesis_instrumentation__.Notify(579898)
		colName := storingNode.String()
		if _, found := columnRefs[colName]; found {
			__antithesis_instrumentation__.Notify(579900)
			panic(pgerror.Newf(pgcode.InvalidObjectDefinition,
				"index %q contains duplicate column %q", n.Name, colName))
		} else {
			__antithesis_instrumentation__.Notify(579901)
		}
		__antithesis_instrumentation__.Notify(579899)
		columnRefs[colName] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(579852)

	keyColNames := make([]string, len(n.Columns))
	for i, columnNode := range n.Columns {
		__antithesis_instrumentation__.Notify(579902)
		colName := columnNode.Column.String()
		if columnNode.Expr != nil {
			__antithesis_instrumentation__.Notify(579907)
			tbl, ok := relation.(*scpb.Table)
			if !ok {
				__antithesis_instrumentation__.Notify(579909)
				panic(scerrors.NotImplementedErrorf(n,
					"indexing virtual column expressions in materialized views is not supported"))
			} else {
				__antithesis_instrumentation__.Notify(579910)
			}
			__antithesis_instrumentation__.Notify(579908)
			colName = createVirtualColumnForIndex(b, &n.Table, tbl, columnNode.Expr)
			relationElements = b.QueryByID(index.TableID)
		} else {
			__antithesis_instrumentation__.Notify(579911)
		}
		__antithesis_instrumentation__.Notify(579903)
		var columnID catid.ColumnID
		scpb.ForEachColumnName(relationElements, func(_ scpb.Status, target scpb.TargetStatus, e *scpb.ColumnName) {
			__antithesis_instrumentation__.Notify(579912)
			if target == scpb.ToPublic && func() bool {
				__antithesis_instrumentation__.Notify(579913)
				return e.Name == colName == true
			}() == true {
				__antithesis_instrumentation__.Notify(579914)
				columnID = e.ColumnID
			} else {
				__antithesis_instrumentation__.Notify(579915)
			}
		})
		__antithesis_instrumentation__.Notify(579904)
		if columnID == 0 {
			__antithesis_instrumentation__.Notify(579916)
			panic(colinfo.NewUndefinedColumnError(colName))
		} else {
			__antithesis_instrumentation__.Notify(579917)
		}
		__antithesis_instrumentation__.Notify(579905)
		keyColNames[i] = colName
		index.KeyColumnIDs = append(index.KeyColumnIDs, columnID)
		direction := scpb.Index_ASC
		if columnNode.Direction == tree.Descending {
			__antithesis_instrumentation__.Notify(579918)
			direction = scpb.Index_DESC
		} else {
			__antithesis_instrumentation__.Notify(579919)
		}
		__antithesis_instrumentation__.Notify(579906)
		index.KeyColumnDirections = append(index.KeyColumnDirections, direction)
	}
	__antithesis_instrumentation__.Notify(579853)

	keyColIDs := catalog.MakeTableColSet(index.KeyColumnIDs...)
	for _, id := range source.KeyColumnIDs {
		__antithesis_instrumentation__.Notify(579920)
		if !keyColIDs.Contains(id) {
			__antithesis_instrumentation__.Notify(579921)
			index.KeySuffixColumnIDs = append(index.KeySuffixColumnIDs, id)
		} else {
			__antithesis_instrumentation__.Notify(579922)
		}
	}
	__antithesis_instrumentation__.Notify(579854)

	for _, storingNode := range n.Storing {
		__antithesis_instrumentation__.Notify(579923)
		var columnID catid.ColumnID
		scpb.ForEachColumnName(relationElements, func(_ scpb.Status, target scpb.TargetStatus, e *scpb.ColumnName) {
			__antithesis_instrumentation__.Notify(579926)
			if target == scpb.ToPublic && func() bool {
				__antithesis_instrumentation__.Notify(579927)
				return tree.Name(e.Name) == storingNode == true
			}() == true {
				__antithesis_instrumentation__.Notify(579928)
				columnID = e.ColumnID
			} else {
				__antithesis_instrumentation__.Notify(579929)
			}
		})
		__antithesis_instrumentation__.Notify(579924)
		if columnID == 0 {
			__antithesis_instrumentation__.Notify(579930)
			panic(colinfo.NewUndefinedColumnError(storingNode.String()))
		} else {
			__antithesis_instrumentation__.Notify(579931)
		}
		__antithesis_instrumentation__.Notify(579925)
		index.StoringColumnIDs = append(index.StoringColumnIDs, columnID)
	}
	__antithesis_instrumentation__.Notify(579855)

	if n.Sharded != nil {
		__antithesis_instrumentation__.Notify(579932)
		buckets, err := tabledesc.EvalShardBucketCount(b, b.SemaCtx(), b.EvalCtx(), n.Sharded.ShardBuckets, n.StorageParams)
		if err != nil {
			__antithesis_instrumentation__.Notify(579934)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(579935)
		}
		__antithesis_instrumentation__.Notify(579933)
		shardColName := maybeCreateAndAddShardCol(b, int(buckets), relation.(*scpb.Table), keyColNames)
		index.Sharding = &catpb.ShardedDescriptor{
			IsSharded:    true,
			Name:         shardColName,
			ShardBuckets: buckets,
			ColumnNames:  keyColNames,
		}
	} else {
		__antithesis_instrumentation__.Notify(579936)
	}
	__antithesis_instrumentation__.Notify(579856)

	index.SourceIndexID = source.IndexID
	switch t := relation.(type) {
	case *scpb.Table:
		__antithesis_instrumentation__.Notify(579937)
		index.IndexID = b.NextTableIndexID(t)
	case *scpb.View:
		__antithesis_instrumentation__.Notify(579938)
		index.IndexID = b.NextViewIndexID(t)
	}
	__antithesis_instrumentation__.Notify(579857)
	sec := &scpb.SecondaryIndex{Index: index}
	b.Add(sec)
	b.Add(&scpb.IndexName{
		TableID: index.TableID,
		IndexID: index.IndexID,
		Name:    string(n.Name),
	})
	if n.PartitionByIndex.ContainsPartitions() {
		__antithesis_instrumentation__.Notify(579939)
		b.Add(&scpb.IndexPartitioning{
			TableID:                index.TableID,
			IndexID:                index.IndexID,
			PartitioningDescriptor: b.SecondaryIndexPartitioningDescriptor(sec, n.PartitionByIndex.PartitionBy),
		})
	} else {
		__antithesis_instrumentation__.Notify(579940)
	}
}

func maybeCreateAndAddShardCol(
	b BuildCtx, shardBuckets int, tbl *scpb.Table, colNames []string,
) (shardColName string) {
	__antithesis_instrumentation__.Notify(579941)
	shardColName = tabledesc.GetShardColumnName(colNames, int32(shardBuckets))
	elts := b.QueryByID(tbl.TableID)

	var existingShardColID catid.ColumnID
	scpb.ForEachColumnName(elts, func(_ scpb.Status, target scpb.TargetStatus, name *scpb.ColumnName) {
		__antithesis_instrumentation__.Notify(579946)
		if target == scpb.ToPublic && func() bool {
			__antithesis_instrumentation__.Notify(579947)
			return name.Name == shardColName == true
		}() == true {
			__antithesis_instrumentation__.Notify(579948)
			existingShardColID = name.ColumnID
		} else {
			__antithesis_instrumentation__.Notify(579949)
		}
	})
	__antithesis_instrumentation__.Notify(579942)
	scpb.ForEachColumn(elts, func(_ scpb.Status, _ scpb.TargetStatus, col *scpb.Column) {
		__antithesis_instrumentation__.Notify(579950)
		if col.ColumnID == existingShardColID && func() bool {
			__antithesis_instrumentation__.Notify(579951)
			return !col.IsHidden == true
		}() == true {
			__antithesis_instrumentation__.Notify(579952)

			panic(pgerror.Newf(pgcode.DuplicateColumn,
				"column %s already specified; can't be used for sharding", shardColName))
		} else {
			__antithesis_instrumentation__.Notify(579953)
		}
	})
	__antithesis_instrumentation__.Notify(579943)
	if existingShardColID != 0 {
		__antithesis_instrumentation__.Notify(579954)
		return shardColName
	} else {
		__antithesis_instrumentation__.Notify(579955)
	}
	__antithesis_instrumentation__.Notify(579944)
	expr := schemaexpr.MakeHashShardComputeExpr(colNames, shardBuckets)
	parsedExpr, err := parser.ParseExpr(*expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(579956)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(579957)
	}
	__antithesis_instrumentation__.Notify(579945)
	shardColID := b.NextTableColumnID(tbl)
	spec := addColumnSpec{
		tbl: tbl,
		col: &scpb.Column{
			TableID:  tbl.TableID,
			ColumnID: shardColID,
			IsHidden: true,
		},
		name: &scpb.ColumnName{
			TableID:  tbl.TableID,
			ColumnID: shardColID,
			Name:     shardColName,
		},
		colType: &scpb.ColumnType{
			TableID:     tbl.TableID,
			ColumnID:    shardColID,
			TypeT:       scpb.TypeT{Type: types.Int4},
			ComputeExpr: b.WrapExpression(parsedExpr),
		},
	}
	addColumn(b, spec)
	return shardColName
}

func createVirtualColumnForIndex(
	b BuildCtx, tn *tree.TableName, tbl *scpb.Table, expr tree.Expr,
) string {
	__antithesis_instrumentation__.Notify(579958)
	elts := b.QueryByID(tbl.TableID)
	colName := tabledesc.GenerateUniqueName("crdb_internal_idx_expr", func(name string) (found bool) {
		__antithesis_instrumentation__.Notify(579961)
		scpb.ForEachColumnName(elts, func(_ scpb.Status, target scpb.TargetStatus, cn *scpb.ColumnName) {
			__antithesis_instrumentation__.Notify(579963)
			if target == scpb.ToPublic && func() bool {
				__antithesis_instrumentation__.Notify(579964)
				return cn.Name == name == true
			}() == true {
				__antithesis_instrumentation__.Notify(579965)
				found = true
			} else {
				__antithesis_instrumentation__.Notify(579966)
			}
		})
		__antithesis_instrumentation__.Notify(579962)
		return found
	})
	__antithesis_instrumentation__.Notify(579959)

	d := &tree.ColumnTableDef{
		Name:   tree.Name(colName),
		Hidden: true,
	}
	d.Computed.Computed = true
	d.Computed.Virtual = true
	d.Computed.Expr = expr
	d.Nullable.Nullability = tree.Null

	{
		__antithesis_instrumentation__.Notify(579967)
		colLookupFn := func(columnName tree.Name) (exists bool, accessible bool, id catid.ColumnID, typ *types.T) {
			__antithesis_instrumentation__.Notify(579971)
			scpb.ForEachColumnName(elts, func(_ scpb.Status, target scpb.TargetStatus, cn *scpb.ColumnName) {
				__antithesis_instrumentation__.Notify(579976)
				if target == scpb.ToPublic && func() bool {
					__antithesis_instrumentation__.Notify(579977)
					return tree.Name(cn.Name) == columnName == true
				}() == true {
					__antithesis_instrumentation__.Notify(579978)
					id = cn.ColumnID
				} else {
					__antithesis_instrumentation__.Notify(579979)
				}
			})
			__antithesis_instrumentation__.Notify(579972)
			if id == 0 {
				__antithesis_instrumentation__.Notify(579980)
				return false, false, 0, nil
			} else {
				__antithesis_instrumentation__.Notify(579981)
			}
			__antithesis_instrumentation__.Notify(579973)
			scpb.ForEachColumn(elts, func(_ scpb.Status, target scpb.TargetStatus, col *scpb.Column) {
				__antithesis_instrumentation__.Notify(579982)
				if target == scpb.ToPublic && func() bool {
					__antithesis_instrumentation__.Notify(579983)
					return col.ColumnID == id == true
				}() == true {
					__antithesis_instrumentation__.Notify(579984)
					exists = true
					accessible = !col.IsInaccessible
				} else {
					__antithesis_instrumentation__.Notify(579985)
				}
			})
			__antithesis_instrumentation__.Notify(579974)
			scpb.ForEachColumnType(elts, func(_ scpb.Status, target scpb.TargetStatus, col *scpb.ColumnType) {
				__antithesis_instrumentation__.Notify(579986)
				if target == scpb.ToPublic && func() bool {
					__antithesis_instrumentation__.Notify(579987)
					return col.ColumnID == id == true
				}() == true {
					__antithesis_instrumentation__.Notify(579988)
					typ = col.Type
				} else {
					__antithesis_instrumentation__.Notify(579989)
				}
			})
			__antithesis_instrumentation__.Notify(579975)
			return exists, accessible, id, typ
		}
		__antithesis_instrumentation__.Notify(579968)
		replacedExpr, _, err := schemaexpr.ReplaceColumnVars(expr, colLookupFn)
		if err != nil {
			__antithesis_instrumentation__.Notify(579990)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(579991)
		}
		__antithesis_instrumentation__.Notify(579969)
		typedExpr, err := tree.TypeCheck(b, replacedExpr, b.SemaCtx(), types.Any)
		if err != nil {
			__antithesis_instrumentation__.Notify(579992)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(579993)
		}
		__antithesis_instrumentation__.Notify(579970)
		d.Type = typedExpr.ResolvedType()
	}
	__antithesis_instrumentation__.Notify(579960)
	alterTableAddColumn(b, tn, tbl, &tree.AlterTableAddColumn{ColumnDef: d})
	return colName
}
