package span

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/constraint"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

type Builder struct {
	evalCtx *tree.EvalContext
	codec   keys.SQLCodec

	keyAndPrefixCols []descpb.IndexFetchSpec_KeyColumn

	KeyPrefix []byte
	alloc     tree.DatumAlloc
}

func (s *Builder) Init(
	evalCtx *tree.EvalContext,
	codec keys.SQLCodec,
	table catalog.TableDescriptor,
	index catalog.Index,
) {
	__antithesis_instrumentation__.Notify(623513)
	s.evalCtx = evalCtx
	s.codec = codec
	s.keyAndPrefixCols = table.IndexFetchSpecKeyAndSuffixColumns(index)
	s.KeyPrefix = rowenc.MakeIndexKeyPrefix(codec, table.GetID(), index.GetID())
}

func (s *Builder) InitWithFetchSpec(
	evalCtx *tree.EvalContext, codec keys.SQLCodec, spec *descpb.IndexFetchSpec,
) {
	__antithesis_instrumentation__.Notify(623514)
	s.evalCtx = evalCtx
	s.codec = codec
	s.keyAndPrefixCols = spec.KeyAndSuffixColumns
	s.KeyPrefix = rowenc.MakeIndexKeyPrefix(codec, spec.TableID, spec.IndexID)
}

func (s *Builder) SpanFromEncDatums(
	values rowenc.EncDatumRow,
) (_ roachpb.Span, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(623515)
	return rowenc.MakeSpanFromEncDatums(values, s.keyAndPrefixCols, &s.alloc, s.KeyPrefix)
}

func (s *Builder) SpanFromEncDatumsWithRange(
	values rowenc.EncDatumRow,
	prefixLen int,
	startDatum tree.Datum,
	startInclusive bool,
	endDatum tree.Datum,
	endInclusive bool,
) (_ roachpb.Span, containsNull bool, err error) {
	__antithesis_instrumentation__.Notify(623516)

	if s.keyAndPrefixCols[prefixLen-1].Direction == descpb.IndexDescriptor_DESC {
		__antithesis_instrumentation__.Notify(623523)
		startDatum, endDatum = endDatum, startDatum
		startInclusive, endInclusive = endInclusive, startInclusive
	} else {
		__antithesis_instrumentation__.Notify(623524)
	}
	__antithesis_instrumentation__.Notify(623517)

	makeKeyFromRow := func(r rowenc.EncDatumRow) (_ roachpb.Key, containsNull bool, _ error) {
		__antithesis_instrumentation__.Notify(623525)
		return rowenc.MakeKeyFromEncDatums(r, s.keyAndPrefixCols, &s.alloc, s.KeyPrefix)
	}
	__antithesis_instrumentation__.Notify(623518)

	var startKey, endKey roachpb.Key
	var startContainsNull, endContainsNull bool
	if startDatum != nil {
		__antithesis_instrumentation__.Notify(623526)
		values[prefixLen-1] = rowenc.EncDatum{Datum: startDatum}
		startKey, startContainsNull, err = makeKeyFromRow(values[:prefixLen])
		if !startInclusive {
			__antithesis_instrumentation__.Notify(623527)
			startKey = startKey.Next()
		} else {
			__antithesis_instrumentation__.Notify(623528)
		}
	} else {
		__antithesis_instrumentation__.Notify(623529)
		startKey, startContainsNull, err = makeKeyFromRow(values[:prefixLen-1])

		if s.keyAndPrefixCols[prefixLen-1].Direction == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(623531)
			startKey = encoding.EncodeNullAscending(startKey)
		} else {
			__antithesis_instrumentation__.Notify(623532)
		}
		__antithesis_instrumentation__.Notify(623530)
		startKey = startKey.Next()
	}
	__antithesis_instrumentation__.Notify(623519)

	if err != nil {
		__antithesis_instrumentation__.Notify(623533)
		return roachpb.Span{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(623534)
	}
	__antithesis_instrumentation__.Notify(623520)

	if endDatum != nil {
		__antithesis_instrumentation__.Notify(623535)
		values[prefixLen-1] = rowenc.EncDatum{Datum: endDatum}
		endKey, endContainsNull, err = makeKeyFromRow(values[:prefixLen])
		if endInclusive {
			__antithesis_instrumentation__.Notify(623536)
			endKey = endKey.PrefixEnd()
		} else {
			__antithesis_instrumentation__.Notify(623537)
		}
	} else {
		__antithesis_instrumentation__.Notify(623538)
		endKey, endContainsNull, err = makeKeyFromRow(values[:prefixLen-1])

		if s.keyAndPrefixCols[prefixLen-1].Direction == descpb.IndexDescriptor_DESC {
			__antithesis_instrumentation__.Notify(623539)
			endKey = encoding.EncodeNullDescending(endKey)
		} else {
			__antithesis_instrumentation__.Notify(623540)
			endKey = endKey.PrefixEnd()
		}
	}
	__antithesis_instrumentation__.Notify(623521)

	if err != nil {
		__antithesis_instrumentation__.Notify(623541)
		return roachpb.Span{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(623542)
	}
	__antithesis_instrumentation__.Notify(623522)

	return roachpb.Span{Key: startKey, EndKey: endKey}, startContainsNull || func() bool {
		__antithesis_instrumentation__.Notify(623543)
		return endContainsNull == true
	}() == true, nil
}

func (s *Builder) SpanFromDatumRow(
	values tree.Datums, prefixLen int, colMap catalog.TableColMap,
) (_ roachpb.Span, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(623544)
	return rowenc.EncodePartialIndexSpan(s.keyAndPrefixCols[:prefixLen], colMap, values, s.KeyPrefix)
}

func (s *Builder) SpanToPointSpan(span roachpb.Span, family descpb.FamilyID) roachpb.Span {
	__antithesis_instrumentation__.Notify(623545)
	key := keys.MakeFamilyKey(span.Key, uint32(family))
	return roachpb.Span{Key: key, EndKey: roachpb.Key(key).PrefixEnd()}
}

func (s *Builder) SpansFromConstraint(
	c *constraint.Constraint, splitter Splitter,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(623546)
	var spans roachpb.Spans
	var err error
	if c == nil || func() bool {
		__antithesis_instrumentation__.Notify(623549)
		return c.IsUnconstrained() == true
	}() == true {
		__antithesis_instrumentation__.Notify(623550)

		spans, err = s.appendSpansFromConstraintSpan(spans, &constraint.UnconstrainedSpan, splitter)
		if err != nil {
			__antithesis_instrumentation__.Notify(623552)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623553)
		}
		__antithesis_instrumentation__.Notify(623551)
		return spans, nil
	} else {
		__antithesis_instrumentation__.Notify(623554)
	}
	__antithesis_instrumentation__.Notify(623547)

	spans = make(roachpb.Spans, 0, c.Spans.Count())
	for i := 0; i < c.Spans.Count(); i++ {
		__antithesis_instrumentation__.Notify(623555)
		spans, err = s.appendSpansFromConstraintSpan(spans, c.Spans.Get(i), splitter)
		if err != nil {
			__antithesis_instrumentation__.Notify(623556)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623557)
		}
	}
	__antithesis_instrumentation__.Notify(623548)
	return spans, nil
}

func (s *Builder) UnconstrainedSpans() (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(623558)
	return s.SpansFromConstraint(nil, NoopSplitter())
}

func (s *Builder) appendSpansFromConstraintSpan(
	appendTo roachpb.Spans, cs *constraint.Span, splitter Splitter,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(623559)
	var span roachpb.Span
	var err error
	var containsNull bool

	span.Key, containsNull, err = s.encodeConstraintKey(cs.StartKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(623566)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623567)
	}
	__antithesis_instrumentation__.Notify(623560)
	if cs.StartBoundary() == constraint.IncludeBoundary {
		__antithesis_instrumentation__.Notify(623568)
		if cs.StartKey().IsEmpty() {
			__antithesis_instrumentation__.Notify(623569)
			span.Key = append(span.Key, s.KeyPrefix...)
		} else {
			__antithesis_instrumentation__.Notify(623570)
		}
	} else {
		__antithesis_instrumentation__.Notify(623571)

		span.Key = span.Key.PrefixEnd()
	}
	__antithesis_instrumentation__.Notify(623561)

	span.EndKey, _, err = s.encodeConstraintKey(cs.EndKey())
	if err != nil {
		__antithesis_instrumentation__.Notify(623572)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623573)
	}
	__antithesis_instrumentation__.Notify(623562)
	if cs.EndKey().IsEmpty() {
		__antithesis_instrumentation__.Notify(623574)
		span.EndKey = append(span.EndKey, s.KeyPrefix...)
	} else {
		__antithesis_instrumentation__.Notify(623575)
	}
	__antithesis_instrumentation__.Notify(623563)

	if splitter.CanSplitSpanIntoFamilySpans(cs.StartKey().Length(), containsNull) && func() bool {
		__antithesis_instrumentation__.Notify(623576)
		return span.Key.Equal(span.EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(623577)
		return rowenc.SplitRowKeyIntoFamilySpans(appendTo, span.Key, splitter.neededFamilies), nil
	} else {
		__antithesis_instrumentation__.Notify(623578)
	}
	__antithesis_instrumentation__.Notify(623564)

	if cs.EndBoundary() == constraint.IncludeBoundary {
		__antithesis_instrumentation__.Notify(623579)
		span.EndKey = span.EndKey.PrefixEnd()
	} else {
		__antithesis_instrumentation__.Notify(623580)
	}
	__antithesis_instrumentation__.Notify(623565)

	return append(appendTo, span), nil
}

func (s *Builder) encodeConstraintKey(
	ck constraint.Key,
) (key roachpb.Key, containsNull bool, _ error) {
	__antithesis_instrumentation__.Notify(623581)
	if ck.IsEmpty() {
		__antithesis_instrumentation__.Notify(623584)
		return key, containsNull, nil
	} else {
		__antithesis_instrumentation__.Notify(623585)
	}
	__antithesis_instrumentation__.Notify(623582)
	key = append(key, s.KeyPrefix...)
	for i := 0; i < ck.Length(); i++ {
		__antithesis_instrumentation__.Notify(623586)
		val := ck.Value(i)
		if val == tree.DNull {
			__antithesis_instrumentation__.Notify(623589)
			containsNull = true
		} else {
			__antithesis_instrumentation__.Notify(623590)
		}
		__antithesis_instrumentation__.Notify(623587)

		dir, err := s.keyAndPrefixCols[i].Direction.ToEncodingDirection()
		if err != nil {
			__antithesis_instrumentation__.Notify(623591)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(623592)
		}
		__antithesis_instrumentation__.Notify(623588)

		key, err = keyside.Encode(key, val, dir)
		if err != nil {
			__antithesis_instrumentation__.Notify(623593)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(623594)
		}
	}
	__antithesis_instrumentation__.Notify(623583)
	return key, containsNull, nil
}

type InvertedSpans interface {
	Len() int

	Start(i int) []byte

	End(i int) []byte
}

var _ InvertedSpans = inverted.Spans{}
var _ InvertedSpans = inverted.SpanExpressionProtoSpans{}

func (s *Builder) SpansFromInvertedSpans(
	invertedSpans InvertedSpans, c *constraint.Constraint, scratch roachpb.Spans,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(623595)
	if invertedSpans == nil {
		__antithesis_instrumentation__.Notify(623599)
		return nil, errors.AssertionFailedf("invertedSpans cannot be nil")
	} else {
		__antithesis_instrumentation__.Notify(623600)
	}
	__antithesis_instrumentation__.Notify(623596)

	var scratchRows []rowenc.EncDatumRow
	if c != nil {
		__antithesis_instrumentation__.Notify(623601)

		scratchRows = make([]rowenc.EncDatumRow, c.Spans.Count())
		for i, n := 0, c.Spans.Count(); i < n; i++ {
			__antithesis_instrumentation__.Notify(623602)
			span := c.Spans.Get(i)

			if !span.HasSingleKey(s.evalCtx) {
				__antithesis_instrumentation__.Notify(623604)
				return nil, errors.AssertionFailedf("constraint span %s does not have a single key", span)
			} else {
				__antithesis_instrumentation__.Notify(623605)
			}
			__antithesis_instrumentation__.Notify(623603)

			keyLength := span.StartKey().Length()
			scratchRows[i] = make(rowenc.EncDatumRow, keyLength+1)
			for j := 0; j < keyLength; j++ {
				__antithesis_instrumentation__.Notify(623606)
				val := span.StartKey().Value(j)
				scratchRows[i][j] = rowenc.DatumToEncDatum(val.ResolvedType(), val)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(623607)

		scratchRows = make([]rowenc.EncDatumRow, 1)
		scratchRows[0] = make(rowenc.EncDatumRow, 1)
	}
	__antithesis_instrumentation__.Notify(623597)

	scratch = scratch[:0]
	for i := range scratchRows {
		__antithesis_instrumentation__.Notify(623608)
		for j, n := 0, invertedSpans.Len(); j < n; j++ {
			__antithesis_instrumentation__.Notify(623609)
			var indexSpan roachpb.Span
			var err error
			if indexSpan.Key, err = s.generateInvertedSpanKey(invertedSpans.Start(j), scratchRows[i]); err != nil {
				__antithesis_instrumentation__.Notify(623612)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623613)
			}
			__antithesis_instrumentation__.Notify(623610)
			if indexSpan.EndKey, err = s.generateInvertedSpanKey(invertedSpans.End(j), scratchRows[i]); err != nil {
				__antithesis_instrumentation__.Notify(623614)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623615)
			}
			__antithesis_instrumentation__.Notify(623611)
			scratch = append(scratch, indexSpan)
		}
	}
	__antithesis_instrumentation__.Notify(623598)
	sort.Sort(scratch)
	return scratch, nil
}

func (s *Builder) generateInvertedSpanKey(
	enc []byte, scratchRow rowenc.EncDatumRow,
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(623616)
	keyLen := len(scratchRow) - 1
	scratchRow = scratchRow[:keyLen]
	if len(enc) > 0 {
		__antithesis_instrumentation__.Notify(623618)

		encDatum := rowenc.EncDatumFromEncoded(descpb.DatumEncoding_ASCENDING_KEY, enc)
		scratchRow = append(scratchRow, encDatum)
		keyLen++
	} else {
		__antithesis_instrumentation__.Notify(623619)
	}
	__antithesis_instrumentation__.Notify(623617)

	span, _, err := s.SpanFromEncDatums(scratchRow[:keyLen])
	return span.Key, err
}
