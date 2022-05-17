package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func (s *Smither) GenerateTLP() (unpartitioned, partitioned string, args []interface{}) {
	__antithesis_instrumentation__.Notify(69935)

	originalDisableImpureFns := s.disableImpureFns
	s.disableImpureFns = true
	defer func() {
		__antithesis_instrumentation__.Notify(69938)
		s.disableImpureFns = originalDisableImpureFns
	}()
	__antithesis_instrumentation__.Notify(69936)

	switch tlpType := s.rnd.Intn(5); tlpType {
	case 0:
		__antithesis_instrumentation__.Notify(69939)
		partitioned, unpartitioned, args = s.generateWhereTLP()
	case 1:
		__antithesis_instrumentation__.Notify(69940)
		partitioned, unpartitioned = s.generateOuterJoinTLP()
	case 2:
		__antithesis_instrumentation__.Notify(69941)
		partitioned, unpartitioned = s.generateInnerJoinTLP()
	case 3:
		__antithesis_instrumentation__.Notify(69942)
		partitioned, unpartitioned = s.generateDistinctTLP()
	default:
		__antithesis_instrumentation__.Notify(69943)
		partitioned, unpartitioned = s.generateAggregationTLP()
	}
	__antithesis_instrumentation__.Notify(69937)
	return partitioned, unpartitioned, args
}

func (s *Smither) generateWhereTLP() (unpartitioned, partitioned string, args []interface{}) {
	__antithesis_instrumentation__.Notify(69944)
	f := tree.NewFmtCtx(tree.FmtParsable)

	table, _, _, cols, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69947)
		panic(errors.AssertionFailedf("failed to find random table"))
	} else {
		__antithesis_instrumentation__.Notify(69948)
	}
	__antithesis_instrumentation__.Notify(69945)
	table.Format(f)
	tableName := f.CloseAndGetString()

	var pred tree.Expr
	if s.coin() {
		__antithesis_instrumentation__.Notify(69949)
		pred = makeBoolExpr(s, cols)
	} else {
		__antithesis_instrumentation__.Notify(69950)
		pred, args = makeBoolExprWithPlaceholders(s, cols)
	}
	__antithesis_instrumentation__.Notify(69946)
	pred.Format(f)
	predicate := f.CloseAndGetString()

	allPreds := fmt.Sprintf("%[1]s, NOT (%[1]s), (%[1]s) IS NULL", predicate)

	unpartitioned = fmt.Sprintf("SELECT *, %s, true, false, false FROM %s", allPreds, tableName)

	pred1 := predicate
	pred2 := fmt.Sprintf("NOT (%s)", predicate)
	pred3 := fmt.Sprintf("(%s) IS NULL", predicate)

	part1 := fmt.Sprintf(`SELECT *,
%s,
%s, %s, %s
FROM %s
WHERE %s`, allPreds, pred1, pred2, pred3, tableName, pred1)
	part2 := fmt.Sprintf(`SELECT *,
%s,
%s, %s, %s
FROM %s
WHERE %s`, allPreds, pred2, pred1, pred3, tableName, pred2)
	part3 := fmt.Sprintf(`SELECT *,
%s,
%s, (%s) IS NOT NULL, (%s) IS NOT NULL
FROM %s
WHERE %s`, allPreds, pred3, pred1, pred2, tableName, pred3)

	partitioned = fmt.Sprintf(
		`(%s)
UNION ALL (%s)
UNION ALL (%s)`,
		part1, part2, part3,
	)

	return unpartitioned, partitioned, args
}

func (s *Smither) generateOuterJoinTLP() (unpartitioned, partitioned string) {
	__antithesis_instrumentation__.Notify(69951)
	f := tree.NewFmtCtx(tree.FmtParsable)

	table1, _, _, cols1, ok1 := s.getSchemaTable()
	table2, _, _, _, ok2 := s.getSchemaTable()
	if !ok1 || func() bool {
		__antithesis_instrumentation__.Notify(69953)
		return !ok2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69954)
		panic(errors.AssertionFailedf("failed to find random tables"))
	} else {
		__antithesis_instrumentation__.Notify(69955)
	}
	__antithesis_instrumentation__.Notify(69952)
	table1.Format(f)
	tableName1 := f.CloseAndGetString()
	table2.Format(f)
	tableName2 := f.CloseAndGetString()

	leftJoinTrue := fmt.Sprintf(
		"SELECT * FROM %s LEFT JOIN %s ON TRUE",
		tableName1, tableName2,
	)
	leftJoinFalse := fmt.Sprintf(
		"SELECT * FROM %s LEFT JOIN %s ON FALSE",
		tableName1, tableName2,
	)

	unpartitioned = fmt.Sprintf(
		"(%s) UNION ALL (%s) UNION ALL (%s)",
		leftJoinTrue, leftJoinFalse, leftJoinFalse,
	)

	pred := makeBoolExpr(s, cols1)
	pred.Format(f)
	predicate := f.CloseAndGetString()

	part1 := fmt.Sprintf(
		"SELECT * FROM %s LEFT JOIN %s ON %s",
		tableName1, tableName2, predicate,
	)
	part2 := fmt.Sprintf(
		"SELECT * FROM %s LEFT JOIN %s ON NOT (%s)",
		tableName1, tableName2, predicate,
	)
	part3 := fmt.Sprintf(
		"SELECT * FROM %s LEFT JOIN %s ON (%s) IS NULL",
		tableName1, tableName2, predicate,
	)

	partitioned = fmt.Sprintf(
		"(%s) UNION ALL (%s) UNION ALL (%s)",
		part1, part2, part3,
	)

	return unpartitioned, partitioned
}

func (s *Smither) generateInnerJoinTLP() (unpartitioned, partitioned string) {
	__antithesis_instrumentation__.Notify(69956)
	f := tree.NewFmtCtx(tree.FmtParsable)

	table1, _, _, cols1, ok1 := s.getSchemaTable()
	table2, _, _, cols2, ok2 := s.getSchemaTable()
	if !ok1 || func() bool {
		__antithesis_instrumentation__.Notify(69958)
		return !ok2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(69959)
		panic(errors.AssertionFailedf("failed to find random tables"))
	} else {
		__antithesis_instrumentation__.Notify(69960)
	}
	__antithesis_instrumentation__.Notify(69957)
	table1.Format(f)
	tableName1 := f.CloseAndGetString()
	table2.Format(f)
	tableName2 := f.CloseAndGetString()

	unpartitioned = fmt.Sprintf(
		"SELECT * FROM %s JOIN %s ON true",
		tableName1, tableName2,
	)

	cols := cols1.extend(cols2...)
	pred := makeBoolExpr(s, cols)
	pred.Format(f)
	predicate := f.CloseAndGetString()

	part1 := fmt.Sprintf(
		"SELECT * FROM %s JOIN %s ON %s",
		tableName1, tableName2, predicate,
	)
	part2 := fmt.Sprintf(
		"SELECT * FROM %s JOIN %s ON NOT (%s)",
		tableName1, tableName2, predicate,
	)
	part3 := fmt.Sprintf(
		"SELECT * FROM %s JOIN %s ON (%s) IS NULL",
		tableName1, tableName2, predicate,
	)

	partitioned = fmt.Sprintf(
		"(%s) UNION ALL (%s) UNION ALL (%s)",
		part1, part2, part3,
	)

	return unpartitioned, partitioned
}

func (s *Smither) generateAggregationTLP() (unpartitioned, partitioned string) {
	__antithesis_instrumentation__.Notify(69961)
	f := tree.NewFmtCtx(tree.FmtParsable)

	table, _, _, cols, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69964)
		panic(errors.AssertionFailedf("failed to find random table"))
	} else {
		__antithesis_instrumentation__.Notify(69965)
	}
	__antithesis_instrumentation__.Notify(69962)
	table.Format(f)
	tableName := f.CloseAndGetString()
	tableNameAlias := strings.TrimSpace(strings.Split(tableName, "AS")[1])

	var innerAgg, outerAgg string
	switch aggType := s.rnd.Intn(3); aggType {
	case 0:
		__antithesis_instrumentation__.Notify(69966)
		innerAgg, outerAgg = "MAX", "MAX"
	case 1:
		__antithesis_instrumentation__.Notify(69967)
		innerAgg, outerAgg = "MIN", "MIN"
	default:
		__antithesis_instrumentation__.Notify(69968)
		innerAgg, outerAgg = "COUNT", "SUM"
	}
	__antithesis_instrumentation__.Notify(69963)

	unpartitioned = fmt.Sprintf(
		"SELECT %s(first) FROM (SELECT * FROM %s) %s(first)",
		innerAgg, tableName, tableNameAlias,
	)

	pred := makeBoolExpr(s, cols)
	pred.Format(f)
	predicate := f.CloseAndGetString()

	part1 := fmt.Sprintf(
		"SELECT %s(first) AS agg FROM (SELECT * FROM %s WHERE %s) %s(first)",
		innerAgg, tableName, predicate, tableNameAlias,
	)
	part2 := fmt.Sprintf(
		"SELECT %s(first) AS agg FROM (SELECT * FROM %s WHERE NOT (%s)) %s(first)",
		innerAgg, tableName, predicate, tableNameAlias,
	)
	part3 := fmt.Sprintf(
		"SELECT %s(first) AS agg FROM (SELECT * FROM %s WHERE (%s) IS NULL) %s(first)",
		innerAgg, tableName, predicate, tableNameAlias,
	)

	partitioned = fmt.Sprintf(
		"SELECT %s(agg) FROM (%s UNION ALL %s UNION ALL %s)",
		outerAgg, part1, part2, part3,
	)

	return unpartitioned, partitioned
}

func (s *Smither) generateDistinctTLP() (unpartitioned, partitioned string) {
	__antithesis_instrumentation__.Notify(69969)
	f := tree.NewFmtCtx(tree.FmtParsable)

	table, _, _, cols, ok := s.getSchemaTable()
	if !ok {
		__antithesis_instrumentation__.Notify(69974)
		panic(errors.AssertionFailedf("failed to find random table"))
	} else {
		__antithesis_instrumentation__.Notify(69975)
	}
	__antithesis_instrumentation__.Notify(69970)
	table.Format(f)
	tableName := f.CloseAndGetString()

	s.rnd.Shuffle(len(cols), func(i, j int) { __antithesis_instrumentation__.Notify(69976); cols[i], cols[j] = cols[j], cols[i] })
	__antithesis_instrumentation__.Notify(69971)
	n := s.rnd.Intn(len(cols))
	if n == 0 {
		__antithesis_instrumentation__.Notify(69977)
		n = 1
	} else {
		__antithesis_instrumentation__.Notify(69978)
	}
	__antithesis_instrumentation__.Notify(69972)
	colStrs := make([]string, n)
	for i, ref := range cols[:n] {
		__antithesis_instrumentation__.Notify(69979)
		colStrs[i] = tree.AsStringWithFlags(ref.typedExpr(), tree.FmtParsable)
	}
	__antithesis_instrumentation__.Notify(69973)
	distinctCols := strings.Join(colStrs, ",")
	unpartitioned = fmt.Sprintf("SELECT DISTINCT %s FROM %s", distinctCols, tableName)

	pred := makeBoolExpr(s, cols)
	pred.Format(f)
	predicate := f.CloseAndGetString()

	part1 := fmt.Sprintf("SELECT DISTINCT %s FROM %s WHERE %s", distinctCols, tableName, predicate)
	part2 := fmt.Sprintf("SELECT DISTINCT %s FROM %s WHERE NOT (%s)", distinctCols, tableName, predicate)
	part3 := fmt.Sprintf("SELECT DISTINCT %s FROM %s WHERE (%s) IS NULL", distinctCols, tableName, predicate)
	partitioned = fmt.Sprintf(
		"(%s) UNION (%s) UNION (%s)", part1, part2, part3,
	)

	return unpartitioned, partitioned
}
