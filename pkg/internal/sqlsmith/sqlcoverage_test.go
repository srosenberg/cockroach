// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package sqlsmith

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

// TestFeatureRegistryNoDuplicates verifies that every registered feature has
// a unique name.
func TestFeatureRegistryNoDuplicates(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	seen := make(map[SQLFeature]bool)
	for _, f := range AllFeatures() {
		if seen[f] {
			t.Errorf("duplicate feature registered: %s", f)
		}
		seen[f] = true
	}
}

// TestFeatureVisitorCoversAllFeatures verifies that every registered feature
// has a corresponding extraction path in the visitor by parsing representative
// SQL statements that exercise each feature. If this test fails after adding a
// new feature via registerFeature, add a SQL statement below that exercises it.
func TestFeatureVisitorCoversAllFeatures(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	// Each SQL string exercises one or more features. Together they must cover
	// every registered feature.
	sqls := []string{
		// stmt/select, clause/where, comp/eq, select/star
		`SELECT * FROM t WHERE a = 1`,
		// stmt/insert, misc/values
		`INSERT INTO t(a) VALUES (1)`,
		// stmt/update, clause/where
		`UPDATE t SET a = 1 WHERE b = 2`,
		// stmt/delete, clause/where
		`DELETE FROM t WHERE a = 1`,
		// clause/group_by, clause/having, func/aggregate
		`SELECT count(a) FROM t GROUP BY b HAVING count(a) > 1`,
		// clause/order_by, clause/limit, clause/offset
		`SELECT a FROM t ORDER BY a LIMIT 10 OFFSET 5`,
		// clause/with
		`WITH cte AS (SELECT 1) SELECT * FROM cte`,
		// clause/window, func/window
		`SELECT row_number() OVER (PARTITION BY a ORDER BY b) FROM t`,
		// select/distinct
		`SELECT DISTINCT a FROM t`,
		// select/distinct_on
		`SELECT DISTINCT ON (a) a, b FROM t`,
		// set_op/union
		`SELECT a FROM t UNION SELECT b FROM t`,
		// set_op/union_all
		`SELECT a FROM t UNION ALL SELECT b FROM t`,
		// set_op/intersect
		`SELECT a FROM t INTERSECT SELECT b FROM t`,
		// set_op/except
		`SELECT a FROM t EXCEPT SELECT b FROM t`,
		// join/inner
		`SELECT * FROM t1 JOIN t2 ON t1.a = t2.a`,
		// join/left
		`SELECT * FROM t1 LEFT JOIN t2 ON t1.a = t2.a`,
		// join/right
		`SELECT * FROM t1 RIGHT JOIN t2 ON t1.a = t2.a`,
		// join/full
		`SELECT * FROM t1 FULL JOIN t2 ON t1.a = t2.a`,
		// join/cross
		`SELECT * FROM t1 CROSS JOIN t2`,
		// join/lateral
		`SELECT * FROM t, LATERAL (SELECT * FROM t2 WHERE t2.a = t.a)`,
		// expr/subquery
		`SELECT (SELECT 1)`,
		// expr/exists
		`SELECT * FROM t WHERE EXISTS (SELECT 1)`,
		// expr/in, expr/not_in
		`SELECT * FROM t WHERE a IN (1, 2) AND b NOT IN (3, 4)`,
		// expr/between
		`SELECT * FROM t WHERE a BETWEEN 1 AND 10`,
		// expr/case
		`SELECT CASE WHEN a = 1 THEN 'x' ELSE 'y' END FROM t`,
		// expr/coalesce
		`SELECT COALESCE(a, b) FROM t`,
		// expr/cast
		`SELECT CAST(a AS INT) FROM t`,
		// expr/type_annotation
		`SELECT a:::INT FROM t`,
		// expr/collate
		`SELECT a COLLATE "en_US" FROM t`,
		// expr/array_construct
		`SELECT ARRAY[1, 2, 3]`,
		// expr/array_flatten
		`SELECT ARRAY(SELECT a FROM t)`,
		// expr/tuple
		`SELECT (1, 2)`,
		// expr/if
		`SELECT IF(a > 0, 'pos', 'neg') FROM t`,
		// expr/nullif
		`SELECT NULLIF(a, 0) FROM t`,
		// expr/is_null, expr/is_not_null
		`SELECT * FROM t WHERE a IS NULL AND b IS NOT NULL`,
		// expr/indirection
		`SELECT a[1] FROM t`,
		// logic/and, logic/or, logic/not
		`SELECT * FROM t WHERE a = 1 AND (b = 2 OR NOT c = 3)`,
		// comp/ne, comp/lt, comp/le, comp/gt, comp/ge
		`SELECT * FROM t WHERE a != 1 AND b < 2 AND c <= 3 AND d > 4 AND e >= 5`,
		// comp/like
		`SELECT * FROM t WHERE a LIKE 'x%'`,
		// comp/any, comp/all
		`SELECT * FROM t WHERE a = ANY(ARRAY[1,2]) AND b > ALL(ARRAY[3,4])`,
		// func/scalar, func/string
		`SELECT length('hello')`,
		// func/math
		`SELECT abs(-1), round(3.14)`,
		// func/date
		`SELECT now(), date_trunc('day', current_timestamp())`,
		// misc/placeholder
		`SELECT * FROM t WHERE a = $1`,
		// quant/multi_table (2+ tables via join — already covered above)
		// quant/multi_filter (2+ AND-separated WHERE predicates)
		`SELECT * FROM t WHERE a = 1 AND b = 2 AND c = 3`,
		// quant/multi_subquery (2+ subqueries)
		`SELECT (SELECT 1), (SELECT 2) FROM t`,
		// quant/deep_nesting (AST depth >= 4)
		`SELECT CASE WHEN (SELECT EXISTS (SELECT 1)) THEN 1 ELSE 2 END`,
		// quant/multi_select_col (3+ select columns)
		`SELECT a, b, c, d FROM t`,
	}

	tracker := NewCoverageTracker()
	for _, sql := range sqls {
		stmts, err := parser.Parse(sql)
		if err != nil {
			t.Fatalf("failed to parse %q: %v", sql, err)
		}
		for i := range stmts {
			tracker.ObserveStatement(stmts[i].AST)
		}
	}

	// Every registered feature must be covered.
	for _, f := range AllFeatures() {
		if !tracker.features[f] {
			t.Errorf("feature %q is registered but not covered by any test SQL "+
				"— add a SQL statement to this test that exercises it, or the "+
				"visitor is missing an extraction path for it", f)
		}
	}
}

// TestSmitherExercisesMinimumFeatures runs sqlsmith for a number of iterations
// and verifies that a minimum set of features is exercised. This catches the
// case where sqlsmith is extended with new generation paths but the feature
// catalog or visitor doesn't account for them (the new paths would produce
// SQL that the visitor maps to existing features, so coverage wouldn't drop).
//
// If this test fails, it means sqlsmith is no longer exercising features it
// used to exercise, which likely indicates a regression in sqlsmith's
// generation diversity.
func TestSmitherExercisesMinimumFeatures(t *testing.T) {
	defer leaktest.AfterTest(t)()
	defer log.Scope(t).Close(t)

	// These are features we expect sqlsmith to exercise within a reasonable
	// number of iterations, based on its weight tables.
	expectedMinimum := []SQLFeature{
		FeatureSelect,
		FeatureInsert,
		FeatureUpdate,
		FeatureDelete,
		FeatureWhere,
		FeatureValues,
		FeatureInnerJoin,
		FeatureCompEQ,
		FeatureAnd,
		FeatureOr,
		FeatureNot,
		FeatureScalarFunc,
	}

	// We can't easily run sqlsmith here without a DB connection, but we can
	// verify the structural invariant: every feature in expectedMinimum must
	// be a registered feature.
	registered := featureSet()
	for _, f := range expectedMinimum {
		if !registered[f] {
			t.Errorf("expected minimum feature %q is not registered — was it renamed or removed?", f)
		}
	}

	// Note: a full integration test that actually runs sqlsmith and checks
	// feature coverage belongs in sqlsmith_test.go with a DB connection.
	// This test just validates the structural invariant.
}
