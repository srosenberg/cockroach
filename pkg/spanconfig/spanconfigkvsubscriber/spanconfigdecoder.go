package spanconfigkvsubscriber

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangefeed/rangefeedbuffer"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type spanConfigDecoder struct {
	alloc   tree.DatumAlloc
	columns []catalog.Column
	decoder valueside.Decoder
}

func newSpanConfigDecoder() *spanConfigDecoder {
	__antithesis_instrumentation__.Notify(240637)
	columns := systemschema.SpanConfigurationsTable.PublicColumns()
	return &spanConfigDecoder{
		columns: columns,
		decoder: valueside.MakeDecoder(columns),
	}
}

func (sd *spanConfigDecoder) decode(kv roachpb.KeyValue) (spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(240638)

	var rawSp roachpb.Span
	var conf roachpb.SpanConfig
	{
		__antithesis_instrumentation__.Notify(240645)
		types := []*types.T{sd.columns[0].GetType()}
		startKeyRow := make([]rowenc.EncDatum, 1)
		_, _, err := rowenc.DecodeIndexKey(keys.SystemSQLCodec, types, startKeyRow, nil, kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(240648)
			return spanconfig.Record{}, errors.Wrapf(err, "failed to decode key: %v", kv.Key)
		} else {
			__antithesis_instrumentation__.Notify(240649)
		}
		__antithesis_instrumentation__.Notify(240646)
		if err := startKeyRow[0].EnsureDecoded(types[0], &sd.alloc); err != nil {
			__antithesis_instrumentation__.Notify(240650)
			return spanconfig.Record{}, err
		} else {
			__antithesis_instrumentation__.Notify(240651)
		}
		__antithesis_instrumentation__.Notify(240647)
		rawSp.Key = []byte(tree.MustBeDBytes(startKeyRow[0].Datum))
	}
	__antithesis_instrumentation__.Notify(240639)
	if !kv.Value.IsPresent() {
		__antithesis_instrumentation__.Notify(240652)
		return spanconfig.Record{},
			errors.AssertionFailedf("missing value for start key: %s", rawSp.Key)
	} else {
		__antithesis_instrumentation__.Notify(240653)
	}
	__antithesis_instrumentation__.Notify(240640)

	bytes, err := kv.Value.GetTuple()
	if err != nil {
		__antithesis_instrumentation__.Notify(240654)
		return spanconfig.Record{}, err
	} else {
		__antithesis_instrumentation__.Notify(240655)
	}
	__antithesis_instrumentation__.Notify(240641)

	datums, err := sd.decoder.Decode(&sd.alloc, bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(240656)
		return spanconfig.Record{}, err
	} else {
		__antithesis_instrumentation__.Notify(240657)
	}
	__antithesis_instrumentation__.Notify(240642)
	if endKey := datums[1]; endKey != tree.DNull {
		__antithesis_instrumentation__.Notify(240658)
		rawSp.EndKey = []byte(tree.MustBeDBytes(endKey))
	} else {
		__antithesis_instrumentation__.Notify(240659)
	}
	__antithesis_instrumentation__.Notify(240643)
	if config := datums[2]; config != tree.DNull {
		__antithesis_instrumentation__.Notify(240660)
		if err := protoutil.Unmarshal([]byte(tree.MustBeDBytes(config)), &conf); err != nil {
			__antithesis_instrumentation__.Notify(240661)
			return spanconfig.Record{}, err
		} else {
			__antithesis_instrumentation__.Notify(240662)
		}
	} else {
		__antithesis_instrumentation__.Notify(240663)
	}
	__antithesis_instrumentation__.Notify(240644)

	return spanconfig.MakeRecord(spanconfig.DecodeTarget(rawSp), conf)
}

func (sd *spanConfigDecoder) translateEvent(
	ctx context.Context, ev *roachpb.RangeFeedValue,
) rangefeedbuffer.Event {
	__antithesis_instrumentation__.Notify(240664)
	deleted := !ev.Value.IsPresent()
	var value roachpb.Value
	if deleted {
		__antithesis_instrumentation__.Notify(240669)
		if !ev.PrevValue.IsPresent() {
			__antithesis_instrumentation__.Notify(240671)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(240672)
		}
		__antithesis_instrumentation__.Notify(240670)

		value = ev.PrevValue
	} else {
		__antithesis_instrumentation__.Notify(240673)
		value = ev.Value
	}
	__antithesis_instrumentation__.Notify(240665)
	record, err := sd.decode(roachpb.KeyValue{
		Key:   ev.Key,
		Value: value,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(240674)
		log.Fatalf(ctx, "failed to decode row: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(240675)
	}
	__antithesis_instrumentation__.Notify(240666)

	if log.ExpensiveLogEnabled(ctx, 1) {
		__antithesis_instrumentation__.Notify(240676)
		log.Infof(ctx, "received span configuration update for %s (deleted=%t)",
			record.GetTarget(), deleted)
	} else {
		__antithesis_instrumentation__.Notify(240677)
	}
	__antithesis_instrumentation__.Notify(240667)

	var update spanconfig.Update
	if deleted {
		__antithesis_instrumentation__.Notify(240678)
		update, err = spanconfig.Deletion(record.GetTarget())
		if err != nil {
			__antithesis_instrumentation__.Notify(240679)
			log.Fatalf(ctx, "failed to construct Deletion: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(240680)
		}
	} else {
		__antithesis_instrumentation__.Notify(240681)
		update = spanconfig.Update(record)
	}
	__antithesis_instrumentation__.Notify(240668)

	return &bufferEvent{update, ev.Value.Timestamp}
}

func TestingDecoderFn() func(roachpb.KeyValue) (spanconfig.Record, error) {
	__antithesis_instrumentation__.Notify(240682)
	return newSpanConfigDecoder().decode
}
