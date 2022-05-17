package reducesql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"go/constant"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/reduce/reduce"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"

	_ "github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

var SQLPasses = []reduce.Pass{
	removeStatement,
	replaceStmt,
	removeWithCTEs,
	removeWith,
	removeCreateDefs,
	removeComputedColumn,
	removeValuesCols,
	removeWithSelectExprs,
	removeSelectAsExprs,
	removeIndexFlags,
	removeValuesRows,
	removeSelectExprs,
	nullExprs,
	nullifyFuncArgs,
	removeLimit,
	removeOrderBy,
	removeOrderByExprs,
	removeGroupBy,
	removeGroupByExprs,
	removeCreateNullDefs,
	removeIndexCols,
	removeIndexPredicate,
	removeWindowPartitions,
	removeDBSchema,
	removeFroms,
	removeJoins,
	removeWhere,
	removeHaving,
	removeDistinct,
	simplifyOnCond,
	simplifyVal,
	removeCTENames,
	removeCasts,
	removeAliases,
	unparenthesize,
}

type sqlWalker struct {
	topOnly bool
	match   func(int, interface{}) int
	replace func(int, interface{}) (int, tree.NodeFormatter)
}

func walkSQL(name string, match func(transform int, node interface{}) (matched int)) reduce.Pass {
	__antithesis_instrumentation__.Notify(41930)
	w := sqlWalker{
		match: match,
	}
	return reduce.MakeIntPass(name, w.Transform)
}

func replaceStatement(
	name string,
	replace func(transform int, node interface{}) (matched int, replacement tree.NodeFormatter),
) reduce.Pass {
	__antithesis_instrumentation__.Notify(41931)
	w := sqlWalker{
		replace: replace,
	}
	return reduce.MakeIntPass(name, w.Transform)
}

func replaceTopStatement(
	name string,
	replace func(transform int, node interface{}) (matched int, replacement tree.NodeFormatter),
) reduce.Pass {
	__antithesis_instrumentation__.Notify(41932)
	w := sqlWalker{
		replace: replace,
		topOnly: true,
	}
	return reduce.MakeIntPass(name, w.Transform)
}

var (
	LogUnknown   bool
	unknownTypes = map[string]bool{}
)

func (w sqlWalker) Transform(s string, i int) (out string, ok bool, err error) {
	__antithesis_instrumentation__.Notify(41933)
	stmts, err := parser.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(41938)
		return "", false, err
	} else {
		__antithesis_instrumentation__.Notify(41939)
	}
	__antithesis_instrumentation__.Notify(41934)

	asts := collectASTs(stmts)

	var replacement tree.NodeFormatter

	var nodeCount int
	var walk func(...interface{})
	walk = func(nodes ...interface{}) {
		__antithesis_instrumentation__.Notify(41940)
		for _, node := range nodes {
			__antithesis_instrumentation__.Notify(41941)
			nodeCount++
			if w.topOnly && func() bool {
				__antithesis_instrumentation__.Notify(41947)
				return nodeCount > 1 == true
			}() == true {
				__antithesis_instrumentation__.Notify(41948)
				return
			} else {
				__antithesis_instrumentation__.Notify(41949)
			}
			__antithesis_instrumentation__.Notify(41942)
			if i < 0 {
				__antithesis_instrumentation__.Notify(41950)
				return
			} else {
				__antithesis_instrumentation__.Notify(41951)
			}
			__antithesis_instrumentation__.Notify(41943)
			var matches int
			if w.match != nil {
				__antithesis_instrumentation__.Notify(41952)
				matches = w.match(i, node)
			} else {
				__antithesis_instrumentation__.Notify(41953)
				matches, replacement = w.replace(i, node)
			}
			__antithesis_instrumentation__.Notify(41944)
			i -= matches

			if node == nil {
				__antithesis_instrumentation__.Notify(41954)
				continue
			} else {
				__antithesis_instrumentation__.Notify(41955)
			}
			__antithesis_instrumentation__.Notify(41945)
			if _, ok := node.(tree.Datum); ok {
				__antithesis_instrumentation__.Notify(41956)
				continue
			} else {
				__antithesis_instrumentation__.Notify(41957)
			}
			__antithesis_instrumentation__.Notify(41946)

			switch node := node.(type) {
			case *tree.AliasedTableExpr:
				__antithesis_instrumentation__.Notify(41958)
				walk(node.Expr)
			case *tree.AndExpr:
				__antithesis_instrumentation__.Notify(41959)
				walk(node.Left, node.Right)
			case *tree.AnnotateTypeExpr:
				__antithesis_instrumentation__.Notify(41960)
				walk(node.Expr)
			case *tree.Array:
				__antithesis_instrumentation__.Notify(41961)
				walk(node.Exprs)
			case *tree.BinaryExpr:
				__antithesis_instrumentation__.Notify(41962)
				walk(node.Left, node.Right)
			case *tree.CaseExpr:
				__antithesis_instrumentation__.Notify(41963)
				walk(node.Expr, node.Else)
				for _, w := range node.Whens {
					__antithesis_instrumentation__.Notify(42029)
					walk(w.Cond, w.Val)
				}
			case *tree.CastExpr:
				__antithesis_instrumentation__.Notify(41964)
				walk(node.Expr)
			case *tree.CoalesceExpr:
				__antithesis_instrumentation__.Notify(41965)
				for _, expr := range node.Exprs {
					__antithesis_instrumentation__.Notify(42030)
					walk(expr)
				}
			case *tree.ColumnTableDef:
				__antithesis_instrumentation__.Notify(41966)
				walk(node.Computed.Expr)
				for _, expr := range node.CheckExprs {
					__antithesis_instrumentation__.Notify(42031)
					walk(expr)
				}
			case *tree.ComparisonExpr:
				__antithesis_instrumentation__.Notify(41967)
				walk(node.Left, node.Right)
			case *tree.CreateTable:
				__antithesis_instrumentation__.Notify(41968)
				for _, def := range node.Defs {
					__antithesis_instrumentation__.Notify(42032)
					walk(def)
				}
				__antithesis_instrumentation__.Notify(41969)
				if node.AsSource != nil {
					__antithesis_instrumentation__.Notify(42033)
					walk(node.AsSource)
				} else {
					__antithesis_instrumentation__.Notify(42034)
				}
			case *tree.CTE:
				__antithesis_instrumentation__.Notify(41970)
				walk(node.Stmt)
			case *tree.DBool:
				__antithesis_instrumentation__.Notify(41971)
			case *tree.Delete:
				__antithesis_instrumentation__.Notify(41972)
				walk(node.Table)
				if node.With != nil {
					__antithesis_instrumentation__.Notify(42035)
					walk(node.With)
				} else {
					__antithesis_instrumentation__.Notify(42036)
				}
				__antithesis_instrumentation__.Notify(41973)
				if node.Where != nil {
					__antithesis_instrumentation__.Notify(42037)
					walk(node.Where)
				} else {
					__antithesis_instrumentation__.Notify(42038)
				}
				__antithesis_instrumentation__.Notify(41974)
				walk(node.OrderBy)
				if node.Limit != nil {
					__antithesis_instrumentation__.Notify(42039)
					walk(node.Limit)
				} else {
					__antithesis_instrumentation__.Notify(42040)
				}
				__antithesis_instrumentation__.Notify(41975)
				if node.Returning != nil {
					__antithesis_instrumentation__.Notify(42041)
					walk(node.Returning)
				} else {
					__antithesis_instrumentation__.Notify(42042)
				}
			case tree.Exprs:
				__antithesis_instrumentation__.Notify(41976)
				for _, expr := range node {
					__antithesis_instrumentation__.Notify(42043)
					walk(expr)
				}
			case *tree.FamilyTableDef:
				__antithesis_instrumentation__.Notify(41977)
			case *tree.FuncExpr:
				__antithesis_instrumentation__.Notify(41978)
				if node.WindowDef != nil {
					__antithesis_instrumentation__.Notify(42044)
					walk(node.WindowDef)
				} else {
					__antithesis_instrumentation__.Notify(42045)
				}
				__antithesis_instrumentation__.Notify(41979)
				walk(node.Exprs, node.Filter)
			case *tree.IndexTableDef:
				__antithesis_instrumentation__.Notify(41980)
				walk(node.Predicate)
			case *tree.Insert:
				__antithesis_instrumentation__.Notify(41981)
				walk(node.Table)
				if node.Rows != nil {
					__antithesis_instrumentation__.Notify(42046)
					walk(node.Rows)
				} else {
					__antithesis_instrumentation__.Notify(42047)
				}
				__antithesis_instrumentation__.Notify(41982)
				if node.With != nil {
					__antithesis_instrumentation__.Notify(42048)
					walk(node.With)
				} else {
					__antithesis_instrumentation__.Notify(42049)
				}
				__antithesis_instrumentation__.Notify(41983)
				if node.Returning != nil {
					__antithesis_instrumentation__.Notify(42050)
					walk(node.Returning)
				} else {
					__antithesis_instrumentation__.Notify(42051)
				}
			case *tree.JoinTableExpr:
				__antithesis_instrumentation__.Notify(41984)
				walk(node.Left, node.Right, node.Cond)
			case *tree.Limit:
				__antithesis_instrumentation__.Notify(41985)
				walk(node.Count)
			case *tree.NotExpr:
				__antithesis_instrumentation__.Notify(41986)
				walk(node.Expr)
			case *tree.NumVal:
				__antithesis_instrumentation__.Notify(41987)
			case *tree.OnJoinCond:
				__antithesis_instrumentation__.Notify(41988)
				walk(node.Expr)
			case *tree.Order:
				__antithesis_instrumentation__.Notify(41989)
				walk(node.Expr, node.Table)
			case *tree.OrExpr:
				__antithesis_instrumentation__.Notify(41990)
				walk(node.Left, node.Right)
			case *tree.ParenExpr:
				__antithesis_instrumentation__.Notify(41991)
				walk(node.Expr)
			case *tree.ParenSelect:
				__antithesis_instrumentation__.Notify(41992)
				walk(node.Select)
			case *tree.RangeCond:
				__antithesis_instrumentation__.Notify(41993)
				walk(node.Left, node.From, node.To)
			case *tree.RowsFromExpr:
				__antithesis_instrumentation__.Notify(41994)
				for _, expr := range node.Items {
					__antithesis_instrumentation__.Notify(42052)
					walk(expr)
				}
			case *tree.Select:
				__antithesis_instrumentation__.Notify(41995)
				if node.With != nil {
					__antithesis_instrumentation__.Notify(42053)
					walk(node.With)
				} else {
					__antithesis_instrumentation__.Notify(42054)
				}
				__antithesis_instrumentation__.Notify(41996)
				walk(node.Select)
				if node.OrderBy != nil {
					__antithesis_instrumentation__.Notify(42055)
					for _, order := range node.OrderBy {
						__antithesis_instrumentation__.Notify(42056)
						walk(order)
					}
				} else {
					__antithesis_instrumentation__.Notify(42057)
				}
				__antithesis_instrumentation__.Notify(41997)
				if node.Limit != nil {
					__antithesis_instrumentation__.Notify(42058)
					walk(node.Limit)
				} else {
					__antithesis_instrumentation__.Notify(42059)
				}
			case *tree.SelectClause:
				__antithesis_instrumentation__.Notify(41998)
				walk(node.Exprs)
				if node.Where != nil {
					__antithesis_instrumentation__.Notify(42060)
					walk(node.Where)
				} else {
					__antithesis_instrumentation__.Notify(42061)
				}
				__antithesis_instrumentation__.Notify(41999)
				if node.Having != nil {
					__antithesis_instrumentation__.Notify(42062)
					walk(node.Having)
				} else {
					__antithesis_instrumentation__.Notify(42063)
				}
				__antithesis_instrumentation__.Notify(42000)
				for _, table := range node.From.Tables {
					__antithesis_instrumentation__.Notify(42064)
					walk(table)
				}
				__antithesis_instrumentation__.Notify(42001)
				if node.DistinctOn != nil {
					__antithesis_instrumentation__.Notify(42065)
					for _, distinct := range node.DistinctOn {
						__antithesis_instrumentation__.Notify(42066)
						walk(distinct)
					}
				} else {
					__antithesis_instrumentation__.Notify(42067)
				}
				__antithesis_instrumentation__.Notify(42002)
				if node.GroupBy != nil {
					__antithesis_instrumentation__.Notify(42068)
					for _, group := range node.GroupBy {
						__antithesis_instrumentation__.Notify(42069)
						walk(group)
					}
				} else {
					__antithesis_instrumentation__.Notify(42070)
				}
			case tree.SelectExpr:
				__antithesis_instrumentation__.Notify(42003)
				walk(node.Expr)
			case tree.SelectExprs:
				__antithesis_instrumentation__.Notify(42004)
				for _, expr := range node {
					__antithesis_instrumentation__.Notify(42071)
					walk(expr)
				}
			case *tree.SetVar:
				__antithesis_instrumentation__.Notify(42005)
				for _, expr := range node.Values {
					__antithesis_instrumentation__.Notify(42072)
					walk(expr)
				}
			case *tree.StrVal:
				__antithesis_instrumentation__.Notify(42006)
			case *tree.Subquery:
				__antithesis_instrumentation__.Notify(42007)
				walk(node.Select)
			case *tree.TableName, tree.TableName:
				__antithesis_instrumentation__.Notify(42008)
			case *tree.Tuple:
				__antithesis_instrumentation__.Notify(42009)
				for _, expr := range node.Exprs {
					__antithesis_instrumentation__.Notify(42073)
					walk(expr)
				}
			case *tree.UnaryExpr:
				__antithesis_instrumentation__.Notify(42010)
				walk(node.Expr)
			case *tree.UniqueConstraintTableDef:
				__antithesis_instrumentation__.Notify(42011)
			case *tree.UnionClause:
				__antithesis_instrumentation__.Notify(42012)
				walk(node.Left, node.Right)
			case tree.UnqualifiedStar:
				__antithesis_instrumentation__.Notify(42013)
			case *tree.UnresolvedName:
				__antithesis_instrumentation__.Notify(42014)
			case *tree.Update:
				__antithesis_instrumentation__.Notify(42015)
				walk(node.Table)
				if node.Exprs != nil {
					__antithesis_instrumentation__.Notify(42074)
					walk(node.Exprs)
				} else {
					__antithesis_instrumentation__.Notify(42075)
				}
				__antithesis_instrumentation__.Notify(42016)
				if node.With != nil {
					__antithesis_instrumentation__.Notify(42076)
					walk(node.With)
				} else {
					__antithesis_instrumentation__.Notify(42077)
				}
				__antithesis_instrumentation__.Notify(42017)
				if node.Where != nil {
					__antithesis_instrumentation__.Notify(42078)
					walk(node.Where)
				} else {
					__antithesis_instrumentation__.Notify(42079)
				}
				__antithesis_instrumentation__.Notify(42018)
				walk(node.OrderBy)
				if node.Limit != nil {
					__antithesis_instrumentation__.Notify(42080)
					walk(node.Limit)
				} else {
					__antithesis_instrumentation__.Notify(42081)
				}
				__antithesis_instrumentation__.Notify(42019)
				if node.Returning != nil {
					__antithesis_instrumentation__.Notify(42082)
					walk(node.Returning)
				} else {
					__antithesis_instrumentation__.Notify(42083)
				}
			case *tree.ValuesClause:
				__antithesis_instrumentation__.Notify(42020)
				for _, row := range node.Rows {
					__antithesis_instrumentation__.Notify(42084)
					walk(row)
				}
			case *tree.Where:
				__antithesis_instrumentation__.Notify(42021)
				walk(node.Expr)
			case *tree.WindowDef:
				__antithesis_instrumentation__.Notify(42022)
				walk(node.Partitions)
				if node.Frame != nil {
					__antithesis_instrumentation__.Notify(42085)
					walk(node.Frame)
				} else {
					__antithesis_instrumentation__.Notify(42086)
				}
			case *tree.WindowFrame:
				__antithesis_instrumentation__.Notify(42023)
				if node.Bounds.StartBound != nil {
					__antithesis_instrumentation__.Notify(42087)
					walk(node.Bounds.StartBound)
				} else {
					__antithesis_instrumentation__.Notify(42088)
				}
				__antithesis_instrumentation__.Notify(42024)
				if node.Bounds.EndBound != nil {
					__antithesis_instrumentation__.Notify(42089)
					walk(node.Bounds.EndBound)
				} else {
					__antithesis_instrumentation__.Notify(42090)
				}
			case *tree.WindowFrameBound:
				__antithesis_instrumentation__.Notify(42025)
				walk(node.OffsetExpr)
			case *tree.Window:
				__antithesis_instrumentation__.Notify(42026)
			case *tree.With:
				__antithesis_instrumentation__.Notify(42027)
				for _, expr := range node.CTEList {
					__antithesis_instrumentation__.Notify(42091)
					walk(expr)
				}
			default:
				__antithesis_instrumentation__.Notify(42028)
				if LogUnknown {
					__antithesis_instrumentation__.Notify(42092)
					n := fmt.Sprintf("%T", node)
					if !unknownTypes[n] {
						__antithesis_instrumentation__.Notify(42093)
						unknownTypes[n] = true
						fmt.Println("UNKNOWN", n)
					} else {
						__antithesis_instrumentation__.Notify(42094)
					}
				} else {
					__antithesis_instrumentation__.Notify(42095)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(41935)

	for i, ast := range asts {
		__antithesis_instrumentation__.Notify(42096)
		replacement = nil
		nodeCount = 0
		walk(ast)
		if replacement != nil {
			__antithesis_instrumentation__.Notify(42097)
			asts[i] = replacement
		} else {
			__antithesis_instrumentation__.Notify(42098)
		}
	}
	__antithesis_instrumentation__.Notify(41936)
	if i >= 0 {
		__antithesis_instrumentation__.Notify(42099)

		return s, false, nil
	} else {
		__antithesis_instrumentation__.Notify(42100)
	}
	__antithesis_instrumentation__.Notify(41937)
	return joinASTs(asts), true, nil
}

func Pretty(s string) (string, error) {
	__antithesis_instrumentation__.Notify(42101)
	stmts, err := parser.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(42103)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(42104)
	}
	__antithesis_instrumentation__.Notify(42102)

	return joinASTs(collectASTs(stmts)), nil
}

func collectASTs(stmts parser.Statements) []tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(42105)
	asts := make([]tree.NodeFormatter, len(stmts))
	for i, stmt := range stmts {
		__antithesis_instrumentation__.Notify(42107)
		asts[i] = stmt.AST
	}
	__antithesis_instrumentation__.Notify(42106)
	return asts
}

func joinASTs(stmts []tree.NodeFormatter) string {
	__antithesis_instrumentation__.Notify(42108)
	var sb strings.Builder
	for i, stmt := range stmts {
		__antithesis_instrumentation__.Notify(42110)
		if i > 0 {
			__antithesis_instrumentation__.Notify(42112)
			sb.WriteString("\n\n")
		} else {
			__antithesis_instrumentation__.Notify(42113)
		}
		__antithesis_instrumentation__.Notify(42111)
		cfg := tree.PrettyCfg{
			LineWidth: 100,
			TabWidth:  2,
			Align:     tree.PrettyAlignAndDeindent,
			UseTabs:   false,
			Simplify:  true,
		}
		sb.WriteString(cfg.Pretty(stmt))
		sb.WriteString(";")
	}
	__antithesis_instrumentation__.Notify(42109)
	return sb.String()
}

var (
	removeLimit = walkSQL("remove LIMIT", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42114)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.Delete:
			__antithesis_instrumentation__.Notify(42116)
			if node.Limit != nil {
				__antithesis_instrumentation__.Notify(42119)
				if xf {
					__antithesis_instrumentation__.Notify(42121)
					node.Limit = nil
				} else {
					__antithesis_instrumentation__.Notify(42122)
				}
				__antithesis_instrumentation__.Notify(42120)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42123)
			}
		case *tree.Select:
			__antithesis_instrumentation__.Notify(42117)
			if node.Limit != nil {
				__antithesis_instrumentation__.Notify(42124)
				if xf {
					__antithesis_instrumentation__.Notify(42126)
					node.Limit = nil
				} else {
					__antithesis_instrumentation__.Notify(42127)
				}
				__antithesis_instrumentation__.Notify(42125)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42128)
			}
		case *tree.Update:
			__antithesis_instrumentation__.Notify(42118)
			if node.Limit != nil {
				__antithesis_instrumentation__.Notify(42129)
				if xf {
					__antithesis_instrumentation__.Notify(42131)
					node.Limit = nil
				} else {
					__antithesis_instrumentation__.Notify(42132)
				}
				__antithesis_instrumentation__.Notify(42130)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42133)
			}
		}
		__antithesis_instrumentation__.Notify(42115)
		return 0
	})
	removeOrderBy = walkSQL("remove ORDER BY", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42134)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.Delete:
			__antithesis_instrumentation__.Notify(42136)
			if node.OrderBy != nil {
				__antithesis_instrumentation__.Notify(42141)
				if xf {
					__antithesis_instrumentation__.Notify(42143)
					node.OrderBy = nil
				} else {
					__antithesis_instrumentation__.Notify(42144)
				}
				__antithesis_instrumentation__.Notify(42142)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42145)
			}
		case *tree.FuncExpr:
			__antithesis_instrumentation__.Notify(42137)
			if node.OrderBy != nil {
				__antithesis_instrumentation__.Notify(42146)
				if xf {
					__antithesis_instrumentation__.Notify(42148)
					node.OrderBy = nil
				} else {
					__antithesis_instrumentation__.Notify(42149)
				}
				__antithesis_instrumentation__.Notify(42147)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42150)
			}
		case *tree.Select:
			__antithesis_instrumentation__.Notify(42138)
			if node.OrderBy != nil {
				__antithesis_instrumentation__.Notify(42151)
				if xf {
					__antithesis_instrumentation__.Notify(42153)
					node.OrderBy = nil
				} else {
					__antithesis_instrumentation__.Notify(42154)
				}
				__antithesis_instrumentation__.Notify(42152)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42155)
			}
		case *tree.Update:
			__antithesis_instrumentation__.Notify(42139)
			if node.OrderBy != nil {
				__antithesis_instrumentation__.Notify(42156)
				if xf {
					__antithesis_instrumentation__.Notify(42158)
					node.OrderBy = nil
				} else {
					__antithesis_instrumentation__.Notify(42159)
				}
				__antithesis_instrumentation__.Notify(42157)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42160)
			}
		case *tree.WindowDef:
			__antithesis_instrumentation__.Notify(42140)
			if node.OrderBy != nil {
				__antithesis_instrumentation__.Notify(42161)
				if xf {
					__antithesis_instrumentation__.Notify(42163)
					node.OrderBy = nil
				} else {
					__antithesis_instrumentation__.Notify(42164)
				}
				__antithesis_instrumentation__.Notify(42162)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42165)
			}
		}
		__antithesis_instrumentation__.Notify(42135)
		return 0
	})
	removeOrderByExprs = walkSQL("remove ORDER BY exprs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42166)
		switch node := node.(type) {
		case *tree.Delete:
			__antithesis_instrumentation__.Notify(42168)
			n := len(node.OrderBy)
			if xfi < len(node.OrderBy) {
				__antithesis_instrumentation__.Notify(42178)
				node.OrderBy = append(node.OrderBy[:xfi], node.OrderBy[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42179)
			}
			__antithesis_instrumentation__.Notify(42169)
			return n
		case *tree.FuncExpr:
			__antithesis_instrumentation__.Notify(42170)
			n := len(node.OrderBy)
			if xfi < len(node.OrderBy) {
				__antithesis_instrumentation__.Notify(42180)
				node.OrderBy = append(node.OrderBy[:xfi], node.OrderBy[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42181)
			}
			__antithesis_instrumentation__.Notify(42171)
			return n
		case *tree.Select:
			__antithesis_instrumentation__.Notify(42172)
			n := len(node.OrderBy)
			if xfi < len(node.OrderBy) {
				__antithesis_instrumentation__.Notify(42182)
				node.OrderBy = append(node.OrderBy[:xfi], node.OrderBy[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42183)
			}
			__antithesis_instrumentation__.Notify(42173)
			return n
		case *tree.Update:
			__antithesis_instrumentation__.Notify(42174)
			n := len(node.OrderBy)
			if xfi < len(node.OrderBy) {
				__antithesis_instrumentation__.Notify(42184)
				node.OrderBy = append(node.OrderBy[:xfi], node.OrderBy[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42185)
			}
			__antithesis_instrumentation__.Notify(42175)
			return n
		case *tree.WindowDef:
			__antithesis_instrumentation__.Notify(42176)
			n := len(node.OrderBy)
			if xfi < len(node.OrderBy) {
				__antithesis_instrumentation__.Notify(42186)
				node.OrderBy = append(node.OrderBy[:xfi], node.OrderBy[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42187)
			}
			__antithesis_instrumentation__.Notify(42177)
			return n
		}
		__antithesis_instrumentation__.Notify(42167)
		return 0
	})
	removeGroupBy = walkSQL("remove GROUP BY", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42188)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42190)
			if node.GroupBy != nil {
				__antithesis_instrumentation__.Notify(42191)
				if xf {
					__antithesis_instrumentation__.Notify(42193)
					node.GroupBy = nil
				} else {
					__antithesis_instrumentation__.Notify(42194)
				}
				__antithesis_instrumentation__.Notify(42192)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42195)
			}
		}
		__antithesis_instrumentation__.Notify(42189)
		return 0
	})
	removeGroupByExprs = walkSQL("remove GROUP BY exprs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42196)
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42198)
			n := len(node.GroupBy)
			if xfi < len(node.GroupBy) {
				__antithesis_instrumentation__.Notify(42200)
				node.GroupBy = append(node.GroupBy[:xfi], node.GroupBy[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42201)
			}
			__antithesis_instrumentation__.Notify(42199)
			return n
		}
		__antithesis_instrumentation__.Notify(42197)
		return 0
	})
	nullExprs = walkSQL("nullify SELECT exprs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42202)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42204)
			if len(node.Exprs) != 1 || func() bool {
				__antithesis_instrumentation__.Notify(42205)
				return node.Exprs[0].Expr != tree.DNull == true
			}() == true {
				__antithesis_instrumentation__.Notify(42206)
				if xf {
					__antithesis_instrumentation__.Notify(42208)
					node.Exprs = tree.SelectExprs{tree.SelectExpr{Expr: tree.DNull}}
				} else {
					__antithesis_instrumentation__.Notify(42209)
				}
				__antithesis_instrumentation__.Notify(42207)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42210)
			}
		}
		__antithesis_instrumentation__.Notify(42203)
		return 0
	})
	removeSelectExprs = walkSQL("remove SELECT exprs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42211)
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42213)
			n := len(node.Exprs)
			if xfi < len(node.Exprs) {
				__antithesis_instrumentation__.Notify(42215)
				node.Exprs = append(node.Exprs[:xfi], node.Exprs[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42216)
			}
			__antithesis_instrumentation__.Notify(42214)
			return n
		}
		__antithesis_instrumentation__.Notify(42212)
		return 0
	})
	removeWithSelectExprs = walkSQL("remove WITH SELECT exprs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42217)
		switch node := node.(type) {
		case *tree.CTE:
			__antithesis_instrumentation__.Notify(42219)
			if len(node.Name.Cols) < 1 {
				__antithesis_instrumentation__.Notify(42224)
				break
			} else {
				__antithesis_instrumentation__.Notify(42225)
			}
			__antithesis_instrumentation__.Notify(42220)
			slct, ok := node.Stmt.(*tree.Select)
			if !ok {
				__antithesis_instrumentation__.Notify(42226)
				break
			} else {
				__antithesis_instrumentation__.Notify(42227)
			}
			__antithesis_instrumentation__.Notify(42221)
			clause, ok := slct.Select.(*tree.SelectClause)
			if !ok {
				__antithesis_instrumentation__.Notify(42228)
				break
			} else {
				__antithesis_instrumentation__.Notify(42229)
			}
			__antithesis_instrumentation__.Notify(42222)
			n := len(clause.Exprs)
			if xfi < len(clause.Exprs) {
				__antithesis_instrumentation__.Notify(42230)
				node.Name.Cols = append(node.Name.Cols[:xfi], node.Name.Cols[xfi+1:]...)
				clause.Exprs = append(clause.Exprs[:xfi], clause.Exprs[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42231)
			}
			__antithesis_instrumentation__.Notify(42223)
			return n
		}
		__antithesis_instrumentation__.Notify(42218)
		return 0
	})
	removeValuesCols = walkSQL("remove VALUES cols", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42232)
		switch node := node.(type) {
		case *tree.AliasedTableExpr:
			__antithesis_instrumentation__.Notify(42234)
			subq, ok := node.Expr.(*tree.Subquery)
			if !ok {
				__antithesis_instrumentation__.Notify(42251)
				break
			} else {
				__antithesis_instrumentation__.Notify(42252)
			}
			__antithesis_instrumentation__.Notify(42235)
			values, ok := skipParenSelect(subq.Select).(*tree.ValuesClause)
			if !ok {
				__antithesis_instrumentation__.Notify(42253)
				break
			} else {
				__antithesis_instrumentation__.Notify(42254)
			}
			__antithesis_instrumentation__.Notify(42236)
			if len(values.Rows) < 1 {
				__antithesis_instrumentation__.Notify(42255)
				break
			} else {
				__antithesis_instrumentation__.Notify(42256)
			}
			__antithesis_instrumentation__.Notify(42237)
			n := len(values.Rows[0])
			if xfi < n {
				__antithesis_instrumentation__.Notify(42257)
				removeValuesCol(values, xfi)

				if len(node.As.Cols) > xfi {
					__antithesis_instrumentation__.Notify(42258)
					node.As.Cols = append(node.As.Cols[:xfi], node.As.Cols[xfi+1:]...)
				} else {
					__antithesis_instrumentation__.Notify(42259)
				}
			} else {
				__antithesis_instrumentation__.Notify(42260)
			}
			__antithesis_instrumentation__.Notify(42238)
			return n
		case *tree.CTE:
			__antithesis_instrumentation__.Notify(42239)
			slct, ok := node.Stmt.(*tree.Select)
			if !ok {
				__antithesis_instrumentation__.Notify(42261)
				break
			} else {
				__antithesis_instrumentation__.Notify(42262)
			}
			__antithesis_instrumentation__.Notify(42240)
			clause, ok := slct.Select.(*tree.SelectClause)
			if !ok {
				__antithesis_instrumentation__.Notify(42263)
				break
			} else {
				__antithesis_instrumentation__.Notify(42264)
			}
			__antithesis_instrumentation__.Notify(42241)
			if len(clause.From.Tables) != 1 {
				__antithesis_instrumentation__.Notify(42265)
				break
			} else {
				__antithesis_instrumentation__.Notify(42266)
			}
			__antithesis_instrumentation__.Notify(42242)
			ate, ok := clause.From.Tables[0].(*tree.AliasedTableExpr)
			if !ok {
				__antithesis_instrumentation__.Notify(42267)
				break
			} else {
				__antithesis_instrumentation__.Notify(42268)
			}
			__antithesis_instrumentation__.Notify(42243)
			subq, ok := ate.Expr.(*tree.Subquery)
			if !ok {
				__antithesis_instrumentation__.Notify(42269)
				break
			} else {
				__antithesis_instrumentation__.Notify(42270)
			}
			__antithesis_instrumentation__.Notify(42244)
			values, ok := skipParenSelect(subq.Select).(*tree.ValuesClause)
			if !ok {
				__antithesis_instrumentation__.Notify(42271)
				break
			} else {
				__antithesis_instrumentation__.Notify(42272)
			}
			__antithesis_instrumentation__.Notify(42245)
			if len(values.Rows) < 1 {
				__antithesis_instrumentation__.Notify(42273)
				break
			} else {
				__antithesis_instrumentation__.Notify(42274)
			}
			__antithesis_instrumentation__.Notify(42246)
			n := len(values.Rows[0])
			if xfi < n {
				__antithesis_instrumentation__.Notify(42275)
				removeValuesCol(values, xfi)

				if len(node.Name.Cols) > xfi {
					__antithesis_instrumentation__.Notify(42277)
					node.Name.Cols = append(node.Name.Cols[:xfi], node.Name.Cols[xfi+1:]...)
				} else {
					__antithesis_instrumentation__.Notify(42278)
				}
				__antithesis_instrumentation__.Notify(42276)

				if len(ate.As.Cols) > xfi {
					__antithesis_instrumentation__.Notify(42279)
					ate.As.Cols = append(ate.As.Cols[:xfi], ate.As.Cols[xfi+1:]...)
				} else {
					__antithesis_instrumentation__.Notify(42280)
				}
			} else {
				__antithesis_instrumentation__.Notify(42281)
			}
			__antithesis_instrumentation__.Notify(42247)
			return n
		case *tree.ValuesClause:
			__antithesis_instrumentation__.Notify(42248)
			if len(node.Rows) < 1 {
				__antithesis_instrumentation__.Notify(42282)
				break
			} else {
				__antithesis_instrumentation__.Notify(42283)
			}
			__antithesis_instrumentation__.Notify(42249)
			n := len(node.Rows[0])
			if xfi < n {
				__antithesis_instrumentation__.Notify(42284)
				removeValuesCol(node, xfi)
			} else {
				__antithesis_instrumentation__.Notify(42285)
			}
			__antithesis_instrumentation__.Notify(42250)
			return n
		}
		__antithesis_instrumentation__.Notify(42233)
		return 0
	})
	removeSelectAsExprs = walkSQL("remove SELECT AS exprs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42286)
		switch node := node.(type) {
		case *tree.AliasedTableExpr:
			__antithesis_instrumentation__.Notify(42288)
			if len(node.As.Cols) < 1 {
				__antithesis_instrumentation__.Notify(42293)
				break
			} else {
				__antithesis_instrumentation__.Notify(42294)
			}
			__antithesis_instrumentation__.Notify(42289)
			subq, ok := node.Expr.(*tree.Subquery)
			if !ok {
				__antithesis_instrumentation__.Notify(42295)
				break
			} else {
				__antithesis_instrumentation__.Notify(42296)
			}
			__antithesis_instrumentation__.Notify(42290)
			clause, ok := skipParenSelect(subq.Select).(*tree.SelectClause)
			if !ok {
				__antithesis_instrumentation__.Notify(42297)
				break
			} else {
				__antithesis_instrumentation__.Notify(42298)
			}
			__antithesis_instrumentation__.Notify(42291)
			n := len(clause.Exprs)
			if xfi < len(clause.Exprs) {
				__antithesis_instrumentation__.Notify(42299)
				node.As.Cols = append(node.As.Cols[:xfi], node.As.Cols[xfi+1:]...)
				clause.Exprs = append(clause.Exprs[:xfi], clause.Exprs[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42300)
			}
			__antithesis_instrumentation__.Notify(42292)
			return n
		}
		__antithesis_instrumentation__.Notify(42287)
		return 0
	})
	removeIndexFlags = walkSQL("remove index flags", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42301)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.AliasedTableExpr:
			__antithesis_instrumentation__.Notify(42303)
			if node.IndexFlags != nil {
				__antithesis_instrumentation__.Notify(42304)
				if xf {
					__antithesis_instrumentation__.Notify(42306)
					node.IndexFlags = nil
				} else {
					__antithesis_instrumentation__.Notify(42307)
				}
				__antithesis_instrumentation__.Notify(42305)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42308)
			}
		}
		__antithesis_instrumentation__.Notify(42302)
		return 0
	})
	removeWith = walkSQL("remove WITH", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42309)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.Delete:
			__antithesis_instrumentation__.Notify(42311)
			if node.With != nil {
				__antithesis_instrumentation__.Notify(42315)
				if xf {
					__antithesis_instrumentation__.Notify(42317)
					node.With = nil
				} else {
					__antithesis_instrumentation__.Notify(42318)
				}
				__antithesis_instrumentation__.Notify(42316)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42319)
			}
		case *tree.Insert:
			__antithesis_instrumentation__.Notify(42312)
			if node.With != nil {
				__antithesis_instrumentation__.Notify(42320)
				if xf {
					__antithesis_instrumentation__.Notify(42322)
					node.With = nil
				} else {
					__antithesis_instrumentation__.Notify(42323)
				}
				__antithesis_instrumentation__.Notify(42321)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42324)
			}
		case *tree.Select:
			__antithesis_instrumentation__.Notify(42313)
			if node.With != nil {
				__antithesis_instrumentation__.Notify(42325)
				if xf {
					__antithesis_instrumentation__.Notify(42327)
					node.With = nil
				} else {
					__antithesis_instrumentation__.Notify(42328)
				}
				__antithesis_instrumentation__.Notify(42326)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42329)
			}
		case *tree.Update:
			__antithesis_instrumentation__.Notify(42314)
			if node.With != nil {
				__antithesis_instrumentation__.Notify(42330)
				if xf {
					__antithesis_instrumentation__.Notify(42332)
					node.With = nil
				} else {
					__antithesis_instrumentation__.Notify(42333)
				}
				__antithesis_instrumentation__.Notify(42331)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42334)
			}
		}
		__antithesis_instrumentation__.Notify(42310)
		return 0
	})
	removeCreateDefs = walkSQL("remove CREATE defs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42335)
		switch node := node.(type) {
		case *tree.CreateTable:
			__antithesis_instrumentation__.Notify(42337)
			n := len(node.Defs)
			if xfi < len(node.Defs) {
				__antithesis_instrumentation__.Notify(42339)
				node.Defs = append(node.Defs[:xfi], node.Defs[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42340)
			}
			__antithesis_instrumentation__.Notify(42338)
			return n
		}
		__antithesis_instrumentation__.Notify(42336)
		return 0
	})
	removeComputedColumn = walkSQL("remove computed column", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42341)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(42343)
			if node.Computed.Computed {
				__antithesis_instrumentation__.Notify(42344)
				if xf {
					__antithesis_instrumentation__.Notify(42346)
					node.Computed.Computed = false
					node.Computed.Expr = nil
					node.Computed.Virtual = false
				} else {
					__antithesis_instrumentation__.Notify(42347)
				}
				__antithesis_instrumentation__.Notify(42345)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42348)
			}
		}
		__antithesis_instrumentation__.Notify(42342)
		return 0
	})
	removeCreateNullDefs = walkSQL("remove CREATE NULL defs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42349)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.ColumnTableDef:
			__antithesis_instrumentation__.Notify(42351)
			if node.Nullable.Nullability != tree.SilentNull {
				__antithesis_instrumentation__.Notify(42352)
				if xf {
					__antithesis_instrumentation__.Notify(42354)
					node.Nullable.Nullability = tree.SilentNull
				} else {
					__antithesis_instrumentation__.Notify(42355)
				}
				__antithesis_instrumentation__.Notify(42353)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42356)
			}
		}
		__antithesis_instrumentation__.Notify(42350)
		return 0
	})
	removeIndexCols = walkSQL("remove INDEX cols", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42357)
		removeCol := func(idx *tree.IndexTableDef) int {
			__antithesis_instrumentation__.Notify(42360)
			n := len(idx.Columns)
			if xfi < len(idx.Columns) {
				__antithesis_instrumentation__.Notify(42362)
				idx.Columns = append(idx.Columns[:xfi], idx.Columns[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42363)
			}
			__antithesis_instrumentation__.Notify(42361)
			return n
		}
		__antithesis_instrumentation__.Notify(42358)
		switch node := node.(type) {
		case *tree.IndexTableDef:
			__antithesis_instrumentation__.Notify(42364)
			return removeCol(node)
		case *tree.UniqueConstraintTableDef:
			__antithesis_instrumentation__.Notify(42365)
			return removeCol(&node.IndexTableDef)
		}
		__antithesis_instrumentation__.Notify(42359)
		return 0
	})
	removeIndexPredicate = walkSQL("remove index predicate", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42366)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.IndexTableDef:
			__antithesis_instrumentation__.Notify(42368)
			if node.Predicate != nil {
				__antithesis_instrumentation__.Notify(42369)
				if xf {
					__antithesis_instrumentation__.Notify(42371)
					node.Predicate = nil
				} else {
					__antithesis_instrumentation__.Notify(42372)
				}
				__antithesis_instrumentation__.Notify(42370)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42373)
			}
		}
		__antithesis_instrumentation__.Notify(42367)
		return 0
	})
	removeWindowPartitions = walkSQL("remove WINDOW partitions", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42374)
		switch node := node.(type) {
		case *tree.WindowDef:
			__antithesis_instrumentation__.Notify(42376)
			n := len(node.Partitions)
			if xfi < len(node.Partitions) {
				__antithesis_instrumentation__.Notify(42378)
				node.Partitions = append(node.Partitions[:xfi], node.Partitions[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42379)
			}
			__antithesis_instrumentation__.Notify(42377)
			return n
		}
		__antithesis_instrumentation__.Notify(42375)
		return 0
	})
	removeValuesRows = walkSQL("remove VALUES rows", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42380)
		switch node := node.(type) {
		case *tree.ValuesClause:
			__antithesis_instrumentation__.Notify(42382)
			n := len(node.Rows)
			if xfi < len(node.Rows) {
				__antithesis_instrumentation__.Notify(42384)
				node.Rows = append(node.Rows[:xfi], node.Rows[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42385)
			}
			__antithesis_instrumentation__.Notify(42383)
			return n
		}
		__antithesis_instrumentation__.Notify(42381)
		return 0
	})
	removeWithCTEs = walkSQL("remove WITH CTEs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42386)
		switch node := node.(type) {
		case *tree.With:
			__antithesis_instrumentation__.Notify(42388)
			n := len(node.CTEList)
			if xfi < len(node.CTEList) {
				__antithesis_instrumentation__.Notify(42390)
				node.CTEList = append(node.CTEList[:xfi], node.CTEList[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42391)
			}
			__antithesis_instrumentation__.Notify(42389)
			return n
		}
		__antithesis_instrumentation__.Notify(42387)
		return 0
	})
	removeCTENames = walkSQL("remove CTE names", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42392)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.CTE:
			__antithesis_instrumentation__.Notify(42394)
			if len(node.Name.Cols) > 0 {
				__antithesis_instrumentation__.Notify(42395)
				if xf {
					__antithesis_instrumentation__.Notify(42397)
					node.Name.Cols = nil
				} else {
					__antithesis_instrumentation__.Notify(42398)
				}
				__antithesis_instrumentation__.Notify(42396)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42399)
			}
		}
		__antithesis_instrumentation__.Notify(42393)
		return 0
	})
	removeFroms = walkSQL("remove FROMs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42400)
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42402)
			n := len(node.From.Tables)
			if xfi < len(node.From.Tables) {
				__antithesis_instrumentation__.Notify(42404)
				node.From.Tables = append(node.From.Tables[:xfi], node.From.Tables[xfi+1:]...)
			} else {
				__antithesis_instrumentation__.Notify(42405)
			}
			__antithesis_instrumentation__.Notify(42403)
			return n
		}
		__antithesis_instrumentation__.Notify(42401)
		return 0
	})
	removeJoins = walkSQL("remove JOINs", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42406)

		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42408)
			idx := xfi / 2
			n := 0
			for i, t := range node.From.Tables {
				__antithesis_instrumentation__.Notify(42410)
				switch t := t.(type) {
				case *tree.JoinTableExpr:
					__antithesis_instrumentation__.Notify(42411)
					if n == idx {
						__antithesis_instrumentation__.Notify(42413)
						if xfi%2 == 0 {
							__antithesis_instrumentation__.Notify(42414)
							node.From.Tables[i] = t.Left
						} else {
							__antithesis_instrumentation__.Notify(42415)
							node.From.Tables[i] = t.Right
						}
					} else {
						__antithesis_instrumentation__.Notify(42416)
					}
					__antithesis_instrumentation__.Notify(42412)
					n += 2
				}
			}
			__antithesis_instrumentation__.Notify(42409)
			return n
		}
		__antithesis_instrumentation__.Notify(42407)
		return 0
	})
	simplifyVal = walkSQL("simplify vals", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42417)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.StrVal:
			__antithesis_instrumentation__.Notify(42419)
			if node.RawString() != "" {
				__antithesis_instrumentation__.Notify(42421)
				if xf {
					__antithesis_instrumentation__.Notify(42423)
					*node = *tree.NewStrVal("")
				} else {
					__antithesis_instrumentation__.Notify(42424)
				}
				__antithesis_instrumentation__.Notify(42422)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42425)
			}
		case *tree.NumVal:
			__antithesis_instrumentation__.Notify(42420)
			if node.OrigString() != "0" {
				__antithesis_instrumentation__.Notify(42426)
				if xf {
					__antithesis_instrumentation__.Notify(42428)
					*node = *tree.NewNumVal(constant.MakeInt64(0), "0", false)
				} else {
					__antithesis_instrumentation__.Notify(42429)
				}
				__antithesis_instrumentation__.Notify(42427)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42430)
			}
		}
		__antithesis_instrumentation__.Notify(42418)
		return 0
	})
	removeWhere = walkSQL("remove WHERE", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42431)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42433)
			if node.Where != nil {
				__antithesis_instrumentation__.Notify(42434)
				if xf {
					__antithesis_instrumentation__.Notify(42436)
					node.Where = nil
				} else {
					__antithesis_instrumentation__.Notify(42437)
				}
				__antithesis_instrumentation__.Notify(42435)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42438)
			}
		}
		__antithesis_instrumentation__.Notify(42432)
		return 0
	})
	removeHaving = walkSQL("remove HAVING", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42439)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42441)
			if node.Having != nil {
				__antithesis_instrumentation__.Notify(42442)
				if xf {
					__antithesis_instrumentation__.Notify(42444)
					node.Having = nil
				} else {
					__antithesis_instrumentation__.Notify(42445)
				}
				__antithesis_instrumentation__.Notify(42443)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42446)
			}
		}
		__antithesis_instrumentation__.Notify(42440)
		return 0
	})
	removeDistinct = walkSQL("remove DISTINCT", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42447)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.SelectClause:
			__antithesis_instrumentation__.Notify(42449)
			if node.Distinct {
				__antithesis_instrumentation__.Notify(42450)
				if xf {
					__antithesis_instrumentation__.Notify(42452)
					node.Distinct = false
				} else {
					__antithesis_instrumentation__.Notify(42453)
				}
				__antithesis_instrumentation__.Notify(42451)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42454)
			}
		}
		__antithesis_instrumentation__.Notify(42448)
		return 0
	})
	unparenthesize = walkSQL("unparenthesize", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42455)
		switch node := node.(type) {
		case tree.Exprs:
			__antithesis_instrumentation__.Notify(42457)
			n := 0
			for i, x := range node {
				__antithesis_instrumentation__.Notify(42459)
				if x, ok := x.(*tree.ParenExpr); ok {
					__antithesis_instrumentation__.Notify(42460)
					if n == xfi {
						__antithesis_instrumentation__.Notify(42462)
						node[i] = x.Expr
					} else {
						__antithesis_instrumentation__.Notify(42463)
					}
					__antithesis_instrumentation__.Notify(42461)
					n++
				} else {
					__antithesis_instrumentation__.Notify(42464)
				}
			}
			__antithesis_instrumentation__.Notify(42458)
			return n
		}
		__antithesis_instrumentation__.Notify(42456)
		return 0
	})
	nullifyFuncArgs = walkSQL("nullify function args", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42465)
		switch node := node.(type) {
		case *tree.FuncExpr:
			__antithesis_instrumentation__.Notify(42467)
			n := 0
			for i, x := range node.Exprs {
				__antithesis_instrumentation__.Notify(42469)
				if x != tree.DNull {
					__antithesis_instrumentation__.Notify(42470)
					if n == xfi {
						__antithesis_instrumentation__.Notify(42472)
						node.Exprs[i] = tree.DNull
					} else {
						__antithesis_instrumentation__.Notify(42473)
					}
					__antithesis_instrumentation__.Notify(42471)
					n++
				} else {
					__antithesis_instrumentation__.Notify(42474)
				}
			}
			__antithesis_instrumentation__.Notify(42468)
			return n
		}
		__antithesis_instrumentation__.Notify(42466)
		return 0
	})
	simplifyOnCond = walkSQL("simplify ON conditions", func(xfi int, node interface{}) int {
		__antithesis_instrumentation__.Notify(42475)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.OnJoinCond:
			__antithesis_instrumentation__.Notify(42477)
			if node.Expr != tree.DBoolTrue {
				__antithesis_instrumentation__.Notify(42478)
				if xf {
					__antithesis_instrumentation__.Notify(42480)
					node.Expr = tree.DBoolTrue
				} else {
					__antithesis_instrumentation__.Notify(42481)
				}
				__antithesis_instrumentation__.Notify(42479)
				return 1
			} else {
				__antithesis_instrumentation__.Notify(42482)
			}
		}
		__antithesis_instrumentation__.Notify(42476)
		return 0
	})

	removeStatement = replaceTopStatement("remove statements", func(xfi int, node interface{}) (int, tree.NodeFormatter) {
		__antithesis_instrumentation__.Notify(42483)
		xf := xfi == 0
		if _, ok := node.(tree.Statement); ok {
			__antithesis_instrumentation__.Notify(42485)
			if xf {
				__antithesis_instrumentation__.Notify(42487)
				return 1, emptyStatement{}
			} else {
				__antithesis_instrumentation__.Notify(42488)
			}
			__antithesis_instrumentation__.Notify(42486)
			return 1, nil
		} else {
			__antithesis_instrumentation__.Notify(42489)
		}
		__antithesis_instrumentation__.Notify(42484)
		return 0, nil
	})
	replaceStmt = replaceStatement("replace statements", func(xfi int, node interface{}) (int, tree.NodeFormatter) {
		__antithesis_instrumentation__.Notify(42490)
		xf := xfi == 0
		switch node := node.(type) {
		case *tree.ParenSelect:
			__antithesis_instrumentation__.Notify(42492)
			if xf {
				__antithesis_instrumentation__.Notify(42498)
				return 1, node.Select
			} else {
				__antithesis_instrumentation__.Notify(42499)
			}
			__antithesis_instrumentation__.Notify(42493)
			return 1, nil
		case *tree.Subquery:
			__antithesis_instrumentation__.Notify(42494)
			if xf {
				__antithesis_instrumentation__.Notify(42500)
				return 1, node.Select
			} else {
				__antithesis_instrumentation__.Notify(42501)
			}
			__antithesis_instrumentation__.Notify(42495)
			return 1, nil
		case *tree.With:
			__antithesis_instrumentation__.Notify(42496)
			n := len(node.CTEList)
			if xfi < len(node.CTEList) {
				__antithesis_instrumentation__.Notify(42502)
				return n, node.CTEList[xfi].Stmt
			} else {
				__antithesis_instrumentation__.Notify(42503)
			}
			__antithesis_instrumentation__.Notify(42497)
			return n, nil
		}
		__antithesis_instrumentation__.Notify(42491)
		return 0, nil
	})

	removeCastsRE = regexp.MustCompile(`:::?[a-zA-Z0-9]+`)
	removeCasts   = reduce.MakeIntPass("remove casts", func(s string, i int) (string, bool, error) {
		__antithesis_instrumentation__.Notify(42504)
		out := removeCastsRE.ReplaceAllStringFunc(s, func(found string) string {
			__antithesis_instrumentation__.Notify(42506)
			i--
			if i == -1 {
				__antithesis_instrumentation__.Notify(42508)
				return ""
			} else {
				__antithesis_instrumentation__.Notify(42509)
			}
			__antithesis_instrumentation__.Notify(42507)
			return found
		})
		__antithesis_instrumentation__.Notify(42505)
		return out, i < 0, nil
	})
	removeAliasesRE = regexp.MustCompile(`\sAS\s+\w+`)
	removeAliases   = reduce.MakeIntPass("remove aliases", func(s string, i int) (string, bool, error) {
		__antithesis_instrumentation__.Notify(42510)
		out := removeAliasesRE.ReplaceAllStringFunc(s, func(found string) string {
			__antithesis_instrumentation__.Notify(42512)
			i--
			if i == -1 {
				__antithesis_instrumentation__.Notify(42514)
				return ""
			} else {
				__antithesis_instrumentation__.Notify(42515)
			}
			__antithesis_instrumentation__.Notify(42513)
			return found
		})
		__antithesis_instrumentation__.Notify(42511)
		return out, i < 0, nil
	})
	removeDBSchemaRE = regexp.MustCompile(`\w+\.\w+\.`)
	removeDBSchema   = reduce.MakeIntPass("remove DB schema", func(s string, i int) (string, bool, error) {
		__antithesis_instrumentation__.Notify(42516)

		out := removeDBSchemaRE.ReplaceAllStringFunc(s, func(found string) string {
			__antithesis_instrumentation__.Notify(42518)
			i--
			if i == -1 {
				__antithesis_instrumentation__.Notify(42520)
				return ""
			} else {
				__antithesis_instrumentation__.Notify(42521)
			}
			__antithesis_instrumentation__.Notify(42519)
			return found
		})
		__antithesis_instrumentation__.Notify(42517)
		return out, i < 0, nil
	})
)

func skipParenSelect(stmt tree.SelectStatement) tree.SelectStatement {
	__antithesis_instrumentation__.Notify(42522)
	for {
		__antithesis_instrumentation__.Notify(42523)
		ps, ok := stmt.(*tree.ParenSelect)
		if !ok {
			__antithesis_instrumentation__.Notify(42525)
			return stmt
		} else {
			__antithesis_instrumentation__.Notify(42526)
		}
		__antithesis_instrumentation__.Notify(42524)
		stmt = ps.Select.Select
	}

}

func removeValuesCol(values *tree.ValuesClause, col int) {
	__antithesis_instrumentation__.Notify(42527)
	for i, row := range values.Rows {
		__antithesis_instrumentation__.Notify(42528)
		values.Rows[i] = append(row[:col], row[col+1:]...)
	}
}

type emptyStatement struct{}

func (e emptyStatement) Format(*tree.FmtCtx) { __antithesis_instrumentation__.Notify(42529) }

type SQLChunkReducer struct {
	maxConsecutiveFailures int
	asts                   []tree.NodeFormatter
}

func NewSQLChunkReducer(maxConsecutiveFailures int) *SQLChunkReducer {
	__antithesis_instrumentation__.Notify(42530)
	return &SQLChunkReducer{
		maxConsecutiveFailures: maxConsecutiveFailures,
	}
}

func (smr *SQLChunkReducer) HaltAfter() int {
	__antithesis_instrumentation__.Notify(42531)
	return smr.maxConsecutiveFailures
}

func (smr *SQLChunkReducer) Init(s string) error {
	__antithesis_instrumentation__.Notify(42532)
	stmts, err := parser.Parse(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(42534)
		return err
	} else {
		__antithesis_instrumentation__.Notify(42535)
	}
	__antithesis_instrumentation__.Notify(42533)
	smr.asts = collectASTs(stmts)
	return nil
}

func (smr SQLChunkReducer) NumSegments() int {
	__antithesis_instrumentation__.Notify(42536)
	return len(smr.asts)
}

func (smr SQLChunkReducer) DeleteSegments(start, end int) string {
	__antithesis_instrumentation__.Notify(42537)
	asts := append(smr.asts[:start], smr.asts[end:]...)
	return joinASTs(asts)
}
