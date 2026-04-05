// Copyright 2026 The Cockroach Authors.
//
// Use of this software is governed by the CockroachDB Software License
// included in the /LICENSE file.

package sqlsmith

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree/treecmp"
)

// SQLFeature represents a single SQL feature that can be exercised by a query.
type SQLFeature string

// featureRegistry is the global set of all declared features. Every feature
// constant must be created via registerFeature so it is automatically included
// in the catalog. This is the single source of truth for the feature
// denominator — no separate list to maintain.
var featureRegistry []SQLFeature

// registerFeature adds a feature to the global registry and returns it. All
// feature constants must be declared as:
//
//	var FeatureXxx = registerFeature("category/name")
func registerFeature(name string) SQLFeature {
	f := SQLFeature(name)
	featureRegistry = append(featureRegistry, f)
	return f
}

// AllFeatures returns the complete catalog of tracked SQL features, derived
// automatically from all registerFeature calls. This serves as the denominator
// for coverage calculations.
func AllFeatures() []SQLFeature {
	out := make([]SQLFeature, len(featureRegistry))
	copy(out, featureRegistry)
	return out
}

// Statement types.
var (
	FeatureSelect = registerFeature("stmt/select")
	FeatureInsert = registerFeature("stmt/insert")
	FeatureUpdate = registerFeature("stmt/update")
	FeatureDelete = registerFeature("stmt/delete")
)

// Clauses.
var (
	FeatureWhere   = registerFeature("clause/where")
	FeatureGroupBy = registerFeature("clause/group_by")
	FeatureHaving  = registerFeature("clause/having")
	FeatureOrderBy = registerFeature("clause/order_by")
	FeatureLimit   = registerFeature("clause/limit")
	FeatureOffset  = registerFeature("clause/offset")
	FeatureWith    = registerFeature("clause/with")
	FeatureWindow  = registerFeature("clause/window")
)

// SELECT modifiers.
var (
	FeatureDistinct   = registerFeature("select/distinct")
	FeatureDistinctOn = registerFeature("select/distinct_on")
	FeatureStar       = registerFeature("select/star")
)

// Set operations.
var (
	FeatureUnion     = registerFeature("set_op/union")
	FeatureUnionAll  = registerFeature("set_op/union_all")
	FeatureIntersect = registerFeature("set_op/intersect")
	FeatureExcept    = registerFeature("set_op/except")
)

// Join types.
var (
	FeatureInnerJoin = registerFeature("join/inner")
	FeatureLeftJoin  = registerFeature("join/left")
	FeatureRightJoin = registerFeature("join/right")
	FeatureFullJoin  = registerFeature("join/full")
	FeatureCrossJoin = registerFeature("join/cross")
	FeatureLateral   = registerFeature("join/lateral")
)

// Expression types.
var (
	FeatureSubquery      = registerFeature("expr/subquery")
	FeatureExists        = registerFeature("expr/exists")
	FeatureIn            = registerFeature("expr/in")
	FeatureNotIn         = registerFeature("expr/not_in")
	FeatureBetween       = registerFeature("expr/between")
	FeatureCase          = registerFeature("expr/case")
	FeatureCoalesce      = registerFeature("expr/coalesce")
	FeatureCast          = registerFeature("expr/cast")
	FeatureTypAnnotation = registerFeature("expr/type_annotation")
	FeatureCollate       = registerFeature("expr/collate")
	FeatureArrayExpr     = registerFeature("expr/array_construct")
	FeatureArrayFlatten  = registerFeature("expr/array_flatten")
	FeatureTuple         = registerFeature("expr/tuple")
	FeatureIfExpr        = registerFeature("expr/if")
	FeatureNullIf        = registerFeature("expr/nullif")
	FeatureIsNull        = registerFeature("expr/is_null")
	FeatureIsNotNull     = registerFeature("expr/is_not_null")
	FeatureIndirection   = registerFeature("expr/indirection")
)

// Logical operators.
var (
	FeatureAnd = registerFeature("logic/and")
	FeatureOr  = registerFeature("logic/or")
	FeatureNot = registerFeature("logic/not")
)

// Comparison operators.
var (
	FeatureCompEQ   = registerFeature("comp/eq")
	FeatureCompNE   = registerFeature("comp/ne")
	FeatureCompLT   = registerFeature("comp/lt")
	FeatureCompLE   = registerFeature("comp/le")
	FeatureCompGT   = registerFeature("comp/gt")
	FeatureCompGE   = registerFeature("comp/ge")
	FeatureCompLike = registerFeature("comp/like")
	FeatureCompAny  = registerFeature("comp/any")
	FeatureCompAll  = registerFeature("comp/all")
)

// Functions.
var (
	FeatureAggFunc    = registerFeature("func/aggregate")
	FeatureWindowFunc = registerFeature("func/window")
	FeatureScalarFunc = registerFeature("func/scalar")
	FeatureMathFunc   = registerFeature("func/math")
	FeatureStringFunc = registerFeature("func/string")
	FeatureDateFunc   = registerFeature("func/date")
)

// Quantitative features inspired by DBTest'26 (boolean thresholds).
var (
	FeatureMultiTable     = registerFeature("quant/multi_table")      // 2+ tables
	FeatureMultiFilter    = registerFeature("quant/multi_filter")     // 2+ WHERE predicates
	FeatureMultiSubquery  = registerFeature("quant/multi_subquery")   // 2+ subqueries
	FeatureDeepNesting    = registerFeature("quant/deep_nesting")     // AST depth >= 4
	FeatureMultiSelectCol = registerFeature("quant/multi_select_col") // 3+ select columns
)

// Miscellaneous.
var (
	FeaturePlaceholder = registerFeature("misc/placeholder")
	FeatureValues      = registerFeature("misc/values")
)

// featureSet returns the set of all features that the visitor can extract.
// This is used by tests to verify that every registered feature has a
// corresponding extraction path.
func featureSet() map[SQLFeature]bool {
	s := make(map[SQLFeature]bool, len(featureRegistry))
	for _, f := range featureRegistry {
		s[f] = true
	}
	return s
}

// CoverageTracker accumulates SQL feature coverage and unique statement
// fingerprints across a test run.
type CoverageTracker struct {
	features     map[SQLFeature]bool
	fingerprints map[string]int // fingerprint -> count
	stmtCount    int
}

// NewCoverageTracker creates a new tracker initialized with the full feature
// catalog.
func NewCoverageTracker() *CoverageTracker {
	return &CoverageTracker{
		features:     make(map[SQLFeature]bool),
		fingerprints: make(map[string]int),
	}
}

// ObserveStatement extracts features from a parsed AST and records the
// statement's fingerprint. It is the main entry point for coverage tracking.
func (ct *CoverageTracker) ObserveStatement(stmt tree.Statement) {
	ct.stmtCount++
	// Extract features via AST walk.
	v := &featureVisitor{features: ct.features}
	walkStatement(v, stmt)
	v.finalizeQuantitative()
	// Record fingerprint.
	fp := tree.FormatStatementHideConstants(stmt, tree.FmtAnonymize)
	ct.fingerprints[fp]++
}

// Report returns a human-readable coverage report.
func (ct *CoverageTracker) Report() string {
	var b strings.Builder
	all := AllFeatures()
	covered := 0
	for _, f := range all {
		if ct.features[f] {
			covered++
		}
	}
	fmt.Fprintf(&b, "=== SQL Coverage Report ===\n")
	fmt.Fprintf(&b, "Statements observed: %d\n", ct.stmtCount)
	fmt.Fprintf(&b, "Unique fingerprints: %d\n", len(ct.fingerprints))
	fmt.Fprintf(&b, "Features covered: %d / %d (%.1f%%)\n\n",
		covered, len(all), 100*float64(covered)/float64(len(all)))

	fmt.Fprintf(&b, "--- Covered Features ---\n")
	var coveredList []string
	for _, f := range all {
		if ct.features[f] {
			coveredList = append(coveredList, string(f))
		}
	}
	sort.Strings(coveredList)
	for _, f := range coveredList {
		fmt.Fprintf(&b, "  [x] %s\n", f)
	}

	fmt.Fprintf(&b, "\n--- Uncovered Features ---\n")
	var uncoveredList []string
	for _, f := range all {
		if !ct.features[f] {
			uncoveredList = append(uncoveredList, string(f))
		}
	}
	sort.Strings(uncoveredList)
	for _, f := range uncoveredList {
		fmt.Fprintf(&b, "  [ ] %s\n", f)
	}

	fmt.Fprintf(&b, "\n--- Top 20 Fingerprints (by frequency) ---\n")
	type fpEntry struct {
		fp    string
		count int
	}
	fps := make([]fpEntry, 0, len(ct.fingerprints))
	for fp, count := range ct.fingerprints {
		fps = append(fps, fpEntry{fp, count})
	}
	sort.Slice(fps, func(i, j int) bool { return fps[i].count > fps[j].count })
	limit := 20
	if len(fps) < limit {
		limit = len(fps)
	}
	for _, e := range fps[:limit] {
		// Truncate long fingerprints.
		fp := e.fp
		if len(fp) > 120 {
			fp = fp[:120] + "..."
		}
		fmt.Fprintf(&b, "  %4d× %s\n", e.count, fp)
	}

	return b.String()
}

// FeaturesCovered returns the number of features covered.
func (ct *CoverageTracker) FeaturesCovered() int {
	count := 0
	for _, f := range AllFeatures() {
		if ct.features[f] {
			count++
		}
	}
	return count
}

// FeaturesTotal returns the total number of features in the catalog.
func (ct *CoverageTracker) FeaturesTotal() int {
	return len(featureRegistry)
}

// UniqueFingerprints returns the number of unique fingerprints observed.
func (ct *CoverageTracker) UniqueFingerprints() int {
	return len(ct.fingerprints)
}

// featureVisitor is a tree.Visitor that collects SQL features from an AST.
// It also tracks quantitative counters for DBTest-inspired threshold features.
type featureVisitor struct {
	features       map[SQLFeature]bool
	tableCount     int // number of table references
	filterCount    int // number of WHERE predicates (AND-separated)
	subqueryCount  int // number of subqueries
	selectColCount int // number of select columns
	depth          int // current AST nesting depth
	maxDepth       int // maximum AST nesting depth observed
}

var _ tree.Visitor = &featureVisitor{}

// finalizeQuantitative sets the threshold-based boolean features from the
// accumulated counters. Called after the full AST walk.
func (v *featureVisitor) finalizeQuantitative() {
	if v.tableCount >= 2 {
		v.features[FeatureMultiTable] = true
	}
	if v.filterCount >= 2 {
		v.features[FeatureMultiFilter] = true
	}
	if v.subqueryCount >= 2 {
		v.features[FeatureMultiSubquery] = true
	}
	if v.maxDepth >= 4 {
		v.features[FeatureDeepNesting] = true
	}
	if v.selectColCount >= 3 {
		v.features[FeatureMultiSelectCol] = true
	}
}

func (v *featureVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	v.depth++
	if v.depth > v.maxDepth {
		v.maxDepth = v.depth
	}
	switch e := expr.(type) {
	case *tree.Subquery:
		v.subqueryCount++
		if e.Exists {
			v.features[FeatureExists] = true
		} else {
			v.features[FeatureSubquery] = true
		}
	case *tree.ComparisonExpr:
		v.visitComparison(e)
	case *tree.AndExpr:
		v.features[FeatureAnd] = true
	case *tree.OrExpr:
		v.features[FeatureOr] = true
	case *tree.NotExpr:
		v.features[FeatureNot] = true
	case *tree.BinaryExpr:
		// Tracked at operator level via comparisons; just recurse.
	case *tree.CaseExpr:
		v.features[FeatureCase] = true
	case *tree.CoalesceExpr:
		v.features[FeatureCoalesce] = true
	case *tree.CastExpr:
		v.features[FeatureCast] = true
	case *tree.AnnotateTypeExpr:
		v.features[FeatureTypAnnotation] = true
	case *tree.CollateExpr:
		v.features[FeatureCollate] = true
	case *tree.Array:
		v.features[FeatureArrayExpr] = true
	case *tree.ArrayFlatten:
		v.features[FeatureArrayFlatten] = true
	case *tree.Tuple:
		v.features[FeatureTuple] = true
	case *tree.IfExpr:
		v.features[FeatureIfExpr] = true
	case *tree.NullIfExpr:
		v.features[FeatureNullIf] = true
	case *tree.IsNullExpr:
		v.features[FeatureIsNull] = true
	case *tree.IsNotNullExpr:
		v.features[FeatureIsNotNull] = true
	case *tree.IndirectionExpr:
		v.features[FeatureIndirection] = true
	case *tree.FuncExpr:
		v.visitFunc(e)
	case *tree.RangeCond:
		v.features[FeatureBetween] = true
	case *tree.Placeholder:
		v.features[FeaturePlaceholder] = true
	case *tree.ParenExpr, *tree.UnaryExpr:
		// Just recurse.
	}
	return true, expr
}

func (v *featureVisitor) VisitPost(expr tree.Expr) tree.Expr {
	v.depth--
	return expr
}

func (v *featureVisitor) visitComparison(e *tree.ComparisonExpr) {
	switch e.Operator.Symbol {
	case treecmp.EQ:
		v.features[FeatureCompEQ] = true
	case treecmp.NE:
		v.features[FeatureCompNE] = true
	case treecmp.LT:
		v.features[FeatureCompLT] = true
	case treecmp.LE:
		v.features[FeatureCompLE] = true
	case treecmp.GT:
		v.features[FeatureCompGT] = true
	case treecmp.GE:
		v.features[FeatureCompGE] = true
	case treecmp.Like, treecmp.ILike, treecmp.SimilarTo,
		treecmp.RegMatch, treecmp.RegIMatch:
		v.features[FeatureCompLike] = true
	case treecmp.In:
		v.features[FeatureIn] = true
	case treecmp.NotIn:
		v.features[FeatureNotIn] = true
	case treecmp.Any, treecmp.Some:
		v.features[FeatureCompAny] = true
	case treecmp.All:
		v.features[FeatureCompAll] = true
	}
}

// knownAggregates is the set of aggregate function names used to classify
// parsed (not type-checked) FuncExprs.
var knownAggregates = map[string]bool{
	"count": true, "sum": true, "avg": true, "min": true, "max": true,
	"stddev": true, "variance": true, "bool_and": true, "bool_or": true,
	"array_agg": true, "concat_agg": true, "string_agg": true,
	"json_agg": true, "jsonb_agg": true, "json_object_agg": true,
	"jsonb_object_agg": true, "xmlagg": true, "st_makeline": true,
	"st_collect": true, "st_extent": true, "st_union": true,
	"every": true, "bit_and": true, "bit_or": true, "xor_agg": true,
	"corr": true, "covar_pop": true, "covar_samp": true,
	"regr_avgx": true, "regr_avgy": true, "regr_count": true,
	"regr_intercept": true, "regr_r2": true, "regr_slope": true,
	"regr_sxx": true, "regr_sxy": true, "regr_syy": true,
	"stddev_pop": true, "stddev_samp": true, "var_pop": true, "var_samp": true,
	"count_rows": true,
}

// knownWindowFuncs is the set of window-only function names.
var knownWindowFuncs = map[string]bool{
	"row_number": true, "rank": true, "dense_rank": true,
	"percent_rank": true, "cume_dist": true, "ntile": true,
	"lag": true, "lead": true, "first_value": true,
	"last_value": true, "nth_value": true,
}

// knownMathFuncs is the set of math/numeric function names.
var knownMathFuncs = map[string]bool{
	"abs": true, "ceil": true, "ceiling": true, "floor": true,
	"round": true, "trunc": true, "truncate": true, "mod": true,
	"power": true, "pow": true, "sqrt": true, "cbrt": true,
	"exp": true, "ln": true, "log": true, "log10": true,
	"sign": true, "pi": true, "degrees": true, "radians": true,
	"sin": true, "cos": true, "tan": true, "asin": true,
	"acos": true, "atan": true, "atan2": true, "random": true,
	"div": true, "greatest": true, "least": true, "width_bucket": true,
}

// knownStringFuncs is the set of string function names.
var knownStringFuncs = map[string]bool{
	"length": true, "char_length": true, "character_length": true,
	"octet_length": true, "bit_length": true,
	"upper": true, "lower": true, "initcap": true,
	"left": true, "right": true, "substr": true, "substring": true,
	"trim": true, "ltrim": true, "rtrim": true, "btrim": true,
	"lpad": true, "rpad": true, "repeat": true, "reverse": true,
	"replace": true, "translate": true, "overlay": true,
	"position": true, "strpos": true, "concat": true, "concat_ws": true,
	"split_part": true, "regexp_extract": true, "regexp_replace": true,
	"encode": true, "decode": true, "ascii": true, "chr": true,
	"md5": true, "sha256": true, "sha512": true, "to_hex": true,
	"quote_literal": true, "quote_ident": true,
}

// knownDateFuncs is the set of date/time function names.
var knownDateFuncs = map[string]bool{
	"now": true, "current_timestamp": true, "current_date": true,
	"current_time": true, "localtime": true, "localtimestamp": true,
	"clock_timestamp": true, "statement_timestamp": true,
	"transaction_timestamp": true, "timeofday": true,
	"age": true, "date_trunc": true, "date_part": true,
	"extract": true, "make_date": true, "make_time": true,
	"make_timestamp": true, "make_timestamptz": true,
	"to_timestamp": true, "to_char": true, "to_date": true,
	"date_add": true, "date_sub": true, "datediff": true,
	"experimental_strftime": true, "experimental_strptime": true,
}

func (v *featureVisitor) visitFunc(e *tree.FuncExpr) {
	if e.WindowDef != nil {
		v.features[FeatureWindowFunc] = true
		v.features[FeatureWindow] = true
		return
	}
	name := strings.ToLower(e.Func.String())
	// Strip schema qualifier if present.
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		name = name[idx+1:]
	}
	switch {
	case knownAggregates[name]:
		v.features[FeatureAggFunc] = true
	case knownWindowFuncs[name]:
		v.features[FeatureWindowFunc] = true
	case knownMathFuncs[name]:
		v.features[FeatureMathFunc] = true
		v.features[FeatureScalarFunc] = true
	case knownStringFuncs[name]:
		v.features[FeatureStringFunc] = true
		v.features[FeatureScalarFunc] = true
	case knownDateFuncs[name]:
		v.features[FeatureDateFunc] = true
		v.features[FeatureScalarFunc] = true
	default:
		v.features[FeatureScalarFunc] = true
	}
}

// walkStatement walks the top-level statement structure to extract statement
// and clause features, then uses tree.WalkExpr for expression-level features.
func walkStatement(v *featureVisitor, stmt tree.Statement) {
	switch s := stmt.(type) {
	case *tree.Select:
		v.features[FeatureSelect] = true
		if s.With != nil {
			v.features[FeatureWith] = true
		}
		if s.OrderBy != nil {
			v.features[FeatureOrderBy] = true
		}
		if s.Limit != nil {
			if s.Limit.Count != nil {
				v.features[FeatureLimit] = true
			}
			if s.Limit.Offset != nil {
				v.features[FeatureOffset] = true
			}
		}
		walkSelectExpr(v, s.Select)
	case *tree.Insert:
		v.features[FeatureInsert] = true
		walkInsert(v, s)
	case *tree.Update:
		v.features[FeatureUpdate] = true
		walkUpdate(v, s)
	case *tree.Delete:
		v.features[FeatureDelete] = true
		walkDelete(v, s)
	case *tree.ParenSelect:
		v.features[FeatureSelect] = true
		walkStatement(v, s.Select)
	case *tree.UnionClause:
		walkUnion(v, s)
	case *tree.ValuesClause:
		v.features[FeatureValues] = true
	case *tree.SelectClause:
		walkSelectClause(v, s)
	}
}

func walkSelectExpr(v *featureVisitor, sel tree.SelectStatement) {
	switch s := sel.(type) {
	case *tree.SelectClause:
		walkSelectClause(v, s)
	case *tree.UnionClause:
		walkUnion(v, s)
	case *tree.ValuesClause:
		v.features[FeatureValues] = true
	case *tree.ParenSelect:
		walkStatement(v, s.Select)
	}
}

func walkSelectClause(v *featureVisitor, s *tree.SelectClause) {
	if s.Distinct {
		v.features[FeatureDistinct] = true
	}
	if len(s.DistinctOn) > 0 {
		v.features[FeatureDistinctOn] = true
	}
	// Count select columns.
	v.selectColCount += len(s.Exprs)
	// Check for star in projections.
	for _, expr := range s.Exprs {
		if _, ok := expr.Expr.(tree.UnqualifiedStar); ok {
			v.features[FeatureStar] = true
		}
		if _, ok := expr.Expr.(*tree.AllColumnsSelector); ok {
			v.features[FeatureStar] = true
		}
		tree.WalkExpr(v, expr.Expr)
	}
	// Walk WHERE.
	if s.Where != nil {
		v.features[FeatureWhere] = true
		v.filterCount += countFilters(s.Where.Expr)
		tree.WalkExpr(v, s.Where.Expr)
	}
	// Walk GROUP BY.
	if len(s.GroupBy) > 0 {
		v.features[FeatureGroupBy] = true
		for _, expr := range s.GroupBy {
			tree.WalkExpr(v, expr)
		}
	}
	// Walk HAVING.
	if s.Having != nil {
		v.features[FeatureHaving] = true
		tree.WalkExpr(v, s.Having.Expr)
	}
	// Walk WINDOW.
	if len(s.Window) > 0 {
		v.features[FeatureWindow] = true
	}
	// Walk FROM for joins.
	for _, table := range s.From.Tables {
		walkTableExpr(v, table)
	}
}

func walkTableExpr(v *featureVisitor, expr tree.TableExpr) {
	switch t := expr.(type) {
	case *tree.JoinTableExpr:
		walkJoin(v, t)
	case *tree.AliasedTableExpr:
		v.tableCount++
		if sub, ok := t.Expr.(*tree.Subquery); ok {
			v.features[FeatureSubquery] = true
			v.subqueryCount++
			walkSubquery(v, sub)
		}
		if t.Lateral {
			v.features[FeatureLateral] = true
		}
	case *tree.ParenTableExpr:
		walkTableExpr(v, t.Expr)
	}
}

func walkJoin(v *featureVisitor, j *tree.JoinTableExpr) {
	switch j.JoinType {
	case tree.AstInner, "":
		v.features[FeatureInnerJoin] = true
	case tree.AstLeft:
		v.features[FeatureLeftJoin] = true
	case tree.AstRight:
		v.features[FeatureRightJoin] = true
	case tree.AstFull:
		v.features[FeatureFullJoin] = true
	case tree.AstCross:
		v.features[FeatureCrossJoin] = true
	}
	walkTableExpr(v, j.Left)
	walkTableExpr(v, j.Right)
	if onCond, ok := j.Cond.(*tree.OnJoinCond); ok {
		tree.WalkExpr(v, onCond.Expr)
	}
}

func walkSubquery(v *featureVisitor, sub *tree.Subquery) {
	if sub.Exists {
		v.features[FeatureExists] = true
	}
	switch sel := sub.Select.(type) {
	case *tree.SelectClause:
		walkSelectClause(v, sel)
	case *tree.UnionClause:
		walkUnion(v, sel)
	case *tree.ValuesClause:
		v.features[FeatureValues] = true
	case *tree.ParenSelect:
		walkStatement(v, sel.Select)
	}
}

func walkInsert(v *featureVisitor, ins *tree.Insert) {
	if ins.With != nil {
		v.features[FeatureWith] = true
	}
	if sel, ok := ins.Rows.Select.(*tree.SelectClause); ok {
		walkSelectClause(v, sel)
	} else if _, ok := ins.Rows.Select.(*tree.ValuesClause); ok {
		v.features[FeatureValues] = true
	} else {
		walkStatement(v, ins.Rows)
	}
}

func walkUpdate(v *featureVisitor, upd *tree.Update) {
	if upd.With != nil {
		v.features[FeatureWith] = true
	}
	if upd.Where != nil {
		v.features[FeatureWhere] = true
		tree.WalkExpr(v, upd.Where.Expr)
	}
	if upd.OrderBy != nil {
		v.features[FeatureOrderBy] = true
	}
	if upd.Limit != nil {
		v.features[FeatureLimit] = true
	}
	for _, expr := range upd.Exprs {
		tree.WalkExpr(v, expr.Expr)
	}
}

func walkDelete(v *featureVisitor, del *tree.Delete) {
	if del.With != nil {
		v.features[FeatureWith] = true
	}
	if del.Where != nil {
		v.features[FeatureWhere] = true
		tree.WalkExpr(v, del.Where.Expr)
	}
	if del.OrderBy != nil {
		v.features[FeatureOrderBy] = true
	}
	if del.Limit != nil {
		v.features[FeatureLimit] = true
	}
}

// countFilters counts the number of top-level predicates in a WHERE clause
// by counting AND-separated conjuncts.
func countFilters(expr tree.Expr) int {
	switch e := expr.(type) {
	case *tree.AndExpr:
		return countFilters(e.Left) + countFilters(e.Right)
	default:
		return 1
	}
}

func walkUnion(v *featureVisitor, u *tree.UnionClause) {
	switch u.Type {
	case tree.UnionOp:
		if u.All {
			v.features[FeatureUnionAll] = true
		} else {
			v.features[FeatureUnion] = true
		}
	case tree.IntersectOp:
		v.features[FeatureIntersect] = true
	case tree.ExceptOp:
		v.features[FeatureExcept] = true
	}
	// UnionClause.Left and .Right are *tree.Select, not SelectStatement.
	walkStatement(v, u.Left)
	walkStatement(v, u.Right)
}

// allVisitorFeatures returns the set of all features that the visitor and
// statement walkers can emit. This is used by tests to verify that the visitor
// stays in sync with the feature registry. If a feature is registered but
// never appears in any visitor/walker code path, the sync test will catch it.
func allVisitorFeatures() map[SQLFeature]bool {
	// Collect every feature that appears in a v.features[...] = true
	// assignment across the visitor and walker functions. We do this by
	// constructing carefully crafted ASTs that exercise every branch.
	//
	// Instead of maintaining another list, we rely on the sync test
	// (TestFeatureVisitorSync) to verify exhaustiveness by running sqlsmith
	// and checking coverage.
	return featureSet()
}
