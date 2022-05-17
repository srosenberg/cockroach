package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

var _ sqlutil.InternalExecutor = &InternalExecutor{}

type InternalExecutor struct {
	s *Server

	mon *mon.BytesMonitor

	memMetrics MemoryMetrics

	sessionDataStack *sessiondata.Stack

	syntheticDescriptors []catalog.Descriptor
}

func (ie *InternalExecutor) WithSyntheticDescriptors(
	descs []catalog.Descriptor, run func() error,
) error {
	__antithesis_instrumentation__.Notify(497746)
	ie.syntheticDescriptors = descs
	defer func() {
		__antithesis_instrumentation__.Notify(497748)
		ie.syntheticDescriptors = nil
	}()
	__antithesis_instrumentation__.Notify(497747)
	return run()
}

func MakeInternalExecutor(
	ctx context.Context, s *Server, memMetrics MemoryMetrics, settings *cluster.Settings,
) InternalExecutor {
	__antithesis_instrumentation__.Notify(497749)
	monitor := mon.NewMonitor(
		"internal SQL executor",
		mon.MemoryResource,
		memMetrics.CurBytesCount,
		memMetrics.MaxBytesHist,
		-1,
		math.MaxInt64,
		settings,
	)
	monitor.Start(ctx, s.pool, mon.BoundAccount{})
	return InternalExecutor{
		s:          s,
		mon:        monitor,
		memMetrics: memMetrics,
	}
}

func (ie *InternalExecutor) SetSessionData(sessionData *sessiondata.SessionData) {
	__antithesis_instrumentation__.Notify(497750)
	ie.s.populateMinimalSessionData(sessionData)
	ie.sessionDataStack = sessiondata.NewStack(sessionData)
}

func (ie *InternalExecutor) initConnEx(
	ctx context.Context,
	txn *kv.Txn,
	w ieResultWriter,
	sd *sessiondata.SessionData,
	stmtBuf *StmtBuf,
	wg *sync.WaitGroup,
	syncCallback func([]resWithPos),
	errCallback func(error),
) {
	__antithesis_instrumentation__.Notify(497751)
	clientComm := &internalClientComm{
		w: w,

		lastDelivered: -1,
		sync:          syncCallback,
	}

	var appStatsBucketName string
	if !strings.HasPrefix(sd.ApplicationName, catconstants.InternalAppNamePrefix) {
		__antithesis_instrumentation__.Notify(497754)
		appStatsBucketName = catconstants.DelegatedAppNamePrefix + sd.ApplicationName
	} else {
		__antithesis_instrumentation__.Notify(497755)

		appStatsBucketName = sd.ApplicationName
	}
	__antithesis_instrumentation__.Notify(497752)
	applicationStats := ie.s.sqlStats.GetApplicationStats(appStatsBucketName)

	sds := sessiondata.NewStack(sd)
	sdMutIterator := ie.s.makeSessionDataMutatorIterator(sds, nil)
	var ex *connExecutor
	if txn == nil {
		__antithesis_instrumentation__.Notify(497756)
		ex = ie.s.newConnExecutor(
			ctx,
			sdMutIterator,
			stmtBuf,
			clientComm,
			ie.memMetrics,
			&ie.s.InternalMetrics,
			applicationStats,
		)
	} else {
		__antithesis_instrumentation__.Notify(497757)
		ex = ie.s.newConnExecutorWithTxn(
			ctx,
			sdMutIterator,
			stmtBuf,
			clientComm,
			ie.mon,
			ie.memMetrics,
			&ie.s.InternalMetrics,
			txn,
			ie.syntheticDescriptors,
			applicationStats,
		)
	}
	__antithesis_instrumentation__.Notify(497753)

	ex.executorType = executorTypeInternal

	wg.Add(1)
	go func() {
		__antithesis_instrumentation__.Notify(497758)
		if err := ex.run(ctx, ie.mon, mon.BoundAccount{}, nil); err != nil {
			__antithesis_instrumentation__.Notify(497761)
			sqltelemetry.RecordError(ctx, err, &ex.server.cfg.Settings.SV)
			errCallback(err)
		} else {
			__antithesis_instrumentation__.Notify(497762)
		}
		__antithesis_instrumentation__.Notify(497759)
		w.finish()
		closeMode := normalClose
		if txn != nil {
			__antithesis_instrumentation__.Notify(497763)
			closeMode = externalTxnClose
		} else {
			__antithesis_instrumentation__.Notify(497764)
		}
		__antithesis_instrumentation__.Notify(497760)
		ex.close(ctx, closeMode)
		wg.Done()
	}()
}

type ieIteratorResult struct {
	row                   tree.Datums
	rowsAffectedIncrement *int
	cols                  colinfo.ResultColumns
	err                   error
}

type rowsIterator struct {
	r ieResultReader

	rowsAffected int
	resultCols   colinfo.ResultColumns

	first *ieIteratorResult

	lastRow tree.Datums
	lastErr error
	done    bool

	errCallback func(err error) error

	stmtBuf *StmtBuf

	wg *sync.WaitGroup

	sp *tracing.Span
}

var _ sqlutil.InternalRows = &rowsIterator{}
var _ tree.InternalRows = &rowsIterator{}

func (r *rowsIterator) Next(ctx context.Context) (_ bool, retErr error) {
	__antithesis_instrumentation__.Notify(497765)

	defer func() {
		__antithesis_instrumentation__.Notify(497771)

		if r.done {
			__antithesis_instrumentation__.Notify(497774)

			_ = r.Close()
		} else {
			__antithesis_instrumentation__.Notify(497775)
		}
		__antithesis_instrumentation__.Notify(497772)
		if r.errCallback != nil {
			__antithesis_instrumentation__.Notify(497776)
			r.lastErr = r.errCallback(r.lastErr)
			r.errCallback = nil
		} else {
			__antithesis_instrumentation__.Notify(497777)
		}
		__antithesis_instrumentation__.Notify(497773)
		retErr = r.lastErr
	}()
	__antithesis_instrumentation__.Notify(497766)

	if r.done {
		__antithesis_instrumentation__.Notify(497778)
		return false, r.lastErr
	} else {
		__antithesis_instrumentation__.Notify(497779)
	}
	__antithesis_instrumentation__.Notify(497767)

	handleDataObject := func(data ieIteratorResult) (bool, error) {
		__antithesis_instrumentation__.Notify(497780)
		if data.row != nil {
			__antithesis_instrumentation__.Notify(497785)
			r.rowsAffected++

			r.lastRow = data.row
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(497786)
		}
		__antithesis_instrumentation__.Notify(497781)
		if data.rowsAffectedIncrement != nil {
			__antithesis_instrumentation__.Notify(497787)
			r.rowsAffected += *data.rowsAffectedIncrement
			return r.Next(ctx)
		} else {
			__antithesis_instrumentation__.Notify(497788)
		}
		__antithesis_instrumentation__.Notify(497782)
		if data.cols != nil {
			__antithesis_instrumentation__.Notify(497789)

			if r.resultCols == nil {
				__antithesis_instrumentation__.Notify(497791)
				r.resultCols = data.cols
			} else {
				__antithesis_instrumentation__.Notify(497792)
			}
			__antithesis_instrumentation__.Notify(497790)
			return r.Next(ctx)
		} else {
			__antithesis_instrumentation__.Notify(497793)
		}
		__antithesis_instrumentation__.Notify(497783)
		if data.err == nil {
			__antithesis_instrumentation__.Notify(497794)
			data.err = errors.AssertionFailedf("unexpectedly empty ieIteratorResult object")
		} else {
			__antithesis_instrumentation__.Notify(497795)
		}
		__antithesis_instrumentation__.Notify(497784)
		r.lastErr = data.err
		r.done = true
		return false, r.lastErr
	}
	__antithesis_instrumentation__.Notify(497768)

	if r.first != nil {
		__antithesis_instrumentation__.Notify(497796)

		first := r.first
		r.first = nil
		return handleDataObject(*first)
	} else {
		__antithesis_instrumentation__.Notify(497797)
	}
	__antithesis_instrumentation__.Notify(497769)

	var next ieIteratorResult
	next, r.done, r.lastErr = r.r.nextResult(ctx)
	if r.done || func() bool {
		__antithesis_instrumentation__.Notify(497798)
		return r.lastErr != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(497799)
		return false, r.lastErr
	} else {
		__antithesis_instrumentation__.Notify(497800)
	}
	__antithesis_instrumentation__.Notify(497770)
	return handleDataObject(next)
}

func (r *rowsIterator) Cur() tree.Datums {
	__antithesis_instrumentation__.Notify(497801)
	return r.lastRow
}

func (r *rowsIterator) Close() error {
	__antithesis_instrumentation__.Notify(497802)

	r.stmtBuf.Close()

	defer func() {
		__antithesis_instrumentation__.Notify(497805)
		if r.sp != nil {
			__antithesis_instrumentation__.Notify(497806)
			r.wg.Wait()
			r.sp.Finish()
			r.sp = nil
		} else {
			__antithesis_instrumentation__.Notify(497807)
		}
	}()
	__antithesis_instrumentation__.Notify(497803)

	if err := r.r.close(); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(497808)
		return r.lastErr == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(497809)
		r.lastErr = err
	} else {
		__antithesis_instrumentation__.Notify(497810)
	}
	__antithesis_instrumentation__.Notify(497804)
	return r.lastErr
}

func (r *rowsIterator) Types() colinfo.ResultColumns {
	__antithesis_instrumentation__.Notify(497811)
	return r.resultCols
}

func (ie *InternalExecutor) QueryBuffered(
	ctx context.Context, opName string, txn *kv.Txn, stmt string, qargs ...interface{},
) ([]tree.Datums, error) {
	__antithesis_instrumentation__.Notify(497812)
	return ie.QueryBufferedEx(ctx, opName, txn, ie.maybeRootSessionDataOverride(opName), stmt, qargs...)
}

func (ie *InternalExecutor) QueryBufferedEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) ([]tree.Datums, error) {
	__antithesis_instrumentation__.Notify(497813)
	datums, _, err := ie.queryInternalBuffered(ctx, opName, txn, session, stmt, 0, qargs...)
	return datums, err
}

func (ie *InternalExecutor) QueryBufferedExWithCols(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) ([]tree.Datums, colinfo.ResultColumns, error) {
	__antithesis_instrumentation__.Notify(497814)
	datums, cols, err := ie.queryInternalBuffered(ctx, opName, txn, session, stmt, 0, qargs...)
	return datums, cols, err
}

func (ie *InternalExecutor) queryInternalBuffered(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	sessionDataOverride sessiondata.InternalExecutorOverride,
	stmt string,

	limit int,
	qargs ...interface{},
) ([]tree.Datums, colinfo.ResultColumns, error) {
	__antithesis_instrumentation__.Notify(497815)

	rw := newAsyncIEResultChannel()
	it, err := ie.execInternal(ctx, opName, rw, txn, sessionDataOverride, stmt, qargs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(497819)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(497820)
	}
	__antithesis_instrumentation__.Notify(497816)
	var rows []tree.Datums
	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(497821)
		rows = append(rows, it.Cur())
		if limit != 0 && func() bool {
			__antithesis_instrumentation__.Notify(497822)
			return len(rows) == limit == true
		}() == true {
			__antithesis_instrumentation__.Notify(497823)

			err = it.Close()
			break
		} else {
			__antithesis_instrumentation__.Notify(497824)
		}
	}
	__antithesis_instrumentation__.Notify(497817)
	if err != nil {
		__antithesis_instrumentation__.Notify(497825)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(497826)
	}
	__antithesis_instrumentation__.Notify(497818)
	return rows, it.Types(), nil
}

func (ie *InternalExecutor) QueryRow(
	ctx context.Context, opName string, txn *kv.Txn, stmt string, qargs ...interface{},
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(497827)
	return ie.QueryRowEx(ctx, opName, txn, ie.maybeRootSessionDataOverride(opName), stmt, qargs...)
}

func (ie *InternalExecutor) QueryRowEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(497828)
	rows, _, err := ie.QueryRowExWithCols(ctx, opName, txn, session, stmt, qargs...)
	return rows, err
}

func (ie *InternalExecutor) QueryRowExWithCols(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (tree.Datums, colinfo.ResultColumns, error) {
	__antithesis_instrumentation__.Notify(497829)
	rows, cols, err := ie.queryInternalBuffered(ctx, opName, txn, session, stmt, 2, qargs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(497831)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(497832)
	}
	__antithesis_instrumentation__.Notify(497830)
	switch len(rows) {
	case 0:
		__antithesis_instrumentation__.Notify(497833)
		return nil, nil, nil
	case 1:
		__antithesis_instrumentation__.Notify(497834)
		return rows[0], cols, nil
	default:
		__antithesis_instrumentation__.Notify(497835)
		return nil, nil, &tree.MultipleResultsError{SQL: stmt}
	}
}

func (ie *InternalExecutor) Exec(
	ctx context.Context, opName string, txn *kv.Txn, stmt string, qargs ...interface{},
) (int, error) {
	__antithesis_instrumentation__.Notify(497836)
	return ie.ExecEx(ctx, opName, txn, ie.maybeRootSessionDataOverride(opName), stmt, qargs...)
}

func (ie *InternalExecutor) ExecEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (int, error) {
	__antithesis_instrumentation__.Notify(497837)

	rw := newAsyncIEResultChannel()
	it, err := ie.execInternal(ctx, opName, rw, txn, session, stmt, qargs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(497841)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(497842)
	}
	__antithesis_instrumentation__.Notify(497838)

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(497843)
	}
	__antithesis_instrumentation__.Notify(497839)
	if err != nil {
		__antithesis_instrumentation__.Notify(497844)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(497845)
	}
	__antithesis_instrumentation__.Notify(497840)
	return it.rowsAffected, nil
}

func (ie *InternalExecutor) QueryIterator(
	ctx context.Context, opName string, txn *kv.Txn, stmt string, qargs ...interface{},
) (sqlutil.InternalRows, error) {
	__antithesis_instrumentation__.Notify(497846)
	return ie.QueryIteratorEx(ctx, opName, txn, ie.maybeRootSessionDataOverride(opName), stmt, qargs...)
}

func (ie *InternalExecutor) QueryIteratorEx(
	ctx context.Context,
	opName string,
	txn *kv.Txn,
	session sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (sqlutil.InternalRows, error) {
	__antithesis_instrumentation__.Notify(497847)
	return ie.execInternal(
		ctx, opName, newSyncIEResultChannel(), txn, session, stmt, qargs...,
	)
}

func applyOverrides(o sessiondata.InternalExecutorOverride, sd *sessiondata.SessionData) {
	__antithesis_instrumentation__.Notify(497848)
	if !o.User.Undefined() {
		__antithesis_instrumentation__.Notify(497854)
		sd.UserProto = o.User.EncodeProto()
	} else {
		__antithesis_instrumentation__.Notify(497855)
	}
	__antithesis_instrumentation__.Notify(497849)
	if o.Database != "" {
		__antithesis_instrumentation__.Notify(497856)
		sd.Database = o.Database
	} else {
		__antithesis_instrumentation__.Notify(497857)
	}
	__antithesis_instrumentation__.Notify(497850)
	if o.ApplicationName != "" {
		__antithesis_instrumentation__.Notify(497858)
		sd.ApplicationName = o.ApplicationName
	} else {
		__antithesis_instrumentation__.Notify(497859)
	}
	__antithesis_instrumentation__.Notify(497851)
	if o.SearchPath != nil {
		__antithesis_instrumentation__.Notify(497860)
		sd.SearchPath = *o.SearchPath
	} else {
		__antithesis_instrumentation__.Notify(497861)
	}
	__antithesis_instrumentation__.Notify(497852)
	if o.DatabaseIDToTempSchemaID != nil {
		__antithesis_instrumentation__.Notify(497862)
		sd.DatabaseIDToTempSchemaID = o.DatabaseIDToTempSchemaID
	} else {
		__antithesis_instrumentation__.Notify(497863)
	}
	__antithesis_instrumentation__.Notify(497853)
	if o.QualityOfService != nil {
		__antithesis_instrumentation__.Notify(497864)
		sd.DefaultTxnQualityOfService = o.QualityOfService.ValidateInternal()
	} else {
		__antithesis_instrumentation__.Notify(497865)
	}
}

func (ie *InternalExecutor) maybeRootSessionDataOverride(
	opName string,
) sessiondata.InternalExecutorOverride {
	__antithesis_instrumentation__.Notify(497866)
	if ie.sessionDataStack == nil {
		__antithesis_instrumentation__.Notify(497870)
		return sessiondata.InternalExecutorOverride{
			User:            security.RootUserName(),
			ApplicationName: catconstants.InternalAppNamePrefix + "-" + opName,
		}
	} else {
		__antithesis_instrumentation__.Notify(497871)
	}
	__antithesis_instrumentation__.Notify(497867)
	o := sessiondata.InternalExecutorOverride{}
	sd := ie.sessionDataStack.Top()
	if sd.User().Undefined() {
		__antithesis_instrumentation__.Notify(497872)
		o.User = security.RootUserName()
	} else {
		__antithesis_instrumentation__.Notify(497873)
	}
	__antithesis_instrumentation__.Notify(497868)
	if sd.ApplicationName == "" {
		__antithesis_instrumentation__.Notify(497874)
		o.ApplicationName = catconstants.InternalAppNamePrefix + "-" + opName
	} else {
		__antithesis_instrumentation__.Notify(497875)
	}
	__antithesis_instrumentation__.Notify(497869)
	return o
}

var rowsAffectedResultColumns = colinfo.ResultColumns{
	colinfo.ResultColumn{
		Name: "rows_affected",
		Typ:  types.Int,
	},
}

func (ie *InternalExecutor) execInternal(
	ctx context.Context,
	opName string,
	rw *ieResultChannel,
	txn *kv.Txn,
	sessionDataOverride sessiondata.InternalExecutorOverride,
	stmt string,
	qargs ...interface{},
) (r *rowsIterator, retErr error) {
	__antithesis_instrumentation__.Notify(497876)
	ctx = logtags.AddTag(ctx, "intExec", opName)

	var sd *sessiondata.SessionData
	if ie.sessionDataStack != nil {
		__antithesis_instrumentation__.Notify(497890)

		sd = ie.sessionDataStack.Top().Clone()
	} else {
		__antithesis_instrumentation__.Notify(497891)
		sd = ie.s.newSessionData(SessionArgs{})
	}
	__antithesis_instrumentation__.Notify(497877)
	applyOverrides(sessionDataOverride, sd)
	sd.Internal = true
	if sd.User().Undefined() {
		__antithesis_instrumentation__.Notify(497892)
		return nil, errors.AssertionFailedf("no user specified for internal query")
	} else {
		__antithesis_instrumentation__.Notify(497893)
	}
	__antithesis_instrumentation__.Notify(497878)
	if sd.ApplicationName == "" {
		__antithesis_instrumentation__.Notify(497894)
		sd.ApplicationName = catconstants.InternalAppNamePrefix + "-" + opName
	} else {
		__antithesis_instrumentation__.Notify(497895)
	}
	__antithesis_instrumentation__.Notify(497879)

	ctx, sp := tracing.EnsureChildSpan(ctx, ie.s.cfg.AmbientCtx.Tracer, opName)
	stmtBuf := NewStmtBuf()
	var wg sync.WaitGroup

	defer func() {
		__antithesis_instrumentation__.Notify(497896)

		if retErr != nil || func() bool {
			__antithesis_instrumentation__.Notify(497897)
			return r == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(497898)

			if retErr != nil && func() bool {
				__antithesis_instrumentation__.Notify(497900)
				return !errIsRetriable(retErr) == true
			}() == true {
				__antithesis_instrumentation__.Notify(497901)
				retErr = errors.Wrapf(retErr, "%s", opName)
			} else {
				__antithesis_instrumentation__.Notify(497902)
			}
			__antithesis_instrumentation__.Notify(497899)
			stmtBuf.Close()
			wg.Wait()
			sp.Finish()
		} else {
			__antithesis_instrumentation__.Notify(497903)
			r.errCallback = func(err error) error {
				__antithesis_instrumentation__.Notify(497905)
				if err != nil && func() bool {
					__antithesis_instrumentation__.Notify(497907)
					return !errIsRetriable(err) == true
				}() == true {
					__antithesis_instrumentation__.Notify(497908)
					err = errors.Wrapf(err, "%s", opName)
				} else {
					__antithesis_instrumentation__.Notify(497909)
				}
				__antithesis_instrumentation__.Notify(497906)
				return err
			}
			__antithesis_instrumentation__.Notify(497904)
			r.sp = sp
		}
	}()
	__antithesis_instrumentation__.Notify(497880)

	timeReceived := timeutil.Now()
	parseStart := timeReceived
	parsed, err := parser.ParseOne(stmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(497910)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(497911)
	}
	__antithesis_instrumentation__.Notify(497881)
	parseEnd := timeutil.Now()

	datums, err := golangFillQueryArguments(qargs...)
	if err != nil {
		__antithesis_instrumentation__.Notify(497912)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(497913)
	}
	__antithesis_instrumentation__.Notify(497882)

	var resPos CmdPos

	syncCallback := func(results []resWithPos) {
		__antithesis_instrumentation__.Notify(497914)

		stmtBuf.Close()
		for _, res := range results {
			__antithesis_instrumentation__.Notify(497916)
			if res.Err() != nil {
				__antithesis_instrumentation__.Notify(497918)

				_ = rw.addResult(ctx, ieIteratorResult{err: res.Err()})
				return
			} else {
				__antithesis_instrumentation__.Notify(497919)
			}
			__antithesis_instrumentation__.Notify(497917)
			if res.pos == resPos {
				__antithesis_instrumentation__.Notify(497920)
				return
			} else {
				__antithesis_instrumentation__.Notify(497921)
			}
		}
		__antithesis_instrumentation__.Notify(497915)
		_ = rw.addResult(ctx, ieIteratorResult{
			err: errors.AssertionFailedf(
				"missing result for pos: %d and no previous error", resPos,
			),
		})
	}
	__antithesis_instrumentation__.Notify(497883)

	errCallback := func(err error) {
		__antithesis_instrumentation__.Notify(497922)
		_ = rw.addResult(ctx, ieIteratorResult{err: err})
	}
	__antithesis_instrumentation__.Notify(497884)
	ie.initConnEx(ctx, txn, rw, sd, stmtBuf, &wg, syncCallback, errCallback)

	typeHints := make(tree.PlaceholderTypes, len(datums))
	for i, d := range datums {
		__antithesis_instrumentation__.Notify(497923)

		typeHints[tree.PlaceholderIdx(i)] = d.ResolvedType()
	}
	__antithesis_instrumentation__.Notify(497885)
	if len(qargs) == 0 {
		__antithesis_instrumentation__.Notify(497924)
		resPos = 0
		if err := stmtBuf.Push(
			ctx,
			ExecStmt{
				Statement:    parsed,
				TimeReceived: timeReceived,
				ParseStart:   parseStart,
				ParseEnd:     parseEnd,

				LastInBatch: true,
			}); err != nil {
			__antithesis_instrumentation__.Notify(497925)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(497926)
		}
	} else {
		__antithesis_instrumentation__.Notify(497927)
		resPos = 2
		if err := stmtBuf.Push(
			ctx,
			PrepareStmt{
				Statement:  parsed,
				ParseStart: parseStart,
				ParseEnd:   parseEnd,
				TypeHints:  typeHints,
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(497930)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(497931)
		}
		__antithesis_instrumentation__.Notify(497928)

		if err := stmtBuf.Push(ctx, BindStmt{internalArgs: datums}); err != nil {
			__antithesis_instrumentation__.Notify(497932)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(497933)
		}
		__antithesis_instrumentation__.Notify(497929)

		if err := stmtBuf.Push(ctx,
			ExecPortal{
				TimeReceived: timeReceived,

				FollowedBySync: true,
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(497934)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(497935)
		}
	}
	__antithesis_instrumentation__.Notify(497886)
	if err := stmtBuf.Push(ctx, Sync{}); err != nil {
		__antithesis_instrumentation__.Notify(497936)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(497937)
	}
	__antithesis_instrumentation__.Notify(497887)
	r = &rowsIterator{
		r:       rw,
		stmtBuf: stmtBuf,
		wg:      &wg,
	}

	if parsed.AST.StatementReturnType() != tree.Rows {
		__antithesis_instrumentation__.Notify(497938)
		r.resultCols = rowsAffectedResultColumns
	} else {
		__antithesis_instrumentation__.Notify(497939)
	}

	{
		__antithesis_instrumentation__.Notify(497940)
		var first ieIteratorResult
		if first, r.done, r.lastErr = rw.firstResult(ctx); !r.done {
			__antithesis_instrumentation__.Notify(497941)
			r.first = &first
		} else {
			__antithesis_instrumentation__.Notify(497942)
		}
	}
	__antithesis_instrumentation__.Notify(497888)
	if !r.done && func() bool {
		__antithesis_instrumentation__.Notify(497943)
		return r.first.cols != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(497944)

		if r.resultCols == nil {
			__antithesis_instrumentation__.Notify(497946)
			r.resultCols = r.first.cols
		} else {
			__antithesis_instrumentation__.Notify(497947)
		}
		__antithesis_instrumentation__.Notify(497945)
		var first ieIteratorResult
		first, r.done, r.lastErr = rw.nextResult(ctx)
		if !r.done {
			__antithesis_instrumentation__.Notify(497948)
			r.first = &first
		} else {
			__antithesis_instrumentation__.Notify(497949)
		}
	} else {
		__antithesis_instrumentation__.Notify(497950)
	}
	__antithesis_instrumentation__.Notify(497889)

	return r, nil
}

type internalClientComm struct {
	results []resWithPos

	w ieResultWriter

	lastDelivered CmdPos

	sync func([]resWithPos)
}

var _ ClientComm = &internalClientComm{}

type resWithPos struct {
	*streamingCommandResult
	pos CmdPos
}

func (icc *internalClientComm) CreateStatementResult(
	_ tree.Statement,
	_ RowDescOpt,
	pos CmdPos,
	_ []pgwirebase.FormatCode,
	_ sessiondatapb.DataConversionConfig,
	_ *time.Location,
	_ int,
	_ string,
	_ bool,
) CommandResult {
	__antithesis_instrumentation__.Notify(497951)
	return icc.createRes(pos, nil)
}

func (icc *internalClientComm) createRes(pos CmdPos, onClose func()) *streamingCommandResult {
	__antithesis_instrumentation__.Notify(497952)
	res := &streamingCommandResult{
		w: icc.w,
		closeCallback: func(res *streamingCommandResult, typ resCloseType) {
			__antithesis_instrumentation__.Notify(497954)
			if typ == discarded {
				__antithesis_instrumentation__.Notify(497956)
				return
			} else {
				__antithesis_instrumentation__.Notify(497957)
			}
			__antithesis_instrumentation__.Notify(497955)
			icc.results = append(icc.results, resWithPos{streamingCommandResult: res, pos: pos})
			if onClose != nil {
				__antithesis_instrumentation__.Notify(497958)
				onClose()
			} else {
				__antithesis_instrumentation__.Notify(497959)
			}
		},
	}
	__antithesis_instrumentation__.Notify(497953)
	return res
}

func (icc *internalClientComm) CreatePrepareResult(pos CmdPos) ParseResult {
	__antithesis_instrumentation__.Notify(497960)
	return icc.createRes(pos, nil)
}

func (icc *internalClientComm) CreateBindResult(pos CmdPos) BindResult {
	__antithesis_instrumentation__.Notify(497961)
	return icc.createRes(pos, nil)
}

func (icc *internalClientComm) CreateSyncResult(pos CmdPos) SyncResult {
	__antithesis_instrumentation__.Notify(497962)
	return icc.createRes(pos, func() {
		__antithesis_instrumentation__.Notify(497963)
		results := make([]resWithPos, len(icc.results))
		copy(results, icc.results)
		icc.results = icc.results[:0]
		icc.sync(results)
		icc.lastDelivered = pos
	})
}

func (icc *internalClientComm) LockCommunication() ClientLock {
	__antithesis_instrumentation__.Notify(497964)
	return (*noopClientLock)(icc)
}

func (icc *internalClientComm) Flush(pos CmdPos) error {
	__antithesis_instrumentation__.Notify(497965)
	return nil
}

func (icc *internalClientComm) CreateDescribeResult(pos CmdPos) DescribeResult {
	__antithesis_instrumentation__.Notify(497966)
	return icc.createRes(pos, nil)
}

func (icc *internalClientComm) CreateDeleteResult(pos CmdPos) DeleteResult {
	__antithesis_instrumentation__.Notify(497967)
	panic("unimplemented")
}

func (icc *internalClientComm) CreateFlushResult(pos CmdPos) FlushResult {
	__antithesis_instrumentation__.Notify(497968)
	panic("unimplemented")
}

func (icc *internalClientComm) CreateErrorResult(pos CmdPos) ErrorResult {
	__antithesis_instrumentation__.Notify(497969)
	panic("unimplemented")
}

func (icc *internalClientComm) CreateEmptyQueryResult(pos CmdPos) EmptyQueryResult {
	__antithesis_instrumentation__.Notify(497970)
	panic("unimplemented")
}

func (icc *internalClientComm) CreateCopyInResult(pos CmdPos) CopyInResult {
	__antithesis_instrumentation__.Notify(497971)
	panic("unimplemented")
}

func (icc *internalClientComm) CreateDrainResult(pos CmdPos) DrainResult {
	__antithesis_instrumentation__.Notify(497972)
	panic("unimplemented")
}

type noopClientLock internalClientComm

func (ncl *noopClientLock) Close() { __antithesis_instrumentation__.Notify(497973) }

func (ncl *noopClientLock) ClientPos() CmdPos {
	__antithesis_instrumentation__.Notify(497974)
	return ncl.lastDelivered
}

func (ncl *noopClientLock) RTrim(_ context.Context, pos CmdPos) {
	__antithesis_instrumentation__.Notify(497975)
	var i int
	var r resWithPos
	for i, r = range ncl.results {
		__antithesis_instrumentation__.Notify(497977)
		if r.pos >= pos {
			__antithesis_instrumentation__.Notify(497978)
			break
		} else {
			__antithesis_instrumentation__.Notify(497979)
		}
	}
	__antithesis_instrumentation__.Notify(497976)
	ncl.results = ncl.results[:i]
}
