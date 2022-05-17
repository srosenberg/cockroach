package clisqlexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding/csv"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/olekukonko/tablewriter"
)

type RowStrIter interface {
	Next() (row []string, err error)
	ToSlice() (allRows [][]string, err error)
	Align() []int
}

type rowSliceIter struct {
	allRows [][]string
	index   int
	align   []int
}

func (iter *rowSliceIter) Next() (row []string, err error) {
	__antithesis_instrumentation__.Notify(28983)
	if iter.index >= len(iter.allRows) {
		__antithesis_instrumentation__.Notify(28985)
		return nil, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(28986)
	}
	__antithesis_instrumentation__.Notify(28984)
	row = iter.allRows[iter.index]
	iter.index = iter.index + 1
	return row, nil
}

func (iter *rowSliceIter) ToSlice() ([][]string, error) {
	__antithesis_instrumentation__.Notify(28987)
	return iter.allRows, nil
}

func (iter *rowSliceIter) Align() []int {
	__antithesis_instrumentation__.Notify(28988)
	return iter.align
}

func convertAlign(align string) []int {
	__antithesis_instrumentation__.Notify(28989)
	result := make([]int, len(align))
	for i, v := range align {
		__antithesis_instrumentation__.Notify(28991)
		switch v {
		case 'l':
			__antithesis_instrumentation__.Notify(28992)
			result[i] = tablewriter.ALIGN_LEFT
		case 'r':
			__antithesis_instrumentation__.Notify(28993)
			result[i] = tablewriter.ALIGN_RIGHT
		case 'c':
			__antithesis_instrumentation__.Notify(28994)
			result[i] = tablewriter.ALIGN_CENTER
		default:
			__antithesis_instrumentation__.Notify(28995)
			result[i] = tablewriter.ALIGN_DEFAULT
		}
	}
	__antithesis_instrumentation__.Notify(28990)
	return result
}

func NewRowSliceIter(allRows [][]string, align string) RowStrIter {
	__antithesis_instrumentation__.Notify(28996)
	return &rowSliceIter{
		allRows: allRows,
		index:   0,
		align:   convertAlign(align),
	}
}

type rowIter struct {
	rows          clisqlclient.Rows
	colTypes      []string
	showMoreChars bool
}

func (iter *rowIter) Next() (row []string, err error) {
	__antithesis_instrumentation__.Notify(28997)
	nextRowString, err := getNextRowStrings(iter.rows, iter.colTypes, iter.showMoreChars)
	if err != nil {
		__antithesis_instrumentation__.Notify(29000)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(29001)
	}
	__antithesis_instrumentation__.Notify(28998)
	if nextRowString == nil {
		__antithesis_instrumentation__.Notify(29002)
		return nil, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(29003)
	}
	__antithesis_instrumentation__.Notify(28999)
	return nextRowString, nil
}

func (iter *rowIter) ToSlice() ([][]string, error) {
	__antithesis_instrumentation__.Notify(29004)
	return getAllRowStrings(iter.rows, iter.colTypes, iter.showMoreChars)
}

func (iter *rowIter) Align() []int {
	__antithesis_instrumentation__.Notify(29005)
	cols := iter.rows.Columns()
	align := make([]int, len(cols))
	for i := range align {
		__antithesis_instrumentation__.Notify(29007)
		switch iter.rows.ColumnTypeScanType(i).Kind() {
		case reflect.String:
			__antithesis_instrumentation__.Notify(29008)
			align[i] = tablewriter.ALIGN_LEFT
		case reflect.Slice:
			__antithesis_instrumentation__.Notify(29009)
			align[i] = tablewriter.ALIGN_LEFT
		case reflect.Int64:
			__antithesis_instrumentation__.Notify(29010)
			align[i] = tablewriter.ALIGN_RIGHT
		case reflect.Float64:
			__antithesis_instrumentation__.Notify(29011)
			align[i] = tablewriter.ALIGN_RIGHT
		case reflect.Bool:
			__antithesis_instrumentation__.Notify(29012)
			align[i] = tablewriter.ALIGN_CENTER
		default:
			__antithesis_instrumentation__.Notify(29013)
			align[i] = tablewriter.ALIGN_DEFAULT
		}
	}
	__antithesis_instrumentation__.Notify(29006)
	return align
}

func newRowIter(rows clisqlclient.Rows, showMoreChars bool) *rowIter {
	__antithesis_instrumentation__.Notify(29014)
	return &rowIter{
		rows:          rows,
		colTypes:      rows.ColumnTypeNames(),
		showMoreChars: showMoreChars,
	}
}

type rowReporter interface {
	describe(w io.Writer, cols []string) error
	beforeFirstRow(w io.Writer, allRows RowStrIter) error
	iter(w, ew io.Writer, rowIdx int, row []string) error
	doneRows(w io.Writer, seenRows int) error
	doneNoRows(w io.Writer) error
}

func render(
	r rowReporter,
	w, ew io.Writer,
	cols []string,
	iter RowStrIter,
	completedHook func(),
	noRowsHook func() (bool, error),
) (err error) {
	__antithesis_instrumentation__.Notify(29015)
	described := false
	nRows := 0
	defer func() {
		__antithesis_instrumentation__.Notify(29018)

		if !described {
			__antithesis_instrumentation__.Notify(29023)
			err = errors.WithSecondaryError(err, r.describe(w, cols))
		} else {
			__antithesis_instrumentation__.Notify(29024)
		}
		__antithesis_instrumentation__.Notify(29019)

		if completedHook != nil {
			__antithesis_instrumentation__.Notify(29025)
			completedHook()
		} else {
			__antithesis_instrumentation__.Notify(29026)
		}
		__antithesis_instrumentation__.Notify(29020)

		var handled bool
		if nRows == 0 && func() bool {
			__antithesis_instrumentation__.Notify(29027)
			return noRowsHook != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(29028)
			handled, err = noRowsHook()
			if err != nil {
				__antithesis_instrumentation__.Notify(29029)
				return
			} else {
				__antithesis_instrumentation__.Notify(29030)
			}
		} else {
			__antithesis_instrumentation__.Notify(29031)
		}
		__antithesis_instrumentation__.Notify(29021)
		if handled {
			__antithesis_instrumentation__.Notify(29032)
			err = errors.WithSecondaryError(err, r.doneNoRows(w))
		} else {
			__antithesis_instrumentation__.Notify(29033)
			err = errors.WithSecondaryError(err, r.doneRows(w, nRows))
		}
		__antithesis_instrumentation__.Notify(29022)

		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(29034)
			return nRows > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(29035)
			fmt.Fprintf(ew, "(error encountered after some results were delivered)\n")
		} else {
			__antithesis_instrumentation__.Notify(29036)
		}
	}()
	__antithesis_instrumentation__.Notify(29016)

	for {
		__antithesis_instrumentation__.Notify(29037)

		row, err := iter.Next()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(29042)

			break
		} else {
			__antithesis_instrumentation__.Notify(29043)
		}
		__antithesis_instrumentation__.Notify(29038)
		if err != nil {
			__antithesis_instrumentation__.Notify(29044)

			return err
		} else {
			__antithesis_instrumentation__.Notify(29045)
		}
		__antithesis_instrumentation__.Notify(29039)
		if nRows == 0 {
			__antithesis_instrumentation__.Notify(29046)

			described = true
			if err = r.describe(w, cols); err != nil {
				__antithesis_instrumentation__.Notify(29048)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29049)
			}
			__antithesis_instrumentation__.Notify(29047)
			if err = r.beforeFirstRow(w, iter); err != nil {
				__antithesis_instrumentation__.Notify(29050)
				return err
			} else {
				__antithesis_instrumentation__.Notify(29051)
			}
		} else {
			__antithesis_instrumentation__.Notify(29052)
		}
		__antithesis_instrumentation__.Notify(29040)

		if err = r.iter(w, ew, nRows, row); err != nil {
			__antithesis_instrumentation__.Notify(29053)
			return err
		} else {
			__antithesis_instrumentation__.Notify(29054)
		}
		__antithesis_instrumentation__.Notify(29041)

		nRows++
	}
	__antithesis_instrumentation__.Notify(29017)

	return nil
}

type asciiTableReporter struct {
	rows  int
	table *tablewriter.Table
	buf   bytes.Buffer
	w     *tabwriter.Writer

	tableBorderMode int
}

func newASCIITableReporter(tableBorderMode int) *asciiTableReporter {
	__antithesis_instrumentation__.Notify(29055)
	n := &asciiTableReporter{tableBorderMode: tableBorderMode}

	n.w = tabwriter.NewWriter(&n.buf, 4, 0, 1, ' ', 0)
	return n
}

func (p *asciiTableReporter) describe(w io.Writer, cols []string) error {
	__antithesis_instrumentation__.Notify(29056)
	if len(cols) > 0 {
		__antithesis_instrumentation__.Notify(29058)

		expandedCols := make([]string, len(cols))
		for i, c := range cols {
			__antithesis_instrumentation__.Notify(29061)
			p.buf.Reset()
			fmt.Fprint(p.w, c)
			_ = p.w.Flush()
			expandedCols[i] = p.buf.String()
		}
		__antithesis_instrumentation__.Notify(29059)

		p.table = tablewriter.NewWriter(w)
		p.table.SetAutoFormatHeaders(false)
		p.table.SetAutoWrapText(false)
		var outsideBorders, insideLines bool

		switch p.tableBorderMode {
		case 0:
			__antithesis_instrumentation__.Notify(29062)
			outsideBorders, insideLines = false, false
		case 1:
			__antithesis_instrumentation__.Notify(29063)
			outsideBorders, insideLines = false, true
		case 2:
			__antithesis_instrumentation__.Notify(29064)
			outsideBorders, insideLines = true, false
		case 3:
			__antithesis_instrumentation__.Notify(29065)
			outsideBorders, insideLines = true, true
		default:
			__antithesis_instrumentation__.Notify(29066)
		}
		__antithesis_instrumentation__.Notify(29060)
		p.table.SetBorder(outsideBorders)
		p.table.SetRowLine(insideLines)
		p.table.SetReflowDuringAutoWrap(false)
		p.table.SetHeader(expandedCols)
		p.table.SetTrimWhiteSpaceAtEOL(true)

		p.table.SetColWidth(72)
	} else {
		__antithesis_instrumentation__.Notify(29067)
	}
	__antithesis_instrumentation__.Notify(29057)
	return nil
}

func (p *asciiTableReporter) beforeFirstRow(w io.Writer, iter RowStrIter) error {
	__antithesis_instrumentation__.Notify(29068)
	if p.table == nil {
		__antithesis_instrumentation__.Notify(29070)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(29071)
	}
	__antithesis_instrumentation__.Notify(29069)

	p.table.SetColumnAlignment(iter.Align())
	return nil
}

const asciiTableWarnRows = 10000

func (p *asciiTableReporter) iter(_, ew io.Writer, _ int, row []string) error {
	__antithesis_instrumentation__.Notify(29072)
	if p.table == nil {
		__antithesis_instrumentation__.Notify(29076)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(29077)
	}
	__antithesis_instrumentation__.Notify(29073)

	if p.rows == asciiTableWarnRows {
		__antithesis_instrumentation__.Notify(29078)
		fmt.Fprintf(ew,
			"warning: buffering more than %d result rows in client "+
				"- RAM usage growing, consider another formatter instead\n",
			asciiTableWarnRows)
	} else {
		__antithesis_instrumentation__.Notify(29079)
	}
	__antithesis_instrumentation__.Notify(29074)

	for i, r := range row {
		__antithesis_instrumentation__.Notify(29080)
		p.buf.Reset()
		fmt.Fprint(p.w, r)
		_ = p.w.Flush()
		row[i] = p.buf.String()
	}
	__antithesis_instrumentation__.Notify(29075)
	p.table.Append(row)
	p.rows++
	return nil
}

func (p *asciiTableReporter) doneRows(w io.Writer, seenRows int) error {
	__antithesis_instrumentation__.Notify(29081)
	if p.table != nil {
		__antithesis_instrumentation__.Notify(29083)
		p.table.Render()
		p.table = nil
	} else {
		__antithesis_instrumentation__.Notify(29084)

		fmt.Fprintln(w, "--")
	}
	__antithesis_instrumentation__.Notify(29082)

	fmt.Fprintf(w, "(%d row%s)\n", seenRows, util.Pluralize(int64(seenRows)))
	return nil
}

func (p *asciiTableReporter) doneNoRows(_ io.Writer) error {
	__antithesis_instrumentation__.Notify(29085)
	p.table = nil
	return nil
}

type csvReporter struct {
	mu struct {
		syncutil.Mutex
		csvWriter *csv.Writer
	}
	stop chan struct{}
}

const csvFlushInterval = 5 * time.Second

func makeCSVReporter(w io.Writer, format TableDisplayFormat) (*csvReporter, func()) {
	__antithesis_instrumentation__.Notify(29086)
	r := &csvReporter{}
	r.mu.csvWriter = csv.NewWriter(w)
	if format == TableDisplayTSV {
		__antithesis_instrumentation__.Notify(29090)
		r.mu.csvWriter.Comma = '\t'
	} else {
		__antithesis_instrumentation__.Notify(29091)
	}
	__antithesis_instrumentation__.Notify(29087)

	r.stop = make(chan struct{}, 1)
	go func() {
		__antithesis_instrumentation__.Notify(29092)
		ticker := time.NewTicker(csvFlushInterval)
		for {
			__antithesis_instrumentation__.Notify(29093)
			select {
			case <-ticker.C:
				__antithesis_instrumentation__.Notify(29094)
				r.mu.Lock()
				r.mu.csvWriter.Flush()
				r.mu.Unlock()
			case <-r.stop:
				__antithesis_instrumentation__.Notify(29095)
				return
			}
		}
	}()
	__antithesis_instrumentation__.Notify(29088)
	cleanup := func() {
		__antithesis_instrumentation__.Notify(29096)
		close(r.stop)
	}
	__antithesis_instrumentation__.Notify(29089)
	return r, cleanup
}

func (p *csvReporter) describe(w io.Writer, cols []string) error {
	__antithesis_instrumentation__.Notify(29097)
	p.mu.Lock()
	if len(cols) == 0 {
		__antithesis_instrumentation__.Notify(29099)
		_ = p.mu.csvWriter.Write([]string{"# no columns"})
	} else {
		__antithesis_instrumentation__.Notify(29100)
		_ = p.mu.csvWriter.Write(cols)
	}
	__antithesis_instrumentation__.Notify(29098)
	p.mu.Unlock()
	return nil
}

func (p *csvReporter) iter(_, _ io.Writer, _ int, row []string) error {
	__antithesis_instrumentation__.Notify(29101)
	p.mu.Lock()
	if len(row) == 0 {
		__antithesis_instrumentation__.Notify(29103)
		_ = p.mu.csvWriter.Write([]string{"# empty"})
	} else {
		__antithesis_instrumentation__.Notify(29104)
		_ = p.mu.csvWriter.Write(row)
	}
	__antithesis_instrumentation__.Notify(29102)
	p.mu.Unlock()
	return nil
}

func (p *csvReporter) beforeFirstRow(_ io.Writer, _ RowStrIter) error {
	__antithesis_instrumentation__.Notify(29105)
	return nil
}
func (p *csvReporter) doneNoRows(_ io.Writer) error {
	__antithesis_instrumentation__.Notify(29106)
	return nil
}

func (p *csvReporter) doneRows(w io.Writer, seenRows int) error {
	__antithesis_instrumentation__.Notify(29107)
	p.mu.Lock()
	p.mu.csvWriter.Flush()
	p.mu.Unlock()
	return nil
}

type rawReporter struct{}

func (p *rawReporter) describe(w io.Writer, cols []string) error {
	__antithesis_instrumentation__.Notify(29108)
	fmt.Fprintf(w, "# %d column%s\n", len(cols),
		util.Pluralize(int64(len(cols))))
	return nil
}

func (p *rawReporter) iter(w, _ io.Writer, rowIdx int, row []string) error {
	__antithesis_instrumentation__.Notify(29109)
	fmt.Fprintf(w, "# row %d\n", rowIdx+1)
	for _, r := range row {
		__antithesis_instrumentation__.Notify(29111)
		fmt.Fprintf(w, "## %d\n%s\n", len(r), r)
	}
	__antithesis_instrumentation__.Notify(29110)
	return nil
}

func (p *rawReporter) beforeFirstRow(_ io.Writer, _ RowStrIter) error {
	__antithesis_instrumentation__.Notify(29112)
	return nil
}
func (p *rawReporter) doneNoRows(_ io.Writer) error {
	__antithesis_instrumentation__.Notify(29113)
	return nil
}

func (p *rawReporter) doneRows(w io.Writer, nRows int) error {
	__antithesis_instrumentation__.Notify(29114)
	fmt.Fprintf(w, "# %d row%s\n", nRows, util.Pluralize(int64(nRows)))
	return nil
}

type htmlReporter struct {
	nCols int

	escape   bool
	rowStats bool
}

func (p *htmlReporter) describe(w io.Writer, cols []string) error {
	__antithesis_instrumentation__.Notify(29115)
	p.nCols = len(cols)
	fmt.Fprint(w, "<table>\n<thead><tr>")
	if p.rowStats {
		__antithesis_instrumentation__.Notify(29118)
		fmt.Fprint(w, "<th>row</th>")
	} else {
		__antithesis_instrumentation__.Notify(29119)
	}
	__antithesis_instrumentation__.Notify(29116)
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(29120)
		if p.escape {
			__antithesis_instrumentation__.Notify(29122)
			col = html.EscapeString(col)
		} else {
			__antithesis_instrumentation__.Notify(29123)
		}
		__antithesis_instrumentation__.Notify(29121)
		fmt.Fprintf(w, "<th>%s</th>", strings.Replace(col, "\n", "<br/>", -1))
	}
	__antithesis_instrumentation__.Notify(29117)
	fmt.Fprintln(w, "</tr></thead>")
	return nil
}

func (p *htmlReporter) beforeFirstRow(w io.Writer, _ RowStrIter) error {
	__antithesis_instrumentation__.Notify(29124)
	fmt.Fprintln(w, "<tbody>")
	return nil
}

func (p *htmlReporter) iter(w, _ io.Writer, rowIdx int, row []string) error {
	__antithesis_instrumentation__.Notify(29125)
	fmt.Fprint(w, "<tr>")
	if p.rowStats {
		__antithesis_instrumentation__.Notify(29128)
		fmt.Fprintf(w, "<td>%d</td>", rowIdx+1)
	} else {
		__antithesis_instrumentation__.Notify(29129)
	}
	__antithesis_instrumentation__.Notify(29126)
	for _, r := range row {
		__antithesis_instrumentation__.Notify(29130)
		if p.escape {
			__antithesis_instrumentation__.Notify(29132)
			r = html.EscapeString(r)
		} else {
			__antithesis_instrumentation__.Notify(29133)
		}
		__antithesis_instrumentation__.Notify(29131)
		fmt.Fprintf(w, "<td>%s</td>", strings.Replace(r, "\n", "<br/>", -1))
	}
	__antithesis_instrumentation__.Notify(29127)
	fmt.Fprintln(w, "</tr>")
	return nil
}

func (p *htmlReporter) doneNoRows(w io.Writer) error {
	__antithesis_instrumentation__.Notify(29134)
	fmt.Fprintln(w, "</table>")
	return nil
}

func (p *htmlReporter) doneRows(w io.Writer, nRows int) error {
	__antithesis_instrumentation__.Notify(29135)
	fmt.Fprintln(w, "</tbody>")
	if p.rowStats {
		__antithesis_instrumentation__.Notify(29137)
		fmt.Fprintf(w, "<tfoot><tr><td colspan=%d>%d row%s</td></tr></tfoot>",
			p.nCols+1, nRows, util.Pluralize(int64(nRows)))
	} else {
		__antithesis_instrumentation__.Notify(29138)
	}
	__antithesis_instrumentation__.Notify(29136)
	fmt.Fprintln(w, "</table>")
	return nil
}

type recordReporter struct {
	cols        []string
	colRest     [][]string
	maxColWidth int
}

func (p *recordReporter) describe(w io.Writer, cols []string) error {
	__antithesis_instrumentation__.Notify(29139)
	for _, col := range cols {
		__antithesis_instrumentation__.Notify(29141)
		parts := strings.Split(col, "\n")
		for _, part := range parts {
			__antithesis_instrumentation__.Notify(29143)
			colLen := utf8.RuneCountInString(part)
			if colLen > p.maxColWidth {
				__antithesis_instrumentation__.Notify(29144)
				p.maxColWidth = colLen
			} else {
				__antithesis_instrumentation__.Notify(29145)
			}
		}
		__antithesis_instrumentation__.Notify(29142)
		p.cols = append(p.cols, parts[0])
		p.colRest = append(p.colRest, parts[1:])
	}
	__antithesis_instrumentation__.Notify(29140)
	return nil
}

func (p *recordReporter) iter(w, _ io.Writer, rowIdx int, row []string) error {
	__antithesis_instrumentation__.Notify(29146)
	if len(p.cols) == 0 {
		__antithesis_instrumentation__.Notify(29149)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(29150)
	}
	__antithesis_instrumentation__.Notify(29147)
	fmt.Fprintf(w, "-[ RECORD %d ]\n", rowIdx+1)
	for j, r := range row {
		__antithesis_instrumentation__.Notify(29151)
		lines := strings.Split(r, "\n")
		for l, line := range lines {
			__antithesis_instrumentation__.Notify(29152)
			colLabel := p.cols[j]
			if l > 0 {
				__antithesis_instrumentation__.Notify(29156)
				colLabel = ""
			} else {
				__antithesis_instrumentation__.Notify(29157)
			}
			__antithesis_instrumentation__.Notify(29153)
			lineCont := "+"
			if l == len(lines)-1 {
				__antithesis_instrumentation__.Notify(29158)
				lineCont = ""
			} else {
				__antithesis_instrumentation__.Notify(29159)
			}
			__antithesis_instrumentation__.Notify(29154)

			contChar := " "
			if len(p.colRest[j]) > 0 {
				__antithesis_instrumentation__.Notify(29160)
				contChar = "+"
			} else {
				__antithesis_instrumentation__.Notify(29161)
			}
			__antithesis_instrumentation__.Notify(29155)
			fmt.Fprintf(w, "%-*s%s| %s%s\n", p.maxColWidth, colLabel, contChar, line, lineCont)
			for k, cont := range p.colRest[j] {
				__antithesis_instrumentation__.Notify(29162)
				if k == len(p.colRest[j])-1 {
					__antithesis_instrumentation__.Notify(29164)
					contChar = " "
				} else {
					__antithesis_instrumentation__.Notify(29165)
				}
				__antithesis_instrumentation__.Notify(29163)
				fmt.Fprintf(w, "%-*s%s|\n", p.maxColWidth, cont, contChar)
			}
		}
	}
	__antithesis_instrumentation__.Notify(29148)
	return nil
}

func (p *recordReporter) doneRows(w io.Writer, seenRows int) error {
	__antithesis_instrumentation__.Notify(29166)
	if len(p.cols) == 0 {
		__antithesis_instrumentation__.Notify(29168)
		fmt.Fprintf(w, "(%d row%s)\n", seenRows, util.Pluralize(int64(seenRows)))
	} else {
		__antithesis_instrumentation__.Notify(29169)
	}
	__antithesis_instrumentation__.Notify(29167)
	return nil
}

func (p *recordReporter) beforeFirstRow(_ io.Writer, _ RowStrIter) error {
	__antithesis_instrumentation__.Notify(29170)
	return nil
}
func (p *recordReporter) doneNoRows(_ io.Writer) error {
	__antithesis_instrumentation__.Notify(29171)
	return nil
}

type sqlReporter struct {
	noColumns bool
}

func (p *sqlReporter) describe(w io.Writer, cols []string) error {
	__antithesis_instrumentation__.Notify(29172)
	fmt.Fprint(w, "CREATE TABLE results (\n")
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(29174)
		var colName bytes.Buffer
		lexbase.EncodeRestrictedSQLIdent(&colName, col, lexbase.EncNoFlags)
		fmt.Fprintf(w, "  %s STRING", colName.String())
		if i < len(cols)-1 {
			__antithesis_instrumentation__.Notify(29176)
			fmt.Fprint(w, ",")
		} else {
			__antithesis_instrumentation__.Notify(29177)
		}
		__antithesis_instrumentation__.Notify(29175)
		fmt.Fprint(w, "\n")
	}
	__antithesis_instrumentation__.Notify(29173)
	fmt.Fprint(w, ");\n\n")
	p.noColumns = (len(cols) == 0)
	return nil
}

func (p *sqlReporter) iter(w, _ io.Writer, _ int, row []string) error {
	__antithesis_instrumentation__.Notify(29178)
	if p.noColumns {
		__antithesis_instrumentation__.Notify(29181)
		fmt.Fprintln(w, "INSERT INTO results(rowid) VALUES (DEFAULT);")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(29182)
	}
	__antithesis_instrumentation__.Notify(29179)

	fmt.Fprint(w, "INSERT INTO results VALUES (")
	for i, r := range row {
		__antithesis_instrumentation__.Notify(29183)
		var buf bytes.Buffer
		lexbase.EncodeSQLStringWithFlags(&buf, r, lexbase.EncNoFlags)
		fmt.Fprint(w, buf.String())
		if i < len(row)-1 {
			__antithesis_instrumentation__.Notify(29184)
			fmt.Fprint(w, ", ")
		} else {
			__antithesis_instrumentation__.Notify(29185)
		}
	}
	__antithesis_instrumentation__.Notify(29180)
	fmt.Fprint(w, ");\n")
	return nil
}

func (p *sqlReporter) beforeFirstRow(_ io.Writer, _ RowStrIter) error {
	__antithesis_instrumentation__.Notify(29186)
	return nil
}
func (p *sqlReporter) doneNoRows(_ io.Writer) error {
	__antithesis_instrumentation__.Notify(29187)
	return nil
}
func (p *sqlReporter) doneRows(w io.Writer, seenRows int) error {
	__antithesis_instrumentation__.Notify(29188)
	fmt.Fprintf(w, "-- %d row%s\n", seenRows, util.Pluralize(int64(seenRows)))
	return nil
}

func (sqlExecCtx *Context) makeReporter(w io.Writer) (rowReporter, func(), error) {
	__antithesis_instrumentation__.Notify(29189)
	switch sqlExecCtx.TableDisplayFormat {
	case TableDisplayTable:
		__antithesis_instrumentation__.Notify(29190)
		return newASCIITableReporter(sqlExecCtx.TableBorderMode), nil, nil

	case TableDisplayTSV:
		__antithesis_instrumentation__.Notify(29191)
		fallthrough
	case TableDisplayCSV:
		__antithesis_instrumentation__.Notify(29192)
		reporter, cleanup := makeCSVReporter(w, sqlExecCtx.TableDisplayFormat)
		return reporter, cleanup, nil

	case TableDisplayRaw:
		__antithesis_instrumentation__.Notify(29193)
		return &rawReporter{}, nil, nil

	case TableDisplayHTML:
		__antithesis_instrumentation__.Notify(29194)
		return &htmlReporter{escape: true, rowStats: true}, nil, nil

	case TableDisplayRawHTML:
		__antithesis_instrumentation__.Notify(29195)
		return &htmlReporter{escape: false, rowStats: false}, nil, nil

	case TableDisplayRecords:
		__antithesis_instrumentation__.Notify(29196)
		return &recordReporter{}, nil, nil

	case TableDisplaySQL:
		__antithesis_instrumentation__.Notify(29197)
		return &sqlReporter{}, nil, nil

	default:
		__antithesis_instrumentation__.Notify(29198)
		return nil, nil, errors.Errorf("unhandled display format: %d", sqlExecCtx.TableDisplayFormat)
	}
}

func (sqlExecCtx *Context) PrintQueryOutput(
	w, ew io.Writer, cols []string, allRows RowStrIter,
) error {
	__antithesis_instrumentation__.Notify(29199)
	reporter, cleanup, err := sqlExecCtx.makeReporter(w)
	if err != nil {
		__antithesis_instrumentation__.Notify(29202)
		return err
	} else {
		__antithesis_instrumentation__.Notify(29203)
	}
	__antithesis_instrumentation__.Notify(29200)
	if cleanup != nil {
		__antithesis_instrumentation__.Notify(29204)
		defer cleanup()
	} else {
		__antithesis_instrumentation__.Notify(29205)
	}
	__antithesis_instrumentation__.Notify(29201)
	return render(reporter, w, ew, cols, allRows, nil, nil)
}
