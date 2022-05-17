package workloadsql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

type InsertsDataLoader struct {
	BatchSize   int
	Concurrency int
}

func (l InsertsDataLoader) InitialDataLoad(
	ctx context.Context, db *gosql.DB, gen workload.Generator,
) (int64, error) {
	__antithesis_instrumentation__.Notify(699084)
	if gen.Meta().Name == `tpch` {
		__antithesis_instrumentation__.Notify(699093)
		return 0, errors.New(
			`tpch currently doesn't work with the inserts data loader. try --data-loader=import`)
	} else {
		__antithesis_instrumentation__.Notify(699094)
	}
	__antithesis_instrumentation__.Notify(699085)

	if l.BatchSize <= 0 {
		__antithesis_instrumentation__.Notify(699095)
		l.BatchSize = 1000
	} else {
		__antithesis_instrumentation__.Notify(699096)
	}
	__antithesis_instrumentation__.Notify(699086)
	if l.Concurrency < 1 {
		__antithesis_instrumentation__.Notify(699097)
		l.Concurrency = 1
	} else {
		__antithesis_instrumentation__.Notify(699098)
	}
	__antithesis_instrumentation__.Notify(699087)

	tables := gen.Tables()
	var hooks workload.Hooks
	if h, ok := gen.(workload.Hookser); ok {
		__antithesis_instrumentation__.Notify(699099)
		hooks = h.Hooks()
	} else {
		__antithesis_instrumentation__.Notify(699100)
	}
	__antithesis_instrumentation__.Notify(699088)

	if hooks.PreCreate != nil {
		__antithesis_instrumentation__.Notify(699101)
		if err := hooks.PreCreate(db); err != nil {
			__antithesis_instrumentation__.Notify(699102)
			return 0, errors.Wrapf(err, "Could not precreate")
		} else {
			__antithesis_instrumentation__.Notify(699103)
		}
	} else {
		__antithesis_instrumentation__.Notify(699104)
	}
	__antithesis_instrumentation__.Notify(699089)

	for _, table := range tables {
		__antithesis_instrumentation__.Notify(699105)
		createStmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS "%s" %s`, table.Name, table.Schema)
		if _, err := db.ExecContext(ctx, createStmt); err != nil {
			__antithesis_instrumentation__.Notify(699106)
			return 0, errors.Wrapf(err, "could not create table: %s", table.Name)
		} else {
			__antithesis_instrumentation__.Notify(699107)
		}
	}
	__antithesis_instrumentation__.Notify(699090)

	if hooks.PreLoad != nil {
		__antithesis_instrumentation__.Notify(699108)
		if err := hooks.PreLoad(db); err != nil {
			__antithesis_instrumentation__.Notify(699109)
			return 0, errors.Wrapf(err, "Could not preload")
		} else {
			__antithesis_instrumentation__.Notify(699110)
		}
	} else {
		__antithesis_instrumentation__.Notify(699111)
	}
	__antithesis_instrumentation__.Notify(699091)

	var bytesAtomic int64
	for _, table := range tables {
		__antithesis_instrumentation__.Notify(699112)
		if table.InitialRows.NumBatches == 0 {
			__antithesis_instrumentation__.Notify(699116)
			continue
		} else {
			__antithesis_instrumentation__.Notify(699117)
			if table.InitialRows.FillBatch == nil {
				__antithesis_instrumentation__.Notify(699118)
				return 0, errors.Errorf(
					`initial data is not supported for workload %s`, gen.Meta().Name)
			} else {
				__antithesis_instrumentation__.Notify(699119)
			}
		}
		__antithesis_instrumentation__.Notify(699113)
		tableStart := timeutil.Now()
		var tableRowsAtomic int64

		batchesPerWorker := table.InitialRows.NumBatches / l.Concurrency
		g, gCtx := errgroup.WithContext(ctx)
		for i := 0; i < l.Concurrency; i++ {
			__antithesis_instrumentation__.Notify(699120)
			startIdx := i * batchesPerWorker
			endIdx := startIdx + batchesPerWorker
			if i == l.Concurrency-1 {
				__antithesis_instrumentation__.Notify(699122)

				endIdx = table.InitialRows.NumBatches
			} else {
				__antithesis_instrumentation__.Notify(699123)
			}
			__antithesis_instrumentation__.Notify(699121)
			g.Go(func() error {
				__antithesis_instrumentation__.Notify(699124)
				var insertStmtBuf bytes.Buffer
				var params []interface{}
				var numRows int
				flush := func() error {
					__antithesis_instrumentation__.Notify(699127)
					if len(params) > 0 {
						__antithesis_instrumentation__.Notify(699129)
						insertStmt := insertStmtBuf.String()
						if _, err := db.ExecContext(gCtx, insertStmt, params...); err != nil {
							__antithesis_instrumentation__.Notify(699130)
							return errors.Wrapf(err, "failed insert into %s", table.Name)
						} else {
							__antithesis_instrumentation__.Notify(699131)
						}
					} else {
						__antithesis_instrumentation__.Notify(699132)
					}
					__antithesis_instrumentation__.Notify(699128)
					insertStmtBuf.Reset()
					fmt.Fprintf(&insertStmtBuf, `INSERT INTO "%s" VALUES `, table.Name)
					params = params[:0]
					numRows = 0
					return nil
				}
				__antithesis_instrumentation__.Notify(699125)
				_ = flush()

				for batchIdx := startIdx; batchIdx < endIdx; batchIdx++ {
					__antithesis_instrumentation__.Notify(699133)
					for _, row := range table.InitialRows.BatchRows(batchIdx) {
						__antithesis_instrumentation__.Notify(699134)
						atomic.AddInt64(&tableRowsAtomic, 1)
						if len(params) != 0 {
							__antithesis_instrumentation__.Notify(699137)
							insertStmtBuf.WriteString(`,`)
						} else {
							__antithesis_instrumentation__.Notify(699138)
						}
						__antithesis_instrumentation__.Notify(699135)
						insertStmtBuf.WriteString(`(`)
						for i, datum := range row {
							__antithesis_instrumentation__.Notify(699139)
							atomic.AddInt64(&bytesAtomic, workload.ApproxDatumSize(datum))
							if i != 0 {
								__antithesis_instrumentation__.Notify(699141)
								insertStmtBuf.WriteString(`,`)
							} else {
								__antithesis_instrumentation__.Notify(699142)
							}
							__antithesis_instrumentation__.Notify(699140)
							fmt.Fprintf(&insertStmtBuf, `$%d`, len(params)+i+1)
						}
						__antithesis_instrumentation__.Notify(699136)
						params = append(params, row...)
						insertStmtBuf.WriteString(`)`)
						if numRows++; numRows >= l.BatchSize {
							__antithesis_instrumentation__.Notify(699143)
							if err := flush(); err != nil {
								__antithesis_instrumentation__.Notify(699144)
								return err
							} else {
								__antithesis_instrumentation__.Notify(699145)
							}
						} else {
							__antithesis_instrumentation__.Notify(699146)
						}
					}
				}
				__antithesis_instrumentation__.Notify(699126)
				return flush()
			})
		}
		__antithesis_instrumentation__.Notify(699114)
		if err := g.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(699147)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(699148)
		}
		__antithesis_instrumentation__.Notify(699115)
		tableRows := int(atomic.LoadInt64(&tableRowsAtomic))
		log.Infof(ctx, `imported %s (%s, %d rows)`,
			table.Name, timeutil.Since(tableStart).Round(time.Second), tableRows,
		)
	}
	__antithesis_instrumentation__.Notify(699092)
	return atomic.LoadInt64(&bytesAtomic), nil
}
