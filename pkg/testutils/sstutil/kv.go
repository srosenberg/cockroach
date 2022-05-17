package sstutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type KV struct {
	KeyString     string
	WallTimestamp int64
	ValueString   string
}

func (kv KV) Key() roachpb.Key {
	__antithesis_instrumentation__.Notify(646399)
	return roachpb.Key(kv.KeyString)
}

func (kv KV) Timestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(646400)
	return hlc.Timestamp{WallTime: kv.WallTimestamp}
}

func (kv KV) MVCCKey() storage.MVCCKey {
	__antithesis_instrumentation__.Notify(646401)
	return storage.MVCCKey{
		Key:       kv.Key(),
		Timestamp: kv.Timestamp(),
	}
}

func (kv KV) Value() roachpb.Value {
	__antithesis_instrumentation__.Notify(646402)
	value := roachpb.MakeValueFromString(kv.ValueString)
	if kv.ValueString == "" {
		__antithesis_instrumentation__.Notify(646404)
		value = roachpb.Value{}
	} else {
		__antithesis_instrumentation__.Notify(646405)
	}
	__antithesis_instrumentation__.Notify(646403)
	value.InitChecksum(kv.Key())
	return value
}

func (kv KV) ValueBytes() []byte {
	__antithesis_instrumentation__.Notify(646406)
	return kv.Value().RawBytes
}
