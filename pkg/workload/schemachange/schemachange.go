package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"regexp"
	"runtime"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/pflag"
)

const (
	defaultMaxOpsPerWorker    = 5
	defaultErrorRate          = 10
	defaultEnumPct            = 10
	defaultMaxSourceTables    = 3
	defaultSequenceOwnedByPct = 25
	defaultFkParentInvalidPct = 5
	defaultFkChildInvalidPct  = 5
)

type schemaChange struct {
	flags              workload.Flags
	dbOverride         string
	concurrency        int
	maxOpsPerWorker    int
	errorRate          int
	enumPct            int
	verbose            int
	dryRun             bool
	maxSourceTables    int
	sequenceOwnedByPct int
	logFilePath        string
	logFile            *os.File
	dumpLogsOnce       *sync.Once
	workers            []*schemaChangeWorker
	fkParentInvalidPct int
	fkChildInvalidPct  int
}

var schemaChangeMeta = workload.Meta{
	Name:        `schemachange`,
	Description: `schemachange randomly generates concurrent schema changes`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(697155)
		s := &schemaChange{}
		s.flags.FlagSet = pflag.NewFlagSet(`schemachange`, pflag.ContinueOnError)
		s.flags.StringVar(&s.dbOverride, `db`, ``,
			`Override for the SQL database to use. If empty, defaults to the generator name`)
		s.flags.IntVar(&s.concurrency, `concurrency`, 2*runtime.GOMAXPROCS(0),
			`Number of concurrent workers`)
		s.flags.IntVar(&s.maxOpsPerWorker, `max-ops-per-worker`, defaultMaxOpsPerWorker,
			`Number of operations to execute in a single transaction`)
		s.flags.IntVar(&s.errorRate, `error-rate`, defaultErrorRate,
			`Percentage of times to intentionally cause errors due to either existing or non-existing names`)
		s.flags.IntVar(&s.enumPct, `enum-pct`, defaultEnumPct,
			`Percentage of times when picking a type that an enum type is picked`)
		s.flags.IntVarP(&s.verbose, `verbose`, `v`, 0, ``)
		s.flags.BoolVarP(&s.dryRun, `dry-run`, `n`, false, ``)
		s.flags.IntVar(&s.maxSourceTables, `max-source-tables`, defaultMaxSourceTables,
			`Maximum tables or views that a newly created tables or views can depend on`)
		s.flags.IntVar(&s.sequenceOwnedByPct, `seq-owned-pct`, defaultSequenceOwnedByPct,
			`Percentage of times that a sequence is owned by column upon creation.`)
		s.flags.StringVar(&s.logFilePath, `txn-log`, "",
			`If provided, transactions will be written to this file in JSON form`)
		s.flags.IntVar(&s.fkParentInvalidPct, `fk-parent-invalid-pct`, defaultFkParentInvalidPct,
			`Percentage of times to choose an invalid parent column in a fk constraint.`)
		s.flags.IntVar(&s.fkChildInvalidPct, `fk-child-invalid-pct`, defaultFkChildInvalidPct,
			`Percentage of times to choose an invalid child column in a fk constraint.`)
		return s
	},
}

func init() {
	workload.Register(schemaChangeMeta)
}

func (s *schemaChange) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(697156)
	return schemaChangeMeta
}

func (s *schemaChange) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(697157)
	return s.flags
}

func (s *schemaChange) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(697158)
	return nil
}

func (s *schemaChange) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(697159)
	return workload.Hooks{
		PostRun: func(_ time.Duration) error {
			__antithesis_instrumentation__.Notify(697160)
			return s.closeJSONLogFile()
		},
	}
}

func (s *schemaChange) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(697161)
	sqlDatabase, err := workload.SanitizeUrls(s, s.dbOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(697167)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(697168)
	}
	__antithesis_instrumentation__.Notify(697162)
	cfg := workload.MultiConnPoolCfg{
		MaxTotalConnections: s.concurrency * 2,
	}
	pool, err := workload.NewMultiConnPool(ctx, cfg, urls...)
	if err != nil {
		__antithesis_instrumentation__.Notify(697169)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(697170)
	}
	__antithesis_instrumentation__.Notify(697163)

	seqNum, err := s.initSeqNum(ctx, pool)
	if err != nil {
		__antithesis_instrumentation__.Notify(697171)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(697172)
	}
	__antithesis_instrumentation__.Notify(697164)

	ops := newDeck(rand.New(rand.NewSource(timeutil.Now().UnixNano())), opWeights...)
	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}

	stdoutLog := makeAtomicLog(os.Stdout)
	var artifactsLog *atomicLog
	if s.logFilePath != "" {
		__antithesis_instrumentation__.Notify(697173)
		err := s.initJSONLogFile(s.logFilePath)
		if err != nil {
			__antithesis_instrumentation__.Notify(697175)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(697176)
		}
		__antithesis_instrumentation__.Notify(697174)
		artifactsLog = makeAtomicLog(s.logFile)
	} else {
		__antithesis_instrumentation__.Notify(697177)
	}
	__antithesis_instrumentation__.Notify(697165)

	s.dumpLogsOnce = &sync.Once{}

	for i := 0; i < s.concurrency; i++ {
		__antithesis_instrumentation__.Notify(697178)

		opGeneratorParams := operationGeneratorParams{
			seqNum:             seqNum,
			errorRate:          s.errorRate,
			enumPct:            s.enumPct,
			rng:                rand.New(rand.NewSource(timeutil.Now().UnixNano())),
			ops:                ops,
			maxSourceTables:    s.maxSourceTables,
			sequenceOwnedByPct: s.sequenceOwnedByPct,
			fkParentInvalidPct: s.fkParentInvalidPct,
			fkChildInvalidPct:  s.fkChildInvalidPct,
		}

		w := &schemaChangeWorker{
			id:              i,
			workload:        s,
			dryRun:          s.dryRun,
			maxOpsPerWorker: s.maxOpsPerWorker,
			pool:            pool,
			hists:           reg.GetHandle(),
			opGen:           makeOperationGenerator(&opGeneratorParams),
			logger: &logger{
				verbose: s.verbose,
				currentLogEntry: &struct {
					mu struct {
						syncutil.Mutex
						entry *LogEntry
					}
				}{},
				stdoutLog:    stdoutLog,
				artifactsLog: artifactsLog,
			},
			isHoldingEntryLocks: false,
		}

		s.workers = append(s.workers, w)

		ql.WorkerFns = append(ql.WorkerFns, w.run)
		ql.Close = func(ctx2 context.Context) {
			__antithesis_instrumentation__.Notify(697179)
			pool.Close()
		}
	}
	__antithesis_instrumentation__.Notify(697166)
	return ql, nil
}

func (s *schemaChange) initSeqNum(
	ctx context.Context, pool *workload.MultiConnPool,
) (*int64, error) {
	__antithesis_instrumentation__.Notify(697180)
	seqNum := new(int64)

	const q = `
SELECT max(regexp_extract(name, '[0-9]+$')::INT8)
  FROM (
    SELECT name
      FROM (
	           (SELECT table_name FROM [SHOW TABLES]) UNION
						 (SELECT sequence_name FROM [SHOW SEQUENCES]) UNION
						 (SELECT name FROM [SHOW ENUMS]) UNION
	           (SELECT schema_name FROM [SHOW SCHEMAS]) UNION
						 (SELECT column_name FROM information_schema.columns) UNION
						 (SELECT index_name FROM information_schema.statistics)
           ) AS obj (name)
       )
 WHERE name ~ '^(table|view|seq|enum|schema)[0-9]+$'
    OR name ~ '^(col|index)[0-9]+_[0-9]+$';
`
	var max gosql.NullInt64
	if err := pool.Get().QueryRow(ctx, q).Scan(&max); err != nil {
		__antithesis_instrumentation__.Notify(697183)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697184)
	}
	__antithesis_instrumentation__.Notify(697181)
	if max.Valid {
		__antithesis_instrumentation__.Notify(697185)
		*seqNum = max.Int64 + 1
	} else {
		__antithesis_instrumentation__.Notify(697186)
	}
	__antithesis_instrumentation__.Notify(697182)

	return seqNum, nil
}

type schemaChangeWorker struct {
	id                  int
	workload            *schemaChange
	dryRun              bool
	maxOpsPerWorker     int
	pool                *workload.MultiConnPool
	hists               *histogram.Histograms
	opGen               *operationGenerator
	isHoldingEntryLocks bool
	logger              *logger
}

var (
	errRunInTxnFatalSentinel = errors.New("fatal error when running txn")
	errRunInTxnRbkSentinel   = errors.New("txn needs to rollback")
)

type LogEntry struct {
	WorkerID             int      `json:"workerId"`
	ClientTimestamp      string   `json:"clientTimestamp"`
	Ops                  []string `json:"ops"`
	ExpectedExecErrors   string   `json:"expectedExecErrors"`
	ExpectedCommitErrors string   `json:"expectedCommitErrors"`

	Message string `json:"message"`
}

type histBin int

const (
	operationOk histBin = iota
	txnOk
	txnCommitError
	txnRollback
)

func (d histBin) String() string {
	__antithesis_instrumentation__.Notify(697187)
	return [...]string{"opOk", "txnOk", "txnCmtErr", "txnRbk"}[d]
}

func (w *schemaChangeWorker) recordInHist(elapsed time.Duration, bin histBin) {
	__antithesis_instrumentation__.Notify(697188)
	w.hists.Get(bin.String()).Record(elapsed)
}

func (w *schemaChangeWorker) getErrorState() string {
	__antithesis_instrumentation__.Notify(697189)
	return fmt.Sprintf("Dumping state before death:\n"+
		"Expected errors: %s"+
		"==========================="+
		"Executed queries for generating errors: %s"+
		"==========================="+
		"Previous statements %s",
		w.opGen.expectedExecErrors.String(),
		w.opGen.GetOpGenLog(),
		w.opGen.stmtsInTxt)
}

func (w *schemaChangeWorker) runInTxn(ctx context.Context, tx pgx.Tx) error {
	__antithesis_instrumentation__.Notify(697190)
	w.logger.startLog()
	w.logger.writeLog("BEGIN")
	opsNum := 1 + w.opGen.randIntn(w.maxOpsPerWorker)

	for i := 0; i < opsNum; i++ {
		__antithesis_instrumentation__.Notify(697192)

		if !w.opGen.expectedCommitErrors.empty() {
			__antithesis_instrumentation__.Notify(697195)
			break
		} else {
			__antithesis_instrumentation__.Notify(697196)
		}
		__antithesis_instrumentation__.Notify(697193)

		op, err := w.opGen.randOp(ctx, tx)

		if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && func() bool {
			__antithesis_instrumentation__.Notify(697197)
			return pgcode.MakeCode(pgErr.Code) == pgcode.SerializationFailure == true
		}() == true {
			__antithesis_instrumentation__.Notify(697198)
			return errors.Mark(err, errRunInTxnRbkSentinel)
		} else {
			__antithesis_instrumentation__.Notify(697199)
			if err != nil {
				__antithesis_instrumentation__.Notify(697200)
				return errors.Mark(
					errors.Wrapf(err, "***UNEXPECTED ERROR; Failed to generate a random operation\n OpGen log: \n%s",
						w.opGen.GetOpGenLog(),
					),
					errRunInTxnFatalSentinel,
				)
			} else {
				__antithesis_instrumentation__.Notify(697201)
			}
		}
		__antithesis_instrumentation__.Notify(697194)

		w.logger.addExpectedErrors(w.opGen.expectedExecErrors, w.opGen.expectedCommitErrors)
		w.logger.writeLog(op)
		if !w.dryRun {
			__antithesis_instrumentation__.Notify(697202)
			start := timeutil.Now()

			if _, err = tx.Exec(ctx, op); err != nil {
				__antithesis_instrumentation__.Notify(697205)

				pgErr := new(pgconn.PgError)
				if !errors.As(err, &pgErr) {
					__antithesis_instrumentation__.Notify(697209)
					return errors.Mark(
						errors.Wrapf(err, "***UNEXPECTED ERROR; Received a non pg error. %s",
							w.getErrorState()),
						errRunInTxnFatalSentinel,
					)
				} else {
					__antithesis_instrumentation__.Notify(697210)
				}
				__antithesis_instrumentation__.Notify(697206)

				if pgcode.MakeCode(pgErr.Code) == pgcode.SerializationFailure {
					__antithesis_instrumentation__.Notify(697211)
					w.recordInHist(timeutil.Since(start), txnRollback)
					return errors.Mark(
						err,
						errRunInTxnRbkSentinel,
					)
				} else {
					__antithesis_instrumentation__.Notify(697212)
				}
				__antithesis_instrumentation__.Notify(697207)

				if !w.opGen.expectedExecErrors.contains(pgcode.MakeCode(pgErr.Code)) {
					__antithesis_instrumentation__.Notify(697213)
					return errors.Mark(
						errors.Wrapf(err, "***UNEXPECTED ERROR; Received an unexpected execution error. %s",
							w.getErrorState()),
						errRunInTxnFatalSentinel,
					)
				} else {
					__antithesis_instrumentation__.Notify(697214)
				}
				__antithesis_instrumentation__.Notify(697208)

				w.recordInHist(timeutil.Since(start), txnRollback)
				return errors.Mark(
					errors.Wrapf(err, "ROLLBACK; Successfully got expected execution error. %s",
						w.getErrorState()),
					errRunInTxnRbkSentinel,
				)
			} else {
				__antithesis_instrumentation__.Notify(697215)
			}
			__antithesis_instrumentation__.Notify(697203)
			if !w.opGen.expectedExecErrors.empty() {
				__antithesis_instrumentation__.Notify(697216)
				return errors.Mark(
					errors.Newf("***FAIL; Failed to receive an execution error when errors were expected. %s",
						w.getErrorState()),
					errRunInTxnFatalSentinel,
				)
			} else {
				__antithesis_instrumentation__.Notify(697217)
			}
			__antithesis_instrumentation__.Notify(697204)

			w.recordInHist(timeutil.Since(start), operationOk)
		} else {
			__antithesis_instrumentation__.Notify(697218)
		}
	}
	__antithesis_instrumentation__.Notify(697191)
	return nil
}

func (w *schemaChangeWorker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(697219)
	tx, err := w.pool.Get().Begin(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(697224)
		return errors.Wrap(err, "cannot get a connection and begin a txn")
	} else {
		__antithesis_instrumentation__.Notify(697225)
	}
	__antithesis_instrumentation__.Notify(697220)

	defer w.releaseLocksIfHeld()

	start := timeutil.Now()
	w.opGen.resetTxnState()
	err = w.runInTxn(ctx, tx)

	if err != nil {
		__antithesis_instrumentation__.Notify(697226)

		if rbkErr := tx.Rollback(ctx); rbkErr != nil {
			__antithesis_instrumentation__.Notify(697228)
			err = errors.Mark(
				errors.Wrap(rbkErr, "***UNEXPECTED ERROR DURING ROLLBACK;"),
				errRunInTxnFatalSentinel,
			)
		} else {
			__antithesis_instrumentation__.Notify(697229)
		}
		__antithesis_instrumentation__.Notify(697227)

		w.logger.flushLog(tx, err.Error())
		switch {
		case errors.Is(err, errRunInTxnFatalSentinel):
			__antithesis_instrumentation__.Notify(697230)
			w.preErrorHook()
			return err
		case errors.Is(err, errRunInTxnRbkSentinel):
			__antithesis_instrumentation__.Notify(697231)

			return nil
		default:
			__antithesis_instrumentation__.Notify(697232)
			w.preErrorHook()
			return errors.Wrapf(err, "***UNEXPECTED ERROR")
		}
	} else {
		__antithesis_instrumentation__.Notify(697233)
	}
	__antithesis_instrumentation__.Notify(697221)

	w.logger.writeLog("COMMIT")
	if err = tx.Commit(ctx); err != nil {
		__antithesis_instrumentation__.Notify(697234)

		pgErr := new(pgconn.PgError)
		if !errors.As(err, &pgErr) {
			__antithesis_instrumentation__.Notify(697239)
			err = errors.Mark(
				errors.Wrap(err, "***UNEXPECTED COMMIT ERROR; Received a non pg error"),
				errRunInTxnFatalSentinel,
			)
			w.logger.flushLog(tx, err.Error())
			w.preErrorHook()
			return err
		} else {
			__antithesis_instrumentation__.Notify(697240)
		}
		__antithesis_instrumentation__.Notify(697235)

		if pgcode.MakeCode(pgErr.Code) == pgcode.SerializationFailure {
			__antithesis_instrumentation__.Notify(697241)
			w.recordInHist(timeutil.Since(start), txnCommitError)
			w.logger.flushLog(tx, fmt.Sprintf("TXN RETRY ERROR; %v", pgErr))
			return nil
		} else {
			__antithesis_instrumentation__.Notify(697242)
		}
		__antithesis_instrumentation__.Notify(697236)

		if pgErr.Code == pgcode.TransactionCommittedWithSchemaChangeFailure.String() {
			__antithesis_instrumentation__.Notify(697243)
			re := regexp.MustCompile(`\([A-Z0-9]{5}\)`)
			underLyingErrorCode := re.FindString(pgErr.Error())
			if underLyingErrorCode != "" {
				__antithesis_instrumentation__.Notify(697244)
				pgErr.Code = underLyingErrorCode[1 : len(underLyingErrorCode)-1]
			} else {
				__antithesis_instrumentation__.Notify(697245)
			}
		} else {
			__antithesis_instrumentation__.Notify(697246)
		}
		__antithesis_instrumentation__.Notify(697237)

		if !w.opGen.expectedCommitErrors.contains(pgcode.MakeCode(pgErr.Code)) {
			__antithesis_instrumentation__.Notify(697247)
			err = errors.Mark(
				errors.Wrap(err, "***UNEXPECTED COMMIT ERROR; Received an unexpected commit error"),
				errRunInTxnFatalSentinel,
			)
			w.logger.flushLog(tx, err.Error())
			w.preErrorHook()
			return err
		} else {
			__antithesis_instrumentation__.Notify(697248)
		}
		__antithesis_instrumentation__.Notify(697238)

		w.recordInHist(timeutil.Since(start), txnCommitError)
		w.logger.flushLog(tx, "COMMIT; Successfully got expected commit error")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(697249)
	}
	__antithesis_instrumentation__.Notify(697222)

	if !w.opGen.expectedCommitErrors.empty() {
		__antithesis_instrumentation__.Notify(697250)
		err := errors.New("***FAIL; Failed to receive a commit error when at least one commit error was expected")
		w.logger.flushLog(tx, err.Error())
		w.preErrorHook()
		return errors.Mark(err, errRunInTxnFatalSentinel)
	} else {
		__antithesis_instrumentation__.Notify(697251)
	}
	__antithesis_instrumentation__.Notify(697223)

	w.logger.flushLog(tx, "")
	w.recordInHist(timeutil.Since(start), txnOk)
	return nil
}

func (w *schemaChangeWorker) preErrorHook() {
	__antithesis_instrumentation__.Notify(697252)
	w.workload.dumpLogsOnce.Do(func() {
		__antithesis_instrumentation__.Notify(697253)
		for _, worker := range w.workload.workers {
			__antithesis_instrumentation__.Notify(697255)
			worker.logger.flushLogAndLock(nil, "Flushed by pre-error hook", false)
			worker.logger.artifactsLog = nil
		}
		__antithesis_instrumentation__.Notify(697254)
		_ = w.workload.closeJSONLogFile()
		w.isHoldingEntryLocks = true
	})
}

func (w *schemaChangeWorker) releaseLocksIfHeld() {
	__antithesis_instrumentation__.Notify(697256)
	if w.isHoldingEntryLocks && func() bool {
		__antithesis_instrumentation__.Notify(697258)
		return w.logger.verbose >= 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697259)
		for _, worker := range w.workload.workers {
			__antithesis_instrumentation__.Notify(697260)
			worker.logger.currentLogEntry.mu.Unlock()
		}
	} else {
		__antithesis_instrumentation__.Notify(697261)
	}
	__antithesis_instrumentation__.Notify(697257)
	w.isHoldingEntryLocks = false
}

func (l *logger) startLog() {
	__antithesis_instrumentation__.Notify(697262)
	if l.verbose < 1 {
		__antithesis_instrumentation__.Notify(697264)
		return
	} else {
		__antithesis_instrumentation__.Notify(697265)
	}
	__antithesis_instrumentation__.Notify(697263)
	l.currentLogEntry.mu.Lock()
	defer l.currentLogEntry.mu.Unlock()
	l.currentLogEntry.mu.entry = &LogEntry{
		ClientTimestamp: timeutil.Now().Format("15:04:05.999999"),
	}
}

func (l *logger) writeLog(op string) {
	__antithesis_instrumentation__.Notify(697266)
	if l.verbose < 1 {
		__antithesis_instrumentation__.Notify(697268)
		return
	} else {
		__antithesis_instrumentation__.Notify(697269)
	}
	__antithesis_instrumentation__.Notify(697267)
	l.currentLogEntry.mu.Lock()
	defer l.currentLogEntry.mu.Unlock()
	if l.currentLogEntry.mu.entry != nil {
		__antithesis_instrumentation__.Notify(697270)
		l.currentLogEntry.mu.entry.Ops = append(l.currentLogEntry.mu.entry.Ops, op)
	} else {
		__antithesis_instrumentation__.Notify(697271)
	}
}

func (l *logger) addExpectedErrors(execErrors errorCodeSet, commitErrors errorCodeSet) {
	__antithesis_instrumentation__.Notify(697272)
	if l.verbose < 1 {
		__antithesis_instrumentation__.Notify(697274)
		return
	} else {
		__antithesis_instrumentation__.Notify(697275)
	}
	__antithesis_instrumentation__.Notify(697273)
	l.currentLogEntry.mu.Lock()
	defer l.currentLogEntry.mu.Unlock()
	if l.currentLogEntry.mu.entry != nil {
		__antithesis_instrumentation__.Notify(697276)
		l.currentLogEntry.mu.entry.ExpectedExecErrors = execErrors.String()
		l.currentLogEntry.mu.entry.ExpectedCommitErrors = commitErrors.String()
	} else {
		__antithesis_instrumentation__.Notify(697277)
	}
}

func (l *logger) flushLog(tx pgx.Tx, message string) {
	__antithesis_instrumentation__.Notify(697278)
	if l.verbose < 1 {
		__antithesis_instrumentation__.Notify(697280)
		return
	} else {
		__antithesis_instrumentation__.Notify(697281)
	}
	__antithesis_instrumentation__.Notify(697279)
	l.flushLogAndLock(tx, message, true)
	l.currentLogEntry.mu.Unlock()
}

func (l *logger) flushLogAndLock(_ pgx.Tx, message string, stdout bool) {
	__antithesis_instrumentation__.Notify(697282)
	if l.verbose < 1 {
		__antithesis_instrumentation__.Notify(697289)
		return
	} else {
		__antithesis_instrumentation__.Notify(697290)
	}
	__antithesis_instrumentation__.Notify(697283)

	l.currentLogEntry.mu.Lock()

	if l.currentLogEntry.mu.entry == nil || func() bool {
		__antithesis_instrumentation__.Notify(697291)
		return len(l.currentLogEntry.mu.entry.Ops) < 2 == true
	}() == true {
		__antithesis_instrumentation__.Notify(697292)
		return
	} else {
		__antithesis_instrumentation__.Notify(697293)
	}
	__antithesis_instrumentation__.Notify(697284)

	if message != "" {
		__antithesis_instrumentation__.Notify(697294)
		l.currentLogEntry.mu.entry.Message = message
	} else {
		__antithesis_instrumentation__.Notify(697295)
	}
	__antithesis_instrumentation__.Notify(697285)
	jsonBytes, err := json.MarshalIndent(l.currentLogEntry.mu.entry, "", " ")
	if err != nil {
		__antithesis_instrumentation__.Notify(697296)
		return
	} else {
		__antithesis_instrumentation__.Notify(697297)
	}
	__antithesis_instrumentation__.Notify(697286)
	if stdout {
		__antithesis_instrumentation__.Notify(697298)
		l.stdoutLog.printLn(string(jsonBytes))
	} else {
		__antithesis_instrumentation__.Notify(697299)
	}
	__antithesis_instrumentation__.Notify(697287)
	if l.artifactsLog != nil {
		__antithesis_instrumentation__.Notify(697300)
		var jsonBuf bytes.Buffer
		err = json.Compact(&jsonBuf, jsonBytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(697302)
			return
		} else {
			__antithesis_instrumentation__.Notify(697303)
		}
		__antithesis_instrumentation__.Notify(697301)
		l.artifactsLog.printLn(jsonBuf.String())
	} else {
		__antithesis_instrumentation__.Notify(697304)
	}
	__antithesis_instrumentation__.Notify(697288)
	l.currentLogEntry.mu.entry = nil
}

type logger struct {
	verbose         int
	currentLogEntry *struct {
		mu struct {
			syncutil.Mutex
			entry *LogEntry
		}
	}
	stdoutLog    *atomicLog
	artifactsLog *atomicLog
}

type atomicLog struct {
	mu struct {
		syncutil.Mutex
		log io.Writer
	}
}

func makeAtomicLog(w io.Writer) *atomicLog {
	__antithesis_instrumentation__.Notify(697305)
	return &atomicLog{
		mu: struct {
			syncutil.Mutex
			log io.Writer
		}{log: w},
	}
}

func (l *atomicLog) printLn(message string) {
	__antithesis_instrumentation__.Notify(697306)
	l.mu.Lock()
	defer l.mu.Unlock()

	_, _ = l.mu.log.Write(append([]byte(message), '\n'))
}

func (s *schemaChange) initJSONLogFile(filePath string) error {
	__antithesis_instrumentation__.Notify(697307)
	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		__antithesis_instrumentation__.Notify(697309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697310)
	}
	__antithesis_instrumentation__.Notify(697308)
	s.logFile = f
	return nil
}

func (s *schemaChange) closeJSONLogFile() error {
	__antithesis_instrumentation__.Notify(697311)
	if s.logFile == nil {
		__antithesis_instrumentation__.Notify(697314)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(697315)
	}
	__antithesis_instrumentation__.Notify(697312)

	if err := s.logFile.Sync(); err != nil {
		__antithesis_instrumentation__.Notify(697316)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697317)
	}
	__antithesis_instrumentation__.Notify(697313)
	err := s.logFile.Close()
	s.logFile = nil
	return err
}
