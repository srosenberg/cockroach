package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
)

func TryDelegate(
	ctx context.Context, catalog cat.Catalog, evalCtx *tree.EvalContext, stmt tree.Statement,
) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465388)
	d := delegator{
		ctx:     ctx,
		catalog: catalog,
		evalCtx: evalCtx,
	}
	switch t := stmt.(type) {
	case *tree.ShowClusterSettingList:
		__antithesis_instrumentation__.Notify(465389)
		return d.delegateShowClusterSettingList(t)

	case *tree.ShowTenantClusterSettingList:
		__antithesis_instrumentation__.Notify(465390)
		return d.delegateShowTenantClusterSettingList(t)

	case *tree.ShowDatabases:
		__antithesis_instrumentation__.Notify(465391)
		return d.delegateShowDatabases(t)

	case *tree.ShowEnums:
		__antithesis_instrumentation__.Notify(465392)
		return d.delegateShowEnums(t)

	case *tree.ShowTypes:
		__antithesis_instrumentation__.Notify(465393)
		return d.delegateShowTypes()

	case *tree.ShowCreate:
		__antithesis_instrumentation__.Notify(465394)
		return d.delegateShowCreate(t)

	case *tree.ShowCreateAllSchemas:
		__antithesis_instrumentation__.Notify(465395)
		return d.delegateShowCreateAllSchemas()

	case *tree.ShowCreateAllTables:
		__antithesis_instrumentation__.Notify(465396)
		return d.delegateShowCreateAllTables()

	case *tree.ShowCreateAllTypes:
		__antithesis_instrumentation__.Notify(465397)
		return d.delegateShowCreateAllTypes()

	case *tree.ShowDatabaseIndexes:
		__antithesis_instrumentation__.Notify(465398)
		return d.delegateShowDatabaseIndexes(t)

	case *tree.ShowIndexes:
		__antithesis_instrumentation__.Notify(465399)
		return d.delegateShowIndexes(t)

	case *tree.ShowColumns:
		__antithesis_instrumentation__.Notify(465400)
		return d.delegateShowColumns(t)

	case *tree.ShowConstraints:
		__antithesis_instrumentation__.Notify(465401)
		return d.delegateShowConstraints(t)

	case *tree.ShowPartitions:
		__antithesis_instrumentation__.Notify(465402)
		return d.delegateShowPartitions(t)

	case *tree.ShowGrants:
		__antithesis_instrumentation__.Notify(465403)
		return d.delegateShowGrants(t)

	case *tree.ShowJobs:
		__antithesis_instrumentation__.Notify(465404)
		return d.delegateShowJobs(t)

	case *tree.ShowChangefeedJobs:
		__antithesis_instrumentation__.Notify(465405)
		return d.delegateShowChangefeedJobs(t)

	case *tree.ShowQueries:
		__antithesis_instrumentation__.Notify(465406)
		return d.delegateShowQueries(t)

	case *tree.ShowRanges:
		__antithesis_instrumentation__.Notify(465407)
		return d.delegateShowRanges(t)

	case *tree.ShowRangeForRow:
		__antithesis_instrumentation__.Notify(465408)
		return d.delegateShowRangeForRow(t)

	case *tree.ShowSurvivalGoal:
		__antithesis_instrumentation__.Notify(465409)
		return d.delegateShowSurvivalGoal(t)

	case *tree.ShowRegions:
		__antithesis_instrumentation__.Notify(465410)
		return d.delegateShowRegions(t)

	case *tree.ShowRoleGrants:
		__antithesis_instrumentation__.Notify(465411)
		return d.delegateShowRoleGrants(t)

	case *tree.ShowRoles:
		__antithesis_instrumentation__.Notify(465412)
		return d.delegateShowRoles()

	case *tree.ShowSchemas:
		__antithesis_instrumentation__.Notify(465413)
		return d.delegateShowSchemas(t)

	case *tree.ShowSequences:
		__antithesis_instrumentation__.Notify(465414)
		return d.delegateShowSequences(t)

	case *tree.ShowSessions:
		__antithesis_instrumentation__.Notify(465415)
		return d.delegateShowSessions(t)

	case *tree.ShowSyntax:
		__antithesis_instrumentation__.Notify(465416)
		return d.delegateShowSyntax(t)

	case *tree.ShowTables:
		__antithesis_instrumentation__.Notify(465417)
		return d.delegateShowTables(t)

	case *tree.ShowTransactions:
		__antithesis_instrumentation__.Notify(465418)
		return d.delegateShowTransactions(t)

	case *tree.ShowUsers:
		__antithesis_instrumentation__.Notify(465419)
		return d.delegateShowRoles()

	case *tree.ShowVar:
		__antithesis_instrumentation__.Notify(465420)
		return d.delegateShowVar(t)

	case *tree.ShowZoneConfig:
		__antithesis_instrumentation__.Notify(465421)
		return d.delegateShowZoneConfig(t)

	case *tree.ShowTransactionStatus:
		__antithesis_instrumentation__.Notify(465422)
		return d.delegateShowVar(&tree.ShowVar{Name: "transaction_status"})

	case *tree.ShowSchedules:
		__antithesis_instrumentation__.Notify(465423)
		return d.delegateShowSchedules(t)

	case *tree.ShowCompletions:
		__antithesis_instrumentation__.Notify(465424)
		return d.delegateShowCompletions(t)

	case *tree.ControlJobsForSchedules:
		__antithesis_instrumentation__.Notify(465425)
		return d.delegateJobControl(ControlJobsDelegate{
			Schedules: t.Schedules,
			Command:   t.Command,
		})

	case *tree.ControlJobsOfType:
		__antithesis_instrumentation__.Notify(465426)
		return d.delegateJobControl(ControlJobsDelegate{
			Type:    t.Type,
			Command: t.Command,
		})

	case *tree.ShowFullTableScans:
		__antithesis_instrumentation__.Notify(465427)
		return d.delegateShowFullTableScans()

	case *tree.ShowDefaultPrivileges:
		__antithesis_instrumentation__.Notify(465428)
		return d.delegateShowDefaultPrivileges(t)

	case *tree.ShowLastQueryStatistics:
		__antithesis_instrumentation__.Notify(465429)
		return nil, unimplemented.New(
			"show last query statistics",
			"cannot use SHOW LAST QUERY STATISTICS as a statement source",
		)

	case *tree.ShowSavepointStatus:
		__antithesis_instrumentation__.Notify(465430)
		return nil, unimplemented.NewWithIssue(
			47333, "cannot use SHOW SAVEPOINT STATUS as a statement source")

	case *tree.ShowTransferState:
		__antithesis_instrumentation__.Notify(465431)
		return nil, pgerror.Newf(pgcode.FeatureNotSupported,
			"cannot use SHOW TRANSFER STATE as a statement source")

	default:
		__antithesis_instrumentation__.Notify(465432)
		return nil, nil
	}
}

type delegator struct {
	ctx     context.Context
	catalog cat.Catalog
	evalCtx *tree.EvalContext
}

func parse(sql string) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465433)
	s, err := parser.ParseOne(sql)
	return s.AST, err
}
