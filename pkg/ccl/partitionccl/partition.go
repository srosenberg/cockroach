package partitionccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/ccl/utilccl"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func valueEncodePartitionTuple(
	typ tree.PartitionByType, evalCtx *tree.EvalContext, maybeTuple tree.Expr, cols []catalog.Column,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(20644)

	maybeTuple, _ = tree.WalkExpr(replaceMinMaxValVisitor{}, maybeTuple)

	tuple, ok := maybeTuple.(*tree.Tuple)
	if !ok {
		__antithesis_instrumentation__.Notify(20648)

		tuple = &tree.Tuple{Exprs: []tree.Expr{maybeTuple}}
	} else {
		__antithesis_instrumentation__.Notify(20649)
	}
	__antithesis_instrumentation__.Notify(20645)

	if len(tuple.Exprs) != len(cols) {
		__antithesis_instrumentation__.Notify(20650)
		return nil, errors.Errorf("partition has %d columns but %d values were supplied",
			len(cols), len(tuple.Exprs))
	} else {
		__antithesis_instrumentation__.Notify(20651)
	}
	__antithesis_instrumentation__.Notify(20646)

	var value, scratch []byte
	for i, expr := range tuple.Exprs {
		__antithesis_instrumentation__.Notify(20652)
		expr = tree.StripParens(expr)
		switch expr.(type) {
		case tree.DefaultVal:
			__antithesis_instrumentation__.Notify(20658)
			if typ != tree.PartitionByList {
				__antithesis_instrumentation__.Notify(20666)
				return nil, errors.Errorf("%s cannot be used with PARTITION BY %s", expr, typ)
			} else {
				__antithesis_instrumentation__.Notify(20667)
			}
			__antithesis_instrumentation__.Notify(20659)

			value = encoding.EncodeNotNullValue(value, encoding.NoColumnID)
			value = encoding.EncodeNonsortingUvarint(value, uint64(rowenc.PartitionDefaultVal))
			continue
		case tree.PartitionMinVal:
			__antithesis_instrumentation__.Notify(20660)
			if typ != tree.PartitionByRange {
				__antithesis_instrumentation__.Notify(20668)
				return nil, errors.Errorf("%s cannot be used with PARTITION BY %s", expr, typ)
			} else {
				__antithesis_instrumentation__.Notify(20669)
			}
			__antithesis_instrumentation__.Notify(20661)

			value = encoding.EncodeNotNullValue(value, encoding.NoColumnID)
			value = encoding.EncodeNonsortingUvarint(value, uint64(rowenc.PartitionMinVal))
			continue
		case tree.PartitionMaxVal:
			__antithesis_instrumentation__.Notify(20662)
			if typ != tree.PartitionByRange {
				__antithesis_instrumentation__.Notify(20670)
				return nil, errors.Errorf("%s cannot be used with PARTITION BY %s", expr, typ)
			} else {
				__antithesis_instrumentation__.Notify(20671)
			}
			__antithesis_instrumentation__.Notify(20663)

			value = encoding.EncodeNotNullValue(value, encoding.NoColumnID)
			value = encoding.EncodeNonsortingUvarint(value, uint64(rowenc.PartitionMaxVal))
			continue
		case *tree.Placeholder:
			__antithesis_instrumentation__.Notify(20664)
			return nil, unimplemented.NewWithIssuef(
				19464, "placeholders are not supported in PARTITION BY")
		default:
			__antithesis_instrumentation__.Notify(20665)

		}
		__antithesis_instrumentation__.Notify(20653)

		var semaCtx tree.SemaContext
		typedExpr, err := schemaexpr.SanitizeVarFreeExpr(evalCtx.Context, expr, cols[i].GetType(), "partition",
			&semaCtx,
			tree.VolatilityImmutable,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(20672)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(20673)
		}
		__antithesis_instrumentation__.Notify(20654)
		if !tree.IsConst(evalCtx, typedExpr) {
			__antithesis_instrumentation__.Notify(20674)
			return nil, pgerror.Newf(pgcode.Syntax,
				"%s: partition values must be constant", typedExpr)
		} else {
			__antithesis_instrumentation__.Notify(20675)
		}
		__antithesis_instrumentation__.Notify(20655)
		datum, err := typedExpr.Eval(evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(20676)
			return nil, errors.Wrapf(err, "evaluating %s", typedExpr)
		} else {
			__antithesis_instrumentation__.Notify(20677)
		}
		__antithesis_instrumentation__.Notify(20656)
		if err := colinfo.CheckDatumTypeFitsColumnType(cols[i], datum.ResolvedType()); err != nil {
			__antithesis_instrumentation__.Notify(20678)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(20679)
		}
		__antithesis_instrumentation__.Notify(20657)
		value, err = valueside.Encode(value, valueside.NoColumnID, datum, scratch)
		if err != nil {
			__antithesis_instrumentation__.Notify(20680)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(20681)
		}
	}
	__antithesis_instrumentation__.Notify(20647)
	return value, nil
}

type replaceMinMaxValVisitor struct{}

func (v replaceMinMaxValVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(20682)
	if t, ok := expr.(*tree.UnresolvedName); ok && func() bool {
		__antithesis_instrumentation__.Notify(20684)
		return t.NumParts == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(20685)
		switch t.Parts[0] {
		case "minvalue":
			__antithesis_instrumentation__.Notify(20686)
			return false, tree.PartitionMinVal{}
		case "maxvalue":
			__antithesis_instrumentation__.Notify(20687)
			return false, tree.PartitionMaxVal{}
		default:
			__antithesis_instrumentation__.Notify(20688)
		}
	} else {
		__antithesis_instrumentation__.Notify(20689)
	}
	__antithesis_instrumentation__.Notify(20683)
	return true, expr
}

func (replaceMinMaxValVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(20690)
	return expr
}

func createPartitioningImpl(
	ctx context.Context,
	evalCtx *tree.EvalContext,
	columnLookupFn func(tree.Name) (catalog.Column, error),
	newIdxColumnNames []string,
	partBy *tree.PartitionBy,
	allowedNewColumnNames []tree.Name,
	numImplicitColumns int,
	colOffset int,
) (catpb.PartitioningDescriptor, error) {
	__antithesis_instrumentation__.Notify(20691)
	partDesc := catpb.PartitioningDescriptor{}
	if partBy == nil {
		__antithesis_instrumentation__.Notify(20697)
		return partDesc, nil
	} else {
		__antithesis_instrumentation__.Notify(20698)
	}
	__antithesis_instrumentation__.Notify(20692)
	partDesc.NumColumns = uint32(len(partBy.Fields))
	partDesc.NumImplicitColumns = uint32(numImplicitColumns)

	partitioningString := func() string {
		__antithesis_instrumentation__.Notify(20699)

		partCols := append([]string(nil), newIdxColumnNames[:colOffset]...)
		for _, p := range partBy.Fields {
			__antithesis_instrumentation__.Notify(20701)
			partCols = append(partCols, string(p))
		}
		__antithesis_instrumentation__.Notify(20700)
		return strings.Join(partCols, ", ")
	}
	__antithesis_instrumentation__.Notify(20693)

	var cols []catalog.Column
	for i := 0; i < len(partBy.Fields); i++ {
		__antithesis_instrumentation__.Notify(20702)
		if colOffset+i >= len(newIdxColumnNames) {
			__antithesis_instrumentation__.Notify(20705)
			return partDesc, pgerror.Newf(pgcode.Syntax,
				"declared partition columns (%s) exceed the number of columns in index being partitioned (%s)",
				partitioningString(), strings.Join(newIdxColumnNames, ", "))
		} else {
			__antithesis_instrumentation__.Notify(20706)
		}
		__antithesis_instrumentation__.Notify(20703)

		col, err := findColumnByNameOnTable(
			columnLookupFn,
			tree.Name(newIdxColumnNames[colOffset+i]),
			allowedNewColumnNames,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(20707)
			return partDesc, err
		} else {
			__antithesis_instrumentation__.Notify(20708)
		}
		__antithesis_instrumentation__.Notify(20704)
		cols = append(cols, col)
		if string(partBy.Fields[i]) != col.GetName() {
			__antithesis_instrumentation__.Notify(20709)

			n := colOffset + i + 1
			return partDesc, pgerror.Newf(pgcode.Syntax,
				"declared partition columns (%s) do not match first %d columns in index being partitioned (%s)",
				partitioningString(), n, strings.Join(newIdxColumnNames[:n], ", "))
		} else {
			__antithesis_instrumentation__.Notify(20710)
		}
	}
	__antithesis_instrumentation__.Notify(20694)

	for _, l := range partBy.List {
		__antithesis_instrumentation__.Notify(20711)
		p := catpb.PartitioningDescriptor_List{
			Name: string(l.Name),
		}
		for _, expr := range l.Exprs {
			__antithesis_instrumentation__.Notify(20714)
			encodedTuple, err := valueEncodePartitionTuple(
				tree.PartitionByList, evalCtx, expr, cols)
			if err != nil {
				__antithesis_instrumentation__.Notify(20716)
				return partDesc, errors.Wrapf(err, "PARTITION %s", p.Name)
			} else {
				__antithesis_instrumentation__.Notify(20717)
			}
			__antithesis_instrumentation__.Notify(20715)
			p.Values = append(p.Values, encodedTuple)
		}
		__antithesis_instrumentation__.Notify(20712)
		if l.Subpartition != nil {
			__antithesis_instrumentation__.Notify(20718)
			newColOffset := colOffset + int(partDesc.NumColumns)
			if numImplicitColumns > 0 {
				__antithesis_instrumentation__.Notify(20721)
				return catpb.PartitioningDescriptor{}, unimplemented.New(
					"PARTITION BY SUBPARTITION",
					"implicit column partitioning on a subpartition is not yet supported",
				)
			} else {
				__antithesis_instrumentation__.Notify(20722)
			}
			__antithesis_instrumentation__.Notify(20719)
			subpartitioning, err := createPartitioningImpl(
				ctx,
				evalCtx,
				columnLookupFn,
				newIdxColumnNames,
				l.Subpartition,
				allowedNewColumnNames,
				0,
				newColOffset,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(20723)
				return partDesc, err
			} else {
				__antithesis_instrumentation__.Notify(20724)
			}
			__antithesis_instrumentation__.Notify(20720)
			p.Subpartitioning = subpartitioning
		} else {
			__antithesis_instrumentation__.Notify(20725)
		}
		__antithesis_instrumentation__.Notify(20713)
		partDesc.List = append(partDesc.List, p)
	}
	__antithesis_instrumentation__.Notify(20695)

	for _, r := range partBy.Range {
		__antithesis_instrumentation__.Notify(20726)
		p := catpb.PartitioningDescriptor_Range{
			Name: string(r.Name),
		}
		var err error
		p.FromInclusive, err = valueEncodePartitionTuple(
			tree.PartitionByRange, evalCtx, &tree.Tuple{Exprs: r.From}, cols)
		if err != nil {
			__antithesis_instrumentation__.Notify(20730)
			return partDesc, errors.Wrapf(err, "PARTITION %s", p.Name)
		} else {
			__antithesis_instrumentation__.Notify(20731)
		}
		__antithesis_instrumentation__.Notify(20727)
		p.ToExclusive, err = valueEncodePartitionTuple(
			tree.PartitionByRange, evalCtx, &tree.Tuple{Exprs: r.To}, cols)
		if err != nil {
			__antithesis_instrumentation__.Notify(20732)
			return partDesc, errors.Wrapf(err, "PARTITION %s", p.Name)
		} else {
			__antithesis_instrumentation__.Notify(20733)
		}
		__antithesis_instrumentation__.Notify(20728)
		if r.Subpartition != nil {
			__antithesis_instrumentation__.Notify(20734)
			return partDesc, errors.Newf("PARTITION %s: cannot subpartition a range partition", p.Name)
		} else {
			__antithesis_instrumentation__.Notify(20735)
		}
		__antithesis_instrumentation__.Notify(20729)
		partDesc.Range = append(partDesc.Range, p)
	}
	__antithesis_instrumentation__.Notify(20696)

	return partDesc, nil
}

func collectImplicitPartitionColumns(
	columnLookupFn func(tree.Name) (catalog.Column, error),
	indexFirstColumnName string,
	partBy *tree.PartitionBy,
	allowedNewColumnNames []tree.Name,
) (implicitCols []catalog.Column, _ error) {
	__antithesis_instrumentation__.Notify(20736)
	seenImplicitColumnNames := map[string]struct{}{}

	for _, field := range partBy.Fields {
		__antithesis_instrumentation__.Notify(20738)

		if string(field) == indexFirstColumnName {
			__antithesis_instrumentation__.Notify(20742)
			break
		} else {
			__antithesis_instrumentation__.Notify(20743)
		}
		__antithesis_instrumentation__.Notify(20739)

		col, err := findColumnByNameOnTable(
			columnLookupFn,
			field,
			allowedNewColumnNames,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(20744)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(20745)
		}
		__antithesis_instrumentation__.Notify(20740)
		if _, ok := seenImplicitColumnNames[col.GetName()]; ok {
			__antithesis_instrumentation__.Notify(20746)
			return nil, pgerror.Newf(
				pgcode.InvalidObjectDefinition,
				`found multiple definitions in partition using column "%s"`,
				col.GetName(),
			)
		} else {
			__antithesis_instrumentation__.Notify(20747)
		}
		__antithesis_instrumentation__.Notify(20741)
		seenImplicitColumnNames[col.GetName()] = struct{}{}
		implicitCols = append(implicitCols, col)
	}
	__antithesis_instrumentation__.Notify(20737)

	return implicitCols, nil
}

func findColumnByNameOnTable(
	columnLookupFn func(tree.Name) (catalog.Column, error),
	col tree.Name,
	allowedNewColumnNames []tree.Name,
) (catalog.Column, error) {
	__antithesis_instrumentation__.Notify(20748)
	ret, err := columnLookupFn(col)
	if err != nil {
		__antithesis_instrumentation__.Notify(20752)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20753)
	}
	__antithesis_instrumentation__.Notify(20749)
	if ret.Public() {
		__antithesis_instrumentation__.Notify(20754)
		return ret, nil
	} else {
		__antithesis_instrumentation__.Notify(20755)
	}
	__antithesis_instrumentation__.Notify(20750)
	for _, allowedNewColName := range allowedNewColumnNames {
		__antithesis_instrumentation__.Notify(20756)
		if allowedNewColName == col {
			__antithesis_instrumentation__.Notify(20757)
			return ret, nil
		} else {
			__antithesis_instrumentation__.Notify(20758)
		}
	}
	__antithesis_instrumentation__.Notify(20751)
	return nil, colinfo.NewUndefinedColumnError(string(col))
}

func createPartitioning(
	ctx context.Context,
	st *cluster.Settings,
	evalCtx *tree.EvalContext,
	columnLookupFn func(tree.Name) (catalog.Column, error),
	oldNumImplicitColumns int,
	oldKeyColumnNames []string,
	partBy *tree.PartitionBy,
	allowedNewColumnNames []tree.Name,
	allowImplicitPartitioning bool,
) (newImplicitCols []catalog.Column, newPartitioning catpb.PartitioningDescriptor, err error) {
	__antithesis_instrumentation__.Notify(20759)
	org := sql.ClusterOrganization.Get(&st.SV)
	if err := utilccl.CheckEnterpriseEnabled(st, evalCtx.ClusterID, org, "partitions"); err != nil {
		__antithesis_instrumentation__.Notify(20765)
		return nil, newPartitioning, err
	} else {
		__antithesis_instrumentation__.Notify(20766)
	}
	__antithesis_instrumentation__.Notify(20760)

	newIdxColumnNames := oldKeyColumnNames[oldNumImplicitColumns:]

	if allowImplicitPartitioning {
		__antithesis_instrumentation__.Notify(20767)
		newImplicitCols, err = collectImplicitPartitionColumns(
			columnLookupFn,
			newIdxColumnNames[0],
			partBy,
			allowedNewColumnNames,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(20768)
			return nil, newPartitioning, err
		} else {
			__antithesis_instrumentation__.Notify(20769)
		}
	} else {
		__antithesis_instrumentation__.Notify(20770)
	}
	__antithesis_instrumentation__.Notify(20761)
	if len(newImplicitCols) > 0 {
		__antithesis_instrumentation__.Notify(20771)

		newIdxColumnNames = make([]string, len(newImplicitCols), len(newImplicitCols)+len(newIdxColumnNames))
		for i, col := range newImplicitCols {
			__antithesis_instrumentation__.Notify(20773)
			newIdxColumnNames[i] = col.GetName()
		}
		__antithesis_instrumentation__.Notify(20772)
		newIdxColumnNames = append(newIdxColumnNames, oldKeyColumnNames[oldNumImplicitColumns:]...)
	} else {
		__antithesis_instrumentation__.Notify(20774)
	}
	__antithesis_instrumentation__.Notify(20762)

	if oldNumImplicitColumns > 0 {
		__antithesis_instrumentation__.Notify(20775)
		if len(newImplicitCols) != oldNumImplicitColumns {
			__antithesis_instrumentation__.Notify(20777)
			return nil, newPartitioning, errors.AssertionFailedf(
				"mismatching number of implicit columns: old %d vs new %d",
				oldNumImplicitColumns,
				len(newImplicitCols),
			)
		} else {
			__antithesis_instrumentation__.Notify(20778)
		}
		__antithesis_instrumentation__.Notify(20776)
		for i, col := range newImplicitCols {
			__antithesis_instrumentation__.Notify(20779)
			if oldKeyColumnNames[i] != col.GetName() {
				__antithesis_instrumentation__.Notify(20780)
				return nil, newPartitioning, errors.AssertionFailedf("found new implicit partitioning at column ordinal %d", i)
			} else {
				__antithesis_instrumentation__.Notify(20781)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(20782)
	}
	__antithesis_instrumentation__.Notify(20763)

	newPartitioning, err = createPartitioningImpl(
		ctx,
		evalCtx,
		columnLookupFn,
		newIdxColumnNames,
		partBy,
		allowedNewColumnNames,
		len(newImplicitCols),
		0,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(20783)
		return nil, catpb.PartitioningDescriptor{}, err
	} else {
		__antithesis_instrumentation__.Notify(20784)
	}
	__antithesis_instrumentation__.Notify(20764)
	return newImplicitCols, newPartitioning, err
}

func selectPartitionExprs(
	evalCtx *tree.EvalContext, tableDesc catalog.TableDescriptor, partNames tree.NameList,
) (tree.Expr, error) {
	__antithesis_instrumentation__.Notify(20785)
	exprsByPartName := make(map[string]tree.TypedExpr)
	for _, partName := range partNames {
		__antithesis_instrumentation__.Notify(20791)
		exprsByPartName[string(partName)] = nil
	}
	__antithesis_instrumentation__.Notify(20786)

	a := &tree.DatumAlloc{}
	var prefixDatums []tree.Datum
	if err := catalog.ForEachIndex(tableDesc, catalog.IndexOpts{
		AddMutations: true,
	}, func(idx catalog.Index) error {
		__antithesis_instrumentation__.Notify(20792)
		return selectPartitionExprsByName(
			a, evalCtx, tableDesc, idx, idx.GetPartitioning(), prefixDatums, exprsByPartName, true)
	}); err != nil {
		__antithesis_instrumentation__.Notify(20793)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20794)
	}
	__antithesis_instrumentation__.Notify(20787)

	var expr tree.TypedExpr = tree.DBoolFalse
	for _, partName := range partNames {
		__antithesis_instrumentation__.Notify(20795)
		partExpr, ok := exprsByPartName[string(partName)]
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(20797)
			return partExpr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(20798)
			return nil, errors.Errorf("unknown partition: %s", partName)
		} else {
			__antithesis_instrumentation__.Notify(20799)
		}
		__antithesis_instrumentation__.Notify(20796)
		expr = tree.NewTypedOrExpr(expr, partExpr)
	}
	__antithesis_instrumentation__.Notify(20788)

	var err error
	expr, err = evalCtx.NormalizeExpr(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(20800)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(20801)
	}
	__antithesis_instrumentation__.Notify(20789)

	finalExpr, err := tree.SimpleVisit(expr, func(e tree.Expr) (recurse bool, newExpr tree.Expr, _ error) {
		__antithesis_instrumentation__.Notify(20802)
		if ivar, ok := e.(*tree.IndexedVar); ok {
			__antithesis_instrumentation__.Notify(20804)
			col, err := tableDesc.FindColumnWithID(descpb.ColumnID(ivar.Idx))
			if err != nil {
				__antithesis_instrumentation__.Notify(20806)
				return false, nil, err
			} else {
				__antithesis_instrumentation__.Notify(20807)
			}
			__antithesis_instrumentation__.Notify(20805)
			return false, &tree.ColumnItem{ColumnName: tree.Name(col.GetName())}, nil
		} else {
			__antithesis_instrumentation__.Notify(20808)
		}
		__antithesis_instrumentation__.Notify(20803)
		return true, e, nil
	})
	__antithesis_instrumentation__.Notify(20790)
	return finalExpr, err
}

func selectPartitionExprsByName(
	a *tree.DatumAlloc,
	evalCtx *tree.EvalContext,
	tableDesc catalog.TableDescriptor,
	idx catalog.Index,
	part catalog.Partitioning,
	prefixDatums tree.Datums,
	exprsByPartName map[string]tree.TypedExpr,
	genExpr bool,
) error {
	__antithesis_instrumentation__.Notify(20809)
	if part.NumColumns() == 0 {
		__antithesis_instrumentation__.Notify(20815)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(20816)
	}
	__antithesis_instrumentation__.Notify(20810)

	if !genExpr {
		__antithesis_instrumentation__.Notify(20817)
		err := part.ForEachList(func(name string, _ [][]byte, subPartitioning catalog.Partitioning) error {
			__antithesis_instrumentation__.Notify(20820)
			exprsByPartName[name] = tree.DBoolFalse
			var fakeDatums tree.Datums
			return selectPartitionExprsByName(a, evalCtx, tableDesc, idx, subPartitioning, fakeDatums, exprsByPartName, genExpr)
		})
		__antithesis_instrumentation__.Notify(20818)
		if err != nil {
			__antithesis_instrumentation__.Notify(20821)
			return err
		} else {
			__antithesis_instrumentation__.Notify(20822)
		}
		__antithesis_instrumentation__.Notify(20819)
		return part.ForEachRange(func(name string, _, _ []byte) error {
			__antithesis_instrumentation__.Notify(20823)
			exprsByPartName[name] = tree.DBoolFalse
			return nil
		})
	} else {
		__antithesis_instrumentation__.Notify(20824)
	}
	__antithesis_instrumentation__.Notify(20811)

	var colVars tree.Exprs
	{
		__antithesis_instrumentation__.Notify(20825)

		colVars = make(tree.Exprs, len(prefixDatums)+part.NumColumns())
		for i := range colVars {
			__antithesis_instrumentation__.Notify(20826)
			col, err := tabledesc.FindPublicColumnWithID(tableDesc, idx.GetKeyColumnID(i))
			if err != nil {
				__antithesis_instrumentation__.Notify(20828)
				return err
			} else {
				__antithesis_instrumentation__.Notify(20829)
			}
			__antithesis_instrumentation__.Notify(20827)
			colVars[i] = tree.NewTypedOrdinalReference(int(col.GetID()), col.GetType())
		}
	}
	__antithesis_instrumentation__.Notify(20812)

	if part.NumLists() > 0 {
		__antithesis_instrumentation__.Notify(20830)
		type exprAndPartName struct {
			expr tree.TypedExpr
			name string
		}

		partValueExprs := make([][]exprAndPartName, part.NumColumns()+1)

		err := part.ForEachList(func(name string, values [][]byte, subPartitioning catalog.Partitioning) error {
			__antithesis_instrumentation__.Notify(20833)
			for _, valueEncBuf := range values {
				__antithesis_instrumentation__.Notify(20835)
				t, _, err := rowenc.DecodePartitionTuple(a, evalCtx.Codec, tableDesc, idx, part, valueEncBuf, prefixDatums)
				if err != nil {
					__antithesis_instrumentation__.Notify(20839)
					return err
				} else {
					__antithesis_instrumentation__.Notify(20840)
				}
				__antithesis_instrumentation__.Notify(20836)
				allDatums := append(prefixDatums, t.Datums...)

				typContents := make([]*types.T, len(allDatums))
				for i, d := range allDatums {
					__antithesis_instrumentation__.Notify(20841)
					typContents[i] = d.ResolvedType()
				}
				__antithesis_instrumentation__.Notify(20837)
				tupleTyp := types.MakeTuple(typContents)
				partValueExpr := tree.NewTypedComparisonExpr(
					treecmp.MakeComparisonOperator(treecmp.EQ),
					tree.NewTypedTuple(tupleTyp, colVars[:len(allDatums)]),
					tree.NewDTuple(tupleTyp, allDatums...),
				)
				partValueExprs[len(t.Datums)] = append(partValueExprs[len(t.Datums)], exprAndPartName{
					expr: partValueExpr,
					name: name,
				})

				genExpr := true
				if _, ok := exprsByPartName[name]; ok {
					__antithesis_instrumentation__.Notify(20842)

					genExpr = false
				} else {
					__antithesis_instrumentation__.Notify(20843)
				}
				__antithesis_instrumentation__.Notify(20838)
				if err := selectPartitionExprsByName(
					a, evalCtx, tableDesc, idx, subPartitioning, allDatums, exprsByPartName, genExpr,
				); err != nil {
					__antithesis_instrumentation__.Notify(20844)
					return err
				} else {
					__antithesis_instrumentation__.Notify(20845)
				}
			}
			__antithesis_instrumentation__.Notify(20834)
			return nil
		})
		__antithesis_instrumentation__.Notify(20831)
		if err != nil {
			__antithesis_instrumentation__.Notify(20846)
			return err
		} else {
			__antithesis_instrumentation__.Notify(20847)
		}
		__antithesis_instrumentation__.Notify(20832)

		excludeExpr := tree.TypedExpr(tree.DBoolFalse)
		for i := len(partValueExprs) - 1; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(20848)
			nextExcludeExpr := tree.TypedExpr(tree.DBoolFalse)
			for _, v := range partValueExprs[i] {
				__antithesis_instrumentation__.Notify(20850)
				nextExcludeExpr = tree.NewTypedOrExpr(nextExcludeExpr, v.expr)
				partValueExpr := tree.NewTypedAndExpr(v.expr, tree.NewTypedNotExpr(excludeExpr))

				if e, ok := exprsByPartName[v.name]; !ok || func() bool {
					__antithesis_instrumentation__.Notify(20851)
					return e == nil == true
				}() == true {
					__antithesis_instrumentation__.Notify(20852)
					exprsByPartName[v.name] = partValueExpr
				} else {
					__antithesis_instrumentation__.Notify(20853)
					exprsByPartName[v.name] = tree.NewTypedOrExpr(e, partValueExpr)
				}
			}
			__antithesis_instrumentation__.Notify(20849)
			excludeExpr = tree.NewTypedOrExpr(excludeExpr, nextExcludeExpr)
		}
	} else {
		__antithesis_instrumentation__.Notify(20854)
	}
	__antithesis_instrumentation__.Notify(20813)

	if part.NumRanges() > 0 {
		__antithesis_instrumentation__.Notify(20855)
		log.Fatal(evalCtx.Context, "TODO(dan): unsupported for range partitionings")
	} else {
		__antithesis_instrumentation__.Notify(20856)
	}
	__antithesis_instrumentation__.Notify(20814)
	return nil
}

func init() {
	sql.CreatePartitioningCCL = createPartitioning
	scdeps.CreatePartitioningCCL = createPartitioning
}
