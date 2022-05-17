package tenantsettingswatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type RowDecoder struct {
	alloc   tree.DatumAlloc
	columns []catalog.Column
	decoder valueside.Decoder
}

var allTenantOverridesID = roachpb.TenantID{InternalValue: 0}

func MakeRowDecoder() RowDecoder {
	__antithesis_instrumentation__.Notify(238915)
	columns := systemschema.TenantSettingsTable.PublicColumns()
	return RowDecoder{
		columns: columns,
		decoder: valueside.MakeDecoder(columns),
	}
}

func (d *RowDecoder) DecodeRow(
	kv roachpb.KeyValue,
) (_ roachpb.TenantID, _ roachpb.TenantSetting, tombstone bool, _ error) {
	__antithesis_instrumentation__.Notify(238916)

	keyTypes := []*types.T{d.columns[0].GetType(), d.columns[1].GetType()}
	keyVals := make([]rowenc.EncDatum, 2)
	_, _, err := rowenc.DecodeIndexKey(keys.SystemSQLCodec, keyTypes, keyVals, nil, kv.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(238924)
		return roachpb.TenantID{}, roachpb.TenantSetting{}, false, errors.Wrap(err, "failed to decode key")
	} else {
		__antithesis_instrumentation__.Notify(238925)
	}
	__antithesis_instrumentation__.Notify(238917)
	for i := range keyVals {
		__antithesis_instrumentation__.Notify(238926)
		if err := keyVals[i].EnsureDecoded(keyTypes[i], &d.alloc); err != nil {
			__antithesis_instrumentation__.Notify(238927)
			return roachpb.TenantID{}, roachpb.TenantSetting{}, false, err
		} else {
			__antithesis_instrumentation__.Notify(238928)
		}
	}
	__antithesis_instrumentation__.Notify(238918)

	tenantID := roachpb.TenantID{InternalValue: uint64(tree.MustBeDInt(keyVals[0].Datum))}
	var setting roachpb.TenantSetting
	setting.Name = string(tree.MustBeDString(keyVals[1].Datum))
	if !kv.Value.IsPresent() {
		__antithesis_instrumentation__.Notify(238929)
		return tenantID, setting, true, nil
	} else {
		__antithesis_instrumentation__.Notify(238930)
	}
	__antithesis_instrumentation__.Notify(238919)

	bytes, err := kv.Value.GetTuple()
	if err != nil {
		__antithesis_instrumentation__.Notify(238931)
		return roachpb.TenantID{}, roachpb.TenantSetting{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(238932)
	}
	__antithesis_instrumentation__.Notify(238920)

	datums, err := d.decoder.Decode(&d.alloc, bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(238933)
		return roachpb.TenantID{}, roachpb.TenantSetting{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(238934)
	}
	__antithesis_instrumentation__.Notify(238921)

	if value := datums[2]; value != tree.DNull {
		__antithesis_instrumentation__.Notify(238935)
		setting.Value.Value = string(tree.MustBeDString(value))
	} else {
		__antithesis_instrumentation__.Notify(238936)
	}
	__antithesis_instrumentation__.Notify(238922)
	if typ := datums[4]; typ != tree.DNull {
		__antithesis_instrumentation__.Notify(238937)
		setting.Value.Type = string(tree.MustBeDString(typ))
	} else {
		__antithesis_instrumentation__.Notify(238938)

		setting.Value.Type = "s"
	}
	__antithesis_instrumentation__.Notify(238923)

	return tenantID, setting, false, nil
}
