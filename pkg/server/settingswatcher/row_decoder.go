package settingswatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type RowDecoder struct {
	codec   keys.SQLCodec
	alloc   tree.DatumAlloc
	columns []catalog.Column
	decoder valueside.Decoder
}

func MakeRowDecoder(codec keys.SQLCodec) RowDecoder {
	__antithesis_instrumentation__.Notify(235099)
	columns := systemschema.SettingsTable.PublicColumns()
	return RowDecoder{
		codec:   codec,
		columns: columns,
		decoder: valueside.MakeDecoder(columns),
	}
}

func (d *RowDecoder) DecodeRow(
	kv roachpb.KeyValue,
) (setting string, val settings.EncodedValue, tombstone bool, _ error) {
	__antithesis_instrumentation__.Notify(235100)

	{
		__antithesis_instrumentation__.Notify(235107)
		types := []*types.T{d.columns[0].GetType()}
		nameRow := make([]rowenc.EncDatum, 1)
		_, _, err := rowenc.DecodeIndexKey(d.codec, types, nameRow, nil, kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(235110)
			return "", settings.EncodedValue{}, false, errors.Wrap(err, "failed to decode key")
		} else {
			__antithesis_instrumentation__.Notify(235111)
		}
		__antithesis_instrumentation__.Notify(235108)
		if err := nameRow[0].EnsureDecoded(types[0], &d.alloc); err != nil {
			__antithesis_instrumentation__.Notify(235112)
			return "", settings.EncodedValue{}, false, err
		} else {
			__antithesis_instrumentation__.Notify(235113)
		}
		__antithesis_instrumentation__.Notify(235109)
		setting = string(tree.MustBeDString(nameRow[0].Datum))
	}
	__antithesis_instrumentation__.Notify(235101)
	if !kv.Value.IsPresent() {
		__antithesis_instrumentation__.Notify(235114)
		return setting, settings.EncodedValue{}, true, nil
	} else {
		__antithesis_instrumentation__.Notify(235115)
	}
	__antithesis_instrumentation__.Notify(235102)

	bytes, err := kv.Value.GetTuple()
	if err != nil {
		__antithesis_instrumentation__.Notify(235116)
		return "", settings.EncodedValue{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(235117)
	}
	__antithesis_instrumentation__.Notify(235103)

	datums, err := d.decoder.Decode(&d.alloc, bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(235118)
		return "", settings.EncodedValue{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(235119)
	}
	__antithesis_instrumentation__.Notify(235104)

	if value := datums[1]; value != tree.DNull {
		__antithesis_instrumentation__.Notify(235120)
		val.Value = string(tree.MustBeDString(value))
	} else {
		__antithesis_instrumentation__.Notify(235121)
	}
	__antithesis_instrumentation__.Notify(235105)
	if typ := datums[3]; typ != tree.DNull {
		__antithesis_instrumentation__.Notify(235122)
		val.Type = string(tree.MustBeDString(typ))
	} else {
		__antithesis_instrumentation__.Notify(235123)

		val.Type = "s"
	}
	__antithesis_instrumentation__.Notify(235106)

	return setting, val, false, nil
}
