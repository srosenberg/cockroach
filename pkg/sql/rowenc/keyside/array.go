package keyside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

func encodeArrayKey(b []byte, array *tree.DArray, dir encoding.Direction) ([]byte, error) {
	__antithesis_instrumentation__.Notify(570644)
	var err error
	b = encoding.EncodeArrayKeyMarker(b, dir)
	for _, elem := range array.Array {
		__antithesis_instrumentation__.Notify(570646)
		if elem == tree.DNull {
			__antithesis_instrumentation__.Notify(570647)
			b = encoding.EncodeNullWithinArrayKey(b, dir)
		} else {
			__antithesis_instrumentation__.Notify(570648)
			b, err = Encode(b, elem, dir)
			if err != nil {
				__antithesis_instrumentation__.Notify(570649)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(570650)
			}
		}
	}
	__antithesis_instrumentation__.Notify(570645)
	return encoding.EncodeArrayKeyTerminator(b, dir), nil
}

func decodeArrayKey(
	a *tree.DatumAlloc, t *types.T, buf []byte, dir encoding.Direction,
) (tree.Datum, []byte, error) {
	__antithesis_instrumentation__.Notify(570651)
	var err error
	buf, err = encoding.ValidateAndConsumeArrayKeyMarker(buf, dir)
	if err != nil {
		__antithesis_instrumentation__.Notify(570655)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(570656)
	}
	__antithesis_instrumentation__.Notify(570652)

	result := tree.NewDArray(t.ArrayContents())
	if err = result.MaybeSetCustomOid(t); err != nil {
		__antithesis_instrumentation__.Notify(570657)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(570658)
	}
	__antithesis_instrumentation__.Notify(570653)

	for {
		__antithesis_instrumentation__.Notify(570659)
		if len(buf) == 0 {
			__antithesis_instrumentation__.Notify(570663)
			return nil, nil, errors.AssertionFailedf("invalid array encoding (unterminated)")
		} else {
			__antithesis_instrumentation__.Notify(570664)
		}
		__antithesis_instrumentation__.Notify(570660)
		if encoding.IsArrayKeyDone(buf, dir) {
			__antithesis_instrumentation__.Notify(570665)
			buf = buf[1:]
			break
		} else {
			__antithesis_instrumentation__.Notify(570666)
		}
		__antithesis_instrumentation__.Notify(570661)
		var d tree.Datum
		if encoding.IsNextByteArrayEncodedNull(buf, dir) {
			__antithesis_instrumentation__.Notify(570667)
			d = tree.DNull
			buf = buf[1:]
		} else {
			__antithesis_instrumentation__.Notify(570668)
			d, buf, err = Decode(a, t.ArrayContents(), buf, dir)
			if err != nil {
				__antithesis_instrumentation__.Notify(570669)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(570670)
			}
		}
		__antithesis_instrumentation__.Notify(570662)
		if err := result.Append(d); err != nil {
			__antithesis_instrumentation__.Notify(570671)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570672)
		}
	}
	__antithesis_instrumentation__.Notify(570654)
	return result, buf, nil
}
