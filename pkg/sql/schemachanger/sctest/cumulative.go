package sctest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scrun"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/datadriven"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func cumulativeTest(
	t *testing.T, dir string, tf func(t *testing.T, setup, stmts []parser.Statement),
) {
	__antithesis_instrumentation__.Notify(595192)
	datadriven.Walk(t, dir, func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(595193)
		var setup []parser.Statement

		datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
			__antithesis_instrumentation__.Notify(595194)
			stmts, err := parser.Parse(d.Input)
			require.NoError(t, err)
			require.NotEmpty(t, stmts)

			switch d.Cmd {
			case "setup":
				__antithesis_instrumentation__.Notify(595196)

			case "test":
				__antithesis_instrumentation__.Notify(595197)
				var lines []string
				for _, stmt := range stmts {
					__antithesis_instrumentation__.Notify(595200)
					lines = append(lines, stmt.SQL)
				}
				__antithesis_instrumentation__.Notify(595198)
				t.Run(strings.Join(lines, "; "), func(t *testing.T) {
					__antithesis_instrumentation__.Notify(595201)
					tf(t, setup, stmts)
				})
			default:
				__antithesis_instrumentation__.Notify(595199)
				return fmt.Sprintf("unknown command: %s", d.Cmd)
			}
			__antithesis_instrumentation__.Notify(595195)
			setup = append(setup, stmts...)
			return d.Expected
		})
	})
}

func Rollback(t *testing.T, dir string, newCluster NewClusterFunc) {
	__antithesis_instrumentation__.Notify(595202)
	countRevertiblePostCommitStages := func(
		t *testing.T, setup, stmts []parser.Statement,
	) (n int) {
		__antithesis_instrumentation__.Notify(595206)
		processPlanInPhase(
			t, newCluster, setup, stmts, scop.PostCommitPhase,
			func(p scplan.Plan) { __antithesis_instrumentation__.Notify(595208); n = len(p.StagesForCurrentPhase()) },
			func(db *gosql.DB) {
				__antithesis_instrumentation__.Notify(595209)

			},
		)
		__antithesis_instrumentation__.Notify(595207)
		return n
	}
	__antithesis_instrumentation__.Notify(595203)
	var testRollbackCase func(
		t *testing.T, setup, stmts []parser.Statement, ord int,
	)
	testFunc := func(t *testing.T, setup, stmts []parser.Statement) {
		__antithesis_instrumentation__.Notify(595210)
		n := countRevertiblePostCommitStages(t, setup, stmts)
		if n == 0 {
			__antithesis_instrumentation__.Notify(595212)
			t.Logf("test case has no revertible post-commit stages, skipping...")
			return
		} else {
			__antithesis_instrumentation__.Notify(595213)
		}
		__antithesis_instrumentation__.Notify(595211)
		t.Logf("test case has %d revertible post-commit stages", n)
		for i := 1; i <= n; i++ {
			__antithesis_instrumentation__.Notify(595214)
			if !t.Run(
				fmt.Sprintf("rollback stage %d of %d", i, n),
				func(t *testing.T) {
					__antithesis_instrumentation__.Notify(595215)
					testRollbackCase(t, setup, stmts, i)
				},
			) {
				__antithesis_instrumentation__.Notify(595216)
				return
			} else {
				__antithesis_instrumentation__.Notify(595217)
			}
		}
	}
	__antithesis_instrumentation__.Notify(595204)

	testRollbackCase = func(
		t *testing.T, setup, stmts []parser.Statement, ord int,
	) {
		__antithesis_instrumentation__.Notify(595218)
		var numInjectedFailures uint32
		beforeStage := func(p scplan.Plan, stageIdx int) error {
			__antithesis_instrumentation__.Notify(595223)
			if atomic.LoadUint32(&numInjectedFailures) > 0 {
				__antithesis_instrumentation__.Notify(595226)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(595227)
			}
			__antithesis_instrumentation__.Notify(595224)
			s := p.Stages[stageIdx]
			if s.Phase == scop.PostCommitPhase && func() bool {
				__antithesis_instrumentation__.Notify(595228)
				return s.Ordinal == ord == true
			}() == true {
				__antithesis_instrumentation__.Notify(595229)
				atomic.AddUint32(&numInjectedFailures, 1)
				return errors.Errorf("boom %d", ord)
			} else {
				__antithesis_instrumentation__.Notify(595230)
			}
			__antithesis_instrumentation__.Notify(595225)
			return nil
		}
		__antithesis_instrumentation__.Notify(595219)

		db, cleanup := newCluster(t, &scrun.TestingKnobs{
			BeforeStage: beforeStage,
			OnPostCommitError: func(p scplan.Plan, stageIdx int, err error) error {
				__antithesis_instrumentation__.Notify(595231)
				if strings.Contains(err.Error(), "boom") {
					__antithesis_instrumentation__.Notify(595233)
					return err
				} else {
					__antithesis_instrumentation__.Notify(595234)
				}
				__antithesis_instrumentation__.Notify(595232)
				panic(fmt.Sprintf("%+v", err))
			},
		})
		__antithesis_instrumentation__.Notify(595220)
		defer cleanup()

		tdb := sqlutils.MakeSQLRunner(db)
		var before [][]string
		beforeFunc := func() {
			__antithesis_instrumentation__.Notify(595235)
			before = tdb.QueryStr(t, fetchDescriptorStateQuery)
		}
		__antithesis_instrumentation__.Notify(595221)
		onError := func(err error) error {
			__antithesis_instrumentation__.Notify(595236)

			require.Equal(t, before, tdb.QueryStr(t, fetchDescriptorStateQuery))
			return err
		}
		__antithesis_instrumentation__.Notify(595222)
		err := executeSchemaChangeTxn(
			context.Background(), t, setup, stmts, db, beforeFunc, nil, onError,
		)
		if atomic.LoadUint32(&numInjectedFailures) == 0 {
			__antithesis_instrumentation__.Notify(595237)
			require.NoError(t, err)
		} else {
			__antithesis_instrumentation__.Notify(595238)
			require.Regexp(t, fmt.Sprintf("boom %d", ord), err)
		}
	}
	__antithesis_instrumentation__.Notify(595205)
	cumulativeTest(t, dir, testFunc)
}

const fetchDescriptorStateQuery = `
SELECT
	create_statement
FROM
	( 
		SELECT descriptor_id, create_statement FROM crdb_internal.create_schema_statements
		UNION ALL SELECT descriptor_id, create_statement FROM crdb_internal.create_statements
		UNION ALL SELECT descriptor_id, create_statement FROM crdb_internal.create_type_statements
	)
WHERE descriptor_id IN (SELECT id FROM system.namespace)
ORDER BY
	create_statement;`

func Pause(t *testing.T, dir string, newCluster NewClusterFunc) {
	__antithesis_instrumentation__.Notify(595239)
	skip.UnderRace(t)
	var postCommit, nonRevertible int
	countStages := func(
		t *testing.T, setup, stmts []parser.Statement,
	) {
		__antithesis_instrumentation__.Notify(595243)
		processPlanInPhase(t, newCluster, setup, stmts, scop.PostCommitPhase, func(
			p scplan.Plan,
		) {
			__antithesis_instrumentation__.Notify(595244)
			postCommit = len(p.StagesForCurrentPhase())
			nonRevertible = len(p.Stages) - postCommit
		}, nil)
	}
	__antithesis_instrumentation__.Notify(595240)
	var testPauseCase func(
		t *testing.T, setup, stmts []parser.Statement, ord int,
	)
	testFunc := func(t *testing.T, setup, stmts []parser.Statement) {
		__antithesis_instrumentation__.Notify(595245)
		countStages(t, setup, stmts)
		n := postCommit + nonRevertible
		if n == 0 {
			__antithesis_instrumentation__.Notify(595247)
			t.Logf("test case has no revertible post-commit stages, skipping...")
			return
		} else {
			__antithesis_instrumentation__.Notify(595248)
		}
		__antithesis_instrumentation__.Notify(595246)
		t.Logf("test case has %d revertible post-commit stages", n)
		for i := 1; i <= n; i++ {
			__antithesis_instrumentation__.Notify(595249)
			if !t.Run(
				fmt.Sprintf("pause stage %d of %d", i, n),
				func(t *testing.T) { __antithesis_instrumentation__.Notify(595250); testPauseCase(t, setup, stmts, i) },
			) {
				__antithesis_instrumentation__.Notify(595251)
				return
			} else {
				__antithesis_instrumentation__.Notify(595252)
			}
		}
	}
	__antithesis_instrumentation__.Notify(595241)
	testPauseCase = func(t *testing.T, setup, stmts []parser.Statement, ord int) {
		__antithesis_instrumentation__.Notify(595253)
		var numInjectedFailures uint32

		db, cleanup := newCluster(t, &scrun.TestingKnobs{
			BeforeStage: func(p scplan.Plan, stageIdx int) error {
				__antithesis_instrumentation__.Notify(595256)
				if atomic.LoadUint32(&numInjectedFailures) > 0 {
					__antithesis_instrumentation__.Notify(595259)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(595260)
				}
				__antithesis_instrumentation__.Notify(595257)
				s := p.Stages[stageIdx]
				if s.Phase == scop.PostCommitPhase && func() bool {
					__antithesis_instrumentation__.Notify(595261)
					return s.Ordinal == ord == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(595262)
					return (s.Phase == scop.PostCommitNonRevertiblePhase && func() bool {
						__antithesis_instrumentation__.Notify(595263)
						return s.Ordinal+postCommit == ord == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(595264)
					atomic.AddUint32(&numInjectedFailures, 1)
					return jobs.MarkPauseRequestError(errors.Errorf("boom %d", ord))
				} else {
					__antithesis_instrumentation__.Notify(595265)
				}
				__antithesis_instrumentation__.Notify(595258)
				return nil
			},
		})
		__antithesis_instrumentation__.Notify(595254)
		defer cleanup()
		tdb := sqlutils.MakeSQLRunner(db)
		onError := func(err error) error {
			__antithesis_instrumentation__.Notify(595266)

			re := regexp.MustCompile(
				`job (\d+) was paused before it completed with reason: boom (\d+)`,
			)
			match := re.FindStringSubmatch(err.Error())
			require.NotNil(t, match)
			idx, err := strconv.Atoi(match[2])
			require.NoError(t, err)
			require.Equal(t, ord, idx)
			jobID, err := strconv.Atoi(match[1])
			require.NoError(t, err)
			t.Logf("found job %d", jobID)
			tdb.Exec(t, "RESUME JOB $1", jobID)
			tdb.CheckQueryResultsRetry(t, "SELECT status, error FROM [SHOW JOB "+match[1]+"]", [][]string{
				{"succeeded", ""},
			})
			return nil
		}
		__antithesis_instrumentation__.Notify(595255)
		require.NoError(t, executeSchemaChangeTxn(
			context.Background(), t, setup, stmts, db, nil, nil, onError,
		))
		require.Equal(t, uint32(1), atomic.LoadUint32(&numInjectedFailures))
	}
	__antithesis_instrumentation__.Notify(595242)
	cumulativeTest(t, dir, testFunc)
}

func Backup(t *testing.T, dir string, newCluster NewClusterFunc) {
	__antithesis_instrumentation__.Notify(595267)
	skip.UnderRace(t)
	skip.UnderStress(t)
	var after [][]string
	var dbName string
	countStages := func(
		t *testing.T, setup, stmts []parser.Statement,
	) (postCommit, nonRevertible int) {
		__antithesis_instrumentation__.Notify(595272)
		var pl scplan.Plan
		processPlanInPhase(t, newCluster, setup, stmts, scop.PostCommitPhase,
			func(p scplan.Plan) {
				__antithesis_instrumentation__.Notify(595274)
				pl = p
				postCommit = len(p.StagesForCurrentPhase())
				nonRevertible = len(p.Stages) - postCommit
			}, func(db *gosql.DB) {
				__antithesis_instrumentation__.Notify(595275)
				tdb := sqlutils.MakeSQLRunner(db)
				var ok bool
				dbName, ok = maybeGetDatabaseForIDs(t, tdb, screl.AllTargetDescIDs(pl.TargetState))
				if ok {
					__antithesis_instrumentation__.Notify(595277)
					tdb.Exec(t, fmt.Sprintf("USE %q", dbName))
				} else {
					__antithesis_instrumentation__.Notify(595278)
				}
				__antithesis_instrumentation__.Notify(595276)
				after = tdb.QueryStr(t, fetchDescriptorStateQuery)
			})
		__antithesis_instrumentation__.Notify(595273)
		return postCommit, nonRevertible
	}
	__antithesis_instrumentation__.Notify(595268)
	var testBackupRestoreCase func(
		t *testing.T, setup, stmts []parser.Statement, ord int,
	)
	testFunc := func(t *testing.T, setup, stmts []parser.Statement) {
		__antithesis_instrumentation__.Notify(595279)
		postCommit, nonRevertible := countStages(t, setup, stmts)
		if nonRevertible > 0 {
			__antithesis_instrumentation__.Notify(595281)
			postCommit++
		} else {
			__antithesis_instrumentation__.Notify(595282)
		}
		__antithesis_instrumentation__.Notify(595280)
		n := postCommit
		t.Logf("test case has %d revertible post-commit stages", n)
		for i := 1; i <= n; i++ {
			__antithesis_instrumentation__.Notify(595283)
			if !t.Run(
				fmt.Sprintf("backup/restore stage %d of %d", i, n),
				func(t *testing.T) {
					__antithesis_instrumentation__.Notify(595284)
					testBackupRestoreCase(t, setup, stmts, i)
				},
			) {
				__antithesis_instrumentation__.Notify(595285)
				return
			} else {
				__antithesis_instrumentation__.Notify(595286)
			}
		}
	}
	__antithesis_instrumentation__.Notify(595269)
	type stage struct {
		p        scplan.Plan
		stageIdx int
		resume   chan error
	}
	mkStage := func(p scplan.Plan, stageIdx int) stage {
		__antithesis_instrumentation__.Notify(595287)
		return stage{p: p, stageIdx: stageIdx, resume: make(chan error)}
	}
	__antithesis_instrumentation__.Notify(595270)
	testBackupRestoreCase = func(
		t *testing.T, setup, stmts []parser.Statement, ord int,
	) {
		__antithesis_instrumentation__.Notify(595288)
		stageChan := make(chan stage)
		ctx, cancel := context.WithCancel(context.Background())
		db, cleanup := newCluster(t, &scrun.TestingKnobs{
			BeforeStage: func(p scplan.Plan, stageIdx int) error {
				__antithesis_instrumentation__.Notify(595294)
				if p.Stages[stageIdx].Phase < scop.PostCommitPhase {
					__antithesis_instrumentation__.Notify(595297)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(595298)
				}
				__antithesis_instrumentation__.Notify(595295)
				if stageChan != nil {
					__antithesis_instrumentation__.Notify(595299)
					s := mkStage(p, stageIdx)
					select {
					case stageChan <- s:
						__antithesis_instrumentation__.Notify(595301)
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(595302)
						return ctx.Err()
					}
					__antithesis_instrumentation__.Notify(595300)
					select {
					case err := <-s.resume:
						__antithesis_instrumentation__.Notify(595303)
						return err
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(595304)
						return ctx.Err()
					}
				} else {
					__antithesis_instrumentation__.Notify(595305)
				}
				__antithesis_instrumentation__.Notify(595296)
				return nil
			},
		})
		__antithesis_instrumentation__.Notify(595289)

		defer cleanup()
		defer cancel()

		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		tdb := sqlutils.MakeSQLRunner(conn)
		tdb.Exec(t, "create database backups")
		var g errgroup.Group
		var before [][]string
		beforeFunc := func() {
			__antithesis_instrumentation__.Notify(595306)
			tdb.Exec(t, fmt.Sprintf("USE %q", dbName))
			before = tdb.QueryStr(t, fetchDescriptorStateQuery)
		}
		__antithesis_instrumentation__.Notify(595290)
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(595307)
			return executeSchemaChangeTxn(
				context.Background(), t, setup, stmts, db, beforeFunc, nil, nil,
			)
		})
		__antithesis_instrumentation__.Notify(595291)
		type backup struct {
			name       string
			isRollback bool
			url        string
			s          stage
		}
		var backups []backup
		var done bool
		var rollbackStage int
		for i := 0; !done; i++ {
			__antithesis_instrumentation__.Notify(595308)

			s := <-stageChan
			shouldFail := ord == i && func() bool {
				__antithesis_instrumentation__.Notify(595311)
				return s.p.Stages[s.stageIdx].Phase != scop.PostCommitNonRevertiblePhase == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(595312)
				return !s.p.InRollback == true
			}() == true
			done = len(s.p.Stages) == s.stageIdx+1 && func() bool {
				__antithesis_instrumentation__.Notify(595313)
				return !shouldFail == true
			}() == true
			t.Logf("stage %d/%d in %v (rollback=%v) %d %q %v",
				s.stageIdx+1, len(s.p.Stages), s.p.Stages[s.stageIdx].Phase, s.p.InRollback, ord, dbName, done)

			var exists bool
			tdb.QueryRow(t,
				`SELECT count(*) > 0 FROM system.namespace WHERE "parentID" = 0 AND name = $1`,
				dbName).Scan(&exists)
			if !exists || func() bool {
				__antithesis_instrumentation__.Notify(595314)
				return (i < ord && func() bool {
					__antithesis_instrumentation__.Notify(595315)
					return !done == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(595316)
				close(s.resume)
				continue
			} else {
				__antithesis_instrumentation__.Notify(595317)
			}
			__antithesis_instrumentation__.Notify(595309)

			backupURL := fmt.Sprintf("userfile://backups.public.userfiles_$user/data%d", i)
			tdb.Exec(t, fmt.Sprintf(
				"BACKUP DATABASE %s INTO '%s'", dbName, backupURL))
			backups = append(backups, backup{
				name:       dbName,
				isRollback: rollbackStage > 0,
				url:        backupURL,
				s:          s,
			})

			if s.p.InRollback {
				__antithesis_instrumentation__.Notify(595318)
				rollbackStage++
			} else {
				__antithesis_instrumentation__.Notify(595319)
			}
			__antithesis_instrumentation__.Notify(595310)
			if shouldFail {
				__antithesis_instrumentation__.Notify(595320)
				s.resume <- errors.Newf("boom %d", i)
			} else {
				__antithesis_instrumentation__.Notify(595321)
				close(s.resume)
			}
		}
		__antithesis_instrumentation__.Notify(595292)
		if err := g.Wait(); rollbackStage > 0 {
			__antithesis_instrumentation__.Notify(595322)
			require.Regexp(t, fmt.Sprintf("boom %d", ord), err)
		} else {
			__antithesis_instrumentation__.Notify(595323)
			require.NoError(t, err)
		}
		__antithesis_instrumentation__.Notify(595293)
		stageChan = nil
		t.Logf("finished")

		for i, b := range backups {
			__antithesis_instrumentation__.Notify(595324)
			t.Run("", func(t *testing.T) {
				__antithesis_instrumentation__.Notify(595325)
				t.Logf("testing backup %d %v", i, b.isRollback)
				tdb.Exec(t, fmt.Sprintf("DROP DATABASE IF EXISTS %q CASCADE", dbName))
				tdb.Exec(t, "SET use_declarative_schema_changer = 'off'")
				tdb.Exec(t, fmt.Sprintf("RESTORE DATABASE %s FROM LATEST IN '%s'", dbName, b.url))
				tdb.Exec(t, fmt.Sprintf("USE %q", dbName))
				waitForSchemaChangesToFinish(t, tdb)
				afterRestore := tdb.QueryStr(t, fetchDescriptorStateQuery)
				if b.isRollback {
					__antithesis_instrumentation__.Notify(595327)
					require.Equal(t, before, afterRestore)
				} else {
					__antithesis_instrumentation__.Notify(595328)
					require.Equal(t, after, afterRestore)
				}
				__antithesis_instrumentation__.Notify(595326)

				const validateQuery = `
SELECT * FROM crdb_internal.invalid_objects WHERE database_name != 'backups'
`
				tdb.CheckQueryResults(t, validateQuery, [][]string{})
				tdb.Exec(t, fmt.Sprintf("DROP DATABASE %q CASCADE", dbName))
				tdb.Exec(t, "USE backups")
				tdb.CheckQueryResults(t, validateQuery, [][]string{})
			})
		}
	}
	__antithesis_instrumentation__.Notify(595271)
	cumulativeTest(t, dir, testFunc)
}

func maybeGetDatabaseForIDs(
	t *testing.T, tdb *sqlutils.SQLRunner, ids catalog.DescriptorIDSet,
) (dbName string, exists bool) {
	__antithesis_instrumentation__.Notify(595329)
	err := tdb.DB.QueryRowContext(context.Background(), `
SELECT name
  FROM system.namespace
 WHERE id
       IN (
            SELECT DISTINCT
                   COALESCE(
                    d->'database'->>'id',
                    d->'schema'->>'parentId',
                    d->'type'->>'parentId',
                    d->'table'->>'parentId'
                   )::INT8
              FROM (
                    SELECT crdb_internal.pb_to_json('desc', descriptor) AS d
                      FROM system.descriptor
                     WHERE id IN (SELECT * FROM ROWS FROM (unnest($1::INT8[])))
                   )
        )
`, pq.Array(ids.Ordered())).
		Scan(&dbName)
	if errors.Is(err, gosql.ErrNoRows) {
		__antithesis_instrumentation__.Notify(595331)
		return "", false
	} else {
		__antithesis_instrumentation__.Notify(595332)
	}
	__antithesis_instrumentation__.Notify(595330)

	require.NoError(t, err)
	return dbName, true
}

func processPlanInPhase(
	t *testing.T,
	newCluster NewClusterFunc,
	setup, stmt []parser.Statement,
	phaseToProcess scop.Phase,
	processFunc func(p scplan.Plan),
	after func(db *gosql.DB),
) {
	__antithesis_instrumentation__.Notify(595333)
	var processOnce sync.Once
	db, cleanup := newCluster(t, &scrun.TestingKnobs{
		BeforeStage: func(p scplan.Plan, _ int) error {
			__antithesis_instrumentation__.Notify(595335)
			if p.Params.ExecutionPhase == phaseToProcess {
				__antithesis_instrumentation__.Notify(595337)
				processOnce.Do(func() { __antithesis_instrumentation__.Notify(595338); processFunc(p) })
			} else {
				__antithesis_instrumentation__.Notify(595339)
			}
			__antithesis_instrumentation__.Notify(595336)
			return nil
		},
	})
	__antithesis_instrumentation__.Notify(595334)
	defer cleanup()
	require.NoError(t, executeSchemaChangeTxn(
		context.Background(), t, setup, stmt, db, nil, nil, nil,
	))
	if after != nil {
		__antithesis_instrumentation__.Notify(595340)
		after(db)
	} else {
		__antithesis_instrumentation__.Notify(595341)
	}
}

func executeSchemaChangeTxn(
	ctx context.Context,
	t *testing.T,
	setup []parser.Statement,
	stmts []parser.Statement,
	db *gosql.DB,
	before func(),
	txnStartCallback func(),
	onError func(err error) error,
) (err error) {
	__antithesis_instrumentation__.Notify(595342)

	tdb := sqlutils.MakeSQLRunner(db)

	tdb.Exec(t, "SET use_declarative_schema_changer = 'off'")
	for _, stmt := range setup {
		__antithesis_instrumentation__.Notify(595347)
		tdb.Exec(t, stmt.SQL)
	}
	__antithesis_instrumentation__.Notify(595343)
	waitForSchemaChangesToSucceed(t, tdb)
	if before != nil {
		__antithesis_instrumentation__.Notify(595348)
		before()
	} else {
		__antithesis_instrumentation__.Notify(595349)
	}

	{
		__antithesis_instrumentation__.Notify(595350)
		c := make(chan error, 1)
		go func() {
			__antithesis_instrumentation__.Notify(595352)
			conn, err := db.Conn(ctx)
			if err != nil {
				__antithesis_instrumentation__.Notify(595355)
				c <- err
				return
			} else {
				__antithesis_instrumentation__.Notify(595356)
			}
			__antithesis_instrumentation__.Notify(595353)
			defer func() { __antithesis_instrumentation__.Notify(595357); _ = conn.Close() }()
			__antithesis_instrumentation__.Notify(595354)
			c <- crdb.Execute(func() (err error) {
				__antithesis_instrumentation__.Notify(595358)
				_, err = conn.ExecContext(
					ctx, "SET use_declarative_schema_changer = 'unsafe_always'",
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(595364)
					return err
				} else {
					__antithesis_instrumentation__.Notify(595365)
				}
				__antithesis_instrumentation__.Notify(595359)
				var tx *gosql.Tx
				tx, err = conn.BeginTx(ctx, nil)
				if err != nil {
					__antithesis_instrumentation__.Notify(595366)
					return err
				} else {
					__antithesis_instrumentation__.Notify(595367)
				}
				__antithesis_instrumentation__.Notify(595360)
				defer func() {
					__antithesis_instrumentation__.Notify(595368)
					if err != nil {
						__antithesis_instrumentation__.Notify(595369)
						err = errors.WithSecondaryError(err, tx.Rollback())
					} else {
						__antithesis_instrumentation__.Notify(595370)
						err = tx.Commit()
					}
				}()
				__antithesis_instrumentation__.Notify(595361)
				if txnStartCallback != nil {
					__antithesis_instrumentation__.Notify(595371)
					txnStartCallback()
				} else {
					__antithesis_instrumentation__.Notify(595372)
				}
				__antithesis_instrumentation__.Notify(595362)
				for _, stmt := range stmts {
					__antithesis_instrumentation__.Notify(595373)
					if _, err := tx.Exec(stmt.SQL); err != nil {
						__antithesis_instrumentation__.Notify(595374)
						return err
					} else {
						__antithesis_instrumentation__.Notify(595375)
					}
				}
				__antithesis_instrumentation__.Notify(595363)
				return nil
			})
		}()
		__antithesis_instrumentation__.Notify(595351)
		testutils.SucceedsSoon(t, func() error {
			__antithesis_instrumentation__.Notify(595376)
			select {
			case e := <-c:
				__antithesis_instrumentation__.Notify(595377)
				err = e
				return nil
			default:
				__antithesis_instrumentation__.Notify(595378)
				return errors.New("waiting for statements to execute")
			}
		})
	}
	__antithesis_instrumentation__.Notify(595344)

	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(595379)
		return onError != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(595380)
		err = onError(err)
	} else {
		__antithesis_instrumentation__.Notify(595381)
	}
	__antithesis_instrumentation__.Notify(595345)
	if err != nil {
		__antithesis_instrumentation__.Notify(595382)
		return err
	} else {
		__antithesis_instrumentation__.Notify(595383)
	}
	__antithesis_instrumentation__.Notify(595346)

	waitForSchemaChangesToSucceed(t, tdb)
	return nil
}
