package sctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdecomp"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestutils"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/datadriven"
	"github.com/stretchr/testify/require"
)

func DecomposeToElements(t *testing.T, dir string, newCluster NewClusterFunc) {
	__antithesis_instrumentation__.Notify(595384)
	skip.UnderRace(t)
	skip.UnderStress(t)
	ctx := context.Background()
	datadriven.Walk(t, dir, func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(595385)

		db, cleanup := newCluster(t, nil)
		tdb := sqlutils.MakeSQLRunner(db)
		defer cleanup()

		tdb.Exec(t, `SET CLUSTER SETTING sql.defaults.use_declarative_schema_changer = 'off'`)
		datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
			__antithesis_instrumentation__.Notify(595386)
			return runDecomposeTest(ctx, t, d, tdb)
		})
	})
}

func runDecomposeTest(
	ctx context.Context, t *testing.T, d *datadriven.TestData, tdb *sqlutils.SQLRunner,
) string {
	__antithesis_instrumentation__.Notify(595387)
	switch d.Cmd {
	case "setup":
		__antithesis_instrumentation__.Notify(595388)
		stmts, err := parser.Parse(d.Input)
		require.NoError(t, err)
		require.NotEmpty(t, stmts, "missing statement(s) for setup command")
		for _, stmt := range stmts {
			__antithesis_instrumentation__.Notify(595394)
			tdb.Exec(t, stmt.SQL)
		}
		__antithesis_instrumentation__.Notify(595389)
		return ""

	case "decompose":
		__antithesis_instrumentation__.Notify(595390)
		fields := strings.Fields(d.Input)
		require.Lenf(t, fields, 1, "'decompose' requires one simple name, invalid input: %s", d.Input)
		name := fields[0]
		var desc catalog.Descriptor
		allDescs := sctestdeps.ReadDescriptorsFromDB(ctx, t, tdb)
		_ = allDescs.ForEachDescriptorEntry(func(d catalog.Descriptor) error {
			__antithesis_instrumentation__.Notify(595395)
			if d.GetName() == name {
				__antithesis_instrumentation__.Notify(595397)
				desc = d
			} else {
				__antithesis_instrumentation__.Notify(595398)
			}
			__antithesis_instrumentation__.Notify(595396)
			return nil
		})
		__antithesis_instrumentation__.Notify(595391)
		require.NotNilf(t, desc, "descriptor with name %q not found", name)
		m := make(map[scpb.Element]scpb.Status)
		visitor := func(status scpb.Status, element scpb.Element) {
			__antithesis_instrumentation__.Notify(595399)
			m[element] = status
		}
		__antithesis_instrumentation__.Notify(595392)
		backRefs := scdecomp.WalkDescriptor(desc, allDescs.LookupDescriptorEntry, visitor)
		return marshalResult(t, m, backRefs)

	default:
		__antithesis_instrumentation__.Notify(595393)
		return fmt.Sprintf("unknown command: %s", d.Cmd)
	}
}

func marshalResult(
	t *testing.T, m map[scpb.Element]scpb.Status, backRefs catalog.DescriptorIDSet,
) string {
	__antithesis_instrumentation__.Notify(595400)
	var b strings.Builder
	str := make(map[scpb.Element]string, len(m))
	rank := make(map[scpb.Element]int, len(m))
	elts := make([]scpb.Element, 0, len(m))
	for e := range m {
		__antithesis_instrumentation__.Notify(595405)
		{
			__antithesis_instrumentation__.Notify(595407)

			var ep scpb.ElementProto
			ep.SetValue(e)
			v := reflect.ValueOf(ep)
			for i := 0; i < v.NumField(); i++ {
				__antithesis_instrumentation__.Notify(595408)
				if !v.Field(i).IsNil() {
					__antithesis_instrumentation__.Notify(595409)
					rank[e] = i
					break
				} else {
					__antithesis_instrumentation__.Notify(595410)
				}
			}
		}
		__antithesis_instrumentation__.Notify(595406)
		elts = append(elts, e)
		yaml, err := sctestutils.ProtoToYAML(e)
		require.NoError(t, err)
		str[e] = yaml
	}
	__antithesis_instrumentation__.Notify(595401)
	sort.Slice(elts, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(595411)
		if d := rank[elts[i]] - rank[elts[j]]; d != 0 {
			__antithesis_instrumentation__.Notify(595413)
			return d < 0
		} else {
			__antithesis_instrumentation__.Notify(595414)
		}
		__antithesis_instrumentation__.Notify(595412)
		return str[elts[i]] < str[elts[j]]
	})
	__antithesis_instrumentation__.Notify(595402)
	b.WriteString("BackReferencedIDs:\n")
	for _, id := range backRefs.Ordered() {
		__antithesis_instrumentation__.Notify(595415)
		b.WriteString(fmt.Sprintf("  - %d\n", id))
	}
	__antithesis_instrumentation__.Notify(595403)
	b.WriteString("ElementState:\n")
	for _, e := range elts {
		__antithesis_instrumentation__.Notify(595416)
		typeName := fmt.Sprintf("%T", e)
		b.WriteString(strings.ReplaceAll(typeName, "*scpb.", "- ") + ":\n")
		b.WriteString(indentText(str[e], "    "))
		b.WriteString(fmt.Sprintf("  Status: %s\n", m[e].String()))
	}
	__antithesis_instrumentation__.Notify(595404)
	return b.String()
}

func indentText(input string, tab string) string {
	__antithesis_instrumentation__.Notify(595417)
	result := strings.Builder{}
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(595419)
		line := scanner.Text()
		result.WriteString(tab)
		result.WriteString(line)
		result.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(595418)
	return result.String()
}
