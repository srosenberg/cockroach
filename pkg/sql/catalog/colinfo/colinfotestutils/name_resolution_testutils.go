package colinfotestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/testutils"
)

type ColumnItemResolverTester interface {
	GetColumnItemResolver() colinfo.ColumnItemResolver

	AddTable(tabName tree.TableName, colNames []tree.Name)

	ResolveQualifiedStarTestResults(
		srcName *tree.TableName, srcMeta colinfo.ColumnSourceMeta,
	) (string, string, error)

	ResolveColumnItemTestResults(colRes colinfo.ColumnResolutionResult) (string, error)
}

func initColumnItemResolverTester(t *testing.T, ct ColumnItemResolverTester) {
	__antithesis_instrumentation__.Notify(250949)
	ct.AddTable(tree.MakeTableNameWithSchema("", "crdb_internal", "tables"), []tree.Name{"table_name"})
	ct.AddTable(tree.MakeTableNameWithSchema("db1", tree.PublicSchemaName, "foo"), []tree.Name{"x"})
	ct.AddTable(tree.MakeTableNameWithSchema("db2", tree.PublicSchemaName, "foo"), []tree.Name{"x"})
	ct.AddTable(tree.MakeUnqualifiedTableName("bar"), []tree.Name{"x"})
	ct.AddTable(tree.MakeTableNameWithSchema("db1", tree.PublicSchemaName, "kv"), []tree.Name{"k", "v"})
}

func RunResolveQualifiedStarTest(t *testing.T, ct ColumnItemResolverTester) {
	__antithesis_instrumentation__.Notify(250950)
	testCases := []struct {
		in    string
		tnout string
		csout string
		err   string
	}{
		{`a.*`, ``, ``, `no data source matches pattern: a.*`},
		{`foo.*`, ``, ``, `ambiguous source name: "foo"`},
		{`db1.public.foo.*`, `db1.public.foo`, `x`, ``},
		{`db1.foo.*`, `db1.public.foo`, `x`, ``},
		{`dbx.foo.*`, ``, ``, `no data source matches pattern: dbx.foo.*`},
		{`kv.*`, `db1.public.kv`, `k, v`, ``},
	}

	initColumnItemResolverTester(t, ct)
	resolver := ct.GetColumnItemResolver()
	for _, tc := range testCases {
		__antithesis_instrumentation__.Notify(250951)
		t.Run(tc.in, func(t *testing.T) {
			__antithesis_instrumentation__.Notify(250952)
			tnout, csout, err := func() (string, string, error) {
				__antithesis_instrumentation__.Notify(250957)
				stmt, err := parser.ParseOne(fmt.Sprintf("SELECT %s", tc.in))
				if err != nil {
					__antithesis_instrumentation__.Notify(250962)
					return "", "", err
				} else {
					__antithesis_instrumentation__.Notify(250963)
				}
				__antithesis_instrumentation__.Notify(250958)
				v := stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs[0].Expr.(tree.VarName)
				c, err := v.NormalizeVarName()
				if err != nil {
					__antithesis_instrumentation__.Notify(250964)
					return "", "", err
				} else {
					__antithesis_instrumentation__.Notify(250965)
				}
				__antithesis_instrumentation__.Notify(250959)
				acs, ok := c.(*tree.AllColumnsSelector)
				if !ok {
					__antithesis_instrumentation__.Notify(250966)
					return "", "", fmt.Errorf("var name %s (%T) did not resolve to AllColumnsSelector, found %T instead",
						v, v, c)
				} else {
					__antithesis_instrumentation__.Notify(250967)
				}
				__antithesis_instrumentation__.Notify(250960)
				tn, res, err := colinfo.ResolveAllColumnsSelector(context.Background(), resolver, acs)
				if err != nil {
					__antithesis_instrumentation__.Notify(250968)
					return "", "", err
				} else {
					__antithesis_instrumentation__.Notify(250969)
				}
				__antithesis_instrumentation__.Notify(250961)
				return ct.ResolveQualifiedStarTestResults(tn, res)
			}()
			__antithesis_instrumentation__.Notify(250953)
			if !testutils.IsError(err, tc.err) {
				__antithesis_instrumentation__.Notify(250970)
				t.Fatalf("%s: expected %s, but found %v", tc.in, tc.err, err)
			} else {
				__antithesis_instrumentation__.Notify(250971)
			}
			__antithesis_instrumentation__.Notify(250954)
			if tc.err != "" {
				__antithesis_instrumentation__.Notify(250972)
				return
			} else {
				__antithesis_instrumentation__.Notify(250973)
			}
			__antithesis_instrumentation__.Notify(250955)

			if tc.tnout != tnout {
				__antithesis_instrumentation__.Notify(250974)
				t.Fatalf("%s: expected tn %s, but found %s", tc.in, tc.tnout, tnout)
			} else {
				__antithesis_instrumentation__.Notify(250975)
			}
			__antithesis_instrumentation__.Notify(250956)
			if tc.csout != csout {
				__antithesis_instrumentation__.Notify(250976)
				t.Fatalf("%s: expected cs %s, but found %s", tc.in, tc.csout, csout)
			} else {
				__antithesis_instrumentation__.Notify(250977)
			}
		})
	}
}

func RunResolveColumnItemTest(t *testing.T, ct ColumnItemResolverTester) {
	__antithesis_instrumentation__.Notify(250978)
	testCases := []struct {
		in  string
		out string
		err string
	}{
		{`a`, ``, `column "a" does not exist`},
		{`x`, ``, `column reference "x" is ambiguous \(candidates: db1.public.foo.x, db2.public.foo.x, bar.x\)`},
		{`k`, `db1.public.kv.k`, ``},
		{`v`, `db1.public.kv.v`, ``},
		{`table_name`, `"".crdb_internal.tables.table_name`, ``},

		{`blix.x`, ``, `no data source matches prefix: blix`},
		{`"".x`, ``, `invalid column name: ""\.x`},
		{`foo.x`, ``, `ambiguous source name`},
		{`kv.k`, `db1.public.kv.k`, ``},
		{`bar.x`, `bar.x`, ``},
		{`tables.table_name`, `"".crdb_internal.tables.table_name`, ``},

		{`a.b.x`, ``, `no data source matches prefix: a\.b`},
		{`crdb_internal.tables.table_name`, `"".crdb_internal.tables.table_name`, ``},
		{`public.foo.x`, ``, `ambiguous source name`},
		{`public.kv.k`, `db1.public.kv.k`, ``},

		{`db1.foo.x`, `db1.public.foo.x`, ``},
		{`db2.foo.x`, `db2.public.foo.x`, ``},

		{`a.b.c.x`, ``, `no data source matches prefix: a\.b\.c`},
		{`"".crdb_internal.tables.table_name`, `"".crdb_internal.tables.table_name`, ``},
		{`db1.public.foo.x`, `db1.public.foo.x`, ``},
		{`db2.public.foo.x`, `db2.public.foo.x`, ``},
		{`db1.public.kv.v`, `db1.public.kv.v`, ``},
	}

	initColumnItemResolverTester(t, ct)
	resolver := ct.GetColumnItemResolver()
	for _, tc := range testCases {
		__antithesis_instrumentation__.Notify(250979)
		t.Run(tc.in, func(t *testing.T) {
			__antithesis_instrumentation__.Notify(250980)
			out, err := func() (string, error) {
				__antithesis_instrumentation__.Notify(250984)
				stmt, err := parser.ParseOne(fmt.Sprintf("SELECT %s", tc.in))
				if err != nil {
					__antithesis_instrumentation__.Notify(250989)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(250990)
				}
				__antithesis_instrumentation__.Notify(250985)
				v := stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs[0].Expr.(tree.VarName)
				c, err := v.NormalizeVarName()
				if err != nil {
					__antithesis_instrumentation__.Notify(250991)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(250992)
				}
				__antithesis_instrumentation__.Notify(250986)
				ci, ok := c.(*tree.ColumnItem)
				if !ok {
					__antithesis_instrumentation__.Notify(250993)
					return "", fmt.Errorf("var name %s (%T) did not resolve to ColumnItem, found %T instead",
						v, v, c)
				} else {
					__antithesis_instrumentation__.Notify(250994)
				}
				__antithesis_instrumentation__.Notify(250987)
				res, err := colinfo.ResolveColumnItem(context.Background(), resolver, ci)
				if err != nil {
					__antithesis_instrumentation__.Notify(250995)
					return "", err
				} else {
					__antithesis_instrumentation__.Notify(250996)
				}
				__antithesis_instrumentation__.Notify(250988)
				return ct.ResolveColumnItemTestResults(res)
			}()
			__antithesis_instrumentation__.Notify(250981)
			if !testutils.IsError(err, tc.err) {
				__antithesis_instrumentation__.Notify(250997)
				t.Fatalf("%s: expected %s, but found %v", tc.in, tc.err, err)
			} else {
				__antithesis_instrumentation__.Notify(250998)
			}
			__antithesis_instrumentation__.Notify(250982)
			if tc.err != "" {
				__antithesis_instrumentation__.Notify(250999)
				return
			} else {
				__antithesis_instrumentation__.Notify(251000)
			}
			__antithesis_instrumentation__.Notify(250983)

			if tc.out != out {
				__antithesis_instrumentation__.Notify(251001)
				t.Fatalf("%s: expected %s, but found %s", tc.in, tc.out, out)
			} else {
				__antithesis_instrumentation__.Notify(251002)
			}
		})
	}
}
