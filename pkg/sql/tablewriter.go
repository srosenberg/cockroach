package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/mutations"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type expressionCarrier interface {
	walkExprs(func(desc string, index int, expr tree.TypedExpr))
}

type tableWriter interface {
	expressionCarrier

	init(context.Context, *kv.Txn, *tree.EvalContext, *settings.Values) error

	row(context.Context, tree.Datums, row.PartialIndexUpdateHelper, bool) error

	flushAndStartNewBatch(context.Context) error

	finalize(context.Context) error

	tableDesc() catalog.TableDescriptor

	close(context.Context)

	desc() string

	enableAutoCommit()
}

type autoCommitOpt int

const (
	autoCommitDisabled autoCommitOpt = 0
	autoCommitEnabled  autoCommitOpt = 1
)

type tableWriterBase struct {
	txn *kv.Txn

	desc catalog.TableDescriptor

	autoCommit autoCommitOpt

	b *kv.Batch

	lockTimeout time.Duration

	maxBatchSize int

	maxBatchByteSize int

	currentBatchSize int

	lastBatchSize int

	rowsWritten int64

	rowsWrittenLimit int64

	rows *rowcontainer.RowContainer

	forceProductionBatchSizes bool

	sv *settings.Values
}

var maxBatchBytes = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.mutations.mutation_batch_byte_size",
	"byte size - in key and value lengths -- for mutation batches",
	4<<20,
)

func (tb *tableWriterBase) init(
	txn *kv.Txn,
	tableDesc catalog.TableDescriptor,
	evalCtx *tree.EvalContext,
	settings *settings.Values,
) {
	tb.txn = txn
	tb.desc = tableDesc
	tb.lockTimeout = 0
	if evalCtx != nil {
		tb.lockTimeout = evalCtx.SessionData().LockTimeout
	}
	tb.forceProductionBatchSizes = evalCtx != nil && evalCtx.TestingKnobs.ForceProductionBatchSizes
	tb.maxBatchSize = mutations.MaxBatchSize(tb.forceProductionBatchSizes)
	batchMaxBytes := int(maxBatchBytes.Default())
	if evalCtx != nil {
		batchMaxBytes = int(maxBatchBytes.Get(&evalCtx.Settings.SV))
	}
	tb.maxBatchByteSize = mutations.MaxBatchByteSize(batchMaxBytes, tb.forceProductionBatchSizes)
	tb.sv = settings
	tb.initNewBatch()
}

func (tb *tableWriterBase) setRowsWrittenLimit(sd *sessiondata.SessionData) {
	__antithesis_instrumentation__.Notify(627713)
	if sd != nil && func() bool {
		__antithesis_instrumentation__.Notify(627714)
		return !sd.Internal == true
	}() == true {
		__antithesis_instrumentation__.Notify(627715)

		tb.rowsWrittenLimit = sd.TxnRowsWrittenErr
	} else {
		__antithesis_instrumentation__.Notify(627716)
	}
}

func (tb *tableWriterBase) flushAndStartNewBatch(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(627717)
	if err := tb.txn.Run(ctx, tb.b); err != nil {
		__antithesis_instrumentation__.Notify(627720)
		return row.ConvertBatchError(ctx, tb.desc, tb.b)
	} else {
		__antithesis_instrumentation__.Notify(627721)
	}
	__antithesis_instrumentation__.Notify(627718)
	if err := tb.tryDoResponseAdmission(ctx); err != nil {
		__antithesis_instrumentation__.Notify(627722)
		return err
	} else {
		__antithesis_instrumentation__.Notify(627723)
	}
	__antithesis_instrumentation__.Notify(627719)
	tb.initNewBatch()
	tb.rowsWritten += int64(tb.currentBatchSize)
	tb.lastBatchSize = tb.currentBatchSize
	tb.currentBatchSize = 0
	return nil
}

func (tb *tableWriterBase) finalize(ctx context.Context) (err error) {
	__antithesis_instrumentation__.Notify(627724)

	tb.rowsWritten += int64(tb.currentBatchSize)
	if tb.autoCommit == autoCommitEnabled && func() bool {
		__antithesis_instrumentation__.Notify(627727)
		return (tb.rowsWrittenLimit == 0 || func() bool {
			__antithesis_instrumentation__.Notify(627728)
			return tb.rowsWritten <= tb.rowsWrittenLimit == true
		}() == true) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(627729)
		return !tb.txn.DeadlineLikelySufficient(tb.sv) == true
	}() == true {
		__antithesis_instrumentation__.Notify(627730)
		log.Event(ctx, "autocommit enabled")

		err = tb.txn.CommitInBatch(ctx, tb.b)
	} else {
		__antithesis_instrumentation__.Notify(627731)
		err = tb.txn.Run(ctx, tb.b)
	}
	__antithesis_instrumentation__.Notify(627725)
	tb.lastBatchSize = tb.currentBatchSize
	if err != nil {
		__antithesis_instrumentation__.Notify(627732)
		return row.ConvertBatchError(ctx, tb.desc, tb.b)
	} else {
		__antithesis_instrumentation__.Notify(627733)
	}
	__antithesis_instrumentation__.Notify(627726)
	return tb.tryDoResponseAdmission(ctx)
}

func (tb *tableWriterBase) tryDoResponseAdmission(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(627734)

	responseAdmissionQ := tb.txn.DB().SQLKVResponseAdmissionQ
	if responseAdmissionQ != nil {
		__antithesis_instrumentation__.Notify(627736)
		requestAdmissionHeader := tb.txn.AdmissionHeader()
		responseAdmission := admission.WorkInfo{
			TenantID:   roachpb.SystemTenantID,
			Priority:   admission.WorkPriority(requestAdmissionHeader.Priority),
			CreateTime: requestAdmissionHeader.CreateTime,
		}
		if _, err := responseAdmissionQ.Admit(ctx, responseAdmission); err != nil {
			__antithesis_instrumentation__.Notify(627737)
			return err
		} else {
			__antithesis_instrumentation__.Notify(627738)
		}
	} else {
		__antithesis_instrumentation__.Notify(627739)
	}
	__antithesis_instrumentation__.Notify(627735)
	return nil
}

func (tb *tableWriterBase) enableAutoCommit() {
	__antithesis_instrumentation__.Notify(627740)
	tb.autoCommit = autoCommitEnabled
}

func (tb *tableWriterBase) initNewBatch() {
	__antithesis_instrumentation__.Notify(627741)
	tb.b = tb.txn.NewBatch()
	tb.b.Header.LockTimeout = tb.lockTimeout
}

func (tb *tableWriterBase) clearLastBatch(ctx context.Context) {
	__antithesis_instrumentation__.Notify(627742)
	tb.lastBatchSize = 0
	if tb.rows != nil {
		__antithesis_instrumentation__.Notify(627743)
		tb.rows.Clear(ctx)
	} else {
		__antithesis_instrumentation__.Notify(627744)
	}
}

func (tb *tableWriterBase) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(627745)
	if tb.rows != nil {
		__antithesis_instrumentation__.Notify(627746)
		tb.rows.Close(ctx)
		tb.rows = nil
	} else {
		__antithesis_instrumentation__.Notify(627747)
	}
}
