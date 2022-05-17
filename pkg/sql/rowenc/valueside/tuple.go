package valueside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
)

func encodeTuple(t *tree.DTuple, appendTo []byte, colID uint32, scratch []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(571609)
	appendTo = encoding.EncodeValueTag(appendTo, colID, encoding.Tuple)
	return encodeUntaggedTuple(t, appendTo, colID, scratch)
}

func encodeUntaggedTuple(
	t *tree.DTuple, appendTo []byte, colID uint32, scratch []byte,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(571610)
	appendTo = encoding.EncodeNonsortingUvarint(appendTo, uint64(len(t.D)))

	var err error
	for _, dd := range t.D {
		__antithesis_instrumentation__.Notify(571612)
		appendTo, err = Encode(appendTo, NoColumnID, dd, scratch)
		if err != nil {
			__antithesis_instrumentation__.Notify(571613)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571614)
		}
	}
	__antithesis_instrumentation__.Notify(571611)
	return appendTo, nil
}

func decodeTuple(a *tree.DatumAlloc, tupTyp *types.T, b []byte) (tree.Datum, []byte, error) {
	__antithesis_instrumentation__.Notify(571615)
	b, _, _, err := encoding.DecodeNonsortingUvarint(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(571618)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(571619)
	}
	__antithesis_instrumentation__.Notify(571616)

	result := *(tree.NewDTuple(tupTyp))
	result.D = a.NewDatums(len(tupTyp.TupleContents()))
	var datum tree.Datum
	for i := range tupTyp.TupleContents() {
		__antithesis_instrumentation__.Notify(571620)
		datum, b, err = Decode(a, tupTyp.TupleContents()[i], b)
		if err != nil {
			__antithesis_instrumentation__.Notify(571622)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571623)
		}
		__antithesis_instrumentation__.Notify(571621)
		result.D[i] = datum
	}
	__antithesis_instrumentation__.Notify(571617)
	return a.NewDTuple(result), b, nil
}
