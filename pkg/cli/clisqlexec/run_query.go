package clisqlexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"database/sql/driver"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (sqlExecCtx *Context) RunQuery(
	ctx context.Context, conn clisqlclient.Conn, fn clisqlclient.QueryFn, showMoreChars bool,
) ([]string, [][]string, error) {
	__antithesis_instrumentation__.Notify(29300)
	rows, _, err := fn(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(29303)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(29304)
	}
	__antithesis_instrumentation__.Notify(29301)

	defer func() { __antithesis_instrumentation__.Notify(29305); _ = rows.Close() }()
	__antithesis_instrumentation__.Notify(29302)
	return sqlRowsToStrings(rows, showMoreChars)
}

func (sqlExecCtx *Context) RunQueryAndFormatResults(
	ctx context.Context, conn clisqlclient.Conn, w, ew io.Writer, fn clisqlclient.QueryFn,
) (err error) {
	__antithesis_instrumentation__.Notify(29306)
	startTime := timeutil.Now()
	rows, isMultiStatementQuery, err := fn(ctx, conn)
	if err != nil {
		__antithesis_instrumentation__.Notify(29309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(29310)
	}
	__antithesis_instrumentation__.Notify(29307)
	defer func() {
		__antithesis_instrumentation__.Notify(29311)
		closeErr := rows.Close()
		err = errors.CombineErrors(err, closeErr)
	}()
	__antithesis_instrumentation__.Notify(29308)
	for {
		__antithesis_instrumentation__.Notify(29312)

		noRowsHook := func() (bool, error) {
			__antithesis_instrumentation__.Notify(29317)
			res := rows.Result()
			if ra, ok := res.(driver.RowsAffected); ok {
				__antithesis_instrumentation__.Notify(29319)
				nRows, err := ra.RowsAffected()
				if err != nil {
					__antithesis_instrumentation__.Notify(29322)
					return false, err
				} else {
					__antithesis_instrumentation__.Notify(29323)
				}
				__antithesis_instrumentation__.Notify(29320)

				tag := rows.Tag()
				if tag == "SELECT" && func() bool {
					__antithesis_instrumentation__.Notify(29324)
					return nRows == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(29325)

					return false, nil
				} else {
					__antithesis_instrumentation__.Notify(29326)
					if _, ok := tagsWithRowsAffected[tag]; ok {
						__antithesis_instrumentation__.Notify(29327)

						nRows, err := ra.RowsAffected()
						if err != nil {
							__antithesis_instrumentation__.Notify(29329)
							return false, err
						} else {
							__antithesis_instrumentation__.Notify(29330)
						}
						__antithesis_instrumentation__.Notify(29328)
						fmt.Fprintf(w, "%s %d\n", tag, nRows)
					} else {
						__antithesis_instrumentation__.Notify(29331)

						if tag == "" {
							__antithesis_instrumentation__.Notify(29333)
							tag = "OK"
						} else {
							__antithesis_instrumentation__.Notify(29334)
						}
						__antithesis_instrumentation__.Notify(29332)
						fmt.Fprintln(w, tag)
					}
				}
				__antithesis_instrumentation__.Notify(29321)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(29335)
			}
			__antithesis_instrumentation__.Notify(29318)

			return false, nil
		}
		__antithesis_instrumentation__.Notify(29313)

		cols := getColumnStrings(rows, true)
		reporter, cleanup, err := sqlExecCtx.makeReporter(w)
		if err != nil {
			__antithesis_instrumentation__.Notify(29336)
			return err
		} else {
			__antithesis_instrumentation__.Notify(29337)
		}
		__antithesis_instrumentation__.Notify(29314)

		var queryCompleteTime time.Time
		completedHook := func() { __antithesis_instrumentation__.Notify(29338); queryCompleteTime = timeutil.Now() }
		__antithesis_instrumentation__.Notify(29315)

		if err := func() error {
			__antithesis_instrumentation__.Notify(29339)
			if cleanup != nil {
				__antithesis_instrumentation__.Notify(29341)
				defer cleanup()
			} else {
				__antithesis_instrumentation__.Notify(29342)
			}
			__antithesis_instrumentation__.Notify(29340)
			return render(reporter, w, ew, cols, newRowIter(rows, true), completedHook, noRowsHook)
		}(); err != nil {
			__antithesis_instrumentation__.Notify(29343)
			return err
		} else {
			__antithesis_instrumentation__.Notify(29344)
		}
		__antithesis_instrumentation__.Notify(29316)

		sqlExecCtx.maybeShowTimes(ctx, conn, w, ew, isMultiStatementQuery, startTime, queryCompleteTime)

		if more, err := rows.NextResultSet(); err != nil {
			__antithesis_instrumentation__.Notify(29345)
			return err
		} else {
			__antithesis_instrumentation__.Notify(29346)
			if !more {
				__antithesis_instrumentation__.Notify(29347)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(29348)
			}
		}
	}
}

func (sqlExecCtx *Context) maybeShowTimes(
	ctx context.Context,
	conn clisqlclient.Conn,
	w, ew io.Writer,
	isMultiStatementQuery bool,
	startTime, queryCompleteTime time.Time,
) {
	__antithesis_instrumentation__.Notify(29349)
	if !sqlExecCtx.ShowTimes {
		__antithesis_instrumentation__.Notify(29359)
		return
	} else {
		__antithesis_instrumentation__.Notify(29360)
	}
	__antithesis_instrumentation__.Notify(29350)

	defer func() {
		__antithesis_instrumentation__.Notify(29361)

		renderDelay := timeutil.Since(queryCompleteTime)
		if renderDelay >= 1*time.Second && func() bool {
			__antithesis_instrumentation__.Notify(29363)
			return sqlExecCtx.IsInteractive() == true
		}() == true {
			__antithesis_instrumentation__.Notify(29364)
			fmt.Fprintf(ew,
				"\nNote: an additional delay of %s was spent formatting the results.\n"+
					"You can use \\set display_format to change the formatting.\n",
				renderDelay)
		} else {
			__antithesis_instrumentation__.Notify(29365)
		}
		__antithesis_instrumentation__.Notify(29362)

		fmt.Fprintln(w)
	}()
	__antithesis_instrumentation__.Notify(29351)

	clientSideQueryLatency := queryCompleteTime.Sub(startTime)

	if isMultiStatementQuery {
		__antithesis_instrumentation__.Notify(29366)

		if sqlExecCtx.IsInteractive() {
			__antithesis_instrumentation__.Notify(29368)
			fmt.Fprintf(ew, "\nNote: timings for multiple statements on a single line are not supported. See %s.\n",
				build.MakeIssueURL(48180))
		} else {
			__antithesis_instrumentation__.Notify(29369)
		}
		__antithesis_instrumentation__.Notify(29367)
		return
	} else {
		__antithesis_instrumentation__.Notify(29370)
	}
	__antithesis_instrumentation__.Notify(29352)

	fmt.Fprintln(w)

	var stats strings.Builder

	fmt.Fprintln(&stats)

	unit := "s"
	multiplier := 1.
	precision := 3
	if clientSideQueryLatency.Seconds() < 1 {
		__antithesis_instrumentation__.Notify(29371)
		unit = "ms"
		multiplier = 1000.
		precision = 0
	} else {
		__antithesis_instrumentation__.Notify(29372)
	}
	__antithesis_instrumentation__.Notify(29353)

	if sqlExecCtx.VerboseTimings {
		__antithesis_instrumentation__.Notify(29373)
		fmt.Fprintf(&stats, "Time: %s", clientSideQueryLatency)
	} else {
		__antithesis_instrumentation__.Notify(29374)

		fmt.Fprintf(&stats, "Time: %.*f%s", precision, clientSideQueryLatency.Seconds()*multiplier, unit)
	}
	__antithesis_instrumentation__.Notify(29354)

	detailedStats, err := conn.GetLastQueryStatistics(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(29375)
		fmt.Fprintln(w, stats.String())
		fmt.Fprintf(ew, "\nwarning: %v", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(29376)
	}
	__antithesis_instrumentation__.Notify(29355)
	if !detailedStats.Enabled {
		__antithesis_instrumentation__.Notify(29377)
		fmt.Fprintln(w, stats.String())
		return
	} else {
		__antithesis_instrumentation__.Notify(29378)
	}
	__antithesis_instrumentation__.Notify(29356)

	fmt.Fprint(&stats, " total")

	containsJobLat := detailedStats.PostCommitJobs.Valid
	parseLat := detailedStats.Parse.Value
	serviceLat := detailedStats.Service.Value
	planLat := detailedStats.Plan.Value
	execLat := detailedStats.Exec.Value
	jobsLat := detailedStats.PostCommitJobs.Value

	networkLat := clientSideQueryLatency - (serviceLat + jobsLat)

	if networkLat.Seconds() < 0 {
		__antithesis_instrumentation__.Notify(29379)
		networkLat = clientSideQueryLatency
	} else {
		__antithesis_instrumentation__.Notify(29380)
	}
	__antithesis_instrumentation__.Notify(29357)
	otherLat := serviceLat - parseLat - planLat - execLat
	if sqlExecCtx.VerboseTimings {
		__antithesis_instrumentation__.Notify(29381)

		if containsJobLat {
			__antithesis_instrumentation__.Notify(29382)
			fmt.Fprintf(&stats, " (parse %s / plan %s / exec %s / schema change %s / other %s / network %s)",
				parseLat, planLat, execLat, jobsLat, otherLat, networkLat)
		} else {
			__antithesis_instrumentation__.Notify(29383)
			fmt.Fprintf(&stats, " (parse %s / plan %s / exec %s / other %s / network %s)",
				parseLat, planLat, execLat, otherLat, networkLat)
		}
	} else {
		__antithesis_instrumentation__.Notify(29384)

		sep := " ("
		reportTiming := func(label string, lat time.Duration) {
			__antithesis_instrumentation__.Notify(29386)
			fmt.Fprintf(&stats, "%s%s %.*f%s", sep, label, precision, lat.Seconds()*multiplier, unit)
			sep = " / "
		}
		__antithesis_instrumentation__.Notify(29385)
		reportTiming("execution", serviceLat+jobsLat)
		reportTiming("network", networkLat)
		fmt.Fprint(&stats, ")")
	}
	__antithesis_instrumentation__.Notify(29358)
	fmt.Fprintln(w, stats.String())
}

var tagsWithRowsAffected = map[string]struct{}{
	"INSERT":    {},
	"UPDATE":    {},
	"DELETE":    {},
	"MOVE":      {},
	"DROP USER": {},
	"COPY":      {},

	"SELECT": {},
}

func sqlRowsToStrings(rows clisqlclient.Rows, showMoreChars bool) ([]string, [][]string, error) {
	__antithesis_instrumentation__.Notify(29387)
	cols := getColumnStrings(rows, showMoreChars)
	allRows, err := getAllRowStrings(rows, rows.ColumnTypeNames(), showMoreChars)
	if err != nil {
		__antithesis_instrumentation__.Notify(29389)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(29390)
	}
	__antithesis_instrumentation__.Notify(29388)
	return cols, allRows, nil
}

func getColumnStrings(rows clisqlclient.Rows, showMoreChars bool) []string {
	__antithesis_instrumentation__.Notify(29391)
	srcCols := rows.Columns()
	cols := make([]string, len(srcCols))
	for i, c := range srcCols {
		__antithesis_instrumentation__.Notify(29393)
		cols[i] = FormatVal(c, "NAME", showMoreChars, showMoreChars)
	}
	__antithesis_instrumentation__.Notify(29392)
	return cols
}
