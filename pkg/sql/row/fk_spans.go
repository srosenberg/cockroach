package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
)

func FKCheckSpan(
	builder *span.Builder,
	splitter span.Splitter,
	values []tree.Datum,
	colMap catalog.TableColMap,
	numCols int,
) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(568020)
	span, containsNull, err := builder.SpanFromDatumRow(values, numCols, colMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(568022)
		return roachpb.Span{}, err
	} else {
		__antithesis_instrumentation__.Notify(568023)
	}
	__antithesis_instrumentation__.Notify(568021)
	return splitter.ExistenceCheckSpan(span, numCols, containsNull), nil
}
