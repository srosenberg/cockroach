package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
)

const (
	NodelocalFileUploadTable = "nodelocal_file_upload"

	UserFileUploadTable = "user_file_upload"
)

var _ copyMachineInterface = &fileUploadMachine{}

type fileUploadMachine struct {
	c              *copyMachine
	w              io.WriteCloser
	cancel         func()
	failureCleanup func()
}

var _ io.ReadSeeker = &noopReadSeeker{}

type noopReadSeeker struct {
	*io.PipeReader
}

func (n *noopReadSeeker) Seek(int64, int) (int64, error) {
	__antithesis_instrumentation__.Notify(460994)
	return 0, errors.New("illegal seek")
}

func checkIfFileExists(ctx context.Context, c *copyMachine, dest, copyTargetTable string) error {
	__antithesis_instrumentation__.Notify(460995)
	if copyTargetTable == UserFileUploadTable {
		__antithesis_instrumentation__.Notify(460999)
		dest = strings.TrimSuffix(dest, ".tmp")
	} else {
		__antithesis_instrumentation__.Notify(461000)
	}
	__antithesis_instrumentation__.Notify(460996)
	store, err := c.p.execCfg.DistSQLSrv.ExternalStorageFromURI(ctx, dest, c.p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(461001)
		return err
	} else {
		__antithesis_instrumentation__.Notify(461002)
	}
	__antithesis_instrumentation__.Notify(460997)
	defer store.Close()

	_, err = store.ReadFile(ctx, "")
	if err == nil {
		__antithesis_instrumentation__.Notify(461003)

		uri, _ := url.Parse(dest)
		return errors.Newf("destination file already exists for %s", uri.Path)
	} else {
		__antithesis_instrumentation__.Notify(461004)
	}
	__antithesis_instrumentation__.Notify(460998)

	return nil
}

func newFileUploadMachine(
	ctx context.Context,
	conn pgwirebase.Conn,
	n *tree.CopyFrom,
	txnOpt copyTxnOpt,
	execCfg *ExecutorConfig,
) (f *fileUploadMachine, retErr error) {
	__antithesis_instrumentation__.Notify(461005)
	if len(n.Columns) != 0 {
		__antithesis_instrumentation__.Notify(461016)
		return nil, errors.New("expected 0 columns specified for file uploads")
	} else {
		__antithesis_instrumentation__.Notify(461017)
	}
	__antithesis_instrumentation__.Notify(461006)
	c := &copyMachine{
		conn: conn,

		p: planner{execCfg: execCfg, alloc: &tree.DatumAlloc{}},
	}
	f = &fileUploadMachine{
		c: c,
	}

	cleanup := c.p.preparePlannerForCopy(ctx, txnOpt)
	defer func() {
		__antithesis_instrumentation__.Notify(461018)
		retErr = cleanup(ctx, retErr)
	}()
	__antithesis_instrumentation__.Notify(461007)
	c.parsingEvalCtx = c.p.EvalContext()

	if n.Table.Table() == NodelocalFileUploadTable {
		__antithesis_instrumentation__.Notify(461019)
		if err := c.p.RequireAdminRole(ctx, "upload to nodelocal"); err != nil {
			__antithesis_instrumentation__.Notify(461020)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(461021)
		}
	} else {
		__antithesis_instrumentation__.Notify(461022)
	}
	__antithesis_instrumentation__.Notify(461008)

	if n.Options.Destination == nil {
		__antithesis_instrumentation__.Notify(461023)
		return nil, errors.Newf("destination required")
	} else {
		__antithesis_instrumentation__.Notify(461024)
	}
	__antithesis_instrumentation__.Notify(461009)
	destFn, err := f.c.p.TypeAsString(ctx, n.Options.Destination, "COPY")
	if err != nil {
		__antithesis_instrumentation__.Notify(461025)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461026)
	}
	__antithesis_instrumentation__.Notify(461010)
	dest, err := destFn()
	if err != nil {
		__antithesis_instrumentation__.Notify(461027)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461028)
	}
	__antithesis_instrumentation__.Notify(461011)

	store, err := c.p.execCfg.DistSQLSrv.ExternalStorageFromURI(ctx, dest, c.p.User())
	if err != nil {
		__antithesis_instrumentation__.Notify(461029)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461030)
	}
	__antithesis_instrumentation__.Notify(461012)

	err = checkIfFileExists(ctx, c, dest, n.Table.Table())
	if err != nil {
		__antithesis_instrumentation__.Notify(461031)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461032)
	}
	__antithesis_instrumentation__.Notify(461013)

	writeCtx, canecelWriteCtx := context.WithCancel(ctx)
	f.cancel = canecelWriteCtx
	f.w, err = store.Writer(writeCtx, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(461033)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(461034)
	}
	__antithesis_instrumentation__.Notify(461014)

	f.failureCleanup = func() {
		__antithesis_instrumentation__.Notify(461035)

		_ = store.Delete(ctx, "")
	}
	__antithesis_instrumentation__.Notify(461015)

	c.resultColumns = make(colinfo.ResultColumns, 1)
	c.resultColumns[0] = colinfo.ResultColumn{Typ: types.Bytes}
	c.parsingEvalCtx = c.p.EvalContext()
	c.rowsMemAcc = c.p.extendedEvalCtx.Mon.MakeBoundAccount()
	c.bufMemAcc = c.p.extendedEvalCtx.Mon.MakeBoundAccount()
	c.processRows = f.writeFile
	c.forceNotNull = true
	c.format = tree.CopyFormatText
	c.null = `\N`
	c.delimiter = '\t'
	return
}

func CopyInFileStmt(destination, schema, table string) string {
	__antithesis_instrumentation__.Notify(461036)
	return fmt.Sprintf(
		`COPY %s.%s FROM STDIN WITH destination = %s`,
		pq.QuoteIdentifier(schema),
		pq.QuoteIdentifier(table),
		lexbase.EscapeSQLString(destination),
	)
}

func (f *fileUploadMachine) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(461037)
	err := f.c.run(ctx)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(461040)
		return f.cancel != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(461041)
		f.cancel()
	} else {
		__antithesis_instrumentation__.Notify(461042)
	}
	__antithesis_instrumentation__.Notify(461038)
	err = errors.CombineErrors(f.w.Close(), err)

	if err != nil {
		__antithesis_instrumentation__.Notify(461043)
		f.failureCleanup()
	} else {
		__antithesis_instrumentation__.Notify(461044)
	}
	__antithesis_instrumentation__.Notify(461039)
	return err
}

func (f *fileUploadMachine) writeFile(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(461045)
	for _, r := range f.c.rows {
		__antithesis_instrumentation__.Notify(461048)
		b := []byte(*r[0].(*tree.DBytes))
		n, err := f.w.Write(b)
		if err != nil {
			__antithesis_instrumentation__.Notify(461050)
			return err
		} else {
			__antithesis_instrumentation__.Notify(461051)
		}
		__antithesis_instrumentation__.Notify(461049)
		if n < len(b) {
			__antithesis_instrumentation__.Notify(461052)
			return io.ErrShortWrite
		} else {
			__antithesis_instrumentation__.Notify(461053)
		}
	}
	__antithesis_instrumentation__.Notify(461046)

	_, err := f.w.Write(nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(461054)
		return err
	} else {
		__antithesis_instrumentation__.Notify(461055)
	}
	__antithesis_instrumentation__.Notify(461047)
	f.c.insertedRows += len(f.c.rows)
	f.c.rows = f.c.rows[:0]
	f.c.rowsMemAcc.Clear(ctx)
	return nil
}
