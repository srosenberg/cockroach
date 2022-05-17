package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/gob"
	"io"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
)

func DumpRawTo(src tspb.TimeSeries_DumpRawClient, out io.Writer) error {
	__antithesis_instrumentation__.Notify(650196)
	enc := gob.NewEncoder(out)
	for {
		__antithesis_instrumentation__.Notify(650197)
		data, err := src.Recv()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(650200)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(650201)
		}
		__antithesis_instrumentation__.Notify(650198)
		if err != nil {
			__antithesis_instrumentation__.Notify(650202)
			return err
		} else {
			__antithesis_instrumentation__.Notify(650203)
		}
		__antithesis_instrumentation__.Notify(650199)
		if err := enc.Encode(data); err != nil {
			__antithesis_instrumentation__.Notify(650204)
			return err
		} else {
			__antithesis_instrumentation__.Notify(650205)
		}
	}
}
