package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

func DecodeDatum(datumAlloc *tree.DatumAlloc, typ *types.T, data []byte) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(471514)
	datum, rem, err := valueside.Decode(datumAlloc, typ, data)
	if err != nil {
		__antithesis_instrumentation__.Notify(471517)
		return nil, errors.NewAssertionErrorWithWrappedErrf(err,
			"error decoding %d bytes", errors.Safe(len(data)))
	} else {
		__antithesis_instrumentation__.Notify(471518)
	}
	__antithesis_instrumentation__.Notify(471515)
	if len(rem) != 0 {
		__antithesis_instrumentation__.Notify(471519)
		return nil, errors.AssertionFailedf(
			"%d trailing bytes in encoded value", errors.Safe(len(rem)))
	} else {
		__antithesis_instrumentation__.Notify(471520)
	}
	__antithesis_instrumentation__.Notify(471516)
	return datum, nil
}
