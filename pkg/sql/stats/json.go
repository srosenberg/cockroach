package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type JSONStatistic struct {
	Name          string   `json:"name,omitempty"`
	CreatedAt     string   `json:"created_at"`
	Columns       []string `json:"columns"`
	RowCount      uint64   `json:"row_count"`
	DistinctCount uint64   `json:"distinct_count"`
	NullCount     uint64   `json:"null_count"`
	AvgSize       uint64   `json:"avg_size"`

	HistogramColumnType string            `json:"histo_col_type"`
	HistogramBuckets    []JSONHistoBucket `json:"histo_buckets,omitempty"`
	HistogramVersion    HistogramVersion  `json:"histo_version,omitempty"`
}

type JSONHistoBucket struct {
	NumEq         int64   `json:"num_eq"`
	NumRange      int64   `json:"num_range"`
	DistinctRange float64 `json:"distinct_range"`

	UpperBound string `json:"upper_bound"`
}

func (js *JSONStatistic) SetHistogram(h *HistogramData) error {
	__antithesis_instrumentation__.Notify(626662)
	typ := h.ColumnType
	if typ == nil {
		__antithesis_instrumentation__.Notify(626665)
		return fmt.Errorf("histogram type is unset")
	} else {
		__antithesis_instrumentation__.Notify(626666)
	}
	__antithesis_instrumentation__.Notify(626663)
	js.HistogramColumnType = typ.SQLString()
	js.HistogramBuckets = make([]JSONHistoBucket, len(h.Buckets))
	js.HistogramVersion = h.Version
	var a tree.DatumAlloc
	for i := range h.Buckets {
		__antithesis_instrumentation__.Notify(626667)
		b := &h.Buckets[i]
		js.HistogramBuckets[i].NumEq = b.NumEq
		js.HistogramBuckets[i].NumRange = b.NumRange
		js.HistogramBuckets[i].DistinctRange = b.DistinctRange

		if b.UpperBound == nil {
			__antithesis_instrumentation__.Notify(626670)
			return fmt.Errorf("histogram bucket upper bound is unset")
		} else {
			__antithesis_instrumentation__.Notify(626671)
		}
		__antithesis_instrumentation__.Notify(626668)
		datum, _, err := keyside.Decode(&a, typ, b.UpperBound, encoding.Ascending)
		if err != nil {
			__antithesis_instrumentation__.Notify(626672)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626673)
		}
		__antithesis_instrumentation__.Notify(626669)

		js.HistogramBuckets[i] = JSONHistoBucket{
			NumEq:         b.NumEq,
			NumRange:      b.NumRange,
			DistinctRange: b.DistinctRange,
			UpperBound:    tree.AsStringWithFlags(datum, tree.FmtExport),
		}
	}
	__antithesis_instrumentation__.Notify(626664)
	return nil
}

func (js *JSONStatistic) DecodeAndSetHistogram(
	ctx context.Context, semaCtx *tree.SemaContext, datum tree.Datum,
) error {
	__antithesis_instrumentation__.Notify(626674)
	if datum == tree.DNull {
		__antithesis_instrumentation__.Notify(626680)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(626681)
	}
	__antithesis_instrumentation__.Notify(626675)
	if datum.ResolvedType().Family() != types.BytesFamily {
		__antithesis_instrumentation__.Notify(626682)
		return fmt.Errorf("histogram datum type should be Bytes")
	} else {
		__antithesis_instrumentation__.Notify(626683)
	}
	__antithesis_instrumentation__.Notify(626676)
	if len(*datum.(*tree.DBytes)) == 0 {
		__antithesis_instrumentation__.Notify(626684)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(626685)
	}
	__antithesis_instrumentation__.Notify(626677)
	h := &HistogramData{}
	if err := protoutil.Unmarshal([]byte(*datum.(*tree.DBytes)), h); err != nil {
		__antithesis_instrumentation__.Notify(626686)
		return err
	} else {
		__antithesis_instrumentation__.Notify(626687)
	}
	__antithesis_instrumentation__.Notify(626678)

	if h.ColumnType.UserDefined() {
		__antithesis_instrumentation__.Notify(626688)
		resolver := semaCtx.GetTypeResolver()
		if resolver == nil {
			__antithesis_instrumentation__.Notify(626691)
			return errors.AssertionFailedf("attempt to resolve user defined type with nil TypeResolver")
		} else {
			__antithesis_instrumentation__.Notify(626692)
		}
		__antithesis_instrumentation__.Notify(626689)
		typ, err := resolver.ResolveTypeByOID(ctx, h.ColumnType.Oid())
		if err != nil {
			__antithesis_instrumentation__.Notify(626693)
			return err
		} else {
			__antithesis_instrumentation__.Notify(626694)
		}
		__antithesis_instrumentation__.Notify(626690)
		h.ColumnType = typ
	} else {
		__antithesis_instrumentation__.Notify(626695)
	}
	__antithesis_instrumentation__.Notify(626679)
	return js.SetHistogram(h)
}

func (js *JSONStatistic) GetHistogram(
	semaCtx *tree.SemaContext, evalCtx *tree.EvalContext,
) (*HistogramData, error) {
	__antithesis_instrumentation__.Notify(626696)
	if len(js.HistogramBuckets) == 0 {
		__antithesis_instrumentation__.Notify(626701)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(626702)
	}
	__antithesis_instrumentation__.Notify(626697)
	h := &HistogramData{}
	colTypeRef, err := parser.GetTypeFromValidSQLSyntax(js.HistogramColumnType)
	if err != nil {
		__antithesis_instrumentation__.Notify(626703)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(626704)
	}
	__antithesis_instrumentation__.Notify(626698)
	colType, err := tree.ResolveType(evalCtx.Context, colTypeRef, semaCtx.GetTypeResolver())
	if err != nil {
		__antithesis_instrumentation__.Notify(626705)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(626706)
	}
	__antithesis_instrumentation__.Notify(626699)
	h.ColumnType = colType
	h.Version = js.HistogramVersion
	h.Buckets = make([]HistogramData_Bucket, len(js.HistogramBuckets))
	for i := range h.Buckets {
		__antithesis_instrumentation__.Notify(626707)
		hb := &js.HistogramBuckets[i]
		upperVal, err := rowenc.ParseDatumStringAs(colType, hb.UpperBound, evalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(626709)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(626710)
		}
		__antithesis_instrumentation__.Notify(626708)
		h.Buckets[i].NumEq = hb.NumEq
		h.Buckets[i].NumRange = hb.NumRange
		h.Buckets[i].DistinctRange = hb.DistinctRange
		h.Buckets[i].UpperBound, err = keyside.Encode(nil, upperVal, encoding.Ascending)
		if err != nil {
			__antithesis_instrumentation__.Notify(626711)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(626712)
		}
	}
	__antithesis_instrumentation__.Notify(626700)
	return h, nil
}
