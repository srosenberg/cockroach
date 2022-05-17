package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func (s *Smither) makeStmt() (stmt tree.Statement, ok bool) {
	__antithesis_instrumentation__.Notify(68990)
	return s.stmtSampler.Next()(s)
}

func (s *Smither) makeSelectStmt(
	desiredTypes []*types.T, refs colRefs, withTables tableRefs,
) (stmt tree.SelectStatement, stmtRefs colRefs, ok bool) {
	__antithesis_instrumentation__.Notify(68991)
	if s.canRecurse() {
		__antithesis_instrumentation__.Notify(68993)
		for {
			__antithesis_instrumentation__.Notify(68994)
			expr, exprRefs, ok := s.selectStmtSampler.Next()(s, desiredTypes, refs, withTables)
			if ok {
				__antithesis_instrumentation__.Notify(68995)
				return expr, exprRefs, ok
			} else {
				__antithesis_instrumentation__.Notify(68996)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(68997)
	}
	__antithesis_instrumentation__.Notify(68992)
	return makeValuesSelect(s, desiredTypes, refs, withTables)
}

func makeSchemaTable(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(68998)

	if len(s.tables) == 0 {
		__antithesis_instrumentation__.Notify(69000)
		return makeValuesTable(s, refs, forJoin)
	} else {
		__antithesis_instrumentation__.Notify(69001)
	}
	__antithesis_instrumentation__.Notify(68999)
	expr, _, _, exprRefs, ok := s.getSchemaTableWithIndexHint()
	return expr, exprRefs, ok
}

func (s *Smither) getSchemaTable() (tree.TableExpr, *tree.TableName, *tableRef, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69002)
	ate, name, tableRef, refs, ok := s.getSchemaTableWithIndexHint()
	if ate != nil {
		__antithesis_instrumentation__.Notify(69004)
		ate.IndexFlags = nil
	} else {
		__antithesis_instrumentation__.Notify(69005)
	}
	__antithesis_instrumentation__.Notify(69003)
	return ate, name, tableRef, refs, ok
}

func (s *Smither) getSchemaTableWithIndexHint() (
	*tree.AliasedTableExpr,
	*tree.TableName,
	*tableRef,
	colRefs,
	bool,
) {
	__antithesis_instrumentation__.Notify(69006)
	table, ok := s.getRandTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69008)
		return nil, nil, nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69009)
	}
	__antithesis_instrumentation__.Notify(69007)
	alias := s.name("tab")
	name := tree.NewUnqualifiedTableName(alias)
	expr, refs := s.tableExpr(table.tableRef, name)
	return &tree.AliasedTableExpr{
		Expr:       expr,
		IndexFlags: table.indexFlags,
		As:         tree.AliasClause{Alias: alias},
	}, name, table.tableRef, refs, true
}

func (s *Smither) tableExpr(table *tableRef, name *tree.TableName) (tree.TableExpr, colRefs) {
	__antithesis_instrumentation__.Notify(69010)
	refs := make(colRefs, len(table.Columns))
	for i, c := range table.Columns {
		__antithesis_instrumentation__.Notify(69012)
		refs[i] = &colRef{
			typ: tree.MustBeStaticallyKnownType(c.Type),
			item: tree.NewColumnItem(
				name,
				c.Name,
			),
		}
	}
	__antithesis_instrumentation__.Notify(69011)
	return table.TableName, refs
}

var (
	mutatingStatements = []statementWeight{
		{10, makeInsert},
		{10, makeDelete},
		{10, makeUpdate},
		{1, makeAlter},
		{1, makeBegin},
		{2, makeRollback},
		{6, makeCommit},
		{1, makeBackup},
		{1, makeRestore},
		{1, makeExport},
		{1, makeImport},
	}
	nonMutatingStatements = []statementWeight{
		{10, makeSelect},
	}
	allStatements = append(mutatingStatements, nonMutatingStatements...)

	mutatingTableExprs = []tableExprWeight{
		{1, makeInsertReturning},
		{1, makeDeleteReturning},
		{1, makeUpdateReturning},
	}
	nonMutatingTableExprs = []tableExprWeight{
		{40, makeMergeJoinExpr},
		{40, makeEquiJoinExpr},
		{20, makeSchemaTable},
		{10, makeJoinExpr},
		{1, makeValuesTable},
		{2, makeSelectTable},
	}
	allTableExprs = append(mutatingTableExprs, nonMutatingTableExprs...)

	selectStmts = []selectStatementWeight{
		{1, makeValuesSelect},
		{1, makeSetOp},
		{1, makeSelectClause},
	}
)

func makeTableExpr(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69013)
	if s.canRecurse() {
		__antithesis_instrumentation__.Notify(69015)
		for i := 0; i < retryCount; i++ {
			__antithesis_instrumentation__.Notify(69016)
			expr, exprRefs, ok := s.tableExprSampler.Next()(s, refs, forJoin)
			if ok {
				__antithesis_instrumentation__.Notify(69017)
				return expr, exprRefs, ok
			} else {
				__antithesis_instrumentation__.Notify(69018)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(69019)
	}
	__antithesis_instrumentation__.Notify(69014)
	return makeSchemaTable(s, refs, forJoin)
}

type typedExpr struct {
	tree.TypedExpr
	typ *types.T
}

func makeTypedExpr(expr tree.TypedExpr, typ *types.T) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69020)
	return typedExpr{
		TypedExpr: expr,
		typ:       typ,
	}
}

func (t typedExpr) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(69021)
	return t.typ
}

var joinTypes = []string{
	"",
	tree.AstFull,
	tree.AstLeft,
	tree.AstRight,
	tree.AstCross,
	tree.AstInner,
}

func makeJoinExpr(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69022)
	left, leftRefs, ok := makeTableExpr(s, refs, true)
	if !ok {
		__antithesis_instrumentation__.Notify(69026)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69027)
	}
	__antithesis_instrumentation__.Notify(69023)
	right, rightRefs, ok := makeTableExpr(s, refs, true)
	if !ok {
		__antithesis_instrumentation__.Notify(69028)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69029)
	}
	__antithesis_instrumentation__.Notify(69024)

	joinExpr := &tree.JoinTableExpr{
		JoinType: joinTypes[s.rnd.Intn(len(joinTypes))],
		Left:     left,
		Right:    right,
	}

	if joinExpr.JoinType != tree.AstCross {
		__antithesis_instrumentation__.Notify(69030)
		on := makeBoolExpr(s, refs)
		joinExpr.Cond = &tree.OnJoinCond{Expr: on}
	} else {
		__antithesis_instrumentation__.Notify(69031)
	}
	__antithesis_instrumentation__.Notify(69025)
	joinRefs := leftRefs.extend(rightRefs...)

	return joinExpr, joinRefs, true
}

func makeEquiJoinExpr(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69032)
	left, leftRefs, ok := makeTableExpr(s, refs, true)
	if !ok {
		__antithesis_instrumentation__.Notify(69039)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69040)
	}
	__antithesis_instrumentation__.Notify(69033)
	right, rightRefs, ok := makeTableExpr(s, refs, true)
	if !ok {
		__antithesis_instrumentation__.Notify(69041)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69042)
	}
	__antithesis_instrumentation__.Notify(69034)

	s.lock.RLock()
	defer s.lock.RUnlock()

	var available [][2]tree.TypedExpr
	for _, leftCol := range leftRefs {
		__antithesis_instrumentation__.Notify(69043)
		for _, rightCol := range rightRefs {
			__antithesis_instrumentation__.Notify(69044)

			if !s.isScalarType(leftCol.typ) || func() bool {
				__antithesis_instrumentation__.Notify(69046)
				return !s.isScalarType(rightCol.typ) == true
			}() == true {
				__antithesis_instrumentation__.Notify(69047)
				continue
			} else {
				__antithesis_instrumentation__.Notify(69048)
			}
			__antithesis_instrumentation__.Notify(69045)

			if leftCol.typ.Equivalent(rightCol.typ) {
				__antithesis_instrumentation__.Notify(69049)
				available = append(available, [2]tree.TypedExpr{
					typedParen(leftCol.item, leftCol.typ),
					typedParen(rightCol.item, rightCol.typ),
				})
			} else {
				__antithesis_instrumentation__.Notify(69050)
			}
		}
	}
	__antithesis_instrumentation__.Notify(69035)
	if len(available) == 0 {
		__antithesis_instrumentation__.Notify(69051)

		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69052)
	}
	__antithesis_instrumentation__.Notify(69036)

	s.rnd.Shuffle(len(available), func(i, j int) {
		__antithesis_instrumentation__.Notify(69053)
		available[i], available[j] = available[j], available[i]
	})
	__antithesis_instrumentation__.Notify(69037)

	var cond tree.TypedExpr
	for (cond == nil || func() bool {
		__antithesis_instrumentation__.Notify(69054)
		return s.coin() == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(69055)
		return len(available) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69056)
		v := available[0]
		available = available[1:]
		expr := tree.NewTypedComparisonExpr(treecmp.MakeComparisonOperator(treecmp.EQ), v[0], v[1])
		if cond == nil {
			__antithesis_instrumentation__.Notify(69057)
			cond = expr
		} else {
			__antithesis_instrumentation__.Notify(69058)
			cond = tree.NewTypedAndExpr(cond, expr)
		}
	}
	__antithesis_instrumentation__.Notify(69038)

	joinExpr := &tree.JoinTableExpr{
		Left:  left,
		Right: right,
		Cond:  &tree.OnJoinCond{Expr: cond},
	}
	joinRefs := leftRefs.extend(rightRefs...)

	ok = cond != nil
	return joinExpr, joinRefs, ok
}

func makeMergeJoinExpr(s *Smither, _ colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69059)

	leftTableName, leftIdx, _, ok := s.getRandIndex()
	if !ok {
		__antithesis_instrumentation__.Notify(69065)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69066)
	}
	__antithesis_instrumentation__.Notify(69060)

	leftAlias := s.name("tab")
	rightAlias := s.name("tab")
	leftAliasName := tree.NewUnqualifiedTableName(leftAlias)
	rightAliasName := tree.NewUnqualifiedTableName(rightAlias)

	rightTableName, cols, ok := func() (*tree.TableIndexName, [][2]colRef, bool) {
		__antithesis_instrumentation__.Notify(69067)
		s.lock.RLock()
		defer s.lock.RUnlock()
		for tbl, idxs := range s.indexes {
			__antithesis_instrumentation__.Notify(69069)
			for idxName, idx := range idxs {
				__antithesis_instrumentation__.Notify(69070)
				rightTableName := &tree.TableIndexName{
					Table: tbl,
					Index: tree.UnrestrictedName(idxName),
				}

				var cols [][2]colRef
				for _, rightColElem := range idx.Columns {
					__antithesis_instrumentation__.Notify(69072)
					rightCol := s.columns[tbl][rightColElem.Column]
					leftColElem := leftIdx.Columns[len(cols)]
					leftCol := s.columns[leftTableName.Table][leftColElem.Column]
					if rightColElem.Direction != leftColElem.Direction {
						__antithesis_instrumentation__.Notify(69077)
						break
					} else {
						__antithesis_instrumentation__.Notify(69078)
					}
					__antithesis_instrumentation__.Notify(69073)
					if leftCol == nil || func() bool {
						__antithesis_instrumentation__.Notify(69079)
						return rightCol == nil == true
					}() == true {
						__antithesis_instrumentation__.Notify(69080)

						break
					} else {
						__antithesis_instrumentation__.Notify(69081)
					}
					__antithesis_instrumentation__.Notify(69074)
					if !tree.MustBeStaticallyKnownType(rightCol.Type).Equivalent(tree.MustBeStaticallyKnownType(leftCol.Type)) {
						__antithesis_instrumentation__.Notify(69082)
						break
					} else {
						__antithesis_instrumentation__.Notify(69083)
					}
					__antithesis_instrumentation__.Notify(69075)
					leftType := tree.MustBeStaticallyKnownType(leftCol.Type)
					rightType := tree.MustBeStaticallyKnownType(rightCol.Type)

					if !s.isScalarType(leftType) || func() bool {
						__antithesis_instrumentation__.Notify(69084)
						return !s.isScalarType(rightType) == true
					}() == true {
						__antithesis_instrumentation__.Notify(69085)
						break
					} else {
						__antithesis_instrumentation__.Notify(69086)
					}
					__antithesis_instrumentation__.Notify(69076)
					cols = append(cols, [2]colRef{
						{
							typ:  leftType,
							item: tree.NewColumnItem(leftAliasName, leftColElem.Column),
						},
						{
							typ:  rightType,
							item: tree.NewColumnItem(rightAliasName, rightColElem.Column),
						},
					})
					if len(cols) >= len(leftIdx.Columns) {
						__antithesis_instrumentation__.Notify(69087)
						break
					} else {
						__antithesis_instrumentation__.Notify(69088)
					}
				}
				__antithesis_instrumentation__.Notify(69071)
				if len(cols) > 0 {
					__antithesis_instrumentation__.Notify(69089)
					return rightTableName, cols, true
				} else {
					__antithesis_instrumentation__.Notify(69090)
				}
			}
		}
		__antithesis_instrumentation__.Notify(69068)
		return nil, nil, false
	}()
	__antithesis_instrumentation__.Notify(69061)
	if !ok {
		__antithesis_instrumentation__.Notify(69091)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69092)
	}
	__antithesis_instrumentation__.Notify(69062)

	var joinRefs colRefs
	for _, pair := range cols {
		__antithesis_instrumentation__.Notify(69093)
		joinRefs = append(joinRefs, &pair[0], &pair[1])
	}
	__antithesis_instrumentation__.Notify(69063)

	var cond tree.TypedExpr

	for (cond == nil || func() bool {
		__antithesis_instrumentation__.Notify(69094)
		return s.coin() == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(69095)
		return len(cols) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69096)
		v := cols[0]
		cols = cols[1:]
		expr := tree.NewTypedComparisonExpr(
			treecmp.MakeComparisonOperator(treecmp.EQ),
			typedParen(v[0].item, v[0].typ),
			typedParen(v[1].item, v[1].typ),
		)
		if cond == nil {
			__antithesis_instrumentation__.Notify(69097)
			cond = expr
		} else {
			__antithesis_instrumentation__.Notify(69098)
			cond = tree.NewTypedAndExpr(cond, expr)
		}
	}
	__antithesis_instrumentation__.Notify(69064)

	joinExpr := &tree.JoinTableExpr{
		Left: &tree.AliasedTableExpr{
			Expr: &leftTableName.Table,
			As:   tree.AliasClause{Alias: leftAlias},
		},
		Right: &tree.AliasedTableExpr{
			Expr: &rightTableName.Table,
			As:   tree.AliasClause{Alias: rightAlias},
		},
		Cond: &tree.OnJoinCond{Expr: cond},
	}

	ok = cond != nil
	return joinExpr, joinRefs, ok
}

func (s *Smither) makeWith() (*tree.With, tableRefs) {
	__antithesis_instrumentation__.Notify(69099)
	if s.disableWith {
		__antithesis_instrumentation__.Notify(69104)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(69105)
	}
	__antithesis_instrumentation__.Notify(69100)

	if s.coin() {
		__antithesis_instrumentation__.Notify(69106)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(69107)
	}
	__antithesis_instrumentation__.Notify(69101)
	ctes := make([]*tree.CTE, 0, s.d6()/2)
	if cap(ctes) == 0 {
		__antithesis_instrumentation__.Notify(69108)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(69109)
	}
	__antithesis_instrumentation__.Notify(69102)
	var tables tableRefs
	for i := 0; i < cap(ctes); i++ {
		__antithesis_instrumentation__.Notify(69110)
		var ok bool
		var stmt tree.SelectStatement
		var stmtRefs colRefs
		stmt, stmtRefs, ok = s.makeSelectStmt(s.makeDesiredTypes(), nil, tables)
		if !ok {
			__antithesis_instrumentation__.Notify(69113)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69114)
		}
		__antithesis_instrumentation__.Notify(69111)
		alias := s.name("with")
		tblName := tree.NewUnqualifiedTableName(alias)
		cols := make(tree.NameList, len(stmtRefs))
		defs := make([]*tree.ColumnTableDef, len(stmtRefs))
		for i, r := range stmtRefs {
			__antithesis_instrumentation__.Notify(69115)
			var err error
			cols[i] = r.item.ColumnName
			defs[i], err = tree.NewColumnTableDef(r.item.ColumnName, r.typ, false, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(69116)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(69117)
			}
		}
		__antithesis_instrumentation__.Notify(69112)
		tables = append(tables, &tableRef{
			TableName: tblName,
			Columns:   defs,
		})
		ctes = append(ctes, &tree.CTE{
			Name: tree.AliasClause{
				Alias: alias,
				Cols:  cols,
			},
			Stmt: stmt,
		})
	}
	__antithesis_instrumentation__.Notify(69103)
	return &tree.With{
		CTEList: ctes,
	}, tables
}

var orderDirections = []tree.Direction{
	tree.DefaultDirection,
	tree.Ascending,
	tree.Descending,
}

func (s *Smither) randDirection() tree.Direction {
	__antithesis_instrumentation__.Notify(69118)
	return orderDirections[s.rnd.Intn(len(orderDirections))]
}

var nullabilities = []tree.Nullability{
	tree.NotNull,
	tree.Null,
	tree.SilentNull,
}

func (s *Smither) randNullability() tree.Nullability {
	__antithesis_instrumentation__.Notify(69119)
	return nullabilities[s.rnd.Intn(len(nullabilities))]
}

var dropBehaviors = []tree.DropBehavior{
	tree.DropDefault,
	tree.DropRestrict,
	tree.DropCascade,
}

func (s *Smither) randDropBehavior() tree.DropBehavior {
	__antithesis_instrumentation__.Notify(69120)
	return dropBehaviors[s.rnd.Intn(len(dropBehaviors))]
}

var stringComparisons = []treecmp.ComparisonOperatorSymbol{
	treecmp.Like,
	treecmp.NotLike,
	treecmp.ILike,
	treecmp.NotILike,
	treecmp.SimilarTo,
	treecmp.NotSimilarTo,
	treecmp.RegMatch,
	treecmp.NotRegMatch,
	treecmp.RegIMatch,
	treecmp.NotRegIMatch,
}

func (s *Smither) randStringComparison() treecmp.ComparisonOperator {
	__antithesis_instrumentation__.Notify(69121)
	return treecmp.MakeComparisonOperator(stringComparisons[s.rnd.Intn(len(stringComparisons))])
}

func makeSelectTable(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69122)
	stmt, stmtRefs, ok := s.makeSelect(nil, refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69125)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69126)
	}
	__antithesis_instrumentation__.Notify(69123)

	table := s.name("tab")
	names := make(tree.NameList, len(stmtRefs))
	clauseRefs := make(colRefs, len(stmtRefs))
	for i, ref := range stmtRefs {
		__antithesis_instrumentation__.Notify(69127)
		names[i] = s.name("col")
		clauseRefs[i] = &colRef{
			typ: ref.typ,
			item: tree.NewColumnItem(
				tree.NewUnqualifiedTableName(table),
				names[i],
			),
		}
	}
	__antithesis_instrumentation__.Notify(69124)

	return &tree.AliasedTableExpr{
		Expr: &tree.Subquery{
			Select: &tree.ParenSelect{Select: stmt},
		},
		As: tree.AliasClause{
			Alias: table,
			Cols:  names,
		},
	}, clauseRefs, true
}

func makeSelectClause(
	s *Smither, desiredTypes []*types.T, refs colRefs, withTables tableRefs,
) (tree.SelectStatement, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69128)
	stmt, selectRefs, _, ok := s.makeSelectClause(desiredTypes, refs, withTables)
	return stmt, selectRefs, ok
}

func (s *Smither) makeSelectClause(
	desiredTypes []*types.T, refs colRefs, withTables tableRefs,
) (clause *tree.SelectClause, selectRefs, orderByRefs colRefs, ok bool) {
	__antithesis_instrumentation__.Notify(69129)
	if desiredTypes == nil && func() bool {
		__antithesis_instrumentation__.Notify(69134)
		return s.d9() == 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69135)
		return s.makeOrderedAggregate()
	} else {
		__antithesis_instrumentation__.Notify(69136)
	}
	__antithesis_instrumentation__.Notify(69130)

	clause = &tree.SelectClause{}

	var fromRefs colRefs

	requireFrom := s.d6() != 1
	for (requireFrom && func() bool {
		__antithesis_instrumentation__.Notify(69137)
		return len(clause.From.Tables) < 1 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(69138)
		return s.canRecurse() == true
	}() == true {
		__antithesis_instrumentation__.Notify(69139)
		var from tree.TableExpr
		if len(withTables) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(69141)
			return s.coin() == true
		}() == true {
			__antithesis_instrumentation__.Notify(69142)

			source, sourceRefs, sourceOk := makeTableExpr(s, refs, false)
			if !sourceOk {
				__antithesis_instrumentation__.Notify(69144)
				return nil, nil, nil, false
			} else {
				__antithesis_instrumentation__.Notify(69145)
			}
			__antithesis_instrumentation__.Notify(69143)
			from = source
			fromRefs = append(fromRefs, sourceRefs...)
		} else {
			__antithesis_instrumentation__.Notify(69146)

			table := withTables[s.rnd.Intn(len(withTables))]

			alias := s.name("cte_ref")
			name := tree.NewUnqualifiedTableName(alias)
			expr, exprRefs := s.tableExpr(table, name)

			from = &tree.AliasedTableExpr{
				Expr: expr,
				As:   tree.AliasClause{Alias: alias},
			}
			fromRefs = append(fromRefs, exprRefs...)
		}
		__antithesis_instrumentation__.Notify(69140)
		clause.From.Tables = append(clause.From.Tables, from)

		if len(clause.From.Tables) >= 4 {
			__antithesis_instrumentation__.Notify(69147)
			break
		} else {
			__antithesis_instrumentation__.Notify(69148)
		}
	}
	__antithesis_instrumentation__.Notify(69131)

	selectListRefs := refs
	ctx := emptyCtx

	if len(clause.From.Tables) > 0 {
		__antithesis_instrumentation__.Notify(69149)
		clause.Where = s.makeWhere(fromRefs)
		orderByRefs = fromRefs
		selectListRefs = selectListRefs.extend(fromRefs...)

		if s.d6() <= 2 && func() bool {
			__antithesis_instrumentation__.Notify(69150)
			return s.canRecurse() == true
		}() == true {
			__antithesis_instrumentation__.Notify(69151)

			var groupByRefs colRefs
			for _, r := range fromRefs {
				__antithesis_instrumentation__.Notify(69155)
				if s.postgres && func() bool {
					__antithesis_instrumentation__.Notify(69157)
					return r.typ.Family() == types.Box2DFamily == true
				}() == true {
					__antithesis_instrumentation__.Notify(69158)
					continue
				} else {
					__antithesis_instrumentation__.Notify(69159)
				}
				__antithesis_instrumentation__.Notify(69156)
				groupByRefs = append(groupByRefs, r)
			}
			__antithesis_instrumentation__.Notify(69152)
			s.rnd.Shuffle(len(groupByRefs), func(i, j int) {
				__antithesis_instrumentation__.Notify(69160)
				groupByRefs[i], groupByRefs[j] = groupByRefs[j], groupByRefs[i]
			})
			__antithesis_instrumentation__.Notify(69153)
			var groupBy tree.GroupBy
			for (len(groupBy) < 1 || func() bool {
				__antithesis_instrumentation__.Notify(69161)
				return s.coin() == true
			}() == true) && func() bool {
				__antithesis_instrumentation__.Notify(69162)
				return len(groupBy) < len(groupByRefs) == true
			}() == true {
				__antithesis_instrumentation__.Notify(69163)
				groupBy = append(groupBy, groupByRefs[len(groupBy)].item)
			}
			__antithesis_instrumentation__.Notify(69154)
			groupByRefs = groupByRefs[:len(groupBy)]
			clause.GroupBy = groupBy
			clause.Having = s.makeHaving(fromRefs)
			selectListRefs = groupByRefs
			orderByRefs = groupByRefs

			ctx = groupByCtx
		} else {
			__antithesis_instrumentation__.Notify(69164)
			if s.d6() <= 1 && func() bool {
				__antithesis_instrumentation__.Notify(69165)
				return s.canRecurse() == true
			}() == true {
				__antithesis_instrumentation__.Notify(69166)

				ctx = windowCtx
			} else {
				__antithesis_instrumentation__.Notify(69167)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(69168)
	}
	__antithesis_instrumentation__.Notify(69132)

	selectList, selectRefs, ok := s.makeSelectList(ctx, desiredTypes, selectListRefs)
	if !ok {
		__antithesis_instrumentation__.Notify(69169)
		return nil, nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69170)
	}
	__antithesis_instrumentation__.Notify(69133)
	clause.Exprs = selectList

	return clause, selectRefs, orderByRefs, true
}

func (s *Smither) makeOrderedAggregate() (
	clause *tree.SelectClause,
	selectRefs, orderByRefs colRefs,
	ok bool,
) {
	__antithesis_instrumentation__.Notify(69171)

	tableExpr, tableAlias, table, tableColRefs, ok := s.getSchemaTableWithIndexHint()
	if !ok {
		__antithesis_instrumentation__.Notify(69175)
		return nil, nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69176)
	}
	__antithesis_instrumentation__.Notify(69172)
	_, _, idxRefs, ok := s.getRandTableIndex(*table.TableName, *tableAlias)
	if !ok {
		__antithesis_instrumentation__.Notify(69177)
		return nil, nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69178)
	}
	__antithesis_instrumentation__.Notify(69173)

	var groupBy tree.GroupBy
	for (len(groupBy) < 1 || func() bool {
		__antithesis_instrumentation__.Notify(69179)
		return s.coin() == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(69180)
		return len(groupBy) < len(idxRefs) == true
	}() == true {
		__antithesis_instrumentation__.Notify(69181)
		groupBy = append(groupBy, idxRefs[len(groupBy)].item)
	}
	__antithesis_instrumentation__.Notify(69174)
	idxRefs = idxRefs[:len(groupBy)]
	alias := s.name("col")
	selectRefs = colRefs{
		{
			typ:  types.Int,
			item: &tree.ColumnItem{ColumnName: alias},
		},
	}
	return &tree.SelectClause{
		Exprs: tree.SelectExprs{
			{
				Expr: countStar,
				As:   tree.UnrestrictedName(alias),
			},
		},
		From: tree.From{
			Tables: tree.TableExprs{tableExpr},
		},
		Where:   s.makeWhere(tableColRefs),
		GroupBy: groupBy,
		Having:  s.makeHaving(idxRefs),
	}, selectRefs, idxRefs, true
}

var countStar = func() tree.TypedExpr {
	__antithesis_instrumentation__.Notify(69182)
	fn := tree.FunDefs["count"]
	typ := types.Int
	return tree.NewTypedFuncExpr(
		tree.ResolvableFunctionReference{FunctionReference: fn},
		0,
		tree.TypedExprs{tree.UnqualifiedStar{}},
		nil,
		nil,
		typ,
		&fn.FunctionProperties,
		fn.Definition[0].(*tree.Overload),
	)
}()

func makeSelect(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69183)
	stmt, refs, ok := s.makeSelect(nil, nil)
	if !ok {
		__antithesis_instrumentation__.Notify(69186)
		return stmt, ok
	} else {
		__antithesis_instrumentation__.Notify(69187)
	}
	__antithesis_instrumentation__.Notify(69184)
	if s.outputSort {
		__antithesis_instrumentation__.Notify(69188)
		order := make(tree.OrderBy, len(refs))
		for i, r := range refs {
			__antithesis_instrumentation__.Notify(69190)
			var expr tree.Expr = r.item

			if s.postgres && func() bool {
				__antithesis_instrumentation__.Notify(69192)
				return r.typ.Family() == types.Box2DFamily == true
			}() == true {
				__antithesis_instrumentation__.Notify(69193)
				expr = &tree.CastExpr{Expr: r.item, Type: types.String}
			} else {
				__antithesis_instrumentation__.Notify(69194)
			}
			__antithesis_instrumentation__.Notify(69191)
			order[i] = &tree.Order{
				Expr:       expr,
				NullsOrder: tree.NullsFirst,
			}
		}
		__antithesis_instrumentation__.Notify(69189)
		stmt = &tree.Select{
			Select: &tree.SelectClause{
				Exprs: tree.SelectExprs{
					tree.SelectExpr{
						Expr: tree.UnqualifiedStar{},
					},
				},
				From: tree.From{
					Tables: tree.TableExprs{
						&tree.AliasedTableExpr{
							Expr: &tree.Subquery{
								Select: &tree.ParenSelect{Select: stmt},
							},
							As: tree.AliasClause{
								Alias: s.name("tab"),
							},
						},
					},
				},
			},
			OrderBy: order,
		}
	} else {
		__antithesis_instrumentation__.Notify(69195)
	}
	__antithesis_instrumentation__.Notify(69185)
	return stmt, ok
}

func (s *Smither) makeSelect(desiredTypes []*types.T, refs colRefs) (*tree.Select, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69196)
	withStmt, withTables := s.makeWith()

	clause, selectRefs, orderByRefs, ok := s.makeSelectClause(desiredTypes, refs, withTables)
	if !ok {
		__antithesis_instrumentation__.Notify(69198)
		return nil, nil, ok
	} else {
		__antithesis_instrumentation__.Notify(69199)
	}
	__antithesis_instrumentation__.Notify(69197)

	stmt := tree.Select{
		Select:  clause,
		With:    withStmt,
		OrderBy: s.makeOrderBy(orderByRefs),
		Limit:   makeLimit(s),
	}

	return &stmt, selectRefs, true
}

func (s *Smither) makeSelectList(
	ctx Context, desiredTypes []*types.T, refs colRefs,
) (tree.SelectExprs, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69200)

	if len(desiredTypes) == 0 {
		__antithesis_instrumentation__.Notify(69204)
		s.sample(len(refs), 6, func(i int) {
			__antithesis_instrumentation__.Notify(69205)
			desiredTypes = append(desiredTypes, refs[i].typ)
		})
	} else {
		__antithesis_instrumentation__.Notify(69206)
	}
	__antithesis_instrumentation__.Notify(69201)

	if len(desiredTypes) == 0 {
		__antithesis_instrumentation__.Notify(69207)
		desiredTypes = s.makeDesiredTypes()
	} else {
		__antithesis_instrumentation__.Notify(69208)
	}
	__antithesis_instrumentation__.Notify(69202)
	result := make(tree.SelectExprs, len(desiredTypes))
	selectRefs := make(colRefs, len(desiredTypes))
	for i, t := range desiredTypes {
		__antithesis_instrumentation__.Notify(69209)
		result[i].Expr = makeScalarContext(s, ctx, t, refs)
		alias := s.name("col")
		result[i].As = tree.UnrestrictedName(alias)
		selectRefs[i] = &colRef{
			typ:  t,
			item: &tree.ColumnItem{ColumnName: alias},
		}
	}
	__antithesis_instrumentation__.Notify(69203)
	return result, selectRefs, true
}

func makeDelete(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69210)
	stmt, _, ok := s.makeDelete(nil)
	return stmt, ok
}

func (s *Smither) makeDelete(refs colRefs) (*tree.Delete, *tableRef, bool) {
	__antithesis_instrumentation__.Notify(69211)
	table, _, tableRef, tableRefs, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69214)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69215)
	}
	__antithesis_instrumentation__.Notify(69212)

	del := &tree.Delete{
		Table:     table,
		Where:     s.makeWhere(tableRefs),
		OrderBy:   s.makeOrderBy(tableRefs),
		Limit:     makeLimit(s),
		Returning: &tree.NoReturningClause{},
	}
	if del.Limit == nil {
		__antithesis_instrumentation__.Notify(69216)
		del.OrderBy = nil
	} else {
		__antithesis_instrumentation__.Notify(69217)
	}
	__antithesis_instrumentation__.Notify(69213)

	return del, tableRef, true
}

func makeDeleteReturning(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69218)
	if forJoin {
		__antithesis_instrumentation__.Notify(69220)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69221)
	}
	__antithesis_instrumentation__.Notify(69219)
	return s.makeDeleteReturning(refs)
}

func (s *Smither) makeDeleteReturning(refs colRefs) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69222)
	del, delRef, ok := s.makeDelete(refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69224)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69225)
	}
	__antithesis_instrumentation__.Notify(69223)
	var returningRefs colRefs
	del.Returning, returningRefs = s.makeReturning(delRef)
	return &tree.StatementSource{
		Statement: del,
	}, returningRefs, true
}

func makeUpdate(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69226)
	stmt, _, ok := s.makeUpdate(nil)
	return stmt, ok
}

func (s *Smither) makeUpdate(refs colRefs) (*tree.Update, *tableRef, bool) {
	__antithesis_instrumentation__.Notify(69227)
	table, _, tableRef, tableRefs, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69233)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69234)
	}
	__antithesis_instrumentation__.Notify(69228)
	cols := make(map[tree.Name]*tree.ColumnTableDef)
	for _, c := range tableRef.Columns {
		__antithesis_instrumentation__.Notify(69235)
		cols[c.Name] = c
	}
	__antithesis_instrumentation__.Notify(69229)

	update := &tree.Update{
		Table:     table,
		Where:     s.makeWhere(tableRefs),
		OrderBy:   s.makeOrderBy(tableRefs),
		Limit:     makeLimit(s),
		Returning: &tree.NoReturningClause{},
	}

	upRefs := tableRefs.extend()
	for (len(update.Exprs) < 1 || func() bool {
		__antithesis_instrumentation__.Notify(69236)
		return s.coin() == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(69237)
		return len(upRefs) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69238)
		n := s.rnd.Intn(len(upRefs))
		ref := upRefs[n]
		upRefs = append(upRefs[:n], upRefs[n+1:]...)
		col := cols[ref.item.ColumnName]

		if col == nil || func() bool {
			__antithesis_instrumentation__.Notify(69241)
			return col.Computed.Computed == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(69242)
			return col.Hidden == true
		}() == true {
			__antithesis_instrumentation__.Notify(69243)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69244)
		}
		__antithesis_instrumentation__.Notify(69239)
		var expr tree.TypedExpr
		for {
			__antithesis_instrumentation__.Notify(69245)
			expr = makeScalar(s, ref.typ, tableRefs)

			if col.Nullable.Nullability != tree.NotNull || func() bool {
				__antithesis_instrumentation__.Notify(69246)
				return expr != tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(69247)
				break
			} else {
				__antithesis_instrumentation__.Notify(69248)
			}
		}
		__antithesis_instrumentation__.Notify(69240)
		update.Exprs = append(update.Exprs, &tree.UpdateExpr{
			Names: tree.NameList{ref.item.ColumnName},
			Expr:  expr,
		})
	}
	__antithesis_instrumentation__.Notify(69230)
	if len(update.Exprs) == 0 {
		__antithesis_instrumentation__.Notify(69249)
		panic("empty")
	} else {
		__antithesis_instrumentation__.Notify(69250)
	}
	__antithesis_instrumentation__.Notify(69231)
	if update.Limit == nil {
		__antithesis_instrumentation__.Notify(69251)
		update.OrderBy = nil
	} else {
		__antithesis_instrumentation__.Notify(69252)
	}
	__antithesis_instrumentation__.Notify(69232)

	return update, tableRef, true
}

func makeUpdateReturning(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69253)
	if forJoin {
		__antithesis_instrumentation__.Notify(69255)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69256)
	}
	__antithesis_instrumentation__.Notify(69254)
	return s.makeUpdateReturning(refs)
}

func (s *Smither) makeUpdateReturning(refs colRefs) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69257)
	update, updateRef, ok := s.makeUpdate(refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69259)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69260)
	}
	__antithesis_instrumentation__.Notify(69258)
	var returningRefs colRefs
	update.Returning, returningRefs = s.makeReturning(updateRef)
	return &tree.StatementSource{
		Statement: update,
	}, returningRefs, true
}

func makeInsert(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69261)
	stmt, _, ok := s.makeInsert(nil)
	return stmt, ok
}

func makeBegin(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69262)
	return &tree.BeginTransaction{}, true
}

const letters = "abcdefghijklmnopqrstuvwxyz"

func makeSavepoint(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69263)
	savepointName := randgen.RandString(s.rnd, s.d9(), letters)
	s.activeSavepoints = append(s.activeSavepoints, savepointName)
	return &tree.Savepoint{Name: tree.Name(savepointName)}, true
}

func makeReleaseSavepoint(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69264)
	if len(s.activeSavepoints) == 0 {
		__antithesis_instrumentation__.Notify(69266)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69267)
	}
	__antithesis_instrumentation__.Notify(69265)
	idx := s.rnd.Intn(len(s.activeSavepoints))
	savepoint := s.activeSavepoints[idx]

	s.activeSavepoints = append(s.activeSavepoints[:idx], s.activeSavepoints[idx+1:]...)
	return &tree.ReleaseSavepoint{Savepoint: tree.Name(savepoint)}, true
}

func makeRollbackToSavepoint(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69268)
	if len(s.activeSavepoints) == 0 {
		__antithesis_instrumentation__.Notify(69270)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(69271)
	}
	__antithesis_instrumentation__.Notify(69269)
	idx := s.rnd.Intn(len(s.activeSavepoints))
	savepoint := s.activeSavepoints[idx]

	s.activeSavepoints = s.activeSavepoints[:idx+1]
	return &tree.RollbackToSavepoint{Savepoint: tree.Name(savepoint)}, true
}

func makeCommit(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69272)
	return &tree.CommitTransaction{}, true
}

func makeRollback(s *Smither) (tree.Statement, bool) {
	__antithesis_instrumentation__.Notify(69273)
	return &tree.RollbackTransaction{}, true
}

func (s *Smither) makeInsert(refs colRefs) (*tree.Insert, *tableRef, bool) {
	__antithesis_instrumentation__.Notify(69274)
	table, _, tableRef, _, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69282)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69283)
	}
	__antithesis_instrumentation__.Notify(69275)

	insert := &tree.Insert{
		Table:     table,
		Rows:      &tree.Select{},
		Returning: &tree.NoReturningClause{},
	}

	if s.d9() == 1 {
		__antithesis_instrumentation__.Notify(69284)
		return insert, tableRef, true
	} else {
		__antithesis_instrumentation__.Notify(69285)
	}
	__antithesis_instrumentation__.Notify(69276)

	if s.coin() {
		__antithesis_instrumentation__.Notify(69286)
		var names tree.NameList
		var row tree.Exprs
		for _, c := range tableRef.Columns {
			__antithesis_instrumentation__.Notify(69289)

			if c.Computed.Computed || func() bool {
				__antithesis_instrumentation__.Notify(69291)
				return c.Hidden == true
			}() == true {
				__antithesis_instrumentation__.Notify(69292)
				continue
			} else {
				__antithesis_instrumentation__.Notify(69293)
			}
			__antithesis_instrumentation__.Notify(69290)
			notNull := c.Nullable.Nullability == tree.NotNull
			if notNull || func() bool {
				__antithesis_instrumentation__.Notify(69294)
				return s.coin() == true
			}() == true {
				__antithesis_instrumentation__.Notify(69295)
				names = append(names, c.Name)
				row = append(row, randgen.RandDatum(s.rnd, tree.MustBeStaticallyKnownType(c.Type), !notNull))
			} else {
				__antithesis_instrumentation__.Notify(69296)
			}
		}
		__antithesis_instrumentation__.Notify(69287)

		if len(names) == 0 {
			__antithesis_instrumentation__.Notify(69297)
			return nil, nil, false
		} else {
			__antithesis_instrumentation__.Notify(69298)
		}
		__antithesis_instrumentation__.Notify(69288)

		insert.Columns = names
		insert.Rows = &tree.Select{
			Select: &tree.ValuesClause{
				Rows: []tree.Exprs{row},
			},
		}
		return insert, tableRef, true
	} else {
		__antithesis_instrumentation__.Notify(69299)
	}
	__antithesis_instrumentation__.Notify(69277)

	var desiredTypes []*types.T
	var names tree.NameList
	unnamed := s.coin()
	for _, c := range tableRef.Columns {
		__antithesis_instrumentation__.Notify(69300)

		if c.Computed.Computed || func() bool {
			__antithesis_instrumentation__.Notify(69302)
			return c.Hidden == true
		}() == true {
			__antithesis_instrumentation__.Notify(69303)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69304)
		}
		__antithesis_instrumentation__.Notify(69301)
		if unnamed || func() bool {
			__antithesis_instrumentation__.Notify(69305)
			return c.Nullable.Nullability == tree.NotNull == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(69306)
			return s.coin() == true
		}() == true {
			__antithesis_instrumentation__.Notify(69307)
			names = append(names, c.Name)
			desiredTypes = append(desiredTypes, tree.MustBeStaticallyKnownType(c.Type))
		} else {
			__antithesis_instrumentation__.Notify(69308)
		}
	}
	__antithesis_instrumentation__.Notify(69278)

	if len(desiredTypes) == 0 {
		__antithesis_instrumentation__.Notify(69309)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69310)
	}
	__antithesis_instrumentation__.Notify(69279)

	if !unnamed {
		__antithesis_instrumentation__.Notify(69311)
		insert.Columns = names
	} else {
		__antithesis_instrumentation__.Notify(69312)
	}
	__antithesis_instrumentation__.Notify(69280)

	insert.Rows, _, ok = s.makeSelect(desiredTypes, refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69313)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69314)
	}
	__antithesis_instrumentation__.Notify(69281)

	return insert, tableRef, true
}

func makeInsertReturning(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69315)
	if forJoin {
		__antithesis_instrumentation__.Notify(69317)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69318)
	}
	__antithesis_instrumentation__.Notify(69316)
	return s.makeInsertReturning(refs)
}

func (s *Smither) makeInsertReturning(refs colRefs) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69319)
	insert, insertRef, ok := s.makeInsert(refs)
	if !ok {
		__antithesis_instrumentation__.Notify(69321)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69322)
	}
	__antithesis_instrumentation__.Notify(69320)
	var returningRefs colRefs
	insert.Returning, returningRefs = s.makeReturning(insertRef)
	return &tree.StatementSource{
		Statement: insert,
	}, returningRefs, true
}

func makeValuesTable(s *Smither, refs colRefs, forJoin bool) (tree.TableExpr, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69323)
	types := s.makeDesiredTypes()
	values, valuesRefs := makeValues(s, types, refs)
	return values, valuesRefs, true
}

func makeValues(s *Smither, desiredTypes []*types.T, refs colRefs) (tree.TableExpr, colRefs) {
	__antithesis_instrumentation__.Notify(69324)
	numRowsToInsert := s.d6()
	values := tree.ValuesClause{
		Rows: make([]tree.Exprs, numRowsToInsert),
	}

	for i := 0; i < numRowsToInsert; i++ {
		__antithesis_instrumentation__.Notify(69327)
		tuple := make([]tree.Expr, len(desiredTypes))
		for j, t := range desiredTypes {
			__antithesis_instrumentation__.Notify(69329)
			tuple[j] = makeScalar(s, t, refs)
		}
		__antithesis_instrumentation__.Notify(69328)
		values.Rows[i] = tuple
	}
	__antithesis_instrumentation__.Notify(69325)
	table := s.name("tab")
	names := make(tree.NameList, len(desiredTypes))
	valuesRefs := make(colRefs, len(desiredTypes))
	for i, typ := range desiredTypes {
		__antithesis_instrumentation__.Notify(69330)
		names[i] = s.name("col")
		valuesRefs[i] = &colRef{
			typ: typ,
			item: tree.NewColumnItem(
				tree.NewUnqualifiedTableName(table),
				names[i],
			),
		}
	}
	__antithesis_instrumentation__.Notify(69326)

	return &tree.AliasedTableExpr{
		Expr: &tree.Subquery{
			Select: &tree.ParenSelect{
				Select: &tree.Select{
					Select: &values,
				},
			},
		},
		As: tree.AliasClause{
			Alias: table,
			Cols:  names,
		},
	}, valuesRefs
}

func makeValuesSelect(
	s *Smither, desiredTypes []*types.T, refs colRefs, withTables tableRefs,
) (tree.SelectStatement, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69331)
	values, valuesRefs := makeValues(s, desiredTypes, refs)

	return &tree.SelectClause{
		Exprs: tree.SelectExprs{tree.StarSelectExpr()},
		From: tree.From{
			Tables: tree.TableExprs{values},
		},
	}, valuesRefs, true
}

var setOps = []tree.UnionType{
	tree.UnionOp,
	tree.IntersectOp,
	tree.ExceptOp,
}

func makeSetOp(
	s *Smither, desiredTypes []*types.T, refs colRefs, withTables tableRefs,
) (tree.SelectStatement, colRefs, bool) {
	__antithesis_instrumentation__.Notify(69332)
	left, leftRefs, ok := s.makeSelectStmt(desiredTypes, refs, withTables)
	if !ok {
		__antithesis_instrumentation__.Notify(69335)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69336)
	}
	__antithesis_instrumentation__.Notify(69333)

	right, _, ok := s.makeSelectStmt(desiredTypes, refs, withTables)
	if !ok {
		__antithesis_instrumentation__.Notify(69337)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(69338)
	}
	__antithesis_instrumentation__.Notify(69334)

	return &tree.UnionClause{
		Type:  setOps[s.rnd.Intn(len(setOps))],
		Left:  &tree.Select{Select: left},
		Right: &tree.Select{Select: right},
		All:   s.coin(),
	}, leftRefs, true
}

func (s *Smither) makeWhere(refs colRefs) *tree.Where {
	__antithesis_instrumentation__.Notify(69339)
	if s.coin() {
		__antithesis_instrumentation__.Notify(69341)
		where := makeBoolExpr(s, refs)
		return tree.NewWhere("WHERE", where)
	} else {
		__antithesis_instrumentation__.Notify(69342)
	}
	__antithesis_instrumentation__.Notify(69340)
	return nil
}

func (s *Smither) makeHaving(refs colRefs) *tree.Where {
	__antithesis_instrumentation__.Notify(69343)
	if s.coin() {
		__antithesis_instrumentation__.Notify(69345)
		where := makeBoolExprContext(s, havingCtx, refs)
		return tree.NewWhere("HAVING", where)
	} else {
		__antithesis_instrumentation__.Notify(69346)
	}
	__antithesis_instrumentation__.Notify(69344)
	return nil
}

func (s *Smither) makeOrderBy(refs colRefs) tree.OrderBy {
	__antithesis_instrumentation__.Notify(69347)
	if len(refs) == 0 {
		__antithesis_instrumentation__.Notify(69350)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(69351)
	}
	__antithesis_instrumentation__.Notify(69348)
	var ob tree.OrderBy
	for s.coin() {
		__antithesis_instrumentation__.Notify(69352)
		ref := refs[s.rnd.Intn(len(refs))]

		if ref.typ.Family() == types.JsonFamily {
			__antithesis_instrumentation__.Notify(69355)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69356)
		}
		__antithesis_instrumentation__.Notify(69353)

		if s.postgres && func() bool {
			__antithesis_instrumentation__.Notify(69357)
			return ref.typ.Family() == types.Box2DFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(69358)
			continue
		} else {
			__antithesis_instrumentation__.Notify(69359)
		}
		__antithesis_instrumentation__.Notify(69354)
		ob = append(ob, &tree.Order{
			Expr:      ref.item,
			Direction: s.randDirection(),
		})
	}
	__antithesis_instrumentation__.Notify(69349)
	return ob
}

func makeLimit(s *Smither) *tree.Limit {
	__antithesis_instrumentation__.Notify(69360)
	if s.disableLimits {
		__antithesis_instrumentation__.Notify(69363)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(69364)
	}
	__antithesis_instrumentation__.Notify(69361)
	if s.d6() > 2 {
		__antithesis_instrumentation__.Notify(69365)
		return &tree.Limit{Count: tree.NewDInt(tree.DInt(s.d100()))}
	} else {
		__antithesis_instrumentation__.Notify(69366)
	}
	__antithesis_instrumentation__.Notify(69362)
	return nil
}

func (s *Smither) makeReturning(table *tableRef) (*tree.ReturningExprs, colRefs) {
	__antithesis_instrumentation__.Notify(69367)
	desiredTypes := s.makeDesiredTypes()

	refs := make(colRefs, len(table.Columns))
	for i, c := range table.Columns {
		__antithesis_instrumentation__.Notify(69370)
		refs[i] = &colRef{
			typ:  tree.MustBeStaticallyKnownType(c.Type),
			item: &tree.ColumnItem{ColumnName: c.Name},
		}
	}
	__antithesis_instrumentation__.Notify(69368)

	returning := make(tree.ReturningExprs, len(desiredTypes))
	returningRefs := make(colRefs, len(desiredTypes))
	for i, t := range desiredTypes {
		__antithesis_instrumentation__.Notify(69371)
		returning[i].Expr = makeScalar(s, t, refs)
		alias := s.name("ret")
		returning[i].As = tree.UnrestrictedName(alias)
		returningRefs[i] = &colRef{
			typ: t,
			item: &tree.ColumnItem{
				ColumnName: alias,
			},
		}
	}
	__antithesis_instrumentation__.Notify(69369)
	return &returning, returningRefs
}
