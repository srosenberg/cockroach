package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding/csv"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type copyMachineInterface interface {
	run(ctx context.Context) error
}

type copyMachine struct {
	table                    tree.TableExpr
	columns                  tree.NameList
	resultColumns            colinfo.ResultColumns
	expectedHiddenColumnIdxs []int
	format                   tree.CopyFormat
	csvEscape                rune
	delimiter                byte

	textDelim   []byte
	null        string
	binaryState binaryState

	forceNotNull bool
	csvInput     bytes.Buffer
	csvReader    *csv.Reader

	buf bytes.Buffer

	rows []tree.Exprs

	insertedRows int

	rowsMemAcc mon.BoundAccount

	bufMemAcc mon.BoundAccount

	conn pgwirebase.Conn

	execInsertPlan func(ctx context.Context, p *planner, res RestrictedCommandResult) error

	txnOpt copyTxnOpt

	p planner

	parsingEvalCtx *tree.EvalContext

	processRows func(ctx context.Context) error
}

func newCopyMachine(
	ctx context.Context,
	conn pgwirebase.Conn,
	n *tree.CopyFrom,
	txnOpt copyTxnOpt,
	execCfg *ExecutorConfig,
	execInsertPlan func(ctx context.Context, p *planner, res RestrictedCommandResult) error,
) (_ *copyMachine, retErr error) {
	__antithesis_instrumentation__.Notify(460616)
	c := &copyMachine{
		conn: conn,

		table:   &n.Table,
		columns: n.Columns,
		format:  n.Options.CopyFormat,
		txnOpt:  txnOpt,

		p:              planner{execCfg: execCfg, alloc: &tree.DatumAlloc{}},
		execInsertPlan: execInsertPlan,
	}

	cleanup := c.p.preparePlannerForCopy(ctx, txnOpt)
	defer func() {
		__antithesis_instrumentation__.Notify(460627)
		retErr = cleanup(ctx, retErr)
	}()
	__antithesis_instrumentation__.Notify(460617)
	c.parsingEvalCtx = c.p.EvalContext()

	switch c.format {
	case tree.CopyFormatText:
		__antithesis_instrumentation__.Notify(460628)
		c.null = `\N`
		c.delimiter = '\t'
	case tree.CopyFormatCSV:
		__antithesis_instrumentation__.Notify(460629)
		c.null = ""
		c.delimiter = ','
	default:
		__antithesis_instrumentation__.Notify(460630)
	}
	__antithesis_instrumentation__.Notify(460618)

	if n.Options.Delimiter != nil {
		__antithesis_instrumentation__.Notify(460631)
		if c.format == tree.CopyFormatBinary {
			__antithesis_instrumentation__.Notify(460636)
			return nil, errors.Newf("DELIMITER unsupported in BINARY format")
		} else {
			__antithesis_instrumentation__.Notify(460637)
		}
		__antithesis_instrumentation__.Notify(460632)
		fn, err := c.p.TypeAsString(ctx, n.Options.Delimiter, "COPY")
		if err != nil {
			__antithesis_instrumentation__.Notify(460638)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(460639)
		}
		__antithesis_instrumentation__.Notify(460633)
		delim, err := fn()
		if err != nil {
			__antithesis_instrumentation__.Notify(460640)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(460641)
		}
		__antithesis_instrumentation__.Notify(460634)
		if len(delim) != 1 || func() bool {
			__antithesis_instrumentation__.Notify(460642)
			return !utf8.ValidString(delim) == true
		}() == true {
			__antithesis_instrumentation__.Notify(460643)
			return nil, errors.Newf("delimiter must be a single-byte character")
		} else {
			__antithesis_instrumentation__.Notify(460644)
		}
		__antithesis_instrumentation__.Notify(460635)
		c.delimiter = delim[0]
	} else {
		__antithesis_instrumentation__.Notify(460645)
	}
	__antithesis_instrumentation__.Notify(460619)
	if n.Options.Null != nil {
		__antithesis_instrumentation__.Notify(460646)
		if c.format == tree.CopyFormatBinary {
			__antithesis_instrumentation__.Notify(460649)
			return nil, errors.Newf("NULL unsupported in BINARY format")
		} else {
			__antithesis_instrumentation__.Notify(460650)
		}
		__antithesis_instrumentation__.Notify(460647)
		fn, err := c.p.TypeAsString(ctx, n.Options.Null, "COPY")
		if err != nil {
			__antithesis_instrumentation__.Notify(460651)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(460652)
		}
		__antithesis_instrumentation__.Notify(460648)
		c.null, err = fn()
		if err != nil {
			__antithesis_instrumentation__.Notify(460653)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(460654)
		}
	} else {
		__antithesis_instrumentation__.Notify(460655)
	}
	__antithesis_instrumentation__.Notify(460620)
	if n.Options.Escape != nil {
		__antithesis_instrumentation__.Notify(460656)
		s := n.Options.Escape.RawString()
		if len(s) != 1 {
			__antithesis_instrumentation__.Notify(460659)
			return nil, pgerror.Newf(
				pgcode.FeatureNotSupported,
				"ESCAPE must be a single rune",
			)
		} else {
			__antithesis_instrumentation__.Notify(460660)
		}
		__antithesis_instrumentation__.Notify(460657)

		if c.format != tree.CopyFormatCSV {
			__antithesis_instrumentation__.Notify(460661)
			return nil, pgerror.Newf(
				pgcode.FeatureNotSupported,
				"ESCAPE can only be specified for CSV",
			)
		} else {
			__antithesis_instrumentation__.Notify(460662)
		}
		__antithesis_instrumentation__.Notify(460658)

		c.csvEscape, _ = utf8.DecodeRuneInString(s)
	} else {
		__antithesis_instrumentation__.Notify(460663)
	}
	__antithesis_instrumentation__.Notify(460621)

	flags := tree.ObjectLookupFlagsWithRequiredTableKind(tree.ResolveRequireTableDesc)
	_, tableDesc, err := resolver.ResolveExistingTableObject(ctx, &c.p, &n.Table, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(460664)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(460665)
	}
	__antithesis_instrumentation__.Notify(460622)
	if err := c.p.CheckPrivilege(ctx, tableDesc, privilege.INSERT); err != nil {
		__antithesis_instrumentation__.Notify(460666)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(460667)
	}
	__antithesis_instrumentation__.Notify(460623)
	cols, err := colinfo.ProcessTargetColumns(tableDesc, n.Columns,
		true, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(460668)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(460669)
	}
	__antithesis_instrumentation__.Notify(460624)
	c.resultColumns = make(colinfo.ResultColumns, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(460670)
		c.resultColumns[i] = colinfo.ResultColumn{
			Name:           col.GetName(),
			Typ:            col.GetType(),
			TableID:        tableDesc.GetID(),
			PGAttributeNum: col.GetPGAttributeNum(),
		}
	}
	__antithesis_instrumentation__.Notify(460625)

	if c.p.SessionData().ExpectAndIgnoreNotVisibleColumnsInCopy && func() bool {
		__antithesis_instrumentation__.Notify(460671)
		return len(n.Columns) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(460672)
		for i, col := range tableDesc.PublicColumns() {
			__antithesis_instrumentation__.Notify(460673)
			if col.IsHidden() {
				__antithesis_instrumentation__.Notify(460674)
				c.expectedHiddenColumnIdxs = append(c.expectedHiddenColumnIdxs, i)
			} else {
				__antithesis_instrumentation__.Notify(460675)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(460676)
	}
	__antithesis_instrumentation__.Notify(460626)
	c.rowsMemAcc = c.p.extendedEvalCtx.Mon.MakeBoundAccount()
	c.bufMemAcc = c.p.extendedEvalCtx.Mon.MakeBoundAccount()
	c.processRows = c.insertRows
	return c, nil
}

type copyTxnOpt struct {
	txn           *kv.Txn
	txnTimestamp  time.Time
	stmtTimestamp time.Time
	resetPlanner  func(ctx context.Context, p *planner, txn *kv.Txn, txnTS time.Time, stmtTS time.Time)

	resetExtraTxnState func(ctx context.Context) error
}

func (c *copyMachine) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(460677)
	defer c.rowsMemAcc.Close(ctx)
	defer c.bufMemAcc.Close(ctx)

	format := pgwirebase.FormatText
	if c.format == tree.CopyFormatBinary {
		__antithesis_instrumentation__.Notify(460682)
		format = pgwirebase.FormatBinary
	} else {
		__antithesis_instrumentation__.Notify(460683)
	}
	__antithesis_instrumentation__.Notify(460678)

	if err := c.conn.BeginCopyIn(ctx, c.resultColumns, format); err != nil {
		__antithesis_instrumentation__.Notify(460684)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460685)
	}
	__antithesis_instrumentation__.Notify(460679)

	readBuf := pgwirebase.MakeReadBuffer(
		pgwirebase.ReadBufferOptionWithClusterSettings(&c.p.execCfg.Settings.SV),
	)

	switch c.format {
	case tree.CopyFormatText:
		__antithesis_instrumentation__.Notify(460686)
		c.textDelim = []byte{c.delimiter}
	case tree.CopyFormatCSV:
		__antithesis_instrumentation__.Notify(460687)
		c.csvInput.Reset()
		c.csvReader = csv.NewReader(&c.csvInput)
		c.csvReader.Comma = rune(c.delimiter)
		c.csvReader.ReuseRecord = true
		c.csvReader.FieldsPerRecord = len(c.resultColumns) + len(c.expectedHiddenColumnIdxs)
		if c.csvEscape != 0 {
			__antithesis_instrumentation__.Notify(460689)
			c.csvReader.Escape = c.csvEscape
		} else {
			__antithesis_instrumentation__.Notify(460690)
		}
	default:
		__antithesis_instrumentation__.Notify(460688)
	}
	__antithesis_instrumentation__.Notify(460680)

Loop:
	for {
		__antithesis_instrumentation__.Notify(460691)
		typ, _, err := readBuf.ReadTypedMsg(c.conn.Rd())
		if err != nil {
			__antithesis_instrumentation__.Notify(460693)
			if pgwirebase.IsMessageTooBigError(err) && func() bool {
				__antithesis_instrumentation__.Notify(460695)
				return typ == pgwirebase.ClientMsgCopyData == true
			}() == true {
				__antithesis_instrumentation__.Notify(460696)

				_, slurpErr := readBuf.SlurpBytes(c.conn.Rd(), pgwirebase.GetMessageTooBigSize(err))
				if slurpErr != nil {
					__antithesis_instrumentation__.Notify(460698)
					return errors.CombineErrors(err, errors.Wrapf(slurpErr, "error slurping remaining bytes in COPY"))
				} else {
					__antithesis_instrumentation__.Notify(460699)
				}
				__antithesis_instrumentation__.Notify(460697)

				for {
					__antithesis_instrumentation__.Notify(460700)
					typ, _, slurpErr = readBuf.ReadTypedMsg(c.conn.Rd())
					if typ == pgwirebase.ClientMsgCopyDone || func() bool {
						__antithesis_instrumentation__.Notify(460703)
						return typ == pgwirebase.ClientMsgCopyFail == true
					}() == true {
						__antithesis_instrumentation__.Notify(460704)
						break
					} else {
						__antithesis_instrumentation__.Notify(460705)
					}
					__antithesis_instrumentation__.Notify(460701)
					if slurpErr != nil && func() bool {
						__antithesis_instrumentation__.Notify(460706)
						return !pgwirebase.IsMessageTooBigError(slurpErr) == true
					}() == true {
						__antithesis_instrumentation__.Notify(460707)
						return errors.CombineErrors(err, errors.Wrapf(slurpErr, "error slurping remaining bytes in COPY"))
					} else {
						__antithesis_instrumentation__.Notify(460708)
					}
					__antithesis_instrumentation__.Notify(460702)

					_, slurpErr = readBuf.SlurpBytes(c.conn.Rd(), pgwirebase.GetMessageTooBigSize(slurpErr))
					if slurpErr != nil {
						__antithesis_instrumentation__.Notify(460709)
						return errors.CombineErrors(err, errors.Wrapf(slurpErr, "error slurping remaining bytes in COPY"))
					} else {
						__antithesis_instrumentation__.Notify(460710)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(460711)
			}
			__antithesis_instrumentation__.Notify(460694)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460712)
		}
		__antithesis_instrumentation__.Notify(460692)

		switch typ {
		case pgwirebase.ClientMsgCopyData:
			__antithesis_instrumentation__.Notify(460713)
			if err := c.processCopyData(
				ctx, string(readBuf.Msg), false,
			); err != nil {
				__antithesis_instrumentation__.Notify(460719)
				return err
			} else {
				__antithesis_instrumentation__.Notify(460720)
			}
		case pgwirebase.ClientMsgCopyDone:
			__antithesis_instrumentation__.Notify(460714)
			if err := c.processCopyData(
				ctx, "", true,
			); err != nil {
				__antithesis_instrumentation__.Notify(460721)
				return err
			} else {
				__antithesis_instrumentation__.Notify(460722)
			}
			__antithesis_instrumentation__.Notify(460715)
			break Loop
		case pgwirebase.ClientMsgCopyFail:
			__antithesis_instrumentation__.Notify(460716)
			return errors.Newf("client canceled COPY")
		case pgwirebase.ClientMsgFlush, pgwirebase.ClientMsgSync:
			__antithesis_instrumentation__.Notify(460717)

		default:
			__antithesis_instrumentation__.Notify(460718)
			return pgwirebase.NewUnrecognizedMsgTypeErr(typ)
		}
	}
	__antithesis_instrumentation__.Notify(460681)

	dummy := tree.CopyFrom{}
	tag := []byte(dummy.StatementTag())
	tag = append(tag, ' ')
	tag = strconv.AppendInt(tag, int64(c.insertedRows), 10)
	return c.conn.SendCommandComplete(tag)
}

const (
	lineDelim = '\n'
	endOfData = `\.`
)

func (c *copyMachine) processCopyData(ctx context.Context, data string, final bool) (retErr error) {
	__antithesis_instrumentation__.Notify(460723)

	defer func() {
		__antithesis_instrumentation__.Notify(460729)
		if err := c.bufMemAcc.ResizeTo(ctx, int64(c.buf.Cap())); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(460730)
			return retErr == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(460731)
			retErr = err
		} else {
			__antithesis_instrumentation__.Notify(460732)
		}
	}()
	__antithesis_instrumentation__.Notify(460724)

	const copyBatchRowSize = 100

	if len(data) > (c.buf.Cap() - c.buf.Len()) {
		__antithesis_instrumentation__.Notify(460733)

		if err := c.bufMemAcc.ResizeTo(ctx, int64(len(data))); err != nil {
			__antithesis_instrumentation__.Notify(460734)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460735)
		}
	} else {
		__antithesis_instrumentation__.Notify(460736)
	}
	__antithesis_instrumentation__.Notify(460725)
	c.buf.WriteString(data)
	var readFn func(ctx context.Context, final bool) (brk bool, err error)
	switch c.format {
	case tree.CopyFormatText:
		__antithesis_instrumentation__.Notify(460737)
		readFn = c.readTextData
	case tree.CopyFormatBinary:
		__antithesis_instrumentation__.Notify(460738)
		readFn = c.readBinaryData
	case tree.CopyFormatCSV:
		__antithesis_instrumentation__.Notify(460739)
		readFn = c.readCSVData
	default:
		__antithesis_instrumentation__.Notify(460740)
		panic("unknown copy format")
	}
	__antithesis_instrumentation__.Notify(460726)
	for c.buf.Len() > 0 {
		__antithesis_instrumentation__.Notify(460741)
		brk, err := readFn(ctx, final)
		if err != nil {
			__antithesis_instrumentation__.Notify(460743)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460744)
		}
		__antithesis_instrumentation__.Notify(460742)
		if brk {
			__antithesis_instrumentation__.Notify(460745)
			break
		} else {
			__antithesis_instrumentation__.Notify(460746)
		}
	}
	__antithesis_instrumentation__.Notify(460727)

	if ln := len(c.rows); !final && func() bool {
		__antithesis_instrumentation__.Notify(460747)
		return (ln == 0 || func() bool {
			__antithesis_instrumentation__.Notify(460748)
			return ln < copyBatchRowSize == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(460749)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(460750)
	}
	__antithesis_instrumentation__.Notify(460728)
	return c.processRows(ctx)
}

func (c *copyMachine) readTextData(ctx context.Context, final bool) (brk bool, err error) {
	__antithesis_instrumentation__.Notify(460751)
	line, err := c.buf.ReadBytes(lineDelim)
	if err != nil {
		__antithesis_instrumentation__.Notify(460754)
		if err != io.EOF {
			__antithesis_instrumentation__.Notify(460755)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(460756)
			if !final {
				__antithesis_instrumentation__.Notify(460757)

				c.buf.Write(line)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(460758)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(460759)

		line = line[:len(line)-1]

		if len(line) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(460760)
			return line[len(line)-1] == '\r' == true
		}() == true {
			__antithesis_instrumentation__.Notify(460761)
			line = line[:len(line)-1]
		} else {
			__antithesis_instrumentation__.Notify(460762)
		}
	}
	__antithesis_instrumentation__.Notify(460752)
	if c.buf.Len() == 0 && func() bool {
		__antithesis_instrumentation__.Notify(460763)
		return bytes.Equal(line, []byte(`\.`)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(460764)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(460765)
	}
	__antithesis_instrumentation__.Notify(460753)
	err = c.readTextTuple(ctx, line)
	return false, err
}

func (c *copyMachine) readCSVData(ctx context.Context, final bool) (brk bool, err error) {
	__antithesis_instrumentation__.Notify(460766)
	var fullLine []byte
	quoteCharsSeen := 0

	for {
		__antithesis_instrumentation__.Notify(460770)
		line, err := c.buf.ReadBytes(lineDelim)
		fullLine = append(fullLine, line...)
		if err != nil {
			__antithesis_instrumentation__.Notify(460772)
			if err == io.EOF {
				__antithesis_instrumentation__.Notify(460773)
				if final {
					__antithesis_instrumentation__.Notify(460774)

					break
				} else {
					__antithesis_instrumentation__.Notify(460775)

					c.buf.Write(fullLine)
					return true, nil
				}
			} else {
				__antithesis_instrumentation__.Notify(460776)
				return false, err
			}
		} else {
			__antithesis_instrumentation__.Notify(460777)
		}
		__antithesis_instrumentation__.Notify(460771)

		quoteCharsSeen += bytes.Count(line, []byte{'"'})
		if quoteCharsSeen%2 == 0 {
			__antithesis_instrumentation__.Notify(460778)
			break
		} else {
			__antithesis_instrumentation__.Notify(460779)
		}
	}
	__antithesis_instrumentation__.Notify(460767)

	c.csvInput.Write(fullLine)
	record, err := c.csvReader.Read()

	if len(record) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(460780)
		return record[0] == endOfData == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(460781)
		return c.buf.Len() == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(460782)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(460783)
	}
	__antithesis_instrumentation__.Notify(460768)
	if err != nil {
		__antithesis_instrumentation__.Notify(460784)
		return false, pgerror.Wrap(err, pgcode.BadCopyFileFormat,
			"read CSV record")
	} else {
		__antithesis_instrumentation__.Notify(460785)
	}
	__antithesis_instrumentation__.Notify(460769)
	err = c.readCSVTuple(ctx, record)
	return false, err
}

func (c *copyMachine) maybeIgnoreHiddenColumnsStr(in []string) []string {
	__antithesis_instrumentation__.Notify(460786)
	if len(c.expectedHiddenColumnIdxs) == 0 {
		__antithesis_instrumentation__.Notify(460789)
		return in
	} else {
		__antithesis_instrumentation__.Notify(460790)
	}
	__antithesis_instrumentation__.Notify(460787)
	ret := in[:0]
	nextStartIdx := 0
	for _, toIdx := range c.expectedHiddenColumnIdxs {
		__antithesis_instrumentation__.Notify(460791)
		ret = append(ret, in[nextStartIdx:toIdx]...)
		nextStartIdx = toIdx + 1
	}
	__antithesis_instrumentation__.Notify(460788)
	ret = append(ret, in[nextStartIdx:]...)
	return ret
}

func (c *copyMachine) readCSVTuple(ctx context.Context, record []string) error {
	__antithesis_instrumentation__.Notify(460792)
	if expected := len(c.resultColumns) + len(c.expectedHiddenColumnIdxs); expected != len(record) {
		__antithesis_instrumentation__.Notify(460796)
		return pgerror.Newf(pgcode.BadCopyFileFormat,
			"expected %d values, got %d", expected, len(record))
	} else {
		__antithesis_instrumentation__.Notify(460797)
	}
	__antithesis_instrumentation__.Notify(460793)
	record = c.maybeIgnoreHiddenColumnsStr(record)
	exprs := make(tree.Exprs, len(record))
	for i, s := range record {
		__antithesis_instrumentation__.Notify(460798)
		if s == c.null {
			__antithesis_instrumentation__.Notify(460802)
			exprs[i] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(460803)
		}
		__antithesis_instrumentation__.Notify(460799)
		d, _, err := tree.ParseAndRequireString(c.resultColumns[i].Typ, s, c.parsingEvalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(460804)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460805)
		}
		__antithesis_instrumentation__.Notify(460800)

		sz := d.Size()
		if err := c.rowsMemAcc.Grow(ctx, int64(sz)); err != nil {
			__antithesis_instrumentation__.Notify(460806)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460807)
		}
		__antithesis_instrumentation__.Notify(460801)

		exprs[i] = d
	}
	__antithesis_instrumentation__.Notify(460794)
	if err := c.rowsMemAcc.Grow(ctx, int64(unsafe.Sizeof(exprs))); err != nil {
		__antithesis_instrumentation__.Notify(460808)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460809)
	}
	__antithesis_instrumentation__.Notify(460795)

	c.rows = append(c.rows, exprs)
	return nil
}

func (c *copyMachine) readBinaryData(ctx context.Context, final bool) (brk bool, err error) {
	__antithesis_instrumentation__.Notify(460810)
	if len(c.expectedHiddenColumnIdxs) > 0 {
		__antithesis_instrumentation__.Notify(460813)
		return false, pgerror.Newf(
			pgcode.FeatureNotSupported,
			"expect_and_ignore_not_visible_columns_in_copy not supported in binary mode",
		)
	} else {
		__antithesis_instrumentation__.Notify(460814)
	}
	__antithesis_instrumentation__.Notify(460811)
	switch c.binaryState {
	case binaryStateNeedSignature:
		__antithesis_instrumentation__.Notify(460815)
		if readSoFar, err := c.readBinarySignature(); err != nil {
			__antithesis_instrumentation__.Notify(460820)

			if !final && func() bool {
				__antithesis_instrumentation__.Notify(460822)
				return (err == io.EOF || func() bool {
					__antithesis_instrumentation__.Notify(460823)
					return err == io.ErrUnexpectedEOF == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(460824)
				c.buf.Write(readSoFar)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(460825)
			}
			__antithesis_instrumentation__.Notify(460821)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(460826)
		}
	case binaryStateRead:
		__antithesis_instrumentation__.Notify(460816)
		if readSoFar, err := c.readBinaryTuple(ctx); err != nil {
			__antithesis_instrumentation__.Notify(460827)

			if !final && func() bool {
				__antithesis_instrumentation__.Notify(460829)
				return (err == io.EOF || func() bool {
					__antithesis_instrumentation__.Notify(460830)
					return err == io.ErrUnexpectedEOF == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(460831)
				c.buf.Write(readSoFar)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(460832)
			}
			__antithesis_instrumentation__.Notify(460828)
			return false, errors.Wrapf(err, "read binary tuple")
		} else {
			__antithesis_instrumentation__.Notify(460833)
		}
	case binaryStateFoundTrailer:
		__antithesis_instrumentation__.Notify(460817)
		if !final {
			__antithesis_instrumentation__.Notify(460834)
			return false, pgerror.New(pgcode.BadCopyFileFormat,
				"copy data present after trailer")
		} else {
			__antithesis_instrumentation__.Notify(460835)
		}
		__antithesis_instrumentation__.Notify(460818)
		return true, nil
	default:
		__antithesis_instrumentation__.Notify(460819)
		panic("unknown binary state")
	}
	__antithesis_instrumentation__.Notify(460812)
	return false, nil
}

func (c *copyMachine) readBinaryTuple(ctx context.Context) (readSoFar []byte, err error) {
	__antithesis_instrumentation__.Notify(460836)
	var fieldCount int16
	var fieldCountBytes [2]byte
	n, err := io.ReadFull(&c.buf, fieldCountBytes[:])
	readSoFar = append(readSoFar, fieldCountBytes[:n]...)
	if err != nil {
		__antithesis_instrumentation__.Notify(460842)
		return readSoFar, err
	} else {
		__antithesis_instrumentation__.Notify(460843)
	}
	__antithesis_instrumentation__.Notify(460837)
	fieldCount = int16(binary.BigEndian.Uint16(fieldCountBytes[:]))
	if fieldCount == -1 {
		__antithesis_instrumentation__.Notify(460844)
		c.binaryState = binaryStateFoundTrailer
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(460845)
	}
	__antithesis_instrumentation__.Notify(460838)
	if fieldCount < 1 {
		__antithesis_instrumentation__.Notify(460846)
		return nil, pgerror.Newf(pgcode.BadCopyFileFormat,
			"unexpected field count: %d", fieldCount)
	} else {
		__antithesis_instrumentation__.Notify(460847)
	}
	__antithesis_instrumentation__.Notify(460839)
	exprs := make(tree.Exprs, fieldCount)
	var byteCount int32
	var byteCountBytes [4]byte
	for i := range exprs {
		__antithesis_instrumentation__.Notify(460848)
		n, err := io.ReadFull(&c.buf, byteCountBytes[:])
		readSoFar = append(readSoFar, byteCountBytes[:n]...)
		if err != nil {
			__antithesis_instrumentation__.Notify(460854)
			return readSoFar, err
		} else {
			__antithesis_instrumentation__.Notify(460855)
		}
		__antithesis_instrumentation__.Notify(460849)
		byteCount = int32(binary.BigEndian.Uint32(byteCountBytes[:]))
		if byteCount == -1 {
			__antithesis_instrumentation__.Notify(460856)
			exprs[i] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(460857)
		}
		__antithesis_instrumentation__.Notify(460850)
		data := make([]byte, byteCount)
		n, err = io.ReadFull(&c.buf, data)
		readSoFar = append(readSoFar, data[:n]...)
		if err != nil {
			__antithesis_instrumentation__.Notify(460858)
			return readSoFar, err
		} else {
			__antithesis_instrumentation__.Notify(460859)
		}
		__antithesis_instrumentation__.Notify(460851)
		d, err := pgwirebase.DecodeDatum(
			c.parsingEvalCtx,
			c.resultColumns[i].Typ,
			pgwirebase.FormatBinary,
			data,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(460860)
			return nil, pgerror.Wrapf(err, pgcode.BadCopyFileFormat,
				"decode datum as %s: %s", c.resultColumns[i].Typ.SQLString(), data)
		} else {
			__antithesis_instrumentation__.Notify(460861)
		}
		__antithesis_instrumentation__.Notify(460852)
		sz := d.Size()
		if err := c.rowsMemAcc.Grow(ctx, int64(sz)); err != nil {
			__antithesis_instrumentation__.Notify(460862)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(460863)
		}
		__antithesis_instrumentation__.Notify(460853)
		exprs[i] = d
	}
	__antithesis_instrumentation__.Notify(460840)
	if err = c.rowsMemAcc.Grow(ctx, int64(unsafe.Sizeof(exprs))); err != nil {
		__antithesis_instrumentation__.Notify(460864)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(460865)
	}
	__antithesis_instrumentation__.Notify(460841)
	c.rows = append(c.rows, exprs)
	return nil, nil
}

func (c *copyMachine) readBinarySignature() ([]byte, error) {
	__antithesis_instrumentation__.Notify(460866)

	const binarySignature = "PGCOPY\n\377\r\n\000" + "\x00\x00\x00\x00" + "\x00\x00\x00\x00"
	var sig [11 + 8]byte
	if n, err := io.ReadFull(&c.buf, sig[:]); err != nil {
		__antithesis_instrumentation__.Notify(460869)
		return sig[:n], err
	} else {
		__antithesis_instrumentation__.Notify(460870)
	}
	__antithesis_instrumentation__.Notify(460867)
	if !bytes.Equal(sig[:], []byte(binarySignature)) {
		__antithesis_instrumentation__.Notify(460871)
		return sig[:], pgerror.New(pgcode.BadCopyFileFormat,
			"unrecognized binary copy signature")
	} else {
		__antithesis_instrumentation__.Notify(460872)
	}
	__antithesis_instrumentation__.Notify(460868)
	c.binaryState = binaryStateRead
	return sig[:], nil
}

func (p *planner) preparePlannerForCopy(
	ctx context.Context, txnOpt copyTxnOpt,
) func(context.Context, error) error {
	__antithesis_instrumentation__.Notify(460873)
	txn := txnOpt.txn
	txnTs := txnOpt.txnTimestamp
	stmtTs := txnOpt.stmtTimestamp
	autoCommit := false
	if txn == nil {
		__antithesis_instrumentation__.Notify(460875)
		nodeID, _ := p.execCfg.NodeID.OptionalNodeID()

		txn = kv.NewTxnWithSteppingEnabled(ctx, p.execCfg.DB, nodeID, sessiondatapb.Normal)
		txnTs = p.execCfg.Clock.PhysicalTime()
		stmtTs = txnTs
		autoCommit = true
	} else {
		__antithesis_instrumentation__.Notify(460876)
	}
	__antithesis_instrumentation__.Notify(460874)
	txnOpt.resetPlanner(ctx, p, txn, txnTs, stmtTs)
	p.autoCommit = autoCommit

	return func(ctx context.Context, prevErr error) (err error) {
		__antithesis_instrumentation__.Notify(460877)

		if txnOpt.resetExtraTxnState != nil {
			__antithesis_instrumentation__.Notify(460880)
			defer func() {
				__antithesis_instrumentation__.Notify(460881)

				err = errors.CombineErrors(err, txnOpt.resetExtraTxnState(ctx))
			}()
		} else {
			__antithesis_instrumentation__.Notify(460882)
		}
		__antithesis_instrumentation__.Notify(460878)
		if prevErr == nil {
			__antithesis_instrumentation__.Notify(460883)

			if autoCommit && func() bool {
				__antithesis_instrumentation__.Notify(460885)
				return !txn.IsCommitted() == true
			}() == true {
				__antithesis_instrumentation__.Notify(460886)
				return txn.CommitOrCleanup(ctx)
			} else {
				__antithesis_instrumentation__.Notify(460887)
			}
			__antithesis_instrumentation__.Notify(460884)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(460888)
		}
		__antithesis_instrumentation__.Notify(460879)
		txn.CleanupOnError(ctx, prevErr)
		return prevErr
	}
}

func (c *copyMachine) insertRows(ctx context.Context) (retErr error) {
	__antithesis_instrumentation__.Notify(460889)
	if len(c.rows) == 0 {
		__antithesis_instrumentation__.Notify(460896)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(460897)
	}
	__antithesis_instrumentation__.Notify(460890)
	cleanup := c.p.preparePlannerForCopy(ctx, c.txnOpt)
	defer func() {
		__antithesis_instrumentation__.Notify(460898)
		retErr = cleanup(ctx, retErr)
	}()
	__antithesis_instrumentation__.Notify(460891)

	vc := &tree.ValuesClause{Rows: c.rows}
	numRows := len(c.rows)

	c.rows = c.rows[:0]
	c.rowsMemAcc.Clear(ctx)

	c.p.stmt = Statement{}
	c.p.stmt.AST = &tree.Insert{
		Table:   c.table,
		Columns: c.columns,
		Rows: &tree.Select{
			Select: vc,
		},
		Returning: tree.AbsentReturningClause,
	}
	if err := c.p.makeOptimizerPlan(ctx); err != nil {
		__antithesis_instrumentation__.Notify(460899)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460900)
	}
	__antithesis_instrumentation__.Notify(460892)

	var res streamingCommandResult
	err := c.execInsertPlan(ctx, &c.p, &res)
	if err != nil {
		__antithesis_instrumentation__.Notify(460901)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460902)
	}
	__antithesis_instrumentation__.Notify(460893)
	if err := res.Err(); err != nil {
		__antithesis_instrumentation__.Notify(460903)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460904)
	}
	__antithesis_instrumentation__.Notify(460894)

	if rows := res.RowsAffected(); rows != numRows {
		__antithesis_instrumentation__.Notify(460905)
		log.Fatalf(ctx, "didn't insert all buffered rows and yet no error was reported. "+
			"Inserted %d out of %d rows.", rows, numRows)
	} else {
		__antithesis_instrumentation__.Notify(460906)
	}
	__antithesis_instrumentation__.Notify(460895)
	c.insertedRows += numRows

	return nil
}

func (c *copyMachine) maybeIgnoreHiddenColumnsBytes(in [][]byte) [][]byte {
	__antithesis_instrumentation__.Notify(460907)
	if len(c.expectedHiddenColumnIdxs) == 0 {
		__antithesis_instrumentation__.Notify(460910)
		return in
	} else {
		__antithesis_instrumentation__.Notify(460911)
	}
	__antithesis_instrumentation__.Notify(460908)
	ret := in[:0]
	nextStartIdx := 0
	for _, toIdx := range c.expectedHiddenColumnIdxs {
		__antithesis_instrumentation__.Notify(460912)
		ret = append(ret, in[nextStartIdx:toIdx]...)
		nextStartIdx = toIdx + 1
	}
	__antithesis_instrumentation__.Notify(460909)
	ret = append(ret, in[nextStartIdx:]...)
	return ret
}

func (c *copyMachine) readTextTuple(ctx context.Context, line []byte) error {
	__antithesis_instrumentation__.Notify(460913)
	parts := bytes.Split(line, c.textDelim)
	if expected := len(c.resultColumns) + len(c.expectedHiddenColumnIdxs); expected != len(parts) {
		__antithesis_instrumentation__.Notify(460917)
		return pgerror.Newf(pgcode.BadCopyFileFormat,
			"expected %d values, got %d", expected, len(parts))
	} else {
		__antithesis_instrumentation__.Notify(460918)
	}
	__antithesis_instrumentation__.Notify(460914)
	parts = c.maybeIgnoreHiddenColumnsBytes(parts)
	exprs := make(tree.Exprs, len(parts))
	for i, part := range parts {
		__antithesis_instrumentation__.Notify(460919)
		s := string(part)

		if !c.forceNotNull && func() bool {
			__antithesis_instrumentation__.Notify(460925)
			return s == c.null == true
		}() == true {
			__antithesis_instrumentation__.Notify(460926)
			exprs[i] = tree.DNull
			continue
		} else {
			__antithesis_instrumentation__.Notify(460927)
		}
		__antithesis_instrumentation__.Notify(460920)
		switch t := c.resultColumns[i].Typ; t.Family() {
		case types.BytesFamily,
			types.DateFamily,
			types.IntervalFamily,
			types.INetFamily,
			types.StringFamily,
			types.TimestampFamily,
			types.TimestampTZFamily,
			types.UuidFamily:
			__antithesis_instrumentation__.Notify(460928)
			s = decodeCopy(s)
		default:
			__antithesis_instrumentation__.Notify(460929)
		}
		__antithesis_instrumentation__.Notify(460921)

		var d tree.Datum
		var err error
		switch c.resultColumns[i].Typ.Family() {
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(460930)
			d = tree.NewDBytes(tree.DBytes(s))
		default:
			__antithesis_instrumentation__.Notify(460931)
			d, _, err = tree.ParseAndRequireString(c.resultColumns[i].Typ, s, c.parsingEvalCtx)
		}
		__antithesis_instrumentation__.Notify(460922)

		if err != nil {
			__antithesis_instrumentation__.Notify(460932)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460933)
		}
		__antithesis_instrumentation__.Notify(460923)

		sz := d.Size()
		if err := c.rowsMemAcc.Grow(ctx, int64(sz)); err != nil {
			__antithesis_instrumentation__.Notify(460934)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460935)
		}
		__antithesis_instrumentation__.Notify(460924)

		exprs[i] = d
	}
	__antithesis_instrumentation__.Notify(460915)
	if err := c.rowsMemAcc.Grow(ctx, int64(unsafe.Sizeof(exprs))); err != nil {
		__antithesis_instrumentation__.Notify(460936)
		return err
	} else {
		__antithesis_instrumentation__.Notify(460937)
	}
	__antithesis_instrumentation__.Notify(460916)

	c.rows = append(c.rows, exprs)
	return nil
}

func decodeCopy(in string) string {
	__antithesis_instrumentation__.Notify(460938)
	var buf strings.Builder
	start := 0
	for i, n := 0, len(in); i < n; i++ {
		__antithesis_instrumentation__.Notify(460942)
		if in[i] != '\\' {
			__antithesis_instrumentation__.Notify(460945)
			continue
		} else {
			__antithesis_instrumentation__.Notify(460946)
		}
		__antithesis_instrumentation__.Notify(460943)
		buf.WriteString(in[start:i])
		i++

		if i >= n {
			__antithesis_instrumentation__.Notify(460947)

			buf.WriteByte('\\')
		} else {
			__antithesis_instrumentation__.Notify(460948)
			ch := in[i]
			if decodedChar := decodeMap[ch]; decodedChar != 0 {
				__antithesis_instrumentation__.Notify(460949)
				buf.WriteByte(decodedChar)
			} else {
				__antithesis_instrumentation__.Notify(460950)
				if ch == 'x' {
					__antithesis_instrumentation__.Notify(460951)

					if i+1 >= n {
						__antithesis_instrumentation__.Notify(460952)

						buf.WriteByte('x')
					} else {
						__antithesis_instrumentation__.Notify(460953)
						ch = in[i+1]
						digit, ok := decodeHexDigit(ch)
						if !ok {
							__antithesis_instrumentation__.Notify(460954)

							buf.WriteByte('x')
						} else {
							__antithesis_instrumentation__.Notify(460955)
							i++
							if i+1 < n {
								__antithesis_instrumentation__.Notify(460957)
								if v, ok := decodeHexDigit(in[i+1]); ok {
									__antithesis_instrumentation__.Notify(460958)
									i++
									digit <<= 4
									digit += v
								} else {
									__antithesis_instrumentation__.Notify(460959)
								}
							} else {
								__antithesis_instrumentation__.Notify(460960)
							}
							__antithesis_instrumentation__.Notify(460956)
							buf.WriteByte(digit)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(460961)
					if ch >= '0' && func() bool {
						__antithesis_instrumentation__.Notify(460962)
						return ch <= '7' == true
					}() == true {
						__antithesis_instrumentation__.Notify(460963)
						digit, _ := decodeOctDigit(ch)

						if i+1 < n {
							__antithesis_instrumentation__.Notify(460966)
							if v, ok := decodeOctDigit(in[i+1]); ok {
								__antithesis_instrumentation__.Notify(460967)
								i++
								digit <<= 3
								digit += v
							} else {
								__antithesis_instrumentation__.Notify(460968)
							}
						} else {
							__antithesis_instrumentation__.Notify(460969)
						}
						__antithesis_instrumentation__.Notify(460964)
						if i+1 < n {
							__antithesis_instrumentation__.Notify(460970)
							if v, ok := decodeOctDigit(in[i+1]); ok {
								__antithesis_instrumentation__.Notify(460971)
								i++
								digit <<= 3
								digit += v
							} else {
								__antithesis_instrumentation__.Notify(460972)
							}
						} else {
							__antithesis_instrumentation__.Notify(460973)
						}
						__antithesis_instrumentation__.Notify(460965)
						buf.WriteByte(digit)
					} else {
						__antithesis_instrumentation__.Notify(460974)

						buf.WriteByte(ch)
					}
				}
			}
		}
		__antithesis_instrumentation__.Notify(460944)
		start = i + 1
	}
	__antithesis_instrumentation__.Notify(460939)

	if start == 0 {
		__antithesis_instrumentation__.Notify(460975)
		return in
	} else {
		__antithesis_instrumentation__.Notify(460976)
	}
	__antithesis_instrumentation__.Notify(460940)
	if start < len(in) {
		__antithesis_instrumentation__.Notify(460977)
		buf.WriteString(in[start:])
	} else {
		__antithesis_instrumentation__.Notify(460978)
	}
	__antithesis_instrumentation__.Notify(460941)
	return buf.String()
}

func decodeDigit(c byte, onlyOctal bool) (byte, bool) {
	__antithesis_instrumentation__.Notify(460979)
	switch {
	case c >= '0' && func() bool {
		__antithesis_instrumentation__.Notify(460985)
		return c <= '7' == true
	}() == true:
		__antithesis_instrumentation__.Notify(460980)
		return c - '0', true
	case !onlyOctal && func() bool {
		__antithesis_instrumentation__.Notify(460986)
		return c >= '8' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(460987)
		return c <= '9' == true
	}() == true:
		__antithesis_instrumentation__.Notify(460981)
		return c - '0', true
	case !onlyOctal && func() bool {
		__antithesis_instrumentation__.Notify(460988)
		return c >= 'a' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(460989)
		return c <= 'f' == true
	}() == true:
		__antithesis_instrumentation__.Notify(460982)
		return c - 'a' + 10, true
	case !onlyOctal && func() bool {
		__antithesis_instrumentation__.Notify(460990)
		return c >= 'A' == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(460991)
		return c <= 'F' == true
	}() == true:
		__antithesis_instrumentation__.Notify(460983)
		return c - 'A' + 10, true
	default:
		__antithesis_instrumentation__.Notify(460984)
		return 0, false
	}
}

func decodeOctDigit(c byte) (byte, bool) {
	__antithesis_instrumentation__.Notify(460992)
	return decodeDigit(c, true)
}
func decodeHexDigit(c byte) (byte, bool) {
	__antithesis_instrumentation__.Notify(460993)
	return decodeDigit(c, false)
}

var decodeMap = map[byte]byte{
	'b':  '\b',
	'f':  '\f',
	'n':  '\n',
	'r':  '\r',
	't':  '\t',
	'v':  '\v',
	'\\': '\\',
}

type binaryState int

const (
	binaryStateNeedSignature binaryState = iota
	binaryStateRead
	binaryStateFoundTrailer
)
