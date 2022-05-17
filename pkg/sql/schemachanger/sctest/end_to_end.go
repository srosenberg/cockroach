// Package sctest contains tools to run end-to-end datadriven tests in both
// ccl and non-ccl settings.
package sctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps/sctestutils"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/datadriven"
	"github.com/stretchr/testify/require"
)

type NewClusterFunc func(
	t *testing.T, knobs *scrun.TestingKnobs,
) (_ *gosql.DB, cleanup func())

func SingleNodeCluster(t *testing.T, knobs *scrun.TestingKnobs) (*gosql.DB, func()) {
	__antithesis_instrumentation__.Notify(595420)
	s, db, _ := serverutils.StartServer(t, base.TestServerArgs{
		Knobs: base.TestingKnobs{
			SQLDeclarativeSchemaChanger: knobs,
			JobsTestingKnobs:            jobs.NewTestingKnobsWithShortIntervals(),
		},
	})
	return db, func() {
		__antithesis_instrumentation__.Notify(595421)
		s.Stopper().Stop(context.Background())
	}
}

func EndToEndSideEffects(t *testing.T, dir string, newCluster NewClusterFunc) {
	__antithesis_instrumentation__.Notify(595422)
	ctx := context.Background()
	datadriven.Walk(t, dir, func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(595423)

		db, cleanup := newCluster(t, nil)
		tdb := sqlutils.MakeSQLRunner(db)
		defer cleanup()
		datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
			__antithesis_instrumentation__.Notify(595424)
			stmts, err := parser.Parse(d.Input)
			require.NoError(t, err)
			require.NotEmpty(t, stmts)
			execStmts := func() {
				__antithesis_instrumentation__.Notify(595426)
				for _, stmt := range stmts {
					__antithesis_instrumentation__.Notify(595428)
					tdb.Exec(t, stmt.SQL)
				}
				__antithesis_instrumentation__.Notify(595427)
				waitForSchemaChangesToSucceed(t, tdb)
			}
			__antithesis_instrumentation__.Notify(595425)

			switch d.Cmd {
			case "setup":
				__antithesis_instrumentation__.Notify(595429)
				a := prettyNamespaceDump(t, tdb)
				execStmts()
				b := prettyNamespaceDump(t, tdb)
				return sctestutils.Diff(a, b, sctestutils.DiffArgs{CompactLevel: 1})

			case "test":
				__antithesis_instrumentation__.Notify(595430)
				require.Len(t, stmts, 1)
				stmt := stmts[0]

				defer execStmts()

				sctestdeps.WaitForNoRunningSchemaChanges(t, tdb)
				var deps *sctestdeps.TestState

				deps = sctestdeps.NewTestDependencies(
					sctestdeps.WithDescriptors(sctestdeps.ReadDescriptorsFromDB(ctx, t, tdb).Catalog),
					sctestdeps.WithNamespace(sctestdeps.ReadNamespaceFromDB(t, tdb).Catalog),
					sctestdeps.WithCurrentDatabase(sctestdeps.ReadCurrentDatabaseFromDB(t, tdb)),
					sctestdeps.WithSessionData(sctestdeps.ReadSessionDataFromDB(t, tdb, func(
						sd *sessiondata.SessionData,
					) {
						__antithesis_instrumentation__.Notify(595433)

						sd.NewSchemaChangerMode = sessiondatapb.UseNewSchemaChangerUnsafe
						sd.ApplicationName = ""
					})),
					sctestdeps.WithTestingKnobs(&scrun.TestingKnobs{
						BeforeStage: func(p scplan.Plan, stageIdx int) error {
							__antithesis_instrumentation__.Notify(595434)
							deps.LogSideEffectf("## %s", p.Stages[stageIdx].String())
							return nil
						},
					}),
					sctestdeps.WithStatements(stmt.SQL))
				__antithesis_instrumentation__.Notify(595431)
				execStatementWithTestDeps(ctx, t, deps, stmt)
				return replaceNonDeterministicOutput(deps.SideEffectLog())

			default:
				__antithesis_instrumentation__.Notify(595432)
				return fmt.Sprintf("unknown command: %s", d.Cmd)
			}
		})
	})
}

var scheduleIDRegexp = regexp.MustCompile(`scheduleId: "?[0-9]+"?`)

var dropTimeRegexp = regexp.MustCompile("dropTime: \"[0-9]+")

func replaceNonDeterministicOutput(text string) string {
	__antithesis_instrumentation__.Notify(595435)

	nextString := scheduleIDRegexp.ReplaceAllString(text, "scheduleId: <redacted>")
	return dropTimeRegexp.ReplaceAllString(nextString, "dropTime: <redacted>")
}

func execStatementWithTestDeps(
	ctx context.Context, t *testing.T, deps *sctestdeps.TestState, stmt parser.Statement,
) {
	__antithesis_instrumentation__.Notify(595436)
	state, err := scbuild.Build(ctx, deps, scpb.CurrentState{}, stmt.AST)
	require.NoError(t, err, "error in builder")

	var jobID jobspb.JobID
	deps.WithTxn(func(s *sctestdeps.TestState) {
		__antithesis_instrumentation__.Notify(595438)

		deps.IncrementPhase()
		deps.LogSideEffectf("# begin %s", deps.Phase())
		state, _, err = scrun.RunStatementPhase(ctx, s.TestingKnobs(), s, state)
		require.NoError(t, err, "error in %s", s.Phase())
		deps.LogSideEffectf("# end %s", deps.Phase())

		deps.IncrementPhase()
		deps.LogSideEffectf("# begin %s", deps.Phase())
		state, jobID, err = scrun.RunPreCommitPhase(ctx, s.TestingKnobs(), s, state)
		require.NoError(t, err, "error in %s", s.Phase())
		deps.LogSideEffectf("# end %s", deps.Phase())
	})
	__antithesis_instrumentation__.Notify(595437)

	if job := deps.JobRecord(jobID); job != nil {
		__antithesis_instrumentation__.Notify(595439)

		deps.IncrementPhase()
		deps.LogSideEffectf("# begin %s", deps.Phase())
		const rollback = false
		err = scrun.RunSchemaChangesInJob(
			ctx, deps.TestingKnobs(), deps.ClusterSettings(), deps, jobID, job.DescriptorIDs, rollback,
		)
		require.NoError(t, err, "error in mock schema change job execution")
		deps.LogSideEffectf("# end %s", deps.Phase())
	} else {
		__antithesis_instrumentation__.Notify(595440)
	}
}

func prettyNamespaceDump(t *testing.T, tdb *sqlutils.SQLRunner) string {
	__antithesis_instrumentation__.Notify(595441)
	rows := tdb.QueryStr(t, fmt.Sprintf(`
		SELECT "parentID", "parentSchemaID", name, id
		FROM system.namespace
		WHERE id NOT IN (%d, %d) AND "parentID" <> %d
		ORDER BY id ASC`,
		keys.PublicSchemaID,
		keys.SystemDatabaseID,
		keys.SystemDatabaseID,
	))
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		__antithesis_instrumentation__.Notify(595443)
		parentID, parentSchemaID, name, id := row[0], row[1], row[2], row[3]
		nameType := "object"
		if parentSchemaID == "0" {
			__antithesis_instrumentation__.Notify(595445)
			if parentID == "0" {
				__antithesis_instrumentation__.Notify(595446)
				nameType = "database"
			} else {
				__antithesis_instrumentation__.Notify(595447)
				nameType = "schema"
			}
		} else {
			__antithesis_instrumentation__.Notify(595448)
		}
		__antithesis_instrumentation__.Notify(595444)
		line := fmt.Sprintf("%s {%s %s %s} -> %s", nameType, parentID, parentSchemaID, name, id)
		lines = append(lines, line)
	}
	__antithesis_instrumentation__.Notify(595442)
	return strings.Join(lines, "\n")
}

func waitForSchemaChangesToSucceed(t *testing.T, tdb *sqlutils.SQLRunner) {
	__antithesis_instrumentation__.Notify(595449)
	tdb.CheckQueryResultsRetry(
		t, schemaChangeWaitQuery(`('succeeded')`), [][]string{},
	)
}

func waitForSchemaChangesToFinish(t *testing.T, tdb *sqlutils.SQLRunner) {
	__antithesis_instrumentation__.Notify(595450)
	tdb.CheckQueryResultsRetry(
		t, schemaChangeWaitQuery(`('succeeded', 'failed')`), [][]string{},
	)
}

func schemaChangeWaitQuery(statusInString string) string {
	__antithesis_instrumentation__.Notify(595451)
	q := fmt.Sprintf(
		`SELECT status, job_type, description FROM [SHOW JOBS] WHERE job_type IN ('%s', '%s', '%s') AND status NOT IN %s`,
		jobspb.TypeSchemaChange,
		jobspb.TypeTypeSchemaChange,
		jobspb.TypeNewSchemaChange,
		statusInString,
	)
	return q
}
