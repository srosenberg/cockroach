package config

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
)

func MakeZoneKeyPrefix(codec keys.SQLCodec, id descpb.ID) roachpb.Key {
	__antithesis_instrumentation__.Notify(55788)
	return codec.ZoneKeyPrefix(uint32(id))
}

func MakeZoneKey(codec keys.SQLCodec, id descpb.ID) roachpb.Key {
	__antithesis_instrumentation__.Notify(55789)
	return codec.ZoneKey(uint32(id))
}

func DecodeObjectID(codec keys.SQLCodec, key roachpb.RKey) (ObjectID, []byte, bool) {
	__antithesis_instrumentation__.Notify(55790)
	rem, id, err := codec.DecodeTablePrefix(key.AsRawKey())
	return ObjectID(id), rem, err == nil
}
