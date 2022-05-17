package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catalogkeys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
	"github.com/dustin/go-humanize"
)

type DiagramFlags struct {
	ShowInputTypes bool

	MakeDeterministic bool
}

type diagramCellType interface {
	summary() (title string, details []string)
}

func (ord *Ordering) diagramString() string {
	__antithesis_instrumentation__.Notify(478038)
	var buf bytes.Buffer
	for i, c := range ord.Columns {
		__antithesis_instrumentation__.Notify(478040)
		if i > 0 {
			__antithesis_instrumentation__.Notify(478042)
			buf.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(478043)
		}
		__antithesis_instrumentation__.Notify(478041)
		fmt.Fprintf(&buf, "@%d", c.ColIdx+1)
		if c.Direction == Ordering_Column_DESC {
			__antithesis_instrumentation__.Notify(478044)
			buf.WriteByte('-')
		} else {
			__antithesis_instrumentation__.Notify(478045)
			buf.WriteByte('+')
		}
	}
	__antithesis_instrumentation__.Notify(478039)
	return buf.String()
}

func colListStr(cols []uint32) string {
	__antithesis_instrumentation__.Notify(478046)
	var buf bytes.Buffer
	for i, c := range cols {
		__antithesis_instrumentation__.Notify(478048)
		if i > 0 {
			__antithesis_instrumentation__.Notify(478050)
			buf.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(478051)
		}
		__antithesis_instrumentation__.Notify(478049)
		fmt.Fprintf(&buf, "@%d", c+1)
	}
	__antithesis_instrumentation__.Notify(478047)
	return buf.String()
}

func (*NoopCoreSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478052)
	return "No-op", []string{}
}

func (f *FiltererSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478053)
	return "Filterer", []string{
		fmt.Sprintf("Filter: %s", f.Filter),
	}
}

func (mts *MetadataTestSenderSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478054)
	return "MetadataTestSender", []string{mts.ID}
}

func (*MetadataTestReceiverSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478055)
	return "MetadataTestReceiver", []string{}
}

func (v *ValuesCoreSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478056)
	var bytes uint64
	for _, b := range v.RawBytes {
		__antithesis_instrumentation__.Notify(478058)
		bytes += uint64(len(b))
	}
	__antithesis_instrumentation__.Notify(478057)
	detail := fmt.Sprintf("%s (%d chunks)", humanize.IBytes(bytes), len(v.RawBytes))
	return "Values", []string{detail}
}

func (a *AggregatorSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478059)
	details := make([]string, 0, len(a.Aggregations)+1)
	if len(a.GroupCols) > 0 {
		__antithesis_instrumentation__.Notify(478063)
		details = append(details, colListStr(a.GroupCols))
	} else {
		__antithesis_instrumentation__.Notify(478064)
	}
	__antithesis_instrumentation__.Notify(478060)
	if len(a.OrderedGroupCols) > 0 {
		__antithesis_instrumentation__.Notify(478065)
		details = append(details, fmt.Sprintf("Ordered: %s", colListStr(a.OrderedGroupCols)))
	} else {
		__antithesis_instrumentation__.Notify(478066)
	}
	__antithesis_instrumentation__.Notify(478061)
	for _, agg := range a.Aggregations {
		__antithesis_instrumentation__.Notify(478067)
		var buf bytes.Buffer
		buf.WriteString(agg.Func.String())
		buf.WriteByte('(')

		if agg.Distinct {
			__antithesis_instrumentation__.Notify(478070)
			buf.WriteString("DISTINCT ")
		} else {
			__antithesis_instrumentation__.Notify(478071)
		}
		__antithesis_instrumentation__.Notify(478068)
		buf.WriteString(colListStr(agg.ColIdx))
		buf.WriteByte(')')
		if agg.FilterColIdx != nil {
			__antithesis_instrumentation__.Notify(478072)
			fmt.Fprintf(&buf, " FILTER @%d", *agg.FilterColIdx+1)
		} else {
			__antithesis_instrumentation__.Notify(478073)
		}
		__antithesis_instrumentation__.Notify(478069)

		details = append(details, buf.String())
	}
	__antithesis_instrumentation__.Notify(478062)

	return "Aggregator", details
}

func indexDetail(desc *descpb.TableDescriptor, indexIdx uint32) string {
	__antithesis_instrumentation__.Notify(478074)
	var index string
	if indexIdx > 0 {
		__antithesis_instrumentation__.Notify(478076)
		index = desc.Indexes[indexIdx-1].Name
	} else {
		__antithesis_instrumentation__.Notify(478077)
		index = desc.PrimaryIndex.Name
	}
	__antithesis_instrumentation__.Notify(478075)
	return fmt.Sprintf("%s@%s", desc.Name, index)
}

func appendColumns(details []string, columns []descpb.IndexFetchSpec_Column) []string {
	__antithesis_instrumentation__.Notify(478078)
	var b strings.Builder
	b.WriteString("Columns:")
	const wrapAt = 100
	for i := range columns {
		__antithesis_instrumentation__.Notify(478080)
		if i > 0 {
			__antithesis_instrumentation__.Notify(478083)
			b.WriteByte(',')
		} else {
			__antithesis_instrumentation__.Notify(478084)
		}
		__antithesis_instrumentation__.Notify(478081)
		name := columns[i].Name
		if b.Len()+len(name)+1 > wrapAt {
			__antithesis_instrumentation__.Notify(478085)
			details = append(details, b.String())
			b.Reset()
		} else {
			__antithesis_instrumentation__.Notify(478086)
		}
		__antithesis_instrumentation__.Notify(478082)
		b.WriteByte(' ')
		b.WriteString(name)
	}
	__antithesis_instrumentation__.Notify(478079)
	details = append(details, b.String())
	return details
}

func (tr *TableReaderSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478087)
	details := make([]string, 0, 3)
	details = append(details, fmt.Sprintf("%s@%s", tr.FetchSpec.TableName, tr.FetchSpec.IndexName))
	details = appendColumns(details, tr.FetchSpec.FetchedColumns)

	if len(tr.Spans) > 0 {
		__antithesis_instrumentation__.Notify(478089)

		keyDirs := make([]encoding.Direction, len(tr.FetchSpec.KeyAndSuffixColumns))
		for i := range keyDirs {
			__antithesis_instrumentation__.Notify(478093)
			keyDirs[i] = encoding.Ascending
			if tr.FetchSpec.KeyAndSuffixColumns[i].Direction == descpb.IndexDescriptor_DESC {
				__antithesis_instrumentation__.Notify(478094)
				keyDirs[i] = encoding.Descending
			} else {
				__antithesis_instrumentation__.Notify(478095)
			}
		}
		__antithesis_instrumentation__.Notify(478090)

		var spanStr strings.Builder
		spanStr.WriteString("Spans: ")
		spanStr.WriteString(catalogkeys.PrettySpan(keyDirs, tr.Spans[0], 2))

		if len(tr.Spans) > 1 {
			__antithesis_instrumentation__.Notify(478096)
			spanStr.WriteString(fmt.Sprintf(" and %d other", len(tr.Spans)-1))
		} else {
			__antithesis_instrumentation__.Notify(478097)
		}
		__antithesis_instrumentation__.Notify(478091)

		if len(tr.Spans) > 2 {
			__antithesis_instrumentation__.Notify(478098)
			spanStr.WriteString("s")
		} else {
			__antithesis_instrumentation__.Notify(478099)
		}
		__antithesis_instrumentation__.Notify(478092)

		details = append(details, spanStr.String())
	} else {
		__antithesis_instrumentation__.Notify(478100)
	}
	__antithesis_instrumentation__.Notify(478088)

	return "TableReader", details
}

func (jr *JoinReaderSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478101)
	details := make([]string, 0, 5)
	if jr.Type != descpb.InnerJoin {
		__antithesis_instrumentation__.Notify(478109)
		details = append(details, joinTypeDetail(jr.Type))
	} else {
		__antithesis_instrumentation__.Notify(478110)
	}
	__antithesis_instrumentation__.Notify(478102)
	details = append(details, fmt.Sprintf("%s@%s", jr.FetchSpec.TableName, jr.FetchSpec.IndexName))
	if len(jr.LookupColumns) > 0 {
		__antithesis_instrumentation__.Notify(478111)
		details = append(details, fmt.Sprintf("Lookup join on: %s", colListStr(jr.LookupColumns)))
	} else {
		__antithesis_instrumentation__.Notify(478112)
	}
	__antithesis_instrumentation__.Notify(478103)
	if !jr.LookupExpr.Empty() {
		__antithesis_instrumentation__.Notify(478113)
		details = append(details, fmt.Sprintf("Lookup join on: %s", jr.LookupExpr))
	} else {
		__antithesis_instrumentation__.Notify(478114)
	}
	__antithesis_instrumentation__.Notify(478104)
	if !jr.RemoteLookupExpr.Empty() {
		__antithesis_instrumentation__.Notify(478115)
		details = append(details, fmt.Sprintf("Remote lookup join on: %s", jr.RemoteLookupExpr))
	} else {
		__antithesis_instrumentation__.Notify(478116)
	}
	__antithesis_instrumentation__.Notify(478105)
	if !jr.OnExpr.Empty() {
		__antithesis_instrumentation__.Notify(478117)
		details = append(details, fmt.Sprintf("ON %s", jr.OnExpr))
	} else {
		__antithesis_instrumentation__.Notify(478118)
	}
	__antithesis_instrumentation__.Notify(478106)
	if jr.LeftJoinWithPairedJoiner {
		__antithesis_instrumentation__.Notify(478119)
		details = append(details, "second join in paired-join")
	} else {
		__antithesis_instrumentation__.Notify(478120)
	}
	__antithesis_instrumentation__.Notify(478107)
	if jr.OutputGroupContinuationForLeftRow {
		__antithesis_instrumentation__.Notify(478121)
		details = append(details, "first join in paired-join")
	} else {
		__antithesis_instrumentation__.Notify(478122)
	}
	__antithesis_instrumentation__.Notify(478108)
	details = appendColumns(details, jr.FetchSpec.FetchedColumns)
	return "JoinReader", details
}

func joinTypeDetail(joinType descpb.JoinType) string {
	__antithesis_instrumentation__.Notify(478123)
	typeStr := strings.Replace(joinType.String(), "_", " ", -1)
	if joinType == descpb.IntersectAllJoin || func() bool {
		__antithesis_instrumentation__.Notify(478125)
		return joinType == descpb.ExceptAllJoin == true
	}() == true {
		__antithesis_instrumentation__.Notify(478126)
		return fmt.Sprintf("Type: %s", typeStr)
	} else {
		__antithesis_instrumentation__.Notify(478127)
	}
	__antithesis_instrumentation__.Notify(478124)
	return fmt.Sprintf("Type: %s JOIN", typeStr)
}

func (hj *HashJoinerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478128)
	name := "HashJoiner"
	if hj.Type.IsSetOpJoin() {
		__antithesis_instrumentation__.Notify(478133)
		name = "HashSetOp"
	} else {
		__antithesis_instrumentation__.Notify(478134)
	}
	__antithesis_instrumentation__.Notify(478129)

	details := make([]string, 0, 4)

	if hj.Type != descpb.InnerJoin {
		__antithesis_instrumentation__.Notify(478135)
		details = append(details, joinTypeDetail(hj.Type))
	} else {
		__antithesis_instrumentation__.Notify(478136)
	}
	__antithesis_instrumentation__.Notify(478130)
	if len(hj.LeftEqColumns) > 0 {
		__antithesis_instrumentation__.Notify(478137)
		details = append(details, fmt.Sprintf(
			"left(%s)=right(%s)",
			colListStr(hj.LeftEqColumns), colListStr(hj.RightEqColumns),
		))
	} else {
		__antithesis_instrumentation__.Notify(478138)
	}
	__antithesis_instrumentation__.Notify(478131)
	if !hj.OnExpr.Empty() {
		__antithesis_instrumentation__.Notify(478139)
		details = append(details, fmt.Sprintf("ON %s", hj.OnExpr))
	} else {
		__antithesis_instrumentation__.Notify(478140)
	}
	__antithesis_instrumentation__.Notify(478132)

	return name, details
}

func (ifs *InvertedFiltererSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478141)
	name := "InvertedFilterer"
	var b strings.Builder
	for i := range ifs.InvertedExpr.SpansToRead {
		__antithesis_instrumentation__.Notify(478143)
		if i > 0 {
			__antithesis_instrumentation__.Notify(478144)
			fmt.Fprintf(&b, " and %d others", len(ifs.InvertedExpr.SpansToRead)-1)
			break
		} else {
			__antithesis_instrumentation__.Notify(478145)
			fmt.Fprintf(&b, "%s", ifs.InvertedExpr.SpansToRead[i].String())
		}
	}
	__antithesis_instrumentation__.Notify(478142)
	details := append([]string(nil), fmt.Sprintf(
		"InvertedExpr on @%d: spans %s", ifs.InvertedColIdx, b.String()))
	return name, details
}

func orderedJoinDetails(
	joinType descpb.JoinType, left, right Ordering, onExpr Expression,
) []string {
	__antithesis_instrumentation__.Notify(478146)
	details := make([]string, 0, 3)

	if joinType != descpb.InnerJoin {
		__antithesis_instrumentation__.Notify(478149)
		details = append(details, joinTypeDetail(joinType))
	} else {
		__antithesis_instrumentation__.Notify(478150)
	}
	__antithesis_instrumentation__.Notify(478147)
	details = append(details, fmt.Sprintf(
		"left(%s)=right(%s)", left.diagramString(), right.diagramString(),
	))

	if !onExpr.Empty() {
		__antithesis_instrumentation__.Notify(478151)
		details = append(details, fmt.Sprintf("ON %s", onExpr))
	} else {
		__antithesis_instrumentation__.Notify(478152)
	}
	__antithesis_instrumentation__.Notify(478148)

	return details
}

func (mj *MergeJoinerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478153)
	name := "MergeJoiner"
	if mj.Type.IsSetOpJoin() {
		__antithesis_instrumentation__.Notify(478155)
		name = "MergeSetOp"
	} else {
		__antithesis_instrumentation__.Notify(478156)
	}
	__antithesis_instrumentation__.Notify(478154)
	return name, orderedJoinDetails(mj.Type, mj.LeftOrdering, mj.RightOrdering, mj.OnExpr)
}

func (zj *ZigzagJoinerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478157)
	name := "ZigzagJoiner"
	tables := zj.Tables
	details := make([]string, 0, len(tables)+1)
	for i, table := range tables {
		__antithesis_instrumentation__.Notify(478160)
		details = append(details, fmt.Sprintf(
			"Side %d: %s", i, indexDetail(&table, zj.IndexOrdinals[i]),
		))
	}
	__antithesis_instrumentation__.Notify(478158)
	if !zj.OnExpr.Empty() {
		__antithesis_instrumentation__.Notify(478161)
		details = append(details, fmt.Sprintf("ON %s", zj.OnExpr))
	} else {
		__antithesis_instrumentation__.Notify(478162)
	}
	__antithesis_instrumentation__.Notify(478159)
	return name, details
}

func (ij *InvertedJoinerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478163)
	details := make([]string, 0, 5)
	if ij.Type != descpb.InnerJoin {
		__antithesis_instrumentation__.Notify(478167)
		details = append(details, joinTypeDetail(ij.Type))
	} else {
		__antithesis_instrumentation__.Notify(478168)
	}
	__antithesis_instrumentation__.Notify(478164)
	details = append(details, indexDetail(&ij.Table, ij.IndexIdx))
	details = append(details, fmt.Sprintf("InvertedExpr %s", ij.InvertedExpr))
	if !ij.OnExpr.Empty() {
		__antithesis_instrumentation__.Notify(478169)
		details = append(details, fmt.Sprintf("ON %s", ij.OnExpr))
	} else {
		__antithesis_instrumentation__.Notify(478170)
	}
	__antithesis_instrumentation__.Notify(478165)
	if ij.OutputGroupContinuationForLeftRow {
		__antithesis_instrumentation__.Notify(478171)
		details = append(details, "first join in paired-join")
	} else {
		__antithesis_instrumentation__.Notify(478172)
	}
	__antithesis_instrumentation__.Notify(478166)
	return "InvertedJoiner", details
}

func (s *SorterSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478173)
	details := []string{s.OutputOrdering.diagramString()}
	if s.OrderingMatchLen != 0 {
		__antithesis_instrumentation__.Notify(478176)
		details = append(details, fmt.Sprintf("match len: %d", s.OrderingMatchLen))
	} else {
		__antithesis_instrumentation__.Notify(478177)
	}
	__antithesis_instrumentation__.Notify(478174)
	if s.Limit > 0 {
		__antithesis_instrumentation__.Notify(478178)
		details = append(details, fmt.Sprintf("TopK: %d", s.Limit))
	} else {
		__antithesis_instrumentation__.Notify(478179)
	}
	__antithesis_instrumentation__.Notify(478175)
	return "Sorter", details
}

func (bf *BackfillerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478180)
	details := []string{
		bf.Table.Name,
		fmt.Sprintf("Type: %s", bf.Type.String()),
	}
	return "Backfiller", details
}

func (m *BackupDataSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478181)
	var spanStr strings.Builder
	if len(m.Spans) > 0 {
		__antithesis_instrumentation__.Notify(478183)
		spanStr.WriteString(fmt.Sprintf("Spans [%d]: ", len(m.Spans)))
		const limit = 3
		for i := 0; i < len(m.Spans) && func() bool {
			__antithesis_instrumentation__.Notify(478185)
			return i < limit == true
		}() == true; i++ {
			__antithesis_instrumentation__.Notify(478186)
			if i > 0 {
				__antithesis_instrumentation__.Notify(478188)
				spanStr.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(478189)
			}
			__antithesis_instrumentation__.Notify(478187)
			spanStr.WriteString(m.Spans[i].String())
		}
		__antithesis_instrumentation__.Notify(478184)
		if len(m.Spans) > limit {
			__antithesis_instrumentation__.Notify(478190)
			spanStr.WriteString("...")
		} else {
			__antithesis_instrumentation__.Notify(478191)
		}
	} else {
		__antithesis_instrumentation__.Notify(478192)
	}
	__antithesis_instrumentation__.Notify(478182)

	details := []string{
		spanStr.String(),
	}
	return "BACKUP", details
}

func (d *DistinctSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478193)
	details := []string{
		colListStr(d.DistinctColumns),
	}
	if len(d.OrderedColumns) > 0 {
		__antithesis_instrumentation__.Notify(478195)
		details = append(details, fmt.Sprintf("Ordered: %s", colListStr(d.OrderedColumns)))
	} else {
		__antithesis_instrumentation__.Notify(478196)
	}
	__antithesis_instrumentation__.Notify(478194)
	return "Distinct", details
}

func (o *OrdinalitySpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478197)
	return "Ordinality", []string{}
}

func (d *ProjectSetSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478198)
	var details []string
	for _, expr := range d.Exprs {
		__antithesis_instrumentation__.Notify(478200)
		details = append(details, expr.String())
	}
	__antithesis_instrumentation__.Notify(478199)
	return "ProjectSet", details
}

func (s *SamplerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478201)
	details := []string{fmt.Sprintf("SampleSize: %d", s.SampleSize)}
	for _, sk := range s.Sketches {
		__antithesis_instrumentation__.Notify(478203)
		details = append(details, fmt.Sprintf("Stat: %s", colListStr(sk.Columns)))
	}
	__antithesis_instrumentation__.Notify(478202)

	return "Sampler", details
}

func (s *SampleAggregatorSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478204)
	details := []string{
		fmt.Sprintf("SampleSize: %d", s.SampleSize),
	}
	for _, sk := range s.Sketches {
		__antithesis_instrumentation__.Notify(478206)
		s := fmt.Sprintf("Stat: %s", colListStr(sk.Columns))
		if sk.GenerateHistogram {
			__antithesis_instrumentation__.Notify(478208)
			s = fmt.Sprintf("%s (%d buckets)", s, sk.HistogramMaxBuckets)
		} else {
			__antithesis_instrumentation__.Notify(478209)
		}
		__antithesis_instrumentation__.Notify(478207)
		details = append(details, s)
	}
	__antithesis_instrumentation__.Notify(478205)

	return "SampleAggregator", details
}

func (is *InputSyncSpec) summary(showTypes bool) (string, []string) {
	__antithesis_instrumentation__.Notify(478210)
	typs := make([]string, 0, len(is.ColumnTypes)+1)
	if showTypes {
		__antithesis_instrumentation__.Notify(478212)
		for _, typ := range is.ColumnTypes {
			__antithesis_instrumentation__.Notify(478213)
			typs = append(typs, typ.Name())
		}
	} else {
		__antithesis_instrumentation__.Notify(478214)
	}
	__antithesis_instrumentation__.Notify(478211)
	switch is.Type {
	case InputSyncSpec_PARALLEL_UNORDERED:
		__antithesis_instrumentation__.Notify(478215)
		return "unordered", typs
	case InputSyncSpec_ORDERED:
		__antithesis_instrumentation__.Notify(478216)
		return "ordered", append(typs, is.Ordering.diagramString())
	case InputSyncSpec_SERIAL_UNORDERED:
		__antithesis_instrumentation__.Notify(478217)
		return "serial unordered", typs
	default:
		__antithesis_instrumentation__.Notify(478218)
		return "unknown", []string{}
	}
}

func (r *LocalPlanNodeSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478219)
	return fmt.Sprintf("local %s %d", r.Name, r.RowSourceIdx), []string{}
}

func (r *OutputRouterSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478220)
	switch r.Type {
	case OutputRouterSpec_PASS_THROUGH:
		__antithesis_instrumentation__.Notify(478221)
		return "", []string{}
	case OutputRouterSpec_MIRROR:
		__antithesis_instrumentation__.Notify(478222)
		return "mirror", []string{}
	case OutputRouterSpec_BY_HASH:
		__antithesis_instrumentation__.Notify(478223)
		return "by hash", []string{colListStr(r.HashColumns)}
	case OutputRouterSpec_BY_RANGE:
		__antithesis_instrumentation__.Notify(478224)
		return "by range", []string{}
	default:
		__antithesis_instrumentation__.Notify(478225)
		return "unknown", []string{}
	}
}

func (post *PostProcessSpec) summary() []string {
	__antithesis_instrumentation__.Notify(478226)
	var res []string
	if post.Projection {
		__antithesis_instrumentation__.Notify(478231)
		outputColumns := "None"
		outputCols := post.OutputColumns
		if post.OriginalOutputColumns != nil {
			__antithesis_instrumentation__.Notify(478234)
			outputCols = post.OriginalOutputColumns
		} else {
			__antithesis_instrumentation__.Notify(478235)
		}
		__antithesis_instrumentation__.Notify(478232)
		if len(outputCols) > 0 {
			__antithesis_instrumentation__.Notify(478236)
			outputColumns = colListStr(outputCols)
		} else {
			__antithesis_instrumentation__.Notify(478237)
		}
		__antithesis_instrumentation__.Notify(478233)
		res = append(res, fmt.Sprintf("Out: %s", outputColumns))
	} else {
		__antithesis_instrumentation__.Notify(478238)
	}
	__antithesis_instrumentation__.Notify(478227)
	renderExprs := post.RenderExprs
	if post.OriginalRenderExprs != nil {
		__antithesis_instrumentation__.Notify(478239)
		renderExprs = post.OriginalRenderExprs
	} else {
		__antithesis_instrumentation__.Notify(478240)
	}
	__antithesis_instrumentation__.Notify(478228)
	if len(renderExprs) > 0 {
		__antithesis_instrumentation__.Notify(478241)
		var buf bytes.Buffer
		buf.WriteString("Render: ")
		for i, expr := range renderExprs {
			__antithesis_instrumentation__.Notify(478243)
			if i > 0 {
				__antithesis_instrumentation__.Notify(478245)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(478246)
			}
			__antithesis_instrumentation__.Notify(478244)

			buf.WriteString(strings.Replace(expr.String(), " ", "", -1))
		}
		__antithesis_instrumentation__.Notify(478242)
		res = append(res, buf.String())
	} else {
		__antithesis_instrumentation__.Notify(478247)
	}
	__antithesis_instrumentation__.Notify(478229)
	if post.Limit != 0 || func() bool {
		__antithesis_instrumentation__.Notify(478248)
		return post.Offset != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(478249)
		var buf bytes.Buffer
		if post.Limit != 0 {
			__antithesis_instrumentation__.Notify(478252)
			fmt.Fprintf(&buf, "Limit %d", post.Limit)
		} else {
			__antithesis_instrumentation__.Notify(478253)
		}
		__antithesis_instrumentation__.Notify(478250)
		if post.Offset != 0 {
			__antithesis_instrumentation__.Notify(478254)
			if buf.Len() != 0 {
				__antithesis_instrumentation__.Notify(478256)
				buf.WriteByte(' ')
			} else {
				__antithesis_instrumentation__.Notify(478257)
			}
			__antithesis_instrumentation__.Notify(478255)
			fmt.Fprintf(&buf, "Offset %d", post.Offset)
		} else {
			__antithesis_instrumentation__.Notify(478258)
		}
		__antithesis_instrumentation__.Notify(478251)
		res = append(res, buf.String())
	} else {
		__antithesis_instrumentation__.Notify(478259)
	}
	__antithesis_instrumentation__.Notify(478230)
	return res
}

func (c *RestoreDataSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478260)
	return "RestoreDataSpec", []string{}
}

func (c *SplitAndScatterSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478261)
	detail := fmt.Sprintf("%d chunks", len(c.Chunks))
	return "SplitAndScatterSpec", []string{detail}
}

func (c *ReadImportDataSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478262)
	ss := make([]string, 0, len(c.Uri))
	for _, s := range c.Uri {
		__antithesis_instrumentation__.Notify(478264)
		ss = append(ss, s)
	}
	__antithesis_instrumentation__.Notify(478263)
	return "ReadImportData", ss
}

func (s *StreamIngestionDataSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478265)
	return "StreamIngestionData", []string{}
}

func (s *StreamIngestionFrontierSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478266)
	return "StreamIngestionFrontier", []string{}
}

func (s *IndexBackfillMergerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478267)
	return "IndexBackfillMerger", []string{}
}

func (s *ExportSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478268)
	return "Exporter", []string{s.Destination}
}

func (s *BulkRowWriterSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478269)
	return "BulkRowWriterSpec", []string{}
}

func (w *WindowerSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478270)
	details := make([]string, 0, len(w.WindowFns))
	if len(w.PartitionBy) > 0 {
		__antithesis_instrumentation__.Notify(478273)
		details = append(details, fmt.Sprintf("PARTITION BY: %s", colListStr(w.PartitionBy)))
	} else {
		__antithesis_instrumentation__.Notify(478274)
	}
	__antithesis_instrumentation__.Notify(478271)
	for _, windowFn := range w.WindowFns {
		__antithesis_instrumentation__.Notify(478275)
		var buf bytes.Buffer
		if windowFn.Func.WindowFunc != nil {
			__antithesis_instrumentation__.Notify(478278)
			buf.WriteString(windowFn.Func.WindowFunc.String())
		} else {
			__antithesis_instrumentation__.Notify(478279)
			buf.WriteString(windowFn.Func.AggregateFunc.String())
		}
		__antithesis_instrumentation__.Notify(478276)
		buf.WriteByte('(')
		buf.WriteString(colListStr(windowFn.ArgsIdxs))
		buf.WriteByte(')')
		if len(windowFn.Ordering.Columns) > 0 {
			__antithesis_instrumentation__.Notify(478280)
			buf.WriteString(" (ORDER BY ")
			buf.WriteString(windowFn.Ordering.diagramString())
			buf.WriteByte(')')
		} else {
			__antithesis_instrumentation__.Notify(478281)
		}
		__antithesis_instrumentation__.Notify(478277)
		details = append(details, buf.String())
	}
	__antithesis_instrumentation__.Notify(478272)

	return "Windower", details
}

func (s *ChangeAggregatorSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478282)
	var details []string
	for _, watch := range s.Watches {
		__antithesis_instrumentation__.Notify(478284)
		details = append(details, watch.Span.String())
	}
	__antithesis_instrumentation__.Notify(478283)
	return "ChangeAggregator", details
}

func (s *ChangeFrontierSpec) summary() (string, []string) {
	__antithesis_instrumentation__.Notify(478285)
	return "ChangeFrontier", []string{}
}

type diagramCell struct {
	Title   string   `json:"title"`
	Details []string `json:"details"`
}

type diagramProcessor struct {
	NodeIdx int           `json:"nodeIdx"`
	Inputs  []diagramCell `json:"inputs"`
	Core    diagramCell   `json:"core"`
	Outputs []diagramCell `json:"outputs"`
	StageID int32         `json:"stage"`

	processorID int32
}

type diagramEdge struct {
	SourceProc   int      `json:"sourceProc"`
	SourceOutput int      `json:"sourceOutput"`
	DestProc     int      `json:"destProc"`
	DestInput    int      `json:"destInput"`
	Stats        []string `json:"stats,omitempty"`

	streamID StreamID
}

type FlowDiagram interface {
	ToURL() (string, url.URL, error)

	AddSpans([]tracingpb.RecordedSpan)
}

type diagramData struct {
	SQL        string             `json:"sql"`
	NodeNames  []string           `json:"nodeNames"`
	Processors []diagramProcessor `json:"processors"`
	Edges      []diagramEdge      `json:"edges"`

	flags          DiagramFlags
	flowID         FlowID
	sqlInstanceIDs []base.SQLInstanceID
}

var _ FlowDiagram = &diagramData{}

func (d diagramData) ToURL() (string, url.URL, error) {
	__antithesis_instrumentation__.Notify(478286)
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(d); err != nil {
		__antithesis_instrumentation__.Notify(478288)
		return "", url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(478289)
	}
	__antithesis_instrumentation__.Notify(478287)
	return encodeJSONToURL(buf)
}

func (d *diagramData) AddSpans(spans []tracingpb.RecordedSpan) {
	__antithesis_instrumentation__.Notify(478290)
	statsMap := ExtractStatsFromSpans(spans, d.flags.MakeDeterministic)
	for i := range d.Processors {
		__antithesis_instrumentation__.Notify(478292)
		p := &d.Processors[i]
		sqlInstanceID := d.sqlInstanceIDs[p.NodeIdx]
		component := ProcessorComponentID(sqlInstanceID, d.flowID, p.processorID)
		if compStats := statsMap[component]; compStats != nil {
			__antithesis_instrumentation__.Notify(478293)
			p.Core.Details = append(p.Core.Details, compStats.StatsForQueryPlan()...)
		} else {
			__antithesis_instrumentation__.Notify(478294)
		}
	}
	__antithesis_instrumentation__.Notify(478291)
	for i := range d.Edges {
		__antithesis_instrumentation__.Notify(478295)
		originSQLInstanceID := d.sqlInstanceIDs[d.Processors[d.Edges[i].SourceProc].NodeIdx]
		component := StreamComponentID(originSQLInstanceID, d.flowID, d.Edges[i].streamID)
		if compStats := statsMap[component]; compStats != nil {
			__antithesis_instrumentation__.Notify(478296)
			d.Edges[i].Stats = compStats.StatsForQueryPlan()
		} else {
			__antithesis_instrumentation__.Notify(478297)
		}
	}
}

func generateDiagramData(
	sql string, flows []FlowSpec, sqlInstanceIDs []base.SQLInstanceID, flags DiagramFlags,
) (FlowDiagram, error) {
	__antithesis_instrumentation__.Notify(478298)
	d := &diagramData{
		SQL:            sql,
		sqlInstanceIDs: sqlInstanceIDs,
		flags:          flags,
	}
	d.NodeNames = make([]string, len(sqlInstanceIDs))
	for i := range d.NodeNames {
		__antithesis_instrumentation__.Notify(478304)
		d.NodeNames[i] = sqlInstanceIDs[i].String()
	}
	__antithesis_instrumentation__.Notify(478299)

	if len(flows) > 0 {
		__antithesis_instrumentation__.Notify(478305)
		d.flowID = flows[0].FlowID
		for i := 1; i < len(flows); i++ {
			__antithesis_instrumentation__.Notify(478306)
			if flows[i].FlowID != d.flowID {
				__antithesis_instrumentation__.Notify(478307)
				return nil, errors.AssertionFailedf("flow ID mismatch within a diagram")
			} else {
				__antithesis_instrumentation__.Notify(478308)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(478309)
	}
	__antithesis_instrumentation__.Notify(478300)

	inPorts := make(map[StreamID]diagramEdge)
	syncResponseNode := -1

	pIdx := 0
	for n := range flows {
		__antithesis_instrumentation__.Notify(478310)
		for _, p := range flows[n].Processors {
			__antithesis_instrumentation__.Notify(478311)
			proc := diagramProcessor{NodeIdx: n}
			proc.Core.Title, proc.Core.Details = p.Core.GetValue().(diagramCellType).summary()
			proc.Core.Title += fmt.Sprintf("/%d", p.ProcessorID)
			proc.processorID = p.ProcessorID
			proc.Core.Details = append(proc.Core.Details, p.Post.summary()...)

			if len(p.Input) > 1 || func() bool {
				__antithesis_instrumentation__.Notify(478316)
				return (len(p.Input) == 1 && func() bool {
					__antithesis_instrumentation__.Notify(478317)
					return len(p.Input[0].Streams) > 1 == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(478318)
				proc.Inputs = make([]diagramCell, len(p.Input))
				for i, s := range p.Input {
					__antithesis_instrumentation__.Notify(478319)
					proc.Inputs[i].Title, proc.Inputs[i].Details = s.summary(flags.ShowInputTypes)
				}
			} else {
				__antithesis_instrumentation__.Notify(478320)
				proc.Inputs = []diagramCell{}
			}
			__antithesis_instrumentation__.Notify(478312)

			for i, input := range p.Input {
				__antithesis_instrumentation__.Notify(478321)
				val := diagramEdge{
					DestProc: pIdx,
				}
				if len(proc.Inputs) > 0 {
					__antithesis_instrumentation__.Notify(478323)
					val.DestInput = i + 1
				} else {
					__antithesis_instrumentation__.Notify(478324)
				}
				__antithesis_instrumentation__.Notify(478322)
				for _, stream := range input.Streams {
					__antithesis_instrumentation__.Notify(478325)
					inPorts[stream.StreamID] = val
				}
			}
			__antithesis_instrumentation__.Notify(478313)

			for _, r := range p.Output {
				__antithesis_instrumentation__.Notify(478326)
				for _, o := range r.Streams {
					__antithesis_instrumentation__.Notify(478327)
					if o.Type == StreamEndpointSpec_SYNC_RESPONSE {
						__antithesis_instrumentation__.Notify(478328)
						if syncResponseNode != -1 && func() bool {
							__antithesis_instrumentation__.Notify(478330)
							return syncResponseNode != n == true
						}() == true {
							__antithesis_instrumentation__.Notify(478331)
							return nil, errors.Errorf("multiple nodes with SyncResponse")
						} else {
							__antithesis_instrumentation__.Notify(478332)
						}
						__antithesis_instrumentation__.Notify(478329)
						syncResponseNode = n
					} else {
						__antithesis_instrumentation__.Notify(478333)
					}
				}
			}
			__antithesis_instrumentation__.Notify(478314)

			if len(p.Output) > 1 || func() bool {
				__antithesis_instrumentation__.Notify(478334)
				return (len(p.Output) == 1 && func() bool {
					__antithesis_instrumentation__.Notify(478335)
					return len(p.Output[0].Streams) > 1 == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(478336)
				proc.Outputs = make([]diagramCell, len(p.Output))
				for i, r := range p.Output {
					__antithesis_instrumentation__.Notify(478337)
					proc.Outputs[i].Title, proc.Outputs[i].Details = r.summary()
				}
			} else {
				__antithesis_instrumentation__.Notify(478338)
				proc.Outputs = []diagramCell{}
			}
			__antithesis_instrumentation__.Notify(478315)
			proc.StageID = p.StageID
			d.Processors = append(d.Processors, proc)
			pIdx++
		}
	}
	__antithesis_instrumentation__.Notify(478301)

	if syncResponseNode != -1 {
		__antithesis_instrumentation__.Notify(478339)
		d.Processors = append(d.Processors, diagramProcessor{
			NodeIdx: syncResponseNode,
			Core:    diagramCell{Title: "Response", Details: []string{}},
			Inputs:  []diagramCell{},
			Outputs: []diagramCell{},

			processorID: -1,
		})
	} else {
		__antithesis_instrumentation__.Notify(478340)
	}
	__antithesis_instrumentation__.Notify(478302)

	pIdx = 0
	for n := range flows {
		__antithesis_instrumentation__.Notify(478341)
		for _, p := range flows[n].Processors {
			__antithesis_instrumentation__.Notify(478342)
			for i, output := range p.Output {
				__antithesis_instrumentation__.Notify(478344)
				srcOutput := 0
				if len(d.Processors[pIdx].Outputs) > 0 {
					__antithesis_instrumentation__.Notify(478346)
					srcOutput = i + 1
				} else {
					__antithesis_instrumentation__.Notify(478347)
				}
				__antithesis_instrumentation__.Notify(478345)
				for _, o := range output.Streams {
					__antithesis_instrumentation__.Notify(478348)
					edge := diagramEdge{
						SourceProc:   pIdx,
						SourceOutput: srcOutput,
						streamID:     o.StreamID,
					}
					if o.Type == StreamEndpointSpec_SYNC_RESPONSE {
						__antithesis_instrumentation__.Notify(478350)
						edge.DestProc = len(d.Processors) - 1
					} else {
						__antithesis_instrumentation__.Notify(478351)
						to, ok := inPorts[o.StreamID]
						if !ok {
							__antithesis_instrumentation__.Notify(478353)
							return nil, errors.Errorf("stream %d has no destination", o.StreamID)
						} else {
							__antithesis_instrumentation__.Notify(478354)
						}
						__antithesis_instrumentation__.Notify(478352)
						edge.DestProc = to.DestProc
						edge.DestInput = to.DestInput
					}
					__antithesis_instrumentation__.Notify(478349)
					d.Edges = append(d.Edges, edge)
				}
			}
			__antithesis_instrumentation__.Notify(478343)
			pIdx++
		}
	}
	__antithesis_instrumentation__.Notify(478303)

	return d, nil
}

func GeneratePlanDiagram(
	sql string, flows map[base.SQLInstanceID]*FlowSpec, flags DiagramFlags,
) (FlowDiagram, error) {
	__antithesis_instrumentation__.Notify(478355)

	sqlInstanceIDs := make([]base.SQLInstanceID, 0, len(flows))
	for n := range flows {
		__antithesis_instrumentation__.Notify(478359)
		sqlInstanceIDs = append(sqlInstanceIDs, n)
	}
	__antithesis_instrumentation__.Notify(478356)
	sort.Slice(sqlInstanceIDs, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(478360)
		return sqlInstanceIDs[i] < sqlInstanceIDs[j]
	})
	__antithesis_instrumentation__.Notify(478357)

	flowSlice := make([]FlowSpec, len(sqlInstanceIDs))
	for i, n := range sqlInstanceIDs {
		__antithesis_instrumentation__.Notify(478361)
		flowSlice[i] = *flows[n]
	}
	__antithesis_instrumentation__.Notify(478358)

	return generateDiagramData(sql, flowSlice, sqlInstanceIDs, flags)
}

func GeneratePlanDiagramURL(
	sql string, flows map[base.SQLInstanceID]*FlowSpec, flags DiagramFlags,
) (string, url.URL, error) {
	__antithesis_instrumentation__.Notify(478362)
	d, err := GeneratePlanDiagram(sql, flows, flags)
	if err != nil {
		__antithesis_instrumentation__.Notify(478364)
		return "", url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(478365)
	}
	__antithesis_instrumentation__.Notify(478363)
	return d.ToURL()
}

func encodeJSONToURL(json bytes.Buffer) (string, url.URL, error) {
	__antithesis_instrumentation__.Notify(478366)
	var compressed bytes.Buffer
	jsonStr := json.String()

	encoder := base64.NewEncoder(base64.URLEncoding, &compressed)
	compressor := zlib.NewWriter(encoder)
	if _, err := json.WriteTo(compressor); err != nil {
		__antithesis_instrumentation__.Notify(478370)
		return "", url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(478371)
	}
	__antithesis_instrumentation__.Notify(478367)
	if err := compressor.Close(); err != nil {
		__antithesis_instrumentation__.Notify(478372)
		return "", url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(478373)
	}
	__antithesis_instrumentation__.Notify(478368)
	if err := encoder.Close(); err != nil {
		__antithesis_instrumentation__.Notify(478374)
		return "", url.URL{}, err
	} else {
		__antithesis_instrumentation__.Notify(478375)
	}
	__antithesis_instrumentation__.Notify(478369)
	url := url.URL{
		Scheme:   "https",
		Host:     "cockroachdb.github.io",
		Path:     "distsqlplan/decode.html",
		Fragment: compressed.String(),
	}
	return jsonStr, url, nil
}
