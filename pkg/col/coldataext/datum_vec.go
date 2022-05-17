package coldataext

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type datumVec struct {
	t *types.T

	data []tree.Datum

	evalCtx *tree.EvalContext

	scratch []byte
	da      tree.DatumAlloc
}

var _ coldata.DatumVec = &datumVec{}

func newDatumVec(t *types.T, n int, evalCtx *tree.EvalContext) coldata.DatumVec {
	__antithesis_instrumentation__.Notify(54656)
	return &datumVec{
		t:       t,
		data:    make([]tree.Datum, n),
		evalCtx: evalCtx,
	}
}

func CompareDatum(d, dVec, other interface{}) int {
	__antithesis_instrumentation__.Notify(54657)
	return d.(tree.Datum).Compare(dVec.(*datumVec).evalCtx, convertToDatum(other))
}

func Hash(d tree.Datum, da *tree.DatumAlloc) []byte {
	__antithesis_instrumentation__.Notify(54658)
	ed := rowenc.EncDatum{Datum: convertToDatum(d)}

	b, err := ed.Fingerprint(context.TODO(), d.ResolvedType(), da, nil, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(54660)

		colexecerror.ExpectedError(err)
	} else {
		__antithesis_instrumentation__.Notify(54661)
	}
	__antithesis_instrumentation__.Notify(54659)
	return b
}

func (dv *datumVec) Get(i int) coldata.Datum {
	__antithesis_instrumentation__.Notify(54662)
	dv.maybeSetDNull(i)
	return dv.data[i]
}

func (dv *datumVec) Set(i int, v coldata.Datum) {
	__antithesis_instrumentation__.Notify(54663)
	datum := convertToDatum(v)
	dv.assertValidDatum(datum)
	dv.data[i] = datum
}

func (dv *datumVec) Window(start, end int) coldata.DatumVec {
	__antithesis_instrumentation__.Notify(54664)
	return &datumVec{
		t:       dv.t,
		data:    dv.data[start:end],
		evalCtx: dv.evalCtx,
	}
}

func (dv *datumVec) CopySlice(src coldata.DatumVec, destIdx, srcStartIdx, srcEndIdx int) {
	__antithesis_instrumentation__.Notify(54665)
	castSrc := src.(*datumVec)
	dv.assertSameTypeFamily(castSrc.t)
	copy(dv.data[destIdx:], castSrc.data[srcStartIdx:srcEndIdx])
}

func (dv *datumVec) AppendSlice(src coldata.DatumVec, destIdx, srcStartIdx, srcEndIdx int) {
	__antithesis_instrumentation__.Notify(54666)
	castSrc := src.(*datumVec)
	dv.assertSameTypeFamily(castSrc.t)
	dv.data = append(dv.data[:destIdx], castSrc.data[srcStartIdx:srcEndIdx]...)
}

func (dv *datumVec) AppendVal(v coldata.Datum) {
	__antithesis_instrumentation__.Notify(54667)
	datum := convertToDatum(v)
	dv.assertValidDatum(datum)
	dv.data = append(dv.data, datum)
}

func (dv *datumVec) Len() int {
	__antithesis_instrumentation__.Notify(54668)
	return len(dv.data)
}

func (dv *datumVec) Cap() int {
	__antithesis_instrumentation__.Notify(54669)
	return cap(dv.data)
}

func (dv *datumVec) MarshalAt(appendTo []byte, i int) ([]byte, error) {
	__antithesis_instrumentation__.Notify(54670)
	dv.maybeSetDNull(i)
	return valueside.Encode(
		appendTo, valueside.NoColumnID, dv.data[i], dv.scratch,
	)
}

func (dv *datumVec) UnmarshalTo(i int, b []byte) error {
	__antithesis_instrumentation__.Notify(54671)
	var err error
	dv.data[i], _, err = valueside.Decode(&dv.da, dv.t, b)
	return err
}

func (dv *datumVec) valuesSize(startIdx int) int64 {
	__antithesis_instrumentation__.Notify(54672)
	var size int64

	if startIdx < dv.Len() {
		__antithesis_instrumentation__.Notify(54674)
		for _, d := range dv.data[startIdx:dv.Len()] {
			__antithesis_instrumentation__.Notify(54675)
			if d != nil {
				__antithesis_instrumentation__.Notify(54676)
				size += int64(d.Size())
			} else {
				__antithesis_instrumentation__.Notify(54677)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(54678)
	}
	__antithesis_instrumentation__.Notify(54673)
	return size
}

func (dv *datumVec) Size(startIdx int) int64 {
	__antithesis_instrumentation__.Notify(54679)

	if startIdx >= dv.Cap() {
		__antithesis_instrumentation__.Notify(54682)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(54683)
	}
	__antithesis_instrumentation__.Notify(54680)
	if startIdx < 0 {
		__antithesis_instrumentation__.Notify(54684)
		startIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(54685)
	}
	__antithesis_instrumentation__.Notify(54681)

	return memsize.DatumOverhead*int64(dv.Cap()-startIdx) + dv.valuesSize(startIdx)
}

func (dv *datumVec) Reset() int64 {
	__antithesis_instrumentation__.Notify(54686)
	released := dv.valuesSize(0)
	for i := range dv.data {
		__antithesis_instrumentation__.Notify(54688)
		dv.data[i] = nil
	}
	__antithesis_instrumentation__.Notify(54687)
	return released
}

func (dv *datumVec) assertValidDatum(datum tree.Datum) {
	__antithesis_instrumentation__.Notify(54689)
	if datum != tree.DNull {
		__antithesis_instrumentation__.Notify(54690)
		dv.assertSameTypeFamily(datum.ResolvedType())
	} else {
		__antithesis_instrumentation__.Notify(54691)
	}
}

func (dv *datumVec) assertSameTypeFamily(t *types.T) {
	__antithesis_instrumentation__.Notify(54692)
	if dv.t.Family() != t.Family() {
		__antithesis_instrumentation__.Notify(54693)
		colexecerror.InternalError(
			errors.AssertionFailedf("cannot use value of type %+v on a datumVec of type %+v", t, dv.t),
		)
	} else {
		__antithesis_instrumentation__.Notify(54694)
	}
}

func (dv *datumVec) maybeSetDNull(i int) {
	__antithesis_instrumentation__.Notify(54695)
	if dv.data[i] == nil {
		__antithesis_instrumentation__.Notify(54696)
		dv.data[i] = tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(54697)
	}
}

func convertToDatum(v coldata.Datum) tree.Datum {
	__antithesis_instrumentation__.Notify(54698)
	if v == nil {
		__antithesis_instrumentation__.Notify(54701)
		return tree.DNull
	} else {
		__antithesis_instrumentation__.Notify(54702)
	}
	__antithesis_instrumentation__.Notify(54699)
	if datum, ok := v.(tree.Datum); ok {
		__antithesis_instrumentation__.Notify(54703)
		return datum
	} else {
		__antithesis_instrumentation__.Notify(54704)
	}
	__antithesis_instrumentation__.Notify(54700)
	colexecerror.InternalError(errors.AssertionFailedf("unexpected value: %v", v))

	return nil
}
