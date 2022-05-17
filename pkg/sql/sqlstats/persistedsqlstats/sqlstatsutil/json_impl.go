package sqlstatsutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/hex"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/errors"
)

type jsonDecoder interface {
	decodeJSON(js json.JSON) error
}

type jsonEncoder interface {
	encodeJSON() (json.JSON, error)
}

type jsonMarshaler interface {
	jsonEncoder
	jsonDecoder
}

var (
	_ jsonMarshaler = &stmtFingerprintIDArray{}
	_ jsonMarshaler = &stmtStats{}
	_ jsonMarshaler = &txnStats{}
	_ jsonMarshaler = &innerTxnStats{}
	_ jsonMarshaler = &innerStmtStats{}
	_ jsonMarshaler = &execStats{}
	_ jsonMarshaler = &numericStats{}
	_ jsonMarshaler = jsonFields{}
	_ jsonMarshaler = &decimal{}
	_ jsonMarshaler = (*jsonFloat)(nil)
	_ jsonMarshaler = (*jsonString)(nil)
	_ jsonMarshaler = (*jsonBool)(nil)
	_ jsonMarshaler = (*jsonInt)(nil)
	_ jsonMarshaler = (*stmtFingerprintID)(nil)
	_ jsonMarshaler = (*int64Array)(nil)
)

type txnStats roachpb.TransactionStatistics

func (t *txnStats) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625081)
	return jsonFields{
		{"statistics", (*innerTxnStats)(t)},
		{"execution_statistics", (*execStats)(&t.ExecStats)},
	}
}

func (t *txnStats) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625082)
	return t.jsonFields().decodeJSON(js)
}

func (t *txnStats) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625083)
	return t.jsonFields().encodeJSON()
}

type stmtStats roachpb.StatementStatistics

func (s *stmtStats) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625084)
	return jsonFields{
		{"statistics", (*innerStmtStats)(s)},
		{"execution_statistics", (*execStats)(&s.ExecStats)},
	}
}

func (s *stmtStats) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625085)
	return s.jsonFields().decodeJSON(js)
}

func (s *stmtStats) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625086)
	return s.jsonFields().encodeJSON()
}

type stmtStatsMetadata roachpb.CollectedStatementStatistics

func (s *stmtStatsMetadata) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625087)
	return jsonFields{
		{"stmtTyp", (*jsonString)(&s.Stats.SQLType)},
		{"query", (*jsonString)(&s.Key.Query)},
		{"querySummary", (*jsonString)(&s.Key.QuerySummary)},
		{"db", (*jsonString)(&s.Key.Database)},
		{"distsql", (*jsonBool)(&s.Key.DistSQL)},
		{"failed", (*jsonBool)(&s.Key.Failed)},
		{"implicitTxn", (*jsonBool)(&s.Key.ImplicitTxn)},
		{"vec", (*jsonBool)(&s.Key.Vec)},
		{"fullScan", (*jsonBool)(&s.Key.FullScan)},
	}
}

type aggregatedMetadata roachpb.AggregatedStatementMetadata

func (s *aggregatedMetadata) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625088)
	return jsonFields{
		{"db", (*stringArray)(&s.Databases)},
		{"appNames", (*stringArray)(&s.AppNames)},
		{"distSQLCount", (*jsonInt)(&s.DistSQLCount)},
		{"failedCount", (*jsonInt)(&s.FailedCount)},
		{"fullScanCount", (*jsonInt)(&s.FullScanCount)},
		{"implicitTxn", (*jsonBool)(&s.ImplicitTxn)},
		{"query", (*jsonString)(&s.Query)},
		{"formattedQuery", (*jsonString)(&s.FormattedQuery)},
		{"querySummary", (*jsonString)(&s.QuerySummary)},
		{"stmtType", (*jsonString)(&s.StmtType)},
		{"vecCount", (*jsonInt)(&s.VecCount)},
		{"totalCount", (*jsonInt)(&s.TotalCount)},
	}
}

type int64Array []int64

func (a *int64Array) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625089)
	arrLen := js.Len()
	for i := 0; i < arrLen; i++ {
		__antithesis_instrumentation__.Notify(625091)
		var value jsonInt
		valJSON, err := js.FetchValIdx(i)
		if err != nil {
			__antithesis_instrumentation__.Notify(625094)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625095)
		}
		__antithesis_instrumentation__.Notify(625092)
		if err := value.decodeJSON(valJSON); err != nil {
			__antithesis_instrumentation__.Notify(625096)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625097)
		}
		__antithesis_instrumentation__.Notify(625093)
		*a = append(*a, int64(value))
	}
	__antithesis_instrumentation__.Notify(625090)

	return nil
}

func (a *int64Array) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625098)
	builder := json.NewArrayBuilder(len(*a))

	for _, value := range *a {
		__antithesis_instrumentation__.Notify(625100)
		jsVal, err := (*jsonInt)(&value).encodeJSON()
		if err != nil {
			__antithesis_instrumentation__.Notify(625102)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625103)
		}
		__antithesis_instrumentation__.Notify(625101)
		builder.Add(jsVal)
	}
	__antithesis_instrumentation__.Notify(625099)

	return builder.Build(), nil
}

type stringArray []string

func (a *stringArray) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625104)
	arrLen := js.Len()
	for i := 0; i < arrLen; i++ {
		__antithesis_instrumentation__.Notify(625106)
		var value jsonString
		valJSON, err := js.FetchValIdx(i)
		if err != nil {
			__antithesis_instrumentation__.Notify(625109)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625110)
		}
		__antithesis_instrumentation__.Notify(625107)
		if err := value.decodeJSON(valJSON); err != nil {
			__antithesis_instrumentation__.Notify(625111)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625112)
		}
		__antithesis_instrumentation__.Notify(625108)
		*a = append(*a, string(value))
	}
	__antithesis_instrumentation__.Notify(625105)

	return nil
}

func (a *stringArray) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625113)
	builder := json.NewArrayBuilder(len(*a))

	for _, value := range *a {
		__antithesis_instrumentation__.Notify(625115)
		jsVal, err := (*jsonString)(&value).encodeJSON()
		if err != nil {
			__antithesis_instrumentation__.Notify(625117)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625118)
		}
		__antithesis_instrumentation__.Notify(625116)
		builder.Add(jsVal)
	}
	__antithesis_instrumentation__.Notify(625114)

	return builder.Build(), nil
}

type stmtFingerprintIDArray []roachpb.StmtFingerprintID

func (s *stmtFingerprintIDArray) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625119)
	arrLen := js.Len()
	for i := 0; i < arrLen; i++ {
		__antithesis_instrumentation__.Notify(625121)
		var fingerprintID stmtFingerprintID
		fingerprintIDJSON, err := js.FetchValIdx(i)
		if err != nil {
			__antithesis_instrumentation__.Notify(625124)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625125)
		}
		__antithesis_instrumentation__.Notify(625122)
		if err := fingerprintID.decodeJSON(fingerprintIDJSON); err != nil {
			__antithesis_instrumentation__.Notify(625126)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625127)
		}
		__antithesis_instrumentation__.Notify(625123)
		*s = append(*s, roachpb.StmtFingerprintID(fingerprintID))
	}
	__antithesis_instrumentation__.Notify(625120)

	return nil
}

func (s *stmtFingerprintIDArray) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625128)
	builder := json.NewArrayBuilder(len(*s))

	for _, fingerprintID := range *s {
		__antithesis_instrumentation__.Notify(625130)
		jsVal, err := (*stmtFingerprintID)(&fingerprintID).encodeJSON()
		if err != nil {
			__antithesis_instrumentation__.Notify(625132)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625133)
		}
		__antithesis_instrumentation__.Notify(625131)
		builder.Add(jsVal)
	}
	__antithesis_instrumentation__.Notify(625129)

	return builder.Build(), nil
}

type stmtFingerprintID roachpb.StmtFingerprintID

func (s *stmtFingerprintID) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625134)
	var str jsonString
	if err := str.decodeJSON(js); err != nil {
		__antithesis_instrumentation__.Notify(625138)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625139)
	}
	__antithesis_instrumentation__.Notify(625135)

	decodedString, err := hex.DecodeString(string(str))
	if err != nil {
		__antithesis_instrumentation__.Notify(625140)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625141)
	}
	__antithesis_instrumentation__.Notify(625136)

	_, fingerprintID, err := encoding.DecodeUint64Ascending(decodedString)
	if err != nil {
		__antithesis_instrumentation__.Notify(625142)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625143)
	}
	__antithesis_instrumentation__.Notify(625137)

	*s = stmtFingerprintID(fingerprintID)
	return nil
}

func (s *stmtFingerprintID) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625144)
	return json.FromString(
		encodeStmtFingerprintIDToString((roachpb.StmtFingerprintID)(*s))), nil
}

type innerTxnStats roachpb.TransactionStatistics

func (t *innerTxnStats) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625145)
	return jsonFields{
		{"cnt", (*jsonInt)(&t.Count)},
		{"maxRetries", (*jsonInt)(&t.MaxRetries)},
		{"numRows", (*numericStats)(&t.NumRows)},
		{"svcLat", (*numericStats)(&t.ServiceLat)},
		{"retryLat", (*numericStats)(&t.RetryLat)},
		{"commitLat", (*numericStats)(&t.CommitLat)},
		{"bytesRead", (*numericStats)(&t.BytesRead)},
		{"rowsRead", (*numericStats)(&t.RowsRead)},
		{"rowsWritten", (*numericStats)(&t.RowsWritten)},
	}
}

func (t *innerTxnStats) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625146)
	return t.jsonFields().decodeJSON(js)
}

func (t *innerTxnStats) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625147)
	return t.jsonFields().encodeJSON()
}

type innerStmtStats roachpb.StatementStatistics

func (s *innerStmtStats) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625148)
	return jsonFields{
		{"cnt", (*jsonInt)(&s.Count)},
		{"firstAttemptCnt", (*jsonInt)(&s.FirstAttemptCount)},
		{"maxRetries", (*jsonInt)(&s.MaxRetries)},
		{"lastExecAt", (*jsonTime)(&s.LastExecTimestamp)},
		{"numRows", (*numericStats)(&s.NumRows)},
		{"parseLat", (*numericStats)(&s.ParseLat)},
		{"planLat", (*numericStats)(&s.PlanLat)},
		{"runLat", (*numericStats)(&s.RunLat)},
		{"svcLat", (*numericStats)(&s.ServiceLat)},
		{"ovhLat", (*numericStats)(&s.OverheadLat)},
		{"bytesRead", (*numericStats)(&s.BytesRead)},
		{"rowsRead", (*numericStats)(&s.RowsRead)},
		{"rowsWritten", (*numericStats)(&s.RowsWritten)},
		{"nodes", (*int64Array)(&s.Nodes)},
		{"planGists", (*stringArray)(&s.PlanGists)},
	}
}

func (s *innerStmtStats) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625149)
	return s.jsonFields().decodeJSON(js)
}

func (s *innerStmtStats) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625150)
	return s.jsonFields().encodeJSON()
}

type execStats roachpb.ExecStats

func (e *execStats) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625151)
	return jsonFields{
		{"cnt", (*jsonInt)(&e.Count)},
		{"networkBytes", (*numericStats)(&e.NetworkBytes)},
		{"maxMemUsage", (*numericStats)(&e.MaxMemUsage)},
		{"contentionTime", (*numericStats)(&e.ContentionTime)},
		{"networkMsgs", (*numericStats)(&e.NetworkMessages)},
		{"maxDiskUsage", (*numericStats)(&e.MaxDiskUsage)},
	}
}

func (e *execStats) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625152)
	return e.jsonFields().decodeJSON(js)
}

func (e *execStats) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625153)
	return e.jsonFields().encodeJSON()
}

type numericStats roachpb.NumericStat

func (n *numericStats) jsonFields() jsonFields {
	__antithesis_instrumentation__.Notify(625154)
	return jsonFields{
		{"mean", (*jsonFloat)(&n.Mean)},
		{"sqDiff", (*jsonFloat)(&n.SquaredDiffs)},
	}
}

func (n *numericStats) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625155)
	return n.jsonFields().decodeJSON(js)
}

func (n *numericStats) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625156)
	return n.jsonFields().encodeJSON()
}

type jsonFields []jsonField

func (jf jsonFields) decodeJSON(js json.JSON) (err error) {
	__antithesis_instrumentation__.Notify(625157)
	var fieldName string
	defer func() {
		__antithesis_instrumentation__.Notify(625160)
		if err != nil {
			__antithesis_instrumentation__.Notify(625161)
			err = errors.Wrapf(err, "decoding field %s", fieldName)
		} else {
			__antithesis_instrumentation__.Notify(625162)
		}
	}()
	__antithesis_instrumentation__.Notify(625158)

	for i := range jf {
		__antithesis_instrumentation__.Notify(625163)
		fieldName = jf[i].field
		field, err := js.FetchValKey(fieldName)
		if err != nil {
			__antithesis_instrumentation__.Notify(625165)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625166)
		}
		__antithesis_instrumentation__.Notify(625164)
		if field != nil {
			__antithesis_instrumentation__.Notify(625167)
			err = jf[i].val.decodeJSON(field)
			if err != nil {
				__antithesis_instrumentation__.Notify(625168)
				return err
			} else {
				__antithesis_instrumentation__.Notify(625169)
			}
		} else {
			__antithesis_instrumentation__.Notify(625170)
		}
	}
	__antithesis_instrumentation__.Notify(625159)

	return nil
}

func (jf jsonFields) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625171)
	builder := json.NewObjectBuilder(len(jf))
	for i := range jf {
		__antithesis_instrumentation__.Notify(625173)
		jsVal, err := jf[i].val.encodeJSON()
		if err != nil {
			__antithesis_instrumentation__.Notify(625175)
			return nil, errors.Wrapf(err, "encoding field %s", jf[i].field)
		} else {
			__antithesis_instrumentation__.Notify(625176)
		}
		__antithesis_instrumentation__.Notify(625174)
		builder.Add(jf[i].field, jsVal)
	}
	__antithesis_instrumentation__.Notify(625172)
	return builder.Build(), nil
}

type jsonField struct {
	field string
	val   jsonMarshaler
}

type jsonTime time.Time

func (t *jsonTime) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625177)
	var s jsonString
	if err := s.decodeJSON(js); err != nil {
		__antithesis_instrumentation__.Notify(625180)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625181)
	}
	__antithesis_instrumentation__.Notify(625178)

	tm := (*time.Time)(t)
	if err := tm.UnmarshalText([]byte(s)); err != nil {
		__antithesis_instrumentation__.Notify(625182)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625183)
	}
	__antithesis_instrumentation__.Notify(625179)

	return nil
}

func (t *jsonTime) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625184)
	str, err := (time.Time)(*t).MarshalText()
	if err != nil {
		__antithesis_instrumentation__.Notify(625186)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625187)
	}
	__antithesis_instrumentation__.Notify(625185)
	return json.FromString(string(str)), nil
}

type jsonString string

func (s *jsonString) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625188)
	text, err := js.AsText()
	if err != nil {
		__antithesis_instrumentation__.Notify(625190)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625191)
	}
	__antithesis_instrumentation__.Notify(625189)
	*s = (jsonString)(*text)
	return nil
}

func (s *jsonString) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625192)
	return json.FromString(string(*s)), nil
}

type jsonFloat float64

func (f *jsonFloat) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625193)
	var d apd.Decimal
	if err := (*decimal)(&d).decodeJSON(js); err != nil {
		__antithesis_instrumentation__.Notify(625196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625197)
	}
	__antithesis_instrumentation__.Notify(625194)

	val, err := d.Float64()
	if err != nil {
		__antithesis_instrumentation__.Notify(625198)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625199)
	}
	__antithesis_instrumentation__.Notify(625195)
	*f = (jsonFloat)(val)

	return nil
}

func (f *jsonFloat) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625200)
	return json.FromFloat64(float64(*f))
}

type jsonBool bool

func (b *jsonBool) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625201)
	switch js.Type() {
	case json.TrueJSONType:
		__antithesis_instrumentation__.Notify(625203)
		*b = true
	case json.FalseJSONType:
		__antithesis_instrumentation__.Notify(625204)
		*b = false
	default:
		__antithesis_instrumentation__.Notify(625205)
		return errors.New("invalid boolean json value type")
	}
	__antithesis_instrumentation__.Notify(625202)
	return nil
}

func (b *jsonBool) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625206)
	return json.FromBool(bool(*b)), nil
}

type jsonInt int64

func (i *jsonInt) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625207)
	var d apd.Decimal
	if err := (*decimal)(&d).decodeJSON(js); err != nil {
		__antithesis_instrumentation__.Notify(625210)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625211)
	}
	__antithesis_instrumentation__.Notify(625208)
	val, err := d.Int64()
	if err != nil {
		__antithesis_instrumentation__.Notify(625212)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625213)
	}
	__antithesis_instrumentation__.Notify(625209)
	*i = (jsonInt)(val)
	return nil
}

func (i *jsonInt) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625214)
	return json.FromInt64(int64(*i)), nil
}

type decimal apd.Decimal

func (d *decimal) decodeJSON(js json.JSON) error {
	__antithesis_instrumentation__.Notify(625215)
	dec, ok := js.AsDecimal()
	if !ok {
		__antithesis_instrumentation__.Notify(625217)
		return errors.New("unable to decode decimal")
	} else {
		__antithesis_instrumentation__.Notify(625218)
	}
	__antithesis_instrumentation__.Notify(625216)
	*d = (decimal)(*dec)
	return nil
}

func (d *decimal) encodeJSON() (json.JSON, error) {
	__antithesis_instrumentation__.Notify(625219)
	return json.FromDecimal(*(*apd.Decimal)(d)), nil
}
