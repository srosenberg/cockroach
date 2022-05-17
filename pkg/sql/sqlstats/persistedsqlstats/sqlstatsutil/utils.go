package sqlstatsutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
)

func DatumToUint64(d tree.Datum) (uint64, error) {
	__antithesis_instrumentation__.Notify(625264)
	b := []byte(tree.MustBeDBytes(d))

	_, val, err := encoding.DecodeUint64Ascending(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(625266)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(625267)
	}
	__antithesis_instrumentation__.Notify(625265)

	return val, nil
}
