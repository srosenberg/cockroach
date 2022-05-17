package instancestorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

type rowCodec struct {
	codec   keys.SQLCodec
	columns []catalog.Column
	decoder valueside.Decoder
}

func makeRowCodec(codec keys.SQLCodec) rowCodec {
	__antithesis_instrumentation__.Notify(624121)
	columns := systemschema.SQLInstancesTable.PublicColumns()
	return rowCodec{
		codec:   codec,
		columns: columns,
		decoder: valueside.MakeDecoder(columns),
	}
}

func (d *rowCodec) encodeRow(
	instanceID base.SQLInstanceID,
	addr string,
	sessionID sqlliveness.SessionID,
	codec keys.SQLCodec,
	tableID descpb.ID,
) (kv kv.KeyValue, err error) {
	__antithesis_instrumentation__.Notify(624122)
	addrDatum := tree.NewDString(addr)
	var valueBuf []byte
	valueBuf, err = valueside.Encode(
		[]byte(nil), valueside.MakeColumnIDDelta(0, d.columns[1].GetID()), addrDatum, []byte(nil))
	if err != nil {
		__antithesis_instrumentation__.Notify(624125)
		return kv, err
	} else {
		__antithesis_instrumentation__.Notify(624126)
	}
	__antithesis_instrumentation__.Notify(624123)
	sessionDatum := tree.NewDBytes(tree.DBytes(sessionID.UnsafeBytes()))
	sessionColDiff := valueside.MakeColumnIDDelta(d.columns[1].GetID(), d.columns[2].GetID())
	valueBuf, err = valueside.Encode(valueBuf, sessionColDiff, sessionDatum, []byte(nil))
	if err != nil {
		__antithesis_instrumentation__.Notify(624127)
		return kv, err
	} else {
		__antithesis_instrumentation__.Notify(624128)
	}
	__antithesis_instrumentation__.Notify(624124)
	var v roachpb.Value
	v.SetTuple(valueBuf)
	kv.Value = &v
	kv.Key = makeInstanceKey(codec, tableID, instanceID)
	return kv, nil
}

func (d *rowCodec) decodeRow(
	kv kv.KeyValue,
) (
	instanceID base.SQLInstanceID,
	addr string,
	sessionID sqlliveness.SessionID,
	timestamp hlc.Timestamp,
	tombstone bool,
	_ error,
) {
	__antithesis_instrumentation__.Notify(624129)
	var alloc tree.DatumAlloc

	{
		__antithesis_instrumentation__.Notify(624136)
		types := []*types.T{d.columns[0].GetType()}
		row := make([]rowenc.EncDatum, 1)
		_, _, err := rowenc.DecodeIndexKey(d.codec, types, row, nil, kv.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(624139)
			return base.SQLInstanceID(0), "", "", hlc.Timestamp{}, false, errors.Wrap(err, "failed to decode key")
		} else {
			__antithesis_instrumentation__.Notify(624140)
		}
		__antithesis_instrumentation__.Notify(624137)
		if err := row[0].EnsureDecoded(types[0], &alloc); err != nil {
			__antithesis_instrumentation__.Notify(624141)
			return base.SQLInstanceID(0), "", "", hlc.Timestamp{}, false, err
		} else {
			__antithesis_instrumentation__.Notify(624142)
		}
		__antithesis_instrumentation__.Notify(624138)
		instanceID = base.SQLInstanceID(tree.MustBeDInt(row[0].Datum))
	}
	__antithesis_instrumentation__.Notify(624130)
	if !kv.Value.IsPresent() {
		__antithesis_instrumentation__.Notify(624143)
		return instanceID, "", "", hlc.Timestamp{}, true, nil
	} else {
		__antithesis_instrumentation__.Notify(624144)
	}
	__antithesis_instrumentation__.Notify(624131)
	timestamp = kv.Value.Timestamp

	bytes, err := kv.Value.GetTuple()
	if err != nil {
		__antithesis_instrumentation__.Notify(624145)
		return instanceID, "", "", hlc.Timestamp{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(624146)
	}
	__antithesis_instrumentation__.Notify(624132)

	datums, err := d.decoder.Decode(&alloc, bytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(624147)
		return instanceID, "", "", hlc.Timestamp{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(624148)
	}
	__antithesis_instrumentation__.Notify(624133)

	if addrVal := datums[1]; addrVal != tree.DNull {
		__antithesis_instrumentation__.Notify(624149)
		addr = string(tree.MustBeDString(addrVal))
	} else {
		__antithesis_instrumentation__.Notify(624150)
	}
	__antithesis_instrumentation__.Notify(624134)
	if sessionIDVal := datums[2]; sessionIDVal != tree.DNull {
		__antithesis_instrumentation__.Notify(624151)
		sessionID = sqlliveness.SessionID(tree.MustBeDBytes(sessionIDVal))
	} else {
		__antithesis_instrumentation__.Notify(624152)
	}
	__antithesis_instrumentation__.Notify(624135)

	return instanceID, addr, sessionID, timestamp, false, nil
}

func makeTablePrefix(codec keys.SQLCodec, tableID descpb.ID) roachpb.Key {
	__antithesis_instrumentation__.Notify(624153)
	return codec.IndexPrefix(uint32(tableID), 1)
}

func makeInstanceKey(
	codec keys.SQLCodec, tableID descpb.ID, instanceID base.SQLInstanceID,
) roachpb.Key {
	__antithesis_instrumentation__.Notify(624154)
	return keys.MakeFamilyKey(encoding.EncodeVarintAscending(makeTablePrefix(codec, tableID), int64(instanceID)), 0)
}
