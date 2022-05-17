package spanconfigsqlwatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts/ptpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type protectedTimestampDecoder struct {
	alloc   tree.DatumAlloc
	decoder valueside.Decoder
}

func newProtectedTimestampDecoder() *protectedTimestampDecoder {
	__antithesis_instrumentation__.Notify(241326)
	columns := systemschema.ProtectedTimestampsRecordsTable.PublicColumns()
	return &protectedTimestampDecoder{
		decoder: valueside.MakeDecoder(columns),
	}
}

func (d *protectedTimestampDecoder) decode(kv roachpb.KeyValue) (target ptpb.Target, _ error) {
	__antithesis_instrumentation__.Notify(241327)
	if !kv.Value.IsPresent() {
		__antithesis_instrumentation__.Notify(241333)
		return ptpb.Target{},
			errors.AssertionFailedf("missing value for key in system.protected_ts_records: %v", kv)
	} else {
		__antithesis_instrumentation__.Notify(241334)
	}
	__antithesis_instrumentation__.Notify(241328)

	bytes, err := kv.Value.GetTuple()
	if err != nil {
		__antithesis_instrumentation__.Notify(241335)
		return ptpb.Target{}, err
	} else {
		__antithesis_instrumentation__.Notify(241336)
	}
	__antithesis_instrumentation__.Notify(241329)

	datums, err := d.decoder.Decode(&d.alloc, bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(241337)
		return ptpb.Target{}, err
	} else {
		__antithesis_instrumentation__.Notify(241338)
	}
	__antithesis_instrumentation__.Notify(241330)

	if len(datums) != 8 {
		__antithesis_instrumentation__.Notify(241339)
		return ptpb.Target{}, errors.AssertionFailedf("expected len(datums) == 8, but found %d", len(datums))
	} else {
		__antithesis_instrumentation__.Notify(241340)
	}
	__antithesis_instrumentation__.Notify(241331)

	if t := datums[7]; t != tree.DNull {
		__antithesis_instrumentation__.Notify(241341)
		targetBytes := tree.MustBeDBytes(t)
		if err := protoutil.Unmarshal([]byte(targetBytes), &target); err != nil {
			__antithesis_instrumentation__.Notify(241342)
			return ptpb.Target{}, errors.Wrap(err, "failed to unmarshal target")
		} else {
			__antithesis_instrumentation__.Notify(241343)
		}
	} else {
		__antithesis_instrumentation__.Notify(241344)
	}
	__antithesis_instrumentation__.Notify(241332)

	return target, nil
}

func TestingProtectedTimestampDecoderFn() func(roachpb.KeyValue) (ptpb.Target, error) {
	__antithesis_instrumentation__.Notify(241345)
	return newProtectedTimestampDecoder().decode
}
