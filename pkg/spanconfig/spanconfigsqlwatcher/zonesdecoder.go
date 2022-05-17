package spanconfigsqlwatcher

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
)

type zonesDecoder struct {
	alloc tree.DatumAlloc
	codec keys.SQLCodec
}

func newZonesDecoder(codec keys.SQLCodec) *zonesDecoder {
	__antithesis_instrumentation__.Notify(241460)
	return &zonesDecoder{
		codec: codec,
	}
}

func (zd *zonesDecoder) DecodePrimaryKey(key roachpb.Key) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(241461)

	tbl := systemschema.ZonesTable
	types := []*types.T{tbl.PublicColumns()[0].GetType()}
	startKeyRow := make([]rowenc.EncDatum, 1)
	_, _, err := rowenc.DecodeIndexKey(
		zd.codec, types, startKeyRow, nil, key,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(241464)
		return descpb.InvalidID, errors.NewAssertionErrorWithWrappedErrf(err, "failed to decode key in system.zones %v", key)
	} else {
		__antithesis_instrumentation__.Notify(241465)
	}
	__antithesis_instrumentation__.Notify(241462)
	if err := startKeyRow[0].EnsureDecoded(types[0], &zd.alloc); err != nil {
		__antithesis_instrumentation__.Notify(241466)
		return descpb.InvalidID, errors.NewAssertionErrorWithWrappedErrf(err, "failed to decode key in system.zones %v", key)
	} else {
		__antithesis_instrumentation__.Notify(241467)
	}
	__antithesis_instrumentation__.Notify(241463)
	descID := descpb.ID(tree.MustBeDInt(startKeyRow[0].Datum))
	return descID, nil
}

func TestingZonesDecoderDecodePrimaryKey(codec keys.SQLCodec, key roachpb.Key) (descpb.ID, error) {
	__antithesis_instrumentation__.Notify(241468)
	return newZonesDecoder(codec).DecodePrimaryKey(key)
}
